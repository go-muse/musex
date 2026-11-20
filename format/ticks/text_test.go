package ticks_test

import (
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type textValue[T any] interface {
	Equal(T) bool
	String() string
	MarshalText() ([]byte, error)
	AppendText([]byte) ([]byte, error)
}

type textPointer[T any] interface {
	*T
	UnmarshalText([]byte) error
}

// checkCanonicalText checks the canonical text, Parse, MarshalText,
// AppendText and UnmarshalText of one value (4.15, обратимость).
func checkCanonicalText[T textValue[T], P textPointer[T]](t *testing.T, value T, text string, parse func(string) (T, error)) {
	t.Helper()
	assert.Equal(t, text, value.String())
	marshaled, err := value.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, text, string(marshaled))
	appended, err := value.AppendText([]byte("x="))
	require.NoError(t, err)
	assert.Equal(t, "x="+text, string(appended))
	parsed, err := parse(text)
	require.NoError(t, err)
	assert.True(t, value.Equal(parsed), "parsed %s", parsed.String())
	var u T
	require.NoError(t, P(&u).UnmarshalText(marshaled))
	assert.True(t, value.Equal(u), "unmarshaled %s", u.String())
}

// checkParsedAs checks that a non-canonical text is accepted and canonicalized.
func checkParsedAs[T textValue[T]](t *testing.T, text, canonical string, parse func(string) (T, error)) {
	t.Helper()
	parsed, err := parse(text)
	require.NoError(t, err, text)
	assert.Equal(t, canonical, parsed.String(), text)
}

