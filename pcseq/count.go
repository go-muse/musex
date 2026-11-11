package pcseq

import (
	"math/big"

	"github.com/go-muse/musex/pcset"
)

func factorialBig(n int) *big.Int {
	result := big.NewInt(1)
	for factor := 2; factor <= n; factor++ {
		result.Mul(result, big.NewInt(int64(factor)))
	}
	return result
}

func bigToInt(value *big.Int) (int, error) {
	if !value.IsInt64() {
		return 0, ErrOverflow
	}
	return int(value.Int64()), nil
}

// RowCount returns modulus factorial when it fits in int.
func RowCount(modulus int) (int, error) {
	value, err := RowCountBig(modulus)
	if err != nil {
		return 0, err
	}
	return bigToInt(value)
}

// RowCountBig returns exact modulus factorial in a new big integer.
func RowCountBig(modulus int) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}
	return factorialBig(modulus), nil
}

// RowClassCount returns a supported row-class count when it fits in int.
func RowClassCount(modulus int, kind RowClassKind) (int, error) {
	value, err := RowClassCountBig(modulus, kind)
	if err != nil {
		return 0, err
	}
	return bigToInt(value)
}

// RowClassCountBig returns an exact supported row-class count.
func RowClassCountBig(modulus int, kind RowClassKind) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}
	if !kind.IsValid() {
		return nil, invalidValue()
	}
	if kind != TIR {
		if modulus != 12 {
			return nil, ErrNotModulus12
		}
		if kind == TIRRotations {
			return big.NewInt(836_017), nil
		}
		return big.NewInt(4_995_360), nil
	}
	if err := checkModulusLimit(modulus, MaxAnalysisModulus); err != nil {
		return nil, err
	}
	if modulus <= 2 {
		return big.NewInt(1), nil
	}
	result := factorialBig(modulus - 1)
	product := big.NewInt(1)
	if modulus%2 != 0 {
		for factor := modulus - 1; factor >= 2; factor -= 2 {
			product.Mul(product, big.NewInt(int64(factor)))
		}
	} else {
		product.Mul(product, big.NewInt(int64(modulus/2+1)))
		for factor := modulus - 2; factor >= 2; factor -= 2 {
			product.Mul(product, big.NewInt(int64(factor)))
		}
	}
	result.Add(result, product)
	return result.Quo(result, big.NewInt(4)), nil
}

// CountRowsWithSegmentLengths returns the product of segment-length factorials.
func CountRowsWithSegmentLengths(modulus int, lengths ...int) (int, error) {
	value, err := CountRowsWithSegmentLengthsBig(modulus, lengths...)
	if err != nil {
		return 0, err
	}
	return bigToInt(value)
}

// CountRowsWithSegmentLengthsBig returns the exact product of length factorials.
func CountRowsWithSegmentLengthsBig(modulus int, lengths ...int) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}
	if len(lengths) == 0 {
		return nil, invalidSegmentContent()
	}
	remaining := modulus
	for _, length := range lengths {
		if length < 1 || length > remaining {
			return nil, invalidSegmentContent()
		}
		remaining -= length
	}
	if remaining != 0 {
		return nil, invalidSegmentContent()
	}
	result := big.NewInt(1)
	for _, length := range lengths {
		result.Mul(result, factorialBig(length))
	}
	return result, nil
}

// CountRowsWithSegmentContents returns the product of segment-content cardinalities factorial.
func CountRowsWithSegmentContents(modulus int, contents ...pcset.Set) (int, error) {
	value, err := CountRowsWithSegmentContentsBig(modulus, contents...)
	if err != nil {
		return 0, err
	}
	return bigToInt(value)
}

// CountRowsWithSegmentContentsBig returns the exact product for concrete contents.
func CountRowsWithSegmentContentsBig(modulus int, contents ...pcset.Set) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}
	for _, content := range contents {
		if !content.IsValid() {
			return nil, invalidValue()
		}
		if content.Modulus() != modulus {
			return nil, modulusMismatch(modulus, content.Modulus())
		}
	}
	if len(contents) == 0 {
		return nil, invalidSegmentContent()
	}
	union := pcset.MustEmpty(modulus)
	lengths := make([]int, len(contents))
	remaining := modulus
	for i, content := range contents {
		cardinality := content.Cardinality()
		if cardinality < 1 || cardinality > remaining {
			return nil, invalidSegmentContent()
		}
		disjoint, err := union.IsDisjoint(content)
		if err != nil || !disjoint {
			return nil, invalidSegmentContent()
		}
		union = union.MustUnion(content)
		lengths[i] = cardinality
		remaining -= cardinality
	}
	if remaining != 0 || !union.IsUniverse() {
		return nil, invalidSegmentContent()
	}
	return CountRowsWithSegmentLengthsBig(modulus, lengths...)
}

// AllIntervalCount returns a confirmed count of all-interval rows of modulus 12.
func AllIntervalCount(modulus int, kind AllIntervalCountKind) (int, error) {
	value, err := AllIntervalCountBig(modulus, kind)
	if err != nil {
		return 0, err
	}
	return bigToInt(value)
}

// AllIntervalCountBig returns a confirmed exact count of all-interval rows.
func AllIntervalCountBig(modulus int, kind AllIntervalCountKind) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}
	if !kind.IsValid() {
		return nil, invalidValue()
	}
	if modulus != 12 {
		return nil, ErrNotModulus12
	}
	values := [...]int64{46_272, 3_856, 1_928, 1_008, 88, 0}
	return big.NewInt(values[int(kind)-int(AllIntervalTotal)]), nil
}
