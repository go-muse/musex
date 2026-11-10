package system_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for defining and reading a system: OP-1…OP-3, OP-5…OP-9, OP-39,
// P-2 (without p(l) = N), P-3, P-11, P-12, contract 5.6, criteria K-2 (without S5, S1),
// K-14 (systems), K-16, K-18 (readers).

func TestNew_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		def  system.Definition
	}{
		{name: "traditional layout", def: tradLayoutDef(system.Integers())},
		{name: "SA: boundary letter A (K-2)", def: defSA()},
		{name: "S14 (K-2)", def: defS14()},
		{name: "irregular 12-class system (K-2)", def: defS12Irregular()},
		{name: "S2: equal offsets are allowed", def: defS2()},
		{name: "N = MaxClasses", def: system.Definition{
			Classes: system.MaxClasses, Boundary: letter.C, Offsets: offs(0, 1, 2, 3, 4, 5, 6),
			Chromatic: 1, Alterations: system.Integers(),
		}},
		{name: "c = MaxChromatic", def: system.Definition{
			Classes: 12, Boundary: letter.C, Offsets: tradOffsets(),
			Chromatic: system.MaxChromatic, Alterations: system.Integers(),
		}},
		{name: "c = MaxChromatic with steps 1/MaxDenominator", def: system.Definition{
			Classes: system.MaxClasses, Boundary: letter.C, Offsets: offs(0, 1, 2, 3, 4, 5, 6),
			Chromatic: system.MaxChromatic, Alterations: system.MustMultiples(system.MaxDenominator),
		}},
		{name: "thirds with c = 3", def: system.Definition{
			Classes: 36, Boundary: letter.C, Offsets: offs(0, 6, 12, 15, 21, 27, 33),
			Chromatic: 3, Alterations: system.MustMultiples(3),
		}},
		{name: "traditional layout with A = {0}", def: tradLayoutDef(system.Single(a(0)))},
		{name: "empty name", def: system.Definition{
			Classes: 12, Boundary: letter.C, Offsets: tradOffsets(), Chromatic: 1, Alterations: system.Integers(),
		}},
		{name: "boundary letter B", def: system.Definition{
			Classes: 12, Boundary: letter.B, Offsets: offs(1, 3, 5, 6, 8, 10, 0),
			Chromatic: 1, Alterations: system.Integers(),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := system.New(tt.def)
			require.NoError(t, err)
			require.NotNil(t, s)
			assert.True(t, s.IsValid())
			assert.Equal(t, tt.def, s.Definition())
			assert.True(t, s.Equal(system.MustNew(tt.def)))
		})
	}
}

