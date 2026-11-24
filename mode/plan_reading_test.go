package mode_test

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the letter plan of a mode, of rebuilding a mode from its numbers
// and plan, of the Must forms and of concurrent use (issue #146, item 4;
// business-spec, KP-18, F-15 and F-16).

func TestModeLetterPlan(t *testing.T) {
	t.Parallel()
	shortened := mode.MustFromCode("ionian")
	shortened.MustRemoveDegree(7).MustRemoveDegree(4)
	for _, tt := range []struct {
		name    string
		m       mode.Mode
		degrees []int
		cycle   int
		closed  bool
	}{
		{"major pentatonic", mode.MustFromCode("pentatonic-major"), []int{1, 2, 3, 5, 6}, 8, true},
		{"minor pentatonic", mode.MustFromCode("pentatonic-minor"), []int{1, 3, 4, 5, 7}, 8, true},
		{"major", mode.MustFromCode("ionian"), []int{1, 2, 3, 4, 5, 6, 7}, 8, true},
		{"whole tone", mode.MustFromCode("whole-tone"), []int{1, 2, 3, 4, 5, 6}, 8, true},
		{"blues repeats a letter", mode.MustFromCode("blues"), []int{1, 3, 4, 4, 5, 7}, 8, true},
		{"blues b5 repeats a letter", mode.MustFromCode("blues-b5"), []int{1, 3, 4, 5, 5, 7}, 8, true},
		{"octatonic", mode.MustFromCode("octatonic-hw"), []int{1, 2, 3, 3, 4, 5, 6, 7}, 8, true},
		{"bebop major", mode.MustFromCode("bebop-major"), []int{1, 2, 3, 4, 5, 5, 6, 7}, 8, true},
		{"ten degrees", mode.MustFromCode("messiaen-7-degraaf"), []int{1, 1, 2, 2, 4, 4, 5, 5, 6, 7}, 8, true},
		{"cycle of a fourth", mode.MustParse("M2 M2 m2"), []int{1, 2, 3}, 4, true},
		{"cycle of two octaves", mode.MustParse("M9 m7"), []int{1, 9}, 15, true},
		{"cycle of a diminished fifth", mode.MustParse("M2 m2 M2 m2"), []int{1, 2, 3, 4}, 5, true},
		{"one degree", mode.MustParse("P8"), []int{1}, 8, true},
		{"tritones", mode.MustParse("A4 d5"), []int{1, 4}, 8, true},
		{"diminished second", mode.MustParse("d2 A7"), []int{1, 2}, 8, true},
		{"doubly diminished second", mode.MustParse("dd2 AA7"), []int{1, 2}, 8, true},
		{"unison degree", mode.MustParse("A1 dd2 A7"), []int{1, 1, 2}, 8, true},
		{"size 8 before the cycle", mode.MustParse("d8 A1"), []int{1, 8}, 8, true},
		{"open thirds", mode.MustParse("m3 M2 |"), []int{1, 3, 4}, 0, false},
		{"open seconds", mode.MustParse("M2 m2 |"), []int{1, 2, 3}, 0, false},
		{"open unison", mode.MustParse("A1 |"), []int{1, 1}, 0, false},
		{"origin does not matter", shortened, []int{1, 2, 3, 5, 6}, 8, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, closed := tt.m.LetterPlan()
			assert.Equal(t, mode.LetterPlan{Degrees: tt.degrees, Cycle: tt.cycle}, plan)
			assert.Equal(t, tt.closed, closed)
			_, hasCycle := tt.m.Cycle()
			assert.Equal(t, hasCycle, closed, "the second result agrees with Cycle")
			if tt.closed {
				// The plan depends only on the structure, not on the origin.
				unnamed := mode.MustClosedFromIntervals(tt.m.AdjacentIntervals())
				samePlan, _ := unnamed.LetterPlan()
				assert.Equal(t, plan, samePlan)
			}
		})
	}
	assert.True(t, shortened.Changed())
	checkMode(t, shortened, "M2 M2 m3 M2 m3")

	var unset mode.Mode
	plan, closed := unset.LetterPlan()
	assert.False(t, closed)
	assert.Equal(t, mode.LetterPlan{}, plan)
	assert.Nil(t, plan.Degrees)
	assert.Zero(t, plan.Cycle)
}

func TestModeLetterPlanIsOwnedByTheCaller(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("pentatonic-major")
	original := m
	plan, _ := m.LetterPlan()
	plan.Degrees[1] = 3
	plan.Degrees = append(plan.Degrees, 9)
	plan.Cycle = 15

	again, closed := m.LetterPlan()
	assert.True(t, closed)
	assert.Equal(t, mode.LetterPlan{Degrees: []int{1, 2, 3, 5, 6}, Cycle: 8}, again)
	assert.True(t, m.Equal(original))
	checkMode(t, m, "M2 M2 m3 M2 m3")
	assert.True(t, m.HasOrigin())
	assert.False(t, m.Changed())
}

