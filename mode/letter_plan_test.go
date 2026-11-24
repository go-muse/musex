package mode_test

import (
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the suggested letter plan and of the closed mode built by a
// spelling rule (issue #146, item 1; business-spec, KP-15 and F-12).
// Spellings are written in the machine form of package spelling:
// "E-1" is E flat, "F+2" is F double sharp, "E-1/2" is E lowered by 1/2.

// stepList converts plain integers to system steps.
func stepList(values ...int) []system.Step {
	result := make([]system.Step, len(values))
	for i, v := range values {
		result[i] = system.Step(v)
	}
	return result
}

// ones returns count steps of 1.
func ones(count int) []system.Step {
	result := make([]system.Step, count)
	for i := range result {
		result[i] = 1
	}
	return result
}

// naturalsOnlySystem is T0 of the business spec: the traditional layout
// where only the natural names are admissible.
func naturalsOnlySystem() *system.System {
	return system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(0, 0), "T0: naturals only")
}

// singleAccidentalSystem is T1 of the business spec: the traditional
// layout with the alteration values -1, 0 and +1.
func singleAccidentalSystem() *system.System {
	return system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(-1, 1), "T1: single accidentals")
}

// chromaticMajorUp is the rule "chromatic scale" of C major, ascending.
func chromaticMajorUp() system.SpellingRule {
	return system.MustChromaticScale(spelling.MustParse("C"), system.Major, system.Ascending)
}

// mappedRule spells the listed class numbers with fixed names and every
// other class with fallback. A fallback failure gives an unset name, which
// system reports as a bad custom spelling.
func mappedRule(names map[int]string, fallback system.SpellingRule) system.SpellingRule {
	return system.MustCustomSpelling(func(s *system.System, c pitchclass.Class) spelling.Spelling {
		if name, ok := names[c.Number()]; ok {
			return spelling.MustParse(name)
		}
		chosen, err := s.Spell(c, fallback)
		if err != nil {
			return spelling.Spelling{}
		}
		return chosen
	})
}

// degreeNames returns the names of the notes of m built from start in sys.
func degreeNames(t *testing.T, m mode.Mode, sys *system.System, start string) string {
	t.Helper()
	first, err := note.FromSpelling(spelling.MustParse(start), cents.Cents{}, nil)
	require.NoError(t, err)
	built, err := m.NotesFromIn(sys, first)
	require.NoError(t, err)
	return noteNames(built)
}

// allModeErrors lists every sentinel error of package mode.
func allModeErrors() []error {
	return append(modeErrors(), mode.ErrIndex, mode.ErrInsufficientNotes, mode.ErrNonMonotone)
}

// spellingCauses lists the causes of package system that a spelling rule
// can report.
func spellingCauses() []error {
	return []error{system.ErrNotApplicable, system.ErrInadmissible, system.ErrNoSpelling, system.ErrBadSpelling}
}

// checkCategory requires the mode category, the listed causes and no other
// mode sentinel in the chain of err.
func checkCategory(t *testing.T, err, category error, causes ...error) {
	t.Helper()
	require.ErrorIs(t, err, category)
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
	for _, other := range allModeErrors() {
		if other != category {
			assert.NotErrorIs(t, err, other)
		}
	}
}

// checkForeignCause requires the listed causes of other packages and no
// sentinel of package mode in the chain of err.
func checkForeignCause(t *testing.T, err error, causes ...error) {
	t.Helper()
	require.Error(t, err)
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
	for _, sentinel := range allModeErrors() {
		assert.NotErrorIs(t, err, sentinel)
	}
}

type ruleRow struct {
	name     string
	sys      *system.System
	start    string
	adjacent []system.Step
	rule     system.SpellingRule
	names    string
	plan     []int
	want     string
	// vsDefault compares the result with FromNumbersFrom without a plan:
	// "same" requires equal modes, "differs" requires different ones.
	vsDefault string
}

