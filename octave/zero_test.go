package octave_test

import (
	"fmt"
	"iter"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-18, K-24, E-5: the zero value is the only invalid value; it is the same
// whether declared or written as a literal.
func TestZeroValueIsInvalid(t *testing.T) {
	t.Parallel()

	var z octave.Octave
	assert.False(t, z.IsValid())
	assert.True(t, z == octave.Octave{}, "var z == octave.Octave{}")
	for _, a := range []int{0, 4, minInt, maxInt} {
		assert.True(t, z != oct(a), "Z != O(%d)", a)
	}
}

// K-24, E-5: operations with a number or an octave as a result give
// ErrInvalid for the invalid value as the receiver or an argument, with the
// value 0 or octave.Octave{}; no octave such as ASPN 0 or ASPN 4 is
// substituted. The Must variants panic with ErrInvalid.
func TestInvalidValueOperations(t *testing.T) {
	t.Parallel()

	var z octave.Octave

	intCases := []struct {
		name string
		f    func() (int, error)
		must func()
	}{
		{
			name: "Z.Number(ASPN)",
			f:    func() (int, error) { return z.Number(number.ASPN()) },
			must: func() { z.MustNumber(number.ASPN()) },
		},
		{
			name: "Z.Number(FB)",
			f:    func() (int, error) { return z.Number(fb()) },
			must: func() { z.MustNumber(fb()) },
		},
		{
			name: "Z.Number(undefined)",
			f:    func() (int, error) { return z.Number(number.Numbering{}) },
			must: func() { z.MustNumber(number.Numbering{}) },
		},
		{
			name: "Z.ASPN()",
			f:    func() (int, error) { return z.ASPN() },
			must: func() { z.MustASPN() },
		},
		{
			name: "Z.Sub(O(4))",
			f:    func() (int, error) { return z.Sub(oct(4)) },
			must: func() { z.MustSub(oct(4)) },
		},
		{
			name: "O(4).Sub(Z)",
			f:    func() (int, error) { return oct(4).Sub(z) },
			must: func() { oct(4).MustSub(z) },
		},
		{
			name: "Z.Sub(Z)",
			f:    func() (int, error) { return z.Sub(z) },
			must: func() { z.MustSub(z) },
		},
		{
			name: "O(MIN).Sub(Z)",
			f:    func() (int, error) { return oct(minInt).Sub(z) },
			must: func() { oct(minInt).MustSub(z) },
		},
		{
			name: "Z.Distance(O(4))",
			f:    func() (int, error) { return z.Distance(oct(4)) },
			must: func() { z.MustDistance(oct(4)) },
		},
		{
			name: "O(4).Distance(Z)",
			f:    func() (int, error) { return oct(4).Distance(z) },
			must: func() { oct(4).MustDistance(z) },
		},
		{
			name: "Z.Distance(Z)",
			f:    func() (int, error) { return z.Distance(z) },
			must: func() { z.MustDistance(z) },
		},
		{
			name: "O(MAX).Distance(Z)",
			f:    func() (int, error) { return oct(maxInt).Distance(z) },
			must: func() { oct(maxInt).MustDistance(z) },
		},
	}

	for _, tc := range intCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got int
				err error
			)
			require.NotPanics(t, func() { got, err = tc.f() })
			assertCause(t, err, causeInvalid, tc.name)
			assert.Zero(t, got, "value on error, not a substituted number")
			assertPanicsCause(t, causeInvalid, tc.must, "Must variant of "+tc.name)
		})
	}

	type octaveCase struct {
		name string
		f    func() (octave.Octave, error)
		must func()
	}

	octaveCases := []octaveCase{
		{name: "Z.Up()", f: z.Up, must: func() { z.MustUp() }},
		{name: "Z.Down()", f: z.Down, must: func() { z.MustDown() }},
	}
	for _, k := range []int{0, 1, -1, minInt, maxInt} {
		octaveCases = append(octaveCases,
			octaveCase{
				name: fmt.Sprintf("Z.Shift(%d)", k),
				f:    func() (octave.Octave, error) { return z.Shift(k) },
				must: func() { z.MustShift(k) },
			},
			octaveCase{
				name: fmt.Sprintf("Z.ShiftDown(%d)", k),
				f:    func() (octave.Octave, error) { return z.ShiftDown(k) },
				must: func() { z.MustShiftDown(k) },
			},
		)
	}

	for _, tc := range octaveCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = tc.f() })
			assertCause(t, err, causeInvalid, tc.name)
			assert.Equal(t, octave.Octave{}, got, "value on error")
			assert.False(t, got.IsValid(), "value on error")
			assert.False(t, got.Equal(oct(1)) || got.Equal(oct(5)) || got.Equal(oct(0)), "no substituted octave")
			assertPanicsCause(t, causeInvalid, tc.must, "Must variant of "+tc.name)
		})
	}
}

