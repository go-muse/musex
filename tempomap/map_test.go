package tempomap_test

import (
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

func TestKP01AndKP02ConstantJumpAndInverse(t *testing.T) {
	t.Parallel()
	m := newJumpMap()
	checkSecondsResult(t, m.MustStart(at(3, 1), 0, rounding.HalfUp), "10/1", false, 0, "0/1")
	checkPositionResult(t, m.MustPosition(sec(10, 1), 0, rounding.HalfUp), "3/1", false, 0, "0/1")
}

func TestKP03TempoISeconds(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewJump(at(2, 1), quarter(120))).
		MustWithChange(tempomap.MustNewTempoI(at(3, 1)))
	checkSecondsResult(t, m.MustStart(at(4, 1), 1000, rounding.Down), "14/1", false, 0, "0/1")
}

func TestKP04ATempoAndNewMainTempo(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewJump(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewATempo(at(2, 1))).
		MustWithChange(tempomap.MustNewMain(at(3, 1), quarter(90))).
		MustWithChange(tempomap.MustNewJump(at(4, 1), quarter(120))).
		MustWithChange(tempomap.MustNewATempo(at(5, 1)))

	for _, tt := range []struct {
		name      string
		position  int
		current   tempo.Tempo
		main      tempo.Tempo
		deviating bool
	}{
		{"after a tempo", 2, quarter(60), quarter(60), false},
		{"after new main", 3, quarter(90), quarter(90), false},
		{"new deviation", 4, quarter(120), quarter(90), true},
		{"return to new main", 5, quarter(90), quarter(90), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkState(t, m.MustState(at(tt.position, 1)), tt.current, tt.main, tt.deviating)
		})
	}
}

func TestKP05TempoIChangesMainTempo(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(90))).
		MustWithChange(tempomap.MustNewJump(at(2, 1), quarter(120))).
		MustWithChange(tempomap.MustNewTempoI(at(3, 1))).
		MustWithChange(tempomap.MustNewJump(at(4, 1), quarter(120))).
		MustWithChange(tempomap.MustNewATempo(at(5, 1)))
	checkState(t, m.MustState(at(3, 1)), quarter(60), quarter(60), false)
	checkState(t, m.MustState(at(5, 1)), quarter(60), quarter(60), false)
}

func TestKP06MetricModulation(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(120)).MustWithChange(tempomap.MustNewModulation(
		at(1, 1), duration.Quarter.MustWritten().MustAddDot(), duration.MustNew(1, 4),
	))
	checkState(t, m.MustState(at(1, 1)), dottedQuarter(120), dottedQuarter(120), false)
	checkSecondsResult(t, m.MustStart(at(2, 1), 0, rounding.HalfUp), "10/3", false, 0, "0/1")
}

func TestKP07AndKP08LinearAccelerationAndRoundedSeconds(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))

	speed := m.MustSpeed(at(1, 2))
	assert.True(t, fraction.MustNew(45, 2).Equal(speed))
	assert.True(t, quarter(90).Equal(m.MustTempoAt(at(1, 2), duration.Quarter.MustWritten())))
	assert.True(t, quarter(120).Equal(m.MustTempoAt(at(1, 1), duration.Quarter.MustWritten())))

	for _, tt := range []struct {
		name  string
		mode  rounding.Mode
		value string
		bound string
	}{
		{"down", rounding.Down, "1386/500", "1/1000"},
		{"toward zero", rounding.TowardZero, "1386/500", "1/1000"},
		{"up", rounding.Up, "2773/1000", "1/1000"},
		{"away from zero", rounding.AwayFromZero, "2773/1000", "1/1000"},
		{"half up", rounding.HalfUp, "2773/1000", "1/2000"},
		{"half down", rounding.HalfDown, "2773/1000", "1/2000"},
		{"half away", rounding.HalfAwayFromZero, "2773/1000", "1/2000"},
		{"half toward", rounding.HalfTowardZero, "2773/1000", "1/2000"},
		{"half even", rounding.HalfEven, "2773/1000", "1/2000"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkSecondsResult(t, m.MustStart(at(1, 1), 3, tt.mode), tt.value, true, 3, tt.bound)
		})
	}
}

