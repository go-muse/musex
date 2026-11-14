package tempo_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew covers С-1, С-2, С-50, С-51 and architecture 4.6, "Создание",
// "Чтение": unit as given, bpm stored reduced, the default unit of
// NewQuarter, speed and quarters per minute.
func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		unit       duration.Written
		bpm        fraction.Fraction
		wantBPM    fraction.Fraction
		unitLen    duration.Duration
		speed, qpm fraction.Fraction
	}{
		{"С-1 ♩ = 60", q, fi(60), fi(60), dur(1, 4), fi(15), fi(60)},
		{"С-1 С-2 ♩. = 60", dq, fi(60), fi(60), dur(3, 8), fr(45, 2), fi(90)},
		{"С-1 ♩ = 185/2", q, fr(185, 2), fr(185, 2), dur(1, 4), fr(185, 8), fr(185, 2)},
		{"С-3 ♩ = 120/2 is stored as 60", q, fraction.MustNew(120, 2), fi(60), dur(1, 4), fi(15), fi(60)},
		{"♪ = 180", e, fi(180), fi(180), dur(1, 8), fr(45, 2), fi(90)},
		{"С-50 ♩ = 1/1000", q, fr(1, 1000), fr(1, 1000), dur(1, 4), fr(1, 4000), fr(1, 1000)},
		{"С-50 ♩ = 100000", q, fi(100000), fi(100000), dur(1, 4), fi(25000), fi(100000)},
		{"С-51 eighth in a triplet = 360", e3, fi(360), fi(360), dur(1, 12), fi(30), fi(120)},
		{"𝅗𝅥. = 60", dh, fi(60), fi(60), dur(3, 4), fi(45), fi(180)},
		{"breve = 1/3", breve, fr(1, 3), fr(1, 3), dur(2, 1), fr(2, 3), fr(8, 3)},
		{"bpm at the range limit", q, maxF(), maxF(), dur(1, 4), fbig(fraction.MaxComponentBig(), big.NewInt(4)), maxF()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tempo.New(tt.unit, tt.bpm)
			require.NoError(t, err)
			assert.True(t, got.IsSet())

			assert.Equal(t, tt.unit, got.MustUnit())
			assertDur(t, tt.unitLen, got.MustUnitDuration())

			bpm := got.MustBPM()
			assert.Equal(t, tt.wantBPM, bpm, "BPM is reduced: %s", bpm)

			assertFrac(t, tt.speed, got.MustSpeed(), "Speed")
			assertFrac(t, tt.qpm, got.MustQuartersPerMinute(), "QuartersPerMinute")
		})
	}

	t.Run("С-1 NewQuarter is the quarter unit", func(t *testing.T) {
		t.Parallel()

		got, err := tempo.NewQuarter(fi(60))
		require.NoError(t, err)
		assert.Equal(t, tm(q, fi(60)), got)
		assert.Equal(t, duration.Quarter.MustWritten(), got.MustUnit())
		assert.Equal(t, tm(q, fr(185, 2)), tempo.MustNewQuarter(fraction.MustNew(370, 4)))
	})
}

// TestNewErrors covers С-30, С-31 and architecture 4.3, 4.6: unset unit or
// bpm — Е-2 without dependency sentinels; bpm <= 0 — Е-3; order
// Е-2 (unit) -> Е-2 (bpm) -> Е-3. The result is Tempo{}.
func TestNewErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func() (tempo.Tempo, error)
		want errClass
	}{
		{"С-31 unset unit", func() (tempo.Tempo, error) { return tempo.New(duration.Written{}, fi(60)) }, argClass},
		{"С-31 unset bpm", func() (tempo.Tempo, error) { return tempo.New(q, fraction.Fraction{}) }, argClass},
		{"С-30 bpm 0", func() (tempo.Tempo, error) { return tempo.New(q, fi(0)) }, noTempoClass},
		{"С-30 negative bpm", func() (tempo.Tempo, error) { return tempo.New(dq, fi(-60)) }, noTempoClass},
		{"С-30 negative fraction", func() (tempo.Tempo, error) { return tempo.New(dq, fr(-1, 2)) }, noTempoClass},
		{"order: unset unit before unset bpm",
			func() (tempo.Tempo, error) { return tempo.New(duration.Written{}, fraction.Fraction{}) }, argClass},
		{"order: unset unit before bpm 0", func() (tempo.Tempo, error) { return tempo.New(duration.Written{}, fi(0)) }, argClass},
		{"NewQuarter unset", func() (tempo.Tempo, error) { return tempo.NewQuarter(fraction.Fraction{}) }, argClass},
		{"С-30 NewQuarter 0", func() (tempo.Tempo, error) { return tempo.NewQuarter(fraction.Zero()) }, noTempoClass},
		{"С-30 NewQuarter negative", func() (tempo.Tempo, error) { return tempo.NewQuarter(fi(-1)) }, noTempoClass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			tt.want.check(t, err)
			assert.Equal(t, tempo.Tempo{}, got)
		})
	}
}

