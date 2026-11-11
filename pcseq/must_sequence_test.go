package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceCoreMustVariants(t *testing.T) {
	t.Parallel()
	assert.True(t, pcseq.MustEmpty(12).IsEmpty())
	_, err := pcseq.Empty(0)
	assert.PanicsWithValue(t, err, func() { pcseq.MustEmpty(0) })

	assert.Equal(t, []int{11, 1}, pcseq.MustReduce(12, -1, 13).Numbers())
	_, err = pcseq.Reduce(0)
	assert.PanicsWithValue(t, err, func() { pcseq.MustReduce(0) })

	classes := []pitchclass.Class{mustClass(t, 0, 12), mustClass(t, 1, 12)}
	assert.Equal(t, []int{0, 1}, pcseq.MustFromClasses(classes...).Numbers())
	_, err = pcseq.FromClasses()
	assert.PanicsWithValue(t, err, func() { pcseq.MustFromClasses() })
	assert.Equal(t, []int{0, 1}, pcseq.MustFromClassesWithModulus(12, classes...).Numbers())
	_, err = pcseq.FromClassesWithModulus(0)
	assert.PanicsWithValue(t, err, func() { pcseq.MustFromClassesWithModulus(0) })

	s := mustSequence(t, 12, 0, 1, 2)
	assert.Equal(t, mustClass(t, 1, 12), s.MustAt(1))
	_, err = s.At(3)
	assert.PanicsWithValue(t, err, func() { s.MustAt(3) })
	assert.Equal(t, []int{1, 2, 3}, s.MustTranspose(1).Numbers())
	assert.Equal(t, []int{0, 11, 10}, s.MustInvert(0).Numbers())
	assert.Equal(t, []int{0, 2, 4}, s.MustMultiply(2).Numbers())
	assert.Equal(t, []int{0, 5, 10}, s.MustApply(mustOperator(t, 5, 0, 12)).Numbers())
	assert.Equal(t, []int{2, 1, 0}, s.MustRetrograde().Numbers())
	assert.Equal(t, []int{0, 1, 2}, s.MustRetrogradeInversion(2).Numbers())
	assert.Equal(t, []int{1, 2, 0}, s.MustRotate(1).Numbers())
	assert.Len(t, s.MustIntervals(), 2)
	assert.Equal(t, []int{0, 1, 2}, s.MustContent().Numbers())

	for name, call := range map[string]func() (pcseq.Sequence, error){
		"Transpose":           func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Transpose(1) },
		"Invert":              func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Invert(1) },
		"Multiply":            func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Multiply(2) },
		"Apply":               func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Apply(mustOperator(t, 1, 0, 12)) },
		"Retrograde":          func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Retrograde() },
		"RetrogradeInversion": func() (pcseq.Sequence, error) { return pcseq.Sequence{}.RetrogradeInversion(1) },
		"Rotate":              func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Rotate(1) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := call()
			assert.PanicsWithValue(t, err, func() {
				switch name {
				case "Transpose":
					pcseq.Sequence{}.MustTranspose(1)
				case "Invert":
					pcseq.Sequence{}.MustInvert(1)
				case "Multiply":
					pcseq.Sequence{}.MustMultiply(2)
				case "Apply":
					pcseq.Sequence{}.MustApply(mustOperator(t, 1, 0, 12))
				case "Retrograde":
					pcseq.Sequence{}.MustRetrograde()
				case "RetrogradeInversion":
					pcseq.Sequence{}.MustRetrogradeInversion(1)
				case "Rotate":
					pcseq.Sequence{}.MustRotate(1)
				}
			})
		})
	}
	_, err = pcseq.Sequence{}.Intervals()
	assert.PanicsWithValue(t, err, func() { pcseq.Sequence{}.MustIntervals() })
	_, err = pcseq.Sequence{}.Content()
	assert.PanicsWithValue(t, err, func() { pcseq.Sequence{}.MustContent() })
}

func TestRowCoreMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []int{0, 1, 2}, pcseq.MustReduceRow(3, 0, 1, 2).Numbers())
	_, err := pcseq.ReduceRow(0)
	assert.PanicsWithValue(t, err, func() { pcseq.MustReduceRow(0) })
	assert.Equal(t, []int{0, 1}, pcseq.MustRowFromClasses(mustClass(t, 0, 2), mustClass(t, 1, 2)).Numbers())
	_, err = pcseq.RowFromClasses()
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowFromClasses() })
	assert.Equal(t, []int{0, 1}, pcseq.MustRowFromClassesWithModulus(2, mustClass(t, 0, 2), mustClass(t, 1, 2)).Numbers())
	_, err = pcseq.RowFromClassesWithModulus(0)
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowFromClassesWithModulus(0) })
	r := mustRow(t, 3, 0, 1, 2)
	assert.Equal(t, r, pcseq.MustRowFromSequence(r.Sequence()))
	assert.Equal(t, mustClass(t, 2, 3), r.MustAt(2))
	_, err = r.At(3)
	assert.PanicsWithValue(t, err, func() { r.MustAt(3) })
	assert.Equal(t, []int{1, 2, 0}, r.MustTranspose(1).Numbers())
	assert.Equal(t, []int{0, 2, 1}, r.MustInvert(0).Numbers())
	assert.Equal(t, []int{0, 2, 1}, r.MustMultiply(2).Numbers())
	assert.Equal(t, []int{0, 2, 1}, r.MustApply(mustOperator(t, 2, 0, 3)).Numbers())
	assert.Equal(t, []int{2, 1, 0}, r.MustRetrograde().Numbers())
	assert.Equal(t, []int{0, 1, 2}, r.MustRetrogradeInversion(2).Numbers())
	assert.Equal(t, []int{1, 2, 0}, r.MustRotate(1).Numbers())

	for name, call := range map[string]func(){
		"Transpose":           func() { pcseq.Row{}.MustTranspose(1) },
		"Invert":              func() { pcseq.Row{}.MustInvert(1) },
		"Multiply":            func() { pcseq.Row{}.MustMultiply(2) },
		"Apply":               func() { pcseq.Row{}.MustApply(mustOperator(t, 1, 0, 12)) },
		"Retrograde":          func() { pcseq.Row{}.MustRetrograde() },
		"RetrogradeInversion": func() { pcseq.Row{}.MustRetrogradeInversion(1) },
		"Rotate":              func() { pcseq.Row{}.MustRotate(1) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var err error
			switch name {
			case "Transpose":
				_, err = pcseq.Row{}.Transpose(1)
			case "Invert":
				_, err = pcseq.Row{}.Invert(1)
			case "Multiply":
				_, err = pcseq.Row{}.Multiply(2)
			case "Apply":
				_, err = pcseq.Row{}.Apply(mustOperator(t, 1, 0, 12))
			case "Retrograde":
				_, err = pcseq.Row{}.Retrograde()
			case "RetrogradeInversion":
				_, err = pcseq.Row{}.RetrogradeInversion(1)
			case "Rotate":
				_, err = pcseq.Row{}.Rotate(1)
			}
			assert.PanicsWithValue(t, err, call)
		})
	}
}

func TestRowFormMustVariants(t *testing.T) {
	t.Parallel()
	r := mustRow(t, 3, 0, 1, 2)
	label := mustLabel(t, pcseq.P, 0, 3)
	assert.Equal(t, r, r.MustForm(label))
	assert.Equal(t, []string{"P0"}, labelStrings(r.MustLabelsOf(r)))
	_, err := r.Form(pcseq.FormLabel{})
	assert.PanicsWithValue(t, err, func() { r.MustForm(pcseq.FormLabel{}) })
	_, err = r.LabelsOf(pcseq.Row{})
	assert.PanicsWithValue(t, err, func() { r.MustLabelsOf(pcseq.Row{}) })
}

