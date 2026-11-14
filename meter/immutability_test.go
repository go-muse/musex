package meter_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArgumentsAreNotKept covers ИР-12, Т-14, КП-48 and architecture 3.2:
// changing a slice passed to a constructor or to WithGrouping after the call
// does not change the value.
func TestArgumentsAreNotKept(t *testing.T) {
	t.Parallel()

	t.Run("NewAdditive", func(t *testing.T) {
		t.Parallel()

		addends := []int{3, 2, 3}
		m := meter.MustNewAdditive(addends, 8)
		want := m
		addends[0], addends[1] = 1, 5

		assert.Equal(t, want, m)
		assert.Equal(t, []int{3, 2, 3}, m.MustAddends())
		assert.Equal(t, []int{3, 2, 3}, m.MustGrouping())
		assert.Equal(t, "(3+2+3)/8", m.String())
	})

	t.Run("WithGrouping", func(t *testing.T) {
		t.Parallel()

		groups := []int{3, 2, 2}
		m := meter.MustNew(7, 8).MustWithGrouping(groups...)
		groups[0], groups[1] = 2, 3

		assert.Equal(t, []int{3, 2, 2}, m.MustGrouping())
		assert.Equal(t, "7/8[3+2+2]", m.String())
	})

	t.Run("NewSum", func(t *testing.T) {
		t.Parallel()

		parts := mts("2/4", "3/8")
		m := meter.MustNewSum(parts...)
		parts[0] = mt("1/4")

		assert.Equal(t, mts("2/4", "3/8"), m.MustParts())
		assert.Equal(t, "2/4+3/8", m.String())
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		t.Parallel()

		text := []byte("5/8[2+3]")

		var m meter.Meter
		require.NoError(t, m.UnmarshalText(text))
		copy(text, "3/4[1+2]")

		assert.Equal(t, mt("5/8[2+3]"), m)
	})
}

// TestResultsAreOwned covers ИР-12, КП-48 and architecture 3.2: slices
// returned by the package are new on every call; changing them does not
// change the value or the next results.
func TestResultsAreOwned(t *testing.T) {
	t.Parallel()

	m := mt("7/8[3+2+2]")
	sum := mt("2/4+3/8")
	want := m

	parts := sum.MustParts()
	parts[0] = mt("C|")
	assert.Equal(t, mts("2/4", "3/8"), sum.MustParts())

	addends := sum.MustAddends()
	addends[0] = 99
	assert.Equal(t, []int{2, 3}, sum.MustAddends())

	units := m.MustUnits()
	units[0] = sp("1@0")
	assert.Equal(t, equalSpans(7, "1/8"), m.MustUnits())

	grouping := m.MustGrouping()
	grouping[0] = 7
	assert.Equal(t, []int{3, 2, 2}, m.MustGrouping())

	groups := m.MustGroups()
	groups[1] = sp("1@0")
	assert.Equal(t, sps("3/8@0", "1/4@3/8", "1/4@5/8"), m.MustGroups())

	gw := m.MustGroupsWritten()
	gw[0] = duration.MustParseWritten("1")
	assert.Equal(t, wrs("4.", "4", "4"), m.MustGroupsWritten())

	beats := m.MustBeats()
	beats[0] = sp("1@0")
	assert.Equal(t, sps("3/8@0", "1/4@3/8", "1/4@5/8"), m.MustBeats())

	bw := m.MustBeatsWritten()
	bw[0] = duration.MustParseWritten("1")
	assert.Equal(t, wrs("4.", "4", "4"), m.MustBeatsWritten())

	weights := m.MustUnitWeights()
	weights[0] = meter.WeightOffUnit
	assert.Equal(t, []meter.Weight{wS, wW, wW, wR, wW, wR, wW}, m.MustUnitWeights())

	_, rel := m.MustStrongBeats()
	rel[0] = pos(0, 1)
	_, rel = m.MustStrongBeats()
	assert.Equal(t, poss("3/8", "5/8"), rel)

	variants := m.MustGroupingVariants()
	variants[0][0] = 9
	variants[1] = nil
	assert.Equal(t, compositions(7), m.MustGroupingVariants())

	fv := meter.MustGroupingVariants(7)
	fv[0][1] = 9
	assert.Equal(t, compositions(7), meter.MustGroupingVariants(7))

	data := noErr(m.MarshalText())
	data[0] = '9'
	assert.Equal(t, "7/8[3+2+2]", string(noErr(m.MarshalText())))

	assert.Equal(t, want, m, "the value is unchanged")
}

// TestAppendText covers architecture 3.2: AppendText does not change
// dst[:len(dst)] and appends after it, also when dst has spare capacity.
func TestAppendText(t *testing.T) {
	t.Parallel()

	dst := make([]byte, 3, 64)
	copy(dst, "ab:")

	out, err := mt("5/8[2+3]").AppendText(dst)
	require.NoError(t, err)
	assert.Equal(t, "ab:5/8[2+3]", string(out))
	assert.Equal(t, "ab:", string(dst))

	out, err = meter.Meter{}.AppendText(dst)
	assertErr(t, err, meter.ErrUnset)
	assert.Equal(t, "ab:", string(out))

	out, err = meter.CommonTime().AppendText(nil)
	require.NoError(t, err)
	assert.Equal(t, "C", string(out))
}

// TestInputsUnchangedOnError covers КП-48: the inputs are the same after
// failing calls and after the panics of Must variants.
func TestInputsUnchangedOnError(t *testing.T) {
	t.Parallel()

	m := mt("5/8")
	want := m

	_, err := m.Beats()
	require.Error(t, err)

	_, err = m.WithGrouping(2, 2)
	require.Error(t, err)

	_, err = m.ToSum()
	require.Error(t, err)

	_, _ = capturePanic(func() { m.MustGroups() })
	_, _ = capturePanic(func() { m.MustWithGrouping(1) })

	assert.Equal(t, want, m)

	addends := []int{1000, 25}
	_, err = meter.NewAdditive(addends, 8)
	require.Error(t, err)
	assert.Equal(t, []int{1000, 25}, addends)

	parts := []meter.Meter{mt("2/4"), {}}
	_, err = meter.NewSum(parts...)
	require.Error(t, err)
	assert.Equal(t, []meter.Meter{mt("2/4"), {}}, parts)

	groups := []int{2, 2}
	_, _ = capturePanic(func() { m.MustWithGrouping(groups...) })
	assert.Equal(t, []int{2, 2}, groups)

	got := meter.SixEight()
	require.Error(t, got.UnmarshalText([]byte("x")))
	assert.Equal(t, meter.SixEight(), got)
}
