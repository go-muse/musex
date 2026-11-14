package tempo

import (
	"math/big"
	"strconv"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
)

// The Maelzel metronome scale: ScaleSize divisions from ScaleMin to ScaleMax.
const (
	// ScaleMin is the lowest division of the scale.
	ScaleMin = 40
	// ScaleMax is the highest division of the scale.
	ScaleMax = 208
	// ScaleSize is the number of divisions of the scale.
	ScaleSize = 39
)

// scale holds the divisions of the metronome scale in ascending order: from 40
// to 60 by 2, to 72 by 3, to 120 by 4, to 144 by 6, and to 208 by 8.
var scale = [ScaleSize]int{
	40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60,
	63, 66, 69, 72,
	76, 80, 84, 88, 92, 96, 100, 104, 108, 112, 116, 120,
	126, 132, 138, 144,
	152, 160, 168, 176, 184, 192, 200, 208,
}

// ScaleDivisions returns the 39 divisions of the metronome scale in ascending
// order, from 40 to 208, as a new slice that belongs to the caller.
func ScaleDivisions() []int {
	out := make([]int, ScaleSize)
	copy(out, scale[:])
	return out
}

// ScaleDivision returns the division with the given index, numbered from 0:
// 40 for 0 and 208 for 38. It returns an error wrapping ErrInvalidArgument if
// index is outside 0..38.
func ScaleDivision(index int) (int, error) {
	if index < 0 || index >= ScaleSize {
		return 0, invalidError("ScaleDivision", "index "+strconv.Itoa(index)+" outside 0..38")
	}
	return scale[index], nil
}

// MustScaleDivision is like ScaleDivision but panics with the returned error.
func MustScaleDivision(index int) int {
	return must(ScaleDivision(index))
}

// checkBeats returns the error of an unset number of beats or of one that is
// zero or negative.
func checkBeats(op string, beats fraction.Fraction) error {
	if unsetFraction(beats) {
		return invalidError(op, "unset number of beats")
	}
	if beats.Sign() <= 0 {
		return simpleError(ErrNoTempo, op, "number of beats not positive")
	}
	return nil
}

// position returns, for a positive number of beats, the index of the highest
// division not greater than beats (−1 below the scale) and whether beats
// equals that division.
func position(beats fraction.Fraction) (int, bool) {
	b := ratOf(beats)
	i := -1
	for k, d := range scale {
		if b.Cmp(ratInt(int64(d))) < 0 {
			break
		}
		i = k
	}
	if i < 0 {
		return -1, false
	}
	return i, b.Cmp(ratInt(int64(scale[i]))) == 0
}

// IsScaleDivision reports whether beats equals a division of the scale: 100
// does, 110, 123/2, and 30 do not. In order of checking, it returns an error
// wrapping ErrInvalidArgument if beats is unset and ErrNoTempo if beats is
// zero or negative.
func IsScaleDivision(beats fraction.Fraction) (bool, error) {
	if err := checkBeats("IsScaleDivision", beats); err != nil {
		return false, err
	}
	_, on := position(beats)
	return on, nil
}

// MustIsScaleDivision is like IsScaleDivision but panics with the returned
// error.
func MustIsScaleDivision(beats fraction.Fraction) bool {
	return must(IsScaleDivision(beats))
}

// ScaleIndex returns the index of the division equal to beats: 21 for 100. In
// order of checking, it returns an error wrapping ErrInvalidArgument if beats
// is unset, ErrNoTempo if beats is zero or negative, and ErrOffScale if beats
// is not a division.
func ScaleIndex(beats fraction.Fraction) (int, error) {
	const op = "ScaleIndex"
	if err := checkBeats(op, beats); err != nil {
		return 0, err
	}
	i, on := position(beats)
	if !on {
		return 0, simpleError(ErrOffScale, op, "not a division")
	}
	return i, nil
}

// MustScaleIndex is like ScaleIndex but panics with the returned error.
func MustScaleIndex(beats fraction.Fraction) int {
	return must(ScaleIndex(beats))
}

// NextScaleDivision returns the smallest division strictly greater than
// beats: 63 after 60, 112 after 110, and 40 after 30. In order of checking, it
// returns an error wrapping ErrInvalidArgument if beats is unset, ErrNoTempo
// if beats is zero or negative, and ErrOffScale if beats is 208 or more.
func NextScaleDivision(beats fraction.Fraction) (int, error) {
	const op = "NextScaleDivision"
	if err := checkBeats(op, beats); err != nil {
		return 0, err
	}
	i, _ := position(beats)
	if i == ScaleSize-1 {
		return 0, simpleError(ErrOffScale, op, "no division above")
	}
	return scale[i+1], nil
}

// MustNextScaleDivision is like NextScaleDivision but panics with the returned
// error.
func MustNextScaleDivision(beats fraction.Fraction) int {
	return must(NextScaleDivision(beats))
}

