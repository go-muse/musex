package number

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

// Errors returned by the package. Each returned error matches exactly one of
// them with [errors.Is]; it may be the error itself or wrap it with context.
// The text of the errors is not part of the API.
var (
	// ErrSkipped reports that a number is skipped in a numbering: no octave
	// has this number. It is returned by [Numbering.ToASPN] and [Convert].
	ErrSkipped = errors.New("number: number is skipped in the numbering")

	// ErrInvalidDefinition reports an invalid definition of an own numbering:
	// the anchor number is among the skipped numbers, or a skipped number is
	// repeated. It is returned by [New].
	ErrInvalidDefinition = errors.New("number: invalid numbering definition")

	// ErrOutOfRange reports that the exact result of an operation does not fit
	// in an int. It is returned by [Numbering.FromASPN], [Numbering.ToASPN],
	// and [Convert]. The limits of int are not musical limits.
	ErrOutOfRange = errors.New("number: number out of range")

	// ErrUndefined reports that a numbering is the zero [Numbering], which is
	// not a numbering. It is returned by [Numbering.FromASPN],
	// [Numbering.ToASPN], and [Convert].
	ErrUndefined = errors.New("number: undefined numbering")
)

// Labels of the built-in numberings.
const (
	labelASPN          = "ASPN"
	labelFrancoBelgian = "Franco-Belgian"
	labelMexican       = "Mexican"
	labelMiddleC3      = "Middle C = C3"
	labelMiddleC5      = "Middle C = C5"
)

// errorSkippedLimit bounds the skipped numbers shown in an error description.
const errorSkippedLimit = 4

// francoBelgianSkipped holds the skipped numbers of the Franco-Belgian
// numbering. It is read-only and never returned to callers.
var francoBelgianSkipped = []int{0}

// Numbering is an octave numbering: a rule that gives each octave, identified
// by its ASPN number, a visible number.
//
// A numbering is defined by an anchor pair (the octave with ASPN number a
// has number x) and a finite set of skipped numbers. Octave numbers strictly
// increase with the ASPN number, take every integer except the skipped
// numbers, and continue without limit in both directions. A built-in
// numbering also has a label; an own numbering created by [New] has none.
//
// Create a numbering with [ASPN], [FrancoBelgian], [Mexican], [MiddleC3],
// [MiddleC5], [Builtins], [New], or [MustNew]. The zero value is not a
// numbering: it is undefined, it is never treated as ASPN or any other
// numbering, and [Numbering.IsDefined] reports false for it. Operations that
// return a number return [ErrUndefined] for it.
//
// Numbering is an immutable value. It can be copied by assignment and passed
// by value at any time; a copy behaves like the original and is equal to it.
// No function or method changes a numbering, its arguments, or the data of
// the package.
//
// The type is not comparable: == does not compile and a Numbering cannot be
// a map key. Use [Numbering.Equal], which compares the numbers given to all
// octaves and ignores the label. A struct with a Numbering field is not
// comparable either.
//
// Numbering has no exported fields and no serialization of its own:
// encoding/json writes {} for it. Store a built-in numbering by its label.
type Numbering struct {
	// skipped holds the skipped numbers in ascending order, without
	// duplicates. It may be shared between copies but is never modified
	// after creation.
	skipped []int
	// offset is r(x) - a for every octave with ASPN number a and number x,
	// where r(x) is x minus the count of skipped numbers below x. The pair
	// (skipped, offset) determines the numbers of all octaves.
	offset int128
	// aspn and num are the anchor pair as given; they are used by String.
	aspn, num int
	// label is the label of a built-in numbering, or "" for an own one.
	label string
	// defined is false only for the zero value.
	defined bool
}

// define returns the numbering with the anchor pair aspn -> num, the skipped
// numbers sorted in ascending order without duplicates, and the label.
// The caller must ensure that num is not skipped.
func define(aspn, num int, sorted []int, label string) Numbering {
	n := Numbering{
		skipped: sorted,
		aspn:    aspn,
		num:     num,
		label:   label,
		defined: true,
	}
	n.offset = n.rank(num).sub(int128Of(aspn))

	return n
}

