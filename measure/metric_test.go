package measure_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-5: the place in the meter, the unit count, MEI tstamp, the
// functions of one meter, and the closing length (architecture 4.9, 4.10;
// О-14…О-16, О-18; ИС-8…ИС-10; П-6, П-7; КП-7, КП-17…КП-25, КП-40,
// КП-42, КП-43, КП-58).

// TestMetricPlace covers О-14, ИС-8, П-6, and КП-25: the meter of the bar
// and the offset from the start of a full bar of the meter; package meter
// gives the unit and the weight at that place.
func TestMetricPlace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		address measure.Address
		meter   string
		place   string
		weight  meter.Weight
	}{
		{"КП-25 pickup (0, 0)", "3/4;pickup:1/4", ad(0, "0"), "3/4", "1/2", meter.WeightWeak},
		{"КП-25 (1, 0)", "3/4;pickup:1/4", ad(1, "0"), "3/4", "0", meter.WeightStrong},
		{"КП-25 (1, 1/2)", "3/4;pickup:1/4", ad(1, "1/2"), "3/4", "1/2", meter.WeightWeak},
		{"КП-25 to the end (5, 0)", "4/4;5:length:1/4:end", ad(5, "0"), "4/4", "3/4", meter.WeightWeak},
		{"КП-25 from the start (5, 0)", "4/4;5:length:1/4", ad(5, "0"), "4/4", "0", meter.WeightStrong},
		{"4/4 pickup 1/4 (0, 1/8)", "4/4;pickup:1/4", ad(0, "1/8"), "4/4", "7/8", meter.WeightOffUnit},
		{"whole-bar pickup (0, 1/2)", "4/4;pickup:1/1", ad(0, "1/2"), "4/4", "1/2", meter.WeightRelativelyStrong},
		{"extended before the size", "4/4;2:length:5/4", ad(2, "3/4"), "4/4", "3/4", meter.WeightWeak},
		{"after a change", "4/4;3:6/8", ad(4, "3/8"), "6/8", "3/8", meter.WeightRelativelyStrong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, u, err := gridOf(t, tt.grid).MetricPlace(tt.address)
			require.NoError(t, err)
			assertText(t, tt.meter, m)
			assertText(t, canonicalRat(tt.place), u)

			w, err := m.WeightAt(u)
			require.NoError(t, err)
			assert.Equal(t, tt.weight, w)
		})
	}
}

// TestMetricPlaceErrors covers architecture 4.2 and КП-41…КП-43 for
// MetricPlace: Е-4, then Е-5 for an offset at or after the length of the
// bar, then Е-6 for a place at or after the size of the meter.
func TestMetricPlaceErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		address measure.Address
		err     error
	}{
		{"КП-42 4/4 (1, 1)", "4/4", ad(1, "1"), measure.ErrOutsideBar},
		{"КП-43 extended (2, 9/8)", "4/4;2:length:5/4", ad(2, "9/8"), measure.ErrUndefined},
		{"extended at the size", "4/4;2:length:5/4", ad(2, "1"), measure.ErrUndefined},
		{"extended at its length", "4/4;2:length:5/4", ad(2, "5/4"), measure.ErrOutsideBar},
		{"bar 0 without pickup", "4/4", ad(0, "0"), measure.ErrOutsideGrid},
		{"after N", "4/4;bars:4", ad(5, "0"), measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, u, err := gridOf(t, tt.grid).MetricPlace(tt.address)
			assertErr(t, err, tt.err)
			assert.False(t, m.IsSet())
			assert.False(t, u.IsSet())
		})
	}
}

// countCase is a unit count of a bar at an offset by both ways of
// architecture 4.10.
type countCase struct {
	name   string
	grid   string
	bar    int
	t      string
	count  string // from the start of the bar
	metric string // by the place in the meter
}

