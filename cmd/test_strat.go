package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/netrixframework/netrix/config"
	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/strategies"
	"github.com/netrixframework/netrix/strategies/unittest"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/testcases/rskip"
	"github.com/netrixframework/tendermint-testing/util"
	"github.com/spf13/cobra"
)

var testStrat = &cobra.Command{
	Use: "test",
	RunE: func(cmd *cobra.Command, args []string) error {
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

		var strategy strategies.Strategy = unittest.NewTestCaseStrategy(
			rskip.RoundSkip(common.NewSystemParams(4), 1, 2),
		)

		stateMachine := sm.NewStateMachine()
		roundReached := stateMachine.Builder().
			On(common.HeightReached(1), "SkipRounds").
			On(common.RoundReached(2), "roundReached")

		roundReached.MarkSuccess()
		roundReached.On(
			common.DiffCommits(),
			sm.FailStateLabel,
		)

		strategy = strategies.NewStrategyWithProperty(strategy, stateMachine)

		driver := strategies.NewStrategyDriver(
			&config.Config{
				APIServerAddr: "10.0.0.2:7074",
				NumReplicas:   4,
				LogConfig: config.LogConfig{
					Format: "json",
					Path:   "/Users/srinidhin/Local/data/testing/tendermint/t/checker.log",
				},
			},
			&util.TMessageParser{},
			strategy,
			&strategies.StrategyConfig{
				Iterations:       100,
				IterationTimeout: 45 * time.Second,
			},
		)

		go func() {
			<-termCh
			driver.Stop()
		}()
		return driver.Start()
	},
}
