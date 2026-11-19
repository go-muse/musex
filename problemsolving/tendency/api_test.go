package tendency_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func arguments(values ...any) []reflect.Value {
	out := make([]reflect.Value, 0, len(values))
	for _, v := range values {
		out = append(out, reflect.ValueOf(v))
	}
	return out
}

func TestEveryFallibleOperationAndMustWrapper(t *testing.T) {
	t.Parallel()
	m := ionian()
	support := built(tendency.AMSupports)
	roles := built(tendency.AMMajorRoles)
	targets := built(tendency.AMNeighborTargets)
	classification := built(tendency.AMConsonance)
	modal := built(tendency.AMModalResolution)
	inter := built(tendency.AMIntervalResolution)
	schemes := []*tendency.Scheme{support}
	nilSchemes := []*tendency.Scheme{nil}
	moves := tendency.Motions{First: iv("M2").MustDown(), Second: iv("P1")}
	cases := []struct {
		name           string
		normal, must   any
		valid, invalid []reflect.Value
		cause          error
	}{
		{"Builtin", tendency.Builtin, tendency.MustBuiltin, arguments(tendency.AMSupports), arguments(tendency.ID("unknown")), tendency.ErrUnknownScheme},
		{"Catalog", tendency.Catalog, tendency.MustCatalog, arguments(tendency.CatalogQuery{}), arguments(tendency.CatalogQuery{Task: tendency.Task(200)}), tendency.ErrInvalidArgument},
		{"New", tendency.New, tendency.MustNew, arguments(finiteDefinition()), arguments(tendency.Definition{}), tendency.ErrInvalidDefinition},
		{"ValidatePlace", tendency.ValidatePlace, tendency.MustValidatePlace, arguments(m, place(1, 0)), arguments(m, place(0, 0)), tendency.ErrInvalidPlace},
		{"Between", tendency.Between, tendency.MustBetween, arguments(m, place(1, 0), place(2, 0)), arguments(m, place(1, 0), place(8, 0)), tendency.ErrInvalidPlace},
		{"Applicability", support.Applicability, support.MustApplicability, arguments(m, tendency.SupportTask), arguments(m, tendency.TaskUnset), tendency.ErrInvalidArgument},
		{"Support", support.Support, support.MustSupport, arguments(m, 1), arguments(m, 0), tendency.ErrInvalidPlace},
		{"Supports", support.Supports, support.MustSupports, arguments(m), arguments(mode.Mode{}), tendency.ErrInvalidMode},
		{"CompareSupport", support.CompareSupport, support.MustCompareSupport, arguments(m, 1, 5), arguments(m, 1, 0), tendency.ErrInvalidPlace},
		{"Role", roles.Role, roles.MustRole, arguments(m, 1), arguments(m, 0), tendency.ErrInvalidPlace},
		{"Roles", roles.Roles, roles.MustRoles, arguments(m), arguments(mode.Mode{}), tendency.ErrInvalidMode},
		{"Tendency", targets.Tendency, targets.MustTendency, arguments(m, place(2, 0)), arguments(m, place(0, 0)), tendency.ErrInvalidPlace},
		{"Tendencies", targets.Tendencies, targets.MustTendencies, arguments(m, 0), arguments(mode.Mode{}, 0), tendency.ErrInvalidMode},
		{"Classify", classification.Classify, classification.MustClassify, arguments(iv("P4")), arguments(interval.Interval{}), tendency.ErrInvalidInterval},
		{"ResolveModal", modal.ResolveModal, modal.MustResolveModal, arguments(m, pair(2, 4)), arguments(m, pair(0, 4)), tendency.ErrInvalidPlace},
		{"ResolveInterval", inter.ResolveInterval, inter.MustResolveInterval, arguments(iv("m2")), arguments(interval.Interval{}), tendency.ErrInvalidInterval},
		{"CheckTransition", targets.CheckTransition, targets.MustCheckTransition, arguments(m, place(2, 0), place(1, 0)), arguments(m, place(2, 0), place(0, 0)), tendency.ErrInvalidPlace},
		{"CheckModal", modal.CheckModal, modal.MustCheckModal, arguments(m, pair(2, 4), pair(1, 3)), arguments(m, pair(2, 4), pair(1, 0)), tendency.ErrInvalidPlace},
		{"CheckInterval", inter.CheckInterval, inter.MustCheckInterval, arguments(iv("m2"), moves), arguments(iv("m2"), tendency.Motions{}), tendency.ErrInvalidInterval},
		{"SelectGoals", targets.SelectGoals, targets.MustSelectGoals, arguments(m, place(2, 0), tendency.GoalFilter{}), arguments(m, place(1, 0), tendency.GoalFilter{Directions: []tendency.Motion{tendency.MotionUnset}}), tendency.ErrInvalidArgument},
		{"SelectModal", modal.SelectModal, modal.MustSelectModal, arguments(m, pair(2, 4), tendency.ModalFilter{}), arguments(m, pair(1, 3), tendency.ModalFilter{Keep: tendency.Keep(200)}), tendency.ErrInvalidArgument},
		{"SelectInterval", inter.SelectInterval, inter.MustSelectInterval, arguments(iv("m2"), tendency.IntervalFilter{}), arguments(iv("P1"), tendency.IntervalFilter{SecondDirections: []tendency.Motion{tendency.Motion(200)}}), tendency.ErrInvalidArgument},
		{"CompareSupports", tendency.CompareSupports, tendency.MustCompareSupports, arguments(m, 1, schemes), arguments(m, 1, nilSchemes), tendency.ErrUnsetScheme},
		{"CompareRoles", tendency.CompareRoles, tendency.MustCompareRoles, arguments(m, 1, schemes), arguments(m, 1, nilSchemes), tendency.ErrUnsetScheme},
		{"CompareTendencies", tendency.CompareTendencies, tendency.MustCompareTendencies, arguments(m, place(2, 0), []*tendency.Scheme{targets}, tendency.ExactGoal), arguments(m, place(2, 0), nilSchemes, tendency.ExactGoal), tendency.ErrUnsetScheme},
		{"CompareClassifications", tendency.CompareClassifications, tendency.MustCompareClassifications, arguments(iv("P4"), []*tendency.Scheme{classification}), arguments(interval.Interval{}, schemes), tendency.ErrInvalidInterval},
		{"CompareModalResolutions", tendency.CompareModalResolutions, tendency.MustCompareModalResolutions, arguments(m, pair(2, 4), []*tendency.Scheme{modal}), arguments(m, pair(2, 4), nilSchemes), tendency.ErrUnsetScheme},
		{"CompareIntervalResolutions", tendency.CompareIntervalResolutions, tendency.MustCompareIntervalResolutions, arguments(iv("m2"), []*tendency.Scheme{inter}), arguments(iv("m2"), nilSchemes), tendency.ErrUnsetScheme},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			normal := reflect.ValueOf(tc.normal)
			must := reflect.ValueOf(tc.must)
			good := normal.Call(tc.valid)
			require.True(t, good[len(good)-1].IsNil())
			actual := must.Call(tc.valid)
			require.Len(t, actual, len(good)-1)
			for i, value := range actual {
				assert.Equal(t, good[i].Interface(), value.Interface())
			}
			bad := normal.Call(tc.invalid)
			err, ok := bad[len(bad)-1].Interface().(error)
			require.True(t, ok)
			assert.ErrorIs(t, err, tc.cause)
			assert.NotEmpty(t, err.Error())
			for _, value := range bad[:len(bad)-1] {
				assert.True(t, value.IsZero(), "all error results must be zero")
			}
			var recovered any
			func() { defer func() { recovered = recover() }(); must.Call(tc.invalid) }()
			panicError, ok := recovered.(error)
			require.True(t, ok, "Must must panic with an error, not a string")
			assert.True(t, errors.Is(panicError, tc.cause))
			assert.Equal(t, err.Error(), panicError.Error())
		})
	}
}