func TestKP10InverseThroughAcceleration(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))
	for _, tt := range []struct {
		mode  rounding.Mode
		value string
		bound string
	}{
		{rounding.Down, "284025/1000000", "1/1000000"},
		{rounding.Up, "284026/1000000", "1/1000000"},
		{rounding.HalfUp, "284025/1000000", "1/2000000"},
		{rounding.HalfEven, "284025/1000000", "1/2000000"},
	} {
		checkPositionResult(t, m.MustPosition(sec(1, 1), 6, tt.mode), tt.value, true, 6, tt.bound)
	}
}

func TestKP11EqualGradualEndpointsAreExact(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), eighth(120))
	checkSecondsResult(t, m.MustStart(at(1, 1), 77, rounding.HalfUp), "4/1", false, 0, "0/1")
	checkPositionResult(t, m.MustPosition(sec(4, 1), 1000, rounding.Up), "1/1", false, 0, "0/1")
}

func TestKP12SecondsToMeasureAddress(t *testing.T) {
	t.Parallel()
	grid := grid34Pickup()
	m := tempomap.MustNewMeasured(grid, quarter(60))
	address, result, err := m.AddressAt(grid, sec(9, 2), 3, rounding.HalfUp)
	require.NoError(t, err)
	assert.True(t, measure.MustNewAddress(2, pos(1, 8)).Equal(address))
	checkPositionResult(t, result, "9/8", false, 0, "0/1")
}

func TestKP13AddressingRoundTripAndTimelineEquality(t *testing.T) {
	t.Parallel()
	grid := grid34Pickup()
	measured := tempomap.MustNewMeasured(grid, quarter(60)).
		MustWithChange(tempomap.MustNewJump(bar(2, 1, 8), quarter(120)), grid).
		MustWithExtension(tempomap.MustNewExtension(bar(1, 1, 2), sec(1, 1)), grid)
	positions := measured.MustToPositions(grid)
	assert.Equal(t, tempomap.PositionAddressing, positions.MustAddressing())
	assert.False(t, measured.Equal(positions))
	assert.True(t, measured.MustEqualTimeline(positions, grid))

	roundTrip := positions.MustToMeasures(grid)
	assert.True(t, measured.Equal(roundTrip))
	assert.True(t, roundTrip.MustEqualTimeline(measured, grid, grid))
}

func TestKP14AndKP15ExtensionStartEndAndInverse(t *testing.T) {
	t.Parallel()
	m := newHoldMap()
	checkSecondsResult(t, m.MustStart(at(1, 2), 0, rounding.HalfUp), "2/1", false, 0, "0/1")
	checkSecondsResult(t, m.MustEnd(at(1, 2), 0, rounding.HalfUp), "3/1", false, 0, "0/1")
	checkSecondsResult(t, m.MustStart(at(3, 4), 0, rounding.HalfUp), "4/1", false, 0, "0/1")

	for _, tt := range []struct {
		seconds tempo.Seconds
		want    string
	}{
		{sec(2, 1), "1/2"},
		{sec(5, 2), "1/2"},
		{sec(3, 1), "1/2"},
		{sec(4, 1), "3/4"},
	} {
		checkPositionResult(t, m.MustPosition(tt.seconds, 0, rounding.HalfUp), tt.want, false, 0, "0/1")
	}
}

func TestMeasuredStartEndAndBetween(t *testing.T) {
	t.Parallel()
	grid := grid34Pickup()
	m := tempomap.MustNewMeasured(grid, quarter(60)).MustWithExtension(
		tempomap.MustNewExtension(bar(1, 1, 2), sec(1, 1)), grid,
	)
	checkSecondsResult(t, m.MustStart(bar(1, 1, 2), 0, rounding.HalfUp, grid), "3/1", false, 0, "0/1")
	checkSecondsResult(t, m.MustEnd(bar(1, 1, 2), 0, rounding.HalfUp, grid), "4/1", false, 0, "0/1")
	checkSecondsResult(t, m.MustBetween(bar(1, 1, 2), bar(1, 3, 4), 0, rounding.HalfUp, grid), "2/1", false, 0, "0/1")
}

