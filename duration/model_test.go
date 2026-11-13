package duration_test

import (
	"math/big"
	"slices"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"
)

// This file holds independent models of business-spec on math/big. They do
// not call the operations they check: the value of a written duration
// (ID-1), the single-sign rule (ID-4, ID-5) by enumeration, and the tie rule
// of P-3 by enumeration of the allowed signs.

// baseRow is one row of business-spec 2.2 and architecture 4.11.
type baseRow struct {
	base    duration.Base
	num     int64 // value in whole notes, num/den
	den     int64
	text    string // architecture 4.12
	us      string // "" means "no name"
	uk      string
	ru      string
	lilyPad string // LilyPond word form, "" when the text is the LilyPond number
}

// baseTable is the series of 14 bases from the longest (P-1). The names
// are the first table of architecture 4.11 as changed by task 28.2 (P-12):
// maxima is "maxima" in both variants by the product convention of P-12.
var baseTable = []baseRow{
	{duration.Maxima, 8, 1, "maxima", "maxima", "maxima", "", `\maxima`},
	{duration.Longa, 4, 1, "longa", "long", "longa", "", `\longa`},
	{duration.Breve, 2, 1, "breve", "double whole note", "breve", "бревис", `\breve`},
	{duration.Whole, 1, 1, "1", "whole note", "semibreve", "целая", ""},
	{duration.Half, 1, 2, "2", "half note", "minim", "половинная", ""},
	{duration.Quarter, 1, 4, "4", "quarter note", "crotchet", "четвертная", ""},
	{duration.Eighth, 1, 8, "8", "eighth note", "quaver", "восьмая", ""},
	{duration.Sixteenth, 1, 16, "16", "sixteenth note", "semiquaver", "шестнадцатая", ""},
	{duration.ThirtySecond, 1, 32, "32", "thirty-second note", "demisemiquaver", "тридцатьвторая", ""},
	{duration.SixtyFourth, 1, 64, "64", "sixty-fourth note", "hemidemisemiquaver", "шестьдесятчетвёртая", ""},
	{duration.HundredTwentyEighth, 1, 128, "128", "hundred twenty-eighth note", "semihemidemisemiquaver", "стодвадцатьвосьмая", ""},
	{duration.TwoHundredFiftySixth, 1, 256, "256", "two hundred fifty-sixth note", "demisemihemidemisemiquaver", "", ""},
	{duration.FiveHundredTwelfth, 1, 512, "512", "", "", "", ""},
	{duration.ThousandTwentyFourth, 1, 1024, "1024", "", "", "", ""},
}

// parseOnlyNames is the second table of architecture 4.11 (P-12, KP-54):
// other English names and spellings that ParseBase accepts, but that Name
// and BritishName never return.
var parseOnlyNames = []struct {
	base  duration.Base
	names []string
}{
	{duration.Maxima, []string{"large", "duplex longa", "octuple note", "octuple whole note", "octuple entire musical note"}},
	{duration.Longa, []string{"quadruple note", "quadruple whole note"}},
	{duration.Breve, []string{"double-whole note"}},
	{duration.Sixteenth, []string{"16th note"}},
	{duration.ThirtySecond, []string{"32nd note"}},
	{duration.SixtyFourth, []string{"64th note"}},
	{duration.HundredTwentyEighth, []string{"one-hundred-twenty-eighth note", "128th note"}},
	{duration.TwoHundredFiftySixth, []string{"two-hundred-fifty-sixth note", "256th note"}},
}

// baseForm is one string that ParseBase accepts and the base it denotes.
type baseForm struct {
	text string
	base duration.Base
}

// acceptedBaseForms lists every string of architecture 4.12 that ParseBase
// accepts, in the case of the tables: the text of a base, the LilyPond word,
// the US, UK, and Russian names of the first table of 4.11, and the strings
// of its second table. Equal strings of one base may repeat.
func acceptedBaseForms() []baseForm {
	var forms []baseForm

	for _, row := range baseTable {
		for _, text := range []string{row.text, row.lilyPad, row.us, row.uk, row.ru} {
			if text != "" {
				forms = append(forms, baseForm{text: text, base: row.base})
			}
		}
	}

	for _, row := range parseOnlyNames {
		for _, text := range row.names {
			forms = append(forms, baseForm{text: text, base: row.base})
		}
	}

	return forms
}

// expectedParseBase is the model of ParseBase by architecture 4.12: the base
// of the accepted string equal to text under strings.EqualFold, if any.
func expectedParseBase(text string) (duration.Base, bool) {
	for _, form := range acceptedBaseForms() {
		if strings.EqualFold(form.text, text) {
			return form.base, true
		}
	}

	return 0, false
}

// seriesIndex returns the 0-based place of a valid base in the series.
func seriesIndex(b duration.Base) int {
	for i, row := range baseTable {
		if row.base == b {
			return i
		}
	}

	panic("seriesIndex: invalid base")
}

// baseRat is the value of a base by the table.
func baseRat(b duration.Base) *big.Rat {
	row := baseTable[seriesIndex(b)]

	return rat(row.num, row.den)
}

// chainRat is the value of the chain coefficient: the product of m/n.
func chainRat(chain []tuplet.Tuplet) *big.Rat {
	c := big.NewRat(1, 1)

	for _, t := range chain {
		n, m := t.MustPair()
		c.Mul(c, new(big.Rat).SetFrac(big.NewInt(int64(m)), big.NewInt(int64(n))))
	}

	return c
}

// dotFactor is 2 − 1/2^k.
func dotFactor(k int) *big.Rat { return new(big.Rat).Sub(big.NewRat(2, 1), pow2(-k)) }

