package duration_test

import (
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPositionCreate covers O-12, KP-39, P-7, and architecture 4.9.
func TestPositionCreate(t *testing.T) {
	t.Parallel()

	origin := duration.Origin()
	assert.True(t, origin.IsSet())
	assert.NotEqual(t, duration.Position{}, origin, "the origin is not the unset position")
	assert.True(t, origin.MustFraction().EqualExact(fraction.Zero()), "the origin is 0/1")
	assert.Equal(t, "0/1", origin.String())

	for _, c := range []struct {
		name     string
		num, den int
		want     duration.Position
	}{
		{"zero", 0, 1, origin},
		{"zero over seven", 0, 7, origin},
		{"zero over negative", 0, -3, origin},
		{"quarter", 1, 4, posRat(rat(1, 4))},
		{"both negative", -1, -4, posRat(rat(1, 4))},
		{"unreduced", 6, 16, posRat(rat(3, 8))},
		{"max int", math.MaxInt, 1, posRat(new(big.Rat).SetInt64(math.MaxInt))},
	} {
		p, err := duration.NewPosition(c.num, c.den)
		require.NoError(t, err, c.name)
		assert.Equal(t, c.want, p, c.name)
	}

	assert.True(t, pos(6, 16).MustFraction().EqualExact(frac(3, 8)), "stored reduced")

	for _, c := range []struct {
		name     string
		num, den int
		want     error
		causes   []error
	}{
		{"negative", -1, 4, duration.ErrNegativePosition, nil},
		{"negative denominator", 1, -4, duration.ErrNegativePosition, nil},
		{"min int", math.MinInt, 1, duration.ErrNegativePosition, nil},
		{"zero denominator", 1, 0, duration.ErrInvalidArgument, []error{fraction.ErrZeroDenominator}},
		{"negative over zero", -1, 0, duration.ErrInvalidArgument, []error{fraction.ErrZeroDenominator}},
	} {
		p, err := duration.NewPosition(c.num, c.den)
		assertErr(t, err, c.want, c.causes...)
		assert.Equal(t, duration.Position{}, p, c.name)
	}

	p, err := duration.PositionFromFraction(frac(0, 5))
	require.NoError(t, err)
	assert.Equal(t, origin, p)

	p, err = duration.PositionFromFraction(frac(-6, -16))
	require.NoError(t, err)
	assert.Equal(t, pos(3, 8), p)

	for _, c := range []struct {
		f    fraction.Fraction
		want error
	}{
		{fraction.Fraction{}, duration.ErrInvalidArgument},
		{frac(-1, 4), duration.ErrNegativePosition},
		{frac(1, -4), duration.ErrNegativePosition},
	} {
		p, err := duration.PositionFromFraction(c.f)
		assertErr(t, err, c.want)
		assert.Equal(t, duration.Position{}, p)
	}
}

// TestPositionArithmetic covers O-13, KP-15, KP-16, KP-39, KP-52, T-6, and
// architecture 4.9.
func TestPositionArithmetic(t *testing.T) {
	t.Parallel()

	// KP-15: positions of a voice from the start.
	p := duration.Origin()

	var got []duration.Position

	for _, d := range []duration.Duration{dur(1, 4), dur(1, 8), dur(1, 8), dur(1, 2)} {
		p = p.MustAdd(d)
		got = append(got, p)
	}

	assert.Equal(t, []duration.Position{pos(1, 4), pos(3, 8), pos(1, 2), pos(1, 1)}, got)

	// KP-39.
	assert.Equal(t, pos(1, 4), duration.Origin().MustAdd(dur(1, 4)))

	// KP-52.
	assert.Equal(t, pos(3, 8), pos(1, 2).MustSub(dur(1, 8)))
	assert.Equal(t, duration.Origin(), pos(1, 4).MustSub(dur(1, 4)), "zero is a valid position")

	// KP-16.
	assert.Equal(t, pos(1, 4), pos(1, 4).MustOffsetTo(pos(1, 2)))
	assert.Equal(t, dur(1, 4), pos(1, 4).MustDurationTo(pos(1, 2)))
	assert.Equal(t, duration.Origin(), pos(1, 2).MustOffsetTo(pos(1, 2)))
	assert.Equal(t, pos(3, 4), duration.Origin().MustOffsetTo(pos(3, 4)))
	assert.Equal(t, dur(3, 4), duration.Origin().MustDurationTo(pos(3, 4)))

	// KP-30.
	for _, c := range []struct {
		name string
		run  func() (any, error)
		want error
		zero any
	}{
		{"Sub below zero", func() (any, error) { return pos(1, 4).Sub(dur(3, 8)) }, duration.ErrNegativePosition, duration.Position{}},
		{"OffsetTo backwards", func() (any, error) { return pos(1, 2).OffsetTo(pos(1, 4)) }, duration.ErrNegativePosition, duration.Position{}},
		{"DurationTo backwards", func() (any, error) { return pos(1, 2).DurationTo(pos(1, 4)) }, duration.ErrNotPositive, duration.Duration{}},
		{"DurationTo same", func() (any, error) { return pos(1, 2).DurationTo(pos(1, 2)) }, duration.ErrNotPositive, duration.Duration{}},
		{"DurationTo origin", func() (any, error) { return duration.Origin().DurationTo(duration.Origin()) }, duration.ErrNotPositive, duration.Duration{}},
	} {
		v, err := c.run()
		assertErr(t, err, c.want)
		assert.Equal(t, c.zero, v, c.name)
	}

	// Model: the operations agree with exact arithmetic.
	positions := []duration.Position{duration.Origin(), pos(1, 3), pos(1, 4), pos(3, 8), pos(7, 5), pos(1024, 1)}
	durations := []duration.Duration{dur(1, 1), dur(1, 12), dur(3, 8), dur(2, 15)}

	for _, a := range positions {
		ra := ratOfPosition(a)

		for _, d := range durations {
			rd := ratOfDuration(d)
			assert.Equal(t, posRat(new(big.Rat).Add(ra, rd)), a.MustAdd(d))

			sub, err := a.Sub(d)
			if diff := new(big.Rat).Sub(ra, rd); diff.Sign() >= 0 {
				require.NoError(t, err)
				assert.Equal(t, posRat(diff), sub)
			} else {
				assertErr(t, err, duration.ErrNegativePosition)
			}
		}

		for _, b := range positions {
			rb := ratOfPosition(b)
			diff := new(big.Rat).Sub(rb, ra)

			offset, err := a.OffsetTo(b)
			if diff.Sign() >= 0 {
				require.NoError(t, err)
				assert.Equal(t, posRat(diff), offset)
			} else {
				assertErr(t, err, duration.ErrNegativePosition)
			}

			length, err := a.DurationTo(b)
			if diff.Sign() > 0 {
				require.NoError(t, err)
				assert.Equal(t, durRat(diff), length)
			} else {
				assertErr(t, err, duration.ErrNotPositive)
			}

			assert.Equal(t, ra.Cmp(rb), a.Compare(b))
			assert.Equal(t, ra.Cmp(rb) == 0, a.Equal(b))
			assert.Equal(t, a == b, a.Equal(b))
		}
	}

	sorted := slices.Clone(positions)
	slices.SortFunc(sorted, duration.Position.Compare)
	assert.Equal(t, duration.Origin(), sorted[0], "the origin is first")
}

// TestPositionRange covers E-11 and architecture 4.9.
func TestPositionRange(t *testing.T) {
	t.Parallel()

	maxPos := posRat(ratInt(maxComponent()))
	a := posRat(new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, -1)))
	b := posRat(new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, 1)))
	db := durRat(new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, 1)))

	for _, c := range []struct {
		name string
		run  func() (any, error)
		zero any
	}{
		{"Add", func() (any, error) { return maxPos.Add(dur(1, 1)) }, duration.Position{}},
		{"Add denominators", func() (any, error) { return a.Add(db) }, duration.Position{}},
		{"Sub", func() (any, error) { return a.Sub(db) }, duration.Position{}},
		{"OffsetTo", func() (any, error) { return b.OffsetTo(a) }, duration.Position{}},
		{"DurationTo", func() (any, error) { return b.DurationTo(a) }, duration.Duration{}},
	} {
		v, err := c.run()
		assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, c.zero, v, c.name)
	}

	assert.Equal(t, maxPos, maxPos.MustSub(dur(1, 1)).MustAdd(dur(1, 1)))
}

// TestPositionString covers architecture 4.12: the reduced n/d.
func TestPositionString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "0/1", duration.Origin().String())
	assert.Equal(t, "0/1", pos(0, 9).String())
	assert.Equal(t, "3/8", pos(6, 16).String())
	assert.Equal(t, "2/1", pos(2, 1).String())
}
