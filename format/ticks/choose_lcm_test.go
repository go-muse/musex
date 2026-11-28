package ticks_test

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file check MinResolution (ticks.md 4.5) where task 26.8
// changes its computation: the control examples of architecture 4.3, the
// refusal Е-8 and the order of checks, and a property against an oracle of
// the definition on math/big over all bb 1…255.

// minresK is K = fraction.MaxComponent.
func minresK() *big.Int { return fraction.MaxComponentBig() }

func minresPow2(n uint) *big.Int { return new(big.Int).Lsh(big.NewInt(1), n) }

// minresRat is num/den.
func minresRat(num, den *big.Int) *big.Rat { return new(big.Rat).SetFrac(num, den) }

// minresInv is 1/den.
func minresInv(den *big.Int) *big.Rat { return minresRat(big.NewInt(1), den) }

func minresDuration(v *big.Rat) duration.Duration {
	return duration.MustFromFraction(fraction.MustFromBigRat(v))
}

func minresPosition(v *big.Rat) duration.Position {
	return duration.MustPositionFromFraction(fraction.MustFromBigRat(v))
}

func minresDurations(vs []*big.Rat) []duration.Duration {
	if vs == nil {
		return nil
	}

	out := make([]duration.Duration, 0, len(vs))
	for _, v := range vs {
		out = append(out, minresDuration(v))
	}

	return out
}

func minresPositions(vs []*big.Rat) []duration.Position {
	if vs == nil {
		return nil
	}

	out := make([]duration.Position, 0, len(vs))
	for _, v := range vs {
		out = append(out, minresPosition(v))
	}

	return out
}

// minresOracle is D₀ of 4.5 for the written quarter bb: the least common
// multiple of the denominators of the reduced v/w, w = bb/32, over all
// values; 1 for no values. ok reports whether D₀ is at most math.MaxInt.
func minresOracle(bb int, values ...[]*big.Rat) (d0 *big.Int, ok bool) {
	d0 = big.NewInt(1)

	for _, vs := range values {
		for _, v := range vs {
			q := new(big.Rat).Mul(v, big.NewRat(32, int64(bb)))
			d := q.Denom()

			gcd := new(big.Int).GCD(nil, nil, d0, d)
			d0.Mul(d0, d)
			d0.Quo(d0, gcd)
		}
	}

	return d0, d0.Cmp(big.NewInt(math.MaxInt)) <= 0
}

// minresQuarter is the variadic argument for bb: nil for the standard quarter
// unless explicit is set.
func minresQuarter(bb int, explicit bool) []ticks.MIDIQuarter {
	if bb == ticks.StandardQuarter().MustThirtySeconds() && !explicit {
		return nil
	}

	return []ticks.MIDIQuarter{ticks.MustNewMIDIQuarter(bb)}
}

// assertMinResolutionOutOfRange checks the refusal Е-8 of MinResolution:
// ticks.ErrOutOfRange without fraction.ErrOutOfRange and without any other
// sentinel of the package, the text of architecture 4.3, and a zero
// Resolution.
func assertMinResolutionOutOfRange(t *testing.T, r ticks.Resolution, err error) {
	t.Helper()
	requireTicksError(t, err, ticks.ErrOutOfRange)
	assert.NotErrorIs(t, err, fraction.ErrOutOfRange, "the cause is an integer outside int, not a fraction")
	assert.ErrorContains(t, err, "result outside int")
	assert.False(t, r.IsSet())
	assert.True(t, r.Equal(ticks.Resolution{}))
}