// TestTempoEquality covers С-3 and architecture 4.6, "Равенство и порядок":
// equal by notation (Equal, ==), equal by speed, comparison and ratio of
// speeds, and the total order Compare.
func TestTempoEquality(t *testing.T) {
	t.Parallel()

	dq60 := tm(dq, fi(60))
	e180 := tm(e, fi(180))
	q90 := tq(90)
	q120 := tq(120)

	t.Run("С-3 notation and speed", func(t *testing.T) {
		t.Parallel()

		assert.False(t, dq60.Equal(e180))
		assert.NotEqual(t, dq60, e180)
		assert.True(t, dq60.MustEqualSpeed(e180))
		assert.True(t, dq60.MustEqualSpeed(q90))
		assert.True(t, e180.MustEqualSpeed(q90))
		assert.False(t, q120.MustEqualSpeed(dq60))
		assert.True(t, tq(60).Equal(tm(q, fraction.MustNew(120, 2))))
		assert.Equal(t, tq(60), tm(q, fraction.MustNew(120, 2)))
	})

	t.Run("С-3 faster and ratio", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 1, q120.MustCompareSpeed(dq60))
		assert.Equal(t, -1, dq60.MustCompareSpeed(q120))
		assert.Equal(t, 0, dq60.MustCompareSpeed(e180))
		assertFrac(t, fr(4, 3), q120.MustSpeedRatio(dq60))
		assertFrac(t, fr(3, 4), dq60.MustSpeedRatio(q120))
		assertFrac(t, fi(1), dq60.MustSpeedRatio(e180))
	})

	t.Run("extreme values compare without Е-11", func(t *testing.T) {
		t.Parallel()

		big := tm(breve, maxF())
		small := tm(q, invMaxF())

		got, err := big.CompareSpeed(small)
		require.NoError(t, err)
		assert.Equal(t, 1, got)

		eq, err := big.EqualSpeed(tm(whole, maxF()))
		require.NoError(t, err)
		assert.False(t, eq)

		eq, err = tm(breve, maxF()).EqualSpeed(tm(breve, maxF()))
		require.NoError(t, err)
		assert.True(t, eq)

		assert.Equal(t, 1, big.Compare(small))
	})

	t.Run("Compare: speed, then unit", func(t *testing.T) {
		t.Parallel()

		sorted := []tempo.Tempo{{}, tm(q, fr(1, 1000)), e180, q90, dq60, q120, tm(breve, maxF())}
		for i, a := range sorted {
			for j, b := range sorted {
				want := 0
				if i < j {
					want = -1
				} else if i > j {
					want = 1
				}

				assert.Equal(t, want, a.Compare(b), "%s vs %s", a, b)
			}
		}

		shuffled := []tempo.Tempo{q120, dq60, tm(breve, maxF()), e180, {}, q90, tm(q, fr(1, 1000))}
		slices.SortFunc(shuffled, tempo.Tempo.Compare)
		assert.Equal(t, sorted, shuffled)
	})

	t.Run("== matches Equal", func(t *testing.T) {
		t.Parallel()

		values := []tempo.Tempo{{}, dq60, e180, q90, q120, tq(60), tm(q, fraction.MustNew(120, 2)), tm(e3, fi(360))}
		for _, a := range values {
			for _, b := range values {
				assert.Equal(t, a == b, a.Equal(b), "%s vs %s", a, b)
			}
		}

		m := map[tempo.Tempo]int{tq(60): 1}
		m[tm(q, fraction.MustNew(120, 2))]++
		assert.Equal(t, 2, m[tq(60)], "a Tempo is usable as a map key")
	})
}

