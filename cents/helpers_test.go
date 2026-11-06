package cents_test

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests assume a 64-bit int (the project supports only 64-bit systems).
const (
	minInt = math.MinInt
	maxInt = math.MaxInt
)

// The oracle below follows the mathematical definition of the architecture:
// every value is an exact record r + sum e_p * J(p) over the odd primes
// 3...31, J(p) = 1200 * log2(p), with rational r and e_p. All record
// arithmetic is exact (math/big.Rat) and never calls the package under test.
// Numeric views of irrational values use J(p) with 80 digits after the point
// (Python decimal, 120 significant digits: 1200 * ln(p) / ln(2)).

// oddPrimes returns the odd primes of an exact record in increasing order.
func oddPrimes() [10]int {
	return [10]int{3, 5, 7, 11, 13, 17, 19, 23, 29, 31}
}

// allBases returns 2 and the odd primes of a record in increasing order.
func allBases() []int {
	ps := oddPrimes()

	return append([]int{2}, ps[:]...)
}

// primeIndex returns the index of p in oddPrimes or -1.
func primeIndex(p int) int {
	ps := oddPrimes()

	return slices.Index(ps[:], p)
}

// oraclePrec is the precision of big.Float views, well above the 80 digits
// of the J(p) literals.
const oraclePrec = 512

// jTable holds J(p) for the odd primes, parsed once; read-only afterwards.
var jTable = sync.OnceValue(func() [10]*big.Float {
	digits := [10]string{
		"1901.95500086538741774448673273737981051177728923097727254690318544931787335323027503",
		"2786.31371386483481744438331538726821103779767162949673446570767497912173193035025902",
		"3368.82590646912492893036318067819697036923195115936894041275006888438501863463158380",
		"4151.31794236475670743923565607095155044387783081812168557536197487159674366402202325",
		"4440.52766176931059247617518510803368035412368214924484344615623101142622395963036018",
		"4904.95540950040728990487921297248522481352073881378482575193087726403581085685848196",
		"5097.51301613230259255222330748820130723220908359384081749782967703747824984655786416",
		"5428.27434726841544675297789299520261339859015053106607139333247841773377467531763533",
		"5829.57719415308654486372798955358171497223260009112445766233213134571560365372412197",
		"5945.03557246425025056734831901066530825882678591565680264075846649512483023015038421",
	}

	var out [10]*big.Float
	for i, s := range digits {
		f, _, err := big.ParseFloat(s, 10, oraclePrec, big.ToNearestEven)
		if err != nil {
			panic(err)
		}

		out[i] = f
	}

	return out
})

// rat returns num/den as a new big.Rat.
func rat(num, den int64) *big.Rat {
	return big.NewRat(num, den)
}

// bigInt parses a decimal integer.
func bigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("bad integer literal " + s)
	}

	return n
}

// pow10 returns 10^n.
func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

// comp is the oracle of a value: the record (r; e_3...e_31). bad collects
// parts of a composition that are not a record (another base, a zero
// denominator); a comp with bad parts equals no proper record.
type comp struct {
	r   *big.Rat
	e   [10]*big.Rat
	bad string
}

func zeroComp() comp {
	c := comp{r: new(big.Rat)}
	for i := range c.e {
		c.e[i] = new(big.Rat)
	}

	return c
}

// ratC returns the rational value num/den cents.
func ratC(num, den int64) comp {
	c := zeroComp()
	c.r.SetFrac64(num, den)

	return c
}

// bigRatC returns the rational value x cents.
func bigRatC(x *big.Rat) comp {
	c := zeroComp()
	c.r.Set(x)

	return c
}

// clone returns a deep copy of c.
func (c comp) clone() comp {
	out := comp{r: new(big.Rat).Set(c.r), bad: c.bad}
	for i := range c.e {
		out.e[i] = new(big.Rat).Set(c.e[i])
	}

	return out
}

// withExp returns c with num/den added to the exponent of base p
// (2 adds 1200 * num/den cents to r).
func (c comp) withExp(p int, x *big.Rat) comp {
	out := c.clone()
	if p == 2 {
		out.r.Add(out.r, new(big.Rat).Mul(x, rat(1200, 1)))

		return out
	}

	i := primeIndex(p)
	if i < 0 {
		panic(fmt.Sprintf("base %d is not an odd prime of the record", p))
	}

	out.e[i].Add(out.e[i], x)

	return out
}

func (c comp) add(d comp) comp {
	out := c.clone()
	out.r.Add(out.r, d.r)
	for i := range out.e {
		out.e[i].Add(out.e[i], d.e[i])
	}

	out.bad += d.bad

	return out
}

func (c comp) neg() comp {
	out := c.clone()
	out.r.Neg(out.r)
	for i := range out.e {
		out.e[i].Neg(out.e[i])
	}

	return out
}

func (c comp) sub(d comp) comp {
	return c.add(d.neg())
}

// scale returns c * k.
func (c comp) scale(k *big.Rat) comp {
	out := c.clone()
	out.r.Mul(out.r, k)
	for i := range out.e {
		out.e[i].Mul(out.e[i], k)
	}

	return out
}

