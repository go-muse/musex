package notation_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/require"
)

var builtInIDOrder = []string{
	notation.EnglishUnicodeID,
	notation.EnglishASCIIMT21ID,
	notation.GermanAMID,
	notation.RussianAMID,
	notation.RomanceItalianID,
	notation.ABCID,
	notation.LilyPondNederlandsID,
	notation.LilyPondEnglishID,
	notation.UnicodeSteinID,
	notation.UnicodeGouldID,
	notation.LilyPondItalianoID,
	notation.UnicodeSimsID,
	notation.UnicodeWyschnegradskyID,
	notation.UnicodeAEUID,
	notation.UnicodePersianID,
	notation.UnicodeArabicID,
	notation.UnicodeStockhausenID,
	notation.ASCIIUpDownID,
	notation.ASCIIFokkerID,
}

var builtInDomainCounts = map[string]int{
	notation.EnglishUnicodeID:        35,
	notation.EnglishASCIIMT21ID:      28,
	notation.GermanAMID:              33,
	notation.RussianAMID:             35,
	notation.RomanceItalianID:        35,
	notation.ABCID:                   35,
	notation.LilyPondNederlandsID:    63,
	notation.LilyPondEnglishID:       63,
	notation.UnicodeSteinID:          63,
	notation.UnicodeGouldID:          77,
	notation.LilyPondItalianoID:      63,
	notation.UnicodeSimsID:           77,
	notation.UnicodeWyschnegradskyID: 175,
	notation.UnicodeAEUID:            77,
	notation.UnicodePersianID:        49,
	notation.UnicodeArabicID:         63,
	notation.UnicodeStockhausenID:    56,
	notation.ASCIIFokkerID:           63,
}

func TestBuiltInIDs(t *testing.T) {
	t.Parallel()
	ids := notation.BuiltInIDs()
	require.Equal(t, builtInIDOrder, ids)
	ids[0] = "changed"
	require.Equal(t, builtInIDOrder, notation.BuiltInIDs())
}

func TestBuiltInRecordsAndNameRoundTrip(t *testing.T) {
	t.Parallel()
	for _, expected := range expectedBuiltIns(t) {
		t.Run(expected.id, func(t *testing.T) {
			t.Parallel()
			n, err := notation.BuiltIn(expected.id)
			require.NoError(t, err)
			require.True(t, n.IsDefined())
			require.Equal(t, expected.id, n.Name())
			require.Equal(t, expected.style, n.OctaveStyle())
			require.Empty(t, n.OctaveSeparator())
			_, parameterized := n.ChromaticUnit()
			require.False(t, parameterized)
			rules := n.NameRules()
			require.True(t, rules.IsDefined())
			require.Equal(t, builtInDomainCounts[expected.id], len(expected.forms))
			require.Equal(t, expected.names(), rules.Names())
			require.Equal(t, expected.names(), sortedNames(expected.forms))

			unique := make(map[string]spelling.Spelling, len(expected.forms))
			for _, name := range expected.names() {
				canonical := expected.forms[name]
				unique[canonical] = name
				actual, err := n.Format(name)
				require.NoError(t, err)
				require.Equal(t, canonical, actual)
				require.Equal(t, canonical, n.MustFormat(name))
				fromRules, err := rules.Format(name)
				require.NoError(t, err)
				require.Equal(t, canonical, fromRules)
				require.Equal(t, canonical, rules.MustFormat(name))

				parsed, err := n.Parse(canonical)
				require.NoError(t, err)
				require.True(t, parsed.Equal(name))
				require.True(t, n.MustParse(canonical).Equal(name))
				parsed, err = rules.Parse(canonical)
				require.NoError(t, err)
				require.True(t, parsed.Equal(name))
				require.True(t, rules.MustParse(canonical).Equal(name))
			}
			require.Len(t, unique, len(expected.forms))

			aliases := make([]string, 0, len(expected.aliases))
			for alias := range expected.aliases {
				aliases = append(aliases, alias)
			}
			slices.Sort(aliases)
			for _, alias := range aliases {
				name := expected.aliases[alias]
				parsed, err := n.Parse(alias)
				require.NoError(t, err, "%q", alias)
				require.True(t, parsed.Equal(name), "%q", alias)
				require.True(t, rules.MustParse(alias).Equal(name), "%q", alias)
				require.Equal(t, expected.forms[name], n.MustFormat(name), "%q", alias)
			}
		})
	}
}

