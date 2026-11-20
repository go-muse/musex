package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMIDIQuarter checks 4.3: the written length of the MIDI quarter
// (field bb, w = bb/32), ИТ-1, П-16.
func TestMIDIQuarter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		bb       int
		written  duration.Duration
		standard bool
		text     string
	}{
		{name: "standard bb 8", bb: 8, written: dur(1, 4), standard: true, text: "1/4"},
		{name: "half bb 16", bb: 16, written: dur(1, 2), standard: false, text: "1/2"},
		{name: "dotted quarter bb 12", bb: 12, written: dur(3, 8), standard: false, text: "3/8"},
		{name: "thirty-second bb 1", bb: 1, written: dur(1, 32), standard: false, text: "1/32"},
		{name: "largest bb 255", bb: 255, written: dur(255, 32), standard: false, text: "255/32"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, err := ticks.NewMIDIQuarter(tt.bb)
			require.NoError(t, err)
			assert.True(t, q.IsSet())
			bb, err := q.ThirtySeconds()
			require.NoError(t, err)
			assert.Equal(t, tt.bb, bb)
			assert.Equal(t, tt.bb, q.MustThirtySeconds())
			w, err := q.Duration()
			require.NoError(t, err)
			assertEqualValue(t, tt.written, w)
			assertEqualValue(t, tt.written, q.MustDuration())
			standard, err := q.IsStandard()
			require.NoError(t, err)
			assert.Equal(t, tt.standard, standard)
			assert.Equal(t, tt.standard, q.MustIsStandard())
			assert.Equal(t, tt.text, q.String())

			of, err := ticks.MIDIQuarterOf(tt.written)
			require.NoError(t, err)
			assert.True(t, q.Equal(of))
			assert.True(t, q == of)
			assert.Equal(t, q, ticks.MustMIDIQuarterOf(tt.written))
			assert.Equal(t, q, ticks.MustNewMIDIQuarter(tt.bb))
		})
	}
	assert.True(t, ticks.StandardQuarter().Equal(ticks.MustNewMIDIQuarter(8)))
	assert.True(t, ticks.StandardQuarter().MustIsStandard())
}

func TestMIDIQuarterErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		bb   int
		own  error
	}{
		{name: "zero", bb: 0, own: ticks.ErrZero},
		{name: "negative", bb: -1, own: ticks.ErrInvalidArgument},
		{name: "above byte", bb: 256, own: ticks.ErrFormatLimit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, err := ticks.NewMIDIQuarter(tt.bb)
			requireTicksError(t, err, tt.own)
			assert.Equal(t, ticks.MIDIQuarter{}, q)
			requireMustPanic(t, err, tt.own, func() { ticks.MustNewMIDIQuarter(tt.bb) })
		})
	}
	for _, tt := range []struct {
		name    string
		written duration.Duration
		own     error
		deps    []error
	}{
		{name: "not a multiple of 1/32", written: dur(1, 3), own: ticks.ErrNotRepresentable},
		{name: "1/64", written: dur(1, 64), own: ticks.ErrNotRepresentable},
		{name: "more than 255 thirty-seconds", written: dur(8, 1), own: ticks.ErrFormatLimit},
		{name: "unset duration", written: duration.Duration{}, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, err := ticks.MIDIQuarterOf(tt.written)
			requireTicksError(t, err, tt.own, tt.deps...)
			assert.Equal(t, ticks.MIDIQuarter{}, q)
			requireMustPanic(t, err, tt.own, func() { ticks.MustMIDIQuarterOf(tt.written) })
		})
	}
}