func TestNew_Errors(t *testing.T) {
	t.Parallel()

	withT := func(change func(d *system.Definition)) system.Definition {
		d := tradLayoutDef(system.Integers())
		change(&d)
		return d
	}

	tests := []struct {
		name     string
		def      system.Definition
		category error
		cause    error
	}{
		// E-1 before every cause (K-16, K-18).
		{name: "zero definition", def: system.Definition{}, category: system.ErrInvalid},
		{name: "unset boundary letter", def: withT(func(d *system.Definition) { d.Boundary = 0 }), category: system.ErrInvalid},
		{name: "invalid boundary letter", def: withT(func(d *system.Definition) { d.Boundary = letter.Letter(9) }), category: system.ErrInvalid},
		{name: "nil rule", def: withT(func(d *system.Definition) { d.Alterations = nil }), category: system.ErrInvalid},
		{name: "unset rule", def: withT(func(d *system.Definition) { d.Alterations = &system.Rule{} }), category: system.ErrInvalid},
		{name: "unset rule and N = 0", def: withT(func(d *system.Definition) {
			d.Alterations = nil
			d.Classes = 0
		}), category: system.ErrInvalid},

		// Cause 1.
		{name: "N = 0", def: withT(func(d *system.Definition) { d.Classes = 0 }), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "N = -12", def: withT(func(d *system.Definition) { d.Classes = -12 }), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "N = MaxClasses+1", def: withT(func(d *system.Definition) { d.Classes = system.MaxClasses + 1 }), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "N = MinInt", def: withT(func(d *system.Definition) { d.Classes = math.MinInt }), category: system.ErrDefinition, cause: system.ErrClassCount},

		// Cause 2.
		{name: "boundary letter C on offset 1", def: withT(func(d *system.Definition) { d.Offsets[0] = 1 }), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "negative offset", def: withT(func(d *system.Definition) { d.Offsets[1] = -1 }), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "offset far above N", def: withT(func(d *system.Definition) { d.Offsets[6] = 100 }), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "offset MaxInt", def: withT(func(d *system.Definition) { d.Offsets[6] = math.MaxInt }), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "boundary letter A on a nonzero offset", def: func() system.Definition {
			d := defSA()
			d.Offsets[5] = 1
			return d
		}(), category: system.ErrDefinition, cause: system.ErrOffsets},

		// Cause 3.
		{name: "D 3, E 2", def: withT(func(d *system.Definition) {
			d.Offsets[1] = 3
			d.Offsets[2] = 2
		}), category: system.ErrDefinition, cause: system.ErrDescending},
		{name: "descending around the boundary letter A", def: func() system.Definition {
			d := defSA()
			d.Offsets[0] = 1 // C below B (2) when counted from A
			return d
		}(), category: system.ErrDefinition, cause: system.ErrDescending},

		// Cause 4.
		{name: "c = 0", def: withT(func(d *system.Definition) { d.Chromatic = 0 }), category: system.ErrDefinition, cause: system.ErrChromatic},
		{name: "c = -1", def: withT(func(d *system.Definition) { d.Chromatic = -1 }), category: system.ErrDefinition, cause: system.ErrChromatic},
		{name: "c = MaxChromatic+1", def: withT(func(d *system.Definition) { d.Chromatic = system.MaxChromatic + 1 }), category: system.ErrDefinition, cause: system.ErrChromatic},

		// Cause 5.
		{name: "A = {-1, 1}", def: withT(func(d *system.Definition) {
			d.Alterations = system.MustUnion(system.Single(a(-1)), system.Single(a(1)))
		}), category: system.ErrDefinition, cause: system.ErrNoNatural},
		{name: "A = 1/2 + Z", def: withT(func(d *system.Definition) {
			d.Alterations = system.MustNewRule(system.MustNewProgression(q(1, 2), a(1)))
		}), category: system.ErrDefinition, cause: system.ErrNoNatural},

		// Cause 6.
		{name: "c = 1, A = {0, 1/3}", def: withT(func(d *system.Definition) {
			d.Alterations = system.MustUnion(system.Single(a(0)), system.Single(q(1, 3)))
		}), category: system.ErrDefinition, cause: system.ErrFractionalShift},
		{name: "c = 1, A = Z/2", def: withT(func(d *system.Definition) {
			d.Alterations = system.MustMultiples(2)
		}), category: system.ErrDefinition, cause: system.ErrFractionalShift},
		{name: "c = 2, A = Z/3", def: withT(func(d *system.Definition) {
			d.Chromatic = 2
			d.Alterations = system.MustMultiples(3)
		}), category: system.ErrDefinition, cause: system.ErrFractionalShift},
		{name: "c = 2, A = Z with 1/4", def: withT(func(d *system.Definition) {
			d.Chromatic = 2
			d.Alterations = system.MustUnion(system.Integers(), system.Single(q(1, 4)))
		}), category: system.ErrDefinition, cause: system.ErrFractionalShift},

		// The first cause by order wins.
		{name: "N and c invalid", def: withT(func(d *system.Definition) {
			d.Classes = 0
			d.Chromatic = 0
		}), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "offsets and c invalid", def: withT(func(d *system.Definition) {
			d.Offsets[0] = 1
			d.Chromatic = 0
		}), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "descending and no natural", def: withT(func(d *system.Definition) {
			d.Offsets[1] = 3
			d.Offsets[2] = 2
			d.Alterations = system.Single(a(1))
		}), category: system.ErrDefinition, cause: system.ErrDescending},
		{name: "c invalid and no natural", def: withT(func(d *system.Definition) {
			d.Chromatic = 0
			d.Alterations = system.Single(a(1))
		}), category: system.ErrDefinition, cause: system.ErrChromatic},
		{name: "no natural and fractional shift", def: withT(func(d *system.Definition) {
			d.Alterations = system.Single(q(1, 3))
		}), category: system.ErrDefinition, cause: system.ErrNoNatural},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := system.New(tt.def)
			requireErr(t, err, tt.category, tt.cause)
			assert.Nil(t, s)
			requirePanicErr(t, tt.category, tt.cause, func() { system.MustNew(tt.def) })
		})
	}
}

