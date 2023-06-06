package sakura

import (
	"net"
	"os"
	"sync"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/frrouting/vtysh"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/mariadb"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/process"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/systemd"
	"golang.org/x/exp/slog"
)

const (
	// SAKURAControllerStateCandidate is the controller state that is ready to be promoted to Primary state.
	SAKURAControllerStateCandidate controller.State = "Candidate"
	// SAKURAControllerStateAnchor indicates the controller is not exist and the node is under the anchor mode.
	// The State is not used in db-controller.
	SAKURAControllerStateAnchor controller.State = "Anchor"
	mariaDBServerDefaultIFName                   = "eth0"
	mariaDBServerDefaultPort                     = 3306
	nftablesMariaDBChain                         = "mariadb"
	mariaDBSerivceName                           = "mariadb"
)

var (
	SAKURAControllerAllStates = map[controller.State]bool{
		controller.StateInitial:        true,
		controller.StateFault:          true,
		controller.StatePrimary:        true,
		SAKURAControllerStateCandidate: true,
		controller.StateReplica:        true,
	}
)

type SAKURAController struct {
	Logger *slog.Logger
	// prevState is the previous state of the controller.
	prevState controller.State
	// m is a read-write-mutex that is used for sharing controller's state btw controller/http-api goroutines.
	m sync.RWMutex
	// selfAddr is an IP address of the eth0.
	selfAddr string
	// dbReplicaPassword is credential used by replica to establish replication link for primary
	dbReplicaPassword string
	// replicationStatusCheckFailCount is a counter of the MariaDB's replication status checker in replica state.
	replicationStatusCheckFailCount uint
	// writeTestDataFailCount is a counter that the controller tries to write test data to MariaDB.
	// if the count overs the pre-declared threshold, the controller urgently exits.
	writeTestDataFailCount uint
	// CurrentNeighbors holds the current BGP neighbors of the dbserver.
	// that discovered in each loop of the controller.
	CurrentNeighbors *NeighborSet
	// CurrentMariaDBHealth holds the most recent healthcheck's result.
	CurrentMariaDBHealth MariaDBHealthCheckResult
	// ReadyToPrimary
	ReadyToPrimary ReadyToPrimaryJudge

	// nftablesConnector communicates with FRRouting BGPd via vtysh.
	nftablesConnector nftables.Connector
	// currentState is the current state of the controller.
	// for prevending unexpected transition, the state isn't exposed.
	currentState controller.State
	// bgpdConnector communicates with FRRouting BGPd via vtysh.
	bgpdConnector vtysh.BGPdConnector
	// processControlConnector manages the linux process.
	processControlConnector process.ProcessControlConnector
	// systemdConnector manages the systemd services.
	systemdConnector systemd.Connector
	// mariaDBConnector communicates with MariaDB via mysql-client.
	mariaDBConnector mariadb.Connector
}

// GetState implements controller.Controller
func (c *SAKURAController) GetState() controller.State {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.currentState
}

// MakeDecision implements controller.Controller
func (c *SAKURAController) MakeDecision() controller.State {
	currentNeighbors := c.CurrentNeighbors
	currentMariaDBHealth := c.CurrentMariaDBHealth
	readyToPrimary := c.ReadyToPrimary
	switch c.GetState() {
	case controller.StateFault:
		return makeDecisionOnFault(currentNeighbors)
	case SAKURAControllerStateCandidate:
		return makeDecisionOnCandidate(
			c.Logger,
			currentNeighbors,
			currentMariaDBHealth,
			readyToPrimary)
		/*
			case StatePrimary:
				return c.decisionNextStateOnPrimary(ns, mariaDBHealth)
			case StateReplica:
				return c.decisionNextStateOnReplica(ns, mariaDBHealth)
		*/
	case controller.StateInitial:
		// just initialized controller take this case.
		return controller.StateFault
	default:
		panic("unreachable")
	}
}

