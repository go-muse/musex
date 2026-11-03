package spelling_test

import (
	"strconv"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

// The guaranteed range of alteration values, taken from the published
// alteration contract: M is the largest magnitude, Q the largest denominator.
const (
	maxMagnitude   = alteration.MaxMagnitude
	maxDenominator = alteration.MaxDenominator
)

// nameCase is a spelling described by its parts. The parts are the test
// oracle; build creates the spelling under test from them.
type nameCase struct {
	name string
	l    letter.Letter
	a    alteration.Alteration
}

func (c nameCase) build() spelling.Spelling {
	return spelling.MustNew(c.l, c.a)
}

// altCase is a named alteration value.
type altCase struct {
	name string
	a    alteration.Alteration
}

// predefinedCase is one of the 35 ready-made names with its expected parts.
type predefinedCase struct {
	name string
	fn   func() spelling.Spelling
	l    letter.Letter
	a    alteration.Alteration
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func frac(num, den int) alteration.Alteration {
	return alteration.MustFromFraction(num, den)
}

func whole(n int) alteration.Alteration {
	return alteration.MustFromInt(n)
}

// nc builds a nameCase whose name is the expected machine form.
func nc(l letter.Letter, a alteration.Alteration) nameCase {
	return nameCase{name: formOf(l, a), l: l, a: a}
}

// formOf returns the expected machine form of the name (l, a), built from the
// grammar in the architecture: the letter, then, unless a is the natural, the
// sign, the absolute numerator, and "/denominator" when the denominator is
// greater than 1. It uses only the public contracts of letter and alteration.
func formOf(l letter.Letter, a alteration.Alteration) string {
	num, den := a.Fraction()
	if num == 0 {
		return l.String()
	}

	sign := "+"
	if num < 0 {
		sign = "-"
		num = -num
	}

	form := l.String() + sign + strconv.Itoa(num)
	if den > 1 {
		form += "/" + strconv.Itoa(den)
	}

	return form
}

// setN is the checking set N of the acceptance criteria:
// C, C#, Db, B#, Cb, E#, F, Cx, D, Fbb, (A, +1/2), (D, -3/2), (E, +3), (B, -7/2).
func setN() []nameCase {
	return []nameCase{
		nc(letter.C, alteration.Natural()),
		nc(letter.C, alteration.Sharp()),
		nc(letter.D, alteration.Flat()),
		nc(letter.B, alteration.Sharp()),
		nc(letter.C, alteration.Flat()),
		nc(letter.E, alteration.Sharp()),
		nc(letter.F, alteration.Natural()),
		nc(letter.C, alteration.DoubleSharp()),
		nc(letter.D, alteration.Natural()),
		nc(letter.F, alteration.DoubleFlat()),
		nc(letter.A, frac(1, 2)),
		nc(letter.D, frac(-3, 2)),
		nc(letter.E, whole(3)),
		nc(letter.B, frac(-7, 2)),
	}
}

// setS is the set S of alteration values of the acceptance criteria.
func setS() []altCase {
	values := []alteration.Alteration{
		alteration.DoubleFlat(),
		alteration.Flat(),
		frac(-1, 2),
		alteration.Natural(),
		frac(1, 3),
		frac(1, 2),
		alteration.Sharp(),
		frac(3, 2),
		alteration.DoubleSharp(),
		whole(3),
		frac(-7, 2),
	}

	cases := make([]altCase, 0, len(values))
	for _, a := range values {
		cases = append(cases, altCase{name: a.String(), a: a})
	}

	return cases
}

// boundaryNames are the boundary names of criterion K-10 item 4:
// (C, +-M), (C, +-(M - 1)), (C, +-(M - 1/2)), (C, +-1/Q).
func boundaryNames() []nameCase {
	return []nameCase{
		nc(letter.C, whole(maxMagnitude)),
		nc(letter.C, whole(-maxMagnitude)),
		nc(letter.C, whole(maxMagnitude-1)),
		nc(letter.C, whole(-(maxMagnitude - 1))),
		nc(letter.C, frac(2*maxMagnitude-1, 2)),
		nc(letter.C, frac(-(2*maxMagnitude-1), 2)),
		nc(letter.C, frac(1, maxDenominator)),
		nc(letter.C, frac(-1, maxDenominator)),
	}
}

// edgeNames are (C, +-(M*Q - 1)/Q): the values just below the magnitude limit
// with the largest denominator.
func edgeNames() []nameCase {
	return []nameCase{
		nc(letter.C, frac(maxMagnitude*maxDenominator-1, maxDenominator)),
		nc(letter.C, frac(-(maxMagnitude*maxDenominator-1), maxDenominator)),
	}
}

func concatNames(sets ...[]nameCase) []nameCase {
	var all []nameCase
	for _, set := range sets {
		all = append(all, set...)
	}

	return all
}

// predefinedCases lists the 35 ready-made names in the order of the
// acceptance criterion K-23: by letter C D E F G A B and, within a letter,
// double flat, flat, natural, sharp, double sharp.
func predefinedCases() []predefinedCase {
	return []predefinedCase{
		{"CDoubleFlat", spelling.CDoubleFlat, letter.C, alteration.DoubleFlat()},
		{"CFlat", spelling.CFlat, letter.C, alteration.Flat()},
		{"C", spelling.C, letter.C, alteration.Natural()},
		{"CSharp", spelling.CSharp, letter.C, alteration.Sharp()},
		{"CDoubleSharp", spelling.CDoubleSharp, letter.C, alteration.DoubleSharp()},
		{"DDoubleFlat", spelling.DDoubleFlat, letter.D, alteration.DoubleFlat()},
		{"DFlat", spelling.DFlat, letter.D, alteration.Flat()},
		{"D", spelling.D, letter.D, alteration.Natural()},
		{"DSharp", spelling.DSharp, letter.D, alteration.Sharp()},
		{"DDoubleSharp", spelling.DDoubleSharp, letter.D, alteration.DoubleSharp()},
		{"EDoubleFlat", spelling.EDoubleFlat, letter.E, alteration.DoubleFlat()},
		{"EFlat", spelling.EFlat, letter.E, alteration.Flat()},
		{"E", spelling.E, letter.E, alteration.Natural()},
		{"ESharp", spelling.ESharp, letter.E, alteration.Sharp()},
		{"EDoubleSharp", spelling.EDoubleSharp, letter.E, alteration.DoubleSharp()},
		{"FDoubleFlat", spelling.FDoubleFlat, letter.F, alteration.DoubleFlat()},
		{"FFlat", spelling.FFlat, letter.F, alteration.Flat()},
		{"F", spelling.F, letter.F, alteration.Natural()},
		{"FSharp", spelling.FSharp, letter.F, alteration.Sharp()},
		{"FDoubleSharp", spelling.FDoubleSharp, letter.F, alteration.DoubleSharp()},
		{"GDoubleFlat", spelling.GDoubleFlat, letter.G, alteration.DoubleFlat()},
		{"GFlat", spelling.GFlat, letter.G, alteration.Flat()},
		{"G", spelling.G, letter.G, alteration.Natural()},
		{"GSharp", spelling.GSharp, letter.G, alteration.Sharp()},
		{"GDoubleSharp", spelling.GDoubleSharp, letter.G, alteration.DoubleSharp()},
		{"ADoubleFlat", spelling.ADoubleFlat, letter.A, alteration.DoubleFlat()},
		{"AFlat", spelling.AFlat, letter.A, alteration.Flat()},
		{"A", spelling.A, letter.A, alteration.Natural()},
		{"ASharp", spelling.ASharp, letter.A, alteration.Sharp()},
		{"ADoubleSharp", spelling.ADoubleSharp, letter.A, alteration.DoubleSharp()},
		{"BDoubleFlat", spelling.BDoubleFlat, letter.B, alteration.DoubleFlat()},
		{"BFlat", spelling.BFlat, letter.B, alteration.Flat()},
		{"B", spelling.B, letter.B, alteration.Natural()},
		{"BSharp", spelling.BSharp, letter.B, alteration.Sharp()},
		{"BDoubleSharp", spelling.BDoubleSharp, letter.B, alteration.DoubleSharp()},
	}
}

// assertAlteration checks that got is the value want, by Equal and by the
// fraction in lowest terms.
func assertAlteration(t *testing.T, want, got alteration.Alteration) bool {
	t.Helper()

	wantNum, wantDen := want.Fraction()
	gotNum, gotDen := got.Fraction()

	return assert.Truef(t, want.Equal(got), "alteration: want %v, got %v", want, got) &&
		assert.Equal(t, [2]int{wantNum, wantDen}, [2]int{gotNum, gotDen}, "alteration fraction")
}

// assertName checks that got is a set name with the letter wantLetter and
// the value wantValue, through its parts.
func assertName(t *testing.T, wantLetter letter.Letter, wantValue alteration.Alteration, got spelling.Spelling) bool {
	t.Helper()

	ok := assert.True(t, got.IsValid(), "the name must be set")
	ok = assert.Equal(t, wantLetter, got.Letter(), "letter") && ok
	ok = assertAlteration(t, wantValue, got.Alteration()) && ok

	return ok
}

// assertUnset checks that got is the unset name: the zero Spelling.
func assertUnset(t *testing.T, got spelling.Spelling) bool {
	t.Helper()

	return assert.Equal(t, spelling.Spelling{}, got, "the result must be the unset name") &&
		assert.False(t, got.IsValid(), "IsValid of the unset name")
}

// assertOutOfRange checks that err reports a result outside the guaranteed
// range and no other error of the contract.
func assertOutOfRange(t *testing.T, err error) bool {
	t.Helper()

	if !assert.ErrorIs(t, err, alteration.ErrOutOfRange) {
		return false
	}

	ok := assert.NotErrorIs(t, err, letter.ErrInvalid)
	ok = assert.NotErrorIs(t, err, spelling.ErrInvalidForm) && ok
	ok = assert.NotErrorIs(t, err, alteration.ErrZeroDenominator) && ok

	return ok
}

// assertInvalidLetter checks that err reports an invalid letter or an unset
// name and no other error of the contract.
func assertInvalidLetter(t *testing.T, err error) bool {
	t.Helper()

	if !assert.ErrorIs(t, err, letter.ErrInvalid) {
		return false
	}

	ok := assert.NotErrorIs(t, err, alteration.ErrOutOfRange)
	ok = assert.NotErrorIs(t, err, spelling.ErrInvalidForm) && ok
	ok = assert.NotErrorIs(t, err, alteration.ErrZeroDenominator) && ok

	return ok
}

// formErrorClass names the secondary alteration error, if any, that an error
// of reading the machine form must carry.
type formErrorClass int

const (
	formOnly formErrorClass = iota
	formZeroDenominator
	formOutOfRange
)

// assertInvalidForm checks that err is a reading error of the class.
func assertInvalidForm(t *testing.T, class formErrorClass, err error) bool {
	t.Helper()

	if !assert.ErrorIs(t, err, spelling.ErrInvalidForm) {
		return false
	}

	ok := assert.NotErrorIs(t, err, letter.ErrInvalid)

	switch class {
	case formOnly:
		ok = assert.NotErrorIs(t, err, alteration.ErrOutOfRange) && ok
		ok = assert.NotErrorIs(t, err, alteration.ErrZeroDenominator) && ok
	case formZeroDenominator:
		ok = assert.ErrorIs(t, err, alteration.ErrZeroDenominator) && ok
		ok = assert.NotErrorIs(t, err, alteration.ErrOutOfRange) && ok
	case formOutOfRange:
		ok = assert.ErrorIs(t, err, alteration.ErrOutOfRange) && ok
		ok = assert.NotErrorIs(t, err, alteration.ErrZeroDenominator) && ok
	}

	return ok
}

// catchPanic runs f and reports whether it panicked and with which value.
func catchPanic(f func()) (value any, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			value = r
			panicked = true
		}
	}()

	f()

	return nil, false
}

