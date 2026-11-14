package note_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Creation (architecture 6.4; OP-1...OP-7; K-9, K-9a, K-10, K-10a, K-34; P-3a).

// TestFromSpelling covers OP-1: a note without octave, bound or unbound,
// failures in the order E-1, E-3, E-5 (architecture 6.4).
func TestFromSpelling(t *testing.T) {
	t.Parallel()

	s := std()
	unset := &unsetSource{}
	half := cHalf()
	tests := []struct {
		name  string
		input spelling.Spelling
		c     cents.Cents
		src   settings.Source
		want  want
		err   error
		extra []error
	}{
		{name: "unbound", input: spelling.CSharp(), c: ct(0), want: wn(spelling.CSharp(), ct(0)).free()},
		{name: "zero cents value is 0", input: spelling.CSharp(), c: cents.Cents{}, want: wn(spelling.CSharp(), ct(0)).free()},
		{name: "bound with cents", input: spelling.CSharp(), c: ct(30), src: s, want: wn(spelling.CSharp(), ct(30)).boundTo(s)},
		{name: "unbound inadmissible name is not checked", input: half, c: ct(0), want: wn(half, ct(0)).free()},
		{name: "unset name", input: spelling.Spelling{}, src: s, err: note.ErrUnset},
		{name: "unset name before unset settings", input: spelling.Spelling{}, src: unset, err: note.ErrUnset},
		{name: "source with unset settings", input: spelling.C(), src: unset, err: note.ErrUnsetSettings, extra: []error{settings.ErrUnset}},
		{name: "non-nil source with nil settings", input: spelling.C(), src: nilSettingsSource(), err: note.ErrUnsetSettings, extra: []error{settings.ErrUnset}},
		{name: "inadmissible in the set", input: half, src: s, err: note.ErrInadmissible, extra: []error{system.ErrInadmissible}},
		{name: "unset settings before inadmissible", input: half, src: unset, err: note.ErrUnsetSettings},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := note.FromSpelling(tt.input, tt.c, tt.src)
			if tt.err != nil {
				assertErr(t, err, tt.err, tt.extra...)
				assert.True(t, got == note.Note{}, "the result on error is Note{}")
				return
			}
			require.NoError(t, err)
			assertNote(t, tt.want, got)
		})
	}
}

