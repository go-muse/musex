package interval_test

import (
	"math"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Task 18.2: numeric shifts retain their system, rather than a spelling.
func TestChromaticValue(t *testing.T) {
	t.Parallel()
	var zero interval.Chromatic
	assert.False(t, zero.IsValid())
	assert.Nil(t, zero.System())
	assert.Zero(t, zero.Steps())
	assert.Equal(t, "Chromatic(unset)", zero.String())
	for _, tc := range []struct {
		steps int
		text  string
	}{{0, "0"}, {3, "+3"}, {-18, "-18"}, {math.MinInt, "-9223372036854775808"}, {math.MaxInt, "+9223372036854775807"}} {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			sys := trad()
			c, err := interval.NewChromatic(sys, tc.steps)
			require.NoError(t, err)
			assert.True(t, c.IsValid())
			assert.Same(t, sys, c.System())
			assert.Equal(t, tc.steps, c.Steps())
			assert.Equal(t, tc.text, c.String())
			copy := c
			assert.Equal(t, c, copy)
			assert.True(t, c.MustSameShift(copy))
			assert.Equal(t, c, interval.MustNewChromatic(sys, tc.steps))
		})
	}
	for _, sys := range []*system.System{nil, new(system.System)} {
		c, err := interval.NewChromatic(sys, 0)
		assert.ErrorIs(t, err, interval.ErrUnset)
		assert.Equal(t, zero, c)
	}
	c := interval.MustNewChromatic(trad(), 3)
	equal := system.MustNew(trad().Definition())
	assert.NotSame(t, c.System(), equal)
	assert.True(t, c.MustSameShift(interval.MustNewChromatic(equal, 3)))
	assert.False(t, c.MustSameShift(interval.MustNewChromatic(equal, -3)))
	for _, other := range []interval.Chromatic{zero, interval.MustNewChromatic(qt24(), 3), interval.MustNewChromatic(irregular(), 3)} {
		ok, err := c.SameShift(other)
		assert.False(t, ok)
		if other.IsValid() {
			assert.ErrorIs(t, err, interval.ErrDifferentSystems)
		} else {
			assert.ErrorIs(t, err, interval.ErrUnset)
		}
	}
	_, err := zero.SameShift(c)
	assert.ErrorIs(t, err, interval.ErrUnset)
	assert.ErrorIs(t, panicError(t, func() { interval.MustNewChromatic(nil, 0) }), interval.ErrUnset)
	assert.ErrorIs(t, panicError(t, func() { zero.MustSameShift(c) }), interval.ErrUnset)
}

type chromaticSource struct {
	set   *settings.Settings
	reads atomic.Int64
}

func (s *chromaticSource) Settings() *settings.Settings { s.reads.Add(1); return s.set }

func TestChromaticMeasurement(t *testing.T) {
	t.Parallel()
	sys := trad()
	for _, tc := range []struct {
		name     string
		from, to pitch.Pitch
		steps    int
	}{
		{"compound", wp(sys, spelling.C(), 4), wp(sys, spelling.FSharp(), 5), 18},
		{"descending", wp(sys, spelling.FSharp(), 5), wp(sys, spelling.C(), 4), -18},
		{"written zero", wp(sys, spelling.BSharp(), 3), wp(sys, spelling.C(), 4), 0},
		{"minor third", wp(sys, spelling.C(), 4), wp(sys, spelling.EFlat(), 4), 3},
		{"augmented second", wp(sys, spelling.C(), 4), wp(sys, spelling.DSharp(), 4), 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			src := &chromaticSource{set: std()}
			from, to := noteAt(t, tc.from, 31, src), noteAt(t, tc.to, -17, src)
			src.reads.Store(0)
			c, err := interval.ChromaticBetweenNotes(from, to)
			require.NoError(t, err)
			assert.Equal(t, int64(1), src.reads.Load())
			assert.Equal(t, tc.steps, c.Steps())
			assert.True(t, c.System().Equal(sys))
			assert.True(t, c.MustSameShift(interval.MustChromaticBetweenNotesIn(sys, from, to)))
			assert.True(t, c.MustSameShift(interval.MustChromaticBetweenNotes(from, to)))
			src.set = set24()
			assert.True(t, c.System().Equal(sys), "measurement keeps its system snapshot")
		})
	}
	first, second := std(), std()
	from, to := noteAt(t, wp(sys, spelling.C(), 4), 0, first), noteAt(t, wp(sys, spelling.EFlat(), 4), 0, second)
	c, err := interval.ChromaticBetweenNotes(from, to)
	require.NoError(t, err)
	assert.Same(t, first.Tuning().System(), c.System())
	// Explicit systems ignore even sources that became unavailable.
	src := newSwitch(std())
	a, b := noteAt(t, wp(sys, spelling.C(), 4), 0, src), noteAt(t, wp(sys, spelling.E(), 4), 0, src)
	src.set(nil)
	c, err = interval.ChromaticBetweenNotesIn(qt24(), a, b)
	require.NoError(t, err)
	assert.Equal(t, 8, c.Steps())
}

