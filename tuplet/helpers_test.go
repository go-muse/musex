package tuplet_test

import (
	"errors"
	"math"
	"math/big"
	"slices"
	"strconv"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// namedError pairs a sentinel error with its exported name for diagnostics.
type namedError struct {
	name string
	err  error
}

// tupletSentinels lists the seven sentinel errors of the package
// (architecture 4.1, 4.2; requirements E-1...E-7).
var tupletSentinels = []namedError{
	{"ErrInvalidCount", tuplet.ErrInvalidCount},
	{"ErrUnset", tuplet.ErrUnset},
	{"ErrNoDefault", tuplet.ErrNoDefault},
	{"ErrInvalidText", tuplet.ErrInvalidText},
	{"ErrOutOfRange", tuplet.ErrOutOfRange},
	{"ErrInvalidArgument", tuplet.ErrInvalidArgument},
	{"ErrNotRepresentable", tuplet.ErrNotRepresentable},
}

// fractionSentinels lists the six sentinel errors of common/fraction.
var fractionSentinels = []namedError{
	{"ErrInvalid", fraction.ErrInvalid},
	{"ErrZeroDenominator", fraction.ErrZeroDenominator},
	{"ErrDivisionByZero", fraction.ErrDivisionByZero},
	{"ErrNotRepresentable", fraction.ErrNotRepresentable},
	{"ErrOutOfRange", fraction.ErrOutOfRange},
	{"ErrInvalidText", fraction.ErrInvalidText},
}

// assertErr checks the error classification of architecture 4.2: err matches
// exactly one tuplet sentinel, want, and among the common/fraction sentinels
// exactly the listed causes. With no causes, err must not match any
// common/fraction sentinel.
func assertErr(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := true
	for _, s := range tupletSentinels {
		expected := s.err == want
		if errors.Is(err, s.err) != expected {
			ok = false
			t.Errorf("errors.Is(err, tuplet.%s) = %t, want %t; err: %s", s.name, !expected, expected, err)
		}
	}

	for _, s := range fractionSentinels {
		expected := slices.Contains(causes, s.err)
		if errors.Is(err, s.err) != expected {
			ok = false
			t.Errorf("errors.Is(err, fraction.%s) = %t, want %t; err: %s", s.name, !expected, expected, err)
		}
	}

	return ok
}

// capturePanic runs f and reports the recovered panic value, if any.
func capturePanic(f func()) (value any, panicked bool) {
	panicked = true

	defer func() {
		if panicked {
			value = recover()
		}
	}()

	f()

	panicked = false

	return nil, false
}

// assertPanicErr checks that f panics with an error value classified as in
// assertErr and returns that error.
func assertPanicErr(t *testing.T, f func(), want error, causes ...error) error {
	t.Helper()

	value, panicked := capturePanic(f)
	if !assert.True(t, panicked, "expected a panic") {
		return nil
	}

	err, isErr := value.(error)
	if !assert.True(t, isErr, "panic value is %T, want error", value) {
		return nil
	}

	assertErr(t, err, want, causes...)

	return err
}

// assertNoPanic checks that f returns without panicking.
func assertNoPanic(t *testing.T, f func()) bool {
	t.Helper()

	value, panicked := capturePanic(f)

	return assert.False(t, panicked, "unexpected panic: %v", value)
}

// frac builds a fraction with the exact source pair num/den.
func frac(num, den int) fraction.Fraction {
	return fraction.MustNew(num, den)
}

// bigFrac builds a fraction with an exact source pair given as big integers.
func bigFrac(t *testing.T, num, den *big.Int) fraction.Fraction {
	t.Helper()

	f, err := fraction.FromBigInt(num, den)
	require.NoError(t, err)

	return f
}

// pow2 returns 2**n as a new big integer.
func pow2(n uint) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), n)
}

// bigOf returns v as a new big integer.
func bigOf(v int) *big.Int {
	return big.NewInt(int64(v))
}

// maxComponentFrac returns fraction.MaxComponent/1.
func maxComponentFrac(t *testing.T) fraction.Fraction {
	t.Helper()

	return bigFrac(t, fraction.MaxComponentBig(), big.NewInt(1))
}

// ratOf returns the value of a valid fraction as a big.Rat oracle.
func ratOf(t *testing.T, f fraction.Fraction) *big.Rat {
	t.Helper()

	num, den, ok := f.BigPair()
	require.True(t, ok, "fraction %s is not valid", f)

	return new(big.Rat).SetFrac(num, den)
}

// ratInts returns num/den as a big.Rat oracle.
func ratInts(num, den int) *big.Rat {
	return new(big.Rat).SetFrac(bigOf(num), bigOf(den))
}

// assertPair checks the exact source pair of f.
func assertPair(t *testing.T, f fraction.Fraction, num, den int) bool {
	t.Helper()

	gotNum, gotDen, ok := f.Pair()
	if !assert.True(t, ok, "fraction %s has no int pair", f) {
		return false
	}

	return assert.Equal(t, [2]int{num, den}, [2]int{gotNum, gotDen}, "source pair of %s", f)
}

