package measure_test

import (
	"errors"
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
)

// Group Т-9, errors (architecture 4.2; 2.6; Т-15; КП-39…КП-48). The
// failures of each operation are also covered by its group; this group
// checks the sentinels and the order of checks on inputs with two
// violations of neighboring links of the order.

// TestErrorsSentinels covers 2.6 and architecture 4.1: the ten sentinels
// are set, pairwise distinct, and none wraps another or a sentinel of a
// dependency.
func TestErrorsSentinels(t *testing.T) {
	t.Parallel()

	for i, a := range measureSentinels {
		assert.Error(t, a.err, a.name)

		for j, b := range measureSentinels {
			assert.Equal(t, i == j, errors.Is(a.err, b.err), "errors.Is(%s, %s)", a.name, b.name)
		}

		for _, d := range dependencySentinels {
			assert.False(t, errors.Is(a.err, d.err), "errors.Is(%s, %s)", a.name, d.name)
		}
	}
}

// TestErrorsOrder covers the order of checks of architecture 4.2: an input
// with two violations returns the earlier one; a dependency sentinel is
// wrapped only when the earlier violation is its failure.
func TestErrorsOrder(t *testing.T) {
	t.Parallel()

	hugeLengths := func(t *testing.T) measure.Grid {
		huge := durRat(ratInt(pow2(254)))

		return build(t, mt("4/4"),
			func(g measure.Grid) (measure.Grid, error) { return g.WithLength(1, huge, measure.AlignStart) },
			func(g measure.Grid) (measure.Grid, error) { return g.WithLength(2, huge, measure.AlignStart) },
			withBarCount(2))
	}

	tests := []struct {
		name    string
		call    func(t *testing.T) error
		err     error
		causes  []error
		without []error
	}{
		{"unset grid and a negative bar", func(t *testing.T) error { return e(measure.Grid{}.Bar(-1)) }, measure.ErrUnset, nil, nil},
		{"unset grid and an unset address", func(t *testing.T) error { return e(measure.Grid{}.PositionOf(measure.Address{})) }, measure.ErrUnset, nil, nil},
		{"unset grid and an unset meter", func(t *testing.T) error { return e(measure.Grid{}.WithMark(3, meter.Meter{})) }, measure.ErrUnset, nil, []error{meter.ErrUnset}},
		{"WithMark: bar < 1 and an unset meter", func(t *testing.T) error {
			return e(gridOf(t, "4/4").WithMark(0, meter.Meter{}))
		}, measure.ErrInvalidArgument, nil, []error{meter.ErrUnset}},
		{"WithMark: unset meter and after N", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").WithMark(5, meter.Meter{}))
		}, measure.ErrInvalidArgument, []error{meter.ErrUnset}, nil},
		{"WithLength: bar < 1 and an unset length", func(t *testing.T) error {
			return e(gridOf(t, "4/4").WithLength(0, duration.Duration{}, measure.AlignStart))
		}, measure.ErrInvalidArgument, nil, []error{duration.ErrUnset}},
		{"WithLength: unset length and an invalid alignment", func(t *testing.T) error {
			return e(gridOf(t, "4/4").WithLength(3, duration.Duration{}, 0))
		}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"WithLength: invalid alignment and after N", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").WithLength(6, dur("1/2"), 0))
		}, measure.ErrInvalidArgument, nil, nil},
		{"WithLength: after N and longer to the end", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").WithLength(6, dur("5/4"), measure.AlignEnd))
		}, measure.ErrOutsideGrid, nil, nil},
		{"WithMetersIn: first > last and an unset meter", func(t *testing.T) error {
			return e(gridOf(t, "4/4").WithMetersIn(3, 2, meter.Meter{}))
		}, measure.ErrInvalidArgument, nil, []error{meter.ErrUnset}},
		{"WithMetersIn: unset meter and after N", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").WithMetersIn(5, 6, meter.Meter{}))
		}, measure.ErrInvalidArgument, []error{meter.ErrUnset}, nil},
		{"WithMetersIn: after N and too many", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").WithMetersIn(1, 2000, mt("3/4"), mt("2/4")))
		}, measure.ErrOutsideGrid, nil, nil},
		{"WithMetersIn: too many and longer", func(t *testing.T) error {
			return e(gridOf(t, "4/4;pickup:3/4").WithMetersIn(1, 2000, mt("2/4"), mt("2/4")))
		}, measure.ErrTooMany, nil, nil},
		{"InsertBars: after N + 1 and a number past math.MaxInt", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").InsertBars(6, math.MaxInt))
		}, measure.ErrOutsideGrid, nil, nil},
		{"DeleteBars: all bars and a longer pickup", func(t *testing.T) error {
			return e(gridOf(t, "4/4;pickup:3/4;bars:4").DeleteBars(1, 4))
		}, measure.ErrInvalidArgument, nil, nil},
		{"AddBars: m < 1 of an open grid", func(t *testing.T) error { return e(gridOf(t, "4/4").AddBars(0)) }, measure.ErrInvalidArgument, nil, nil},
		{"AddBars: open grid and a number past math.MaxInt", func(t *testing.T) error {
			return e(gridOf(t, "4/4").AddBars(math.MaxInt))
		}, measure.ErrUndefined, nil, nil},
		{"Append: unset argument of an open receiver", func(t *testing.T) error {
			return e(gridOf(t, "4/4").Append(measure.Grid{}))
		}, measure.ErrUnset, nil, nil},
		{"Append: open receiver and a number past math.MaxInt", func(t *testing.T) error {
			return e(gridOf(t, "4/4").Append(gridOf(t, "4/4;9223372036854775807:3/4")))
		}, measure.ErrUndefined, nil, nil},
		{"Slice: last < 1 and bar 0 missing", func(t *testing.T) error { return e(gridOf(t, "4/4").Slice(0, 0)) }, measure.ErrInvalidArgument, nil, nil},
		{"Slice: missing first and missing last", func(t *testing.T) error { return e(gridOf(t, "4/4;bars:4").Slice(5, 6)) }, measure.ErrOutsideGrid, nil, nil},
		{"PartOfBar: unset value and a missing bar", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").PartOfBar(9, duration.Duration{}))
		}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"CountAt: unset t and a missing bar", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").CountAt(9, duration.Position{}))
		}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"TstampOffset: unset ts and a missing bar", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").TstampOffset(9, fraction.Fraction{}))
		}, measure.ErrInvalidArgument, []error{fraction.ErrInvalid}, nil},
		{"MetricPlace: missing bar and an offset after the length", func(t *testing.T) error {
			return e2(gridOf(t, "4/4;bars:4").MetricPlace(ad(5, "9/8")))
		}, measure.ErrOutsideGrid, nil, nil},
		{"MetricPlace: offset after the length of an extended bar", func(t *testing.T) error {
			return e2(gridOf(t, "4/4;2:length:5/4").MetricPlace(ad(2, "3/2")))
		}, measure.ErrOutsideBar, nil, nil},
		{"Elapsed: missing bar and an offset after the length", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").Elapsed(ad(5, "9")))
		}, measure.ErrOutsideGrid, nil, nil},
		{"BarStart: missing bar of huge lengths", func(t *testing.T) error { return e(hugeLengths(t).BarStart(5)) }, measure.ErrOutsideGrid, nil, nil},
		{"Bars: missing last of huge lengths", func(t *testing.T) error { return e(hugeLengths(t).Bars(1, 5)) }, measure.ErrOutsideGrid, nil, nil},
		{"RangeLength: first > last and missing bars", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").RangeLength(9, 8))
		}, measure.ErrInvalidArgument, nil, nil},
		{"Locate: unset position of a finite grid", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").Locate(duration.Position{}))
		}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"Normalize: missing bar and an offset past the end", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:4").Normalize(ad(0, "9")))
		}, measure.ErrOutsideGrid, nil, nil},
		{"SplitBy: no grouping and more than MaxParts parts", func(t *testing.T) error {
			return e(gridOf(t, "7/8").SplitBy(pos("0"), dur("2000"), measure.LevelBeats))
		}, measure.ErrGroupingRequired, []error{meter.ErrGroupingRequired}, nil},
		{"SplitBy: unset value and an invalid level", func(t *testing.T) error {
			return e(gridOf(t, "4/4").SplitBy(pos("0"), duration.Duration{}, 0))
		}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"SplitBy: second bar without grouping", func(t *testing.T) error {
			return e(gridOf(t, "4/4;2:5/8").SplitBy(pos("1/2"), dur("1"), measure.LevelGroups))
		}, measure.ErrGroupingRequired, []error{meter.ErrGroupingRequired}, nil},
		{"Complement: unset length", func(t *testing.T) error {
			return e2(measure.Complement(mt("3/4"), duration.Duration{}))
		}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"MeterCountOffset: unset meter and a count below 1", func(t *testing.T) error {
			return e(measure.MeterCountOffset(meter.Meter{}, frac("0")))
		}, measure.ErrInvalidArgument, []error{meter.ErrUnset}, nil},
		{"MeterTstampOffset: no common denominator and a tstamp below 1", func(t *testing.T) error {
			return e(measure.MeterTstampOffset(mt("2/4+3/8"), frac("1/2")))
		}, measure.ErrUndefined, []error{meter.ErrNoCommonDenominator}, nil},
		{"NewAddress: negative bar and an unset offset", func(t *testing.T) error {
			return e(measure.NewAddress(-1, duration.Position{}))
		}, measure.ErrInvalidArgument, nil, []error{duration.ErrUnset}},
		{"Ordinal: missing bar of a grid with pickup", func(t *testing.T) error {
			return e(gridOf(t, "4/4;pickup:1/4;bars:4").Ordinal(5))
		}, measure.ErrOutsideGrid, nil, nil},
		{"BarAtOrdinal: n < 1 and missing", func(t *testing.T) error { return e(gridOf(t, "4/4;bars:4").BarAtOrdinal(-9)) }, measure.ErrInvalidArgument, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			err := tt.call(t)
			assertErr(t, err, tt.err, tt.causes...)
			assertNoCause(t, err, tt.without...)
		})
	}
}
