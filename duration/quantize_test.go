package duration_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oracleNode is the node of the grid by the accepted package common/rounding.
func oracleNode(t *testing.T, x *big.Rat, step fraction.Fraction, mode rounding.Mode) *big.Rat {
	t.Helper()

	_, v, err := mode.RatMultiple(x, ratOf(step))
	require.NoError(t, err)

	return v
}

// TestQuantizeExamples covers O-14, P-8, KP-19, KP-20, and KP-21.
func TestQuantizeExamples(t *testing.T) {
	t.Parallel()

	eighth := frac(1, 8)

	// KP-19.
	assert.Equal(t, pos(1, 4), pos(7, 24).MustQuantize(eighth, rounding.HalfEven))

	for _, c := range []struct {
		mode rounding.Mode
		want duration.Position
	}{
		{rounding.HalfUp, pos(3, 8)},
		{rounding.HalfDown, pos(1, 4)},
		{rounding.HalfEven, pos(1, 4)},
		{rounding.Down, pos(1, 4)},
		{rounding.Up, pos(3, 8)},
	} {
		got, err := pos(5, 16).Quantize(eighth, c.mode)
		require.NoError(t, err)
		assert.Equal(t, c.want, got, "5/16 %s", c.mode)
	}

	// A position may quantize to the origin.
	assert.Equal(t, duration.Origin(), pos(1, 32).MustQuantize(eighth, rounding.HalfEven))
	assert.Equal(t, duration.Origin(), duration.Origin().MustQuantize(eighth, rounding.Up))

	// KP-20.
	assert.Equal(t, dur(5, 16), dur(1, 3).MustQuantize(frac(1, 16), rounding.HalfEven))
	assert.Equal(t, dur(1, 8), dur(1, 32).MustQuantize(eighth, rounding.Up))

	for _, mode := range allModes() {
		assert.Equal(t, dur(1, 3), dur(1, 3).MustQuantize(frac(1, 12), mode), "KP-20 on the grid, %s", mode)
	}

	// KP-21.
	start, length, err := duration.QuantizeSpan(pos(1, 10), dur(1, 5), eighth, rounding.HalfEven)
	require.NoError(t, err)
	assert.Equal(t, pos(1, 8), start)
	assert.Equal(t, dur(1, 8), length)
	assert.Equal(t, pos(1, 4), start.MustAdd(length), "the end is 1/4")

	// An unreduced step is the same grid.
	assert.Equal(t, pos(1, 4), pos(7, 24).MustQuantize(frac(2, 16), rounding.HalfEven))
}

// TestQuantizeModel covers P-8 and architecture 4.10: results agree with
// RatMultiple of common/rounding on a grid from zero in all nine modes.
func TestQuantizeModel(t *testing.T) {
	t.Parallel()

	steps := []fraction.Fraction{frac(1, 1), frac(1, 8), frac(1, 12), frac(3, 16), frac(2, 3), frac(1, 1024), frac(5, 1)}

	var xs []*big.Rat
	for _, d := range []int64{1, 3, 7, 16, 24, 48, 100} {
		for n := int64(0); n <= 30; n++ {
			xs = append(xs, rat(n, d))
		}
	}

	for _, mode := range allModes() {
		for _, step := range steps {
			for _, x := range xs {
				want := oracleNode(t, x, step, mode)

				gotPos, err := posRat(x).Quantize(step, mode)
				require.NoError(t, err)
				assert.Equal(t, posRat(want), gotPos, "position %s step %s %s", x.RatString(), step, mode)

				if x.Sign() == 0 {
					continue
				}

				gotDur, err := durRat(x).Quantize(step, mode)
				if want.Sign() == 0 {
					assertErr(t, err, duration.ErrQuantizedToZero)
					assert.Equal(t, duration.Duration{}, gotDur)

					continue
				}

				require.NoError(t, err)
				assert.Equal(t, durRat(want), gotDur, "duration %s step %s %s", x.RatString(), step, mode)
			}
		}
	}
}

