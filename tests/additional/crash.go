package additional

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
)

func CrashReplica(sp *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()
	roundOne := init.On(
		common.RoundReached(1),
		"roundOne",
	)
	roundOne.On(
		common.IsCommit(),
		sm.SuccessStateLabel,
	)

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundTwoThirds)
	// Need to figure out a way around this
	// filters.AddFilter(
	// 	testlib.If(
	// 		testlib.Once(sm.InState("roundOne")),
	// 	).Then(
	// 		testlib.StopReplica(common.RandomReplicaFromPart("faulty")),
	// 	),
	// )
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
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

	testcase := testlib.NewTestCase(
		"CrashReplica",
		2*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
