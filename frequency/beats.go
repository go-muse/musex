package frequency

import (
	"math"
	"math/big"
)

// BeatFrequency returns |f2-f1| hertz for two simple tones. The exact binary
// inputs are subtracted exactly and rounded directly to the nearest float64,
// without an intermediate 53-bit rounding; zero and subnormal results are
// valid. The result is not a measure of consonance or a hearing threshold.
// Zero means no periodic beating, not absence of sound. Validation of f1 and
// then f2 is the only source of errors.
func BeatFrequency(f1, f2 float64) (float64, error) {
	if err := validateFrequency(f1, 1); err != nil {
		return 0, err
	}
	if err := validateFrequency(f2, 2); err != nil {
		return 0, err
	}
	difference := new(big.Rat).Sub(binaryRat(f2), binaryRat(f1))
	difference.Abs(difference)
	result, _ := difference.Float64()
	return result, nil
}

// MustBeatFrequency calls BeatFrequency and panics with its error, if any.
func MustBeatFrequency(f1, f2 float64) float64 {
	result, err := BeatFrequency(f1, f2)
	if err != nil {
		panic(err)
	}
	return result
}

// Period returns the oscillation period 1/f in seconds. The exact reciprocal
// of the binary input is rounded directly to float64, without intermediate
// 53-bit rounding; a positive subnormal result is valid. Milliseconds are only
// a display unit derived from the returned seconds. This is neither a musical
// duration nor the duration of the whole sound. It returns ErrInvalidFrequency
// for invalid f and ErrOutOfRange, without a cents error, if the result
// overflows to positive infinity.
func Period(f float64) (float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return 0, err
	}
	period := new(big.Rat).Inv(binaryRat(f))
	result, _ := period.Float64()
	if math.IsInf(result, 1) {
		return 0, ErrOutOfRange
	}
	return result, nil
}

// MustPeriod calls Period and panics with its error, if any.
func MustPeriod(f float64) float64 {
	result, err := Period(f)
	if err != nil {
		panic(err)
	}
	return result
}
