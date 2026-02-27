package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// Runner executes the ordered sequence of computers for a single event.
type Runner struct {
	computers []Computer
}

// NewRunner creates a Runner with the given ordered list of computers.
func NewRunner(comps ...Computer) *Runner {
	return &Runner{computers: comps}
}

// Result captures the outcome of a pipeline run.
type Result struct {
	LogID       int64
	GKID        int64
	FinalAwards []FinalAward
	Halted      bool
	HaltReason  string
	Duration    time.Duration
}

// Run executes all computers in order for the given event.
// It returns the result of the run, including all awarded points.
func (r *Runner) Run(ctx context.Context, event Event) (*Result, error) {
	start := time.Now()

	pipeCtx := &Context{Event: event}
	acc := NewAccumulator()

	result := &Result{
		LogID: event.LogID,
		GKID:  event.GKID,
	}

	for _, comp := range r.computers {
		if err := comp.Process(ctx, pipeCtx, acc); err != nil {
			if IsHalt(err) {
				haltErr := err.(*HaltError)
				result.Halted = true
				result.HaltReason = haltErr.Reason
				result.Duration = time.Since(start)
				log.Debug().
					Int64("log_id", event.LogID).
					Str("computer", comp.Name()).
					Str("reason", haltErr.Reason).
					Msg("pipeline halted")
				return result, nil
			}
			return nil, fmt.Errorf("computer %s: %w", comp.Name(), err)
		}
	}

	// Collect final awards from the accumulator.
	// The aggregator (computer 14) stores them in pipeCtx after processing.
	result.FinalAwards = pipeCtx.AggregatedAwards
	result.Duration = time.Since(start)

	log.Debug().
		Int64("log_id", event.LogID).
		Int64("gk_id", event.GKID).
		Int("awards", len(result.FinalAwards)).
		Dur("duration", result.Duration).
		Msg("pipeline completed")

	return result, nil
}
