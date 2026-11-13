package duration_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWrittenModel covers ID-1, O-3, T-2, and KP-38: the value of every
// base with 0 to 6 dots under each test chain agrees with the big.Rat model;
// the record is stored as given.
func TestNewWrittenModel(t *testing.T) {
	t.Parallel()

	for _, row := range baseTable {
		for dots := range 7 {
			for _, chain := range testChains() {
				w, err := duration.NewWritten(row.base, dots, chain...)
				require.NoError(t, err)

				assert.True(t, w.IsSet())
				assert.Equal(t, row.base, w.MustBase())
				assert.Equal(t, dots, w.MustDots())
				assert.Equal(t, chain, w.MustChain())
				assert.True(t, w.MustCoefficient().EqualExact(tuplet.MustChainCoefficient(chain...)))
				assert.Equal(t, 0, ratOfDuration(w.MustDuration()).Cmp(writtenRat(row.base, dots, chain)),
					"%s with %d dots and chain %v", row.text, dots, chain)
				assert.Equal(t, dots > 0, w.MustHasDots())
				assert.Equal(t, len(chain) > 0, w.MustHasTuplets())
				assert.Equal(t, dots == 0 && len(chain) == 0, w.MustIsBasic())
			}
		}
	}
}

// TestWrittenExamples covers KP-1 to KP-4, KP-38, KP-40, and KP-49.
func TestWrittenExamples(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		w    duration.Written
		want duration.Duration
	}{
		{"KP-1 one dot", wr(duration.Quarter, 1), dur(3, 8)},
		{"KP-1 two dots", wr(duration.Quarter, 2), dur(7, 16)},
		{"KP-1 three dots", wr(duration.Quarter, 3), dur(15, 32)},
		{"KP-2 dotted quarter in 3:2", wr(duration.Quarter, 1, tuplet.Triplet()), dur(1, 4)},
		{"KP-2 eighth in 3:2", wr(duration.Eighth, 0, tuplet.Triplet()), dur(1, 12)},
		{"KP-2 eighth in 2:3", wr(duration.Eighth, 0, tuplet.Duplet()), dur(3, 16)},
		{"KP-3 nested", wr(duration.Quarter, 0, tuplet.Quintuplet(), tuplet.Triplet()), dur(2, 15)},
		{"KP-4 maxima", wr(duration.Maxima, 0), dur(8, 1)},
		{"KP-4 longa", wr(duration.Longa, 0), dur(4, 1)},
		{"KP-4 breve", wr(duration.Breve, 0), dur(2, 1)},
		{"KP-4 1024th", wr(duration.ThousandTwentyFourth, 0), dur(1, 1024)},
		{"KP-4 dotted 1024th", wr(duration.ThousandTwentyFourth, 1), dur(3, 2048)},
		{"KP-38 empty chain", wr(duration.Quarter, 0), dur(1, 4)},
		{"KP-40 neutral tuplet", wr(duration.Quarter, 0, tp(1, 1)), dur(1, 4)},
		{"maxima with 254 dots", wr(duration.Maxima, 254), durRat(ratDiv(ratInt(maxComponent()), pow2(251)))},
		{"quarter with 252 dots", wr(duration.Quarter, 252), durRat(writtenRat(duration.Quarter, 252, nil))},
	} {
		assert.Equal(t, c.want, c.w.MustDuration(), c.name)
	}

	// KP-38: an empty chain has coefficient 1 and no tuplets.
	q := wr(duration.Quarter, 0)
	assert.True(t, q.MustCoefficient().EqualExact(fraction.One()))
	assert.False(t, q.MustHasTuplets())
	assert.Nil(t, q.MustChain(), "an empty chain is nil")

	// KP-40: 1:1 changes the record but not the value.
	n := wr(duration.Quarter, 0, tp(1, 1))
	assert.NotEqual(t, q, n)
	assert.False(t, q.Equal(n))
	assert.True(t, q.MustEqualDuration(n))
	assert.True(t, n.MustHasTuplets())
	assert.False(t, n.MustIsBasic())

	// KP-49.
	w := wr(duration.Quarter, 1, tuplet.Triplet())
	assert.Equal(t, duration.Quarter, w.MustBase())
	assert.Equal(t, 1, w.MustDots())
	assert.Equal(t, []tuplet.Tuplet{tuplet.Triplet()}, w.MustChain())
	assert.True(t, w.MustCoefficient().EqualExact(frac(2, 3)))
	assert.Equal(t, dur(1, 4), w.MustDuration())
	assert.True(t, w.MustHasDots())
	assert.True(t, w.MustHasTuplets())
	assert.False(t, w.MustIsBasic())
	assert.True(t, wr(duration.Quarter, 0).MustIsBasic())

	// The coefficient keeps the unreduced pair of the chain (KR-D4).
	assert.True(t, wr(duration.Quarter, 0, tp(6, 4)).MustCoefficient().EqualExact(frac(4, 6)))
	assert.True(t, wr(duration.Quarter, 0, tuplet.Quintuplet(), tuplet.Triplet()).MustCoefficient().EqualExact(frac(8, 15)))

	// The chain is stored as given (tuplet IP-2, IP-6).
	assert.Equal(t, []tuplet.Tuplet{tp(6, 4)}, wr(duration.Quarter, 0, tp(6, 4)).MustChain())
	assert.Equal(t, []tuplet.Tuplet{tuplet.Triplet(), tuplet.Quintuplet()},
		wr(duration.Quarter, 0, tuplet.Triplet(), tuplet.Quintuplet()).MustChain())
}

