package rounding_test

import (
	"cmp"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/go-muse/musex/common/rounding"
)

func Example() {
	for _, mode := range rounding.Modes() {
		fmt.Printf("%s: %.0f %.0f\n", mode, mode.MustFloat(2.5), mode.MustFloat(-2.5))
	}
	// Output:
	// down: 2 -3
	// up: 3 -2
	// toward-zero: 2 -2
	// away-from-zero: 3 -3
	// half-up: 3 -2
	// half-down: 2 -3
	// half-away-from-zero: 3 -3
	// half-toward-zero: 2 -2
	// half-even: 2 -2
}

func ExampleMode_Choose() {
	// The caller compares exact distances on its own non-uniform grid.
	comparison := cmp.Compare(3, 3)
	neighbor := rounding.HalfAwayFromZero.MustChoose(comparison, rounding.Negative, false)
	fmt.Println(neighbor)
	// Output: lower
}

func ExampleMode_ChooseIndex() {
	const nearestMode = rounding.HalfDown
	fmt.Println(nearestMode.MustChooseIndex(0, 0))
	fmt.Println(rounding.Up.MustChooseIndex(0, -1))
	fmt.Println(rounding.HalfEven.MustChooseIndex(-3, 0))
	// Output:
	// 0
	// 1
	// -2
}

func ExampleMode_QuoRem() {
	q, r := rounding.HalfUp.MustQuoRem(1234, 480)
	fmt.Println(q, r)
	q, r = rounding.HalfDown.MustQuoRem(1234, 480)
	fmt.Println(q, r)
	// Output:
	// 3 -206
	// 2 274
}

func ExampleMode_IntMultiple() {
	fmt.Println(rounding.HalfUp.MustIntMultiple(1234, 480))
	fmt.Println(rounding.HalfDown.MustIntMultiple(1200, 480))
	// Output:
	// 1440
	// 960
}

func ExampleMode_RatDecimal() {
	x := big.NewRat(1, 8)
	fmt.Println(rounding.HalfEven.MustRatDecimal(x, 2))
	fmt.Println(rounding.HalfAwayFromZero.MustRatDecimal(x, 2))
	// Output:
	// 0.12
	// 0.13
}

func ExampleMode_FloatDecimal() {
	fmt.Println(rounding.HalfUp.MustFloatDecimal(0.145, 2))
	fmt.Println(strconv.FormatFloat(0.145, 'f', 2, 64))
	// Output:
	// 0.15
	// 0.14
}

func ExampleMode_FloatMultiple() {
	fmt.Println(rounding.HalfUp.MustFloatMultiple(0.15, 0.1))
	fmt.Println(rounding.HalfDown.MustFloatMultiple(0.15, 0.1))
	// Output:
	// 0.2
	// 0.1
}

func ExampleMode_FloatFraction() {
	fmt.Println(rounding.HalfDown.MustFloatFraction(0.5, 3))
	fmt.Println(rounding.HalfEven.MustFloatFraction(0.5, 3))
	// Output:
	// 0.3333333333333333
	// 0.6666666666666666
}

func ExampleMode_Mirror() {
	fmt.Println(rounding.HalfUp.Mirror())
	fmt.Println(rounding.Down.Mirror())
	// Output:
	// half-down
	// up
}

func ExampleModes() {
	for _, mode := range rounding.Modes() {
		fmt.Println(mode)
	}
	// Output:
	// down
	// up
	// toward-zero
	// away-from-zero
	// half-up
	// half-down
	// half-away-from-zero
	// half-toward-zero
	// half-even
}

func ExampleParse() {
	mode, err := rounding.Parse("half-even")
	fmt.Println(mode, err)
	_, err = rounding.Parse("HALF-EVEN")
	fmt.Println(errors.Is(err, rounding.ErrInvalidArgument))
	// Output:
	// half-even <nil>
	// true
}

func ExampleMode_MarshalText() {
	type config struct {
		Mode rounding.Mode `json:"mode"`
	}
	encoded, err := json.Marshal(config{Mode: rounding.HalfEven})
	if err != nil {
		panic(err)
	}
	var decoded config
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))
	fmt.Println(decoded.Mode)
	// Output:
	// {"mode":"half-even"}
	// half-even
}

func ExampleMode_Float_unset() {
	var mode rounding.Mode
	_, err := mode.Float(1.5)
	fmt.Println(errors.Is(err, rounding.ErrUnset))
	// Output: true
}
