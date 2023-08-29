package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func RoundSkipTest(sysParams *common.SystemParams, height, round int) *testlib.TestCase {
	stateMachine := RoundSkipProperty()

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundAll)
	filters.AddFilter(
		testlib.If(
			common.IsFromHeight(height).Not(),
		).Then(
			testlib.DeliverMessage(),
		),
	)
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
			sm.OnceCondition("DeliverDelayed", stateMachine.InState("roundReached")),
		).Then(
			testlib.DeliverAllFromSet(sm.Set("DelayedPrevotes")),
			testlib.DeliverMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromPart("h")).
				And(common.IsMessageType(util.Prevote)),
		).Then(
			testlib.StoreInSet(sm.Set("DelayedPrevotes")),
			testlib.DropMessage(),
		),
	)

	testCase := testlib.NewTestCase(
		"RoundSkipWithPrevotes",
		1*time.Minute,
		stateMachine,
		filters,
	)
	testCase.SetupFunc(common.Setup(sysParams))
	return testCase
}

func RoundSkipProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	roundReached := property.Builder().
		On(common.HeightReached(1), "SkipRounds").
		On(common.RoundReached(2), "roundReached")

	roundReached.MarkSuccess()
	roundReached.On(
		common.DiffCommits(),
		sm.FailStateLabel,
	)
	return property
}
