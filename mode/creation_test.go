package mode_test

import (
	"math"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
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

func TestCreateFromCatalog(t *testing.T) {
	t.Parallel()
	fromCode, err := mode.FromCode("dorian")
	require.NoError(t, err)
	checkMode(t, fromCode, dorian)
	origin, ok := fromCode.Origin()
	require.True(t, ok)
	assert.True(t, fromCode.HasOrigin())
	assert.Equal(t, "dorian", origin.Code())
	name, selected := fromCode.SelectedName()
	assert.Empty(t, name)
	assert.False(t, selected)
	assert.False(t, fromCode.Changed())

	for _, want := range []string{"дорийский", "Dorian"} {
		m, err := mode.FromName(want)
		require.NoError(t, err)
		assert.True(t, m.Equal(fromCode))
		got, ok := m.SelectedName()
		assert.True(t, ok)
		assert.Equal(t, want, got)
	}
	a := mode.MustFromName("natural minor")
	b := mode.MustFromName("Aeolian")
	assert.True(t, a.Equal(b))
	assert.Equal(t, "natural minor", firstOf(a.SelectedName()))
	assert.Equal(t, "Aeolian", firstOf(b.SelectedName()))

	first := entry("Chosen", "m3 M2 |", "Original", "Café Mode")
	duplicate := entry("Other", "m3 M2 |", "CAFÉ MODE")
	cat := template.MustNewCatalog(first, duplicate)
	m, err := mode.FromNameIn(&cat, "  CAFE\u0301\t  MODE  ")
	require.NoError(t, err)
	checkMode(t, m, "m3 M2 |")
	assert.True(t, first.Equal(firstOf(m.Origin())))
	assert.Equal(t, "Café Mode", firstOf(m.SelectedName()))
	m, err = mode.FromCodeIn(&cat, "Chosen")
	require.NoError(t, err)
	assert.True(t, first.Equal(firstOf(m.Origin())))
	assert.Equal(t, "Original: m3 M2 |", m.String())
	m, err = mode.FromEntry(first)
	require.NoError(t, err)
	assert.True(t, first.Equal(firstOf(m.Origin())))
	assert.False(t, secondOf(m.SelectedName()))
	m, err = mode.FromTemplate(first.Template())
	require.NoError(t, err)
	checkMode(t, m, "m3 M2 |")
	assert.False(t, m.HasOrigin())

	ambiguous := template.MustNewCatalog(first, entry("different", "M2 |", "Café Mode"))
	var empty template.Catalog
	tests := []struct {
		name     string
		makeMode func() (mode.Mode, error)
		category error
	}{
		{"unset template", func() (mode.Mode, error) { return mode.FromTemplate(template.Template{}) }, mode.ErrUnset},
		{"unset entry", func() (mode.Mode, error) { return mode.FromEntry(template.Entry{}) }, mode.ErrUnset},
		{"nil catalog code", func() (mode.Mode, error) { return mode.FromCodeIn(nil, "dorian") }, mode.ErrUnset},
		{"nil catalog name", func() (mode.Mode, error) { return mode.FromNameIn(nil, "Dorian") }, mode.ErrUnset},
		{"empty catalog code", func() (mode.Mode, error) { return mode.FromCodeIn(&empty, "dorian") }, mode.ErrNoEntry},
		{"empty catalog name", func() (mode.Mode, error) { return mode.FromNameIn(&empty, "Dorian") }, mode.ErrNoEntry},
		{"unknown code", func() (mode.Mode, error) { return mode.FromCode("nope") }, mode.ErrNoEntry},
		{"exact code case", func() (mode.Mode, error) { return mode.FromCodeIn(&cat, "chosen") }, mode.ErrNoEntry},
		{"exact code whitespace", func() (mode.Mode, error) { return mode.FromCodeIn(&cat, " Chosen") }, mode.ErrNoEntry},
		{"unknown name", func() (mode.Mode, error) { return mode.FromName("no such mode") }, mode.ErrNoEntry},
		{"built-in ambiguous name", func() (mode.Mode, error) { return mode.FromName("blues scale") }, mode.ErrAmbiguousName},
		{"custom ambiguous name", func() (mode.Mode, error) { return mode.FromNameIn(&ambiguous, "CAFÉ MODE") }, mode.ErrAmbiguousName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.makeMode()
			checkError(t, err, tt.category)
			assert.Zero(t, got)
		})
	}
}

