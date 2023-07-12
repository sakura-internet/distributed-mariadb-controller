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

package nftables

import (
	"fmt"
	"strings"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bash"
	"golang.org/x/exp/slog"
)

// Connector is an interface that communicates with nftables.
type Connector interface {
	FlushChain(
		table string,
		chain string,
	) error
	AddRule(
		table string,
		chain string,
		matches []Match,
		statement Statement,
	) error
}

func NewDefaultConnector(logger *slog.Logger) Connector {
	return &NftCommandConnector{Logger: logger}
}

// NftComandConnector is a default implementation of Connector.
// this impl uses "nft" commands to interact with nftables.
type NftCommandConnector struct {
	Logger *slog.Logger
}

// AddRule implements Connector
func (c *NftCommandConnector) AddRule(
	table string, chain string, matches []Match, statement Statement) error {
	matchesStr := make([]string, len(matches))
	for i := range matches {
		matchesStr[i] = string(matches[i])
	}

	cmd := fmt.Sprintf("nft add rule %s %s %s %s", table, chain, strings.Join(matchesStr, " "), statement)
	c.Logger.Info("execute command", "command", cmd, "callerFn", "AddRule")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to add rule to chain %s on table %s: %w", chain, table, err)
	}

	return nil
}

// FlushChain implements Connector
func (c *NftCommandConnector) FlushChain(
	table string,
	chain string,
) error {
	cmd := fmt.Sprintf("nft flush chain %s %s", table, chain)
	c.Logger.Info("execute command", "command", cmd, "callerFn", "FlushChain")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to flush chain %s on table %s: %w", chain, table, err)
	}

	return nil
}
