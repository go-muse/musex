package system_test

import (
	"cmp"
	"math"
	"slices"
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

// Tests for pitches in a system: OP-17, OP-19, OP-23…OP-26, OP-28, OP-31, P-8, P-9,
// contract 5.8 and 6.2, criteria K-4, K-7, K-8, K-9 (pitches), K-10 (OP-31), K-17,
// K-22, K-25.

// samplePitches returns names with |a| <= 2 in octaves from -2 to 5.
func samplePitches() []pitch.Pitch {
	names := namesUpTo(2)
	out := make([]pitch.Pitch, 0, len(names)*8)
	for oct := -2; oct <= 5; oct++ {
		for _, n := range names {
			out = append(out, pt(n, oct))
		}
	}
	return out
}

func TestSystem_AdmitsPitch(t *testing.T) {
	t.Parallel()

	half := pt(dHalfSharp(), 4)
	assert.False(t, system.Traditional().MustAdmitsPitch(half), "K-4")
	assert.True(t, system.QuarterTones24().MustAdmitsPitch(half))
	got, err := system.Traditional().AdmitsPitch(pt(spelling.BSharp(), math.MaxInt))
	require.NoError(t, err)
	assert.True(t, got)

	_, err = system.Traditional().AdmitsPitch(pitch.Pitch{})
	requireErr(t, err, system.ErrInvalid, nil)
	var unset *system.System
	_, err = unset.AdmitsPitch(half)
	requireErr(t, err, system.ErrInvalid, nil)
	requirePanicErr(t, system.ErrInvalid, nil, func() { unset.MustAdmitsPitch(half) })
}

func TestSystem_PitchClass(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	assert.Equal(t, cls(0, 12), s.MustPitchClass(pt(spelling.BSharp(), 4)))
	assert.Equal(t, cls(11, 12), s.MustPitchClass(pt(spelling.CFlat(), 3)))
	assert.Equal(t, cls(0, 12), s.MustPitchClass(pt(spelling.C(), math.MinInt)))

	for _, m := range append(integerModels(), modelS24()) {
		sys := m.build(t)
		for _, p := range samplePitches() {
			got, err := sys.PitchClass(p)
			require.NoError(t, err)
			assert.Equal(t, cls(m.class(t, p.Spelling()), m.n), got, "%s: PitchClass(%s)", m.name, p)
		}
	}

	c, err := s.PitchClass(pt(dHalfSharp(), 4))
	requireErr(t, err, system.ErrInadmissible, nil)
	assert.Equal(t, pitchclass.Class{}, c)
	_, err = s.PitchClass(pitch.Pitch{})
	requireErr(t, err, system.ErrInvalid, nil)
	requirePanicErr(t, system.ErrInadmissible, nil, func() { s.MustPitchClass(pt(dHalfSharp(), 4)) })
}

func TestSystem_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		x, y  pitch.Pitch
		want  int
	}{
		{"T: B sharp 3 and C4 (K-7)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.BSharp(), 4), pt(spelling.C(), 4), 0},
		{"T: C3 below A3 (K-7)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.C(), 3), pt(spelling.A(), 3), -1},
		{"SA: C3 above A3 (K-7)", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, pt(spelling.C(), 3), pt(spelling.A(), 3), 1},
		{"T: C flat 4 below C4", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.CFlat(), 3), pt(spelling.C(), 4), -1},
		{"T: edges of the octave axis", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.C(), math.MinInt), pt(spelling.C(), math.MaxInt), -1},
		{"T: C, D at Octave MAX (K-22)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.C(), math.MaxInt), pt(spelling.D(), math.MaxInt), -1},
		{"T: B and C flat at Octave MIN are one position", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.B(), math.MinInt), pt(spelling.CFlat(), math.MinInt), 0},
		{"S19: B sharp and C flat at one Octave", func(testing.TB) *system.System { return system.Positions19() }, pt(spelling.BSharp(), 5), pt(spelling.CFlat(), 5), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Compare(tt.x, tt.y)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, -tt.want, s.MustCompare(tt.y, tt.x))
		})
	}
}

