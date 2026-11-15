package measure_test

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-2, the range check of Bars (architecture 4.6, "Iteration": the
// Е-10 criterion by runs of equal length, its guarantee and sufficiency,
// examples ПБ-1 to ПБ-6). Values near the range of common/fraction are
// built from *big.Rat; every expected value is taken from ПБ-1 to ПБ-6.

// pbD is d = 2⁶⁰ + 1 of ПБ-1 and ПБ-2.
const pbD = 1<<60 + 1

// pbA is a = 1/(2²⁵⁰ + 1) of ПБ-1 and ПБ-2.
func pbA() *big.Rat { return ratFrac(big.NewInt(1), new(big.Int).Add(pow2(250), big.NewInt(1))) }

// pbR is r = 2¹⁹⁴ + 1 of ПБ-3.
func pbR() *big.Int { return new(big.Int).Add(pow2(194), big.NewInt(1)) }

// pbMeter is the meter 1/d of ПБ-1 and ПБ-2.
func pbMeter() meter.Meter { return meter.MustNew(1, pbD) }

// pb1 is ПБ-1: the meter 1/d with the pickup a.
func pb1(t *testing.T) measure.Grid {
	t.Helper()

	return build(t, pbMeter(), func(g measure.Grid) (measure.Grid, error) { return g.WithPickup(durRat(pbA())) })
}

// pb2 is ПБ-2: the meter 1/d, bar 1 of the explicit length a from the start.
func pb2(t *testing.T) measure.Grid {
	t.Helper()

	return build(t, pbMeter(), func(g measure.Grid) (measure.Grid, error) {
		return g.WithLength(1, durRat(pbA()), measure.AlignStart)
	})
}

// pb3 is ПБ-3: the meter 1/4 with the pickup 1/r.
func pb3(t *testing.T) measure.Grid {
	t.Helper()

	return build(t, meter.MustNew(1, 4), func(g measure.Grid) (measure.Grid, error) {
		return g.WithPickup(durRat(ratFrac(big.NewInt(1), pbR())))
	})
}

// pb4S0 and pb4L are S₀ = 1/(5³⁵·7²⁸·11⁶) and L = b/(3⁵⁰·5³⁵·7²⁸) of ПБ-4.
func pb4S0() *big.Rat {
	d := new(big.Int).Mul(powInt(5, 35), powInt(7, 28))

	return ratFrac(big.NewInt(1), d.Mul(d, powInt(11, 6)))
}

func pb4L() *big.Rat {
	b, ok := new(big.Int).SetString("2655824496171892984250609480479768198769199113291", 10)
	if !ok {
		panic("pb4L")
	}

	d := new(big.Int).Mul(powInt(3, 50), powInt(5, 35))

	return ratFrac(b, d.Mul(d, powInt(7, 28)))
}

// pb4 is ПБ-4: 4/4, bar 1 of the length S₀, bars 2 and 3 of the length L.
func pb4(t *testing.T) measure.Grid {
	t.Helper()

	s0, l := durRat(pb4S0()), durRat(pb4L())

	return build(t, meter.FourFour(),
		func(g measure.Grid) (measure.Grid, error) { return g.WithLength(1, s0, measure.AlignStart) },
		func(g measure.Grid) (measure.Grid, error) { return g.WithLength(2, l, measure.AlignStart) },
		func(g measure.Grid) (measure.Grid, error) { return g.WithLength(3, l, measure.AlignStart) },
	)
}

// pb5L is L₅ = (2²⁵⁴ + 1)/2 of ПБ-5.
func pb5L() *big.Rat { return ratFrac(new(big.Int).Add(pow2(254), big.NewInt(1)), big.NewInt(2)) }

// pb5 is ПБ-5: 4/4, bars 1 and 2 of the length L₅.
func pb5(t *testing.T) measure.Grid {
	t.Helper()

	l := durRat(pb5L())

	return build(t, meter.FourFour(),
		func(g measure.Grid) (measure.Grid, error) { return g.WithLength(1, l, measure.AlignStart) },
		func(g measure.Grid) (measure.Grid, error) { return g.WithLength(2, l, measure.AlignStart) },
	)
}

// pb6 is ПБ-6 with a pickup: 4/4 with the pickup 1/4.
func pb6(t *testing.T) measure.Grid {
	t.Helper()

	return build(t, meter.FourFour(), withPickup("1/4"))
}

// pbOpen is the open 4/4 of ПБ-6.
func pbOpen(t *testing.T) measure.Grid {
	t.Helper()

	return build(t, meter.FourFour())
}

// powInt returns b^e as a big.Int.
func powInt(b, e int64) *big.Int { return new(big.Int).Exp(big.NewInt(b), big.NewInt(e), nil) }

// ratNum returns n/d as a big.Rat from decimal texts.
func ratNum(n, d string) *big.Rat {
	bn, ok1 := new(big.Int).SetString(n, 10)
	bd, ok2 := new(big.Int).SetString(d, 10)

	if !ok1 || !ok2 {
		panic("ratNum: " + n + "/" + d)
	}

	return ratFrac(bn, bd)
}

// assertBarsFail checks a refusal of Bars by the criterion of architecture
// 4.6: Е-10 with fraction.ErrOutOfRange and no other package sentinel, a nil
// sequence, and the panic of MustBars with the same error.
func assertBarsFail(t *testing.T, g measure.Grid, first, last int) {
	t.Helper()

	seq, err := g.Bars(first, last)
	assertErr(t, err, measure.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Nil(t, seq, "the sequence on an error is nil")

	panicErr := panicError(t, func() { g.MustBars(first, last) })
	if assert.Error(t, panicErr) && err != nil {
		assert.Equal(t, err.Error(), panicErr.Error(), "MustBars panics with the same error")

		for _, s := range allSentinels() {
			assert.Equal(t, errors.Is(err, s.err), errors.Is(panicErr, s.err), "errors.Is(err, %s)", s.name)
		}
	}
}

// assertBarsAgree iterates Bars(first, last) without an error and checks
// that each step equals Bar(k), in order from first; iteration stops after
// limit bars (0: all bars). It returns the bars seen.
func assertBarsAgree(t *testing.T, g measure.Grid, first, last, limit int) []measure.Bar {
	t.Helper()

	seq, err := g.Bars(first, last)
	require.NoError(t, err, "Bars(%d, %d)", first, last)
	require.NotNil(t, seq)

	var seen []measure.Bar

	k := first
	for b := range seq {
		want, err := g.Bar(k)
		if assert.NoError(t, err, "Bar(%d) of a range that Bars accepts", k) {
			assert.Equal(t, want, b, "step %d equals Bar(%d)", len(seen), k)
		}

		seen = append(seen, b)

		if (limit > 0 && len(seen) == limit) || k == last {
			break
		}

		k++
	}

	// last − first < limit − 1 avoids the overflow of last − first + 1 for
	// the range 0…math.MaxInt.
	if limit == 0 || last-first < limit-1 {
		assert.Len(t, seen, last-first+1, "the whole range is iterated")
	} else {
		assert.Len(t, seen, limit, "the iteration stops after %d bars", limit)
	}

	return seen
}

// TestBarsIterRangeRefused covers the refusals of ПБ-1 to ПБ-4: Е-10 by
// item 1 (an end of a run), item 2 (inside a run with a representable end),
// and the shift.
func TestBarsIterRangeRefused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        func(t *testing.T) measure.Grid
		first, last int
	}{
		{"ПБ-1 Bars(1, d) by item 2", pb1, 1, pbD},
		{"ПБ-1 Bars(1, 1) by item 1", pb1, 1, 1},
		{"ПБ-1 Bars(0, 0) by the shift", pb1, 0, 0},
		{"ПБ-2 Bars(2, d+1) by item 2", pb2, 2, pbD + 1},
		{"ПБ-2 Bars(1, d+1) by item 2", pb2, 1, pbD + 1},
		{"ПБ-3 Bars(1, 2^61+1) by item 1", pb3, 1, 1<<61 + 1},
		{"ПБ-3 Bars(1, 2^61+2) by item 2", pb3, 1, 1<<61 + 2},
		{"ПБ-3 Bars(1, 2^62) by item 2", pb3, 1, 1 << 62},
		{"ПБ-4 Bars(2, 3) false refusal by item 2", pb4, 2, 3},
		{"ПБ-4 Bars(1, 3) false refusal by item 2", pb4, 1, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertBarsFail(t, tt.grid(t), tt.first, tt.last)
		})
	}
}

