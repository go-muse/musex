package duration_test

import (
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FreeContents lets a group hold any number of values whose written sum is
// the actual count times the unit; it is off by default.
func TestReadingsFreeContents(t *testing.T) {
	t.Parallel()

	triplet := tuplet.Triplet()
	eighth := duration.MustNewWritten(duration.Eighth, 0)

	t.Run("quarter and eighth", func(t *testing.T) {
		t.Parallel()

		ds := []duration.Duration{duration.MustNew(1, 6), duration.MustNew(1, 12)}

		off, err := duration.Readings(ds, duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}})
		require.NoError(t, err)
		assert.Empty(t, off)

		on, err := duration.Readings(ds, duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}, FreeContents: true})
		require.NoError(t, err)
		require.Len(t, on, 1)
		items := on[0].MustContents()
		require.Len(t, items, 1)
		g := items[0].MustGroup()
		assert.True(t, g.MustTuplet().Equal(triplet))
		assert.True(t, g.MustUnit().Equal(eighth))
		contents := g.MustContents()
		require.Len(t, contents, 2)
		assert.True(t, contents[0].MustWritten().Equal(duration.MustNewWritten(duration.Quarter, 0, triplet)))
		assert.True(t, contents[1].MustWritten().Equal(duration.MustNewWritten(duration.Eighth, 0, triplet)))
		spans := g.MustSpans()
		assert.Equal(t, fraction.MustNewReduced(1, 6), spans[0][1])
		assert.Equal(t, fraction.MustNewReduced(1, 4), spans[1][1])
	})

	t.Run("two eighths and two sixteenths", func(t *testing.T) {
		t.Parallel()

		ds := []duration.Duration{
			duration.MustNew(1, 12), duration.MustNew(1, 12),
			duration.MustNew(1, 24), duration.MustNew(1, 24),
		}
		on, err := duration.Readings(ds, duration.ReadingsOptions{Allowed: []tuplet.Tuplet{triplet}, FreeContents: true})
		require.NoError(t, err)
		require.Len(t, on, 1)
		g := on[0].MustContents()[0].MustGroup()
		assert.True(t, g.MustUnit().Equal(eighth))
		assert.Len(t, g.MustContents(), 4)
	})
}

// Nesting works for tuplets whose coefficient is not 1: the unit of a nested
// group is written under the enclosing chain.
func TestReadingsNestingCoefficient(t *testing.T) {
	t.Parallel()

	ds := []duration.Duration{
		duration.MustNew(1, 10), duration.MustNew(1, 10),
		duration.MustNew(1, 30), duration.MustNew(1, 30), duration.MustNew(1, 30),
		duration.MustNew(1, 10), duration.MustNew(1, 10),
	}
	readings, err := duration.Readings(ds, duration.ReadingsOptions{
		Allowed:  []tuplet.Tuplet{tuplet.Quintuplet(), tuplet.Triplet()},
		MaxDepth: 1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, readings)

	var found bool
	for _, r := range readings {
		items := r.MustContents()
		if len(items) != 1 || !items[0].IsGroup() {
			continue
		}
		outer := items[0].MustGroup()
		if !outer.MustTuplet().Equal(tuplet.Quintuplet()) {
			continue
		}
		contents := outer.MustContents()
		if len(contents) != 5 || !contents[2].IsGroup() {
			continue
		}
		inner := contents[2].MustGroup()
		if !inner.MustTuplet().Equal(tuplet.Triplet()) {
			continue
		}
		// The inner unit is a sixteenth written under 5:4, not a plain sign.
		assert.True(t, inner.MustUnit().Equal(duration.MustNewWritten(duration.Sixteenth, 0, tuplet.Quintuplet())))
		assert.Len(t, inner.MustContents(), 3)
		found = true
	}
	assert.True(t, found, "a quintuplet of eighths with a triplet of sixteenths inside")
}

// The work of a call is bounded independently of the result count. On a long
// input that has readings, the call returns within the time limit either with
// at most MaxReadings distinct exact readings or with ErrSearchLimit, never
// with ErrTooManyParts and never with an empty result and no error. The
// contract bounds the work but does not prescribe which outcome this input
// has.
func TestReadingsWorkBound(t *testing.T) {
	t.Parallel()

	ds := make([]duration.Duration, 240)
	for i := range ds {
		ds[i] = duration.MustNew(1, 12)
	}

	opts := duration.ReadingsOptions{
		Allowed:     []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet()},
		MaxReadings: 4,
	}

	r := within(t, 10*time.Second, func() result[[]duration.Reading] {
		readings, err := duration.Readings(ds, opts)

		return result[[]duration.Reading]{readings, err}
	})
	assertBoundedReadings(t, ds, opts, r)
}

// MaxDots excludes readings that need more dots.
func TestReadingsMaxDots(t *testing.T) {
	t.Parallel()

	ds := []duration.Duration{duration.MustNew(3, 8)}
	none, err := duration.Readings(ds, duration.ReadingsOptions{MaxDots: 0})
	require.NoError(t, err)
	assert.Empty(t, none)

	one, err := duration.Readings(ds, duration.ReadingsOptions{MaxDots: 1})
	require.NoError(t, err)
	require.Len(t, one, 1)
	assert.True(t, one[0].MustContents()[0].MustWritten().Equal(duration.MustNewWritten(duration.Quarter, 1)))
}
