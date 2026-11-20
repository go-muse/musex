package ticks

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/fraction"
)

// Quantized is a rounded tick count with its exact deviation: the rounded
// value minus the original value, in whole notes. The zero value is unset.
// Quantized is comparable: a == b reports the same as a.Equal(b).
type Quantized struct {
	ticks     int
	deviation fraction.Fraction
	set       bool
}

// IsSet reports whether q is set.
func (q Quantized) IsSet() bool { return q.set }

// Ticks returns the rounded number of ticks.
func (q Quantized) Ticks() (int, error) {
	if !q.IsSet() {
		return 0, unsetError("Quantized.Ticks", "receiver")
	}
	return q.ticks, nil
}

// MustTicks is like Ticks but panics with its error.
func (q Quantized) MustTicks() int { return must(q.Ticks()) }

// Deviation returns the result minus the original value in whole notes. It is
// negative when the value was rounded down.
func (q Quantized) Deviation() (fraction.Fraction, error) {
	if !q.IsSet() {
		return fraction.Fraction{}, unsetError("Quantized.Deviation", "receiver")
	}
	return q.deviation, nil
}

// MustDeviation is like Deviation but panics with its error.
func (q Quantized) MustDeviation() fraction.Fraction { return must(q.Deviation()) }

// Equal reports whether q and other have the same ticks and deviation.
func (q Quantized) Equal(other Quantized) bool { return q == other }

// String returns the ticks, the sign and the absolute deviation, such as
// "69+1/4480", "68-1/3360" or "720+0/1", or "Quantized(unset)".
func (q Quantized) String() string {
	if !q.IsSet() {
		return "Quantized(unset)"
	}
	deviation := ratOfFraction(q.deviation)
	sign := "+"
	if deviation.Sign() < 0 {
		sign = "-"
		deviation.Neg(deviation)
	}
	return strconv.Itoa(q.ticks) + sign + ratText(deviation)
}

// MarshalText implements encoding.TextMarshaler.
func (q Quantized) MarshalText() ([]byte, error) {
	if !q.IsSet() {
		return nil, unsetError("Quantized.MarshalText", "receiver")
	}
	return []byte(q.String()), nil
}

// AppendText implements encoding.TextAppender.
func (q Quantized) AppendText(dst []byte) ([]byte, error) {
	if !q.IsSet() {
		return dst, unsetError("Quantized.AppendText", "receiver")
	}
	return append(dst, q.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves q unchanged on
// error.
func (q *Quantized) UnmarshalText(text []byte) error {
	if q == nil {
		return invalidError("Quantized.UnmarshalText", "nil receiver")
	}
	v, err := ParseQuantized(string(text))
	if err != nil {
		return err
	}
	*q = v
	return nil
}

// ParseQuantized parses the text of String. Unreduced deviations are accepted
// and reduced; "-0/1" is a zero deviation.
func ParseQuantized(text string) (Quantized, error) {
	i := strings.IndexAny(text, "+-")
	if i < 0 || !isUint(text[:i]) {
		return Quantized{}, grammarError(text)
	}
	num, den, ok := scanRatio(text[i+1:])
	if !ok {
		return Quantized{}, grammarError(text)
	}
	ticks, err := uintValue(text, text[:i])
	if err != nil {
		return Quantized{}, err
	}
	deviation, err := ratioValue(text, num, den)
	if err != nil {
		return Quantized{}, err
	}
	if text[i] == '-' {
		deviation.Neg(deviation)
	}
	dev, err := fraction.FromBigRat(deviation)
	if err != nil {
		return Quantized{}, numberRangeTextError(text)
	}
	return Quantized{ticks: ticks, deviation: dev, set: true}, nil
}

// MustParseQuantized is like ParseQuantized but panics with its error.
func MustParseQuantized(text string) Quantized { return must(ParseQuantized(text)) }
