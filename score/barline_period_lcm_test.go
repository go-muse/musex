package score_test

import (
	"errors"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file check BarlinePeriod against the formula of score.md,
// "Совпадение тактовых черт": for reduced lengths a/b and c/d the period is
// lcm(a, c)/gcd(b, d), the intermediate integers are unbounded, and the
// result is checked against the range of duration and common/fraction. Every
// expected value is computed by periodOracle on math/big from the lengths of
// the meters, which the tests build themselves.

// periodMeter is a meter together with its exact length.
type periodMeter struct {
	meter  meter.Meter
	length *big.Rat
}

// periodSimple is the simple meter n/d.
func periodSimple(n, d int) periodMeter {
	return periodMeter{meter.MustNew(n, d), big.NewRat(int64(n), int64(d))}
}

// periodSum is the sum of the meters 1/d for the denominators dens.
func periodSum(dens ...int) periodMeter {
	parts := make([]meter.Meter, len(dens))
	length := new(big.Rat)

	for i, d := range dens {
		parts[i] = meter.MustNew(1, d)
		length.Add(length, big.NewRat(1, int64(d)))
	}

	return periodMeter{meter.MustNewSum(parts...), length}
}

// grid is a finite grid of count bars of the meter.
func (m periodMeter) grid(count int) measure.Grid {
	return measure.MustNew(m.meter).MustWithBarCount(count)
}

// periodPart is a part with the grid g.
func periodPart(g measure.Grid) *track.Track {
	return track.MustNew(instrument.MustNew(instrument.Piano()), g)
}

// periodOracle is lcm(a, c)/gcd(b, d) for the reduced lengths a/b and c/d.
func periodOracle(x, y *big.Rat) *big.Rat {
	a, c := x.Num(), y.Num()
	lcm := new(big.Int).Mul(a, c)
	lcm.Quo(lcm, new(big.Int).GCD(nil, nil, a, c))

	return new(big.Rat).SetFrac(lcm, new(big.Int).GCD(nil, nil, x.Denom(), y.Denom()))
}

// periodInRange reports whether both components of r are at most
// fraction.MaxComponent.
func periodInRange(r *big.Rat) bool {
	k := fraction.MaxComponentBig()

	return new(big.Int).Abs(r.Num()).Cmp(k) <= 0 && r.Denom().Cmp(k) <= 0
}

// periodDuration is the duration of an exact positive value in range.
func periodDuration(r *big.Rat) duration.Duration {
	return duration.MustFromFraction(fraction.MustFromBigRat(r))
}

// periodPanic runs f and returns the error it panics with.
func periodPanic(t *testing.T, f func()) (err error) {
	t.Helper()

	defer func() {
		r := recover()
		require.NotNil(t, r, "the Must variant does not panic")

		var ok bool
		err, ok = r.(error)
		require.True(t, ok, "the panic value is %T, want error", r)
	}()

	f()

	return nil
}

// assertPeriod checks that BarlinePeriod(a, b), BarlinePeriod(b, a), and
// MustBarlinePeriod return want.
func assertPeriod(t *testing.T, s *score.Score, a, b score.PartID, want *big.Rat) {
	t.Helper()

	expected := periodDuration(want)

	got, err := s.BarlinePeriod(a, b)
	require.NoError(t, err)
	assert.True(t, got.Equal(expected), "BarlinePeriod = %s, want %s", got, want.RatString())

	back, err := s.BarlinePeriod(b, a)
	require.NoError(t, err)
	assert.True(t, back.Equal(expected), "the period does not depend on the order of the parts")

	assert.True(t, s.MustBarlinePeriod(a, b).Equal(expected), "MustBarlinePeriod")
}

// assertPeriodOutOfRange checks the refusal of a period whose numerator is
// greater than fraction.MaxComponent: score.ErrOutOfRange together with
// fraction.ErrOutOfRange, a zero Duration, and the panic of
// MustBarlinePeriod with the same error.
func assertPeriodOutOfRange(t *testing.T, s *score.Score, a, b score.PartID) {
	t.Helper()

	got, err := s.BarlinePeriod(a, b)
	require.Error(t, err)
	assert.ErrorIs(t, err, score.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrOutOfRange)
	assert.NotErrorIs(t, err, score.ErrPeriod)
	assert.NotErrorIs(t, err, score.ErrPart)
	assert.NotErrorIs(t, err, score.ErrInvalidArgument)
	assert.False(t, got.IsSet())
	assert.True(t, got == duration.Duration{}, "the result on an error is the zero Duration")

	panicked := periodPanic(t, func() { s.MustBarlinePeriod(a, b) })
	assert.EqualError(t, panicked, err.Error(), "MustBarlinePeriod panics with the same error")
	assert.ErrorIs(t, panicked, score.ErrOutOfRange)
	assert.ErrorIs(t, panicked, fraction.ErrOutOfRange)

	_, back := s.BarlinePeriod(b, a)
	assert.ErrorIs(t, back, score.ErrOutOfRange, "the refusal does not depend on the order of the parts")
}

// Primes below 2⁶³ used for denominators of sum meters: periodP0 is the
// largest prime below 2⁶³, and the next ones follow downward.
const (
	periodP0 = 1<<63 - 25
	periodP1 = 1<<63 - 165
	periodP2 = 1<<63 - 259
	periodP3 = 1<<63 - 301
	periodP4 = 1<<63 - 375
	periodP5 = 1<<63 - 387
)

// Near the bound K = fraction.MaxComponent the period is built from two sums
// of three meters 1/p: c = 1/periodP0 + 1/periodP1 + 1/periodP2 and
// a = 1/periodP3 + 1/periodP4 + 1/d3. The reduced numerator of a is
// periodP3·periodP4 + d3·(periodP3 + periodP4), so the product of the
// numerators grows by (periodP3 + periodP4)·num(c) < 2¹⁹² per unit of d3.
// periodBelowK is the largest d3 with lcm(num a, num c) ≤ K and
// periodBelowK + 1 the smallest with lcm > K; both keep the numerators
// coprime. lcm = K exactly is not reached by meters, whose numerators are
// sums of at most 1024 units over denominators of at most 63 bits; lcm = K + 1
// = 2²⁵⁵ is unreachable by any two numerators of at most K.
const periodBelowK = 7686143364045647525

func periodNearK(d3 int) (a, c periodMeter) {
	return periodSum(periodP3, periodP4, d3), periodSum(periodP0, periodP1, periodP2)
}

// periodWide is a pair of sums of four meters 1/p whose numerators have about
// 189 bits each, so that their LCM is far greater than K.
func periodWide() (a, c periodMeter) {
	return periodSum(periodP0, periodP1, periodP2, periodP3), periodSum(periodP4, periodP5, 1<<62-57, 1<<62-87)
}

// TestBarlinePeriodLCMFormula checks the period lcm(a, c)/gcd(b, d) for equal,
// multiple, coprime, and wide lengths, and for lengths with common factors
// of numerators and denominators.
func TestBarlinePeriodLCMFormula(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x, y periodMeter
		want string
	}{
		{"equal values of different notation", periodSimple(3, 4), periodSimple(6, 8), "3/4"},
		{"equal meters", periodSimple(5, 7), periodSimple(5, 7), "5/7"},
		{"multiple numerators, equal denominators", periodSimple(1, 4), periodSimple(3, 4), "3/4"},
		{"multiple lengths", periodSimple(1, 2), periodSimple(3, 2), "3/2"},
		{"multiple denominators", periodSimple(1, 4), periodSimple(1, 8), "1/4"},
		{"coprime numerators and denominators", periodSimple(3, 4), periodSimple(5, 7), "15"},
		{"coprime numerators, coprime denominators", periodSimple(7, 8), periodSimple(5, 3), "35"},
		{"common factor of numerators and of denominators", periodSimple(6, 8), periodSimple(9, 8), "9/4"},
		{"equal numerators, common factor of denominators", periodSimple(10, 12), periodSimple(15, 9), "5/3"},
		{"coprime numerators, coprime denominators after reduction", periodSimple(4, 6), periodSimple(6, 10), "6"},
		{"powers of two near the int limit", periodSimple(1, 1<<62), periodSimple(3, 1<<61), "3/2305843009213693952"},
		{"primes near the int limit", periodSimple(1, periodP0), periodSimple(1, periodP1), "1"},
		{"numerators of prime products", periodSimple(1021, periodP0), periodSimple(1019, periodP1), "1040399"},
		{"wide numerator against a whole", periodSum(periodP0, periodP1, periodP2), periodSimple(4, 4), ""},
		{"wide numerators of two sums", periodSum(periodP0, periodP1, periodP2), periodSum(periodP3, periodP4), ""},
		{"wide sums with shared denominators", periodSum(periodP0, periodP1, periodP2), periodSum(periodP0, periodP1), ""},
		{"wide sum against a simple meter", periodSum(periodP0, periodP1, periodP2), periodSimple(1024, periodP3), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := periodOracle(tt.x.length, tt.y.length)
			if tt.want != "" {
				r, ok := new(big.Rat).SetString(tt.want)
				require.True(t, ok)
				require.Equal(t, 0, r.Cmp(want), "the table value %s differs from lcm/gcd %s", tt.want, want.RatString())
			} else {
				require.Greater(t, want.Num().BitLen(), 64, "the case has a numerator wider than 64 bits")
			}

			require.True(t, periodInRange(want))

			s, ids := assemble(clock60(), periodPart(tt.x.grid(2)), periodPart(tt.y.grid(3)))
			assertPeriod(t, s, ids[0], ids[1], want)
		})
	}
}

