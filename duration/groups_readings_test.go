package duration_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SubdivideTuplet divides a written duration into the actual count of the
// tuplet, preserving the enclosing chain.
func TestSubdivideTupletAcceptance(t *testing.T) {
	t.Parallel()

	quarter := duration.MustNewWritten(duration.Quarter, 0)
	dottedQuarter := duration.MustNewWritten(duration.Quarter, 1)

	t.Run("quarter 3:2", func(t *testing.T) {
		t.Parallel()

		count, part, err := quarter.SubdivideTuplet(tuplet.Triplet())
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, duration.MustNewWritten(duration.Eighth, 0, tuplet.Triplet()), part)
		assert.Equal(t, fraction.MustNewReduced(1, 12), part.MustDuration().MustFraction())
	})

	t.Run("dotted quarter 2:3", func(t *testing.T) {
		t.Parallel()

		count, part, err := dottedQuarter.SubdivideTuplet(tuplet.Duplet())
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, duration.MustNewWritten(duration.Eighth, 0, tuplet.Duplet()), part)
		assert.Equal(t, fraction.MustNewReduced(3, 16), part.MustDuration().MustFraction())
	})

	t.Run("dotted quarter 3:2 keeps dots", func(t *testing.T) {
		t.Parallel()

		count, part, err := dottedQuarter.SubdivideTuplet(tuplet.Triplet())
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, duration.MustNewWritten(duration.Eighth, 1, tuplet.Triplet()), part)
	})

	t.Run("enclosing chain preserved", func(t *testing.T) {
		t.Parallel()

		outer := duration.MustNewWritten(duration.Quarter, 0, tuplet.Quintuplet())
		count, part, err := outer.SubdivideTuplet(tuplet.Triplet())
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, []tuplet.Tuplet{tuplet.Quintuplet(), tuplet.Triplet()}, part.MustChain())
	})
}

// A written tuplet group exposes its contents' sounding spans and shortfall
func TestGroupAcceptance(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	eighth := duration.MustNewWritten(duration.Eighth, 0)
	quarterInTriplet := duration.MustNewWritten(duration.Quarter, 0, triplet)
	eighthInTriplet := duration.MustNewWritten(duration.Eighth, 0, triplet)

	g, err := duration.NewGroup(triplet, eighth,
		duration.WrittenContent(quarterInTriplet),
		duration.WrittenContent(eighthInTriplet),
	)
	require.NoError(t, err)

	assert.Equal(t, fraction.MustNewReduced(1, 4), g.MustSpan())
	assert.Equal(t, fraction.MustNewReduced(0, 1), g.MustShortfall())

	spans := g.MustSpans()
	require.Len(t, spans, 2)
	assert.Equal(t, fraction.MustNewReduced(0, 1), spans[0][0])
	assert.Equal(t, fraction.MustNewReduced(1, 6), spans[0][1])
	assert.Equal(t, fraction.MustNewReduced(1, 6), spans[1][0])
	assert.Equal(t, fraction.MustNewReduced(1, 4), spans[1][1])

	// One 6:4 group stays distinguishable from two 3:2 groups.
	sex := duration.MustNewWritten(duration.Eighth, 0, tuplet.Sextuplet())
	one64 := duration.MustNewGroup(tuplet.Sextuplet(), eighth,
		duration.WrittenContent(sex), duration.WrittenContent(sex), duration.WrittenContent(sex),
		duration.WrittenContent(sex), duration.WrittenContent(sex), duration.WrittenContent(sex),
	)
	three32 := duration.MustNewWritten(duration.Eighth, 0, triplet)
	two32 := duration.MustNewGroup(triplet, eighth,
		duration.WrittenContent(three32), duration.WrittenContent(three32), duration.WrittenContent(three32),
	)
	assert.False(t, one64.Equal(two32))
}

// Readings finds both exact readings of six sounding 1/12 over a half note
// with 3:2 and 6:4 allowed.
func TestReadingsAcceptance(t *testing.T) {
	t.Parallel()

	ds := make([]duration.Duration, 6)
	for i := range ds {
		ds[i] = duration.MustNew(1, 12)
	}

	readings, err := duration.Readings(ds, duration.ReadingsOptions{
		Allowed: []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()},
	})
	require.NoError(t, err)
	require.Len(t, readings, 2)

	// Fewer groups first: the single 6:4 group, then two 3:2 groups.
	first := readings[0].MustContents()
	require.Len(t, first, 1)
	require.True(t, first[0].IsGroup())
	g := first[0].MustGroup()
	assert.Equal(t, tuplet.Sextuplet(), g.MustTuplet())
	assert.Len(t, g.MustContents(), 6)

	second := readings[1].MustContents()
	require.Len(t, second, 2)
	for _, item := range second {
		require.True(t, item.IsGroup())
		g := item.MustGroup()
		assert.Equal(t, tuplet.Triplet(), g.MustTuplet())
		assert.Len(t, g.MustContents(), 3)
	}

	// No reading is an empty result, not an error.
	none, err := duration.Readings([]duration.Duration{duration.MustNew(1, 12)}, duration.ReadingsOptions{})
	require.NoError(t, err)
	assert.Empty(t, none)
}