// TestBarsIterRangeAccepted covers the successes of ПБ-2 to ПБ-6: no error,
// and each step equals Bar(k); large ranges are iterated with a break.
func TestBarsIterRangeAccepted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        func(t *testing.T) measure.Grid
		first, last int
		limit       int
	}{
		{"ПБ-2 Bars(1, 1)", pb2, 1, 1, 0},
		{"ПБ-3 Bars(1, 2^61)", pb3, 1, 1 << 61, 3},
		{"ПБ-3 Bars(0, 2^61)", pb3, 0, 1 << 61, 3},
		{"ПБ-3 a range inside 1…2^61", pb3, 1<<61 - 2, 1 << 61, 0},
		{"ПБ-4 Bars(2, 2)", pb4, 2, 2, 0},
		{"ПБ-4 Bars(3, 3)", pb4, 3, 3, 0},
		{"ПБ-4 Bars(1, 2)", pb4, 1, 2, 0},
		{"ПБ-5 Bars(1, 2) at the edge of the numerator", pb5, 1, 2, 0},
		{"ПБ-5 Bars(2, 2)", pb5, 2, 2, 0},
		{"ПБ-6 Bars(0, math.MaxInt) with a pickup", pb6, 0, math.MaxInt, 3},
		{"ПБ-6 Bars(1, math.MaxInt) open", pbOpen, 1, math.MaxInt, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertBarsAgree(t, tt.grid(t), tt.first, tt.last, tt.limit)
		})
	}
}

