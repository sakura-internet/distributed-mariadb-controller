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

// SystemdConnector generates a config that sets the systemd.Connector into Controller.
func SystemdConnector(connector systemd.Connector) ControllerConfig {
	return func(c *Controller) {
		c.systemdConnector = connector
	}
}

func MariaDBConnector(connector mariadb.Connector) ControllerConfig {
	return func(c *Controller) {
		c.mariaDBConnector = connector
	}
}

// NftablesConnector generates a config that sets the nftables.Connector into Controller.
func NftablesConnector(connector nftables.Connector) ControllerConfig {
	return func(c *Controller) {
		c.nftablesConnector = connector
	}
}

// BGPdConnector generates a config that sets the vtysh.BGPdConnector into Controller.
func BGPdConnector(connector bgpd.BGPdConnector) ControllerConfig {
	return func(c *Controller) {
		c.bgpdConnector = connector
	}
}