// TestResolutionBasics checks О-1, П-2, КП-19, КП-30 and 4.3.
func TestResolutionBasics(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		d          int
		fitsSMF    bool
		inMusicXML bool
		tickLength duration.Duration
	}{
		{name: "one", d: 1, fitsSMF: true, inMusicXML: true, tickLength: dur(1, 4)},
		{name: "480", d: 480, fitsSMF: true, inMusicXML: true, tickLength: dur(1, 1920)},
		{name: "MusicXML limit", d: 16383, fitsSMF: true, inMusicXML: true, tickLength: dur(1, 65532)},
		{name: "above MusicXML limit", d: 16384, fitsSMF: true, inMusicXML: false, tickLength: dur(1, 65536)},
		{name: "KP-30 20000", d: 20000, fitsSMF: true, inMusicXML: false, tickLength: dur(1, 80000)},
		{name: "SMF limit", d: 32767, fitsSMF: true, inMusicXML: false, tickLength: dur(1, 131068)},
		{name: "above SMF limit is still a value", d: 32768, fitsSMF: false, inMusicXML: false, tickLength: dur(1, 131072)},
		{name: "very large", d: 1 << 40, fitsSMF: false, inMusicXML: false, tickLength: duration.MustFromFraction(ratFrac("1/4398046511104"))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := ticks.NewResolution(tt.d)
			require.NoError(t, err)
			assert.True(t, r.IsSet())
			d, err := r.PerQuarter()
			require.NoError(t, err)
			assert.Equal(t, tt.d, d)
			assert.Equal(t, tt.d, r.MustPerQuarter())
			q, err := r.Quarter()
			require.NoError(t, err)
			assert.True(t, q.Equal(ticks.StandardQuarter()))
			assert.True(t, r.MustQuarter().Equal(ticks.StandardQuarter()))
			fits, err := r.FitsSMF()
			require.NoError(t, err)
			assert.Equal(t, tt.fitsSMF, fits)
			assert.Equal(t, tt.fitsSMF, r.MustFitsSMF())
			within, err := r.WithinMusicXML()
			require.NoError(t, err)
			assert.Equal(t, tt.inMusicXML, within)
			assert.Equal(t, tt.inMusicXML, r.MustWithinMusicXML())
			length, err := r.TickLength()
			require.NoError(t, err)
			assertEqualValue(t, tt.tickLength, length)
			assertEqualValue(t, tt.tickLength, r.MustTickLength())
		})
	}
}

func TestResolutionErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		d    int
		own  error
	}{
		{name: "KP-19 zero", d: 0, own: ticks.ErrZero},
		{name: "negative", d: -480, own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := ticks.NewResolution(tt.d)
			requireTicksError(t, err, tt.own)
			assert.Equal(t, ticks.Resolution{}, r)
			requireMustPanic(t, err, tt.own, func() { ticks.MustNewResolution(tt.d) })
		})
	}
}

// TestResolutionWithQuarter checks КР-TK-2, ТК-19, КП-45: the tick length is
// w/D and equality includes the written quarter.
func TestResolutionWithQuarter(t *testing.T) {
	t.Parallel()
	r := res(480)
	half := ticks.MustNewMIDIQuarter(16)
	h, err := r.WithQuarter(half)
	require.NoError(t, err)
	assert.Equal(t, 480, h.MustPerQuarter())
	assert.True(t, h.MustQuarter().Equal(half))
	assertEqualValue(t, dur(1, 960), h.MustTickLength())
	assert.False(t, h.Equal(r))
	assert.False(t, h == r)
	assert.True(t, r.MustWithQuarter(ticks.StandardQuarter()).Equal(r))
	assert.True(t, r.MustWithQuarter(ticks.StandardQuarter()) == r)
	assert.True(t, h.MustWithQuarter(ticks.StandardQuarter()).Equal(r))
	assert.False(t, res(480).Equal(res(96)))
	assert.True(t, h.Equal(resQ(480, 16)))

	_, err = r.WithQuarter(ticks.MIDIQuarter{})
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.Resolution{}.WithQuarter(half)
	requireTicksError(t, err, ticks.ErrUnset)
}

// TestResolutionWord checks О-3, КП-14, КП-20 for metric resolutions.
func TestResolutionWord(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		r    ticks.Resolution
		word uint16
	}{
		{name: "KP-14 96", r: res(96), word: 0x0060},
		{name: "KP-14 480", r: res(480), word: 0x01E0},
		{name: "one", r: res(1), word: 0x0001},
		{name: "SMF limit", r: res(32767), word: 0x7FFF},
		{name: "written quarter is not encoded", r: resQ(480, 16), word: 0x01E0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w, err := tt.r.Word()
			require.NoError(t, err)
			assert.Equal(t, tt.word, w)
			assert.Equal(t, tt.word, tt.r.MustWord())
		})
	}
	w, err := res(32768).Word()
	requireTicksError(t, err, ticks.ErrFormatLimit)
	assert.Zero(t, w)
	requireMustPanic(t, err, ticks.ErrFormatLimit, func() { res(32768).MustWord() })
}

