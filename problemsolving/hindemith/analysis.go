package hindemith

import (
	"sort"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/system"
)

// indeterminateTriad reports whether the distinct pitch classes are three
// tones a major third or a perfect fourth apart, in any rotation: the
// augmented triad and the perfect quartal triad of group V.
func indeterminateTriad(pcs map[int]bool) bool {
	if len(pcs) != 3 {
		return false
	}
	for base := range pcs {
		rels := make([]int, 0, 3)
		for x := range pcs {
			rels = append(rels, ((x-base)%12+12)%12)
		}
		sort.Ints(rels)
		if (rels[0] == 0 && rels[1] == 4 && rels[2] == 8) ||
			(rels[0] == 0 && rels[1] == 5 && rels[2] == 10) {
			return true
		}
	}
	return false
}

// AnalyzeDyad measures the interval from the note from to the note to in
// sys and returns its Series II position and root. The tritone has no root.
// Both notes must have octaves; cents do not take part.
//
// Errors, in this order: ErrUnset for a nil or unset system or an unset
// note; ErrUnsupportedSystem for a system that is not twelve classes with
// chromatic unit 1; ErrNoOctave; ErrInadmissible together with
// system.ErrInadmissible for a name the system does not admit;
// ErrOutOfRange together with system.ErrOutOfRange when the distance does
// not fit.
func AnalyzeDyad(sys *system.System, from, to note.Note) (Dyad, error) {
	const op = "AnalyzeDyad"

	if err := checkSystem(op, sys); err != nil {
		return Dyad{}, err
	}
	if err := checkNotes(op, []note.Note{from, to}); err != nil {
		return Dyad{}, err
	}
	if err := checkAdmitted(op, sys, []note.Note{from, to}); err != nil {
		return Dyad{}, err
	}

	s, err := distance(op, sys, from, to)
	if err != nil {
		return Dyad{}, err
	}

	return classify(from, to, s), nil
}

// MustAnalyzeDyad is like AnalyzeDyad but panics with its error.
func MustAnalyzeDyad(sys *system.System, from, to note.Note) Dyad {
	return must(AnalyzeDyad(sys, from, to))
}

// Group is one of Hindemith's six chord groups. The zero value is unset.
type Group uint8

// The chord groups I to VI of the table of chord groups.
const (
	GroupI Group = iota + 1
	GroupII
	GroupIII
	GroupIV
	GroupV
	GroupVI
)

// String returns the roman numeral, or "Group(unset)".
func (g Group) String() string {
	switch g {
	case GroupI:
		return "I"
	case GroupII:
		return "II"
	case GroupIII:
		return "III"
	case GroupIV:
		return "IV"
	case GroupV:
		return "V"
	case GroupVI:
		return "VI"
	default:
		return "Group(unset)"
	}
}

// Subgroup is the subdivision of a group: 1 and 2 for root in the bass and
// root above the bass, a and b1, b2, b3 for group II. SubNone marks the
// undivided groups V and VI and the zero value.
type Subgroup uint8

// The subgroups of the table of chord groups.
const (
	SubNone Subgroup = iota
	Sub1
	Sub2
	SubA
	SubB1
	SubB2
	SubB3
)

// String returns "1", "2", "a", "b1", "b2", "b3", or "" for SubNone.
func (s Subgroup) String() string {
	switch s {
	case Sub1:
		return "1"
	case Sub2:
		return "2"
	case SubA:
		return "a"
	case SubB1:
		return "b1"
	case SubB2:
		return "b2"
	case SubB3:
		return "b3"
	default:
		return ""
	}
}

// Chord is one analyzed voiced chord: its group and subgroup, its bass, its
// root when determinate, and the intra-chord dyads that justify the result.
// The zero Chord is unset.
type Chord struct {
	notes   []note.Note
	dyads   []Dyad
	best    Dyad
	group   Group
	sub     Subgroup
	bass    note.Note
	root    note.Note
	hasRoot bool
}

// IsValid reports whether the chord was analyzed.
func (c Chord) IsValid() bool { return c.group != 0 }

// Notes returns a copy of the input notes in input order.
func (c Chord) Notes() []note.Note { return append([]note.Note(nil), c.notes...) }

// Dyads returns a copy of the intra-chord dyads, one per unordered pair of
// notes, in input order of the pairs.
func (c Chord) Dyads() []Dyad { return append([]Dyad(nil), c.dyads...) }

// Best returns the dyad whose root is the chord root, or the unset dyad for
// groups V and VI and for an unset chord.
func (c Chord) Best() Dyad { return c.best }

// Group returns the chord group, or the unset group.
func (c Chord) Group() Group { return c.group }

// Subgroup returns the subgroup, or SubNone for groups V and VI.
func (c Chord) Subgroup() Subgroup { return c.sub }

// Bass returns the lowest note; equal positions keep the first input note.
func (c Chord) Bass() note.Note { return c.bass }

