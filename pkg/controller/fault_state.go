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
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
)

// decideNextStateOnFault determines the next state on fault state
func (c *Controller) decideNextStateOnFault() State {
	if c.currentNeighbors.primaryNodeExists() {
		return StateReplica
	}

	if c.currentNeighbors.candidateNodeExists() || c.currentNeighbors.replicaNodeExists() {
		c.logger.Info("another candidate or replica exists")
		return StateFault
	}

	// the fault controller is ready to transition to candidate state
	// because network reachability is ok and no one candidate is here.
	return StateCandidate
}

// triggerRunOnStateChangesToFault transition to fault state in main loop.
// In fault state, the controller just reflect the fault state to external resources.
func (c *Controller) triggerRunOnStateChangesToFault() error {
	// [STEP1]: configure bgp route
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		c.logger.Warn("failed to advertise self-address in BGP but ignored because i'm fault", "error", err)
	}

	// [STEP2]: setting nftables state
	if err := c.rejectDatabaseServiceTraffic(); err != nil {
		c.logger.Warn("failed to reject tcp traffic to database serving port but ignored because i'm fault", "error", err)
	}

	// [STEP3]: setting MariaDB state
	if err := c.systemdConnector.KillService(mariadb.SystemdSerivceName); err != nil {
		c.logger.Warn("failed to kill db service but ignored because i'm fault", "error", err)
	}
	if err := c.stopMariaDBService(); err != nil {
		c.logger.Warn("failed to stop systemd mariadb process but ignored because i'm fault", "error", err)
	}

	c.logger.Info("fault state handler succeed")
	return nil
}

// rejectDatabaseServiceTraffic sets the reject rule that denies the inbound communication from the outsider of the network.
func (c *Controller) rejectDatabaseServiceTraffic() error {
	if err := c.nftablesConnector.FlushChain(
		c.dbAclChainName,
	); err != nil {
		return err
	}

	rejectMatches := []nftables.Match{
		nftables.IFNameMatch(c.globalInterfaceName),
		nftables.TCPDstPortMatch(uint16(c.dbServingPort)),
	}
	if err := c.nftablesConnector.AddRule(
		c.dbAclChainName,
		rejectMatches,
		nftables.RejectStatement(),
	); err != nil {
		return err
	}

	return nil
}

// stopMariaDBService stops the mariadb's systemd service.
func (c *Controller) stopMariaDBService() error {
	if err := c.systemdConnector.StopService(mariadb.SystemdSerivceName); err != nil {
		return err
	}

	return nil
}
