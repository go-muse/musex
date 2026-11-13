package duration

import (
	"fmt"
	"strings"

	"github.com/go-muse/musex/common/fraction"
)

// Base is a note value without dots: one of 14 values from maxima to 1024th.
//
// The constants are declared from the longest base to the shortest, so a
// smaller constant is a longer base and cmp.Compare of two bases orders them
// from longest to shortest. Base(0) and values greater than
// ThousandTwentyFourth are not bases: IsValid reports false for them, and
// operations that need a base return an error wrapping ErrInvalidArgument.
//
// The values of longa (4 whole notes) and maxima (8 whole notes) are a
// convention of the product that continues the doubling of the series;
// mensural relations, including triple ones, are not modeled.
type Base uint8

// The 14 bases of the series, from the longest to the shortest. The comment
// of each constant gives its duration in whole notes.
const (
	Maxima               Base = iota + 1 // 8 whole notes (product convention)
	Longa                                // 4 whole notes (product convention)
	Breve                                // 2
	Whole                                // 1
	Half                                 // 1/2
	Quarter                              // 1/4
	Eighth                               // 1/8
	Sixteenth                            // 1/16
	ThirtySecond                         // 1/32
	SixtyFourth                          // 1/64
	HundredTwentyEighth                  // 1/128
	TwoHundredFiftySixth                 // 1/256
	FiveHundredTwelfth                   // 1/512
	ThousandTwentyFourth                 // 1/1024
)

// baseCount is the number of bases in the series.
const baseCount = int(ThousandTwentyFourth)

// invalidBaseString is the String form of a value that is not a base.
const invalidBaseString = "Base(invalid)"

// nameTable selects a table of texts and names of the bases.
type nameTable uint8

const (
	textTable nameTable = iota // canonical text of String
	lilyTable                  // LilyPond notations that differ from the text
	usTable                    // American English names
	ukTable                    // British English names
	ruTable                    // Russian names
)

// nameTables lists the tables accepted by ParseBase.
const nameTables = 5

// entry returns the entry of table t for the valid base b; an empty string
// means that the base has no entry. The tables are array values, not mutable
// package variables.
func (b Base) entry(t nameTable) string {
	switch t {
	case textTable:
		return [...]string{
			Maxima: "maxima", Longa: "longa", Breve: "breve", Whole: "1", Half: "2", Quarter: "4",
			Eighth: "8", Sixteenth: "16", ThirtySecond: "32", SixtyFourth: "64",
			HundredTwentyEighth: "128", TwoHundredFiftySixth: "256", FiveHundredTwelfth: "512",
			ThousandTwentyFourth: "1024",
		}[b]
	case lilyTable:
		return [...]string{Maxima: `\maxima`, Longa: `\longa`, Breve: `\breve`, ThousandTwentyFourth: ""}[b]
	case usTable:
		return [...]string{
			Maxima: "maxima", Longa: "long", Breve: "double whole note", Whole: "whole note",
			Half: "half note", Quarter: "quarter note", Eighth: "eighth note", Sixteenth: "sixteenth note",
			ThirtySecond: "thirty-second note", SixtyFourth: "sixty-fourth note",
			HundredTwentyEighth: "hundred twenty-eighth note", TwoHundredFiftySixth: "two hundred fifty-sixth note",
			ThousandTwentyFourth: "",
		}[b]
	case ukTable:
		return [...]string{
			Maxima: "maxima", Longa: "longa", Breve: "breve", Whole: "semibreve", Half: "minim",
			Quarter: "crotchet", Eighth: "quaver", Sixteenth: "semiquaver", ThirtySecond: "demisemiquaver",
			SixtyFourth: "hemidemisemiquaver", HundredTwentyEighth: "semihemidemisemiquaver",
			TwoHundredFiftySixth: "demisemihemidemisemiquaver", ThousandTwentyFourth: "",
		}[b]
	default:
		return [...]string{
			Breve: "бревис", Whole: "целая", Half: "половинная", Quarter: "четвертная", Eighth: "восьмая",
			Sixteenth: "шестнадцатая", ThirtySecond: "тридцатьвторая", SixtyFourth: "шестьдесятчетвёртая",
			HundredTwentyEighth: "стодвадцатьвосьмая", ThousandTwentyFourth: "",
		}[b]
	}
}

