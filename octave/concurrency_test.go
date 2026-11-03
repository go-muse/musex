package octave_test

import (
	"fmt"
	"iter"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// concurrencyRange is the range of ASPN numbers -r...r used by the workers.
const concurrencyRange = 6

// sharedValues are created in the test goroutine before the workers start
// and are only read by them: octaves, built-in numberings, one catalog, one
// range sequence and the inputs of NewNumbering and NewCatalog.
type sharedValues struct {
	octaves    []octave.Octave // octaves[i] is O(i - r)
	numberings []numberingCase
	catalog    *naming.Catalog
	russian    map[int]string
	own        *naming.Catalog
	ownNames   map[octave.Octave]string
	skipped    []int
	seq        iter.Seq[octave.Octave]
	seqWant    []octave.Octave
}

func newSharedValues(t *testing.T) sharedValues {
	t.Helper()

	r := concurrencyRange
	s := sharedValues{
		octaves:    octs(span(-r, r)...),
		numberings: builtinNumberings(),
		catalog:    naming.Russian(),
		russian:    russianNames(),
		ownNames:   map[octave.Octave]string{oct(-2): "low", oct(4): "middle", oct(6): "high"},
		skipped:    []int{0},
		seqWant:    octs(span(-r, r)...),
	}

	var err error
	s.own, err = octave.NewCatalog(s.ownNames)
	require.NoError(t, err)
	seq, err := octave.Range(oct(-r), oct(r))
	require.NoError(t, err)
	s.seq = seq

	return s
}

// check runs every operation on the shared values once and returns the
// mismatches with the expected results. Expected values come from the
// oracles of the tests (formulas, literal names, exact arithmetic), not from
// earlier calls. It must not call testing functions: it runs in workers.
func (s sharedValues) check() []string {
	var bad []string
	report := func(format string, args ...any) {
		bad = append(bad, fmt.Sprintf(format, args...))
	}

	r := concurrencyRange
	for i, o := range s.octaves {
		a := i - r

		if !o.IsValid() {
			report("O(%d).IsValid() false", a)
		}
		if got, err := o.ASPN(); err != nil || got != a {
			report("O(%d).ASPN() = %d, %v", a, got, err)
		}
		for _, nc := range s.numberings {
			want, _ := nc.number(a)
			if got, err := o.Number(nc.n); err != nil || got != want {
				report("O(%d).Number(%s) = %d, %v; want %d", a, nc.name, got, err, want)
			}
			if got, err := octave.FromNumber(want, nc.n); err != nil || got != o {
				report("FromNumber(%d, %s) = %v, %v", want, nc.name, got, err)
			}
		}
		if got := o.Name(s.catalog); got != s.russian[a] {
			report("O(%d).Name(Russian) = %q", a, got)
		}
		if got := o.Name(s.own); got != s.ownNames[o] {
			report("O(%d).Name(own) = %q", a, got)
		}
		if name := s.russian[a]; name != "" {
			if got, err := octave.FromName(name, s.catalog); err != nil || got != o {
				report("FromName(%q) = %v, %v", name, got, err)
			}
		}
		if got := o.String(); got != "ASPN "+strconv.Itoa(a) {
			report("O(%d).String() = %q", a, got)
		}
		if got := o.GoString(); got != "octave.FromASPN("+strconv.Itoa(a)+")" {
			report("O(%d).GoString() = %q", a, got)
		}
		if got, err := o.Shift(1); err != nil || got != oct(a+1) {
			report("O(%d).Shift(1) = %v, %v", a, got, err)
		}
		if got, err := o.ShiftDown(1); err != nil || got != oct(a-1) {
			report("O(%d).ShiftDown(1) = %v, %v", a, got, err)
		}
		if got, err := o.Up(); err != nil || got != oct(a+1) {
			report("O(%d).Up() = %v, %v", a, got, err)
		}
		if got, err := o.Down(); err != nil || got != oct(a-1) {
			report("O(%d).Down() = %v, %v", a, got, err)
		}

		for j, p := range s.octaves {
			b := j - r
			if got := o.Equal(p); got != (a == b) {
				report("O(%d).Equal(O(%d)) = %t", a, b, got)
			}
			if got := o.Compare(p); got != sign(a-b) {
				report("O(%d).Compare(O(%d)) = %d", a, b, got)
			}
			if got := o.Above(p); got != (a > b) {
				report("O(%d).Above(O(%d)) = %t", a, b, got)
			}
			if got := o.Below(p); got != (a < b) {
				report("O(%d).Below(O(%d)) = %t", a, b, got)
			}
			if got := o.Between(p, oct(0)); got != (min(b, 0) <= a && a <= max(b, 0)) {
				report("O(%d).Between(O(%d), O(0)) = %t", a, b, got)
			}
			if got, err := o.Sub(p); err != nil || got != a-b {
				report("O(%d).Sub(O(%d)) = %d, %v", a, b, got, err)
			}
			if got, err := o.Distance(p); err != nil || got != max(a-b, b-a) {
				report("O(%d).Distance(O(%d)) = %d, %v", a, b, got, err)
			}
		}

		if n, err := octave.NewNumbering(o, a+100, s.skipped...); err != nil || !n.Equal(number.MustNew(a, a+100, s.skipped...)) {
			report("NewNumbering(O(%d), %d, 0) = %v, %v", a, a+100, n, err)
		}
	}

	if c, err := octave.NewCatalog(s.ownNames); err != nil || !c.Equal(s.own) {
		report("NewCatalog(shared map) = %v", err)
	}

	got := collect(s.seq, 100)
	if len(got) != len(s.seqWant) {
		report("shared Range: %d elements", len(got))
	} else {
		for i := range got {
			if got[i] != s.seqWant[i] {
				report("shared Range element %d = %v", i, got[i])
			}
		}
	}

	return bad
}

// K-28, NF-3: all operations called concurrently from many goroutines on the
// same values, including one *naming.Catalog, one range sequence and the same
// input map and slice, give the results of K-1...K-19. Data races are found
// when the test runs under go test -race, as in make test and CI. Workers
// report mismatches through their own slot; assertions run only in the test
// goroutine after Wait.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 3

	shared := newSharedValues(t)
	require.Empty(t, shared.check(), "sequential run before the workers")

	workers := runtime.GOMAXPROCS(0) * 4
	failures := make([][]string, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			<-start
			for range iterations {
				failures[w] = append(failures[w], shared.check()...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, f := range failures {
		assert.Empty(t, f, "worker %d", w)
	}
}
