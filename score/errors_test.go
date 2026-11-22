package score_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingCall struct {
	name  string
	args  []any
	cause error
}

// Reflection here verifies the uniform public Must contract without replacing
// behavioral tests. It also requires every promised wrapper to exist.
func checkFailureAndMust(t *testing.T, s *score.Score, tc failingCall) {
	t.Helper()
	args := make([]reflect.Value, len(tc.args))
	for i, a := range tc.args {
		args[i] = reflect.ValueOf(a)
	}
	method := reflect.ValueOf(s).MethodByName(tc.name)
	require.True(t, method.IsValid(), tc.name)
	out := method.Call(args)
	require.NotEmpty(t, out)
	err, ok := out[len(out)-1].Interface().(error)
	require.True(t, ok, "%s must return error", tc.name)
	require.ErrorIs(t, err, tc.cause)
	for _, v := range out[:len(out)-1] {
		assert.True(t, v.IsZero(), "%s returned nonzero failure result: %v", tc.name, v.Interface())
	}
	must := reflect.ValueOf(s).MethodByName("Must" + tc.name)
	require.True(t, must.IsValid(), "Must%s", tc.name)
	var recovered any
	func() { defer func() { recovered = recover() }(); must.Call(args) }()
	panicError, ok := recovered.(error)
	require.True(t, ok, "Must%s must panic with error, got %v", tc.name, recovered)
	assert.ErrorIs(t, panicError, tc.cause)
}

func TestUnsetReceiversAndAllMustFailures(t *testing.T) {
	t.Parallel()
	calls := []failingCall{
		{"Parts", nil, score.ErrUnset}, {"Part", []any{score.PartID{}}, score.ErrUnset},
		{"End", nil, score.ErrUnset}, {"Map", nil, score.ErrUnset}, {"WithMap", []any{tempomap.Map{}}, score.ErrUnset},
		{"Add", []any{(*track.Track)(nil)}, score.ErrUnset}, {"Replace", []any{score.PartID{}, (*track.Track)(nil)}, score.ErrUnset},
		{"Remove", []any{score.PartID{}}, score.ErrUnset}, {"Reorder", nil, score.ErrUnset},
		{"AddWithoutClock", []any{(*track.Track)(nil)}, score.ErrUnset}, {"ReplaceWithoutClock", []any{score.PartID{}, (*track.Track)(nil)}, score.ErrUnset},
		{"ExtractWithMap", []any{score.PartID{}}, score.ErrUnset},
		{"At", []any{duration.Position{}}, score.ErrUnset}, {"Sounds", nil, score.ErrUnset}, {"SoundingAt", []any{duration.Position{}}, score.ErrUnset},
		{"SoundSeconds", []any{-1, rounding.Mode(0)}, score.ErrUnset},
		{"BarView", []any{score.PartID{}, track.VoiceID{}, 0}, score.ErrUnset},
		{"ChangeMeter", []any{0, 0, meter.Meter{}}, score.ErrUnset},
		{"CommonBarlines", []any{duration.Position{}, duration.Position{}}, score.ErrUnset},
		{"BarlinePeriod", []any{score.PartID{}, score.PartID{}}, score.ErrUnset},
	}
	for _, s := range []*score.Score{nil, {}} {
		assert.False(t, s.IsSet())
		assert.True(t, s.Equal(nil))
		assert.True(t, s.Equal(&score.Score{}))
		assert.False(t, s.Equal(score.MustNew(clock60())))
		for _, tc := range calls {
			t.Run(tc.name, func(t *testing.T) { checkFailureAndMust(t, s, tc) })
		}
	}
}

func TestInvalidConstructorAndMap(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		clock tempomap.Map
		cause error
	}{
		{"unset", tempomap.Map{}, tempomap.ErrUnset},
		{"measured", tempomap.MustNewMeasured(grid(4, 4, 1), bpm(60)), score.ErrInvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := score.New(tc.clock)
			require.ErrorIs(t, err, score.ErrInvalidArgument)
			require.ErrorIs(t, err, tc.cause)
			assert.Nil(t, s)
			assert.Panics(t, func() { score.MustNew(tc.clock) })
			checkFailureAndMust(t, score.MustNew(clock60()), failingCall{"WithMap", []any{tc.clock}, score.ErrInvalidArgument})
		})
	}
	for _, p := range []*track.Track{nil, {}} {
		s, id, err := score.MustNew(clock60()).Add(p)
		require.ErrorIs(t, err, score.ErrInvalidArgument)
		require.ErrorIs(t, err, track.ErrUnset)
		assert.Nil(t, s)
		assert.Zero(t, id)
	}
}

