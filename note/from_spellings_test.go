package note_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exact function types of the new constructor.
var (
	_ func([]spelling.Spelling, settings.Source) (note.Notes, error) = note.FromSpellings
	_ func([]spelling.Spelling, settings.Source) note.Notes          = note.MustFromSpellings
)

// FromSpellings builds notes without octaves and with zero correction,
// preserving order and duplicates; the binding follows FromSpelling.
func TestFromSpellings(t *testing.T) {
	t.Parallel()

	t.Run("order duplicates binding and zero parts", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		names := []spelling.Spelling{spelling.C(), spelling.EFlat(), spelling.G(), spelling.C()}
		reads := src.reads()
		got, err := note.FromSpellings(names, src)
		require.NoError(t, err)
		assertNotes(t, []want{
			wn(spelling.C(), ct(0)).boundTo(src),
			wn(spelling.EFlat(), ct(0)).boundTo(src),
			wn(spelling.G(), ct(0)).boundTo(src),
			wn(spelling.C(), ct(0)).boundTo(src),
		}, got)
		assert.Equal(t, reads+1, src.reads(), "the source is read once")
	})

	t.Run("nil source gives unbound notes", func(t *testing.T) {
		t.Parallel()
		got, err := note.FromSpellings([]spelling.Spelling{spelling.C(), spelling.CSharp()}, nil)
		require.NoError(t, err)
		assertNotes(t, []want{
			wn(spelling.C(), ct(0)).free(),
			wn(spelling.CSharp(), ct(0)).free(),
		}, got)
	})

	t.Run("empty and nil names", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		reads := src.reads()
		for _, names := range [][]spelling.Spelling{nil, {}, make([]spelling.Spelling, 0)} {
			got, err := note.FromSpellings(names, src)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		}
		assert.Equal(t, reads, src.reads(), "an empty list reads no source")
		got, err := note.FromSpellings(nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("unset name", func(t *testing.T) {
		t.Parallel()
		names := []spelling.Spelling{spelling.C(), {}}
		got, err := note.FromSpellings(names, std())
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnset)
		assertPanicErr(t, func() { note.MustFromSpellings(names, std()) }, note.ErrUnset)
	})

	t.Run("unset settings", func(t *testing.T) {
		t.Parallel()
		got, err := note.FromSpellings([]spelling.Spelling{spelling.C()}, &unsetSource{})
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnsetSettings, settings.ErrUnset)
	})

	t.Run("inadmissible name", func(t *testing.T) {
		t.Parallel()
		names := []spelling.Spelling{spelling.C(), spelling.CDoubleSharp()}
		got, err := note.FromSpellings(names, setK25())
		assert.Nil(t, got)
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)
		assertPanicErr(t, func() { note.MustFromSpellings(names, setK25()) }, note.ErrInadmissible, system.ErrInadmissible)
	})

	t.Run("Must variant", func(t *testing.T) {
		t.Parallel()
		got := note.MustFromSpellings([]spelling.Spelling{spelling.C()}, nil)
		assertNotes(t, []want{wn(spelling.C(), ct(0)).free()}, got)
	})
}

// FromSpellings builds a list of notes without octaves from names, keeping
// order and duplicates.
func ExampleFromSpellings() {
	ns := note.MustFromSpellings([]spelling.Spelling{
		spelling.C(), spelling.EFlat(), spelling.G(), spelling.C(),
	}, nil)
	fmt.Println(len(ns), ns[0].HasOctave(), ns[1].Spelling().Equal(spelling.EFlat()))
	// Output:
	// 4 false true
}
