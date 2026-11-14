package term_test

import (
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
)

// namedError is a sentinel error with its qualified name for diagnostics.
type namedError struct {
	name string
	err  error
}

// termSentinels are the four sentinels of architecture 4.1.
var termSentinels = []namedError{
	{"term.ErrUnset", term.ErrUnset},
	{"term.ErrInvalidArgument", term.ErrInvalidArgument},
	{"term.ErrNoTempo", term.ErrNoTempo},
	{"term.ErrInvalidText", term.ErrInvalidText},
}

// fractionSentinels are the sentinels of common/fraction, the only musex
// import of tempo/term. No error of tempo/term wraps any of them
// (architecture 4.3: the "additionally" column of the term table is empty).
var fractionSentinels = []namedError{
	{"fraction.ErrInvalid", fraction.ErrInvalid},
	{"fraction.ErrZeroDenominator", fraction.ErrZeroDenominator},
	{"fraction.ErrDivisionByZero", fraction.ErrDivisionByZero},
	{"fraction.ErrNotRepresentable", fraction.ErrNotRepresentable},
	{"fraction.ErrOutOfRange", fraction.ErrOutOfRange},
	{"fraction.ErrInvalidText", fraction.ErrInvalidText},
}

// assertErr checks the error class of architecture 4.3: errors.Is holds for
// the sentinel want of tempo/term and for no other sentinel of the package or
// of common/fraction.
func assertErr(t *testing.T, err error, want error) bool {
	t.Helper()

	if !assert.Error(t, err) {
		return false
	}

	ok := true

	for _, s := range termSentinels {
		expected := s.err == want
		if errors.Is(err, s.err) != expected {
			ok = false

			t.Errorf("errors.Is(err, %s) = %t, want %t; err: %.200s", s.name, !expected, expected, err)
		}
	}

	for _, s := range fractionSentinels {
		if errors.Is(err, s.err) {
			ok = false

			t.Errorf("errors.Is(err, %s) = true, want false; err: %.200s", s.name, err)
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

// fastLimit bounds calls that must not do work proportional to the size of a
// huge input.
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

// maxF is fraction.MaxComponent as a fraction.
func maxF() fraction.Fraction { return fraction.MustParse(fraction.MaxComponentString) }

// invMaxF is 1/fraction.MaxComponent.
func invMaxF() fraction.Fraction {
	return fraction.MustFromBigIntReduced(big.NewInt(1), fraction.MaxComponentBig())
}

// between, atLeast, and atMost return guides of integer bounds.
func between(lo, hi int) term.Guide { return term.MustBetween(fi(lo), fi(hi)) }

func atLeast(lo int) term.Guide { return term.MustAtLeast(fi(lo)) }

func atMost(hi int) term.Guide { return term.MustAtMost(fi(hi)) }

// guideOf returns the guide of integer bounds where 0 means an open bound.
func guideOf(lo, hi int) term.Guide {
	switch {
	case lo == 0:
		return atMost(hi)
	case hi == 0:
		return atLeast(lo)
	default:
		return between(lo, hi)
	}
}

// guidePtr returns a pointer to a new copy of g.
func guidePtr(g term.Guide) *term.Guide { return &g }

// names returns the names of terms, nil for nil.
func names(t *testing.T, terms []term.Term) []string {
	t.Helper()

	if terms == nil {
		return nil
	}

	out := make([]string, 0, len(terms))

	for _, w := range terms {
		n, err := w.Name()
		assert.NoError(t, err)

		out = append(out, n)
	}

	return out
}

// rep repeats u to n bytes for a one-byte u, or floor(n/len(u)) whole copies.
func rep(u string, n int) string { return strings.Repeat(u, n/len(u)) }

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
// units of x that is at most 64 bytes. A unit is a valid rune or a single
// byte that does not start one.
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

// customRows are the rows of the user list of scenario С-26.
func customRows() []term.Row {
	return []term.Row{
		{Entries: []term.Entry{{Name: "Медленно"}}, Group: term.Slow, Guide: guidePtr(between(40, 60))},
		{Entries: []term.Entry{{Name: "Ровно"}}, Group: term.NoGroup, Guide: guidePtr(between(100, 100))},
		{
			Entries: []term.Entry{{Name: "Скоро"}, {Name: "Живо"}},
			Group:   term.Fast,
			Guide:   guidePtr(between(120, 160)),
		},
	}
}

// customList is the user list of scenario С-26.
func customList() term.List { return term.MustNewList(customRows()...) }
