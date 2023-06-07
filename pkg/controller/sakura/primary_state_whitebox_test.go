package sakura

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestDecideNextStateOnPrimary_MariaDBIsUnhealthy(t *testing.T) {
	ns := NewNeighborSet()

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnPrimary(logger, ns, MariaDBHealthCheckResultNG)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnPrimary_InDualPrimarySituation(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnPrimary(logger, ns, MariaDBHealthCheckResultOK)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnPrimary_OKPath(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateReplica] = append(ns.NeighborMatrix[controller.StateReplica], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnPrimary(logger, ns, MariaDBHealthCheckResultOK)
	assert.Equal(t, controller.StatePrimary, nextState)
}

func TestTriggerRunOnStateKeepsPrimary_WriteTestDataFailPath(t *testing.T) {
	c := _newFakeSAKURAController()
	// inject the mariadb connector that fails to write testdata.
	c.mariaDBConnector = mariadb.NewFakeMariaDBFailWriteTestDataConnector()

	err := c.triggerRunOnStateKeepsPrimary()
	assert.NoError(t, err)

	// check the challenge count of writing test data is incremented.
	assert.Equal(t, uint(1), c.writeTestDataFailCount)
}

func TestTriggerRunOnStateKeepsPrimary_WriteTestDataFailedCountOversThreshold(t *testing.T) {
	c := _newFakeSAKURAController()
	// inject the mariadb connector that fails to write testdata.
	c.mariaDBConnector = mariadb.NewFakeMariaDBFailWriteTestDataConnector()
	c.writeTestDataFailCount = writeTestDataFailCountThreshold

	err := c.triggerRunOnStateKeepsPrimary()
	assert.Error(t, err)
}

func TestTriggerRunOnStateChangesToPrimary_OKPath(t *testing.T) {
	c := _newFakeSAKURAController()

	// for checking the triggerRunOnStateChangesToPrimary() resets this count to 0
	c.writeTestDataFailCount = 5

	err := c.triggerRunOnStateChangesToPrimary()
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
