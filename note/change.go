package note

import (
	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// The changing methods change the note itself and have a pointer receiver.
// Each one computes the whole result first and then assigns it to the note:
// on an error the note stays unchanged whole (record, correction, source).
// The Must variant of a changing method returns the receiver itself, so
// changes can be chained: n.MustRaise().MustUp(). A chain is not atomic: a
// failed step panics and leaves the earlier steps done.
//
// A nil *Note and the unset note report ErrUnset; the Must variants panic
// with that error.

// mustDo returns n or panics with err.
func (n *Note) mustDo(err error) *Note {
	if err != nil {
		panic(err)
	}

	return n
}

// current returns the note behind n, or ErrUnset for a nil pointer or the
// unset note.
func (n *Note) current(op string) (Note, error) {
	if n == nil {
		return Note{}, unsetError(op, "nil *Note")
	}

	if err := checkNote(op, *n); err != nil {
		return Note{}, err
	}

	return *n, nil
}

// noCheck is a check of the arguments of a changing method that never fails.
func noCheck() error { return nil }

// changeKeepingOctave changes the record of n keeping its Octave (or keeping
// no octave): args checks the arguments (ErrUnset), next computes the new
// note. A bound note reads its source once and checks the name of the result
// in the system of its settings.
func (n *Note) changeKeepingOctave(op string, args func() error, next func(cur Note) (Note, error)) error {
	cur, err := n.current(op)
	if err != nil {
		return err
	}

	if err := args(); err != nil {
		return err
	}

	var sys *system.System

	if cur.src != nil {
		set, err := newReader(op).read(cur.src)
		if err != nil {
			return err
		}

		sys = set.System()
	}

	res, err := next(cur)
	if err != nil {
		return err
	}

	if sys != nil {
		if err := admit(op, sys, res.name); err != nil {
			return err
		}
	}

	*n = res

	return nil
}

// changeWritten changes the name of n keeping its written octave: the
// Octave of the result is w + k of the new name, w being the written octave
// of n (system.System.RespellWritten). A note without an octave keeps having
// no octave. With explicit set (an In form) the computation is in the
// settings of explicit, and the name of the result must also be admissible in
// the own settings of a bound note.
func (n *Note) changeWritten(op string, explicit settings.Source, inForm bool, args func() error,
	name func(cur spelling.Spelling) (spelling.Spelling, error),
) error {
	cur, err := n.current(op)
	if err != nil {
		return err
	}

	if inForm {
		if err := checkSource(op, explicit); err != nil {
			return err
		}
	}

	if err := args(); err != nil {
		return err
	}

	r := newReader(op)

	var (
		sys    *system.System // the system of the computation; nil when none
		ownSys *system.System // the own system of an In form
	)

	switch {
	case inForm:
		f, err := sourceFrame(r, explicit)
		if err != nil {
			return err
		}

		sys = f.sys

		if cur.src != nil {
			own, err := r.read(cur.src)
			if err != nil {
				return err
			}

			ownSys = own.System()
		}
	case cur.oct.IsValid():
		f, err := ownFrame(r, cur)
		if err != nil {
			return err
		}

		sys = f.sys
	case cur.src != nil:
		own, err := r.read(cur.src)
		if err != nil {
			return err
		}

		sys = own.System()
	}

	withOctave := cur.oct.IsValid()

	if withOctave {
		if err := admit(op, sys, cur.name); err != nil {
			return err
		}
	}

	nn, err := name(cur.name)
	if err != nil {
		return depError(op, err)
	}

	for _, s := range [...]*system.System{sys, ownSys} {
		if s == nil {
			continue
		}

		if err := admit(op, s, nn); err != nil {
			return err
		}
	}

	res := cur
	res.name = nn

	if withOctave {
		p, err := sys.RespellWritten(cur.pitch(), nn)
		if err != nil {
			return depError(op, err)
		}

		res = res.withPitch(p)
	}

	*n = res

	return nil
}

// setWritten gives n the written octave returned by written: Octave = w + k
// of the name of n (system.System.FromWritten). A note without an octave gets
// one. With explicit set (an In form) the computation is in the settings of
// explicit, and the name must also be admissible in the own settings of a
// bound note.
func (n *Note) setWritten(op string, explicit settings.Source, inForm bool, args func() error,
	written func(set *settings.Settings) (octave.Octave, error),
) error {
	cur, err := n.current(op)
	if err != nil {
		return err
	}

	if inForm {
		if err := checkSource(op, explicit); err != nil {
			return err
		}
	}

	if err := args(); err != nil {
		return err
	}

	r := newReader(op)

	var (
		f      frame
		ownSys *system.System
	)

	if inForm {
		f, err = sourceFrame(r, explicit)
		if err != nil {
			return err
		}

		if cur.src != nil {
			own, err := r.read(cur.src)
			if err != nil {
				return err
			}

			ownSys = own.System()
		}
	} else {
		f, err = ownFrame(r, cur)
		if err != nil {
			return err
		}
	}

	for _, s := range [...]*system.System{f.sys, ownSys} {
		if s == nil {
			continue
		}

		if err := admit(op, s, cur.name); err != nil {
			return err
		}
	}

	w, err := written(f.set)
	if err != nil {
		return depError(op, err)
	}

	p, err := f.sys.FromWritten(cur.name, w)
	if err != nil {
		return depError(op, err)
	}

	*n = cur.withPitch(p)

	return nil
}

// Respell replaces the name of n by name, keeping the Octave; a note without
// an octave keeps having no octave. A bound note reads its source and checks
// that name is admissible in the system of its settings.
//
// Errors, in this order: ErrUnset for the note or an unset name;
// ErrUnsetSettings; ErrInadmissible.
func (n *Note) Respell(name spelling.Spelling) error {
	const op = "Respell"

	return n.changeKeepingOctave(op, nameArg(op, name), func(cur Note) (Note, error) {
		cur.name = name

		return cur, nil
	})
}

// MustRespell is like Respell but panics with its error; it returns n.
func (n *Note) MustRespell(name spelling.Spelling) *Note {
	return n.mustDo(n.Respell(name))
}

// nameArg returns a check that reports ErrUnset for an unset name.
func nameArg(op string, name spelling.Spelling) func() error {
	return func() error {
		if !name.IsValid() {
			return unsetError(op, "name")
		}

		return nil
	}
}

// RespellWritten replaces the name of n by name, keeping the written octave:
// the Octave of the result is w + k of name in the system of the settings of
// n (system.System.RespellWritten). A note without an octave changes as with
// Respell.
//
// Errors, in this order: ErrUnset for the note or an unset name; ErrNoSource
// for an unbound note with an octave; ErrUnsetSettings; ErrInadmissible for
// the name of n, then for name; ErrOutOfRange (also system.ErrOutOfRange).
func (n *Note) RespellWritten(name spelling.Spelling) error {
	const op = "RespellWritten"

	return n.changeWritten(op, nil, false, nameArg(op, name), func(spelling.Spelling) (spelling.Spelling, error) {
		return name, nil
	})
}

// MustRespellWritten is like RespellWritten but panics with its error; it
// returns n.
func (n *Note) MustRespellWritten(name spelling.Spelling) *Note {
	return n.mustDo(n.RespellWritten(name))
}

// RespellWrittenIn is like RespellWritten in the settings of the explicit
// source src; the binding of n does not change. The name of the result must be
// admissible in the settings of src and, for a bound note, in its own
// settings. A nil src reports ErrUnset.
func (n *Note) RespellWrittenIn(src settings.Source, name spelling.Spelling) error {
	const op = "RespellWrittenIn"

	return n.changeWritten(op, src, true, nameArg(op, name), func(spelling.Spelling) (spelling.Spelling, error) {
		return name, nil
	})
}

// MustRespellWrittenIn is like RespellWrittenIn but panics with its error; it
// returns n.
func (n *Note) MustRespellWrittenIn(src settings.Source, name spelling.Spelling) *Note {
	return n.mustDo(n.RespellWrittenIn(src, name))
}

// SetOctave sets the Octave of n to o; a note without an octave gets one.
// The source is not read.
//
// Errors: ErrUnset for the note or an unset o.
func (n *Note) SetOctave(o octave.Octave) error {
	const op = "SetOctave"

	cur, err := n.current(op)
	if err != nil {
		return err
	}

	if !o.IsValid() {
		return unsetError(op, "octave")
	}

	cur.oct = o
	*n = cur

	return nil
}

// MustSetOctave is like SetOctave but panics with its error; it returns n.
func (n *Note) MustSetOctave(o octave.Octave) *Note {
	return n.mustDo(n.SetOctave(o))
}

// SetASPN sets the Octave of n to the Octave with ASPN number aspn.
//
// Errors: ErrUnset for the note.
func (n *Note) SetASPN(aspn int) error {
	const op = "SetASPN"

	cur, err := n.current(op)
	if err != nil {
		return err
	}

	cur.oct = octave.FromASPN(aspn)
	*n = cur

	return nil
}

// MustSetASPN is like SetASPN but panics with its error; it returns n.
func (n *Note) MustSetASPN(aspn int) *Note {
	return n.mustDo(n.SetASPN(aspn))
}

// SetWrittenOctave sets the written octave of n to w: the Octave becomes
// w + k, the carry of the name in the system of the settings of n
// (system.System.FromWritten). A note without an octave gets one.
//
// Errors, in this order: ErrUnset for the note or an unset w; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange (also
// system.ErrOutOfRange).
func (n *Note) SetWrittenOctave(w octave.Octave) error {
	const op = "SetWrittenOctave"

	return n.setWritten(op, nil, false, octaveArg(op, w), fixedOctave(w))
}

// MustSetWrittenOctave is like SetWrittenOctave but panics with its error; it
// returns n.
func (n *Note) MustSetWrittenOctave(w octave.Octave) *Note {
	return n.mustDo(n.SetWrittenOctave(w))
}

// octaveArg returns a check that reports ErrUnset for an unset octave.
func octaveArg(op string, o octave.Octave) func() error {
	return func() error {
		if !o.IsValid() {
			return unsetError(op, "octave")
		}

		return nil
	}
}

// fixedOctave returns a written octave that does not depend on the settings.
func fixedOctave(w octave.Octave) func(*settings.Settings) (octave.Octave, error) {
	return func(*settings.Settings) (octave.Octave, error) { return w, nil }
}

// numberedOctave returns the written octave with the visible number num in
// the numbering of the settings.
func numberedOctave(num int) func(*settings.Settings) (octave.Octave, error) {
	return func(set *settings.Settings) (octave.Octave, error) {
		return octave.FromNumber(num, set.Numbering())
	}
}

// SetWrittenASPN is like SetWrittenOctave with the written octave given by its
// ASPN number.
func (n *Note) SetWrittenASPN(aspn int) error {
	return n.setWritten("SetWrittenASPN", nil, false, noCheck, fixedOctave(octave.FromASPN(aspn)))
}

// MustSetWrittenASPN is like SetWrittenASPN but panics with its error; it
// returns n.
func (n *Note) MustSetWrittenASPN(aspn int) *Note {
	return n.mustDo(n.SetWrittenASPN(aspn))
}

// SetWrittenNumber is like SetWrittenOctave with the written octave given by
// its visible number num in the numbering of the settings of n. A number
// skipped in the numbering reports number.ErrSkipped without a sentinel of
// note; a number out of range reports ErrOutOfRange.
func (n *Note) SetWrittenNumber(num int) error {
	return n.setWritten("SetWrittenNumber", nil, false, noCheck, numberedOctave(num))
}

// MustSetWrittenNumber is like SetWrittenNumber but panics with its error; it
// returns n.
func (n *Note) MustSetWrittenNumber(num int) *Note {
	return n.mustDo(n.SetWrittenNumber(num))
}

// SetWrittenOctaveIn is like SetWrittenOctave in the settings of the explicit
// source src; the binding of n does not change. The name of a bound note must
// also be admissible in its own settings. A nil src reports ErrUnset.
func (n *Note) SetWrittenOctaveIn(src settings.Source, w octave.Octave) error {
	const op = "SetWrittenOctaveIn"

	return n.setWritten(op, src, true, octaveArg(op, w), fixedOctave(w))
}

// MustSetWrittenOctaveIn is like SetWrittenOctaveIn but panics with its error;
// it returns n.
func (n *Note) MustSetWrittenOctaveIn(src settings.Source, w octave.Octave) *Note {
	return n.mustDo(n.SetWrittenOctaveIn(src, w))
}

// SetWrittenNumberIn is like SetWrittenNumber in the settings of the explicit
// source src, with num in the numbering of src.
func (n *Note) SetWrittenNumberIn(src settings.Source, num int) error {
	return n.setWritten("SetWrittenNumberIn", src, true, noCheck, numberedOctave(num))
}

// MustSetWrittenNumberIn is like SetWrittenNumberIn but panics with its error;
// it returns n.
func (n *Note) MustSetWrittenNumberIn(src settings.Source, num int) *Note {
	return n.mustDo(n.SetWrittenNumberIn(src, num))
}

// RemoveOctave removes the octave of n, keeping its name and correction;
// a note without an octave does not change.
//
// Errors: ErrUnset for the note.
func (n *Note) RemoveOctave() error {
	cur, err := n.current("RemoveOctave")
	if err != nil {
		return err
	}

	cur.oct = octave.Octave{}
	*n = cur

	return nil
}

// MustRemoveOctave is like RemoveOctave but panics with its error; it returns
// n.
func (n *Note) MustRemoveOctave() *Note {
	return n.mustDo(n.RemoveOctave())
}

// ReplaceWithPitch replaces the whole record of n by the written pitch p,
// keeping the correction and the source.
//
// Errors, in this order: ErrUnset for the note or an unset p;
// ErrUnsetSettings; ErrInadmissible.
func (n *Note) ReplaceWithPitch(p pitch.Pitch) error {
	const op = "ReplaceWithPitch"

	args := func() error {
		if !p.IsValid() {
			return unsetError(op, "pitch")
		}

		return nil
	}

	return n.changeKeepingOctave(op, args, func(cur Note) (Note, error) {
		return cur.withPitch(p), nil
	})
}

// MustReplaceWithPitch is like ReplaceWithPitch but panics with its error; it
// returns n.
func (n *Note) MustReplaceWithPitch(p pitch.Pitch) *Note {
	return n.mustDo(n.ReplaceWithPitch(p))
}

// ReplaceWithSpelling replaces the whole record of n by name: the note
// becomes a note without an octave, with the same correction and source.
// Respell, by contrast, keeps the octave.
//
// Errors, in this order: ErrUnset for the note or an unset name;
// ErrUnsetSettings; ErrInadmissible.
func (n *Note) ReplaceWithSpelling(name spelling.Spelling) error {
	const op = "ReplaceWithSpelling"

	return n.changeKeepingOctave(op, nameArg(op, name), func(cur Note) (Note, error) {
		cur.name = name
		cur.oct = octave.Octave{}

		return cur, nil
	})
}

// MustReplaceWithSpelling is like ReplaceWithSpelling but panics with its
// error; it returns n.
func (n *Note) MustReplaceWithSpelling(name spelling.Spelling) *Note {
	return n.mustDo(n.ReplaceWithSpelling(name))
}

// SetCents replaces the correction of n by c.
//
// Errors: ErrUnset for the note.
func (n *Note) SetCents(c cents.Cents) error {
	cur, err := n.current("SetCents")
	if err != nil {
		return err
	}

	cur.c = c
	*n = cur

	return nil
}

// MustSetCents is like SetCents but panics with its error; it returns n.
func (n *Note) MustSetCents(c cents.Cents) *Note {
	return n.mustDo(n.SetCents(c))
}

// alter changes the alteration of n keeping its Octave.
func (n *Note) alter(op string, f func(spelling.Spelling) (spelling.Spelling, error)) error {
	return n.changeKeepingOctave(op, noCheck, func(cur Note) (Note, error) {
		name, err := f(cur.name)
		if err != nil {
			return Note{}, depError(op, err)
		}

		cur.name = name

		return cur, nil
	})
}

// Raise raises the alteration of n by 1, keeping the Octave, as
// pitch.Pitch.Raise: B3 becomes B♯ in octave 3, written "B♯2". Use
// RaiseWritten to keep the written octave instead. A bound note checks that
// the new name is admissible in the system of its settings.
//
// Errors, in this order: ErrUnset for the note; ErrUnsetSettings;
// ErrOutOfRange for an alteration out of range (also
// alteration.ErrOutOfRange); ErrInadmissible for the new name.
func (n *Note) Raise() error {
	return n.alter("Raise", spelling.Spelling.Raise)
}

// MustRaise is like Raise but panics with its error; it returns n.
func (n *Note) MustRaise() *Note {
	return n.mustDo(n.Raise())
}

// Lower lowers the alteration of n by 1, keeping the Octave; see Raise.
func (n *Note) Lower() error {
	return n.alter("Lower", spelling.Spelling.Lower)
}

// MustLower is like Lower but panics with its error; it returns n.
func (n *Note) MustLower() *Note {
	return n.mustDo(n.Lower())
}

// RaiseBy raises the alteration of n by k, keeping the Octave; a negative k
// lowers it. See Raise.
func (n *Note) RaiseBy(k int) error {
	return n.alter("RaiseBy", func(s spelling.Spelling) (spelling.Spelling, error) { return s.RaiseBy(k) })
}

// MustRaiseBy is like RaiseBy but panics with its error; it returns n.
func (n *Note) MustRaiseBy(k int) *Note {
	return n.mustDo(n.RaiseBy(k))
}

// LowerBy lowers the alteration of n by k, keeping the Octave; a negative k
// raises it. Every int is accepted, including math.MinInt. See Raise.
func (n *Note) LowerBy(k int) error {
	return n.alter("LowerBy", func(s spelling.Spelling) (spelling.Spelling, error) { return s.LowerBy(k) })
}

// MustLowerBy is like LowerBy but panics with its error; it returns n.
func (n *Note) MustLowerBy(k int) *Note {
	return n.mustDo(n.LowerBy(k))
}

// RaiseByAlteration raises the alteration of n by the value d, keeping the
// Octave. See Raise.
func (n *Note) RaiseByAlteration(d alteration.Alteration) error {
	return n.alter("RaiseByAlteration", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.RaiseByAlteration(d)
	})
}

