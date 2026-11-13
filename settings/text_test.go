package settings_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNames covers K-9, K-23, K-24 and A 5.7 for names: expressibility,
// formatting and parsing.
func TestNames(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	lily := setIn(t, system.Traditional(), builtin(notation.LilyPondEnglishID))
	italian := std.MustWithNotation(builtin(notation.RomanceItalianID))
	quarter := altered(spelling.C(), 1, 2)
	plus3 := spelling.C().WithAlteration(alteration.MustFromInt(3))
	for _, tc := range []struct {
		name      string
		set       *settings.Settings
		spelling  spelling.Spelling
		expresses bool
		text      string
		want      []error
		notWant   []error
	}{
		{name: "K-9 C sharp in the standard settings", set: std, spelling: spelling.CSharp(), expresses: true, text: "C♯"},
		{name: "K-24 D flat in the standard settings", set: std, spelling: spelling.DFlat(), expresses: true, text: "D♭"},
		{name: "K-9 C +3 is admissible but not writable", set: std, spelling: plus3, want: []error{notation.ErrUnwritableName}, notWant: []error{settings.ErrInadmissible}},
		{
			name: "K-9 C +1/2 is neither writable nor admissible: E-4 first", set: std, spelling: quarter,
			want: []error{settings.ErrInadmissible, system.ErrInadmissible}, notWant: []error{notation.ErrUnwritableName},
		},
		{name: "K-24 C +1/2 in the LilyPond domain, not admissible", set: lily, spelling: quarter, want: []error{settings.ErrInadmissible, system.ErrInadmissible}},
		{name: "K-24 LilyPond C sharp", set: lily, spelling: spelling.CSharp(), expresses: true, text: "cs"},
		{name: "K-23 Italian C sharp", set: italian, spelling: spelling.CSharp(), expresses: true, text: "do diesis"},
		{name: "German A double flat is not writable", set: std.MustWithNotation(builtin(notation.GermanAMID)), spelling: spelling.ADoubleFlat(), want: []error{notation.ErrUnwritableName}},
		{
			name: "quarter tone in a quarter-tone system", set: setIn(t, system.QuarterTones24(), builtin(notation.LilyPondEnglishID)),
			spelling: quarter, expresses: true, text: "cqs",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, err := tc.set.Expresses(tc.spelling)
			require.NoError(t, err, "Expresses answers instead of failing")
			assert.Equal(t, tc.expresses, ok)

			text, err := tc.set.Format(tc.spelling)
			if !tc.expresses {
				require.Error(t, err)
				assert.Empty(t, text)
				for _, want := range tc.want {
					assert.ErrorIs(t, err, want)
				}
				for _, notWant := range tc.notWant {
					assert.NotErrorIs(t, err, notWant)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.text, text)
			back, err := tc.set.Parse(text)
			require.NoError(t, err)
			assert.Equal(t, tc.spelling, back)
		})
	}
}

// TestParse covers K-23, K-24 and the order of causes of A 5.7 for Parse.
func TestParse(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	lily := setIn(t, system.Traditional(), builtin(notation.LilyPondEnglishID))
	var unsetNil *settings.Settings
	for _, tc := range []struct {
		name    string
		set     *settings.Settings
		text    string
		want    spelling.Spelling
		errs    []error
		notErrs []error
	}{
		{name: "K-24 D flat", set: std, text: "D♭", want: spelling.DFlat()},
		{name: "K-23 Italian do diesis", set: std.MustWithNotation(builtin(notation.RomanceItalianID)), text: "do diesis", want: spelling.CSharp()},
		{name: "LilyPond alias", set: lily, text: "c-sharp", want: spelling.CSharp()},
		{name: "unset set", set: unsetNil, text: "C♯", errs: []error{settings.ErrUnset}},
		{name: "unset set before a bad string", set: &settings.Settings{}, text: "?", errs: []error{settings.ErrUnset}, notErrs: []error{notation.ErrParse}},
		{name: "string does not match", set: std, text: "H", errs: []error{notation.ErrParse}, notErrs: []error{settings.ErrInadmissible}},
		{name: "empty string", set: std, text: "", errs: []error{notation.ErrParse}},
		{name: "K-24 cqs is not admissible", set: lily, text: "cqs", errs: []error{settings.ErrInadmissible, system.ErrInadmissible}, notErrs: []error{notation.ErrParse}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.set.Parse(tc.text)
			if len(tc.errs) == 0 {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.True(t, tc.set.MustExpresses(got), "a parsed name is expressible")
				return
			}
			require.Error(t, err)
			assert.Equal(t, spelling.Spelling{}, got)
			for _, want := range tc.errs {
				assert.ErrorIs(t, err, want)
			}
			for _, notWant := range tc.notErrs {
				assert.NotErrorIs(t, err, notWant)
			}
		})
	}
}

