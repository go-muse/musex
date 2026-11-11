package pcseq

import (
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/setclass"
)

// Sequence is an immutable finite ordered list of pitch classes of one
// modulus. Repetition is significant. Its zero value is invalid.
type Sequence struct {
	modulus  int
	elements string
}

// Empty returns an empty sequence of modulus.
func Empty(modulus int) (Sequence, error) { return Sequence{}, nil }

// New returns a sequence from strict class numbers.
func New(modulus int, numbers ...int) (Sequence, error) { return Sequence{}, nil }

// Reduce returns a sequence after reducing arbitrary integers modulo modulus.
func Reduce(modulus int, xs ...int) (Sequence, error) { return Sequence{}, nil }

// FromClasses returns a sequence and infers its modulus from nonempty input.
func FromClasses(classes ...pitchclass.Class) (Sequence, error) { return Sequence{}, nil }

// FromClassesWithModulus returns a sequence of an explicit modulus.
func FromClassesWithModulus(modulus int, classes ...pitchclass.Class) (Sequence, error) {
	return Sequence{}, nil
}

// IsValid reports whether s is a valid sequence, including a valid empty one.
func (s Sequence) IsValid() bool { return false }

// IsEmpty reports whether s is valid and has no elements.
func (s Sequence) IsEmpty() bool { return false }

// IsRow reports whether s contains every class of its modulus exactly once.
func (s Sequence) IsRow() bool { return false }

// Modulus returns the sequence modulus, or 0 for an invalid value.
func (s Sequence) Modulus() int { return 0 }

// Len returns the number of elements.
func (s Sequence) Len() int { return 0 }

// At returns the class at a zero-based position.
func (s Sequence) At(index int) (pitchclass.Class, error) { return pitchclass.Class{}, nil }

// Classes returns a caller-owned slice in sequence order.
func (s Sequence) Classes() []pitchclass.Class { return nil }

// Numbers returns a caller-owned slice of class numbers in sequence order.
func (s Sequence) Numbers() []int { return nil }

// Equal reports domain equality of two valid sequences.
func (s Sequence) Equal(other Sequence) bool { return false }

// Compare orders by modulus, length, and lexicographic element order.
func (s Sequence) Compare(other Sequence) int { return 0 }

// Transpose applies T_n to every element.
func (s Sequence) Transpose(n int) (Sequence, error) { return Sequence{}, nil }

// Invert applies I_n to every element.
func (s Sequence) Invert(n int) (Sequence, error) { return Sequence{}, nil }

// Multiply applies M_m to every element.
func (s Sequence) Multiply(m int) (Sequence, error) { return Sequence{}, nil }

// Apply applies a pitch-class operator to every element.
func (s Sequence) Apply(op pitchclass.Operator) (Sequence, error) { return Sequence{}, nil }

// Retrograde reverses element order.
func (s Sequence) Retrograde() (Sequence, error) { return Sequence{}, nil }

// RetrogradeInversion reverses the inverted sequence.
func (s Sequence) RetrogradeInversion(n int) (Sequence, error) { return Sequence{}, nil }

// Rotate rotates a nonempty sequence left by k modulo its length.
func (s Sequence) Rotate(k int) (Sequence, error) { return Sequence{}, nil }

// Intervals returns ordered intervals between adjacent elements.
func (s Sequence) Intervals() ([]pitchclass.Class, error) { return nil, nil }

// Content returns the unordered set of classes in s.
func (s Sequence) Content() (pcset.Set, error) { return pcset.Set{}, nil }

// Segment returns a contiguous segment.
func (s Sequence) Segment(start, length int) (Sequence, error) { return Sequence{}, nil }

// OverlappingSegments returns every contiguous segment of length.
func (s Sequence) OverlappingSegments(length int) ([]Sequence, error) { return nil, nil }

// DiscreteSegments partitions s into nonoverlapping segments of length.
func (s Sequence) DiscreteSegments(length int) ([]Sequence, error) { return nil, nil }

// ContentClass classifies segment content under a set-class relation.
func (s Sequence) ContentClass(kind setclass.Kind) (setclass.SetClass, error) {
	return setclass.SetClass{}, nil
}

// SegmentInvariance lists transposition and inversion indices relating content.
type SegmentInvariance struct {
	Transpositions []int
	Inversions     []int
}

// InvariancesTo returns all content-preserving transpositions and inversions.
func (s Sequence) InvariancesTo(other Sequence) (SegmentInvariance, error) {
	return SegmentInvariance{}, nil
}
