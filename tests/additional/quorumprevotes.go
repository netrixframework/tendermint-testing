package additional

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func QuorumPrevotes(sysParams *common.SystemParams) *testlib.TestCase {

	filters := testlib.NewFilterSet()

	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("proposal"),
			testlib.DeliverMessage(),
		),
	)

	filters.AddFilter(
		testlib.If(
			sm.IsMessageReceive().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote)).
				And(common.IsVoteForProposal("proposal")),
		).Then(
			testlib.IncrCounter(sm.Count("prevotesDelivered")),
		),
	)

	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()

	quorumDelivered := init.On(
		sm.Count("prevotesDelivered").Geq(2*sysParams.F+1),
		"quorumDelivered",
	)
	quorumDelivered.On(
		sm.IsMessageSend().
			And(common.IsVoteFromPart("h")).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteForProposal("proposal")),
		sm.SuccessStateLabel,
	)

	testcase := testlib.NewTestCase(
		"QuorumPrevotes",
		1*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase
}