func ruleRows() []ruleRow {
	traditional := system.Traditional()
	irregular := irregularSystem()
	sharps, flats := system.Sharps(), system.Flats()
	twelve := ones(12)
	return []ruleRow{
		{name: "C 3 2 2 3 2 flats", sys: traditional, start: "C", adjacent: stepList(3, 2, 2, 3, 2), rule: flats,
			names: "C E-1 F G B-1", plan: []int{1, 3, 4, 5, 7}, want: "m3 M2 M2 m3 M2"},
		{name: "C 3 2 2 3 2 sharps", sys: traditional, start: "C", adjacent: stepList(3, 2, 2, 3, 2), rule: sharps,
			names: "C D+1 F G A+1", plan: []int{1, 2, 4, 5, 6}, want: "A2 d3 M2 A2 d3"},
		{name: "C 2 2 3 2 3 sharps", sys: traditional, start: "C", adjacent: stepList(2, 2, 3, 2, 3), rule: sharps,
			names: "C D E G A", plan: []int{1, 2, 3, 5, 6}, want: "M2 M2 m3 M2 m3"},
		{name: "C 2 2 3 2 3 flats", sys: traditional, start: "C", adjacent: stepList(2, 2, 3, 2, 3), rule: flats,
			names: "C D E G A", plan: []int{1, 2, 3, 5, 6}, want: "M2 M2 m3 M2 m3"},
		{name: "C 3 2 1 1 3 2 sharps", sys: traditional, start: "C", adjacent: stepList(3, 2, 1, 1, 3, 2), rule: sharps,
			names: "C D+1 F F+1 G A+1", plan: []int{1, 2, 4, 4, 5, 6}, want: "A2 d3 A1 m2 A2 d3"},
		{name: "C 3 2 1 1 3 2 flats", sys: traditional, start: "C", adjacent: stepList(3, 2, 1, 1, 3, 2), rule: flats,
			names: "C E-1 F G-1 G B-1", plan: []int{1, 3, 4, 5, 5, 7}, want: "m3 M2 m2 A1 m3 M2"},
		{name: "C major sharps", sys: traditional, start: "C", adjacent: stepList(2, 2, 1, 2, 2, 2, 1), rule: sharps,
			names: "C D E F G A B", plan: []int{1, 2, 3, 4, 5, 6, 7}, want: ionian, vsDefault: "same"},
		{name: "C major flats", sys: traditional, start: "C", adjacent: stepList(2, 2, 1, 2, 2, 2, 1), rule: flats,
			names: "C D E F G A B", plan: []int{1, 2, 3, 4, 5, 6, 7}, want: ionian, vsDefault: "same"},
		{name: "A harmonic minor sharps", sys: traditional, start: "A", adjacent: stepList(2, 1, 2, 2, 1, 3, 1), rule: sharps,
			names: "A B C D E F G+1", plan: []int{1, 2, 3, 4, 5, 6, 7}, want: "M2 m2 M2 M2 m2 A2 m2", vsDefault: "same"},
		{name: "A harmonic minor flats", sys: traditional, start: "A", adjacent: stepList(2, 1, 2, 2, 1, 3, 1), rule: flats,
			names: "A B C D E F A-1", plan: []int{1, 2, 3, 4, 5, 6, 8}, want: "M2 m2 M2 M2 m2 m3 A1", vsDefault: "differs"},
		{name: "D flat pentatonic flats", sys: traditional, start: "D-1", adjacent: stepList(2, 2, 3, 2, 3), rule: flats,
			names: "D-1 E-1 F A-1 B-1", plan: []int{1, 2, 3, 5, 6}, want: "M2 M2 m3 M2 m3"},
		{name: "D flat keeps its name under sharps", sys: traditional, start: "D-1", adjacent: stepList(2, 2, 3, 2, 3), rule: sharps,
			names: "D-1 D+1 F G+1 A+1", plan: []int{1, 1, 3, 4, 5}, want: "AA1 d3 A2 M2 dd4"},
		{name: "F sharp pentatonic sharps", sys: traditional, start: "F+1", adjacent: stepList(2, 2, 3, 2, 3), rule: sharps,
			names: "F+1 G+1 A+1 C+1 D+1", plan: []int{1, 2, 3, 5, 6}, want: "M2 M2 m3 M2 m3"},
		{name: "F sharp pentatonic flats", sys: traditional, start: "F+1", adjacent: stepList(2, 2, 3, 2, 3), rule: flats,
			names: "F+1 A-1 B-1 D-1 E-1", plan: []int{1, 3, 4, 6, 7}, want: "d3 M2 m3 M2 A2"},
		{name: "F sharp 11 1 sharps gives size 8", sys: traditional, start: "F+1", adjacent: stepList(11, 1), rule: sharps,
			names: "F+1 F", plan: []int{1, 8}, want: "d8 A1"},
		{name: "F sharp 11 1 flats gives size 8", sys: traditional, start: "F+1", adjacent: stepList(11, 1), rule: flats,
			names: "F+1 F", plan: []int{1, 8}, want: "d8 A1"},
		{name: "C 6 6 sharps", sys: traditional, start: "C", adjacent: stepList(6, 6), rule: sharps,
			names: "C F+1", plan: []int{1, 4}, want: "A4 d5"},
		{name: "C 6 6 flats", sys: traditional, start: "C", adjacent: stepList(6, 6), rule: flats,
			names: "C G-1", plan: []int{1, 5}, want: "d5 A4"},
		{name: "one degree sharps", sys: traditional, start: "C", adjacent: stepList(12), rule: sharps,
			names: "C", plan: []int{1}, want: "P8"},
		{name: "one degree flats", sys: traditional, start: "C", adjacent: stepList(12), rule: flats,
			names: "C", plan: []int{1}, want: "P8"},
		{name: "one degree chromatic scale", sys: traditional, start: "C", adjacent: stepList(12), rule: chromaticMajorUp(),
			names: "C", plan: []int{1}, want: "P8"},
		{name: "one degree custom rule", sys: traditional, start: "C", adjacent: stepList(12), rule: mappedRule(nil, sharps),
			names: "C", plan: []int{1}, want: "P8"},
		{name: "twelve degrees sharps", sys: traditional, start: "C", adjacent: twelve, rule: sharps,
			names: "C C+1 D D+1 E F F+1 G G+1 A A+1 B", plan: []int{1, 1, 2, 2, 3, 4, 4, 5, 5, 6, 6, 7},
			want: "A1 m2 A1 m2 m2 A1 m2 A1 m2 A1 m2 m2"},
		{name: "twelve degrees flats", sys: traditional, start: "C", adjacent: twelve, rule: flats,
			names: "C D-1 D E-1 E F G-1 G A-1 A B-1 B", plan: []int{1, 2, 2, 3, 3, 4, 5, 5, 6, 6, 7, 7},
			want: "m2 A1 m2 A1 m2 m2 A1 m2 A1 m2 A1 m2"},
		{name: "twelve degrees chromatic scale", sys: traditional, start: "C", adjacent: twelve, rule: chromaticMajorUp(),
			names: "C C+1 D D+1 E F F+1 G G+1 A B-1 B", plan: []int{1, 1, 2, 2, 3, 4, 4, 5, 5, 6, 7, 7},
			want: "A1 m2 A1 m2 m2 A1 m2 A1 m2 m2 A1 m2"},
		{name: "three degrees on one letter", sys: traditional, start: "F-1", adjacent: stepList(1, 1, 10), rule: sharps,
			names: "F-1 F F+1", plan: []int{1, 1, 1}, want: "A1 A1 dd8"},
		{name: "B sharp 1 11 sharps", sys: traditional, start: "B+1", adjacent: stepList(1, 11), rule: sharps,
			names: "B+1 C+1", plan: []int{1, 2}, want: "m2 M7"},
		{name: "B sharp 1 11 flats", sys: traditional, start: "B+1", adjacent: stepList(1, 11), rule: flats,
			names: "B+1 D-1", plan: []int{1, 3}, want: "dd3 AA6"},
		{name: "G double flat 2 10 sharps", sys: traditional, start: "G-2", adjacent: stepList(2, 10), rule: sharps,
			names: "G-2 G", plan: []int{1, 1}, want: "AA1 dd8"},
		{name: "G double flat 2 10 flats", sys: traditional, start: "G-2", adjacent: stepList(2, 10), rule: flats,
			names: "G-2 G", plan: []int{1, 1}, want: "AA1 dd8"},
		{name: "G double flat 1 11 flats", sys: traditional, start: "G-2", adjacent: stepList(1, 11), rule: flats,
			names: "G-2 G-1", plan: []int{1, 1}, want: "A1 d8"},
		{name: "F double sharp 11 1 sharps", sys: traditional, start: "F+2", adjacent: stepList(11, 1), rule: sharps,
			names: "F+2 F+1", plan: []int{1, 8}, want: "d8 A1"},
		{name: "naturals only C 2 2 3 2 3 sharps", sys: naturalsOnlySystem(), start: "C", adjacent: stepList(2, 2, 3, 2, 3), rule: sharps,
			names: "C D E G A", plan: []int{1, 2, 3, 5, 6}, want: "M2 M2 m3 M2 m3"},
		{name: "19-position sharps", sys: system.Positions19(), start: "C", adjacent: stepList(5, 3, 3, 5, 3), rule: sharps,
			names: "C E-1 F G B-1", plan: []int{1, 3, 4, 5, 7}, want: "m3 M2 M2 m3 M2"},
		{name: "19-position flats", sys: system.Positions19(), start: "C", adjacent: stepList(5, 3, 3, 5, 3), rule: flats,
			names: "C E-1 F G B-1", plan: []int{1, 3, 4, 5, 7}, want: "m3 M2 M2 m3 M2"},
		{name: "19-position one degree ignores an inapplicable rule", sys: system.Positions19(), start: "C", adjacent: stepList(19),
			rule: chromaticMajorUp(), names: "C", plan: []int{1}, want: "P8"},
		{name: "31 sharps and flats C 1 30 flats", sys: system.SharpsAndFlats31(), start: "C", adjacent: stepList(1, 30), rule: flats,
			names: "C D-2", plan: []int{1, 2}, want: "d2 A7"},
		{name: "24-position sharps", sys: system.QuarterTones24(), start: "C", adjacent: stepList(4, 3, 3, 4, 4, 3, 3), rule: sharps,
			names: "C D E-1/2 F G A B-1/2", plan: []int{1, 2, 3, 4, 5, 6, 7}, want: "M2 n2 n2 M2 M2 n2 n2", vsDefault: "same"},
		{name: "24-position flats", sys: system.QuarterTones24(), start: "C", adjacent: stepList(4, 3, 3, 4, 4, 3, 3), rule: flats,
			names: "C D E-1/2 F G A B-1/2", plan: []int{1, 2, 3, 4, 5, 6, 7}, want: "M2 n2 n2 M2 M2 n2 n2", vsDefault: "same"},
		{name: "irregular from C flats", sys: irregular, start: "C", adjacent: stepList(3, 2, 2, 3, 2), rule: flats,
			names: "C D F G B-1", plan: []int{1, 2, 4, 5, 7}, want: "M2 m3 M2 m3 M2"},
		{name: "irregular from C sharps", sys: irregular, start: "C", adjacent: stepList(3, 2, 2, 3, 2), rule: sharps,
			names: "C D F G A+1", plan: []int{1, 2, 4, 5, 6}, want: "M2 m3 M2 A2 d3"},
		{name: "irregular from D flats", sys: irregular, start: "D", adjacent: stepList(2, 2, 3, 2, 3), rule: flats,
			names: "D F G B-1 C", plan: []int{1, 3, 4, 6, 7}, want: "m3 M2 m3 M2 M2"},
		{name: "irregular from D sharps", sys: irregular, start: "D", adjacent: stepList(2, 2, 3, 2, 3), rule: sharps,
			names: "D F G A+1 C", plan: []int{1, 3, 4, 5, 7}, want: "m3 M2 A2 d3 M2"},
	}
}