// TestQuantizeSpanModel covers P-8 and architecture 4.10: the start and the
// end are quantized on one grid; the length is end' − start'; E-8 when they
// meet.
func TestQuantizeSpanModel(t *testing.T) {
	t.Parallel()

	starts := []*big.Rat{rat(0, 1), rat(1, 10), rat(1, 3), rat(5, 16), rat(7, 4)}
	lengths := []*big.Rat{rat(1, 40), rat(1, 16), rat(1, 5), rat(1, 3), rat(2, 1)}
	steps := []fraction.Fraction{frac(1, 8), frac(1, 12), frac(1, 1)}

	for _, mode := range allModes() {
		for _, step := range steps {
			for _, s := range starts {
				for _, l := range lengths {
					wantStart := oracleNode(t, s, step, mode)
					wantEnd := oracleNode(t, new(big.Rat).Add(s, l), step, mode)
					wantLength := new(big.Rat).Sub(wantEnd, wantStart)

					start, length, err := duration.QuantizeSpan(posRat(s), durRat(l), step, mode)
					if wantLength.Sign() == 0 {
						assertErr(t, err, duration.ErrQuantizedToZero)
						assert.Equal(t, duration.Position{}, start)
						assert.Equal(t, duration.Duration{}, length)

						continue
					}

					require.NoError(t, err)
					assert.Equal(t, posRat(wantStart), start)
					assert.Equal(t, durRat(wantLength), length)
				}
			}
		}
	}

	// KP-31.
	start, length, err := duration.QuantizeSpan(pos(1, 10), dur(1, 40), frac(1, 8), rounding.HalfEven)
	assertErr(t, err, duration.ErrQuantizedToZero)
	assert.Equal(t, duration.Position{}, start)
	assert.Equal(t, duration.Duration{}, length)
}

// TestQuantizeOnGrid covers KP-41: a value on the grid is returned in every
// mode.
func TestQuantizeOnGrid(t *testing.T) {
	t.Parallel()

	for _, mode := range allModes() {
		for _, c := range []struct {
			x    *big.Rat
			step fraction.Fraction
		}{
			{rat(1, 4), frac(1, 8)},
			{rat(3, 8), frac(1, 8)},
			{rat(1, 3), frac(1, 12)},
			{rat(2, 1), frac(1, 1)},
			{rat(5, 1), frac(5, 1)},
		} {
			assert.Equal(t, durRat(c.x), durRat(c.x).MustQuantize(c.step, mode))
			assert.Equal(t, posRat(c.x), posRat(c.x).MustQuantize(c.step, mode))
		}

		assert.Equal(t, duration.Origin(), duration.Origin().MustQuantize(frac(1, 8), mode))
	}
}

// TestQuantizeModePairs covers P-8: for values >= 0 the modes coincide in
// pairs: TowardZero = Down, AwayFromZero = Up, HalfTowardZero = HalfDown,
// HalfAwayFromZero = HalfUp.
func TestQuantizeModePairs(t *testing.T) {
	t.Parallel()

	pairs := [][2]rounding.Mode{
		{rounding.TowardZero, rounding.Down},
		{rounding.AwayFromZero, rounding.Up},
		{rounding.HalfTowardZero, rounding.HalfDown},
		{rounding.HalfAwayFromZero, rounding.HalfUp},
	}

	for _, pair := range pairs {
		for _, step := range []fraction.Fraction{frac(1, 8), frac(1, 12), frac(2, 3)} {
			for d := int64(1); d <= 48; d++ {
				for n := int64(0); n <= 50; n++ {
					x := rat(n, d)

					a, errA := posRat(x).Quantize(step, pair[0])
					b, errB := posRat(x).Quantize(step, pair[1])
					require.NoError(t, errA)
					require.NoError(t, errB)
					assert.Equal(t, a, b, "%s: %s and %s", x.RatString(), pair[0], pair[1])

					if n == 0 {
						continue
					}

					da, errA := durRat(x).Quantize(step, pair[0])
					db, errB := durRat(x).Quantize(step, pair[1])
					assert.Equal(t, da, db)
					assert.Equal(t, errA == nil, errB == nil)
				}
			}
		}
	}
}

