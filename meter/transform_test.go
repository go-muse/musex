package meter_test

import (
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toFunc is one of the five methods of changing the record (О-13).
type toFunc struct {
	name string
	f    func(m meter.Meter) (meter.Meter, error)
}

var (
	toNumeric  = toFunc{"ToNumeric", meter.Meter.ToNumeric}
	toSymbol   = toFunc{"ToSymbol", meter.Meter.ToSymbol}
	toSimple   = toFunc{"ToSimple", meter.Meter.ToSimple}
	toAdditive = toFunc{"ToAdditive", meter.Meter.ToAdditive}
	toSum      = toFunc{"ToSum", meter.Meter.ToSum}
	toAll      = []toFunc{toNumeric, toSymbol, toSimple, toAdditive, toSum}
)

// TestChangeRecord covers О-13, ИР-13, Т-10, the table 2.4.1, КП-20, КП-20а,
// КП-20б, КП-20в and architecture 4.12: the result has the same structure.
func TestChangeRecord(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name     string
		to       toFunc
		in, want string
	}{
		// КП-20: in numbers.
		{"КП-20 C", toNumeric, "C", "4/4"},
		{"КП-20 C|", toNumeric, "C|", "2/2"},
		{"КП-20 C[3+1]", toNumeric, "C[3+1]", "4/4[3+1]"},
		{"КП-20 3/4 the same", toNumeric, "3/4", "3/4"},
		{"(2+3)/8 the same", toNumeric, "(2+3)/8", "(2+3)/8"},
		{"2/4+3/8 the same", toNumeric, "2/4+3/8", "2/4+3/8"},
		{"5/8[2+3] the same", toNumeric, "5/8[2+3]", "5/8[2+3]"},
		{"КП-20в 5/8+2/8 the same", toNumeric, "5/8+2/8", "5/8+2/8"},
		{"5/8 the same", toNumeric, "5/8", "5/8"},

		// КП-20: as a symbol.
		{"КП-20 4/4", toSymbol, "4/4", "C"},
		{"КП-20 2/2", toSymbol, "2/2", "C|"},
		{"КП-20 4/4[3+1]", toSymbol, "4/4[3+1]", "C[3+1]"},
		{"КП-20 C the same", toSymbol, "C", "C"},
		{"C| the same", toSymbol, "C|", "C|"},
		{"2/2[1+1]", toSymbol, "2/2[1+1]", "C|[1+1]"},

		// КП-20а: as a simple signature.
		{"КП-20а (2+3)/8", toSimple, "(2+3)/8", "5/8[2+3]"},
		{"КП-20а (3+3)/8", toSimple, "(3+3)/8", "6/8"},
		{"КП-20а 2/8+3/8", toSimple, "2/8+3/8", "5/8[2+3]"},
		{"КП-20а 4/8+3/8", toSimple, "4/8+3/8", "7/8[2+2+3]"},
		{"КП-20а 5/8+2/8", toSimple, "5/8+2/8", "7/8"},
		{"КП-20а C", toSimple, "C", "4/4"},
		{"КП-20а 3/4 the same", toSimple, "3/4", "3/4"},
		{"C|", toSimple, "C|", "2/2"},
		{"C[3+1]", toSimple, "C[3+1]", "4/4[3+1]"},
		{"КП-20в 5/8 the same", toSimple, "5/8", "5/8"},
		{"5/8[2+3] the same", toSimple, "5/8[2+3]", "5/8[2+3]"},
		{"(2+2)/4", toSimple, "(2+2)/4", "4/4"},
		{"2/4+2/4", toSimple, "2/4+2/4", "4/4"},
		{"(4+4)/8", toSimple, "(4+4)/8", "8/8[4+4]"},
		{"(2+3)/8+2/8", toSimple, "(2+3)/8+2/8", "7/8[2+3+2]"},
		{"3/8+3/8", toSimple, "3/8+3/8", "6/8"},

		// КП-20б: as a compound numerator.
		{"КП-20б 5/8[2+3]", toAdditive, "5/8[2+3]", "(2+3)/8"},
		{"КП-20б 6/8", toAdditive, "6/8", "(3+3)/8"},
		{"КП-20б C", toAdditive, "C", "(2+2)/4"},
		{"КП-20б 2/8+3/8", toAdditive, "2/8+3/8", "(2+3)/8"},
		{"КП-20б (2+3)/8 the same", toAdditive, "(2+3)/8", "(2+3)/8"},
		{"КП-20в (4+4)/8 the same", toAdditive, "(4+4)/8", "(4+4)/8"},
		{"4/4", toAdditive, "4/4", "(2+2)/4"},
		{"C[3+1]", toAdditive, "C[3+1]", "(3+1)/4"},
		{"(2+3)/8+2/8", toAdditive, "(2+3)/8+2/8", "(2+3+2)/8"},
		{"12/8", toAdditive, "12/8", "(3+3+3+3)/8"},

		// КП-20б: as a sum.
		{"КП-20б 6/8", toSum, "6/8", "3/8+3/8"},
		{"КП-20б 4/4", toSum, "4/4", "2/4+2/4"},
		{"КП-20б (2+3)/8", toSum, "(2+3)/8", "2/8+3/8"},
		{"КП-20б C", toSum, "C", "2/4+2/4"},
		{"КП-20б 2/4+3/8 the same", toSum, "2/4+3/8", "2/4+3/8"},
		{"КП-20в 5/8+2/8 the same", toSum, "5/8+2/8", "5/8+2/8"},
		{"C[3+1]", toSum, "C[3+1]", "3/4+1/4"},
		{"(3+2+3)/8", toSum, "(3+2+3)/8", "3/8+2/8+3/8"},
		{"12/8", toSum, "12/8", "3/8+3/8+3/8+3/8"},
		{"6/8[2+2+2]", toSum, "6/8[2+2+2]", "2/8+2/8+2/8"},
		{"5/8[2+3]", toSum, "5/8[2+3]", "2/8+3/8"},
	} {
		t.Run(c.to.name+" "+c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			in := mt(c.in)

			got, err := c.to.f(in)
			require.NoError(t, err)

			want := mt(c.want)
			assert.Equal(t, want, got)
			assert.True(t, got.Equal(want))
			assert.True(t, noErr(in.EqualStructure(got)), "ИР-13: the structure is kept")
		})
	}
}

