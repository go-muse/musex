package meter_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// Example shows the units and beats of 6/8, its kind in both traditions,
// and the division of its beat.
func Example() {
	m := meter.SixEight()
	fmt.Println(m, m.MustDuration(), m.MustUnitCount(), m.MustBeatCount())
	fmt.Println(m.MustBeats(), m.MustBeatsWritten())

	name, ok := m.MustKindName()
	fmt.Println(m.MustKind(), name, ok, m.MustRussianKind())
	fmt.Println(m.MustBeatDivision())

	// Output:
	// 6/8 3/4 6 2
	// [3/8@0/1 3/8@3/8] [4. 4.]
	// compound compound duple true compound-uniform
	// 3 6
}

// ExampleNew shows a simple signature and the reading of its record.
func ExampleNew() {
	m, err := meter.New(3, 4)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(m, m.MustNumerator(), m.MustDenominator(), m.MustUnit(), m.MustUnitWritten())
	fmt.Println(m.MustUnits())

	// Output:
	// 3/4 3 4 1/4 4
	// [1/4@0/1 1/4@1/4 1/4@1/2]
}

// ExampleNewSum shows a sum of signatures with different denominators.
func ExampleNewSum() {
	sum := meter.MustNewSum(meter.MustNew(2, 4), meter.MustNew(3, 8))
	fmt.Println(sum, sum.MustDuration(), sum.MustGroups())

	_, err := sum.Denominator()
	fmt.Println(errors.Is(err, meter.ErrNoCommonDenominator))

	// Output:
	// 2/4+3/8 7/8 [1/2@0/1 3/8@1/2]
	// true
}

// ExampleMeter_StrongBeats shows the strong beats of 7/8 grouped as 3+2+2.
func ExampleMeter_StrongBeats() {
	s7 := meter.MustNew(7, 8).MustWithGrouping(3, 2, 2)
	strong, rel := s7.MustStrongBeats()
	fmt.Println(s7, strong, rel)

	var weights []string
	for _, w := range s7.MustUnitWeights() {
		weights = append(weights, w.String())
	}

	fmt.Println(weights)

	// Output:
	// 7/8[3+2+2] 0/1 [3/8 5/8]
	// [strong weak weak relatively-strong weak relatively-strong weak]
}

// ExampleMeter_UnitAt shows the place of a position in the measure.
func ExampleMeter_UnitAt() {
	m := meter.SixEight()
	t := duration.MustNewPosition(7, 16)

	n, off := m.MustUnitAt(t)
	beat, beatOff := m.MustBeatAt(t)
	fmt.Println(n, off, beat, beatOff, m.MustWeightAt(t))
	fmt.Println(m.MustBeatStart(2))

	// Output:
	// 4 1/16 2 1/16 off-unit
	// 3/8
}

// ExampleMeter_Kind shows the kinds of meters in both traditions.
func ExampleMeter_Kind() {
	for _, m := range []meter.Meter{
		meter.ThreeFour(),
		meter.MustNew(5, 8),
		meter.MustNew(5, 8).MustWithGrouping(2, 3),
		meter.MustNew(15, 8),
	} {
		name, ok := m.MustKindName()
		ru, _ := m.MustKindRussianName()
		fmt.Println(m, m.MustKind(), m.MustRussianKind(), name, ok, ru)
	}

	// Output:
	// 3/4 simple simple simple triple true простой трёхдольный
	// 5/8 asymmetrical compound-mixed asymmetrical true асимметричный
	// 5/8[2+3] asymmetrical compound-mixed asymmetrical true асимметричный
	// 15/8 compound compound-uniform compound true сложный
}

// ExampleMeter_EqualStructure shows the three equalities of meters.
func ExampleMeter_EqualStructure() {
	fmt.Println(meter.ThreeFour().Equal(meter.SixEight()), meter.ThreeFour().MustEqualDuration(meter.SixEight()))
	fmt.Println(meter.CommonTime().Equal(meter.FourFour()), meter.CommonTime().MustEqualStructure(meter.FourFour()))

	// Output:
	// false true
	// false true
}

