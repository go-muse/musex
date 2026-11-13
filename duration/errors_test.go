package duration_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSentinels covers business-spec 2.6, T-10, and architecture 4.1: 11
// sentinels, pairwise distinct; none wraps another or a dependency sentinel.
func TestSentinels(t *testing.T) {
	t.Parallel()

	require.Len(t, durationSentinels, 11)

	for _, a := range durationSentinels {
		require.Error(t, a.err, a.name)
		assert.NotEmpty(t, a.err.Error(), a.name)

		for _, b := range durationSentinels {
			assert.Equal(t, a.name == b.name, errors.Is(a.err, b.err), "errors.Is(%s, %s)", a.name, b.name)
		}

		for _, d := range dependencySentinels {
			assert.False(t, errors.Is(a.err, d.err), "errors.Is(%s, %s)", a.name, d.name)
			assert.False(t, errors.Is(d.err, a.err), "errors.Is(%s, %s)", d.name, a.name)
		}
	}
}
