package mode_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type observedMode struct {
	valid, open, closed, hasCycle, hasOrigin, hasName, changed              bool
	degreeCount                                                             int
	kind                                                                    mode.Kind
	structure, own, adjacent, text, debug, originCode, originTemplate, name string
	originNames                                                             []string
	cycle                                                                   interval.Interval
}

func observe(m mode.Mode) observedMode {
	cycle, hasCycle := m.Cycle()
	origin, hasOrigin := m.Origin()
	name, hasName := m.SelectedName()
	return observedMode{
		valid: m.IsValid(), open: m.IsOpen(), closed: m.IsClosed(), hasCycle: hasCycle,
		hasOrigin: hasOrigin, hasName: hasName, changed: m.Changed(),
		degreeCount: m.Len(), kind: m.Kind(), structure: m.Template().String(),
		own: intervalText(m.DegreeIntervals()), adjacent: intervalText(m.AdjacentIntervals()),
		text: m.String(), debug: m.GoString(), originCode: origin.Code(), originTemplate: origin.Template().String(),
		name: name, originNames: origin.Names(), cycle: cycle,
	}
}

func TestRebuildReplacesOrigin(t *testing.T) {
	t.Parallel()
	c := template.MustNewCatalog(entry("open", "m3 M2 |", "Open shape", "Open alias"))
	for _, tt := range []struct {
		name                               string
		mutate                             func(*mode.Mode) error
		structure, code, selected, display string
	}{
		{"template", func(m *mode.Mode) error { return m.Rebuild(template.MustParse("P8")) }, "P8", "", "", "P8"},
		{"entry", func(m *mode.Mode) error { return m.RebuildFromEntry(entry("two", "M2 |", "Two", "Alias")) }, "M2 |", "two", "", "Two: M2 |"},
		{"code", func(m *mode.Mode) error { return m.RebuildFromCode("dorian") }, dorian, "dorian", "", "дорийский: " + dorian},
		{"custom code", func(m *mode.Mode) error { return m.RebuildFromCodeIn(&c, "open") }, "m3 M2 |", "open", "", "Open shape: m3 M2 |"},
		{"name", func(m *mode.Mode) error { return m.RebuildFromName("  DORIAN ") }, dorian, "dorian", "Dorian", "Dorian: " + dorian},
		{"custom name", func(m *mode.Mode) error { return m.RebuildFromNameIn(&c, "open alias") }, "m3 M2 |", "open", "Open alias", "Open alias: m3 M2 |"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mode.MustFromName("Ionian")
			require.NoError(t, m.Rotate(2))
			require.True(t, m.Changed())
			save := m
			require.NoError(t, tt.mutate(&m))
			checkMode(t, m, tt.structure)
			assert.Equal(t, tt.code, firstOf(m.Origin()).Code())
			assert.Equal(t, tt.code != "", m.HasOrigin())
			assert.Equal(t, tt.selected, firstOf(m.SelectedName()))
			assert.Equal(t, tt.selected != "", secondOf(m.SelectedName()))
			assert.False(t, m.Changed())
			assert.Equal(t, tt.display, m.String())
			assert.True(t, save.Changed())
			assert.Equal(t, "Ionian", firstOf(save.SelectedName()))
		})
	}
}

