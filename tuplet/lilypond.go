package tuplet

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/fraction"
)

// ParseLilyPondTuplet parses the LilyPond group marker "\tuplet a/b": the
// prefix "\tuplet " followed by two counts of ASCII decimal digits without
// signs or spaces, separated by a single slash. A count may have leading
// zeros, which do not change its value: "\tuplet 03/2" gives 3:2 and
// "\tuplet 006/004" gives 6:4. A count that is empty or zero in value, such
// as "00", is rejected. The result is the tuplet a:b, stored as given: the
// fraction a/b is the reciprocal coefficient of the group, as in
// [FromReciprocalCoefficient].
//
// The optional span duration of LilyPond ("\tuplet 3/2 4") is not supported:
// any trailing token gives an error wrapping ErrInvalidText. The span chooses
// the written unit of a group, which the caller sets when building the group
// (package duration); it is not a property of the ratio.
//
// In order of checking, it returns an error wrapping ErrInvalidText if text
// does not match the grammar and ErrOutOfRange if the value of a count
// exceeds math.MaxInt; leading zeros are not counted, so a long run of zeros
// before a count in range is accepted. Every error message quotes an excerpt
// of text as [Parse] does.
func ParseLilyPondTuplet(text string) (Tuplet, error) {
	body, ok := strings.CutPrefix(text, "\\tuplet ")
	if !ok {
		return Tuplet{}, parseError(ErrInvalidText, text)
	}

	before, after, found := strings.Cut(body, "/")
	before, okBefore := lyCount(before)
	after, okAfter := lyCount(after)
	if !found || !okBefore || !okAfter {
		return Tuplet{}, parseError(ErrInvalidText, text)
	}

	actual, okActual := countValue(before)
	normal, okNormal := countValue(after)
	if !okActual || !okNormal {
		return Tuplet{}, parseError(ErrOutOfRange, text)
	}

	return Tuplet{actual: actual, normal: normal}, nil
}

// MustParseLilyPondTuplet is like ParseLilyPondTuplet but panics with the
// returned error.
func MustParseLilyPondTuplet(text string) Tuplet {
	return must(ParseLilyPondTuplet(text))
}

// FormatLilyPondTuplet returns the LilyPond group marker of t: "\tuplet "
// followed by the actual and normal counts, without leading zeros, separated
// by a slash, as [ParseLilyPondTuplet] reads them. It returns an error
// wrapping ErrUnset if t is unset.
func FormatLilyPondTuplet(t Tuplet) (string, error) {
	if !t.IsSet() {
		return "", unsetError("receiver")
	}
	return "\\tuplet " + strconv.Itoa(t.actual) + "/" + strconv.Itoa(t.normal), nil
}

// MustFormatLilyPondTuplet is like FormatLilyPondTuplet but panics with the
// returned error.
func MustFormatLilyPondTuplet(t Tuplet) string {
	return must(FormatLilyPondTuplet(t))
}

// ParseLilyPondMultiplier parses the LilyPond duration multiplier "*a/b" or
// its integer form "*a" (as in R1*4): the prefix "*" followed by one or two
// counts of ASCII decimal digits without signs or spaces, separated by a
// single slash. As LilyPond's number syntax allows, a count may have leading
// zeros, which do not change its value: "*02/3" gives the pair (2, 3) and
// "*04" the pair (4, 1). A count that is empty or zero in value is rejected,
// so "*0", "*00", and "*2/0" are errors. The result is the fraction a/b (a/1
// for the integer form) with the stored pair (a, b): a duration multiplied by
// it sounds for that part of its written value. The multiplier declares no
// group, so the result is a coefficient, not a tuplet.
//
// In order of checking, it returns an error wrapping ErrInvalidText if text
// does not match the grammar and ErrOutOfRange if the value of a count
// exceeds math.MaxInt (leading zeros are not counted) or the fraction is
// outside the range of package common/fraction. Every error message quotes
// an excerpt of text as [Parse] does.
func ParseLilyPondMultiplier(text string) (fraction.Fraction, error) {
	body, ok := strings.CutPrefix(text, "*")
	if !ok {
		return fraction.Fraction{}, parseError(ErrInvalidText, text)
	}

	before, after, found := strings.Cut(body, "/")
	before, okBefore := lyCount(before)
	after, okAfter := lyCount(after)
	if !okBefore || (found && !okAfter) {
		return fraction.Fraction{}, parseError(ErrInvalidText, text)
	}

	num, okNum := countValue(before)
	den := 1
	var okDen = true
	if found {
		den, okDen = countValue(after)
	}
	if !okNum || !okDen {
		return fraction.Fraction{}, parseError(ErrOutOfRange, text)
	}

	c, err := fraction.New(num, den)
	if err != nil {
		return fraction.Fraction{}, fractionRangeError("lilypond multiplier", err)
	}
	return c, nil
}

// MustParseLilyPondMultiplier is like ParseLilyPondMultiplier but panics with
// the returned error.
func MustParseLilyPondMultiplier(text string) fraction.Fraction {
	return must(ParseLilyPondMultiplier(text))
}

// FormatLilyPondMultiplier returns the LilyPond duration multiplier of c: "*"
// followed by the stored numerator and denominator, without leading zeros,
// separated by a slash, as [ParseLilyPondMultiplier] reads them; the pair is
// not reduced. A positive coefficient stored with two negative components is
// written with both signs removed, since the grammar has no signs: the pair
// (-2, -3) is written as "*2/3" and (-4, -6) as "*4/6". A denominator of 1,
// after that, is written in the integer form "*a", so both "*4" and "*4/1"
// parse to the same coefficient and format as "*4", and the pair (-4, -1) is
// written as "*4". Parsing the result gives a coefficient equal to c by
// fraction.Fraction.Equal.
//
// In order of checking, it returns an error wrapping ErrInvalidArgument if c
// is unset, ErrInvalidCount if c is not positive, and ErrOutOfRange if a
// component of the stored pair, with its sign removed, does not fit in int,
// such as the pair (math.MinInt, -1).
func FormatLilyPondMultiplier(c fraction.Fraction) (string, error) {
	if !c.IsValid() {
		return "", fmt.Errorf("%w: unset fraction", ErrInvalidArgument)
	}
	if c.Sign() <= 0 {
		return "", fmt.Errorf("%w: coefficient %s is not positive", ErrInvalidCount, c)
	}

	num, den, ok := c.Pair()
	if ok && num < 0 {
		// A positive coefficient has components of the same sign; the
		// negation of math.MinInt does not fit in int.
		ok = num != math.MinInt && den != math.MinInt
		num, den = -num, -den
	}
	if !ok {
		return "", fmt.Errorf("%w: coefficient component exceeds int", ErrOutOfRange)
	}
	if den == 1 {
		return "*" + strconv.Itoa(num), nil
	}
	return "*" + strconv.Itoa(num) + "/" + strconv.Itoa(den), nil
}

// MustFormatLilyPondMultiplier is like FormatLilyPondMultiplier but panics
// with the returned error.
func MustFormatLilyPondMultiplier(c fraction.Fraction) string {
	return must(FormatLilyPondMultiplier(c))
}

// lyCount matches s against the LilyPond count grammar ly-count = *"0" count:
// ASCII decimal digits with optional leading zeros and a value of at least 1.
// It returns s without its leading zeros, ready for countValue, so the range
// check is by value and the cost is linear in len(s) however long the run of
// zeros is.
func lyCount(s string) (string, bool) {
	digits := strings.TrimLeft(s, "0")
	return digits, isCount(digits)
}
