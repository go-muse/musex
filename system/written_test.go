package system_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the written register: OP-21, OP-22, OP-36 (whole changes), OP-37, OP-38,
// P-7, P-10, contract 5.9, criteria K-6 (without S5), K-13 (without S24), K-17, K-21, K-22.

// shiftFits reports whether the ASPN number w + k fits in int.
func shiftFits(w, k int) bool {
	if k > 0 {
		return w <= math.MaxInt-k
	}
	return w >= math.MinInt-k
}

func edgeOctaves() []int {
	return []int{math.MinInt, math.MinInt + 1, math.MinInt + 2, -1, 0, 4, math.MaxInt - 2, math.MaxInt - 1, math.MaxInt}
}

func TestSystem_WrittenExamples(t *testing.T) {
	t.Parallel()

	trad := func(testing.TB) *system.System { return system.Traditional() }

	t.Run("WrittenOctave (K-6)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			build func(t testing.TB) *system.System
			p     pitch.Pitch
			want  int
		}{
			{"T: (B sharp, Octave 4) is written B sharp 3", trad, pt(spelling.BSharp(), 4), 3},
			{"T: (C flat, Octave 3) is written C flat 4", trad, pt(spelling.CFlat(), 3), 4},
			{"T: (B sharp, Octave 5) is written 4", trad, pt(spelling.BSharp(), 5), 4},
			{"S19: (B sharp, Octave 5) is written 5", func(testing.TB) *system.System { return system.Positions19() }, pt(spelling.BSharp(), 5), 5},
			{"T: (B with 13 sharps, Octave 6) is written 4", trad, pt(spi(letter.B, 13), 6), 4},
			{"SA: (A flat, Octave 3) is written 4", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, pt(spelling.AFlat(), 3), 4},
			{"T: (C, Octave 4) is written 4", trad, pt(spelling.C(), 4), 4},
			{"T: (C flat, Octave MIN) is written MIN + 1", trad, pt(spelling.CFlat(), math.MinInt), math.MinInt + 1},
		}
		for _, tt := range tests {
			s := tt.build(t)
			got, err := s.WrittenOctave(tt.p)
			require.NoError(t, err, tt.name)
			assert.Equal(t, o(tt.want), got, tt.name)
			assert.Equal(t, o(tt.want), s.MustWrittenOctave(tt.p), tt.name)
		}
	})

	t.Run("FromWritten (K-6)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			build   func(t testing.TB) *system.System
			s       spelling.Spelling
			written int
			octave  int
		}{
			{"T: B sharp 3 is Octave 4", trad, spelling.BSharp(), 3, 4},
			{"T: C flat 4 is Octave 3", trad, spelling.CFlat(), 4, 3},
			{"T: B with 13 sharps, written 4, is Octave 6", trad, spi(letter.B, 13), 4, 6},
			{"T: C4 is Octave 4", trad, spelling.C(), 4, 4},
			{"SA: A flat 4 is Octave 3 (K-25)", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, spelling.AFlat(), 4, 3},
			{"SA: G sharp 3 is Octave 3 (K-25)", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, spelling.GSharp(), 3, 3},
			{"S19: B sharp 5 is Octave 5", func(testing.TB) *system.System { return system.Positions19() }, spelling.BSharp(), 5, 5},
			{"T: C flat written MAX is Octave MAX - 1 (K-22)", trad, spelling.CFlat(), math.MaxInt, math.MaxInt - 1},
			{"T: B sharp written MIN is Octave MIN + 1", trad, spelling.BSharp(), math.MinInt, math.MinInt + 1},
		}
		for _, tt := range tests {
			s := tt.build(t)
			got, err := s.FromWritten(tt.s, o(tt.written))
			require.NoError(t, err, tt.name)
			assert.Equal(t, pt(tt.s, tt.octave), got, tt.name)
			assert.Equal(t, got, s.MustFromWritten(tt.s, o(tt.written)), tt.name)
		}
	})
}

