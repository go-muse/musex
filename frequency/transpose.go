package frequency

import (
	"github.com/go-muse/musex/cents"
)

// Transpose returns f transposed by k steps of the equal division of the
// octave into n parts: f·2^(k/n). Seven steps of 12-EDO are seven equal
// semitones, not the just ratio 3:2. Validation order is f, then n < 1, then
// the generated result. Errors are ErrInvalidFrequency, ErrInvalidParameter,
// and ErrOutOfRange respectively; the latter two preserve the corresponding
// cents argument or range error.
func Transpose(f float64, k, n int) (float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, invalidTransposeParameter()
	}
	step, err := cents.Step(k, n)
	if err != nil {
		return 0, resultOutOfRange("transpose step")
	}
	result, err := step.ApplyTo(f)
	if err != nil {
		return 0, resultOutOfRange("transpose result")
	}
	return result, nil
}

// MustTranspose calls Transpose and panics with its error, if any.
func MustTranspose(f float64, k, n int) float64 {
	result, err := Transpose(f, k, n)
	if err != nil {
		panic(err)
	}
	return result
}

// CentsBetween returns the signed distance from f1 to f2 using cents.Between:
// 1200·log2(f2/f1) on the 10^-6-cent grid, with both frequencies interpreted
// as their shortest decimal representations. It validates f1 and then f2 and
// has no other failure for valid frequencies.
func CentsBetween(f1, f2 float64) (cents.Cents, error) {
	if err := validateFrequency(f1, 1); err != nil {
		return cents.Cents{}, err
	}
	if err := validateFrequency(f2, 2); err != nil {
		return cents.Cents{}, err
	}
	distance, err := cents.Between(f1, f2)
	if err != nil {
		return cents.Cents{}, invalidFrequency(1)
	}
	return distance, nil
}

// MustCentsBetween calls CentsBetween and panics with its error, if any.
func MustCentsBetween(f1, f2 float64) cents.Cents {
	distance, err := CentsBetween(f1, f2)
	if err != nil {
		panic(err)
	}
	return distance
}
