package duration_test

import (
	"errors"
	"math/big"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Independent recognizers of the ABNF of architecture 4.12. A text that
// does not match is E-10; a matching text is accepted or fails with E-11.
const writtenPattern = `(maxima|longa|breve|1024|512|256|128|64|32|16|8|4|2|1)\.*` +
	`(\[[1-9][0-9]*:[1-9][0-9]*(,[1-9][0-9]*:[1-9][0-9]*)*\])?`

var (
	durationGrammar = regexp.MustCompile(`^[1-9][0-9]*/[1-9][0-9]*$`)
	positionGrammar = regexp.MustCompile(`^(0|[1-9][0-9]*)/[1-9][0-9]*$`)
	writtenGrammar  = regexp.MustCompile(`^` + writtenPattern + `$`)
	tiedGrammar     = regexp.MustCompile(`^` + writtenPattern + `(~` + writtenPattern + `)+$`)
)

// textSeeds are seeds shared by the parse fuzz tests.
var textSeeds = []string{
	"3/8", "6/16", "0/1", "0/7", "1/0", "03/8", "", " ", "4", "4.", "4..[3:2]", "8[3:2]", "4.[5:4,3:2]", "breve..", "maxima",
	"2~4", "4[3:2]~8[3:2]", "4[]", "4[3:2,]", "16.", "3", "1024.", "Breve", `\breve`, "crotchet", "четвертная", "Quarter Note",
	"Duration(unset)", "Tied(unset)", "4[9223372036854775808:1]", maxCompPlus + "/1", "4" + strings.Repeat(".", 300),
	"\xff", "ё", strings.Repeat("9", 80) + "/1",
}

// checkParseFailure checks a rejected text: zero result, E-10 exactly when
// the recognizer does not match, otherwise E-11, and the error text.
func checkParseFailure(t *testing.T, err error, text string, grammar *regexp.Regexp) {
	t.Helper()

	if grammar.MatchString(text) {
		assert.True(t, errors.Is(err, duration.ErrOutOfRange), "a text of the grammar fails only with E-11: %v", err)
		assert.False(t, errors.Is(err, duration.ErrInvalidText))
	} else {
		assertErr(t, err, duration.ErrInvalidText)
	}

	checkErrorText(t, err, text, "fuzz")
}

// FuzzParse covers O-16 and architecture 4.12 for Parse.
func FuzzParse(f *testing.F) {
	for _, s := range textSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			got duration.Duration
			err error
		)

		_, panicked := capturePanic(func() { got, err = duration.Parse(text) })
		require.False(t, panicked)

		if err != nil {
			assert.Equal(t, duration.Duration{}, got)
			checkParseFailure(t, err, text, durationGrammar)

			if durationGrammar.MatchString(text) {
				assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
			}

			return
		}

		require.True(t, durationGrammar.MatchString(text), "accepted text matches the grammar")
		assert.Equal(t, 1, got.MustFraction().Sign())
		assert.Equal(t, got, duration.MustParse(got.String()), "round trip")
		assert.Equal(t, got.String(), duration.MustParse(got.String()).String(), "the output is canonical")
		assert.True(t, got.MustFraction().IsReduced())
	})
}

// FuzzParsePosition covers O-16 and architecture 4.12 for ParsePosition.
func FuzzParsePosition(f *testing.F) {
	for _, s := range textSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			got duration.Position
			err error
		)

		_, panicked := capturePanic(func() { got, err = duration.ParsePosition(text) })
		require.False(t, panicked)

		if err != nil {
			assert.Equal(t, duration.Position{}, got)
			checkParseFailure(t, err, text, positionGrammar)

			if positionGrammar.MatchString(text) {
				assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
			}

			return
		}

		require.True(t, positionGrammar.MatchString(text))
		assert.GreaterOrEqual(t, got.MustFraction().Sign(), 0)
		assert.Equal(t, got, duration.MustParsePosition(got.String()))
		assert.True(t, got.MustFraction().IsReduced())
	})
}

