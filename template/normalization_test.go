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
		{name: "composed acute", original: "Jué", query: "jue", match: true},
		{name: "decomposed acute", original: "Jué", query: "jue", match: true},
		{name: "edge Unicode spaces", original: "Mode", query: "\u2003mode\u00a0", match: true},
		{name: "repeated Unicode spaces", original: "My mode", query: "my\u2003\u00a0 mode", match: true},
		{name: "tabs and newlines", original: "My mode", query: "my\t\nmode", match: true},
		{name: "non Latin mark retained", original: "кириллица́", query: "кириллица", match: false},
		{name: "Greek mark retained", original: "λάδά", query: "λάδα", match: false},
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