// countCases are КП-17…КП-21 and КП-58.
func countCases() []countCase {
	return []countCase{
		{"КП-17 4/4 1/8", "4/4", 1, "1/8", "3/2", "3/2"},
		{"КП-17 6/8 3/8", "6/8", 1, "3/8", "4", "4"},
		{"КП-17 3/4 0", "3/4", 1, "0", "1", "1"},
		{"КП-17 4/4 1 right bar line", "4/4", 1, "1", "5", "5"},
		{"КП-18 pickup 0", "4/4;pickup:1/4", 0, "0", "1", "4"},
		{"КП-18 pickup 1/8", "4/4;pickup:1/4", 0, "1/8", "3/2", "9/2"},
		{"КП-18 pickup 1/4", "4/4;pickup:1/4", 0, "1/4", "2", "5"},
		{"КП-18 3/4 pickup 1/4", "3/4;pickup:1/4", 0, "0", "1", "3"},
		{"КП-18 6/8 pickup 1/8", "6/8;pickup:1/8", 0, "0", "1", "6"},
		{"КП-19 0", "2/4+3/8", 1, "0", "1", "1"},
		{"КП-19 1/8", "2/4+3/8", 1, "1/8", "3/2", "3/2"},
		{"КП-19 1/4", "2/4+3/8", 1, "1/4", "2", "2"},
		{"КП-19 1/2", "2/4+3/8", 1, "1/2", "3", "3"},
		{"КП-19 9/16", "2/4+3/8", 1, "9/16", "7/2", "7/2"},
		{"КП-19 3/4", "2/4+3/8", 1, "3/4", "5", "5"},
		{"КП-19 7/8", "2/4+3/8", 1, "7/8", "6", "6"},
		{"КП-20 extended 1", "4/4;2:length:5/4", 2, "1", "5", "5"},
		{"КП-20 extended 9/8", "4/4;2:length:5/4", 2, "9/8", "11/2", "11/2"},
		{"КП-20 extended 5/4", "4/4;2:length:5/4", 2, "5/4", "6", "6"},
		{"КП-21 to the end", "4/4;5:length:1/4:end", 5, "0", "1", "4"},
		{"КП-58 4/3", "4/3", 1, "1/3", "2", "2"},
		{"incomplete from the start", "4/4;3:length:1/2", 3, "1/2", "3", "3"},
		{"after a pickup bar 1", "4/4;pickup:1/4", 1, "1/4", "2", "2"},
		{"pickup of a whole bar", "4/4;pickup:1/1", 0, "1/2", "3", "3"},
	}
}

// TestCountAt covers О-15, ИС-9, П-7, and КП-17…КП-21, КП-58: the unit
// count from the start of the bar (CountAt) and by the place in the meter
// (MetricCountAt), and the inverse offsets.
func TestCountAt(t *testing.T) {
	t.Parallel()

	for _, tt := range countCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			c, err := g.CountAt(tt.bar, pos(tt.t))
			require.NoError(t, err)
			assertFrac(t, tt.count, c, "CountAt")

			mc, err := g.MetricCountAt(tt.bar, pos(tt.t))
			require.NoError(t, err)
			assertFrac(t, tt.metric, mc, "MetricCountAt")

			off, err := g.CountOffset(tt.bar, frac(tt.count))
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.t), off, "CountOffset")

			moff, err := g.MetricCountOffset(tt.bar, frac(tt.metric))
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.t), moff, "MetricCountOffset")
		})
	}
}