// TestTempoChange covers С-4, С-5, С-52 and architecture 4.6, "Изменение":
// another unit with the same speed, another bpm, multiplication.
func TestTempoChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func() (tempo.Tempo, error)
		want tempo.Tempo
	}{
		{"С-4 ♩. = 60 -> ♪", func() (tempo.Tempo, error) { return tm(dq, fi(60)).WithUnit(e) }, tm(e, fi(180))},
		{"С-4 ♩ = 90 -> half", func() (tempo.Tempo, error) { return tq(90).WithUnit(h) }, tm(h, fi(45))},
		{"С-4 ♩ = 60 -> ♩.", func() (tempo.Tempo, error) { return tq(60).WithUnit(dq) }, tm(dq, fi(40))},
		{"С-4 ♩ = 100 -> ♩.", func() (tempo.Tempo, error) { return tq(100).WithUnit(dq) }, tm(dq, fr(200, 3))},
		{"С-52 the same unit", func() (tempo.Tempo, error) { return tm(dq, fi(60)).WithUnit(dq) }, tm(dq, fi(60))},
		{"С-51 triplet eighth -> ♩", func() (tempo.Tempo, error) { return tm(e3, fi(360)).WithUnit(q) }, tq(120)},
		{"WithBPM", func() (tempo.Tempo, error) { return tm(dq, fi(60)).WithBPM(fr(185, 2)) }, tm(dq, fr(185, 2))},
		{"WithBPM reduced", func() (tempo.Tempo, error) { return tq(1).WithBPM(fraction.MustNew(120, 2)) }, tq(60)},
		{"С-5 Double", func() (tempo.Tempo, error) { return tq(60).Double() }, tq(120)},
		{"С-5 Halve", func() (tempo.Tempo, error) { return tq(60).Halve() }, tq(30)},
		{"С-5 Mul 3/2", func() (tempo.Tempo, error) { return tq(60).Mul(fr(3, 2)) }, tq(90)},
		{"Mul keeps the unit", func() (tempo.Tempo, error) { return tm(dq, fi(60)).Mul(fr(1, 3)) }, tm(dq, fi(20))},
		{"Halve to a fraction", func() (tempo.Tempo, error) { return tq(45).Halve() }, tm(q, fr(45, 2))},
		{"Mul by 1", func() (tempo.Tempo, error) { return tq(45).Mul(fi(1)) }, tq(45)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "got %s", got)
		})
	}

	t.Run("С-4 the result keeps the speed", func(t *testing.T) {
		t.Parallel()

		for _, u := range []duration.Written{q, dq, e, h, dh, e3, whole, breve} {
			got := tq(100).MustWithUnit(u)
			assert.Equal(t, u, got.MustUnit())
			assert.True(t, got.MustEqualSpeed(tq(100)), "%s", got)
		}
	})
}

// TestModulation covers С-12, С-13, С-52 and architecture 4.6: Modulate
// gives v = N·u/w, the old duration w is any value (П-1); ModulationTo finds
// w = N·u/N′.
func TestModulation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from tempo.Tempo
		v    duration.Written
		w    duration.Duration
		want tempo.Tempo
	}{
		{"С-12 new ♩. = old ♩", tq(120), dq, dur(1, 4), tm(dq, fi(120))},
		{"С-12 new ♩ = old ♩.", tq(60), q, dur(3, 8), tq(40)},
		{"С-12 new ♪ = old triplet eighth", tq(120), e, dur(1, 12), tm(e, fi(360))},
		{"С-52 new u = old u", tm(dq, fi(60)), dq, dur(3, 8), tm(dq, fi(60))},
		{"П-1 old duration without a notation", tq(100), q, dur(5, 16), tq(80)},
		{"tied old duration", tq(60), q, dur(5, 8), tq(24)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.from.Modulate(tt.v, tt.w)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "got %s", got)

			back, err := tt.from.ModulationTo(got)
			require.NoError(t, err)
			assertDur(t, tt.w, back, "ModulationTo finds the old duration back")
		})
	}

	t.Run("С-12 equal speeds", func(t *testing.T) {
		t.Parallel()

		assert.True(t, tq(120).MustModulate(dq, dur(1, 4)).MustEqualSpeed(tq(180)))
		assert.True(t, tq(120).MustModulate(e, dur(1, 12)).MustEqualSpeed(tq(180)))
	})

	t.Run("С-13", func(t *testing.T) {
		t.Parallel()

		got, err := tq(120).ModulationTo(tm(dq, fi(120)))
		require.NoError(t, err)
		assertDur(t, dur(1, 4), got)

		assertDur(t, dur(3, 8), tq(60).MustModulationTo(tq(40)))
		assertDur(t, dur(1, 12), tq(120).MustModulationTo(tm(e, fi(360))))
		assertDur(t, dur(1, 4), tq(60).MustModulationTo(tq(60)))
	})
}

