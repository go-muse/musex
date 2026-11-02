package alteration_test

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The guaranteed range of the contract, defined independently of the package
// under test: |value| <= oracleMaxMagnitude and the denominator in lowest
// terms <= oracleMaxDenominator. TestRangeConstants checks that the package
// constants have these values.
const (
	oracleMaxMagnitude   = 1 << 24
	oracleMaxDenominator = 1 << 16
)

// Short names for the edges of the range used in the boundary tables:
// M is the largest value, m = -M the smallest, Q the largest denominator.
const (
	bigM = oracleMaxMagnitude
	bigQ = oracleMaxDenominator
)

// frac is a pair (numerator, denominator) given as test input. It is not
// necessarily in lowest terms; the oracle normalizes it with math/big.
type frac struct {
	num, den int
}

// String renders the pair for test names and messages. Parentheses are used
// instead of a slash so that subtest names stay single-level.
func (f frac) String() string {
	return fmt.Sprintf("(%d,%d)", f.num, f.den)
}

// rat returns the exact rational value of the pair; den must not be zero.
func (f frac) rat() *big.Rat {
	return ratOf(f.num, f.den)
}

// The proof set S of the acceptance criteria in ascending order:
// -7/2 < double flat < flat < -1/2 < natural < 1/3 < 1/2 < sharp < 3/2 <
// double sharp < +3. A fresh slice is returned on every call.
func setS() []frac {
	return []frac{
		{-7, 2}, {-2, 1}, {-1, 1}, {-1, 2}, {0, 1}, {1, 3}, {1, 2}, {1, 1}, {3, 2}, {2, 1}, {3, 1},
	}
}

// Edges of the guaranteed range used by the boundary and zero-value tests.
// They are functions, not package variables: no test state is shared.

// edgeTop returns M.
func edgeTop() frac { return frac{bigM, 1} }

// edgeBottom returns m = -M.
func edgeBottom() frac { return frac{-bigM, 1} }

// edgeInvQ returns 1/Q.
func edgeInvQ() frac { return frac{1, bigQ} }

// edgeNegInvQ returns -1/Q.
func edgeNegInvQ() frac { return frac{-1, bigQ} }

// edgeNearTop returns M - 1/Q, with the denominator Q.
func edgeNearTop() frac { return frac{bigM*bigQ - 1, bigQ} }

// edgeNearBottom returns m + 1/Q, with the denominator Q.
func edgeNearBottom() frac { return frac{1 - bigM*bigQ, bigQ} }

// ratOf returns num/den as an exact rational; den must not be zero.
// math.MinInt is handled exactly: big.Int holds every int64.
func ratOf(num, den int) *big.Rat {
	return new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
}

func ratInt(n int) *big.Rat {
	return new(big.Rat).SetInt64(int64(n))
}

func ratAdd(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Add(a, b)
}

func ratSub(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Sub(a, b)
}

func ratMul(a *big.Rat, k int) *big.Rat {
	return new(big.Rat).Mul(a, ratInt(k))
}

// inRange reports whether the exact value r lies in the guaranteed range.
func inRange(r *big.Rat) bool {
	limit := ratInt(oracleMaxMagnitude)
	magnitude := new(big.Rat).Abs(r)

	return magnitude.Cmp(limit) <= 0 && r.Denom().Cmp(big.NewInt(oracleMaxDenominator)) <= 0
}

// expectedFrac returns the lowest-terms pair of an in-range rational.
func expectedFrac(t *testing.T, r *big.Rat) frac {
	t.Helper()

	require.True(t, r.Num().IsInt64() && r.Denom().IsInt64(), "oracle value %s fits int64", r.RatString())

	return frac{int(r.Num().Int64()), int(r.Denom().Int64())}
}

// fractionOf returns the exact value reported by the package.
func fractionOf(a alteration.Alteration) frac {
	num, den := a.Fraction()

	return frac{num, den}
}