// assertBigPair checks the exact source pair of f against big integers.
func assertBigPair(t *testing.T, f fraction.Fraction, num, den *big.Int) bool {
	t.Helper()

	gotNum, gotDen, ok := f.BigPair()
	if !assert.True(t, ok, "fraction %s is not valid", f) {
		return false
	}

	return assert.Zero(t, gotNum.Cmp(num), "numerator of %s, want %s", f, num) &&
		assert.Zero(t, gotDen.Cmp(den), "denominator of %s, want %s", f, den)
}

// assertReducedValue checks that f is the fully reduced form of want: equal
// value, positive denominator, and coprime components (architecture 4.9).
func assertReducedValue(t *testing.T, f fraction.Fraction, want *big.Rat) bool {
	t.Helper()

	return assertBigPair(t, f, want.Num(), want.Denom())
}

// assertTuplet checks that got is the set tuplet actual:normal.
func assertTuplet(t *testing.T, got tuplet.Tuplet, actual, normal int) bool {
	t.Helper()

	if !assert.True(t, got.IsSet(), "tuplet is not set") {
		return false
	}

	a, n, err := got.Pair()
	if !assert.NoError(t, err) {
		return false
	}

	return assert.Equal(t, [2]int{actual, normal}, [2]int{a, n}, "pair of %s", got)
}

// isSeriesMember reports whether count is a member of the series of the
// division by walking the series: 2, 4, 8, ... for Binary and 3, 6, 12, ...
// for Ternary (requirement IP-8).
func isSeriesMember(count int, division tuplet.Division) bool {
	start, ok := seriesStart(division)
	if !ok {
		return false
	}

	for v := start; v <= count; v *= 2 {
		if v == count {
			return true
		}

		if v > math.MaxInt/2 {
			break
		}
	}

	return false
}

// seriesStart returns the first member of the series of the division.
func seriesStart(division tuplet.Division) (int, bool) {
	switch division {
	case tuplet.Binary:
		return 2, true
	case tuplet.Ternary:
		return 3, true
	default:
		return 0, false
	}
}

// oracleDefault returns the default normal count by the definition of rule
// P-5: the largest member of the series of the division that is less than
// actual; for Ternary and actual 2 the default is 3; there is no default for
// actual 1 or for a member of the series. It walks the series instead of
// reproducing any closed formula.
func oracleDefault(actual int, division tuplet.Division) (int, bool) {
	start, ok := seriesStart(division)
	if !ok || actual < 1 {
		return 0, false
	}

	if division == tuplet.Ternary && actual == 2 {
		return 3, true
	}

	best, found := 0, false

	for v := start; v <= actual; v *= 2 {
		if v == actual {
			return 0, false
		}

		best, found = v, true

		if v > math.MaxInt/2 {
			break
		}
	}

	return best, found
}

// oracleDivision returns the division whose series contains count.
func oracleDivision(count int) (tuplet.Division, bool) {
	for _, d := range []tuplet.Division{tuplet.Binary, tuplet.Ternary} {
		if isSeriesMember(count, d) {
			return d, true
		}
	}

	return 0, false
}

// itoa formats v for subtest names.
func itoa(v int) string {
	return strconv.Itoa(v)
}

// sign returns -1, 0, or +1 for the sign of v.
func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	default:
		return 0
	}
}

// oracleCompare is the total order of architecture 4.7 computed with big
// rationals: unset first; then by coefficient normal/actual ascending; then
// by actual; then by normal.
func oracleCompare(a, b [2]int, aSet, bSet bool) int {
	switch {
	case !aSet && !bSet:
		return 0
	case !aSet:
		return -1
	case !bSet:
		return 1
	}

	if c := ratInts(a[1], a[0]).Cmp(ratInts(b[1], b[0])); c != 0 {
		return c
	}

	if a[0] != b[0] {
		return sign(a[0] - b[0])
	}

	return sign(a[1] - b[1])
}

// sampleTuplets returns a varied set of set tuplets, including counts near
// math.MaxInt, for property checks.
func sampleTuplets() []tuplet.Tuplet {
	pairs := [][2]int{
		{1, 1}, {2, 3}, {3, 2}, {4, 3}, {5, 4}, {6, 4}, {7, 4}, {8, 6},
		{9, 8}, {10, 8}, {5, 3}, {7, 8}, {7, 11}, {4, 4}, {1, 7}, {12, 1},
		{14, 8}, {6, 6},
		{math.MaxInt, 1}, {1, math.MaxInt}, {math.MaxInt, math.MaxInt},
		{math.MaxInt, math.MaxInt - 1}, {math.MaxInt - 1, math.MaxInt},
		{1 << 62, 3 << 61},
	}

	out := make([]tuplet.Tuplet, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, tuplet.MustNew(p[0], p[1]))
	}

	return out
}
