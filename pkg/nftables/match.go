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

package nftables

import (
	"strconv"
)

type Match []string

func IPSrcAddrMatch(srcAddr string) Match {
	return []string{"ip", "saddr", srcAddr}
}

func TCPDstPortMatch(dport uint16) Match {
	return []string{"tcp", "dport", strconv.Itoa(int(dport))}
}

func IFNameMatch(ifname string) Match {
	return []string{"iifname", ifname}
}
