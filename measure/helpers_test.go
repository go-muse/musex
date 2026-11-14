package measure_test

import (
	"errors"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The helpers build inputs only with the public API of the dependencies and
// with New and the operations of architecture 4.4. Parse of the package is
// used only by the groups Т-8 and Т-9.

// namedError is a sentinel error with its qualified name for diagnostics.
type namedError struct {
	name string
	err  error
}

// measureSentinels are the ten sentinels of architecture 4.1 (Е-1 to Е-10).
var measureSentinels = []namedError{
	{"measure.ErrUnset", measure.ErrUnset},
	{"measure.ErrInvalidArgument", measure.ErrInvalidArgument},
	{"measure.ErrLongerThanMeter", measure.ErrLongerThanMeter},
	{"measure.ErrOutsideGrid", measure.ErrOutsideGrid},
	{"measure.ErrOutsideBar", measure.ErrOutsideBar},
	{"measure.ErrUndefined", measure.ErrUndefined},
	{"measure.ErrGroupingRequired", measure.ErrGroupingRequired},
	{"measure.ErrTooMany", measure.ErrTooMany},
	{"measure.ErrInvalidText", measure.ErrInvalidText},
	{"measure.ErrOutOfRange", measure.ErrOutOfRange},
}

// dependencySentinels are the sentinels of meter, duration, and
// common/fraction.
var dependencySentinels = []namedError{
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
	return append(slices.Clone(measureSentinels), dependencySentinels...)
}

// assertErr checks the error class of architecture 4.2: errors.Is holds for
// the package sentinel want and for no other package sentinel, and for every
// dependency sentinel in causes. Other wrapped errors of the dependencies are
// not part of the contract and are not checked.
func assertErr(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := checkPackageSentinel(t, err, want)

	for _, s := range dependencySentinels {
		if slices.Contains(causes, s.err) && !errors.Is(err, s.err) {
			ok = false

			t.Errorf("errors.Is(err, %s) = false, want true; err: %.300s", s.name, err)
		}
	}

	return ok
}

// assertNoCause checks that err does not wrap the named dependency
// sentinels: the check order of architecture 4.2 returns a violation whose
// cause is not a failure of that dependency.
func assertNoCause(t *testing.T, err error, causes ...error) {
	t.Helper()

	for _, s := range dependencySentinels {
		if slices.Contains(causes, s.err) && errors.Is(err, s.err) {
			t.Errorf("errors.Is(err, %s) = true, want false; err: %.300s", s.name, err)
		}
	}
}

// assertErrExact checks the class of a parse error (architecture 4.16): the
// package sentinel want and exactly the dependency sentinels causes.
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

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.300s", s.name, !expected, expected, err)
		}
	}

	return ok
}

// checkPackageSentinel checks that err wraps exactly the package sentinel
// want among the package sentinels.
func checkPackageSentinel(t *testing.T, err error, want error) bool {
	t.Helper()

	ok := true

	for _, s := range measureSentinels {
		expected := s.err == want
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.300s", s.name, !expected, expected, err)
		}
	}

	return ok
}

