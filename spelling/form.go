package spelling

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
)

// ErrInvalidForm reports that a text or JSON value is not the machine form of
// a name.
//
// Parse, Spelling.UnmarshalText, and Spelling.UnmarshalJSON return errors that
// match ErrInvalidForm with errors.Is, and MustParse panics with such an error
// instead of returning it. Some of these errors also match
// [alteration.ErrZeroDenominator] or [alteration.ErrOutOfRange], as described
// in the Machine form section of the package documentation; they never match
// [letter.ErrInvalid]. The text of the error is not part of the API.
var ErrInvalidForm = errors.New("spelling: invalid machine form")

const (
	// unsetForm is the machine form of the unset name.
	unsetForm = "unset"

	// letterChars are the letters of the machine form, indexed by
	// letter.Letter.Ordinal.
	letterChars = "CDEFGAB"

	// maxFormLen is the largest length of a machine form: a letter, a sign,
	// a numerator of at most 13 digits (MaxMagnitude*MaxDenominator is 2^40),
	// a slash, and a denominator of at most 5 digits.
	maxFormLen = 1 + 1 + 13 + 1 + 5
)

// String returns the machine form of s, such as "C", "C+1", "C+1/2", or
// "B-7/2", and "unset" for the unset name. The form is described in the
// Machine form section of the package documentation. It is not a notation:
// it uses no accidental symbols, is not localized, and Parse reads it back.
func (s Spelling) String() string {
	return string(s.appendForm(make([]byte, 0, maxFormLen)))
}

// MarshalText implements [encoding.TextMarshaler]. It returns the machine
// form of s, the same text as Spelling.String; for the unset name it is
// "unset". The error is always nil.
func (s Spelling) MarshalText() ([]byte, error) {
	return s.appendForm(make([]byte, 0, maxFormLen)), nil
}

// AppendText implements [encoding.TextAppender]. It appends the machine form
// of s, the same text as Spelling.String, to b and returns the extended
// buffer. It never changes b[:len(b)], but it may write into the spare
// capacity of b; the result is owned by the caller. The error is always nil.
//
// Because AppendText may write into b, concurrent calls with the same b, or
// with buffers that share a backing array, must be synchronized by the
// caller.
func (s Spelling) AppendText(b []byte) ([]byte, error) {
	return s.appendForm(b), nil
}

// MarshalJSON implements [encoding/json.Marshaler]. It returns the machine
// form of s as a JSON string, such as "C+1", and "unset" for the unset name.
// The form contains only ASCII characters that need no escaping. The error is
// always nil.
func (s Spelling) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, maxFormLen+2)
	b = append(b, '"')
	b = s.appendForm(b)
	b = append(b, '"')

	return b, nil
}

// Parse returns the name whose machine form is text, such as "C", "C+1",
// "C+1/2", or "B-7/2". The text "unset" gives the unset name.
//
// Only the exact canonical form is accepted, as described in the Machine form
// section of the package documentation. Otherwise Parse returns the unset name
// and an error that matches ErrInvalidForm; for some inputs the error also
// matches [alteration.ErrZeroDenominator] or [alteration.ErrOutOfRange].
// The machine form is not a notation: "C#", "C♯", and "Cis" are errors.
func Parse(text string) (Spelling, error) {
	return parse(text)
}

// MustParse is like Parse but panics if text is not the machine form of a
// name. The panic value is the error returned by Parse, which matches
// ErrInvalidForm.
//
// It is intended for literals and values known to be valid, for example
// spelling.MustParse("B-7/2").
func MustParse(text string) Spelling {
	return must(Parse(text))
}

// UnmarshalText implements [encoding.TextUnmarshaler]. It reads the machine
// form of a name like Parse(string(text)) and stores the name in s. An empty
// or nil text is an error.
//
// On error, s is not changed. UnmarshalText neither modifies nor retains text.
// The receiver s must not be nil: calling UnmarshalText through a nil
// *Spelling is a caller error and may panic.
func (s *Spelling) UnmarshalText(text []byte) error {
	v, err := parse(string(text))
	if err != nil {
		return err
	}

	*s = v

	return nil
}

