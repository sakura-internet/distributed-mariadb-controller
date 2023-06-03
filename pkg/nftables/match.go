package nftables

import "fmt"

type Match string

func IPSrcAddrMatch(srcAddr string) Match {
	return Match(fmt.Sprintf("ip saddr %s", srcAddr))
}

func TCPDstPortMatch(dport uint16) Match {
	return Match(fmt.Sprintf("tcp dport %d", dport))
}

func IFNameMatch(ifname string) Match {
	return Match(fmt.Sprintf("iifname \"%s\"", ifname))
}