func TestSuggestLetterPlanAndBuildByRule(t *testing.T) {
	t.Parallel()
	for _, tt := range ruleRows() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start := spelling.MustParse(tt.start)
			input := slices.Clone(tt.adjacent)
			want := mode.LetterPlan{Degrees: tt.plan, Cycle: 8}

			plan, err := mode.SuggestLetterPlan(tt.sys, input, start, tt.rule)
			require.NoError(t, err)
			assert.Equal(t, want, plan)

			m, err := mode.FromNumbersFromWithRule(tt.sys, input, start, tt.rule)
			require.NoError(t, err)
			assert.Equal(t, tt.adjacent, input, "the input numbers must not change")
			expected := mode.MustParse(tt.want)
			assert.True(t, m.Equal(expected), "got %s, want %s", m, expected)
			assert.True(t, m.IsClosed())
			assert.True(t, m.IsOctave())
			assert.False(t, m.HasOrigin())
			assert.False(t, m.Changed())

			// F-12: the rule result equals the explicit plan result, reads back
			// the same plan, has the passed numbers and the rule's names.
			byPlan, err := mode.FromNumbersFromWithPlan(tt.sys, tt.adjacent, start, want)
			require.NoError(t, err)
			assert.True(t, m.Equal(byPlan))
			read, closed := m.LetterPlan()
			assert.True(t, closed)
			assert.Equal(t, want, read)
			assert.Equal(t, tt.adjacent, m.MustNumbersFrom(tt.sys, start).Adjacent)
			assert.Equal(t, tt.names, degreeNames(t, m, tt.sys, tt.start))

			switch tt.vsDefault {
			case "same":
				assert.True(t, m.Equal(mode.MustFromNumbersFrom(tt.sys, tt.adjacent, start)))
			case "differs":
				// R-22: with seven degrees the rule is still applied; the default
				// plan of one letter per degree is not substituted.
				byDefault := mode.MustFromNumbersFrom(tt.sys, tt.adjacent, start)
				assert.False(t, m.Equal(byDefault))
				defaultPlan, _ := byDefault.LetterPlan()
				assert.NotEqual(t, plan.Degrees, defaultPlan.Degrees)
			}
		})
	}
}