// FuzzParseWritten covers O-16 and architecture 4.12 for ParseWritten: an
// accepted text is canonical.
func FuzzParseWritten(f *testing.F) {
	for _, s := range textSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			got duration.Written
			err error
		)

		_, panicked := capturePanic(func() { got, err = duration.ParseWritten(text) })
		require.False(t, panicked)

		if err != nil {
			assert.Equal(t, duration.Written{}, got)
			checkParseFailure(t, err, text, writtenGrammar)

			return
		}

		require.True(t, writtenGrammar.MatchString(text))
		assert.Equal(t, text, got.String(), "an accepted text is canonical")
		assert.Equal(t, got, duration.MustParseWritten(got.String()))
		assert.Equal(t, 0, ratOfDuration(got.MustDuration()).Cmp(writtenRat(got.MustBase(), got.MustDots(), got.MustChain())))
	})
}

// FuzzParseTied covers O-16 and architecture 4.12 for ParseTied.
func FuzzParseTied(f *testing.F) {
	for _, s := range textSeeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			got duration.Tied
			err error
		)

		_, panicked := capturePanic(func() { got, err = duration.ParseTied(text) })
		require.False(t, panicked)

		if err != nil {
			assert.Equal(t, duration.Tied{}, got)
			checkParseFailure(t, err, text, tiedGrammar)

			return
		}

		require.True(t, tiedGrammar.MatchString(text))
		assert.Equal(t, text, got.String())
		assert.Equal(t, got, duration.MustParseTied(got.String()))
		assert.GreaterOrEqual(t, got.MustLen(), 2)
	})
}

// baseSeeds are additional seeds of FuzzParseBase: the names and spellings
// of P-12 (task 28.2) and texts close to them that are E-10.
var baseSeeds = []string{
	"large", "Duplex Longa", "long", "longa", "QUADRUPLE WHOLE NOTE", "double-whole note", "demisemiquaver",
	"hundred twenty-eighth note", "one-hundred-twenty-eighth note", "256th note", "32nd note", "ſemibreve",
	"octuple", "8th note", "512th note", "1024th note", "hundred-twenty-eighth note", "two hundred fifty sixth note",
	" long", "large ", "16th",
}

// FuzzParseBase covers O-15, P-12, and architecture 4.12 for ParseBase: the
// result agrees with the model of the accepted strings.
func FuzzParseBase(f *testing.F) {
	for _, s := range slices.Concat(textSeeds, baseSeeds) {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			got duration.Base
			err error
		)

		_, panicked := capturePanic(func() { got, err = duration.ParseBase(text) })
		require.False(t, panicked)

		want, accepted := expectedParseBase(text)
		if accepted {
			require.NoError(t, err, "%q is a string of architecture 4.11, 4.12", text)
			assert.Equal(t, want, got)
		} else {
			assert.Error(t, err, "%q is not a string of architecture 4.11, 4.12", text)
		}

		if err != nil {
			assert.Equal(t, duration.Base(0), got)
			assertErr(t, err, duration.ErrInvalidText)
			checkErrorText(t, err, text, "fuzz")

			return
		}

		require.True(t, got.IsValid())
		assert.Equal(t, got, duration.MustParseBase(got.String()))
		assert.Equal(t, text, strings.TrimSpace(text), "no surrounding spaces are accepted")
	})
}

// fuzzChains are chains selected by the fuzz tests.
var fuzzChains = [][]tuplet.Tuplet{
	nil,
	{tuplet.Triplet()},
	{tp(6, 4)},
	{tuplet.Duplet()},
	{tuplet.Quintuplet(), tuplet.Triplet()},
	{tp(7, 8), tp(1, 1)},
}

