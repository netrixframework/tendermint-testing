package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/netrixframework/netrix/config"
	"github.com/netrixframework/netrix/log"
	"github.com/netrixframework/netrix/strategies"
	"github.com/netrixframework/netrix/strategies/timeout"
	"github.com/netrixframework/netrix/types"
	"github.com/netrixframework/tendermint-testing/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
)

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

var strategyCmd = &cobra.Command{
	Use: "strat",
	RunE: func(cmd *cobra.Command, args []string) error {
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

		records := newRecords()

		dist := timeout.NewExpDistribution(1.5)
		dist.SetSrc(rand.NewSource(uint64(time.Now().UnixMilli())))

		strategy, err := timeout.NewTimeoutStrategy(&timeout.TimeoutStrategyConfig{
			Nondeterministic:  true,
			SpuriousCheck:     true,
			ClockDrift:        5,
			MaxMessageDelay:   100 * time.Millisecond,
			DelayDistribution: dist,
			// PendingEventThreshold: 20,
			RecordFilePath: "/Users/srinidhin/Local/data/testing/tendermint/run",
		})
		if err != nil {
			return err
		}
		driver := strategies.NewStrategyDriver(
			&config.Config{
				APIServerAddr: "172.23.37.208:7074",
				NumReplicas:   4,
				LogConfig: config.LogConfig{
					Format: "json",
					Level:  "info",
					Path:   "/Users/srinidhin/Local/data/testing/tendermint/run/checker.log",
				},
			},
			&util.TMessageParser{},
			strategy,
			&strategies.StrategyConfig{
				Iterations:       30,
				IterationTimeout: 40 * time.Second,
				StepFunc:         records.stepFunc,
				FinalizeFunc:     records.finalize,
			},
		)

		go func() {
			<-termCh
			driver.Stop()
		}()

		if err := driver.Start(); err != nil {
			panic(err)
		}
		return err
	},
}
