package pcset_test

import (
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentSharedReadSafety(t *testing.T) {
	t.Parallel()

	shared12A := pcset.MustNew(12, 0, 1, 3, 7, 8)
	shared12B := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)
	shared65A := pcset.MustNew(65, 0, 10, 32, 64)
	shared65B := pcset.MustNew(65, 0, 10, 11, 32, 40, 64)
	sharedIV := shared12A.IntervalVector()
	sharedInvV := shared12A.MustInvarianceVector()
	sharedSeqSet := shared12A.All()
	sharedSeqAll := pcset.MustAll(4)
	sharedSeqCard := pcset.MustOfCardinality(5, 2)
	op := mustOp(t, 5, 3, 12)
	c4 := mustClass(t, 4, 12)

	workers := max(4, 4*runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup

	for w := range workers {
		wg.Go(func() {
			for i := range 25 {
				step := w + i

				assert.Equal(t, 12, shared12A.Modulus())
				assert.Equal(t, 5, shared12A.Cardinality())
				assert.True(t, shared12A.ContainsNumber(0))
				assert.True(t, shared12A.Contains(mustClass(t, 0, 12)))

				nums := shared12A.Numbers()
				nums[0] = 999
				classes := shared12A.Classes()
				classes[0] = c4

				_ = shared12A.MustBitmask()
				bmBig := shared12A.MustBitmaskBig()
				bmBig.SetInt64(0)

				_ = shared12A.Transpose(step)
				_ = shared12A.Invert(step)
				_ = shared12A.MustInvertAround(c4)
				_ = shared12A.MustInvertSwapping(c4, c4)
				_ = shared12A.Multiply(step)
				_ = shared12A.MustApply(op)
				_ = shared12A.Complement()
				_ = shared12A.MustUnion(shared12B)
				_ = shared12A.MustIntersection(shared12B)
				_ = shared12A.MustDifference(shared12B)
				_ = shared12A.MustSymmetricDifference(shared12B)
				_ = shared12A.MustIsDisjoint(shared12B)
				_ = shared12A.MustIsSubsetOf(shared12B)
				_ = shared12A.MustCoveringTranspositionsOf(shared12B)
				_ = shared12A.MustTranspositionEmbeddingsInto(shared12B)
				_ = shared12A.MustInversionEmbeddingsInto(shared12B)
				_ = shared12A.MustEquivalenceOperatorsTI(shared12A)
				_ = shared12A.MustEquivalenceOperatorsTTO(shared12A)
				_ = shared12A.MustEmbeddingOperatorsTI(shared12B)
				_ = shared12A.MustEmbeddingOperatorsTTO(shared12B)
				_ = shared12A.MustCommonToneCountTranspose(step)
				_ = shared12A.MustCommonToneCountInvert(step)
				_ = shared12A.MustCommonToneCountWith(shared12B, op)

				rots := shared12A.Rotations()
				rots[0][0] = c4
				nf := shared12A.NormalForm()
				nf[0] = c4
				nfForte := shared12A.MustNormalFormWith(pcset.Forte)
				nfForte[0] = c4

				_ = shared12A.MustDegreeOfSymmetryTI()
				_ = shared12A.MustDegreeOfSymmetryTTO()
				_ = shared12A.MustDistinctTranspositionCount()
				_ = shared12A.MustOrbitSizeTI()
				_ = shared12A.String()
				_ = shared12A.MustNotation(pitchclass.TE)
				_ = shared12A.MustCompactNotation(pitchclass.TE)
				_ = shared12A.MustNormalFormNotation(pitchclass.TE, pcset.Rahn)
				_ = shared12A.MustNormalFormCompactNotation(pitchclass.TE, pcset.Forte)

				// Large modulus (> 64) concurrent reads.
				_ = shared65A.Transpose(step)
				_ = shared65A.Invert(step)
				_ = shared65A.Multiply(step)
				_ = shared65A.Complement()
				_ = shared65A.MustUnion(shared65B)
				_ = shared65A.MustCoveringTranspositionsOf(shared65B)
				_ = shared65A.IntervalVector().Slice()

				// Shared IntervalVector and InvarianceVector reads.
				ivSlice := sharedIV.Slice()
				ivSlice[0] = -1
				_ = sharedIV.MustCount(1)
				_ = sharedIV.String()
				_ = sharedIV.MustNotation()
				_ = sharedIV.MustCompactNotation()

				invSlice := sharedInvV.Slice()
				invSlice[0] = -1
				_ = sharedInvV.SymmetryTotal()
				_ = sharedInvV.ComplementTotal()
				_ = sharedInvV.String()
				_ = sharedInvV.MustNotation()

				// Shared iterators.
				assert.Len(t, slices.Collect(sharedSeqSet), 5)
				assert.Len(t, slices.Collect(sharedSeqAll), 16)
				assert.Len(t, slices.Collect(sharedSeqCard), 10)
			}
		})
	}

	wg.Wait()
}
