package midi

import "errors"

var (
	// ErrInvalidNumber reports that a value is not a MIDI note number.
	ErrInvalidNumber = errors.New("midi: invalid MIDI number")

	// ErrUnset reports that a required input was not supplied.
	ErrUnset = errors.New("midi: unset input")

	// ErrNoOctave reports that a note has no positional octave.
	ErrNoOctave = errors.New("midi: note has no octave")

	// ErrNoSource reports that a note has no settings source of its own.
	ErrNoSource = errors.New("midi: note has no settings source")

	// ErrUnsetSettings reports that a settings source returned an unset snapshot.
	ErrUnsetSettings = errors.New("midi: source returned unset settings")

	// ErrNotTraditionalC reports that a system does not use the traditional C layout.
	ErrNotTraditionalC = errors.New("midi: not traditional C layout")

	// ErrInadmissible reports that an input note name is not admissible in its system.
	ErrInadmissible = errors.New("midi: name not admissible")

	// ErrOutOfRange reports that a computed MIDI note number is outside 0 through 127.
	ErrOutOfRange = errors.New("midi: result out of range")

	// ErrInvalidClass reports an unset class or a class whose modulus is not 12.
	ErrInvalidClass = errors.New("midi: invalid class for C layout")

	// ErrNoSpellingRule reports that no spelling rule was supplied.
	ErrNoSpellingRule = errors.New("midi: spelling rule is not set")

	// ErrNotApplicable reports a spelling rule that MIDI conversion does not accept.
	ErrNotApplicable = errors.New("midi: spelling rule not applicable")
)
