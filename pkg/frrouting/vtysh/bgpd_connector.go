package vtysh

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bash"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"golang.org/x/exp/slog"
)

func NewDefaultBGPdConnector(logger *slog.Logger) bgpd.BGPdConnector {
	return &VtyshBGPdConnector{Logger: logger}
}

// VtyshBGPdConnector is a default implementation of BGPdConnector.
// this impl uses "vtysh" commands to interact with frrouting bgpd.
type VtyshBGPdConnector struct {
	Logger *slog.Logger
}

// ShowRoutesWithBGPCommunityList implements BGPdConnector
func (c *VtyshBGPdConnector) ShowRoutesWithBGPCommunityList(
	communityList string,
) (bgpd.BGP, error) {
	bgp := bgpd.BGP{}

	showCmd := fmt.Sprintf("show ip bgp community-list %s json", communityList)
	cmd := fmt.Sprintf("vtysh -H /dev/null -c '%s'", showCmd)

	c.Logger.Debug("execute command", "command", cmd, "callerFn", "ShowRoutesWithBGPCommunityList")
	out, err := bash.RunCommand(cmd)
	if err != nil {
		return bgp, fmt.Errorf("failed to show ip bgp community-list: %w", err)
	}

	if err := json.Unmarshal(out, &bgp); err != nil {
		return bgp, fmt.Errorf("failed to unmarchal to bgpd.BGP: %w", err)
	}

	return bgp, nil
}

// ConfigureRouteWithRouteMap implements BGPdConnector
func (c *VtyshBGPdConnector) ConfigureRouteWithRouteMap(
	prefix net.IPNet,
	routeMap string,
) error {
	configCommand := fmt.Sprintf("network %s route-map %s", prefix.String(), routeMap)
	cmd := fmt.Sprintf("vtysh -H /dev/null -c 'conf t' -c 'router bgp' -c '%s'", configCommand)

	c.Logger.Info("execute command", "command", cmd, "callerFn", "ConfigureUnicastRouteWithRouteMap")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to advertise %s route: %w", prefix.String(), err)
	}

	return nil
}