// TestBarsIterRangeSingleBar covers the exactness of the criterion for a run
// of one bar (architecture 4.6): Bars(k, k) fails if and only if Bar(k)
// fails, on the bars named in ПБ-1 to ПБ-5.
func TestBarsIterRangeSingleBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid func(t *testing.T) measure.Grid
		bar  int
		fail bool
	}{
		{"ПБ-1 bar 0", pb1, 0, true},
		{"ПБ-1 bar 1", pb1, 1, true},
		{"ПБ-2 bar 1", pb2, 1, false},
		{"ПБ-2 bar 2", pb2, 2, true},
		{"ПБ-3 bar 0", pb3, 0, false},
		{"ПБ-3 bar 2^61", pb3, 1 << 61, false},
		{"ПБ-3 bar 2^61+1", pb3, 1<<61 + 1, true},
		{"ПБ-3 bar 2^61+2", pb3, 1<<61 + 2, true},
		{"ПБ-4 bar 1", pb4, 1, false},
		{"ПБ-4 bar 2", pb4, 2, false},
		{"ПБ-4 bar 3", pb4, 3, false},
		{"ПБ-5 bar 1", pb5, 1, false},
		{"ПБ-5 bar 2", pb5, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := tt.grid(t)

			_, barErr := g.Bar(tt.bar)
			if tt.fail {
				assertErr(t, barErr, measure.ErrOutOfRange, fraction.ErrOutOfRange)
				assertBarsFail(t, g, tt.bar, tt.bar)

				return
			}

			require.NoError(t, barErr)
			assertBarsAgree(t, g, tt.bar, tt.bar, 0)
		})
	}
}

