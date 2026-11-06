package pcset_test

import (
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyAndUniverse(t *testing.T) {
	t.Parallel()

	moduli := []int{1, 2, 7, 10, 12, 19, 24, 31, 53, 64, 65}
	for _, n := range moduli {
		t.Run("modulus_"+itoa(n), func(t *testing.T) {
			t.Parallel()

			empty, err := pcset.Empty(n)
			require.NoError(t, err)
			assertSet(t, empty, n, []int{})
			assertEmptySeqContract(t, empty.All())
			assert.Equal(t, empty, pcset.MustEmpty(n))

			wantUniv := make([]int, n)
			for i := range n {
				wantUniv[i] = i
			}

			univ, err := pcset.Universe(n)
			require.NoError(t, err)
			assertSet(t, univ, n, wantUniv)
			assertSeqContract(t, univ.All(), univ.Classes())
			assert.Equal(t, univ, pcset.MustUniverse(n))
		})
	}

	t.Run("modulus_65536_boundary", func(t *testing.T) {
		t.Parallel()

		empty, err := pcset.Empty(pcset.MaxModulus)
		require.NoError(t, err)
		assertSet(t, empty, pcset.MaxModulus, []int{})
		assert.Equal(t, empty, pcset.MustEmpty(pcset.MaxModulus))

		univ, err := pcset.Universe(pcset.MaxModulus)
		require.NoError(t, err)
		require.True(t, univ.IsValid())
		assert.Equal(t, pcset.MaxModulus, univ.Modulus())
		assert.Equal(t, pcset.MaxModulus, univ.Cardinality())
		assert.False(t, univ.IsEmpty())
		assert.True(t, univ.IsUniverse())
		assert.True(t, univ.ContainsNumber(0))
		assert.True(t, univ.ContainsNumber(32768))
		assert.True(t, univ.ContainsNumber(pcset.MaxModulus-1))
		assert.False(t, univ.ContainsNumber(-1))
		assert.False(t, univ.ContainsNumber(pcset.MaxModulus))
		assert.Equal(t, univ, pcset.MustUniverse(pcset.MaxModulus))

		firstThree := takeSeq(univ.All(), 3)
		assert.Equal(t, []int{0, 1, 2}, numbersOfClasses(firstThree))
		assert.Zero(t, callsAfterStopSeq(univ.All(), 3))

		nums := univ.Numbers()
		require.Len(t, nums, pcset.MaxModulus)
		assert.Equal(t, 0, nums[0])
		assert.Equal(t, pcset.MaxModulus-1, nums[pcset.MaxModulus-1])

		classes := univ.Classes()
		require.Len(t, classes, pcset.MaxModulus)
		assert.Equal(t, mustClass(t, 0, pcset.MaxModulus), classes[0])
		assert.Equal(t, mustClass(t, pcset.MaxModulus-1, pcset.MaxModulus), classes[pcset.MaxModulus-1])
	})

	errCases := []struct {
		name    string
		modulus int
		wantErr error
	}{
		{name: "zero_modulus", modulus: 0, wantErr: pcset.ErrInvalidModulus},
		{name: "negative_modulus", modulus: -1, wantErr: pcset.ErrInvalidModulus},
		{name: "min_int_modulus", modulus: minInt, wantErr: pcset.ErrInvalidModulus},
		{name: "above_max_modulus_65537", modulus: pcset.MaxModulus + 1, wantErr: pcset.ErrModulusTooLarge},
		{name: "above_max_modulus_100000", modulus: 100_000, wantErr: pcset.ErrModulusTooLarge},
		{name: "max_int_modulus", modulus: maxInt, wantErr: pcset.ErrModulusTooLarge},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, err := pcset.Empty(tc.modulus)
			requireSentinelError(t, err, tc.wantErr)
			assert.False(t, s.IsValid())
			requirePanicsWithSentinel(t, func() { _ = pcset.MustEmpty(tc.modulus) }, tc.wantErr)

			u, err := pcset.Universe(tc.modulus)
			requireSentinelError(t, err, tc.wantErr)
			assert.False(t, u.IsValid())
			requirePanicsWithSentinel(t, func() { _ = pcset.MustUniverse(tc.modulus) }, tc.wantErr)
		})
	}
}

