package common

import (
	"strconv"

	"github.com/netrixframework/netrix/log"
	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/util"
)

func IsCommit() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if ok && eType.T == "CommittingBlock" {
			blockID, ok := eType.Params["block_id"]
			if ok {
				c.Vars.Set(commitBlockIDKey, blockID)
			}
			c.Logger.Debug("Block committed")
			return true
		}
		return false
	}
}

func IsNilCommit() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if ok && eType.T == "CommittingBlock" {
			blockID, ok := eType.Params["block_id"]
			return ok && blockID == ""
		}
		return false
	}
}

func IsCommitForProposal(prop string) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		proposal, ok := c.Vars.GetString(prop)
		if !ok {
			return false
		}
		eType, ok := e.Type.(*types.GenericEventType)
		if ok && eType.T == "CommittingBlock" {
			blockID, ok := eType.Params["block_id"]
			return ok && blockID == proposal
		}
		return false
	}
}

func IsMessageFromRound(round int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		return m.Round() == round
	}
}

func IsConsensusMessage() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		return m.Round() != -1
	}
}

func IsVoteFromPart(partS string) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		if m.Type != util.Precommit && m.Type != util.Prevote {
			return false
		}
		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return false
		}
		val, ok := util.GetVoteValidator(m)
		if !ok {
			return false
		}
		return part.ContainsVal(val)
	}
}

func IsVoteFromFaulty() sm.Condition {
	return IsVoteFromPart("faulty")
}

func getPartition(c *sm.Context) (*util.Partition, bool) {
	p, exists := c.Vars.Get("partition")
	if !exists {
		return nil, false
	}
	partition, ok := p.(*util.Partition)
	return partition, ok
}

func IsMessageFromPart(partS string) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		if !e.IsMessageSend() && !e.IsMessageReceive() {
			return false
		}
		m, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return false
		}
		return part.Contains(m.From)
	}
}

func IsMessageToPart(partS string) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		if !e.IsMessageSend() && !e.IsMessageReceive() {
			return false
		}
		m, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return false
		}
		return part.Contains(m.To)
	}
}

func IsMessageType(t util.MessageType) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		tMessage, ok := util.GetParsedMessage(message)
		if !ok {
			return false
		}
		return tMessage.Type == t
	}
}

func IsNewHeightRoundFromPart(p string, h, r int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(p)
		if !ok {
			return false
		}
		return part.Contains(e.Replica) && IsNewHeightRound(h, r)(e, c)
	}
}

func IsNewHeightRound(h int, r int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if !ok {
			return false
		}
		if eType.T != "newStep" {
			return false
		}
		roundS, ok := eType.Params["round"]
		if !ok {
			return false
		}
		round, err := strconv.Atoi(roundS)
		if err != nil {
			return false
		}
		heightS, ok := eType.Params["height"]
		if !ok {
			return false
		}
		height, err := strconv.Atoi(heightS)
		if err != nil {
			return false
		}
		return height == h && round == r
	}
}

// RoundReached returns true if all replicas have reached the specified round
// Should be used with TrackRound handler!
func RoundReached(r int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		curRound, ok := c.Vars.GetInt(curRoundKey)
		if ok && curRound >= r {
			c.Logger.With(log.LogParams{"round": r}).Debug("Round reached")
			return true
		}
		return false
	}
}

func TwoFMinus1() func(*types.Event, *testlib.Context) (int, bool) {
	return func(e *types.Event, c *testlib.Context) (int, bool) {
		f, ok := c.Vars.GetInt("faults")
		if !ok {
			return 0, false
		}
		return 2*f + 1, true
	}
}

func IsVoteForProposal(proposalLabel string) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		proposal, ok := c.Vars.GetString(proposalLabel)
		if !ok {
			return false
		}
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		tMsg, ok := util.GetParsedMessage(message)
		if !ok {
			return false
		}
		voteBlockID, ok := util.GetVoteBlockIDS(tMsg)
		if !ok {
			return false
		}
		return voteBlockID == proposal
	}
}

func IsNilVote() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		tMsg, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		blockID, ok := util.GetVoteBlockIDS(tMsg)
		return ok && blockID == ""
	}
}

func IsNotNilVote() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		tMsg, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		blockID, ok := util.GetVoteBlockIDS(tMsg)
		return ok && blockID != ""
	}
}

func IsProposalEq(proposalLabel string) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		proposal, ok := c.Vars.GetString(proposalLabel)
		if !ok {
			return false
		}
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		tMsg, ok := util.GetParsedMessage(message)
		if !ok {
			return false
		}
		proposalBlockID, ok := util.GetProposalBlockIDS(tMsg)
		if !ok {
			return false
		}
		return proposalBlockID == proposal
	}
}

func IsFromHeight(height int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		return m.Height() == height
	}
}

func HeightReached(h int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if !ok {
			return false
		}
		if eType.T != "newStep" {
			return false
		}
		heightS, ok := eType.Params["height"]
		if !ok {
			return false
		}
		height, err := strconv.Atoi(heightS)
		if err != nil {
			return false
		}
		return height == h
	}
}

func IsEventNewRound(r int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if !ok {
			return false
		}
		if eType.T != "newStep" {
			return false
		}
		roundS, ok := eType.Params["round"]
		if !ok {
			return false
		}
		round, err := strconv.Atoi(roundS)
		if err != nil {
			return false
		}
		return round >= r
	}
}

func DiffCommits() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if ok && eType.T == "CommittingBlock" {
			blockID, ok := eType.Params["block_id"]
			if ok {
				curBlockID, exists := c.Vars.GetString(commitBlockIDKey)
				if !exists {
					c.Vars.Set(commitBlockIDKey, blockID)
					return false
				}
				return blockID != curBlockID
			}
		}
		return false
	}
}

func MessageCurRoundGt(m int) sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		tMsg, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		curRound, ok := c.Vars.GetInt(curRoundKey)
		if !ok {
			return false
		}
		return tMsg.Round() >= curRound-m
	}
}

func IsMessageFromCurRound() sm.Condition {
	return func(e *types.Event, c *sm.Context) bool {
		tMsg, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		curRound, ok := c.Vars.GetInt(curRoundKey)
		if !ok {
			return false
		}
		return tMsg.Round() == curRound
	}
}
