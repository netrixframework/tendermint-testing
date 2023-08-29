package tests

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

// Change vote to nil if we haven't seen new proposal, to the new proposal otherwise
func changeVote() testlib.Action {
	return func(e *types.Event, c *testlib.Context) []*types.Message {
		_, ok := c.Vars.Get("newProposalMessage")
		if !ok {
			return common.ChangeVoteToNil()(e, c)
		}
		return common.ChangeVoteToProposalMessage("newProposalMessage")(e, c)
	}
}

func RelockedTest(sysParams *common.SystemParams) *testlib.TestCase {

	stateMachine := RelockedProperty()

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundAll)
	// Change faulty replicas votes to nil if not seen new proposal
	// New proposal otherwise
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				// And(common.IsMessageType())
				And(common.IsVoteFromFaulty()),
		).Then(changeVote()),
	)
	// Record round 0 proposal
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
	// Do not deliver votes from "h".
	// This along with changing votes from faulty will ensure rounds are always skipped
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
				And(common.IsMessageType(util.Proposal)).
				And(common.IsProposalEq("zeroProposal")),
		).Then(
			testlib.DropMessage(),
		),
	)
	// Record the new proposal message
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0).Not()).
				And(common.IsMessageType(util.Proposal)).
				And(common.IsProposalEq("zeroProposal").Not()),
		).Then(
			common.RecordProposal("newProposal"),
			testlib.RecordMessageAs("newProposalMessage"),
			testlib.DeliverMessage(),
		),
	)

	testcase := testlib.NewTestCase("Relocking", 3*time.Minute, stateMachine, filters)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase

}

func RelockedProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()
	init.On(common.IsCommit(), "Committed")
	// We observe a precommit for round 0 proposal from replica "h"
	valueLocked := init.On(
		sm.IsMessageSend().
			And(common.IsVoteFromPart("h")).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteForProposal("zeroProposal")),
		"ValueLocked",
	)
	// Wait until all move to round 1
	roundOne := valueLocked.On(common.RoundReached(1), "RoundOne")
	// We observe a precommit for the new proposal from h
	roundOne.On(
		sm.IsMessageSend().
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("newProposal")),
		sm.SuccessStateLabel,
	)
	return property
}