// mulInt returns c * k.
func (c comp) mulInt(k int) comp {
	return c.scale(new(big.Rat).SetInt64(int64(k)))
}

// equal reports whether two records are the same.
func (c comp) equal(d comp) bool {
	if c.bad != d.bad || c.r.Cmp(d.r) != 0 {
		return false
	}

	for i := range c.e {
		if c.e[i].Cmp(d.e[i]) != 0 {
			return false
		}
	}

	return true
}

// rational reports whether all exponents are zero.
func (c comp) rational() bool {
	for _, e := range c.e {
		if e.Sign() != 0 {
			return false
		}
	}

	return true
}

// ratioParts returns n and d of the ratio n:d whose cents the value is
// (all exponents integer and r a multiple of 1200).
func (c comp) ratioParts() (n, d *big.Int, ok bool) {
	two := new(big.Rat).Quo(c.r, rat(1200, 1))
	if !two.IsInt() {
		return nil, nil, false
	}

	n, d = big.NewInt(1), big.NewInt(1)
	mulPow := func(p int64, e *big.Int) {
		if e.Sign() == 0 {
			return
		}

		pw := new(big.Int).Exp(big.NewInt(p), new(big.Int).Abs(e), nil)
		if e.Sign() > 0 {
			n.Mul(n, pw)
		} else {
			d.Mul(d, pw)
		}
	}

	mulPow(2, two.Num())
	ps := oddPrimes()
	for i, e := range c.e {
		if !e.IsInt() {
			return nil, nil, false
		}

		mulPow(int64(ps[i]), e.Num())
	}

	return n, d, true
}

// float returns the value of c in cents (big.Float, about 80 digits).
func (c comp) float() *big.Float {
	v := new(big.Float).SetPrec(oraclePrec).SetRat(c.r)
	js := jTable()
	for i, e := range c.e {
		if e.Sign() == 0 {
			continue
		}

		term := new(big.Float).SetPrec(oraclePrec).SetRat(e)
		term.Mul(term, js[i])
		v.Add(v, term)
	}

	return v
}

// newFloat returns a big.Float of the oracle precision.
func newFloat() *big.Float {
	return new(big.Float).SetPrec(oraclePrec)
}

// errBound returns an upper bound of |c.float() - value of c|, doubled for
// safety. Each J(p) literal is within 10^-80 of J(p), so the exponents
// contribute 10^-80 * sum |e_p|; the big.Float operations (512 bits, a
// relative error of 2^-512 each, fewer than 30 operations) contribute less
// than 2^-480 * (|r| + 6000 * sum |e_p|), J(p) < 6000.
func errBound(c comp) *big.Float {
	sumE := new(big.Rat)
	for _, e := range c.e {
		sumE.Add(sumE, new(big.Rat).Abs(e))
	}

	literal := newFloat().SetRat(new(big.Rat).Mul(sumE, new(big.Rat).SetFrac(big.NewInt(1), pow10(80))))

	size := new(big.Rat).Add(new(big.Rat).Abs(c.r), new(big.Rat).Mul(sumE, rat(6000, 1)))
	size.Add(size, rat(1, 1))
	ops := newFloat().SetRat(size)
	ops.Mul(ops, newFloat().SetMantExp(big.NewFloat(1), -480))

	out := literal.Add(literal, ops)

	return out.Mul(out, newFloat().SetInt64(2))
}

// scaledBound returns errBound(c) * |m|: the bound of value * m.
func scaledBound(c comp, m *big.Rat) *big.Float {
	b := errBound(c)

	return b.Mul(b, newFloat().SetRat(new(big.Rat).Abs(m)))
}

// assertMargin panics when an irrational view is not farther than bound
// from a decision point (zero, a range limit, a rounding midpoint, an
// integer), so the oracle cannot decide from the view; this never happens
// for the values of the tests.
func assertMargin(dist, bound *big.Float, what string) {
	if newFloat().Abs(dist).Cmp(bound) <= 0 {
		panic("oracle precision is not enough for " + what)
	}
}

// sign returns the sign of the value.
func (c comp) sign() int {
	if c.rational() {
		return c.r.Sign()
	}

	v := c.float()
	assertMargin(v, errBound(c), "the sign")

	return v.Sign()
}

// cmp compares the values of c and d.
func (c comp) cmp(d comp) int {
	return c.sub(d).sign()
}

func limitDen() *big.Int { return pow10(30) }
func limitExp() *big.Int { return pow10(24) }

// tooComplex reports whether the record is outside the record limits:
// denominators up to 10^30, |e_p| up to 10^24.
func (c comp) tooComplex() bool {
	if c.r.Denom().Cmp(limitDen()) > 0 {
		return true
	}

	for _, e := range c.e {
		if e.Denom().Cmp(limitDen()) > 0 {
			return true
		}

		if new(big.Rat).Abs(e).Cmp(new(big.Rat).SetInt(limitExp())) > 0 {
			return true
		}
	}

	return false
}

