package meter_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentReads covers the non-functional requirement of concurrent
// reading (architecture 3.2): shared values are read by every read
// operation from several goroutines without synchronization. The test is
// meaningful under go test -race.
func TestConcurrentReads(t *testing.T) {
	t.Parallel()

	shared := mts("7/8[3+2+2]", "2/4+3/8", "C", "5/8", "(3+2+3)/8")
	want := make([]string, len(shared))

	for i, m := range shared {
		want[i] = readAll(m)
	}

	const workers = 8

	results := make([][]string, workers)

	var wg sync.WaitGroup

	for w := range workers {
		wg.Go(func() {
			out := make([]string, len(shared))
			for i, m := range shared {
				out[i] = readAll(m)
			}

			results[w] = out
		})
	}

	wg.Wait()

	for w := range workers {
		assert.Equal(t, want, results[w], "worker %d", w)
	}
}

// readAll calls every read operation of m and of the package functions and
// describes the outcomes.
func readAll(m meter.Meter) string {
	p := pos(3, 8)
	v := dur(1, 8)
	other := meter.SixEight()

	var out []string

	add := func(v any, err error) { out = append(out, describe(v, err)) }
	add2 := func(v, w any, err error) { out = append(out, describe([]any{v, w}, err)) }

	out = append(out, m.String(), describe(m.IsSet(), nil), describe(m.Equal(other), nil), describe(m.Compare(other), nil))
	add(m.Form())
	add2(m.Symbol())
	add(m.PartCount())
	add(m.Parts())
	add(m.Addends())
	add(m.Numerator())
	add(m.Denominator())
	add(m.HasCommonDenominator())
	add(m.HasPowerOfTwoDenominators())
	add(m.DenominatorExponent())
	add(m.Duration())
	add2(m.QuoRem(v))
	add(m.UnitCount())
	add(m.Unit())
	add(m.UnitWritten())
	add(m.Units())
	add(m.UnitStart(2))
	add2(m.UnitAt(p))
	add(m.HasGrouping())
	add(m.HasExplicitGrouping())
	add(m.Grouping())
	add(m.GroupCount())
	add(m.Groups())
	add(m.GroupsWritten())
	add(m.GroupStart(2))
	add2(m.GroupAt(p))
	add(m.WithGrouping(1, 1, 1, 1, 1, 1, 1))
	add(m.ResetGrouping())
	add(m.GroupingVariants())
	add(m.BeatCount())
	add(m.Beats())
	add(m.BeatsWritten())
	add(m.BeatStart(2))
	add2(m.BeatAt(p))
	add(m.Kind())
	add(m.RussianKind())
	add2(m.BeatDivision())
	add2(m.KindName())
	add2(m.KindRussianName())
	add2(m.StrongBeats())
	add(m.UnitWeights())
	add(m.WeightAt(p))
	add(m.ToNumeric())
	add(m.ToSymbol())
	add(m.ToSimple())
	add(m.ToAdditive())
	add(m.ToSum())
	add(m.Shorter())
	add(m.Longer())
	add(m.WithDenominator(16))
	add(m.EqualStructure(other))
	add(m.EqualDuration(other))
	add(m.CompareDuration(other))
	add(m.MarshalText())
	add(m.AppendText(nil))
	add(meter.GroupingVariants(9))
	add(meter.Parse(m.String()))

	k, _ := m.Kind()
	add2(k.Name())

	return describe(out, nil)
}
