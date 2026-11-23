package interval_test

import (
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// measureCountingSource counts its reads.
type measureCountingSource struct {
	set   *settings.Settings
	reads atomic.Int64
}

func (s *measureCountingSource) Settings() *settings.Settings {
	s.reads.Add(1)
	return s.set
}

// measureAlternatingSource returns the standard set and the 19-position set
// in turn: an operation that reads it twice sees two different systems.
type measureAlternatingSource struct {
	reads atomic.Int64
}

func (s *measureAlternatingSource) Settings() *settings.Settings {
	if s.reads.Add(1)%2 == 1 {
		return std()
	}
	return set19()
}

// restrictedSystem admits only the alterations -1, 0 and +1.
func restrictedSystem() *system.System {
	return system.MustNew(system.Definition{
		Name:        "restricted",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 2, 4, 5, 7, 9, 11},
		Chromatic:   1,
		Alterations: system.MustIntegerRange(-1, 1),
	})
}

// TestMeasureNotes covers the acceptance pairs of the task: C4 to D#4 is A2
// and +3, C4 to Eb4 is m3 and +3, D#4 to Eb4 is an ascending diminished
// second and 0.
func TestMeasureNotes(t *testing.T) {
	t.Parallel()

	t.Run("acceptance pairs", func(t *testing.T) {
		t.Parallel()
		src := std()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		ds4 := noteAt(t, pt(spelling.DSharp(), 4), 0, src)
		ef4 := noteAt(t, pt(spelling.EFlat(), 4), 0, src)

		i, c, err := interval.MeasureNotes(c4, ds4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("A2")), "written interval: %v", i)
		assert.Equal(t, 3, c.Steps())
		assert.True(t, c.System().Equal(trad()))
		assert.Equal(t, 2, i.Generic().Size())

		i, c, err = interval.MeasureNotes(c4, ef4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("m3")), "written interval: %v", i)
		assert.Equal(t, 3, c.Steps())
		assert.Equal(t, 3, i.Generic().Size())

		i, c, err = interval.MeasureNotes(ds4, ef4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("d2")), "written interval: %v", i)
		assert.Equal(t, interval.Up, i.Direction(), "the diminished second ascends")
		assert.Equal(t, 0, c.Steps())
	})

	t.Run("descending and compound", func(t *testing.T) {
		t.Parallel()
		src := std()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		d4 := noteAt(t, pt(spelling.D(), 4), 0, src)
		g5 := noteAt(t, pt(spelling.G(), 5), 0, src)

		i, c, err := interval.MeasureNotes(d4, c4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("↓M2")), "written interval: %v", i)
		assert.Equal(t, -2, c.Steps())

		i, c, err = interval.MeasureNotes(c4, g5)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("P12")), "written interval: %v", i)
		assert.Equal(t, 19, c.Steps())
	})

	t.Run("cents do not take part", func(t *testing.T) {
		t.Parallel()
		src := std()
		c4 := noteAt(t, pt(spelling.C(), 4), 30, src)
		e4 := noteAt(t, pt(spelling.E(), 4), -20, src)
		i, c, err := interval.MeasureNotes(c4, e4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("M3")))
		assert.Equal(t, 4, c.Steps())
	})

	t.Run("one snapshot per distinct source", func(t *testing.T) {
		t.Parallel()
		src := &measureCountingSource{set: std()}
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, src)
		reads := src.reads.Load() // creation reads the source too
		i, c, err := interval.MeasureNotes(c4, e4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("M3")))
		assert.Equal(t, 4, c.Steps())
		assert.Equal(t, reads+1, src.reads.Load(), "the shared source is read once")
	})

	t.Run("equal distinct sources are read once each", func(t *testing.T) {
		t.Parallel()
		a := &measureCountingSource{set: std()}
		b := &measureCountingSource{set: std()}
		c4 := noteAt(t, pt(spelling.C(), 4), 0, a)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, b)
		readsA, readsB := a.reads.Load(), b.reads.Load()
		_, _, err := interval.MeasureNotes(c4, e4)
		require.NoError(t, err)
		assert.Equal(t, readsA+1, a.reads.Load())
		assert.Equal(t, readsB+1, b.reads.Load())
	})

	t.Run("both results share one snapshot of a changing source", func(t *testing.T) {
		t.Parallel()
		src := &measureAlternatingSource{}
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, src)
		reads := src.reads.Load()
		i, c, err := interval.MeasureNotes(c4, e4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("M3")))
		assert.Equal(t, 4, c.Steps())
		assert.True(t, c.System().Equal(trad()))
		assert.Equal(t, reads+1, src.reads.Load(), "both results use one read")
	})

	t.Run("different systems", func(t *testing.T) {
		t.Parallel()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, std())
		c4other := noteAt(t, pt(spelling.C(), 4), 0, set19())
		_, _, err := interval.MeasureNotes(c4, c4other)
		assertErr(t, err, interval.ErrDifferentSystems)
	})

	t.Run("unbound note", func(t *testing.T) {
		t.Parallel()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, nil)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, std())
		_, _, err := interval.MeasureNotes(c4, e4)
		assertErr(t, err, interval.ErrNoSystem)
	})

	t.Run("note without octave", func(t *testing.T) {
		t.Parallel()
		c := noteName(t, spelling.C(), 0, std())
		e4 := noteAt(t, pt(spelling.E(), 4), 0, std())
		_, _, err := interval.MeasureNotes(c, e4)
		assertErr(t, err, interval.ErrNoOctave)
		_, _, err = interval.MeasureNotes(e4, c)
		assertErr(t, err, interval.ErrNoOctave)
	})

	t.Run("unset note", func(t *testing.T) {
		t.Parallel()
		e4 := noteAt(t, pt(spelling.E(), 4), 0, std())
		_, _, err := interval.MeasureNotes(note.Note{}, e4)
		assertErr(t, err, interval.ErrUnset)
	})

	t.Run("inadmissible name in the own system", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		x4 := noteAt(t, pt(spelling.CDoubleSharp(), 4), 0, src)
		src.set(std().MustWithSystem(restrictedSystem()))
		_, _, err := interval.MeasureNotes(c4, x4)
		assertErr(t, err, nil, note.ErrInadmissible, system.ErrInadmissible)
	})

	t.Run("Must variants", func(t *testing.T) {
		t.Parallel()
		src := std()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, src)
		i, c := interval.MustMeasureNotes(c4, e4)
		assert.True(t, i.Equal(iv("M3")))
		assert.Equal(t, 4, c.Steps())
		err := panicError(t, func() { interval.MustMeasureNotes(note.Note{}, e4) })
		assertErr(t, err, interval.ErrUnset)
	})
}

