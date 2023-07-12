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

package sakura

import (
	"os"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/process"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"golang.org/x/exp/slog"
)

func _newFakeSAKURAController() *SAKURAController {
	logger := slog.New(slog.NewJSONHandler(os.Stderr))
	c := NewSAKURAController(
		logger,
		SystemdConnector(systemd.NewFakeSystemdConnector()),
		MariaDBConnector(mariadb.NewFakeMariaDBConnector()),
		NftablesConnector(nftables.NewFakeNftablesConnector()),
		BGPdConnector(bgpd.NewFakeBGPdConnector()),
		ProcessControlConnector(process.NewFakeProcessControlConnector()),
	)

	c.HostAddress = "10.0.0.1"
	c.MariaDBReplicaPassword = "dummy-db-replica-password"
	return c
}