// value creates the value of an input pair with the fallible constructor and
// stops the test if it cannot be created.
func value(t *testing.T, f frac) alteration.Alteration {
	t.Helper()

	a, err := alteration.FromFraction(f.num, f.den)
	require.NoError(t, err, "FromFraction%s", f)

	return a
}

// values creates the values of the input pairs in order.
func values(t *testing.T, fs []frac) []alteration.Alteration {
	t.Helper()

	out := make([]alteration.Alteration, len(fs))
	for i, f := range fs {
		out[i] = value(t, f)
	}

	return out
}

// assertSame checks that got == want and reports both values and their exact
// fractions when they differ. assert.Equal is not used for values: it would
// print the unexported fields of the type.
func assertSame(t *testing.T, want, got alteration.Alteration, msgAndArgs ...any) bool {
	t.Helper()

	if got == want {
		return true
	}

	return assert.Fail(t, fmt.Sprintf("values differ: want %v %s, got %v %s", want, fractionOf(want), got, fractionOf(got)), msgAndArgs...)
}

// assertExact checks that got has the exact value want, which must lie in
// the guaranteed range.
func assertExact(t *testing.T, want *big.Rat, got alteration.Alteration, msgAndArgs ...any) {
	t.Helper()

	require.True(t, inRange(want), "test expectation %s lies in the range", want.RatString())
	assert.Equal(t, expectedFrac(t, want), fractionOf(got), msgAndArgs...)
}

// assertOutcome checks the result of a fallible operation against the exact
// value: in range — no error and the exact value; out of range —
// ErrOutOfRange, not ErrZeroDenominator, and the zero value. want == nil
// means a zero denominator: ErrZeroDenominator, not ErrOutOfRange, and the
// zero value.
func assertOutcome(t *testing.T, want *big.Rat, got alteration.Alteration, err error, msgAndArgs ...any) {
	t.Helper()

	switch {
	case want == nil:
		require.ErrorIs(t, err, alteration.ErrZeroDenominator, msgAndArgs...)
		assert.NotErrorIs(t, err, alteration.ErrOutOfRange, msgAndArgs...)
		assert.True(t, got == alteration.Alteration{}, "the zero value is returned with the error, got %v", got)
	case inRange(want):
		require.NoError(t, err, msgAndArgs...)
		assert.Equal(t, expectedFrac(t, want), fractionOf(got), msgAndArgs...)
	default:
		require.ErrorIs(t, err, alteration.ErrOutOfRange, msgAndArgs...)
		assert.NotErrorIs(t, err, alteration.ErrZeroDenominator, msgAndArgs...)
		assert.True(t, got == alteration.Alteration{}, "the zero value is returned with the error, got %v", got)
	}
}

// recoverPanic calls f and reports whether it panicked and with what value.
func recoverPanic(f func()) (recovered any, panicked bool) {
	defer func() {
		recovered = recover()
	}()

	panicked = true
	f()
	panicked = false

	return nil, panicked
}

// assertPanicsWithSentinel checks that f panics with an error value for
// which errors.Is reports sentinel and not the other sentinel error.
// assert.PanicsWithError is not used: it compares the error text, which is
// not part of the contract.
func assertPanicsWithSentinel(t *testing.T, sentinel error, f func(), msgAndArgs ...any) {
	t.Helper()

	recovered, panicked := recoverPanic(f)
	require.True(t, panicked, msgAndArgs...)

	err, ok := recovered.(error)
	require.True(t, ok, "the panic value is an error, got %#v", recovered)
	assert.ErrorIs(t, err, sentinel, msgAndArgs...)

	for _, other := range sentinels() {
		if other != sentinel {
			assert.NotErrorIs(t, err, other, msgAndArgs...)
		}
	}
}

// sentinels returns the sentinel errors of the contract.
func sentinels() []error {
	return []error{alteration.ErrZeroDenominator, alteration.ErrOutOfRange}
}

// outcome is the observable result of a fallible call or of its Must
// variant: the returned value and the returned error or the error the Must
// variant panicked with.
type outcome struct {
	value alteration.Alteration
	err   error
}