// TestDecodeDivision checks О-3, П-5, КП-14 and 4.3.
func TestDecodeDivision(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		word  uint16
		r     ticks.Resolution
		smpte ticks.SMPTE
	}{
		{name: "KP-14 96", word: 0x0060, r: res(96)},
		{name: "KP-14 480", word: 0x01E0, r: res(480)},
		{name: "largest metric", word: 0x7FFF, r: res(32767)},
		{name: "KP-14 -25x40", word: 0xE728, smpte: smpte(-25, 40)},
		{name: "KP-14 -30x80", word: 0xE250, smpte: smpte(-30, 80)},
		{name: "-24x1", word: 0xE801, smpte: smpte(-24, 1)},
		{name: "-29x255 has no reading", word: 0xE3FF, smpte: smpte(-29, 255)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, s, err := ticks.DecodeDivision(tt.word)
			require.NoError(t, err)
			assert.True(t, tt.r.Equal(r), "resolution %s", r)
			assert.True(t, tt.smpte.Equal(s), "smpte %s", s)
			mr, ms := ticks.MustDecodeDivision(tt.word)
			assert.True(t, tt.r.Equal(mr))
			assert.True(t, tt.smpte.Equal(ms))
			if tt.r.IsSet() {
				assert.Equal(t, tt.word, r.MustWord())
				assert.True(t, r.MustQuarter().MustIsStandard())
			} else {
				assert.Equal(t, tt.word, s.MustWord())
				_, ok := s.MustReading()
				assert.False(t, ok)
			}
		})
	}
	for _, tt := range []struct {
		name string
		word uint16
		own  error
	}{
		{name: "metric zero", word: 0x0000, own: ticks.ErrZero},
		{name: "zero ticks per frame", word: 0xE700, own: ticks.ErrZero},
		{name: "KP-20 code -28", word: 0xE428, own: ticks.ErrInvalidArgument},
		{name: "code -128", word: 0x8028, own: ticks.ErrInvalidArgument},
		{name: "code -1", word: 0xFF28, own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, s, err := ticks.DecodeDivision(tt.word)
			requireTicksError(t, err, tt.own)
			assert.Equal(t, ticks.Resolution{}, r)
			assert.Equal(t, ticks.SMPTE{}, s)
			requireMustPanic(t, err, tt.own, func() { ticks.MustDecodeDivision(tt.word) })
		})
	}
}

// TestUnsetResolutionAndQuarter checks 4.2 for the zero values.
func TestUnsetResolutionAndQuarter(t *testing.T) {
	t.Parallel()
	var r ticks.Resolution
	assert.False(t, r.IsSet())
	assert.Equal(t, "Resolution(unset)", r.String())
	assert.True(t, r.Equal(ticks.Resolution{}))
	assert.False(t, r.Equal(res(1)))
	assert.False(t, res(1).Equal(r))
	for name, call := range map[string]func() error{
		"PerQuarter":     func() error { _, err := r.PerQuarter(); return err },
		"Quarter":        func() error { _, err := r.Quarter(); return err },
		"TickLength":     func() error { _, err := r.TickLength(); return err },
		"FitsSMF":        func() error { _, err := r.FitsSMF(); return err },
		"WithinMusicXML": func() error { _, err := r.WithinMusicXML(); return err },
		"Word":           func() error { _, err := r.Word(); return err },
		"AtLeast":        func() error { _, err := r.AtLeast(1); return err },
		"AtMost":         func() error { _, err := r.AtMost(1); return err },
		"Ticks":          func() error { _, err := r.Ticks(dur(1, 4)); return err },
		"Position":       func() error { _, err := r.Position(1); return err },
		"Rescale target": func() error { _, err := res(1).Rescale(1, r); return err },
		"MarshalText":    func() error { _, err := r.MarshalText(); return err },
		"TicksPerClock":  func() error { _, err := r.TicksPerClock(); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrUnset)
		})
	}

	var q ticks.MIDIQuarter
	assert.False(t, q.IsSet())
	assert.Equal(t, "MIDIQuarter(unset)", q.String())
	assert.True(t, q.Equal(ticks.MIDIQuarter{}))
	assert.False(t, q.Equal(ticks.StandardQuarter()))
	_, err := q.ThirtySeconds()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = q.Duration()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = q.IsStandard()
	requireTicksError(t, err, ticks.ErrUnset)
	requireMustPanic(t, err, ticks.ErrUnset, func() { q.MustIsStandard() })
	_, err = q.MarshalText()
	requireTicksError(t, err, ticks.ErrUnset)
	dst := []byte("prefix")
	out, err := q.AppendText(dst)
	requireTicksError(t, err, ticks.ErrUnset)
	assert.Equal(t, []byte("prefix"), out)
}
