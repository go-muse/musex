package system_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for names in a system: OP-16, OP-18, OP-20 (whole n), OP-27, OP-29, OP-30, OP-40,
// P-4…P-6, P-9, contract 5.7, criteria K-3, K-4, K-5 (S19), K-9 (names), K-10, K-17,
// K-26, K-27, K-28 (without S5).

// integerModels are systems with A = Z for model-based checks.
func integerModels() []model {
	return []model{modelT(), modelS19(), modelSA(), modelS14(), modelS12Irregular(), modelS2()}
}

// halfNames returns names with alteration values in Z/2 from -limit to limit.
func halfNames(limit int) []spelling.Spelling {
	out := make([]spelling.Spelling, 0, letter.Count*(4*limit+1))
	for _, l := range letter.All() {
		for n := -2 * limit; n <= 2*limit; n++ {
			out = append(out, spq(l, n, 2))
		}
	}
	return out
}

func TestSystem_Admits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		in    []spelling.Spelling
		out   []spelling.Spelling
	}{
		{
			name:  "traditional (K-4)",
			build: func(testing.TB) *system.System { return system.Traditional() },
			in:    append(namesUpTo(3), spi(letter.C, maxAlt), spi(letter.B, -maxAlt)),
			out:   []spelling.Spelling{dHalfSharp(), spq(letter.C, -1, 2), spq(letter.E, 1, 3)},
		},
		{
			name:  "24-class (K-4)",
			build: func(testing.TB) *system.System { return system.QuarterTones24() },
			in:    append(halfNames(2), dHalfSharp()),
			out:   []spelling.Spelling{spq(letter.C, 1, 4), spq(letter.E, 1, 3)},
		},
		{
			name:  "traditional layout with A from -2 to 2",
			build: func(t testing.TB) *system.System { return newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2))) },
			in:    namesUpTo(2),
			out:   []spelling.Spelling{spi(letter.C, 3), spi(letter.B, -3), dHalfSharp()},
		},
		{
			name:  "traditional layout with A = {0}",
			build: func(t testing.TB) *system.System { return newSystem(t, tradLayoutDef(system.Single(a(0)))) },
			in:    namesUpTo(0),
			out:   []spelling.Spelling{spelling.CSharp(), spelling.BFlat()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			for _, n := range tt.in {
				got, err := s.Admits(n)
				require.NoError(t, err)
				assert.True(t, got, "%s must be admissible", n)
				assert.True(t, s.MustAdmits(n))
			}
			for _, n := range tt.out {
				got, err := s.Admits(n)
				require.NoError(t, err)
				assert.False(t, got, "%s must not be admissible", n)
				assert.False(t, s.MustAdmits(n))
			}
		})
	}
}