// PrevScaleDivision returns the largest division strictly less than beats:
// 116 before 120, 108 before 110, and 208 before 250. In order of checking, it
// returns an error wrapping ErrInvalidArgument if beats is unset, ErrNoTempo
// if beats is zero or negative, and ErrOffScale if beats is 40 or less.
func PrevScaleDivision(beats fraction.Fraction) (int, error) {
	const op = "PrevScaleDivision"
	if err := checkBeats(op, beats); err != nil {
		return 0, err
	}
	i, on := position(beats)
	if on {
		i--
	}
	if i < 0 {
		return 0, simpleError(ErrOffScale, op, "no division below")
	}
	return scale[i], nil
}

// MustPrevScaleDivision is like PrevScaleDivision but panics with the returned
// error.
func MustPrevScaleDivision(beats fraction.Fraction) int {
	return must(PrevScaleDivision(beats))
}

// neighbors returns the index of the lower neighbor of a positive number of
// beats on the scale, the lower and upper neighbors, and cmp, or an error
// wrapping ErrOffScale if beats is outside 40..208.
func neighbors(op string, beats fraction.Fraction) (idx, lower, upper, cmp int, err error) {
	i, on := position(beats)
	if i < 0 || (i == ScaleSize-1 && !on) {
		return 0, 0, 0, 0, simpleError(ErrOffScale, op, "outside 40..208")
	}
	if on {
		return i, scale[i], scale[i], 0, nil
	}
	b := ratOf(beats)
	lo, hi := ratInt(int64(scale[i])), ratInt(int64(scale[i+1]))
	below := new(big.Rat).Sub(b, lo)
	above := new(big.Rat).Sub(hi, b)
	return i, scale[i], scale[i+1], below.Cmp(above), nil
}

// ScaleNeighbors returns the neighboring divisions of beats: lower, the
// largest division not greater than beats, upper, the smallest division not
// less than beats, and cmp, the sign of (beats − lower) − (upper − beats): −1
// if the lower is nearer, 0 for a tie, and +1 if the upper is nearer. The
// measure of nearness is the difference. On a division lower and upper are
// that division and cmp is 0. 62 gives 60, 63, +1; 110 gives 108, 112, 0;
// 121/2 gives 60, 63, −1. In order of checking, it returns an error wrapping
// ErrInvalidArgument if beats is unset, ErrNoTempo if beats is zero or
// negative, and ErrOffScale if beats is less than 40 or greater than 208.
func ScaleNeighbors(beats fraction.Fraction) (lower, upper, cmp int, err error) {
	const op = "ScaleNeighbors"
	if err := checkBeats(op, beats); err != nil {
		return 0, 0, 0, err
	}
	_, lower, upper, cmp, err = neighbors(op, beats)
	return lower, upper, cmp, err
}

// MustScaleNeighbors is like ScaleNeighbors but panics with the returned
// error.
func MustScaleNeighbors(beats fraction.Fraction) (lower, upper, cmp int) {
	return must3(ScaleNeighbors(beats))
}

// NearestScaleDivision returns the division chosen by mode: beats itself on a
// division in every mode, and otherwise one of the neighbors of
// ScaleNeighbors, with nearness measured by difference. The scale lies above
// zero, so TowardZero gives the same result as Down, AwayFromZero as Up,
// HalfTowardZero as HalfDown, and HalfAwayFromZero as HalfUp; HalfEven breaks
// a tie toward the neighbor with an even index, decided by the index of the
// lower neighbor. There is no default mode. 110 gives 112 with HalfUp and
// HalfEven and 108 with HalfDown and Down; 74 gives 72 with HalfEven. In order
// of checking, it returns an error wrapping ErrInvalidArgument if beats is
// unset, ErrNoTempo if beats is zero or negative, ErrInvalidArgument together
// with rounding.ErrUnset or rounding.ErrUnknownMode if mode is unset or
// unknown, and ErrOffScale if beats is less than 40 or greater than 208.
func NearestScaleDivision(beats fraction.Fraction, mode rounding.Mode) (int, error) {
	const op = "NearestScaleDivision"
	if err := checkBeats(op, beats); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	return nearest(op, beats, mode)
}

// MustNearestScaleDivision is like NearestScaleDivision but panics with the
// returned error.
func MustNearestScaleDivision(beats fraction.Fraction, mode rounding.Mode) int {
	return must(NearestScaleDivision(beats, mode))
}

// nearest returns the division chosen by the valid mode for a positive number
// of beats, or an error wrapping ErrOffScale.
func nearest(op string, beats fraction.Fraction, mode rounding.Mode) (int, error) {
	idx, lower, upper, cmp, err := neighbors(op, beats)
	if err != nil {
		return 0, err
	}
	if lower == upper {
		return lower, nil
	}
	n, err := mode.Choose(cmp, rounding.Positive, idx%2 == 0)
	if err != nil {
		return 0, modeError(op, mode)
	}
	if n == rounding.Lower {
		return lower, nil
	}
	return upper, nil
}
