package tempomap_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNineRoundingModesAtZeroPlaces(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))
	for _, tt := range []struct {
		mode  rounding.Mode
		value string
		bound string
	}{
		{rounding.Down, "2/1", "1/1"},
		{rounding.TowardZero, "2/1", "1/1"},
		{rounding.Up, "3/1", "1/1"},
		{rounding.AwayFromZero, "3/1", "1/1"},
		{rounding.HalfUp, "3/1", "1/2"},
		{rounding.HalfDown, "3/1", "1/2"},
		{rounding.HalfAwayFromZero, "3/1", "1/2"},
		{rounding.HalfTowardZero, "3/1", "1/2"},
		{rounding.HalfEven, "3/1", "1/2"},
	} {
		checkSecondsResult(t, m.MustStart(at(1, 1), 0, tt.mode), tt.value, true, 0, tt.bound)
	}
}

func TestNearestModesImmediatelyAroundMidpoints(t *testing.T) {
	t.Parallel()
	// Independently evaluated full-segment times:
	// BPM 335.32 (speed 83.830) -> 1.49999192580057... (below),
	// BPM 335.316 (speed 83.829) -> 1.50000332003022... (above),
	// BPM 144.132 (speed 36.033) -> 2.50002848135288... (above),
	// BPM 144.136 (speed 36.034) -> 2.49998878793752... (below).
	for _, tt := range []struct {
		name       string
		finalBPM   string
		halfUp     string
		halfDown   string
		halfEven   string
		halfAway   string
		halfToward string
		down       string
		up         string
	}{
		{"below odd midpoint", "83830/250", "1/1", "1/1", "1/1", "1/1", "1/1", "1/1", "2/1"},
		{"above odd midpoint", "83829/250", "2/1", "2/1", "2/1", "2/1", "2/1", "1/1", "2/1"},
		{"above even midpoint", "36033/250", "3/1", "3/1", "3/1", "3/1", "3/1", "2/1", "3/1"},
		{"below even midpoint", "36034/250", "2/1", "2/1", "2/1", "2/1", "2/1", "2/1", "3/1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newGradualMap(pos(1, 1), tempo.MustNewQuarter(fractionRat(tt.finalBPM)))
			for _, mode := range allRoundingModes {
				result := m.MustStart(at(1, 1), 0, mode)
				approximate, err := result.Approximate()
				require.NoError(t, err)
				assert.True(t, approximate)
			}
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.HalfUp), tt.halfUp, true, 0, "1/2")
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.HalfDown), tt.halfDown, true, 0, "1/2")
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.HalfEven), tt.halfEven, true, 0, "1/2")
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.HalfAwayFromZero), tt.halfAway, true, 0, "1/2")
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.HalfTowardZero), tt.halfToward, true, 0, "1/2")
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.Down), tt.down, true, 0, "1/1")
			checkSecondsResult(t, m.MustStart(at(1, 1), 0, rounding.Up), tt.up, true, 0, "1/1")
		})
	}
}

func TestMultipleGradualSegmentsRoundOnlyTheWholeExpression(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewGradual(at(2, 1), tempomap.LinearSpeed)).
		MustWithChange(tempomap.MustNewMain(at(3, 1), quarter(240)))

	// Start(3) = 4 ln 2 + 2 + 4 ln 2 = 7.54517744447956....
	checkSecondsResult(t, m.MustStart(at(3, 1), 12, rounding.Down), "7545177444479/1000000000000", true, 12, "1/1000000000000")
	checkSecondsResult(t, m.MustStart(at(3, 1), 12, rounding.HalfUp), "754517744448/100000000000", true, 12, "1/2000000000000")
	checkSecondsResult(t, m.MustStart(at(3, 1), 0, rounding.Down), "7/1", true, 0, "1/1")

	segments := m.MustSegments()
	require.Len(t, segments, 4)
	individual := 0
	for _, segment := range segments[:3] {
		seconds, ok := segment.MustSeconds(0, rounding.Down)
		require.True(t, ok)
		value, err := seconds.Value()
		require.NoError(t, err)
		numerator, ok := value.MustFraction().Int()
		require.True(t, ok)
		individual += numerator
	}
	assert.Equal(t, 6, individual, "rounding each segment separately loses the exact fractional sum")
	checkSecondsResult(t, m.MustStart(at(3, 1), 0, rounding.Down), "7/1", true, 0, "1/1")

	// At t=5 the second exponent is x=(3-4 ln 2)/2. It contains the
	// preceding logarithm and must not be replaced by an independently
	// rounded prefix. Algebraically p=1+exp((3-4 ln 2)/2).
	checkPositionResult(t, m.MustPosition(sec(5, 1), 12, rounding.Down), "2120422267584/1000000000000", true, 12, "1/1000000000000")
	checkPositionResult(t, m.MustPosition(sec(5, 1), 12, rounding.HalfUp), "2120422267585/1000000000000", true, 12, "1/2000000000000")
}

