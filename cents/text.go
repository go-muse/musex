package cents

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Kind is an exact text form of a value.
type Kind uint8

// The exact text forms. Every value has exactly one canonical form,
// returned by [Cents.Kind]; [Cents.TextAs] writes any form a value has.
const (
	KindDecimal  Kind = iota // decimal cents: "700", "7.1", "-13.7"
	KindFraction             // fraction of cents with unit: "1200/19 cents"
	KindRatio                // frequency ratio: "3:2"
	KindPower                // product of powers: "5^(1/4)"
)

// String returns "decimal", "fraction", "ratio", or "power", and "Kind(n)"
// for another value.
func (k Kind) String() string {
	switch k {
	case KindDecimal:
		return "decimal"
	case KindFraction:
		return "fraction"
	case KindRatio:
		return "ratio"
	case KindPower:
		return "power"
	}
	return "Kind(" + strconv.Itoa(int(k)) + ")"
}

// Options is a set of options of the text of cents. They apply to the forms
// KindDecimal and KindFraction and to [Cents.Decimal]; the forms KindRatio
// and KindPower are always ASCII and have no sign. The zero Options is ASCII
// without the optional unit and without "+". Unknown bits are ignored.
type Options uint8

// The options of the text of cents.
const (
	Typographic Options = 1 << iota // minus U+2212 and unit "¢" instead of "-" and "cents"
	WithUnit                        // append the unit to decimal cents
	WithPlus                        // "+" before positive cents
)

const (
	typoMinus = "−" // −
	centSign  = "¢" // ¢
	nbsp      = " "
)

// maxRatioBits bounds the terms of a ratio of at most MaxDigits digits:
// 10^100 < 2^333.
const maxRatioBits = 333

var maxRatioTerm = pow10(MaxDigits) // a term of KindRatio is less than this

// decimalDigits returns the number of digits after the point of the exact
// decimal of a rational with the reduced denominator den, or -1 if the
// decimal does not terminate.
func decimalDigits(den *big.Int) int {
	d := new(big.Int).Set(den)
	twos := int(d.TrailingZeroBits())
	d.Rsh(d, uint(twos))
	fives := 0
	var q, m big.Int
	five := big.NewInt(5)
	for {
		q.QuoRem(d, five, &m)
		if m.Sign() != 0 {
			break
		}
		d.Set(&q)
		fives++
	}
	if d.Cmp(bigOne) != 0 {
		return -1
	}
	return max(twos, fives)
}

// ratioText returns the terms of KindRatio of the record, if it has the form.
func (rc *record) ratioText() (n, d *big.Int, ok bool) {
	n, d, ok = rc.ratioParts(maxRatioBits)
	if !ok || n.Cmp(maxRatioTerm) >= 0 || d.Cmp(maxRatioTerm) >= 0 {
		return nil, nil, false
	}
	return n, d, true
}

// kind returns the canonical form of the record.
func (rc *record) kind() Kind {
	if rc.isRational() {
		if decimalDigits(rc.r.Denom()) >= 0 {
			return KindDecimal
		}
		return KindFraction
	}
	if _, _, ok := rc.ratioText(); ok {
		return KindRatio
	}
	return KindPower
}

// Kind returns the canonical text form of c: KindDecimal for a rational
// value with a terminating decimal, KindFraction for another rational value,
// KindRatio for the cents of a ratio of integers of at most MaxDigits digits
// each, and KindPower for any other value.
func (c Cents) Kind() Kind {
	return c.value().kind()
}

// String returns the canonical text of c: the canonical form, ASCII, without
// "+" and without the optional unit, such as "700", "7.1", "-13.7",
// "1200/19 cents", "3:2", "5^(1/4)", or "0". It is also the output of %v.
// [Parse] reads it back to an equal value.
func (c Cents) String() string {
	rc := c.value()
	s, _ := rc.text(rc.kind(), 0)
	return s
}

// Text returns the canonical form of c with the options: 13.7 cents below
// zero with Typographic|WithUnit are "−13.7 ¢", and 7.1 cents with WithPlus
// are "+7.1". The options do not apply to KindRatio and KindPower.
func (c Cents) Text(opts Options) string {
	rc := c.value()
	s, _ := rc.text(rc.kind(), opts)
	return s
}

// TextAs returns c in the exact form k with the options, or ("", false) if c
// has no such form or k is unknown. Octave().TextAs(KindRatio, 0) is
// ("2:1", true), and 700 cents have no KindRatio.
func (c Cents) TextAs(k Kind, opts Options) (string, bool) {
	return c.value().text(k, opts)
}

// text returns the record in the form k with the options.
func (rc *record) text(k Kind, opts Options) (string, bool) {
	switch k {
	case KindDecimal:
		if !rc.isRational() {
			return "", false
		}
		digits := decimalDigits(rc.r.Denom())
		if digits < 0 {
			return "", false
		}
		m := new(big.Int).Mul(rc.r.Num(), pow10(digits))
		m.Quo(m, rc.r.Denom())
		sign := m.Sign()
		return centsText(sign, scaledDecimal(m.Abs(m), digits), opts, opts&WithUnit != 0), true
	case KindFraction:
		if !rc.isRational() {
			return "", false
		}
		number := new(big.Int).Abs(rc.r.Num()).String() + "/" + rc.r.Denom().String()
		return centsText(rc.r.Sign(), number, opts, true), true
	case KindRatio:
		n, d, ok := rc.ratioText()
		if !ok {
			return "", false
		}
		return n.String() + ":" + d.String(), true
	case KindPower:
		return rc.powerText(), true
	}
	return "", false
}

