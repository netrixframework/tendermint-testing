package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
)

func NotNilDecideTest(sp *common.SystemParams) *testlib.TestCase {
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
			sm.IsMessageSend().
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"NotNilDecide",
		2*time.Minute,
		NotNilDecideProperty(),
		filters,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}

func NotNilDecideProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()
	init.MarkSuccess()
	init.On(
		common.IsNilCommit(),
		"NilDecided",
	)
	return property
}
