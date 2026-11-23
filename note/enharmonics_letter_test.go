package note_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EnharmonicsOnLetter filters Enharmonics by the written letter, keeping the
// order, the position, the correction, the octave presence and the source.
func TestEnharmonicsOnLetter(t *testing.T) {
	t.Parallel()

	t.Run("C sharp 4 on D with limit one gives D flat 4", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		n := mk(t, spelling.CSharp(), 4, ct(30), src)
		before := n
		seq, err := n.EnharmonicsOnLetter(letter.D, alteration.Sharp())
		require.NoError(t, err)
		wantOne := []want{wp(spelling.DFlat(), 4, ct(30)).boundTo(src)}
		assertNotes(t, wantOne, collect(seq))
		assertNotes(t, wantOne, collect(seq))
		assertSame(t, before, n)
		assertNotes(t, wantOne, collect(n.MustEnharmonicsOnLetter(letter.D, alteration.Sharp())))
	})

	t.Run("keeps the order of Enharmonics", func(t *testing.T) {
		t.Parallel()
		src := std()
		n := mk(t, spelling.CSharp(), 4, ct(30), src)
		got := collect(n.MustEnharmonicsOnLetter(letter.B, alteration.DoubleSharp()))
		var ws []want
		for p := range system.Traditional().MustEnharmonicPitches(pt(spelling.CSharp(), 4), alteration.DoubleSharp()) {
			if p.Spelling().Letter().Equal(letter.B) {
				ws = append(ws, wp(p.Spelling(), p.MustASPN(), ct(30)).boundTo(src))
			}
		}
		require.NotEmpty(t, ws)
		assertNotes(t, ws, got)
	})

	t.Run("no match gives an empty sequence", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), std())
		seq, err := n.EnharmonicsOnLetter(letter.D, alteration.Natural())
		require.NoError(t, err)
		assert.Empty(t, collect(seq))
	})

	t.Run("without octave", func(t *testing.T) {
		t.Parallel()
		src := std()
		n := mkName(t, spelling.CSharp(), ct(-4), src)
		got := collect(n.MustEnharmonicsOnLetter(letter.D, alteration.Sharp()))
		assertNotes(t, []want{wn(spelling.DFlat(), ct(-4)).boundTo(src)}, got)
	})

	t.Run("the sequence can be stopped early", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), std())
		// The traditional system admits every integer alteration, so within a
		// limit of twelve units the letter C has three spellings at this
		// position: C-12, C and C+12. Breaking after the first one is a real
		// check: without the return after !yield(x) the next match would call
		// yield again and panic.
		limit := alteration.MustFromInt(12)
		full := collect(n.MustEnharmonicsOnLetter(letter.C, limit))
		require.GreaterOrEqual(t, len(full), 2, "precondition: several results on the requested letter")
		seq, err := n.EnharmonicsOnLetter(letter.C, limit)
		require.NoError(t, err)
		count := 0
		for range seq {
			count++
			break
		}
		assert.Equal(t, 1, count)
	})

	t.Run("unset note", func(t *testing.T) {
		t.Parallel()
		var n note.Note
		seq, err := n.EnharmonicsOnLetter(letter.D, alteration.Sharp())
		assert.Nil(t, seq)
		assertErr(t, err, note.ErrUnset)
		assertPanicErr(t, func() { n.MustEnharmonicsOnLetter(letter.D, alteration.Sharp()) }, note.ErrUnset)
	})

	t.Run("unset letter", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), std())
		seq, err := n.EnharmonicsOnLetter(letter.Letter(0), alteration.Sharp())
		assert.Nil(t, seq)
		assertErr(t, err, note.ErrUnset)
	})

	t.Run("negative limit", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), std())
		seq, err := n.EnharmonicsOnLetter(letter.D, alteration.Flat())
		assert.Nil(t, seq)
		assertErr(t, err, nil, system.ErrLimit)
	})

	t.Run("unbound note", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), nil)
		seq, err := n.EnharmonicsOnLetter(letter.D, alteration.Sharp())
		assert.Nil(t, seq)
		assertErr(t, err, note.ErrNoSource)
	})
}

// C♯4 spelled on the letter D with a limit of one chromatic unit is D♭4 in
// the traditional system.
func ExampleNote_EnharmonicsOnLetter() {
	std := settings.Standard()
	cs := note.MustFromASPN(spelling.CSharp(), 4, cents.Zero(), std)
	for n := range cs.MustEnharmonicsOnLetter(letter.D, alteration.Sharp()) {
		fmt.Println(n.MustFormatIn(std))
	}
	// Output:
	// D♭4
}
