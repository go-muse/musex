package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRead covers O-4 and S-1: reading the counts of a set tuplet.
func TestRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		actual, normal int
	}{
		{actual: 3, normal: 2},
		{actual: 6, normal: 4},
		{actual: 1, normal: 1},
		{actual: math.MaxInt, normal: 1},
		{actual: 1, normal: math.MaxInt},
		{actual: math.MaxInt, normal: math.MaxInt - 1},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.actual)+":"+itoa(tt.normal), func(t *testing.T) {
			t.Parallel()

			tp := tuplet.MustNew(tt.actual, tt.normal)
			assert.True(t, tp.IsSet())

			actual, err := tp.Actual()
			require.NoError(t, err)
			assert.Equal(t, tt.actual, actual)

			normal, err := tp.Normal()
			require.NoError(t, err)
			assert.Equal(t, tt.normal, normal)

			a, n, err := tp.Pair()
			require.NoError(t, err)
			assert.Equal(t, tt.actual, a)
			assert.Equal(t, tt.normal, n)

			assert.Equal(t, tt.actual, tp.MustActual())
			assert.Equal(t, tt.normal, tp.MustNormal())

			a, n = tp.MustPair()
			assert.Equal(t, tt.actual, a)
			assert.Equal(t, tt.normal, n)
		})
	}
}

// TestIsSet covers P-3: the zero value is the unset tuplet.
func TestIsSet(t *testing.T) {
	t.Parallel()

	var zero tuplet.Tuplet

	assert.False(t, zero.IsSet())
	assert.False(t, tuplet.Tuplet{}.IsSet())
	assert.True(t, tuplet.Triplet().IsSet())
	assert.True(t, tuplet.MustNew(1, 1).IsSet())
}

// TestKind covers KP-10, G-1 and IP-7.
func TestKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		actual, normal int
		kind           tuplet.Kind
	}{
		{actual: 3, normal: 2, kind: tuplet.Compressing},
		{actual: 2, normal: 3, kind: tuplet.Expanding},
		{actual: 4, normal: 4, kind: tuplet.Neutral},
		{actual: 1, normal: 1, kind: tuplet.Neutral},
		{actual: 1, normal: 2, kind: tuplet.Expanding},
		{actual: 2, normal: 1, kind: tuplet.Compressing},
		{actual: 6, normal: 4, kind: tuplet.Compressing},
		{actual: math.MaxInt, normal: math.MaxInt - 1, kind: tuplet.Compressing},
		{actual: math.MaxInt - 1, normal: math.MaxInt, kind: tuplet.Expanding},
		{actual: math.MaxInt, normal: math.MaxInt, kind: tuplet.Neutral},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.actual)+":"+itoa(tt.normal), func(t *testing.T) {
			t.Parallel()

			tp := tuplet.MustNew(tt.actual, tt.normal)

			kind, err := tp.Kind()
			require.NoError(t, err)
			assert.Equal(t, tt.kind, kind)
			assert.Equal(t, tt.kind, tp.MustKind())

			compressing, err := tp.IsCompressing()
			require.NoError(t, err)
			assert.Equal(t, tt.kind == tuplet.Compressing, compressing)
			assert.Equal(t, compressing, tp.MustIsCompressing())

			expanding, err := tp.IsExpanding()
			require.NoError(t, err)
			assert.Equal(t, tt.kind == tuplet.Expanding, expanding)
			assert.Equal(t, expanding, tp.MustIsExpanding())

			neutral, err := tp.IsNeutral()
			require.NoError(t, err)
			assert.Equal(t, tt.kind == tuplet.Neutral, neutral)
			assert.Equal(t, neutral, tp.MustIsNeutral())
		})
	}
}
