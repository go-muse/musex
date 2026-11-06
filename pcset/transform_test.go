package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranspose(t *testing.T) {
	t.Parallel()

	s := pcset.MustNew(12, 0, 4, 7)
	oracleS := newOracleSet(12, 0, 4, 7)

	assertSet(t, s.Transpose(2), 12, []int{2, 6, 9})
	assertSet(t, s.Transpose(-1), 12, []int{3, 6, 11})
	assertSet(t, s.Transpose(-2), 12, []int{2, 5, 10})
	assertSet(t, s.Transpose(maxInt), 12, []int{2, 7, 11})
	assertSet(t, s.Transpose(minInt), 12, []int{4, 8, 11})
	assert.Equal(t, s, s.Transpose(0))
	assert.Equal(t, s, s.Transpose(12))
	assert.Equal(t, s, s.Transpose(-24))

	for _, step := range []int{-100, -13, -12, -1, 0, 1, 5, 11, 12, 13, 100, minInt, maxInt} {
		want := transposeOracle(oracleS, step)
		assertSet(t, s.Transpose(step), 12, want.elems)
	}

	for _, mod := range []int{1, 2, 7, 12, 19} {
		x := pcset.MustReduce(mod, 0, 4, 7)
		for step := -25; step <= 25; step++ {
			tn := x.Transpose(step)
			assert.Equal(t, x.Cardinality(), tn.Cardinality())
			assert.Equal(t, x, tn.Transpose(-step))
			assert.Equal(t, x.Complement().Transpose(step), tn.Complement())
		}
	}

	for a := -15; a <= 15; a += 5 {
		for b := -15; b <= 15; b += 7 {
			assert.Equal(t, s.Transpose(a+b), s.Transpose(a).Transpose(b))
		}
	}

	for _, n := range []int{1, 12, 64, 65, 65_536} {
		empty := pcset.MustEmpty(n)
		univ := pcset.MustUniverse(n)
		for _, step := range []int{0, 1, -1, n, minInt, maxInt} {
			assert.Equal(t, empty, empty.Transpose(step))
			assert.Equal(t, univ, univ.Transpose(step))
		}
	}

	s65 := pcset.MustNew(65, 0, 64)
	assertSet(t, s65.Transpose(1), 65, []int{0, 1})
	assertSet(t, s65.Transpose(-1), 65, []int{63, 64})

	s65536 := pcset.MustNew(65_536, 0, 65_535)
	assertSet(t, s65536.Transpose(1), 65_536, []int{0, 1})
	assertSet(t, s65536.Transpose(minInt), 65_536, transposeOracle(newOracleSet(65_536, 0, 65_535), minInt).elems)
	assertSet(t, s65536.Transpose(maxInt), 65_536, transposeOracle(newOracleSet(65_536, 0, 65_535), maxInt).elems)

	var zero pcset.Set
	assert.False(t, zero.Transpose(3).IsValid())
}

