package testcases

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/util"
)

func handler(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
	if !e.IsMessageSend() {
		return []*types.Message{}, false
	}
	messageID, _ := e.MessageID()
	message, ok := c.MessagePool.Get(messageID)
	if ok {
		return []*types.Message{message}, true
	}
	return []*types.Message{}, true
}

func cond(e *types.Event, c *sm.Context) bool {
	if !e.IsMessageSend() {
		return false
	}

	message, ok := util.GetMessageFromEvent(e, c)
	if !ok {
		return false
	}
	return message.Type == util.Precommit
}

func DummyTestCaseStateMachine() *testlib.TestCase {
	stateMachine := sm.NewStateMachine()
	stateMachine.Builder().On(cond, sm.SuccessStateLabel)

	h := testlib.NewFilterSet()
	h.AddFilter(handler)

	testcase := testlib.NewTestCase("DummySM", 30*time.Second, stateMachine, h)
	return testcase
}
