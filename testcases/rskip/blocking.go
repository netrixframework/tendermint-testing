package rskip

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func BlockVotes(sysParams *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()
	init.MarkSuccess()
	init.On(common.IsCommit(), sm.FailStateLabel)
	init.On(common.IsEventNewRound(1), sm.FailStateLabel)

	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageType(util.Prevote)).
				And(sm.CountTo("votes").Lt(2*sysParams.F)),
		).Then(
			testlib.IncrCounter(sm.CountTo("votes")),
			testlib.DeliverMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageType(util.Prevote)).
				And(sm.CountTo("votes").Geq(2 * sysParams.F)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"BlockVotes",
		50*time.Second,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase
}
