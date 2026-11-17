package midi

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// FromPitch returns the MIDI number of written pitch p in the explicit system sys.
// The system must have the traditional C layout, and p's name must be admissible in it.
func FromPitch(p pitch.Pitch, sys *system.System) (Number, error) {
	if !p.IsValid() {
		return Number{}, fmt.Errorf("midi: from pitch: %w: %w", ErrUnset, pitch.ErrInvalid)
	}
	if sys == nil || !sys.IsValid() {
		return Number{}, fmt.Errorf("midi: from pitch: system: %w: %w", ErrUnset, system.ErrInvalid)
	}
	if !hasTraditionalCLayout(sys) {
		return Number{}, fmt.Errorf("midi: from pitch: %w", ErrNotTraditionalC)
	}

	class, err := sys.PitchClass(p)
	if err != nil {
		return Number{}, fmt.Errorf("midi: from pitch: %w: %w", ErrInadmissible, err)
	}

	o, err := p.ASPN()
	if err != nil {
		return Number{}, fmt.Errorf("midi: from pitch: octave: %w: %w", ErrUnset, err)
	}

	return fromPartsASPN(o, class.Number())
}

// MustFromPitch calls FromPitch and panics with the same error if it fails.
func MustFromPitch(p pitch.Pitch, sys *system.System) Number {
	n, err := FromPitch(p, sys)
	if err != nil {
		panic(err)
	}

	return n
}

// Pitch returns a written pitch for n in sys. The spelling is selected only by
// the explicit Sharps, Flats, or CustomSpelling rule; modal and unknown rules
// are rejected before system.PitchAt is called.
func (n Number) Pitch(sys *system.System, rule system.SpellingRule) (pitch.Pitch, error) {
	if !n.IsValid() {
		return pitch.Pitch{}, fmt.Errorf("midi: pitch: %w", ErrInvalidNumber)
	}
	if sys == nil || !sys.IsValid() {
		return pitch.Pitch{}, fmt.Errorf("midi: pitch: system: %w: %w", ErrUnset, system.ErrInvalid)
	}
	if !hasTraditionalCLayout(sys) {
		return pitch.Pitch{}, fmt.Errorf("midi: pitch: %w", ErrNotTraditionalC)
	}
	if err := checkSpellingRule(rule); err != nil {
		return pitch.Pitch{}, err
	}

	o, class, _, err := n.Parts()
	if err != nil {
		return pitch.Pitch{}, fmt.Errorf("midi: pitch: %w", err)
	}

	p, err := sys.PitchAt(class, o, rule)
	if err != nil {
		if errors.Is(err, system.ErrNotApplicable) {
			return pitch.Pitch{}, fmt.Errorf("midi: pitch: %w: %w", ErrNotApplicable, err)
		}

		return pitch.Pitch{}, fmt.Errorf("midi: pitch: %w", err)
	}

	return p, nil
}

// MustPitch calls Pitch and panics with the same error if it fails.
func (n Number) MustPitch(sys *system.System, rule system.SpellingRule) pitch.Pitch {
	p, err := n.Pitch(sys, rule)
	if err != nil {
		panic(err)
	}

	return p
}

func hasTraditionalCLayout(sys *system.System) bool {
	definition := sys.Definition()
	traditional := system.Traditional().Definition()

	return definition.Classes == traditional.Classes &&
		definition.Offsets == traditional.Offsets &&
		definition.Chromatic == traditional.Chromatic &&
		definition.Boundary.Equal(traditional.Boundary)
}

func checkSpellingRule(rule system.SpellingRule) error {
	if !rule.IsValid() {
		return fmt.Errorf("midi: spelling rule: %w: %w", ErrNoSpellingRule, system.ErrInvalid)
	}

	switch rule.String() {
	case "sharps", "flats", "custom":
		return nil
	default:
		return fmt.Errorf("midi: spelling rule %q: %w", rule.String(), ErrNotApplicable)
	}
}
