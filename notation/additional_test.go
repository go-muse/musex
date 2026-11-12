package notation_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func TestPitchABCMixedMarks(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.ABCID)
	tests := []struct {
		text      string
		name      spelling.Spelling
		w         int
		canonical string
	}{
		{"C,',", spelling.C(), 3, "C,"},
		{"C,", spelling.C(), 3, "C,"},
		{"C'", spelling.C(), 5, "c"},
		{"c", spelling.C(), 5, "c"},
		{"^^C,',", spelling.CDoubleSharp(), 3, "^^C,"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			expected := pitch.MustFromASPN(tt.name, tt.w)
			require.True(t, n.MustParsePitch(tt.text, system.Traditional(), notation.OctaveContext{}).Equal(expected))
			require.Equal(t, tt.canonical, n.MustFormatPitch(expected, system.Traditional(), notation.OctaveContext{}))
		})
	}
}

func TestPitchHelmholtzNumberedAliases(t *testing.T) {
	t.Parallel()
	n := customNotation(t, notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules(), notation.OctaveStyleHelmholtzNumbered, "")
	for _, tt := range []struct {
		text string
		w    int
	}{
		{"2C", 0},
		{"1C", 1},
	} {
		expected := pitch.MustFromASPN(spelling.C(), tt.w)
		require.True(t, n.MustParsePitch(tt.text, system.Traditional(), notation.OctaveContext{}).Equal(expected))
		require.Equal(t, expectedSpecialOctave(notation.OctaveStyleHelmholtzNumbered, "C", tt.w), n.MustFormatPitch(expected, system.Traditional(), notation.OctaveContext{}))
	}
	for _, text := range []string{"2C1", "1C1", "0C", "3C"} {
		_, err := n.ParsePitch(text, system.Traditional(), notation.OctaveContext{})
		require.ErrorIs(t, err, notation.ErrParse)
	}
}

func TestPitchRegisterChangeAfterPrefix(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.ABCID)
	upper := pitch.MustFromASPN(spelling.CSharp(), 4)
	lower := pitch.MustFromASPN(spelling.CSharp(), 5)
	require.Equal(t, "^C", n.MustFormatPitch(upper, system.Traditional(), notation.OctaveContext{}))
	require.Equal(t, "^c", n.MustFormatPitch(lower, system.Traditional(), notation.OctaveContext{}))
	require.True(t, n.MustParsePitch("^c", system.Traditional(), notation.OctaveContext{}).Equal(lower))
}

func TestPitchSameTextDifferentNotationMeaning(t *testing.T) {
	t.Parallel()
	englishRules := notation.MustBuiltIn(notation.EnglishUnicodeID).NameRules()
	abc := notation.MustBuiltIn(notation.ABCID)
	helmholtz := customNotation(t, englishRules, notation.OctaveStyleHelmholtz, "")
	numbered := customNotation(t, englishRules, notation.OctaveStyleHelmholtzNumbered, "")
	require.True(t, abc.MustParsePitch("c", system.Traditional(), notation.OctaveContext{}).Equal(pitch.MustFromASPN(spelling.C(), 5)))
	require.True(t, helmholtz.MustParsePitch("c", system.Traditional(), notation.OctaveContext{}).Equal(pitch.MustFromASPN(spelling.C(), 3)))
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	ctx := notation.OctaveContext{Numbering: number.ASPN()}
	require.True(t, english.MustParsePitch("C2", system.Traditional(), ctx).Equal(pitch.MustFromASPN(spelling.C(), 2)))
	require.True(t, numbered.MustParsePitch("C2", system.Traditional(), notation.OctaveContext{}).Equal(pitch.MustFromASPN(spelling.C(), 0)))
}

func TestNotationGeneratedLettersUseOrdinal(t *testing.T) {
	t.Parallel()
	rules := generatedNameRules(t, basicGeneratedDefinition())
	for _, l := range letter.All() {
		name := spelling.MustNew(l, alteration.Natural())
		require.Equal(t, l.String(), rules.MustFormat(name))
	}
}

func TestNotationMustConstructors(t *testing.T) {
	t.Parallel()
	def := basicGeneratedDefinition()
	rules := notation.MustNewNameRules(def)
	other, err := notation.NewNameRules(def)
	require.NoError(t, err)
	require.True(t, rules.Equal(other))
	n := notation.MustNew(notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleNumber, OctaveSeparator: ":"})
	otherNotation, err := notation.New(notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleNumber, OctaveSeparator: ":"})
	require.NoError(t, err)
	require.True(t, n.Equal(otherNotation))
	updown := notation.MustNewUpsAndDownsNameRules(3)
	otherUpdown, err := notation.NewUpsAndDownsNameRules(3)
	require.NoError(t, err)
	require.True(t, updown.Equal(otherUpdown))
}

func TestNotationHelmholtzEnglishAllowedPairPitchRoundTrip(t *testing.T) {
	t.Parallel()
	rules := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "C͵"},
		notation.NameForm{Name: spelling.D(), Text: "CC"},
	)
	n := customNotation(t, rules, notation.OctaveStyleHelmholtzEnglish, "")
	for _, w := range []int{-1, 0, 1, 2, 3, 4} {
		for _, name := range []spelling.Spelling{spelling.C(), spelling.D()} {
			p := pitch.MustFromASPN(name, w)
			text, err := n.FormatPitch(p, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			require.Equal(t, expectedSpecialOctave(notation.OctaveStyleHelmholtzEnglish, rules.MustFormat(name), w), text)
			parsed, err := n.ParsePitch(text, system.Traditional(), notation.OctaveContext{})
			require.NoError(t, err)
			require.True(t, parsed.Equal(p))
		}
	}
}