type ruleFailure struct {
	name     string
	sys      *system.System
	start    spelling.Spelling
	adjacent []system.Step
	rule     system.SpellingRule
	// category is the mode sentinel; nil means a direct cause of system
	// without any mode sentinel.
	category error
	causes   []error
	// noSpellingCause additionally excludes every spelling cause of system;
	// it is set for an incorrect record found after all names are chosen.
	noSpellingCause bool
}

func ruleFailures() []ruleFailure {
	traditional := system.Traditional()
	t0 := naturalsOnlySystem()
	t1 := singleAccidentalSystem()
	c := spelling.MustParse("C")
	cDoubleSharp := spelling.MustParse("C+2")
	sharps, flats := system.Sharps(), system.Flats()
	tooMany := ones(mode.MaxDegrees + 1)
	tooMany[0] = 0
	return []ruleFailure{
		{name: "sum below N", sys: traditional, start: c, adjacent: stepList(3, 2, 2, 3), rule: sharps,
			category: mode.ErrSystemOctaveRequired},
		{name: "sum equal to the traditional N in the 19-position system", sys: system.Positions19(), start: c,
			adjacent: stepList(2, 2, 3, 2, 3), rule: sharps, category: mode.ErrSystemOctaveRequired},
		{name: "exact sum is not wrapped modulo 2^64", sys: traditional, start: c,
			adjacent: stepList(math.MaxInt, math.MaxInt, 14), rule: sharps, category: mode.ErrSystemOctaveRequired},
		{name: "zero number", sys: traditional, start: c, adjacent: stepList(3, 0, 4, 3, 2), rule: sharps,
			category: mode.ErrInvalidStructure},
		{name: "negative number", sys: traditional, start: c, adjacent: stepList(3, -2, 6, 3, 2), rule: sharps,
			category: mode.ErrInvalidStructure},
		{name: "nil numbers", sys: traditional, start: c, adjacent: nil, rule: sharps, category: mode.ErrEmptyInput},
		{name: "empty numbers", sys: traditional, start: c, adjacent: []system.Step{}, rule: sharps, category: mode.ErrEmptyInput},
		{name: "nil system", sys: nil, start: c, adjacent: stepList(12), rule: sharps, category: mode.ErrUnset},
		{name: "unset system", sys: new(system.System), start: c, adjacent: stepList(12), rule: sharps, category: mode.ErrUnset},
		{name: "unset start", sys: traditional, start: spelling.Spelling{}, adjacent: stepList(12), rule: sharps, category: mode.ErrUnset},
		{name: "unset rule", sys: traditional, start: c, adjacent: stepList(12), rule: system.SpellingRule{}, category: mode.ErrUnset},
		{name: "unset rule before missing numbers", sys: traditional, start: c, adjacent: nil, rule: system.SpellingRule{},
			category: mode.ErrUnset},
		{name: "nil system before missing numbers", sys: nil, start: c, adjacent: nil, rule: sharps, category: mode.ErrUnset},
		{name: "unset start before missing numbers", sys: traditional, start: spelling.Spelling{}, adjacent: nil, rule: sharps,
			category: mode.ErrUnset},
		{name: "too many numbers before reading them", sys: traditional, start: c, adjacent: tooMany, rule: sharps,
			category: mode.ErrInvalidArgument},
		{name: "the degree limit itself is allowed", sys: traditional, start: c, adjacent: ones(mode.MaxDegrees), rule: sharps,
			category: mode.ErrSystemOctaveRequired},
		{name: "number checked before the sum", sys: traditional, start: c, adjacent: stepList(0, 5), rule: sharps,
			category: mode.ErrInvalidStructure},
		{name: "no admissible spelling of a class", sys: t0, start: c, adjacent: stepList(3, 2, 2, 3, 2), rule: sharps,
			causes: []error{system.ErrNoSpelling}},
		{name: "inadmissible start", sys: t1, start: cDoubleSharp, adjacent: stepList(2, 2, 3, 2, 3), rule: sharps,
			causes: []error{system.ErrInadmissible}},
		{name: "inadmissible start with one degree", sys: t1, start: cDoubleSharp, adjacent: stepList(12), rule: sharps,
			causes: []error{system.ErrInadmissible}},
		{name: "rule not applicable", sys: system.Positions19(), start: c, adjacent: stepList(5, 3, 3, 5, 3), rule: chromaticMajorUp(),
			causes: []error{system.ErrNotApplicable}},
		{name: "start checked before the applicability of the rule", sys: system.Positions19(), start: spelling.MustParse("C+1/2"),
			adjacent: stepList(5, 3, 3, 5, 3), rule: chromaticMajorUp(), causes: []error{system.ErrInadmissible}},
		{name: "number checked before the names", sys: t0, start: c, adjacent: stepList(3, 0, 4, 3, 2), rule: sharps,
			category: mode.ErrInvalidStructure},
		{name: "sum checked before the start", sys: t1, start: cDoubleSharp, adjacent: stepList(3, 2, 2, 3), rule: sharps,
			category: mode.ErrSystemOctaveRequired},
		{name: "bad custom name of degree 3 before the incorrect degree 2", sys: traditional, start: spelling.MustParse("G-2"),
			adjacent: stepList(1, 1, 10), rule: mappedRule(map[int]string{6: "F+1", 7: "A"}, sharps),
			causes: []error{system.ErrBadSpelling}},
		{name: "degree written below the first degree", sys: traditional, start: spelling.MustParse("G-2"), adjacent: stepList(1, 11),
			rule: sharps, category: mode.ErrInvalidStructure, noSpellingCause: true},
		{name: "degree written above the next first degree", sys: traditional, start: spelling.MustParse("F+2"), adjacent: stepList(11, 1),
			rule: flats, category: mode.ErrInvalidStructure, noSpellingCause: true},
		{name: "degree written below the previous degree", sys: traditional, start: c, adjacent: stepList(4, 1, 7),
			rule: mappedRule(map[int]string{4: "F-1", 5: "E+1"}, sharps), category: mode.ErrInvalidStructure, noSpellingCause: true},
		{name: "natural start in 31 sharps and flats", sys: system.SharpsAndFlats31(), start: c, adjacent: stepList(1, 30),
			rule: sharps, category: mode.ErrInvalidStructure, noSpellingCause: true},
		{name: "decreasing sizes in 31 sharps and flats", sys: system.SharpsAndFlats31(), start: c, adjacent: stepList(1, 1, 29),
			rule: flats, category: mode.ErrInvalidStructure, noSpellingCause: true},
	}
}

