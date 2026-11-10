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

// Tests for a letter on the boundary of the next octave, p(l) = N (step Sh-3):
// P-2, ODV-6, contract 5.6 (cause 2), position.inv-12 and inv-13, criteria K-2 (S5, S1),
// K-6, K-8, K-9 (S5), K-11 (S5), K-23, K-24, K-28 (S5).

// defS5 is S5 of the spec: C0 D1 E2 F2 G3 A4 B5, N = 5, c = 1, A = Z; B is on N.
func defS5() system.Definition {
	return system.Definition{
		Name: "S5", Classes: 5, Boundary: letter.C,
		Offsets: offs(0, 1, 2, 2, 3, 4, 5), Chromatic: 1, Alterations: system.Integers(),
	}
}

// defS1 is S1 of the spec: N = 1, C0, the other letters on 1 = N, c = 1, A = Z.
func defS1() system.Definition {
	return system.Definition{
		Name: "S1", Classes: 1, Boundary: letter.C,
		Offsets: offs(0, 1, 1, 1, 1, 1, 1), Chromatic: 1, Alterations: system.Integers(),
	}
}

func modelS5() model {
	d := defS5()
	return model{
		name: "S5", n: d.Classes, offsets: d.Offsets, c: d.Chromatic,
		build: func(t testing.TB) *system.System { return newSystem(t, defS5()) },
	}
}

func modelS1() model {
	d := defS1()
	return model{
		name: "S1", n: d.Classes, offsets: d.Offsets, c: d.Chromatic,
		build: func(t testing.TB) *system.System { return newSystem(t, defS1()) },
	}
}

func boundaryModels() []model { return []model{modelS5(), modelS1()} }

func TestBoundaryLetter_Definition(t *testing.T) {
	t.Parallel()

	t.Run("an offset equal to N is accepted", func(t *testing.T) {
		t.Parallel()

		for _, def := range []system.Definition{
			defS5(),
			defS1(),
			{Classes: 12, Boundary: letter.C, Offsets: offs(0, 2, 4, 5, 7, 9, 12), Chromatic: 1, Alterations: system.Integers()},
			{Classes: 12, Boundary: letter.A, Offsets: offs(3, 5, 7, 8, 12, 0, 2), Chromatic: 1, Alterations: system.Integers()},
		} {
			s, err := system.New(def)
			require.NoError(t, err, "%+v", def)
			assert.True(t, s.IsValid())
			assert.Equal(t, def, s.Definition())
		}
	})

	tests := []struct {
		name  string
		def   func() system.Definition
		cause error
	}{
		{"S5 with B on N + 1", func() system.Definition {
			d := defS5()
			d.Offsets[6] = 6
			return d
		}, system.ErrOffsets},
		{"S1 with D on N + 1", func() system.Definition {
			d := defS1()
			d.Offsets[1] = 2
			return d
		}, system.ErrOffsets},
		{"boundary letter on N", func() system.Definition {
			d := defS5()
			d.Offsets[0] = 5
			return d
		}, system.ErrOffsets},
		{"G on N followed by A on 4 goes down", func() system.Definition {
			d := defS5()
			d.Offsets[4] = 5
			return d
		}, system.ErrDescending},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := system.New(tt.def())
			requireErr(t, err, system.ErrDefinition, tt.cause)
			assert.Nil(t, s)
		})
	}

	t.Run("FromSteps and NewRegular give S5 (K-2)", func(t *testing.T) {
		t.Parallel()

		s5 := newSystem(t, defS5())
		steps := [letter.Count]int{1, 1, 0, 1, 1, 1, 0}

		fs, err := system.FromSteps(letter.C, steps, 1, system.Integers(), "five")
		require.NoError(t, err)
		requireDefinition(t, system.Definition{Name: "five", Classes: 5, Boundary: letter.C, Offsets: offs(0, 1, 2, 2, 3, 4, 5), Chromatic: 1, Alterations: system.Integers()}, fs.Definition())
		assert.True(t, fs.Equal(s5))

		reg, err := system.NewRegular(1, 0, letter.C, system.Integers(), "")
		require.NoError(t, err)
		assert.True(t, reg.Equal(s5))
	})
}

