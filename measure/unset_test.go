package measure_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
)

// Group Т-9, unset values (architecture 4.2; П-14; Т-14; КП-46, КП-47,
// КП-59).

// unsetCall is an operation on a grid that returns its error and whether
// its other results are the values of architecture 4.2 for an error.
type unsetCall struct {
	name string
	call func(g measure.Grid) (zero bool, err error)
}

// gridCalls lists every operation of Grid that returns an error, with
// arguments that are valid (valid true) or that are themselves invalid, to
// check that Е-1 of the receiver comes first.
func gridCalls(valid bool) []unsetCall {
	bar, first, last, n := 3, 1, 4, 2
	p, d, a := pos("1/8"), dur("1/4"), ad(2, "1/8")
	m, c, level, align := mt("3/4"), frac("3/2"), measure.LevelBeats, measure.AlignStart
	other := measure.MustNew(mt("3/4"))

	if !valid {
		bar, first, last, n = -1, 5, 2, 0
		p, d, a = duration.Position{}, duration.Duration{}, measure.Address{}
		m, c, level, align = meter.Meter{}, fraction.Fraction{}, measure.Level(0), measure.Alignment(0)
		other = measure.Grid{}
	}

	return []unsetCall{
		{"Pickup", func(g measure.Grid) (bool, error) { l, ok, err := g.Pickup(); return !l.IsSet() && !ok, err }},
		{"HasPickup", func(g measure.Grid) (bool, error) { v, err := g.HasPickup(); return !v, err }},
		{"Marks", func(g measure.Grid) (bool, error) { v, err := g.Marks(); return v == nil, err }},
		{"HasMark", func(g measure.Grid) (bool, error) { v, err := g.HasMark(bar); return !v, err }},
		{"Changes", func(g measure.Grid) (bool, error) { v, err := g.Changes(); return v == nil, err }},
		{"ChangesAt", func(g measure.Grid) (bool, error) { v, err := g.ChangesAt(bar); return !v, err }},
		{"Lengths", func(g measure.Grid) (bool, error) { v, err := g.Lengths(); return v == nil, err }},
		{"BarCount", func(g measure.Grid) (bool, error) { v, f, err := g.BarCount(); return v == 0 && !f, err }},
		{"IsFinite", func(g measure.Grid) (bool, error) { v, err := g.IsFinite(); return !v, err }},
		{"TotalBars", func(g measure.Grid) (bool, error) { v, f, err := g.TotalBars(); return v == 0 && !f, err }},
		{"End", func(g measure.Grid) (bool, error) { v, f, err := g.End(); return !v.IsSet() && !f, err }},
		{"FirstBar", func(g measure.Grid) (bool, error) { v, err := g.FirstBar(); return v == 0, err }},
		{"HasBar", func(g measure.Grid) (bool, error) { v, err := g.HasBar(bar); return !v, err }},
		{"Bar", func(g measure.Grid) (bool, error) { v, err := g.Bar(bar); return v == measure.Bar{}, err }},
		{"BarMeter", func(g measure.Grid) (bool, error) { v, err := g.BarMeter(bar); return !v.IsSet(), err }},
		{"BarLength", func(g measure.Grid) (bool, error) { v, err := g.BarLength(bar); return !v.IsSet(), err }},
		{"BarStart", func(g measure.Grid) (bool, error) { v, err := g.BarStart(bar); return !v.IsSet(), err }},
		{"BarEnd", func(g measure.Grid) (bool, error) { v, err := g.BarEnd(bar); return !v.IsSet(), err }},
		{"Bars", func(g measure.Grid) (bool, error) { v, err := g.Bars(first, last); return v == nil, err }},
		{"Ordinal", func(g measure.Grid) (bool, error) { v, err := g.Ordinal(bar); return v == 0, err }},
		{"BarAtOrdinal", func(g measure.Grid) (bool, error) { v, err := g.BarAtOrdinal(n); return v == 0, err }},
		{"Locate", func(g measure.Grid) (bool, error) { v, err := g.Locate(p); return v == measure.Address{}, err }},
		{"BarAt", func(g measure.Grid) (bool, error) { v, err := g.BarAt(p); return v == measure.Bar{}, err }},
		{"Remaining", func(g measure.Grid) (bool, error) { v, err := g.Remaining(p); return !v.IsSet(), err }},
		{"OnBarline", func(g measure.Grid) (bool, error) { v, err := g.OnBarline(p); return !v, err }},
		{"PositionOf", func(g measure.Grid) (bool, error) { v, err := g.PositionOf(a); return !v.IsSet(), err }},
		{"Contains", func(g measure.Grid) (bool, error) { v, err := g.Contains(a); return !v, err }},
		{"Normalize", func(g measure.Grid) (bool, error) { v, err := g.Normalize(a); return v == measure.Address{}, err }},
		{"RangeLength", func(g measure.Grid) (bool, error) { v, err := g.RangeLength(first, last); return !v.IsSet(), err }},
		{"ChangesWithin", func(g measure.Grid) (bool, error) { v, err := g.ChangesWithin(first, last); return !v, err }},
		{"AllFull", func(g measure.Grid) (bool, error) { v, err := g.AllFull(first, last); return !v, err }},
		{"MeterRuns", func(g measure.Grid) (bool, error) { v, err := g.MeterRuns(first, last); return v == nil, err }},
		{"BarsSpanned", func(g measure.Grid) (bool, error) { f, l, err := g.BarsSpanned(p, d); return f == 0 && l == 0, err }},
		{"Elapsed", func(g measure.Grid) (bool, error) { v, err := g.Elapsed(a); return !v.IsValid(), err }},
		{"PartOfBar", func(g measure.Grid) (bool, error) { v, err := g.PartOfBar(bar, d); return !v.IsValid(), err }},
		{"MetricPlace", func(g measure.Grid) (bool, error) {
			v, u, err := g.MetricPlace(a)
			return !v.IsSet() && !u.IsSet(), err
		}},
		{"CountAt", func(g measure.Grid) (bool, error) { v, err := g.CountAt(bar, p); return !v.IsValid(), err }},
		{"MetricCountAt", func(g measure.Grid) (bool, error) { v, err := g.MetricCountAt(bar, p); return !v.IsValid(), err }},
		{"CountOffset", func(g measure.Grid) (bool, error) { v, err := g.CountOffset(bar, c); return !v.IsSet(), err }},
		{"MetricCountOffset", func(g measure.Grid) (bool, error) { v, err := g.MetricCountOffset(bar, c); return !v.IsSet(), err }},
		{"TstampAt", func(g measure.Grid) (bool, error) { v, err := g.TstampAt(bar, p); return !v.IsValid(), err }},
		{"TstampOffset", func(g measure.Grid) (bool, error) { v, err := g.TstampOffset(bar, c); return !v.IsSet(), err }},
		{"Split", func(g measure.Grid) (bool, error) { v, err := g.Split(p, d); return v == nil, err }},
		{"SplitBy", func(g measure.Grid) (bool, error) { v, err := g.SplitBy(p, d, level); return v == nil, err }},
		{"WithPickup", func(g measure.Grid) (bool, error) { v, err := g.WithPickup(d); return !v.IsSet(), err }},
		{"WithoutPickup", func(g measure.Grid) (bool, error) { v, err := g.WithoutPickup(); return !v.IsSet(), err }},
		{"WithMark", func(g measure.Grid) (bool, error) { v, err := g.WithMark(bar, m); return !v.IsSet(), err }},
		{"WithoutMark", func(g measure.Grid) (bool, error) { v, err := g.WithoutMark(bar); return !v.IsSet(), err }},
		{"WithMetersIn", func(g measure.Grid) (bool, error) { v, err := g.WithMetersIn(first, last, m); return !v.IsSet(), err }},
		{"WithLength", func(g measure.Grid) (bool, error) { v, err := g.WithLength(bar, d, align); return !v.IsSet(), err }},
		{"WithoutLength", func(g measure.Grid) (bool, error) { v, err := g.WithoutLength(bar); return !v.IsSet(), err }},
		{"WithBarCount", func(g measure.Grid) (bool, error) { v, err := g.WithBarCount(n); return !v.IsSet(), err }},
		{"WithoutBarCount", func(g measure.Grid) (bool, error) { v, err := g.WithoutBarCount(); return !v.IsSet(), err }},
		{"Truncate", func(g measure.Grid) (bool, error) { v, err := g.Truncate(n); return !v.IsSet(), err }},
		{"AddBars", func(g measure.Grid) (bool, error) { v, err := g.AddBars(n); return !v.IsSet(), err }},
		{"InsertBars", func(g measure.Grid) (bool, error) { v, err := g.InsertBars(bar, n); return !v.IsSet(), err }},
		{"DeleteBars", func(g measure.Grid) (bool, error) { v, err := g.DeleteBars(first, last); return !v.IsSet(), err }},
		{"Slice", func(g measure.Grid) (bool, error) { v, err := g.Slice(first, last); return !v.IsSet(), err }},
		{"Append", func(g measure.Grid) (bool, error) { v, err := g.Append(other); return !v.IsSet(), err }},
		{"Simplify", func(g measure.Grid) (bool, error) { v, err := g.Simplify(); return !v.IsSet(), err }},
		{"EqualMarkup", func(g measure.Grid) (bool, error) { v, err := g.EqualMarkup(other); return !v, err }},
		{"EqualBarLengths", func(g measure.Grid) (bool, error) { v, err := g.EqualBarLengths(other); return !v, err }},
		{"MarshalText", func(g measure.Grid) (bool, error) { v, err := g.MarshalText(); return v == nil, err }},
		{"AppendText", func(g measure.Grid) (bool, error) {
			dst := []byte("keep")
			v, err := g.AppendText(dst)

			return string(v) == "keep" && len(v) == 4, err
		}},
	}
}

