package measure_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
)

// The examples follow architecture 4.17 and section 8.

// Example shows a grid with a pickup: bar starts, the address of a
// position, the place in the meter, and the unit count of the pickup.
func Example() {
	g := measure.MustNew(meter.ThreeFour()).MustWithPickup(duration.MustNew(1, 4))
	fmt.Println(g)
	fmt.Println(g.MustBarStart(0), g.MustBarStart(1), g.MustBarStart(2))

	a := g.MustLocate(duration.MustNewPosition(9, 8))
	fmt.Println(a, g.MustPositionOf(a))

	m, u := g.MustMetricPlace(measure.MustNewAddress(0, duration.Origin()))
	fmt.Println(m, u, m.MustWeightAt(u))
	fmt.Println(g.MustMetricCountAt(0, duration.Origin()), g.MustCountAt(0, duration.Origin()))

	// Output:
	// 3/4;pickup:1/4
	// 0/1 1/4 1/1
	// 2:1/8 9/8
	// 3/4 1/2 weak
	// 3/1 1/1
}

// Example_errors shows errors checked with errors.Is.
func Example_errors() {
	g2 := measure.MustNew(meter.FourFour()).
		MustWithMark(3, meter.ThreeFour()).
		MustWithLength(4, duration.MustNew(1, 2), measure.AlignStart).
		MustWithBarCount(8)
	four := measure.MustNew(meter.FourFour())

	_, err := g2.TstampAt(4, duration.Origin())
	fmt.Println(errors.Is(err, measure.ErrUndefined))

	_, err = measure.MustNew(meter.MustNew(5, 8)).SplitBy(duration.Origin(), duration.MustNew(5, 8), measure.LevelGroups)
	fmt.Println(errors.Is(err, measure.ErrGroupingRequired), errors.Is(err, meter.ErrGroupingRequired))

	_, err = g2.Locate(duration.MustNewPosition(25, 4))
	fmt.Println(errors.Is(err, measure.ErrOutsideGrid))

	_, err = four.WithPickup(duration.MustNew(5, 4))
	fmt.Println(errors.Is(err, measure.ErrLongerThanMeter))

	_, err = four.PositionOf(measure.MustNewAddress(1, duration.MustNewPosition(1, 1)))
	fmt.Println(errors.Is(err, measure.ErrOutsideBar))

	var unset measure.Grid
	_, err = unset.BarStart(1)
	fmt.Println(unset, errors.Is(err, measure.ErrUnset))

	// Output:
	// true
	// true true
	// true
	// true
	// true
	// Grid(unset) true
}

// ExampleGrid_Bars shows the information of bars in order.
func ExampleGrid_Bars() {
	g := measure.MustNew(meter.ThreeFour()).MustWithPickup(duration.MustNew(1, 4))

	for b := range g.MustBars(0, 2) {
		fmt.Println(b.Number, b.Length, b.Start, b.Kind, b.Shift)
	}

	// Output:
	// 0 1/4 0/1 incomplete 1/2
	// 1 3/4 1/4 full 0/1
	// 2 3/4 1/1 full 0/1
}

// ExampleGrid_Split shows a finite grid with a change of the meter and an
// explicit length, its runs, and a split at bar lines and at groups.
func ExampleGrid_Split() {
	g2 := measure.MustNew(meter.FourFour()).
		MustWithMark(3, meter.ThreeFour()).
		MustWithLength(4, duration.MustNew(1, 2), measure.AlignStart).
		MustWithBarCount(8)
	fmt.Println(g2)

	end, finite := g2.MustEnd()
	fmt.Println(end, finite)
	fmt.Println(g2.MustChanges())
	fmt.Println(g2.MustMeterRuns(1, 5))
	fmt.Println(g2.MustSplit(duration.MustNewPosition(1, 1), duration.MustNew(5, 2)))

	four := measure.MustNew(meter.FourFour())
	fmt.Println(four.MustSplitBy(duration.MustNewPosition(3, 8), duration.MustNew(3, 8), measure.LevelGroups))
	fmt.Println(four.MustBarStart(1 << 62))

	// Output:
	// 4/4;3:3/4;4:length:1/2;bars:8
	// 25/4 true
	// [3:3/4]
	// [1-2:4/4 3-5:3/4]
	// [1/1@2:0/1 3/4@3:0/1 1/2@4:0/1 1/4@5:0/1]
	// [1/8@1:3/8 1/4@1:1/2]
	// 4611686018427387903/1
}

