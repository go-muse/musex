package meter_test

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
)

// namedError is a sentinel error with its qualified name for diagnostics.
type namedError struct {
	name string
	err  error
}

// meterSentinels are the ten sentinels of architecture 4.1 (E-1 to E-10).
var meterSentinels = []namedError{
	{"meter.ErrUnset", meter.ErrUnset},
	{"meter.ErrInvalidArgument", meter.ErrInvalidArgument},
	{"meter.ErrGroupingRequired", meter.ErrGroupingRequired},
	{"meter.ErrNoCommonDenominator", meter.ErrNoCommonDenominator},
	{"meter.ErrUndefined", meter.ErrUndefined},
	{"meter.ErrNotNotatable", meter.ErrNotNotatable},
	{"meter.ErrOutsideMeter", meter.ErrOutsideMeter},
	{"meter.ErrTooMany", meter.ErrTooMany},
	{"meter.ErrInvalidText", meter.ErrInvalidText},
	{"meter.ErrOutOfRange", meter.ErrOutOfRange},
}

// dependencySentinels are the sentinels of the dependencies duration and
// common/fraction.
var dependencySentinels = []namedError{
	{"duration.ErrUnset", duration.ErrUnset},
	{"duration.ErrInvalidArgument", duration.ErrInvalidArgument},
	{"duration.ErrNotPositive", duration.ErrNotPositive},
	{"duration.ErrNegativePosition", duration.ErrNegativePosition},
	{"duration.ErrNotNotatable", duration.ErrNotNotatable},
	{"duration.ErrOutsideSeries", duration.ErrOutsideSeries},
	{"duration.ErrNoSubdivision", duration.ErrNoSubdivision},
	{"duration.ErrQuantizedToZero", duration.ErrQuantizedToZero},
	{"duration.ErrTooManyParts", duration.ErrTooManyParts},
	{"duration.ErrInvalidText", duration.ErrInvalidText},
	{"duration.ErrOutOfRange", duration.ErrOutOfRange},
	{"fraction.ErrInvalid", fraction.ErrInvalid},
	{"fraction.ErrZeroDenominator", fraction.ErrZeroDenominator},
	{"fraction.ErrDivisionByZero", fraction.ErrDivisionByZero},
	{"fraction.ErrNotRepresentable", fraction.ErrNotRepresentable},
	{"fraction.ErrOutOfRange", fraction.ErrOutOfRange},
	{"fraction.ErrInvalidText", fraction.ErrInvalidText},
}

// allSentinels lists the sentinels of the package and of its dependencies.
func allSentinels() []namedError {
	return append(slices.Clone(meterSentinels), dependencySentinels...)
}

// assertErr checks the error class of architecture 4.2: errors.Is holds for
// the package sentinel want and for no other package sentinel, and for every
// dependency sentinel in causes. Without causes the error wraps no
// dependency sentinel ("without wrapping of dependencies"). With causes,
// other wrapped dependency errors are not part of the contract and are not
// checked.
func assertErr(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := checkPackageSentinel(t, err, want)

	for _, s := range dependencySentinels {
		required := slices.Contains(causes, s.err)

		switch {
		case required && !errors.Is(err, s.err):
			ok = false

			t.Errorf("errors.Is(err, %s) = false, want true; err: %.200s", s.name, err)
		case len(causes) == 0 && errors.Is(err, s.err):
			ok = false

			t.Errorf("errors.Is(err, %s) = true, want false; err: %.200s", s.name, err)
		}
	}

	return ok
}

// assertErrExact checks the error class of a parse error (architecture
// 4.15): the package sentinel want and exactly the dependency sentinels
// causes, and no other sentinel.
func assertErrExact(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := checkPackageSentinel(t, err, want)

	for _, s := range dependencySentinels {
		expected := slices.Contains(causes, s.err)
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.200s", s.name, !expected, expected, err)
		}
	}

	return ok
}

// checkPackageSentinel checks that err wraps exactly the package sentinel
// want among the package sentinels.
func checkPackageSentinel(t *testing.T, err error, want error) bool {
	t.Helper()

	ok := true

	for _, s := range meterSentinels {
		expected := s.err == want
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.200s", s.name, !expected, expected, err)
		}
	}

	return ok
}

