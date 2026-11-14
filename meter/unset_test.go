package meter_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
)

// TestErrorSentinels covers #model 2.6, Т-13 and architecture 4.1, 4.2: the
// ten sentinels are set, pairwise distinct, and none wraps another or a
// sentinel of a dependency.
func TestErrorSentinels(t *testing.T) {
	t.Parallel()
	defer guard(t)

	assert.Len(t, meterSentinels, 10)

	for _, a := range meterSentinels {
		assert.Error(t, a.err, a.name)

		for _, b := range allSentinels() {
			if a.err == b.err {
				continue
			}

			assert.False(t, errors.Is(a.err, b.err), "%s wraps %s", a.name, b.name)
		}
	}
}

// unsetCall is an operation of a meter with an error result.
type unsetCall struct {
	name string
	call func(m meter.Meter) error
}

// meterCalls lists every method of Meter with an error, with valid other
// arguments, except UnmarshalText (architecture 4.1).
func meterCalls() []unsetCall {
	p := pos(1, 8)
	v := dur(1, 8)
	other := meter.ThreeFour()

	return []unsetCall{
		{"Form", func(m meter.Meter) error { return e(m.Form()) }},
		{"Symbol", func(m meter.Meter) error { return e2(m.Symbol()) }},
		{"PartCount", func(m meter.Meter) error { return e(m.PartCount()) }},
		{"Parts", func(m meter.Meter) error { return e(m.Parts()) }},
		{"Addends", func(m meter.Meter) error { return e(m.Addends()) }},
		{"Numerator", func(m meter.Meter) error { return e(m.Numerator()) }},
		{"Denominator", func(m meter.Meter) error { return e(m.Denominator()) }},
		{"HasCommonDenominator", func(m meter.Meter) error { return e(m.HasCommonDenominator()) }},
		{"HasPowerOfTwoDenominators", func(m meter.Meter) error { return e(m.HasPowerOfTwoDenominators()) }},
		{"DenominatorExponent", func(m meter.Meter) error { return e(m.DenominatorExponent()) }},
		{"Duration", func(m meter.Meter) error { return e(m.Duration()) }},
		{"QuoRem", func(m meter.Meter) error { _, _, err := m.QuoRem(v); return err }},
		{"UnitCount", func(m meter.Meter) error { return e(m.UnitCount()) }},
		{"Unit", func(m meter.Meter) error { return e(m.Unit()) }},
		{"UnitWritten", func(m meter.Meter) error { return e(m.UnitWritten()) }},
		{"Units", func(m meter.Meter) error { return e(m.Units()) }},
		{"UnitStart", func(m meter.Meter) error { return e(m.UnitStart(1)) }},
		{"UnitAt", func(m meter.Meter) error { return e2(m.UnitAt(p)) }},
		{"HasGrouping", func(m meter.Meter) error { return e(m.HasGrouping()) }},
		{"HasExplicitGrouping", func(m meter.Meter) error { return e(m.HasExplicitGrouping()) }},
		{"Grouping", func(m meter.Meter) error { return e(m.Grouping()) }},
		{"GroupCount", func(m meter.Meter) error { return e(m.GroupCount()) }},
		{"Groups", func(m meter.Meter) error { return e(m.Groups()) }},
		{"GroupsWritten", func(m meter.Meter) error { return e(m.GroupsWritten()) }},
		{"GroupStart", func(m meter.Meter) error { return e(m.GroupStart(1)) }},
		{"GroupAt", func(m meter.Meter) error { return e2(m.GroupAt(p)) }},
		{"WithGrouping", func(m meter.Meter) error { return e(m.WithGrouping(2, 1)) }},
		{"ResetGrouping", func(m meter.Meter) error { return e(m.ResetGrouping()) }},
		{"GroupingVariants", func(m meter.Meter) error { return e(m.GroupingVariants()) }},
		{"BeatCount", func(m meter.Meter) error { return e(m.BeatCount()) }},
		{"Beats", func(m meter.Meter) error { return e(m.Beats()) }},
		{"BeatsWritten", func(m meter.Meter) error { return e(m.BeatsWritten()) }},
		{"BeatStart", func(m meter.Meter) error { return e(m.BeatStart(1)) }},
		{"BeatAt", func(m meter.Meter) error { return e2(m.BeatAt(p)) }},
		{"Kind", func(m meter.Meter) error { return e(m.Kind()) }},
		{"RussianKind", func(m meter.Meter) error { return e(m.RussianKind()) }},
		{"BeatDivision", func(m meter.Meter) error { return e2(m.BeatDivision()) }},
		{"KindName", func(m meter.Meter) error { return e2(m.KindName()) }},
		{"KindRussianName", func(m meter.Meter) error { return e2(m.KindRussianName()) }},
		{"StrongBeats", func(m meter.Meter) error { return e2(m.StrongBeats()) }},
		{"UnitWeights", func(m meter.Meter) error { return e(m.UnitWeights()) }},
		{"WeightAt", func(m meter.Meter) error { return e(m.WeightAt(p)) }},
		{"ToNumeric", func(m meter.Meter) error { return e(m.ToNumeric()) }},
		{"ToSymbol", func(m meter.Meter) error { return e(m.ToSymbol()) }},
		{"ToSimple", func(m meter.Meter) error { return e(m.ToSimple()) }},
		{"ToAdditive", func(m meter.Meter) error { return e(m.ToAdditive()) }},
		{"ToSum", func(m meter.Meter) error { return e(m.ToSum()) }},
		{"Shorter", func(m meter.Meter) error { return e(m.Shorter()) }},
		{"Longer", func(m meter.Meter) error { return e(m.Longer()) }},
		{"WithDenominator", func(m meter.Meter) error { return e(m.WithDenominator(4)) }},
		{"EqualStructure", func(m meter.Meter) error { return e(m.EqualStructure(other)) }},
		{"EqualDuration", func(m meter.Meter) error { return e(m.EqualDuration(other)) }},
		{"CompareDuration", func(m meter.Meter) error { return e(m.CompareDuration(other)) }},
		{"MarshalText", func(m meter.Meter) error { return e(m.MarshalText()) }},
		{"AppendText", func(m meter.Meter) error { return e(m.AppendText(nil)) }},
	}
}

