package note_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Must variants (architecture 6.11; OP-54): a Must variant calls the
// operation; on an error it panics with that error value (recover gives an
// error, errors.Is works), otherwise it returns the same results. The Must
// variants of the changing methods, computations, comparisons and sequences
// are checked next to their operations; this file covers the creation
// functions and the general rules.

// TestMustCreation checks the Must variant of every creation function on a
// success and on a failure.
func TestMustCreation(t *testing.T) {
	t.Parallel()

	s := std()
	fb := stdFB()
	tu := tuning.Standard()
	unset := &unsetSource{}
	tests := []struct {
		name    string
		plain   func(settings.Source) (note.Note, error)
		must    func(settings.Source) note.Note
		failSrc settings.Source // makes the call fail with failErr
		failErr error
	}{
		{"FromSpelling",
			func(src settings.Source) (note.Note, error) { return note.FromSpelling(spelling.CSharp(), ct(1), src) },
			func(src settings.Source) note.Note { return note.MustFromSpelling(spelling.CSharp(), ct(1), src) },
			unset, note.ErrUnsetSettings},
		{"FromPitch",
			func(src settings.Source) (note.Note, error) { return note.FromPitch(pt(spelling.D(), 2), ct(1), src) },
			func(src settings.Source) note.Note { return note.MustFromPitch(pt(spelling.D(), 2), ct(1), src) },
			unset, note.ErrUnsetSettings},
		{"FromOctave",
			func(src settings.Source) (note.Note, error) { return note.FromOctave(spelling.D(), oct(2), ct(1), src) },
			func(src settings.Source) note.Note { return note.MustFromOctave(spelling.D(), oct(2), ct(1), src) },
			unset, note.ErrUnsetSettings},
		{"FromASPN",
			func(src settings.Source) (note.Note, error) { return note.FromASPN(spelling.D(), 2, ct(1), src) },
			func(src settings.Source) note.Note { return note.MustFromASPN(spelling.D(), 2, ct(1), src) },
			unset, note.ErrUnsetSettings},
		{"FromWritten",
			func(src settings.Source) (note.Note, error) {
				return note.FromWritten(src, spelling.BSharp(), oct(3), ct(1))
			},
			func(src settings.Source) note.Note {
				return note.MustFromWritten(src, spelling.BSharp(), oct(3), ct(1))
			},
			nil, note.ErrUnset},
		{"FromWrittenASPN",
			func(src settings.Source) (note.Note, error) {
				return note.FromWrittenASPN(src, spelling.BSharp(), 3, ct(1))
			},
			func(src settings.Source) note.Note { return note.MustFromWrittenASPN(src, spelling.BSharp(), 3, ct(1)) },
			nil, note.ErrUnset},
		{"FromWrittenNumber",
			func(src settings.Source) (note.Note, error) {
				return note.FromWrittenNumber(src, spelling.BSharp(), 3, ct(1))
			},
			func(src settings.Source) note.Note {
				return note.MustFromWrittenNumber(src, spelling.BSharp(), 3, ct(1))
			},
			unset, note.ErrUnsetSettings},
		{"Parse",
			func(src settings.Source) (note.Note, error) { return note.Parse(src, "E♭") },
			func(src settings.Source) note.Note { return note.MustParse(src, "E♭") },
			nil, note.ErrUnset},
		{"ParsePitch",
			func(src settings.Source) (note.Note, error) { return note.ParsePitch(src, "E♭5") },
			func(src settings.Source) note.Note { return note.MustParsePitch(src, "E♭5") },
			unset, note.ErrUnsetSettings},
		{"FromStep",
			func(src settings.Source) (note.Note, error) { return note.FromStep(src, 5, system.Flats(), ct(1)) },
			func(src settings.Source) note.Note { return note.MustFromStep(src, 5, system.Flats(), ct(1)) },
			nil, note.ErrUnset},
		{"Nearest",
			func(src settings.Source) (note.Note, error) { return note.Nearest(src, 300, system.Flats()) },
			func(src settings.Source) note.Note { return note.MustNearest(src, 300, system.Flats()) },
			unset, note.ErrUnsetSettings},
		{"NearestBy",
			func(src settings.Source) (note.Note, error) {
				return note.NearestBy(src, 300, rounding.Up, system.Flats())
			},
			func(src settings.Source) note.Note { return note.MustNearestBy(src, 300, rounding.Up, system.Flats()) },
			nil, note.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := settings.Source(s)
			if tt.name == "FromWrittenNumber" {
				src = fb
			}
			want, err := tt.plain(src)
			require.NoError(t, err)
			var got note.Note
			assertNoPanic(t, func() { got = tt.must(src) })
			assert.True(t, want == got, "the same result")

			_, err = tt.plain(tt.failSrc)
			assertErr(t, err, tt.failErr)
			assertPanicErr(t, func() { tt.must(tt.failSrc) }, tt.failErr)
		})
	}

	t.Run("explicit tuning", func(t *testing.T) {
		t.Parallel()
		a, err := note.FromStepInTuning(tu, 5, system.Flats(), ct(1))
		require.NoError(t, err)
		assert.True(t, a == note.MustFromStepInTuning(tu, 5, system.Flats(), ct(1)))
		assertPanicErr(t, func() { note.MustFromStepInTuning(nil, 5, system.Flats(), ct(1)) }, note.ErrUnset)

		b, err := note.NearestInTuning(tu, 300, system.Flats())
		require.NoError(t, err)
		assert.True(t, b == note.MustNearestInTuning(tu, 300, system.Flats()))
		assertPanicErr(t, func() { note.MustNearestInTuning(&tuning.Tuning{}, 300, system.Flats()) }, note.ErrUnset)

		c, err := note.NearestByInTuning(tu, 300, rounding.Down, system.Flats())
		require.NoError(t, err)
		assert.True(t, c == note.MustNearestByInTuning(tu, 300, rounding.Down, system.Flats()))
		assertPanicErr(t, func() { note.MustNearestByInTuning(tu, 300, rounding.Mode(0), system.Flats()) }, note.ErrUnset)
	})

	t.Run("a reason of a dependency without a note sentinel", func(t *testing.T) {
		t.Parallel()
		assertPanicErr(t, func() { note.MustParse(s, "xyz") }, nil)
		assertPanicErr(t, func() { note.MustFromStep(s, 0, badRule(), ct(0)) }, nil, tuning.ErrSpelling)
	})
}

// TestMustPanicsWithTheSameError checks that the panic value of a Must
// variant is the error the plain form returns: an error with the same text
// that matches the same sentinels.
func TestMustPanicsWithTheSameError(t *testing.T) {
	t.Parallel()

	free := mk(t, spelling.C(), 4, ct(0), nil)
	_, plainErr := free.Frequency()
	require.Error(t, plainErr)
	v := panicValue(func() { free.MustFrequency() })
	panicErr, ok := v.(error)
	require.True(t, ok, "the panic value is an error")
	assert.Equal(t, plainErr.Error(), panicErr.Error())
	assert.ErrorIs(t, panicErr, note.ErrNoSource)

	n := mk(t, spelling.CSharp(), 4, ct(0), setK25())
	plain := n
	errRaise := plain.Raise()
	require.Error(t, errRaise)
	v = panicValue(func() { n.MustRaise() })
	panicErr, ok = v.(error)
	require.True(t, ok)
	assert.Equal(t, errRaise.Error(), panicErr.Error())
}
