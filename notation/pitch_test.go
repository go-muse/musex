package notation_test

import (
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func TestPitchSpecialOctaveStylesRoundTrip(t *testing.T) {
	t.Parallel()
	rules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	for _, style := range []notation.OctaveStyle{
		notation.OctaveStyleAM,
		notation.OctaveStyleHelmholtz,
		notation.OctaveStyleHelmholtzEnglish,
		notation.OctaveStyleHelmholtzNumbered,
		notation.OctaveStyleLilyPond,
		notation.OctaveStyleABC,
	} {
		t.Run(style.String(), func(t *testing.T) {
			t.Parallel()
			n := customNotation(t, rules, style, "")
			for _, w := range []int{-3, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 12} {
				if style == notation.OctaveStyleAM && (w < 0 || w > 8) {
					continue
				}
				t.Run(strconv.Itoa(w), func(t *testing.T) {
					p := pitch.MustFromASPN(spelling.C(), w)
					expected := expectedSpecialOctave(style, "C", w)
					text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
					require.NoError(t, err)
					require.Equal(t, expected, text)
					require.Equal(t, expected, n.MustFormatPitch(p, system.Traditional(), notation.OctaveContext{}))
					parsed, err := n.ParsePitch(expected, system.Traditional(), notation.OctaveContext{})
					require.NoError(t, err)
					require.True(t, parsed.Equal(p))
					require.True(t, n.MustParsePitch(expected, system.Traditional(), notation.OctaveContext{}).Equal(p))
				})
			}
		})
	}
}

func TestPitchOctaveMarksFixture(t *testing.T) {
	t.Parallel()
	styleByFixture := map[string]notation.OctaveStyle{
		"aspn":               notation.OctaveStyleNumber,
		"am":                 notation.OctaveStyleAM,
		"helmholtz":          notation.OctaveStyleHelmholtz,
		"helmholtz-en":       notation.OctaveStyleHelmholtzEnglish,
		"helmholtz-numbered": notation.OctaveStyleHelmholtzNumbered,
		"lilypond":           notation.OctaveStyleLilyPond,
		"abc":                notation.OctaveStyleABC,
	}
	rules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	for _, row := range readFixture(t, octaveMarksFixture) {
		fixtureStyle, aspnText, expected := row[0], row[1], row[2]
		style := styleByFixture[fixtureStyle]
		w, err := strconv.Atoi(aspnText)
		require.NoError(t, err)
		t.Run(fixtureStyle+"/"+aspnText, func(t *testing.T) {
			t.Parallel()
			n := customNotation(t, rules, style, "")
			ctx := notation.OctaveContext{Numbering: number.ASPN()}
			p := pitch.MustFromASPN(spelling.C(), w)
			require.Equal(t, expected, n.MustFormatPitch(p, system.Traditional(), ctx))
			require.True(t, n.MustParsePitch(expected, system.Traditional(), ctx).Equal(p))
		})
	}
}

func TestPitchAMOutsideRange(t *testing.T) {
	t.Parallel()
	n := customNotation(t, notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), notation.OctaveStyleAM, "")
	for _, w := range []int{-1, 9} {
		p := pitch.MustFromASPN(spelling.C(), w)
		text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
		require.ErrorIs(t, err, notation.ErrNoOctave)
		require.Empty(t, text)
		requirePanicError(t, notation.ErrNoOctave, func() { n.MustFormatPitch(p, system.Traditional(), notation.OctaveContext{}) })
	}
}

func TestPitchVisibleNumbering(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	tests := []struct {
		name      string
		numbering number.Numbering
		w         int
		text      string
	}{
		{"ASPN", number.ASPN(), 4, "C4"},
		{"Franco-Belgian", number.FrancoBelgian(), 4, "C3"},
		{"Middle C5", number.MiddleC5(), 4, "C5"},
		{"negative", number.ASPN(), -12, "C-12"},
		{"large", number.ASPN(), 12345, "C12345"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := notation.OctaveContext{Numbering: tt.numbering}
			p := pitch.MustFromASPN(spelling.C(), tt.w)
			require.Equal(t, tt.text, n.MustFormatPitch(p, system.Traditional(), ctx))
			require.True(t, n.MustParsePitch(tt.text, system.Traditional(), ctx).Equal(p))
		})
	}
}

