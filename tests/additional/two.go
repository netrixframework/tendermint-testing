package additional

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func ExpectNoUnlock(sysParams *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()

	roundOne := init.On(common.RoundReached(1), "RoundOne")

	roundOne.On(
		sm.IsMessageSend().
			And(common.IsMessageFromRound(1).Not()).
			And(common.IsMessageFromRound(0).Not()).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal")),
		sm.SuccessStateLabel,
	)
	roundOne.On(
		sm.IsMessageSend().
			And(common.IsMessageFromRound(1).Not()).
			And(common.IsMessageFromRound(0).Not()).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal").Not()),
		sm.FailStateLabel,
	)
	init.On(
		common.IsCommit(), sm.FailStateLabel,
	)

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundAll)
	// Change faulty replicas votes to nil
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(common.ChangeVoteToNil()),
	)
	// Record round 0 proposal
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
		),
	)
	// Do not deliver votes from "h"
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.StoreInSet(sm.Set("zeroDelayedPrevotes")),
			testlib.DropMessage(),
		),
	)
	// For higher rounds, we do not deliver proposal until we see a new one
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0).Not()).
				And(common.IsProposalEq("zeroProposal")),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"ExpectNoUnlock",
		1*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase
}