// FuzzNotate covers P-3 and architecture 4.8: for a value of two small
// numbers, a longest base, D <= 3, and a chain of up to two tuplets, the
// result agrees with the model of the rule and has its properties.
func FuzzNotate(f *testing.F) {
	f.Add(uint16(5), uint16(16), uint8(2), uint8(0), uint8(0))
	f.Add(uint16(11), uint16(16), uint8(2), uint8(1), uint8(0))
	f.Add(uint16(5), uint16(12), uint8(2), uint8(0), uint8(1))
	f.Add(uint16(12), uint16(1), uint8(0), uint8(1), uint8(0))
	f.Add(uint16(1), uint16(3), uint8(2), uint8(3), uint8(0))
	f.Add(uint16(5), uint16(2048), uint8(2), uint8(1), uint8(0))

	f.Fuzz(func(t *testing.T, num, den uint16, longestIndex, maxDots, chainIndex uint8) {
		if num == 0 || den == 0 {
			t.Skip("values are positive")
		}

		v := rat(int64(num), int64(den))
		longest := baseTable[int(longestIndex)%len(baseTable)].base
		dots := int(maxDots % 4)
		chain := fuzzChains[int(chainIndex)%len(fuzzChains)]
		d := durRat(v)

		got, err := d.NotateWith(longest, dots, chain...)
		checkNotateProperties(t, d, longest, dots, chain, got, err)

		vPrime := new(big.Rat).Quo(v, chainRat(chain))
		m := greedyTie(vPrime, longest, dots)

		switch {
		case m.outcome == tieNotNotable:
			assertErr(t, err, duration.ErrNotNotatable)
		case m.count.Cmp(big.NewInt(duration.MaxParts)) > 0:
			assertErr(t, err, duration.ErrTooManyParts)
		default:
			require.NoError(t, err)
			assert.Equal(t, writtenOfSigns(m.parts, chain), got)
		}
	})
}

// FuzzToWritten covers ID-4 and ID-5: agreement with the enumeration of
// single signs.
func FuzzToWritten(f *testing.F) {
	f.Add(uint32(3), uint32(8), uint8(0))
	f.Add(uint32(1), uint32(3), uint8(1))
	f.Add(uint32(15), uint32(1), uint8(0))
	f.Add(uint32(3), uint32(2048), uint8(0))
	f.Add(uint32(2), uint32(15), uint8(4))

	f.Fuzz(func(t *testing.T, num, den uint32, chainIndex uint8) {
		if num == 0 || den == 0 {
			t.Skip("values are positive")
		}

		v := rat(int64(num), int64(den))
		chain := fuzzChains[int(chainIndex)%len(fuzzChains)]
		signs := singleSigns(new(big.Rat).Quo(v, chainRat(chain)))
		require.LessOrEqual(t, len(signs), 1)

		got, err := durRat(v).ToWritten(chain...)
		if len(signs) == 0 {
			assertErr(t, err, duration.ErrNotNotatable)

			return
		}

		require.NoError(t, err)
		assert.Equal(t, wr(signs[0].base, signs[0].dots, chain...), got)
	})
}

