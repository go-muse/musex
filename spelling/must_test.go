package spelling_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contractErrors are all the errors named by the contract of the package.
var contractErrors = []struct {
	name string
	err  error
}{
	{"letter.ErrInvalid", letter.ErrInvalid},
	{"alteration.ErrOutOfRange", alteration.ErrOutOfRange},
	{"alteration.ErrZeroDenominator", alteration.ErrZeroDenominator},
	{"spelling.ErrInvalidForm", spelling.ErrInvalidForm},
}

// K-16: each Must variant panics exactly when its operation returns an error,
// with an error value that matches the same errors of the contract.
func TestMustPanics(t *testing.T) {
	t.Parallel()

	var unset spelling.Spelling

	top := spelling.MustNew(letter.C, whole(maxMagnitude))
	bottom := spelling.MustNew(letter.C, whole(-maxMagnitude))
	half := frac(1, 2)

	cases := []struct {
		name   string
		target error
		op     func() error
		must   func()
	}{
		{
			"New with the zero letter", letter.ErrInvalid,
			func() error { _, err := spelling.New(letter.Letter(0), alteration.Sharp()); return err },
			func() { _ = spelling.MustNew(letter.Letter(0), alteration.Sharp()) },
		},
		{
			"New with letter 8", letter.ErrInvalid,
			func() error { _, err := spelling.New(letter.Letter(8), alteration.Natural()); return err },
			func() { _ = spelling.MustNew(letter.Letter(8), alteration.Natural()) },
		},
		{
			"Raise of (C, M)", alteration.ErrOutOfRange,
			func() error { _, err := top.Raise(); return err },
			func() { _ = top.MustRaise() },
		},
		{
			"Lower of (C, -M)", alteration.ErrOutOfRange,
			func() error { _, err := bottom.Lower(); return err },
			func() { _ = bottom.MustLower() },
		},
		{
			"RaiseBy(2) of (C, M)", alteration.ErrOutOfRange,
			func() error { _, err := top.RaiseBy(2); return err },
			func() { _ = top.MustRaiseBy(2) },
		},
		{
			"LowerBy(2) of (C, -M)", alteration.ErrOutOfRange,
			func() error { _, err := bottom.LowerBy(2); return err },
			func() { _ = bottom.MustLowerBy(2) },
		},
		{
			"RaiseByAlteration(1/2) of (C, M)", alteration.ErrOutOfRange,
			func() error { _, err := top.RaiseByAlteration(half); return err },
			func() { _ = top.MustRaiseByAlteration(half) },
		},
		{
			"LowerByAlteration(1/2) of (C, -M)", alteration.ErrOutOfRange,
			func() error { _, err := bottom.LowerByAlteration(half); return err },
			func() { _ = bottom.MustLowerByAlteration(half) },
		},
		{
			"AlterationDifference of (C, M) and (C, -M)", alteration.ErrOutOfRange,
			func() error { _, err := top.AlterationDifference(bottom); return err },
			func() { _ = top.MustAlterationDifference(bottom) },
		},
		{
			"Raise of the unset name", letter.ErrInvalid,
			func() error { _, err := unset.Raise(); return err },
			func() { _ = unset.MustRaise() },
		},
		{
			"Lower of the unset name", letter.ErrInvalid,
			func() error { _, err := unset.Lower(); return err },
			func() { _ = unset.MustLower() },
		},
		{
			"RaiseBy(0) of the unset name", letter.ErrInvalid,
			func() error { _, err := unset.RaiseBy(0); return err },
			func() { _ = unset.MustRaiseBy(0) },
		},
		{
			"LowerBy(0) of the unset name", letter.ErrInvalid,
			func() error { _, err := unset.LowerBy(0); return err },
			func() { _ = unset.MustLowerBy(0) },
		},
		{
			"RaiseByAlteration(natural) of the unset name", letter.ErrInvalid,
			func() error { _, err := unset.RaiseByAlteration(alteration.Natural()); return err },
			func() { _ = unset.MustRaiseByAlteration(alteration.Natural()) },
		},
		{
			"LowerByAlteration(natural) of the unset name", letter.ErrInvalid,
			func() error { _, err := unset.LowerByAlteration(alteration.Natural()); return err },
			func() { _ = unset.MustLowerByAlteration(alteration.Natural()) },
		},
		{
			"AlterationDifference of the unset name and C", letter.ErrInvalid,
			func() error { _, err := unset.AlterationDifference(spelling.C()); return err },
			func() { _ = unset.MustAlterationDifference(spelling.C()) },
		},
		{
			"AlterationDifference of C and the unset name", letter.ErrInvalid,
			func() error { _, err := spelling.C().AlterationDifference(unset); return err },
			func() { _ = spelling.C().MustAlterationDifference(unset) },
		},
		{
			"Parse of the empty string", spelling.ErrInvalidForm,
			func() error { _, err := spelling.Parse(""); return err },
			func() { _ = spelling.MustParse("") },
		},
		{
			"Parse of C#", spelling.ErrInvalidForm,
			func() error { _, err := spelling.Parse("C#"); return err },
			func() { _ = spelling.MustParse("C#") },
		},
		{
			"Parse of C+1/0", spelling.ErrInvalidForm,
			func() error { _, err := spelling.Parse("C+1/0"); return err },
			func() { _ = spelling.MustParse("C+1/0") },
		},
		{
			"Parse of C+16777217", spelling.ErrInvalidForm,
			func() error { _, err := spelling.Parse("C+16777217"); return err },
			func() { _ = spelling.MustParse("C+16777217") },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opErr := tc.op()
			require.ErrorIs(t, opErr, tc.target)

			mustErr := panicError(t, tc.must)
			require.ErrorIs(t, mustErr, tc.target)

			for _, e := range contractErrors {
				assert.Equalf(t, errors.Is(opErr, e.err), errors.Is(mustErr, e.err),
					"the panic value and the returned error must agree on %s", e.name)
			}
		})
	}
}

