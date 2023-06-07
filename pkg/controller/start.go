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
