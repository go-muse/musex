package note_test

import (
	"testing"

	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reads of sources (P-5, P-5a, KP-2; architecture 6.3): an operation reads
// each needed source exactly once; notes of one source share one snapshot.
// The number of reads is observed only on counting sources; for two separate
// standard sets only the result is checked.

// multiOp is an operation over two notes, or a list of three, with a set.
type multiOp struct {
	name string
	call func(x, y note.Note) error
}

func multiOps() []multiOp {
	return []multiOp{
		{"Equivalent", func(x, y note.Note) error { _, err := x.Equivalent(y); return err }},
		{"ComparePosition", func(x, y note.Note) error { _, err := x.ComparePosition(y); return err }},
		{"Between", func(x, y note.Note) error { _, err := x.Between(y, x); return err }},
		{"Distance", func(x, y note.Note) error { _, err := x.Distance(y); return err }},
		{"CompareSound", func(x, y note.Note) error { _, err := x.CompareSound(y); return err }},
		{"CompareSoundBy", func(x, y note.Note) error {
			_, err := x.CompareSoundBy(y, frequency.MustWithinCents(ct(1)))
			return err
		}},
		{"SoundsSame", func(x, y note.Note) error { _, err := x.SoundsSame(y); return err }},
		{"SoundsSameBy", func(x, y note.Note) error { _, err := x.SoundsSameBy(y, frequency.Strict()); return err }},
		{"CentsTo", func(x, y note.Note) error { _, err := x.CentsTo(y); return err }},
		{"Walk", func(x, y note.Note) error { _, err := x.Walk(y, system.Sharps()); return err }},
		{"SortedByPosition", func(x, y note.Note) error { _, err := note.SortedByPosition([]note.Note{y, x, y}); return err }},
		{"SortedBySound", func(x, y note.Note) error { _, err := note.SortedBySound([]note.Note{y, x, y}); return err }},
	}
}

// TestReadsOfSeveralNotes covers P-5 and P-5a for the operations over
// several notes with their own sources.
func TestReadsOfSeveralNotes(t *testing.T) {
	t.Parallel()

	for _, op := range multiOps() {
		t.Run(op.name+" one source", func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			x, y := mk(t, spelling.C(), 4, ct(0), src), mk(t, spelling.E(), 4, ct(0), src)
			before := src.reads()
			require.NoError(t, op.call(x, y))
			assert.EqualValues(t, 1, src.reads()-before)
		})
		t.Run(op.name+" two sources", func(t *testing.T) {
			t.Parallel()
			a, b := newCounting(std()), newCounting(std())
			x, y := mk(t, spelling.C(), 4, ct(0), a), mk(t, spelling.E(), 4, ct(0), b)
			ra, rb := a.reads(), b.reads()
			require.NoError(t, op.call(x, y))
			assert.EqualValues(t, 1, a.reads()-ra)
			assert.EqualValues(t, 1, b.reads()-rb)
		})
		t.Run(op.name+" alternating source gives one snapshot", func(t *testing.T) {
			t.Parallel()
			src := &alternatingSource{a: std(), b: set19()}
			x, y := mk(t, spelling.C(), 4, ct(0), src), mk(t, spelling.E(), 4, ct(0), src)
			for range 4 {
				require.NoError(t, op.call(x, y), "K-24a: no false E-6 and no mixture")
			}
		})
		t.Run(op.name+" two standard sets", func(t *testing.T) {
			t.Parallel()
			x, y := mk(t, spelling.C(), 4, ct(0), settings.Standard()), mk(t, spelling.E(), 4, ct(0), settings.Standard())
			require.NoError(t, op.call(x, y), "K-36: one source, no false E-6")
		})
	}
}

