package degree

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Decimal returns the canonical decimal form of the number of d: ASCII
// digits without a sign, spaces, or leading zeros. The note is not included.
// An unset degree reports ErrUnset and returns an empty string.
func (d Degree) Decimal() (string, error) {
	if !d.IsValid() {
		return "", unsetError("Decimal", "unset degree")
	}

	return strconv.Itoa(d.number), nil
}

// Roman returns the canonical uppercase Roman form of the number of d. It is
// defined for numbers 1 through MaxRomanNumber. An unset degree reports
// ErrUnset; a greater number reports ErrRomanOutOfRange. The string is empty
// on either error.
func (d Degree) Roman() (string, error) {
	if !d.IsValid() {
		return "", unsetError("Roman", "unset degree")
	}

	if d.number > MaxRomanNumber {
		return "", romanRangeError("Roman", d.number)
	}

	return formatRoman(d.number), nil
}

// Parse returns a degree without a note from one canonical decimal or Roman
// number. It accepts no spaces, signs, leading zeros, mixed forms, lowercase
// or Unicode Roman numerals, or noncanonical subtraction.
//
// Errors: ErrParse for an invalid form, ErrInvalidNumber for "0", and
// ErrOutOfRange for a decimal number above MaxNumber. The result is then the
// unset degree.
func Parse(text string) (Degree, error) {
	if text == "" {
		return Degree{}, parseError(text)
	}

	if isDecimal(text) {
		if text == "0" {
			return Degree{}, invalidNumberError("Parse", "number is zero")
		}

		if text[0] == '0' {
			return Degree{}, parseError(text)
		}

		number, err := strconv.Atoi(text)
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				return Degree{}, outOfRangeError("Parse", "decimal number exceeds MaxNumber")
			}

			return Degree{}, parseError(text)
		}

		return Degree{number: number}, nil
	}

	return parseRoman(text)
}

// String returns a debugging form with the number and either the debugging
// form of the note or "no note". The unset degree is "Degree(unset)". It is
// not a serialization and is never parsed back.
func (d Degree) String() string {
	if !d.IsValid() {
		return "Degree(unset)"
	}

	if !d.HasNote() {
		return fmt.Sprintf("Degree(%d, no note)", d.number)
	}

	return fmt.Sprintf("Degree(%d, %s)", d.number, d.n.String())
}

// isDecimal reports whether text consists only of ASCII decimal digits. The
// caller has already excluded the empty string.
func isDecimal(text string) bool {
	for i := range len(text) {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}

	return true
}

// formatRoman returns the canonical Roman form of a number from 1 through
// MaxRomanNumber.
func formatRoman(number int) string {
	places := [...]struct {
		value  int
		symbol string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var b strings.Builder
	for _, place := range places {
		for number >= place.value {
			b.WriteString(place.symbol)
			number -= place.value
		}
	}

	return b.String()
}

// parseRoman decodes a candidate and accepts it only when formatting the
// decoded number reproduces the candidate exactly. Early range rejection also
// keeps arbitrarily long invalid input from overflowing the accumulator.
func parseRoman(text string) (Degree, error) {
	total, previous := 0, 0

	for i := len(text) - 1; i >= 0; i-- {
		var value int
		switch text[i] {
		case 'I':
			value = 1
		case 'V':
			value = 5
		case 'X':
			value = 10
		case 'L':
			value = 50
		case 'C':
			value = 100
		case 'D':
			value = 500
		case 'M':
			value = 1000
		default:
			return Degree{}, parseError(text)
		}

		if value < previous {
			total -= value
		} else {
			total += value
			previous = value
		}

		if total > MaxRomanNumber || total < -MaxRomanNumber {
			return Degree{}, parseError(text)
		}
	}

	if total < 1 || formatRoman(total) != text {
		return Degree{}, parseError(text)
	}

	return Degree{number: total}, nil
}
