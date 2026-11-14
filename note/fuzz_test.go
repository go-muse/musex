package note_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
)

// Fuzz tests check properties (plan, Sh1 "Fuzzing"), not a copy of an
// algorithm: the agreed order is antisymmetric and agrees with Equal; raising
// and lowering, shifting up and down return the same note or fail with E-8
// without a change; copies are independent.

// fuzzNote builds an unbound note from fuzz inputs; kind 0 is the unset note,
// an odd kind has an octave.
func fuzzNote(kind uint8, ord uint8, alt int16, aspn int, c int32) note.Note {
	if kind%4 == 0 {
		return note.Note{}
	}
	l := letter.MustFromOrdinal(int(ord % letter.Count))
	a, err := alteration.FromInt(int(alt))
	if err != nil {
		return note.Note{}
	}
	name := spelling.MustNew(l, a)
	x, err := cents.FromInt(int(c))
	if err != nil {
		x = cents.Zero()
	}
	if kind%2 == 1 {
		p, err := pitch.FromASPN(name, aspn)
		if err != nil {
			return note.Note{}
		}
		return note.MustFromPitch(p, x, nil)
	}
	return note.MustFromSpelling(name, x, nil)
}

func FuzzCompare(f *testing.F) {
	f.Add(uint8(1), uint8(0), int16(0), 4, int32(0), uint8(1), uint8(0), int16(0), 4, int32(0))
	f.Add(uint8(1), uint8(6), int16(1), 3, int32(0), uint8(1), uint8(0), int16(0), 4, int32(0))
	f.Add(uint8(2), uint8(0), int16(1), 0, int32(0), uint8(3), uint8(0), int16(1), 4, int32(0))
	f.Add(uint8(0), uint8(0), int16(0), 0, int32(0), uint8(1), uint8(2), int16(-2), -100, int32(-50))
	f.Add(uint8(1), uint8(0), int16(0), 4, int32(10), uint8(1), uint8(0), int16(0), 4, int32(-10))

	f.Fuzz(func(t *testing.T, k1, o1 uint8, a1 int16, p1 int, c1 int32, k2, o2 uint8, a2 int16, p2 int, c2 int32) {
		x, y := fuzzNote(k1, o1, a1, p1, c1), fuzzNote(k2, o2, a2, p2, c2)
		xy, yx := x.Compare(y), y.Compare(x)
		if xy != -yx {
			t.Fatalf("Compare is not antisymmetric: %v vs %v: %d, %d", x, y, xy, yx)
		}
		if xy < -1 || xy > 1 {
			t.Fatalf("Compare out of -1..1: %d", xy)
		}
		if x.IsValid() && y.IsValid() && (xy == 0) != x.Equal(y) {
			t.Fatalf("Compare == 0 differs from Equal: %v vs %v", x, y)
		}
		if x.Equal(y) != y.Equal(x) {
			t.Fatalf("Equal is not symmetric: %v vs %v", x, y)
		}
	})
}

func FuzzRoundTrip(f *testing.F) {
	f.Add(uint8(1), uint8(0), int16(0), 4, int32(30), 1)
	f.Add(uint8(3), uint8(6), int16(2), -7, int32(0), -5)
	f.Add(uint8(2), uint8(3), int16(-1), 0, int32(1), 1<<30)
	f.Add(uint8(1), uint8(0), int16(0), 1<<62, int32(0), 1<<62)

	f.Fuzz(func(t *testing.T, kind, ord uint8, alt int16, aspn int, c int32, k int) {
		n := fuzzNote(kind, ord, alt, aspn, c)
		if !n.IsValid() {
			return
		}
		check := func(name string, there func(*note.Note) error, back func(*note.Note) error) {
			m := n
			if err := there(&m); err != nil {
				if !isOutOfRange(err) && !isNoOctave(err, n) {
					t.Fatalf("%s: unexpected error %v", name, err)
				}
				if m != n {
					t.Fatalf("%s: a failed change changed the note", name)
				}
				return
			}
			if err := back(&m); err != nil {
				t.Fatalf("%s: the way back failed: %v", name, err)
			}
			if !m.Equal(n) || m != n {
				t.Fatalf("%s: %v came back as %v", name, n, m)
			}
		}
		check("RaiseBy/LowerBy",
			func(m *note.Note) error { return m.RaiseBy(k) },
			func(m *note.Note) error { return m.LowerBy(k) })
		check("Shift/ShiftDown",
			func(m *note.Note) error { return m.Shift(k) },
			func(m *note.Note) error { return m.ShiftDown(k) })
	})
}

func FuzzCopy(f *testing.F) {
	f.Add(uint8(1), uint8(0), int16(0), 4, int32(30), uint8(0b1011), 3)
	f.Add(uint8(2), uint8(5), int16(-1), 0, int32(-7), uint8(0xff), -2)

	f.Fuzz(func(t *testing.T, kind, ord uint8, alt int16, aspn int, c int32, ops uint8, k int) {
		n := fuzzNote(kind, ord, alt, aspn, c)
		before := n
		m := n
		for i := range 8 {
			switch (ops >> i) & 1 {
			case 0:
				_ = m.RaiseBy(k % 3)
			default:
				_ = m.Shift(k % 5)
			}
			_ = m.SetCents(cents.MustFromInt(i))
		}
		if n != before {
			t.Fatalf("changing a copy changed the original")
		}
		_ = n.RemoveOctave()
		if n.IsValid() && m.IsValid() && n.HasOctave() {
			t.Fatalf("RemoveOctave failed on the original")
		}
	})
}

// isOutOfRange reports whether err is E-8.
func isOutOfRange(err error) bool { return errors.Is(err, note.ErrOutOfRange) }

// isNoOctave reports whether err is E-4 of a note without octave.
func isNoOctave(err error, n note.Note) bool {
	return !n.HasOctave() && errors.Is(err, note.ErrNoOctave)
}