// TestCountOffset covers КП-22: the offset by a count from the start of the
// bar and by the place in the meter.
func TestCountOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		bar    int
		count  string
		metric bool
		want   string
	}{
		{"КП-22 4/4 3/2", "4/4", 1, "3/2", false, "1/8"},
		{"КП-22 6/8 4", "6/8", 1, "4", false, "3/8"},
		{"КП-22 4/4 5", "4/4", 1, "5", false, "1"},
		{"КП-22 2/4+3/8 7/2", "2/4+3/8", 1, "7/2", false, "9/16"},
		{"КП-22 pickup by place 4", "4/4;pickup:1/4", 0, "4", true, "0"},
		{"КП-22 pickup by place 9/2", "4/4;pickup:1/4", 0, "9/2", true, "1/8"},
		{"КП-22 pickup by place 5", "4/4;pickup:1/4", 0, "5", true, "1/4"},
		{"unnormalized fraction", "4/4", 1, "6/4", false, "1/8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			c := frac(tt.count)
			if tt.name == "unnormalized fraction" {
				c = fraction.MustNew(6, 4)
			}

			var (
				off duration.Position
				err error
			)

			if tt.metric {
				off, err = g.MetricCountOffset(tt.bar, c)
			} else {
				off, err = g.CountOffset(tt.bar, c)
			}

			require.NoError(t, err)
			assertText(t, canonicalRat(tt.want), off)
		})
	}
}

// TestCountErrors covers architecture 4.2 and КП-39, КП-41, КП-42 for the
// unit count in a grid: Е-2 for a negative bar, then for an unset offset or
// count; Е-4; Е-5 for an offset after the length or a count below 1 or out
// of the bar.
func TestCountErrors(t *testing.T) {
	t.Parallel()

	type call func(measure.Grid) (fraction.Fraction, duration.Position, error)

	at := func(bar int, tv duration.Position) call {
		return func(g measure.Grid) (fraction.Fraction, duration.Position, error) {
			f, err := g.CountAt(bar, tv)

			return f, duration.Position{}, err
		}
	}
	metricAt := func(bar int, tv duration.Position) call {
		return func(g measure.Grid) (fraction.Fraction, duration.Position, error) {
			f, err := g.MetricCountAt(bar, tv)

			return f, duration.Position{}, err
		}
	}
	offset := func(bar int, c fraction.Fraction) call {
		return func(g measure.Grid) (fraction.Fraction, duration.Position, error) {
			p, err := g.CountOffset(bar, c)

			return fraction.Fraction{}, p, err
		}
	}
	metricOffset := func(bar int, c fraction.Fraction) call {
		return func(g measure.Grid) (fraction.Fraction, duration.Position, error) {
			p, err := g.MetricCountOffset(bar, c)

			return fraction.Fraction{}, p, err
		}
	}

	tests := []struct {
		name    string
		grid    string
		call    call
		err     error
		causes  []error
		without []error
	}{
		{"КП-42 CountAt 9/8", "4/4", at(1, pos("9/8")), measure.ErrOutsideBar, nil, nil},
		{"КП-42 MetricCountAt 9/8", "4/4", metricAt(1, pos("9/8")), measure.ErrOutsideBar, nil, nil},
		{"КП-42 CountOffset 6", "4/4", offset(1, frac("6")), measure.ErrOutsideBar, nil, nil},
		{"КП-42 CountOffset 1/2", "4/4", offset(1, frac("1/2")), measure.ErrOutsideBar, nil, nil},
		{"КП-42 CountOffset 0", "4/4", offset(1, frac("0")), measure.ErrOutsideBar, nil, nil},
		{"negative count", "4/4", offset(1, frac("-2")), measure.ErrOutsideBar, nil, nil},
		{"КП-42 pickup by place 3", "4/4;pickup:1/4", metricOffset(0, frac("3")), measure.ErrOutsideBar, nil, nil},
		{"pickup by place after the end", "4/4;pickup:1/4", metricOffset(0, frac("11/2")), measure.ErrOutsideBar, nil, nil},
		{"pickup from the start after the end", "4/4;pickup:1/4", offset(0, frac("3")), measure.ErrOutsideBar, nil, nil},
		{"incomplete after its end", "4/4;3:length:1/2", offset(3, frac("4")), measure.ErrOutsideBar, nil, nil},
		{"pickup CountAt after its length", "4/4;pickup:1/4", at(0, pos("3/8")), measure.ErrOutsideBar, nil, nil},
		{"КП-39 unset t", "4/4", at(1, duration.Position{}), measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"unset t metric", "4/4", metricAt(1, duration.Position{}), measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"unset c", "4/4", offset(1, fraction.Fraction{}), measure.ErrInvalidArgument, []error{fraction.ErrInvalid}, nil},
		{"unset c metric", "4/4", metricOffset(1, fraction.Fraction{}), measure.ErrInvalidArgument, []error{fraction.ErrInvalid}, nil},
		{"negative bar", "4/4", at(-1, pos("0")), measure.ErrInvalidArgument, nil, nil},
		{"negative bar before unset t", "4/4", at(-1, duration.Position{}), measure.ErrInvalidArgument, nil, []error{duration.ErrUnset}},
		{"negative bar before unset c", "4/4", offset(-1, fraction.Fraction{}), measure.ErrInvalidArgument, nil, []error{fraction.ErrInvalid}},
		{"bar 0 without pickup", "4/4", at(0, pos("0")), measure.ErrOutsideGrid, nil, nil},
		{"after N", "4/4;bars:4", metricOffset(5, frac("1")), measure.ErrOutsideGrid, nil, nil},
		{"missing bar before outside bar", "4/4;bars:4", at(5, pos("9")), measure.ErrOutsideGrid, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			f, p, err := tt.call(gridOf(t, tt.grid))
			assertErr(t, err, tt.err, tt.causes...)
			assertNoCause(t, err, tt.without...)
			assert.False(t, f.IsValid())
			assert.False(t, p.IsSet())
		})
	}
}

