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
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"net/netip"
	"slices"
	"sync"
	"time"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bgpserver"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
)

// State specifies the controller state.
type State string

const (
	StateInitial   State = "initial"
	StateFault     State = "fault"
	StateCandidate State = "candidate"
	StatePrimary   State = "primary"
	StateReplica   State = "replica"
	StateAnchor    State = "anchor"
)

var (
	controllerAllStates = map[State]bool{
		StateInitial:   true,
		StateFault:     true,
		StatePrimary:   true,
		StateCandidate: true,
		StateReplica:   true,
	}
)

// dbHealthCheckResult is the result of the mariadb's healthcheck
type dbHealthCheckResult uint

const (
	dbHealthCheckResultOK dbHealthCheckResult = iota
	dbHealthCheckResultNG
)

// readyToPrimaryJudge is the result of the judgement to be promoted to primary state.
type readyToPrimaryJudge uint

const (
	// readytoPrimaryJudgeOK is OK for being promoted to primary state
	readytoPrimaryJudgeOK readyToPrimaryJudge = iota
	// readytoPrimaryJudgeNG is NG for being promoted to primary state
	readytoPrimaryJudgeNG
)

var (
	bgpCommunityFault     = bgpserver.MustParseCommunity("65000:1")
	bgpCommunityCandidate = bgpserver.MustParseCommunity("65000:2")
	bgpCommunityPrimary   = bgpserver.MustParseCommunity("65000:3")
	bgpCommunityReplica   = bgpserver.MustParseCommunity("65000:4")
	bgpCommunityAnchor    = bgpserver.MustParseCommunity("65000:10")
)

var (
	bgpCommunityToState = map[bgpserver.Community]State{
		bgpCommunityFault:     StateFault,
		bgpCommunityCandidate: StateCandidate,
		bgpCommunityPrimary:   StatePrimary,
		bgpCommunityReplica:   StateReplica,
		bgpCommunityAnchor:    StateAnchor,
	}
	stateToBgpCommunity = map[State]bgpserver.Community{
		StateFault:     bgpCommunityFault,
		StateCandidate: bgpCommunityCandidate,
		StatePrimary:   bgpCommunityPrimary,
		StateReplica:   bgpCommunityReplica,
		StateAnchor:    bgpCommunityAnchor,
	}
)

type Controller struct {
	logger *slog.Logger
	// globalInterfaceName is DB service interface name.
	globalInterfaceName string
	// hostAddress is an IP address of the global interface.
	hostAddress string
	// dbServingPort is the port number of database service
	dbServingPort uint16
	// dbReplicaUserName is the username for replication
	dbReplicaUserName string
	// dbReplicaSourcePort is the port of primary as replication source.
	dbReplicaSourcePort uint16
	// dbReplicaPassword is credential used by replica to establish replication link for primary
	dbReplicaPassword string
	// dbAclChainName is the nftables chain name for database access control.
	dbAclChainName string

	// currentState is the current state of the controller.
	// for prevending unexpected transition, the state isn't exposed.
	currentState State
	// prevState is the previous state of the controller.
	prevState State
	// m is a read-write-mutex that is used for sharing controller's state btw controller/http-api goroutines.
	m sync.RWMutex
	// replicationStatusCheckFailCount is a counter of the MariaDB's replication status checker in replica state.
	replicationStatusCheckFailCount uint
	// writeTestDataFailCount is a counter that the controller tries to write test data to MariaDB.
	// if the count overs the pre-declared threshold, the controller urgently exits.
	writeTestDataFailCount uint
	// currentNeighbors holds the current BGP neighbors of the dbserver.
	// that discovered in each loop of the controller.
	currentNeighbors neighborSet
	// currentMariaDBHealth holds the most recent healthcheck's result.
	currentMariaDBHealth dbHealthCheckResult
	// readyToPrimary
	readyToPrimary readyToPrimaryJudge

	// nftablesConnector communicates with nftables.
	nftablesConnector nftables.Connector
	// systemdConnector manages the systemd services.
	systemdConnector systemd.Connector
	// mariaDBConnector communicates with MariaDB via mysql-client.
	mariaDBConnector mariadb.Connector
	// bgpServerConnector communicates with gobgp
	bgpServerConnector bgpserver.Connector
}

func NewController(
	logger *slog.Logger,
	configs ...ControllerConfig,
) *Controller {
	c := &Controller{
		logger: logger,

		currentState:     StateInitial,
		currentNeighbors: newNeighborSet(),

		nftablesConnector:  nftables.NewDefaultConnector(logger),
		mariaDBConnector:   mariadb.NewDefaultConnector(logger),
		systemdConnector:   systemd.NewDefaultConnector(logger),
		bgpServerConnector: bgpserver.NewDefaultConnector(logger),
	}

	for _, cfg := range configs {
		cfg(c)
	}
	return c
}

