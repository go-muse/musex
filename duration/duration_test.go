package duration_test

import (
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew covers O-1, KP-23, and architecture 4.4: signs of components count
// by value; E-2 with fraction.ErrZeroDenominator for den == 0; E-3 for <= 0.
func TestNew(t *testing.T) {
	t.Parallel()

	ok := []struct {
		name     string
		num, den int
		want     *big.Rat
	}{
		{"quarter", 1, 4, rat(1, 4)},
		{"both negative", -1, -4, rat(1, 4)},
		{"unreduced", 6, 16, rat(3, 8)},
		{"whole", 1, 1, rat(1, 1)},
		{"long", 2049, 1, rat(2049, 1)},
		{"max int", math.MaxInt, 1, new(big.Rat).SetInt64(math.MaxInt)},
		{"min fraction", 1, math.MaxInt, new(big.Rat).SetFrac64(1, math.MaxInt)},
	}

	for _, c := range ok {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			d, err := duration.New(c.num, c.den)
			require.NoError(t, err)
			assert.True(t, d.IsSet())
			assert.Equal(t, 0, ratOfDuration(d).Cmp(c.want))
		})
	}

	failing := []struct {
		name     string
		num, den int
		want     error
		causes   []error
	}{
		{"zero", 0, 1, duration.ErrNotPositive, nil},
		{"negative", -1, 4, duration.ErrNotPositive, nil},
		{"negative denominator", 1, -4, duration.ErrNotPositive, nil},
		{"zero denominator", 1, 0, duration.ErrInvalidArgument, []error{fraction.ErrZeroDenominator}},
		{"zero over zero", 0, 0, duration.ErrInvalidArgument, []error{fraction.ErrZeroDenominator}},
		{"min int", math.MinInt, 1, duration.ErrNotPositive, nil},
	}

	for _, c := range failing {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			d, err := duration.New(c.num, c.den)
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, duration.Duration{}, d)
		})
	}
}

// TestNewCanonical covers KR-D4 and architecture 4.4: a value is stored
// reduced, so equal values are == and Fraction returns the reduced pair.
func TestNewCanonical(t *testing.T) {
	t.Parallel()

	a := dur(6, 16)
	b := dur(3, 8)
	c := duration.MustFromFraction(frac(-6, -16))

	assert.True(t, a == b)
	assert.True(t, a == c)
	assert.Equal(t, "3/8", a.String())

	f, err := a.Fraction()
	require.NoError(t, err)
	assert.True(t, f.EqualExact(frac(3, 8)), "pair %s", f)

	assert.True(t, dur(4, 2).MustFraction().EqualExact(frac(2, 1)))
	assert.True(t, dur(-3, -1).MustFraction().EqualExact(frac(3, 1)))
}

// TestFromFraction covers O-1 and architecture 4.2: an unset fraction is E-2
// without fraction errors; <= 0 is E-3.
func TestFromFraction(t *testing.T) {
	t.Parallel()

	d, err := duration.FromFraction(frac(3, 8))
	require.NoError(t, err)
	assert.Equal(t, dur(3, 8), d)

	maxValue := fraction.MustFromBigInt(maxComponent(), big.NewInt(1))
	d, err = duration.FromFraction(maxValue)
	require.NoError(t, err, "the largest value in range")
	assert.Equal(t, 0, ratOfDuration(d).Cmp(ratInt(maxComponent())))

	smallest := fraction.MustFromBigInt(big.NewInt(1), maxComponent())
	d, err = duration.FromFraction(smallest)
	require.NoError(t, err)
	assert.Equal(t, 0, ratOfDuration(d).Cmp(new(big.Rat).SetFrac(big.NewInt(1), maxComponent())))

	for _, c := range []struct {
		name string
		f    fraction.Fraction
		want error
	}{
		{"unset", fraction.Fraction{}, duration.ErrInvalidArgument},
		{"zero", fraction.Zero(), duration.ErrNotPositive},
		{"zero over three", frac(0, 3), duration.ErrNotPositive},
		{"negative", frac(-1, 4), duration.ErrNotPositive},
		{"negative denominator", frac(1, -4), duration.ErrNotPositive},
	} {
		d, err := duration.FromFraction(c.f)
		assertErr(t, err, c.want)
		assert.Equal(t, duration.Duration{}, d, c.name)
	}
}

// arithmeticValues are durations for model checks of arithmetic.
func arithmeticValues() []duration.Duration {
	return []duration.Duration{
		dur(1, 1), dur(1, 2), dur(1, 4), dur(3, 8), dur(1, 3), dur(2, 15), dur(5, 16),
		dur(7, 2), dur(1024, 1), dur(1, 1024), dur(22, 7), dur(9, 1),
	}
}

