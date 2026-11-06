package pcset_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/pcset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymmetryAndOrbitSizes(t *testing.T) {
	t.Parallel()

	t.Run("canonical_cases_k21", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name        string
			set         pcset.Set
			wantST      []int
			wantSI      []int
			wantDST     int
			wantDSI     int
			wantDSTI    int
			wantDSTTO   int
			wantDistT   int
			wantOrbitTI int
		}{
			{
				name:        "c_major_asymmetric_mod_12",
				set:         pcset.MustNew(12, 0, 4, 7),
				wantST:      []int{0},
				wantSI:      []int{},
				wantDST:     1,
				wantDSI:     0,
				wantDSTI:    1,
				wantDSTTO:   1,
				wantDistT:   12,
				wantOrbitTI: 24,
			},
			{
				name:        "chromatic_trichord_inversionally_symmetric_mod_12",
				set:         pcset.MustNew(12, 0, 1, 2),
				wantST:      []int{0},
				wantSI:      []int{2},
				wantDST:     1,
				wantDSI:     1,
				wantDSTI:    2,
				wantDSTTO:   2,
				wantDistT:   12,
				wantOrbitTI: 12,
			},
			{
				name:        "augmented_triad_mod_12",
				set:         pcset.MustNew(12, 0, 4, 8),
				wantST:      []int{0, 4, 8},
				wantSI:      []int{0, 4, 8},
				wantDST:     3,
				wantDSI:     3,
				wantDSTI:    6,
				wantDSTTO:   12,
				wantDistT:   4,
				wantOrbitTI: 4,
			},
			{
				name:        "whole_tone_scale_mod_12",
				set:         pcset.MustNew(12, 0, 2, 4, 6, 8, 10),
				wantST:      []int{0, 2, 4, 6, 8, 10},
				wantSI:      []int{0, 2, 4, 6, 8, 10},
				wantDST:     6,
				wantDSI:     6,
				wantDSTI:    12,
				wantDSTTO:   24,
				wantDistT:   2,
				wantOrbitTI: 2,
			},
			{
				name:        "empty_mod_12",
				set:         pcset.MustEmpty(12),
				wantST:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				wantSI:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				wantDST:     12,
				wantDSI:     12,
				wantDSTI:    24,
				wantDSTTO:   48,
				wantDistT:   1,
				wantOrbitTI: 1,
			},
			{
				name:        "universe_mod_12",
				set:         pcset.MustUniverse(12),
				wantST:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				wantSI:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				wantDST:     12,
				wantDSI:     12,
				wantDSTI:    24,
				wantDSTTO:   48,
				wantDistT:   1,
				wantOrbitTI: 1,
			},
			{
				name:        "singleton_mod_2_group_collapse",
				set:         pcset.MustNew(2, 0),
				wantST:      []int{0},
				wantSI:      []int{0},
				wantDST:     1,
				wantDSI:     1,
				wantDSTI:    1,
				wantDSTTO:   1,
				wantDistT:   2,
				wantOrbitTI: 2,
			},
			{
				name:        "universe_mod_2_group_collapse",
				set:         pcset.MustUniverse(2),
				wantST:      []int{0, 1},
				wantSI:      []int{0, 1},
				wantDST:     2,
				wantDSI:     2,
				wantDSTI:    2,
				wantDSTTO:   2,
				wantDistT:   1,
				wantOrbitTI: 1,
			},
			{
				name:        "universe_mod_1_group_collapse",
				set:         pcset.MustUniverse(1),
				wantST:      []int{0},
				wantSI:      []int{0},
				wantDST:     1,
				wantDSI:     1,
				wantDSTI:    1,
				wantDSTTO:   1,
				wantDistT:   1,
				wantOrbitTI: 1,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				mod := tc.set.Modulus()

				st, err := tc.set.TranspositionSymmetryIndices()
				require.NoError(t, err)
				assertSet(t, st, mod, tc.wantST)
				assert.Equal(t, st, tc.set.MustTranspositionSymmetryIndices())

				si, err := tc.set.InversionSymmetryIndices()
				require.NoError(t, err)
				assertSet(t, si, mod, tc.wantSI)
				assert.Equal(t, si, tc.set.MustInversionSymmetryIndices())

				dst, err := tc.set.DegreeOfTranspositionSymmetry()
				require.NoError(t, err)
				assert.Equal(t, tc.wantDST, dst)
				assert.Equal(t, tc.wantDST, tc.set.MustDegreeOfTranspositionSymmetry())

				dsi, err := tc.set.DegreeOfInversionSymmetry()
				require.NoError(t, err)
				assert.Equal(t, tc.wantDSI, dsi)
				assert.Equal(t, tc.wantDSI, tc.set.MustDegreeOfInversionSymmetry())

				dsti, err := tc.set.DegreeOfSymmetryTI()
				require.NoError(t, err)
				assert.Equal(t, tc.wantDSTI, dsti)
				assert.Equal(t, tc.wantDSTI, tc.set.MustDegreeOfSymmetryTI())

				dstto, err := tc.set.DegreeOfSymmetryTTO()
				require.NoError(t, err)
				assert.Equal(t, tc.wantDSTTO, dstto)
				assert.Equal(t, tc.wantDSTTO, tc.set.MustDegreeOfSymmetryTTO())

				opsTI, err := tc.set.SymmetryOperatorsTI()
				require.NoError(t, err)
				assert.Len(t, opsTI, tc.wantDSTI)
				assert.Equal(t, opsTI, tc.set.MustSymmetryOperatorsTI())
				assert.Equal(t, tc.set.MustEquivalenceOperatorsTI(tc.set), opsTI)

				opsTTO, err := tc.set.SymmetryOperatorsTTO()
				require.NoError(t, err)
				assert.Len(t, opsTTO, tc.wantDSTTO)
				assert.Equal(t, opsTTO, tc.set.MustSymmetryOperatorsTTO())
				assert.Equal(t, tc.set.MustEquivalenceOperatorsTTO(tc.set), opsTTO)

				distT, err := tc.set.DistinctTranspositionCount()
				require.NoError(t, err)
				assert.Equal(t, tc.wantDistT, distT)
				assert.Equal(t, tc.wantDistT, tc.set.MustDistinctTranspositionCount())

				orbTI, err := tc.set.OrbitSizeTI()
				require.NoError(t, err)
				assert.Equal(t, tc.wantOrbitTI, orbTI)
				assert.Equal(t, tc.wantOrbitTI, tc.set.MustOrbitSizeTI())
			})
		}
	})

	t.Run("orbit_stabilizer_exhaustive_mod_1_2_7_12_and_65", func(t *testing.T) {
		t.Parallel()

		for _, mod := range []int{1, 2, 7, 12} {
			totalSets := 1 << mod
			for mask := range totalSets {
				s := pcset.MustFromBitmask(uint64(mask), mod)

				seenT := make(map[pcset.Set]bool, mod)
				seenTI := make(map[pcset.Set]bool, 2*mod)
				for n := range mod {
					tn := s.Transpose(n)
					in := s.Invert(n)
					seenT[tn] = true
					seenTI[tn] = true
					seenTI[in] = true
				}

				assert.Equal(t, len(seenT), s.MustDistinctTranspositionCount())
				assert.Equal(t, len(seenTI), s.MustOrbitSizeTI())
			}
		}

		// Symmetric set at N = 65: 13-element subgroup {0, 5, 10, ..., 60}.
		s65 := pcset.MustNew(65, 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60)
		assert.Equal(t, 13, s65.MustDegreeOfTranspositionSymmetry())
		assert.Equal(t, 13, s65.MustDegreeOfInversionSymmetry())
		assert.Equal(t, 26, s65.MustDegreeOfSymmetryTI())
		assert.Equal(t, 5, s65.MustDistinctTranspositionCount())
		assert.Equal(t, 5, s65.MustOrbitSizeTI())
	})

	t.Run("invalid_set_errors", func(t *testing.T) {
		t.Parallel()

		var zero pcset.Set

		_, err := zero.TranspositionSymmetryIndices()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustTranspositionSymmetryIndices() }, pcset.ErrInvalidValue)

		_, err = zero.InversionSymmetryIndices()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustInversionSymmetryIndices() }, pcset.ErrInvalidValue)

		_, err = zero.SymmetryOperatorsTI()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustSymmetryOperatorsTI() }, pcset.ErrInvalidValue)

		_, err = zero.SymmetryOperatorsTTO()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustSymmetryOperatorsTTO() }, pcset.ErrInvalidValue)

		_, err = zero.DegreeOfTranspositionSymmetry()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustDegreeOfTranspositionSymmetry() }, pcset.ErrInvalidValue)

		_, err = zero.DegreeOfInversionSymmetry()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustDegreeOfInversionSymmetry() }, pcset.ErrInvalidValue)

		_, err = zero.DegreeOfSymmetryTI()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustDegreeOfSymmetryTI() }, pcset.ErrInvalidValue)

		_, err = zero.DegreeOfSymmetryTTO()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustDegreeOfSymmetryTTO() }, pcset.ErrInvalidValue)

		_, err = zero.DistinctTranspositionCount()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustDistinctTranspositionCount() }, pcset.ErrInvalidValue)

		_, err = zero.OrbitSizeTI()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustOrbitSizeTI() }, pcset.ErrInvalidValue)
	})
}

