package mainpath

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func NilPrevotes(sysParams *common.SystemParams) *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	init := stateMachine.Builder()

	nilQuorumDelivered := init.On(
		sm.Count("nilPrevotesDelivered").Geq(2*sysParams.F+1),
		"nilQuorumDelivered",
	)
	nilQuorumDelivered.On(
		sm.IsMessageSend().
			And(sm.IsMessageFromF(common.GetRandomReplica)).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsNilVote()),
		sm.SuccessStateLabel,
	)

	filters := testlib.NewFilterSet()
	// We don't deliver any proposal and hence we should see that replicas other than the proposer prevote nil.
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageType(util.Proposal)),
		).Then(
			testlib.DropMessage(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageReceive().
				And(sm.IsMessageToF(common.GetRandomReplica)).
				And(common.IsMessageType(util.Prevote)).
				And(common.IsNilVote()),
		).Then(
			testlib.IncrCounter(sm.Count("nilPrevotesDelivered")),
		),
	)

	testcase := testlib.NewTestCase(
		"NilPrevotes",
		1*time.Minute,
		stateMachine,
		filters,
	)
	testcase.SetupFunc(common.Setup(sysParams, common.PickRandomReplica()))
	return testcase
}
