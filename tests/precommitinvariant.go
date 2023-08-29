package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func PrecommitsInvariantTest() *testlib.TestCase {
	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"PrecommitInvariant",
		1*time.Minute,
		PrecommitInvariantProperty(),
		filters,
	)
	return testcase
}

func PrecommitInvariantProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()
	init.MarkSuccess()
	proposalReceived := init.On(
		sm.ConditionWithAction(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
			func(e *types.Event, ctx *sm.Context) {
				tMsg, ok := util.GetMessageFromEvent(e, ctx)
				if !ok {
					return
				}
				proposalS, ok := util.GetProposalBlockIDS(tMsg)
				if !ok {
					return
				}
				ctx.Vars.Set("zeroProposal", proposalS)
			},
		),
		"ProposalReceived",
	)
	proposalReceived.MarkSuccess()

	proposalReceived.On(
		sm.IsMessageSend().
			And(common.IsMessageFromRound(0)).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteForProposal("zeroProposal")),
		"ObservedVoteForProposal",
	)
	return property
}
