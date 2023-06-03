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
func (c *FakeNftablesConnector) AddRule(table string, chain string, matches []Match, statement Statement) error {
	c.Timestamp["AddRule"] = time.Now()
	return nil
}

// FlushChain implements nftables.Connector
func (c *FakeNftablesConnector) FlushChain(table string, chain string) error {
	c.Timestamp["FlushChain"] = time.Now()
	return nil
}

func NewFakeNftablesConnector() Connector {
	return &FakeNftablesConnector{
		Timestamp: make(map[string]time.Time),
	}
}
