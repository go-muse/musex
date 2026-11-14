package interval_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
)

// T-7, concurrency (NF-1, P-1; architecture 3.2): shared values, predefined
// intervals, systems, notes and sequences used from many goroutines give the
// results of a sequential run. Run with -race.

const workers = 16

// runShared runs compute once sequentially and then in many goroutines at
// once, and checks that every run gives the same results.
func runShared(t *testing.T, compute func() []any) {
	t.Helper()

	want := compute()
	results := make([][]any, workers)

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() { results[w] = compute() })
	}

	wg.Wait()

	for w, got := range results {
		assert.Equal(t, want, got, "worker %d", w)
	}
}

func TestConcurrencyIntervals(t *testing.T) {
	t.Parallel()

	shared := predefinedBoth()
	m3 := interval.MajorThird()

	runShared(t, func() []any {
		var out []any

		for _, i := range shared {
			s, _ := i.Simple()
			inv, invErr := i.Invert()
			sum, _ := i.Add(m3)
			name, ok, _ := i.Name()

			out = append(out, i.Size(), i.Deviation(), i.Quality(), i.IsBasic(), s, inv, invErr == nil, sum,
				i.Compare(m3), i.Equal(m3), i.String(), name, ok, interval.MustParse(i.String()))
		}

		out = append(out, interval.Predefined(), interval.PerfectFifth(), interval.MustParseName("ув.4"))

		return out
	})
}

func TestConcurrencySystems(t *testing.T) {
	t.Parallel()

	systems := []*system.System{trad(), pos19(), qt24(), irregular()}
	shared := predefinedBoth()
	starts := []spelling.Spelling{spelling.C(), spelling.FSharp()}

	runShared(t, func() []any {
		var out []any

		for _, sys := range systems {
			for _, x := range starts {
				p := wp(sys, x, 4)

				for _, i := range shared {
					st, err := i.Steps(sys, x)
					y, _ := interval.TransposePitch(sys, p, i)
					back, _ := interval.BetweenPitches(sys, p, y)
					out = append(out, st, err == nil, y, back, interval.MustTranspose(x, i))
				}
			}
		}

		return out
	})
}

func TestConcurrencySequences(t *testing.T) {
	t.Parallel()

	tr := trad()
	all := interval.MustAllFromSteps(tr, 6, 3, noStart)
	enh := iv("A4").MustEnharmonics(tr, 3, noStart)
	resp := interval.MustRespellings(tr, pt(spelling.C(), 4), pt(spelling.E(), 4), ai(2), interval.ReplaceBoth)

	runShared(t, func() []any {
		return []any{slices.Collect(all), slices.Collect(enh), slices.Collect(resp), slices.Collect(all)}
	})

	assert.NotEmpty(t, slices.Collect(all))
	assert.NotEmpty(t, slices.Collect(resp))
}

func TestConcurrencyNotes(t *testing.T) {
	t.Parallel()

	src := std()
	x := noteAt(t, pt(spelling.C(), 4), 30, src)
	before := x
	shared := predefinedBoth()

	runShared(t, func() []any {
		var out []any

		for _, i := range shared {
			y, err := interval.TransposeNote(x, i)

			var got interval.Interval
			if err == nil {
				got, _ = interval.BetweenNotes(x, y, interval.DirectionUnset)
			}

			out = append(out, err == nil, y.Spelling(), got)
		}

		return out
	})

	assert.True(t, x.Equal(before) && x.SameSource(before), "the shared note does not change")
}

func TestConcurrencyUnmarshal(t *testing.T) {
	t.Parallel()

	texts := []string{"M3", "-m6", "{+1/4}3", "dd5", "P15"}
	targets := make([]interval.Interval, workers)

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() { _ = targets[w].UnmarshalText([]byte(texts[w%len(texts)])) })
	}

	wg.Wait()

	for w, got := range targets {
		assert.True(t, got.Equal(interval.MustParse(texts[w%len(texts)])), "worker %d", w)
	}
}