func TestNew_CopiesDefinition(t *testing.T) {
	t.Parallel()

	def := defSA()
	s := newSystem(t, def)
	want := def

	def.Name = "changed"
	def.Classes = 13
	def.Offsets[0] = 4
	def.Chromatic = 2
	def.Alterations = system.MustMultiples(2)

	assert.Equal(t, want, s.Definition())
	assert.Equal(t, "SA", s.Name())

	got := s.Definition()
	got.Offsets[1] = 99
	got.Name = "other"
	assert.Equal(t, want, s.Definition(), "changing a returned definition does not change the system")
}

func TestFromSteps(t *testing.T) {
	t.Parallel()

	tradSteps := [letter.Count]int{2, 2, 1, 2, 2, 2, 1}

	t.Run("traditional steps from C equal the traditional system (K-2)", func(t *testing.T) {
		t.Parallel()

		s, err := system.FromSteps(letter.C, tradSteps, 1, system.Integers(), "steps")
		require.NoError(t, err)
		assert.True(t, s.Equal(system.Traditional()))
		assert.Equal(t, "steps", s.Name())
		assert.Equal(t, 12, s.Classes())
		assert.Equal(t, tradOffsets(), s.Definition().Offsets)
		assert.Equal(t, tradSteps, s.Steps())
		assert.True(t, s.Equal(system.MustFromSteps(letter.C, tradSteps, 1, system.Integers(), "")))
	})

	t.Run("traditional steps from A give SA", func(t *testing.T) {
		t.Parallel()

		s, err := system.FromSteps(letter.A, tradSteps, 1, system.Integers(), "SA")
		require.NoError(t, err)
		requireDefinition(t, defSA(), s.Definition())
		assert.True(t, s.Equal(newSystem(t, defSA())))
	})

	t.Run("steps of S14", func(t *testing.T) {
		t.Parallel()

		s, err := system.FromSteps(letter.C, [letter.Count]int{2, 2, 2, 2, 2, 2, 2}, 1, system.Integers(), "S14")
		require.NoError(t, err)
		requireDefinition(t, defS14(), s.Definition())
	})

	tests := []struct {
		name     string
		b        letter.Letter
		steps    [letter.Count]int
		c        int
		rule     *system.Rule
		category error
		cause    error
	}{
		{name: "unset boundary letter", b: 0, steps: tradSteps, c: 1, rule: system.Integers(), category: system.ErrInvalid},
		{name: "nil rule", b: letter.C, steps: tradSteps, c: 1, rule: nil, category: system.ErrInvalid},
		{name: "all steps 0", b: letter.C, c: 1, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "sum above MaxClasses", b: letter.C, steps: [letter.Count]int{system.MaxClasses, 1}, c: 1, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "sum overflows int", b: letter.C, steps: [letter.Count]int{math.MaxInt, 1}, c: 1, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "sum underflows int", b: letter.C, steps: [letter.Count]int{math.MinInt, -1}, c: 1, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "intermediate sum overflows int", b: letter.C, steps: [letter.Count]int{math.MaxInt, 1, -math.MaxInt}, c: 1, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "negative step", b: letter.C, steps: [letter.Count]int{3, -1, 2, 2, 2, 2, 2}, c: 1, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrDescending},
		{name: "c = 0", b: letter.C, steps: tradSteps, c: 0, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrChromatic},
		{name: "no natural", b: letter.C, steps: tradSteps, c: 1, rule: system.Single(a(1)), category: system.ErrDefinition, cause: system.ErrNoNatural},
		{name: "fractional shift", b: letter.C, steps: tradSteps, c: 1, rule: system.MustMultiples(2), category: system.ErrDefinition, cause: system.ErrFractionalShift},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := system.FromSteps(tt.b, tt.steps, tt.c, tt.rule, "x")
			requireErr(t, err, tt.category, tt.cause)
			assert.Nil(t, s)
			requirePanicErr(t, tt.category, tt.cause, func() {
				system.MustFromSteps(tt.b, tt.steps, tt.c, tt.rule, "x")
			})
		})
	}
}

