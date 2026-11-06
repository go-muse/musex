package pcset_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranspositionAndInversionIndicesTo(t *testing.T) {
	t.Parallel()

	cMaj := pcset.MustNew(12, 0, 4, 7)
	dMaj := pcset.MustNew(12, 2, 6, 9)
	cMin := pcset.MustNew(12, 0, 3, 7)
	aug0 := pcset.MustNew(12, 0, 4, 8)
	aug1 := pcset.MustNew(12, 1, 5, 9)
	empty12 := pcset.MustEmpty(12)
	dom7 := pcset.MustNew(12, 0, 4, 7, 10)

	cases := []struct {
		name  string
		src   pcset.Set
		dst   pcset.Set
		wantT []int
		wantI []int
	}{
		{
			name:  "c_major_to_d_major",
			src:   cMaj,
			dst:   dMaj,
			wantT: []int{2},
			wantI: []int{},
		},
		{
			name:  "c_major_to_c_minor",
			src:   cMaj,
			dst:   cMin,
			wantT: []int{},
			wantI: []int{7},
		},
		{
			name:  "augmented_triads",
			src:   aug0,
			dst:   aug1,
			wantT: []int{1, 5, 9},
			wantI: []int{1, 5, 9},
		},
		{
			name:  "empty_to_empty_universe",
			src:   empty12,
			dst:   empty12,
			wantT: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			wantI: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		{
			name:  "different_cardinalities_empty",
			src:   cMaj,
			dst:   dom7,
			wantT: []int{},
			wantI: []int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotT, err := tc.src.TranspositionIndicesTo(tc.dst)
			require.NoError(t, err)
			assertSet(t, gotT, 12, tc.wantT)
			assert.Equal(t, gotT, tc.src.MustTranspositionIndicesTo(tc.dst))

			gotI, err := tc.src.InversionIndicesTo(tc.dst)
			require.NoError(t, err)
			assertSet(t, gotI, 12, tc.wantI)
			assert.Equal(t, gotI, tc.src.MustInversionIndicesTo(tc.dst))
		})
	}

	// Oracle verification across moduli including N > 64.
	for _, mod := range []int{1, 2, 7, 12, 65} {
		a := pcset.MustReduce(mod, 0, 3, 7)
		for shift := range min(mod, 5) {
			bT := a.Transpose(shift)
			bI := a.Invert(shift)
			for _, target := range []pcset.Set{bT, bI, pcset.MustEmpty(mod)} {
				gotT := a.MustTranspositionIndicesTo(target)
				gotI := a.MustInversionIndicesTo(target)
				for n := range mod {
					assert.Equal(t, a.Transpose(n) == target, gotT.ContainsNumber(n))
					assert.Equal(t, a.Invert(n) == target, gotI.ContainsNumber(n))
				}
			}
		}
	}

	var zero pcset.Set
	s19 := pcset.MustEmpty(19)
	errPairs := []struct {
		name    string
		lhs     pcset.Set
		rhs     pcset.Set
		wantErr error
	}{
		{name: "invalid_lhs", lhs: zero, rhs: cMaj, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_rhs", lhs: cMaj, rhs: zero, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_lhs_over_mismatch", lhs: zero, rhs: s19, wantErr: pcset.ErrInvalidValue},
		{name: "modulus_mismatch", lhs: cMaj, rhs: s19, wantErr: pcset.ErrModulusMismatch},
	}

	for _, tc := range errPairs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.lhs.TranspositionIndicesTo(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustTranspositionIndicesTo(tc.rhs) }, tc.wantErr)

			_, err = tc.lhs.InversionIndicesTo(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustInversionIndicesTo(tc.rhs) }, tc.wantErr)
		})
	}
}

