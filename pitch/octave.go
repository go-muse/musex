package pitch

import (
	"fmt"
	"iter"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/spelling"
)

// Shift returns p shifted by k octaves upward (or downward for negative k),
// keeping the same spelling.
//
// Validity of p is checked before range: if p is invalid, Shift returns
// Pitch{} and an error matching [ErrInvalid] for any k. If p is valid and the
// resulting octave overflows int, Shift returns Pitch{} and an error matching
// [octave.ErrOutOfRange].
func (p Pitch) Shift(k int) (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: shift %d: %w", k, ErrInvalid)
	}

	o, err := p.o.Shift(k)
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: shift %d: %w", k, err)
	}

	return Pitch{s: p.s, o: o}, nil
}

// MustShift is like [Pitch.Shift] but panics with the returned error if Shift
// fails.
func (p Pitch) MustShift(k int) Pitch {
	return must(p.Shift(k))
}

// ShiftDown returns p shifted by k octaves downward (or upward for negative
// k), keeping the same spelling. ShiftDown handles k == math.MinInt exactly
// without sign overflow and reverses [Pitch.Shift] across the entire int
// range.
//
// Validity of p is checked before range: if p is invalid, ShiftDown returns
// Pitch{} and an error matching [ErrInvalid] for any k. If p is valid and the
// resulting octave overflows int, ShiftDown returns Pitch{} and an error
// matching [octave.ErrOutOfRange].
func (p Pitch) ShiftDown(k int) (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: shift down %d: %w", k, ErrInvalid)
	}

	o, err := p.o.ShiftDown(k)
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: shift down %d: %w", k, err)
	}

	return Pitch{s: p.s, o: o}, nil
}

// MustShiftDown is like [Pitch.ShiftDown] but panics with the returned error
// if ShiftDown fails.
func (p Pitch) MustShiftDown(k int) Pitch {
	return must(p.ShiftDown(k))
}

// Up returns p in the next higher octave, equivalent to p.Shift(1), keeping
// the same spelling.
//
// If p is invalid, Up returns Pitch{} and an error matching [ErrInvalid]. If
// p is in the highest representable octave (ASPN math.MaxInt), Up returns
// Pitch{} and an error matching [octave.ErrOutOfRange].
func (p Pitch) Up() (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: up: %w", ErrInvalid)
	}

	o, err := p.o.Up()
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: up: %w", err)
	}

	return Pitch{s: p.s, o: o}, nil
}

// MustUp is like [Pitch.Up] but panics with the returned error if Up fails.
func (p Pitch) MustUp() Pitch {
	return must(p.Up())
}

// Down returns p in the next lower octave, equivalent to p.Shift(-1), keeping
// the same spelling.
//
// If p is invalid, Down returns Pitch{} and an error matching [ErrInvalid]. If
// p is in the lowest representable octave (ASPN math.MinInt), Down returns
// Pitch{} and an error matching [octave.ErrOutOfRange].
func (p Pitch) Down() (Pitch, error) {
	if !p.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: down: %w", ErrInvalid)
	}

	o, err := p.o.Down()
	if err != nil {
		return Pitch{}, fmt.Errorf("pitch: down: %w", err)
	}

	return Pitch{s: p.s, o: o}, nil
}

// MustDown is like [Pitch.Down] but panics with the returned error if Down
// fails.
func (p Pitch) MustDown() Pitch {
	return must(p.Down())
}

// OctaveSub returns the signed difference in octaves between p and other
// (the ASPN number of p's octave minus the ASPN number of other's octave).
//
// OctaveSub depends only on the positional octaves of p and other, not on
// their spellings: for any two valid pitches in the same octave (such as
// (C, ASPN 4) and (B♯, ASPN 4)), OctaveSub returns 0.
//
// Validity of both operands is checked before range: if p or other is invalid,
// OctaveSub returns 0 and an error matching [ErrInvalid]. If both are valid
// and the difference does not fit in an int, OctaveSub returns 0 and an error
// matching [octave.ErrOutOfRange].
func (p Pitch) OctaveSub(other Pitch) (int, error) {
	if !p.IsValid() || !other.IsValid() {
		return 0, fmt.Errorf("pitch: octave sub: %w", ErrInvalid)
	}

	d, err := p.o.Sub(other.o)
	if err != nil {
		return 0, fmt.Errorf("pitch: octave sub: %w", err)
	}

	return d, nil
}

// MustOctaveSub is like [Pitch.OctaveSub] but panics with the returned error
// if OctaveSub fails.
func (p Pitch) MustOctaveSub(other Pitch) int {
	return must(p.OctaveSub(other))
}

// OctaveDistance returns the unsigned distance in octaves between p and other
// (the absolute value of p.OctaveSub(other)).
//
// OctaveDistance depends only on the positional octaves of p and other, not on
// their spellings, and is independent of operand order. Because |math.MinInt|
// does not fit in an int, OctaveDistance fails with [octave.ErrOutOfRange]
// when the signed difference is math.MinInt (for example between ASPN
// math.MinInt and ASPN 0), even though OctaveSub succeeds there.
//
// Validity of both operands is checked before range: if p or other is invalid,
// OctaveDistance returns 0 and an error matching [ErrInvalid]. If both are
// valid and the distance does not fit in an int, OctaveDistance returns 0 and
// an error matching [octave.ErrOutOfRange].
func (p Pitch) OctaveDistance(other Pitch) (int, error) {
	if !p.IsValid() || !other.IsValid() {
		return 0, fmt.Errorf("pitch: octave distance: %w", ErrInvalid)
	}

	d, err := p.o.Distance(other.o)
	if err != nil {
		return 0, fmt.Errorf("pitch: octave distance: %w", err)
	}

	return d, nil
}

// MustOctaveDistance is like [Pitch.OctaveDistance] but panics with the
// returned error if OctaveDistance fails.
func (p Pitch) MustOctaveDistance(other Pitch) int {
	return must(p.OctaveDistance(other))
}

// Range returns a lazy sequence of pitches with spelling s for every octave
// from from to to inclusive, stepping by one octave upward when to is above
// from, downward when to is below from, or yielding a single pitch when from
// and to are equal.
//
// The returned sequence runs in the caller's goroutine without allocating a
// slice for the range, stops immediately when the loop breaks, terminates
// cleanly after yielding to at math.MinInt or math.MaxInt, and can be iterated
// repeatedly and concurrently.
//
// If s, from, or to is invalid, Range returns nil and an error matching
// [ErrInvalid].
func Range(s spelling.Spelling, from, to octave.Octave) (iter.Seq[Pitch], error) {
	if !s.IsValid() || !from.IsValid() || !to.IsValid() {
		return nil, fmt.Errorf("pitch: range: %w", ErrInvalid)
	}

	octSeq := octave.MustRange(from, to)

	return func(yield func(Pitch) bool) {
		for o := range octSeq {
			if !yield(Pitch{s: s, o: o}) {
				return
			}
		}
	}, nil
}

// MustRange is like [Range] but panics with the returned error if s, from, or
// to is invalid.
func MustRange(s spelling.Spelling, from, to octave.Octave) iter.Seq[Pitch] {
	return must(Range(s, from, to))
}