// guard turns an unexpected panic of the code under test into a failure of
// the current test instead of a crash of the test binary.
func guard(t *testing.T) {
	t.Helper()

	if r := recover(); r != nil {
		t.Errorf("unexpected panic: %v", r)
	}
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

// e returns only the error of a call with one result.
func e[T any](_ T, err error) error { return err }

// e2 returns only the error of a call with two results.
func e2[T, U any](_ T, _ U, err error) error { return err }

// Values of the dependencies. They panic on misuse, which is a test error
// rather than a package error.

// mt is a meter by its text in package meter, such as "3/4", "C", or
// "2/4+3/8".
func mt(text string) meter.Meter { return meter.MustParse(text) }

// mts returns meters by their texts.
func mts(texts ...string) []meter.Meter {
	out := make([]meter.Meter, len(texts))
	for i, s := range texts {
		out[i] = mt(s)
	}

	return out
}

// ratText parses an exact rational number such as "3/8", "0", or "5".
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

// dur, pos, and frac are a duration, a position, and a fraction by text
// such as "3/8" or "2".
func dur(s string) duration.Duration { return durRat(ratText(s)) }

func pos(s string) duration.Position { return posRat(ratText(s)) }

func frac(s string) fraction.Fraction { return fracRat(ratText(s)) }

// posInt is the position n/1.
func posInt(n int) duration.Position { return duration.MustNewPosition(n, 1) }

// durInt is the duration n/1.
func durInt(n int) duration.Duration { return duration.MustNew(n, 1) }

// pow2 returns 2^e as a big.Int for e ≥ 0.
func pow2(e int) *big.Int { return new(big.Int).Lsh(big.NewInt(1), uint(e)) }

// ratInt returns n/1 as a big.Rat.
func ratInt(n *big.Int) *big.Rat { return new(big.Rat).SetInt(n) }

// ratFrac returns n/d as a big.Rat.
func ratFrac(n, d *big.Int) *big.Rat { return new(big.Rat).SetFrac(n, d) }

// assertText checks the canonical text of a value of a dependency: reduced
// durations, positions, and fractions, and the record of a meter.
func assertText(t *testing.T, want string, v interface{ String() string }, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Equal(t, want, v.String(), msgAndArgs...)
}

// assertFrac checks that a fraction result is reduced with a positive
// denominator and equals want (architecture 4.2, "Results").
func assertFrac(t *testing.T, want string, f fraction.Fraction, msgAndArgs ...any) bool {
	t.Helper()

	if !assert.True(t, f.IsValid(), msgAndArgs...) {
		return false
	}

	assert.True(t, f.IsReduced(), "the result fraction %s is not reduced", f)

	n, d, ok := f.Pair()
	if ok {
		rn, rd, rok := f.ReducedPair()
		assert.True(t, rok)
		assert.Equal(t, [2]int{rn, rd}, [2]int{n, d}, "Pair equals ReducedPair")
		assert.Positive(t, d, "the denominator is positive")
	}

	want = canonicalRat(want)

	return assert.Equal(t, want, f.String(), msgAndArgs...)
}

// canonicalRat writes an exact rational number as "n/d" in lowest terms.
func canonicalRat(s string) string {
	r := ratText(s)

	return r.Num().String() + "/" + r.Denom().String()
}

// Grids by the notation of business-spec 3, built only with New and the
// operations of architecture 4.4, without Parse of the package.

// step is one change of a grid.
type step func(measure.Grid) (measure.Grid, error)

// build returns the grid of meter m changed by steps in order.
func build(t *testing.T, m meter.Meter, steps ...step) measure.Grid {
	t.Helper()

	g, err := measure.New(m)
	require.NoError(t, err, "New(%s)", m)

	for i, s := range steps {
		g, err = s(g)
		require.NoError(t, err, "step %d", i)
	}

	return g
}

func withPickup(d string) step {
	return func(g measure.Grid) (measure.Grid, error) { return g.WithPickup(dur(d)) }
}

func withMark(bar int, m string) step {
	return func(g measure.Grid) (measure.Grid, error) { return g.WithMark(bar, mt(m)) }
}

func withLength(bar int, d string, a measure.Alignment) step {
	return func(g measure.Grid) (measure.Grid, error) { return g.WithLength(bar, dur(d), a) }
}

func withBarCount(n int) step {
	return func(g measure.Grid) (measure.Grid, error) { return g.WithBarCount(n) }
}

// gridOf builds a grid from a text in the notation of the canonical text of
// architecture 4.16, such as "4/4;pickup:1/4;3:3/4;4:length:1/2:end;bars:8",
// by New, WithPickup, WithMark, WithLength, and WithBarCount in this order.
// The text is split by the test, not parsed by the package.
func gridOf(t *testing.T, text string) measure.Grid {
	t.Helper()

	items := strings.Split(text, ";")

	var (
		pickups []step
		marks   []step
		lengths []step
		counts  []step
	)

	for _, item := range items[1:] {
		head, rest, ok := strings.Cut(item, ":")
		require.True(t, ok, "gridOf: %s", item)

		switch {
		case head == "pickup":
			pickups = append(pickups, withPickup(rest))
		case head == "bars":
			counts = append(counts, withBarCount(atoi(rest)))
		case strings.HasPrefix(rest, "length:"):
			d := strings.TrimPrefix(rest, "length:")

			a := measure.AlignStart
			if s, ok := strings.CutSuffix(d, ":end"); ok {
				d, a = s, measure.AlignEnd
			}

			lengths = append(lengths, withLength(atoi(head), d, a))
		default:
			marks = append(marks, withMark(atoi(head), rest))
		}
	}

	steps := slices.Concat(pickups, marks, lengths, counts)

	return build(t, mt(items[0]), steps...)
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return n
}

// record is the record of a grid read by Pickup, Marks, Lengths, and
// BarCount, in the texts of the records of architecture 4.13.
type record struct {
	Pickup  string
	Marks   []string
	Lengths []string
	Bars    int
}

// recordOf reads the record of g.
func recordOf(t *testing.T, g measure.Grid) record {
	t.Helper()

	var r record

	p, ok, err := g.Pickup()
	require.NoError(t, err)

	if ok {
		r.Pickup = p.String()
	}

	marks, err := g.Marks()
	require.NoError(t, err)

	for _, m := range marks {
		r.Marks = append(r.Marks, m.String())
	}

	lengths, err := g.Lengths()
	require.NoError(t, err)

	for _, l := range lengths {
		r.Lengths = append(r.Lengths, l.String())
	}

	n, finite, err := g.BarCount()
	require.NoError(t, err)

	if finite {
		r.Bars = n
	}

	return r
}

// recordText is the expected record of a text in the notation of gridOf.
func recordText(text string) record {
	items := strings.Split(text, ";")
	r := record{Marks: []string{"1:" + items[0]}}

	for _, item := range items[1:] {
		head, rest, _ := strings.Cut(item, ":")

		switch {
		case head == "pickup":
			r.Pickup = rest
		case head == "bars":
			r.Bars = atoi(rest)
		case strings.HasPrefix(rest, "length:"):
			r.Lengths = append(r.Lengths, item)
		default:
			r.Marks = append(r.Marks, item)
		}
	}

	return r
}

// assertRecord checks the record of g against a text in the notation of
// gridOf.
func assertRecord(t *testing.T, want string, g measure.Grid, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Equal(t, recordText(want), recordOf(t, g), msgAndArgs...)
}

// ad is the address (bar, offset) with an offset by text.
func ad(bar int, offset string) measure.Address { return measure.MustNewAddress(bar, pos(offset)) }

// barOf is the expected information of a bar, by texts of its values.
type barOf struct {
	number      int
	meter       string
	length      string
	start       string
	end         string
	kind        measure.Kind
	pickup      bool
	explicit    bool
	alignment   measure.Alignment
	shift       string
	marked      bool
	meterChange bool
}

// value returns the expected Bar.
func (b barOf) value() measure.Bar {
	return measure.Bar{
		Number:      b.number,
		Meter:       mt(b.meter),
		Length:      dur(b.length),
		Start:       pos(b.start),
		End:         pos(b.end),
		Kind:        b.kind,
		Pickup:      b.pickup,
		Explicit:    b.explicit,
		Alignment:   b.alignment,
		Shift:       pos(b.shift),
		Marked:      b.marked,
		MeterChange: b.meterChange,
	}
}

// partTexts returns the texts of parts, such as "1/4@2:0/1".
func partTexts(parts []measure.Part) []string {
	if parts == nil {
		return nil
	}

	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = p.String()
	}

	return out
}

// collect returns the bars of a sequence in order.
func collect(seq func(func(measure.Bar) bool)) []measure.Bar {
	var out []measure.Bar

	for b := range seq {
		out = append(out, b)
	}

	return out
}
