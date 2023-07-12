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
	return &ShellCommandConnector{Logger: logger}
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
