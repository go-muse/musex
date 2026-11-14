package note_test

import (
	"errors"
	"math"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The source of a note: change, check, unbind (architecture 6.7; OP-28...OP-31;
// P-11; K-13...K-18a, K-37, K-49; G-5).

// TestSetSource covers OP-28 and OP-29 on the control examples.
func TestSetSource(t *testing.T) {
	t.Parallel()

	t.Run("K-13", func(t *testing.T) {
		t.Parallel()
		s, s19 := std(), set19()
		n := mk(t, spelling.BSharp(), 4, ct(0), s) // "B♯3"
		copyBefore := n
		x, y := n, n

		require.NoError(t, x.SetSource(s19))
		assertNote(t, wp(spelling.BSharp(), 4, ct(0)).boundTo(s19), x)
		assert.Equal(t, "B♯4", x.MustFormat())
		assert.InDelta(t, 509.1285, x.MustFrequency(), 0.00005)
		assert.Equal(t, s19.Tuning().MustFrequency(pt(spelling.BSharp(), 4)), x.MustFrequency())

		require.NoError(t, y.SetSourceWritten(s19))
		assertNote(t, wp(spelling.BSharp(), 3, ct(0)).boundTo(s19), y)
		assert.Equal(t, "B♯3", y.MustFormat())
		assert.InDelta(t, 254.5643, y.MustFrequency(), 0.00005)

		assert.True(t, copyBefore == n, "a copy taken before the change keeps the old source")
		assert.True(t, n.Source() == settings.Source(s))
		assert.InDelta(t, 261.6256, n.MustFrequency(), 0.00005)
	})

	t.Run("K-14", func(t *testing.T) {
		t.Parallel()
		s24 := set24()
		n := mk(t, cHalf(), 4, ct(0), s24)
		before := n
		assertErr(t, n.SetSource(std()), note.ErrInadmissible, system.ErrInadmissible)
		assertSame(t, before, n)
		assert.True(t, n.Source() == settings.Source(s24))
	})

	t.Run("K-15", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.ADoubleFlat(), 4, ct(0), std())
		before := n
		assertErr(t, n.SetSource(stdNotation("german-am")), note.ErrInexpressible)
		assertSame(t, before, n)
	})

	t.Run("K-16", func(t *testing.T) {
		t.Parallel()
		ga := stdNotation("german-am")
		n := mk(t, spi(letter.C, 3), 4, ct(0), std())
		require.NoError(t, n.SetSource(ga), "not expressible before: nothing to lose")
		assert.True(t, n.Source() == settings.Source(ga))
	})

	t.Run("K-17", func(t *testing.T) {
		t.Parallel()
		ri := stdNotation("romance-italian")
		n := mk(t, spelling.CSharp(), 4, ct(0), std())
		assertErr(t, n.SetSource(ri), note.ErrInexpressible)
		m := mkName(t, spelling.CSharp(), ct(0), std())
		require.NoError(t, m.SetSource(ri))
		assert.True(t, m.Source() == settings.Source(ri))
	})

	t.Run("K-18a", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.CSharp(), 4, ct(3), nil)
		require.NoError(t, n.SetSource(s))
		assertNote(t, wp(spelling.CSharp(), 4, ct(3)).boundTo(s), n)

		h := mk(t, cHalf(), 4, ct(0), nil)
		assertErr(t, h.SetSource(s), note.ErrInadmissible, system.ErrInadmissible)
		assert.False(t, h.IsBound())

		u := mk(t, spelling.CSharp(), 4, ct(0), nil)
		chk, err := u.CheckSetSourceWritten(s)
		require.NoError(t, err)
		assert.False(t, chk.Possible())
		v := chk.Violations()
		require.Len(t, v, 1)
		assertErr(t, v[0], note.ErrNoSource)
	})

	t.Run("K-49 and G-5", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.C(), 4, ct(5), s)
		copyBefore := n

		same := n
		require.NoError(t, same.SetSource(s))
		assert.True(t, same.Equal(copyBefore))
		assert.True(t, same.SameSource(copyBefore))
		assert.True(t, same.Source() == settings.Source(s))

		other := settings.Standard()
		moved := n
		require.NoError(t, moved.SetSource(other))
		assert.True(t, moved.Equal(copyBefore))
		assert.True(t, moved.SameSource(copyBefore), "equal sets are one source")
		assert.True(t, moved.Source() == settings.Source(other), "the source is the given object")

		w := n
		require.NoError(t, w.SetSourceWritten(other))
		assert.True(t, w.Equal(copyBefore))
		assert.True(t, w.Source() == settings.Source(other))
	})

	t.Run("source with unset settings: unbind first", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		n := mk(t, spelling.C(), 4, ct(5), src)
		src.set(nil)
		s := std()
		assertErr(t, n.SetSource(s), note.ErrUnsetSettings, settings.ErrUnset)
		assertErr(t, n.SetSourceWritten(s), note.ErrUnsetSettings, settings.ErrUnset)
		require.NoError(t, n.Unbind())
		require.NoError(t, n.SetSource(s))
		assertNote(t, wp(spelling.C(), 4, ct(5)).boundTo(s), n)
	})

	t.Run("SetSourceWritten without octave works like SetSource", func(t *testing.T) {
		t.Parallel()
		s19 := set19()
		n := mkName(t, spelling.BSharp(), ct(0), nil)
		require.NoError(t, n.SetSourceWritten(s19))
		assertNote(t, wn(spelling.BSharp(), ct(0)).boundTo(s19), n)
	})

	t.Run("SetSourceWritten unbound with octave", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), nil)
		before := n
		assertErr(t, n.SetSourceWritten(std()), note.ErrNoSource)
		assertSame(t, before, n)
	})

	t.Run("SetSourceWritten written octave out of range", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.BSharp(), math.MinInt, ct(0), std())
		assertErr(t, n.SetSourceWritten(set19()), note.ErrOutOfRange, system.ErrOutOfRange)
	})

	t.Run("failures of arguments", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), std())
		assertErr(t, n.SetSource(nil), note.ErrUnset)
		assertErr(t, n.SetSourceWritten(nil), note.ErrUnset)
		assertErr(t, n.SetSource(&unsetSource{}), note.ErrUnsetSettings, settings.ErrUnset)
		assertErr(t, n.SetSource(nilSettingsSource()), note.ErrUnsetSettings, settings.ErrUnset)
		var u note.Note
		assertErr(t, u.SetSource(std()), note.ErrUnset)
	})
}

