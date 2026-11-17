package mode

import (
	"fmt"

	"github.com/go-muse/musex/alteration"
)

// DegreeDifference is the nonzero deviation delta at one degree.
type DegreeDifference struct {
	Degree int
	Delta  alteration.Alteration
}

// Difference gives the degree deviations and an optional cycle deviation.
type Difference struct {
	Degrees  []DegreeDifference
	Cycle    alteration.Alteration
	HasCycle bool
}

// Mismatches lists the degrees whose intervals differ in a selected measure.
type Mismatches struct {
	Degrees []int
}

// Count returns the number of mismatches.
func (r Mismatches) Count() int { return len(r.Degrees) }

// Difference subtracts the deviations of other from those of m. The two
// structures must have the same kind, degree count, and interval sizes.
func (m Mode) Difference(other Mode) (Difference, error) {
	if !m.IsValid() || !other.IsValid() {
		return Difference{}, ErrUnset
	}
	if m.Kind() != other.Kind() || m.Len() != other.Len() || m.cycle.Size() != other.cycle.Size() {
		return Difference{}, ErrIncomparable
	}
	for i, p := range m.prefixes {
		if p.Size() != other.prefixes[i].Size() {
			return Difference{}, ErrIncomparable
		}
	}
	result := Difference{Degrees: []DegreeDifference{}, HasCycle: m.IsClosed()}
	for i, p := range m.prefixes {
		delta, err := p.Deviation().Sub(other.prefixes[i].Deviation())
		if err != nil {
			return Difference{}, fmt.Errorf("%w: degree %d: %w", ErrOutOfRange, i+1, err)
		}
		if delta.Sign() != 0 {
			result.Degrees = append(result.Degrees, DegreeDifference{Degree: i + 1, Delta: delta})
		}
	}
	if result.HasCycle {
		delta, err := m.cycle.Deviation().Sub(other.cycle.Deviation())
		if err != nil {
			return Difference{}, fmt.Errorf("%w: cycle: %w", ErrOutOfRange, err)
		}
		result.Cycle = delta
	}
	return result, nil
}

// MismatchingDegrees gives the differing degree numbers in measure.
func (m Mode) MismatchingDegrees(other Mode, measure Measure) (Mismatches, error) {
	places, err := m.structure.MismatchingPlaces(other.structure, measure)
	if err != nil {
		return Mismatches{}, templateError(err)
	}
	return Mismatches{Degrees: places}, nil
}
