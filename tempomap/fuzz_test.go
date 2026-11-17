package tempomap_test

import (
	"testing"

	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzParseAddressRoundTrip(f *testing.F) {
	for _, seed := range []string{"p:0/1", "p:3/8", "m:2:1/8", "Address(unset)", "bad", ""} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		value, err := tempomap.ParseAddress(text)
		if err != nil {
			return
		}
		roundTrip, err := tempomap.ParseAddress(value.String())
		require.NoError(t, err)
		assert.True(t, value.Equal(roundTrip))
	})
}

func FuzzParseChangeRoundTrip(f *testing.F) {
	for _, seed := range []string{
		"first@p:0/1:main:4=60",
		"single@p:2/1:jump:4=120",
		"second@p:0/1:gradual:linear-speed",
		"single@p:3/1:a-tempo",
		"single@p:4/1:tempo-i",
		"single@p:1/1:modulation:4.@1/4",
		"bad",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		value, err := tempomap.ParseChange(text)
		if err != nil {
			return
		}
		roundTrip, err := tempomap.ParseChange(value.String())
		require.NoError(t, err)
		assert.True(t, value.Equal(roundTrip))
	})
}

func FuzzParseMapRoundTrip(f *testing.F) {
	for _, seed := range []string{
		"tempomap[first@p:0/1:main:4=60]",
		"tempomap[first@p:0/1:main:4=60;hold@p:1/4:0/1]",
		"tempomap[first@m:1:0/1:main:4=60]",
		"tempomap[]",
		"bad",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		value, err := tempomap.Parse(text)
		if err != nil {
			return
		}
		roundTrip, err := tempomap.Parse(value.String())
		require.NoError(t, err)
		assert.True(t, value.Equal(roundTrip))
	})
}

func FuzzParseResultsRoundTrip(f *testing.F) {
	for _, seed := range []string{"s:10/1", "s~2773/1000@3:1/2000", "p:3/1", "p~284/1000@3:1/2000", "bad"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if seconds, err := tempomap.ParseSecondsResult(text); err == nil {
			roundTrip, err := tempomap.ParseSecondsResult(seconds.String())
			require.NoError(t, err)
			assert.True(t, seconds.Equal(roundTrip))
		}
		if position, err := tempomap.ParsePositionResult(text); err == nil {
			roundTrip, err := tempomap.ParsePositionResult(position.String())
			require.NoError(t, err)
			assert.True(t, position.Equal(roundTrip))
		}
	})
}
