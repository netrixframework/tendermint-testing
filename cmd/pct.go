package cmd

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/netrixframework/netrix/config"
	"github.com/netrixframework/netrix/strategies"
	"github.com/netrixframework/netrix/strategies/pct"
	tests "github.com/netrixframework/tendermint-testing/testcases/invariant"
	"github.com/netrixframework/tendermint-testing/util"
	"github.com/spf13/cobra"
)

var pctStrategy = &cobra.Command{
	Use: "pct",
	RunE: func(cmd *cobra.Command, args []string) error {
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

		var strategy strategies.Strategy = pct.NewPCTStrategy(&pct.PCTStrategyConfig{
			RandSrc:        rand.NewSource(time.Now().UnixMilli()),
			MaxEvents:      1000,
			Depth:          10,
			RecordFilePath: "/home/nagendra/data/testing/tendermint/t",
		})

		strategy = strategies.NewStrategyWithProperty(strategy, tests.PrecommitInvariantProperty())

		driver := strategies.NewStrategyDriver(
			&config.Config{
				APIServerAddr: "127.0.0.1:7074",
				NumReplicas:   4,
				LogConfig: config.LogConfig{
					Format: "json",
					Level:  "info",
					Path:   "/home/nagendra/data/testing/tendermint/t/checker.log",
				},
			},
			&util.TMessageParser{},
			strategy,
			&strategies.StrategyConfig{
				Iterations:       1000,
				IterationTimeout: 40 * time.Second,
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