// TestPitches covers K-9, K-10, K-11, K-23, K-24, K-25 and A 5.7 for pitches:
// expressibility, formatting and parsing.
func TestPitches(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	russian := std.MustWithNotation(builtin(notation.RussianAMID))
	named := std.MustWithNotation(englishNamed(t)).MustWithCatalog(naming.Russian())
	namedEmpty := named.MustWithCatalog(nil)
	italian := std.MustWithNotation(builtin(notation.RomanceItalianID))
	german := std.MustWithNotation(builtin(notation.GermanAMID))
	lily := setIn(t, system.Traditional(), builtin(notation.LilyPondEnglishID))
	for _, tc := range []struct {
		name      string
		set       *settings.Settings
		pitch     pitch.Pitch
		expresses bool
		text      string
		want      []error
		notWant   []error
	}{
		{name: "K-9 C sharp 4", set: std, pitch: at(spelling.CSharp(), 4), expresses: true, text: "C♯4"},
		{name: "K-11 B sharp with Octave 4", set: std, pitch: at(spelling.BSharp(), 4), expresses: true, text: "B♯3"},
		{name: "C flat with Octave 3", set: std, pitch: at(spelling.CFlat(), 3), expresses: true, text: "C♭4"},
		{name: "negative octave", set: std, pitch: at(spelling.C(), -3), expresses: true, text: "C-3"},
		{name: "German C sharp 4", set: german, pitch: at(spelling.CSharp(), 4), expresses: true, text: "cis¹"},
		{name: "K-10 Russian AM C with Octave 0", set: russian, pitch: at(spelling.C(), 0), expresses: true, text: russian.Notation().MustFormatPitch(at(spelling.C(), 0), system.Traditional(), notation.OctaveContext{})},
		{name: "K-10 Russian AM C with Octave 8", set: russian, pitch: at(spelling.C(), 8), expresses: true, text: russian.Notation().MustFormatPitch(at(spelling.C(), 8), system.Traditional(), notation.OctaveContext{})},
		{name: "K-10 Russian AM C with Octave -1", set: russian, pitch: at(spelling.C(), -1), want: []error{notation.ErrNoOctave}},
		{name: "K-10 Russian AM C with Octave 9", set: russian, pitch: at(spelling.C(), 9), want: []error{notation.ErrNoOctave}},
		{name: "Russian AM B sharp with Octave 9 is written 8", set: russian, pitch: at(spelling.BSharp(), 9), expresses: true, text: russian.Notation().MustFormatPitch(at(spelling.BSharp(), 9), system.Traditional(), notation.OctaveContext{})},
		{name: "K-10 name style with the Russian catalog", set: named, pitch: at(spelling.C(), 4), expresses: true, text: "C " + naming.Russian().Name(4)},
		{name: "name style beyond the Russian catalog", set: named, pitch: at(spelling.C(), 10), want: []error{notation.ErrNoOctave}},
		{name: "K-10 name style with the empty catalog", set: namedEmpty, pitch: at(spelling.C(), 4), want: []error{notation.ErrNoOctave}},
		{name: "K-23 name-only notation", set: italian, pitch: at(spelling.CSharp(), 4), want: []error{notation.ErrNoOctave}},
		{name: "own numbering", set: std.MustWithNumbering(number.FrancoBelgian()), pitch: at(spelling.C(), 1), expresses: true, text: "C-1"},
		{
			name: "K-24 C +1/2 with Octave 4", set: std, pitch: at(altered(spelling.C(), 1, 2), 4),
			want: []error{settings.ErrInadmissible, system.ErrInadmissible}, notWant: []error{notation.ErrUnwritableName},
		},
		{name: "K-24 LilyPond C +1/2 with Octave 4", set: lily, pitch: at(altered(spelling.C(), 1, 2), 4), want: []error{settings.ErrInadmissible, system.ErrInadmissible}},
		{name: "inadmissible before the octave range", set: std, pitch: at(altered(spelling.C(), 1, 2), math.MinInt), want: []error{settings.ErrInadmissible}, notWant: []error{system.ErrOutOfRange}},
		{name: "C +3 is not writable", set: std, pitch: at(spelling.C().WithAlteration(alteration.MustFromInt(3)), 4), want: []error{notation.ErrUnwritableName}, notWant: []error{settings.ErrInadmissible}},
		{
			name: "unwritable name before the octave range", set: std, pitch: at(spelling.B().WithAlteration(alteration.MustFromInt(3)), math.MinInt),
			want: []error{notation.ErrUnwritableName}, notWant: []error{system.ErrOutOfRange},
		},
		{name: "K-25 B sharp with the smallest Octave", set: std, pitch: at(spelling.BSharp(), math.MinInt), want: []error{system.ErrOutOfRange}},
		{name: "C flat with the largest Octave", set: std, pitch: at(spelling.CFlat(), math.MaxInt), want: []error{system.ErrOutOfRange}},
		{name: "Mexican number of the largest octave", set: std.MustWithNumbering(number.Mexican()), pitch: at(spelling.C(), math.MaxInt), want: []error{number.ErrOutOfRange}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, err := tc.set.ExpressesPitch(tc.pitch)
			require.NoError(t, err, "ExpressesPitch answers instead of failing")
			assert.Equal(t, tc.expresses, ok)

			text, err := tc.set.FormatPitch(tc.pitch)
			if !tc.expresses {
				require.Error(t, err)
				assert.Empty(t, text)
				for _, want := range tc.want {
					assert.ErrorIs(t, err, want)
				}
				for _, notWant := range tc.notWant {
					assert.NotErrorIs(t, err, notWant)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.text, text)
			back, err := tc.set.ParsePitch(text)
			require.NoError(t, err)
			assert.True(t, tc.pitch.Equal(back), "ParsePitch(%q) = %v, want %v", text, back, tc.pitch)
		})
	}
}

