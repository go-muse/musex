package ticks_test

import (
	"errors"
	"math/big"
	"strconv"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/format/ticks"
)

// BenchmarkMinResolution measures ticks.MinResolution(ds, ps, q...) on the
// classes of architecture 6.2 of task 26.8, named <class>/<n> for n values.
// Before the loop each case is checked once against the oracle of 4.5: a
// success with the value of the oracle, or Е-8 for out-of-range.
func BenchmarkMinResolution(b *testing.B) {
	k := minresK()

	// small: ordinary durations, triplets, and positions, alternately.
	smallDurations := []*big.Rat{
		big.NewRat(1, 4), big.NewRat(1, 8), big.NewRat(1, 12), big.NewRat(3, 16),
		big.NewRat(1, 6), big.NewRat(3, 8), big.NewRat(1, 24), big.NewRat(1, 2),
	}
	smallPositions := []*big.Rat{
		big.NewRat(3, 8), big.NewRat(5, 12), new(big.Rat), big.NewRat(7, 16),
		big.NewRat(1, 3), big.NewRat(9, 4), big.NewRat(11, 24), big.NewRat(2, 1),
	}

	small := func(n int) (ds, ps []*big.Rat) {
		for i := range n {
			if i%2 == 0 {
				ds = append(ds, smallDurations[(i/2)%len(smallDurations)])
			} else {
				ps = append(ps, smallPositions[(i/2)%len(smallPositions)])
			}
		}

		return ds, ps
	}

	// wide: (K − 2j)/2⁶⁷ with bb = 1, D₀ = 2⁶² (6.2).
	wide := func(n int) (ds, ps []*big.Rat) {
		for j := range n {
			ds = append(ds, minresRat(new(big.Int).Sub(k, big.NewInt(int64(2*j))), minresPow2(67)))
		}

		return ds, nil
	}

	// out-of-range: the small values and 1/2⁶⁵ last; D₀ ≥ 2⁶³ with bb = 8.
	outOfRange := func(n int) (ds, ps []*big.Rat) {
		ds, ps = small(n - 1)

		return append(ds, minresInv(minresPow2(65))), ps
	}

	classes := []struct {
		name   string
		bb     int
		values func(n int) (ds, ps []*big.Rat)
		fails  bool
	}{
		{"small", 8, small, false},
		{"small-q255", 255, small, false},
		{"wide", 1, wide, false},
		{"out-of-range", 8, outOfRange, true},
	}

	for _, class := range classes {
		b.Run(class.name, func(b *testing.B) {
			for _, n := range []int{1, 8, 64} {
				b.Run(strconv.Itoa(n), func(b *testing.B) {
					dv, pv := class.values(n)
					if len(dv)+len(pv) != n {
						b.Fatalf("%d values, want %d", len(dv)+len(pv), n)
					}

					want, ok := minresOracle(class.bb, dv, pv)
					if ok == class.fails {
						b.Fatalf("the oracle gives D₀ = %s for bb = %d", want, class.bb)
					}

					ds, ps := minresDurations(dv), minresPositions(pv)
					q := minresQuarter(class.bb, false)

					r, err := ticks.MinResolution(ds, ps, q...)
					refused := errors.Is(err, ticks.ErrOutOfRange) && !errors.Is(err, fraction.ErrOutOfRange)

					switch {
					case class.fails && !refused:
						b.Fatalf("MinResolution: %v, want Е-8 without fraction.ErrOutOfRange", err)
					case !class.fails && (err != nil || int64(r.MustPerQuarter()) != want.Int64()):
						b.Fatalf("MinResolution: %v, %v, want D₀ = %s", r, err, want)
					}

					b.ReportAllocs()

					for b.Loop() {
						_, _ = ticks.MinResolution(ds, ps, q...)
					}
				})
			}
		})
	}
}
