package fraction

import (
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

func formatSignedU256(u u256, neg bool) string {
	if u.isZero() {
		return "0"
	}
	if u.isUint64() {
		if neg {
			return "-" + strconv.FormatUint(u[0], 10)
		}
		return strconv.FormatUint(u[0], 10)
	}
	return u.toBigInt(neg).String()
}

// String returns the stored pair formatted as "<n>/<d>" (or "0/0" if f is
// invalid).
func (f Fraction) String() string {
	if !f.IsValid() {
		return "0/0"
	}
	return formatSignedU256(f.num, f.numNeg()) + "/" + formatSignedU256(f.den, f.denNeg())
}

// ReducedString returns the reduced canonical pair formatted as
// "<n_red>/<d_red>" (or "0/0" if f is invalid).
func (f Fraction) ReducedString() string {
	if !f.IsValid() {
		return "0/0"
	}
	return f.Reduce().String()
}

// CompactString returns "<n_red>" if f is a valid integer fraction,
// "<n_red>/<d_red>" if f is a valid non-integer fraction, or "0/0" if f is
// invalid.
func (f Fraction) CompactString() string {
	if !f.IsValid() {
		return "0/0"
	}
	red := f.Reduce()
	if red.den.isOne() {
		return formatSignedU256(red.num, red.numNeg())
	}
	return red.String()
}

// MixedString formats a valid improper non-integer fraction as "<q> <rem>/<d>"
// using its reduced form, a valid integer as "<n_red>", a valid proper
// fraction as "<n_red>/<d_red>", and an invalid fraction as "0/0".
func (f Fraction) MixedString() string {
	if !f.IsValid() {
		return "0/0"
	}
	red := f.Reduce()
	if red.den.isOne() {
		return formatSignedU256(red.num, red.numNeg())
	}
	if u256Cmp(red.num, red.den) < 0 {
		return red.String()
	}
	qMag, remMag := u256DivMod(red.num, red.den)
	return formatSignedU256(qMag, red.numNeg()) + " " + formatSignedU256(remMag, false) + "/" + formatSignedU256(red.den, false)
}

// Decimal formats f as a fixed-point decimal string with digits fractional
// digits using [RoundHalfAwayFromZero].
func (f Fraction) Decimal(digits int) (string, error) {
	return f.FormatDecimal(digits, RoundHalfAwayFromZero)
}

// MustDecimal is like [Fraction.Decimal] but panics on error.
func (f Fraction) MustDecimal(digits int) string {
	s, err := f.Decimal(digits)
	if err != nil {
		panic(err)
	}
	return s
}

// FormatDecimal formats f as a fixed-point decimal string with digits
// fractional digits rounded according to mode. It returns [ErrInvalid] if f is
// invalid and [ErrNotRepresentable] if digits < 0, digits > [MaxDigits], or
// mode is unknown.
func (f Fraction) FormatDecimal(digits int, mode RoundingMode) (string, error) {
	if !f.IsValid() {
		return "", ErrInvalid
	}
	if digits < 0 || digits > MaxDigits || mode > RoundTrunc {
		return "", ErrNotRepresentable
	}
	pow10 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	scaledNum := new(big.Int).Mul(f.num.toBigInt(false), pow10)
	denBig := f.den.toBigInt(false)
	q, rem := new(big.Int).QuoRem(scaledNum, denBig, new(big.Int))
	neg := f.Sign() < 0
	if rem.Sign() != 0 {
		var inc bool
		switch mode {
		case RoundTrunc:
			inc = false
		case RoundFloor:
			inc = neg
		case RoundCeil:
			inc = !neg
		case RoundHalfAwayFromZero:
			rem2 := new(big.Int).Lsh(rem, 1)
			inc = rem2.Cmp(denBig) >= 0
		case RoundHalfToEven:
			rem2 := new(big.Int).Lsh(rem, 1)
			c := rem2.Cmp(denBig)
			inc = c > 0 || (c == 0 && q.Bit(0) == 1)
		}
		if inc {
			q.Add(q, big.NewInt(1))
		}
	}
	if q.Sign() == 0 {
		neg = false
	}
	signPrefix := ""
	if neg {
		signPrefix = "-"
	}
	if digits == 0 {
		return signPrefix + q.String(), nil
	}
	intPart, fracPart := new(big.Int).QuoRem(q, pow10, new(big.Int))
	fracStr := fracPart.String()
	if len(fracStr) < digits {
		fracStr = strings.Repeat("0", digits-len(fracStr)) + fracStr
	}
	return signPrefix + intPart.String() + "." + fracStr, nil
}

// MustFormatDecimal is like [Fraction.FormatDecimal] but panics on error.
func (f Fraction) MustFormatDecimal(digits int, mode RoundingMode) string {
	s, err := f.FormatDecimal(digits, mode)
	if err != nil {
		panic(err)
	}
	return s
}

// MarshalText implements [encoding.TextMarshaler].
func (f Fraction) MarshalText() ([]byte, error) {
	if !f.IsValid() {
		return nil, ErrInvalid
	}
	return []byte(f.String()), nil
}

// AppendText implements [encoding.TextAppender].
func (f Fraction) AppendText(b []byte) ([]byte, error) {
	if !f.IsValid() {
		return b, ErrInvalid
	}
	return append(b, f.String()...), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (f *Fraction) UnmarshalText(text []byte) error {
	if f == nil {
		return ErrInvalid
	}
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*f = parsed
	return nil
}

func splitSign(s string) (neg bool, rest string) {
	if len(s) > 0 {
		switch s[0] {
		case '-':
			return true, s[1:]
		case '+':
			return false, s[1:]
		}
	}
	return false, s
}

func isASCIIAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isAllZeroDigits(s string) bool {
	for i := range len(s) {
		if s[i] != '0' {
			return false
		}
	}
	return true
}

func parseDigitsToU256(digits string) (u256, bool) {
	if len(digits) <= 19 {
		var v uint64
		for i := range len(digits) {
			v = v*10 + uint64(digits[i]-'0')
		}
		return u256{v, 0, 0, 0}, true
	}
	var acc u256
	for i := range len(digits) {
		d := uint64(digits[i] - '0')
		carry := d
		for j := range 4 {
			hi, lo := bits.Mul64(acc[j], 10)
			var c uint64
			acc[j], c = bits.Add64(lo, carry, 0)
			carry = hi + c
		}
		if carry != 0 || acc.exceedsMax() {
			return u256Zero, false
		}
	}
	return acc, true
}

// Parse parses a Fraction from text in ordinary ("a/b"), integer ("n"), mixed
// ("q a/b"), or finite decimal ("a.b") form. Each error returned wraps
// exactly one sentinel error: [ErrInvalidText], [ErrZeroDenominator], or
// [ErrOutOfRange].
func Parse(text string) (Fraction, error) {
	if text == "" {
		return Fraction{}, parseError(ErrInvalidText, text)
	}
	if wholePart, fracPart, hasSpace := strings.Cut(text, " "); hasSpace {
		numPart, denPart, hasSlash := strings.Cut(fracPart, "/")
		if !hasSlash || strings.Contains(denPart, "/") {
			return Fraction{}, parseError(ErrInvalidText, text)
		}
		wNeg, wDigits := splitSign(wholePart)
		if !isASCIIAllDigits(wDigits) || !isASCIIAllDigits(numPart) || !isASCIIAllDigits(denPart) {
			return Fraction{}, parseError(ErrInvalidText, text)
		}
		if len(wDigits) > MaxDigits || len(numPart) > MaxDigits || len(denPart) > MaxDigits {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		if isAllZeroDigits(denPart) {
			return Fraction{}, parseError(ErrZeroDenominator, text)
		}
		wMag, okW := parseDigitsToU256(wDigits)
		aMag, okA := parseDigitsToU256(numPart)
		bMag, okB := parseDigitsToU256(denPart)
		if !okW || !okA || !okB {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		wb := u256Mul(wMag, bMag)
		total512 := u512Add(wb, u512FromU256(aMag))
		if !total512.fitsU256Max() {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		return makeFraction(total512.low256(), bMag, wNeg, false), nil
	}

	if numPart, denPart, hasSlash := strings.Cut(text, "/"); hasSlash {
		if strings.Contains(denPart, "/") {
			return Fraction{}, parseError(ErrInvalidText, text)
		}
		nNeg, nDigits := splitSign(numPart)
		dNeg, dDigits := splitSign(denPart)
		if !isASCIIAllDigits(nDigits) || !isASCIIAllDigits(dDigits) {
			return Fraction{}, parseError(ErrInvalidText, text)
		}
		if len(nDigits) > MaxDigits || len(dDigits) > MaxDigits {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		if isAllZeroDigits(dDigits) {
			return Fraction{}, parseError(ErrZeroDenominator, text)
		}
		nMag, okN := parseDigitsToU256(nDigits)
		dMag, okD := parseDigitsToU256(dDigits)
		if !okN || !okD {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		return makeFraction(nMag, dMag, nNeg, dNeg), nil
	}

	if intPart, fracPart, hasDot := strings.Cut(text, "."); hasDot {
		if strings.Contains(fracPart, ".") {
			return Fraction{}, parseError(ErrInvalidText, text)
		}
		neg, iDigits := splitSign(intPart)
		if !isASCIIAllDigits(iDigits) || !isASCIIAllDigits(fracPart) {
			return Fraction{}, parseError(ErrInvalidText, text)
		}
		if len(iDigits) > MaxDigits || len(fracPart) > MaxDigits {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		if isAllZeroDigits(iDigits) && isAllZeroDigits(fracPart) {
			return Zero(), nil
		}
		numBig, _ := new(big.Int).SetString(iDigits+fracPart, 10)
		if neg {
			numBig.Neg(numBig)
		}
		denBig := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(fracPart))), nil)
		f, err := FromBigIntReduced(numBig, denBig)
		if err != nil {
			return Fraction{}, parseError(ErrOutOfRange, text)
		}
		return f, nil
	}

	neg, digits := splitSign(text)
	if !isASCIIAllDigits(digits) {
		return Fraction{}, parseError(ErrInvalidText, text)
	}
	if len(digits) > MaxDigits {
		return Fraction{}, parseError(ErrOutOfRange, text)
	}
	nMag, ok := parseDigitsToU256(digits)
	if !ok {
		return Fraction{}, parseError(ErrOutOfRange, text)
	}
	return makeFraction(nMag, u256One, neg, false), nil
}

// MustParse is like [Parse] but panics on error.
func MustParse(text string) Fraction {
	f, err := Parse(text)
	if err != nil {
		panic(err)
	}
	return f
}
