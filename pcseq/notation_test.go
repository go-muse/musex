package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceStringCanonicalRoundTrip(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		value    pcseq.Sequence
		expected string
	}{
		{name: "repetitions", value: mustSequence(t, 12, 7, 11, 9, 6, 7), expected: "<7, 11, 9, 6, 7> mod 12"},
		{name: "empty", value: mustSequence(t, 12), expected: "<> mod 12"},
		{name: "single zero mod one", value: mustSequence(t, 1, 0), expected: "<0> mod 1"},
		{name: "single zero", value: mustSequence(t, 12, 0), expected: "<0> mod 12"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.value.String())
			parsed, err := pcseq.Parse(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, tt.value, parsed)
			assert.Equal(t, tt.value, pcseq.MustParse(tt.expected))
		})
	}
}

func TestRowStringCanonicalRoundTrip(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		value    pcseq.Row
		expected string
	}{
		{name: "MT21", value: mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9), expected: "<3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9> mod 12"},
		{name: "N1", value: mustRow(t, 1, 0), expected: "<0> mod 1"},
		{name: "N2", value: mustRow(t, 2, 1, 0), expected: "<1, 0> mod 2"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.value.String())
			parsed, err := pcseq.ParseRow(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, tt.value, parsed)
			assert.Equal(t, tt.value, pcseq.MustParseRow(tt.expected))
		})
	}
}

func TestParseCanonicalErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		text   string
		target error
	}{
		{name: "invalid diagnostic", text: "Sequence(invalid)", target: pcseq.ErrNotation},
		{name: "row diagnostic", text: "Row(invalid)", target: pcseq.ErrNotation},
		{name: "wrong brackets", text: "{0, 1} mod 12", target: pcseq.ErrNotation},
		{name: "round brackets", text: "(0, 1) mod 12", target: pcseq.ErrNotation},
		{name: "missing modulus", text: "<0, 1>", target: pcseq.ErrNotation},
		{name: "extra text", text: "<0, 1> mod 12 trailing", target: pcseq.ErrNotation},
		{name: "signed number", text: "<+0> mod 12", target: pcseq.ErrNotation},
		{name: "leading zero", text: "<00> mod 12", target: pcseq.ErrNotation},
		{name: "embedded leading zero", text: "<0, 01> mod 12", target: pcseq.ErrNotation},
		{name: "wrong separator", text: "<0,1> mod 12", target: pcseq.ErrNotation},
		{name: "extra whitespace", text: "<0,  1> mod 12", target: pcseq.ErrNotation},
		{name: "element outside modulus", text: "<12> mod 12", target: pcseq.ErrNotation},
		{name: "zero modulus", text: "<0> mod 0", target: pcseq.ErrInvalidModulus},
		{name: "large modulus", text: "<> mod 65537", target: pcseq.ErrModulusTooLarge},
		{name: "syntax precedes modulus", text: "<01> mod 0", target: pcseq.ErrNotation},
		{name: "integer overflow", text: "<0> mod 999999999999999999999999999", target: pcseq.ErrNotation},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			value, err := pcseq.Parse(tt.text)
			assert.ErrorIs(t, err, tt.target)
			assert.Equal(t, pcseq.Sequence{}, value)
			row, err := pcseq.ParseRow(tt.text)
			if tt.target == pcseq.ErrInvalidModulus || tt.target == pcseq.ErrModulusTooLarge || tt.target == pcseq.ErrNotation {
				assert.ErrorIs(t, err, tt.target)
			}
			assert.Equal(t, pcseq.Row{}, row)
		})
	}
	_, err := pcseq.ParseRow("<0, 0> mod 2")
	assert.ErrorIs(t, err, pcseq.ErrNotRow)
	_, err = pcseq.ParseRow("<0, 1, 2> mod 12")
	assert.ErrorIs(t, err, pcseq.ErrNotRow)
}

func TestNotationStylesAndCompactForms(t *testing.T) {
	t.Parallel()
	s := mustSequence(t, 12, 0, 4, 10, 11)
	for _, tt := range []struct {
		style     pitchclass.Style
		separated string
		compact   string
	}{
		{style: pitchclass.Decimal, separated: "⟨0, 4, 10, 11⟩", compact: ""},
		{style: pitchclass.TE, separated: "⟨0, 4, T, E⟩", compact: "⟨04TE⟩"},
		{style: pitchclass.LowerTE, separated: "⟨0, 4, t, e⟩", compact: "⟨04te⟩"},
		{style: pitchclass.AB, separated: "⟨0, 4, A, B⟩", compact: "⟨04AB⟩"},
	} {
		t.Run(tt.style.String(), func(t *testing.T) {
			t.Parallel()
			separated, err := s.Notation(tt.style)
			require.NoError(t, err)
			assert.Equal(t, tt.separated, separated)
			parsed, err := pcseq.ParseNotation(separated, tt.style, 12)
			require.NoError(t, err)
			assert.Equal(t, s, parsed)
			compact, err := s.CompactNotation(tt.style)
			if tt.style == pitchclass.Decimal {
				assert.ErrorIs(t, err, pcseq.ErrNotation)
				assert.Empty(t, compact)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.compact, compact)
				parsed, err = pcseq.ParseNotation(compact, tt.style, 12)
				require.NoError(t, err)
				assert.Equal(t, s, parsed)
				ascii := "<" + compact[len("⟨"):len(compact)-len("⟩")] + ">"
				parsed, err = pcseq.ParseNotation(ascii, tt.style, 12)
				require.NoError(t, err)
				assert.Equal(t, s, parsed)
			}
		})
	}

	decimal := mustSequence(t, 10, 0, 4, 7)
	compact, err := decimal.CompactNotation(pitchclass.Decimal)
	require.NoError(t, err)
	assert.Equal(t, "⟨047⟩", compact)
	parsed, err := pcseq.ParseNotation(compact, pitchclass.Decimal, 10)
	require.NoError(t, err)
	assert.Equal(t, decimal, parsed)
	twoZeros, err := pcseq.ParseNotation("⟨00⟩", pitchclass.Decimal, 10)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 0}, twoZeros.Numbers())

	empty, err := mustSequence(t, 12).Notation(pitchclass.TE)
	require.NoError(t, err)
	assert.Equal(t, "⟨⟩", empty)
	parsedEmpty, err := pcseq.ParseNotation(empty, pitchclass.TE, 12)
	require.NoError(t, err)
	assert.True(t, parsedEmpty.IsEmpty())
}

