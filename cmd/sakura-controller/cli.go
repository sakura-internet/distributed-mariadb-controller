package main

import (
	"flag"
	"fmt"
)

var (
	// LogLevelFlag is a cli-flag that specifies the log level on the db-controller.
	LogLevelFlag string
	// LockFilePathFlag is a cli-flag that specifies the filepath of the exclusive lock.
	LockFilePathFlag string
	// DBReplicaPasswordFilePathFlag is a cli-flag that specifies the filepath of the DB replica password.
	DBReplicaPasswordFilePathFlag string

	// MainPollingSpanSecondFlag is a cli-flag that specifies the span seconds of the loop in main.go.
	MainPollingSpanSecondFlag int
	// HTTPAPIServerPortFlag is a cli-flag that specifies the port the HTTP API server listens.
	HTTPAPIServerPortFlag int
	// PrometheusExporterPortFlag is a cli-flag that specifies the port the prometheus exporter listens.
	PrometheusExporterPortFlag int

	// EnablePrometheusExporterFlag is a cli-flag that enables the prometheus exporter.
	EnablePrometheusExporterFlag bool
	// EnableHTTPAPIFlag is a cli-flag that enables the http api server.
	EnableHTTPAPIFlag bool
)

// ParseAllFlags parses all defined cmd-flags.
func parseAllFlags(args []string) error {
	fs := flag.NewFlagSet("db-controller", flag.PanicOnError)

	fs.StringVar(&LockFilePathFlag, "lock-filepath", "/var/run/db-controller/lock", "the filepath of the exclusive lock")
	fs.StringVar(&DBReplicaPasswordFilePathFlag, "db-repilica-password-filepath", "/var/run/db-controller/.db-replica-password", "the filepath of the DB replica password")
	fs.StringVar(&LogLevelFlag, "log-level", "warning", "the log level(debug/info/warning/error)")

	fs.IntVar(&MainPollingSpanSecondFlag, "main-polling-span-second", 4, "the span seconds of the loop in main.go")
	fs.IntVar(&PrometheusExporterPortFlag, "prometheus-exporter-port", 50505, "the port the prometheus exporter listens")
	fs.IntVar(&HTTPAPIServerPortFlag, "http-api-server-port", 54545, "the port the http api server listens")

	fs.BoolVar(&EnablePrometheusExporterFlag, "prometheus-exporter", true, "enables the prometheus exporter")
	fs.BoolVar(&EnableHTTPAPIFlag, "http-api", true, "enables the http api server")

	return fs.Parse(args)
}

// ValidateAllFlags validates all cmd flags.
func validateAllFlags() error {
	if invalidLogLevelFlag(LogLevelFlag) {
		return fmt.Errorf("--log-level must be one of debug/info/warning/error")
	}

	if PrometheusExporterPortFlag < 0 || 65535 < PrometheusExporterPortFlag {
		return fmt.Errorf("--prometheus-exporter-port must be the range of uint16(tcp port)")
	}

	return nil
}

func invalidLogLevelFlag(l string) bool {
	valid := l == "debug" || l == "info" || l == "warning" || l == "error"
	return !valid
}
