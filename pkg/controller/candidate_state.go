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
)

// decideNextStateOnCandidate determines the next state on candidate state.
func (c *Controller) decideNextStateOnCandidate() State {
	if c.currentMariaDBHealth == dbHealthCheckResultNG {
		c.logger.Warn("MariaDB is down. falling back to fault state.")
		return StateFault
	}

	if c.currentNeighbors.candidateNodeExists() || c.currentNeighbors.primaryNodeExists() {
		c.logger.Info("another candidate or primary exists. falling back to fault state.")
		return StateFault
	}

	if c.readyToPrimary == readytoPrimaryJudgeOK {
		return StatePrimary
	}

	c.logger.Info("I'm not ready to primary. staying candidate state.")
	return StateCandidate
}

// triggerRunOnStateChangesToCandidate transition to candidate in main loop.
func (c *Controller) triggerRunOnStateChangesToCandidate() error {
	// [STEP1]: setting MariaDB State.
	if err := c.startMariaDBService(); err != nil {
		return err
	}
	if health := c.checkMariaDBHealth(); health == dbHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}
	if err := c.syncReadOnlyVariable( /* read_only=1 */ true); err != nil {
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

	c.logger.Info("candidate state handler succeed")
	return nil
}
