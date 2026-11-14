package note_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Comparisons and ordering (architecture 6.10; OP-39...OP-52, OP-57;
// K-2, K-4, K-5, K-21...K-24a, K-27, K-36, K-39...K-41, K-48, K-50, K-51).

// TestCompareWithoutSet covers OP-39, OP-40 (table 6.10.1) on notes with and
// without octave, of different sources and unset (K-19, K-21, K-28).
func TestCompareWithoutSet(t *testing.T) {
	t.Parallel()

	s, s19 := std(), set19()
	type res struct{ equal, record, spelling, letter, alteration, octave, cents bool }
	tests := []struct {
		name string
		a, b func(*testing.T) note.Note
		want res
	}{
		{"K-19 bound and unbound C4", onPitch(spelling.C(), 4, ct(0), s), onPitch(spelling.C(), 4, ct(0), nil),
			res{true, true, true, true, true, true, true}},
		{"different sources", onPitch(spelling.C(), 4, ct(0), s), onPitch(spelling.C(), 4, ct(0), s19),
			res{true, true, true, true, true, true, true}},
		{"K-21 C4 and C4 +5", onPitch(spelling.C(), 4, ct(0), s), onPitch(spelling.C(), 4, ct(5), s),
			res{false, true, true, true, true, true, false}},
		{"K-21 C4 +1200 and C5", onPitch(spelling.C(), 4, ct(1200), s), onPitch(spelling.C(), 5, ct(0), s),
			res{false, false, true, true, true, false, false}},
		{"C♯ and C♯4", onName(spelling.CSharp(), ct(0), s), onPitch(spelling.CSharp(), 4, ct(0), s),
			res{false, false, true, true, true, false, true}},
		{"C♯ and C♯ without octave", onName(spelling.CSharp(), ct(3), s), onName(spelling.CSharp(), ct(3), nil),
			res{true, true, true, true, true, false, true}},
		{"K-22 B♯3 and C4", onPitch(spelling.BSharp(), 4, ct(0), s), onPitch(spelling.C(), 4, ct(0), s),
			res{false, false, false, false, false, true, true}},
		{"C♯4 and D♯2", onPitch(spelling.CSharp(), 4, ct(0), s), onPitch(spelling.DSharp(), 2, ct(0), s),
			res{false, false, false, false, true, false, true}},
		{"C and C♯", onName(spelling.C(), ct(0), nil), onName(spelling.CSharp(), ct(0), nil),
			res{false, false, false, true, false, false, true}},
		{"fractional cents", onName(spelling.C(), cents.MustFromFraction(1, 3), nil), onName(spelling.C(), cents.MustFromFraction(2, 6), nil),
			res{true, true, true, true, true, false, true}},
		{"K-28 two unset notes", onUnset(), onUnset(), res{}},
		{"unset and C4", onUnset(), onPitch(spelling.C(), 4, ct(0), nil), res{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := tt.a(t), tt.b(t)
			ab, bb := a, b
			for _, pair := range [][2]note.Note{{a, b}, {b, a}} {
				x, y := pair[0], pair[1]
				assert.Equal(t, tt.want.equal, x.Equal(y), "Equal")
				assert.Equal(t, tt.want.record, x.SameRecord(y), "SameRecord")
				assert.Equal(t, tt.want.spelling, x.SameSpelling(y), "SameSpelling")
				assert.Equal(t, tt.want.letter, x.SameLetter(y), "SameLetter")
				assert.Equal(t, tt.want.alteration, x.SameAlteration(y), "SameAlteration")
				assert.Equal(t, tt.want.octave, x.SameOctave(y), "SameOctave")
				assert.Equal(t, tt.want.cents, x.SameCents(y), "SameCents")
				if tt.want.equal {
					assert.Zero(t, x.Compare(y), "Compare == 0 for equal notes")
				} else if x.IsValid() && y.IsValid() {
					assert.NotZero(t, x.Compare(y), "Compare != 0 for unequal set notes")
				}
			}
			assertSame(t, ab, a)
			assertSame(t, bb, b)
		})
	}

	t.Run("Equal is reflexive for set notes", func(t *testing.T) {
		t.Parallel()
		for _, n := range []note.Note{mk(t, spelling.C(), 4, ct(5), s), mkName(t, spelling.DFlat(), ct(0), nil)} {
			assert.True(t, n.Equal(n))
			assert.Zero(t, n.Compare(n))
		}
		var u note.Note
		assert.False(t, u.Equal(u), "R-5: two unset notes are not equal")
		assert.Zero(t, u.Compare(u), "two unset notes compare as 0")
	})
}

// TestSameSource covers OP-41 and K-36.
func TestSameSource(t *testing.T) {
	t.Parallel()

	user := newCounting(std())
	a := mk(t, spelling.C(), 4, ct(0), user)
	b := mkName(t, spelling.D(), ct(0), user)
	assert.True(t, a.SameSource(b), "K-36: one user source")

	x := mk(t, spelling.C(), 4, ct(0), settings.Standard())
	y := mk(t, spelling.C(), 4, ct(0), settings.Standard())
	assert.True(t, x.SameSource(y), "K-36: standard settings obtained separately")
	assert.False(t, x == y, "== tells the two standard pointers apart")

	assert.False(t, x.SameSource(a), "K-36: standard and a user source")
	assert.False(t, mk(t, spelling.C(), 4, ct(0), std()).SameSource(mk(t, spelling.C(), 4, ct(0), set19())))

	free := mk(t, spelling.C(), 4, ct(0), nil)
	assert.False(t, free.SameSource(free), "unbound notes have no source")
	assert.False(t, free.SameSource(a))
	var u note.Note
	assert.False(t, u.SameSource(u))
	before := user.reads()
	_ = a.SameSource(b)
	assert.Equal(t, before, user.reads())
}

// TestCompareOrder covers OP-42 and K-27: the agreed order.
func TestCompareOrder(t *testing.T) {
	t.Parallel()

	s := std()
	// K-27, in the expected order, with the unset note first.
	ordered := []note.Note{
		{},
		mkName(t, spelling.CSharp(), ct(0), s),
		mk(t, spelling.C(), 4, ct(0), s),
		mk(t, spelling.C(), 4, ct(10), nil),
		mk(t, spelling.BSharp(), 4, ct(0), s),
	}
	for i := range ordered {
		for j := range ordered {
			want := 0
			switch {
			case i < j:
				want = -1
			case i > j:
				want = 1
			}
			assert.Equal(t, want, ordered[i].Compare(ordered[j]), "%d vs %d", i, j)
		}
	}

	t.Run("names and pitches by their own order, then cents", func(t *testing.T) {
		t.Parallel()
		names := []spelling.Spelling{spelling.CFlat(), spelling.C(), spelling.CSharp(), spelling.DFlat(), spelling.D()}
		for _, x := range names {
			for _, y := range names {
				a, b := mkName(t, x, ct(0), nil), mkName(t, y, ct(0), nil)
				assert.Equal(t, x.Compare(y), a.Compare(b), "%v %v", x, y)
				pa, pb := mk(t, x, 3, ct(0), nil), mk(t, y, 4, ct(0), nil)
				assert.Equal(t, pt(x, 3).Compare(pt(y, 4)), pa.Compare(pb), "%v3 %v4", x, y)
			}
		}
		lo, hi := mk(t, spelling.C(), 4, ct(-1), nil), mk(t, spelling.C(), 4, ct(1), nil)
		assert.Equal(t, -1, lo.Compare(hi))
		assert.Equal(t, 1, hi.Compare(lo))
	})
}

// TestOctaveSubAndDistance covers OP-43.
func TestOctaveSubAndDistance(t *testing.T) {
	t.Parallel()

	c5, c3 := mk(t, spelling.C(), 5, ct(0), nil), mk(t, spelling.B(), 3, ct(9), std())
	assert.Equal(t, 2, c5.MustOctaveSub(c3))
	assert.Equal(t, -2, c3.MustOctaveSub(c5))
	assert.Equal(t, 2, c3.MustOctaveDistance(c5))
	assert.Equal(t, 2, c5.MustOctaveDistance(c3))

	noOct := mkName(t, spelling.C(), ct(0), nil)
	_, err := c5.OctaveSub(noOct)
	assertErr(t, err, note.ErrNoOctave)
	_, err = noOct.OctaveDistance(c5)
	assertErr(t, err, note.ErrNoOctave)
	_, err = c5.OctaveSub(note.Note{})
	assertErr(t, err, note.ErrUnset)
	_, err = note.Note{}.OctaveDistance(noOct)
	assertErr(t, err, note.ErrUnset, nil...)

	lo, hi := mk(t, spelling.C(), math.MinInt, ct(0), nil), mk(t, spelling.C(), math.MaxInt, ct(0), nil)
	_, err = hi.OctaveSub(lo)
	assertErr(t, err, note.ErrOutOfRange, octave.ErrOutOfRange)
	_, err = lo.OctaveDistance(hi)
	assertErr(t, err, note.ErrOutOfRange, octave.ErrOutOfRange)
	assertPanicErr(t, func() { hi.MustOctaveSub(lo) }, note.ErrOutOfRange)
	assertPanicErr(t, func() { lo.MustOctaveDistance(noOct) }, note.ErrNoOctave)
}

// TestPosition covers OP-44...OP-47 with their forms (6.10.2): K-22, K-23,
// K-24a; E-4 rules of Equivalent.
func TestPosition(t *testing.T) {
	t.Parallel()

	s, s19, s24 := std(), set19(), set24()
	tr := system.Traditional()

	t.Run("K-22 B♯3 and C4", func(t *testing.T) {
		t.Parallel()
		b, c := mk(t, spelling.BSharp(), 4, ct(0), s), mk(t, spelling.C(), 4, ct(50), s)
		assert.True(t, b.MustEquivalent(c))
		assert.Zero(t, b.MustDistance(c))
		assert.Zero(t, b.MustComparePosition(c))
		assert.False(t, b.Equal(c))
	})

	t.Run("results compared with system", func(t *testing.T) {
		t.Parallel()
		pitches := []struct {
			name spelling.Spelling
			aspn int
		}{{spelling.C(), 4}, {spelling.E(), 4}, {spelling.DSharp(), 4}, {spelling.FFlat(), 4}, {spelling.BSharp(), 3}, {spelling.A(), 2}}
		for _, x := range pitches {
			for _, y := range pitches {
				px, py := pt(x.name, x.aspn), pt(y.name, y.aspn)
				a, b := mk(t, x.name, x.aspn, ct(30), s), mk(t, y.name, y.aspn, ct(-30), std())
				fa, fb := mk(t, x.name, x.aspn, ct(0), nil), mk(t, y.name, y.aspn, ct(0), nil)

				assert.Equal(t, tr.MustEquivalent(px, py), a.MustEquivalent(b))
				assert.Equal(t, tr.MustEquivalent(px, py), fa.MustEquivalentIn(s, fb))
				assert.Equal(t, tr.MustEquivalent(px, py), fa.MustEquivalentInTuning(tuning.Standard(), fb))
				assert.Equal(t, tr.MustCompare(px, py), a.MustComparePosition(b))
				assert.Equal(t, tr.MustCompare(px, py), fa.MustComparePositionIn(s, fb))
				assert.Equal(t, tr.MustCompare(px, py), fa.MustComparePositionInTuning(tuning.Standard(), fb))
				assert.Equal(t, tr.MustDistance(px, py), a.MustDistance(b))
				assert.Equal(t, tr.MustDistance(px, py), fa.MustDistanceIn(s, fb))
				assert.Equal(t, tr.MustDistance(px, py), fa.MustDistanceInTuning(tuning.Standard(), fb))
				for _, z := range pitches {
					pz := pt(z.name, z.aspn)
					c := mk(t, z.name, z.aspn, ct(0), s)
					fc := mk(t, z.name, z.aspn, ct(0), nil)
					assert.Equal(t, tr.MustBetween(pz, px, py), c.MustBetween(a, b))
					assert.Equal(t, tr.MustBetween(pz, px, py), fc.MustBetweenIn(s, fa, fb))
					assert.Equal(t, tr.MustBetween(pz, px, py), fc.MustBetweenInTuning(tuning.Standard(), fa, fb))
				}
			}
		}
	})

	t.Run("in set 24 by an explicit tuning", func(t *testing.T) {
		t.Parallel()
		a, b := mk(t, cHalf(), 4, ct(0), nil), mk(t, sp(letter.D, -3, 2), 4, ct(0), nil)
		assert.True(t, a.MustEquivalentInTuning(tuning24(), b))
		assert.True(t, a.MustEquivalentIn(s24, b))
		assert.Equal(t, system.Step(1), mk(t, spelling.C(), 4, ct(0), nil).MustDistanceInTuning(tuning24(), a))
	})

	t.Run("without octave", func(t *testing.T) {
		t.Parallel()
		a, b := mkName(t, spelling.CSharp(), ct(0), s), mkName(t, spelling.DFlat(), ct(9), s)
		assert.True(t, a.MustEquivalent(b), "both without octave: one class")
		assert.False(t, a.MustEquivalent(mkName(t, spelling.D(), ct(0), s)))
		_, err := a.Equivalent(mk(t, spelling.DFlat(), 4, ct(0), s))
		assertErr(t, err, note.ErrNoOctave)
		_, err = a.ComparePosition(b)
		assertErr(t, err, note.ErrNoOctave)
		_, err = a.Distance(b)
		assertErr(t, err, note.ErrNoOctave)
		_, err = mk(t, spelling.C(), 4, ct(0), s).Between(a, mk(t, spelling.D(), 4, ct(0), s))
		assertErr(t, err, note.ErrNoOctave)
		assert.True(t, mkName(t, spelling.CSharp(), ct(0), nil).MustEquivalentInTuning(tuning.Standard(), mkName(t, spelling.DFlat(), ct(0), nil)))
	})

	t.Run("K-23 different systems", func(t *testing.T) {
		t.Parallel()
		a, b := mk(t, spelling.C(), 4, ct(0), s), mk(t, spelling.C(), 4, ct(0), s19)
		_, err := a.ComparePosition(b)
		assertErr(t, err, note.ErrDifferentSystems)
		_, err = a.Equivalent(b)
		assertErr(t, err, note.ErrDifferentSystems)
		_, err = a.Distance(b)
		assertErr(t, err, note.ErrDifferentSystems)
		_, err = a.Between(a, b)
		assertErr(t, err, note.ErrDifferentSystems)
		assert.Equal(t, 0, a.MustComparePositionIn(s19, b), "one explicit system for all notes")
	})

	t.Run("K-24a an alternating source is read once", func(t *testing.T) {
		t.Parallel()
		src := &alternatingSource{a: std(), b: set19()}
		c4 := mk(t, spelling.C(), 4, ct(0), src)
		d4 := mk(t, spelling.D(), 4, ct(0), src)
		for range 4 {
			got, err := c4.ComparePosition(d4)
			require.NoError(t, err)
			assert.Equal(t, -1, got)
			_, err = c4.Between(d4, c4)
			require.NoError(t, err)
			_, err = c4.Distance(d4)
			require.NoError(t, err)
		}
	})

	t.Run("distance out of range", func(t *testing.T) {
		t.Parallel()
		lo, hi := mk(t, spelling.C(), math.MinInt, ct(0), s), mk(t, spelling.C(), math.MaxInt, ct(0), s)
		_, err := lo.Distance(hi)
		assertErr(t, err, note.ErrOutOfRange, system.ErrOutOfRange)
		_, err = lo.DistanceInTuning(tuning.Standard(), hi)
		assertErr(t, err, note.ErrOutOfRange, system.ErrOutOfRange)
	})

	t.Run("failures in order", func(t *testing.T) {
		t.Parallel()
		c4, free := mk(t, spelling.C(), 4, ct(0), s), mk(t, spelling.C(), 4, ct(0), nil)
		noOctFree := mkName(t, spelling.C(), ct(0), nil)
		inadm, _ := inadmissibleNote(t)
		stale := onStale(spelling.D(), 4)(t)
		cases := []struct {
			name  string
			call  func() error
			err   error
			extra []error
		}{
			{"unset note", func() error { _, err := note.Note{}.ComparePosition(c4); return err }, note.ErrUnset, nil},
			{"unset argument", func() error { _, err := c4.Distance(note.Note{}); return err }, note.ErrUnset, nil},
			{"unset end", func() error { _, err := c4.Between(c4, note.Note{}); return err }, note.ErrUnset, nil},
			{"no octave before no source", func() error { _, err := c4.ComparePosition(noOctFree); return err }, note.ErrNoOctave, nil},
			{"no source", func() error { _, err := c4.ComparePosition(free); return err }, note.ErrNoSource, nil},
			{"no source of the receiver", func() error { _, err := free.Equivalent(c4); return err }, note.ErrNoSource, nil},
			{"unset settings", func() error { _, err := c4.Distance(stale); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"inadmissible", func() error { _, err := c4.Equivalent(inadm); return err }, note.ErrInadmissible, []error{system.ErrInadmissible}},
			{"inadmissible before different systems", func() error {
				_, err := inadm.Between(c4, mk(t, spelling.C(), 4, ct(0), s19))
				return err
			}, note.ErrInadmissible, nil},
			{"In nil source", func() error { _, err := free.ComparePositionIn(nil, free); return err }, note.ErrUnset, nil},
			{"In unset settings", func() error { _, err := free.EquivalentIn(&unsetSource{}, free); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"In inadmissible", func() error {
				_, err := free.DistanceIn(setK25(), mk(t, spelling.CDoubleSharp(), 4, ct(0), nil))
				return err
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
			{"InTuning nil", func() error { _, err := free.BetweenInTuning(nil, free, free); return err }, note.ErrUnset, nil},
			{"InTuning unset", func() error { _, err := free.EquivalentInTuning(&tuning.Tuning{}, free); return err }, note.ErrUnset, nil},
			{"InTuning inadmissible", func() error {
				_, err := free.ComparePositionInTuning(tuning.Standard(), mk(t, cHalf(), 4, ct(0), nil))
				return err
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
		}
		for _, c := range cases {
			if !assertErr(t, c.call(), c.err, c.extra...) {
				t.Log(c.name)
			}
		}
		assertPanicErr(t, func() { c4.MustEquivalent(free) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustEquivalentIn(nil, free) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustEquivalentInTuning(nil, free) }, note.ErrUnset)
		assertPanicErr(t, func() { c4.MustComparePosition(free) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustComparePositionIn(nil, free) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustComparePositionInTuning(nil, free) }, note.ErrUnset)
		assertPanicErr(t, func() { c4.MustBetween(free, c4) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustBetweenIn(nil, free, free) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustBetweenInTuning(nil, free, free) }, note.ErrUnset)
		assertPanicErr(t, func() { c4.MustDistance(free) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustDistanceIn(nil, free) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustDistanceInTuning(nil, free) }, note.ErrUnset)
	})
}

// TestSortedByPosition covers OP-48 and K-48: a new stable slice; the input
// and its notes are unchanged (R-16).
func TestSortedByPosition(t *testing.T) {
	t.Parallel()

	s := std()
	e4, c4, d4 := mk(t, spelling.E(), 4, ct(0), s), mk(t, spelling.C(), 4, ct(0), s), mk(t, spelling.D(), 4, ct(0), s)

	t.Run("K-48", func(t *testing.T) {
		t.Parallel()
		in := []note.Note{e4, c4, d4}
		got, err := note.SortedByPosition(in)
		require.NoError(t, err)
		assert.True(t, sameNotes([]note.Note{c4, d4, e4}, got))
		assert.True(t, sameNotes([]note.Note{e4, c4, d4}, in), "the input is unchanged")
		got[0] = note.Note{}
		assert.True(t, in[1] == c4, "the result is a new slice")
	})

	t.Run("stable for one position", func(t *testing.T) {
		t.Parallel()
		b := mk(t, spelling.BSharp(), 4, ct(0), s)
		c := mk(t, spelling.C(), 4, ct(9), s)
		d := mk(t, spelling.DDoubleFlat(), 4, ct(-9), s)
		assert.True(t, sameNotes([]note.Note{b, c, d, e4}, note.MustSortedByPosition([]note.Note{b, e4, c, d})))
		assert.True(t, sameNotes([]note.Note{d, c, b, e4}, note.MustSortedByPosition([]note.Note{e4, d, c, b})))
	})

	t.Run("explicit forms", func(t *testing.T) {
		t.Parallel()
		fe, fc := mk(t, spelling.E(), 4, ct(0), nil), mk(t, spelling.C(), 4, ct(0), nil)
		assert.True(t, sameNotes([]note.Note{fc, fe}, note.MustSortedByPositionIn(s, []note.Note{fe, fc})))
		assert.True(t, sameNotes([]note.Note{fc, fe}, note.MustSortedByPositionInTuning(tuning.Standard(), []note.Note{fe, fc})))
	})

	t.Run("empty and nil", func(t *testing.T) {
		t.Parallel()
		for _, in := range [][]note.Note{nil, {}} {
			got, err := note.SortedByPosition(in)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
			got, err = note.SortedByPositionIn(&unsetSource{}, in)
			require.NoError(t, err, "the explicit source is not read for an empty list")
			assert.NotNil(t, got)
			got, err = note.SortedByPositionInTuning(tuning.Standard(), in)
			require.NoError(t, err)
			assert.NotNil(t, got)
			_, err = note.SortedByPositionIn(nil, in)
			assertErr(t, err, note.ErrUnset)
			_, err = note.SortedByPositionInTuning(nil, in)
			assertErr(t, err, note.ErrUnset)
		}
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		_, err := note.SortedByPosition([]note.Note{c4, {}})
		assertErr(t, err, note.ErrUnset)
		_, err = note.SortedByPosition([]note.Note{c4, mkName(t, spelling.C(), ct(0), s)})
		assertErr(t, err, note.ErrNoOctave)
		_, err = note.SortedByPosition([]note.Note{c4, mk(t, spelling.C(), 4, ct(0), nil)})
		assertErr(t, err, note.ErrNoSource)
		_, err = note.SortedByPosition([]note.Note{c4, mk(t, spelling.C(), 4, ct(0), set19())})
		assertErr(t, err, note.ErrDifferentSystems)
		assertPanicErr(t, func() { note.MustSortedByPosition([]note.Note{c4, {}}) }, note.ErrUnset)
		assertPanicErr(t, func() { note.MustSortedByPositionIn(nil, nil) }, note.ErrUnset)
		assertPanicErr(t, func() { note.MustSortedByPositionInTuning(nil, nil) }, note.ErrUnset)
	})

	t.Run("K-24a list with an alternating source", func(t *testing.T) {
		t.Parallel()
		src := &alternatingSource{a: std(), b: set19()}
		in := []note.Note{
			mk(t, spelling.E(), 4, ct(0), src), mk(t, spelling.C(), 4, ct(0), src), mk(t, spelling.D(), 4, ct(0), src),
		}
		for range 3 {
			got, err := note.SortedByPosition(in)
			require.NoError(t, err, "no false E-6")
			assert.True(t, got[0].Equal(in[1]))
		}
	})
}

// TestSound covers OP-49...OP-51 with their forms (6.10.3): K-2, K-4, K-5,
// K-23, K-39, K-41; the exact path for equal tunings (R-14).
func TestSound(t *testing.T) {
	t.Parallel()

	s := std()

	t.Run("K-2 A4 +100 and A♯4", func(t *testing.T) {
		t.Parallel()
		a, b := mk(t, spelling.A(), 4, ct(100), s), mk(t, spelling.ASharp(), 4, ct(0), s)
		assert.True(t, a.MustSoundsSame(b))
		assert.Zero(t, a.MustCompareSound(b))
		assert.True(t, a.MustCentsTo(b).Equal(ct(0)))
	})

	t.Run("K-4 C4 +150 and C♯4", func(t *testing.T) {
		t.Parallel()
		a, b := mk(t, spelling.C(), 4, ct(150), s), mk(t, spelling.CSharp(), 4, ct(0), s)
		assert.Equal(t, 1, a.MustCompareSound(b), "higher by sound")
		assert.Equal(t, -1, a.MustComparePosition(b), "lower by position")
		got := b.MustCentsTo(a)
		assert.True(t, got.Equal(ct(50)), "got %v", got)
	})

	t.Run("K-5 A4 and C5 -10 exactly", func(t *testing.T) {
		t.Parallel()
		for _, set := range []*settings.Settings{s, stdAt(415)} {
			a, c := mk(t, spelling.A(), 4, ct(0), set), mk(t, spelling.C(), 5, ct(-10), set)
			got := a.MustCentsTo(c)
			assert.True(t, got.Equal(ct(290)), "got %v", got)
			assert.True(t, c.MustCentsTo(a).Equal(ct(-290)))
		}
	})

	t.Run("K-23 by sound in different systems", func(t *testing.T) {
		t.Parallel()
		a, b := mk(t, spelling.C(), 4, ct(0), s), mk(t, spelling.C(), 4, ct(0), set19())
		assert.Equal(t, -1, a.MustCompareSound(b))
		assert.Equal(t, 1, b.MustCompareSound(a))
		assert.False(t, a.MustSoundsSame(b))
		want := cents.MustBetween(tuning.Standard().MustFrequency(pt(spelling.C(), 4)), tuning19().MustFrequency(pt(spelling.C(), 4)))
		assert.True(t, a.MustCentsTo(b).Equal(want), "cents.Between of the frequencies")
	})

	t.Run("K-39 A4 at 440 and at 441", func(t *testing.T) {
		t.Parallel()
		a, b := mk(t, spelling.A(), 4, ct(0), s), mk(t, spelling.A(), 4, ct(0), stdAt(441))
		five := frequency.MustWithinCents(ct(5))
		assert.False(t, a.MustSoundsSame(b))
		assert.True(t, a.MustSoundsSameBy(b, five))
		assert.Zero(t, a.MustCompareSoundBy(b, five))
		assert.Equal(t, -1, a.MustCompareSoundBy(b, frequency.Strict()))
		assert.InDelta(t, 3.930158, a.MustCentsTo(b).Float64(), 1e-6)
	})

	t.Run("K-41 unset criterion", func(t *testing.T) {
		t.Parallel()
		a := mk(t, spelling.A(), 4, ct(0), s)
		assert.True(t, a.MustSoundsSame(a))
		_, err := a.SoundsSameBy(a, frequency.Criterion{})
		assertErr(t, err, note.ErrUnset, frequency.ErrUnset)
		_, err = a.CompareSoundBy(a, frequency.Criterion{})
		assertErr(t, err, note.ErrUnset, frequency.ErrUnset)
		_, err = a.SoundsSameByIn(s, a, frequency.Criterion{})
		assertErr(t, err, note.ErrUnset, frequency.ErrUnset)
		_, err = a.CompareSoundByInTuning(tuning.Standard(), a, frequency.Criterion{})
		assertErr(t, err, note.ErrUnset, frequency.ErrUnset)
		assertPanicErr(t, func() { a.MustSoundsSameBy(a, frequency.Criterion{}) }, note.ErrUnset, frequency.ErrUnset)
	})

	t.Run("exact path for equal tunings, frequency path for tolerances", func(t *testing.T) {
		t.Parallel()
		requireTinyMerges(t)
		hi, lo := mk(t, spelling.C(), 4, tiny(), s), mk(t, spelling.C(), 4, ct(0), std())
		zeroHz := frequency.MustWithinHertz(0)
		zeroRel := frequency.MustWithinRelative(0)
		allStrict := frequency.MustAllOf(frequency.Strict())

		assert.Equal(t, 1, hi.MustCompareSound(lo), "strict: exact heights")
		assert.Equal(t, 1, hi.MustCompareSoundBy(lo, frequency.Strict()), "strict kind: exact heights")
		assert.False(t, hi.MustSoundsSame(lo))
		assert.False(t, hi.MustSoundsSameBy(lo, frequency.Strict()))
		assert.True(t, hi.MustCentsTo(lo).Equal(tiny().Neg()), "exact cents")

		for _, c := range []frequency.Criterion{zeroHz, zeroRel, allStrict} {
			assert.Zero(t, hi.MustCompareSoundBy(lo, c), "%v: frequencies", c)
			assert.True(t, hi.MustSoundsSameBy(lo, c), "%v: frequencies", c)
		}

		// In and InTuning forms have one tuning: always the exact path.
		fh, fl := mk(t, spelling.C(), 4, tiny(), nil), mk(t, spelling.C(), 4, ct(0), nil)
		assert.Equal(t, 1, fh.MustCompareSoundIn(s, fl))
		assert.Equal(t, 1, fh.MustCompareSoundInTuning(tuning.Standard(), fl))
		assert.Equal(t, 1, fh.MustCompareSoundByIn(s, fl, frequency.Strict()))
		assert.Equal(t, 1, fh.MustCompareSoundByInTuning(tuning.Standard(), fl, frequency.Strict()))
		assert.Zero(t, fh.MustCompareSoundByIn(s, fl, zeroHz))
		assert.Zero(t, fh.MustCompareSoundByInTuning(tuning.Standard(), fl, zeroHz))
		assert.False(t, fh.MustSoundsSameIn(s, fl))
		assert.False(t, fh.MustSoundsSameInTuning(tuning.Standard(), fl))
		assert.True(t, fh.MustSoundsSameByIn(s, fl, zeroHz))
		assert.True(t, fh.MustSoundsSameByInTuning(tuning.Standard(), fl, zeroHz))
		assert.True(t, fh.MustCentsToIn(s, fl).Equal(tiny().Neg()))
		assert.True(t, fh.MustCentsToInTuning(tuning.Standard(), fl).Equal(tiny().Neg()))
	})

	t.Run("different tunings go by frequencies", func(t *testing.T) {
		t.Parallel()
		// A4 at 440 Hz with the reference A4 and with the reference C5 at the
		// frequency of C5: the same frequencies, different tunings.
		c5 := tuning.Standard().MustFrequency(pt(spelling.C(), 5))
		moved := std().MustWithTuning(tuning.Standard().MustWithReferenceAt(pt(spelling.C(), 5), c5))
		a := mk(t, spelling.A(), 4, ct(0), s)
		b := mk(t, spelling.A(), 4, ct(0), moved)
		fa, fb := s.Tuning().MustFrequency(pt(spelling.A(), 4)), moved.Tuning().MustFrequency(pt(spelling.A(), 4))
		want := frequency.MustCompare(fa, fb)
		assert.Equal(t, want, a.MustCompareSound(b))
		assert.Equal(t, want == 0, a.MustSoundsSame(b))
		assert.True(t, a.MustCentsTo(b).Equal(frequency.MustCentsBetween(fa, fb)))
	})

	t.Run("compared with the frequencies of tuning", func(t *testing.T) {
		t.Parallel()
		crit := frequency.MustWithinCents(ct(20))
		type in struct {
			set  *settings.Settings
			name spelling.Spelling
			c    int
		}
		inputs := []in{{s, spelling.C(), 0}, {set19(), spelling.C(), 15}, {stdAt(442), spelling.CSharp(), -90}, {s, spelling.DFlat(), 0}}
		for _, x := range inputs {
			for _, y := range inputs {
				fx := x.set.Tuning().MustShiftedFrequency(pt(x.name, 4), ct(x.c))
				fy := y.set.Tuning().MustShiftedFrequency(pt(y.name, 4), ct(y.c))
				nx, ny := mk(t, x.name, 4, ct(x.c), x.set), mk(t, y.name, 4, ct(y.c), y.set)
				assert.Equal(t, crit.MustCompare(fx, fy), nx.MustCompareSoundBy(ny, crit))
				assert.Equal(t, crit.MustNear(fx, fy), nx.MustSoundsSameBy(ny, crit))
			}
		}
	})

	t.Run("failures in order", func(t *testing.T) {
		t.Parallel()
		a, free := mk(t, spelling.A(), 4, ct(0), s), mk(t, spelling.A(), 4, ct(0), nil)
		noOct := mkName(t, spelling.A(), ct(0), nil)
		inadm, _ := inadmissibleNote(t)
		stale := onStale(spelling.A(), 4)(t)
		huge := mk(t, spelling.C(), 2000, ct(0), s)
		cases := []struct {
			name  string
			call  func() error
			err   error
			extra []error
		}{
			{"unset note", func() error { _, err := note.Note{}.CompareSound(a); return err }, note.ErrUnset, nil},
			{"unset criterion before no octave", func() error { _, err := noOct.SoundsSameBy(a, frequency.Criterion{}); return err }, note.ErrUnset, []error{frequency.ErrUnset}},
			{"K-19 no source", func() error { _, err := free.CompareSound(a); return err }, note.ErrNoSource, nil},
			{"no octave before no source", func() error { _, err := free.SoundsSame(noOct); return err }, note.ErrNoOctave, nil},
			{"unset settings", func() error { _, err := a.CentsTo(stale); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"inadmissible", func() error { _, err := a.CompareSound(inadm); return err }, note.ErrInadmissible, []error{system.ErrInadmissible}},
			{"frequency out of range", func() error { _, err := huge.CompareSound(mk(t, spelling.C(), 4, ct(0), set19())); return err }, note.ErrOutOfRange, nil},
			{"In nil source", func() error { _, err := free.CentsToIn(nil, free); return err }, note.ErrUnset, nil},
			{"In unset settings", func() error { _, err := free.SoundsSameIn(&unsetSource{}, free); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"InTuning nil", func() error { _, err := free.CompareSoundInTuning(nil, free); return err }, note.ErrUnset, nil},
			{"InTuning unset", func() error { _, err := free.CentsToInTuning(&tuning.Tuning{}, free); return err }, note.ErrUnset, nil},
			{"InTuning inadmissible", func() error {
				_, err := free.SoundsSameByInTuning(tuning.Standard(), mk(t, cHalf(), 4, ct(0), nil), frequency.Strict())
				return err
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
		}
		for _, c := range cases {
			if !assertErr(t, c.call(), c.err, c.extra...) {
				t.Log(c.name)
			}
		}
		assertPanicErr(t, func() { free.MustCompareSound(a) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustCompareSoundIn(nil, a) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustCompareSoundInTuning(nil, a) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustCompareSoundBy(a, frequency.Strict()) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustCompareSoundByIn(nil, a, frequency.Strict()) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustCompareSoundByInTuning(nil, a, frequency.Strict()) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustSoundsSame(a) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustSoundsSameIn(nil, a) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustSoundsSameInTuning(nil, a) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustSoundsSameByIn(nil, a, frequency.Strict()) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustSoundsSameByInTuning(nil, a, frequency.Strict()) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustCentsTo(a) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustCentsToIn(nil, a) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustCentsToInTuning(nil, a) }, note.ErrUnset)
	})
}

// TestSortedBySound covers OP-52 and K-50: the exact key when all tunings are
// equal, the frequency otherwise (KR-12); stable; the input is unchanged.
func TestSortedBySound(t *testing.T) {
	t.Parallel()

	s := std()

	t.Run("K-50", func(t *testing.T) {
		t.Parallel()
		cs := mk(t, spelling.CSharp(), 4, ct(0), s)
		c19 := mk(t, spelling.C(), 4, ct(0), set19())
		c4 := mk(t, spelling.C(), 4, ct(0), s)
		in := []note.Note{cs, c19, c4}
		got, err := note.SortedBySound(in)
		require.NoError(t, err)
		assert.True(t, sameNotes([]note.Note{c4, c19, cs}, got))
		assert.True(t, sameNotes([]note.Note{cs, c19, c4}, in))
		assert.InDelta(t, 261.6256, got[0].MustFrequency(), 0.00005)
		assert.InDelta(t, 264.0226, got[1].MustFrequency(), 0.00005)
		assert.InDelta(t, 277.1826, got[2].MustFrequency(), 0.00005)
	})

	t.Run("equal tunings: exact key", func(t *testing.T) {
		t.Parallel()
		requireTinyMerges(t)
		hi, lo := mk(t, spelling.C(), 4, tiny(), s), mk(t, spelling.C(), 4, ct(0), std())
		assert.True(t, sameNotes([]note.Note{lo, hi}, note.MustSortedBySound([]note.Note{hi, lo})))
		fh, fl := mk(t, spelling.C(), 4, tiny(), nil), mk(t, spelling.C(), 4, ct(0), nil)
		assert.True(t, sameNotes([]note.Note{fl, fh}, note.MustSortedBySoundIn(s, []note.Note{fh, fl})))
		assert.True(t, sameNotes([]note.Note{fl, fh}, note.MustSortedBySoundInTuning(tuning.Standard(), []note.Note{fh, fl})))
	})

	t.Run("mixed tunings: frequency key, stable on ties", func(t *testing.T) {
		t.Parallel()
		requireTinyMerges(t)
		hi, lo := mk(t, spelling.C(), 4, tiny(), s), mk(t, spelling.C(), 4, ct(0), s)
		e19 := mk(t, spelling.E(), 4, ct(0), set19())
		got := note.MustSortedBySound([]note.Note{e19, hi, lo})
		assert.True(t, sameNotes([]note.Note{hi, lo, e19}, got), "equal frequencies keep their order")
	})

	t.Run("enharmonic notes keep their order", func(t *testing.T) {
		t.Parallel()
		b := mk(t, spelling.BSharp(), 4, ct(0), s)
		c := mk(t, spelling.C(), 4, ct(0), s)
		assert.True(t, sameNotes([]note.Note{c, b}, note.MustSortedBySound([]note.Note{c, b})))
		assert.True(t, sameNotes([]note.Note{b, c}, note.MustSortedBySound([]note.Note{b, c})))
	})

	t.Run("empty and nil", func(t *testing.T) {
		t.Parallel()
		for _, in := range [][]note.Note{nil, {}} {
			got, err := note.SortedBySound(in)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
			got, err = note.SortedBySoundIn(&unsetSource{}, in)
			require.NoError(t, err)
			assert.NotNil(t, got)
			got, err = note.SortedBySoundInTuning(tuning.Standard(), in)
			require.NoError(t, err)
			assert.NotNil(t, got)
			_, err = note.SortedBySoundIn(nil, in)
			assertErr(t, err, note.ErrUnset)
			_, err = note.SortedBySoundInTuning(&tuning.Tuning{}, in)
			assertErr(t, err, note.ErrUnset)
		}
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		c4 := mk(t, spelling.C(), 4, ct(0), s)
		_, err := note.SortedBySound([]note.Note{c4, {}})
		assertErr(t, err, note.ErrUnset)
		_, err = note.SortedBySound([]note.Note{mk(t, spelling.C(), 4, ct(0), nil), mkName(t, spelling.C(), ct(0), s)})
		assertErr(t, err, note.ErrNoOctave, nil...)
		_, err = note.SortedBySound([]note.Note{c4, mk(t, spelling.C(), 4, ct(0), nil)})
		assertErr(t, err, note.ErrNoSource)
		_, err = note.SortedBySound([]note.Note{c4, mk(t, spelling.C(), 2000, ct(0), set19())})
		assertErr(t, err, note.ErrOutOfRange)
		assertPanicErr(t, func() { note.MustSortedBySound([]note.Note{{}}) }, note.ErrUnset)
		assertPanicErr(t, func() { note.MustSortedBySoundIn(nil, nil) }, note.ErrUnset)
		assertPanicErr(t, func() { note.MustSortedBySoundInTuning(nil, nil) }, note.ErrUnset)
	})
}

// TestOffset covers OP-57: K-40, K-51; the order of failures with the invalid
// frequency before E-2 and before E-3 of the explicit set.
func TestOffset(t *testing.T) {
	t.Parallel()

	s := std()

	t.Run("K-40", func(t *testing.T) {
		t.Parallel()
		a := mk(t, spelling.A(), 4, ct(0), s)
		five := frequency.MustWithinCents(ct(5))

		where, dev, err := a.Offset(441)
		require.NoError(t, err)
		assert.Equal(t, 1, where)
		assert.InDelta(t, 3.930158, dev.Float64(), 1e-6)
		assert.True(t, dev.Equal(frequency.MustCentsBetween(440, 441)))

		where, dev = a.MustOffsetBy(441, five)
		assert.Equal(t, 0, where)
		assert.InDelta(t, 3.930158, dev.Float64(), 1e-6)

		where, dev = a.MustOffsetBy(445, five)
		assert.Equal(t, 1, where)
		assert.InDelta(t, 19.562175, dev.Float64(), 1e-6)

		where, dev = a.MustOffset(430)
		assert.Equal(t, -1, where)
		assert.True(t, dev.Equal(frequency.MustCentsBetween(440, 430)))
	})

	t.Run("K-51 unbound", func(t *testing.T) {
		t.Parallel()
		a := mk(t, spelling.A(), 4, ct(0), nil)
		_, _, err := a.Offset(441)
		assertErr(t, err, note.ErrNoSource)
		where, dev := a.MustOffsetIn(s, 441)
		assert.Equal(t, 1, where)
		assert.InDelta(t, 3.930158, dev.Float64(), 1e-6)
		where, dev = a.MustOffsetInTuning(tuning.Standard(), 441)
		assert.Equal(t, 1, where)
		assert.InDelta(t, 3.930158, dev.Float64(), 1e-6)
		where, _ = a.MustOffsetByIn(s, 441, frequency.MustWithinCents(ct(5)))
		assert.Equal(t, 0, where)
		where, _ = a.MustOffsetByInTuning(tuning.Standard(), 441, frequency.MustWithinCents(ct(5)))
		assert.Equal(t, 0, where)
		assert.False(t, a.IsBound())
	})

	t.Run("compared with frequency", func(t *testing.T) {
		t.Parallel()
		n := mk(t, spelling.CSharp(), 5, ct(-17), set19())
		f := tuning19().MustShiftedFrequency(pt(spelling.CSharp(), 5), ct(-17))
		for _, x := range []float64{f, f * 1.001, f / 1.001, 1, 20000} {
			where, dev := n.MustOffset(x)
			assert.Equal(t, frequency.MustCompare(x, f), where)
			assert.True(t, dev.Equal(frequency.MustCentsBetween(f, x)))
		}
	})

	t.Run("In does not read the own source", func(t *testing.T) {
		t.Parallel()
		own := newCounting(std())
		a := mk(t, spelling.A(), 4, ct(0), own)
		before := own.reads()
		where, _ := a.MustOffsetIn(stdAt(441), 441)
		assert.Equal(t, before, own.reads())
		f441 := stdAt(441).Tuning().MustFrequency(pt(spelling.A(), 4))
		assert.Equal(t, frequency.MustCompare(441, f441), where, "the explicit set gives A4 at 441 Hz")
	})

	t.Run("failures in order", func(t *testing.T) {
		t.Parallel()
		a, free := mk(t, spelling.A(), 4, ct(0), s), mk(t, spelling.A(), 4, ct(0), nil)
		noOct := mkName(t, spelling.A(), ct(0), s)
		stale := onStale(spelling.A(), 4)(t)
		inadm, _ := inadmissibleNote(t)
		invalid := []error{frequency.ErrInvalidFrequency, cents.ErrInvalidArgument}
		cases := []struct {
			name  string
			call  func() error
			err   error
			extra []error
		}{
			{"unset note", func() error { _, _, err := note.Note{}.Offset(440); return err }, note.ErrUnset, nil},
			{"unset criterion", func() error { _, _, err := a.OffsetBy(440, frequency.Criterion{}); return err }, note.ErrUnset, []error{frequency.ErrUnset}},
			{"unset criterion before no octave", func() error { _, _, err := noOct.OffsetBy(-1, frequency.Criterion{}); return err }, note.ErrUnset, nil},
			{"no octave before invalid frequency", func() error { _, _, err := noOct.Offset(-1); return err }, note.ErrNoOctave, nil},
			{"invalid frequency", func() error { _, _, err := a.Offset(math.NaN()); return err }, nil, invalid},
			{"zero frequency", func() error { _, _, err := a.Offset(0); return err }, nil, invalid},
			{"invalid frequency before no source", func() error { _, _, err := free.Offset(-1); return err }, nil, invalid},
			{"invalid frequency before unset settings", func() error { _, _, err := stale.Offset(math.Inf(1)); return err }, nil, invalid},
			{"no source", func() error { _, _, err := free.Offset(440); return err }, note.ErrNoSource, nil},
			{"unset settings", func() error { _, _, err := stale.Offset(440); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"inadmissible", func() error { _, _, err := inadm.Offset(440); return err }, note.ErrInadmissible, []error{system.ErrInadmissible}},
			{"frequency of the note out of range", func() error {
				_, _, err := mk(t, spelling.C(), 2000, ct(0), s).Offset(440)
				return err
			}, note.ErrOutOfRange, nil},
			{"In nil source", func() error { _, _, err := free.OffsetIn(nil, 440); return err }, note.ErrUnset, nil},
			{"In invalid frequency before unset settings", func() error { _, _, err := free.OffsetIn(&unsetSource{}, -1); return err }, nil, invalid},
			{"In unset settings", func() error { _, _, err := free.OffsetByIn(&unsetSource{}, 440, frequency.Strict()); return err }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"In unset criterion", func() error { _, _, err := free.OffsetByIn(s, 440, frequency.Criterion{}); return err }, note.ErrUnset, []error{frequency.ErrUnset}},
			{"InTuning nil", func() error { _, _, err := free.OffsetInTuning(nil, 440); return err }, note.ErrUnset, nil},
			{"InTuning invalid frequency", func() error {
				_, _, err := free.OffsetByInTuning(tuning.Standard(), -1, frequency.Strict())
				return err
			}, nil, invalid},
			{"InTuning inadmissible", func() error {
				_, _, err := mk(t, cHalf(), 4, ct(0), nil).OffsetInTuning(tuning.Standard(), 440)
				return err
			}, note.ErrInadmissible, []error{system.ErrInadmissible}},
		}
		for _, c := range cases {
			if !assertErr(t, c.call(), c.err, c.extra...) {
				t.Log(c.name)
			}
		}
		assertPanicErr(t, func() { free.MustOffset(440) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustOffsetBy(440, frequency.Strict()) }, note.ErrNoSource)
		assertPanicErr(t, func() { free.MustOffsetIn(nil, 440) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustOffsetByIn(nil, 440, frequency.Strict()) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustOffsetInTuning(nil, 440) }, note.ErrUnset)
		assertPanicErr(t, func() { free.MustOffsetByInTuning(nil, 440, frequency.Strict()) }, note.ErrUnset)
	})
}
