package pitch

import (
	"fmt"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/spelling"
)

// WithSpelling returns a pitch with spelling s and the positional octave of p.
// It preserves the positional octave, not the written register. If the new
// spelling crosses a positional-octave boundary, the written register and the
// system position (and therefore the sounding pitch at a fixed tuning) can
// change even though the positional octave stays the same.
//
// If p or s is invalid, WithSpelling returns Pitch{} and an error matching
// [ErrInvalid].
func (p Pitch) WithSpelling(s spelling.Spelling) (Pitch, error) {
	if !p.IsValid() || !s.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: with spelling: %w", ErrInvalid)
	}

	return Pitch{s: s, o: p.o}, nil
}

// MustWithSpelling is like [Pitch.WithSpelling] but panics with the returned
// error if p or s is invalid. It preserves the positional octave, not the
// written register.
func (p Pitch) MustWithSpelling(s spelling.Spelling) Pitch {
	return must(p.WithSpelling(s))
}

// WithOctave returns a pitch with the spelling of p and positional octave o.
//
// If p or o is invalid, WithOctave returns Pitch{} and an error matching
// [ErrInvalid].
func (p Pitch) WithOctave(o octave.Octave) (Pitch, error) {
	if !p.IsValid() || !o.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: with octave: %w", ErrInvalid)
	}

	return Pitch{s: p.s, o: o}, nil
}

// MustWithOctave is like [Pitch.WithOctave] but panics with the returned error
// if p or o is invalid.
func (p Pitch) MustWithOctave(o octave.Octave) Pitch {
	return must(p.WithOctave(o))
}

// WithASPN returns a pitch with the spelling of p and the positional octave
// whose ASPN number is aspn.
//
// If p is invalid, WithASPN returns Pitch{} and an error matching
// [ErrInvalid].
func (p Pitch) WithASPN(aspn int) (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: with ASPN: %w", ErrInvalid)
	}

	return Pitch{s: p.s, o: octave.FromASPN(aspn)}, nil
}

// MustWithASPN is like [Pitch.WithASPN] but panics with the returned error if
// p is invalid.
func (p Pitch) MustWithASPN(aspn int) Pitch {
	return must(p.WithASPN(aspn))
}

// Natural returns the pitch with the same letter and positional octave as p
// and alteration value 0 (natural). It preserves the positional octave, not
// the written register. If the natural spelling crosses a positional-octave
// boundary, the written register and the system position (and therefore the
// sounding pitch at a fixed tuning) can change even though the positional
// octave stays the same.
//
// Letter and octave are preserved without enharmonic substitution: for
// (B♯, ASPN 4), Natural returns (B, ASPN 4), not C. For the zero value
// Pitch{}, Natural returns Pitch{}.
func (p Pitch) Natural() Pitch {
	if !p.IsValid() {
		return Pitch{}
	}

	return Pitch{s: p.s.Natural(), o: p.o}
}

// Raise returns p with its alteration value raised by 1 chromatic unit,
// keeping the same letter and positional octave (for example, (B, ASPN 4)
// becomes (B♯, ASPN 4)). It preserves the positional octave, not the written
// register. If the raised spelling crosses a positional-octave boundary, the
// written register and the system position (and therefore the sounding pitch
// at a fixed tuning) can change even though the positional octave stays the
// same.
//
// If p is invalid, Raise returns Pitch{} and an error matching [ErrInvalid].
// If the resulting alteration exceeds the guaranteed range of package
// [github.com/go-muse/musex/alteration], Raise returns Pitch{} and an error
// matching [github.com/go-muse/musex/alteration.ErrOutOfRange].
func (p Pitch) Raise() (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: raise: %w", ErrInvalid)
	}

	s, err := p.s.Raise()
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: raise: %w", err)
	}

	return Pitch{s: s, o: p.o}, nil
}

// MustRaise is like [Pitch.Raise] but panics with the returned error if Raise
// fails. It preserves the positional octave, not the written register.
func (p Pitch) MustRaise() Pitch {
	return must(p.Raise())
}

// Lower returns p with its alteration value lowered by 1 chromatic unit,
// keeping the same letter and positional octave (for example, (C, ASPN 4)
// becomes (C♭, ASPN 4)). It preserves the positional octave, not the written
// register. If the lowered spelling crosses a positional-octave boundary, the
// written register and the system position (and therefore the sounding pitch
// at a fixed tuning) can change even though the positional octave stays the
// same.
//
// If p is invalid, Lower returns Pitch{} and an error matching [ErrInvalid].
// If the resulting alteration exceeds the guaranteed range of package
// [github.com/go-muse/musex/alteration], Lower returns Pitch{} and an error
// matching [github.com/go-muse/musex/alteration.ErrOutOfRange].
func (p Pitch) Lower() (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: lower: %w", ErrInvalid)
	}

	s, err := p.s.Lower()
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: lower: %w", err)
	}

	return Pitch{s: s, o: p.o}, nil
}

// MustLower is like [Pitch.Lower] but panics with the returned error if Lower
// fails. It preserves the positional octave, not the written register.
func (p Pitch) MustLower() Pitch {
	return must(p.Lower())
}

// RaiseBy returns p with its alteration value raised by n chromatic units (or
// lowered for negative n), keeping the same letter and positional octave. It
// preserves the positional octave, not the written register. If the new
// spelling crosses a positional-octave boundary, the written register and the
// system position (and therefore the sounding pitch at a fixed tuning) can
// change even though the positional octave stays the same.
//
// Validity of p is checked before range: if p is invalid, RaiseBy returns
// Pitch{} and an error matching [ErrInvalid] for any n. If p is valid and the
// resulting alteration exceeds the guaranteed range of package
// [github.com/go-muse/musex/alteration], RaiseBy returns Pitch{} and an error
// matching [github.com/go-muse/musex/alteration.ErrOutOfRange].
func (p Pitch) RaiseBy(n int) (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: raise by %d: %w", n, ErrInvalid)
	}

	s, err := p.s.RaiseBy(n)
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: raise by %d: %w", n, err)
	}

	return Pitch{s: s, o: p.o}, nil
}

// MustRaiseBy is like [Pitch.RaiseBy] but panics with the returned error if
// RaiseBy fails. It preserves the positional octave, not the written register.
func (p Pitch) MustRaiseBy(n int) Pitch {
	return must(p.RaiseBy(n))
}

// LowerBy returns p with its alteration value lowered by n chromatic units (or
// raised for negative n), keeping the same letter and positional octave.
// LowerBy handles n == math.MinInt exactly without sign overflow. It preserves
// the positional octave, not the written register. If the new spelling crosses
// a positional-octave boundary, the written register and the system position
// (and therefore the sounding pitch at a fixed tuning) can change even though
// the positional octave stays the same.
//
// Validity of p is checked before range: if p is invalid, LowerBy returns
// Pitch{} and an error matching [ErrInvalid] for any n. If p is valid and the
// resulting alteration exceeds the guaranteed range of package
// [github.com/go-muse/musex/alteration], LowerBy returns Pitch{} and an error
// matching [github.com/go-muse/musex/alteration.ErrOutOfRange].
func (p Pitch) LowerBy(n int) (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: lower by %d: %w", n, ErrInvalid)
	}

	s, err := p.s.LowerBy(n)
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: lower by %d: %w", n, err)
	}

	return Pitch{s: s, o: p.o}, nil
}

// MustLowerBy is like [Pitch.LowerBy] but panics with the returned error if
// LowerBy fails. It preserves the positional octave, not the written register.
func (p Pitch) MustLowerBy(n int) Pitch {
	return must(p.LowerBy(n))
}
