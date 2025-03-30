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

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"github.com/stretchr/testify/assert"
)

func TestDecideNextStateOnReplica_MariaDBIsUnhealthy(t *testing.T) {
	c := _newFakeController()
	c.currentMariaDBHealth = dbHealthCheckResultNG

	nextState := c.decideNextStateOnReplica()
	assert.Equal(t, StateFault, nextState)
}

func TestDecisionNextState_OnReplica_RemainReplica(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors[StatePrimary] = []neighbor{""}
	c.currentMariaDBHealth = dbHealthCheckResultOK

	nextState := c.decideNextStateOnReplica()
	assert.Equal(t, StateReplica, nextState)
}

func TestDecisionNextState_OnReplica_NoOnePrimaryAndCandidate(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors[StateFault] = []neighbor{""}
	c.currentMariaDBHealth = dbHealthCheckResultOK

	nextState := c.decideNextStateOnReplica()
	assert.Equal(t, StateCandidate, nextState)
}

func TestTriggerRunOnStateChangesToReplica_OKPath(t *testing.T) {
	c := _newFakeController()

	// for checking the triggerRunOnStateChangesToReplica() resets this count to 0
	c.replicationStatusCheckFailCount = 5

	primaryNeighbor := neighbor("10.0.0.2")
	ns := newNeighborSet()
	ns[StatePrimary] = append(ns[StatePrimary], primaryNeighbor)
	c.currentNeighbors = ns

	err := c.triggerRunOnStateChangesToReplica()
	assert.NoError(t, err)

	// test with MariaDB Connector
	fakeMariaDBConn := c.mariaDBConnector.(*mariadb.FakeMariaDBConnector)
	t.Run("TestTriggerRunOnStateChangesToReplica_OKPath_shouldResetReplicationStatusCheckCount", _shouldResetReplicationStatusCheckCount(c))
	t.Run("TestTriggerRunOnStateChangesToReplica_OKPath_mustTurnOnMariaDBReadOnlyVariable", _mustTurnOnMariaDBReadOnlyVariable(fakeMariaDBConn))
	t.Run("TestTriggerRunOnStateChangesToReplica_OKPath_shouldBeCorrectReplicationCommandsExecutionOrder", _shouldBeCorrectReplicationCommandsExecutionOrder(fakeMariaDBConn))
	t.Run("TestTriggerRunOnStateChangesToReplica_OKPath_mustCallChangeMasterToWithCorrectArgs", _mustCallChangeMasterToWithCorrectArgs(fakeMariaDBConn, string(primaryNeighbor), "dummy-db-replica-password"))

	// test with Nftables Connector
	fakeNftablesConn := c.nftablesConnector.(*nftables.FakeNftablesConnector)
	t.Run("TestTriggerRunOnStateChangesToReplica_OKPath_shouldBeCorrectNftablesRejectTCP3306TrafficCommandsOrder", _shouldBeCorrectNftablesRejectTCP3306TrafficCommandsOrder(fakeNftablesConn))

	// Systemd Connector test
	fakeSystemdConnector := c.systemdConnector.(*systemd.FakeSystemdConnector)
	t.Run("TestTriggerRunOnStateChangesToReplica_OKPath_mustStartMariaDBService", _mustStartMariaDBService(fakeSystemdConnector))
}

func TestTriggerRunOnStateKeepsReplica_CheckReplicationStatusFailPath(t *testing.T) {
	c := _newFakeController()
	// inject the mariadb connector that fails to check replication status.
	c.mariaDBConnector = mariadb.NewFakeMariaDBFailedReplicationConnector()

	{
		ns := newNeighborSet()
		ns[StatePrimary] = []neighbor{""}
		c.currentNeighbors = ns
	}
	err := c.triggerRunOnStateKeepsReplica()
	assert.NoError(t, err)

	// check the challenge count of checking replication status is incremented.
	assert.Equal(t, uint(1), c.replicationStatusCheckFailCount)

	// advertiseSelfNetIFAddress() must not be called
	fakeBGPdConnector := c.bgpdConnector.(*bgpd.FakeBGPdConnector)
	_, ok := fakeBGPdConnector.RouteConfigured["10.0.0.1"]
	assert.False(t, ok)
}

func TestTriggerRunOnStateKeepsReplica_ReplicationStatusCheckCountOversThreshold(t *testing.T) {
	c := _newFakeController()
	// inject the mariadb connector that fails to check replication status.
	c.mariaDBConnector = mariadb.NewFakeMariaDBFailedReplicationConnector()

	c.replicationStatusCheckFailCount = replicationStatusCheckThreshold

	{
		ns := newNeighborSet()
		ns[StatePrimary] = []neighbor{""}
		c.currentNeighbors = ns
	}
	err := c.triggerRunOnStateKeepsReplica()
	assert.Error(t, err)
}

func _shouldResetReplicationStatusCheckCount(c *Controller) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// check the count reset to 0 if triggerRunOnStateChangesToReplica() succeeded
		assert.Equal(t, uint(0), c.replicationStatusCheckFailCount)
	}
}

func _mustTurnOnMariaDBReadOnlyVariable(
	conn *mariadb.FakeMariaDBConnector,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// check whether the controller turn on the read_only variable
		_, ok := conn.Timestamp["TurnOnReadOnly"]
		assert.True(t, ok)
	}
}

func _shouldBeCorrectReplicationCommandsExecutionOrder(
	conn *mariadb.FakeMariaDBConnector,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// check whether the execution order of the important commands is correct.
		stopTime := conn.Timestamp["StopReplica"]
		changeMasterTime := conn.Timestamp["ChangeMasterTo"]
		restartTime := conn.Timestamp["StartReplica"]
		assert.True(t, stopTime.Before(restartTime))
		assert.True(t, changeMasterTime.Before(restartTime))
	}
}

func _mustCallChangeMasterToWithCorrectArgs(
	conn *mariadb.FakeMariaDBConnector,
	expectedPrimaryAddress string,
	expectedPassword string,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// check whether the ChangeMasterTo() is called with the properties of a primary neighbor.
		// FakeMariaDBConnector holds the argument of ChangeMasterTo() to .MasterConfig directly.
		expectedReplicaUserName := "repl"
		assert.Equal(t, expectedPrimaryAddress, conn.MasterConfig.Host)
		assert.Equal(t, expectedReplicaUserName, conn.MasterConfig.User)
		assert.Equal(t, expectedPassword, conn.MasterConfig.Password)
		assert.Equal(t, mariadb.MasterUseGTIDValueCurrentPos, conn.MasterConfig.UseGTID)
	}
}

func _shouldBeCorrectNftablesRejectTCP3306TrafficCommandsOrder(
	conn *nftables.FakeNftablesConnector,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// check whether the FlushChain() is called.
		flushTime, ok := conn.Timestamp["FlushChain"]
		assert.True(t, ok)

		// check whether the AddRule() is called.
		addRuleTime, ok := conn.Timestamp["AddRule"]
		assert.True(t, ok)

		// check whether the AddRule() is called after calling FlushChain().
		assert.True(t, flushTime.Before(addRuleTime))
	}
}

func _mustStartMariaDBService(
	conn *systemd.FakeSystemdConnector,
) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// check whether the StartService() method is called with the "mariadb" name.
		started, ok := conn.ServiceStarted["mariadb"]
		assert.True(t, ok)
		assert.True(t, started)
	}
}