func TestChromaticMeasurementErrors(t *testing.T) {
	t.Parallel()
	sys := trad()
	bound := noteAt(t, wp(sys, spelling.C(), 4), 0, std())
	free := noteAt(t, wp(sys, spelling.C(), 4), 0, nil)
	noOctave := noteName(t, spelling.C(), 0, nil)
	badSource := newSwitch(std())
	bad := noteAt(t, wp(sys, spelling.C(), 4), 0, badSource)
	badSource.set(nil)
	narrowed := newSwitch(set24())
	half := noteAt(t, wp(qt24(), eHalfFlat(), 4), 0, narrowed)
	narrowed.set(std())
	for _, tc := range []struct {
		name string
		a, b note.Note
		want error
	}{
		{"first unset", note.Note{}, noOctave, interval.ErrUnset},
		{"second unset", bound, note.Note{}, interval.ErrUnset},
		{"missing first octave", noOctave, bound, interval.ErrNoOctave},
		{"missing second octave before system", free, noOctave, interval.ErrNoOctave},
		{"unbound first", free, bound, interval.ErrNoSystem},
		{"unbound second", bound, free, interval.ErrNoSystem},
		{"different systems", bound, noteAt(t, wp(qt24(), spelling.C(), 4), 0, set24()), interval.ErrDifferentSystems},
		{"settings unavailable", bad, bound, note.ErrUnsetSettings},
		{"incompatible name", bound, half, system.ErrInadmissible},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := interval.ChromaticBetweenNotes(tc.a, tc.b)
			assert.ErrorIs(t, err, tc.want)
			assert.Equal(t, interval.Chromatic{}, c)
		})
	}
	c, err := interval.ChromaticBetweenNotesIn(nil, note.Note{}, note.Note{})
	assert.ErrorIs(t, err, interval.ErrUnset)
	assert.Equal(t, interval.Chromatic{}, c)
	_, err = interval.ChromaticBetweenNotesIn(sys, bound, noOctave)
	assert.ErrorIs(t, err, interval.ErrNoOctave)
	_, err = interval.ChromaticBetweenNotesIn(sys, bound, half)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	assert.ErrorIs(t, panicError(t, func() { interval.MustChromaticBetweenNotes(note.Note{}, bound) }), interval.ErrUnset)
	assert.ErrorIs(t, panicError(t, func() { interval.MustChromaticBetweenNotesIn(nil, bound, bound) }), interval.ErrUnset)
}

func TestNamedChromatic(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		i     interval.Interval
		sys   *system.System
		start spelling.Spelling
		want  int
	}{
		{"traditional minor third", iv("m3"), trad(), noStart, 3},
		{"quarter tone minor third", iv("m3"), qt24(), noStart, 6},
		{"neutral third", iv("n3"), qt24(), spelling.C(), 7},
		{"irregular up", iv("M2"), irregular(), spelling.D(), 1},
		{"irregular down", iv("↓M2"), irregular(), spelling.D(), -3},
		{"written up negative", iv("dd2"), trad(), spelling.C(), -1},
		{"compound", iv("↓M10"), trad(), spelling.C(), -16},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := tc.i.Chromatic(tc.sys, tc.start)
			require.NoError(t, err)
			assert.Equal(t, tc.want, c.Steps())
			assert.Same(t, tc.sys, c.System())
			assert.Equal(t, c, tc.i.MustChromatic(tc.sys, tc.start))
		})
	}
	for _, tc := range []struct {
		i     interval.Interval
		sys   *system.System
		start spelling.Spelling
		want  error
	}{
		{interval.Interval{}, trad(), noStart, interval.ErrUnset},
		{iv("M3"), nil, noStart, interval.ErrUnset},
		{iv("M2"), irregular(), noStart, interval.ErrStartRequired},
		{iv("n3"), trad(), noStart, interval.ErrFractionalSteps},
		{iv("n3"), trad(), spelling.C(), system.ErrInadmissible},
	} {
		c, err := tc.i.Chromatic(tc.sys, tc.start)
		assert.ErrorIs(t, err, tc.want)
		assert.Equal(t, interval.Chromatic{}, c)
	}
	assert.ErrorIs(t, panicError(t, func() { iv("M3").MustChromatic(nil, noStart) }), interval.ErrUnset)
}