// assertMinResolution checks MinResolution and MustMinResolution on the
// values against the oracle.
func assertMinResolution(t *testing.T, bb int, explicit bool, ds, ps []*big.Rat) {
	t.Helper()

	want, ok := minresOracle(bb, ds, ps)
	durations, positions := minresDurations(ds), minresPositions(ps)
	q := minresQuarter(bb, explicit)

	r, err := ticks.MinResolution(durations, positions, q...)
	if !ok {
		assertMinResolutionOutOfRange(t, r, err)
		requireMustPanic(t, err, ticks.ErrOutOfRange, func() { ticks.MustMinResolution(durations, positions, q...) })

		return
	}

	require.NoError(t, err)
	assert.Equal(t, int(want.Int64()), r.MustPerQuarter())
	assert.True(t, r.MustQuarter().Equal(ticks.MustNewMIDIQuarter(bb)), "the written quarter is bb = %d", bb)
	assert.True(t, r.Equal(ticks.MustMinResolution(durations, positions, q...)))
}

// TestMinResolutionLCMControl checks the control examples of architecture
// 4.3 (D₀ by the definition of 4.5), as durations, as positions, and both
// together, with a zero position; the oracle confirms each expected value.
func TestMinResolutionLCMControl(t *testing.T) {
	t.Parallel()

	k := minresK()
	one := big.NewRat(1, 1)

	tests := []struct {
		name   string
		bb     int
		values []*big.Rat
		want   int // 0: Е-8
	}{
		{"bb 8, K/1: v/w = 4K outside the fraction range", 8, []*big.Rat{minresRat(k, big.NewInt(1))}, 1},
		{"bb 8, 1/2⁶³: the common denominator 2⁶³ exceeds int", 8, []*big.Rat{minresInv(minresPow2(63))}, 1 << 61},
		{"bb 8, 1/(2⁶³ − 1): the bound math.MaxInt", 8, []*big.Rat{big.NewRat(1, math.MaxInt64)}, math.MaxInt},
		{"bb 8, 1/2⁶⁵: D₀ = 2⁶³ > math.MaxInt", 8, []*big.Rat{minresInv(minresPow2(65))}, 0},
		{"bb 1, 1/2⁶⁷: b = 32", 1, []*big.Rat{minresInv(minresPow2(67))}, 1 << 62},
		{"bb 1, 1/2⁶⁸: D₀ = 2⁶³ > math.MaxInt", 1, []*big.Rat{minresInv(minresPow2(68))}, 0},
		{"bb 3, 1/1", 3, []*big.Rat{one}, 3},
		{"bb 3, 3/1: a common factor with a", 3, []*big.Rat{big.NewRat(3, 1)}, 1},
		{"bb 3, 1/64", 3, []*big.Rat{big.NewRat(1, 64)}, 6},
		{"bb 255, 17/1: a/gcd(a, p)", 255, []*big.Rat{big.NewRat(17, 1)}, 15},
		{"bb 255, 1/K: v/a outside the fraction range", 255, []*big.Rat{minresInv(k)}, 0},
		{"bb 128, 1/1: w = 4, b = 1", 128, []*big.Rat{one}, 4},
		{"bb 128, 2/1", 128, []*big.Rat{big.NewRat(2, 1)}, 2},
		{"bb 8, {3/8, 1/12}: common factors", 8, []*big.Rat{big.NewRat(3, 8), big.NewRat(1, 12)}, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want, ok := minresOracle(tt.bb, tt.values)
			if tt.want == 0 {
				require.False(t, ok, "the oracle gives D₀ = %s", want)
			} else {
				require.True(t, ok, "the oracle gives D₀ = %s", want)
				require.Equal(t, int64(tt.want), want.Int64(), "the oracle")
			}

			zero := []*big.Rat{new(big.Rat)}
			for _, explicit := range []bool{false, true} {
				assertMinResolution(t, tt.bb, explicit, tt.values, nil)
				assertMinResolution(t, tt.bb, explicit, nil, tt.values)
				assertMinResolution(t, tt.bb, explicit, tt.values, zero)
				assertMinResolution(t, tt.bb, explicit, tt.values, tt.values)
			}

			if tt.want != 0 {
				assert.Equal(t, tt.want, ticks.MustMinResolution(minresDurations(tt.values), nil, ticks.MustNewMIDIQuarter(tt.bb)).MustPerQuarter())
			}
		})
	}
}