func TestSystem_Between(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	bs3 := pt(spelling.BSharp(), 4)
	c4 := pt(spelling.C(), 4)
	d4 := pt(spelling.D(), 4)
	e2 := pt(spelling.E(), 2)

	tests := []struct {
		name    string
		x, a, b pitch.Pitch
		want    bool
	}{
		{"C4 in [B sharp 3, C4] (K-7)", c4, bs3, c4, true},
		{"C4 in [C4, B sharp 3] (K-7)", c4, c4, bs3, true},
		{"D4 in [E2, C4] (K-7)", d4, e2, c4, false},
		{"D4 in [C4, E2]", d4, c4, e2, false},
		{"lower end is included", e2, e2, c4, true},
		{"upper end is included", c4, c4, e2, true},
		{"inside", pt(spelling.A(), 3), e2, c4, true},
		{"below", pt(spelling.D(), 2), e2, c4, false},
		{"degenerate segment", pt(spelling.DDoubleFlat(), 4), c4, bs3, true},
		{"edges of the axis", c4, pt(spelling.C(), math.MinInt), pt(spelling.C(), math.MaxInt), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := s.Between(tt.x, tt.a, tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustBetween(tt.x, tt.a, tt.b))
		})
	}
}

func TestSystem_Sorted(t *testing.T) {
	t.Parallel()

	s := system.Traditional()

	t.Run("example (K-7)", func(t *testing.T) {
		t.Parallel()

		c4 := pt(spelling.C(), 4)
		bs3 := pt(spelling.BSharp(), 4)
		cf4 := pt(spelling.CFlat(), 3)
		b3 := pt(spelling.B(), 3)
		df4 := pt(spelling.DFlat(), 4)
		cs4 := pt(spelling.CSharp(), 4)

		input := []pitch.Pitch{c4, bs3, cf4, b3, df4, cs4}
		before := slices.Clone(input)

		got, err := s.Sorted(input)
		require.NoError(t, err)
		assert.Equal(t, []pitch.Pitch{cf4, b3, c4, bs3, df4, cs4}, got)
		assert.Equal(t, before, input, "the input is not changed")

		got[0] = pt(spelling.G(), 9)
		assert.Equal(t, before, input, "the result does not share the input")
		assert.Equal(t, []pitch.Pitch{cf4, b3, c4, bs3, df4, cs4}, s.MustSorted(input))
	})

	t.Run("stable sort by position", func(t *testing.T) {
		t.Parallel()

		for _, m := range []model{modelT(), modelS19(), modelSA()} {
			sys := m.build(t)
			input := samplePitches()
			// A deterministic shuffle keeps duplicates of positions in a nontrivial order.
			slices.SortFunc(input, func(x, y pitch.Pitch) int {
				return cmp.Compare((x.Spelling().Letter().Ordinal()*7+x.Octave().MustASPN()*3)%11,
					(y.Spelling().Letter().Ordinal()*7+y.Octave().MustASPN()*3)%11)
			})
			want := slices.Clone(input)
			slices.SortStableFunc(want, func(x, y pitch.Pitch) int {
				return cmp.Compare(m.position(t, x), m.position(t, y))
			})
			got, err := sys.Sorted(input)
			require.NoError(t, err)
			assert.Equal(t, want, got, m.name)
		}
	})

	t.Run("empty and single", func(t *testing.T) {
		t.Parallel()

		got, err := s.Sorted(nil)
		require.NoError(t, err)
		assert.Empty(t, got)
		got, err = s.Sorted([]pitch.Pitch{})
		require.NoError(t, err)
		assert.Empty(t, got)
		one := []pitch.Pitch{pt(spelling.E(), 1)}
		assert.Equal(t, one, s.MustSorted(one))
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		valid := pt(spelling.C(), 4)
		inadmissible := pt(dHalfSharp(), 4)

		_, err := s.Sorted([]pitch.Pitch{valid, inadmissible, {}})
		requireErr(t, err, system.ErrInvalid, nil) // E-1 over all elements before E-6
		_, err = s.Sorted([]pitch.Pitch{valid, inadmissible})
		requireErr(t, err, system.ErrInadmissible, nil)
		var unset *system.System
		res, err := unset.Sorted([]pitch.Pitch{valid})
		requireErr(t, err, system.ErrInvalid, nil)
		assert.Nil(t, res)
		requirePanicErr(t, system.ErrInadmissible, nil, func() { s.MustSorted([]pitch.Pitch{inadmissible}) })
	})
}

