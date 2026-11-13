package tuning

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// StepOf returns the signed step number of p from the reference note.
func (t *Tuning) StepOf(p pitch.Pitch) (system.Step, error) {
	if !t.IsValid() || !p.IsValid() {
		return 0, ErrUnset
	}
	step, err := t.sys.Distance(t.ref.note, p)
	if err != nil {
		return 0, systemDistanceError(err)
	}
	return step, nil
}

// MustStepOf calls StepOf and panics with its error on failure.
func (t *Tuning) MustStepOf(p pitch.Pitch) system.Step {
	step, err := t.StepOf(p)
	if err != nil {
		panic(err)
	}
	return step
}

// PitchOfStep returns the pitch k steps from the reference note, using rule to
// select its spelling. A spelling rule is always required.
func (t *Tuning) PitchOfStep(k system.Step, rule system.SpellingRule) (pitch.Pitch, error) {
	if !t.IsValid() || !rule.IsValid() {
		return pitch.Pitch{}, ErrUnset
	}
	result, err := t.sys.Move(t.ref.note, k, rule)
	if err != nil {
		return pitch.Pitch{}, systemMoveError(err)
	}
	return result, nil
}

// MustPitchOfStep calls PitchOfStep and panics with its error on failure.
func (t *Tuning) MustPitchOfStep(k system.Step, rule system.SpellingRule) pitch.Pitch {
	result, err := t.PitchOfStep(k, rule)
	if err != nil {
		panic(err)
	}
	return result
}

// CentsFromReference returns the exact cent height of p relative to the
// reference note.
func (t *Tuning) CentsFromReference(p pitch.Pitch) (cents.Cents, error) {
	if !t.IsValid() || !p.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	step, err := t.StepOf(p)
	if err != nil {
		return cents.Cents{}, err
	}
	return t.heightAtStep(int(step))
}

// MustCentsFromReference calls CentsFromReference and panics with its error on failure.
func (t *Tuning) MustCentsFromReference(p pitch.Pitch) cents.Cents {
	height, err := t.CentsFromReference(p)
	if err != nil {
		panic(err)
	}
	return height
}

// CentsFromReferenceOfStep returns the exact cent height of step k relative to
// the reference note.
func (t *Tuning) CentsFromReferenceOfStep(k system.Step) (cents.Cents, error) {
	if !t.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	return t.heightAtStep(int(k))
}

// MustCentsFromReferenceOfStep calls CentsFromReferenceOfStep and panics with
// its error on failure.
func (t *Tuning) MustCentsFromReferenceOfStep(k system.Step) cents.Cents {
	height, err := t.CentsFromReferenceOfStep(k)
	if err != nil {
		panic(err)
	}
	return height
}

// CentsBetween returns the exact directed cent distance from one written pitch
// to another. The reference note and frequency do not affect the result.
func (t *Tuning) CentsBetween(from, to pitch.Pitch) (cents.Cents, error) {
	if !t.IsValid() || !from.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	fromClass, err := t.sys.PitchClass(from)
	if err != nil {
		return cents.Cents{}, systemDistanceError(err)
	}
	if !to.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	if _, err := t.sys.PitchClass(to); err != nil {
		return cents.Cents{}, systemDistanceError(err)
	}
	distance, err := t.sys.Distance(from, to)
	if err != nil {
		return cents.Cents{}, systemDistanceError(err)
	}
	base := positiveMod(fromClass.Number()-t.root.Number(), t.sys.Classes())
	q, m := decompose(base, int(distance), t.sys.Classes())
	return t.heightFromParts(q, m, base)
}

// MustCentsBetween calls CentsBetween and panics with its error on failure.
func (t *Tuning) MustCentsBetween(from, to pitch.Pitch) cents.Cents {
	distance, err := t.CentsBetween(from, to)
	if err != nil {
		panic(err)
	}
	return distance
}