// OnExit implements controller.Controller
func (c *SAKURAController) OnExit() error {
	c.SetState(controller.StateFault)
	if err := c.triggerRunOnStateChangesToFault(); err != nil {
		c.Logger.Info("failed to TriggerRunOnStateChanges while going to fault. Ignore errors.", err)
	}

	return nil
}

// OnStateHandler implements controller.Controller
func (c *SAKURAController) OnStateHandler(nextState controller.State) error {
	if cannotTransitionTo(c.GetState(), nextState) {
		panic("unreachable controller state was picked")
	}
	c.SetState(nextState)

	if c.keepStateInPrevTransition() {
		if err := c.triggerRunOnStateKeeps(); err != nil {
			slog.Error("failed to triggerRunOnStateKeeps. transition to fault state and exit", err, "state", string(c.GetState()))
			c.forceTransitionToFault()
			panic("urgently exit")
		}

		return nil
	}

	if err := c.triggerRunOnStateChanges(); err != nil {
		// we urgently transition to fault state
		slog.Error("failed to TriggerRunOnStateChanges. transition to fault state.", err, "state", string(c.GetState()))
		c.forceTransitionToFault()
	}

	return nil
}

// PreMakeDecisionHandler implements controller.Controller
func (c *SAKURAController) PreMakeDecisionHandler() error {
	prevNeighbors := c.CurrentNeighbors
	prefixes, err := c.collectStateCommunityRoutePrefixes()
	if err != nil {
		// we urgently transition to fault state
		slog.Error("failed to collect BGP routes", err, "state", c.GetState())
		c.forceTransitionToFault()

		return nil
	}

	c.CurrentNeighbors = c.extractNeighborAddresses(prefixes)
	// to avoiding unnecessary calculation, we checks the logger's level.
	if c.Logger.Enabled(slog.LevelInfo) {

		if prevNeighbors.Different(c.CurrentNeighbors) {
			addrs := c.CurrentNeighbors.NeighborAddresses()
			c.Logger.Info("neighbor set is updated", "addresses", addrs)
		}
	}

	c.CurrentMariaDBHealth = c.checkMariaDBHealth()
	c.ReadyToPrimary = c.readyToBePromotedToPrimary()
	return nil
}

// SetState sets the given state as the current state of the controller.
func (c *SAKURAController) SetState(nextState controller.State) {
	c.prevState = c.GetState()
	{
		c.m.Lock()
		c.currentState = nextState
		c.m.Unlock()
	}

	curState := c.GetState()
	if c.prevState == curState {
		c.Logger.Debug("controller transitions the state", "from", c.prevState, "to", curState)
	} else {
		c.Logger.Info("controller transitions the state", "from", c.prevState, "to", curState)
	}

	// modify state metric(s)
	DBControllerStateTransitionCounterVec.WithLabelValues(string(curState)).Inc()

	DBControllerStateGaugeVec.WithLabelValues(string(curState)).Set(1)

	for s := range SAKURAControllerAllStates {
		if s == curState {
			continue
		}

		DBControllerStateGaugeVec.WithLabelValues(string(s)).Set(0)
	}
}

func NewSAKURAController(logger *slog.Logger, configs ...ControllerConfig) *SAKURAController {
	c := &SAKURAController{
		Logger: logger,
	}

	for _, cfg := range configs {
		cfg(c)
	}
	return c
}

// triggerRunOnStateChanges triggers the state handler if the previous state is not the current state.
func (c *SAKURAController) triggerRunOnStateChanges() error {
	switch c.GetState() {
	case controller.StatePrimary:
		/*
			if err := c.triggerRunOnStateChangesToPrimary(currentNeighbors); err != nil {
				return err
			}
		*/
	case controller.StateFault:
		if err := c.triggerRunOnStateChangesToFault(); err != nil {
			return err
		}
	case SAKURAControllerStateCandidate:
		if err := c.triggerRunOnStateChangesToCandidate(); err != nil {
			return err
		}
	case controller.StateReplica:
		/*
			if err := c.triggerRunOnStateChangesToReplica(currentNeighbors); err != nil {
				return err
			}
		*/
	case SAKURAControllerStateAnchor:
		panic("unreachable")
	}

	return nil
}

