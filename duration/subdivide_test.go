package duration_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubdivide covers O-7, P-11, KP-17, T-5, and architecture 4.6: a base
// divides into 2 and a base with one dot into 3 parts one step shorter,
// without dots, with the same chain; count·part equals the value.
func TestSubdivide(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		w     duration.Written
		count int
		part  duration.Written
	}{
		{"KP-17 quarter", wr(duration.Quarter, 0), 2, wr(duration.Eighth, 0)},
		{"KP-17 dotted quarter", wr(duration.Quarter, 1), 3, wr(duration.Eighth, 0)},
		{"KP-17 dotted half", wr(duration.Half, 1), 3, wr(duration.Quarter, 0)},
		{"KP-17 quarter in 3:2", wr(duration.Quarter, 0, tuplet.Triplet()), 2, wr(duration.Eighth, 0, tuplet.Triplet())},
		{"maxima", wr(duration.Maxima, 0), 2, wr(duration.Longa, 0)},
		{"dotted maxima", wr(duration.Maxima, 1), 3, wr(duration.Longa, 0)},
		{"512th", wr(duration.FiveHundredTwelfth, 0), 2, wr(duration.ThousandTwentyFourth, 0)},
		{"dotted 512th under a chain", wr(duration.FiveHundredTwelfth, 1, tp(6, 4), tuplet.Quintuplet()), 3,
			wr(duration.ThousandTwentyFourth, 0, tp(6, 4), tuplet.Quintuplet())},
	} {
		count, part, err := c.w.Subdivide()
		require.NoError(t, err, c.name)
		assert.Equal(t, c.count, count, c.name)
		assert.Equal(t, c.part, part, c.name)
		assert.Equal(t, c.w.MustDuration(), part.MustDuration().MustMulInt(count), "%s: count·part", c.name)
	}

	// Every base of the model with at most one dot.
	for i, row := range baseTable[:len(baseTable)-1] {
		for dots := range 2 {
			for _, chain := range testChains() {
				count, part := wr(row.base, dots, chain...).MustSubdivide()
				assert.Equal(t, 2+dots, count)
				assert.Equal(t, wr(baseTable[i+1].base, 0, chain...), part)
			}
		}
	}
}

// TestSubdivideErrors covers KP-33 and architecture 4.6: two or more dots
// are E-7 (also for the 1024th, by the order of checks), parts beyond the
// 1024th are E-6, and a part out of range is E-11.
func TestSubdivideErrors(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name   string
		w      duration.Written
		want   error
		causes []error
	}{
		{"KP-33 two dots", wr(duration.Quarter, 2), duration.ErrNoSubdivision, nil},
		{"three dots", wr(duration.Breve, 3, tuplet.Triplet()), duration.ErrNoSubdivision, nil},
		{"KP-33 1024th", wr(duration.ThousandTwentyFourth, 0), duration.ErrOutsideSeries, nil},
		{"KP-33 dotted 1024th", wr(duration.ThousandTwentyFourth, 1), duration.ErrOutsideSeries, nil},
		{"1024th with two dots", wr(duration.ThousandTwentyFourth, 2), duration.ErrNoSubdivision, nil},
		{"part out of range", edgeWritten(), duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	} {
		count, part, err := c.w.Subdivide()
		assertErr(t, err, c.want, c.causes...)
		assert.Equal(t, 0, count, c.name)
		assert.Equal(t, duration.Written{}, part, c.name)
	}
}