// TestMinResolutionLCMMixed checks durations and positions together, where
// D₀ comes from both lists, and the zero position, which adds nothing.
func TestMinResolutionLCMMixed(t *testing.T) {
	t.Parallel()

	k := minresK()
	zero := new(big.Rat)

	tests := []struct {
		name   string
		bb     int
		ds, ps []*big.Rat
		want   int // 0: Е-8
	}{
		{"bb 3: 1/64 and the position 3", 3, []*big.Rat{big.NewRat(1, 64)}, []*big.Rat{big.NewRat(3, 1)}, 6},
		{"bb 3: 3/1 and the position 1/64", 3, []*big.Rat{big.NewRat(3, 1)}, []*big.Rat{big.NewRat(1, 64)}, 6},
		{"bb 8: 3/8 and the position 1/12", 8, []*big.Rat{big.NewRat(3, 8)}, []*big.Rat{big.NewRat(1, 12)}, 6},
		{"bb 8: K/1 and the position 1/2⁶³", 8, []*big.Rat{minresRat(k, big.NewInt(1))}, []*big.Rat{minresInv(minresPow2(63))}, 1 << 61},
		{"bb 8: 1/2⁶³ and the position 1/2⁶⁵", 8, []*big.Rat{minresInv(minresPow2(63))}, []*big.Rat{minresInv(minresPow2(65))}, 0},
		{"bb 255: 17/1 and the position 1/K", 255, []*big.Rat{big.NewRat(17, 1)}, []*big.Rat{minresInv(k)}, 0},
		{"bb 255: 1/K and the zero position", 255, []*big.Rat{minresInv(k)}, []*big.Rat{zero}, 0},
		{"bb 255: the zero position only", 255, nil, []*big.Rat{zero}, 1},
		{"bb 1: the zero position only", 1, nil, []*big.Rat{zero, zero}, 1},
		{"bb 8: the zero position and 1/(2⁶³ − 1)", 8, []*big.Rat{big.NewRat(1, math.MaxInt64)}, []*big.Rat{zero}, math.MaxInt},
		{"bb 128: 1/1 and the positions 0, 2", 128, []*big.Rat{big.NewRat(1, 1)}, []*big.Rat{zero, big.NewRat(2, 1)}, 4},
		{"bb 1: 1/2⁶⁷ and the position 1/2⁶⁸", 1, []*big.Rat{minresInv(minresPow2(67))}, []*big.Rat{minresInv(minresPow2(68))}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want, ok := minresOracle(tt.bb, tt.ds, tt.ps)
			if tt.want == 0 {
				require.False(t, ok, "the oracle gives D₀ = %s", want)
			} else {
				require.True(t, ok, "the oracle gives D₀ = %s", want)
				require.Equal(t, int64(tt.want), want.Int64(), "the oracle")
			}

			assertMinResolution(t, tt.bb, true, tt.ds, tt.ps)
			assertMinResolution(t, tt.bb, false, tt.ds, tt.ps)
		})
	}
}

// TestMinResolutionLCMEmpty checks КП-27 for every written quarter: an empty
// set gives D₀ = 1 with that quarter.
func TestMinResolutionLCMEmpty(t *testing.T) {
	t.Parallel()

	for bb := 1; bb <= 255; bb++ {
		q := ticks.MustNewMIDIQuarter(bb)

		r, err := ticks.MinResolution(nil, nil, q)
		require.NoError(t, err, "bb = %d", bb)
		assert.Equal(t, 1, r.MustPerQuarter(), "bb = %d", bb)
		assert.True(t, r.MustQuarter().Equal(q), "bb = %d", bb)

		r, err = ticks.MinResolution([]duration.Duration{}, []duration.Position{}, q)
		require.NoError(t, err, "bb = %d", bb)
		assert.Equal(t, 1, r.MustPerQuarter(), "bb = %d", bb)
	}

	assert.Equal(t, 1, ticks.MustMinResolution(nil, nil).MustPerQuarter())
}