// triggerRunOnStateKeeps triggers the state handler if the previous state is same as the current state.
func (c *SAKURAController) triggerRunOnStateKeeps() error {
	switch c.GetState() {
	case controller.StatePrimary:
		/*
			if err := c.triggerRunOnStateKeepsPrimary(ns); err != nil {
				return err
			}
		*/

	case controller.StateReplica:
		/*
			if err := c.triggerRunOnStateKeepsReplica(ns); err != nil {
				return err
			}
		*/
	}

	return nil
}

// advertiseSelfNetIFAddress updates the configuration of the advertising route.
// the BGP community of the advertising route will be updated with the current controller-state.
func (c *SAKURAController) advertiseSelfNetIFAddress() error {
	_, selfAddr, err := net.ParseCIDR(c.selfAddr + "/32")
	if err != nil {
		return err
	}
	return c.bgpdConnector.ConfigureRouteWithRouteMap(*selfAddr, string(c.GetState()))
}

// forceTransitionToFault set state to fault and triggers fault handler
func (c *SAKURAController) forceTransitionToFault() {
	c.SetState(controller.StateFault)
	if err := c.triggerRunOnStateChanges(); err != nil {
		slog.Info("failed to TriggerRunOnStateChanges while going to fault. Ignore errors.", err)
	}
}

// keepStateInPrevTransition determins wheather a state transition has occurred
func (c *SAKURAController) keepStateInPrevTransition() bool {
	prev := c.getPreviousState()
	cur := c.GetState()
	return prev == cur
}

// getPreviousState returns the controller's previous state.
func (c *SAKURAController) getPreviousState() controller.State {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.prevState
}

// MariaDBHealthCheckResult is the result of the mariadb's healthcheck
type MariaDBHealthCheckResult uint

const (
	MariaDBHealthCheckResultOK MariaDBHealthCheckResult = iota
	MariaDBHealthCheckResultNG
)

// checkMariaDBHealth checks whether the MariaDB server is healthy or not.
func (c *SAKURAController) checkMariaDBHealth() MariaDBHealthCheckResult {
	if err := c.systemdConnector.CheckServiceStatus(mariaDBSerivceName); err != nil {
		c.Logger.Debug("'systemctl status mariadb' exit with returning error", "error", err)
		return MariaDBHealthCheckResultNG
	}

	return MariaDBHealthCheckResultOK
}

// ReadyToPrimaryJudge is the result of the judgement to be promoted to primary state.
type ReadyToPrimaryJudge uint

const (
	// ReadytoPrimaryJudgeOK is OK for being promoted to primary state
	ReadytoPrimaryJudgeOK ReadyToPrimaryJudge = iota
	// ReadytoPrimaryJudgeNG is NG for being promoted to primary state
	ReadytoPrimaryJudgeNG
)

// readyToBePromotedToPrimary returns true when the controller satisfies the conditions to be promoted to primary state.
func (c *SAKURAController) readyToBePromotedToPrimary() ReadyToPrimaryJudge {
	status, err := c.mariaDBConnector.ShowReplicationStatus()
	if err != nil {
		slog.Debug("failed to show replication status", "error", err)
		return ReadytoPrimaryJudgeNG
	}

	readMasterLogPos, ok := status[mariadb.ReplicationStatusReadMasterLogPos]
	if !ok {
		return ReadytoPrimaryJudgeOK
	}

	if readMasterLogPos == status[mariadb.ReplicationStatusExecMasterLogPos] &&
		status[mariadb.ReplicationStatusMasterLogFile] == status[mariadb.ReplicationStatusRelayMasterLogFile] {
		return ReadytoPrimaryJudgeOK
	}

	return ReadytoPrimaryJudgeNG
}