func TestBoundaryLetter_Readers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		def        system.Definition
		steps      [letter.Count]int
		natural    [letter.Count]int
		regularity system.Regularity
		layout     bool
		str        string
	}{
		{
			name:       "S5 (K-2)",
			def:        defS5(),
			steps:      [letter.Count]int{1, 1, 0, 1, 1, 1, 0},
			natural:    [letter.Count]int{0, 1, 2, 2, 3, 4, 0},
			regularity: system.Regularity{Regular: true, Large: 1, Small: 0, RegularByClasses: true},
			str:        `System("S5"; N=5; b=C; C=0 D=1 E=2 F=2 G=3 A=4 B=5; c=1; A=0 + Z/1 in [-inf, +inf])`,
		},
		{
			name:       "S1 (K-2, K-23)",
			def:        defS1(),
			steps:      [letter.Count]int{1, 0, 0, 0, 0, 0, 0},
			natural:    [letter.Count]int{0, 0, 0, 0, 0, 0, 0},
			regularity: system.Regularity{RegularByClasses: true},
			str:        `System("S1"; N=1; b=C; C=0 D=1 E=1 F=1 G=1 A=1 B=1; c=1; A=0 + Z/1 in [-inf, +inf])`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := newSystem(t, tt.def)
			assert.Equal(t, tt.steps, s.Steps())
			for _, l := range letter.All() {
				assert.Equal(t, tt.def.Offsets[l.Ordinal()], s.MustOffset(l), "Offset(%s)", l)
				assert.Equal(t, tt.steps[l.Ordinal()], s.MustStepAfter(l), "StepAfter(%s)", l)
				assert.Equal(t, cls(tt.natural[l.Ordinal()], tt.def.Classes), s.MustNaturalClass(l), "NaturalClass(%s)", l)
			}
			assert.Equal(t, tt.regularity, s.MustRegularity())
			assert.Equal(t, tt.layout, s.MustHasTraditionalLayout())
			assert.Equal(t, tt.str, s.String())
		})
	}
}

func TestBoundaryLetter_Names(t *testing.T) {
	t.Parallel()

	s5 := newSystem(t, defS5())
	s1 := newSystem(t, defS1())

	t.Run("S5 examples (K-9, K-11, K-24, K-28)", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, cls(0, 5), s5.MustClass(spelling.B()), "K-24: B has class 0")
		assert.Equal(t, 1, s5.MustCarry(spelling.B()), "K-28: B carries 1")
		assert.Equal(t, 0, s5.MustCarry(spelling.BFlat()))
		assert.Equal(t, 1, s5.MustCarry(spelling.BSharp()))
		assert.Equal(t, -1, s5.MustCarry(spelling.CFlat()))
		assert.True(t, s5.MustEquivalentNames(spelling.E(), spelling.F()), "K-9: E ≡ F")
		assert.True(t, s5.MustEquivalentNames(spelling.B(), spelling.C()))
		assert.Equal(t, spelling.E(), s5.MustSpell(cls(2, 5), system.Sharps()), "K-11: class 2 → E")
		assert.Equal(t, spelling.E(), s5.MustSpell(cls(2, 5), system.Flats()))
		assert.Equal(t, spelling.C(), s5.MustSpell(cls(0, 5), system.Sharps()), "C before B on the circle from C")
		assert.Equal(t, []spelling.Spelling{spelling.C(), spelling.B()}, collect(t, s5.MustSpellings(cls(0, 5), a(0))))
		assert.Equal(t, []spelling.Spelling{spelling.E(), spelling.F()}, collect(t, s5.MustSpellings(cls(2, 5), a(0))))
	})

	t.Run("S1: every name has class 0 (K-23)", func(t *testing.T) {
		t.Parallel()

		for _, n := range namesUpTo(3) {
			assert.Equal(t, cls(0, 1), s1.MustClass(n), "Class(%s)", n)
		}
		assert.Equal(t, 1, s1.MustCarry(spelling.D()))
		assert.Equal(t, 0, s1.MustCarry(spelling.C()))
		assert.Equal(t, spelling.C(), s1.MustSpell(cls(0, 1), system.Sharps()))
	})

	t.Run("model: class, carry, shifted class", func(t *testing.T) {
		t.Parallel()

		for _, m := range boundaryModels() {
			s := m.build(t)
			for _, n := range namesUpTo(3) {
				assert.Equal(t, cls(m.class(t, n), m.n), s.MustClass(n), "%s: Class(%s)", m.name, n)
				assert.Equal(t, m.carry(t, n), s.MustCarry(n), "%s: Carry(%s)", m.name, n)
				for shift := -7; shift <= 7; shift++ {
					want := m.class(t, n) + shift*m.c
					want -= floorDiv(want, m.n) * m.n
					assert.Equal(t, cls(want, m.n), s.MustShiftedClass(n, shift))
				}
			}
		}
	})

	t.Run("model: spellings and sharps, flats", func(t *testing.T) {
		t.Parallel()

		for _, m := range boundaryModels() {
			s := m.build(t)
			values := collect(t, s.Alterations().MustValues(a(2)))
			for c := range m.n {
				var want []spelling.Spelling
				for _, l := range letter.All() {
					for _, v := range values {
						n := spelling.MustNew(l, v)
						if m.class(t, n) == c {
							want = append(want, n)
						}
					}
				}
				got := collect(t, s.MustSpellings(cls(c, m.n), a(2)))
				require.NotEmpty(t, want)
				assert.Equal(t, want, got, "%s class %d", m.name, c)

				assert.Equal(t, bestSpelling(want, s.Boundary(), true), s.MustSpell(cls(c, m.n), system.Sharps()), "%s class %d sharps", m.name, c)
				assert.Equal(t, bestSpelling(want, s.Boundary(), false), s.MustSpell(cls(c, m.n), system.Flats()), "%s class %d flats", m.name, c)
			}
		}
	})
}

