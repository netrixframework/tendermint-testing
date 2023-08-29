package cmd

import (
	"time"

	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/testlib"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/tests"
)

func GetTest(name string, sysParams *common.SystemParams) (*testlib.TestCase, *sm.StateMachine) {
	switch name {
	case "ExpectUnlock":
		return tests.ExpectUnlockTest(sysParams), tests.ExpectUnlockProperty()
	case "Relocked":
		return tests.RelockedTest(sysParams), tests.RelockedProperty()
	case "LockedCommit":
		return tests.LockedCommitTest(sysParams), tests.LockedCommitProperty()
	case "LaggingReplica":
		return tests.LaggingReplicaTest(sysParams, 10, 10*time.Minute), tests.LaggingReplicaProperty(10)
	case "ForeverLaggingReplica":
		return tests.ForeverLaggingReplicaTest(sysParams), tests.ForeverLaggingReplicaProperty()
	case "RoundSkip":
		return tests.RoundSkipTest(sysParams, 1, 2), tests.RoundSkipProperty()
	case "BlockVotes":
		return tests.BlockVotesTest(sysParams), tests.BlockVotesProperty()
	case "PrecommitInvariant":
		return tests.PrecommitsInvariantTest(), tests.PrecommitInvariantProperty()
	case "CommitAfterRoundSkip":
		return tests.CommitAfterRoundSkipTest(sysParams), tests.CommitAfterRoundSkipProperty()
	case "DifferentDecisions":
		return tests.DifferentDecisionsTest(sysParams), tests.DifferentDecisionsProperty()
	case "NilPrevotes":
		return tests.NilPrevotesTest(sysParams), tests.NilPrevotesProperty(sysParams)
	case "ProposalNilPrevote":
		return tests.ProposalNilPrevoteTest(sysParams), tests.ProposalNilPrevoteProperty()
	case "NotNilDecide":
		return tests.NotNilDecideTest(sysParams), tests.NotNilDecideProperty()
	case "GarbledMessage":
		return tests.GarbledMessageTest(sysParams), tests.GarbledMessageProperty()
	case "HigherRound":
		return tests.HigherRoundTest(sysParams), tests.HigherRoundProperty()
	}
	return nil, nil
}