func TestCanonicalText(t *testing.T) {
	t.Parallel()
	t.Run("MIDIQuarter", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.MIDIQuarter](t, ticks.StandardQuarter(), "1/4", ticks.ParseMIDIQuarter)
		checkCanonicalText[ticks.MIDIQuarter](t, ticks.MustNewMIDIQuarter(16), "1/2", ticks.ParseMIDIQuarter)
		checkCanonicalText[ticks.MIDIQuarter](t, ticks.MustNewMIDIQuarter(12), "3/8", ticks.ParseMIDIQuarter)
		checkCanonicalText[ticks.MIDIQuarter](t, ticks.MustNewMIDIQuarter(255), "255/32", ticks.ParseMIDIQuarter)
		checkParsedAs(t, "2/8", "1/4", ticks.ParseMIDIQuarter)
		checkParsedAs(t, "16/32", "1/2", ticks.ParseMIDIQuarter)
		assert.True(t, ticks.MustParseMIDIQuarter("1/4").MustIsStandard())
	})
	t.Run("Resolution", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.Resolution](t, res(480), "480", ticks.ParseResolution)
		checkCanonicalText[ticks.Resolution](t, resQ(480, 16), "480@1/2", ticks.ParseResolution)
		checkCanonicalText[ticks.Resolution](t, resQ(1, 1), "1@1/32", ticks.ParseResolution)
		checkParsedAs(t, "480@1/4", "480", ticks.ParseResolution)
		checkParsedAs(t, "480@2/4", "480@1/2", ticks.ParseResolution)
		assert.True(t, ticks.MustParseResolution("96").Equal(res(96)))
	})
	t.Run("SMPTE", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.SMPTE](t, smpte(-25, 40), "-25x40", ticks.ParseSMPTE)
		checkCanonicalText[ticks.SMPTE](t, smpte(-29, 80), "-29x80", ticks.ParseSMPTE)
		checkCanonicalText[ticks.SMPTE](t, smpte(-29, 80).MustWithReading(ticks.ReadingDropFrame), "-29x80:df", ticks.ParseSMPTE)
		checkCanonicalText[ticks.SMPTE](t, smpte(-29, 80).MustWithReading(ticks.Reading30), "-29x80:30", ticks.ParseSMPTE)
		checkCanonicalText[ticks.SMPTE](t, smpte(-29, 80).MustWithReading(ticks.Reading29), "-29x80:29", ticks.ParseSMPTE)
		checkCanonicalText[ticks.SMPTE](t, smpte(-24, 1), "-24x1", ticks.ParseSMPTE)
		checkCanonicalText[ticks.SMPTE](t, smpte(-30, 255), "-30x255", ticks.ParseSMPTE)
		assert.True(t, ticks.MustParseSMPTE("-25x40").Equal(smpte(-25, 40)))
	})
	t.Run("MIDITempo", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.MIDITempo](t, mtempo(500000), "500000", ticks.ParseMIDITempo)
		checkCanonicalText[ticks.MIDITempo](t, mtempo(1), "1", ticks.ParseMIDITempo)
		checkCanonicalText[ticks.MIDITempo](t, mtempo(16777215), "16777215", ticks.ParseMIDITempo)
		assert.True(t, ticks.MustParseMIDITempo("600000").Equal(mtempo(600000)))
	})
	t.Run("MIDITempoChange", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.MIDITempoChange](t, tchange(192, 1000000), "192:1000000", ticks.ParseMIDITempoChange)
		checkCanonicalText[ticks.MIDITempoChange](t, tchange(0, 500000), "0:500000", ticks.ParseMIDITempoChange)
		assert.True(t, ticks.MustParseMIDITempoChange("5:7").Equal(tchange(5, 7)))
	})
	t.Run("MIDITempoMap", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.MIDITempoMap](t, kp10Map(), "miditempo[96;500000;0:500000;192:1000000]", ticks.ParseMIDITempoMap)
		checkCanonicalText[ticks.MIDITempoMap](t, ticks.MustNewMIDITempoMap(res(96)).MustWithInitial(mtempo(600000)), "miditempo[96;600000]", ticks.ParseMIDITempoMap)
		checkCanonicalText[ticks.MIDITempoMap](t, ticks.MustNewMIDITempoMap(resQ(480, 16), tchange(0, 1), tchange(0, 2)), "miditempo[480@1/2;500000;0:1;0:2]", ticks.ParseMIDITempoMap)
		assert.True(t, ticks.MustParseMIDITempoMap("miditempo[96;500000]").Equal(ticks.MustNewMIDITempoMap(res(96))))
	})
	t.Run("MIDIMeter", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.MIDIMeter](t, mmeter(6, 3, 36, 8), "6,3,36,8", ticks.ParseMIDIMeter)
		checkCanonicalText[ticks.MIDIMeter](t, mmeter(1, 254, 255, 255), "1,254,255,255", ticks.ParseMIDIMeter)
		checkCanonicalText[ticks.MIDIMeter](t, mmeter(2, 0, 96, 8), "2,0,96,8", ticks.ParseMIDIMeter)
		assert.True(t, ticks.MustParseMIDIMeter("4,2,24,8").Equal(ticks.DefaultMIDIMeter()))
	})
	t.Run("MIDIMeterChange", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.MIDIMeterChange](t, mchange(0, 4, 2, 24, 8), "0:4,2,24,8", ticks.ParseMIDIMeterChange)
		assert.True(t, ticks.MustParseMIDIMeterChange("3840:3,2,24,8").Equal(mchange(3840, 3, 2, 24, 8)))
	})
	t.Run("MIDIMeterMap", func(t *testing.T) {
		t.Parallel()
		m := ticks.MustNewMIDIMeterMap(res(480), mchange(0, 4, 2, 24, 8), mchange(3840, 3, 2, 24, 8))
		checkCanonicalText[ticks.MIDIMeterMap](t, m, "midimeter[480;0:4,2,24,8;3840:3,2,24,8]", ticks.ParseMIDIMeterMap)
		checkCanonicalText[ticks.MIDIMeterMap](t, ticks.MustNewMIDIMeterMap(res(96)), "midimeter[96]", ticks.ParseMIDIMeterMap)
		checkCanonicalText[ticks.MIDIMeterMap](t, ticks.MustNewMIDIMeterMap(resQ(96, 16), mchange(0, 4, 2, 24, 16), mchange(0, 3, 2, 24, 16)), "midimeter[96@1/2;0:4,2,24,16;0:3,2,24,16]", ticks.ParseMIDIMeterMap)
		assert.True(t, ticks.MustParseMIDIMeterMap("midimeter[480;0:4,2,24,8;3840:3,2,24,8]").Equal(m))
	})
	t.Run("Click", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.Click](t, ticks.ClickOnBeat(), "beat", ticks.ParseClick)
		checkCanonicalText[ticks.Click](t, ticks.MustClickEvery(dur(3, 8)), "3/8", ticks.ParseClick)
		checkParsedAs(t, "6/16", "3/8", ticks.ParseClick)
		assert.True(t, ticks.MustParseClick("beat").Equal(ticks.ClickOnBeat()))
	})
	t.Run("Quantized", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.Quantized](t, res(480).MustQuantize(dur(1, 28), rounding.HalfEven), "69+1/4480", ticks.ParseQuantized)
		checkCanonicalText[ticks.Quantized](t, res(480).MustQuantize(dur(1, 28), rounding.Down), "68-1/3360", ticks.ParseQuantized)
		checkCanonicalText[ticks.Quantized](t, res(480).MustQuantize(dur(3, 8), rounding.Down), "720+0/1", ticks.ParseQuantized)
		checkCanonicalText[ticks.Quantized](t, res(1).MustQuantizePosition(pos(1, 64), rounding.Down), "0-1/64", ticks.ParseQuantized)
		checkParsedAs(t, "69+2/8960", "69+1/4480", ticks.ParseQuantized)
		checkParsedAs(t, "720-0/1", "720+0/1", ticks.ParseQuantized)
		checkParsedAs(t, "720+0/7", "720+0/1", ticks.ParseQuantized)
		q := ticks.MustParseQuantized("68-1/3360")
		assert.Equal(t, 68, q.MustTicks())
		assertFraction(t, frac(-1, 3360), q.MustDeviation())
		assert.True(t, ticks.MustParseQuantized("720-0/1").Equal(ticks.MustParseQuantized("720+0/1")))
	})
	t.Run("BarLoss", func(t *testing.T) {
		t.Parallel()
		checkCanonicalText[ticks.BarLoss](t, ticks.BarLoss{Bar: 3, Losses: []ticks.Loss{ticks.LostSymbol, ticks.LostGrouping}}, "3:symbol,grouping", ticks.ParseBarLoss)
		checkCanonicalText[ticks.BarLoss](t, ticks.BarLoss{Bar: 3}, "3:", ticks.ParseBarLoss)
		checkCanonicalText[ticks.BarLoss](t, ticks.BarLoss{Bar: 0, Losses: []ticks.Loss{ticks.LostParts, ticks.LostParts}}, "0:parts,parts", ticks.ParseBarLoss)
		checkCanonicalText[ticks.BarLoss](t, ticks.BarLoss{Bar: 2, Losses: []ticks.Loss{ticks.LostGrouping, ticks.LostSymbol}}, "2:grouping,symbol", ticks.ParseBarLoss)
		parsed := ticks.MustParseBarLoss("2:grouping,symbol")
		assert.Equal(t, []ticks.Loss{ticks.LostGrouping, ticks.LostSymbol}, parsed.Losses)
		assert.Equal(t, 2, parsed.Bar)
	})
	t.Run("ExportOptions", func(t *testing.T) {
		t.Parallel()
		o := ticks.MustNewExportOptions(rounding.HalfUp)
		checkCanonicalText[ticks.ExportOptions](t, o, "export[tempo=half-up;positions=exact;steps=none;clock=false;holds=refuse]", ticks.ParseExportOptions)
		checkCanonicalText[ticks.ExportOptions](t, o.MustWithSteps(dur(1, 4), ticks.StepEqualTime).MustWithHoldsStretched(dur(1, 4)),
			"export[tempo=half-up;positions=exact;steps=1/4:equal-time;clock=false;holds=stretch:1/4]", ticks.ParseExportOptions)
		checkCanonicalText[ticks.ExportOptions](t, ticks.MustNewExportOptions(rounding.Down).MustWithPositionMode(rounding.HalfEven).MustWithClockSteps(true).MustWithHoldsSkipped(),
			"export[tempo=down;positions=half-even;steps=none;clock=true;holds=skip]", ticks.ParseExportOptions)
		checkCanonicalText[ticks.ExportOptions](t, ticks.MustNewExportOptions(rounding.HalfTowardZero).MustWithSteps(dur(3, 8), ticks.StepStart),
			"export[tempo=half-toward-zero;positions=exact;steps=3/8:start;clock=false;holds=refuse]", ticks.ParseExportOptions)
		checkCanonicalText[ticks.ExportOptions](t, ticks.MustNewExportOptions(rounding.AwayFromZero).MustWithSteps(dur(1, 8), ticks.StepEnd),
			"export[tempo=away-from-zero;positions=exact;steps=1/8:end;clock=false;holds=refuse]", ticks.ParseExportOptions)
		for _, mode := range allModes {
			name, err := mode.Name()
			require.NoError(t, err)
			text := "export[tempo=" + name + ";positions=" + name + ";steps=none;clock=false;holds=refuse]"
			checkCanonicalText[ticks.ExportOptions](t, ticks.MustNewExportOptions(mode).MustWithPositionMode(mode), text, ticks.ParseExportOptions)
		}
		checkParsedAs(t, "export[tempo=half-up;positions=exact;steps=2/8:start;clock=false;holds=stretch:2/4]",
			"export[tempo=half-up;positions=exact;steps=1/4:start;clock=false;holds=stretch:1/2]", ticks.ParseExportOptions)
	})
}

