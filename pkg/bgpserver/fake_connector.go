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
	"fmt"
)

type FakeBgpServerConnector struct {
	RouteConfigured map[string]bool
}

func NewFakeBgpServerConnector() Connector {
	return &FakeBgpServerConnector{
		RouteConfigured: make(map[string]bool),
	}
}

func (bs *FakeBgpServerConnector) Start() error {
	return nil
}

func (bs *FakeBgpServerConnector) AddPath(prefix string, prefixLen uint32, _ string, _ uint32) error {
	p := fmt.Sprintf("%s/%d", prefix, prefixLen)
	bs.RouteConfigured[p] = true

	return nil
}

func (bs *FakeBgpServerConnector) ListPath() ([]Route, error) {
	return nil, nil
}

func (bs *FakeBgpServerConnector) Stop() {
}
