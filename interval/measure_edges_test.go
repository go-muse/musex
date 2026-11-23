package interval_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/spelling"
)

// TestMeasureNotesErrorPaths covers the settings error of either note and
// the out-of-range result at the edges of int.
func TestMeasureNotesErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("second note settings error", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		from := noteAt(t, pt(spelling.C(), 4), 0, std())
		to := noteAt(t, pt(spelling.E(), 4), 0, src)
		src.set(nil)
		_, _, err := interval.MeasureNotes(from, to)
		assertErr(t, err, nil, note.ErrUnsetSettings)
	})

	t.Run("first note settings error", func(t *testing.T) {
		t.Parallel()
		src := newSwitch(std())
		from := noteAt(t, pt(spelling.C(), 4), 0, src)
		to := noteAt(t, pt(spelling.E(), 4), 0, std())
		src.set(nil)
		_, _, err := interval.MeasureNotes(from, to)
		assertErr(t, err, nil, note.ErrUnsetSettings)
	})

	t.Run("displacement overflow", func(t *testing.T) {
		t.Parallel()
		hi := noteAt(t, pt(spelling.C(), math.MaxInt), 0, std())
		lo := noteAt(t, pt(spelling.C(), math.MinInt), 0, std())
		_, _, err := interval.MeasureNotes(hi, lo)
		assertErr(t, err, interval.ErrOutOfRange, octave.ErrOutOfRange)
		_, _, err = interval.MeasureNotesIn(trad(), hi, lo)
		assertErr(t, err, interval.ErrOutOfRange, octave.ErrOutOfRange)
		assertErr(t, panicError(t, func() { interval.MustMeasureNotes(hi, lo) }), interval.ErrOutOfRange, octave.ErrOutOfRange)
		assertErr(t, panicError(t, func() { interval.MustMeasureNotesIn(trad(), hi, lo) }), interval.ErrOutOfRange, octave.ErrOutOfRange)
	})
}
