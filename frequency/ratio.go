package frequency

import (
	"math/big"
	"strconv"
)

// Ratio returns the reduced positive rational ratio f2/f1. Each frequency is
// interpreted as the exact rational value of its shortest decimal
// representation, not as its binary float64 value or a nearby simple ratio.
// It validates f1 and then f2 and returns ErrInvalidFrequency for the first
// invalid input. The result is a new *big.Rat owned by the caller; later calls
// are independent of any modification of earlier results.
func Ratio(f1, f2 float64) (*big.Rat, error) {
	if err := validateFrequency(f1, 1); err != nil {
		return nil, err
	}
	if err := validateFrequency(f2, 2); err != nil {
		return nil, err
	}
	return new(big.Rat).Quo(decimalRat(f2), decimalRat(f1)), nil
}

// MustRatio calls Ratio and panics with its error, if any. The returned
// *big.Rat has the same independent ownership as the result of Ratio.
func MustRatio(f1, f2 float64) *big.Rat {
	ratio, err := Ratio(f1, f2)
	if err != nil {
		panic(err)
	}
	return ratio
}

func binaryRat(f float64) *big.Rat {
	return new(big.Rat).SetFloat64(f)
}

func decimalRat(f float64) *big.Rat {
	text := strconv.FormatFloat(f, 'g', -1, 64)
	value, _ := new(big.Rat).SetString(text)
	return value
}
