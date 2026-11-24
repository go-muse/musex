package mode_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// The suggested plan is ordinary data: the caller can edit it and build the
// edited plan with FromNumbersFromWithPlan.
func ExampleSuggestLetterPlan() {
	traditional := system.Traditional()
	c := spelling.MustParse("C")
	steps := []system.Step{3, 2, 2, 3, 2}

	plan, err := mode.SuggestLetterPlan(traditional, steps, c, system.Flats())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(plan.Degrees, plan.Cycle) // C E flat F G B flat

	plan.Degrees[1] = 2 // write the second degree as D sharp instead
	edited := mode.MustFromNumbersFromWithPlan(traditional, steps, c, plan)
	fmt.Println(edited.MustIntervalsString())
	// Output:
	// [1 3 4 5 7] 8
	// A2 d3 M2 m3 M2
}

// The rule chooses the names of the degrees after the first one; when the
// names it chooses cannot form a mode, pass an explicit plan instead.
func ExampleFromNumbersFromWithRule() {
	traditional := system.Traditional()
	c := spelling.MustParse("C")
	steps := []system.Step{3, 2, 2, 3, 2}

	for _, rule := range []system.SpellingRule{system.Flats(), system.Sharps()} {
		m, err := mode.FromNumbersFromWithRule(traditional, steps, c, rule)
		if err != nil {
			fmt.Println(err)
			return
		}
		notes := m.MustNotesFromIn(traditional, note.MustFromSpelling(c, cents.Cents{}, nil))
		names := make([]string, len(notes))
		for i, n := range notes {
			names[i] = n.Spelling().String()
		}
		fmt.Printf("%s: %s\n", m.MustIntervalsString(), strings.Join(names, " "))
	}

	// In 31-sharps-and-flats the sharps rule writes B double sharp, a name
	// below C, for the second degree.
	_, err := mode.FromNumbersFromWithRule(system.SharpsAndFlats31(), []system.Step{1, 30}, c, system.Sharps())
	fmt.Println(errors.Is(err, mode.ErrInvalidStructure))
	// Output:
	// m3 M2 M2 m3 M2: C E-1 F G B-1
	// A2 d3 M2 A2 d3: C D+1 F G A+1
	// true
}

func ExampleRepository_IdentifyNumbers() {
	repo := mode.BuiltInRepository()
	traditional := system.Traditional()

	found, err := repo.IdentifyNumbers(mode.Closed, traditional, []system.Step{2, 2, 3, 2, 3})
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, e := range found {
		fmt.Println(e.Code())
	}
	m := mode.MustFromEntry(found[0])
	fmt.Println(m.MustIntervalsString(), m.HasOrigin())

	none := repo.MustIdentifyNumbers(mode.Closed, traditional, []system.Step{2, 2, 3, 2})
	fmt.Println(len(none))
	// Output:
	// pentatonic-major
	// M2 M2 m3 M2 m3 true
	// 0
}

// In an irregular system the numbers of an entry depend on the name of the
// first degree, so the search needs that name.
func ExampleRepository_IdentifyNumbersFrom() {
	irregular := system.MustFromSteps(letter.C, [letter.Count]int{3, 1, 1, 2, 2, 2, 1}, 1, system.Integers(), "Irregular")
	repo := mode.BuiltInRepository()
	steps := []system.Step{3, 2, 2, 3, 2}

	found, err := repo.IdentifyNumbersFrom(mode.Closed, irregular, steps, spelling.MustParse("C"))
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, e := range found {
		fmt.Println(e.Code())
	}

	_, err = repo.IdentifyNumbers(mode.Closed, irregular, steps)
	fmt.Println(errors.Is(err, interval.ErrStartRequired))
	// Output:
	// pentatonic-minor
	// pentatonic-suspended
	// true
}

func ExampleOpenFromNumbers() {
	traditional := system.Traditional()

	m, err := mode.OpenFromNumbers(traditional, []system.Step{2, 1}, []int{1, 2, 3})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(m.MustIntervalsString(), m.IsOpen())

	// Three degrees need three sizes.
	_, err = mode.OpenFromNumbers(traditional, []system.Step{2, 1}, []int{1, 2})
	fmt.Println(errors.Is(err, mode.ErrWrongDegreeCount))
	// Output:
	// M2 m2 | true
	// true
}

func ExampleOpenFromNumbersFrom() {
	traditional := system.Traditional()
	c := spelling.MustParse("C")

	m, err := mode.OpenFromNumbersFrom(traditional, []system.Step{2, 1}, c, []int{1, 2, 3})
	if err != nil {
		fmt.Println(err)
		return
	}
	notes := m.MustNotesFromIn(traditional, note.MustFromSpelling(c, cents.Cents{}, nil))
	names := make([]string, len(notes))
	for i, n := range notes {
		names[i] = n.Spelling().String()
	}
	_, hasCycle := m.Cycle()
	fmt.Println(strings.Join(names, " "))
	fmt.Println(m.MustIntervalsString(), hasCycle)
	// Output:
	// C D E-1
	// M2 m2 | false
}

// The plan and the numbers of a mode rebuild it.
func ExampleMode_LetterPlan() {
	traditional := system.Traditional()
	m := mode.MustFromCode("pentatonic-major")

	plan, closed := m.LetterPlan()
	fmt.Println(plan.Degrees, plan.Cycle, closed)
	numbers := m.MustNumbers(traditional).Adjacent
	fmt.Println(numbers)
	fmt.Println(mode.MustFromNumbersWithPlan(traditional, numbers, plan).Equal(m))

	open := mode.MustParse("m3 M2 |")
	openPlan, closed := open.LetterPlan()
	fmt.Println(openPlan.Degrees, openPlan.Cycle, closed)
	// Output:
	// [1 2 3 5 6] 8 true
	// [2 2 3 2 3]
	// true
	// [1 3 4] 0 false
}
