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

package nftables

import (
	"time"
)

// FakeNftablersConnector is for testing the controller.
type FakeNftablesConnector struct {
	// Timestamp holds the method calling's timestamp.
	Timestamp map[string]time.Time
}

// AddRule implements nftables.Connector
func (c *FakeNftablesConnector) AddRule(chain string, matches []Match, statement statement) error {
	c.Timestamp["AddRule"] = time.Now()
	return nil
}

// FlushChain implements nftables.Connector
func (c *FakeNftablesConnector) FlushChain(chain string) error {
	c.Timestamp["FlushChain"] = time.Now()
	return nil
}

// CreateChain implements nftables.Connector
func (c *FakeNftablesConnector) CreateChain(chain string) error {
	c.Timestamp["CreateChain"] = time.Now()
	return nil
}

func NewFakeNftablesConnector() Connector {
	return &FakeNftablesConnector{
		Timestamp: make(map[string]time.Time),
	}
}