// TestMinResolutionLCMOutOfRangeText checks that every refusal Е-8 of
// MinResolution has one text: when D₀ exceeds int with every v/a in range,
// when v/a itself is outside the fraction range, and when v/w alone is.
func TestMinResolutionLCMOutOfRangeText(t *testing.T) {
	t.Parallel()

	k := minresK()

	calls := []struct {
		name string
		call func() (ticks.Resolution, error)
	}{
		{"bb 8, 1/2⁶⁵", func() (ticks.Resolution, error) {
			return ticks.MinResolution([]duration.Duration{minresDuration(minresInv(minresPow2(65)))}, nil)
		}},
		{"bb 255, 1/K", func() (ticks.Resolution, error) {
			return ticks.MinResolution([]duration.Duration{minresDuration(minresInv(k))}, nil, ticks.MustNewMIDIQuarter(255))
		}},
		{"bb 255, the position 1/K", func() (ticks.Resolution, error) {
			return ticks.MinResolution(nil, []duration.Position{minresPosition(minresInv(k))}, ticks.MustNewMIDIQuarter(255))
		}},
		{"bb 8, primes near 10⁶", func() (ticks.Resolution, error) {
			return ticks.MinResolution([]duration.Duration{dur(1, 1000003), dur(1, 1000033), dur(1, 1000037), dur(1, 1000039)}, nil)
		}},
		{"bb 3, 1/K and 1/2⁶⁵", func() (ticks.Resolution, error) {
			return ticks.MinResolution([]duration.Duration{minresDuration(minresInv(k)), minresDuration(minresInv(minresPow2(65)))}, nil, ticks.MustNewMIDIQuarter(3))
		}},
	}

	var texts []string

	for _, c := range calls {
		r, err := c.call()
		assertMinResolutionOutOfRange(t, r, err)

		if err != nil {
			texts = append(texts, err.Error())
		}
	}

	require.Len(t, texts, len(calls))

	for i, text := range texts {
		assert.Equal(t, texts[0], text, "the text of %q", calls[i].name)
	}
}

