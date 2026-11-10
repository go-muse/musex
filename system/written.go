package system

import (
	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
)

// WrittenOctave returns the written octave w of p: the octave whose number is
// written with the letter, w = Octave − k, where k is the carry of the name
// (see Carry). In the traditional system the pitch (B♯, Octave 4) is written
// "B♯3", so its written octave is 3.
//
// Errors, in this order: ErrInvalid when s or p is unset; ErrInadmissible when
// the name of p is not admissible; ErrOutOfRange when w does not fit in int.
func (s *System) WrittenOctave(p pitch.Pitch) (octave.Octave, error) {
	if err := s.checkPitches(p); err != nil {
		return octave.Octave{}, err
	}

	o, k, err := s.carried(p)
	if err != nil {
		return octave.Octave{}, err
	}

	w, ok := subInt(o, k)
	if !ok {
		return octave.Octave{}, errorf(ErrOutOfRange, "written octave of %v", p)
	}

	return octave.FromASPN(w), nil
}

// MustWrittenOctave is like WrittenOctave but panics on error.
func (s *System) MustWrittenOctave(p pitch.Pitch) octave.Octave {
	return must(s.WrittenOctave(p))
}

// FromWritten returns the pitch with the name name written in the octave
// written: the pitch (name, Octave = written + k), where k is the carry of
// name. In the traditional system "B♯3" is the pitch (B♯, Octave 4). It is the
// inverse of WrittenOctave.
//
// Errors, in this order: ErrInvalid when s, name or written is unset;
// ErrInadmissible when name is not admissible; ErrOutOfRange when the Octave
// does not fit in int.
func (s *System) FromWritten(name spelling.Spelling, written octave.Octave) (pitch.Pitch, error) {
	if err := s.checkName(name); err != nil {
		return pitch.Pitch{}, err
	}

	w, err := written.ASPN()
	if err != nil {
		return pitch.Pitch{}, errorf(ErrInvalid, "unset written octave")
	}

	sh, err := s.admittedShift(name)
	if err != nil {
		return pitch.Pitch{}, err
	}

	o, ok := addInt(w, int(floorDiv(sh, int64(s.n))))
	if !ok {
		return pitch.Pitch{}, errorf(ErrOutOfRange, "%v written in %v", name, written)
	}

	return pitch.MustNew(name, octave.FromASPN(o)), nil
}

// MustFromWritten is like FromWritten but panics on error.
func (s *System) MustFromWritten(name spelling.Spelling, written octave.Octave) pitch.Pitch {
	return must(s.FromWritten(name, written))
}

// RaiseWritten returns p with the alteration value raised by 1 and the same
// written octave: in the traditional system the record "B3" becomes "B♯3", the
// pitch (B♯, Octave 4), one step above. This differs from pitch.Pitch.Raise,
// which keeps the Octave and gives (B♯, Octave 3), written "B♯2".
//
// For a pitch (l, a, Octave) and the resulting name t, the result is
// (t, Octave + k(t) − k(l, a)); its position is the position of p plus
// s(t) − s(l, a), where s = p(letter) + a·c. The written octave is used only
// through this difference of carries, so a written octave outside int is not
// an error when the result fits.
//
// Errors, in this order: ErrInvalid when s or p is unset; ErrInadmissible when
// the name of p is not admissible; ErrOutOfRange when the new alteration value
// is outside the range of package alteration; ErrInadmissible when the new
// name is not admissible; ErrOutOfRange when the resulting Octave does not fit
// in int. The other methods that keep the written octave report errors in the
// same order.
func (s *System) RaiseWritten(p pitch.Pitch) (pitch.Pitch, error) {
	return s.RaiseWrittenBy(p, 1)
}

// MustRaiseWritten is like RaiseWritten but panics on error.
func (s *System) MustRaiseWritten(p pitch.Pitch) pitch.Pitch {
	return must(s.RaiseWritten(p))
}

// LowerWritten returns p with the alteration value lowered by 1 and the same
// written octave: in the traditional system "C4" becomes "C♭4", the pitch
// (C♭, Octave 3). See RaiseWritten.
func (s *System) LowerWritten(p pitch.Pitch) (pitch.Pitch, error) {
	return s.LowerWrittenBy(p, 1)
}

// MustLowerWritten is like LowerWritten but panics on error.
func (s *System) MustLowerWritten(p pitch.Pitch) pitch.Pitch {
	return must(s.LowerWritten(p))
}

// RaiseWrittenBy returns p with the alteration value raised by n (any int) and
// the same written octave. See RaiseWritten.
func (s *System) RaiseWrittenBy(p pitch.Pitch, n int) (pitch.Pitch, error) {
	return s.keepWritten(p, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.RaiseBy(n)
	})
}

// MustRaiseWrittenBy is like RaiseWrittenBy but panics on error.
func (s *System) MustRaiseWrittenBy(p pitch.Pitch, n int) pitch.Pitch {
	return must(s.RaiseWrittenBy(p, n))
}