// bestSpelling applies the rules "sharps" and "flats" of spec 2.5 to candidates.
func bestSpelling(candidates []spelling.Spelling, b letter.Letter, preferSharp bool) spelling.Spelling {
	abs := func(v alteration.Alteration) alteration.Alteration {
		if v.Sign() < 0 {
			return v.MustMul(-1)
		}
		return v
	}
	better := func(x, y spelling.Spelling) bool {
		ax, ay := x.Alteration(), y.Alteration()
		if d := abs(ax).Compare(abs(ay)); d != 0 {
			return d < 0
		}
		if ax.Sign() != ay.Sign() {
			if preferSharp {
				return ax.Sign() > ay.Sign()
			}
			return ax.Sign() < ay.Sign()
		}
		return b.DistanceUp(x.Letter()) < b.DistanceUp(y.Letter())
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if better(c, best) {
			best = c
		}
	}
	return best
}

func TestBoundaryLetter_Written(t *testing.T) {
	t.Parallel()

	s5 := newSystem(t, defS5())
	s1 := newSystem(t, defS1())

	t.Run("examples (K-6, K-23, K-24)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			s       *system.System
			n       spelling.Spelling
			written int
			octave  int
		}{
			{"S5: B written 4 is Octave 5 (K-6, K-24)", s5, spelling.B(), 4, 5},
			{"S5: C flat written 5 is Octave 4 (K-6)", s5, spelling.CFlat(), 5, 4},
			{"S5: B flat written 4 is Octave 4", s5, spelling.BFlat(), 4, 4},
			{"S5: B sharp written 4 is Octave 5", s5, spelling.BSharp(), 4, 5},
			{"S5: B written MAX - 1 is Octave MAX", s5, spelling.B(), math.MaxInt - 1, math.MaxInt},
			{"S1: D written 4 is Octave 5 (K-23)", s1, spelling.D(), 4, 5},
			{"S1: C written 4 is Octave 4", s1, spelling.C(), 4, 4},
		}
		for _, tt := range tests {
			got, err := tt.s.FromWritten(tt.n, o(tt.written))
			require.NoError(t, err, tt.name)
			assert.Equal(t, pt(tt.n, tt.octave), got, tt.name)
			assert.Equal(t, o(tt.written), tt.s.MustWrittenOctave(got), tt.name)
		}

		_, err := s5.FromWritten(spelling.B(), o(math.MaxInt))
		requireErr(t, err, system.ErrOutOfRange, nil)
		_, err = s5.WrittenOctave(pt(spelling.B(), math.MinInt))
		requireErr(t, err, system.ErrOutOfRange, nil)
	})

	t.Run("round trip for all names with |a| <= 2", func(t *testing.T) {
		t.Parallel()

		for _, m := range boundaryModels() {
			s := m.build(t)
			for _, n := range namesUpTo(2) {
				k := m.carry(t, n)
				for _, w := range edgeOctaves() {
					p, err := s.FromWritten(n, o(w))
					if !shiftFits(w, k) {
						requireErr(t, err, system.ErrOutOfRange, nil)
					} else {
						require.NoError(t, err, "%s: FromWritten(%s, %d)", m.name, n, w)
						assert.Equal(t, pt(n, w+k), p)
						assert.Equal(t, o(w), s.MustWrittenOctave(p))
					}

					got, err := s.WrittenOctave(pt(n, w))
					if !shiftFits(w, -k) {
						requireErr(t, err, system.ErrOutOfRange, nil)
					} else {
						require.NoError(t, err)
						assert.Equal(t, o(w-k), got)
						assert.Equal(t, pt(n, w), s.MustFromWritten(n, got))
					}
				}
			}
		}
	})

	t.Run("changes keeping the register", func(t *testing.T) {
		t.Parallel()

		b4 := pt(spelling.B(), 5) // written B4
		c5 := pt(spelling.C(), 5) // written C5

		r := s5.MustRaiseWritten(b4)
		assert.Equal(t, pt(spelling.BSharp(), 5), r, "B sharp 4")
		assert.Equal(t, o(4), s5.MustWrittenOctave(r))
		assert.Equal(t, system.Step(1), s5.MustDistance(b4, r))

		l := s5.MustLowerWritten(b4)
		assert.Equal(t, pt(spelling.BFlat(), 4), l, "B flat 4")
		assert.Equal(t, system.Step(-1), s5.MustDistance(b4, l))

		nat := s5.MustNaturalWritten(pt(spelling.CFlat(), 4)) // written C flat 5
		assert.Equal(t, c5, nat)

		re := s5.MustRespellWritten(c5, spelling.B())
		assert.Equal(t, pt(spelling.B(), 6), re, "written B5")
		assert.Equal(t, system.Step(5), s5.MustDistance(c5, re))
	})
}

