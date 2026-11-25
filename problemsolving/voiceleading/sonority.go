package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Sonority is an immutable, measured snapshot of simultaneous notes named by
// voice from the bass upward. A standalone sonority keeps the record order of
// its chord and the LowToHigh order; a part sonority of an analysis keeps its
// sounding parts in input-slot order, the silent parts, and the order of the
// analysis. Notes are never sorted.
//
// The zero value and nil pointer are unset; except IsSet and Equal, their
// methods return ErrUnset. A sonority without voices is set. Assignment
// copies share only immutable data; values are not comparable with ==.
// Methods never read sources of settings and are safe for concurrent use.
// Returned slices and notes are independent copies; the system is borrowed.
type Sonority struct {
	set      bool
	notes    []note.Note
	pitches  []pitch.Pitch
	sys      *system.System
	order    Order
	parts    []int
	hasParts bool
	silent   []int
	adjacent []VoiceInterval
	span     VoiceInterval
	// soundErr explains why a single sounding voice of a part sonority has
	// no class by sound: no system or an inadmissible spelling.
	soundErr error
}

// NewSonority measures a chord as voices from the bass upward in the order of
// record: notes are not sorted, repeats are allowed, and the empty chord gives
// a sonority without voices. With at least one note every note is measured:
// a non-nil sys is used without reading sources and must admit every
// spelling; a nil sys is chosen from the notes (every note bound, the
// settings of each note read exactly once in order, equal systems, admissible
// spellings). Adjacent intervals and the span are measured at construction.
// Notes are kept as copies with their corrections and sources.
//
// Errors, in this order: ErrUnset with chord.ErrUnset in a ChordError{0, -1}
// for an unset chord; ErrUnset with system.ErrInvalid for an unset explicit
// system; ErrNoOctave in a ChordError{0, k}; ErrNoSystem, a settings failure,
// ErrDifferentSystems, and system.ErrInadmissible in a ChordError{0, k};
// measurement failures in a ChordError{0, k} with the upper note of the pair.
func NewSonority(notes chord.Chord, sys *system.System) (*Sonority, error) {
	if !notes.IsValid() {
		return nil, chordErr(0, -1, unset("chord", chord.ErrUnset))
	}
	if sys != nil && !sys.IsValid() {
		return nil, unset("system", system.ErrInvalid)
	}
	ns := notes.Notes()
	for k, n := range ns {
		if !n.HasOctave() {
			return nil, chordErr(0, k, ErrNoOctave)
		}
	}
	s := &Sonority{set: true, notes: ns, sys: sys, order: LowToHigh, silent: []int{}, pitches: make([]pitch.Pitch, len(ns))}
	for k, n := range ns {
		s.pitches[k] = n.MustPitch()
	}
	if len(ns) == 0 {
		s.adjacent = []VoiceInterval{}
		return s, nil
	}
	if s.sys == nil {
		chosen, err := chooseChordSystem([][]note.Note{ns}, 0)
		if err != nil {
			return nil, err
		}
		s.sys = chosen
	}
	if err := admitNotes(s.sys, [][]note.Note{ns}, 0); err != nil {
		return nil, err
	}
	if err := s.measure(); err != nil {
		return nil, err
	}
	return s, nil
}

// MustNewSonority is like NewSonority but panics with its error.
func MustNewSonority(notes chord.Chord, sys *system.System) *Sonority {
	return must(NewSonority(notes, sys))
}

// chooseChordSystem chooses a system from all notes of chords numbered from
// first, in (chord, note) order: bindings, then settings, then equality.
func chooseChordSystem(chords [][]note.Note, first int) (*system.System, error) {
	for i, ns := range chords {
		for k, n := range ns {
			if !n.IsBound() {
				return nil, chordErr(first+i, k, ErrNoSystem)
			}
		}
	}
	type read struct {
		i, k int
		sys  *system.System
	}
	var systems []read
	for i, ns := range chords {
		for k, n := range ns {
			st, err := n.Settings()
			if err != nil {
				return nil, chordErr(first+i, k, err)
			}
			systems = append(systems, read{i: first + i, k: k, sys: st.System()})
		}
	}
	if len(systems) == 0 {
		return nil, nil
	}
	for _, r := range systems {
		if !systems[0].sys.Equal(r.sys) {
			return nil, chordErr(r.i, r.k, ErrDifferentSystems)
		}
	}
	return systems[0].sys, nil
}

func admitNotes(sys *system.System, chords [][]note.Note, first int) error {
	for i, ns := range chords {
		for k, n := range ns {
			if err := admit(sys, n.Spelling()); err != nil {
				return chordErr(first+i, k, err)
			}
		}
	}
	return nil
}

