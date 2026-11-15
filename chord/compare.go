package chord

import (
	"github.com/go-muse/musex/note"
)

// Equal reports whether c and other are set and have equal contents: every
// note has one and the same multiplicity in both, the notes compared by
// note.Note.Equal. The order of record and the sources of the notes do not
// take part. Two empty chords are equal; the unset chord equals no chord, not
// even the unset one.
//
// Equal is reflexive for set chords, symmetric and transitive.
// EqualInOrder implies Equal, and Equal implies SameRecords.
func (c Chord) Equal(other Chord) bool {
	return c.sameContents(other, note.Note.Equal)
}

// EqualInOrder reports whether c and other are set, have the same number of
// notes, and their notes with equal numbers are equal (note.Note.Equal).
func (c Chord) EqualInOrder(other Chord) bool {
	if !c.set || !other.set || len(c.notes) != len(other.notes) {
		return false
	}

	for i, n := range c.notes {
		if !n.Equal(other.notes[i]) {
			return false
		}
	}

	return true
}

// SameRecords reports whether c and other are set and have equal contents
// when the notes are compared by note.Note.SameRecord, regardless of the
// corrections: [C4, C4 +5 cents] and [C4, C4] have the same records.
func (c Chord) SameRecords(other Chord) bool {
	return c.sameContents(other, note.Note.SameRecord)
}

// Includes reports whether c and other are set and c contains all the notes
// of other, counting multiplicity. A set chord includes the empty chord and
// itself; c and other include each other only when they are Equal.
func (c Chord) Includes(other Chord) bool {
	if !c.set || !other.set {
		return false
	}

	_, inC, inOther := multiplicities(c.notes, other.notes, note.Note.Equal)
	for id, need := range inOther {
		if need > inC[id] {
			return false
		}
	}

	return true
}

// HasCommonNotes reports whether c and other are set and some note enters
// both.
func (c Chord) HasCommonNotes(other Chord) bool {
	if !c.set || !other.set {
		return false
	}

	_, inC, inOther := multiplicities(c.notes, other.notes, note.Note.Equal)
	for id, given := range inOther {
		if given > 0 && inC[id] > 0 {
			return true
		}
	}

	return false
}

// Compare returns -1, 0 or +1 by an agreed order of chords: the unset chord
// before every set one, two unset chords compare as 0. Set chords are
// compared by their SortedNotes, note by note with note.Note.Compare: the
// first difference decides; when one list is the beginning of the other, the
// shorter comes first. For set chords Compare returns 0 exactly when they are
// Equal.
//
// Compare is a total order, fit for slices.SortFunc(chords,
// chord.Chord.Compare). It is an order for sorting, not "higher" or "lower".
func (c Chord) Compare(other Chord) int {
	switch {
	case !c.set && !other.set:
		return 0
	case !c.set:
		return -1
	case !other.set:
		return 1
	}

	a, b := c.SortedNotes(), other.SortedNotes()
	for i := range min(len(a), len(b)) {
		if r := a[i].Compare(b[i]); r != 0 {
			return r
		}
	}

	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	default:
		return 0
	}
}

// sameContents reports whether c and other are set and every class of notes
// equal by same has one multiplicity in both.
func (c Chord) sameContents(other Chord, same func(x, y note.Note) bool) bool {
	if !c.set || !other.set || len(c.notes) != len(other.notes) {
		return false
	}

	_, inC, inOther := multiplicities(c.notes, other.notes, same)
	for id, given := range inOther {
		if given != inC[id] {
			return false
		}
	}

	return true
}