// TestMeasureNotesIn covers the explicit-system form: the sources are not
// read, and the system is checked before the notes.
func TestMeasureNotesIn(t *testing.T) {
	t.Parallel()

	t.Run("acceptance pairs without sources", func(t *testing.T) {
		t.Parallel()
		src := &measureCountingSource{set: std()}
		c4 := noteAt(t, pt(spelling.C(), 4), 0, src)
		ds4 := noteAt(t, pt(spelling.DSharp(), 4), 0, src)
		ef4 := noteAt(t, pt(spelling.EFlat(), 4), 0, src)
		reads := src.reads.Load()

		i, c, err := interval.MeasureNotesIn(trad(), c4, ds4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("A2")))
		assert.Equal(t, 3, c.Steps())

		i, c, err = interval.MeasureNotesIn(trad(), ds4, ef4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("d2")))
		assert.Equal(t, 0, c.Steps())
		assert.Equal(t, reads, src.reads.Load(), "the explicit form reads no source")
	})

	t.Run("unbound notes work", func(t *testing.T) {
		t.Parallel()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, nil)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, nil)
		i, c, err := interval.MeasureNotesIn(trad(), c4, e4)
		require.NoError(t, err)
		assert.True(t, i.Equal(iv("M3")))
		assert.Equal(t, 4, c.Steps())
	})

	t.Run("unset system first", func(t *testing.T) {
		t.Parallel()
		_, _, err := interval.MeasureNotesIn(nil, note.Note{}, note.Note{})
		assertErr(t, err, interval.ErrUnset)
	})

	t.Run("unset note", func(t *testing.T) {
		t.Parallel()
		e4 := noteAt(t, pt(spelling.E(), 4), 0, nil)
		_, _, err := interval.MeasureNotesIn(trad(), note.Note{}, e4)
		assertErr(t, err, interval.ErrUnset)
	})

	t.Run("note without octave", func(t *testing.T) {
		t.Parallel()
		c := noteName(t, spelling.C(), 0, nil)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, nil)
		_, _, err := interval.MeasureNotesIn(trad(), c, e4)
		assertErr(t, err, interval.ErrNoOctave)
	})

	t.Run("inadmissible name", func(t *testing.T) {
		t.Parallel()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, nil)
		x4 := noteAt(t, pt(spelling.CDoubleSharp(), 4), 0, nil)
		_, _, err := interval.MeasureNotesIn(restrictedSystem(), c4, x4)
		assertErr(t, err, nil, system.ErrInadmissible)
	})

	t.Run("Must variants", func(t *testing.T) {
		t.Parallel()
		c4 := noteAt(t, pt(spelling.C(), 4), 0, nil)
		e4 := noteAt(t, pt(spelling.E(), 4), 0, nil)
		i, c := interval.MustMeasureNotesIn(trad(), c4, e4)
		assert.True(t, i.Equal(iv("M3")))
		assert.Equal(t, 4, c.Steps())
		err := panicError(t, func() { interval.MustMeasureNotesIn(nil, c4, e4) })
		assertErr(t, err, interval.ErrUnset)
	})
}
