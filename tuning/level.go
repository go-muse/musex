package tuning

import (
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
)

// WithReferenceNote reexpresses the same pitch level through note. The new
// reference frequency is exactly the frequency currently assigned to note.
// Strict SoundsSame comparison can nevertheless change because of float64 rounding.
func (t *Tuning) WithReferenceNote(note pitch.Pitch) (*Tuning, error) {
	if !t.IsValid() || !note.IsValid() {
		return nil, ErrUnset
	}
	freq, err := t.Frequency(note)
	if err != nil {
		return nil, err
	}
	return &Tuning{
		sys:  t.sys,
		temp: t.temp,
		root: t.root,
		ref:  Reference{note: note, freq: freq},
	}, nil
}

// MustWithReferenceNote calls WithReferenceNote and panics with its error on failure.
func (t *Tuning) MustWithReferenceNote(note pitch.Pitch) *Tuning {
	result, err := t.WithReferenceNote(note)
	if err != nil {
		panic(err)
	}
	return result
}

// LevelDifference returns how many cents higher other's pitch level is than
// t's. The system, temperament, and root must be equal. Frequencies of pitches
// are not computed, so a large reference-note displacement does not cause a
// false range error when the level difference itself is representable.
func (t *Tuning) LevelDifference(other *Tuning) (cents.Cents, error) {
	if !t.IsValid() || !other.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	if !t.sys.Equal(other.sys) || !t.temp.Equal(other.temp) || t.root != other.root {
		return cents.Cents{}, incomparable()
	}
	base := t.referenceOffset()
	otherBase := positiveMod(other.referenceClass().Number()-t.root.Number(), t.sys.Classes())
	difference, err := t.offsetDifference(otherBase, base)
	if err != nil {
		return cents.Cents{}, err
	}
	distance, err := t.sys.Distance(t.ref.note, other.ref.note)
	if err != nil {
		return cents.Cents{}, centsOutOfRange()
	}
	q, _ := decompose(base, int(distance), t.sys.Classes())
	between, err := cents.Between(t.ref.freq, other.ref.freq)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	shifted, err := between.Sub(difference)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	// Split the small frequency-distance part before subtracting the octaves.
	// This keeps every intermediate cents value in range whenever the final
	// level difference is in range, even if the reference height is not.
	baseQ, rest := shifted.SplitOctaves()
	remainingQ, ok := subInt(q, baseQ)
	if !ok {
		return cents.Cents{}, centsOutOfRange()
	}
	octaves, err := cents.Octave().Mul(remainingQ)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	level, err := rest.Sub(octaves)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	return level, nil
}

// MustLevelDifference calls LevelDifference and panics with its error on failure.
func (t *Tuning) MustLevelDifference(other *Tuning) cents.Cents {
	difference, err := t.LevelDifference(other)
	if err != nil {
		panic(err)
	}
	return difference
}

// LevelRatio returns the frequency multiplier from t's pitch level to other's.
// It returns ErrOutOfRange with cents.ErrOutOfRange if the multiplier is not a
// normal finite positive float64.
func (t *Tuning) LevelRatio(other *Tuning) (float64, error) {
	difference, err := t.LevelDifference(other)
	if err != nil {
		return 0, err
	}
	ratio, err := difference.Multiplier()
	if err != nil {
		return 0, centsError(err)
	}
	return ratio, nil
}

// MustLevelRatio calls LevelRatio and panics with its error on failure.
func (t *Tuning) MustLevelRatio(other *Tuning) float64 {
	ratio, err := t.LevelRatio(other)
	if err != nil {
		panic(err)
	}
	return ratio
}
