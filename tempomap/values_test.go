package tempomap_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnumerations(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 1000, tempomap.MaxPlaces)

	for _, tt := range []struct {
		value tempomap.Addressing
		text  string
	}{
		{tempomap.PositionAddressing, "positions"},
		{tempomap.MeasureAddressing, "measures"},
	} {
		assert.True(t, tt.value.IsValid())
		assert.True(t, tt.value.Equal(tt.value))
		assert.Equal(t, tt.text, tt.value.String())
	}
	for _, value := range []tempomap.Addressing{0, 3, 255} {
		assert.False(t, value.IsValid())
		assert.Equal(t, "Addressing(invalid)", value.String())
		assert.True(t, value.Equal(value))
	}

	for _, tt := range []struct {
		value tempomap.Slot
		text  string
	}{
		{tempomap.SlotSingle, "single"},
		{tempomap.SlotFirst, "first"},
		{tempomap.SlotSecond, "second"},
	} {
		assert.True(t, tt.value.IsValid())
		assert.Equal(t, tt.text, tt.value.String())
	}
	for _, value := range []tempomap.Slot{0, 4, 255} {
		assert.False(t, value.IsValid())
		assert.Equal(t, "Slot(invalid)", value.String())
		assert.True(t, value.Equal(value))
	}

	for _, tt := range []struct {
		value tempomap.ChangeKind
		text  string
	}{
		{tempomap.ChangeMain, "main"},
		{tempomap.ChangeJump, "jump"},
		{tempomap.ChangeGradual, "gradual"},
		{tempomap.ChangeATempo, "a-tempo"},
		{tempomap.ChangeTempoI, "tempo-i"},
		{tempomap.ChangeModulation, "modulation"},
	} {
		assert.True(t, tt.value.IsValid())
		assert.Equal(t, tt.text, tt.value.String())
	}
	for _, value := range []tempomap.ChangeKind{0, 7, 255} {
		assert.False(t, value.IsValid())
		assert.Equal(t, "ChangeKind(invalid)", value.String())
		assert.True(t, value.Equal(value))
	}

	assert.True(t, tempomap.LinearSpeed.IsValid())
	assert.True(t, tempomap.LinearSpeed.Equal(tempomap.LinearSpeed))
	assert.Equal(t, "linear-speed", tempomap.LinearSpeed.String())
	for _, value := range []tempomap.Law{0, 2, 255} {
		assert.False(t, value.IsValid())
		assert.Equal(t, "Law(invalid)", value.String())
		assert.True(t, value.Equal(value))
	}
}