// TestSecondsOfDuration covers С-6, С-7, С-8, С-9, С-17 and architecture 4.6,
// "Секунды и обратные расчёты".
func TestSecondsOfDuration(t *testing.T) {
	t.Parallel()

	secTests := []struct {
		name string
		t    tempo.Tempo
		d    duration.Duration
		want tempo.Seconds
	}{
		{"С-6 ♩ = 60 half", tq(60), dur(1, 2), sec(2, 1)},
		{"С-6 ♩. = 60 eighth", tm(dq, fi(60)), dur(1, 8), sec(1, 3)},
		{"С-6 ♩ = 120 whole", tq(120), dur(1, 1), sec(2, 1)},
		{"С-6 bpm 60 without a unit", tempo.MustNewQuarter(fi(60)), duration.Half.MustDuration(), sec(2, 1)},
		{"С-6 6/8 bar at ♩. = 60", tm(dq, fi(60)), dur(3, 4), sec(2, 1)},
		{"С-6 ♩ = 90 triplet eighth", tq(90), dur(1, 12), sec(2, 9)},
		{"С-6 ♩ = 185/2 quarter", tm(q, fr(185, 2)), dur(1, 4), sec(24, 37)},
		{"С-6 ♩ = 60 dotted quarter", tq(60), dur(3, 8), sec(3, 2)},
		{"С-51 triplet eighth = 360 whole", tm(e3, fi(360)), dur(1, 1), sec(2, 1)},
	}

	for _, tt := range secTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.t.Seconds(tt.d)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "got %s", got)
		})
	}

	t.Run("С-7 unit seconds and units per second", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, sec(1, 1), tq(60).MustUnitSeconds())
		assert.Equal(t, sec(1, 2), tq(120).MustUnitSeconds())
		assert.Equal(t, sec(3, 4), tm(dq, fi(80)).MustUnitSeconds())
		assertFrac(t, fi(2), tq(120).MustUnitsPerSecond())
		assertFrac(t, fi(1), tq(60).MustUnitsPerSecond())
		assertFrac(t, fr(4, 3), tm(dq, fi(80)).MustUnitsPerSecond())
	})

	t.Run("С-8 duration in a time", func(t *testing.T) {
		t.Parallel()

		got, err := tq(60).DurationIn(sec(2, 1))
		require.NoError(t, err)
		assertDur(t, dur(1, 2), got)
		assertDur(t, dur(1, 8), tq(120).MustDurationIn(sec(1, 4)))
		assertDur(t, dur(3, 4), tm(dq, fi(60)).MustDurationIn(sec(2, 1)))
	})

	t.Run("С-9 tempo from the time of a duration", func(t *testing.T) {
		t.Parallel()

		got, err := tempo.FromSeconds(q, dur(1, 4), sec(1, 2))
		require.NoError(t, err)
		assert.Equal(t, tq(120), got)
		assert.Equal(t, tq(60), tempo.MustFromSeconds(q, dur(1, 2), sec(2, 1)))
		assert.Equal(t, tm(dq, fi(60)), tempo.MustFromSeconds(dq, dur(1, 8), sec(1, 3)))
		assert.Equal(t, tm(q, fr(185, 2)), tempo.MustFromSeconds(q, dur(1, 4), sec(24, 37)))
	})

	t.Run("С-17 counting units", func(t *testing.T) {
		t.Parallel()

		got, err := tq(90).UnitsIn(sec(40, 1))
		require.NoError(t, err)
		assertFrac(t, fi(60), got)
		assertFrac(t, fi(0), tq(90).MustUnitsIn(tempo.ZeroSeconds()))
		assert.Equal(t, sec(40, 1), tq(90).MustSecondsOfUnits(fi(60)))
		assert.Equal(t, tempo.ZeroSeconds(), tq(90).MustSecondsOfUnits(fi(0)))
		assert.Equal(t, sec(1, 3), tq(90).MustSecondsOfUnits(fr(1, 2)))
	})

	t.Run("С-34 no solution", func(t *testing.T) {
		t.Parallel()

		d, err := tq(60).DurationIn(tempo.ZeroSeconds())
		noSolution.check(t, err)
		assert.Equal(t, duration.Duration{}, d)

		x, err := tempo.FromSeconds(q, dur(1, 4), tempo.ZeroSeconds())
		noSolution.check(t, err)
		assert.Equal(t, tempo.Tempo{}, x)

		s, err := tq(90).SecondsOfUnits(fi(-1))
		noSolution.check(t, err)
		assert.Equal(t, tempo.Seconds{}, s)
	})
}

