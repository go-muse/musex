package tempomap_test

import (
	"errors"
	"math"
	"math/big"
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

func TestKP21InitialChangeMustBeMain(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"tempomap[]",
		"tempomap[single@p:0/1:jump:4=60]",
		"tempomap[second@p:0/1:gradual:linear-speed]",
		"tempomap[single@p:0/1:a-tempo]",
		"tempomap[single@p:0/1:tempo-i]",
		"tempomap[single@p:0/1:modulation:4.@1/4]",
		"tempomap[single@p:0/1:main:4=60]",
	} {
		m, err := tempomap.Parse(text)
		assert.False(t, m.IsSet())
		requireTempomapError(t, err, tempomap.ErrNoInitialTempo)
	}
}

func TestKP22DuplicateRulesAndCompatibleHold(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"tempomap[first@p:0/1:main:4=60;single@p:1/1:jump:4=90;single@p:1/1:main:4=120]",
		"tempomap[first@p:0/1:main:4=60;hold@p:1/2:1/1;hold@p:1/2:2/1]",
	} {
		_, err := tempomap.Parse(text)
		requireTempomapError(t, err, tempomap.ErrDuplicateAddress)
	}

	m, err := tempomap.Parse("tempomap[first@p:0/1:main:4=60;single@p:1/2:jump:4=90;second@p:0/1:gradual:linear-speed]")
	assert.False(t, m.IsSet())
	requireTempomapError(t, err, tempomap.ErrDuplicateAddress)

	valid := tempomap.MustParse("tempomap[first@p:0/1:main:4=60;second@p:0/1:gradual:linear-speed;single@p:1/1:main:4=120;hold@p:1/1:1/1]")
	require.Len(t, valid.MustChanges(), 3)
	require.Len(t, valid.MustExtensions(), 1)
}

func TestKP23ATempoRequiresCurrentDeviation(t *testing.T) {
	t.Parallel()
	base := tempomap.MustNew(quarter(60))
	before := base.String()
	m, err := base.WithChange(tempomap.MustNewATempo(at(1, 1)))
	assert.False(t, m.IsSet())
	requireTempomapError(t, err, tempomap.ErrNoDeviation)
	assert.Equal(t, before, base.String())

	m = base.MustWithChange(tempomap.MustNewJump(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewATempo(at(2, 1)))
	before = m.String()
	m, err = m.WithChange(tempomap.MustNewATempo(at(3, 1)))
	assert.False(t, m.IsSet())
	requireTempomapError(t, err, tempomap.ErrNoDeviation)
	assert.Equal(t, before, m.String())
}

func TestKP24InvalidGradualSequences(t *testing.T) {
	t.Parallel()
	gradualTail := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed))
	for _, tt := range []struct {
		name   string
		change tempomap.Change
	}{
		{"gradual after gradual", tempomap.MustNewGradual(at(2, 1), tempomap.LinearSpeed)},
		{"modulation after gradual", tempomap.MustNewModulation(at(2, 1), duration.Quarter.MustWritten().MustAddDot(), duration.MustNew(1, 4))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before := gradualTail.String()
			m, err := gradualTail.WithChange(tt.change)
			assert.False(t, m.IsSet())
			requireTempomapError(t, err, tempomap.ErrInvalidSequence)
			assert.Equal(t, before, gradualTail.String())
		})
	}
}

func TestKP25UnfinishedGradualTail(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed))
	unfinished, err := m.Validate()
	require.NoError(t, err)
	assert.True(t, unfinished)
	assert.True(t, m.MustUnfinished())

	checkSecondsResult(t, m.MustStart(at(1, 2), 1000, rounding.HalfUp), "2/1", false, 0, "0/1")
	assert.True(t, fraction.FromInt(15).Equal(m.MustSpeed(at(1, 2))))
	segments := m.MustSegments()
	require.Len(t, segments, 1)
	seconds, ok := segments[0].MustSeconds(0, rounding.HalfUp)
	require.True(t, ok)
	checkSecondsResult(t, seconds, "4/1", false, 0, "0/1")

	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"start at tail", func() error { _, err := m.Start(at(1, 1), 0, rounding.HalfUp); return err }},
		{"start after tail", func() error { _, err := m.Start(at(2, 1), 0, rounding.HalfUp); return err }},
		{"speed at tail", func() error { _, err := m.Speed(at(1, 1)); return err }},
		{"tempo at tail", func() error { _, err := m.TempoAt(at(1, 1), duration.Quarter.MustWritten()); return err }},
		{"state at tail", func() error { _, err := m.State(at(1, 1)); return err }},
		{"position after tail", func() error { _, err := m.Position(sec(5, 1), 0, rounding.HalfUp); return err }},
		{"between across tail", func() error { _, err := m.Between(at(1, 2), at(2, 1), 0, rounding.HalfUp); return err }},
		{"timeline equality", func() error { _, err := m.EqualTimeline(tempomap.MustNew(quarter(60))); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, operation.call(), tempomap.ErrUnfinished)
		})
	}

	completed := m.MustWithChange(tempomap.MustNewMain(at(2, 1), quarter(120)))
	assert.False(t, completed.MustUnfinished())
	unfinished, err = completed.Validate()
	require.NoError(t, err)
	assert.False(t, unfinished)
}

