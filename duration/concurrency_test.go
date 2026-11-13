package duration_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentUse covers architecture 3.2 and the non-functional
// requirement of concurrent reading: operations of all types on shared
// values run concurrently without synchronization; run with -race. Each
// goroutine unmarshals only into its own variables.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	d := dur(5, 16)
	w := wr(duration.Quarter, 1, tuplet.Triplet())
	tt := tie(wr(duration.Half, 0), wr(duration.Quarter, 0))
	p := pos(7, 24)
	chain := []tuplet.Tuplet{tuplet.Triplet()}
	step := frac(1, 8)

	const workers = 8

	type observed struct {
		notate    []duration.Written
		written   duration.Written
		quantized duration.Position
		text      string
		parsed    duration.Tied
		chain     []tuplet.Tuplet
		sum       duration.Duration
		base      duration.Base
	}

	results := make([]observed, workers)

	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			for range 50 {
				var o observed

				o.notate = d.MustNotate()
				o.written = dur(1, 4).MustToWritten(chain...)
				o.quantized = p.MustQuantize(step, rounding.HalfEven)
				o.text = w.String() + tt.String() + d.String() + p.String()
				o.chain = w.MustChain()
				o.sum = duration.MustSum(d, w.MustDuration(), tt.MustDuration())
				o.base = duration.Bases()[i%14]

				_, _, _ = w.Subdivide()
				_, _ = tt.Append(w)
				_, _ = tt.ToWritten()
				_, _, _ = duration.QuantizeSpan(p, d, step, rounding.Up)
				_, _ = d.NotateWith(duration.Whole, 2, chain...)
				_ = w.Compare(w.MustAddDot())
				_, _ = duration.ParseBase("crotchet")

				var parsed duration.Tied
				if err := parsed.UnmarshalText([]byte("2~4")); err == nil {
					o.parsed = parsed
				}

				results[i] = o
			}
		})
	}

	wg.Wait()

	for i, o := range results {
		assert.Equal(t, []duration.Written{wr(duration.Quarter, 0), wr(duration.Sixteenth, 0)}, o.notate)
		assert.Equal(t, wr(duration.Quarter, 1, tuplet.Triplet()), o.written)
		assert.Equal(t, pos(1, 4), o.quantized)
		assert.Equal(t, "4.[3:2]2~45/16"+"7/24", o.text)
		assert.Equal(t, tt, o.parsed)
		assert.Equal(t, []tuplet.Tuplet{tuplet.Triplet()}, o.chain)
		assert.Equal(t, dur(5, 16).MustAdd(dur(1, 4)).MustAdd(dur(3, 4)), o.sum)
		assert.Equal(t, baseTable[i%14].base, o.base)
	}
}