func TestNewAndMustNew(t *testing.T) {
	t.Parallel()

	t.Run("valid_cases", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			modulus  int
			input    []int
			wantNums []int
		}{
			{name: "empty_slice_mod_12", modulus: 12, input: nil, wantNums: []int{}},
			{name: "single_zero_mod_1", modulus: 1, input: []int{0}, wantNums: []int{0}},
			{name: "sorted_c_major_mod_12", modulus: 12, input: []int{0, 4, 7}, wantNums: []int{0, 4, 7}},
			{name: "unsorted_c_major_mod_12", modulus: 12, input: []int{7, 0, 4}, wantNums: []int{0, 4, 7}},
			{name: "duplicates_c_major_mod_12", modulus: 12, input: []int{0, 4, 4, 7, 0}, wantNums: []int{0, 4, 7}},
			{
				name:     "universe_mod_12",
				modulus:  12,
				input:    []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
				wantNums: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			},
			{name: "boundary_mod_64", modulus: 64, input: []int{63, 0, 31}, wantNums: []int{0, 31, 63}},
			{name: "boundary_mod_65", modulus: 65, input: []int{64, 0, 64}, wantNums: []int{0, 64}},
			{name: "boundary_mod_65536", modulus: 65_536, input: []int{65_535, 0}, wantNums: []int{0, 65_535}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.New(tc.modulus, tc.input...)
				require.NoError(t, err)
				assertSet(t, s, tc.modulus, tc.wantNums)
				assertSeqContract(t, s.All(), s.Classes())
				assert.Equal(t, s, pcset.MustNew(tc.modulus, tc.input...))
			})
		}

		s1 := pcset.MustNew(12, 0, 4, 7)
		s2 := pcset.MustNew(12, 7, 0, 4)
		s3 := pcset.MustNew(12, 0, 4, 4, 7, 0)
		assert.Equal(t, s1, s2)
		assert.Equal(t, s2, s3)
		assert.True(t, s1.Equal(s2))
		assert.True(t, s2.Equal(s3))
		assert.Zero(t, s1.Compare(s2))
		assert.Zero(t, s2.Compare(s3))

		empty12 := pcset.MustNew(12)
		assert.Equal(t, pcset.MustEmpty(12), empty12)

		univ12 := pcset.MustNew(12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
		assert.Equal(t, pcset.MustUniverse(12), univ12)
	})

	t.Run("error_cases", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			modulus int
			input   []int
			wantErr error
		}{
			{name: "modulus_zero", modulus: 0, input: []int{0}, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_negative", modulus: -12, input: nil, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_min_int", modulus: minInt, input: nil, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", modulus: 65_537, input: []int{0}, wantErr: pcset.ErrModulusTooLarge},
			{name: "modulus_max_int", modulus: maxInt, input: nil, wantErr: pcset.ErrModulusTooLarge},
			{name: "negative_element", modulus: 12, input: []int{-1}, wantErr: pcset.ErrInvalidClass},
			{name: "element_equals_modulus_12", modulus: 12, input: []int{12}, wantErr: pcset.ErrInvalidClass},
			{name: "element_out_of_range_among_valid", modulus: 12, input: []int{0, 4, 12}, wantErr: pcset.ErrInvalidClass},
			{name: "element_equals_modulus_1", modulus: 1, input: []int{1}, wantErr: pcset.ErrInvalidClass},
			{name: "element_equals_modulus_65536", modulus: 65_536, input: []int{0, 65_536}, wantErr: pcset.ErrInvalidClass},
			{name: "element_min_int", modulus: 12, input: []int{minInt}, wantErr: pcset.ErrInvalidClass},
			{name: "element_max_int", modulus: 12, input: []int{maxInt}, wantErr: pcset.ErrInvalidClass},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.New(tc.modulus, tc.input...)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustNew(tc.modulus, tc.input...)
				}, tc.wantErr)
			})
		}
	})
}

