package measure_test

import (
	"errors"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-9, consistency (КП-61; Т-16; architecture 4.7, 4.8, 4.10, 4.11,
// 4.14…4.16) on the grids of the examples and on grids built by a
// deterministic generator.

// ratD, ratP, and ratF are exact values of a duration, a position, and a
// fraction.
func ratD(d duration.Duration) *big.Rat {
	r, _ := d.MustFraction().BigRat()

	return r
}

func ratP(p duration.Position) *big.Rat {
	r, _ := p.MustFraction().BigRat()

	return r
}

func ratF(f fraction.Fraction) *big.Rat {
	r, _ := f.BigRat()

	return r
}

func ratAdd(a, b *big.Rat) *big.Rat { return new(big.Rat).Add(a, b) }

func ratSub(a, b *big.Rat) *big.Rat { return new(big.Rat).Sub(a, b) }

func ratMul(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }

func ratQuo(a, b *big.Rat) *big.Rat { return new(big.Rat).Quo(a, b) }

// generatedGrids builds grids with a pickup, time signatures, explicit
// lengths of both alignments, and open and finite counts by New and the
// operations of architecture 4.4, from a fixed seed.
func generatedGrids(t *testing.T) []measure.Grid {
	t.Helper()

	rng := rand.New(rand.NewPCG(28, 2))
	meters := mts("4/4", "3/4", "6/8", "2/4+3/8", "5/8", "7/8[3+2+2]", "C", "4/3", "3/8", "2/2", "(3+2+3)/8", "5/8[2+3]")

	out := make([]measure.Grid, 0, 48)

	for range 48 {
		g, err := measure.New(meters[rng.IntN(len(meters))])
		require.NoError(t, err)

		if rng.IntN(2) == 0 {
			m, err := g.BarMeter(1)
			require.NoError(t, err)

			size := ratD(m.MustDuration())
			a := ratMul(size, big.NewRat(int64(1+rng.IntN(8)), 8))
			g, err = g.WithPickup(durRat(a))
			require.NoError(t, err)
		}

		last := 1

		for range rng.IntN(4) {
			bar := 2 + rng.IntN(14)
			g, err = g.WithMark(bar, meters[rng.IntN(len(meters))])
			require.NoError(t, err)

			last = max(last, bar)
		}

		for range rng.IntN(4) {
			bar := 1 + rng.IntN(16)

			m, err := g.BarMeter(bar)
			require.NoError(t, err)

			l := durRat(ratMul(ratD(m.MustDuration()), big.NewRat(int64(1+rng.IntN(6)), 4)))

			align := measure.AlignStart
			if rng.IntN(2) == 0 {
				align = measure.AlignEnd
			}

			next, err := g.WithLength(bar, l, align)
			if errors.Is(err, measure.ErrLongerThanMeter) {
				next, err = g.WithLength(bar, l, measure.AlignStart)
			}

			require.NoError(t, err)

			g = next
			last = max(last, bar)
		}

		if rng.IntN(2) == 0 {
			g, err = g.WithBarCount(last + rng.IntN(4))
			require.NoError(t, err)
		}

		out = append(out, g)
	}

	return out
}

// exampleGrids are the grids of the examples of the other groups.
func exampleGrids(t *testing.T) []measure.Grid {
	t.Helper()

	texts := append([]string{}, canonicalGrids...)
	for _, c := range equalCases() {
		texts = append(texts, c.a, c.b)
	}

	texts = append(texts, "4/4;pickup:1/4;3:3/4;4:length:1/2", "4/4;2:length:3/2", "4/4;3:length:1/4:end", "6/8;pickup:1/2")

	out := make([]measure.Grid, len(texts))
	for i, s := range texts {
		out[i] = gridOf(t, s)
	}

	return out
}

// TestConsistency covers КП-61 and the properties of
// implementation-plan, "Consistency", on each grid.
func TestConsistency(t *testing.T) {
	t.Parallel()

	var grids []measure.Grid

	func() {
		defer guard(t)

		grids = append(exampleGrids(t), generatedGrids(t)...)
	}()

	for _, g := range grids {
		t.Run(g.String(), func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			checkText(t, g)
			checkBars(t, g)
			checkPositions(t, g)
			checkCounts(t, g)
			checkSplit(t, g)
		})
	}

	t.Run("equalities imply each other", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		for _, a := range grids {
			for _, b := range grids {
				markup, err := a.EqualMarkup(b)
				require.NoError(t, err)

				lengths, err := a.EqualBarLengths(b)
				require.NoError(t, err)

				if a.Equal(b) {
					assert.True(t, markup, "%s and %s: equal by record, not by markup", a, b)
				}

				if markup {
					assert.True(t, lengths, "%s and %s: equal by markup, not by bar lengths", a, b)
				}
			}
		}
	})
}