// TestChangeRecordErrors covers КП-30, КП-31, КП-32, the conditions under the
// table 2.4.1 and the order of checks of the To… methods (architecture 4.2).
func TestChangeRecordErrors(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name string
		to   toFunc
		in   string
		want error
	}{
		// ToSymbol: E-5.
		{"КП-32 3/4", toSymbol, "3/4", meter.ErrUndefined},
		{"КП-32 (2+2)/4", toSymbol, "(2+2)/4", meter.ErrUndefined},
		{"КП-32 2/4+2/4", toSymbol, "2/4+2/4", meter.ErrUndefined},
		{"(2+3)/8", toSymbol, "(2+3)/8", meter.ErrUndefined},
		{"4/2", toSymbol, "4/2", meter.ErrUndefined},
		{"8/8", toSymbol, "8/8", meter.ErrUndefined},
		{"2/4+3/8", toSymbol, "2/4+3/8", meter.ErrUndefined},

		// ToSimple.
		{"КП-31 2/4+3/8", toSimple, "2/4+3/8", meter.ErrNoCommonDenominator},
		{"КП-32 5/8+1/8", toSimple, "5/8+1/8", meter.ErrUndefined},
		{"КП-32 7/8+2/8", toSimple, "7/8+2/8", meter.ErrUndefined},
		{"5/8+7/8", toSimple, "5/8+7/8", meter.ErrUndefined},
		{"order: E-4 before E-5", toSimple, "5/8+1/4", meter.ErrNoCommonDenominator},
		{"different denominators without grouping", toSimple, "5/8+2/4", meter.ErrNoCommonDenominator},

		// ToAdditive.
		{"КП-31 2/4+3/8", toAdditive, "2/4+3/8", meter.ErrNoCommonDenominator},
		{"КП-30 5/8", toAdditive, "5/8", meter.ErrGroupingRequired},
		{"КП-30 7/8", toAdditive, "7/8", meter.ErrGroupingRequired},
		{"КП-30 8/8", toAdditive, "8/8", meter.ErrGroupingRequired},
		{"КП-30 5/8+2/8", toAdditive, "5/8+2/8", meter.ErrGroupingRequired},
		{"КП-32 3/4 one group", toAdditive, "3/4", meter.ErrUndefined},
		{"КП-32 C| one group", toAdditive, "C|", meter.ErrUndefined},
		{"5/8[5] one group", toAdditive, "5/8[5]", meter.ErrUndefined},
		{"1/4 one group", toAdditive, "1/4", meter.ErrUndefined},
		{"order: E-4 before E-3", toAdditive, "5/8+2/4", meter.ErrNoCommonDenominator},

		// ToSum.
		{"КП-30 5/8", toSum, "5/8", meter.ErrGroupingRequired},
		{"КП-30 7/8", toSum, "7/8", meter.ErrGroupingRequired},
		{"КП-30 8/8", toSum, "8/8", meter.ErrGroupingRequired},
		{"КП-32 3/4 one group", toSum, "3/4", meter.ErrUndefined},
		{"КП-32 C| one group", toSum, "C|", meter.ErrUndefined},
		{"КП-32 (4+4)/8 group of four", toSum, "(4+4)/8", meter.ErrUndefined},
		{"КП-32 5/8[5]", toSum, "5/8[5]", meter.ErrUndefined},
		{"5/8[1+4] group of four", toSum, "5/8[1+4]", meter.ErrUndefined},
		{"КП-30 13/8", toSum, "13/8", meter.ErrGroupingRequired},
	} {
		t.Run(c.to.name+" "+c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := c.to.f(mt(c.in))
			assertErr(t, err, c.want)
			assert.Equal(t, meter.Meter{}, got)
		})
	}

	for _, to := range toAll {
		t.Run(to.name+" unset", func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := to.f(meter.Meter{})
			assertErr(t, err, meter.ErrUnset)
			assert.Equal(t, meter.Meter{}, got)
		})
	}
}

