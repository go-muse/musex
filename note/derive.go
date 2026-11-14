package note

import (
	"iter"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
)

// The operations below return new notes and never change the note they are
// called on or the note arguments. All errors are reported before the
// sequence is returned; the sources are read once, at the call. A sequence is
// lazy, can be stopped at any element, and yields the same notes when it is
// iterated again or concurrently; every yielded note is a new value.

func (n Note) enharmonics(op string, c opContext, limit alteration.Alteration) (iter.Seq[Note], error) {
	if err := checkNote(op, n); err != nil {
		return nil, err
	}

	if err := c.check(op); err != nil {
		return nil, err
	}

	r := newReader(op)

	f, err := c.frame(r, n)
	if err != nil {
		return nil, err
	}

	var ownSys *system.System

	if c.in && n.src != nil {
		own, err := r.read(n.src)
		if err != nil {
			return nil, err
		}

		ownSys = own.System()
	}

	if err := admit(op, f.sys, n.name); err != nil {
		return nil, err
	}

	if n.oct.IsValid() {
		ps, err := f.sys.EnharmonicPitches(n.pitch(), limit)
		if err != nil {
			return nil, depError(op, err)
		}

		if ownSys != nil {
			for p := range ps {
				if err := admit(op, ownSys, p.Spelling()); err != nil {
					return nil, err
				}
			}
		}

		return func(yield func(Note) bool) {
			for p := range ps {
				if !yield(n.withPitch(p)) {
					return
				}
			}
		}, nil
	}

	names, err := f.sys.EnharmonicNames(n.name, limit)
	if err != nil {
		return nil, depError(op, err)
	}

	if ownSys != nil {
		for name := range names {
			if err := admit(op, ownSys, name); err != nil {
				return nil, err
			}
		}
	}

	return func(yield func(Note) bool) {
		for name := range names {
			res := n
			res.name = name

			if !yield(res) {
				return
			}
		}
	}, nil
}

// Enharmonics returns the notes at the position of n with the admissible names
// whose alteration is at most limit in magnitude: for a note with an octave
// the pitches of system.System.EnharmonicPitches, for a note without an octave
// the names of system.System.EnharmonicNames, in their order. n itself is
// included when |a| <= limit. The notes have the correction and the source
// of n.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; system.ErrLimit for a negative limit,
// without a sentinel of note.
func (n Note) Enharmonics(limit alteration.Alteration) (iter.Seq[Note], error) {
	return n.enharmonics("Enharmonics", ownContext(), limit)
}

// MustEnharmonics is like Enharmonics but panics with its error.
func (n Note) MustEnharmonics(limit alteration.Alteration) iter.Seq[Note] {
	return must(n.Enharmonics(limit))
}

// EnharmonicsIn is like Enharmonics in the settings of the explicit source
// src. The notes stay bound to the source of n; for a bound note every name
// of the result must also be admissible in its own settings, which is checked
// before the sequence is returned. A nil src reports ErrUnset.
func (n Note) EnharmonicsIn(src settings.Source, limit alteration.Alteration) (iter.Seq[Note], error) {
	return n.enharmonics("EnharmonicsIn", sourceContext(src), limit)
}

// MustEnharmonicsIn is like EnharmonicsIn but panics with its error.
func (n Note) MustEnharmonicsIn(src settings.Source, limit alteration.Alteration) iter.Seq[Note] {
	return must(n.EnharmonicsIn(src, limit))
}

// Range returns the notes with the name, the correction and the source of n
// in each Octave from from to to, both included, going up or down as
// pitch.Range. The source is not read; the sequence does not allocate memory
// for the whole range.
//
// Errors: ErrUnset for the note, an unset from or an unset to.
func (n Note) Range(from, to octave.Octave) (iter.Seq[Note], error) {
	const op = "Range"

	if err := checkNote(op, n); err != nil {
		return nil, err
	}

	if !from.IsValid() {
		return nil, unsetError(op, "from")
	}

	if !to.IsValid() {
		return nil, unsetError(op, "to")
	}

	octaves, err := octave.Range(from, to)
	if err != nil {
		return nil, depError(op, err)
	}

	return func(yield func(Note) bool) {
		for o := range octaves {
			res := n
			res.oct = o

			if !yield(res) {
				return
			}
		}
	}, nil
}

