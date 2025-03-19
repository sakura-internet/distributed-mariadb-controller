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

func TestDecideNextStateOnCandidate_MariaDBIsUnhealthy(t *testing.T) {
	c := _newFakeController()
	c.currentMariaDBHealth = dbHealthCheckResultNG
	c.readyToPrimary = readytoPrimaryJudgeNG

	nextState := c.decideNextStateOnCandidate()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnCandidate_InMultiCandidateSituation(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StateCandidate] = []neighbor{{}}
	c.currentMariaDBHealth = dbHealthCheckResultOK
	c.readyToPrimary = readytoPrimaryJudgeNG

	nextState := c.decideNextStateOnCandidate()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnCandidate_PrimaryIsAlreadyExist(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StatePrimary] = []neighbor{{}}
	c.currentMariaDBHealth = dbHealthCheckResultOK
	c.readyToPrimary = readytoPrimaryJudgeNG

	nextState := c.decideNextStateOnCandidate()
	assert.Equal(t, StateFault, nextState)
}

func TestDecideNextStateOnCandidate_ToBePromotedToPrimary(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StateFault] = []neighbor{{}}
	c.currentMariaDBHealth = dbHealthCheckResultOK
	c.readyToPrimary = readytoPrimaryJudgeOK

	nextState := c.decideNextStateOnCandidate()
	assert.Equal(t, StatePrimary, nextState)
}

func TestDecideNextStateCandidate_RemainCandidate(t *testing.T) {
	c := _newFakeController()
	c.currentNeighbors.neighborMatrix[StateFault] = []neighbor{{}}
	c.currentMariaDBHealth = dbHealthCheckResultOK
	c.readyToPrimary = readytoPrimaryJudgeNG

	nextState := c.decideNextStateOnCandidate()
	assert.Equal(t, StateCandidate, nextState)
}

func TestTriggerRunOnStateChangesToCandidate_OKPath(t *testing.T) {
	c := _newFakeController()

	err := c.triggerRunOnStateChangesToCandidate()
	assert.NoError(t, err)
}
