package ticks

import (
	"math"
	"strconv"
)

// Deltas returns the delta-times of non-decreasing absolute ticks: the first
// delta is absolute[0], then the differences of neighbors. Elements are
// checked in order; within an element a negative tick is ErrInvalidArgument, a
// tick less than the previous one ErrOrder and a delta above MaxDelta
// ErrFormatLimit. An empty input gives nil.
func Deltas(absolute []int) ([]int, error) {
	const op = "Deltas"
	if len(absolute) == 0 {
		return nil, nil
	}
	out := make([]int, len(absolute))
	previous := 0
	for i, tick := range absolute {
		at := "element " + strconv.Itoa(i)
		switch {
		case tick < 0:
			return nil, invalidError(op, at+" is negative")
		case tick < previous:
			return nil, newError(ErrOrder, op, at+" decreases")
		case tick-previous > MaxDelta:
			return nil, newError(ErrFormatLimit, op, at+" has a delta above 268435455")
		}
		out[i] = tick - previous
		previous = tick
	}
	return out, nil
}

// MustDeltas is like Deltas but panics with its error.
func MustDeltas(absolute []int) []int { return must(Deltas(absolute)) }

// Absolute returns the absolute ticks of delta-times: their running sums. A
// negative delta is ErrInvalidArgument, a delta above MaxDelta ErrFormatLimit
// and a sum outside int ErrOutOfRange. An empty input gives nil.
func Absolute(deltas []int) ([]int, error) {
	const op = "Absolute"
	if len(deltas) == 0 {
		return nil, nil
	}
	out := make([]int, len(deltas))
	sum := 0
	for i, delta := range deltas {
		at := "element " + strconv.Itoa(i)
		switch {
		case delta < 0:
			return nil, invalidError(op, at+" is negative")
		case delta > MaxDelta:
			return nil, newError(ErrFormatLimit, op, at+" is above 268435455")
		case sum > math.MaxInt-delta:
			return nil, intRangeError(op, at+" makes the sum exceed int")
		}
		sum += delta
		out[i] = sum
	}
	return out, nil
}

// MustAbsolute is like Absolute but panics with its error.
func MustAbsolute(deltas []int) []int { return must(Absolute(deltas)) }

// IsDelta reports whether delta can be written as an SMF delta-time:
// 0 <= delta <= MaxDelta.
func IsDelta(delta int) bool { return delta >= 0 && delta <= MaxDelta }
