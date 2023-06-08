package v0

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller/sakura"
)

const (
	controllerStateCtxKey = "controllerState"
)

// UseControllerState is an echo middleware that injects the current state of the db-controller into othe request context.
func UseControllerState(ctrler *sakura.SAKURAController) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(controllerStateCtxKey, ctrler.GetState())
			return next(c)
		}
	}
}

// ExtractControllerState is an utility for retrieving the controller state from request context.
func ExtractControllerState(c echo.Context) (controller.State, error) {
	v := c.Get(controllerStateCtxKey)
	if v == nil {
		return controller.StateFault, fmt.Errorf("failed to get controller state from context")
	}

	return v.(controller.State), nil
}
