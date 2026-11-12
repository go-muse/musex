package tuplet

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// unsetString is the String form of the unset tuplet. Parse rejects it.
const unsetString = "Tuplet(unset)"

// String returns the canonical notation "actual:normal" in ASCII decimal
// digits without signs or leading zeros, such as "3:2" or "6:4". For the
// unset tuplet it returns "Tuplet(unset)", which Parse rejects.
func (t Tuplet) String() string {
	if !t.IsSet() {
		return unsetString
	}
	return string(t.appendText(make([]byte, 0, 2*len("9223372036854775807")+1)))
}

// appendText appends the canonical notation of a set tuplet to b.
func (t Tuplet) appendText(b []byte) []byte {
	b = strconv.AppendInt(b, int64(t.actual), 10)
	b = append(b, ':')
	return strconv.AppendInt(b, int64(t.normal), 10)
}

// ShortString returns the short notation: the actual count alone, such as
// "3" for 3:2. ParseShort restores the default tuplet for a division, not
// necessarily t. It returns an error wrapping ErrUnset if t is unset.
func (t Tuplet) ShortString() (string, error) {
	if !t.IsSet() {
		return "", unsetError("receiver")
	}
	return strconv.Itoa(t.actual), nil
}

// MustShortString is like ShortString but panics with the returned error.
func (t Tuplet) MustShortString() string {
	return must(t.ShortString())
}

// MarshalText implements encoding.TextMarshaler with the canonical notation
// of String. It returns nil and an error wrapping ErrUnset if t is unset, so
// an unset struct field is not encoded; use the omitzero tag option for an
// optional field.
func (t Tuplet) MarshalText() ([]byte, error) {
	if !t.IsSet() {
		return nil, unsetError("receiver")
	}
	return t.appendText(nil), nil
}

// AppendText implements encoding.TextAppender: it appends the canonical
// notation of String to b. The existing contents b[:len(b)] are not
// modified. It returns b unchanged and an error wrapping ErrUnset if t is
// unset.
func (t Tuplet) AppendText(b []byte) ([]byte, error) {
	if !t.IsSet() {
		return b, unsetError("receiver")
	}
	return t.appendText(b), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// Parse. On error the receiver is not modified. It returns an error wrapping
// ErrInvalidArgument if t is nil, and otherwise the errors of Parse. The text
// is not retained.
func (t *Tuplet) UnmarshalText(text []byte) error {
	if t == nil {
		return fmt.Errorf("%w: nil *Tuplet receiver", ErrInvalidArgument)
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// isCount reports whether s matches the grammar count = nzdigit *digit.
func isCount(s string) bool {
	if s == "" || s[0] < '1' || s[0] > '9' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// countValue converts a string that matches the count grammar to an int,
// reporting false if the value exceeds math.MaxInt.
func countValue(s string) (int, bool) {
	n := 0
	for i := 0; i < len(s); i++ {
		d := int(s[i] - '0')
		if n > (math.MaxInt-d)/10 {
			return 0, false
		}
		n = n*10 + d
	}
	return n, true
}

// Parse parses the canonical notation "actual:normal": two counts of ASCII
// decimal digits without signs, leading zeros, or zero values, separated by
// a single colon. In order of checking, it returns an error wrapping
// ErrInvalidText if text does not match this grammar (spaces, signs, leading
// zeros, zero, other separators, names, and "Tuplet(unset)" are rejected) and
// ErrOutOfRange if a count exceeds math.MaxInt. The grammar is checked over
// the whole text before the counts are converted. The error message quotes an
// excerpt of text of at most 64 bytes followed by "..." if text is longer.
func Parse(text string) (Tuplet, error) {
	before, after, found := strings.Cut(text, ":")
	if !found || !isCount(before) || !isCount(after) {
		return Tuplet{}, parseError(ErrInvalidText, text)
	}
	actual, okActual := countValue(before)
	normal, okNormal := countValue(after)
	if !okActual || !okNormal {
		return Tuplet{}, parseError(ErrOutOfRange, text)
	}
	return Tuplet{actual: actual, normal: normal}, nil
}

// MustParse is like Parse but panics with the returned error.
func MustParse(text string) Tuplet {
	return must(Parse(text))
}

// ParseShort parses the short notation, a single count of ASCII decimal
// digits without sign or leading zeros, and returns Default(count, division).
// In order of checking, it returns an error wrapping ErrInvalidText if text
// does not match the grammar, ErrOutOfRange if the count exceeds math.MaxInt,
// and ErrNoDefault if division is unknown or no default normal count exists.
// Every error message quotes an excerpt of text as Parse does.
func ParseShort(text string, division Division) (Tuplet, error) {
	if !isCount(text) {
		return Tuplet{}, parseError(ErrInvalidText, text)
	}
	actual, ok := countValue(text)
	if !ok {
		return Tuplet{}, parseError(ErrOutOfRange, text)
	}
	if !division.IsValid() {
		return Tuplet{}, fmt.Errorf("%w: unknown division %d: %s",
			ErrNoDefault, uint8(division), quoteExcerpt(text))
	}
	normal, ok := defaultNormal(actual, division)
	if !ok {
		return Tuplet{}, fmt.Errorf("%w: %s division: %s",
			ErrNoDefault, division, quoteExcerpt(text))
	}
	return Tuplet{actual: actual, normal: normal}, nil
}

// MustParseShort is like ParseShort but panics with the returned error.
func MustParseShort(text string, division Division) Tuplet {
	return must(ParseShort(text, division))
}

// ParseName parses a tuplet name in lower case and returns the named tuplet
// for its actual count (for example, Triplet() for "triplet"). It accepts the
// English names "duplet" through "decuplet", the alternative forms "quartole",
// "quintolet", "pentuplet", "sextolet", "sestole", "septolet", and
// "septimole", and the Russian names "дуоль" through "децимоль". Any other
// text, including other letter case and surrounding spaces, yields an error
// wrapping ErrInvalidText that quotes an excerpt of text as Parse does.
func ParseName(text string) (Tuplet, error) {
	n, ok := countByName(text)
	if !ok {
		return Tuplet{}, parseError(ErrInvalidText, text)
	}
	return named(n), nil
}

// MustParseName is like ParseName but panics with the returned error.
func MustParseName(text string) Tuplet {
	return must(ParseName(text))
}
