package pipeline

import "math"

// Accumulator collects awards throughout pipeline execution.
// Each computer appends or modifies awards in the accumulator.
type Accumulator struct {
	awards []Award
}

// NewAccumulator creates an empty accumulator.
func NewAccumulator() *Accumulator {
	return &Accumulator{awards: make([]Award, 0)}
}

// Add appends an award to the accumulator.
func (a *Accumulator) Add(award Award) {
	a.awards = append(a.awards, award)
}

// ZeroByLabel zeros out (removes) all awards with a matching label.
// Used by modules 03 and 04 to apply penalties.
func (a *Accumulator) ZeroByLabel(label string) {
	for i := range a.awards {
		if a.awards[i].Label == label {
			a.awards[i].Points = 0
		}
	}
}

// ScaleByLabel multiplies all awards with the given label by factor.
func (a *Accumulator) ScaleByLabel(label string, factor float64) {
	for i := range a.awards {
		if a.awards[i].Label == label {
			a.awards[i].Points *= factor
		}
	}
}

// HasLabel returns true if any award with the given label exists and has points > 0.
func (a *Accumulator) HasLabel(label string) bool {
	for _, aw := range a.awards {
		if aw.Label == label && aw.Points > 0 {
			return true
		}
	}
	return false
}

// Awards returns a copy of the award list.
func (a *Accumulator) Awards() []Award {
	result := make([]Award, len(a.awards))
	copy(result, a.awards)
	return result
}

// Total returns the total points in the accumulator.
func (a *Accumulator) Total() float64 {
	var total float64
	for _, aw := range a.awards {
		total += aw.Points
	}
	return total
}

// Len returns the number of awards in the accumulator.
func (a *Accumulator) Len() int {
	return len(a.awards)
}

// roundHalfUp rounds x to n decimal places using half-up (school rounding).
func roundHalfUp(x float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(x*pow) / pow
}

// RoundHalfUp rounds all points in the accumulator to 2 decimal places.
func (a *Accumulator) RoundHalfUp() {
	for i := range a.awards {
		a.awards[i].Points = roundHalfUp(a.awards[i].Points, 2)
	}
}
