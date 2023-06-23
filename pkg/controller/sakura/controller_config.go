package sakura

import (
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/process"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
)

// ControllerConfig is the configuration that is applied into SAKURAController.
type ControllerConfig func(c *SAKURAController)

// SystemdConnector generates a config that sets the systemd.Connector into SAKURAController.
func SystemdConnector(connector systemd.Connector) ControllerConfig {
	return func(c *SAKURAController) {
		c.systemdConnector = connector
	}
}

func MariaDBConnector(connector mariadb.Connector) ControllerConfig {
	return func(c *SAKURAController) {
		c.mariaDBConnector = connector
	}
}

// NftablesConnector generates a config that sets the nftables.Connector into SAKURAController.
func NftablesConnector(connector nftables.Connector) ControllerConfig {
	return func(c *SAKURAController) {
		c.NftablesConnector = connector
	}
}

// BGPdConnector generates a config that sets the vtysh.BGPdConnector into SAKURAController.
func BGPdConnector(connector bgpd.BGPdConnector) ControllerConfig {
	return func(c *SAKURAController) {
		c.bgpdConnector = connector
	}
}

// ProcessControlConnector generates a config that sets the process.ProcessControlConnector into SAKURAController.
func ProcessControlConnector(connector process.ProcessControlConnector) ControllerConfig {
	return func(c *SAKURAController) {
		c.processControlConnector = connector
	}
}
