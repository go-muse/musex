package spelling_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// appendBufferCase is one call of AppendText on a buffer whose backing array
// holds prefix followed by spare bytes of free capacity.
type appendBufferCase struct {
	name   string
	s      spelling.Spelling
	prefix string
	spare  int
}

// appendBufferCases combines names with forms of different lengths, up to
// the longest machine form, two prefixes, and spare capacities that are too
// small for the form (none, less than the form, the form length minus one) or
// large enough (exactly the form length, more than the form).
func appendBufferCases() []appendBufferCase {
	var unset spelling.Spelling

	names := []spelling.Spelling{
		spelling.C(),
		spelling.CSharp(),
		spelling.MustNew(letter.B, frac(-7, 2)),
		unset,
		// A long form with a five-digit denominator: "B-16777215/65536".
		spelling.MustNew(letter.B, frac(-(maxMagnitude-1), maxDenominator)),
		// The longest form: the largest numerator allowed with the largest
		// denominator, "B-1099511627775/65536" (21 bytes).
		spelling.MustNew(letter.B, frac(-(maxMagnitude*maxDenominator-1), maxDenominator)),
	}

	var cases []appendBufferCase

	for _, s := range names {
		n := len(s.String())

		spares := []int{0, 1, n / 2, n - 1, n, n + 3}
		slices.Sort(spares)
		spares = slices.Compact(spares)

		for _, prefix := range []string{"a", "key="} {
			for _, spare := range spares {
				cases = append(cases, appendBufferCase{
					name:   fmt.Sprintf("%s prefix %q spare %d", s.String(), prefix, spare),
					s:      s,
					prefix: prefix,
					spare:  spare,
				})
			}
		}
	}

	return cases
}

// newAppendBuffer returns a backing array that holds prefix and then spare
// marker bytes, and a buffer over it with length len(prefix) and capacity
// len(prefix)+spare.
func newAppendBuffer(prefix string, spare int) (array, buf []byte) {
	array = []byte(prefix + strings.Repeat(".", spare))

	return array, array[:len(prefix):len(array)]
}

// AppendText behaves as one call of append(b, s.String()...): when the spare
// capacity of b cannot hold the whole form, the old backing array is left
// untouched and the result is a new array; when it can, the result shares the
// array of b and the form follows b[:len(b)]. The oracle is append on an
// independent copy of the same array.
func TestAppendTextBufferReuse(t *testing.T) {
	t.Parallel()

	for _, tc := range appendBufferCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			form := tc.s.String()
			fits := tc.spare >= len(form)

			wantArray, wantBuf := newAppendBuffer(tc.prefix, tc.spare)
			want := append(wantBuf, form...)

			array, buf := newAppendBuffer(tc.prefix, tc.spare)

			got, err := tc.s.AppendText(buf)
			require.NoError(t, err)

			assert.Equal(t, string(want), string(got), "the returned text")
			assert.Equal(t, tc.prefix+form, string(got), "the prefix followed by the form")
			assert.Equal(t, string(wantArray), string(array), "the old backing array")
			assert.Equal(t, tc.prefix, string(buf), "b[:len(b)]")

			if fits {
				require.NotEmpty(t, got)
				assert.Same(t, &array[0], &got[0], "the result must reuse the array of b")
				assert.Equal(t, form, string(array[len(tc.prefix):len(tc.prefix)+len(form)]),
					"the form must follow the prefix in the array of b")

				return
			}

			assert.Equal(t, strings.Repeat(".", tc.spare), string(array[len(tc.prefix):]),
				"the spare capacity must stay untouched")

			if len(array) > 0 && len(got) > 0 {
				assert.NotSame(t, &array[0], &got[0], "the result must be a new array")
			}

		})
	}
}
