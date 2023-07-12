// Copyright 2023 The distributed-mariadb-controller Authors
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
func (*FakeBGPdConnector) ShowRoutesWithBGPCommunityList(communityList string) (BGP, error) {
	return BGP{}, nil
}

func NewFakeBGPdConnector() BGPdConnector {
	return &FakeBGPdConnector{
		RouteConfigured: make(map[string]bool),
	}
}