// outOfRange reports whether |value| > 9 000 000.
func (c comp) outOfRange() bool {
	limit := ratC(cents.MaxMagnitude, 1)

	return c.cmp(limit) > 0 || c.cmp(limit.neg()) < 0
}

// wantErr returns the error of an operation whose exact result is c, by the
// order of checks: the record limit, then the range.
func (c comp) wantErr() error {
	switch {
	case c.tooComplex():
		return cents.ErrTooComplex
	case c.outOfRange():
		return cents.ErrOutOfRange
	default:
		return nil
	}
}

// parts returns the composition of c for FromComposition: r and the
// non-zero exponents of the odd primes in increasing order.
func (c comp) parts() (num, den *big.Int, ps []cents.Power) {
	num, den = new(big.Int).Set(c.r.Num()), new(big.Int).Set(c.r.Denom())
	ps = []cents.Power{}
	for i, p := range oddPrimes() {
		if c.e[i].Sign() == 0 {
			continue
		}

		ps = append(ps, cents.Power{
			Base: p,
			Num:  new(big.Int).Set(c.e[i].Num()),
			Den:  new(big.Int).Set(c.e[i].Denom()),
		})
	}

	return num, den, ps
}

// String is the record in the notation of the criteria: "(r; p → e, ...)".
func (c comp) String() string {
	var b strings.Builder
	b.WriteString("(" + c.r.RatString() + ";")
	sep := " "
	for i, p := range oddPrimes() {
		if c.e[i].Sign() == 0 {
			continue
		}

		fmt.Fprintf(&b, "%s%d → %s", sep, p, c.e[i].RatString())
		sep = ", "
	}

	if sep == " " {
		b.WriteString(" —")
	}

	if c.bad != "" {
		b.WriteString("; bad: " + c.bad)
	}

	b.WriteString(")")

	return b.String()
}

// compOf converts the composition reported by the package into a record.
// It does not trust the shape: nil numbers, base 2, unknown bases and zero
// denominators are tolerated and recorded (shape checks are in values_test).
func compOf(v cents.Cents) comp {
	num, den, ps := v.Composition()
	out := zeroComp()

	if num != nil {
		out.r.SetInt(num)
	}

	if den != nil {
		if den.Sign() == 0 {
			out.bad += " zero denominator"
		} else {
			out.r.Quo(out.r, new(big.Rat).SetInt(den))
		}
	}

	for _, p := range ps {
		e := new(big.Rat)
		if p.Num != nil {
			e.SetInt(p.Num)
		}

		if p.Den != nil {
			if p.Den.Sign() == 0 {
				out.bad += fmt.Sprintf(" base %d: zero denominator", p.Base)

				continue
			}

			e.Quo(e, new(big.Rat).SetInt(p.Den))
		}

		if p.Base != 2 && primeIndex(p.Base) < 0 {
			out.bad += fmt.Sprintf(" base %d", p.Base)

			continue
		}

		out = out.withExp(p.Base, e)
	}

	return out
}

// ratioComp returns the record of R(a:b) for a, b > 0 by reducing a/b and
// factoring both numbers; ok is false when a prime greater than 31 remains
// in the reduced ratio (primes are checked after reducing: 111:74 is 3:2).
func ratioComp(a, b *big.Int) (comp, bool) {
	g := new(big.Int).GCD(nil, nil, a, b)
	a, b = new(big.Int).Quo(a, g), new(big.Int).Quo(b, g)

	out := zeroComp()
	for _, part := range []struct {
		n    *big.Int
		sign int64
	}{{a, 1}, {b, -1}} {
		n := new(big.Int).Set(part.n)
		for _, p := range allBases() {
			bp := big.NewInt(int64(p))
			m := new(big.Int)
			for {
				q, r := new(big.Int).QuoRem(n, bp, m)
				if r.Sign() != 0 {
					break
				}

				n = q
				out = out.withExp(p, rat(part.sign, 1))
			}
		}

		if n.Cmp(big.NewInt(1)) != 0 {
			return comp{}, false
		}
	}

	return out, true
}

// ratio returns the record of R(a:b) for small a, b > 0 with primes <= 31.
func ratio(a, b int64) comp {
	c, ok := ratioComp(big.NewInt(a), big.NewInt(b))
	if !ok {
		panic(fmt.Sprintf("ratio %d:%d has a prime greater than 31", a, b))
	}

	return c
}

// stepC returns the record of E(k, n) = k * 1200 / n.
func stepC(k, n int64) comp {
	return bigRatC(new(big.Rat).SetFrac(new(big.Int).Mul(big.NewInt(k), big.NewInt(1200)), big.NewInt(n)))
}

// noErr takes the results of a call with an error, for example
// noErr(cents.FromRatio(3, 2))(t, "R(3:2)"), and returns the value after
// checking that the error is nil; msgAndArgs describe the call in the
// failure message (the line of the call is reported in any case).
func noErr[T any](v T, err error) func(t testing.TB, msgAndArgs ...any) T {
	return func(t testing.TB, msgAndArgs ...any) T {
		t.Helper()
		require.NoError(t, err, msgAndArgs...)

		return v
	}
}

