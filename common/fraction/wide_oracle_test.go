package fraction_test

import (
	"errors"
	"math/big"
	"slices"
	"strings"

	"github.com/go-muse/musex/common/fraction"
)

// This file holds the contract oracle for the wide-value tests. Every expected
// result is computed from the formulas of the package contract (sections 3–9)
// on math/big values: exact values and reduced pairs through big.Rat, the
// original pairs of non-reducing modes through big.Int. The oracle never calls
// the operation it predicts; it reads operands only through BigPair.

// wideWant is the expected outcome of an operation that returns a Fraction:
// either the exact original pair of the result or the sentinel error.
type wideWant struct {
	num, den *big.Int
	err      error
}

// wideArith names the four arithmetic operations of the binary family.
type wideArith uint8

const (
	wideAdd wideArith = iota
	wideSub
	wideMul
	wideDiv
)

func (op wideArith) String() string {
	return [...]string{"Add", "Sub", "Mul", "Div"}[op]
}

// wideModes lists every DenomMode of the contract plus one unknown value.
var wideModes = []fraction.DenomMode{
	fraction.DenomReduced,
	fraction.DenomUnreducedLCM,
	fraction.DenomUnreducedProduct,
	fraction.DenomSameUnit,
	fraction.DenomLargerUnit,
	fraction.DenomSmallerUnit,
	fraction.DenomSmallerUnit + 1,
}

// wideKnownRounding lists the five named rounding modes. The tests identify
// modes by these names only, never by their numeric values or order, so they
// hold for any numbering of RoundingMode.
var wideKnownRounding = []fraction.RoundingMode{
	fraction.RoundHalfAwayFromZero,
	fraction.RoundHalfToEven,
	fraction.RoundFloor,
	fraction.RoundCeil,
	fraction.RoundTrunc,
}

// wideUnknownRounding is a value outside every set of named rounding modes.
const wideUnknownRounding = fraction.RoundingMode(99)

// wideRoundingModes lists the named rounding modes plus one unknown value.
var wideRoundingModes = append(slices.Clone(wideKnownRounding), wideUnknownRounding)

// wKnownRounding reports whether mode is one of the five named modes.
func wKnownRounding(mode fraction.RoundingMode) bool {
	return slices.Contains(wideKnownRounding, mode)
}

func wPow2(e uint) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), e)
}

// wOff returns x + d as a new value.
func wOff(x *big.Int, d int64) *big.Int {
	return new(big.Int).Add(x, big.NewInt(d))
}

// wMul returns the product of the factors as a new value.
func wMul(factors ...*big.Int) *big.Int {
	p := big.NewInt(1)
	for _, f := range factors {
		p.Mul(p, f)
	}
	return p
}

func wNeg(x *big.Int) *big.Int {
	return new(big.Int).Neg(x)
}

func wGCD(a, b *big.Int) *big.Int {
	return new(big.Int).GCD(nil, nil, new(big.Int).Abs(a), new(big.Int).Abs(b))
}

// wHex parses a hexadecimal constant used in boundary tables.
func wHex(s string) *big.Int {
	v, ok := new(big.Int).SetString(strings.ReplaceAll(s, "_", ""), 16)
	if !ok {
		panic("invalid hexadecimal constant " + s)
	}
	return v
}

// wDec parses a decimal constant used in boundary tables.
func wDec(s string) *big.Int {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid decimal constant " + s)
	}
	return v
}

func wInRange(x *big.Int) bool {
	return x.CmpAbs(fraction.MaxComponentBig()) <= 0
}

// wFitsInt reports whether x is a 64-bit int value (the project targets 64-bit platforms only).
func wFitsInt(x *big.Int) bool {
	return x.IsInt64()
}

// wFrac builds a fraction with the original pair (num, den); it panics on invalid
// test data, which is a defect of the table, not of the package under test.
func wFrac(num, den *big.Int) fraction.Fraction {
	return fraction.MustFromBigInt(num, den)
}

