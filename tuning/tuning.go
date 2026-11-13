package tuning

import (
	"strconv"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning/temperament"
)

// Tuning is an immutable musical tuning: a system, temperament, root class,
// and reference. It contains no derived fields or caches. The nil pointer and
// a pointer to the zero value are unset. Use New, NewAt, Standard, or a With
// method to create a valid tuning. Tuning values are safe for concurrent use.
type Tuning struct {
	sys  *system.System
	temp temperament.Temperament
	root pitchclass.Class
	ref  Reference
}

// New returns a tuning from its four parts. Failures are reported in this
// order: ErrUnset, ErrClassCount, ErrInvalidArgument for a root class of another
// modulus, and ErrInadmissible for a root name or reference pitch.
func New(sys *system.System, temp temperament.Temperament, root Root, ref Reference) (*Tuning, error) {
	if !sys.IsValid() || !temp.IsValid() || !root.IsValid() || !ref.IsValid() {
		return nil, ErrUnset
	}
	if temp.StepsPerOctave() != sys.Classes() {
		return nil, classCountMismatch()
	}
	rootClass, err := resolveRoot(sys, root)
	if err != nil {
		return nil, err
	}
	if err := admitPitch(sys, ref.note); err != nil {
		return nil, err
	}
	return &Tuning{sys: sys, temp: temp, root: rootClass, ref: ref}, nil
}

// MustNew calls New and panics with its error on failure.
func MustNew(sys *system.System, temp temperament.Temperament, root Root, ref Reference) *Tuning {
	t, err := New(sys, temp, root, ref)
	if err != nil {
		panic(err)
	}
	return t
}

// NewAt returns a tuning with a reference supplied as a note-frequency pair.
// Failures follow New, with ErrInvalidArgument for the frequency reported last.
func NewAt(sys *system.System, temp temperament.Temperament, root Root, note pitch.Pitch, freq float64) (*Tuning, error) {
	if !sys.IsValid() || !temp.IsValid() || !root.IsValid() || !note.IsValid() {
		return nil, ErrUnset
	}
	if temp.StepsPerOctave() != sys.Classes() {
		return nil, classCountMismatch()
	}
	rootClass, err := resolveRoot(sys, root)
	if err != nil {
		return nil, err
	}
	if err := admitPitch(sys, note); err != nil {
		return nil, err
	}
	if !cents.IsNormalFrequency(freq) {
		return nil, invalidArgument()
	}
	return &Tuning{sys: sys, temp: temp, root: rootClass, ref: Reference{note: note, freq: freq}}, nil
}

// MustNewAt calls NewAt and panics with its error on failure.
func MustNewAt(sys *system.System, temp temperament.Temperament, root Root, note pitch.Pitch, freq float64) *Tuning {
	t, err := NewAt(sys, temp, root, note, freq)
	if err != nil {
		panic(err)
	}
	return t
}

// Standard returns a fresh tuning with the traditional system, 12-EDO, root C,
// and A4 at 440 Hz. Each call returns a new pointer to an equal tuning.
func Standard() *Tuning {
	return &Tuning{
		sys:  system.Traditional(),
		temp: temperament.Equal12(),
		root: pitchclass.MustNew(0, system.Traditional().Classes()),
		ref:  A4At440(),
	}
}

// IsValid reports whether t is a fully specified tuning.
func (t *Tuning) IsValid() bool {
	return t != nil && t.sys.IsValid() && t.temp.IsValid() && t.root.IsValid() &&
		t.ref.IsValid() && t.temp.StepsPerOctave() == t.sys.Classes() &&
		t.root.Modulus() == t.sys.Classes()
}

// System returns the tuning's musical system, or nil for an unset tuning.
// The returned system is shared and immutable; do not overwrite it by
// dereferencing the pointer.
func (t *Tuning) System() *system.System {
	if t == nil {
		return nil
	}
	return t.sys
}

// Temperament returns the tuning's temperament, or its zero value if t is unset.
func (t *Tuning) Temperament() temperament.Temperament {
	if t == nil {
		return temperament.Temperament{}
	}
	return t.temp
}

// RootClass returns the root pitch class, or an unset class if t is unset.
func (t *Tuning) RootClass() pitchclass.Class {
	if t == nil {
		return pitchclass.Class{}
	}
	return t.root
}

// Reference returns the reference note-frequency pair, or an unset reference.
func (t *Tuning) Reference() Reference {
	if t == nil {
		return Reference{}
	}
	return t.ref
}

// Classes returns the number of classes per octave, or zero if t is unset.
func (t *Tuning) Classes() int {
	if t == nil || !t.sys.IsValid() {
		return 0
	}
	return t.sys.Classes()
}

// WithRoot returns a tuning with a different root and the same other parts.
func (t *Tuning) WithRoot(root Root) (*Tuning, error) {
	if !t.IsValid() || !root.IsValid() {
		return nil, ErrUnset
	}
	rootClass, err := resolveRoot(t.sys, root)
	if err != nil {
		return nil, err
	}
	return &Tuning{sys: t.sys, temp: t.temp, root: rootClass, ref: t.ref}, nil
}

// MustWithRoot calls WithRoot and panics with its error on failure.
func (t *Tuning) MustWithRoot(root Root) *Tuning {
	result, err := t.WithRoot(root)
	if err != nil {
		panic(err)
	}
	return result
}

// WithTemperament returns a tuning with a different temperament of the same
// class count and the same system, root, and reference.
func (t *Tuning) WithTemperament(temp temperament.Temperament) (*Tuning, error) {
	if !t.IsValid() || !temp.IsValid() {
		return nil, ErrUnset
	}
	if temp.StepsPerOctave() != t.sys.Classes() {
		return nil, classCountMismatch()
	}
	return &Tuning{sys: t.sys, temp: temp, root: t.root, ref: t.ref}, nil
}

