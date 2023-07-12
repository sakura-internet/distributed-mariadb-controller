// Copyright 2023 The distributed-mariadb-controller Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
