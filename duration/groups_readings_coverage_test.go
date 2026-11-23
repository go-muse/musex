package duration_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupZeroAccessors(t *testing.T) {
	t.Parallel()

	var g duration.Group
	_, err := g.Tuplet()
	assert.ErrorIs(t, err, duration.ErrUnset)
	_, err = g.Unit()
	assert.ErrorIs(t, err, duration.ErrUnset)
	_, err = g.Contents()
	assert.ErrorIs(t, err, duration.ErrUnset)
	_, err = g.Shortfall()
	assert.ErrorIs(t, err, duration.ErrUnset)
	_, err = g.Spans()
	assert.ErrorIs(t, err, duration.ErrUnset)
	assert.Panics(t, func() { g.MustUnit() })
	assert.Panics(t, func() { g.MustTuplet() })
	assert.Panics(t, func() { g.MustContents() })
	assert.Panics(t, func() { g.MustSpan() })
	assert.Panics(t, func() { g.MustShortfall() })
	assert.Panics(t, func() { g.MustSpans() })

	assert.False(t, duration.GroupContent(duration.Group{}).IsValid())
	assert.False(t, duration.Content{}.IsValid())
	_, err = duration.Content{}.Written()
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)
	_, err = duration.Content{}.Group()
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)
	assert.Panics(t, func() { duration.Content{}.MustWritten() })
	assert.Panics(t, func() { duration.Content{}.MustGroup() })

	w := duration.WrittenContent(duration.MustNewWritten(duration.Eighth, 0))
	_, err = w.Group()
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)
	gc := duration.GroupContent(duration.MustNewGroup(tuplet.Triplet(), duration.MustNewWritten(duration.Eighth, 0)))
	_, err = gc.Written()
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)
}

func TestReadingEqualAndOrder(t *testing.T) {
	t.Parallel()

	var r duration.Reading
	_, err := r.Contents()
	assert.ErrorIs(t, err, duration.ErrInvalidArgument)
	assert.Panics(t, func() { r.MustContents() })
	assert.True(t, r.Equal(duration.Reading{}))

	// Several readings with the same group count exercise the content order.
	ds := []duration.Duration{
		duration.MustNew(1, 8), duration.MustNew(1, 8),
		duration.MustNew(1, 8), duration.MustNew(1, 8),
	}
	opts := duration.ReadingsOptions{Allowed: []tuplet.Tuplet{tuplet.Triplet(), tuplet.MustNew(2, 2)}}
	a, err := duration.Readings(ds, opts)
	require.NoError(t, err)
	b, err := duration.Readings(ds, opts)
	require.NoError(t, err)
	require.Equal(t, len(a), len(b))
	require.GreaterOrEqual(t, len(a), 3)
	for i := range a {
		assert.True(t, a[i].Equal(b[i]))
	}
	assert.False(t, a[0].Equal(a[len(a)-1]))
}

func TestSubdivideTupletMustSuccess(t *testing.T) {
	t.Parallel()

	count, part := duration.MustNewWritten(duration.Quarter, 0).MustSubdivideTuplet(tuplet.Triplet())
	assert.Equal(t, 3, count)
	assert.True(t, part.IsSet())
}
