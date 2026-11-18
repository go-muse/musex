package tempomap_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gradualBoundaryMapText = "tempomap[first@p:0/1:main:4=60;second@p:0/1:gradual:linear-speed;single@p:1/1:main:4=120]"

func parseRegressionMap(t *testing.T, text string) tempomap.Map {
	t.Helper()

	var m tempomap.Map
	var err error
	require.NotPanics(t, func() {
		m, err = tempomap.Parse(text)
	})
	require.NoError(t, err)
	return m
}

func regressionAddress(position string) tempomap.Address {
	return tempomap.MustAt(duration.MustParsePosition(position))
}

func assertSecondsResult(t *testing.T, expectedText string, actual tempomap.SecondsResult) {
	t.Helper()

	expected := tempomap.MustParseSecondsResult(expectedText)
	assert.True(t, actual.Equal(expected), "seconds result = %s, want %s", actual.String(), expected.String())
}

func assertMapSnapshot(t *testing.T, snapshot string, actual tempomap.Map) {
	t.Helper()

	assert.Equal(t, snapshot, actual.String())
	assert.True(t, actual.Equal(parseRegressionMap(t, snapshot)), "map record changed")
}

func assertMustPanic(t *testing.T, baseErr error, call func()) {
	t.Helper()

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		call()
	}()
	require.NotNil(t, recovered)
	actual, ok := recovered.(error)
	require.True(t, ok, "panic value has type %T, want error", recovered)
	assert.IsType(t, baseErr, actual)
	assert.EqualError(t, actual, baseErr.Error())
}

