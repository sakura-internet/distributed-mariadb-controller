package sakura

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
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

func TestMakeDecisionOnFault_WithPrimaryNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], Neighbor{})

	nextState := makeDecisionOnFault(ns)
	assert.Equal(t, controller.StateReplica, nextState)
}

func TestMakeDecisionOnFault_WithCandidateNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[SAKURAControllerStateCandidate] = append(ns.NeighborMatrix[SAKURAControllerStateCandidate], Neighbor{})

	nextState := makeDecisionOnFault(ns)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestMakeDecisionOnFault_WithReplicaNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateReplica] = append(ns.NeighborMatrix[controller.StateReplica], Neighbor{})

	nextState := makeDecisionOnFault(ns)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestMakeDecisionOnFault_WithoutNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], Neighbor{})

	nextState := makeDecisionOnFault(ns)
	assert.Equal(t, SAKURAControllerStateCandidate, nextState)
}
