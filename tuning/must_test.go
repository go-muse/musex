package tuning_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requirePanicError(t *testing.T, want error, call func()) {
	t.Helper()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		call()
	}()
	require.NotNil(t, recovered, "Must variant must panic")
	err, ok := recovered.(error)
	require.True(t, ok, "panic value must be an error, got %T", recovered)
	assert.ErrorIs(t, err, want)
}

func TestMustSuccessMatchesPlainOperations(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	a4 := aspn(spelling.A(), 4)
	c5 := aspn(spelling.C(), 5)
	ref := tuning.A4At435()
	for _, tc := range []struct {
		name  string
		plain func() (any, error)
		must  func() any
	}{
		{name: "NewReference", plain: func() (any, error) { v, err := tuning.NewReference(a4, 440); return v, err }, must: func() any { return tuning.MustNewReference(a4, 440) }},
		{name: "Reference.WithFrequency", plain: func() (any, error) { v, err := tuning.A4At440().WithFrequency(435); return v, err }, must: func() any { return tuning.A4At440().MustWithFrequency(435) }},
		{name: "Reference.FrequencyRatio", plain: func() (any, error) { v, err := tuning.A4At440().FrequencyRatio(435); return v, err }, must: func() any { return tuning.A4At440().MustFrequencyRatio(435) }},
		{name: "Reference.CentsTo", plain: func() (any, error) { v, err := tuning.A4At440().CentsTo(ref); return v, err }, must: func() any { return tuning.A4At440().MustCentsTo(ref) }},
		{name: "New", plain: func() (any, error) {
			v, err := tuning.New(system.Traditional(), temperament.Equal12(), root, tuning.A4At440())
			return v, err
		}, must: func() any { return tuning.MustNew(system.Traditional(), temperament.Equal12(), root, tuning.A4At440()) }},
		{name: "NewAt", plain: func() (any, error) {
			v, err := tuning.NewAt(system.Traditional(), temperament.Equal12(), root, a4, 440)
			return v, err
		}, must: func() any { return tuning.MustNewAt(system.Traditional(), temperament.Equal12(), root, a4, 440) }},
		{name: "StepOf", plain: func() (any, error) { v, err := std.StepOf(c5); return v, err }, must: func() any { return std.MustStepOf(c5) }},
		{name: "PitchOfStep", plain: func() (any, error) { v, err := std.PitchOfStep(1, system.Sharps()); return v, err }, must: func() any { return std.MustPitchOfStep(1, system.Sharps()) }},
		{name: "CentsFromReference", plain: func() (any, error) { v, err := std.CentsFromReference(c5); return v, err }, must: func() any { return std.MustCentsFromReference(c5) }},
		{name: "CentsFromReferenceOfStep", plain: func() (any, error) { v, err := std.CentsFromReferenceOfStep(3); return v, err }, must: func() any { return std.MustCentsFromReferenceOfStep(3) }},
		{name: "CentsBetween", plain: func() (any, error) { v, err := std.CentsBetween(a4, c5); return v, err }, must: func() any { return std.MustCentsBetween(a4, c5) }},
		{name: "CentsBetweenSteps", plain: func() (any, error) { v, err := std.CentsBetweenSteps(0, 3); return v, err }, must: func() any { return std.MustCentsBetweenSteps(0, 3) }},
		{name: "Frequency", plain: func() (any, error) { v, err := std.Frequency(c5); return v, err }, must: func() any { return std.MustFrequency(c5) }},
		{name: "FrequencyOfStep", plain: func() (any, error) { v, err := std.FrequencyOfStep(3); return v, err }, must: func() any { return std.MustFrequencyOfStep(3) }},
		{name: "ShiftedFrequency", plain: func() (any, error) { v, err := std.ShiftedFrequency(c5, cents.MustFromInt(5)); return v, err }, must: func() any { return std.MustShiftedFrequency(c5, cents.MustFromInt(5)) }},
		{name: "ShiftedFrequencyOfStep", plain: func() (any, error) { v, err := std.ShiftedFrequencyOfStep(3, cents.MustFromInt(5)); return v, err }, must: func() any { return std.MustShiftedFrequencyOfStep(3, cents.MustFromInt(5)) }},
		{name: "Nearest", plain: func() (any, error) { k, d, err := std.Nearest(445); return []any{k, d}, err }, must: func() any { k, d := std.MustNearest(445); return []any{k, d} }},
		{name: "NearestBy", plain: func() (any, error) { k, d, err := std.NearestBy(445, rounding.Up); return []any{k, d}, err }, must: func() any { k, d := std.MustNearestBy(445, rounding.Up); return []any{k, d} }},
		{name: "NearestPitch", plain: func() (any, error) { p, d, err := std.NearestPitch(445, system.Sharps()); return []any{p, d}, err }, must: func() any { p, d := std.MustNearestPitch(445, system.Sharps()); return []any{p, d} }},
		{name: "NearestPitchBy", plain: func() (any, error) {
			p, d, err := std.NearestPitchBy(445, rounding.Up, system.Sharps())
			return []any{p, d}, err
		}, must: func() any { p, d := std.MustNearestPitchBy(445, rounding.Up, system.Sharps()); return []any{p, d} }},
		{name: "WithReferenceNote", plain: func() (any, error) { v, err := std.WithReferenceNote(c5); return v, err }, must: func() any { return std.MustWithReferenceNote(c5) }},
		{name: "WithReferenceFrequency", plain: func() (any, error) { v, err := std.WithReferenceFrequency(435); return v, err }, must: func() any { return std.MustWithReferenceFrequency(435) }},
		{name: "FrequencyRatio", plain: func() (any, error) { v, err := std.FrequencyRatio(435); return v, err }, must: func() any { return std.MustFrequencyRatio(435) }},
		{name: "WithReference", plain: func() (any, error) { v, err := std.WithReference(ref); return v, err }, must: func() any { return std.MustWithReference(ref) }},
		{name: "WithReferenceAt", plain: func() (any, error) { v, err := std.WithReferenceAt(c5, 523); return v, err }, must: func() any { return std.MustWithReferenceAt(c5, 523) }},
		{name: "LevelDifference", plain: func() (any, error) { v, err := std.LevelDifference(std.MustWithReferenceFrequency(435)); return v, err }, must: func() any { return std.MustLevelDifference(std.MustWithReferenceFrequency(435)) }},
		{name: "LevelRatio", plain: func() (any, error) { v, err := std.LevelRatio(std.MustWithReferenceFrequency(435)); return v, err }, must: func() any { return std.MustLevelRatio(std.MustWithReferenceFrequency(435)) }},
		{name: "WithRoot", plain: func() (any, error) { v, err := std.WithRoot(tuning.RootFromName(spelling.A())); return v, err }, must: func() any { return std.MustWithRoot(tuning.RootFromName(spelling.A())) }},
		{name: "WithTemperament", plain: func() (any, error) {
			v, err := std.WithTemperament(temperament.PythagoreanEFlatGSharp())
			return v, err
		}, must: func() any { return std.MustWithTemperament(temperament.PythagoreanEFlatGSharp()) }},
		{name: "WithSystem", plain: func() (any, error) {
			v, err := std.WithSystem(traditionalSystemWithAlterations(-1, 1, "limited"))
			return v, err
		}, must: func() any { return std.MustWithSystem(traditionalSystemWithAlterations(-1, 1, "limited")) }},
		{name: "ChangedClasses", plain: func() (any, error) { v, err := std.ChangedClasses(tuning.RootFromName(spelling.A())); return v, err }, must: func() any { return std.MustChangedClasses(tuning.RootFromName(spelling.A())) }},
		{name: "SoundsSame", plain: func() (any, error) {
			v, err := std.SoundsSame(std.MustWithReferenceNote(aspn(spelling.A(), 5)))
			return v, err
		}, must: func() any { return std.MustSoundsSame(std.MustWithReferenceNote(aspn(spelling.A(), 5))) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			want, err := tc.plain()
			require.NoError(t, err)
			assert.Equal(t, want, tc.must())
		})
	}
}