func TestAddressBranchesAndTotalOrder(t *testing.T) {
	t.Parallel()
	var unset tempomap.Address
	assert.False(t, unset.IsSet())
	assert.True(t, unset.Equal(tempomap.Address{}))
	assert.Equal(t, "Address(unset)", unset.String())
	assert.Zero(t, unset.Compare(tempomap.Address{}))

	position := at(3, 8)
	addressing, err := position.Addressing()
	require.NoError(t, err)
	assert.Equal(t, tempomap.PositionAddressing, addressing)
	actualPosition, ok, err := position.Position()
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, pos(3, 8).Equal(actualPosition))
	_, ok, err = position.Bar()
	require.NoError(t, err)
	assert.False(t, ok)

	measured := bar(2, 1, 8)
	addressing, err = measured.Addressing()
	require.NoError(t, err)
	assert.Equal(t, tempomap.MeasureAddressing, addressing)
	actualBar, ok, err := measured.Bar()
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, measure.MustNewAddress(2, pos(1, 8)).Equal(actualBar))
	_, ok, err = measured.Position()
	require.NoError(t, err)
	assert.False(t, ok)

	ordered := []tempomap.Address{unset, tempomap.MustAt(duration.Origin()), position, bar(1, 0, 1), measured}
	for i, left := range ordered {
		for j, right := range ordered {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			assert.Equal(t, want, left.Compare(right), "Compare(%s, %s)", left, right)
		}
	}

	_, err = tempomap.At(duration.Position{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, duration.ErrUnset)
	_, err = tempomap.AtBar(measure.Address{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, measure.ErrUnset)
}

func TestChangeConstructorsSlotsAndData(t *testing.T) {
	t.Parallel()
	main := tempomap.MustNewMain(at(1, 1), quarter(90))
	assert.True(t, main.IsSet())
	assert.True(t, at(1, 1).Equal(main.MustAddress()))
	assert.Equal(t, tempomap.SlotSingle, main.MustSlot())
	assert.Equal(t, tempomap.ChangeMain, main.MustKind())
	value, ok, err := main.Tempo()
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, quarter(90).Equal(value))
	_, ok, err = main.Law()
	require.NoError(t, err)
	assert.False(t, ok)
	_, _, ok, err = main.Modulation()
	require.NoError(t, err)
	assert.False(t, ok)

	jump := tempomap.MustNewJump(at(2, 1), quarter(120))
	assert.Equal(t, tempomap.ChangeJump, jump.MustKind())
	gradual := tempomap.MustNewGradual(at(3, 1), tempomap.LinearSpeed)
	law, ok, err := gradual.Law()
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, tempomap.LinearSpeed, law)
	_, ok, err = gradual.Tempo()
	require.NoError(t, err)
	assert.False(t, ok)

	aTempo := tempomap.MustNewATempo(at(4, 1))
	assert.Equal(t, tempomap.ChangeATempo, aTempo.MustKind())
	tempoI := tempomap.MustNewTempoI(at(5, 1))
	assert.Equal(t, tempomap.ChangeTempoI, tempoI.MustKind())
	modulation := tempomap.MustNewModulation(at(6, 1), duration.Quarter.MustWritten().MustAddDot(), duration.MustNew(1, 4))
	newUnit, old, ok, err := modulation.Modulation()
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, duration.Quarter.MustWritten().MustAddDot().Equal(newUnit))
	assert.True(t, duration.MustNew(1, 4).Equal(old))

	first := main.MustWithSlot(tempomap.SlotFirst)
	assert.Equal(t, tempomap.SlotFirst, first.MustSlot())
	assert.False(t, first.Equal(main))
	assert.True(t, first.Equal(first.MustWithSlot(tempomap.SlotFirst)))

	_, err = tempomap.NewGradual(at(1, 1), tempomap.Law(0))
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	_, err = main.WithSlot(tempomap.Slot(0))
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)

	for _, constructor := range []struct {
		name string
		call func() error
	}{
		{"main", func() error { _, err := tempomap.NewMain(tempomap.Address{}, quarter(60)); return err }},
		{"jump", func() error { _, err := tempomap.NewJump(tempomap.Address{}, quarter(60)); return err }},
		{"gradual", func() error { _, err := tempomap.NewGradual(tempomap.Address{}, tempomap.LinearSpeed); return err }},
		{"a tempo", func() error { _, err := tempomap.NewATempo(tempomap.Address{}); return err }},
		{"tempo i", func() error { _, err := tempomap.NewTempoI(tempomap.Address{}); return err }},
		{"modulation", func() error {
			_, err := tempomap.NewModulation(tempomap.Address{}, duration.Quarter.MustWritten(), duration.MustNew(1, 4))
			return err
		}},
	} {
		t.Run(constructor.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, constructor.call(), tempomap.ErrUnset)
		})
	}
}

func TestExtensionValueAndReplacement(t *testing.T) {
	t.Parallel()
	extension := tempomap.MustNewExtension(at(1, 2), sec(3, 2))
	assert.True(t, extension.IsSet())
	assert.True(t, at(1, 2).Equal(extension.MustAddress()))
	assert.Equal(t, sec(3, 2), extension.MustSeconds())
	assert.True(t, extension.Equal(extension))

	m := tempomap.MustNew(quarter(60)).MustWithExtension(extension)
	m = m.MustWithExtension(tempomap.MustNewExtension(at(1, 2), sec(2, 1)))
	require.Len(t, m.MustExtensions(), 1)
	assert.Equal(t, sec(2, 1), m.MustExtensions()[0].MustSeconds())
	m = m.MustWithoutExtension(at(1, 2))
	assert.Empty(t, m.MustExtensions())

	_, err := tempomap.NewExtension(tempomap.Address{}, sec(1, 1))
	requireTempomapError(t, err, tempomap.ErrUnset)
	_, err = tempomap.NewExtension(at(1, 1), tempo.Seconds{})
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, tempo.ErrUnset)
}