func TestSystem_WrittenRoundTrip(t *testing.T) {
	t.Parallel()

	for _, m := range []model{modelT(), modelS19(), modelSA(), modelS24()} {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()

			s := m.build(t)
			names := namesUpTo(2)
			if m.c == 2 {
				names = halfNames(2)
			}
			for _, n := range names {
				k := m.carry(t, n)
				for _, w := range edgeOctaves() {
					// Written register → Octave → written register.
					p, err := s.FromWritten(n, o(w))
					if !shiftFits(w, k) {
						requireErr(t, err, system.ErrOutOfRange, nil)
						assert.Equal(t, pitch.Pitch{}, p)
					} else {
						require.NoError(t, err, "FromWritten(%s, %d)", n, w)
						assert.Equal(t, pt(n, w+k), p, "FromWritten(%s, %d)", n, w)
						back, err := s.WrittenOctave(p)
						require.NoError(t, err)
						assert.Equal(t, o(w), back, "WrittenOctave(FromWritten(%s, %d))", n, w)
					}

					// Octave → written register → Octave; here w is used as the Octave.
					got, err := s.WrittenOctave(pt(n, w))
					if !shiftFits(w, -k) {
						requireErr(t, err, system.ErrOutOfRange, nil)
						assert.Equal(t, octave.Octave{}, got)
					} else {
						require.NoError(t, err, "WrittenOctave(%s, Octave %d)", n, w)
						assert.Equal(t, o(w-k), got)
						assert.Equal(t, pt(n, w), s.MustFromWritten(n, got))
					}
				}
			}
		})
	}
}

func TestSystem_WrittenErrors(t *testing.T) {
	t.Parallel()

	trad := system.Traditional()
	bad := dHalfSharp()

	t.Run("FromWritten", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			s        *system.System
			n        spelling.Spelling
			w        octave.Octave
			category error
		}{
			{"B sharp written MAX (K-22)", trad, spelling.BSharp(), o(math.MaxInt), system.ErrOutOfRange},
			{"C flat written MIN (K-22)", trad, spelling.CFlat(), o(math.MinInt), system.ErrOutOfRange},
			{"inadmissible name (K-17)", trad, bad, o(4), system.ErrInadmissible},
			{"inadmissible name at MAX", trad, bad, o(math.MaxInt), system.ErrInadmissible},
			{"unset octave", trad, spelling.C(), octave.Octave{}, system.ErrInvalid},
			{"inadmissible name and unset octave", trad, bad, octave.Octave{}, system.ErrInvalid},
			{"unset name", trad, spelling.Spelling{}, o(4), system.ErrInvalid},
			{"unset system", nil, spelling.C(), o(4), system.ErrInvalid},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := tt.s.FromWritten(tt.n, tt.w)
				requireErr(t, err, tt.category, nil)
				assert.Equal(t, pitch.Pitch{}, got)
				requirePanicErr(t, tt.category, nil, func() { tt.s.MustFromWritten(tt.n, tt.w) })
			})
		}
	})

	t.Run("WrittenOctave", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			s        *system.System
			p        pitch.Pitch
			category error
		}{
			{"C flat at Octave MAX (K-22)", trad, pt(spelling.CFlat(), math.MaxInt), system.ErrOutOfRange},
			{"B sharp at Octave MIN", trad, pt(spelling.BSharp(), math.MinInt), system.ErrOutOfRange},
			{"inadmissible name (K-17)", trad, pt(bad, 4), system.ErrInadmissible},
			{"unset pitch", trad, pitch.Pitch{}, system.ErrInvalid},
			{"unset system", &system.System{}, pt(spelling.C(), 4), system.ErrInvalid},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := tt.s.WrittenOctave(tt.p)
				requireErr(t, err, tt.category, nil)
				assert.Equal(t, octave.Octave{}, got)
				requirePanicErr(t, tt.category, nil, func() { tt.s.MustWrittenOctave(tt.p) })
			})
		}
	})
}

