// Copyright 2025 The distributed-mariadb-controller Authors
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
