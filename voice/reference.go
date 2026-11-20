package voice

import (
	"fmt"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/system"
)

type referenceKind uint8

const (
	referenceUnset referenceKind = iota
	referenceRange
	referencePreset
	referenceInstrument
	referencePlacement
)

// Reference is a reference range for the tessitura and the location of the
// sounds of a line: a range, a preset, an instrument, or the reference range
// of a role by place.
//
// The constructors do not check their arguments and never panic: the
// operations check them. Reference borrows a preset or an instrument: it
// neither owns nor changes it, and an instrument is read once per operation
// by Instrument.State, so its range and its system come from one moment.
// A copy of a Reference shares the borrowed pointer. The zero value
// Reference{} is unset.
type Reference struct {
	kind       referenceKind
	rng        instrument.Range
	preset     *instrument.Preset
	instrument *instrument.Instrument
	placement  Placement
}

// RangeReference returns the reference range r. The system of an operation
// is then taken from the sounds or given explicitly.
func RangeReference(r instrument.Range) Reference {
	return Reference{kind: referenceRange, rng: r}
}

// PresetReference returns the range of the preset p. Without an explicit
// system an operation works in the system of the set of p.
func PresetReference(p *instrument.Preset) Reference {
	return Reference{kind: referencePreset, preset: p}
}

// InstrumentReference returns the current range of the instrument i. Without
// an explicit system an operation works in the system of its current set.
func InstrumentReference(i *instrument.Instrument) Reference {
	return Reference{kind: referenceInstrument, instrument: i}
}

// PlacementReference returns the reference range of the role p
// (Placement.Range). It is a value, not a preset: without an explicit system
// an operation works in the system of the sounds.
func PlacementReference(p Placement) Reference {
	return Reference{kind: referencePlacement, placement: p}
}

// IsSet reports whether r was made by a constructor, whatever its argument;
// false only for Reference{}.
func (r Reference) IsSet() bool {
	return r.kind != referenceUnset
}

// String returns a debugging form, or "Reference(unset)". It is not a
// notation and has no parser.
func (r Reference) String() string {
	switch r.kind {
	case referenceRange:
		return "Reference(range " + r.rng.String() + ")"
	case referencePreset:
		if r.preset == nil {
			return "Reference(preset nil)"
		}

		return "Reference(preset " + r.preset.String() + ")"
	case referenceInstrument:
		if r.instrument == nil {
			return "Reference(instrument nil)"
		}

		return "Reference(instrument " + r.instrument.String() + ")"
	case referencePlacement:
		return "Reference(placement " + r.placement.String() + ")"
	default:
		return "Reference(unset)"
	}
}

// check reports ErrUnset for an unset reference, a nil or unset preset or
// instrument (with instrument.ErrUnset), and a placement outside the list.
func (r Reference) check() error {
	switch r.kind {
	case referenceUnset:
		return unsetError("reference range", nil)
	case referencePreset:
		if r.preset == nil || !r.preset.IsValid() {
			return unsetError("preset", instrument.ErrUnset)
		}
	case referenceInstrument:
		if r.instrument == nil || !r.instrument.IsValid() {
			return unsetError("instrument", instrument.ErrUnset)
		}
	case referencePlacement:
		if r.placement != PlacementUnset && !r.placement.IsValid() {
			return unsetError("placement", nil)
		}
	case referenceRange:
	}

	return nil
}

// resolve returns the range of a checked reference and, for a preset or an
// instrument, the system of its set; the instrument is read once by State.
// A missing range reports ErrNoRange.
func (r Reference) resolve() (instrument.Range, *system.System, error) {
	var (
		rng    instrument.Range
		sys    *system.System
		exists bool
	)

	switch r.kind {
	case referenceRange:
		rng, exists = r.rng, r.rng.IsValid()
	case referencePreset:
		rng, sys, exists = r.preset.Range(), r.preset.SettingsSet().System(), r.preset.HasRange()
	case referenceInstrument:
		state, err := r.instrument.State()
		if err != nil {
			return instrument.Range{}, nil, fmt.Errorf("voice: state of the instrument: %w", err)
		}

		rng, sys, exists = state.Range(), state.SettingsSet().System(), state.HasRange()
	case referencePlacement:
		rng, exists = r.placement.Range()
	case referenceUnset:
	}

	if !exists {
		return instrument.Range{}, nil, fmt.Errorf("%w: %s", ErrNoRange, r)
	}

	return rng, sys, nil
}
