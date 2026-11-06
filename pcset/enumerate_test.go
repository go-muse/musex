package pcset_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/pcset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllAndOfCardinality(t *testing.T) {
	t.Parallel()

	t.Run("all_mod_1_3_12", func(t *testing.T) {
		t.Parallel()

		seq1, err := pcset.All(1)
		require.NoError(t, err)
		want1 := []pcset.Set{
			pcset.MustEmpty(1),
			pcset.MustNew(1, 0),
		}
		assertSeqContract(t, seq1, want1)
		assertSeqContract(t, pcset.MustAll(1), want1)

		seq3, err := pcset.All(3)
		require.NoError(t, err)
		want3 := []pcset.Set{
			pcset.MustEmpty(3),
			pcset.MustNew(3, 0),
			pcset.MustNew(3, 1),
			pcset.MustNew(3, 0, 1),
			pcset.MustNew(3, 2),
			pcset.MustNew(3, 0, 2),
			pcset.MustNew(3, 1, 2),
			pcset.MustNew(3, 0, 1, 2),
		}
		assertSeqContract(t, seq3, want3)
		assertSeqContract(t, pcset.MustAll(3), want3)

		all12 := slices.Collect(pcset.MustAll(12))
		require.Len(t, all12, 4096)
		for mask, s := range all12 {
			assert.Equal(t, uint64(mask), s.MustBitmask())
		}
	})

	t.Run("of_cardinality_lexicographical_order", func(t *testing.T) {
		t.Parallel()

		seq42, err := pcset.OfCardinality(4, 2)
		require.NoError(t, err)
		want42 := []pcset.Set{
			pcset.MustNew(4, 0, 1),
			pcset.MustNew(4, 0, 2),
			pcset.MustNew(4, 0, 3),
			pcset.MustNew(4, 1, 2),
			pcset.MustNew(4, 1, 3),
			pcset.MustNew(4, 2, 3),
		}
		assertSeqContract(t, seq42, want42)
		assertSeqContract(t, pcset.MustOfCardinality(4, 2), want42)

		assertSeqContract(t, pcset.MustOfCardinality(12, 0), []pcset.Set{pcset.MustEmpty(12)})
		assertSeqContract(t, pcset.MustOfCardinality(12, 12), []pcset.Set{pcset.MustUniverse(12)})

		hexachords := slices.Collect(pcset.MustOfCardinality(12, 6))
		require.Len(t, hexachords, 924)
		for i, s := range hexachords {
			assert.Equal(t, 6, s.Cardinality())
			if i > 0 {
				assert.Equal(t, -1, hexachords[i-1].Compare(s), "strictly ascending Compare order at %d", i)
			}
		}
	})

	t.Run("lazy_early_termination_at_max_modulus_65536", func(t *testing.T) {
		t.Parallel()

		seqAll := pcset.MustAll(pcset.MaxModulus)
		firstFour := takeSeq(seqAll, 4)
		assert.Equal(t, []pcset.Set{
			pcset.MustEmpty(pcset.MaxModulus),
			pcset.MustNew(pcset.MaxModulus, 0),
			pcset.MustNew(pcset.MaxModulus, 1),
			pcset.MustNew(pcset.MaxModulus, 0, 1),
		}, firstFour)
		assert.Zero(t, callsAfterStopSeq(seqAll, 4))

		seqK2 := pcset.MustOfCardinality(pcset.MaxModulus, 2)
		firstThree := takeSeq(seqK2, 3)
		assert.Equal(t, []pcset.Set{
			pcset.MustNew(pcset.MaxModulus, 0, 1),
			pcset.MustNew(pcset.MaxModulus, 0, 2),
			pcset.MustNew(pcset.MaxModulus, 0, 3),
		}, firstThree)
		assert.Zero(t, callsAfterStopSeq(seqK2, 3))

		assertSeqContract(t, pcset.MustOfCardinality(pcset.MaxModulus, 0), []pcset.Set{pcset.MustEmpty(pcset.MaxModulus)})
		assertSeqContract(t, pcset.MustOfCardinality(pcset.MaxModulus, pcset.MaxModulus), []pcset.Set{pcset.MustUniverse(pcset.MaxModulus)})
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name    string
			modulus int
			wantErr error
		}{
			{name: "zero_modulus", modulus: 0, wantErr: pcset.ErrInvalidModulus},
			{name: "negative_modulus", modulus: -1, wantErr: pcset.ErrInvalidModulus},
			{name: "min_int_modulus", modulus: minInt, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", modulus: 65_537, wantErr: pcset.ErrModulusTooLarge},
			{name: "max_int_modulus", modulus: maxInt, wantErr: pcset.ErrModulusTooLarge},
		} {
			t.Run("all_"+tc.name, func(t *testing.T) {
				t.Parallel()

				seq, err := pcset.All(tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				assertEmptySeqContract(t, seq)
				requirePanicsWithSentinel(t, func() { _ = pcset.MustAll(tc.modulus) }, tc.wantErr)

				seqK, err := pcset.OfCardinality(tc.modulus, 0)
				requireSentinelError(t, err, tc.wantErr)
				assertEmptySeqContract(t, seqK)
				requirePanicsWithSentinel(t, func() { _ = pcset.MustOfCardinality(tc.modulus, 0) }, tc.wantErr)
			})
		}

		for _, badK := range []int{-1, minInt, 13, maxInt} {
			seqK, err := pcset.OfCardinality(12, badK)
			requireSentinelError(t, err, pcset.ErrInvalidCardinality)
			assertEmptySeqContract(t, seqK)
			requirePanicsWithSentinel(t, func() { _ = pcset.MustOfCardinality(12, badK) }, pcset.ErrInvalidCardinality)
		}
	})
}

func TestCountAndCountOfCardinality(t *testing.T) {
	t.Parallel()

	t.Run("total_count_and_overflow_boundary", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			modulus int
			want    int
		}{
			{modulus: 1, want: 2},
			{modulus: 12, want: 4096},
			{modulus: 24, want: 16_777_216},
			{modulus: 62, want: 1 << 62},
		}

		for _, tc := range cases {
			got, err := pcset.Count(tc.modulus)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, pcset.MustCount(tc.modulus))

			gotBig, err := pcset.CountBig(tc.modulus)
			require.NoError(t, err)
			wantBig := new(big.Int).Lsh(big.NewInt(1), uint(tc.modulus))
			assert.Zero(t, wantBig.Cmp(gotBig))
			assert.Zero(t, wantBig.Cmp(pcset.MustCountBig(tc.modulus)))
		}

		for _, mod := range []int{63, 64, 65, pcset.MaxModulus} {
			_, err := pcset.Count(mod)
			requireSentinelError(t, err, pcset.ErrOverflow)
			requirePanicsWithSentinel(t, func() { _ = pcset.MustCount(mod) }, pcset.ErrOverflow)

			gotBig, err := pcset.CountBig(mod)
			require.NoError(t, err)
			assert.Equal(t, mod+1, gotBig.BitLen())
			assert.Equal(t, uint(1), gotBig.Bit(mod))
			assert.Equal(t, uint(0), gotBig.Bit(0))
			assert.Zero(t, gotBig.Cmp(pcset.MustCountBig(mod)))
		}
	})

	t.Run("binomial_count_and_overflow_boundary", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 1, res(pcset.CountOfCardinality(12, 0)).get(t))
		assert.Equal(t, 220, res(pcset.CountOfCardinality(12, 3)).get(t))
		assert.Equal(t, 924, res(pcset.CountOfCardinality(12, 6)).get(t))
		assert.Equal(t, 1, res(pcset.CountOfCardinality(12, 12)).get(t))

		sum12 := 0
		for k := 0; k <= 12; k++ {
			c := pcset.MustCountOfCardinality(12, k)
			sum12 += c
			assert.Zero(t, binomBigOracle(12, k).Cmp(pcset.MustCountOfCardinalityBig(12, k)))
		}
		assert.Equal(t, 4096, sum12)

		// C(66, 33) = 7219428434016265740 <= math.MaxInt64: fits in int.
		c66, err := pcset.CountOfCardinality(66, 33)
		require.NoError(t, err)
		assert.Equal(t, 7_219_428_434_016_265_740, c66)
		assert.Equal(t, c66, pcset.MustCountOfCardinality(66, 33))

		// C(67, 33) = 14226520737620288370 > math.MaxInt64: overflows int.
		_, err = pcset.CountOfCardinality(67, 33)
		requireSentinelError(t, err, pcset.ErrOverflow)
		requirePanicsWithSentinel(t, func() { _ = pcset.MustCountOfCardinality(67, 33) }, pcset.ErrOverflow)

		big67, err := pcset.CountOfCardinalityBig(67, 33)
		require.NoError(t, err)
		want67, ok := new(big.Int).SetString("14226520737620288370", 10)
		require.True(t, ok)
		assert.Zero(t, want67.Cmp(big67))
		assert.Zero(t, want67.Cmp(pcset.MustCountOfCardinalityBig(67, 33)))

		// Boundary values at MaxModulus = 65536.
		assert.Equal(t, 1, pcset.MustCountOfCardinality(pcset.MaxModulus, 0))
		assert.Equal(t, pcset.MaxModulus, pcset.MustCountOfCardinality(pcset.MaxModulus, 1))
		assert.Equal(t, pcset.MaxModulus*(pcset.MaxModulus-1)/2, pcset.MustCountOfCardinality(pcset.MaxModulus, 2))
		assert.Equal(t, pcset.MaxModulus, pcset.MustCountOfCardinality(pcset.MaxModulus, pcset.MaxModulus-1))
		assert.Equal(t, 1, pcset.MustCountOfCardinality(pcset.MaxModulus, pcset.MaxModulus))
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name    string
			modulus int
			wantErr error
		}{
			{name: "zero_modulus", modulus: 0, wantErr: pcset.ErrInvalidModulus},
			{name: "negative_modulus", modulus: -5, wantErr: pcset.ErrInvalidModulus},
			{name: "min_int_modulus", modulus: minInt, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", modulus: 65_537, wantErr: pcset.ErrModulusTooLarge},
			{name: "max_int_modulus", modulus: maxInt, wantErr: pcset.ErrModulusTooLarge},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				_, err := pcset.Count(tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = pcset.MustCount(tc.modulus) }, tc.wantErr)

				_, err = pcset.CountBig(tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = pcset.MustCountBig(tc.modulus) }, tc.wantErr)

				_, err = pcset.CountOfCardinality(tc.modulus, 0)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = pcset.MustCountOfCardinality(tc.modulus, 0) }, tc.wantErr)

				_, err = pcset.CountOfCardinalityBig(tc.modulus, 0)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = pcset.MustCountOfCardinalityBig(tc.modulus, 0) }, tc.wantErr)
			})
		}

		for _, badK := range []int{-1, minInt, 13, maxInt} {
			_, err := pcset.CountOfCardinality(12, badK)
			requireSentinelError(t, err, pcset.ErrInvalidCardinality)
			requirePanicsWithSentinel(t, func() { _ = pcset.MustCountOfCardinality(12, badK) }, pcset.ErrInvalidCardinality)

			_, err = pcset.CountOfCardinalityBig(12, badK)
			requireSentinelError(t, err, pcset.ErrInvalidCardinality)
			requirePanicsWithSentinel(t, func() { _ = pcset.MustCountOfCardinalityBig(12, badK) }, pcset.ErrInvalidCardinality)
		}
	})
}
