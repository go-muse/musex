package mode_test

import (
	"strings"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ionian = "M2 M2 m2 M2 M2 M2 m2"
const dorian = "M2 m2 M2 M2 M2 m2 M2"
const aeolian = "M2 m2 M2 M2 m2 M2 M2"

func intervals(text string) []interval.Interval {
	var result []interval.Interval
	for _, token := range strings.Fields(text) {
		result = append(result, interval.MustParse(token))
	}
	return result
}

func intervalText(values []interval.Interval) string {
	tokens := make([]string, len(values))
	for i, value := range values {
		tokens[i] = value.String()
	}
	return strings.Join(tokens, " ")
}

func checkMode(t *testing.T, m mode.Mode, want string) {
	t.Helper()
	require.True(t, m.IsValid())
	got, err := m.IntervalsString()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func checkDegrees(t *testing.T, ds []degree.Degree, numbers ...int) {
	t.Helper()
	require.Len(t, ds, len(numbers))
	for i, d := range ds {
		assert.True(t, d.IsValid(), "index %d", i)
		assert.Equal(t, numbers[i], d.Number(), "index %d", i)
		assert.False(t, d.HasNote(), "index %d", i)
	}
}

func entry(code, structure string, names ...string) template.Entry {
	return template.MustNewEntry(code, names, template.MustParse(structure))
}

func codes(es []template.Entry) []string {
	result := make([]string, len(es))
	for i, e := range es {
		result[i] = e.Code()
	}
	return result
}

func notes(t *testing.T, names string, delta cents.Cents, src settings.Source) []note.Note {
	t.Helper()
	var result []note.Note
	for _, name := range strings.Fields(names) {
		n, err := note.FromSpelling(spelling.MustParse(name), delta, src)
		require.NoError(t, err)
		result = append(result, n)
	}
	return result
}

func noteNames(ns []note.Note) string {
	names := make([]string, len(ns))
	for i, n := range ns {
		names[i] = n.Spelling().String()
	}
	return strings.Join(names, " ")
}

func irregularSystem() *system.System {
	return system.MustFromSteps(letter.C, [letter.Count]int{3, 1, 1, 2, 2, 2, 1}, 1, system.Integers(), "Irregular test system")
}

func extremeSeconds() []interval.Interval {
	magnitude := alteration.MaxMagnitude
	result := make([]interval.Interval, 4)
	for i, delta := range []int{-magnitude, magnitude, magnitude - 1, -magnitude} {
		result[i] = interval.MustFromDeviation(2, alteration.MustFromInt(delta))
	}
	return result
}

func checkError(t *testing.T, err, category error, causes ...error) {
	t.Helper()
	require.ErrorIs(t, err, category)
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
	for _, other := range modeErrors() {
		if other != category {
			assert.NotErrorIs(t, err, other)
		}
	}
}

func modeErrors() []error {
	return []error{
		mode.ErrUnset, mode.ErrInvalidArgument, mode.ErrEmptyInput, mode.ErrInvalidStructure,
		mode.ErrSystemOctaveRequired, mode.ErrWrongDegreeCount, mode.ErrIncomparable,
		mode.ErrNonPositiveNumber, mode.ErrNoEntry, mode.ErrAmbiguousName, mode.ErrOutOfRange,
		mode.ErrParse, mode.ErrNoSystem, mode.ErrCannotBuildNote, mode.ErrOpenMode,
	}
}