// lastChecked is the last bar checked in a grid: N, but not after bar 18.
func lastChecked(t *testing.T, g measure.Grid) int {
	t.Helper()

	n, finite, err := g.BarCount()
	require.NoError(t, err)

	if finite {
		return min(n, 18)
	}

	return 18
}

// checkText covers Parse(String()) and Simplify.
func checkText(t *testing.T, g measure.Grid) {
	t.Helper()

	p, err := measure.Parse(g.String())
	require.NoError(t, err)
	assert.True(t, p.Equal(g), "Parse(String()) equals by record")

	s, err := g.Simplify()
	require.NoError(t, err)

	markup, err := s.EqualMarkup(g)
	require.NoError(t, err)
	assert.True(t, markup, "Simplify keeps the markup")
}

// checkBars covers the sequence of bars, BarEnd(k) = BarStart(k + 1), and
// RangeLength(f, l) = BarEnd(l) − BarStart(f).
func checkBars(t *testing.T, g measure.Grid) {
	t.Helper()

	first, err := g.FirstBar()
	require.NoError(t, err)

	last := lastChecked(t, g)

	seq, err := g.Bars(first, last)
	require.NoError(t, err)

	bars := collect(seq)
	require.Len(t, bars, last-first+1)

	for i, b := range bars {
		k := first + i

		want, err := g.Bar(k)
		require.NoError(t, err)
		assert.Equal(t, want, b, "bar %d", k)

		assert.Equal(t, 0, ratAdd(ratP(b.Start), ratD(b.Length)).Cmp(ratP(b.End)), "end = start + length")

		if k < last {
			next, err := g.BarStart(k + 1)
			require.NoError(t, err)
			assert.Equal(t, b.End, next, "BarEnd(%d) = BarStart(%d)", k, k+1)
		}

		for f := first; f <= k; f++ {
			l, err := g.RangeLength(f, k)
			require.NoError(t, err)

			s, err := g.BarStart(f)
			require.NoError(t, err)
			assert.Equal(t, 0, ratD(l).Cmp(ratSub(ratP(b.End), ratP(s))), "RangeLength(%d, %d)", f, k)
		}
	}
}

// offsetsOf returns offsets in a bar of length l: 0, l/3, l/2, 7l/8.
func offsetsOf(l duration.Duration) []*big.Rat {
	r := ratD(l)

	return []*big.Rat{new(big.Rat), ratMul(r, big.NewRat(1, 3)), ratMul(r, big.NewRat(1, 2)), ratMul(r, big.NewRat(7, 8))}
}