// TestBarlinePeriodLCMRange checks the refusal when the numerator of the LCM
// is greater than K and the nearest values on both sides of K that meters
// reach.
func TestBarlinePeriodLCMRange(t *testing.T) {
	t.Parallel()

	k := fraction.MaxComponentBig()
	window := new(big.Int).Lsh(big.NewInt(1), 191)

	t.Run("largest reached value below K", func(t *testing.T) {
		t.Parallel()

		a, c := periodNearK(periodBelowK)
		want := periodOracle(a.length, c.length)
		require.True(t, want.IsInt(), "the denominators are coprime")
		gap := new(big.Int).Sub(k, want.Num())
		require.GreaterOrEqual(t, gap.Sign(), 0, "lcm ≤ K")
		require.Negative(t, gap.Cmp(window), "K − lcm < 2¹⁹¹")
		require.Equal(t, 255, want.Num().BitLen())

		s, ids := assemble(clock60(), periodPart(a.grid(1)), periodPart(c.grid(1)))
		assertPeriod(t, s, ids[0], ids[1], want)
	})

	t.Run("smallest reached value above K", func(t *testing.T) {
		t.Parallel()

		a, c := periodNearK(periodBelowK + 1)
		want := periodOracle(a.length, c.length)
		excess := new(big.Int).Sub(want.Num(), k)
		require.Positive(t, excess.Sign(), "lcm > K")
		require.Negative(t, excess.Cmp(window), "lcm − K < 2¹⁹¹")
		require.True(t, periodInRange(a.length))
		require.True(t, periodInRange(c.length))

		s, ids := assemble(clock60(), periodPart(a.grid(1)), periodPart(c.grid(1)))
		assertPeriodOutOfRange(t, s, ids[0], ids[1])
	})

	t.Run("sums of four meters", func(t *testing.T) {
		t.Parallel()

		a, c := periodWide()
		want := periodOracle(a.length, c.length)
		require.Greater(t, want.Num().BitLen(), fraction.MaxComponentBits)

		s, ids := assemble(clock60(), periodPart(a.grid(2)), periodPart(c.grid(2)))
		assertPeriodOutOfRange(t, s, ids[0], ids[1])
	})
}

