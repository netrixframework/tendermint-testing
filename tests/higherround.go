package tests

import (
	"bytes"
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func HigherRoundTest(sysParams *common.SystemParams) *testlib.TestCase {

	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(
			changeVoteRound(),
		),
	)

	testcase := testlib.NewTestCase(
		"HigherRound",
		2*time.Minute,
		HigherRoundProperty(),
		filters,
	)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase
}

func changeVoteRound() testlib.Action {
	return func(e *types.Event, c *testlib.Context) []*types.Message {
		m, ok := c.GetMessage(e)
		if !ok {
			return []*types.Message{}
		}
		tMsg, ok := util.GetParsedMessage(m)
		if !ok {
			return []*types.Message{m}
		}
		if tMsg.Type != util.Precommit && tMsg.Type != util.Prevote {
			return []*types.Message{}
		}
		valAddr, ok := util.GetVoteValidator(tMsg)
		if !ok {
			return []*types.Message{}
		}
		var replica *types.Replica = nil
		for _, r := range c.ReplicaStore.Iter() {
			addr, err := util.GetReplicaAddress(r)
			if err != nil {
				continue
			}
			if bytes.Equal(addr, valAddr) {
				replica = r
				break
			}
		}
		if replica == nil {
			return []*types.Message{}
		}
		newVote, err := util.ChangeVoteRound(replica, tMsg, int32(tMsg.Round()+2))
		if err != nil {
			return []*types.Message{}
		}
		msgB, err := newVote.Marshal()
		if err != nil {
			return []*types.Message{}
		}
		return []*types.Message{c.NewMessage(m, msgB, newVote)}
	}
}

func HigherRoundProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	init := property.Builder()
	init.On(
		common.IsCommit(),
		sm.SuccessStateLabel,
	)
	init.On(
		common.DiffCommits(),
		"DifferentCommits",
	)
	return property
}
