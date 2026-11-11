package pcseq

import (
	"strings"

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
func Empty(modulus int) (Sequence, error) {
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	return newSequenceUnchecked(modulus, ""), nil
}

// New returns a sequence from strict class numbers.
func New(modulus int, numbers ...int) (Sequence, error) {
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	for _, number := range numbers {
		if number < 0 || number >= modulus {
			return Sequence{}, invalidNumber(number, modulus)
		}
	}
	return newSequenceUnchecked(modulus, encodeNumbers(numbers)), nil
}

// Reduce returns a sequence after reducing arbitrary integers modulo modulus.
func Reduce(modulus int, xs ...int) (Sequence, error) {
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	buf := make([]byte, 2*len(xs))
	for i, x := range xs {
		putNumber(buf, i, pitchclass.MustReduce(x, modulus).Number())
	}
	return newSequenceUnchecked(modulus, string(buf)), nil
}

// FromClasses returns a sequence and infers its modulus from nonempty input.
func FromClasses(classes ...pitchclass.Class) (Sequence, error) {
	if len(classes) == 0 {
		return Sequence{}, invalidValue()
	}
	for _, class := range classes {
		if !class.IsValid() {
			return Sequence{}, invalidValue()
		}
	}
	modulus := classes[0].Modulus()
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	for _, class := range classes[1:] {
		if class.Modulus() != modulus {
			return Sequence{}, modulusMismatch(modulus, class.Modulus())
		}
	}
	return newSequenceUnchecked(modulus, encodeClasses(classes)), nil
}

// FromClassesWithModulus returns a sequence of an explicit modulus.
func FromClassesWithModulus(modulus int, classes ...pitchclass.Class) (Sequence, error) {
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	for _, class := range classes {
		if !class.IsValid() {
			return Sequence{}, invalidValue()
		}
		if class.Modulus() != modulus {
			return Sequence{}, modulusMismatch(modulus, class.Modulus())
		}
	}
	return newSequenceUnchecked(modulus, encodeClasses(classes)), nil
}

// IsValid reports whether s is a valid sequence, including a valid empty one.
func (s Sequence) IsValid() bool { return sequenceIsValid(s) }

// IsEmpty reports whether s is valid and has no elements.
func (s Sequence) IsEmpty() bool { return s.IsValid() && len(s.elements) == 0 }

// IsRow reports whether s contains every class of its modulus exactly once.
func (s Sequence) IsRow() bool {
	if !s.IsValid() || s.Len() != s.modulus {
		return false
	}
	seen := make([]bool, s.modulus)
	for i := range s.Len() {
		number := numberAt(s.elements, i)
		if seen[number] {
			return false
		}
		seen[number] = true
	}
	return true
}

// Modulus returns the sequence modulus, or 0 for an invalid value.
func (s Sequence) Modulus() int { return s.modulus }

// Len returns the number of elements.
func (s Sequence) Len() int { return len(s.elements) / 2 }

// At returns the class at a zero-based position.
func (s Sequence) At(index int) (pitchclass.Class, error) {
	if !s.IsValid() {
		return pitchclass.Class{}, invalidValue()
	}
	if index < 0 || index >= s.Len() {
		return pitchclass.Class{}, invalidSegment()
	}
	return classAt(s.modulus, s.elements, index), nil
}

// Classes returns a caller-owned slice in sequence order.
func (s Sequence) Classes() []pitchclass.Class {
	if !s.IsValid() {
		return nil
	}
	classes := make([]pitchclass.Class, s.Len())
	for i := range s.Len() {
		classes[i] = classAt(s.modulus, s.elements, i)
	}
	return classes
}

// Numbers returns a caller-owned slice of class numbers in sequence order.
func (s Sequence) Numbers() []int {
	if !s.IsValid() {
		return nil
	}
	numbers := make([]int, s.Len())
	for i := range s.Len() {
		numbers[i] = numberAt(s.elements, i)
	}
	return numbers
}

// Equal reports domain equality of two valid sequences.
func (s Sequence) Equal(other Sequence) bool {
	return s.IsValid() && other.IsValid() && s == other
}

// Compare orders by modulus, length, and lexicographic element order.
func (s Sequence) Compare(other Sequence) int {
	if !s.IsValid() || !other.IsValid() {
		switch {
		case !s.IsValid() && !other.IsValid():
			return 0
		case !s.IsValid():
			return -1
		default:
			return 1
		}
	}
	return compareEncoded(s.modulus, s.elements, other.modulus, other.elements)
}

func (s Sequence) mapValues(transform func(pitchclass.Class) pitchclass.Class) Sequence {
	buf := make([]byte, len(s.elements))
	for i := range s.Len() {
		putNumber(buf, i, transform(classAt(s.modulus, s.elements, i)).Number())
	}
	return newSequenceUnchecked(s.modulus, string(buf))
}

// Transpose applies T_n to every element.
func (s Sequence) Transpose(n int) (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	return s.mapValues(func(class pitchclass.Class) pitchclass.Class {
		return class.Transpose(n)
	}), nil
}

// Invert applies I_n to every element.
func (s Sequence) Invert(n int) (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	return s.mapValues(func(class pitchclass.Class) pitchclass.Class {
		return class.Invert(n)
	}), nil
}

// Multiply applies M_m to every element.
func (s Sequence) Multiply(m int) (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	return s.mapValues(func(class pitchclass.Class) pitchclass.Class {
		return class.Multiply(m)
	}), nil
}

// Apply applies a pitch-class operator to every element.
func (s Sequence) Apply(op pitchclass.Operator) (Sequence, error) {
	if !s.IsValid() || !op.IsValid() {
		return Sequence{}, invalidValue()
	}
	if s.modulus != op.Modulus() {
		return Sequence{}, modulusMismatch(s.modulus, op.Modulus())
	}
	return s.mapValues(op.MustApply), nil
}

// Retrograde reverses element order.
func (s Sequence) Retrograde() (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	buf := make([]byte, len(s.elements))
	for i := range s.Len() {
		putNumber(buf, i, numberAt(s.elements, s.Len()-1-i))
	}
	return newSequenceUnchecked(s.modulus, string(buf)), nil
}

// RetrogradeInversion reverses the inverted sequence.
func (s Sequence) RetrogradeInversion(n int) (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	buf := make([]byte, len(s.elements))
	for i := range s.Len() {
		putNumber(buf, i, classAt(s.modulus, s.elements, s.Len()-1-i).Invert(n).Number())
	}
	return newSequenceUnchecked(s.modulus, string(buf)), nil
}

// Rotate rotates a nonempty sequence left by k modulo its length.
func (s Sequence) Rotate(k int) (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	if s.Len() == 0 {
		return Sequence{}, invalidSegment()
	}
	shift := pitchclass.MustReduce(k, s.Len()).Number()
	buf := make([]byte, len(s.elements))
	copy(buf, s.elements[2*shift:])
	copy(buf[len(s.elements)-2*shift:], s.elements[:2*shift])
	return newSequenceUnchecked(s.modulus, string(buf)), nil
}

// Intervals returns ordered intervals between adjacent elements.
func (s Sequence) Intervals() ([]pitchclass.Class, error) {
	if !s.IsValid() {
		return nil, invalidValue()
	}
	intervals := make([]pitchclass.Class, max(0, s.Len()-1))
	for i := range len(intervals) {
		intervals[i] = pitchclass.MustReduce(classAt(s.modulus, s.elements, i+1).Number()-
			classAt(s.modulus, s.elements, i).Number(), s.modulus)
	}
	return intervals, nil
}

// Content returns the unordered set of classes in s.
func (s Sequence) Content() (pcset.Set, error) {
	if !s.IsValid() {
		return pcset.Set{}, invalidValue()
	}
	return pcset.FromClassesWithModulus(s.modulus, s.Classes()...)
}

// Segment returns a contiguous segment.
func (s Sequence) Segment(start, length int) (Sequence, error) {
	if !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	if start < 0 || length < 0 || length > s.Len() || start > s.Len()-length {
		return Sequence{}, invalidSegment()
	}
	elements := strings.Clone(s.elements[2*start : 2*(start+length)])
	return newSequenceUnchecked(s.modulus, elements), nil
}

// OverlappingSegments returns every contiguous segment of length.
func (s Sequence) OverlappingSegments(length int) ([]Sequence, error) {
	if !s.IsValid() {
		return nil, invalidValue()
	}
	if length < 1 || length > s.Len() {
		return nil, invalidSegment()
	}
	segments := make([]Sequence, s.Len()-length+1)
	for i := range len(segments) {
		segments[i] = newSequenceUnchecked(s.modulus,
			strings.Clone(s.elements[2*i:2*(i+length)]))
	}
	return segments, nil
}

// DiscreteSegments partitions s into nonoverlapping segments of length.
func (s Sequence) DiscreteSegments(length int) ([]Sequence, error) {
	if !s.IsValid() {
		return nil, invalidValue()
	}
	if length < 1 || s.Len()%length != 0 {
		return nil, invalidPartition()
	}
	segments := make([]Sequence, s.Len()/length)
	for i := range len(segments) {
		start := i * length
		segments[i] = newSequenceUnchecked(s.modulus,
			strings.Clone(s.elements[2*start:2*(start+length)]))
	}
	return segments, nil
}

// ContentClass classifies segment content under a set-class relation.
func (s Sequence) ContentClass(kind setclass.Kind) (setclass.SetClass, error) {
	if !s.IsValid() {
		return setclass.SetClass{}, invalidValue()
	}
	if !kind.IsValid() {
		return setclass.SetClass{}, invalidValue()
	}
	if kind == setclass.TTO && s.modulus > MaxTTOClassModulus {
		return setclass.SetClass{}, checkModulusLimit(s.modulus, MaxTTOClassModulus)
	}
	content, err := s.Content()
	if err != nil {
		return setclass.SetClass{}, err
	}
	class, err := setclass.New(content, kind)
	if err != nil {
		return setclass.SetClass{}, err
	}
	return class, nil
}

// SegmentInvariance lists transposition and inversion indices relating content.
type SegmentInvariance struct {
	Transpositions []int
	Inversions     []int
}

// InvariancesTo returns all content-preserving transpositions and inversions.
func (s Sequence) InvariancesTo(other Sequence) (SegmentInvariance, error) {
	if !s.IsValid() || !other.IsValid() {
		return SegmentInvariance{}, invalidValue()
	}
	if s.modulus != other.modulus {
		return SegmentInvariance{}, modulusMismatch(s.modulus, other.modulus)
	}
	if s.Len() != other.Len() {
		return SegmentInvariance{}, invalidSegment()
	}
	source, err := s.Content()
	if err != nil {
		return SegmentInvariance{}, err
	}
	target, err := other.Content()
	if err != nil {
		return SegmentInvariance{}, err
	}
	result := SegmentInvariance{
		Transpositions: make([]int, 0),
		Inversions:     make([]int, 0),
	}
	for n := range s.modulus {
		if source.Transpose(n).Equal(target) {
			result.Transpositions = append(result.Transpositions, n)
		}
		if source.Invert(n).Equal(target) {
			result.Inversions = append(result.Inversions, n)
		}
	}
	return result, nil
}