// ExampleMeterCountAt shows the unit count and MEI tstamp of one meter and
// the closing length of a pickup.
func ExampleMeterCountAt() {
	sum := meter.MustNewSum(meter.MustNew(2, 4), meter.MustNew(3, 8))
	fmt.Println(measure.MustMeterCountAt(sum, duration.MustNewPosition(9, 16)))
	fmt.Println(measure.MustMeterTstampAt(meter.SixEight(), duration.MustNewPosition(3, 8)))

	closing, ok := measure.MustComplement(meter.FourFour(), duration.MustNew(1, 4))
	fmt.Println(closing, ok)

	// Output:
	// 7/2
	// 4/1
	// 3/4 true
}

// ExampleGrid_TstampAt shows MEI tstamp in a full bar and the two unit
// counts of a pickup.
func ExampleGrid_TstampAt() {
	g := measure.MustNew(meter.FourFour()).MustWithPickup(duration.MustNew(1, 4))
	fmt.Println(g.MustTstampAt(1, duration.MustNewPosition(1, 8)))

	_, err := g.TstampAt(0, duration.Origin())
	fmt.Println(errors.Is(err, measure.ErrUndefined))
	fmt.Println(g.MustCountAt(0, duration.MustNewPosition(1, 8)), g.MustMetricCountAt(0, duration.MustNewPosition(1, 8)))

	// Output:
	// 3/2
	// true
	// 3/2 9/2
}

// ExampleGrid_InsertBars shows inserting and deleting bars.
func ExampleGrid_InsertBars() {
	g := measure.MustNew(meter.FourFour()).MustWithMark(3, meter.ThreeFour())
	fmt.Println(g.MustInsertBars(3, 2))
	fmt.Println(g.MustDeleteBars(1, 2))

	// Output:
	// 4/4;5:3/4
	// 3/4
}

// ExampleGrid_Append shows joining grids and slicing the result.
func ExampleGrid_Append() {
	first := measure.MustNew(meter.FourFour()).MustWithBarCount(2)
	second := measure.MustNew(meter.ThreeFour()).MustWithPickup(duration.MustNew(1, 4)).MustWithBarCount(2)

	joined := first.MustAppend(second)
	fmt.Println(joined)
	fmt.Println(joined.MustEnd())
	fmt.Println(joined.MustSlice(3, 5))

	// Output:
	// 4/4;3:3/4;3:length:1/4:end;bars:5
	// 15/4 true
	// 3/4;1:length:1/4:end;bars:3
}

// ExampleGrid_EqualMarkup shows the three equalities and Simplify.
func ExampleGrid_EqualMarkup() {
	four := measure.MustNew(meter.FourFour())
	fmt.Println(four.MustWithMark(5, meter.FourFour()).MustEqualMarkup(four))
	fmt.Println(measure.MustParse("4/4;5:4/4").Equal(four))
	fmt.Println(measure.MustNew(meter.ThreeFour()).MustEqualBarLengths(measure.MustNew(meter.SixEight())))

	g := four.MustWithMark(5, meter.FourFour()).MustWithLength(3, duration.MustNew(1, 1), measure.AlignStart)
	s := g.MustSimplify()
	fmt.Println(s, s.Equal(g), s.MustEqualMarkup(g))

	// Output:
	// true
	// false
	// true
	// 4/4 false true
}

// ExampleParse shows parsing, the canonical text, and JSON.
func ExampleParse() {
	g := measure.MustParse("3/4;pickup:2/8;5:¢")
	fmt.Println(g)

	b, err := json.Marshal(struct {
		Grid measure.Grid    `json:"grid"`
		At   measure.Address `json:"at"`
	}{g, measure.MustParseAddress("2:2/16")})
	fmt.Println(string(b), err)

	// Output:
	// 3/4;pickup:1/4;5:C|
	// {"grid":"3/4;pickup:1/4;5:C|","at":"2:1/8"} <nil>
}

// ExampleAddress shows an address as a value.
func ExampleAddress() {
	a := measure.MustNewAddress(2, duration.MustNewPosition(1, 8))
	fmt.Println(a, a.MustBar(), a.MustOffset(), a.Compare(measure.MustParseAddress("3:0/1")))

	// Output:
	// 2:1/8 2 1/8 -1
}
