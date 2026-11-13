package tuning

import (
	"errors"

	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/pitchclass"
)

// ChangedClasses returns, in increasing class-number order, the classes whose
// exact cent heights would change if the current root were replaced by root.
// The returned slice is never nil; an equal temperament produces an empty slice.
func (t *Tuning) ChangedClasses(root Root) ([]pitchclass.Class, error) {
	if !t.IsValid() || !root.IsValid() {
		return nil, ErrUnset
	}
	newRoot, err := resolveRoot(t.sys, root)
	if err != nil {
		return nil, err
	}
	changed := make([]pitchclass.Class, 0)
	for classNumber := range t.sys.Classes() {
		current, err := t.heightForClass(classNumber, t.root)
		if err != nil {
			return nil, err
		}
		updated, err := t.heightForClass(classNumber, newRoot)
		if err != nil {
			return nil, err
		}
		if !current.Equal(updated) {
			changed = append(changed, pitchclass.MustNew(classNumber, t.sys.Classes()))
		}
	}
	return changed, nil
}

// MustChangedClasses calls ChangedClasses and panics with its error on failure.
func (t *Tuning) MustChangedClasses(root Root) []pitchclass.Class {
	changed, err := t.ChangedClasses(root)
	if err != nil {
		panic(err)
	}
	return changed
}

// SoundsSame reports whether t and other assign exactly equal float64
// frequencies to every written pitch. It compares one octave-sized window with
// frequency.Equal and no tolerance. Different systems return false. An
// unrepresentable position or frequency in other also returns false; excessive
// exact-record complexity is the only non-unset error.
func (t *Tuning) SoundsSame(other *Tuning) (bool, error) {
	if !t.IsValid() || !other.IsValid() {
		return false, ErrUnset
	}
	if !t.sys.Equal(other.sys) {
		return false, nil
	}
	otherBase, err := other.sys.Distance(other.ref.note, t.ref.note)
	if err != nil {
		return false, nil
	}
	upward := t.ref.freq < 1
	for i := range t.sys.Classes() {
		k := -i
		if upward {
			k = i
		}
		heightA, err := t.heightAtStep(k)
		if err != nil {
			if errors.Is(err, ErrTooComplex) {
				return false, err
			}
			return false, nil
		}
		freqA, err := t.frequencyFromHeight(heightA)
		if err != nil {
			return false, nil
		}
		otherK, ok := addInt(int(otherBase), k)
		if !ok {
			return false, nil
		}
		heightB, err := other.heightAtStep(otherK)
		if err != nil {
			if errors.Is(err, ErrTooComplex) {
				return false, err
			}
			return false, nil
		}
		freqB, err := other.frequencyFromHeight(heightB)
		if err != nil {
			return false, nil
		}
		if !frequency.MustEqual(freqA, freqB) {
			return false, nil
		}
	}
	return true, nil
}

// MustSoundsSame calls SoundsSame and panics with its error on failure.
func (t *Tuning) MustSoundsSame(other *Tuning) bool {
	same, err := t.SoundsSame(other)
	if err != nil {
		panic(err)
	}
	return same
}