func checkRuleFailure(t *testing.T, err error, tt ruleFailure) {
	t.Helper()
	if tt.category == nil {
		checkForeignCause(t, err, tt.causes...)
		return
	}
	checkCategory(t, err, tt.category, tt.causes...)
	if tt.noSpellingCause {
		for _, cause := range spellingCauses() {
			assert.NotErrorIs(t, err, cause)
		}
	}
}

func TestSuggestLetterPlanAndBuildByRuleFailures(t *testing.T) {
	t.Parallel()
	for _, tt := range ruleFailures() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, err := mode.SuggestLetterPlan(tt.sys, tt.adjacent, tt.start, tt.rule)
			checkRuleFailure(t, err, tt)
			assert.Equal(t, mode.LetterPlan{}, plan)
			assert.Nil(t, plan.Degrees)

			m, err := mode.FromNumbersFromWithRule(tt.sys, tt.adjacent, tt.start, tt.rule)
			checkRuleFailure(t, err, tt)
			assert.Zero(t, m)
			assert.False(t, m.IsValid())
		})
	}
}

// classRecorder is a custom spelling rule that records, in call order, the
// class numbers passed to its function.
type classRecorder struct {
	mu      sync.Mutex
	classes []int
}

func (r *classRecorder) rule(names map[int]string, fallback system.SpellingRule) system.SpellingRule {
	inner := mappedRule(names, fallback)
	return system.MustCustomSpelling(func(s *system.System, c pitchclass.Class) spelling.Spelling {
		r.mu.Lock()
		r.classes = append(r.classes, c.Number())
		r.mu.Unlock()
		chosen, err := s.Spell(c, inner)
		if err != nil {
			return spelling.Spelling{}
		}
		return chosen
	})
}

