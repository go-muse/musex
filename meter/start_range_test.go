package meter_test

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
)

// This file holds the inputs of the E-10 case in which the size of a sum is
// within the range of common/fraction while the start of some unit of the sum
// is not (architecture 4.2, row ErrOutOfRange and order of checks; 4.4,
// NewSum; 4.12, Shorter and Longer; 4.15, order of parse failures; section 7,
// "Exact computation"). The existing КП-39 inputs exceed the range by the size
// itself, so they do not reach this case.
//
// TestUnitStartOutOfRangeInputs confirms the properties of the inputs by an
// exact computation with math/big that does not depend on the package. The
// cases themselves are rows of TestNewSum, TestParseErrors, TestShorter and
// TestLonger.

// sumPart is a simple signature n/d, one part of a sum.
type sumPart struct{ n, d int }

// partMeters returns the parts built by meter.MustNew.
func partMeters(parts []sumPart) []meter.Meter {
	out := make([]meter.Meter, len(parts))
	for i, p := range parts {
		out[i] = meter.MustNew(p.n, p.d)
	}

	return out
}

// sumOf returns the sum of the parts built by meter.MustNewSum.
func sumOf(parts []sumPart) meter.Meter { return meter.MustNewSum(partMeters(parts)...) }

// sumText returns the text of the sum of the parts by the grammar of
// architecture 4.15, such as "1/3+1/6".
func sumText(parts []sumPart) string {
	texts := make([]string, len(parts))
	for i, p := range parts {
		texts[i] = strconv.Itoa(p.n) + "/" + strconv.Itoa(p.d)
	}

	return strings.Join(texts, "+")
}

// shifted returns c·2^shift + 1 for each c.
func shifted(shift int, cs ...int) []int {
	out := make([]int, len(cs))
	for i, c := range cs {
		out[i] = c<<shift + 1
	}

	return out
}

// pairedParts returns the parts 1/(k·p) for each p in order, then the parts
// 1/(k·p·(p−1)) for each p in order. Since 1/p + 1/(p·(p−1)) = 1/(p−1), the
// whole sum is Σ 1/(k·(p−1)): the second half brings the denominator of the
// size back from the product of the p to a short one, while the starts of the
// units between the halves keep long denominators. It panics if a
// denominator exceeds math.MaxInt, which is a test error.
func pairedParts(k int, ps []int) []sumPart {
	out := make([]sumPart, 0, 2*len(ps))
	for _, p := range ps {
		out = append(out, sumPart{1, intOf(new(big.Int).Mul(big.NewInt(int64(k)), big.NewInt(int64(p))))})
	}

	for _, p := range ps {
		d := new(big.Int).Mul(big.NewInt(int64(k)), big.NewInt(int64(p)))
		d.Mul(d, big.NewInt(int64(p-1)))
		out = append(out, sumPart{1, intOf(d)})
	}

	return out
}

// intOf returns x as an int; it panics if x exceeds math.MaxInt.
func intOf(x *big.Int) int {
	if !x.IsInt64() || x.Int64() > math.MaxInt {
		panic("intOf: " + x.String() + " exceeds math.MaxInt")
	}

	return int(x.Int64())
}

// unitStartInput is the input of NewSum and Parse: nine parts 1/p and nine
// parts 1/(p·(p−1)) for the primes p = c·2^24 + 1 near 2^31. The start of the
// tenth unit, Σ 1/p, has the product of the nine p (about 2^279) as its
// denominator; the size Σ 1/(p−1) is in range. 18 units.
func unitStartInput() []sumPart {
	return pairedParts(1, shifted(24, 126, 127, 136, 148, 151, 157, 162, 171, 172))
}

// shorterStartInput is the input of Shorter: nine parts 1/p and nine parts
// 1/(p·(p−1)) for the primes p = c·2^20 + 1 near 2^28, whose product lies
// between 2^254 and 2^255. The size and the starts of all units are in
// range. Shorter doubles every denominator: the start of the tenth unit of
// the result, Σ 1/(2p), has a denominator of twice that product, outside the
// range, while the size of the result is in range.
func shorterStartInput() []sumPart {
	return pairedParts(1, shifted(20, 202, 211, 220, 235, 238, 391, 438, 447, 471))
}