// MustRaiseByAlteration is like RaiseByAlteration but panics with its error;
// it returns n.
func (n *Note) MustRaiseByAlteration(d alteration.Alteration) *Note {
	return n.mustDo(n.RaiseByAlteration(d))
}

// LowerByAlteration lowers the alteration of n by the value d, keeping the
// Octave. See Raise.
func (n *Note) LowerByAlteration(d alteration.Alteration) error {
	return n.alter("LowerByAlteration", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.LowerByAlteration(d)
	})
}

// MustLowerByAlteration is like LowerByAlteration but panics with its error;
// it returns n.
func (n *Note) MustLowerByAlteration(d alteration.Alteration) *Note {
	return n.mustDo(n.LowerByAlteration(d))
}

// SetAlteration replaces the alteration of n by a, keeping the letter and the
// Octave.
//
// Errors, in this order: ErrUnset for the note; ErrUnsetSettings;
// ErrInadmissible for the new name.
func (n *Note) SetAlteration(a alteration.Alteration) error {
	return n.alter("SetAlteration", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.WithAlteration(a), nil
	})
}

// MustSetAlteration is like SetAlteration but panics with its error; it
// returns n.
func (n *Note) MustSetAlteration(a alteration.Alteration) *Note {
	return n.mustDo(n.SetAlteration(a))
}