func TestPitchStrictNumbersAndNumberErrors(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	for _, text := range []string{"C04", "C+4", "C-0", "C", "C 4", "C4x"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			p, err := n.ParsePitch(text, system.Traditional(), notation.OctaveContext{Numbering: number.ASPN()})
			require.ErrorIs(t, err, notation.ErrParse)
			require.False(t, p.IsValid())
		})
	}
	p, err := n.ParsePitch("C999999999999999999999999999", system.Traditional(), notation.OctaveContext{Numbering: number.ASPN()})
	require.ErrorIs(t, err, number.ErrOutOfRange)
	require.False(t, p.IsValid())

	p, err = n.ParsePitch("C0", system.Traditional(), notation.OctaveContext{Numbering: number.FrancoBelgian()})
	require.ErrorIs(t, err, number.ErrSkipped)
	require.False(t, p.IsValid())
	requirePanicError(t, number.ErrSkipped, func() {
		n.MustParsePitch("C0", system.Traditional(), notation.OctaveContext{Numbering: number.FrancoBelgian()})
	})

	_, err = n.FormatPitch(pitch.MustFromASPN(spelling.C(), math.MaxInt), system.Traditional(), notation.OctaveContext{Numbering: number.MiddleC5()})
	require.ErrorIs(t, err, number.ErrOutOfRange)
}

func TestPitchRequiredInputs(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	p := pitch.MustFromASPN(spelling.C(), 4)
	for _, s := range []*system.System{nil, new(system.System)} {
		text, err := n.FormatPitch(p, s, notation.OctaveContext{Numbering: number.ASPN()})
		require.ErrorIs(t, err, notation.ErrUndefined)
		require.Empty(t, text)
		parsed, err := n.ParsePitch("C4", s, notation.OctaveContext{Numbering: number.ASPN()})
		require.ErrorIs(t, err, notation.ErrUndefined)
		require.False(t, parsed.IsValid())
	}
	text, err := n.FormatPitch(pitch.Pitch{}, system.Traditional(), notation.OctaveContext{Numbering: number.ASPN()})
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.Empty(t, text)
	text, err = n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.Empty(t, text)
	parsed, err := n.ParsePitch("C4", system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.False(t, parsed.IsValid())
	var zero notation.Notation
	parsed, err = zero.ParsePitch("C4", system.Traditional(), notation.OctaveContext{Numbering: number.ASPN()})
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.False(t, parsed.IsValid())
}

func TestPitchWrittenOctaveVersusPitchOctave(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	ctx := notation.OctaveContext{Numbering: number.ASPN()}
	tests := []struct {
		name spelling.Spelling
		oct  int
		text string
		sys  *system.System
	}{
		{spelling.BSharp(), 4, "B♯3", system.Traditional()},
		{spelling.CFlat(), 3, "C♭4", system.Traditional()},
		{spelling.BSharp(), 5, "B♯5", system.Positions19()},
	}
	for _, tt := range tests {
		p := pitch.MustFromASPN(tt.name, tt.oct)
		require.Equal(t, tt.text, n.MustFormatPitch(p, tt.sys, ctx))
		require.True(t, n.MustParsePitch(tt.text, tt.sys, ctx).Equal(p))
	}
}

func TestPitchNameCatalogAndSeparator(t *testing.T) {
	t.Parallel()
	rules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	n := customNotation(t, rules, notation.OctaveStyleName, " ")
	ctx := notation.OctaveContext{Catalog: naming.Russian()}
	p := pitch.MustFromASPN(spelling.C(), 4)
	require.Equal(t, "C первая", n.MustFormatPitch(p, system.Traditional(), ctx))
	require.True(t, n.MustParsePitch("C первая", system.Traditional(), ctx).Equal(p))

	_, err := n.FormatPitch(pitch.MustFromASPN(spelling.C(), 10), system.Traditional(), ctx)
	require.ErrorIs(t, err, notation.ErrNoOctave)
	parsed, err := n.ParsePitch("C неизвестно", system.Traditional(), ctx)
	require.ErrorIs(t, err, naming.ErrUnknownName)
	require.False(t, parsed.IsValid())

	custom := customNotation(t, rules, notation.OctaveStyleName, "::")
	require.Equal(t, "C::первая", custom.MustFormatPitch(p, system.Traditional(), ctx))
	require.True(t, custom.MustParsePitch("C::первая", system.Traditional(), ctx).Equal(p))
	for _, text := range []string{"C первая", "C:первая", "C:: Первая"} {
		_, err := custom.ParsePitch(text, system.Traditional(), ctx)
		require.Error(t, err)
	}
}

func TestPitchEmptyCatalogForms(t *testing.T) {
	t.Parallel()
	rules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	var zero naming.Catalog
	emptyFromNew, err := naming.New(nil)
	require.NoError(t, err)
	for _, separator := range []string{" ", ""} {
		n := customNotation(t, rules, notation.OctaveStyleName, separator)
		for _, catalog := range []*naming.Catalog{nil, &zero, emptyFromNew} {
			ctx := notation.OctaveContext{Catalog: catalog}
			p := pitch.MustFromASPN(spelling.C(), 4)
			text, err := n.FormatPitch(p, system.Traditional(), ctx)
			require.ErrorIs(t, err, notation.ErrNoOctave)
			require.Empty(t, text)
			input := "C" + separator
			parsed, err := n.ParsePitch(input, system.Traditional(), ctx)
			require.ErrorIs(t, err, naming.ErrUnknownName)
			require.False(t, parsed.IsValid())
			require.False(t, errorsIs(err, notation.ErrParse))
		}
	}
}

func errorsIs(err error, target error) bool {
	return errors.Is(err, target)
}

func TestPitchNoneStyleRejectsOperations(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.RomanceItalianID)
	p := pitch.MustFromASPN(spelling.C(), 4)
	text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, notation.ErrNoOctave)
	require.Empty(t, text)
	parsed, err := n.ParsePitch("do", system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, notation.ErrNoOctave)
	require.False(t, parsed.IsValid())
}