// otherName is an English name or spelling of a base that only ParseBase
// accepts; Name and BritishName do not return it.
type otherName struct {
	text string
	base Base
}

// otherNameCount is the number of other English names and spellings.
const otherNameCount = 15

// otherNames returns the other English names and spellings accepted by
// ParseBase, each with its base. The list is an array value, not a mutable
// package variable; every text differs, ignoring case, from the texts of all
// other bases.
func otherNames() [otherNameCount]otherName {
	return [otherNameCount]otherName{
		{"large", Maxima},
		{"duplex longa", Maxima},
		{"octuple note", Maxima},
		{"octuple whole note", Maxima},
		{"octuple entire musical note", Maxima},
		{"quadruple note", Longa},
		{"quadruple whole note", Longa},
		{"double-whole note", Breve},
		{"16th note", Sixteenth},
		{"32nd note", ThirtySecond},
		{"64th note", SixtyFourth},
		{"one-hundred-twenty-eighth note", HundredTwentyEighth},
		{"128th note", HundredTwentyEighth},
		{"two-hundred-fifty-sixth note", TwoHundredFiftySixth},
		{"256th note", TwoHundredFiftySixth},
	}
}

// Bases returns the 14 bases in declaration order, from maxima to 1024th.
// The returned slice has length and capacity 14 and belongs to the caller.
func Bases() []Base {
	bases := make([]Base, baseCount)
	for i := range bases {
		bases[i] = Base(i + 1)
	}
	return bases
}

// BaseOf returns the base whose duration equals d: 1/4 gives Quarter. It
// returns an error wrapping ErrUnset if d is unset and ErrInvalidArgument if
// d is not the duration of a base, such as 3/8, 16, or 1/2048.
func BaseOf(d Duration) (Base, error) {
	const op = "BaseOf"
	if !d.IsSet() {
		return 0, unsetError(op, "duration")
	}
	num, den, _ := d.v.ReducedBigPair()
	var e int
	switch {
	case den.IsInt64() && den.Int64() == 1 && isPow2(num):
		e = num.BitLen() - 1
	case num.IsInt64() && num.Int64() == 1 && isPow2(den):
		e = -(den.BitLen() - 1)
	default:
		return 0, invalidError(op, "not the duration of a base")
	}
	b, ok := baseOfExp(e)
	if !ok {
		return 0, invalidError(op, "not the duration of a base")
	}
	return b, nil
}

// MustBaseOf is like BaseOf but panics with the returned error.
func MustBaseOf(d Duration) Base {
	return must(BaseOf(d))
}

// exp returns the binary exponent of the duration of a valid base: 3 for
// maxima, 0 for whole, and −10 for 1024th.
func (b Base) exp() int {
	return int(Whole) - int(b)
}

// baseOfExp returns the base with duration 2^e and reports whether it exists.
func baseOfExp(e int) (Base, bool) {
	if e > Maxima.exp() || e < ThousandTwentyFourth.exp() {
		return 0, false
	}
	return Base(int(Whole) - e), true
}

// IsValid reports whether b is one of the 14 base constants.
func (b Base) IsValid() bool {
	return b >= Maxima && b <= ThousandTwentyFourth
}

// Equal reports whether b and other are the same value. It is equivalent to
// == for all values, including values that are not bases.
func (b Base) Equal(other Base) bool {
	return b == other
}

// invalidBaseError returns the error for a value that is not a base.
func invalidBaseError(op string, b Base) error {
	return invalidError(op, fmt.Sprintf("base %d outside the series", uint8(b)))
}

// fraction returns the duration of a valid base as a reduced fraction.
func (b Base) fraction() fraction.Fraction {
	f, _ := fromRat(pow2Rat(b.exp()))
	return f
}

