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
	"time"
)

// FakeSystemdConnector is for testing the controller.
type FakeSystemdConnector struct {
	// Timestamp holds the method execution's timestamp.
	Timestamp map[string]time.Time
	// ServiceStarted checks whether the (fake) service is started.
	ServiceStarted map[string]bool
}

// CheckServiceStatus implements systemd.Connector
func (c *FakeSystemdConnector) CheckServiceStatus(serviceName string) error {
	c.Timestamp["CheckServiceStatus"] = time.Now()
	return nil
}

// StartService implements systemd.Connector
func (c *FakeSystemdConnector) StartService(serviceName string) error {
	c.ServiceStarted[serviceName] = true
	c.Timestamp["StartService"] = time.Now()
	return nil
}

// StopService implements systemd.Connector
func (c *FakeSystemdConnector) StopService(serviceName string) error {
	c.ServiceStarted[serviceName] = false
	c.Timestamp["StopService"] = time.Now()
	return nil
}

// KillService implements systemd.Connector
func (c *FakeSystemdConnector) KillService(serviceName string) error {
	c.ServiceStarted[serviceName] = false
	c.Timestamp["KillService"] = time.Now()
	return nil
}

func NewFakeSystemdConnector() Connector {
	return &FakeSystemdConnector{
		Timestamp:      make(map[string]time.Time),
		ServiceStarted: make(map[string]bool),
	}
}