func TestRowStringLiteraryRoundTrip(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	separated, err := row.Notation(pitchclass.TE)
	require.NoError(t, err)
	assert.Equal(t, "⟨3, 7, E, 1, 5, 0, 2, T, 6, 4, 8, 9⟩", separated)
	compact, err := row.CompactNotation(pitchclass.TE)
	require.NoError(t, err)
	assert.Equal(t, "⟨37E1502T6489⟩", compact)
	for _, text := range []string{separated, compact, "<37E1502T6489>"} {
		parsed, err := pcseq.ParseRowNotation(text, pitchclass.TE, 12)
		require.NoError(t, err)
		assert.Equal(t, row, parsed)
	}
	_, err = pcseq.ParseRowNotation("⟨0, 0⟩", pitchclass.Decimal, 2)
	assert.ErrorIs(t, err, pcseq.ErrNotRow)
}

func TestParseNotationErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		text    string
		style   pitchclass.Style
		modulus int
		target  error
	}{
		{name: "invalid modulus", text: "⟨0⟩", style: pitchclass.Decimal, modulus: 0, target: pcseq.ErrInvalidModulus},
		{name: "large modulus", text: "⟨0⟩", style: pitchclass.Decimal, modulus: 65_537, target: pcseq.ErrModulusTooLarge},
		{name: "unknown style", text: "⟨0⟩", style: pitchclass.Style(99), modulus: 12, target: pcseq.ErrNotation},
		{name: "letters above twelve", text: "⟨0⟩", style: pitchclass.TE, modulus: 13, target: pcseq.ErrNotation},
		{name: "wrong case", text: "⟨t⟩", style: pitchclass.TE, modulus: 12, target: pcseq.ErrNotation},
		{name: "wrong letters", text: "⟨A⟩", style: pitchclass.TE, modulus: 12, target: pcseq.ErrNotation},
		{name: "unbalanced", text: "⟨0", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "round brackets", text: "(0)", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "square brackets", text: "[0]", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "extra text", text: "⟨0⟩x", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "leading zero", text: "⟨01⟩", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "compact decimal above ten", text: "⟨012⟩", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "wrong comma spacing", text: "⟨0,1⟩", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "empty token", text: "⟨0, , 1⟩", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
		{name: "sign", text: "⟨+1⟩", style: pitchclass.Decimal, modulus: 12, target: pcseq.ErrNotation},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			value, err := pcseq.ParseNotation(tt.text, tt.style, tt.modulus)
			assert.ErrorIs(t, err, tt.target)
			assert.Equal(t, pcseq.Sequence{}, value)
		})
	}
	for _, style := range []pitchclass.Style{pitchclass.TE, pitchclass.LowerTE, pitchclass.AB} {
		_, err := mustSequence(t, 13, 0).Notation(style)
		assert.ErrorIs(t, err, pcseq.ErrNotation)
		_, err = mustSequence(t, 13, 0).CompactNotation(style)
		assert.ErrorIs(t, err, pcseq.ErrNotation)
	}
	_, err := pcseq.Sequence{}.Notation(pitchclass.Decimal)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Row{}.CompactNotation(pitchclass.TE)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestFormLabelStringAndParse(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"P0", "I3", "R6", "RI7"} {
		label, err := pcseq.ParseFormLabel(text, 12)
		require.NoError(t, err)
		assert.Equal(t, text, label.String())
		assert.Equal(t, label, pcseq.MustParseFormLabel(text, 12))
	}
	for _, tt := range []struct {
		text   string
		target error
	}{
		{text: "P00", target: pcseq.ErrNotation},
		{text: "P01", target: pcseq.ErrNotation},
		{text: "p0", target: pcseq.ErrNotation},
		{text: "P 0", target: pcseq.ErrNotation},
		{text: "P0x", target: pcseq.ErrNotation},
		{text: "IR0", target: pcseq.ErrInvalidFormLabel},
		{text: "P12", target: pcseq.ErrInvalidFormLabel},
		{text: "P-1", target: pcseq.ErrNotation},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			label, err := pcseq.ParseFormLabel(tt.text, 12)
			assert.ErrorIs(t, err, tt.target)
			assert.Equal(t, pcseq.FormLabel{}, label)
		})
	}
	_, err := pcseq.ParseFormLabel("P0", 0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.ParseFormLabel("P0", 65_537)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
}
