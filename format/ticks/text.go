package ticks

import (
	"math/big"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
)

// The grammar of the package texts (ABNF):
//
//	uint     = "0" / count               ; no leading zeros
//	count    = nzdigit *digit
//	ratio    = uint "/" count
//	decimal  = uint "." 1*digit
//
// Parsing checks the grammar of the whole text first (ErrInvalidText), then
// numbers outside int and fraction components outside common/fraction
// (ErrOutOfRange), then values from left to right by the rules of the
// constructors.

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// isCount reports whether text is a positive decimal integer without leading
// zeros.
func isCount(text string) bool {
	if text == "" || text[0] < '1' || text[0] > '9' {
		return false
	}
	for i := 1; i < len(text); i++ {
		if !isDigit(text[i]) {
			return false
		}
	}
	return true
}

// isUint reports whether text is "0" or a count.
func isUint(text string) bool { return text == "0" || isCount(text) }

// scanRatio splits uint "/" count.
func scanRatio(text string) (num, den string, ok bool) {
	num, den, found := strings.Cut(text, "/")
	if !found || !isUint(num) || !isCount(den) {
		return "", "", false
	}
	return num, den, true
}

// isDecimal reports whether text is uint "." 1*digit.
func isDecimal(text string) bool {
	whole, frac, found := strings.Cut(text, ".")
	if !found || !isUint(whole) || frac == "" {
		return false
	}
	for i := range len(frac) {
		if !isDigit(frac[i]) {
			return false
		}
	}
	return true
}

func intRangeTextError(text string) error {
	return &textError{message: ErrOutOfRange.Error() + ": " + quoteExcerpt(text), causes: []error{ErrOutOfRange}}
}

// uintValue converts a grammatical uint to int; a value above math.MaxInt is
// ErrOutOfRange.
func uintValue(text, digits string) (int, error) {
	n, ok := new(big.Int).SetString(digits, 10)
	if !ok || !n.IsInt64() {
		return 0, intRangeTextError(text)
	}
	return int(n.Int64()), nil
}

// ratioValue converts a grammatical ratio to an exact value; a component
// outside common/fraction is ErrOutOfRange with fraction.ErrOutOfRange.
func ratioValue(text, num, den string) (*big.Rat, error) {
	n, okN := new(big.Int).SetString(num, 10)
	d, okD := new(big.Int).SetString(den, 10)
	limit := fraction.MaxComponentBig()
	if !okN || !okD || n.Cmp(limit) > 0 || d.Cmp(limit) > 0 {
		return nil, numberRangeTextError(text)
	}
	return new(big.Rat).SetFrac(n, d), nil
}

// decimalValue converts a grammatical decimal to an exact value; a value whose
// reduced components are outside common/fraction is ErrOutOfRange.
func decimalValue(text, digits string) (*big.Rat, error) {
	value, ok := new(big.Rat).SetString(digits)
	if !ok {
		return nil, grammarError(text)
	}
	if _, err := fraction.FromBigRat(value); err != nil {
		return nil, numberRangeTextError(text)
	}
	return value, nil
}

// ratText returns the reduced text n/d of a rational value.
func ratText(value *big.Rat) string { return value.String() }

// modeName returns the canonical name of a valid rounding mode.
func modeName(mode rounding.Mode) string { return mode.String() }

// parseModeName returns the mode with the canonical name, or false.
func parseModeName(name string) (rounding.Mode, bool) {
	mode, err := rounding.Parse(name)
	if err != nil {
		return 0, false
	}
	return mode, true
}

// splitList splits text by sep; an empty text gives no items.
func splitList(text string, sep string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, sep)
}
