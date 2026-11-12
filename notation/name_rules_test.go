package notation_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/require"
)

func basicGeneratedDefinition() notation.NameRulesDefinition {
	return notation.NameRulesDefinition{
		Letters: [letter.Count]string{"C", "D", "E", "F", "G", "A", "B"},
		Alterations: []notation.AlterationForm{
			{Value: alteration.Natural(), Text: "", Aliases: []string{"♮", "n"}},
			{Value: alteration.Sharp(), Text: "#", Aliases: []string{"♯", "s"}},
			{Value: alteration.Flat(), Text: "b", Aliases: []string{"♭"}},
		},
		Position: notation.AlterationSuffix,
	}
}

func TestNameRulesGeneratedDomainExceptionsAndOwnership(t *testing.T) {
	t.Parallel()
	def := basicGeneratedDefinition()
	def.Names = []notation.NameForm{
		{Name: spelling.BFlat(), Text: "bb", Aliases: []string{"flat-b", "si-flat"}},
		{Name: spelling.CSharp(), Text: "C-sharp", Aliases: []string{"C♯"}},
	}
	def.Omit = []spelling.Spelling{spelling.EDoubleSharp()}
	def.Alterations = append(def.Alterations,
		notation.AlterationForm{Value: alteration.DoubleSharp(), Text: "x"},
		notation.AlterationForm{Value: alteration.DoubleFlat(), Text: "bb"},
	)
	rules := generatedNameRules(t, def)
	require.True(t, rules.IsDefined())
	_, updown := rules.ChromaticUnit()
	require.False(t, updown)
	require.Len(t, rules.Names(), 7*5-1)
	require.Equal(t, "C-sharp", rules.MustFormat(spelling.CSharp()))
	require.Equal(t, "bb", rules.MustFormat(spelling.BFlat()))
	for _, alias := range []string{"flat-b", "si-flat"} {
		require.True(t, rules.MustParse(alias).Equal(spelling.BFlat()))
	}
	for _, alias := range []string{"C♮", "Cn"} {
		require.True(t, rules.MustParse(alias).Equal(spelling.C()))
	}
	for _, alias := range []string{"C♯", "Cs"} {
		require.True(t, rules.MustParse(alias).Equal(spelling.CSharp()))
	}
	require.Equal(t, "C-sharp", rules.MustFormat(spelling.CSharp()))

	names := rules.Names()
	names[0] = spelling.BDoubleSharp()
	require.NotEqual(t, names, rules.Names())

	def.Letters[0] = "X"
	def.Alterations[0].Text = "changed"
	def.Alterations[0].Aliases[0] = "changed"
	def.Names[0].Text = "changed"
	def.Names[0].Aliases[0] = "changed"
	def.Omit[0] = spelling.C()
	require.Equal(t, "C", rules.MustFormat(spelling.C()))
	require.Equal(t, "bb", rules.MustFormat(spelling.BFlat()))
	require.True(t, rules.MustParse("C♮").Equal(spelling.C()))
	require.True(t, rules.MustParse("flat-b").Equal(spelling.BFlat()))
}

func TestNameRulesExplicitOrderEqualityAndCopies(t *testing.T) {
	t.Parallel()
	first := explicitNameRules(t,
		notation.NameForm{Name: spelling.CSharp(), Text: "sharp-c", Aliases: []string{"sc", "c#"}},
		notation.NameForm{Name: spelling.C(), Text: "see", Aliases: []string{"natural-c", "cee"}},
	)
	second := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "see", Aliases: []string{"cee", "natural-c"}},
		notation.NameForm{Name: spelling.CSharp(), Text: "sharp-c", Aliases: []string{"c#", "sc"}},
	)
	require.True(t, first.Equal(second))
	require.True(t, second.Equal(first))
	require.Equal(t, []spelling.Spelling{spelling.C(), spelling.CSharp()}, first.Names())
	require.False(t, first.Equal(notation.NameRules{}))
	require.False(t, (notation.NameRules{}).Equal(first))
	require.False(t, (notation.NameRules{}).Equal(notation.NameRules{}))

	differentText := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "different", Aliases: []string{"cee", "natural-c"}},
		notation.NameForm{Name: spelling.CSharp(), Text: "sharp-c", Aliases: []string{"c#", "sc"}},
	)
	require.False(t, first.Equal(differentText))
	differentAlias := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "see", Aliases: []string{"cee", "other"}},
		notation.NameForm{Name: spelling.CSharp(), Text: "sharp-c", Aliases: []string{"c#", "sc"}},
	)
	require.False(t, first.Equal(differentAlias))
}

