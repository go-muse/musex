package scale_test

import (
	"errors"
	"iter"
	"reflect"
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Check every error-returning Scale method against the universal nil and Must contracts.
func TestEveryMethodNilAndMustFailure(t *testing.T) {
	t.Parallel()
	var s *scale.Scale
	typ := reflect.TypeOf(s)
	errType := reflect.TypeFor[error]()
	for i := range typ.NumMethod() {
		method := typ.Method(i)
		if strings.HasPrefix(method.Name, "Must") || method.Type.NumOut() == 0 || method.Type.Out(method.Type.NumOut()-1) != errType {
			continue
		}
		t.Run(method.Name, func(t *testing.T) {
			ordinary := reflect.ValueOf(s).MethodByName(method.Name)
			must := reflect.ValueOf(s).MethodByName("Must" + method.Name)
			require.True(t, must.IsValid(), "missing Must wrapper")
			args := make([]reflect.Value, ordinary.Type().NumIn())
			for i := range args {
				args[i] = reflect.Zero(ordinary.Type().In(i))
			}
			call := func(fn reflect.Value) []reflect.Value {
				if fn.Type().IsVariadic() {
					return fn.CallSlice(args)
				}
				return fn.Call(args)
			}
			result := call(ordinary)
			err := result[len(result)-1].Interface().(error)
			assert.ErrorIs(t, err, scale.ErrUnset)
			for _, value := range result[:len(result)-1] {
				assert.True(t, value.IsZero(), "failure must return zero %s", value.Type())
			}
			panicError(t, scale.ErrUnset, func() { call(must) })
		})
	}
}

func TestEveryMustSuccessMatchesOrdinary(t *testing.T) {
	t.Parallel()
	n := at("G", 4, cents.Zero(), settings.Standard())
	other := scale.MustNew(n)
	args := map[string][]any{
		"At": {0}, "First": {}, "Last": {}, "Replace": {[]note.Note{n}}, "Set": {0, n}, "Insert": {0, []note.Note{n}}, "Append": {[]note.Note{n}}, "Remove": {0}, "Delete": {0, 1}, "Clear": {},
		"Contains": {n, []scale.Comparison{}}, "Find": {n, []scale.Comparison{}}, "FindFirst": {n, []scale.Comparison{}}, "FindLast": {n, []scale.Comparison{}}, "Count": {n, []scale.Comparison{}}, "FindSound": {n, frequency.Strict()}, "FindSoundInTuning": {tuning.Standard(), n, frequency.Strict()},
		"EqualBy": {other, scale.ByNote()}, "SameSet": {other, []scale.Comparison{}}, "SameMultiset": {other, []scale.Comparison{}}, "IsSubsetOf": {other, []scale.Comparison{}}, "Intersection": {other, []scale.Comparison{}}, "Difference": {other, []scale.Comparison{}}, "Union": {other, []scale.Comparison{}}, "RotationStarts": {other, []scale.Comparison{}}, "IsRotationOf": {other, []scale.Comparison{}},
		"Move": {0, 1}, "Next": {0}, "Previous": {0}, "MoveFrom": {scale.Selection{Note: n}, 1, []scale.Comparison{}}, "DistanceForward": {0, 0}, "DistanceBackward": {0, 0}, "MirrorIndex": {0, 0},
		"Take": {0, 2, scale.Forward}, "OneRound": {0, scale.Forward}, "Iter": {0, 2, scale.Forward}, "Round": {0, scale.Forward}, "Cycle": {0, scale.Forward},
		"ShiftMelody": {[]note.Note{n}, 1, []scale.Comparison{}}, "ShiftSelected": {[]scale.Selection{{Note: n}}, 1, []scale.Comparison{}}, "MirrorMelody": {[]note.Note{n}, 0, []scale.Comparison{}}, "MirrorSelected": {[]scale.Selection{{Note: n}}, 0, []scale.Comparison{}},
		"Rotate": {[]int{}}, "Reverse": {}, "Slice": {0, 1}, "Repeat": {2}, "Unique": {[]scale.Comparison{}}, "WithoutOctaves": {}, "ShiftOctaves": {1}, "Transpose": {interval.MustParse("M2")}, "TransposeIn": {system.Traditional(), interval.MustParse("M2")}, "TransposeTo": {n, interval.Up}, "TransposeToIn": {system.Traditional(), n, interval.Up},
		"Enharmonics": {0, alteration.MustFromInt(1)}, "EnharmonicsIn": {settings.Standard(), 0, alteration.MustFromInt(1)}, "WithEnharmonic": {0, spelling.MustParse("G")}, "WithEnharmonicIn": {settings.Standard(), 0, spelling.MustParse("G")},
		"PitchClasses": {system.Traditional()}, "KeySignature": {}, "Format": {}, "FormatIn": {settings.Standard()},
		"Tokens": {}, "TokensIn": {settings.Standard()}, "PitchClassSequence": {system.Traditional()},
	}
	typ := reflect.TypeFor[*scale.Scale]()
	errorType := reflect.TypeFor[error]()
	for i := range typ.NumMethod() {
		method := typ.Method(i)
		if strings.HasPrefix(method.Name, "Must") || method.Type.NumOut() == 0 || method.Type.Out(method.Type.NumOut()-1) != errorType {
			continue
		}
		t.Run(method.Name, func(t *testing.T) {
			raw, ok := args[method.Name]
			require.True(t, ok, "success case must be specified for every error-returning method")
			a, b := scale.MustNew(n), scale.MustNew(n)
			values := make([]reflect.Value, len(raw))
			for i, arg := range raw {
				values[i] = reflect.ValueOf(arg)
			}
			call := func(fn reflect.Value) []reflect.Value {
				if fn.Type().IsVariadic() {
					return fn.CallSlice(values)
				}
				return fn.Call(values)
			}
			ordinary := call(reflect.ValueOf(a).MethodByName(method.Name))
			require.True(t, ordinary[len(ordinary)-1].IsNil(), "ordinary method should succeed")
			mustMethod := reflect.ValueOf(b).MethodByName("Must" + method.Name)
			require.True(t, mustMethod.IsValid())
			wrapped := call(mustMethod)
			if len(ordinary) == 1 {
				require.Len(t, wrapped, 1)
				assert.Same(t, b, wrapped[0].Interface())
				assert.True(t, a.Equal(b))
				return
			}
			require.Len(t, wrapped, len(ordinary)-1)
			for i, expected := range ordinary[:len(ordinary)-1] {
				switch value := expected.Interface().(type) {
				case *scale.Scale:
					checkScale(t, value.Notes(), wrapped[i].Interface().(*scale.Scale))
				case iter.Seq[note.Note]:
					checkNotes(t, collect(value, 4), collect(wrapped[i].Interface().(iter.Seq[note.Note]), 4))
				default:
					assert.Equal(t, value, wrapped[i].Interface())
				}
			}
		})
	}
}

func TestFactoryMustAndSentinelTexts(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		err  error
		text string
	}{{scale.ErrUnset, "scale: unset value"}, {scale.ErrInvalidArgument, "scale: invalid argument"}, {scale.ErrEmpty, "scale: empty scale"}, {scale.ErrIndex, "scale: index out of range"}, {scale.ErrNotFound, "scale: note not found"}, {scale.ErrAmbiguous, "scale: ambiguous occurrence"}, {scale.ErrNotEquivalent, "scale: spelling is not equivalent"}, {scale.ErrOutOfRange, "scale: result out of range"}} {
		assert.Equal(t, tt.text, tt.err.Error())
		assert.True(t, errors.Is(tt.err, tt.err))
	}
	panicError(t, scale.ErrUnset, func() { scale.MustNew(note.Note{}) })
	panicError(t, system.ErrInvalid, func() { scale.MustByPositionIn(nil) })
	panicError(t, tuning.ErrUnset, func() { scale.MustBySoundIn(nil) })
	checkScale(t, names("G"), scale.MustNew(names("G")...))
	assert.False(t, reflect.TypeFor[scale.Comparison]().Comparable())
}

