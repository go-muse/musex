package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fuzzRoundTrip checks for any text: parsing does not panic, a failure wraps a
// sentinel of the package, and an accepted value round-trips through String.
func fuzzRoundTrip[T textValue[T]](f *testing.F, seeds []string, parse func(string) (T, error)) {
	f.Helper()
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		value, err := parse(text)
		if err != nil {
			assert.False(t, errNotTicks(err), "error %v wraps no ticks sentinel", err)
			assert.LessOrEqual(t, len(err.Error()), 512)
			return
		}
		again, err := parse(value.String())
		require.NoError(t, err, "canonical text %q", value.String())
		assert.True(t, value.Equal(again))
		assert.Equal(t, value.String(), again.String())
		marshaled, err := value.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, value.String(), string(marshaled))
	})
}

func FuzzParseMIDIQuarter(f *testing.F) {
	fuzzRoundTrip(f, []string{"1/4", "1/2", "2/8", "0/4", "1/3", "9/1", "1/0", ""}, ticks.ParseMIDIQuarter)
}

func FuzzParseResolution(f *testing.F) {
	fuzzRoundTrip(f, []string{"480", "480@1/2", "480@2/4", "0", "0480", "99999999999999999999", "1@1/32"}, ticks.ParseResolution)
}

func FuzzParseQuantized(f *testing.F) {
	fuzzRoundTrip(f, []string{"69+1/4480", "68-1/3360", "720+0/1", "720-0/1", "69+2/8960", "1+1/0"}, ticks.ParseQuantized)
}

func FuzzParseSMPTE(f *testing.F) {
	fuzzRoundTrip(f, []string{"-25x40", "-29x80:df", "-29x80:30", "-25x40:df", "-28x40", "-30x0", "-24x256"}, ticks.ParseSMPTE)
}

func FuzzParseMIDITempo(f *testing.F) {
	fuzzRoundTrip(f, []string{"500000", "1", "16777215", "16777216", "0", "-1"}, ticks.ParseMIDITempo)
}

func FuzzParseMIDITempoChange(f *testing.F) {
	fuzzRoundTrip(f, []string{"192:1000000", "0:1", ":1", "1:0"}, ticks.ParseMIDITempoChange)
}

func FuzzParseMIDITempoMap(f *testing.F) {
	fuzzRoundTrip(f, []string{
		"miditempo[96;500000;0:500000;192:1000000]",
		"miditempo[480@1/2;500000;0:1;0:2]",
		"miditempo[96;500000;192:1;0:2]",
		"miditempo[96]",
	}, ticks.ParseMIDITempoMap)
}

func FuzzParseExportOptions(f *testing.F) {
	fuzzRoundTrip(f, []string{
		"export[tempo=half-up;positions=exact;steps=1/4:equal-time;clock=false;holds=stretch:1/4]",
		"export[tempo=down;positions=half-even;steps=none;clock=true;holds=skip]",
		"export[tempo=half-up;positions=exact;steps=0/1:start;clock=false;holds=refuse]",
	}, ticks.ParseExportOptions)
}

func FuzzParseClick(f *testing.F) {
	fuzzRoundTrip(f, []string{"beat", "3/8", "0/1", "Beat"}, ticks.ParseClick)
}

func FuzzParseMIDIMeter(f *testing.F) {
	fuzzRoundTrip(f, []string{"6,3,36,8", "4,255,24,8", "0,2,24,8", "1,254,255,255"}, ticks.ParseMIDIMeter)
}

func FuzzParseMIDIMeterChange(f *testing.F) {
	fuzzRoundTrip(f, []string{"0:4,2,24,8", "3840:3,2,24,8", "-1:4,2,24,8"}, ticks.ParseMIDIMeterChange)
}

func FuzzParseBarLoss(f *testing.F) {
	fuzzRoundTrip(f, []string{"3:symbol,grouping", "3:", "0:parts,parts", "3:sym"}, ticks.ParseBarLoss)
}

func FuzzParseMIDIMeterMap(f *testing.F) {
	fuzzRoundTrip(f, []string{"midimeter[480;0:4,2,24,8;3840:3,2,24,8]", "midimeter[96]", "midimeter[480;5:4,2,24,8;0:4,2,24,8]"}, ticks.ParseMIDIMeterMap)
}

// FuzzParseEnumerations checks that enumeration parsers do not panic and
// round-trip accepted values.
func FuzzParseEnumerations(f *testing.F) {
	for _, seed := range []string{"-25", "df", "start", "equal-time", "refuse", "stretch", "strict", "lossy", "symbol", "event", "skip", ""} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if v, err := ticks.ParseFrameRate(text); err == nil {
			assert.Equal(t, v, ticks.MustParseFrameRate(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
		if v, err := ticks.ParseReading(text); err == nil {
			assert.Equal(t, v, ticks.MustParseReading(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
		if v, err := ticks.ParseStepMethod(text); err == nil {
			assert.Equal(t, v, ticks.MustParseStepMethod(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
		if v, err := ticks.ParseHoldPolicy(text); err == nil {
			assert.Equal(t, v, ticks.MustParseHoldPolicy(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
		if v, err := ticks.ParseFidelity(text); err == nil {
			assert.Equal(t, v, ticks.MustParseFidelity(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
		if v, err := ticks.ParseLoss(text); err == nil {
			assert.Equal(t, v, ticks.MustParseLoss(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
		if v, err := ticks.ParseIrregularBars(text); err == nil {
			assert.Equal(t, v, ticks.MustParseIrregularBars(v.String()))
		} else {
			assert.ErrorIs(t, err, ticks.ErrInvalidText)
		}
	})
}

// FuzzParseQuartersPerMinute checks that the parser does not panic and that
// failures wrap a sentinel of the package.
func FuzzParseQuartersPerMinute(f *testing.F) {
	for _, seed := range []string{"90", "85.71", "600/7", "0", "0.0", "090", "1/0", "-1"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		value, err := ticks.ParseQuartersPerMinute(text)
		if err != nil {
			assert.False(t, errNotTicks(err))
			return
		}
		assert.True(t, value.IsValid())
		assert.True(t, value.Greater(fraction.Zero()))
	})
}

// FuzzTicksOfDuration checks ИТ-1: Ticks(Duration(t)) == t and
// PositionTicks(Position(t)) == t for any resolution and t > 0.
func FuzzTicksOfDuration(f *testing.F) {
	f.Add(480, 8, 720)
	f.Add(1, 1, 1)
	f.Add(7, 255, 1<<40)
	f.Add(32767, 16, 3)
	f.Fuzz(func(t *testing.T, perQuarter, thirtySeconds, n int) {
		if perQuarter < 1 || thirtySeconds < 1 || thirtySeconds > 255 || n < 1 {
			t.Skip()
		}
		r := ticks.MustNewResolution(perQuarter).MustWithQuarter(ticks.MustNewMIDIQuarter(thirtySeconds))
		d, err := r.Duration(n)
		if err != nil {
			require.ErrorIs(t, err, ticks.ErrOutOfRange)
			return
		}
		back, err := r.Ticks(d)
		require.NoError(t, err)
		assert.Equal(t, n, back)
		p := r.MustPosition(n)
		assert.Equal(t, n, r.MustPositionTicks(p))
	})
}