// TestChangeRecordProperties covers ИР-13 and architecture 4.12 on the
// corpus: every successful change keeps the structure, applying the same
// method to the result returns it unchanged, and a meter already in the
// requested form is returned as it is.
func TestChangeRecordProperties(t *testing.T) {
	t.Parallel()
	defer guard(t)

	inForm := map[string]func(meter.Meter) bool{
		"ToNumeric":  func(m meter.Meter) bool { return noErr(m.Form()) != meter.FormSymbol },
		"ToSymbol":   func(m meter.Meter) bool { return noErr(m.Form()) == meter.FormSymbol },
		"ToSimple":   func(m meter.Meter) bool { return noErr(m.Form()) == meter.FormSimple },
		"ToAdditive": func(m meter.Meter) bool { return noErr(m.Form()) == meter.FormAdditive },
		"ToSum":      func(m meter.Meter) bool { return noErr(m.Form()) == meter.FormSum },
	}

	for _, m := range corpus() {
		t.Run(m.String(), func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			for _, to := range toAll {
				got, err := to.f(m)

				if inForm[to.name](m) {
					require.NoError(t, err, to.name)
					assert.Equal(t, m, got, "%s: already in the form", to.name)

					continue
				}

				if err != nil {
					continue
				}

				assert.True(t, noErr(m.EqualStructure(got)), "%s keeps the structure", to.name)
				assert.Equal(t, got, noErr(to.f(got)), "%s is idempotent", to.name)
			}
		})
	}
}

