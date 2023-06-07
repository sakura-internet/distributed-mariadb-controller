package bgpd

import (
	"net"
)

// Connector is an interface that communicates with FRRouting BGPd.
type BGPdConnector interface {
	// ShowRoutesWithBGPCommunityList shows the routes that with the given bgp community-list.
	ShowRoutesWithBGPCommunityList(
		communityList string,
	) (BGP, error)

	// ConfigureRouteWithRouteMap configs the route-advertising with the given route-map.
	ConfigureRouteWithRouteMap(
		prefix net.IPNet,
		routeMap string,
	) error
}
