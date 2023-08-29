package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func LaggingReplicaTest(sp *common.SystemParams, rounds int, timeout time.Duration) *testlib.TestCase {
	stateMachine := LaggingReplicaProperty(rounds)

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
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote).Or(common.IsMessageType(util.Precommit))).
				And(stateMachine.InState("allowCatchUp").Not()),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"LaggingReplica",
		timeout,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}

func LaggingReplicaProperty(rounds int) *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()
	init.On(common.IsCommit(), sm.FailStateLabel)

	allowCatchUp := init.On(common.RoundReached(rounds), "allowCatchUp")
	allowCatchUp.On(
		common.IsCommit(),
		sm.SuccessStateLabel,
	)
	allowCatchUp.On(
		common.DiffCommits(),
		"DifferentCommits",
	)
	return property
}
