package tuning

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// NearestMode is the rounding mode of Nearest and NearestPitch: the nearest
// step wins, and an exact midpoint resolves to the lower step.
const NearestMode = rounding.HalfDown

// Nearest returns the step nearest to freq using NearestMode and the signed
// deviation freq-height in exact cents. A positive deviation means freq is
// higher than the selected step.
func (t *Tuning) Nearest(freq float64) (system.Step, cents.Cents, error) {
	return t.NearestBy(freq, NearestMode)
}

// MustNearest calls Nearest and panics with its error on failure.
func (t *Tuning) MustNearest(freq float64) (system.Step, cents.Cents) {
	step, deviation, err := t.Nearest(freq)
	if err != nil {
		panic(err)
	}
	return step, deviation
}

// NearestBy returns the step nearest to freq using mode to choose between the
// two adjacent steps. An unset mode returns ErrUnset; an unknown mode returns
// ErrInvalidArgument. Any finite positive frequency, including a subnormal one,
// is accepted.
func (t *Tuning) NearestBy(freq float64, mode rounding.Mode) (system.Step, cents.Cents, error) {
	if !t.IsValid() || mode == 0 {
		return 0, cents.Cents{}, ErrUnset
	}
	if !mode.IsValid() || !cents.IsValidFrequency(freq) {
		return 0, cents.Cents{}, invalidArgument()
	}
	step, deviation, err := t.nearestValid(freq, mode)
	if err != nil {
		return 0, cents.Cents{}, err
	}
	return system.Step(step), deviation, nil
}

// MustNearestBy calls NearestBy and panics with its error on failure.
func (t *Tuning) MustNearestBy(freq float64, mode rounding.Mode) (system.Step, cents.Cents) {
	step, deviation, err := t.NearestBy(freq, mode)
	if err != nil {
		panic(err)
	}
	return step, deviation
}

// NearestPitch returns the pitch nearest to freq using NearestMode and rule to
// select its spelling, together with the signed deviation in exact cents.
func (t *Tuning) NearestPitch(freq float64, rule system.SpellingRule) (pitch.Pitch, cents.Cents, error) {
	return t.NearestPitchBy(freq, NearestMode, rule)
}

// MustNearestPitch calls NearestPitch and panics with its error on failure.
func (t *Tuning) MustNearestPitch(freq float64, rule system.SpellingRule) (pitch.Pitch, cents.Cents) {
	result, deviation, err := t.NearestPitch(freq, rule)
	if err != nil {
		panic(err)
	}
	return result, deviation
}

// NearestPitchBy returns the pitch nearest to freq using mode and rule. Tuner
// values are checked before the spelling rule is applied.
func (t *Tuning) NearestPitchBy(freq float64, mode rounding.Mode, rule system.SpellingRule) (pitch.Pitch, cents.Cents, error) {
	if !t.IsValid() || mode == 0 || !rule.IsValid() {
		return pitch.Pitch{}, cents.Cents{}, ErrUnset
	}
	if !mode.IsValid() || !cents.IsValidFrequency(freq) {
		return pitch.Pitch{}, cents.Cents{}, invalidArgument()
	}
	step, deviation, err := t.nearestValid(freq, mode)
	if err != nil {
		return pitch.Pitch{}, cents.Cents{}, err
	}
	result, err := t.PitchOfStep(system.Step(step), rule)
	if err != nil {
		return pitch.Pitch{}, cents.Cents{}, err
	}
	return result, deviation, nil
}

// MustNearestPitchBy calls NearestPitchBy and panics with its error on failure.
func (t *Tuning) MustNearestPitchBy(freq float64, mode rounding.Mode, rule system.SpellingRule) (pitch.Pitch, cents.Cents) {
	result, deviation, err := t.NearestPitchBy(freq, mode, rule)
	if err != nil {
		panic(err)
	}
	return result, deviation
}

func (t *Tuning) nearestValid(freq float64, mode rounding.Mode) (int, cents.Cents, error) {
	distance, err := cents.Between(t.ref.freq, freq)
	if err != nil {
		return 0, cents.Cents{}, centsError(err)
	}
	base := t.referenceOffset()
	baseOffset, err := t.temp.Offset(base)
	if err != nil {
		return 0, cents.Cents{}, centsError(err)
	}
	x, err := distance.Add(baseOffset)
	if err != nil {
		return 0, cents.Cents{}, centsError(err)
	}
	nearest, nearestDeviation, err := t.temp.NearestCumulativeStep(x)
	if err != nil {
		return 0, cents.Cents{}, centsError(err)
	}
	lower := nearest - base
	if nearestDeviation.Sign() < 0 {
		lower--
	}
	lowerHeight, err := t.heightAtStep(lower)
	if err != nil {
		return 0, cents.Cents{}, err
	}
	upperHeight, err := t.heightAtStep(lower + 1)
	if err != nil {
		return 0, cents.Cents{}, err
	}
	lowerDistance, err := distance.Sub(lowerHeight)
	if err != nil {
		return 0, cents.Cents{}, centsError(err)
	}
	upperDistance, err := upperHeight.Sub(distance)
	if err != nil {
		return 0, cents.Cents{}, centsError(err)
	}
	if lowerDistance.Sign() == 0 {
		return lower, lowerDistance, nil
	}
	selected, err := mode.ChooseIndex(lower, lowerDistance.Compare(upperDistance))
	if err != nil {
		return 0, cents.Cents{}, invalidArgument()
	}
	if selected == lower {
		return selected, lowerDistance, nil
	}
	return selected, upperDistance.Neg(), nil
}