func TestSystem_Distance(t *testing.T) {
	t.Parallel()

	n1 := func(t testing.TB) *system.System {
		// N = 1, all letters on 0 (no letter on N).
		return newSystem(t, system.Definition{Classes: 1, Boundary: letter.C, Chromatic: 1, Alterations: system.Integers()})
	}
	trad := func(testing.TB) *system.System { return system.Traditional() }
	far := 768614336404564651 // 12·far = 2^63 + 4

	tests := []struct {
		name     string
		build    func(t testing.TB) *system.System
		from, to pitch.Pitch
		want     system.Step
	}{
		{"C4 to E4 (K-8)", trad, pt(spelling.C(), 4), pt(spelling.E(), 4), 4},
		{"C4 to E5 (K-8)", trad, pt(spelling.C(), 4), pt(spelling.E(), 5), 16},
		{"E5 to C4 (K-8)", trad, pt(spelling.E(), 5), pt(spelling.C(), 4), -16},
		{"B sharp 3 to C4 (K-8)", trad, pt(spelling.BSharp(), 4), pt(spelling.C(), 4), 0},
		{"B3 to C flat 4", trad, pt(spelling.B(), 3), pt(spelling.CFlat(), 3), 0},
		{"SA: G4 to A5 (K-25)", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, pt(spelling.G(), 4), pt(spelling.A(), 5), 2},
		{"S19: C4 to D4", func(testing.TB) *system.System { return system.Positions19() }, pt(spelling.C(), 4), pt(spelling.D(), 4), 3},
		{"C to D at Octave MAX (K-22)", trad, pt(spelling.C(), math.MaxInt), pt(spelling.D(), math.MaxInt), 2},
		{"C to B at Octave MIN (K-22)", trad, pt(spelling.C(), math.MinInt), pt(spelling.B(), math.MinInt), 11},
		{"result MaxInt (contract 6.2)", trad, pt(spelling.F(), 0), pt(spelling.C(), far), math.MaxInt},
		{"result MinInt (contract 6.2)", trad, pt(spelling.C(), far), pt(spelling.E(), 0), math.MinInt},
		{"N = 1: MaxInt", n1, pt(spelling.C(), 0), pt(spelling.G(), math.MaxInt), math.MaxInt},
		{"N = 1: MinInt", n1, pt(spelling.C(), 0), pt(spelling.G(), math.MinInt), math.MinInt},
		{"N = 1: -MaxInt", n1, pt(spelling.C(), math.MaxInt), pt(spelling.A(), 0), -math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Distance(tt.from, tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustDistance(tt.from, tt.to))
		})
	}

	errTests := []struct {
		name     string
		build    func(t testing.TB) *system.System
		from, to pitch.Pitch
		category error
	}{
		{"MIN to MAX (K-22)", trad, pt(spelling.C(), math.MinInt), pt(spelling.C(), math.MaxInt), system.ErrOutOfRange},
		{"MAX to MIN", trad, pt(spelling.C(), math.MaxInt), pt(spelling.C(), math.MinInt), system.ErrOutOfRange},
		{"one past MaxInt (contract 6.2)", trad, pt(spelling.F(), -1), pt(spelling.C(), far), system.ErrOutOfRange},
		{"one past MinInt", trad, pt(spelling.C(), far), pt(spelling.EFlat(), 0), system.ErrOutOfRange},
		{"C0 to C at Octave MaxInt", trad, pt(spelling.C(), 0), pt(spelling.C(), math.MaxInt), system.ErrOutOfRange},
		{"N = 1: one past MaxInt", n1, pt(spelling.C(), -1), pt(spelling.C(), math.MaxInt), system.ErrOutOfRange},
		{"N = 1: MaxInt to MinInt", n1, pt(spelling.C(), math.MaxInt), pt(spelling.C(), math.MinInt), system.ErrOutOfRange},
		{"inadmissible from (K-17)", trad, pt(dHalfSharp(), 4), pt(spelling.C(), 4), system.ErrInadmissible},
		{"inadmissible to (K-17)", trad, pt(spelling.C(), 4), pt(dHalfSharp(), 4), system.ErrInadmissible},
		{"inadmissible before out of range", trad, pt(dHalfSharp(), math.MinInt), pt(spelling.C(), math.MaxInt), system.ErrInadmissible},
		{"unset from", trad, pitch.Pitch{}, pt(spelling.C(), 4), system.ErrInvalid},
		{"inadmissible from and unset to", trad, pt(dHalfSharp(), 4), pitch.Pitch{}, system.ErrInvalid},
		{"unset system", func(testing.TB) *system.System { return nil }, pt(spelling.C(), 4), pt(spelling.C(), 4), system.ErrInvalid},
	}

	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Distance(tt.from, tt.to)
			requireErr(t, err, tt.category, nil)
			assert.Zero(t, got)
			requirePanicErr(t, tt.category, nil, func() { s.MustDistance(tt.from, tt.to) })
		})
	}
}

