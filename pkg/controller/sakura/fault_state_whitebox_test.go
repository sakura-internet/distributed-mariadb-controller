package sakura

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/vtysh"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/process"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestTriggerRunOnStateChangesToFault_OKPath(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr))

	nftablesConnector := nftables.NewFakeNftablesConnector()
	systemdConnector := systemd.NewFakeSystemdConnector()
	bgpdConnector := vtysh.NewFakeBGPdConnector()
	processControlConnector := process.NewFakeProcessControlConnector()
	c := NewSAKURAController(
		logger,
		NftablesConnector(nftablesConnector),
		SystemdConnector(systemdConnector),
		BGPdConnector(bgpdConnector),
		ProcessControlConnector(processControlConnector),
	)

	err := c.triggerRunOnStateChangesToFault()
	assert.NoError(t, err)

	// Systemd Connector test
	fakeSystemdConnector := c.systemdConnector.(*systemd.FakeSystemdConnector)
	// check whether the StartService() method is called with the "mariadb" name.
	started, ok := fakeSystemdConnector.ServiceStarted["mariadb"]
	assert.True(t, !ok || !started)
}
