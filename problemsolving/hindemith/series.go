package hindemith

import (
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
	entry    seriesEntry
	root     note.Note
}

// From returns the lower input note of the measurement as given.
func (d Dyad) From() note.Note { return d.from }

// To returns the upper input note of the measurement as given.
func (d Dyad) To() note.Note { return d.to }

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

// distance returns the signed distance in steps from a to b.
func distance(op string, sys *system.System, a, b note.Note) (int, error) {
	pa, _ := a.Pitch()
	pb, _ := b.Pitch()
	s, err := sys.Distance(pa, pb)
	if err != nil {
		return 0, rangeError(op, err)
	}
	return int(s), nil
}

// classify builds the dyad of the pair (a, b) at the signed distance s.
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
	if entry.hasRoot {
		lower, upper := a, b
		if s < 0 {
			lower, upper = b, a
		}
		if entry.rootLow {
			out.root = lower
		} else {
			out.root = upper
		}
	}
	return out
}
