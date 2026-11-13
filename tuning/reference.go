package tuning

import (
	"strconv"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
)

// Reference is an immutable reference note and its frequency in hertz.
// The zero value is unset. A Reference does not belong to a musical system;
// New checks that its note is admissible in the tuning's system.
//
// Reference is comparable and can be used as a map key. For every value,
// including the zero value, == and Equal agree. Copies are independent values
// and are safe for concurrent use.
type Reference struct {
	note pitch.Pitch
	freq float64
}

// NewReference returns a reference with the given written pitch and frequency.
// It returns ErrUnset for an unset pitch and ErrInvalidArgument unless freq is
// a normal, finite, positive float64 as defined by cents.IsNormalFrequency.
func NewReference(note pitch.Pitch, freq float64) (Reference, error) {
	if !note.IsValid() {
		return Reference{}, ErrUnset
	}
	if !cents.IsNormalFrequency(freq) {
		return Reference{}, invalidArgument()
	}
	return Reference{note: note, freq: freq}, nil
}

// MustNewReference calls NewReference and panics with its error on failure.
func MustNewReference(note pitch.Pitch, freq float64) Reference {
	r, err := NewReference(note, freq)
	if err != nil {
		panic(err)
	}
	return r
}

// IsValid reports whether r has a written pitch and a normal positive frequency.
func (r Reference) IsValid() bool {
	return r.note.IsValid() && cents.IsNormalFrequency(r.freq)
}

// Pitch returns the reference note, or the unset pitch for an unset reference.
func (r Reference) Pitch() pitch.Pitch {
	return r.note
}

// Frequency returns the reference frequency in hertz, or zero for an unset reference.
func (r Reference) Frequency() float64 {
	return r.freq
}

// WithFrequency returns a reference with the same note and a new frequency.
// It returns ErrUnset for an unset reference and ErrInvalidArgument for an
// invalid frequency. It does not reject a frequency ratio that overflows.
func (r Reference) WithFrequency(freq float64) (Reference, error) {
	if !r.IsValid() {
		return Reference{}, ErrUnset
	}
	if !cents.IsNormalFrequency(freq) {
		return Reference{}, invalidArgument()
	}
	return Reference{note: r.note, freq: freq}, nil
}

// MustWithFrequency calls WithFrequency and panics with its error on failure.
func (r Reference) MustWithFrequency(freq float64) Reference {
	result, err := r.WithFrequency(freq)
	if err != nil {
		panic(err)
	}
	return result
}

// FrequencyRatio returns freq divided by the reference frequency. It returns
// ErrUnset, ErrInvalidArgument, or ErrOutOfRange together with
// cents.ErrOutOfRange when the quotient is not a normal finite positive float64.
func (r Reference) FrequencyRatio(freq float64) (float64, error) {
	if !r.IsValid() {
		return 0, ErrUnset
	}
	if !cents.IsNormalFrequency(freq) {
		return 0, invalidArgument()
	}
	ratio, err := cents.NormalFrequency(freq / r.freq)
	if err != nil {
		return 0, centsError(err)
	}
	return ratio, nil
}

// MustFrequencyRatio calls FrequencyRatio and panics with its error on failure.
func (r Reference) MustFrequencyRatio(freq float64) float64 {
	ratio, err := r.FrequencyRatio(freq)
	if err != nil {
		panic(err)
	}
	return ratio
}

// CentsTo returns the signed distance in cents from r's frequency to other's
// frequency on the 1e-6-cent grid. Both references must have equal written pitches.
func (r Reference) CentsTo(other Reference) (cents.Cents, error) {
	if !r.IsValid() || !other.IsValid() {
		return cents.Cents{}, ErrUnset
	}
	if !r.note.Equal(other.note) {
		return cents.Cents{}, incomparable()
	}
	distance, err := cents.Between(r.freq, other.freq)
	if err != nil {
		return cents.Cents{}, centsError(err)
	}
	return distance, nil
}

// MustCentsTo calls CentsTo and panics with its error on failure.
func (r Reference) MustCentsTo(other Reference) cents.Cents {
	distance, err := r.CentsTo(other)
	if err != nil {
		panic(err)
	}
	return distance
}

// Equal reports whether r and other have the same note and exactly the same
// frequency. Two unset references are equal; an unset and a set reference are not.
func (r Reference) Equal(other Reference) bool {
	return r == other
}

// String returns a debug representation such as "A ASPN 4 = 440 Hz".
// It returns "Reference(unset)" for an unset reference.
func (r Reference) String() string {
	if !r.IsValid() {
		return "Reference(unset)"
	}
	return r.note.String() + " = " + strconv.FormatFloat(r.freq, 'g', -1, 64) + " Hz"
}

// A4At440 returns A4 at 440 Hz, the current international reference standard.
func A4At440() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.A(), 4), 440)
}

// A4At435 returns A4 at 435 Hz, the former international reference standard
// in Vienna, preceding the 440 Hz standard.
func A4At435() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.A(), 4), 435)
}

// A4At442 returns A4 at 442 Hz, the lower end of the 442-444 Hz range that
// reflects a modern tendency toward a higher orchestra pitch; it is not a standard.
func A4At442() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.A(), 4), 442)
}

// A4At444 returns A4 at 444 Hz, the upper end of the 442-444 Hz range that
// reflects a modern tendency toward a higher orchestra pitch; it is not a standard.
func A4At444() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.A(), 4), 444)
}

// A4At453 returns A4 at 453 Hz, a historical tuning-fork frequency reached at
// times in the Paris and Vienna opera houses; it is not a standard.
func A4At453() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.A(), 4), 453)
}

// A4At466 returns A4 at 466 Hz, a historical tuning-fork frequency reached at
// times in the Paris and Vienna opera houses; it is not a standard.
func A4At466() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.A(), 4), 466)
}

// C5At523 returns C5 at 523 Hz, an ordinary tuning fork in the second octave,
// used less often than A4 at 440 Hz. It is not a standard, and its source does
// not establish a temperament. It is not the same pitch level as equal
// temperament with A4 at 440 Hz, whose C5 is approximately 523.2511 Hz.
func C5At523() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.C(), 5), 523)
}

// C0At16 returns C0 at 16 Hz, the scientific pitch standard.
func C0At16() Reference {
	return MustNewReference(pitch.MustFromASPN(spelling.C(), 0), 16)
}

// References returns all eight ready-made references in their documented order.
// The returned slice belongs to the caller and does not affect later calls.
func References() []Reference {
	return []Reference{
		A4At440(),
		A4At435(),
		A4At442(),
		A4At444(),
		A4At453(),
		A4At466(),
		C5At523(),
		C0At16(),
	}
}