func TestEmbeddingsAndCoveringTranspositions(t *testing.T) {
	t.Parallel()

	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)
	x1CMaj := pcset.MustNew(12, 0, 4, 7)
	x2DMaj := pcset.MustNew(12, 2, 6, 9)
	x3Chromatic := pcset.MustNew(12, 0, 1, 2)
	x4Empty := pcset.MustEmpty(12)

	cases := []struct {
		name         string
		sub          pcset.Set
		wantCovering []int
		wantEmbedT   []int
		wantEmbedI   []int
	}{
		{
			name:         "c_major_in_major_scale",
			sub:          x1CMaj,
			wantCovering: []int{0, 5, 7},
			wantEmbedT:   []int{0, 5, 7},
			wantEmbedI:   []int{4, 9, 11},
		},
		{
			name:         "d_major_in_major_scale_asymmetric",
			sub:          x2DMaj,
			wantCovering: []int{2, 7, 9},
			wantEmbedT:   []int{3, 5, 10},
			wantEmbedI:   []int{1, 6, 11},
		},
		{
			name:         "chromatic_trichord_not_in_major_scale",
			sub:          x3Chromatic,
			wantCovering: []int{},
			wantEmbedT:   []int{},
			wantEmbedI:   []int{},
		},
		{
			name:         "empty_set_embedded_everywhere",
			sub:          x4Empty,
			wantCovering: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			wantEmbedT:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			wantEmbedI:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cov, err := tc.sub.CoveringTranspositionsOf(majorScale)
			require.NoError(t, err)
			assertSet(t, cov, 12, tc.wantCovering)
			assert.Equal(t, cov, tc.sub.MustCoveringTranspositionsOf(majorScale))

			embT, err := tc.sub.TranspositionEmbeddingsInto(majorScale)
			require.NoError(t, err)
			assertSet(t, embT, 12, tc.wantEmbedT)
			assert.Equal(t, embT, tc.sub.MustTranspositionEmbeddingsInto(majorScale))

			embI, err := tc.sub.InversionEmbeddingsInto(majorScale)
			require.NoError(t, err)
			assertSet(t, embI, 12, tc.wantEmbedI)
			assert.Equal(t, embI, tc.sub.MustInversionEmbeddingsInto(majorScale))

			// Verify duality: tau in Covering <=> (-tau) mod N in Embeddings, i.e., cov == embT.Invert(0).
			assert.Equal(t, embT.Invert(0), cov)

			// Verify intersection formula for non-empty sub:
			// {tau : X <= T_tau(S)} == intersection_{x in X} T_x(I_0(S)).
			if !tc.sub.IsEmpty() {
				inter := pcset.MustUniverse(12)
				invS := majorScale.Invert(0)
				for _, x := range tc.sub.Numbers() {
					inter = inter.MustIntersection(invS.Transpose(x))
				}
				assert.Equal(t, inter, cov)
			}

			for n := range 12 {
				assert.Equal(t, tc.sub.MustIsSubsetOf(majorScale.Transpose(n)), cov.ContainsNumber(n))
				assert.Equal(t, tc.sub.Transpose(n).MustIsSubsetOf(majorScale), embT.ContainsNumber(n))
				assert.Equal(t, tc.sub.Invert(n).MustIsSubsetOf(majorScale), embI.ContainsNumber(n))
			}
		})
	}

	// Verify for N > 64.
	s65 := pcset.MustNew(65, 0, 5, 12, 30, 64)
	x65 := pcset.MustNew(65, 2, 7, 14)
	cov65 := x65.MustCoveringTranspositionsOf(s65)
	embT65 := x65.MustTranspositionEmbeddingsInto(s65)
	embI65 := x65.MustInversionEmbeddingsInto(s65)
	assert.Equal(t, embT65.Invert(0), cov65)
	for n := range 65 {
		assert.Equal(t, x65.MustIsSubsetOf(s65.Transpose(n)), cov65.ContainsNumber(n))
		assert.Equal(t, x65.Transpose(n).MustIsSubsetOf(s65), embT65.ContainsNumber(n))
		assert.Equal(t, x65.Invert(n).MustIsSubsetOf(s65), embI65.ContainsNumber(n))
	}

	var zero pcset.Set
	s19 := pcset.MustEmpty(19)
	for _, tc := range []struct {
		name    string
		lhs     pcset.Set
		rhs     pcset.Set
		wantErr error
	}{
		{name: "invalid_lhs", lhs: zero, rhs: majorScale, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_rhs", lhs: x1CMaj, rhs: zero, wantErr: pcset.ErrInvalidValue},
		{name: "modulus_mismatch", lhs: x1CMaj, rhs: s19, wantErr: pcset.ErrModulusMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.lhs.CoveringTranspositionsOf(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustCoveringTranspositionsOf(tc.rhs) }, tc.wantErr)

			_, err = tc.lhs.TranspositionEmbeddingsInto(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustTranspositionEmbeddingsInto(tc.rhs) }, tc.wantErr)

			_, err = tc.lhs.InversionEmbeddingsInto(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustInversionEmbeddingsInto(tc.rhs) }, tc.wantErr)
		})
	}
}