func TestKP26MeasuredInitialAndGridAddresses(t *testing.T) {
	t.Parallel()
	withPickup := grid34Pickup()
	withoutPickup := measure.MustNew(withPickup.MustBarMeter(1))
	for _, tt := range []struct {
		name string
		text string
		grid measure.Grid
	}{
		{"pickup requires bar zero", "tempomap[first@m:1:0/1:main:4=60]", withPickup},
		{"no pickup requires bar one", "tempomap[first@m:0:0/1:main:4=60]", withoutPickup},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tempomap.Parse(tt.text, tt.grid)
			requireTempomapError(t, err, tempomap.ErrInvalidAddress)
		})
	}

	finite := grid44Finite(1)
	m := tempomap.MustNewMeasured(finite, quarter(60))
	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"change beyond grid", func() error {
			_, err := m.WithChange(tempomap.MustNewJump(bar(2, 0, 1), quarter(120)), finite)
			return err
		}},
		{"extension beyond grid", func() error {
			_, err := m.WithExtension(tempomap.MustNewExtension(bar(2, 0, 1), sec(1, 1)), finite)
			return err
		}},
		{"offset outside bar", func() error {
			_, err := m.WithExtension(tempomap.MustNewExtension(bar(1, 1, 1), sec(1, 1)), finite)
			return err
		}},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			err := operation.call()
			require.Error(t, err)
			assert.ErrorIs(t, err, tempomap.ErrInvalidAddress)
			assert.True(t,
				errorIsAny(err, measure.ErrOutsideGrid, measure.ErrOutsideBar),
				"got %v, want a transparent measure address cause", err,
			)
		})
	}
}

func TestKP27PositionToMeasuresRejectsGridEndAndBeyond(t *testing.T) {
	t.Parallel()
	grid := grid44Finite(1)
	for _, tt := range []struct {
		name  string
		build func(tempomap.Map) tempomap.Map
	}{
		{"change at end", func(m tempomap.Map) tempomap.Map {
			return m.MustWithChange(tempomap.MustNewJump(at(1, 1), quarter(120)))
		}},
		{"change beyond end", func(m tempomap.Map) tempomap.Map {
			return m.MustWithChange(tempomap.MustNewJump(at(2, 1), quarter(120)))
		}},
		{"extension at end", func(m tempomap.Map) tempomap.Map {
			return m.MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
		}},
		{"extension beyond end", func(m tempomap.Map) tempomap.Map {
			return m.MustWithExtension(tempomap.MustNewExtension(at(2, 1), sec(1, 1)))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := tt.build(tempomap.MustNew(quarter(60)))
			before := m.String()
			converted, err := m.ToMeasures(grid)
			assert.False(t, converted.IsSet())
			requireTempomapError(t, err, tempomap.ErrInvalidAddress, measure.ErrOutsideGrid)
			assert.Equal(t, before, m.String())
		})
	}
}

