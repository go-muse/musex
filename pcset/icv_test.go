package pcset_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/pcset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntervalVectorCanonicalAndBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("canonical_cases_mod_1_2_7_12", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name        string
			set         pcset.Set
			wantSlice   []int
			wantTotal   int
			wantString  string
			wantNot     string
			wantCompact string
			compactErr  error
		}{
			{
				name:        "empty_mod_1",
				set:         pcset.MustEmpty(1),
				wantSlice:   []int{},
				wantTotal:   0,
				wantString:  "<> mod 1",
				wantNot:     "<>",
				wantCompact: "<>",
			},
			{
				name:        "universe_mod_1",
				set:         pcset.MustUniverse(1),
				wantSlice:   []int{},
				wantTotal:   0,
				wantString:  "<> mod 1",
				wantNot:     "<>",
				wantCompact: "<>",
			},
			{
				name:        "singleton_mod_2",
				set:         pcset.MustNew(2, 0),
				wantSlice:   []int{0},
				wantTotal:   0,
				wantString:  "<0> mod 2",
				wantNot:     "<0>",
				wantCompact: "<0>",
			},
			{
				name:        "universe_mod_2_half_octave",
				set:         pcset.MustUniverse(2),
				wantSlice:   []int{1},
				wantTotal:   1,
				wantString:  "<1> mod 2",
				wantNot:     "<1>",
				wantCompact: "<1>",
			},
			{
				name:        "universe_mod_7",
				set:         pcset.MustUniverse(7),
				wantSlice:   []int{7, 7, 7},
				wantTotal:   21,
				wantString:  "<7, 7, 7> mod 7",
				wantNot:     "<7, 7, 7>",
				wantCompact: "<777>",
			},
			{
				name:        "empty_mod_12",
				set:         pcset.MustEmpty(12),
				wantSlice:   []int{0, 0, 0, 0, 0, 0},
				wantTotal:   0,
				wantString:  "<0, 0, 0, 0, 0, 0> mod 12",
				wantNot:     "<0, 0, 0, 0, 0, 0>",
				wantCompact: "<000000>",
			},
			{
				name:        "singleton_mod_12",
				set:         pcset.MustNew(12, 0),
				wantSlice:   []int{0, 0, 0, 0, 0, 0},
				wantTotal:   0,
				wantString:  "<0, 0, 0, 0, 0, 0> mod 12",
				wantNot:     "<0, 0, 0, 0, 0, 0>",
				wantCompact: "<000000>",
			},
			{
				name:        "c_major_mod_12",
				set:         pcset.MustNew(12, 0, 4, 7),
				wantSlice:   []int{0, 0, 1, 1, 1, 0},
				wantTotal:   3,
				wantString:  "<0, 0, 1, 1, 1, 0> mod 12",
				wantNot:     "<0, 0, 1, 1, 1, 0>",
				wantCompact: "<001110>",
			},
			{
				name:        "c_minor_mod_12",
				set:         pcset.MustNew(12, 0, 3, 7),
				wantSlice:   []int{0, 0, 1, 1, 1, 0},
				wantTotal:   3,
				wantString:  "<0, 0, 1, 1, 1, 0> mod 12",
				wantNot:     "<0, 0, 1, 1, 1, 0>",
				wantCompact: "<001110>",
			},
			{
				name:        "major_scale_mod_12",
				set:         pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11),
				wantSlice:   []int{2, 5, 4, 3, 6, 1},
				wantTotal:   21,
				wantString:  "<2, 5, 4, 3, 6, 1> mod 12",
				wantNot:     "<2, 5, 4, 3, 6, 1>",
				wantCompact: "<254361>",
			},
			{
				name:        "all_interval_tetrachord_4_z15",
				set:         pcset.MustNew(12, 0, 1, 4, 6),
				wantSlice:   []int{1, 1, 1, 1, 1, 1},
				wantTotal:   6,
				wantString:  "<1, 1, 1, 1, 1, 1> mod 12",
				wantNot:     "<1, 1, 1, 1, 1, 1>",
				wantCompact: "<111111>",
			},
			{
				name:       "universe_mod_12_multi_digit_component",
				set:        pcset.MustUniverse(12),
				wantSlice:  []int{12, 12, 12, 12, 12, 6},
				wantTotal:  66,
				wantString: "<12, 12, 12, 12, 12, 6> mod 12",
				wantNot:    "<12, 12, 12, 12, 12, 6>",
				compactErr: pcset.ErrNotation,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				v := tc.set.IntervalVector()
				require.True(t, v.IsValid())
				assert.Equal(t, tc.set.Modulus(), v.Modulus())
				assert.Equal(t, len(tc.wantSlice), v.Len())
				assert.Equal(t, tc.wantTotal, v.Total())

				gotSlice := v.Slice()
				require.NotNil(t, gotSlice)
				assert.Equal(t, tc.wantSlice, gotSlice)
				assert.Equal(t, tc.wantString, v.String())

				not, err := v.Notation()
				require.NoError(t, err)
				assert.Equal(t, tc.wantNot, not)
				assert.Equal(t, tc.wantNot, v.MustNotation())

				for d := 1; d <= v.Len(); d++ {
					cnt, err := v.Count(d)
					require.NoError(t, err)
					assert.Equal(t, tc.wantSlice[d-1], cnt)
					assert.Equal(t, tc.wantSlice[d-1], v.MustCount(d))
				}

				for _, badD := range []int{-1, 0, v.Len() + 1, minInt, maxInt} {
					_, err := v.Count(badD)
					requireSentinelError(t, err, pcset.ErrInvalidIntervalClass)
					requirePanicsWithSentinel(t, func() { _ = v.MustCount(badD) }, pcset.ErrInvalidIntervalClass)
				}

				if tc.compactErr != nil {
					_, err := v.CompactNotation()
					requireSentinelError(t, err, tc.compactErr)
					requirePanicsWithSentinel(t, func() { _ = v.MustCompactNotation() }, tc.compactErr)
				} else {
					comp, err := v.CompactNotation()
					require.NoError(t, err)
					assert.Equal(t, tc.wantCompact, comp)
					assert.Equal(t, tc.wantCompact, v.MustCompactNotation())
				}
			})
		}

		// Z-related sets {0, 1, 4, 6}/12 and {0, 1, 3, 7}/12 share identical IntervalVector.
		z1 := pcset.MustNew(12, 0, 1, 4, 6).IntervalVector()
		z2 := pcset.MustNew(12, 0, 1, 3, 7).IntervalVector()
		assert.Equal(t, z1, z2)
		assert.True(t, z1.Equal(z2))
		assert.Zero(t, z1.Compare(z2))
	})

	t.Run("universe_65536_fast_via_complement_theorem", func(t *testing.T) {
		t.Parallel()

		univ := pcset.MustUniverse(pcset.MaxModulus)
		v := univ.IntervalVector()
		require.True(t, v.IsValid())
		assert.Equal(t, pcset.MaxModulus, v.Modulus())
		assert.Equal(t, 32_768, v.Len())
		assert.Equal(t, 65_536*65_535/2, v.Total())
		assert.Equal(t, 65_536, v.MustCount(1))
		assert.Equal(t, 65_536, v.MustCount(32_767))
		assert.Equal(t, 32_768, v.MustCount(32_768))

		// Near-universe set with k = 65535 (complement has cardinality 1).
		nearUniv := pcset.MustNew(pcset.MaxModulus, 0).Complement()
		vNear := nearUniv.IntervalVector()
		assert.Equal(t, 65_534, vNear.MustCount(1))
		assert.Equal(t, 32_767, vNear.MustCount(32_768))
	})
}