// HasRoot reports whether the chord has a determinate root; groups V and VI
// have none in isolation.
func (c Chord) HasRoot() bool { return c.hasRoot }

// Root returns the chord root, or the unset note when there is none.
func (c Chord) Root() note.Note { return c.root }

// Equal reports whether c and other classify the same notes in the same
// order to the same group, subgroup, bass and root. Two unset chords are
// not equal, as for note.Note.Equal.
func (c Chord) Equal(other Chord) bool {
	if len(c.notes) != len(other.notes) || c.group != other.group || c.sub != other.sub ||
		c.hasRoot != other.hasRoot || !c.bass.Equal(other.bass) || !c.root.Equal(other.root) {
		return false
	}
	for i := range c.notes {
		if !c.notes[i].Equal(other.notes[i]) {
			return false
		}
	}
	return true
}

// Value returns the position of the class in the fluctuation order I.1,
// I.2, II.a, II.b1, II.b2, II.b3, III.1, III.2, IV.1, IV.2, V, VI: 1 to 12,
// or 0 for an unset chord. Higher values mean higher harmonic tension.
func (c Chord) Value() int {
	if !c.IsValid() {
		return 0
	}
	switch c.group {
	case GroupI:
		return int(c.sub) // Sub1, Sub2
	case GroupII:
		switch c.sub {
		case SubA:
			return 3
		case SubB1:
			return 4
		case SubB2:
			return 5
		default:
			return 6
		}
	case GroupIII:
		return 6 + int(c.sub)
	case GroupIV:
		return 8 + int(c.sub)
	case GroupV:
		return 11
	default:
		return 12
	}
}

// AnalyzeChord classifies the voiced chord notes in sys: group, subgroup,
// bass, root and the intra-chord evidence. At least two notes with octaves
// are required; cents do not take part.
//
// Errors, in this order: ErrUnset for a nil or unset system;
// ErrUnsupportedSystem; ErrTooFewNotes for fewer than two notes; ErrUnset
// for an unset note; ErrNoOctave; ErrInadmissible together with
// system.ErrInadmissible for a name the system does not admit;
// ErrOutOfRange together with system.ErrOutOfRange.
func AnalyzeChord(sys *system.System, notes []note.Note) (Chord, error) {
	const op = "AnalyzeChord"

	if err := checkSystem(op, sys); err != nil {
		return Chord{}, err
	}
	if len(notes) < 2 {
		return Chord{}, opError(op, ErrTooFewNotes)
	}
	if err := checkNotes(op, notes); err != nil {
		return Chord{}, err
	}
	if err := checkAdmitted(op, sys, notes); err != nil {
		return Chord{}, err
	}

	out := Chord{notes: append([]note.Note(nil), notes...)}

	bass := notes[0]
	for _, n := range notes[1:] {
		pb, _ := bass.Pitch()
		pn, _ := n.Pitch()
		r, err := sys.Compare(pb, pn)
		if err != nil {
			return Chord{}, rangeError(op, err)
		}
		if r > 0 {
			bass = n
		}
	}
	out.bass = bass

	// Pitch classes relative to the first note identify doublings and
	// distinct tritones independently of the listing order.
	rel := make([]int, len(notes))
	pcs := make(map[int]bool, len(notes))
	for i, n := range notes {
		s, err := distance(op, sys, notes[0], n)
		if err != nil {
			return Chord{}, err
		}
		rel[i] = ((s % 12) + 12) % 12
		pcs[rel[i]] = true
	}

	var seconds, sevenths, minorSeconds, majorSevenths, majorSeconds bool
	onlyTritoneThirdsSixths := true
	tritonePairs := make(map[[2]int]bool)

	for i := range notes {
		for j := i + 1; j < len(notes); j++ {
			s, err := distance(op, sys, notes[i], notes[j])
			if err != nil {
				return Chord{}, err
			}
			d := classify(notes[i], notes[j], s)
			out.dyads = append(out.dyads, d)

			switch d.entry.series {
			case 11:
				minorSeconds = true
				seconds = true
			case 9:
				majorSeconds = true
				seconds = true
			case 10:
				sevenths = true
			case 12:
				majorSevenths = true
				sevenths = true
			}

			semis := s
			if semis < 0 {
				semis = -semis
			}
			semis %= 12
			if semis == 6 {
				a, b := rel[i], rel[j]
				if a > b {
					a, b = b, a
				}
				tritonePairs[[2]int{a, b}] = true
			}
			switch semis {
			case 3, 6, 9:
			default:
				onlyTritoneThirdsSixths = false
			}
		}
	}

	hasTritone := len(tritonePairs) > 0
	switch {
	case !hasTritone && indeterminateTriad(pcs):
		out.group, out.sub = GroupV, SubNone
	case !hasTritone && !seconds && !sevenths:
		out.group = GroupI
	case !hasTritone:
		out.group = GroupIII
	case onlyTritoneThirdsSixths:
		out.group, out.sub = GroupVI, SubNone
	case !minorSeconds && !majorSevenths:
		out.group = GroupII
	default:
		out.group = GroupIV
	}

	if out.group == GroupV || out.group == GroupVI {
		return out, nil
	}

	best := Dyad{}
	for _, d := range out.dyads {
		if d.entry.series <= 2 {
			continue // unison and octave pairs are doublings: no root information
		}
		if !best.IsValid() {
			best = d
			continue
		}
		if d.entry.series != best.entry.series {
			if d.entry.series < best.entry.series {
				best = d
			}
			continue
		}
		// Equal best intervals: the lower one decides, then the upper.
		pl, _ := d.low.Pitch()
		bl, _ := best.low.Pitch()
		r, err := sys.Compare(pl, bl)
		if err != nil {
			return Chord{}, rangeError(op, err)
		}
		if r > 0 {
			continue
		}
		if r == 0 {
			pu, _ := d.up.Pitch()
			bu, _ := best.up.Pitch()
			r, err = sys.Compare(pu, bu)
			if err != nil {
				return Chord{}, rangeError(op, err)
			}
			if r >= 0 {
				continue
			}
		}
		best = d
	}
	if !best.IsValid() {
		// Only doubling pairs: the doubled note is its own root.
		out.best, out.root, out.hasRoot = out.dyads[0], bass, true
	} else {
		out.best, out.root, out.hasRoot = best, best.root, true
	}

	pr, _ := out.root.Pitch()
	pb, _ := bass.Pitch()
	cmp, err := sys.Compare(pr, pb)
	if err != nil {
		return Chord{}, rangeError(op, err)
	}
	rootInBass := cmp == 0
	switch out.group {
	case GroupII:
		switch {
		case len(tritonePairs) > 1:
			out.sub = SubB3
		case !majorSeconds && rootInBass:
			out.sub = SubA
		case rootInBass:
			out.sub = SubB1
		default:
			out.sub = SubB2
		}
	default:
		if rootInBass {
			out.sub = Sub1
		} else {
			out.sub = Sub2
		}
	}

	return out, nil
}