func sortedNames(forms map[spelling.Spelling]string) []spelling.Spelling {
	names := make([]spelling.Spelling, 0, len(forms))
	for name := range forms {
		names = append(names, name)
	}
	slices.SortFunc(names, spelling.Spelling.Compare)
	return names
}

func TestBuiltInUpDownDomainAndValues(t *testing.T) {
	t.Parallel()
	for _, c := range []int{3, 4, 65_536} {
		t.Run(strconvItoa(c), func(t *testing.T) {
			t.Parallel()
			expected := expectedUpDown(t, c)
			n, err := notation.BuiltIn(notation.ASCIIUpDownID, c)
			require.NoError(t, err)
			require.Equal(t, notation.ASCIIUpDownID, n.Name())
			require.Equal(t, notation.OctaveStyleNone, n.OctaveStyle())
			unit, ok := n.ChromaticUnit()
			require.True(t, ok)
			require.Equal(t, c, unit)
			rules := n.NameRules()
			unit, ok = rules.ChromaticUnit()
			require.True(t, ok)
			require.Equal(t, c, unit)
			require.Len(t, rules.Names(), 105)
			require.Equal(t, expected.names(), rules.Names())
			unique := make(map[string]bool, 105)
			for _, name := range expected.names() {
				canonical := expected.forms[name]
				unique[canonical] = true
				require.Equal(t, canonical, n.MustFormat(name))
				require.True(t, n.MustParse(canonical).Equal(name))
				require.True(t, rules.MustParse(canonical).Equal(name))
			}
			require.Len(t, unique, 105)
			for alias, name := range expected.aliases {
				require.True(t, n.MustParse(alias).Equal(name))
			}
		})
	}
}

func strconvItoa(value int) string {
	return fmtSprint(value)
}

func fmtSprint(value int) string {
	return fmt.Sprintf("%d", value)
}

func TestBuiltInLookupErrorsAndMust(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		id   string
		c    []int
		want error
	}{
		{"unknown id", "unknown", nil, notation.ErrUnknownID},
		{"empty id", "", nil, notation.ErrUnknownID},
		{"argument on fixed notation", notation.EnglishUnicodeID, []int{3}, notation.ErrInvalidChromaticUnit},
		{"missing updown unit", notation.ASCIIUpDownID, nil, notation.ErrMissingChromaticUnit},
		{"negative unit", notation.ASCIIUpDownID, []int{-1}, notation.ErrInvalidChromaticUnit},
		{"zero unit", notation.ASCIIUpDownID, []int{0}, notation.ErrInvalidChromaticUnit},
		{"ambiguous unit two", notation.ASCIIUpDownID, []int{2}, notation.ErrInvalidChromaticUnit},
		{"unit above maximum", notation.ASCIIUpDownID, []int{65_537}, notation.ErrInvalidChromaticUnit},
		{"too many units", notation.ASCIIUpDownID, []int{3, 4}, notation.ErrInvalidChromaticUnit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, err := notation.BuiltIn(tt.id, tt.c...)
			require.ErrorIs(t, err, tt.want)
			require.False(t, n.IsDefined())
			requirePanicError(t, tt.want, func() { notation.MustBuiltIn(tt.id, tt.c...) })
		})
	}
	for _, c := range []int{3, 65_536} {
		require.True(t, notation.MustBuiltIn(notation.ASCIIUpDownID, c).Equal(mustBuiltInValue(t, notation.ASCIIUpDownID, c)))
	}
	require.True(t, notation.MustBuiltIn(notation.EnglishUnicodeID).Equal(mustBuiltInValue(t, notation.EnglishUnicodeID)))
}

func mustBuiltInValue(t *testing.T, id string, c ...int) notation.Notation {
	t.Helper()
	n, err := notation.BuiltIn(id, c...)
	require.NoError(t, err)
	return n
}

