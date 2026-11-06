package cents_test

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allocRepeats is the number of measurements of one call; the least one is
// taken.
const allocRepeats = 3

// allocatedBytes returns the least number of bytes allocated by one call of
// run(i), i = 0...allocRepeats-1: the difference of
// runtime.MemStats.TotalAlloc around the call, after runtime.GC.
func allocatedBytes(run func(i int)) uint64 {
	least := uint64(math.MaxUint64)

	var before, after runtime.MemStats
	for i := range allocRepeats {
		runtime.GC()
		runtime.ReadMemStats(&before)
		run(i)
		runtime.ReadMemStats(&after)
		least = min(least, after.TotalAlloc-before.TotalAlloc)
	}

	return least
}

// assertLinearMemory checks the memory criterion of a rejected long text:
// the bytes of the long input are at most the bytes of the reference input
// of the same form plus 2*len(long input).
func assertLinearMemory(t *testing.T, long, ref uint64, longLen, refLen int) bool {
	t.Helper()

	limit := ref + 2*uint64(longLen)
	msg := fmt.Sprintf("long input %d bytes: %d bytes allocated (%.1f per input byte); "+
		"reference input %d bytes: %d bytes allocated; limit %d bytes (%.1f times the limit)",
		longLen, long, float64(long)/float64(longLen), refLen, ref, limit, float64(long)/float64(limit))

	return assert.LessOrEqual(t, long, limit, msg)
}

// NF-11, KR-C15, K-25, K-32, E-2; regression test of review-tests-from-code
// T-0: a text of the general form with far more than MaxFactors factors is
// rejected with only ErrTooComplex, and its parse allocates no more than
// the shortest text rejected by the limit (MaxFactors+1 factors of the same
// form) plus 2 bytes per byte of input: the cost of a rejected text is set
// by the limits, beyond a linear reading of the input (architecture "Limits
// of the text length", NF-11, doc.go "Performance"). Time is not checked.
//
// The test does not call t.Parallel and has no parallel subtests:
// TotalAlloc counts the allocations of the whole process, so it runs in the
// sequential phase of the package tests, while the parallel tests are
// paused. The file is sorted after concurrency_test.go, so the first use of
// the package stays in TestConcurrentUse.
func TestParseManyFactorsMemory(t *testing.T) {
	const many = 1_000_000

	for _, f := range []string{"2", "2^3", "3^(1/2)"} {
		long := repeatFactors(f, f, many)
		ref := repeatFactors(f, f, cents.MaxFactors)
		longBytes, refBytes := []byte(long), []byte(ref)
		require.Equal(t, many+1, strings.Count(long, "*")+1, "factors of the long input")
		require.Equal(t, cents.MaxFactors+1, strings.Count(ref, "*")+1, "factors of the reference input")

		t.Run(f, func(t *testing.T) {
			t.Run("Parse", func(t *testing.T) {
				var got [allocRepeats]cents.Cents
				var errs [allocRepeats]error

				// Warm-up: lazy package state is not counted in either input.
				_, err := cents.Parse(ref)
				requireOnlyError(t, err, cents.ErrTooComplex, "warm-up")

				check := func(what string) {
					for i := range allocRepeats {
						assertOnlyError(t, errs[i], cents.ErrTooComplex, "%s input, call %d", what, i)
						assert.True(t, got[i] == cents.Cents{}, "%s input, call %d: the zero value", what, i)
					}
				}

				refAlloc := allocatedBytes(func(i int) { got[i], errs[i] = cents.Parse(ref) })
				check("reference")

				longAlloc := allocatedBytes(func(i int) { got[i], errs[i] = cents.Parse(long) })
				check("long")

				assertLinearMemory(t, longAlloc, refAlloc, len(long), len(ref))
			})

			t.Run("UnmarshalText", func(t *testing.T) {
				var recv [allocRepeats]cents.Cents
				var errs [allocRepeats]error

				warm := cents.SyntonicComma()
				requireOnlyError(t, warm.UnmarshalText(refBytes), cents.ErrTooComplex, "warm-up")

				reset := func() {
					for i := range recv {
						recv[i] = cents.SyntonicComma()
					}
				}

				check := func(what string) {
					for i := range allocRepeats {
						assertOnlyError(t, errs[i], cents.ErrTooComplex, "%s input, call %d", what, i)
						assertCents(t, cents.SyntonicComma(), recv[i], "%s input, call %d: the receiver is unchanged", what, i)
					}
				}

				reset()
				refAlloc := allocatedBytes(func(i int) { errs[i] = recv[i].UnmarshalText(refBytes) })
				check("reference")

				reset()
				longAlloc := allocatedBytes(func(i int) { errs[i] = recv[i].UnmarshalText(longBytes) })
				check("long")

				assertLinearMemory(t, longAlloc, refAlloc, len(longBytes), len(refBytes))
			})
		})
	}
}