func TestSystem_ChangeKeepingRegister(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	b3 := pt(spelling.B(), 3)       // written B3
	c4 := pt(spelling.C(), 4)       // written C4
	bs3 := pt(spelling.BSharp(), 4) // written B sharp 3

	tests := []struct {
		name     string
		call     func() (pitch.Pitch, error)
		must     func() pitch.Pitch
		from     pitch.Pitch
		want     pitch.Pitch
		written  int
		distance system.Step
	}{
		{
			name: "B3 raised is written B sharp 3 (K-13)",
			call: func() (pitch.Pitch, error) { return s.RaiseWritten(b3) },
			must: func() pitch.Pitch { return s.MustRaiseWritten(b3) },
			from: b3, want: pt(spelling.BSharp(), 4), written: 3, distance: 1,
		},
		{
			name: "C4 lowered is written C flat 4 (K-13)",
			call: func() (pitch.Pitch, error) { return s.LowerWritten(c4) },
			must: func() pitch.Pitch { return s.MustLowerWritten(c4) },
			from: c4, want: pt(spelling.CFlat(), 3), written: 4, distance: -1,
		},
		{
			name: "B3 raised by 2 is written B double sharp 3 (K-13)",
			call: func() (pitch.Pitch, error) { return s.RaiseWrittenBy(b3, 2) },
			must: func() pitch.Pitch { return s.MustRaiseWrittenBy(b3, 2) },
			from: b3, want: pt(spelling.BDoubleSharp(), 4), written: 3, distance: 2,
		},
		{
			name: "C4 lowered by 1",
			call: func() (pitch.Pitch, error) { return s.LowerWrittenBy(c4, 1) },
			must: func() pitch.Pitch { return s.MustLowerWrittenBy(c4, 1) },
			from: c4, want: pt(spelling.CFlat(), 3), written: 4, distance: -1,
		},
		{
			name: "B3 lowered by -2 is raised by 2",
			call: func() (pitch.Pitch, error) { return s.LowerWrittenBy(b3, -2) },
			must: func() pitch.Pitch { return s.MustLowerWrittenBy(b3, -2) },
			from: b3, want: pt(spelling.BDoubleSharp(), 4), written: 3, distance: 2,
		},
		{
			name: "C4 raised by -1 is lowered",
			call: func() (pitch.Pitch, error) { return s.RaiseWrittenBy(c4, -1) },
			must: func() pitch.Pitch { return s.MustRaiseWrittenBy(c4, -1) },
			from: c4, want: pt(spelling.CFlat(), 3), written: 4, distance: -1,
		},
		{
			name: "raised by 0 is the same pitch",
			call: func() (pitch.Pitch, error) { return s.RaiseWrittenBy(bs3, 0) },
			must: func() pitch.Pitch { return s.MustRaiseWrittenBy(bs3, 0) },
			from: bs3, want: bs3, written: 3, distance: 0,
		},
		{
			name: "B sharp 3 made natural is written B3 (K-13)",
			call: func() (pitch.Pitch, error) { return s.NaturalWritten(bs3) },
			must: func() pitch.Pitch { return s.MustNaturalWritten(bs3) },
			from: bs3, want: b3, written: 3, distance: -1,
		},
		{
			name: "C flat 4 made natural is written C4",
			call: func() (pitch.Pitch, error) { return s.NaturalWritten(pt(spelling.CFlat(), 3)) },
			must: func() pitch.Pitch { return s.MustNaturalWritten(pt(spelling.CFlat(), 3)) },
			from: pt(spelling.CFlat(), 3), want: c4, written: 4, distance: 1,
		},
		{
			name: "C4 respelled as B sharp is written B sharp 4 (K-13)",
			call: func() (pitch.Pitch, error) { return s.RespellWritten(c4, spelling.BSharp()) },
			must: func() pitch.Pitch { return s.MustRespellWritten(c4, spelling.BSharp()) },
			from: c4, want: pt(spelling.BSharp(), 5), written: 4, distance: 12,
		},
		{
			name: "B sharp 3 respelled as C is written C3",
			call: func() (pitch.Pitch, error) { return s.RespellWritten(bs3, spelling.C()) },
			must: func() pitch.Pitch { return s.MustRespellWritten(bs3, spelling.C()) },
			from: bs3, want: pt(spelling.C(), 3), written: 3, distance: -12,
		},
		{
			name: "C flat at Octave MAX lowered keeps Octave MAX (K-22)",
			call: func() (pitch.Pitch, error) { return s.LowerWritten(pt(spelling.CFlat(), math.MaxInt)) },
			must: func() pitch.Pitch { return s.MustLowerWritten(pt(spelling.CFlat(), math.MaxInt)) },
			from: pt(spelling.CFlat(), math.MaxInt), want: pt(spelling.CDoubleFlat(), math.MaxInt), written: 0, distance: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.must())
			assert.Equal(t, tt.distance, s.MustDistance(tt.from, got))
			if tt.written != 0 {
				assert.Equal(t, o(tt.written), s.MustWrittenOctave(got))
			}
		})
	}

	t.Run("pitch.Raise keeps the Octave, not the register (K-13)", func(t *testing.T) {
		t.Parallel()

		raised := b3.MustRaise()
		assert.Equal(t, pt(spelling.BSharp(), 3), raised)
		assert.Equal(t, o(2), s.MustWrittenOctave(raised), "written B sharp 2")
		assert.Equal(t, system.Step(-11), s.MustDistance(b3, raised), "position C3")
	})
}

