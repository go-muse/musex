package score_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/score"
)

// BenchmarkBarlinePeriod measures s.BarlinePeriod(a, b) for three classes of
// lengths (architecture 6.2 of task 26.8): small — components in one machine
// word; wide — numerators wider than 64 bits with a period in range;
// out-of-range — the numerator of the LCM is greater than K.
func BenchmarkBarlinePeriod(b *testing.B) {
	nearA, nearC := periodNearK(periodBelowK)
	wideA, wideC := periodWide()

	tests := []struct {
		name string
		x, y periodMeter
		fail bool
	}{
		{"small", periodSimple(6, 8), periodSimple(5, 12), false},
		{"wide", nearA, nearC, false},
		{"out-of-range", wideA, wideC, true},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			s, ids := assemble(clock60(), periodPart(tt.x.grid(4)), periodPart(tt.y.grid(4)))
			x, y := ids[0], ids[1]

			got, err := s.BarlinePeriod(x, y)

			want := periodOracle(tt.x.length, tt.y.length)
			if tt.fail {
				if periodInRange(want) || !errorsAre(err, score.ErrOutOfRange, fraction.ErrOutOfRange) {
					b.Fatalf("the class %s expects score.ErrOutOfRange, got %v", tt.name, err)
				}
			} else if err != nil || !got.Equal(periodDuration(want)) {
				b.Fatalf("the class %s expects %s, got %s, %v", tt.name, want.RatString(), got, err)
			}

			b.ReportAllocs()

			for b.Loop() {
				_, _ = s.BarlinePeriod(x, y)
			}
		})
	}
}

// errorsAre reports whether err wraps every target.
func errorsAre(err error, targets ...error) bool {
	for _, target := range targets {
		if !errors.Is(err, target) {
			return false
		}
	}

	return err != nil
}