func firstOf[A, B any](a A, _ B) A  { return a }
func secondOf[A, B any](_ A, b B) B { return b }

func TestCreateFromIntervals(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, adj, own, cycle, code string
	}{
		{"ionian", ionian, "P1 M2 M3 P4 P5 M6 M7", "P8", "ionian"},
		{"singleton", "P8", "P1", "P8", ""},
		{"split octave", "P5 P4", "P1 P5", "P8", ""},
		{"tritone pair", "A4 d5", "P1 A4", "P8", ""},
		{"blues", "m3 M2 A1 m2 m3 M2", "P1 m3 P4 A4 P5 m7", "P8", "blues"},
		{"flat fifth blues", "m3 M2 m2 A1 m3 M2", "P1 m3 P4 d5 P5 m7", "P8", "blues-b5"},
		{"zero numeric edge", "d2 A7", "P1 d2", "P8", ""},
		{"short cycle", "M2 M2 m2", "P1 M2 M3", "P4", ""},
		{"wide cycle", "M9 m7", "P1 M9", "P15", ""},
		{"open", "m3 M2", "P1 m3 P4", "", ""},
		{"open minimum", "M2", "P1 M2", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			adj, own := intervals(tt.adj), intervals(tt.own)
			var m, fromOwn mode.Mode
			var err error
			if tt.cycle == "" {
				m, err = mode.OpenFromIntervals(adj)
				require.NoError(t, err)
				fromOwn, err = mode.OpenFromDegreeIntervals(own)
			} else {
				m, err = mode.ClosedFromIntervals(adj)
				require.NoError(t, err)
				fromOwn, err = mode.ClosedFromDegreeIntervals(own, interval.MustParse(tt.cycle))
			}
			require.NoError(t, err)
			assert.Equal(t, tt.own, intervalText(m.DegreeIntervals()))
			assert.True(t, m.Equal(fromOwn))
			assert.Equal(t, len(own), m.Len())
			cycle, ok := m.Cycle()
			assert.Equal(t, tt.cycle != "", ok)
			if ok {
				assert.Equal(t, tt.cycle, cycle.String())
			} else {
				assert.Zero(t, cycle)
			}
			assert.Equal(t, tt.cycle == "P8", m.IsOctave())
			if tt.code != "" {
				assert.True(t, m.Equal(mode.MustFromCode(tt.code)))
			}
			assert.False(t, m.HasOrigin())
			adj[0], own[0] = interval.Interval{}, interval.Interval{}
			assert.Equal(t, tt.adj, intervalText(m.AdjacentIntervals()))
			assert.Equal(t, tt.own, intervalText(fromOwn.DegreeIntervals()))
		})
	}
	for _, tt := range []struct {
		name            string
		call            func() (mode.Mode, error)
		category, cause error
	}{
		{"nil adjacency", func() (mode.Mode, error) { return mode.ClosedFromIntervals(nil) }, mode.ErrEmptyInput, template.ErrEmptyInput},
		{"empty open adjacency", func() (mode.Mode, error) { return mode.OpenFromIntervals([]interval.Interval{}) }, mode.ErrEmptyInput, template.ErrEmptyInput},
		{"empty own", func() (mode.Mode, error) { return mode.ClosedFromDegreeIntervals(nil, interval.MustParse("P8")) }, mode.ErrEmptyInput, template.ErrEmptyInput},
		{"one open place", func() (mode.Mode, error) { return mode.OpenFromDegreeIntervals(intervals("P1")) }, mode.ErrEmptyInput, template.ErrEmptyInput},
		{"unison edge", func() (mode.Mode, error) { return mode.ClosedFromIntervals(intervals("M2 P1 M7")) }, mode.ErrInvalidStructure, template.ErrInvalidInterval},
		{"descending edge", func() (mode.Mode, error) { return mode.OpenFromIntervals(intervals("-M2 M2")) }, mode.ErrInvalidStructure, template.ErrInvalidInterval},
		{"unset in final edge", func() (mode.Mode, error) {
			return mode.ClosedFromIntervals(append(intervals("P1"), interval.Interval{}))
		}, mode.ErrUnset, template.ErrUnset},
		{"unset cycle before empty input", func() (mode.Mode, error) { return mode.ClosedFromDegreeIntervals(nil, interval.Interval{}) }, mode.ErrUnset, template.ErrUnset},
		{"first not unison", func() (mode.Mode, error) { return mode.OpenFromDegreeIntervals(intervals("M2 M3")) }, mode.ErrInvalidStructure, template.ErrInvalidInterval},
		{"places decrease", func() (mode.Mode, error) { return mode.OpenFromDegreeIntervals(intervals("P1 M3 M2")) }, mode.ErrInvalidStructure, template.ErrInvalidInterval},
		{"cycle equals final place", func() (mode.Mode, error) {
			return mode.ClosedFromDegreeIntervals(intervals("P1 P5 P8"), interval.MustParse("P8"))
		}, mode.ErrInvalidStructure, template.ErrInvalidInterval},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.call()
			checkError(t, err, tt.category, tt.cause)
			assert.Zero(t, got)
		})
	}
}