func TestRationalBoundaryTimeWithUnrepresentableIntermediateSpeed(t *testing.T) {
	t.Parallel()

	probe := regressionAddress("1/" + fraction.MaxComponentBig().String())
	origin := regressionAddress("0/1")
	tests := []struct {
		name     string
		call     func(tempomap.Map) (tempomap.SecondsResult, error)
		mustCall func(tempomap.Map) tempomap.SecondsResult
	}{
		{
			name: "Start",
			call: func(m tempomap.Map) (tempomap.SecondsResult, error) {
				return m.Start(probe, 3, rounding.HalfUp)
			},
			mustCall: func(m tempomap.Map) tempomap.SecondsResult {
				return m.MustStart(probe, 3, rounding.HalfUp)
			},
		},
		{
			name: "End_without_hold",
			call: func(m tempomap.Map) (tempomap.SecondsResult, error) {
				return m.End(probe, 3, rounding.HalfUp)
			},
			mustCall: func(m tempomap.Map) tempomap.SecondsResult {
				return m.MustEnd(probe, 3, rounding.HalfUp)
			},
		},
		{
			name: "Between_origin_and_probe",
			call: func(m tempomap.Map) (tempomap.SecondsResult, error) {
				return m.Between(origin, probe, 3, rounding.HalfUp)
			},
			mustCall: func(m tempomap.Map) tempomap.SecondsResult {
				return m.MustBetween(origin, probe, 3, rounding.HalfUp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := parseRegressionMap(t, gradualBoundaryMapText)
			snapshot := m.String()
			var actual tempomap.SecondsResult
			var err error
			require.NotPanics(t, func() {
				actual, err = tt.call(m)
			})
			require.NoError(t, err)
			assertSecondsResult(t, "s~0/1@3:1/2000", actual)

			var mustActual tempomap.SecondsResult
			require.NotPanics(t, func() {
				mustActual = tt.mustCall(m)
			})
			assertSecondsResult(t, "s~0/1@3:1/2000", mustActual)

			if tt.name == "Between_origin_and_probe" {
				require.NotPanics(t, func() {
					actual, err = m.Between(probe, origin, 3, rounding.HalfUp)
				})
				require.Error(t, err)
				assert.Equal(t, tempomap.SecondsResult{}, actual)
				assert.ErrorIs(t, err, tempomap.ErrInvalidArgument)
			}
			assertMapSnapshot(t, snapshot, m)
		})
	}

	t.Run("Speed_and_TempoAt_remain_bounded", func(t *testing.T) {
		t.Parallel()

		m := parseRegressionMap(t, gradualBoundaryMapText)
		snapshot := m.String()
		var speed fraction.Fraction
		var speedErr error
		require.NotPanics(t, func() {
			speed, speedErr = m.Speed(probe)
		})
		require.Error(t, speedErr)
		assert.Equal(t, fraction.Fraction{}, speed)
		assert.ErrorIs(t, speedErr, tempomap.ErrOutOfRange)
		assert.ErrorIs(t, speedErr, fraction.ErrOutOfRange)
		assertMustPanic(t, speedErr, func() {
			m.MustSpeed(probe)
		})

		var marking tempo.Tempo
		var tempoErr error
		require.NotPanics(t, func() {
			marking, tempoErr = m.TempoAt(probe, duration.Quarter.MustWritten())
		})
		require.Error(t, tempoErr)
		assert.Equal(t, tempo.Tempo{}, marking)
		assert.ErrorIs(t, tempoErr, tempomap.ErrOutOfRange)
		assert.ErrorIs(t, tempoErr, fraction.ErrOutOfRange)
		assertMustPanic(t, tempoErr, func() {
			m.MustTempoAt(probe, duration.Quarter.MustWritten())
		})
		assertMapSnapshot(t, snapshot, m)
	})
}

func TestRationalBoundaryHoldAtSamePosition(t *testing.T) {
	t.Parallel()

	probe := regressionAddress("1/" + fraction.MaxComponentBig().String())
	origin := regressionAddress("0/1")
	initial := tempo.MustNewQuarter(fraction.FromInt(60))
	final := tempo.MustNewQuarter(fraction.FromInt(120))

	var base tempomap.Map
	var err error
	require.NotPanics(t, func() {
		base, err = tempomap.New(initial)
	})
	require.NoError(t, err)
	gradual := tempomap.MustNewGradual(origin, tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)
	require.NotPanics(t, func() {
		base, err = base.WithChange(gradual)
	})
	require.NoError(t, err)
	last := tempomap.MustNewMain(regressionAddress("1/1"), final)
	require.NotPanics(t, func() {
		base, err = base.WithChange(last)
	})
	require.NoError(t, err)
	baseSnapshot := base.String()

	hold := tempomap.MustNewExtension(probe, tempo.MustNewSeconds(1, 1))
	var withHold tempomap.Map
	require.NotPanics(t, func() {
		withHold, err = base.WithExtension(hold)
	})
	require.NoError(t, err)
	assertMapSnapshot(t, baseSnapshot, base)
	holdSnapshot := withHold.String()

	var segments []tempomap.Segment
	require.NotPanics(t, func() {
		segments, err = withHold.Segments()
	})
	require.NoError(t, err)
	require.Len(t, segments, 2)
	require.NotPanics(t, func() {
		segments = withHold.MustSegments()
	})
	require.Len(t, segments, 2)

	var start tempomap.SecondsResult
	require.NotPanics(t, func() {
		start, err = withHold.Start(probe, 3, rounding.HalfUp)
	})
	require.NoError(t, err)
	assertSecondsResult(t, "s~0/1@3:1/2000", start)

	var end tempomap.SecondsResult
	require.NotPanics(t, func() {
		end, err = withHold.End(probe, 3, rounding.HalfUp)
	})
	require.NoError(t, err)
	assertSecondsResult(t, "s~1/1@3:1/2000", end)

	var mustStart, mustEnd tempomap.SecondsResult
	require.NotPanics(t, func() {
		mustStart = withHold.MustStart(probe, 3, rounding.HalfUp)
		mustEnd = withHold.MustEnd(probe, 3, rounding.HalfUp)
	})
	assertSecondsResult(t, "s~0/1@3:1/2000", mustStart)
	assertSecondsResult(t, "s~1/1@3:1/2000", mustEnd)

	var speed fraction.Fraction
	require.NotPanics(t, func() {
		speed, err = withHold.Speed(probe)
	})
	require.Error(t, err)
	assert.Equal(t, fraction.Fraction{}, speed)
	assert.ErrorIs(t, err, tempomap.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrOutOfRange)
	assertSecondsResult(t, "s~0/1@3:1/2000", start)
	assertMapSnapshot(t, baseSnapshot, base)
	assertMapSnapshot(t, holdSnapshot, withHold)
}

func TestEqualTimelineUnrepresentableMidpoint(t *testing.T) {
	t.Parallel()

	redundant := "tempomap[first@p:0/1:main:4=60;single@p:1/" + fraction.MaxComponentBig().String() + ":main:4=60]"
	equivalent := "tempomap[first@p:0/1:main:4=60]"
	tests := []struct {
		name        string
		left        string
		right       string
		recordEqual bool
	}{
		{name: "self", left: redundant, right: redundant, recordEqual: true},
		{name: "equivalent_redundant_change", left: redundant, right: equivalent, recordEqual: false},
		{name: "equivalent_redundant_change_reverse", left: equivalent, right: redundant, recordEqual: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			left := parseRegressionMap(t, tt.left)
			right := parseRegressionMap(t, tt.right)
			leftSnapshot := left.String()
			rightSnapshot := right.String()
			var equal bool
			var err error
			require.NotPanics(t, func() {
				equal, err = left.EqualTimeline(right)
			})
			require.NoError(t, err)
			assert.True(t, equal)

			var mustEqual bool
			require.NotPanics(t, func() {
				mustEqual = left.MustEqualTimeline(right)
			})
			assert.True(t, mustEqual)
			assert.Equal(t, tt.recordEqual, left.Equal(right))
			assertMapSnapshot(t, leftSnapshot, left)
			assertMapSnapshot(t, rightSnapshot, right)
		})
	}
}