func TestReduceAndMustReduce(t *testing.T) {
	t.Parallel()

	t.Run("valid_cases", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			modulus  int
			input    []int
			wantNums []int
		}{
			{
				name:     "negative_and_overflowing_mod_12",
				modulus:  12,
				input:    []int{-1, 12, 16, 7, -5},
				wantNums: []int{0, 4, 7, 11},
			},
			{
				name:     "min_int_and_max_int_mod_12",
				modulus:  12,
				input:    []int{minInt, maxInt},
				wantNums: reduceOracleSet(12, minInt, maxInt).elems,
			},
			{
				name:     "empty_input_mod_19",
				modulus:  19,
				input:    nil,
				wantNums: []int{},
			},
			{
				name:     "non_empty_mod_1_collapses_to_universe",
				modulus:  1,
				input:    []int{-99, 0, 42, minInt, maxInt},
				wantNums: []int{0},
			},
			{
				name:     "mod_65_large_residues",
				modulus:  65,
				input:    []int{-1, 65, 130, -65},
				wantNums: []int{0, 64},
			},
			{
				name:     "mod_65536_large_residues",
				modulus:  65_536,
				input:    []int{-1, 65_536, 131_073},
				wantNums: []int{0, 1, 65_535},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.Reduce(tc.modulus, tc.input...)
				require.NoError(t, err)
				assertSet(t, s, tc.modulus, tc.wantNums)
				assert.Equal(t, s, pcset.MustReduce(tc.modulus, tc.input...))
			})
		}

		assert.Equal(t, []int{4, 7}, pcset.MustReduce(12, minInt, maxInt).Numbers())
	})

	t.Run("error_cases", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			modulus int
			input   []int
			wantErr error
		}{
			{name: "modulus_zero", modulus: 0, input: []int{1, 2}, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_negative", modulus: -5, input: []int{1}, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_min_int", modulus: minInt, input: nil, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", modulus: 65_537, input: []int{1}, wantErr: pcset.ErrModulusTooLarge},
			{name: "modulus_100000", modulus: 100_000, input: nil, wantErr: pcset.ErrModulusTooLarge},
			{name: "modulus_max_int", modulus: maxInt, input: nil, wantErr: pcset.ErrModulusTooLarge},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.Reduce(tc.modulus, tc.input...)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustReduce(tc.modulus, tc.input...)
				}, tc.wantErr)
			})
		}
	})
}

