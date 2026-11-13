package duration_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"
)

// Example shows ordinary use (architecture 4.13).
func Example() {
	dq := duration.MustNewWritten(duration.Quarter, 1)
	fmt.Println(dq, dq.MustDuration())

	t8 := duration.MustNewWritten(duration.Eighth, 0, tuplet.Triplet())
	fmt.Println(t8, t8.MustDuration())

	dh := duration.MustNewWritten(duration.Half, 1)
	tie := duration.MustNewTied(duration.Half.MustWritten(), duration.Quarter.MustWritten())
	fmt.Println(tie, dh.MustDuration().Equal(tie.MustDuration()))

	q1 := duration.Quarter.MustWritten().MustWithChain(tuplet.Triplet())
	q2 := duration.Quarter.MustWritten().MustWithChain(tuplet.Sextuplet())
	fmt.Println(q1.Equal(q2), q1.MustEqualDuration(q2))

	fmt.Println(duration.MustNew(5, 16).MustNotate())
	fmt.Println(duration.MustNew(11, 16).MustNotateWith(duration.Breve, 1))
	fmt.Println(duration.MustNew(1, 3).MustToWritten(tuplet.Triplet()))

	p := duration.Origin().MustAdd(duration.Quarter.MustDuration()).MustAdd(duration.MustNew(1, 8))
	fmt.Println(p)

	eighth := fraction.MustNew(1, 8)
	fmt.Println(duration.MustNewPosition(7, 24).MustQuantize(eighth, rounding.HalfEven))

	name, ok := duration.Quarter.MustName()
	fmt.Println(name, ok)

	// Output:
	// 4. 3/8
	// 8[3:2] 1/12
	// 2~4 true
	// false true
	// [4 16]
	// [2 8.]
	// 2[3:2]
	// 3/8
	// 1/4
	// quarter note true
}

// Example_errors shows errors and bounds (architecture 4.13).
func Example_errors() {
	_, err := duration.MustNew(5, 16).ToWritten()
	fmt.Println(errors.Is(err, duration.ErrNotNotatable))

	_, err = duration.New(0, 1)
	fmt.Println(errors.Is(err, duration.ErrNotPositive))

	_, err = duration.MustNew(1, 32).Quantize(fraction.MustNew(1, 8), rounding.HalfEven)
	fmt.Println(errors.Is(err, duration.ErrQuantizedToZero))

	_, err = duration.MustNew(1, 4).Quantize(fraction.MustNew(1, 8), rounding.Mode(0))
	fmt.Println(errors.Is(err, duration.ErrInvalidArgument), errors.Is(err, rounding.ErrUnset))

	var w duration.Written
	_, err = w.Duration()
	fmt.Println(w, errors.Is(err, duration.ErrUnset))

	_, err = duration.ThousandTwentyFourth.MustWritten().AddDot()
	fmt.Println(err)

	_, _, err = duration.ThousandTwentyFourth.MustWritten().Subdivide()
	fmt.Println(errors.Is(err, duration.ErrOutsideSeries))

	// Output:
	// true
	// true
	// true
	// true true
	// Written(unset) true
	// <nil>
	// true
}

// ExampleNew shows a value and its arithmetic.
func ExampleNew() {
	d := duration.MustNew(6, 16)
	fmt.Println(d)
	fmt.Println(d.MustAdd(duration.MustNew(1, 8)))
	fmt.Println(d.MustRatio(duration.MustNew(1, 8)))

	quo, rem := duration.MustNew(1, 1).MustQuoRem(d)
	fmt.Println(quo, rem)

	fmt.Println(duration.MustSum(duration.Quarter.MustDuration(), duration.MustNew(1, 8), duration.MustNew(1, 8), duration.MustNew(1, 2)))

	// Output:
	// 3/8
	// 1/2
	// 3/1
	// 2 1/4
	// 1/1
}

// ExampleNewWritten shows a written duration with dots and a tuplet.
func ExampleNewWritten() {
	w := duration.MustNewWritten(duration.Quarter, 1, tuplet.Triplet())
	fmt.Println(w, w.MustDuration(), w.MustCoefficient())
	fmt.Println(w.MustAddDot(), w.MustAddDot().MustDuration())

	// Output:
	// 4.[3:2] 1/4 2/3
	// 4..[3:2] 7/24
}

