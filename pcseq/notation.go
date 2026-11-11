package pcseq

import "github.com/go-muse/musex/pitchclass"

// String returns the canonical sequence notation or Sequence(invalid).
func (s Sequence) String() string { return "" }

// Parse parses canonical sequence notation.
func Parse(s string) (Sequence, error) { return Sequence{}, nil }

// Notation returns delimited literary sequence notation.
func (s Sequence) Notation(style pitchclass.Style) (string, error) { return "", nil }

// CompactNotation returns compact literary sequence notation.
func (s Sequence) CompactNotation(style pitchclass.Style) (string, error) { return "", nil }

// ParseNotation parses literary sequence notation in an explicit style and modulus.
func ParseNotation(s string, style pitchclass.Style, modulus int) (Sequence, error) {
	return Sequence{}, nil
}

// String returns the canonical row notation or Row(invalid).
func (r Row) String() string { return "" }

// ParseRow parses canonical notation and validates the row invariant.
func ParseRow(s string) (Row, error) { return Row{}, nil }

// Notation returns delimited literary row notation.
func (r Row) Notation(style pitchclass.Style) (string, error) { return "", nil }

// CompactNotation returns compact literary row notation.
func (r Row) CompactNotation(style pitchclass.Style) (string, error) { return "", nil }

// ParseRowNotation parses literary notation and validates the row invariant.
func ParseRowNotation(s string, style pitchclass.Style, modulus int) (Row, error) {
	return Row{}, nil
}
