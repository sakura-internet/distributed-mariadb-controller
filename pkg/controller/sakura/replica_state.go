package sakura

import (
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"golang.org/x/exp/slog"
)

const (
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
func (c *SAKURAController) triggerRunOnStateChangesToReplica(
	neighbors *NeighborSet,
) error {
	// [STEP1]: START of setting MariaDB State.
	if err := c.startMariaDBService(); err != nil {
		return err
	}
	if health := c.checkMariaDBHealth(); health == MariaDBHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}

	if !neighbors.primaryNodeExists() {
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

	primaryNode := neighbors.NeighborMatrix[controller.StatePrimary][0]
	master := mariadb.MasterInstance{
		Host:     primaryNode.Address,
		Port:     mariaDBMasterDefaultPort,
		User:     mariaDBMasterDefaultUser,
		Password: c.dbReplicaPassword,
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

	slog.Info("replica state handler succeed")
	return nil
}
