package octave_test

import (
	"fmt"
	"iter"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
)

// The tests assume a 64-bit int (the project supports only 64-bit systems).
const (
	minInt = math.MinInt
	maxInt = math.MaxInt
)

// oct returns the octave with ASPN number a. By the contract it is the octave
// with that ASPN number for every int, so it serves as the reference octave
// O(a) of the criteria.
func oct(a int) octave.Octave {
	return octave.FromASPN(a)
}

// span returns the integers lo...hi in ascending order.
func span(lo, hi int) []int {
	out := make([]int, 0, hi-lo+1)
	for v := lo; v <= hi; v++ {
		out = append(out, v)
	}

	return out
}

// edgeASPN returns the ASPN numbers at the limits of int, around zero, the
// octave of middle C, and far octaves (K-23, K-26, K-31).
func edgeASPN() []int {
	return []int{
		minInt, minInt + 1, minInt + 2,
		-1000, -1, 0, 4, 1000,
		maxInt - 2, maxInt - 1, maxInt,
	}
}

// edgeShifts returns shift amounts at the limits of int and around zero.
func edgeShifts() []int {
	return []int{
		minInt, minInt + 1, minInt + 2,
		-1000, -2, -1, 0, 1, 2, 1000,
		maxInt - 2, maxInt - 1, maxInt,
	}
}

// exactAdd returns a + k and whether the exact sum fits in an int. It is the
// oracle of Shift: an independent check of overflow in plain integers.
func exactAdd(a, k int) (int, bool) {
	s := a + k
	if (k >= 0 && s < a) || (k < 0 && s >= a) {
		return 0, false
	}

	return s, true
}

// exactSub returns a - k and whether the exact difference fits in an int.
// It is the oracle of ShiftDown and Sub, including k = math.MinInt.
func exactSub(a, k int) (int, bool) {
	d := a - k
	if (k > 0 && d >= a) || (k <= 0 && d < a) {
		return 0, false
	}

	return d, true
}

// sign returns -1, 0 or +1 by the sign of v.
func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	default:
		return 0
	}
}

// numberingCase is a numbering under test with its oracle: the formula of the
// table of built-in numberings in the living architecture (octave/number) or
// the definition of an own numbering. number returns the number of the octave
// with ASPN number a and false when the exact number does not fit in an int.
type numberingCase struct {
	name    string
	n       number.Numbering
	number  func(a int) (int, bool)
	skipped []int
}

// shiftBy returns the formula "ASPN number + d".
func shiftBy(d int) func(int) (int, bool) {
	return func(a int) (int, bool) {
		return exactAdd(a, d)
	}
}

// splitAt2 returns the formula of a numbering with skipped numbers between
// the octaves with ASPN numbers 1 and 2: "a - high" for a >= 2 and "a - low"
// for a <= 1. It covers Franco-Belgian (1, 2) and New(4, 4, 0, 1) (0, 2).
func splitAt2(high, low int) func(int) (int, bool) {
	return func(a int) (int, bool) {
		if a >= 2 {
			return exactAdd(a, -high)
		}

		return exactAdd(a, -low)
	}
}

// builtinNumberings returns the five built-in numberings in the order of
// number.Builtins with their formulas: ASPN a; Franco-Belgian a - 1 for
// a >= 2 and a - 2 for a <= 1; Mexican a + 1; Middle C = C3 a - 1;
// Middle C = C5 a + 1.
func builtinNumberings() []numberingCase {
	return []numberingCase{
		{name: "ASPN", n: number.ASPN(), number: shiftBy(0)},
		{name: "Franco-Belgian", n: number.FrancoBelgian(), number: splitAt2(1, 2), skipped: []int{0}},
		{name: "Mexican", n: number.Mexican(), number: shiftBy(1)},
		{name: "Middle C = C3", n: number.MiddleC3(), number: shiftBy(-1)},
		{name: "Middle C = C5", n: number.MiddleC5(), number: shiftBy(1)},
	}
}

