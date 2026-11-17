package mode

import "fmt"

// String returns the intervals, prefixed by the catalog name when available.
func (m Mode) String() string {
	if !m.IsValid() {
		return "Mode(unset)"
	}
	text := m.structure.String()
	if !m.HasOrigin() {
		return text
	}
	name := m.name
	if name == "" {
		name = m.origin.Names()[0]
	}
	if m.changed {
		name += " (edited)"
	}
	return name + ": " + text
}

// IntervalsString returns the parseable interval syntax for the structure.
func (m Mode) IntervalsString() (string, error) {
	if !m.IsValid() {
		return "", ErrUnset
	}
	return m.structure.String(), nil
}

// RussianIntervals returns a Russian abbreviation for each adjacent interval
// when all of them have an abbreviation.
func (m Mode) RussianIntervals() (string, bool, error) {
	text, ok, err := m.structure.Russian()
	return text, ok, templateError(err)
}

// GoString returns a deterministic diagnostic representation of m.
func (m Mode) GoString() string {
	if !m.IsValid() {
		return "mode.Mode{unset}"
	}
	return fmt.Sprintf("mode.Mode{intervals:%q, origin:%q, selectedName:%q, changed:%t}",
		m.structure.String(), m.origin.Code(), m.name, m.changed)
}