// ratDiv returns a / b.
func ratDiv(a, b *big.Rat) *big.Rat { return new(big.Rat).Quo(a, b) }

// TestNewWrittenErrors covers KP-24, KP-37, and architecture 4.5.
func TestNewWrittenErrors(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name   string
		base   duration.Base
		dots   int
		chain  []tuplet.Tuplet
		want   error
		causes []error
	}{
		{"base zero", 0, 0, nil, duration.ErrInvalidArgument, nil},
		{"base 15", 15, 0, nil, duration.ErrInvalidArgument, nil},
		{"negative dots", duration.Quarter, -1, nil, duration.ErrInvalidArgument, nil},
		{"min int dots", duration.Quarter, math.MinInt, nil, duration.ErrInvalidArgument, nil},
		{"unset tuplet", duration.Quarter, 0, []tuplet.Tuplet{{}}, duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"unset inner tuplet", duration.Quarter, 0, []tuplet.Tuplet{tuplet.Triplet(), {}}, duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"KP-37 300 dots", duration.Quarter, 300, nil, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"quarter with 253 dots", duration.Quarter, 253, nil, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"maxima with 255 dots", duration.Maxima, 255, nil, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"511 dots", duration.Maxima, 511, nil, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"coefficient", duration.Quarter, 0, bigChain(5), duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"value under four tuplets", duration.ThousandTwentyFourth, 0, bigChain(4), duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"sixteenth under four tuplets", duration.Sixteenth, 0, bigChain(4), duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			w, err := duration.NewWritten(c.base, c.dots, c.chain...)
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, duration.Written{}, w)
		})
	}

	// The edge value is in range (architecture 4.5).
	_, err := duration.NewWritten(duration.Eighth, 0, bigChain(4)...)
	require.NoError(t, err)
}