func TestNameRulesDefinitionErrors(t *testing.T) {
	t.Parallel()
	validExplicit := []notation.NameForm{{Name: spelling.C(), Text: "C"}}
	tests := []struct {
		name string
		def  notation.NameRulesDefinition
		want error
	}{
		{"empty definition", notation.NameRulesDefinition{}, notation.ErrInvalidDefinition},
		{"partial letters", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Letters[3] = ""
			return d
		}(), notation.ErrInvalidDefinition},
		{"no alterations", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Alterations = nil
			return d
		}(), notation.ErrInvalidDefinition},
		{"unset position", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Position = notation.AlterationPositionUnset
			return d
		}(), notation.ErrInvalidDefinition},
		{"unknown position", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Position = notation.AlterationPosition(255)
			return d
		}(), notation.ErrInvalidDefinition},
		{"exception collides with generated canonical", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Names = []notation.NameForm{{Name: spelling.D(), Text: "C"}}
			return d
		}(), notation.ErrAmbiguous},
		{"explicit with position", notation.NameRulesDefinition{Names: validExplicit, Position: notation.AlterationSuffix}, notation.ErrInvalidDefinition},
		{"explicit with letters", func() notation.NameRulesDefinition {
			d := notation.NameRulesDefinition{Names: validExplicit}
			d.Letters[0] = "C"
			return d
		}(), notation.ErrInvalidDefinition},
		{"empty canonical", notation.NameRulesDefinition{Names: []notation.NameForm{{Name: spelling.C()}}}, notation.ErrInvalidDefinition},
		{"empty alias", notation.NameRulesDefinition{Names: []notation.NameForm{{Name: spelling.C(), Text: "C", Aliases: []string{""}}}}, notation.ErrInvalidDefinition},
		{"duplicate alteration", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Alterations[1] = d.Alterations[0]
			return d
		}(), notation.ErrInvalidDefinition},
		{"duplicate name", notation.NameRulesDefinition{Names: []notation.NameForm{validExplicit[0], validExplicit[0]}}, notation.ErrInvalidDefinition},
		{"duplicate alias at one name", notation.NameRulesDefinition{Names: []notation.NameForm{{Name: spelling.C(), Text: "C", Aliases: []string{"c", "c"}}}}, notation.ErrInvalidDefinition},
		{"unknown omit", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Omit = []spelling.Spelling{spelling.CDoubleSharp()}
			return d
		}(), notation.ErrInvalidDefinition},
		{"duplicate omit", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Omit = []spelling.Spelling{spelling.C(), spelling.C()}
			return d
		}(), notation.ErrInvalidDefinition},
		{"omit and exception", func() notation.NameRulesDefinition {
			d := basicGeneratedDefinition()
			d.Omit = []spelling.Spelling{spelling.C()}
			d.Names = []notation.NameForm{{Name: spelling.C(), Text: "see"}}
			return d
		}(), notation.ErrInvalidDefinition},
		{"same canonical for two names", notation.NameRulesDefinition{Names: []notation.NameForm{
			{Name: spelling.C(), Text: "same"},
			{Name: spelling.D(), Text: "same"},
		}}, notation.ErrAmbiguous},
		{"alias is another canonical", notation.NameRulesDefinition{Names: []notation.NameForm{
			{Name: spelling.C(), Text: "C", Aliases: []string{"D"}},
			{Name: spelling.D(), Text: "D"},
		}}, notation.ErrAmbiguous},
		{"alias of two names", notation.NameRulesDefinition{Names: []notation.NameForm{
			{Name: spelling.C(), Text: "C", Aliases: []string{"same"}},
			{Name: spelling.D(), Text: "D", Aliases: []string{"same"}},
		}}, notation.ErrAmbiguous},
		{"canonical also alias of another", notation.NameRulesDefinition{Names: []notation.NameForm{
			{Name: spelling.C(), Text: "C"},
			{Name: spelling.D(), Text: "D", Aliases: []string{"C"}},
		}}, notation.ErrAmbiguous},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rules, err := notation.NewNameRules(tt.def)
			require.ErrorIs(t, err, tt.want)
			require.False(t, rules.IsDefined())
			requirePanicError(t, tt.want, func() { notation.MustNewNameRules(tt.def) })
		})
	}
}

