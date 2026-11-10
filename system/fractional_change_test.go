package system_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for changes by an alteration value d (step Sh-3): OP-20 and OP-36 with d,
// musical-system.op-7, contract 5.7 and 5.9, criteria K-5 (S24), K-13 (S24).

func TestShiftedClassByAlteration(t *testing.T) {
	t.Parallel()

	s24 := system.QuarterTones24()

	t.Run("examples (K-5)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			s     *system.System
			input spelling.Spelling
			d     alteration.Alteration
			want  pitchclass.Class
		}{
			{"S24: D, d = 1/2 (K-5)", s24, spelling.D(), q(1, 2), cls(5, 24)},
			{"S24: C, d = -1/2", s24, spelling.C(), q(-1, 2), cls(23, 24)},
			{"S24: E, d = 3/2", s24, spelling.E(), q(3, 2), cls(11, 24)},
			{"S24: D+1/2, d = -1/2", s24, dHalfSharp(), q(-1, 2), cls(4, 24)},
			{"S19: C, d = 2 (K-5)", system.Positions19(), spelling.C(), a(2), cls(2, 19)},
			{"T: B sharp, d = 0", system.Traditional(), spelling.BSharp(), a(0), cls(0, 12)},
			{"T: C, d = M", system.Traditional(), spelling.C(), a(maxAlt), cls(4, 12)},
		}
		for _, tt := range tests {
			got, err := tt.s.ShiftedClassByAlteration(tt.input, tt.d)
			require.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.want, tt.s.MustShiftedClassByAlteration(tt.input, tt.d), tt.name)
		}
	})

	t.Run("whole d agrees with ShiftedClass", func(t *testing.T) {
		t.Parallel()

		for _, m := range []model{modelT(), modelS19(), modelS24(), modelSA(), modelS5(), modelS1()} {
			s := m.build(t)
			for _, n := range namesUpTo(2) {
				for d := -5; d <= 5; d++ {
					assert.Equal(t, s.MustShiftedClass(n, d), s.MustShiftedClassByAlteration(n, a(d)), "%s: %s, d = %d", m.name, n, d)
				}
			}
		}
	})

	t.Run("model: (κ + d·c) mod N for halves in S24", func(t *testing.T) {
		t.Parallel()

		m := modelS24()
		for _, n := range halfNames(2) {
			for num := -6; num <= 6; num++ {
				want := m.class(t, n) + num*m.c/2
				want -= floorDiv(want, m.n) * m.n
				assert.Equal(t, cls(want, m.n), s24.MustShiftedClassByAlteration(n, q(num, 2)), "%s, d = %d/2", n, num)
			}
		}
	})

	limited := newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2)))

	tests := []struct {
		name     string
		s        *system.System
		input    spelling.Spelling
		d        alteration.Alteration
		category error
	}{
		{"unset system", nil, spelling.C(), q(1, 2), system.ErrInvalid},
		{"zero system", &system.System{}, dHalfSharp(), a(maxAlt), system.ErrInvalid},
		{"unset name", s24, spelling.Spelling{}, q(1, 2), system.ErrInvalid},
		{"inadmissible source", system.Traditional(), dHalfSharp(), q(1, 2), system.ErrInadmissible},
		{"inadmissible source before range", system.Traditional(), dHalfSharp(), a(maxAlt), system.ErrInadmissible},
		{"a + d above the range", system.Traditional(), spi(letter.C, maxAlt), a(1), system.ErrOutOfRange},
		{"a + d above the range by a half", s24, spi(letter.C, maxAlt), q(1, 2), system.ErrOutOfRange},
		{"a + d below the range", system.Traditional(), spi(letter.C, -maxAlt), a(-1), system.ErrOutOfRange},
		{"range before admissibility of a + d", system.Traditional(), spi(letter.C, maxAlt), q(1, 2), system.ErrOutOfRange},
		{"d·c not whole: a + d inadmissible", system.Traditional(), spelling.C(), q(1, 2), system.ErrInadmissible},
		{"a + d outside A", limited, spelling.CDoubleSharp(), a(1), system.ErrInadmissible},
		{"a third in S24", s24, spelling.C(), q(1, 3), system.ErrInadmissible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.s.ShiftedClassByAlteration(tt.input, tt.d)
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, pitchclass.Class{}, got)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustShiftedClassByAlteration(tt.input, tt.d) })
		})
	}
}

