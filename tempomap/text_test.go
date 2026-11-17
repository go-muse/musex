package tempomap_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressTextRoundTrip(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"p:0/1", "p:3/8", "m:2:1/8"} {
		address, err := tempomap.ParseAddress(text)
		require.NoError(t, err)
		assert.Equal(t, text, address.String())
		marshaled, err := address.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, text, string(marshaled))
		appended, err := address.AppendText([]byte("prefix:"))
		require.NoError(t, err)
		assert.Equal(t, "prefix:"+text, string(appended))

		var decoded tempomap.Address
		require.NoError(t, decoded.UnmarshalText(marshaled))
		assert.True(t, address.Equal(decoded))
	}
}

func TestChangeTextRoundTripAllKinds(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"first@p:0/1:main:4=60",
		"single@p:2/1:jump:4=120",
		"second@p:0/1:gradual:linear-speed",
		"single@p:3/1:a-tempo",
		"single@p:4/1:tempo-i",
		"single@p:1/1:modulation:4.@1/4",
	} {
		change, err := tempomap.ParseChange(text)
		require.NoError(t, err)
		assert.Equal(t, text, change.String())
		marshaled, err := change.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, text, string(marshaled))
		var decoded tempomap.Change
		require.NoError(t, decoded.UnmarshalText(marshaled))
		assert.True(t, change.Equal(decoded))
	}
}

func TestExtensionAndLawTextRoundTrip(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"hold@p:1/2:1/1", "hold@p:1/4:0/1", "hold@m:2:1/8:3/2"} {
		extension, err := tempomap.ParseExtension(text)
		require.NoError(t, err)
		assert.Equal(t, text, extension.String())
		marshaled, err := extension.MarshalText()
		require.NoError(t, err)
		var decoded tempomap.Extension
		require.NoError(t, decoded.UnmarshalText(marshaled))
		assert.True(t, extension.Equal(decoded))
	}

	law, err := tempomap.ParseLaw("linear-speed")
	require.NoError(t, err)
	assert.Equal(t, tempomap.LinearSpeed, law)
	marshaled, err := law.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "linear-speed", string(marshaled))
	appended, err := law.AppendText([]byte("law="))
	require.NoError(t, err)
	assert.Equal(t, "law=linear-speed", string(appended))
	var decoded tempomap.Law
	require.NoError(t, decoded.UnmarshalText(marshaled))
	assert.Equal(t, law, decoded)
}

func TestKP40MapTextRoundTripAllChangesAndExtensions(t *testing.T) {
	t.Parallel()
	text := "tempomap[" + strings.Join([]string{
		"first@p:0/1:main:4=60",
		"second@p:0/1:gradual:linear-speed",
		"single@p:1/1:main:4=120",
		"single@p:2/1:jump:4=90",
		"single@p:3/1:a-tempo",
		"single@p:4/1:tempo-i",
		"single@p:5/1:modulation:4.@1/4",
		"hold@p:1/4:0/1",
		"hold@p:1/2:1/1",
	}, ";") + "]"
	m, err := tempomap.Parse(text)
	require.NoError(t, err)
	assert.Equal(t, text, m.String())
	marshaled, err := m.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, text, string(marshaled))
	appended, err := m.AppendText([]byte("map="))
	require.NoError(t, err)
	assert.Equal(t, "map="+text, string(appended))
	var decoded tempomap.Map
	require.NoError(t, decoded.UnmarshalText(marshaled))
	assert.True(t, m.Equal(decoded))

	measuredText := "tempomap[first@m:1:0/1:main:4=60;hold@m:1:1/2:1/1]"
	measured, err := tempomap.Parse(measuredText)
	require.NoError(t, err)
	assert.Equal(t, measuredText, measured.String())
	unfinished, err := measured.Validate(grid44Finite(2))
	require.NoError(t, err)
	assert.False(t, unfinished)
}

