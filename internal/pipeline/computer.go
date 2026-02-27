package pipeline

import "context"

// HaltError signals that the pipeline should stop processing cleanly.
// It is not a true error — the event guard returns it when an event is not
// eligible for scoring, causing the runner to halt without reporting failure.
type HaltError struct {
	Reason string
}

func (e *HaltError) Error() string {
	return "pipeline halted: " + e.Reason
}

// IsHalt returns true if err is a *HaltError.
func IsHalt(err error) bool {
	_, ok := err.(*HaltError)
	return ok
}

// Computer is the interface all scoring rule modules must implement.
// Each computer is stateless with respect to pipeline runs — all persistent
// state is accessed via a Store (injected at construction time).
type Computer interface {
	// Name returns a unique, human-readable identifier for the computer
	// (used in logging and award metadata).
	Name() string

	// Process executes the computer's scoring logic.
	//   ctx     – Go context for cancellation / deadline
	//   pipeCtx – shared pipeline context (event + state)
	//   acc     – accumulator where awards are appended / modified
	//
	// Returns:
	//   nil        – success, pipeline continues
	//   *HaltError – clean stop (event not eligible)
	//   other      – unexpected error, pipeline aborts
	Process(ctx context.Context, pipeCtx *Context, acc *Accumulator) error
}