func TestKP16AndKP17IntervalsWithExtensions(t *testing.T) {
	t.Parallel()
	m := newHoldMap()
	checkSecondsResult(t, m.MustBetween(at(1, 2), at(3, 4), 0, rounding.HalfUp), "2/1", false, 0, "0/1")

	atOrigin := tempomap.MustNew(quarter(60)).MustWithExtension(
		tempomap.MustNewExtension(tempomap.MustAt(duration.Origin()), sec(1, 1)),
	)
	checkSecondsResult(t, atOrigin.MustStart(tempomap.MustAt(duration.Origin()), 0, rounding.HalfUp), "0/1", false, 0, "0/1")
	checkSecondsResult(t, atOrigin.MustEnd(tempomap.MustAt(duration.Origin()), 0, rounding.HalfUp), "1/1", false, 0, "0/1")
	checkSecondsResult(t, atOrigin.MustBetween(tempomap.MustAt(duration.Origin()), at(1, 4), 0, rounding.HalfUp), "2/1", false, 0, "0/1")
}

func TestKP18ZeroExtensionIsRecordedButNotTemporal(t *testing.T) {
	t.Parallel()
	base := tempomap.MustNew(quarter(60))
	withZero := base.MustWithExtension(tempomap.MustNewExtension(at(1, 4), tempo.ZeroSeconds()))
	require.Len(t, withZero.MustExtensions(), 1)
	assert.Equal(t, withZero.MustStart(at(1, 4), 0, rounding.HalfUp), withZero.MustEnd(at(1, 4), 0, rounding.HalfUp))
	assert.False(t, base.Equal(withZero))
	assert.True(t, base.MustEqualTimeline(withZero))
	checkPositionResult(t, withZero.MustPosition(sec(1, 1), 1000, rounding.HalfUp), "1/4", false, 0, "0/1")
}

func TestKP19ApproximationSurvivesExactExtension(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120)).MustWithExtension(
		tempomap.MustNewExtension(at(1, 1), sec(1, 4)),
	)
	checkSecondsResult(t, m.MustStart(at(1, 1), 3, rounding.HalfUp), "2773/1000", true, 3, "1/2000")
	checkSecondsResult(t, m.MustEnd(at(1, 1), 3, rounding.HalfUp), "3023/1000", true, 3, "1/2000")
}

func TestKP20Segments(t *testing.T) {
	t.Parallel()
	segments := newJumpMap().MustSegments()
	require.Len(t, segments, 2)

	first := segments[0]
	assert.True(t, first.MustChange().Equal(tempomap.MustNewMain(tempomap.MustAt(duration.Origin()), quarter(60)).MustWithSlot(tempomap.SlotFirst)))
	assert.True(t, first.MustStart().Equal(tempomap.MustAt(duration.Origin())))
	assert.True(t, fraction.FromInt(15).Equal(first.MustStartSpeed()))
	end, ok := first.MustEnd()
	require.True(t, ok)
	assert.True(t, at(2, 1).Equal(end))
	endSpeed, ok := first.MustEndSpeed()
	require.True(t, ok)
	assert.True(t, fraction.FromInt(15).Equal(endSpeed))
	seconds, ok := first.MustSeconds(1000, rounding.HalfUp)
	require.True(t, ok)
	checkSecondsResult(t, seconds, "8/1", false, 0, "0/1")

	last := segments[1]
	_, ok = last.MustEnd()
	assert.False(t, ok)
	_, ok = last.MustEndSpeed()
	assert.False(t, ok)
	_, ok = last.MustSeconds(1000, rounding.HalfUp)
	assert.False(t, ok)
	assert.True(t, last.Equal(segments[1]))
	assert.NotEmpty(t, last.String())
}

func TestKP33OriginAndZeroSeconds(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	origin := tempomap.MustAt(duration.Origin())
	assert.Equal(t, m.MustStart(origin, 0, rounding.HalfUp), m.MustEnd(origin, 1000, rounding.Down))
	checkSecondsResult(t, m.MustStart(origin, 0, rounding.HalfUp), "0/1", false, 0, "0/1")
	checkPositionResult(t, m.MustPosition(tempo.ZeroSeconds(), 0, rounding.HalfUp), "0/1", false, 0, "0/1")
}

