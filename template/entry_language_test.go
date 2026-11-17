package template_test

import (
	"testing"
	"unicode"

	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntry_LanguageClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		names           []string
		wantRussian     []string
		wantEnglish     []string
		wantRussianName string
		wantRussianOK   bool
		wantEnglishName string
		wantEnglishOK   bool
	}{
		{
			name:            "S-1_Cyrillic_letter",
			names:           []string{"Ионийский"},
			wantRussian:     []string{"Ионийский"},
			wantEnglish:     []string{},
			wantRussianName: "Ионийский",
			wantRussianOK:   true,
			wantEnglishName: "",
			wantEnglishOK:   false,
		},
		{
			name:            "S-2_Cyrillic_symbol_is_not_a_letter",
			names:           []string{"Mode \u0482"},
			wantRussian:     []string{},
			wantEnglish:     []string{"Mode \u0482"},
			wantRussianName: "",
			wantRussianOK:   false,
			wantEnglishName: "Mode \u0482",
			wantEnglishOK:   true,
		},
		{
			name:            "S-3_Cyrillic_combining_mark_is_not_a_letter",
			names:           []string{"Mode \u0484"},
			wantRussian:     []string{},
			wantEnglish:     []string{"Mode \u0484"},
			wantRussianName: "",
			wantRussianOK:   false,
			wantEnglishName: "Mode \u0484",
			wantEnglishOK:   true,
		},
		{
			name:            "S-4_Mixed_names_preserve_order",
			names:           []string{"Mode \u0482", "Mode Л", "Plain", "Ионийский"},
			wantRussian:     []string{"Mode Л", "Ионийский"},
			wantEnglish:     []string{"Mode \u0482", "Plain"},
			wantRussianName: "Mode Л",
			wantRussianOK:   true,
			wantEnglishName: "Mode \u0482",
			wantEnglishOK:   true,
		},
		{
			name:            "S-5_No_Russian_names",
			names:           []string{"Mode \u0482", "Mode \u0484"},
			wantRussian:     []string{},
			wantEnglish:     []string{"Mode \u0482", "Mode \u0484"},
			wantRussianName: "",
			wantRussianOK:   false,
			wantEnglishName: "Mode \u0482",
			wantEnglishOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			entry := newLanguageEntry(t, tt.names)

			is.Equal(tt.names, entry.Names())

			russianNames := entry.RussianNames()
			is.NotNil(russianNames)
			is.Equal(tt.wantRussian, russianNames)

			englishNames := entry.EnglishNames()
			is.NotNil(englishNames)
			is.Equal(tt.wantEnglish, englishNames)

			russianName, russianOK := entry.RussianName()
			is.Equal(tt.wantRussianName, russianName)
			is.Equal(tt.wantRussianOK, russianOK)

			englishName, englishOK := entry.EnglishName()
			is.Equal(tt.wantEnglishName, englishName)
			is.Equal(tt.wantEnglishOK, englishOK)
		})
	}
}

func TestEntry_LanguageResultSlicesAreIndependent(t *testing.T) {
	t.Parallel()

	names := []string{"Mode \u0482", "Mode Л", "Plain", "Ионийский"}
	tests := []struct {
		name string
		read func(template.Entry) []string
		want []string
	}{
		{
			name: "Names",
			read: template.Entry.Names,
			want: []string{"Mode \u0482", "Mode Л", "Plain", "Ионийский"},
		},
		{
			name: "RussianNames",
			read: template.Entry.RussianNames,
			want: []string{"Mode Л", "Ионийский"},
		},
		{
			name: "EnglishNames",
			read: template.Entry.EnglishNames,
			want: []string{"Mode \u0482", "Plain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)
			entry := newLanguageEntry(t, names)

			result := tt.read(entry)
			must.NotEmpty(result)
			result[0] = "mutated"

			is.Equal(tt.want, tt.read(entry))
			is.Equal(names, entry.Names())
		})
	}
}

func TestEntry_ZeroValueLanguageReads(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	var entry template.Entry

	is.Nil(entry.Names())
	is.Nil(entry.RussianNames())
	is.Nil(entry.EnglishNames())

	russianName, russianOK := entry.RussianName()
	is.Equal("", russianName)
	is.False(russianOK)

	englishName, englishOK := entry.EnglishName()
	is.Equal("", englishName)
	is.False(englishOK)
}

func TestCyrillicNonLetterControls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    rune
	}{
		{name: "U+0482", r: '\u0482'},
		{name: "U+0484", r: '\u0484'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			is.True(unicode.Is(unicode.Cyrillic, tt.r))
			is.False(unicode.IsLetter(tt.r))
		})
	}
}

func newLanguageEntry(t *testing.T, names []string) template.Entry {
	t.Helper()
	return template.MustNewEntry("language-test", names, template.Ionian())
}