func TestSystem_ChangeKeepingRegisterModel(t *testing.T) {
	t.Parallel()

	for _, m := range []model{modelT(), modelS19(), modelSA(), modelS14(), modelS24()} {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()

			s := m.build(t)
			targets := namesUpTo(1)
			for _, p := range samplePitches() {
				w := s.MustWrittenOctave(p)
				l, alt := p.Spelling().Parts()

				for n := -3; n <= 3; n++ {
					r, err := s.RaiseWrittenBy(p, n)
					require.NoError(t, err, "RaiseWrittenBy(%s, %d)", p, n)
					assert.Equal(t, spelling.MustNew(l, alt.MustRaiseBy(n)), r.Spelling())
					assert.Equal(t, w, s.MustWrittenOctave(r), "register of RaiseWrittenBy(%s, %d)", p, n)
					assert.Equal(t, system.Step(n*m.c), s.MustDistance(p, r), "Distance after RaiseWrittenBy(%s, %d)", p, n)
					assert.Equal(t, r, s.MustLowerWrittenBy(p, -n))
				}
				assert.Equal(t, s.MustRaiseWrittenBy(p, 1), s.MustRaiseWritten(p))
				assert.Equal(t, s.MustRaiseWrittenBy(p, -1), s.MustLowerWritten(p))

				nat := s.MustNaturalWritten(p)
				assert.Equal(t, p.Spelling().Natural(), nat.Spelling())
				assert.Equal(t, w, s.MustWrittenOctave(nat))

				for _, target := range targets {
					r, err := s.RespellWritten(p, target)
					require.NoError(t, err)
					assert.Equal(t, target, r.Spelling())
					assert.Equal(t, w, s.MustWrittenOctave(r), "register of RespellWritten(%s, %s)", p, target)
					assert.Equal(t, system.Step(m.shift(t, target)-m.shift(t, p.Spelling())), s.MustDistance(p, r),
						"Distance after RespellWritten(%s, %s)", p, target)
				}
			}
		})
	}
}