func TestRebuildFromNumbersAndPlan(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	positions19 := system.Positions19()

	pentatonic := mode.MustFromCode("pentatonic-major")
	plan, _ := pentatonic.LetterPlan()
	numbers := pentatonic.MustNumbers(traditional).Adjacent
	assert.Equal(t, stepList(2, 2, 3, 2, 3), numbers)
	assert.True(t, pentatonic.Equal(mode.MustFromNumbersWithPlan(traditional, numbers, plan)))

	blues := mode.MustFromCode("blues")
	bluesPlan, _ := blues.LetterPlan()
	bluesNumbers := blues.MustNumbers(positions19).Adjacent
	assert.Equal(t, stepList(5, 3, 1, 2, 5, 3), bluesNumbers)
	assert.True(t, blues.Equal(mode.MustFromNumbersWithPlan(positions19, bluesNumbers, bluesPlan)))
	// The same numbers with the plan of blues-b5 give another mode.
	b5Plan, _ := mode.MustFromCode("blues-b5").LetterPlan()
	other := mode.MustFromNumbersWithPlan(positions19, bluesNumbers, b5Plan)
	checkMode(t, other, "m3 M2 d2 AA1 m3 M2")
	assert.False(t, other.Equal(blues))

	for _, tt := range []struct {
		structure string
		numbers   []system.Step
	}{
		{"M2 M2 m2", stepList(2, 2, 1)},
		{"M9 m7", stepList(14, 10)},
	} {
		m := mode.MustParse(tt.structure)
		p, _ := m.LetterPlan()
		assert.Equal(t, tt.numbers, m.MustNumbers(traditional).Adjacent)
		assert.True(t, m.Equal(mode.MustFromNumbersWithPlan(traditional, tt.numbers, p)), tt.structure)
	}

	ionianMode := mode.MustFromCode("ionian")
	ionianPlan, _ := ionianMode.LetterPlan()
	d := spelling.MustParse("D")
	irregularNumbers := ionianMode.MustNumbersFrom(irregularSystem(), d).Adjacent
	assert.Equal(t, stepList(1, 2, 1, 2, 2, 2, 2), irregularNumbers)
	assert.True(t, ionianMode.Equal(mode.MustFromNumbersFromWithPlan(irregularSystem(), irregularNumbers, d, ionianPlan)))

	open := mode.MustParse("m3 M2 |")
	openPlan, _ := open.LetterPlan()
	assert.True(t, open.Equal(mode.MustOpenFromNumbers(traditional, stepList(3, 2), openPlan.Degrees)))

	// A zero number does not rebuild: creation by numbers needs numbers >= 1.
	zero := mode.MustParse("d2 A7")
	zeroPlan, _ := zero.LetterPlan()
	zeroNumbers := zero.MustNumbers(traditional).Adjacent
	assert.Equal(t, stepList(0, 12), zeroNumbers)
	rebuilt, err := mode.FromNumbersWithPlan(traditional, zeroNumbers, zeroPlan)
	checkCategory(t, err, mode.ErrInvalidStructure)
	assert.Zero(t, rebuilt)
}

// rebuild creates a mode from numbers and a plan in the form selected by
// start ("" for the form without a name).
func rebuild(sys *system.System, start string, closed bool, numbers []system.Step, plan mode.LetterPlan) (mode.Mode, error) {
	switch {
	case closed && start == "":
		return mode.FromNumbersWithPlan(sys, numbers, plan)
	case closed:
		return mode.FromNumbersFromWithPlan(sys, numbers, spelling.MustParse(start), plan)
	case start == "":
		return mode.OpenFromNumbers(sys, numbers, plan.Degrees)
	default:
		return mode.OpenFromNumbersFrom(sys, numbers, spelling.MustParse(start), plan.Degrees)
	}
}

