package tempo

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// mark is the content of an exact or an approximate marking: a written unit
// and beats per minute in reduced form. The zero mark is unset.
type mark struct {
	unit duration.Written
	bpm  fraction.Fraction
}

// isSet reports whether m is set.
func (m mark) isSet() bool {
	return m.unit.IsSet()
}

// newMark checks a unit and beats per minute in the order of New and returns
// the mark.
func newMark(op string, unit duration.Written, bpm fraction.Fraction) (mark, error) {
	if !unit.IsSet() {
		return mark{}, invalidError(op, "unset unit")
	}
	if unsetFraction(bpm) {
		return mark{}, invalidError(op, "unset beats per minute")
	}
	if bpm.Sign() <= 0 {
		return mark{}, simpleError(ErrNoTempo, op, "beats per minute not positive")
	}
	return mark{unit: unit, bpm: bpm.Reduce()}, nil
}

// markOf returns the mark of a unit and the positive value r, or an error
// wrapping ErrOutOfRange if r is outside the range of package common/fraction.
func markOf(op string, unit duration.Written, r *big.Rat) (mark, error) {
	f, err := toFraction(op, r)
	if err != nil {
		return mark{}, err
	}
	return mark{unit: unit, bpm: f}, nil
}

// unitDuration returns the duration of the unit of the set mark m.
func (m mark) unitDuration() duration.Duration {
	d, _ := m.unit.Duration()
	return d
}

// speed returns u·N, the number of whole notes per minute of the set mark m.
func (m mark) speed() *big.Rat {
	return mul(unitRat(m.unit), ratOf(m.bpm))
}

// unitFraction returns u, the duration of the unit of the set mark m in whole
// notes.
func (m mark) unitFraction() fraction.Fraction {
	return durFraction(m.unitDuration())
}

// speedFraction returns the speed u·N of m as a fraction.
func (m mark) speedFraction(op string) (fraction.Fraction, error) {
	speed, err := m.unitFraction().Mul(m.bpm)
	return fractionResult(op, speed, err)
}

// quartersPerMinute returns 4·u·N for the set mark m. It is one product of
// three factors rather than 4 times the speed: the speed u·N may be outside
// the range when 4·u·N is not, since its reduced denominator can be four
// times larger.
func (m mark) quartersPerMinute(op string) (fraction.Fraction, error) {
	q, err := fraction.Product(four, m.unitFraction(), m.bpm)
	return fractionResult(op, q, err)
}

// compare orders two set marks by speed and then by unit, as
// duration.Written.Compare does.
func (m mark) compare(other mark) int {
	if c := m.speed().Cmp(other.speed()); c != 0 {
		return c
	}
	return m.unit.Compare(other.unit)
}

// compareMarks orders unset marks before set ones and set marks by compare.
func compareMarks(a, b mark) int {
	if !a.isSet() || !b.isSet() {
		return compareSet(a.isSet(), b.isSet())
	}
	return a.compare(b)
}

// withUnit returns the mark of the same speed with unit u: N·u/u′. It
// checks u and the range of the result.
func (m mark) withUnit(op string, u duration.Written) (mark, error) {
	if !u.IsSet() {
		return mark{}, invalidError(op, "unset unit")
	}
	return markOf(op, u, quo(m.speed(), unitRat(u)))
}

// appendMark appends the text of the set mark m with the sign between the
// unit and the number.
func appendMark(b []byte, m mark, sign string) []byte {
	b, _ = m.unit.AppendText(b)
	b = append(b, sign...)
	return appendNumber(b, m.bpm)
}
