package mode_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotationsSnapshotAndIndependentResults(t *testing.T) {
	t.Parallel()
	source := mode.MustFromName("Ionian")
	saved := source
	seq, err := mode.Rotations(source)
	require.NoError(t, err)
	require.NotNil(t, seq)
	changed, err := mode.FromDegree(source, 2)
	require.NoError(t, err)
	checkMode(t, changed, dorian)
	assert.Equal(t, "ionian", firstOf(changed.Origin()).Code())
	assert.Equal(t, "Ionian", firstOf(changed.SelectedName()))
	assert.True(t, changed.Changed())
	checkMode(t, source, ionian)
	require.NoError(t, source.RebuildFromCode("blues"))

	reached := 0
	for first := range seq {
		checkMode(t, first, ionian)
		require.NoError(t, first.Rotate(2))
		checkMode(t, first, dorian)
		reached++
		break
	}
	assert.Equal(t, 1, reached)
	want := []string{
		ionian, dorian, "m2 M2 M2 M2 m2 M2 M2", "M2 M2 M2 m2 M2 M2 m2",
		"M2 M2 m2 M2 M2 m2 M2", aeolian, "m2 M2 M2 m2 M2 M2 M2",
	}
	for range 2 {
		got := slices.Collect(seq)
		require.Len(t, got, 7)
		for i, candidate := range got {
			checkMode(t, candidate, want[i])
			assert.Equal(t, "ionian", firstOf(candidate.Origin()).Code())
			assert.Equal(t, "Ionian", firstOf(candidate.SelectedName()))
			assert.Equal(t, i > 0, candidate.Changed())
			candidate.MustMirror()
		}
		assert.True(t, got[0].Equal(saved))
	}
	periodic := mode.MustParse("M2 m2 M2 m2")
	got := slices.Collect(mode.MustRotations(periodic))
	require.Len(t, got, 4)
	for i, candidate := range got {
		checkMode(t, candidate, []string{"M2 m2 M2 m2", "m2 M2 m2 M2"}[i%2])
	}
	open := mode.MustParse("m3 M2 |")
	gotMode, err := mode.FromDegree(open, 1)
	require.NoError(t, err)
	assert.True(t, open.Equal(gotMode))
	gotMode, err = mode.FromDegree(open, 2)
	checkError(t, err, mode.ErrOpenMode)
	assert.Zero(t, gotMode)
	gotMode, err = mode.FromDegree(open, 4)
	checkError(t, err, mode.ErrInvalidArgument)
	assert.Zero(t, gotMode)
	gotMode, err = mode.FromDegree(mode.Mode{}, 0)
	checkError(t, err, mode.ErrUnset)
	assert.Zero(t, gotMode)
	seq, err = mode.Rotations(open)
	checkError(t, err, mode.ErrOpenMode)
	assert.Nil(t, seq)
	seq, err = mode.Rotations(mode.Mode{})
	checkError(t, err, mode.ErrUnset)
	assert.Nil(t, seq)
}

func TestIteratorCallbackPanicIsNotSwallowed(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("harmonic-minor")
	marker := new(int)
	a := mode.MustRotations(m)
	b := m.MustIntervals()
	c := m.MustChromaticIntervals()
	assert.Same(t, marker, capturePanic(func() { a(func(mode.Mode) bool { panic(marker) }) }))
	assert.Same(t, marker, capturePanic(func() { b(func(mode.IntervalAt) bool { panic(marker) }) }))
	assert.Same(t, marker, capturePanic(func() { c(func(mode.IntervalAt) bool { panic(marker) }) }))
	assert.Len(t, slices.Collect(a), 7)
	assert.Len(t, slices.Collect(b), 42)
	assert.Len(t, slices.Collect(c), 4)
}

func TestConcurrentReadsIteratorsAndIndependentCopies(t *testing.T) {
	t.Parallel()
	m := mode.MustFromName("harmonic minor")
	saved := m
	all := m.MustIntervals()
	chromatic := m.MustChromaticIntervals()
	rotations := mode.MustRotations(m)
	wantRows := []intervalRow{{3, 4, "A5"}, {6, 1, "A2"}, {7, 3, "d4"}, {7, 6, "d7"}}
	// All three iterators must retain the same state after the original moves.
	require.NoError(t, m.RebuildFromCode("dorian"))
	checkMode(t, m, dorian)
	workers := make([]struct {
		intervals, chromatic              []intervalRow
		firstRotation, readOwn, copyShape string
		rotCount                          int
		err                               error
	}, 8)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			<-start
			local := saved
			var result []mode.IntervalAt
			for r := range all {
				result = append(result, r)
			}
			workers[i].intervals = intervalRows(result)
			workers[i].chromatic = intervalRows(slices.Collect(chromatic))
			for candidate := range rotations {
				if workers[i].rotCount == 0 {
					workers[i].firstRotation = candidate.MustIntervalsString()
				}
				workers[i].rotCount++
				if err := candidate.Rotate(2); err != nil {
					workers[i].err = err
					return
				}
			}
			workers[i].readOwn = intervalText(saved.DegreeIntervals())
			if err := local.Mirror(); err != nil {
				workers[i].err = err
				return
			}
			workers[i].copyShape, workers[i].err = local.IntervalsString()
		})
	}
	close(start)
	wg.Wait()
	for i, worker := range workers {
		require.NoError(t, worker.err, "worker %d", i)
		require.Len(t, worker.intervals, 42)
		assert.Equal(t, intervalRow{1, 1, "M2"}, worker.intervals[0])
		assert.Equal(t, worker.intervals, workers[0].intervals)
		assert.Equal(t, wantRows, worker.chromatic)
		assert.Equal(t, 7, worker.rotCount)
		assert.Equal(t, "M2 m2 M2 M2 m2 A2 m2", worker.firstRotation)
		assert.Equal(t, "P1 M2 m3 P4 P5 m6 M7", worker.readOwn)
		assert.Equal(t, "m2 A2 m2 M2 M2 m2 M2", worker.copyShape)
	}
	checkMode(t, saved, "M2 m2 M2 M2 m2 A2 m2")
	checkMode(t, m, dorian)
}

func FuzzModeTextAndInverseRotation(f *testing.F) {
	f.Add([]byte{0, 0, 1, 3}, false, uint8(2))
	f.Add([]byte{1, 2, 1, 2, 0}, true, uint8(0))
	f.Add([]byte{}, false, uint8(0))
	f.Fuzz(func(t *testing.T, input []byte, open bool, first uint8) {
		if len(input) > 12 {
			input = input[:12]
		}
		alphabet := []string{"M2", "m2", "M3", "A1", "d2", "{+1/2}4"}
		text := "P4"
		if len(input) > 0 {
			text = ""
			for i, b := range input {
				if i > 0 {
					text += " "
				}
				text += alphabet[int(b)%len(alphabet)]
			}
		}
		if open {
			text += " |"
		}
		m, err := mode.Parse(text)
		require.NoErrorf(t, err, "generated %q", text)
		assert.Equal(t, text, m.MustIntervalsString())
		after := m
		require.NoError(t, after.Mirror())
		require.NoError(t, after.Mirror())
		assert.True(t, m.Equal(after))
		if !open && m.Len() > 1 {
			j := 2 + int(first)%(m.Len()-1)
			turned, err := mode.FromDegree(m, j)
			require.NoError(t, err)
			require.NoError(t, turned.Rotate(m.Len()-j+2))
			assert.True(t, m.Equal(turned))
		}
	})
}