// centsText returns the text of a number of cents: the sign of the written
// number, the number without a sign, and the unit if unit is true.
func centsText(sign int, number string, opts Options, unit bool) string {
	var b strings.Builder
	switch {
	case sign < 0 && opts&Typographic != 0:
		b.WriteString(typoMinus)
	case sign < 0:
		b.WriteByte('-')
	case sign > 0 && opts&WithPlus != 0:
		b.WriteByte('+')
	}
	b.WriteString(number)
	if unit {
		b.WriteByte(' ')
		switch {
		case opts&Typographic != 0:
			b.WriteString(centSign)
		case number == "1":
			b.WriteString("cent")
		default:
			b.WriteString("cents")
		}
	}
	return b.String()
}

// powerText returns the record as a product of powers.
func (rc *record) powerText() string {
	var b strings.Builder
	factors := 0
	single := false // the only factor so far has the exponent 1
	write := func(p int64, e *big.Rat) {
		if e.Sign() == 0 {
			return
		}
		if factors > 0 {
			b.WriteByte('*')
		}
		factors++
		b.WriteString(strconv.FormatInt(p, 10))
		single = false
		switch {
		case e.IsInt() && e.Num().Cmp(bigOne) == 0:
			single = factors == 1
		case e.IsInt():
			b.WriteByte('^')
			b.WriteString(e.Num().String())
		default:
			b.WriteString("^(")
			b.WriteString(e.Num().String())
			b.WriteByte('/')
			b.WriteString(e.Denom().String())
			b.WriteByte(')')
		}
	}
	write(2, new(big.Rat).Quo(&rc.r, rat1200))
	for i := range rc.e {
		write(primes[i], &rc.e[i])
	}
	switch {
	case factors == 0:
		return "2^0"
	case factors == 1 && single:
		b.WriteString("^1")
	}
	return b.String()
}

// Decimal returns c in decimal with digits digits after the point, correctly
// rounded, for reading: an exact half, which only a rational value can have,
// is rounded away from zero, so 1/8 cent with 2 digits is "0.13". A negative
// digits means 0, and with 0 digits there is no point. A number rounded to
// zero has no sign. The options apply as to KindDecimal. The decimal of an
// irrational value is rounded; use [Cents.String] or [Cents.MarshalText] to
// exchange values exactly. The cost grows with the number of digits.
func (c Cents) Decimal(digits int, opts Options) string {
	digits = max(digits, 0)
	m := c.value().roundValue(pow10(digits), bigOne)
	return centsText(m.Sign(), scaledDecimal(new(big.Int).Abs(m), digits), opts, opts&WithUnit != 0)
}

// Scala returns c as the pitch value of a line of a Scala tuning file
// (.scl), ASCII: the ratio "n/d" if c is the cents of a ratio whose terms are
// at most 2^31-1, which any conforming program reads, such as "3/2", "2/1",
// or "1/1" for 0; otherwise cents with at least one digit after the point,
// Decimal(max(digits, 1), 0), such as "700.0" or "63.15789". Irrational
// values are rounded in this form, a limitation of the format.
func (c Cents) Scala(digits int) string {
	rc := c.value()
	if n, d, ok := rc.ratioParts(31); ok {
		return n.String() + "/" + d.String()
	}
	return c.Decimal(max(digits, 1), 0)
}

// Format implements [fmt.Formatter]. The verbs %v and %s write
// [Cents.String], with the flag + [Cents.Text](WithPlus); %q writes
// [strconv.Quote] of String; %f and %F write [Cents.Decimal] with the
// precision (6 by default) and WithPlus for the flag +, without a unit. A
// width pads with spaces on the left, or on the right with the flag -,
// counting Unicode characters. The flags 0, # and space, the flag + of %q,
// and the precision of %v, %s and %q are ignored. Another verb writes
// "%!verb(cents.Cents=String)", without width, precision, or flags.
func (c Cents) Format(f fmt.State, verb rune) {
	var s string
	switch verb {
	case 'v', 's':
		if f.Flag('+') {
			s = c.Text(WithPlus)
		} else {
			s = c.String()
		}
	case 'q':
		s = strconv.Quote(c.String())
	case 'f', 'F':
		prec, ok := f.Precision()
		if !ok {
			prec = 6
		}
		var opts Options
		if f.Flag('+') {
			opts = WithPlus
		}
		s = c.Decimal(prec, opts)
	default:
		_, _ = fmt.Fprintf(f, "%%!%c(cents.Cents=%s)", verb, c.String())
		return
	}
	if width, ok := f.Width(); ok {
		if pad := width - utf8.RuneCountInString(s); pad > 0 {
			if f.Flag('-') {
				s += strings.Repeat(" ", pad)
			} else {
				s = strings.Repeat(" ", pad) + s
			}
		}
	}
	_, _ = f.Write([]byte(s))
}

// MarshalText implements [encoding.TextMarshaler]: it returns
// []byte(c.String()) and a nil error. The slice is new and belongs to the
// caller.
func (c Cents) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// AppendText implements [encoding.TextAppender]: it appends the canonical
// text [Cents.String] to b and returns the extended buffer; the error is
// always nil. If b has enough spare capacity for the text, the text is
// written into it and the result shares the array of b; the bytes after the
// appended text are not touched. Otherwise the result is a new array, and
// the array of b is not modified at all. b[:len(b)] is never modified, and b
// is not retained. A call writes into the spare capacity of b, so the caller
// synchronizes calls that share it (see Concurrency in the package
// documentation).
func (c Cents) AppendText(b []byte) ([]byte, error) {
	return append(b, c.String()...), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]: it parses text as
// [Parse] does and, on success, sets c to the result. On error, c is not
// changed. A nil or empty text gives [ErrInvalidText]. text is neither
// modified nor retained.
func (c *Cents) UnmarshalText(text []byte) error {
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*c = v
	return nil
}