func TestResultTextAndMetadataRoundTrip(t *testing.T) {
	t.Parallel()
	seconds, err := tempomap.ParseSecondsResult("s:10/1")
	require.NoError(t, err)
	checkSecondsResult(t, seconds, "10/1", false, 0, "0/1")
	assert.Equal(t, "s:10/1", seconds.String())
	seconds, err = tempomap.ParseSecondsResult("s~2773/1000@3:1/2000")
	require.NoError(t, err)
	checkSecondsResult(t, seconds, "2773/1000", true, 3, "1/2000")
	assert.Equal(t, "s~2773/1000@3:1/2000", seconds.String())
	marshaled, err := seconds.MarshalText()
	require.NoError(t, err)
	var decodedSeconds tempomap.SecondsResult
	require.NoError(t, decodedSeconds.UnmarshalText(marshaled))
	assert.True(t, seconds.Equal(decodedSeconds))
	appended, err := seconds.AppendText([]byte("result="))
	require.NoError(t, err)
	assert.Equal(t, "result=s~2773/1000@3:1/2000", string(appended))

	position, err := tempomap.ParsePositionResult("p:3/1")
	require.NoError(t, err)
	checkPositionResult(t, position, "3/1", false, 0, "0/1")
	position, err = tempomap.ParsePositionResult("p~284/1000@3:1/2000")
	require.NoError(t, err)
	checkPositionResult(t, position, "284/1000", true, 3, "1/2000")
	assert.Equal(t, "p~284/1000@3:1/2000", position.String())
	marshaled, err = position.MarshalText()
	require.NoError(t, err)
	var decodedPosition tempomap.PositionResult
	require.NoError(t, decodedPosition.UnmarshalText(marshaled))
	assert.True(t, position.Equal(decodedPosition))

	exactA := tempomap.MustNew(quarter(60)).MustStart(at(1, 1), 0, roundingHalfUp())
	exactB := tempomap.MustNew(quarter(60)).MustStart(at(1, 1), 1000, roundingHalfUp())
	assert.True(t, exactA.Equal(exactB))
}

func TestKP32InvalidTextAndSpecialUnsetStrings(t *testing.T) {
	t.Parallel()
	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"address unset", func() error { _, err := tempomap.ParseAddress("Address(unset)"); return err }},
		{"change unset", func() error { _, err := tempomap.ParseChange("Change(unset)"); return err }},
		{"extension unset", func() error { _, err := tempomap.ParseExtension("Extension(unset)"); return err }},
		{"map unset", func() error { _, err := tempomap.Parse("Map(unset)"); return err }},
		{"seconds unset", func() error { _, err := tempomap.ParseSecondsResult("SecondsResult(unset)"); return err }},
		{"position unset", func() error { _, err := tempomap.ParsePositionResult("PositionResult(unset)"); return err }},
		{"law unset", func() error { _, err := tempomap.ParseLaw("Law(invalid)"); return err }},
		{"address grammar", func() error { _, err := tempomap.ParseAddress("P:1/2"); return err }},
		{"change grammar", func() error { _, err := tempomap.ParseChange("single@p:1/1:jump"); return err }},
		{"extension grammar", func() error { _, err := tempomap.ParseExtension("hold@p:1/2"); return err }},
		{"map grammar", func() error { _, err := tempomap.Parse("tempomap[first@p:0/1:main:4=60] "); return err }},
		{"law grammar", func() error { _, err := tempomap.ParseLaw("LINEAR-SPEED"); return err }},
		{"seconds grammar", func() error { _, err := tempomap.ParseSecondsResult("s:1/2 "); return err }},
		{"position grammar", func() error { _, err := tempomap.ParsePositionResult("p:1/2 "); return err }},
		{"seconds value off decimal grid", func() error { _, err := tempomap.ParseSecondsResult("s~1/3@3:1/2000"); return err }},
		{"position value off decimal grid", func() error { _, err := tempomap.ParsePositionResult("p~1/3@3:1/2000"); return err }},
		{"seconds wrong bound", func() error { _, err := tempomap.ParseSecondsResult("s~1/1@3:1/7"); return err }},
		{"position wrong bound", func() error { _, err := tempomap.ParsePositionResult("p~1/1@3:1/7"); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, operation.call(), tempomap.ErrInvalidText)
		})
	}

	_, secondsErr := tempo.ParseSeconds("s~2773/1000@3:1/2000")
	require.Error(t, secondsErr)
	_, positionErr := duration.ParsePosition("p~284/1000@3:1/2000")
	require.Error(t, positionErr)

	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"seconds places above maximum", func() error { _, err := tempomap.ParseSecondsResult("s~1/1@1001:1/1"); return err }},
		{"position places above maximum", func() error { _, err := tempomap.ParsePositionResult("p~1/1@1001:1/1"); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, operation.call(), tempomap.ErrInvalidArgument)
		})
	}

	_, err := tempomap.ParseSecondsResult("s~1/1@77:1/2000000000000000000000000000000000000000000000000000000000000000000000000000000")
	requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrOutOfRange)
}

