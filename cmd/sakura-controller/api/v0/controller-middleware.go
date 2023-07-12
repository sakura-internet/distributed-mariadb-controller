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