func TestNewRegular(t *testing.T) {
	t.Parallel()

	t.Run("B = 2, M = 1 equals the traditional system (K-2)", func(t *testing.T) {
		t.Parallel()

		s, err := system.NewRegular(2, 1, letter.C, system.Integers(), "regular")
		require.NoError(t, err)
		assert.True(t, s.Equal(system.Traditional()))
		assert.Equal(t, "regular", s.Name())
	})

	t.Run("B = 5, M = 3 gives N = 31, c = 2 (K-2)", func(t *testing.T) {
		t.Parallel()

		s, err := system.NewRegular(5, 3, letter.C, system.Integers(), "31")
		require.NoError(t, err)
		assert.Equal(t, 31, s.Classes())
		assert.Equal(t, 2, s.Chromatic())
		assert.Equal(t, [letter.Count]int{5, 5, 3, 5, 5, 5, 3}, s.Steps())
		assert.Equal(t, offs(0, 5, 10, 13, 18, 23, 28), s.Definition().Offsets)
		assert.Equal(t, system.Regularity{Regular: true, Large: 5, Small: 3, RegularByClasses: true}, s.MustRegularity())
	})

	t.Run("B = 3, M = 2 equals the 19-position system", func(t *testing.T) {
		t.Parallel()

		s := system.MustNewRegular(3, 2, letter.C, system.Integers(), "")
		assert.True(t, s.Equal(system.Positions19()))
	})

	t.Run("boundary letter A", func(t *testing.T) {
		t.Parallel()

		s := system.MustNewRegular(2, 1, letter.A, system.Integers(), "SA")
		requireDefinition(t, defSA(), s.Definition())
	})

	tests := []struct {
		name         string
		large, small int
		b            letter.Letter
		rule         *system.Rule
		category     error
		cause        error
	}{
		{name: "unset boundary letter", large: 2, small: 1, b: 0, rule: system.Integers(), category: system.ErrInvalid},
		{name: "unset rule", large: 2, small: 1, b: letter.C, rule: &system.Rule{}, category: system.ErrInvalid},
		{name: "B = M gives c = 0 (OP-3)", large: 2, small: 2, b: letter.C, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrChromatic},
		{name: "B < M gives c < 0 (OP-3)", large: 1, small: 2, b: letter.C, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrChromatic},
		{name: "B = M = 0 gives N = 0", large: 0, small: 0, b: letter.C, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "N above MaxClasses", large: system.MaxClasses, small: 1, b: letter.C, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "N overflows int", large: math.MaxInt / 2, small: 0, b: letter.C, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrClassCount},
		{name: "negative M puts B above N", large: 3, small: -1, b: letter.C, rule: system.Integers(), category: system.ErrDefinition, cause: system.ErrOffsets},
		{name: "no natural", large: 2, small: 1, b: letter.C, rule: system.Single(a(1)), category: system.ErrDefinition, cause: system.ErrNoNatural},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := system.NewRegular(tt.large, tt.small, tt.b, tt.rule, "x")
			requireErr(t, err, tt.category, tt.cause)
			assert.Nil(t, s)
			requirePanicErr(t, tt.category, tt.cause, func() {
				system.MustNewRegular(tt.large, tt.small, tt.b, tt.rule, "x")
			})
		})
	}
}

