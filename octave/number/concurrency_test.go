package number_test

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
)

// concurrencyRange is the range of ASPN numbers -r...r used by the workers.
const concurrencyRange = 10

// expectations are the oracle results for the shared numberings, computed in
// the test goroutine before the workers start and only read by them.
type expectations struct {
	shared []numbering
	from   [][]int  // from[i][a+r]: number of ASPN a in shared[i]
	equal  [][]bool // equal[i][j]: shared[i] equals shared[j]
}

func newExpectations(shared []numbering) expectations {
	e := expectations{
		shared: shared,
		from:   make([][]int, len(shared)),
		equal:  make([][]bool, len(shared)),
	}
	for i, p := range shared {
		for a := -concurrencyRange; a <= concurrencyRange; a++ {
			// The values fit into int: K-18 checks them without errors.
			e.from[i] = append(e.from[i], p.spec.fromASPN(a).value)
		}
		for _, q := range shared {
			e.equal[i] = append(e.equal[i], equalSpec(p.spec, q.spec))
		}
	}

	return e
}

// K-17, NF-3: all operations called concurrently from many goroutines on the
// same values give the results of K-1...K-9. Data races are detected when the
// test runs under go test -race, as in make test and CI. The shared values are
// created before the goroutines start; workers report mismatches through their
// own slot, and assertions run only in the test goroutine.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 3

	want := newExpectations(criteriaSet(t))
	workers := runtime.GOMAXPROCS(0) * 4
	failures := make([][]string, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			// A panic in a worker is recorded in its slot instead of
			// stopping the whole test binary.
			defer func() {
				if r := recover(); r != nil {
					failures[w] = append(failures[w], fmt.Sprintf("panic: %v", r))
				}
			}()

			<-start
			for range iterations {
				failures[w] = append(failures[w], checkOperations(want)...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, f := range failures {
		assert.Empty(t, f, "worker %d", w)
	}
}

// checkOperations calls every public operation on the shared numberings and
// on new values, compares the results with the precomputed expectations and
// changes the slices it owns. It returns a description of every mismatch.
func checkOperations(want expectations) []string {
	var failures []string
	fail := func(format string, args ...any) {
		failures = append(failures, fmt.Sprintf(format, args...))
	}

	table := builtinTable()
	all := number.Builtins()
	if len(all) != len(table) {
		fail("len(Builtins()) = %d", len(all))
	} else {
		for i, b := range table {
			if all[i].Label() != b.label || !all[i].Equal(b.get()) {
				fail("Builtins()[%d] = %s", i, all[i])
			}
		}
	}
	clear(all)

	for i, b := range table {
		n := b.get()
		if n.Label() != b.label || n.String() != b.label {
			fail("%s: Label %q, String %q", b.label, n.Label(), n.String())
		}
		for j, c := range table {
			if got := n.Equal(c.get()); got != builtinsEqual(i, j) {
				fail("%s.Equal(%s) = %t", b.label, c.label, got)
			}
		}
	}

	if m, err := number.New(4, 4, 1, 0); err != nil || m.String() != "ASPN 4 -> 4, skip [0 1]" || m.Label() != "" {
		fail("New(4, 4, 1, 0) = %s, %v", m, err)
	}
	if m := number.MustNew(4, 3, 0); !m.Equal(number.FrancoBelgian()) {
		fail("MustNew(4, 3, 0) is not equal to Franco-Belgian")
	}
	if _, err := number.New(4, 3, 0, 0); !isOnly(err, number.ErrInvalidDefinition) {
		fail("New(4, 3, 0, 0) error = %v", err)
	}
	if _, err := number.FrancoBelgian().ToASPN(0); !isOnly(err, number.ErrSkipped) {
		fail("FrancoBelgian().ToASPN(0) error = %v", err)
	}

	for i, p := range want.shared {
		if !p.n.IsDefined() {
			fail("%s.IsDefined() = false", p.name)
		}
		_ = p.n.Label()
		_ = p.n.String()

		for k, x := range want.from[i] {
			a := k - concurrencyRange

			if got, err := p.n.FromASPN(a); err != nil || got != x {
				fail("%s.FromASPN(%d) = %d, %v; want %d", p.name, a, got, err, x)
			}
			if got := p.n.MustFromASPN(a); got != x {
				fail("%s.MustFromASPN(%d) = %d; want %d", p.name, a, got, x)
			}
			if got, err := p.n.ToASPN(x); err != nil || got != a {
				fail("%s.ToASPN(%d) = %d, %v; want %d", p.name, x, got, err, a)
			}
			if got := p.n.MustToASPN(x); got != a {
				fail("%s.MustToASPN(%d) = %d; want %d", p.name, x, got, a)
			}

			for j, q := range want.shared {
				y := want.from[j][k]
				if got, err := number.Convert(x, p.n, q.n); err != nil || got != y {
					fail("Convert(%d, %s, %s) = %d, %v; want %d", x, p.name, q.name, got, err, y)
				}
				if got := number.MustConvert(x, p.n, q.n); got != y {
					fail("MustConvert(%d, %s, %s) = %d; want %d", x, p.name, q.name, got, y)
				}
			}
		}

		for j, q := range want.shared {
			if got := p.n.Equal(q.n); got != want.equal[i][j] {
				fail("%s.Equal(%s) = %t", p.name, q.name, got)
			}
		}
	}

	return failures
}

// isOnly reports whether err matches want and none of the other sentinels.
func isOnly(err, want error) bool {
	for _, e := range allErrors() {
		if errors.Is(err, e) != (e == want) {
			return false
		}
	}

	return true
}
