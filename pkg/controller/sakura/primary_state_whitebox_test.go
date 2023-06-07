package sakura

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
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

/*
func TestTriggerRunOnStateKeepsPrimary_WriteTestDataFailPath(t *testing.T) {
	c := _newFakeSAKURAController()
	// inject the mariadb connector that fails to write testdata.
	c.mariaDBConnector = testhelper.NewFakeMariaDBFailWriteTestDataConnector()

	ns := NewNeighborSet()
	err := c.triggerRunOnStateKeepsPrimary(ns)
	assert.NoError(t, err)

	// check the challenge count of writing test data is incremented.
	assert.Equal(t, uint(1), c.writeTestDataFailCount)
}

func TestTriggerRunOnStateKeepsPrimary_WriteTestDataFailedCountOversThreshold(t *testing.T) {
	c := _newFakeController("10.0.0.1", "dummy-db-replica-password")
	// inject the mariadb connector that fails to write testdata.
	c.mariadbConnector = testhelper.NewFakeMariaDBFailWriteTestDataConnector()
	c.writeTestDataFailCount = writeTestDataFailCountThreshold

	ns := NewNeighborSet()
	err := c.triggerRunOnStateKeepsPrimary(ns)
	assert.Error(t, err)
}
*/
