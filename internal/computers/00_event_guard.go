package computers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// EventGuard is computer 00.
// It is the first gate in the pipeline: rejects non-scoreable events before
// any state is loaded or processing begins.
type EventGuard struct {
	store store.Store
}

// NewEventGuard creates a new EventGuard computer.
func NewEventGuard(s store.Store) *EventGuard {
	return &EventGuard{store: s}
}

// Name returns the computer's name.
func (c *EventGuard) Name() string {
	return "00_event_guard"
}

// Process implements the Computer interface.
func (c *EventGuard) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event

	// Condition 1 – Authenticated user only.
	if !event.HasUser() {
		return &HaltError{Reason: "anonymous moves earn 0 points"}
	}

	// Condition 2 – Scoreable log type only.
	if !event.LogType.IsScoreable() {
		return &HaltError{
			Reason: fmt.Sprintf("log type %s (%d) is not scoreable", event.LogType, int(event.LogType)),
		}
	}

	// Condition 3 – No duplicate processing.
	processed, err := c.store.IsEventProcessed(ctx, event.LogID)
	if err != nil {
		return fmt.Errorf("checking duplicate: %w", err)
	}
	if processed {
		log.Warn().Int64("log_id", event.LogID).Msg("duplicate event, already scored")
		return &HaltError{Reason: "duplicate event, already scored"}
	}

	return nil
}