// TestParsePitch covers K-11, K-23, K-24 and the order of causes of P-17a
// for ParsePitch.
func TestParsePitch(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	lily := setIn(t, system.Traditional(), builtin(notation.LilyPondEnglishID))
	lilyNumbered := newSet(t, tuning.Standard(), styled(t, builtin(notation.LilyPondEnglishID).NameRules(), notation.OctaveStyleNumber), number.FrancoBelgian(), nil)
	var unsetNil *settings.Settings
	for _, tc := range []struct {
		name    string
		set     *settings.Settings
		text    string
		want    pitch.Pitch
		errs    []error
		notErrs []error
	}{
		{name: "K-11 B sharp 3", set: std, text: "B♯3", want: at(spelling.BSharp(), 4)},
		{name: "C flat 4", set: std, text: "C♭4", want: at(spelling.CFlat(), 3)},
		{name: "LilyPond cs'", set: lily, text: "cs'", want: at(spelling.CSharp(), 4)},
		{name: "own numbering", set: lilyNumbered, text: "c1", want: at(spelling.C(), 2)},
		{name: "unset set", set: unsetNil, text: "C4", errs: []error{settings.ErrUnset}},
		{name: "zero set before a bad string", set: &settings.Settings{}, text: "?", errs: []error{settings.ErrUnset}, notErrs: []error{notation.ErrParse}},
		{name: "string does not match", set: std, text: "C", errs: []error{notation.ErrParse}},
		{name: "K-23 name-only notation", set: std.MustWithNotation(builtin(notation.RomanceItalianID)), text: "do diesis", errs: []error{notation.ErrNoOctave}},
		{name: "unknown octave name", set: std.MustWithNotation(englishNamed(t)), text: "C nowhere", errs: []error{naming.ErrUnknownName}},
		{
			name: "K-24 cqs' restores the octave and fails with E-4", set: lily, text: "cqs'",
			errs: []error{settings.ErrInadmissible, system.ErrInadmissible}, notErrs: []error{notation.ErrParse},
		},
		{name: "inadmissible name with a valid number", set: lilyNumbered, text: "cqs1", errs: []error{settings.ErrInadmissible, system.ErrInadmissible}},
		{
			name: "P-17a skipped number reported before the inadmissible name", set: lilyNumbered, text: "cqs0",
			errs: []error{number.ErrSkipped}, notErrs: []error{settings.ErrInadmissible},
		},
		{name: "number out of range", set: std, text: "C99999999999999999999", errs: []error{number.ErrOutOfRange}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.set.ParsePitch(tc.text)
			if len(tc.errs) == 0 {
				require.NoError(t, err)
				assert.True(t, tc.want.Equal(got), "got %v, want %v", got, tc.want)
				assert.True(t, tc.set.MustExpressesPitch(got), "a parsed pitch is expressible")
				return
			}
			require.Error(t, err)
			assert.Equal(t, pitch.Pitch{}, got)
			for _, want := range tc.errs {
				assert.ErrorIs(t, err, want)
			}
			for _, notWant := range tc.notErrs {
				assert.NotErrorIs(t, err, notWant)
			}
		})
	}
}

