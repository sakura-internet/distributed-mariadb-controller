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

package sakura

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
)

// NeighborSet holds the set of the BGP neighbors.
type NeighborSet struct {
	// Neighbors is the set of the BGP neighbors.
	NeighborMatrix map[controller.State][]Neighbor
}

// NewNeighborSet initializes the empty NeighborSet.
func NewNeighborSet() *NeighborSet {
	return &NeighborSet{
		NeighborMatrix: map[controller.State][]Neighbor{
			controller.StateFault:          make([]Neighbor, 0),
			SAKURAControllerStateCandidate: make([]Neighbor, 0),
			controller.StatePrimary:        make([]Neighbor, 0),
			controller.StateReplica:        make([]Neighbor, 0),
		},
	}
}

// Different returns true if the n and other is differenct.
func (n NeighborSet) Different(other *NeighborSet) bool {
	if len(n.NeighborMatrix) != len(other.NeighborMatrix) {
		return true
	}

	for k, nNeighbors := range n.NeighborMatrix {
		oNeighbors, ok := other.NeighborMatrix[k]
		if !ok {
			return true
		}

		if !reflect.DeepEqual(nNeighbors, oNeighbors) {
			return true
		}
	}

	return false
}

// NeighborAddresses construct the addresses of the neighbors into a string.
func (n NeighborSet) NeighborAddresses() string {
	addressesByState := []string{}

	for state, neighbors := range n.NeighborMatrix {
		addrs := make([]string, len(neighbors))
		for i, neighbor := range neighbors {
			addrs[i] = neighbor.Address
		}

		addressesByState = append(addressesByState, fmt.Sprintf("%s: [%s]", state, strings.Join(addrs, ",")))
	}

	return strings.Join(addressesByState, ", ")
}

// primaryNodeExists returns true if the set contains primary-state node(s).
func (n *NeighborSet) primaryNodeExists() bool {
	return len(n.NeighborMatrix[controller.StatePrimary]) != 0
}

// replicaNodeExists returns true if the set contains replica-state node(s).
func (n *NeighborSet) replicaNodeExists() bool {
	return len(n.NeighborMatrix[controller.StateReplica]) != 0
}

// candidateNodeExists returns true if the set contains candidate-state node(s).
func (n *NeighborSet) candidateNodeExists() bool {
	return len(n.NeighborMatrix[SAKURAControllerStateCandidate]) != 0
}

// faultNodeExists returns true if the set contains fault-state node(s).
func (n *NeighborSet) faultNodeExists() bool {
	return len(n.NeighborMatrix[controller.StateFault]) != 0
}

// anchorNodeExists returns true if the set contains anchor-mode node(s).
func (n *NeighborSet) anchorNodeExists() bool {
	return len(n.NeighborMatrix[SAKURAControllerStateAnchor]) != 0
}

// Neighbor is the BGP neighbor.
type Neighbor struct {
	// Address is the address of the BGP speaker.
	// e.g., "192.168.0.1"
	Address string
}
