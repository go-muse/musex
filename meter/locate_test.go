package meter_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// locateFunc is one of UnitAt, GroupAt, BeatAt.
type locateFunc func(m meter.Meter, t duration.Position) (int, duration.Position, error)

func unitAt(m meter.Meter, t duration.Position) (int, duration.Position, error)  { return m.UnitAt(t) }
func groupAt(m meter.Meter, t duration.Position) (int, duration.Position, error) { return m.GroupAt(t) }
func beatAt(m meter.Meter, t duration.Position) (int, duration.Position, error)  { return m.BeatAt(t) }

// TestLocate covers О-11, П-11, Т-8, КП-16, КП-46: the number of the unit,
// beat, and group that contains a position, counted from 1, and the offset
// in it.
func TestLocate(t *testing.T) {
	t.Parallel()

	type place struct {
		n      int
		offset string
	}

	for _, c := range []struct {
		name, text, t     string
		unit, beat, group place
	}{
		{"КП-16 6/8 at 3/8", "6/8", "3/8", place{4, "0"}, place{2, "0"}, place{2, "0"}},
		{"КП-16 6/8 at 1/2", "6/8", "1/2", place{5, "0"}, place{2, "1/8"}, place{2, "1/8"}},
		{"КП-16 6/8 at 7/16", "6/8", "7/16", place{4, "1/16"}, place{2, "1/16"}, place{2, "1/16"}},
		{"КП-16 4/4 at 3/4", "4/4", "3/4", place{4, "0"}, place{4, "0"}, place{2, "1/4"}},
		{"КП-46 start of the measure", "4/4", "0", place{1, "0"}, place{1, "0"}, place{1, "0"}},
		{"КП-46 just before the end", "4/4", "15/16", place{4, "3/16"}, place{4, "3/16"}, place{2, "7/16"}},
		{"3/4 at 1/2", "3/4", "1/2", place{3, "0"}, place{3, "0"}, place{1, "1/2"}},
		{"5/8[2+3] at 1/4", "5/8[2+3]", "1/4", place{3, "0"}, place{2, "0"}, place{2, "0"}},
		{"5/8[2+3] at 3/16", "5/8[2+3]", "3/16", place{2, "1/16"}, place{1, "3/16"}, place{1, "3/16"}},
		{"2/4+3/8 at 5/8", "2/4+3/8", "5/8", place{4, "0"}, place{2, "1/8"}, place{2, "1/8"}},
		{"2/4+3/8 at 11/16", "2/4+3/8", "11/16", place{4, "1/16"}, place{2, "3/16"}, place{2, "3/16"}},
		{"2/4+3/8 at 1/4", "2/4+3/8", "1/4", place{2, "0"}, place{1, "1/4"}, place{1, "1/4"}},
		{"(3+2+3)/8 at 3/4", "(3+2+3)/8", "3/4", place{7, "0"}, place{3, "1/8"}, place{3, "1/8"}},
		{"4/3 at 1", "4/3", "1", place{4, "0"}, place{4, "0"}, place{2, "1/3"}},
		{"C| at 3/4", "C|", "3/4", place{2, "1/4"}, place{2, "1/4"}, place{1, "3/4"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			m := mt(c.text)
			at := posT(c.t)

			for _, x := range []struct {
				what string
				f    locateFunc
				want place
			}{
				{"unit", unitAt, c.unit},
				{"beat", beatAt, c.beat},
				{"group", groupAt, c.group},
			} {
				n, offset, err := x.f(m, at)
				require.NoError(t, err, x.what)
				assert.Equal(t, x.want.n, n, x.what)
				assert.Equal(t, posT(x.want.offset), offset, x.what)
			}
		})
	}

	t.Run("КП-16 units without grouping", func(t *testing.T) {
		t.Parallel()

		n, offset, err := meter.MustNew(5, 8).UnitAt(pos(1, 4))
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, duration.Origin(), offset)

		_, _, err = meter.MustNew(5, 8).BeatAt(pos(1, 4))
		assertErr(t, err, meter.ErrGroupingRequired)

		_, _, err = meter.MustNew(5, 8).GroupAt(pos(1, 4))
		assertErr(t, err, meter.ErrGroupingRequired)
	})
}