func TestEqualTimelineUnrepresentableProbeSpeed(t *testing.T) {
	t.Parallel()

	maximum := fraction.MaxComponentBig().String()
	previous := new(big.Int).Sub(fraction.MaxComponentBig(), big.NewInt(1)).String()
	m := parseRegressionMap(t, "tempomap[first@p:0/1:main:1="+maximum+";second@p:0/1:gradual:linear-speed;single@p:1/1:main:1="+previous+"]")
	snapshot := m.String()
	probe := regressionAddress("1/2")

	var equal bool
	var err error
	require.NotPanics(t, func() {
		equal, err = m.EqualTimeline(m)
	})
	require.NoError(t, err)
	assert.True(t, equal)
	require.NotPanics(t, func() {
		equal = m.MustEqualTimeline(m)
	})
	assert.True(t, equal)

	var speed fraction.Fraction
	require.NotPanics(t, func() {
		speed, err = m.Speed(probe)
	})
	require.Error(t, err)
	assert.Equal(t, fraction.Fraction{}, speed)
	assert.ErrorIs(t, err, tempomap.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrOutOfRange)
	assertMustPanic(t, err, func() {
		m.MustSpeed(probe)
	})
	assertMapSnapshot(t, snapshot, m)
}

func TestEqualTimelineLastOpenInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		boundary string
	}{
		{
			name:     "boundary_M_minus_1_midpoint_numerator_2M_minus_1",
			boundary: new(big.Int).Sub(fraction.MaxComponentBig(), big.NewInt(1)).String(),
		},
		{
			name:     "boundary_M_second_probe_equals_start",
			boundary: fraction.MaxComponentBig().String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			same := parseRegressionMap(t, "tempomap[first@p:0/1:main:4=60;single@p:"+tt.boundary+"/1:main:4=120]")
			different := parseRegressionMap(t, "tempomap[first@p:0/1:main:4=60;single@p:"+tt.boundary+"/1:main:4=121]")
			sameSnapshot := same.String()
			differentSnapshot := different.String()

			var equal bool
			var err error
			require.NotPanics(t, func() {
				equal, err = same.EqualTimeline(same)
			})
			require.NoError(t, err)
			assert.True(t, equal)
			require.NotPanics(t, func() {
				equal = same.MustEqualTimeline(same)
			})
			assert.True(t, equal)

			require.NotPanics(t, func() {
				equal, err = same.EqualTimeline(different)
			})
			require.NoError(t, err)
			assert.False(t, equal)
			require.NotPanics(t, func() {
				equal = same.MustEqualTimeline(different)
			})
			assert.False(t, equal)
			assertMapSnapshot(t, sameSnapshot, same)
			assertMapSnapshot(t, differentSnapshot, different)
		})
	}
}

func TestSegmentSecondsRangeErrorPresence(t *testing.T) {
	t.Parallel()

	m := parseRegressionMap(t, gradualBoundaryMapText)
	snapshot := m.String()
	var segments []tempomap.Segment
	var err error
	require.NotPanics(t, func() {
		segments, err = m.Segments()
	})
	require.NoError(t, err)
	require.Len(t, segments, 2)
	bounded := segments[0]
	open := segments[1]

	var actual tempomap.SecondsResult
	var ok bool
	require.NotPanics(t, func() {
		actual, ok, err = bounded.Seconds(77, rounding.HalfUp)
	})
	require.Error(t, err)
	assert.Equal(t, tempomap.SecondsResult{}, actual)
	assert.False(t, ok)
	assert.ErrorIs(t, err, tempomap.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrOutOfRange)
	assertMustPanic(t, err, func() {
		bounded.MustSeconds(77, rounding.HalfUp)
	})

	require.NotPanics(t, func() {
		actual, ok, err = bounded.Seconds(3, rounding.HalfUp)
	})
	require.NoError(t, err)
	assert.True(t, ok)
	assertSecondsResult(t, "s~2773/1000@3:1/2000", actual)

	var mustActual tempomap.SecondsResult
	var mustOK bool
	require.NotPanics(t, func() {
		mustActual, mustOK = bounded.MustSeconds(3, rounding.HalfUp)
	})
	assert.True(t, mustOK)
	assertSecondsResult(t, "s~2773/1000@3:1/2000", mustActual)

	require.NotPanics(t, func() {
		actual, ok, err = open.Seconds(3, rounding.HalfUp)
	})
	require.NoError(t, err)
	assert.Equal(t, tempomap.SecondsResult{}, actual)
	assert.False(t, ok)
	require.NotPanics(t, func() {
		mustActual, mustOK = open.MustSeconds(3, rounding.HalfUp)
	})
	assert.Equal(t, tempomap.SecondsResult{}, mustActual)
	assert.False(t, mustOK)
	assertMapSnapshot(t, snapshot, m)
}
