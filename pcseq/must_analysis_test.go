package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"

	"github.com/stretchr/testify/assert"
)

func TestDerivationMustVariants(t *testing.T) {
	t.Parallel()
	op := pcseq.MustNewDerivationOperator(pcseq.DerivationT, 1, 4)
	assert.Equal(t, "T1", op.String())
	_, err := pcseq.NewDerivationOperator(pcseq.DerivationFamily(0), 0, 4)
	assert.PanicsWithValue(t, err, func() { pcseq.MustNewDerivationOperator(pcseq.DerivationFamily(0), 0, 4) })
	assert.Equal(t, []int{1, 2, 3}, op.MustApply(mustSequence(t, 4, 0, 1, 2)).Numbers())
	_, err = op.Apply(mustSequence(t, 12))
	assert.PanicsWithValue(t, err, func() { op.MustApply(mustSequence(t, 12)) })
	row := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	derived, err := row.IsDerived(3)
	assert.NoError(t, err)
	assert.True(t, derived)
	assert.True(t, row.MustIsDerived(3))
	assert.Len(t, row.MustDerive(3), 4)
	_, err = row.IsDerived(5)
	assert.PanicsWithValue(t, err, func() { row.MustIsDerived(5) })
}

func TestCombinatorialMustVariants(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	assert.True(t, row.MustIsCombinatorialPair(mustLabel(t, pcseq.P, 0, 12), mustLabel(t, pcseq.R, 0, 12)))
	assert.Len(t, row.MustCombinatorialForms(), 9)
	assert.NotNil(t, row.MustCombinatorialRotations())
	analysis := row.MustHexachordAnalysis()
	assert.Equal(t, []int{2, 6, 10}, analysis.Transpositions)
	direct := pcseq.MustAnalyzeHexachord(mustSet(t, 12, 0, 1, 4, 5, 8, 9))
	assert.Equal(t, []int{2, 6, 10}, direct.Transpositions)

	_, err := pcseq.Row{}.IsCombinatorialPair(mustLabel(t, pcseq.P, 0, 12), mustLabel(t, pcseq.R, 0, 12))
	assert.PanicsWithValue(t, err, func() {
		pcseq.Row{}.MustIsCombinatorialPair(mustLabel(t, pcseq.P, 0, 12), mustLabel(t, pcseq.R, 0, 12))
	})
	_, err = pcseq.Row{}.CombinatorialForms()
	assert.PanicsWithValue(t, err, func() { pcseq.Row{}.MustCombinatorialForms() })
	_, err = pcseq.Row{}.CombinatorialRotations()
	assert.PanicsWithValue(t, err, func() { pcseq.Row{}.MustCombinatorialRotations() })
	_, err = pcseq.Row{}.HexachordAnalysis()
	assert.PanicsWithValue(t, err, func() { pcseq.Row{}.MustHexachordAnalysis() })
	_, err = pcseq.AnalyzeHexachord(mustSet(t, 12, 0, 1))
	assert.PanicsWithValue(t, err, func() { pcseq.MustAnalyzeHexachord(mustSet(t, 12, 0, 1)) })
}

func TestRowCountMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 479_001_600, pcseq.MustRowCount(12))
	assert.Equal(t, "479001600", pcseq.MustRowCountBig(12).String())
	_, err := pcseq.RowCount(21)
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowCount(21) })
	_, err = pcseq.RowCountBig(0)
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowCountBig(0) })
}

func TestRowClassCountMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 9_985_920, pcseq.MustRowClassCount(12, pcseq.TIR))
	assert.Equal(t, "9985920", pcseq.MustRowClassCountBig(12, pcseq.TIR).String())
	_, err := pcseq.RowClassCount(23, pcseq.TIR)
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowClassCount(23, pcseq.TIR) })
	_, err = pcseq.RowClassCountBig(11, pcseq.TTOR)
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowClassCountBig(11, pcseq.TTOR) })
}

func TestCountRowsMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 518_400, pcseq.MustCountRowsWithSegmentLengths(12, 6, 6))
	assert.Equal(t, "518400", pcseq.MustCountRowsWithSegmentLengthsBig(12, 6, 6).String())
	contents := partitionContents(t, 12, 6, 6)
	assert.Equal(t, 518_400, pcseq.MustCountRowsWithSegmentContents(12, contents...))
	assert.Equal(t, "518400", pcseq.MustCountRowsWithSegmentContentsBig(12, contents...).String())
	_, err := pcseq.CountRowsWithSegmentLengths(12, 5, 5)
	assert.PanicsWithValue(t, err, func() { pcseq.MustCountRowsWithSegmentLengths(12, 5, 5) })
	_, err = pcseq.CountRowsWithSegmentLengthsBig(12, 6, 0, 6)
	assert.PanicsWithValue(t, err, func() { pcseq.MustCountRowsWithSegmentLengthsBig(12, 6, 0, 6) })
	_, err = pcseq.CountRowsWithSegmentContents(12)
	assert.PanicsWithValue(t, err, func() { pcseq.MustCountRowsWithSegmentContents(12) })
	_, err = pcseq.CountRowsWithSegmentContentsBig(12)
	assert.PanicsWithValue(t, err, func() { pcseq.MustCountRowsWithSegmentContentsBig(12) })
}

func TestAllIntervalCountMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 46_272, pcseq.MustAllIntervalCount(12, pcseq.AllIntervalTotal))
	assert.Equal(t, "46272", pcseq.MustAllIntervalCountBig(12, pcseq.AllIntervalTotal).String())
	_, err := pcseq.AllIntervalCount(11, pcseq.AllIntervalTotal)
	assert.PanicsWithValue(t, err, func() { pcseq.MustAllIntervalCount(11, pcseq.AllIntervalTotal) })
	_, err = pcseq.AllIntervalCountBig(12, pcseq.AllIntervalCountKind(0))
	assert.PanicsWithValue(t, err, func() { pcseq.MustAllIntervalCountBig(12, pcseq.AllIntervalCountKind(0)) })
}
