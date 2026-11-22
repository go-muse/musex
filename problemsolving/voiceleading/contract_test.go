package voiceleading_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeValidationAndIdentity(t *testing.T) {
	t.Parallel()
	empty := voice.MustNew()
	cases := []struct {
		name  string
		lines []*voice.Line
		opts  voiceleading.Options
		cause error
		part  int
	}{
		{"nil input", nil, voiceleading.Options{}, voiceleading.ErrPartCount, -1},
		{"empty input", []*voice.Line{}, voiceleading.Options{}, voiceleading.ErrPartCount, -1},
		{"one input before unset validation", []*voice.Line{nil}, voiceleading.Options{Order: 255}, voiceleading.ErrPartCount, -1},
		{"nil line", []*voice.Line{empty, nil}, voiceleading.Options{}, voiceleading.ErrUnset, 1},
		{"zero line before options", []*voice.Line{new(voice.Line), nil}, voiceleading.Options{Order: 255}, voiceleading.ErrUnset, 0},
		{"invalid order before system", []*voice.Line{empty, empty}, voiceleading.Options{Order: 255, System: new(system.System)}, voiceleading.ErrOption, -1},
		{"unset explicit system even on empty material", []*voice.Line{empty, empty}, voiceleading.Options{System: new(system.System)}, voiceleading.ErrUnset, -1},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, err := voiceleading.Analyze(tt.lines, tt.opts)
			assert.Nil(t, a)
			require.ErrorIs(t, err, tt.cause)
			assert.Contains(t, err.Error(), "voiceleading: ")
			if tt.part >= 0 {
				assertPartError(t, err, voice.ErrUnset, tt.part, -1)
			}
			if tt.opts.System != nil && tt.opts.Order == voiceleading.Unordered {
				assert.ErrorIs(t, err, system.ErrInvalid)
			}
			assertOnlySentinel(t, err, tt.cause)
			assertPanicCause(t, tt.cause, func() { voiceleading.MustAnalyze(tt.lines, tt.opts) })
		})
	}
	shared := line(t, "C4").MustWithName("same").MustWithPlacement(voice.Soprano)
	sameName := line(t, "G4").MustWithName("same").MustWithPlacement(voice.Bass)
	a := voiceleading.MustAnalyze([]*voice.Line{shared, shared, sameName, empty}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	assert.True(t, a.IsSet())
	assert.Equal(t, 4, a.MustPartCount())
	assert.Equal(t, voiceleading.LowToHigh, a.MustOrder())
	assert.Equal(t, "same", a.MustName(0))
	assert.Equal(t, "same", a.MustName(1))
	assert.Equal(t, "same", a.MustName(2))
	assert.Empty(t, a.MustName(3))
	assert.Equal(t, []voiceleading.Pair{{First: 0, Second: 1}, {First: 0, Second: 2}, {First: 0, Second: 3}, {First: 1, Second: 2}, {First: 1, Second: 3}, {First: 2, Second: 3}}, a.MustPairs())
	assert.Empty(t, a.MustVerticals(voiceleading.Pair{First: 2, Second: 3}))
	assert.Equal(t, system.Step(7), a.MustVerticals(voiceleading.Pair{First: 0, Second: 2})[0].Chromatic) // Placement does not reorder parts.
	key := map[voiceleading.Pair]string{pair: "pair"}
	assert.Equal(t, "pair", key[voiceleading.Pair{First: 0, Second: 1}])
	assert.False(t, reflect.TypeFor[voiceleading.Analysis]().Comparable())
}

func assertOnlySentinel(t *testing.T, err, want error) {
	t.Helper()
	for _, sentinel := range []error{voiceleading.ErrUnset, voiceleading.ErrPartCount, voiceleading.ErrPart, voiceleading.ErrOption, voiceleading.ErrNoOrder, voiceleading.ErrNoSystem, voiceleading.ErrDifferentSystems, voiceleading.ErrTransition} {
		assert.Equal(t, sentinel == want, errors.Is(err, sentinel), "sentinel %v", sentinel)
	}
}

func assertPanicCause(t *testing.T, want error, call func()) {
	t.Helper()
	var recovered any
	func() { defer func() { recovered = recover() }(); call() }()
	err, ok := recovered.(error)
	require.True(t, ok, "Must must panic with an error, got %T", recovered)
	assert.ErrorIs(t, err, want)
}

type operation struct {
	name         string
	normal, must any
	args         []any
}

