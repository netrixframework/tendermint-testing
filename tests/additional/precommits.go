package additional

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

// When quorum precommit and are delivered, you expect a decision
func QuorumPrecommits(sp *common.SystemParams) *testlib.TestCase {
	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("proposal"),
			testlib.DeliverMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Precommit)).
				And(common.IsVoteForProposal("proposal")),
		).Then(
			testlib.IncrCounter(sm.Count("precommitsSeen")),
		),
	)

	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()
	init.On(
		common.IsCommitForProposal("proposal"),
		sm.SuccessStateLabel,
	)
	init.On(
		sm.Count("precommitsSeen").Geq(2*sp.F+1),
		"quorumPrecommitsSeen",
	).On(
		common.IsCommitForProposal("proposal"),
		sm.SuccessStateLabel,
	)

	testcase := testlib.NewTestCase(
		"QuorumPrecommits",
		1*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