// alterWritten changes the alteration of n keeping its written octave.
func (n *Note) alterWritten(op string, f func(spelling.Spelling) (spelling.Spelling, error)) error {
	return n.changeWritten(op, nil, false, noCheck, f)
}

// RaiseWritten raises the alteration of n by 1, keeping the written octave,
// as system.System.RaiseWritten: B3 becomes B♯ in octave 4, written "B♯3".
// A note without an octave changes as with Raise.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource for an unbound
// note with an octave; ErrUnsetSettings; ErrInadmissible for the name of n;
// ErrOutOfRange for an alteration out of range; ErrInadmissible for the new
// name; ErrOutOfRange for the Octave (also system.ErrOutOfRange).
func (n *Note) RaiseWritten() error {
	return n.alterWritten("RaiseWritten", spelling.Spelling.Raise)
}

// MustRaiseWritten is like RaiseWritten but panics with its error; it returns
// n.
func (n *Note) MustRaiseWritten() *Note {
	return n.mustDo(n.RaiseWritten())
}

// LowerWritten lowers the alteration of n by 1, keeping the written octave;
// see RaiseWritten.
func (n *Note) LowerWritten() error {
	return n.alterWritten("LowerWritten", spelling.Spelling.Lower)
}

// MustLowerWritten is like LowerWritten but panics with its error; it returns
// n.
func (n *Note) MustLowerWritten() *Note {
	return n.mustDo(n.LowerWritten())
}

