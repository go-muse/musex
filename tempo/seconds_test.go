package tempo_test

import (
	"math"
	"math/big"
	"slices"
	"testing"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecondsCreate covers С-10, С-35, С-53 and architecture 4.7,
// "Создание", "Чтение".
func TestSecondsCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func() (tempo.Seconds, error)
		want fraction.Fraction
	}{
		{"С-53 ZeroSeconds", func() (tempo.Seconds, error) { return tempo.ZeroSeconds(), nil }, fraction.Zero()},
		{"NewSeconds", func() (tempo.Seconds, error) { return tempo.NewSeconds(1, 3) }, fr(1, 3)},
		{"NewSeconds reduced", func() (tempo.Seconds, error) { return tempo.NewSeconds(4, 2) }, fi(2)},
		{"NewSeconds signs by value", func() (tempo.Seconds, error) { return tempo.NewSeconds(-1, -2) }, fr(1, 2)},
		{"NewSeconds zero", func() (tempo.Seconds, error) { return tempo.NewSeconds(0, 5) }, fraction.Zero()},
		{"SecondsFromFraction", func() (tempo.Seconds, error) { return tempo.SecondsFromFraction(fraction.MustNew(6, 4)) }, fr(3, 2)},
		{"SecondsFromFraction zero", func() (tempo.Seconds, error) { return tempo.SecondsFromFraction(fraction.Zero()) }, fraction.Zero()},
		{"SecondsFromFraction max", func() (tempo.Seconds, error) { return tempo.SecondsFromFraction(maxF()) }, maxF()},
		{"С-10 2 minutes", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(fi(2)) }, fi(120)},
		{"minutes 1/120", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(fr(1, 120)) }, fr(1, 2)},
		{"minutes zero", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(fraction.Zero()) }, fraction.Zero()},
		{"С-11 system time 500 ms", func() (tempo.Seconds, error) { return tempo.SecondsFromTimeDuration(500 * time.Millisecond) }, fr(1, 2)},
		{"system time 1 ns", func() (tempo.Seconds, error) { return tempo.SecondsFromTimeDuration(1) }, fr(1, 1000000000)},
		{"system time zero", func() (tempo.Seconds, error) { return tempo.SecondsFromTimeDuration(0) }, fraction.Zero()},
		{"system time max", func() (tempo.Seconds, error) { return tempo.SecondsFromTimeDuration(math.MaxInt64) },
			fbig(big.NewInt(math.MaxInt64), big.NewInt(1000000000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			require.NoError(t, err)
			assert.True(t, got.IsSet())
			assert.Equal(t, tt.want, got.MustFraction(), "the fraction is reduced, 0/1 for zero")
		})
	}

	t.Run("С-10 minutes", func(t *testing.T) {
		t.Parallel()

		got, err := sec(90, 1).Minutes()
		require.NoError(t, err)
		assertFrac(t, fr(3, 2), got)
		assertFrac(t, fi(2), sec(120, 1).MustMinutes())
		assertFrac(t, fi(0), tempo.ZeroSeconds().MustMinutes())
	})

	t.Run("С-53 zero is not unset", func(t *testing.T) {
		t.Parallel()

		assert.NotEqual(t, tempo.Seconds{}, tempo.ZeroSeconds())
		assert.False(t, tempo.ZeroSeconds().Equal(tempo.Seconds{}))
		assert.Equal(t, tempo.ZeroSeconds(), sec(0, 1))
	})
}

