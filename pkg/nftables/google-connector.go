package nftables

import (
	"github.com/google/nftables"
)

func NewGoogleNftablesConnector() Connector {
	return &GoogleNftablesConnector{}
}

type GoogleNftablesConnector struct {
}

// AddRule implements Connector
func (c *GoogleNftablesConnector) AddRule(tableName string, chainName string, matches []Match, statement Statement) (err error) {
	conn, err := nftables.New()
	if err != nil {
		return err
	}

	defer func() {
		err = conn.CloseLasting()
	}()

	return err
}

// FlushChain implements Connector
func (c *GoogleNftablesConnector) FlushChain(tableName string, chainName string) error {
	return nil
}
