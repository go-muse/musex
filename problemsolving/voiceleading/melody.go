package voiceleading

import (
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"
)

// Move is the melodic interval between two consecutive sounds of one voice.
// Interval is the full written interval from From to To with its written
// direction; Chromatic is the signed distance in positions of the system.
// ThroughRest reports that at least one rest lies between the two sounds:
// such a move is listed descriptively but takes part in no norm or run.
type Move struct {
	From, To    ElementAt
	Interval    interval.Interval
	Chromatic   system.Step
	ThroughRest bool
}

// IsRepeat reports whether the move is a perfect unison: no direction, neither
// a step nor a leap. It is false for an unset Interval.
func (m Move) IsRepeat() bool { return m.Interval.IsPerfectUnison() }

// IsStep reports whether the move is a second of any quality, including
// augmented and diminished seconds.
func (m Move) IsStep() bool { return m.Interval.Size() == 2 }

// IsLeap reports whether the move is a leap under def: a third or wider for
// LeapFromThird, a fourth or wider for LeapFromFourth. Any other def, and an
// unset Interval, give false.
func (m Move) IsLeap(def LeapDefinition) bool { return isLeap(m.Interval, def) }

func isLeap(i interval.Interval, def LeapDefinition) bool {
	switch def {
	case LeapFromThird:
		return i.Size() >= 3
	case LeapFromFourth:
		return i.Size() >= 4
	default:
		return false
	}
}

// LeapRun identifies a maximal run of consecutive leaps by the indices of its
// first and last move in Melody.Moves.
type LeapRun struct{ First, Last int }

// Melody is an immutable, measured snapshot of one voice: all elements of a
// line and the moves between its consecutive sounds. The zero value and nil
// pointer are unset; except IsSet and Equal, their methods return ErrUnset.
// Assignment copies share only immutable data; values are not comparable with
// ==. Methods never read sources of settings and are safe for concurrent use.
// Returned slices are independent; the system is borrowed and immutable.
// A melody of a part may include elements after all pairwise intersections,
// which Analysis.Equal does not compare; compare melodies with Melody.Equal.
type Melody struct {
	set      bool
	part     int
	elements []ElementAt
	sys      *system.System
	moves    []Move
}

// NewMelody measures one line. Sounds are measured only when there are at
// least two of them; then every sound is measured. A non-nil sys is used
// without reading sources, and each sound's spelling must be admissible in it.
// A nil sys is chosen from the sounds of the line: every sound must be bound,
// then the settings of each sound are read exactly once in order, the systems
// must be equal, and the spellings admissible. With fewer than two sounds no
// system is chosen and no source is read.
//
// Errors, in this order: ErrUnset with voice.ErrUnset in a PartError{0, -1}
// for an unset line; ErrUnset with system.ErrInvalid for an unset explicit
// system; ErrNoSystem, a settings failure, ErrDifferentSystems, and
// system.ErrInadmissible in a PartError{0, element}; measurement failures
// of the moves in a PartError{0, element of To}.
func NewMelody(line *voice.Line, sys *system.System) (*Melody, error) {
	if !line.IsSet() {
		return nil, &PartError{Part: 0, Element: -1, Err: unset("line", voice.ErrUnset)}
	}
	if sys != nil && !sys.IsValid() {
		return nil, unset("system", system.ErrInvalid)
	}
	elems := line.MustElements()
	m := &Melody{set: true, sys: sys, elements: make([]ElementAt, len(elems))}
	for j, e := range elems {
		m.elements[j] = ElementAt{Index: j, Element: e, Start: line.MustStartOf(j), End: line.MustEndOf(j)}
	}
	sounds := soundIndices(m.elements)
	if len(sounds) < 2 {
		m.moves = []Move{}
		return m, nil
	}
	if m.sys == nil {
		chosen, err := chooseLineSystem(m.elements, sounds)
		if err != nil {
			return nil, err
		}
		m.sys = chosen
	}
	if err := m.admit(sounds); err != nil {
		return nil, err
	}
	if err := m.measure(sounds); err != nil {
		return nil, err
	}
	return m, nil
}

// MustNewMelody is like NewMelody but panics with its error.
func MustNewMelody(line *voice.Line, sys *system.System) *Melody {
	return must(NewMelody(line, sys))
}

func soundIndices(elems []ElementAt) []int {
	out := []int{}
	for i, e := range elems {
		if e.Element.IsSound() {
			out = append(out, i)
		}
	}
	return out
}

func chooseLineSystem(elems []ElementAt, sounds []int) (*system.System, error) {
	for _, i := range sounds {
		if n, _ := elems[i].Element.MustNote(); !n.IsBound() {
			return nil, &PartError{Part: 0, Element: i, Err: ErrNoSystem}
		}
	}
	systems := make([]*system.System, len(sounds))
	for k, i := range sounds {
		n, _ := elems[i].Element.MustNote()
		s, err := n.Settings()
		if err != nil {
			return nil, &PartError{Part: 0, Element: i, Err: err}
		}
		systems[k] = s.System()
	}
	for k, i := range sounds {
		if !systems[0].Equal(systems[k]) {
			return nil, &PartError{Part: 0, Element: i, Err: ErrDifferentSystems}
		}
	}
	return systems[0], nil
}