type changeableSource struct {
	current atomic.Pointer[settings.Settings]
}

func (s *changeableSource) Settings() *settings.Settings { return s.current.Load() }

func TestCreateFromNotes(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ names, want string }{
		{"C D E♭ F G A♭ B", "M2 m2 M2 M2 m2 A2 m2"},
		{"G♯ A♯ B C♯ D♯ E F𝄪", "M2 m2 M2 M2 m2 A2 m2"},
		{"B C D", "m2 M2 M6"},
		{"C D C♭", "M2 d7 A1"},
	} {
		t.Run(tt.names, func(t *testing.T) {
			input := notes(t, tt.names, cents.MustFromInt(30), nil)
			m, err := mode.FromNotesInOctave(input)
			require.NoError(t, err)
			checkMode(t, m, tt.want)
			checkMode(t, mode.MustClosedFromNotes(input, interval.MustParse("P8")), tt.want)
			input[0] = note.Note{}
			checkMode(t, m, tt.want)
		})
	}
	mixed := notes(t, "D E F G A B C", cents.MustFromInt(30), settings.Standard())
	d4, err := note.FromASPN(spelling.MustParse("D"), 4, cents.Cents{}, nil)
	require.NoError(t, err)
	c2, err := note.FromASPN(spelling.MustParse("C"), 2, cents.Cents{}, nil)
	require.NoError(t, err)
	mixed[0], mixed[6] = d4, c2
	checkMode(t, mode.MustFromNotesInOctave(mixed), dorian)
	// The names already exist even when their source temporarily has no settings.
	src := new(changeableSource)
	src.current.Store(settings.Standard())
	bound := notes(t, "D E F G A B C", cents.Cents{}, src)
	src.current.Store(nil)
	checkMode(t, mode.MustFromNotesInOctave(bound), dorian)

	checkMode(t, mode.MustClosedFromNotes(notes(t, "G A B", cents.Cents{}, nil), interval.MustParse("P4")), "M2 M2 m2")
	checkMode(t, mode.MustOpenFromNotes(notes(t, "E G A", cents.Cents{}, nil)), "m3 M2 |")
	for _, tt := range []struct {
		names    string
		category error
	}{
		{"C E D", mode.ErrInvalidStructure}, {"C C", mode.ErrInvalidStructure}, {"", mode.ErrEmptyInput},
	} {
		got, err := mode.FromNotesInOctave(notes(t, tt.names, cents.Cents{}, nil))
		checkError(t, err, tt.category)
		assert.Zero(t, got)
	}
	got, err := mode.OpenFromNotes(notes(t, "C", cents.Cents{}, nil))
	checkError(t, err, mode.ErrEmptyInput)
	assert.Zero(t, got)
	got, err = mode.OpenFromNotes(append(notes(t, "C C", cents.Cents{}, nil), note.Note{}))
	checkError(t, err, mode.ErrUnset)
	assert.Zero(t, got)

	magnitude := alteration.MaxMagnitude
	low, err := note.FromSpelling(spelling.MustNew(letter.C, alteration.MustFromInt(-magnitude)), cents.Cents{}, nil)
	require.NoError(t, err)
	high, err := note.FromSpelling(spelling.MustNew(letter.D, alteration.MustFromInt(magnitude)), cents.Cents{}, nil)
	require.NoError(t, err)
	got, err = mode.OpenFromNotes([]note.Note{low, high})
	checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Zero(t, got)

}

