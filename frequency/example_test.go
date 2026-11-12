package frequency_test

import (
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
)

func Example() {
	tuner := frequency.MustWithinCents(cents.MustFromInt(5))
	for _, f := range []float64{435, 441, 445} {
		fmt.Println(f, tuner.MustCompare(f, 440))
	}
	// Output:
	// 435 -1
	// 441 0
	// 445 1
}

func ExampleEqual() {
	fmt.Println(frequency.MustEqual(440, 440))
	a, b := 0.1, 0.2
	fmt.Println(frequency.MustEqual(a+b, 0.3))
	_, err := frequency.Equal(440, 0)
	fmt.Println(errors.Is(err, frequency.ErrInvalidFrequency), errors.Is(err, cents.ErrInvalidArgument))
	// Output:
	// true
	// false
	// true true
}

func ExampleCompare() {
	fmt.Println(frequency.MustCompare(440, 441))
	fmt.Println(frequency.MustCompare(441, 440))
	fmt.Println(frequency.MustCompare(440, 440))
	// Output:
	// -1
	// 1
	// 0
}

func ExampleCriterion_Near() {
	soundsLike := func(c frequency.Criterion, f1, f2 float64) (bool, error) {
		return c.Near(f1, f2)
	}
	criteria := []frequency.Criterion{
		frequency.Strict(),
		frequency.MustWithinCents(cents.MustFromInt(5)),
		frequency.MustWithinHertz(0.5),
	}
	for _, criterion := range criteria {
		near, err := soundsLike(criterion, 440, 441)
		if err != nil {
			panic(err)
		}
		fmt.Println(criterion, near)
	}
	// Output:
	// strict false
	// cents(5) true
	// hertz(0.5) false
}

func ExampleAnyOf() {
	either := frequency.MustAnyOf(
		frequency.MustWithinCents(cents.Cent()),
		frequency.MustWithinHertz(0.5),
	)
	fmt.Println(either, either.MustNear(440, 441), either.MustCompare(440, 441))
	both := frequency.MustAllOf(
		frequency.MustWithinCents(cents.MustFromInt(5)),
		frequency.MustWithinHertz(2),
	)
	fmt.Println(both, both.MustNear(440, 441))
	// Output:
	// any(cents(1), hertz(0.5)) false -1
	// all(cents(5), hertz(2)) true
}

func ExampleWithinCents_zero() {
	adjacent := math.Nextafter(440, math.Inf(1))
	fmt.Println(frequency.Strict().MustNear(440, adjacent))
	fmt.Println(frequency.MustWithinCents(cents.Zero()).MustNear(440, adjacent))
	fmt.Println(frequency.MustWithinHertz(0).MustNear(440, adjacent))
	// Output:
	// false
	// true
	// false
}

func ExampleParse() {
	criterion := frequency.MustParse("any(cents(1200/19 cents), hertz(1e-1))")
	fmt.Println(criterion)
	_, err := frequency.Parse("any(cents(1), hertz(-1))")
	fmt.Println(errors.Is(err, frequency.ErrInvalidTolerance))
	_, err = frequency.Parse("cents(3/2)")
	fmt.Println(errors.Is(err, frequency.ErrInvalidText), errors.Is(err, cents.ErrInvalidText))
	// Output:
	// any(cents(1200/19 cents), hertz(0.1))
	// true
	// true true
}

func ExampleCriterion_Components() {
	criterion := frequency.MustParse("any(cents(1), all(hertz(1)))")
	components := criterion.Components()
	fmt.Println(criterion.Kind(), len(components))
	for _, component := range components {
		fmt.Println(component.Kind(), component)
	}
	tolerance, ok := components[0].CentsTolerance()
	fmt.Println(tolerance, ok)
	// Output:
	// any 2
	// cents cents(1)
	// all all(hertz(1))
	// 1 true
}

func ExampleCriterion_MarshalText() {
	type settings struct {
		Match frequency.Criterion `json:"match"`
	}
	data, err := jsonv2.Marshal(settings{Match: frequency.MustWithinCents(cents.MustFromInt(5))})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	var decoded settings
	if err := jsonv2.Unmarshal([]byte(`{"match":"strict"}`), &decoded); err != nil {
		panic(err)
	}
	fmt.Println(decoded.Match.Kind())
	// Output:
	// {"match":"cents(5)"}
	// strict
}

func ExampleCriterion_Compare_unset() {
	var criterion frequency.Criterion
	_, err := criterion.Compare(440, 440)
	fmt.Println(errors.Is(err, frequency.ErrUnset))
	// Output:
	// true
}

func ExampleTranspose() {
	fmt.Printf("%.6f\n", frequency.MustTranspose(440, 7, 12))
	// Output:
	// 659.255114
}

func ExampleRatio() {
	ratio := frequency.MustRatio(440, 660)
	fmt.Println(ratio.RatString())
	distance := frequency.MustCentsBetween(440, 660)
	fmt.Println(distance)
	// Output:
	// 3/2
	// 701.955001
}

func ExamplePartials() {
	partials := frequency.MustPartials(110, []int{2, 3, 4})
	fmt.Println(partials)
	subharmonics := frequency.MustSubharmonics(440, []int{1, 2, 3, 4})
	fmt.Printf("%.6f\n", subharmonics)
	// Output:
	// [220 330 440]
	// [440.000000 220.000000 146.666667 110.000000]
}

func ExampleBeatFrequency() {
	fmt.Println(frequency.MustBeatFrequency(440, 442))
	fmt.Println(frequency.MustBeatFrequency(440, 440))
	// Output:
	// 2
	// 0
}

func ExampleShiftOctaves() {
	fmt.Println(frequency.MustShiftOctaves(440, 1))
	fmt.Println(frequency.MustShiftOctaves(440, -1))
	// Output:
	// 880
	// 220
}

func ExampleNormalizeOctave() {
	for _, input := range []float64{880, 440, 220} {
		representative, k := frequency.MustNormalizeOctave(input, 110)
		fmt.Println(representative, k)
	}
	// Output:
	// 110 -3
	// 110 -2
	// 110 -1
}

func ExamplePeriod() {
	fmt.Printf("%.6f\n", 1000*frequency.MustPeriod(440))
	// Output:
	// 2.272727
}
