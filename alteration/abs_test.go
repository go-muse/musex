package alteration_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Abs returns the exact magnitude of a value. The guaranteed range is
// symmetric, so the magnitude of every value is representable and Abs never
// fails.
func TestAbs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   frac
		want frac
	}{
		{"natural", frac{0, 1}, frac{0, 1}},
		{"sharp", frac{1, 1}, frac{1, 1}},
		{"flat", frac{-1, 1}, frac{1, 1}},
		{"double sharp", frac{2, 1}, frac{2, 1}},
		{"double flat", frac{-2, 1}, frac{2, 1}},
		{"negative half", frac{-1, 2}, frac{1, 2}},
		{"positive half", frac{1, 2}, frac{1, 2}},
		{"negative three halves", frac{-3, 2}, frac{3, 2}},
		{"unreduced negative", frac{-2, 4}, frac{1, 2}},
		{"top", edgeTop(), edgeTop()},
		{"bottom", edgeBottom(), edgeTop()},
		{"inverse denominator", edgeInvQ(), edgeInvQ()},
		{"negative inverse denominator", edgeNegInvQ(), edgeInvQ()},
		{"near top", edgeNearTop(), edgeNearTop()},
		{"near bottom", edgeNearBottom(), edgeNearTop()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := value(t, tc.in)
			before := a
			got := a.Abs()

			assert.Equal(t, tc.want, fractionOf(got))
			assertExact(t, new(big.Rat).Abs(tc.in.rat()), got)
			assert.True(t, got.Sign() >= 0, "the magnitude does not lower")
			assertSame(t, before, a, "Abs does not change the receiver")
			assertSame(t, got, got.Abs(), "Abs is idempotent")
		})
	}

	t.Run("opposite values have one magnitude", func(t *testing.T) {
		t.Parallel()

		for _, f := range setS() {
			a := value(t, f)
			neg, err := a.Mul(-1)
			require.NoError(t, err)
			assertSame(t, a.Abs(), neg.Abs(), "%s", f)
		}
	})

	t.Run("the proof set keeps its magnitudes", func(t *testing.T) {
		t.Parallel()

		for _, f := range setS() {
			a := value(t, f)
			assertExact(t, new(big.Rat).Abs(f.rat()), a.Abs(), "%s", f)
		}
	})
}

// The magnitude of a negative value is the positive value of the same size;
// the natural is its own magnitude.
func ExampleAlteration_Abs() {
	fmt.Println(alteration.DoubleFlat().Abs())
	fmt.Println(alteration.MustFromFraction(-1, 2).Abs())
	fmt.Println(alteration.Natural().Abs())
	// Output:
	// +2
	// +1/2
	// 0
}