func TestLogarithmFactorizationAndNearCancellation(t *testing.T) {
	t.Parallel()
	// 4 ln 4 and 8 ln 2 are the same value. The exact zero detector must
	// recognize ln 4 - 2 ln 2 without a floating-point comparison.
	ratioFour := newGradualMap(pos(1, 1), quarter(240))
	ratioTwoTwice := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewGradual(at(2, 1), tempomap.LinearSpeed)).
		MustWithChange(tempomap.MustNewMain(at(3, 1), quarter(240)))

	// Start(3) is 4 ln 4 + 2 for ratioFour and 4 ln 2 + 2 + 4 ln 2
	// for ratioTwoTwice. Both independently equal 7.54517744447956....
	first := ratioFour.MustStart(at(3, 1), 12, rounding.HalfUp)
	second := ratioTwoTwice.MustStart(at(3, 1), 12, rounding.HalfUp)
	checkSecondsResult(t, first, "754517744448/100000000000", true, 12, "1/2000000000000")
	checkSecondsResult(t, second, "754517744448/100000000000", true, 12, "1/2000000000000")

	// The inverse case with x=(3-4 ln 2)/2 also has a nonzero rational
	// term and an algebraic irrational exponential candidate. It exercises
	// the non-cancelling and almost-cancelling certificate paths.
	chained := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewGradual(at(2, 1), tempomap.LinearSpeed)).
		MustWithChange(tempomap.MustNewMain(at(3, 1), quarter(240)))
	result := chained.MustPosition(sec(5, 1), 12, rounding.HalfEven)
	checkPositionResult(t, result, "2120422267585/1000000000000", true, 12, "1/2000000000000")
}

func TestAccelerationAndSlowingBoundariesAndInverse(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name         string
		final        tempo.Tempo
		startSpeed   string
		endSpeed     string
		inverseValue string
	}{
		{"acceleration", quarter(120), "15/1", "30/1", "284025416688/1000000000000"},
		{"slowing", quarter(30), "15/1", "15/2", "235006194831/1000000000000"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newGradualMap(pos(1, 1), tt.final)
			assert.True(t, fractionRat(tt.startSpeed).Equal(m.MustSpeed(tempomap.MustAt(duration.Origin()))))
			assert.True(t, fractionRat(tt.endSpeed).Equal(m.MustSpeed(at(1, 1))))
			checkPositionResult(t, m.MustPosition(tempo.ZeroSeconds(), 12, rounding.HalfUp), "0/1", false, 0, "0/1")
			if tt.name == "slowing" {
				checkSecondsResult(t, m.MustStart(at(1, 1), 12, rounding.HalfUp), "554517744448/100000000000", true, 12, "1/2000000000000")
			} else {
				checkSecondsResult(t, m.MustStart(at(1, 1), 12, rounding.HalfUp), "277258872224/100000000000", true, 12, "1/2000000000000")
			}
			checkPositionResult(t, m.MustPosition(sec(1, 1), 12, rounding.HalfUp), tt.inverseValue, true, 12, "1/2000000000000")
		})
	}

	// The slowing map is the public case A=2, B=-2. Its inverse is
	// p=2-2exp(-t/8), so R=0 and small negative R occur when certifying
	// rounding comparisons. For y=4 and E_min=1/2 the independent lower
	// bound is |B|(E_min-R)=1, not 2; the observable results must remain
	// correctly rounded on both sides of the candidate grid.
	slowing := newGradualMap(pos(1, 1), quarter(30))
	checkPositionResult(t, slowing.MustPosition(sec(1, 1), 0, rounding.Down), "0/1", true, 0, "1/1")
	checkPositionResult(t, slowing.MustPosition(sec(1, 1), 0, rounding.Up), "1/1", true, 0, "1/1")
	checkPositionResult(t, slowing.MustPosition(sec(1, 1), 0, rounding.HalfUp), "0/1", true, 0, "1/2")
}

