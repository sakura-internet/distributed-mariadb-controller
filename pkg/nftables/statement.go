package nftables

import "fmt"

type Statement string

func AcceptStatement() Statement {
	return Statement("accept")
}

func RejectStatement() Statement {
	return Statement("reject")
}

func RejectStatementWithProto(
	proto string,
	protoType string,
) Statement {
	return Statement(fmt.Sprintf("reject with %s type %s", proto, protoType))
}