// fromComp creates the value of a record with FromComposition.
func fromComp(t testing.TB, c comp) cents.Cents {
	t.Helper()

	num, den, ps := c.parts()

	return noErr(cents.FromComposition(num, den, ps...))(t)
}

// assertCents checks that got is the value want: Equal, == and the same
// canonical text (the text makes the failure message readable).
func assertCents(t testing.TB, want, got cents.Cents, msgAndArgs ...any) bool {
	t.Helper()

	ok := assert.Equal(t, want.String(), got.String(), msgAndArgs...)
	ok = assert.True(t, want.Equal(got), msgAndArgs...) && ok
	ok = assert.True(t, want == got, msgAndArgs...) && ok

	return ok
}

// assertComp checks that the composition of got is the record want.
func assertComp(t testing.TB, want comp, got cents.Cents, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Equal(t, want.String(), compOf(got).String(), msgAndArgs...)
}

// allErrors returns the five sentinel errors.
func allErrors() []error {
	return []error{
		cents.ErrOutOfRange,
		cents.ErrTooComplex,
		cents.ErrPrimeLimit,
		cents.ErrInvalidArgument,
		cents.ErrInvalidText,
	}
}

// errName returns a short name of a sentinel error for test names.
func errName(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, cents.ErrOutOfRange):
		return "out of range"
	case errors.Is(err, cents.ErrTooComplex):
		return "too complex"
	case errors.Is(err, cents.ErrPrimeLimit):
		return "prime limit"
	case errors.Is(err, cents.ErrInvalidArgument):
		return "invalid argument"
	case errors.Is(err, cents.ErrInvalidText):
		return "invalid text"
	default:
		return "unknown error"
	}
}

// assertOnlyError checks that err is the sentinel want and none of the
// other four (the kinds are distinguishable, K-32).
func assertOnlyError(t testing.TB, err, want error, msgAndArgs ...any) bool {
	t.Helper()

	ok := assert.ErrorIs(t, err, want, msgAndArgs...)
	for _, other := range allErrors() {
		if other == want {
			continue
		}

		ok = assert.NotErrorIs(t, err, other, msgAndArgs...) && ok
	}

	return ok
}

// requireOnlyError is assertOnlyError that stops the test on a wrong error.
func requireOnlyError(t testing.TB, err, want error, msgAndArgs ...any) {
	t.Helper()

	require.ErrorIs(t, err, want, msgAndArgs...)
	assertOnlyError(t, err, want, msgAndArgs...)
}

// assertResult checks a result with an error: the value want and no error,
// or the zero value and only the error wantErr.
func assertResult(t testing.TB, want comp, wantErr error, got cents.Cents, err error, msgAndArgs ...any) bool {
	t.Helper()

	if wantErr != nil {
		ok := assertOnlyError(t, err, wantErr, msgAndArgs...)

		return assert.True(t, got == cents.Cents{}, msgAndArgs...) && ok
	}

	if !assert.NoError(t, err, msgAndArgs...) {
		return false
	}

	return assertComp(t, want, got, msgAndArgs...)
}

// panicValue calls f and reports whether it panicked and with what value.
func panicValue(f func()) (recovered any, panicked bool) {
	panicked = true

	defer func() {
		if panicked {
			recovered = recover()
		}
	}()

	f()

	panicked = false

	return nil, false
}

// assertPanicsWithErrorIs checks that f panics with an error value that is
// the sentinel want (errors.Is), not by comparing texts.
func assertPanicsWithErrorIs(t testing.TB, want error, f func(), msgAndArgs ...any) bool {
	t.Helper()

	v, panicked := panicValue(f)
	if !assert.True(t, panicked, msgAndArgs...) {
		return false
	}

	err, isErr := v.(error)
	if !assert.True(t, isErr, msgAndArgs...) {
		return false
	}

	return assertOnlyError(t, err, want, msgAndArgs...)
}

// sentinelOf returns the sentinel error that err is (errors.Is), or err
// itself when it is none of the five.
func sentinelOf(err error) error {
	for _, s := range allErrors() {
		if errors.Is(err, s) {
			return s
		}
	}

	return err
}

// assertMust checks the Must variant of an operation on the inputs of a
// call whose results were got and err (K-31, E-6): when the operation
// returned an error, the variant panics with an error of the same kind
// (errors.Is on the recovered value); otherwise it does not panic and
// returns the same results (==). The callers check got and err against the
// oracle of the case.
func assertMust[T comparable](t testing.TB, got T, err error, must func() T, msgAndArgs ...any) bool {
	t.Helper()

	if err != nil {
		return assertPanicsWithErrorIs(t, sentinelOf(err), func() { must() }, msgAndArgs...)
	}

	var fromMust T

	v, panicked := panicValue(func() { fromMust = must() })
	if panicked {
		return assert.Fail(t, fmt.Sprintf("the Must variant panicked with %v where the operation has no error", v), msgAndArgs...)
	}

	if got != fromMust {
		return assert.Fail(t, fmt.Sprintf("the Must variant returned %v, the operation %v", fromMust, got), msgAndArgs...)
	}

	return true
}