func TestSystem_Readers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		build   func(t testing.TB) *system.System
		def     system.Definition
		steps   [letter.Count]int
		natural [letter.Count]int // class numbers of the natural letters
	}{
		{
			name:    "traditional",
			build:   func(testing.TB) *system.System { return system.Traditional() },
			def:     system.Definition{Name: "traditional", Classes: 12, Boundary: letter.C, Offsets: tradOffsets(), Chromatic: 1},
			steps:   [letter.Count]int{2, 2, 1, 2, 2, 2, 1},
			natural: [letter.Count]int{0, 2, 4, 5, 7, 9, 11},
		},
		{
			name:    "SA",
			build:   func(t testing.TB) *system.System { return newSystem(t, defSA()) },
			def:     defSA(),
			steps:   [letter.Count]int{2, 2, 1, 2, 2, 2, 1},
			natural: [letter.Count]int{3, 5, 7, 8, 10, 0, 2},
		},
		{
			name:    "S2",
			build:   func(t testing.TB) *system.System { return newSystem(t, defS2()) },
			def:     defS2(),
			steps:   [letter.Count]int{1, 0, 0, 0, 0, 0, 1},
			natural: [letter.Count]int{0, 1, 1, 1, 1, 1, 1},
		},
		{
			name:    "S14",
			build:   func(t testing.TB) *system.System { return newSystem(t, defS14()) },
			def:     defS14(),
			steps:   [letter.Count]int{2, 2, 2, 2, 2, 2, 2},
			natural: [letter.Count]int{0, 2, 4, 6, 8, 10, 12},
		},
		{
			name:    "irregular 12-class system",
			build:   func(t testing.TB) *system.System { return newSystem(t, defS12Irregular()) },
			def:     defS12Irregular(),
			steps:   [letter.Count]int{3, 1, 1, 2, 2, 2, 1},
			natural: [letter.Count]int{0, 3, 4, 5, 7, 9, 11},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			assert.True(t, s.IsValid())
			assert.Equal(t, tt.def.Name, s.Name())
			assert.Equal(t, tt.def.Classes, s.Classes())
			assert.Equal(t, tt.def.Boundary, s.Boundary())
			assert.Equal(t, tt.def.Chromatic, s.Chromatic())
			assert.Equal(t, tt.steps, s.Steps())

			def := s.Definition()
			require.NotNil(t, def.Alterations)
			assert.Same(t, s.Alterations(), def.Alterations, "Definition and Alterations share the rule")
			def.Alterations = nil
			want := tt.def
			want.Alterations = nil
			assert.Equal(t, want, def)

			sum := 0
			for _, st := range s.Steps() {
				sum += st
			}
			assert.Equal(t, s.Classes(), sum, "steps sum to N")

			for _, l := range letter.All() {
				i := l.Ordinal()

				off, err := s.Offset(l)
				require.NoError(t, err)
				assert.Equal(t, tt.def.Offsets[i], off, "Offset(%s)", l)
				assert.Equal(t, off, s.MustOffset(l))

				st, err := s.StepAfter(l)
				require.NoError(t, err)
				assert.Equal(t, tt.steps[i], st, "StepAfter(%s)", l)
				assert.Equal(t, st, s.MustStepAfter(l))

				nc, err := s.NaturalClass(l)
				require.NoError(t, err)
				assert.Equal(t, cls(tt.natural[i], s.Classes()), nc, "NaturalClass(%s)", l)
				assert.Equal(t, nc, s.MustNaturalClass(l))
			}

			for _, bad := range []letter.Letter{0, letter.Letter(8), letter.Letter(255)} {
				_, err := s.Offset(bad)
				requireErr(t, err, system.ErrInvalid, nil)
				_, err = s.StepAfter(bad)
				requireErr(t, err, system.ErrInvalid, nil)
				nc, err := s.NaturalClass(bad)
				requireErr(t, err, system.ErrInvalid, nil)
				assert.Equal(t, pitchclass.Class{}, nc)
				requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustOffset(bad) })
				requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustStepAfter(bad) })
				requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustNaturalClass(bad) })
			}
		})
	}
}

func TestSystem_Unset(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		s    *system.System
	}{
		{"nil", nil},
		{"zero value", &system.System{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.s
			assert.False(t, s.IsValid())
			assert.Equal(t, system.Definition{}, s.Definition())
			assert.Empty(t, s.Name())
			assert.Zero(t, s.Classes())
			assert.False(t, s.Boundary().IsValid())
			assert.Zero(t, s.Chromatic())
			assert.Nil(t, s.Alterations())
			assert.Equal(t, [letter.Count]int{}, s.Steps())
			assert.Equal(t, "System(unset)", s.String())

			_, err := s.Offset(letter.C)
			requireErr(t, err, system.ErrInvalid, nil)
			_, err = s.StepAfter(letter.C)
			requireErr(t, err, system.ErrInvalid, nil)
			_, err = s.NaturalClass(letter.C)
			requireErr(t, err, system.ErrInvalid, nil)

			r, err := s.Regularity()
			requireErr(t, err, system.ErrInvalid, nil)
			assert.Equal(t, system.Regularity{}, r)
			requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustRegularity() })

			_, err = s.HasTraditionalLayout()
			requireErr(t, err, system.ErrInvalid, nil)
			requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustHasTraditionalLayout() })

			_, err = s.NamesSubsetOf(system.Traditional())
			requireErr(t, err, system.ErrInvalid, nil)
			_, err = system.Traditional().NamesSubsetOf(s)
			requireErr(t, err, system.ErrInvalid, nil)
			requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustNamesSubsetOf(system.Traditional()) })

			assert.False(t, s.Equal(s))
			assert.False(t, s.Equal(system.Traditional()))
			assert.False(t, system.Traditional().Equal(s))
		})
	}
}

