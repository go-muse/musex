// Package note provides the note: a written record (a name, or a written
// pitch), a correction in cents and an optional reference to a source of
// settings.
//
// # Bound and unbound notes
//
// A note bound to a source of settings (settings.Source) reads the source at
// the moment of each operation that needs settings and works with the
// snapshot it gets. The note does not store the settings, the frequency, the
// step or the text, and it does not compute them itself: the frequency comes
// from package tuning, the position from package system, the text from
// package settings. The note adds only its correction to the written pitch.
// A note knows only where to read its settings from; it knows nothing of an
// instrument, which is one possible owner and operator of a source.
//
// An unbound note has no source. Operations that do not need settings work
// for it as for a bound note; operations that need them report ErrNoSource
// unless an explicit context is given. Standard settings are never
// substituted silently: pass settings.Standard() as a source when you need
// them.
//
// # The frequency of a note
//
// Frequency returns the frequency of the written pitch of the note in the
// tuning of its settings, shifted by its correction with one rounding:
//
//	std := settings.Standard()
//	a4 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), std)
//	f := a4.MustFrequency() // 440
//
// Each call reads the source again: when the owner of the source retunes it,
// for example to A4 at 415 Hz, the next call of Frequency of the same note
// returns the new frequency without recreating the note. A note without an
// octave has no frequency (ErrNoOctave).
//
// # Changing the source
//
// SetSource binds the note to another source keeping its Octave; with another
// system its written octave can change. SetSourceWritten keeps the written
// octave and recomputes the Octave in the new system. Both refuse a name
// inadmissible in the new system (ErrInadmissible) and a change that would
// make an expressible note inexpressible (ErrInexpressible).
// CheckSetSource and CheckSetSourceWritten report all the violated conditions
// without changing the note. Unbind removes the source.
//
// # Two ways to change the alteration
//
// The short methods keep the Octave, as package pitch does: Raise of B3 gives
// B♯ in octave 3, written "B♯2". The methods with Written keep the written
// octave, as package system does: RaiseWritten of B3 gives B♯ in octave 4,
// written "B♯3". The short methods work without settings; the Written ones
// need the system of the settings.
//
// # Changing a note in place and chains
//
// A changing method changes the note itself; it has a pointer receiver. Its
// variant with an error returns only the error, and on an error the note does
// not change at all. Its Must variant returns the same note, not a copy, so
// changes are chained with Must variants:
//
//	n := note.MustFromPitch(c4, cents.MustFromInt(30), std)
//	n.MustRaise().MustUp().MustSetCents(cents.MustFromInt(-10)) // n is C♯5 −10 cents
//
// A chain is not atomic: when a step fails, its Must variant panics and the
// earlier steps stay done.
//
// A copy of a note is made by assignment; the copy and the original are
// independent, and only the reference to the source is shared:
//
//	third := root               // an independent copy
//	third.MustMove(4, rule)     // root does not change
//
// This differs from package pitch, whose values are immutable: there one
// writes p = p.MustRaise(), here n.MustRaise(). The result of n.MustRaise() is
// n itself: hi := n.MustRaise() changes n, and hi points to it.
//
// A note can be changed only where it is addressable: a variable, an element
// of a slice, a field of an addressable struct, or through a *Note:
//
//	notes[i].MustLower()                    // changes the element
//	for _, x := range notes { x.MustLower() } // changes copies only
//	v := m[k]; v.MustLower(); m[k] = v      // an element of a map is not addressable
//
// The result of a function is not addressable either, so a chain cannot start
// right at a creating function; use a variable, or new:
//
//	p := new(note.MustFromPitch(c4, cents.Zero(), std)).MustRaise() // p is *note.Note
//
// # Equality
//
// Equal compares the records and the corrections, without the sources. ==
// compiles for notes but is not the equality of records: it also compares the
// sources, so two notes bound to two separate standard settings differ by ==,
// and it panics when two sources have one incomparable dynamic type. The
// same holds for notes as keys of a map. Use Equal.
//
// # Concurrency
//
// The reading methods and the methods that return new notes never change the
// note; they can be called concurrently on one note as long as concurrent
// reading of its source is safe, which the contract of settings.Source
// requires. A changing method needs exclusive access to the note: nobody else
// reads or changes it meanwhile. This is the duty of the caller; a note has no
// internal synchronization. Different copies can be changed concurrently. The
// package has no global mutable state, caches or goroutines.
//
// # Explicit context
//
// The operations that need settings have forms with an explicit source (the
// In forms, such as FrequencyIn) and, where only the tuning is needed, with an
// explicit tuning (the InTuning forms, such as FrequencyInTuning). The binding
// of the note does not change. A bound note changed with an explicit source,
// and the notes made from it, stay bound to its own source and must also be
// admissible in its own settings.
//
// # Levels of comparison
//
// Without settings, any notes can be compared: Equal, SameRecord, the Same
// methods, Compare (an agreed order of records), OctaveSub. By position
// (Equivalent, ComparePosition, Between, Distance, SortedByPosition) the
// systems of the notes must be equal (ErrDifferentSystems); the correction
// does not take part. By sound (CompareSound, SoundsSame, CentsTo,
// SortedBySound, Offset) each note is in its own settings, which may differ.
// With equal tunings a strict comparison is exact: it compares the heights
// h + c in cents, not the rounded frequencies. Otherwise frequencies are
// compared, strictly or by a criterion of package frequency.
//
// # Errors
//
// Errors are tested with errors.Is against the sentinels of the package and
// of its dependencies; the documentation of the errors of the package
// describes the order in which the causes are reported. Must variants panic
// with the same error.
// Format can panic inside package notation for octaves extremely far from the
// middle; use IsExpressible to test such notes.
package note