// TestEnumerationText checks the Parse functions of the enumerations (4.15).
func TestEnumerationText(t *testing.T) {
	t.Parallel()
	for _, v := range []ticks.FrameRate{ticks.FrameRate24, ticks.FrameRate25, ticks.FrameRate29, ticks.FrameRate30} {
		got, err := ticks.ParseFrameRate(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseFrameRate(v.String()))
	}
	for _, v := range []ticks.Reading{ticks.ReadingDropFrame, ticks.Reading30, ticks.Reading29} {
		got, err := ticks.ParseReading(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseReading(v.String()))
	}
	for _, v := range []ticks.StepMethod{ticks.StepStart, ticks.StepEnd, ticks.StepEqualTime} {
		got, err := ticks.ParseStepMethod(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseStepMethod(v.String()))
	}
	for _, v := range []ticks.HoldPolicy{ticks.HoldsRefuse, ticks.HoldsSkip, ticks.HoldsStretch} {
		got, err := ticks.ParseHoldPolicy(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseHoldPolicy(v.String()))
	}
	for _, v := range []ticks.Fidelity{ticks.Strict, ticks.Lossy} {
		got, err := ticks.ParseFidelity(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseFidelity(v.String()))
	}
	for _, v := range []ticks.Loss{ticks.LostSymbol, ticks.LostGrouping, ticks.LostParts} {
		got, err := ticks.ParseLoss(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseLoss(v.String()))
	}
	for _, v := range []ticks.IrregularBars{ticks.IrregularEvent, ticks.IrregularSkip} {
		got, err := ticks.ParseIrregularBars(v.String())
		require.NoError(t, err)
		assert.Equal(t, v, got)
		assert.Equal(t, v, ticks.MustParseIrregularBars(v.String()))
	}
	for name, call := range map[string]func() error{
		"frame rate -28":          func() error { _, err := ticks.ParseFrameRate("-28"); return err },
		"frame rate 25":           func() error { _, err := ticks.ParseFrameRate("25"); return err },
		"frame rate invalid":      func() error { _, err := ticks.ParseFrameRate("FrameRate(invalid)"); return err },
		"reading DF":              func() error { _, err := ticks.ParseReading("DF"); return err },
		"reading invalid":         func() error { _, err := ticks.ParseReading("Reading(invalid)"); return err },
		"step method equal time":  func() error { _, err := ticks.ParseStepMethod("equal time"); return err },
		"step method invalid":     func() error { _, err := ticks.ParseStepMethod("StepMethod(invalid)"); return err },
		"hold policy stretch:1/4": func() error { _, err := ticks.ParseHoldPolicy("stretch:1/4"); return err },
		"hold policy invalid":     func() error { _, err := ticks.ParseHoldPolicy("HoldPolicy(invalid)"); return err },
		"fidelity empty":          func() error { _, err := ticks.ParseFidelity(""); return err },
		"fidelity invalid":        func() error { _, err := ticks.ParseFidelity("Fidelity(invalid)"); return err },
		"loss Symbol":             func() error { _, err := ticks.ParseLoss("Symbol"); return err },
		"loss invalid":            func() error { _, err := ticks.ParseLoss("Loss(invalid)"); return err },
		"irregular events":        func() error { _, err := ticks.ParseIrregularBars("events"); return err },
		"irregular invalid":       func() error { _, err := ticks.ParseIrregularBars("IrregularBars(invalid)"); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidText)
		})
	}
	_, err := ticks.ParseLoss("bogus")
	requireMustPanic(t, err, ticks.ErrInvalidText, func() { ticks.MustParseLoss("bogus") })
}

