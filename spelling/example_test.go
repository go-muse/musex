package spelling_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"
)

func Example() {
	cs := spelling.MustNew(letter.C, alteration.Sharp())
	fmt.Println(cs)

	up := cs.MustRaise()
	fmt.Println(up)
	fmt.Println(up.Equal(spelling.CDoubleSharp()))

	// The same sound, but a different spelling.
	fmt.Println(spelling.CSharp().Equal(spelling.DFlat()))
	// Output:
	// C+1
	// C+2
	// true
	// false
}

func ExampleNew() {
	sharp, err := spelling.New(letter.C, alteration.Sharp())
	fmt.Println(sharp, err == nil)

	quarterUp, err := spelling.New(letter.C, alteration.MustFromFraction(1, 2))
	fmt.Println(quarterUp, err == nil)

	tripleSharp, err := spelling.New(letter.E, alteration.MustFromInt(3))
	fmt.Println(tripleSharp, err == nil)

	low, err := spelling.New(letter.B, alteration.MustFromFraction(-7, 2))
	fmt.Println(low, err == nil)
	// Output:
	// C+1 true
	// C+1/2 true
	// E+3 true
	// B-7/2 true
}

func ExampleNew_invalidLetter() {
	s, err := spelling.New(letter.Letter(0), alteration.Sharp())
	fmt.Println(errors.Is(err, letter.ErrInvalid))
	fmt.Println(s.IsValid())
	// Output:
	// true
	// false
}

func ExampleMustNew() {
	s := spelling.MustNew(letter.D, alteration.Flat())
	fmt.Println(s, s.Equal(spelling.DFlat()))
	// Output:
	// D-1 true
}

func ExampleSpelling_Parts() {
	l, a := spelling.FSharp().Parts()
	fmt.Println(l, a)
	fmt.Println(spelling.FSharp().Letter(), spelling.FSharp().Alteration())
	// Output:
	// F +1
	// F +1
}

func Example_readyNames() {
	fmt.Println(spelling.C(), spelling.CSharp(), spelling.DFlat(), spelling.BDoubleFlat())
	fmt.Println(spelling.CSharp().Equal(spelling.MustNew(letter.C, alteration.Sharp())))
	// Output:
	// C C+1 D-1 B-2
	// true
}

func ExamplePredefined() {
	names := spelling.Predefined()
	fmt.Println(len(names))
	fmt.Println(names[:5])
	// Output:
	// 35
	// [C-2 C-1 C C+1 C+2]
}

func ExampleSpelling_Equal() {
	fmt.Println(spelling.CSharp().Equal(spelling.DFlat()))
	fmt.Println(spelling.BSharp().Equal(spelling.C()))

	twoQuarters := spelling.MustNew(letter.C, alteration.MustFromFraction(2, 4))
	half := spelling.MustNew(letter.C, alteration.MustFromFraction(1, 2))
	fmt.Println(twoQuarters.Equal(half))
	// Output:
	// false
	// false
	// true
}

func ExampleSpelling_Raise() {
	s := spelling.C()
	for range 3 {
		s = s.MustRaise()
		fmt.Println(s)
	}

	fmt.Println(spelling.DFlat().MustRaise())
	fmt.Println(spelling.CFlat().MustLower())
	// Output:
	// C+1
	// C+2
	// C+3
	// D
	// C-2
}

func ExampleSpelling_Lower() {
	s, err := spelling.D().Lower()
	fmt.Println(s, err == nil)
	// Output:
	// D-1 true
}

func ExampleSpelling_RaiseBy() {
	fmt.Println(spelling.C().MustRaiseBy(3))
	fmt.Println(spelling.CSharp().MustRaiseBy(-2))
	fmt.Println(spelling.CDoubleSharp().MustLowerBy(4))
	// Output:
	// C+3
	// C-1
	// C-2
}

func ExampleSpelling_LowerBy() {
	s, err := spelling.ESharp().LowerBy(3)
	fmt.Println(s, err == nil)
	// Output:
	// E-2 true
}

func ExampleSpelling_Natural() {
	fmt.Println(spelling.CSharp().Natural())
	fmt.Println(spelling.MustNew(letter.A, alteration.MustFromFraction(1, 2)).Natural())
	// Output:
	// C
	// A
}

func ExampleSpelling_WithAlteration() {
	fmt.Println(spelling.CSharp().WithAlteration(alteration.Flat()))
	fmt.Println(spelling.C().WithAlteration(alteration.MustFromFraction(1, 2)))
	// Output:
	// C-1
	// C+1/2
}

func ExampleSpelling_RaiseByAlteration() {
	half := alteration.MustFromFraction(1, 2)

	fmt.Println(spelling.CSharp().MustRaiseByAlteration(half))
	fmt.Println(spelling.MustNew(letter.A, half).MustRaiseByAlteration(half))
	fmt.Println(spelling.C().MustLowerByAlteration(half))
	// Output:
	// C+3/2
	// A+1
	// C-1/2
}

