package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func CommitAfterRoundSkipTest(sp *common.SystemParams) *testlib.TestCase {
	stateMachine := CommitAfterRoundSkipProperty()

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundAll)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.StoreInSet(sm.Set("delayedVotes")),
			testlib.DropMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().Not().
				And(stateMachine.InState("Round1")),
		).Then(
			testlib.DeliverAllFromSet(sm.Set("delayedVotes")),
			testlib.DeliverMessage(),
		),
	)
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
		"CommitAfterRoundSkip",
		2*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}

func CommitAfterRoundSkipProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()

	roundOne := init.On(
		common.RoundReached(1),
		"Round1",
	)
	roundOne.On(
		common.IsCommitForProposal("zeroProposal"),
		sm.SuccessStateLabel,
	)
	roundOne.On(
		common.DiffCommits(),
		sm.FailStateLabel,
	)
	return property
}
