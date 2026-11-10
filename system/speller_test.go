package system_test

import (
	"math"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for spelling choice: OP-32…OP-35, P-14, spec 2.5, contract 5.10, criteria K-11
// (without S5), K-12, K-19, K-20, K-22 (Move).

func TestScaleModeAndMotion(t *testing.T) {
	t.Parallel()

	assert.EqualValues(t, 0, system.ScaleModeUnset)
	assert.EqualValues(t, 1, system.Major)
	assert.EqualValues(t, 2, system.Minor)
	assert.EqualValues(t, 0, system.MotionUnset)
	assert.EqualValues(t, 1, system.Ascending)
	assert.EqualValues(t, 2, system.Descending)

	assert.Equal(t, "unset", system.ScaleModeUnset.String())
	assert.Equal(t, "major", system.Major.String())
	assert.Equal(t, "minor", system.Minor.String())
	assert.Equal(t, "ScaleMode(3)", system.ScaleMode(3).String())
	assert.Equal(t, "ScaleMode(255)", system.ScaleMode(255).String())

	assert.Equal(t, "unset", system.MotionUnset.String())
	assert.Equal(t, "ascending", system.Ascending.String())
	assert.Equal(t, "descending", system.Descending.String())
	assert.Equal(t, "Motion(3)", system.Motion(3).String())
	assert.Equal(t, "Motion(200)", system.Motion(200).String())
}

func sharpNames(spellings ...spelling.Spelling) func(*system.System, pitchclass.Class) spelling.Spelling {
	return func(_ *system.System, c pitchclass.Class) spelling.Spelling { return spellings[c.Number()] }
}

// tradSharpNames are the names of classes 0…11 of the traditional system with sharps.
func tradSharpNames() []spelling.Spelling {
	return []spelling.Spelling{
		spelling.C(), spelling.CSharp(), spelling.D(), spelling.DSharp(), spelling.E(), spelling.F(),
		spelling.FSharp(), spelling.G(), spelling.GSharp(), spelling.A(), spelling.ASharp(), spelling.B(),
	}
}

func TestSpellingRule_Construction(t *testing.T) {
	t.Parallel()

	custom, err := system.CustomSpelling(sharpNames(tradSharpNames()...))
	require.NoError(t, err)

	tests := []struct {
		name string
		rule system.SpellingRule
		str  string
	}{
		{"Sharps", system.Sharps(), "sharps"},
		{"Flats", system.Flats(), "flats"},
		{"chromatic C major ascending", system.MustChromaticScale(spelling.C(), system.Major, system.Ascending), "chromatic(C, major, ascending)"},
		{"chromatic G sharp minor descending", system.MustChromaticScale(spelling.GSharp(), system.Minor, system.Descending), "chromatic(G+1, minor, descending)"},
		{"custom", custom, "custom"},
		{"MustCustomSpelling", system.MustCustomSpelling(sharpNames(tradSharpNames()...)), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, tt.rule.IsValid())
			assert.Equal(t, tt.str, tt.rule.String())
		})
	}

	t.Run("unset", func(t *testing.T) {
		t.Parallel()

		var r system.SpellingRule
		assert.False(t, r.IsValid())
		assert.Equal(t, "SpellingRule(unset)", r.String())
	})

	errTests := []struct {
		name   string
		tonic  spelling.Spelling
		mode   system.ScaleMode
		motion system.Motion
	}{
		{"unset tonic", spelling.Spelling{}, system.Major, system.Ascending},
		{"unset mode", spelling.C(), system.ScaleModeUnset, system.Ascending},
		{"unknown mode", spelling.C(), system.ScaleMode(3), system.Ascending},
		{"unset motion", spelling.C(), system.Major, system.MotionUnset},
		{"unknown motion", spelling.C(), system.Minor, system.Motion(7)},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := system.ChromaticScale(tt.tonic, tt.mode, tt.motion)
			requireErr(t, err, system.ErrInvalid, nil)
			assert.False(t, r.IsValid())
			requirePanicErr(t, system.ErrInvalid, nil, func() { system.MustChromaticScale(tt.tonic, tt.mode, tt.motion) })
		})
	}

	t.Run("nil custom function", func(t *testing.T) {
		t.Parallel()

		r, err := system.CustomSpelling(nil)
		requireErr(t, err, system.ErrInvalid, nil)
		assert.False(t, r.IsValid())
		requirePanicErr(t, system.ErrInvalid, nil, func() { system.MustCustomSpelling(nil) })
	})
}