func TestAbsentSchemeAndValidationOrder(t *testing.T) {
	t.Parallel()
	for _, s := range []*tendency.Scheme{nil, new(tendency.Scheme)} {
		assert.False(t, s.IsValid())
		assert.Equal(t, tendency.Info{}, s.Info())
		def, ok := s.Definition()
		assert.False(t, ok)
		assert.Equal(t, tendency.Definition{}, def)
		assert.True(t, s.Equal(nil))
		assert.True(t, s.Equal(new(tendency.Scheme)))
		calls := []struct {
			fn   any
			args []reflect.Value
		}{
			{s.Applicability, arguments(mode.Mode{}, tendency.TaskUnset)},
			{s.Support, arguments(mode.Mode{}, 0)}, {s.Supports, arguments(mode.Mode{})}, {s.CompareSupport, arguments(mode.Mode{}, 0, 0)},
			{s.Role, arguments(mode.Mode{}, 0)}, {s.Roles, arguments(mode.Mode{})},
			{s.Tendency, arguments(mode.Mode{}, tendency.Place{})}, {s.Tendencies, arguments(mode.Mode{}, 0)},
			{s.Classify, arguments(interval.Interval{})}, {s.ResolveModal, arguments(mode.Mode{}, tendency.Pair{})}, {s.ResolveInterval, arguments(interval.Interval{})},
			{s.CheckTransition, arguments(mode.Mode{}, tendency.Place{}, tendency.Place{})}, {s.CheckModal, arguments(mode.Mode{}, tendency.Pair{}, tendency.Pair{})}, {s.CheckInterval, arguments(interval.Interval{}, tendency.Motions{})},
			{s.SelectGoals, arguments(mode.Mode{}, tendency.Place{}, tendency.GoalFilter{})}, {s.SelectModal, arguments(mode.Mode{}, tendency.Pair{}, tendency.ModalFilter{})}, {s.SelectInterval, arguments(interval.Interval{}, tendency.IntervalFilter{})},
		}
		for _, call := range calls {
			results := reflect.ValueOf(call.fn).Call(call.args)
			err, ok := results[len(results)-1].Interface().(error)
			require.True(t, ok)
			assert.ErrorIs(t, err, tendency.ErrUnsetScheme)
			for _, r := range results[:len(results)-1] {
				assert.True(t, r.IsZero())
			}
		}
	}
	s := built(tendency.AMNeighborTargets)
	_, err := s.SelectGoals(mode.Mode{}, place(0, 0), tendency.GoalFilter{Directions: []tendency.Motion{200}})
	assert.ErrorIs(t, err, tendency.ErrInvalidMode)
	_, err = s.SelectGoals(ionian(), place(0, 0), tendency.GoalFilter{Directions: []tendency.Motion{200}})
	assert.ErrorIs(t, err, tendency.ErrInvalidPlace)
	_, err = s.SelectGoals(ionian(), place(1, 0), tendency.GoalFilter{Support: new(tendency.Scheme)})
	assert.ErrorIs(t, err, tendency.ErrUnsetScheme)
	_, err = s.SelectGoals(ionian(), place(1, 0), tendency.GoalFilter{Targets: []tendency.Place{place(8, 0)}})
	assert.ErrorIs(t, err, tendency.ErrInvalidPlace)
	_, err = tendency.CompareRoles(ionian(), 1, []*tendency.Scheme{built(tendency.AMMajorRoles), built(tendency.AMMajorRoles)})
	assert.ErrorIs(t, err, tendency.ErrInvalidArgument)
}

func TestConcurrentImmutableSchemeReads(t *testing.T) {
	t.Parallel()
	s := tendency.MustNew(finiteDefinition())
	m := ionian()
	want := s.MustTendency(m, place(2, 0))
	const workers = 16
	results := make(chan tendency.TendencyResult, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			r, err := s.Tendency(m, place(2, 0))
			results <- r
			errs <- err
			info := s.Info()
			info.Tasks[0] = tendency.TaskUnset
			def, _ := s.Definition()
			def.Tendencies[0].Goals[0].Target.Degree = 1
		})
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	for r := range results {
		assert.Equal(t, want, r)
	}
	assert.Equal(t, want, s.MustTendency(m, place(2, 0)))
}
