package rounding

import (
	"encoding"
	"fmt"
	"unicode/utf8"
)

var (
	_ fmt.Stringer             = Mode(0)
	_ encoding.TextMarshaler   = Mode(0)
	_ encoding.TextAppender    = Mode(0)
	_ encoding.TextUnmarshaler = (*Mode)(nil)
)

// Parse returns the mode with the exact canonical name. Names are
// case-sensitive and surrounding whitespace is not accepted. Any other string
// returns ErrInvalidArgument.
func Parse(name string) (Mode, error) {
	for m := Down; m <= HalfEven; m++ {
		if name == modeNames[m] {
			return m, nil
		}
	}
	return 0, invalidName(name)
}

// MustParse calls Parse and panics with its error if the name is invalid.
func MustParse(name string) Mode {
	m, err := Parse(name)
	if err != nil {
		panic(err)
	}
	return m
}

// Name returns the canonical name of m. It returns ErrUnset for the zero mode
// and ErrUnknownMode for an unknown mode.
func (m Mode) Name() (string, error) {
	if err := validateMode(m); err != nil {
		return "", err
	}
	return modeNames[m], nil
}

// MustName calls Name and panics with its error for an unset or unknown mode.
func (m Mode) MustName() string {
	name, err := m.Name()
	if err != nil {
		panic(err)
	}
	return name
}

// MarshalText returns a newly allocated byte slice containing the canonical
// name. It has the same failures as Name and returns nil on failure.
func (m Mode) MarshalText() ([]byte, error) {
	name, err := m.Name()
	if err != nil {
		return nil, err
	}
	return []byte(name), nil
}

// AppendText appends the canonical name to b without changing b's existing
// bytes. The result shares b's array when its spare capacity suffices. It has
// the same failures as Name and returns nil on failure.
func (m Mode) AppendText(b []byte) ([]byte, error) {
	name, err := m.Name()
	if err != nil {
		return nil, err
	}
	return append(b, name...), nil
}

// UnmarshalText parses an exact canonical name and assigns it to m. On failure
// it returns ErrInvalidArgument and leaves m unchanged. It neither changes nor
// retains text.
func (m *Mode) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

func invalidName(name string) error {
	if len(name) > 64 {
		name = name[:64]
		for len(name) > 0 && !utf8.ValidString(name) {
			name = name[:len(name)-1]
		}
	}
	return fmt.Errorf("%w: mode name %q", ErrInvalidArgument, name)
}