func TestChromaticTranspose(t *testing.T) {
	t.Parallel()
	sys := trad()
	src := std()
	for _, tc := range []struct {
		name       string
		steps      int
		from, want spelling.Spelling
		rule       system.SpellingRule
	}{
		{"minor third", 3, spelling.D(), spelling.F(), system.Flats()},
		{"zero respells", 0, spelling.DSharp(), spelling.EFlat(), system.Flats()},
		{"descending", -3, spelling.F(), spelling.D(), system.Sharps()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, source := range []settings.Source{nil, src} {
				from := noteAt(t, wp(sys, tc.from, 4), 23, source)
				before := from
				c := interval.MustNewChromatic(sys, tc.steps)
				got, err := c.TransposeNote(from, tc.rule)
				require.NoError(t, err)
				want := noteAt(t, wp(sys, tc.want, 4), 23, source)
				sameNote(t, want, got, from)
				unchanged(t, before, from)
				assert.True(t, got.Equal(c.MustTransposeNote(from, tc.rule)))
			}
		})
	}
	valid := interval.MustNewChromatic(sys, 1)
	for _, tc := range []struct {
		c    interval.Chromatic
		n    note.Note
		rule system.SpellingRule
		want error
	}{
		{interval.Chromatic{}, noteAt(t, wp(sys, spelling.C(), 4), 0, nil), system.Sharps(), interval.ErrUnset},
		{valid, note.Note{}, system.Sharps(), interval.ErrUnset},
		{valid, noteName(t, spelling.C(), 0, nil), system.Sharps(), interval.ErrNoOctave},
		{valid, noteAt(t, wp(qt24(), spelling.C(), 4), 0, set24()), system.Sharps(), interval.ErrDifferentSystems},
	} {
		got, err := tc.c.TransposeNote(tc.n, tc.rule)
		assert.ErrorIs(t, err, tc.want)
		assert.Equal(t, note.Note{}, got)
	}
	assert.ErrorIs(t, panicError(t, func() { valid.MustTransposeNote(note.Note{}, system.Sharps()) }), interval.ErrUnset)
}

func TestChromaticDistinctSourcesAndSystemParameters(t *testing.T) {
	t.Parallel()
	left, right := &chromaticSource{set: std()}, &chromaticSource{set: std()}
	a := noteAt(t, wp(trad(), spelling.C(), 4), 0, left)
	b := noteAt(t, wp(trad(), spelling.EFlat(), 4), 0, right)
	left.reads.Store(0)
	right.reads.Store(0)
	got, err := interval.ChromaticBetweenNotes(a, b)
	require.NoError(t, err)
	assert.Equal(t, int64(1), left.reads.Load())
	assert.Equal(t, int64(1), right.reads.Load())
	assert.Equal(t, 3, got.Steps())
	// Equal N and equal name alone are insufficient.
	def := trad().Definition()
	def.Name = "same label"
	first := system.MustNew(def)
	def.Offsets = [7]int{0, 3, 4, 5, 7, 9, 11}
	second := system.MustNew(def)
	ok, err := interval.MustNewChromatic(first, 3).SameShift(interval.MustNewChromatic(second, 3))
	assert.ErrorIs(t, err, interval.ErrDifferentSystems)
	assert.False(t, ok)
	// Different names alone do not make unequal musical systems.
	def = first.Definition()
	def.Name = "another label"
	equal := system.MustNew(def)
	assert.True(t, interval.MustNewChromatic(first, 3).MustSameShift(interval.MustNewChromatic(equal, 3)))
}

func TestChromaticOverflowAndSpellingErrors(t *testing.T) {
	t.Parallel()
	// The pitch registers are representable; their positional difference is not.
	low := noteAt(t, pt(spelling.C(), -math.MaxInt/24), 0, nil)
	high := noteAt(t, pt(spelling.C(), math.MaxInt/24), 0, nil)
	c, err := interval.ChromaticBetweenNotesIn(trad(), low, high)
	require.NoError(t, err)
	assert.Equal(t, 24*(math.MaxInt/24), c.Steps())
	// A wider pair has individually valid registers and no representable shift.
	low = noteAt(t, pt(spelling.C(), -math.MaxInt/12), 0, nil)
	high = noteAt(t, pt(spelling.C(), math.MaxInt/12), 0, nil)
	c, err = interval.ChromaticBetweenNotesIn(trad(), low, high)
	assert.ErrorIs(t, err, interval.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)
	assert.Equal(t, interval.Chromatic{}, c)
	from := noteAt(t, wp(trad(), spelling.C(), 4), 0, nil)
	c = interval.MustNewChromatic(trad(), 1)
	moved, err := c.TransposeNote(from, system.SpellingRule{})
	assert.ErrorIs(t, err, system.ErrInvalid)
	assert.Equal(t, note.Note{}, moved)
}

func TestChromaticTransposeSourceFailure(t *testing.T) {
	t.Parallel()
	source := newSwitch(std())
	from := noteAt(t, wp(trad(), spelling.C(), 4), 0, source)
	source.set(nil)
	result, err := interval.MustNewChromatic(trad(), 3).TransposeNote(from, system.Flats())
	assert.ErrorIs(t, err, note.ErrUnsetSettings)
	assert.ErrorIs(t, err, settings.ErrUnset)
	assert.Equal(t, note.Note{}, result)
}

func TestChromaticTransposePreservesImpossibleSpellingCause(t *testing.T) {
	t.Parallel()
	definition := trad().Definition()
	definition.Alterations = system.Single(ai(0))
	naturalOnly := system.MustNew(definition)
	from := noteAt(t, wp(naturalOnly, spelling.C(), 4), 0, nil)
	got, err := interval.MustNewChromatic(naturalOnly, 1).TransposeNote(from, system.Sharps())
	assert.ErrorIs(t, err, system.ErrNoSpelling)
	assert.Equal(t, note.Note{}, got)
}
