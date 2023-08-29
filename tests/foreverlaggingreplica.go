package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func ForeverLaggingReplicaTest(sp *common.SystemParams) *testlib.TestCase {
	stateMachine := ForeverLaggingReplicaProperty()

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundTwoThirds)
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
			testlib.DropMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote).Or(common.IsMessageType(util.Precommit))).
				And(stateMachine.InState("allowCatchUp").Not()),
		).Then(
			testlib.DropMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageFromCurRound()),
		).Then(
			testlib.DeliverMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote).Or(common.IsMessageType(util.Precommit))).
				And(common.MessageCurRoundGt(2)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"LaggingReplica",
		25*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}

func ForeverLaggingReplicaProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()
	init.On(common.IsCommit(), "Committed")

	allowCatchUp := init.On(common.RoundReached(5), "allowCatchUp")
	allowCatchUp.On(
		common.IsCommit(),
		sm.SuccessStateLabel,
	)
	allowCatchUp.On(
		common.DiffCommits(),
		"DiffCommits",
	)
	return property
}
