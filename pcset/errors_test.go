package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrorsDistinctness(t *testing.T) {
	t.Parallel()

	sentinels := allSentinelErrors()
	require.Len(t, sentinels, 11)

	for i, a := range sentinels {
		require.Error(t, a, "sentinel %d must not be nil", i)
		assert.NotEmpty(t, a.Error(), "sentinel %d message must not be empty", i)
		requireSentinelError(t, a, a)

		for j, b := range sentinels {
			if i == j {
				assert.ErrorIs(t, a, b)
			} else {
				assert.NotErrorIs(t, a, b, "sentinel %d (%v) must not match %d (%v)", i, a, j, b)
			}
		}
	}
}

func TestAll66MustVariantsPanicContract(t *testing.T) {
	t.Parallel()

	var zeroSet pcset.Set
	var zeroIV pcset.IntervalVector
	var zeroInvV pcset.InvarianceVector
	var zeroClass pitchclass.Class
	var zeroOp pitchclass.Operator
	valid12 := pcset.MustNew(12, 0, 4, 7)
	valid19 := pcset.MustEmpty(19)

	cases := []struct {
		name    string
		fn      func()
		wantErr error
	}{
		{name: "01_MustEmpty", fn: func() { _ = pcset.MustEmpty(0) }, wantErr: pcset.ErrInvalidModulus},
		{name: "02_MustUniverse", fn: func() { _ = pcset.MustUniverse(65_537) }, wantErr: pcset.ErrModulusTooLarge},
		{name: "03_MustNew", fn: func() { _ = pcset.MustNew(12, 12) }, wantErr: pcset.ErrInvalidClass},
		{name: "04_MustReduce", fn: func() { _ = pcset.MustReduce(0, 1) }, wantErr: pcset.ErrInvalidModulus},
		{name: "05_MustFromClasses", fn: func() { _ = pcset.MustFromClasses() }, wantErr: pcset.ErrInvalidClass},
		{
			name:    "06_MustFromClassesWithModulus",
			fn:      func() { _ = pcset.MustFromClassesWithModulus(12, zeroClass) },
			wantErr: pcset.ErrInvalidValue,
		},
		{name: "07_MustFromBitmask", fn: func() { _ = pcset.MustFromBitmask(0, 65) }, wantErr: pcset.ErrOverflow},
		{name: "08_MustFromBitmaskBig", fn: func() { _ = pcset.MustFromBitmaskBig(nil, 12) }, wantErr: pcset.ErrInvalidValue},
		{name: "09_Set_MustBitmask", fn: func() { _ = pcset.MustEmpty(65).MustBitmask() }, wantErr: pcset.ErrOverflow},
		{name: "10_Set_MustBitmaskBig", fn: func() { _ = zeroSet.MustBitmaskBig() }, wantErr: pcset.ErrInvalidValue},
		{name: "11_Set_MustInvertAround", fn: func() { _ = valid12.MustInvertAround(zeroClass) }, wantErr: pcset.ErrInvalidValue},
		{
			name:    "12_Set_MustInvertSwapping",
			fn:      func() { _ = valid12.MustInvertSwapping(zeroClass, zeroClass) },
			wantErr: pcset.ErrInvalidValue,
		},
		{name: "13_Set_MustApply", fn: func() { _ = valid12.MustApply(zeroOp) }, wantErr: pcset.ErrInvalidValue},
		{name: "14_MustOperatorFixedClasses", fn: func() { _ = pcset.MustOperatorFixedClasses(zeroOp) }, wantErr: pcset.ErrInvalidValue},
		{
			name:    "15_MustMultiplicationFixedClasses",
			fn:      func() { _ = pcset.MustMultiplicationFixedClasses(1, 0) },
			wantErr: pcset.ErrInvalidModulus,
		},
		{name: "16_Set_MustUnion", fn: func() { _ = valid12.MustUnion(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{name: "17_Set_MustIntersection", fn: func() { _ = valid12.MustIntersection(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{name: "18_Set_MustDifference", fn: func() { _ = valid12.MustDifference(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{
			name:    "19_Set_MustSymmetricDifference",
			fn:      func() { _ = valid12.MustSymmetricDifference(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{name: "20_Set_MustIsDisjoint", fn: func() { _ = valid12.MustIsDisjoint(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{name: "21_Set_MustIsSubsetOf", fn: func() { _ = valid12.MustIsSubsetOf(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{name: "22_Set_MustIsProperSubsetOf", fn: func() { _ = valid12.MustIsProperSubsetOf(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{name: "23_Set_MustIsSupersetOf", fn: func() { _ = valid12.MustIsSupersetOf(valid19) }, wantErr: pcset.ErrModulusMismatch},
		{
			name:    "24_Set_MustIsProperSupersetOf",
			fn:      func() { _ = valid12.MustIsProperSupersetOf(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "25_Set_MustTranspositionIndicesTo",
			fn:      func() { _ = valid12.MustTranspositionIndicesTo(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "26_Set_MustInversionIndicesTo",
			fn:      func() { _ = valid12.MustInversionIndicesTo(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "27_Set_MustTranspositionEmbeddingsInto",
			fn:      func() { _ = valid12.MustTranspositionEmbeddingsInto(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "28_Set_MustCoveringTranspositionsOf",
			fn:      func() { _ = valid12.MustCoveringTranspositionsOf(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "29_Set_MustInversionEmbeddingsInto",
			fn:      func() { _ = valid12.MustInversionEmbeddingsInto(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "30_Set_MustEquivalenceOperatorsTI",
			fn:      func() { _ = valid12.MustEquivalenceOperatorsTI(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "31_Set_MustEquivalenceOperatorsTTO",
			fn:      func() { _ = valid12.MustEquivalenceOperatorsTTO(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "32_Set_MustEmbeddingOperatorsTI",
			fn:      func() { _ = valid12.MustEmbeddingOperatorsTI(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "33_Set_MustEmbeddingOperatorsTTO",
			fn:      func() { _ = valid12.MustEmbeddingOperatorsTTO(valid19) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "34_Set_MustCommonToneCountTranspose",
			fn:      func() { _ = zeroSet.MustCommonToneCountTranspose(1) },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "35_Set_MustCommonToneCountInvert",
			fn:      func() { _ = zeroSet.MustCommonToneCountInvert(1) },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "36_Set_MustCommonToneCountTransposeWith",
			fn:      func() { _ = valid12.MustCommonToneCountTransposeWith(valid19, 1) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "37_Set_MustCommonToneCountInvertWith",
			fn:      func() { _ = valid12.MustCommonToneCountInvertWith(valid19, 1) },
			wantErr: pcset.ErrModulusMismatch,
		},
		{
			name:    "38_Set_MustCommonToneCountWith",
			fn:      func() { _ = valid12.MustCommonToneCountWith(valid12, zeroOp) },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "39_Set_MustNormalFormWith",
			fn:      func() { _ = valid12.MustNormalFormWith(pcset.Algorithm(99)) },
			wantErr: pcset.ErrInvalidAlgorithm,
		},
		{
			name:    "40_Set_MustTranspositionSymmetryIndices",
			fn:      func() { _ = zeroSet.MustTranspositionSymmetryIndices() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "41_Set_MustInversionSymmetryIndices",
			fn:      func() { _ = zeroSet.MustInversionSymmetryIndices() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "42_Set_MustSymmetryOperatorsTI",
			fn:      func() { _ = zeroSet.MustSymmetryOperatorsTI() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "43_Set_MustSymmetryOperatorsTTO",
			fn:      func() { _ = zeroSet.MustSymmetryOperatorsTTO() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "44_Set_MustDegreeOfTranspositionSymmetry",
			fn:      func() { _ = zeroSet.MustDegreeOfTranspositionSymmetry() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "45_Set_MustDegreeOfInversionSymmetry",
			fn:      func() { _ = zeroSet.MustDegreeOfInversionSymmetry() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "46_Set_MustDegreeOfSymmetryTI",
			fn:      func() { _ = zeroSet.MustDegreeOfSymmetryTI() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "47_Set_MustDegreeOfSymmetryTTO",
			fn:      func() { _ = zeroSet.MustDegreeOfSymmetryTTO() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "48_Set_MustDistinctTranspositionCount",
			fn:      func() { _ = zeroSet.MustDistinctTranspositionCount() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "49_Set_MustOrbitSizeTI",
			fn:      func() { _ = zeroSet.MustOrbitSizeTI() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "50_IntervalVector_MustCount",
			fn:      func() { _ = valid12.IntervalVector().MustCount(7) },
			wantErr: pcset.ErrInvalidIntervalClass,
		},
		{
			name:    "51_IntervalVector_MustNotation",
			fn:      func() { _ = zeroIV.MustNotation() },
			wantErr: pcset.ErrInvalidValue,
		},
		{
			name:    "52_IntervalVector_MustCompactNotation",
			fn:      func() { _ = pcset.MustUniverse(12).IntervalVector().MustCompactNotation() },
			wantErr: pcset.ErrNotation,
		},
		{
			name:    "53_Set_MustInvarianceVector",
			fn:      func() { _ = valid19.MustInvarianceVector() },
			wantErr: pcset.ErrNotModulus12,
		},
		{
			name:    "54_InvarianceVector_MustNotation",
			fn:      func() { _ = zeroInvV.MustNotation() },
			wantErr: pcset.ErrInvalidValue,
		},
		{name: "55_MustAll", fn: func() { _ = pcset.MustAll(0) }, wantErr: pcset.ErrInvalidModulus},
		{
			name:    "56_MustOfCardinality",
			fn:      func() { _ = pcset.MustOfCardinality(12, 13) },
			wantErr: pcset.ErrInvalidCardinality,
		},
		{name: "57_MustCount", fn: func() { _ = pcset.MustCount(63) }, wantErr: pcset.ErrOverflow},
		{name: "58_MustCountBig", fn: func() { _ = pcset.MustCountBig(0) }, wantErr: pcset.ErrInvalidModulus},
		{
			name:    "59_MustCountOfCardinality",
			fn:      func() { _ = pcset.MustCountOfCardinality(67, 33) },
			wantErr: pcset.ErrOverflow,
		},
		{
			name:    "60_MustCountOfCardinalityBig",
			fn:      func() { _ = pcset.MustCountOfCardinalityBig(12, -1) },
			wantErr: pcset.ErrInvalidCardinality,
		},
		{name: "61_MustParse", fn: func() { _ = pcset.MustParse("invalid") }, wantErr: pcset.ErrNotation},
		{
			name:    "62_Set_MustNotation",
			fn:      func() { _ = valid19.MustNotation(pitchclass.TE) },
			wantErr: pcset.ErrNotation,
		},
		{
			name:    "63_Set_MustCompactNotation",
			fn:      func() { _ = valid12.MustCompactNotation(pitchclass.Decimal) },
			wantErr: pcset.ErrNotation,
		},
		{
			name:    "64_Set_MustNormalFormNotation",
			fn:      func() { _ = valid12.MustNormalFormNotation(pitchclass.TE, pcset.Algorithm(99)) },
			wantErr: pcset.ErrInvalidAlgorithm,
		},
		{
			name:    "65_Set_MustNormalFormCompactNotation",
			fn:      func() { _ = valid12.MustNormalFormCompactNotation(pitchclass.Decimal, pcset.Rahn) },
			wantErr: pcset.ErrNotation,
		},
		{
			name:    "66_MustParseNotation",
			fn:      func() { _ = pcset.MustParseNotation("{047}", pitchclass.Decimal, 12) },
			wantErr: pcset.ErrNotation,
		},
	}

	require.Len(t, cases, 66)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requirePanicsWithSentinel(t, tc.fn, tc.wantErr)
		})
	}
}
