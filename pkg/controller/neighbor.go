// Copyright 2025 The distributed-mariadb-controller Authors
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

package controller

import (
	"fmt"
	"reflect"
	"strings"
)

// neighbor is the BGP neighbor.
type neighbor struct {
	// address is the address of the BGP speaker.
	// e.g., "192.168.0.1"
	address string
}

// neighborSet holds the set of the BGP neighbors.
type neighborSet struct {
	// Neighbors is the set of the BGP neighbors.
	neighborMatrix map[State][]neighbor
}

// newNeighborSet initializes the empty NeighborSet.
func newNeighborSet() *neighborSet {
	return &neighborSet{
		neighborMatrix: map[State][]neighbor{
			StateFault:     make([]neighbor, 0),
			StateCandidate: make([]neighbor, 0),
			StatePrimary:   make([]neighbor, 0),
			StateReplica:   make([]neighbor, 0),
		},
	}
}

// different returns true if the n and other is differenct.
func (n neighborSet) different(other *neighborSet) bool {
	if len(n.neighborMatrix) != len(other.neighborMatrix) {
		return true
	}

	for k, nNeighbors := range n.neighborMatrix {
		oNeighbors, ok := other.neighborMatrix[k]
		if !ok {
			return true
		}

		if !reflect.DeepEqual(nNeighbors, oNeighbors) {
			return true
		}
	}

	return false
}

// neighborAddresses construct the addresses of the neighbors into a string.
func (n neighborSet) neighborAddresses() string {
	addressesByState := []string{}

	for state, neighbors := range n.neighborMatrix {
		addrs := make([]string, len(neighbors))
		for i, neighbor := range neighbors {
			addrs[i] = neighbor.address
		}

		addressesByState = append(addressesByState, fmt.Sprintf("%s: [%s]", state, strings.Join(addrs, ",")))
	}

	return strings.Join(addressesByState, ", ")
}

// primaryNodeExists returns true if the set contains primary-state node(s).
func (n *neighborSet) primaryNodeExists() bool {
	return len(n.neighborMatrix[StatePrimary]) != 0
}

// replicaNodeExists returns true if the set contains replica-state node(s).
func (n *neighborSet) replicaNodeExists() bool {
	return len(n.neighborMatrix[StateReplica]) != 0
}

// candidateNodeExists returns true if the set contains candidate-state node(s).
func (n *neighborSet) candidateNodeExists() bool {
	return len(n.neighborMatrix[StateCandidate]) != 0
}

// faultNodeExists returns true if the set contains fault-state node(s).
func (n *neighborSet) faultNodeExists() bool {
	return len(n.neighborMatrix[StateFault]) != 0
}

// anchorNodeExists returns true if the set contains anchor-mode node(s).
func (n *neighborSet) anchorNodeExists() bool {
	return len(n.neighborMatrix[StateAnchor]) != 0
}

// isNetworkParted returns true if there is no neighbor on the network.
func (n *neighborSet) isNetworkParted() bool {
	if n.primaryNodeExists() ||
		n.candidateNodeExists() ||
		n.replicaNodeExists() ||
		n.faultNodeExists() ||
		n.anchorNodeExists() {
		return false
	}

	return true
}
