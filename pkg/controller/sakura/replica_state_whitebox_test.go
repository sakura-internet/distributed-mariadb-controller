package sakura

import (
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
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