// TestReadsOfExplicitSources checks that the In forms over several notes read
// the explicit source once and the own sources never.
func TestReadsOfExplicitSources(t *testing.T) {
	t.Parallel()

	ops := []struct {
		name string
		call func(src settings.Source, x, y note.Note) error
	}{
		{"EquivalentIn", func(s settings.Source, x, y note.Note) error { _, err := x.EquivalentIn(s, y); return err }},
		{"ComparePositionIn", func(s settings.Source, x, y note.Note) error { _, err := x.ComparePositionIn(s, y); return err }},
		{"BetweenIn", func(s settings.Source, x, y note.Note) error { _, err := x.BetweenIn(s, y, x); return err }},
		{"DistanceIn", func(s settings.Source, x, y note.Note) error { _, err := x.DistanceIn(s, y); return err }},
		{"CompareSoundIn", func(s settings.Source, x, y note.Note) error { _, err := x.CompareSoundIn(s, y); return err }},
		{"CompareSoundByIn", func(s settings.Source, x, y note.Note) error {
			_, err := x.CompareSoundByIn(s, y, frequency.Strict())
			return err
		}},
		{"SoundsSameIn", func(s settings.Source, x, y note.Note) error { _, err := x.SoundsSameIn(s, y); return err }},
		{"SoundsSameByIn", func(s settings.Source, x, y note.Note) error {
			_, err := x.SoundsSameByIn(s, y, frequency.Strict())
			return err
		}},
		{"CentsToIn", func(s settings.Source, x, y note.Note) error { _, err := x.CentsToIn(s, y); return err }},
		{"SortedByPositionIn", func(s settings.Source, x, y note.Note) error {
			_, err := note.SortedByPositionIn(s, []note.Note{y, x})
			return err
		}},
		{"SortedBySoundIn", func(s settings.Source, x, y note.Note) error {
			_, err := note.SortedBySoundIn(s, []note.Note{y, x})
			return err
		}},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			t.Parallel()
			own, explicit := newCounting(std()), newCounting(std())
			x, y := mk(t, spelling.C(), 4, ct(0), own), mk(t, spelling.E(), 4, ct(0), own)
			before := own.reads()
			require.NoError(t, op.call(explicit, x, y))
			assert.EqualValues(t, 1, explicit.reads())
			assert.Equal(t, before, own.reads(), "the own sources are not read")
		})
	}

	t.Run("WalkIn of a bound note: explicit and own once each", func(t *testing.T) {
		t.Parallel()
		own, explicit := newCounting(std()), newCounting(std())
		x, y := mk(t, spelling.C(), 4, ct(0), own), mk(t, spelling.E(), 4, ct(0), own)
		before := own.reads()
		_, err := x.WalkIn(explicit, y, system.Sharps())
		require.NoError(t, err)
		assert.EqualValues(t, 1, explicit.reads())
		assert.EqualValues(t, 1, own.reads()-before)
	})

	t.Run("WalkIn when the explicit source is the own one", func(t *testing.T) {
		t.Parallel()
		own := newCounting(std())
		x, y := mk(t, spelling.C(), 4, ct(0), own), mk(t, spelling.E(), 4, ct(0), own)
		before := own.reads()
		_, err := x.WalkIn(own, y, system.Sharps())
		require.NoError(t, err)
		assert.EqualValues(t, 1, own.reads()-before)
	})

	t.Run("OffsetIn reads only the explicit source", func(t *testing.T) {
		t.Parallel()
		own, explicit := newCounting(std()), newCounting(std())
		x := mk(t, spelling.A(), 4, ct(0), own)
		before := own.reads()
		_, _, err := x.OffsetIn(explicit, 440)
		require.NoError(t, err)
		assert.EqualValues(t, 1, explicit.reads())
		assert.Equal(t, before, own.reads())
	})

	t.Run("Offset reads the own source once", func(t *testing.T) {
		t.Parallel()
		own := newCounting(std())
		x := mk(t, spelling.A(), 4, ct(0), own)
		before := own.reads()
		_, _, err := x.OffsetBy(440, frequency.Strict())
		require.NoError(t, err)
		assert.EqualValues(t, 1, own.reads()-before)
	})
}

// TestAlternatingSourceGivesWholeSets covers KP-2 and G-6 by sound: a source
// that changes its set on every read never gives a mixed result.
func TestAlternatingSourceGivesWholeSets(t *testing.T) {
	t.Parallel()

	src := &alternatingSource{a: std(), b: set19()}
	x, y := mk(t, spelling.C(), 4, ct(0), src), mk(t, spelling.C(), 4, ct(0), src)
	for range 6 {
		same, err := x.SoundsSame(y)
		require.NoError(t, err)
		assert.True(t, same, "both notes in one snapshot")
		d, err := x.CentsTo(y)
		require.NoError(t, err)
		assert.Zero(t, d.Sign())
	}
}