// mustNearest is the results of MustNearestStep as one comparable value.
type mustNearest struct {
	k   int
	dev cents.Cents
}

// mustScala is the results of MustParseScala and MustParseScalaApprox.
type mustScala struct {
	c    cents.Cents
	rest string
}

// named is a value of a test set together with its oracle record.
type named struct {
	name string
	v    cents.Cents
	c    comp
}

// setS returns the check set S of the criteria (20 values) in the order of
// the criteria list. The records are written from the definitions:
// R(a:b) by factoring, E(k, N) = k*1200/N, Q = R(3:2) - SYNT/4,
// MK = R(3:2)*53 - <37200>, SAV = (6/5; 5 → 1/1000).
func setS(t testing.TB) []named {
	t.Helper()

	r32 := noErr(cents.FromRatio(3, 2))(t)
	synt := cents.SyntonicComma()
	quarter := noErr(synt.Div(4))(t)

	return []named{
		{"0", cents.Zero(), zeroComp()},
		{"<100>", noErr(cents.FromInt(100))(t), ratC(100, 1)},
		{"<1200>", noErr(cents.FromInt(1200))(t), ratC(1200, 1)},
		{"<-1200>", noErr(cents.FromInt(-1200))(t), ratC(-1200, 1)},
		{"<700>", noErr(cents.FromInt(700))(t), ratC(700, 1)},
		{"E(1,19)", noErr(cents.Step(1, 19))(t), stepC(1, 19)},
		{"<7.1>", noErr(cents.FromFraction(71, 10))(t), ratC(71, 10)},
		{"<-13.7>", noErr(cents.FromFraction(-137, 10))(t), ratC(-137, 10)},
		{"R(3:2)", r32, ratio(3, 2)},
		{"R(5:4)", noErr(cents.FromRatio(5, 4))(t), ratio(5, 4)},
		{"R(2:3)", noErr(cents.FromRatio(2, 3))(t), ratio(2, 3)},
		{"SYNT", synt, ratio(81, 80)},
		{"PYTH", cents.PythagoreanComma(), ratio(531441, 524288)},
		{"Q", noErr(r32.Sub(quarter))(t), zeroComp().withExp(5, rat(1, 4))},
		{"MK", noErr(noErr(r32.Mul(53))(t).Sub(noErr(cents.FromInt(37200))(t)))(t), ratC(-100800, 1).withExp(3, rat(53, 1))},
		{"HOLDER", cents.HoldrianComma(), stepC(1, 53)},
		{"R(442:440)", noErr(cents.FromRatio(442, 440))(t), ratio(442, 440)},
		{"SAV", cents.Savart(), ratC(6, 5).withExp(5, rat(1, 1000))},
		{"<9000000>", noErr(cents.FromInt(9000000))(t), ratC(9000000, 1)},
		{"<-9000000>", noErr(cents.FromInt(-9000000))(t), ratC(-9000000, 1)},
	}
}

// sAscending returns the names of S in increasing order (K-15 literal).
func sAscending() []string {
	return []string{
		"<-9000000>", "<-1200>", "R(2:3)", "<-13.7>", "0", "MK", "SAV", "<7.1>",
		"R(442:440)", "SYNT", "HOLDER", "PYTH", "E(1,19)", "<100>", "R(5:4)", "Q",
		"<700>", "R(3:2)", "<1200>", "<9000000>",
	}
}

// byName returns the element of a set with the name.
func byName(t testing.TB, set []named, name string) named {
	t.Helper()

	i := slices.IndexFunc(set, func(n named) bool { return n.name == name })
	require.GreaterOrEqual(t, i, 0, "no value %q in the set", name)

	return set[i]
}

// roundHalfAway returns the integer nearest to x, half away from zero.
func roundHalfAway(x *big.Rat) *big.Int {
	num := new(big.Int).Abs(x.Num())
	num.Mul(num, big.NewInt(2))
	num.Add(num, x.Denom())
	q := new(big.Int).Quo(num, new(big.Int).Mul(x.Denom(), big.NewInt(2)))
	if x.Sign() < 0 {
		q.Neg(q)
	}

	return q
}

// roundedScaled returns the integer nearest to |value| * 10^digits (half
// away from zero; only rational values have exact halves) and the sign of
// the value.
func roundedScaled(c comp, digits int) (q *big.Int, sign int) {
	scale := new(big.Rat).SetInt(pow10(digits))
	if c.rational() {
		x := new(big.Rat).Mul(c.r, scale)

		return new(big.Int).Abs(roundHalfAway(x)), x.Sign()
	}

	v := c.float()
	v.Mul(v, new(big.Float).SetPrec(oraclePrec).SetRat(scale))
	sign = v.Sign()
	v.Abs(v)

	floor, _ := v.Int(nil)
	frac := new(big.Float).SetPrec(oraclePrec).Sub(v, new(big.Float).SetPrec(oraclePrec).SetInt(floor))
	half := new(big.Float).SetPrec(oraclePrec).SetFloat64(0.5)
	assertMargin(newFloat().Sub(frac, half), scaledBound(c, scale), "decimal rounding")

	if frac.Cmp(half) > 0 {
		floor.Add(floor, big.NewInt(1))
	}

	return floor, sign
}

