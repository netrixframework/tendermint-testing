package cmd

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/netrixframework/netrix/config"
	"github.com/netrixframework/netrix/sm"
	"github.com/netrixframework/netrix/strategies"
	"github.com/netrixframework/netrix/strategies/pct"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/testcases/rskip"
	"github.com/netrixframework/tendermint-testing/util"
	"github.com/spf13/cobra"
)

var pctTestStrategy = &cobra.Command{
	Use: "pct-test",
	RunE: func(cmd *cobra.Command, args []string) error {
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

		stateMachine := sm.NewStateMachine()
		roundReached := stateMachine.Builder().
			On(common.HeightReached(1), "SkipRounds").
			On(common.RoundReached(2), "roundReached")

		roundReached.MarkSuccess()
		roundReached.On(
			common.DiffCommits(),
			sm.FailStateLabel,
		)

		var strategy strategies.Strategy = pct.NewPCTStrategyWithTestCase(
			&pct.PCTStrategyConfig{
				RandSrc:        rand.NewSource(time.Now().UnixMilli()),
				MaxEvents:      1000,
				Depth:          6,
				RecordFilePath: "/Users/srinidhin/Local/data/testing/tendermint/t",
			},
			rskip.RoundSkip(common.NewSystemParams(4), 1, 2),
			true,
		)

		strategy = strategies.NewStrategyWithProperty(strategy, stateMachine)

		driver := strategies.NewStrategyDriver(
			&config.Config{
				APIServerAddr: "10.0.0.2:7074",
				NumReplicas:   4,
				LogConfig: config.LogConfig{
					Format: "json",
					Level:  "info",
					Path:   "/Users/srinidhin/Local/data/testing/tendermint/t/checker.log",
				},
			},
			&util.TMessageParser{},
			strategy,
			&strategies.StrategyConfig{
				Iterations:       200,
				IterationTimeout: 45 * time.Second,
			},
		)

		go func() {
			<-termCh
			driver.Stop()
		}()

		if err := driver.Start(); err != nil {
			panic(err)
		}
		return nil
	},
}