// K-16: on input without an error, each Must variant returns the same result
// as its operation and does not panic.
func TestMustMatchesOperation(t *testing.T) {
	t.Parallel()

	t.Run("MustNew", func(t *testing.T) {
		t.Parallel()

		for _, l := range letter.All() {
			for _, a := range setS() {
				want, err := spelling.New(l, a.a)
				require.NoError(t, err)

				require.NotPanics(t, func() { assert.Equal(t, want, spelling.MustNew(l, a.a)) })
			}
		}
	})

	for _, tc := range setN() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			x := tc.build()

			same := func(name string, op func() (spelling.Spelling, error), must func() spelling.Spelling) {
				want, err := op()
				require.NoError(t, err, name)
				require.NotPanics(t, func() { assert.Equal(t, want, must(), name) }, name)
			}

			same("Raise", x.Raise, x.MustRaise)
			same("Lower", x.Lower, x.MustLower)

			for n := -5; n <= 5; n++ {
				same("RaiseBy", func() (spelling.Spelling, error) { return x.RaiseBy(n) }, func() spelling.Spelling { return x.MustRaiseBy(n) })
				same("LowerBy", func() (spelling.Spelling, error) { return x.LowerBy(n) }, func() spelling.Spelling { return x.MustLowerBy(n) })
			}

			for _, d := range setS() {
				same("RaiseByAlteration", func() (spelling.Spelling, error) { return x.RaiseByAlteration(d.a) },
					func() spelling.Spelling { return x.MustRaiseByAlteration(d.a) })
				same("LowerByAlteration", func() (spelling.Spelling, error) { return x.LowerByAlteration(d.a) },
					func() spelling.Spelling { return x.MustLowerByAlteration(d.a) })
			}

			for _, y := range setN() {
				ys := y.build()
				want, err := x.AlterationDifference(ys)
				require.NoError(t, err)
				require.NotPanics(t, func() { assertAlteration(t, want, x.MustAlterationDifference(ys)) })
			}

			form := tc.name
			want, err := spelling.Parse(form)
			require.NoError(t, err)
			require.NotPanics(t, func() { assert.Equal(t, want, spelling.MustParse(form)) })
		})
	}
}
