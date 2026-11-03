package octave_test

import (
	"testing"

	"github.com/go-muse/musex/octave"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rangeLimit bounds the number of elements collected from a sequence, so a
// wrong sequence over the whole axis cannot run without end.
const rangeLimit = 100

// K-15, C-15, G-7: the range from one octave to another, inclusive, upward,
// downward or of one octave; MustRange gives the same sequence.
func TestRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		from, to octave.Octave
		want     []octave.Octave
	}{
		{name: "O(2)...O(5)", from: oct(2), to: oct(5), want: octs(2, 3, 4, 5)},
		{name: "O(5)...O(2)", from: oct(5), to: oct(2), want: octs(5, 4, 3, 2)},
		{name: "O(4)...O(4)", from: oct(4), to: oct(4), want: octs(4)},
		{name: "O(-1)...O(1)", from: oct(-1), to: oct(1), want: octs(-1, 0, 1)},
		{name: "O(MAX-2)...O(MAX)", from: oct(maxInt - 2), to: oct(maxInt), want: octs(maxInt-2, maxInt-1, maxInt)},
		{name: "O(MIN+2)...O(MIN)", from: oct(minInt + 2), to: oct(minInt), want: octs(minInt+2, minInt+1, minInt)},
		{name: "O(MAX)...O(MAX-2)", from: oct(maxInt), to: oct(maxInt - 2), want: octs(maxInt, maxInt-1, maxInt-2)},
		{name: "O(MIN)...O(MIN+2)", from: oct(minInt), to: oct(minInt + 2), want: octs(minInt, minInt+1, minInt+2)},
		{name: "O(MAX)...O(MAX)", from: oct(maxInt), to: oct(maxInt), want: octs(maxInt)},
		{name: "O(MIN)...O(MIN)", from: oct(minInt), to: oct(minInt), want: octs(minInt)},
		{name: "FB 1...FB -1", from: octave.MustFromNumber(1, fb()), to: octave.MustFromNumber(-1, fb()), want: octs(2, 1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			seq, err := octave.Range(tc.from, tc.to)
			require.NoError(t, err)
			require.NotNil(t, seq)

			var got []octave.Octave
			require.NotPanics(t, func() { got = collect(seq, rangeLimit) })
			assert.Equal(t, tc.want, got, "Range: %s", describe(got))

			var must []octave.Octave
			require.NotPanics(t, func() { must = collect(octave.MustRange(tc.from, tc.to), rangeLimit) })
			assert.Equal(t, tc.want, must, "MustRange: %s", describe(must))
		})
	}
}

// K-15: for all a, b in -5...5, the range has Distance + 1 elements, starts
// at from, ends at to, and neighboring elements differ by one octave in the
// direction from from to to.
func TestRangeProperties(t *testing.T) {
	t.Parallel()

	for _, a := range span(-5, 5) {
		for _, b := range span(-5, 5) {
			seq, err := octave.Range(oct(a), oct(b))
			require.NoError(t, err, "Range(O(%d), O(%d))", a, b)
			got := collect(seq, rangeLimit)

			require.Len(t, got, oct(a).MustDistance(oct(b))+1, "Range(O(%d), O(%d)): %s", a, b, describe(got))
			assert.Equal(t, oct(a), got[0], "first of Range(O(%d), O(%d))", a, b)
			assert.Equal(t, oct(b), got[len(got)-1], "last of Range(O(%d), O(%d))", a, b)

			step := sign(b - a)
			for i := 1; i < len(got); i++ {
				d, err := got[i].Sub(got[i-1])
				require.NoError(t, err, "Range(O(%d), O(%d)) elements %d, %d", a, b, i-1, i)
				assert.Equal(t, step, d, "Range(O(%d), O(%d)) elements %d, %d", a, b, i-1, i)
			}
		}
	}
}

// K-15, G-7: a range across the whole axis is lazy and can be stopped: after
// three elements the consumer breaks; the iterator stops calling yield, which
// the Go runtime would otherwise report with a panic.
func TestRangeBreak(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		from, to octave.Octave
		want     []octave.Octave
	}{
		{name: "O(MIN)...O(MAX)", from: oct(minInt), to: oct(maxInt), want: octs(minInt, minInt+1, minInt+2)},
		{name: "O(MAX)...O(MIN)", from: oct(maxInt), to: oct(minInt), want: octs(maxInt, maxInt-1, maxInt-2)},
		{name: "O(0)...O(10)", from: oct(0), to: oct(10), want: octs(0, 1, 2)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			seq, err := octave.Range(tc.from, tc.to)
			require.NoError(t, err)

			var got []octave.Octave
			require.NotPanics(t, func() {
				for o := range seq {
					got = append(got, o)
					if len(got) == 3 {
						break
					}
				}
			})
			assert.Equal(t, tc.want, got, "%s", describe(got))
		})
	}
}

// K-15: each iteration of the same sequence starts again from from and gives
// the same elements, also after an earlier iteration was stopped.
func TestRangeRepeatable(t *testing.T) {
	t.Parallel()

	seq, err := octave.Range(oct(2), oct(6))
	require.NoError(t, err)

	first := collect(seq, rangeLimit)
	stopped := collect(seq, 2)
	second := collect(seq, rangeLimit)

	assert.Equal(t, octs(2, 3, 4, 5, 6), first, "first: %s", describe(first))
	assert.Equal(t, octs(2, 3), stopped, "stopped: %s", describe(stopped))
	assert.Equal(t, first, second, "second: %s", describe(second))

	whole := octave.MustRange(oct(minInt), oct(maxInt))
	assert.Equal(t, octs(minInt, minInt+1), collect(whole, 2), "whole axis, first iteration")
	assert.Equal(t, octs(minInt, minInt+1, minInt+2), collect(whole, 3), "whole axis, second iteration")
}