// wNorm returns (sign(d)·n, |d|) of a valid fraction as fresh values.
func wNorm(f fraction.Fraction) (n, d *big.Int) {
	n, d, ok := f.BigPair()
	if !ok {
		panic("wNorm: invalid fraction")
	}
	if d.Sign() < 0 {
		n.Neg(n)
		d.Neg(d)
	}
	return n, d
}

// wRat returns the exact value of a valid fraction.
func wRat(f fraction.Fraction) *big.Rat {
	n, d := wNorm(f)
	return new(big.Rat).SetFrac(n, d)
}

func wErr(err error) wideWant {
	return wideWant{err: err}
}

// wRange returns the pair (num, den) as is, or ErrOutOfRange when a component
// exceeds MaxComponent (contract, section 9, step 4).
func wRange(num, den *big.Int) wideWant {
	if !wInRange(num) || !wInRange(den) {
		return wErr(fraction.ErrOutOfRange)
	}
	return wideWant{num: new(big.Int).Set(num), den: new(big.Int).Set(den)}
}

// wFromRat returns the reduced canonical pair of r with the range check.
func wFromRat(r *big.Rat) wideWant {
	return wRange(r.Num(), r.Denom())
}

// wInTarget expresses the exact value r in the positive denominator target:
// ErrNotRepresentable when target mod Q != 0, otherwise (P·(target/Q), target)
// with the range check (contract, sections 3 and 6, modes 4–7).
func wInTarget(r *big.Rat, target *big.Int) wideWant {
	q := r.Denom()
	k, m := new(big.Int).QuoRem(target, q, new(big.Int))
	if m.Sign() != 0 {
		return wErr(fraction.ErrNotRepresentable)
	}
	return wRange(new(big.Int).Mul(r.Num(), k), target)
}

// wValue returns the exact value of x op y; the second result is false for division by zero.
func wValue(op wideArith, x, y fraction.Fraction) (*big.Rat, bool) {
	rx, ry := wRat(x), wRat(y)
	switch op {
	case wideAdd:
		return new(big.Rat).Add(rx, ry), true
	case wideSub:
		return new(big.Rat).Sub(rx, ry), true
	case wideMul:
		return new(big.Rat).Mul(rx, ry), true
	default:
		if ry.Sign() == 0 {
			return nil, false
		}
		return new(big.Rat).Quo(rx, ry), true
	}
}

// wantArith predicts AddWith/SubWith/MulWith/DivWith in the given mode
// (contract, section 6, modes 1–6, and the unknown-mode note).
func wantArith(op wideArith, x, y fraction.Fraction, mode fraction.DenomMode) wideWant {
	n1, d1 := wNorm(x)
	n2, d2 := wNorm(y)
	if op == wideDiv && n2.Sign() == 0 {
		return wErr(fraction.ErrDivisionByZero)
	}
	if mode > fraction.DenomSmallerUnit {
		return wErr(fraction.ErrNotRepresentable)
	}
	value, _ := wValue(op, x, y)

	mulPair := func() wideWant {
		return wRange(wMul(n1, n2), wMul(d1, d2))
	}
	divPair := func() wideWant {
		num := wMul(n1, d2)
		if n2.Sign() < 0 {
			num.Neg(num)
		}
		return wRange(num, wMul(d1, new(big.Int).Abs(n2)))
	}
	sum := func(a, b *big.Int) *big.Int {
		if op == wideSub {
			return new(big.Int).Sub(a, b)
		}
		return new(big.Int).Add(a, b)
	}

	switch mode {
	case fraction.DenomReduced:
		return wFromRat(value)
	case fraction.DenomUnreducedLCM:
		switch op {
		case wideAdd, wideSub:
			g := wGCD(d1, d2)
			k1 := new(big.Int).Quo(d2, g)
			k2 := new(big.Int).Quo(d1, g)
			return wRange(sum(wMul(n1, k1), wMul(n2, k2)), wMul(k2, d2))
		case wideMul:
			return mulPair()
		default:
			return divPair()
		}
	case fraction.DenomUnreducedProduct:
		switch op {
		case wideAdd, wideSub:
			return wRange(sum(wMul(n1, d2), wMul(n2, d1)), wMul(d1, d2))
		case wideMul:
			return mulPair()
		default:
			return divPair()
		}
	case fraction.DenomSameUnit:
		if d1.Cmp(d2) != 0 {
			return wErr(fraction.ErrNotRepresentable)
		}
		if op == wideAdd || op == wideSub {
			return wRange(sum(n1, n2), d1)
		}
		return wInTarget(value, d1)
	case fraction.DenomLargerUnit:
		if d1.Cmp(d2) <= 0 {
			return wInTarget(value, d1)
		}
		return wInTarget(value, d2)
	default: // fraction.DenomSmallerUnit
		if d1.Cmp(d2) >= 0 {
			return wInTarget(value, d1)
		}
		return wInTarget(value, d2)
	}
}

