package score_test

import (
	"math"
	"math/big"
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

// K-11 and K-12: actual finite boundaries are distinct from a period.
func TestPolymetricBarlines(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name              string
		n, ac, bc, period int
		pickup            bool
	}{
		{"three against four", 3, 4, 3, 3, false},
		{"seven against four", 7, 4, 7, 7, false},
		{"same pickup", 3, 4, 3, 3, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := emptyPart(tc.n, 4, tc.ac), emptyPart(4, 4, tc.bc)
			end := pos(tc.period, 1)
			if tc.pickup {
				ga, gb := a.MustGrid().MustWithPickup(dur(1, 4)), b.MustGrid().MustWithPickup(dur(1, 4))
				a = a.MustEdit(track.Edit{Grid: &ga})
				b = b.MustEdit(track.Edit{Grid: &gb})
				end = pos(13, 4)
			}
			s, ids := assemble(clock60(), a, b)
			got, err := s.CommonBarlines(pos(0, 1), end, ids[1], ids[0])
			require.NoError(t, err)
			want := []duration.Position{pos(0, 1), end}
			if tc.pickup {
				want = []duration.Position{pos(0, 1), pos(1, 4), end}
			}
			require.Len(t, got, len(want))
			for i, p := range want {
				assertPos(t, p, got[i].Position)
				require.Len(t, got[i].Boundaries, 2)
				assertKey(t, ids[0], got[i].Boundaries[0].Part)
				assertKey(t, ids[1], got[i].Boundaries[1].Part)
			}
			for _, boundary := range got[0].Boundaries {
				assert.Nil(t, boundary.PreviousBar)
				require.NotNil(t, boundary.NextBar)
				expected := 1
				if tc.pickup {
					expected = 0
				}
				assert.Equal(t, expected, *boundary.NextBar)
			}
			last := got[len(got)-1].Boundaries
			for i, wantBar := range []int{tc.ac, tc.bc} {
				require.NotNil(t, last[i].PreviousBar)
				assert.Equal(t, wantBar, *last[i].PreviousBar)
				assert.Nil(t, last[i].NextBar)
			}
			assert.Len(t, s.MustCommonBarlines(end, end, ids...), 1)
			assert.Empty(t, s.MustCommonBarlines(pos(1, 2), pos(1, 2), ids...))
			period, err := s.BarlinePeriod(ids[0], ids[1])
			if tc.pickup {
				require.ErrorIs(t, err, score.ErrPeriod)
				assert.False(t, period.IsSet())
			} else {
				require.NoError(t, err)
				assert.True(t, period.Equal(dur(tc.period, 1)))
			}
		})
	}
}

func TestBarlineShortAndMultipleGrids(t *testing.T) {
	t.Parallel()
	s, ids := assemble(clock60(), emptyPart(4, 4, 1), emptyPart(4, 4, 3), emptyPart(2, 4, 6))
	lines := s.MustCommonBarlines(pos(0, 1), pos(3, 1), ids[2], ids[0], ids[1])
	require.Len(t, lines, 2)
	assertPos(t, pos(1, 1), lines[1].Position)
	assert.Nil(t, lines[1].Boundaries[0].NextBar)
	assert.Equal(t, 2, *lines[1].Boundaries[1].NextBar)
	assert.Equal(t, 3, *lines[1].Boundaries[2].NextBar)
	assert.Empty(t, s.MustCommonBarlines(pos(3, 2), pos(3, 1), ids...))
	assert.True(t, s.MustBarlinePeriod(ids[0], ids[2]).Equal(dur(1, 1)))
	// Period is allowed to exceed the total finite duration.
	s, ids = assemble(clock60(), emptyPart(3, 4, 1), emptyPart(4, 4, 1))
	assert.True(t, s.MustBarlinePeriod(ids[0], ids[1]).Equal(dur(3, 1)))
}

