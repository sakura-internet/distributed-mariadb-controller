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
	"time"
)

// FakeProcessControlConnector is for testing the controller.
type FakeProcessControlConnector struct {
	// Timestamp holds the method execution's timestamp.
	Timestamp map[string]time.Time
	// ProcessLived checks whether the (fake) process is lived.
	ProcessLived map[string]bool
}

// KillProcessWithFullName implements process.ProcessControlConnector
func (c *FakeProcessControlConnector) KillProcessWithFullName(processName string) error {
	c.ProcessLived[processName] = false
	c.Timestamp["KillProcessWithFullName"] = time.Now()

	return nil
}

func NewFakeProcessControlConnector() ProcessControlConnector {
	return &FakeProcessControlConnector{
		Timestamp:    make(map[string]time.Time),
		ProcessLived: make(map[string]bool),
	}
}