// TestQuantizeErrors covers KP-31, KP-32, and architecture 4.10: invalid
// steps are E-2 without dependency errors; invalid modes are E-2 with the
// error of common/rounding, also for a value on the grid; E-8; E-11.
func TestQuantizeErrors(t *testing.T) {
	t.Parallel()

	// KP-31.
	d, err := dur(1, 32).Quantize(frac(1, 8), rounding.HalfEven)
	assertErr(t, err, duration.ErrQuantizedToZero)
	assert.Equal(t, duration.Duration{}, d)

	// KP-32.
	badSteps := []fraction.Fraction{fraction.Zero(), frac(0, 5), frac(-1, 8), frac(1, -8), {}}
	badModes := []struct {
		mode  rounding.Mode
		cause error
	}{
		{rounding.Mode(0), rounding.ErrUnset},
		{rounding.Mode(10), rounding.ErrUnknownMode},
		{rounding.Mode(200), rounding.ErrUnknownMode},
	}

	for _, x := range []*big.Rat{rat(5, 16), rat(1, 4)} {
		for _, step := range badSteps {
			got, err := durRat(x).Quantize(step, rounding.HalfEven)
			assertErr(t, err, duration.ErrInvalidArgument)
			assert.Equal(t, duration.Duration{}, got)

			gotPos, err := posRat(x).Quantize(step, rounding.HalfEven)
			assertErr(t, err, duration.ErrInvalidArgument)
			assert.Equal(t, duration.Position{}, gotPos)

			start, length, err := duration.QuantizeSpan(posRat(x), dur(1, 4), step, rounding.HalfEven)
			assertErr(t, err, duration.ErrInvalidArgument)
			assert.Equal(t, duration.Position{}, start)
			assert.Equal(t, duration.Duration{}, length)
		}

		for _, m := range badModes {
			// 1/4 is on the grid of 1/8: the mode is still checked.
			got, err := durRat(x).Quantize(frac(1, 8), m.mode)
			assertErr(t, err, duration.ErrInvalidArgument, m.cause)
			assert.Equal(t, duration.Duration{}, got)

			gotPos, err := posRat(x).Quantize(frac(1, 8), m.mode)
			assertErr(t, err, duration.ErrInvalidArgument, m.cause)
			assert.Equal(t, duration.Position{}, gotPos)

			start, length, err := duration.QuantizeSpan(posRat(x), dur(1, 4), frac(1, 8), m.mode)
			assertErr(t, err, duration.ErrInvalidArgument, m.cause)
			assert.Equal(t, duration.Position{}, start)
			assert.Equal(t, duration.Duration{}, length)
		}
	}

	_, err = duration.Origin().Quantize(frac(1, 8), rounding.Mode(0))
	assertErr(t, err, duration.ErrInvalidArgument, rounding.ErrUnset)

	// E-11: rounding up at the edge of the range.
	maxRat := ratInt(maxComponent())

	got, err := durRat(maxRat).Quantize(frac(2, 1), rounding.Up)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, duration.Duration{}, got)

	gotPos, err := posRat(maxRat).Quantize(frac(2, 1), rounding.Up)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, duration.Position{}, gotPos)

	start, length, err := duration.QuantizeSpan(duration.Origin(), durRat(maxRat), frac(2, 1), rounding.Up)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Equal(t, duration.Position{}, start)
	assert.Equal(t, duration.Duration{}, length)

	// Rounding down at the edge is in range.
	assert.Equal(t, durRat(new(big.Rat).Sub(maxRat, big.NewRat(1, 1))), durRat(maxRat).MustQuantize(frac(2, 1), rounding.Down))
}

// TestQuantizeSpanExactSum covers architecture 4.10: start + length is
// computed exactly and is not checked against the range by itself.
func TestQuantizeSpanExactSum(t *testing.T) {
	t.Parallel()

	// 1/(2^200 − 1) + 1 + 1/(2^200 + 1) has a denominator near 2^400.
	start := posRat(new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, -1)))
	length := durRat(new(big.Rat).Add(big.NewRat(1, 1), new(big.Rat).SetFrac(big.NewInt(1), near2pow(200, 1))))

	gotStart, gotLength, err := duration.QuantizeSpan(start, length, frac(1, 8), rounding.HalfEven)
	require.NoError(t, err)
	assert.Equal(t, duration.Origin(), gotStart)
	assert.Equal(t, dur(1, 1), gotLength)

	// The sum exceeds the range by its numerator, the nodes do not.
	maxRat := ratInt(maxComponent())
	// Start 2^255 − 3, length 3, step 3: the end 2^255 is out of range,
	// its node 2^255 − 2 is not.
	step := frac(3, 1)
	gotStart, gotLength, err = duration.QuantizeSpan(posRat(new(big.Rat).Sub(maxRat, big.NewRat(2, 1))), dur(3, 1), step, rounding.Down)
	require.NoError(t, err)

	wantStart := oracleNode(t, new(big.Rat).Sub(maxRat, big.NewRat(2, 1)), step, rounding.Down)
	wantEnd := oracleNode(t, new(big.Rat).Add(maxRat, big.NewRat(1, 1)), step, rounding.Down)
	require.Equal(t, -1, wantEnd.Cmp(new(big.Rat).Add(maxRat, big.NewRat(1, 1))))
	assert.Equal(t, posRat(wantStart), gotStart)
	assert.Equal(t, durRat(new(big.Rat).Sub(wantEnd, wantStart)), gotLength)
}
