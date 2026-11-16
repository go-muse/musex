package template

import (
	"fmt"
	"strings"

	"github.com/go-muse/musex/interval"
)

// String returns the short notation: adjacent intervals separated by one ASCII
// space. An open template ends in " |". The zero Template returns
// "Template(unset)".
func (t Template) String() string {
	if !t.IsValid() {
		return "Template(unset)"
	}
	var text strings.Builder
	for i, current := range t.intervals {
		if i > 0 {
			text.WriteByte(' ')
		}
		text.WriteString(current.String())
	}
	if t.open {
		text.WriteString(" |")
	}
	return text.String()
}

// Russian returns adjacent intervals using Russian abbreviations, followed by
// " |" for an open template. If any interval has no Russian abbreviation, it
// returns ("", false, nil). An unset template returns ErrUnset.
func (t Template) Russian() (text string, ok bool, err error) {
	if !t.IsValid() {
		return "", false, opError("Russian", ErrUnset)
	}
	parts := make([]string, 0, len(t.intervals)+1)
	for _, current := range t.intervals {
		abbreviation, exists, err := current.RussianAbbreviation()
		if err != nil {
			return "", false, fmt.Errorf("template: Russian: %w: %w", ErrUnset, err)
		}
		if !exists {
			return "", false, nil
		}
		parts = append(parts, abbreviation)
	}
	if t.open {
		parts = append(parts, "|")
	}
	return strings.Join(parts, " "), true, nil
}

// MustRussian is Russian without the error result. It preserves the boolean
// result and panics with the same error.
func (t Template) MustRussian() (text string, ok bool) {
	text, ok, err := t.Russian()
	if err != nil {
		panic(err)
	}
	return text, ok
}

// Parse reads short interval notation or Russian interval abbreviations
// separated by exactly one ASCII space. A final " |" selects an open template;
// otherwise the result is closed.
func Parse(text string) (Template, error) {
	if text == "" {
		return Template{}, opError("Parse", ErrParse)
	}
	parts := strings.Split(text, " ")
	for _, part := range parts {
		if part == "" {
			return Template{}, opError("Parse", ErrParse)
		}
	}
	open := parts[len(parts)-1] == "|"
	if open {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return Template{}, opError("Parse", ErrParse)
	}
	intervals := make([]interval.Interval, len(parts))
	for i, part := range parts {
		if part == "|" {
			return Template{}, opError("Parse", ErrParse)
		}
		current, err := interval.Parse(part)
		if err != nil {
			shortError := err
			current, err = interval.ParseName(part)
			if err != nil {
				return Template{}, fmt.Errorf("template: Parse: %w: %w: %w", ErrParse, shortError, err)
			}
		}
		intervals[i] = current
	}
	if open {
		result, err := OpenFromIntervals(intervals)
		if err != nil {
			return Template{}, fmt.Errorf("template: Parse: %w", err)
		}
		return result, nil
	}
	result, err := ClosedFromIntervals(intervals)
	if err != nil {
		return Template{}, fmt.Errorf("template: Parse: %w", err)
	}
	return result, nil
}

// MustParse is Parse without the error result. It panics with the same error.
func MustParse(text string) Template {
	return must(Parse(text))
}

// MarshalText returns the short notation of a set template.
func (t Template) MarshalText() ([]byte, error) {
	if !t.IsValid() {
		return nil, opError("MarshalText", ErrUnset)
	}
	return []byte(t.String()), nil
}

// AppendText appends the short notation of a set template to b. On error it
// returns b unchanged and does not modify its existing prefix.
func (t Template) AppendText(b []byte) ([]byte, error) {
	if !t.IsValid() {
		return b, opError("AppendText", ErrUnset)
	}
	return append(b, t.String()...), nil
}

// UnmarshalText replaces t with the parsed template only on success. A nil
// receiver returns ErrUnset.
func (t *Template) UnmarshalText(text []byte) error {
	if t == nil {
		return opError("UnmarshalText", ErrUnset)
	}
	result, err := Parse(string(text))
	if err != nil {
		return fmt.Errorf("template: UnmarshalText: %w", err)
	}
	*t = result
	return nil
}