// TestBarsIterRangeValues covers the values named in ПБ-1 to ПБ-6: the ends
// of refused ranges are in range, Bar of the named bars, and the first
// bars of large ranges.
func TestBarsIterRangeValues(t *testing.T) {
	t.Parallel()

	t.Run("ПБ-1", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := pb1(t)

		end, err := g.BarEnd(pbD)
		require.NoError(t, err)
		assert.Equal(t, 0, ratP(end).Cmp(ratAdd(big.NewRat(1, 1), pbA())), "BarEnd(d) = (2²⁵⁰+2)/(2²⁵⁰+1)")

		start, err := g.BarStart(1)
		require.NoError(t, err)
		assert.Equal(t, 0, ratP(start).Cmp(pbA()), "BarStart(1) = a")
	})

	t.Run("ПБ-2", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := pb2(t)

		end, err := g.BarEnd(pbD + 1)
		require.NoError(t, err)
		assert.Equal(t, 0, ratP(end).Cmp(ratAdd(big.NewRat(1, 1), pbA())), "BarEnd(d + 1) = (2²⁵⁰+2)/(2²⁵⁰+1)")

		b, err := g.Bar(1)
		require.NoError(t, err)

		want := measure.Bar{
			Number:      1,
			Meter:       pbMeter(),
			Length:      durRat(pbA()),
			Start:       pos("0"),
			End:         posRat(pbA()),
			Kind:        measure.KindIncomplete,
			Pickup:      false,
			Explicit:    true,
			Alignment:   measure.AlignStart,
			Shift:       pos("0"),
			Marked:      true,
			MeterChange: false,
		}
		assert.Equal(t, want, b)

		seen := assertBarsAgree(t, g, 1, 1, 0)
		assert.Equal(t, []measure.Bar{want}, seen)
	})

	t.Run("ПБ-3", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := pb3(t)
		r := ratInt(pbR())
		inv := new(big.Rat).Inv(r)

		b, err := g.Bar(1 << 61)
		require.NoError(t, err)
		// (1 + 2⁵⁹·r)/r = 1/r + 2⁵⁹.
		assert.Equal(t, 0, ratP(b.End).Cmp(ratAdd(inv, ratInt(pow2(59)))), "end of bar 2⁶¹")

		end, err := g.BarEnd(1<<61 + 2)
		require.NoError(t, err)
		// (2²⁵⁴ + 2¹⁹⁴ + 2⁶⁰ + 3)/(2¹⁹⁵ + 2).
		want := ratFrac(
			new(big.Int).Add(new(big.Int).Add(pow2(254), pow2(194)), new(big.Int).Add(pow2(60), big.NewInt(3))),
			new(big.Int).Add(pow2(195), big.NewInt(2)))
		assert.Equal(t, 0, ratP(end).Cmp(want), "BarEnd(2⁶¹ + 2)")

		end, err = g.BarEnd(1 << 62)
		require.NoError(t, err)
		// (2²⁵⁴ + 2⁶⁰ + 1)/(2¹⁹⁴ + 1).
		want = ratFrac(new(big.Int).Add(pow2(254), new(big.Int).Add(pow2(60), big.NewInt(1))), pbR())
		assert.Equal(t, 0, ratP(end).Cmp(want), "BarEnd(2⁶²)")

		// The first bars of 1…2⁶¹ start at 1/r + j/4.
		seen := assertBarsAgree(t, g, 1, 1<<61, 3)
		if assert.Len(t, seen, 3) {
			for j, b := range seen {
				assert.Equal(t, 0, ratP(b.Start).Cmp(ratAdd(inv, big.NewRat(int64(j), 4))), "start of bar %d", j+1)
			}
		}
	})

	t.Run("ПБ-4", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := pb4(t)

		b2, err := g.Bar(2)
		require.NoError(t, err)

		b3, err := g.Bar(3)
		require.NoError(t, err)

		mid := ratNum("1616610262782161538566692261564", "585010912584312344346707148530270571875343078282503889")
		end := ratNum("20456925263964346000220136067951", "3701425381509371986654619362347875721752643585205078125")

		assert.Equal(t, 0, ratP(b2.Start).Cmp(pb4S0()), "start of bar 2 = S₀")
		assert.Equal(t, 0, ratP(b2.End).Cmp(mid), "end of bar 2 = S₀ + L")
		assert.Equal(t, 0, ratP(b3.Start).Cmp(mid), "start of bar 3 = S₀ + L")
		assert.Equal(t, 0, ratP(b3.End).Cmp(end), "end of bar 3 = S₀ + 2L")
		assert.Equal(t, 0, ratD(b2.Length).Cmp(pb4L()))
		assert.Equal(t, 0, ratD(b3.Length).Cmp(pb4L()))

		e3, err := g.BarEnd(3)
		require.NoError(t, err, "the end of a refused range is in range")
		assert.Equal(t, 0, ratP(e3).Cmp(end))
	})

	t.Run("ПБ-5", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := pb5(t)
		l := pb5L()

		want := []measure.Bar{
			{
				Number: 1, Meter: meter.FourFour(), Length: durRat(l), Start: pos("0"), End: posRat(l),
				Kind: measure.KindExtended, Explicit: true, Alignment: measure.AlignStart, Shift: pos("0"), Marked: true,
			},
			{
				Number: 2, Meter: meter.FourFour(), Length: durRat(l), Start: posRat(l), End: posRat(ratAdd(l, l)),
				Kind: measure.KindExtended, Explicit: true, Alignment: measure.AlignStart, Shift: pos("0"),
			},
		}

		assert.Equal(t, want, assertBarsAgree(t, g, 1, 2, 0))
		assert.Equal(t, 0, ratAdd(l, l).Cmp(ratInt(new(big.Int).Add(pow2(254), big.NewInt(1)))), "end of bar 2 = 2²⁵⁴ + 1")
	})

	t.Run("ПБ-6", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		seen := assertBarsAgree(t, pb6(t), 0, math.MaxInt, 3)

		var starts []duration.Position
		for _, b := range seen {
			starts = append(starts, b.Start)
		}

		assert.Equal(t, []duration.Position{pos("0"), pos("1/4"), pos("5/4")}, starts)
	})
}