func TestBoundaryLetter_Positions(t *testing.T) {
	t.Parallel()

	s5 := newSystem(t, defS5())
	b4 := pt(spelling.B(), 5)         // written B4
	c5 := pt(spelling.C(), 5)         // written C5
	a4 := pt(spelling.A(), 4)         // written A4
	bFlat4 := pt(spelling.BFlat(), 4) // written B flat 4
	bSharp4 := pt(spelling.BSharp(), 5)
	d5 := pt(spelling.D(), 5)
	cFlat5 := pt(spelling.CFlat(), 4) // written C flat 5

	t.Run("distance (K-8)", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, system.Step(0), s5.MustDistance(b4, c5))
		assert.Equal(t, system.Step(0), s5.MustDistance(a4, bFlat4))
		assert.Equal(t, system.Step(1), s5.MustDistance(a4, b4))
		assert.Equal(t, system.Step(5), s5.MustDistance(c5, pt(spelling.B(), 6)))
	})

	t.Run("equivalence (K-9)", func(t *testing.T) {
		t.Parallel()

		assert.True(t, s5.MustEquivalent(b4, c5))
		assert.True(t, s5.MustEquivalent(bSharp4, d5))
		assert.True(t, s5.MustEquivalent(cFlat5, a4))
		assert.True(t, s5.MustEquivalent(a4, bFlat4))
		assert.True(t, s5.MustEquivalent(cFlat5, bFlat4))
		assert.False(t, s5.MustEquivalent(b4, a4))
		assert.Equal(t, 0, s5.MustCompare(b4, c5))
		assert.Equal(t, 1, s5.MustCompare(b4, a4))
	})

	t.Run("enharmonic pitches keep the Octave", func(t *testing.T) {
		t.Parallel()

		got := collect(t, s5.MustEnharmonicPitches(c5, a(0)))
		assert.Equal(t, []pitch.Pitch{c5, b4}, got)
	})

	t.Run("move and walk across the boundary", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, c5, s5.MustMove(b4, 0, system.Sharps()), "the name is chosen again")
		assert.Equal(t, c5, s5.MustMove(a4, 1, system.Sharps()))
		assert.Equal(t, []pitch.Pitch{a4, c5, d5}, collect(t, s5.MustWalk(a4, d5, system.Sharps())))
	})

	t.Run("properties", func(t *testing.T) {
		t.Parallel()

		for _, m := range boundaryModels() {
			s := m.build(t)
			sample := samplePitches()
			for i := 0; i < len(sample); i += 11 {
				x := sample[i]
				for j := 0; j < len(sample); j += 13 {
					y := sample[j]
					d := s.MustDistance(x, y)
					assert.Equal(t, system.Step(m.position(t, y)-m.position(t, x)), d, "%s: Distance(%s, %s)", m.name, x, y)
					assert.Equal(t, d == 0, s.MustEquivalent(x, y))
				}
				for n := -12; n <= 12; n += 5 {
					r := s.MustMove(x, system.Step(n), system.Sharps())
					assert.Equal(t, system.Step(n), s.MustDistance(x, r))
					want := m.class(t, x.Spelling()) + n
					want -= floorDiv(want, m.n) * m.n
					assert.Equal(t, cls(want, m.n), s.MustPitchClass(r))
				}
			}
		}
	})

	t.Run("S1: Distance and PitchClass", func(t *testing.T) {
		t.Parallel()

		s1 := newSystem(t, defS1())
		assert.Equal(t, system.Step(1), s1.MustDistance(pt(spelling.C(), 4), pt(spelling.D(), 5)))
		assert.Equal(t, pitchclass.MustNew(0, 1), s1.MustPitchClass(pt(spelling.BSharp(), -3)))
	})
}