func TestMatrixMustVariants(t *testing.T) {
	t.Parallel()
	r := mustRow(t, 3, 1, 0, 2)
	matrix := pcseq.MustNewMatrix(r)
	assert.True(t, matrix.IsValid())
	_, err := pcseq.NewMatrix(pcseq.Row{})
	assert.PanicsWithValue(t, err, func() { pcseq.MustNewMatrix(pcseq.Row{}) })
	assert.Equal(t, 0, matrix.MustCell(0, 0).Number())
	assert.Equal(t, "P0", matrix.MustRowLabel(0).String())
	assert.Equal(t, "I0", matrix.MustColumnLabel(0).String())
	assert.Equal(t, []int{0, 2, 1}, matrix.MustRow(0).Numbers())
	assert.Equal(t, []int{0, 1, 2}, matrix.MustColumn(0).Numbers())
	assert.Equal(t, []int{1, 2, 0}, matrix.MustRetrogradeRow(0).Numbers())
	assert.Equal(t, []int{2, 1, 0}, matrix.MustRetrogradeInversionColumn(0).Numbers())
	assert.Len(t, matrix.MustCells(), 3)

	for name, call := range map[string]func() error{
		"Cell":                      func() error { _, err := matrix.Cell(3, 0); return err },
		"RowLabel":                  func() error { _, err := matrix.RowLabel(3); return err },
		"ColumnLabel":               func() error { _, err := matrix.ColumnLabel(3); return err },
		"Row":                       func() error { _, err := matrix.Row(3); return err },
		"Column":                    func() error { _, err := matrix.Column(3); return err },
		"RetrogradeRow":             func() error { _, err := matrix.RetrogradeRow(3); return err },
		"RetrogradeInversionColumn": func() error { _, err := matrix.RetrogradeInversionColumn(3); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := call()
			assert.PanicsWithValue(t, err, func() {
				switch name {
				case "Cell":
					matrix.MustCell(3, 0)
				case "RowLabel":
					matrix.MustRowLabel(3)
				case "ColumnLabel":
					matrix.MustColumnLabel(3)
				case "Row":
					matrix.MustRow(3)
				case "Column":
					matrix.MustColumn(3)
				case "RetrogradeRow":
					matrix.MustRetrogradeRow(3)
				case "RetrogradeInversionColumn":
					matrix.MustRetrogradeInversionColumn(3)
				}
			})
		})
	}
	_, err = pcseq.Matrix{}.Cells()
	assert.PanicsWithValue(t, err, func() { pcseq.Matrix{}.MustCells() })
}

func TestRowClassModelMustVariants(t *testing.T) {
	t.Parallel()
	r := mustRow(t, 2, 0, 1)
	class := pcseq.MustNewRowClass(r, pcseq.TIR)
	assert.Equal(t, 2, class.Size())
	_, err := pcseq.NewRowClass(r, pcseq.RowClassKind(0))
	assert.PanicsWithValue(t, err, func() { pcseq.MustNewRowClass(r, pcseq.RowClassKind(0)) })
	equivalent, err := pcseq.AreEquivalentRows(r, r, pcseq.TIR)
	require.NoError(t, err)
	assert.True(t, equivalent)
	assert.True(t, pcseq.MustAreEquivalentRows(r, r, pcseq.TIR))
	_, err = pcseq.AreEquivalentRows(r, mustRow(t, 3, 0, 1, 2), pcseq.TIR)
	assert.PanicsWithValue(t, err, func() { pcseq.MustAreEquivalentRows(r, mustRow(t, 3, 0, 1, 2), pcseq.TIR) })
	assert.Len(t, class.MustMembers(), 2)
	_, err = pcseq.RowClass{}.Members()
	assert.PanicsWithValue(t, err, func() { pcseq.RowClass{}.MustMembers() })
}
