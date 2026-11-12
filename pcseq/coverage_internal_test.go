package pcseq

import (
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/setclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func coverageRow(modulus int, numbers ...int) Row {
	return newRowUnchecked(modulus, encodeNumbers(numbers))
}

func coverageSequence(modulus int, numbers ...int) Sequence {
	return newSequenceUnchecked(modulus, encodeNumbers(numbers))
}

func coverageValidRow(modulus int) Row {
	numbers := make([]int, modulus)
	for i := range modulus {
		numbers[i] = i
	}
	return coverageRow(modulus, numbers...)
}

func coverageSet(t *testing.T, modulus int, numbers ...int) pcset.Set {
	t.Helper()
	set, err := pcset.New(modulus, numbers...)
	require.NoError(t, err)
	return set
}

func coverageEmpty(t *testing.T, modulus int) Sequence {
	t.Helper()
	sequence, err := Empty(modulus)
	require.NoError(t, err)
	return sequence
}

func coverageNew(t *testing.T, modulus int, numbers ...int) Sequence {
	t.Helper()
	sequence, err := New(modulus, numbers...)
	require.NoError(t, err)
	return sequence
}

func coverageLabel(kind FormKind, index, modulus int) FormLabel {
	return FormLabel{kind: kind, index: index, modulus: modulus}
}

func coverageOperator(family DerivationFamily, index, modulus int) DerivationOperator {
	return DerivationOperator{family: family, index: index, modulus: modulus}
}

func TestInternalFirstHalfContentScenarios(t *testing.T) {
	t.Parallel()
	t.Run("FHC-1 invalid row", func(t *testing.T) {
		t.Parallel()
		got, err := firstHalfContent(Row{})
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.Equal(t, pcset.Set{}, got)
	})
	for _, tt := range []struct {
		name     string
		row      Row
		expected pcset.Set
	}{
		{name: "FHC-2 even row", row: coverageRow(4, 0, 1, 2, 3), expected: coverageSet(t, 4, 0, 1)},
		{name: "FHC-3 odd row rounds down", row: coverageRow(3, 0, 1, 2), expected: coverageSet(t, 3, 0)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := firstHalfContent(tt.row)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.expected))
		})
	}
}

func TestInternalCombinatorialContentScenarios(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		a        Row
		b        Row
		expected bool
		target   error
	}{
		{name: "CC-1 odd first row skips second", a: coverageValidRow(3), b: Row{}, expected: false},
		{name: "CC-2 invalid first row", a: Row{}, b: coverageRow(2, 0, 1), expected: false, target: ErrInvalidValue},
		{name: "CC-3 invalid second row", a: coverageRow(2, 0, 1), b: Row{}, expected: false, target: ErrInvalidValue},
		{name: "CC-4 complementary halves", a: coverageRow(2, 0, 1), b: coverageRow(2, 1, 0), expected: true},
		{name: "CC-5 same halves", a: coverageRow(2, 0, 1), b: coverageRow(2, 0, 1), expected: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := combinatorialContent(tt.a, tt.b)
			if tt.target == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.target)
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestInternalAnalyzeHexachordScenarios(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		content pcset.Set
		target  error
	}{
		{name: "HA-1 invalid set", content: pcset.Set{}, target: ErrInvalidValue},
		{name: "HA-2 modulus before cardinality", content: coverageSet(t, 11, 0, 1, 2, 3, 4), target: ErrNotModulus12},
		{name: "HA-3 wrong cardinality", content: coverageSet(t, 12, 0, 1, 2, 3, 4), target: ErrInvalidSegmentContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := AnalyzeHexachord(tt.content)
			assert.ErrorIs(t, err, tt.target)
		})
	}

	for _, tt := range []struct {
		name              string
		content           pcset.Set
		transpositions    []int
		inversions        []int
		retrograde        []int
		all               bool
		inversionOnly     bool
		transpositionOnly bool
		riOnly            bool
	}{
		{name: "HA-4 all combinatorial", content: coverageSet(t, 12, 0, 1, 2, 3, 4, 5), transpositions: []int{6}, inversions: []int{11}, retrograde: []int{5}, all: true},
		{name: "HA-5 transposition only", content: coverageSet(t, 12, 0, 1, 2, 4, 5, 9), transpositions: []int{6}, inversions: []int{}, retrograde: []int{}, transpositionOnly: true},
		{name: "HA-6 inversion only", content: coverageSet(t, 12, 0, 1, 2, 3, 4, 6), transpositions: []int{}, inversions: []int{11}, retrograde: []int{}, inversionOnly: true},
		{name: "HA-7 retrograde inversion only", content: coverageSet(t, 12, 0, 1, 2, 3, 4, 8), transpositions: []int{}, inversions: []int{}, retrograde: []int{4}, riOnly: true},
		{name: "HA-8 no combinatoriality", content: coverageSet(t, 12, 0, 1, 2, 3, 4, 7), transpositions: []int{}, inversions: []int{}, retrograde: []int{}},
		{name: "HA-9 whole tone", content: coverageSet(t, 12, 0, 2, 4, 6, 8, 10), transpositions: []int{1, 3, 5, 7, 9, 11}, inversions: []int{1, 3, 5, 7, 9, 11}, retrograde: []int{0, 2, 4, 6, 8, 10}, all: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := AnalyzeHexachord(tt.content)
			require.NoError(t, err)
			assert.Equal(t, tt.transpositions, got.Transpositions)
			assert.Equal(t, tt.inversions, got.Inversions)
			assert.Equal(t, tt.retrograde, got.RetrogradeInversions)
			assert.NotNil(t, got.Transpositions)
			assert.NotNil(t, got.Inversions)
			assert.NotNil(t, got.RetrogradeInversions)
			assert.Equal(t, tt.all, got.IsAllCombinatorial())
			assert.Equal(t, tt.inversionOnly, got.IsInversionOnly())
			assert.Equal(t, tt.transpositionOnly, got.IsTranspositionOnly())
			assert.Equal(t, tt.riOnly, got.IsRetrogradeInversionOnly())
		})
	}
}

