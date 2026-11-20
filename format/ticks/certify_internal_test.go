package ticks

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file cover the internal seam of architecture 5.3:
// roundCertified with fake evaluators. Expectations follow only from the
// described behavior of the seam and the definitions of the rounding modes.

var certifyModes = []struct {
	name string
	mode rounding.Mode
	// tie2 is the expected result for exactly 5/2, tie3 for exactly 7/2.
	tie2, tie3 int64
	// inner is the expected result for a value in (2, 5/2); upper for a value in (5/2, 3).
	inner, upper int64
}{
	{name: "down", mode: rounding.Down, tie2: 2, tie3: 3, inner: 2, upper: 2},
	{name: "up", mode: rounding.Up, tie2: 3, tie3: 4, inner: 3, upper: 3},
	{name: "toward zero", mode: rounding.TowardZero, tie2: 2, tie3: 3, inner: 2, upper: 2},
	{name: "away from zero", mode: rounding.AwayFromZero, tie2: 3, tie3: 4, inner: 3, upper: 3},
	{name: "half up", mode: rounding.HalfUp, tie2: 3, tie3: 4, inner: 2, upper: 3},
	{name: "half down", mode: rounding.HalfDown, tie2: 2, tie3: 3, inner: 2, upper: 3},
	{name: "half away from zero", mode: rounding.HalfAwayFromZero, tie2: 3, tie3: 4, inner: 2, upper: 3},
	{name: "half toward zero", mode: rounding.HalfTowardZero, tie2: 2, tie3: 3, inner: 2, upper: 3},
	{name: "half even", mode: rounding.HalfEven, tie2: 2, tie3: 4, inner: 2, upper: 3},
}

type evalStep struct {
	lo, hi *big.Rat
	exact  bool
	err    error
}

// recordingEval returns an evaluator answering by the number of decimal
// places from steps, and the list of places it was called with.
func recordingEval(t *testing.T, steps map[int]evalStep) (evalFunc, *[]int) {
	t.Helper()
	calls := &[]int{}
	return func(places int) (*big.Rat, *big.Rat, bool, error) {
		*calls = append(*calls, places)
		step, ok := steps[places]
		if !ok {
			t.Errorf("unexpected call with %d places", places)
			return nil, nil, false, errors.New("unexpected call")
		}
		return step.lo, step.hi, step.exact, step.err
	}, calls
}

func rat(num, den int64) *big.Rat {
	return big.NewRat(num, den)
}

func sameSteps(step evalStep) map[int]evalStep {
	return map[int]evalStep{20: step, 40: step, 60: step, 76: step}
}

func requireCertified(t *testing.T, got *big.Int, err error, want int64) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 0, got.Cmp(big.NewInt(want)), "got %s, want %d", got, want)
}

func TestPlacesSchedule(t *testing.T) {
	t.Parallel()
	assert.Equal(t, [...]int{20, 40, 60, 76}, placesSchedule)
}

// TestRoundCertifiedExactTies checks step 2: an exact value is rounded by
// mode.RatInt with ties by the rule of the mode; one call is made.
func TestRoundCertifiedExactTies(t *testing.T) {
	t.Parallel()
	for _, mode := range certifyModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			for _, tt := range []struct {
				num  int64
				want int64
			}{
				{num: 5, want: mode.tie2},
				{num: 7, want: mode.tie3},
			} {
				eval, calls := recordingEval(t, map[int]evalStep{20: {lo: rat(tt.num, 2), hi: rat(tt.num, 2), exact: true}})
				got, err := roundCertified(eval, mode.mode, big.NewInt(1000))
				requireCertified(t, got, err, tt.want)
				assert.Equal(t, []int{20}, *calls)
			}
			// An exact integer.
			eval, calls := recordingEval(t, map[int]evalStep{20: {lo: rat(4, 1), hi: rat(4, 1), exact: true}})
			got, err := roundCertified(eval, mode.mode, big.NewInt(4))
			requireCertified(t, got, err, 4)
			assert.Equal(t, []int{20}, *calls)
		})
	}
}