// ExampleWritten_Equal shows the two equalities: by record and by value.
func ExampleWritten_Equal() {
	a := duration.MustParseWritten("4[3:2]")
	b := duration.MustParseWritten("4[6:4]")
	fmt.Println(a.Equal(b), a == b, a.MustEqualDuration(b), a.Compare(b))

	// Output:
	// false false true -1
}

// ExampleNewTied shows a tied duration.
func ExampleNewTied() {
	t := duration.MustNewTied(duration.Half.MustWritten(), duration.Quarter.MustWritten())
	fmt.Println(t, t.MustLen(), t.MustDuration())

	t = t.MustAppend(duration.Eighth.MustWritten())
	fmt.Println(t, t.MustDuration())
	fmt.Println(t.MustToWritten())

	// Output:
	// 2~4 2 3/4
	// 2~4~8 7/8
	// 2..
}

// ExampleWritten_Subdivide shows the normal subdivision.
func ExampleWritten_Subdivide() {
	for _, text := range []string{"4", "4.", "8[3:2]"} {
		count, part := duration.MustParseWritten(text).MustSubdivide()
		fmt.Println(text, count, part)
	}

	_, _, err := duration.MustParseWritten("4..").Subdivide()
	fmt.Println(errors.Is(err, duration.ErrNoSubdivision))

	// Output:
	// 4 2 8
	// 4. 3 8
	// 8[3:2] 2 16[3:2]
	// true
}

// ExampleDuration_ToWritten shows writing a value as one sign.
func ExampleDuration_ToWritten() {
	fmt.Println(duration.MustNew(3, 8).MustToWritten())
	fmt.Println(duration.MustNew(1, 12).MustToWritten(tuplet.Triplet()))

	_, err := duration.MustNew(5, 16).ToWritten()
	fmt.Println(errors.Is(err, duration.ErrNotNotatable))

	// Output:
	// 4.
	// 8[3:2]
	// true
}

// ExampleDuration_Notate shows writing a value with ties by default.
func ExampleDuration_Notate() {
	fmt.Println(duration.MustNew(11, 16).MustNotate())
	fmt.Println(duration.MustNew(5, 1).MustNotate())
	fmt.Println(duration.MustNew(5, 12).MustNotate(tuplet.Triplet()))

	_, err := duration.MustNew(1, 3).Notate()
	fmt.Println(errors.Is(err, duration.ErrNotNotatable))

	// Output:
	// [2 8 16]
	// [breve breve 1]
	// [2[3:2] 8[3:2]]
	// true
}

// ExampleDuration_NotateWith shows ties with dots and another longest base.
func ExampleDuration_NotateWith() {
	fmt.Println(duration.MustNew(7, 16).MustNotateWith(duration.Breve, 2))
	fmt.Println(duration.MustNew(12, 1).MustNotateWith(duration.Maxima, 0))
	fmt.Println(duration.MustNew(3, 1).MustNotateWith(duration.Whole, 0))

	// Output:
	// [4..]
	// [maxima longa]
	// [1 1 1]
}

// ExampleTied_ToWritten shows reducing a tie to one sign.
func ExampleTied_ToWritten() {
	t := duration.MustParseTied("4[3:2]~8[3:2]")
	fmt.Println(t.MustDuration(), t.MustToWritten(), t.MustToWritten(tuplet.Triplet()))

	// Output:
	// 1/4 4 4.[3:2]
}

// ExamplePosition_Add shows positions of a voice from its start.
func ExamplePosition_Add() {
	p := duration.Origin()
	for _, text := range []string{"4", "8", "8", "2"} {
		p = p.MustAdd(duration.MustParseWritten(text).MustDuration())
		fmt.Println(p)
	}

	a, b := duration.MustNewPosition(1, 4), duration.MustNewPosition(1, 2)
	fmt.Println(a.MustOffsetTo(b), a.MustDurationTo(b), b.MustOffsetTo(b))

	// Output:
	// 1/4
	// 3/8
	// 1/2
	// 1/1
	// 1/4 1/4 0/1
}