func TestSystem_SpellSharpsFlats(t *testing.T) {
	t.Parallel()

	// N = 3, boundary letter A: A0 B1 C1 D2 E2 F2 G2. Ties of letters are broken by the
	// circle upward from A, not by the order C…B.
	fromA := func(t testing.TB) *system.System {
		return newSystem(t, system.Definition{
			Classes: 3, Boundary: letter.A, Offsets: offs(1, 2, 2, 2, 2, 0, 1),
			Chromatic: 1, Alterations: system.Integers(),
		})
	}
	// Traditional layout with A = {0, 5, -5}.
	far := func(t testing.TB) *system.System {
		return newSystem(t, tradLayoutDef(system.MustUnion(system.Single(a(0)), system.Single(a(5)), system.Single(a(-5)))))
	}
	trad := func(testing.TB) *system.System { return system.Traditional() }
	s19 := func(testing.TB) *system.System { return system.Positions19() }

	tests := []struct {
		name   string
		build  func(t testing.TB) *system.System
		class  int
		sharps spelling.Spelling
		flats  spelling.Spelling
	}{
		{"T class 1 (K-11)", trad, 1, spelling.CSharp(), spelling.DFlat()},
		{"T class 0", trad, 0, spelling.C(), spelling.C()},
		{"T class 10", trad, 10, spelling.ASharp(), spelling.BFlat()},
		{"S19 class 7 (K-11)", s19, 7, spelling.ESharp(), spelling.FFlat()},
		{"S19 class 1 (K-11)", s19, 1, spelling.CSharp(), spelling.CSharp()},
		{"S19 class 2 (K-11)", s19, 2, spelling.DFlat(), spelling.DFlat()},
		{"S19 class 18", s19, 18, spelling.BSharp(), spelling.CFlat()},
		{"S2 class 1: six natural letters, first up from C (K-11)", func(t testing.TB) *system.System { return newSystem(t, defS2()) }, 1, spelling.D(), spelling.D()},
		{"boundary A: class 0 starts with the boundary letter itself", fromA, 0, spelling.A(), spelling.A()},
		{"boundary A: class 1 is B, not C", fromA, 1, spelling.B(), spelling.B()},
		{"boundary A: class 2 is D", fromA, 2, spelling.D(), spelling.D()},
		{"A = {0, ±5}: class 10 has only F+5", far, 10, spi(letter.F, 5), spi(letter.F, 5)},
		{"A = {0, ±5}: class 0 prefers the natural C", far, 0, spelling.C(), spelling.C()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			c := cls(tt.class, s.Classes())

			got, err := s.Spell(c, system.Sharps())
			require.NoError(t, err)
			assert.Equal(t, tt.sharps, got, "sharps")
			assert.Equal(t, tt.sharps, s.MustSpell(c, system.Sharps()))

			got, err = s.Spell(c, system.Flats())
			require.NoError(t, err)
			assert.Equal(t, tt.flats, got, "flats")
		})
	}

	t.Run("no admissible name (K-20)", func(t *testing.T) {
		t.Parallel()

		zero := newSystem(t, tradLayoutDef(system.Single(a(0))))
		for _, r := range []system.SpellingRule{system.Sharps(), system.Flats()} {
			got, err := zero.Spell(cls(1, 12), r)
			requireErr(t, err, system.ErrNoSpelling, nil)
			assert.Equal(t, spelling.Spelling{}, got)
			requirePanicErr(t, system.ErrNoSpelling, nil, func() { zero.MustSpell(cls(1, 12), r) })
		}
		_, err := far(t).Spell(cls(3, 12), system.Sharps())
		requireErr(t, err, system.ErrNoSpelling, nil)
	})

	t.Run("minimality against Spellings", func(t *testing.T) {
		t.Parallel()

		for _, m := range []model{modelT(), modelS19(), modelS24(), modelSA(), modelS14(), modelS12Irregular(), modelS2()} {
			s := m.build(t)
			b := s.Boundary()
			for c := range m.n {
				class := cls(c, m.n)
				candidates := collect(t, s.MustSpellings(class, a(4)))
				require.NotEmpty(t, candidates, "%s class %d", m.name, c)

				for _, preferSharp := range []bool{true, false} {
					best := candidates[0]
					better := func(x, y spelling.Spelling) bool {
						ax, ay := x.Alteration(), y.Alteration()
						absX, absY := ax, ay
						if absX.Sign() < 0 {
							absX = absX.MustMul(-1)
						}
						if absY.Sign() < 0 {
							absY = absY.MustMul(-1)
						}
						if d := absX.Compare(absY); d != 0 {
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
					for _, cand := range candidates[1:] {
						if better(cand, best) {
							best = cand
						}
					}
					rule := system.Flats()
					if preferSharp {
						rule = system.Sharps()
					}
					assert.Equal(t, best, s.MustSpell(class, rule), "%s class %d, %s", m.name, c, rule)
				}
			}
		}
	})
}

func TestSystem_SpellChromaticScale(t *testing.T) {
	t.Parallel()

	sp := func(names ...spelling.Spelling) []spelling.Spelling { return names }
	tests := []struct {
		name   string
		build  func(t testing.TB) *system.System
		tonic  spelling.Spelling
		mode   system.ScaleMode
		motion system.Motion
		// names in the order of the scale, starting with the tonic
		scale []spelling.Spelling
	}{
		{
			name: "C major ascending (K-11)", build: func(testing.TB) *system.System { return system.Traditional() },
			tonic: spelling.C(), mode: system.Major, motion: system.Ascending,
			scale: sp(spelling.C(), spelling.CSharp(), spelling.D(), spelling.DSharp(), spelling.E(), spelling.F(),
				spelling.FSharp(), spelling.G(), spelling.GSharp(), spelling.A(), spelling.BFlat(), spelling.B()),
		},
		{
			name: "C major descending (K-11)", build: func(testing.TB) *system.System { return system.Traditional() },
			tonic: spelling.C(), mode: system.Major, motion: system.Descending,
			scale: sp(spelling.C(), spelling.B(), spelling.BFlat(), spelling.A(), spelling.AFlat(), spelling.G(),
				spelling.FSharp(), spelling.F(), spelling.E(), spelling.EFlat(), spelling.D(), spelling.DFlat()),
		},
		{
			name: "C minor ascending (K-11)", build: func(testing.TB) *system.System { return system.Traditional() },
			tonic: spelling.C(), mode: system.Minor, motion: system.Ascending,
			scale: sp(spelling.C(), spelling.DFlat(), spelling.D(), spelling.EFlat(), spelling.E(), spelling.F(),
				spelling.FSharp(), spelling.G(), spelling.AFlat(), spelling.A(), spelling.BFlat(), spelling.B()),
		},
		{
			name: "C minor descending is as C major descending", build: func(testing.TB) *system.System { return system.Traditional() },
			tonic: spelling.C(), mode: system.Minor, motion: system.Descending,
			scale: sp(spelling.C(), spelling.B(), spelling.BFlat(), spelling.A(), spelling.AFlat(), spelling.G(),
				spelling.FSharp(), spelling.F(), spelling.E(), spelling.EFlat(), spelling.D(), spelling.DFlat()),
		},
		{
			name: "G major ascending (K-11)", build: func(testing.TB) *system.System { return system.Traditional() },
			tonic: spelling.G(), mode: system.Major, motion: system.Ascending,
			scale: sp(spelling.G(), spelling.GSharp(), spelling.A(), spelling.ASharp(), spelling.B(), spelling.C(),
				spelling.CSharp(), spelling.D(), spelling.DSharp(), spelling.E(), spelling.F(), spelling.FSharp()),
		},
		{
			// Contract 5.10, steps 1 and 2: the VI degree D is not raised, the VII degree E is lowered.
			name: "F major ascending", build: func(testing.TB) *system.System { return system.Traditional() },
			tonic: spelling.F(), mode: system.Major, motion: system.Ascending,
			scale: sp(spelling.F(), spelling.FSharp(), spelling.G(), spelling.GSharp(), spelling.A(), spelling.BFlat(),
				spelling.B(), spelling.C(), spelling.CSharp(), spelling.D(), spelling.EFlat(), spelling.E()),
		},
		{
			name: "SA: C major ascending (K-11)", build: func(t testing.TB) *system.System { return newSystem(t, defSA()) },
			tonic: spelling.C(), mode: system.Major, motion: system.Ascending,
			scale: sp(spelling.C(), spelling.CSharp(), spelling.D(), spelling.DSharp(), spelling.E(), spelling.F(),
				spelling.FSharp(), spelling.G(), spelling.GSharp(), spelling.A(), spelling.BFlat(), spelling.B()),
		},
		{
			name: "SA: C major descending (K-11)", build: func(t testing.TB) *system.System { return newSystem(t, defSA()) },
			tonic: spelling.C(), mode: system.Major, motion: system.Descending,
			scale: sp(spelling.C(), spelling.B(), spelling.BFlat(), spelling.A(), spelling.AFlat(), spelling.G(),
				spelling.FSharp(), spelling.F(), spelling.E(), spelling.EFlat(), spelling.D(), spelling.DFlat()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			rule, err := system.ChromaticScale(tt.tonic, tt.mode, tt.motion)
			require.NoError(t, err)
			require.Len(t, tt.scale, 12)

			tonicClass := s.MustClass(tt.tonic).Number()
			for i, want := range tt.scale {
				step := i
				if tt.motion == system.Descending {
					step = -i
				}
				c := cls(((tonicClass+step)%12+12)%12, 12)
				require.Equal(t, c, s.MustClass(want), "the expected name %s has the class of step %d", want, i)

				got, err := s.Spell(c, rule)
				require.NoError(t, err)
				assert.Equal(t, want, got, "class %d", c.Number())
			}
		})
	}

	t.Run("single classes (K-11)", func(t *testing.T) {
		t.Parallel()

		s := system.Traditional()
		up := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
		down := system.MustChromaticScale(spelling.C(), system.Major, system.Descending)
		assert.Equal(t, spelling.DSharp(), s.MustSpell(cls(3, 12), up))
		assert.Equal(t, spelling.EFlat(), s.MustSpell(cls(3, 12), down))
		assert.Equal(t, spelling.BFlat(), s.MustSpell(cls(10, 12), up))

		sa := newSystem(t, defSA())
		assert.Equal(t, spelling.CSharp(), sa.MustSpell(cls(4, 12), up))
		assert.Equal(t, spelling.DSharp(), sa.MustSpell(cls(6, 12), up))
		assert.Equal(t, spelling.EFlat(), sa.MustSpell(cls(6, 12), down))
	})

	t.Run("names that cannot be represented (contract 5.10)", func(t *testing.T) {
		t.Parallel()

		// Tonic C+M: every degree has alteration M; raised degrees would need M+1.
		s := system.Traditional()
		rule := system.MustChromaticScale(spi(letter.C, maxAlt), system.Major, system.Ascending)
		assert.Equal(t, spi(letter.C, maxAlt), s.MustSpell(cls(4, 12), rule))
		assert.Equal(t, spi(letter.B, maxAlt-1), s.MustSpell(cls(2, 12), rule), "lowered VII")
		got, err := s.Spell(cls(5, 12), rule)
		requireErr(t, err, system.ErrNoSpelling, nil)
		assert.Equal(t, spelling.Spelling{}, got)
	})

	errTests := []struct {
		name     string
		build    func(t testing.TB) *system.System
		class    pitchclass.Class
		rule     func() system.SpellingRule
		category error
	}{
		{
			name: "G sharp major ascending needs F double sharp (K-20)",
			build: func(t testing.TB) *system.System {
				return newSystem(t, tradLayoutDef(system.MustIntegerRange(-1, 1)))
			},
			class: cls(7, 12),
			rule: func() system.SpellingRule {
				return system.MustChromaticScale(spelling.GSharp(), system.Major, system.Ascending)
			},
			category: system.ErrNoSpelling,
		},
		{
			name: "tonic G sharp with A = {0} (K-20)",
			build: func(t testing.TB) *system.System {
				return newSystem(t, tradLayoutDef(system.Single(a(0))))
			},
			class: cls(8, 12),
			rule: func() system.SpellingRule {
				return system.MustChromaticScale(spelling.GSharp(), system.Major, system.Ascending)
			},
			category: system.ErrInadmissible,
		},
		{
			name:  "19-position system (K-20)",
			build: func(testing.TB) *system.System { return system.Positions19() },
			class: cls(0, 19),
			rule: func() system.SpellingRule {
				return system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
			},
			category: system.ErrNotApplicable,
		},
		{
			name:  "not applicable before inadmissible tonic",
			build: func(testing.TB) *system.System { return system.Positions19() },
			class: cls(0, 19),
			rule: func() system.SpellingRule {
				return system.MustChromaticScale(dHalfSharp(), system.Major, system.Ascending)
			},
			category: system.ErrNotApplicable,
		},
		{
			name:  "class of another system before not applicable",
			build: func(testing.TB) *system.System { return system.Positions19() },
			class: cls(0, 12),
			rule: func() system.SpellingRule {
				return system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
			},
			category: system.ErrModulusMismatch,
		},
		{
			name:  "S14 has no traditional layout",
			build: func(t testing.TB) *system.System { return newSystem(t, defS14()) },
			class: cls(0, 14),
			rule: func() system.SpellingRule {
				return system.MustChromaticScale(spelling.C(), system.Minor, system.Descending)
			},
			category: system.ErrNotApplicable,
		},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Spell(tt.class, tt.rule())
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, spelling.Spelling{}, got)
		})
	}
}