func TestSystem_ClassExamples(t *testing.T) {
	t.Parallel()

	type check struct {
		name  spelling.Spelling
		class int
	}
	tests := []struct {
		name   string
		build  func(t testing.TB) *system.System
		n      int
		checks []check
	}{
		{
			name:  "traditional (K-3)",
			build: func(testing.TB) *system.System { return system.Traditional() },
			n:     12,
			checks: []check{
				{spelling.BSharp(), 0}, {spelling.CFlat(), 11},
				{spelling.D(), 2}, {spelling.CDoubleSharp(), 2}, {spelling.EDoubleFlat(), 2},
				// The 24 names of the OMT list.
				{spelling.C(), 0}, {spelling.BSharp(), 0},
				{spelling.CSharp(), 1}, {spelling.DFlat(), 1},
				{spelling.D(), 2}, {spelling.CDoubleSharp(), 2},
				{spelling.DSharp(), 3}, {spelling.EFlat(), 3},
				{spelling.E(), 4}, {spelling.FFlat(), 4},
				{spelling.F(), 5}, {spelling.ESharp(), 5},
				{spelling.FSharp(), 6}, {spelling.GFlat(), 6},
				{spelling.G(), 7}, {spelling.FDoubleSharp(), 7},
				{spelling.GSharp(), 8}, {spelling.AFlat(), 8},
				{spelling.A(), 9}, {spelling.GDoubleSharp(), 9},
				{spelling.ASharp(), 10}, {spelling.BFlat(), 10},
				{spelling.B(), 11}, {spelling.CFlat(), 11},
				// Large alteration values.
				{spi(letter.C, maxAlt), 4}, {spi(letter.C, -maxAlt), 8},
			},
		},
		{
			name:  "19-position (K-3)",
			build: func(testing.TB) *system.System { return system.Positions19() },
			n:     19,
			checks: []check{
				{spelling.CSharp(), 1}, {spelling.DFlat(), 2},
				{spelling.ESharp(), 7}, {spelling.FFlat(), 7},
				{spelling.BSharp(), 18}, {spelling.CFlat(), 18},
			},
		},
		{
			name:  "24-class: the C major triad (K-3)",
			build: func(testing.TB) *system.System { return system.QuarterTones24() },
			n:     24,
			checks: []check{
				{spelling.C(), 0}, {spelling.E(), 8}, {spelling.G(), 14},
				{dHalfSharp(), 5}, {spq(letter.C, -1, 2), 23},
			},
		},
		{
			name:  "SA: C gets class 3 (K-2)",
			build: func(t testing.TB) *system.System { return newSystem(t, defSA()) },
			n:     12,
			checks: []check{
				{spelling.C(), 3}, {spelling.A(), 0}, {spelling.AFlat(), 11}, {spelling.GSharp(), 11},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			for _, c := range tt.checks {
				got, err := s.Class(c.name)
				require.NoError(t, err, "Class(%s)", c.name)
				assert.Equal(t, cls(c.class, tt.n), got, "Class(%s)", c.name)
				assert.Equal(t, tt.n, got.Modulus())
				assert.Equal(t, got, s.MustClass(c.name))
			}
		})
	}
}

func TestSystem_ClassAndCarryModel(t *testing.T) {
	t.Parallel()

	models := append(integerModels(), modelS24())
	for _, m := range models {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()

			s := m.build(t)
			names := namesUpTo(3)
			if m.c == 2 {
				names = halfNames(3)
			}
			for _, n := range names {
				c, err := s.Class(n)
				require.NoError(t, err, "Class(%s)", n)
				assert.Equal(t, cls(m.class(t, n), m.n), c, "Class(%s)", n)

				k, err := s.Carry(n)
				require.NoError(t, err, "Carry(%s)", n)
				assert.Equal(t, m.carry(t, n), k, "Carry(%s)", n)
				assert.Equal(t, k, s.MustCarry(n))
			}
		})
	}
}