func TestFromClassesAndFromClassesWithModulus(t *testing.T) {
	t.Parallel()

	c0 := mustClass(t, 0, 12)
	c4 := mustClass(t, 4, 12)
	c7 := mustClass(t, 7, 12)
	c0Mod19 := mustClass(t, 0, 19)
	c0Mod70000 := mustClass(t, 0, 70_000)
	c65535 := mustClass(t, 65_535, 65_536)

	t.Run("valid_from_classes", func(t *testing.T) {
		t.Parallel()

		s, err := pcset.FromClasses(c0, c4, c7, c4)
		require.NoError(t, err)
		assertSet(t, s, 12, []int{0, 4, 7})
		assert.Equal(t, s, pcset.MustFromClasses(c0, c4, c7, c4))

		sWithMod, err := pcset.FromClassesWithModulus(12, c0, c4, c7, c4)
		require.NoError(t, err)
		assert.Equal(t, s, sWithMod)
		assert.Equal(t, s, pcset.MustFromClassesWithModulus(12, c0, c4, c7, c4))

		emptyWithMod, err := pcset.FromClassesWithModulus(12)
		require.NoError(t, err)
		assertSet(t, emptyWithMod, 12, []int{})
		assert.Equal(t, emptyWithMod, pcset.MustFromClassesWithModulus(12))

		sLarge, err := pcset.FromClasses(c65535)
		require.NoError(t, err)
		assertSet(t, sLarge, 65_536, []int{65_535})
		assert.Equal(t, sLarge, pcset.MustFromClassesWithModulus(65_536, c65535))
	})

	t.Run("from_classes_errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			classes []pitchclass.Class
			wantErr error
		}{
			{
				name:    "empty_classes_slice",
				classes: nil,
				wantErr: pcset.ErrInvalidClass,
			},
			{
				name:    "zero_value_class_first",
				classes: []pitchclass.Class{{}},
				wantErr: pcset.ErrInvalidValue,
			},
			{
				name:    "zero_value_class_second",
				classes: []pitchclass.Class{c0, {}},
				wantErr: pcset.ErrInvalidValue,
			},
			{
				name:    "zero_value_class_takes_precedence_over_mismatch",
				classes: []pitchclass.Class{c0, c0Mod19, {}},
				wantErr: pcset.ErrInvalidValue,
			},
			{
				name:    "modulus_mismatch",
				classes: []pitchclass.Class{c0, c0Mod19},
				wantErr: pcset.ErrModulusMismatch,
			},
			{
				name:    "class_modulus_too_large_70000",
				classes: []pitchclass.Class{c0Mod70000},
				wantErr: pcset.ErrModulusTooLarge,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.FromClasses(tc.classes...)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustFromClasses(tc.classes...)
				}, tc.wantErr)
			})
		}
	})

	t.Run("from_classes_with_modulus_errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			modulus int
			classes []pitchclass.Class
			wantErr error
		}{
			{
				name:    "invalid_modulus_zero",
				modulus: 0,
				classes: nil,
				wantErr: pcset.ErrInvalidModulus,
			},
			{
				name:    "invalid_modulus_negative",
				modulus: -1,
				classes: []pitchclass.Class{c0},
				wantErr: pcset.ErrInvalidModulus,
			},
			{
				name:    "modulus_too_large_70000",
				modulus: 70_000,
				classes: []pitchclass.Class{c0Mod70000},
				wantErr: pcset.ErrModulusTooLarge,
			},
			{
				name:    "zero_value_class",
				modulus: 12,
				classes: []pitchclass.Class{c0, {}},
				wantErr: pcset.ErrInvalidValue,
			},
			{
				name:    "zero_value_class_takes_precedence_over_mismatch",
				modulus: 12,
				classes: []pitchclass.Class{c0Mod19, {}},
				wantErr: pcset.ErrInvalidValue,
			},
			{
				name:    "class_modulus_mismatch",
				modulus: 12,
				classes: []pitchclass.Class{c0, c0Mod19},
				wantErr: pcset.ErrModulusMismatch,
			},
			{
				name:    "class_modulus_70000_against_modulus_12",
				modulus: 12,
				classes: []pitchclass.Class{c0Mod70000},
				wantErr: pcset.ErrModulusTooLarge,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.FromClassesWithModulus(tc.modulus, tc.classes...)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustFromClassesWithModulus(tc.modulus, tc.classes...)
				}, tc.wantErr)
			})
		}
	})
}

