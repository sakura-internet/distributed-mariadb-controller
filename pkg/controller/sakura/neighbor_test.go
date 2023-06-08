package sakura_test

import (
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller/sakura"
	"github.com/stretchr/testify/assert"
)

func TestDifferent_Same(t *testing.T) {
	a := sakura.NewNeighborSet()
	b := sakura.NewNeighborSet()
	assert.False(t, a.Different(b))
}

func TestDifferent_DiffLen(t *testing.T) {
	a := sakura.NewNeighborSet()
	b := sakura.NewNeighborSet()
	b.NeighborMatrix[controller.StateInitial] = append(b.NeighborMatrix[controller.StateInitial], sakura.Neighbor{})
	assert.True(t, a.Different(b))
}

func TestDifferent_DiffNeigh(t *testing.T) {
	a := sakura.NewNeighborSet()
	a.NeighborMatrix[controller.StateInitial] = append(a.NeighborMatrix[controller.StateInitial], sakura.Neighbor{})

	b := sakura.NewNeighborSet()
	b.NeighborMatrix[controller.StateInitial] = append(b.NeighborMatrix[controller.StateInitial], sakura.Neighbor{Address: "10.0.0.1"})
	assert.True(t, a.Different(b))

}
