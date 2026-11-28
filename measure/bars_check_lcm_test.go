package measure_test

import (
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file check the Е-10 criterion of Bars (architecture 4.6,
// "Перебор") at the bounds of its second item, where D = lcm(den S, den L)
// is computed, and on generated grids against a model of the criterion on
// math/big. The model, lcmModel, computes the lengths, starts, and shifts of
// the bars from the record of the grid by the rules of 4.6 (ИС-8) and applies
// the criterion literally: runs of equal length, item 1, item 2, and the
// shifts. It does not use the package.

// lcmK is K = fraction.MaxComponent.
func lcmK() *big.Int { return fraction.MaxComponentBig() }

// lcmInRange reports whether the reduced components of r are at most K.
func lcmInRange(r *big.Rat) bool {
	k := lcmK()

	return new(big.Int).Abs(r.Num()).Cmp(k) <= 0 && r.Denom().Cmp(k) <= 0
}

// lcmMeter is a meter with its exact length.
type lcmMeter struct {
	meter  meter.Meter
	length *big.Rat
}

func lcmSimple(n, d int) lcmMeter {
	return lcmMeter{meter.MustNew(n, d), big.NewRat(int64(n), int64(d))}
}

// lcmMark is the meter m from bar on.
type lcmMark struct {
	bar int
	m   lcmMeter
}

// lcmLength is an explicit length of a bar; end selects AlignEnd.
type lcmLength struct {
	bar    int
	length *big.Rat
	end    bool
}

// lcmSpec is the record of a grid: the meter of bar 1, an optional pickup,
// marks of bars from 2, explicit lengths, and the number of bars (0: open).
type lcmSpec struct {
	first   lcmMeter
	pickup  *big.Rat
	marks   []lcmMark
	lengths []lcmLength
	bars    int
}

// lcmRun repeats length for the bars from, from + 1, …, to.
func lcmRun(from, to int, length *big.Rat) []lcmLength {
	out := make([]lcmLength, 0, to-from+1)
	for bar := from; bar <= to; bar++ {
		out = append(out, lcmLength{bar: bar, length: length})
	}

	return out
}

// lcmBuild builds the grid of spec by New, WithPickup, WithMark, WithLength,
// and WithBarCount in this order and returns it with the model of the
// applied record. With tolerant, a step refused with ErrLongerThanMeter is
// dropped from the record; any other refusal fails the test.
func lcmBuild(t testing.TB, spec lcmSpec, tolerant bool) (measure.Grid, lcmModel) {
	t.Helper()

	g, err := measure.New(spec.first.meter)
	require.NoError(t, err)

	model := lcmModel{first: spec.first.length, marks: map[int]*big.Rat{}, lengths: map[int]lcmLength{}}

	apply := func(next measure.Grid, err error, what string) bool {
		if err == nil {
			g = next

			return true
		}

		require.True(t, tolerant, "%s: %v", what, err)
		require.ErrorIs(t, err, measure.ErrLongerThanMeter, what)

		return false
	}

	if spec.pickup != nil {
		next, err := g.WithPickup(durRat(spec.pickup))
		if apply(next, err, "WithPickup") {
			model.pickup = spec.pickup
		}
	}

	for _, m := range spec.marks {
		next, err := g.WithMark(m.bar, m.m.meter)
		if apply(next, err, "WithMark") {
			model.marks[m.bar] = m.m.length
		}
	}

	for _, l := range spec.lengths {
		align := measure.AlignStart
		if l.end {
			align = measure.AlignEnd
		}

		next, err := g.WithLength(l.bar, durRat(l.length), align)
		if apply(next, err, "WithLength") {
			model.lengths[l.bar] = l
		}
	}

	if spec.bars > 0 {
		next, err := g.WithBarCount(spec.bars)
		require.NoError(t, err)

		g = next
	}

	for bar := range model.marks {
		model.horizon = max(model.horizon, bar)
	}

	for bar := range model.lengths {
		model.horizon = max(model.horizon, bar)
	}

	return g, model
}

// lcmModel is the record of a grid in exact values. Beyond the bar horizon
// there are no marks and no explicit lengths.
type lcmModel struct {
	first   *big.Rat
	pickup  *big.Rat
	marks   map[int]*big.Rat
	lengths map[int]lcmLength
	horizon int
}

// meterAt is the length of the meter in effect at bar k; the pickup has the
// meter of bar 1.
func (m lcmModel) meterAt(k int) *big.Rat {
	best, length := 1, m.first

	for bar, l := range m.marks {
		if bar <= k && bar > best {
			best, length = bar, l
		}
	}

	return length
}

// bar is the length and the shift of bar k (ИС-8).
func (m lcmModel) bar(k int) (length, shift *big.Rat) {
	if k == 0 {
		return m.pickup, new(big.Rat).Sub(m.first, m.pickup)
	}

	meterLength := m.meterAt(k)

	l, ok := m.lengths[k]
	if !ok {
		return meterLength, new(big.Rat)
	}

	if l.end {
		return l.length, new(big.Rat).Sub(meterLength, l.length)
	}

	return l.length, new(big.Rat)
}

// start is the start of bar k: the pickup and the lengths of bars 1…k−1.
func (m lcmModel) start(k int) *big.Rat {
	s := new(big.Rat)
	if k == 0 {
		return s
	}

	if m.pickup != nil {
		s.Add(s, m.pickup)
	}

	for j := 1; j < k && j <= m.horizon; j++ {
		l, _ := m.bar(j)
		s.Add(s, l)
	}

	if k-1 > m.horizon {
		tail := m.meterAt(m.horizon + 1)
		s.Add(s, new(big.Rat).Mul(tail, ratInt(big.NewInt(int64(k-1-m.horizon)))))
	}

	return s
}

// lcmReason is the reason of a refusal by the criterion.
type lcmReason int

const (
	lcmAccepted lcmReason = iota
	lcmByShift            // a nonzero shift out of range
	lcmByEnds             // item 1: S or F of a run out of range
	lcmByItem2            // item 2: n ≥ 2 and D > K or (F − L)·D > K
)

// lcmRunOf is a run of equal length: S, L, n, F, and D.
type lcmRunOf struct {
	s, l, f *big.Rat
	n, d    *big.Int
}

// runs splits first…last into runs of equal length (4.6).
func (m lcmModel) runs(first, last int) []lcmRunOf {
	var out []lcmRunOf

	for b := first; ; {
		l, _ := m.bar(b)

		e := b

		for e < last {
			if e+1 > m.horizon {
				if m.meterAt(m.horizon+1).Cmp(l) == 0 {
					e = last
				}

				break
			}

			next, _ := m.bar(e + 1)
			if next.Cmp(l) != 0 {
				break
			}

			e++
		}

		n := new(big.Int).Add(big.NewInt(int64(e-b)), big.NewInt(1))
		s := m.start(b)
		f := new(big.Rat).Add(s, new(big.Rat).Mul(l, ratInt(n)))

		d := new(big.Int).Mul(s.Denom(), l.Denom())
		d.Quo(d, new(big.Int).GCD(nil, nil, s.Denom(), l.Denom()))

		out = append(out, lcmRunOf{s: s, l: l, f: f, n: n, d: d})

		if e == last {
			return out
		}

		b = e + 1
	}
}

// top is (F − L)·D of a run, an integer.
func (r lcmRunOf) top() *big.Int {
	v := new(big.Rat).Sub(r.f, r.l)
	v.Mul(v, ratInt(r.d))

	if !v.IsInt() {
		panic("(F − L)·D is not an integer")
	}

	return new(big.Int).Set(v.Num())
}

// refuses applies the criterion of 4.6 to first…last.
func (m lcmModel) refuses(first, last int) lcmReason {
	for b := first; b <= last && b <= m.horizon+1; b++ {
		if b == 0 && m.pickup == nil {
			continue
		}

		if _, shift := m.bar(b); shift.Sign() != 0 && !lcmInRange(shift) {
			return lcmByShift
		}
	}

	k := lcmK()
	runs := m.runs(first, last)

	for _, r := range runs {
		if !lcmInRange(r.s) || !lcmInRange(r.f) {
			return lcmByEnds
		}
	}

	for _, r := range runs {
		if r.n.Cmp(big.NewInt(2)) >= 0 && (r.d.Cmp(k) > 0 || r.top().Cmp(k) > 0) {
			return lcmByItem2
		}
	}

	return lcmAccepted
}

// assertBarsVerdict checks Bars(first, last) against the verdict of the
// model: Е-10 with fraction.ErrOutOfRange and a nil sequence on a refusal;
// otherwise no error, the first steps equal Bar(k), and MustBars agrees.
func assertBarsVerdict(t *testing.T, g measure.Grid, first, last int, refused bool) {
	t.Helper()

	if refused {
		assertBarsFail(t, g, first, last)

		return
	}

	assertBarsAgree(t, g, first, last, 4)

	for b := range g.MustBars(first, last) {
		want, err := g.Bar(first)
		require.NoError(t, err)
		assert.Equal(t, want, b, "the first step of MustBars equals Bar(%d)", first)

		break
	}
}

// Values at the bound of item 2.

// lcmP9 is (2²⁵⁵ + 1)/9, an integer prime to 3: lcm(lcmP9, 9) = K + 2.
func lcmP9() *big.Int {
	p := new(big.Int).Add(pow2(255), big.NewInt(1))

	return p.Quo(p, big.NewInt(9))
}

// lcmLambda is λ = K/7, an integer: 7 divides 2²⁵⁵ − 1.
func lcmLambda() *big.Int { return new(big.Int).Quo(lcmK(), big.NewInt(7)) }

// lcmB is B = 2²⁵³ + 3, odd and prime to 3, with 3·B ≤ K < 6·B.
func lcmB() *big.Int { return new(big.Int).Add(pow2(253), big.NewInt(3)) }

func lcmInv(n *big.Int) *big.Rat { return ratFrac(big.NewInt(1), n) }

// TestBarsCheckLCMBounds checks the criterion of Bars where it depends on D:
// D = K and D > K, D smaller than the product of the denominators of S and L
// while the product is greater than K, (F − L)·D = K and > K, S = 0,
// den S | den L and den L | den S, runs of one bar, runs of a pickup, of
// explicit lengths, and of meter lengths, and runs across a change of meter
// without a change of length. Each case also names the values of S, L, D,
// and (F − L)·D of the first run, which the model confirms.
func TestBarsCheckLCMBounds(t *testing.T) {
	t.Parallel()

	k := lcmK()
	lambda16 := ratFrac(lcmLambda(), big.NewInt(16))
	l250 := ratFrac(lcmK(), pow2(250))
	twoB := new(big.Int).Lsh(lcmB(), 1)
	shortEnd := lcmInv(new(big.Int).Add(pow2(254), big.NewInt(1)))

	pickupK := lcmSpec{first: lcmSimple(1, 7), pickup: lcmInv(k)}
	pickupP9 := lcmSpec{first: lcmSimple(1, 9), pickup: lcmInv(lcmP9())}
	smallD := lcmSpec{first: lcmSimple(1, 1<<62), pickup: lcmInv(new(big.Int).Mul(big.NewInt(3), pow2(200)))}
	explicit := lcmSpec{first: lcmSimple(4, 4), lengths: lcmRun(1, 16, lambda16)}
	pickupRun := lcmSpec{first: lcmSimple(33, 1), pickup: l250, lengths: lcmRun(1, 15, l250)}
	sDividesL := lcmSpec{first: lcmSimple(4, 4), pickup: lcmInv(pow2(200)), lengths: lcmRun(1, 4, lcmInv(pow2(254)))}
	lDividesS := lcmSpec{first: lcmSimple(1, 1<<62), pickup: lcmInv(pow2(254))}
	oneBar := lcmSpec{first: lcmSimple(4, 4), pickup: big.NewRat(1, 6), lengths: lcmRun(1, 2, lcmInv(twoB))}
	acrossK := lcmSpec{
		first: lcmSimple(1, 7), pickup: lcmInv(k),
		marks:   []lcmMark{{3, lcmSimple(2, 14)}},
		lengths: lcmRun(5, 5, big.NewRat(1, 7)),
	}
	acrossP9 := lcmSpec{
		first: lcmSimple(1, 9), pickup: lcmInv(lcmP9()),
		marks:   []lcmMark{{4, lcmSimple(2, 18)}},
		lengths: lcmRun(6, 6, big.NewRat(1, 9)),
	}
	shiftEnd := lcmSpec{first: lcmSimple(3, 4), lengths: []lcmLength{{bar: 1, length: shortEnd, end: true}}}

	tests := []struct {
		name        string
		spec        lcmSpec
		first, last int
		reason      lcmReason
		d, top      *big.Int // D and (F − L)·D of the first run, if named
	}{
		{"D = K: meter 1/7 after the pickup 1/K, two bars", pickupK, 1, 2, lcmAccepted, k, nil},
		{"D = K, den S · den L = 7K > K: six bars", pickupK, 1, 6, lcmAccepted, k, nil},
		{"D = K: with the pickup", pickupK, 0, 6, lcmAccepted, nil, nil},
		{"D = K: the end of bar 7 is out of range", pickupK, 1, 7, lcmByEnds, k, nil},
		{"D = K + 2: meter 1/9 after the pickup (2²⁵⁵+1)/9⁻¹", pickupP9, 1, 9, lcmByItem2, new(big.Int).Add(k, big.NewInt(2)), nil},
		{"D = K + 2: one bar, the end is out of range", pickupP9, 1, 1, lcmByEnds, nil, nil},
		{"D = 3·2²⁰⁰ smaller than den S · den L = 3·2²⁶²", smallD, 1, 1000, lcmAccepted, new(big.Int).Mul(big.NewInt(3), pow2(200)), nil},
		{"D = 3·2²⁰⁰ with the pickup", smallD, 0, 1000, lcmAccepted, nil, nil},
		{"S = 0, (F − L)·D = K: explicit lengths λ/16", explicit, 1, 8, lcmAccepted, big.NewInt(16), k},
		{"S = 0, (F − L)·D = 15λ > K, F = λ in range", explicit, 1, 16, lcmByItem2, big.NewInt(16), new(big.Int).Mul(big.NewInt(15), lcmLambda())},
		{"S = 0, (F − L)·D = 11λ > K, F = 3λ/4 in range", explicit, 1, 12, lcmByItem2, nil, nil},
		{"S = 0: the end 9λ/16 is out of range", explicit, 1, 9, lcmByEnds, nil, nil},
		{"pickup run: (F − L)·D = K with L = K/2²⁵⁰", pickupRun, 0, 1, lcmAccepted, pow2(250), k},
		{"pickup run: (F − L)·D = 15K", pickupRun, 0, 15, lcmByItem2, pow2(250), new(big.Int).Mul(big.NewInt(15), k)},
		{"pickup run: the end 3K/2²⁵⁰ is out of range", pickupRun, 0, 2, lcmByEnds, nil, nil},
		{"pickup run of one bar", pickupRun, 0, 0, lcmAccepted, nil, nil},
		{"den S | den L: S = 2⁻²⁰⁰, L = 2⁻²⁵⁴", sDividesL, 1, 4, lcmAccepted, pow2(254), nil},
		{"den S | den L with the pickup", sDividesL, 0, 4, lcmAccepted, nil, nil},
		{"den L | den S: S = 2⁻²⁵⁴, L = 2⁻⁶²", lDividesS, 1, 1 << 20, lcmAccepted, pow2(254), nil},
		{"den L | den S with the pickup", lDividesS, 0, 1 << 20, lcmAccepted, nil, nil},
		{"one bar with D = 6B > K is accepted", oneBar, 1, 1, lcmAccepted, new(big.Int).Mul(big.NewInt(3), twoB), nil},
		{"one bar whose end is out of range", oneBar, 2, 2, lcmByEnds, nil, nil},
		{"two bars with D = 6B > K", oneBar, 1, 2, lcmByEnds, nil, nil},
		{"pickup and one bar", oneBar, 0, 1, lcmAccepted, nil, nil},
		{"run across a mark and an explicit length, D = K", acrossK, 1, 6, lcmAccepted, k, nil},
		{"run across a mark and an explicit length, D = K + 2", acrossP9, 1, 9, lcmByItem2, new(big.Int).Add(k, big.NewInt(2)), nil},
		{"shift of an explicit length at the end out of range", shiftEnd, 1, 1, lcmByShift, nil, nil},
		{"shift out of range in a longer range", shiftEnd, 1, 2, lcmByShift, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g, model := lcmBuild(t, tt.spec, false)

			require.Equal(t, tt.reason, model.refuses(tt.first, tt.last), "the model of the criterion")

			run := model.runs(tt.first, tt.last)[0]
			if tt.d != nil {
				require.Equal(t, 0, tt.d.Cmp(run.d), "D of the first run is %s", run.d)
			}

			if tt.top != nil {
				require.Equal(t, 0, tt.top.Cmp(run.top()), "(F − L)·D of the first run is %s", run.top())
			}

			assertBarsVerdict(t, g, tt.first, tt.last, tt.reason != lcmAccepted)
		})
	}
}

// TestBarsCheckLCMOrder checks that the refusals of the arguments precede
// Е-10 on grids whose ranges Bars refuses by item 2: Е-1, Е-2 for first < 0,
// Е-2 for first > last, Е-4 for first, then for last.
func TestBarsCheckLCMOrder(t *testing.T) {
	t.Parallel()

	withPickup, _ := lcmBuild(t, lcmSpec{first: lcmSimple(1, 9), pickup: lcmInv(lcmP9()), bars: 9}, false)
	noPickup, _ := lcmBuild(t, lcmSpec{first: lcmSimple(4, 4), lengths: lcmRun(1, 16, ratFrac(lcmLambda(), big.NewInt(16))), bars: 16}, false)

	assertBarsFail(t, withPickup, 1, 9)
	assertBarsFail(t, noPickup, 1, 16)

	tests := []struct {
		name        string
		g           measure.Grid
		first, last int
		want        error
	}{
		{"unset grid", measure.Grid{}, 1, 9, measure.ErrUnset},
		{"first < 0", withPickup, -1, 9, measure.ErrInvalidArgument},
		{"first < 0 and first > last", withPickup, -1, -2, measure.ErrInvalidArgument},
		{"first > last", withPickup, 9, 1, measure.ErrInvalidArgument},
		{"first > last beyond the grid", withPickup, 12, 10, measure.ErrInvalidArgument},
		{"first beyond the grid", withPickup, 10, 12, measure.ErrOutsideGrid},
		{"last beyond the grid", withPickup, 1, 10, measure.ErrOutsideGrid},
		{"last at math.MaxInt beyond the grid", noPickup, 1, math.MaxInt, measure.ErrOutsideGrid},
		{"bar 0 without a pickup", noPickup, 0, 16, measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			seq, err := tt.g.Bars(tt.first, tt.last)
			assertErr(t, err, tt.want)
			assertNoCause(t, err, fraction.ErrOutOfRange)
			assert.Nil(t, seq)

			panicErr := panicError(t, func() { tt.g.MustBars(tt.first, tt.last) })
			if assert.Error(t, panicErr) && err != nil {
				assert.Equal(t, err.Error(), panicErr.Error())
				assert.ErrorIs(t, panicErr, tt.want)
			}
		})
	}
}

