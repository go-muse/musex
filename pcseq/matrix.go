package pcseq

import "github.com/go-muse/musex/pitchclass"

// Matrix is the square twelve-tone matrix determined solely by a source row.
type Matrix struct {
	source Row
}

// NewMatrix returns the matrix of source.
func NewMatrix(source Row) (Matrix, error) {
	if !source.IsValid() {
		return Matrix{}, invalidValue()
	}
	if err := checkModulusLimit(source.modulus, MaxMatrixModulus); err != nil {
		return Matrix{}, err
	}
	return Matrix{source: source}, nil
}

// IsValid reports whether m has a valid source within the matrix limit.
func (m Matrix) IsValid() bool {
	return m.source.IsValid() && m.source.modulus <= MaxMatrixModulus
}

// Modulus returns the matrix modulus, or 0 for an invalid value.
func (m Matrix) Modulus() int { return m.source.modulus }

// Source returns the source row.
func (m Matrix) Source() Row { return m.source }

func (m Matrix) checkIndices(i, j int) error {
	if !m.IsValid() {
		return invalidValue()
	}
	if i < 0 || i >= m.source.modulus || j < 0 || j >= m.source.modulus {
		return invalidSegment()
	}
	return nil
}

// Cell returns m[i][j] = (p_j-p_i) mod N.
func (m Matrix) Cell(i, j int) (pitchclass.Class, error) {
	if err := m.checkIndices(i, j); err != nil {
		return pitchclass.Class{}, err
	}
	return pitchclass.MustNew(classAt(m.source.modulus, m.source.elements, i).MustIntervalTo(
		classAt(m.source.modulus, m.source.elements, j)), m.source.modulus), nil
}

// RowLabel returns the P label of row i.
func (m Matrix) RowLabel(i int) (FormLabel, error) {
	if err := m.checkIndices(i, 0); err != nil {
		return FormLabel{}, err
	}
	index := classAt(m.source.modulus, m.source.elements, i).MustIntervalTo(
		classAt(m.source.modulus, m.source.elements, 0))
	return NewFormLabel(P, index, m.source.modulus)
}

// ColumnLabel returns the I label of column j.
func (m Matrix) ColumnLabel(j int) (FormLabel, error) {
	if err := m.checkIndices(0, j); err != nil {
		return FormLabel{}, err
	}
	index := classAt(m.source.modulus, m.source.elements, 0).MustIntervalTo(
		classAt(m.source.modulus, m.source.elements, j))
	return NewFormLabel(I, index, m.source.modulus)
}

// Row returns matrix row i.
func (m Matrix) Row(i int) (Row, error) {
	label, err := m.RowLabel(i)
	if err != nil {
		return Row{}, err
	}
	return m.source.Form(label)
}

// Column returns matrix column j.
func (m Matrix) Column(j int) (Row, error) {
	label, err := m.ColumnLabel(j)
	if err != nil {
		return Row{}, err
	}
	return m.source.Form(label)
}

// RetrogradeRow returns matrix row i in reverse order.
func (m Matrix) RetrogradeRow(i int) (Row, error) {
	label, err := m.RowLabel(i)
	if err != nil {
		return Row{}, err
	}
	label.kind = R
	return m.source.Form(label)
}

// RetrogradeInversionColumn returns matrix column j in reverse order.
func (m Matrix) RetrogradeInversionColumn(j int) (Row, error) {
	label, err := m.ColumnLabel(j)
	if err != nil {
		return Row{}, err
	}
	label.kind = RI
	return m.source.Form(label)
}

// Cells returns a newly allocated square matrix of classes.
func (m Matrix) Cells() ([][]pitchclass.Class, error) {
	if !m.IsValid() {
		return nil, invalidValue()
	}
	n := m.source.modulus
	cells := make([][]pitchclass.Class, n)
	for i := range n {
		cells[i] = make([]pitchclass.Class, n)
		for j := range n {
			cells[i][j] = pitchclass.MustNew(classAt(m.source.modulus, m.source.elements, i).MustIntervalTo(
				classAt(m.source.modulus, m.source.elements, j)), m.source.modulus)
		}
	}
	return cells, nil
}

// String returns Matrix(source), or Matrix(invalid).
func (m Matrix) String() string {
	if !m.IsValid() {
		return "Matrix(invalid)"
	}
	return "Matrix(" + m.source.String() + ")"
}
