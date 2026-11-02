package number_test

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests assume a 64-bit int (the project supports only 64-bit systems).
const (
	minInt = math.MinInt
	maxInt = math.MaxInt
)

// spec describes a numbering by the mathematical definition of the
// architecture: the anchor pair (aspn -> num) and the finite set of skipped
// numbers. Together with the functions below it is the test oracle: all
// arithmetic is exact (math/big) and does not depend on the package under
// test.
type spec struct {
	aspn    int
	num     int
	skipped []int // ascending; the definition is valid, so no repeats
}

// newSpec returns the definition of New(aspn, num, skipped...). The skipped
// numbers are copied and sorted, so the caller's slice is never shared.
func newSpec(aspn, num int, skipped ...int) spec {
	s := slices.Clone(skipped)
	slices.Sort(s)

	return spec{aspn: aspn, num: num, skipped: s}
}

func bigInt(v int) *big.Int {
	return big.NewInt(int64(v))
}

// numberOf returns N(a): the number of the octave with ASPN number a.
// For a >= aspn it counts used numbers upward from num, for a < aspn downward.
func (s spec) numberOf(a *big.Int) *big.Int {
	a0, x0 := bigInt(s.aspn), bigInt(s.num)
	y := new(big.Int)
	one := big.NewInt(1)

	if a.Cmp(a0) >= 0 {
		y.Sub(a, a0)
		y.Add(y, x0)
		for _, v := range s.skipped {
			sv := bigInt(v)
			if sv.Cmp(x0) > 0 && sv.Cmp(y) <= 0 {
				y.Add(y, one)
			}
		}

		return y
	}

	y.Sub(a0, a)
	y.Sub(x0, y)
	for i := len(s.skipped) - 1; i >= 0; i-- {
		sv := bigInt(s.skipped[i])
		if sv.Cmp(y) >= 0 && sv.Cmp(x0) < 0 {
			y.Sub(y, one)
		}
	}

	return y
}

// aspnOf returns N^-1(x), the inverse of N: the ASPN number of the octave
// with number x, and false when x is skipped.
func (s spec) aspnOf(x int) (*big.Int, bool) {
	if s.isSkipped(x) {
		return nil, false
	}

	a := new(big.Int)
	count := 0
	if x >= s.num {
		for _, v := range s.skipped {
			if v > s.num && v < x {
				count++
			}
		}
		a.Sub(bigInt(x), bigInt(s.num))
		a.Add(a, bigInt(s.aspn))
		a.Sub(a, bigInt(count))

		return a, true
	}

	for _, v := range s.skipped {
		if v > x && v < s.num {
			count++
		}
	}
	a.Sub(bigInt(s.num), bigInt(x))
	a.Sub(bigInt(s.aspn), a)
	a.Add(a, bigInt(count))

	return a, true
}

func (s spec) isSkipped(x int) bool {
	return slices.Contains(s.skipped, x)
}

// outcome is the expected result of an operation that returns (int, error):
// the value when err is nil, otherwise the sentinel error and the value 0.
type outcome struct {
	value int
	err   error
}

func (o outcome) String() string {
	if o.err != nil {
		return fmt.Sprintf("error %v", o.err)
	}

	return fmt.Sprint(o.value)
}

// fitInt turns an exact result into an outcome: a value that does not fit
// into a 64-bit int is ErrOutOfRange.
func fitInt(b *big.Int) outcome {
	if !b.IsInt64() {
		return outcome{err: number.ErrOutOfRange}
	}

	return outcome{value: int(b.Int64())}
}

// fromASPN is the expected result of FromASPN(a).
func (s spec) fromASPN(a int) outcome {
	return fitInt(s.numberOf(bigInt(a)))
}

// toASPN is the expected result of ToASPN(x).
func (s spec) toASPN(x int) outcome {
	a, ok := s.aspnOf(x)
	if !ok {
		return outcome{err: number.ErrSkipped}
	}

	return fitInt(a)
}

// convertSpec is the expected result of Convert(x, from, to): To(From^-1(x))
// computed exactly, the intermediate ASPN number may exceed int.
func convertSpec(x int, from, to spec) outcome {
	a, ok := from.aspnOf(x)
	if !ok {
		return outcome{err: number.ErrSkipped}
	}

	return fitInt(to.numberOf(a))
}

// equalSpec reports whether two numberings give the same numbers to all
// octaves: the same skipped numbers and the same number of one octave.
func equalSpec(p, q spec) bool {
	return slices.Equal(p.skipped, q.skipped) && q.numberOf(bigInt(p.aspn)).Cmp(bigInt(p.num)) == 0
}

// Indexes of the built-in numberings in the order of Builtins.
const (
	idxASPN = iota
	idxFrancoBelgian
	idxMexican
	idxMiddleC3
	idxMiddleC5
	builtinCount
)

// builtin describes one built-in numbering by the table of the architecture.
type builtin struct {
	label string
	get   func() number.Numbering
	spec  spec
}