func TestMapCreationAndReading(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	assert.True(t, m.IsSet())
	assert.Equal(t, tempomap.PositionAddressing, m.MustAddressing())
	assert.True(t, quarter(60).Equal(m.MustInitial()))
	assert.False(t, m.MustUnfinished())
	require.Len(t, m.MustChanges(), 1)
	assert.Equal(t, tempomap.SlotFirst, m.MustChanges()[0].MustSlot())
	assert.Empty(t, m.MustExtensions())
	unfinished, err := m.Validate()
	require.NoError(t, err)
	assert.False(t, unfinished)

	grid := grid34Pickup()
	measured := tempomap.MustNewMeasured(grid, quarter(60))
	assert.Equal(t, tempomap.MeasureAddressing, measured.MustAddressing())
	assert.True(t, bar(0, 0, 1).Equal(measured.MustChanges()[0].MustAddress()))
	unfinished, err = measured.Validate(grid)
	require.NoError(t, err)
	assert.False(t, unfinished)
}

func TestUnsetValuesAllowOnlyInspectionEqualityAndString(t *testing.T) {
	t.Parallel()
	var address tempomap.Address
	var change tempomap.Change
	var extension tempomap.Extension
	var m tempomap.Map
	var state tempomap.State
	var segment tempomap.Segment
	var seconds tempomap.SecondsResult
	var position tempomap.PositionResult

	assert.False(t, change.IsSet())
	assert.False(t, extension.IsSet())
	assert.False(t, m.IsSet())
	assert.False(t, state.IsSet())
	assert.False(t, segment.IsSet())
	assert.False(t, seconds.IsSet())
	assert.False(t, position.IsSet())
	assert.True(t, change.Equal(tempomap.Change{}))
	assert.True(t, extension.Equal(tempomap.Extension{}))
	assert.True(t, m.Equal(tempomap.Map{}))
	assert.True(t, state.Equal(tempomap.State{}))
	assert.True(t, segment.Equal(tempomap.Segment{}))
	assert.True(t, seconds.Equal(tempomap.SecondsResult{}))
	assert.True(t, position.Equal(tempomap.PositionResult{}))
	assert.Equal(t, "Change(unset)", change.String())
	assert.Equal(t, "Extension(unset)", extension.String())
	assert.Equal(t, "Map(unset)", m.String())
	assert.Equal(t, "State(unset)", state.String())
	assert.Equal(t, "Segment(unset)", segment.String())
	assert.Equal(t, "SecondsResult(unset)", seconds.String())
	assert.Equal(t, "PositionResult(unset)", position.String())

	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"address addressing", func() error { _, err := address.Addressing(); return err }},
		{"address position", func() error { _, _, err := address.Position(); return err }},
		{"address bar", func() error { _, _, err := address.Bar(); return err }},
		{"address marshal", func() error { _, err := address.MarshalText(); return err }},
		{"address append", func() error { _, err := address.AppendText([]byte("prefix")); return err }},
		{"change address", func() error { _, err := change.Address(); return err }},
		{"change slot", func() error { _, err := change.Slot(); return err }},
		{"change kind", func() error { _, err := change.Kind(); return err }},
		{"change tempo", func() error { _, _, err := change.Tempo(); return err }},
		{"change law", func() error { _, _, err := change.Law(); return err }},
		{"change modulation", func() error { _, _, _, err := change.Modulation(); return err }},
		{"change with slot", func() error { _, err := change.WithSlot(tempomap.SlotSingle); return err }},
		{"change marshal", func() error { _, err := change.MarshalText(); return err }},
		{"extension address", func() error { _, err := extension.Address(); return err }},
		{"extension seconds", func() error { _, err := extension.Seconds(); return err }},
		{"map addressing", func() error { _, err := m.Addressing(); return err }},
		{"map initial", func() error { _, err := m.Initial(); return err }},
		{"map changes", func() error { _, err := m.Changes(); return err }},
		{"map extensions", func() error { _, err := m.Extensions(); return err }},
		{"map unfinished", func() error { _, err := m.Unfinished(); return err }},
		{"map validate", func() error { _, err := m.Validate(); return err }},
		{"map state", func() error { _, err := m.State(at(1, 1)); return err }},
		{"map speed", func() error { _, err := m.Speed(at(1, 1)); return err }},
		{"map tempo", func() error { _, err := m.TempoAt(at(1, 1), duration.Quarter.MustWritten()); return err }},
		{"map segments", func() error { _, err := m.Segments(); return err }},
		{"map start", func() error { _, err := m.Start(at(1, 1), 0, rounding.HalfUp); return err }},
		{"map end", func() error { _, err := m.End(at(1, 1), 0, rounding.HalfUp); return err }},
		{"map position", func() error { _, err := m.Position(sec(1, 1), 0, rounding.HalfUp); return err }},
		{"map address at", func() error { _, _, err := m.AddressAt(grid34Pickup(), sec(1, 1), 0, rounding.HalfUp); return err }},
		{"map between", func() error { _, err := m.Between(at(0, 1), at(1, 1), 0, rounding.HalfUp); return err }},
		{"map to positions", func() error { _, err := m.ToPositions(grid34Pickup()); return err }},
		{"map to measures", func() error { _, err := m.ToMeasures(grid34Pickup()); return err }},
		{"map timeline", func() error { _, err := m.EqualTimeline(m); return err }},
		{"map marshal", func() error { _, err := m.MarshalText(); return err }},
		{"state current", func() error { _, err := state.Current(); return err }},
		{"state main", func() error { _, err := state.Main(); return err }},
		{"state deviating", func() error { _, err := state.Deviating(); return err }},
		{"segment change", func() error { _, err := segment.Change(); return err }},
		{"segment start", func() error { _, err := segment.Start(); return err }},
		{"segment start speed", func() error { _, err := segment.StartSpeed(); return err }},
		{"segment end", func() error { _, _, err := segment.End(); return err }},
		{"segment end speed", func() error { _, _, err := segment.EndSpeed(); return err }},
		{"segment seconds", func() error { _, _, err := segment.Seconds(0, rounding.HalfUp); return err }},
		{"seconds value", func() error { _, err := seconds.Value(); return err }},
		{"seconds approximate", func() error { _, err := seconds.Approximate(); return err }},
		{"seconds places", func() error { _, _, err := seconds.Places(); return err }},
		{"seconds bound", func() error { _, err := seconds.ErrorBound(); return err }},
		{"seconds marshal", func() error { _, err := seconds.MarshalText(); return err }},
		{"position value", func() error { _, err := position.Value(); return err }},
		{"position approximate", func() error { _, err := position.Approximate(); return err }},
		{"position places", func() error { _, _, err := position.Places(); return err }},
		{"position bound", func() error { _, err := position.ErrorBound(); return err }},
		{"position marshal", func() error { _, err := position.MarshalText(); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, operation.call(), tempomap.ErrUnset)
		})
	}
}

