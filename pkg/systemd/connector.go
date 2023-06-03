package systemd

import (
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bash"
	"golang.org/x/exp/slog"
)

// Connector is an interface that communicates with systemd.
type Connector interface {
	// StartSerivce starts a systemd service.
	// preHook are triggered before starting a service.
	// postHook are triggered after starting a service.
	StartService(
		serviceName string,
		preHook func() error,
		postHook func() error,
	) error

	StopService(
		serviceName string,
	) error
	// CheckServiceStatus checks the status of a systemd service.
	CheckServiceStatus(
		serviceName string,
	) error
}

func NewDefaultConnector(logger *slog.Logger) Connector {
	return &SystemCtlConnector{Logger: logger}
}

// SystemCtlConnector is a default implementation of Connector.
// this impl uses "systemctl" commands to interact with systemd.
type SystemCtlConnector struct {
	Logger *slog.Logger
}

// StartSerivce implements Connector
func (c *SystemCtlConnector) StartService(
	serviceName string,
	preHook func() error,
	postHook func() error,
) error {
	if preHook != nil {
		if err := preHook(); err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("systemctl start %s", serviceName)
	c.Logger.Info("execute command", "command", cmd, "callerFn", "CheckServiceStatus")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to start %s service: %w", serviceName, err)
	}

	if postHook != nil {
		if err := postHook(); err != nil {
			return err
		}
	}

	return nil
}

// CheckServiceStatus implements Connector
func (c *SystemCtlConnector) CheckServiceStatus(
	serviceName string,
) error {
	cmd := fmt.Sprintf("systemctl status %s", serviceName)
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "CheckServiceStatus")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to check %s service: %w", serviceName, err)
	}

	return nil
}

// StopService implements Connector
func (c *SystemCtlConnector) StopService(
	serviceName string,
) error {
	cmd := fmt.Sprintf("systemctl stop %s", serviceName)

	c.Logger.Info("execute command", "command", cmd, "callerFn", "StopService")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to stop service %s : %w", serviceName, err)
	}

	return nil
}
