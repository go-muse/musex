package note_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Concurrent use (architecture 6.1.1; P-10, NF-2; G-6, G-7, G-8). Run with
// -race: reading one note from many goroutines while its source is retuned,
// changing separate copies, iterating one sequence, comparing and sorting
// with shared arguments.

const goroutines = 16

// TestConcurrentReads reads one note from many goroutines while the source
// switches between A4 = 440 Hz and A4 = 415 Hz: each result belongs to one of
// the two sets, never to a mixture.
func TestConcurrentReads(t *testing.T) {
	t.Parallel()

	a440, a415 := std(), stdAt(415)
	src := newSwitch(a440)
	n := mk(t, spelling.A(), 4, ct(0), src)
	other := mk(t, spelling.C(), 5, ct(0), src)
	before := n

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Go(func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if i%2 == 0 {
				src.set(a415)
			} else {
				src.set(a440)
			}
		}
	})

	results := make(chan float64, goroutines*50)
	var readers sync.WaitGroup
	for range goroutines {
		readers.Go(func() {
			for range 50 {
				f, err := n.Frequency()
				if !assert.NoError(t, err) {
					return
				}
				results <- f
				_ = n.String()
				_, _ = n.Step()
				_, _ = n.CentsFromReference()
				_, _ = n.Format()
				_, _ = n.ComparePosition(other)
				_, _ = n.CompareSound(other)
				_, _ = n.CentsTo(other)
				_, _, _ = n.Offset(440)
				_ = n.Equal(other)
				_, _ = n.Settings()
			}
		})
	}
	readers.Wait()
	close(stop)
	wg.Wait()
	close(results)

	for f := range results {
		assert.True(t, f == 440 || f == 415, "a whole set: %v", f)
	}
	assertSame(t, before, n, "reading does not change the note")
}

// TestConcurrentCopies changes separate copies of one note at the same time
// (G-8): each copy ends in its own state, the original is unchanged.
func TestConcurrentCopies(t *testing.T) {
	t.Parallel()

	s := std()
	n := mk(t, spelling.C(), 4, ct(30), s)
	before := n
	got := make([]note.Note, goroutines)
	var wg sync.WaitGroup
	for i := range goroutines {
		m := n
		wg.Go(func() {
			m.MustShift(i).MustSetCents(ct(i))
			if i%2 == 1 {
				m.MustRaise()
			}
			got[i] = m
		})
	}
	wg.Wait()
	assertSame(t, before, n)
	for i, m := range got {
		name := spelling.C()
		if i%2 == 1 {
			name = spelling.CSharp()
		}
		assertNote(t, wp(name, 4+i, ct(i)).boundTo(s), m, "copy %d", i)
	}
}

// TestConcurrentSequences iterates one sequence from many goroutines (6.9).
func TestConcurrentSequences(t *testing.T) {
	t.Parallel()

	s := std()
	from := mk(t, spelling.C(), 4, ct(0), s)
	to := mk(t, spelling.C(), 6, ct(0), s)
	walk := from.MustWalk(to, system.Sharps())
	enh := from.MustEnharmonics(alteration.DoubleSharp())
	rng := from.MustRange(oct(0), oct(9))
	wantWalk, wantEnh, wantRange := collect(walk), collect(enh), collect(rng)

	var wg sync.WaitGroup
	ok := make([]bool, goroutines)
	for i := range goroutines {
		wg.Go(func() {
			ok[i] = sameNotes(wantWalk, collect(walk)) && sameNotes(wantEnh, collect(enh)) && sameNotes(wantRange, collect(rng))
		})
	}
	wg.Wait()
	for i := range ok {
		assert.True(t, ok[i], "goroutine %d", i)
	}
}

// TestConcurrentComparisons compares and sorts with shared notes and a shared
// input slice (NF-2): the arguments and the input are unchanged.
func TestConcurrentComparisons(t *testing.T) {
	t.Parallel()

	src := newSwitch(std())
	in := []note.Note{
		mk(t, spelling.E(), 4, ct(0), src),
		mk(t, spelling.C(), 4, ct(10), src),
		mk(t, spelling.D(), 4, ct(0), std()),
		mk(t, spelling.C(), 4, ct(0), set19()),
	}
	snapshot := append([]note.Note(nil), in...)
	wantPos, err := note.SortedByPosition(in[:3])
	require.NoError(t, err)
	wantSound, err := note.SortedBySound(in)
	require.NoError(t, err)

	var wg sync.WaitGroup
	ok := make([]bool, goroutines)
	for i := range goroutines {
		wg.Go(func() {
			pos, err1 := note.SortedByPosition(in[:3])
			sound, err2 := note.SortedBySound(in)
			_, err3 := in[0].ComparePosition(in[1])
			_, err4 := in[0].CompareSound(in[3])
			_, err5 := in[2].Between(in[0], in[1])
			ok[i] = err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil &&
				sameNotes(wantPos, pos) && sameNotes(wantSound, sound)
		})
	}
	wg.Wait()
	for i := range ok {
		assert.True(t, ok[i], "goroutine %d", i)
	}
	assert.True(t, sameNotes(snapshot, in), "the input slice and its notes are unchanged")
}