// TestBarlinePeriodLCMOracle compares BarlinePeriod with periodOracle on pairs
// of generated meters with a fixed seed: simple meters with denominators up to
// the int limit and sums of up to four meters, whose LCM lies on both sides
// of K.
func TestBarlinePeriodLCMOracle(t *testing.T) {
	t.Parallel()

	const seed = 2608

	rng := rand.New(rand.NewPCG(seed, seed))

	count := 3000
	if testing.Short() {
		count = 100
	}

	var inside, outside int

	for i := range count {
		x, y := periodRandom(rng), periodRandom(rng)
		want := periodOracle(x.length, y.length)

		s, ids := assemble(clock60(), periodPart(x.grid(1)), periodPart(y.grid(1)))

		if periodInRange(want) {
			inside++

			got, err := s.BarlinePeriod(ids[0], ids[1])
			if assert.NoError(t, err, "case %d (seed %d): %s and %s", i, seed, x.meter, y.meter) {
				assert.True(t, got.Equal(periodDuration(want)), "case %d (seed %d): got %s, want %s", i, seed, got, want.RatString())
			}

			continue
		}

		outside++

		got, err := s.BarlinePeriod(ids[0], ids[1])
		assert.ErrorIs(t, err, score.ErrOutOfRange, "case %d (seed %d): %s and %s", i, seed, x.meter, y.meter)
		assert.ErrorIs(t, err, fraction.ErrOutOfRange, "case %d (seed %d)", i, seed)
		assert.False(t, got.IsSet(), "case %d (seed %d)", i, seed)
	}

	assert.Positive(t, inside, "the generated pairs have periods in range")
	assert.Positive(t, outside, "the generated pairs have periods out of range")
}

