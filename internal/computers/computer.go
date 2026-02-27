// Package computers contains all scoring rule implementations.
// Each file implements one pipeline.Computer for a single rule module (00–14).
//
// The pipeline.Computer interface and pipeline.HaltError type live in the
// pipeline package so that the runner can reference them without creating a
// circular import.  This file re-exports HaltError as a convenience alias so
// individual computers can write &HaltError{...} without a package qualifier.
package computers

import "github.com/geokrety/geokrety-points-system/internal/pipeline"

// HaltError is a convenience alias for pipeline.HaltError.
// Computer implementations can return &HaltError{Reason: "..."} to cleanly
// stop pipeline processing without propagating a real error.
type HaltError = pipeline.HaltError

// IsHalt reports whether err is a pipeline halt signal.
func IsHalt(err error) bool {
	return pipeline.IsHalt(err)
}
