package number_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
)

// zeroValues returns the three forms of the undefined numbering: a declared
// variable, the composite literal and the result of New with an error.
func zeroValues() []struct {
	name string
	n    number.Numbering
} {
	var declared number.Numbering
	failed, _ := number.New(4, 3, 3)

	return []struct {
		name string
		n    number.Numbering
	}{
		{name: "var", n: declared},
		{name: "literal", n: number.Numbering{}},
		{name: "New with an error", n: failed},
	}
}

// K-13, E-4: the zero value is an undefined numbering. Every operation is
// defined for it and none returns the result of ASPN or another numbering.
func TestZeroValue(t *testing.T) {
	t.Parallel()

	for _, z := range zeroValues() {
		t.Run(z.name, func(t *testing.T) {
			t.Parallel()

			n := z.n
			undefined := outcome{err: number.ErrUndefined}

			assert.False(t, n.IsDefined(), "IsDefined")
			assert.Empty(t, n.Label(), "Label")
			assert.Equal(t, "Numbering(undefined)", n.String(), "String")

			for _, a := range []int{minInt, -1, 0, 4, maxInt} {
				got, err := n.FromASPN(a)
				assertOutcome(t, undefined, got, err, fmt.Sprintf("FromASPN(%d)", a))
				assertPanicsWithErrorIs(t, number.ErrUndefined, func() { n.MustFromASPN(a) }, fmt.Sprintf("MustFromASPN(%d)", a))
			}

			for _, x := range []int{minInt, 0, 4, maxInt} {
				got, err := n.ToASPN(x)
				assertOutcome(t, undefined, got, err, fmt.Sprintf("ToASPN(%d)", x))
				assertPanicsWithErrorIs(t, number.ErrUndefined, func() { n.MustToASPN(x) }, fmt.Sprintf("MustToASPN(%d)", x))
			}

			for _, b := range builtins() {
				for _, x := range []int{minInt, 0, 4, maxInt} {
					got, err := number.Convert(x, n, b.n)
					assertOutcome(t, undefined, got, err, fmt.Sprintf("Convert(%d, zero, %s)", x, b.name))

					got, err = number.Convert(x, b.n, n)
					assertOutcome(t, undefined, got, err, fmt.Sprintf("Convert(%d, %s, zero)", x, b.name))

					assertPanicsWithErrorIs(t, number.ErrUndefined, func() { number.MustConvert(x, n, b.n) },
						fmt.Sprintf("MustConvert(%d, zero, %s)", x, b.name))
					assertPanicsWithErrorIs(t, number.ErrUndefined, func() { number.MustConvert(x, b.n, n) },
						fmt.Sprintf("MustConvert(%d, %s, zero)", x, b.name))
				}

				assert.False(t, n.Equal(b.n), "zero.Equal(%s)", b.name)
				assert.False(t, b.n.Equal(n), "%s.Equal(zero)", b.name)
			}

			for _, x := range []int{minInt, 0, 4, maxInt} {
				got, err := number.Convert(x, n, n)
				assertOutcome(t, undefined, got, err, fmt.Sprintf("Convert(%d, zero, zero)", x))
			}

			mine := own(t, 4, 4)
			assert.False(t, n.Equal(mine.n), "zero.Equal(New(4, 4))")
			assert.False(t, mine.n.Equal(n), "New(4, 4).Equal(zero)")

			assert.False(t, n.Equal(n), "zero.Equal(zero)")
			for _, other := range zeroValues() {
				assert.False(t, n.Equal(other.n), "zero.Equal(%s)", other.name)
			}
		})
	}
}
