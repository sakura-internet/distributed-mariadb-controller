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

package controller

import (
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
)

// ControllerConfig is the configuration that is applied into Controller.
type ControllerConfig func(c *Controller)

func WithGlobalInterfaceName(globalInterfaceName string) ControllerConfig {
	return func(c *Controller) {
		c.globalInterfaceName = globalInterfaceName
	}
}

func WithHostAddress(hostAddress string) ControllerConfig {
	return func(c *Controller) {
		c.hostAddress = hostAddress
	}
}

func WithDBServingPort(dbServingPort uint16) ControllerConfig {
	return func(c *Controller) {
		c.dbServingPort = dbServingPort
	}
}

func WithDBReplicaUserName(dbReplicaUserName string) ControllerConfig {
	return func(c *Controller) {
		c.dbReplicaUserName = dbReplicaUserName
	}
}

func WithDBReplicaPassword(dbReplicaPassword string) ControllerConfig {
	return func(c *Controller) {
		c.dbReplicaPassword = dbReplicaPassword
	}
}

func WithDBReplicaSourcePort(dbReplicaSourcePort uint16) ControllerConfig {
	return func(c *Controller) {
		c.dbReplicaSourcePort = dbReplicaSourcePort
	}
}

func WithDBAclChainName(dbAclChainName string) ControllerConfig {
	return func(c *Controller) {
		c.dbAclChainName = dbAclChainName
	}
}

// WithSystemdConnector generates a config that sets the systemd.Connector into Controller.
func WithSystemdConnector(connector systemd.Connector) ControllerConfig {
	return func(c *Controller) {
		c.systemdConnector = connector
	}
}

// WithMariaDBConnector generates a config that sets the mariadb.Connector into Controller.
func WithMariaDBConnector(connector mariadb.Connector) ControllerConfig {
	return func(c *Controller) {
		c.mariaDBConnector = connector
	}
}

// WithNftablesConnector generates a config that sets the nftables.Connector into Controller.
func WithNftablesConnector(connector nftables.Connector) ControllerConfig {
	return func(c *Controller) {
		c.nftablesConnector = connector
	}
}

// WithBGPdConnector generates a config that sets the vtysh.WithBGPdConnector into Controller.
func WithBGPdConnector(connector bgpd.BGPdConnector) ControllerConfig {
	return func(c *Controller) {
		c.bgpdConnector = connector
	}
}
