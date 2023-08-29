package tests

import (
	"math/rand"
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
)

func GarbledMessageTest(sysParams *common.SystemParams) *testlib.TestCase {
	filters := testlib.NewFilterSet()
	filters.AddFilter(
		testlib.If(
			sm.IsMessageSend().
				And(common.IsMessageFromPart("faulty")),
		).Then(
			garbleMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"GarbledMessages",
		2*time.Minute,
		GarbledMessageProperty(),
		filters,
	)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase
}

func GarbledMessageProperty() *sm.StateMachine {
	property := sm.NewStateMachine()
	property.Builder().On(
		common.IsCommit(),
		sm.SuccessStateLabel,
	)
	return property
}

func garbleMessage() testlib.Action {
	return func(e *types.Event, c *testlib.Context) []*types.Message {
		m, ok := c.GetMessage(e)
		if !ok {
			return []*types.Message{}
		}
		tMsg, ok := util.GetParsedMessage(m)
		if !ok {
			return []*types.Message{m}
		}
		randBytes := make([]byte, 100)
		rand.Read(randBytes)
		tMsg.MsgB = randBytes
		tMsg.Data = nil
		newMsg, err := tMsg.Marshal()
		if err != nil {
			return []*types.Message{m}
		}
		return []*types.Message{c.NewMessage(m, newMsg, tMsg)}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