func TestEditDegreeIntervals(t *testing.T) {
	t.Parallel()
	m := mode.MustFromName("дорийский")
	copyOfOriginal := m
	require.NoError(t, m.SetDegreeInterval(6, interval.MustParse("m6")))
	checkMode(t, m, aeolian)
	assert.True(t, m.Equal(mode.MustFromCode("aeolian")))
	assert.False(t, copyOfOriginal.Changed())
	assert.Equal(t, dorian, copyOfOriginal.MustIntervalsString())
	assert.Equal(t, "dorian", firstOf(m.Origin()).Code())
	assert.Equal(t, "дорийский", firstOf(m.SelectedName()))
	assert.True(t, m.Changed())
	assert.Equal(t, "дорийский (edited): "+aeolian, m.String())
	require.NoError(t, m.SetDegreeInterval(6, interval.MustParse("M6")))
	assert.True(t, m.Equal(copyOfOriginal))
	assert.False(t, m.Changed())
	assert.Equal(t, "дорийский: "+dorian, m.String())

	for _, tt := range []struct {
		name, initial, want string
		edit                func(*mode.Mode) error
	}{
		{"change final closed degree", ionian, "M2 M2 m2 M2 M2 m2 M2", func(m *mode.Mode) error { return m.SetDegreeInterval(7, interval.MustParse("m7")) }},
		{"insert inside", "M2 M2 m3 M2 m3", "M2 M2 m2 M2 M2 m3", func(m *mode.Mode) error { return m.AddDegreeInterval(interval.MustParse("P4")) }},
		{"remove inside", ionian, "M2 M2 m3 M2 m3", func(m *mode.Mode) error {
			if err := m.RemoveDegree(7); err != nil {
				return err
			}
			return m.RemoveDegree(4)
		}},
		{"remove first closed", ionian, "M2 m2 M2 M2 M2 m3", func(m *mode.Mode) error { return m.RemoveDegree(1) }},
		{"last open degree has no upper bound", "m3 M2 |", "m3 P4 |", func(m *mode.Mode) error { return m.SetDegreeInterval(3, interval.MustParse("m7")) }},
		{"append to open", "m3 M2 |", "m3 M2 M3 |", func(m *mode.Mode) error { return m.AddDegreeInterval(interval.MustParse("M6")) }},
		{"remove first open", "m3 M2 |", "M2 |", func(m *mode.Mode) error { return m.RemoveDegree(1) }},
		{"remove last open", "m3 M2 |", "m3 |", func(m *mode.Mode) error { return m.RemoveDegree(3) }},
		{"remove to one degree", "P5 P4", "P8", func(m *mode.Mode) error { return m.RemoveDegree(1) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mode.MustParse(tt.initial)
			before, beforeOK := m.Cycle()
			require.NoError(t, tt.edit(&m))
			checkMode(t, m, tt.want)
			after, afterOK := m.Cycle()
			assert.Equal(t, beforeOK, afterOK)
			if beforeOK {
				assert.True(t, before.Equal(after))
			}
			assert.False(t, m.HasOrigin())
			assert.False(t, m.Changed())
			if tt.name == "insert inside" {
				assert.Equal(t, "P4", m.MustDegreeInterval(4).String())
			}
		})
	}
}

func TestRotateMirrorRepeatAndReduce(t *testing.T) {
	t.Parallel()
	m := mode.MustFromName("Ionian")
	before := observe(m)
	for i := range 7 {
		require.NoError(t, m.Rotate(2))
		assert.Equal(t, i != 6, m.Changed())
		assert.Equal(t, "Ionian", firstOf(m.SelectedName()))
		assert.Equal(t, "ionian", firstOf(m.Origin()).Code())
		if i == 0 {
			checkMode(t, m, dorian)
		}
	}
	assert.Equal(t, before, observe(m))
	for j := 2; j <= m.Len(); j++ {
		candidate := mode.MustFromDegree(m, j)
		require.NoError(t, candidate.Rotate(m.Len()-j+2))
		assert.Equal(t, before, observe(candidate))
	}
	for _, tt := range []struct{ initial, want string }{
		{ionian, "m2 M2 M2 M2 m2 M2 M2"},
		{dorian, dorian},
		{"M2 m2 M2 M2 m2 A2 m2", "m2 A2 m2 M2 M2 m2 M2"},
		{"m3 M2 |", "M2 m3 |"},
	} {
		candidate := mode.MustParse(tt.initial)
		require.NoError(t, candidate.Mirror())
		checkMode(t, candidate, tt.want)
		require.NoError(t, candidate.Mirror())
		checkMode(t, candidate, tt.initial)
	}
	open := mode.MustParse("m3 M2 |")
	require.NoError(t, open.Rotate(1))
	checkMode(t, open, "m3 M2 |")
	m = mode.MustFromCode("whole-tone")
	before = observe(m)
	require.NoError(t, m.ReduceToMinimalPart())
	assert.Equal(t, before, observe(m))
	require.NoError(t, m.Repeat(1))
	assert.Equal(t, before, observe(m))
	m = mode.MustFromEntry(entry("repeat", "M2 m2 M2 m2", "Repeat"))
	require.NoError(t, m.ReduceToMinimalPart())
	checkMode(t, m, "M2 m2")
	assert.Equal(t, "m3", firstOf(m.Cycle()).String())
	assert.True(t, m.Changed())
	require.NoError(t, m.Repeat(2))
	checkMode(t, m, "M2 m2 M2 m2")
	assert.False(t, m.Changed())
	m = mode.MustParse("M2 M2 m2")
	require.NoError(t, m.Repeat(2))
	checkMode(t, m, "M2 M2 m2 M2 M2 m2")
	assert.Equal(t, "m7", firstOf(m.Cycle()).String())

	m = mode.MustFromCode("ionian")
	assert.Same(t, &m, m.MustRotate(6).MustRemoveDegree(7))
	checkMode(t, m, "M2 m2 M2 M2 m2 M3")
}

