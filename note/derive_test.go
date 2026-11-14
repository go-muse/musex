package note_test

import (
	"iter"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// New notes from a note (architecture 6.9; OP-25...OP-27; K-26, K-38, K-47;
// P-6, P-10).

// TestEnharmonics covers OP-25 and K-26: the notes of the same position in
// the order of system OP-31, with the cents and the source of the note.
func TestEnharmonics(t *testing.T) {
	t.Parallel()

	t.Run("K-26", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.CSharp(), 4, ct(30), s)
		before := n
		seq, err := n.Enharmonics(alteration.DoubleSharp())
		require.NoError(t, err)
		got := collect(seq)

		var ws []want
		for p := range system.Traditional().MustEnharmonicPitches(pt(spelling.CSharp(), 4), alteration.DoubleSharp()) {
			ws = append(ws, wp(p.Spelling(), p.MustASPN(), ct(30)).boundTo(s))
		}
		assertNotes(t, ws, got)
		require.Len(t, got, 3)
		assert.ElementsMatch(t,
			[]string{spelling.CSharp().String(), spelling.DFlat().String(), spelling.BDoubleSharp().String()},
			[]string{got[0].Spelling().String(), got[1].Spelling().String(), got[2].Spelling().String()})
		for _, x := range got {
			if x.Spelling().Equal(spelling.BDoubleSharp()) {
				assert.True(t, x.MustOctave().Equal(oct(4)), "B𝄪3 has Octave 4")
			}
		}
		assertSame(t, before, n)
		assert.Len(t, collect(n.MustEnharmonics(alteration.DoubleSharp())), 3)
	})

	t.Run("limit excludes the note itself", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.BDoubleSharp(), 3, ct(0), std())
		got := collect(n.MustEnharmonics(alteration.Sharp()))
		for _, x := range got {
			assert.False(t, x.Spelling().Equal(spelling.BDoubleSharp()))
		}
		assert.NotEmpty(t, got)
	})

	t.Run("without octave", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mkName(t, spelling.CSharp(), ct(-4), s)
		got := collect(n.MustEnharmonics(alteration.DoubleSharp()))
		var ws []want
		for name := range system.Traditional().MustEnharmonicNames(spelling.CSharp(), alteration.DoubleSharp()) {
			ws = append(ws, wn(name, ct(-4)).boundTo(s))
		}
		assertNotes(t, ws, got)
	})

	t.Run("EnharmonicsIn unbound stays unbound", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), nil)
		got := collect(n.MustEnharmonicsIn(set24(), alteration.DoubleSharp()))
		var ws []want
		for p := range system.QuarterTones24().MustEnharmonicPitches(pt(spelling.C(), 4), alteration.DoubleSharp()) {
			ws = append(ws, wp(p.Spelling(), p.MustASPN(), ct(0)).free())
		}
		assertNotes(t, ws, got)
	})

	t.Run("EnharmonicsIn checks every name in the own set before returning", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.C(), 4, ct(0), setK25())
		seq, err := n.EnharmonicsIn(std(), alteration.DoubleSharp()) // D𝄫 is not admissible in K-25
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)
		assert.Nil(t, seq)
		assertPanicErr(t, func() { n.MustEnharmonicsIn(std(), alteration.DoubleSharp()) }, note.ErrInadmissible)

		ok, err := n.EnharmonicsIn(std(), alteration.Sharp())
		require.NoError(t, err, "with limit 1 every name is admissible in K-25")
		for x := range ok {
			assert.True(t, x.Source() == n.Source())
		}
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.C(), 4, ct(0), s)
		cases := []struct {
			name  string
			call  func() (iter.Seq[note.Note], error)
			err   error
			extra []error
		}{
			{"unset note", func() (iter.Seq[note.Note], error) { return note.Note{}.Enharmonics(alteration.Sharp()) }, note.ErrUnset, nil},
			{"unbound", func() (iter.Seq[note.Note], error) {
				return mk(t, spelling.C(), 4, ct(0), nil).Enharmonics(alteration.Sharp())
			}, note.ErrNoSource, nil},
			{"unset settings", func() (iter.Seq[note.Note], error) {
				return onStale(spelling.C(), 4)(t).Enharmonics(alteration.Sharp())
			}, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"inadmissible", func() (iter.Seq[note.Note], error) {
				x, _ := inadmissibleNote(t)
				return x.Enharmonics(alteration.Sharp())
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
			{"negative limit", func() (iter.Seq[note.Note], error) { return n.Enharmonics(alteration.Flat()) }, nil, []error{system.ErrLimit}},
			{"inadmissible before negative limit", func() (iter.Seq[note.Note], error) {
				x, _ := inadmissibleNote(t)
				return x.Enharmonics(alteration.Flat())
			}, note.ErrInadmissible, nil},
			{"In nil source", func() (iter.Seq[note.Note], error) { return n.EnharmonicsIn(nil, alteration.Sharp()) }, note.ErrUnset, nil},
			{"In unset settings", func() (iter.Seq[note.Note], error) {
				return n.EnharmonicsIn(&unsetSource{}, alteration.Sharp())
			}, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"In name inadmissible in the explicit set", func() (iter.Seq[note.Note], error) {
				return mk(t, spelling.CDoubleSharp(), 4, ct(0), nil).EnharmonicsIn(setK25(), alteration.DoubleSharp())
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
		}
		for _, c := range cases {
			seq, err := c.call()
			assertErr(t, err, c.err, c.extra...)
			assert.Nil(t, seq, c.name)
		}
	})
}

// TestRange covers OP-26 and K-38: the same name, cents and source in each
// Octave from "from" to "to", up or down; the source is not read.
func TestRange(t *testing.T) {
	t.Parallel()

	t.Run("K-38", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		n := mkName(t, spelling.CSharp(), ct(10), src)
		reads := src.reads()
		got := collect(n.MustRange(oct(3), oct(5)))
		assertNotes(t, []want{
			wp(spelling.CSharp(), 3, ct(10)).boundTo(src),
			wp(spelling.CSharp(), 4, ct(10)).boundTo(src),
			wp(spelling.CSharp(), 5, ct(10)).boundTo(src),
		}, got)
		assert.Equal(t, reads, src.reads(), "group N")
	})

	t.Run("down", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.D(), 1, ct(0), nil)
		seq, err := n.Range(oct(5), oct(3))
		require.NoError(t, err)
		assertNotes(t, []want{wp(spelling.D(), 5, ct(0)).free(), wp(spelling.D(), 4, ct(0)).free(), wp(spelling.D(), 3, ct(0)).free()}, collect(seq))
	})

	t.Run("one octave", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.D(), 1, ct(0), nil)
		assertNotes(t, []want{wp(spelling.D(), 7, ct(0))}, collect(n.MustRange(oct(7), oct(7))))
	})

	t.Run("edges of int", func(t *testing.T) {
		t.Parallel()
		n := mkName(t, spelling.E(), ct(0), nil)
		var got []note.Note
		assertNoPanic(t, func() { got = collect(n.MustRange(oct(math.MinInt), oct(math.MinInt+2))) })
		assertNotes(t, []want{wp(spelling.E(), math.MinInt, ct(0)), wp(spelling.E(), math.MinInt+1, ct(0)), wp(spelling.E(), math.MinInt+2, ct(0))}, got)
		assertNoPanic(t, func() { got = collect(n.MustRange(oct(math.MaxInt), oct(math.MaxInt-2))) })
		assertNotes(t, []want{wp(spelling.E(), math.MaxInt, ct(0)), wp(spelling.E(), math.MaxInt-1, ct(0)), wp(spelling.E(), math.MaxInt-2, ct(0))}, got)
	})

	t.Run("the whole range is lazy", func(t *testing.T) {
		t.Parallel()
		n := mkName(t, spelling.E(), ct(0), nil)
		seq := n.MustRange(oct(math.MinInt), oct(math.MaxInt))
		var got []note.Note
		for x := range seq {
			got = append(got, x)
			if len(got) == 2 {
				break
			}
		}
		assertNotes(t, []want{wp(spelling.E(), math.MinInt, ct(0)), wp(spelling.E(), math.MinInt+1, ct(0))}, got)
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.D(), 1, ct(0), nil)
		for _, call := range []func() (iter.Seq[note.Note], error){
			func() (iter.Seq[note.Note], error) { return note.Note{}.Range(oct(1), oct(2)) },
			func() (iter.Seq[note.Note], error) { return n.Range(octave.Octave{}, oct(2)) },
			func() (iter.Seq[note.Note], error) { return n.Range(oct(1), octave.Octave{}) },
		} {
			seq, err := call()
			assertErr(t, err, note.ErrUnset)
			assert.Nil(t, seq)
		}
		assertPanicErr(t, func() { n.MustRange(octave.Octave{}, oct(2)) }, note.ErrUnset)
	})
}