// TestFromPitchAndOctave covers OP-2 and OP-3: FromPitch, FromOctave, FromASPN.
func TestFromPitchAndOctave(t *testing.T) {
	t.Parallel()

	s := std()
	unset := &unsetSource{}
	half := cHalf()
	tests := []struct {
		name  string
		make  func() (note.Note, error)
		want  want
		err   error
		extra []error
	}{
		{
			name: "FromPitch bound",
			make: func() (note.Note, error) { return note.FromPitch(pt(spelling.BSharp(), 4), ct(10), s) },
			want: wp(spelling.BSharp(), 4, ct(10)).boundTo(s),
		},
		{
			name: "FromPitch unbound",
			make: func() (note.Note, error) { return note.FromPitch(pt(spelling.C(), 4), cents.Cents{}, nil) },
			want: wp(spelling.C(), 4, ct(0)).free(),
		},
		{
			name: "FromPitch unset pitch",
			make: func() (note.Note, error) { return note.FromPitch(pitch.Pitch{}, ct(0), s) },
			err:  note.ErrUnset,
		},
		{
			name: "FromPitch unset settings",
			make: func() (note.Note, error) { return note.FromPitch(pt(spelling.C(), 4), ct(0), unset) },
			err:  note.ErrUnsetSettings, extra: []error{settings.ErrUnset},
		},
		{
			name: "FromPitch inadmissible",
			make: func() (note.Note, error) { return note.FromPitch(pt(half, 4), ct(0), s) },
			err:  note.ErrInadmissible, extra: []error{system.ErrInadmissible},
		},
		{
			name: "FromPitch unbound inadmissible name",
			make: func() (note.Note, error) { return note.FromPitch(pt(half, 4), ct(0), nil) },
			want: wp(half, 4, ct(0)).free(),
		},
		{
			name: "FromOctave",
			make: func() (note.Note, error) { return note.FromOctave(spelling.B(), oct(3), ct(-5), s) },
			want: wp(spelling.B(), 3, ct(-5)).boundTo(s),
		},
		{
			name: "FromOctave unset name",
			make: func() (note.Note, error) { return note.FromOctave(spelling.Spelling{}, oct(3), ct(0), s) },
			err:  note.ErrUnset,
		},
		{
			name: "FromOctave unset octave",
			make: func() (note.Note, error) { return note.FromOctave(spelling.B(), octave.Octave{}, ct(0), s) },
			err:  note.ErrUnset,
		},
		{
			name: "FromOctave unset octave before unset settings",
			make: func() (note.Note, error) { return note.FromOctave(spelling.B(), octave.Octave{}, ct(0), unset) },
			err:  note.ErrUnset,
		},
		{
			name: "FromOctave inadmissible",
			make: func() (note.Note, error) { return note.FromOctave(half, oct(3), ct(0), s) },
			err:  note.ErrInadmissible, extra: []error{system.ErrInadmissible},
		},
		{
			name: "FromASPN",
			make: func() (note.Note, error) { return note.FromASPN(spelling.CSharp(), 4, ct(30), s) },
			want: wp(spelling.CSharp(), 4, ct(30)).boundTo(s),
		},
		{
			name: "FromASPN smallest int",
			make: func() (note.Note, error) { return note.FromASPN(spelling.C(), math.MinInt, ct(0), nil) },
			want: wp(spelling.C(), math.MinInt, ct(0)).free(),
		},
		{
			name: "FromASPN largest int",
			make: func() (note.Note, error) { return note.FromASPN(spelling.C(), math.MaxInt, ct(0), s) },
			want: wp(spelling.C(), math.MaxInt, ct(0)).boundTo(s),
		},
		{
			name: "FromASPN unset name",
			make: func() (note.Note, error) { return note.FromASPN(spelling.Spelling{}, 4, ct(0), nil) },
			err:  note.ErrUnset,
		},
		{
			name: "FromASPN unset settings",
			make: func() (note.Note, error) { return note.FromASPN(spelling.C(), 4, ct(0), unset) },
			err:  note.ErrUnsetSettings, extra: []error{settings.ErrUnset},
		},
		{
			name: "FromASPN inadmissible",
			make: func() (note.Note, error) { return note.FromASPN(half, 4, ct(0), s) },
			err:  note.ErrInadmissible, extra: []error{system.ErrInadmissible},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.make()
			if tt.err != nil {
				assertErr(t, err, tt.err, tt.extra...)
				assert.True(t, got == note.Note{}, "the result on error is Note{}")
				return
			}
			require.NoError(t, err)
			assertNote(t, tt.want, got)
		})
	}
}

