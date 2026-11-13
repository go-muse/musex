package tuning_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tuningSentinels = []error{
	tuning.ErrUnset,
	tuning.ErrInvalidArgument,
	tuning.ErrClassCount,
	tuning.ErrInadmissible,
	tuning.ErrIncomparable,
	tuning.ErrOutOfRange,
	tuning.ErrTooComplex,
	tuning.ErrSpelling,
}

func requireTuningError(t *testing.T, err error, want error, msgAndArgs ...any) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
	assert.ErrorIs(t, err, want, msgAndArgs...)

	matched := make([]string, 0, 1)
	for _, candidate := range tuningSentinels {
		if errors.Is(err, candidate) {
			matched = append(matched, candidate.Error())
		}
	}
	assert.Len(t, matched, 1, "error must match exactly one tuning sentinel: %v", err)
}

func requireNoDependencyError(t *testing.T, err error) {
	t.Helper()
	for _, candidate := range dependencySentinels() {
		assert.NotErrorIs(t, err, candidate)
	}
}

func requireCentsEqual(t *testing.T, want, got cents.Cents, msgAndArgs ...any) {
	t.Helper()
	assert.True(t, want.Equal(got), "cents mismatch: want %s, got %s%s", want.String(), got.String(), fmt.Sprint(msgAndArgs...))
}

func requireTuning(t *testing.T, got *tuning.Tuning, err error) *tuning.Tuning {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, got)
	return got
}

func aspn(name spelling.Spelling, octave int) pitch.Pitch {
	return pitch.MustFromASPN(name, octave)
}

func traditionalSystemWithClasses(classes int) *system.System {
	def := system.Traditional().Definition()
	def.Name = fmt.Sprintf("test-%d-classes", classes)
	def.Classes = classes
	for i := range def.Offsets {
		def.Offsets[i] = 0
	}
	return system.MustNew(def)
}

func traditionalSystemWithAlterations(from, to int, name string) *system.System {
	def := system.Traditional().Definition()
	def.Name = name
	def.Alterations = system.MustIntegerRange(from, to)
	return system.MustNew(def)
}

func newTestTuning(t *testing.T, sys *system.System, temp temperament.Temperament, root tuning.Root, ref tuning.Reference) *tuning.Tuning {
	t.Helper()
	got, err := tuning.New(sys, temp, root, ref)
	return requireTuning(t, got, err)
}

func newTestTuningAt(t *testing.T, sys *system.System, temp temperament.Temperament, root tuning.Root, note pitch.Pitch, freq float64) *tuning.Tuning {
	t.Helper()
	got, err := tuning.NewAt(sys, temp, root, note, freq)
	return requireTuning(t, got, err)
}

func newPythagoreanTuning(t *testing.T, root tuning.Root) *tuning.Tuning {
	t.Helper()
	return newTestTuning(t, system.Traditional(), temperament.PythagoreanEFlatGSharp(), root, tuning.A4At440())
}

func floorDiv(value, divisor int) (int, int) {
	q := value / divisor
	r := value % divisor
	if r < 0 {
		q--
		r += divisor
	}
	return q, r
}

// heightAt evaluates architecture 6.3 independently through dependency APIs.
func heightAt(t *testing.T, sys *system.System, temp temperament.Temperament, rootNumber int, reference pitch.Pitch, k system.Step) cents.Cents {
	t.Helper()
	n := sys.Classes()
	require.Equal(t, n, temp.StepsPerOctave())
	referenceClass := sys.MustPitchClass(reference)
	m0 := (referenceClass.Number() - rootNumber) % n
	if m0 < 0 {
		m0 += n
	}
	q, m := floorDiv(m0+int(k), n)
	octaves := cents.MustFromInt(1200 * q)
	delta := temp.MustOffset(m).MustSub(temp.MustOffset(m0))
	return octaves.MustAdd(delta)
}

func frequencyAt(t *testing.T, sys *system.System, temp temperament.Temperament, rootNumber int, ref tuning.Reference, k system.Step) float64 {
	t.Helper()
	return heightAt(t, sys, temp, rootNumber, ref.Pitch(), k).MustApplyTo(ref.Frequency())
}

func complexTuning(t *testing.T) *tuning.Tuning {
	t.Helper()
	const (
		largerDenominator  = 10_000_000_000_000_039
		smallerDenominator = 10_000_000_000_000_033
	)
	temp := temperament.MustFromOffsets("record-limit", []cents.Cents{
		cents.Zero(),
		cents.MustFromFraction(1, largerDenominator),
		cents.MustFromFraction(1, smallerDenominator),
	})
	sys := traditionalSystemWithClasses(3)
	ref := tuning.MustNewReference(aspn(spelling.CSharp(), 4), 440)
	return newTestTuning(t, sys, temp, tuning.RootFromClass(pitchclass.MustNew(0, 3)), ref)
}
