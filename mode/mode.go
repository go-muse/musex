package mode

import (
	"slices"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"
)

// MaxDegrees is the largest supported number of degrees in a mode.
const MaxDegrees = template.MaxPlaces

// Kind distinguishes a cyclical structure from an open one.
type Kind = template.Kind

const (
	KindUnset = template.KindUnset
	Closed    = template.Closed
	Open      = template.Open
)

// Measure selects how to compare the intervals of two modes.
type Measure = template.Measure

const (
	MeasureUnset = template.MeasureUnset
	FromFirst    = template.FromFirst
	Adjacent     = template.Adjacent
)

// Mode is a musical structure with numbered degrees and optional catalog
// provenance. The zero value is unset. Use Equal to compare structures.
type Mode struct {
	structure template.Template
	degrees   []degree.Degree
	prefixes  []interval.Interval
	cycle     interval.Interval
	origin    template.Entry
	name      string
	changed   bool
}

func buildMode(structure template.Template, origin template.Entry, name string) (Mode, error) {
	prefixes, err := structure.PlaceIntervals()
	if err != nil {
		return Mode{}, templateError(err)
	}
	return modeFromPrefixes(structure, prefixes, origin, name), nil
}

// modeFromPrefixes takes ownership of the checked intervals of the actual
// degrees. Neither these intervals nor the degree slice are ever changed.
func modeFromPrefixes(structure template.Template, prefixes []interval.Interval, origin template.Entry, name string) Mode {
	degrees := make([]degree.Degree, len(prefixes))
	for i := range degrees {
		degrees[i] = degree.MustNew(i + 1)
	}
	cycle, _ := structure.Cycle()
	return Mode{
		structure: structure,
		degrees:   degrees,
		prefixes:  prefixes,
		cycle:     cycle,
		origin:    origin,
		name:      name,
		changed:   origin.IsValid() && !structure.Equal(origin.Template()),
	}
}

// IsValid reports whether m is set.
func (m Mode) IsValid() bool { return m.structure.IsValid() }

// Kind returns the kind of the structure.
func (m Mode) Kind() Kind { return m.structure.Kind() }

// IsClosed reports whether m has a cycle.
func (m Mode) IsClosed() bool { return m.structure.IsClosed() }

// IsOpen reports whether m is set without a cycle.
func (m Mode) IsOpen() bool { return m.structure.IsOpen() }

// Len returns the number of actual degrees.
func (m Mode) Len() int { return len(m.degrees) }

// Cycle returns the interval to the first degree of the next cycle, if any.
func (m Mode) Cycle() (interval.Interval, bool) { return m.cycle, m.IsClosed() }

// IsOctave reports whether the cycle is a perfect octave.
func (m Mode) IsOctave() bool { return m.cycle.Equal(interval.PerfectOctave()) }

// AdjacentIntervals returns an independent slice in degree order.
func (m Mode) AdjacentIntervals() []interval.Interval { return m.structure.AdjacentIntervals() }

// DegreeIntervals returns a copy of the intervals from the first degree.
func (m Mode) DegreeIntervals() []interval.Interval { return slices.Clone(m.prefixes) }

// Template returns the current immutable structure.
func (m Mode) Template() template.Template { return m.structure }

// HasOrigin reports whether a catalog record supplied the original structure.
func (m Mode) HasOrigin() bool { return m.origin.IsValid() }

// Origin returns that catalog record, if any.
func (m Mode) Origin() (template.Entry, bool) { return m.origin, m.HasOrigin() }

// SelectedName returns the original catalog spelling of an explicitly chosen name.
func (m Mode) SelectedName() (string, bool) { return m.name, m.name != "" }

// Changed reports whether the current structure differs from its origin.
func (m Mode) Changed() bool { return m.changed }

// Equal reports structural equality. Unset modes are never equal.
func (m Mode) Equal(other Mode) bool { return m.structure.Equal(other.structure) }

// Compare orders modes by their structure, with unset modes first.
func (m Mode) Compare(other Mode) int { return m.structure.Compare(other.structure) }

// RotationOf reports the smallest degree of source that gives m.
func (m Mode) RotationOf(source Mode) (int, bool) { return m.structure.RotationOf(source.structure) }

func (m Mode) checkDegrees(numbers ...int) error {
	if !m.IsValid() {
		return ErrUnset
	}
	for _, number := range numbers {
		if number < 1 || number > m.Len() {
			return ErrInvalidArgument
		}
	}
	return nil
}

func (m Mode) checkClosed() error {
	if !m.IsValid() {
		return ErrUnset
	}
	if m.IsOpen() {
		return ErrOpenMode
	}
	return nil
}