func TestBuiltInControlNamesAndAliases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		id   string
		name spelling.Spelling
		text string
	}{
		{notation.EnglishUnicodeID, spelling.CSharp(), "C♯"},
		{notation.EnglishUnicodeID, spelling.DFlat(), "D♭"},
		{notation.EnglishUnicodeID, spelling.CDoubleSharp(), "C𝄪"},
		{notation.EnglishUnicodeID, spelling.EDoubleFlat(), "E𝄫"},
		{notation.EnglishASCIIMT21ID, spelling.CSharp(), "C#"},
		{notation.EnglishASCIIMT21ID, spelling.DFlat(), "Db"},
		{notation.EnglishASCIIMT21ID, spelling.CDoubleSharp(), "Cx"},
		{notation.GermanAMID, spelling.BDoubleFlat(), "heses"},
		{notation.GermanAMID, spelling.BFlat(), "b"},
		{notation.GermanAMID, spelling.B(), "h"},
		{notation.GermanAMID, spelling.BSharp(), "his"},
		{notation.RussianAMID, spelling.CDoubleSharp(), "до-дубль-диез"},
		{notation.RomanceItalianID, spelling.CSharp(), "do diesis"},
		{notation.RomanceItalianID, spelling.BFlat(), "si bemolle"},
		{notation.ABCID, spelling.CSharp(), "^C"},
		{notation.ABCID, spelling.BFlat(), "_B"},
		{notation.ABCID, spelling.FDoubleSharp(), "^^F"},
		{notation.ABCID, spelling.DDoubleFlat(), "__D"},
		{notation.LilyPondNederlandsID, nameWithAlteration(letter.C, 3, 2), "cisih"},
		{notation.LilyPondNederlandsID, nameWithAlteration(letter.C, -3, 2), "ceseh"},
		{notation.LilyPondEnglishID, nameWithAlteration(letter.C, 1, 2), "cqs"},
		{notation.LilyPondEnglishID, nameWithAlteration(letter.C, 3, 2), "ctqs"},
		{notation.LilyPondEnglishID, spelling.CDoubleSharp(), "css"},
		{notation.LilyPondItalianoID, nameWithAlteration(letter.C, 1, 2), "dosd"},
		{notation.LilyPondItalianoID, nameWithAlteration(letter.C, 3, 2), "dodsd"},
		{notation.ASCIIFokkerID, nameWithAlteration(letter.C, 3, 2), "C#+"},
		{notation.ASCIIFokkerID, nameWithAlteration(letter.B, -3, 2), "Bb-"},
	}
	for _, tt := range tests {
		t.Run(tt.id+"/"+tt.text, func(t *testing.T) {
			t.Parallel()
			n := notation.MustBuiltIn(tt.id)
			require.Equal(t, tt.text, n.MustFormat(tt.name))
			require.True(t, n.MustParse(tt.text).Equal(tt.name))
		})
	}
	require.True(t, notation.MustBuiltIn(notation.ABCID).MustParse("=E").Equal(spelling.E()))
	require.Equal(t, "E", notation.MustBuiltIn(notation.ABCID).MustFormat(spelling.E()))
	for _, c := range []int{3, 4} {
		n := notation.MustBuiltIn(notation.ASCIIUpDownID, c)
		require.True(t, n.MustParse("^C").Equal(nameWithAlteration(letter.C, 1, c)))
		require.True(t, n.MustParse("vC♯").Equal(nameWithAlteration(letter.C, c-1, c)))
	}
}