// ownNumberings returns the own numberings of K-1: ASPN 4 -> 5 (a + 1) and
// ASPN 4 -> 4 with 0 and 1 skipped (a for a >= 2, a - 2 for a <= 1).
func ownNumberings() []numberingCase {
	return []numberingCase{
		{name: "New(4, 5)", n: number.MustNew(4, 5), number: shiftBy(1)},
		{name: "New(4, 4, 0, 1)", n: number.MustNew(4, 4, 0, 1), number: splitAt2(0, 2), skipped: []int{0, 1}},
	}
}

// criteriaNumberings returns the built-in and own numberings of K-1.
func criteriaNumberings() []numberingCase {
	return append(builtinNumberings(), ownNumberings()...)
}

// fb, mex, c3 and c5 are short names of the built-in numberings.
func fb() number.Numbering  { return number.FrancoBelgian() }
func mex() number.Numbering { return number.Mexican() }
func c3() number.Numbering  { return number.MiddleC3() }
func c5() number.Numbering  { return number.MiddleC5() }

// The names of the built-in catalogs by ASPN number, exactly as in the
// criteria K-1...K-3 of octave/naming (living requirements).
func russianNames() map[int]string {
	return map[int]string{
		-1: "субсубконтроктава",
		0:  "субконтроктава",
		1:  "контроктава",
		2:  "большая",
		3:  "малая",
		4:  "первая",
		5:  "вторая",
		6:  "третья",
		7:  "четвёртая",
		8:  "пятая",
		9:  "шестая",
	}
}

func englishNames() map[int]string {
	return map[int]string{
		0: "subcontra",
		1: "contra",
		2: "great",
		3: "small",
		4: "one-line",
		5: "two-line",
		6: "three-line",
		7: "four-line",
	}
}

func germanNames() map[int]string {
	return map[int]string{
		0: "Subkontraoktave",
		1: "Kontraoktave",
		2: "große",
		3: "kleine",
		4: "eingestrichene",
		5: "zweigestrichene",
		6: "dreigestrichene",
		7: "viergestrichene",
	}
}

// catalogCase is a catalog under test with the names it gives by ASPN number.
type catalogCase struct {
	name  string
	c     *naming.Catalog
	names map[int]string
}

// builtinCatalogs returns the three built-in catalogs with their names.
// Each call returns new catalogs and maps.
func builtinCatalogs() []catalogCase {
	return []catalogCase{
		{name: "Russian", c: naming.Russian(), names: russianNames()},
		{name: "English", c: naming.English(), names: englishNames()},
		{name: "German", c: naming.German(), names: germanNames()},
	}
}

// emptyCatalogs returns catalogs without names: created from nil and from an
// empty map, the zero Catalog and the nil *Catalog (octave/naming contract).
func emptyCatalogs() []catalogCase {
	return []catalogCase{
		{name: "naming.MustNew(nil)", c: naming.MustNew(nil)},
		{name: "naming.MustNew(empty map)", c: naming.MustNew(map[int]string{})},
		{name: "&naming.Catalog{}", c: &naming.Catalog{}},
		{name: "nil *naming.Catalog", c: nil},
	}
}

// cause is an expected cause of an error: the sentinel errors it must match
// with errors.Is. Every other sentinel of allSentinels must not match.
type cause struct {
	name  string
	match []error
}

func (c cause) String() string {
	return c.name
}

// The seven causes of the contract (architecture, "Errors"). An out-of-range
// error found by octave/number matches both octave.ErrOutOfRange and
// number.ErrOutOfRange; one found by the arithmetic of octave matches only
// octave.ErrOutOfRange.
var (
	causeInvalid           = cause{name: "ErrInvalid", match: []error{octave.ErrInvalid}}
	causeOutOfRange        = cause{name: "ErrOutOfRange", match: []error{octave.ErrOutOfRange}}
	causeNumberOutOfRange  = cause{name: "ErrOutOfRange+number.ErrOutOfRange", match: []error{octave.ErrOutOfRange, number.ErrOutOfRange}}
	causeUndefined         = cause{name: "number.ErrUndefined", match: []error{number.ErrUndefined}}
	causeSkipped           = cause{name: "number.ErrSkipped", match: []error{number.ErrSkipped}}
	causeInvalidDefinition = cause{name: "number.ErrInvalidDefinition", match: []error{number.ErrInvalidDefinition}}
	causeUnknownName       = cause{name: "naming.ErrUnknownName", match: []error{naming.ErrUnknownName}}
	causeDuplicateName     = cause{name: "naming.ErrDuplicateName", match: []error{naming.ErrDuplicateName}}
)

