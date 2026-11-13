package frequency

import (
	"math"
	"math/big"

	"github.com/go-muse/musex/cents"
)

// ShiftOctaves returns f·2^k for any integer k. It composes cents.Octave,
// Cents.Mul, and Cents.ApplyTo rather than approximating a logarithm. It
// returns ErrInvalidFrequency for invalid f and ErrOutOfRange with
// cents.ErrOutOfRange when the generated result is not a normal frequency.
func ShiftOctaves(f float64, k int) (float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return 0, err
	}
	shift, err := cents.Octave().Mul(k)
	if err != nil {
		return 0, resultOutOfRange("octave shift")
	}
	result, err := shift.ApplyTo(f)
	if err != nil {
		return 0, resultOutOfRange("octave shift result")
	}
	return result, nil
}

// MustShiftOctaves calls ShiftOctaves and panics with its error, if any.
func MustShiftOctaves(f float64, k int) float64 {
	result, err := ShiftOctaves(f, k)
	if err != nil {
		panic(err)
	}
	return result
}

// NormalizeOctave returns representative = f·2^k in the exact half-open band
// [base, 2·base) and the integer shift k. The band and shift are determined
// from the exact binary inputs, without a cents grid or a float64 ratio. The
// representative is rounded and range-checked before conversion to float64.
// If no normal representative exists, it returns (0, 0, ErrOutOfRange) with
// cents.ErrOutOfRange. The result is a numeric octave equivalent, not a pitch
// class or a note name. Validation order is f, then base, then the result.
func NormalizeOctave(f, base float64) (representative float64, k int, err error) {
	if err := validateFrequency(f, 1); err != nil {
		return 0, 0, err
	}
	if err := validateFrequency(base, 2); err != nil {
		return 0, 0, err
	}

	fractionF, exponentF := math.Frexp(f)
	fractionBase, exponentBase := math.Frexp(base)
	q := exponentF - exponentBase
	if fractionF < fractionBase {
		q--
	}
	k = -q

	candidate := new(big.Rat).Set(binaryRat(f))
	power := new(big.Int).Lsh(big.NewInt(1), uint(absInt(k)))
	factor := new(big.Rat).SetInt(power)
	if k >= 0 {
		candidate.Mul(candidate, factor)
	} else {
		candidate.Quo(candidate, factor)
	}

	baseValue := binaryRat(base)
	twiceBase := new(big.Rat).Mul(baseValue, big.NewRat(2, 1))
	if candidate.Cmp(baseValue) < 0 || candidate.Cmp(twiceBase) >= 0 {
		return 0, 0, resultOutOfRange("octave normalization band")
	}
	representative, err = generatedFrequency(candidate, "octave normalization")
	if err != nil {
		return 0, 0, err
	}
	return representative, k, nil
}

// MustNormalizeOctave calls NormalizeOctave and panics with its error, if any.
// On success it returns the same representative and shift as NormalizeOctave.
func MustNormalizeOctave(f, base float64) (representative float64, k int) {
	representative, k, err := NormalizeOctave(f, base)
	if err != nil {
		panic(err)
	}
	return representative, k
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