// ASPN returns American Standard Pitch Notation octave numbering, also known
// as SPN and IPN.
//
// Label "ASPN"; the octave of middle C is 4; no skipped numbers; the number
// of an octave equals its ASPN number. It is used in English-language
// literature and in science. Labels on Roland and Korg instruments match
// ASPN.
func ASPN() Numbering {
	return define(4, 4, nil, labelASPN)
}

// FrancoBelgian returns the Franco-Belgian octave numbering.
//
// Label "Franco-Belgian"; the octave of middle C is 3; number 0 is skipped:
// ASPN 2 is number 1 and ASPN 1 is number -1. The number is the ASPN number
// minus 1 for ASPN 2 and above, and minus 2 for ASPN 1 and below. It is used
// in France, Belgium, and parts of Spain; in Italy it is prescribed by a law
// of 1989, but ASPN is common there in practice.
//
// Sources confirm the numbers -2 to 7. Below -2, the numbers continue in
// sequence without confirmation by sources; above 7, the source sets no
// limit, and the numbers continue in sequence.
func FrancoBelgian() Numbering {
	return define(4, 3, francoBelgianSkipped, labelFrancoBelgian)
}

// Mexican returns the Mexican octave numbering, the numbering of the Spanish
// translations of Riemann.
//
// Label "Mexican"; the octave of middle C is 5; no skipped numbers; the
// number is the ASPN number plus 1. It is used in Mexico alongside ASPN.
//
// A single encyclopedia article confirms the numbers from 1 (the lowest
// audible C, ASPN 0). Below 1, the numbers continue in sequence without
// confirmation by sources. Mexican gives the same numbers as [MiddleC5], so
// the two are equal by [Numbering.Equal], but their labels differ.
func Mexican() Numbering {
	return define(4, 5, nil, labelMexican)
}

// MiddleC3 returns the octave numbering in which middle C is C3.
//
// Label "Middle C = C3"; the octave of middle C is 3; no skipped numbers;
// the number is the ASPN number minus 1. It is used by Yamaha instruments
// and by default in Logic Pro, and reportedly in Cubase.
func MiddleC3() Numbering {
	return define(4, 3, nil, labelMiddleC3)
}

// MiddleC5 returns the octave numbering in which middle C is C5.
//
// Label "Middle C = C5"; the octave of middle C is 5; no skipped numbers;
// the number is the ASPN number plus 1. It is used in FL Studio. It gives
// the same numbers as [Mexican].
func MiddleC5() Numbering {
	return define(4, 5, nil, labelMiddleC5)
}

// Builtins returns the five built-in numberings in a fixed order: [ASPN],
// [FrancoBelgian], [Mexican], [MiddleC3], and [MiddleC5]. Each element is
// equal to the result of the corresponding function and has the same label.
//
// Each call returns a new slice with length and capacity 5. The caller owns
// it and may modify or retain it; this does not affect later calls, the
// built-in numberings, or other callers. The list is not a registry: own
// numberings cannot be added to it.
func Builtins() []Numbering {
	return []Numbering{ASPN(), FrancoBelgian(), Mexican(), MiddleC3(), MiddleC5()}
}

// New returns an own numbering in which the octave with ASPN number aspn has
// number num and the numbers in skipped are skipped.
//
// Any int values are accepted. The skipped numbers may be given in any order,
// on either side of num, next to each other or not; without skipped numbers,
// the numbering is a shift: number = ASPN number + (num - aspn). For example,
// New(4, 3) equals [MiddleC3], and New(4, 3, 0) equals [FrancoBelgian].
//
// New copies skipped: changing the slice after the call does not affect the
// numbering. The result has no label, even when it equals a built-in
// numbering.
//
// If num is among the skipped numbers or a skipped number is repeated, New
// returns the zero Numbering and an error that wraps [ErrInvalidDefinition].
func New(aspn, num int, skipped ...int) (Numbering, error) {
	var sorted []int
	if len(skipped) > 0 {
		sorted = slices.Clone(skipped)
		slices.Sort(sorted)
	}

	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			return Numbering{}, fmt.Errorf("%w: skipped number %d is repeated", ErrInvalidDefinition, sorted[i])
		}
	}

	if _, found := slices.BinarySearch(sorted, num); found {
		return Numbering{}, fmt.Errorf("%w: anchor number %d is skipped", ErrInvalidDefinition, num)
	}

	return define(aspn, num, sorted, ""), nil
}

