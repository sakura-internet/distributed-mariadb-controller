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

// BGP is an instance of the BGP speaker.
// NOTE: the struct is incomplete. this struct only contains the used fields.
type BGP struct {
	// Routes is the sequence of the BGP routes.
	// this key is notated with the CIDR-block.
	Routes map[string][]BGPRoute `json:"routes"`
}

// BGPRoute is a route that contains various attributes in BGP.
// NOTE: the struct is incomplete. this struct only contains the used fields.
type BGPRoute struct {
}