// Duration returns the duration of b: 8, 4, 2, 1, 1/2, and so on to 1/1024.
// It returns an error wrapping ErrInvalidArgument if b is not a base.
func (b Base) Duration() (Duration, error) {
	if !b.IsValid() {
		return Duration{}, invalidBaseError("Base.Duration", b)
	}
	return Duration{v: b.fraction()}, nil
}

// MustDuration is like Duration but panics with the returned error.
func (b Base) MustDuration() Duration {
	return must(b.Duration())
}

// Written returns the written duration of b without dots and tuplets. It
// returns an error wrapping ErrInvalidArgument if b is not a base.
func (b Base) Written() (Written, error) {
	if !b.IsValid() {
		return Written{}, invalidBaseError("Base.Written", b)
	}
	return Written{base: b, value: b.fraction()}, nil
}

// MustWritten is like Written but panics with the returned error.
func (b Base) MustWritten() Written {
	return must(b.Written())
}

// Longer returns the neighboring base twice as long: Breve for Whole. It
// returns an error wrapping ErrInvalidArgument if b is not a base and
// ErrOutsideSeries for Maxima.
func (b Base) Longer() (Base, error) {
	return b.shift("Base.Longer", 1)
}

// MustLonger is like Longer but panics with the returned error.
func (b Base) MustLonger() Base {
	return must(b.Longer())
}

// Shorter returns the neighboring base half as long: Half for Whole. It
// returns an error wrapping ErrInvalidArgument if b is not a base and
// ErrOutsideSeries for ThousandTwentyFourth.
func (b Base) Shorter() (Base, error) {
	return b.shift("Base.Shorter", -1)
}

// MustShorter is like Shorter but panics with the returned error.
func (b Base) MustShorter() Base {
	return must(b.Shorter())
}

// Shift returns the base steps steps longer (steps > 0) or shorter
// (steps < 0); zero returns b. It returns an error wrapping
// ErrInvalidArgument if b is not a base and ErrOutsideSeries if the result is
// beyond either end of the series, for any int including math.MinInt and
// math.MaxInt.
func (b Base) Shift(steps int) (Base, error) {
	return b.shift("Base.Shift", steps)
}

// MustShift is like Shift but panics with the returned error.
func (b Base) MustShift(steps int) Base {
	return must(b.Shift(steps))
}

// shift implements Shift for the operation op.
func (b Base) shift(op string, steps int) (Base, error) {
	if !b.IsValid() {
		return 0, invalidBaseError(op, b)
	}
	return b.shiftValid(op, steps)
}

// shiftValid shifts a valid base without overflow: steps is compared with the
// allowed interval before the subtraction.
func (b Base) shiftValid(op string, steps int) (Base, error) {
	if steps > int(b)-int(Maxima) || steps < int(b)-int(ThousandTwentyFourth) {
		return 0, simpleError(ErrOutsideSeries, op)
	}
	return Base(int(b) - steps), nil
}

// name returns the entry of table t for b.
func (b Base) name(op string, t nameTable) (string, bool, error) {
	if !b.IsValid() {
		return "", false, invalidBaseError(op, b)
	}
	n := b.entry(t)
	return n, n != "", nil
}

// Name returns the American English name of b, such as "quarter note" or
// "hundred twenty-eighth note", and true. If the base has no such name (the
// 512th and the 1024th), it returns "", false, and no error: the package does
// not invent names, and the identifiers of the constants are not names. The
// name "maxima" of Maxima is a convention of the product. It returns an error
// wrapping ErrInvalidArgument if b is not a base.
func (b Base) Name() (name string, ok bool, err error) {
	return b.name("Base.Name", usTable)
}

// MustName is like Name but panics with the returned error.
func (b Base) MustName() (name string, ok bool) {
	return must2(b.Name())
}

// BritishName returns the British English name of b, such as "crotchet" or
// "demisemiquaver", and true. If the base has no such name (the 512th and the
// 1024th), it returns "", false, and no error. The name "maxima" of Maxima is
// a convention of the product. It returns an error wrapping
// ErrInvalidArgument if b is not a base.
func (b Base) BritishName() (name string, ok bool, err error) {
	return b.name("Base.BritishName", ukTable)
}

