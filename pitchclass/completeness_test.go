package pitchclass_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-33, NF-4: for all classes of N in {1, 2, 12, 19} and n, m in -30...30
// the operations C-3...C-5, C-7, C-8, C-14, C-18, C-22, C-25, C-26, C-29,
// C-31 (Decimal) and C-35 give no error and do not panic.
func TestOperationsWithoutErrors(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("classes N=%d", size), func(t *testing.T) {
			t.Parallel()

			cs := classesOf(t, size)
			other := newClass(t, 0, 19)
			for _, c := range cs {
				cycled := 0
				require.NotPanics(t, func() {
					_ = c.Number()
					_ = c.Modulus()
					_ = c.IsValid()
					_ = c.Equal(other)
					_ = c.Compare(other)
					_ = c.String()
					for range c.Cycle() {
						cycled++
					}

					for v := -30; v <= 30; v++ {
						_ = c.Transpose(v)
						_ = c.Invert(v)
						_ = c.Multiply(v)
					}
				}, "%v", c)
				assert.Equal(t, size, cycled, "Cycle of %v", c)

				s, err := c.Notation(pitchclass.Decimal)
				assert.NoError(t, err, "Notation(Decimal) of %v", c)
				assert.Equal(t, fmt.Sprint(c.Number()), s)

				for _, d := range cs {
					assert.NotPanics(t, func() {
						_ = c.Equal(d)
						_ = c.Compare(d)
					})
				}
			}
		})

		t.Run(fmt.Sprintf("operators N=%d", size), func(t *testing.T) {
			t.Parallel()

			var ops []pitchclass.Operator
			for m := -30; m <= 30; m++ {
				for n := -30; n <= 30; n++ {
					if f, err := pitchclass.NewOperator(m, n, size); err == nil {
						ops = append(ops, f)
					}
				}
				ops = append(ops, pitchclass.MustTransposition(m, size), pitchclass.MustInversion(m, size))
			}
			require.NotEmpty(t, ops)

			other := pitchclass.MustIdentity(19)
			for _, f := range ops {
				require.NotPanics(t, func() {
					_ = f.Modulus()
					_ = f.Multiplier()
					_ = f.Index()
					_ = f.IsValid()
					_ = f.IsTransposition()
					_ = f.IsInversion()
					_ = f.Inverse()
					_ = f.Equal(other)
					_ = f.Compare(other)
					_ = f.String()
				}, "%v", f)
			}
		})
	}
}