// errNonErrorPanic marks a Must panic whose value is not an error.
var errNonErrorPanic = errors.New("panic value is not an error")

// call records the outcome of a fallible operation.
func call(f func() (alteration.Alteration, error)) outcome {
	v, err := f()

	return outcome{value: v, err: err}
}

// callMust records the outcome of a Must variant, turning a panic into the
// error it carries.
func callMust(f func() alteration.Alteration) outcome {
	var v alteration.Alteration

	recovered, panicked := recoverPanic(func() { v = f() })
	if !panicked {
		return outcome{value: v}
	}

	if err, ok := recovered.(error); ok {
		return outcome{err: err}
	}

	return outcome{err: fmt.Errorf("%w: %#v", errNonErrorPanic, recovered)}
}

// assertSameOutcome checks that two outcomes agree: both succeed with equal
// values or both fail with the same sentinel error.
func assertSameOutcome(t *testing.T, want, got outcome, msgAndArgs ...any) {
	t.Helper()

	if want.err == nil {
		require.NoError(t, got.err, msgAndArgs...)
		assert.True(t, want.value == got.value, "values: want %v, got %v", want.value, got.value)
		assert.Equal(t, fractionOf(want.value), fractionOf(got.value), msgAndArgs...)

		return
	}

	require.Error(t, got.err, msgAndArgs...)
	assert.NotErrorIs(t, got.err, errNonErrorPanic, msgAndArgs...)

	for _, s := range sentinels() {
		assert.Equal(t, errors.Is(want.err, s), errors.Is(got.err, s), "errors.Is(err, %v): want %v, got %v", s, want.err, got.err)
	}
}

// fallibleOp is one call of an operation that may return an error, with its
// value inputs given as pairs, the call itself, its Must variant and the
// exact expected result computed with math/big.
type fallibleOp struct {
	name   string
	inputs []frac
	call   func(in []alteration.Alteration) (alteration.Alteration, error)
	must   func(in []alteration.Alteration) alteration.Alteration
	want   *big.Rat // exact result; nil for a zero denominator
}

func opFromInt(n int) fallibleOp {
	return fallibleOp{
		name: fmt.Sprintf("FromInt(%d)", n),
		call: func([]alteration.Alteration) (alteration.Alteration, error) { return alteration.FromInt(n) },
		must: func([]alteration.Alteration) alteration.Alteration { return alteration.MustFromInt(n) },
		want: ratInt(n),
	}
}

func opFromFraction(num, den int) fallibleOp {
	var want *big.Rat
	if den != 0 {
		want = ratOf(num, den)
	}

	return fallibleOp{
		name: fmt.Sprintf("FromFraction(%d,%d)", num, den),
		call: func([]alteration.Alteration) (alteration.Alteration, error) {
			return alteration.FromFraction(num, den)
		},
		must: func([]alteration.Alteration) alteration.Alteration { return alteration.MustFromFraction(num, den) },
		want: want,
	}
}

func opRaise(a frac) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.Raise()", a),
		inputs: []frac{a},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].Raise() },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustRaise() },
		want:   ratAdd(a.rat(), ratInt(1)),
	}
}

func opLower(a frac) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.Lower()", a),
		inputs: []frac{a},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].Lower() },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustLower() },
		want:   ratSub(a.rat(), ratInt(1)),
	}
}

func opRaiseBy(a frac, n int) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.RaiseBy(%d)", a, n),
		inputs: []frac{a},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].RaiseBy(n) },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustRaiseBy(n) },
		want:   ratAdd(a.rat(), ratInt(n)),
	}
}

func opLowerBy(a frac, n int) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.LowerBy(%d)", a, n),
		inputs: []frac{a},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].LowerBy(n) },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustLowerBy(n) },
		want:   ratSub(a.rat(), ratInt(n)),
	}
}