// LowerWrittenBy returns p with the alteration value lowered by n (any int)
// and the same written octave. See RaiseWritten.
func (s *System) LowerWrittenBy(p pitch.Pitch, n int) (pitch.Pitch, error) {
	return s.keepWritten(p, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.LowerBy(n)
	})
}

// MustLowerWrittenBy is like LowerWrittenBy but panics on error.
func (s *System) MustLowerWrittenBy(p pitch.Pitch, n int) pitch.Pitch {
	return must(s.LowerWrittenBy(p, n))
}

// RaiseWrittenByAlteration returns p with the alteration value raised by d,
// which can be fractional, and the same written octave: in the 24-class
// quarter-tone system the record "D4" raised by 1/2 becomes D with +1/2
// written in octave 4, one step above. See RaiseWritten.
func (s *System) RaiseWrittenByAlteration(p pitch.Pitch, d alteration.Alteration) (pitch.Pitch, error) {
	return s.keepWritten(p, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.RaiseByAlteration(d)
	})
}

// MustRaiseWrittenByAlteration is like RaiseWrittenByAlteration but panics on
// error.
func (s *System) MustRaiseWrittenByAlteration(p pitch.Pitch, d alteration.Alteration) pitch.Pitch {
	return must(s.RaiseWrittenByAlteration(p, d))
}

// LowerWrittenByAlteration returns p with the alteration value lowered by d,
// which can be fractional, and the same written octave. See RaiseWritten.
func (s *System) LowerWrittenByAlteration(p pitch.Pitch, d alteration.Alteration) (pitch.Pitch, error) {
	return s.keepWritten(p, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.LowerByAlteration(d)
	})
}

// MustLowerWrittenByAlteration is like LowerWrittenByAlteration but panics on
// error.
func (s *System) MustLowerWrittenByAlteration(p pitch.Pitch, d alteration.Alteration) pitch.Pitch {
	return must(s.LowerWrittenByAlteration(p, d))
}

// RespellWritten returns the pitch with the name name in the written octave of
// p: in the traditional system "C4" respelled as B♯ becomes "B♯4", the pitch
// (B♯, Octave 5), 12 steps above. See RaiseWritten; in addition, an unset
// name gives an error that matches ErrInvalid.
func (s *System) RespellWritten(p pitch.Pitch, name spelling.Spelling) (pitch.Pitch, error) {
	if err := s.checkPitches(p); err != nil {
		return pitch.Pitch{}, err
	}

	if !name.IsValid() {
		return pitch.Pitch{}, errorf(ErrInvalid, "unset name")
	}

	return s.keepWritten(p, func(spelling.Spelling) (spelling.Spelling, error) {
		return name, nil
	})
}

// MustRespellWritten is like RespellWritten but panics on error.
func (s *System) MustRespellWritten(p pitch.Pitch, name spelling.Spelling) pitch.Pitch {
	return must(s.RespellWritten(p, name))
}

// NaturalWritten returns the pitch with the natural name of the letter of p in
// the written octave of p: "B♯3" becomes "B3". The natural name is always
// admissible. See RaiseWritten.
func (s *System) NaturalWritten(p pitch.Pitch) (pitch.Pitch, error) {
	return s.keepWritten(p, func(sp spelling.Spelling) (spelling.Spelling, error) {
		return sp.Natural(), nil
	})
}

// MustNaturalWritten is like NaturalWritten but panics on error.
func (s *System) MustNaturalWritten(p pitch.Pitch) pitch.Pitch {
	return must(s.NaturalWritten(p))
}

// carried returns the Octave (ASPN) of the set pitch p and the carry of its
// name, or an error that matches ErrInadmissible.
func (s *System) carried(p pitch.Pitch) (o, k int, err error) {
	sh, err := s.admittedShift(p.Spelling())
	if err != nil {
		return 0, 0, err
	}

	return p.Octave().MustASPN(), int(floorDiv(sh, int64(s.n))), nil
}

// keepWritten returns the pitch with the name change(name of p) and the same
// written octave as p.
func (s *System) keepWritten(p pitch.Pitch, change func(spelling.Spelling) (spelling.Spelling, error)) (pitch.Pitch, error) {
	if err := s.checkPitches(p); err != nil {
		return pitch.Pitch{}, err
	}

	o, k, err := s.carried(p)
	if err != nil {
		return pitch.Pitch{}, err
	}

	t, err := change(p.Spelling())
	if err != nil {
		return pitch.Pitch{}, errorf(ErrOutOfRange, "changed alteration of %v", p)
	}

	sh, ok := s.shift(t)
	if !ok {
		return pitch.Pitch{}, errorf(ErrInadmissible, "%v in %v", t, s.name)
	}

	r, ok := addInt(o, int(floorDiv(sh, int64(s.n)))-k)
	if !ok {
		return pitch.Pitch{}, errorf(ErrOutOfRange, "octave of %v with the written octave of %v", t, p)
	}

	return pitch.MustNew(t, octave.FromASPN(r)), nil
}