func TestBitmaskAndBitmaskBig(t *testing.T) {
	t.Parallel()

	t.Run("uint64_and_big_roundtrip_le_64", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			modulus  int
			mask     uint64
			wantNums []int
		}{
			{name: "empty_mod_1", modulus: 1, mask: 0, wantNums: []int{}},
			{name: "universe_mod_1", modulus: 1, mask: 1, wantNums: []int{0}},
			{name: "empty_mod_12", modulus: 12, mask: 0, wantNums: []int{}},
			{name: "singleton_zero_mod_12", modulus: 12, mask: 1, wantNums: []int{0}},
			{name: "c_major_145_mod_12", modulus: 12, mask: 145, wantNums: []int{0, 4, 7}},
			{
				name:     "universe_4095_mod_12",
				modulus:  12,
				mask:     4095,
				wantNums: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			},
			{
				name:     "extremes_mod_64",
				modulus:  64,
				mask:     1 | (uint64(1) << 63),
				wantNums: []int{0, 63},
			},
			{
				name:    "universe_max_uint64_mod_64",
				modulus: 64,
				mask:    math.MaxUint64,
				wantNums: func() []int {
					out := make([]int, 64)
					for i := range 64 {
						out[i] = i
					}

					return out
				}(),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.FromBitmask(tc.mask, tc.modulus)
				require.NoError(t, err)
				assertSet(t, s, tc.modulus, tc.wantNums)
				assert.Equal(t, s, pcset.MustFromBitmask(tc.mask, tc.modulus))

				gotMask, err := s.Bitmask()
				require.NoError(t, err)
				assert.Equal(t, tc.mask, gotMask)
				assert.Equal(t, tc.mask, s.MustBitmask())

				wantBig := bitmaskBigOracle(tc.wantNums)
				gotBig, err := s.BitmaskBig()
				require.NoError(t, err)
				assert.Zero(t, wantBig.Cmp(gotBig))
				assert.Zero(t, wantBig.Cmp(s.MustBitmaskBig()))

				sFromBig, err := pcset.FromBitmaskBig(wantBig, tc.modulus)
				require.NoError(t, err)
				assert.Equal(t, s, sFromBig)
				assert.Equal(t, s, pcset.MustFromBitmaskBig(wantBig, tc.modulus))
			})
		}
	})

	t.Run("large_modulus_gt_64_big_int_roundtrip", func(t *testing.T) {
		t.Parallel()

		s65 := pcset.MustNew(65, 0, 64)
		_, err := s65.Bitmask()
		requireSentinelError(t, err, pcset.ErrOverflow)
		requirePanicsWithSentinel(t, func() { _ = s65.MustBitmask() }, pcset.ErrOverflow)

		empty65 := pcset.MustEmpty(65)
		_, err = empty65.Bitmask()
		requireSentinelError(t, err, pcset.ErrOverflow)

		bigMask65, err := s65.BitmaskBig()
		require.NoError(t, err)
		wantBig65 := bitmaskBigOracle([]int{0, 64})
		assert.Zero(t, wantBig65.Cmp(bigMask65))
		assert.Zero(t, wantBig65.Cmp(s65.MustBitmaskBig()))

		round65, err := pcset.FromBitmaskBig(bigMask65, 65)
		require.NoError(t, err)
		assert.Equal(t, s65, round65)
		assert.Equal(t, s65, pcset.MustFromBitmaskBig(bigMask65, 65))

		s65536 := pcset.MustNew(65_536, 0, 65_535)
		_, err = s65536.Bitmask()
		requireSentinelError(t, err, pcset.ErrOverflow)

		bigMask65536, err := s65536.BitmaskBig()
		require.NoError(t, err)
		assert.Equal(t, uint(1), bigMask65536.Bit(0))
		assert.Equal(t, uint(0), bigMask65536.Bit(1))
		assert.Equal(t, uint(1), bigMask65536.Bit(65_535))
		assert.Equal(t, 65_536, bigMask65536.BitLen())

		round65536, err := pcset.FromBitmaskBig(bigMask65536, 65_536)
		require.NoError(t, err)
		assert.Equal(t, s65536, round65536)
		assert.Equal(t, s65536, pcset.MustFromBitmaskBig(bigMask65536, 65_536))
	})

	t.Run("from_bitmask_errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			mask    uint64
			modulus int
			wantErr error
		}{
			{name: "modulus_zero", mask: 0, modulus: 0, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_negative", mask: 0, modulus: -1, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", mask: 0, modulus: 65_537, wantErr: pcset.ErrModulusTooLarge},
			{name: "modulus_65_overflow", mask: 0, modulus: 65, wantErr: pcset.ErrOverflow},
			{name: "modulus_65536_overflow", mask: 1, modulus: 65_536, wantErr: pcset.ErrOverflow},
			{name: "bit_at_modulus_12_4096", mask: 4096, modulus: 12, wantErr: pcset.ErrInvalidClass},
			{name: "max_uint64_mod_12", mask: math.MaxUint64, modulus: 12, wantErr: pcset.ErrInvalidClass},
			{name: "bit_1_mod_1", mask: 2, modulus: 1, wantErr: pcset.ErrInvalidClass},
			{name: "high_bit_mod_63", mask: uint64(1) << 63, modulus: 63, wantErr: pcset.ErrInvalidClass},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.FromBitmask(tc.mask, tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustFromBitmask(tc.mask, tc.modulus)
				}, tc.wantErr)
			})
		}
	})

	t.Run("from_bitmask_big_errors", func(t *testing.T) {
		t.Parallel()

		twoPow12 := new(big.Int).Lsh(big.NewInt(1), 12)
		twoPow65 := new(big.Int).Lsh(big.NewInt(1), 65)
		twoPow65536 := new(big.Int).Lsh(big.NewInt(1), 65_536)

		cases := []struct {
			name    string
			mask    *big.Int
			modulus int
			wantErr error
		}{
			{name: "nil_mask_valid_modulus", mask: nil, modulus: 12, wantErr: pcset.ErrInvalidValue},
			{name: "modulus_zero", mask: big.NewInt(0), modulus: 0, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_negative", mask: big.NewInt(0), modulus: -12, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", mask: big.NewInt(0), modulus: 65_537, wantErr: pcset.ErrModulusTooLarge},
			{name: "negative_mask_mod_12", mask: big.NewInt(-1), modulus: 12, wantErr: pcset.ErrInvalidClass},
			{name: "negative_mask_mod_65", mask: big.NewInt(-42), modulus: 65, wantErr: pcset.ErrInvalidClass},
			{name: "mask_2_pow_12_mod_12", mask: twoPow12, modulus: 12, wantErr: pcset.ErrInvalidClass},
			{name: "mask_2_pow_65_mod_65", mask: twoPow65, modulus: 65, wantErr: pcset.ErrInvalidClass},
			{name: "mask_2_pow_65536_mod_65536", mask: twoPow65536, modulus: 65_536, wantErr: pcset.ErrInvalidClass},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := pcset.FromBitmaskBig(tc.mask, tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, s.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustFromBitmaskBig(tc.mask, tc.modulus)
				}, tc.wantErr)
			})
		}
	})
}