func TestSystem_Regularity(t *testing.T) {
	t.Parallel()

	regular := func(large, small int) system.Regularity {
		return system.Regularity{Regular: true, Large: large, Small: small, RegularByClasses: true}
	}

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		want  system.Regularity
	}{
		{name: "traditional (K-1)", build: func(testing.TB) *system.System { return system.Traditional() }, want: regular(2, 1)},
		{name: "19-position (K-1)", build: func(testing.TB) *system.System { return system.Positions19() }, want: regular(3, 2)},
		{name: "24-class (K-1)", build: func(testing.TB) *system.System { return system.QuarterTones24() }, want: regular(4, 2)},
		{name: "SA (K-2)", build: func(t testing.TB) *system.System { return newSystem(t, defSA()) }, want: regular(2, 1)},
		{name: "irregular 12-class system (K-2)", build: func(t testing.TB) *system.System { return newSystem(t, defS12Irregular()) }},
		{name: "S14 is not regular in any sense (K-2)", build: func(t testing.TB) *system.System { return newSystem(t, defS14()) }},
		{name: "S2", build: func(t testing.TB) *system.System { return newSystem(t, defS2()) }},
		{
			// Steps 3 3 1 3 3 3 1 with N = 17 and c = 1: steps after C, D, F, G, A agree,
			// after E and B they are c less mod N only if 1 = 3 - 1 mod 17, which is false.
			name: "17 classes with c = 1 is not regular",
			build: func(testing.TB) *system.System {
				return system.MustFromSteps(letter.C, [letter.Count]int{3, 3, 1, 3, 3, 3, 1}, 1, system.Integers(), "")
			},
		},
		{
			// Steps 3 3 1 3 3 3 1 with c = 2 = 3 - 1: regular.
			name: "17 classes with c = 2",
			build: func(testing.TB) *system.System {
				return system.MustFromSteps(letter.C, [letter.Count]int{3, 3, 1, 3, 3, 3, 1}, 2, system.Integers(), "")
			},
			want: regular(3, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Regularity()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustRegularity())
		})
	}
}

func TestSystem_HasTraditionalLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		want  bool
	}{
		{name: "traditional", build: func(testing.TB) *system.System { return system.Traditional() }, want: true},
		{name: "SA", build: func(t testing.TB) *system.System { return newSystem(t, defSA()) }, want: true},
		{name: "traditional layout with A = {0}", build: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.Single(a(0))))
		}, want: true},
		{name: "19-position", build: func(testing.TB) *system.System { return system.Positions19() }},
		{name: "24-class", build: func(testing.TB) *system.System { return system.QuarterTones24() }},
		{name: "S14", build: func(t testing.TB) *system.System { return newSystem(t, defS14()) }},
		{name: "irregular 12-class system", build: func(t testing.TB) *system.System { return newSystem(t, defS12Irregular()) }},
		{name: "traditional steps with c = 2", build: func(t testing.TB) *system.System {
			d := tradLayoutDef(system.Integers())
			d.Chromatic = 2
			return newSystem(t, d)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.HasTraditionalLayout()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustHasTraditionalLayout())
		})
	}
}

func TestSystem_NamesSubsetOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x, y func(t testing.TB) *system.System
		want bool
	}{
		{name: "traditional in 24-class (K-14)", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(testing.TB) *system.System { return system.QuarterTones24() }, want: true},
		{name: "24-class in traditional (K-14)", x: func(testing.TB) *system.System { return system.QuarterTones24() }, y: func(testing.TB) *system.System { return system.Traditional() }},
		{name: "traditional in 19-position (K-14)", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(testing.TB) *system.System { return system.Positions19() }, want: true},
		{name: "N, offsets and c do not matter", x: func(t testing.TB) *system.System { return newSystem(t, defS14()) }, y: func(t testing.TB) *system.System { return newSystem(t, defSA()) }, want: true},
		{name: "traditional in traditional layout with A from -2 to 2", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2)))
		}},
		{name: "A from -2 to 2 in traditional", x: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2)))
		}, y: func(testing.TB) *system.System { return system.Traditional() }, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, y := tt.x(t), tt.y(t)
			got, err := x.NamesSubsetOf(y)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, x.MustNamesSubsetOf(y))
		})
	}
}

