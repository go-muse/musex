package note_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reading parts, signs and the source (architecture 6.5; OP-8...OP-10;
// K-28 for the unset note) and the current set (architecture 6.8, OP-11).

// TestReadParts reads every part of notes with and without octave; none of
// the readers reads the source.
func TestReadParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     spelling.Spelling
		aspn      int
		hasOctave bool
		c         cents.Cents
		bound     bool
	}{
		{name: "C♯4 +30 bound", input: spelling.CSharp(), aspn: 4, hasOctave: true, c: ct(30), bound: true},
		{name: "B♭2 -12 unbound", input: spelling.BFlat(), aspn: 2, hasOctave: true, c: ct(-12)},
		{name: "D +1/2 octave -1 unbound", input: sp(letter.D, 1, 2), aspn: -1, hasOctave: true, c: cents.MustFromFraction(1, 3)},
		{name: "E without octave bound", input: spelling.E(), c: ct(0), bound: true},
		{name: "F𝄫 without octave unbound", input: spelling.FDoubleFlat(), c: ct(5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			var s settings.Source
			if tt.bound {
				s = src
			}
			var n note.Note
			if tt.hasOctave {
				n = mk(t, tt.input, tt.aspn, tt.c, s)
			} else {
				n = mkName(t, tt.input, tt.c, s)
			}
			before := src.reads()

			assert.True(t, n.IsValid())
			assert.Equal(t, tt.hasOctave, n.HasOctave())
			assert.True(t, n.Spelling().Equal(tt.input))
			assert.Equal(t, tt.input.Letter(), n.Letter())
			assert.True(t, n.Alteration().Equal(tt.input.Alteration()))
			assert.True(t, n.Cents().Equal(tt.c))
			assert.Equal(t, tt.input.IsNatural(), n.IsNatural())
			assert.Equal(t, tt.input.IsInteger(), n.IsInteger())
			assert.Equal(t, tt.input.Direction(), n.Direction())
			assert.Equal(t, tt.bound, n.IsBound())
			assert.True(t, n.Source() == s)

			p, perr := n.Pitch()
			o, oerr := n.Octave()
			a, aerr := n.ASPN()
			if tt.hasOctave {
				require.NoError(t, perr)
				require.NoError(t, oerr)
				require.NoError(t, aerr)
				assert.True(t, p.Equal(pt(tt.input, tt.aspn)))
				assert.True(t, o.Equal(oct(tt.aspn)))
				assert.Equal(t, tt.aspn, a)
				assert.True(t, n.MustPitch().Equal(p))
				assert.True(t, n.MustOctave().Equal(o))
				assert.Equal(t, a, n.MustASPN())
			} else {
				assertErr(t, perr, note.ErrNoOctave)
				assertErr(t, oerr, note.ErrNoOctave)
				assertErr(t, aerr, note.ErrNoOctave)
				assert.True(t, p == pitch.Pitch{})
				assert.True(t, o == octave.Octave{})
				assert.Zero(t, a)
				assertPanicErr(t, func() { n.MustPitch() }, note.ErrNoOctave)
				assertPanicErr(t, func() { n.MustOctave() }, note.ErrNoOctave)
				assertPanicErr(t, func() { n.MustASPN() }, note.ErrNoOctave)
			}
			assert.Equal(t, before, src.reads(), "readers do not read the source")
		})
	}
}

// TestReadUnsetNote covers K-28 for OP-8...OP-10 and U-1: the unset note gives
// zero values without errors, and the Must variants of the readers do not
// panic (architecture 6.5).
func TestReadUnsetNote(t *testing.T) {
	t.Parallel()

	var n note.Note
	assert.False(t, n.IsValid())
	assert.False(t, n.HasOctave())
	assert.True(t, n.Spelling() == spelling.Spelling{})
	assert.False(t, n.Letter().IsValid())
	assert.True(t, n.Alteration().Equal(alteration.Natural()))
	assert.True(t, n.Cents().Equal(cents.Zero()))
	assert.False(t, n.IsNatural())
	assert.False(t, n.IsInteger())
	assert.Equal(t, spelling.DirectionUnset, n.Direction())
	assert.Nil(t, n.Source())
	assert.False(t, n.IsBound())

	p, err := n.Pitch()
	assert.NoError(t, err)
	assert.True(t, p == pitch.Pitch{})
	o, err := n.Octave()
	assert.NoError(t, err)
	assert.True(t, o == octave.Octave{})
	a, err := n.ASPN()
	assert.NoError(t, err)
	assert.Zero(t, a)

	assertNoPanic(t, func() { _ = n.MustPitch() })
	assertNoPanic(t, func() { _ = n.MustOctave() })
	assertNoPanic(t, func() { _ = n.MustASPN() })
}

// TestSignsOfNames checks the signs of names (OP-9) on several directions.
func TestSignsOfNames(t *testing.T) {
	t.Parallel()

	for _, name := range []spelling.Spelling{
		spelling.C(), spelling.CSharp(), spelling.DFlat(), sp(letter.E, -1, 2), spelling.BDoubleSharp(),
	} {
		n := mkName(t, name, ct(0), nil)
		assert.Equal(t, name.IsNatural(), n.IsNatural(), "%v", name)
		assert.Equal(t, name.IsInteger(), n.IsInteger(), "%v", name)
		assert.Equal(t, name.Direction(), n.Direction(), "%v", name)
	}
}

// TestSettings covers OP-11: the snapshot of the own source; E-1, E-2, E-3;
// one read (architecture 6.8).
func TestSettings(t *testing.T) {
	t.Parallel()

	t.Run("bound returns the set of the source", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.C(), 4, ct(0), s)
		got, err := n.Settings()
		require.NoError(t, err)
		assert.Same(t, s, got, "a set is a source that returns itself")
		assert.Same(t, s, n.MustSettings())
	})

	t.Run("snapshot of a switching source", func(t *testing.T) {
		t.Parallel()
		a, b := std(), set19()
		src := newSwitch(a)
		n := mk(t, spelling.C(), 4, ct(0), src)
		assert.Same(t, a, n.MustSettings())
		src.set(b)
		assert.Same(t, b, n.MustSettings())
	})

	t.Run("one read", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), src)
		before := src.reads()
		_, err := n.Settings()
		require.NoError(t, err)
		assert.EqualValues(t, 1, src.reads()-before)
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		unbound := mk(t, spelling.C(), 4, ct(0), nil)
		_, err := unbound.Settings()
		assertErr(t, err, note.ErrNoSource)
		assertPanicErr(t, func() { unbound.MustSettings() }, note.ErrNoSource)

		var unset note.Note
		got, err := unset.Settings()
		assertErr(t, err, note.ErrUnset)
		assert.Nil(t, got)

		src := newSwitch(std())
		bound := mk(t, spelling.C(), 4, ct(0), src)
		src.set(nil)
		got, err = bound.Settings()
		assertErr(t, err, note.ErrUnsetSettings, settings.ErrUnset)
		assert.Nil(t, got)
	})
}