func TestContainsAndContainsNumber(t *testing.T) {
	t.Parallel()

	s := pcset.MustNew(12, 0, 4, 7)

	assert.True(t, s.Contains(mustClass(t, 0, 12)))
	assert.True(t, s.Contains(mustClass(t, 4, 12)))
	assert.True(t, s.Contains(mustClass(t, 7, 12)))
	assert.False(t, s.Contains(mustClass(t, 2, 12)))
	assert.False(t, s.Contains(mustClass(t, 11, 12)))
	assert.False(t, s.Contains(mustClass(t, 0, 19)))
	assert.False(t, s.Contains(mustClass(t, 4, 24)))
	assert.False(t, s.Contains(mustClass(t, 0, 70_000)))
	assert.False(t, s.Contains(pitchclass.Class{}))

	assert.True(t, s.ContainsNumber(0))
	assert.True(t, s.ContainsNumber(4))
	assert.True(t, s.ContainsNumber(7))
	assert.False(t, s.ContainsNumber(2))
	assert.False(t, s.ContainsNumber(-1))
	assert.False(t, s.ContainsNumber(-5))
	assert.False(t, s.ContainsNumber(12))
	assert.False(t, s.ContainsNumber(16))
	assert.False(t, s.ContainsNumber(minInt))
	assert.False(t, s.ContainsNumber(maxInt))

	sLarge := pcset.MustNew(65_536, 0, 65_535)
	assert.True(t, sLarge.Contains(mustClass(t, 0, 65_536)))
	assert.True(t, sLarge.Contains(mustClass(t, 65_535, 65_536)))
	assert.False(t, sLarge.Contains(mustClass(t, 1, 65_536)))
	assert.True(t, sLarge.ContainsNumber(0))
	assert.True(t, sLarge.ContainsNumber(65_535))
	assert.False(t, sLarge.ContainsNumber(1))
	assert.False(t, sLarge.ContainsNumber(-1))
	assert.False(t, sLarge.ContainsNumber(65_536))

	var zero pcset.Set
	assert.False(t, zero.Contains(mustClass(t, 0, 12)))
	assert.False(t, zero.Contains(pitchclass.Class{}))
	assert.False(t, zero.ContainsNumber(0))
	assert.False(t, zero.ContainsNumber(-1))
}

