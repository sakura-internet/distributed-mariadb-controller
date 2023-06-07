package sakura

import (
	"os"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/process"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"golang.org/x/exp/slog"
)

func _newFakeSAKURAController() *SAKURAController {
	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	c := NewSAKURAController(
		logger,
		SystemdConnector(systemd.NewFakeSystemdConnector()),
		MariaDBConnector(mariadb.NewFakeMariaDBConnector()),
		NftablesConnector(nftables.NewFakeNftablesConnector()),
		BGPdConnector(bgpd.NewFakeBGPdConnector()),
		ProcessControlConnector(process.NewFakeProcessControlConnector()),
	)

	c.HostAddress = "10.0.0.1"
	c.MariaDBReplicaPassword = "dummy-db-replica-password"
	return c
}
