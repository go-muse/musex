package duration_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubdivideTupletErrors(t *testing.T) {
	t.Parallel()

	_, _, err := duration.Written{}.SubdivideTuplet(tuplet.Triplet())
	assert.ErrorIs(t, err, duration.ErrUnset)

	quarter := duration.MustNewWritten(duration.Quarter, 0)
	_, _, err = quarter.SubdivideTuplet(tuplet.Tuplet{})
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)
	assert.ErrorIs(t, err, tuplet.ErrUnset)

	// 1/4 divided by 3 is not dyadic: no single written sign.
	_, _, err = quarter.SubdivideTuplet(tuplet.Duplet())
	assert.ErrorIs(t, err, duration.ErrNotNotatable)

	// The part would be shorter than the 1024th.
	tiny := duration.MustNewWritten(duration.ThousandTwentyFourth, 0)
	_, _, err = tiny.SubdivideTuplet(tuplet.MustNew(1, 2))
	assert.ErrorIs(t, err, duration.ErrOutsideSeries)

	assert.Panics(t, func() { quarter.MustSubdivideTuplet(tuplet.Duplet()) })
}

func TestGroupEdges(t *testing.T) {
	t.Parallel()

	eighth := duration.MustNewWritten(duration.Eighth, 0)
	triplet := tuplet.Triplet()
	inTriplet := duration.MustNewWritten(duration.Eighth, 0, triplet)

	_, err := duration.NewGroup(tuplet.Tuplet{}, eighth)
	assert.ErrorIs(t, err, tuplet.ErrUnset)

	_, err = duration.NewGroup(triplet, duration.Written{})
	assert.ErrorIs(t, err, duration.ErrUnset)

	_, err = duration.NewGroup(triplet, eighth, duration.Content{})
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)

	// Contents exceeding the span are rejected.
	quarterInTriplet := duration.MustNewWritten(duration.Quarter, 0, triplet)
	_, err = duration.NewGroup(triplet, eighth,
		duration.WrittenContent(quarterInTriplet),
		duration.WrittenContent(quarterInTriplet),
	)
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)

	// A shortfall is exposed, not an error.
	g := duration.MustNewGroup(triplet, eighth, duration.WrittenContent(inTriplet))
	assert.Equal(t, fraction.MustNewReduced(1, 6), g.MustShortfall())

	// Nested groups stay distinguishable and equal recursively.
	inner := duration.MustNewGroup(triplet, eighth,
		duration.WrittenContent(inTriplet), duration.WrittenContent(inTriplet), duration.WrittenContent(inTriplet))
	outer := duration.MustNewGroup(tuplet.Sextuplet(), eighth, duration.GroupContent(inner), duration.GroupContent(inner))
	outer2 := duration.MustNewGroup(tuplet.Sextuplet(), eighth, duration.GroupContent(inner), duration.GroupContent(inner))
	assert.True(t, outer.Equal(outer2))
	assert.False(t, outer.Equal(inner))

	// Contents are copied.
	contents := outer.MustContents()
	contents[0] = duration.Content{}
	assert.True(t, outer.Equal(outer2))

	var zero duration.Group
	assert.False(t, zero.IsSet())
	_, err = zero.Span()
	assert.ErrorIs(t, err, duration.ErrUnset)
	assert.True(t, zero.Equal(duration.Group{}))
}

func TestReadingsEdges(t *testing.T) {
	t.Parallel()

	_, err := duration.Readings(nil, duration.ReadingsOptions{})
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)

	_, err = duration.Readings([]duration.Duration{{}}, duration.ReadingsOptions{})
	assert.ErrorIs(t, err, duration.ErrUnset)

	_, err = duration.Readings([]duration.Duration{duration.MustNew(1, 4)}, duration.ReadingsOptions{MaxDots: -1})
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)

	_, err = duration.Readings([]duration.Duration{duration.MustNew(1, 4)}, duration.ReadingsOptions{Allowed: []tuplet.Tuplet{{}}})
	assert.ErrorIs(t, err, tuplet.ErrUnset)

	// A plain quarter needs no tuplet and stays outside groups.
	readings, err := duration.Readings([]duration.Duration{duration.MustNew(1, 4)}, duration.ReadingsOptions{})
	require.NoError(t, err)
	require.Len(t, readings, 1)
	items := readings[0].MustContents()
	require.Len(t, items, 1)
	assert.False(t, items[0].IsGroup())
	assert.Equal(t, duration.MustNewWritten(duration.Quarter, 0), items[0].MustWritten())

	// Nesting is bounded by MaxDepth: with depth 1 a neutral 2:2 group may
	// hold two 3:2 groups over six sounding 1/12, a reading unavailable at
	// depth 0.
	ds := make([]duration.Duration, 6)
	for i := range ds {
		ds[i] = duration.MustNew(1, 12)
	}
	neutral := tuplet.MustNew(2, 2)
	flat, err := duration.Readings(ds, duration.ReadingsOptions{
		Allowed: []tuplet.Tuplet{tuplet.Triplet(), neutral},
	})
	require.NoError(t, err)

	nested, err := duration.Readings(ds, duration.ReadingsOptions{
		Allowed:  []tuplet.Tuplet{tuplet.Triplet(), neutral},
		MaxDepth: 1,
	})
	require.NoError(t, err)
	assert.Greater(t, len(nested), len(flat), "nesting adds readings")

	var hasNested bool
	for _, r := range nested {
		for _, item := range r.MustContents() {
			if item.IsGroup() {
				for _, c := range item.MustGroup().MustContents() {
					if c.IsGroup() {
						hasNested = true
					}
				}
			}
		}
	}
	assert.True(t, hasNested, "a reading has a group inside a group")

	// MaxReadings truncates deterministically.
	one, err := duration.Readings(ds, duration.ReadingsOptions{
		Allowed:     []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()},
		MaxReadings: 1,
	})
	require.NoError(t, err)
	assert.Len(t, one, 1)
}