// scaledFloat returns value * m as a big.Float.
func scaledFloat(c comp, m *big.Rat) *big.Float {
	v := c.float()

	return v.Mul(v, new(big.Float).SetPrec(oraclePrec).SetRat(m))
}

// nearestScaled returns the integer nearest to value * m, half away from
// zero (exact halves exist only for rational values).
func nearestScaled(c comp, m *big.Rat) *big.Int {
	if c.rational() {
		return roundHalfAway(new(big.Rat).Mul(c.r, m))
	}

	v := scaledFloat(c, m)
	neg := v.Sign() < 0
	v.Abs(v)

	n, _ := v.Int(nil)
	frac := new(big.Float).SetPrec(oraclePrec).Sub(v, new(big.Float).SetPrec(oraclePrec).SetInt(n))
	half := new(big.Float).SetPrec(oraclePrec).SetFloat64(0.5)
	assertMargin(newFloat().Sub(frac, half), scaledBound(c, m), "rounding to an integer")

	if frac.Cmp(half) > 0 {
		n.Add(n, big.NewInt(1))
	}

	if neg {
		n.Neg(n)
	}

	return n
}

// floorScaled returns floor(value * m).
func floorScaled(c comp, m *big.Rat) *big.Int {
	if c.rational() {
		x := new(big.Rat).Mul(c.r, m)

		return new(big.Int).Div(x.Num(), x.Denom()) // Euclidean: floor for den > 0
	}

	v := scaledFloat(c, m)
	n, acc := v.Int(nil) // truncates toward zero
	frac := new(big.Float).SetPrec(oraclePrec).Sub(v, new(big.Float).SetPrec(oraclePrec).SetInt(n))
	// The fractional part decides near both integers: near 0 and near ±1.
	bound := scaledBound(c, m)
	assertMargin(frac, bound, "floor")
	assertMargin(newFloat().Sub(newFloat().Abs(frac), newFloat().SetInt64(1)), bound, "floor")

	if v.Sign() < 0 && acc != big.Exact {
		n.Sub(n, big.NewInt(1))
	}

	return n
}

// unsignedDecimal formats q / 10^digits without a sign.
func unsignedDecimal(q *big.Int, digits int) string {
	s := q.String()
	if digits == 0 {
		return s
	}

	if len(s) <= digits {
		s = strings.Repeat("0", digits-len(s)+1) + s
	}

	return s[:len(s)-digits] + "." + s[len(s)-digits:]
}

// unitOf returns the unit printed after the unsigned number s.
func unitOf(s string, opts cents.Options) string {
	switch {
	case opts&cents.Typographic != 0:
		return "¢"
	case s == "1":
		return "cent"
	default:
		return "cents"
	}
}

// styled adds the sign and the unit to an unsigned number: the minus "-"
// or U+2212, "+" with WithPlus for positive numbers, no sign for a printed
// zero; the unit after a space when unit is true.
func styled(s string, sign int, zero bool, opts cents.Options, unit bool) string {
	switch {
	case zero:
	case sign < 0 && opts&cents.Typographic != 0:
		s = "−" + s
	case sign < 0:
		s = "-" + s
	case sign > 0 && opts&cents.WithPlus != 0:
		s = "+" + s
	}

	if unit {
		s += " " + unitOf(strings.TrimLeft(s, "+-−"), opts)
	}

	return s
}

// decimalOracle is the expected Decimal(digits, opts): correctly rounded,
// half away from zero, no minus for a value rounded to zero, digits < 0 as 0.
func decimalOracle(c comp, digits int, opts cents.Options) string {
	digits = max(digits, 0)
	q, sign := roundedScaled(c, digits)

	return styled(unsignedDecimal(q, digits), sign, q.Sign() == 0, opts, opts&cents.WithUnit != 0)
}

// exactDecimal returns the shortest exact decimal of a rational whose
// denominator is 2^a * 5^b.
func exactDecimal(x *big.Rat) (string, bool) {
	d := new(big.Int).Set(x.Denom())
	digits := 0
	for _, p := range []int64{2, 5} {
		bp := big.NewInt(p)
		m := new(big.Int)
		for {
			q, r := new(big.Int).QuoRem(d, bp, m)
			if r.Sign() != 0 {
				break
			}

			d = q
		}
	}

	if d.Cmp(big.NewInt(1)) != 0 {
		return "", false
	}

	for !new(big.Rat).Mul(x, new(big.Rat).SetInt(pow10(digits))).IsInt() {
		digits++
	}

	scaled := new(big.Rat).Mul(x, new(big.Rat).SetInt(pow10(digits)))

	return unsignedDecimal(new(big.Int).Abs(scaled.Num()), digits), true
}

