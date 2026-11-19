package modal_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type query struct {
	name      string
	absolute  bool
	addressed bool
	compare   bool
	zero      any
	run       func(mode.Mode, int, int) (any, error)
	must      func(mode.Mode, int, int) any
}

// Shared only to enumerate public entry points and their zero results; no
// modal calculation or classification is implemented by this test helper.
func queries() []query {
	return []query{
		{name: "Relative", addressed: true, zero: modal.RelativePosition{},
			run:  func(m mode.Mode, a, _ int) (any, error) { return modal.Relative(m, a) },
			must: func(m mode.Mode, a, _ int) any { return modal.MustRelative(m, a) }},
		{name: "RelativePositions", zero: []modal.RelativePosition(nil),
			run:  func(m mode.Mode, _, _ int) (any, error) { return modal.RelativePositions(m) },
			must: func(m mode.Mode, _, _ int) any { return modal.MustRelativePositions(m) }},
		{name: "GroupRelative", zero: modal.RelativeGroups{},
			run:  func(m mode.Mode, _, _ int) (any, error) { return modal.GroupRelative(m) },
			must: func(m mode.Mode, _, _ int) any { return modal.MustGroupRelative(m) }},
		{name: "Absolute", absolute: true, addressed: true, zero: modal.AbsolutePosition{},
			run:  func(m mode.Mode, a, _ int) (any, error) { return modal.Absolute(m, a) },
			must: func(m mode.Mode, a, _ int) any { return modal.MustAbsolute(m, a) }},
		{name: "AbsolutePositions", absolute: true, zero: []modal.AbsolutePosition(nil),
			run:  func(m mode.Mode, _, _ int) (any, error) { return modal.AbsolutePositions(m) },
			must: func(m mode.Mode, _, _ int) any { return modal.MustAbsolutePositions(m) }},
		{name: "Order", absolute: true, zero: []modal.AbsoluteGroup(nil),
			run:  func(m mode.Mode, _, _ int) (any, error) { return modal.Order(m) },
			must: func(m mode.Mode, _, _ int) any { return modal.MustOrder(m) }},
		{name: "CompareAbsolute", absolute: true, addressed: true, compare: true, zero: 0,
			run:  func(m mode.Mode, a, b int) (any, error) { return modal.CompareAbsolute(m, a, b) },
			must: func(m mode.Mode, a, b int) any { return modal.MustCompareAbsolute(m, a, b) }},
		{name: "AnalyzeAbsolute", absolute: true, zero: modal.AbsoluteAnalysis{},
			run:  func(m mode.Mode, _, _ int) (any, error) { return modal.AnalyzeAbsolute(m) },
			must: func(m mode.Mode, _, _ int) any { return modal.MustAnalyzeAbsolute(m) }},
	}
}

func checkPanicIs(t *testing.T, action func(), causes ...error) {
	t.Helper()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		action()
	}()
	require.NotNil(t, recovered, "Must must panic on an error")
	err, ok := recovered.(error)
	require.True(t, ok, "panic must contain error, got %T", recovered)
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
}

func checkFailure(t *testing.T, q query, m mode.Mode, a, b int, causes ...error) {
	t.Helper()
	got, err := q.run(m, a, b)
	require.Error(t, err)
	assert.Equal(t, q.zero, got, "no partial result on failure")
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
	for _, category := range []error{modal.ErrUnset, modal.ErrInvalidDegree, modal.ErrUnsupportedStructure, modal.ErrUnsupportedQuality} {
		if !slices.Contains(causes, category) {
			assert.NotErrorIs(t, err, category, "only the highest-priority error category applies")
		}
	}
	checkPanicIs(t, func() { q.must(m, a, b) }, causes...)
}

func TestSentinelsAreDistinct(t *testing.T) {
	t.Parallel()
	values := []error{modal.ErrUnset, modal.ErrInvalidDegree, modal.ErrUnsupportedStructure, modal.ErrUnsupportedQuality}
	for i, err := range values {
		require.NotNil(t, err)
		for j, other := range values {
			if i != j {
				assert.NotErrorIs(t, err, other)
			}
		}
	}
}

func TestUnsetHasPriorityForEveryQuery(t *testing.T) {
	t.Parallel()
	for _, q := range queries() {
		t.Run(q.name, func(t *testing.T) {
			t.Parallel()
			for _, number := range []int{1, 0, -1, math.MinInt, math.MaxInt} {
				checkFailure(t, q, mode.Mode{}, number, number, modal.ErrUnset, mode.ErrUnset)
				_, err := q.run(mode.Mode{}, number, number)
				assert.NotErrorIs(t, err, modal.ErrInvalidDegree)
			}
		})
	}
}

func TestInvalidDegreesPrecedeStructureAndIntervals(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		m    func(*testing.T) mode.Mode
	}{
		{"supported", dorianMode},
		{"open", func(t *testing.T) mode.Mode { return openMode(t, "P1 M2 n3") }},
		{"non-octave", func(t *testing.T) mode.Mode { return closedMode(t, "P1 n2", "P4") }},
		{"unsupported quality", func(t *testing.T) mode.Mode { return closedMode(t, "P1 n2", "P8") }},
		{"unrepresentable table", func(t *testing.T) mode.Mode { return rangeFailureMode(t, false) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := tc.m(t)
			for _, q := range queries() {
				if !q.addressed {
					continue
				}
				t.Run(q.name, func(t *testing.T) {
					for _, number := range []int{0, -1, m.Len() + 1, math.MinInt, math.MaxInt} {
						t.Run(fmt.Sprint(number), func(t *testing.T) {
							checkFailure(t, q, m, number, 1, modal.ErrInvalidDegree)
							if q.compare {
								checkFailure(t, q, m, 1, number, modal.ErrInvalidDegree)
								checkFailure(t, q, m, number, math.MaxInt, modal.ErrInvalidDegree)
							}
						})
					}
				})
			}
		})
	}
}

