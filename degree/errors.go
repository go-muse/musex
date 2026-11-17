package degree

import "errors"

// Errors reported by the package. Test them with errors.Is; error texts are
// not part of the API. An error of the package matches at most one sentinel of
// degree. The error of a user transform is the exception: it is preserved
// inside ErrTransformFailed and may itself match any sentinel.
var (
	// ErrUnset reports an unset degree, note, or transform, including a nil
	// *Degree at a changing method and an unset note returned by a transform.
	ErrUnset = errors.New("degree: unset value")

	// ErrInvalidNumber reports a number below one, an invalid circle size, a
	// number outside a circle where membership is required, or negative cycles.
	ErrInvalidNumber = errors.New("degree: invalid number")

	// ErrNoNote reports a note transformation requested for a degree without a
	// note.
	ErrNoNote = errors.New("degree: note is absent")

	// ErrParse reports text that is not one canonical decimal or Roman number.
	ErrParse = errors.New("degree: cannot parse")

	// ErrRomanOutOfRange reports a number above the Roman numeral range.
	ErrRomanOutOfRange = errors.New("degree: roman numeral out of range")

	// ErrOutOfRange reports a decimal input or denormalized result above
	// MaxNumber.
	ErrOutOfRange = errors.New("degree: result out of range")

	// ErrTransformFailed reports an error returned by a user transform. The
	// original error is preserved in the chain.
	ErrTransformFailed = errors.New("degree: transform failed")
)

func unsetError(_, _ string) error {
	return ErrUnset
}

func invalidNumberError(_, _ string) error {
	return ErrInvalidNumber
}

func noNoteError(_ string) error {
	return ErrNoNote
}

func parseError(_ string) error {
	return ErrParse
}

func romanRangeError(_ string, _ int) error {
	return ErrRomanOutOfRange
}

func outOfRangeError(_, _ string) error {
	return ErrOutOfRange
}

// transformFailure is a value so an operation and its Must wrapper produce
// equal errors for the same stable cause. Unwrap preserves both categories.
type transformFailure struct {
	cause error
}

func (e transformFailure) Error() string {
	return "degree: TransformNote: degree: transform failed: " + e.cause.Error()
}

func (e transformFailure) Unwrap() []error {
	return []error{ErrTransformFailed, e.cause}
}

func transformError(err error) error {
	return transformFailure{cause: err}
}