// CentsBetweenSteps returns the exact directed cent distance between two step
// numbers relative to the reference note. It does not construct their separate
// heights, so a small distance remains valid when either height is out of range.
func (t *Tuning) CentsBetweenSteps(from, to system.Step) (cents.Cents, error) {
	if !t.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	base := t.referenceOffset()
	fromQ, fromM := decompose(base, int(from), t.sys.Classes())
	toQ, toM := decompose(base, int(to), t.sys.Classes())
	difference, err := t.offsetDifference(toM, fromM)
	if err != nil {
		return cents.Cents{}, err
	}
	q, ok := subInt(toQ, fromQ)
	if !ok {
		return cents.Cents{}, centsOutOfRange()
	}
	octaves, err := cents.Octave().Mul(q)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	result, err := octaves.Add(difference)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	return result, nil
}

// MustCentsBetweenSteps calls CentsBetweenSteps and panics with its error on failure.
func (t *Tuning) MustCentsBetweenSteps(from, to system.Step) cents.Cents {
	distance, err := t.CentsBetweenSteps(from, to)
	if err != nil {
		panic(err)
	}
	return distance
}

// Frequency returns the frequency assigned to p in hertz.
func (t *Tuning) Frequency(p pitch.Pitch) (float64, error) {
	if !t.IsValid() || !p.IsValid() {
		return 0, ErrUnset
	}
	height, err := t.CentsFromReference(p)
	if err != nil {
		return 0, err
	}
	return t.frequencyFromHeight(height)
}

// MustFrequency calls Frequency and panics with its error on failure.
func (t *Tuning) MustFrequency(p pitch.Pitch) float64 {
	freq, err := t.Frequency(p)
	if err != nil {
		panic(err)
	}
	return freq
}

// FrequencyOfStep returns the frequency assigned to step k in hertz.
func (t *Tuning) FrequencyOfStep(k system.Step) (float64, error) {
	if !t.IsValid() {
		return 0, ErrUnset
	}
	height, err := t.heightAtStep(int(k))
	if err != nil {
		return 0, err
	}
	return t.frequencyFromHeight(height)
}

// MustFrequencyOfStep calls FrequencyOfStep and panics with its error on failure.
func (t *Tuning) MustFrequencyOfStep(k system.Step) float64 {
	freq, err := t.FrequencyOfStep(k)
	if err != nil {
		panic(err)
	}
	return freq
}

// ShiftedFrequency returns the frequency of p after adding shift cents to its
// exact height. The height itself must be within the guaranteed cents range
// before the shift is added. A zero shift gives a bit-for-bit Frequency result.
func (t *Tuning) ShiftedFrequency(p pitch.Pitch, shift cents.Cents) (float64, error) {
	if !t.IsValid() || !p.IsValid() {
		return 0, ErrUnset
	}
	height, err := t.CentsFromReference(p)
	if err != nil {
		return 0, err
	}
	return t.shiftedFrequencyFromHeight(height, shift)
}

// MustShiftedFrequency calls ShiftedFrequency and panics with its error on failure.
func (t *Tuning) MustShiftedFrequency(p pitch.Pitch, shift cents.Cents) float64 {
	freq, err := t.ShiftedFrequency(p, shift)
	if err != nil {
		panic(err)
	}
	return freq
}

// ShiftedFrequencyOfStep returns the frequency of step k after adding shift
// cents to its exact height.
func (t *Tuning) ShiftedFrequencyOfStep(k system.Step, shift cents.Cents) (float64, error) {
	if !t.IsValid() {
		return 0, ErrUnset
	}
	height, err := t.heightAtStep(int(k))
	if err != nil {
		return 0, err
	}
	return t.shiftedFrequencyFromHeight(height, shift)
}

// MustShiftedFrequencyOfStep calls ShiftedFrequencyOfStep and panics with its
// error on failure.
func (t *Tuning) MustShiftedFrequencyOfStep(k system.Step, shift cents.Cents) float64 {
	freq, err := t.ShiftedFrequencyOfStep(k, shift)
	if err != nil {
		panic(err)
	}
	return freq
}

func (t *Tuning) shiftedFrequencyFromHeight(height, shift cents.Cents) (float64, error) {
	shifted, err := height.Add(shift)
	if err != nil {
		return 0, centsError(err)
	}
	return t.frequencyFromHeight(shifted)
}