// TestParseErrors checks the order and kinds of failures of 4.15.
func TestParseErrors(t *testing.T) {
	t.Parallel()
	hugeDenominator := "1/1" + strings.Repeat("0", 90)
	for _, tt := range []struct {
		name  string
		parse func(string) error
		text  string
		own   error
		deps  []error
	}{
		// Grammar.
		{name: "resolution leading zero", parse: parseResolution, text: "0480", own: ticks.ErrInvalidText},
		{name: "resolution spaces", parse: parseResolution, text: " 480", own: ticks.ErrInvalidText},
		{name: "resolution sign", parse: parseResolution, text: "+480", own: ticks.ErrInvalidText},
		{name: "resolution empty", parse: parseResolution, text: "", own: ticks.ErrInvalidText},
		{name: "resolution unset string", parse: parseResolution, text: "Resolution(unset)", own: ticks.ErrInvalidText},
		{name: "resolution zero denominator", parse: parseResolution, text: "480@1/0", own: ticks.ErrInvalidText},
		{name: "quarter unset string", parse: parseQuarter, text: "MIDIQuarter(unset)", own: ticks.ErrInvalidText},
		{name: "quarter integer", parse: parseQuarter, text: "1", own: ticks.ErrInvalidText},
		{name: "smpte code -28", parse: parseSMPTE, text: "-28x40", own: ticks.ErrInvalidText},
		{name: "smpte positive code", parse: parseSMPTE, text: "25x40", own: ticks.ErrInvalidText},
		{name: "smpte reading DF", parse: parseSMPTE, text: "-29x80:DF", own: ticks.ErrInvalidText},
		{name: "smpte unset string", parse: parseSMPTE, text: "SMPTE(unset)", own: ticks.ErrInvalidText},
		{name: "tempo negative", parse: parseMIDITempo, text: "-1", own: ticks.ErrInvalidText},
		{name: "tempo decimal", parse: parseMIDITempo, text: "500000.0", own: ticks.ErrInvalidText},
		{name: "tempo unset string", parse: parseMIDITempo, text: "MIDITempo(unset)", own: ticks.ErrInvalidText},
		{name: "tempo change without tick", parse: parseMIDITempoChange, text: ":500000", own: ticks.ErrInvalidText},
		{name: "tempo change with unset tempo string", parse: parseMIDITempoChange, text: "5:MIDITempo(unset)", own: ticks.ErrInvalidText},
		{name: "tempo map without initial tempo", parse: parseMIDITempoMap, text: "miditempo[96]", own: ticks.ErrInvalidText},
		{name: "tempo map trailing separator", parse: parseMIDITempoMap, text: "miditempo[96;500000;]", own: ticks.ErrInvalidText},
		{name: "tempo map unset string", parse: parseMIDITempoMap, text: "MIDITempoMap(unset)", own: ticks.ErrInvalidText},
		{name: "meter with three fields", parse: parseMIDIMeter, text: "4,2,24", own: ticks.ErrInvalidText},
		{name: "meter with spaces", parse: parseMIDIMeter, text: "4, 2, 24, 8", own: ticks.ErrInvalidText},
		{name: "meter map with an initial tempo", parse: parseMIDIMeterMap, text: "midimeter[480;500000]", own: ticks.ErrInvalidText},
		{name: "click Beat", parse: parseClick, text: "Beat", own: ticks.ErrInvalidText},
		{name: "click unset string", parse: parseClick, text: "Click(unset)", own: ticks.ErrInvalidText},
		{name: "quantized without sign", parse: parseQuantized, text: "69 1/4480", own: ticks.ErrInvalidText},
		{name: "quantized negative ticks", parse: parseQuantized, text: "-1+0/1", own: ticks.ErrInvalidText},
		{name: "quantized integer deviation", parse: parseQuantized, text: "69+1", own: ticks.ErrInvalidText},
		{name: "bar loss unknown loss", parse: parseBarLoss, text: "3:sym", own: ticks.ErrInvalidText},
		{name: "bar loss negative bar", parse: parseBarLoss, text: "-1:", own: ticks.ErrInvalidText},
		{name: "bar loss trailing comma", parse: parseBarLoss, text: "3:symbol,", own: ticks.ErrInvalidText},
		{name: "export unknown mode", parse: parseExport, text: "export[tempo=sideways;positions=exact;steps=none;clock=false;holds=refuse]", own: ticks.ErrInvalidText},
		{name: "export fields out of order", parse: parseExport, text: "export[positions=exact;tempo=half-up;steps=none;clock=false;holds=refuse]", own: ticks.ErrInvalidText},
		{name: "export unset string", parse: parseExport, text: "ExportOptions(unset)", own: ticks.ErrInvalidText},
		{name: "export stretch without length", parse: parseExport, text: "export[tempo=half-up;positions=exact;steps=none;clock=false;holds=stretch]", own: ticks.ErrInvalidText},
		// Grammar of the whole text is checked before ranges and values.
		{name: "grammar before range", parse: parseMIDITempoMap, text: "miditempo[99999999999999999999;0;x]", own: ticks.ErrInvalidText},
		// Ranges.
		{name: "resolution outside int", parse: parseResolution, text: "99999999999999999999", own: ticks.ErrOutOfRange},
		{name: "range before values", parse: parseMIDITempoMap, text: "miditempo[0;99999999999999999999]", own: ticks.ErrOutOfRange},
		{name: "fraction component outside range", parse: parseQuarter, text: hugeDenominator, own: ticks.ErrOutOfRange, deps: []error{fraction.ErrOutOfRange}},
		{name: "tick outside int", parse: parseMIDITempoChange, text: "99999999999999999999:1", own: ticks.ErrOutOfRange},
		// Values from left to right.
		{name: "resolution zero", parse: parseResolution, text: "0", own: ticks.ErrZero},
		{name: "resolution zero quarter", parse: parseResolution, text: "480@0/1", own: ticks.ErrZero},
		{name: "resolution quarter not a multiple of 1/32", parse: parseResolution, text: "480@1/3", own: ticks.ErrNotRepresentable},
		{name: "resolution quarter above 255/32", parse: parseResolution, text: "480@9/1", own: ticks.ErrFormatLimit},
		{name: "quarter zero", parse: parseQuarter, text: "0/4", own: ticks.ErrZero},
		{name: "quarter 1/64", parse: parseQuarter, text: "1/64", own: ticks.ErrNotRepresentable},
		{name: "smpte zero ticks per frame", parse: parseSMPTE, text: "-25x0", own: ticks.ErrZero},
		{name: "smpte ticks per frame above a byte", parse: parseSMPTE, text: "-25x256", own: ticks.ErrFormatLimit},
		{name: "smpte reading not for -29", parse: parseSMPTE, text: "-25x40:df", own: ticks.ErrInvalidArgument},
		{name: "tempo zero", parse: parseMIDITempo, text: "0", own: ticks.ErrZero},
		{name: "tempo above Set Tempo", parse: parseMIDITempo, text: "16777216", own: ticks.ErrFormatLimit},
		{name: "tempo map decreasing ticks", parse: parseMIDITempoMap, text: "miditempo[96;500000;192:1;0:2]", own: ticks.ErrOrder},
		{name: "tempo map zero resolution", parse: parseMIDITempoMap, text: "miditempo[0;500000]", own: ticks.ErrZero},
		{name: "tempo map values left to right", parse: parseMIDITempoMap, text: "miditempo[0;16777216]", own: ticks.ErrZero},
		{name: "tempo map zero initial tempo", parse: parseMIDITempoMap, text: "miditempo[96;0]", own: ticks.ErrZero},
		{name: "tempo map zero tempo of a change", parse: parseMIDITempoMap, text: "miditempo[96;500000;5:0]", own: ticks.ErrZero},
		{name: "meter zero numerator", parse: parseMIDIMeter, text: "0,2,24,8", own: ticks.ErrZero},
		{name: "meter dd 255", parse: parseMIDIMeter, text: "4,255,24,8", own: ticks.ErrOutOfRange},
		{name: "meter dd 256", parse: parseMIDIMeter, text: "4,256,24,8", own: ticks.ErrFormatLimit},
		{name: "meter bb zero", parse: parseMIDIMeter, text: "4,2,24,0", own: ticks.ErrZero},
		{name: "meter cc above a byte", parse: parseMIDIMeter, text: "4,2,256,8", own: ticks.ErrFormatLimit},
		{name: "meter map decreasing ticks", parse: parseMIDIMeterMap, text: "midimeter[480;5:4,2,24,8;0:4,2,24,8]", own: ticks.ErrOrder},
		{name: "meter map zero field", parse: parseMIDIMeterMap, text: "midimeter[480;0:4,2,0,8]", own: ticks.ErrZero},
		{name: "click zero", parse: parseClick, text: "0/1", own: ticks.ErrInvalidArgument},
		{name: "click zero eighths", parse: parseClick, text: "0/8", own: ticks.ErrInvalidArgument},
		{name: "export zero step", parse: parseExport, text: "export[tempo=half-up;positions=exact;steps=0/1:start;clock=false;holds=refuse]", own: ticks.ErrInvalidArgument},
		{name: "export zero stretch", parse: parseExport, text: "export[tempo=half-up;positions=exact;steps=none;clock=false;holds=stretch:0/4]", own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.parse(tt.text), tt.own, tt.deps...)
		})
	}
}