func TestKP34ChangeAndBarlineBoundaries(t *testing.T) {
	t.Parallel()
	m := newJumpMap()
	checkState(t, m.MustState(at(2, 1)), quarter(120), quarter(60), true)
	checkSecondsResult(t, m.MustStart(at(2, 1), 0, rounding.HalfUp), "8/1", false, 0, "0/1")
	checkPositionResult(t, m.MustPosition(sec(8, 1), 0, rounding.HalfUp), "2/1", false, 0, "0/1")

	grid := grid34Pickup()
	measured := tempomap.MustNewMeasured(grid, quarter(60))
	address, result := measured.MustAddressAt(grid, sec(1, 1), 0, rounding.HalfUp)
	assert.True(t, measure.MustNewAddress(1, duration.Origin()).Equal(address))
	checkPositionResult(t, result, "1/4", false, 0, "0/1")
}

func TestKP35RepeatedTempoAffectsRecordNotTimeline(t *testing.T) {
	t.Parallel()
	plain := tempomap.MustNew(quarter(60))
	repeated := plain.MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(60)))
	assert.False(t, plain.Equal(repeated))
	assert.True(t, plain.MustEqualTimeline(repeated))
	require.Len(t, repeated.MustChanges(), 2)
}

func TestTimelineRejectsDifferentSpeeds(t *testing.T) {
	t.Parallel()
	slow := tempomap.MustNew(quarter(60))
	fast := tempomap.MustNew(quarter(120))
	assert.False(t, slow.MustEqualTimeline(fast))
	assert.False(t, slow.Equal(fast))
}

func TestKP36GradualCompletionByATempoAndTempoI(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		returnKind tempomap.ChangeKind
		mid        tempo.Tempo
		final      tempo.Tempo
		exact      bool
	}{
		{"a tempo unequal", tempomap.ChangeATempo, quarter(90), quarter(60), false},
		{"a tempo equal", tempomap.ChangeATempo, eighth(120), quarter(60), true},
		{"tempo i unequal", tempomap.ChangeTempoI, quarter(90), quarter(60), false},
		{"tempo i equal", tempomap.ChangeTempoI, eighth(120), quarter(60), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := tempomap.MustNew(quarter(60)).
				MustWithChange(tempomap.MustNewMain(at(1, 4), tt.mid)).
				MustWithChange(tempomap.MustNewGradual(at(1, 2), tempomap.LinearSpeed))
			var change tempomap.Change
			if tt.returnKind == tempomap.ChangeATempo {
				change = tempomap.MustNewATempo(at(1, 1))
			} else {
				change = tempomap.MustNewTempoI(at(1, 1))
			}
			m = m.MustWithChange(change)
			checkState(t, m.MustState(at(1, 1)), tt.final, tt.final, false)
			if tt.exact {
				checkSecondsResult(t, m.MustStart(at(1, 1), 1000, rounding.HalfUp), "4/1", false, 0, "0/1")
			} else {
				result := m.MustStart(at(1, 1), 6, rounding.HalfUp)
				approximate, err := result.Approximate()
				require.NoError(t, err)
				assert.True(t, approximate)
			}
		})
	}
}

func TestKP37OpenFinalConstantSegment(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	checkSecondsResult(t, m.MustStart(at(1000000, 1), 1000, rounding.HalfUp), "4000000/1", false, 0, "0/1")
}

func TestKP41ExactResultIgnoresPlacesButValidatesThem(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	for _, places := range []int{0, 76, 77, 1000} {
		checkSecondsResult(t, m.MustStart(at(1, 1), places, rounding.HalfUp), "4/1", false, 0, "0/1")
	}
	for _, places := range []int{-1, 1001} {
		result, err := m.Start(at(1, 1), places, rounding.HalfUp)
		assert.False(t, result.IsSet())
		requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	}
}

