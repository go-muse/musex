package tempo_test

import (
	"errors"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// namedError is a sentinel error with its qualified name for diagnostics.
type namedError struct {
	name string
	err  error
}

// tempoSentinels are the 11 sentinels of architecture 4.2 (Е-1 to Е-11).
var tempoSentinels = []namedError{
	{"tempo.ErrUnset", tempo.ErrUnset},
	{"tempo.ErrInvalidArgument", tempo.ErrInvalidArgument},
	{"tempo.ErrNoTempo", tempo.ErrNoTempo},
	{"tempo.ErrNoSolution", tempo.ErrNoSolution},
	{"tempo.ErrNegativeTime", tempo.ErrNegativeTime},
	{"tempo.ErrOffScale", tempo.ErrOffScale},
	{"tempo.ErrNoRange", tempo.ErrNoRange},
	{"tempo.ErrQuantizedToZero", tempo.ErrQuantizedToZero},
	{"tempo.ErrTimeOverflow", tempo.ErrTimeOverflow},
	{"tempo.ErrInvalidText", tempo.ErrInvalidText},
	{"tempo.ErrOutOfRange", tempo.ErrOutOfRange},
}

// dependencySentinels are all sentinels of the packages that tempo imports
// directly or through duration. An error of tempo wraps exactly the
// dependency sentinels named by architecture 4.3 and no other.
var dependencySentinels = []namedError{
	{"term.ErrUnset", term.ErrUnset},
	{"term.ErrInvalidArgument", term.ErrInvalidArgument},
	{"term.ErrNoTempo", term.ErrNoTempo},
	{"term.ErrInvalidText", term.ErrInvalidText},
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

// assertErr checks the error class of architecture 4.3: errors.Is holds for
// the sentinel want of tempo and for exactly the dependency sentinels causes,
// and for no other sentinel of the package or its dependencies.
func assertErr(t *testing.T, err error, want error, causes ...error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := true

	for _, s := range tempoSentinels {
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

// errClass is the class of an error: the package sentinel and the wrapped
// dependency sentinels.
type errClass struct {
	want   error
	causes []error
}

// Error classes of architecture 4.3 and 4.14.
var (
	unsetClass    = errClass{want: tempo.ErrUnset}
	argClass      = errClass{want: tempo.ErrInvalidArgument}
	noTempoClass  = errClass{want: tempo.ErrNoTempo}
	noSolution    = errClass{want: tempo.ErrNoSolution}
	negTime       = errClass{want: tempo.ErrNegativeTime}
	offScale      = errClass{want: tempo.ErrOffScale}
	noRange       = errClass{want: tempo.ErrNoRange}
	quantZero     = errClass{want: tempo.ErrQuantizedToZero}
	overflow      = errClass{want: tempo.ErrTimeOverflow}
	textClass     = errClass{want: tempo.ErrInvalidText}
	fracRange     = errClass{want: tempo.ErrOutOfRange, causes: []error{fraction.ErrOutOfRange}}
	modeUnset     = errClass{want: tempo.ErrInvalidArgument, causes: []error{rounding.ErrUnset}}
	modeUnknown   = errClass{want: tempo.ErrInvalidArgument, causes: []error{rounding.ErrUnknownMode}}
	zeroDen       = errClass{want: tempo.ErrInvalidArgument, causes: []error{fraction.ErrZeroDenominator}}
	unitTupletErr = errClass{want: tempo.ErrOutOfRange, causes: []error{duration.ErrOutOfRange, tuplet.ErrOutOfRange}}
	unitChainErr  = errClass{
		want:   tempo.ErrOutOfRange,
		causes: []error{duration.ErrOutOfRange, tuplet.ErrOutOfRange, fraction.ErrOutOfRange},
	}
	unitValueErr = errClass{want: tempo.ErrOutOfRange, causes: []error{duration.ErrOutOfRange, fraction.ErrOutOfRange}}
)

// check asserts that err has the class c.
func (c errClass) check(t *testing.T, err error) bool {
	t.Helper()

	return assertErr(t, err, c.want, c.causes...)
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

// fastLimit bounds calls that must not do work proportional to the size of a
// huge input (architecture 4.14, 6, 7: parsing in O(len)).
const fastLimit = 5 * time.Second

// fast runs f in a goroutine and returns its result. If f does not return
// within fastLimit, the test fails at once.
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

// Short constructors of test values. They panic on misuse, which is a test
// error rather than a package error.

// fi returns the integer n as a fraction.
func fi(n int) fraction.Fraction { return fraction.FromInt(n) }

// fr returns the reduced fraction n/d.
func fr(n, d int) fraction.Fraction { return fraction.MustNewReduced(n, d) }

// fbig returns the reduced fraction num/den of big integers.
func fbig(num, den *big.Int) fraction.Fraction { return fraction.MustFromBigIntReduced(num, den) }

// maxF is fraction.MaxComponent as a fraction.
func maxF() fraction.Fraction { return fraction.MustParse(fraction.MaxComponentString) }

// invMaxF is 1/fraction.MaxComponent.
func invMaxF() fraction.Fraction { return fbig(big.NewInt(1), fraction.MaxComponentBig()) }

// maxPlusOne is the decimal text of fraction.MaxComponent + 1.
func maxPlusOne() string {
	return new(big.Int).Add(fraction.MaxComponentBig(), big.NewInt(1)).String()
}

// dur returns the duration n/d of whole notes.
func dur(n, d int) duration.Duration { return duration.MustNew(n, d) }

// durF returns the duration of a positive fraction.
func durF(f fraction.Fraction) duration.Duration { return duration.MustFromFraction(f) }

// sec returns the time n/d seconds.
func sec(n, d int) tempo.Seconds { return tempo.MustNewSeconds(n, d) }

// secF returns the time of a fraction of seconds.
func secF(f fraction.Fraction) tempo.Seconds { return tempo.MustSecondsFromFraction(f) }

// Written units of the scenarios (business-spec 3): ♩ 1/4, ♩. 3/8, ♪ 1/8,
// half 1/2, dotted half 3/4, eighth in a triplet 1/12, whole, breve.
var (
	q     = duration.Quarter.MustWritten()
	dq    = duration.MustNewWritten(duration.Quarter, 1)
	e     = duration.Eighth.MustWritten()
	h     = duration.Half.MustWritten()
	dh    = duration.MustNewWritten(duration.Half, 1)
	e3    = duration.MustNewWritten(duration.Eighth, 0, tuplet.MustNew(3, 2))
	whole = duration.Whole.MustWritten()
	breve = duration.Breve.MustWritten()
)

// tm returns the exact tempo unit = bpm.
func tm(unit duration.Written, bpm fraction.Fraction) tempo.Tempo { return tempo.MustNew(unit, bpm) }

// tq returns the exact tempo ♩ = n.
func tq(n int) tempo.Tempo { return tempo.MustNewQuarter(fi(n)) }

// ap returns the approximate tempo unit ≈ bpm.
func ap(unit duration.Written, bpm fraction.Fraction) tempo.Approx {
	return tempo.MustNewApprox(unit, bpm)
}

// rg returns the tempo range unit = lo-hi.
func rg(unit duration.Written, lo, hi fraction.Fraction) tempo.Range {
	return tempo.MustNewRange(unit, lo, hi)
}

// assertFrac checks that got is a valid fraction equal to want by value.
func assertFrac(t *testing.T, want, got fraction.Fraction, msgAndArgs ...any) bool {
	t.Helper()

	if !assert.True(t, got.IsValid(), msgAndArgs...) {
		return false
	}

	return assert.True(t, want.Equal(got), "want %s, got %s %v", want.ReducedString(), got.ReducedString(), msgAndArgs)
}

// assertDur checks that got is a set duration equal to want.
func assertDur(t *testing.T, want, got duration.Duration, msgAndArgs ...any) bool {
	t.Helper()

	return assert.True(t, want.Equal(got), "want %s, got %s %v", want, got, msgAndArgs)
}

// rep repeats u to n bytes for a one-byte u, or floor(n/len(u)) whole copies.
func rep(u string, n int) string { return strings.Repeat(u, n/len(u)) }

// allModes are the nine modes of common/rounding.
func allModes() []rounding.Mode { return rounding.Modes() }

// Sizes and limits of the error text of parsing (architecture 4.14, the rule
// of duration 4.12).
const (
	kib            = 1024
	mib            = 1024 * 1024
	excerptLimit   = 64
	dependentBytes = 67
	errorTextLimit = 512
)

// excerptLen returns the length of the longest prefix of whole UTF-8 decoding
// units of x that is at most 64 bytes.
func excerptLen(x string) int {
	if len(x) <= excerptLimit {
		return len(x)
	}

	k := 0
	for k < len(x) {
		_, size := utf8.DecodeRuneInString(x[k:])
		if k+size > excerptLimit {
			break
		}

		k += size
	}

	return k
}

// customList is a user list with names that need quotes in the text of an
// indication and names that do not (business-spec С-28; architecture 4.14).
func customList() term.List {
	names := []string{
		"4=60", "Скоро, но не слишком", `"Presto"`, "a≈b", "x=\t\\y", "q\"uote", "back\\slash", " spaced ",
		"line\nbreak", "≈", "=", "4≈108", "Tempo(unset)", "x, 4=120", "x, ", "ё", "\x00",
	}

	rows := make([]term.Row, 0, len(names))
	for _, n := range names {
		rows = append(rows, term.Row{Entries: []term.Entry{{Name: n}}})
	}

	return term.MustNewList(rows...)
}

// roundingMode and the nine modes of common/rounding, short names for tables.
type roundingMode = rounding.Mode

const (
	down             = rounding.Down
	up               = rounding.Up
	towardZero       = rounding.TowardZero
	awayFromZero     = rounding.AwayFromZero
	halfUp           = rounding.HalfUp
	halfDown         = rounding.HalfDown
	halfAwayFromZero = rounding.HalfAwayFromZero
	halfTowardZero   = rounding.HalfTowardZero
	halfEven         = rounding.HalfEven
)

// pow10 returns 10^n.
func pow10(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }
