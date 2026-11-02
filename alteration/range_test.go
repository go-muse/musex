package alteration_test

import (
	"errors"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-14, K-22: the range constants have the values of the contract and are
// untyped: they are usable in arithmetic of different integer types.
func TestRangeConstants(t *testing.T) {
	t.Parallel()

	magnitudeInt := alteration.MaxMagnitude
	denominatorInt := alteration.MaxDenominator

	var (
		magnitudeInt64  int64  = alteration.MaxMagnitude
		magnitudeUint32 uint32 = alteration.MaxMagnitude
		denominatorI64  int64  = alteration.MaxDenominator
		denominatorU32  uint32 = alteration.MaxDenominator
	)

	assert.Equal(t, 1<<24, magnitudeInt)
	assert.Equal(t, int64(1<<24), magnitudeInt64)
	assert.Equal(t, uint32(1<<24), magnitudeUint32)
	assert.Equal(t, 1<<16, denominatorInt)
	assert.Equal(t, uint32(1<<16), denominatorU32)
	assert.Equal(t, int64(1<<16), denominatorI64)
	assert.Equal(t, oracleMaxMagnitude, magnitudeInt, "oracle constant")
	assert.Equal(t, oracleMaxDenominator, denominatorInt, "oracle constant")
}

// K-13, K-14: the two sentinel errors exist, differ and do not wrap each
// other.
func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	require.Error(t, alteration.ErrZeroDenominator)
	require.Error(t, alteration.ErrOutOfRange)
	assert.NotErrorIs(t, alteration.ErrZeroDenominator, alteration.ErrOutOfRange)
	assert.NotErrorIs(t, alteration.ErrOutOfRange, alteration.ErrZeroDenominator)
	assert.False(t, errors.Is(alteration.ErrZeroDenominator, alteration.ErrOutOfRange))
}

// K-13: a zero denominator is a distinguishable error for any numerator;
// the zero value is returned; the Must variant panics with that error.
func TestFromFraction_ZeroDenominator(t *testing.T) {
	t.Parallel()

	ops := make([]fallibleOp, 0, 5)
	for _, num := range []int{1, 0, -5, math.MinInt, math.MaxInt} {
		ops = append(ops, opFromFraction(num, 0))
	}

	runFallible(t, ops)
}

// rangeCase is a boundary case with the expected side of the range stated
// explicitly; the test also checks that the math/big oracle agrees with it.
type rangeCase struct {
	op      fallibleOp
	inRange bool
}

func inside(op fallibleOp) rangeCase {
	return rangeCase{op: op, inRange: true}
}

func outside(op fallibleOp) rangeCase {
	return rangeCase{op: op, inRange: false}
}

// runRange runs every case once as a named parallel subtest. A case that
// repeats an earlier name runs once; its stated side of the range must match
// the earlier one, so a contradicting duplicate does not disappear silently.
func runRange(t *testing.T, cases []rangeCase) {
	t.Helper()

	seen := make(map[string]bool, len(cases))
	for _, tc := range cases {
		if side, ok := seen[tc.op.name]; ok {
			require.Equal(t, side, tc.inRange, "repeated case %s states the same side of the range", tc.op.name)

			continue
		}
		seen[tc.op.name] = tc.inRange

		t.Run(tc.op.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.op.want, "a boundary case has a nonzero denominator")
			require.Equal(t, tc.inRange, inRange(tc.op.want), "the case table agrees with the oracle: %s", tc.op.want.RatString())
			checkFallible(t, tc.op)
		})
	}
}

