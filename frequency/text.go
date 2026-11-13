package frequency

import (
	"encoding"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	_ fmt.Formatter            = Criterion{}
	_ encoding.TextMarshaler   = Criterion{}
	_ encoding.TextAppender    = Criterion{}
	_ encoding.TextUnmarshaler = (*Criterion)(nil)
)

// String returns the canonical ASCII text of c. Equal criteria have identical
// texts; unequal criteria have different texts. Parsing this text recovers an
// equal criterion, including for the unset value "unset". The traversal is
// iterative and does not depend on nesting depth.
func (c Criterion) String() string {
	var builder strings.Builder
	type frame struct {
		criterion Criterion
		next      int
		started   bool
	}
	stack := []frame{{criterion: c}}
	for len(stack) > 0 {
		current := &stack[len(stack)-1]
		if !current.criterion.kind.IsComposite() {
			builder.WriteString(simpleCriterionText(current.criterion))
			stack = stack[:len(stack)-1]
			continue
		}
		if !current.started {
			builder.WriteString(current.criterion.kind.String())
			builder.WriteByte('(')
			current.started = true
		}
		if current.next < len(current.criterion.parts) {
			if current.next > 0 {
				builder.WriteString(", ")
			}
			child := current.criterion.parts[current.next]
			current.next++
			stack = append(stack, frame{criterion: child})
			continue
		}
		builder.WriteByte(')')
		stack = stack[:len(stack)-1]
	}
	return builder.String()
}

func simpleCriterionText(c Criterion) string {
	switch c.kind {
	case KindCents:
		return "cents(" + c.tolerance.String() + ")"
	case KindHertz, KindRelative:
		return c.kind.String() + "(" + strconv.FormatFloat(c.number, 'g', -1, 64) + ")"
	default:
		return c.kind.String()
	}
}

// Format implements fmt.Formatter. Verbs v and s write String, q writes its
// quoted form, and width pads with spaces (on the right with the minus flag).
// Other flags and precision are ignored. An unsupported verb writes the
// standard diagnostic %!verb(frequency.Criterion=text) without width.
func (c Criterion) Format(state fmt.State, verb rune) {
	text := c.String()
	if verb != 'v' && verb != 's' && verb != 'q' {
		_, _ = io.WriteString(state, "%!"+string(verb)+"(frequency.Criterion="+text+")")
		return
	}
	if verb == 'q' {
		text = strconv.Quote(text)
	}
	width, hasWidth := state.Width()
	padding := ""
	if hasWidth && width > len(text) {
		padding = strings.Repeat(" ", width-len(text))
	}
	if state.Flag('-') {
		_, _ = io.WriteString(state, text+padding)
	} else {
		_, _ = io.WriteString(state, padding+text)
	}
}

// MarshalText returns a new byte slice containing String. It never returns an
// error, including for the unset criterion.
func (c Criterion) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// AppendText appends String to b and never returns an error. It has the same
// buffer ownership behavior as append(b, c.String()...): b[:len(b)] is never
// changed, and the result may share b's array when spare capacity suffices.
func (c Criterion) AppendText(b []byte) ([]byte, error) {
	return append(b, c.String()...), nil
}

// UnmarshalText parses text and assigns the resulting criterion only on
// success. It does not modify or retain text. A nil or empty text returns
// ErrInvalidText. Calling this method through a nil receiver is a programming
// error and panics; the zero Criterion value itself is safe to use.
func (c *Criterion) UnmarshalText(text []byte) error {
	if c == nil {
		panic("frequency: nil Criterion receiver")
	}
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}
