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
)

type records struct {
	duration     map[int]time.Duration
	curStartTime time.Time
	timeSet      bool
	lock         *sync.Mutex
}

func newRecords() *records {
	return &records{
		duration: make(map[int]time.Duration),
		lock:     new(sync.Mutex),
		timeSet:  false,
	}
}

func (r *records) stepFunc(e *types.Event, ctx *strategies.Context) {
	switch eType := e.Type.(type) {
	case *types.MessageSendEventType:
		messageID, _ := e.MessageID()
		message, ok := ctx.Messages.Get(messageID)
		if !ok {
			return
		}
		tMsg, ok := util.GetParsedMessage(message)
		if !ok {
			return
		}
		height, round := tMsg.HeightRound()
		r.lock.Lock()
		if tMsg.Type == util.Proposal &&
			height == 1 &&
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
				duration := time.Since(r.curStartTime)
				_, ok := r.duration[ctx.CurIteration()]
				if !ok {
					r.duration[ctx.CurIteration()] = duration
				}
				r.timeSet = false
			}
			r.lock.Unlock()
		}
	}
}

func (r *records) finalize(ctx *strategies.Context) {
	ctx.Logger.Info("Finalizing")
	sum := 0
	r.lock.Lock()
	for _, dur := range r.duration {
		sum = sum + int(dur)
	}
	count := len(r.duration)
	r.lock.Unlock()
	if count != 0 {
		avg := time.Duration(sum / count)
		ctx.Logger.With(log.LogParams{
			"completed_runs": count,
			"average_time":   avg.String(),
		}).Info("Metrics")
	}
}

var strategyCmd = &cobra.Command{
	Use: "strat",
	RunE: func(cmd *cobra.Command, args []string) error {
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

		records := newRecords()

		strategy, err := timeout.NewTimeoutStrategy(&timeout.TimeoutStrategyConfig{
			Nondeterministic: true,
			ClockDrift:       5,
			MaxMessageDelay:  200 * time.Millisecond,
		})
		if err != nil {
			return err
		}
		driver := strategies.NewStrategyDriver(
			&config.Config{
				APIServerAddr: "10.0.0.8:7074",
				NumReplicas:   4,
				LogConfig: config.LogConfig{
					Format: "json",
					Level:  "info",
					Path:   "/tmp/tendermint/log/checker.log",
				},
			},
			&util.TMessageParser{},
			strategy,
			&strategies.StrategyConfig{
				Iterations:       15,
				IterationTimeout: 40 * time.Second,
				StepFunc:         records.stepFunc,
				FinalizeFunc:     records.finalize,
			},
		)

		go func() {
			<-termCh
			driver.Stop()
		}()

		driver.Start()
		return nil
	},
}