// MustRange is like Range but panics with its error.
func (n Note) MustRange(from, to octave.Octave) iter.Seq[Note] {
	return must(n.Range(from, to))
}

func (n Note) walk(op string, c opContext, to Note, rule system.SpellingRule) (iter.Seq[Note], error) {
	notes := [2]Note{n, to}

	if err := checkNotes(op, notes[:]); err != nil {
		return nil, err
	}

	if err := c.check(op); err != nil {
		return nil, err
	}

	if !rule.IsValid() {
		return nil, unsetError(op, "spelling rule")
	}

	if err := checkOctaves(op, notes[:]); err != nil {
		return nil, err
	}

	r := newReader(op)

	// In an In form, the unset settings of the explicit source, then of the
	// own source of n, are reported before any name is checked; the reader
	// then gives positionSystem the explicit settings without a second read.
	var ownSys *system.System

	if c.in {
		if _, err := c.frame(r, n); err != nil {
			return nil, err
		}

		if n.src != nil {
			own, err := r.read(n.src)
			if err != nil {
				return nil, err
			}

			ownSys = own.System()
		}
	}

	sys, err := positionSystem(op, r, c, notes[:])
	if err != nil {
		return nil, err
	}

	from, last := n.pitch(), to.pitch()

	ps, err := sys.Walk(from, last, rule)
	if err != nil {
		return nil, depError(op, err)
	}

	if ownSys != nil {
		if err := checkWalk(op, ownSys, sys, ps, from, last); err != nil {
			return nil, err
		}
	}

	return func(yield func(Note) bool) {
		for p := range ps {
			if !yield(n.withPitch(p)) {
				return
			}
		}
	}, nil
}

// checkWalk checks the names of the first min(length, N) pitches of a walk in
// the system own; the names repeat after N pitches, N being the number of
// classes of sys.
func checkWalk(op string, own, sys *system.System, ps iter.Seq[pitch.Pitch], from, to pitch.Pitch) error {
	limit := sys.Classes()

	if d, err := sys.Distance(from, to); err == nil {
		if d < 0 {
			d = -d
		}

		if d >= 0 && int(d) < limit {
			limit = int(d) + 1
		}
	}

	checked := 0

	for p := range ps {
		if checked == limit {
			break
		}

		if err := admit(op, own, p.Spelling()); err != nil {
			return err
		}

		checked++
	}

	return nil
}

// Walk returns the notes at the positions from the position of n to the
// position of to, both included, by one step up or down, with names chosen by
// rule (system.System.Walk). The notes have the correction and the source of
// n. Each note is in its own settings (read once for one source); the systems
// must be equal.
//
// Errors, in this order: ErrUnset for the notes or an unset rule; ErrNoOctave;
// ErrNoSource; ErrUnsetSettings; ErrInadmissible; ErrDifferentSystems; the
// errors of the rule (system.ErrNotApplicable, system.ErrInadmissible for its
// tonic, system.ErrNoSpelling, system.ErrBadSpelling) without a sentinel of
// note.
func (n Note) Walk(to Note, rule system.SpellingRule) (iter.Seq[Note], error) {
	return n.walk("Walk", ownContext(), to, rule)
}

// MustWalk is like Walk but panics with its error.
func (n Note) MustWalk(to Note, rule system.SpellingRule) iter.Seq[Note] {
	return must(n.Walk(to, rule))
}

// WalkIn is like Walk in the settings of the explicit source src for both
// notes. The notes stay bound to the source of n; for a bound n the names of
// the result must also be admissible in its own settings, which is checked
// before the sequence is returned. A nil src reports ErrUnset.
func (n Note) WalkIn(src settings.Source, to Note, rule system.SpellingRule) (iter.Seq[Note], error) {
	return n.walk("WalkIn", sourceContext(src), to, rule)
}

// MustWalkIn is like WalkIn but panics with its error.
func (n Note) MustWalkIn(src settings.Source, to Note, rule system.SpellingRule) iter.Seq[Note] {
	return must(n.WalkIn(src, to, rule))
}