func TestInvarianceVectorMorris(t *testing.T) {
	t.Parallel()

	t.Run("canonical_cases_k22", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name         string
			set          pcset.Set
			want         [8]int
			wantSymTotal int
			wantCmpTotal int
			wantString   string
			wantNotation string
		}{
			{
				name:         "empty_mod_12",
				set:          pcset.MustEmpty(12),
				want:         [8]int{12, 12, 12, 12, 12, 12, 12, 12},
				wantSymTotal: 48,
				wantCmpTotal: 48,
				wantString:   "<12, 12, 12, 12, 12, 12, 12, 12> mod 12",
				wantNotation: "<12, 12, 12, 12, 12, 12, 12, 12>",
			},
			{
				name:         "universe_mod_12",
				set:          pcset.MustUniverse(12),
				want:         [8]int{12, 12, 12, 12, 0, 0, 0, 0},
				wantSymTotal: 48,
				wantCmpTotal: 0,
				wantString:   "<12, 12, 12, 12, 0, 0, 0, 0> mod 12",
				wantNotation: "<12, 12, 12, 12, 0, 0, 0, 0>",
			},
			{
				name:         "augmented_triad_mod_12",
				set:          pcset.MustNew(12, 0, 4, 8),
				want:         [8]int{3, 3, 3, 3, 9, 9, 9, 9},
				wantSymTotal: 12,
				wantCmpTotal: 36,
				wantString:   "<3, 3, 3, 3, 9, 9, 9, 9> mod 12",
				wantNotation: "<3, 3, 3, 3, 9, 9, 9, 9>",
			},
			{
				name:         "whole_tone_scale_mod_12",
				set:          pcset.MustNew(12, 0, 2, 4, 6, 8, 10),
				want:         [8]int{6, 6, 6, 6, 6, 6, 6, 6},
				wantSymTotal: 24,
				wantCmpTotal: 24,
				wantString:   "<6, 6, 6, 6, 6, 6, 6, 6> mod 12",
				wantNotation: "<6, 6, 6, 6, 6, 6, 6, 6>",
			},
			{
				name:         "chromatic_trichord_mod_12",
				set:          pcset.MustNew(12, 0, 1, 2),
				want:         [8]int{1, 1, 0, 0, 7, 7, 4, 4},
				wantSymTotal: 2,
				wantCmpTotal: 22,
				wantString:   "<1, 1, 0, 0, 7, 7, 4, 4> mod 12",
				wantNotation: "<1, 1, 0, 0, 7, 7, 4, 4>",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				v, err := tc.set.InvarianceVector()
				require.NoError(t, err)
				assert.Equal(t, v, tc.set.MustInvarianceVector())
				require.True(t, v.IsValid())
				assert.Equal(t, 12, v.Modulus())
				assert.Equal(t, 8, v.Len())

				assert.Equal(t, tc.want[0], v.T())
				assert.Equal(t, tc.want[1], v.I())
				assert.Equal(t, tc.want[2], v.M5())
				assert.Equal(t, tc.want[3], v.M7())
				assert.Equal(t, tc.want[4], v.ComplementT())
				assert.Equal(t, tc.want[5], v.ComplementI())
				assert.Equal(t, tc.want[6], v.ComplementM5())
				assert.Equal(t, tc.want[7], v.ComplementM7())
				assert.Equal(t, tc.wantSymTotal, v.SymmetryTotal())
				assert.Equal(t, tc.wantCmpTotal, v.ComplementTotal())
				assert.Equal(t, tc.want[:], v.Slice())
				assert.Equal(t, tc.wantString, v.String())

				not, err := v.Notation()
				require.NoError(t, err)
				assert.Equal(t, tc.wantNotation, not)
				assert.Equal(t, tc.wantNotation, v.MustNotation())
			})
		}
	})

	t.Run("exhaustive_4096_sets_mod_12", func(t *testing.T) {
		t.Parallel()

		for mask := range 4096 {
			s := pcset.MustFromBitmask(uint64(mask), 12)
			v := s.MustInvarianceVector()
			want := morrisVectorOracle(s.Numbers())

			assert.Equal(t, want[:], v.Slice(), "mask=%d", mask)
			assert.Equal(t, s.MustDegreeOfSymmetryTTO(), v.SymmetryTotal(), "mask=%d", mask)

			if s.Cardinality() > 6 {
				assert.Zero(t, v.ComplementT())
				assert.Zero(t, v.ComplementI())
				assert.Zero(t, v.ComplementM5())
				assert.Zero(t, v.ComplementM7())
				assert.Zero(t, v.ComplementTotal())
			}
		}
	})

	t.Run("order_and_map_key", func(t *testing.T) {
		t.Parallel()

		chain := []pcset.InvarianceVector{
			{},
			pcset.MustNew(12, 0, 4, 7).MustInvarianceVector(),
			pcset.MustNew(12, 0, 1, 2).MustInvarianceVector(),
			pcset.MustNew(12, 0, 4, 8).MustInvarianceVector(),
			pcset.MustNew(12, 0, 2, 4, 6, 8, 10).MustInvarianceVector(),
			pcset.MustUniverse(12).MustInvarianceVector(),
			pcset.MustEmpty(12).MustInvarianceVector(),
		}

		for i, a := range chain {
			for j, b := range chain {
				switch {
				case i < j:
					assert.Equal(t, -1, a.Compare(b), "i=%d j=%d", i, j)
					assert.Equal(t, 1, b.Compare(a), "j=%d i=%d", j, i)
					assert.False(t, a.Equal(b))
					assert.False(t, a == b)
				case i == j:
					assert.Zero(t, a.Compare(b))
					assert.Equal(t, i != 0, a.Equal(b))
					assert.True(t, a == b)
				default:
					assert.Equal(t, 1, a.Compare(b))
					assert.Equal(t, -1, b.Compare(a))
					assert.False(t, a.Equal(b))
					assert.False(t, a == b)
				}
			}
		}

		reversed := slices.Clone(chain)
		slices.Reverse(reversed)
		slices.SortFunc(reversed, func(a, b pcset.InvarianceVector) int {
			return a.Compare(b)
		})
		assert.Equal(t, chain, reversed)

		m := make(map[pcset.InvarianceVector]int, len(chain))
		for i, v := range chain {
			m[v] = i + 10
		}
		require.Len(t, m, len(chain))
		// Transposed sets have identical Morris invariance vectors.
		assert.Equal(t, 11, m[pcset.MustNew(12, 2, 6, 9).MustInvarianceVector()])
	})

	t.Run("errors_not_modulus_12_and_invalid_value", func(t *testing.T) {
		t.Parallel()

		for _, mod := range []int{1, 2, 7, 10, 11, 13, 19, 24, 65, 65_536} {
			s := pcset.MustEmpty(mod)
			v, err := s.InvarianceVector()
			requireSentinelError(t, err, pcset.ErrNotModulus12)
			assert.False(t, v.IsValid())
			requirePanicsWithSentinel(t, func() { _ = s.MustInvarianceVector() }, pcset.ErrNotModulus12)
		}

		var zeroSet pcset.Set
		_, err := zeroSet.InvarianceVector()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zeroSet.MustInvarianceVector() }, pcset.ErrInvalidValue)

		var zeroVec pcset.InvarianceVector
		_, err = zeroVec.Notation()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zeroVec.MustNotation() }, pcset.ErrInvalidValue)
	})
}
