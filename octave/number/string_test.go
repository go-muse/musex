package number_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Architecture KR-O10: the fixed forms of String. Built-in numberings print
// their labels; own numberings print the arguments of New with the skipped
// numbers in ascending order; the zero value prints Numbering(undefined).
func TestNumbering_String(t *testing.T) {
	t.Parallel()

	newNumbering := func(aspn, num int, skipped ...int) func() (number.Numbering, error) {
		return func() (number.Numbering, error) { return number.New(aspn, num, skipped...) }
	}
	builtin := func(get func() number.Numbering) func() (number.Numbering, error) {
		return func() (number.Numbering, error) { return get(), nil }
	}

	cases := []struct {
		name string
		n    func() (number.Numbering, error)
		want string
	}{
		{name: "ASPN", n: builtin(number.ASPN), want: "ASPN"},
		{name: "Franco-Belgian", n: builtin(number.FrancoBelgian), want: "Franco-Belgian"},
		{name: "Mexican", n: builtin(number.Mexican), want: "Mexican"},
		{name: "Middle C = C3", n: builtin(number.MiddleC3), want: "Middle C = C3"},
		{name: "Middle C = C5", n: builtin(number.MiddleC5), want: "Middle C = C5"},
		{name: "New(4, 5)", n: newNumbering(4, 5), want: "ASPN 4 -> 5"},
		{name: "New(-1, 0)", n: newNumbering(-1, 0), want: "ASPN -1 -> 0"},
		{name: "New(4, 3)", n: newNumbering(4, 3), want: "ASPN 4 -> 3"},
		{name: "New(4, 3, []int{}...)", n: newNumbering(4, 3, []int{}...), want: "ASPN 4 -> 3"},
		{name: "New(4, 3, nil...)", n: newNumbering(4, 3, nil...), want: "ASPN 4 -> 3"},
		{name: "New(4, 3, 0)", n: newNumbering(4, 3, 0), want: "ASPN 4 -> 3, skip [0]"},
		{name: "New(4, 4, 1, 0)", n: newNumbering(4, 4, 1, 0), want: "ASPN 4 -> 4, skip [0 1]"},
		{name: "New(4, 4, -10)", n: newNumbering(4, 4, -10), want: "ASPN 4 -> 4, skip [-10]"},
		{name: "New(0, 0, 5, -3, 2)", n: newNumbering(0, 0, 5, -3, 2), want: "ASPN 0 -> 0, skip [-3 2 5]"},
		{name: "New(0, 0, MinInt)", n: newNumbering(0, 0, minInt), want: "ASPN 0 -> 0, skip [-9223372036854775808]"},
		{
			name: "New(MinInt, MaxInt)",
			n:    newNumbering(minInt, maxInt),
			want: "ASPN -9223372036854775808 -> 9223372036854775807",
		},
		{
			name: "zero value",
			n:    func() (number.Numbering, error) { return number.Numbering{}, nil },
			want: "Numbering(undefined)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n, err := tc.n()
			require.NoError(t, err)

			assert.Equal(t, tc.want, n.String())
			assert.Equal(t, tc.want, fmt.Sprint(n), "fmt uses String")
			assert.Equal(t, tc.want, fmt.Sprintf("%v", n), "fmt %v uses String")
		})
	}
}

// String is not equality: equal numberings may print differently.
func TestNumbering_StringIsNotEquality(t *testing.T) {
	t.Parallel()

	mine := number.MustNew(4, 3, 0)
	fb := number.FrancoBelgian()

	assert.True(t, mine.Equal(fb))
	assert.Equal(t, "ASPN 4 -> 3, skip [0]", mine.String())
	assert.Equal(t, "Franco-Belgian", fb.String())
}
