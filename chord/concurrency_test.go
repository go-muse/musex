package chord_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, concurrency: НФ-2, П-14, Г-7, Г-8; architecture 5.13. The checks are
// meaningful under the race detector (go test -race).

const workers = 8

// readAll calls every reading method of c and returns a description of the
// results that can be compared between goroutines.
type readResult struct {
	valid, empty, repeats, withOct, withoutOct bool
	length, count                              int
	notes, sorted, distinct, all, values, fn   []note.Note
	at                                         note.Note
	found                                      int
	indices, fnIdx                             []int
	contains, all2, any2, rec, spell, fnAny    bool
	equal, inOrder, same, incl, common         bool
	cmp, fnCount                               int
	str                                        string
}

func readAll(c chord.Chord, other chord.Chord) readResult {
	var r readResult
	r.valid, r.empty, r.length = c.IsValid(), c.IsEmpty(), c.Len()
	r.notes, r.sorted, r.distinct = c.Notes(), c.SortedNotes(), c.DistinctNotes()
	r.repeats, r.withOct, r.withoutOct = c.HasRepeats(), c.HasNotesWithOctave(), c.HasNotesWithoutOctave()
	for _, n := range c.All() {
		r.all = append(r.all, n)
	}
	for n := range c.Values() {
		r.values = append(r.values, n)
	}
	r.at, _ = c.At(1)
	r.contains, r.count = c.Contains(c4), c.Count(c4)
	r.found, _, _ = c.Find(e4)
	r.indices = c.Indices(c4)
	r.all2, r.any2 = c.ContainsAll(c4, e4), c.ContainsAny(d4, e4)
	r.rec, r.spell = c.ContainsRecord(c4p5), c.ContainsSpelling(c5)
	r.fnAny, _ = c.ContainsFunc(hasOctave)
	r.fnCount, _ = c.CountFunc(hasOctave)
	r.fnIdx, _ = c.IndicesFunc(hasOctave)
	r.fn, _ = c.NotesFunc(hasOctave)
	r.equal, r.inOrder, r.same = c.Equal(other), c.EqualInOrder(other), c.SameRecords(other)
	r.incl, r.common, r.cmp = c.Includes(other), c.HasCommonNotes(other), c.Compare(other)
	r.str = c.String()
	return r
}

// assertSameRead compares two results of readAll.
func assertSameRead(t *testing.T, want, got readResult) bool {
	t.Helper()
	ok := assert.Equal(t, []bool{want.valid, want.empty, want.repeats, want.withOct, want.withoutOct, want.contains, want.all2, want.any2, want.rec, want.spell, want.fnAny, want.equal, want.inOrder, want.same, want.incl, want.common},
		[]bool{got.valid, got.empty, got.repeats, got.withOct, got.withoutOct, got.contains, got.all2, got.any2, got.rec, got.spell, got.fnAny, got.equal, got.inOrder, got.same, got.incl, got.common})
	ok = assert.Equal(t, []int{want.length, want.count, want.found, want.cmp, want.fnCount}, []int{got.length, got.count, got.found, got.cmp, got.fnCount}) && ok
	ok = assert.Equal(t, want.indices, got.indices) && ok
	ok = assert.Equal(t, want.fnIdx, got.fnIdx) && ok
	ok = assert.Equal(t, want.str, got.str) && ok
	ok = assert.True(t, sameNote(want.at, got.at)) && ok
	for name, pair := range map[string][2][]note.Note{
		"Notes": {want.notes, got.notes}, "SortedNotes": {want.sorted, got.sorted}, "DistinctNotes": {want.distinct, got.distinct},
		"All": {want.all, got.all}, "Values": {want.values, got.values}, "NotesFunc": {want.fn, got.fn},
	} {
		ok = assertNotes(t, pair[0], pair[1], name) && ok
	}
	return ok
}