func TestMutationFailuresAreAtomic(t *testing.T) {
	t.Parallel()
	m := mode.MustFromName("Ionian")
	require.NoError(t, m.Rotate(2))
	open := mode.MustFromEntry(entry("open", "m3 M2 |", "Open"))
	max := mode.MustClosedFromIntervals(slices.Repeat(intervals("M2"), mode.MaxDegrees))
	huge := mode.MustClosedFromIntervals([]interval.Interval{interval.MustFromDeviation(math.MaxInt, alteration.MustFromInt(0))})
	for _, tt := range []struct {
		name     string
		input    mode.Mode
		mutate   func(*mode.Mode) error
		category error
	}{
		{"unset template", m, func(m *mode.Mode) error { return m.Rebuild(template.Template{}) }, mode.ErrUnset},
		{"unset entry", m, func(m *mode.Mode) error { return m.RebuildFromEntry(template.Entry{}) }, mode.ErrUnset},
		{"unknown code", m, func(m *mode.Mode) error { return m.RebuildFromCode("unknown") }, mode.ErrNoEntry},
		{"nil code catalog", m, func(m *mode.Mode) error { return m.RebuildFromCodeIn(nil, "dorian") }, mode.ErrUnset},
		{"ambiguous name", m, func(m *mode.Mode) error { return m.RebuildFromName("blues scale") }, mode.ErrAmbiguousName},
		{"empty name catalog", m, func(m *mode.Mode) error { return m.RebuildFromNameIn(&template.Catalog{}, "Dorian") }, mode.ErrNoEntry},
		{"bad rotation degree", m, func(m *mode.Mode) error { return m.Rotate(8) }, mode.ErrInvalidArgument},
		{"open rotation", open, func(m *mode.Mode) error { return m.Rotate(2) }, mode.ErrOpenMode},
		{"bad open rotation before kind", open, func(m *mode.Mode) error { return m.Rotate(4) }, mode.ErrInvalidArgument},
		{"unset replacement before degree", m, func(m *mode.Mode) error { return m.SetDegreeInterval(0, interval.Interval{}) }, mode.ErrUnset},
		{"first degree", m, func(m *mode.Mode) error { return m.SetDegreeInterval(1, interval.MustParse("M2")) }, mode.ErrInvalidArgument},
		{"misordered replacement", m, func(m *mode.Mode) error { return m.SetDegreeInterval(2, interval.MustParse("A4")) }, mode.ErrInvalidStructure},
		{"replacement reaches cycle", m, func(m *mode.Mode) error { return m.SetDegreeInterval(m.Len(), interval.MustParse("P8")) }, mode.ErrInvalidStructure},
		{"unset insertion", m, func(m *mode.Mode) error { return m.AddDegreeInterval(interval.Interval{}) }, mode.ErrUnset},
		{"duplicate insertion", m, func(m *mode.Mode) error { return m.AddDegreeInterval(interval.MustParse("m3")) }, mode.ErrInvalidArgument},
		{"insertion at cycle", m, func(m *mode.Mode) error { return m.AddDegreeInterval(interval.MustParse("P8")) }, mode.ErrInvalidStructure},
		{"insertion below first", m, func(m *mode.Mode) error { return m.AddDegreeInterval(interval.MustParse("-M2")) }, mode.ErrInvalidStructure},
		{"insertion at degree limit", max, func(m *mode.Mode) error { return m.AddDegreeInterval(interval.MustParse("A1")) }, mode.ErrInvalidArgument},
		{"remove nonexistent", m, func(m *mode.Mode) error { return m.RemoveDegree(0) }, mode.ErrInvalidArgument},
		{"remove from singleton", mode.MustParse("P8"), func(m *mode.Mode) error { return m.RemoveDegree(1) }, mode.ErrInvalidArgument},
		{"remove from smallest open", mode.MustParse("M2 |"), func(m *mode.Mode) error { return m.RemoveDegree(1) }, mode.ErrInvalidArgument},
		{"unset mirror", mode.Mode{}, (*mode.Mode).Mirror, mode.ErrUnset},
		{"zero repeats", m, func(m *mode.Mode) error { return m.Repeat(0) }, mode.ErrInvalidArgument},
		{"repeat above limit", m, func(m *mode.Mode) error { return m.Repeat(math.MaxInt) }, mode.ErrInvalidArgument},
		{"invalid repeat count before open", open, func(m *mode.Mode) error { return m.Repeat(0) }, mode.ErrInvalidArgument},
		{"repeat open", open, func(m *mode.Mode) error { return m.Repeat(2) }, mode.ErrOpenMode},
		{"repeat unrepresentable cycle", huge, func(m *mode.Mode) error { return m.Repeat(2) }, mode.ErrOutOfRange},
		{"reduce open", open, (*mode.Mode).ReduceToMinimalPart, mode.ErrOpenMode},
	} {
		t.Run(tt.name, func(t *testing.T) {
			candidate := tt.input
			before := observe(candidate)
			savedEdges := candidate.AdjacentIntervals()
			savedDegrees := candidate.DegreeIntervals()
			checkError(t, tt.mutate(&candidate), tt.category)
			assert.Equal(t, before, observe(candidate))
			assert.Equal(t, before.adjacent, intervalText(savedEdges))
			assert.Equal(t, before.own, intervalText(savedDegrees))
		})
	}
}

