package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apiv0 "github.com/sakura-internet/distributed-mariadb-controller/cmd/sakura-controller/api/v0"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bash"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller/sakura"
	"github.com/vishvananda/netlink"

	"golang.org/x/exp/rand"
	"golang.org/x/exp/slog"
)

func main() {
	rand.Seed(uint64(time.Now().UnixNano()))

	ctx, cancel := context.WithCancel(context.Background())

	if err := parseAllFlags(os.Args[1:]); err != nil {
		panic(err)
	}
	if err := validateAllFlags(); err != nil {
		panic(err)
	}

	logger := setupGlobalLogger(os.Stderr, LogLevelFlag)

	// mkdir for lock file
	if err := os.MkdirAll(filepath.Dir(LockFilePathFlag), 0o755); err != nil {
		panic(err)
	}

	lockf, err := tryToGetTheExclusiveLockWithoutBlocking(LockFilePathFlag)
	if err != nil {
		panic(err)
	}
	defer lockf.Close()

	// for controlling the traffics that they're to the DB server port.
	// the function returns nil if the expected chain is already exist.
	if err := createNftablesChain(logger); err != nil {
		panic(err)
	}

	c := sakura.NewSAKURAController(logger)

	{
		eth0Address, err := getEth0NetIFAddress()
		if err != nil {
			panic(err)
		}

		logger.Debug("eth0 address", "address", eth0Address)
		c.HostAddress = eth0Address
	}
	{
		dbReplicaPassword, err := readDBReplicaPassword(DBReplicaPasswordFilePathFlag)
		if err != nil {
			panic(err)
		}

		c.MariaDBReplicaPassword = dbReplicaPassword
	}

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, c *sakura.SAKURAController) {
		wg.Done()
		controller.Start(ctx, logger, controller.Controller(c), time.Second*time.Duration(MainPollingSpanSecondFlag))
	}(ctx, wg, c)

	if EnablePrometheusExporterFlag {
		wg.Add(1)
		go startPrometheusExporterServer(ctx, wg)
	}

	if EnableHTTPAPIFlag {
		wg.Add(1)
		go startHTTPAPIServer(ctx, wg, c)
	}

	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	stopSigCh := make(chan os.Signal, 3)
	signal.Notify(stopSigCh, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

mainLoop:
	for range stopSigCh {
		logger.Info("got stop signal. exiting.")

		// for stopping all goroutine.
		cancel()
		break mainLoop
	}

	wg.Wait()

	logger.Info("db-controller exited. see you again, bye.")
}

// startPrometheusExporterServer starts the HTTP server that serves the prometheus-exporter endpoint.
func startPrometheusExporterServer(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Setup
	e := echo.New()

	switch LogLevelFlag {
	case "info":
		e.Logger.SetLevel(log.INFO)
	case "debug":
		e.Logger.SetLevel(log.DEBUG)
		e.Use(middleware.Logger())
	case "warning":
		e.Logger.SetLevel(log.WARN)
	case "error":
		e.Logger.SetLevel(log.ERROR)
	}

	reg := sakura.NewPrometheusMetricRegistry()
	e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	// Start server
	addr := fmt.Sprintf(":%d", PrometheusExporterPortFlag)

	ch := make(chan bool, 1)
	go func(ch chan<- bool) {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}

		ch <- true
	}(ch)

waitLoop:
	for {
		select {
		case <-ctx.Done():
			if err := e.Shutdown(ctx); err != nil {
				e.Logger.Fatal(err)
			}
			<-ch
			break waitLoop
		}
	}
}

// startHTTPAPIServer starts the HTTP API server that serves the controller status responder.
func startHTTPAPIServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	c *sakura.SAKURAController,
) {
	defer wg.Done()

	// Setup
	e := echo.New()
	e.Use(apiv0.UseControllerState(c))

	switch LogLevelFlag {
	case "info":
		e.Logger.SetLevel(log.INFO)
	case "debug":
		e.Logger.SetLevel(log.DEBUG)
		e.Use(middleware.Logger())
	case "warning":
		e.Logger.SetLevel(log.WARN)
	case "error":
		e.Logger.SetLevel(log.ERROR)
	}

	e.HEAD("/healthcheck", apiv0.GSLBHealthCheckEndpoint)
	e.GET("/healthcheck", apiv0.GSLBHealthCheckEndpoint)
	e.GET("/status", apiv0.GetDBControllerStatus)
	// Start server
	addr := fmt.Sprintf(":%d", HTTPAPIServerPortFlag)

	ch := make(chan bool, 1)
	go func(ch chan<- bool) {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}

		ch <- true
	}(ch)

waitLoop:
	for {
		select {
		case <-ctx.Done():
			if err := e.Shutdown(ctx); err != nil {
				e.Logger.Fatal(err)
			}
			<-ch
			break waitLoop
		}
	}
}

// createNftablesChain tries to create an nftables chain on filter table.
func createNftablesChain(
	logger *slog.Logger,
) error {
	const (
		chainName = "mariadb"
	)
	// nft add chain comand returns ok if the chain is already exist.
	cmd := fmt.Sprintf("nft add chain filter %s { type filter hook input priority 0\\; }", chainName)
	logger.Info("execute command", "command", cmd, "callerFn", "createNftablesChain")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to add nft chain: %w", err)
	}

	return nil
}

// getEth0NetIFAddress tries to get the IP address of the eth0 I/F using Netlink messages.
func getEth0NetIFAddress() (string, error) {
	eth, err := netlink.LinkByName("eth0")
	if err != nil {
		return "", err
	}

	addrs, err := netlink.AddrList(eth, netlink.FAMILY_V4)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("eth0 doesn't have any IP addresses")
	}

	return addrs[0].IP.String(), nil
}

// tryToGetTheExclusiveLockWithoutBlocking uses flock(2) to get the exclusive lock of the path.
func tryToGetTheExclusiveLockWithoutBlocking(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, err
	}

	return f, nil
}

// setupGlobalLogger setups a slog.Logger and sets it as the global logger of the slog packages.
func setupGlobalLogger(w io.Writer, level string) *slog.Logger {
	opts := slog.HandlerOptions{
		AddSource: true,
	}

	switch level {
	case "info":
		opts.Level = slog.LevelInfo
	case "debug":
		opts.Level = slog.LevelDebug
	case "warning":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	}

	return slog.New(opts.NewTextHandler(w))
}

// readDBReplicaPassword reads the contents from db replica password file.
func readDBReplicaPassword(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}