func TestEquivalenceAndEmbeddingOperators(t *testing.T) {
	t.Parallel()

	cMaj := pcset.MustNew(12, 0, 4, 7)
	cMin := pcset.MustNew(12, 0, 3, 7)
	chrom := pcset.MustNew(12, 0, 1, 2)
	quartal := pcset.MustNew(12, 0, 5, 10)
	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)

	// {0, 4, 7} -> {0, 3, 7}: only I_7 in both T/I and TTO.
	eqTI, err := cMaj.EquivalenceOperatorsTI(cMin)
	require.NoError(t, err)
	assert.Equal(t, []pitchclass.Operator{mustOp(t, 11, 7, 12)}, eqTI)
	assert.Equal(t, eqTI, cMaj.MustEquivalenceOperatorsTI(cMin))

	eqTTO, err := cMaj.EquivalenceOperatorsTTO(cMin)
	require.NoError(t, err)
	assert.Equal(t, []pitchclass.Operator{mustOp(t, 11, 7, 12)}, eqTTO)
	assert.Equal(t, eqTTO, cMaj.MustEquivalenceOperatorsTTO(cMin))

	// {0, 1, 2} -> {0, 5, 10}: empty in T/I, [T_0 M_5, T_10 M_7] in TTO.
	chromTI, err := chrom.EquivalenceOperatorsTI(quartal)
	require.NoError(t, err)
	require.NotNil(t, chromTI)
	assert.Empty(t, chromTI)

	chromTTO, err := chrom.EquivalenceOperatorsTTO(quartal)
	require.NoError(t, err)
	assert.Equal(t, []pitchclass.Operator{
		mustOp(t, 5, 0, 12),
		mustOp(t, 7, 10, 12),
	}, chromTTO)
	assert.Equal(t, chromTTO, chrom.MustEquivalenceOperatorsTTO(quartal))

	// Embedding of C major into C major scale in T/I: [T_0, T_5, T_7, I_4, I_9, I_11].
	embTI, err := cMaj.EmbeddingOperatorsTI(majorScale)
	require.NoError(t, err)
	assert.Equal(t, []pitchclass.Operator{
		mustOp(t, 1, 0, 12),
		mustOp(t, 1, 5, 12),
		mustOp(t, 1, 7, 12),
		mustOp(t, 11, 4, 12),
		mustOp(t, 11, 9, 12),
		mustOp(t, 11, 11, 12),
	}, embTI)
	assert.Equal(t, embTI, cMaj.MustEmbeddingOperatorsTI(majorScale))

	// Exhaustive oracle comparison across moduli {1, 2, 7, 12}.
	for _, mod := range []int{1, 2, 7, 12} {
		x := pcset.MustReduce(mod, 0, 1, 3)
		y := pcset.MustReduce(mod, 0, 2, 3, 5, 6)
		z := x.Multiply(5).Transpose(2)

		for _, target := range []pcset.Set{x, y, z, pcset.MustEmpty(mod), pcset.MustUniverse(mod)} {
			assertOperatorListOracle(t, x, target, false, false, x.MustEquivalenceOperatorsTI(target))
			assertOperatorListOracle(t, x, target, true, false, x.MustEquivalenceOperatorsTTO(target))
			assertOperatorListOracle(t, x, target, false, true, x.MustEmbeddingOperatorsTI(target))
			assertOperatorListOracle(t, x, target, true, true, x.MustEmbeddingOperatorsTTO(target))
		}
	}

	var zero pcset.Set
	s19 := pcset.MustEmpty(19)
	for _, tc := range []struct {
		name    string
		lhs     pcset.Set
		rhs     pcset.Set
		wantErr error
	}{
		{name: "invalid_lhs", lhs: zero, rhs: cMaj, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_rhs", lhs: cMaj, rhs: zero, wantErr: pcset.ErrInvalidValue},
		{name: "modulus_mismatch", lhs: cMaj, rhs: s19, wantErr: pcset.ErrModulusMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.lhs.EquivalenceOperatorsTI(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustEquivalenceOperatorsTI(tc.rhs) }, tc.wantErr)

			_, err = tc.lhs.EquivalenceOperatorsTTO(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustEquivalenceOperatorsTTO(tc.rhs) }, tc.wantErr)

			_, err = tc.lhs.EmbeddingOperatorsTI(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustEmbeddingOperatorsTI(tc.rhs) }, tc.wantErr)

			_, err = tc.lhs.EmbeddingOperatorsTTO(tc.rhs)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustEmbeddingOperatorsTTO(tc.rhs) }, tc.wantErr)
		})
	}
}