func TestBuiltInRejectedAlterations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		id           string
		c            []int
		spellingName spelling.Spelling
	}{
		{"MT21 double flat", notation.EnglishASCIIMT21ID, nil, spelling.CDoubleFlat()},
		{"German A double flat", notation.GermanAMID, nil, spelling.ADoubleFlat()},
		{"German E double flat", notation.GermanAMID, nil, spelling.EDoubleFlat()},
		{"ABC half", notation.ABCID, nil, nameWithAlteration(letter.C, 1, 2)},
		{"Russian half", notation.RussianAMID, nil, nameWithAlteration(letter.C, 1, 2)},
		{"LilyPond third", notation.LilyPondEnglishID, nil, nameWithAlteration(letter.C, 1, 3)},
		{"Stein five halves", notation.UnicodeSteinID, nil, nameWithAlteration(letter.C, 5, 2)},
		{"Gould seven halves", notation.UnicodeGouldID, nil, nameWithAlteration(letter.C, 7, 2)},
		{"Sims seventh", notation.UnicodeSimsID, nil, nameWithAlteration(letter.C, 1, 7)},
		{"Wyschnegradsky seventh", notation.UnicodeWyschnegradskyID, nil, nameWithAlteration(letter.C, 1, 7)},
		{"AEU negative three halves", notation.UnicodeAEUID, nil, nameWithAlteration(letter.C, -3, 2)},
		{"Persian three halves", notation.UnicodePersianID, nil, nameWithAlteration(letter.C, 3, 2)},
		{"Arabic five halves", notation.UnicodeArabicID, nil, nameWithAlteration(letter.C, 5, 2)},
		{"Stockhausen negative three halves", notation.UnicodeStockhausenID, nil, nameWithAlteration(letter.C, -3, 2)},
		{"updown half at three", notation.ASCIIUpDownID, []int{3}, nameWithAlteration(letter.C, 1, 2)},
		{"Fokker five halves", notation.ASCIIFokkerID, nil, nameWithAlteration(letter.C, 5, 2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := notation.MustBuiltIn(tt.id, tt.c...)
			text, err := n.Format(tt.spellingName)
			require.ErrorIs(t, err, notation.ErrUnwritableName)
			require.Empty(t, text)
			text, err = n.NameRules().Format(tt.spellingName)
			require.ErrorIs(t, err, notation.ErrUnwritableName)
			require.Empty(t, text)
			requirePanicError(t, notation.ErrUnwritableName, func() { n.MustFormat(tt.spellingName) })
			requirePanicError(t, notation.ErrUnwritableName, func() { n.NameRules().MustFormat(tt.spellingName) })
		})
	}
}

func TestBuiltInStrictParsing(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	for _, text := range []string{"", "C♯ ", " C♯", "c♯", "C#", "C♯x", "C♯♯", "C♮♮", "unknown"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			parsed, err := n.Parse(text)
			require.ErrorIs(t, err, notation.ErrParse)
			require.False(t, parsed.IsValid())
			requirePanicError(t, notation.ErrParse, func() { n.MustParse(text) })
		})
	}
	require.True(t, n.MustParse("C♮").Equal(spelling.C()))
}

func TestBuiltInMicrotonalCodePoints(t *testing.T) {
	t.Parallel()
	for _, row := range readFixture(t, microtonalSymbolsFixture) {
		family, valueText, code := row[0], row[1], row[2]
		if family == notation.ASCIIUpDownID || family == notation.ASCIIFokkerID {
			continue
		}
		t.Run(family+"/"+valueText, func(t *testing.T) {
			t.Parallel()
			n := notation.MustBuiltIn(family)
			name := spelling.MustNew(letter.C, alterationValue(t, valueText))
			canonical := "C"
			alias := "C" + symbolText(t, code)
			if valueText != "0" {
				canonical = alias
			}
			require.Equal(t, canonical, n.MustFormat(name))
			require.True(t, n.MustParse(alias).Equal(name))
			require.Equal(t, canonical, n.MustFormat(n.MustParse(alias)))
		})
	}
}

func TestBuiltInProductCSVHashes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want string
	}{
		{"data/notation-built-in-rules.csv", "064f6a00942a206ecdac14ebc35eeaab935b7975fe756c4f0ea951df5c774843"},
		{"data/notation-microtonal-symbols.csv", "18c881c88ad3f3c27c086b172c61ad5d5af114acd4f11fc8dc49e1f213804f20"},
		{"data/notation-octave-marks.csv", "96f28b6314b98a7e0461359503c51d20d278b6052af305a6112e0d3141918c95"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			sum := sha256.Sum256(readProductFile(t, tt.path))
			require.Equal(t, tt.want, hex.EncodeToString(sum[:]))
		})
	}
}

func TestBuiltInNoEnharmonicSubstitution(t *testing.T) {
	t.Parallel()
	n := notation.MustBuiltIn(notation.EnglishASCIIMT21ID)
	_, err := n.Format(spelling.CDoubleFlat())
	require.ErrorIs(t, err, notation.ErrUnwritableName)
	for _, candidate := range []spelling.Spelling{spelling.D(), spelling.BSharp(), spelling.C()} {
		text, formatErr := n.Format(candidate)
		require.NoError(t, formatErr)
		require.NotEmpty(t, text)
	}
	require.NotEqual(t, alteration.Flat(), alteration.Natural())
}