// panicError checks that f panics with an error value and returns that
// error, or nil after reporting a failure. The text of the error is not part
// of the contract and is not checked.
func panicError(t *testing.T, f func()) error {
	t.Helper()

	value, panicked := catchPanic(f)
	if !assert.True(t, panicked, "expected a panic") {
		return nil
	}

	err, ok := value.(error)
	if !assert.Truef(t, ok, "the panic value %#v is not an error", value) {
		return nil
	}

	return err
}

// assertPanicsWithErrorIs checks that f panics with an error value that
// matches target with errors.Is.
func assertPanicsWithErrorIs(t *testing.T, target error, f func()) bool {
	t.Helper()

	err := panicError(t, f)

	return err != nil && assert.ErrorIs(t, err, target)
}

// directionOf is the expected Direction of a set name with the value a,
// from the sign of the value.
func directionOf(a alteration.Alteration) spelling.Direction {
	switch a.Sign() {
	case -1:
		return spelling.DirectionLowered
	case 1:
		return spelling.DirectionRaised
	default:
		return spelling.DirectionNatural
	}
}

// compareOracle is the expected contractual order of two set names: by the
// ordinal of the letter, then by the alteration value.
func compareOracle(x, y nameCase) int {
	switch {
	case x.l.Ordinal() < y.l.Ordinal():
		return -1
	case x.l.Ordinal() > y.l.Ordinal():
		return 1
	default:
		return x.a.Compare(y.a)
	}
}