// Pools of the generated grids: meters and lengths near the bounds of the
// criterion together with ordinary values.

func lcmMeters() []lcmMeter {
	return []lcmMeter{
		lcmSimple(4, 4), lcmSimple(3, 4), lcmSimple(6, 8), lcmSimple(1, 7), lcmSimple(2, 14),
		lcmSimple(1, 9), lcmSimple(2, 18), lcmSimple(1, 1<<62), lcmSimple(33, 1), lcmSimple(5, 1<<63-25),
		{meter.MustNewSum(meter.MustNew(1, 1<<63-25), meter.MustNew(1, 1<<63-165)), ratAdd(big.NewRat(1, 1<<63-25), big.NewRat(1, 1<<63-165))},
	}
}

func lcmLengths() []*big.Rat {
	k := lcmK()

	return []*big.Rat{
		lcmInv(k), ratFrac(k, pow2(250)), ratFrac(lcmLambda(), big.NewInt(16)), lcmInv(lcmP9()),
		lcmInv(new(big.Int).Mul(big.NewInt(3), pow2(200))), lcmInv(pow2(254)), lcmInv(pow2(200)),
		lcmInv(new(big.Int).Add(pow2(250), big.NewInt(1))), ratFrac(new(big.Int).Add(pow2(254), big.NewInt(1)), big.NewInt(2)),
		big.NewRat(1, 6), lcmInv(new(big.Int).Lsh(lcmB(), 1)), lcmInv(new(big.Int).Add(pow2(254), big.NewInt(1))),
		big.NewRat(1, 4), big.NewRat(1, 2), big.NewRat(3, 8), big.NewRat(1, 3), big.NewRat(1, 1), big.NewRat(1, 7), big.NewRat(1, 9),
		big.NewRat(1, pbD),
	}
}