// MustNew is like [New] but panics if New returns an error. The panic value
// is that error, which wraps [ErrInvalidDefinition].
//
// It is intended for literals and values known to be valid.
func MustNew(aspn, num int, skipped ...int) Numbering {
	n, err := New(aspn, num, skipped...)
	if err != nil {
		panic(err)
	}

	return n
}

// Convert returns the number in numbering to of the octave that has number x
// in numbering from. The result is exact: it is returned whenever it fits in
// an int, even if the ASPN number of the octave does not.
//
// Convert(x, A, A) returns x for every number x used by A, and converting
// the result back returns x. Convert(x, A, ASPN()) equals A.ToASPN(x), and
// Convert(a, ASPN(), B) equals B.FromASPN(a).
//
// On error, Convert returns 0 and an error. The conditions are checked in
// this order:
//
//  1. from or to is undefined: the error wraps [ErrUndefined];
//  2. x is skipped in from: the error wraps [ErrSkipped];
//  3. the result does not fit in an int: the error wraps [ErrOutOfRange].
func Convert(x int, from, to Numbering) (int, error) {
	if !from.defined || !to.defined {
		return 0, ErrUndefined
	}

	if from.isSkipped(x) {
		return 0, fmt.Errorf("%w: number %d in %s", ErrSkipped, x, from.errorDescription())
	}

	// The rank of the octave in to: from.rank(x) - from.offset is its ASPN
	// number, and adding to.offset gives its rank in to.
	result, ok := to.unrank(from.rank(x).sub(from.offset).add(to.offset)).toInt()
	if !ok {
		return 0, fmt.Errorf("%w: number %d converted from %s to %s",
			ErrOutOfRange, x, from.errorDescription(), to.errorDescription())
	}

	return result, nil
}

// MustConvert is like [Convert] but panics if Convert returns an error. The
// panic value is that error.
//
// It is intended for literals and values known to be valid.
func MustConvert(x int, from, to Numbering) int {
	result, err := Convert(x, from, to)
	if err != nil {
		panic(err)
	}

	return result
}

// FromASPN returns the number in n of the octave with ASPN number aspn.
// The result is never a skipped number and strictly increases with aspn.
//
// On error, FromASPN returns 0 and an error: if n is undefined, the error
// wraps [ErrUndefined]; if the result does not fit in an int, the error wraps
// [ErrOutOfRange].
func (n Numbering) FromASPN(aspn int) (int, error) {
	if !n.defined {
		return 0, ErrUndefined
	}

	result, ok := n.unrank(int128Of(aspn).add(n.offset)).toInt()
	if !ok {
		return 0, fmt.Errorf("%w: ASPN %d in %s", ErrOutOfRange, aspn, n.errorDescription())
	}

	return result, nil
}

// MustFromASPN is like [Numbering.FromASPN] but panics if FromASPN returns
// an error. The panic value is that error.
//
// It is intended for literals and values known to be valid.
func (n Numbering) MustFromASPN(aspn int) int {
	result, err := n.FromASPN(aspn)
	if err != nil {
		panic(err)
	}

	return result
}

// ToASPN returns the ASPN number of the octave that has number num in n.
// For every aspn for which FromASPN succeeds, n.ToASPN(n.MustFromASPN(aspn))
// returns aspn.
//
// On error, ToASPN returns 0 and an error. The conditions are checked in this
// order:
//
//  1. n is undefined: the error wraps [ErrUndefined];
//  2. num is skipped in n: the error wraps [ErrSkipped]; no nearby number
//     is used instead;
//  3. the result does not fit in an int: the error wraps [ErrOutOfRange].
func (n Numbering) ToASPN(num int) (int, error) {
	if !n.defined {
		return 0, ErrUndefined
	}

	if n.isSkipped(num) {
		return 0, fmt.Errorf("%w: number %d in %s", ErrSkipped, num, n.errorDescription())
	}

	result, ok := n.rank(num).sub(n.offset).toInt()
	if !ok {
		return 0, fmt.Errorf("%w: number %d in %s", ErrOutOfRange, num, n.errorDescription())
	}

	return result, nil
}

