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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDifferent_Same(t *testing.T) {
	a := newNeighborSet()
	b := newNeighborSet()
	assert.False(t, a.different(b))
}

func TestDifferent_DiffLen(t *testing.T) {
	a := newNeighborSet()
	b := newNeighborSet()
	b[StateInitial] = []neighbor{""}
	assert.True(t, a.different(b))
}

func TestDifferent_DiffNeigh(t *testing.T) {
	a := newNeighborSet()
	a[StateInitial] = []neighbor{""}

	b := newNeighborSet()
	b[StateInitial] = []neighbor{"10.0.0.1"}
	assert.True(t, a.different(b))

}