func TestCreateFromNumbers(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, want string
		sys        *system.System
		steps      []system.Step
		start      string
		plan       mode.LetterPlan
	}{
		{"traditional", ionian, system.Traditional(), []system.Step{2, 2, 1, 2, 2, 2, 1}, "", mode.LetterPlan{}},
		{"harmonic minor", "M2 m2 M2 M2 m2 A2 m2", system.Traditional(), []system.Step{2, 1, 2, 2, 1, 3, 1}, "", mode.LetterPlan{}},
		{"19 positions", ionian, system.Positions19(), []system.Step{3, 3, 2, 3, 3, 3, 2}, "", mode.LetterPlan{}},
		{"24 positions", "M2 n2 n2 M2 M2 n2 n2", system.QuarterTones24(), []system.Step{4, 3, 3, 4, 4, 3, 3}, "", mode.LetterPlan{}},
		{"irregular with start", "m2 A2 m2 M2 M2 M2 m2", irregularSystem(), []system.Step{2, 2, 1, 2, 2, 2, 1}, "C", mode.LetterPlan{}},
		{"pentatonic plan", "M2 M2 m3 M2 m3", system.Traditional(), []system.Step{2, 2, 3, 2, 3}, "", mode.LetterPlan{Degrees: []int{1, 2, 3, 5, 6}, Cycle: 8}},
		{"reused letter", "m3 M2 A1 m2 m3 M2", system.Traditional(), []system.Step{3, 2, 1, 1, 3, 2}, "", mode.LetterPlan{Degrees: []int{1, 3, 4, 4, 5, 7}, Cycle: 8}},
		{"explicit short cycle", "M2 M2 m2", system.Traditional(), []system.Step{2, 2, 1}, "G", mode.LetterPlan{Degrees: []int{1, 2, 3}, Cycle: 4}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			steps, plan := slices.Clone(tt.steps), tt.plan
			plan.Degrees = slices.Clone(plan.Degrees)
			var m mode.Mode
			var err error
			switch {
			case tt.start != "" && plan.Cycle != 0:
				m, err = mode.FromNumbersFromWithPlan(tt.sys, steps, spelling.MustParse(tt.start), plan)
			case tt.start != "":
				m, err = mode.FromNumbersFrom(tt.sys, steps, spelling.MustParse(tt.start))
			case plan.Cycle != 0:
				m, err = mode.FromNumbersWithPlan(tt.sys, steps, plan)
			default:
				m, err = mode.FromNumbers(tt.sys, steps)
			}
			require.NoError(t, err)
			assert.Equal(t, tt.steps, steps)
			assert.Equal(t, tt.plan.Degrees, plan.Degrees)
			clear(steps)
			clear(plan.Degrees)
			checkMode(t, m, tt.want)
			assert.True(t, m.IsClosed())
			assert.False(t, m.HasOrigin())
		})
	}
	sys := system.Traditional()
	steps := []system.Step{2, 2, 1, 2, 2, 2, 1}
	plan := mode.LetterPlan{Degrees: []int{1, 2}, Cycle: 3}
	for _, tt := range []struct {
		name  string
		call  func() (mode.Mode, error)
		cause error
	}{
		{"nil system first", func() (mode.Mode, error) { return mode.FromNumbers(nil, nil) }, mode.ErrUnset},
		{"zero system", func() (mode.Mode, error) { return mode.FromNumbers(&system.System{}, steps) }, mode.ErrUnset},
		{"unset start", func() (mode.Mode, error) { return mode.FromNumbersFrom(sys, steps, spelling.Spelling{}) }, mode.ErrUnset},
		{"nil input", func() (mode.Mode, error) { return mode.FromNumbers(sys, nil) }, mode.ErrEmptyInput},
		{"negative number before count", func() (mode.Mode, error) { return mode.FromNumbers(sys, []system.Step{-1}) }, mode.ErrInvalidStructure},
		{"zero step", func() (mode.Mode, error) { return mode.FromNumbers(sys, []system.Step{0, 2, 1, 2, 2, 2, 1}) }, mode.ErrInvalidStructure},
		{"wrong count before octave", func() (mode.Mode, error) { return mode.FromNumbers(sys, []system.Step{2, 2, 3, 2, 3}) }, mode.ErrWrongDegreeCount},
		{"wrong octave", func() (mode.Mode, error) { return mode.FromNumbers(system.Positions19(), steps) }, mode.ErrSystemOctaveRequired},
		{"plan length", func() (mode.Mode, error) { return mode.FromNumbersWithPlan(sys, steps, plan) }, mode.ErrWrongDegreeCount},
		{"bad plan before negative step", func() (mode.Mode, error) {
			return mode.FromNumbersWithPlan(sys, []system.Step{-1, 1}, mode.LetterPlan{Degrees: []int{2, 3}, Cycle: 4})
		}, mode.ErrInvalidArgument},
		{"decreasing plan", func() (mode.Mode, error) {
			return mode.FromNumbersWithPlan(sys, []system.Step{2, 2, 1}, mode.LetterPlan{Degrees: []int{1, 3, 2}, Cycle: 4})
		}, mode.ErrInvalidArgument},
		{"cycle too small", func() (mode.Mode, error) {
			return mode.FromNumbersWithPlan(sys, []system.Step{2, 2}, mode.LetterPlan{Degrees: []int{1, 2}, Cycle: 1})
		}, mode.ErrInvalidArgument},
		{"irregular without start", func() (mode.Mode, error) { return mode.FromNumbers(irregularSystem(), steps) }, interval.ErrStartRequired},
		{"inexact accumulated step", func() (mode.Mode, error) { return mode.FromNumbersWithPlan(sys, []system.Step{math.MaxInt, 1}, plan) }, mode.ErrOutOfRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.call()
			assert.ErrorIs(t, err, tt.cause)
			assert.Zero(t, got)
		})
	}
}

func TestCreationDegreeLimit(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 65536, mode.MaxDegrees)
	assert.Equal(t, template.MaxPlaces, mode.MaxDegrees)
	edges := slices.Repeat(intervals("M2"), mode.MaxDegrees+1)
	for _, tt := range []struct {
		name     string
		count    int
		makeMode func([]interval.Interval) (mode.Mode, error)
	}{
		{"closed", mode.MaxDegrees, mode.ClosedFromIntervals},
		{"open", mode.MaxDegrees - 1, mode.OpenFromIntervals},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.makeMode(edges[:tt.count])
			require.NoError(t, err)
			assert.Equal(t, mode.MaxDegrees, m.Len())
			got, err := tt.makeMode(edges[:tt.count+1])
			checkError(t, err, mode.ErrInvalidArgument, template.ErrInvalidArgument)
			assert.Zero(t, got)
		})
	}
	got, err := mode.FromNumbers(system.Traditional(), slices.Repeat([]system.Step{1}, mode.MaxDegrees+1))
	checkError(t, err, mode.ErrInvalidArgument)
	assert.Zero(t, got)
}
