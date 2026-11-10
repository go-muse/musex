package system_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fuzz targets with independent oracles (plan, section 4; risks R-2, R-3).

func ratOf(v alteration.Alteration) *big.Rat {
	num, den := v.Fraction()
	return big.NewRat(int64(num), int64(den))
}

// FuzzProgressionContains checks membership and bounds of a progression against a
// rational oracle.
func FuzzProgressionContains(f *testing.F) {
	f.Add(0, uint8(0), uint8(1), 0, 0, uint8(0), 0, uint8(1))
	f.Add(1, uint8(1), uint8(2), -3, 3, uint8(3), 1, uint8(2))
	f.Add(1, uint8(2), uint8(3), 0, 1, uint8(1), 2, uint8(3))
	f.Add(5, uint8(3), uint8(0), -2, 2, uint8(3), 5, uint8(4))
	f.Add(-7, uint8(5), uint8(6), -1, 0, uint8(2), -7, uint8(6))
	f.Add(1, uint8(1), uint8(1), 1, 0, uint8(3), 0, uint8(1))

	f.Fuzz(func(t *testing.T, startNum int, startDenRaw, dRaw uint8, loNum, hiNum int, flags uint8, qNum int, qDenRaw uint8) {
		// Small values: denominators from 1 to 6 and 1 to 24, numerators within ±60.
		startDen := 1 + int(startDenRaw%6)
		d := int(dRaw % 7) // 0 means a single value
		startNum %= 60
		loNum %= 60
		hiNum %= 60
		qNum %= 200
		qDen := 1 + int(qDenRaw%24)
		const boundDen = 4

		start := q(startNum, startDen)
		step := a(0)
		if d > 0 {
			step = q(1, d)
		}
		p, err := system.NewProgression(start, step)
		require.NoError(t, err)

		startR := ratOf(start)
		var lo, hi *big.Rat

		onLattice := func(x *big.Rat) bool {
			if d == 0 {
				return x.Cmp(startR) == 0
			}
			diff := new(big.Rat).Sub(x, startR)
			diff.Mul(diff, big.NewRat(int64(d), 1))
			return diff.IsInt()
		}
		inside := func(x *big.Rat) bool {
			if lo != nil && x.Cmp(lo) < 0 {
				return false
			}
			if hi != nil && x.Cmp(hi) > 0 {
				return false
			}
			return onLattice(x)
		}
		// nonEmpty reports whether some lattice point lies in [lo, hi].
		nonEmpty := func() bool {
			if d == 0 {
				return inside(startR)
			}
			if lo == nil || hi == nil {
				return true
			}
			// Indices n with lo <= start + n/d <= hi.
			l := new(big.Rat).Mul(new(big.Rat).Sub(lo, startR), big.NewRat(int64(d), 1))
			h := new(big.Rat).Mul(new(big.Rat).Sub(hi, startR), big.NewRat(int64(d), 1))
			ceil := new(big.Int).Neg(new(big.Int).Div(new(big.Int).Neg(l.Num()), l.Denom()))
			floor := new(big.Int).Div(h.Num(), h.Denom())
			return ceil.Cmp(floor) <= 0
		}

		if flags&1 != 0 {
			lo = big.NewRat(int64(loNum), boundDen)
			r, err := p.WithLower(q(loNum, boundDen))
			if !nonEmpty() {
				requireErr(t, err, system.ErrDefinition, system.ErrRule)
				return
			}
			require.NoError(t, err)
			p = r
		}
		if flags&2 != 0 {
			hi = big.NewRat(int64(hiNum), boundDen)
			r, err := p.WithUpper(q(hiNum, boundDen))
			if !nonEmpty() {
				requireErr(t, err, system.ErrDefinition, system.ErrRule)
				return
			}
			require.NoError(t, err)
			p = r
		}

		x := q(qNum, qDen)
		assert.Equal(t, inside(ratOf(x)), p.Contains(x), "%s contains %s", p, x)
		assert.True(t, p.Contains(p.First()), "%s contains First", p)
		assert.True(t, p.Contains(p.Last()), "%s contains Last", p)
		assert.True(t, inside(ratOf(p.First())) && inside(ratOf(p.Last())))
		assert.LessOrEqual(t, p.First().Compare(p.Last()), 0)
	})
}

