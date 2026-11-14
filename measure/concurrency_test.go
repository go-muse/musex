package measure_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/measure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-9, concurrency (architecture 3.2; the nonfunctional requirement of
// safe concurrent reading): the same grids, addresses, and sequences of bars
// are used from many goroutines under -race and give the same results.

// snapshot is a set of results of read operations on one grid.
type snapshot struct {
	text    string
	bars    []measure.Bar
	address measure.Address
	start   string
	parts   []string
	count   string
	markup  bool
	sorted  int
}

// takeSnapshot reads g, seq, and a without changing them. Errors become
// part of the snapshot text so that goroutines do not fail the test.
func takeSnapshot(g measure.Grid, seq func(func(measure.Bar) bool), a measure.Address) snapshot {
	var s snapshot

	s.text = g.String()
	s.bars = collect(seq)

	if loc, err := g.Locate(pos("17/8")); err == nil {
		s.address = loc
	}

	if p, err := g.PositionOf(a); err == nil {
		s.start = p.String()
	} else {
		s.start = err.Error()
	}

	if parts, err := g.SplitBy(pos("1/8"), dur("3"), measure.LevelUnits); err == nil {
		s.parts = partTexts(parts)
	}

	if c, err := g.MetricCountAt(0, pos("1/8")); err == nil {
		s.count = c.String()
	}

	if ok, err := g.EqualMarkup(g); err == nil {
		s.markup = ok
	}

	s.sorted = a.Compare(measure.MustNewAddress(1, pos("0")))

	if p, err := measure.Parse(s.text); err == nil && !p.Equal(g) {
		s.text += " (parse differs)"
	}

	return s
}

// TestConcurrency covers architecture 3.2: concurrent reads of shared grids,
// addresses, and sequences give the results of sequential reads.
func TestConcurrency(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "4/4;pickup:1/4;3:3/4;4:length:1/2:end;6:2/4+3/8;bars:12")
	a := ad(3, "1/4")

	seq, err := g.Bars(0, 12)
	require.NoError(t, err)

	want := takeSnapshot(g, seq, a)
	require.Len(t, want.bars, 13)

	const workers = 16

	results := make([]snapshot, workers*4)

	var wg sync.WaitGroup

	for i := range results {
		wg.Go(func() {
			defer func() { _ = recover() }()

			results[i] = takeSnapshot(g, seq, a)
		})
	}

	wg.Wait()

	for i, got := range results {
		assert.Equal(t, want, got, "goroutine %d", i)
	}

	assert.Equal(t, "4/4;pickup:1/4;3:3/4;4:length:1/2:end;6:2/4+3/8;bars:12", g.String(), "the shared grid is unchanged")
}

// TestConcurrencyChanges covers architecture 3.2: concurrent changes of the
// same grid return new grids and leave the shared one unchanged.
func TestConcurrencyChanges(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "4/4;3:3/4;bars:8")

	const workers = 32

	texts := make([]string, workers)

	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			defer func() { _ = recover() }()

			next, err := g.WithMark(4+i%4, mt("2/4"))
			if err != nil {
				texts[i] = err.Error()

				return
			}

			next, err = next.InsertBars(2, 1+i%3)
			if err != nil {
				texts[i] = err.Error()

				return
			}

			texts[i] = next.String()
		})
	}

	wg.Wait()

	for i, got := range texts {
		bar := 4 + i%4
		m := 1 + i%3

		next, err := g.WithMark(bar, mt("2/4"))
		require.NoError(t, err)

		next, err = next.InsertBars(2, m)
		require.NoError(t, err)
		assert.Equal(t, next.String(), got, "goroutine %d", i)
	}

	assert.Equal(t, "4/4;3:3/4;bars:8", g.String())
}
