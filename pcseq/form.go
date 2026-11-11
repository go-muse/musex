package pcseq

import (
	"cmp"
	"strconv"
)

// FormLabel identifies one P, I, R, or RI form by its result class.
type FormLabel struct {
	kind    FormKind
	index   int
	modulus int
}

// NewFormLabel returns a label with an index in 0..modulus-1.
func NewFormLabel(kind FormKind, index, modulus int) (FormLabel, error) {
	if err := checkModulus(modulus); err != nil {
		return FormLabel{}, err
	}
	if !kind.IsValid() || index < 0 || index >= modulus {
		return FormLabel{}, invalidFormLabel()
	}
	return FormLabel{kind: kind, index: index, modulus: modulus}, nil
}

// ParseFormLabel parses P0, I3, R6, or RI7 for modulus.
func ParseFormLabel(s string, modulus int) (FormLabel, error) {
	if err := checkModulus(modulus); err != nil {
		return FormLabel{}, err
	}
	split := 0
	for split < len(s) && (s[split] < '0' || s[split] > '9') {
		split++
	}
	if split == 0 || split == len(s) {
		return FormLabel{}, invalidNotation()
	}
	index, ok := parseDecimal(s[split:])
	if !ok {
		return FormLabel{}, invalidNotation()
	}
	var kind FormKind
	switch s[:split] {
	case "P":
		kind = P
	case "I":
		kind = I
	case "R":
		kind = R
	case "RI":
		kind = RI
	default:
		return FormLabel{}, invalidFormLabel()
	}
	return NewFormLabel(kind, index, modulus)
}

// IsValid reports whether l is a valid label.
func (l FormLabel) IsValid() bool {
	return l.kind.IsValid() && l.modulus >= 1 && l.modulus <= MaxModulus &&
		l.index >= 0 && l.index < l.modulus
}

// Kind returns the form family, or the invalid zero kind.
func (l FormLabel) Kind() FormKind { return l.kind }

// Index returns the label index, or 0 for an invalid value.
func (l FormLabel) Index() int { return l.index }

// Modulus returns the label modulus, or 0 for an invalid value.
func (l FormLabel) Modulus() int { return l.modulus }

// Equal reports domain equality of two valid labels.
func (l FormLabel) Equal(other FormLabel) bool {
	return l.IsValid() && other.IsValid() && l == other
}

// Compare orders by modulus, kind, and index.
func (l FormLabel) Compare(other FormLabel) int {
	if !l.IsValid() || !other.IsValid() {
		switch {
		case !l.IsValid() && !other.IsValid():
			return 0
		case !l.IsValid():
			return -1
		default:
			return 1
		}
	}
	if result := cmp.Compare(l.modulus, other.modulus); result != 0 {
		return result
	}
	if result := l.kind.Compare(other.kind); result != 0 {
		return result
	}
	return cmp.Compare(l.index, other.index)
}

// String returns P0, I3, R6, RI7, or FormLabel(invalid).
func (l FormLabel) String() string {
	if !l.IsValid() {
		return "FormLabel(invalid)"
	}
	return l.kind.String() + strconv.Itoa(l.index)
}

func formUnchecked(row Row, label FormLabel) Row {
	first := numberAt(row.elements, 0)
	var sequence Sequence
	switch label.kind {
	case P:
		sequence = row.Sequence().MustTranspose(label.index - first)
	case I:
		sequence = row.Sequence().MustInvert(label.index + first)
	case R:
		sequence = row.Sequence().MustTranspose(label.index - first).MustRetrograde()
	case RI:
		sequence = row.Sequence().MustInvert(label.index + first).MustRetrograde()
	}
	return newRowUnchecked(sequence.modulus, sequence.elements)
}

// Form returns the labelled P, I, R, or RI form.
func (r Row) Form(label FormLabel) (Row, error) {
	if !r.IsValid() {
		return Row{}, invalidValue()
	}
	if !label.IsValid() {
		return Row{}, invalidFormLabel()
	}
	if r.modulus != label.modulus {
		return Row{}, modulusMismatch(r.modulus, label.modulus)
	}
	return formUnchecked(r, label), nil
}

// LabelsOf returns every label under which other is a form of r.
func (r Row) LabelsOf(other Row) ([]FormLabel, error) {
	if !r.IsValid() || !other.IsValid() {
		return nil, invalidValue()
	}
	if r.modulus != other.modulus {
		return nil, modulusMismatch(r.modulus, other.modulus)
	}
	labels := make([]FormLabel, 0, 4)
	for _, kind := range []FormKind{P, I, R, RI} {
		index := numberAt(other.elements, 0)
		if kind == R || kind == RI {
			index = numberAt(other.elements, other.Len()-1)
		}
		label := FormLabel{kind: kind, index: index, modulus: r.modulus}
		if formUnchecked(r, label) == other {
			labels = append(labels, label)
		}
	}
	return labels, nil
}