func TestSystem_Carry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		input spelling.Spelling
		want  int
	}{
		{name: "traditional C (K-28)", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.C(), want: 0},
		{name: "traditional B sharp (K-28)", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.BSharp(), want: 1},
		{name: "traditional C flat (K-28)", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.CFlat(), want: -1},
		{name: "traditional B with 13 sharps (K-28)", build: func(testing.TB) *system.System { return system.Traditional() }, input: spi(letter.B, 13), want: 2},
		{name: "traditional B double sharp", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.BDoubleSharp(), want: 1},
		{name: "traditional C double flat", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.CDoubleFlat(), want: -1},
		{name: "traditional C with 13 flats", build: func(testing.TB) *system.System { return system.Traditional() }, input: spi(letter.C, -13), want: -2},
		{name: "traditional C+M", build: func(testing.TB) *system.System { return system.Traditional() }, input: spi(letter.C, maxAlt), want: 1398101},
		{name: "traditional C-M", build: func(testing.TB) *system.System { return system.Traditional() }, input: spi(letter.C, -maxAlt), want: -1398102},
		{name: "19-position B sharp (K-28)", build: func(testing.TB) *system.System { return system.Positions19() }, input: spelling.BSharp(), want: 0},
		{name: "SA A flat", build: func(t testing.TB) *system.System { return newSystem(t, defSA()) }, input: spelling.AFlat(), want: -1},
		{name: "SA G sharp", build: func(t testing.TB) *system.System { return newSystem(t, defSA()) }, input: spelling.GSharp(), want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.build(t).Carry(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("traditional, |a| <= 2: only B sharps and C flats carry", func(t *testing.T) {
		t.Parallel()

		s := system.Traditional()
		carrying := map[spelling.Spelling]int{
			spelling.BSharp(): 1, spelling.BDoubleSharp(): 1,
			spelling.CFlat(): -1, spelling.CDoubleFlat(): -1,
		}
		for _, n := range namesUpTo(2) {
			assert.Equal(t, carrying[n], s.MustCarry(n), "Carry(%s)", n)
		}
	})
}

func TestSystem_LargeNumbers(t *testing.T) {
	t.Parallel()

	// K-27: N = 2^16, C0 D1 E2 F3 G4 A5 B6, c = 2^16, A = Z.
	s := newSystem(t, system.Definition{
		Classes: 1 << 16, Boundary: letter.C, Offsets: offs(0, 1, 2, 3, 4, 5, 6),
		Chromatic: 1 << 16, Alterations: system.Integers(),
	})

	top := spi(letter.C, maxAlt)
	assert.Equal(t, cls(0, 1<<16), s.MustClass(top))
	assert.Equal(t, maxAlt, s.MustCarry(top))
	p, err := s.FromWritten(top, o(4))
	require.NoError(t, err)
	assert.Equal(t, 4+maxAlt, p.Octave().MustASPN())
	assert.Equal(t, 4, s.MustWrittenOctave(p).MustASPN())

	bottom := spi(letter.B, -maxAlt)
	assert.Equal(t, cls(6, 1<<16), s.MustClass(bottom))
	assert.Equal(t, -maxAlt, s.MustCarry(bottom))

	t.Run("steps 1/MaxDenominator with c = MaxChromatic", func(t *testing.T) {
		t.Parallel()

		f := newSystem(t, system.Definition{
			Classes: system.MaxClasses, Boundary: letter.C, Offsets: offs(0, 1, 2, 3, 4, 5, 6),
			Chromatic: system.MaxChromatic, Alterations: system.MustMultiples(system.MaxDenominator),
		})
		tiny := spq(letter.C, 1, system.MaxDenominator)
		assert.True(t, f.MustAdmits(tiny))
		assert.Equal(t, cls(1, system.MaxClasses), f.MustClass(tiny))
		assert.Equal(t, cls(system.MaxClasses-1, system.MaxClasses), f.MustClass(spq(letter.C, -1, system.MaxDenominator)))
		assert.Equal(t, -1, f.MustCarry(spq(letter.C, -1, system.MaxDenominator)))
	})
}

func TestSystem_ShiftedClass(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, cls(2, 19), system.Positions19().MustShiftedClass(spelling.C(), 2), "K-5: C, n = 2 in S19 is the class of D flat")
		assert.Equal(t, system.Positions19().MustClass(spelling.DFlat()), system.Positions19().MustShiftedClass(spelling.C(), 2))
		assert.Equal(t, cls(0, 12), system.Traditional().MustShiftedClass(spelling.CSharp(), -1))
		assert.Equal(t, cls(10, 24), system.QuarterTones24().MustShiftedClass(spelling.E(), 1))
		assert.Equal(t, cls(4, 12), system.Traditional().MustShiftedClass(spelling.C(), maxAlt))
		assert.Equal(t, cls(8, 12), system.Traditional().MustShiftedClass(spelling.C(), -maxAlt))
		assert.Equal(t, cls(0, 12), system.Traditional().MustShiftedClass(spelling.BSharp(), 0))
	})

	t.Run("model: (κ + n·c) mod N", func(t *testing.T) {
		t.Parallel()

		for _, m := range append(integerModels(), modelS24()) {
			s := m.build(t)
			for _, n := range namesUpTo(2) {
				for shift := -30; shift <= 30; shift++ {
					want := m.class(t, n) + shift*m.c
					want = want - floorDiv(want, m.n)*m.n
					got, err := s.ShiftedClass(n, shift)
					require.NoError(t, err, "%s: ShiftedClass(%s, %d)", m.name, n, shift)
					assert.Equal(t, cls(want, m.n), got, "%s: ShiftedClass(%s, %d)", m.name, n, shift)
				}
			}
		}
	})

	limited := func(t testing.TB) *system.System { return newSystem(t, tradLayoutDef(system.MustIntegerRange(-2, 2))) }

	tests := []struct {
		name     string
		build    func(t testing.TB) *system.System
		input    spelling.Spelling
		n        int
		category error
	}{
		{name: "unset system", build: func(testing.TB) *system.System { return nil }, input: spelling.C(), n: 1, category: system.ErrInvalid},
		{name: "unset name", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.Spelling{}, n: 1, category: system.ErrInvalid},
		{name: "unset system and inadmissible name", build: func(testing.TB) *system.System { return &system.System{} }, input: dHalfSharp(), n: math.MaxInt, category: system.ErrInvalid},
		{name: "inadmissible source (K-17)", build: func(testing.TB) *system.System { return system.Traditional() }, input: dHalfSharp(), n: 1, category: system.ErrInadmissible},
		{name: "inadmissible source before range", build: func(testing.TB) *system.System { return system.Traditional() }, input: dHalfSharp(), n: math.MaxInt, category: system.ErrInadmissible},
		{name: "a + n above the alteration range", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.C(), n: maxAlt + 1, category: system.ErrOutOfRange},
		{name: "a + n below the alteration range", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.C(), n: -maxAlt - 1, category: system.ErrOutOfRange},
		{name: "n = MaxInt", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.CSharp(), n: math.MaxInt, category: system.ErrOutOfRange},
		{name: "n = MinInt", build: func(testing.TB) *system.System { return system.Traditional() }, input: spelling.CFlat(), n: math.MinInt, category: system.ErrOutOfRange},
		{name: "a = M, n = 1", build: func(testing.TB) *system.System { return system.Traditional() }, input: spi(letter.C, maxAlt), n: 1, category: system.ErrOutOfRange},
		{name: "a + n not admissible", build: limited, input: spelling.CDoubleSharp(), n: 1, category: system.ErrInadmissible},
		{name: "range before admissibility of a + n", build: limited, input: spelling.CDoubleSharp(), n: math.MaxInt, category: system.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.ShiftedClass(tt.input, tt.n)
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, pitchclass.Class{}, got)
			requirePanicErr(t, tt.category, nil, func() { s.MustShiftedClass(tt.input, tt.n) })
		})
	}
}

