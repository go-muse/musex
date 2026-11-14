package tempo_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBarUnits covers С-14 and architecture 4.9: r = L/u, L = u·r, u = L/r.
func TestBarUnits(t *testing.T) {
	t.Parallel()

	got, err := tq(60).BarUnits(dur(1, 1))
	require.NoError(t, err)
	assertFrac(t, fi(4), got)
	assertFrac(t, fi(2), tm(dq, fi(60)).MustBarUnits(dur(3, 4)))

	got, err = tempo.BarUnits(dur(1, 1), dur(1, 4))
	require.NoError(t, err)
	assertFrac(t, fi(4), got)
	assertFrac(t, fi(2), tempo.MustBarUnits(dur(3, 4), dur(3, 8)))
	assertFrac(t, fr(5, 2), tempo.MustBarUnits(dur(5, 8), dur(1, 4)))

	l, err := tempo.BarLength(dur(1, 4), fi(2))
	require.NoError(t, err)
	assertDur(t, dur(1, 2), l)
	assertDur(t, dur(1, 4), tempo.MustBarLength(dur(1, 8), fi(2)))
	assertDur(t, dur(5, 8), tempo.MustBarLength(dur(1, 4), fr(5, 2)))

	u, err := tempo.UnitLength(dur(1, 2), fi(2))
	require.NoError(t, err)
	assertDur(t, dur(1, 4), u)
	assertDur(t, dur(3, 8), tempo.MustUnitLength(dur(3, 4), fi(2)))
}

// TestFormulaWithTempo covers С-15, С-16 and architecture 4.9: bars, time,
// tempo, and r with L from a tempo; bars per minute.
func TestFormulaWithTempo(t *testing.T) {
	t.Parallel()

	t.Run("С-15 bars", func(t *testing.T) {
		t.Parallel()

		got, err := tq(120).Bars(dur(1, 1), sec(120, 1))
		require.NoError(t, err)
		assertFrac(t, fi(60), got)
		assertFrac(t, fi(30), tm(dq, fi(60)).MustBars(dur(3, 4), sec(60, 1)))
		assertFrac(t, fr(100, 3), tq(100).MustBars(dur(3, 4), sec(60, 1)), "time ends off the bar line")
	})

	t.Run("С-15 time of bars", func(t *testing.T) {
		t.Parallel()

		got, err := tq(120).SecondsOfBars(dur(1, 1), fi(60))
		require.NoError(t, err)
		assert.Equal(t, sec(120, 1), got)
		assert.Equal(t, sec(60, 1), tq(100).MustSecondsOfBars(dur(3, 4), fr(100, 3)))
	})

	t.Run("С-15 tempo from bars", func(t *testing.T) {
		t.Parallel()

		got, err := tempo.FromBars(q, dur(1, 2), fi(30), sec(60, 1))
		require.NoError(t, err)
		assert.Equal(t, tq(60), got)
		assert.Equal(t, tq(120), tempo.MustFromBars(q, dur(1, 1), fi(60), sec(120, 1)))
		assert.Equal(t, tm(dq, fi(60)), tempo.MustFromBars(dq, dur(3, 4), fi(30), sec(60, 1)))
	})

	t.Run("С-15 bar for bars and time", func(t *testing.T) {
		t.Parallel()

		r, l, err := tq(60).BarFor(fi(30), sec(60, 1))
		require.NoError(t, err)
		assertFrac(t, fi(2), r)
		assertDur(t, dur(1, 2), l)

		r, l = tq(100).MustBarFor(fr(100, 3), sec(60, 1))
		assertFrac(t, fi(3), r)
		assertDur(t, dur(3, 4), l)
	})

	t.Run("С-16 bars per minute", func(t *testing.T) {
		t.Parallel()

		got, err := tq(180).BarsPerMinute(dur(3, 4))
		require.NoError(t, err)
		assertFrac(t, fi(60), got)

		x, err := tempo.FromBarsPerMinute(q, dur(3, 4), fi(60))
		require.NoError(t, err)
		assert.Equal(t, tq(180), x)
		assert.Equal(t, tm(dh, fi(60)), tempo.MustFromBarsPerMinute(dh, dur(3, 4), fi(60)))
	})
}

// TestFormulaNumbers covers С-15 and architecture 4.9: any three of n, r, N,
// and time give the fourth.
func TestFormulaNumbers(t *testing.T) {
	t.Parallel()

	got, err := tempo.SolveBarUnits(fi(30), fi(60), sec(60, 1))
	require.NoError(t, err)
	assertFrac(t, fi(2), got, "С-15 n = 30, t = 1 min, N = 60 -> r = 2")

	got, err = tempo.SolveBars(fi(2), fi(60), sec(60, 1))
	require.NoError(t, err)
	assertFrac(t, fi(30), got)

	got, err = tempo.SolveBPM(fi(30), fi(2), sec(60, 1))
	require.NoError(t, err)
	assertFrac(t, fi(60), got)

	s, err := tempo.SolveSeconds(fi(30), fi(2), fi(60))
	require.NoError(t, err)
	assert.Equal(t, sec(60, 1), s)

	assert.Equal(t, sec(120, 1), tempo.MustSolveSeconds(fi(60), fi(4), fi(120)), "С-15 4/4 at ♩ = 120, n = 60")
	assertFrac(t, fr(100, 3), tempo.MustSolveBars(fi(3), fi(100), sec(60, 1)))
	assertFrac(t, fr(1, 2), tempo.MustSolveBPM(fr(1, 4), fi(2), sec(60, 1)))
	assertFrac(t, fr(7, 3), tempo.MustSolveBarUnits(fi(30), fr(70, 1), sec(60, 1)))
}

