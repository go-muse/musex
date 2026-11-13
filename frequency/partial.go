package frequency

import (
	"math/big"

	"github.com/go-muse/musex/cents"
)

// Partial returns the n-th partial tone f·n. Index 1 is the fundamental and
// index 2 is the first overtone. It validates f, then n >= 1, then requires
// the rounded generated frequency to be normal. Errors are
// ErrInvalidFrequency, ErrInvalidParameter, and ErrOutOfRange respectively.
func Partial(f float64, n int) (float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, invalidParameter("partial index")
	}
	return generatedFrequency(new(big.Rat).Mul(binaryRat(f), big.NewRat(int64(n), 1)), "partial")
}

// MustPartial calls Partial and panics with its error, if any.
func MustPartial(f float64, n int) float64 {
	result, err := Partial(f, n)
	if err != nil {
		panic(err)
	}
	return result
}

// Partials returns f·n for every index in the same order. It validates f, the
// nonempty list, and every index before computing any result. On any error it
// returns nil rather than a partial list. On success it returns a new slice
// whose length and capacity equal len(indices), owned by the caller. The
// input slice is only read during the call and is neither changed nor kept.
func Partials(f float64, indices []int) ([]float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return nil, err
	}
	if err := validateIndices(indices, "partial"); err != nil {
		return nil, err
	}
	results := make([]float64, len(indices))
	for i, n := range indices {
		result, err := generatedFrequency(new(big.Rat).Mul(binaryRat(f), big.NewRat(int64(n), 1)), "partials")
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// MustPartials calls Partials and panics with its error, if any. Its result
// has the same independent ownership as the result of Partials.
func MustPartials(f float64, indices []int) []float64 {
	results, err := Partials(f, indices)
	if err != nil {
		panic(err)
	}
	return results
}

// Subharmonic returns f/n, where index 1 is the fundamental and indices n >= 2
// are lower tones. This ratio does not assert that every subharmonic is
// physically present in a sound; the definition and numbering are supported
// only by two weak domain sources. Errors follow the same order as Partial.
func Subharmonic(f float64, n int) (float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, invalidParameter("subharmonic index")
	}
	return generatedFrequency(new(big.Rat).Quo(binaryRat(f), big.NewRat(int64(n), 1)), "subharmonic")
}

// MustSubharmonic calls Subharmonic and panics with its error, if any.
func MustSubharmonic(f float64, n int) float64 {
	result, err := Subharmonic(f, n)
	if err != nil {
		panic(err)
	}
	return result
}

// Subharmonics returns f/n for every index in the same order. Validation,
// atomic error behavior, and result ownership are the same as for Partials.
// A subnormal generated frequency is outside the normal-result range even
// though a subnormal input frequency is valid.
func Subharmonics(f float64, indices []int) ([]float64, error) {
	if err := validateFrequency(f, 1); err != nil {
		return nil, err
	}
	if err := validateIndices(indices, "subharmonic"); err != nil {
		return nil, err
	}
	results := make([]float64, len(indices))
	for i, n := range indices {
		result, err := generatedFrequency(new(big.Rat).Quo(binaryRat(f), big.NewRat(int64(n), 1)), "subharmonics")
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// MustSubharmonics calls Subharmonics and panics with its error, if any. Its
// result has the same independent ownership as the result of Subharmonics.
func MustSubharmonics(f float64, indices []int) []float64 {
	results, err := Subharmonics(f, indices)
	if err != nil {
		panic(err)
	}
	return results
}

func validateIndices(indices []int, operation string) error {
	if len(indices) == 0 {
		return invalidParameter(operation + " indices are empty")
	}
	for _, n := range indices {
		if n < 1 {
			return invalidParameter(operation + " index")
		}
	}
	return nil
}

func generatedFrequency(exact *big.Rat, operation string) (float64, error) {
	rounded := new(big.Float).SetPrec(53).SetMode(big.ToNearestEven).SetRat(exact)
	minimum := new(big.Float).SetPrec(53).SetFloat64(cents.MinNormalFrequency)
	maximum := new(big.Float).SetPrec(53).SetFloat64(cents.MaxFrequency)
	if rounded.Cmp(minimum) < 0 || rounded.Cmp(maximum) > 0 {
		return 0, resultOutOfRange(operation)
	}
	result, _ := rounded.Float64()
	if !cents.IsNormalFrequency(result) {
		return 0, resultOutOfRange(operation)
	}
	return result, nil
}