func TestSystem_SpellCustom(t *testing.T) {
	t.Parallel()

	s := system.Traditional()

	t.Run("chooses the returned name", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int64
		rule := system.MustCustomSpelling(func(got *system.System, c pitchclass.Class) spelling.Spelling {
			calls.Add(1)
			assert.True(t, got.Equal(s), "the function gets the system")
			assert.Equal(t, 12, c.Modulus())
			if c.Number() == 0 {
				return spelling.BSharp()
			}
			return tradSharpNames()[c.Number()]
		})
		assert.Equal(t, spelling.BSharp(), s.MustSpell(cls(0, 12), rule))
		assert.Equal(t, spelling.FSharp(), s.MustSpell(cls(6, 12), rule))
		assert.Positive(t, calls.Load())
	})

	tests := []struct {
		name   string
		choose func(*system.System, pitchclass.Class) spelling.Spelling
	}{
		{"unset name", func(*system.System, pitchclass.Class) spelling.Spelling { return spelling.Spelling{} }},
		{"inadmissible name", func(*system.System, pitchclass.Class) spelling.Spelling { return dHalfSharp() }},
		{"name of another class", func(*system.System, pitchclass.Class) spelling.Spelling { return spelling.D() }},
	}
	for _, tt := range tests {
		t.Run(tt.name+" (K-20)", func(t *testing.T) {
			t.Parallel()

			rule := system.MustCustomSpelling(tt.choose)
			got, err := s.Spell(cls(1, 12), rule)
			requireErr(t, err, system.ErrBadSpelling, nil)
			assert.Equal(t, spelling.Spelling{}, got)
			requirePanicErr(t, system.ErrBadSpelling, nil, func() { s.MustSpell(cls(1, 12), rule) })
		})
	}

	t.Run("a panic of the function is not recovered", func(t *testing.T) {
		t.Parallel()

		rule := system.MustCustomSpelling(func(*system.System, pitchclass.Class) spelling.Spelling { panic("boom") })
		assert.PanicsWithValue(t, "boom", func() { _, _ = s.Spell(cls(1, 12), rule) })
	})
}

