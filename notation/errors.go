package notation

import "errors"

// Errors reported by notation operations. Use errors.Is to test their causes;
// error texts are not part of the public contract.
var (
	// ErrUndefined reports an unset notation, name rules, name, pitch, musical
	// system, or numbering required by the operation.
	ErrUndefined = errors.New("notation: undefined input")

	// ErrUnknownID reports an identifier that is not one of the built-in
	// notation identifiers.
	ErrUnknownID = errors.New("notation: unknown built-in id")

	// ErrMissingChromaticUnit reports the omitted chromatic-unit argument of
	// the parameterized ups-and-downs notation.
	ErrMissingChromaticUnit = errors.New("notation: missing chromatic unit")

	// ErrInvalidChromaticUnit reports a chromatic unit outside 3..65536 or a
	// chromatic-unit argument supplied where it is not allowed.
	ErrInvalidChromaticUnit = errors.New("notation: invalid chromatic unit")

	// ErrInvalidDefinition reports structurally invalid name or notation rules.
	ErrInvalidDefinition = errors.New("notation: invalid definition")

	// ErrAmbiguous reports rules that do not determine one unambiguous reading.
	ErrAmbiguous = errors.New("notation: ambiguous rules")

	// ErrParse reports text that does not match the notation grammar.
	ErrParse = errors.New("notation: text does not match notation")

	// ErrUnwritableName reports a defined name outside the notation's domain.
	ErrUnwritableName = errors.New("notation: name is not writable")

	// ErrNoOctave reports an octave that cannot be represented by the selected
	// style, including a pitch operation on a name-only notation.
	ErrNoOctave = errors.New("notation: octave is not representable")

	// ErrIncompatibleSystem reports a musical system whose chromatic unit does
	// not match the unit of an ups-and-downs notation.
	ErrIncompatibleSystem = errors.New("notation: incompatible musical system")
)