// MustAnalyzeChord is like AnalyzeChord but panics with its error.
func MustAnalyzeChord(sys *system.System, notes []note.Note) Chord {
	return must(AnalyzeChord(sys, notes))
}

// Direction is the change of harmonic tension of one progression step.
type Direction uint8

// The directions of harmonic fluctuation.
const (
	Level Direction = iota
	Rise
	Fall
)

// String returns "level", "rise" or "fall".
func (d Direction) String() string {
	switch d {
	case Rise:
		return "rise"
	case Fall:
		return "fall"
	default:
		return "level"
	}
}

// Step is one adjacent pair of a progression: the fluctuation values of
// the two chords and the direction of the change.
type Step struct {
	from, to int
}

// From returns the fluctuation value of the earlier chord.
func (s Step) From() int { return s.from }

// To returns the fluctuation value of the later chord.
func (s Step) To() int { return s.to }

// Delta returns To minus From: positive when tension rises.
func (s Step) Delta() int { return s.to - s.from }

// Direction returns Rise, Fall or Level.
func (s Step) Direction() Direction {
	switch {
	case s.to > s.from:
		return Rise
	case s.to < s.from:
		return Fall
	default:
		return Level
	}
}

// Progression is an analyzed sequence of chords with its harmonic
// fluctuation. The zero value is an empty progression.
type Progression struct {
	chords []Chord
	steps  []Step
}

// Chords returns a copy of the chord classifications in input order.
func (p Progression) Chords() []Chord { return append([]Chord(nil), p.chords...) }

// Steps returns a copy of the fluctuation steps between adjacent chords.
func (p Progression) Steps() []Step { return append([]Step(nil), p.steps...) }

// AnalyzeProgression classifies every chord and the fluctuation between
// adjacent chords. An empty sequence gives an empty progression.
//
// Errors: those of AnalyzeChord for the first chord that fails.
func AnalyzeProgression(sys *system.System, chords [][]note.Note) (Progression, error) {
	const op = "AnalyzeProgression"

	if err := checkSystem(op, sys); err != nil {
		return Progression{}, err
	}

	out := Progression{
		chords: make([]Chord, 0, len(chords)),
		steps:  make([]Step, 0, len(chords)),
	}
	for _, ns := range chords {
		c, err := AnalyzeChord(sys, ns)
		if err != nil {
			return Progression{}, err
		}
		if len(out.chords) > 0 {
			prev := out.chords[len(out.chords)-1]
			out.steps = append(out.steps, Step{from: prev.Value(), to: c.Value()})
		}
		out.chords = append(out.chords, c)
	}

	return out, nil
}

// MustAnalyzeProgression is like AnalyzeProgression but panics with its
// error.
func MustAnalyzeProgression(sys *system.System, chords [][]note.Note) Progression {
	return must(AnalyzeProgression(sys, chords))
}
