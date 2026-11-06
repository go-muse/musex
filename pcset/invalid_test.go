package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
)

func TestZeroAndInvalidValuesContract(t *testing.T) {
	t.Parallel()

	t.Run("zero_set_contract", func(t *testing.T) {
		t.Parallel()

		var zero pcset.Set
		valid12 := pcset.MustNew(12, 0, 4, 7)
		valid19 := pcset.MustEmpty(19)

		// (1) IsValid == false.
		assert.False(t, zero.IsValid())
		assert.True(t, valid12.IsValid())

		// (2) Numeric and boolean properties on invalid Set.
		assert.Zero(t, zero.Modulus())
		assert.Zero(t, zero.Cardinality())
		assert.False(t, zero.IsEmpty())
		assert.False(t, zero.IsUniverse())
		assert.False(t, zero.Contains(mustClass(t, 0, 12)))
		assert.False(t, zero.Contains(pitchclass.Class{}))
		assert.False(t, zero.ContainsNumber(0))

		// (3) Equal with itself and valid sets is false.
		assert.False(t, zero.Equal(zero))
		assert.False(t, zero.Equal(valid12))
		assert.False(t, valid12.Equal(zero))

		// (4) Compare with itself is 0, with any valid set is -1 (+1 in reverse).
		assert.Zero(t, zero.Compare(zero))
		assert.Equal(t, -1, zero.Compare(valid12))
		assert.Equal(t, 1, valid12.Compare(zero))
		assert.Equal(t, -1, zero.Compare(pcset.MustEmpty(1)))

		// (5) String returns "Set(invalid)".
		assert.Equal(t, "Set(invalid)", zero.String())

		// (6) Non-error methods do not panic and return nil / empty / invalid.
		assert.Nil(t, zero.Classes())
		assert.Nil(t, zero.Numbers())
		assert.Nil(t, zero.NormalForm())
		assert.Nil(t, zero.Rotations())
		assertEmptySeqContract(t, zero.All())
		assert.Equal(t, zero, zero.Transpose(3))
		assert.Equal(t, zero, zero.Invert(3))
		assert.Equal(t, zero, zero.Multiply(5))
		assert.Equal(t, zero, zero.Complement())
		assert.Equal(t, pcset.IntervalVector{}, zero.IntervalVector())
		assert.False(t, zero.IntervalVector().IsValid())

		// Error-returning methods return ErrInvalidValue with highest precedence.
		_, err := zero.Bitmask()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		_, err = zero.BitmaskBig()
		requireSentinelError(t, err, pcset.ErrInvalidValue)

		_, err = zero.Union(valid19)
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		_, err = valid19.Union(zero)
		requireSentinelError(t, err, pcset.ErrInvalidValue)

		_, err = zero.InvarianceVector()
		requireSentinelError(t, err, pcset.ErrInvalidValue)

		_, err = zero.NormalFormWith(pcset.Algorithm(99))
		requireSentinelError(t, err, pcset.ErrInvalidValue)

		_, err = zero.Notation(pitchclass.Style(99))
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		_, err = zero.CompactNotation(pitchclass.Style(99))
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		_, err = zero.NormalFormNotation(pitchclass.Style(99), pcset.Algorithm(99))
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		_, err = zero.NormalFormCompactNotation(pitchclass.Style(99), pcset.Algorithm(99))
		requireSentinelError(t, err, pcset.ErrInvalidValue)

		// (7) Built-in == is true for Set{} == Set{} and false against valid sets.
		assert.True(t, zero == pcset.Set{})
		assert.False(t, zero == valid12)
	})

	t.Run("zero_interval_vector_contract", func(t *testing.T) {
		t.Parallel()

		var zero pcset.IntervalVector
		valid := pcset.MustNew(12, 0, 4, 7).IntervalVector()

		assert.False(t, zero.IsValid())
		assert.True(t, valid.IsValid())

		assert.Zero(t, zero.Modulus())
		assert.Zero(t, zero.Len())
		assert.Zero(t, zero.Total())
		assert.Nil(t, zero.Slice())

		assert.False(t, zero.Equal(zero))
		assert.False(t, zero.Equal(valid))
		assert.False(t, valid.Equal(zero))

		assert.Zero(t, zero.Compare(zero))
		assert.Equal(t, -1, zero.Compare(valid))
		assert.Equal(t, 1, valid.Compare(zero))

		assert.Equal(t, "IntervalVector(invalid)", zero.String())

		// ErrInvalidValue takes precedence over ErrInvalidIntervalClass.
		for _, d := range []int{0, 1, -1, 100} {
			_, err := zero.Count(d)
			requireSentinelError(t, err, pcset.ErrInvalidValue)
			requirePanicsWithSentinel(t, func() { _ = zero.MustCount(d) }, pcset.ErrInvalidValue)
		}

		_, err := zero.Notation()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNotation() }, pcset.ErrInvalidValue)

		_, err = zero.CompactNotation()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustCompactNotation() }, pcset.ErrInvalidValue)

		assert.True(t, zero == pcset.IntervalVector{})
		assert.False(t, zero == valid)
	})

	t.Run("zero_invariance_vector_contract", func(t *testing.T) {
		t.Parallel()

		var zero pcset.InvarianceVector
		valid := pcset.MustNew(12, 0, 4, 7).MustInvarianceVector()

		assert.False(t, zero.IsValid())
		assert.True(t, valid.IsValid())

		assert.Zero(t, zero.Modulus())
		assert.Zero(t, zero.Len())
		assert.Zero(t, zero.T())
		assert.Zero(t, zero.I())
		assert.Zero(t, zero.M5())
		assert.Zero(t, zero.M7())
		assert.Zero(t, zero.ComplementT())
		assert.Zero(t, zero.ComplementI())
		assert.Zero(t, zero.ComplementM5())
		assert.Zero(t, zero.ComplementM7())
		assert.Zero(t, zero.SymmetryTotal())
		assert.Zero(t, zero.ComplementTotal())
		assert.Nil(t, zero.Slice())

		assert.False(t, zero.Equal(zero))
		assert.False(t, zero.Equal(valid))
		assert.False(t, valid.Equal(zero))

		assert.Zero(t, zero.Compare(zero))
		assert.Equal(t, -1, zero.Compare(valid))
		assert.Equal(t, 1, valid.Compare(zero))

		assert.Equal(t, "InvarianceVector(invalid)", zero.String())

		_, err := zero.Notation()
		requireSentinelError(t, err, pcset.ErrInvalidValue)
		requirePanicsWithSentinel(t, func() { _ = zero.MustNotation() }, pcset.ErrInvalidValue)

		assert.True(t, zero == pcset.InvarianceVector{})
		assert.False(t, zero == valid)
	})
}