// TestUnbind covers OP-31 and K-37.
func TestUnbind(t *testing.T) {
	t.Parallel()

	s := std()
	n := mk(t, spelling.C(), 4, ct(5), s)
	copyBefore := n
	require.NoError(t, n.Unbind())
	assertNote(t, wp(spelling.C(), 4, ct(5)).free(), n)
	assert.True(t, n.Equal(copyBefore))
	assert.True(t, copyBefore.IsBound())

	require.NoError(t, n.Unbind(), "an unbound note stays unbound without an error")
	assertNote(t, wp(spelling.C(), 4, ct(5)).free(), n)

	var u note.Note
	assertErr(t, u.Unbind(), note.ErrUnset)
}

// TestSourceReads checks the reads of OP-28...OP-31 (P-5, P-5a).
func TestSourceReads(t *testing.T) {
	t.Parallel()

	t.Run("two sources", func(t *testing.T) {
		t.Parallel()
		a, b := newCounting(std()), newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), a)
		before := a.reads()
		require.NoError(t, n.SetSource(b))
		assert.EqualValues(t, 1, a.reads()-before)
		assert.EqualValues(t, 1, b.reads())
	})

	t.Run("one source", func(t *testing.T) {
		t.Parallel()
		a := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), a)
		before := a.reads()
		require.NoError(t, n.SetSourceWritten(a))
		assert.EqualValues(t, 1, a.reads()-before)
		_, err := n.CheckSetSource(a)
		require.NoError(t, err)
		assert.EqualValues(t, 2, a.reads()-before)
	})

	t.Run("unbind does not read", func(t *testing.T) {
		t.Parallel()
		a := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), a)
		before := a.reads()
		require.NoError(t, n.Unbind())
		assert.Equal(t, before, a.reads())
	})

	t.Run("unbound note reads only the new source", func(t *testing.T) {
		t.Parallel()
		b := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), nil)
		require.NoError(t, n.SetSource(b))
		assert.EqualValues(t, 1, b.reads())
	})
}