// TestShorter covers О-14, П-3, Т-10, КП-21, КП-52 and the order of checks of
// Shorter.
func TestShorter(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct{ name, in, want string }{
		{"КП-21 3/4", "3/4", "3/8"},
		{"КП-21 (3+2+3)/8", "(3+2+3)/8", "(3+2+3)/16"},
		{"КП-21 C", "C", "4/8"},
		{"КП-21 2/4+3/8", "2/4+3/8", "2/8+3/16"},
		{"5/8[2+3]", "5/8[2+3]", "5/16[2+3]"},
		{"C[3+1]", "C[3+1]", "4/8[3+1]"},
		{"C|", "C|", "2/4"},
		{"4/3", "4/3", "4/6"},
		{"5/8", "5/8", "5/16"},
		{"denominator 2^61", "1/2305843009213693952", "1/4611686018427387904"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := mt(c.in).Shorter()
			require.NoError(t, err)
			assert.Equal(t, mt(c.want), got)
		})
	}

	for _, c := range []struct {
		name   string
		m      meter.Meter
		want   error
		causes []error
	}{
		{"КП-52 denominator MaxInt", meter.MustNew(1, math.MaxInt), meter.ErrOutOfRange, nil},
		{"denominator 2^62", meter.MustNew(3, 1<<62), meter.ErrOutOfRange, nil},
		{"denominator above MaxInt in a sum", meter.MustNewSum(meter.MustNew(1, 1<<62), meter.MustNew(1, 3)),
			meter.ErrOutOfRange, nil},
		{"order: denominator before range in a sum", shorterBothViolations(), meter.ErrOutOfRange, nil},
		{"sum out of range", sumNearRange(), meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"unset", meter.Meter{}, meter.ErrUnset, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := c.m.Shorter()
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, meter.Meter{}, got)
		})
	}
}

// sumNearRange returns a sum of five parts 1/p with odd pairwise coprime
// denominators whose product P lies in [2^254, 2^255): the size and the
// starts of the units are in range, while the size of Shorter has the
// denominator 2P, outside the range of common/fraction.
func sumNearRange() meter.Meter {
	return meter.MustNewSum(unitParts(nearRangeDenominators(kp39Denominators()[:4]))...)
}

// shorterBothViolations is a sum like sumNearRange with a part 1/MaxInt:
// Shorter would double that denominator above math.MaxInt and also give a
// size outside the range of common/fraction. The denominator is checked
// first, so E-10 wraps nothing (architecture 4.2, 4.12).
func shorterBothViolations() meter.Meter {
	return meter.MustNewSum(unitParts(nearRangeDenominators([]int{math.MaxInt, mersenne(61), mersenne(59), mersenne(53)}))...)
}

// nearRangeDenominators appends to odd pairwise coprime denominators the
// smallest odd denominator coprime with them that brings their product into
// [2^254, 2^255).
func nearRangeDenominators(dens []int) []int {
	product := big.NewInt(1)
	for _, d := range dens {
		product.Mul(product, big.NewInt(int64(d)))
	}

	low := new(big.Int).Div(new(big.Int).Add(pow2(254), new(big.Int).Sub(product, big.NewInt(1))), product)
	d := coprimeOdd(low, product)

	if new(big.Int).Mul(product, d).Cmp(maxComponent()) > 0 {
		panic("nearRangeDenominators: no denominator found")
	}

	return append(slices.Clone(dens), int(d.Int64()))
}

// coprimeOdd returns the smallest odd number not less than from that is
// coprime with product.
func coprimeOdd(from, product *big.Int) *big.Int {
	d := new(big.Int).Set(from)
	if d.Bit(0) == 0 {
		d.Add(d, big.NewInt(1))
	}

	for new(big.Int).GCD(nil, nil, d, product).Cmp(big.NewInt(1)) != 0 {
		d.Add(d, big.NewInt(2))
	}

	return d
}

