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
