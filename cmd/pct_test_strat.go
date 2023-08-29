package cmd

import (
	"errors"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/netrixframework/netrix/config"
	"github.com/netrixframework/netrix/strategies"
	"github.com/netrixframework/netrix/strategies/pct"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
	"github.com/spf13/cobra"
)

func PCTTestStrategy() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "pct-test [test]",
		Long: "Run PCT with a test case to guide the exploration and measure outcomes",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			termCh := make(chan os.Signal, 1)
			signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

			sysParams := common.NewSystemParams(4)

			testCase, property := GetTest(args[0], sysParams)
			if testCase == nil || property == nil {
				return errors.New("invalid test")
			}

			var strategy strategies.Strategy = pct.NewPCTStrategyWithTestCase(
				&pct.PCTStrategyConfig{
					RandSrc:        rand.NewSource(time.Now().UnixMilli()),
					MaxEvents:      1000,
					Depth:          6,
					RecordFilePath: "results",
				},
				testCase,
			)

			strategy = strategies.NewStrategyWithProperty(strategy, property)

			driver := strategies.NewStrategyDriver(
				&config.Config{
					APIServerAddr: "127.0.0.1:7074",
					NumReplicas:   4,
					LogConfig: config.LogConfig{
						Format: "json",
						Level:  "info",
						Path:   "results/checker.log",
					},
				},
				&util.TMessageParser{},
				strategy,
				&strategies.StrategyConfig{
					Iterations:       60,
					IterationTimeout: 80 * time.Second,
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
	return cmd
}
