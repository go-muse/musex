package pcseq

// FormLabel identifies one P, I, R, or RI form by its result class.
type FormLabel struct {
	kind    FormKind
	index   int
	modulus int
}

// NewFormLabel returns a label with an index in 0..modulus-1.
func NewFormLabel(kind FormKind, index, modulus int) (FormLabel, error) { return FormLabel{}, nil }

// ParseFormLabel parses P0, I3, R6, or RI7 for modulus.
func ParseFormLabel(s string, modulus int) (FormLabel, error) { return FormLabel{}, nil }

// IsValid reports whether l is a valid label.
func (l FormLabel) IsValid() bool { return false }

// Kind returns the form family, or the invalid zero kind.
func (l FormLabel) Kind() FormKind { return 0 }

// Index returns the label index, or 0 for an invalid value.
func (l FormLabel) Index() int { return 0 }

// Modulus returns the label modulus, or 0 for an invalid value.
func (l FormLabel) Modulus() int { return 0 }

// Equal reports domain equality of two valid labels.
func (l FormLabel) Equal(other FormLabel) bool { return false }

// Compare orders by modulus, kind, and index.
func (l FormLabel) Compare(other FormLabel) int { return 0 }

// String returns P0, I3, R6, RI7, or FormLabel(invalid).
func (l FormLabel) String() string { return "" }