// wantInDenominator predicts the mode-7 operations (contract, section 6, mode 7).
// A zero divisor is checked before the target denominator (section 9, step 2).
func wantInDenominator(op wideArith, x, y fraction.Fraction, den int) wideWant {
	value, ok := wValue(op, x, y)
	switch {
	case !ok:
		return wErr(fraction.ErrDivisionByZero)
	case den <= 0:
		return wErr(fraction.ErrZeroDenominator)
	}
	return wInTarget(value, big.NewInt(int64(den)))
}

// wideDivResult is the expected outcome of QuoRem, DivModFloor, Split and their wide forms.
type wideDivResult struct {
	q   *big.Int // nil on error
	rem wideWant // rem.err carries the error of the whole call
}

// wantQuoRem predicts QuoRem/QuoRemBig (floor == false) and DivModFloor/DivModFloorBig
// (floor == true); narrow selects the int forms (contract, section 7).
func wantQuoRem(x, step fraction.Fraction, floor, narrow bool) wideDivResult {
	rs := wRat(step)
	if rs.Sign() == 0 {
		return wideDivResult{rem: wErr(fraction.ErrDivisionByZero)}
	}
	if floor && rs.Sign() < 0 {
		return wideDivResult{rem: wErr(fraction.ErrZeroDenominator)}
	}
	rx := wRat(x)
	ratio := new(big.Rat).Quo(rx, rs)
	q := new(big.Int)
	if floor {
		q.Div(ratio.Num(), ratio.Denom()) // Euclidean division by a positive divisor is the floor.
	} else {
		q.Quo(ratio.Num(), ratio.Denom())
	}
	if narrow && !wFitsInt(q) {
		return wideDivResult{rem: wErr(fraction.ErrNotRepresentable)}
	}
	rem := new(big.Rat).Sub(rx, new(big.Rat).Mul(new(big.Rat).SetInt(q), rs))
	want := wFromRat(rem)
	if want.err != nil {
		return wideDivResult{rem: want}
	}
	return wideDivResult{q: q, rem: want}
}

// wantSplit predicts Split/SplitBig (floor == false) and SplitFloor/SplitFloorBig
// (floor == true); the remainder keeps |d| without reduction (contract, section 7).
func wantSplit(x fraction.Fraction, floor, narrow bool) wideDivResult {
	n, d := wNorm(x)
	q, r := new(big.Int), new(big.Int)
	if floor {
		q.DivMod(n, d, r)
	} else {
		q.QuoRem(n, d, r)
	}
	if narrow && !wFitsInt(q) {
		return wideDivResult{rem: wErr(fraction.ErrNotRepresentable)}
	}
	return wideDivResult{q: q, rem: wideWant{num: r, den: d}}
}

// wantIsMultipleOf predicts IsMultipleOf.
func wantIsMultipleOf(x, step fraction.Fraction) (bool, error) {
	rs := wRat(step)
	if rs.Sign() == 0 {
		return false, fraction.ErrDivisionByZero
	}
	return new(big.Rat).Quo(wRat(x), rs).IsInt(), nil
}