func TestSystem_Equal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x, y func(t testing.TB) *system.System
		want bool
	}{
		{name: "same system", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(testing.TB) *system.System { return system.Traditional() }, want: true},
		{name: "another name (K-14)", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.Integers()))
		}, want: true},
		{name: "same rule written differently", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.MustIntegerRange(-maxAlt, maxAlt)))
		}, want: true},
		{name: "traditional and SA (K-14)", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System { return newSystem(t, defSA()) }},
		{name: "another rule", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2)))
		}},
		{name: "another chromatic unit", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System {
			d := tradLayoutDef(system.Integers())
			d.Chromatic = 2
			return newSystem(t, d)
		}},
		{name: "another offset", x: func(testing.TB) *system.System { return system.Traditional() }, y: func(t testing.TB) *system.System { return newSystem(t, defS12Irregular()) }},
		{name: "another N", x: func(t testing.TB) *system.System { return newSystem(t, defS14()) }, y: func(t testing.TB) *system.System {
			d := defS14()
			d.Classes = 15
			return newSystem(t, d)
		}},
		{name: "another boundary letter with the same offsets", x: func(t testing.TB) *system.System {
			// N = 1, all letters on 0: both C and D can be the boundary letter.
			return newSystem(t, system.Definition{Classes: 1, Boundary: letter.C, Chromatic: 1, Alterations: system.Integers()})
		}, y: func(t testing.TB) *system.System {
			return newSystem(t, system.Definition{Classes: 1, Boundary: letter.D, Chromatic: 1, Alterations: system.Integers()})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, y := tt.x(t), tt.y(t)
			assert.Equal(t, tt.want, x.Equal(y))
			assert.Equal(t, tt.want, y.Equal(x))
			assert.True(t, x.Equal(x))
		})
	}
}

func TestSystem_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		want  string
	}{
		{
			name:  "traditional",
			build: func(testing.TB) *system.System { return system.Traditional() },
			want:  `System("traditional"; N=12; b=C; C=0 D=2 E=4 F=5 G=7 A=9 B=11; c=1; A=0 + Z/1 in [-inf, +inf])`,
		},
		{
			name:  "24-class",
			build: func(testing.TB) *system.System { return system.QuarterTones24() },
			want:  `System("24-quarter-tone"; N=24; b=C; C=0 D=4 E=8 F=10 G=14 A=18 B=22; c=2; A=0 + Z/2 in [-inf, +inf])`,
		},
		{
			name:  "SA",
			build: func(t testing.TB) *system.System { return newSystem(t, defSA()) },
			want:  `System("SA"; N=12; b=A; C=3 D=5 E=7 F=8 G=10 A=0 B=2; c=1; A=0 + Z/1 in [-inf, +inf])`,
		},
		{
			name: "quoted name and union rule",
			build: func(t testing.TB) *system.System {
				d := tradLayoutDef(system.MustUnion(system.MustIntegerRange(-2, 2), system.Single(q(1, 2))))
				d.Name = "a \"b\""
				d.Chromatic = 2
				d.Classes = 24
				d.Offsets = offs(0, 4, 8, 10, 14, 18, 22)
				return newSystem(t, d)
			},
			want: `System("a \"b\""; N=24; b=C; C=0 D=4 E=8 F=10 G=14 A=18 B=22; c=2; A=0 + Z/1 in [-2, +2] | {+1/2})`,
		},
		{
			name: "empty name",
			build: func(t testing.TB) *system.System {
				d := tradLayoutDef(system.Single(a(0)))
				d.Name = ""
				return newSystem(t, d)
			},
			want: `System(""; N=12; b=C; C=0 D=2 E=4 F=5 G=7 A=9 B=11; c=1; A={0})`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.build(t).String())
		})
	}
}