func TestPitchSystemErrors(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.LilyPondEnglishID)
	p := pitch.MustFromASPN(nameWithAlteration(letter.C, 1, 2), 4)
	text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, system.ErrInadmissible)
	require.Empty(t, text)
	parsed, err := n.ParsePitch("cqs'", system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, system.ErrInadmissible)
	require.False(t, parsed.IsValid())

	numbered := customNotation(t, notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), notation.OctaveStyleHelmholtzNumbered, "")
	text = "b♯" + strconv.Itoa(math.MaxInt-3)
	parsed, err = numbered.ParsePitch(text, system.Traditional(), notation.OctaveContext{})
	require.ErrorIs(t, err, system.ErrOutOfRange)
	require.False(t, parsed.IsValid())
}

func TestPitchUpDownChromaticCompatibility(t *testing.T) {
	t.Parallel()
	rules := notation.MustNewUpsAndDownsNameRules(4)
	n := customNotation(t, rules, notation.OctaveStyleNumber, "")
	ctx := notation.OctaveContext{Numbering: number.ASPN()}
	p := pitch.MustFromASPN(nameWithAlteration(letter.C, 1, 4), 4)
	require.Equal(t, "^C4", n.MustFormatPitch(p, system.UpsAndDowns41(), ctx))
	require.True(t, n.MustParsePitch("^C4", system.UpsAndDowns41(), ctx).Equal(p))
	for _, incompatible := range []*system.System{system.UpsAndDowns22(), system.Traditional()} {
		text, err := n.FormatPitch(p, incompatible, ctx)
		require.ErrorIs(t, err, notation.ErrIncompatibleSystem)
		require.Empty(t, text)
		parsed, err := n.ParsePitch("^C4", incompatible, ctx)
		require.ErrorIs(t, err, notation.ErrIncompatibleSystem)
		require.False(t, parsed.IsValid())
	}
}

func TestPitchCheckedOctaveArithmetic(t *testing.T) {
	t.Parallel()
	for _, style := range []notation.OctaveStyle{
		notation.OctaveStyleHelmholtz,
		notation.OctaveStyleHelmholtzEnglish,
		notation.OctaveStyleHelmholtzNumbered,
		notation.OctaveStyleLilyPond,
		notation.OctaveStyleABC,
	} {
		t.Run(style.String()+"/MinInt", func(t *testing.T) {
			t.Parallel()
			n := customNotation(t, notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), style, "")
			p := pitch.MustFromASPN(spelling.C(), math.MinInt)
			require.NotPanics(t, func() {
				text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
				require.ErrorIs(t, err, octave.ErrOutOfRange)
				require.Empty(t, text)
			})
		})
	}

	numbered := customNotation(t, notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), notation.OctaveStyleHelmholtzNumbered, "")
	p := pitch.MustFromASPN(spelling.C(), math.MaxInt)
	expected := "c" + strconv.Itoa(math.MaxInt-3)
	require.Equal(t, expected, numbered.MustFormatPitch(p, system.Traditional(), notation.OctaveContext{}))
	require.True(t, numbered.MustParsePitch(expected, system.Traditional(), notation.OctaveContext{}).Equal(p))

	for _, tt := range []struct {
		style notation.OctaveStyle
		text  string
	}{
		{notation.OctaveStyleHelmholtz, "CCCC"},
		{notation.OctaveStyleHelmholtzEnglish, "CCCC"},
		{notation.OctaveStyleLilyPond, "CCCC"},
		{notation.OctaveStyleABC, "CCCCCC"},
	} {
		t.Run(tt.style.String()+"/MaxIntLength", func(t *testing.T) {
			t.Parallel()
			rules := explicitNameRules(t, notation.NameForm{Name: spelling.C(), Text: tt.text})
			n := customNotation(t, rules, tt.style, "")
			require.NotPanics(t, func() {
				text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
				require.ErrorIs(t, err, octave.ErrOutOfRange)
				require.Empty(t, text)
			})
		})
	}
}