// TestUnsetReceiver covers П-10, Т-12, КП-38: every operation with an unset
// receiver gives E-1, also with an invalid argument (E-1 is checked first).
func TestUnsetReceiver(t *testing.T) {
	t.Parallel()
	defer guard(t)

	var u meter.Meter

	calls := meterCalls()
	assert.Len(t, calls, 55)

	invalid := []unsetCall{
		{"QuoRem unset", func(m meter.Meter) error { _, _, err := m.QuoRem(duration.Duration{}); return err }},
		{"UnitStart 0", func(m meter.Meter) error { return e(m.UnitStart(0)) }},
		{"UnitAt unset", func(m meter.Meter) error { return e2(m.UnitAt(duration.Position{})) }},
		{"GroupStart 0", func(m meter.Meter) error { return e(m.GroupStart(0)) }},
		{"GroupAt unset", func(m meter.Meter) error { return e2(m.GroupAt(duration.Position{})) }},
		{"BeatStart 0", func(m meter.Meter) error { return e(m.BeatStart(0)) }},
		{"BeatAt unset", func(m meter.Meter) error { return e2(m.BeatAt(duration.Position{})) }},
		{"WeightAt unset", func(m meter.Meter) error { return e(m.WeightAt(duration.Position{})) }},
		{"WithGrouping none", func(m meter.Meter) error { return e(m.WithGrouping()) }},
		{"WithDenominator 0", func(m meter.Meter) error { return e(m.WithDenominator(0)) }},
		{"EqualStructure unset", func(m meter.Meter) error { return e(m.EqualStructure(meter.Meter{})) }},
		{"EqualDuration unset", func(m meter.Meter) error { return e(m.EqualDuration(meter.Meter{})) }},
		{"CompareDuration unset", func(m meter.Meter) error { return e(m.CompareDuration(meter.Meter{})) }},
	}

	for _, c := range append(calls, invalid...) {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertErr(t, c.call(u), meter.ErrUnset)
		})
	}
}