// ExampleMeter_Compare shows sorting by size and then by text.
func ExampleMeter_Compare() {
	ms := []meter.Meter{meter.SixEight(), meter.CommonTime(), meter.ThreeFour(), meter.TwoFour(), meter.FourFour()}
	slices.SortFunc(ms, meter.Meter.Compare)
	fmt.Println(ms)

	// Output:
	// [2/4 3/4 6/8 4/4 C]
}

// ExampleMeter_ToSum shows changes of the record that keep the structure.
func ExampleMeter_ToSum() {
	fmt.Println(meter.MustNewAdditive([]int{3, 2, 3}, 8).MustToSum())
	fmt.Println(meter.MustNewAdditive([]int{2, 3}, 8).MustToSimple())
	fmt.Println(meter.FourFour().MustToSymbol(), meter.CutTime().MustToNumeric())

	// Output:
	// 3/8+2/8+3/8
	// 5/8[2+3]
	// C 2/2
}

// ExampleMeter_Shorter shows other units and the signature by size.
func ExampleMeter_Shorter() {
	fmt.Println(meter.ThreeFour().MustShorter(), meter.ThreeFour().MustLonger())
	fmt.Println(meter.SixEight().MustWithDenominator(4))
	fmt.Println(meter.MustFromDuration(duration.MustNew(3, 4), 8))

	// Output:
	// 3/8 3/2
	// 6/4
	// 6/8
}

// ExampleGroupingVariants shows the groupings of seven units into groups
// of two and three.
func ExampleGroupingVariants() {
	fmt.Println(meter.MustGroupingVariants(7))

	// Output:
	// [[2 2 3] [2 3 2] [3 2 2]]
}

// ExampleSymbol_Name shows the names of the cut time symbol and its text.
func ExampleSymbol_Name() {
	cut, _ := meter.SymbolCut.MustName()
	ru, _ := meter.SymbolCut.MustRussianName()
	fmt.Println(cut, "/", ru)
	fmt.Println(meter.MustParse("¢"), meter.MustParse("¢").Equal(meter.CutTime()))

	// Output:
	// cut time / alla breve
	// C| true
}

// Example_json shows a meter in JSON; omitzero omits an unset field.
func Example_json() {
	type bar struct {
		Meter    meter.Meter `json:"meter"`
		Previous meter.Meter `json:"previous,omitzero"`
	}

	data, err := json.Marshal(bar{Meter: meter.MustParse("5/8[2+3]")})
	fmt.Println(string(data), err)

	var back bar

	err = json.Unmarshal([]byte(`{"meter":"¢","previous":"(3+2+3)/8"}`), &back)
	fmt.Println(back.Meter, back.Previous, err)

	// Output:
	// {"meter":"5/8[2+3]"} <nil>
	// C| (3+2+3)/8 <nil>
}

// Example_errors shows the handling of errors and limits.
func Example_errors() {
	_, err := meter.MustNew(5, 8).Beats()
	fmt.Println(errors.Is(err, meter.ErrGroupingRequired))
	fmt.Println(meter.MustNew(5, 8).MustKind())

	_, err = meter.MustNew(4, 3).UnitWritten()
	fmt.Println(errors.Is(err, meter.ErrNotNotatable), errors.Is(err, duration.ErrNotNotatable))

	_, err = meter.New(1025, 8)
	fmt.Println(errors.Is(err, meter.ErrTooMany))

	_, err = meter.ThreeFour().BeatStart(4)
	fmt.Println(errors.Is(err, meter.ErrOutsideMeter))

	var u meter.Meter

	_, err = u.Duration()
	fmt.Println(u, errors.Is(err, meter.ErrUnset))

	// Output:
	// true
	// asymmetrical
	// true true
	// true
	// true
	// Meter(unset) true
}
