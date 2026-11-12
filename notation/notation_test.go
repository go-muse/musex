package notation_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

func TestNotationDefinitionAccessorsAndEquality(t *testing.T) {
	t.Parallel()
	rules := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "C", Aliases: []string{"see"}},
		notation.NameForm{Name: spelling.D(), Text: "D"},
	)
	first := customNotation(t, rules, notation.OctaveStyleNumber, ":")
	second, err := notation.New(notation.Definition{
		Name:            "debug-name",
		Rules:           rules,
		OctaveStyle:     notation.OctaveStyleNumber,
		OctaveSeparator: ":",
	})
	require.NoError(t, err)
	require.True(t, first.Equal(second))
	require.Empty(t, first.Name())
	require.Equal(t, "debug-name", second.Name())
	require.True(t, second.NameRules().Equal(rules))
	require.Equal(t, notation.OctaveStyleNumber, second.OctaveStyle())
	require.Equal(t, ":", second.OctaveSeparator())
	_, updown := second.ChromaticUnit()
	require.False(t, updown)
	require.True(t, second.IsDefined())
	require.Equal(t, "C", second.MustFormat(spelling.C()))
	require.True(t, second.MustParse("see").Equal(spelling.C()))

	differentStyle := customNotation(t, rules, notation.OctaveStyleName, ":")
	require.False(t, first.Equal(differentStyle))
	differentSeparator := customNotation(t, rules, notation.OctaveStyleNumber, ";")
	require.False(t, first.Equal(differentSeparator))
	updownThree := notation.MustNew(notation.Definition{Rules: notation.MustNewUpsAndDownsNameRules(3), OctaveStyle: notation.OctaveStyleNone})
	updownFour := notation.MustNew(notation.Definition{Rules: notation.MustNewUpsAndDownsNameRules(4), OctaveStyle: notation.OctaveStyleNone})
	require.False(t, updownThree.Equal(updownFour))
	unit, ok := updownThree.ChromaticUnit()
	require.True(t, ok)
	require.Equal(t, 3, unit)

	var zero notation.Notation
	require.False(t, zero.IsDefined())
	require.False(t, zero.Equal(zero))
	require.False(t, zero.Equal(first))
	require.False(t, first.Equal(zero))
	require.Empty(t, zero.Name())
	require.False(t, zero.NameRules().IsDefined())
	require.Equal(t, notation.OctaveStyleUnset, zero.OctaveStyle())
	require.Empty(t, zero.OctaveSeparator())
	unit, ok = zero.ChromaticUnit()
	require.False(t, ok)
	require.Zero(t, unit)
	text, err := zero.Format(spelling.C())
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.Empty(t, text)
	name, err := zero.Parse("C")
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.False(t, name.IsValid())
}

func TestNotationDefinitionErrors(t *testing.T) {
	t.Parallel()
	rules := explicitNameRules(t, notation.NameForm{Name: spelling.C(), Text: "C"})
	noRegister := explicitNameRules(t, notation.NameForm{Name: spelling.C(), Text: "123"})
	tests := []struct {
		name string
		def  notation.Definition
		want error
	}{
		{"zero rules", notation.Definition{OctaveStyle: notation.OctaveStyleNumber}, notation.ErrUndefined},
		{"unset style", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleUnset}, notation.ErrUndefined},
		{"unknown style", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyle(255)}, notation.ErrInvalidDefinition},
		{"none with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleNone, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"AM with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleAM, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"Helmholtz with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleHelmholtz, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"English with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleHelmholtzEnglish, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"numbered with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleHelmholtzNumbered, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"LilyPond with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleLilyPond, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"ABC with separator", notation.Definition{Rules: rules, OctaveStyle: notation.OctaveStyleABC, OctaveSeparator: " "}, notation.ErrInvalidDefinition},
		{"AM without register letter", notation.Definition{Rules: noRegister, OctaveStyle: notation.OctaveStyleAM}, notation.ErrInvalidDefinition},
		{"Helmholtz without register letter", notation.Definition{Rules: noRegister, OctaveStyle: notation.OctaveStyleHelmholtz}, notation.ErrInvalidDefinition},
		{"English without register letter", notation.Definition{Rules: noRegister, OctaveStyle: notation.OctaveStyleHelmholtzEnglish}, notation.ErrInvalidDefinition},
		{"numbered without register letter", notation.Definition{Rules: noRegister, OctaveStyle: notation.OctaveStyleHelmholtzNumbered}, notation.ErrInvalidDefinition},
		{"LilyPond without register letter", notation.Definition{Rules: noRegister, OctaveStyle: notation.OctaveStyleLilyPond}, notation.ErrInvalidDefinition},
		{"ABC without register letter", notation.Definition{Rules: noRegister, OctaveStyle: notation.OctaveStyleABC}, notation.ErrInvalidDefinition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, err := notation.New(tt.def)
			require.ErrorIs(t, err, tt.want)
			require.False(t, n.IsDefined())
			requirePanicError(t, tt.want, func() { notation.MustNew(tt.def) })
		})
	}
}

func TestNotationOctaveStyle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		style notation.OctaveStyle
		text  string
		valid bool
	}{
		{notation.OctaveStyleUnset, "OctaveStyle(unset)", false},
		{notation.OctaveStyleNone, "none", true},
		{notation.OctaveStyleNumber, "number", true},
		{notation.OctaveStyleName, "name", true},
		{notation.OctaveStyleAM, "am", true},
		{notation.OctaveStyleHelmholtz, "helmholtz", true},
		{notation.OctaveStyleHelmholtzEnglish, "helmholtz-english", true},
		{notation.OctaveStyleHelmholtzNumbered, "helmholtz-numbered", true},
		{notation.OctaveStyleLilyPond, "lilypond", true},
		{notation.OctaveStyleABC, "abc", true},
		{notation.OctaveStyle(255), "OctaveStyle(255)", false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.text, tt.style.String())
			require.Equal(t, tt.valid, tt.style.IsValid())
		})
	}
}