// TestUnsetArgument covers П-10, КП-38: an unset meter as the argument of
// EqualStructure, EqualDuration, CompareDuration gives E-1.
func TestUnsetArgument(t *testing.T) {
	t.Parallel()
	defer guard(t)

	var u meter.Meter

	m := meter.SixEight()

	for _, c := range []unsetCall{
		{"EqualStructure", func(m meter.Meter) error { return e(m.EqualStructure(u)) }},
		{"EqualDuration", func(m meter.Meter) error { return e(m.EqualDuration(u)) }},
		{"CompareDuration", func(m meter.Meter) error { return e(m.CompareDuration(u)) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertErr(t, c.call(m), meter.ErrUnset)
		})
	}
}

// TestUnsetWithoutError covers П-10, КП-47 and architecture 4.2: IsSet,
// Equal, Compare, and String work without an error.
func TestUnsetWithoutError(t *testing.T) {
	t.Parallel()
	defer guard(t)

	var u meter.Meter

	assert.False(t, u.IsSet())
	assert.True(t, u.Equal(meter.Meter{}))
	assert.False(t, u.Equal(meter.ThreeFour()))
	assert.False(t, meter.ThreeFour().Equal(u))
	assert.Equal(t, 0, u.Compare(meter.Meter{}))
	assert.Equal(t, -1, u.Compare(meter.ThreeFour()))
	assert.Equal(t, 1, meter.ThreeFour().Compare(u))
	assert.Equal(t, "Meter(unset)", u.String())

	_, err := meter.Parse(u.String())
	assertErrExact(t, err, meter.ErrInvalidText)
}

// TestZeroResultsOnError covers architecture 4.2: on an error every result
// is the zero value of its type.
func TestZeroResultsOnError(t *testing.T) {
	t.Parallel()
	defer guard(t)

	var u meter.Meter

	form, err := u.Form()
	assert.Error(t, err)
	assert.Equal(t, meter.Form(0), form)

	s, ok, err := u.Symbol()
	assert.Error(t, err)
	assert.Equal(t, meter.Symbol(0), s)
	assert.False(t, ok)

	n, err := u.PartCount()
	assert.Error(t, err)
	assert.Zero(t, n)

	parts, err := u.Parts()
	assert.Error(t, err)
	assert.Nil(t, parts)

	addends, err := u.Addends()
	assert.Error(t, err)
	assert.Nil(t, addends)

	has, err := u.HasGrouping()
	assert.Error(t, err)
	assert.False(t, has)

	size, err := u.Duration()
	assert.Error(t, err)
	assert.Equal(t, duration.Duration{}, size)

	units, err := u.Units()
	assert.Error(t, err)
	assert.Nil(t, units)

	written, err := mt("5/8").BeatsWritten()
	assert.Error(t, err)
	assert.Nil(t, written)

	kind, err := mt("8/8").Kind()
	assert.Error(t, err)
	assert.Equal(t, meter.Kind(0), kind)

	weights, err := mt("8/8").UnitWeights()
	assert.Error(t, err)
	assert.Nil(t, weights)

	strong, rel, err := mt("8/8").StrongBeats()
	assert.Error(t, err)
	assert.Equal(t, duration.Position{}, strong)
	assert.Nil(t, rel)

	name, ok, err := mt("8/8").KindName()
	assert.Error(t, err)
	assert.Empty(t, name)
	assert.False(t, ok)

	eq, err := u.EqualStructure(meter.ThreeFour())
	assert.Error(t, err)
	assert.False(t, eq)

	cmp, err := meter.ThreeFour().CompareDuration(u)
	assert.Error(t, err)
	assert.Zero(t, cmp)

	variants, err := u.GroupingVariants()
	assert.Error(t, err)
	assert.Nil(t, variants)

	w, err := u.WeightAt(pos(0, 1))
	assert.Error(t, err)
	assert.Equal(t, meter.Weight(0), w)

	rk, err := u.RussianKind()
	assert.Error(t, err)
	assert.Equal(t, meter.RussianKind(0), rk)
}