// MustWithTemperament calls WithTemperament and panics with its error on failure.
func (t *Tuning) MustWithTemperament(temp temperament.Temperament) *Tuning {
	result, err := t.WithTemperament(temp)
	if err != nil {
		panic(err)
	}
	return result
}

// WithSystem returns a tuning with a different system of the same class count.
// The root number and reference are preserved; the reference pitch must be
// admissible in the new system.
func (t *Tuning) WithSystem(sys *system.System) (*Tuning, error) {
	if !t.IsValid() || !sys.IsValid() {
		return nil, ErrUnset
	}
	if sys.Classes() != t.temp.StepsPerOctave() {
		return nil, classCountMismatch()
	}
	if err := admitPitch(sys, t.ref.note); err != nil {
		return nil, err
	}
	return &Tuning{sys: sys, temp: t.temp, root: t.root, ref: t.ref}, nil
}

// MustWithSystem calls WithSystem and panics with its error on failure.
func (t *Tuning) MustWithSystem(sys *system.System) *Tuning {
	result, err := t.WithSystem(sys)
	if err != nil {
		panic(err)
	}
	return result
}

// WithReference returns a tuning with a different reference and the same
// system, temperament, and root.
func (t *Tuning) WithReference(ref Reference) (*Tuning, error) {
	if !t.IsValid() || !ref.IsValid() {
		return nil, ErrUnset
	}
	if err := admitPitch(t.sys, ref.note); err != nil {
		return nil, err
	}
	return &Tuning{sys: t.sys, temp: t.temp, root: t.root, ref: ref}, nil
}

// MustWithReference calls WithReference and panics with its error on failure.
func (t *Tuning) MustWithReference(ref Reference) *Tuning {
	result, err := t.WithReference(ref)
	if err != nil {
		panic(err)
	}
	return result
}

// WithReferenceAt returns a tuning with a reference supplied as a note-frequency
// pair. The note is checked before the frequency.
func (t *Tuning) WithReferenceAt(note pitch.Pitch, freq float64) (*Tuning, error) {
	if !t.IsValid() || !note.IsValid() {
		return nil, ErrUnset
	}
	if err := admitPitch(t.sys, note); err != nil {
		return nil, err
	}
	if !cents.IsNormalFrequency(freq) {
		return nil, invalidArgument()
	}
	return &Tuning{sys: t.sys, temp: t.temp, root: t.root, ref: Reference{note: note, freq: freq}}, nil
}

// MustWithReferenceAt calls WithReferenceAt and panics with its error on failure.
func (t *Tuning) MustWithReferenceAt(note pitch.Pitch, freq float64) *Tuning {
	result, err := t.WithReferenceAt(note, freq)
	if err != nil {
		panic(err)
	}
	return result
}

// WithReferenceFrequency returns a tuning with the same reference note and a
// different reference frequency. It does not reject an overflowing ratio; use
// FrequencyRatio when the ratio itself is needed.
func (t *Tuning) WithReferenceFrequency(freq float64) (*Tuning, error) {
	if !t.IsValid() {
		return nil, ErrUnset
	}
	ref, err := t.ref.WithFrequency(freq)
	if err != nil {
		return nil, err
	}
	return &Tuning{sys: t.sys, temp: t.temp, root: t.root, ref: ref}, nil
}

// MustWithReferenceFrequency calls WithReferenceFrequency and panics with its
// error on failure.
func (t *Tuning) MustWithReferenceFrequency(freq float64) *Tuning {
	result, err := t.WithReferenceFrequency(freq)
	if err != nil {
		panic(err)
	}
	return result
}

// FrequencyRatio returns freq divided by the tuning's reference frequency.
func (t *Tuning) FrequencyRatio(freq float64) (float64, error) {
	if !t.IsValid() {
		return 0, ErrUnset
	}
	return t.ref.FrequencyRatio(freq)
}

// MustFrequencyRatio calls FrequencyRatio and panics with its error on failure.
func (t *Tuning) MustFrequencyRatio(freq float64) float64 {
	ratio, err := t.FrequencyRatio(freq)
	if err != nil {
		panic(err)
	}
	return ratio
}

// Equal reports whether all four parts of t and other are equal. System and
// temperament names do not affect equality. Two unset tunings are equal; an
// unset and a set tuning are not. Pointer identity is not tuning equality.
func (t *Tuning) Equal(other *Tuning) bool {
	tValid, otherValid := t.IsValid(), other.IsValid()
	if tValid != otherValid {
		return false
	}
	if !tValid {
		return true
	}
	return t.sys.Equal(other.sys) && t.temp.Equal(other.temp) &&
		t.root == other.root && t.ref.Equal(other.ref)
}

// String returns a debug representation with the names and all four parts.
// It returns "Tuning(unset)" for an unset tuning.
func (t *Tuning) String() string {
	if !t.IsValid() {
		return "Tuning(unset)"
	}
	return "Tuning(system " + strconv.Quote(t.sys.Name()) +
		"; temperament " + strconv.Quote(t.temp.Name()) +
		"; root " + t.root.String() + "; reference " + t.ref.String() + ")"
}

func admitPitch(sys *system.System, note pitch.Pitch) error {
	admitted, err := sys.AdmitsPitch(note)
	if err != nil {
		return systemDistanceError(err)
	}
	if !admitted {
		return inadmissible()
	}
	return nil
}
