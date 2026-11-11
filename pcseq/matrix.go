package pcseq

import "github.com/go-muse/musex/pitchclass"

// Matrix is the square twelve-tone matrix determined solely by a source row.
type Matrix struct {
	source Row
}

// NewMatrix returns the matrix of source.
func NewMatrix(source Row) (Matrix, error) { return Matrix{}, nil }

// IsValid reports whether m has a valid source within the matrix limit.
func (m Matrix) IsValid() bool { return false }

// Modulus returns the matrix modulus, or 0 for an invalid value.
func (m Matrix) Modulus() int { return 0 }

// Source returns the source row.
func (m Matrix) Source() Row { return Row{} }

// Cell returns m[i][j] = (p_j-p_i) mod N.
func (m Matrix) Cell(i, j int) (pitchclass.Class, error) { return pitchclass.Class{}, nil }

// RowLabel returns the P label of row i.
func (m Matrix) RowLabel(i int) (FormLabel, error) { return FormLabel{}, nil }

// ColumnLabel returns the I label of column j.
func (m Matrix) ColumnLabel(j int) (FormLabel, error) { return FormLabel{}, nil }

// Row returns matrix row i.
func (m Matrix) Row(i int) (Row, error) { return Row{}, nil }

// Column returns matrix column j.
func (m Matrix) Column(j int) (Row, error) { return Row{}, nil }

// RetrogradeRow returns matrix row i in reverse order.
func (m Matrix) RetrogradeRow(i int) (Row, error) { return Row{}, nil }

// RetrogradeInversionColumn returns matrix column j in reverse order.
func (m Matrix) RetrogradeInversionColumn(j int) (Row, error) { return Row{}, nil }

// Cells returns a newly allocated square matrix of classes.
func (m Matrix) Cells() ([][]pitchclass.Class, error) { return nil, nil }

// String returns Matrix(source), or Matrix(invalid).
func (m Matrix) String() string { return "" }
