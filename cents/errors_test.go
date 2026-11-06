package cents_test

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
)

// K-32: every error is recognized as its own kind with errors.Is; the five
// kinds are pairwise distinguishable and do not wrap each other.
func TestErrorKinds(t *testing.T) {
	t.Parallel()

	for _, a := range allErrors() {
		for _, b := range allErrors() {
			assert.Equal(t, a == b, errors.Is(a, b), "errors.Is(%v, %v)", a, b)
		}
	}

	// One real error of every kind: each is exactly its own kind.
	_, errRange := cents.FromInt(9000001)
	_, errComplex := cents.Parse(strings.Repeat("1", 101))
	_, errPrime := cents.FromRatio(37, 32)
	_, errArg := cents.Step(1, 0)
	_, errText := cents.Parse("3/2")

	for _, tc := range []struct {
		err, want error
	}{
		{errRange, cents.ErrOutOfRange},
		{errComplex, cents.ErrTooComplex},
		{errPrime, cents.ErrPrimeLimit},
		{errArg, cents.ErrInvalidArgument},
		{errText, cents.ErrInvalidText},
	} {
		assertOnlyError(t, tc.err, tc.want, errName(tc.want))
	}
}

// K-32, NF-8: with several violations the error of the first check is
// returned. Values and numbers: invalid argument, prime limit, too complex,
// out of range. Text: syntax, text length, prime limit, record, range.
func TestErrorOrder(t *testing.T) {
	t.Parallel()

	p := new(big.Int).Add(pow10(99), big.NewInt(7)).String() // 100 digits

	// A value with e_3 = 10^20 and r = -floor(10^20 * J(3)) + 8000000 (floor
	// by Python decimal, 120 digits): in range; times 10^5 its exponent is
	// 10^25 > 10^24 (too complex) and its value is far out of range.
	bigExp := fromComp(t, bigRatC(new(big.Rat).SetInt(new(big.Int).Add(bigInt("-190195500086538741774448"), big.NewInt(8000000)))).
		withExp(3, new(big.Rat).SetInt(pow10(20))))

	top := noErr(cents.FromInt(9000000))(t)

	cases := []struct {
		name string
		call func() (cents.Cents, error)
		want error
	}{
		{"composition (1/0; 37 → 1)", func() (cents.Cents, error) {
			return cents.FromComposition(big.NewInt(1), big.NewInt(0), pw(37, 1, 1))
		}, cents.ErrInvalidArgument},
		{"composition (0; 37 → 1, 3 → 6000)", func() (cents.Cents, error) {
			return cents.FromComposition(nil, nil, pw(37, 1, 1), pw(3, 6000, 1))
		}, cents.ErrPrimeLimit},
		{"composition (9000000; 3 → 1/1000000)", func() (cents.Cents, error) {
			return cents.FromComposition(big.NewInt(9000000), nil, pw(3, 1, 1000000))
		}, cents.ErrOutOfRange},
		{"composition too complex and out of range", func() (cents.Cents, error) {
			return cents.FromComposition(big.NewInt(9000000), nil, cents.Power{Base: 3, Num: new(big.Int).Add(pow10(24), big.NewInt(1))})
		}, cents.ErrTooComplex},
		{"ratio 0:37", func() (cents.Cents, error) { return cents.FromRatio(0, 37) }, cents.ErrInvalidArgument},
		{"text 37:32", func() (cents.Cents, error) { return cents.Parse("37:32") }, cents.ErrPrimeLimit},
		{"text 3/2", func() (cents.Cents, error) { return cents.Parse("3/2") }, cents.ErrInvalidText},
		{"text 37 and 100 zeros : 1", func() (cents.Cents, error) {
			return cents.Parse("37" + strings.Repeat("0", 100) + ":1")
		}, cents.ErrTooComplex},
		{"text 65 factors with 37", func() (cents.Cents, error) {
			return cents.Parse(repeatFactors("37", "2", 64))
		}, cents.ErrTooComplex},
		{"text 65 factors and a trailing star", func() (cents.Cents, error) {
			return cents.Parse(repeatFactors("37", "2", 64) + "*")
		}, cents.ErrInvalidText},
		{"text 65 factors out of range", func() (cents.Cents, error) {
			return cents.Parse(repeatFactors("2^100000", "2", 64))
		}, cents.ErrTooComplex},
		{"text 2^100000*37", func() (cents.Cents, error) { return cents.Parse("2^100000*37") }, cents.ErrPrimeLimit},
		{"text record before range", func() (cents.Cents, error) {
			return cents.Parse("2^100000*3^(1/" + p + ")")
		}, cents.ErrTooComplex},
		{"Scala 37 and 100 zeros / 1", func() (cents.Cents, error) {
			c, _, err := cents.ParseScala("37" + strings.Repeat("0", 100) + "/1")

			return c, err
		}, cents.ErrTooComplex},
		{"Scala 5/4E with 37", func() (cents.Cents, error) {
			c, _, err := cents.ParseScala("37/32E")

			return c, err
		}, cents.ErrInvalidText},
		{"Mul too complex before range", func() (cents.Cents, error) { return bigExp.Mul(100000) }, cents.ErrTooComplex},
		{"Scale den 0 before range", func() (cents.Cents, error) {
			return top.Scale(maxInt, 0)
		}, cents.ErrInvalidArgument},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.call()
			assertResult(t, comp{}, tc.want, got, err)
		})
	}

	t.Run("NearestStep too complex before out of range", func(t *testing.T) {
		t.Parallel()

		c := fromComp(t, ratC(8999998, 1).add(ratC(1, 1000000)).add(stepC(1, 999983)))
		k, dev, err := c.NearestStep(maxInt)
		assertOnlyError(t, err, cents.ErrTooComplex)
		assert.Zero(t, k)
		assert.True(t, dev == cents.Cents{})
	})

	t.Run("ApplyTo invalid argument before range", func(t *testing.T) {
		t.Parallel()

		f, err := noErr(cents.FromInt(9000000))(t).ApplyTo(-1)
		assertOnlyError(t, err, cents.ErrInvalidArgument)
		assert.Zero(t, f)
	})
}