// longerOutOfRange returns 3/2 + 1/(2q₁) + … + 1/(2q₅) with odd pairwise
// coprime q: q₁…q₄ are 2^61−1, 2^59−1, 2^53−1, 2^47−1, and q₅ is the
// smallest odd number coprime with them for which the size of Longer,
// 3 + 1/q₁ + … + 1/q₅, has a numerator above 2^255−1. The numerator of that
// sum is even, so the size of the input, half of it, and the starts of its
// units are in range (architecture 4.12: Longer, E-10).
func longerOutOfRange() meter.Meter {
	qs := []*big.Int{
		big.NewInt(int64(mersenne(61))), big.NewInt(int64(mersenne(59))),
		big.NewInt(int64(mersenne(53))), big.NewInt(int64(mersenne(47))),
	}

	q4 := big.NewInt(1)
	for _, q := range qs {
		q4.Mul(q4, q)
	}

	inRange := func(q5 *big.Int) bool {
		longer := big.NewRat(3, 1)
		for _, q := range append(slices.Clone(qs), q5) {
			longer.Add(longer, new(big.Rat).SetFrac(big.NewInt(1), q))
		}

		return longer.Num().CmpAbs(maxComponent()) <= 0 && longer.Denom().Cmp(maxComponent()) <= 0
	}

	q5 := coprimeOdd(new(big.Int).Div(maxComponent(), new(big.Int).Mul(big.NewInt(3), q4)), q4)
	for inRange(q5) {
		q5 = coprimeOdd(new(big.Int).Add(q5, big.NewInt(2)), q4)
	}

	parts := []meter.Meter{meter.MustNew(3, 2)}
	for _, q := range append(qs, q5) {
		parts = append(parts, meter.MustNew(1, int(2*q.Int64())))
	}

	return meter.MustNewSum(parts...)
}

// TestLonger covers О-14, Т-10, КП-21, КП-32.
func TestLonger(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct{ name, in, want string }{
		{"КП-21 3/4", "3/4", "3/2"},
		{"КП-21 5/8[2+3]", "5/8[2+3]", "5/4[2+3]"},
		{"6/8", "6/8", "6/4"},
		{"C", "C", "4/2"},
		{"C|", "C|", "2/1"},
		{"C|[1+1]", "C|[1+1]", "2/1[1+1]"},
		{"(3+2+3)/16", "(3+2+3)/16", "(3+2+3)/8"},
		{"2/8+3/16", "2/8+3/16", "2/4+3/8"},
		{"1/2", "1/2", "1/1"},
		{"4/6", "4/6", "4/3"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := mt(c.in).Longer()
			require.NoError(t, err)
			assert.Equal(t, mt(c.want), got)
		})
	}

	for _, c := range []struct {
		name   string
		m      meter.Meter
		want   error
		causes []error
	}{
		{"КП-32 3/1", mt("3/1"), meter.ErrUndefined, nil},
		{"КП-32 4/3", mt("4/3"), meter.ErrUndefined, nil},
		{"odd denominator in a part", mt("2/4+2/3"), meter.ErrUndefined, nil},
		{"unset", meter.Meter{}, meter.ErrUnset, nil},
		{"size out of range", longerOutOfRange(), meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := c.m.Longer()
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, meter.Meter{}, got)
		})
	}
}

// TestShorterLonger covers architecture 4.12 on the corpus: Shorter then
// Longer gives ToNumeric of the input.
func TestShorterLonger(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, m := range corpus() {
		short, err := m.Shorter()
		require.NoError(t, err, "%v", m)
		assert.Equal(t, noErr(m.ToNumeric()), noErr(short.Longer()), "%v", m)
		assert.Equal(t, describe(m.Grouping()), describe(short.Grouping()), "%v: the grouping is kept", m)
		assert.Equal(t, describe(m.Addends()), describe(short.Addends()), "%v: the addends are kept", m)
	}
}

