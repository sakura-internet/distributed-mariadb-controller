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

package bgpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"

	gobgpapi "github.com/osrg/gobgp/v3/api"
	gobgpbgp "github.com/osrg/gobgp/v3/pkg/packet/bgp"
	gobgpserver "github.com/osrg/gobgp/v3/pkg/server"
	apb "google.golang.org/protobuf/types/known/anypb"
)

const (
	// set dummy nexthop to advertise route because nexthop is meaningless.
	dummyBgpRouteNexthop = "192.0.2.1"
)

type Route struct {
	Prefix    netip.Prefix
	Community Community
}

type Peer struct {
	Neighbor             string
	RemoteAS             uint32
	RemotePort           uint32
	KeepaliveIntervalSec uint64
}

type Connector interface {
	Start() error
	AddPath(Route) error
	ListPath() ([]Route, error)
	Stop()
}

type bgpServerConnector struct {
	logger     *slog.Logger
	server     *gobgpserver.BgpServer
	localAsn   uint32
	routerId   string
	listenPort int32
	grpcPort   int
	peers      []Peer
}

func NewDefaultConnector(logger *slog.Logger, configs ...func(*bgpServerConnector)) Connector {
	bs := &bgpServerConnector{
		logger: logger,
	}
	for _, f := range configs {
		f(bs)
	}
	bs.server = gobgpserver.NewBgpServer(
		gobgpserver.GrpcListenAddress(fmt.Sprintf("127.0.0.1:%d", bs.grpcPort)),
	)

	return bs
}

func WithLocalAsn(localAsn uint32) func(*bgpServerConnector) {
	return func(c *bgpServerConnector) {
		c.localAsn = localAsn
	}
}

func WithRouterId(routerId string) func(*bgpServerConnector) {
	return func(c *bgpServerConnector) {
		c.routerId = routerId
	}
}

func WithListenPort(listenPort int32) func(*bgpServerConnector) {
	return func(c *bgpServerConnector) {
		c.listenPort = listenPort
	}
}

func WithGrpcPort(grpcPort int) func(*bgpServerConnector) {
	return func(c *bgpServerConnector) {
		c.grpcPort = grpcPort
	}
}

func WithPeers(peers []Peer) func(*bgpServerConnector) {
	return func(c *bgpServerConnector) {
		c.peers = peers
	}
}

func (bs *bgpServerConnector) Start() error {
	go bs.server.Serve()

	err := bs.server.StartBgp(context.Background(), &gobgpapi.StartBgpRequest{
		Global: &gobgpapi.Global{
			Asn:             bs.localAsn,
			RouterId:        bs.routerId,
			ListenAddresses: []string{"0.0.0.0"},
			ListenPort:      bs.listenPort,
		},
	})
	if err != nil {
		return err
	}

	for _, peer := range bs.peers {
		p := &gobgpapi.Peer{
			Conf: &gobgpapi.PeerConf{
				NeighborAddress: peer.Neighbor,
				PeerAsn:         peer.RemoteAS,
			},
			Transport: &gobgpapi.Transport{
				RemoteAddress: peer.Neighbor,
				RemotePort:    peer.RemotePort,
			},
			Timers: &gobgpapi.Timers{
				Config: &gobgpapi.TimersConfig{
					KeepaliveInterval: peer.KeepaliveIntervalSec,
					HoldTime:          peer.KeepaliveIntervalSec * 3,
				},
			},
			// ebgp multihop is always enabled
			EbgpMultihop: &gobgpapi.EbgpMultihop{
				Enabled:     true,
				MultihopTtl: 255,
			},
		}

		err := bs.server.AddPeer(
			context.Background(),
			&gobgpapi.AddPeerRequest{
				Peer: p,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bs *bgpServerConnector) AddPath(route Route) error {
	nlri1, err := apb.New(&gobgpapi.IPAddressPrefix{
		Prefix:    route.Prefix.Addr().String(),
		PrefixLen: uint32(route.Prefix.Bits()),
	})
	if err != nil {
		return err
	}

	var attrs []*apb.Any
	{
		attrOrigin, _ := apb.New(&gobgpapi.OriginAttribute{
			Origin: uint32(gobgpbgp.BGP_ORIGIN_ATTR_TYPE_IGP),
		})
		attrNextHop, _ := apb.New(&gobgpapi.NextHopAttribute{
			NextHop: dummyBgpRouteNexthop,
		})
		attrCommunities, _ := apb.New(&gobgpapi.CommunitiesAttribute{
			Communities: []uint32{
				uint32(route.Community),
			},
		})
		attrs = []*apb.Any{attrOrigin, attrNextHop, attrCommunities}
	}

	_, err = bs.server.AddPath(context.Background(), &gobgpapi.AddPathRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Path: &gobgpapi.Path{
			Family: &gobgpapi.Family{
				Afi:  gobgpapi.Family_AFI_IP,
				Safi: gobgpapi.Family_SAFI_UNICAST,
			},
			Nlri:   nlri1,
			Pattrs: attrs,
		},
	})

	return err
}

func (bs *bgpServerConnector) ListPath() ([]Route, error) {
	var routes []Route

	err := bs.server.ListPath(context.Background(), &gobgpapi.ListPathRequest{
		TableType: gobgpapi.TableType_GLOBAL,
		Family: &gobgpapi.Family{
			Afi:  gobgpapi.Family_AFI_IP,
			Safi: gobgpapi.Family_SAFI_UNICAST,
		},
		EnableFiltered: true,
	}, func(d *gobgpapi.Destination) {
		prefix, err := netip.ParsePrefix(d.Prefix)
		if err != nil {
			slog.Warn("ListPath: failed to ParsePrefix", "prefix", d.Prefix)
			return
		}
		for _, path := range d.Paths {
			for _, attr := range path.GetPattrs() {
				m, err := attr.UnmarshalNew()
				if err != nil {
					slog.Warn("ListPath: failed to attr.UnmarshalNew", "attr", attr)
					return
				}

				ca, ok := m.(*gobgpapi.CommunitiesAttribute)
				if !ok {
					continue
				}

				for _, comm := range ca.Communities {
					routes = append(routes, Route{
						Prefix:    prefix,
						Community: Community(comm),
					})
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (bs *bgpServerConnector) Stop() {
	bs.server.Stop()
}
