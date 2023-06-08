package sakura

import (
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestDecideNextStateOnCandidate_MariaDBIsUnhealthy(t *testing.T) {
	ns := NewNeighborSet()
	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnCandidate(logger, ns, MariaDBHealthCheckResultNG, ReadytoPrimaryJudgeNG)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnCandidate_InMultiCandidateSituation(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[SAKURAControllerStateCandidate] = append(ns.NeighborMatrix[SAKURAControllerStateCandidate], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnCandidate(logger, ns, MariaDBHealthCheckResultOK, ReadytoPrimaryJudgeNG)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnCandidate_PrimaryIsAlreadyExist(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[SAKURAControllerStateCandidate], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnCandidate(logger, ns, MariaDBHealthCheckResultOK, ReadytoPrimaryJudgeNG)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnCandidate_ToBePromotedToPrimary(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnCandidate(logger, ns, MariaDBHealthCheckResultOK, ReadytoPrimaryJudgeOK)
	assert.Equal(t, controller.StatePrimary, nextState)
}

func TestDecideNextStateCandidate_RemainCandidate(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnCandidate(logger, ns, MariaDBHealthCheckResultOK, ReadytoPrimaryJudgeNG)
	assert.Equal(t, SAKURAControllerStateCandidate, nextState)
}

func TestTriggerRunOnStateChangesToCandidate_OKPath(t *testing.T) {
	c := _newFakeSAKURAController()

	err := c.triggerRunOnStateChangesToCandidate()
	assert.NoError(t, err)
}
