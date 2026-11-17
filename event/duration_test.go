package event_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrittenDurationPreservesNotation(t *testing.T) {
	t.Parallel()

	triplet := tuplet.MustNew(3, 2)
	cases := []struct {
		name  string
		base  duration.Base
		dots  int
		chain []tuplet.Tuplet
		num   int
		den   int
	}{
		{"dotted quarter", duration.Quarter, 1, nil, 3, 8},
		{"triplet half", duration.Half, 0, []tuplet.Tuplet{triplet}, 1, 3},
		{"ordered nested groups", duration.Quarter, 0, []tuplet.Tuplet{triplet, tuplet.MustNew(5, 4)}, 2, 15},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := duration.MustNewWritten(tt.base, tt.dots, tt.chain...)
			d, err := event.WrittenDuration(w)
			require.NoError(t, err)
			assert.True(t, d.IsSet())
			assert.Equal(t, event.DurationWritten, d.Kind())
			assert.True(t, event.MustWrittenDuration(w).Equal(d))

			got, err := d.Written()
			require.NoError(t, err)
			assert.True(t, w.Equal(got))
			assert.Equal(t, tt.base, got.MustBase())
			assert.Equal(t, tt.dots, got.MustDots())
			assert.Len(t, got.MustChain(), len(tt.chain))
			for i, group := range got.MustChain() {
				require.Less(t, i, len(tt.chain))
				assert.True(t, tt.chain[i].Equal(group), "group %d", i)
			}
			assert.True(t, w.Equal(d.MustWritten()))

			expected := fraction.MustNew(tt.num, tt.den)
			v, err := d.Value()
			require.NoError(t, err)
			assertFraction(t, expected, v.MustFraction())
			assertFraction(t, expected, d.MustValue().MustFraction())

			assertFailure(t, event.ErrWrongDurationKind, d.MeasureCount, d.MustMeasureCount)
		})
	}
}

func TestValueDurationHasOnlyExactValue(t *testing.T) {
	t.Parallel()

	input := fraction.MustNew(2, 6)
	d, err := event.ValueDuration(input)
	require.NoError(t, err)
	assert.True(t, d.IsSet())
	assert.Equal(t, event.DurationValue, d.Kind())

	v, err := d.Value()
	require.NoError(t, err)
	want := fraction.MustNew(1, 3)
	assertFraction(t, want, v.MustFraction())
	assert.True(t, want.EqualExact(v.MustFraction()), "value should be in reduced form")
	assertFraction(t, want, d.MustValue().MustFraction())
	assert.True(t, event.MustValueDuration(input).Equal(d))
	assert.True(t, fraction.MustNew(2, 6).EqualExact(input))

	assertFailure(t, event.ErrWrongDurationKind, d.Written, d.MustWritten)
	assertFailure(t, event.ErrWrongDurationKind, d.MeasureCount, d.MustMeasureCount)
}

func TestMeasureDurationAcceptsPositiveCounts(t *testing.T) {
	t.Parallel()

	for _, count := range []int{1, 3, math.MaxInt} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			t.Parallel()

			d, err := event.MeasureDuration(count)
			require.NoError(t, err)
			assert.True(t, d.IsSet())
			assert.Equal(t, event.DurationMeasures, d.Kind())
			got, err := d.MeasureCount()
			require.NoError(t, err)
			assert.Equal(t, count, got)
			assert.Equal(t, count, d.MustMeasureCount())
			assert.True(t, event.MustMeasureDuration(count).Equal(d))

			assertFailure(t, event.ErrWrongDurationKind, d.Written, d.MustWritten)
			assertFailure(t, event.ErrNeedsMeasureContext, d.Value, d.MustValue)
		})
	}
}

func TestDurationConstructorFailures(t *testing.T) {
	t.Parallel()

	t.Run("unset written duration", func(t *testing.T) {
		var w duration.Written
		assertFailure(t, duration.ErrUnset,
			func() (event.Duration, error) { return event.WrittenDuration(w) },
			func() event.Duration { return event.MustWrittenDuration(w) })
	})

	values := []struct {
		name  string
		value fraction.Fraction
		err   error
	}{
		{"unset", fraction.Fraction{}, duration.ErrInvalidArgument},
		{"zero", fraction.Zero(), duration.ErrNotPositive},
		{"negative", fraction.MustNew(-1, 8), duration.ErrNotPositive},
	}
	for _, tt := range values {
		t.Run(tt.name+" fraction", func(t *testing.T) {
			assertFailure(t, tt.err,
				func() (event.Duration, error) { return event.ValueDuration(tt.value) },
				func() event.Duration { return event.MustValueDuration(tt.value) })
		})
	}
	for _, count := range []int{0, -1, math.MinInt} {
		t.Run("count "+strconv.Itoa(count), func(t *testing.T) {
			assertFailure(t, event.ErrMeasureCount,
				func() (event.Duration, error) { return event.MeasureDuration(count) },
				func() event.Duration { return event.MustMeasureDuration(count) })
		})
	}
}

func TestDurationEqualityAndComparability(t *testing.T) {
	t.Parallel()

	triplet := tuplet.MustNew(3, 2)
	quintuplet := tuplet.MustNew(5, 4)
	written := func(groups ...tuplet.Tuplet) event.Duration {
		return event.MustWrittenDuration(duration.MustNewWritten(duration.Half, 0, groups...))
	}
	nested := written(triplet, quintuplet)
	set := []struct {
		name string
		a, b event.Duration
		same bool
	}{
		{"same written record", nested, written(triplet, quintuplet), true},
		{"different group order", nested, written(quintuplet, triplet), false},
		{"equivalent group ratio", written(triplet), written(tuplet.MustNew(6, 4)), false},
		{"same length across forms", written(triplet), valueDuration(1, 3), false},
		{"equivalent fractions", valueDuration(1, 3), valueDuration(2, 6), true},
		{"different exact values", valueDuration(1, 3), valueDuration(1, 4), false},
		{"same measure count", event.MustMeasureDuration(3), event.MustMeasureDuration(3), true},
		{"different measure counts", event.MustMeasureDuration(3), event.MustMeasureDuration(1), false},
	}
	for _, tt := range set {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.same, tt.a.Equal(tt.b))
			assert.Equal(t, tt.same, tt.b.Equal(tt.a))
			assert.Equal(t, tt.same, tt.a == tt.b)
			_, found := map[event.Duration]struct{}{tt.a: {}}[tt.b]
			assert.Equal(t, tt.same, found)
		})
	}
}

func TestUnsetDuration(t *testing.T) {
	t.Parallel()

	var d event.Duration
	assert.False(t, d.IsSet())
	assert.Equal(t, event.DurationUnset, d.Kind())
	assert.True(t, d == (event.Duration{}))
	assert.False(t, d.Equal(d))
	assert.False(t, d.Equal(valueDuration(1, 2)))
	assert.False(t, valueDuration(1, 2).Equal(d))
	assertFailure(t, event.ErrUnsetDuration, d.Written, d.MustWritten)
	assertFailure(t, event.ErrUnsetDuration, d.Value, d.MustValue)
	assertFailure(t, event.ErrUnsetDuration, d.MeasureCount, d.MustMeasureCount)
}
