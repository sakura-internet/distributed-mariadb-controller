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

package controller

import (
	"log/slog"
	"os"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
)

func _newFakeController() *Controller {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	c := NewController(
		logger,
		"dummy-global-interface-name",
		"10.0.0.1",
		3306,
		"repl",
		"dummy-db-replica-password",
		0,
		"dummy-chain-name",
		SystemdConnector(systemd.NewFakeSystemdConnector()),
		MariaDBConnector(mariadb.NewFakeMariaDBConnector()),
		NftablesConnector(nftables.NewFakeNftablesConnector()),
		BGPdConnector(bgpd.NewFakeBGPdConnector()),
	)

	return c
}