func parseResolution(text string) error { _, err := ticks.ParseResolution(text); return err }
func parseQuarter(text string) error    { _, err := ticks.ParseMIDIQuarter(text); return err }
func parseSMPTE(text string) error      { _, err := ticks.ParseSMPTE(text); return err }
func parseMIDITempo(text string) error  { _, err := ticks.ParseMIDITempo(text); return err }
func parseMIDITempoChange(text string) error {
	_, err := ticks.ParseMIDITempoChange(text)
	return err
}
func parseMIDITempoMap(text string) error { _, err := ticks.ParseMIDITempoMap(text); return err }
func parseMIDIMeter(text string) error    { _, err := ticks.ParseMIDIMeter(text); return err }
func parseMIDIMeterMap(text string) error { _, err := ticks.ParseMIDIMeterMap(text); return err }
func parseClick(text string) error        { _, err := ticks.ParseClick(text); return err }
func parseQuantized(text string) error    { _, err := ticks.ParseQuantized(text); return err }
func parseBarLoss(text string) error      { _, err := ticks.ParseBarLoss(text); return err }
func parseExport(text string) error       { _, err := ticks.ParseExportOptions(text); return err }

// TestUnmarshalTextFailures checks that UnmarshalText keeps the receiver on
// failure and rejects a nil receiver (4.15).
func TestUnmarshalTextFailures(t *testing.T) {
	t.Parallel()
	r := res(480)
	err := r.UnmarshalText([]byte("0"))
	requireTicksError(t, err, ticks.ErrZero)
	assert.True(t, r.Equal(res(480)))
	m := kp10Map()
	err = m.UnmarshalText([]byte("miditempo[96;500000;192:1;0:2]"))
	requireTicksError(t, err, ticks.ErrOrder)
	assert.True(t, m.Equal(kp10Map()))
	l := ticks.BarLoss{Bar: 1, Losses: []ticks.Loss{ticks.LostParts}}
	err = l.UnmarshalText([]byte("x"))
	requireTicksError(t, err, ticks.ErrInvalidText)
	assert.True(t, l.Equal(ticks.BarLoss{Bar: 1, Losses: []ticks.Loss{ticks.LostParts}}))

	for name, call := range map[string]func() error{
		"MIDIQuarter":     func() error { var p *ticks.MIDIQuarter; return p.UnmarshalText([]byte("1/4")) },
		"Resolution":      func() error { var p *ticks.Resolution; return p.UnmarshalText([]byte("480")) },
		"Quantized":       func() error { var p *ticks.Quantized; return p.UnmarshalText([]byte("1+0/1")) },
		"SMPTE":           func() error { var p *ticks.SMPTE; return p.UnmarshalText([]byte("-25x40")) },
		"MIDITempo":       func() error { var p *ticks.MIDITempo; return p.UnmarshalText([]byte("1")) },
		"MIDITempoChange": func() error { var p *ticks.MIDITempoChange; return p.UnmarshalText([]byte("0:1")) },
		"MIDITempoMap":    func() error { var p *ticks.MIDITempoMap; return p.UnmarshalText([]byte("miditempo[1;1]")) },
		"ExportOptions": func() error {
			var p *ticks.ExportOptions
			return p.UnmarshalText([]byte("export[tempo=down;positions=exact;steps=none;clock=false;holds=refuse]"))
		},
		"Click":           func() error { var p *ticks.Click; return p.UnmarshalText([]byte("beat")) },
		"MIDIMeter":       func() error { var p *ticks.MIDIMeter; return p.UnmarshalText([]byte("4,2,24,8")) },
		"MIDIMeterChange": func() error { var p *ticks.MIDIMeterChange; return p.UnmarshalText([]byte("0:4,2,24,8")) },
		"BarLoss":         func() error { var p *ticks.BarLoss; return p.UnmarshalText([]byte("1:")) },
		"MIDIMeterMap":    func() error { var p *ticks.MIDIMeterMap; return p.UnmarshalText([]byte("midimeter[1]")) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidArgument)
		})
	}
}

