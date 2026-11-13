package tuning_test

import (
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRootAndChangedClassesPythagorean(t *testing.T) {
	t.Parallel()
	rootC := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	rootA := tuning.RootFromClass(pitchclass.MustNew(9, 12))
	tune := newPythagoreanTuning(t, rootC)
	changed := tune.MustWithRoot(rootA)

	assert.Equal(t, pitchclass.MustNew(9, 12), changed.RootClass())
	assert.Equal(t, pitchclass.MustNew(0, 12), tune.RootClass())
	assert.Equal(t, "704.7949", strconv.FormatFloat(changed.MustFrequency(aspn(spelling.F(), 5)), 'f', 4, 64))
	assert.Equal(t, "521.4815", strconv.FormatFloat(changed.MustFrequency(aspn(spelling.C(), 5)), 'f', 4, 64))
	assert.Equal(t, 440.0, changed.MustFrequency(aspn(spelling.A(), 4)))

	classes, err := tune.ChangedClasses(rootA)
	require.NoError(t, err)
	assert.Equal(t, []pitchclass.Class{
		pitchclass.MustNew(3, 12),
		pitchclass.MustNew(5, 12),
		pitchclass.MustNew(10, 12),
	}, classes, "K-5: E-flat, F and B-flat")
	classes[0] = pitchclass.MustNew(0, 12)
	fresh := tune.MustChangedClasses(rootA)
	assert.Equal(t, pitchclass.MustNew(3, 12), fresh[0], "each result is a new slice")
}

func TestWithRootJustIntonation(t *testing.T) {
	t.Parallel()
	tune := newTestTuning(t, system.Traditional(), temperament.Just5Limit(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), tuning.A4At440())
	changed := tune.MustWithRoot(tuning.RootFromName(spelling.A()))
	assert.Equal(t, pitchclass.MustNew(9, 12), changed.RootClass())
	assert.Equal(t, "550.0000", strconv.FormatFloat(changed.MustFrequency(aspn(spelling.CSharp(), 5)), 'f', 4, 64))
	assert.Equal(t, "563.2000", strconv.FormatFloat(tune.MustFrequency(aspn(spelling.CSharp(), 5)), 'f', 4, 64))
	assert.Equal(t, 440.0, changed.MustFrequency(aspn(spelling.A(), 4)))
}

func TestChangedClassesEqualTemperament(t *testing.T) {
	t.Parallel()
	tune := tuning.Standard()
	rootA := tuning.RootFromClass(pitchclass.MustNew(9, 12))
	classes, err := tune.ChangedClasses(rootA)
	require.NoError(t, err)
	assert.NotNil(t, classes)
	assert.Len(t, classes, 0, "G-8")
	assert.True(t, tune.MustWithRoot(rootA).MustSoundsSame(tune))
}

func TestWithRootErrors(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	_, err := std.WithRoot(tuning.Root{})
	requireTuningError(t, err, tuning.ErrUnset)
	_, err = std.ChangedClasses(tuning.Root{})
	requireTuningError(t, err, tuning.ErrUnset)

	badClass := tuning.RootFromClass(pitchclass.MustNew(0, 19))
	_, err = std.WithRoot(badClass)
	requireTuningError(t, err, tuning.ErrInvalidArgument)
	_, err = std.ChangedClasses(badClass)
	requireTuningError(t, err, tuning.ErrInvalidArgument)

	badName := tuning.RootFromName(spelling.MustParse("C+1/2"))
	_, err = std.WithRoot(badName)
	requireTuningError(t, err, tuning.ErrInadmissible)
	_, err = std.ChangedClasses(badName)
	requireTuningError(t, err, tuning.ErrInadmissible)
	assert.True(t, tuning.Standard().Equal(std))
}

func TestChangedClassesComplexity(t *testing.T) {
	t.Parallel()
	tune := complexTuning(t)
	_, err := tune.ChangedClasses(tuning.RootFromClass(pitchclass.MustNew(2, 3)))
	requireTuningError(t, err, tuning.ErrTooComplex)
	assert.ErrorIs(t, err, cents.ErrTooComplex)
}

func TestWithTemperament(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	changed, err := std.WithTemperament(temperament.PythagoreanEFlatGSharp())
	require.NoError(t, err)
	assert.True(t, temperament.PythagoreanEFlatGSharp().Equal(changed.Temperament()))
	assert.True(t, temperament.Equal12().Equal(std.Temperament()))

	_, err = std.WithTemperament(temperament.Temperament{})
	requireTuningError(t, err, tuning.ErrUnset)
	_, err = std.WithTemperament(temperament.Equal19())
	requireTuningError(t, err, tuning.ErrClassCount)
	requireNoDependencyError(t, err)
}

func TestWithSystem(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	limited := traditionalSystemWithAlterations(-1, 1, "traditional layout with single accidentals")
	changed, err := std.WithSystem(limited)
	require.NoError(t, err)
	assert.True(t, limited.Equal(changed.System()))
	assert.True(t, system.Traditional().Equal(std.System()))
	assert.Equal(t, "523.2511", strconv.FormatFloat(changed.MustFrequency(aspn(spelling.C(), 5)), 'f', 4, 64), "K-24")

	doubleFlatReference := std.MustFrequency(aspn(spelling.BDoubleFlat(), 4))
	withDoubleFlat := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.BDoubleFlat(), 4), doubleFlatReference)
	_, err = withDoubleFlat.WithSystem(limited)
	requireTuningError(t, err, tuning.ErrInadmissible)

	_, err = std.WithSystem(system.Positions19())
	requireTuningError(t, err, tuning.ErrClassCount)
	_, err = std.WithSystem(nil)
	requireTuningError(t, err, tuning.ErrUnset)
	_, err = std.WithSystem(&system.System{})
	requireTuningError(t, err, tuning.ErrUnset)
}

func TestWithOperationsDoNotMutateInput(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	baseline := tuning.Standard()
	for _, tc := range []struct {
		name string
		call func() error
	}{
		{name: "root", call: func() error { _, err := std.WithRoot(tuning.RootFromName(spelling.A())); return err }},
		{name: "temperament", call: func() error { _, err := std.WithTemperament(temperament.PythagoreanEFlatGSharp()); return err }},
		{name: "system", call: func() error { _, err := std.WithSystem(traditionalSystemWithAlterations(-1, 1, "limited")); return err }},
		{name: "reference", call: func() error { _, err := std.WithReference(tuning.A4At435()); return err }},
		{name: "reference at", call: func() error { _, err := std.WithReferenceAt(aspn(spelling.C(), 5), 523); return err }},
		{name: "reference note", call: func() error { _, err := std.WithReferenceNote(aspn(spelling.C(), 5)); return err }},
		{name: "reference frequency", call: func() error { _, err := std.WithReferenceFrequency(435); return err }},
		{name: "root error", call: func() error { _, err := std.WithRoot(tuning.RootFromClass(pitchclass.MustNew(0, 19))); return err }},
		{name: "temperament error", call: func() error { _, err := std.WithTemperament(temperament.Equal19()); return err }},
		{name: "system error", call: func() error { _, err := std.WithSystem(system.Positions19()); return err }},
		{name: "reference error", call: func() error { _, err := std.WithReference(tuning.Reference{}); return err }},
		{name: "frequency error", call: func() error { _, err := std.WithReferenceFrequency(0); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_ = tc.call()
			assert.True(t, baseline.Equal(std))
			assert.Equal(t, 440.0, std.MustFrequency(aspn(spelling.A(), 4)))
		})
	}
}
