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

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/stretchr/testify/assert"
)

func TestDecideNextStateOnPrimary_MariaDBIsUnhealthy(t *testing.T) {
	c := _newFakeController()
	c.currentMariaDBHealth = dbHealthCheckResultNG

	nextState := c.decideNextStateOnPrimary()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnPrimary_InDualPrimarySituation(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors[StatePrimary] = []neighbor{""}
	c.currentMariaDBHealth = dbHealthCheckResultOK

	nextState := c.decideNextStateOnPrimary()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnPrimary_OKPath(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors[StateReplica] = []neighbor{""}
	c.currentMariaDBHealth = dbHealthCheckResultOK

	nextState := c.decideNextStateOnPrimary()
	assert.Equal(t, StatePrimary, nextState)
}

func TestTriggerRunOnStateKeepsPrimary_WriteTestDataFailPath(t *testing.T) {
	c := _newFakeController()
	// inject the mariadb connector that fails to write testdata.
	c.mariaDBConnector = mariadb.NewFakeMariaDBFailWriteTestDataConnector()

	err := c.triggerRunOnStateKeepsPrimary()
	assert.NoError(t, err)

	// check the challenge count of writing test data is incremented.
	assert.Equal(t, uint(1), c.writeTestDataFailCount)
}

func TestTriggerRunOnStateKeepsPrimary_WriteTestDataFailedCountOversThreshold(t *testing.T) {
	c := _newFakeController()
	// inject the mariadb connector that fails to write testdata.
	c.mariaDBConnector = mariadb.NewFakeMariaDBFailWriteTestDataConnector()
	c.writeTestDataFailCount = writeTestDataFailCountThreshold

	err := c.triggerRunOnStateKeepsPrimary()
	assert.Error(t, err)
}

func TestTriggerRunOnStateChangesToPrimary_OKPath(t *testing.T) {
	c := _newFakeController()
	c.setState(StateCandidate)

	// for checking the triggerRunOnStateChangesToPrimary() resets this count to 0
	c.writeTestDataFailCount = 5

	err := c.triggerRunOnStateChangesToPrimary()
	assert.NoError(t, err)
	assert.Equal(t, uint(0), c.writeTestDataFailCount)

	// test with MariaDB Connector
	fakeMariaDBConn := c.mariaDBConnector.(*mariadb.FakeMariaDBConnector)
	_, ok := fakeMariaDBConn.Timestamp["StopReplica"]
	assert.True(t, ok)

	t.Run("TestTriggerRunOnStateChangesToPrimary_OKPath_mustTurnOffMariaDBReadOnlyVariable", _mustTurnOffMariaDBReadOnlyVairable(fakeMariaDBConn))
}

func _mustTurnOffMariaDBReadOnlyVairable(conn *mariadb.FakeMariaDBConnector) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		_, ok := conn.Timestamp["StopReplica"]
		assert.True(t, ok)
		assert.False(t, conn.ReadOnlyVariable)
	}
}