func TestErrorResultsAreZeroAndAppendIsUnchanged(t *testing.T) {
	t.Parallel()
	m, err := tempomap.New(tempo.Tempo{})
	assert.Equal(t, tempomap.Map{}, m)
	require.Error(t, err)

	address, err := tempomap.At(duration.Position{})
	assert.Equal(t, tempomap.Address{}, address)
	require.Error(t, err)

	change, err := tempomap.NewMain(tempomap.Address{}, quarter(60))
	assert.Equal(t, tempomap.Change{}, change)
	require.Error(t, err)

	extension, err := tempomap.NewExtension(tempomap.Address{}, sec(1, 1))
	assert.Equal(t, tempomap.Extension{}, extension)
	require.Error(t, err)

	result, err := tempomap.MustNew(quarter(60)).Start(at(1, 1), -1, rounding.HalfUp)
	assert.Equal(t, tempomap.SecondsResult{}, result)
	require.Error(t, err)
	position, err := tempomap.MustNew(quarter(60)).Position(sec(1, 1), -1, rounding.HalfUp)
	assert.Equal(t, tempomap.PositionResult{}, position)
	require.Error(t, err)

	prefix := []byte("prefix")
	appended, err := tempomap.Law(0).AppendText(prefix)
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	assert.Equal(t, prefix, appended)
	assert.Equal(t, "prefix", string(prefix))
}