// TestUnsetArguments covers K-22, E-3 and the order E-1 before E-3.
func TestUnsetArguments(t *testing.T) {
	t.Parallel()
	std := settings.Standard()
	var unsetNil *settings.Settings
	for _, tc := range []struct {
		name string
		set  *settings.Settings
		want error
		not  error
	}{
		{name: "set", set: std, want: settings.ErrUnsetName, not: settings.ErrUnset},
		{name: "nil set first", set: unsetNil, want: settings.ErrUnset, not: settings.ErrUnsetName},
		{name: "zero set first", set: &settings.Settings{}, want: settings.ErrUnset, not: settings.ErrUnsetName},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, err := tc.set.Expresses(spelling.Spelling{})
			assert.False(t, ok)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, tc.not)

			ok, err = tc.set.ExpressesPitch(pitch.Pitch{})
			assert.False(t, ok)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, tc.not)

			text, err := tc.set.Format(spelling.Spelling{})
			assert.Empty(t, text)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, tc.not)

			text, err = tc.set.FormatPitch(pitch.Pitch{})
			assert.Empty(t, text)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, tc.not)
		})
	}
}

// TestExtremeOctaves covers K-25, P-17a and the table of A 5.7: the exact
// technical conditions of ExpressesPitch at the limits of int. FormatPitch is
// called only where it fails or its result is short (A 5.7, "Huge octaves").
func TestExtremeOctaves(t *testing.T) {
	t.Parallel()
	eu := builtin(notation.EnglishUnicodeID).NameRules()
	trad := system.Traditional()
	helmholtz := setIn(t, trad, styled(t, eu, notation.OctaveStyleHelmholtz))
	english := setIn(t, trad, styled(t, eu, notation.OctaveStyleHelmholtzEnglish))
	numbered := setIn(t, trad, styled(t, eu, notation.OctaveStyleHelmholtzNumbered))
	lily := setIn(t, trad, builtin(notation.LilyPondEnglishID))
	abc := setIn(t, trad, builtin(notation.ABCID))
	std := settings.Standard()
	russian := std.MustWithNotation(builtin(notation.RussianAMID))
	edges := naming.MustNew(map[int]string{math.MaxInt: "top", math.MinInt: "bottom", 4: "four"})
	named := std.MustWithNotation(englishNamed(t)).MustWithCatalog(edges)

	const (
		// The last w with a Helmholtz string of C that fits in int, going down:
		// (2 − w)·len("͵") + len("C") <= MaxInt.
		helmLowC = 3 - 1<<62
		// The same for C♯, written "C♯" (4 bytes).
		helmLowCSharp = 5 - 1<<62
		// The last w going up: (w − 3)·len("′") + len("c") <= MaxInt.
		helmHighC = 3074457345618258605
		// The same for C♯ (4 bytes).
		helmHighCSharp = 3074457345618258604
	)

	c, cs := spelling.C(), spelling.CSharp()
	for _, tc := range []struct {
		name string
		set  *settings.Settings
		p    pitch.Pitch
		want bool
		// short marks a pitch whose formatted string is short, so that
		// FormatPitch may be called to check P-17a.
		short bool
	}{
		{name: "K-25 B sharp with the smallest Octave", set: std, p: at(spelling.BSharp(), math.MinInt), want: false, short: true},
		{name: "B sharp with the second smallest Octave", set: std, p: at(spelling.BSharp(), math.MinInt+1), want: true, short: true},
		{name: "C flat with the largest Octave", set: std, p: at(spelling.CFlat(), math.MaxInt), want: false, short: true},
		{name: "C flat with the second largest Octave", set: std, p: at(spelling.CFlat(), math.MaxInt-1), want: true, short: true},
		{name: "ASPN smallest", set: std, p: at(c, math.MinInt), want: true, short: true},
		{name: "ASPN largest", set: std, p: at(c, math.MaxInt), want: true, short: true},
		{name: "Mexican largest", set: std.MustWithNumbering(number.Mexican()), p: at(c, math.MaxInt), want: false, short: true},
		{name: "Mexican second largest", set: std.MustWithNumbering(number.Mexican()), p: at(c, math.MaxInt-1), want: true, short: true},
		{name: "Mexican smallest", set: std.MustWithNumbering(number.Mexican()), p: at(c, math.MinInt), want: true, short: true},
		{name: "middle C = C5 largest", set: std.MustWithNumbering(number.MiddleC5()), p: at(c, math.MaxInt), want: false, short: true},
		{name: "middle C = C3 smallest", set: std.MustWithNumbering(number.MiddleC3()), p: at(c, math.MinInt), want: false, short: true},
		{name: "middle C = C3 second smallest", set: std.MustWithNumbering(number.MiddleC3()), p: at(c, math.MinInt+1), want: true, short: true},
		{name: "Franco-Belgian smallest", set: std.MustWithNumbering(number.FrancoBelgian()), p: at(c, math.MinInt), want: false, short: true},
		{name: "Franco-Belgian MinInt+1", set: std.MustWithNumbering(number.FrancoBelgian()), p: at(c, math.MinInt+1), want: false, short: true},
		{name: "Franco-Belgian MinInt+2", set: std.MustWithNumbering(number.FrancoBelgian()), p: at(c, math.MinInt+2), want: true, short: true},
		{name: "Franco-Belgian largest", set: std.MustWithNumbering(number.FrancoBelgian()), p: at(c, math.MaxInt), want: true, short: true},
		{name: "AM 0", set: russian, p: at(c, 0), want: true, short: true},
		{name: "AM 8", set: russian, p: at(c, 8), want: true, short: true},
		{name: "AM -1", set: russian, p: at(c, -1), want: false, short: true},
		{name: "AM 9", set: russian, p: at(c, 9), want: false, short: true},
		{name: "AM smallest", set: russian, p: at(c, math.MinInt), want: false, short: true},
		{name: "AM largest", set: russian, p: at(c, math.MaxInt), want: false, short: true},
		{name: "name of the largest octave", set: named, p: at(c, math.MaxInt), want: true, short: true},
		{name: "unnamed second largest octave", set: named, p: at(c, math.MaxInt-1), want: false, short: true},
		{name: "name of the smallest octave", set: named, p: at(c, math.MinInt), want: true, short: true},
		{name: "named octave 4", set: named, p: at(c, 4), want: true, short: true},
		{name: "unnamed octave 5", set: named, p: at(c, 5), want: false, short: true},
		{name: "B sharp in the smallest named octave", set: named, p: at(spelling.BSharp(), math.MinInt+1), want: true, short: true},
		{name: "Helmholtz smallest", set: helmholtz, p: at(c, math.MinInt), want: false, short: true},
		{name: "Helmholtz 2 − w overflows", set: helmholtz, p: at(c, math.MinInt+2), want: false, short: true},
		{name: "Helmholtz length overflows below", set: helmholtz, p: at(c, helmLowC-1), want: false, short: true},
		{name: "Helmholtz last length below", set: helmholtz, p: at(c, helmLowC), want: true},
		{name: "Helmholtz C sharp length overflows below", set: helmholtz, p: at(cs, helmLowCSharp-1), want: false, short: true},
		{name: "Helmholtz C sharp last length below", set: helmholtz, p: at(cs, helmLowCSharp), want: true},
		{name: "Helmholtz last length above", set: helmholtz, p: at(c, helmHighC), want: true},
		{name: "Helmholtz length overflows above", set: helmholtz, p: at(c, helmHighC+1), want: false, short: true},
		{name: "Helmholtz C sharp last length above", set: helmholtz, p: at(cs, helmHighCSharp), want: true},
		{name: "Helmholtz C sharp length overflows above", set: helmholtz, p: at(cs, helmHighCSharp+1), want: false, short: true},
		{name: "Helmholtz largest", set: helmholtz, p: at(c, math.MaxInt), want: false, short: true},
		{name: "Helmholtz ordinary", set: helmholtz, p: at(c, 7), want: true, short: true},
		{name: "English Helmholtz 0", set: english, p: at(c, 0), want: true, short: true},
		{name: "English Helmholtz 1", set: english, p: at(c, 1), want: true, short: true},
		{name: "English Helmholtz 2", set: english, p: at(c, 2), want: true, short: true},
		{name: "English Helmholtz smallest", set: english, p: at(c, math.MinInt), want: false, short: true},
		{name: "English Helmholtz length overflows below", set: english, p: at(c, helmLowC-1), want: false, short: true},
		{name: "English Helmholtz last length below", set: english, p: at(c, helmLowC), want: true},
		{name: "English Helmholtz length overflows above", set: english, p: at(c, helmHighC+1), want: false, short: true},
		{name: "English Helmholtz largest", set: english, p: at(c, math.MaxInt), want: false, short: true},
		{name: "numbered Helmholtz smallest", set: numbered, p: at(c, math.MinInt), want: false, short: true},
		{name: "numbered Helmholtz MinInt+2", set: numbered, p: at(c, math.MinInt+2), want: false, short: true},
		{name: "numbered Helmholtz MinInt+3", set: numbered, p: at(c, math.MinInt+3), want: true, short: true},
		{name: "numbered Helmholtz largest", set: numbered, p: at(c, math.MaxInt), want: true, short: true},
		{name: "LilyPond smallest", set: lily, p: at(c, math.MinInt), want: false, short: true},
		{name: "LilyPond length overflows below", set: lily, p: at(c, math.MinInt+4), want: false, short: true},
		{name: "LilyPond last length below", set: lily, p: at(c, math.MinInt+5), want: true},
		{name: "LilyPond cs length overflows below", set: lily, p: at(cs, math.MinInt+5), want: false, short: true},
		{name: "LilyPond cs last length below", set: lily, p: at(cs, math.MinInt+6), want: true},
		{name: "LilyPond largest", set: lily, p: at(c, math.MaxInt), want: true},
		{name: "ABC smallest", set: abc, p: at(c, math.MinInt), want: false, short: true},
		{name: "ABC length overflows below", set: abc, p: at(c, math.MinInt+5), want: false, short: true},
		{name: "ABC last length below", set: abc, p: at(c, math.MinInt+6), want: true},
		{name: "ABC ^C length overflows below", set: abc, p: at(cs, math.MinInt+6), want: false, short: true},
		{name: "ABC ^C last length below", set: abc, p: at(cs, math.MinInt+7), want: true},
		{name: "ABC 4", set: abc, p: at(c, 4), want: true, short: true},
		{name: "ABC 5", set: abc, p: at(c, 5), want: true, short: true},
		{name: "ABC largest", set: abc, p: at(c, math.MaxInt), want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.set.ExpressesPitch(tc.p)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			if !tc.short {
				return
			}
			text, err := tc.set.FormatPitch(tc.p)
			if tc.want {
				assert.NoError(t, err)
				assert.NotEmpty(t, text)
			} else {
				assert.Error(t, err)
				assert.Empty(t, text)
			}
		})
	}
}

