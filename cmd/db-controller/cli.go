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
	"flag"
	"fmt"
	"strings"
)

var (
	// logLevelFlag is a cli-flag that specifies the log level on the db-controller.
	logLevelFlag string
	// lockFilePathFlag is a cli-flag that specifies the filepath of the exclusive lock.
	lockFilePathFlag string
	// dbServingPortFlag is a cli-flag that specifies portnumber of database service
	dbServingPortFlag int
	// dbReplicaUserNameFlag is a cli-flag that specifies the username for replication
	dbReplicaUserNameFlag string
	// dbReplicaPasswordFilePathFlag is a cli-flag that specifies the filepath of the DB replica password.
	dbReplicaPasswordFilePathFlag string
	// globalInterfaceNameFlag is a cli-flag that specifies the global network interface for get my IPaddress.
	globalInterfaceNameFlag string
	// chainNameForDBAclFlag is a cli-flag that specifies the nftables chain name for DB access control list.
	chainNameForDBAclFlag string
	// bgpAsNumberFlag is a cli-flag that specifies the as number of ours
	bgpAsNumberFlag int
	// bgpServingPortFlag is a cli-flag that specifies the port of bgp
	bgpServingPortFlag int
	// bgpKeepaliveIntervalSecFlag is a cli-flag that specifies the interval seconds of bgp keepalive
	bgpKeepaliveIntervalSecFlag int
	// bgpPeerAddressesFlag is a cli-flag that specifies comma sparated peers of bgp.
	bgpPeerAddressesFlag string
	// gobgpGrpcPortFlag is a cli-flag that specifies port of gobgp gRPC
	gobgpGrpcPortFlag int

	// mainPollingSpanSecondFlag is a cli-flag that specifies the span seconds of the loop in main.go.
	mainPollingSpanSecondFlag int
	// httpAPIServerPortFlag is a cli-flag that specifies the port the HTTP API server listens.
	httpAPIServerPortFlag int
	// prometheusExporterPortFlag is a cli-flag that specifies the port the prometheus exporter listens.
	prometheusExporterPortFlag int
	// dbReplicaSourcePortFlag is a cli-flag that specifies the port of primary as replication source.
	dbReplicaSourcePortFlag int

	// enablePrometheusExporterFlag is a cli-flag that enables the prometheus exporter.
	enablePrometheusExporterFlag bool
	// enableHTTPAPIFlag is a cli-flag that enables the http api server.
	enableHTTPAPIFlag bool
)

// ParseAllFlags parses all defined cmd-flags.
func parseAllFlags(args []string) error {
	fs := flag.NewFlagSet("db-controller", flag.PanicOnError)

	fs.StringVar(&logLevelFlag, "log-level", "warning", "the log level(debug/info/warning/error)")
	fs.StringVar(&lockFilePathFlag, "lock-filepath", "/var/run/db-controller/lock", "the filepath of the exclusive lock")
	fs.StringVar(&dbReplicaPasswordFilePathFlag, "db-repilica-password-filepath", "/var/run/db-controller/.db-replica-password", "the filepath of the DB replica password")
	fs.StringVar(&globalInterfaceNameFlag, "global-interface-name", "eth0", "the interface name of global")
	fs.StringVar(&chainNameForDBAclFlag, "chain-name-for-db-acl", "mariadb", "the chain name for DB access control")
	fs.StringVar(&dbReplicaUserNameFlag, "db-replica-user-name", "repl", "the username for replication")
	fs.StringVar(&bgpPeerAddressesFlag, "bgp-peer-addresses", "", "the peers of bgp")

	fs.IntVar(&mainPollingSpanSecondFlag, "main-polling-span-second", 4, "the span seconds of the loop in main.go")
	fs.IntVar(&httpAPIServerPortFlag, "http-api-server-port", 54545, "the port the http api server listens")
	fs.IntVar(&prometheusExporterPortFlag, "prometheus-exporter-port", 50505, "the port the prometheus exporter listens")
	fs.IntVar(&dbReplicaSourcePortFlag, "db-replica-source-port", 13306, "the port of primary as replication source")
	fs.IntVar(&dbServingPortFlag, "db-serving-port", 3306, "the port of database service")
	fs.IntVar(&bgpAsNumberFlag, "bgp-as-number", 65001, "the as number of ours")
	fs.IntVar(&bgpServingPortFlag, "bgp-serving-port", 179, "the port of bgp")
	fs.IntVar(&bgpKeepaliveIntervalSecFlag, "bgp-keepalive-interval-sec", 3, "the interval seconds of bgp keepalive")
	fs.IntVar(&gobgpGrpcPortFlag, "gobgp-grpc-port", 50051, "the listen port of gobgp gRPC")

	fs.BoolVar(&enablePrometheusExporterFlag, "prometheus-exporter", true, "enables the prometheus exporter")
	fs.BoolVar(&enableHTTPAPIFlag, "http-api", true, "enables the http api server")

	return fs.Parse(args)
}

// ValidateAllFlags validates all cmd flags.
func validateAllFlags() error {
	if !isValidLogLevelFlag(logLevelFlag) {
		return fmt.Errorf("--log-level must be one of debug/info/warning/error")
	}

	if prometheusExporterPortFlag < 0 || 65535 < prometheusExporterPortFlag {
		return fmt.Errorf("--prometheus-exporter-port must be the range of uint16(tcp port)")
	}

	peers := strings.Split(bgpPeerAddressesFlag, ",")
	if len(peers) != 2 {
		return fmt.Errorf("insufficient bgp peer addresses: %s", bgpPeerAddressesFlag)
	}

	return nil
}

func isValidLogLevelFlag(l string) bool {
	return l == "debug" || l == "info" || l == "warning" || l == "error"
}