func TestUnmarshalFailurePreservesReceiverAndNilReceiverIsInvalid(t *testing.T) {
	t.Parallel()
	address := at(1, 2)
	require.Error(t, address.UnmarshalText([]byte("bad")))
	assert.True(t, address.Equal(at(1, 2)))
	change := tempomap.MustNewMain(at(0, 1), quarter(60))
	require.Error(t, change.UnmarshalText([]byte("bad")))
	assert.True(t, change.Equal(tempomap.MustNewMain(at(0, 1), quarter(60))))
	extension := tempomap.MustNewExtension(at(1, 2), sec(1, 1))
	require.Error(t, extension.UnmarshalText([]byte("bad")))
	assert.True(t, extension.Equal(tempomap.MustNewExtension(at(1, 2), sec(1, 1))))
	m := tempomap.MustNew(quarter(60))
	require.Error(t, m.UnmarshalText([]byte("bad")))
	assert.True(t, m.Equal(tempomap.MustNew(quarter(60))))
	law := tempomap.LinearSpeed
	require.Error(t, law.UnmarshalText([]byte("bad")))
	assert.Equal(t, tempomap.LinearSpeed, law)
	seconds := tempomap.MustParseSecondsResult("s:1/1")
	require.Error(t, seconds.UnmarshalText([]byte("bad")))
	assert.Equal(t, "s:1/1", seconds.String())
	position := tempomap.MustParsePositionResult("p:1/1")
	require.Error(t, position.UnmarshalText([]byte("bad")))
	assert.Equal(t, "p:1/1", position.String())

	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"address", func() error { var value *tempomap.Address; return value.UnmarshalText([]byte("p:0/1")) }},
		{"change", func() error { var value *tempomap.Change; return value.UnmarshalText([]byte("bad")) }},
		{"extension", func() error { var value *tempomap.Extension; return value.UnmarshalText([]byte("bad")) }},
		{"map", func() error { var value *tempomap.Map; return value.UnmarshalText([]byte("bad")) }},
		{"law", func() error { var value *tempomap.Law; return value.UnmarshalText([]byte("bad")) }},
		{"seconds", func() error { var value *tempomap.SecondsResult; return value.UnmarshalText([]byte("bad")) }},
		{"position", func() error { var value *tempomap.PositionResult; return value.UnmarshalText([]byte("bad")) }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			requireTempomapError(t, operation.call(), tempomap.ErrInvalidArgument)
		})
	}
}

func TestParseErrorExcerptIsBounded(t *testing.T) {
	t.Parallel()
	short := "bad"
	long := strings.Repeat("x", 1<<20)
	unicodeInput := strings.Repeat("я", 40)
	for _, operation := range []struct {
		name string
		call func(string) error
	}{
		{"address", func(text string) error { _, err := tempomap.ParseAddress(text); return err }},
		{"change", func(text string) error { _, err := tempomap.ParseChange(text); return err }},
		{"extension", func(text string) error { _, err := tempomap.ParseExtension(text); return err }},
		{"map", func(text string) error { _, err := tempomap.Parse(text); return err }},
		{"law", func(text string) error { _, err := tempomap.ParseLaw(text); return err }},
		{"seconds", func(text string) error { _, err := tempomap.ParseSecondsResult(text); return err }},
		{"position", func(text string) error { _, err := tempomap.ParsePositionResult(text); return err }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			err := operation.call(short)
			requireTempomapError(t, err, tempomap.ErrInvalidText)
			assert.Contains(t, err.Error(), strconv.Quote(short))
			err = operation.call(long)
			requireTempomapError(t, err, tempomap.ErrInvalidText)
			assert.LessOrEqual(t, len(err.Error()), 512)
			assert.Contains(t, err.Error(), "...")
			assert.NotContains(t, err.Error(), long)
			err = operation.call(unicodeInput)
			requireTempomapError(t, err, tempomap.ErrInvalidText)
			assert.Contains(t, err.Error(), strconv.Quote(strings.Repeat("я", 32))+"...")
			assert.LessOrEqual(t, len(err.Error()), 512)
		})
	}
}

func roundingHalfUp() (mode rounding.Mode) {
	return rounding.HalfUp
}