// FuzzRuleSubsetOf checks Rule.SubsetOf and Rule.Equal against the enumeration oracle.
func FuzzRuleSubsetOf(f *testing.F) {
	f.Add(uint16(0), uint16(1), uint16(2), uint16(3))
	f.Add(uint16(5), uint16(5), uint16(17), uint16(40))
	f.Add(uint16(100), uint16(7), uint16(64), uint16(13))

	rules := oracleRules()
	f.Fuzz(func(t *testing.T, i, j, k, l uint16) {
		x := append(append(oracleRule{}, rules[int(i)%len(rules)]...), rules[int(k)%len(rules)]...)
		y := append(append(oracleRule{}, rules[int(j)%len(rules)]...), rules[int(l)%len(rules)]...)
		bx, by := x.build(t), y.build(t)

		got, err := bx.SubsetOf(by)
		require.NoError(t, err)
		assert.Equal(t, x.subsetOf(y), got, "%v ⊆ %v", x, y)
		assert.Equal(t, x.subsetOf(y) && y.subsetOf(x), bx.Equal(by), "%v = %v", x, y)
	})
}

// FuzzWrittenRoundTrip checks the written register against P-7 for any octave.
func FuzzWrittenRoundTrip(f *testing.F) {
	f.Add(uint8(0), uint8(6), int16(1), 3)
	f.Add(uint8(0), uint8(0), int16(-1), math.MinInt)
	f.Add(uint8(1), uint8(6), int16(1), math.MaxInt)
	f.Add(uint8(2), uint8(5), int16(-1), 4)
	f.Add(uint8(3), uint8(1), int16(1), -7)

	models := []model{modelT(), modelS19(), modelSA(), modelS24()}
	f.Fuzz(func(t *testing.T, sysIdx, letterIdx uint8, altRaw int16, w int) {
		m := models[int(sysIdx)%len(models)]
		s := m.build(t)
		l := letter.MustFromOrdinal(int(letterIdx) % letter.Count)
		var n spelling.Spelling
		if m.c == 2 {
			n = spq(l, int(altRaw), 2)
		} else {
			n = spi(l, int(altRaw))
		}
		k := m.carry(t, n)

		p, err := s.FromWritten(n, o(w))
		if !shiftFits(w, k) {
			requireErr(t, err, system.ErrOutOfRange, nil)
		} else {
			require.NoError(t, err)
			assert.Equal(t, pt(n, w+k), p)
			assert.Equal(t, o(w), s.MustWrittenOctave(p))
		}

		got, err := s.WrittenOctave(pt(n, w))
		if !shiftFits(w, -k) {
			requireErr(t, err, system.ErrOutOfRange, nil)
		} else {
			require.NoError(t, err)
			assert.Equal(t, o(w-k), got)
			assert.Equal(t, pt(n, w), s.MustFromWritten(n, got))
		}
	})
}

// FuzzDistanceMove checks Distance against P-8 and Move against Distance.
func FuzzDistanceMove(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(2), int8(0), int8(0), int32(4), int32(5), int32(16))
	f.Add(uint8(1), uint8(6), uint8(0), int8(1), int8(-1), int32(5), int32(5), int32(-1))
	f.Add(uint8(2), uint8(5), uint8(4), int8(-1), int8(1), int32(3), int32(3), int32(0))
	f.Add(uint8(3), uint8(3), uint8(3), int8(2), int8(-2), int32(-100), int32(100), int32(1000))
	f.Add(uint8(4), uint8(1), uint8(6), int8(5), int8(-5), int32(math.MinInt32), int32(math.MaxInt32), int32(math.MinInt32))

	models := []model{modelT(), modelS19(), modelSA(), modelS14(), modelS2()}
	f.Fuzz(func(t *testing.T, sysIdx, l1, l2 uint8, a1, a2 int8, o1, o2, n int32) {
		m := models[int(sysIdx)%len(models)]
		s := m.build(t)
		x := pitch.MustNew(spi(letter.MustFromOrdinal(int(l1)%letter.Count), int(a1)), o(int(o1)))
		y := pitch.MustNew(spi(letter.MustFromOrdinal(int(l2)%letter.Count), int(a2)), o(int(o2)))

		d, err := s.Distance(x, y)
		require.NoError(t, err)
		assert.Equal(t, system.Step(m.position(t, y)-m.position(t, x)), d)
		assert.Equal(t, -d, s.MustDistance(y, x))
		c := s.MustCompare(x, y)
		switch {
		case d > 0:
			assert.Equal(t, -1, c)
		case d < 0:
			assert.Equal(t, 1, c)
		default:
			assert.Equal(t, 0, c)
		}

		for _, rule := range []system.SpellingRule{system.Sharps(), system.Flats()} {
			r, err := s.Move(x, system.Step(n), rule)
			require.NoError(t, err)
			assert.Equal(t, system.Step(n), s.MustDistance(x, r))
			want := m.class(t, x.Spelling()) + int(n)
			want -= floorDiv(want, m.n) * m.n
			assert.Equal(t, cls(want, m.n), s.MustPitchClass(r))
		}
	})
}