func TestInternalRowHexachordAnalysisScenarios(t *testing.T) {
	t.Parallel()
	_, err := Row{}.HexachordAnalysis()
	assert.ErrorIs(t, err, ErrInvalidValue)
	_, err = coverageValidRow(3).HexachordAnalysis()
	assert.ErrorIs(t, err, ErrNotModulus12)
	got, err := coverageValidRow(12).HexachordAnalysis()
	require.NoError(t, err)
	assert.Equal(t, []int{6}, got.Transpositions)
	assert.Equal(t, []int{11}, got.Inversions)
	assert.Equal(t, []int{5}, got.RetrogradeInversions)
}

func TestInternalCombinatorialFormsScenarios(t *testing.T) {
	t.Parallel()
	forms, err := Row{}.CombinatorialForms()
	assert.ErrorIs(t, err, ErrInvalidValue)
	assert.Nil(t, forms)
	forms, err = coverageValidRow(1_025).CombinatorialForms()
	assert.ErrorIs(t, err, ErrModulusTooLarge)
	assert.Nil(t, forms)
	forms, err = coverageValidRow(3).CombinatorialForms()
	require.NoError(t, err)
	assert.NotNil(t, forms)
	assert.Empty(t, forms)
	forms, err = coverageRow(2, 0, 1).CombinatorialForms()
	require.NoError(t, err)
	assert.Equal(t, []FormLabel{
		coverageLabel(P, 1, 2),
		coverageLabel(I, 1, 2),
		coverageLabel(RI, 0, 2),
	}, forms)
}

func TestInternalCombinatorialRotationsScenarios(t *testing.T) {
	t.Parallel()
	rotations, err := Row{}.CombinatorialRotations()
	assert.ErrorIs(t, err, ErrInvalidValue)
	assert.Nil(t, rotations)
	rotations, err = coverageValidRow(1_025).CombinatorialRotations()
	assert.ErrorIs(t, err, ErrModulusTooLarge)
	assert.Nil(t, rotations)
	rotations, err = coverageValidRow(3).CombinatorialRotations()
	require.NoError(t, err)
	assert.NotNil(t, rotations)
	assert.Empty(t, rotations)
	rotations, err = coverageRow(2, 0, 1).CombinatorialRotations()
	require.NoError(t, err)
	assert.Equal(t, []RotationCombinatoriality{
		{Rotation: 0, Forms: []FormLabel{coverageLabel(I, 1, 2)}},
		{Rotation: 1, Forms: []FormLabel{coverageLabel(I, 0, 2)}},
	}, rotations)
}

func TestInternalContentClassScenarios(t *testing.T) {
	t.Parallel()
	_, err := Sequence{}.ContentClass(setclass.TnIn)
	assert.ErrorIs(t, err, ErrInvalidValue)
	_, err = coverageEmpty(t, 12).ContentClass(setclass.Kind(0))
	assert.ErrorIs(t, err, ErrInvalidValue)
	_, err = coverageEmpty(t, 257).ContentClass(setclass.TTO)
	assert.ErrorIs(t, err, ErrModulusTooLarge)
	class, err := coverageEmpty(t, 257).ContentClass(setclass.TnIn)
	require.NoError(t, err)
	assert.Equal(t, 257, class.Modulus())
	assert.Zero(t, class.Cardinality())
	assert.True(t, class.IsEmpty())

	for _, kind := range []setclass.Kind{setclass.Tn, setclass.TnIn, setclass.TTO} {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()
			class, err := coverageNew(t, 12, 0).ContentClass(kind)
			require.NoError(t, err)
			assert.Equal(t, kind, class.Kind())
			assert.Equal(t, 12, class.Modulus())
			assert.Equal(t, 1, class.Cardinality())
			assert.True(t, class.PrimeForm().Equal(coverageSet(t, 12, 0)))
		})
	}
}

