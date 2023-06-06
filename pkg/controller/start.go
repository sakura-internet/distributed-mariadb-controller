package controller

import (
	"context"
	"math/rand"
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
			if err := ctrler.PreMakeDecisionHandler(); err != nil {
				logger.Error("controller.PreMakeDecisionHandler()", err, "state", ctrler.GetState())
				continue
			}

			time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))

			nextState := ctrler.MakeDecision()

			if err := ctrler.OnStateHandler(nextState); err != nil {
				logger.Error("controller.OnStateHandler()", err, "state", nextState)
			}
		}
	}
}