func TestSystem_SpellErrors(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	tests := []struct {
		name     string
		s        *system.System
		class    pitchclass.Class
		rule     system.SpellingRule
		category error
	}{
		{"unset system", nil, cls(0, 12), system.Sharps(), system.ErrInvalid},
		{"unset class", s, pitchclass.Class{}, system.Sharps(), system.ErrInvalid},
		{"unset rule", s, cls(0, 12), system.SpellingRule{}, system.ErrInvalid},
		{"unset rule and class of another system", s, cls(0, 19), system.SpellingRule{}, system.ErrInvalid},
		{"class of another system (K-19)", s, cls(0, 19), system.Sharps(), system.ErrModulusMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.s.Spell(tt.class, tt.rule)
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, spelling.Spelling{}, got)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustSpell(tt.class, tt.rule) })
		})
	}
}

func TestSystem_PitchAt(t *testing.T) {
	t.Parallel()

	s := system.Traditional()

	tests := []struct {
		name  string
		class pitchclass.Class
		oct   int
		rule  system.SpellingRule
		want  pitch.Pitch
	}{
		{"class 0, Octave 4, sharps (K-12)", cls(0, 12), 4, system.Sharps(), pt(spelling.C(), 4)},
		{"class 11, Octave 3, sharps (K-12)", cls(11, 12), 3, system.Sharps(), pt(spelling.B(), 3)},
		{"class 1, Octave 4, flats", cls(1, 12), 4, system.Flats(), pt(spelling.DFlat(), 4)},
		{"class 0 named B sharp keeps the Octave", cls(0, 12), 4, system.MustCustomSpelling(func(*system.System, pitchclass.Class) spelling.Spelling {
			return spelling.BSharp()
		}), pt(spelling.BSharp(), 4)},
		{"Octave MAX", cls(3, 12), math.MaxInt, system.MustChromaticScale(spelling.C(), system.Major, system.Descending), pt(spelling.EFlat(), math.MaxInt)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := s.PitchAt(tt.class, o(tt.oct), tt.rule)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustPitchAt(tt.class, o(tt.oct), tt.rule))
		})
	}

	errTests := []struct {
		name     string
		s        *system.System
		class    pitchclass.Class
		oct      octave.Octave
		rule     system.SpellingRule
		category error
	}{
		{"unset octave", s, cls(0, 12), octave.Octave{}, system.Sharps(), system.ErrInvalid},
		{"unset octave and class of another system", s, cls(0, 19), octave.Octave{}, system.Sharps(), system.ErrInvalid},
		{"unset rule", s, cls(0, 12), o(4), system.SpellingRule{}, system.ErrInvalid},
		{"unset system", nil, cls(0, 12), o(4), system.Sharps(), system.ErrInvalid},
		{"class of another system (K-19)", s, cls(0, 19), o(4), system.Sharps(), system.ErrModulusMismatch},
		{"not applicable", system.Positions19(), cls(0, 19), o(4), system.MustChromaticScale(spelling.C(), system.Major, system.Ascending), system.ErrNotApplicable},
		{"no admissible name", newSystem(t, tradLayoutDef(system.Single(a(0)))), cls(1, 12), o(4), system.Sharps(), system.ErrNoSpelling},
		{"bad custom name", s, cls(1, 12), o(4), system.MustCustomSpelling(func(*system.System, pitchclass.Class) spelling.Spelling { return spelling.C() }), system.ErrBadSpelling},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.s.PitchAt(tt.class, tt.oct, tt.rule)
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, pitch.Pitch{}, got)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustPitchAt(tt.class, tt.oct, tt.rule) })
		})
	}
}