func ExampleSpelling_Compare() {
	names := []spelling.Spelling{
		spelling.BSharp(),
		spelling.C(),
		spelling.CFlat(),
		spelling.B(),
		spelling.DFlat(),
		spelling.CDoubleSharp(),
	}
	slices.SortFunc(names, spelling.Spelling.Compare)
	fmt.Println(names)

	// The order is not the pitch order: B sharp comes after C.
	fmt.Println(spelling.BSharp().Compare(spelling.C()))
	// Output:
	// [C-1 C C+2 D-1 B B+1]
	// 1
}

func ExampleSpelling_LetterDistanceUp() {
	from, to := spelling.CSharp(), spelling.EFlat()
	fmt.Println(from.LetterDistanceUp(to), from.LetterDistanceDown(to))
	// Output:
	// 2 5
}

func ExampleSpelling_LetterDistanceDown() {
	fmt.Println(spelling.BSharp().LetterDistanceDown(spelling.CFlat()))
	// Output:
	// 6
}

func ExampleSpelling_AlterationDifference() {
	fmt.Println(spelling.CSharp().MustAlterationDifference(spelling.CFlat()))
	fmt.Println(spelling.DFlat().MustAlterationDifference(spelling.CSharp()))
	// Output:
	// +2
	// -2
}

func ExampleParse() {
	s, err := spelling.Parse("C+1/2")
	fmt.Println(s, err == nil)

	// The machine form is not a notation.
	_, err = spelling.Parse("C#")
	fmt.Println(errors.Is(err, spelling.ErrInvalidForm))
	// Output:
	// C+1/2 true
	// true
}

func ExampleMustParse() {
	fmt.Println(spelling.MustParse("B-7/2").Equal(spelling.MustNew(letter.B, alteration.MustFromFraction(-7, 2))))
	// Output:
	// true
}

func ExampleSpelling_String() {
	var unset spelling.Spelling

	fmt.Println(spelling.CSharp().String(), spelling.BFlat().String(), unset.String())
	// Output:
	// C+1 B-1 unset
}

func ExampleSpelling_MarshalText() {
	text, err := spelling.CSharp().MarshalText()
	fmt.Println(string(text), err == nil)

	var unset spelling.Spelling

	text, err = unset.MarshalText()
	fmt.Println(string(text), err == nil)
	// Output:
	// C+1 true
	// unset true
}

func Example_json() {
	type key struct {
		Tonic     spelling.Spelling            `json:"tonic"`
		Spellings []spelling.Spelling          `json:"spellings"`
		By        map[spelling.Spelling]string `json:"by"`
	}

	in := key{
		Tonic:     spelling.CSharp(),
		Spellings: []spelling.Spelling{spelling.DFlat(), {}},
		By:        map[spelling.Spelling]string{spelling.BFlat(): "B", spelling.C(): "C"},
	}

	data, err := json.Marshal(in)
	if err != nil {
		fmt.Println("marshal failed")

		return
	}

	fmt.Println(string(data))

	var out key
	if err := json.Unmarshal(data, &out); err != nil {
		fmt.Println("unmarshal failed")

		return
	}

	fmt.Println(out.Tonic.Equal(in.Tonic), out.Spellings, out.By[spelling.BFlat()])
	fmt.Println(out.Spellings[1].IsValid())
	// Output:
	// {"tonic":"C+1","spellings":["D-1","unset"],"by":{"B-1":"B","C":"C"}}
	// true [D-1 unset] B
	// false
}

func ExampleSpelling_IsNatural() {
	var unset spelling.Spelling

	fmt.Println(spelling.C().IsNatural())
	fmt.Println(spelling.CSharp().IsNatural())
	fmt.Println(unset.IsNatural())
	// Output:
	// true
	// false
	// false
}

func ExampleSpelling_SameLetter() {
	fmt.Println(spelling.CSharp().SameLetter(spelling.CFlat()))
	fmt.Println(spelling.CSharp().SameLetter(spelling.DFlat()))
	// Output:
	// true
	// false
}

func ExampleSpelling_IsInteger() {
	fmt.Println(spelling.CSharp().IsInteger())
	fmt.Println(spelling.MustNew(letter.A, alteration.MustFromFraction(1, 2)).IsInteger())
	// Output:
	// true
	// false
}

func ExampleSpelling_Direction() {
	var unset spelling.Spelling

	fmt.Println(spelling.CSharp().Direction(), spelling.DFlat().Direction(), spelling.C().Direction(), unset.Direction())
	// Output:
	// raised lowered natural unset
}

func ExampleDirection_String() {
	fmt.Println(spelling.DirectionRaised.String(), spelling.Direction(9).String())
	// Output:
	// raised Direction(9)
}

func ExampleSpelling_SameAlteration() {
	fmt.Println(spelling.CSharp().SameAlteration(spelling.DSharp()))
	fmt.Println(spelling.CSharp().SameAlteration(spelling.CFlat()))
	// Output:
	// true
	// false
}

func ExampleSpelling_IsValid() {
	var s spelling.Spelling

	fmt.Println(s.IsValid(), s.String())
	fmt.Println(spelling.C().IsValid())
	// Output:
	// false unset
	// true
}

func Example_outOfRange() {
	top := spelling.MustNew(letter.C, alteration.MustFromInt(alteration.MaxMagnitude))

	_, err := top.Raise()
	fmt.Println(errors.Is(err, alteration.ErrOutOfRange))
	fmt.Println(top)
	// Output:
	// true
	// C+16777216
}