func TestEqualCompareAndMapKey(t *testing.T) {
	t.Parallel()

	chain := []pcset.Set{
		{},
		pcset.MustEmpty(1),
		pcset.MustNew(1, 0),
		pcset.MustEmpty(12),
		pcset.MustNew(12, 0),
		pcset.MustNew(12, 1),
		pcset.MustNew(12, 0, 4),
		pcset.MustNew(12, 0, 4, 7),
		pcset.MustNew(12, 0, 4, 8),
		pcset.MustNew(12, 1, 4, 7),
		pcset.MustUniverse(12),
		pcset.MustEmpty(19),
		pcset.MustNew(65, 0, 64),
		pcset.MustNew(65_536, 0, 65_535),
	}

	for i, a := range chain {
		for j, b := range chain {
			switch {
			case i < j:
				assert.Equal(t, -1, a.Compare(b), "i=%d j=%d", i, j)
				assert.Equal(t, 1, b.Compare(a), "j=%d i=%d", j, i)
				assert.False(t, a.Equal(b), "i=%d j=%d", i, j)
				assert.False(t, a == b, "i=%d j=%d", i, j)
			case i == j:
				assert.Zero(t, a.Compare(b), "i=%d", i)
				assert.Equal(t, i > 0, a.Equal(b), "i=%d", i)
				assert.True(t, a == b, "i=%d", i)
			default:
				assert.Equal(t, 1, a.Compare(b), "i=%d j=%d", i, j)
				assert.Equal(t, -1, b.Compare(a), "j=%d i=%d", j, i)
				assert.False(t, a.Equal(b), "i=%d j=%d", i, j)
				assert.False(t, a == b, "i=%d j=%d", i, j)
			}
		}
	}

	reversed := slices.Clone(chain)
	slices.Reverse(reversed)
	slices.SortFunc(reversed, func(a, b pcset.Set) int {
		return a.Compare(b)
	})
	assert.Equal(t, chain, reversed)

	m := make(map[pcset.Set]int, len(chain))
	for i, s := range chain {
		m[s] = i + 100
	}
	require.Len(t, m, len(chain))

	assert.Equal(t, 100, m[pcset.Set{}])
	assert.Equal(t, 107, m[pcset.MustFromBitmask(145, 12)])
	assert.Equal(t, 107, m[pcset.MustReduce(12, 12, 16, 7)])
	assert.Equal(t, 112, m[pcset.MustFromBitmaskBig(bitmaskBigOracle([]int{0, 64}), 65)])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	neg := n < 0
	var digits []byte
	for n != 0 {
		d := n % 10
		if d < 0 {
			d = -d
		}
		digits = append(digits, byte('0'+d))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	slices.Reverse(digits)

	return string(digits)
}