// MustBritishName is like BritishName but panics with the returned error.
func (b Base) MustBritishName() (name string, ok bool) {
	return must2(b.BritishName())
}

// RussianName returns the Russian name of b, such as "четвертная", and true.
// If the base has no such name (maxima, longa, and bases shorter than the
// 128th), it returns "", false, and no error. It returns an error wrapping
// ErrInvalidArgument if b is not a base.
func (b Base) RussianName() (name string, ok bool, err error) {
	return b.name("Base.RussianName", ruTable)
}

// MustRussianName is like RussianName but panics with the returned error.
func (b Base) MustRussianName() (name string, ok bool) {
	return must2(b.RussianName())
}

// String returns the canonical text of b: "maxima", "longa", "breve", or the
// LilyPond number "1", "2", "4", and so on to "1024". For a value that is not
// a base it returns "Base(invalid)", which ParseBase rejects.
func (b Base) String() string {
	if !b.IsValid() {
		return invalidBaseString
	}
	return b.entry(textTable)
}

// MarshalText implements encoding.TextMarshaler with the canonical text of
// String. It returns nil and an error wrapping ErrInvalidArgument if b is not
// a base.
func (b Base) MarshalText() ([]byte, error) {
	if !b.IsValid() {
		return nil, invalidBaseError("Base.MarshalText", b)
	}
	return []byte(b.entry(textTable)), nil
}

// AppendText implements encoding.TextAppender: it appends the canonical text
// of String to dst without modifying dst[:len(dst)]. It returns dst unchanged
// and an error wrapping ErrInvalidArgument if b is not a base.
func (b Base) AppendText(dst []byte) ([]byte, error) {
	if !b.IsValid() {
		return dst, invalidBaseError("Base.AppendText", b)
	}
	return append(dst, b.entry(textTable)...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the rules of
// ParseBase. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if b is nil, and otherwise the errors of
// ParseBase. The text is not retained.
func (b *Base) UnmarshalText(text []byte) error {
	if b == nil {
		return invalidError("Base.UnmarshalText", "nil receiver")
	}
	v, err := ParseBase(string(text))
	if err != nil {
		return err
	}
	*b = v
	return nil
}

// ParseBase returns the base named by text, ignoring the case of letters
// (strings.EqualFold). It accepts:
//
//   - the canonical text of String, such as "breve" or "4";
//   - the LilyPond notations \maxima, \longa, and \breve;
//   - the American English, British English, and Russian names returned by
//     Name, BritishName, and RussianName;
//   - the other English names and spellings that Name and BritishName do not
//     return: "large", "duplex longa", "octuple note", "octuple whole note",
//     and "octuple entire musical note" for Maxima; "quadruple note" and
//     "quadruple whole note" for Longa; "double-whole note" for Breve;
//     "16th note" for Sixteenth; "32nd note" for ThirtySecond; "64th note"
//     for SixtyFourth; "one-hundred-twenty-eighth note" and "128th note" for
//     HundredTwentyEighth; and "two-hundred-fifty-sixth note" and
//     "256th note" for TwoHundredFiftySixth.
//
// Each accepted text names exactly one base. Apart from case, text must equal
// an accepted form: surrounding spaces, other spaces or hyphens inside, and
// other forms such as "8th note", "512th note", or "octuple" are not
// accepted. Any other text returns an error wrapping ErrInvalidText.
func ParseBase(text string) (Base, error) {
	for b := Maxima; b <= ThousandTwentyFourth; b++ {
		for t := range nameTable(nameTables) {
			if n := b.entry(t); n != "" && strings.EqualFold(text, n) {
				return b, nil
			}
		}
	}
	for _, o := range otherNames() {
		if strings.EqualFold(text, o.text) {
			return o.base, nil
		}
	}
	return 0, parseError(classInvalidText, text)
}

// MustParseBase is like ParseBase but panics with the returned error.
func MustParseBase(text string) Base {
	return must(ParseBase(text))
}
