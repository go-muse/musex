package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlgorithmContract(t *testing.T) {
	t.Parallel()

	assert.Equal(t, pcset.Algorithm(0), pcset.Rahn)
	assert.Equal(t, pcset.Algorithm(1), pcset.Forte)
	assert.True(t, pcset.Rahn.IsValid())
	assert.True(t, pcset.Forte.IsValid())
	assert.False(t, pcset.Algorithm(2).IsValid())
	assert.False(t, pcset.Algorithm(99).IsValid())
	assert.False(t, pcset.Algorithm(255).IsValid())

	assert.Equal(t, "Rahn", pcset.Rahn.String())
	assert.Equal(t, "Forte", pcset.Forte.String())
	assert.Equal(t, "Algorithm(2)", pcset.Algorithm(2).String())
	assert.Equal(t, "Algorithm(99)", pcset.Algorithm(99).String())
}

func TestRotationsAndNormalForm(t *testing.T) {
	t.Parallel()

	t.Run("rotations_and_canonical_cases", func(t *testing.T) {
		t.Parallel()

		s := pcset.MustNew(12, 2, 3, 7, 11)
		rots := s.Rotations()
		require.Len(t, rots, 4)
		assert.Equal(t, []int{2, 3, 7, 11}, numbersOfClasses(rots[0]))
		assert.Equal(t, []int{3, 7, 11, 2}, numbersOfClasses(rots[1]))
		assert.Equal(t, []int{7, 11, 2, 3}, numbersOfClasses(rots[2]))
		assert.Equal(t, []int{11, 2, 3, 7}, numbersOfClasses(rots[3]))

		assert.Equal(t, []int{11, 2, 3, 7}, numbersOfClasses(s.NormalForm()))
		assert.Equal(t, []int{11, 2, 3, 7}, numbersOfClasses(res(s.NormalFormWith(pcset.Rahn)).get(t)))
		assert.Equal(t, []int{11, 2, 3, 7}, numbersOfClasses(s.MustNormalFormWith(pcset.Rahn)))
		assert.Equal(t, []int{11, 2, 3, 7}, numbersOfClasses(res(s.NormalFormWith(pcset.Forte)).get(t)))
		assert.Equal(t, []int{11, 2, 3, 7}, numbersOfClasses(s.MustNormalFormWith(pcset.Forte)))

		// Symmetric sets (G-6): tie-break chooses smallest starting pitch class.
		for _, nums := range [][]int{
			{0, 4, 8},
			{1, 5, 9},
			{0, 3, 6, 9},
			{0, 2, 4, 7, 9},
		} {
			sym := pcset.MustNew(12, nums...)
			assert.Equal(t, nums, numbersOfClasses(sym.NormalForm()))
			assert.Equal(t, nums, numbersOfClasses(sym.MustNormalFormWith(pcset.Rahn)))
			assert.Equal(t, nums, numbersOfClasses(sym.MustNormalFormWith(pcset.Forte)))
		}

		// Empty, singleton, universe (G-3, G-4).
		empty := pcset.MustEmpty(12)
		emptyRots := empty.Rotations()
		require.NotNil(t, emptyRots)
		assert.Empty(t, emptyRots)

		emptyNF := empty.NormalForm()
		require.NotNil(t, emptyNF)
		assert.Empty(t, emptyNF)
		assert.Empty(t, empty.MustNormalFormWith(pcset.Rahn))
		assert.Empty(t, empty.MustNormalFormWith(pcset.Forte))

		single := pcset.MustNew(12, 5)
		require.Len(t, single.Rotations(), 1)
		assert.Equal(t, []int{5}, numbersOfClasses(single.Rotations()[0]))
		assert.Equal(t, []int{5}, numbersOfClasses(single.NormalForm()))
		assert.Equal(t, []int{5}, numbersOfClasses(single.MustNormalFormWith(pcset.Forte)))

		univ := pcset.MustUniverse(12)
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, numbersOfClasses(univ.NormalForm()))
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, numbersOfClasses(univ.MustNormalFormWith(pcset.Forte)))
	})

	t.Run("six_divergence_classes_rahn_vs_forte", func(t *testing.T) {
		t.Parallel()

		// All 6 Forte set-classes where Rahn and Forte normal forms diverge (lit-pitch-class-sets.md section 2.3, G-7).
		divergences := []struct {
			name      string
			elems     []int
			wantRahn  []int
			wantForte []int
		}{
			{
				name:      "5-20_form_A",
				elems:     []int{0, 1, 3, 7, 8},
				wantRahn:  []int{7, 8, 0, 1, 3},
				wantForte: []int{0, 1, 3, 7, 8},
			},
			{
				name:      "5-20_form_B",
				elems:     []int{0, 1, 5, 6, 8},
				wantRahn:  []int{0, 1, 5, 6, 8},
				wantForte: []int{5, 6, 8, 0, 1},
			},
			{
				name:      "6-Z29_form_A",
				elems:     []int{0, 2, 3, 6, 7, 9},
				wantRahn:  []int{0, 2, 3, 6, 7, 9},
				wantForte: []int{6, 7, 9, 0, 2, 3},
			},
			{
				name:      "6-Z29_form_B",
				elems:     []int{0, 1, 3, 6, 8, 9},
				wantRahn:  []int{6, 8, 9, 0, 1, 3},
				wantForte: []int{0, 1, 3, 6, 8, 9},
			},
			{
				name:      "6-31_form_A",
				elems:     []int{0, 1, 4, 5, 7, 9},
				wantRahn:  []int{0, 1, 4, 5, 7, 9},
				wantForte: []int{4, 5, 7, 9, 0, 1},
			},
			{
				name:      "6-31_form_B",
				elems:     []int{0, 1, 3, 5, 8, 9},
				wantRahn:  []int{8, 9, 0, 1, 3, 5},
				wantForte: []int{0, 1, 3, 5, 8, 9},
			},
			{
				name:      "7-18_form_A",
				elems:     []int{0, 1, 4, 5, 6, 7, 9},
				wantRahn:  []int{0, 1, 4, 5, 6, 7, 9},
				wantForte: []int{4, 5, 6, 7, 9, 0, 1},
			},
			{
				name:      "7-18_form_B",
				elems:     []int{0, 1, 2, 3, 5, 8, 9},
				wantRahn:  []int{8, 9, 0, 1, 2, 3, 5},
				wantForte: []int{0, 1, 2, 3, 5, 8, 9},
			},
			{
				name:      "7-20_form_A",
				elems:     []int{0, 1, 2, 5, 6, 7, 9},
				wantRahn:  []int{0, 1, 2, 5, 6, 7, 9},
				wantForte: []int{5, 6, 7, 9, 0, 1, 2},
			},
			{
				name:      "7-20_form_B",
				elems:     []int{0, 1, 2, 4, 7, 8, 9},
				wantRahn:  []int{7, 8, 9, 0, 1, 2, 4},
				wantForte: []int{0, 1, 2, 4, 7, 8, 9},
			},
			{
				name:      "8-26_form_A",
				elems:     []int{0, 1, 3, 4, 5, 7, 8, 10},
				wantRahn:  []int{0, 1, 3, 4, 5, 7, 8, 10},
				wantForte: []int{3, 4, 5, 7, 8, 10, 0, 1},
			},
			{
				name:      "8-26_form_B",
				elems:     []int{0, 1, 2, 4, 5, 7, 9, 10},
				wantRahn:  []int{9, 10, 0, 1, 2, 4, 5, 7},
				wantForte: []int{0, 1, 2, 4, 5, 7, 9, 10},
			},
		}

		for _, tc := range divergences {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s := pcset.MustNew(12, tc.elems...)
				gotRahn := numbersOfClasses(s.MustNormalFormWith(pcset.Rahn))
				gotForte := numbersOfClasses(s.MustNormalFormWith(pcset.Forte))

				assert.Equal(t, tc.wantRahn, gotRahn)
				assert.Equal(t, tc.wantRahn, numbersOfClasses(s.NormalForm()))
				assert.Equal(t, tc.wantForte, gotForte)
				assert.NotEqual(t, gotRahn, gotForte)
			})
		}
	})

	t.Run("exhaustive_oracle_mod_1_2_7_12_and_large_moduli", func(t *testing.T) {
		t.Parallel()

		for _, mod := range []int{1, 2, 7, 12} {
			total := 1 << mod
			for mask := range total {
				s := pcset.MustFromBitmask(uint64(mask), mod)
				elems := s.Numbers()

				wantRots := rotationsOracle(elems)
				gotRots := s.Rotations()
				require.Len(t, gotRots, len(wantRots))
				for i := range wantRots {
					assert.Equal(t, wantRots[i], numbersOfClasses(gotRots[i]))
				}

				wantRahn := normalFormRahnOracle(mod, elems)
				wantForte := normalFormForteOracle(mod, elems)
				assert.Equal(t, wantRahn, numbersOfClasses(s.NormalForm()))
				assert.Equal(t, wantRahn, numbersOfClasses(s.MustNormalFormWith(pcset.Rahn)))
				assert.Equal(t, wantForte, numbersOfClasses(s.MustNormalFormWith(pcset.Forte)))
			}
		}

		s65536 := pcset.MustNew(65_536, 0, 2, 65_535)
		assert.Equal(t, []int{65_535, 0, 2}, numbersOfClasses(s65536.NormalForm()))
		assert.Equal(t, []int{65_535, 0, 2}, numbersOfClasses(s65536.MustNormalFormWith(pcset.Forte)))
	})

	t.Run("errors_and_invalid_values", func(t *testing.T) {
		t.Parallel()

		var zero pcset.Set
		assert.Nil(t, zero.Rotations())
		assert.Nil(t, zero.NormalForm())

		_, err := zero.NormalFormWith(pcset.Rahn)
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNormalFormWith(pcset.Rahn) }, pcset.ErrInvalidValue)

		// ErrInvalidValue on Set{} takes precedence over ErrInvalidAlgorithm.
		_, err = zero.NormalFormWith(pcset.Algorithm(99))
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNormalFormWith(pcset.Algorithm(99)) }, pcset.ErrInvalidValue)

		valid := pcset.MustNew(12, 0, 4, 7)
		_, err = valid.NormalFormWith(pcset.Algorithm(99))
		requireSentinelError(t, err, pcset.ErrInvalidAlgorithm)
		requirePanicsWithSentinel(t, func() { _ = valid.MustNormalFormWith(pcset.Algorithm(99)) }, pcset.ErrInvalidAlgorithm)
	})
}