func TestSystem_EquivalentNames(t *testing.T) {
	t.Parallel()

	t.Run("examples (K-9)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			s     *system.System
			x, y  spelling.Spelling
			equiv bool
		}{
			{"T: C sharp, D flat", system.Traditional(), spelling.CSharp(), spelling.DFlat(), true},
			{"T: E sharp, F", system.Traditional(), spelling.ESharp(), spelling.F(), true},
			{"T: B sharp, C flat", system.Traditional(), spelling.BSharp(), spelling.CFlat(), false},
			{"T: C sharp, B double sharp", system.Traditional(), spelling.CSharp(), spelling.BDoubleSharp(), true},
			{"T: B flat, C double flat", system.Traditional(), spelling.BFlat(), spelling.CDoubleFlat(), true},
			{"S19: C sharp, D flat", system.Positions19(), spelling.CSharp(), spelling.DFlat(), false},
			{"S19: E sharp, F", system.Positions19(), spelling.ESharp(), spelling.F(), false},
			{"S19: B sharp, C flat", system.Positions19(), spelling.BSharp(), spelling.CFlat(), true},
			{"S19: E sharp, F flat", system.Positions19(), spelling.ESharp(), spelling.FFlat(), true},
		}
		for _, tt := range tests {
			got, err := tt.s.EquivalentNames(tt.x, tt.y)
			require.NoError(t, err, tt.name)
			assert.Equal(t, tt.equiv, got, tt.name)
			assert.Equal(t, tt.equiv, tt.s.MustEquivalentNames(tt.y, tt.x), tt.name)
		}
	})

	t.Run("equivalence relation on 35 names of T (K-9)", func(t *testing.T) {
		t.Parallel()

		s := system.Traditional()
		m := modelT()
		names := namesUpTo(2)
		require.Len(t, names, 35)
		eq := func(x, y spelling.Spelling) bool {
			got, err := s.EquivalentNames(x, y)
			require.NoError(t, err)
			return got
		}
		for _, x := range names {
			assert.True(t, eq(x, x), "reflexive: %s", x)
			for _, y := range names {
				exy := eq(x, y)
				assert.Equal(t, m.class(t, x) == m.class(t, y), exy, "%s ≡ %s", x, y)
				assert.Equal(t, exy, eq(y, x), "symmetric: %s, %s", x, y)
				if !exy {
					continue
				}
				for _, z := range names {
					if eq(y, z) {
						assert.True(t, eq(x, z), "transitive: %s, %s, %s", x, y, z)
					}
				}
			}
		}
	})

	tests := []struct {
		name     string
		s        *system.System
		x, y     spelling.Spelling
		category error
	}{
		{"unset system", nil, spelling.C(), spelling.C(), system.ErrInvalid},
		{"unset first name", system.Traditional(), spelling.Spelling{}, spelling.C(), system.ErrInvalid},
		{"unset second name", system.Traditional(), spelling.C(), spelling.Spelling{}, system.ErrInvalid},
		{"inadmissible first and unset second", system.Traditional(), dHalfSharp(), spelling.Spelling{}, system.ErrInvalid},
		{"inadmissible first (K-17)", system.Traditional(), dHalfSharp(), spelling.C(), system.ErrInadmissible},
		{"inadmissible second (K-17)", system.Traditional(), spelling.C(), dHalfSharp(), system.ErrInadmissible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.s.EquivalentNames(tt.x, tt.y)
			requireErr(t, err, tt.category, nil)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustEquivalentNames(tt.x, tt.y) })
		})
	}
}