// K-14 with M = MaxMagnitude, m = -M and Q = MaxDenominator (architecture,
// "Guaranteed range" and "Boundary cases"). Errors are ErrOutOfRange and not
// ErrZeroDenominator, with the zero value; Must variants panic with the same
// error; inputs are unchanged (K-17). The points follow the criterion.
func TestRange(t *testing.T) {
	t.Parallel()

	var (
		top        = edgeTop()
		bottom     = edgeBottom()
		natural    = frac{0, 1}
		half       = frac{1, 2}
		sharp      = frac{1, 1}
		topHalf    = frac{2*bigM - 1, 2} // M - 1/2
		bottomHalf = frac{1 - 2*bigM, 2} // m + 1/2
	)

	t.Run("1 edges are created exactly", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			inside(opFromInt(bigM)),
			inside(opFromInt(-bigM)),
			inside(opFromFraction(bigM, 1)),
			inside(opFromFraction(-bigM, 1)),
			inside(opFromFraction(1, bigQ)),
			inside(opFromFraction(-1, bigQ)),
			inside(opFromFraction(bigM*bigQ-1, bigQ)),
			inside(opFromFraction(1-bigM*bigQ, bigQ)),
			inside(opFromFraction(bigM*bigQ, bigQ)),
			inside(opFromFraction(-bigM*bigQ, bigQ)),
			inside(opFromFraction(bigQ-1, bigQ)),
			inside(opFromFraction(2*bigM-1, 2)),
		})
	})

	t.Run("2 creation out of range", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			outside(opFromInt(bigM + 1)),
			outside(opFromInt(-bigM - 1)),
			outside(opFromInt(math.MaxInt)),
			outside(opFromInt(math.MinInt)),
			outside(opFromFraction(bigM+1, 1)),
			outside(opFromFraction(-bigM-1, 1)),
			outside(opFromFraction(2*bigM+1, 2)),
			outside(opFromFraction(bigM*bigQ+1, bigQ)),
			outside(opFromFraction(-bigM*bigQ-1, bigQ)),
			outside(opFromFraction(math.MaxInt, 1)),
			outside(opFromFraction(math.MinInt, 1)),
			outside(opFromFraction(math.MinInt, 2)),
			outside(opFromFraction(math.MinInt, -1)),
			outside(opFromFraction(math.MaxInt, -1)),
			outside(opFromFraction(1, bigQ+1)),
			outside(opFromFraction(1, -(bigQ + 1))),
			outside(opFromFraction(-1, bigQ+1)),
			outside(opFromFraction(3, 2*bigQ)),
			outside(opFromFraction(1, math.MinInt)),
			outside(opFromFraction(1, math.MaxInt)),
			outside(opFromFraction(-1, math.MaxInt)),
			outside(opFromFraction(math.MaxInt, math.MaxInt-1)),
			outside(opFromFraction(math.MinInt, 1<<38)),
		})
	})

	t.Run("3 by one", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			outside(opRaise(top)),
			outside(opLower(bottom)),
			outside(opRaise(edgeNearTop())),
			outside(opLower(edgeNearBottom())),
			outside(opRaise(topHalf)),
			outside(opLower(bottomHalf)),
			inside(opRaise(frac{bigM - 1, 1})),
			inside(opLower(frac{1 - bigM, 1})),
			inside(opLower(top)),
			inside(opRaise(bottom)),
		})
	})

	t.Run("4 by an integer", func(t *testing.T) {
		t.Parallel()

		var cases []rangeCase
		for _, n := range []int{1, 2, math.MaxInt} {
			cases = append(cases,
				outside(opRaiseBy(top, n)),
				outside(opLowerBy(bottom, n)),
				outside(opRaiseBy(bottom, -n)),
				outside(opLowerBy(top, -n)),
			)
		}

		for _, x := range []frac{natural, half, top, bottom} {
			cases = append(cases,
				outside(opRaiseBy(x, math.MaxInt)),
				outside(opLowerBy(x, math.MaxInt)),
				outside(opRaiseBy(x, math.MinInt)),
				outside(opLowerBy(x, math.MinInt)),
			)
		}

		cases = append(cases,
			outside(opRaiseBy(natural, bigM+1)),
			outside(opLowerBy(natural, bigM+1)),
			outside(opRaiseBy(bottom, 2*bigM+1)),
			outside(opLowerBy(top, 2*bigM+1)),
			outside(opRaiseBy(edgeNearTop(), 1)),
			outside(opLowerBy(edgeNearBottom(), 1)),
			inside(opRaiseBy(natural, bigM)),
			inside(opLowerBy(natural, bigM)),
			inside(opRaiseBy(natural, -bigM)),
			inside(opLowerBy(natural, -bigM)),
		)

		runRange(t, cases)
	})

	t.Run("5 by a value", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			outside(opAdd(top, sharp)),
			outside(opSub(bottom, sharp)),
			outside(opAdd(top, half)),
			outside(opSub(bottom, half)),
			outside(opAdd(top, edgeInvQ())),
			outside(opSub(bottom, edgeInvQ())),
			outside(opAdd(topHalf, sharp)),
			outside(opSub(bottomHalf, sharp)),
			outside(opAdd(edgeNearTop(), frac{2, bigQ})),
			outside(opAdd(frac{bigM/2 + 1, 1}, frac{bigM / 2, 1})),
			outside(opSub(frac{-bigM/2 - 1, 1}, frac{bigM / 2, 1})),
			inside(opAdd(edgeNearTop(), edgeInvQ())),
			inside(opSub(edgeNearBottom(), edgeInvQ())),
			inside(opAdd(topHalf, half)),
			inside(opSub(bottomHalf, half)),
		})
	})

	t.Run("6 product", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			outside(opMul(top, 2)),
			outside(opMul(bottom, 2)),
			outside(opMul(top, -2)),
			outside(opMul(bottom, -2)),
			outside(opMul(edgeNearTop(), 2)),
			outside(opMul(edgeInvQ(), math.MaxInt)),
			outside(opMul(edgeInvQ(), math.MinInt)),
			outside(opMul(edgeInvQ(), bigM*bigQ+1)),
			outside(opMul(sharp, math.MinInt)),
			outside(opMul(sharp, math.MaxInt)),
			outside(opMul(sharp, bigM+1)),
			outside(opMul(half, math.MinInt)),
			outside(opMul(half, math.MaxInt)),
			outside(opMul(half, 2*bigM+1)),
			outside(opMul(top, math.MinInt)),
			outside(opMul(bottom, math.MaxInt)),
		})
	})

	t.Run("7 difference and denominators", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			outside(opSub(top, bottom)),
			outside(opSub(bottom, top)),
			outside(opAdd(top, top)),
			outside(opAdd(bottom, bottom)),
			outside(opAdd(edgeInvQ(), frac{1, bigQ - 1})),
			outside(opSub(edgeInvQ(), frac{1, bigQ - 1})),
			outside(opAdd(edgeInvQ(), frac{1, 3})),
			outside(opAdd(frac{1, 257}, frac{1, 263})),
			outside(opSub(frac{1, 257}, frac{1, 263})),
		})
	})

	t.Run("8 exact results in range", func(t *testing.T) {
		t.Parallel()

		runRange(t, []rangeCase{
			inside(opAdd(edgeInvQ(), frac{bigQ - 1, bigQ})),
			inside(opAdd(frac{1, bigQ - 1}, frac{bigQ - 2, bigQ - 1})),
			inside(opAdd(frac{1, 255}, frac{1, 257})),
			inside(opSub(edgeInvQ(), edgeInvQ())),
			inside(opSub(edgeInvQ(), edgeNegInvQ())),
			inside(opAdd(top, bottom)),
			inside(opSub(top, top)),
			inside(opMul(natural, math.MaxInt)),
			inside(opMul(natural, math.MinInt)),
			inside(opMul(edgeInvQ(), bigM*bigQ)),
			inside(opMul(edgeInvQ(), -bigM*bigQ)),
			inside(opMul(edgeInvQ(), bigQ)),
			inside(opMul(top, -1)),
			inside(opMul(bottom, -1)),
			inside(opMul(top, 0)),
			inside(opMul(bottom, 0)),
			inside(opMul(top, 1)),
			inside(opRaiseBy(bottom, 2*bigM)),
			inside(opLowerBy(top, 2*bigM)),
			inside(opFromFraction(2*(bigQ+1), bigQ+1)),
			inside(opFromFraction(math.MinInt, math.MinInt)),
			inside(opFromFraction(math.MaxInt, math.MaxInt)),
			inside(opFromFraction(math.MinInt, -(1 << 62))),
			inside(opFromFraction(-(1 << 62), math.MinInt)),
			inside(opFromFraction(0, math.MinInt)),
			inside(opFromFraction(0, math.MaxInt)),
			inside(opFromFraction(math.MinInt, 1<<39)),
			inside(opFromFraction(1<<62, 1<<38)),
			inside(opFromFraction(math.MinInt, 1<<40)),
		})
	})

	t.Run("8 named results", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			got  func() (alteration.Alteration, error)
			want alteration.Alteration
		}{
			{
				name: "MinInt over -(1<<62) is a double sharp, not a double flat",
				got:  func() (alteration.Alteration, error) { return alteration.FromFraction(math.MinInt, -(1 << 62)) },
				want: alteration.DoubleSharp(),
			},
			{
				name: "MinInt over MinInt is a sharp",
				got:  func() (alteration.Alteration, error) { return alteration.FromFraction(math.MinInt, math.MinInt) },
				want: alteration.Sharp(),
			},
			{
				name: "2(Q+1) over Q+1 is a double sharp",
				got:  func() (alteration.Alteration, error) { return alteration.FromFraction(2*(bigQ+1), bigQ+1) },
				want: alteration.DoubleSharp(),
			},
			{
				name: "natural times MaxInt is natural",
				got:  func() (alteration.Alteration, error) { return alteration.Natural().Mul(math.MaxInt) },
				want: alteration.Natural(),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.got()
				require.NoError(t, err)
				assertSame(t, tc.want, got, tc.name)
			})
		}

		positiveHalf, err := alteration.FromFraction(-(1 << 62), math.MinInt)
		require.NoError(t, err)
		assert.Equal(t, frac{1, 2}, fractionOf(positiveHalf), "-(1<<62) over MinInt is +1/2, not -1/2")
	})
}

