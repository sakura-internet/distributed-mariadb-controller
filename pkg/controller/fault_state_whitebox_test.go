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
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"github.com/stretchr/testify/assert"
)

func TestTriggerRunOnStateChangesToFault_OKPath(t *testing.T) {
	c := _newFakeController()
	err := c.triggerRunOnStateChangesToFault()
	assert.NoError(t, err)

	// Systemd Connector test
	fakeSystemdConnector := c.systemdConnector.(*systemd.FakeSystemdConnector)
	// check whether the StartService() method is called with the "mariadb" name.
	started, ok := fakeSystemdConnector.ServiceStarted["mariadb"]
	assert.True(t, !ok || !started)
}

func TestDecideNextStateOnFault_WithPrimaryNeighbors(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StatePrimary] = []neighbor{{}}

	nextState := c.decideNextStateOnFault()
	assert.Equal(t, StateReplica, nextState)
}

func TestDecideNextStateOnFault_WithCandidateNeighbors(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StateCandidate] = []neighbor{{}}

	nextState := c.decideNextStateOnFault()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnFault_WithReplicaNeighbors(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StateReplica] = []neighbor{{}}

	nextState := c.decideNextStateOnFault()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnFault_WithoutNeighbors(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StateFault] = []neighbor{{}}

	nextState := c.decideNextStateOnFault()
	assert.Equal(t, StateCandidate, nextState)
}
