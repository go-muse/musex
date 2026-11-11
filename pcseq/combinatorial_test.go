package pcseq_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/setclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombinatorialFormsAndPairs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		row      pcseq.Row
		expected []string
	}{
		{
			name:     "Webern op24",
			row:      mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10),
			expected: []string{"P2", "P6", "P10", "I1", "I5", "I9", "RI3", "RI7", "RI11"},
		},
		{
			name:     "Webern op28",
			row:      mustRow(t, 12, 0, 11, 2, 1, 5, 6, 3, 4, 8, 7, 10, 9),
			expected: []string{"I9"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			forms, err := tt.row.CombinatorialForms()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, labelStrings(forms))
		})
	}

	row := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	for _, tt := range []struct {
		name string
		a    pcseq.FormLabel
		b    pcseq.FormLabel
	}{
		{name: "P-R", a: mustLabel(t, pcseq.P, 0, 12), b: mustLabel(t, pcseq.R, 0, 12)},
		{name: "I-RI", a: mustLabel(t, pcseq.I, 0, 12), b: mustLabel(t, pcseq.RI, 0, 12)},
		{name: "known P-P", a: mustLabel(t, pcseq.P, 0, 12), b: mustLabel(t, pcseq.P, 2, 12)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ok, err := row.IsCombinatorialPair(tt.a, tt.b)
			require.NoError(t, err)
			assert.True(t, ok)
		})
	}

	notCombinatorial, err := row.IsCombinatorialPair(mustLabel(t, pcseq.P, 0, 12), mustLabel(t, pcseq.P, 1, 12))
	require.NoError(t, err)
	assert.False(t, notCombinatorial)

	odd := mustRow(t, 3, 0, 1, 2)
	ok, err := odd.IsCombinatorialPair(mustLabel(t, pcseq.P, 0, 3), mustLabel(t, pcseq.R, 0, 3))
	require.NoError(t, err)
	assert.False(t, ok)
	forms, err := odd.CombinatorialForms()
	require.NoError(t, err)
	assert.NotNil(t, forms)
	assert.Empty(t, forms)
	rotations, err := odd.CombinatorialRotations()
	require.NoError(t, err)
	assert.NotNil(t, rotations)
	assert.Empty(t, rotations)

	_, err = row.IsCombinatorialPair(mustLabel(t, pcseq.P, 0, 19), mustLabel(t, pcseq.P, 0, 12))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = row.IsCombinatorialPair(pcseq.FormLabel{}, mustLabel(t, pcseq.P, 0, 12))
	assert.ErrorIs(t, err, pcseq.ErrInvalidFormLabel)
	_, err = pcseq.Row{}.CombinatorialForms()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestCombinatorialRotations(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 12, 0, 2, 6, 10, 7, 4, 11, 9, 8, 3, 5, 1)
	got, err := row.CombinatorialRotations()
	require.NoError(t, err)
	expected := []struct {
		rotation int
		forms    []string
	}{
		{rotation: 0, forms: []string{"I3"}},
		{rotation: 1, forms: []string{"I5"}},
		{rotation: 4, forms: []string{"I2"}},
		{rotation: 6, forms: []string{"I4"}},
		{rotation: 7, forms: []string{"I10"}},
		{rotation: 10, forms: []string{"I4"}},
	}
	require.Len(t, got, len(expected))
	for i := range expected {
		assert.Equal(t, expected[i].rotation, got[i].Rotation)
		assert.Equal(t, expected[i].forms, labelStrings(got[i].Forms))
	}
}