// wantInProgression predicts InProgression(start, step) on x.
func wantInProgression(x, start, step fraction.Fraction) (int, bool, error) {
	rs := wRat(step)
	switch {
	case rs.Sign() == 0:
		return 0, false, fraction.ErrDivisionByZero
	case rs.Sign() < 0:
		return 0, false, fraction.ErrZeroDenominator
	}
	k := new(big.Rat).Quo(new(big.Rat).Sub(wRat(x), wRat(start)), rs)
	if !k.IsInt() {
		return 0, false, nil
	}
	if !wFitsInt(k.Num()) {
		return 0, false, fraction.ErrNotRepresentable
	}
	return int(k.Num().Int64()), true, nil
}

// wantScaleIntExact predicts ScaleIntExact(c).
func wantScaleIntExact(x fraction.Fraction, c int) (int, bool, error) {
	r := wRat(x)
	k, m := new(big.Int).QuoRem(big.NewInt(int64(c)), r.Denom(), new(big.Int))
	if m.Sign() != 0 {
		return 0, false, nil
	}
	prod := new(big.Int).Mul(r.Num(), k)
	if !wFitsInt(prod) {
		return 0, false, fraction.ErrNotRepresentable
	}
	return int(prod.Int64()), true, nil
}

// wRound rounds x/y (y > 0) to an integer by the rule of the contract (section 7).
func wRound(x, y *big.Int, mode fraction.RoundingMode) *big.Int {
	q, r := new(big.Int).QuoRem(x, y, new(big.Int))
	r.Abs(r)
	twice := new(big.Int).Lsh(r, 1)
	step := big.NewInt(int64(x.Sign()))
	switch mode {
	case fraction.RoundFloor:
		if x.Sign() < 0 && r.Sign() > 0 {
			q.Sub(q, big.NewInt(1))
		}
	case fraction.RoundCeil:
		if x.Sign() > 0 && r.Sign() > 0 {
			q.Add(q, big.NewInt(1))
		}
	case fraction.RoundHalfAwayFromZero:
		if twice.Cmp(y) >= 0 {
			q.Add(q, step)
		}
	case fraction.RoundHalfToEven:
		if c := twice.Cmp(y); c > 0 || c == 0 && q.Bit(0) == 1 {
			q.Add(q, step)
		}
	}
	return q
}

// wantRoundToBigInt predicts RoundToBigInt; RoundToInt adds the int check.
func wantRoundToBigInt(x fraction.Fraction, mode fraction.RoundingMode) (*big.Int, error) {
	if !wKnownRounding(mode) {
		return nil, fraction.ErrNotRepresentable
	}
	n, d := wNorm(x)
	return wRound(n, d, mode), nil
}

// wantRoundToBigDenominator predicts RoundToBigDenominator; RoundToDenominator
// is the same rule for an int denominator.
func wantRoundToBigDenominator(x fraction.Fraction, den *big.Int, mode fraction.RoundingMode) wideWant {
	if den == nil || den.Sign() <= 0 {
		return wErr(fraction.ErrZeroDenominator)
	}
	if !wKnownRounding(mode) {
		return wErr(fraction.ErrNotRepresentable)
	}
	if !wInRange(den) {
		return wErr(fraction.ErrOutOfRange)
	}
	n, d := wNorm(x)
	return wRange(wRound(wMul(n, den), d, mode), den)
}

// wantWithBigDenominator predicts WithBigDenominator; WithDenominator is the
// same rule for an int denominator.
func wantWithBigDenominator(x fraction.Fraction, den *big.Int) wideWant {
	if den == nil || den.Sign() <= 0 {
		return wErr(fraction.ErrZeroDenominator)
	}
	return wInTarget(wRat(x), den)
}

