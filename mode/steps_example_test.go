package mode_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

func ExampleMode_StepsBetween() {
	m := mode.MustFromCode("ionian")
	sys := system.Traditional()
	up, err := m.StepsBetween(sys, mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 3})
	if err != nil {
		panic(err)
	}
	// The distance is signed: the seventh degree is above the first one.
	down := m.MustStepsBetween(sys, mode.DegreePlace{Number: 7}, mode.DegreePlace{Number: 1})
	// The first degree of the next cycle is one step above the seventh.
	next := m.MustStepsBetween(sys, mode.DegreePlace{Number: 7}, mode.DegreePlace{Number: 1, Cycle: 1})
	fmt.Println(up, down, next)

	// Without a name of the first degree, only regular systems are accepted.
	irregular := system.MustFromSteps(letter.C, [letter.Count]int{3, 1, 1, 2, 2, 2, 1}, 1, system.Integers(), "Irregular")
	_, err = m.StepsBetween(irregular, mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 3})
	fmt.Println(errors.Is(err, interval.ErrStartRequired))
	// Output:
	// 4 -11 1
	// true
}

func ExampleMode_StepsBetweenFrom() {
	// C 0, D 3, E 4, F 5, G 7, A 9, B 11: an irregular 12-position system.
	irregular := system.MustFromSteps(letter.C, [letter.Count]int{3, 1, 1, 2, 2, 2, 1}, 1, system.Integers(), "Irregular")
	m := mode.MustFromCode("ionian")
	first, second := mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 2}
	fromC := m.MustStepsBetweenFrom(irregular, spelling.MustParse("C"), first, second)
	fromD, err := m.StepsBetweenFrom(irregular, spelling.MustParse("D"), first, second)
	if err != nil {
		panic(err)
	}
	fmt.Println(fromC, fromD)

	// A mode repeating at a fourth: from D, its cycles take 4 and then 5 steps.
	fourth := mode.MustParse("M2 M2 m2")
	d := spelling.MustParse("D")
	one := fourth.MustStepsBetweenFrom(irregular, d, mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 1, Cycle: 1})
	two := fourth.MustStepsBetweenFrom(irregular, d, mode.DegreePlace{Number: 1, Cycle: 1}, mode.DegreePlace{Number: 1, Cycle: 2})
	fmt.Println(one, two)
	// Output:
	// 3 1
	// 4 5
}

func ExamplePositions_DegreesAt() {
	numbers := mode.MustFromCode("ionian").MustNumbers(system.Traditional())
	fmt.Println(numbers.DegreesAt(4), numbers.DegreesAt(1), numbers.DegreesAt(12))
	// Two degrees can share an offset.
	fmt.Println(mode.MustParse("d2 A7").MustNumbers(system.Traditional()).DegreesAt(0))
	// Output:
	// [3] [] []
	// [1 2]
}

func ExamplePositions_OccupiedSteps() {
	numbers := mode.MustParse("dd2 AA7").MustNumbers(system.Traditional())
	// Offsets ascend; the second degree is one step below the first.
	for offset, degrees := range numbers.OccupiedSteps() {
		fmt.Println(offset, degrees)
	}
	// Output:
	// -1 [2]
	// 0 [1]
}

func ExamplePositions_StepsIn() {
	numbers := mode.MustFromCode("ionian").MustNumbers(system.Traditional())
	for offset, degrees := range numbers.StepsIn(0, 4) {
		fmt.Println(offset, degrees)
	}
	// Output:
	// 0 [1]
	// 1 []
	// 2 [2]
	// 3 []
	// 4 [3]
}

func ExamplePositions_StepsInBackward() {
	numbers := mode.MustFromCode("ionian").MustNumbers(system.Traditional())
	// The cycle boundary is not crossed: offset 12 has no degree.
	for offset, degrees := range numbers.StepsInBackward(9, 12) {
		fmt.Println(offset, degrees)
	}
	for offset, degrees := range numbers.OccupiedStepsBackward() {
		fmt.Println("highest:", offset, degrees)
		break
	}
	// Output:
	// 12 []
	// 11 [7]
	// 10 []
	// 9 [6]
	// highest: 11 [7]
}
