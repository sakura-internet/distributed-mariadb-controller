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
	"fmt"
	"log/slog"
	"time"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/command"
)

const (
	builtinTableFilter = "filter"
)

var (
	nftCommandTimeout = 5 * time.Second
)

// Connector is an interface that communicates with nftables.
type Connector interface {
	FlushChain(chain string) error
	CreateChain(chain string) error
	AddRule(chain string, matches []Match, statement statement) error
}

// nftCommandConnector is a default implementation of Connector.
// this impl uses "nft" commands to interact with nftables.
type nftCommandConnector struct {
	logger *slog.Logger
}

func NewDefaultConnector(logger *slog.Logger) Connector {
	return &nftCommandConnector{logger: logger}
}

// AddRule implements Connector
func (c *nftCommandConnector) AddRule(
	chain string, matches []Match, stmt statement,
) error {
	name := "nft"
	args := []string{"add", "rule", builtinTableFilter, chain}
	for _, match := range matches {
		args = append(args, match...)
	}
	args = append(args, stmt...)
	c.logger.Info("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(nftCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to add rule to chain %s on table %s: %w", chain, builtinTableFilter, err)
	}

	return nil
}

// FlushChain implements Connector
func (c *nftCommandConnector) FlushChain(
	chain string,
) error {
	name := "nft"
	args := []string{"flush", "chain", builtinTableFilter, chain}
	c.logger.Info("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(nftCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to flush chain %s on table %s: %w", chain, builtinTableFilter, err)
	}

	return nil
}

// CreateChain implements Connector
func (c *nftCommandConnector) CreateChain(
	chain string,
) error {
	// nft add chain comand returns ok if the chain is already exist.
	name := "nft"
	args := []string{"add", "chain", builtinTableFilter, chain, "{ type filter hook input priority 0; }"}
	c.logger.Info("execute command", "name", name, "args", args)
	if _, err := command.RunWithTimeout(nftCommandTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to add nft chain: %w", err)
	}

	return nil
}