// RaiseWrittenBy raises the alteration of n by k, keeping the written octave;
// see RaiseWritten.
func (n *Note) RaiseWrittenBy(k int) error {
	return n.alterWritten("RaiseWrittenBy", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.RaiseBy(k)
	})
}

// MustRaiseWrittenBy is like RaiseWrittenBy but panics with its error; it
// returns n.
func (n *Note) MustRaiseWrittenBy(k int) *Note {
	return n.mustDo(n.RaiseWrittenBy(k))
}

// LowerWrittenBy lowers the alteration of n by k, keeping the written octave;
// see RaiseWritten.
func (n *Note) LowerWrittenBy(k int) error {
	return n.alterWritten("LowerWrittenBy", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.LowerBy(k)
	})
}

// MustLowerWrittenBy is like LowerWrittenBy but panics with its error; it
// returns n.
func (n *Note) MustLowerWrittenBy(k int) *Note {
	return n.mustDo(n.LowerWrittenBy(k))
}

// RaiseWrittenByAlteration raises the alteration of n by the value d, keeping
// the written octave; see RaiseWritten.
func (n *Note) RaiseWrittenByAlteration(d alteration.Alteration) error {
	return n.alterWritten("RaiseWrittenByAlteration", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.RaiseByAlteration(d)
	})
}

