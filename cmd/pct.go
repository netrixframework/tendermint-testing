package cmd

import (
	"errors"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/netrixframework/netrix/config"
	"github.com/netrixframework/netrix/log"
	"github.com/netrixframework/netrix/strategies"
	"github.com/netrixframework/netrix/strategies/pct"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/common"
	"github.com/netrixframework/tendermint-testing/util"
	"github.com/spf13/cobra"
)

func PCTStrategy() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "pct [test]",
		Args: cobra.ExactArgs(1),
		Long: "Run PCT without any guidance and check if the property is satisfied for a specific test",
		RunE: func(cmd *cobra.Command, args []string) error {
			termCh := make(chan os.Signal, 1)
			signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

			_, property := GetTest(args[0], common.NewSystemParams(4))
			if property == nil {
				return errors.New("invalid test")
			}

			var strategy strategies.Strategy = pct.NewPCTStrategy(&pct.PCTStrategyConfig{
				RandSrc:        rand.NewSource(time.Now().UnixMilli()),
				MaxEvents:      1000,
				Depth:          10,
				RecordFilePath: "results",
			})

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
	return cmd
}

type records struct {
	duration     map[int][]time.Duration
	curStartTime time.Time
	timeSet      bool
	lock         *sync.Mutex
}

func newRecords() *records {
	return &records{
		duration: make(map[int][]time.Duration),
		lock:     new(sync.Mutex),
		timeSet:  false,
	}
}

func (r *records) stepFunc(e *types.Event, ctx *strategies.Context) {
	switch eType := e.Type.(type) {
	case *types.MessageSendEventType:
		messageID, _ := e.MessageID()
		message, ok := ctx.MessagePool.Get(messageID)
		if !ok {
			return
		}
		tMsg, ok := util.GetParsedMessage(message)
		if !ok {
			return
		}
		_, round := tMsg.HeightRound()
		r.lock.Lock()
		if tMsg.Type == util.Proposal &&
			round == 0 &&
			!r.timeSet {
			r.curStartTime = time.Now()
			r.timeSet = true
		}
		r.lock.Unlock()
	case *types.GenericEventType:
		if eType.T == "Committing block" {
			r.lock.Lock()
			if r.timeSet {
				_, ok := r.duration[ctx.CurIteration()]
				if !ok {
					r.duration[ctx.CurIteration()] = make([]time.Duration, 0)
				}
				r.duration[ctx.CurIteration()] = append(r.duration[ctx.CurIteration()], time.Since(r.curStartTime))
				r.timeSet = false
			}
			r.lock.Unlock()
		}
	}
}

func (r *records) finalize(ctx *strategies.Context) {
	sum := 0
	count := 0
	r.lock.Lock()
	for _, dur := range r.duration {
		for _, d := range dur {
			sum = sum + int(d)
			count = count + 1
		}
	}
	r.lock.Unlock()
	if count != 0 {
		iterations := len(r.duration)
		avg := time.Duration(sum / count)
		ctx.Logger.With(log.LogParams{
			"completed_runs":       iterations,
			"average_time":         avg.String(),
			"blocks_per_iteration": count / iterations,
		}).Info("Metrics")
	}
}
