package template_test

import (
	"testing"

	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNameNormalization(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		original string
		query    string
		match    bool
	}{
		{name: "ASCII case", original: "DoRiAn", query: "dorian", match: true},
		{name: "Unicode case fold", original: "ΣΊΣΥΦΟΣ", query: "σίσυφος", match: true},
		{name: "Cyrillic case", original: "ДОРИЙСКИЙ", query: "дорийский", match: true},
		{name: "yo", original: "Ёлка", query: "елка", match: true},
		{name: "yo composed record decomposed query", original: "Ёлка", query: "е\u0308лка", match: true},
		{name: "yo composed record lowercase composed query", original: "Ёлка", query: "ёлка", match: true},
		{name: "yo decomposed record composed query", original: "Е\u0308лка", query: "Ёлка", match: true},
		{name: "yo decomposed record lowercase composed query", original: "Е\u0308лка", query: "ёлка", match: true},
		{name: "yo decomposed record without diaeresis query", original: "Е\u0308лка", query: "елка", match: true},
		{name: "yo uppercase composed record decomposed uppercase query", original: "ЁЛКА", query: "Е\u0308ЛКА", match: true},
		{name: "yo uppercase decomposed record composed uppercase query", original: "Е\u0308ЛКА", query: "ЁЛКА", match: true},
		{name: "yo uppercase composed record decomposed lowercase query", original: "ЁЛКА", query: "е\u0308лка", match: true},
		{name: "yo uppercase decomposed record composed lowercase query", original: "Е\u0308ЛКА", query: "ёлка", match: true},
		{name: "yo uppercase composed record lowercase composed query", original: "ЁЛКА", query: "ёлка", match: true},
		{name: "yo uppercase decomposed record lowercase decomposed query", original: "Е\u0308ЛКА", query: "е\u0308лка", match: true},
		{name: "yo lowercase composed record decomposed uppercase query", original: "ёлка", query: "Е\u0308ЛКА", match: true},
		{name: "yo lowercase decomposed record composed uppercase query", original: "е\u0308лка", query: "ЁЛКА", match: true},
		{name: "yo lowercase composed record decomposed lowercase query", original: "ёлка", query: "е\u0308лка", match: true},
		{name: "yo lowercase decomposed record composed lowercase query", original: "е\u0308лка", query: "ёлка", match: true},
		{name: "yo lowercase composed record uppercase composed query", original: "ёлка", query: "ЁЛКА", match: true},
		{name: "yo lowercase decomposed record uppercase decomposed query", original: "е\u0308лка", query: "Е\u0308ЛКА", match: true},
		{name: "flat", original: "Dorian ♭2", query: "dorian b2", match: true},
		{name: "sharp", original: "Locrian ♯2", query: "locrian #2", match: true},
		{name: "double flat", original: "Super-Locrian 𝄫7", query: "super-locrian bb7", match: true},
		{name: "double sharp", original: "Lydian 𝄪2", query: "lydian ##2", match: true},
		{name: "hyphen", original: "Hira-Choshi", query: "hira‐choshi", match: true},
		{name: "non breaking hyphen", original: "Hira‑Choshi", query: "hira-choshi", match: true},
		{name: "figure dash", original: "Hira‒Choshi", query: "hira-choshi", match: true},
		{name: "en dash", original: "Hira–Choshi", query: "hira-choshi", match: true},
		{name: "em dash", original: "Hira—Choshi", query: "hira-choshi", match: true},
		{name: "minus sign", original: "Hira−Choshi", query: "hira-choshi", match: true},
		{name: "composed macron", original: "Hirajōshi", query: "hirajoshi", match: true},
		{name: "decomposed macron", original: "Hirajōshi", query: "hirajoshi", match: true},
		{name: "composed macron record decomposed query", original: "Hirajōshi", query: "Hirajo\u0304shi", match: true},
		{name: "decomposed macron record composed query", original: "Hirajo\u0304shi", query: "Hirajōshi", match: true},
		{name: "composed acute", original: "Jué", query: "jue", match: true},
		{name: "decomposed acute", original: "Jué", query: "jue", match: true},
		{name: "edge Unicode spaces", original: "Mode", query: "\u2003mode\u00a0", match: true},
		{name: "repeated Unicode spaces", original: "My mode", query: "my\u2003\u00a0 mode", match: true},
		{name: "tabs and newlines", original: "My mode", query: "my\t\nmode", match: true},
		{name: "non Latin mark retained", original: "кириллица́", query: "кириллица", match: false},
		{name: "Greek mark retained", original: "λάδά", query: "λάδα", match: false},
		{name: "short i is not i", original: "Йод", query: "Иод", match: false},
		{name: "i is not short i", original: "Иод", query: "Йод", match: false},
		{name: "composed short i record decomposed query", original: "Йод", query: "И\u0306од", match: true},
		{name: "decomposed short i record composed query", original: "И\u0306од", query: "Йод", match: true},
		{name: "decomposed short i is not i", original: "И\u0306од", query: "Иод", match: false},
		{name: "i is not decomposed short i", original: "Иод", query: "И\u0306од", match: false},
		{name: "not a substring", original: "Dorian", query: "dori", match: false},
		{name: "empty query", original: "Dorian", query: "", match: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entry := template.MustNewEntry("entry", []string{tc.original}, template.Ionian())
			catalog := template.MustNewCatalog(entry)
			got := catalog.ByName(tc.query)
			assert.NotNil(t, got)
			if tc.match {
				require.Len(t, got, 1)
				assert.True(t, got[0].Equal(entry))
				assert.Equal(t, tc.original, got[0].Names()[0])
			} else {
				assert.Empty(t, got)
			}
		})
	}
}

func TestCyrillicYoDuplicateNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		names []string
	}{
		{name: "composed then decomposed", names: []string{"Ёлка", "Е\u0308лка"}},
		{name: "decomposed then composed", names: []string{"Е\u0308лка", "Ёлка"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.NewEntry("entry", tc.names, template.Ionian())
			require.ErrorIs(t, err, template.ErrInvalidArgument)
			assert.False(t, got.IsValid())
		})
	}
}

func TestCyrillicShortINamesAreDistinct(t *testing.T) {
	t.Parallel()
	names := []string{"Йод", "Иод"}
	entry, err := template.NewEntry("short-i", names, template.Ionian())
	require.NoError(t, err)
	assert.True(t, entry.IsValid())
	assert.Equal(t, names, entry.Names())
}

func TestCyrillicYoOriginalNames(t *testing.T) {
	t.Parallel()
	names := []string{"Первый", "Е\u0308лка", "Third"}
	entry := template.MustNewEntry("entry", names, template.Ionian())
	assert.Equal(t, []string{"Первый", "Е\u0308лка", "Third"}, entry.Names())
}

func TestCyrillicYoEntryEquality(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		leftNames  []string
		rightNames []string
		want       bool
	}{
		{
			name:       "composed and decomposed originals differ",
			leftNames:  []string{"Первый", "Ёлка", "Third"},
			rightNames: []string{"Первый", "Е\u0308лка", "Third"},
			want:       false,
		},
		{
			name:       "decomposed and composed originals differ",
			leftNames:  []string{"Первый", "Е\u0308лка", "Third"},
			rightNames: []string{"Первый", "Ёлка", "Third"},
			want:       false,
		},
		{
			name:       "identical decomposed originals match",
			leftNames:  []string{"Первый", "Е\u0308лка", "Third"},
			rightNames: []string{"Первый", "Е\u0308лка", "Third"},
			want:       true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			left := template.MustNewEntry("entry", tc.leftNames, template.Ionian())
			right := template.MustNewEntry("entry", tc.rightNames, template.Ionian())
			assert.Equal(t, tc.want, left.Equal(right))
		})
	}
}