func TestSystem_Move(t *testing.T) {
	t.Parallel()

	trad := func(testing.TB) *system.System { return system.Traditional() }
	n1 := func(t testing.TB) *system.System {
		return newSystem(t, system.Definition{Classes: 1, Boundary: letter.C, Chromatic: 1, Alterations: system.Integers()})
	}

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		p     pitch.Pitch
		n     system.Step
		rule  system.SpellingRule
		want  pitch.Pitch
	}{
		{"C4 + 16 is E5 (K-12)", trad, pt(spelling.C(), 4), 16, system.Sharps(), pt(spelling.E(), 5)},
		{"C4 + 3 is D sharp 4 (K-12)", trad, pt(spelling.C(), 4), 3, system.Sharps(), pt(spelling.DSharp(), 4)},
		{"C4 + 3 is E flat 4 with flats (K-12)", trad, pt(spelling.C(), 4), 3, system.Flats(), pt(spelling.EFlat(), 4)},
		{"B3 + 1 is C4 (K-12)", trad, pt(spelling.B(), 3), 1, system.Sharps(), pt(spelling.C(), 4)},
		{"C4 - 1 is B3 (K-12)", trad, pt(spelling.C(), 4), -1, system.Sharps(), pt(spelling.B(), 3)},
		{"n = 0 chooses the name again", trad, pt(spelling.BSharp(), 4), 0, system.Sharps(), pt(spelling.C(), 4)},
		{"C4 - 13 is B2", trad, pt(spelling.C(), 4), -13, system.Flats(), pt(spelling.B(), 2)},
		{"S19: C4 + 1 is C sharp 4", func(testing.TB) *system.System { return system.Positions19() }, pt(spelling.C(), 4), 1, system.Sharps(), pt(spelling.CSharp(), 4)},
		{"SA: G sharp 3 + 1 is A4 written, Octave 4", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, pt(spelling.GSharp(), 3), 1, system.Sharps(), pt(spelling.A(), 4)},
		{"n = MaxInt from Octave 0", trad, pt(spelling.C(), 0), math.MaxInt, system.Sharps(), pt(spelling.G(), 768614336404564650)},
		{"n = MinInt from Octave 0", trad, pt(spelling.C(), 0), math.MinInt, system.Sharps(), pt(spelling.E(), -768614336404564651)},
		{"to the top of the axis", trad, pt(spelling.C(), math.MaxInt), 11, system.Sharps(), pt(spelling.B(), math.MaxInt)},
		{"to the bottom of the axis", trad, pt(spelling.B(), math.MinInt), -11, system.Sharps(), pt(spelling.C(), math.MinInt)},
		{"N = 1: n = MaxInt", n1, pt(spelling.D(), 0), math.MaxInt, system.Sharps(), pt(spelling.C(), math.MaxInt)},
		{"N = 1: n = MinInt", n1, pt(spelling.D(), 0), math.MinInt, system.Flats(), pt(spelling.C(), math.MinInt)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Move(tt.p, tt.n, tt.rule)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustMove(tt.p, tt.n, tt.rule))
		})
	}

	t.Run("properties", func(t *testing.T) {
		t.Parallel()

		for _, m := range []model{modelT(), modelS19(), modelSA(), modelS14(), modelS2()} {
			s := m.build(t)
			all := samplePitches()
			for i := 0; i < len(all); i += 5 {
				p := all[i]
				for n := -30; n <= 30; n += 7 {
					for _, rule := range []system.SpellingRule{system.Sharps(), system.Flats()} {
						r, err := s.Move(p, system.Step(n), rule)
						require.NoError(t, err)
						assert.Equal(t, system.Step(n), s.MustDistance(p, r), "%s: Distance(%s, Move(%s, %d))", m.name, p, p, n)
						want := m.class(t, p.Spelling()) + n
						want -= floorDiv(want, m.n) * m.n
						assert.Equal(t, cls(want, m.n), s.MustPitchClass(r))
						assert.Equal(t, s.MustSpell(cls(want, m.n), rule), r.Spelling(), "the name is chosen by the rule")
					}
				}
			}
		}
	})

	errTests := []struct {
		name     string
		s        *system.System
		p        pitch.Pitch
		n        system.Step
		rule     system.SpellingRule
		category error
	}{
		{"B at Octave MAX + 1 (K-22)", system.Traditional(), pt(spelling.B(), math.MaxInt), 1, system.Sharps(), system.ErrOutOfRange},
		{"C at Octave MIN - 1", system.Traditional(), pt(spelling.C(), math.MinInt), -1, system.Sharps(), system.ErrOutOfRange},
		{"n = MaxInt from Octave MAX", system.Traditional(), pt(spelling.C(), math.MaxInt), math.MaxInt, system.Sharps(), system.ErrOutOfRange},
		{"n = MinInt from Octave MIN", system.Traditional(), pt(spelling.C(), math.MinInt), math.MinInt, system.Flats(), system.ErrOutOfRange},
		{"inadmissible pitch (K-17)", system.Traditional(), pt(dHalfSharp(), 4), 1, system.Sharps(), system.ErrInadmissible},
		{"unset pitch", system.Traditional(), pitch.Pitch{}, 1, system.Sharps(), system.ErrInvalid},
		{"unset rule", system.Traditional(), pt(spelling.C(), 4), 1, system.SpellingRule{}, system.ErrInvalid},
		{"unset system", nil, pt(spelling.C(), 4), 1, system.Sharps(), system.ErrInvalid},
		{"not applicable before inadmissible", system.Positions19(), pt(dHalfSharp(), 4), 1, system.MustChromaticScale(spelling.C(), system.Major, system.Ascending), system.ErrNotApplicable},
		{"no admissible name", newSystem(t, tradLayoutDef(system.Single(a(0)))), pt(spelling.C(), 4), 1, system.Sharps(), system.ErrNoSpelling},
		{"no admissible name before out of range", newSystem(t, tradLayoutDef(system.Single(a(0)))), pt(spelling.C(), math.MaxInt), 13, system.Sharps(), system.ErrNoSpelling},
		{"bad custom name before out of range", system.Traditional(), pt(spelling.B(), math.MaxInt), 1, system.MustCustomSpelling(func(*system.System, pitchclass.Class) spelling.Spelling {
			return dHalfSharp()
		}), system.ErrBadSpelling},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.s.Move(tt.p, tt.n, tt.rule)
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, pitch.Pitch{}, got)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustMove(tt.p, tt.n, tt.rule) })
		})
	}
}