func TestInvertAndContextualInversions(t *testing.T) {
	t.Parallel()

	s := pcset.MustNew(12, 0, 4, 7)
	oracleS := newOracleSet(12, 0, 4, 7)

	assertSet(t, s.Invert(0), 12, []int{0, 5, 8})
	assertSet(t, s.Invert(7), 12, []int{0, 3, 7})
	assertSet(t, pcset.MustNew(12, 2, 4, 7).Invert(8), 12, []int{1, 4, 6})
	assertSet(t, s.Invert(maxInt), 12, []int{0, 3, 7})
	assertSet(t, s.Invert(minInt), 12, []int{0, 4, 9})
	assertSet(t, s.Invert(4), 12, []int{0, 4, 9})
	assertSet(t, s.Invert(-8), 12, []int{0, 4, 9})

	for _, idx := range []int{-100, -12, -8, -1, 0, 1, 4, 7, 11, 12, 100, minInt, maxInt} {
		inv := s.Invert(idx)
		assertSet(t, inv, 12, invertOracle(oracleS, idx).elems)
		assert.Equal(t, s, inv.Invert(idx), "involution at index %d", idx)
		assert.Equal(t, s.Complement().Invert(idx), inv.Complement(), "complement commutation at index %d", idx)
	}

	axis := mustClass(t, 2, 12)
	invAround, err := s.InvertAround(axis)
	require.NoError(t, err)
	assertSet(t, invAround, 12, []int{0, 4, 9})
	assert.Equal(t, invAround, s.MustInvertAround(axis))

	c0 := mustClass(t, 0, 12)
	c4 := mustClass(t, 4, 12)
	invSwap, err := s.InvertSwapping(c0, c4)
	require.NoError(t, err)
	assertSet(t, invSwap, 12, []int{0, 4, 9})
	assert.Equal(t, invSwap, s.MustInvertSwapping(c0, c4))

	c7 := mustClass(t, 7, 12)
	invSwap07, err := s.InvertSwapping(c0, c7)
	require.NoError(t, err)
	assertSet(t, invSwap07, 12, []int{0, 3, 7})
	assert.Equal(t, invSwap07, s.MustInvertSwapping(c0, c7))

	s65 := pcset.MustNew(65, 0, 10, 64)
	assertSet(t, s65.Invert(0), 65, invertOracle(newOracleSet(65, 0, 10, 64), 0).elems)
	assertSet(t, s65.MustInvertAround(mustClass(t, 5, 65)), 65, invertOracle(newOracleSet(65, 0, 10, 64), 10).elems)
	assertSet(
		t,
		s65.MustInvertSwapping(mustClass(t, 0, 65), mustClass(t, 10, 65)),
		65,
		invertOracle(newOracleSet(65, 0, 10, 64), 10).elems,
	)

	s65536 := pcset.MustNew(65_536, 0, 1, 65_535)
	assertSet(t, s65536.Invert(0), 65_536, []int{0, 1, 65_535})

	var zeroSet pcset.Set
	var zeroClass pitchclass.Class
	c0Mod19 := mustClass(t, 0, 19)

	t.Run("invert_around_errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			recv    pcset.Set
			axis    pitchclass.Class
			wantErr error
		}{
			{name: "invalid_set", recv: zeroSet, axis: axis, wantErr: pcset.ErrInvalidValue},
			{name: "invalid_axis", recv: s, axis: zeroClass, wantErr: pcset.ErrInvalidValue},
			{name: "invalid_set_and_axis", recv: zeroSet, axis: zeroClass, wantErr: pcset.ErrInvalidValue},
			{name: "modulus_mismatch", recv: s, axis: c0Mod19, wantErr: pcset.ErrModulusMismatch},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.recv.InvertAround(tc.axis)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, got.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = tc.recv.MustInvertAround(tc.axis)
				}, tc.wantErr)
			})
		}
	})

	t.Run("invert_swapping_errors", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			recv    pcset.Set
			x, y    pitchclass.Class
			wantErr error
		}{
			{name: "invalid_set", recv: zeroSet, x: c0, y: c4, wantErr: pcset.ErrInvalidValue},
			{name: "invalid_x", recv: s, x: zeroClass, y: c4, wantErr: pcset.ErrInvalidValue},
			{name: "invalid_y", recv: s, x: c0, y: zeroClass, wantErr: pcset.ErrInvalidValue},
			{
				name:    "invalid_y_takes_precedence_over_mismatch",
				recv:    s,
				x:       c0Mod19,
				y:       zeroClass,
				wantErr: pcset.ErrInvalidValue,
			},
			{name: "x_modulus_mismatch", recv: s, x: c0Mod19, y: c4, wantErr: pcset.ErrModulusMismatch},
			{name: "y_modulus_mismatch", recv: s, x: c0, y: c0Mod19, wantErr: pcset.ErrModulusMismatch},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.recv.InvertSwapping(tc.x, tc.y)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, got.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = tc.recv.MustInvertSwapping(tc.x, tc.y)
				}, tc.wantErr)
			})
		}
	})
}