// TestCountMeter covers О-15 and КП-23, КП-42: the unit count of one meter
// without a grid and its inverse.
func TestCountMeter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		meter string
		u     string
		count string
	}{
		{"КП-23 4/4 1/8", "4/4", "1/8", "3/2"},
		{"КП-23 6/8 3/8", "6/8", "3/8", "4"},
		{"КП-23 4/3 1/3", "4/3", "1/3", "2"},
		{"КП-23 2/8+3/8 1/4", "2/8+3/8", "1/4", "3"},
		{"2/4+3/8 9/16", "2/4+3/8", "9/16", "7/2"},
		{"right bar line", "3/4", "3/4", "4"},
		{"start", "C", "0", "1"},
		{"5/8 without grouping", "5/8", "3/16", "5/2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			c, err := measure.MeterCountAt(mt(tt.meter), pos(tt.u))
			require.NoError(t, err)
			assertFrac(t, tt.count, c)

			u, err := measure.MeterCountOffset(mt(tt.meter), frac(tt.count))
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.u), u)
		})
	}
}

// TestCountMeterErrors covers architecture 4.2 and КП-42 for the functions
// of one meter.
func TestCountMeterErrors(t *testing.T) {
	t.Parallel()

	t.Run("MeterCountAt", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			meter   meter.Meter
			u       duration.Position
			err     error
			causes  []error
			without []error
		}{
			{"КП-42 9/8", mt("4/4"), pos("9/8"), measure.ErrOutsideBar, nil, nil},
			{"unset meter", meter.Meter{}, pos("0"), measure.ErrInvalidArgument, []error{meter.ErrUnset}, nil},
			{"unset u", mt("4/4"), duration.Position{}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
			{"unset meter before unset u", meter.Meter{}, duration.Position{}, measure.ErrInvalidArgument, []error{meter.ErrUnset}, []error{duration.ErrUnset}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				defer guard(t)

				c, err := measure.MeterCountAt(tt.meter, tt.u)
				assertErr(t, err, tt.err, tt.causes...)
				assertNoCause(t, err, tt.without...)
				assert.False(t, c.IsValid())
			})
		}
	})

	t.Run("MeterCountOffset", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			meter   meter.Meter
			c       fraction.Fraction
			err     error
			causes  []error
			without []error
		}{
			{"below 1", mt("4/4"), frac("1/2"), measure.ErrOutsideBar, nil, nil},
			{"after n + 1", mt("4/4"), frac("6"), measure.ErrOutsideBar, nil, nil},
			{"just after n + 1", mt("3/4"), frac("4001/1000"), measure.ErrOutsideBar, nil, nil},
			{"unset meter", meter.Meter{}, frac("1"), measure.ErrInvalidArgument, []error{meter.ErrUnset}, nil},
			{"unset c", mt("4/4"), fraction.Fraction{}, measure.ErrInvalidArgument, []error{fraction.ErrInvalid}, nil},
			{"unset meter before unset c", meter.Meter{}, fraction.Fraction{}, measure.ErrInvalidArgument, []error{meter.ErrUnset}, []error{fraction.ErrInvalid}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				defer guard(t)

				u, err := measure.MeterCountOffset(tt.meter, tt.c)
				assertErr(t, err, tt.err, tt.causes...)
				assertNoCause(t, err, tt.without...)
				assert.False(t, u.IsSet())
			})
		}
	})
}

