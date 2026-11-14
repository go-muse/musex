package meter_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wS   = meter.WeightStrong
	wR   = meter.WeightRelativelyStrong
	wW   = meter.WeightWeak
	wOff = meter.WeightOffUnit
)

// TestStrongBeats covers О-10, ИР-8, П-7, Т-7, КП-14, КП-43, КП-50.
func TestStrongBeats(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, text string
		relative   []duration.Position
	}{
		{"КП-14 4/4", "4/4", poss("1/2")},
		{"КП-14 6/8", "6/8", poss("3/8")},
		{"КП-14 3/4", "3/4", nil},
		{"КП-14 5/8[2+3]", "5/8[2+3]", poss("1/4")},
		{"КП-14 5/8[3+2]", "5/8[3+2]", poss("3/8")},
		{"КП-14 7/8[3+2+2]", "7/8[3+2+2]", poss("3/8", "5/8")},
		{"КП-14 12/8", "12/8", poss("3/8", "3/4", "9/8")},
		{"КП-14 2/4+3/8", "2/4+3/8", poss("1/2")},
		{"КП-43 5/8[5]", "5/8[5]", nil},
		{"КП-50 4/3", "4/3", poss("2/3")},
		{"C|", "C|", nil},
		{"C[3+1]", "C[3+1]", poss("3/4")},
		{"6/8[2+2+2]", "6/8[2+2+2]", poss("1/4", "1/2")},
		{"(2+3)/8+2/4", "(2+3)/8+2/4", poss("1/4", "5/8")},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			strong, relative, err := mt(c.text).StrongBeats()
			require.NoError(t, err)
			assert.Equal(t, duration.Origin(), strong, "the strong beat is at 0")
			assert.Equal(t, c.relative, relative)
		})
	}
}

// TestUnitWeights covers О-10, ИР-8, П-7, Т-7, КП-15.
func TestUnitWeights(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, text string
		weights    []meter.Weight
	}{
		{"КП-15 4/4", "4/4", []meter.Weight{wS, wW, wR, wW}},
		{"КП-15 6/8", "6/8", []meter.Weight{wS, wW, wW, wR, wW, wW}},
		{"3/4", "3/4", []meter.Weight{wS, wW, wW}},
		{"7/8[3+2+2]", "7/8[3+2+2]", []meter.Weight{wS, wW, wW, wR, wW, wR, wW}},
		{"5/8[2+3]", "5/8[2+3]", []meter.Weight{wS, wW, wR, wW, wW}},
		{"1/4", "1/4", []meter.Weight{wS}},
		{"C[3+1]", "C[3+1]", []meter.Weight{wS, wW, wW, wR}},
		{"2/4+3/8", "2/4+3/8", []meter.Weight{wS, wW, wR, wW, wW}},
		{"2/4[1+1]", "2/4[1+1]", []meter.Weight{wS, wR}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := mt(c.text).UnitWeights()
			require.NoError(t, err)
			assert.Equal(t, c.weights, got)
		})
	}
}

// TestWeightAt covers О-11, ИР-8, Т-7, Т-8, КП-16, КП-34, КП-38, КП-46 and the
// order of checks of WeightAt: E-1 → E-2 → E-3 → E-7, without E-10.
func TestWeightAt(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name, text string
		t          duration.Position
		want       meter.Weight
	}{
		{"4/4 at 0", "4/4", pos(0, 1), wS},
		{"4/4 at 1/4", "4/4", pos(1, 4), wW},
		{"КП-26а 4/4 at 1/2", "4/4", pos(1, 2), wR},
		{"КП-16 4/4 at 3/4", "4/4", pos(3, 4), wW},
		{"4/4 at 1/8", "4/4", pos(1, 8), wOff},
		{"КП-16 6/8 at 3/8", "6/8", pos(3, 8), wR},
		{"КП-16 6/8 at 1/2", "6/8", pos(1, 2), wW},
		{"КП-16 6/8 at 7/16", "6/8", pos(7, 16), wOff},
		{"2/4+3/8 at 1/2", "2/4+3/8", pos(1, 2), wR},
		{"2/4+3/8 at 5/8", "2/4+3/8", pos(5, 8), wW},
		{"2/4+3/8 at 3/8", "2/4+3/8", pos(3, 8), wOff},
		{"no E-10 off a unit", "3/4", rangeEdge(), wOff},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := mt(c.text).WeightAt(c.t)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}

	var unset duration.Position

	for _, c := range []struct {
		name, text string
		t          duration.Position
		want       error
		causes     []error
	}{
		{"КП-34 at the size", "3/4", pos(3, 4), meter.ErrOutsideMeter, nil},
		{"КП-34 after the size", "3/4", pos(1, 1), meter.ErrOutsideMeter, nil},
		{"КП-38 unset position", "3/4", unset, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"КП-30 without grouping", "5/8", pos(0, 1), meter.ErrGroupingRequired, nil},
		{"order: E-2 before E-3", "5/8", unset, meter.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"order: E-3 before E-7", "7/8", pos(1, 1), meter.ErrGroupingRequired, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := mt(c.text).WeightAt(c.t)
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, meter.Weight(0), got)
		})
	}
}

// TestAccentConsistency covers ИР-8 on the corpus: the strong beat and the
// relatively strong beats are the starts of the groups, UnitWeights agrees
// with the groups, and WeightAt at the start of unit k is UnitWeights()[k−1]
// and off-unit between starts.
func TestAccentConsistency(t *testing.T) {
	t.Parallel()

	for _, m := range corpus() {
		if !noErr(m.HasGrouping()) {
			continue
		}

		t.Run(m.String(), func(t *testing.T) {
			t.Parallel()

			groups := noErr(m.Groups())
			strong, relative := noErr2(m.StrongBeats())

			assert.Equal(t, groups[0].Start, strong)

			var starts []duration.Position
			for _, g := range groups[1:] {
				starts = append(starts, g.Start)
			}

			assert.Equal(t, starts, relative)

			weights := noErr(m.UnitWeights())
			units := noErr(m.Units())
			require.Len(t, weights, len(units))

			groupStart := map[duration.Position]bool{}
			for _, g := range groups {
				groupStart[g.Start] = true
			}

			step := max(1, len(units)/40)

			for i := 0; i < len(units); i += step {
				u := units[i]

				want := wW

				switch {
				case i == 0:
					want = wS
				case groupStart[u.Start]:
					want = wR
				}

				assert.Equal(t, want, weights[i], "unit %d", i+1)
				assert.Equal(t, weights[i], noErr(m.WeightAt(u.Start)), "WeightAt at unit %d", i+1)

				mid := u.Start.MustAdd(u.Duration.MustDivInt(2))
				assert.Equal(t, wOff, noErr(m.WeightAt(mid)), "WeightAt inside unit %d", i+1)
			}
		})
	}
}

// noErr2 returns the two values of a successful call and panics with the
// error of a failed one.
func noErr2[T, U any](v T, w U, err error) (T, U) {
	if err != nil {
		panic(err)
	}

	return v, w
}