// checkPositions covers PositionOf(Locate(p)) = p, Locate(PositionOf(a)) = a,
// Contains, BarAt, Remaining, OnBarline, Elapsed, Normalize, and the place in
// the meter.
func checkPositions(t *testing.T, g measure.Grid) {
	t.Helper()

	first, err := g.FirstBar()
	require.NoError(t, err)

	firstStart, err := g.BarStart(first)
	require.NoError(t, err)

	for k := first; k <= lastChecked(t, g); k++ {
		b, err := g.Bar(k)
		require.NoError(t, err)

		size := ratD(b.Meter.MustDuration())

		for _, off := range offsetsOf(b.Length) {
			p := posRat(ratAdd(ratP(b.Start), off))
			want := measure.MustNewAddress(k, posRat(off))

			a, err := g.Locate(p)
			require.NoError(t, err)
			assert.Equal(t, want, a, "Locate(%s)", p)

			back, err := g.PositionOf(want)
			require.NoError(t, err)
			assert.Equal(t, p, back, "PositionOf(%s)", want)

			ok, err := g.Contains(want)
			require.NoError(t, err)
			assert.True(t, ok)

			at, err := g.BarAt(p)
			require.NoError(t, err)
			assert.Equal(t, b, at)

			r, err := g.Remaining(p)
			require.NoError(t, err)
			assert.Equal(t, 0, ratD(r).Cmp(ratSub(ratD(b.Length), off)), "Remaining(%s)", p)

			on, err := g.OnBarline(p)
			require.NoError(t, err)
			assert.Equal(t, off.Sign() == 0, on, "OnBarline(%s)", p)

			el, err := g.Elapsed(want)
			require.NoError(t, err)
			assert.Equal(t, 0, ratF(el).Cmp(ratQuo(off, ratD(b.Length))), "Elapsed(%s)", want)

			n, err := g.Normalize(measure.MustNewAddress(first, posRat(ratSub(ratP(p), ratP(firstStart)))))
			require.NoError(t, err)
			assert.Equal(t, want, n, "Normalize from bar %d", first)

			m, u, err := g.MetricPlace(want)
			place := ratAdd(ratP(b.Shift), off)

			if place.Cmp(size) < 0 {
				require.NoError(t, err)
				assert.Equal(t, b.Meter, m)
				assert.Equal(t, 0, ratP(u).Cmp(place), "MetricPlace(%s)", want)
			} else {
				assertErr(t, err, measure.ErrUndefined)
			}
		}

		ok, err := g.Contains(measure.MustNewAddress(k, posRat(ratD(b.Length))))
		require.NoError(t, err)
		assert.False(t, ok, "an offset equal to the length is not valid")
	}
}

// checkCounts covers the inverse unit counts of both ways and MEI tstamp
// in its domain.
func checkCounts(t *testing.T, g measure.Grid) {
	t.Helper()

	first, err := g.FirstBar()
	require.NoError(t, err)

	for k := first; k <= lastChecked(t, g); k++ {
		b, err := g.Bar(k)
		require.NoError(t, err)

		common, err := b.Meter.HasCommonDenominator()
		require.NoError(t, err)

		l := ratD(b.Length)

		for _, tr := range []*big.Rat{new(big.Rat), ratMul(l, big.NewRat(1, 2)), ratMul(l, big.NewRat(2, 3)), l} {
			tv := posRat(tr)

			c, err := g.CountAt(k, tv)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, ratF(c).Cmp(big.NewRat(1, 1)), 0, "a count is at least 1")

			back, err := g.CountOffset(k, c)
			require.NoError(t, err)
			assert.Equal(t, tv, back, "CountOffset(CountAt(%d, %s))", k, tv)

			mc, err := g.MetricCountAt(k, tv)
			require.NoError(t, err)

			mback, err := g.MetricCountOffset(k, mc)
			require.NoError(t, err)
			assert.Equal(t, tv, mback, "MetricCountOffset(MetricCountAt(%d, %s))", k, tv)

			if b.Shift.Equal(duration.Origin()) {
				assert.Equal(t, c, mc, "the two ways agree without a shift")
			}

			ts, err := g.TstampAt(k, tv)
			if b.Kind == measure.KindFull && common {
				require.NoError(t, err)
				assert.Equal(t, c, ts, "MEI tstamp equals the unit count in its domain")

				tsBack, err := g.TstampOffset(k, ts)
				require.NoError(t, err)
				assert.Equal(t, tv, tsBack)
			} else {
				assertErr(t, err, measure.ErrUndefined)
			}
		}
	}
}

// checkSplit covers the properties of a split at every level.
func checkSplit(t *testing.T, g measure.Grid) {
	t.Helper()

	first, err := g.FirstBar()
	require.NoError(t, err)

	last := lastChecked(t, g)

	b, err := g.Bar(first)
	require.NoError(t, err)

	whole, err := g.RangeLength(first, last)
	require.NoError(t, err)

	start := posRat(ratAdd(ratP(b.Start), ratMul(ratD(b.Length), big.NewRat(1, 3))))
	value := durRat(ratMul(ratD(whole), big.NewRat(1, 2)))

	for _, level := range []measure.Level{measure.LevelBars, measure.LevelGroups, measure.LevelBeats, measure.LevelUnits} {
		parts, err := g.SplitBy(start, value, level)
		if err != nil && (level == measure.LevelGroups || level == measure.LevelBeats) {
			assertErr(t, err, measure.ErrGroupingRequired, meter.ErrGroupingRequired)

			continue
		}

		require.NoError(t, err, "level %s", level)
		assertParts(t, g, start, value, parts)
	}
}