func TestSystem_Walk(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	c4, e4 := pt(spelling.C(), 4), pt(spelling.E(), 4)

	t.Run("C4 to E4 and back (K-12)", func(t *testing.T) {
		t.Parallel()

		up := []pitch.Pitch{c4, pt(spelling.CSharp(), 4), pt(spelling.D(), 4), pt(spelling.DSharp(), 4), e4}
		seq, err := s.Walk(c4, e4, system.Sharps())
		require.NoError(t, err)
		assert.Equal(t, up, collect(t, seq))
		assert.Equal(t, up, collect(t, seq), "repeated traversal")
		assert.Equal(t, up[:1], first(t, seq, 1), "early stop")

		down := []pitch.Pitch{e4, pt(spelling.DSharp(), 4), pt(spelling.D(), 4), pt(spelling.CSharp(), 4), c4}
		assert.Equal(t, down, collect(t, s.MustWalk(e4, c4, system.Sharps())))
	})

	t.Run("one pitch when from = to", func(t *testing.T) {
		t.Parallel()

		got := collect(t, s.MustWalk(pt(spelling.BSharp(), 4), c4, system.Sharps()))
		assert.Equal(t, []pitch.Pitch{c4}, got)
	})

	t.Run("chromatic scale C4 to C5", func(t *testing.T) {
		t.Parallel()

		rule := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
		want := []pitch.Pitch{
			c4, pt(spelling.CSharp(), 4), pt(spelling.D(), 4), pt(spelling.DSharp(), 4), e4, pt(spelling.F(), 4),
			pt(spelling.FSharp(), 4), pt(spelling.G(), 4), pt(spelling.GSharp(), 4), pt(spelling.A(), 4),
			pt(spelling.BFlat(), 4), pt(spelling.B(), 4), pt(spelling.C(), 5),
		}
		assert.Equal(t, want, collect(t, s.MustWalk(c4, pt(spelling.C(), 5), rule)))
	})

	t.Run("the whole axis is walkable lazily", func(t *testing.T) {
		t.Parallel()

		lo, hi := pt(spelling.C(), math.MinInt), pt(spelling.C(), math.MaxInt)
		seq, err := s.Walk(lo, hi, system.Sharps())
		require.NoError(t, err)
		assert.Equal(t, []pitch.Pitch{lo, pt(spelling.CSharp(), math.MinInt), pt(spelling.D(), math.MinInt)}, first(t, seq, 3))

		seq, err = s.Walk(hi, lo, system.Flats())
		require.NoError(t, err)
		assert.Equal(t, []pitch.Pitch{hi, pt(spelling.B(), math.MaxInt-1), pt(spelling.BFlat(), math.MaxInt-1)}, first(t, seq, 3))
	})

	t.Run("custom function is called once per class before the first pitch", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int64
		rule := system.MustCustomSpelling(func(_ *system.System, c pitchclass.Class) spelling.Spelling {
			calls.Add(1)
			return tradSharpNames()[c.Number()]
		})

		seq, err := s.Walk(c4, pt(spelling.C(), 5), rule)
		require.NoError(t, err)
		assert.EqualValues(t, 12, calls.Load(), "12 classes among 13 positions")
		assert.Len(t, collect(t, seq), 13)
		assert.Len(t, collect(t, seq), 13)
		assert.EqualValues(t, 12, calls.Load(), "traversals use the names chosen in advance")

		calls.Store(0)
		_, err = s.Walk(c4, e4, rule)
		require.NoError(t, err)
		assert.EqualValues(t, 5, calls.Load())
	})

	t.Run("properties", func(t *testing.T) {
		t.Parallel()

		for _, m := range []model{modelT(), modelS19(), modelSA(), modelS14()} {
			sys := m.build(t)
			pairs := [][2]pitch.Pitch{
				{pt(spelling.C(), 4), pt(spelling.G(), 5)},
				{pt(spelling.BSharp(), 3), pt(spelling.CFlat(), 2)},
				{pt(spelling.EDoubleFlat(), 0), pt(spelling.ESharp(), 0)},
			}
			for _, pair := range pairs {
				from, to := pair[0], pair[1]
				d := sys.MustDistance(from, to)
				dir := system.Step(1)
				if d < 0 {
					dir = -1
				}
				got := collect(t, sys.MustWalk(from, to, system.Sharps()))
				require.Len(t, got, int(d*dir)+1, "%s: Walk(%s, %s)", m.name, from, to)
				assert.True(t, sys.MustEquivalent(from, got[0]))
				assert.True(t, sys.MustEquivalent(to, got[len(got)-1]))
				for i, p := range got {
					assert.Equal(t, system.Step(i)*dir, sys.MustDistance(from, p))
					assert.Equal(t, sys.MustSpell(sys.MustPitchClass(p), system.Sharps()), p.Spelling())
					if i > 0 {
						assert.Equal(t, dir, sys.MustDistance(got[i-1], p))
					}
				}
			}
		}
	})

	errTests := []struct {
		name     string
		s        *system.System
		from, to pitch.Pitch
		rule     system.SpellingRule
		category error
	}{
		{"unset system", nil, c4, e4, system.Sharps(), system.ErrInvalid},
		{"unset from", s, pitch.Pitch{}, e4, system.Sharps(), system.ErrInvalid},
		{"unset to", s, c4, pitch.Pitch{}, system.Sharps(), system.ErrInvalid},
		{"unset rule", s, c4, e4, system.SpellingRule{}, system.ErrInvalid},
		{"inadmissible from and unset rule", s, pt(dHalfSharp(), 4), e4, system.SpellingRule{}, system.ErrInvalid},
		{"not applicable before inadmissible", system.Positions19(), pt(dHalfSharp(), 4), e4, system.MustChromaticScale(spelling.C(), system.Major, system.Ascending), system.ErrNotApplicable},
		{"inadmissible from", s, pt(dHalfSharp(), 4), e4, system.Sharps(), system.ErrInadmissible},
		{"inadmissible to", s, c4, pt(dHalfSharp(), 4), system.Sharps(), system.ErrInadmissible},
		{"a class without a name on the way", newSystem(t, tradLayoutDef(system.Single(a(0)))), c4, pt(spelling.D(), 4), system.Sharps(), system.ErrNoSpelling},
		{"bad custom name", s, c4, e4, system.MustCustomSpelling(func(_ *system.System, c pitchclass.Class) spelling.Spelling {
			if c.Number() == 3 {
				return spelling.E()
			}
			return tradSharpNames()[c.Number()]
		}), system.ErrBadSpelling},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := tt.s.Walk(tt.from, tt.to, tt.rule)
			requireErr(t, err, tt.category, nil)
			assert.Nil(t, seq)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustWalk(tt.from, tt.to, tt.rule) })
		})
	}

	t.Run("no error for classes outside the walk", func(t *testing.T) {
		t.Parallel()

		zero := newSystem(t, tradLayoutDef(system.Single(a(0))))
		got := collect(t, zero.MustWalk(c4, c4, system.Sharps()))
		assert.Equal(t, []pitch.Pitch{c4}, got)
		got = collect(t, zero.MustWalk(pt(spelling.E(), 4), pt(spelling.F(), 4), system.Sharps()))
		assert.Equal(t, []pitch.Pitch{pt(spelling.E(), 4), pt(spelling.F(), 4)}, got)
	})
}