func panicWithCauses(t *testing.T, run func(), causes ...error) {
	t.Helper()
	defer func() {
		p := recover()
		err, ok := p.(error)
		require.True(t, ok, "panic must retain the original error, got %v", p)
		for _, cause := range causes {
			assert.ErrorIs(t, err, cause)
		}
	}()
	run()
	t.Fatal("expected panic")
}

func TestMustPreservesLocalAndDependencyCauses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		sys  *system.System
	}{{"nil", nil}, {"zero system", new(system.System)}} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := scale.ByPositionIn(tt.sys)
			assert.ErrorIs(t, err, scale.ErrUnset)
			assert.ErrorIs(t, err, system.ErrInvalid)
			panicWithCauses(t, func() { scale.MustByPositionIn(tt.sys) }, scale.ErrUnset, system.ErrInvalid)
			s := list("G")
			_, err = s.PitchClasses(tt.sys)
			assert.ErrorIs(t, err, scale.ErrUnset)
			assert.ErrorIs(t, err, system.ErrInvalid)
			panicWithCauses(t, func() { s.MustPitchClasses(tt.sys) }, scale.ErrUnset, system.ErrInvalid)
		})
	}
	s := list("G A")
	_, err := s.At(2)
	assert.ErrorIs(t, err, scale.ErrIndex)
	panicWithCauses(t, func() { s.MustAt(2) }, scale.ErrIndex)
	err = s.Set(0, note.Note{})
	assert.ErrorIs(t, err, scale.ErrUnset)
	assert.ErrorIs(t, err, note.ErrUnset)
	panicWithCauses(t, func() { s.MustSet(0, note.Note{}) }, scale.ErrUnset, note.ErrUnset)
	checkScale(t, names("G A"), s)
}
