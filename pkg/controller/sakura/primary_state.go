package sakura

import (
	"fmt"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/nftables"
	"golang.org/x/exp/slog"
)

// makeDecisionNextStateOnPrimary determines the next state on primary state
func makeDecisionNextStateOnPrimary(
	logger *slog.Logger,
	currentNeighbors *NeighborSet,
	mariaDBHealth MariaDBHealthCheckResult,
) controller.State {
	if mariaDBHealth == MariaDBHealthCheckResultNG {
		logger.Warn("MariaDB instance is down")
		return controller.StateFault
	}

	// found dual-primary situation.
	if currentNeighbors.primaryNodeExists() {
		logger.Warn("dual primary detected")
		return controller.StateFault
	}

	// won't transition to other state.
	return controller.StatePrimary
}

// triggerRunOnStateChangesToPrimary processes transition to primary state in main loop.
func (c *SAKURAController) triggerRunOnStateChangesToPrimary(
	ns *NeighborSet,
) error {
	if health := c.checkMariaDBHealth(); health == MariaDBHealthCheckResultNG {
		return fmt.Errorf("MariaDB instance is down")
	}
	if ns.primaryNodeExists() {
		return fmt.Errorf("dual primary detected")
	}

	// [STEP1]: START of setting MariaDB state
	if err := c.mariaDBConnector.StopReplica(); err != nil {
		return err
	}
	if err := c.mariaDBConnector.ResetAllReplicas(); err != nil {
		return err
	}
	if err := c.syncReadOnlyVariable( /* read_only=0 */ false); err != nil {
		return err
	}
	// [STEP1]: END of setting MariaDB state

	// [STEP2]: START of setting nftables state
	if err := c.acceptTCP3306TrafficFromSakuraPrefixes(); err != nil {
		return err
	}
	// [STEP2]: END of setting nftables state

	// [STEP3]: START of configurating frrouting
	if err := c.advertiseSelfNetIFAddress(); err != nil {
		return err
	}
	// [STEP3]: END of configurating frrouting

	// reset the count because the controller is healthy.
	c.writeTestDataFailCount = 0

	c.Logger.Info("primary state handler succeed")
	return nil
}

// acceptTCP3306TrafficFromSakuraPrefixes sets the rule that accepts the inbound communication from the SAKURA hosts.
func (c *SAKURAController) acceptTCP3306TrafficFromSakuraPrefixes() error {
	const (
		sakuraSrcAddrGroupName = "@sakura_addrs"
	)
	if err := c.nftablesConnector.FlushChain(nftables.BuiltinTableFilter, nftablesMariaDBChain); err != nil {
		return err
	}

	acceptMatches := []nftables.Match{
		nftables.IFNameMatch(mariaDBServerDefaultIFName),
		nftables.IPSrcAddrMatch(sakuraSrcAddrGroupName),
		nftables.TCPDstPortMatch(mariaDBServerDefaultPort),
	}

	if err := c.nftablesConnector.AddRule(nftables.BuiltinTableFilter, nftablesMariaDBChain, acceptMatches, nftables.AcceptStatement()); err != nil {
		return err
	}

	rejectMatches := []nftables.Match{
		nftables.IFNameMatch(mariaDBServerDefaultIFName),
		nftables.TCPDstPortMatch(mariaDBServerDefaultPort),
	}

	if err := c.nftablesConnector.AddRule(nftables.BuiltinTableFilter, nftablesMariaDBChain, rejectMatches, nftables.RejectStatement()); err != nil {
		return err
	}

	return nil
}