// powerText returns the product of powers of a record (KindPower).
func powerText(c comp) string {
	type factor struct {
		p int
		e *big.Rat
	}

	fs := []factor{{2, new(big.Rat).Quo(c.r, rat(1200, 1))}}
	for i, p := range oddPrimes() {
		fs = append(fs, factor{p, c.e[i]})
	}

	fs = slices.DeleteFunc(fs, func(f factor) bool { return f.e.Sign() == 0 })
	if len(fs) == 0 {
		return "2^0"
	}

	parts := make([]string, 0, len(fs))
	for _, f := range fs {
		base := strconv.Itoa(f.p)
		switch {
		case f.e.Cmp(rat(1, 1)) == 0 && len(fs) == 1:
			parts = append(parts, base+"^1")
		case f.e.Cmp(rat(1, 1)) == 0:
			parts = append(parts, base)
		case f.e.IsInt():
			parts = append(parts, base+"^"+f.e.Num().String())
		default:
			parts = append(parts, base+"^("+f.e.RatString()+")")
		}
	}

	return strings.Join(parts, "*")
}

// textOracle is the expected TextAs(k, opts) of a record (architecture,
// "Text: kinds, styles, output").
func textOracle(c comp, k cents.Kind, opts cents.Options) (string, bool) {
	switch k {
	case cents.KindDecimal:
		if !c.rational() {
			return "", false
		}

		s, ok := exactDecimal(c.r)
		if !ok {
			return "", false
		}

		return styled(s, c.r.Sign(), c.r.Sign() == 0, opts, opts&cents.WithUnit != 0), true
	case cents.KindFraction:
		if !c.rational() {
			return "", false
		}

		s := new(big.Int).Abs(c.r.Num()).String() + "/" + c.r.Denom().String()

		return styled(s, c.r.Sign(), c.r.Sign() == 0, opts, true), true
	case cents.KindRatio:
		n, d, ok := c.ratioParts()
		if !ok || len(n.String()) > cents.MaxDigits || len(d.String()) > cents.MaxDigits {
			return "", false
		}

		return n.String() + ":" + d.String(), true
	case cents.KindPower:
		return powerText(c), true
	default:
		return "", false
	}
}

// kindOracle is the expected canonical kind (Kind()).
func kindOracle(c comp) cents.Kind {
	for _, k := range []cents.Kind{cents.KindDecimal, cents.KindFraction, cents.KindRatio} {
		if _, ok := textOracle(c, k, 0); ok {
			return k
		}
	}

	return cents.KindPower
}

// stringOracle is the expected String() (canonical kind, ASCII, no options).
func stringOracle(c comp) string {
	s, _ := textOracle(c, kindOracle(c), 0)

	return s
}

// scalaOracle is the expected Scala(digits): n/d for a ratio with n, d up to
// 2^31-1, otherwise cents with at least one digit after the point.
func scalaOracle(c comp, digits int) string {
	n, d, ok := c.ratioParts()
	limit := big.NewInt(1<<31 - 1)
	if ok && n.Cmp(limit) <= 0 && d.Cmp(limit) <= 0 {
		return n.String() + "/" + d.String()
	}

	return decimalOracle(c, max(digits, 1), 0)
}

// floatOracle is the expected Float64(): the nearest float64 (ties to even;
// only rational values have ties, they are rounded exactly by big.Rat).
func floatOracle(c comp) float64 {
	if c.rational() {
		f, _ := c.r.Float64()

		return f
	}

	return nearestFloat64(c.float(), errBound(c), "Float64")
}

// nearestFloat64 returns the float64 nearest to a view v whose error is at
// most bound; it panics when v - bound and v + bound round differently (the
// view is too close to a midpoint between two float64 for the oracle).
func nearestFloat64(v, bound *big.Float, what string) float64 {
	f, _ := v.Float64()
	lo, _ := newFloat().Sub(v, bound).Float64()
	hi, _ := newFloat().Add(v, bound).Float64()

	if lo != f || hi != f {
		panic("oracle precision is not enough for " + what)
	}

	return f
}

// gridOracle is the expected value of FromFloat(x) for a finite x: the
// shortest decimal of x as an exact rational, rounded to a multiple of
// 10^-6 cent, half away from zero.
func gridOracle(x float64) comp {
	s := strconv.FormatFloat(x, 'g', -1, 64)
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic("cannot parse " + s)
	}

	grid := new(big.Rat).SetInt64(cents.GridDenominator)
	q := roundHalfAway(new(big.Rat).Mul(r, grid))

	return bigRatC(new(big.Rat).SetFrac(q, big.NewInt(cents.GridDenominator)))
}

// optionSets returns the 8 combinations of the three options.
func optionSets() []cents.Options {
	out := make([]cents.Options, 0, 8)
	for o := range cents.Options(8) {
		out = append(out, o)
	}

	return out
}

// kinds returns the four text kinds.
func kinds() []cents.Kind {
	return []cents.Kind{cents.KindDecimal, cents.KindFraction, cents.KindRatio, cents.KindPower}
}