func TestSystem_PositionProperties(t *testing.T) {
	t.Parallel()

	for _, m := range []model{modelT(), modelS19(), modelSA(), modelS14(), modelS2()} {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()

			s := m.build(t)
			// A sparse sample keeps the triple loop small.
			all := samplePitches()
			sample := make([]pitch.Pitch, 0, len(all)/7+1)
			for i := 0; i < len(all); i += 7 {
				sample = append(sample, all[i])
			}

			for _, x := range sample {
				for _, y := range sample {
					d, err := s.Distance(x, y)
					require.NoError(t, err)
					assert.Equal(t, system.Step(m.position(t, y)-m.position(t, x)), d, "Distance(%s, %s)", x, y)
					assert.Equal(t, -d, s.MustDistance(y, x), "antisymmetry")

					c := s.MustCompare(x, y)
					assert.Equal(t, -cmp.Compare(int(d), 0), c, "Compare(%s, %s) is the opposite of the sign of Distance(%s, %s)", x, y, x, y)
					assert.Equal(t, -c, s.MustCompare(y, x))
					assert.Equal(t, c == 0, s.MustEquivalent(x, y), "Equivalent ⇔ Compare == 0")
				}
			}
			for i := 0; i+2 < len(sample); i += 3 {
				x, y, z := sample[i], sample[i+1], sample[i+2]
				assert.Equal(t, s.MustDistance(x, z), s.MustDistance(x, y)+s.MustDistance(y, z), "additivity")
			}
		})
	}
}

func TestSystem_Equivalent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t testing.TB) *system.System
		x, y  pitch.Pitch
		want  bool
	}{
		{"T: B sharp 3 and C4 (K-9)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.BSharp(), 4), pt(spelling.C(), 4), true},
		{"T: C flat 4 and B3 (K-9)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.CFlat(), 3), pt(spelling.B(), 3), true},
		{"T: B sharp 3 and C3 (K-9)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.BSharp(), 4), pt(spelling.C(), 3), false},
		{"T: B sharp and C at Octave MAX (K-22)", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.BSharp(), math.MaxInt), pt(spelling.C(), math.MaxInt), true},
		{"SA: A flat 4 and G sharp 3 (K-25)", func(t testing.TB) *system.System { return newSystem(t, defSA()) }, pt(spelling.AFlat(), 3), pt(spelling.GSharp(), 3), true},
		{"S19: C sharp 4 and D flat 4", func(testing.TB) *system.System { return system.Positions19() }, pt(spelling.CSharp(), 4), pt(spelling.DFlat(), 4), false},
		{"pitch equality implies equivalence", func(testing.TB) *system.System { return system.Traditional() }, pt(spelling.ESharp(), -7), pt(spelling.ESharp(), -7), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.build(t)
			got, err := s.Equivalent(tt.x, tt.y)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, s.MustEquivalent(tt.y, tt.x))
		})
	}
}