// TestQuantize covers С-18, С-20, С-37 and architecture 4.6, 4.8: bpm is
// replaced by a node k·q by the mode; the unit is kept; a node of 0 is Е-8.
func TestQuantize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from tempo.Tempo
		step fraction.Fraction
		mode roundingMode
		want tempo.Tempo
	}{
		{"С-18 HalfUp", tm(dq, fr(200, 3)), fi(1), halfUp, tm(dq, fi(67))},
		{"С-18 Down", tm(dq, fr(200, 3)), fi(1), down, tm(dq, fi(66))},
		{"С-18 step 1/2 Down", tm(dq, fr(200, 3)), fr(1, 2), down, tm(dq, fr(133, 2))},
		{"С-20 HalfEven of 251/2", tm(q, fr(251, 2)), fi(1), halfEven, tq(126)},
		{"С-37 Up of 1/4", tm(q, fr(1, 4)), fi(1), up, tq(1)},
		{"node in Down", tq(120), fi(1), down, tq(120)},
		{"node in Up", tq(120), fi(10), up, tq(120)},
		{"step 5 HalfDown tie", tq(125), fi(5), halfDown, tq(125)},
		{"step 10 HalfDown tie", tq(125), fi(10), halfDown, tq(120)},
		{"step 10 HalfUp tie", tq(125), fi(10), halfUp, tq(130)},
		{"step 10 HalfEven tie to 12·10", tq(125), fi(10), halfEven, tq(120)},
		{"step 10 HalfEven tie to 14·10", tq(135), fi(10), halfEven, tq(140)},
		{"С-59 TowardZero as Down", tm(dq, fr(200, 3)), fi(1), towardZero, tm(dq, fi(66))},
		{"С-59 AwayFromZero as Up", tm(dq, fr(200, 3)), fi(1), awayFromZero, tm(dq, fi(67))},
		{"С-59 HalfTowardZero as HalfDown", tq(125), fi(10), halfTowardZero, tq(120)},
		{"С-59 HalfAwayFromZero as HalfUp", tq(125), fi(10), halfAwayFromZero, tq(130)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.from.Quantize(tt.step, tt.mode)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "got %s", got)
		})
	}

	t.Run("С-37 quantized to zero", func(t *testing.T) {
		t.Parallel()

		for _, m := range []roundingMode{down, towardZero, halfUp, halfEven} {
			got, err := tm(q, fr(1, 4)).Quantize(fi(1), m)
			quantZero.check(t, err)
			assert.Equal(t, tempo.Tempo{}, got)
		}
	})

	t.Run("С-33 step and mode", func(t *testing.T) {
		t.Parallel()

		for _, step := range []fraction.Fraction{fi(0), fi(-1), {}} {
			got, err := tq(60).Quantize(step, down)
			argClass.check(t, err)
			assert.Equal(t, tempo.Tempo{}, got)
		}

		_, err := tq(60).Quantize(fi(1), roundingMode(0))
		modeUnset.check(t, err)

		_, err = tq(60).Quantize(fi(1), roundingMode(42))
		modeUnknown.check(t, err)
	})

	t.Run("С-43 out of range", func(t *testing.T) {
		t.Parallel()

		got, err := tm(q, maxF()).Quantize(fr(3, 2), up)
		fracRange.check(t, err)
		assert.Equal(t, tempo.Tempo{}, got)

		_, err = tm(q, maxF()).Quantize(fr(3, 2), down)
		assert.NoError(t, err)
	})
}

// TestSnapToScale covers С-22 and architecture 4.6, "Шкала".
func TestSnapToScale(t *testing.T) {
	t.Parallel()

	got, err := tm(dq, fi(62)).SnapToScale(halfUp)
	require.NoError(t, err)
	assert.Equal(t, tm(dq, fi(63)), got)

	assert.Equal(t, tm(dq, fi(60)), tm(dq, fi(62)).MustSnapToScale(down))
	assert.Equal(t, tq(112), tq(110).MustSnapToScale(halfEven))
	assert.Equal(t, tq(108), tq(110).MustSnapToScale(halfDown))
	assert.Equal(t, tq(120), tq(120).MustSnapToScale(up))
	assert.Equal(t, tm(e, fi(60)), tm(e, fr(121, 2)).MustSnapToScale(halfUp))

	for _, bpm := range []fraction.Fraction{fi(30), fi(250), fr(79, 2), fr(417, 2)} {
		got, err := tq(1).MustWithBPM(bpm).SnapToScale(halfUp)
		offScale.check(t, err)
		assert.Equal(t, tempo.Tempo{}, got)
	}

	_, err = tq(110).SnapToScale(roundingMode(0))
	modeUnset.check(t, err)

	_, err = tq(30).SnapToScale(roundingMode(0))
	modeUnset.check(t, err)
}
