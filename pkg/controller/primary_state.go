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

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
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
func (c *Controller) decideNextStateOnPrimary() State {
	if c.currentMariaDBHealth == dbHealthCheckResultNG {
		c.logger.Warn("MariaDB instance is down")
		return StateFault
	}

	// found dual-primary situation.
	if c.currentNeighbors.primaryNodeExists() {
		c.logger.Warn("dual primary detected")
		return StateFault
	}

	// won't transition to other state.
	return StatePrimary
}

// triggerRunOnStateChangesToPrimary processes transition to primary state in main loop.
func (c *Controller) triggerRunOnStateChangesToPrimary() error {
	if health := c.checkMariaDBHealth(); health == dbHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}
	if c.currentNeighbors.primaryNodeExists() {
		return fmt.Errorf("dual primary detected")
	}

	// [STEP1]: setting MariaDB state
	if err := c.mariaDBConnector.StopReplica(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.ResetAllReplicas(); err != nil {
		return err
	}
	if err := c.syncReadOnlyVariable( /* read_only=0 */ false); err != nil {
		return err
	}

	// [STEP2]: setting nftables state
	if err := c.acceptDatabaseServiceTraffic(); err != nil {
		return err
	}

	// [STEP3]: configurating frrouting
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		return err
	}

	// reset the count because the controller is healthy.
	c.writeTestDataFailCount = 0

	c.logger.Info("primary state handler succeed")
	return nil
}

// triggerRunOnStateKeepsPrimary is the handler that is triggered when the prev/current state is different.
func (c *Controller) triggerRunOnStateKeepsPrimary() error {
	if c.writeTestDataFailCount >= writeTestDataFailCountThreshold {
		return fmt.Errorf("reached the maximum fail count of write test data")
	}

	if err := c.writeTestDataToMariaDB(); err != nil {
		c.writeTestDataFailCount++
		c.logger.Warn("failed to write test data to mariadb", "error", err, "failedCount", c.writeTestDataFailCount)
		// return noerror because this is soft fail
		return nil
	}

	// reset the count because the controller is healthy.
	c.writeTestDataFailCount = 0
	return nil
}

// acceptDatabaseServiceTraffic sets the rule that accepts the inbound communication.
func (c *Controller) acceptDatabaseServiceTraffic() error {
	if err := c.nftablesConnector.FlushChain(c.dbAclChainName); err != nil {
		return err
	}

	acceptMatches := []nftables.Match{
		nftables.IFNameMatch(c.globalInterfaceName),
		nftables.TCPDstPortMatch(uint16(c.dbServingPort)),
	}

	if err := c.nftablesConnector.AddRule(c.dbAclChainName, acceptMatches, nftables.AcceptStatement()); err != nil {
		return err
	}

	return nil
}

// writeTestDataToMariaDB tries to write the testdata to MariaDB.
func (c *Controller) writeTestDataToMariaDB() error {
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

	return nil
}

// createManagementDatabase tries to create the management database.
// if the management database is already exist, the function does nothing.
func (c *Controller) createManagementDatabase() error {
	return c.mariaDBConnector.CreateDatabase(managementDatabaseName)
}

// createManagementDatabase tries to create alive-check table on the management database.
// if the alive-check table is already exist, the function does nothing.
func (c *Controller) createAliveCheckTableOnManagementDB() error {
	return c.mariaDBConnector.CreateIDTable(managementDatabaseName, aliveCheckTableName)
}

// insertTemporaryRecordToAliveCheck tries to insert temporary record to alive-check table.
func (c *Controller) insertTemporaryRecordToAliveCheck() error {
	// id has no meaning.
	return c.mariaDBConnector.InsertIDRecord(managementDatabaseName, aliveCheckTableName, 1)
}

// deleteTemporaryRecordOnAliveCheck tries to delete records on alive-check table.
func (c *Controller) deleteTemporaryRecordOnAliveCheck() error {
	return c.mariaDBConnector.DeleteRecords(managementDatabaseName, aliveCheckTableName)
}
