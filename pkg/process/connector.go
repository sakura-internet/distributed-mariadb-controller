package process

import (
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bash"
	"golang.org/x/exp/slog"
)

// ProcessControlConnector is an interface that communicates with Linux processes.
type ProcessControlConnector interface {
	KillProcessWithFullName(processName string) error
}

func NewDefaultConnector(logger *slog.Logger) ProcessControlConnector {
	return &ShellCommandConnector{}
}

// ShellCommandConnector is a default implementation of ProcessControlConnector.
// this impl uses "pkill/etc" commands to interact with Linux processes.
type ShellCommandConnector struct {
	Logger *slog.Logger
}

// KillProcessWithFullName implements Connector
func (c *ShellCommandConnector) KillProcessWithFullName(
	processName string,
) error {
	cmd := fmt.Sprintf("pkill -9 -f %s", processName)

	c.Logger.Info("execute command", "command", cmd, "callerFn", "KillProcessWithFullName")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to kill %s process: %w", processName, err)
	}

	return nil
}