func TestNotationFullGrammarAmbiguity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		forms     []notation.NameForm
		style     notation.OctaveStyle
		separator string
	}{
		{
			"number canonical prefix",
			[]notation.NameForm{{Name: spelling.C(), Text: "C+1"}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleNumber,
			"+",
		},
		{
			"number alias prefix",
			[]notation.NameForm{{Name: spelling.C(), Text: "A", Aliases: []string{"C+1"}}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleNumber,
			"+",
		},
		{
			"name canonical prefix",
			[]notation.NameForm{{Name: spelling.C(), Text: "C"}, {Name: spelling.D(), Text: "C sharp"}},
			notation.OctaveStyleName,
			" ",
		},
		{
			"name alias prefix",
			[]notation.NameForm{{Name: spelling.C(), Text: "C"}, {Name: spelling.D(), Text: "D", Aliases: []string{"C sharp"}}},
			notation.OctaveStyleName,
			" ",
		},
		{
			"AM case collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "C"}, {Name: spelling.D(), Text: "c"}},
			notation.OctaveStyleAM,
			"",
		},
		{
			"Helmholtz suffix collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "C"}, {Name: spelling.D(), Text: "C′"}},
			notation.OctaveStyleHelmholtz,
			"",
		},
		{
			"LilyPond suffix collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "C"}, {Name: spelling.D(), Text: "C,"}},
			notation.OctaveStyleLilyPond,
			"",
		},
		{
			"ABC mixed suffix collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "C"}, {Name: spelling.D(), Text: "C,'"}},
			notation.OctaveStyleABC,
			"",
		},
		{
			"Helmholtz English repeat collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "C͵͵͵"}, {Name: spelling.D(), Text: "CC"}},
			notation.OctaveStyleHelmholtzEnglish,
			"",
		},
		{
			"Helmholtz English alias repeat collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "C", Aliases: []string{"C͵͵͵"}}, {Name: spelling.D(), Text: "CC"}},
			notation.OctaveStyleHelmholtzEnglish,
			"",
		},
		{
			"Helmholtz numbered alias collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "2C"}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleHelmholtzNumbered,
			"",
		},
		{
			"AM user alias case collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "A", Aliases: []string{"C"}}, {Name: spelling.D(), Text: "c"}},
			notation.OctaveStyleAM,
			"",
		},
		{
			"Helmholtz user alias suffix collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "A", Aliases: []string{"C′"}}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleHelmholtz,
			"",
		},
		{
			"LilyPond user alias suffix collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "A", Aliases: []string{"C,"}}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleLilyPond,
			"",
		},
		{
			"ABC user alias mixed suffix collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "A", Aliases: []string{"C,'"}}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleABC,
			"",
		},
		{
			"Helmholtz numbered user alias collision",
			[]notation.NameForm{{Name: spelling.C(), Text: "A", Aliases: []string{"2C"}}, {Name: spelling.D(), Text: "C"}},
			notation.OctaveStyleHelmholtzNumbered,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rules := explicitNameRules(t, tt.forms...)
			n, err := notation.New(notation.Definition{Rules: rules, OctaveStyle: tt.style, OctaveSeparator: tt.separator})
			require.ErrorIs(t, err, notation.ErrAmbiguous)
			require.False(t, n.IsDefined())
			requirePanicError(t, notation.ErrAmbiguous, func() {
				notation.MustNew(notation.Definition{Rules: rules, OctaveStyle: tt.style, OctaveSeparator: tt.separator})
			})
		})
	}
}

func TestNotationHelmholtzEnglishAllowedRepeatPair(t *testing.T) {
	t.Parallel()
	rules := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "C͵"},
		notation.NameForm{Name: spelling.D(), Text: "CC"},
	)
	n := customNotation(t, rules, notation.OctaveStyleHelmholtzEnglish, "")
	for _, w := range []int{-1, 0, 1, 2, 3, 4} {
		for _, name := range []spelling.Spelling{spelling.C(), spelling.D()} {
			base := rules.MustFormat(name)
			require.NotEmpty(t, expectedSpecialOctave(n.OctaveStyle(), base, w))
		}
	}
}

func TestNotationMultipleAliasesDoNotCreateFalseAmbiguity(t *testing.T) {
	t.Parallel()
	rules := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "C", Aliases: []string{"see", "cee"}},
		notation.NameForm{Name: spelling.D(), Text: "D"},
	)
	n := customNotation(t, rules, notation.OctaveStyleNumber, "")
	ctx := notation.OctaveContext{Numbering: number.ASPN()}
	p := pitch.MustFromASPN(spelling.C(), 4)
	require.Equal(t, "C4", n.MustFormatPitch(p, system.Traditional(), ctx))
	for _, alias := range []string{"see4", "cee4"} {
		parsed, err := n.ParsePitch(alias, system.Traditional(), ctx)
		require.NoError(t, err)
		require.True(t, parsed.Equal(p))
	}
	require.Equal(t, "C", n.MustFormat(spelling.C()))
	require.True(t, n.MustParse("see").Equal(spelling.C()))
	require.True(t, n.MustParse("cee").Equal(spelling.C()))
}