// TestSecondsCreateErrors covers С-35 and architecture 4.3, 4.7: Е-2 with
// fraction.ErrZeroDenominator for a zero denominator, Е-2 for an unset
// fraction, Е-5 for a negative value, Е-11 for 60·f out of range, and the
// order of checks.
func TestSecondsCreateErrors(t *testing.T) {
	t.Parallel()

	negMax := fraction.MustFromBigIntReduced(new(big.Int).Neg(fraction.MaxComponentBig()), big.NewInt(1))

	tests := []struct {
		name string
		call func() (tempo.Seconds, error)
		want errClass
	}{
		{"zero denominator", func() (tempo.Seconds, error) { return tempo.NewSeconds(1, 0) }, zeroDen},
		{"order: zero denominator before negative", func() (tempo.Seconds, error) { return tempo.NewSeconds(-1, 0) }, zeroDen},
		{"С-35 −1 s", func() (tempo.Seconds, error) { return tempo.NewSeconds(-1, 1) }, negTime},
		{"negative by signs", func() (tempo.Seconds, error) { return tempo.NewSeconds(1, -3) }, negTime},
		{"unset fraction", func() (tempo.Seconds, error) { return tempo.SecondsFromFraction(fraction.Fraction{}) }, argClass},
		{"negative fraction", func() (tempo.Seconds, error) { return tempo.SecondsFromFraction(fi(-1)) }, negTime},
		{"unset minutes", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(fraction.Fraction{}) }, argClass},
		{"negative minutes", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(fr(-1, 60)) }, negTime},
		{"С-43 minutes out of range", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(maxF()) }, fracRange},
		{"order: negative before out of range", func() (tempo.Seconds, error) { return tempo.SecondsFromMinutes(negMax) }, negTime},
		{"С-35 negative system time", func() (tempo.Seconds, error) { return tempo.SecondsFromTimeDuration(-1) }, negTime},
		{"most negative system time", func() (tempo.Seconds, error) { return tempo.SecondsFromTimeDuration(math.MinInt64) }, negTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			tt.want.check(t, err)
			assert.Equal(t, tempo.Seconds{}, got)
		})
	}
}

// TestSecondsArithmetic covers С-10, С-35 and architecture 4.7: order and
// equality by value, sum and difference.
func TestSecondsArithmetic(t *testing.T) {
	t.Parallel()

	got, err := sec(2, 1).Add(sec(1, 3))
	require.NoError(t, err)
	assert.Equal(t, sec(7, 3), got)

	got, err = sec(2, 1).Sub(sec(1, 3))
	require.NoError(t, err)
	assert.Equal(t, sec(5, 3), got)

	assert.Equal(t, tempo.ZeroSeconds(), sec(1, 3).MustSub(sec(1, 3)))
	assert.Equal(t, sec(1, 3), tempo.ZeroSeconds().MustAdd(sec(1, 3)))

	got, err = sec(1, 1).Sub(sec(2, 1))
	negTime.check(t, err)
	assert.Equal(t, tempo.Seconds{}, got)

	got, err = secF(maxF()).Add(secF(maxF()))
	fracRange.check(t, err)
	assert.Equal(t, tempo.Seconds{}, got)

	got, err = secF(maxF()).Sub(secF(invMaxF()))
	fracRange.check(t, err)
	assert.Equal(t, tempo.Seconds{}, got)

	sorted := []tempo.Seconds{{}, tempo.ZeroSeconds(), secF(invMaxF()), sec(1, 3), sec(1, 2), sec(2, 1), secF(maxF())}
	for i, a := range sorted {
		for j, b := range sorted {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}

			assert.Equal(t, want, a.Compare(b), "%s vs %s", a, b)
			assert.Equal(t, i == j, a.Equal(b))
			assert.Equal(t, a == b, a.Equal(b))
		}
	}

	shuffled := []tempo.Seconds{sec(2, 1), {}, secF(maxF()), sec(1, 2), tempo.ZeroSeconds(), sec(1, 3), secF(invMaxF())}
	slices.SortFunc(shuffled, tempo.Seconds.Compare)
	assert.Equal(t, sorted, shuffled)

	m := map[tempo.Seconds]int{sec(1, 2): 1}
	m[tempo.MustNewSeconds(2, 4)]++
	assert.Equal(t, 2, m[sec(1, 2)])
}