func operations(a *voiceleading.Analysis) []operation {
	return []operation{
		{"PartCount", a.PartCount, a.MustPartCount, nil}, {"Name", a.Name, a.MustName, []any{0}},
		{"Order", a.Order, a.MustOrder, nil}, {"System", a.System, a.MustSystem, nil}, {"Pairs", a.Pairs, a.MustPairs, nil},
		{"Verticals", a.Verticals, a.MustVerticals, []any{pair}}, {"VerticalAt", a.VerticalAt, a.MustVerticalAt, []any{pair, pos(0, 1)}},
		{"Transitions", a.Transitions, a.MustTransitions, []any{pair}}, {"TransitionAt", a.TransitionAt, a.MustTransitionAt, []any{pair, pos(0, 1), pos(1, 4)}},
		{"FindMotion", a.FindMotion, a.MustFindMotion, []any{voiceleading.MotionParallel}},
		{"FindParallels", a.FindParallels, a.MustFindParallels, []any{voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}}},
		{"FindConsecutive", a.FindConsecutive, a.MustFindConsecutive, []any{voiceleading.PerfectBoth}},
		{"FindDirect", a.FindDirect, a.MustFindDirect, []any{voiceleading.PerfectBoth, voiceleading.AnyOrigin}},
		{"FindCrossings", a.FindCrossings, a.MustFindCrossings, nil}, {"FindOverlaps", a.FindOverlaps, a.MustFindOverlaps, nil},
		{"CheckOMTFirstSpecies", a.CheckOMTFirstSpecies, a.MustCheckOMTFirstSpecies, []any{pair}},
	}
}
func call(fn any, args []any) []reflect.Value {
	values := make([]reflect.Value, len(args))
	for i, v := range args {
		values[i] = reflect.ValueOf(v)
	}
	return reflect.ValueOf(fn).Call(values)
}

func TestEveryOperationOnUnsetAndEveryMust(t *testing.T) {
	t.Parallel()
	for _, a := range []*voiceleading.Analysis{nil, new(voiceleading.Analysis)} {
		assert.False(t, a.IsSet())
		for _, op := range operations(a) {
			t.Run(op.name, func(t *testing.T) {
				t.Parallel()
				results := call(op.normal, op.args)
				err, ok := results[len(results)-1].Interface().(error)
				require.True(t, ok)
				assertOnlySentinel(t, err, voiceleading.ErrUnset)
				for _, v := range results[:len(results)-1] {
					assert.True(t, v.IsZero(), "nonzero result: %v", v.Interface())
				}
				assertPanicCause(t, voiceleading.ErrUnset, func() { call(op.must, op.args) })
			})
		}
	}
	a := analyze(t, []string{"C4", "D4"}, []string{"G4", "A4"})
	for _, op := range operations(a) {
		t.Run("success "+op.name, func(t *testing.T) {
			t.Parallel()
			want := call(op.normal, op.args)
			require.True(t, want[len(want)-1].IsNil())
			got := call(op.must, op.args)
			require.Len(t, got, len(want)-1)
			for i, v := range got {
				assertResult(t, want[i].Interface(), v.Interface())
			}
		})
	}
	b := voiceleading.MustAnalyze([]*voice.Line{line(t, "C4", "D4"), line(t, "G4", "A4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	assert.True(t, a.Equal(b))
}

func TestInvalidArgumentsAndZeroResults(t *testing.T) {
	t.Parallel()
	a := analyze(t, []string{"C4", "D4"}, []string{"G4", "A4"})
	cases := []struct {
		name              string
		fn                any
		args              []any
		cause, dependency error
	}{
		{"negative name", a.Name, []any{-1}, voiceleading.ErrPart, nil}, {"past last name", a.Name, []any{2}, voiceleading.ErrPart, nil},
		{"unset at", a.VerticalAt, []any{pair, duration.Position{}}, voiceleading.ErrUnset, duration.ErrUnset},
		{"unset from before unset to", a.TransitionAt, []any{pair, duration.Position{}, duration.Position{}}, voiceleading.ErrUnset, duration.ErrUnset},
		{"unset to", a.TransitionAt, []any{pair, pos(0, 1), duration.Position{}}, voiceleading.ErrUnset, duration.ErrUnset},
		{"unknown motion", a.FindMotion, []any{voiceleading.Motion(255)}, voiceleading.ErrOption, nil},
		{"unclassified is not a search", a.FindMotion, []any{voiceleading.MotionUnclassified}, voiceleading.ErrOption, nil},
		{"empty perfect", a.FindConsecutive, []any{voiceleading.PerfectSet(0)}, voiceleading.ErrOption, nil},
		{"unknown perfect bits", a.FindConsecutive, []any{voiceleading.PerfectSet(4)}, voiceleading.ErrOption, nil},
		{"parallel empty options", a.FindParallels, []any{voiceleading.ParallelOptions{}}, voiceleading.ErrOption, nil},
		{"parallel unknown bits", a.FindParallels, []any{voiceleading.ParallelOptions{Perfect: 255}}, voiceleading.ErrOption, nil},
		{"direct unset origin", a.FindDirect, []any{voiceleading.PerfectBoth, voiceleading.OriginUnset}, voiceleading.ErrOption, nil},
		{"direct unknown origin", a.FindDirect, []any{voiceleading.PerfectBoth, voiceleading.DirectOrigin(255)}, voiceleading.ErrOption, nil},
		{"direct invalid perfect", a.FindDirect, []any{voiceleading.PerfectSet(0), voiceleading.AnyOrigin}, voiceleading.ErrOption, nil},
	}
	invalidPairs := []voiceleading.Pair{{}, {First: 1, Second: 0}, {First: 1, Second: 1}, {First: -1, Second: 1}, {First: 0, Second: 2}}
	for _, p := range invalidPairs {
		for _, op := range []struct {
			name string
			fn   any
			args []any
		}{{"Verticals", a.Verticals, []any{p}}, {"VerticalAt", a.VerticalAt, []any{p, duration.Position{}}}, {"Transitions", a.Transitions, []any{p}}, {"TransitionAt", a.TransitionAt, []any{p, duration.Position{}, duration.Position{}}}, {"OMT", a.CheckOMTFirstSpecies, []any{p}}} {
			cases = append(cases, struct {
				name              string
				fn                any
				args              []any
				cause, dependency error
			}{op.name, op.fn, op.args, voiceleading.ErrPart, nil})
		}
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			results := call(tt.fn, tt.args)
			err, ok := results[len(results)-1].Interface().(error)
			require.True(t, ok)
			assert.ErrorIs(t, err, tt.cause)
			assertOnlySentinel(t, err, tt.cause)
			if tt.dependency != nil {
				assert.ErrorIs(t, err, tt.dependency)
			}
			for _, v := range results[:len(results)-1] {
				assert.True(t, v.IsZero())
			}
		})
	}
	for _, lines := range [][]*voice.Line{{line(t), line(t)}, {line(t, "F4"), line(t, "D4")}} {
		unordered := voiceleading.MustAnalyze(lines, voiceleading.Options{System: system.Traditional()})
		assert.Equal(t, voiceleading.Unordered, unordered.MustOrder())
		xs, err := unordered.FindCrossings()
		assert.Nil(t, xs)
		assert.ErrorIs(t, err, voiceleading.ErrNoOrder)
		os, err := unordered.FindOverlaps()
		assert.Nil(t, os)
		assert.ErrorIs(t, err, voiceleading.ErrNoOrder)
		assertPanicCause(t, voiceleading.ErrNoOrder, func() { unordered.MustFindCrossings() })
		assertPanicCause(t, voiceleading.ErrNoOrder, func() { unordered.MustFindOverlaps() })
	}
}

func TestSuccessfulEmptyCollections(t *testing.T) {
	t.Parallel()
	a := analyze(t, nil, nil)
	for _, op := range operations(a) {
		if op.name == "TransitionAt" {
			continue
		}
		results := call(op.normal, op.args)
		require.True(t, results[len(results)-1].IsNil())
		for _, v := range results[:len(results)-1] {
			if v.Kind() == reflect.Slice {
				assert.False(t, v.IsNil(), op.name)
			}
		}
	}
	v, ok, err := a.VerticalAt(pair, pos(0, 1))
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, v)
}