func TestInternalInvariancesToScenarios(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		s      Sequence
		other  Sequence
		target error
	}{
		{name: "SI-1 invalid receiver", s: Sequence{}, other: coverageEmpty(t, 12), target: ErrInvalidValue},
		{name: "SI-2 invalid other", s: coverageEmpty(t, 12), other: Sequence{}, target: ErrInvalidValue},
		{name: "SI-3 modulus before length", s: coverageEmpty(t, 12), other: coverageEmpty(t, 19), target: ErrModulusMismatch},
		{name: "SI-4 different lengths", s: coverageEmpty(t, 12), other: coverageNew(t, 12, 0), target: ErrInvalidSegment},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.s.InvariancesTo(tt.other)
			assert.ErrorIs(t, err, tt.target)
		})
	}

	got, err := coverageEmpty(t, 3).InvariancesTo(coverageEmpty(t, 3))
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2}, got.Transpositions)
	assert.Equal(t, []int{0, 1, 2}, got.Inversions)
	got, err = coverageNew(t, 12, 0, 11, 2, 1).InvariancesTo(coverageNew(t, 12, 5, 6, 3, 4))
	require.NoError(t, err)
	assert.Equal(t, []int{4}, got.Transpositions)
	assert.Equal(t, []int{5}, got.Inversions)
	got, err = coverageSequence(4, 0, 1).InvariancesTo(coverageSequence(4, 0, 2))
	require.NoError(t, err)
	assert.NotNil(t, got.Transpositions)
	assert.NotNil(t, got.Inversions)
	assert.Empty(t, got.Transpositions)
	assert.Empty(t, got.Inversions)
}

func TestInternalDerivationApplyScenarios(t *testing.T) {
	t.Parallel()
	_, err := DerivationOperator{}.Apply(coverageNew(t, 12, 0))
	assert.ErrorIs(t, err, ErrInvalidValue)
	_, err = coverageOperator(DerivationT, 0, 12).Apply(Sequence{})
	assert.ErrorIs(t, err, ErrInvalidValue)
	_, err = coverageOperator(DerivationT, 0, 12).Apply(coverageEmpty(t, 19))
	assert.ErrorIs(t, err, ErrModulusMismatch)

	for _, tt := range []struct {
		name     string
		op       DerivationOperator
		expected []int
	}{
		{name: "DA-4 transpose", op: coverageOperator(DerivationT, 2, 12), expected: []int{2, 3, 4}},
		{name: "DA-5 inversion", op: coverageOperator(DerivationI, 2, 12), expected: []int{2, 1, 0}},
		{name: "DA-6 transposed retrograde", op: coverageOperator(DerivationTR, 2, 12), expected: []int{4, 3, 2}},
		{name: "DA-7 inverted retrograde", op: coverageOperator(DerivationIR, 2, 12), expected: []int{0, 1, 2}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.op.Apply(coverageNew(t, 12, 0, 1, 2))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got.Numbers())
		})
	}
}

func TestInternalDerivationStringScenarios(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "DerivationOperator(invalid)", DerivationOperator{}.String())
	for _, tt := range []struct {
		op       DerivationOperator
		expected string
	}{
		{op: coverageOperator(DerivationT, 2, 12), expected: "T2"},
		{op: coverageOperator(DerivationI, 3, 12), expected: "I3"},
		{op: coverageOperator(DerivationTR, 4, 12), expected: "T4R"},
		{op: coverageOperator(DerivationIR, 5, 12), expected: "I5R"},
	} {
		assert.Equal(t, tt.expected, tt.op.String())
	}
}

func TestInternalRowIsValidScenarios(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		row      Row
		expected bool
	}{
		{name: "RV-1 zero modulus", row: Row{}, expected: false},
		{name: "RV-2 modulus above limit", row: Row{modulus: MaxModulus + 1, elements: ""}, expected: false},
		{name: "RV-3 odd byte count", row: Row{modulus: 1, elements: "x"}, expected: false},
		{name: "RV-4 wrong length", row: coverageRow(2, 0), expected: false},
		{name: "RV-5 number above modulus", row: coverageRow(2, 0, 2), expected: false},
		{name: "RV-6 duplicate", row: coverageRow(2, 0, 0), expected: false},
		{name: "RV-7 complete permutation", row: coverageRow(2, 0, 1), expected: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, rowIsValid(tt.row))
		})
	}
}

func TestInternalMinimalRotationScenarios(t *testing.T) {
	t.Parallel()
	assert.Equal(t, Row{}, minimalRotation(Row{}))
	for _, tt := range []struct {
		name     string
		row      Row
		expected []int
	}{
		{name: "MR-2 single", row: coverageRow(1, 0), expected: []int{0}},
		{name: "MR-3 repeated values", row: coverageRow(2, 0, 0), expected: []int{0, 0}},
		{name: "MR-4 mixed comparisons", row: coverageRow(3, 1, 2, 0), expected: []int{0, 1, 2}},
		{name: "MR-5 greater and lower", row: coverageRow(3, 2, 1, 0), expected: []int{0, 2, 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, minimalRotation(tt.row).Numbers())
		})
	}
}
