package degree_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func degreeTestNote(t *testing.T, name spelling.Spelling, c cents.Cents, src settings.Source) note.Note {
	t.Helper()
	n, err := note.FromSpelling(name, c, src)
	require.NoError(t, err)
	return n
}

func degreeTestASPNNote(t *testing.T, name spelling.Spelling, aspn int, c cents.Cents, src settings.Source) note.Note {
	t.Helper()
	n, err := note.FromASPN(name, aspn, c, src)
	require.NoError(t, err)
	return n
}

func assertDegreeState(t *testing.T, d degree.Degree, number int, hasNote bool, expected note.Note) {
	t.Helper()
	require.True(t, d.IsValid())
	assert.Equal(t, number, d.Number())
	assert.Equal(t, hasNote, d.HasNote())
	actual := d.Note()
	if !hasNote {
		assert.False(t, actual.IsValid())
		return
	}
	require.True(t, actual.IsValid())
	assert.True(t, expected.Equal(actual), "degree note differs: got %s, want %s", actual, expected)
}

func assertSameDegree(t *testing.T, expected, actual degree.Degree) {
	t.Helper()
	if !expected.IsValid() {
		assertUnsetDegree(t, actual)
		return
	}
	require.True(t, actual.IsValid())
	assert.True(t, expected.Equal(actual), "degrees differ: got %s, want %s", actual, expected)
}

func assertUnsetDegree(t *testing.T, d degree.Degree) {
	t.Helper()
	assert.False(t, d.IsValid())
	assert.Zero(t, d.Number())
	assert.False(t, d.HasNote())
	assert.False(t, d.Note().IsValid())
	assert.Equal(t, "Degree(unset)", d.String())
}

func degreeErrorSentinels() []error {
	return []error{
		degree.ErrUnset,
		degree.ErrInvalidNumber,
		degree.ErrNoNote,
		degree.ErrParse,
		degree.ErrRomanOutOfRange,
		degree.ErrOutOfRange,
		degree.ErrTransformFailed,
	}
}

func assertDegreeError(t *testing.T, err, expected error) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, expected)
	for _, sentinel := range degreeErrorSentinels() {
		if errors.Is(sentinel, expected) {
			continue
		}
		assert.NotErrorIs(t, err, sentinel)
	}
}

func assertMustPanic(t *testing.T, expected error, fn func()) {
	t.Helper()
	require.Error(t, expected)
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		fn()
	}()
	require.NotNil(t, recovered, "Must variant did not panic")
	actual, ok := recovered.(error)
	require.True(t, ok, "panic value has type %T, not error", recovered)
	assert.True(t, expected == actual, "Must panic is not the error returned by the wrapped operation")
	assert.ErrorIs(t, actual, expected)
}
