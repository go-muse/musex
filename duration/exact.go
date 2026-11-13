package duration

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
)

// ratOf returns a new *big.Rat equal to the valid fraction f.
func ratOf(f fraction.Fraction) *big.Rat {
	r, ok := f.BigRat()
	if !ok {
		return new(big.Rat)
	}
	return r
}

// fromRat returns r as a reduced fraction and reports whether r is within the
// range of package common/fraction.
func fromRat(r *big.Rat) (fraction.Fraction, bool) {
	f, err := fraction.FromBigRat(r)
	if err != nil {
		return fraction.Fraction{}, false
	}
	return f, true
}

// pow2Rat returns 2^e as a new *big.Rat for any integer e that fits the
// allocation of the result.
func pow2Rat(e int) *big.Rat {
	if e >= 0 {
		return new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(e)))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), uint(-e)))
}

// isPow2 reports whether x is a positive power of two, including 1.
func isPow2(x *big.Int) bool {
	return x.Sign() > 0 && x.BitLen()-1 == int(x.TrailingZeroBits())
}

// floorRat returns ⌊r⌋ for r ≥ 0.
func floorRat(r *big.Rat) *big.Int {
	return new(big.Int).Quo(r.Num(), r.Denom())
}

// dotFactor returns the multiplier of k dots, 2 − 1/2^k = (2^(k+1) − 1)/2^k.
func dotFactor(k int) *big.Rat {
	one := big.NewInt(1)
	den := new(big.Int).Lsh(one, uint(k))
	num := new(big.Int).Lsh(one, uint(k+1))
	num.Sub(num, one)
	return new(big.Rat).SetFrac(num, den)
}