func TestKP28DependencyTypesRejectNegativeInputs(t *testing.T) {
	t.Parallel()
	_, err := duration.NewPosition(-1, 1)
	assert.ErrorIs(t, err, duration.ErrNegativePosition)
	_, err = tempo.NewSeconds(-1, 1)
	assert.ErrorIs(t, err, tempo.ErrNegativeTime)
	_, err = tempomap.NewExtension(at(1, 2), tempo.Seconds{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, tempo.ErrUnset)
}

func TestKP29PrecisionArgumentsAndRoundingModes(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))
	for _, places := range []int{-1, 1001} {
		for _, operation := range []struct {
			name string
			call func() error
		}{
			{"start", func() error { _, err := m.Start(at(1, 1), places, rounding.HalfUp); return err }},
			{"position", func() error { _, err := m.Position(sec(1, 1), places, rounding.HalfUp); return err }},
			{"segment", func() error { _, _, err := m.MustSegments()[0].Seconds(places, rounding.HalfUp); return err }},
		} {
			t.Run(operation.name, func(t *testing.T) {
				t.Parallel()
				requireTempomapError(t, operation.call(), tempomap.ErrInvalidArgument)
			})
		}
	}

	for _, tt := range []struct {
		name       string
		mode       rounding.Mode
		dependency error
	}{
		{"unset", rounding.Mode(0), rounding.ErrUnset},
		{"unknown", rounding.Mode(255), rounding.ErrUnknownMode},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, operation := range []struct {
				name string
				call func() error
			}{
				{"start", func() error { _, err := m.Start(at(1, 1), 3, tt.mode); return err }},
				{"position", func() error { _, err := m.Position(sec(1, 1), 3, tt.mode); return err }},
				{"segment", func() error { _, _, err := m.MustSegments()[0].Seconds(3, tt.mode); return err }},
				{"open segment still validates", func() error {
					_, _, err := tempomap.MustNew(quarter(60)).MustSegments()[0].Seconds(3, tt.mode)
					return err
				}},
			} {
				t.Run(operation.name, func(t *testing.T) {
					t.Parallel()
					requireTempomapError(t, operation.call(), tempomap.ErrInvalidArgument, tt.dependency)
				})
			}
		})
	}
}

func TestKP30RemovalValidatesFollowingStates(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewJump(at(1, 1), quarter(120))).
		MustWithChange(tempomap.MustNewATempo(at(2, 1)))
	before := m.String()
	changed, err := m.WithoutChange(at(1, 1), tempomap.SlotSingle)
	assert.False(t, changed.IsSet())
	requireTempomapError(t, err, tempomap.ErrNoDeviation)
	assert.Equal(t, before, m.String())

	changed, err = m.WithoutChange(tempomap.MustAt(duration.Origin()), tempomap.SlotFirst)
	assert.False(t, changed.IsSet())
	requireTempomapError(t, err, tempomap.ErrNoInitialTempo)
	assert.Equal(t, before, m.String())

	assert.True(t, m.Equal(m.MustWithoutChange(at(100, 1), tempomap.SlotSingle)))
	assert.True(t, m.Equal(m.MustWithoutExtension(at(100, 1))))
}

func TestKP31UnsetInputsAndDependencySentinels(t *testing.T) {
	t.Parallel()
	var unset tempomap.Map
	result, err := unset.Start(at(1, 1), 0, rounding.HalfUp)
	assert.False(t, result.IsSet())
	requireTempomapError(t, err, tempomap.ErrUnset)

	grid := grid34Pickup()
	measured := tempomap.MustNewMeasured(grid, quarter(60))
	_, err = measured.Start(bar(0, 0, 1), 0, rounding.HalfUp, measure.Grid{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, measure.ErrUnset)

	_, err = tempomap.MustNew(quarter(60)).Position(tempo.Seconds{}, 0, rounding.HalfUp)
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, tempo.ErrUnset)

	_, err = tempomap.New(tempo.Tempo{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, tempo.ErrUnset)

	_, err = tempomap.NewMain(at(1, 1), tempo.Tempo{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, tempo.ErrUnset)
}

func TestKP46FiniteGridEndIsNotAnAddress(t *testing.T) {
	t.Parallel()
	grid := grid44Finite(1)
	m := tempomap.MustNewMeasured(grid, quarter(60))
	lastValid := bar(1, 3, 4)
	m = m.MustWithChange(tempomap.MustNewJump(lastValid, quarter(120)), grid)
	m = m.MustWithExtension(tempomap.MustNewExtension(lastValid, sec(1, 1)), grid)
	assert.True(t, m.MustState(lastValid, grid).IsSet())
	assert.True(t, m.MustStart(lastValid, 0, rounding.HalfUp, grid).IsSet())

	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"change at end", func() error {
			_, err := m.WithChange(tempomap.MustNewJump(bar(1, 1, 1), quarter(120)), grid)
			return err
		}},
		{"extension at end", func() error {
			_, err := m.WithExtension(tempomap.MustNewExtension(bar(1, 1, 1), sec(1, 1)), grid)
			return err
		}},
		{"state at end", func() error { _, err := m.State(bar(1, 1, 1), grid); return err }},
		{"start at end", func() error { _, err := m.Start(bar(1, 1, 1), 0, rounding.HalfUp, grid); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			err := operation.call()
			require.Error(t, err)
			assert.ErrorIs(t, err, tempomap.ErrInvalidAddress)
			assert.True(t, errorIsAny(err, measure.ErrOutsideGrid, measure.ErrOutsideBar))
		})
	}
}