func TestIntervalVectorExhaustiveAndComplementTheorem(t *testing.T) {
	t.Parallel()

	hexachordCount := 0
	for _, mod := range []int{1, 2, 7, 12} {
		totalSets := 1 << mod
		for mask := range totalSets {
			s := pcset.MustFromBitmask(uint64(mask), mod)
			v := s.IntervalVector()
			k := s.Cardinality()
			want := icvOracle(mod, s.Numbers())

			assert.Equal(t, want, v.Slice())
			assert.Equal(t, k*(k-1)/2, v.Total())

			// T_n and I_n invariance.
			assert.Equal(t, v, s.Transpose(3).IntervalVector())
			assert.Equal(t, v, s.Invert(5).IntervalVector())

			// Complement Theorem P-22.
			vComp := s.Complement().IntervalVector()
			for d := 1; d <= v.Len(); d++ {
				md := v.MustCount(d)
				mdComp := vComp.MustCount(d)
				if mod%2 == 0 && d == mod/2 {
					assert.Equal(t, md+mod/2-k, mdComp, "P-22 tritone failed for mask=%d mod=%d", mask, mod)
				} else {
					assert.Equal(t, md+mod-2*k, mdComp, "P-22 failed for mask=%d mod=%d d=%d", mask, mod, d)
				}
			}

			if mod == 12 && k == 6 {
				hexachordCount++
				assert.Equal(t, v, vComp, "hexachord ICV must equal its complement ICV")
			}
		}
	}
	assert.Equal(t, 924, hexachordCount)
}

func TestIntervalVectorOrderAndMapKey(t *testing.T) {
	t.Parallel()

	chain := []pcset.IntervalVector{
		{},
		pcset.MustEmpty(1).IntervalVector(),
		pcset.MustEmpty(2).IntervalVector(),
		pcset.MustUniverse(2).IntervalVector(),
		pcset.MustEmpty(12).IntervalVector(),
		pcset.MustNew(12, 0, 4, 7).IntervalVector(),
		pcset.MustNew(12, 0, 1, 4, 6).IntervalVector(),
		pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11).IntervalVector(),
		pcset.MustUniverse(12).IntervalVector(),
		pcset.MustEmpty(19).IntervalVector(),
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
				assert.Zero(t, a.Compare(b), "i=%d", i)
				assert.Equal(t, i != 0, a.Equal(b), "i=%d", i)
				assert.True(t, a == b)
			default:
				assert.Equal(t, 1, a.Compare(b), "i=%d j=%d", i, j)
				assert.Equal(t, -1, b.Compare(a), "j=%d i=%d", j, i)
				assert.False(t, a.Equal(b))
				assert.False(t, a == b)
			}
		}
	}

	shuffled := slices.Clone(chain)
	slices.Reverse(shuffled)
	slices.SortFunc(shuffled, func(a, b pcset.IntervalVector) int {
		return a.Compare(b)
	})
	assert.Equal(t, chain, shuffled)

	m := make(map[pcset.IntervalVector]int, len(chain))
	for i, v := range chain {
		m[v] = i + 1
	}
	require.Len(t, m, len(chain))
	// Z-partner {0, 1, 3, 7}/12 hits the same map entry as {0, 1, 4, 6}/12.
	assert.Equal(t, 7, m[pcset.MustNew(12, 0, 1, 3, 7).IntervalVector()])
}
