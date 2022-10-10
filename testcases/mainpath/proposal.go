package mainpath

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func ProposalNilPrevote(sp *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()

	init.On(
		sm.IsMessageSend().
			And(common.IsMessageFromRound(0)).
			And(common.IsVoteFromPart("h")).
			And(common.IsNotNilVote()),
		sm.FailStateLabel,
	)
	init.On(
		sm.IsMessageSend().
			And(common.IsMessageFromRound(0)).
			And(common.IsVoteFromPart("h")).
			And(common.IsNilVote()),
		sm.SuccessStateLabel,
	)

	cascade := testlib.NewFilterSet()

	cascade.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"ProposalNilPrevote",
		30*time.Second,
		stateMachine,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