func TestKeyAndSelectionValidation(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 1)
	p, v2 := p.MustAddVoice()
	v1 := p.MustVoices()[0]
	s, ids := assemble(clock60(), p, p)
	foreign, other := assemble(clock60(), emptyPart(4, 4, 1))
	_ = foreign
	unknownVoice := emptyPart(4, 4, 1).MustVoices()[0]
	for _, tc := range []failingCall{
		{"Part", []any{other[0]}, score.ErrPart}, {"Remove", []any{score.PartID{}}, score.ErrPart},
		{"Replace", []any{other[0], (*track.Track)(nil)}, score.ErrPart},
		{"ReplaceWithoutClock", []any{other[0], (*track.Track)(nil)}, score.ErrPart},
		{"ExtractWithMap", []any{other[0]}, score.ErrPart},
		{"Reorder", nil, score.ErrInvalidArgument}, {"Reorder", []any{ids[0], ids[0]}, score.ErrInvalidArgument},
		{"Reorder", []any{ids[0], other[0]}, score.ErrPart},
		{"At", []any{duration.Position{}}, score.ErrInvalidArgument},
		{"SoundingAt", []any{pos(2, 1)}, score.ErrPosition},
		{"BarView", []any{ids[0], unknownVoice, 1}, score.ErrVoice},
		{"BarView", []any{other[0], v1, 1}, score.ErrPart},
		{"ChangeMeter", []any{0, 1, meter.MustNew(3, 4)}, score.ErrInvalidArgument},
		{"ChangeMeter", []any{2, 1, meter.MustNew(3, 4)}, score.ErrInvalidArgument},
		{"ChangeMeter", []any{1, 1, meter.Meter{}}, score.ErrInvalidArgument},
		{"CommonBarlines", []any{pos(1, 1), pos(0, 1), ids[0], ids[1]}, score.ErrInvalidArgument},
		{"CommonBarlines", []any{pos(0, 1), pos(1, 1), ids[0]}, score.ErrInvalidArgument},
		{"CommonBarlines", []any{pos(0, 1), pos(1, 1), ids[0], ids[0]}, score.ErrInvalidArgument},
		{"CommonBarlines", []any{pos(0, 1), pos(2, 1), ids[0], other[0]}, score.ErrPart},
		{"CommonBarlines", []any{pos(0, 1), pos(2, 1), ids[0], ids[1]}, score.ErrPosition},
		{"BarlinePeriod", []any{ids[0], ids[0]}, score.ErrInvalidArgument},
		{"BarlinePeriod", []any{ids[0], other[0]}, score.ErrPart},
	} {
		t.Run(tc.name, func(t *testing.T) { checkFailureAndMust(t, s, tc) })
	}
	for _, tc := range []struct {
		name       string
		selections []score.Selection
		cause      error
	}{
		{"zero part", []score.Selection{{}}, score.ErrPart},
		{"foreign part", []score.Selection{{Part: other[0]}}, score.ErrPart},
		{"zero voice", []score.Selection{{Part: ids[0], Voices: []track.VoiceID{{}}}}, score.ErrVoice},
		{"foreign voice", []score.Selection{{Part: ids[0], Voices: []track.VoiceID{unknownVoice}}}, score.ErrVoice},
		{"repeated voice", []score.Selection{{Part: ids[0], Voices: []track.VoiceID{v1, v1}}}, score.ErrInvalidArgument},
		{"repeated part disjoint voices", []score.Selection{{Part: ids[0], Voices: []track.VoiceID{v1}}, {Part: ids[0], Voices: []track.VoiceID{v2}}}, score.ErrInvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, method := range []string{"At", "SoundingAt", "Sounds", "SoundSeconds"} {
				args := []any{}
				if method == "At" || method == "SoundingAt" {
					args = append(args, pos(1, 1))
				}
				if method == "SoundSeconds" {
					args = append(args, 3, rounding.HalfUp)
				}
				for _, sel := range tc.selections {
					args = append(args, sel)
				}
				checkFailureAndMust(t, s, failingCall{method, args, tc.cause})
			}
			if tc.cause == score.ErrVoice {
				_, err := s.Sounds(tc.selections...)
				assert.ErrorIs(t, err, track.ErrVoice)
			}
		})
	}
	empty := score.MustNew(clock60())
	assert.Empty(t, empty.MustReorder().MustParts())
	checkFailureAndMust(t, empty, failingCall{"ChangeMeter", []any{1, 1, meter.MustNew(3, 4)}, score.ErrInvalidArgument})
	checkFailureAndMust(t, empty, failingCall{"At", []any{pos(0, 1), score.Selection{Part: ids[0]}}, score.ErrPart})
}

func TestProblemsErrorAndUnwrap(t *testing.T) {
	t.Parallel()
	for _, p := range []*score.Problems{nil, {}} {
		assert.Empty(t, p.Error())
		assert.Nil(t, p.Unwrap())
	}
	a, b := errors.New("first cause"), errors.New("second cause")
	p := &score.Problems{Issues: []score.Issue{{Cause: a}, {Cause: nil}, {Cause: b}}}
	assert.NotEmpty(t, p.Error())
	assert.Contains(t, p.Error(), a.Error())
	assert.Contains(t, p.Error(), b.Error())
	assert.ErrorIs(t, p, a)
	assert.ErrorIs(t, p, b)
	assert.Equal(t, []error{a, b}, p.Unwrap())
	unwrapped := p.Unwrap()
	unwrapped[0] = nil
	assert.ErrorIs(t, p, a)
	// Sentinels represent distinguishable public error classes.
	sentinels := []error{score.ErrUnset, score.ErrInvalidArgument, score.ErrPart, score.ErrVoice, score.ErrPosition, score.ErrLocalClock, score.ErrGridShift, score.ErrMapAddress, score.ErrMapConflict, score.ErrPeriod, score.ErrOutOfRange}
	for i, e := range sentinels {
		require.NotNil(t, e)
		for j, other := range sentinels {
			assert.Equal(t, i == j, errors.Is(e, other))
		}
	}
}