// K-4, K-5, K-6: at the edges, Raise agrees with RaiseBy(1) and Add(Sharp()),
// Lower with LowerBy(1) and Sub(Sharp()), LowerBy(n) with RaiseBy(-n) —
// including the errors.
func TestEquivalentFormsAtEdges(t *testing.T) {
	t.Parallel()

	edges := []frac{edgeTop(), edgeBottom(), edgeNearTop(), edgeNearBottom(), edgeInvQ(), {0, 1}, {1, 2}}

	for _, f := range edges {
		t.Run(f.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, f)

			raise := call(a.Raise)
			assertSameOutcome(t, raise, call(func() (alteration.Alteration, error) { return a.RaiseBy(1) }), "RaiseBy(1)")
			assertSameOutcome(t, raise, call(func() (alteration.Alteration, error) { return a.Add(alteration.Sharp()) }), "Add(Sharp())")
			assertSameOutcome(t, raise, callMust(a.MustRaise), "MustRaise")

			lower := call(a.Lower)
			assertSameOutcome(t, lower, call(func() (alteration.Alteration, error) { return a.LowerBy(1) }), "LowerBy(1)")
			assertSameOutcome(t, lower, call(func() (alteration.Alteration, error) { return a.Sub(alteration.Sharp()) }), "Sub(Sharp())")
			assertSameOutcome(t, lower, callMust(a.MustLower), "MustLower")

			for _, n := range []int{0, 1, -1, 2, bigM, -bigM, 2 * bigM, -2 * bigM, math.MaxInt, -math.MaxInt} {
				assertSameOutcome(t,
					call(func() (alteration.Alteration, error) { return a.RaiseBy(-n) }),
					call(func() (alteration.Alteration, error) { return a.LowerBy(n) }),
					"LowerBy(%d) = RaiseBy(%d)", n, -n)
			}
		})
	}
}