func opAdd(a, d frac) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.Add%s", a, d),
		inputs: []frac{a, d},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].Add(in[1]) },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustAdd(in[1]) },
		want:   ratAdd(a.rat(), d.rat()),
	}
}

func opSub(a, d frac) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.Sub%s", a, d),
		inputs: []frac{a, d},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].Sub(in[1]) },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustSub(in[1]) },
		want:   ratSub(a.rat(), d.rat()),
	}
}

func opMul(a frac, k int) fallibleOp {
	return fallibleOp{
		name:   fmt.Sprintf("%s.Mul(%d)", a, k),
		inputs: []frac{a},
		call:   func(in []alteration.Alteration) (alteration.Alteration, error) { return in[0].Mul(k) },
		must:   func(in []alteration.Alteration) alteration.Alteration { return in[0].MustMul(k) },
		want:   ratMul(a.rat(), k),
	}
}

// checkFallible runs one fallible operation and its Must variant:
//   - the result agrees with the exact value (K-3...K-14);
//   - the Must variant panics with the same sentinel error exactly when the
//     operation fails and otherwise returns the same value (K-21);
//   - the value inputs are unchanged after the call, the error and the
//     panic (K-17).
func checkFallible(t *testing.T, op fallibleOp) {
	t.Helper()

	in := values(t, op.inputs)
	before := slices.Clone(in)

	got, err := op.call(in)
	assertOutcome(t, op.want, got, err, op.name)
	assertInputsUnchanged(t, op.inputs, before, in, op.name)

	var mustGot alteration.Alteration

	mustRun := func() { mustGot = op.must(in) }

	switch {
	case op.want == nil:
		assertPanicsWithSentinel(t, alteration.ErrZeroDenominator, mustRun, "Must variant of %s", op.name)
	case !inRange(op.want):
		assertPanicsWithSentinel(t, alteration.ErrOutOfRange, mustRun, "Must variant of %s", op.name)
	default:
		require.NotPanics(t, mustRun, "Must variant of %s", op.name)
		assert.True(t, got == mustGot, "Must variant of %s: want %v, got %v", op.name, got, mustGot)
		assert.Equal(t, fractionOf(got), fractionOf(mustGot), "Must variant of %s", op.name)
	}

	assertInputsUnchanged(t, op.inputs, before, in, "after the Must variant of "+op.name)
}

// assertInputsUnchanged checks that every input value equals its copy taken
// before the call and still has the exact value of its pair (K-17).
func assertInputsUnchanged(t *testing.T, pairs []frac, before, after []alteration.Alteration, msg string) {
	t.Helper()

	for i := range pairs {
		assert.True(t, before[i] == after[i], "%s: input %d changed from %v to %v", msg, i, before[i], after[i])
		assert.True(t, before[i].Equal(after[i]), "%s: input %d", msg, i)
		assert.Equal(t, expectedFrac(t, pairs[i].rat()), fractionOf(after[i]), "%s: input %d", msg, i)
	}
}

// runFallible runs every operation as a named parallel subtest.
func runFallible(t *testing.T, ops []fallibleOp) {
	t.Helper()

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			t.Parallel()
			checkFallible(t, op)
		})
	}
}

// sign returns -1, 0 or +1 by the sign of n.
func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

// gcd returns the greatest common divisor of |a| and |b| for small values.
func gcd(a, b int) int {
	a, b = max(a, -a), max(b, -b)
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// contractString renders an exact rational in the fixed String form of the
// contract: "0"; an integer with a sign ("+3", "-1"); a fraction with a sign,
// the magnitude of the numerator, "/" and the denominator ("+1/2", "-7/2").
func contractString(r *big.Rat) string {
	if r.Sign() == 0 {
		return "0"
	}

	signText := "+"
	if r.Sign() < 0 {
		signText = "-"
	}

	magnitude := new(big.Int).Abs(r.Num())
	if r.IsInt() {
		return signText + magnitude.String()
	}

	return signText + magnitude.String() + "/" + r.Denom().String()
}
