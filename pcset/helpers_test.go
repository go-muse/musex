package pcset_test

import (
	"iter"
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	minInt = math.MinInt
	maxInt = math.MaxInt
)

// allSentinelErrors returns all 11 exported sentinel errors of package pcset.
func allSentinelErrors() []error {
	return []error{
		pcset.ErrInvalidModulus,
		pcset.ErrModulusTooLarge,
		pcset.ErrInvalidClass,
		pcset.ErrInvalidCardinality,
		pcset.ErrInvalidIntervalClass,
		pcset.ErrModulusMismatch,
		pcset.ErrNotModulus12,
		pcset.ErrInvalidValue,
		pcset.ErrInvalidAlgorithm,
		pcset.ErrNotation,
		pcset.ErrOverflow,
	}
}

// requireSentinelError verifies that err matches target via errors.Is and
// does not match any of the other 10 sentinel errors of package pcset.
func requireSentinelError(t testing.TB, err, target error) {
	t.Helper()

	require.ErrorIs(t, err, target)
	for _, other := range allSentinelErrors() {
		if other == target {
			continue
		}

		assert.NotErrorIs(t, err, other)
	}
}

// panicValue executes fn and captures any panic value.
func panicValue(fn func()) (recovered any, panicked bool) {
	defer func() {
		recovered = recover()
	}()

	panicked = true
	fn()

	return nil, false
}

// requirePanicsWithSentinel verifies that fn panics with an error value that
// matches target via requireSentinelError.
func requirePanicsWithSentinel(t testing.TB, fn func(), target error) {
	t.Helper()

	v, panicked := panicValue(fn)
	require.True(t, panicked, "expected panic with %v", target)

	err, ok := v.(error)
	require.True(t, ok, "panic value %v (%T) does not implement error", v, v)
	requireSentinelError(t, err, target)
}

type result[T any] struct {
	v   T
	err error
}

func res[T any](v T, err error) result[T] {
	return result[T]{v: v, err: err}
}

func (r result[T]) get(t testing.TB) T {
	t.Helper()

	require.NoError(t, r.err)

	return r.v
}

func mustClass(t testing.TB, number, modulus int) pitchclass.Class {
	t.Helper()

	c, err := pitchclass.New(number, modulus)
	require.NoError(t, err)

	return c
}

func mustOp(t testing.TB, m, n, modulus int) pitchclass.Operator {
	t.Helper()

	op, err := pitchclass.NewOperator(m, n, modulus)
	require.NoError(t, err)

	return op
}

func numbersOfClasses(classes []pitchclass.Class) []int {
	out := make([]int, len(classes))
	for i, c := range classes {
		out[i] = c.Number()
	}

	return out
}

// assertSet checks all structural properties of a valid Set against the expected
// modulus and strictly ascending slice of element numbers.
func assertSet(t testing.TB, got pcset.Set, wantModulus int, wantNums []int) {
	t.Helper()

	if wantNums == nil {
		wantNums = []int{}
	}

	require.True(t, got.IsValid(), "IsValid")
	assert.Equal(t, wantModulus, got.Modulus(), "Modulus")
	assert.Equal(t, len(wantNums), got.Cardinality(), "Cardinality")
	assert.Equal(t, len(wantNums) == 0, got.IsEmpty(), "IsEmpty")
	assert.Equal(t, len(wantNums) == wantModulus, got.IsUniverse(), "IsUniverse")

	gotNums := got.Numbers()
	require.NotNil(t, gotNums, "Numbers must not be nil for valid Set")
	assert.Equal(t, wantNums, gotNums, "Numbers")

	gotClasses := got.Classes()
	require.NotNil(t, gotClasses, "Classes must not be nil for valid Set")
	assert.Equal(t, wantNums, numbersOfClasses(gotClasses), "Classes numbers")
	for _, c := range gotClasses {
		assert.True(t, c.IsValid())
		assert.Equal(t, wantModulus, c.Modulus())
	}

	iterClasses := slices.Collect(got.All())
	if iterClasses == nil {
		iterClasses = []pitchclass.Class{}
	}
	assert.Equal(t, gotClasses, iterClasses, "All iterator")
}

