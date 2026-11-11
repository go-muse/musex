package pcseq

import (
	"math/big"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// MustEmpty is like Empty but panics on error.
func MustEmpty(modulus int) Sequence { return must(Empty(modulus)) }

// MustNew is like New but panics on error.
func MustNew(modulus int, numbers ...int) Sequence { return must(New(modulus, numbers...)) }

// MustReduce is like Reduce but panics on error.
func MustReduce(modulus int, xs ...int) Sequence { return must(Reduce(modulus, xs...)) }

// MustFromClasses is like FromClasses but panics on error.
func MustFromClasses(classes ...pitchclass.Class) Sequence { return must(FromClasses(classes...)) }

// MustFromClassesWithModulus is like FromClassesWithModulus but panics on error.
func MustFromClassesWithModulus(modulus int, classes ...pitchclass.Class) Sequence {
	return must(FromClassesWithModulus(modulus, classes...))
}

// MustAt is like Sequence.At but panics on error.
func (s Sequence) MustAt(index int) pitchclass.Class { return must(s.At(index)) }

// MustTranspose is like Sequence.Transpose but panics on error.
func (s Sequence) MustTranspose(n int) Sequence { return must(s.Transpose(n)) }

// MustInvert is like Sequence.Invert but panics on error.
func (s Sequence) MustInvert(n int) Sequence { return must(s.Invert(n)) }

// MustMultiply is like Sequence.Multiply but panics on error.
func (s Sequence) MustMultiply(m int) Sequence { return must(s.Multiply(m)) }

// MustApply is like Sequence.Apply but panics on error.
func (s Sequence) MustApply(op pitchclass.Operator) Sequence { return must(s.Apply(op)) }

// MustRetrograde is like Sequence.Retrograde but panics on error.
func (s Sequence) MustRetrograde() Sequence { return must(s.Retrograde()) }

// MustRetrogradeInversion is like Sequence.RetrogradeInversion but panics on error.
func (s Sequence) MustRetrogradeInversion(n int) Sequence {
	return must(s.RetrogradeInversion(n))
}

// MustRotate is like Sequence.Rotate but panics on error.
func (s Sequence) MustRotate(k int) Sequence { return must(s.Rotate(k)) }

// MustIntervals is like Sequence.Intervals but panics on error.
func (s Sequence) MustIntervals() []pitchclass.Class { return must(s.Intervals()) }

// MustContent is like Sequence.Content but panics on error.
func (s Sequence) MustContent() pcset.Set { return must(s.Content()) }

// MustSegment is like Sequence.Segment but panics on error.
func (s Sequence) MustSegment(start, length int) Sequence { return must(s.Segment(start, length)) }

// MustOverlappingSegments is like Sequence.OverlappingSegments but panics on error.
func (s Sequence) MustOverlappingSegments(length int) []Sequence {
	return must(s.OverlappingSegments(length))
}

// MustDiscreteSegments is like Sequence.DiscreteSegments but panics on error.
func (s Sequence) MustDiscreteSegments(length int) []Sequence {
	return must(s.DiscreteSegments(length))
}

// MustContentClass is like Sequence.ContentClass but panics on error.
func (s Sequence) MustContentClass(kind setclass.Kind) setclass.SetClass {
	return must(s.ContentClass(kind))
}

// MustInvariancesTo is like Sequence.InvariancesTo but panics on error.
func (s Sequence) MustInvariancesTo(other Sequence) SegmentInvariance {
	return must(s.InvariancesTo(other))
}

// MustNewRow is like NewRow but panics on error.
func MustNewRow(modulus int, numbers ...int) Row { return must(NewRow(modulus, numbers...)) }

// MustReduceRow is like ReduceRow but panics on error.
func MustReduceRow(modulus int, xs ...int) Row { return must(ReduceRow(modulus, xs...)) }

// MustRowFromClasses is like RowFromClasses but panics on error.
func MustRowFromClasses(classes ...pitchclass.Class) Row { return must(RowFromClasses(classes...)) }

// MustRowFromClassesWithModulus is like RowFromClassesWithModulus but panics on error.
func MustRowFromClassesWithModulus(modulus int, classes ...pitchclass.Class) Row {
	return must(RowFromClassesWithModulus(modulus, classes...))
}

// MustRowFromSequence is like RowFromSequence but panics on error.
func MustRowFromSequence(s Sequence) Row { return must(RowFromSequence(s)) }

// MustAt is like Row.At but panics on error.
func (r Row) MustAt(index int) pitchclass.Class { return must(r.At(index)) }

// MustTranspose is like Row.Transpose but panics on error.
func (r Row) MustTranspose(n int) Row { return must(r.Transpose(n)) }

// MustInvert is like Row.Invert but panics on error.
func (r Row) MustInvert(n int) Row { return must(r.Invert(n)) }

// MustMultiply is like Row.Multiply but panics on error.
func (r Row) MustMultiply(m int) Sequence { return must(r.Multiply(m)) }

// MustApply is like Row.Apply but panics on error.
func (r Row) MustApply(op pitchclass.Operator) Row { return must(r.Apply(op)) }

// MustRetrograde is like Row.Retrograde but panics on error.
func (r Row) MustRetrograde() Row { return must(r.Retrograde()) }

// MustRetrogradeInversion is like Row.RetrogradeInversion but panics on error.
func (r Row) MustRetrogradeInversion(n int) Row { return must(r.RetrogradeInversion(n)) }

// MustRotate is like Row.Rotate but panics on error.
func (r Row) MustRotate(k int) Row { return must(r.Rotate(k)) }

// MustForm is like Row.Form but panics on error.
func (r Row) MustForm(label FormLabel) Row { return must(r.Form(label)) }

// MustLabelsOf is like Row.LabelsOf but panics on error.
func (r Row) MustLabelsOf(other Row) []FormLabel { return must(r.LabelsOf(other)) }

// MustIsDerived is like Row.IsDerived but panics on error.
func (r Row) MustIsDerived(length int) bool { return must(r.IsDerived(length)) }

// MustDerive is like Row.Derive but panics on error.
func (r Row) MustDerive(length int) [][]DerivationOperator { return must(r.Derive(length)) }

// MustIsCombinatorialPair is like Row.IsCombinatorialPair but panics on error.
func (r Row) MustIsCombinatorialPair(a, b FormLabel) bool {
	return must(r.IsCombinatorialPair(a, b))
}

// MustCombinatorialForms is like Row.CombinatorialForms but panics on error.
func (r Row) MustCombinatorialForms() []FormLabel { return must(r.CombinatorialForms()) }

// MustCombinatorialRotations is like Row.CombinatorialRotations but panics on error.
func (r Row) MustCombinatorialRotations() []RotationCombinatoriality {
	return must(r.CombinatorialRotations())
}

// MustHexachordAnalysis is like Row.HexachordAnalysis but panics on error.
func (r Row) MustHexachordAnalysis() HexachordAnalysis { return must(r.HexachordAnalysis()) }

// MustNewFormLabel is like NewFormLabel but panics on error.
func MustNewFormLabel(kind FormKind, index, modulus int) FormLabel {
	return must(NewFormLabel(kind, index, modulus))
}

// MustParseFormLabel is like ParseFormLabel but panics on error.
func MustParseFormLabel(s string, modulus int) FormLabel { return must(ParseFormLabel(s, modulus)) }

// MustNewMatrix is like NewMatrix but panics on error.
func MustNewMatrix(source Row) Matrix { return must(NewMatrix(source)) }

// MustCell is like Matrix.Cell but panics on error.
func (m Matrix) MustCell(i, j int) pitchclass.Class { return must(m.Cell(i, j)) }

// MustRowLabel is like Matrix.RowLabel but panics on error.
func (m Matrix) MustRowLabel(i int) FormLabel { return must(m.RowLabel(i)) }

// MustColumnLabel is like Matrix.ColumnLabel but panics on error.
func (m Matrix) MustColumnLabel(j int) FormLabel { return must(m.ColumnLabel(j)) }

// MustRow is like Matrix.Row but panics on error.
func (m Matrix) MustRow(i int) Row { return must(m.Row(i)) }

// MustColumn is like Matrix.Column but panics on error.
func (m Matrix) MustColumn(j int) Row { return must(m.Column(j)) }

// MustRetrogradeRow is like Matrix.RetrogradeRow but panics on error.
func (m Matrix) MustRetrogradeRow(i int) Row { return must(m.RetrogradeRow(i)) }

// MustRetrogradeInversionColumn is like Matrix.RetrogradeInversionColumn but panics on error.
func (m Matrix) MustRetrogradeInversionColumn(j int) Row {
	return must(m.RetrogradeInversionColumn(j))
}

// MustCells is like Matrix.Cells but panics on error.
func (m Matrix) MustCells() [][]pitchclass.Class { return must(m.Cells()) }

// MustNewRowClass is like NewRowClass but panics on error.
func MustNewRowClass(source Row, kind RowClassKind) RowClass {
	return must(NewRowClass(source, kind))
}

// MustAreEquivalentRows is like AreEquivalentRows but panics on error.
func MustAreEquivalentRows(a, b Row, kind RowClassKind) bool {
	return must(AreEquivalentRows(a, b, kind))
}

// MustMembers is like RowClass.Members but panics on error.
func (c RowClass) MustMembers() []Row { return must(c.Members()) }

// MustNewDerivationOperator is like NewDerivationOperator but panics on error.
func MustNewDerivationOperator(family DerivationFamily, index, modulus int) DerivationOperator {
	return must(NewDerivationOperator(family, index, modulus))
}

// MustApply is like DerivationOperator.Apply but panics on error.
func (o DerivationOperator) MustApply(s Sequence) Sequence { return must(o.Apply(s)) }

// MustAnalyzeHexachord is like AnalyzeHexachord but panics on error.
func MustAnalyzeHexachord(content pcset.Set) HexachordAnalysis {
	return must(AnalyzeHexachord(content))
}

// MustRowCount is like RowCount but panics on error.
func MustRowCount(modulus int) int { return must(RowCount(modulus)) }

// MustRowCountBig is like RowCountBig but panics on error.
func MustRowCountBig(modulus int) *big.Int { return must(RowCountBig(modulus)) }

// MustRowClassCount is like RowClassCount but panics on error.
func MustRowClassCount(modulus int, kind RowClassKind) int {
	return must(RowClassCount(modulus, kind))
}

// MustRowClassCountBig is like RowClassCountBig but panics on error.
func MustRowClassCountBig(modulus int, kind RowClassKind) *big.Int {
	return must(RowClassCountBig(modulus, kind))
}

// MustCountRowsWithSegmentLengths is like CountRowsWithSegmentLengths but panics on error.
func MustCountRowsWithSegmentLengths(modulus int, lengths ...int) int {
	return must(CountRowsWithSegmentLengths(modulus, lengths...))
}

// MustCountRowsWithSegmentLengthsBig is like CountRowsWithSegmentLengthsBig but panics on error.
func MustCountRowsWithSegmentLengthsBig(modulus int, lengths ...int) *big.Int {
	return must(CountRowsWithSegmentLengthsBig(modulus, lengths...))
}

// MustCountRowsWithSegmentContents is like CountRowsWithSegmentContents but panics on error.
func MustCountRowsWithSegmentContents(modulus int, contents ...pcset.Set) int {
	return must(CountRowsWithSegmentContents(modulus, contents...))
}

// MustCountRowsWithSegmentContentsBig is like CountRowsWithSegmentContentsBig but panics on error.
func MustCountRowsWithSegmentContentsBig(modulus int, contents ...pcset.Set) *big.Int {
	return must(CountRowsWithSegmentContentsBig(modulus, contents...))
}

// MustAllIntervalCount is like AllIntervalCount but panics on error.
func MustAllIntervalCount(modulus int, kind AllIntervalCountKind) int {
	return must(AllIntervalCount(modulus, kind))
}

// MustAllIntervalCountBig is like AllIntervalCountBig but panics on error.
func MustAllIntervalCountBig(modulus int, kind AllIntervalCountKind) *big.Int {
	return must(AllIntervalCountBig(modulus, kind))
}

// MustParse is like Parse but panics on error.
func MustParse(s string) Sequence { return must(Parse(s)) }

// MustNotation is like Sequence.Notation but panics on error.
func (s Sequence) MustNotation(style pitchclass.Style) string { return must(s.Notation(style)) }

// MustCompactNotation is like Sequence.CompactNotation but panics on error.
func (s Sequence) MustCompactNotation(style pitchclass.Style) string {
	return must(s.CompactNotation(style))
}

// MustParseNotation is like ParseNotation but panics on error.
func MustParseNotation(s string, style pitchclass.Style, modulus int) Sequence {
	return must(ParseNotation(s, style, modulus))
}

// MustParseRow is like ParseRow but panics on error.
func MustParseRow(s string) Row { return must(ParseRow(s)) }

// MustNotation is like Row.Notation but panics on error.
func (r Row) MustNotation(style pitchclass.Style) string { return must(r.Notation(style)) }

// MustCompactNotation is like Row.CompactNotation but panics on error.
func (r Row) MustCompactNotation(style pitchclass.Style) string {
	return must(r.CompactNotation(style))
}

// MustParseRowNotation is like ParseRowNotation but panics on error.
func MustParseRowNotation(s string, style pitchclass.Style, modulus int) Row {
	return must(ParseRowNotation(s, style, modulus))
}