// writtenRat is ID-1: b · (2 − 1/2^k) · C.
func writtenRat(b duration.Base, k int, chain []tuplet.Tuplet) *big.Rat {
	v := new(big.Rat).Mul(baseRat(b), dotFactor(k))

	return v.Mul(v, chainRat(chain))
}

// isDyadic reports whether the reduced denominator of r is a power of two.
func isDyadic(r *big.Rat) bool {
	d := r.Denom()

	return new(big.Int).And(d, new(big.Int).Sub(d, big.NewInt(1))).Sign() == 0
}

// sign is an allowed sign of the tie rule: a base with k dots.
type sign struct {
	base duration.Base
	dots int
	v    *big.Rat
}

// maxSingleDots bounds the enumeration of single signs: any value with 512
// or more dots is outside the fraction range (architecture 4.5).
const maxSingleDots = 520

// singleSigns enumerates every base and every 0 <= k <= maxSingleDots whose
// value without tuplets equals v. ID-4 says there is at most one.
func singleSigns(v *big.Rat) []sign {
	var found []sign

	for _, row := range baseTable {
		b := rat(row.num, row.den)

		// b·(2 − 1/2^k) grows with k from b toward 2b, so only bases with
		// b <= v < 2b can match and the enumeration stops above v.
		if v.Cmp(b) < 0 || v.Cmp(new(big.Rat).Mul(b, big.NewRat(2, 1))) >= 0 {
			continue
		}

		for k := 0; k <= maxSingleDots; k++ {
			s := new(big.Rat).Mul(b, dotFactor(k))

			c := s.Cmp(v)
			if c == 0 {
				found = append(found, sign{base: row.base, dots: k, v: s})
			}

			if c > 0 {
				break
			}
		}
	}

	return found
}

// allowedSigns is S(L, D) of architecture 4.8: bases from longest to the
// 1024th with 0 <= k <= maxDots, sorted by decreasing value.
func allowedSigns(longest duration.Base, maxDots int) []sign {
	var signs []sign

	for _, row := range baseTable[seriesIndex(longest):] {
		for k := 0; k <= maxDots; k++ {
			signs = append(signs, sign{base: row.base, dots: k, v: new(big.Rat).Mul(rat(row.num, row.den), dotFactor(k))})
		}
	}

	slices.SortFunc(signs, func(a, b sign) int { return b.v.Cmp(a.v) })

	return signs
}

// tieOutcome is the result of the model of the tie rule.
type tieOutcome int

const (
	tieOK         tieOutcome = iota // the rule finishes
	tieNotNotable                   // E-5: no sign fits a positive remainder, or v' is not dyadic
)

// tieModel is the result of the rule of P-3 for v': the outcome, the number
// of parts, and the parts when there are at most buildLimit of them.
type tieModel struct {
	outcome tieOutcome
	count   *big.Int
	parts   []sign
}

// buildLimit bounds the parts the model builds; MaxParts + 1 tells E-9.
const buildLimit = 1025

// greedyTie runs the rule of P-3 by its text: while r > 0 take the largest
// allowed sign not greater than r. A non-dyadic v' never reaches zero, since
// every sign is dyadic. While r is at least the largest sign of S, that
// sign is taken, so the repeated first steps are counted by one division;
// the remaining steps are taken one by one, and each sign is taken at most
// once there, since neighboring signs of S differ at most twice.
func greedyTie(v *big.Rat, longest duration.Base, maxDots int) tieModel {
	if !isDyadic(v) {
		return tieModel{outcome: tieNotNotable}
	}

	signs := allowedSigns(longest, maxDots)
	largest := signs[0]

	quotient := new(big.Rat).Quo(v, largest.v)
	repeats := new(big.Int).Quo(quotient.Num(), quotient.Denom())
	r := new(big.Rat).Sub(v, new(big.Rat).Mul(largest.v, new(big.Rat).SetInt(repeats)))

	var tail []sign

	i := 0
	for r.Sign() > 0 {
		for i < len(signs) && signs[i].v.Cmp(r) > 0 {
			i++
		}

		if i == len(signs) {
			return tieModel{outcome: tieNotNotable}
		}

		tail = append(tail, signs[i])
		r.Sub(r, signs[i].v)
	}

	count := new(big.Int).Add(repeats, big.NewInt(int64(len(tail))))
	m := tieModel{outcome: tieOK, count: count}

	if count.Cmp(big.NewInt(buildLimit)) <= 0 {
		for range repeats.Int64() {
			m.parts = append(m.parts, largest)
		}

		m.parts = append(m.parts, tail...)
	}

	return m
}

// writtenOfSigns builds the expected parts with the chain as given.
func writtenOfSigns(signs []sign, chain []tuplet.Tuplet) []duration.Written {
	out := make([]duration.Written, len(signs))
	for i, s := range signs {
		out[i] = wr(s.base, s.dots, chain...)
	}

	return out
}

// testChains are chains for model checks, including the empty chain,
// unreduced and neutral tuplets, and nesting.
func testChains() [][]tuplet.Tuplet {
	return [][]tuplet.Tuplet{
		nil,
		{tuplet.Triplet()},
		{tp(6, 4)},
		{tuplet.Duplet()},
		{tp(1, 1)},
		{tuplet.Quintuplet(), tuplet.Triplet()},
		{tp(7, 4)},
	}
}

// testValues are positive values for model checks: dyadic and non-dyadic,
// short and long, including bounds of the series.
func testValues() []*big.Rat {
	var values []*big.Rat

	for _, d := range []int64{1, 2, 3, 4, 5, 6, 8, 12, 16, 24, 32, 48, 64, 128, 256, 1024, 2048, 4096} {
		for n := int64(1); n <= 40; n++ {
			values = append(values, rat(n, d))
		}
	}

	return values
}
