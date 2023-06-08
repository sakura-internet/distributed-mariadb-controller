package v0

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type GetDBControllerStatusResponse struct {
	State string `json:"state"`
}

// GetDBControllerStatus is an http handler that returns the current state of the db-controller.
// that assumes the `UseControllerState` middleware before triggered this.
func GetDBControllerStatus(c echo.Context) error {
	state, err := ExtractControllerState(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, GetDBControllerStatusResponse{State: string(state)})
}