// TestWalk covers OP-27 and K-47: the notes on the positions from n to "to",
// one step at a time, names by the rule, the cents and the source of n.
func TestWalk(t *testing.T) {
	t.Parallel()

	t.Run("K-47", func(t *testing.T) {
		t.Parallel()
		s := std()
		from := mk(t, spelling.C(), 4, ct(7), s)
		to := mk(t, spelling.E(), 4, ct(-3), s)
		fromBefore, toBefore := from, to
		got := collect(from.MustWalk(to, system.Sharps()))
		assertNotes(t, []want{
			wp(spelling.C(), 4, ct(7)).boundTo(s),
			wp(spelling.CSharp(), 4, ct(7)).boundTo(s),
			wp(spelling.D(), 4, ct(7)).boundTo(s),
			wp(spelling.DSharp(), 4, ct(7)).boundTo(s),
			wp(spelling.E(), 4, ct(7)).boundTo(s),
		}, got)
		assertSame(t, fromBefore, from)
		assertSame(t, toBefore, to)
	})

	t.Run("down, compared with system.Walk", func(t *testing.T) {
		t.Parallel()
		s := std()
		from := mk(t, spelling.D(), 5, ct(0), s)
		to := mk(t, spelling.A(), 4, ct(0), s)
		seq, err := from.Walk(to, system.Flats())
		require.NoError(t, err)
		var ws []want
		for p := range system.Traditional().MustWalk(pt(spelling.D(), 5), pt(spelling.A(), 4), system.Flats()) {
			ws = append(ws, wp(p.Spelling(), p.MustASPN(), ct(0)).boundTo(s))
		}
		assertNotes(t, ws, collect(seq))
	})

	t.Run("same position gives one note", func(t *testing.T) {
		t.Parallel()
		s := std()
		from := mk(t, spelling.BSharp(), 4, ct(0), s)
		to := mk(t, spelling.C(), 4, ct(0), s)
		got := collect(from.MustWalk(to, system.Sharps()))
		assertNotes(t, []want{wp(spelling.C(), 4, ct(0)).boundTo(s)}, got)
	})

	t.Run("different sources with equal systems", func(t *testing.T) {
		t.Parallel()
		a := std()
		from := mk(t, spelling.C(), 4, ct(0), a)
		to := mk(t, spelling.D(), 4, ct(0), stdAt(415))
		got := collect(from.MustWalk(to, system.Sharps()))
		require.Len(t, got, 3)
		assert.True(t, got[2].Source() == settings.Source(a), "the source of n")
	})

	t.Run("WalkIn unbound", func(t *testing.T) {
		t.Parallel()
		from := mk(t, spelling.C(), 4, ct(0), nil)
		to := mk(t, spelling.D(), 4, ct(0), nil)
		got := collect(from.MustWalkIn(set24(), to, system.Sharps()))
		var ws []want
		for p := range system.QuarterTones24().MustWalk(pt(spelling.C(), 4), pt(spelling.D(), 4), system.Sharps()) {
			ws = append(ws, wp(p.Spelling(), p.MustASPN(), ct(0)).free())
		}
		assertNotes(t, ws, got)
	})

	t.Run("WalkIn names inadmissible in the own set: first step and step N", func(t *testing.T) {
		t.Parallel()
		k25, s := setK25(), std()
		// classRule names class k as C with alteration k: from D4 (class 2) the
		// first name is C𝄪, inadmissible in K-25.
		first := mk(t, spelling.D(), 4, ct(0), k25)
		end := mk(t, spelling.E(), 4, ct(0), k25)
		_, err := first.WalkIn(s, end, classRule())
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)

		// From C4 the names are C, C♯, C𝄪: the third is inadmissible.
		third := mk(t, spelling.C(), 4, ct(0), k25)
		_, err = third.WalkIn(s, end, classRule())
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)

		// Over more than N steps the names repeat; the check stays correct.
		far := mk(t, spelling.C(), 6, ct(0), k25)
		_, err = third.WalkIn(s, far, classRule())
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)

		// lastRule gives an inadmissible name only to the last class of the
		// cycle (class 11 is A𝄪): the walk C4...B4 of N = 12 notes fails on
		// its 12th note, so every one of the first N names must be checked.
		lastRule := system.MustCustomSpelling(func(sys *system.System, c pitchclass.Class) spelling.Spelling {
			if c.Number() == 11 {
				return spelling.ADoubleSharp()
			}
			return sys.MustSpell(c, system.Sharps())
		})
		b4 := mk(t, spelling.B(), 4, ct(0), k25)
		_, err = third.WalkIn(s, b4, lastRule)
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)
		assertPanicErr(t, func() { third.MustWalkIn(s, b4, lastRule) }, note.ErrInadmissible)

		// Longer than N: the same class comes again and fails as well.
		_, err = third.WalkIn(s, far, lastRule)
		assertErr(t, err, note.ErrInadmissible, system.ErrInadmissible)

		// Eleven notes C4...A♯4 never reach class 11: the walk works.
		aSharp4 := mk(t, spelling.ASharp(), 4, ct(0), k25)
		got11 := collect(third.MustWalkIn(s, aSharp4, lastRule))
		assert.Len(t, got11, 11)
		for _, x := range got11 {
			assert.True(t, x.Source() == settings.Source(k25), "the own source is kept")
		}

		// The same walk of an unbound note works.
		free := mk(t, spelling.C(), 4, ct(0), nil)
		got := collect(free.MustWalkIn(s, mk(t, spelling.E(), 4, ct(0), nil), classRule()))
		assert.Len(t, got, 5)
	})

	t.Run("failures in order", func(t *testing.T) {
		t.Parallel()
		s, s19 := std(), set19()
		c4 := mk(t, spelling.C(), 4, ct(0), s)
		e4 := mk(t, spelling.E(), 4, ct(0), s)
		cases := []struct {
			name  string
			call  func() (iter.Seq[note.Note], error)
			err   error
			extra []error
		}{
			{"unset note", func() (iter.Seq[note.Note], error) { return note.Note{}.Walk(e4, system.Sharps()) }, note.ErrUnset, nil},
			{"unset target", func() (iter.Seq[note.Note], error) { return c4.Walk(note.Note{}, system.Sharps()) }, note.ErrUnset, nil},
			{"unset rule", func() (iter.Seq[note.Note], error) { return c4.Walk(e4, system.SpellingRule{}) }, note.ErrUnset, nil},
			{"unset rule before no octave", func() (iter.Seq[note.Note], error) {
				return c4.Walk(mkName(t, spelling.E(), ct(0), s), system.SpellingRule{})
			}, note.ErrUnset, nil},
			{"target without octave", func() (iter.Seq[note.Note], error) {
				return c4.Walk(mkName(t, spelling.E(), ct(0), s), system.Sharps())
			}, note.ErrNoOctave, nil},
			{"no octave before no source", func() (iter.Seq[note.Note], error) {
				return mkName(t, spelling.C(), ct(0), nil).Walk(e4, system.Sharps())
			}, note.ErrNoOctave, nil},
			{"unbound target", func() (iter.Seq[note.Note], error) {
				return c4.Walk(mk(t, spelling.E(), 4, ct(0), nil), system.Sharps())
			}, note.ErrNoSource, nil},
			{"unset settings of the target", func() (iter.Seq[note.Note], error) {
				return c4.Walk(onStale(spelling.E(), 4)(t), system.Sharps())
			}, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"inadmissible", func() (iter.Seq[note.Note], error) {
				x, _ := inadmissibleNote(t)
				return x.Walk(e4, system.Sharps())
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
			{"different systems", func() (iter.Seq[note.Note], error) {
				return c4.Walk(mk(t, spelling.E(), 4, ct(0), s19), system.Sharps())
			}, note.ErrDifferentSystems, nil},
			{"bad rule", func() (iter.Seq[note.Note], error) { return c4.Walk(e4, badRule()) }, nil, []error{system.ErrBadSpelling}},
			{"different systems before bad rule", func() (iter.Seq[note.Note], error) {
				return c4.Walk(mk(t, spelling.E(), 4, ct(0), s19), badRule())
			}, note.ErrDifferentSystems, nil},
			{"In nil source", func() (iter.Seq[note.Note], error) { return c4.WalkIn(nil, e4, system.Sharps()) }, note.ErrUnset, nil},
			{"In unset settings", func() (iter.Seq[note.Note], error) {
				return c4.WalkIn(&unsetSource{}, e4, system.Sharps())
			}, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"In name inadmissible in the explicit set", func() (iter.Seq[note.Note], error) {
				return mk(t, spelling.CDoubleSharp(), 4, ct(0), nil).WalkIn(setK25(), mk(t, spelling.E(), 4, ct(0), nil), system.Sharps())
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
		}
		for _, c := range cases {
			seq, err := c.call()
			if assertErr(t, err, c.err, c.extra...) {
				assert.Nil(t, seq, c.name)
			} else {
				t.Log(c.name)
			}
		}
		assertPanicErr(t, func() { c4.MustWalk(note.Note{}, system.Sharps()) }, note.ErrUnset)
		assertPanicErr(t, func() { c4.MustWalkIn(nil, e4, system.Sharps()) }, note.ErrUnset)
	})
}

// TestEnharmonicsReads covers KP-2 for OP-25: the own source once; the In
// form reads the explicit source once and the own source of a bound note once
// (6.3), or one read when they are one source.
func TestEnharmonicsReads(t *testing.T) {
	t.Parallel()

	t.Run("own source", func(t *testing.T) {
		t.Parallel()
		own := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), own)
		before := own.reads()
		_, err := n.Enharmonics(alteration.DoubleSharp())
		require.NoError(t, err)
		assert.EqualValues(t, 1, own.reads()-before)
	})

	t.Run("In with two sources", func(t *testing.T) {
		t.Parallel()
		own, explicit := newCounting(std()), newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), own)
		before := own.reads()
		_, err := n.EnharmonicsIn(explicit, alteration.DoubleSharp())
		require.NoError(t, err)
		assert.EqualValues(t, 1, explicit.reads())
		assert.EqualValues(t, 1, own.reads()-before)
	})

	t.Run("In with one source", func(t *testing.T) {
		t.Parallel()
		own := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), own)
		before := own.reads()
		_, err := n.EnharmonicsIn(own, alteration.DoubleSharp())
		require.NoError(t, err)
		assert.EqualValues(t, 1, own.reads()-before)
	})

	t.Run("In of an unbound note", func(t *testing.T) {
		t.Parallel()
		explicit := newCounting(std())
		n := mk(t, spelling.C(), 4, ct(0), nil)
		_, err := n.EnharmonicsIn(explicit, alteration.DoubleSharp())
		require.NoError(t, err)
		assert.EqualValues(t, 1, explicit.reads())
	})
}