// isASCII reports whether s has only bytes below 0x80.
func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= 0x80 {
			return false
		}
	}

	return true
}

// TestOracleSelfCheck checks the oracle on literals of the criteria before
// other tests rely on it.
func TestOracleSelfCheck(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "701.955001", decimalOracle(ratio(3, 2), 6, 0))
	assert.Equal(t, "701.955000865387418", decimalOracle(ratio(3, 2), 15, 0))
	assert.Equal(t, "-701.955", decimalOracle(ratio(2, 3), 3, 0))
	assert.Equal(t, "−701.955", decimalOracle(ratio(2, 3), 3, cents.Typographic))
	assert.Equal(t, "702", decimalOracle(ratio(3, 2), 0, 0))
	assert.Equal(t, "0.000", decimalOracle(ratC(-4, 10000), 3, 0))
	assert.Equal(t, "0.13", decimalOracle(ratC(1, 8), 2, 0))
	assert.Equal(t, "-0.13", decimalOracle(ratC(-1, 8), 2, 0))
	assert.Equal(t, "696.578428", decimalOracle(zeroComp().withExp(5, rat(1, 4)), 6, 0))
	assert.Equal(t, "3.615046", decimalOracle(ratC(-100800, 1).withExp(3, rat(53, 1)), 6, 0))
	assert.Equal(t, "1 cent", decimalOracle(ratC(2499, 2500), 0, cents.WithUnit))
	assert.Equal(t, "1.000 cents", decimalOracle(ratC(2499, 2500), 3, cents.WithUnit))
	assert.Equal(t, "0 cents", decimalOracle(ratC(-1, 2500), 0, cents.WithUnit|cents.WithPlus))

	for _, tc := range []struct {
		c    comp
		want string
	}{
		{ratC(700, 1), "700"},
		{ratC(71, 10), "7.1"},
		{ratC(-137, 10), "-13.7"},
		{stepC(1, 19), "1200/19 cents"},
		{ratio(3, 2), "3:2"},
		{ratio(2, 3), "2:3"},
		{ratio(531441, 524288), "531441:524288"},
		{ratC(-100800, 1).withExp(3, rat(53, 1)), "19383245667680019896796723:19342813113834066795298816"},
		{zeroComp().withExp(5, rat(1, 4)), "5^(1/4)"},
		{ratC(6, 5).withExp(5, rat(1, 1000)), "2^(1/1000)*5^(1/1000)"},
		{ratio(3, 2).add(ratC(1, 1)), "2^(-1199/1200)*3"},
		{ratC(1200, 1), "1200"},
		{zeroComp(), "0"},
		{ratio(3, 2).mulInt(5000), "2^-5000*3^5000"},
	} {
		assert.Equal(t, tc.want, stringOracle(tc.c), tc.c.String())
	}

	s, ok := textOracle(ratC(1200, 1), cents.KindRatio, 0)
	assert.True(t, ok)
	assert.Equal(t, "2:1", s)
	assert.Equal(t, "2^-84*3^53", powerText(ratC(-100800, 1).withExp(3, rat(53, 1))))
	assert.Equal(t, "2^1", powerText(ratC(1200, 1)))
	assert.Equal(t, "3^1", powerText(ratio(3, 1)))
	assert.Equal(t, "2^-1*3", powerText(ratio(3, 2)))

	assert.Equal(t, "3/2", scalaOracle(ratio(3, 2), 1))
	assert.Equal(t, "1/1", scalaOracle(zeroComp(), 1))
	assert.Equal(t, "700.0", scalaOracle(ratC(700, 1), 0))
	assert.Equal(t, "3.61505", scalaOracle(ratC(-100800, 1).withExp(3, rat(53, 1)), 5))

	assert.Equal(t, ratC(1, 1000000).String(), gridOracle(5e-7).String())
	assert.Equal(t, ratC(-1, 1000000).String(), gridOracle(-5e-7).String())
	assert.Equal(t, zeroComp().String(), gridOracle(4.9999999e-7).String())
	assert.Equal(t, ratC(1234568, 1000000).String(), gridOracle(1.23456789).String())

	assert.Equal(t, 701.9550008653874, floatOracle(ratio(3, 2)))
	assert.True(t, ratio(3, 2).cmp(ratC(701955000865387, 1000000000000)) > 0)
	assert.True(t, ratio(3, 2).cmp(ratC(701955000865388, 1000000000000)) < 0)
	assert.True(t, ratio(3, 2).mulInt(12).sub(ratC(8400, 1)).equal(ratio(531441, 524288)))
	_, ok = ratioComp(big.NewInt(37), big.NewInt(32))
	assert.False(t, ok)

	r, ok := ratioComp(big.NewInt(111), big.NewInt(74))
	assert.True(t, ok)
	assert.True(t, r.equal(ratio(3, 2)))

	r, ok = ratioComp(big.NewInt(maxInt), big.NewInt(maxInt))
	assert.True(t, ok)
	assert.True(t, r.equal(zeroComp()))
}
