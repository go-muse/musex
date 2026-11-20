package voice_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Т-11, architecture 6.2: eighteen distinct sentinels with the prefix
// "voice: "; no sentinel wraps another.
func TestSentinels(t *testing.T) {
	t.Parallel()
	sentinels := voiceErrors()
	require.Len(t, sentinels, 18)
	for i, a := range sentinels {
		require.NotNil(t, a)
		assert.True(t, strings.HasPrefix(a.Error(), "voice: "), "%q", a.Error())
		for j, b := range sentinels {
			if i == j {
				continue
			}
			assert.NotErrorIs(t, a, b, "%v must not match %v", a, b)
		}
	}
}

// Е-16 and Е-17 do not match the system sentinels of interval (КР-V6).
func TestSystemSentinelsDifferFromInterval(t *testing.T) {
	t.Parallel()
	assert.NotErrorIs(t, voice.ErrNoSystem, interval.ErrNoSystem)
	assert.NotErrorIs(t, voice.ErrDifferentSystems, interval.ErrDifferentSystems)
	assert.NotErrorIs(t, interval.ErrNoSystem, voice.ErrNoSystem)
	assert.NotErrorIs(t, interval.ErrDifferentSystems, voice.ErrDifferentSystems)
}

func TestItemError(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()
		var e *voice.ItemError
		assert.Equal(t, "voice: item error", e.Error())
		assert.NoError(t, e.Unwrap())
	})

	t.Run("without a cause", func(t *testing.T) {
		t.Parallel()
		e := &voice.ItemError{Index: 3}
		assert.NotEmpty(t, e.Error())
		assert.NoError(t, e.Unwrap())
	})

	t.Run("with a cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.Join(voice.ErrUnset, note.ErrUnset)
		e := &voice.ItemError{Index: 2, Err: cause}
		assert.NotEmpty(t, e.Error())
		assert.Equal(t, cause, e.Unwrap())
		var err error = e
		assert.ErrorIs(t, err, voice.ErrUnset)
		assert.ErrorIs(t, err, note.ErrUnset)
		got, ok := errors.AsType[*voice.ItemError](err)
		require.True(t, ok)
		assert.Equal(t, 2, got.Index)
	})
}
