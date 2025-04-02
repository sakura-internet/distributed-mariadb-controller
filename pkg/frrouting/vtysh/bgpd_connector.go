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

package vtysh

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/command"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
)

var (
	vtyshTimeout = 5 * time.Second
)

// vtyshBGPdConnector is a default implementation of BGPdConnector.
// this impl uses "vtysh" commands to interact with frrouting bgpd.
type vtyshBGPdConnector struct {
	logger *slog.Logger
}

func NewDefaultBGPdConnector(logger *slog.Logger) bgpd.BGPdConnector {
	return &vtyshBGPdConnector{logger: logger}
}

// ShowRoutesWithBGPCommunityList implements BGPdConnector
func (c *vtyshBGPdConnector) ShowRoutesWithBGPCommunityList(
	communityList string,
) (bgpd.BGP, error) {
	bgp := bgpd.BGP{}

	name := "vtysh"
	args := []string{"-H", "/dev/null", "-c", fmt.Sprintf("show ip bgp community-list %s json", communityList)}

	c.logger.Debug("execute command", "name", name, "args", args, "callerFn", "ShowRoutesWithBGPCommunityList")
	out, err := command.RunWithTimeout(vtyshTimeout, name, args...)
	if err != nil {
		return bgp, fmt.Errorf("failed to show ip bgp community-list: %w", err)
	}

	if err := json.Unmarshal(out, &bgp); err != nil {
		return bgp, fmt.Errorf("failed to unmarshal to bgpd.BGP: %w", err)
	}

	return bgp, nil
}

// ConfigureRouteWithRouteMap implements BGPdConnector
func (c *vtyshBGPdConnector) ConfigureRouteWithRouteMap(
	prefix net.IPNet,
	routeMap string,
) error {
	name := "vtysh"
	configCommand := fmt.Sprintf("network %s route-map %s", prefix.String(), routeMap)
	args := []string{"-H", "/dev/null", "-c", "conf t", "-c", "router bgp", "-c", configCommand}

	c.logger.Info("execute command", "name", name, "args", args, "callerFn", "ConfigureUnicastRouteWithRouteMap")
	if _, err := command.RunWithTimeout(vtyshTimeout, name, args...); err != nil {
		return fmt.Errorf("failed to advertise %s route: %w", prefix.String(), err)
	}

	return nil
}