// sequence returns the recorded classes with consecutive repetitions
// collapsed: the number of calls per class is not part of the contract.
func (r *classRecorder) sequence() []int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Compact(slices.Clone(r.classes))
}

func TestCustomRuleReceivesOnlyDegreeClassesInOrder(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	sharps := system.Sharps()
	for _, tt := range []struct {
		name     string
		sys      *system.System
		start    string
		adjacent []system.Step
		names    map[int]string
		classes  []int
		plan     []int
		err      error
	}{
		{name: "all degrees in order, start excluded", sys: traditional, start: "C", adjacent: stepList(3, 2, 2, 3, 2),
			classes: []int{3, 5, 7, 10}, plan: []int{1, 2, 4, 5, 6}},
		{name: "classes are counted from the class of the start", sys: traditional, start: "D-1", adjacent: stepList(2, 2, 3, 2, 3),
			classes: []int{3, 5, 8, 10}, plan: []int{1, 1, 3, 4, 5}},
		{name: "classes wrap around the octave", sys: traditional, start: "F+1", adjacent: stepList(2, 2, 3, 2, 3),
			classes: []int{8, 10, 1, 3}, plan: []int{1, 2, 3, 5, 6}},
		{name: "irregular system", sys: irregularSystem(), start: "D", adjacent: stepList(2, 2, 3, 2, 3),
			classes: []int{5, 7, 10, 0}, plan: []int{1, 3, 4, 5, 7}},
		{name: "one degree calls nothing", sys: traditional, start: "C", adjacent: stepList(12), classes: []int{}, plan: []int{1}},
		{name: "all names are chosen before an incorrect record is reported", sys: traditional, start: "G-2", adjacent: stepList(1, 1, 10),
			classes: []int{6, 7}, err: mode.ErrInvalidStructure},
		{name: "decreasing custom names", sys: traditional, start: "C", adjacent: stepList(4, 1, 7),
			names: map[int]string{4: "F-1", 5: "E+1"}, classes: []int{4, 5}, err: mode.ErrInvalidStructure},
		{name: "bad name of degree 3", sys: traditional, start: "G-2", adjacent: stepList(1, 1, 10),
			names: map[int]string{6: "F+1", 7: "A"}, classes: []int{6, 7}, err: system.ErrBadSpelling},
		{name: "a failure at degree 2 stops the walk", sys: traditional, start: "C", adjacent: stepList(3, 2, 2, 3, 2),
			names: map[int]string{3: "D"}, classes: []int{3}, err: system.ErrBadSpelling},
		{name: "unset start", sys: traditional, start: "unset", adjacent: stepList(3, 2, 2, 3, 2), classes: []int{}, err: mode.ErrUnset},
		{name: "nil system", sys: nil, start: "C", adjacent: stepList(3, 2, 2, 3, 2), classes: []int{}, err: mode.ErrUnset},
		{name: "missing numbers", sys: traditional, start: "C", adjacent: nil, classes: []int{}, err: mode.ErrEmptyInput},
		{name: "too many numbers", sys: traditional, start: "C", adjacent: ones(mode.MaxDegrees + 1), classes: []int{},
			err: mode.ErrInvalidArgument},
		{name: "number below 1", sys: traditional, start: "C", adjacent: stepList(3, 0, 4, 3, 2), classes: []int{},
			err: mode.ErrInvalidStructure},
		{name: "wrong sum", sys: traditional, start: "C", adjacent: stepList(3, 2, 2, 3), classes: []int{},
			err: mode.ErrSystemOctaveRequired},
		{name: "inadmissible start", sys: singleAccidentalSystem(), start: "C+2", adjacent: stepList(2, 2, 3, 2, 3), classes: []int{},
			err: system.ErrInadmissible},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start := spelling.MustParse(tt.start)

			suggested := new(classRecorder)
			plan, err := mode.SuggestLetterPlan(tt.sys, tt.adjacent, start, suggested.rule(tt.names, sharps))
			built := new(classRecorder)
			m, buildErr := mode.FromNumbersFromWithRule(tt.sys, tt.adjacent, start, built.rule(tt.names, sharps))

			assert.Equal(t, tt.classes, suggested.sequence())
			assert.Equal(t, tt.classes, built.sequence(), "building by the rule repeats the same classes")
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.ErrorIs(t, buildErr, tt.err)
				assert.Nil(t, plan.Degrees)
				assert.False(t, m.IsValid())
				return
			}
			require.NoError(t, err)
			require.NoError(t, buildErr)
			assert.Equal(t, mode.LetterPlan{Degrees: tt.plan, Cycle: 8}, plan)
			read, _ := m.LetterPlan()
			assert.Equal(t, plan, read)
		})
	}
}

