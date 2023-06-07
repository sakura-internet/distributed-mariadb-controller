package sakura

import (
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
)

const (
	// replicationStatusCheckThreshold is a threshold.
	// if the counter of the controller overs this, the controller goes panic.
	replicationStatusCheckThreshold = 20

	mariaDBMasterDefaultPort = 13306
	mariaDBMasterDefaultUser = "repl"
)

// decideNextStateOnReplica determines the next state on replica state.
func decideNextStateOnReplica(
	neighbors *NeighborSet,
	mariaDBHealth MariaDBHealthCheckResult,
) controller.State {
	if mariaDBHealth == MariaDBHealthCheckResultNG {
		return controller.StateFault
	}

	noPrimary := !neighbors.primaryNodeExists()
	noCandidate := !neighbors.candidateNodeExists()
	if noPrimary && noCandidate {
		// you may be the next primary node!
		return SAKURAControllerStateCandidate
	}

	return controller.StateReplica
}

// triggerRunOnStateChangesToReplica transition to replica state in main loop.
func (c *SAKURAController) triggerRunOnStateChangesToReplica() error {
	// [STEP1]: START of setting MariaDB State.
	if err := c.startMariaDBService(); err != nil {
		return err
	}
	if health := c.checkMariaDBHealth(); health == MariaDBHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}

	if !c.CurrentNeighbors.primaryNodeExists() {
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

	primaryNode := c.CurrentNeighbors.NeighborMatrix[controller.StatePrimary][0]
	master := mariadb.MasterInstance{
		Host:     primaryNode.Address,
		Port:     mariaDBMasterDefaultPort,
		User:     mariaDBMasterDefaultUser,
		Password: c.MariaDBReplicaPassword,
		UseGTID:  mariadb.MasterUseGTIDValueCurrentPos,
	}
	if err := c.mariaDBConnector.ChangeMasterTo(master); err != nil {
		return err
	}

	if err := c.mariaDBConnector.StartReplica(); err != nil {
		return err
	}
	// [STEP1]: END of setting MariaDB State.

	// [STEP2]: START of setting Nftables State.
	if err := c.rejectTCP3306TrafficFromExternal(); err != nil {
		return err
	}
	// [STEP2]: END of setting Nftables State.

	// [STEP3]: START of configurating frrouting.
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		return err
	}
	// [STEP3]: END of configurating frrouting.

	// reset the count because the controller is healthy for replica mode.
	c.replicationStatusCheckFailCount = 0

	c.Logger.Info("replica state handler succeed")
	return nil
}

func (c *SAKURAController) triggerRunOnStateKeepsReplica() error {
	if c.replicationStatusCheckFailCount >= replicationStatusCheckThreshold {
		// we should manually operate the case for recovering.
		// ref: https://github.sakura.codes/ohkubo/sacloud-multi-az-database/issues/28
		return fmt.Errorf("reached the maximum retry limit for replication")
	}

	if err := c.checkMariaDBReplicationStatus(); err != nil {
		// we should keep trying to challenge that the replication status satisfies our conditions.
		c.replicationStatusCheckFailCount++
		c.Logger.Warn("failed to satisfy replication conditions", "error", err, "replicationCount", c.replicationStatusCheckFailCount)

		if err := c.restartMariaDBReplica(); err != nil {
			c.Logger.Warn("failed to restart replica", "error", err)
		}
	} else {
		// reset the count because the controller is healthy.
		c.replicationStatusCheckFailCount = 0
	}

	return nil
}

// checkMariaDBReplicationStatus returns true if the status of replication is satisfied.
// if the challenge failed to satisfy the conditions, this function returns false.
func (c *SAKURAController) checkMariaDBReplicationStatus() error {
	status, err := c.mariaDBConnector.ShowReplicationStatus()
	if err != nil {
		return err
	}

	if !c.checkRequiredReplicationStatusIsOK(status) {
		return fmt.Errorf("failed to satisfy the replication conditions")
	}

	// lastNotifiedReplicationDelay := time.Unix(0, 0)
	// c.validateReplicationDelaySeconds(replicaStatus, lastNotifiedReplicationDelay)
	return nil
}

// checkRequiredReplicationStatusIsOK checks the replication status satisfies the required conditions.
func (c *SAKURAController) checkRequiredReplicationStatusIsOK(status mariadb.ReplicationStatus) bool {
	ioRunning, ok1 := status[mariadb.ReplicationStatusSlaveIORunning]
	sqlRunning, ok2 := status[mariadb.ReplicationStatusSlaveSQLRunning]

	if !ok1 || !ok2 {
		msg := fmt.Sprintf("failed to retrieve %s or %s",
			mariadb.ReplicationStatusSlaveIORunning,
			mariadb.ReplicationStatusSlaveSQLRunning)
		c.Logger.Debug(msg)
		return false
	}

	if ioRunning != mariadb.ReplicationStatusSlaveIORunningYes {
		msg := fmt.Sprintf("unexpected %s status", mariadb.ReplicationStatusSlaveIORunning)
		c.Logger.Debug(msg, "expected", "Yes", "actual", ioRunning)
		return false
	}

	if sqlRunning != mariadb.ReplicationStatusSlaveSQLRunningYes {
		msg := fmt.Sprintf("unexpected %s status", mariadb.ReplicationStatusSlaveSQLRunning)
		c.Logger.Debug(msg, "expected", "Yes", "actual", sqlRunning)
		return false
	}

	return true
}

func (c *SAKURAController) restartMariaDBReplica() error {
	if err := c.mariaDBConnector.StopReplica(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.StartReplica(); err != nil {
		return err
	}

	return nil
}