// sentinelName returns the name of the package sentinel wrapped by err, or
// "nil" for a nil error, for comparing outcomes in property tests.
func sentinelName(err error) string {
	if err == nil {
		return "nil"
	}

	for _, s := range meterSentinels {
		if errors.Is(err, s.err) {
			return s.name
		}
	}

	return "unknown: " + err.Error()
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

// panicError runs f, expects a panic with an error value, and returns it.
func panicError(t *testing.T, f func()) error {
	t.Helper()

	value, panicked := capturePanic(f)
	if !assert.True(t, panicked, "expected a panic") {
		return nil
	}

	err, ok := value.(error)
	assert.True(t, ok, "panic value is %T, want error", value)

	return err
}

// fastLimit bounds calls that architecture 4.9 and 6 require to finish
// without work proportional to their argument (E-8 of GroupingVariants) and
// parsing of long texts.
const fastLimit = 5 * time.Second

// fast runs f in a goroutine and returns its result. If f does not return
// within fastLimit, the test fails without waiting for the goroutine.
// A panic of f fails the test instead of the test binary.
func fast[T any](t *testing.T, f func() T) T {
	t.Helper()

	type outcome struct {
		v        T
		panicked bool
		value    any
	}

	results := make(chan outcome, 1)

	go func() {
		var o outcome

		o.value, o.panicked = capturePanic(func() { o.v = f() })
		results <- o
	}()

	select {
	case o := <-results:
		if o.panicked {
			t.Fatalf("the call panicked: %v", o.value)
		}

		return o.v
	case <-time.After(fastLimit):
		t.Fatalf("the call did not return within %s", fastLimit)

		var zero T

		return zero
	}
}

// result pairs a value with an error for fast calls.
type result[T any] struct {
	v   T
	err error
}

// noErr returns the value of a successful call and panics with the error of
// a failed one; callers recover the panic as a test failure.
func noErr[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// e returns only the error of a call with one result.
func e[T any](_ T, err error) error { return err }

// e2 returns only the error of a call with two results.
func e2[T, U any](_ T, _ U, err error) error { return err }

// Short constructors of values of the dependencies. They panic on misuse,
// which is a test error rather than a package error.

func dur(n, d int) duration.Duration { return duration.MustNew(n, d) }

func pos(n, d int) duration.Position { return duration.MustNewPosition(n, d) }

// wrs returns written durations by their canonical texts in duration.
func wrs(texts ...string) []duration.Written {
	out := make([]duration.Written, len(texts))
	for i, s := range texts {
		out[i] = duration.MustParseWritten(s)
	}

	return out
}

// ratText parses an exact rational number such as "3/8" or "0".
func ratText(s string) *big.Rat {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic("ratText: " + s)
	}

	return r
}

// fracRat converts an exact rational number to a reduced fraction in range.
func fracRat(r *big.Rat) fraction.Fraction { return fraction.MustFromBigRat(r) }

// durRat converts an exact positive rational number to a duration.
func durRat(r *big.Rat) duration.Duration { return duration.MustFromFraction(fracRat(r)) }

// posRat converts an exact non-negative rational number to a position.
func posRat(r *big.Rat) duration.Position { return duration.MustPositionFromFraction(fracRat(r)) }

// durT and posT are a duration and a position by text such as "3/8".
func durT(s string) duration.Duration { return durRat(ratText(s)) }

func posT(s string) duration.Position { return posRat(ratText(s)) }

// poss returns positions by their texts; no texts give nil.
func poss(texts ...string) []duration.Position {
	if len(texts) == 0 {
		return nil
	}

	out := make([]duration.Position, len(texts))
	for i, s := range texts {
		out[i] = posT(s)
	}

	return out
}

// sp returns the span "duration@start", such as "3/8@1/4".
func sp(s string) meter.Span {
	d, st, ok := strings.Cut(s, "@")
	if !ok {
		panic("sp: " + s)
	}

	return meter.Span{Start: posT(st), Duration: durT(d)}
}

// sps returns spans by their texts.
func sps(texts ...string) []meter.Span {
	out := make([]meter.Span, len(texts))
	for i, s := range texts {
		out[i] = sp(s)
	}

	return out
}

// equalSpans returns n spans of duration d that follow each other from 0.
func equalSpans(n int, d string) []meter.Span {
	out := make([]meter.Span, n)
	step := ratText(d)
	start := new(big.Rat)

	for i := range out {
		out[i] = meter.Span{Start: posRat(start), Duration: durRat(step)}
		start = new(big.Rat).Add(start, step)
	}

	return out
}

// pow2 returns 2^e as a big.Int for e ≥ 0.
func pow2(e int) *big.Int { return new(big.Int).Lsh(big.NewInt(1), uint(e)) }

// mersenne returns 2^p − 1 as an int for p ≤ 63.
func mersenne(p int) int { return int(new(big.Int).Sub(pow2(p), big.NewInt(1)).Int64()) }

// maxComponent is fraction.MaxComponent, 2^255 − 1.
func maxComponent() *big.Int { return fraction.MaxComponentBig() }

// kp39Denominators are the pairwise coprime denominators of КП-39 and of
// implementation-plan 4: 2^61−1, 2^59−1, 2^53−1, 2^47−1, 2^43−1. The first
// four parts 1/p give a sum within range; all five do not.
func kp39Denominators() []int {
	return []int{mersenne(61), mersenne(59), mersenne(53), mersenne(47), mersenne(43)}
}

// unitParts returns the simple signatures 1/d for each denominator.
func unitParts(dens []int) []meter.Meter {
	out := make([]meter.Meter, len(dens))
	for i, d := range dens {
		out[i] = meter.MustNew(1, d)
	}

	return out
}

// kp39Text is the text of the КП-39 sum of five parts.
func kp39Text() string {
	parts := make([]string, 0, 5)
	for _, d := range kp39Denominators() {
		parts = append(parts, "1/"+strconv.Itoa(d))
	}

	return strings.Join(parts, "+")
}

// mt builds a meter from the notation of business-spec 3 and of the
// canonical text of architecture 4.15 by the constructors only, without
// meter.Parse: "3/4", "5/8[2+3]", "(3+2+3)/8", "2/4+3/8", "C", "C|",
// "C[3+1]". Spaces are ignored. It panics on misuse, which is a test error.
func mt(text string) meter.Meter {
	s := strings.ReplaceAll(text, " ", "")

	parts := splitTop(s)
	if len(parts) > 1 {
		ms := make([]meter.Meter, len(parts))
		for i, p := range parts {
			ms[i] = mtSingle(p)
		}

		return meter.MustNewSum(ms...)
	}

	return mtSingle(s)
}

// splitTop splits s at "+" outside parentheses and brackets.
func splitTop(s string) []string {
	var (
		out   []string
		depth int
		from  int
	)

	for i, r := range s {
		switch r {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case '+':
			if depth == 0 {
				out = append(out, s[from:i])
				from = i + 1
			}
		}
	}

	return append(out, s[from:])
}

// mtSingle builds one simple signature, symbol, or additive signature with
// an optional explicit grouping.
func mtSingle(s string) meter.Meter {
	var groups []int

	if head, g, ok := strings.Cut(s, "["); ok {
		s = head
		groups = ints(strings.Split(strings.TrimSuffix(g, "]"), "+"))
	}

	var m meter.Meter

	switch {
	case s == "C":
		m = meter.MustNewSymbol(meter.SymbolCommon)
	case s == "C|" || s == "¢":
		m = meter.MustNewSymbol(meter.SymbolCut)
	case strings.HasPrefix(s, "("):
		inner, den, ok := strings.Cut(strings.TrimPrefix(s, "("), ")/")
		if !ok {
			panic("mt: " + s)
		}

		m = meter.MustNewAdditive(ints(strings.Split(inner, "+")), atoi(den))
	default:
		n, d, ok := strings.Cut(s, "/")
		if !ok {
			panic("mt: " + s)
		}

		m = meter.MustNew(atoi(n), atoi(d))
	}

	if groups != nil {
		m = m.MustWithGrouping(groups...)
	}

	return m
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return n
}

func ints(ss []string) []int {
	out := make([]int, len(ss))
	for i, s := range ss {
		out[i] = atoi(s)
	}

	return out
}

// mts builds meters by mt.
func mts(texts ...string) []meter.Meter {
	out := make([]meter.Meter, len(texts))
	for i, s := range texts {
		out[i] = mt(s)
	}

	return out
}

// compositions returns, as an independent oracle for О-17, all ordered
// partitions of n into parts 2 and 3 in lexicographic order (2 before 3).
func compositions(n int) [][]int {
	if n == 0 {
		return [][]int{nil}
	}

	var out [][]int

	for _, first := range []int{2, 3} {
		if first > n {
			continue
		}

		for _, rest := range compositions(n - first) {
			out = append(out, append([]int{first}, rest...))
		}
	}

	return out
}

// corpusTexts are the signatures of the corpus of implementation-plan 4
// besides simple n/d: symbols with and without explicit grouping, compound
// numerators, sums of two and three parts, and explicit groupings.
var corpusTexts = []string{
	"C", "C|", "C[3+1]", "C[1+3]", "C[4]", "C[1+1+1+1]", "C|[1+1]",
	"(2+2)/4", "(3+3)/8", "(2+3)/8", "(3+2)/8", "(3+2+3)/8", "(2+2+2+2)/8", "(4+4)/8",
	"(1+1)/4", "(2+3+2)/8", "(1+2+3+4)/16", "(3+3)/6", "(2+2)/3",
	"2/4+2/4", "2/4+3/8", "2/8+3/8", "3/8+2/8", "4/4+3/8", "2/4+2/8", "3/8+3/4",
	"(2+3)/8+2/4", "4/8+3/8", "2/4+2/3", "3/8+3/8", "2/8+2/8+2/8", "1/4+1/4",
	"3/4+3/4+3/4", "(2+2)/4+3/8", "2/2+(3+3)/8", "1/1+1/2+1/4", "5/8+2/8", "5/8+2/4",
	"5/8+1/8", "(1+2)/4+1/4",
	"5/8[2+3]", "5/8[3+2]", "5/8[5]", "5/8[1+4]", "7/8[3+2+2]", "7/8[2+2+3]", "7/8[7]",
	"7/8[5+2]", "5/4[3+2]", "6/8[2+2+2]", "6/8[1+5]", "6/4[2+2+2]", "4/4[3+1]",
	"4/4[1+1+1+1]", "3/4[2+1]", "2/4[1+1]", "12/8[2+2+2+2+2+2]", "9/8[2+2+2+3]",
	"8/8[3+3+2]", "8/8[4+4]", "4/3[1+3]", "6/6[2+2+2]", "15/8", "1/1024", "3/2048",
	"1024/8", "1023/8", "6/1024", "3/6", "6/6", "6/3",
}

// corpus is the corpus of implementation-plan 4 for property tests: simple
// n/d for n from 1 to 13 and d from {1, 2, 3, 4, 8, 16}, and corpusTexts.
func corpus() []meter.Meter {
	var out []meter.Meter

	for n := 1; n <= 13; n++ {
		for _, d := range []int{1, 2, 3, 4, 8, 16} {
			out = append(out, meter.MustNew(n, d))
		}
	}

	return append(out, mts(corpusTexts...)...)
}

// describe returns a short description of a call outcome for property
// comparisons: the value printed by fmt or the package sentinel.
func describe(v any, err error) string {
	if err != nil {
		return "error " + sentinelName(err)
	}

	return fmt.Sprint(v)
}

// ratOfDuration returns the exact value of a set duration.
func ratOfDuration(d duration.Duration) *big.Rat {
	r, ok := d.MustFraction().BigRat()
	if !ok {
		panic("ratOfDuration")
	}

	return r
}

// lastWord returns the text after the last space of a case name, so that
// "КП-3 3/4" names the signature "3/4".
func lastWord(s string) string {
	if i := strings.LastIndex(s, " "); i >= 0 {
		return s[i+1:]
	}

	return s
}

// itoa is strconv.Itoa for case names.
func itoa(n int) string { return strconv.Itoa(n) }

// guard turns a panic of a test body into a failure of that test, so that a
// panicking helper such as noErr or a Must variant does not stop the whole
// test binary. It is deferred directly after t.Parallel.
func guard(t *testing.T) {
	t.Helper()

	if r := recover(); r != nil {
		t.Errorf("unexpected panic: %v", r)
	}
}