func TestRepresentabilityAtCreationRotationAndMutation(t *testing.T) {
	t.Parallel()
	edges := extremeSeconds()
	sourceTemplate := template.MustClosedFromIntervals(edges)
	rotatedTemplate := firstOf(sourceTemplate.Rotate(2))
	require.True(t, rotatedTemplate.IsValid())
	invalidEntry := template.MustNewEntry("invalid", []string{"Extreme rotation"}, rotatedTemplate)
	catalog := template.MustNewCatalog(invalidEntry)
	m := mode.MustFromEntry(template.MustNewEntry("extreme", []string{"Extreme"}, sourceTemplate))
	assert.Equal(t, "P5", firstOf(m.Cycle()).String())
	require.Len(t, m.DegreeIntervals(), 4)
	assert.True(t, m.Equal(mode.MustParse(m.MustIntervalsString())))
	for _, tt := range []struct {
		name string
		call func() (mode.Mode, error)
	}{
		{"template", func() (mode.Mode, error) { return mode.FromTemplate(rotatedTemplate) }},
		{"entry", func() (mode.Mode, error) { return mode.FromEntry(invalidEntry) }},
		{"code", func() (mode.Mode, error) { return mode.FromCodeIn(&catalog, "invalid") }},
		{"name", func() (mode.Mode, error) { return mode.FromNameIn(&catalog, "Extreme rotation") }},
		{"rotated text", func() (mode.Mode, error) { return mode.Parse(rotatedTemplate.String()) }},
		{"rotate source", func() (mode.Mode, error) { return mode.FromDegree(m, 2) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.call()
			checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
			assert.Zero(t, got)
		})
	}
	ambiguous := template.MustNewCatalog(invalidEntry, entry("different", "P8", "Extreme rotation"))
	got, err := mode.FromNameIn(&ambiguous, "Extreme rotation")
	checkError(t, err, mode.ErrAmbiguousName)
	assert.Zero(t, got)

	seq, err := mode.Rotations(m)
	checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Nil(t, seq)
	before := observe(m)
	checkError(t, m.Rotate(2), mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Equal(t, before, observe(m))
	checkError(t, m.Rebuild(rotatedTemplate), mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Equal(t, before, observe(m))

	open := mode.MustFromEntry(template.MustNewEntry("open", []string{"Open"}, template.MustOpenFromIntervals(edges[:3])))
	assert.True(t, open.Equal(mode.MustParse(open.MustIntervalsString())))
	for _, tt := range []struct {
		name string
		call func(*mode.Mode) error
	}{
		{"remove middle", func(m *mode.Mode) error { return m.RemoveDegree(3) }},
		{"mirror", (*mode.Mode).Mirror},
	} {
		t.Run(tt.name, func(t *testing.T) {
			candidate := open
			before := observe(candidate)
			checkError(t, tt.call(&candidate), mode.ErrOutOfRange, alteration.ErrOutOfRange)
			assert.Equal(t, before, observe(candidate))
		})
	}
	candidate := open
	panicked := capturePanic(func() { candidate.MustRemoveDegree(3) })
	failure, ok := panicked.(error)
	require.True(t, ok, "panic value: %v", panicked)
	checkError(t, failure, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Equal(t, observe(open), observe(candidate))
}
