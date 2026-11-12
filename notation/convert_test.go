package notation_test

import (
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func TestConvertName(t *testing.T) {
	t.Parallel()
	german := notation.MustBuiltIn(notation.GermanAMID)
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	lilyEnglish := notation.MustBuiltIn(notation.LilyPondEnglishID)
	abc := notation.MustBuiltIn(notation.ABCID)
	tests := []struct {
		name string
		text string
		from notation.Notation
		to   notation.Notation
		want string
	}{
		{"German canonical", "cis", german, english, "C♯"},
		{"LilyPond long alias", "c-sharp", lilyEnglish, english, "C♯"},
		{"German whole exception", "b", german, english, "B♭"},
		{"English B is natural", "B", english, german, "h"},
		{"ABC natural alias", "=E", abc, english, "E"},
		{"ABC to German", "^C", abc, german, "cis"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			text, err := notation.Convert(tt.text, tt.from, tt.to)
			require.NoError(t, err)
			require.Equal(t, tt.want, text)
			require.Equal(t, tt.want, notation.MustConvert(tt.text, tt.from, tt.to))
		})
	}
}

func TestConvertNameErrors(t *testing.T) {
	t.Parallel()
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	var zero notation.Notation
	for _, tt := range []struct {
		name string
		from notation.Notation
		to   notation.Notation
		want error
	}{
		{"undefined source", zero, english, notation.ErrUndefined},
		{"undefined target", english, zero, notation.ErrUndefined},
		{"unknown source text", english, english, notation.ErrParse},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := "C"
			if tt.name == "unknown source text" {
				input = "not-a-name"
			}
			text, err := notation.Convert(input, tt.from, tt.to)
			require.ErrorIs(t, err, tt.want)
			require.Empty(t, text)
			requirePanicError(t, tt.want, func() { notation.MustConvert(input, tt.from, tt.to) })
		})
	}
	halfRules := explicitNameRules(t, notation.NameForm{Name: nameWithAlteration(letter.C, 1, 2), Text: "half"})
	halfNotation := customNotation(t, halfRules, notation.OctaveStyleNone, "")
	text, err := notation.Convert("half", halfNotation, english)
	require.ErrorIs(t, err, notation.ErrUnwritableName)
	require.Empty(t, text)
}

func TestConvertPitchBetweenNotations(t *testing.T) {
	t.Parallel()
	german := notation.MustBuiltIn(notation.GermanAMID)
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	abc := notation.MustBuiltIn(notation.ABCID)
	aspn := notation.OctaveContext{Numbering: number.ASPN()}
	text, err := notation.ConvertPitch("cis¹", german, english, system.Traditional(), notation.OctaveContext{}, aspn)
	require.NoError(t, err)
	require.Equal(t, "C♯4", text)
	require.Equal(t, "C♯4", notation.MustConvertPitch("cis¹", german, english, system.Traditional(), notation.OctaveContext{}, aspn))

	text, err = notation.ConvertPitch(text, english, abc, system.Traditional(), aspn, notation.OctaveContext{})
	require.NoError(t, err)
	require.Equal(t, "^C", text)
	text, err = notation.ConvertPitch(text, abc, german, system.Traditional(), notation.OctaveContext{}, notation.OctaveContext{})
	require.NoError(t, err)
	require.Equal(t, "cis¹", text)

	helmholtz := customNotation(t, english.NameRules(), notation.OctaveStyleHelmholtz, "")
	text, err = notation.ConvertPitch("c", helmholtz, abc, system.Traditional(), notation.OctaveContext{}, notation.OctaveContext{})
	require.NoError(t, err)
	require.Equal(t, "C,", text)

	franco := notation.OctaveContext{Numbering: number.FrancoBelgian()}
	text, err = notation.ConvertPitch("C4", english, english, system.Traditional(), aspn, franco)
	require.NoError(t, err)
	require.Equal(t, "C3", text)
}

func TestConvertPitchNameCatalog(t *testing.T) {
	t.Parallel()
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	named := customNotation(t, english.NameRules(), notation.OctaveStyleName, " ")
	fromCtx := notation.OctaveContext{Numbering: number.ASPN()}
	toCtx := notation.OctaveContext{Catalog: naming.Russian()}
	text, err := notation.ConvertPitch("C4", english, named, system.Traditional(), fromCtx, toCtx)
	require.NoError(t, err)
	require.Equal(t, "C первая", text)
	parsed, err := named.ParsePitch(text, system.Traditional(), toCtx)
	require.NoError(t, err)
	require.True(t, parsed.Equal(pitch.MustFromASPN(spelling.C(), 4)))
}

func TestConvertPitchErrors(t *testing.T) {
	t.Parallel()
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	none := notation.MustBuiltIn(notation.RomanceItalianID)
	var zero notation.Notation
	aspn := notation.OctaveContext{Numbering: number.ASPN()}
	for _, tt := range []struct {
		name    string
		from    notation.Notation
		to      notation.Notation
		sys     *system.System
		fromCtx notation.OctaveContext
		toCtx   notation.OctaveContext
		want    error
	}{
		{"undefined source", zero, english, system.Traditional(), aspn, aspn, notation.ErrUndefined},
		{"undefined target", english, zero, system.Traditional(), aspn, aspn, notation.ErrUndefined},
		{"missing system", english, english, nil, aspn, aspn, notation.ErrUndefined},
		{"missing source numbering", english, english, system.Traditional(), notation.OctaveContext{}, aspn, notation.ErrUndefined},
		{"missing target numbering", english, english, system.Traditional(), aspn, notation.OctaveContext{}, notation.ErrUndefined},
		{"target has no octave", english, none, system.Traditional(), aspn, aspn, notation.ErrNoOctave},
		{"source text does not match", english, english, system.Traditional(), aspn, aspn, notation.ErrParse},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := "C4"
			if tt.name == "source text does not match" {
				input = "not-a-pitch"
			}
			text, err := notation.ConvertPitch(input, tt.from, tt.to, tt.sys, tt.fromCtx, tt.toCtx)
			require.ErrorIs(t, err, tt.want)
			require.Empty(t, text)
			requirePanicError(t, tt.want, func() {
				notation.MustConvertPitch(input, tt.from, tt.to, tt.sys, tt.fromCtx, tt.toCtx)
			})
		})
	}
}

func TestConvertPitchPreservesComponentEquality(t *testing.T) {
	t.Parallel()
	german := notation.MustBuiltIn(notation.GermanAMID)
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	abc := notation.MustBuiltIn(notation.ABCID)
	aspn := notation.OctaveContext{Numbering: number.ASPN()}
	original := pitch.MustFromASPN(spelling.CSharp(), 4)
	text := notation.MustConvertPitch("cis¹", german, abc, system.Traditional(), notation.OctaveContext{}, notation.OctaveContext{})
	roundTrip := abc.MustParsePitch(text, system.Traditional(), notation.OctaveContext{})
	require.True(t, roundTrip.Equal(original))
	englishText := notation.MustConvertPitch(text, abc, english, system.Traditional(), notation.OctaveContext{}, aspn)
	require.Equal(t, "C♯4", englishText)
}