// UnmarshalJSON implements [encoding/json.Unmarshaler]. It reads a name from
// the JSON value data and stores it in s:
//
//   - a JSON string is read by its decoded content, like Parse; escaped
//     characters are allowed, so "C\u002B1" reads as C sharp;
//   - JSON null gives the unset name.
//
// Any other JSON value, such as a number, a boolean, an array, or an object,
// and empty or malformed JSON, returns an error that matches ErrInvalidForm.
// On error, s is not changed. UnmarshalJSON neither modifies nor retains data.
// The receiver s must not be nil: calling UnmarshalJSON through a nil
// *Spelling is a caller error and may panic.
func (s *Spelling) UnmarshalJSON(data []byte) error {
	// JSON whitespace around the value is allowed.
	v := bytes.Trim(data, " \t\r\n")

	switch {
	case string(v) == "null":
		*s = Spelling{}

		return nil
	case len(v) > 0 && v[0] == '"':
		// AppendUnquote rejects anything but one valid JSON string,
		// including trailing bytes and invalid escapes or UTF-8.
		text, err := jsontext.AppendUnquote(nil, v)
		if err != nil {
			return fmt.Errorf("%w: JSON %q: %v", ErrInvalidForm, data, err)
		}

		return s.UnmarshalText(text)
	default:
		return fmt.Errorf("%w: JSON %q is not a string or null", ErrInvalidForm, data)
	}
}

// appendForm appends the machine form of s to b.
func (s Spelling) appendForm(b []byte) []byte {
	if !s.IsValid() {
		return append(b, unsetForm...)
	}

	b = append(b, letterChars[s.l.Ordinal()])

	num, den := s.a.Fraction()

	switch {
	case num > 0:
		b = append(b, '+')
	case num < 0:
		b = append(b, '-')
		num = -num
	default:
		return b
	}

	b = strconv.AppendInt(b, int64(num), 10)
	if den > 1 {
		b = append(b, '/')
		b = strconv.AppendInt(b, int64(den), 10)
	}

	return b
}

// parse reads the machine form text. The checks follow the order described
// in the Machine form section of the package documentation: the shape of the
// text first, then the numbers, the value, and the canonical form.
func parse(text string) (Spelling, error) {
	if text == unsetForm {
		return Spelling{}, nil
	}

	l, sign, num, den, ok := splitForm(text)
	if !ok {
		return Spelling{}, invalidForm(text, nil)
	}

	if sign == 0 {
		// A single letter: the natural name.
		return Spelling{l: l}, nil
	}

	if den == "0" {
		return Spelling{}, invalidForm(text, alteration.ErrZeroDenominator)
	}

	n, err := strconv.Atoi(string(sign) + num)
	if err != nil {
		return Spelling{}, invalidForm(text, alteration.ErrOutOfRange)
	}

	d := 1
	if den != "" {
		if d, err = strconv.Atoi(den); err != nil {
			return Spelling{}, invalidForm(text, alteration.ErrOutOfRange)
		}
	}

	a, err := alteration.FromFraction(n, d)
	if err != nil {
		return Spelling{}, invalidForm(text, err)
	}

	s := Spelling{l: l, a: a}
	if s.String() != text {
		return Spelling{}, invalidForm(text, nil)
	}

	return s, nil
}

// splitForm splits text that has the shape of a name or of a number:
//
//	letter [ sign number [ "/" ( number | "0" ) ] ]
//
// where number is ASCII decimal digits without leading zeros. It reports
// false for any other text. For a single letter, sign is 0. The denominator is
// empty if there is no slash.
func splitForm(text string) (l letter.Letter, sign byte, num, den string, ok bool) {
	if text == "" {
		return 0, 0, "", "", false
	}

	l = parseLetter(text[0])
	if !l.IsValid() {
		return 0, 0, "", "", false
	}

	rest := text[1:]
	if rest == "" {
		return l, 0, "", "", true
	}

	sign = rest[0]
	if sign != '+' && sign != '-' {
		return 0, 0, "", "", false
	}

	num, den, slash := strings.Cut(rest[1:], "/")
	if !isNumber(num) {
		return 0, 0, "", "", false
	}

	if slash && den != "0" && !isNumber(den) {
		return 0, 0, "", "", false
	}

	return l, sign, num, den, true
}

// parseLetter returns the letter written as the ASCII character c, or the
// zero letter if c is not one of A to G.
func parseLetter(c byte) letter.Letter {
	i := strings.IndexByte(letterChars, c)
	if i < 0 {
		return 0
	}

	return letter.MustFromOrdinal(i)
}

// isNumber reports whether s is a non-empty sequence of ASCII decimal digits
// without a leading zero.
func isNumber(s string) bool {
	if s == "" || s[0] == '0' {
		return false
	}

	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// invalidForm returns an error for text that is not a machine form. The error
// matches ErrInvalidForm and, if cause is not nil, also cause.
func invalidForm(text string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w %q", ErrInvalidForm, text)
	}

	return fmt.Errorf("%w %q: %w", ErrInvalidForm, text, cause)
}
