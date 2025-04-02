// Copyright 2025 The distributed-mariadb-controller Authors
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

package systemd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/command"
)

const (
	systemctlCommandTimeout = 60 * time.Second
)

// Connector is an interface that communicates with systemd.
type Connector interface {
	// StartSerivce starts a systemd service.
	StartService(serviceName string) error

	// StopService stops a systemd service.
	StopService(serviceName string) error

	// KillService kills a systemd service.
	KillService(serviceName string) error

	// CheckServiceStatus checks the status of a systemd service.
	CheckServiceStatus(serviceName string) error
}

// systemCtlConnector is a default implementation of Connector.
// this impl uses "systemctl" commands to interact with systemd.
type systemCtlConnector struct {
	logger *slog.Logger
}

func NewDefaultConnector(logger *slog.Logger) Connector {
	return &systemCtlConnector{logger: logger}
}

// StartSerivce implements Connector
func (c *systemCtlConnector) StartService(
	serviceName string,
) error {
	name := "systemctl"
	args := []string{"start", serviceName}
	c.logger.Info("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(systemctlCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to start %s service: %w", serviceName, err)
	}

	return nil
}

// CheckServiceStatus implements Connector
func (c *systemCtlConnector) CheckServiceStatus(
	serviceName string,
) error {
	name := "systemctl"
	args := []string{"status", serviceName}
	c.logger.Debug("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(systemctlCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to check %s service: %w", serviceName, err)
	}

	return nil
}

// StopService implements Connector
func (c *systemCtlConnector) StopService(
	serviceName string,
) error {
	name := "systemctl"
	args := []string{"stop", serviceName}
	c.logger.Info("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(systemctlCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to stop service %s : %w", serviceName, err)
	}

	return nil
}

// KillService implements Connector
func (c *systemCtlConnector) KillService(
	serviceName string,
) error {
	name := "systemctl"
	args := []string{"kill", "-s", "SIGKILL", serviceName}
	c.logger.Info("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(systemctlCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to stop service %s : %w", serviceName, err)
	}

	return nil
}
