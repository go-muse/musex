package voice_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Architecture 6.1, Т-11: a Must variant panics with the error of its
// operation, so that recover gives an error that keeps every cause and the
// index of *voice.ItemError. The operation-specific tests check every Must
// variant with assertFailure; this test checks the panic value itself.
func TestMustPanicValue(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), snd(free(spelling.E(), 4), q()))
	value := recovered(func() { l.MustIntervals() })
	panicErr, ok := value.(error)
	require.True(t, ok, "the panic value is an error, got %T", value)
	assert.ErrorIs(t, panicErr, voice.ErrNoSystem)
	itemErr, ok := errors.AsType[*voice.ItemError](panicErr)
	require.True(t, ok)
	assert.Equal(t, 1, itemErr.Index)
}

// The successful Must variants of the counters return the results of their
// operations.
func TestMustCounters(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			n, err := l.Len()
			require.NoError(t, err)
			assert.Equal(t, n, l.MustLen())
			empty, err := l.IsEmpty()
			require.NoError(t, err)
			assert.Equal(t, empty, l.MustIsEmpty())
			assert.Equal(t, n == 0, empty)
			sounds, err := l.SoundCount()
			require.NoError(t, err)
			assert.Equal(t, sounds, l.MustSoundCount())
			rests, err := l.RestCount()
			require.NoError(t, err)
			assert.Equal(t, rests, l.MustRestCount())
			assert.Equal(t, n, sounds+rests)
		})
	}
}

func TestMustTessituraIn(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()))
	assertFailure(t, fails(voice.ErrUnset, system.ErrInvalid),
		func() (voice.Tessitura, error) { return l.TessituraIn(nil, voice.PresetReference(soprano())) },
		func() voice.Tessitura { return l.MustTessituraIn(nil, voice.PresetReference(soprano())) })
	assertFailure(t, fails(voice.ErrNoRange).only(),
		func() (voice.Tessitura, error) {
			return l.TessituraIn(traditional(), voice.PlacementReference(voice.Middle))
		},
		func() voice.Tessitura {
			return l.MustTessituraIn(traditional(), voice.PlacementReference(voice.Middle))
		})
}
