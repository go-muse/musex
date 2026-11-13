package duration_test

import (
	"errors"
	"math"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// namedError is a sentinel error with its qualified name for diagnostics.
type namedError struct {
	name string
	err  error
}

// durationSentinels are the 11 sentinels of architecture 4.2 (E-1 to E-11).
var durationSentinels = []namedError{
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
}

// dependencySentinels are all sentinels of common/fraction, tuplet, and
// common/rounding. An error of the package wraps exactly the dependency
// sentinels named by architecture 4.2 and no other.
var dependencySentinels = []namedError{
	{"fraction.ErrInvalid", fraction.ErrInvalid},
	{"fraction.ErrZeroDenominator", fraction.ErrZeroDenominator},
	{"fraction.ErrDivisionByZero", fraction.ErrDivisionByZero},
	{"fraction.ErrNotRepresentable", fraction.ErrNotRepresentable},
	{"fraction.ErrOutOfRange", fraction.ErrOutOfRange},
	{"fraction.ErrInvalidText", fraction.ErrInvalidText},
	{"tuplet.ErrInvalidCount", tuplet.ErrInvalidCount},
	{"tuplet.ErrUnset", tuplet.ErrUnset},
	{"tuplet.ErrNoDefault", tuplet.ErrNoDefault},
	{"tuplet.ErrInvalidText", tuplet.ErrInvalidText},
	{"tuplet.ErrOutOfRange", tuplet.ErrOutOfRange},
	{"tuplet.ErrInvalidArgument", tuplet.ErrInvalidArgument},
	{"tuplet.ErrNotRepresentable", tuplet.ErrNotRepresentable},
	{"rounding.ErrUnset", rounding.ErrUnset},
	{"rounding.ErrUnknownMode", rounding.ErrUnknownMode},
	{"rounding.ErrInvalidArgument", rounding.ErrInvalidArgument},
	{"rounding.ErrOutOfRange", rounding.ErrOutOfRange},
}

// assertErr checks the error class of architecture 4.2: errors.Is holds for
// the package sentinel want and for exactly the dependency sentinels causes,
// and for no other sentinel of the package or its dependencies.
func assertErr(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := true

	for _, s := range durationSentinels {
		expected := s.err == want
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.200s", s.name, !expected, expected, err)
		}
	}

	for _, s := range dependencySentinels {
		expected := slices.Contains(causes, s.err)
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.200s", s.name, !expected, expected, err)
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

// fastLimit bounds calls that architecture 6 requires to fail without work
// proportional to the size of the value (E-9, E-11 for dots, large maxDots).
const fastLimit = 5 * time.Second

// fast runs f in a goroutine and returns its result. If f does not return
// within fastLimit, the test fails at once without waiting for the goroutine;
// the overall bound of a run is the -timeout of go test. A panic of f is
// recovered in the goroutine and fails the test instead of the test binary.
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

// Short constructors of test values. They panic on misuse, which is a test
// error rather than a package error.

func frac(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

func dur(n, d int) duration.Duration { return duration.MustNew(n, d) }

func pos(n, d int) duration.Position { return duration.MustNewPosition(n, d) }

func tp(n, m int) tuplet.Tuplet { return tuplet.MustNew(n, m) }

func wr(b duration.Base, dots int, chain ...tuplet.Tuplet) duration.Written {
	return duration.MustNewWritten(b, dots, chain...)
}

func tie(parts ...duration.Written) duration.Tied { return duration.MustNewTied(parts...) }

// fracRat converts an exact rational number to a reduced fraction; the
// value must be in range.
func fracRat(r *big.Rat) fraction.Fraction { return fraction.MustFromBigRat(r) }

// durRat converts an exact positive rational number to a duration.
func durRat(r *big.Rat) duration.Duration { return duration.MustFromFraction(fracRat(r)) }

// posRat converts an exact non-negative rational number to a position.
func posRat(r *big.Rat) duration.Position { return duration.MustPositionFromFraction(fracRat(r)) }

// ratOf returns the exact value of a valid fraction.
func ratOf(f fraction.Fraction) *big.Rat {
	r, ok := f.BigRat()
	if !ok {
		panic("ratOf: invalid fraction")
	}

	return r
}

// ratOfDuration returns the exact value of a set duration.
func ratOfDuration(d duration.Duration) *big.Rat { return ratOf(d.MustFraction()) }

// ratOfPosition returns the exact value of a set position.
func ratOfPosition(p duration.Position) *big.Rat { return ratOf(p.MustFraction()) }

// rat returns n/d as a big.Rat.
func rat(n, d int64) *big.Rat { return big.NewRat(n, d) }

// pow2 returns 2^e for any integer e.
func pow2(e int) *big.Rat {
	if e >= 0 {
		return new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(e)))
	}

	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), uint(-e)))
}

// ratInt returns the integer n as a big.Rat.
func ratInt(n *big.Int) *big.Rat { return new(big.Rat).SetInt(n) }

// maxComponent is fraction.MaxComponent, 2^255 - 1.
func maxComponent() *big.Int { return fraction.MaxComponentBig() }

// near2pow returns 2^e + delta.
func near2pow(e int, delta int64) *big.Int {
	return new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), uint(e)), big.NewInt(delta))
}

// bigTuplet is MaxInt:1 with coefficient 1/MaxInt. Four of them have a
// coefficient in range; five do not (tuplet documentation).
var bigTuplet = tuplet.MustNew(math.MaxInt, 1)

// bigChain returns n copies of bigTuplet.
func bigChain(n int) []tuplet.Tuplet {
	chain := make([]tuplet.Tuplet, n)
	for i := range chain {
		chain[i] = bigTuplet
	}

	return chain
}

// edgeWritten is an eighth under four MaxInt:1 tuplets: its value
// 1/(8·MaxInt^4) is in range, while a sixteenth under the same chain is not.
func edgeWritten() duration.Written { return wr(duration.Eighth, 0, bigChain(4)...) }

// rep repeats u to n bytes for a one-byte u, or floor(n/len(u)) whole copies.
func rep(u string, n int) string { return strings.Repeat(u, n/len(u)) }

// allModes are the nine modes of common/rounding.
func allModes() []rounding.Mode { return rounding.Modes() }