// TestMinResolutionLCMOrder checks the order of checks of MinResolution
// before the range: an unset duration, then an unset position, then the
// written quarter (an extra one is Е-2, an unset one is Е-1), and Е-8 last.
// Each combined refusal equals the refusal of its first fault alone; the
// unset elements are at index 1, as in the reference refusals, whose texts
// name the index.
func TestMinResolutionLCMOrder(t *testing.T) {
	t.Parallel()

	k := minresK()
	far := minresDuration(minresInv(minresPow2(65))) // Е-8 with bb 8
	farQ := minresDuration(minresInv(k))             // Е-8 with bb 255, v/a outside the range
	farP := minresPosition(minresInv(minresPow2(65)))
	q255 := ticks.MustNewMIDIQuarter(255)
	two := []ticks.MIDIQuarter{ticks.StandardQuarter(), ticks.StandardQuarter()}
	unsetQ := []ticks.MIDIQuarter{{}}

	call := func(ds []duration.Duration, ps []duration.Position, q ...ticks.MIDIQuarter) error {
		_, err := ticks.MinResolution(ds, ps, q...)

		return err
	}

	unsetDuration := call([]duration.Duration{dur(1, 4), {}}, nil)
	unsetPosition := call(nil, []duration.Position{pos(1, 4), {}})
	extraQuarter := call(nil, nil, two...)
	unsetQuarter := call(nil, nil, unsetQ...)

	requireTicksError(t, unsetDuration, ticks.ErrInvalidArgument, duration.ErrUnset)
	requireTicksError(t, unsetPosition, ticks.ErrInvalidArgument, duration.ErrUnset)
	requireTicksError(t, extraQuarter, ticks.ErrInvalidArgument)
	requireTicksError(t, unsetQuarter, ticks.ErrUnset)

	tests := []struct {
		name string
		ds   []duration.Duration
		ps   []duration.Position
		q    []ticks.MIDIQuarter
		want error
		own  error
	}{
		{"unset duration before an unset position", []duration.Duration{far, {}}, []duration.Position{farP, {}}, nil, unsetDuration, ticks.ErrInvalidArgument},
		{"unset duration after an out-of-range one", []duration.Duration{far, {}}, nil, nil, unsetDuration, ticks.ErrInvalidArgument},
		{"unset duration before an extra quarter", []duration.Duration{farQ, {}}, []duration.Position{farP}, two, unsetDuration, ticks.ErrInvalidArgument},
		{"unset duration before an unset quarter", []duration.Duration{far, {}}, []duration.Position{{}}, unsetQ, unsetDuration, ticks.ErrInvalidArgument},
		{"unset position after out-of-range values", []duration.Duration{far}, []duration.Position{farP, {}}, nil, unsetPosition, ticks.ErrInvalidArgument},
		{"unset position before an extra quarter", []duration.Duration{farQ}, []duration.Position{farP, {}}, two, unsetPosition, ticks.ErrInvalidArgument},
		{"unset position before an unset quarter", nil, []duration.Position{farP, {}}, unsetQ, unsetPosition, ticks.ErrInvalidArgument},
		{"extra quarter before the range", []duration.Duration{far}, []duration.Position{farP}, two, extraQuarter, ticks.ErrInvalidArgument},
		{"extra quarter 255 before the range of v/a", []duration.Duration{farQ}, nil, []ticks.MIDIQuarter{q255, q255}, extraQuarter, ticks.ErrInvalidArgument},
		{"unset quarter before the range", []duration.Duration{far, farQ}, []duration.Position{farP}, unsetQ, unsetQuarter, ticks.ErrUnset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := ticks.MinResolution(tt.ds, tt.ps, tt.q...)
			requireTicksError(t, err, tt.own)
			assert.NotErrorIs(t, err, ticks.ErrOutOfRange)
			assert.Equal(t, tt.want.Error(), err.Error())
			assert.False(t, r.IsSet())
			requireMustPanic(t, err, tt.own, func() { ticks.MustMinResolution(tt.ds, tt.ps, tt.q...) })
		})
	}
}

// minresBits is a random positive integer of exactly n bits, 1 ≤ n ≤ 255.
func minresBits(rng *rand.Rand, n int) *big.Int {
	x := new(big.Int)
	for x.BitLen() < n {
		x.Lsh(x, 64)
		x.Or(x, new(big.Int).SetUint64(rng.Uint64()|1))
	}

	x.Rsh(x, uint(x.BitLen()-n))

	return x.SetBit(x, n-1, 1)
}

// minresSmooth is a random product of small primes.
func minresSmooth(rng *rand.Rand) *big.Int {
	x := big.NewInt(1)
	for _, p := range []int64{2, 3, 5, 7, 11, 13, 17} {
		for range rng.IntN(4) {
			x.Mul(x, big.NewInt(p))
		}
	}

	return x
}

// minresComponent is a random numerator or denominator up to K: small and
// smooth values, powers of two around 2⁶³, values near K, and random values
// of any width.
func minresComponent(rng *rand.Rand, denominator bool) *big.Int {
	k := minresK()

	switch rng.IntN(10) {
	case 0:
		return big.NewInt(1)
	case 1:
		return big.NewInt(1 + rng.Int64N(255))
	case 2, 3:
		return minresSmooth(rng)
	case 4:
		return minresPow2(uint(rng.IntN(72)))
	case 5:
		if denominator {
			return new(big.Int).Mul(minresPow2(uint(55+rng.IntN(15))), big.NewInt(1+rng.Int64N(16)))
		}

		return new(big.Int).Sub(k, big.NewInt(rng.Int64N(64)))
	case 6:
		return big.NewInt(math.MaxInt64 - rng.Int64N(4))
	case 7:
		return new(big.Int).Sub(k, big.NewInt(rng.Int64N(64)))
	default:
		return minresBits(rng, 1+rng.IntN(255))
	}
}