// TestRecordsWithInvalidFields checks 4.2 for records with exported fields.
func TestRecordsWithInvalidFields(t *testing.T) {
	t.Parallel()
	unsetTempo := ticks.MIDITempoChange{Tick: 5}
	assert.Equal(t, "5:MIDITempo(unset)", unsetTempo.String())
	_, err := unsetTempo.MarshalText()
	requireTicksError(t, err, ticks.ErrUnset)
	out, err := unsetTempo.AppendText([]byte("x"))
	requireTicksError(t, err, ticks.ErrUnset)
	assert.Equal(t, []byte("x"), out)
	negative := ticks.MIDITempoChange{Tick: -2, Tempo: mtempo(5)}
	assert.Equal(t, "-2:5", negative.String())
	_, err = negative.MarshalText()
	requireTicksError(t, err, ticks.ErrInvalidArgument)

	negativeMeter := ticks.MIDIMeterChange{Tick: -1, Meter: mmeter(4, 2, 24, 8)}
	assert.Equal(t, "-1:4,2,24,8", negativeMeter.String())
	_, err = negativeMeter.MarshalText()
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	_, err = ticks.MIDIMeterChange{Tick: 1}.MarshalText()
	requireTicksError(t, err, ticks.ErrUnset)
	assert.Equal(t, "1:MIDIMeter(unset)", ticks.MIDIMeterChange{Tick: 1}.String())

	assert.Equal(t, "3:", ticks.BarLoss{Bar: 3}.String())
	assert.Equal(t, "-1:symbol", ticks.BarLoss{Bar: -1, Losses: []ticks.Loss{ticks.LostSymbol}}.String())
	_, err = ticks.BarLoss{Bar: -1}.MarshalText()
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	invalidLoss := ticks.BarLoss{Bar: 1, Losses: []ticks.Loss{ticks.Loss(0)}}
	assert.Equal(t, "1:Loss(invalid)", invalidLoss.String())
	_, err = invalidLoss.MarshalText()
	requireTicksError(t, err, ticks.ErrInvalidArgument)

	// Equal compares fields; nil and empty loss lists are equal.
	assert.True(t, ticks.BarLoss{Bar: 3}.Equal(ticks.BarLoss{Bar: 3, Losses: []ticks.Loss{}}))
	assert.False(t, ticks.BarLoss{Bar: 3}.Equal(ticks.BarLoss{Bar: 4}))
	assert.False(t, ticks.BarLoss{Bar: 3, Losses: []ticks.Loss{ticks.LostSymbol}}.Equal(ticks.BarLoss{Bar: 3, Losses: []ticks.Loss{ticks.LostGrouping}}))
	assert.False(t, ticks.BarLoss{Bar: 3, Losses: []ticks.Loss{ticks.LostSymbol}}.Equal(ticks.BarLoss{Bar: 3, Losses: []ticks.Loss{ticks.LostSymbol, ticks.LostSymbol}}))
	assert.True(t, tchange(1, 2).Equal(ticks.MIDITempoChange{Tick: 1, Tempo: mtempo(2)}))
	assert.False(t, tchange(1, 2).Equal(tchange(1, 3)))
	assert.False(t, tchange(1, 2).Equal(tchange(2, 2)))
	assert.True(t, ticks.MIDITempoChange{}.Equal(ticks.MIDITempoChange{}))
	assert.True(t, mchange(0, 4, 2, 24, 8).Equal(mchange(0, 4, 2, 24, 8)))
	assert.False(t, mchange(0, 4, 2, 24, 8).Equal(mchange(1, 4, 2, 24, 8)))
}

// TestParseErrorText checks the limits of parse error texts (4.15).
func TestParseErrorText(t *testing.T) {
	t.Parallel()
	long := "miditempo[" + strings.Repeat("7", 5000) + "x]"
	for name, call := range map[string]func(string) error{
		"Resolution":    parseResolution,
		"MIDITempoMap":  parseMIDITempoMap,
		"MIDIMeterMap":  parseMIDIMeterMap,
		"ExportOptions": parseExport,
		"BarLoss":       parseBarLoss,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := call(long)
			requireTicksError(t, err, ticks.ErrInvalidText)
			assert.LessOrEqual(t, len(err.Error()), 512)
			assert.NotContains(t, err.Error(), strings.Repeat("7", 65))
		})
	}
	_, err := ticks.ParseQuartersPerMinute(long)
	requireTicksError(t, err, ticks.ErrInvalidText)
	assert.LessOrEqual(t, len(err.Error()), 512)
	assert.NotContains(t, err.Error(), strings.Repeat("7", 65))
}