// TestSequences covers the common rules of architecture 6.9: repeated
// iteration gives the same notes, break stops, a yielded note is a new value,
// the sources are read once at the call and never while iterating.
func TestSequences(t *testing.T) {
	t.Parallel()

	type maker struct {
		name string
		make func(t *testing.T, src settings.Source) iter.Seq[note.Note]
	}
	makers := []maker{
		{"Enharmonics", func(t *testing.T, src settings.Source) iter.Seq[note.Note] {
			return mk(t, spelling.CSharp(), 4, ct(1), src).MustEnharmonics(alteration.DoubleSharp())
		}},
		{"EnharmonicsIn", func(t *testing.T, src settings.Source) iter.Seq[note.Note] {
			return mk(t, spelling.CSharp(), 4, ct(1), src).MustEnharmonicsIn(std(), alteration.DoubleSharp())
		}},
		{"Range", func(t *testing.T, src settings.Source) iter.Seq[note.Note] {
			return mk(t, spelling.CSharp(), 4, ct(1), src).MustRange(oct(1), oct(6))
		}},
		{"Walk", func(t *testing.T, src settings.Source) iter.Seq[note.Note] {
			return mk(t, spelling.C(), 4, ct(1), src).MustWalk(mk(t, spelling.A(), 4, ct(0), src), system.Sharps())
		}},
		{"WalkIn", func(t *testing.T, src settings.Source) iter.Seq[note.Note] {
			return mk(t, spelling.C(), 4, ct(1), src).MustWalkIn(std(), mk(t, spelling.A(), 4, ct(0), src), system.Sharps())
		}},
	}
	for _, m := range makers {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			seq := m.make(t, src)
			readsAfterCall := src.reads()

			first := collect(seq)
			require.GreaterOrEqual(t, len(first), 3)
			second := collect(seq)
			assert.True(t, sameNotes(first, second), "repeated iteration gives the same notes")
			assert.Equal(t, readsAfterCall, src.reads(), "iteration does not read the source")

			var stopped []note.Note
			for x := range seq {
				stopped = append(stopped, x)
				if len(stopped) == 2 {
					break
				}
			}
			assert.True(t, sameNotes(first[:2], stopped), "break stops the sequence")

			for x := range seq {
				x.MustSetCents(ct(500))
				break
			}
			third := collect(seq)
			assert.True(t, sameNotes(first, third), "a yielded note is a new value")
		})
	}

	t.Run("a retuned source does not change a returned sequence", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		from := mk(t, spelling.C(), 4, ct(0), src)
		to := mk(t, spelling.E(), 4, ct(0), src)
		seq := from.MustWalk(to, system.Sharps())
		src.set(nil)
		assert.Len(t, collect(seq), 5, "the snapshot is taken at the call")
	})

	t.Run("K-47 the note and the arguments are unchanged", func(t *testing.T) {
		t.Parallel()
		s := std()
		n := mk(t, spelling.CSharp(), 4, ct(30), s)
		to := mk(t, spelling.E(), 5, ct(0), s)
		nb, tb := n, to
		_ = collect(n.MustEnharmonics(alteration.DoubleSharp()))
		_ = collect(n.MustRange(oct(3), oct(5)))
		_ = collect(n.MustWalk(to, system.Sharps()))
		assertSame(t, nb, n)
		assertSame(t, tb, to)
	})
}