// periodRandom is a simple meter n/d with a random denominator up to the int
// limit or a sum of two to four meters 1/d with random denominators of 63
// bits; the product of the denominators of a sum stays in range.
func periodRandom(rng *rand.Rand) periodMeter {
	if rng.IntN(3) == 0 {
		d := 1 + rng.IntN(1<<62)
		if rng.IntN(2) == 0 {
			d = 1 + rng.IntN(64)
		}

		return periodSimple(1+rng.IntN(1024), d)
	}

	dens := make([]int, 2+rng.IntN(3))
	for i := range dens {
		dens[i] = 1<<62 + rng.IntN(1<<62)
	}

	return periodSum(dens...)
}

// TestBarlinePeriodLCMPriority checks that the checks of the parts precede the
// range check: an unknown part, the same part twice, and ErrPeriod are
// returned even when the lengths of the parts would give a period out of
// range.
func TestBarlinePeriodLCMPriority(t *testing.T) {
	t.Parallel()

	wideA, wideC := periodWide()
	require.Greater(t, periodOracle(wideA.length, wideC.length).Num().BitLen(), fraction.MaxComponentBits,
		"the lengths of the parts give a period out of range")

	pickup := wideA.grid(2).MustWithPickup(periodDuration(new(big.Rat).SetFrac64(1, periodP0)))
	// The second bar is shorter; its denominator divides that of wideC, so
	// the end of the grid stays in range.
	changing := wideC.grid(2).MustWithMark(2, meter.MustNew(1, periodP4))

	s, ids := assemble(clock60(), periodPart(wideA.grid(2)), periodPart(wideC.grid(2)), periodPart(pickup), periodPart(changing))
	_, foreign := assemble(clock60(), periodPart(wideC.grid(2)))
	wide, wideOther, withPickup, withChange := ids[0], ids[1], ids[2], ids[3]

	tests := []struct {
		name   string
		a, b   score.PartID
		want   error
		period score.PartID
	}{
		{"unknown first part", foreign[0], wideOther, score.ErrPart, score.PartID{}},
		{"unknown second part", wide, foreign[0], score.ErrPart, score.PartID{}},
		{"zero second part", wide, score.PartID{}, score.ErrPart, score.PartID{}},
		{"unknown part before a pickup", foreign[0], withPickup, score.ErrPart, score.PartID{}},
		{"unknown part after a pickup", withPickup, foreign[0], score.ErrPart, score.PartID{}},
		{"the same part with a pickup", withPickup, withPickup, score.ErrInvalidArgument, score.PartID{}},
		{"first part with a pickup", withPickup, wideOther, score.ErrPeriod, withPickup},
		{"second part with a pickup", wideOther, withPickup, score.ErrPeriod, withPickup},
		{"second part with a variable length", wide, withChange, score.ErrPeriod, withChange},
		{"both parts not applicable", withChange, withPickup, score.ErrPeriod, withChange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := s.BarlinePeriod(tt.a, tt.b)
			require.ErrorIs(t, err, tt.want)
			assert.NotErrorIs(t, err, score.ErrOutOfRange)
			assert.NotErrorIs(t, err, fraction.ErrOutOfRange)
			assert.False(t, got.IsSet())
			assert.True(t, got == duration.Duration{})

			for _, other := range []error{score.ErrPart, score.ErrInvalidArgument, score.ErrPeriod} {
				if other != tt.want {
					assert.NotErrorIs(t, err, other)
				}
			}

			if tt.want == score.ErrPeriod {
				assertKey(t, tt.period, problems(t, err).Issues[0].Part)
			}

			panicked := periodPanic(t, func() { s.MustBarlinePeriod(tt.a, tt.b) })
			assert.EqualError(t, panicked, err.Error())
			assert.True(t, errors.Is(panicked, tt.want))
		})
	}
}