// takeSeq returns at most limit elements from seq and breaks early.
func takeSeq[T any](seq iter.Seq[T], limit int) []T {
	out := make([]T, 0, max(0, limit))
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

// callsAfterStopSeq invokes seq directly and counts extra calls after yield returns false.
func callsAfterStopSeq[T any](seq iter.Seq[T], stop int) int {
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

// assertSeqContract verifies that seq is non-nil, repeatable, yields want, and
// stops immediately when yield returns false.
func assertSeqContract[T any](t testing.TB, seq iter.Seq[T], want []T) {
	t.Helper()

	require.NotNil(t, seq, "iterator must not be nil")
	if want == nil {
		want = []T{}
	}

	assert.Equal(t, want, takeSeq(seq, len(want)+1), "first full pass")
	assert.Equal(t, want, takeSeq(seq, len(want)+1), "second full pass")

	for k := 1; k <= min(3, len(want)); k++ {
		assert.Equal(t, want[:k], takeSeq(seq, k), "early stop at %d", k)
		assert.Zero(t, callsAfterStopSeq(seq, k), "extra calls after stop at %d", k)
	}
}

// assertEmptySeqContract verifies that seq is a non-nil iterator yielding 0 elements.
func assertEmptySeqContract[T any](t testing.TB, seq iter.Seq[T]) {
	t.Helper()

	require.NotNil(t, seq, "iterator must not be nil")
	assert.Empty(t, takeSeq(seq, 1))
}

// Independent mathematical oracles using math/big and plain slices/maps.

func bigInt(v int) *big.Int {
	return big.NewInt(int64(v))
}

func modOracle(a, n int) int {
	return int(new(big.Int).Mod(bigInt(a), bigInt(n)).Int64())
}

func addModOracle(a, b, n int) int {
	s := new(big.Int).Add(bigInt(a), bigInt(b))

	return int(s.Mod(s, bigInt(n)).Int64())
}

func subModOracle(a, b, n int) int {
	d := new(big.Int).Sub(bigInt(a), bigInt(b))

	return int(d.Mod(d, bigInt(n)).Int64())
}

func mulModOracle(a, b, n int) int {
	p := new(big.Int).Mul(bigInt(a), bigInt(b))

	return int(p.Mod(p, bigInt(n)).Int64())
}

func icOracle(x, y, n int) int {
	d1 := subModOracle(y, x, n)
	d2 := subModOracle(x, y, n)

	return min(d1, d2)
}

type oracleSet struct {
	modulus int
	elems   []int
}

func newOracleSet(modulus int, nums ...int) oracleSet {
	present := make(map[int]bool, len(nums))
	for _, x := range nums {
		present[x] = true
	}

	elems := make([]int, 0, len(present))
	for x := range present {
		elems = append(elems, x)
	}
	slices.Sort(elems)

	return oracleSet{modulus: modulus, elems: elems}
}

func reduceOracleSet(modulus int, xs ...int) oracleSet {
	nums := make([]int, len(xs))
	for i, x := range xs {
		nums[i] = modOracle(x, modulus)
	}

	return newOracleSet(modulus, nums...)
}

func (s oracleSet) has(x int) bool {
	_, found := slices.BinarySearch(s.elems, x)

	return found
}

func bitmaskBigOracle(elems []int) *big.Int {
	b := big.NewInt(0)
	for _, x := range elems {
		b.SetBit(b, x, 1)
	}

	return b
}

func transposeOracle(s oracleSet, n int) oracleSet {
	out := make([]int, len(s.elems))
	for i, x := range s.elems {
		out[i] = addModOracle(x, n, s.modulus)
	}

	return newOracleSet(s.modulus, out...)
}

func invertOracle(s oracleSet, index int) oracleSet {
	out := make([]int, len(s.elems))
	for i, x := range s.elems {
		out[i] = subModOracle(index, x, s.modulus)
	}

	return newOracleSet(s.modulus, out...)
}

func multiplyOracle(s oracleSet, m int) oracleSet {
	out := make([]int, len(s.elems))
	for i, x := range s.elems {
		out[i] = mulModOracle(m, x, s.modulus)
	}

	return newOracleSet(s.modulus, out...)
}

func complementOracle(s oracleSet) oracleSet {
	var out []int
	for x := range s.modulus {
		if !s.has(x) {
			out = append(out, x)
		}
	}

	return newOracleSet(s.modulus, out...)
}

func unionOracle(a, b oracleSet) oracleSet {
	out := slices.Clone(a.elems)
	out = append(out, b.elems...)

	return newOracleSet(a.modulus, out...)
}

func intersectionOracle(a, b oracleSet) oracleSet {
	var out []int
	for _, x := range a.elems {
		if b.has(x) {
			out = append(out, x)
		}
	}

	return newOracleSet(a.modulus, out...)
}

func differenceOracle(a, b oracleSet) oracleSet {
	var out []int
	for _, x := range a.elems {
		if !b.has(x) {
			out = append(out, x)
		}
	}

	return newOracleSet(a.modulus, out...)
}

func symDiffOracle(a, b oracleSet) oracleSet {
	return unionOracle(differenceOracle(a, b), differenceOracle(b, a))
}

func isSubsetOracle(a, b oracleSet) bool {
	for _, x := range a.elems {
		if !b.has(x) {
			return false
		}
	}

	return true
}

func rotationsOracle(elems []int) [][]int {
	k := len(elems)
	if k == 0 {
		return [][]int{}
	}

	out := make([][]int, k)
	for i := range k {
		rot := make([]int, k)
		for j := range k {
			rot[j] = elems[(i+j)%k]
		}
		out[i] = rot
	}

	return out
}

// normalFormRahnOracle computes the normal form by Rahn's algorithm (P-18):
// minimize d_{k-1}, d_{k-2}, ..., d_1, then r_0.
func normalFormRahnOracle(modulus int, elems []int) []int {
	rots := rotationsOracle(elems)
	if len(rots) == 0 {
		return []int{}
	}

	best := rots[0]
	for _, cand := range rots[1:] {
		if compareRahnRotations(modulus, cand, best) < 0 {
			best = cand
		}
	}

	return best
}

func compareRahnRotations(modulus int, a, b []int) int {
	k := len(a)
	for j := k - 1; j >= 1; j-- {
		da := subModOracle(a[j], a[0], modulus)
		db := subModOracle(b[j], b[0], modulus)
		if da != db {
			if da < db {
				return -1
			}

			return 1
		}
	}

	if a[0] < b[0] {
		return -1
	}
	if a[0] > b[0] {
		return 1
	}

	return 0
}

// normalFormForteOracle computes the normal form by Forte's algorithm (P-19):
// minimize d_{k-1}, d_1, d_2, ..., d_{k-2}, then r_0.
func normalFormForteOracle(modulus int, elems []int) []int {
	rots := rotationsOracle(elems)
	if len(rots) == 0 {
		return []int{}
	}

	best := rots[0]
	for _, cand := range rots[1:] {
		if compareForteRotations(modulus, cand, best) < 0 {
			best = cand
		}
	}

	return best
}

func compareForteRotations(modulus int, a, b []int) int {
	k := len(a)
	if k >= 2 {
		daSpan := subModOracle(a[k-1], a[0], modulus)
		dbSpan := subModOracle(b[k-1], b[0], modulus)
		if daSpan != dbSpan {
			if daSpan < dbSpan {
				return -1
			}

			return 1
		}

		for j := 1; j <= k-2; j++ {
			da := subModOracle(a[j], a[0], modulus)
			db := subModOracle(b[j], b[0], modulus)
			if da != db {
				if da < db {
					return -1
				}

				return 1
			}
		}
	}

	if a[0] < b[0] {
		return -1
	}
	if a[0] > b[0] {
		return 1
	}

	return 0
}

// icvOracle computes the interval-class vector of length floor(modulus/2) by
// counting all unordered pairs (P-21).
func icvOracle(modulus int, elems []int) []int {
	counts := make([]int, modulus/2)
	for i := range elems {
		for j := i + 1; j < len(elems); j++ {
			d := icOracle(elems[i], elems[j], modulus)
			counts[d-1]++
		}
	}

	return counts
}

// morrisVectorOracle computes the 8-element Morris invariance vector for N = 12 (P-26).
func morrisVectorOracle(elems []int) [8]int {
	s := newOracleSet(12, elems...)
	comp := complementOracle(s)
	multipliers := [4]int{1, 11, 5, 7}

	var out [8]int
	for idx, m := range multipliers {
		ms := multiplyOracle(s, m)
		for n := range 12 {
			tn := transposeOracle(ms, n)
			if slices.Equal(tn.elems, s.elems) {
				out[idx]++
			}
			if isSubsetOracle(tn, comp) {
				out[idx+4]++
			}
		}
	}

	return out
}

func binomBigOracle(n, k int) *big.Int {
	if k < 0 || k > n {
		return big.NewInt(0)
	}

	return new(big.Int).Binomial(int64(n), int64(k))
}
