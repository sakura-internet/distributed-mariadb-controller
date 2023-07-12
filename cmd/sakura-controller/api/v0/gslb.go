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
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
)

func GSLBHealthCheckEndpoint(c echo.Context) error {
	state, err := ExtractControllerState(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &ErrorResponse{Message: err.Error()})
	}

	if state != controller.StatePrimary {
		return c.NoContent(http.StatusServiceUnavailable)
	}

	return c.NoContent(http.StatusOK)
}