func TestSystem_Spellings(t *testing.T) {
	t.Parallel()

	sp := func(names ...spelling.Spelling) []spelling.Spelling { return names }

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		class pitchclass.Class
		limit alteration.Alteration
		want  []spelling.Spelling
	}{
		{name: "T class 3, L = 2 (K-10)", build: func(testing.TB) *system.System { return system.Traditional() },
			class: cls(3, 12), limit: a(2), want: sp(spelling.DSharp(), spelling.EFlat(), spelling.FDoubleFlat())},
		{name: "T class 8, L = 2 (K-10)", build: func(testing.TB) *system.System { return system.Traditional() },
			class: cls(8, 12), limit: a(2), want: sp(spelling.GSharp(), spelling.AFlat())},
		{name: "T class 0, L = 2: letter order, not pitch order", build: func(testing.TB) *system.System { return system.Traditional() },
			class: cls(0, 12), limit: a(2), want: sp(spelling.C(), spelling.DDoubleFlat(), spelling.BSharp())},
		{name: "T class 0, L = 0: natural names only (K-26)", build: func(testing.TB) *system.System { return system.Traditional() },
			class: cls(0, 12), limit: a(0), want: sp(spelling.C())},
		{name: "T class 1, L = 0: empty (K-26)", build: func(testing.TB) *system.System { return system.Traditional() },
			class: cls(1, 12), limit: a(0), want: sp()},
		{name: "T class 1, fractional limit 3/2", build: func(testing.TB) *system.System { return system.Traditional() },
			class: cls(1, 12), limit: q(3, 2), want: sp(spelling.CSharp(), spelling.DFlat())},
		{name: "traditional layout with A = {0}, class 1: empty", build: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.Single(a(0))))
		}, class: cls(1, 12), limit: a(2), want: sp()},
		{name: "overlapping progressions give each name once", build: func(t testing.TB) *system.System {
			return newSystem(t, tradLayoutDef(system.MustUnion(system.Integers(), system.MustIntegerRange(-2, 2))))
		}, class: cls(3, 12), limit: a(2), want: sp(spelling.DSharp(), spelling.EFlat(), spelling.FDoubleFlat())},
		{name: "S24 class 1, L = 1", build: func(testing.TB) *system.System { return system.QuarterTones24() },
			class: cls(1, 24), limit: a(1), want: sp(spq(letter.C, 1, 2))},
		{name: "S2 class 1, L = 0: six natural letters", build: func(t testing.TB) *system.System { return newSystem(t, defS2()) },
			class: cls(1, 2), limit: a(0), want: sp(spelling.D(), spelling.E(), spelling.F(), spelling.G(), spelling.A(), spelling.B())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			seq, err := s.Spellings(tt.class, tt.limit)
			require.NoError(t, err)
			if len(tt.want) == 0 {
				assert.Empty(t, collect(t, seq))
				assert.Empty(t, collect(t, s.MustSpellings(tt.class, tt.limit)))
				return
			}
			assert.Equal(t, tt.want, collect(t, seq))
			assert.Equal(t, tt.want, collect(t, seq), "repeated traversal")
			assert.Equal(t, tt.want, collect(t, s.MustSpellings(tt.class, tt.limit)))
			assert.Equal(t, tt.want[:1], first(t, seq, 1), "early stop")
		})
	}

	t.Run("every class of T has three names at L = 2, except class 8 (K-10)", func(t *testing.T) {
		t.Parallel()

		s := system.Traditional()
		total := 0
		for c := range 12 {
			got := collect(t, s.MustSpellings(cls(c, 12), a(2)))
			want := 3
			if c == 8 {
				want = 2
			}
			assert.Len(t, got, want, "class %d: %v", c, got)
			total += len(got)
		}
		assert.Equal(t, 35, total)
	})

	t.Run("completeness against Values × letters", func(t *testing.T) {
		t.Parallel()

		for _, m := range append(integerModels(), modelS24()) {
			s := m.build(t)
			for _, limit := range []alteration.Alteration{a(0), q(1, 2), a(1), a(2), a(3)} {
				values := collect(t, s.Alterations().MustValues(limit))
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
					got := collect(t, s.MustSpellings(cls(c, m.n), limit))
					if want == nil {
						assert.Empty(t, got, "%s class %d limit %s", m.name, c, limit)
						continue
					}
					assert.Equal(t, want, got, "%s class %d limit %s", m.name, c, limit)
					for i := 1; i < len(got); i++ {
						assert.Equal(t, -1, got[i-1].Compare(got[i]), "order of %s and %s", got[i-1], got[i])
					}
				}
			}
		}
	})

	errTests := []struct {
		name     string
		s        *system.System
		class    pitchclass.Class
		limit    alteration.Alteration
		category error
	}{
		{"unset system", nil, cls(0, 12), a(1), system.ErrInvalid},
		{"unset class", system.Traditional(), pitchclass.Class{}, a(1), system.ErrInvalid},
		{"unset class and negative limit", system.Traditional(), pitchclass.Class{}, a(-1), system.ErrInvalid},
		{"class of another system (K-19)", system.Traditional(), cls(0, 19), a(1), system.ErrModulusMismatch},
		{"class of another system and negative limit", system.Traditional(), cls(0, 19), a(-1), system.ErrModulusMismatch},
		{"negative limit (K-26)", system.Traditional(), cls(0, 12), a(-1), system.ErrLimit},
		{"negative fractional limit", system.Traditional(), cls(0, 12), q(-1, 2), system.ErrLimit},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := tt.s.Spellings(tt.class, tt.limit)
			requireErr(t, err, tt.category, nil)
			assert.Nil(t, seq)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustSpellings(tt.class, tt.limit) })
		})
	}
}

