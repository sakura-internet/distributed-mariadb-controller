package sakura

import (
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/stretchr/testify/assert"
)

func TestDecideNextStateOnReplica_MariaDBIsUnhealthy(t *testing.T) {
	ns := NewNeighborSet()
	nextState := decideNextStateOnReplica(ns, MariaDBHealthCheckResultNG)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecisionNextState_OnReplica_RemainReplica(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], Neighbor{})

	nextState := decideNextStateOnReplica(ns, MariaDBHealthCheckResultOK)
	assert.Equal(t, controller.StateReplica, nextState)
}

func TestDecisionNextState_OnReplica_NoOnePrimaryAndCandidate(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], Neighbor{})

	nextState := decideNextStateOnReplica(ns, MariaDBHealthCheckResultOK)
	assert.Equal(t, SAKURAControllerStateCandidate, nextState)
}

func TestTriggerRunOnStateChangesToPrimary_OKPath(t *testing.T) {
	c := _newFakeSAKURAController()

	// for checking the triggerRunOnStateChangesToPrimary() resets this count to 0
	c.writeTestDataFailCount = 5

	ns := NewNeighborSet()
	err := c.triggerRunOnStateChangesToPrimary(ns)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), c.writeTestDataFailCount)

	// test with MariaDB Connector
	fakeMariaDBConn := c.mariaDBConnector.(*mariadb.FakeMariaDBConnector)
	_, ok := fakeMariaDBConn.Timestamp["StopReplica"]
	assert.True(t, ok)

	t.Run("TestTriggerRunOnStateChangesToPrimary_OKPath_mustTurnOffMariaDBReadOnlyVariable", _mustTurnOffMariaDBReadOnlyVairable(fakeMariaDBConn))
}

func _mustTurnOffMariaDBReadOnlyVairable(conn *mariadb.FakeMariaDBConnector) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		_, ok := conn.Timestamp["StopReplica"]
		assert.True(t, ok)
		assert.False(t, conn.ReadOnlyVariable)
	}
}
