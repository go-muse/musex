package letter_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/letter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-7: among all 256 values of the underlying uint8 exactly the seven
// constants are valid.
func TestLetter_IsValid_AllValues(t *testing.T) {
	t.Parallel()

	valid := make(map[letter.Letter]bool, oracleCount)
	for v := range math.MaxUint8 + 1 {
		l := letter.Letter(v)
		if l.IsValid() {
			valid[l] = true
		}
	}

	want := make(map[letter.Letter]bool, oracleCount)
	for _, l := range oracle() {
		want[l] = true
	}

	assert.Equal(t, want, valid)
}

// K-7: every operation has the defined result for every invalid value and
// never returns a result possible for a valid letter.
func TestInvalidValues(t *testing.T) {
	t.Parallel()

	shiftSteps := []int{-8, -1, 0, 1, 7, math.MinInt, math.MaxInt}
	names := oracleNames()

	for v := range math.MaxUint8 + 1 {
		invalid := letter.Letter(v)
		if isOracleLetter(invalid) {
			continue
		}

		t.Run("Letter("+strconv.Itoa(v)+")", func(t *testing.T) {
			t.Parallel()

			assert.False(t, invalid.IsValid(), "IsValid")
			assert.Equal(t, -1, invalid.Ordinal(), "Ordinal")
			assert.Nil(t, invalid.Cycle(), "Cycle")
			assert.Len(t, invalid.Cycle(), 0, "Cycle length")

			str := invalid.String()
			assert.Equal(t, fmt.Sprintf("Letter(%d)", v), str, "String")
			assert.NotContains(t, names[:], str, "String never looks like a letter")

			assert.False(t, invalid.Equal(invalid), "Equal with itself")
			assert.False(t, invalid.Equal(letter.Letter(0)), "Equal with the zero value")
			assert.Equal(t, -1, invalid.DistanceUp(invalid), "DistanceUp with itself")
			assert.Equal(t, -1, invalid.DistanceDown(invalid), "DistanceDown with itself")

			for ord, valid := range oracle() {
				name := names[ord]
				assert.False(t, invalid.Equal(valid), "invalid receiver, Equal(%s)", name)
				assert.False(t, valid.Equal(invalid), "invalid argument, %s.Equal", name)
				assert.NotEqual(t, valid, invalid, "%s != invalid", name)
				assert.Equal(t, -1, invalid.DistanceUp(valid), "invalid receiver, DistanceUp(%s)", name)
				assert.Equal(t, -1, valid.DistanceUp(invalid), "invalid argument, %s.DistanceUp", name)
				assert.Equal(t, -1, invalid.DistanceDown(valid), "invalid receiver, DistanceDown(%s)", name)
				assert.Equal(t, -1, valid.DistanceDown(invalid), "invalid argument, %s.DistanceDown", name)
			}

			for _, n := range shiftSteps {
				shifted := invalid.Shift(n)
				assert.Equal(t, letter.Letter(0), shifted, "Shift(%d) is the zero Letter", n)
				assert.False(t, shifted.IsValid(), "Shift(%d) is invalid", n)
			}
		})
	}
}

// K-7: the zero value means "no letter set"; it is not C.
func TestZeroValue(t *testing.T) {
	t.Parallel()

	var zero letter.Letter

	assert.False(t, zero.IsValid())
	assert.NotEqual(t, letter.C, zero)
	assert.False(t, zero.Equal(letter.C))
	assert.False(t, letter.C.Equal(zero))
	assert.False(t, zero.Equal(zero), "an invalid value is not a letter")
	assert.Equal(t, "Letter(0)", zero.String())
	assert.Equal(t, -1, zero.Ordinal())
	assert.Equal(t, -1, zero.DistanceUp(letter.C))
	assert.Equal(t, -1, letter.C.DistanceDown(zero))
	assert.False(t, zero.Shift(1).IsValid())
	assert.Nil(t, zero.Cycle())
}

// K-7: integer arithmetic on letters is not a shift around the circle.
func TestArithmeticIsNotShift(t *testing.T) {
	t.Parallel()

	past := letter.B + 1

	assert.False(t, past.IsValid())
	assert.NotEqual(t, letter.C, past)
	assert.Equal(t, letter.C, letter.B.Shift(1), "Shift wraps around instead")
}

// K-7: invalid ordinals give the zero Letter and an error matching ErrInvalid.
func TestFromOrdinal_Invalid(t *testing.T) {
	t.Parallel()

	for _, ordinal := range invalidOrdinals() {
		t.Run(strconv.Itoa(ordinal), func(t *testing.T) {
			t.Parallel()

			got, err := letter.FromOrdinal(ordinal)
			require.Error(t, err)
			assert.ErrorIs(t, err, letter.ErrInvalid)
			assert.Equal(t, letter.Letter(0), got, "the zero Letter")
			assert.False(t, got.IsValid())
		})
	}
}

// K-7: the Must wrapper panics with an error matching ErrInvalid. The error
// text is not part of the contract, so assert.PanicsWithError is not used.
func TestMustFromOrdinal_Invalid(t *testing.T) {
	t.Parallel()

	for _, ordinal := range invalidOrdinals() {
		t.Run(strconv.Itoa(ordinal), func(t *testing.T) {
			t.Parallel()

			recovered, panicked := recoverPanic(func() { letter.MustFromOrdinal(ordinal) })
			require.True(t, panicked, "MustFromOrdinal(%d) must panic", ordinal)

			err, ok := recovered.(error)
			require.True(t, ok, "panic value %#v must be an error", recovered)
			assert.ErrorIs(t, err, letter.ErrInvalid, "panic error must match ErrInvalid")
		})
	}
}

// K-10: changing a returned list does not affect later results or other
// callers.
func TestAll_Ownership(t *testing.T) {
	t.Parallel()

	want := oracle()

	t.Run("overwrite every element", func(t *testing.T) {
		t.Parallel()

		mine := letter.All()
		other := letter.All()
		for i := range mine {
			mine[i] = letter.B
		}

		assert.Equal(t, want[:], letter.All(), "a later call")
		assert.Equal(t, want[:], other, "a list held by another caller")
	})

	t.Run("append into the spare capacity of a prefix", func(t *testing.T) {
		t.Parallel()

		prefix := letter.All()[:3]
		grown := append(prefix, letter.B, letter.B, letter.B, letter.B)
		require.Len(t, grown, oracleCount)

		assert.Equal(t, want[:], letter.All())
	})
}

// K-10: changing a returned cycle does not affect later results or other
// callers, for every start letter.
func TestLetter_Cycle_Ownership(t *testing.T) {
	t.Parallel()

	names := oracleNames()
	for ord, start := range oracle() {
		t.Run(names[ord], func(t *testing.T) {
			t.Parallel()

			want := make([]letter.Letter, oracleCount)
			for i := range want {
				want[i] = expectedShift(ord, i)
			}

			mine := start.Cycle()
			other := start.Cycle()
			for i := range mine {
				mine[i] = letter.Letter(0)
			}
			grown := append(start.Cycle()[:1], letter.Letter(0), letter.Letter(0))
			require.Len(t, grown, 3)

			assert.Equal(t, want, start.Cycle(), "a later call")
			assert.Equal(t, want, other, "a cycle held by another caller")
		})
	}
}

// invalidOrdinals returns ordinals outside 0...6.
func invalidOrdinals() []int {
	return []int{-1, 7, 8, 100, math.MinInt, math.MaxInt}
}

// isOracleLetter reports whether l is one of the seven constants.
func isOracleLetter(l letter.Letter) bool {
	for _, valid := range oracle() {
		if l == valid {
			return true
		}
	}

	return false
}