func TestSystem_EnharmonicNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     *system.System
		input spelling.Spelling
		limit alteration.Alteration
		want  []spelling.Spelling
	}{
		{"C in T, L = 2 (K-10)", system.Traditional(), spelling.C(), a(2), []spelling.Spelling{spelling.C(), spelling.DDoubleFlat(), spelling.BSharp()}},
		{"C in T, L = 0 (K-10)", system.Traditional(), spelling.C(), a(0), []spelling.Spelling{spelling.C()}},
		{"D sharp in T, L = 0: not even itself", system.Traditional(), spelling.DSharp(), a(0), []spelling.Spelling{}},
		{"B sharp in T, L = 1", system.Traditional(), spelling.BSharp(), a(1), []spelling.Spelling{spelling.C(), spelling.BSharp()}},
		{"G sharp in T, L = 2", system.Traditional(), spelling.GSharp(), a(2), []spelling.Spelling{spelling.GSharp(), spelling.AFlat()}},
		{"E sharp in S19, L = 1", system.Positions19(), spelling.ESharp(), a(1), []spelling.Spelling{spelling.ESharp(), spelling.FFlat()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := tt.s.EnharmonicNames(tt.input, tt.limit)
			require.NoError(t, err)
			got := collect(t, seq)
			if len(tt.want) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
			assert.Equal(t, got, collect(t, tt.s.MustEnharmonicNames(tt.input, tt.limit)))
		})
	}

	errTests := []struct {
		name     string
		s        *system.System
		input    spelling.Spelling
		limit    alteration.Alteration
		category error
	}{
		{"unset system", nil, spelling.C(), a(1), system.ErrInvalid},
		{"unset name", system.Traditional(), spelling.Spelling{}, a(1), system.ErrInvalid},
		{"unset name and negative limit", system.Traditional(), spelling.Spelling{}, a(-1), system.ErrInvalid},
		{"negative limit", system.Traditional(), spelling.C(), a(-1), system.ErrLimit},
		{"negative limit before inadmissible name", system.Traditional(), dHalfSharp(), a(-1), system.ErrLimit},
		{"inadmissible name (K-17)", system.Traditional(), dHalfSharp(), a(1), system.ErrInadmissible},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := tt.s.EnharmonicNames(tt.input, tt.limit)
			requireErr(t, err, tt.category, nil)
			assert.Nil(t, seq)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustEnharmonicNames(tt.input, tt.limit) })
		})
	}
}

func TestSystem_NameErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        *system.System
		input    spelling.Spelling
		category error
	}{
		{"unset system", nil, spelling.C(), system.ErrInvalid},
		{"zero system", &system.System{}, spelling.C(), system.ErrInvalid},
		{"unset name", system.Traditional(), spelling.Spelling{}, system.ErrInvalid},
		{"inadmissible name (K-17)", system.Traditional(), dHalfSharp(), system.ErrInadmissible},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, n := tt.s, tt.input
			if tt.category == system.ErrInvalid {
				_, err := s.Admits(n)
				requireErr(t, err, system.ErrInvalid, nil)
				requirePanicErr(t, system.ErrInvalid, nil, func() { s.MustAdmits(n) })
			}

			c, err := s.Class(n)
			requireErr(t, err, tt.category, nil)
			assert.Equal(t, pitchclass.Class{}, c)
			requirePanicErr(t, tt.category, nil, func() { s.MustClass(n) })

			k, err := s.Carry(n)
			requireErr(t, err, tt.category, nil)
			assert.Zero(t, k)
			requirePanicErr(t, tt.category, nil, func() { s.MustCarry(n) })
		})
	}
}