// Start starts the controller loop.
// the function recognizes a done signal from the given context.
func (c *Controller) Start(
	ctx context.Context,
	ctrlerLoopInterval time.Duration,
) error {
	c.logger.Debug("controller: start bgpserver")
	if err := c.bgpServerConnector.Start(); err != nil {
		return err
	}
	defer c.bgpServerConnector.Stop()

	ticker := time.NewTicker(ctrlerLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.forceTransitionToFault()
			return nil
		case <-ticker.C:
			// random sleep to avoid global synchronization
			time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))

			if err := c.preDecideNextStateHandler(); err != nil {
				c.logger.Error("preDecideNextStateHandler", "error", err, "state", string(c.GetState()))
				// we urgently transition to fault state
				c.forceTransitionToFault()
				continue
			}

			nextState := c.decideNextState()
			c.logger.Debug("controller decided next state", "next state", nextState)

			if err := c.onStateHandler(nextState); err != nil {
				c.logger.Error("onStateHandler", "error", err, "next state", nextState)
			}
		}
	}
}

// GetState returns the current state of the controller.
func (c *Controller) GetState() State {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.currentState
}

// decideNextState determines next state that the controller should transition.
func (c *Controller) decideNextState() State {
	c.logger.Debug("decide next state", "current state", c.GetState())
	if c.currentNeighbors.isNetworkParted() {
		c.logger.Info("detected network partition", "neighbors", c.currentNeighbors.neighborAddresses())
		return StateFault
	}

	switch c.GetState() {
	case StateFault:
		return c.decideNextStateOnFault()
	case StateCandidate:
		return c.decideNextStateOnCandidate()
	case StatePrimary:
		return c.decideNextStateOnPrimary()
	case StateReplica:
		return c.decideNextStateOnReplica()
	case StateInitial:
		// just initialized controller take this case.
		return StateFault
	default:
		panic("unreachable")
	}
}

func (c *Controller) onStateHandler(nextState State) error {
	if cannotTransitionTo(c.GetState(), nextState) {
		panic("unreachable controller state was picked")
	}
	c.setState(nextState)

	if c.keepStateInPrevTransition() {
		if err := c.triggerRunOnStateKeeps(); err != nil {
			c.logger.Error("failed to triggerRunOnStateKeeps. transition to fault state and exit", "error", err, "state", string(c.GetState()))
			c.forceTransitionToFault()
			panic("urgently exit")
		}

		return nil
	}

	if err := c.triggerRunOnStateChanges(); err != nil {
		// we urgently transition to fault state
		c.logger.Error("failed to TriggerRunOnStateChanges. transition to fault state.", "error", err, "state", string(c.GetState()))
		c.forceTransitionToFault()
	}

	return nil
}

// preDecideNextStateHandler is triggered before calling MakeDecision()
func (c *Controller) preDecideNextStateHandler() error {
	prevNeighbors := c.currentNeighbors
	routes, err := c.bgpServerConnector.ListPath()
	if err != nil {
		return err
	}

	currentNeighbors := newNeighborSet()
	for _, route := range routes {
		state, ok := bgpCommunityToState[route.Community]
		if !ok {
			// ignore route with unknown community
			slog.Warn("unknown community", "community", route.Community)
			continue
		}

		if route.Prefix.Bits() != 32 {
			// ignore route with unknown prefix length
			slog.Warn("prefix length must be 32", "prefixlength", route.Prefix.Bits())
			continue
		}
		addr := route.Prefix.Addr().String()

		// skip self originated route
		if addr == c.hostAddress {
			continue
		}

		if !slices.Contains(currentNeighbors[state], neighbor(addr)) {
			currentNeighbors[state] = append(currentNeighbors[state], neighbor(addr))
		}
	}
	c.currentNeighbors = currentNeighbors

	// to avoiding unnecessary calculation, we checks the logger's level.
	if prevNeighbors.different(c.currentNeighbors) {
		addrs := c.currentNeighbors.neighborAddresses()
		c.logger.Info("neighbor set is updated", "addresses", addrs)
	}

	c.currentMariaDBHealth = c.checkMariaDBHealth()

	// judging "not ready" to primary when mariadb is not healthy
	if c.currentMariaDBHealth == dbHealthCheckResultNG {
		c.readyToPrimary = readytoPrimaryJudgeNG
		return nil
	}
	c.readyToPrimary = c.readyToBePromotedToPrimary()

	return nil
}

// setState sets the given state as the current state of the controller.
func (c *Controller) setState(nextState State) {
	c.prevState = c.GetState()
	{
		c.m.Lock()
		c.currentState = nextState
		c.m.Unlock()
	}

	if c.prevState == nextState {
		c.logger.Debug("controller transitions the state(unchanged)", "from", c.prevState, "to", nextState)
	} else {
		c.logger.Info("controller transitions the state(changed)", "from", c.prevState, "to", nextState)
	}

	// modify state metric(s)
	dbControllerStateTransitionCounterVec.WithLabelValues(string(nextState)).Inc()
	for s := range controllerAllStates {
		// clear flag of all state
		dbControllerStateGaugeVec.WithLabelValues(string(s)).Set(0)
	}
	// set flag of next state
	dbControllerStateGaugeVec.WithLabelValues(string(nextState)).Set(1)
}

