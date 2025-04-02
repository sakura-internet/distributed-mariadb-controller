// Copyright 2025 The distributed-mariadb-controller Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	apiv0 "github.com/sakura-internet/distributed-mariadb-controller/cmd/db-controller/api/v0"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/vishvananda/netlink"
)

func main() {
	if err := parseAllFlags(os.Args[1:]); err != nil {
		panic(err)
	}
	if err := validateAllFlags(); err != nil {
		panic(err)
	}

	logger := setupGlobalLogger(os.Stderr, logLevelFlag)

	// mkdir for lock file
	if err := os.MkdirAll(filepath.Dir(lockFilePathFlag), 0o755); err != nil {
		panic(err)
	}

	lockf, err := tryToGetTheExclusiveLockWithoutBlocking(lockFilePathFlag)
	if err != nil {
		panic(err)
	}
	defer lockf.Close()

	// for controlling the traffics that they're to the DB server port.
	// the function returns nil if the expected chain is already exist.
	nftConnect := nftables.NewDefaultConnector(logger)
	if err := nftConnect.CreateChain(chainNameForDBAclFlag); err != nil {
		panic(err)
	}

	// prepare controller instance
	myHostAddress, err := getNetIFAddress(globalInterfaceNameFlag)
	if err != nil {
		panic(err)
	}
	logger.Debug("host address", "address", myHostAddress)

	dbReplicaPassword, err := readDBReplicaPassword(dbReplicaPasswordFilePathFlag)
	if err != nil {
		panic(err)
	}

	c := controller.NewController(
		logger,
		controller.WithGlobalInterfaceName(globalInterfaceNameFlag),
		controller.WithHostAddress(myHostAddress),
		controller.WithDBServingPort(uint16(dbServingPortFlag)),
		controller.WithDBReplicaUserName(dbReplicaUserNameFlag),
		controller.WithDBReplicaPassword(dbReplicaPassword),
		controller.WithDBReplicaSourcePort(uint16(dbReplicaSourcePortFlag)),
		controller.WithDBAclChainName(chainNameForDBAclFlag),
	)

	// start goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, c *controller.Controller) {
		defer wg.Done()
		c.Start(ctx, time.Second*time.Duration(mainPollingSpanSecondFlag))
	}(ctx, wg, c)

	if enablePrometheusExporterFlag {
		wg.Add(1)
		go startPrometheusExporterServer(ctx, wg)
	}

	if enableHTTPAPIFlag {
		wg.Add(1)
		go startHTTPAPIServer(ctx, wg, c)
	}

	// wait for receive signal
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	stopSigCh := make(chan os.Signal, 3)
	signal.Notify(stopSigCh, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	<-stopSigCh
	logger.Info("got stop signal. exiting.")

	// for stopping all goroutine.
	cancel()
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

	switch logLevelFlag {
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

	reg := controller.NewPrometheusMetricRegistry()
	e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	// Start server
	addr := fmt.Sprintf(":%d", prometheusExporterPortFlag)

	ch := make(chan bool, 1)
	go func(ch chan<- bool) {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}

		ch <- true
	}(ch)

	<-ctx.Done()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	<-ch
}

// startHTTPAPIServer starts the HTTP API server that serves the controller status responder.
func startHTTPAPIServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	c *controller.Controller,
) {
	defer wg.Done()

	// Setup
	e := echo.New()
	e.Use(apiv0.UseControllerState(c))

	switch logLevelFlag {
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
	addr := fmt.Sprintf(":%d", httpAPIServerPortFlag)

	ch := make(chan bool, 1)
	go func(ch chan<- bool) {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}

		ch <- true
	}(ch)

	<-ctx.Done()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	<-ch
}

// getNetIFAddress tries to get the IP address of the specified interface name I/F using Netlink messages.
func getNetIFAddress(intfname string) (string, error) {
	eth, err := netlink.LinkByName(intfname)
	if err != nil {
		return "", err
	}

	addrs, err := netlink.AddrList(eth, netlink.FAMILY_V4)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("%s doesn't have any IP addresses", intfname)
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
	opts := &slog.HandlerOptions{
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

	return slog.New(slog.NewTextHandler(w, opts))
}

// readDBReplicaPassword reads the contents from db replica password file.
func readDBReplicaPassword(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}