type customPanic struct{ reason string }

func TestCustomRulePanicIsNotRecovered(t *testing.T) {
	t.Parallel()
	thrown := customPanic{"custom rule failed"}
	panicking := system.MustCustomSpelling(func(*system.System, pitchclass.Class) spelling.Spelling {
		panic(thrown)
	})
	traditional := system.Traditional()
	c := spelling.MustParse("C")
	steps := stepList(3, 2, 2, 3, 2)

	recovered := capturePanic(func() { _, _ = mode.SuggestLetterPlan(traditional, steps, c, panicking) })
	assert.Equal(t, thrown, recovered)
	recovered = capturePanic(func() { _, _ = mode.FromNumbersFromWithRule(traditional, steps, c, panicking) })
	assert.Equal(t, thrown, recovered)

	// The function is not called when no degree after the first exists or
	// when an earlier check fails, so nothing panics there.
	var plan mode.LetterPlan
	var err error
	assert.NotPanics(t, func() { plan, err = mode.SuggestLetterPlan(traditional, stepList(12), c, panicking) })
	require.NoError(t, err)
	assert.Equal(t, mode.LetterPlan{Degrees: []int{1}, Cycle: 8}, plan)
	assert.NotPanics(t, func() { _, err = mode.FromNumbersFromWithRule(traditional, stepList(3, 2, 2, 3), c, panicking) })
	assert.ErrorIs(t, err, mode.ErrSystemOctaveRequired)
}