func TestMotionLabelsAndErrorContainers(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		motion voiceleading.Motion
		label  string
		direct bool
	}{{voiceleading.MotionUnclassified, "unclassified", false}, {voiceleading.MotionStationary, "stationary", false}, {voiceleading.MotionOblique, "oblique", false}, {voiceleading.MotionContrary, "contrary", false}, {voiceleading.MotionDirect, "direct", true}, {voiceleading.MotionParallel, "parallel", true}, {255, "unknown", false}} {
		assert.Equal(t, tt.label, tt.motion.String())
		assert.Equal(t, tt.direct, tt.motion.IsDirect())
	}
	cause := errors.New("test cause")
	for _, e := range []*voiceleading.PartError{nil, {}, {Part: 2, Element: 3, Err: cause}} {
		assert.Contains(t, e.Error(), "voiceleading: ")
		if e == nil {
			assert.Nil(t, e.Unwrap())
		} else {
			assert.Equal(t, e.Err, e.Unwrap())
		}
	}
	for _, e := range []*voiceleading.MeasurementError{nil, {}, {Pair: pair, At: pos(1, 4), Err: cause}} {
		assert.Contains(t, e.Error(), "voiceleading: ")
		if e == nil {
			assert.Nil(t, e.Unwrap())
		} else {
			assert.Equal(t, e.Err, e.Unwrap())
		}
	}
}

func TestUnsetReceiverPrecedesBadArguments(t *testing.T) {
	t.Parallel()
	var a *voiceleading.Analysis
	calls := []struct {
		fn   any
		args []any
	}{
		{a.Name, []any{-1}}, {a.VerticalAt, []any{voiceleading.Pair{}, duration.Position{}}},
		{a.TransitionAt, []any{voiceleading.Pair{}, duration.Position{}, duration.Position{}}},
		{a.FindMotion, []any{voiceleading.Motion(255)}},
		{a.FindDirect, []any{voiceleading.PerfectSet(0), voiceleading.OriginUnset}},
		{a.FindParallels, []any{voiceleading.ParallelOptions{}}},
		{a.CheckOMTFirstSpecies, []any{voiceleading.Pair{}}},
	}
	for _, tt := range calls {
		results := call(tt.fn, tt.args)
		err, ok := results[len(results)-1].Interface().(error)
		require.True(t, ok)
		assertOnlySentinel(t, err, voiceleading.ErrUnset)
	}
}
