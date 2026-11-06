package cents

import (
	"math/big"
	"strings"
)

// ParseScala reads the pitch value of a line of a Scala tuning file (.scl):
//
//	line   = { " " | TAB } token [ ( " " | TAB ) rest ]
//	token  = cents | ratio          ; the longest run without spaces and tabs
//	cents  = [ "-" ] { digit } "." { digit }    ; at least one digit in total
//	ratio  = digits [ "/" digits ]              ; "n" means n/1
//
// Leading spaces and tabs are skipped; the value ends at the first space,
// tab, or the end of the line. A value with a point is exact cents ("408."
// is 408 cents, ".5" is 1/2 cent); a value without a point is a ratio ("5" is
// the ratio 5:1, and "10/20" is -1200 cents). The rest of the line is not
// parsed and is returned as is: it is empty or starts with a space or a tab,
// so ParseScala(" 5/4   E") returns the ratio 5:4 and "   E". A line break is
// a character of the value: pass the line without it. Reading whole .scl
// files is outside the package.
//
// The errors are checked in this order. [ErrInvalidText]: no value, a
// character right after the value (such as "5/4E" or "408.0.5"), "+5.0" or
// "−5.0", a negative ratio, a zero term of a ratio, or a point without
// digits. [ErrTooComplex]: a run of more than MaxDigits digits in the value;
// the limits apply to the value only, not to the leading spaces and the
// rest. [ErrPrimeLimit]: a ratio with a prime greater than 31 (see
// [ParseScalaApprox]). [ErrTooComplex]: the record is beyond the record
// limits. [ErrOutOfRange]: the value is outside the range. On error, the
// results are the zero value and "".
func ParseScala(line string) (Cents, string, error) {
	return parseScala(line, false)
}

// MustParseScala is like [ParseScala] but panics with the error. It is meant
// for literals and values known to be valid.
func MustParseScala(line string) (Cents, string) {
	v, rest, err := ParseScala(line)
	if err != nil {
		panic(err)
	}
	return v, rest
}

// ParseScalaApprox is like [ParseScala], but a ratio with a prime greater
// than 31 gives the cents 1200·log2(n/d) of the exact ratio, rounded
// correctly to the grid of 1/GridDenominator cent, instead of
// [ErrPrimeLimit]: "37/32" gives 251.344039 cents. Other ratios and cents are
// exact, as with ParseScala.
func ParseScalaApprox(line string) (Cents, string, error) {
	return parseScala(line, true)
}

// MustParseScalaApprox is like [ParseScalaApprox] but panics with the error.
// It is meant for literals and values known to be valid.
func MustParseScalaApprox(line string) (Cents, string) {
	v, rest, err := ParseScalaApprox(line)
	if err != nil {
		panic(err)
	}
	return v, rest
}

// parseScala parses a Scala pitch value, approximately for primes greater
// than 31 if approx is true.
func parseScala(line string, approx bool) (Cents, string, error) {
	start := 0
	for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
		start++
	}
	end := start
	for end < len(line) && line[end] != ' ' && line[end] != '\t' {
		end++
	}
	token, rest := line[start:end], line[end:]
	var rc *record
	var err error
	if strings.IndexByte(token, '.') >= 0 {
		rc, err = scalaCents(token)
	} else {
		rc, err = scalaRatio(token, approx)
	}
	if err == nil {
		err = rc.check("ParseScala")
	}
	if err != nil {
		return Cents{}, "", err
	}
	return fromRecord(rc), rest, nil
}

// scalaCents parses a Scala value in cents.
func scalaCents(token string) (*record, error) {
	i := 0
	negative := strings.HasPrefix(token, "-")
	if negative {
		i++
	}
	j := digitsEnd(token, i)
	intPart := token[i:j]
	if j == len(token) || token[j] != '.' {
		return nil, textError(token, "invalid Scala cents")
	}
	k := digitsEnd(token, j+1)
	fracPart := token[j+1 : k]
	if k != len(token) || intPart == "" && fracPart == "" {
		return nil, textError(token, "invalid Scala cents")
	}
	if err := checkDigits(token, intPart, fracPart); err != nil {
		return nil, err
	}
	rc := new(record)
	rc.r.SetFrac(atoi("0"+intPart+fracPart), pow10(len(fracPart)))
	if negative {
		rc.r.Neg(&rc.r)
	}
	return rc, nil
}

// scalaRatio parses a Scala ratio.
func scalaRatio(token string, approx bool) (*record, error) {
	if token == "" {
		return nil, textError(token, "no Scala value")
	}
	i := digitsEnd(token, 0)
	num, den := token[:i], "1"
	if i < len(token) {
		if token[i] != '/' {
			return nil, textError(token, "invalid Scala ratio")
		}
		j := digitsEnd(token, i+1)
		den = token[i+1 : j]
		if j != len(token) || den == "" {
			return nil, textError(token, "invalid Scala ratio")
		}
	}
	if num == "" {
		return nil, textError(token, "invalid Scala ratio")
	}
	if isZeroDigits(num) || isZeroDigits(den) {
		return nil, textError(token, "zero term of a Scala ratio")
	}
	if err := checkDigits(token, num, den); err != nil {
		return nil, err
	}
	n, d := atoi(num), atoi(den)
	rc, err := ratioRecord(n, d, "ParseScala")
	if err == nil || !approx {
		return rc, err
	}
	return gridLog(new(big.Rat).SetFrac(n, d)), nil
}
