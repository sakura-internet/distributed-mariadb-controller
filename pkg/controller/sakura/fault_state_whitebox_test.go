package sakura

import (
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"github.com/stretchr/testify/assert"
)

func TestTriggerRunOnStateChangesToFault_OKPath(t *testing.T) {
	c := _newFakeSAKURAController()
	err := c.triggerRunOnStateChangesToFault()
	assert.NoError(t, err)

	// Systemd Connector test
	fakeSystemdConnector := c.systemdConnector.(*systemd.FakeSystemdConnector)
	// check whether the StartService() method is called with the "mariadb" name.
	started, ok := fakeSystemdConnector.ServiceStarted["mariadb"]
	assert.True(t, !ok || !started)
}

func TestDecideNextStateOnFault_WithPrimaryNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], Neighbor{})

	nextState := decideNextStateOnFault(ns)
	assert.Equal(t, controller.StateReplica, nextState)
}

func TestDecideNextStateOnFault_WithCandidateNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[SAKURAControllerStateCandidate] = append(ns.NeighborMatrix[SAKURAControllerStateCandidate], Neighbor{})

	nextState := decideNextStateOnFault(ns)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnFault_WithReplicaNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateReplica] = append(ns.NeighborMatrix[controller.StateReplica], Neighbor{})

	nextState := decideNextStateOnFault(ns)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnFault_WithoutNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], Neighbor{})

	nextState := decideNextStateOnFault(ns)
	assert.Equal(t, SAKURAControllerStateCandidate, nextState)
}