// TestTstamp covers О-16, ИС-10, and КП-24, КП-58: MEI tstamp only in a full
// bar of a meter with a common denominator, and its inverse.
func TestTstamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		bar    int
		t      string
		tstamp string
	}{
		{"КП-24 4/4 1/8", "4/4", 1, "1/8", "3/2"},
		{"КП-24 4/4 1", "4/4", 1, "1", "5"},
		{"КП-24 6/8 3/8", "6/8", 1, "3/8", "4"},
		{"КП-24 whole-bar pickup", "4/4;pickup:1/1", 0, "1/8", "3/2"},
		{"КП-24 explicit length equal to the meter", "4/4;3:length:1/1", 3, "1/8", "3/2"},
		{"КП-58 4/3", "4/3", 1, "1/3", "2"},
		{"start", "3/4", 2, "0", "1"},
		{"additive", "(3+2+3)/8", 1, "5/8", "6"},
		{"after a change", "4/4;3:6/8", 3, "1/2", "5"},
		{"to the end whole", "4/4;5:length:1/1:end", 5, "1/4", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			ts, err := g.TstampAt(tt.bar, pos(tt.t))
			require.NoError(t, err)
			assertFrac(t, tt.tstamp, ts)

			off, err := g.TstampOffset(tt.bar, frac(tt.tstamp))
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.t), off)

			c, err := g.CountAt(tt.bar, pos(tt.t))
			require.NoError(t, err)
			assert.Equal(t, ts, c, "the unit count equals MEI tstamp in its domain")
		})
	}

	t.Run("КП-24 inverse 3/2", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		off, err := gridOf(t, "4/4").TstampOffset(1, frac("3/2"))
		require.NoError(t, err)
		assertText(t, "1/8", off)
	})
}