// longerStartInput is the input of Longer, with even denominators only:
// 1000/2 + 1/254 + 1/(2q₁) + … + 1/(2q₄) + 1/(2p) + 1/(2p·(p−1)) for the primes
// q near 2^52 and p = 127·2^24 + 1. The size and the starts of all units are
// in range. Longer halves every denominator, which doubles every start: the
// start of the last unit of the result, 1000 + 1/127 + Σ 1/q + 1/p, has a
// numerator above 2^255 − 1. In the size of the result, 1/p + 1/(p·(p−1))
// gives 1/(p−1), and p − 1 = 127·2^24 shares 127 with 254, so the size is in
// range. 1007 units.
func longerStartInput() []sumPart {
	parts := []sumPart{{1000, 2}, {1, 2 * 127}}
	for _, q := range []int{4170701791206407, 4170701791206493, 4170701791206521, 4170701791206611} {
		parts = append(parts, sumPart{1, 2 * q})
	}

	return append(parts, pairedParts(2, shifted(24, 127))...)
}

// doubled returns the parts with every denominator doubled, as Shorter does
// (architecture 4.12). It panics if a doubled denominator exceeds
// math.MaxInt, which is a test error.
func doubled(parts []sumPart) []sumPart {
	out := make([]sumPart, len(parts))
	for i, p := range parts {
		if p.d > math.MaxInt/2 {
			panic("doubled: denominator " + strconv.Itoa(p.d) + " exceeds math.MaxInt/2")
		}

		out[i] = sumPart{p.n, 2 * p.d}
	}

	return out
}

// halved returns the parts with every denominator halved, as Longer does
// (architecture 4.12). It panics on an odd denominator, which is a test
// error.
func halved(parts []sumPart) []sumPart {
	out := make([]sumPart, len(parts))
	for i, p := range parts {
		if p.d%2 != 0 {
			panic("halved: odd denominator " + strconv.Itoa(p.d))
		}

		out[i] = sumPart{p.n, p.d / 2}
	}

	return out
}

// exactSum returns the exact size of the sum of the parts and the exact
// start of each of its units, without the package: the units of a part n/d
// are n units 1/d, and the units of the parts follow each other from 0
// (ИР-1, ИР-2).
func exactSum(parts []sumPart) (*big.Rat, []*big.Rat) {
	size := new(big.Rat)

	var starts []*big.Rat

	for _, p := range parts {
		for j := range p.n {
			step := new(big.Rat).SetFrac(big.NewInt(int64(j)), big.NewInt(int64(p.d)))
			starts = append(starts, new(big.Rat).Add(size, step))
		}

		size.Add(size, new(big.Rat).SetFrac(big.NewInt(int64(p.n)), big.NewInt(int64(p.d))))
	}

	return size, starts
}

// ratInRange reports whether the reduced numerator and denominator of r are
// within the range of common/fraction, ±(2^255 − 1). big.Rat is always
// reduced.
func ratInRange(r *big.Rat) bool {
	return r.Num().CmpAbs(maxComponent()) <= 0 && r.Denom().Cmp(maxComponent()) <= 0
}

// startsOutOfRange returns the numbers, from 1, of the units whose starts
// are outside the range of common/fraction.
func startsOutOfRange(starts []*big.Rat) []int {
	var out []int

	for i, s := range starts {
		if !ratInRange(s) {
			out = append(out, i+1)
		}
	}

	return out
}

// TestUnitStartOutOfRangeInputs confirms by exact computation the properties
// of the inputs of the E-10 case "start of a unit out of range, size in
// range" (task 28a.2), so that each operation reaches that check and no
// earlier one: the sum to be built has at most MaxUnits units, its size is in
// range and the start of some unit is not; the input of Shorter and Longer is
// a valid sum whose size and starts are all in range, with denominators that
// stay at most math.MaxInt when doubled (Shorter) or are all even (Longer).
func TestUnitStartOutOfRangeInputs(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name string
		in   []sumPart
		out  func([]sumPart) []sumPart // the parts of the sum the operation builds
		// validIn requires the input itself to be a sum in range.
		validIn bool
	}{
		{"NewSum and Parse", unitStartInput(), func(p []sumPart) []sumPart { return p }, false},
		{"Shorter", shorterStartInput(), doubled, true},
		{"Longer", longerStartInput(), halved, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			if c.validIn {
				size, starts := exactSum(c.in)
				assert.LessOrEqual(t, len(starts), meter.MaxUnits, "units of the input")
				assert.True(t, ratInRange(size), "size of the input in range: %s", size.RatString())
				assert.Empty(t, startsOutOfRange(starts), "units of the input whose starts are out of range")
			}

			out := c.out(c.in)
			for _, p := range out {
				assert.GreaterOrEqual(t, p.d, 1, "denominator of %v", p)
			}

			size, starts := exactSum(out)
			assert.LessOrEqual(t, len(starts), meter.MaxUnits, "units of the result")
			assert.True(t, ratInRange(size), "size of the result in range: %s", size.RatString())
			assert.NotEmpty(t, startsOutOfRange(starts), "a unit of the result whose start is out of range")
		})
	}
}