// triggerRunOnStateChanges triggers the state handler if the previous state is not the current state.
func (c *Controller) triggerRunOnStateChanges() error {
	switch c.GetState() {
	case StatePrimary:
		return c.triggerRunOnStateChangesToPrimary()
	case StateFault:
		return c.triggerRunOnStateChangesToFault()
	case StateCandidate:
		return c.triggerRunOnStateChangesToCandidate()
	case StateReplica:
		return c.triggerRunOnStateChangesToReplica()
	}

	panic("unreachable")
}

// triggerRunOnStateKeeps triggers the state handler if the previous state is same as the current state.
func (c *Controller) triggerRunOnStateKeeps() error {
	switch c.GetState() {
	case StatePrimary:
		return c.triggerRunOnStateKeepsPrimary()
	case StateReplica:
		return c.triggerRunOnStateKeepsReplica()
	}

	return nil
}

// advertiseSelfNetIFAddress updates the configuration of the advertising route.
// the BGP community of the advertising route will be updated with the current controller-state.
func (c *Controller) advertiseSelfNetIFAddress() error {
	comm, ok := stateToBgpCommunity[c.GetState()]
	if !ok {
		return errors.New("unknown state")
	}
	addr, err := netip.ParseAddr(c.hostAddress)
	if err != nil {
		return err
	}

	prefix := netip.PrefixFrom(addr, 32)
	c.logger.Info("advertising my host address", "prefix", prefix, "community", comm)
	route := bgpserver.Route{
		Prefix:    prefix,
		Community: comm,
	}
	return c.bgpServerConnector.AddPath(route)
}

// forceTransitionToFault set state to fault and triggers fault handler
func (c *Controller) forceTransitionToFault() {
	// do nothing when already state is fault
	if c.GetState() == StateFault {
		return
	}

	c.setState(StateFault)
	if err := c.triggerRunOnStateChanges(); err != nil {
		c.logger.Info("failed to TriggerRunOnStateChanges while going to fault. Ignore errors.", "error", err)
	}
}

// keepStateInPrevTransition determins wheather a state transition has occurred
func (c *Controller) keepStateInPrevTransition() bool {
	prev := c.getPreviousState()
	cur := c.GetState()
	return prev == cur
}

// getPreviousState returns the controller's previous state.
func (c *Controller) getPreviousState() State {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.prevState
}

// checkMariaDBHealth checks whether the MariaDB server is healthy or not.
func (c *Controller) checkMariaDBHealth() dbHealthCheckResult {
	if err := c.systemdConnector.CheckServiceStatus(mariadb.SystemdServiceName); err != nil {
		c.logger.Debug("'systemctl status mariadb' exit with returning error", "error", err)
		return dbHealthCheckResultNG
	}

	return dbHealthCheckResultOK
}

// readyToBePromotedToPrimary returns true when the controller satisfies the conditions to be promoted to primary state.
func (c *Controller) readyToBePromotedToPrimary() readyToPrimaryJudge {
	status, err := c.mariaDBConnector.ShowReplicationStatus()
	if err != nil {
		c.logger.Debug("failed to show replication status", "error", err)
		return readytoPrimaryJudgeNG
	}

	readMasterLogPos, ok := status[mariadb.ReplicationStatusReadMasterLogPos]
	if !ok {
		return readytoPrimaryJudgeOK
	}

	if readMasterLogPos == status[mariadb.ReplicationStatusExecMasterLogPos] &&
		status[mariadb.ReplicationStatusMasterLogFile] == status[mariadb.ReplicationStatusRelayMasterLogFile] {
		return readytoPrimaryJudgeOK
	}

	return readytoPrimaryJudgeNG
}

// syncReadOnlyVariable updates the read_only variable to the given expected value.
// if the current value equals the given value, the variable is already synced.
// otherwise, the function tries to sync the variable.
func (c *Controller) syncReadOnlyVariable(readOnlyToBeTrue bool) error {
	isOn := c.mariaDBConnector.IsReadOnly()

	if readOnlyToBeTrue == isOn {
		// the variable is already the expected value.
		// nothing to do.
		return nil
	}

	if readOnlyToBeTrue {
		return c.mariaDBConnector.TurnOnReadOnly()
	}

	return c.mariaDBConnector.TurnOffReadOnly()
}

// startMariaDBService starts the mariadb service of systemd.
func (c *Controller) startMariaDBService() error {
	if err := c.mariaDBConnector.RemoveMasterInfo(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.RemoveRelayInfo(); err != nil {
		return err
	}
	if err := c.systemdConnector.StartService(mariadb.SystemdServiceName); err != nil {
		return err
	}

	return nil
}

// cannotTransitionTo checks whether the state machine doesn't have the edge from current to next.
func cannotTransitionTo(
	currentState State,
	nextState State,
) bool {
	switch currentState {
	case StateFault:
		return nextState == StatePrimary
	case StateCandidate:
		return nextState == StateReplica
	case StatePrimary:
		return nextState == StateCandidate || nextState == StateReplica
	case StateReplica:
		return nextState == StatePrimary
	case StateInitial:
		return nextState != StateFault
	default:
		// unreachable
		return true
	}
}
