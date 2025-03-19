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

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
)

const (
	// replicationStatusCheckThreshold is a threshold.
	// if the counter of the controller overs this, the controller goes panic.
	replicationStatusCheckThreshold = 20
)

// decideNextStateOnReplica determines the next state on replica state.
func (c *Controller) decideNextStateOnReplica() State {
	if c.currentMariaDBHealth == dbHealthCheckResultNG {
		return StateFault
	}

	noPrimary := !c.currentNeighbors.primaryNodeExists()
	noCandidate := !c.currentNeighbors.candidateNodeExists()
	if noPrimary && noCandidate {
		// you may be the next primary node!
		return StateCandidate
	}

	return StateReplica
}

// triggerRunOnStateChangesToReplica transition to replica state in main loop.
func (c *Controller) triggerRunOnStateChangesToReplica() error {
	// [STEP1]: setting MariaDB State.
	if err := c.startMariaDBService(); err != nil {
		return err
	}
	if health := c.checkMariaDBHealth(); health == dbHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}

	if !c.currentNeighbors.primaryNodeExists() {
		return fmt.Errorf("there is no primary neighbor in replica mode")
	}

	if err := c.syncReadOnlyVariable( /* read_only=1 */ true); err != nil {
		return err
	}
	if err := c.mariaDBConnector.StopReplica(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.ResetAllReplicas(); err != nil {
		return err
	}

	primaryNode := c.currentNeighbors.neighborMatrix[StatePrimary][0]
	master := mariadb.MasterInstance{
		Host:     primaryNode.address,
		Port:     c.dbReplicaSourcePort,
		User:     c.dbReplicaUserName,
		Password: c.dbReplicaPassword,
		UseGTID:  mariadb.MasterUseGTIDValueCurrentPos,
	}
	if err := c.mariaDBConnector.ChangeMasterTo(master); err != nil {
		return err
	}

	if err := c.mariaDBConnector.StartReplica(); err != nil {
		return err
	}

	// [STEP2]: setting Nftables State.
	if err := c.rejectDatabaseServiceTraffic(); err != nil {
		return err
	}

	// [STEP3]: configurating frrouting.
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		return err
	}

	// reset the count because the controller is healthy for replica mode.
	c.replicationStatusCheckFailCount = 0

	c.logger.Info("replica state handler succeed")
	return nil
}

func (c *Controller) triggerRunOnStateKeepsReplica() error {
	if c.replicationStatusCheckFailCount >= replicationStatusCheckThreshold {
		// we should manually operate the case for recovering.
		return fmt.Errorf("reached the maximum retry limit for replication")
	}

	if err := c.checkMariaDBReplicationStatus(); err != nil {
		// we should keep trying to challenge that the replication status satisfies our conditions.
		c.replicationStatusCheckFailCount++
		c.logger.Warn("failed to satisfy replication conditions", "error", err, "replicationCount", c.replicationStatusCheckFailCount)

		if err := c.restartMariaDBReplica(); err != nil {
			c.logger.Warn("failed to restart replica", "error", err)
		}

		// return noerror because this is soft fail
		return nil
	}

	// reset the count because the controller is healthy.
	c.replicationStatusCheckFailCount = 0

	return nil
}

// checkMariaDBReplicationStatus returns true if the status of replication is satisfied.
// if the challenge failed to satisfy the conditions, this function returns false.
func (c *Controller) checkMariaDBReplicationStatus() error {
	status, err := c.mariaDBConnector.ShowReplicationStatus()
	if err != nil {
		return err
	}

	if !c.checkRequiredReplicationStatusIsOK(status) {
		return fmt.Errorf("failed to satisfy the replication conditions")
	}

	return nil
}

// checkRequiredReplicationStatusIsOK checks the replication status satisfies the required conditions.
func (c *Controller) checkRequiredReplicationStatusIsOK(status mariadb.ReplicationStatus) bool {
	ioRunning, ok1 := status[mariadb.ReplicationStatusSlaveIORunning]
	sqlRunning, ok2 := status[mariadb.ReplicationStatusSlaveSQLRunning]

	if !ok1 || !ok2 {
		msg := fmt.Sprintf("failed to retrieve %s or %s",
			mariadb.ReplicationStatusSlaveIORunning,
			mariadb.ReplicationStatusSlaveSQLRunning)
		c.logger.Debug(msg)
		return false
	}

	if ioRunning != mariadb.ReplicationStatusSlaveIORunningYes {
		msg := fmt.Sprintf("unexpected %s status", mariadb.ReplicationStatusSlaveIORunning)
		c.logger.Debug(msg, "expected", "Yes", "actual", ioRunning)
		return false
	}

	if sqlRunning != mariadb.ReplicationStatusSlaveSQLRunningYes {
		msg := fmt.Sprintf("unexpected %s status", mariadb.ReplicationStatusSlaveSQLRunning)
		c.logger.Debug(msg, "expected", "Yes", "actual", sqlRunning)
		return false
	}

	return true
}

func (c *Controller) restartMariaDBReplica() error {
	if err := c.mariaDBConnector.StopReplica(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.StartReplica(); err != nil {
		return err
	}

	return nil
}
