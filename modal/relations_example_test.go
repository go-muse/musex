package modal_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/template"
)

func ExampleRelationsFrom() {
	ionian := mode.MustFromTemplate(template.Ionian())
	row, err := modal.RelationsFrom(ionian, 7)
	if err != nil {
		panic(err)
	}
	// Degrees below VII are taken in the next cycle.
	for _, r := range row {
		fmt.Println(r.From, r.To, r.Cycle, r.Interval, r.Category)
	}
	// Output:
	// 7 7 0 P1 neutral
	// 7 1 1 m2 low
	// 7 2 1 m3 low
	// 7 3 1 P4 neutral
	// 7 4 1 d5 low
	// 7 5 1 m6 low
	// 7 6 1 m7 low
}

func ExampleRelationTo() {
	ionian := mode.MustFromTemplate(template.Ionian())
	up := modal.MustRelationTo(ionian, 1, 7)
	fmt.Println(up.Interval, up.Category, up.Cycle)
	next, err := modal.RelationTo(ionian, 7, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(next.Interval, next.Category, next.Cycle)

	// An open mode has no next cycle.
	open := mode.MustOpenFromDegreeIntervals([]interval.Interval{
		interval.PerfectUnison(), interval.MinorThird(), interval.PerfectFourth(),
	})
	_, err = modal.RelationTo(open, 3, 1)
	fmt.Println(errors.Is(err, modal.ErrUnsupportedStructure))
	// Output:
	// M7 high 0
	// m2 low 1
	// true
}

func ExampleRelationTable() {
	// C D E G A: the major pentatonic scale.
	pentatonic := mode.MustParse("M2 M2 m3 M2 m3")
	table, err := modal.RelationTable(pentatonic)
	if err != nil {
		panic(err)
	}
	// Rows are computed one at a time; low minus high relations of a row
	// give the absolute score of its degree.
	for row := range table {
		balance := 0
		for _, r := range row {
			switch r.Category {
			case modal.Low:
				balance++
			case modal.High:
				balance--
			}
		}
		fmt.Println(row[0].From, len(row), balance)
	}
	// Output:
	// 1 5 -3
	// 2 5 0
	// 3 5 3
	// 4 5 -2
	// 5 5 2
}

func ExampleCompareRelativePositions() {
	open := mode.MustOpenFromDegreeIntervals([]interval.Interval{
		interval.PerfectUnison(), interval.MinorThird(), interval.PerfectFourth(),
	})
	// Degrees P1 M2 M3, repeating at a perfect fourth.
	fourth := mode.MustParse("M2 M2 m2")
	c, err := modal.CompareRelativePositions(open, fourth)
	if err != nil {
		panic(err)
	}
	for _, d := range c.Degrees {
		fmt.Println(d.Number, d.Before.Interval, d.After.Interval, d.Before.Category, d.After.Category)
	}
	fmt.Println(c.IntervalChanges(), c.CategoryChanges())

	_, err = modal.CompareRelativePositions(mode.MustFromTemplate(template.Ionian()), open)
	fmt.Println(errors.Is(err, modal.ErrIncomparable))
	// Output:
	// 1 P1 P1 neutral neutral
	// 2 m3 M2 low high
	// 3 P4 M3 neutral high
	// [2 3] [2 3]
	// true
}

func ExampleComparePositions() {
	natural := mode.MustFromTemplate(template.Aeolian())
	harmonic := mode.MustFromTemplate(template.HarmonicMinor())
	c, err := modal.ComparePositions(natural, harmonic)
	if err != nil {
		panic(err)
	}
	fmt.Println(c.IntervalChanges(), c.CategoryChanges())
	fmt.Println(c.ScoreChanges(), c.RankChanges())
	seventh := c.Degrees[6]
	fmt.Println(seventh.Before.Relative.Interval, seventh.After.Relative.Interval,
		seventh.Before.Absolute.Score, seventh.After.Absolute.Score)
	// The sixth degree keeps its score but shares its rank with III.
	sixth := c.Degrees[5]
	fmt.Println(sixth.Before.Absolute.Score, sixth.After.Absolute.Score,
		sixth.Before.Absolute.Rank, sixth.After.Absolute.Rank)
	// Output:
	// [7] [7]
	// [1 2 3 4 5 7] [1 2 4 5 6 7]
	// m7 M7 -2 6
	// -5 -5 7 6
}
