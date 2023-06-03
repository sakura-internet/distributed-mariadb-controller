package sakura_test

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller/sakura"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestDecisionNextState_OnFault_WithPrimaryNeighbors(t *testing.T) {
	c := sakura.NewSAKURAController(slog.New(slog.NewTextHandler(os.Stderr)))
	c.SetState(controller.StateFault)

	ns := sakura.NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], sakura.Neighbor{})

	c.CurrentNeighbors = ns
	nextState := c.MakeDecision()
	assert.Equal(t, controller.StateReplica, nextState)
}

func TestDecisionNextState_OnFault_WithCandidateNeighbors(t *testing.T) {
	c := sakura.NewSAKURAController(slog.New(slog.NewTextHandler(os.Stderr)))
	c.SetState(controller.StateFault)

	ns := sakura.NewNeighborSet()
	ns.NeighborMatrix[sakura.SAKURAControllerStateCandidate] = append(ns.NeighborMatrix[sakura.SAKURAControllerStateCandidate], sakura.Neighbor{})
	c.CurrentNeighbors = ns

	nextState := c.MakeDecision()
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecisionNextState_OnFault_WithReplicaNeighbors(t *testing.T) {
	c := sakura.NewSAKURAController(slog.New(slog.NewTextHandler(os.Stderr)))
	c.SetState(controller.StateFault)

	ns := sakura.NewNeighborSet()
	ns.NeighborMatrix[controller.StateReplica] = append(ns.NeighborMatrix[controller.StateReplica], sakura.Neighbor{})
	c.CurrentNeighbors = ns

	nextState := c.MakeDecision()
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecisionNextState_OnFault_WithoutNeighbors(t *testing.T) {
	c := sakura.NewSAKURAController(slog.New(slog.NewTextHandler(os.Stderr)))
	c.SetState(controller.StateFault)

	ns := sakura.NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], sakura.Neighbor{})
	c.CurrentNeighbors = ns

	nextState := c.MakeDecision()
	assert.Equal(t, sakura.SAKURAControllerStateCandidate, nextState)
}