// MustRaiseWrittenByAlteration is like RaiseWrittenByAlteration but panics
// with its error; it returns n.
func (n *Note) MustRaiseWrittenByAlteration(d alteration.Alteration) *Note {
	return n.mustDo(n.RaiseWrittenByAlteration(d))
}

// LowerWrittenByAlteration lowers the alteration of n by the value d,
// keeping the written octave; see RaiseWritten.
func (n *Note) LowerWrittenByAlteration(d alteration.Alteration) error {
	return n.alterWritten("LowerWrittenByAlteration", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.LowerByAlteration(d)
	})
}

// MustLowerWrittenByAlteration is like LowerWrittenByAlteration but panics
// with its error; it returns n.
func (n *Note) MustLowerWrittenByAlteration(d alteration.Alteration) *Note {
	return n.mustDo(n.LowerWrittenByAlteration(d))
}

// SetAlterationWritten replaces the alteration of n by a, keeping the letter
// and the written octave; see RaiseWritten.
func (n *Note) SetAlterationWritten(a alteration.Alteration) error {
	return n.alterWritten("SetAlterationWritten", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.WithAlteration(a), nil
	})
}

// MustSetAlterationWritten is like SetAlterationWritten but panics with its
// error; it returns n.
func (n *Note) MustSetAlterationWritten(a alteration.Alteration) *Note {
	return n.mustDo(n.SetAlterationWritten(a))
}