// TestLocateErrors covers КП-34, КП-38, КП-46 and the order of checks of
// UnitAt, GroupAt, BeatAt: E-1 → E-2 → E-3 → E-7 → E-10 (architecture 4.2).
func TestLocateErrors(t *testing.T) {
	t.Parallel()

	var unset duration.Position

	for _, c := range []struct {
		name   string
		text   string
		t      duration.Position
		f      locateFunc
		want   error
		causes []error
	}{
		{"КП-34 unit at 3/4", "3/4", pos(3, 4), unitAt, meter.ErrOutsideMeter, nil},
		{"КП-34 unit at 1", "3/4", pos(1, 1), unitAt, meter.ErrOutsideMeter, nil},
		{"КП-34 beat at 3/4", "3/4", pos(3, 4), beatAt, meter.ErrOutsideMeter, nil},
		{"КП-34 group at 1", "3/4", pos(1, 1), groupAt, meter.ErrOutsideMeter, nil},
		{"sum at its size", "2/4+3/8", pos(7, 8), unitAt, meter.ErrOutsideMeter, nil},
		{"КП-38 unit at an unset position", "3/4", unset, unitAt, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"КП-38 beat at an unset position", "3/4", unset, beatAt, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"КП-38 group at an unset position", "3/4", unset, groupAt, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"order: E-2 before E-3 for a beat", "5/8", unset, beatAt, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"order: E-2 before E-3 for a group", "7/8", unset, groupAt, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"order: E-3 before E-7 for a beat", "5/8", pos(5, 8), beatAt, meter.ErrGroupingRequired, nil},
		{"order: E-3 before E-7 for a group", "5/8", pos(1, 1), groupAt, meter.ErrGroupingRequired, nil},
		{"unit without grouping at the end", "5/8", pos(5, 8), unitAt, meter.ErrOutsideMeter, nil},
		{"E-10 unit offset", "3/4", rangeEdge(), unitAt, meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"E-10 beat offset", "3/4", rangeEdge(), beatAt, meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"E-10 group offset", "2/4[1+1]", rangeEdge(), groupAt, meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			n, offset, err := c.f(mt(c.text), c.t)
			assertErr(t, err, c.want, c.causes...)
			assert.Zero(t, n)
			assert.Equal(t, duration.Position{}, offset)
		})
	}

	t.Run("offset in range at the same position", func(t *testing.T) {
		t.Parallel()

		// The group of 3/4 starts at 0, so the offset is the position itself.
		n, offset, err := meter.ThreeFour().GroupAt(rangeEdge())
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, rangeEdge(), offset)
	})
}

// rangeEdge is the position t = 2^253/(2^255−1) of implementation-plan 4:
// t lies in the second quarter of a measure, and t − 1/4 = 1/(4·(2^255−1))
// is outside the range of common/fraction.
func rangeEdge() duration.Position {
	return posRat(new(big.Rat).SetFrac(pow2(253), maxComponent()))
}

// startFunc is one of UnitStart, GroupStart, BeatStart.
type startFunc func(m meter.Meter, n int) (duration.Position, error)

func unitStart(m meter.Meter, n int) (duration.Position, error)  { return m.UnitStart(n) }
func groupStart(m meter.Meter, n int) (duration.Position, error) { return m.GroupStart(n) }
func beatStart(m meter.Meter, n int) (duration.Position, error)  { return m.BeatStart(n) }