// TestNewWrittenManyDots covers P-2 and architecture 4.5: from 512 dots on
// the value is always out of range, and the failure does not depend on the
// number of dots.
func TestNewWrittenManyDots(t *testing.T) {
	t.Parallel()

	for _, dots := range []int{512, 513, 1000, 1 << 20, 1 << 40, math.MaxInt} {
		for _, chain := range [][]tuplet.Tuplet{nil, {tp(1, 2)}, bigChain(4)} {
			r := fast(t, func() result[duration.Written] {
				w, err := duration.NewWritten(duration.Maxima, dots, chain...)

				return result[duration.Written]{w, err}
			})
			assertErr(t, r.err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
			assert.Equal(t, duration.Written{}, r.v, "%d dots", dots)
		}

		r := fast(t, func() result[duration.Written] {
			w, err := wr(duration.Quarter, 0).WithDots(dots)

			return result[duration.Written]{w, err}
		})
		assertErr(t, r.err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, duration.Written{}, r.v, "WithDots(%d)", dots)
	}
}

// TestWrittenChainOwnership covers architecture 3.2: Chain returns a new
// slice on each call; a nil slice for an empty chain.
func TestWrittenChainOwnership(t *testing.T) {
	t.Parallel()

	w := wr(duration.Quarter, 0, tuplet.Triplet(), tuplet.Quintuplet())

	first := w.MustChain()
	first[0] = tuplet.Septuplet()

	assert.Equal(t, []tuplet.Tuplet{tuplet.Triplet(), tuplet.Quintuplet()}, w.MustChain())
	assert.Equal(t, wr(duration.Quarter, 0, tuplet.Triplet(), tuplet.Quintuplet()), w)

	chain, err := wr(duration.Half, 2).Chain()
	require.NoError(t, err)
	assert.Nil(t, chain)
}

// TestWrittenModify covers O-6, P-10, KP-18, KP-45, KP-50, and the table of
// architecture 4.5.
func TestWrittenModify(t *testing.T) {
	t.Parallel()

	q := wr(duration.Quarter, 0)
	qt := wr(duration.Quarter, 0, tuplet.Triplet())
	q54t := wr(duration.Quarter, 0, tuplet.Quintuplet(), tuplet.Triplet())

	ok := []struct {
		name string
		run  func() (duration.Written, error)
		want duration.Written
	}{
		{"WithBase", func() (duration.Written, error) {
			return wr(duration.Quarter, 1, tuplet.Triplet()).WithBase(duration.Half)
		},
			wr(duration.Half, 1, tuplet.Triplet())},
		{"WithBase same", func() (duration.Written, error) { return qt.WithBase(duration.Quarter) }, qt},
		{"KP-18 ShiftBase longer", func() (duration.Written, error) { return wr(duration.Quarter, 1).ShiftBase(1) }, wr(duration.Half, 1)},
		{"KP-18 ShiftBase shorter", func() (duration.Written, error) { return wr(duration.Eighth, 0, tuplet.Triplet()).ShiftBase(-1) },
			wr(duration.Sixteenth, 0, tuplet.Triplet())},
		{"ShiftBase zero", func() (duration.Written, error) { return qt.ShiftBase(0) }, qt},
		{"ShiftBase to maxima", func() (duration.Written, error) { return wr(duration.Quarter, 2).ShiftBase(5) }, wr(duration.Maxima, 2)},
		{"ShiftBase to 1024th", func() (duration.Written, error) { return wr(duration.Quarter, 2).ShiftBase(-8) }, wr(duration.ThousandTwentyFourth, 2)},
		{"WithDots", func() (duration.Written, error) { return qt.WithDots(2) }, wr(duration.Quarter, 2, tuplet.Triplet())},
		{"WithDots zero", func() (duration.Written, error) { return wr(duration.Quarter, 3).WithDots(0) }, q},
		{"KP-50 AddDot", func() (duration.Written, error) { return q.AddDot() }, wr(duration.Quarter, 1)},
		{"KP-50 RemoveDot", func() (duration.Written, error) { return wr(duration.Quarter, 1).RemoveDot() }, q},
		{"RemoveDot of two", func() (duration.Written, error) { return wr(duration.Quarter, 2, tp(6, 4)).RemoveDot() },
			wr(duration.Quarter, 1, tp(6, 4))},
		{"WithChain", func() (duration.Written, error) { return wr(duration.Quarter, 1).WithChain(tp(6, 4), tuplet.Triplet()) },
			wr(duration.Quarter, 1, tp(6, 4), tuplet.Triplet())},
		{"WithChain empty", func() (duration.Written, error) { return q54t.WithChain() }, q},
		{"KP-50 AddOuterTuplet", func() (duration.Written, error) { return qt.AddOuterTuplet(tuplet.Quintuplet()) }, q54t},
		{"KP-50 AddInnerTuplet", func() (duration.Written, error) { return qt.AddInnerTuplet(tuplet.Quintuplet()) },
			wr(duration.Quarter, 0, tuplet.Triplet(), tuplet.Quintuplet())},
		{"AddOuterTuplet to empty", func() (duration.Written, error) { return q.AddOuterTuplet(tp(6, 4)) }, wr(duration.Quarter, 0, tp(6, 4))},
		{"AddInnerTuplet to empty", func() (duration.Written, error) { return q.AddInnerTuplet(tp(1, 1)) }, wr(duration.Quarter, 0, tp(1, 1))},
		{"KP-50 RemoveOuterTuplet", func() (duration.Written, error) { return q54t.RemoveOuterTuplet() }, qt},
		{"RemoveInnerTuplet", func() (duration.Written, error) { return q54t.RemoveInnerTuplet() }, wr(duration.Quarter, 0, tuplet.Quintuplet())},
		{"RemoveOuterTuplet of one", func() (duration.Written, error) { return qt.RemoveOuterTuplet() }, q},
		{"RemoveInnerTuplet of one", func() (duration.Written, error) { return qt.RemoveInnerTuplet() }, q},
		{"KP-50 RemoveTuplets", func() (duration.Written, error) { return q54t.RemoveTuplets() }, q},
		{"KP-50 RemoveTuplets of empty", func() (duration.Written, error) { return q.RemoveTuplets() }, q},
	}

	for _, c := range ok {
		got, err := c.run()
		require.NoError(t, err, c.name)
		assert.Equal(t, c.want, got, c.name)
	}

	// KP-50 values.
	assert.Equal(t, dur(3, 8), q.MustAddDot().MustDuration())
	assert.Equal(t, dur(2, 15), q54t.MustDuration())
	assert.Equal(t, dur(2, 15), qt.MustAddInnerTuplet(tuplet.Quintuplet()).MustDuration())
	assert.Equal(t, dur(1, 6), q54t.MustRemoveOuterTuplet().MustDuration())
	assert.Equal(t, dur(1, 4), q54t.MustRemoveTuplets().MustDuration())
}

// TestWrittenModifyErrors covers KP-34, KP-45, and architecture 4.5: special
// failures, E-11 for results out of range, and zero results.
func TestWrittenModifyErrors(t *testing.T) {
	t.Parallel()

	q := wr(duration.Quarter, 0)
	edge := edgeWritten()

	// Quarter with 253 dots is out of range; under 1:2 (coefficient 2) it is
	// in range, so removing the tuplet leaves the range.
	doubled := wr(duration.Quarter, 253, tp(1, 2))

	failing := []struct {
		name   string
		run    func() (duration.Written, error)
		want   error
		causes []error
	}{
		{"WithBase invalid", func() (duration.Written, error) { return q.WithBase(0) }, duration.ErrInvalidArgument, nil},
		{"WithBase 15", func() (duration.Written, error) { return q.WithBase(15) }, duration.ErrInvalidArgument, nil},
		{"WithBase range", func() (duration.Written, error) { return edge.WithBase(duration.Sixteenth) },
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"KP-34 ShiftBase longer than maxima", func() (duration.Written, error) { return wr(duration.Maxima, 1).ShiftBase(1) },
			duration.ErrOutsideSeries, nil},
		{"KP-34 ShiftBase shorter than 1024th", func() (duration.Written, error) { return wr(duration.ThousandTwentyFourth, 0).ShiftBase(-1) },
			duration.ErrOutsideSeries, nil},
		{"ShiftBase max int", func() (duration.Written, error) { return q.ShiftBase(math.MaxInt) }, duration.ErrOutsideSeries, nil},
		{"ShiftBase min int", func() (duration.Written, error) { return q.ShiftBase(math.MinInt) }, duration.ErrOutsideSeries, nil},
		{"ShiftBase range", func() (duration.Written, error) { return edge.ShiftBase(-1) }, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"WithDots negative", func() (duration.Written, error) { return q.WithDots(-1) }, duration.ErrInvalidArgument, nil},
		{"WithDots range", func() (duration.Written, error) { return q.WithDots(300) }, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"AddDot range quarter", func() (duration.Written, error) { return wr(duration.Quarter, 252).AddDot() },
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"AddDot range maxima", func() (duration.Written, error) { return wr(duration.Maxima, 254).AddDot() },
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"AddDot range edge", func() (duration.Written, error) { return edge.AddDot() }, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"KP-45 RemoveDot without dots", func() (duration.Written, error) { return q.RemoveDot() }, duration.ErrInvalidArgument, nil},
		{"WithChain unset", func() (duration.Written, error) { return q.WithChain(tuplet.Triplet(), tuplet.Tuplet{}) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"WithChain coefficient", func() (duration.Written, error) { return q.WithChain(bigChain(5)...) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"WithChain value", func() (duration.Written, error) {
			return wr(duration.ThousandTwentyFourth, 0).WithChain(bigChain(4)...)
		},
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"WithChain empty range", func() (duration.Written, error) { return doubled.WithChain() }, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"AddOuterTuplet unset", func() (duration.Written, error) { return q.AddOuterTuplet(tuplet.Tuplet{}) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"AddInnerTuplet unset", func() (duration.Written, error) { return q.AddInnerTuplet(tuplet.Tuplet{}) },
			duration.ErrInvalidArgument, []error{tuplet.ErrUnset}},
		{"AddOuterTuplet coefficient", func() (duration.Written, error) { return edge.AddOuterTuplet(bigTuplet) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"AddInnerTuplet coefficient", func() (duration.Written, error) { return edge.AddInnerTuplet(bigTuplet) },
			duration.ErrOutOfRange, []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"AddOuterTuplet value", func() (duration.Written, error) {
			return wr(duration.ThousandTwentyFourth, 0, bigChain(3)...).AddOuterTuplet(bigTuplet)
		}, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"AddInnerTuplet value", func() (duration.Written, error) {
			return wr(duration.ThousandTwentyFourth, 0, bigChain(3)...).AddInnerTuplet(bigTuplet)
		}, duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"KP-45 RemoveOuterTuplet empty", func() (duration.Written, error) { return q.RemoveOuterTuplet() }, duration.ErrInvalidArgument, nil},
		{"RemoveInnerTuplet empty", func() (duration.Written, error) { return q.RemoveInnerTuplet() }, duration.ErrInvalidArgument, nil},
		{"RemoveOuterTuplet range", func() (duration.Written, error) { return doubled.RemoveOuterTuplet() },
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"RemoveInnerTuplet range", func() (duration.Written, error) { return doubled.RemoveInnerTuplet() },
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"RemoveTuplets range", func() (duration.Written, error) { return doubled.RemoveTuplets() },
			duration.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
	}

	for _, c := range failing {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := c.run()
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, duration.Written{}, got)
		})
	}
}
