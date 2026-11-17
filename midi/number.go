package midi

import (
	"fmt"
	"iter"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitchclass"
)

const (
	minNumber  = 0
	maxNumber  = 127
	classCount = 12
)

// Number is a MIDI note number in the range 0 through 127.
//
// The zero value is invalid; it is not MIDI number 0. Number is an immutable,
// comparable value with no shared data. For valid values, == and Equal agree.
// Two invalid values compare equal with == but not with Equal.
type Number struct {
	value uint8
	valid bool
}

// New returns the MIDI note number with integer value.
// Values outside 0 through 127 return the invalid Number and ErrInvalidNumber.
func New(value int) (Number, error) {
	if value < minNumber || value > maxNumber {
		return Number{}, fmt.Errorf("midi: new %d: %w", value, ErrInvalidNumber)
	}

	return Number{value: uint8(value), valid: true}, nil
}

// MustNew calls New and panics with the same error if value is invalid.
func MustNew(value int) Number {
	n, err := New(value)
	if err != nil {
		panic(err)
	}

	return n
}

// IsValid reports whether n is a MIDI note number rather than the invalid zero value.
func (n Number) IsValid() bool {
	return n.valid
}

// Int returns n as an integer. It returns 0 and ErrInvalidNumber for an invalid Number.
func (n Number) Int() (int, error) {
	if !n.IsValid() {
		return 0, fmt.Errorf("midi: int: %w", ErrInvalidNumber)
	}

	return int(n.value), nil
}

// MustInt calls Int and panics with the same error if n is invalid.
func (n Number) MustInt() int {
	value, err := n.Int()
	if err != nil {
		panic(err)
	}

	return value
}

// Min returns MIDI number 0, C in positional octave -1.
func Min() Number {
	return Number{value: minNumber, valid: true}
}

// Max returns MIDI number 127, G in positional octave 9.
func Max() Number {
	return Number{value: maxNumber, valid: true}
}

// Count returns the number of MIDI note numbers, 128.
func Count() int {
	return maxNumber - minNumber + 1
}

// All returns every MIDI note number exactly once in ascending order.
// Each traversal is independent, repeatable, and safe for concurrent use.
func All() iter.Seq[Number] {
	return func(yield func(Number) bool) {
		for value := minNumber; value <= maxNumber; value++ {
			if !yield(Number{value: uint8(value), valid: true}) {
				return
			}
		}
	}
}

// Equal reports whether n and other are the same valid MIDI note number.
// An invalid Number is unequal to every Number, including itself.
func (n Number) Equal(other Number) bool {
	return n.IsValid() && other.IsValid() && n.value == other.value
}

// Compare returns -1, 0, or 1 as n precedes, equals, or follows other in a total order.
// Invalid values precede valid values and compare equal to one another for sorting.
func (n Number) Compare(other Number) int {
	switch {
	case !n.IsValid() && !other.IsValid():
		return 0
	case !n.IsValid():
		return -1
	case !other.IsValid():
		return 1
	case n.value < other.value:
		return -1
	case n.value > other.value:
		return 1
	default:
		return 0
	}
}

// String returns a debugging form: "MIDI <m>" for a valid number and
// "Number(invalid)" for an invalid one. It is not a serialization format.
func (n Number) String() string {
	if !n.IsValid() {
		return "Number(invalid)"
	}

	return fmt.Sprintf("MIDI %d", n.value)
}

// Shift returns n shifted by k positions. A result outside 0 through 127
// returns the invalid Number and ErrOutOfRange; arithmetic never wraps.
func (n Number) Shift(k int) (Number, error) {
	if !n.IsValid() {
		return Number{}, fmt.Errorf("midi: shift %d: %w", k, ErrInvalidNumber)
	}

	m := int(n.value)
	if k < -m || k > maxNumber-m {
		return Number{}, fmt.Errorf("midi: shift %d by %d: %w", m, k, ErrOutOfRange)
	}

	return Number{value: uint8(m + k), valid: true}, nil
}

// MustShift calls Shift and panics with the same error if it fails.
func (n Number) MustShift(k int) Number {
	shifted, err := n.Shift(k)
	if err != nil {
		panic(err)
	}

	return shifted
}

// ShiftOctaves returns n shifted by k positional octaves, or 12*k positions.
// Its class within the octave is preserved. Results never wrap or saturate.
func (n Number) ShiftOctaves(k int) (Number, error) {
	if !n.IsValid() {
		return Number{}, fmt.Errorf("midi: shift octaves %d: %w", k, ErrInvalidNumber)
	}

	m := int(n.value)
	q := m / classCount
	r := m % classCount
	upper := (maxNumber-r)/classCount - q
	if k < -q || k > upper {
		return Number{}, fmt.Errorf("midi: shift %d by %d octaves: %w", m, k, ErrOutOfRange)
	}

	return Number{value: uint8(m + classCount*k), valid: true}, nil
}

