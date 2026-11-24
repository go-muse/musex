package tuplet

import (
	"fmt"
	"strconv"
	"strings"
)

// ABCMarker is a tuplet marker of ABC notation: the tuplet and the number of
// following events it applies to. The zero value is not a valid marker;
// [ParseABCMarker] returns markers with a set tuplet and Events at least 1.
type ABCMarker struct {
	// Tuplet is the tuplet of the marker.
	Tuplet Tuplet
	// Events is the number of events the marker applies to, at least 1.
	Events int
}

// abcDefaultNormal returns the normal count that the ABC 2.1 standard fixes
// for the short form (p in the given division: 3, 2, 3, 2, 3 for
// p = 2, 3, 4, 6, 8, and 3 in compound (Ternary) meter or 2 otherwise for
// p = 5, 7, 9. Other actual counts have no ABC default. division must be
// valid.
func abcDefaultNormal(actual int, division Division) (int, bool) {
	switch actual {
	case 2, 4, 8:
		return 3, true
	case 3, 6:
		return 2, true
	case 5, 7, 9:
		if division == Ternary {
			return 3, true
		}
		return 2, true
	default:
		return 0, false
	}
}

// ParseABCMarker parses an ABC tuplet marker in the meter context given by
// division: "(p", "(p::", "(p:q", "(p:q:r", or "(p::r", where p, q, and r are
// counts of ASCII decimal digits without signs, leading zeros, or zero values.
// Binary is the context of a simple meter such as 4/4, Ternary of a compound
// meter such as 6/8.
//
// The normal count q defaults by the ABC 2.1 standard, not by [Default]:
// 3, 2, 3, 2, 3 for p = 2, 3, 4, 6, 8, and 3 in Ternary or 2 in Binary for
// p = 5, 7, 9; so "(5" gives 5:2 in Binary and 5:3 in Ternary. The event
// count r defaults to p. "(p::" omits both q and r and reads exactly as "(p"
// in the same division, as section 4.13 of the standard states for "(3::".
// Actual counts other than 2 through 9 have no ABC default and require an
// explicit q. The forms "(p:" and "(p:q:", with an empty last part, are
// rejected.
//
// In order of checking, it returns an error wrapping ErrInvalidText if text
// does not match the grammar, ErrOutOfRange if a count exceeds math.MaxInt,
// and ErrNoDefault if division is unknown or p has no ABC default and q is
// not given. Every error message quotes an excerpt of text as [Parse] does.
func ParseABCMarker(text string, division Division) (ABCMarker, error) {
	body, ok := strings.CutPrefix(text, "(")
	if !ok {
		return ABCMarker{}, parseError(ErrInvalidText, text)
	}

	var pStr, qStr, rStr string
	var hasQ, hasR bool

	if i := strings.Index(body, "::"); i >= 0 {
		pStr, rStr = body[:i], body[i+2:]
		// "(p::" omits both q and r and reads as "(p" (abc 2.1, section 4.13);
		// any other text after "::" must be the event count.
		hasR = rStr != ""
	} else {
		parts := strings.Split(body, ":")
		switch len(parts) {
		case 1:
			pStr = parts[0]
		case 2:
			pStr, qStr = parts[0], parts[1]
			hasQ = true
		case 3:
			pStr, qStr, rStr = parts[0], parts[1], parts[2]
			hasQ, hasR = true, true
		default:
			return ABCMarker{}, parseError(ErrInvalidText, text)
		}
	}

	if !isCount(pStr) || (hasQ && !isCount(qStr)) || (hasR && !isCount(rStr)) {
		return ABCMarker{}, parseError(ErrInvalidText, text)
	}

	p, okP := countValue(pStr)
	q, okQ := countValue(qStr)
	r, okR := countValue(rStr)
	if !okP || (hasQ && !okQ) || (hasR && !okR) {
		return ABCMarker{}, parseError(ErrOutOfRange, text)
	}

	if !division.IsValid() {
		return ABCMarker{}, fmt.Errorf("%w: unknown division %d: %s",
			ErrNoDefault, uint8(division), quoteExcerpt(text))
	}

	if !hasQ {
		normal, hasDefault := abcDefaultNormal(p, division)
		if !hasDefault {
			return ABCMarker{}, fmt.Errorf("%w: actual count %d, %s division: %s",
				ErrNoDefault, p, division, quoteExcerpt(text))
		}
		q = normal
	}

	if !hasR {
		r = p
	}

	return ABCMarker{Tuplet: Tuplet{actual: p, normal: q}, Events: r}, nil
}

// MustParseABCMarker is like ParseABCMarker but panics with the returned
// error.
func MustParseABCMarker(text string, division Division) ABCMarker {
	return must(ParseABCMarker(text, division))
}

// FormatABCMarker returns the ABC marker text of m in the meter context given
// by division. The formatter omits the parts that [ParseABCMarker] restores
// in the same division: "(p" when q is the ABC default for p and r equals p;
// "(p:q" when r equals p; "(p:q:r" otherwise. The "(p::" and "(p::r" forms
// are never produced. Parsing the result in the same division returns an
// equal marker.
//
// In order of checking, it returns an error wrapping ErrUnset if the tuplet
// of m is unset, ErrInvalidCount if Events is less than 1, and ErrNoDefault
// if division is unknown.
func FormatABCMarker(m ABCMarker, division Division) (string, error) {
	if !m.Tuplet.IsSet() {
		return "", unsetError("marker tuplet")
	}
	if m.Events < 1 {
		return "", fmt.Errorf("%w: events %d", ErrInvalidCount, m.Events)
	}
	if !division.IsValid() {
		return "", fmt.Errorf("%w: unknown division %d", ErrNoDefault, uint8(division))
	}

	p, q := m.Tuplet.MustPair()
	text := "(" + strconv.Itoa(p)

	normal, hasDefault := abcDefaultNormal(p, division)
	if !hasDefault || normal != q || m.Events != p {
		text += ":" + strconv.Itoa(q)
		if m.Events != p {
			text += ":" + strconv.Itoa(m.Events)
		}
	}

	return text, nil
}

// MustFormatABCMarker is like FormatABCMarker but panics with the returned
// error.
func MustFormatABCMarker(m ABCMarker, division Division) string {
	return must(FormatABCMarker(m, division))
}
