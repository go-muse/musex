package system

import (
	"errors"
	"fmt"
)

// Technical limits of the package. They are engineering bounds, not musical
// ones: they keep every computation exact in 64-bit integers and bound the cost
// of rule inclusion.
const (
	// MaxClasses is the largest number of classes N of a system (2^16).
	MaxClasses = 1 << 16

	// MaxChromatic is the largest chromatic unit c of a system (2^16).
	MaxChromatic = 1 << 16

	// MaxDenominator is the largest d of a progression step 1/d (2^16). It
	// equals alteration.MaxDenominator.
	MaxDenominator = 1 << 16

	// MaxProgressions is K, the largest number of progressions in one Rule.
	MaxProgressions = 8
)

// Error categories. Every error returned by the package matches exactly one of
// ErrInvalid, ErrDefinition, ErrModulusMismatch, ErrLimit, ErrNotApplicable,
// ErrInadmissible, ErrNoSpelling, ErrBadSpelling and ErrOutOfRange with
// errors.Is. An error that matches ErrDefinition also matches exactly one
// reason: ErrClassCount, ErrOffsets, ErrDescending, ErrChromatic, ErrNoNatural,
// ErrFractionalShift or ErrRule. The text of an error is not part of the API.
var (
	// ErrInvalid reports an unset input: a nil or zero System or Rule, a zero
	// Progression, letter, name, pitch, class, octave or SpellingRule, an unset
	// ScaleMode or Motion, or a nil custom spelling function.
	ErrInvalid = errors.New("system: unset value")

	// ErrDefinition reports an invalid definition of a system, a progression or
	// a rule. It always comes together with exactly one reason.
	ErrDefinition = errors.New("system: invalid definition")

	// ErrClassCount is reason 1 of ErrDefinition: N is outside 1..MaxClasses.
	ErrClassCount = errors.New("system: class count out of range")

	// ErrOffsets is reason 2 of ErrDefinition: the boundary letter is not at
	// offset 0 or an offset is outside 0..N.
	ErrOffsets = errors.New("system: invalid letter offsets")

	// ErrDescending is reason 3 of ErrDefinition: the offsets go down at least
	// once along the circle of letters from the boundary letter.
	ErrDescending = errors.New("system: letters go down")

	// ErrChromatic is reason 4 of ErrDefinition: c is outside 1..MaxChromatic.
	ErrChromatic = errors.New("system: chromatic unit out of range")

	// ErrNoNatural is reason 5 of ErrDefinition: the alteration value 0 is not
	// admissible.
	ErrNoNatural = errors.New("system: zero alteration not admissible")

	// ErrFractionalShift is reason 6 of ErrDefinition: some admissible value a
	// has a non-integer class shift a·c.
	ErrFractionalShift = errors.New("system: non-integer class shift")

	// ErrRule is the reason of ErrDefinition for progressions and rules: an
	// invalid step, an empty set of points, points that are not alteration
	// values, no progressions or more than MaxProgressions of them.
	ErrRule = errors.New("system: invalid alteration rule")

	// ErrModulusMismatch reports a class whose modulus differs from N.
	ErrModulusMismatch = errors.New("system: class of another system")

	// ErrLimit reports a negative enumeration limit.
	ErrLimit = errors.New("system: negative limit")

	// ErrNotApplicable reports a spelling rule that does not apply to the
	// system: ChromaticScale for a system without the traditional layout.
	ErrNotApplicable = errors.New("system: spelling rule not applicable")

	// ErrInadmissible reports a name that is not admissible in the system: an
	// input name, the name of an input pitch, a tonic or a resulting name.
	ErrInadmissible = errors.New("system: name not admissible")

	// ErrNoSpelling reports a class with no admissible name chosen by a
	// predefined spelling rule.
	ErrNoSpelling = errors.New("system: no admissible spelling")

	// ErrBadSpelling reports a custom spelling function that returned an unset
	// or inadmissible name, or a name of another class.
	ErrBadSpelling = errors.New("system: custom rule returned a bad name")

	// ErrOutOfRange reports a result that cannot be represented: an alteration
	// value outside the range of package alteration, or an Octave, a written
	// octave or a distance in steps outside int.
	ErrOutOfRange = errors.New("system: result out of range")
)

// errorf returns an error that wraps the category err with a message.
func errorf(err error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}

// definitionError returns an error that wraps ErrDefinition and the reason.
func definitionError(reason error, format string, args ...any) error {
	return fmt.Errorf("%w: %w: %s", ErrDefinition, reason, fmt.Sprintf(format, args...))
}

// must returns v or panics with err.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