// TestRoundCertifiedDecidedAtFirstPrecision checks step 3 with exactly one call.
func TestRoundCertifiedDecidedAtFirstPrecision(t *testing.T) {
	t.Parallel()
	for _, mode := range certifyModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			eval, calls := recordingEval(t, map[int]evalStep{20: {lo: rat(21, 10), hi: rat(22, 10)}})
			got, err := roundCertified(eval, mode.mode, big.NewInt(1000))
			requireCertified(t, got, err, mode.inner)
			assert.Equal(t, []int{20}, *calls)
		})
	}
}

// TestRoundCertifiedDecidedAtThirdPrecision checks steps 1 and 3: places are
// requested in order 20, 40, 60 and the loop stops at the first decision.
func TestRoundCertifiedDecidedAtThirdPrecision(t *testing.T) {
	t.Parallel()
	wide := evalStep{lo: rat(19, 10), hi: rat(26, 10)}
	for _, mode := range certifyModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			eval, calls := recordingEval(t, map[int]evalStep{
				20: wide,
				40: wide,
				60: {lo: rat(251, 100), hi: rat(252, 100)},
			})
			got, err := roundCertified(eval, mode.mode, big.NewInt(1000))
			requireCertified(t, got, err, mode.upper)
			assert.Equal(t, []int{20, 40, 60}, *calls)
		})
	}
}

// TestRoundCertifiedUndecided checks step 7: no decision after 76 places is
// ErrOutOfRange with fraction.ErrOutOfRange.
func TestRoundCertifiedUndecided(t *testing.T) {
	t.Parallel()
	for _, mode := range certifyModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			eval, calls := recordingEval(t, sameSteps(evalStep{lo: rat(19, 10), hi: rat(26, 10)}))
			got, err := roundCertified(eval, mode.mode, big.NewInt(1000))
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, ErrOutOfRange)
			assert.ErrorIs(t, err, fraction.ErrOutOfRange)
			assert.NotErrorIs(t, err, ErrFormatLimit)
			assert.Equal(t, []int{20, 40, 60, 76}, *calls)
		})
	}
}

// TestRoundCertifiedUnboundedAbove checks hi == nil: the precision does not
// decide and a later precision decides.
func TestRoundCertifiedUnboundedAbove(t *testing.T) {
	t.Parallel()
	for _, mode := range certifyModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			eval, calls := recordingEval(t, map[int]evalStep{
				20: {lo: rat(21, 10)},
				40: {lo: rat(21, 10)},
				60: {lo: rat(21, 10), hi: rat(22, 10)},
			})
			got, err := roundCertified(eval, mode.mode, big.NewInt(1000))
			requireCertified(t, got, err, mode.inner)
			assert.Equal(t, []int{20, 40, 60}, *calls)

			eval, calls = recordingEval(t, sameSteps(evalStep{lo: rat(21, 10)}))
			got, err = roundCertified(eval, mode.mode, big.NewInt(1000))
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, ErrOutOfRange)
			assert.ErrorIs(t, err, fraction.ErrOutOfRange)
			assert.Equal(t, []int{20, 40, 60, 76}, *calls)
		})
	}
}