// MustShiftOctaves calls ShiftOctaves and panics with the same error if it fails.
func (n Number) MustShiftOctaves(k int) Number {
	shifted, err := n.ShiftOctaves(k)
	if err != nil {
		panic(err)
	}

	return shifted
}

// Next returns the next MIDI position, equivalent to Shift(1).
func (n Number) Next() (Number, error) {
	return n.Shift(1)
}

// MustNext calls Next and panics with the same error if it fails.
func (n Number) MustNext() Number {
	next, err := n.Next()
	if err != nil {
		panic(err)
	}

	return next
}

// Previous returns the previous MIDI position, equivalent to Shift(-1).
func (n Number) Previous() (Number, error) {
	return n.Shift(-1)
}

// MustPrevious calls Previous and panics with the same error if it fails.
func (n Number) MustPrevious() Number {
	previous, err := n.Previous()
	if err != nil {
		panic(err)
	}

	return previous
}

// Up returns the same class one positional octave higher, equivalent to ShiftOctaves(1).
func (n Number) Up() (Number, error) {
	return n.ShiftOctaves(1)
}

// MustUp calls Up and panics with the same error if it fails.
func (n Number) MustUp() Number {
	up, err := n.Up()
	if err != nil {
		panic(err)
	}

	return up
}

// Down returns the same class one positional octave lower, equivalent to ShiftOctaves(-1).
func (n Number) Down() (Number, error) {
	return n.ShiftOctaves(-1)
}

// MustDown calls Down and panics with the same error if it fails.
func (n Number) MustDown() Number {
	down, err := n.Down()
	if err != nil {
		panic(err)
	}

	return down
}

// Sub returns the signed position difference n-other. Either invalid operand
// causes ErrInvalidNumber and result 0. A valid difference lies in -127 through 127.
func (n Number) Sub(other Number) (int, error) {
	if !n.IsValid() || !other.IsValid() {
		return 0, fmt.Errorf("midi: subtract: %w", ErrInvalidNumber)
	}

	return int(n.value) - int(other.value), nil
}

// MustSub calls Sub and panics with the same error if either number is invalid.
func (n Number) MustSub(other Number) int {
	difference, err := n.Sub(other)
	if err != nil {
		panic(err)
	}

	return difference
}

// Parts decomposes n into its positional octave in ASPN numbering, its class
// with modulus 12, and its numeric C-layout position P = m-12.
func (n Number) Parts() (octave.Octave, pitchclass.Class, int, error) {
	if !n.IsValid() {
		return octave.Octave{}, pitchclass.Class{}, 0, fmt.Errorf("midi: parts: %w", ErrInvalidNumber)
	}

	m := int(n.value)
	o := m/classCount - 1
	class := pitchclass.MustNew(m%classCount, classCount)

	return octave.MustFromNumber(o, number.ASPN()), class, m - classCount, nil
}

// MustParts calls Parts and panics with the same error if n is invalid.
func (n Number) MustParts() (octave.Octave, pitchclass.Class, int) {
	o, class, position, err := n.Parts()
	if err != nil {
		panic(err)
	}

	return o, class, position
}

// FromParts assembles a MIDI number from a positional octave and a numeric
// C-layout class. It does not consult a musical system.
func FromParts(o octave.Octave, c pitchclass.Class) (Number, error) {
	if !o.IsValid() {
		return Number{}, fmt.Errorf("midi: from parts: octave: %w: %w", ErrUnset, octave.ErrInvalid)
	}
	if !c.IsValid() {
		return Number{}, fmt.Errorf("midi: from parts: class: %w: %w", ErrInvalidClass, pitchclass.ErrInvalidValue)
	}
	if c.Modulus() != classCount {
		return Number{}, fmt.Errorf("midi: from parts: class modulus %d: %w: %w", c.Modulus(), ErrInvalidClass, pitchclass.ErrModulusMismatch)
	}

	return fromPartsASPN(o.MustASPN(), c.Number())
}

// MustFromParts calls FromParts and panics with the same error if it fails.
func MustFromParts(o octave.Octave, c pitchclass.Class) Number {
	n, err := FromParts(o, c)
	if err != nil {
		panic(err)
	}

	return n
}

func fromPartsASPN(o, class int) (Number, error) {
	if o < -1 || o > 9 {
		return Number{}, fmt.Errorf("midi: octave %d: %w", o, ErrOutOfRange)
	}

	m := classCount*(o+1) + class
	if m < minNumber || m > maxNumber {
		return Number{}, fmt.Errorf("midi: assembled number %d: %w", m, ErrOutOfRange)
	}

	return Number{value: uint8(m), valid: true}, nil
}
