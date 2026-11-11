package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
)

func TestParseMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []int{0, 1}, pcseq.MustParse("<0, 1> mod 12").Numbers())
	_, err := pcseq.Parse("<00> mod 12")
	assert.PanicsWithValue(t, err, func() { pcseq.MustParse("<00> mod 12") })
	assert.Equal(t, []int{0, 1}, pcseq.MustParseNotation("⟨0, 1⟩", pitchclass.Decimal, 12).Numbers())
	_, err = pcseq.ParseNotation("⟨01⟩", pitchclass.Decimal, 12)
	assert.PanicsWithValue(t, err, func() { pcseq.MustParseNotation("⟨01⟩", pitchclass.Decimal, 12) })
	assert.Equal(t, []int{0, 1}, pcseq.MustParseRowNotation("⟨0, 1⟩", pitchclass.Decimal, 2).Numbers())
	_, err = pcseq.ParseRowNotation("⟨0, 0⟩", pitchclass.Decimal, 2)
	assert.PanicsWithValue(t, err, func() { pcseq.MustParseRowNotation("⟨0, 0⟩", pitchclass.Decimal, 2) })
}

func TestNotationMustVariants(t *testing.T) {
	t.Parallel()
	s := mustSequence(t, 12, 0, 4, 10, 11)
	assert.Equal(t, "⟨0, 4, T, E⟩", s.MustNotation(pitchclass.TE))
	assert.Equal(t, "⟨04TE⟩", s.MustCompactNotation(pitchclass.TE))
	_, err := s.Notation(pitchclass.Style(99))
	assert.PanicsWithValue(t, err, func() { s.MustNotation(pitchclass.Style(99)) })
	_, err = s.CompactNotation(pitchclass.Decimal)
	assert.PanicsWithValue(t, err, func() { s.MustCompactNotation(pitchclass.Decimal) })
}

func TestSequenceStringMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []int{0, 1}, pcseq.MustParse("<0, 1> mod 12").Numbers())
}

func TestRowStringMustVariants(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	assert.Equal(t, "⟨3, 7, E, 1, 5, 0, 2, T, 6, 4, 8, 9⟩", row.MustNotation(pitchclass.TE))
	assert.Equal(t, "⟨37E1502T6489⟩", row.MustCompactNotation(pitchclass.TE))
	_, err := row.Notation(pitchclass.Style(99))
	assert.PanicsWithValue(t, err, func() { row.MustNotation(pitchclass.Style(99)) })
	_, err = row.CompactNotation(pitchclass.Decimal)
	assert.PanicsWithValue(t, err, func() { row.MustCompactNotation(pitchclass.Decimal) })
}

func TestFormLabelStringMustVariants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "RI7", pcseq.MustNewFormLabel(pcseq.RI, 7, 12).String())
	_, err := pcseq.NewFormLabel(pcseq.P, 12, 12)
	assert.PanicsWithValue(t, err, func() { pcseq.MustNewFormLabel(pcseq.P, 12, 12) })
	assert.Equal(t, "P0", pcseq.MustParseFormLabel("P0", 12).String())
	_, err = pcseq.ParseFormLabel("P00", 12)
	assert.PanicsWithValue(t, err, func() { pcseq.MustParseFormLabel("P00", 12) })
}