// K-24, E-5, D-5 of the model: a range with an invalid end gives ErrInvalid
// and a nil sequence, not an empty one; MustRange panics.
func TestInvalidValueRange(t *testing.T) {
	t.Parallel()

	var z octave.Octave

	cases := []struct {
		name     string
		from, to octave.Octave
	}{
		{name: "Z...O(4)", from: z, to: oct(4)},
		{name: "O(4)...Z", from: oct(4), to: z},
		{name: "Z...Z", from: z, to: z},
		{name: "O(MIN)...Z", from: oct(minInt), to: z},
		{name: "Z...O(MAX)", from: z, to: oct(maxInt)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			// A non-nil starting value shows that Range really returns nil.
			seq := iter.Seq[octave.Octave](func(func(octave.Octave) bool) {})
			require.NotPanics(t, func() { seq, err = octave.Range(tc.from, tc.to) })
			assertCause(t, err, causeInvalid, "Range")
			assert.Nil(t, seq, "value on error")
			assertPanicsCause(t, causeInvalid, func() { octave.MustRange(tc.from, tc.to) }, "MustRange")
		})
	}
}

// K-24, E-5: own numberings and catalogs with the invalid value give
// ErrInvalid, checked before the definition and the names; Must variants
// panic.
func TestInvalidValueDefinitions(t *testing.T) {
	t.Parallel()

	var z octave.Octave

	numberingCases := []struct {
		name    string
		num     int
		skipped []int
	}{
		{name: "Z is 3", num: 3},
		{name: "Z is 3, skip 0", num: 3, skipped: []int{0}},
		{name: "Z is 0, skip 0", num: 0, skipped: []int{0}},
		{name: "Z is 3, skip 0 0", num: 3, skipped: []int{0, 0}},
	}

	for _, tc := range numberingCases {
		t.Run("NewNumbering "+tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got number.Numbering
				err error
			)
			require.NotPanics(t, func() { got, err = octave.NewNumbering(z, tc.num, tc.skipped...) })
			assertCause(t, err, causeInvalid, "NewNumbering")
			assert.False(t, got.IsDefined(), "value on error is the zero Numbering")
			assertPanicsCause(t, causeInvalid, func() { octave.MustNewNumbering(z, tc.num, tc.skipped...) }, "MustNewNumbering")
		})
	}

	catalogCases := []struct {
		name  string
		pairs map[octave.Octave]string
	}{
		{name: "Z empty name", pairs: map[octave.Octave]string{z: ""}},
		{name: "Z named", pairs: map[octave.Octave]string{z: "a"}},
		{name: "Z and O(1) with one name", pairs: map[octave.Octave]string{z: "a", oct(1): "a"}},
		{name: "Z and O(1) with names", pairs: map[octave.Octave]string{z: "x", oct(1): "y"}},
		{name: "Z empty and duplicate", pairs: map[octave.Octave]string{z: "", oct(1): "a", oct(2): "a"}},
	}

	for _, tc := range catalogCases {
		t.Run("NewCatalog "+tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got *naming.Catalog
				err error
			)
			require.NotPanics(t, func() { got, err = octave.NewCatalog(tc.pairs) })
			assertCause(t, err, causeInvalid, "NewCatalog")
			assert.Nil(t, got, "value on error")
			assertPanicsCause(t, causeInvalid, func() { octave.MustNewCatalog(tc.pairs) }, "MustNewCatalog")
		})
	}
}

// K-24, D-3 of the model: the name of the invalid value is "" in every
// catalog, not the name of ASPN 0 or ASPN 4, without an error or a panic.
func TestInvalidValueName(t *testing.T) {
	t.Parallel()

	var z octave.Octave
	catalogs := append(builtinCatalogs(), emptyCatalogs()...)
	catalogs = append(catalogs, catalogCase{
		name: "own with names of ASPN 0 and 4",
		c:    naming.MustNew(map[int]string{0: "zero", 4: "four"}),
	})

	for _, cc := range catalogs {
		var got string
		require.NotPanics(t, func() { got = z.Name(cc.c) }, "Z.Name(%s)", cc.name)
		assert.Empty(t, got, "Z.Name(%s)", cc.name)
	}
}

// K-24, K-31: operations without errors never panic on the invalid value.
func TestInvalidValueNeverPanics(t *testing.T) {
	t.Parallel()

	var z octave.Octave
	others := []octave.Octave{z, oct(minInt), oct(0), oct(4), oct(maxInt)}

	assert.NotPanics(t, func() { _ = z.IsValid() })
	assert.NotPanics(t, func() { _ = z.String() })
	assert.NotPanics(t, func() { _ = z.GoString() })
	assert.NotPanics(t, func() { _ = z.Name(naming.Russian()) })
	assert.NotPanics(t, func() { _ = z.Name(nil) })
	for _, o := range others {
		assert.NotPanics(t, func() { _ = z.Equal(o) }, "Z.Equal(%s)", describe([]octave.Octave{o}))
		assert.NotPanics(t, func() { _ = o.Equal(z) }, "Equal(Z)")
		assert.NotPanics(t, func() { _ = z.Compare(o) }, "Z.Compare")
		assert.NotPanics(t, func() { _ = o.Compare(z) }, "Compare(Z)")
		assert.NotPanics(t, func() { _ = z.Above(o) }, "Z.Above")
		assert.NotPanics(t, func() { _ = z.Below(o) }, "Z.Below")
		assert.NotPanics(t, func() { _ = o.Between(z, o) }, "Between")
		assert.NotPanics(t, func() { _, _ = z.Sub(o) }, "Z.Sub")
		assert.NotPanics(t, func() { _, _ = o.Distance(z) }, "Distance(Z)")
		assert.NotPanics(t, func() { _, _ = octave.Range(o, z) }, "Range(o, Z)")
	}
}