// TestFromWritten covers OP-4: a name in a written octave (register) of the set;
// K-34 (B♯ in ASPN 3 and in Franco-Belgian number 2 is B♯ in Octave 4).
func TestFromWritten(t *testing.T) {
	t.Parallel()

	s := std()
	fb := stdFB()
	unset := &unsetSource{}
	half := cHalf()
	tests := []struct {
		name  string
		make  func() (note.Note, error)
		want  want
		err   error
		extra []error
	}{
		{
			name: "K-34 FromWritten B♯ in written octave 3",
			make: func() (note.Note, error) { return note.FromWritten(s, spelling.BSharp(), oct(3), ct(7)) },
			want: wp(spelling.BSharp(), 4, ct(7)).boundTo(s),
		},
		{
			name: "K-34 FromWrittenASPN B♯ 3",
			make: func() (note.Note, error) { return note.FromWrittenASPN(s, spelling.BSharp(), 3, ct(0)) },
			want: wp(spelling.BSharp(), 4, ct(0)).boundTo(s),
		},
		{
			name: "K-34 FromWrittenNumber Franco-Belgian 2",
			make: func() (note.Note, error) { return note.FromWrittenNumber(fb, spelling.BSharp(), 2, ct(0)) },
			want: wp(spelling.BSharp(), 4, ct(0)).boundTo(fb),
		},
		{
			name: "C♭ written 4 is Octave 3",
			make: func() (note.Note, error) { return note.FromWrittenASPN(s, spelling.CFlat(), 4, ct(0)) },
			want: wp(spelling.CFlat(), 3, ct(0)).boundTo(s),
		},
		{
			name: "nil source",
			make: func() (note.Note, error) { return note.FromWritten(nil, spelling.B(), oct(3), ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "nil source ASPN",
			make: func() (note.Note, error) { return note.FromWrittenASPN(nil, spelling.B(), 3, ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "nil source number",
			make: func() (note.Note, error) { return note.FromWrittenNumber(nil, spelling.B(), 3, ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset name",
			make: func() (note.Note, error) { return note.FromWritten(s, spelling.Spelling{}, oct(3), ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset written octave",
			make: func() (note.Note, error) { return note.FromWritten(s, spelling.B(), octave.Octave{}, ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset settings",
			make: func() (note.Note, error) { return note.FromWrittenASPN(unset, spelling.B(), 3, ct(0)) },
			err:  note.ErrUnsetSettings, extra: []error{settings.ErrUnset},
		},
		{
			name: "inadmissible",
			make: func() (note.Note, error) { return note.FromWrittenASPN(s, half, 3, ct(0)) },
			err:  note.ErrInadmissible, extra: []error{system.ErrInadmissible},
		},
		{
			name: "inadmissible before out of range",
			make: func() (note.Note, error) {
				return note.FromWrittenASPN(setK25(), spelling.BDoubleSharp(), math.MaxInt, ct(0))
			},
			err: note.ErrInadmissible, extra: []error{system.ErrInadmissible},
		},
		{
			name: "Octave out of range",
			make: func() (note.Note, error) { return note.FromWrittenASPN(s, spelling.BSharp(), math.MaxInt, ct(0)) },
			err:  note.ErrOutOfRange, extra: []error{system.ErrOutOfRange},
		},
		{
			name: "skipped visible number",
			make: func() (note.Note, error) { return note.FromWrittenNumber(fb, spelling.C(), 0, ct(0)) },
			err:  nil, extra: []error{number.ErrSkipped},
		},
		{
			name: "visible number out of range",
			make: func() (note.Note, error) { return note.FromWrittenNumber(fb, spelling.C(), math.MaxInt, ct(0)) },
			err:  note.ErrOutOfRange, extra: []error{octave.ErrOutOfRange, number.ErrOutOfRange},
		},
		{
			name: "inadmissible before skipped number",
			make: func() (note.Note, error) { return note.FromWrittenNumber(fb, half, 0, ct(0)) },
			err:  note.ErrInadmissible,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.make()
			if tt.err != nil || tt.extra != nil {
				assertErr(t, err, tt.err, tt.extra...)
				assert.True(t, got == note.Note{}, "the result on error is Note{}")
				return
			}
			require.NoError(t, err)
			assertNote(t, tt.want, got)
		})
	}
}

// TestParse covers OP-5, OP-5a: K-9, K-9a; the reasons of settings without a
// note sentinel; E-5 with settings.ErrInadmissible (architecture 6.2.2).
func TestParse(t *testing.T) {
	t.Parallel()

	s := std()
	ly := stdNotation("lilypond-english")
	unset := &unsetSource{}
	tests := []struct {
		name  string
		make  func() (note.Note, error)
		want  want
		err   error
		extra []error
	}{
		{
			name: "K-9 ParsePitch B♯3",
			make: func() (note.Note, error) { return note.ParsePitch(s, "B♯3") },
			want: wp(spelling.BSharp(), 4, ct(0)).boundTo(s),
		},
		{
			name: "K-9 Parse D♭",
			make: func() (note.Note, error) { return note.Parse(s, "D♭") },
			want: wn(spelling.DFlat(), ct(0)).boundTo(s),
		},
		{
			name: "K-9a lilypond Parse c",
			make: func() (note.Note, error) { return note.Parse(ly, "c") },
			want: wn(spelling.C(), ct(0)).boundTo(ly),
		},
		{
			name: "K-9a lilypond ParsePitch c is Octave 3",
			make: func() (note.Note, error) { return note.ParsePitch(ly, "c") },
			want: wp(spelling.C(), 3, ct(0)).boundTo(ly),
		},
		{
			name: "nil source",
			make: func() (note.Note, error) { return note.Parse(nil, "C") },
			err:  note.ErrUnset,
		},
		{
			name: "nil source pitch",
			make: func() (note.Note, error) { return note.ParsePitch(nil, "C4") },
			err:  note.ErrUnset,
		},
		{
			name: "unset settings",
			make: func() (note.Note, error) { return note.Parse(unset, "C") },
			err:  note.ErrUnsetSettings, extra: []error{settings.ErrUnset},
		},
		{
			name: "unset settings before a bad text",
			make: func() (note.Note, error) { return note.ParsePitch(unset, "xyz") },
			err:  note.ErrUnsetSettings, extra: []error{settings.ErrUnset},
		},
		{
			name:  "text does not match: no note sentinel",
			make:  func() (note.Note, error) { return note.Parse(s, "xyz") },
			extra: []error{notation.ErrParse},
		},
		{
			name:  "pitch text does not match: no note sentinel",
			make:  func() (note.Note, error) { return note.ParsePitch(s, "xyz") },
			extra: []error{notation.ErrParse},
		},
		{
			name:  "out of range in parsing: no E-8, the reason of the dependency",
			make:  func() (note.Note, error) { return note.ParsePitch(s, "C99999999999999999999") },
			extra: []error{number.ErrOutOfRange},
		},
		{
			name: "inadmissible name found by settings",
			make: func() (note.Note, error) { return note.Parse(ly, "cqs") },
			err:  note.ErrInadmissible, extra: []error{settings.ErrInadmissible, system.ErrInadmissible},
		},
		{
			name: "inadmissible pitch found by settings",
			make: func() (note.Note, error) { return note.ParsePitch(ly, "cqs'") },
			err:  note.ErrInadmissible, extra: []error{settings.ErrInadmissible, system.ErrInadmissible},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.make()
			if tt.err != nil || tt.extra != nil {
				assertErr(t, err, tt.err, tt.extra...)
				assert.True(t, got == note.Note{}, "the result on error is Note{}")
				return
			}
			require.NoError(t, err)
			assertNote(t, tt.want, got)
		})
	}
}

// TestFromStep covers OP-6 and K-10a: a bound note from a source, an unbound
// note from an explicit tuning (P-3a).
func TestFromStep(t *testing.T) {
	t.Parallel()

	s := std()
	tests := []struct {
		name  string
		make  func() (note.Note, error)
		want  want
		err   error
		extra []error
	}{
		{
			name: "K-10a bound C4 +5",
			make: func() (note.Note, error) { return note.FromStep(s, -9, system.Sharps(), ct(5)) },
			want: wp(spelling.C(), 4, ct(5)).boundTo(s),
		},
		{
			name: "K-10a explicit tuning unbound C4 +5",
			make: func() (note.Note, error) { return note.FromStepInTuning(tuning.Standard(), -9, system.Sharps(), ct(5)) },
			want: wp(spelling.C(), 4, ct(5)).free(),
		},
		{
			name: "step 1 flats is B♭4",
			make: func() (note.Note, error) { return note.FromStep(s, 1, system.Flats(), ct(0)) },
			want: wp(spelling.BFlat(), 4, ct(0)).boundTo(s),
		},
		{
			name: "step 0 in set 19",
			make: func() (note.Note, error) { return note.FromStepInTuning(tuning19(), 0, system.Sharps(), ct(0)) },
			want: wp(spelling.A(), 4, ct(0)).free(),
		},
		{
			name: "nil source",
			make: func() (note.Note, error) { return note.FromStep(nil, 0, system.Sharps(), ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset rule",
			make: func() (note.Note, error) { return note.FromStep(s, 0, system.SpellingRule{}, ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset rule before unset settings",
			make: func() (note.Note, error) { return note.FromStep(&unsetSource{}, 0, system.SpellingRule{}, ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset settings",
			make: func() (note.Note, error) { return note.FromStep(&unsetSource{}, 0, system.Sharps(), ct(0)) },
			err:  note.ErrUnsetSettings, extra: []error{settings.ErrUnset},
		},
		{
			name: "nil tuning",
			make: func() (note.Note, error) { return note.FromStepInTuning(nil, 0, system.Sharps(), ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name: "unset tuning",
			make: func() (note.Note, error) { return note.FromStepInTuning(&tuning.Tuning{}, 0, system.Sharps(), ct(0)) },
			err:  note.ErrUnset,
		},
		{
			name:  "bad spelling rule: tuning reason without note sentinel",
			make:  func() (note.Note, error) { return note.FromStep(s, 0, badRule(), ct(0)) },
			extra: []error{tuning.ErrSpelling, system.ErrBadSpelling},
		},
		{
			name:  "bad spelling rule in a tuning",
			make:  func() (note.Note, error) { return note.FromStepInTuning(tuning.Standard(), 0, badRule(), ct(0)) },
			extra: []error{tuning.ErrSpelling, system.ErrBadSpelling},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.make()
			if tt.err != nil || tt.extra != nil {
				assertErr(t, err, tt.err, tt.extra...)
				assert.True(t, got == note.Note{}, "the result on error is Note{}")
				return
			}
			require.NoError(t, err)
			assertNote(t, tt.want, got)
		})
	}
}

// TestNearest covers OP-7 and K-10: the nearest note with the deviation δ as
// cents; K-10 is the note and δ, not the equality of frequencies (R-8).
func TestNearest(t *testing.T) {
	t.Parallel()

	t.Run("K-10", func(t *testing.T) {
		t.Parallel()
		s := std()
		for _, tc := range []struct {
			freq float64
			name spelling.Spelling
			dev  float64
		}{
			{445, spelling.A(), 19.562175},
			{262, spelling.C(), 2.475946},
		} {
			got, err := note.Nearest(s, tc.freq, system.Sharps())
			require.NoError(t, err)
			wantP, wantDev := tuning.Standard().MustNearestPitch(tc.freq, system.Sharps())
			assertNote(t, wp(tc.name, 4, wantDev).boundTo(s), got, "%v Hz", tc.freq)
			assert.Truef(t, wantP.Equal(got.MustPitch()), "%v Hz: pitch %v", tc.freq, wantP)
			assert.InDelta(t, tc.dev, got.Cents().Float64(), 1e-6, "%v Hz", tc.freq)
		}
	})

	t.Run("explicit tuning is unbound", func(t *testing.T) {
		t.Parallel()
		got, err := note.NearestInTuning(tuning.Standard(), 445, system.Sharps())
		require.NoError(t, err)
		_, dev := tuning.Standard().MustNearestPitch(445, system.Sharps())
		assertNote(t, wp(spelling.A(), 4, dev).free(), got)
	})

	t.Run("every rounding mode agrees with tuning", func(t *testing.T) {
		t.Parallel()
		s := std()
		tu := tuning.Standard()
		// The midpoint between A4 and A♯4 and an ordinary frequency.
		mid := 440 * math.Pow(2, 1.0/24)
		for _, mode := range rounding.Modes() {
			for _, f := range []float64{mid, 450} {
				wantP, wantDev := tu.MustNearestPitchBy(f, mode, system.Flats())
				got, err := note.NearestBy(s, f, mode, system.Flats())
				require.NoError(t, err, "%v %v", mode, f)
				assertNote(t, wp(wantP.Spelling(), wantP.MustASPN(), wantDev).boundTo(s), got, "%v %v", mode, f)

				gotT, err := note.NearestByInTuning(tu, f, mode, system.Flats())
				require.NoError(t, err, "%v %v", mode, f)
				assertNote(t, wp(wantP.Spelling(), wantP.MustASPN(), wantDev).free(), gotT, "%v %v", mode, f)
			}
		}
	})

	t.Run("default mode is tuning.NearestMode", func(t *testing.T) {
		t.Parallel()
		s := std()
		mid := 440 * math.Pow(2, 1.0/24)
		a := note.MustNearest(s, mid, system.Sharps())
		b := note.MustNearestBy(s, mid, tuning.NearestMode, system.Sharps())
		assert.True(t, a.Equal(b))
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()
		s := std()
		unset := &unsetSource{}
		tests := []struct {
			name  string
			make  func() (note.Note, error)
			err   error
			extra []error
		}{
			{"nil source", func() (note.Note, error) { return note.Nearest(nil, 440, system.Sharps()) }, note.ErrUnset, nil},
			{"unset rule", func() (note.Note, error) { return note.Nearest(s, 440, system.SpellingRule{}) }, note.ErrUnset, nil},
			{"unset mode", func() (note.Note, error) { return note.NearestBy(s, 440, rounding.Mode(0), system.Sharps()) }, note.ErrUnset, nil},
			{"unset mode before unset settings", func() (note.Note, error) {
				return note.NearestBy(unset, 440, rounding.Mode(0), system.Sharps())
			}, note.ErrUnset, nil},
			{"nil tuning", func() (note.Note, error) { return note.NearestInTuning(nil, 440, system.Sharps()) }, note.ErrUnset, nil},
			{"unset tuning", func() (note.Note, error) {
				return note.NearestByInTuning(&tuning.Tuning{}, 440, rounding.HalfUp, system.Sharps())
			}, note.ErrUnset, nil},
			{"unset settings", func() (note.Note, error) { return note.Nearest(unset, 440, system.Sharps()) }, note.ErrUnsetSettings, []error{settings.ErrUnset}},
			{"unset settings before a bad frequency", func() (note.Note, error) { return note.Nearest(unset, -1, system.Sharps()) }, note.ErrUnsetSettings, nil},
			{"negative frequency", func() (note.Note, error) { return note.Nearest(s, -1, system.Sharps()) }, nil, []error{tuning.ErrInvalidArgument}},
			{"NaN frequency", func() (note.Note, error) { return note.NearestInTuning(tuning.Standard(), math.NaN(), system.Sharps()) }, nil, []error{tuning.ErrInvalidArgument}},
			{"unknown mode", func() (note.Note, error) { return note.NearestBy(s, 440, rounding.Mode(255), system.Sharps()) }, nil, []error{tuning.ErrInvalidArgument}},
			{"bad spelling rule", func() (note.Note, error) { return note.Nearest(s, 440, badRule()) }, nil, []error{tuning.ErrSpelling, system.ErrBadSpelling}},
			{"bad frequency before bad rule", func() (note.Note, error) { return note.Nearest(s, math.Inf(1), badRule()) }, nil, []error{tuning.ErrInvalidArgument}},
		}
		for _, tt := range tests {
			got, err := tt.make()
			assertErr(t, err, tt.err, tt.extra...)
			assert.True(t, got == note.Note{}, tt.name)
		}
	})
}

// TestCreationReadsTheSourceOnce covers P-5 for creation: one read per call,
// none without a source.
func TestCreationReadsTheSourceOnce(t *testing.T) {
	t.Parallel()

	calls := []struct {
		name string
		make func(settings.Source) error
	}{
		{"FromSpelling", func(s settings.Source) error { _, err := note.FromSpelling(spelling.C(), ct(0), s); return err }},
		{"FromPitch", func(s settings.Source) error { _, err := note.FromPitch(pt(spelling.C(), 4), ct(0), s); return err }},
		{"FromOctave", func(s settings.Source) error { _, err := note.FromOctave(spelling.C(), oct(4), ct(0), s); return err }},
		{"FromASPN", func(s settings.Source) error { _, err := note.FromASPN(spelling.C(), 4, ct(0), s); return err }},
		{"FromWritten", func(s settings.Source) error { _, err := note.FromWritten(s, spelling.C(), oct(4), ct(0)); return err }},
		{"FromWrittenASPN", func(s settings.Source) error { _, err := note.FromWrittenASPN(s, spelling.C(), 4, ct(0)); return err }},
		{"FromWrittenNumber", func(s settings.Source) error { _, err := note.FromWrittenNumber(s, spelling.C(), 4, ct(0)); return err }},
		{"Parse", func(s settings.Source) error { _, err := note.Parse(s, "C"); return err }},
		{"ParsePitch", func(s settings.Source) error { _, err := note.ParsePitch(s, "C4"); return err }},
		{"FromStep", func(s settings.Source) error { _, err := note.FromStep(s, 0, system.Sharps(), ct(0)); return err }},
		{"Nearest", func(s settings.Source) error { _, err := note.Nearest(s, 440, system.Sharps()); return err }},
		{"NearestBy", func(s settings.Source) error {
			_, err := note.NearestBy(s, 440, rounding.HalfUp, system.Sharps())
			return err
		}},
	}
	for _, c := range calls {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			src := newCounting(std())
			require.NoError(t, c.make(src))
			assert.EqualValues(t, 1, src.reads())
		})
	}
}

// TestCreatedBoundNoteKeepsTheSource checks architecture 6.4: a bound created
// note holds the given source, not the set read from it (P-5).
func TestCreatedBoundNoteKeepsTheSource(t *testing.T) {
	t.Parallel()

	src := newSwitch(std())
	n := note.MustFromPitch(pt(spelling.A(), 4), ct(0), src)
	assert.True(t, n.Source() == settings.Source(src))
	assert.InDelta(t, 440.0, n.MustFrequency(), 1e-9)
	src.set(stdAt(415))
	assert.InDelta(t, 415.0, n.MustFrequency(), 1e-9)
}
