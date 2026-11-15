// Package chord provides the chord: a combination of notes sounding together,
// without a duration.
//
// A chord holds zero or more notes of package note in the order of record.
// It has no duration and no place in time: the duration, ties, arpeggios and
// fermatas belong to an event, which holds a note or a chord. It has no
// harmony either: the intervals of a chord, its quality, root, inversion,
// function and name are left to a separate package. A chord is not a set of
// pitch classes; package note gives the class of each note, and package pcset
// makes sets of them.
//
// # Unset and empty chords
//
// The zero Chord is the unset chord: var c chord.Chord makes a value that does
// not accept notes. Its reading methods give zero results; At and every
// changing method report ErrUnset, and the chord stays unset. A working chord
// is made by New: New() without notes is the empty chord, which accepts notes,
// and New(c4, e4, g4) is a chord with three notes:
//
//	empty := chord.MustNew()
//	empty.MustAdd(c4) // [C4]
//
//	var u chord.Chord
//	err := u.Add(c4) // errors.Is(err, chord.ErrUnset)
//
// No operation turns a set chord into the unset one: removing the last note
// gives the empty chord.
//
// # The order of record and repeats
//
// A chord keeps its notes in the order in which they were given, and gives
// them back in that order. Only Insert and the sorting methods (Sort,
// SortFunc, Reverse) change it; removing and replacing notes keep the order of
// the others. A note can enter a chord several times: [C4, E4, C4] holds C4
// twice. Find gives the first occurrence of a note, Count its multiplicity and
// Indices the numbers of all its occurrences. DistinctNotes, HasRepeats,
// AddMissing and RemoveRepeats work with a chord as with a set.
//
// The notes are compared by record (note.Note.Equal): the record and the
// correction take part, the source does not. C4 and C4 +5 cents are different
// notes, and so are C4 and C without an octave; two C4 notes bound to
// different settings are one note entering twice. A chord can hold any notes:
// with and without an octave, with any corrections, bound to any sources or
// unbound.
//
// # Equality and order
//
// There are three equalities. Equal compares the contents: the same notes
// with the same multiplicities, in any order. EqualInOrder also needs the
// notes with equal numbers to be equal. SameRecords compares the contents
// without the corrections. In all three, the source of a note does not take
// part, and the unset chord equals no chord:
//
//	a := chord.MustNew(c4, e4, g4)
//	b := chord.MustNew(g4, c4, e4)
//	a.Equal(b)        // true
//	a.EqualInOrder(b) // false
//
// Includes and HasCommonNotes compare contents too, counting multiplicity.
// Compare is an agreed order of chords for sorting, consistent with Equal: it
// compares the notes of SortedNotes one by one. It is not "higher" or "lower".
//
// == does not compile for chords, and a chord cannot be a key of a map; a
// struct with a Chord field is not comparable either. Compare chords with the
// methods.
//
// # The notes of a chord are copies
//
// Every note, slice and sequence that a chord gives is new: At, Find, Notes,
// All and Values give copies, and changing them does not change the chord.
// A note inside a chord is changed only by an operation of the chord: Replace,
// ReplaceAll, ReplaceAt, ReplaceFunc, ReplaceAtFunc, Shift, ShiftDown, Up and
// Down. The lists given to New, Add, Insert and AddMissing are copied and not
// kept. All and Values are snapshots: changing the chord while iterating does
// not change the sequence.
//
// # Changing a chord in place and chains
//
// A changing method changes the chord itself; it has a pointer receiver. Its
// variant with an error returns only the error, and on an error the chord does
// not change at all. Its Must variant returns the same chord, not a copy, so
// changes are chained with Must variants:
//
//	c := chord.MustNew(c4, e4, g4)
//	c.MustAdd(e4).MustRemove(c4) // c is [E4, G4, E4]
//
// One operation is atomic, a chain is not: when a step fails, its Must variant
// panics, the earlier steps stay done and the later ones are not done. The
// result of c.MustAdd(n) is c itself: p := c.MustAdd(n) changes c, and p
// points to it.
//
// A copy of a chord is made by assignment. It is cheap, without copying the
// notes, and independent at once, both ways; only the sources of the notes
// are shared:
//
//	d := c        // an independent copy
//	d.MustClear() // c does not change
//
// A chord can be changed only where it is addressable: a variable, an element
// of a slice, a field of an addressable struct, or through a *Chord:
//
//	chords[i].MustAdd(g4)                     // changes the element
//	for _, x := range chords { x.MustAdd(g4) } // changes copies only
//	v := m[k]; v.MustAdd(g4); m[k] = v        // an element of a map is not addressable
//	c.Notes()[0].MustShift(1)                 // changes only the returned copy
//
// The result of a function is not addressable either, so a chain cannot start
// right at New or MustNew; use a variable, or new:
//
//	p := new(chord.MustNew(c4)).MustAdd(e4) // p is *chord.Chord
//
// # Rules of the user
//
// Some operations take a rule of the user: a Condition selects notes
// (ContainsFunc, CountFunc, IndicesFunc, NotesFunc, RemoveFunc, KeepFunc), a
// Transform replaces them (ReplaceFunc, ReplaceAtFunc), and a Comparison
// orders them (SortFunc). A literal function is passed as it is:
//
//	c.MustRemoveFunc(func(n note.Note) (bool, error) { return n.HasOctave(), nil })
//
// A Condition or a Transform gets copies of the notes in the order of record,
// at most once for each note. Any rule may fail: its first error stops the
// operation, which reports ErrRuleFailed together with the error of the rule,
// and a changing operation leaves the chord as it was. A Transform with
// note.Note.Raise thus raises every note of a chord, or none. A panic of a
// rule is not recovered: it passes to the caller, and the chord stays as it
// was. A rule must not change the chord on which the operation runs, even
// through another pointer to it; the result of such an operation is not
// defined.
//
// # Settings belong to the notes
//
// A chord never reads the sources of settings of its notes: the operations of
// a chord work by record. The position, the sound and the frequency are
// computed by package note on the notes of a chord:
//
//	c.MustSortFunc(func(a, b note.Note) (int, error) { return a.ComparePosition(b) })
//	for n := range c.Values() {
//		f, err := n.Frequency()
//		// ...
//	}
//
// A rule may read the sources through package note; that is reading by the
// rule, not by the chord.
//
// # Concurrency
//
// The reading methods, iteration over All and Values included, can be called
// concurrently on one chord. A changing method needs exclusive access to the
// chord: nobody else reads or changes it meanwhile. This is the duty of the
// caller; a chord has no internal synchronization. Different copies can be
// changed and read concurrently, also right after copying. A slice from Notes
// and a sequence from All or Values can be read while the chord changes, as
// long as the call that made them did not run during a change. Methods with a
// rule are as safe as the rule. The package has no global mutable state,
// caches or goroutines.
//
// # Errors
//
// Errors are tested with errors.Is against the sentinels ErrUnset,
// ErrNotFound, ErrIndex and ErrRuleFailed; the documentation of the errors
// describes the order in which the causes are reported. Shift, ShiftDown, Up
// and Down report the error of package note for the first note that cannot
// be moved, such as note.ErrNoOctave for a note without an octave, without a
// sentinel of chord. Must variants panic with the same error.
//
// # Cost
//
// A copy, IsValid, IsEmpty, Len and At take constant time; searching a note
// takes time linear in the number of notes, and the comparisons of contents
// sort the notes. Every change builds a new array of notes, so adding n notes
// one by one costs O(n²) copying; add many notes with one call of New, Add or
// Insert.
package chord
