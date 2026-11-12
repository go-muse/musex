package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// mustCase pairs an operation with its Must variant on the same inputs.
// base returns the results without the error; must returns the results of
// the Must variant. want is nil for a successful call.
type mustCase struct {
	name   string
	base   func() ([]any, error)
	must   func() []any
	want   error
	causes []error
}

// mustCases lists a successful and a failing call for every Must variant of
// summary 4.1 (K-14, O-17).
func mustCases(t *testing.T) []mustCase {
	t.Helper()

	var unset tuplet.Tuplet

	tr := tuplet.Triplet()
	d := frac(1, 8)
	maxFrac := maxComponentFrac(t)
	maxInt := tuplet.MustNew(1, math.MaxInt)
	five := []tuplet.Tuplet{maxInt, maxInt, maxInt, maxInt, maxInt}
	outOfRange := []error{fraction.ErrOutOfRange}

	r1 := func(v any) []any { return []any{v} }
	r2 := func(a, b any) []any { return []any{a, b} }

	return []mustCase{
		{name: "New", base: func() ([]any, error) { v, err := tuplet.New(3, 2); return r1(v), err },
			must: func() []any { return r1(tuplet.MustNew(3, 2)) }},
		{name: "New/error", base: func() ([]any, error) { v, err := tuplet.New(0, 2); return r1(v), err },
			must: func() []any { return r1(tuplet.MustNew(0, 2)) }, want: tuplet.ErrInvalidCount},

		{name: "FromCoefficient", base: func() ([]any, error) { v, err := tuplet.FromCoefficient(frac(4, 6)); return r1(v), err },
			must: func() []any { return r1(tuplet.MustFromCoefficient(frac(4, 6))) }},
		{name: "FromCoefficient/error", base: func() ([]any, error) {
			v, err := tuplet.FromCoefficient(fraction.Fraction{})
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustFromCoefficient(fraction.Fraction{})) }, want: tuplet.ErrInvalidArgument},

		{name: "FromReciprocalCoefficient", base: func() ([]any, error) {
			v, err := tuplet.FromReciprocalCoefficient(frac(6, 4))
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustFromReciprocalCoefficient(frac(6, 4))) }},
		{name: "FromReciprocalCoefficient/error", base: func() ([]any, error) {
			v, err := tuplet.FromReciprocalCoefficient(frac(0, 1))
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustFromReciprocalCoefficient(frac(0, 1))) }, want: tuplet.ErrInvalidCount},

		{name: "FromUnitRatio", base: func() ([]any, error) {
			v, err := tuplet.FromUnitRatio(3, 1, frac(2, 1))
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustFromUnitRatio(3, 1, frac(2, 1))) }},
		{name: "FromUnitRatio/error", base: func() ([]any, error) {
			v, err := tuplet.FromUnitRatio(3, 1, frac(1, 2))
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustFromUnitRatio(3, 1, frac(1, 2))) }, want: tuplet.ErrNotRepresentable},
		{name: "FromUnitRatio/range", base: func() ([]any, error) {
			v, err := tuplet.FromUnitRatio(1, math.MaxInt, frac(2, 1))
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustFromUnitRatio(1, math.MaxInt, frac(2, 1))) },
			want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrNotRepresentable}},

		{name: "Default", base: func() ([]any, error) { v, err := tuplet.Default(5, tuplet.Ternary); return r1(v), err },
			must: func() []any { return r1(tuplet.MustDefault(5, tuplet.Ternary)) }},
		{name: "Default/error", base: func() ([]any, error) { v, err := tuplet.Default(8, tuplet.Binary); return r1(v), err },
			must: func() []any { return r1(tuplet.MustDefault(8, tuplet.Binary)) }, want: tuplet.ErrNoDefault},

		{name: "Actual", base: func() ([]any, error) { v, err := tr.Actual(); return r1(v), err },
			must: func() []any { return r1(tr.MustActual()) }},
		{name: "Actual/error", base: func() ([]any, error) { v, err := unset.Actual(); return r1(v), err },
			must: func() []any { return r1(unset.MustActual()) }, want: tuplet.ErrUnset},

		{name: "Normal", base: func() ([]any, error) { v, err := tr.Normal(); return r1(v), err },
			must: func() []any { return r1(tr.MustNormal()) }},
		{name: "Normal/error", base: func() ([]any, error) { v, err := unset.Normal(); return r1(v), err },
			must: func() []any { return r1(unset.MustNormal()) }, want: tuplet.ErrUnset},

		{name: "Pair", base: func() ([]any, error) { a, n, err := tr.Pair(); return r2(a, n), err },
			must: func() []any { a, n := tr.MustPair(); return r2(a, n) }},
		{name: "Pair/error", base: func() ([]any, error) { a, n, err := unset.Pair(); return r2(a, n), err },
			must: func() []any { a, n := unset.MustPair(); return r2(a, n) }, want: tuplet.ErrUnset},

		{name: "Kind", base: func() ([]any, error) { v, err := tr.Kind(); return r1(v), err },
			must: func() []any { return r1(tr.MustKind()) }},
		{name: "Kind/error", base: func() ([]any, error) { v, err := unset.Kind(); return r1(v), err },
			must: func() []any { return r1(unset.MustKind()) }, want: tuplet.ErrUnset},

		{name: "IsCompressing", base: func() ([]any, error) { v, err := tr.IsCompressing(); return r1(v), err },
			must: func() []any { return r1(tr.MustIsCompressing()) }},
		{name: "IsCompressing/error", base: func() ([]any, error) { v, err := unset.IsCompressing(); return r1(v), err },
			must: func() []any { return r1(unset.MustIsCompressing()) }, want: tuplet.ErrUnset},

		{name: "IsExpanding", base: func() ([]any, error) { v, err := tr.IsExpanding(); return r1(v), err },
			must: func() []any { return r1(tr.MustIsExpanding()) }},
		{name: "IsExpanding/error", base: func() ([]any, error) { v, err := unset.IsExpanding(); return r1(v), err },
			must: func() []any { return r1(unset.MustIsExpanding()) }, want: tuplet.ErrUnset},

		{name: "IsNeutral", base: func() ([]any, error) { v, err := tr.IsNeutral(); return r1(v), err },
			must: func() []any { return r1(tr.MustIsNeutral()) }},
		{name: "IsNeutral/error", base: func() ([]any, error) { v, err := unset.IsNeutral(); return r1(v), err },
			must: func() []any { return r1(unset.MustIsNeutral()) }, want: tuplet.ErrUnset},

		{name: "NormalDivision", base: func() ([]any, error) { v, ok, err := tr.NormalDivision(); return r2(v, ok), err },
			must: func() []any { v, ok := tr.MustNormalDivision(); return r2(v, ok) }},
		{name: "NormalDivision/none", base: func() ([]any, error) {
			v, ok, err := tuplet.MustNew(5, 7).NormalDivision()
			return r2(v, ok), err
		}, must: func() []any { v, ok := tuplet.MustNew(5, 7).MustNormalDivision(); return r2(v, ok) }},
		{name: "NormalDivision/error", base: func() ([]any, error) { v, ok, err := unset.NormalDivision(); return r2(v, ok), err },
			must: func() []any { v, ok := unset.MustNormalDivision(); return r2(v, ok) }, want: tuplet.ErrUnset},

		{name: "HasDefaultNormal", base: func() ([]any, error) { v, err := tr.HasDefaultNormal(); return r1(v), err },
			must: func() []any { return r1(tr.MustHasDefaultNormal()) }},
		{name: "HasDefaultNormal/error", base: func() ([]any, error) { v, err := unset.HasDefaultNormal(); return r1(v), err },
			must: func() []any { return r1(unset.MustHasDefaultNormal()) }, want: tuplet.ErrUnset},

		{name: "HasDefaultNormalFor", base: func() ([]any, error) {
			v, err := tr.HasDefaultNormalFor(tuplet.Ternary)
			return r1(v), err
		}, must: func() []any { return r1(tr.MustHasDefaultNormalFor(tuplet.Ternary)) }},
		{name: "HasDefaultNormalFor/error", base: func() ([]any, error) {
			v, err := tr.HasDefaultNormalFor(tuplet.Division(7))
			return r1(v), err
		}, must: func() []any { return r1(tr.MustHasDefaultNormalFor(tuplet.Division(7))) }, want: tuplet.ErrNoDefault},

		{name: "Coefficient", base: func() ([]any, error) { v, err := tr.Coefficient(); return r1(v), err },
			must: func() []any { return r1(tr.MustCoefficient()) }},
		{name: "Coefficient/error", base: func() ([]any, error) { v, err := unset.Coefficient(); return r1(v), err },
			must: func() []any { return r1(unset.MustCoefficient()) }, want: tuplet.ErrUnset},

		{name: "ReciprocalCoefficient", base: func() ([]any, error) { v, err := tr.ReciprocalCoefficient(); return r1(v), err },
			must: func() []any { return r1(tr.MustReciprocalCoefficient()) }},
		{name: "ReciprocalCoefficient/error", base: func() ([]any, error) {
			v, err := unset.ReciprocalCoefficient()
			return r1(v), err
		}, must: func() []any { return r1(unset.MustReciprocalCoefficient()) }, want: tuplet.ErrUnset},

		{name: "ReducedCoefficient", base: func() ([]any, error) {
			v, err := tuplet.Sextuplet().ReducedCoefficient()
			return r1(v), err
		}, must: func() []any { return r1(tuplet.Sextuplet().MustReducedCoefficient()) }},
		{name: "ReducedCoefficient/error", base: func() ([]any, error) { v, err := unset.ReducedCoefficient(); return r1(v), err },
			must: func() []any { return r1(unset.MustReducedCoefficient()) }, want: tuplet.ErrUnset},

		{name: "EqualCoefficient", base: func() ([]any, error) {
			v, err := tr.EqualCoefficient(tuplet.Sextuplet())
			return r1(v), err
		}, must: func() []any { return r1(tr.MustEqualCoefficient(tuplet.Sextuplet())) }},
		{name: "EqualCoefficient/error", base: func() ([]any, error) { v, err := tr.EqualCoefficient(unset); return r1(v), err },
			must: func() []any { return r1(tr.MustEqualCoefficient(unset)) }, want: tuplet.ErrUnset},

		{name: "CompareCoefficient", base: func() ([]any, error) {
			v, err := tr.CompareCoefficient(tuplet.Duplet())
			return r1(v), err
		}, must: func() []any { return r1(tr.MustCompareCoefficient(tuplet.Duplet())) }},
		{name: "CompareCoefficient/error", base: func() ([]any, error) { v, err := unset.CompareCoefficient(tr); return r1(v), err },
			must: func() []any { return r1(unset.MustCompareCoefficient(tr)) }, want: tuplet.ErrUnset},

		{name: "Reduce", base: func() ([]any, error) { v, err := tuplet.Sextuplet().Reduce(); return r1(v), err },
			must: func() []any { return r1(tuplet.Sextuplet().MustReduce()) }},
		{name: "Reduce/error", base: func() ([]any, error) { v, err := unset.Reduce(); return r1(v), err },
			must: func() []any { return r1(unset.MustReduce()) }, want: tuplet.ErrUnset},

		{name: "Scale", base: func() ([]any, error) { v, err := tr.Scale(3); return r1(v), err },
			must: func() []any { return r1(tr.MustScale(3)) }},
		{name: "Scale/argument", base: func() ([]any, error) { v, err := tr.Scale(0); return r1(v), err },
			must: func() []any { return r1(tr.MustScale(0)) }, want: tuplet.ErrInvalidArgument},
		{name: "Scale/range", base: func() ([]any, error) { v, err := maxInt.Scale(2); return r1(v), err },
			must: func() []any { return r1(maxInt.MustScale(2)) }, want: tuplet.ErrOutOfRange},

		{name: "Apply", base: func() ([]any, error) { v, err := tr.Apply(d); return r1(v), err },
			must: func() []any { return r1(tr.MustApply(d)) }},
		{name: "Apply/argument", base: func() ([]any, error) { v, err := tr.Apply(fraction.Fraction{}); return r1(v), err },
			must: func() []any { return r1(tr.MustApply(fraction.Fraction{})) }, want: tuplet.ErrInvalidArgument},
		{name: "Apply/range", base: func() ([]any, error) { v, err := tuplet.Duplet().Apply(maxFrac); return r1(v), err },
			must: func() []any { return r1(tuplet.Duplet().MustApply(maxFrac)) }, want: tuplet.ErrOutOfRange, causes: outOfRange},

		{name: "Unapply", base: func() ([]any, error) { v, err := tr.Unapply(d); return r1(v), err },
			must: func() []any { return r1(tr.MustUnapply(d)) }},
		{name: "Unapply/range", base: func() ([]any, error) { v, err := tr.Unapply(maxFrac); return r1(v), err },
			must: func() []any { return r1(tr.MustUnapply(maxFrac)) }, want: tuplet.ErrOutOfRange, causes: outOfRange},

		{name: "Span", base: func() ([]any, error) { v, err := tr.Span(frac(1, 4)); return r1(v), err },
			must: func() []any { return r1(tr.MustSpan(frac(1, 4))) }},
		{name: "Span/range", base: func() ([]any, error) { v, err := tr.Span(maxFrac); return r1(v), err },
			must: func() []any { return r1(tr.MustSpan(maxFrac)) }, want: tuplet.ErrOutOfRange, causes: outOfRange},

		{name: "Nest", base: func() ([]any, error) { v, err := tr.Nest(tuplet.Duplet()); return r1(v), err },
			must: func() []any { return r1(tr.MustNest(tuplet.Duplet())) }},
		{name: "Nest/range", base: func() ([]any, error) { v, err := maxInt.Nest(maxInt); return r1(v), err },
			must: func() []any { return r1(maxInt.MustNest(maxInt)) }, want: tuplet.ErrOutOfRange},

		{name: "Flatten", base: func() ([]any, error) { v, err := tuplet.Flatten(tr, tr); return r1(v), err },
			must: func() []any { return r1(tuplet.MustFlatten(tr, tr)) }},
		{name: "Flatten/empty", base: func() ([]any, error) { v, err := tuplet.Flatten(); return r1(v), err },
			must: func() []any { return r1(tuplet.MustFlatten()) }},
		{name: "Flatten/range", base: func() ([]any, error) { v, err := tuplet.Flatten(maxInt, maxInt); return r1(v), err },
			must: func() []any { return r1(tuplet.MustFlatten(maxInt, maxInt)) }, want: tuplet.ErrOutOfRange},

		{name: "ChainCoefficient", base: func() ([]any, error) {
			v, err := tuplet.ChainCoefficient(tuplet.Quintuplet(), tr)
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustChainCoefficient(tuplet.Quintuplet(), tr)) }},
		{name: "ChainCoefficient/range", base: func() ([]any, error) { v, err := tuplet.ChainCoefficient(five...); return r1(v), err },
			must: func() []any { return r1(tuplet.MustChainCoefficient(five...)) }, want: tuplet.ErrOutOfRange, causes: outOfRange},

		{name: "ApplyChain", base: func() ([]any, error) { v, err := tuplet.ApplyChain(d, tr, tr); return r1(v), err },
			must: func() []any { return r1(tuplet.MustApplyChain(d, tr, tr)) }},
		{name: "ApplyChain/argument", base: func() ([]any, error) {
			v, err := tuplet.ApplyChain(fraction.Fraction{}, tr)
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustApplyChain(fraction.Fraction{}, tr)) }, want: tuplet.ErrInvalidArgument},
		{name: "ApplyChain/range", base: func() ([]any, error) { v, err := tuplet.ApplyChain(d, five...); return r1(v), err },
			must: func() []any { return r1(tuplet.MustApplyChain(d, five...)) }, want: tuplet.ErrOutOfRange, causes: outOfRange},
		{name: "ApplyChain/result range", base: func() ([]any, error) {
			v, err := tuplet.ApplyChain(maxFrac, tuplet.Duplet())
			return r1(v), err
		}, must: func() []any { return r1(tuplet.MustApplyChain(maxFrac, tuplet.Duplet())) }, want: tuplet.ErrOutOfRange, causes: outOfRange},

		{name: "Name", base: func() ([]any, error) { v, ok, err := tr.Name(); return r2(v, ok), err },
			must: func() []any { v, ok := tr.MustName(); return r2(v, ok) }},
		{name: "Name/none", base: func() ([]any, error) { v, ok, err := tuplet.MustNew(11, 8).Name(); return r2(v, ok), err },
			must: func() []any { v, ok := tuplet.MustNew(11, 8).MustName(); return r2(v, ok) }},
		{name: "Name/error", base: func() ([]any, error) { v, ok, err := unset.Name(); return r2(v, ok), err },
			must: func() []any { v, ok := unset.MustName(); return r2(v, ok) }, want: tuplet.ErrUnset},

		{name: "RussianName", base: func() ([]any, error) { v, ok, err := tr.RussianName(); return r2(v, ok), err },
			must: func() []any { v, ok := tr.MustRussianName(); return r2(v, ok) }},
		{name: "RussianName/error", base: func() ([]any, error) { v, ok, err := unset.RussianName(); return r2(v, ok), err },
			must: func() []any { v, ok := unset.MustRussianName(); return r2(v, ok) }, want: tuplet.ErrUnset},

		{name: "ShortString", base: func() ([]any, error) { v, err := tr.ShortString(); return r1(v), err },
			must: func() []any { return r1(tr.MustShortString()) }},
		{name: "ShortString/error", base: func() ([]any, error) { v, err := unset.ShortString(); return r1(v), err },
			must: func() []any { return r1(unset.MustShortString()) }, want: tuplet.ErrUnset},

		{name: "Parse", base: func() ([]any, error) { v, err := tuplet.Parse("6:4"); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParse("6:4")) }},
		{name: "Parse/text", base: func() ([]any, error) { v, err := tuplet.Parse("6/4"); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParse("6/4")) }, want: tuplet.ErrInvalidText},
		{name: "Parse/range", base: func() ([]any, error) { v, err := tuplet.Parse("1:9223372036854775808"); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParse("1:9223372036854775808")) }, want: tuplet.ErrOutOfRange},

		{name: "ParseShort", base: func() ([]any, error) { v, err := tuplet.ParseShort("5", tuplet.Binary); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParseShort("5", tuplet.Binary)) }},
		{name: "ParseShort/no default", base: func() ([]any, error) { v, err := tuplet.ParseShort("3", tuplet.Ternary); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParseShort("3", tuplet.Ternary)) }, want: tuplet.ErrNoDefault},

		{name: "ParseName", base: func() ([]any, error) { v, err := tuplet.ParseName("септоль"); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParseName("септоль")) }},
		{name: "ParseName/text", base: func() ([]any, error) { v, err := tuplet.ParseName("Septuplet"); return r1(v), err },
			must: func() []any { return r1(tuplet.MustParseName("Septuplet")) }, want: tuplet.ErrInvalidText},
	}
}

// TestMust covers K-14 and dev.md "Must": on success a Must variant returns
// the same results as its operation; on error it panics with an error value
// of the same sentinel and the same causes, with the same text.
func TestMust(t *testing.T) {
	t.Parallel()

	for _, tc := range mustCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			results, err := tc.base()

			if tc.want == nil {
				if !assert.NoError(t, err) {
					return
				}

				var got []any

				if assertNoPanic(t, func() { got = tc.must() }) {
					assert.Equal(t, results, got)
				}

				return
			}

			assertErr(t, err, tc.want, tc.causes...)

			panicErr := assertPanicErr(t, func() { tc.must() }, tc.want, tc.causes...)
			if err != nil && panicErr != nil {
				assert.Equal(t, err.Error(), panicErr.Error(), "the panic value is the error of the operation")
			}
		})
	}
}
