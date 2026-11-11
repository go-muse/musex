package pcseq

import (
	"math/big"

	"github.com/go-muse/musex/pcset"
)

// RowCount returns modulus factorial when it fits in int.
func RowCount(modulus int) (int, error) { return 0, nil }

// RowCountBig returns exact modulus factorial in a new big integer.
func RowCountBig(modulus int) (*big.Int, error) { return nil, nil }

// RowClassCount returns a supported row-class count when it fits in int.
func RowClassCount(modulus int, kind RowClassKind) (int, error) { return 0, nil }

// RowClassCountBig returns an exact supported row-class count.
func RowClassCountBig(modulus int, kind RowClassKind) (*big.Int, error) { return nil, nil }

// CountRowsWithSegmentLengths returns the product of segment-length factorials.
func CountRowsWithSegmentLengths(modulus int, lengths ...int) (int, error) { return 0, nil }

// CountRowsWithSegmentLengthsBig returns the exact product of length factorials.
func CountRowsWithSegmentLengthsBig(modulus int, lengths ...int) (*big.Int, error) {
	return nil, nil
}

// CountRowsWithSegmentContents returns the product of segment-content cardinalities factorial.
func CountRowsWithSegmentContents(modulus int, contents ...pcset.Set) (int, error) {
	return 0, nil
}

// CountRowsWithSegmentContentsBig returns the exact product for concrete contents.
func CountRowsWithSegmentContentsBig(modulus int, contents ...pcset.Set) (*big.Int, error) {
	return nil, nil
}

// AllIntervalCount returns a confirmed count of all-interval rows of modulus 12.
func AllIntervalCount(modulus int, kind AllIntervalCountKind) (int, error) { return 0, nil }

// AllIntervalCountBig returns a confirmed exact count of all-interval rows.
func AllIntervalCountBig(modulus int, kind AllIntervalCountKind) (*big.Int, error) {
	return nil, nil
}
