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
	"os"
	"testing"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestTriggerRunOnStateChangesToFault_OKPath(t *testing.T) {
	c := _newFakeSAKURAController()
	err := c.triggerRunOnStateChangesToFault()
	assert.NoError(t, err)

	// Systemd Connector test
	fakeSystemdConnector := c.systemdConnector.(*systemd.FakeSystemdConnector)
	// check whether the StartService() method is called with the "mariadb" name.
	started, ok := fakeSystemdConnector.ServiceStarted["mariadb"]
	assert.True(t, !ok || !started)
}

func TestDecideNextStateOnFault_WithPrimaryNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StatePrimary] = append(ns.NeighborMatrix[controller.StatePrimary], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnFault(logger, ns)
	assert.Equal(t, controller.StateReplica, nextState)
}

func TestDecideNextStateOnFault_WithCandidateNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[SAKURAControllerStateCandidate] = append(ns.NeighborMatrix[SAKURAControllerStateCandidate], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnFault(logger, ns)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnFault_WithReplicaNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateReplica] = append(ns.NeighborMatrix[controller.StateReplica], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnFault(logger, ns)
	assert.Equal(t, controller.StateFault, nextState)
}

func TestDecideNextStateOnFault_WithoutNeighbors(t *testing.T) {
	ns := NewNeighborSet()
	ns.NeighborMatrix[controller.StateFault] = append(ns.NeighborMatrix[controller.StateFault], Neighbor{})

	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	nextState := decideNextStateOnFault(logger, ns)
	assert.Equal(t, SAKURAControllerStateCandidate, nextState)
}