// consistencySets lists sets for the agreement of expressibility, formatting
// and parsing (A 5.7): every built-in notation with several systems, and
// custom notations with the octave styles that no built-in notation has.
func consistencySets(t *testing.T) map[string]*settings.Settings {
	t.Helper()
	eu := builtin(notation.EnglishUnicodeID).NameRules()
	plain := []*system.System{system.Traditional(), system.QuarterTones24(), system.Positions19(), system.ManeriSims72()}
	sets := make(map[string]*settings.Settings)
	add := func(n notation.Notation, systems ...*system.System) {
		for _, sys := range systems {
			s, err := settings.New(tuningFor(t, sys), n, number.ASPN(), naming.Russian())
			require.NoError(t, err)
			sets[fmt.Sprintf("%s in %s", n.Name(), sys.Name())] = s
		}
	}
	for _, id := range notation.BuiltInIDs() {
		if id == notation.ASCIIUpDownID {
			continue
		}
		add(builtin(id), plain...)
	}
	add(builtin(notation.ASCIIUpDownID, 3), system.UpsAndDowns15(), system.UpsAndDowns22())
	add(builtin(notation.ASCIIUpDownID, 4), system.UpsAndDowns41())
	add(builtin(notation.ASCIIUpDownID, 5), system.UpsAndDowns53())
	add(builtin(notation.ASCIIUpDownID, 6), system.ManeriSims72())
	for _, style := range []notation.OctaveStyle{
		notation.OctaveStyleName, notation.OctaveStyleHelmholtz,
		notation.OctaveStyleHelmholtzEnglish, notation.OctaveStyleHelmholtzNumbered,
	} {
		add(styled(t, eu, style), system.Traditional(), system.Positions19())
	}
	return sets
}

