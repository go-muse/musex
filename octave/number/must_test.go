package number_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Must variants panic on exactly the inputs on which the variant with an
// error fails. The panic value is an error that matches the same sentinel
// with errors.Is; its text is not compared.
func TestMust_PanicsWithTheSameError(t *testing.T) {
	t.Parallel()

	var zero number.Numbering

	cases := []struct {
		name string
		call func() (int, error)
		must func()
		want error
	}{
		{
			name: "MustNew(4, 0, 0)",
			call: func() (int, error) { _, err := number.New(4, 0, 0); return 0, err },
			must: func() { number.MustNew(4, 0, 0) },
			want: number.ErrInvalidDefinition,
		},
		{
			name: "MustNew(4, 3, 5, 3)",
			call: func() (int, error) { _, err := number.New(4, 3, 5, 3); return 0, err },
			must: func() { number.MustNew(4, 3, 5, 3) },
			want: number.ErrInvalidDefinition,
		},
		{
			name: "MustNew(4, 3, 0, 0)",
			call: func() (int, error) { _, err := number.New(4, 3, 0, 0); return 0, err },
			must: func() { number.MustNew(4, 3, 0, 0) },
			want: number.ErrInvalidDefinition,
		},
		{
			name: "MustNew(MaxInt, MaxInt, MaxInt)",
			call: func() (int, error) { _, err := number.New(maxInt, maxInt, maxInt); return 0, err },
			must: func() { number.MustNew(maxInt, maxInt, maxInt) },
			want: number.ErrInvalidDefinition,
		},
		{
			name: "Franco-Belgian MustToASPN(0)",
			call: func() (int, error) { return number.FrancoBelgian().ToASPN(0) },
			must: func() { number.FrancoBelgian().MustToASPN(0) },
			want: number.ErrSkipped,
		},
		{
			name: "New(4, 4, 0, 1) MustToASPN(1)",
			call: func() (int, error) { return number.MustNew(4, 4, 0, 1).ToASPN(1) },
			must: func() { number.MustNew(4, 4, 0, 1).MustToASPN(1) },
			want: number.ErrSkipped,
		},
		{
			name: "MustConvert(0, Franco-Belgian, ASPN)",
			call: func() (int, error) { return number.Convert(0, number.FrancoBelgian(), number.ASPN()) },
			must: func() { number.MustConvert(0, number.FrancoBelgian(), number.ASPN()) },
			want: number.ErrSkipped,
		},
		{
			name: "Middle C = C5 MustFromASPN(MaxInt)",
			call: func() (int, error) { return number.MiddleC5().FromASPN(maxInt) },
			must: func() { number.MiddleC5().MustFromASPN(maxInt) },
			want: number.ErrOutOfRange,
		},
		{
			name: "Middle C = C3 MustFromASPN(MinInt)",
			call: func() (int, error) { return number.MiddleC3().FromASPN(minInt) },
			must: func() { number.MiddleC3().MustFromASPN(minInt) },
			want: number.ErrOutOfRange,
		},
		{
			name: "Middle C = C3 MustToASPN(MaxInt)",
			call: func() (int, error) { return number.MiddleC3().ToASPN(maxInt) },
			must: func() { number.MiddleC3().MustToASPN(maxInt) },
			want: number.ErrOutOfRange,
		},
		{
			name: "MustConvert(MaxInt, Middle C = C3, Middle C = C5)",
			call: func() (int, error) { return number.Convert(maxInt, number.MiddleC3(), number.MiddleC5()) },
			must: func() { number.MustConvert(maxInt, number.MiddleC3(), number.MiddleC5()) },
			want: number.ErrOutOfRange,
		},
		{
			name: "New(0, MaxInt) MustFromASPN(1)",
			call: func() (int, error) { return number.MustNew(0, maxInt).FromASPN(1) },
			must: func() { number.MustNew(0, maxInt).MustFromASPN(1) },
			want: number.ErrOutOfRange,
		},
		{
			name: "zero MustFromASPN(4)",
			call: func() (int, error) { return zero.FromASPN(4) },
			must: func() { zero.MustFromASPN(4) },
			want: number.ErrUndefined,
		},
		{
			name: "zero MustToASPN(4)",
			call: func() (int, error) { return zero.ToASPN(4) },
			must: func() { zero.MustToASPN(4) },
			want: number.ErrUndefined,
		},
		{
			name: "MustConvert(4, zero, ASPN)",
			call: func() (int, error) { return number.Convert(4, zero, number.ASPN()) },
			must: func() { number.MustConvert(4, zero, number.ASPN()) },
			want: number.ErrUndefined,
		},
		{
			name: "MustConvert(4, ASPN, zero)",
			call: func() (int, error) { return number.Convert(4, number.ASPN(), zero) },
			must: func() { number.MustConvert(4, number.ASPN(), zero) },
			want: number.ErrUndefined,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.call()
			assertOnlyError(t, err, tc.want, "variant with an error")
			assertPanicsWithErrorIs(t, tc.want, tc.must, tc.name)
		})
	}
}