// FuzzQuantize covers P-8 and architecture 4.10: agreement with RatMultiple
// of common/rounding; E-8 for a duration exactly when the node is zero.
func FuzzQuantize(f *testing.F) {
	f.Add(uint32(7), uint32(24), uint16(1), uint16(8), uint8(9))
	f.Add(uint32(5), uint32(16), uint16(1), uint16(8), uint8(5))
	f.Add(uint32(1), uint32(32), uint16(1), uint16(8), uint8(9))
	f.Add(uint32(0), uint32(1), uint16(3), uint16(16), uint8(2))
	f.Add(uint32(1), uint32(3), uint16(1), uint16(12), uint8(0))

	f.Fuzz(func(t *testing.T, num, den uint32, stepNum, stepDen uint16, modeValue uint8) {
		if den == 0 || stepNum == 0 || stepDen == 0 {
			t.Skip("valid inputs only")
		}

		x := rat(int64(num), int64(den))
		step := frac(int(stepNum), int(stepDen))
		mode := rounding.Mode(modeValue)

		gotPos, err := posRat(x).Quantize(step, mode)
		if !mode.IsValid() {
			if mode == 0 {
				assertErr(t, err, duration.ErrInvalidArgument, rounding.ErrUnset)
			} else {
				assertErr(t, err, duration.ErrInvalidArgument, rounding.ErrUnknownMode)
			}

			return
		}

		_, want, oracleErr := mode.RatMultiple(x, ratOf(step))
		require.NoError(t, oracleErr)
		require.NoError(t, err)
		assert.Equal(t, posRat(want), gotPos)

		if x.Sign() == 0 {
			return
		}

		gotDur, err := durRat(x).Quantize(step, mode)
		if want.Sign() == 0 {
			assertErr(t, err, duration.ErrQuantizedToZero)

			return
		}

		require.NoError(t, err)
		assert.Equal(t, durRat(want), gotDur)
	})
}

// FuzzArithmetic covers O-2 and O-13: agreement with big.Rat arithmetic.
func FuzzArithmetic(f *testing.F) {
	f.Add(int64(1), int64(4), int64(1), int64(8))
	f.Add(int64(3), int64(8), int64(3), int64(8))
	f.Add(int64(1), int64(1), int64(3), int64(8))
	f.Add(int64(-1), int64(-4), int64(7), int64(1))

	f.Fuzz(func(t *testing.T, an, ad, bn, bd int64) {
		if ad == 0 || bd == 0 {
			t.Skip("denominators are nonzero")
		}

		ra, rb := big.NewRat(an, ad), big.NewRat(bn, bd)
		if ra.Sign() <= 0 || rb.Sign() <= 0 {
			_, err := duration.New(int(an), int(ad))
			if ra.Sign() <= 0 {
				assertErr(t, err, duration.ErrNotPositive)
			}

			return
		}

		a, b := duration.MustNew(int(an), int(ad)), duration.MustNew(int(bn), int(bd))
		assert.Equal(t, 0, ratOfDuration(a).Cmp(ra))

		assert.Equal(t, durRat(new(big.Rat).Add(ra, rb)), a.MustAdd(b))
		assert.Equal(t, ra.Cmp(rb), a.Compare(b))
		assert.True(t, a.MustRatio(b).EqualExact(fracRat(new(big.Rat).Quo(ra, rb))))

		diff, err := a.Sub(b)
		if exact := new(big.Rat).Sub(ra, rb); exact.Sign() > 0 {
			require.NoError(t, err)
			assert.Equal(t, durRat(exact), diff)
		} else {
			assertErr(t, err, duration.ErrNotPositive)
		}

		quotient := new(big.Rat).Quo(ra, rb)
		wantQuo := new(big.Int).Quo(quotient.Num(), quotient.Denom())
		quo, rem, err := a.QuoRem(b)

		if !wantQuo.IsInt64() || wantQuo.Int64() > int64(^uint(0)>>1) {
			assertErr(t, err, duration.ErrOutOfRange, fraction.ErrNotRepresentable)

			return
		}

		require.NoError(t, err)
		assert.Equal(t, wantQuo.Int64(), int64(quo))
		assert.True(t, rem.EqualExact(fracRat(new(big.Rat).Sub(ra, new(big.Rat).Mul(rb, ratInt(wantQuo))))))

		p := posRat(ra)
		assert.Equal(t, posRat(new(big.Rat).Add(ra, rb)), p.MustAdd(b))

		offset, err := posRat(rb).OffsetTo(p)
		if exact := new(big.Rat).Sub(ra, rb); exact.Sign() >= 0 {
			require.NoError(t, err)
			assert.Equal(t, posRat(exact), offset)
		} else {
			assertErr(t, err, duration.ErrNegativePosition)
		}
	})
}