// allSentinels returns the sentinel errors that identify the causes: the
// seven causes of the contract and number.ErrOutOfRange.
func allSentinels() []error {
	return []error{
		octave.ErrInvalid,
		octave.ErrOutOfRange,
		number.ErrUndefined,
		number.ErrSkipped,
		number.ErrInvalidDefinition,
		naming.ErrUnknownName,
		naming.ErrDuplicateName,
		number.ErrOutOfRange,
	}
}

// sentinelName returns a readable name of a sentinel error for messages.
func sentinelName(err error) string {
	names := []string{
		"octave.ErrInvalid",
		"octave.ErrOutOfRange",
		"number.ErrUndefined",
		"number.ErrSkipped",
		"number.ErrInvalidDefinition",
		"naming.ErrUnknownName",
		"naming.ErrDuplicateName",
		"number.ErrOutOfRange",
	}
	if i := slices.Index(allSentinels(), err); i >= 0 {
		return names[i]
	}

	return fmt.Sprint(err)
}

// assertCause checks that err is not nil, matches every sentinel of want with
// errors.Is and none of the other sentinels (K-25). The text of the error is
// not part of the contract and is not compared.
func assertCause(t *testing.T, err error, want cause, desc string) bool {
	t.Helper()

	if !assert.Error(t, err, "%s: want %s", desc, want) {
		return false
	}

	ok := true
	for _, s := range allSentinels() {
		if slices.Contains(want.match, s) {
			ok = assert.ErrorIs(t, err, s, "%s: want %s", desc, want) && ok
		} else {
			ok = assert.NotErrorIs(t, err, s, "%s: want only %s, not %s", desc, want, sentinelName(s)) && ok
		}
	}

	return ok
}

// panicValue calls f and returns the recovered value and whether f panicked.
func panicValue(f func()) (recovered any, panicked bool) {
	defer func() {
		recovered = recover()
	}()

	panicked = true
	f()
	panicked = false

	return nil, panicked
}

// assertPanicsCause checks that f panics with an error value of the cause
// want (Must variants, K-25). assert.PanicsWithError is not used: it compares
// the text of the error, which is not part of the contract.
func assertPanicsCause(t *testing.T, want cause, f func(), desc string) bool {
	t.Helper()

	v, panicked := panicValue(f)
	if !assert.True(t, panicked, "%s: expected a panic with %s", desc, want) {
		return false
	}

	err, ok := v.(error)
	if !assert.True(t, ok, "%s: panic value %v (%T) is not an error", desc, v, v) {
		return false
	}

	return assertCause(t, err, want, desc+" (panic value)")
}

// collect returns the first limit elements of seq, or all of them if there
// are fewer, and stops the iteration after limit elements.
func collect(seq iter.Seq[octave.Octave], limit int) []octave.Octave {
	var out []octave.Octave
	for o := range seq {
		out = append(out, o)
		if len(out) >= limit {
			break
		}
	}

	return out
}

// octs returns the reference octaves O(a) for the given ASPN numbers.
func octs(aspn ...int) []octave.Octave {
	out := make([]octave.Octave, 0, len(aspn))
	for _, a := range aspn {
		out = append(out, oct(a))
	}

	return out
}

// describe returns a readable list of octaves by their ASPN numbers for
// messages; an invalid value is shown as "Z". It uses only the documented
// error of ASPN, not String.
func describe(list []octave.Octave) string {
	parts := make([]string, 0, len(list))
	for _, o := range list {
		a, err := o.ASPN()
		if err != nil {
			parts = append(parts, "Z")

			continue
		}
		parts = append(parts, fmt.Sprint(a))
	}

	return "[" + strings.Join(parts, " ") + "]"
}
