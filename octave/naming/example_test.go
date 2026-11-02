package naming_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/octave/naming"
)

// The name of an octave by its ASPN number, and the ASPN number by the name.
func Example() {
	ru := naming.Russian()

	fmt.Println(ru.Name(4))
	fmt.Println(ru.MustASPN("малая"))

	// Output:
	// первая
	// 3
}

// The Russian catalog names ASPN octaves -1 to 9; other octaves have the
// empty name.
func ExampleRussian() {
	ru := naming.Russian()

	for aspn := -1; aspn <= 10; aspn++ {
		fmt.Printf("%d %q\n", aspn, ru.Name(aspn))
	}

	// Output:
	// -1 "субсубконтроктава"
	// 0 "субконтроктава"
	// 1 "контроктава"
	// 2 "большая"
	// 3 "малая"
	// 4 "первая"
	// 5 "вторая"
	// 6 "третья"
	// 7 "четвёртая"
	// 8 "пятая"
	// 9 "шестая"
	// 10 ""
}

// The English catalog names ASPN octaves 0 to 7.
func ExampleEnglish() {
	en := naming.English()

	for aspn := 0; aspn <= 8; aspn++ {
		fmt.Printf("%d %q\n", aspn, en.Name(aspn))
	}

	// Output:
	// 0 "subcontra"
	// 1 "contra"
	// 2 "great"
	// 3 "small"
	// 4 "one-line"
	// 5 "two-line"
	// 6 "three-line"
	// 7 "four-line"
	// 8 ""
}

// The German catalog names ASPN octaves 0 to 7.
func ExampleGerman() {
	de := naming.German()

	for aspn := 0; aspn <= 8; aspn++ {
		fmt.Printf("%d %q\n", aspn, de.Name(aspn))
	}

	// Output:
	// 0 "Subkontraoktave"
	// 1 "Kontraoktave"
	// 2 "große"
	// 3 "kleine"
	// 4 "eingestrichene"
	// 5 "zweigestrichene"
	// 6 "dreigestrichene"
	// 7 "viergestrichene"
	// 8 ""
}

// An octave without a name in the catalog, including a negative one, has the
// empty name; this is not an error.
func ExampleCatalog_Name() {
	ru := naming.Russian()

	fmt.Println(ru.Name(2))
	fmt.Printf("%q\n", ru.Name(-5))
	fmt.Println(naming.English().Name(4))

	// Output:
	// большая
	// ""
	// one-line
}

// Lookup is an exact match; an unknown name is reported by ErrUnknownName.
func ExampleCatalog_ASPN() {
	ru := naming.Russian()

	aspn, err := ru.ASPN("первая")
	fmt.Println(aspn, err == nil)

	_, err = ru.ASPN("Первая")
	fmt.Println(errors.Is(err, naming.ErrUnknownName))

	// Output:
	// 4 true
	// true
}

// A custom catalog from "ASPN number -> name" pairs; a name repeated for two
// octaves is reported by ErrDuplicateName.
func ExampleNew() {
	mine, err := naming.New(map[int]string{4: "middle", -3: "deep", 12: "high"})
	if err != nil {
		fmt.Println("unexpected error")

		return
	}

	fmt.Println(mine.Name(4), mine.Name(-3), mine.Name(12))
	fmt.Printf("%q\n", mine.Name(0))

	aspn, err := mine.ASPN("deep")
	fmt.Println(aspn, err == nil)

	_, err = naming.New(map[int]string{1: "a", 2: "a"})
	fmt.Println(errors.Is(err, naming.ErrDuplicateName))

	// Output:
	// middle deep high
	// ""
	// -3 true
	// true
}

// Catalogs are equal when they name the same octaves with the same names;
// pairs with the empty name do not count.
func ExampleCatalog_Equal() {
	fmt.Println(naming.Russian().Equal(naming.Russian()))

	withEmpty := naming.MustNew(map[int]string{1: "a", 2: ""})
	fmt.Println(withEmpty.Equal(naming.MustNew(map[int]string{1: "a"})))

	fmt.Println(naming.Russian().Equal(naming.English()))

	// Output:
	// true
	// true
	// false
}

// MustNew is for literals and data known to be valid; it panics with the
// error of New.
func ExampleMustNew() {
	defer func() {
		err, _ := recover().(error)
		fmt.Println(errors.Is(err, naming.ErrDuplicateName))
	}()

	c := naming.MustNew(map[int]string{0: "low", 1: "high"})
	fmt.Println(c.Name(0), c.Name(1))

	naming.MustNew(map[int]string{1: "x", 2: "x"})

	// Output:
	// low high
	// true
}

// MustASPN gives the ASPN number of a name known to be in the catalog.
func ExampleCatalog_MustASPN() {
	fmt.Println(naming.English().MustASPN("one-line"))
	fmt.Println(naming.German().MustASPN("kleine"))

	// Output:
	// 4
	// 3
}