// TestDurationArithmeticModel covers O-2, T-1, and ID-7: Add, Sub, MulInt,
// DivInt, Mul, Div, Ratio, QuoRem, and Compare agree with exact big.Rat
// arithmetic; results are reduced.
func TestDurationArithmeticModel(t *testing.T) {
	t.Parallel()

	values := arithmeticValues()

	for _, a := range values {
		for _, b := range values {
			ra, rb := ratOfDuration(a), ratOfDuration(b)

			sum, err := a.Add(b)
			require.NoError(t, err)
			assert.Equal(t, durRat(new(big.Rat).Add(ra, rb)), sum, "%s + %s", a, b)

			diff, err := a.Sub(b)
			if exact := new(big.Rat).Sub(ra, rb); exact.Sign() > 0 {
				require.NoError(t, err)
				assert.Equal(t, durRat(exact), diff, "%s - %s", a, b)
			} else {
				assertErr(t, err, duration.ErrNotPositive)
				assert.Equal(t, duration.Duration{}, diff)
			}

			ratio, err := a.Ratio(b)
			require.NoError(t, err)
			assert.True(t, ratio.EqualExact(fracRat(new(big.Rat).Quo(ra, rb))), "%s : %s = %s", a, b, ratio)

			quo, rem, err := a.QuoRem(b)
			require.NoError(t, err)

			wantQuo := new(big.Int).Quo(new(big.Rat).Quo(ra, rb).Num(), new(big.Rat).Quo(ra, rb).Denom())
			wantRem := new(big.Rat).Sub(ra, new(big.Rat).Mul(rb, ratInt(wantQuo)))
			assert.Equal(t, wantQuo.Int64(), int64(quo), "%s QuoRem %s", a, b)
			assert.True(t, rem.EqualExact(fracRat(wantRem)), "%s QuoRem %s rem %s", a, b, rem)
			assert.GreaterOrEqual(t, rem.Sign(), 0)
			assert.Equal(t, -1, ratOf(rem).Cmp(rb), "rem < divisor")

			mul, err := a.Mul(b.MustFraction())
			require.NoError(t, err)
			assert.Equal(t, durRat(new(big.Rat).Mul(ra, rb)), mul)

			div, err := a.Div(b.MustFraction())
			require.NoError(t, err)
			assert.Equal(t, durRat(new(big.Rat).Quo(ra, rb)), div)

			assert.Equal(t, ra.Cmp(rb), a.Compare(b), "%s Compare %s", a, b)
			assert.Equal(t, ra.Cmp(rb) == 0, a.Equal(b))
			assert.Equal(t, a == b, a.Equal(b))
		}

		for n := 1; n <= 7; n++ {
			ra := ratOfDuration(a)

			mul, err := a.MulInt(n)
			require.NoError(t, err)
			assert.Equal(t, durRat(new(big.Rat).Mul(ra, rat(int64(n), 1))), mul)

			div, err := a.DivInt(n)
			require.NoError(t, err)
			assert.Equal(t, durRat(new(big.Rat).Quo(ra, rat(int64(n), 1))), div)
		}
	}
}

// TestDurationArithmeticExamples covers KP-14, KP-29, KP-44, and KP-48.
func TestDurationArithmeticExamples(t *testing.T) {
	t.Parallel()

	// KP-14.
	assert.Equal(t, dur(3, 8), dur(1, 4).MustAdd(dur(1, 8)))
	assert.True(t, dur(3, 8).MustRatio(dur(1, 8)).EqualExact(frac(3, 1)))
	assert.True(t, dur(1, 4).MustRatio(dur(1, 12)).EqualExact(frac(3, 1)))
	assert.True(t, dur(1, 1).MustRatio(dur(1, 64)).EqualExact(frac(64, 1)))
	assert.True(t, dur(3, 8).MustRatio(dur(1, 4)).EqualExact(frac(3, 2)))

	quo, rem := dur(1, 1).MustQuoRem(dur(3, 8))
	assert.Equal(t, 2, quo)
	assert.True(t, rem.EqualExact(frac(1, 4)), "rem %s", rem)

	quo, rem = dur(1, 8).MustQuoRem(dur(1, 4))
	assert.Equal(t, 0, quo)
	assert.True(t, rem.EqualExact(frac(1, 8)), "rem %s", rem)

	quo, rem = dur(3, 4).MustQuoRem(dur(1, 4))
	assert.Equal(t, 3, quo)
	assert.True(t, rem.EqualExact(fraction.Zero()), "a zero remainder is 0/1, got %s", rem)

	// KP-29.
	for _, c := range [][2]duration.Duration{{dur(1, 4), dur(1, 4)}, {dur(1, 8), dur(1, 4)}} {
		got, err := c[0].Sub(c[1])
		assertErr(t, err, duration.ErrNotPositive)
		assert.Equal(t, duration.Duration{}, got)
	}

	// KP-48.
	assert.Equal(t, dur(3, 4), dur(3, 8).MustMulInt(2))
	assert.Equal(t, dur(1, 6), dur(1, 4).MustMul(frac(2, 3)))
	assert.Equal(t, dur(1, 8), dur(3, 8).MustDivInt(3))
	assert.Equal(t, dur(1, 4), dur(3, 8).MustDiv(frac(3, 2)))
	assert.Equal(t, dur(1, 8), dur(1, 4).MustMul(frac(2, 4)), "an unreduced factor")
	assert.Equal(t, dur(1, 1), duration.MustSum(dur(1, 4), dur(1, 8), dur(1, 8), dur(1, 2)))
	assert.Equal(t, dur(1, 8), duration.MustMin(dur(1, 4), dur(3, 8), dur(1, 8)))
	assert.Equal(t, dur(3, 8), duration.MustMax(dur(1, 4), dur(3, 8), dur(1, 8)))
	assert.Equal(t, dur(1, 4), duration.MustSum(dur(1, 4)), "a sum of one value")
	assert.Equal(t, dur(1, 4), duration.MustMin(dur(1, 4)))
	assert.Equal(t, dur(1, 4), duration.MustMax(dur(1, 4), dur(1, 4)))
}