// TestRoundCertifiedAboveLimit checks steps 4 and 5.
func TestRoundCertifiedAboveLimit(t *testing.T) {
	t.Parallel()
	for _, mode := range certifyModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			// The lower end above the limit decides without further calls.
			eval, calls := recordingEval(t, sameSteps(evalStep{lo: rat(100, 1)}))
			got, err := roundCertified(eval, mode.mode, big.NewInt(10))
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, ErrFormatLimit)
			assert.NotErrorIs(t, err, ErrOutOfRange)
			assert.Equal(t, []int{20}, *calls)

			eval, calls = recordingEval(t, sameSteps(evalStep{lo: rat(99, 2), hi: rat(203, 2)}))
			_, err = roundCertified(eval, mode.mode, big.NewInt(10))
			assert.ErrorIs(t, err, ErrFormatLimit)
			assert.Equal(t, []int{20}, *calls)

			// A decided result above the limit.
			eval, calls = recordingEval(t, map[int]evalStep{20: {lo: rat(11, 1), hi: rat(11, 1), exact: true}})
			got, err = roundCertified(eval, mode.mode, big.NewInt(10))
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, ErrFormatLimit)
			assert.Equal(t, []int{20}, *calls)

			eval, _ = recordingEval(t, map[int]evalStep{20: {lo: rat(111, 10), hi: rat(112, 10)}})
			_, err = roundCertified(eval, mode.mode, big.NewInt(10))
			assert.ErrorIs(t, err, ErrFormatLimit)

			// The limit itself is allowed.
			eval, _ = recordingEval(t, map[int]evalStep{20: {lo: rat(10, 1), hi: rat(10, 1), exact: true}})
			got, err = roundCertified(eval, mode.mode, big.NewInt(10))
			requireCertified(t, got, err, 10)

			// The lower end below the limit with an undecided interval goes on.
			eval, calls = recordingEval(t, map[int]evalStep{
				20: {lo: rat(19, 2), hi: rat(23, 2)},
				40: {lo: rat(101, 10), hi: rat(102, 10)},
			})
			got, err = roundCertified(eval, mode.mode, big.NewInt(20))
			want := int64(10)
			if mode.inner == 3 {
				want = 11
			}
			requireCertified(t, got, err, want)
			assert.Equal(t, []int{20, 40}, *calls)
		})
	}
}

// TestRoundCertifiedEvalErrors checks step 6.
func TestRoundCertifiedEvalErrors(t *testing.T) {
	t.Parallel()
	wide := evalStep{lo: rat(19, 10), hi: rat(26, 10)}
	rangeErr := fmt.Errorf("evaluation: %w", tempomap.ErrOutOfRange)
	eval, calls := recordingEval(t, map[int]evalStep{20: wide, 40: {err: rangeErr}})
	got, err := roundCertified(eval, rounding.HalfEven, big.NewInt(1000))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, ErrOutOfRange)
	assert.ErrorIs(t, err, tempomap.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrOutOfRange)
	assert.NotErrorIs(t, err, ErrFormatLimit)
	assert.Equal(t, []int{20, 40}, *calls)

	other := errors.New("other failure")
	eval, calls = recordingEval(t, map[int]evalStep{20: {err: other}})
	got, err = roundCertified(eval, rounding.HalfEven, big.NewInt(1000))
	assert.Nil(t, got)
	assert.True(t, err == other, "got %v", err)
	assert.Equal(t, []int{20}, *calls)

	eval, calls = recordingEval(t, map[int]evalStep{20: wide, 40: wide, 60: {err: other}})
	_, err = roundCertified(eval, rounding.Down, big.NewInt(1000))
	assert.True(t, err == other, "got %v", err)
	assert.Equal(t, []int{20, 40, 60}, *calls)
}

// TestRoundCertifiedDoesNotChangeOrKeepValues checks step 8.
func TestRoundCertifiedDoesNotChangeOrKeepValues(t *testing.T) {
	t.Parallel()
	lo := rat(21, 10)
	hi := rat(22, 10)
	limit := big.NewInt(1000)
	eval, _ := recordingEval(t, map[int]evalStep{20: {lo: lo, hi: hi}})
	got, err := roundCertified(eval, rounding.Up, limit)
	requireCertified(t, got, err, 3)
	assert.Equal(t, "21/10", lo.RatString())
	assert.Equal(t, "11/5", hi.RatString())
	assert.Equal(t, "1000", limit.String())
	got.SetInt64(99)
	assert.Equal(t, "21/10", lo.RatString())
	assert.Equal(t, "1000", limit.String())

	exact := rat(4, 1)
	eval, _ = recordingEval(t, map[int]evalStep{20: {lo: exact, hi: exact, exact: true}})
	got, err = roundCertified(eval, rounding.Down, limit)
	requireCertified(t, got, err, 4)
	got.SetInt64(7)
	assert.Equal(t, "4", exact.RatString())
}