// TestTstampErrors covers architecture 4.2 and КП-42, КП-43 for TstampAt and
// TstampOffset: Е-4; Е-6 for an incomplete or extended bar, then Е-6 with
// meter.ErrNoCommonDenominator; Е-5.
func TestTstampErrors(t *testing.T) {
	t.Parallel()

	type call func(measure.Grid) error

	at := func(bar int, tv duration.Position) call {
		return func(g measure.Grid) error {
			f, err := g.TstampAt(bar, tv)
			if f.IsValid() {
				return nil
			}

			return err
		}
	}
	offset := func(bar int, ts fraction.Fraction) call {
		return func(g measure.Grid) error {
			p, err := g.TstampOffset(bar, ts)
			if p.IsSet() {
				return nil
			}

			return err
		}
	}

	tests := []struct {
		name    string
		grid    string
		call    call
		err     error
		causes  []error
		without []error
	}{
		{"КП-42 9/8", "4/4", at(1, pos("9/8")), measure.ErrOutsideBar, nil, nil},
		{"КП-42 inverse 0", "4/4", offset(1, frac("0")), measure.ErrOutsideBar, nil, nil},
		{"inverse after the bar", "4/4", offset(1, frac("11/2")), measure.ErrOutsideBar, nil, nil},
		{"КП-43 pickup 1/4", "4/4;pickup:1/4", at(0, pos("0")), measure.ErrUndefined, nil, []error{meter.ErrNoCommonDenominator}},
		{"КП-43 incomplete bar 3", "4/4;3:length:1/2", at(3, pos("0")), measure.ErrUndefined, nil, []error{meter.ErrNoCommonDenominator}},
		{"КП-43 extended bar 2", "4/4;2:length:5/4", at(2, pos("0")), measure.ErrUndefined, nil, []error{meter.ErrNoCommonDenominator}},
		{"КП-43 sum 2/4+3/8", "2/4+3/8", at(1, pos("0")), measure.ErrUndefined, []error{meter.ErrNoCommonDenominator}, nil},
		{"sum inverse", "2/4+3/8", offset(1, frac("1")), measure.ErrUndefined, []error{meter.ErrNoCommonDenominator}, nil},
		{"incomplete inverse", "4/4;3:length:1/2", offset(3, frac("1")), measure.ErrUndefined, nil, []error{meter.ErrNoCommonDenominator}},
		{"to the end incomplete", "4/4;5:length:1/4:end", at(5, pos("0")), measure.ErrUndefined, nil, nil},
		{"not full before no common denominator", "2/4+3/8;2:length:1/4", at(2, pos("0")), measure.ErrUndefined, nil, []error{meter.ErrNoCommonDenominator}},
		{"not full before t after the length", "4/4;3:length:1/2", at(3, pos("9/8")), measure.ErrUndefined, nil, nil},
		{"no common denominator before t after the length", "2/4+3/8", at(1, pos("2")), measure.ErrUndefined, []error{meter.ErrNoCommonDenominator}, nil},
		{"unset t", "4/4", at(1, duration.Position{}), measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"unset ts", "4/4", offset(1, fraction.Fraction{}), measure.ErrInvalidArgument, []error{fraction.ErrInvalid}, nil},
		{"negative bar", "4/4", at(-1, pos("0")), measure.ErrInvalidArgument, nil, nil},
		{"bar 0 without pickup", "4/4", at(0, pos("0")), measure.ErrOutsideGrid, nil, nil},
		{"missing bar before not full", "4/4;3:length:1/2;bars:3", at(4, pos("0")), measure.ErrOutsideGrid, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			err := tt.call(gridOf(t, tt.grid))
			assertErr(t, err, tt.err, tt.causes...)
			assertNoCause(t, err, tt.without...)
		})
	}
}

// TestTstampMeter covers О-16 and КП-24, КП-42, КП-43 for the MEI tstamp of
// one meter without a grid.
func TestTstampMeter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		meter  string
		u      string
		tstamp string
	}{
		{"КП-24 4/4 1/8", "4/4", "1/8", "3/2"},
		{"КП-24 6/8 3/8", "6/8", "3/8", "4"},
		{"4/3 1/3", "4/3", "1/3", "2"},
		{"right bar line", "4/4", "1", "5"},
		{"C", "C", "1/2", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			ts, err := measure.MeterTstampAt(mt(tt.meter), pos(tt.u))
			require.NoError(t, err)
			assertFrac(t, tt.tstamp, ts)

			u, err := measure.MeterTstampOffset(mt(tt.meter), frac(tt.tstamp))
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.u), u)
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		_, err := measure.MeterTstampAt(mt("4/4"), pos("9/8"))
		assertErr(t, err, measure.ErrOutsideBar)

		_, err = measure.MeterTstampAt(mt("2/4+3/8"), pos("0"))
		assertErr(t, err, measure.ErrUndefined, meter.ErrNoCommonDenominator)

		_, err = measure.MeterTstampAt(mt("2/4+3/8"), pos("2"))
		assertErr(t, err, measure.ErrUndefined, meter.ErrNoCommonDenominator)

		_, err = measure.MeterTstampAt(meter.Meter{}, pos("0"))
		assertErr(t, err, measure.ErrInvalidArgument, meter.ErrUnset)

		_, err = measure.MeterTstampAt(mt("4/4"), duration.Position{})
		assertErr(t, err, measure.ErrInvalidArgument, duration.ErrUnset)

		_, err = measure.MeterTstampAt(meter.Meter{}, duration.Position{})
		assertErr(t, err, measure.ErrInvalidArgument, meter.ErrUnset)
		assertNoCause(t, err, duration.ErrUnset)

		u, err := measure.MeterTstampOffset(mt("2/4+3/8"), frac("1"))
		assertErr(t, err, measure.ErrUndefined, meter.ErrNoCommonDenominator)
		assert.False(t, u.IsSet())

		_, err = measure.MeterTstampOffset(mt("4/4"), frac("0"))
		assertErr(t, err, measure.ErrOutsideBar)

		_, err = measure.MeterTstampOffset(mt("4/4"), frac("6"))
		assertErr(t, err, measure.ErrOutsideBar)

		_, err = measure.MeterTstampOffset(mt("4/4"), fraction.Fraction{})
		assertErr(t, err, measure.ErrInvalidArgument, fraction.ErrInvalid)

		_, err = measure.MeterTstampOffset(meter.Meter{}, frac("1"))
		assertErr(t, err, measure.ErrInvalidArgument, meter.ErrUnset)

		_, err = measure.MeterTstampOffset(mt("2/4+3/8"), frac("0"))
		assertErr(t, err, measure.ErrUndefined, meter.ErrNoCommonDenominator)
	})
}