// TestWithDenominator covers О-14, Т-10, КП-21, КП-31 and the order of checks
// of WithDenominator.
func TestWithDenominator(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name, in string
		d        int
		want     string
	}{
		{"КП-21 6/8 to 4", "6/8", 4, "6/4"},
		{"C to 4", "C", 4, "4/4"},
		{"C to 8", "C", 8, "4/8"},
		{"C[3+1] to 8", "C[3+1]", 8, "4/8[3+1]"},
		{"(3+2+3)/8 to 16", "(3+2+3)/8", 16, "(3+2+3)/16"},
		{"2/8+3/8 to 4", "2/8+3/8", 4, "2/4+3/4"},
		{"5/8[2+3] to 3", "5/8[2+3]", 3, "5/3[2+3]"},
		{"6/8 to 8", "6/8", 8, "6/8"},
		{"6/8 to MaxInt", "6/8", math.MaxInt, "6/" + maxIntText()},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := mt(c.in).WithDenominator(c.d)
			require.NoError(t, err)
			assert.Equal(t, mt(c.want), got)
		})
	}

	for _, c := range []struct {
		name string
		m    meter.Meter
		d    int
		want error
	}{
		{"denominator 0", mt("6/8"), 0, meter.ErrInvalidArgument},
		{"denominator -1", mt("6/8"), -1, meter.ErrInvalidArgument},
		{"КП-31 2/4+3/8 to 16", mt("2/4+3/8"), 16, meter.ErrNoCommonDenominator},
		{"order: E-2 before E-4", mt("2/4+3/8"), 0, meter.ErrInvalidArgument},
		{"unset", meter.Meter{}, 4, meter.ErrUnset},
		{"order: E-1 before E-2", meter.Meter{}, 0, meter.ErrUnset},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := c.m.WithDenominator(c.d)
			assertErr(t, err, c.want)
			assert.Equal(t, meter.Meter{}, got)
		})
	}
}

// TestFromDuration covers О-15, Т-10, КП-22, КП-36, КП-38 and the order of
// checks of FromDuration.
func TestFromDuration(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		size string
		d    int
		want string
	}{
		{"3/4", 4, "3/4"},
		{"3/4", 8, "6/8"},
		{"3/4", 16, "12/16"},
		{"1", 3, "3/3"},
		{"5/8", 8, "5/8"},
		{"128", 8, "1024/8"},
		{"1/2", 2, "1/2"},
	} {
		t.Run("КП-22 "+c.size+" in "+itoa(c.d), func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := meter.FromDuration(durT(c.size), c.d)
			require.NoError(t, err)
			assert.Equal(t, mt(c.want), got)
		})
	}

	huge := durRat(new(big.Rat).SetInt(pow2(70)))

	for _, c := range []struct {
		name   string
		size   duration.Duration
		d      int
		want   error
		causes []error
	}{
		{"КП-36 3/4 in 2", durT("3/4"), 2, meter.ErrInvalidArgument, nil},
		{"КП-36 1/3 in 4", durT("1/3"), 4, meter.ErrInvalidArgument, nil},
		{"КП-36 denominator 0", durT("3/4"), 0, meter.ErrInvalidArgument, nil},
		{"denominator -4", durT("3/4"), -4, meter.ErrInvalidArgument, nil},
		{"КП-36 2 in 1024", durT("2"), 1024, meter.ErrTooMany, nil},
		{"more units than MaxInt", huge, 8, meter.ErrTooMany, nil},
		{"КП-38 unset size", duration.Duration{}, 4, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"order: unset size before denominator", duration.Duration{}, 0, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"order: not a multiple before E-8", durT("4097/4"), 2, meter.ErrInvalidArgument, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := meter.FromDuration(c.size, c.d)
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, meter.Meter{}, got)
		})
	}
}