// TestFormulaErrors covers С-34, С-43 and architecture 4.3, 4.9: values <= 0
// give Е-4 (bpm in Solve* included, not Е-3), time 0 gives Е-4, results out
// of range give Е-11 with fraction.ErrOutOfRange, and the order of checks.
func TestFormulaErrors(t *testing.T) {
	t.Parallel()

	zero := tempo.ZeroSeconds()
	m60 := sec(60, 1)
	maxDur := durF(maxF())
	tinyDur := durF(invMaxF())

	type call = func() (any, error)

	tests := []struct {
		name string
		call call
		want errClass
		zero any
	}{
		// Е-4 of every formula operation.
		{"BarLength r = 0", func() (any, error) { return tempo.BarLength(dur(1, 4), fi(0)) }, noSolution, duration.Duration{}},
		{"BarLength r < 0", func() (any, error) { return tempo.BarLength(dur(1, 4), fi(-2)) }, noSolution, duration.Duration{}},
		{"UnitLength r = 0", func() (any, error) { return tempo.UnitLength(dur(1, 2), fi(0)) }, noSolution, duration.Duration{}},
		{"Bars time 0", func() (any, error) { return tq(60).Bars(dur(1, 1), zero) }, noSolution, fraction.Fraction{}},
		{"SecondsOfBars n = 0", func() (any, error) { return tq(60).SecondsOfBars(dur(1, 1), fi(0)) }, noSolution, tempo.Seconds{}},
		{"SecondsOfBars n < 0", func() (any, error) { return tq(60).SecondsOfBars(dur(1, 1), fi(-1)) }, noSolution, tempo.Seconds{}},
		{"FromBars n = 0", func() (any, error) { return tempo.FromBars(q, dur(1, 1), fi(0), m60) }, noSolution, tempo.Tempo{}},
		{"FromBars time 0", func() (any, error) { return tempo.FromBars(q, dur(1, 1), fi(30), zero) }, noSolution, tempo.Tempo{}},
		{"С-34 FromBarsPerMinute 0", func() (any, error) { return tempo.FromBarsPerMinute(q, dur(3, 4), fi(0)) }, noSolution, tempo.Tempo{}},
		{"FromBarsPerMinute < 0", func() (any, error) { return tempo.FromBarsPerMinute(q, dur(3, 4), fi(-60)) }, noSolution, tempo.Tempo{}},
		{"С-34 SolveBars r = 0", func() (any, error) { return tempo.SolveBars(fi(0), fi(60), m60) }, noSolution, fraction.Fraction{}},
		{"С-34 SolveBars N = 0", func() (any, error) { return tempo.SolveBars(fi(2), fi(0), m60) }, noSolution, fraction.Fraction{}},
		{"С-34 SolveBars N < 0", func() (any, error) { return tempo.SolveBars(fi(2), fi(-60), m60) }, noSolution, fraction.Fraction{}},
		{"С-34 SolveBars time 0", func() (any, error) { return tempo.SolveBars(fi(2), fi(60), zero) }, noSolution, fraction.Fraction{}},
		{"С-34 SolveBarUnits n = 0", func() (any, error) { return tempo.SolveBarUnits(fi(0), fi(60), m60) }, noSolution, fraction.Fraction{}},
		{"SolveBarUnits N = 0", func() (any, error) { return tempo.SolveBarUnits(fi(30), fi(0), m60) }, noSolution, fraction.Fraction{}},
		{"SolveBarUnits time 0", func() (any, error) { return tempo.SolveBarUnits(fi(30), fi(60), zero) }, noSolution, fraction.Fraction{}},
		{"SolveBPM n < 0", func() (any, error) { return tempo.SolveBPM(fi(-30), fi(2), m60) }, noSolution, fraction.Fraction{}},
		{"SolveBPM r = 0", func() (any, error) { return tempo.SolveBPM(fi(30), fi(0), m60) }, noSolution, fraction.Fraction{}},
		{"SolveBPM time 0", func() (any, error) { return tempo.SolveBPM(fi(30), fi(2), zero) }, noSolution, fraction.Fraction{}},
		{"SolveSeconds n = 0", func() (any, error) { return tempo.SolveSeconds(fi(0), fi(2), fi(60)) }, noSolution, tempo.Seconds{}},
		{"SolveSeconds r = 0", func() (any, error) { return tempo.SolveSeconds(fi(30), fi(0), fi(60)) }, noSolution, tempo.Seconds{}},
		{"SolveSeconds N = 0", func() (any, error) { return tempo.SolveSeconds(fi(30), fi(2), fi(0)) }, noSolution, tempo.Seconds{}},
		// Е-2 for unset values of other packages; Е-1 for unset Seconds.
		{"BarUnits unset", func() (any, error) { return tempo.BarUnits(duration.Duration{}, dur(1, 4)) }, argClass, fraction.Fraction{}},
		{"BarUnits unset unit", func() (any, error) { return tempo.BarUnits(dur(1, 1), duration.Duration{}) }, argClass, fraction.Fraction{}},
		{"BarLength unset unit", func() (any, error) { return tempo.BarLength(duration.Duration{}, fi(2)) }, argClass, duration.Duration{}},
		{"BarLength unset r", func() (any, error) { return tempo.BarLength(dur(1, 4), fraction.Fraction{}) }, argClass, duration.Duration{}},
		{"UnitLength unset r", func() (any, error) { return tempo.UnitLength(dur(1, 2), fraction.Fraction{}) }, argClass, duration.Duration{}},
		{"Tempo.BarUnits unset", func() (any, error) { return tq(60).BarUnits(duration.Duration{}) }, argClass, fraction.Fraction{}},
		{"BarsPerMinute unset", func() (any, error) { return tq(60).BarsPerMinute(duration.Duration{}) }, argClass, fraction.Fraction{}},
		{"Bars unset time", func() (any, error) { return tq(60).Bars(dur(1, 1), tempo.Seconds{}) }, unsetClass, fraction.Fraction{}},
		{"SolveSeconds unset", func() (any, error) { return tempo.SolveSeconds(fi(30), fraction.Fraction{}, fi(60)) }, argClass, tempo.Seconds{}},
		{"SolveBPM unset time", func() (any, error) { return tempo.SolveBPM(fi(30), fi(2), tempo.Seconds{}) }, unsetClass, fraction.Fraction{}},
		// Е-11.
		{"С-43 BarUnits", func() (any, error) { return tempo.BarUnits(maxDur, tinyDur) }, fracRange, fraction.Fraction{}},
		{"С-43 Tempo.BarUnits", func() (any, error) { return tq(60).BarUnits(maxDur) }, fracRange, fraction.Fraction{}},
		{"С-43 BarLength", func() (any, error) { return tempo.BarLength(dur(2, 1), maxF()) }, fracRange, duration.Duration{}},
		{"С-43 UnitLength", func() (any, error) { return tempo.UnitLength(tinyDur, maxF()) }, fracRange, duration.Duration{}},
		{"С-43 Bars", func() (any, error) { return tm(q, maxF()).Bars(tinyDur, m60) }, fracRange, fraction.Fraction{}},
		{"С-43 SecondsOfBars", func() (any, error) { return tm(q, invMaxF()).SecondsOfBars(dur(1, 1), maxF()) }, fracRange, tempo.Seconds{}},
		{"С-43 FromBars", func() (any, error) { return tempo.FromBars(q, maxDur, maxF(), m60) }, fracRange, tempo.Tempo{}},
		{"С-43 BarFor", func() (any, error) { r, l, err := tq(60).BarFor(invMaxF(), secF(maxF())); return []any{r, l}, err },
			fracRange, []any{fraction.Fraction{}, duration.Duration{}}},
		{"С-43 BarFor bar length", func() (any, error) {
			r, l, err := tm(breve, maxF()).BarFor(fi(1), m60)
			return []any{r, l}, err
		}, fracRange, []any{fraction.Fraction{}, duration.Duration{}}},
		{"С-43 BarsPerMinute", func() (any, error) { return tm(q, maxF()).BarsPerMinute(tinyDur) }, fracRange, fraction.Fraction{}},
		{"С-43 FromBarsPerMinute", func() (any, error) { return tempo.FromBarsPerMinute(q, maxDur, maxF()) }, fracRange, tempo.Tempo{}},
		{"С-43 SolveBars", func() (any, error) { return tempo.SolveBars(invMaxF(), maxF(), m60) }, fracRange, fraction.Fraction{}},
		{"С-43 SolveBarUnits", func() (any, error) { return tempo.SolveBarUnits(invMaxF(), maxF(), m60) }, fracRange, fraction.Fraction{}},
		{"С-43 SolveBPM", func() (any, error) { return tempo.SolveBPM(maxF(), maxF(), m60) }, fracRange, fraction.Fraction{}},
		{"С-43 SolveSeconds", func() (any, error) { return tempo.SolveSeconds(maxF(), maxF(), fi(1)) }, fracRange, tempo.Seconds{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			tt.want.check(t, err)
			assert.Equal(t, tt.zero, got)
		})
	}
}