func TestChangeWrittenByAlteration(t *testing.T) {
	t.Parallel()

	s24 := system.QuarterTones24()

	t.Run("examples (K-13)", func(t *testing.T) {
		t.Parallel()

		d4 := pt(spelling.D(), 4)
		b4 := pt(spelling.B(), 4)
		c4 := pt(spelling.C(), 4)

		tests := []struct {
			name     string
			s        *system.System
			from     pitch.Pitch
			raise    bool
			d        alteration.Alteration
			want     pitch.Pitch
			distance system.Step
		}{
			{"S24: D4 raised by 1/2 is written D+1/2 4 (K-13)", s24, d4, true, q(1, 2), pt(dHalfSharp(), 4), 1},
			{"S24: B4 raised by 1/2", s24, b4, true, q(1, 2), pt(spq(letter.B, 1, 2), 4), 1},
			{"S24: B4 raised by 1 is written B sharp 4", s24, b4, true, a(1), pt(spelling.BSharp(), 5), 2},
			{"S24: C4 lowered by 1/2 is written C-1/2 4", s24, c4, false, q(1, 2), pt(spq(letter.C, -1, 2), 3), -1},
			{"S24: C4 raised by -1/2", s24, c4, true, q(-1, 2), pt(spq(letter.C, -1, 2), 3), -1},
			{"S24: D+1/2 4 lowered by -3/2", s24, pt(dHalfSharp(), 4), false, q(-3, 2), pt(spi(letter.D, 2), 4), 3},
			{"S24: d = 0 is the same pitch", s24, b4, true, a(0), b4, 0},
			{"T: B3 raised by 1 is written B sharp 3", system.Traditional(), pt(spelling.B(), 3), true, a(1), pt(spelling.BSharp(), 4), 1},
			{"S24: B at Octave MAX raised by 1/2 stays", s24, pt(spelling.B(), math.MaxInt), true, q(1, 2), pt(spq(letter.B, 1, 2), math.MaxInt), 1},
		}
		for _, tt := range tests {
			var (
				got pitch.Pitch
				err error
				mv  pitch.Pitch
			)
			if tt.raise {
				got, err = tt.s.RaiseWrittenByAlteration(tt.from, tt.d)
				mv = tt.s.MustRaiseWrittenByAlteration(tt.from, tt.d)
			} else {
				got, err = tt.s.LowerWrittenByAlteration(tt.from, tt.d)
				mv = tt.s.MustLowerWrittenByAlteration(tt.from, tt.d)
			}
			require.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.want, mv, tt.name)
			assert.Equal(t, tt.distance, tt.s.MustDistance(tt.from, got), tt.name)
			assert.Equal(t, tt.s.MustWrittenOctave(tt.from), tt.s.MustWrittenOctave(got), "%s: register kept", tt.name)
		}
	})

	t.Run("whole d agrees with RaiseWrittenBy and LowerWrittenBy", func(t *testing.T) {
		t.Parallel()

		for _, m := range []model{modelT(), modelS19(), modelS24(), modelSA(), modelS5(), modelS1()} {
			s := m.build(t)
			sample := samplePitches()
			for i := 0; i < len(sample); i += 3 {
				p := sample[i]
				for d := -3; d <= 3; d++ {
					assert.Equal(t, s.MustRaiseWrittenBy(p, d), s.MustRaiseWrittenByAlteration(p, a(d)), "%s: raise %s by %d", m.name, p, d)
					assert.Equal(t, s.MustLowerWrittenBy(p, d), s.MustLowerWrittenByAlteration(p, a(d)), "%s: lower %s by %d", m.name, p, d)
				}
			}
		}
	})

	t.Run("model for halves in S24: register kept, distance d·c", func(t *testing.T) {
		t.Parallel()

		for _, p := range samplePitches() {
			w := s24.MustWrittenOctave(p)
			l, alt := p.Spelling().Parts()
			for num := -4; num <= 4; num++ {
				d := q(num, 2)
				r := s24.MustRaiseWrittenByAlteration(p, d)
				assert.Equal(t, spelling.MustNew(l, alt.MustAdd(d)), r.Spelling())
				assert.Equal(t, w, s24.MustWrittenOctave(r), "register of %s raised by %s", p, d)
				assert.Equal(t, system.Step(num), s24.MustDistance(p, r), "distance of %s raised by %s", p, d)
				assert.Equal(t, r, s24.MustLowerWrittenByAlteration(p, q(-num, 2)))
			}
		}
	})

	trad := system.Traditional()
	limited := newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2)))
	c4 := pt(spelling.C(), 4)
	bad := pt(dHalfSharp(), 4)

	type op struct {
		call func() (pitch.Pitch, error)
		must func()
	}
	raise := func(s *system.System, p pitch.Pitch, d alteration.Alteration) op {
		return op{
			func() (pitch.Pitch, error) { return s.RaiseWrittenByAlteration(p, d) },
			func() { s.MustRaiseWrittenByAlteration(p, d) },
		}
	}
	lower := func(s *system.System, p pitch.Pitch, d alteration.Alteration) op {
		return op{
			func() (pitch.Pitch, error) { return s.LowerWrittenByAlteration(p, d) },
			func() { s.MustLowerWrittenByAlteration(p, d) },
		}
	}

	tests := []struct {
		name     string
		op       op
		category error
	}{
		// E-1.
		{"raise: unset system", raise(nil, c4, q(1, 2)), system.ErrInvalid},
		{"lower: zero system", lower(&system.System{}, c4, q(1, 2)), system.ErrInvalid},
		{"raise: unset pitch", raise(s24, pitch.Pitch{}, q(1, 2)), system.ErrInvalid},
		{"lower: unset pitch", lower(s24, pitch.Pitch{}, a(maxAlt)), system.ErrInvalid},

		// E-6 for the input name.
		{"raise: inadmissible pitch", raise(trad, bad, q(1, 2)), system.ErrInadmissible},
		{"lower: inadmissible pitch before out of range", lower(trad, bad, a(maxAlt)), system.ErrInadmissible},

		// E-9 for the alteration value of the result.
		{"raise: above the range", raise(trad, pt(spi(letter.C, maxAlt), 4), a(1)), system.ErrOutOfRange},
		{"lower: below the range", lower(trad, pt(spi(letter.C, -maxAlt), 4), a(1)), system.ErrOutOfRange},
		{"raise: above the range by a half", raise(s24, pt(spi(letter.C, maxAlt), 4), q(1, 2)), system.ErrOutOfRange},
		{"lower: by -M from M", lower(trad, pt(spi(letter.C, maxAlt), 4), a(-maxAlt)), system.ErrOutOfRange},
		{"range before admissibility of the result", raise(trad, pt(spi(letter.C, maxAlt), 4), q(1, 2)), system.ErrOutOfRange},

		// E-6 for the result name.
		{"raise: d·c not whole", raise(trad, c4, q(1, 2)), system.ErrInadmissible},
		{"raise: outside A", raise(limited, pt(spelling.CDoubleSharp(), 4), a(1)), system.ErrInadmissible},
		{"raise: a third in S24", raise(s24, c4, q(1, 3)), system.ErrInadmissible},
		{"result name before the octave range", raise(trad, pt(spelling.B(), math.MaxInt), q(1, 2)), system.ErrInadmissible},

		// E-9 for the Octave of the result.
		{"raise: B at Octave MAX by 1", raise(trad, pt(spelling.B(), math.MaxInt), a(1)), system.ErrOutOfRange},
		{"raise: B at Octave MAX by 1 in S24", raise(s24, pt(spelling.B(), math.MaxInt), a(1)), system.ErrOutOfRange},
		{"lower: C at Octave MIN by 1/2 in S24", lower(s24, pt(spelling.C(), math.MinInt), q(1, 2)), system.ErrOutOfRange},
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

func TestUnsetSystem_ByAlteration(t *testing.T) {
	t.Parallel()

	for _, s := range []*system.System{nil, {}} {
		_, err := s.ShiftedClassByAlteration(spelling.C(), q(1, 2))
		requireErr(t, err, system.ErrInvalid, nil)
		_, err = s.RaiseWrittenByAlteration(pt(spelling.C(), 4), q(1, 2))
		requireErr(t, err, system.ErrInvalid, nil)
		_, err = s.LowerWrittenByAlteration(pt(spelling.C(), 4), q(1, 2))
		requireErr(t, err, system.ErrInvalid, nil)
	}
}
