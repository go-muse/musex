package fraction_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
)

// BenchmarkRoundToMultiple measures the new RoundToMultiple of F-1 (task 26.6)
// on the inputs of the division benchmarks, for the report and T-2 only: the
// method has no baseline, so it is kept out of the baseline benchmark files.
func BenchmarkRoundToMultiple(b *testing.B) {
	benchEach(b, benchDivMod, func(in benchInput) error {
		_, _, err := in.x.RoundToMultiple(in.y, fraction.RoundHalfToEven)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _, _ = x.RoundToMultiple(y, fraction.RoundHalfToEven)
		}
	})
}