func TestMultiply(t *testing.T) {
	t.Parallel()

	triad := pcset.MustNew(12, 0, 4, 7)
	assertSet(t, triad.Multiply(1), 12, []int{0, 4, 7})
	assertSet(t, triad.Multiply(5), 12, []int{0, 8, 11})
	assertSet(t, triad.Multiply(7), 12, []int{0, 1, 4})
	assertSet(t, triad.Multiply(11), 12, []int{0, 5, 8})
	assert.Equal(t, triad.Invert(0), triad.Multiply(11))
	assertSet(t, triad.Multiply(maxInt), 12, []int{0, 1, 4})
	assertSet(t, triad.Multiply(0), 12, []int{0})
	assertSet(t, pcset.MustNew(12, 0, 4, 6, 7).Multiply(2), 12, []int{0, 2, 8})
	assertSet(t, pcset.MustNew(12, 0, 1, 4, 7).Multiply(6), 12, []int{0, 6})

	diatonic := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)
	oracleDiatonic := newOracleSet(12, 0, 2, 4, 5, 7, 9, 11)

	m5 := diatonic.Multiply(5)
	assertSet(t, m5, 12, []int{0, 1, 7, 8, 9, 10, 11})
	assert.Equal(t, 7, m5.Cardinality())

	m7 := diatonic.Multiply(7)
	assertSet(t, m7, 12, []int{0, 1, 2, 3, 4, 5, 11})
	assert.Equal(t, 7, m7.Cardinality())
	assert.Equal(t, diatonic, m5.Multiply(5))
	assert.Equal(t, diatonic, m7.Multiply(7))

	m0 := diatonic.Multiply(0)
	assertSet(t, m0, 12, []int{0})
	assertSet(t, pcset.MustEmpty(12).Multiply(0), 12, []int{})

	m2 := diatonic.Multiply(2)
	assertSet(t, m2, 12, []int{0, 2, 4, 6, 8, 10})
	assert.Equal(t, 6, m2.Cardinality())

	// Non-bijective M_2 does not commute with complementation:
	// M_2(complement({0})) == {0, 2, 4, 6, 8, 10} != complement(M_2({0})) == U \ {0}.
	singletonZero := pcset.MustNew(12, 0)
	lhs := singletonZero.Complement().Multiply(2)
	rhs := singletonZero.Multiply(2).Complement()
	assertSet(t, lhs, 12, []int{0, 2, 4, 6, 8, 10})
	assertSet(t, rhs, 12, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	assert.NotEqual(t, lhs, rhs)

	for _, factor := range []int{-25, -12, -7, -5, -1, 0, 1, 2, 3, 4, 5, 6, 7, 11, 12, 17, minInt, maxInt} {
		assertSet(t, diatonic.Multiply(factor), 12, multiplyOracle(oracleDiatonic, factor).elems)
	}

	s65 := pcset.MustNew(65, 0, 1, 13, 64)
	for _, factor := range []int{0, 1, 5, 13, -1, minInt, maxInt} {
		assertSet(t, s65.Multiply(factor), 65, multiplyOracle(newOracleSet(65, 0, 1, 13, 64), factor).elems)
	}

	s65536 := pcset.MustNew(65_536, 0, 1, 32_768, 65_535)
	assertSet(t, s65536.Multiply(2), 65_536, []int{0, 2, 65_534})

	var zero pcset.Set
	assert.False(t, zero.Multiply(5).IsValid())
}

