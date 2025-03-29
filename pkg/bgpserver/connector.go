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

	gobgpapi "github.com/osrg/gobgp/v3/api"
	gobgpbgp "github.com/osrg/gobgp/v3/pkg/packet/bgp"
	gobgpserver "github.com/osrg/gobgp/v3/pkg/server"
	apb "google.golang.org/protobuf/types/known/anypb"
)

type Route struct {
	Prefix    string
	Community uint32
}

type Peer struct {
	Neighbor             string
	RemoteAS             uint32
	RemotePort           uint32
	KeepaliveIntervalSec uint64
}

type Connector interface {
	Start() error
	AddPath(prefix string, prefixLen uint32, nexthop string, community uint32) error
	ListPath() ([]Route, error)
	Stop()
}

type bgpServerConnector struct {
	logger     *slog.Logger
	server     *gobgpserver.BgpServer
	asn        uint32
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

func WithAsn(asn uint32) func(*bgpServerConnector) {
	return func(c *bgpServerConnector) {
		c.asn = asn
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
			Asn:             bs.asn,
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

			// route reflector client is always on
			RouteReflector: &gobgpapi.RouteReflector{
				RouteReflectorClient: true,
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

func (bs *bgpServerConnector) AddPath(prefix string, prefixLen uint32, nexthop string, community uint32) error {
	nlri1, err := apb.New(&gobgpapi.IPAddressPrefix{
		Prefix:    prefix,
		PrefixLen: prefixLen,
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
			NextHop: nexthop,
		})
		attrCommunities, _ := apb.New(&gobgpapi.CommunitiesAttribute{
			Communities: []uint32{community},
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
		for _, path := range d.Paths {
			for _, attr := range path.GetPattrs() {
				m, err := attr.UnmarshalNew()
				if err != nil {
					return
				}

				ca, ok := m.(*gobgpapi.CommunitiesAttribute)
				if !ok {
					continue
				}

				for _, comm := range ca.Communities {
					routes = append(routes, Route{
						Prefix:    d.Prefix,
						Community: comm,
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

// EncodeCommunity converts plain community value to human readable notation(for example 65001:10)
func EncodeCommunity(comm uint32) string {
	upper := comm >> 16
	lower := comm & 0xffff
	commDecoded := fmt.Sprintf("%d:%d", upper, lower)

	return commDecoded
}
