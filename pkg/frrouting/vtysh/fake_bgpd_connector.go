package vtysh

import (
	"net"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
)

// FakeBGPdConnector is for testing the controller.
type FakeBGPdConnector struct {
	// RouteConfigured checks whether the prefix is configured on the (fake) vtysh.
	RouteConfigured map[string]bool
}

// ConfigureRouteWithRouteMap implements vtysh.BGPdConnector
func (c *FakeBGPdConnector) ConfigureRouteWithRouteMap(prefix net.IPNet, routeMap string) error {
	c.RouteConfigured[prefix.String()] = true
	return nil
}

// ShowRoutesWithBGPCommunityList implements vtysh.BGPdConnector
func (*FakeBGPdConnector) ShowRoutesWithBGPCommunityList(communityList string) (bgpd.BGP, error) {
	return bgpd.BGP{}, nil
}

func NewFakeBGPdConnector() BGPdConnector {
	return &FakeBGPdConnector{
		RouteConfigured: make(map[string]bool),
	}
}
