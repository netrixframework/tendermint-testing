package additional

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func ExpectNewRound(sp *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()
	// We want replicas in partition "h" to move to round 1
	init.On(
		common.IsNewHeightRoundFromPart("h", 1, 1),
		sm.SuccessStateLabel,
	)
	newRound := init.On(
		sm.Count("round1ToH").Geq(sp.F+1),
		"newRoundMessagesDelivered",
	).On(
		common.IsNewHeightRoundFromPart("h", 1, 1),
		"NewRound",
	)
	newRound.On(
		common.DiffCommits(),
		sm.FailStateLabel,
	)
	newRound.On(
		common.IsCommit(),
		sm.SuccessStateLabel,
	)

	init.On(
		common.IsCommit(),
		sm.FailStateLabel,
	)

	filters := testlib.NewFilterSet()
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
			sm.IsMessageReceive().
				And(common.IsMessageFromRound(1)).
				And(common.IsMessageToPart("h")).
				And(
					common.IsMessageType(util.Proposal).
						Or(common.IsMessageType(util.Prevote)).
						Or(common.IsMessageType(util.Precommit)),
				),
		).Then(
			testlib.IncrCounter(sm.Count("round1ToH")),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote).Or(common.IsMessageType(util.Precommit))),
		).Then(
			testlib.DropMessage(),
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
		"ExpectNewRound",
		1*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