// TestUnsetGrid covers П-14 and КП-47, КП-59: the unset grid works only with
// IsSet, Equal, and String; every other operation fails with Е-1 before
// checking its arguments and returns the values for an error.
func TestUnsetGrid(t *testing.T) {
	t.Parallel()

	t.Run("allowed", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		var g measure.Grid
		assert.False(t, g.IsSet())
		assert.True(t, g.Equal(measure.Grid{}))
		assert.False(t, g.Equal(measure.MustNew(mt("4/4"))))
		assert.False(t, measure.MustNew(mt("4/4")).Equal(g))
		assert.Equal(t, "Grid(unset)", g.String())
	})

	for _, valid := range []bool{true, false} {
		for _, c := range gridCalls(valid) {
			name := c.name
			if !valid {
				name += " with invalid arguments"
			}

			t.Run(name, func(t *testing.T) {
				t.Parallel()
				defer guard(t)

				zero, err := c.call(measure.Grid{})
				assertErr(t, err, measure.ErrUnset)
				assert.True(t, zero, "the values for an error")
			})
		}
	}

	t.Run("all operations are listed", func(t *testing.T) {
		t.Parallel()
		assert.Len(t, gridCalls(true), 64)
	})
}

// TestUnsetArgument covers П-14 and КП-47: the unset grid and the unset
// address as an argument fail with Е-1.
func TestUnsetArgument(t *testing.T) {
	t.Parallel()

	g := func(t *testing.T) measure.Grid { return gridOf(t, "4/4;bars:4") }
	open := func(t *testing.T) measure.Grid { return gridOf(t, "4/4") }

	tests := []struct {
		name string
		call func(t *testing.T) error
	}{
		{"Append", func(t *testing.T) error { return e(g(t).Append(measure.Grid{})) }},
		{"Append to an open grid", func(t *testing.T) error { return e(open(t).Append(measure.Grid{})) }},
		{"EqualMarkup", func(t *testing.T) error { return e(g(t).EqualMarkup(measure.Grid{})) }},
		{"EqualBarLengths", func(t *testing.T) error { return e(g(t).EqualBarLengths(measure.Grid{})) }},
		{"PositionOf", func(t *testing.T) error { return e(g(t).PositionOf(measure.Address{})) }},
		{"Contains", func(t *testing.T) error { return e(g(t).Contains(measure.Address{})) }},
		{"Normalize", func(t *testing.T) error { return e(g(t).Normalize(measure.Address{})) }},
		{"Elapsed", func(t *testing.T) error { return e(g(t).Elapsed(measure.Address{})) }},
		{"MetricPlace", func(t *testing.T) error { return e2(g(t).MetricPlace(measure.Address{})) }},
		{"EqualMarkup of two unset", func(t *testing.T) error { return e(measure.Grid{}.EqualMarkup(measure.Grid{})) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertErr(t, tt.call(t), measure.ErrUnset)
		})
	}
}