// minresValue is a random value with components up to K; zero is allowed.
func minresValue(rng *rand.Rand, zero bool) *big.Rat {
	if zero && rng.IntN(8) == 0 {
		return new(big.Rat)
	}

	return minresRat(minresComponent(rng, false), minresComponent(rng, true))
}

// minresCase is a generated set of values for one written quarter.
type minresCase struct {
	explicit bool
	ds, ps   []*big.Rat
}

// minresClass classifies the outcome of the oracle for counting: success,
// success with some v/w outside the fraction range, Е-8 with every v/a in
// the fraction range, and Е-8 with some v/a outside it. Here a = bb/gcd(bb,
// 32) is the numerator of the reduced w = bb/32 (architecture 4.3).
func minresClass(bb int, c minresCase) string {
	_, ok := minresOracle(bb, c.ds, c.ps)
	k := minresK()
	a := int64(bb >> min(bits.TrailingZeros(uint(bb)), 5))

	outside := func(q *big.Rat) bool {
		return new(big.Int).Abs(q.Num()).Cmp(k) > 0 || q.Denom().Cmp(k) > 0
	}

	wideW, wideA := false, false

	for _, vs := range [][]*big.Rat{c.ds, c.ps} {
		for _, v := range vs {
			wideW = wideW || outside(new(big.Rat).Mul(v, big.NewRat(32, int64(bb))))
			wideA = wideA || outside(new(big.Rat).Quo(v, big.NewRat(a, 1)))
		}
	}

	switch {
	case ok && wideW:
		return "success, v/w outside the fraction range"
	case ok:
		return "success"
	case wideA:
		return "Е-8, v/a outside the fraction range"
	default:
		return "Е-8, D₀ > math.MaxInt"
	}
}

// TestMinResolutionLCMOracle compares MinResolution with the oracle of 4.5
// on generated sets with a fixed seed for every written quarter bb 1…255:
// components up to K, durations and positions, zero positions, and the
// refusal when D₀ exceeds math.MaxInt.
func TestMinResolutionLCMOracle(t *testing.T) {
	t.Parallel()

	const seed = 2608

	rng := rand.New(rand.NewPCG(seed, seed+1))

	perQuarter := 24
	if testing.Short() {
		perQuarter = 3
	}

	cases := make([][]minresCase, 256)
	counts := map[string]int{}

	for bb := 1; bb <= 255; bb++ {
		for range perQuarter {
			c := minresCase{explicit: rng.IntN(2) == 0}
			for range rng.IntN(5) {
				c.ds = append(c.ds, minresValue(rng, false))
			}

			for range rng.IntN(4) {
				c.ps = append(c.ps, minresValue(rng, true))
			}

			cases[bb] = append(cases[bb], c)
			counts[minresClass(bb, c)]++
		}
	}

	t.Logf("classes of the oracle: %v", counts)

	if !testing.Short() {
		for _, class := range []string{
			"success", "success, v/w outside the fraction range",
			"Е-8, D₀ > math.MaxInt", "Е-8, v/a outside the fraction range",
		} {
			assert.Positive(t, counts[class], "the generated sets include the class %q", class)
		}
	}

	// The sets are generated above, so -run selects one bb or one set with
	// the same values.
	for bb := 1; bb <= 255; bb++ {
		t.Run(fmt.Sprintf("bb=%d", bb), func(t *testing.T) {
			t.Parallel()

			for i, c := range cases[bb] {
				t.Run(fmt.Sprintf("set=%d", i), func(t *testing.T) {
					defer func() {
						if t.Failed() {
							t.Logf("bb %d, set %d (seed %d): durations %v, positions %v", bb, i, seed, c.ds, c.ps)
						}
					}()

					assertMinResolution(t, bb, c.explicit, c.ds, c.ps)
				})
			}
		})
	}
}