// RaiseWrittenByAlterationIn is like RaiseWrittenByAlteration in the settings
// of the explicit source src; a negative d lowers the alteration. The binding
// of n does not change; the new name must be admissible in the settings of src
// and, for a bound note, in its own settings. A nil src reports ErrUnset.
func (n *Note) RaiseWrittenByAlterationIn(src settings.Source, d alteration.Alteration) error {
	return n.changeWritten("RaiseWrittenByAlterationIn", src, true, noCheck,
		func(s spelling.Spelling) (spelling.Spelling, error) { return s.RaiseByAlteration(d) })
}

// MustRaiseWrittenByAlterationIn is like RaiseWrittenByAlterationIn but
// panics with its error; it returns n.
func (n *Note) MustRaiseWrittenByAlterationIn(src settings.Source, d alteration.Alteration) *Note {
	return n.mustDo(n.RaiseWrittenByAlterationIn(src, d))
}

// SetAlterationWrittenIn is like SetAlterationWritten in the settings of the
// explicit source src; see RaiseWrittenByAlterationIn.
func (n *Note) SetAlterationWrittenIn(src settings.Source, a alteration.Alteration) error {
	return n.changeWritten("SetAlterationWrittenIn", src, true, noCheck,
		func(s spelling.Spelling) (spelling.Spelling, error) { return s.WithAlteration(a), nil })
}