// TestUnsetAddress covers П-14 and КП-47, КП-59: the unset address works
// only with IsSet, Equal, Compare, and String.
func TestUnsetAddress(t *testing.T) {
	t.Parallel()
	defer guard(t)

	var a measure.Address

	assert.False(t, a.IsSet())
	assert.True(t, a.Equal(measure.Address{}))
	assert.False(t, a.Equal(ad(0, "0")))
	assert.False(t, ad(0, "0").Equal(a))
	assert.Equal(t, "Address(unset)", a.String())

	assert.Equal(t, 0, a.Compare(measure.Address{}))
	assert.Equal(t, -1, a.Compare(ad(0, "0")), "the unset address comes first")
	assert.Equal(t, 1, ad(0, "0").Compare(a))

	bar, err := a.Bar()
	assertErr(t, err, measure.ErrUnset)
	assert.Zero(t, bar)

	off, err := a.Offset()
	assertErr(t, err, measure.ErrUnset)
	assert.False(t, off.IsSet())

	b, err := a.MarshalText()
	assertErr(t, err, measure.ErrUnset)
	assert.Nil(t, b)

	dst := []byte("keep")
	out, err := a.AppendText(dst)
	assertErr(t, err, measure.ErrUnset)
	assert.Equal(t, "keep", string(out))
}

// TestUnsetText covers КП-46: parsing rejects the strings of the unset
// values.
func TestUnsetText(t *testing.T) {
	t.Parallel()
	defer guard(t)

	_, err := measure.Parse(measure.Grid{}.String())
	assertErrExact(t, err, measure.ErrInvalidText)

	_, err = measure.ParseAddress(measure.Address{}.String())
	assertErrExact(t, err, measure.ErrInvalidText)
}