// builtinTable returns the built-in numberings in the order of Builtins.
// Each call returns new values.
func builtinTable() []builtin {
	return []builtin{
		idxASPN:          {label: "ASPN", get: number.ASPN, spec: newSpec(4, 4)},
		idxFrancoBelgian: {label: "Franco-Belgian", get: number.FrancoBelgian, spec: newSpec(4, 3, 0)},
		idxMexican:       {label: "Mexican", get: number.Mexican, spec: newSpec(4, 5)},
		idxMiddleC3:      {label: "Middle C = C3", get: number.MiddleC3, spec: newSpec(4, 3)},
		idxMiddleC5:      {label: "Middle C = C5", get: number.MiddleC5, spec: newSpec(4, 5)},
	}
}

// builtinsEqual reports the expected Equal between built-ins i and j:
// only the same numbering and the pair Mexican, Middle C = C5.
func builtinsEqual(i, j int) bool {
	pair := func(a, b int) bool { return (i == a && j == b) || (i == b && j == a) }

	return i == j || pair(idxMexican, idxMiddleC5)
}

// numbering is a value under test together with its oracle definition.
type numbering struct {
	name string
	n    number.Numbering
	spec spec
}

// builtins returns the five built-in numberings with their definitions.
func builtins() []numbering {
	table := builtinTable()
	out := make([]numbering, 0, len(table))
	for _, b := range table {
		out = append(out, numbering{name: b.label, n: b.get(), spec: b.spec})
	}

	return out
}

// own creates New(aspn, num, skipped...) and fails the test on an error.
func own(t testing.TB, aspn, num int, skipped ...int) numbering {
	t.Helper()

	n, err := number.New(aspn, num, skipped...)
	require.NoError(t, err, "New(%d, %d, %v)", aspn, num, skipped)

	return numbering{name: ownName(aspn, num, skipped), n: n, spec: newSpec(aspn, num, skipped...)}
}

func ownName(aspn, num int, skipped []int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "New(%d, %d", aspn, num)
	for _, v := range skipped {
		fmt.Fprintf(&b, ", %d", v)
	}
	b.WriteString(")")

	return b.String()
}

// ownCriteria returns the own numberings named in K-7, K-8 and K-9,
// including the empty list of skipped numbers in both forms.
func ownCriteria(t testing.TB) []numbering {
	t.Helper()

	empty := own(t, 4, 3, []int{}...)
	empty.name = "New(4, 3, []int{}...)"
	none := own(t, 4, 3, nil...)
	none.name = "New(4, 3, nil...)"

	return []numbering{
		own(t, 4, 4),
		own(t, 4, 5),
		own(t, 0, 0),
		own(t, 4, 3),
		own(t, -1, 0),
		own(t, 4, 4, 0, 1),
		own(t, 4, 4, 1, 0),
		own(t, 4, 4, 6),
		own(t, 4, 4, -10),
		empty,
		none,
		own(t, 4, 3, 0),
		own(t, 2, 1, 0),
		own(t, 1, -1, 0),
		own(t, 4, 3, 1),
	}
}

// criteriaSet returns the built-in numberings followed by ownCriteria.
func criteriaSet(t testing.TB) []numbering {
	t.Helper()

	return append(builtins(), ownCriteria(t)...)
}

// edgeSet returns the built-in numberings and own numberings with anchors and
// skipped numbers at the limits of int (architecture, "Boundary cases").
func edgeSet(t testing.TB) []numbering {
	t.Helper()

	return append(builtins(),
		own(t, 0, maxInt),
		own(t, 1, minInt),
		own(t, maxInt, minInt),
		own(t, 0, minInt),
		own(t, 1, minInt+1),
		own(t, minInt, 0),
		own(t, 0, 0, maxInt),
		own(t, 0, 0, minInt),
		own(t, minInt, maxInt, minInt, maxInt-1),
		own(t, 0, 0, minInt, minInt+1, maxInt-1, maxInt),
		own(t, maxInt, maxInt, minInt, 0),
	)
}

// edgeValues are inputs near the limits of int and around zero.
func edgeValues() []int {
	return []int{
		minInt, minInt + 1, minInt + 2, minInt + 3,
		-1, 0, 1,
		maxInt - 3, maxInt - 2, maxInt - 1, maxInt,
	}
}

// allErrors returns the four sentinel errors of the package.
func allErrors() []error {
	return []error{
		number.ErrSkipped,
		number.ErrInvalidDefinition,
		number.ErrOutOfRange,
		number.ErrUndefined,
	}
}

// assertOnlyError checks that err matches want with errors.Is and none of the
// other three sentinel errors.
func assertOnlyError(t *testing.T, err, want error, desc string) bool {
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

// assertOutcome compares the result of an (int, error) operation with the
// oracle: on an error, exactly that sentinel and the value 0.
func assertOutcome(t *testing.T, want outcome, got int, err error, desc string) bool {
	t.Helper()

	if want.err != nil {
		ok := assertOnlyError(t, err, want.err, desc)

		return assert.Zero(t, got, desc) && ok
	}

	return assert.NoError(t, err, desc) && assert.Equal(t, want.value, got, desc)
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
// The text of the error is not part of the contract and is not compared.
func assertPanicsWithErrorIs(t *testing.T, want error, f func(), desc string) bool {
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
