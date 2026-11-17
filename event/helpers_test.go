package event_test

import (
	"strconv"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func indexName(index int) string {
	return "index " + strconv.Itoa(index)
}

func assertFailure[T any](t *testing.T, want error, call func() (T, error), must func() T) {
	t.Helper()

	value, err := call()
	assert.Zero(t, value, "failure must return a zero result")
	assertCause(t, want, err)

	panicValue := func() (recovered any) {
		defer func() { recovered = recover() }()
		must()
		return nil
	}()
	panicErr, isError := panicValue.(error)
	require.True(t, isError, "Must should panic with an error, got %T (%v)", panicValue, panicValue)
	assertCause(t, want, panicErr)
}

func assertCause(t *testing.T, want, got error) {
	t.Helper()
	require.ErrorIs(t, got, want)
	for _, other := range []error{
		event.ErrUnset, event.ErrUnsetContent, event.ErrUnsetDuration,
		event.ErrEmptyChord, event.ErrMeasureCount, event.ErrWrongKind, event.ErrWrongDurationKind,
		event.ErrMeasureContent, event.ErrIncompatibleMark, event.ErrIndex, event.ErrTiedNoteChanged,
		event.ErrNeedsMeasureContext, note.ErrUnset, chord.ErrUnset, duration.ErrUnset,
		duration.ErrInvalidArgument, duration.ErrNotPositive, duration.ErrOutOfRange,
		fraction.ErrInvalid, fraction.ErrOutOfRange,
	} {
		if other != want {
			assert.NotErrorIs(t, got, other)
		}
	}
}

func assertFraction(t *testing.T, want, got fraction.Fraction) {
	t.Helper()
	require.True(t, got.IsValid())
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
}

func writtenDuration(base duration.Base) event.Duration {
	return event.MustWrittenDuration(base.MustWritten())
}

func valueDuration(num, den int) event.Duration {
	return event.MustValueDuration(fraction.MustNew(num, den))
}