func assertOperatorListOracle(
	t testing.TB,
	x, y pcset.Set,
	tto bool,
	embedding bool,
	got []pitchclass.Operator,
) {
	t.Helper()

	require.NotNil(t, got)
	mod := x.Modulus()

	seq := pitchclass.MustTranspositionsInversions(mod)
	if tto {
		seq = pitchclass.MustOperators(mod)
	}

	want := make([]pitchclass.Operator, 0)
	for op := range seq {
		img := x.MustApply(op)
		match := img == y
		if embedding {
			match = img.MustIsSubsetOf(y)
		}
		if match {
			want = append(want, op)
		}
	}

	slices.SortFunc(want, func(a, b pitchclass.Operator) int {
		return a.Compare(b)
	})
	assert.Equal(t, want, got)
}

func TestCommonToneCountsAndTheorems(t *testing.T) {
	t.Parallel()

	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)

	// K-18 canonical checks on major scale <2, 5, 4, 3, 6, 1>.
	assert.Equal(t, 7, res(majorScale.CommonToneCountTranspose(0)).get(t))
	assert.Equal(t, 7, majorScale.MustCommonToneCountTranspose(0))
	assert.Equal(t, 6, res(majorScale.CommonToneCountTranspose(7)).get(t))
	assert.Equal(t, 6, majorScale.MustCommonToneCountTranspose(5))
	assert.Equal(t, 5, majorScale.MustCommonToneCountTranspose(2))
	assert.Equal(t, 2, majorScale.MustCommonToneCountTranspose(6))
	assert.Equal(t, 6, majorScale.MustCommonToneCountTranspose(maxInt))
	assert.Equal(t, 3, majorScale.MustCommonToneCountTranspose(minInt))

	// Exhaustive verification of Common-Tone Theorems P-23 and P-24 for all sets of N in {1, 2, 7, 10, 12}.
	for _, mod := range []int{1, 2, 7, 10, 12} {
		totalSets := 1 << mod
		for mask := range totalSets {
			s := pcset.MustFromBitmask(uint64(mask), mod)
			elems := s.Numbers()
			icv := icvOracle(mod, elems)

			for n := range mod {
				gotT, err := s.CommonToneCountTranspose(n)
				require.NoError(t, err)
				assert.Equal(t, s.Transpose(n).MustIntersection(s).Cardinality(), gotT)
				assert.Equal(t, gotT, s.MustCommonToneCountTransposeWith(s, n))

				// P-23: Common tones under T_n from ICV.
				var wantT int
				switch {
				case n == 0:
					wantT = len(elems)
				case mod%2 == 0 && n == mod/2:
					wantT = 2 * icv[mod/2-1]
				default:
					d := min(n, mod-n)
					wantT = icv[d-1]
				}
				assert.Equal(t, wantT, gotT, "P-23 failed for mask=%d mod=%d n=%d", mask, mod, n)

				// P-24: Common tones under I_n from pair sums and fixed points.
				gotI, err := s.CommonToneCountInvert(n)
				require.NoError(t, err)
				assert.Equal(t, s.Invert(n).MustIntersection(s).Cardinality(), gotI)
				assert.Equal(t, gotI, s.MustCommonToneCountInvertWith(s, n))

				pairs, fixed := 0, 0
				for i, x := range elems {
					if addModOracle(x, x, mod) == n {
						fixed++
					}
					for j := i + 1; j < len(elems); j++ {
						if addModOracle(x, elems[j], mod) == n {
							pairs++
						}
					}
				}
				assert.Equal(t, 2*pairs+fixed, gotI, "P-24 failed for mask=%d mod=%d n=%d", mask, mod, n)
			}
		}
	}

	// Cross-set and general operator common tones (including all 48 operators at N=12 and N > 64).
	a := pcset.MustNew(12, 0, 1, 4, 7)
	b := pcset.MustNew(12, 0, 2, 4, 8, 10)
	for op := range pitchclass.MustOperators(12) {
		got, err := a.CommonToneCountWith(b, op)
		require.NoError(t, err)
		assert.Equal(t, a.MustApply(op).MustIntersection(b).Cardinality(), got)
		assert.Equal(t, got, a.MustCommonToneCountWith(b, op))
	}

	s65A := pcset.MustNew(65, 0, 10, 64)
	s65B := pcset.MustNew(65, 1, 11, 0)
	assert.Equal(t, 3, s65A.MustCommonToneCountTransposeWith(s65B, 1))
	assert.Equal(t, 2, s65A.MustCommonToneCountInvertWith(s65B, 10))
	assert.Equal(t, 3, s65A.MustCommonToneCountWith(s65B, mustOp(t, 1, 1, 65)))

	var zeroSet pcset.Set
	var zeroOp pitchclass.Operator
	s19 := pcset.MustEmpty(19)
	op12 := mustOp(t, 1, 0, 12)
	op19 := mustOp(t, 1, 0, 19)

	_, err := zeroSet.CommonToneCountTranspose(1)
	requireSentinelError(t, err, pcset.ErrInvalidValue)
	requirePanicsWithSentinel(t, func() { _ = zeroSet.MustCommonToneCountTranspose(1) }, pcset.ErrInvalidValue)

	_, err = zeroSet.CommonToneCountInvert(1)
	requireSentinelError(t, err, pcset.ErrInvalidValue)
	requirePanicsWithSentinel(t, func() { _ = zeroSet.MustCommonToneCountInvert(1) }, pcset.ErrInvalidValue)

	for _, tc := range []struct {
		name    string
		lhs     pcset.Set
		rhs     pcset.Set
		wantErr error
	}{
		{name: "invalid_lhs", lhs: zeroSet, rhs: a, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_rhs", lhs: a, rhs: zeroSet, wantErr: pcset.ErrInvalidValue},
		{name: "modulus_mismatch", lhs: a, rhs: s19, wantErr: pcset.ErrModulusMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.lhs.CommonToneCountTransposeWith(tc.rhs, 1)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() {
				_ = tc.lhs.MustCommonToneCountTransposeWith(tc.rhs, 1)
			}, tc.wantErr)

			_, err = tc.lhs.CommonToneCountInvertWith(tc.rhs, 1)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() {
				_ = tc.lhs.MustCommonToneCountInvertWith(tc.rhs, 1)
			}, tc.wantErr)
		})
	}

	for _, tc := range []struct {
		name    string
		lhs     pcset.Set
		rhs     pcset.Set
		op      pitchclass.Operator
		wantErr error
	}{
		{name: "invalid_lhs", lhs: zeroSet, rhs: a, op: op12, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_rhs", lhs: a, rhs: zeroSet, op: op12, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_op", lhs: a, rhs: b, op: zeroOp, wantErr: pcset.ErrInvalidValue},
		{name: "invalid_op_over_mismatch", lhs: a, rhs: s19, op: zeroOp, wantErr: pcset.ErrInvalidValue},
		{name: "set_modulus_mismatch", lhs: a, rhs: s19, op: op12, wantErr: pcset.ErrModulusMismatch},
		{name: "op_modulus_mismatch", lhs: a, rhs: b, op: op19, wantErr: pcset.ErrModulusMismatch},
	} {
		t.Run("with_op_"+tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.lhs.CommonToneCountWith(tc.rhs, tc.op)
			requireSentinelError(t, err, tc.wantErr)
			requirePanicsWithSentinel(t, func() {
				_ = tc.lhs.MustCommonToneCountWith(tc.rhs, tc.op)
			}, tc.wantErr)
		})
	}
}