// TestDurationArgumentErrors covers KP-44 and architecture 4.4: factors and
// divisors not set or <= 0 are E-2 without fraction errors; integer factors
// below 1 are E-2; empty lists are E-2.
func TestDurationArgumentErrors(t *testing.T) {
	t.Parallel()

	d := dur(3, 8)

	for _, f := range []fraction.Fraction{{}, fraction.Zero(), frac(0, 5), frac(-1, 2), frac(1, -2)} {
		got, err := d.Mul(f)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, got, "Mul(%s)", f)

		got, err = d.Div(f)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, got, "Div(%s)", f)
	}

	for _, n := range []int{0, -1, math.MinInt} {
		got, err := d.MulInt(n)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, got, "MulInt(%d)", n)

		got, err = d.DivInt(n)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, got, "DivInt(%d)", n)
	}

	for _, f := range []func(...duration.Duration) (duration.Duration, error){duration.Sum, duration.Min, duration.Max} {
		got, err := f()
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, got)

		got, err = f([]duration.Duration{}...)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, got)
	}
}

// TestDurationRange covers E-11 and architecture 4.4: an exact result
// outside the fraction range is E-11 with fraction.ErrOutOfRange.
func TestDurationRange(t *testing.T) {
	t.Parallel()

	maxValue := durRat(ratInt(maxComponent()))
	minValue := durRat(new(big.Rat).SetFrac(big.NewInt(1), maxComponent()))

	// Two values whose denominators are coprime and about 2^200: 1/a − 1/b
	// and 1/a + 1/b have a denominator of about 2^400.
	a := durRat(new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, -1)))
	b := durRat(new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, 1)))

	for _, c := range []struct {
		name string
		run  func() (duration.Duration, error)
	}{
		{"Add", func() (duration.Duration, error) { return maxValue.Add(maxValue) }},
		{"Add denominators", func() (duration.Duration, error) { return a.Add(b) }},
		{"Sum", func() (duration.Duration, error) { return duration.Sum(maxValue, dur(1, 1)) }},
		{"Sub", func() (duration.Duration, error) { return a.Sub(b) }},
		{"MulInt", func() (duration.Duration, error) { return maxValue.MulInt(2) }},
		{"DivInt", func() (duration.Duration, error) { return minValue.DivInt(2) }},
		{"Mul", func() (duration.Duration, error) { return maxValue.Mul(frac(3, 2)) }},
		{"Div", func() (duration.Duration, error) { return minValue.Div(frac(2, 1)) }},
	} {
		got, err := c.run()
		assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, duration.Duration{}, got, c.name)
	}

	ratio, err := maxValue.Ratio(minValue)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, fraction.Fraction{}, ratio)

	// Results at the edge of the range are not errors.
	assert.Equal(t, maxValue, maxValue.MustSub(dur(1, 1)).MustAdd(dur(1, 1)))
	assert.Equal(t, dur(1, 1), minValue.MustMul(fraction.MustFromBigInt(maxComponent(), big.NewInt(1))))
	assert.Equal(t, minValue, dur(1, 1).MustDiv(fraction.MustFromBigInt(maxComponent(), big.NewInt(1))))
}

