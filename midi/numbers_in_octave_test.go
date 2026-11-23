package midi_test

import (
	"fmt"
	"iter"
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/midi"
	"github.com/go-muse/musex/octave"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collectNumbers drains a sequence into the integer values of its numbers.
func collectNumbers(t *testing.T, seq iter.Seq[midi.Number]) []int {
	t.Helper()

	values := make([]int, 0)
	for number := range seq {
		value, err := number.Int()
		require.NoError(t, err)
		values = append(values, value)
	}

	return values
}

// NumbersInOctave enumerates the valid MIDI numbers of a positional octave
// in ascending order: full octaves give twelve numbers, the top octave is
// partial, and a valid octave outside the MIDI range gives an empty result.
func TestNumbersInOctave(t *testing.T) {
	t.Parallel()

	cases := []struct {
		aspn int
		want []int
	}{
		{aspn: -1, want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
		{aspn: 0, want: []int{12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}},
		{aspn: 4, want: []int{60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71}},
		{aspn: 8, want: []int{108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119}},
		{aspn: 9, want: []int{120, 121, 122, 123, 124, 125, 126, 127}},
		{aspn: -2, want: []int{}},
		{aspn: 10, want: []int{}},
		{aspn: math.MinInt, want: []int{}},
		{aspn: math.MaxInt, want: []int{}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("ASPN_%d", tc.aspn), func(t *testing.T) {
			t.Parallel()

			o := octave.FromASPN(tc.aspn)
			seq, err := midi.NumbersInOctave(o)
			require.NoError(t, err)
			require.NotNil(t, seq, "a valid octave gives a sequence, possibly empty")

			got := collectNumbers(t, seq)
			assert.Equal(t, tc.want, got)

			for class, value := range got {
				number := midi.MustNew(value)
				partsOctave, partsClass, _, err := number.Parts()
				require.NoError(t, err)
				assert.True(t, partsOctave.Equal(o), "MIDI %d octave", value)
				assert.Equal(t, class, partsClass.Number(), "MIDI %d class", value)
			}
		})
	}
}

// The invalid octave is an error, not an empty result: it matches ErrUnset
// and octave.ErrInvalid, and the Must variant panics with the same error.
func TestNumbersInOctaveInvalid(t *testing.T) {
	t.Parallel()

	var o octave.Octave
	seq, err := midi.NumbersInOctave(o)
	assert.Nil(t, seq)
	require.Error(t, err)
	assertMidiError(t, err, midi.ErrUnset)
	assert.ErrorIs(t, err, octave.ErrInvalid)

	requirePanicError(t, err, func() { midi.MustNumbersInOctave(o) })
}

// The sequence is repeatable and supports early stop.
func TestNumbersInOctaveSequence(t *testing.T) {
	t.Parallel()

	seq, err := midi.NumbersInOctave(octave.FromASPN(4))
	require.NoError(t, err)

	want := []int{60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71}
	assert.Equal(t, want, collectNumbers(t, seq))
	assert.Equal(t, want, collectNumbers(t, seq), "the sequence must be repeatable")

	prefix := make([]int, 0, 3)
	for number := range seq {
		prefix = append(prefix, number.MustInt())
		if len(prefix) == 3 {
			break
		}
	}

	assert.Equal(t, []int{60, 61, 62}, prefix, "the sequence must support early stop")
	assert.Equal(t, want, collectNumbers(t, seq), "early stop must not consume the sequence")
}

// One sequence may be traversed by multiple goroutines at once.
func TestNumbersInOctaveParallel(t *testing.T) {
	t.Parallel()

	seq, err := midi.NumbersInOctave(octave.FromASPN(9))
	require.NoError(t, err)

	const workers = 8
	results := make([][]int, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			values := make([]int, 0, 8)
			for number := range seq {
				values = append(values, number.MustInt())
			}
			results[i] = values
		}()
	}

	wg.Wait()

	want := []int{120, 121, 122, 123, 124, 125, 126, 127}
	for i, values := range results {
		assert.Equal(t, want, values, "worker %d", i)
	}
}

// The octaves -1 through 9 together give every MIDI number exactly once.
func TestNumbersInOctaveCoversRange(t *testing.T) {
	t.Parallel()

	all := make([]int, 0, midi.Count())
	for aspn := -1; aspn <= 9; aspn++ {
		seq, err := midi.NumbersInOctave(octave.FromASPN(aspn))
		require.NoError(t, err)
		all = append(all, collectNumbers(t, seq)...)
	}

	expected := make([]int, midi.Count())
	for i := range expected {
		expected[i] = i
	}

	assert.Equal(t, expected, all)
}

// MustNumbersInOctave gives the same sequence for a valid octave.
func TestMustNumbersInOctave(t *testing.T) {
	t.Parallel()

	o := octave.FromASPN(-1)
	seq, err := midi.NumbersInOctave(o)
	require.NoError(t, err)

	assert.Equal(t, collectNumbers(t, seq), collectNumbers(t, midi.MustNumbersInOctave(o)))
}
