package additional

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func ProposePrevote(sp *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()

	init := stateMachine.Builder()
	init.On(
		sm.IsMessageSend().
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal")),
		sm.SuccessStateLabel,
	)

	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
			testlib.DeliverMessage(),
		),
	)
	testcase := testlib.NewTestCase(
		"ProposePrevote",
		30*time.Second,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
