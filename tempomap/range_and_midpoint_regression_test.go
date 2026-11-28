package tempomap_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Specification 4.9, EqualTimeline step 2: EqualTimeline compares the exact speed after
// every boundary and in the midpoint of every open interval. The linear
// gradual map and the jump map below have equal speeds at the boundaries 0
// and 4 and different speeds at 2, so they are not equal as timelines.
func TestEqualTimelineDiffersOnlyInsideInterval(t *testing.T) {
	t.Parallel()
	gradual := tempomap.MustParse("tempomap[first@p:0/1:main:4=60;second@p:0/1:gradual:linear-speed;single@p:4/1:main:4=120]")
	steps := tempomap.MustParse("tempomap[first@p:0/1:main:4=60;single@p:4/1:main:4=120]")

	for _, boundary := range []tempomap.Address{at(0, 1), at(4, 1)} {
		assert.True(t, gradual.MustSpeed(boundary).Equal(steps.MustSpeed(boundary)), "speed at the boundary %s", boundary)
	}
	assert.False(t, gradual.MustSpeed(at(2, 1)).Equal(steps.MustSpeed(at(2, 1))), "speed inside the interval")

	for _, tt := range []struct {
		name        string
		left, right tempomap.Map
	}{
		{"gradual against steps", gradual, steps},
		{"steps against gradual", steps, gradual},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			equal, err := tt.left.EqualTimeline(tt.right)
			require.NoError(t, err)
			assert.False(t, equal)
			assert.False(t, tt.left.MustEqualTimeline(tt.right))
		})
	}
}

// Specification 4.5 (TempoAt), 4.11 and Е-10: an exact marking outside the range
// of common/fraction is reported as tempomap.ErrOutOfRange with the cause
// fraction.ErrOutOfRange. With ♩ = M the speed M/4 is in range, while the
// marking in eighths is 2M and the marking in halves M/2 is in range.
func TestTempoAtOutOfRange(t *testing.T) {
	t.Parallel()
	maxBPM := fraction.MustParse(fraction.MaxComponentString)
	m := tempomap.MustNew(tempo.MustNewQuarter(maxBPM))
	origin := tempomap.MustAt(duration.Origin())

	t.Run("marking in eighths", func(t *testing.T) {
		t.Parallel()
		unit := duration.Eighth.MustWritten()
		got, err := m.TempoAt(origin, unit)
		requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, tempo.Tempo{}, got)
		requirePanicError(t, err, tempomap.ErrOutOfRange, func() { m.MustTempoAt(origin, unit) })
	})

	t.Run("markings in range", func(t *testing.T) {
		t.Parallel()
		quarterMark, err := m.TempoAt(origin, duration.Quarter.MustWritten())
		require.NoError(t, err)
		assert.True(t, tempo.MustNewQuarter(maxBPM).Equal(quarterMark), "got %s", quarterMark)

		half := duration.Half.MustWritten()
		halfMark, err := m.TempoAt(origin, half)
		require.NoError(t, err)
		want := tempo.MustNew(half, maxBPM.MustDiv(fraction.FromInt(2)))
		assert.True(t, want.Equal(halfMark), "got %s, want %s", halfMark, want)
	})
}