func TestApplyAndMustApply(t *testing.T) {
	t.Parallel()

	s := pcset.MustNew(12, 0, 1, 4, 7)

	for _, m := range []int{1, 5, 7, 11} {
		for n := range 12 {
			op := mustOp(t, m, n, 12)
			got, err := s.Apply(op)
			require.NoError(t, err)
			want := s.Multiply(m).Transpose(n)
			assert.Equal(t, want, got, "op=%v", op)
			assert.Equal(t, want, s.MustApply(op))

			switch m {
			case 1:
				assert.Equal(t, s.Transpose(n), got)
			case 11:
				assert.Equal(t, s.Invert(n), got)
			}
		}
	}

	cMajor := pcset.MustNew(12, 0, 4, 7)
	t5m7 := mustOp(t, 7, 5, 12)
	assertSet(t, cMajor.MustApply(t5m7), 12, []int{5, 6, 9})

	op1 := mustOp(t, 5, 3, 12)
	op2 := mustOp(t, 11, 7, 12)
	composed := res(op2.Compose(op1)).get(t)
	assert.Equal(t, s.MustApply(op1).MustApply(op2), s.MustApply(composed))

	for _, mod := range []int{1, 2, 7, 12} {
		testSet := pcset.MustReduce(mod, 0, 1, 4, 7)
		for op := range pitchclass.MustOperators(mod) {
			invOp := op.Inverse()
			assert.Equal(t, testSet, testSet.MustApply(op).MustApply(invOp))
			assert.Equal(t, testSet.Cardinality(), testSet.MustApply(op).Cardinality())
			assert.Equal(t, testSet.Complement().MustApply(op), testSet.MustApply(op).Complement())
		}
	}

	s65536 := pcset.MustNew(65_536, 0, 1, 65_535)
	opLarge := mustOp(t, 3, 5, 65_536)
	assert.Equal(t, s65536.Multiply(3).Transpose(5), s65536.MustApply(opLarge))

	var zeroSet pcset.Set
	var zeroOp pitchclass.Operator
	op19 := mustOp(t, 1, 0, 19)

	errCases := []struct {
		name    string
		recv    pcset.Set
		op      pitchclass.Operator
		wantErr error
	}{
		{name: "invalid_set", recv: zeroSet, op: op1, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_operator", recv: s, op: zeroOp, wantErr: pcset.ErrInvalidValue},
		{name: "both_invalid", recv: zeroSet, op: zeroOp, wantErr: pcset.ErrInvalidValue},
		{name: "modulus_mismatch", recv: s, op: op19, wantErr: pcset.ErrModulusMismatch},
	}

	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.recv.Apply(tc.op)
			requireSentinelError(t, err, tc.wantErr)
			assert.False(t, got.IsValid())
			requirePanicsWithSentinel(t, func() {
				_ = tc.recv.MustApply(tc.op)
			}, tc.wantErr)
		})
	}
}