func TestSystem_EnharmonicPitches(t *testing.T) {
	t.Parallel()

	s := system.Traditional()

	tests := []struct {
		name  string
		p     pitch.Pitch
		limit alteration.Alteration
		want  []pitch.Pitch
	}{
		{"C4, L = 2 (K-10)", pt(spelling.C(), 4), a(2),
			[]pitch.Pitch{pt(spelling.C(), 4), pt(spelling.DDoubleFlat(), 4), pt(spelling.BSharp(), 4)}},
		{"C4, L = 0", pt(spelling.C(), 4), a(0), []pitch.Pitch{pt(spelling.C(), 4)}},
		{"B sharp 3, L = 1: same Octave", pt(spelling.BSharp(), 4), a(1),
			[]pitch.Pitch{pt(spelling.C(), 4), pt(spelling.BSharp(), 4)}},
		{"G sharp at Octave MIN, L = 2", pt(spelling.GSharp(), math.MinInt), a(2),
			[]pitch.Pitch{pt(spelling.GSharp(), math.MinInt), pt(spelling.AFlat(), math.MinInt)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := s.EnharmonicPitches(tt.p, tt.limit)
			require.NoError(t, err)
			got := collect(t, seq)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, collect(t, seq), "repeated traversal")
			assert.Equal(t, tt.want[:1], first(t, seq, 1), "early stop")
			assert.Equal(t, tt.want, collect(t, s.MustEnharmonicPitches(tt.p, tt.limit)))
			for _, e := range got {
				assert.True(t, s.MustEquivalent(tt.p, e), "%s ≡ %s", tt.p, e)
			}
		})
	}

	t.Run("D sharp 4, L = 0: empty", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, collect(t, s.MustEnharmonicPitches(pt(spelling.DSharp(), 4), a(0))))
	})

	errTests := []struct {
		name     string
		s        *system.System
		p        pitch.Pitch
		limit    alteration.Alteration
		category error
	}{
		{"unset system", nil, pt(spelling.C(), 4), a(1), system.ErrInvalid},
		{"unset pitch", s, pitch.Pitch{}, a(1), system.ErrInvalid},
		{"negative limit", s, pt(spelling.C(), 4), a(-1), system.ErrLimit},
		{"negative limit before inadmissible name", s, pt(dHalfSharp(), 4), a(-1), system.ErrLimit},
		{"inadmissible name", s, pt(dHalfSharp(), 4), a(1), system.ErrInadmissible},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq, err := tt.s.EnharmonicPitches(tt.p, tt.limit)
			requireErr(t, err, tt.category, nil)
			assert.Nil(t, seq)
			requirePanicErr(t, tt.category, nil, func() { tt.s.MustEnharmonicPitches(tt.p, tt.limit) })
		})
	}
}

func TestSystem_PitchErrors(t *testing.T) {
	t.Parallel()

	s := system.Traditional()
	valid := pt(spelling.C(), 4)
	bad := pt(dHalfSharp(), 4)
	var unsetP pitch.Pitch

	tests := []struct {
		name     string
		call     func() error
		must     func()
		category error
	}{
		{"Compare: inadmissible x (K-17)", func() error { _, err := s.Compare(bad, valid); return err }, func() { s.MustCompare(bad, valid) }, system.ErrInadmissible},
		{"Compare: inadmissible y", func() error { _, err := s.Compare(valid, bad); return err }, func() { s.MustCompare(valid, bad) }, system.ErrInadmissible},
		{"Compare: inadmissible x and unset y", func() error { _, err := s.Compare(bad, unsetP); return err }, func() { s.MustCompare(bad, unsetP) }, system.ErrInvalid},
		{"Compare: unset system", func() error { _, err := (*system.System)(nil).Compare(valid, valid); return err }, func() { (*system.System)(nil).MustCompare(valid, valid) }, system.ErrInvalid},
		{"Between: inadmissible x (K-17)", func() error { _, err := s.Between(bad, valid, valid); return err }, func() { s.MustBetween(bad, valid, valid) }, system.ErrInadmissible},
		{"Between: inadmissible end", func() error { _, err := s.Between(valid, valid, bad); return err }, func() { s.MustBetween(valid, valid, bad) }, system.ErrInadmissible},
		{"Between: inadmissible x and unset end", func() error { _, err := s.Between(bad, valid, unsetP); return err }, func() { s.MustBetween(bad, valid, unsetP) }, system.ErrInvalid},
		{"Equivalent: inadmissible (K-17)", func() error { _, err := s.Equivalent(valid, bad); return err }, func() { s.MustEquivalent(valid, bad) }, system.ErrInadmissible},
		{"Equivalent: inadmissible and unset", func() error { _, err := s.Equivalent(bad, unsetP); return err }, func() { s.MustEquivalent(bad, unsetP) }, system.ErrInvalid},
		{"AdmitsPitch: unset system and pitch", func() error { _, err := (&system.System{}).AdmitsPitch(unsetP); return err }, func() { (&system.System{}).MustAdmitsPitch(unsetP) }, system.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireErr(t, tt.call(), tt.category, nil)
			requirePanicErr(t, tt.category, nil, tt.must)
		})
	}
}