// admit reports system.ErrInadmissible, or the failure of the check, for a
// name that sys does not admit.
func admit(sys *system.System, name spelling.Spelling) error {
	ok, err := sys.Admits(name)
	if err != nil {
		return err
	}
	if !ok {
		return system.ErrInadmissible
	}
	return nil
}

func (s *Sonority) measure() error {
	n := len(s.pitches)
	s.adjacent = make([]VoiceInterval, 0, max(0, n-1))
	for k := 1; k < n; k++ {
		vi, err := measurePair(s.sys, s.pitches, k-1, k)
		if err != nil {
			return chordErr(0, k, err)
		}
		s.adjacent = append(s.adjacent, vi)
	}
	if n >= 2 {
		vi, err := measurePair(s.sys, s.pitches, 0, n-1)
		if err != nil {
			return chordErr(0, n-1, err)
		}
		s.span = vi
	}
	return nil
}

func measurePair(sys *system.System, ps []pitch.Pitch, lower, upper int) (VoiceInterval, error) {
	i, err := interval.BetweenPitches(sys, ps[lower], ps[upper])
	if err != nil {
		return VoiceInterval{}, err
	}
	c, err := sys.Distance(ps[lower], ps[upper])
	if err != nil {
		return VoiceInterval{}, err
	}
	return VoiceInterval{Lower: lower, Upper: upper, Interval: i, Chromatic: c}, nil
}

// IsSet reports whether the sonority was created by this package.
func (s *Sonority) IsSet() bool { return s != nil && s.set }

func (s *Sonority) check() error {
	if !s.IsSet() {
		return ErrUnset
	}
	return nil
}

func (s *Sonority) checkOrder() error {
	if err := s.check(); err != nil {
		return err
	}
	if s.order != LowToHigh {
		return ErrNoOrder
	}
	return nil
}

// Equal reports whether both sonorities are unset, or both are set with the
// same order, presence and value of the system, voice notes (by
// note.Note.Equal, sources ignored), presence and values of the parts, and
// silent parts. A standalone sonority never equals a part sonority.
func (s *Sonority) Equal(other *Sonority) bool {
	if !s.IsSet() || !other.IsSet() {
		return s.IsSet() == other.IsSet()
	}
	return s.order == other.order && equalSystem(s.sys, other.sys) &&
		slices.EqualFunc(s.notes, other.notes, note.Note.Equal) &&
		s.hasParts == other.hasParts && slices.Equal(s.parts, other.parts) && slices.Equal(s.silent, other.silent)
}

// Len returns the number of voices.
func (s *Sonority) Len() (int, error) {
	if err := s.check(); err != nil {
		return 0, err
	}
	return len(s.notes), nil
}

// MustLen is like Len but panics with its error.
func (s *Sonority) MustLen() int { return must(s.Len()) }

// Notes returns copies of the voice notes from the bass upward (or in
// input-slot order for a part sonority), in a new slice.
func (s *Sonority) Notes() ([]note.Note, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	return slices.Clone(s.notes), nil
}

// MustNotes is like Notes but panics with its error.
func (s *Sonority) MustNotes() []note.Note { return must(s.Notes()) }

// Parts returns the input slots of the sounding parts in increasing order and
// true for a part sonority, or nil and false for a standalone sonority.
func (s *Sonority) Parts() ([]int, bool, error) {
	if err := s.check(); err != nil {
		return nil, false, err
	}
	if !s.hasParts {
		return nil, false, nil
	}
	return slices.Clone(s.parts), true, nil
}

// MustParts is like Parts but panics with its error.
func (s *Sonority) MustParts() ([]int, bool) { return must2(s.Parts()) }

// Silent returns the input slots of the silent parts in increasing order;
// it is empty for a standalone sonority.
func (s *Sonority) Silent() ([]int, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	return slices.Clone(s.silent), nil
}

// MustSilent is like Silent but panics with its error.
func (s *Sonority) MustSilent() []int { return must(s.Silent()) }

// Order returns LowToHigh for a standalone sonority and the order of the
// analysis for a part sonority.
func (s *Sonority) Order() (Order, error) {
	if err := s.check(); err != nil {
		return Unordered, err
	}
	return s.order, nil
}

// MustOrder is like Order but panics with its error.
func (s *Sonority) MustOrder() Order { return must(s.Order()) }

// System returns the borrowed system and whether it is present.
func (s *Sonority) System() (*system.System, bool, error) {
	if err := s.check(); err != nil {
		return nil, false, err
	}
	return s.sys, s.sys != nil, nil
}

// MustSystem is like System but panics with its error.
func (s *Sonority) MustSystem() (*system.System, bool) { return must2(s.System()) }
