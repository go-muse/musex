package scale_test

import (
	"iter"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func names(text string) []note.Note {
	result := make([]note.Note, 0)
	for _, text := range strings.Fields(text) {
		result = append(result, note.MustFromSpelling(spelling.MustParse(text), cents.Zero(), nil))
	}
	return result
}
func at(name string, aspn int, delta cents.Cents, src settings.Source) note.Note {
	return note.MustFromASPN(spelling.MustParse(name), aspn, delta, src)
}
func list(text string) *scale.Scale { return scale.MustNew(names(text)...) }
func checkNotes(t *testing.T, want, got []note.Note) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for i := range want {
		assert.True(t, want[i].Equal(got[i]), "note %d: want %s, got %s", i, want[i], got[i])
		assert.True(t, settings.SameSource(want[i].Source(), got[i].Source()), "source %d", i)
	}
}
func checkScale(t *testing.T, want []note.Note, got *scale.Scale) {
	t.Helper()
	require.NotNil(t, got)
	assert.True(t, got.IsValid())
	checkNotes(t, want, got.Notes())
}
func collect(seq iter.Seq[note.Note], limit int) []note.Note {
	result := make([]note.Note, 0)
	for n := range seq {
		result = append(result, n)
		if len(result) == limit {
			break
		}
	}
	return result
}

type source struct {
	set       *settings.Settings
	calls     atomic.Int64
	forbidden bool
}

func (s *source) Settings() *settings.Settings {
	s.calls.Add(1)
	if s.forbidden {
		panic("unexpected Settings call")
	}
	return s.set
}
func panicError(t *testing.T, target error, f func()) {
	t.Helper()
	defer func() {
		p := recover()
		e, ok := p.(error)
		require.True(t, ok, "panic must contain error, got %v", p)
		assert.ErrorIs(t, e, target)
	}()
	f()
	t.Fatal("expected panic")
}