func TestHexachordAnalysisControls(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name           string
		content        pcset.Set
		transpositions []int
		inversions     []int
	}{
		{
			name:           "magic 014589",
			content:        mustSet(t, 12, 0, 1, 4, 5, 8, 9),
			transpositions: []int{2, 6, 10},
			inversions:     []int{3, 7, 11},
		},
		{
			name:           "whole tone",
			content:        mustSet(t, 12, 0, 2, 4, 6, 8, 10),
			transpositions: []int{1, 3, 5, 7, 9, 11},
			inversions:     []int{1, 3, 5, 7, 9, 11},
		},
		{
			name:           "Moses inversion",
			content:        mustSet(t, 12, 0, 1, 4, 5, 6, 7),
			transpositions: []int{},
			inversions:     []int{3},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			analysis, err := pcseq.AnalyzeHexachord(tt.content)
			require.NoError(t, err)
			assert.Equal(t, tt.transpositions, analysis.Transpositions)
			assert.Equal(t, tt.inversions, analysis.Inversions)
		})
	}

	op24 := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	analysis, err := op24.HexachordAnalysis()
	require.NoError(t, err)
	assert.Equal(t, []int{2, 6, 10}, analysis.Transpositions)
	assert.Equal(t, []int{1, 5, 9}, analysis.Inversions)
	assert.Equal(t, []int{3, 7, 11}, analysis.RetrogradeInversions)
	firstHalf, err := op24.Sequence().Segment(0, 6)
	require.NoError(t, err)
	content, err := firstHalf.Content()
	require.NoError(t, err)
	direct, err := pcseq.AnalyzeHexachord(content)
	require.NoError(t, err)
	assert.Equal(t, direct, analysis)

	_, err = pcseq.AnalyzeHexachord(pcset.Set{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.AnalyzeHexachord(mustSet(t, 11, 0, 1, 2, 3, 4))
	assert.ErrorIs(t, err, pcseq.ErrNotModulus12)
	_, err = pcseq.AnalyzeHexachord(mustSet(t, 12, 0, 1, 2, 3, 4))
	assert.ErrorIs(t, err, pcseq.ErrInvalidSegmentContent)
}

func TestHexachordClassificationAndZeroICV(t *testing.T) {
	t.Parallel()
	sequence, err := pcset.OfCardinality(12, 6)
	require.NoError(t, err)
	classes := make(map[setclass.SetClass]pcseq.HexachordAnalysis)
	for content := range sequence {
		analysis, err := pcseq.AnalyzeHexachord(content)
		require.NoError(t, err)
		hasZero := slices.Contains(content.IntervalVector().Slice(), 0)
		assert.Equal(t, hasZero, len(analysis.Transpositions) > 0,
			"transpositional combinatoriality must match a zero ICV component")
		class, err := setclass.New(content, setclass.TnIn)
		require.NoError(t, err)
		if _, exists := classes[class]; !exists {
			classes[class] = analysis
		}
	}
	assert.Len(t, classes, 50)

	var allCombinatorialOrders []int
	inversionOnly := 0
	transpositionOnly := 0
	retrogradeInversionOnly := 0
	for _, analysis := range classes {
		switch {
		case analysis.IsAllCombinatorial():
			allCombinatorialOrders = append(allCombinatorialOrders, len(analysis.Transpositions))
		case analysis.IsInversionOnly():
			inversionOnly++
		case analysis.IsTranspositionOnly():
			transpositionOnly++
		case analysis.IsRetrogradeInversionOnly():
			retrogradeInversionOnly++
		}
	}
	slices.Sort(allCombinatorialOrders)
	assert.Equal(t, []int{1, 1, 1, 2, 3, 6}, allCombinatorialOrders)
	assert.Equal(t, 13, inversionOnly)
	assert.Equal(t, 1, transpositionOnly)
	assert.Equal(t, 14, retrogradeInversionOnly)

	var empty pcseq.HexachordAnalysis
	assert.False(t, empty.IsAllCombinatorial())
	assert.False(t, empty.IsInversionOnly())
	assert.False(t, empty.IsTranspositionOnly())
	assert.False(t, empty.IsRetrogradeInversionOnly())
}

func TestAllIntervalProperty(t *testing.T) {
	t.Parallel()
	babbitt := mustRow(t, 12, 0, 9, 2, 10, 11, 1, 7, 5, 4, 8, 3, 6)
	intervals, err := babbitt.Sequence().Intervals()
	require.NoError(t, err)
	assert.Equal(t, []int{9, 5, 8, 1, 2, 6, 10, 11, 4, 7, 3}, classNumbers(intervals))
	assert.True(t, babbitt.IsAllInterval())
	assert.False(t, mustRow(t, 12, identityNumbers(12)...).IsAllInterval())
	assert.False(t, mustRow(t, 1, 0).IsAllInterval())
	assert.True(t, mustRow(t, 2, 0, 1).IsAllInterval())
	assert.True(t, mustRow(t, 2, 1, 0).IsAllInterval())
	assert.False(t, pcseq.Row{}.IsAllInterval())
}
