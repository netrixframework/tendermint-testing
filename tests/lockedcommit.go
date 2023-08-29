package tests

import (
	"time"

	"github.com/netrixframework/netrix/log"
	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func changeProposalToNil(e *types.Event, c *testlib.Context) []*types.Message {
	message, _ := c.GetMessage(e)
	tMsg, ok := util.GetParsedMessage(message)
	if !ok {
		return []*types.Message{}
	}
	replica, _ := c.ReplicaStore.Get(message.From)
	newProp, err := util.ChangeProposalBlockIDToNil(replica, tMsg)
	if err != nil {
		c.Logger.With(log.LogParams{"error": err}).Error("Failed to change proposal")
		return []*types.Message{message}
	}
	newMsgB, err := newProp.Marshal()
	if err != nil {
		c.Logger.With(log.LogParams{"error": err}).Error("Failed to marshal changed proposal")
		return []*types.Message{message}
	}
	return []*types.Message{c.NewMessage(message, newMsgB, newProp)}
}

// States:
//  1. Ensure replicas skip round by not delivering enough precommits
//     1.1 One replica prevotes and precommits nil
//  2. In the next round change the proposal block value
//  3. Replicas should prevote and precommit the earlier block and commit
func LockedCommitTest(sysParams *common.SystemParams) *testlib.TestCase {

	filters := testlib.NewFilterSet()
	filters.AddFilter(common.TrackRoundTwoThirds)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	// Blanket change of all precommits in round 0 to nil,
	// We expect replicas to lock onto the proposal and this is just to ensure they move to the next round
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromRound(1)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			changeProposalToNil,
		),
	)

	testcase := testlib.NewTestCase("WrongProposal", 30*time.Second, LockedCommitProperty(), filters)
	testcase.SetupFunc(common.Setup(sysParams))

	return testcase
}

func LockedCommitProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	initialState := property.Builder()
	initialState.On(common.IsCommit(), "Committed")
	round1 := initialState.On(common.RoundReached(1), "round1")
	round1.On(common.IsCommit(), sm.SuccessStateLabel)
	return property
}