// Must variants do not panic on valid inputs and return the same result as
// the variant with an error (K-3...K-9 inputs).
func TestMust_ValidInputs(t *testing.T) {
	t.Parallel()

	set := criteriaSet(t)

	for _, n := range set {
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			for a := -20; a <= 20; a++ {
				desc := fmt.Sprintf("%s, ASPN %d", n.name, a)

				x, err := n.n.FromASPN(a)
				require.NoError(t, err, desc)

				var mustX, mustA int
				require.NotPanics(t, func() { mustX = n.n.MustFromASPN(a) }, desc)
				assert.Equal(t, x, mustX, "MustFromASPN: "+desc)

				back, err := n.n.ToASPN(x)
				require.NoError(t, err, desc)
				require.NotPanics(t, func() { mustA = n.n.MustToASPN(x) }, desc)
				assert.Equal(t, back, mustA, "MustToASPN: "+desc)

				for _, to := range set {
					y, err := number.Convert(x, n.n, to.n)
					require.NoError(t, err, desc)

					var mustY int
					require.NotPanics(t, func() { mustY = number.MustConvert(x, n.n, to.n) }, desc)
					assert.Equal(t, y, mustY, "MustConvert to %s: %s", to.name, desc)
				}
			}
		})
	}
}

// MustNew on a valid definition returns a numbering equal to the result of
// New, with the same String and no label.
func TestMustNew_Valid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		aspn, num int
		skipped   []int
	}{
		{aspn: 4, num: 4},
		{aspn: 4, num: 5},
		{aspn: -1, num: 0},
		{aspn: 4, num: 4, skipped: []int{0, 1}},
		{aspn: 4, num: 3, skipped: []int{0}},
		{aspn: minInt, num: maxInt, skipped: []int{minInt, maxInt - 1}},
	}

	for _, tc := range cases {
		name := ownName(tc.aspn, tc.num, tc.skipped)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			n, err := number.New(tc.aspn, tc.num, tc.skipped...)
			require.NoError(t, err)

			var m number.Numbering
			require.NotPanics(t, func() { m = number.MustNew(tc.aspn, tc.num, tc.skipped...) })
			assert.True(t, m.IsDefined())
			assert.True(t, m.Equal(n))
			assert.True(t, n.Equal(m))
			assert.Equal(t, n.String(), m.String())
			assert.Empty(t, m.Label())
		})
	}
}

// Must variants allow chains of calls on literals known to be valid.
func TestMust_Chain(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 2, number.MustNew(4, 3, 0).MustToASPN(1))
	assert.Equal(t, -1, number.MustNew(4, 3, 0).MustFromASPN(1))
	assert.Equal(t, 3, number.FrancoBelgian().MustFromASPN(4))
	assert.Equal(t, 3, number.MustConvert(3, number.MiddleC3(), number.FrancoBelgian()))
}