// TestStart covers О-12, П-11, Т-8, КП-17, КП-34 and the order of checks of
// UnitStart, GroupStart, BeatStart.
func TestStart(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, text string
		f          startFunc
		n          int
		want       string
	}{
		{"КП-17 4/4 beat 3", "4/4", beatStart, 3, "1/2"},
		{"КП-17 6/8 beat 2", "6/8", beatStart, 2, "3/8"},
		{"КП-17 6/8 unit 4", "6/8", unitStart, 4, "3/8"},
		{"КП-17 (3+2+3)/8 group 3", "(3+2+3)/8", groupStart, 3, "5/8"},
		{"first unit", "3/4", unitStart, 1, "0"},
		{"last unit", "3/4", unitStart, 3, "1/2"},
		{"sum unit 5", "2/4+3/8", unitStart, 5, "3/4"},
		{"sum group 2", "2/4+3/8", groupStart, 2, "1/2"},
		{"sum beat 2", "2/4+3/8", beatStart, 2, "1/2"},
		{"5/8[2+3] beat 2", "5/8[2+3]", beatStart, 2, "1/4"},
		{"unit without grouping", "5/8", unitStart, 5, "1/2"},
		{"12/8 group 4", "12/8", groupStart, 4, "9/8"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := c.f(mt(c.text), c.n)
			require.NoError(t, err)
			assert.Equal(t, posT(c.want), got)
		})
	}

	for _, c := range []struct {
		name, text string
		f          startFunc
		n          int
		want       error
	}{
		{"КП-34 beat 0", "3/4", beatStart, 0, meter.ErrOutsideMeter},
		{"КП-34 beat 4", "3/4", beatStart, 4, meter.ErrOutsideMeter},
		{"КП-34 unit 4", "3/4", unitStart, 4, meter.ErrOutsideMeter},
		{"КП-34 group 2", "3/4", groupStart, 2, meter.ErrOutsideMeter},
		{"unit -1", "3/4", unitStart, -1, meter.ErrOutsideMeter},
		{"group 0", "6/8", groupStart, 0, meter.ErrOutsideMeter},
		{"beat 3 of 6/8", "6/8", beatStart, 3, meter.ErrOutsideMeter},
		{"order: E-3 before E-7 for a beat", "5/8", beatStart, 0, meter.ErrGroupingRequired},
		{"order: E-3 before E-7 for a group", "7/8", groupStart, 99, meter.ErrGroupingRequired},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := c.f(mt(c.text), c.n)
			assertErr(t, err, c.want)
			assert.Equal(t, duration.Position{}, got)
		})
	}
}

// TestLocateAgreesWithLists covers О-11, О-12 on the corpus: the start of
// element n is Start of the n-th span, and locating that start gives n with
// offset 0; locating the middle of the element gives n.
func TestLocateAgreesWithLists(t *testing.T) {
	t.Parallel()

	for _, m := range corpus() {
		t.Run(m.String(), func(t *testing.T) {
			t.Parallel()

			checkLocate(t, "unit", noErr(m.Units()), func(n int) (duration.Position, error) { return m.UnitStart(n) },
				func(p duration.Position) (int, duration.Position, error) { return m.UnitAt(p) })

			if !noErr(m.HasGrouping()) {
				return
			}

			checkLocate(t, "group", noErr(m.Groups()), func(n int) (duration.Position, error) { return m.GroupStart(n) },
				func(p duration.Position) (int, duration.Position, error) { return m.GroupAt(p) })
			checkLocate(t, "beat", noErr(m.Beats()), func(n int) (duration.Position, error) { return m.BeatStart(n) },
				func(p duration.Position) (int, duration.Position, error) { return m.BeatAt(p) })
		})
	}
}

// checkLocate compares start and at with the spans; it samples at most 40
// elements of long lists.
func checkLocate(t *testing.T, what string, spans []meter.Span,
	start func(int) (duration.Position, error), at func(duration.Position) (int, duration.Position, error),
) {
	t.Helper()

	step := max(1, len(spans)/40)

	for i := 0; i < len(spans); i += step {
		s := spans[i]
		n := i + 1

		got, err := start(n)
		require.NoError(t, err, "%s %d", what, n)
		assert.Equal(t, s.Start, got, "%s %d start", what, n)

		k, offset, err := at(s.Start)
		require.NoError(t, err, "%s at start of %d", what, n)
		assert.Equal(t, n, k, "%s at start of %d", what, n)
		assert.Equal(t, duration.Origin(), offset)

		half := s.Duration.MustDivInt(2)
		mid := s.Start.MustAdd(half)

		k, offset, err = at(mid)
		require.NoError(t, err, "%s at the middle of %d", what, n)
		assert.Equal(t, n, k, "%s at the middle of %d", what, n)
		assert.Equal(t, posRat(ratOfDuration(half)), offset)
	}

	_, err := start(len(spans) + 1)
	assertErr(t, err, meter.ErrOutsideMeter)
}