// MustSetAlterationWrittenIn is like SetAlterationWrittenIn but panics with
// its error; it returns n.
func (n *Note) MustSetAlterationWrittenIn(src settings.Source, a alteration.Alteration) *Note {
	return n.mustDo(n.SetAlterationWrittenIn(src, a))
}

// Naturalize removes the sign of n, keeping the Octave: C♭4 (octave 3)
// becomes C3. A natural name is admissible in every system.
//
// Errors, in this order: ErrUnset for the note; ErrUnsetSettings.
func (n *Note) Naturalize() error {
	return n.alter("Naturalize", func(s spelling.Spelling) (spelling.Spelling, error) { return s.Natural(), nil })
}

// MustNaturalize is like Naturalize but panics with its error; it returns n.
func (n *Note) MustNaturalize() *Note {
	return n.mustDo(n.Naturalize())
}

// NaturalizeWritten removes the sign of n, keeping the written octave
// (system.System.NaturalWritten): C♭4 (octave 3) becomes C4. A note without an
// octave changes as with Naturalize.
//
// Errors, in this order: ErrUnset for the note; ErrNoSource for an unbound
// note with an octave; ErrUnsetSettings; ErrInadmissible for the name of n;
// ErrOutOfRange (also system.ErrOutOfRange).
func (n *Note) NaturalizeWritten() error {
	return n.alterWritten("NaturalizeWritten", func(s spelling.Spelling) (spelling.Spelling, error) {
		return s.Natural(), nil
	})
}

// MustNaturalizeWritten is like NaturalizeWritten but panics with its error;
// it returns n.
func (n *Note) MustNaturalizeWritten() *Note {
	return n.mustDo(n.NaturalizeWritten())
}

// NaturalizeWrittenIn is like NaturalizeWritten in the settings of the
// explicit source src. A bound note also reads its own source (once when it is
// one source with src). A nil src reports ErrUnset.
func (n *Note) NaturalizeWrittenIn(src settings.Source) error {
	return n.changeWritten("NaturalizeWrittenIn", src, true, noCheck,
		func(s spelling.Spelling) (spelling.Spelling, error) { return s.Natural(), nil })
}

// MustNaturalizeWrittenIn is like NaturalizeWrittenIn but panics with its
// error; it returns n.
func (n *Note) MustNaturalizeWrittenIn(src settings.Source) *Note {
	return n.mustDo(n.NaturalizeWrittenIn(src))
}

// shiftOctave moves the Octave of n by f.
func (n *Note) shiftOctave(op string, f func(octave.Octave) (octave.Octave, error)) error {
	cur, err := n.current(op)
	if err != nil {
		return err
	}

	if err := checkOctave(op, cur); err != nil {
		return err
	}

	o, err := f(cur.oct)
	if err != nil {
		return depError(op, err)
	}

	cur.oct = o
	*n = cur

	return nil
}

// Shift moves n by k octaves up; a negative k moves it down. The source is
// not read.
//
// Errors, in this order: ErrUnset for the note; ErrNoOctave; ErrOutOfRange
// (also octave.ErrOutOfRange).
func (n *Note) Shift(k int) error {
	return n.shiftOctave("Shift", func(o octave.Octave) (octave.Octave, error) { return o.Shift(k) })
}

// MustShift is like Shift but panics with its error; it returns n.
func (n *Note) MustShift(k int) *Note {
	return n.mustDo(n.Shift(k))
}