func TestRebuildEveryCatalogEntry(t *testing.T) {
	t.Parallel()
	builtIn, ok := mode.BuiltInRepository().Catalog()
	require.True(t, ok)
	custom, ok := customRepository().Catalog()
	require.True(t, ok)
	irregular := irregularSystem()
	for _, tt := range []struct {
		name  string
		sys   *system.System
		start string
	}{
		{"traditional", system.Traditional(), ""},
		{"19-position", system.Positions19(), ""},
		{"31 sharps and flats", system.SharpsAndFlats31(), ""},
		{"irregular from C", irregular, "C"},
		{"irregular from D", irregular, "D"},
		{"irregular from F sharp", irregular, "F+1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rebuilt := 0
			for _, e := range slices.Concat(builtIn.Entries(), custom.Entries()) {
				m, err := mode.FromEntry(e)
				if err != nil {
					continue
				}
				numbers, err := numbersOf(m, tt.sys, tt.start)
				if err != nil {
					continue
				}
				plan, closed := m.LetterPlan()
				got, err := rebuild(tt.sys, tt.start, closed, numbers.Adjacent, plan)
				if slices.Min(numbers.Adjacent) < 1 {
					assert.ErrorIs(t, err, mode.ErrInvalidStructure, e.Code())
					continue
				}
				require.NoError(t, err, e.Code())
				assert.True(t, m.Equal(got), "%s: got %s", e.Code(), got)
				rebuilt++
			}
			assert.GreaterOrEqual(t, rebuilt, 40)
		})
	}
}

func TestLetterPlanMustForms(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	c := spelling.MustParse("C")
	steps := stepList(3, 2, 2, 3, 2)

	plan, err := mode.SuggestLetterPlan(traditional, steps, c, system.Flats())
	require.NoError(t, err)
	assert.Equal(t, plan, mode.MustSuggestLetterPlan(traditional, steps, c, system.Flats()))
	byRule, err := mode.FromNumbersFromWithRule(traditional, steps, c, system.Flats())
	require.NoError(t, err)
	assert.Equal(t, observe(byRule), observe(mode.MustFromNumbersFromWithRule(traditional, steps, c, system.Flats())))
	open, err := mode.OpenFromNumbers(traditional, stepList(2, 1), []int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, observe(open), observe(mode.MustOpenFromNumbers(traditional, stepList(2, 1), []int{1, 2, 3})))
	openFrom, err := mode.OpenFromNumbersFrom(traditional, stepList(2, 1), c, []int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, observe(openFrom), observe(mode.MustOpenFromNumbersFrom(traditional, stepList(2, 1), c, []int{1, 2, 3})))
	repo := mode.BuiltInRepository()
	found, err := repo.IdentifyNumbers(mode.Closed, traditional, stepList(2, 2, 3, 2, 3))
	require.NoError(t, err)
	assert.Equal(t, codes(found), codes(repo.MustIdentifyNumbers(mode.Closed, traditional, stepList(2, 2, 3, 2, 3))))
	foundFrom, err := repo.IdentifyNumbersFrom(mode.Closed, traditional, stepList(2, 2, 3, 2, 3), c)
	require.NoError(t, err)
	assert.Equal(t, codes(foundFrom), codes(repo.MustIdentifyNumbersFrom(mode.Closed, traditional, stepList(2, 2, 3, 2, 3), c)))
	empty := repo.MustIdentifyNumbers(mode.Closed, traditional, stepList(2, 2, 3, 2))
	assert.NotNil(t, empty)
	assert.Empty(t, empty)

	t1 := singleAccidentalSystem()
	cDoubleSharp := spelling.MustParse("C+2")
	sum := stepList(3, 2, 2, 3)
	pentatonic := stepList(2, 2, 3, 2, 3)
	major := stepList(2, 2, 1, 2, 2, 2, 1)
	for _, tt := range []struct {
		name   string
		normal func() error
		must   func()
		causes []error
	}{
		{"MustSuggestLetterPlan, incorrect record",
			func() error {
				_, err := mode.SuggestLetterPlan(system.SharpsAndFlats31(), stepList(1, 30), c, system.Sharps())
				return err
			},
			func() { mode.MustSuggestLetterPlan(system.SharpsAndFlats31(), stepList(1, 30), c, system.Sharps()) },
			[]error{mode.ErrInvalidStructure}},
		{"MustSuggestLetterPlan, system cause",
			func() error {
				_, err := mode.SuggestLetterPlan(naturalsOnlySystem(), steps, c, system.Sharps())
				return err
			},
			func() { mode.MustSuggestLetterPlan(naturalsOnlySystem(), steps, c, system.Sharps()) },
			[]error{system.ErrNoSpelling}},
		{"MustFromNumbersFromWithRule, wrong sum",
			func() error {
				_, err := mode.FromNumbersFromWithRule(traditional, sum, c, system.Sharps())
				return err
			},
			func() { mode.MustFromNumbersFromWithRule(traditional, sum, c, system.Sharps()) },
			[]error{mode.ErrSystemOctaveRequired}},
		{"MustFromNumbersFromWithRule, inadmissible start",
			func() error {
				_, err := mode.FromNumbersFromWithRule(t1, pentatonic, cDoubleSharp, system.Sharps())
				return err
			},
			func() { mode.MustFromNumbersFromWithRule(t1, pentatonic, cDoubleSharp, system.Sharps()) },
			[]error{system.ErrInadmissible}},
		{"MustOpenFromNumbers, size count",
			func() error {
				_, err := mode.OpenFromNumbers(traditional, stepList(2, 1), []int{1, 2})
				return err
			},
			func() { mode.MustOpenFromNumbers(traditional, stepList(2, 1), []int{1, 2}) },
			[]error{mode.ErrWrongDegreeCount}},
		{"MustOpenFromNumbers, out of range",
			func() error {
				_, err := mode.OpenFromNumbers(traditional, stepList(1<<40), []int{1, 2})
				return err
			},
			func() { mode.MustOpenFromNumbers(traditional, stepList(1<<40), []int{1, 2}) },
			[]error{mode.ErrOutOfRange, interval.ErrOutOfRange}},
		{"MustOpenFromNumbersFrom, inadmissible name",
			func() error {
				_, err := mode.OpenFromNumbersFrom(naturalsOnlySystem(), stepList(2, 1), c, []int{1, 2, 3})
				return err
			},
			func() { mode.MustOpenFromNumbersFrom(naturalsOnlySystem(), stepList(2, 1), c, []int{1, 2, 3}) },
			[]error{system.ErrInadmissible}},
		{"MustIdentifyNumbers, start required",
			func() error {
				_, err := repo.IdentifyNumbers(mode.Closed, irregularSystem(), pentatonic)
				return err
			},
			func() { repo.MustIdentifyNumbers(mode.Closed, irregularSystem(), pentatonic) },
			[]error{interval.ErrStartRequired}},
		{"MustIdentifyNumbers, unset kind",
			func() error {
				_, err := repo.IdentifyNumbers(mode.KindUnset, traditional, pentatonic)
				return err
			},
			func() { repo.MustIdentifyNumbers(mode.KindUnset, traditional, pentatonic) },
			[]error{mode.ErrInvalidArgument}},
		{"MustIdentifyNumbersFrom, inadmissible start",
			func() error {
				_, err := repo.IdentifyNumbersFrom(mode.Closed, t1, major, cDoubleSharp)
				return err
			},
			func() { repo.MustIdentifyNumbersFrom(mode.Closed, t1, major, cDoubleSharp) },
			[]error{system.ErrInadmissible}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := tt.normal()
			require.Error(t, base)
			recovered := capturePanic(tt.must)
			require.NotNil(t, recovered, "the Must form did not panic")
			failure, ok := recovered.(error)
			require.True(t, ok, "panicked with %T: %v", recovered, recovered)
			assert.Equal(t, base.Error(), failure.Error(), "the Must form panics with the error of the normal form")
			for _, cause := range tt.causes {
				assert.True(t, errors.Is(base, cause), "missing %v in %v", cause, base)
				assert.True(t, errors.Is(failure, cause), "missing %v in %v", cause, failure)
			}
		})
	}
}