// ExamplePosition_Quantize shows quantizing a position in several modes.
func ExamplePosition_Quantize() {
	eighth := fraction.MustNew(1, 8)
	p := duration.MustNewPosition(5, 16)

	for _, mode := range []rounding.Mode{rounding.HalfUp, rounding.HalfDown, rounding.HalfEven, rounding.Down, rounding.Up} {
		fmt.Println(p.MustQuantize(eighth, mode))
	}

	// Output:
	// 3/8
	// 1/4
	// 1/4
	// 1/4
	// 3/8
}

// ExampleDuration_Quantize shows quantizing a value.
func ExampleDuration_Quantize() {
	fmt.Println(duration.MustNew(1, 3).MustQuantize(fraction.MustNew(1, 16), rounding.HalfEven))

	_, err := duration.MustNew(1, 32).Quantize(fraction.MustNew(1, 8), rounding.HalfEven)
	fmt.Println(errors.Is(err, duration.ErrQuantizedToZero))

	// Output:
	// 5/16
	// true
}

// ExampleQuantizeSpan shows quantizing a span on one grid.
func ExampleQuantizeSpan() {
	start, length := duration.MustQuantizeSpan(duration.MustNewPosition(1, 10), duration.MustNew(1, 5),
		fraction.MustNew(1, 8), rounding.HalfEven)
	fmt.Println(start, length, start.MustAdd(length))

	// Output:
	// 1/8 1/8 1/4
}

// ExampleBase_Name shows the names of bases and "no name".
func ExampleBase_Name() {
	for _, b := range []duration.Base{duration.Quarter, duration.Breve, duration.HundredTwentyEighth, duration.TwoHundredFiftySixth} {
		us, okUS := b.MustName()
		uk, okUK := b.MustBritishName()
		ru, okRU := b.MustRussianName()
		fmt.Printf("%s: %q %t, %q %t, %q %t\n", b, us, okUS, uk, okUK, ru, okRU)
	}

	// Output:
	// 4: "quarter note" true, "crotchet" true, "четвертная" true
	// breve: "double whole note" true, "breve" true, "бревис" true
	// 128: "" false, "" false, "стодвадцатьвосьмая" true
	// 256: "" false, "" false, "" false
}

// ExampleParseBase shows parsing texts, LilyPond words, and names.
func ExampleParseBase() {
	for _, text := range []string{"crotchet", "Quarter Note", `\breve`, "ЧЕТВЕРТНАЯ", "16"} {
		fmt.Println(duration.MustParseBase(text))
	}

	_, err := duration.ParseBase("quarter")
	fmt.Println(errors.Is(err, duration.ErrInvalidText))

	// Output:
	// 4
	// 4
	// breve
	// 4
	// 16
	// true
}

// ExampleParseWritten shows the text of the types.
func ExampleParseWritten() {
	w := duration.MustParseWritten("4.[5:4,3:2]")
	fmt.Println(w.MustBase(), w.MustDots(), w.MustChain(), w.MustDuration())
	fmt.Println(duration.MustParse("6/16"), duration.MustParsePosition("0/7"), duration.MustParseTied("2~4").MustDuration())

	// Output:
	// 4 1 [5:4 3:2] 1/5
	// 3/8 0/1 3/4
}

// Example_json shows the values in JSON; omitzero omits an unset field.
func Example_json() {
	type note struct {
		Value duration.Written  `json:"value"`
		At    duration.Position `json:"at"`
		Tie   duration.Tied     `json:"tie,omitzero"`
	}

	data, err := json.Marshal(note{Value: duration.MustParseWritten("8[3:2]"), At: duration.Origin()})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(data))

	var back note
	if err := json.Unmarshal([]byte(`{"value":"4.","at":"3/8","tie":"2~4"}`), &back); err != nil {
		fmt.Println(err)
	}

	fmt.Println(back.Value, back.At, back.Tie)

	// Output:
	// {"value":"8[3:2]","at":"0/1"}
	// 4. 3/8 2~4
}
