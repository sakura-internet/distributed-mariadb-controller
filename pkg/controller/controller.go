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
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/bgpd"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/vtysh"
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
	currentNeighbors *neighborSet
	// currentMariaDBHealth holds the most recent healthcheck's result.
	currentMariaDBHealth dbHealthCheckResult
	// readyToPrimary
	readyToPrimary readyToPrimaryJudge

	// nftablesConnector communicates with FRRouting BGPd via vtysh.
	nftablesConnector nftables.Connector
	// bgpdConnector communicates with FRRouting BGPd via vtysh.
	bgpdConnector bgpd.BGPdConnector
	// systemdConnector manages the systemd services.
	systemdConnector systemd.Connector
	// mariaDBConnector communicates with MariaDB via mysql-client.
	mariaDBConnector mariadb.Connector
}

func NewController(
	logger *slog.Logger,
	globalInterfaceName string,
	hostAddress string,
	dbServingPort uint16,
	dbReplicaUserName string,
	dbReplicaPassword string,
	dbReplicaSourcePort uint16,
	dbAclChainName string,
	configs ...ControllerConfig,
) *Controller {
	c := &Controller{
		logger:              logger,
		globalInterfaceName: globalInterfaceName,
		hostAddress:         hostAddress,
		dbServingPort:       dbServingPort,
		dbReplicaUserName:   dbReplicaUserName,
		dbReplicaPassword:   dbReplicaPassword,
		dbReplicaSourcePort: dbReplicaSourcePort,
		dbAclChainName:      dbAclChainName,

		currentState:     StateInitial,
		currentNeighbors: newNeighborSet(),

		nftablesConnector: nftables.NewDefaultConnector(logger),
		bgpdConnector:     vtysh.NewDefaultBGPdConnector(logger),
		mariaDBConnector:  mariadb.NewDefaultConnector(logger),
		systemdConnector:  systemd.NewDefaultConnector(logger),
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
) {
	c.logger.Info("Hello, Starting db-controller.")

	ticker := time.NewTicker(ctrlerLoopInterval)
	defer ticker.Stop()

controllerLoop:
	for {
		select {
		case <-ctx.Done():
			c.onExit()
			break controllerLoop
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

func (c *Controller) onExit() error {
	c.setState(StateFault)
	if err := c.triggerRunOnStateChangesToFault(); err != nil {
		c.logger.Info("failed to TriggerRunOnStateChanges while going to fault. Ignore errors.", "error", err)
	}

	return nil
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
	prefixes, err := c.collectStateCommunityRoutePrefixes()
	if err != nil {
		return fmt.Errorf("failed to collect BGP routes: %w", err)
	}

	c.currentNeighbors = c.extractNeighborAddresses(prefixes)
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
	_, selfAddr, err := net.ParseCIDR(c.hostAddress + "/32")
	if err != nil {
		return err
	}
	return c.bgpdConnector.ConfigureRouteWithRouteMap(*selfAddr, string(c.GetState()))
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
	if err := c.systemdConnector.CheckServiceStatus(mariadb.SystemdSerivceName); err != nil {
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

// CollectStateCommunityRoutePrefixes collects the BGP route-prefix that they have a community of a controller-state.
func (c *Controller) collectStateCommunityRoutePrefixes() (map[State][]net.IP, error) {
	routes := make(map[State][]net.IP)

	// StateInitial is not needed in the below slice because the state doesn't advertise any routes.
	states := []State{
		StateCandidate,
		StateFault,
		StatePrimary,
		StateReplica,
		StateAnchor,
	}

	for _, state := range states {
		if routes[state] == nil {
			routes[state] = make([]net.IP, 0)
		}

		bgp, err := c.bgpdConnector.ShowRoutesWithBGPCommunityList(string(state))
		if err != nil {
			return nil, err
		}

		for routePrefix := range bgp.Routes {
			// NOTE: we recommend you use net/netip instead of net package
			//       because the netip.Addr is the most prefered way to present an IP address in Go.
			//       but the net/netip package doesn't have the way to parse CIDR notation.
			addr, _, err := net.ParseCIDR(routePrefix)
			if err != nil {
				c.logger.Error("failed to parse route prefix", "error", err)
			}

			routes[state] = append(routes[state], addr)
		}
	}

	return routes, nil
}

// ExtractNeighborAddresses get only the addresses of the neighbors from the given prefixes.
func (c *Controller) extractNeighborAddresses(
	prefixMatrix map[State][]net.IP,
) *neighborSet {
	neighbors := newNeighborSet()

	for state, prefixes := range prefixMatrix {

		for _, prefix := range prefixes {

			// each prefix of the advertised BGP route is the unicast address of other DB instances.
			// if the route prefix(unicast IP) and my address are same,
			// the route is advertised from me so it should be ignored.
			if prefix.String() == c.hostAddress {
				continue
			}

			if neighbors.neighborMatrix[state] == nil {
				neighbors.neighborMatrix[state] = make([]neighbor, 0)
			}

			neighbors.neighborMatrix[state] = append(
				neighbors.neighborMatrix[state],
				neighbor{
					address: prefix.String(),
				},
			)
		}
	}

	return neighbors
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
	if err := c.systemdConnector.StartService(mariadb.SystemdSerivceName); err != nil {
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
