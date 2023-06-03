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
