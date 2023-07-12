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
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"golang.org/x/exp/slog"
)

const (
	MariaDBDaemonProcessPath = "/usr/sbin/mariadbd"
)

// decideNextStateOnFault determines the next state on fault state
func decideNextStateOnFault(
	logger *slog.Logger,
	neighbors *NeighborSet,
) controller.State {
	if neighbors.primaryNodeExists() {
		return controller.StateReplica
	}

	if neighbors.candidateNodeExists() || neighbors.replicaNodeExists() {
		logger.Info("another candidate or replica exists")
		return controller.StateFault
	}

	// the fault controller is ready to transition to candidate state
	// because network reachability is ok and no one candidate is here.
	return SAKURAControllerStateCandidate
}

// triggerRunOnStateChangesToFault transition to fault state in main loop.
// In fault state, the controller just reflect the fault state to external resources.
func (c *SAKURAController) triggerRunOnStateChangesToFault() error {
	// [STEP1]: START of configurating frrouting
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		c.Logger.Warn("failed to advertise self-address in BGP but ignored because i'm fault", "error", err)
	}
	// [STEP1]: END of configurating frrouting

	// [STEP2]: START of setting nftables state
	if err := c.rejectTCP3306TrafficFromExternal(); err != nil {
		c.Logger.Warn("failed to reject tcp traffic to 3306 but ignored because i'm fault", "error", err)
	}
	// [STEP2]: END of setting nftables state

	// [STEP3]: START of setting MariaDB state
	if err := c.processControlConnector.KillProcessWithFullName(MariaDBDaemonProcessPath); err != nil {
		c.Logger.Warn("failed to kill mariadb daemon process but ignored because i'm fault", "error", err)
	}
	if err := c.stopMariaDBService(); err != nil {
		c.Logger.Warn("failed to stop systemd mariadb process but ignored because i'm fault", "error", err)
	}
	// [STEP3]: END of setting MariaDB state

	c.Logger.Info("fault state handler succeed")
	return nil
}

// rejectTCP3306TrafficFromExternal sets the reject rule that denies the inbound communication from the outsider of the network.
func (c *SAKURAController) rejectTCP3306TrafficFromExternal() error {
	if err := c.nftablesConnector.FlushChain(
		nftables.BuiltinTableFilter,
		nftablesMariaDBChain,
	); err != nil {
		return err
	}

	rejectMatches := []nftables.Match{
		nftables.IFNameMatch(mariaDBServerDefaultIFName),
		nftables.TCPDstPortMatch(mariaDBServerDefaultPort),
	}
	if err := c.nftablesConnector.AddRule(
		nftables.BuiltinTableFilter,
		nftablesMariaDBChain,
		rejectMatches,
		nftables.RejectStatement(),
	); err != nil {
		return err
	}

	return nil
}

// stopMariaDBService stops the mariadb's systemd service.
func (c *SAKURAController) stopMariaDBService() error {
	if err := c.systemdConnector.StopService(mariaDBSerivceName); err != nil {
		return err
	}

	return nil
}