func TestSentinelsArePairwiseDistinct(t *testing.T) {
	t.Parallel()
	for i, left := range allSentinels {
		require.NotNil(t, left)
		assert.True(t, errors.Is(left, left))
		for j, right := range allSentinels {
			if i != j {
				assert.False(t, errors.Is(left, right), "%v wraps %v", left, right)
			}
		}
	}
}

func TestMustVariantsPanicWithTheBaseError(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		base func() error
		must func()
		own  error
	}{
		{"At", func() error { _, err := tempomap.At(duration.Position{}); return err }, func() { tempomap.MustAt(duration.Position{}) }, tempomap.ErrInvalidArgument},
		{"AtBar", func() error { _, err := tempomap.AtBar(measure.Address{}); return err }, func() { tempomap.MustAtBar(measure.Address{}) }, tempomap.ErrInvalidArgument},
		{"ParseLaw", func() error { _, err := tempomap.ParseLaw("bad"); return err }, func() { tempomap.MustParseLaw("bad") }, tempomap.ErrInvalidText},
		{"ParseAddress", func() error { _, err := tempomap.ParseAddress("bad"); return err }, func() { tempomap.MustParseAddress("bad") }, tempomap.ErrInvalidText},
		{"NewMain", func() error { _, err := tempomap.NewMain(tempomap.Address{}, quarter(60)); return err }, func() { tempomap.MustNewMain(tempomap.Address{}, quarter(60)) }, tempomap.ErrUnset},
		{"NewJump", func() error { _, err := tempomap.NewJump(tempomap.Address{}, quarter(60)); return err }, func() { tempomap.MustNewJump(tempomap.Address{}, quarter(60)) }, tempomap.ErrUnset},
		{"NewGradual", func() error { _, err := tempomap.NewGradual(tempomap.Address{}, tempomap.LinearSpeed); return err }, func() { tempomap.MustNewGradual(tempomap.Address{}, tempomap.LinearSpeed) }, tempomap.ErrUnset},
		{"NewATempo", func() error { _, err := tempomap.NewATempo(tempomap.Address{}); return err }, func() { tempomap.MustNewATempo(tempomap.Address{}) }, tempomap.ErrUnset},
		{"NewTempoI", func() error { _, err := tempomap.NewTempoI(tempomap.Address{}); return err }, func() { tempomap.MustNewTempoI(tempomap.Address{}) }, tempomap.ErrUnset},
		{"NewModulation", func() error {
			_, err := tempomap.NewModulation(tempomap.Address{}, duration.Quarter.MustWritten(), duration.MustNew(1, 4))
			return err
		}, func() {
			tempomap.MustNewModulation(tempomap.Address{}, duration.Quarter.MustWritten(), duration.MustNew(1, 4))
		}, tempomap.ErrUnset},
		{"NewExtension", func() error { _, err := tempomap.NewExtension(tempomap.Address{}, sec(1, 1)); return err }, func() { tempomap.MustNewExtension(tempomap.Address{}, sec(1, 1)) }, tempomap.ErrUnset},
		{"ParseChange", func() error { _, err := tempomap.ParseChange("bad"); return err }, func() { tempomap.MustParseChange("bad") }, tempomap.ErrInvalidText},
		{"ParseExtension", func() error { _, err := tempomap.ParseExtension("bad"); return err }, func() { tempomap.MustParseExtension("bad") }, tempomap.ErrInvalidText},
		{"New", func() error { _, err := tempomap.New(tempo.Tempo{}); return err }, func() { tempomap.MustNew(tempo.Tempo{}) }, tempomap.ErrInvalidArgument},
		{"NewMeasured", func() error { _, err := tempomap.NewMeasured(measure.Grid{}, quarter(60)); return err }, func() { tempomap.MustNewMeasured(measure.Grid{}, quarter(60)) }, tempomap.ErrInvalidArgument},
		{"Parse", func() error { _, err := tempomap.Parse("bad"); return err }, func() { tempomap.MustParse("bad") }, tempomap.ErrInvalidText},
		{"ParseSecondsResult", func() error { _, err := tempomap.ParseSecondsResult("bad"); return err }, func() { tempomap.MustParseSecondsResult("bad") }, tempomap.ErrInvalidText},
		{"ParsePositionResult", func() error { _, err := tempomap.ParsePositionResult("bad"); return err }, func() { tempomap.MustParsePositionResult("bad") }, tempomap.ErrInvalidText},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expected := tt.base()
			require.Error(t, expected)
			requirePanicError(t, expected, tt.own, tt.must)
		})
	}

	change := tempomap.MustNewMain(at(1, 1), quarter(60))
	expected := func() error { _, err := change.WithSlot(tempomap.Slot(0)); return err }()
	requirePanicError(t, expected, tempomap.ErrInvalidArgument, func() { change.MustWithSlot(tempomap.Slot(0)) })
}

