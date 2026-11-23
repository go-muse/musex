package hindemith

import (
	"math"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/system"
)

// seriesEntry is one Series II interval: its position, its name and the
// side of its root. rootLow false with hasRoot true means the upper tone.
type seriesEntry struct {
	name    string
	series  int
	rootLow bool
	hasRoot bool
}

// seriesByClass maps the tempered distance modulo 12 to the Series II
// entry; class 0 is the unison, whole octaves are the octave (position 2).
var seriesByClass = [12]seriesEntry{
	{"perfect unison", 1, true, true},
	{"minor second", 11, false, true},
	{"major second", 9, false, true},
	{"minor third", 7, true, true},
	{"major third", 5, true, true},
	{"perfect fourth", 4, false, true},
	{"tritone", 13, false, false},
	{"perfect fifth", 3, true, true},
	{"minor sixth", 6, false, true},
	{"major sixth", 8, false, true},
	{"minor seventh", 10, true, true},
	{"major seventh", 12, true, true},
}

var octaveEntry = seriesEntry{"octave", 2, true, true}

// Dyad is one analyzed interval between two notes: its Series II position
// and its root, when the interval has one. The zero Dyad is unset.
type Dyad struct {
	from, to note.Note
	low, up  note.Note
	entry    seriesEntry
	root     note.Note
}

// From returns the first argument of the measurement, which is the lower
// note only when the interval ascends.
func (d Dyad) From() note.Note { return d.from }

// To returns the second argument of the measurement, which is the upper
// note only when the interval ascends.
func (d Dyad) To() note.Note { return d.to }

// Equal reports whether d and other are set and measure the same ordered
// pair of notes (compared with note.Note.Equal) to the same Series II
// position with the same root. The root is compared only when the interval
// has one: two tritone dyads, which have no root, are equal when their
// ordered notes are equal. An unset dyad equals no dyad, not even another
// unset dyad, as for note.Note.Equal.
func (d Dyad) Equal(other Dyad) bool {
	if !d.IsValid() || !other.IsValid() {
		return false
	}

	return d.from.Equal(other.from) && d.to.Equal(other.to) &&
		d.entry.series == other.entry.series &&
		(!d.entry.hasRoot || d.root.Equal(other.root))
}

// IsValid reports whether the dyad was analyzed.
func (d Dyad) IsValid() bool { return d.entry.series != 0 }

// Series returns the Series II position, 1 to 13, or 0 for an unset dyad.
func (d Dyad) Series() int { return d.entry.series }

// Name returns the tempered interval name, such as "perfect fifth" or
// "tritone", or "" for an unset dyad.
func (d Dyad) Name() string { return d.entry.name }

// IsTritone reports whether the interval is the tritone, which has no root.
func (d Dyad) IsTritone() bool { return d.entry.series == 13 }

// HasRoot reports whether the interval has a root; the tritone has none.
func (d Dyad) HasRoot() bool { return d.entry.hasRoot }

// Root returns the root note, or the unset note for the tritone and for an
// unset dyad.
func (d Dyad) Root() note.Note { return d.root }

// checkSystem validates the explicit system of the model.
func checkSystem(op string, sys *system.System) error {
	if !sys.IsValid() {
		return opError(op, ErrUnset)
	}
	if sys.Classes() != 12 || sys.Chromatic() != 1 {
		return opError(op, ErrUnsupportedSystem)
	}
	return nil
}

// checkNotes validates the notes of a measurement.
func checkNotes(op string, ns []note.Note) error {
	for _, n := range ns {
		if !n.IsValid() {
			return opError(op, ErrUnset)
		}
		if !n.HasOctave() {
			return opError(op, ErrNoOctave)
		}
	}
	return nil
}

// checkAdmitted rejects note names that sys does not admit.
func checkAdmitted(op string, sys *system.System, ns []note.Note) error {
	for _, n := range ns {
		ok, err := sys.Admits(n.Spelling())
		if err != nil {
			return opError(op, ErrUnset)
		}
		if !ok {
			return inadmissibleError(op, system.ErrInadmissible)
		}
	}
	return nil
}

// distance returns the signed distance in steps from a to b. A distance of
// exactly math.MinInt steps is out of range: its magnitude does not fit in
// int, so every returned distance can be negated safely.
func distance(op string, sys *system.System, a, b note.Note) (int, error) {
	pa, _ := a.Pitch()
	pb, _ := b.Pitch()
	s, err := sys.Distance(pa, pb)
	if err != nil {
		return 0, rangeError(op, err)
	}
	if int(s) == math.MinInt {
		return 0, rangeError(op, system.ErrOutOfRange)
	}
	return int(s), nil
}

// classify builds the dyad of the pair (a, b) at the signed distance s,
// which distance has already checked to be greater than math.MinInt.
func classify(a, b note.Note, s int) Dyad {
	d := s
	if d < 0 {
		d = -d
	}
	entry := seriesByClass[d%12]
	if d%12 == 0 && d > 0 {
		entry = octaveEntry
	}
	out := Dyad{from: a, to: b, entry: entry}
	out.low, out.up = a, b
	if s < 0 {
		out.low, out.up = b, a
	}
	if entry.hasRoot {
		if entry.rootLow {
			out.root = out.low
		} else {
			out.root = out.up
		}
	}
	return out
}
