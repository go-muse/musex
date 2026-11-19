package modal_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/interval/naming"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/template"
)

func ExampleRelative() {
	m := mode.MustFromTemplate(template.Dorian())
	p, err := modal.Relative(m, 6)
	if err != nil {
		panic(err)
	}
	name, ok, err := naming.English().Specific(p.Interval).NameFull()
	if err != nil {
		panic(err)
	}
	fmt.Println(p.Number, p.Category, p.Interval.Generic().Size(), name, ok)
	fmt.Println(p.Interval.Quality(), p.Interval.Deviation(), p.Interval.Multiplicity())
	// Output:
	// 6 high 6 major sixth true
	// major 0 0
}

func ExampleGroupRelative() {
	m := mode.MustFromTemplate(template.Dorian())
	groups := modal.MustGroupRelative(m)
	fmt.Println(groups.High, groups.Neutral, groups.Low)
	groups.High[0] = 99
	fmt.Println(modal.MustGroupRelative(m).High)
	// Output:
	// [2 6] [1 4 5] [3 7]
	// [2 6]
}

func ExampleAnalyzeAbsolute() {
	m := mode.MustFromTemplate(template.Dorian())
	all := modal.MustAnalyzeAbsolute(m)
	fourth := all.Positions[3]
	fmt.Println(fourth.Number, fourth.Counts, fourth.Score, fourth.Rank)
	for _, group := range all.Order {
		fmt.Println(group.Rank, group.Score, group.Numbers)
	}
	fmt.Println(modal.MustCompareAbsolute(m, 4, 5))
	// Output:
	// 4 {0 1 3 3 0} -2 5
	// 1 5 [6]
	// 2 4 [2]
	// 3 2 [5]
	// 4 0 [1]
	// 5 -2 [4]
	// 6 -4 [7]
	// 7 -5 [3]
	// 1
}

func ExampleRelativePositions() {
	m := mode.MustOpenFromDegreeIntervals([]interval.Interval{interval.PerfectUnison(), interval.MajorSecond(), interval.MustNew(interval.Neutral, 3)})
	p, err := modal.Relative(m, 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(p.Category)
	all, err := modal.RelativePositions(m)
	fmt.Println(all == nil, errors.Is(err, modal.ErrUnsupportedQuality))
	_, err = modal.Absolute(m, 1)
	fmt.Println(errors.Is(err, modal.ErrUnsupportedStructure))
	_, err = modal.Absolute(m, 0)
	fmt.Println(errors.Is(err, modal.ErrInvalidDegree))
	// Output:
	// high
	// true true
	// true
	// true
}