func TestFixedClasses(t *testing.T) {
	t.Parallel()

	t.Run("operator_fixed_classes", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			op       pitchclass.Operator
			wantMod  int
			wantNums []int
		}{
			{
				name:     "T0_mod_12_universe",
				op:       mustOp(t, 1, 0, 12),
				wantMod:  12,
				wantNums: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			},
			{
				name:     "T1_mod_12_empty",
				op:       mustOp(t, 1, 1, 12),
				wantMod:  12,
				wantNums: []int{},
			},
			{
				name:     "T3_mod_12_empty",
				op:       mustOp(t, 1, 3, 12),
				wantMod:  12,
				wantNums: []int{},
			},
			{
				name:     "I0_mod_12_two_fixed",
				op:       mustOp(t, 11, 0, 12),
				wantMod:  12,
				wantNums: []int{0, 6},
			},
			{
				name:     "I1_mod_12_empty",
				op:       mustOp(t, 11, 1, 12),
				wantMod:  12,
				wantNums: []int{},
			},
			{
				name:     "I4_mod_12_two_fixed",
				op:       mustOp(t, 11, 4, 12),
				wantMod:  12,
				wantNums: []int{2, 8},
			},
			{
				name:     "I3_mod_12_empty",
				op:       mustOp(t, 11, 3, 12),
				wantMod:  12,
				wantNums: []int{},
			},
			{
				name:     "T0M5_mod_12",
				op:       mustOp(t, 5, 0, 12),
				wantMod:  12,
				wantNums: []int{0, 3, 6, 9},
			},
			{
				name:     "T0M7_mod_12",
				op:       mustOp(t, 7, 0, 12),
				wantMod:  12,
				wantNums: []int{0, 2, 4, 6, 8, 10},
			},
			{
				name:     "T1M5_mod_12_empty",
				op:       mustOp(t, 5, 1, 12),
				wantMod:  12,
				wantNums: []int{},
			},
			{
				name:     "T4M5_mod_12",
				op:       mustOp(t, 5, 4, 12),
				wantMod:  12,
				wantNums: []int{2, 5, 8, 11},
			},
			{
				name:     "I3_mod_7_odd_modulus_single_fixed",
				op:       mustOp(t, 6, 3, 7),
				wantMod:  7,
				wantNums: []int{5},
			},
			{
				name:     "T1_mod_65536_empty",
				op:       mustOp(t, 1, 1, 65_536),
				wantMod:  65_536,
				wantNums: []int{},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := pcset.OperatorFixedClasses(tc.op)
				require.NoError(t, err)
				assertSet(t, got, tc.wantMod, tc.wantNums)
				assert.Equal(t, got, pcset.MustOperatorFixedClasses(tc.op))
			})
		}

		// Exhaustive verification against pitchclass.Operator.Apply for N in {1, 7, 12}.
		for _, n := range []int{1, 7, 12} {
			for op := range pitchclass.MustOperators(n) {
				var want []int
				for x := range n {
					c := mustClass(t, x, n)
					img := res(op.Apply(c)).get(t)
					if img == c {
						want = append(want, x)
					}
				}
				got := pcset.MustOperatorFixedClasses(op)
				assertSet(t, got, n, want)
			}
		}

		_, err := pcset.OperatorFixedClasses(pitchclass.Operator{})
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() {
			_ = pcset.MustOperatorFixedClasses(pitchclass.Operator{})
		}, pcset.ErrInvalidValue)

		op70000 := mustOp(t, 1, 0, 70_000)
		_, err = pcset.OperatorFixedClasses(op70000)
		requireSentinelError(t, err, pcset.ErrModulusTooLarge)
		requirePanicsWithSentinel(t, func() {
			_ = pcset.MustOperatorFixedClasses(op70000)
		}, pcset.ErrModulusTooLarge)
	})

	t.Run("multiplication_fixed_classes", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			m        int
			modulus  int
			wantNums []int
		}{
			{name: "m5_mod_12", m: 5, modulus: 12, wantNums: []int{0, 3, 6, 9}},
			{name: "m7_mod_12", m: 7, modulus: 12, wantNums: []int{0, 2, 4, 6, 8, 10}},
			{name: "m4_mod_12", m: 4, modulus: 12, wantNums: []int{0, 4, 8}},
			{name: "m3_mod_12", m: 3, modulus: 12, wantNums: []int{0, 6}},
			{name: "m2_mod_12", m: 2, modulus: 12, wantNums: []int{0}},
			{name: "m1_mod_12_universe", m: 1, modulus: 12, wantNums: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
			{name: "m0_mod_12_singleton_zero", m: 0, modulus: 12, wantNums: []int{0}},
			{name: "m_neg7_mod_12", m: -7, modulus: 12, wantNums: []int{0, 3, 6, 9}},
			{name: "m3_mod_7", m: 3, modulus: 7, wantNums: []int{0}},
			{name: "m0_mod_65536", m: 0, modulus: 65_536, wantNums: []int{0}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := pcset.MultiplicationFixedClasses(tc.m, tc.modulus)
				require.NoError(t, err)
				assertSet(t, got, tc.modulus, tc.wantNums)
				assert.Equal(t, got, pcset.MustMultiplicationFixedClasses(tc.m, tc.modulus))
			})
		}

		for _, mod := range []int{1, 2, 7, 12, 19} {
			for m := -20; m <= 20; m++ {
				got := pcset.MustMultiplicationFixedClasses(m, mod)
				var want []int
				for x := range mod {
					if mulModOracle(m, x, mod) == x {
						want = append(want, x)
					}
				}
				assertSet(t, got, mod, want)
			}
		}

		for _, m := range []int{minInt, maxInt} {
			got := pcset.MustMultiplicationFixedClasses(m, 12)
			var want []int
			for x := range 12 {
				if mulModOracle(m, x, 12) == x {
					want = append(want, x)
				}
			}
			assertSet(t, got, 12, want)
		}

		assert.True(t, pcset.MustMultiplicationFixedClasses(1, 65_536).IsUniverse())

		errCases := []struct {
			name    string
			m       int
			modulus int
			wantErr error
		}{
			{name: "modulus_zero", m: 1, modulus: 0, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_negative", m: 1, modulus: -12, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_min_int", m: 1, modulus: minInt, wantErr: pcset.ErrInvalidModulus},
			{name: "modulus_65537", m: 1, modulus: 65_537, wantErr: pcset.ErrModulusTooLarge},
			{name: "modulus_max_int", m: 1, modulus: maxInt, wantErr: pcset.ErrModulusTooLarge},
		}

		for _, tc := range errCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := pcset.MultiplicationFixedClasses(tc.m, tc.modulus)
				requireSentinelError(t, err, tc.wantErr)
				assert.False(t, got.IsValid())
				requirePanicsWithSentinel(t, func() {
					_ = pcset.MustMultiplicationFixedClasses(tc.m, tc.modulus)
				}, tc.wantErr)
			})
		}
	})
}