// consistencyOctaves are the ASPN numbers of the Octaves checked for the
// agreement of ExpressesPitch and FormatPitch: a safe range with short results.
func consistencyOctaves() []int {
	octaves := []int{-64, -63, -33, 32, 63, 64}
	for o := -12; o <= 12; o++ {
		octaves = append(octaves, o)
	}
	return octaves
}

// TestConsistency covers P-17a and A 5.7: Expresses ⇔ Format succeeds,
// ExpressesPitch ⇔ FormatPitch succeeds (in a safe range of Octaves), and the
// results of successful parsing are expressible.
func TestConsistency(t *testing.T) {
	t.Parallel()
	outside := []spelling.Spelling{
		spelling.C().WithAlteration(alteration.MustFromInt(3)),
		altered(spelling.D(), 1, 2),
		altered(spelling.E(), -1, 3),
		altered(spelling.F(), 1, 7),
		altered(spelling.B(), 5, 2),
	}
	octaves := consistencyOctaves()
	for name, s := range consistencySets(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			names := append(s.Notation().NameRules().Names(), outside...)
			for _, n := range names {
				ok, err := s.Expresses(n)
				require.NoError(t, err)
				text, err := s.Format(n)
				if !assert.Equal(t, ok, err == nil, "Expresses(%v) = %v, Format error %v", n, ok, err) {
					continue
				}
				if ok {
					back, err := s.Parse(text)
					if assert.NoError(t, err, "Parse(%q)", text) {
						assert.Equal(t, n, back)
						assert.True(t, s.MustExpresses(back))
					}
				}
				for _, o := range octaves {
					p := at(n, o)
					okPitch, err := s.ExpressesPitch(p)
					require.NoError(t, err)
					text, err := s.FormatPitch(p)
					if !assert.Equal(t, okPitch, err == nil, "ExpressesPitch(%v) = %v, FormatPitch error %v", p, okPitch, err) {
						continue
					}
					if okPitch {
						assert.True(t, ok, "an expressible pitch has an expressible name")
						back, err := s.ParsePitch(text)
						if assert.NoError(t, err, "ParsePitch(%q)", text) {
							assert.True(t, p.Equal(back), "ParsePitch(%q) = %v, want %v", text, back, p)
							assert.True(t, s.MustExpressesPitch(back))
						}
					}
				}
			}
		})
	}
}

// TestOctaveContext covers A 5.4: the context of the set formats like the set.
func TestOctaveContext(t *testing.T) {
	t.Parallel()
	s := settings.Standard().MustWithNotation(englishNamed(t)).MustWithCatalog(naming.German())
	ctx := s.OctaveContext()
	p := at(spelling.C(), 4)
	want, err := s.FormatPitch(p)
	require.NoError(t, err)
	got, err := s.Notation().FormatPitch(p, s.System(), ctx)
	require.NoError(t, err)
	assert.Equal(t, want, got)

	mexican := settings.Standard().MustWithNumbering(number.Mexican())
	assert.True(t, number.Mexican().Equal(mexican.OctaveContext().Numbering))
	assert.Equal(t, "C♯5", mexican.Notation().MustFormatPitch(at(spelling.CSharp(), 4), mexican.System(), mexican.OctaveContext()))
	assert.Equal(t, "C♯5", mexican.MustFormatPitch(at(spelling.CSharp(), 4)))
	assert.True(t, octave.FromASPN(4).Equal(mexican.MustParsePitch("C♯5").Octave()))
}
