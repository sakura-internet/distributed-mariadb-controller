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

package controller

import (
	"context"
	"time"

	"golang.org/x/exp/slog"
)

// Start starts the controller loop.
// the function recognizes a done signal from the given context.
func Start(
	ctx context.Context,
	logger *slog.Logger,
	ctrler Controller,
	ctrlerLoopInterval time.Duration,
) {
	ticker := time.NewTicker(ctrlerLoopInterval)
	defer ticker.Stop()

controllerLoop:
	for {
		select {
		case <-ctx.Done():
			ctrler.OnExit()
			break controllerLoop
		case <-ticker.C:
			if err := ctrler.PreDecideNextStateHandler(); err != nil {
				logger.Error("controller.PreMakeDecisionHandler()", err, "state", ctrler.GetState())
				continue
			}

			nextState := ctrler.DecideNextState()

			logger.Debug("controller decides next state", "state", nextState)
			if err := ctrler.OnStateHandler(nextState); err != nil {
				logger.Error("controller.OnStateHandler()", err, "state", nextState)
			}
		}
	}
}