// TestCheckSetSource covers OP-30 and SourceCheck: K-18, K-18a; the order of
// violations; the zero value; a new slice from Violations.
func TestCheckSetSource(t *testing.T) {
	t.Parallel()

	t.Run("K-18 both conditions", func(t *testing.T) {
		t.Parallel()
		s24 := set24()
		n := mk(t, cHalf(), 4, ct(0), s24)
		require.True(t, s24.MustExpressesPitch(pt(cHalf(), 4)), "precondition: expressible in set 24")
		before := n
		chk, err := n.CheckSetSource(std())
		require.NoError(t, err)
		assertSame(t, before, n, "the check does not change the note")
		assert.False(t, chk.Possible())
		v := chk.Violations()
		require.Len(t, v, 2)
		assertErr(t, v[0], note.ErrInadmissible, system.ErrInadmissible)
		assertErr(t, v[1], note.ErrInexpressible)
		e := chk.Err()
		assert.ErrorIs(t, e, note.ErrInadmissible)
		assert.ErrorIs(t, e, note.ErrInexpressible)
		assert.NotErrorIs(t, e, note.ErrUnset)
	})

	t.Run("possible", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.CSharp(), 4, ct(0), std())
		chk, err := n.CheckSetSource(set19())
		require.NoError(t, err)
		assert.True(t, chk.Possible())
		assert.Empty(t, chk.Violations())
		assert.NoError(t, chk.Err())
		assert.True(t, n.MustCheckSetSource(set19()).Possible())
		assert.True(t, n.MustCheckSetSourceWritten(set19()).Possible())
	})

	t.Run("expressibility", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.ADoubleFlat(), 4, ct(0), std())
		chk, err := n.CheckSetSourceWritten(stdNotation("german-am"))
		require.NoError(t, err)
		require.Len(t, chk.Violations(), 1)
		assertErr(t, chk.Violations()[0], note.ErrInexpressible)
	})

	t.Run("written octave out of range", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.BSharp(), math.MinInt, ct(0), std())
		chk, err := n.CheckSetSourceWritten(set19())
		require.NoError(t, err)
		assert.False(t, chk.Possible())
		v := chk.Violations()
		require.Len(t, v, 1, "E-7 is checked only for a representable result")
		assertErr(t, v[0], note.ErrOutOfRange, system.ErrOutOfRange)
	})

	t.Run("no source skips the conditions that need the old set", func(t *testing.T) {
		t.Parallel()
		n := mk(t, cHalf(), 4, ct(0), nil)
		chk, err := n.CheckSetSourceWritten(std())
		require.NoError(t, err)
		v := chk.Violations()
		require.Len(t, v, 2)
		assertErr(t, v[0], note.ErrNoSource)
		assertErr(t, v[1], note.ErrInadmissible, system.ErrInadmissible)
	})

	t.Run("unbound note: only admissibility", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.ADoubleFlat(), 4, ct(0), nil)
		chk, err := n.CheckSetSource(stdNotation("german-am"))
		require.NoError(t, err)
		assert.True(t, chk.Possible(), "an unbound note has no old set to lose expressibility")
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), std())
		for _, call := range []func() (note.SourceCheck, error){
			func() (note.SourceCheck, error) { return n.CheckSetSource(nil) },
			func() (note.SourceCheck, error) { return n.CheckSetSourceWritten(nil) },
			func() (note.SourceCheck, error) { return note.Note{}.CheckSetSource(std()) },
		} {
			chk, err := call()
			assertErr(t, err, note.ErrUnset)
			assert.False(t, chk.Possible())
		}
		chk, err := n.CheckSetSource(&unsetSource{})
		assertErr(t, err, note.ErrUnsetSettings, settings.ErrUnset)
		assert.False(t, chk.Possible())
		assertErr(t, chk.Err(), note.ErrUnset)

		stale := onStale(spelling.C(), 4)(t)
		_, err = stale.CheckSetSource(std())
		assertErr(t, err, note.ErrUnsetSettings, settings.ErrUnset)

		assertPanicErr(t, func() { n.MustCheckSetSource(nil) }, note.ErrUnset)
		assertPanicErr(t, func() { n.MustCheckSetSourceWritten(&unsetSource{}) }, note.ErrUnsetSettings)
	})

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()
		var chk note.SourceCheck
		assert.False(t, chk.Possible())
		assert.Empty(t, chk.Violations())
		assertErr(t, chk.Err(), note.ErrUnset)
	})

	t.Run("Violations returns a new slice", func(t *testing.T) {
		t.Parallel()
		n := mk(t, cHalf(), 4, ct(0), set24())
		chk, err := n.CheckSetSource(std())
		require.NoError(t, err)
		v := chk.Violations()
		require.Len(t, v, 2)
		v[0] = errors.New("changed")
		v2 := chk.Violations()
		assert.ErrorIs(t, v2[0], note.ErrInadmissible)
		cp := chk
		assert.Len(t, cp.Violations(), 2)
	})
}