func TestSystem_ChangeKeepingRegisterErrors(t *testing.T) {
	t.Parallel()

	trad := system.Traditional()
	limited := newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2)))
	bad := pt(dHalfSharp(), 4)
	c4 := pt(spelling.C(), 4)
	var unsetP pitch.Pitch

	type op struct {
		call func() (pitch.Pitch, error)
		must func()
	}
	raiseBy := func(s *system.System, p pitch.Pitch, n int) op {
		return op{func() (pitch.Pitch, error) { return s.RaiseWrittenBy(p, n) }, func() { s.MustRaiseWrittenBy(p, n) }}
	}
	lowerBy := func(s *system.System, p pitch.Pitch, n int) op {
		return op{func() (pitch.Pitch, error) { return s.LowerWrittenBy(p, n) }, func() { s.MustLowerWrittenBy(p, n) }}
	}
	raise := func(s *system.System, p pitch.Pitch) op {
		return op{func() (pitch.Pitch, error) { return s.RaiseWritten(p) }, func() { s.MustRaiseWritten(p) }}
	}
	lowerOp := func(s *system.System, p pitch.Pitch) op {
		return op{func() (pitch.Pitch, error) { return s.LowerWritten(p) }, func() { s.MustLowerWritten(p) }}
	}
	respell := func(s *system.System, p pitch.Pitch, n spelling.Spelling) op {
		return op{func() (pitch.Pitch, error) { return s.RespellWritten(p, n) }, func() { s.MustRespellWritten(p, n) }}
	}
	natural := func(s *system.System, p pitch.Pitch) op {
		return op{func() (pitch.Pitch, error) { return s.NaturalWritten(p) }, func() { s.MustNaturalWritten(p) }}
	}

	tests := []struct {
		name     string
		op       op
		category error
	}{
		// E-1.
		{"RaiseWritten: unset system", raise(nil, c4), system.ErrInvalid},
		{"RaiseWritten: unset pitch", raise(trad, unsetP), system.ErrInvalid},
		{"LowerWritten: unset pitch", lowerOp(trad, unsetP), system.ErrInvalid},
		{"RaiseWrittenBy: unset pitch", raiseBy(trad, unsetP, math.MaxInt), system.ErrInvalid},
		{"LowerWrittenBy: unset system", lowerBy(&system.System{}, c4, 1), system.ErrInvalid},
		{"RespellWritten: unset name", respell(trad, c4, spelling.Spelling{}), system.ErrInvalid},
		{"RespellWritten: unset pitch and inadmissible name", respell(trad, unsetP, dHalfSharp()), system.ErrInvalid},
		{"RespellWritten: inadmissible pitch and unset name", respell(trad, bad, spelling.Spelling{}), system.ErrInvalid},
		{"NaturalWritten: unset pitch", natural(trad, unsetP), system.ErrInvalid},

		// E-6 for the input name before E-9.
		{"RaiseWritten: inadmissible pitch (K-17)", raise(trad, bad), system.ErrInadmissible},
		{"RaiseWrittenBy: inadmissible pitch before out of range", raiseBy(trad, bad, math.MaxInt), system.ErrInadmissible},
		{"LowerWrittenBy: inadmissible pitch", lowerBy(trad, bad, 1), system.ErrInadmissible},
		{"NaturalWritten: inadmissible pitch", natural(trad, bad), system.ErrInadmissible},
		{"RespellWritten: inadmissible pitch", respell(trad, bad, spelling.C()), system.ErrInadmissible},

		// E-9 for the alteration value of the result.
		{"RaiseWrittenBy MaxInt", raiseBy(trad, c4, math.MaxInt), system.ErrOutOfRange},
		{"RaiseWrittenBy MinInt", raiseBy(trad, c4, math.MinInt), system.ErrOutOfRange},
		{"LowerWrittenBy MaxInt", lowerBy(trad, c4, math.MaxInt), system.ErrOutOfRange},
		{"LowerWrittenBy MinInt", lowerBy(trad, c4, math.MinInt), system.ErrOutOfRange},
		{"RaiseWrittenBy M+1", raiseBy(trad, c4, maxAlt+1), system.ErrOutOfRange},
		{"RaiseWritten from alteration M (K-21)", raise(trad, pt(spi(letter.C, maxAlt), 4)), system.ErrOutOfRange},
		{"LowerWritten from alteration -M", lowerOp(trad, pt(spi(letter.C, -maxAlt), 4)), system.ErrOutOfRange},
		{"out of range before admissibility of the result", raiseBy(limited, pt(spelling.CDoubleSharp(), 4), math.MaxInt), system.ErrOutOfRange},

		// E-6 for the result name.
		{"RaiseWritten: C double sharp with A from -2 to 2 (K-21)", raise(limited, pt(spelling.CDoubleSharp(), 4)), system.ErrInadmissible},
		{"LowerWrittenBy: C double flat by 1 with A from -2 to 2", lowerBy(limited, pt(spelling.CDoubleFlat(), 4), 1), system.ErrInadmissible},
		{"RespellWritten: inadmissible name (K-17)", respell(trad, c4, dHalfSharp()), system.ErrInadmissible},
		{"result name before the octave range", raise(limited, pt(spelling.BDoubleSharp(), math.MaxInt)), system.ErrInadmissible},

		// E-9 for the Octave of the result.
		{"RaiseWritten: B at Octave MAX", raise(trad, pt(spelling.B(), math.MaxInt)), system.ErrOutOfRange},
		{"LowerWritten: C at Octave MIN", lowerOp(trad, pt(spelling.C(), math.MinInt)), system.ErrOutOfRange},
		{"NaturalWritten: B sharp at Octave MIN (K-22)", natural(trad, pt(spelling.BSharp(), math.MinInt)), system.ErrOutOfRange},
		{"NaturalWritten: C flat at Octave MAX", natural(trad, pt(spelling.CFlat(), math.MaxInt)), system.ErrOutOfRange},
		{"RespellWritten: C at Octave MAX as B sharp", respell(trad, pt(spelling.C(), math.MaxInt), spelling.BSharp()), system.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.op.call()
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, pitch.Pitch{}, got)
			requirePanicErr(t, tt.category, nil, tt.op.must)
		})
	}
}
