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
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"golang.org/x/exp/slog"
)

// decideNextStateOnCandidate determines the next state on candidate state.
func decideNextStateOnCandidate(
	logger *slog.Logger,
	neighbors *NeighborSet,
	mariaDBHealth MariaDBHealthCheckResult,
	readyToPrimaryJudge ReadyToPrimaryJudge,
) controller.State {
	if mariaDBHealth == MariaDBHealthCheckResultNG {
		logger.Warn("MariaDB is down. falling back to fault state.")
		return controller.StateFault
	}

	if neighbors.candidateNodeExists() || neighbors.primaryNodeExists() {
		logger.Info("another candidate or primary exists. falling back to fault state.")
		return controller.StateFault
	}

	if readyToPrimaryJudge == ReadytoPrimaryJudgeOK {
		return controller.StatePrimary
	}

	logger.Info("I'm not ready to primary. staying candidate state.")
	return SAKURAControllerStateCandidate
}

// triggerRunOnStateChangesToCandidate transition to candidate in main loop.
func (c *SAKURAController) triggerRunOnStateChangesToCandidate() error {
	// [STEP1]: START of setting MariaDB State.
	if err := c.startMariaDBService(); err != nil {
		return err
	}
	if health := c.checkMariaDBHealth(); health == MariaDBHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}

	if err := c.syncReadOnlyVariable( /* read_only=1 */ true); err != nil {
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

	c.Logger.Info("candidate state handler succeed")
	return nil
}
