package pcset_test

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOwnershipAndImmutability(t *testing.T) {
	t.Parallel()

	t.Run("input_mutation_does_not_affect_constructed_sets", func(t *testing.T) {
		t.Parallel()

		nums := []int{0, 4, 7}
		sNew := pcset.MustNew(12, nums...)
		nums[0] = 11
		assertSet(t, sNew, 12, []int{0, 4, 7})

		xs := []int{-12, 16, 7}
		sRed := pcset.MustReduce(12, xs...)
		xs[1] = 1
		assertSet(t, sRed, 12, []int{0, 4, 7})

		classes := []pitchclass.Class{
			mustClass(t, 0, 12),
			mustClass(t, 4, 12),
			mustClass(t, 7, 12),
		}
		sCls := pcset.MustFromClasses(classes...)
		sClsMod := pcset.MustFromClassesWithModulus(12, classes...)
		classes[0] = mustClass(t, 9, 12)
		assertSet(t, sCls, 12, []int{0, 4, 7})
		assertSet(t, sClsMod, 12, []int{0, 4, 7})

		mask65 := bitmaskBigOracle([]int{0, 64})
		maskCopy := new(big.Int).Set(mask65)
		sBig := pcset.MustFromBitmaskBig(mask65, 65)
		// Constructor must not mutate caller's *big.Int.
		assert.Zero(t, maskCopy.Cmp(mask65))
		// Mutating caller's *big.Int after construction must not affect Set.
		mask65.SetInt64(0)
		assertSet(t, sBig, 65, []int{0, 64})
	})

	t.Run("returned_slices_and_big_ints_are_fresh_copies", func(t *testing.T) {
		t.Parallel()

		for _, s := range []pcset.Set{
			pcset.MustNew(12, 0, 4, 7),
			pcset.MustNew(65, 0, 32, 64),
		} {
			mod := s.Modulus()
			origNums := s.Numbers()
			origClasses := s.Classes()

			mutNums := s.Numbers()
			mutNums[0] = 999
			assert.Equal(t, origNums, s.Numbers())

			mutClasses := s.Classes()
			mutClasses[0] = mustClass(t, 1, mod)
			assert.Equal(t, origClasses, s.Classes())

			origNF := s.NormalForm()
			mutNF := s.NormalForm()
			mutNF[0] = mustClass(t, 2, mod)
			assert.Equal(t, origNF, s.NormalForm())

			mutNFWith := s.MustNormalFormWith(pcset.Forte)
			mutNFWith[0] = mustClass(t, 3, mod)
			assert.Equal(t, origNF, s.MustNormalFormWith(pcset.Forte))

			rots := s.Rotations()
			origRot0 := append([]pitchclass.Class(nil), rots[0]...)
			rots[0][0] = mustClass(t, 5, mod)
			rots[0] = nil
			freshRots := s.Rotations()
			require.NotEmpty(t, freshRots)
			assert.Equal(t, origRot0, freshRots[0])

			b1 := s.MustBitmaskBig()
			b1.SetInt64(12345)
			b2 := s.MustBitmaskBig()
			assert.Zero(t, bitmaskBigOracle(origNums).Cmp(b2))
		}

		// Operator slices ownership.
		aug := pcset.MustNew(12, 0, 4, 8)
		scale := pcset.MustNew(12, 0, 1, 2, 4, 5, 6, 8, 9, 10)
		for _, getter := range []func() []pitchclass.Operator{
			func() []pitchclass.Operator { return aug.MustEquivalenceOperatorsTI(aug) },
			func() []pitchclass.Operator { return aug.MustEquivalenceOperatorsTTO(aug) },
			func() []pitchclass.Operator { return aug.MustEmbeddingOperatorsTI(scale) },
			func() []pitchclass.Operator { return aug.MustEmbeddingOperatorsTTO(scale) },
			func() []pitchclass.Operator { return aug.MustSymmetryOperatorsTI() },
			func() []pitchclass.Operator { return aug.MustSymmetryOperatorsTTO() },
		} {
			first := getter()
			require.NotEmpty(t, first)
			saved := first[0]
			first[0] = pitchclass.Operator{}
			second := getter()
			assert.Equal(t, saved, second[0])
		}

		// IntervalVector.Slice ownership (small and large modulus).
		for _, iv := range []pcset.IntervalVector{
			pcset.MustNew(12, 0, 4, 7).IntervalVector(),
			pcset.MustNew(65, 0, 1, 64).IntervalVector(),
		} {
			s1 := iv.Slice()
			saved0 := s1[0]
			s1[0] = 777_777
			assert.Equal(t, saved0, iv.Slice()[0])
			assert.Equal(t, saved0, iv.MustCount(1))
		}

		// InvarianceVector.Slice ownership.
		invV := pcset.MustNew(12, 0, 4, 8).MustInvarianceVector()
		invSlice := invV.Slice()
		invSlice[0] = 999
		assert.Equal(t, 3, invV.T())
		assert.Equal(t, 3, invV.Slice()[0])

		// CountBig and CountOfCardinalityBig ownership.
		cb1 := pcset.MustCountBig(65)
		cb1.SetInt64(0)
		cb2 := pcset.MustCountBig(65)
		assert.Equal(t, 66, cb2.BitLen())

		ck1 := pcset.MustCountOfCardinalityBig(67, 33)
		ck1.SetInt64(0)
		ck2 := pcset.MustCountOfCardinalityBig(67, 33)
		assert.Zero(t, binomBigOracle(67, 33).Cmp(ck2))
	})
}

func TestTypeReflectionContract(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 65_536, pcset.MaxModulus)

	for _, rt := range []reflect.Type{
		reflect.TypeFor[pcset.Set](),
		reflect.TypeFor[pcset.IntervalVector](),
		reflect.TypeFor[pcset.InvarianceVector](),
	} {
		t.Run(rt.Name(), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, reflect.Struct, rt.Kind())
			assert.True(t, rt.Comparable(), "%s must be comparable", rt.Name())
			for i := range rt.NumField() {
				f := rt.Field(i)
				assert.False(t, f.IsExported(), "%s field %s must not be exported", rt.Name(), f.Name)
			}
		})
	}

	algType := reflect.TypeFor[pcset.Algorithm]()
	assert.Equal(t, reflect.Uint8, algType.Kind())
	assert.True(t, algType.Comparable())
}
