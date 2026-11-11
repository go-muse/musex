package rounding

import (
	"math/big"
	"strings"
)

// QuoRemBig rounds the exact quotient n/d to q and returns r = n-q*d. Nil n or
// d returns ErrUnset and a zero divisor returns ErrInvalidArgument. Both
// results are newly allocated and the inputs are not changed or retained.
func (m Mode) QuoRemBig(n, d *big.Int) (q, r *big.Int, err error) {
	if m == 0 || n == nil || d == nil {
		return nil, nil, ErrUnset
	}
	if err := validateMode(m); err != nil {
		return nil, nil, err
	}
	if d.Sign() == 0 {
		return nil, nil, ErrInvalidArgument
	}

	q = roundRat(m, new(big.Rat).SetFrac(n, d))
	r = new(big.Int).Sub(n, new(big.Int).Mul(q, d))
	return q, r, nil
}

// MustQuoRemBig calls QuoRemBig and panics with its error on failure.
func (m Mode) MustQuoRemBig(n, d *big.Int) (q, r *big.Int) {
	q, r, err := m.QuoRemBig(n, d)
	if err != nil {
		panic(err)
	}
	return q, r
}

// RatInt rounds x to an integer. Nil x returns ErrUnset. The result is newly
// allocated and x is not changed or retained.
func (m Mode) RatInt(x *big.Rat) (*big.Int, error) {
	if m == 0 || x == nil {
		return nil, ErrUnset
	}
	if err := validateMode(m); err != nil {
		return nil, err
	}
	return roundRat(m, x), nil
}

// MustRatInt calls RatInt and panics with its error on failure.
func (m Mode) MustRatInt(x *big.Rat) *big.Int {
	k, err := m.RatInt(x)
	if err != nil {
		panic(err)
	}
	return k
}

// RatMultiple rounds x to a multiple of the positive rational step. It returns
// the selected node index k and its value k*step. Nil x or step returns
// ErrUnset and step <= 0 returns ErrInvalidArgument. Both results are newly
// allocated and do not share storage with the inputs or each other.
func (m Mode) RatMultiple(x, step *big.Rat) (k *big.Int, v *big.Rat, err error) {
	if m == 0 || x == nil || step == nil {
		return nil, nil, ErrUnset
	}
	if err := validateMode(m); err != nil {
		return nil, nil, err
	}
	if step.Sign() <= 0 {
		return nil, nil, ErrInvalidArgument
	}

	t := new(big.Rat).Quo(x, step)
	k = roundRat(m, t)
	numerator := new(big.Int).Mul(k, step.Num())
	v = new(big.Rat).SetFrac(numerator, step.Denom())
	return k, v, nil
}

// MustRatMultiple calls RatMultiple and panics with its error on failure.
func (m Mode) MustRatMultiple(x, step *big.Rat) (k *big.Int, v *big.Rat) {
	k, v, err := m.RatMultiple(x, step)
	if err != nil {
		panic(err)
	}
	return k, v
}

// RatDecimal rounds x to places decimal digits and returns its plain decimal
// representation. Negative places round to tens, hundreds, and so on. Places
// outside [-MaxPlaces, MaxPlaces] return ErrInvalidArgument; nil x returns
// ErrUnset. Zero never has a minus sign.
func (m Mode) RatDecimal(x *big.Rat, places int) (string, error) {
	if m == 0 || x == nil {
		return "", ErrUnset
	}
	if err := validateMode(m); err != nil {
		return "", err
	}
	if places < -MaxPlaces || places > MaxPlaces {
		return "", ErrInvalidArgument
	}
	k := roundRat(m, scaleDecimal(x, places))
	return formatDecimal(k, places), nil
}

// MustRatDecimal calls RatDecimal and panics with its error on failure.
func (m Mode) MustRatDecimal(x *big.Rat, places int) string {
	s, err := m.RatDecimal(x, places)
	if err != nil {
		panic(err)
	}
	return s
}

func roundRat(m Mode, x *big.Rat) *big.Int {
	lower, remainder := new(big.Int), new(big.Int)
	lower.DivMod(x.Num(), x.Denom(), remainder)
	if remainder.Sign() == 0 {
		return lower
	}

	twiceRemainder := new(big.Int).Lsh(new(big.Int).Set(remainder), 1)
	cmp := twiceRemainder.Cmp(x.Denom())
	side := Positive
	if lower.Sign() < 0 {
		side = Negative
	}
	if m.choose(cmp, side, lower.Bit(0) == 0) == Upper {
		lower.Add(lower, big.NewInt(1))
	}
	return lower
}

func scaleDecimal(x *big.Rat, places int) *big.Rat {
	if places >= 0 {
		return new(big.Rat).Mul(x, new(big.Rat).SetInt(pow10(places)))
	}
	return new(big.Rat).Quo(x, new(big.Rat).SetInt(pow10(-places)))
}

func pow10(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}

func formatDecimal(k *big.Int, places int) string {
	negative := k.Sign() < 0
	absolute := new(big.Int).Abs(k)
	var digits string
	if places > 0 {
		scale := pow10(places)
		whole, fraction := new(big.Int), new(big.Int)
		whole.QuoRem(absolute, scale, fraction)
		fractionDigits := fraction.String()
		digits = whole.String() + "." + strings.Repeat("0", places-len(fractionDigits)) + fractionDigits
	} else {
		digits = absolute.Mul(absolute, pow10(-places)).String()
	}
	if negative {
		return "-" + digits
	}
	return digits
}
