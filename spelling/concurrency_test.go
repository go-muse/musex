package spelling_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

// exerciseResult is what one goroutine observed: a record of all results and
// the descriptions of the results that disagree with the oracle.
type exerciseResult struct {
	record     []string
	mismatches []string
}

// nameAgrees reports whether a result of a name operation agrees with the
// result of the same alteration operation: both fail, and the name is the
// unset name, or both succeed, and the name has the letter l and the value want.
func nameAgrees(got spelling.Spelling, err error, l letter.Letter, want alteration.Alteration, wantErr error) bool {
	if (err == nil) != (wantErr == nil) {
		return false
	}

	if err != nil {
		return got == spelling.Spelling{}
	}

	return got.Letter() == l && got.Alteration().Equal(want)
}

// exercise calls every operation of the package on the shared values names
// and values. Receivers of the reading methods are local to the call.
func exercise(names []nameCase, shared []spelling.Spelling, values []altCase) exerciseResult {
	var r exerciseResult

	add := func(v ...any) { r.record = append(r.record, fmt.Sprint(v...)) }
	mismatch := func(format string, v ...any) { r.mismatches = append(r.mismatches, fmt.Sprintf(format, v...)) }

	for _, p := range predefinedCases() {
		s := p.fn()
		add(p.name, s)

		if s.Letter() != p.l || !s.Alteration().Equal(p.a) {
			mismatch("%s() is %v", p.name, s)
		}
	}

	add(spelling.Predefined())

	for i, x := range shared {
		c := names[i]

		l, a := x.Parts()
		add(x, x.Letter(), x.Alteration(), l, a, x.IsValid(), x.IsNatural(), x.IsInteger(), x.Direction(), x.Natural())

		if l != c.l || !a.Equal(c.a) || x.IsNatural() != c.a.IsNatural() || x.IsInteger() != c.a.IsInteger() {
			mismatch("parts or checks of %s", c.name)
		}

		text, err := x.MarshalText()
		add(string(text), err)
		appended, err := x.AppendText(nil)
		add(string(appended), err)
		data, err := x.MarshalJSON()
		add(string(data), err)

		if x.String() != formOf(c.l, c.a) {
			mismatch("form of %s is %s", c.name, x.String())
		}

		parsed, err := spelling.Parse(x.String())
		add(parsed, err)

		if !parsed.Equal(x) {
			mismatch("Parse of %s", c.name)
		}

		var fromText, fromJSON spelling.Spelling

		textErr := fromText.UnmarshalText(text)
		jsonErr := fromJSON.UnmarshalJSON(data)
		add(textErr, fromText, jsonErr, fromJSON)

		if !fromText.Equal(x) || !fromJSON.Equal(x) {
			mismatch("reading of %s", c.name)
		}

		up, err := x.Raise()
		add(up, err)

		if want, wantErr := c.a.Raise(); !nameAgrees(up, err, c.l, want, wantErr) {
			mismatch("Raise of %s", c.name)
		}

		down, err := x.Lower()
		add(down, err)

		for n := -2; n <= 2; n++ {
			byN, err := x.RaiseBy(n)
			add(byN, err)

			if want, wantErr := c.a.RaiseBy(n); !nameAgrees(byN, err, c.l, want, wantErr) {
				mismatch("RaiseBy(%d) of %s", n, c.name)
			}

			byN, err = x.LowerBy(n)
			add(byN, err)
		}

		for _, d := range values {
			add(x.WithAlteration(d.a))

			shifted, err := x.RaiseByAlteration(d.a)
			add(shifted, err)

			if want, wantErr := c.a.Add(d.a); !nameAgrees(shifted, err, c.l, want, wantErr) {
				mismatch("RaiseByAlteration(%s) of %s", d.name, c.name)
			}

			shifted, err = x.LowerByAlteration(d.a)
			add(shifted, err)
		}

		for j, y := range shared {
			diff, err := x.AlterationDifference(y)
			add(x.Equal(y), x.Compare(y), x.SameLetter(y), x.SameAlteration(y),
				x.LetterDistanceUp(y), x.LetterDistanceDown(y), diff, err)

			if x.Compare(y) != compareOracle(c, names[j]) {
				mismatch("Compare of %s and %s", c.name, names[j].name)
			}

			if want, wantErr := c.a.Sub(names[j].a); (err == nil) != (wantErr == nil) || !diff.Equal(want) {
				mismatch("AlterationDifference of %s and %s", c.name, names[j].name)
			}
		}

		created, err := spelling.New(c.l, c.a)
		add(created, err)
	}

	return r
}

// K-14: all operations used at the same time from many goroutines on shared
// values give the same results as sequential use and agree with the oracle.
// Run with the race detector (make test, CI).
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	names := concatNames(setN(), boundaryNames())
	shared := make([]spelling.Spelling, 0, len(names))

	for _, c := range names {
		shared = append(shared, c.build())
	}

	values := setS()

	want := exercise(names, shared, values)
	assert.Empty(t, want.mismatches, "sequential use must agree with the oracle")

	workers := runtime.GOMAXPROCS(0) * 4
	results := make([]exerciseResult, workers)

	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			results[i] = exercise(names, shared, values)
		})
	}

	wg.Wait()

	for i, got := range results {
		assert.Emptyf(t, got.mismatches, "goroutine %d must agree with the oracle", i)
		assert.Equalf(t, want.record, got.record, "goroutine %d must observe the sequential results", i)
	}

	for i, c := range names {
		assert.Truef(t, shared[i].Letter() == c.l && shared[i].Alteration().Equal(c.a), "the shared value %s must not change", c.name)
	}

	assert.Equal(t, letter.C, spelling.CSharp().Letter())
	assert.True(t, spelling.CSharp().Alteration().Equal(alteration.Sharp()))
}
