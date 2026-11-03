package pitchclass_test

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/pitchclass"
)

// The examples print only contractual values: String, numbers, booleans and
// Notation. Error texts and panic values are not printed.

func Example() {
	c := pitchclass.MustNew(11, 12).Transpose(2)
	fmt.Println(c)

	ic := pitchclass.MustNew(10, 12).MustIntervalClass(pitchclass.MustNew(2, 12))
	fmt.Println(ic)
	// Output:
	// 1 mod 12
	// 4
}

func ExampleNew() {
	c, err := pitchclass.New(1, 12)
	fmt.Println(c, err)

	_, err = pitchclass.New(12, 12)
	fmt.Println(errors.Is(err, pitchclass.ErrInvalidClass))
	// Output:
	// 1 mod 12 <nil>
	// true
}

func ExampleReduce() {
	c, _ := pitchclass.Reduce(13, 12)
	fmt.Println(c)

	d, _ := pitchclass.Reduce(-1, 12)
	fmt.Println(d)
	// Output:
	// 1 mod 12
	// 11 mod 12
}

func ExampleAll() {
	all, err := pitchclass.All(12)
	if err != nil {
		fmt.Println(err)

		return
	}

	var numbers []int
	for c := range all {
		numbers = append(numbers, c.Number())
	}
	fmt.Println(numbers)
	// Output:
	// [0 1 2 3 4 5 6 7 8 9 10 11]
}

func ExampleClass_Transpose() {
	fmt.Println(pitchclass.MustNew(11, 12).Transpose(2))
	fmt.Println(pitchclass.MustNew(0, 12).Transpose(-1))
	// Output:
	// 1 mod 12
	// 11 mod 12
}

func ExampleClass_IntervalTo() {
	a, b := pitchclass.MustNew(10, 12), pitchclass.MustNew(2, 12)

	up, _ := a.IntervalTo(b)
	fmt.Println(up)

	back, _ := b.IntervalTo(a)
	fmt.Println(back)
	// Output:
	// 4
	// 8
}

func ExampleClass_IntervalTo_modulusMismatch() {
	_, err := pitchclass.MustNew(1, 12).IntervalTo(pitchclass.MustNew(1, 19))
	fmt.Println(errors.Is(err, pitchclass.ErrModulusMismatch))
	// Output:
	// true
}

func ExampleClass_IntervalClass() {
	ic, _ := pitchclass.MustNew(10, 12).IntervalClass(pitchclass.MustNew(2, 12))
	fmt.Println(ic)

	ic, _ = pitchclass.MustNew(0, 12).IntervalClass(pitchclass.MustNew(7, 12))
	fmt.Println(ic)
	// Output:
	// 4
	// 5
}

func ExampleClass_Invert() {
	for _, k := range []int{2, 4, 7} {
		fmt.Println(pitchclass.MustNew(k, 12).Invert(8).Number())
	}
	// Output:
	// 6
	// 4
	// 1
}

func ExampleInversionFixedClasses() {
	fixed, _ := pitchclass.InversionFixedClasses(4, 12)
	fmt.Println(fixed)

	none, _ := pitchclass.InversionFixedClasses(1, 12)
	fmt.Println(none)
	// Output:
	// [2 mod 12 8 mod 12]
	// []
}

func ExampleClass_Multiply() {
	for i, c := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(pitchclass.MustNew(c, 12).Multiply(7).Number())
	}
	fmt.Println()
	// Output:
	// 0 7 2 9 4 11 6 1 8 3 10 5
}

func ExampleNewOperator() {
	f, _ := pitchclass.NewOperator(7, 5, 12)
	fmt.Println(f)
	fmt.Println(f.IsTransposition())
	// Output:
	// T5M7 mod 12
	// false
}

func ExampleOperator_Apply() {
	f := pitchclass.MustNewOperator(7, 5, 12)

	c, _ := f.Apply(pitchclass.MustNew(1, 12))
	fmt.Println(c)
	// Output:
	// 0 mod 12
}

func ExampleOperator_Compose() {
	t3 := pitchclass.MustTransposition(3, 12)
	i8 := pitchclass.MustInversion(8, 12)

	f, _ := t3.Compose(i8) // I8 first, then T3
	fmt.Println(f)

	g, _ := i8.Compose(t3) // T3 first, then I8
	fmt.Println(g)
	// Output:
	// T11M11 mod 12
	// T5M11 mod 12
}

func ExampleClass_String() {
	c := pitchclass.MustNew(1, 12)
	fmt.Println(c.String())

	p, _ := pitchclass.Parse("1 mod 12")
	fmt.Println(p.Equal(c))
	// Output:
	// 1 mod 12
	// true
}

func ExampleClass_Notation() {
	ten, eleven := pitchclass.MustNew(10, 12), pitchclass.MustNew(11, 12)
	for _, st := range []pitchclass.Style{pitchclass.Decimal, pitchclass.TE, pitchclass.LowerTE, pitchclass.AB} {
		fmt.Println(ten.MustNotation(st), eleven.MustNotation(st))
	}
	// Output:
	// 10 11
	// T E
	// t e
	// A B
}

func ExampleOperators() {
	ops, _ := pitchclass.Operators(12)

	count := 0
	for f := range ops {
		if count < 3 {
			fmt.Println(f)
		}
		count++
	}
	fmt.Println(count)
	// Output:
	// T0M1 mod 12
	// T1M1 mod 12
	// T2M1 mod 12
	// 48
}

func ExampleCountOperators() {
	n, _ := pitchclass.CountOperators(12)
	fmt.Println(n)

	_, err := pitchclass.CountOperators(math.MaxInt)
	fmt.Println(errors.Is(err, pitchclass.ErrOverflow))

	exact, _ := pitchclass.CountOperatorsBig(math.MaxInt)
	fmt.Println(exact)
	// Output:
	// 48
	// true
	// 71139883463926830590344547103944736768
}

func ExampleTotient() {
	phi, _ := pitchclass.Totient(12)
	fmt.Println(phi)

	phi, _ = pitchclass.Totient(math.MaxInt)
	fmt.Println(phi)
	// Output:
	// 4
	// 7713001620195508224
}

func ExampleParseNotation() {
	c, _ := pitchclass.ParseNotation("E", pitchclass.TE, 12)
	fmt.Println(c)

	_, err := pitchclass.ParseNotation("E", pitchclass.TE, 11)
	fmt.Println(errors.Is(err, pitchclass.ErrNotation))
	// Output:
	// 11 mod 12
	// true
}

func Example_zeroValue() {
	var zero pitchclass.Class
	fmt.Println(zero.IsValid(), zero.Number(), zero.String())
	fmt.Println(zero.Equal(zero))

	_, err := zero.IntervalTo(pitchclass.MustNew(0, 12))
	fmt.Println(errors.Is(err, pitchclass.ErrInvalidValue))
	// Output:
	// false -1 Class(invalid)
	// false
	// true
}