func TestBetweenRejectsReverseAndHandlesEqualAddresses(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	result, err := m.Between(at(1, 1), at(0, 1), 0, rounding.HalfUp)
	assert.False(t, result.IsSet())
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	checkSecondsResult(t, m.MustBetween(at(1, 1), at(1, 1), 1000, rounding.HalfUp), "0/1", false, 0, "0/1")

	unfinished := m.MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed))
	_, err = unfinished.Between(at(1, 1), at(1, 1), 0, rounding.HalfUp)
	requireTempomapError(t, err, tempomap.ErrUnfinished)
}

func TestStateInsideCompletedGradualIsInvalid(t *testing.T) {
	t.Parallel()
	m := newGradualMap(pos(1, 1), quarter(120))
	state, err := m.State(at(1, 2))
	assert.False(t, state.IsSet())
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	checkState(t, m.MustState(at(1, 1)), quarter(120), quarter(120), false)
}

func TestInitialSpeedOutOfRange(t *testing.T) {
	t.Parallel()
	initial := tempo.MustNew(duration.Maxima.MustWritten(), fractionRat(fraction.MaxComponentString))
	m, err := tempomap.New(initial)
	assert.False(t, m.IsSet())
	requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrOutOfRange, tempo.ErrOutOfRange)
}

func TestErrorPrecedence(t *testing.T) {
	t.Parallel()
	var unsetMap tempomap.Map
	_, err := unsetMap.Start(tempomap.Address{}, -1, rounding.Mode(0), measure.Grid{})
	requireTempomapError(t, err, tempomap.ErrUnset)

	_, err = tempomap.NewMain(tempomap.Address{}, tempo.Tempo{})
	requireTempomapError(t, err, tempomap.ErrUnset)

	m := tempomap.MustNew(quarter(60))
	_, err = m.Start(at(1, 1), -1, rounding.Mode(0))
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	_, err = m.Start(at(1, 1), 0, rounding.Mode(0))
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, rounding.ErrUnset)

	_, err = m.Position(tempo.Seconds{}, -1, rounding.Mode(0))
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, tempo.ErrUnset)

	_, err = m.Between(tempomap.Address{}, tempomap.Address{}, 0, rounding.HalfUp)
	requireTempomapError(t, err, tempomap.ErrUnset)

	_, err = m.WithChange(tempomap.Change{}, measure.Grid{}, measure.Grid{})
	requireTempomapError(t, err, tempomap.ErrUnset)

	_, err = m.Start(at(1, 1), 0, rounding.HalfUp, grid34Pickup())
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)

	_, err = m.ToPositions(grid34Pickup())
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
}

func TestOpenGridOutOfRangeIsNotInvalidAddress(t *testing.T) {
	t.Parallel()
	// 2^254 is valid; the sum of the first two explicit bars is 2^255,
	// one greater than fraction.MaxComponent.
	huge := duration.MustFromFraction(fraction.MustFromBigRat(new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 254))))
	grid := measure.MustNew(grid44Finite(1).MustBarMeter(1)).
		MustWithLength(1, huge, measure.AlignStart).
		MustWithLength(2, huge, measure.AlignStart)
	m := tempomap.MustParse("tempomap[first@m:1:0/1:main:4=60;single@m:3:0/1:jump:4=120;hold@m:3:0/1:1/1]")

	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"state", func() error { _, err := m.State(bar(3, 0, 1), grid); return err }},
		{"speed", func() error { _, err := m.Speed(bar(3, 0, 1), grid); return err }},
		{"start", func() error { _, err := m.Start(bar(3, 0, 1), 0, rounding.HalfUp, grid); return err }},
		{"editing", func() error {
			_, err := m.WithChange(tempomap.MustNewJump(bar(3, 0, 1), quarter(180)), grid)
			return err
		}},
		{"extension", func() error {
			_, err := m.WithExtension(tempomap.MustNewExtension(bar(3, 0, 1), sec(2, 1)), grid)
			return err
		}},
		{"conversion", func() error { _, err := m.ToPositions(grid); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, operation.call(), tempomap.ErrOutOfRange, measure.ErrOutOfRange, fraction.ErrOutOfRange)
		})
	}

	open := measure.MustNew(grid44Finite(1).MustBarMeter(1))
	measured := tempomap.MustNewMeasured(open, quarter(60))
	seconds := secondsRat(new(big.Rat).Mul(big.NewRat(4, 1), new(big.Rat).SetInt64(math.MaxInt64)).RatString())
	address, result, err := measured.AddressAt(open, seconds, 0, rounding.HalfUp)
	assert.False(t, address.IsSet())
	assert.False(t, result.IsSet())
	requireTempomapError(t, err, tempomap.ErrOutOfRange, measure.ErrOutOfRange, fraction.ErrOutOfRange)
}

func errorIsAny(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
