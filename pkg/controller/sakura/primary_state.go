package sakura

import (
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"golang.org/x/exp/slog"
)

const (
	writeTestDataFailCountThreshold = 15

	// managementDatabaseName is the name of the management database.
	// the database will be used for checking the db-controller is able to write some data to MariaDB.
	managementDatabaseName = "management"
	// aliveCheckTableName is the name of the alive-check table on management DB.
	// the table holds temporary records for checking the db-controller is able to write some data to MariaDB.
	aliveCheckTableName = "alive_check"
)

// decideNextStateOnPrimary determines the next state on primary state
func decideNextStateOnPrimary(
	logger *slog.Logger,
	neighbors *NeighborSet,
	mariaDBHealth MariaDBHealthCheckResult,
) controller.State {
	if mariaDBHealth == MariaDBHealthCheckResultNG {
		logger.Warn("MariaDB instance is down")
		return controller.StateFault
	}

	// found dual-primary situation.
	if neighbors.primaryNodeExists() {
		logger.Warn("dual primary detected")
		return controller.StateFault
	}

	// won't transition to other state.
	return controller.StatePrimary
}

// triggerRunOnStateChangesToPrimary processes transition to primary state in main loop.
func (c *SAKURAController) triggerRunOnStateChangesToPrimary() error {
	if health := c.checkMariaDBHealth(); health == MariaDBHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}
	if c.CurrentNeighbors.primaryNodeExists() {
		return fmt.Errorf("dual primary detected")
	}

	// [STEP1]: START of setting MariaDB state
	if err := c.mariaDBConnector.StopReplica(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.ResetAllReplicas(); err != nil {
		return err
	}
	if err := c.syncReadOnlyVariable( /* read_only=0 */ false); err != nil {
		return err
	}
	// [STEP1]: END of setting MariaDB state

	// [STEP2]: START of setting nftables state
	if err := c.acceptTCP3306Traffic(); err != nil {
		return err
	}
	// [STEP2]: END of setting nftables state

	// [STEP3]: START of configurating frrouting
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		return err
	}
	// [STEP3]: END of configurating frrouting

	// reset the count because the controller is healthy.
	c.writeTestDataFailCount = 0

	c.Logger.Info("primary state handler succeed")
	return nil
}

// triggerRunOnStateKeepsPrimary is the handler that is triggered when the prev/current state is different.
func (c *SAKURAController) triggerRunOnStateKeepsPrimary() error {
	if c.writeTestDataFailCount >= writeTestDataFailCountThreshold {
		return fmt.Errorf("reached the maximum fail count of write test data")
	}

	if err := c.writeTestDataToMariaDB(); err != nil {
		c.writeTestDataFailCount++
		c.Logger.Warn("failed to write test data to mariadb", "error", err, "failedCount", c.writeTestDataFailCount)
	} else {
		// reset the count because the controller is healthy.
		c.writeTestDataFailCount = 0
	}

	return nil
}

// acceptTCP3306Traffic sets the rule that accepts the inbound communication.
func (c *SAKURAController) acceptTCP3306Traffic() error {
	if err := c.nftablesConnector.FlushChain(nftables.BuiltinTableFilter, nftablesMariaDBChain); err != nil {
		return err
	}

	acceptMatches := []nftables.Match{
		nftables.IFNameMatch(mariaDBServerDefaultIFName),
		nftables.TCPDstPortMatch(mariaDBServerDefaultPort),
	}

	if err := c.nftablesConnector.AddRule(nftables.BuiltinTableFilter, nftablesMariaDBChain, acceptMatches, nftables.AcceptStatement()); err != nil {
		return err
	}

	return nil
}

// writeTestDataToMariaDB tries to write the testdata to MariaDB.
func (c *SAKURAController) writeTestDataToMariaDB() error {
	if err := c.createManagementDatabase(); err != nil {
		return err
	}
	if err := c.createAliveCheckTableOnManagementDB(); err != nil {
		return err
	}
	if err := c.insertTemporaryRecordToAliveCheck(); err != nil {
		return err
	}
	if err := c.deleteTemporaryRecordOnAliveCheck(); err != nil {
		return err
	}

	if err := c.systemdConnector.CheckServiceStatus("mariadb"); err != nil {
		return err
	}

	return nil
}

// createManagementDatabase tries to create the management database.
// if the management database is already exist, the function does nothing.
func (c *SAKURAController) createManagementDatabase() error {
	return c.mariaDBConnector.CreateDatabase(managementDatabaseName)
}

// createManagementDatabase tries to create alive-check table on the management database.
// if the alive-check table is already exist, the function does nothing.
func (c *SAKURAController) createAliveCheckTableOnManagementDB() error {
	return c.mariaDBConnector.CreateIDTable(managementDatabaseName, aliveCheckTableName)
}

// insertTemporaryRecordToAliveCheck tries to insert temporary record to alive-check table.
func (c *SAKURAController) insertTemporaryRecordToAliveCheck() error {
	// id has no meaning.
	return c.mariaDBConnector.InsertIDRecord(managementDatabaseName, aliveCheckTableName, 1)
}

// deleteTemporaryRecordOnAliveCheck tries to delete records on alive-check table.
func (c *SAKURAController) deleteTemporaryRecordOnAliveCheck() error {
	return c.mariaDBConnector.DeleteRecords(managementDatabaseName, aliveCheckTableName)
}
