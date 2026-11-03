package pitchclass_test

import (
	"fmt"
	"iter"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests assume a 64-bit int (the project supports only 64-bit systems).
const (
	minInt = math.MinInt
	maxInt = math.MaxInt
)

// The oracle below follows the mathematical definition of the architecture.
// All arithmetic is exact (math/big) and never calls the package under test,
// so it does not depend on its overflow handling.

func bigInt(v int) *big.Int {
	return big.NewInt(int64(v))
}

// modBig returns a mod n, the Euclidean remainder 0...n-1, for n >= 1.
func modBig(a *big.Int, n int) int {
	return int(new(big.Int).Mod(a, bigInt(n)).Int64())
}

// mod returns a mod n for any int a and n >= 1.
func mod(a, n int) int {
	return modBig(bigInt(a), n)
}

// addMod returns (a + b) mod n.
func addMod(a, b, n int) int {
	return modBig(new(big.Int).Add(bigInt(a), bigInt(b)), n)
}

// subMod returns (a - b) mod n.
func subMod(a, b, n int) int {
	return modBig(new(big.Int).Sub(bigInt(a), bigInt(b)), n)
}

// mulMod returns (a * b) mod n.
func mulMod(a, b, n int) int {
	return modBig(new(big.Int).Mul(bigInt(a), bigInt(b)), n)
}

// affineMod returns (m*x + n) mod size.
func affineMod(m, x, n, size int) int {
	v := new(big.Int).Mul(bigInt(m), bigInt(x))
	v.Add(v, bigInt(n))

	return modBig(v, size)
}

// negMod returns (-a) mod n without negating an int.
func negMod(a, n int) int {
	return modBig(new(big.Int).Neg(bigInt(a)), n)
}

// intervalClassOf returns ic(i) = min(i mod n, (-i) mod n).
func intervalClassOf(i, n int) int {
	return min(mod(i, n), negMod(i, n))
}

// gcdInt returns the greatest common divisor of non-negative a and b;
// gcdInt(0, n) = n.
func gcdInt(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// invertibleOracle reports whether m is invertible modulo n:
// gcd(m mod n, n) = 1. For n = 1 every m is invertible.
func invertibleOracle(m, n int) bool {
	g := new(big.Int).GCD(nil, nil, bigInt(mod(m, n)), bigInt(n))

	return g.Cmp(big.NewInt(1)) == 0
}

// inverseOracle returns m^-1 in 0...n-1 for an invertible m; 0 when n = 1.
func inverseOracle(m, n int) int {
	if n == 1 {
		return 0
	}

	r := new(big.Int).ModInverse(bigInt(mod(m, n)), bigInt(n))

	return int(r.Int64())
}

// invertibleList returns the invertible multipliers 0...n-1 in ascending
// order by checking every candidate; only for small n.
func invertibleList(n int) []int {
	out := []int{}
	for m := range n {
		if gcdInt(m, n) == 1 {
			out = append(out, m)
		}
	}

	return out
}

// totientBrute returns phi(n) by checking every candidate; only for small n.
func totientBrute(n int) int {
	return len(invertibleList(n))
}

// fixedClassesBrute returns the numbers z in 0...n-1 with (index - z) mod n
// = z in ascending order by checking every class; only for small n.
func fixedClassesBrute(index, n int) []int {
	out := []int{}
	for z := range n {
		if subMod(index, z, n) == z {
			out = append(out, z)
		}
	}

	return out
}

// fixedClassesFormula returns the solutions of 2z = index (mod n) by the
// formula of P-12, computed exactly; usable for any n.
func fixedClassesFormula(index, n int) []int {
	np := mod(index, n)
	if n%2 == 1 {
		return []int{mulMod(np, inverseOracle(2, n), n)}
	}

	if np%2 == 1 {
		return []int{}
	}

	return []int{np / 2, np/2 + n/2}
}

// opSpec describes the operator x -> m*x + n (mod size) with m and n
// already reduced to 0...size-1. It is the oracle of Operator.
type opSpec struct {
	m, n, size int
}

// newOpSpec reduces m and n modulo size.
func newOpSpec(m, n, size int) opSpec {
	return opSpec{m: mod(m, size), n: mod(n, size), size: size}
}

func identitySpec(size int) opSpec {
	return newOpSpec(1, 0, size)
}

func transpositionSpec(n, size int) opSpec {
	return newOpSpec(1, n, size)
}

func inversionSpec(n, size int) opSpec {
	return newOpSpec(size-1, n, size)
}

func (s opSpec) apply(x int) int {
	return affineMod(s.m, x, s.n, s.size)
}

// compose returns s∘g: first g, then s.
func (s opSpec) compose(g opSpec) opSpec {
	return opSpec{m: mulMod(s.m, g.m, s.size), n: affineMod(s.m, g.n, s.n, s.size), size: s.size}
}

func (s opSpec) inverse() opSpec {
	inv := inverseOracle(s.m, s.size)

	return opSpec{m: inv, n: affineMod(-inv, s.n, 0, s.size), size: s.size}
}

func (s opSpec) isTransposition() bool {
	return s.m == mod(1, s.size)
}

func (s opSpec) isInversion() bool {
	return s.m == mod(s.size-1, s.size)
}

// String is the expected canonical form T<n>M<m> mod <N>.
func (s opSpec) String() string {
	return fmt.Sprintf("T%dM%d mod %d", s.n, s.m, s.size)
}

// compareSpec is the expected Operator.Compare: by N, then m, then n.
func compareSpec(a, b opSpec) int {
	switch {
	case a.size != b.size:
		return sign(a.size - b.size)
	case a.m != b.m:
		return sign(a.m - b.m)
	default:
		return sign(a.n - b.n)
	}
}

// compareClassSpec is the expected Class.Compare of valid classes (k1, n1)
// and (k2, n2): by N, then by number. Inputs are small, no overflow.
func compareClassSpec(k1, n1, k2, n2 int) int {
	if n1 != n2 {
		return sign(n1 - n2)
	}

	return sign(k1 - k2)
}

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

// operatorSpecs returns all operators of size in the order of Operators:
// by ascending invertible m, then n = 0...size-1. Only for small sizes.
func operatorSpecs(size int) []opSpec {
	var out []opSpec
	for _, m := range invertibleList(size) {
		for n := range size {
			out = append(out, opSpec{m: m, n: n, size: size})
		}
	}

	return out
}

// transpositionSpecs returns T0...T(size-1).
func transpositionSpecs(size int) []opSpec {
	out := make([]opSpec, 0, size)
	for n := range size {
		out = append(out, transpositionSpec(n, size))
	}

	return out
}

// tiSpecs returns T0...T(size-1), then I0...I(size-1); for size <= 2 only the
// transpositions, because the inversions coincide with them.
func tiSpecs(size int) []opSpec {
	out := transpositionSpecs(size)
	if size <= 2 {
		return out
	}

	for n := range size {
		out = append(out, inversionSpec(n, size))
	}

	return out
}

// ints returns lo...hi.
func ints(lo, hi int) []int {
	out := make([]int, 0, hi-lo+1)
	for v := lo; v <= hi; v++ {
		out = append(out, v)
	}

	return out
}

// edgeInts are integer inputs at the limits of int and around zero.
func edgeInts() []int {
	return []int{minInt, minInt + 1, -13, -1, 0, 1, 2, 13, maxInt - 1, maxInt}
}

// newClass creates New(k, n) and stops the test on an error.
func newClass(t testing.TB, k, n int) pitchclass.Class {
	t.Helper()

	c, err := pitchclass.New(k, n)
	require.NoError(t, err, "New(%d, %d)", k, n)

	return c
}

// classesOf returns New(k, n) for k = 0...n-1.
func classesOf(t testing.TB, n int) []pitchclass.Class {
	t.Helper()

	out := make([]pitchclass.Class, 0, n)
	for k := range n {
		out = append(out, newClass(t, k, n))
	}

	return out
}

// classesOfAll returns classesOf for each modulus in turn.
func classesOfAll(t testing.TB, moduli ...int) []pitchclass.Class {
	t.Helper()

	var out []pitchclass.Class
	for _, n := range moduli {
		out = append(out, classesOf(t, n)...)
	}

	return out
}

// newOperator creates NewOperator(s.m, s.n, s.size) and stops the test on an
// error.
func newOperator(t testing.TB, s opSpec) pitchclass.Operator {
	t.Helper()

	f, err := pitchclass.NewOperator(s.m, s.n, s.size)
	require.NoError(t, err, "NewOperator(%d, %d, %d)", s.m, s.n, s.size)

	return f
}

// operatorsOf returns all operators of size, created by NewOperator, in the
// order of operatorSpecs, with their specs.
func operatorsOf(t testing.TB, size int) ([]pitchclass.Operator, []opSpec) {
	t.Helper()

	specs := operatorSpecs(size)
	ops := make([]pitchclass.Operator, 0, len(specs))
	for _, s := range specs {
		ops = append(ops, newOperator(t, s))
	}

	return ops, specs
}

// assertClass checks that got is the valid class "number of modulus".
func assertClass(t testing.TB, number, modulus int, got pitchclass.Class, desc string) bool {
	t.Helper()

	ok := assert.True(t, got.IsValid(), "%s: IsValid", desc)
	ok = assert.Equal(t, number, got.Number(), "%s: Number", desc) && ok

	return assert.Equal(t, modulus, got.Modulus(), "%s: Modulus", desc) && ok
}

// assertOperator checks that got is the valid operator described by want.
func assertOperator(t testing.TB, want opSpec, got pitchclass.Operator, desc string) bool {
	t.Helper()

	ok := assert.True(t, got.IsValid(), "%s: IsValid", desc)
	ok = assert.Equal(t, want.size, got.Modulus(), "%s: Modulus", desc) && ok
	ok = assert.Equal(t, want.m, got.Multiplier(), "%s: Multiplier", desc) && ok

	return assert.Equal(t, want.n, got.Index(), "%s: Index", desc) && ok
}

// numbersOf returns the numbers of the classes.
func numbersOf(cs []pitchclass.Class) []int {
	out := make([]int, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Number())
	}

	return out
}

// specsOf describes the operators by their properties.
func specsOf(fs []pitchclass.Operator) []opSpec {
	out := make([]opSpec, 0, len(fs))
	for _, f := range fs {
		out = append(out, opSpec{m: f.Multiplier(), n: f.Index(), size: f.Modulus()})
	}

	return out
}

// take returns at most limit first elements of seq and stops the iteration.
// It never ranges over the whole sequence, so it terminates for sequences
// of any length.
func take[T any](seq iter.Seq[T], limit int) []T {
	out := make([]T, 0, limit)
	if limit <= 0 {
		return out
	}

	for v := range seq {
		out = append(out, v)
		if len(out) == limit {
			break
		}
	}

	return out
}

// callsAfterStop calls seq directly with a yield that returns false on its
// stop-th call and returns how many times yield was called after that.
// A conforming iterator returns 0.
func callsAfterStop[T any](seq iter.Seq[T], stop int) int {
	calls, extra := 0, 0
	seq(func(T) bool {
		calls++
		if calls > stop {
			extra++

			return false
		}

		return calls < stop
	})

	return extra
}

// assertSeq checks the common iterator contract for a finite sequence: the
// function is not nil, it yields want, a second range yields want again, and
// stopping after k elements (k = 1...3) stops the iteration.
func assertSeq[T any](t testing.TB, seq iter.Seq[T], want []T, desc string) bool {
	t.Helper()

	if !assert.NotNil(t, seq, "%s: iterator", desc) {
		return false
	}

	ok := assert.Equal(t, want, take(seq, len(want)+1), "%s: elements", desc)
	ok = assert.Equal(t, want, take(seq, len(want)+1), "%s: second range", desc) && ok

	for k := 1; k <= min(3, len(want)); k++ {
		ok = assert.Equal(t, want[:k], take(seq, k), "%s: first %d", desc, k) && ok
		ok = assert.Zero(t, callsAfterStop(seq, k), "%s: yield called after stop at %d", desc, k) && ok
	}

	return ok
}

// assertEmptySeq checks that seq is a non-nil iterator that yields nothing.
func assertEmptySeq[T any](t testing.TB, seq iter.Seq[T], desc string) bool {
	t.Helper()

	if !assert.NotNil(t, seq, "%s: iterator must not be nil", desc) {
		return false
	}

	return assert.Empty(t, take(seq, 1), "%s: elements", desc)
}

// allErrors returns the seven sentinel errors of the package.
func allErrors() []error {
	return []error{
		pitchclass.ErrInvalidClass,
		pitchclass.ErrInvalidModulus,
		pitchclass.ErrModulusMismatch,
		pitchclass.ErrNotInvertible,
		pitchclass.ErrInvalidValue,
		pitchclass.ErrNotation,
		pitchclass.ErrOverflow,
	}
}

// assertOnlyError checks that err matches want with errors.Is and none of
// the other six sentinel errors. The text of errors is not compared.
func assertOnlyError(t testing.TB, err, want error, desc string) bool {
	t.Helper()

	ok := assert.ErrorIs(t, err, want, desc)
	for _, other := range allErrors() {
		if other == want {
			continue
		}

		ok = assert.NotErrorIs(t, err, other, desc) && ok
	}

	return ok
}

// panicValue calls f and reports whether it panicked and with what value.
func panicValue(f func()) (recovered any, panicked bool) {
	defer func() {
		recovered = recover()
	}()

	panicked = true
	f()
	panicked = false

	return nil, panicked
}

// assertPanicsWithErrorIs checks that f panics with an error value that
// matches want with errors.Is and none of the other sentinel errors.
// assert.PanicsWithError is not used: it compares the text.
func assertPanicsWithErrorIs(t testing.TB, want error, f func(), desc string) bool {
	t.Helper()

	v, panicked := panicValue(f)
	if !assert.True(t, panicked, "expected a panic: %s", desc) {
		return false
	}

	err, ok := v.(error)
	if !assert.True(t, ok, "panic value %v (%T) is not an error: %s", v, v, desc) {
		return false
	}

	return assertOnlyError(t, err, want, desc)
}

// styles returns the four styles of the package.
func styles() []pitchclass.Style {
	return []pitchclass.Style{pitchclass.Decimal, pitchclass.TE, pitchclass.LowerTE, pitchclass.AB}
}

// letters returns the symbols of 10 and 11 in style st, by the table of the
// architecture; Decimal has "10" and "11".
func letters(st pitchclass.Style) (ten, eleven string) {
	switch st {
	case pitchclass.TE:
		return "T", "E"
	case pitchclass.LowerTE:
		return "t", "e"
	case pitchclass.AB:
		return "A", "B"
	default:
		return "10", "11"
	}
}

// notationOf returns the expected literary notation of class number k in a
// valid style at modulus n <= 12 (or any n for Decimal).
func notationOf(k int, st pitchclass.Style) string {
	ten, eleven := letters(st)

	switch k {
	case 10:
		return ten
	case 11:
		return eleven
	default:
		return fmt.Sprint(k)
	}
}

// result holds the value and the error of an operation with an error.
// Tests use res(op(...)).get(t) instead of the Must variant where the
// operation is expected to succeed: an unexpected error then fails only the
// current test instead of panicking the whole test binary. The Must variants
// themselves are checked in must_test.go and next to their operations.
type result[T any] struct {
	v   T
	err error
}

// res captures the two results of an operation with an error.
func res[T any](v T, err error) result[T] {
	return result[T]{v: v, err: err}
}

// get returns the value and stops the test if the operation failed.
func (r result[T]) get(t testing.TB) T {
	t.Helper()

	require.NoError(t, r.err)

	return r.v
}