func TestAbsoluteRejectsUnsupportedStructures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, degrees, cycle string }{
		{"open", "P1 M2 M3", ""},
		{"open quality also unsupported", "P1 n2", ""},
		{"fourth", "P1 M2 M3", "P4"},
		{"two octaves", "P1 M9", "P15"},
		{"diminished ninth has traditional octave size", "P1 M2", "d9"},
		{"non-octave with neutral quality", "P1 n2", "P4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var m mode.Mode
			if tc.cycle == "" {
				m = openMode(t, tc.degrees)
			} else {
				m = closedMode(t, tc.degrees, tc.cycle)
			}
			for _, q := range queries() {
				if !q.absolute {
					continue
				}
				t.Run(q.name, func(t *testing.T) {
					checkFailure(t, q, m, 1, 1, modal.ErrUnsupportedStructure)
					_, err := q.run(m, 1, 1)
					assert.NotErrorIs(t, err, modal.ErrUnsupportedQuality)
				})
			}
		})
	}
}

func TestUnsupportedQualitiesAreAtomic(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		value interval.Interval
	}{
		{"neutral", interval.MustNew(interval.Neutral, 3)},
		{"fractional", interval.MustFromDeviation(3, alteration.MustFromFraction(1, 3))},
		{"fractional perfect group", interval.MustFromDeviation(4, alteration.MustFromFraction(-1, 2))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, err := mode.ClosedFromDegreeIntervals([]interval.Interval{interval.PerfectUnison(), interval.MajorSecond(), tc.value}, interval.PerfectOctave())
			require.NoError(t, err)
			// The addressed relative query must not validate another degree.
			p, err := modal.Relative(m, 2)
			require.NoError(t, err)
			assert.Equal(t, modal.RelativePosition{Number: 2, Category: modal.High, Interval: interval.MajorSecond()}, p)
			assert.Equal(t, p, modal.MustRelative(m, 2))
			for _, q := range queries() {
				t.Run(q.name, func(t *testing.T) {
					checkFailure(t, q, m, 3, 3, modal.ErrUnsupportedQuality)
					if q.absolute {
						checkFailure(t, q, m, 1, 1, modal.ErrUnsupportedQuality)
					}
				})
			}
		})
	}
}

// For C, D with alteration -M, E natural (or -1/2), F with alteration +M:
// own deviations are 0,-M,0(or -1/2),M; adjacent deviations are
// -M,M(or M-1/2),M-1(or M-1/2),-M. All are representable and cycle=P8.
// D -> F is a third with deviation 2M-1, beyond alteration.MaxMagnitude.
// It is outside the root's outgoing row and outside the requested self-pair.
func rangeFailureMode(t *testing.T, withNeutral bool) mode.Mode {
	t.Helper()
	middle := interval.MajorThird()
	if withNeutral {
		middle = interval.MustNew(interval.Neutral, 3)
	}
	m, err := mode.ClosedFromDegreeIntervals([]interval.Interval{
		interval.PerfectUnison(),
		interval.MustFromDeviation(2, alteration.MustFromInt(-alteration.MaxMagnitude)),
		middle,
		interval.MustFromDeviation(4, alteration.MustFromInt(alteration.MaxMagnitude)),
	}, interval.PerfectOctave())
	require.NoError(t, err, "fixture must be a valid Mode before querying modal")
	require.True(t, m.IsOctave())
	_, err = m.Intervals()
	require.ErrorIs(t, err, interval.ErrOutOfRange)
	require.ErrorIs(t, err, alteration.ErrOutOfRange)
	return m
}

func TestAbsolutePreservesWholeTableDependencyErrors(t *testing.T) {
	t.Parallel()
	for _, withNeutral := range []bool{false, true} {
		t.Run(fmt.Sprintf("neutral_%t", withNeutral), func(t *testing.T) {
			t.Parallel()
			m := rangeFailureMode(t, withNeutral)
			root, err := modal.Relative(m, 1)
			require.NoError(t, err)
			assert.Equal(t, modal.Neutral, root.Category)
			if !withNeutral {
				_, err = modal.RelativePositions(m)
				require.NoError(t, err, "relative analysis needs only the representable own intervals")
			}
			for _, q := range queries() {
				if !q.absolute {
					continue
				}
				t.Run(q.name, func(t *testing.T) {
					checkFailure(t, q, m, 1, 1, interval.ErrOutOfRange, alteration.ErrOutOfRange)
					checkFailure(t, q, m, 1, 3, interval.ErrOutOfRange, alteration.ErrOutOfRange)
					_, err := q.run(m, 1, 1)
					assert.NotErrorIs(t, err, modal.ErrUnsupportedQuality, "representability precedes quality")
				})
			}
		})
	}
}
