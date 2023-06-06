package sakura

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestDecisionNextState_OnPrimary_MariaDBIsUnhealthy(t *testing.T) {
	ns := NewNeighborSet()

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := makeDecisionNextStateOnPrimary(logger, ns, MariaDBHealthCheckResultNG)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecisionNextState_OnPrimary_InDualPrimarySituation(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := makeDecisionNextStateOnPrimary(logger, ns, MariaDBHealthCheckResultOK)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecisionNextState_OnPrimary_OKPath(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateReplica] = append(ns.NeighborMatrix[controller.StateReplica], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := makeDecisionNextStateOnPrimary(logger, ns, MariaDBHealthCheckResultOK)
	assert.Equal(t, controller.StatePrimary, nextState)
}