func (m *Melody) admit(sounds []int) error {
	for _, i := range sounds {
		n, _ := m.elements[i].Element.MustNote()
		if err := admit(m.sys, n.Spelling()); err != nil {
			return &PartError{Part: m.part, Element: i, Err: err}
		}
	}
	return nil
}

func (m *Melody) measure(sounds []int) error {
	m.moves = make([]Move, 0, len(sounds)-1)
	for k := 1; k < len(sounds); k++ {
		from, to := m.elements[sounds[k-1]], m.elements[sounds[k]]
		x, y := elementPitch(from), elementPitch(to)
		i, err := interval.BetweenPitches(m.sys, x, y)
		if err != nil {
			return &PartError{Part: m.part, Element: to.Index, Err: err}
		}
		c, err := m.sys.Distance(x, y)
		if err != nil {
			return &PartError{Part: m.part, Element: to.Index, Err: err}
		}
		m.moves = append(m.moves, Move{From: from, To: to, Interval: i, Chromatic: c, ThroughRest: to.Index > from.Index+1})
	}
	return nil
}

// Melody returns the melody of one part from the snapshot, in the system of
// the analysis and without reading sources. A sound outside every sounding
// pairwise vertical is measured in that system; only its spelling is checked.
// Analysis.Equal does not see the elements after all pairwise intersections:
// two equal analyses can give different melodies, which Melody.Equal compares.
// For a check of each note's system by its source use NewMelody(line, nil).
//
// Errors, in this order: ErrUnset; ErrPart for an out-of-range part; with at
// least two sounds, ErrNoSystem for the first sound of the part when the
// analysis has no system, system.ErrInadmissible for a sound, and measurement
// failures of the moves, each in a PartError{part, element}.
func (a *Analysis) Melody(part int) (*Melody, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	if part < 0 || part >= len(a.names) {
		return nil, ErrPart
	}
	m := &Melody{set: true, part: part, sys: a.sys, elements: slices.Clone(a.parts[part])}
	sounds := soundIndices(m.elements)
	if len(sounds) < 2 {
		m.moves = []Move{}
		return m, nil
	}
	if m.sys == nil {
		return nil, &PartError{Part: part, Element: sounds[0], Err: ErrNoSystem}
	}
	if err := m.admit(sounds); err != nil {
		return nil, err
	}
	if err := m.measure(sounds); err != nil {
		return nil, err
	}
	return m, nil
}

// MustMelody is like Melody but panics with its error.
func (a *Analysis) MustMelody(part int) *Melody { return must(a.Melody(part)) }

// IsSet reports whether the melody was created by NewMelody or Analysis.Melody.
func (m *Melody) IsSet() bool { return m != nil && m.set }

func (m *Melody) check() error {
	if !m.IsSet() {
		return ErrUnset
	}
	return nil
}

// Equal reports whether both melodies are unset, or both are set with equal
// elements, the same presence and value of the system, and equal moves.
func (m *Melody) Equal(other *Melody) bool {
	if !m.IsSet() || !other.IsSet() {
		return m.IsSet() == other.IsSet()
	}
	return slices.EqualFunc(m.elements, other.elements, equalElement) && equalSystem(m.sys, other.sys) &&
		slices.EqualFunc(m.moves, other.moves, equalMove)
}

// Elements returns all elements of the voice in order, in a new slice.
func (m *Melody) Elements() ([]ElementAt, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	return slices.Clone(m.elements), nil
}

// MustElements is like Elements but panics with its error.
func (m *Melody) MustElements() []ElementAt { return must(m.Elements()) }

// System returns the borrowed system and whether it is present: an explicit
// system always, a chosen one after measurement, the system of the analysis
// for a part's melody, and nil, false otherwise.
func (m *Melody) System() (*system.System, bool, error) {
	if err := m.check(); err != nil {
		return nil, false, err
	}
	return m.sys, m.sys != nil, nil
}

// MustSystem is like System but panics with its error.
func (m *Melody) MustSystem() (*system.System, bool) { return must2(m.System()) }

// Moves returns the moves between consecutive sounds in order, in a new slice.
// A line with fewer than two sounds gives an empty slice.
func (m *Melody) Moves() ([]Move, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	return slices.Clone(m.moves), nil
}

// MustMoves is like Moves but panics with its error.
func (m *Melody) MustMoves() []Move { return must(m.Moves()) }

// LeapRuns returns the maximal runs of consecutive leaps under def within a
// fragment (a rest breaks a run) with at least minLength moves, in order.
// Nested parts of a run are not reported. This is a description, not a norm.
// Errors, in this order: ErrUnset; ErrOption for def; ErrOption for
// minLength < 1.
func (m *Melody) LeapRuns(def LeapDefinition, minLength int) ([]LeapRun, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	if !def.valid() || minLength < 1 {
		return nil, ErrOption
	}
	out := []LeapRun{}
	start := -1
	flush := func(end int) {
		if start >= 0 && end-start+1 >= minLength {
			out = append(out, LeapRun{First: start, Last: end})
		}
		start = -1
	}
	for k, mv := range m.moves {
		if mv.ThroughRest || !mv.IsLeap(def) {
			flush(k - 1)
			continue
		}
		if start < 0 {
			start = k
		}
	}
	flush(len(m.moves) - 1)
	return out, nil
}

// MustLeapRuns is like LeapRuns but panics with its error.
func (m *Melody) MustLeapRuns(def LeapDefinition, minLength int) []LeapRun {
	return must(m.LeapRuns(def, minLength))
}