func TestKP44LongGradualSegment(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(2, 1), quarter(120))
	checkSecondsResult(t, m.MustStart(at(2, 1), 12, rounding.HalfUp), "554517744448/100000000000", true, 12, "1/2000000000000")

	// At 8 ln(3/2) the exact position is 1. The input type is rational, so
	// use its independently rounded decimal and verify the corresponding
	// mathematical result rather than substituting the ideal boundary.
	seconds := secondsRat("3243720864865/1000000000000")
	checkPositionResult(t, m.MustPosition(seconds, 12, rounding.HalfUp), "1/1", true, 12, "1/2000000000000")
	checkPositionResult(t, m.MustPosition(seconds, 12, rounding.Down), "999999999999/1000000000000", true, 12, "1/1000000000000")

	nearFull := secondsRat("554517744448/100000000000")
	checkPositionResult(t, m.MustPosition(nearFull, 12, rounding.HalfUp), "2/1", true, 12, "1/2000000000000")
}

func TestKP45ApproximationRangeBoundary(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))
	value := "27725887222397812376689284858327062723020005374410210164827200379735744878788/10000000000000000000000000000000000000000000000000000000000000000000000000000"
	checkSecondsResult(t, m.MustStart(at(1, 1), 76, rounding.HalfUp), value, true, 76, "1/"+pow10Times(2, 76))
	checkSecondsResult(t, m.MustStart(at(1, 1), 76, rounding.Down), "27725887222397812376689284858327062723020005374410210164827200379735744878787/10000000000000000000000000000000000000000000000000000000000000000000000000000", true, 76, "1/"+pow10(76))

	for _, mode := range []rounding.Mode{rounding.Down, rounding.HalfUp} {
		result, err := m.Start(at(1, 1), 77, mode)
		assert.False(t, result.IsSet())
		requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrOutOfRange)
		position, err := m.Position(sec(1, 1), 77, mode)
		assert.False(t, position.IsSet())
		requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrOutOfRange)
	}
	checkSecondsResult(t, tempomap.MustNew(quarter(60)).MustStart(at(1, 1), 77, rounding.HalfUp), "4/1", false, 0, "0/1")
}

func TestHighPrecisionChainStress(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewGradual(at(2, 1), tempomap.LinearSpeed)).
		MustWithChange(tempomap.MustNewMain(at(3, 1), quarter(240)))
	value := "21204222675845162056505138650298189547514374670924167641931625206964841686167/10000000000000000000000000000000000000000000000000000000000000000000000000000"
	checkPositionResult(t, m.MustPosition(sec(5, 1), 76, rounding.HalfUp), value, true, 76, "1/"+pow10Times(2, 76))
}

func TestAddressAtRejectsApproximationRoundedToFiniteGridEnd(t *testing.T) {
	t.Parallel()
	grid := grid44Finite(1)
	m := tempomap.MustNewMeasured(grid, quarter(60)).
		MustWithChange(tempomap.MustNewGradual(bar(1, 0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond), grid).
		MustWithChange(tempomap.MustNewMain(bar(1, 3, 4), quarter(120)), grid)
	address, result, err := m.AddressAt(grid, sec(207, 100), 0, rounding.HalfUp)
	assert.False(t, address.IsSet())
	assert.False(t, result.IsSet())
	requireTempomapError(t, err, tempomap.ErrInvalidAddress, measure.ErrOutsideGrid)
}

func pow10(places int) string {
	return "1" + strings.Repeat("0", places)
}

func pow10Times(factor, places int) string {
	return fmt.Sprintf("%d%s", factor, strings.Repeat("0", places))
}
