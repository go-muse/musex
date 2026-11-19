package tendency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBetweenPlaces(t *testing.T) {
	t.Parallel()
	fractional := interval.MustFromDeviation(2, alteration.MustFromFraction(-1, 2))
	cases := []struct {
		name      string
		m         mode.Mode
		from, to  tendency.Place
		value     interval.Interval
		direction tendency.Motion
	}{
		{"stationary", ionian(), place(3, 0), place(3, 0), iv("P1"), tendency.Stationary},
		{"negative circle", ionian(), place(1, -1), place(1, 0), iv("P8"), tendency.Ascending},
		{"descending", ionian(), place(1, 0), place(7, -1), iv("m2").MustDown(), tendency.Descending},
		{"non octave", closed("P1 M3 P5", "P12"), place(3, -1), place(1, 0), iv("P8"), tendency.Ascending},
		{"single degree", closed("P1", "P5"), place(1, 0), place(1, 1), iv("P5"), tendency.Ascending},
		{"single open", open("P1"), place(1, 0), place(1, 0), iv("P1"), tendency.Stationary},
		{"fractional", mode.MustOpenFromDegreeIntervals([]interval.Interval{iv("P1"), fractional}), place(1, 0), place(2, 0), fractional, tendency.Ascending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tendency.Between(tc.m, tc.from, tc.to)
			require.NoError(t, err)
			assert.Equal(t, tendency.Movement{From: tc.from, To: tc.to, Interval: tc.value, Direction: tc.direction}, got)
			assert.Equal(t, got, tendency.MustBetween(tc.m, tc.from, tc.to))
			require.NoError(t, tendency.ValidatePlace(tc.m, tc.from))
			assert.NotPanics(t, func() { tendency.MustValidatePlace(tc.m, tc.to) })
		})
	}
	// Each known coordinate is validated before their difference is taken.
	for _, p := range []tendency.Place{place(1, math.MaxInt), place(1, math.MinInt)} {
		assert.ErrorIs(t, tendency.ValidatePlace(ionian(), p), tendency.ErrOutOfRange)
	}
	for _, p := range []tendency.Place{place(0, 0), place(8, 0), place(-1, 0)} {
		assert.ErrorIs(t, tendency.ValidatePlace(ionian(), p), tendency.ErrInvalidPlace)
	}
	assert.ErrorIs(t, tendency.ValidatePlace(open("P1 M2"), place(1, 1)), tendency.ErrInvalidPlace)
	assert.ErrorIs(t, tendency.ValidatePlace(mode.Mode{}, place(0, 0)), tendency.ErrInvalidMode)
	// The individual coordinates fit, but the final distance does not.
	c := (math.MaxInt - 1) / 7
	r, err := tendency.Between(ionian(), place(1, -c), place(1, c))
	assert.ErrorIs(t, err, tendency.ErrOutOfRange)
	assert.Equal(t, tendency.Movement{}, r)
	// A genuinely unrepresentable reduced deviation is rejected.
	wide := interval.MustFromDeviation(2, alteration.MustFromInt(alteration.MaxMagnitude))
	m := mode.MustClosedFromDegreeIntervals([]interval.Interval{iv("P1")}, wide)
	r, err = tendency.Between(m, place(1, 0), place(1, 2))
	// The final deviation exceeds the alteration range.
	assert.ErrorIs(t, err, tendency.ErrOutOfRange)
	assert.Equal(t, tendency.Movement{}, r)
	// The chromatic coordinate 12*c exceeds int, but the final P(7*c+1) fits.
	cycle := iv("P8")
	m = mode.MustClosedFromDegreeIntervals([]interval.Interval{iv("P1")}, cycle)
	r, err = tendency.Between(m, place(1, 0), place(1, c))
	require.NoError(t, err)
	assert.Equal(t, 7*c+1, r.Interval.Size())
	assert.Equal(t, alteration.MustFromInt(0), r.Interval.Deviation())
	// Reduction after a wide rational product must happen before range checks.
	fractionalCycle := interval.MustFromDeviation(2, alteration.MustFromFraction(-2, 7))
	m = mode.MustClosedFromDegreeIntervals([]interval.Interval{iv("P1")}, fractionalCycle)
	const largeCycle = 1 << 62
	r, err = tendency.Between(m, place(1, 0), place(1, largeCycle))
	require.NoError(t, err)
	assert.Equal(t, largeCycle+1, r.Interval.Size())
	assert.Equal(t, alteration.MustFromFraction(-1, 7), r.Interval.Deviation())
	for from := -2; from <= 2; from++ {
		for to := -2; to <= 2; to++ {
			a, b := place(2, from), place(6, to)
			forward := tendency.MustBetween(ionian(), a, b)
			reverse := tendency.MustBetween(ionian(), b, a)
			assert.True(t, forward.Interval.MustOpposite().Equal(reverse.Interval))
			middle := place(4, 0)
			composed := tendency.MustBetween(ionian(), a, middle).Interval.MustAdd(tendency.MustBetween(ionian(), middle, b).Interval)
			assert.Equal(t, forward.Interval, composed)
		}
	}
}