func TestKP42ExtensionAtChangeDoesNotChangeState(t *testing.T) {
	t.Parallel()
	m := newJumpMap().MustWithExtension(tempomap.MustNewExtension(at(2, 1), sec(1, 1)))
	checkState(t, m.MustState(at(2, 1)), quarter(120), quarter(60), true)
	checkSecondsResult(t, m.MustStart(at(2, 1), 0, rounding.HalfUp), "8/1", false, 0, "0/1")
	checkSecondsResult(t, m.MustEnd(at(2, 1), 0, rounding.HalfUp), "9/1", false, 0, "0/1")
	checkSecondsResult(t, m.MustStart(at(5, 2), 0, rounding.HalfUp), "10/1", false, 0, "0/1")
}

func TestKP43AdjacentNotesUseEndAtBothBoundaries(t *testing.T) {
	t.Parallel()
	m := newHoldMap()
	endA := m.MustEnd(at(1, 2), 0, rounding.HalfUp)
	startB := m.MustEnd(at(1, 2), 0, rounding.HalfUp)
	endB := m.MustEnd(at(3, 4), 0, rounding.HalfUp)
	startA := m.MustEnd(tempomap.MustAt(duration.Origin()), 0, rounding.HalfUp)
	checkSecondsResult(t, startA, "0/1", false, 0, "0/1")
	checkSecondsResult(t, endA, "3/1", false, 0, "0/1")
	checkSecondsResult(t, startB, "3/1", false, 0, "0/1")
	checkSecondsResult(t, endB, "4/1", false, 0, "0/1")
	assert.Equal(t, endA, startB)
	checkSecondsResult(t, m.MustBetween(at(1, 2), at(3, 4), 0, rounding.HalfUp), "2/1", false, 0, "0/1")
}

func TestKP47InitialPairEditing(t *testing.T) {
	t.Parallel()
	base := tempomap.MustNew(quarter(60))
	origin := tempomap.MustAt(duration.Origin())
	gradual := tempomap.MustNewGradual(origin, tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)
	pair := base.MustWithChange(gradual).MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(120)))
	require.Len(t, pair.MustChanges(), 2)
	assert.Equal(t, tempomap.SlotFirst, pair.MustChanges()[0].MustSlot())
	assert.Equal(t, tempomap.SlotSecond, pair.MustChanges()[1].MustSlot())

	replacedFirst := pair.MustWithChange(tempomap.MustNewMain(origin, quarter(90)).MustWithSlot(tempomap.SlotFirst))
	require.Len(t, replacedFirst.MustChanges(), 2)
	assert.True(t, quarter(90).Equal(replacedFirst.MustInitial()))
	assert.Equal(t, tempomap.ChangeGradual, replacedFirst.MustChanges()[1].MustKind())

	replacedSecond := pair.MustWithChange(gradual)
	assert.True(t, pair.Equal(replacedSecond))
	removedSecond := pair.MustWithoutChange(origin, tempomap.SlotSecond)
	assert.True(t, base.Equal(removedSecond))

	before := pair.String()
	third, err := pair.WithChange(tempomap.MustNewJump(origin, quarter(180)))
	assert.False(t, third.IsSet())
	requireTempomapError(t, err, tempomap.ErrDuplicateAddress)
	assert.Equal(t, before, pair.String())

	removedFirst, err := pair.WithoutChange(origin, tempomap.SlotFirst)
	assert.False(t, removedFirst.IsSet())
	requireTempomapError(t, err, tempomap.ErrNoInitialTempo)
	assert.Equal(t, before, pair.String())
}

func TestKP48InitialPairHasNoZeroLengthSegment(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))
	segments := m.MustSegments()
	require.Len(t, segments, 1)
	segment := segments[0]
	assert.Equal(t, tempomap.ChangeGradual, segment.MustChange().MustKind())
	assert.True(t, tempomap.MustAt(duration.Origin()).Equal(segment.MustStart()))
	end, ok := segment.MustEnd()
	require.True(t, ok)
	assert.True(t, at(1, 1).Equal(end))
	assert.True(t, fraction.FromInt(15).Equal(segment.MustStartSpeed()))
	endSpeed, ok := segment.MustEndSpeed()
	require.True(t, ok)
	assert.True(t, fraction.FromInt(30).Equal(endSpeed))
}