func TestNameRulesValidTwoAliasesAndInvalidUTF8(t *testing.T) {
	t.Parallel()
	invalidUTF8 := string([]byte{0xff, 0xfe})
	rules := explicitNameRules(t,
		notation.NameForm{Name: spelling.C(), Text: "C", Aliases: []string{"first", "second"}},
		notation.NameForm{Name: spelling.D(), Text: invalidUTF8},
	)
	require.True(t, rules.MustParse("first").Equal(spelling.C()))
	require.True(t, rules.MustParse("second").Equal(spelling.C()))
	require.Equal(t, "C", rules.MustFormat(spelling.C()))
	require.Equal(t, invalidUTF8, rules.MustFormat(spelling.D()))
	require.True(t, rules.MustParse(invalidUTF8).Equal(spelling.D()))
}

func TestNameRulesZeroValue(t *testing.T) {
	t.Parallel()
	var rules notation.NameRules
	require.False(t, rules.IsDefined())
	require.Nil(t, rules.Names())
	unit, ok := rules.ChromaticUnit()
	require.False(t, ok)
	require.Zero(t, unit)
	text, err := rules.Format(spelling.C())
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.Empty(t, text)
	name, err := rules.Parse("C")
	require.ErrorIs(t, err, notation.ErrUndefined)
	require.False(t, name.IsValid())
	requirePanicError(t, notation.ErrUndefined, func() { rules.MustFormat(spelling.C()) })
	requirePanicError(t, notation.ErrUndefined, func() { rules.MustParse("C") })
}

func TestNameRulesUpsAndDownsConstructors(t *testing.T) {
	t.Parallel()
	for _, c := range []int{-1, 0, 2, 65_537} {
		t.Run(strconvItoa(c), func(t *testing.T) {
			t.Parallel()
			rules, err := notation.NewUpsAndDownsNameRules(c)
			require.ErrorIs(t, err, notation.ErrInvalidChromaticUnit)
			require.False(t, rules.IsDefined())
			requirePanicError(t, notation.ErrInvalidChromaticUnit, func() { notation.MustNewUpsAndDownsNameRules(c) })
		})
	}
	first := notation.MustNewUpsAndDownsNameRules(3)
	second := notation.MustNewUpsAndDownsNameRules(4)
	require.True(t, first.Equal(notation.MustNewUpsAndDownsNameRules(3)))
	require.False(t, first.Equal(second))
	require.Len(t, notation.MustNewUpsAndDownsNameRules(65_536).Names(), 105)
}

func TestNameRulesAndNotationAreNotComparable(t *testing.T) {
	t.Parallel()
	require.False(t, reflect.TypeFor[notation.NameRules]().Comparable())
	require.False(t, reflect.TypeFor[notation.Notation]().Comparable())
}
