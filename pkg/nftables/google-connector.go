package nftables

import (
	"github.com/google/nftables"
)

func NewGoogleNftablesConnector() Connector {
	conn, err := nftables.New()
	if err != nil {
		panic("failed to create a new netlink connection")
	}
	return &GoogleNftablesConnector{
		conn: conn,
	}
}

type GoogleNftablesConnector struct {
	conn *nftables.Conn
}

// AddRule implements Connector
func (c *GoogleNftablesConnector) AddRule(tableName string, chainName string, matches []Match, statement Statement) error {
	chain := nftables.Chain{
		Name: chainName,
	}
	table := nftables.Table{
		Name: tableName,
	}
	rule := nftables.Rule{
		Table: &table,
		Chain: &chain,
	}

	_ = c.conn.AddRule(&rule)
	return nil
}

// FlushChain implements Connector
func (c *GoogleNftablesConnector) FlushChain(tableName string, chainName string) error {
	chain := nftables.Chain{
		Name: chainName,
		Table: &nftables.Table{
			Name: tableName,
		},
	}

	c.conn.FlushChain(&chain)
	return nil
}