// CollectStateCommunityRoutePrefixes collects the BGP route-prefix that they have a community of a controller-state.
func (c *SAKURAController) collectStateCommunityRoutePrefixes() (map[controller.State][]net.IP, error) {
	routes := make(map[controller.State][]net.IP)

	// StateInitial is not needed in the below slice because the state doesn't advertise any routes.
	states := []controller.State{
		SAKURAControllerStateCandidate,
		controller.StateFault,
		controller.StatePrimary,
		controller.StateReplica,
		SAKURAControllerStateAnchor,
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
				c.Logger.Error("failed to parse route prefix", err)
			}

			routes[state] = append(routes[state], addr)
		}

	}

	return routes, nil
}

// ExtractNeighborAddresses get only the addresses of the neighbors from the given prefixes.
func (c *SAKURAController) extractNeighborAddresses(
	prefixMatrix map[controller.State][]net.IP,
) *NeighborSet {
	neighbors := NewNeighborSet()

	for state, prefixes := range prefixMatrix {

		for _, prefix := range prefixes {

			// each prefix of the advertised BGP route is the unicast address of other DB instances.
			// if the route prefix(unicast IP) and my address are same,
			// the route is advertised from me so it should be ignored.
			if prefix.String() == c.selfAddr {
				continue
			}

			if neighbors.NeighborMatrix[state] == nil {
				neighbors.NeighborMatrix[state] = make([]Neighbor, 0)
			}

			neighbors.NeighborMatrix[state] = append(
				neighbors.NeighborMatrix[state],
				Neighbor{
					Address: prefix.String(),
				},
			)
		}
	}

	return neighbors
}

// syncReadOnlyVariable updates the read_only variable to the given expected value.
// if the current value equals the given value, the variable is already synced.
// otherwise, the function tries to sync the variable.
func (c *SAKURAController) syncReadOnlyVariable(readOnlyToBeTrue bool) error {
	isOn := c.mariaDBConnector.CheckBoolVariableIsON(mariadb.ReadOnlyVariableName)

	if readOnlyToBeTrue == isOn {
		// the variable is already the expected value.
		// nothing to do.
		return nil
	}

	if readOnlyToBeTrue {
		return c.mariaDBConnector.TurnOnBoolVariable(mariadb.ReadOnlyVariableName)
	}

	return c.mariaDBConnector.TurnOffBoolVariable(mariadb.ReadOnlyVariableName)
}

// startMariaDBService starts the mariadb service of systemd.
func (c *SAKURAController) startMariaDBService() error {
	const (
		mysqlMasterInfoFilePath = "/var/lib/mysql/master.info"
		mysqllRelayInfoFilePath = "/var/lib/mysql/relay-log.info"
	)

	preHook := func() error {
		if err := os.RemoveAll(mysqlMasterInfoFilePath); err != nil {
			return err
		}

		if err := os.RemoveAll(mysqllRelayInfoFilePath); err != nil {
			return err
		}

		return nil
	}

	if err := c.systemdConnector.StartService(mariaDBSerivceName, preHook, nil); err != nil {
		return err
	}

	return nil
}

// cannotTransitionTo checks whether the state machine doesn't have the edge from current to next.
func cannotTransitionTo(
	currentState controller.State,
	nextState controller.State,
) bool {
	switch currentState {
	case controller.StateFault:
		return nextState == controller.StatePrimary
	case SAKURAControllerStateCandidate:
		return nextState == controller.StateReplica
	case controller.StatePrimary:
		return nextState == SAKURAControllerStateCandidate || nextState == controller.StateReplica
	case controller.StateReplica:
		return nextState == controller.StatePrimary
	case controller.StateInitial:
		return nextState != controller.StateFault
	default:
		// unreachable
		return true
	}
}
