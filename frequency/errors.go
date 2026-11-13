package frequency

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/go-muse/musex/cents"
)

var (
	// ErrInvalidFrequency reports an input that is not a valid frequency argument.
	ErrInvalidFrequency = errors.New("frequency: invalid frequency")
	// ErrInvalidTolerance reports a negative, non-finite, or otherwise invalid tolerance.
	ErrInvalidTolerance = errors.New("frequency: invalid tolerance")
	// ErrUnset reports an unset criterion used for comparison or as a component.
	ErrUnset = errors.New("frequency: unset criterion")
	// ErrNoComponents reports a composite criterion without components.
	ErrNoComponents = errors.New("frequency: no components")
	// ErrInvalidText reports text that does not follow the criterion grammar.
	ErrInvalidText = errors.New("frequency: invalid criterion text")
	// ErrInvalidParameter reports an invalid numeric parameter or partial index.
	ErrInvalidParameter = errors.New("frequency: invalid numeric parameter")
	// ErrOutOfRange reports a result outside the range required by its operation.
	ErrOutOfRange = errors.New("frequency: result out of range")
)

func invalidFrequency(position int) error {
	return fmt.Errorf("frequency: frequency argument %d: %w: %w", position, ErrInvalidFrequency, cents.ErrInvalidArgument)
}

func invalidTolerance() error {
	return fmt.Errorf("frequency: tolerance: %w", ErrInvalidTolerance)
}

func invalidParameter(detail string) error {
	return fmt.Errorf("frequency: %s: %w", detail, ErrInvalidParameter)
}

func invalidTransposeParameter() error {
	return fmt.Errorf("frequency: division count: %w: %w", ErrInvalidParameter, cents.ErrInvalidArgument)
}

func resultOutOfRange(operation string) error {
	return fmt.Errorf("frequency: %s: %w: %w", operation, ErrOutOfRange, cents.ErrOutOfRange)
}

func invalidCriterionText(text string, position int, cause error) error {
	if cause == nil {
		return fmt.Errorf("frequency: criterion text at byte %d near %s: %w", position, quoteInput(text[position:]), ErrInvalidText)
	}
	return fmt.Errorf("frequency: criterion text at byte %d near %s: %w: %w", position, quoteInput(text[position:]), ErrInvalidText, cause)
}

func semanticError(kind error, text string, position int, cause error) error {
	if cause == nil {
		return fmt.Errorf("frequency: criterion element at byte %d near %s: %w", position, quoteInput(text[position:]), kind)
	}
	return fmt.Errorf("frequency: criterion element at byte %d near %s: %w: %w", position, quoteInput(text[position:]), kind, cause)
}

func quoteInput(text string) string {
	const maxBytes = 64
	truncated := len(text) > maxBytes
	if truncated {
		text = text[:maxBytes]
	}
	for !utf8.ValidString(text) {
		text = text[:len(text)-1]
	}
	if truncated {
		text += "..."
	}
	return fmt.Sprintf("%q", text)
}