func TestMustMapMethodsPanicWithTheBaseError(t *testing.T) {
	t.Parallel()
	var m tempomap.Map
	for _, tt := range []struct {
		name string
		base func() error
		must func()
	}{
		{"WithChange", func() error { _, err := m.WithChange(tempomap.Change{}); return err }, func() { m.MustWithChange(tempomap.Change{}) }},
		{"WithoutChange", func() error { _, err := m.WithoutChange(tempomap.Address{}, tempomap.SlotSingle); return err }, func() { m.MustWithoutChange(tempomap.Address{}, tempomap.SlotSingle) }},
		{"WithExtension", func() error { _, err := m.WithExtension(tempomap.Extension{}); return err }, func() { m.MustWithExtension(tempomap.Extension{}) }},
		{"WithoutExtension", func() error { _, err := m.WithoutExtension(tempomap.Address{}); return err }, func() { m.MustWithoutExtension(tempomap.Address{}) }},
		{"Validate", func() error { _, err := m.Validate(); return err }, func() { m.MustValidate() }},
		{"State", func() error { _, err := m.State(tempomap.Address{}); return err }, func() { m.MustState(tempomap.Address{}) }},
		{"Speed", func() error { _, err := m.Speed(tempomap.Address{}); return err }, func() { m.MustSpeed(tempomap.Address{}) }},
		{"TempoAt", func() error { _, err := m.TempoAt(tempomap.Address{}, duration.Quarter.MustWritten()); return err }, func() { m.MustTempoAt(tempomap.Address{}, duration.Quarter.MustWritten()) }},
		{"Segments", func() error { _, err := m.Segments(); return err }, func() { m.MustSegments() }},
		{"Start", func() error { _, err := m.Start(tempomap.Address{}, 0, rounding.HalfUp); return err }, func() { m.MustStart(tempomap.Address{}, 0, rounding.HalfUp) }},
		{"End", func() error { _, err := m.End(tempomap.Address{}, 0, rounding.HalfUp); return err }, func() { m.MustEnd(tempomap.Address{}, 0, rounding.HalfUp) }},
		{"Position", func() error { _, err := m.Position(sec(1, 1), 0, rounding.HalfUp); return err }, func() { m.MustPosition(sec(1, 1), 0, rounding.HalfUp) }},
		{"AddressAt", func() error { _, _, err := m.AddressAt(grid34Pickup(), sec(1, 1), 0, rounding.HalfUp); return err }, func() { m.MustAddressAt(grid34Pickup(), sec(1, 1), 0, rounding.HalfUp) }},
		{"Between", func() error {
			_, err := m.Between(tempomap.Address{}, tempomap.Address{}, 0, rounding.HalfUp)
			return err
		}, func() { m.MustBetween(tempomap.Address{}, tempomap.Address{}, 0, rounding.HalfUp) }},
		{"ToPositions", func() error { _, err := m.ToPositions(grid34Pickup()); return err }, func() { m.MustToPositions(grid34Pickup()) }},
		{"ToMeasures", func() error { _, err := m.ToMeasures(grid34Pickup()); return err }, func() { m.MustToMeasures(grid34Pickup()) }},
		{"EqualTimeline", func() error { _, err := m.EqualTimeline(m); return err }, func() { m.MustEqualTimeline(m) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expected := tt.base()
			requireTempomapError(t, expected, tempomap.ErrUnset)
			requirePanicError(t, expected, tempomap.ErrUnset, tt.must)
		})
	}

	var segment tempomap.Segment
	expected := func() error { _, _, err := segment.Seconds(0, rounding.HalfUp); return err }()
	requirePanicError(t, expected, tempomap.ErrUnset, func() { segment.MustSeconds(0, rounding.HalfUp) })
}
