package instrument

import (
	"github.com/go-muse/musex/settings"
)

// Preset is an immutable description of an instrument: a name, a set of
// settings and a reference range, which can be absent. Built-in presets and
// the caller's own presets are the same type.
//
// A Preset is used through a pointer and never changes; the With methods
// return a new preset. A preset is correct: its name is not empty, its set is
// set, and it has no range or a range valid in the system of its set. The nil
// pointer and a pointer to the zero value are the unset preset: IsValid
// reports false, Name returns "", SettingsSet returns nil, Range returns the
// unset Range, HasRange reports false, String returns "Preset(unset)", Equal
// follows its rule, and the other operations fail with ErrUnset.
//
// A preset has no identity and is not a settings.Source: its set is read with
// SettingsSet, and a note is tied to an Instrument, not to a preset. The set
// is held and returned as the same pointer; it must not be overwritten through
// that pointer (a rule of package settings), and neither may a preset.
type Preset struct {
	name string
	set  *settings.Settings
	rng  Range
}

// NewPreset returns a preset with a name, a set of settings and a range.
// The unset Range gives a preset without a range.
//
// Errors, in this order: ErrUnset when set is unset; ErrEmptyName when name is
// empty; for a set range, ErrInadmissible for its lower bound, then for its
// upper bound, then ErrInverted, all in the system of set.
func NewPreset(name string, set *settings.Settings, r Range) (*Preset, error) {
	const op = "new preset"

	if !set.IsValid() {
		return nil, unsetError(op, "settings")
	}

	if name == "" {
		return nil, emptyNameError(op)
	}

	if err := rangeIn(op, set, r); err != nil {
		return nil, err
	}

	return &Preset{name: name, set: set, rng: r}, nil
}

// MustNewPreset is like NewPreset but panics with its error on failure.
func MustNewPreset(name string, set *settings.Settings, r Range) *Preset {
	p, err := NewPreset(name, set, r)
	if err != nil {
		panic(err)
	}

	return p
}

// IsValid reports whether p is set.
func (p *Preset) IsValid() bool {
	return p != nil && p.set.IsValid()
}

// Name returns the name of p; "" for the unset preset.
func (p *Preset) Name() string {
	if !p.IsValid() {
		return ""
	}

	return p.name
}

// SettingsSet returns the set of settings of p: the same pointer that p holds;
// nil for the unset preset. The set must not be overwritten through the
// pointer.
func (p *Preset) SettingsSet() *settings.Settings {
	if !p.IsValid() {
		return nil
	}

	return p.set
}

// Range returns the range of p; the unset Range when p has no range or is
// unset.
func (p *Preset) Range() Range {
	if !p.IsValid() {
		return Range{}
	}

	return p.rng
}

// HasRange reports whether p has a range.
func (p *Preset) HasRange() bool {
	return p.Range().IsValid()
}

// Equal reports whether p and other are equal: both unset, or both set with
// equal names (==), equal sets (settings.Settings.Equal) and equal ranges
// (Range.Equal; no range equals only no range).
func (p *Preset) Equal(other *Preset) bool {
	if p == other {
		return true
	}

	if !p.IsValid() || !other.IsValid() {
		return p.IsValid() == other.IsValid()
	}

	return p.name == other.name && p.set.Equal(other.set) && p.rng.Equal(other.rng)
}

// String returns a debugging form with the name, the set and the range or
// "no range", or "Preset(unset)". It is not a notation, has no parser and does
// not affect equality.
func (p *Preset) String() string {
	if !p.IsValid() {
		return "Preset(unset)"
	}

	return "Preset(" + p.name + ", " + p.set.String() + ", " + rangeString(p.rng) + ")"
}

// WithName returns a new preset with another name.
//
// Errors, in this order: ErrUnset when p is unset; ErrEmptyName.
func (p *Preset) WithName(name string) (*Preset, error) {
	const op = "preset with name"

	if !p.IsValid() {
		return nil, unsetError(op, "preset")
	}

	if name == "" {
		return nil, emptyNameError(op)
	}

	return &Preset{name: name, set: p.set, rng: p.rng}, nil
}

// MustWithName is like WithName but panics with its error on failure.
func (p *Preset) MustWithName(name string) *Preset {
	return must(p.WithName(name))
}

// WithSettingsSet returns a new preset with another set of settings; its
// range is checked in the system of the new set.
//
// Errors, in this order: ErrUnset when p is unset, then when set is unset;
// for a range, ErrInadmissible for the lower bound, then for the upper bound,
// then ErrInverted.
func (p *Preset) WithSettingsSet(set *settings.Settings) (*Preset, error) {
	const op = "preset with settings"

	if !p.IsValid() {
		return nil, unsetError(op, "preset")
	}

	if !set.IsValid() {
		return nil, unsetError(op, "settings")
	}

	if err := rangeIn(op, set, p.rng); err != nil {
		return nil, err
	}

	return &Preset{name: p.name, set: set, rng: p.rng}, nil
}

// MustWithSettingsSet is like WithSettingsSet but panics with its error on
// failure.
func (p *Preset) MustWithSettingsSet(set *settings.Settings) *Preset {
	return must(p.WithSettingsSet(set))
}

// WithRange returns a new preset with another range, checked in the system of
// the set of p. The unset Range gives a preset without a range, as
// WithoutRange does.
//
// Errors, in this order: ErrUnset when p is unset; for a set range,
// ErrInadmissible for the lower bound, then for the upper bound, then
// ErrInverted.
func (p *Preset) WithRange(r Range) (*Preset, error) {
	const op = "preset with range"

	if !p.IsValid() {
		return nil, unsetError(op, "preset")
	}

	if err := rangeIn(op, p.set, r); err != nil {
		return nil, err
	}

	return &Preset{name: p.name, set: p.set, rng: r}, nil
}

// MustWithRange is like WithRange but panics with its error on failure.
func (p *Preset) MustWithRange(r Range) *Preset {
	return must(p.WithRange(r))
}

// WithoutRange returns a new preset without a range. It fails with ErrUnset
// when p is unset.
func (p *Preset) WithoutRange() (*Preset, error) {
	if !p.IsValid() {
		return nil, unsetError("preset without range", "preset")
	}

	return &Preset{name: p.name, set: p.set}, nil
}

// MustWithoutRange is like WithoutRange but panics with its error on failure.
func (p *Preset) MustWithoutRange() *Preset {
	return must(p.WithoutRange())
}

// rangeIn checks a range for a set of settings that is set: the unset Range is
// no range and passes; a set range must be valid in the system of the set.
func rangeIn(op string, set *settings.Settings, r Range) error {
	if !r.IsValid() {
		return nil
	}

	return r.validIn(op, set.System())
}

// rangeString returns the debugging form of a range, or "no range".
func rangeString(r Range) string {
	if !r.IsValid() {
		return "no range"
	}

	return r.String()
}

// must returns v or panics with err.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