// TestComplement covers О-18 and КП-7, КП-39, КП-40: the closing length of
// a pickup; a pickup of a whole bar gives "no incomplete closing bar".
func TestComplement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		meter   meter.Meter
		pickup  duration.Duration
		closing string
		ok      bool
		err     error
		causes  []error
		without []error
	}{
		{"КП-7 4/4 and 1/4", mt("4/4"), dur("1/4"), "3/4", true, nil, nil, nil},
		{"КП-7 3/4 and 1/8", mt("3/4"), dur("1/8"), "5/8", true, nil, nil, nil},
		{"КП-7 4/4 and 1", mt("4/4"), dur("1"), "", false, nil, nil, nil},
		{"sum", mt("2/4+3/8"), dur("1/8"), "3/4", true, nil, nil, nil},
		{"КП-40 3/4 and 1", mt("3/4"), dur("1"), "", false, measure.ErrLongerThanMeter, nil, nil},
		{"КП-39 unset meter", meter.Meter{}, dur("1/4"), "", false, measure.ErrInvalidArgument, []error{meter.ErrUnset}, nil},
		{"КП-39 unset pickup", mt("4/4"), duration.Duration{}, "", false, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"unset meter before unset pickup", meter.Meter{}, duration.Duration{}, "", false, measure.ErrInvalidArgument, []error{meter.ErrUnset}, []error{duration.ErrUnset}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			closing, ok, err := measure.Complement(tt.meter, tt.pickup)

			if tt.err != nil {
				assertErr(t, err, tt.err, tt.causes...)
				assertNoCause(t, err, tt.without...)
				assert.False(t, ok)
				assert.False(t, closing.IsSet())

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.ok, ok)

			if tt.ok {
				assertText(t, tt.closing, closing)
			} else {
				assert.False(t, closing.IsSet(), "no closing bar gives the unset duration")
			}
		})
	}
}

// TestComplementClosingBar covers О-18 with О-5: the closing length fits a
// grid with a pickup as an explicit length of its last bar.
func TestComplementClosingBar(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "4/4;pickup:1/4;bars:8")

	closing, ok, err := measure.Complement(mt("4/4"), dur("1/4"))
	require.NoError(t, err)
	require.True(t, ok)

	got, err := g.WithLength(8, closing, measure.AlignStart)
	require.NoError(t, err)
	assertRecord(t, "4/4;pickup:1/4;8:length:3/4;bars:8", got)

	end, finite, err := got.End()
	require.NoError(t, err)
	assert.True(t, finite)
	assertText(t, "8/1", end, "the pickup and the closing bar make 8 full bars")
}
