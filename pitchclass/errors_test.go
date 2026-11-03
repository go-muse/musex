package pitchclass_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-30: the seven sentinel errors are distinct: none matches another with
// errors.Is, and a wrapped sentinel still matches only itself.
func TestSentinelErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	all := allErrors()
	for i, a := range all {
		require.Error(t, a, "sentinel %d", i)
		for j, b := range all {
			assert.Equal(t, i == j, errors.Is(a, b), "errors.Is(%v, %v)", a, b)
		}

		wrapped := fmt.Errorf("consumer context: %w", a)
		assertOnlyError(t, wrapped, a, "wrapped sentinel")
	}
}

// K-28, E-3: every two-operand operation with operands of N = 12 and
// N = 19 is ErrModulusMismatch; equality is "not equal" and the order is
// defined without an error.
func TestDifferentModuli(t *testing.T) {
	t.Parallel()

	x12, y12 := newClass(t, 1, 12), newClass(t, 4, 12)
	x19, y19 := newClass(t, 1, 19), newClass(t, 4, 19)
	f12, f19 := pitchclass.MustTransposition(5, 12), pitchclass.MustTransposition(5, 19)

	tests := []struct {
		name string
		call func() (any, error)
	}{
		{name: "IntervalTo 12 -> 19", call: func() (any, error) { return x12.IntervalTo(x19) }},
		{name: "IntervalTo 19 -> 12", call: func() (any, error) { return x19.IntervalTo(x12) }},
		{name: "IntervalClass 12, 19", call: func() (any, error) { return x12.IntervalClass(x19) }},
		{name: "IntervalClass 19, 12", call: func() (any, error) { return x19.IntervalClass(x12) }},
		{name: "IndexSum 12, 19", call: func() (any, error) { return x12.IndexSum(x19) }},
		{name: "IndexSum 19, 12", call: func() (any, error) { return x19.IndexSum(x12) }},
		{name: "InvertAround class 12, axis 19", call: func() (any, error) { return x12.InvertAround(x19) }},
		{name: "InvertAround class 19, axis 12", call: func() (any, error) { return x19.InvertAround(x12) }},
		{name: "InvertSwapping 12 (19, 19)", call: func() (any, error) { return x12.InvertSwapping(x19, y19) }},
		{name: "InvertSwapping 12 (19, 12)", call: func() (any, error) { return x12.InvertSwapping(x19, y12) }},
		{name: "InvertSwapping 12 (12, 19)", call: func() (any, error) { return x12.InvertSwapping(y12, x19) }},
		{name: "InvertSwapping 19 (12, 12)", call: func() (any, error) { return x19.InvertSwapping(x12, y12) }},
		{name: "InversionSwapping (12, 19)", call: func() (any, error) { return pitchclass.InversionSwapping(x12, x19) }},
		{name: "InversionSwapping (19, 12)", call: func() (any, error) { return pitchclass.InversionSwapping(x19, x12) }},
		{name: "Apply operator 12, class 19", call: func() (any, error) { return f12.Apply(x19) }},
		{name: "Apply operator 19, class 12", call: func() (any, error) { return f19.Apply(x12) }},
		{name: "Compose 12∘19", call: func() (any, error) { return f12.Compose(f19) }},
		{name: "Compose 19∘12", call: func() (any, error) { return f19.Compose(f12) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got any
				err error
			)
			require.NotPanics(t, func() { got, err = tc.call() })
			assertOnlyError(t, err, pitchclass.ErrModulusMismatch, tc.name)
			assertZeroResult(t, got, tc.name)
		})
	}

	t.Run("equality and order", func(t *testing.T) {
		t.Parallel()

		for _, pair := range [][2]pitchclass.Class{{x12, x19}, {y12, y19}, {x12, y19}} {
			a, b := pair[0], pair[1]
			assert.False(t, a.Equal(b), "%v.Equal(%v)", a, b)
			assert.False(t, b.Equal(a), "%v.Equal(%v)", b, a)
			require.NotPanics(t, func() { _ = a.Compare(b) })
			assert.Equal(t, -1, a.Compare(b), "%v before %v", a, b)
			assert.Equal(t, 1, b.Compare(a), "%v after %v", b, a)
		}

		assert.False(t, f12.Equal(f19))
		assert.False(t, f19.Equal(f12))
		require.NotPanics(t, func() { _ = f12.Compare(f19) })
		assert.Equal(t, -1, f12.Compare(f19))
		assert.Equal(t, 1, f19.Compare(f12))
	})
}