func TestCommonBarlinesFollowIncompleteAndChangingBars(t *testing.T) {
	t.Parallel()
	type expectedLine struct {
		position       duration.Position
		previous, next [2]int // -1 means that the boundary has no adjacent bar.
	}
	for _, tc := range []struct {
		name   string
		a      measure.Grid
		bCount int
		end    duration.Position
		absent duration.Position
		want   []expectedLine
	}{
		{
			name:   "explicit incomplete first bar",
			a:      grid(4, 4, 2).MustWithLength(1, dur(1, 2), measure.AlignStart),
			bCount: 3, end: pos(3, 2), absent: pos(1, 1),
			want: []expectedLine{
				{pos(0, 1), [2]int{-1, -1}, [2]int{1, 1}},
				{pos(1, 2), [2]int{1, 1}, [2]int{2, 2}},
				{pos(3, 2), [2]int{2, 3}, [2]int{-1, -1}},
			},
		},
		{
			name:   "meter change after first bar",
			a:      grid(4, 4, 3).MustWithMark(2, meter.MustNew(3, 4)),
			bCount: 5, end: pos(5, 2), absent: pos(2, 1),
			want: []expectedLine{
				{pos(0, 1), [2]int{-1, -1}, [2]int{1, 1}},
				{pos(1, 1), [2]int{1, 2}, [2]int{2, 3}},
				{pos(5, 2), [2]int{3, 5}, [2]int{-1, -1}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := track.MustNew(instrument.MustNew(instrument.Piano()), tc.a)
			s, ids := assemble(clock60(), a, emptyPart(2, 4, tc.bCount))
			got, err := s.CommonBarlines(pos(0, 1), tc.end, ids[1], ids[0])
			require.NoError(t, err)
			require.Len(t, got, len(tc.want))
			for i, expected := range tc.want {
				assertPos(t, expected.position, got[i].Position)
				require.Len(t, got[i].Boundaries, 2)
				for part, boundary := range got[i].Boundaries {
					assertKey(t, ids[part], boundary.Part)
					for _, adjacent := range []struct {
						want int
						got  *int
					}{
						{expected.previous[part], boundary.PreviousBar},
						{expected.next[part], boundary.NextBar},
					} {
						if adjacent.want == -1 {
							assert.Nil(t, adjacent.got)
						} else {
							require.NotNil(t, adjacent.got)
							assert.Equal(t, adjacent.want, *adjacent.got)
						}
					}
				}
			}
			// Neither finite grid has a common boundary at the first-meter period.
			assert.Empty(t, s.MustCommonBarlines(tc.absent, tc.absent, ids...))
		})
	}
}

func TestBarlinePeriodAdmissibility(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		grid measure.Grid
		fail bool
	}{
		{"equivalent explicit length", grid(4, 4, 2).MustWithLength(1, dur(1, 1), measure.AlignStart), false},
		{"repeat equal length notation", grid(3, 4, 2).MustWithMark(2, meter.MustNew(6, 8)), false},
		{"short bar", grid(4, 4, 2).MustWithLength(1, dur(1, 2), measure.AlignStart), true},
		{"variable length", grid(4, 4, 2).MustWithMark(2, meter.MustNew(3, 4)), true},
		{"pickup", grid(4, 4, 2).MustWithPickup(dur(1, 4)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := track.MustNew(instrument.MustNew(instrument.Piano()), tc.grid)
			s, ids := assemble(clock60(), p, emptyPart(4, 4, 1))
			got, err := s.BarlinePeriod(ids[0], ids[1])
			if tc.fail {
				require.ErrorIs(t, err, score.ErrPeriod)
				assert.False(t, got.IsSet())
				assertKey(t, ids[0], problems(t, err).Issues[0].Part)
			} else {
				require.NoError(t, err)
				assert.True(t, got.IsSet())
			}
		})
	}
}

func TestBarlinePeriodUsesExactLargeIntegers(t *testing.T) {
	t.Parallel()
	// Sum meters create reduced lengths whose numerators exceed machine ints.
	// A larger pair has a valid individual length but an unrepresentable LCM.
	primes := make([]int, 10)
	candidate := int64(1_000_000_000_001)
	for i := range primes {
		for !big.NewInt(candidate).ProbablyPrime(20) {
			candidate += 2
		}
		primes[i] = int(candidate)
		candidate += 2
	}
	for _, count := range []int{3, 5} {
		t.Run(big.NewInt(int64(count)).String()+" components", func(t *testing.T) {
			parts := make([]*track.Track, 2)
			lengths := make([]*big.Rat, 2)
			for side := range 2 {
				terms := make([]meter.Meter, count)
				lengths[side] = new(big.Rat)
				for i := range count {
					den := primes[side*5+i]
					terms[i] = meter.MustNew(1, den)
					lengths[side].Add(lengths[side], big.NewRat(1, int64(den)))
				}
				g := measure.MustNew(meter.MustNewSum(terms...)).MustWithBarCount(1)
				parts[side] = track.MustNew(instrument.MustNew(instrument.Piano()), g)
			}
			a, c := lengths[0].Num(), lengths[1].Num()
			b, d := lengths[0].Denom(), lengths[1].Denom()
			gcd := new(big.Int).GCD(nil, nil, a, c)
			lcm := new(big.Int).Quo(new(big.Int).Mul(a, c), gcd)
			den := new(big.Int).GCD(nil, nil, b, d)
			want := new(big.Rat).SetFrac(lcm, den)
			s, ids := assemble(clock60(), parts...)
			got, err := s.BarlinePeriod(ids[0], ids[1])
			if count == 5 {
				require.Greater(t, want.Num().BitLen(), fraction.MaxComponentBits)
				require.ErrorIs(t, err, score.ErrOutOfRange)
				require.ErrorIs(t, err, fraction.ErrOutOfRange)
				assert.False(t, got.IsSet())
			} else {
				require.NoError(t, err)
				require.Greater(t, want.Num().BitLen(), 64)
				expected := duration.MustParse(want.Num().String() + "/" + want.Denom().String())
				assert.True(t, got.Equal(expected))
			}
		})
	}
}

func TestMaxIntFiniteGridBoundaries(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		grid measure.Grid
	}{
		{"last meter mark", grid(4, 4, math.MaxInt).MustWithMark(math.MaxInt, meter.MustNew(4, 4))},
		{"last explicit full length", grid(4, 4, math.MaxInt).MustWithLength(math.MaxInt, dur(1, 1), measure.AlignStart)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := track.MustNew(instrument.MustNew(instrument.Piano()), tc.grid)
			s, ids := assemble(clock60(), p, p)
			t.Run("equal replacement", func(t *testing.T) {
				var replaced *score.Score
				var err error
				require.NotPanics(t, func() { replaced, err = s.Replace(ids[0], p) })
				require.NoError(t, err)
				assert.True(t, replaced.MustPart(ids[0]).Equal(p))
				assertKeys(t, ids, replaced.MustParts())
				assertPos(t, pos(math.MaxInt, 1), replaced.MustEnd())
			})
			t.Run("constant full bar period", func(t *testing.T) {
				var period duration.Duration
				var err error
				require.NotPanics(t, func() { period, err = s.BarlinePeriod(ids[0], ids[1]) })
				require.NoError(t, err)
				assert.True(t, period.Equal(dur(1, 1)))
			})
			assert.True(t, s.MustPart(ids[0]).Equal(p))
		})
	}
}
