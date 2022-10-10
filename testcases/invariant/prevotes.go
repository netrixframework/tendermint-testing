package invariant

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func PrecommitsInvariant(sp *common.SystemParams) *testlib.TestCase {
	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
			testlib.DropMessage(),
		),
	)

	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()
	init.On(
		sm.IsMessageSend().
			And(common.IsMessageFromRound(0)).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteForProposal("zeroProposal")),
		sm.FailStateLabel,
	)
	init.MarkSuccess()

	testcase := testlib.NewTestCase(
		"PrecommitInvariant",
		1*time.Minute,
		stateMachine,
		filters,
	)
	return testcase
}
