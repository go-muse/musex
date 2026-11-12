package tuplet_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// Example_namedTuplets shows the nine named tuplets with their default pairs
// and names (NF-7: named tuplets).
func Example_namedTuplets() {
	for _, t := range []tuplet.Tuplet{
		tuplet.Duplet(), tuplet.Triplet(), tuplet.Quadruplet(), tuplet.Quintuplet(), tuplet.Sextuplet(),
		tuplet.Septuplet(), tuplet.Octuplet(), tuplet.Nonuplet(), tuplet.Decuplet(),
	} {
		name, _ := t.MustName()
		russian, _ := t.MustRussianName()
		fmt.Println(t, name, russian)
	}

	// Output:
	// 2:3 duplet дуоль
	// 3:2 triplet триоль
	// 4:3 quadruplet квартоль
	// 5:4 quintuplet квинтоль
	// 6:4 sextuplet секстоль
	// 7:4 septuplet септоль
	// 8:6 octuplet октоль
	// 9:8 nonuplet нонемоль
	// 10:8 decuplet децимоль
}

// ExampleTuplet_Apply shows the coefficient, applying it to a duration, and
// the total duration of a tuplet (NF-7: coefficient and application).
func ExampleTuplet_Apply() {
	t := tuplet.Triplet()
	eighth := fraction.MustNew(1, 8)

	fmt.Println(t.MustCoefficient(), t.MustReciprocalCoefficient())
	fmt.Println(t.MustApply(eighth))
	fmt.Println(t.MustUnapply(fraction.MustNew(1, 12)))
	fmt.Println(tuplet.Duplet().MustApply(eighth))
	fmt.Println(tuplet.MustNew(5, 4).MustSpan(eighth))

	// Output:
	// 2/3 3/2
	// 1/12
	// 1/8
	// 3/16
	// 1/2
}

// ExampleTuplet_Equal shows the two equalities: by notation and by
// coefficient (NF-7: two equalities).
func ExampleTuplet_Equal() {
	t := tuplet.Triplet()
	s := tuplet.Sextuplet()

	fmt.Println(t.Equal(s), t.MustEqualCoefficient(s))
	fmt.Println(s.MustCoefficient(), s.MustReducedCoefficient())
	fmt.Println(t.Compare(s) < 0, t.MustCompareCoefficient(s))
	fmt.Println(s.MustReduce().Equal(t))

	// Output:
	// false true
	// 4/6 2/3
	// true 0
	// true
}

// ExampleFlatten shows nested tuplets (NF-7: nesting).
func ExampleFlatten() {
	t := tuplet.Triplet()

	fmt.Println(tuplet.MustChainCoefficient(tuplet.Quintuplet(), t))
	fmt.Println(tuplet.MustFlatten(t, t))
	fmt.Println(tuplet.MustApplyChain(fraction.MustNew(1, 8), t, t))

	nested := t.MustNest(tuplet.Duplet())
	fmt.Println(nested, nested.MustKind())

	// Output:
	// 8/15
	// 9:4
	// 1/18
	// 6:6 neutral
}

// ExampleDefault shows default normal counts by division (NF-7: defaults).
func ExampleDefault() {
	fmt.Println(tuplet.MustDefault(7, tuplet.Binary))
	fmt.Println(tuplet.MustDefault(7, tuplet.Ternary))
	fmt.Println(tuplet.MustDefault(2, tuplet.Ternary))

	_, err := tuplet.Default(4, tuplet.Binary)
	fmt.Println(errors.Is(err, tuplet.ErrNoDefault))

	division, ok := tuplet.DivisionOf(6)
	fmt.Println(division, ok)

	// Output:
	// 7:4
	// 7:6
	// 2:3
	// true
	// ternary true
}

// ExampleParse shows text and names (NF-7: text).
func ExampleParse() {
	t := tuplet.MustParse("6:4")
	fmt.Println(t, t.MustShortString())
	fmt.Println(tuplet.MustParseShort("5", tuplet.Binary))
	fmt.Println(tuplet.MustParseName("квинтоль"))
	fmt.Println(tuplet.MustParseName("sestole"))

	data, err := json.Marshal(struct {
		Tuplet tuplet.Tuplet `json:"tuplet"`
	}{Tuplet: tuplet.Triplet()})
	fmt.Println(string(data), err)

	// Output:
	// 6:4 6
	// 5:4
	// 5:4
	// 6:4
	// {"tuplet":"3:2"} <nil>
}

// ExampleFromCoefficient shows creation from a coefficient and from a
// reciprocal coefficient.
func ExampleFromCoefficient() {
	fmt.Println(tuplet.MustFromCoefficient(fraction.MustNew(4, 6)))
	fmt.Println(tuplet.MustFromReciprocalCoefficient(fraction.MustNew(3, 2)))

	// Output:
	// 6:4
	// 3:2
}

// ExampleFromUnitRatio shows "3 eighths in the time of 1 quarter".
func ExampleFromUnitRatio() {
	fmt.Println(tuplet.MustFromUnitRatio(3, 1, fraction.MustNew(2, 1)))

	// Output:
	// 3:2
}

// ExampleNew shows error handling (NF-7: errors).
func ExampleNew() {
	_, err := tuplet.New(0, 2)
	fmt.Println(errors.Is(err, tuplet.ErrInvalidCount))

	var unset tuplet.Tuplet

	_, err = unset.Coefficient()
	fmt.Println(errors.Is(err, tuplet.ErrUnset), unset)

	_, err = tuplet.FromUnitRatio(3, 3, fraction.MustNew(1, 2))
	fmt.Println(errors.Is(err, tuplet.ErrNotRepresentable))

	_, err = tuplet.Parse("3 : 2")
	fmt.Println(errors.Is(err, tuplet.ErrInvalidText))

	// Output:
	// true
	// true Tuplet(unset)
	// true
	// true
}
