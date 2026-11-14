package term

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Sentinel errors returned by package operations. Every error returned by the
// package wraps exactly one of these sentinels, so they are pairwise
// distinguishable with [errors.Is]; no sentinel wraps another and no error
// wraps a sentinel of another package. The message text is not part of the
// API, except for the limits on parse error messages described in the package
// documentation.
var (
	// ErrUnset indicates that a receiver or an argument is an unset value of
	// the package: List{} or Term{}.
	ErrUnset = errors.New("term: unset value")

	// ErrInvalidArgument indicates an argument outside the domain of an
	// operation: an unset fraction, a bound of a guide that is not positive,
	// a lower bound greater than the upper one, a guide without bounds, a list
	// without rows, a row without terms, an empty name, a name that is not
	// valid UTF-8, two names of a list equal up to case, a group outside
	// NoGroup..Fast, a place outside the list, or terms of different lists
	// compared by place.
	ErrInvalidArgument = errors.New("term: invalid argument")

	// ErrNoTempo indicates a number of beats that is zero or negative.
	ErrNoTempo = errors.New("term: no tempo")

	// ErrInvalidText indicates a name that is not a name of the list,
	// including the empty string and text that is not valid UTF-8.
	ErrInvalidText = errors.New("term: invalid text")
)

// maxErrorQuoteBytes is the maximum number of input bytes quoted in a parse
// error message.
const maxErrorQuoteBytes = 64

// quoteExcerpt quotes text in full when it is at most maxErrorQuoteBytes bytes
// long. Otherwise it quotes the longest prefix of text that consists of whole
// decoding units (a valid UTF-8 rune or a single byte that does not start a
// valid rune) and fits in maxErrorQuoteBytes bytes, followed by an ellipsis.
// The length of text is not included, so the result depends only on the first
// maxErrorQuoteBytes+3 bytes of text.
func quoteExcerpt(text string) string {
	if len(text) <= maxErrorQuoteBytes {
		return strconv.Quote(text)
	}
	k := 0
	for k < maxErrorQuoteBytes {
		_, width := utf8.DecodeRuneInString(text[k:])
		if k+width > maxErrorQuoteBytes {
			break
		}
		k += width
	}
	return strconv.Quote(text[:k]) + "..."
}

// unsetError returns an error that wraps ErrUnset and names the operation and
// the unset operand.
func unsetError(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s", ErrUnset, op, operand)
}

// invalidError returns an error that wraps ErrInvalidArgument only.
func invalidError(op, detail string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidArgument, op, detail)
}

// noTempoError returns an error that wraps ErrNoTempo only.
func noTempoError(op string) error {
	return fmt.Errorf("%w: %s: number of beats is not positive", ErrNoTempo, op)
}

// textError returns an error that wraps ErrInvalidText and quotes an excerpt
// of the rejected text. The message consists only of the constant sentinel
// text and the excerpt.
func textError(text string) error {
	return fmt.Errorf("%w: %s", ErrInvalidText, quoteExcerpt(text))
}

// must returns v and panics with err if err is not nil.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// must2 returns a and b and panics with err if err is not nil.
func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}