// TestQuoRemRange covers architecture 4.4: a quotient above math.MaxInt is
// E-11 with fraction.ErrNotRepresentable and without fraction.ErrOutOfRange;
// a remainder outside the range is E-11 with fraction.ErrOutOfRange even for
// values in range; when both fail, the quotient is reported.
func TestQuoRemRange(t *testing.T) {
	t.Parallel()

	maxValue := durRat(ratInt(maxComponent()))

	quo, rem, err := maxValue.QuoRem(dur(1, 1))
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrNotRepresentable)
	assert.Equal(t, 0, quo)
	assert.Equal(t, fraction.Fraction{}, rem)

	// The largest quotient that fits int.
	quo, rem = durRat(new(big.Rat).SetInt64(math.MaxInt)).MustQuoRem(dur(1, 1))
	assert.Equal(t, math.MaxInt, quo)
	assert.True(t, rem.EqualExact(fraction.Zero()))

	// d = (p+1)/p and o = (q+1)/(2q) for coprime p < q near 2^200: the
	// quotient is 2 and the remainder (q−p)/(pq) has a denominator near 2^400.
	p, q := near2pow(200, -1), near2pow(200, 1)
	d := durRat(new(big.Rat).SetFrac(new(big.Int).Add(p, big.NewInt(1)), p))
	o := durRat(new(big.Rat).SetFrac(new(big.Int).Add(q, big.NewInt(1)), new(big.Int).Mul(q, big.NewInt(2))))

	quo, rem, err = d.QuoRem(o)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, 0, quo)
	assert.Equal(t, fraction.Fraction{}, rem)

	// Both: d = (2^70·p + 1)/p and o = 1/q for p, q near 2^150. The quotient
	// is 2^70·q + 1 and the remainder (q mod p)/(pq) is out of range.
	p, q = near2pow(150, -1), near2pow(150, 1)
	num := new(big.Int).Add(new(big.Int).Lsh(p, 70), big.NewInt(1))
	d = durRat(new(big.Rat).SetFrac(num, p))
	o = durRat(new(big.Rat).SetFrac(big.NewInt(1), q))

	quo, rem, err = d.QuoRem(o)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrNotRepresentable)
	assert.Equal(t, 0, quo)
	assert.Equal(t, fraction.Fraction{}, rem)
}

// TestSumExact covers KR-D11: E-11 depends on the exact sum, not on partial
// sums or on the order of the list.
func TestSumExact(t *testing.T) {
	t.Parallel()

	// a = 2 − 2^−199, b = 1/MaxInt², c = 2^−199: a + c = 2 is small, while
	// a + b and b + c have a denominator near 2^325.
	a := durRat(new(big.Rat).Sub(big.NewRat(2, 1), pow2(-199)))
	b := durRat(new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Mul(big.NewInt(math.MaxInt), big.NewInt(math.MaxInt))))
	c := durRat(pow2(-199))
	want := durRat(new(big.Rat).Add(big.NewRat(2, 1), ratOfDuration(b)))

	_, err := a.Add(b)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)

	for _, order := range [][]duration.Duration{{a, b, c}, {a, c, b}, {b, a, c}, {b, c, a}, {c, a, b}, {c, b, a}} {
		got, err := duration.Sum(order...)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

// TestSumErrors covers architecture 4.2: an unset element is E-1.
func TestSumErrors(t *testing.T) {
	t.Parallel()

	for _, list := range [][]duration.Duration{
		{duration.Duration{}},
		{dur(1, 4), duration.Duration{}},
		{duration.Duration{}, dur(1, 4), dur(1, 8)},
	} {
		for _, f := range []func(...duration.Duration) (duration.Duration, error){duration.Sum, duration.Min, duration.Max} {
			got, err := f(list...)
			assertErr(t, err, duration.ErrUnset)
			assert.Equal(t, duration.Duration{}, got)
		}
	}
}

// TestDurationCompareSort covers architecture 4.4: Compare is a total order
// by value usable with slices.SortFunc.
func TestDurationCompareSort(t *testing.T) {
	t.Parallel()

	values := arithmeticValues()
	slices.SortFunc(values, duration.Duration.Compare)

	for i := 1; i < len(values); i++ {
		assert.Equal(t, -1, ratOfDuration(values[i-1]).Cmp(ratOfDuration(values[i])))
	}

	assert.Equal(t, -1, dur(1, 4).Compare(dur(3, 8)))
	assert.Equal(t, 0, dur(2, 8).Compare(dur(1, 4)))
	assert.Equal(t, 1, dur(1, 2).Compare(dur(3, 8)))
}

// TestDurationString covers architecture 4.12: the reduced n/d.
func TestDurationString(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		d    duration.Duration
		want string
	}{
		{dur(3, 8), "3/8"},
		{dur(6, 16), "3/8"},
		{dur(1, 1), "1/1"},
		{dur(2, 1), "2/1"},
		{dur(8, 4), "2/1"},
		{dur(1, 1024), "1/1024"},
		{durRat(ratInt(maxComponent())), fraction.MaxComponentString + "/1"},
	} {
		assert.Equal(t, c.want, c.d.String())
	}
}