func TestBuiltInNameLookupControls(t *testing.T) {
	t.Parallel()
	cases := []struct {
		query string
		codes []string
	}{
		{query: "Dorian", codes: []string{"dorian"}},
		{query: "dorian", codes: []string{"dorian"}},
		{query: "ДОРИЙСКИЙ", codes: []string{"dorian"}},
		{query: "эолийский", codes: []string{"aeolian"}},
		{query: "natural minor", codes: []string{"aeolian"}},
		{query: "Dorian ♭2", codes: []string{"dorian-b2"}},
		{query: "dorian b2", codes: []string{"dorian-b2"}},
		{query: "Locrian #2", codes: []string{"locrian-sharp2"}},
		{query: "полууменьшенный", codes: []string{"locrian-sharp2"}},
		{query: "октатоника тон-полутон", codes: []string{"octatonic-wh"}},
		{query: "лад Мессиана 2", codes: []string{"octatonic-hw"}},
		{query: "лад Мессиана 4", codes: []string{"messiaen-4-papadopoulos", "messiaen-4-degraaf", "messiaen-4-commons"}},
		{query: "fourth mode", codes: []string{"messiaen-4-papadopoulos", "messiaen-4-degraaf", "messiaen-4-commons"}},
		{query: "лад Мессиана 6", codes: []string{"messiaen-6"}},
		{query: "Mode 6", codes: []string{"messiaen-6"}},
		{query: "Messiaen's sixth mode", codes: []string{"seventh-flat5-diminished"}},
		{query: "blues scale", codes: []string{"blues", "blues-b5"}},
		{query: "блюзовая", codes: []string{"blues", "blues-b5"}},
		{query: "Hirajōshi", codes: []string{"hirajoshi-burrows", "hirajoshi-sachs", "hirajoshi-kostka"}},
		{query: "hirajoshi", codes: []string{"hirajoshi-burrows", "hirajoshi-sachs", "hirajoshi-kostka"}},
		{query: "hira-choshi", codes: []string{"hirajoshi-burrows", "hirajoshi-sachs", "hirajoshi-kostka"}},
		{query: "Hungarian minor", codes: []string{"double-harmonic-minor"}},
		{query: "Gypsy minor", codes: []string{"double-harmonic-minor"}},
		{query: "Gypsy", codes: []string{"aeolian-sharp4"}},
		{query: "венгерская гамма", codes: []string{"double-harmonic-major", "double-harmonic-minor"}},
		{query: "Lydian-Mixolydian", codes: []string{"prometheus"}},
		{query: "лидийско-миксолидийский", codes: []string{"acoustic"}},
		{query: "Phrygian-Dorian", codes: []string{"dorian-b2"}},
		{query: "Ultralocrian", codes: []string{"ultralocrian"}},
		{query: "super-locrian bb7", codes: []string{"ultralocrian"}},
		{query: "Blues major", codes: []string{"yo"}},
		{query: "major blues scale", codes: []string{"blues-major"}},
		{query: "M2", codes: []string{"octatonic-hw"}},
		{query: "m2", codes: []string{"octatonic-hw"}},
		{query: "ionian-sharp5", codes: []string{}},
		{query: "мажор", codes: []string{}},
		{query: "", codes: []string{}},
	}
	catalog := template.BuiltIn()
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			t.Parallel()
			got := catalog.ByName(tc.query)
			assert.NotNil(t, got)
			assert.Equal(t, tc.codes, entryCodes(got))
		})
	}
}

func TestBuiltInEntryNameReadings(t *testing.T) {
	t.Parallel()
	ionian := template.IonianEntry()
	assert.Equal(t, []string{"натуральный мажор", "ионийский"}, ionian.RussianNames())
	assert.Equal(t, []string{"major", "Ionian"}, ionian.EnglishNames())
	russian, ok := ionian.RussianName()
	assert.True(t, ok)
	assert.Equal(t, "натуральный мажор", russian)
	english, ok := ionian.EnglishName()
	assert.True(t, ok)
	assert.Equal(t, "major", english)

	_, ok = template.MixolydianB2Entry().RussianName()
	assert.False(t, ok)
	english, ok = template.UltralocrianEntry().EnglishName()
	assert.True(t, ok)
	assert.Equal(t, "Super-Locrian 𝄫7", english)
	english, ok = template.Messiaen3Entry().EnglishName()
	assert.True(t, ok)
	assert.Equal(t, "third mode of limited transposition", english)
	english, ok = template.Messiaen6Entry().EnglishName()
	assert.True(t, ok)
	assert.Equal(t, "sixth mode", english)
}