func TestMustPanicsWithOperationError(t *testing.T) {
	t.Parallel()
	var unset *tuning.Tuning
	validNote := aspn(spelling.A(), 4)
	validRef := tuning.A4At440()
	validRoot := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	for _, tc := range []struct {
		name string
		want error
		call func()
	}{
		{name: "MustNewReference", want: tuning.ErrInvalidArgument, call: func() { tuning.MustNewReference(validNote, 0) }},
		{name: "MustWithFrequency", want: tuning.ErrUnset, call: func() { tuning.Reference{}.MustWithFrequency(440) }},
		{name: "Reference.MustFrequencyRatio", want: tuning.ErrUnset, call: func() { tuning.Reference{}.MustFrequencyRatio(440) }},
		{name: "MustCentsTo", want: tuning.ErrUnset, call: func() { tuning.Reference{}.MustCentsTo(validRef) }},
		{name: "MustNew", want: tuning.ErrUnset, call: func() { tuning.MustNew(nil, temperament.Equal12(), validRoot, validRef) }},
		{name: "MustNewAt", want: tuning.ErrUnset, call: func() { tuning.MustNewAt(nil, temperament.Equal12(), validRoot, validNote, 440) }},
		{name: "MustStepOf", want: tuning.ErrUnset, call: func() { unset.MustStepOf(validNote) }},
		{name: "MustPitchOfStep", want: tuning.ErrUnset, call: func() { unset.MustPitchOfStep(0, system.Sharps()) }},
		{name: "MustCentsFromReference", want: tuning.ErrUnset, call: func() { unset.MustCentsFromReference(validNote) }},
		{name: "MustCentsFromReferenceOfStep", want: tuning.ErrUnset, call: func() { unset.MustCentsFromReferenceOfStep(0) }},
		{name: "MustCentsBetween", want: tuning.ErrUnset, call: func() { unset.MustCentsBetween(validNote, validNote) }},
		{name: "MustCentsBetweenSteps", want: tuning.ErrUnset, call: func() { unset.MustCentsBetweenSteps(0, 1) }},
		{name: "MustFrequency", want: tuning.ErrUnset, call: func() { unset.MustFrequency(validNote) }},
		{name: "MustFrequencyOfStep", want: tuning.ErrUnset, call: func() { unset.MustFrequencyOfStep(0) }},
		{name: "MustShiftedFrequency", want: tuning.ErrUnset, call: func() { unset.MustShiftedFrequency(validNote, cents.Zero()) }},
		{name: "MustShiftedFrequencyOfStep", want: tuning.ErrUnset, call: func() { unset.MustShiftedFrequencyOfStep(0, cents.Zero()) }},
		{name: "MustNearest", want: tuning.ErrUnset, call: func() { unset.MustNearest(440) }},
		{name: "MustNearestBy", want: tuning.ErrUnset, call: func() { unset.MustNearestBy(440, rounding.HalfDown) }},
		{name: "MustNearestPitch", want: tuning.ErrUnset, call: func() { unset.MustNearestPitch(440, system.Sharps()) }},
		{name: "MustNearestPitchBy", want: tuning.ErrUnset, call: func() { unset.MustNearestPitchBy(440, rounding.HalfDown, system.Sharps()) }},
		{name: "MustWithReferenceNote", want: tuning.ErrUnset, call: func() { unset.MustWithReferenceNote(validNote) }},
		{name: "MustWithReferenceFrequency", want: tuning.ErrUnset, call: func() { unset.MustWithReferenceFrequency(440) }},
		{name: "Tuning.MustFrequencyRatio", want: tuning.ErrUnset, call: func() { unset.MustFrequencyRatio(440) }},
		{name: "MustWithReference", want: tuning.ErrUnset, call: func() { unset.MustWithReference(validRef) }},
		{name: "MustWithReferenceAt", want: tuning.ErrUnset, call: func() { unset.MustWithReferenceAt(validNote, 440) }},
		{name: "MustLevelDifference", want: tuning.ErrUnset, call: func() { unset.MustLevelDifference(tuning.Standard()) }},
		{name: "MustLevelRatio", want: tuning.ErrUnset, call: func() { unset.MustLevelRatio(tuning.Standard()) }},
		{name: "MustWithRoot", want: tuning.ErrUnset, call: func() { unset.MustWithRoot(validRoot) }},
		{name: "MustWithTemperament", want: tuning.ErrUnset, call: func() { unset.MustWithTemperament(temperament.Equal12()) }},
		{name: "MustWithSystem", want: tuning.ErrUnset, call: func() { unset.MustWithSystem(system.Traditional()) }},
		{name: "MustChangedClasses", want: tuning.ErrUnset, call: func() { unset.MustChangedClasses(validRoot) }},
		{name: "MustSoundsSame", want: tuning.ErrUnset, call: func() { unset.MustSoundsSame(tuning.Standard()) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			requirePanicError(t, tc.want, tc.call)
		})
	}
}