// ShiftDown moves n by k octaves down; a negative k moves it up.
// ShiftDown(math.MinInt) is exact. See Shift.
func (n *Note) ShiftDown(k int) error {
	return n.shiftOctave("ShiftDown", func(o octave.Octave) (octave.Octave, error) { return o.ShiftDown(k) })
}

// MustShiftDown is like ShiftDown but panics with its error; it returns n.
func (n *Note) MustShiftDown(k int) *Note {
	return n.mustDo(n.ShiftDown(k))
}

// Up moves n one octave up. See Shift.
func (n *Note) Up() error {
	return n.shiftOctave("Up", octave.Octave.Up)
}

// MustUp is like Up but panics with its error; it returns n.
func (n *Note) MustUp() *Note {
	return n.mustDo(n.Up())
}

// Down moves n one octave down. See Shift.
func (n *Note) Down() error {
	return n.shiftOctave("Down", octave.Octave.Down)
}

// MustDown is like Down but panics with its error; it returns n.
func (n *Note) MustDown() *Note {
	return n.mustDo(n.Down())
}

// move moves n by k steps of the system, naming the result by rule.
func (n *Note) move(op string, explicit settings.Source, inForm bool, k system.Step, rule system.SpellingRule) error {
	cur, err := n.current(op)
	if err != nil {
		return err
	}

	if inForm {
		if err := checkSource(op, explicit); err != nil {
			return err
		}
	}

	if !rule.IsValid() {
		return unsetError(op, "spelling rule")
	}

	r := newReader(op)

	var (
		f      frame
		ownSys *system.System
	)

	if inForm {
		f, err = sourceFrame(r, explicit)
		if err != nil {
			return err
		}

		if cur.src != nil {
			own, err := r.read(cur.src)
			if err != nil {
				return err
			}

			ownSys = own.System()
		}
	} else {
		f, err = ownFrame(r, cur)
		if err != nil {
			return err
		}
	}

	if err := admit(op, f.sys, cur.name); err != nil {
		return err
	}

	class, err := f.sys.Class(cur.name)
	if err != nil {
		return depError(op, err)
	}

	name, err := f.sys.Spell(class.Transpose(int(k)), rule)
	if err != nil {
		return depError(op, err)
	}

	if ownSys != nil {
		if err := admit(op, ownSys, name); err != nil {
			return err
		}
	}

	res := cur
	res.name = name

	if cur.oct.IsValid() {
		p, err := f.sys.Move(cur.pitch(), k, rule)
		if err != nil {
			return depError(op, err)
		}

		res = res.withPitch(p)
	}

	*n = res

	return nil
}

// Move moves n by k steps of the system of its settings
// (system.System.Move): a note with an octave goes to the position P + k, a
// note without an octave to the class κ + k. The name of the result is chosen
// by rule, also for k = 0.
//
// Errors, in this order: ErrUnset for the note or an unset rule;
// ErrNoSource; ErrUnsetSettings; ErrInadmissible for the name of n; the
// errors of the rule (system.ErrNotApplicable, system.ErrInadmissible for its
// tonic, system.ErrNoSpelling, system.ErrBadSpelling) without a sentinel of
// note; ErrOutOfRange for the Octave (also system.ErrOutOfRange).
func (n *Note) Move(k system.Step, rule system.SpellingRule) error {
	return n.move("Move", nil, false, k, rule)
}

// MustMove is like Move but panics with its error; it returns n.
func (n *Note) MustMove(k system.Step, rule system.SpellingRule) *Note {
	return n.mustDo(n.Move(k, rule))
}

// MoveIn is like Move in the settings of the explicit source src; the binding
// of n does not change, and the name of the result must be admissible in the
// own settings of a bound note. A nil src reports ErrUnset.
func (n *Note) MoveIn(src settings.Source, k system.Step, rule system.SpellingRule) error {
	return n.move("MoveIn", src, true, k, rule)
}

// MustMoveIn is like MoveIn but panics with its error; it returns n.
func (n *Note) MustMoveIn(src settings.Source, k system.Step, rule system.SpellingRule) *Note {
	return n.mustDo(n.MoveIn(src, k, rule))
}