func TestSuggestedPlanIsOwnedByTheCaller(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	c := spelling.MustParse("C")
	sharps := system.Sharps()
	input := stepList(3, 2, 2, 3, 2)
	suggested := mode.LetterPlan{Degrees: []int{1, 2, 4, 5, 6}, Cycle: 8}

	plan := mode.MustSuggestLetterPlan(traditional, input, c, sharps)
	require.Equal(t, suggested, plan)
	built := mode.MustFromNumbersFromWithRule(traditional, input, c, sharps)

	// KP-15: the caller edits the second size to 3 and builds the edited plan.
	plan.Degrees[1] = 3
	edited := mode.MustFromNumbersFromWithPlan(traditional, input, c, plan)
	checkMode(t, edited, "m3 M2 M2 A2 d3")
	assert.Equal(t, "C E-1 F G A+1", degreeNames(t, edited, traditional, "C"))

	// The edit changes neither a new suggestion nor the mode built earlier.
	assert.Equal(t, suggested, mode.MustSuggestLetterPlan(traditional, input, c, sharps))
	checkMode(t, built, "A2 d3 M2 A2 d3")
	read, _ := built.LetterPlan()
	assert.Equal(t, suggested, read)

	// Two suggestions are independent, also after an append.
	first := mode.MustSuggestLetterPlan(traditional, input, c, sharps)
	second := mode.MustSuggestLetterPlan(traditional, input, c, sharps)
	first.Degrees[0] = 99
	first.Degrees = append(first.Degrees, 9)
	assert.Equal(t, suggested, second)

	// Changing the passed numbers after the call does not change the mode.
	input[0], input[1] = 2, 3
	checkMode(t, built, "A2 d3 M2 A2 d3")
	assert.Equal(t, stepList(2, 3, 2, 3, 2), input)
}

func TestFromNumbersWithoutPlanIsUnchanged(t *testing.T) {
	t.Parallel()
	// F-17: five numbers still need an explicit plan or a spelling rule.
	m, err := mode.FromNumbers(system.Traditional(), stepList(3, 2, 2, 3, 2))
	checkCategory(t, err, mode.ErrWrongDegreeCount)
	assert.Zero(t, m)
	m, err = mode.FromNumbersFrom(system.Traditional(), stepList(3, 2, 2, 3, 2), spelling.MustParse("C"))
	checkCategory(t, err, mode.ErrWrongDegreeCount)
	assert.Zero(t, m)
}
