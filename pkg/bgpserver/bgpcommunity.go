package bgpserver

import (
	"fmt"
	"strconv"
	"strings"
)

type Community uint32

// NOTE: panic if fail, so only use to initialize package global variables.
func MustParseCommunity(comm string) Community {
	parts := strings.Split(comm, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("invalid community: %s", comm))
	}
	upper, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(fmt.Sprintf("invalid community: %s", comm))
	}
	lower, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(fmt.Sprintf("invalid community: %s", comm))
	}

	return Community(upper<<16 | lower)
}

// EncodeCommunity converts plain community value to human readable notation(for example 65001:10)
func (c Community) String() string {
	upper := uint32(c) >> 16
	lower := uint32(c) & 0xffff
	return fmt.Sprintf("%d:%d", upper, lower)
}