// MustToASPN is like [Numbering.ToASPN] but panics if ToASPN returns an
// error. The panic value is that error.
//
// It is intended for literals and values known to be valid.
func (n Numbering) MustToASPN(num int) int {
	result, err := n.ToASPN(num)
	if err != nil {
		panic(err)
	}

	return result
}

// Equal reports whether n and other are both defined and give the same
// number to every octave. Labels and the way the numberings were created are
// ignored: Mexican().Equal(MiddleC5()) and New(4, 3, 0).Equal(FrancoBelgian())
// are true.
//
// An undefined numbering is not equal to any numbering, including itself.
// For defined numberings, Equal is reflexive, symmetric, and transitive.
func (n Numbering) Equal(other Numbering) bool {
	return n.defined && other.defined &&
		n.offset == other.offset &&
		slices.Equal(n.skipped, other.skipped)
}

// IsDefined reports whether n is a numbering. It returns false only for the
// zero Numbering, including the result of [New] with an error.
func (n Numbering) IsDefined() bool {
	return n.defined
}

// Label returns the label of a built-in numbering: "ASPN", "Franco-Belgian",
// "Mexican", "Middle C = C3", or "Middle C = C5". The label of a built-in
// numbering is never empty.
//
// Label returns "" for an own numbering created by [New] and for an
// undefined numbering; use [Numbering.IsDefined] to tell them apart.
func (n Numbering) Label() string {
	return n.label
}

// String returns a fixed form of n for output, logs, and debugging:
//
//   - a built-in numbering: its label, such as "Franco-Belgian";
//   - an own numbering without skipped numbers: "ASPN <aspn> -> <num>" with
//     the arguments of [New], such as "ASPN 4 -> 5";
//   - an own numbering with skipped numbers: the same form followed by
//     ", skip " and the skipped numbers in ascending order, separated by
//     spaces in square brackets, such as "ASPN 4 -> 4, skip [0 1]";
//   - an undefined numbering: "Numbering(undefined)".
//
// Equal numberings may have different forms, so String is not a test for
// equality. It is not a label or a notation and has no parser.
func (n Numbering) String() string {
	switch {
	case !n.defined:
		return "Numbering(undefined)"
	case n.label != "":
		return n.label
	}

	b := make([]byte, 0, 32)
	b = append(b, "ASPN "...)
	b = strconv.AppendInt(b, int64(n.aspn), 10)
	b = append(b, " -> "...)
	b = strconv.AppendInt(b, int64(n.num), 10)

	if len(n.skipped) > 0 {
		b = append(b, ", skip ["...)
		for i, s := range n.skipped {
			if i > 0 {
				b = append(b, ' ')
			}

			b = strconv.AppendInt(b, int64(s), 10)
		}

		b = append(b, ']')
	}

	return string(b)
}

// errorDescription returns a bounded description without formatting the full
// skipped list of a long own numbering.
func (n Numbering) errorDescription() string {
	if n.label != "" || len(n.skipped) <= errorSkippedLimit {
		return n.String()
	}

	return fmt.Sprintf("ASPN %d -> %d, skip [%d %d %d %d ...] (%d skipped)",
		n.aspn, n.num, n.skipped[0], n.skipped[1], n.skipped[2], n.skipped[3], len(n.skipped))
}

// isSkipped reports whether x is a skipped number of n.
func (n Numbering) isSkipped(x int) bool {
	_, found := slices.BinarySearch(n.skipped, x)

	return found
}

// rank returns r(x) = x minus the count of skipped numbers below x. On the
// numbers used by n, r is a strictly increasing bijection onto all integers.
func (n Numbering) rank(x int) int128 {
	below, _ := slices.BinarySearch(n.skipped, x)

	return int128Of(x).sub(int128Of(below))
}

// unrank returns the used number x with rank r(x) = r; it is the inverse of
// rank on the numbers used by n.
func (n Numbering) unrank(r int128) int128 {
	x := r
	for _, s := range n.skipped {
		if int128Of(s).cmp(x) > 0 {
			break
		}

		x = x.add(one)
	}

	return x
}
