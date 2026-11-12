package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReduce covers KP-12 and O-10: division by the greatest common divisor.
func TestReduce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want [2]int
	}{
		{in: [2]int{6, 4}, want: [2]int{3, 2}},
		{in: [2]int{3, 2}, want: [2]int{3, 2}},
		{in: [2]int{9, 6}, want: [2]int{3, 2}},
		{in: [2]int{6, 6}, want: [2]int{1, 1}},
		{in: [2]int{1, 1}, want: [2]int{1, 1}},
		{in: [2]int{14, 8}, want: [2]int{7, 4}},
		{in: [2]int{12, 1}, want: [2]int{12, 1}},
		{in: [2]int{math.MaxInt, math.MaxInt}, want: [2]int{1, 1}},
		{in: [2]int{math.MaxInt, math.MaxInt - 1}, want: [2]int{math.MaxInt, math.MaxInt - 1}},
		{in: [2]int{1 << 62, 3 << 61}, want: [2]int{2, 3}},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.in[0])+":"+itoa(tt.in[1]), func(t *testing.T) {
			t.Parallel()

			tp := tuplet.MustNew(tt.in[0], tt.in[1])

			got, err := tp.Reduce()
			require.NoError(t, err)
			assertTuplet(t, got, tt.want[0], tt.want[1])
			assert.Equal(t, got, tp.MustReduce())

			equal, err := got.EqualCoefficient(tp)
			require.NoError(t, err)
			assert.True(t, equal, "Reduce keeps the coefficient value")
		})
	}
}

// TestScale covers KP-12, O-10 and K-7.
func TestScale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   [2]int
		k    int
		want [2]int
	}{
		{in: [2]int{3, 2}, k: 2, want: [2]int{6, 4}},
		{in: [2]int{3, 2}, k: 1, want: [2]int{3, 2}},
		{in: [2]int{2, 3}, k: 4, want: [2]int{8, 12}},
		{in: [2]int{1, 1}, k: math.MaxInt, want: [2]int{math.MaxInt, math.MaxInt}},
		{in: [2]int{math.MaxInt, 1}, k: 1, want: [2]int{math.MaxInt, 1}},
		{in: [2]int{2, 1}, k: 1 << 61, want: [2]int{1 << 62, 1 << 61}},
		{in: [2]int{3, 1}, k: 3 << 59, want: [2]int{9 << 59, 3 << 59}},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.in[0])+":"+itoa(tt.in[1])+"x"+itoa(tt.k), func(t *testing.T) {
			t.Parallel()

			tp := tuplet.MustNew(tt.in[0], tt.in[1])

			got, err := tp.Scale(tt.k)
			require.NoError(t, err)
			assertTuplet(t, got, tt.want[0], tt.want[1])
			assert.Equal(t, got, tp.MustScale(tt.k))

			equal, err := got.EqualCoefficient(tp)
			require.NoError(t, err)
			assert.True(t, equal, "Scale keeps the coefficient value")
		})
	}
}

// TestScale_Errors covers K-7, E-5, E-6 and the order E-2 -> E-6 -> E-5.
func TestScale_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tp   tuplet.Tuplet
		k    int
		want error
	}{
		{name: "zero factor", tp: tuplet.Triplet(), k: 0, want: tuplet.ErrInvalidArgument},
		{name: "negative factor", tp: tuplet.Triplet(), k: -1, want: tuplet.ErrInvalidArgument},
		{name: "min int factor", tp: tuplet.Triplet(), k: math.MinInt, want: tuplet.ErrInvalidArgument},
		{name: "actual overflow", tp: tuplet.MustNew(math.MaxInt, 1), k: 2, want: tuplet.ErrOutOfRange},
		{name: "normal overflow", tp: tuplet.MustNew(1, math.MaxInt), k: 2, want: tuplet.ErrOutOfRange},
		{name: "overflow just above max", tp: tuplet.MustNew(2, 1), k: 1 << 62, want: tuplet.ErrOutOfRange},
		{name: "large factor", tp: tuplet.Triplet(), k: math.MaxInt, want: tuplet.ErrOutOfRange},
		{name: "unset before factor", tp: tuplet.Tuplet{}, k: 0, want: tuplet.ErrUnset},
		{name: "unset before overflow", tp: tuplet.Tuplet{}, k: math.MaxInt, want: tuplet.ErrUnset},
		{name: "factor before overflow", tp: tuplet.MustNew(math.MaxInt, 1), k: -2, want: tuplet.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.tp.Scale(tt.k)
			assertErr(t, err, tt.want)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}
}