// lcmRandomSpec is a record with a fixed horizon of 12 bars.
func lcmRandomSpec(rng *rand.Rand) lcmSpec {
	const horizon = 12

	meters, lengths := lcmMeters(), lcmLengths()
	spec := lcmSpec{first: meters[rng.IntN(len(meters))]}

	if rng.IntN(2) == 0 {
		spec.pickup = lengths[rng.IntN(len(lengths))]
	}

	for range rng.IntN(4) {
		spec.marks = append(spec.marks, lcmMark{2 + rng.IntN(horizon-1), meters[rng.IntN(len(meters))]})
	}

	for range rng.IntN(5) {
		from := 1 + rng.IntN(horizon)
		to := min(horizon, from+rng.IntN(4))

		var length *big.Rat

		switch rng.IntN(3) {
		case 0:
			length = meters[rng.IntN(len(meters))].length
		default:
			length = lengths[rng.IntN(len(lengths))]
		}

		end := rng.IntN(4) == 0
		for bar := from; bar <= to; bar++ {
			spec.lengths = append(spec.lengths, lcmLength{bar: bar, length: length, end: end})
		}
	}

	if rng.IntN(2) == 0 {
		spec.bars = horizon + 1 + rng.IntN(6)
	}

	return spec
}

// TestBarsCheckLCMOracle compares Bars with the model of the criterion on
// generated grids with a fixed seed: ranges inside the record and ranges up
// to math.MaxInt of open grids, and every range of one bar, for which the
// criterion is exact: Bars(k, k) fails if and only if Bar(k) fails.
func TestBarsCheckLCMOracle(t *testing.T) {
	t.Parallel()

	const seed = 2608

	rng := rand.New(rand.NewPCG(seed, seed+1))

	grids := 400
	if testing.Short() {
		grids = 40
	}

	counts := map[lcmReason]int{}

	for i := range grids {
		spec := lcmRandomSpec(rng)
		g, model := lcmBuild(t, spec, true)

		lo := 1
		if model.pickup != nil {
			lo = 0
		}

		hi := math.MaxInt
		if spec.bars > 0 {
			hi = spec.bars
		}

		for j := range 6 {
			first := lo + rng.IntN(min(hi, model.horizon+2)-lo+1)
			last := min(hi, first+rng.IntN(model.horizon+2))

			if spec.bars == 0 && rng.IntN(3) == 0 {
				last = first + rng.IntN(1<<40)
				if rng.IntN(2) == 0 {
					last = math.MaxInt
				}
			}

			reason := model.refuses(first, last)
			counts[reason]++

			// The ranges are drawn outside the subtest, so -run selects one
			// range with the same values.
			t.Run(fmt.Sprintf("grid=%d,range=%d", i, j), func(t *testing.T) {
				defer guard(t)
				defer func() {
					if t.Failed() {
						t.Logf("grid %d (seed %d), Bars(%d, %d), model: %d", i, seed, first, last, reason)
					}
				}()

				assertBarsVerdict(t, g, first, last, reason != lcmAccepted)
			})
		}

		for bar := lo; bar <= min(hi, model.horizon+1); bar++ {
			refused := model.refuses(bar, bar) != lcmAccepted

			_, barErr := g.Bar(bar)
			if refused {
				assertErr(t, barErr, measure.ErrOutOfRange, fraction.ErrOutOfRange)
			} else {
				assert.NoError(t, barErr, "grid %d (seed %d), Bar(%d)", i, seed, bar)
			}

			seq, err := g.Bars(bar, bar)
			assert.Equal(t, refused, err != nil, "grid %d (seed %d), Bars(%d, %d): %v", i, seed, bar, bar, err)
			assert.Equal(t, refused, seq == nil)
		}
	}

	t.Logf("verdicts of the model: %v", counts)

	if !testing.Short() {
		for _, reason := range []lcmReason{lcmAccepted, lcmByShift, lcmByEnds, lcmByItem2} {
			assert.Positive(t, counts[reason], "the generated ranges include the verdict %d", reason)
		}
	}
}