func TestConcurrencyReaders(t *testing.T) {
	t.Parallel()
	// One chord read by many goroutines at once.
	st := std()
	c := mk(t, c4, pn(e4.Spelling(), 4, 0, st), cNo, c4, c4p5, g4)
	other := mk(t, e4, c4, g4)
	want := readAll(c, other)
	before := stateOf(c)

	results := make([]readResult, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			for range 20 {
				results[i] = readAll(c, other)
			}
		})
	}
	wg.Wait()
	for i := range workers {
		assertSameRead(t, want, results[i])
	}
	assertState(t, before, c)
}

func TestConcurrencyCopies(t *testing.T) {
	t.Parallel()
	// Different copies, made right before the goroutines start, changed at
	// once; the original changes too.
	base := list(e4, c4, pn(g4.Spelling(), 4, 0, std()), c4)
	c := mk(t, base...)
	copies := make([]chord.Chord, workers)
	for i := range copies {
		copies[i] = c
	}
	errs := make([][]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			d := &copies[i]
			errs[i] = []error{
				d.Add(b4),
				d.Insert(0, d4),
				d.Remove(c4),
				d.ReplaceAt(1, f4),
				d.Sort(),
				d.Reverse(),
				d.Up(),
				d.RemoveRepeats(),
				d.Union(chord.MustNew(c5, c5)),
				d.Subtract(chord.MustNew(b4)),
			}
		})
	}
	wg.Go(func() {
		for range 50 {
			_ = c.Add(b4)
			_ = c.RemoveAt(c.Len() - 1)
			_ = c.Reverse()
		}
	})
	wg.Wait()

	// The expected result of the sequence for one copy, computed alone.
	want := mk(t, base...)
	for _, err := range []error{
		want.Add(b4), want.Insert(0, d4), want.Remove(c4), want.ReplaceAt(1, f4), want.Sort(), want.Reverse(),
		want.Up(), want.RemoveRepeats(), want.Union(chord.MustNew(c5, c5)), want.Subtract(chord.MustNew(b4)),
	} {
		require.NoError(t, err)
	}
	for i := range workers {
		for _, err := range errs[i] {
			require.NoError(t, err)
		}
		assertNotes(t, want.Notes(), copies[i].Notes(), "copy %d", i)
	}
	assertChord(t, base, c, "the original after 50 times add, remove the last, reverse")
}

func TestConcurrencyReadWhileChanging(t *testing.T) {
	t.Parallel()
	// A slice from Notes and the sequences from All and Values, got before
	// the goroutine that changes the chord starts, are read while it runs.
	base := list(c4, e4, g4, c4)
	c := mk(t, base...)
	notes := c.Notes()
	all := c.All()
	vals := c.Values()

	var wg sync.WaitGroup
	wg.Go(func() {
		for range 50 {
			_ = c.Add(b4)
			_ = c.ReplaceAt(0, d4)
			_ = c.Sort()
			_ = c.Up()
			_ = c.RemoveAt(0)
		}
	})
	var gotAll, gotVals []note.Note
	for range 50 {
		gotAll, gotVals = gotAll[:0], gotVals[:0]
		for _, n := range all {
			gotAll = append(gotAll, n)
		}
		for n := range vals {
			gotVals = append(gotVals, n)
		}
		for i := range notes {
			_ = notes[i].String()
		}
	}
	wg.Wait()
	assertNotes(t, base, notes, "Notes")
	assertNotes(t, base, gotAll, "All")
	assertNotes(t, base, gotVals, "Values")
}

func TestConcurrencyRules(t *testing.T) {
	t.Parallel()
	// Methods with a safe rule on one chord from many goroutines.
	c := mk(t, c4, eNo, g4, c4)
	counts := make([]int, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			n, err := c.CountFunc(hasOctave)
			if err == nil {
				counts[i] = n
			}
		})
	}
	wg.Wait()
	for i := range workers {
		assert.Equal(t, 3, counts[i], "worker %d", i)
	}
}
