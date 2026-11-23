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
// The classification belongs to the chord reduced to the lowest note of
// every pitch class (see AnalyzeChord); Notes and Dyads keep every input
// note and pair, doublings included. The zero Chord is unset.
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

// Notes returns a copy of all input notes in input order, doublings
// included.
func (c Chord) Notes() []note.Note { return append([]note.Note(nil), c.notes...) }

// Dyads returns a copy of the intra-chord dyads, one per unordered pair of
// input notes (n(n-1)/2 for n notes, doubling pairs included), in input
// order of the pairs.
func (c Chord) Dyads() []Dyad { return append([]Dyad(nil), c.dyads...) }

// Best returns the dyad whose root is the chord root: the strongest pair of
// the chord reduced to the lowest note of every pitch class. For a chord of
// one doubled tone it is the first input pair of the bass. Best returns the
// unset dyad for groups V and VI and for an unset chord.
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

// Root returns the chord root, the lowest occurrence of its pitch class, or
// the unset note when there is none.
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
// Doublings never change the classification. The chord is first reduced to
// the lowest note of every pitch class, where a pitch class is the position
// in sys modulo 12 and not the spelling: G♭4 and F♯5, or C4 and B♯4, are
// one class. Of two notes of one class at the same position, such as C♯4
// and D♭4, the first in input order stays, as for Chord.Bass. The group,
// subgroup, tritone count, best interval, root, root-versus-bass comparison
// and Chord.Value follow from the reduced chord alone: a higher repetition
// of a tone adds no interval, while a doubling below the lowest occurrence
// of a tone becomes that tone in the reduced chord. A chord of one doubled
// tone (unisons and octaves only) is group I, subgroup 1, with the bass as
// its root. Chord.Notes and Chord.Dyads still report every input note and
// every input pair.
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

	// Signed position of every note relative to the first one: it orders
	// the notes by pitch and gives the pitch class independently of the
	// spelling.
	pos := make([]int, len(notes))
	for i, n := range notes {
		s, err := distance(op, sys, notes[0], n)
		if err != nil {
			return Chord{}, err
		}
		pos[i] = s
	}

	// The evidence covers every input pair, doublings included.
	dist := make([][]int, len(notes))
	for i := range dist {
		dist[i] = make([]int, len(notes))
	}
	out.dyads = make([]Dyad, 0, len(notes)*(len(notes)-1)/2)
	for i := range notes {
		for j := i + 1; j < len(notes); j++ {
			s, err := distance(op, sys, notes[i], notes[j])
			if err != nil {
				return Chord{}, err
			}
			dist[i][j], dist[j][i] = s, -s
			out.dyads = append(out.dyads, classify(notes[i], notes[j], s))
		}
	}

	red := lowestPerClass(pos)

	// The bass is the lowest note of the reduced chord, which is also the
	// lowest input note, the first one among equal positions.
	bassIdx := red[0]
	for _, i := range red[1:] {
		if pos[i] < pos[bassIdx] {
			bassIdx = i
		}
	}
	out.bass = notes[bassIdx]

	pcs := make(map[int]bool, len(red))
	for _, i := range red {
		pcs[pitchClass(pos[i])] = true
	}

	// The interval content of the reduced chord: one pair per two different
	// pitch classes.
	type pair struct {
		dyad            Dyad
		low, up, rootAt int
	}
	pairs := make([]pair, 0, len(red)*(len(red)-1)/2)
	var seconds, sevenths, minorSeconds, majorSevenths, majorSeconds bool
	onlyTritoneThirdsSixths := true
	tritones := 0

	for x := range red {
		for y := x + 1; y < len(red); y++ {
			i, j := red[x], red[y]
			s := dist[i][j]
			p := pair{dyad: classify(notes[i], notes[j], s), low: i, up: j}
			if s < 0 {
				p.low, p.up = j, i
			}
			p.rootAt = p.up
			if p.dyad.entry.rootLow {
				p.rootAt = p.low
			}
			pairs = append(pairs, p)

			switch p.dyad.entry.series {
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

			switch pitchClass(s) {
			case 6:
				tritones++
			case 3, 9:
			default:
				onlyTritoneThirdsSixths = false
			}
		}
	}

	hasTritone := tritones > 0
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

	rootIdx := bassIdx
	if len(pairs) == 0 {
		// One doubled tone: it is its own root, and the evidence is the
		// first input pair of the bass, whose root is the bass. The pairs of
		// the first note come first, so that pair is (0, 1) when the first
		// note is the bass and (0, bassIdx) otherwise; in the latter case the
		// bass is strictly lower than the first note and is the root.
		out.best = out.dyads[max(bassIdx-1, 0)]
	} else {
		// The strongest interval decides; among equal best intervals the
		// lower pair, then the upper. Positions in the reduced chord are
		// all different, so the choice does not depend on the input order.
		best := pairs[0]
		for _, p := range pairs[1:] {
			switch {
			case p.dyad.entry.series != best.dyad.entry.series:
				if p.dyad.entry.series < best.dyad.entry.series {
					best = p
				}
			case pos[p.low] != pos[best.low]:
				if pos[p.low] < pos[best.low] {
					best = p
				}
			case pos[p.up] < pos[best.up]:
				best = p
			}
		}
		out.best = best.dyad
		rootIdx = best.rootAt
	}
	out.root, out.hasRoot = notes[rootIdx], true

	rootInBass := rootIdx == bassIdx
	switch out.group {
	case GroupII:
		switch {
		case tritones > 1:
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

// pitchClass reduces a signed distance in steps to its class 0 to 11.
func pitchClass(s int) int { return ((s % 12) + 12) % 12 }

// lowestPerClass reduces a chord given by the signed positions of its notes
// to one note per pitch class: the lowest one, and of equal positions the
// first in input order. It returns the input indices of the kept notes in
// the order in which their classes first occur.
func lowestPerClass(pos []int) []int {
	slot := make(map[int]int, len(pos))
	out := make([]int, 0, len(pos))
	for i, p := range pos {
		c := pitchClass(p)
		k, ok := slot[c]
		if !ok {
			slot[c] = len(out)
			out = append(out, i)
			continue
		}
		if p < pos[out[k]] {
			out[k] = i
		}
	}
	return out
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