func TestLetterPlanConcurrentCalls(t *testing.T) {
	t.Parallel()
	const workers = 8
	traditional := system.Traditional()
	c := spelling.MustParse("C")
	steps := stepList(3, 2, 1, 1, 3, 2)
	// The custom function has no mutable state, so concurrent use is safe.
	rule := system.MustCustomSpelling(func(s *system.System, class pitchclass.Class) spelling.Spelling {
		return s.MustSpell(class, system.Flats())
	})
	repo := mode.BuiltInRepository()
	shared := mode.MustFromCode("blues")

	plans := make([]mode.LetterPlan, workers)
	planErrors := make([]error, workers)
	found := make([][]string, workers)
	foundErrors := make([]error, workers)
	read := make([]mode.LetterPlan, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			plans[i], planErrors[i] = mode.SuggestLetterPlan(traditional, steps, c, rule)
			var entries []template.Entry
			entries, foundErrors[i] = repo.IdentifyNumbers(mode.Closed, traditional, steps)
			found[i] = codes(entries)
			read[i], _ = shared.LetterPlan()
		})
	}
	wg.Wait()
	for i := range workers {
		require.NoError(t, planErrors[i])
		require.NoError(t, foundErrors[i])
		assert.Equal(t, mode.LetterPlan{Degrees: []int{1, 3, 4, 5, 5, 7}, Cycle: 8}, plans[i])
		assert.Equal(t, []string{"blues", "blues-b5"}, found[i])
		assert.Equal(t, mode.LetterPlan{Degrees: []int{1, 3, 4, 4, 5, 7}, Cycle: 8}, read[i])
	}
}