// wantPow predicts Pow(exp) (contract, section 5).
func wantPow(x fraction.Fraction, exp int) wideWant {
	if exp == 0 {
		return wideWant{num: big.NewInt(1), den: big.NewInt(1)}
	}
	r := wRat(x)
	if r.Sign() == 0 {
		if exp < 0 {
			return wErr(fraction.ErrDivisionByZero)
		}
		return wideWant{num: big.NewInt(0), den: big.NewInt(1)}
	}
	base, other := new(big.Int).Set(r.Num()), new(big.Int).Set(r.Denom())
	var k uint64
	if exp < 0 {
		base, other = other, base
		if other.Sign() < 0 {
			other.Neg(other)
			base.Neg(base)
		}
		k = uint64(-(exp + 1)) + 1
	} else {
		k = uint64(exp)
	}
	pow := func(v *big.Int) (*big.Int, bool) {
		if v.CmpAbs(big.NewInt(1)) == 0 {
			if v.Sign() < 0 && k%2 == 1 {
				return big.NewInt(-1), true
			}
			return big.NewInt(1), true
		}
		// |v| >= 2: |v|^k >= 2^((bitlen-1)·k) exceeds MaxComponent once that exponent reaches 255.
		if k >= 255 || uint64(v.BitLen()-1)*k >= 255 {
			return nil, false
		}
		p := new(big.Int).Exp(v, new(big.Int).SetUint64(k), nil)
		return p, wInRange(p)
	}
	num, ok1 := pow(base)
	den, ok2 := pow(other)
	if !ok1 || !ok2 {
		return wErr(fraction.ErrOutOfRange)
	}
	return wideWant{num: num, den: den}
}

// wantFromBigIntReduced predicts FromBigIntReduced for a non-nil num and a nonzero den.
func wantFromBigIntReduced(num, den *big.Int) wideWant {
	return wFromRat(new(big.Rat).SetFrac(num, den))
}

// wantMixedString predicts MixedString of a valid fraction (contract, section 8).
func wantMixedString(x fraction.Fraction) string {
	r := wRat(x)
	n, d := r.Num(), r.Denom()
	if d.IsInt64() && d.Int64() == 1 {
		return n.String()
	}
	if n.CmpAbs(d) < 0 {
		return n.String() + "/" + d.String()
	}
	q, rem := new(big.Int).QuoRem(n, d, new(big.Int))
	return q.String() + " " + rem.Abs(rem).String() + "/" + d.String()
}

// wantCompactString predicts CompactString of a valid fraction.
func wantCompactString(x fraction.Fraction) string {
	r := wRat(x)
	if r.IsInt() {
		return r.Num().String()
	}
	return r.Num().String() + "/" + r.Denom().String()
}

// wantFormatDecimal predicts FormatDecimal(digits, mode) (contract, section 8).
func wantFormatDecimal(x fraction.Fraction, digits int, mode fraction.RoundingMode) (string, error) {
	if digits < 0 || digits > fraction.MaxDigits || !wKnownRounding(mode) {
		return "", fraction.ErrNotRepresentable
	}
	n, d := wNorm(x)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	m := wRound(wMul(n, scale), d, mode)
	if digits == 0 {
		return m.String(), nil
	}
	sign := ""
	if m.Sign() < 0 {
		sign = "-"
	}
	s := new(big.Int).Abs(m).String()
	if len(s) <= digits {
		s = strings.Repeat("0", digits+1-len(s)) + s
	}
	return sign + s[:len(s)-digits] + "." + s[len(s)-digits:], nil
}

// wSentinelOnly reports whether err matches want and none of the other sentinels.
func wSentinelOnly(err, want error) bool {
	if !errors.Is(err, want) {
		return false
	}
	for _, s := range allSentinels() {
		if s != want && errors.Is(err, s) {
			return false
		}
	}
	return true
}

// wCatch runs fn and returns the value it panicked with, or nil.
func wCatch(fn func()) (recovered any) {
	defer func() {
		recovered = recover()
	}()
	fn()
	return nil
}