// TestTimeDuration covers С-11, С-36, С-53, С-54, С-59 and architecture 4.7,
// 4.8: rounding to whole nanoseconds by the mode, the nine modes, the bound
// of Е-9, and the mode check also for whole nanoseconds.
func TestTimeDuration(t *testing.T) {
	t.Parallel()

	modes := []roundingMode{
		down, up, towardZero, awayFromZero, halfUp, halfDown, halfAwayFromZero, halfTowardZero, halfEven,
	}

	tests := []struct {
		name string
		s    tempo.Seconds
		want []time.Duration // in the order of modes
	}{
		{"С-11 2 s", sec(2, 1), []time.Duration{2e9, 2e9, 2e9, 2e9, 2e9, 2e9, 2e9, 2e9, 2e9}},
		{"С-53 zero", tempo.ZeroSeconds(), []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{"С-11 С-59 1/3 s", sec(1, 3),
			[]time.Duration{333333333, 333333334, 333333333, 333333334, 333333333, 333333333, 333333333, 333333333, 333333333}},
		{"2/3 s", sec(2, 3),
			[]time.Duration{666666666, 666666667, 666666666, 666666667, 666666667, 666666667, 666666667, 666666667, 666666667}},
		{"С-11 24/37 s", sec(24, 37),
			[]time.Duration{648648648, 648648649, 648648648, 648648649, 648648649, 648648649, 648648649, 648648649, 648648649}},
		{"С-54 half a nanosecond", sec(1, 2000000000), []time.Duration{0, 1, 0, 1, 1, 0, 1, 0, 0}},
		{"С-54 one and a half nanoseconds", sec(3, 2000000000), []time.Duration{1, 2, 1, 2, 2, 1, 2, 1, 2}},
		{"two and a half nanoseconds", sec(5, 2000000000), []time.Duration{2, 3, 2, 3, 3, 2, 3, 2, 2}},
		{"system time max", tempo.MustSecondsFromTimeDuration(math.MaxInt64),
			[]time.Duration{math.MaxInt64, math.MaxInt64, math.MaxInt64, math.MaxInt64, math.MaxInt64,
				math.MaxInt64, math.MaxInt64, math.MaxInt64, math.MaxInt64}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for i, m := range modes {
				got, err := tt.s.TimeDuration(m)
				require.NoError(t, err, "%s", m)
				assert.Equal(t, tt.want[i], got, "%s", m)
			}
		})
	}

	t.Run("С-11 back from system time", func(t *testing.T) {
		t.Parallel()

		got := tempo.MustSecondsFromTimeDuration(500000000)
		assert.Equal(t, sec(1, 2), got)
		assert.Equal(t, 500*time.Millisecond, got.MustTimeDuration(down))
	})

	t.Run("С-36 bound of Е-9", func(t *testing.T) {
		t.Parallel()

		// (2^64 − 1)/(2·10^9) s is 2^63 − 1/2 ns: Down and HalfDown give
		// MaxInt64; Up, HalfUp and HalfEven (the lower 2^63 − 1 is odd) do not fit.
		num := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
		edge := secF(fbig(num, big.NewInt(2000000000)))

		for _, m := range []roundingMode{down, towardZero, halfDown, halfTowardZero} {
			got, err := edge.TimeDuration(m)
			require.NoError(t, err, "%s", m)
			assert.Equal(t, time.Duration(math.MaxInt64), got, "%s", m)
		}

		for _, m := range []roundingMode{up, awayFromZero, halfUp, halfAwayFromZero, halfEven} {
			got, err := edge.TimeDuration(m)
			overflow.check(t, err)
			assert.Equal(t, time.Duration(0), got, "%s", m)
		}

		for _, s := range []tempo.Seconds{sec(1000000000000, 1), secF(maxF())} {
			for _, m := range allModes() {
				got, err := s.TimeDuration(m)
				overflow.check(t, err)
				assert.Equal(t, time.Duration(0), got)
			}
		}
	})

	t.Run("С-33 mode", func(t *testing.T) {
		t.Parallel()

		for _, s := range []tempo.Seconds{sec(1, 3), sec(2, 1), tempo.ZeroSeconds(), secF(maxF())} {
			got, err := s.TimeDuration(roundingMode(0))
			modeUnset.check(t, err)
			assert.Equal(t, time.Duration(0), got)

			_, err = s.TimeDuration(roundingMode(10))
			modeUnknown.check(t, err)

			_, err = s.TimeDuration(roundingMode(255))
			modeUnknown.check(t, err)
		}
	})
}
