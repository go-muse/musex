package pcseq

import "errors"

// Operation-specific modulus limits.
const (
	MaxModulus                  = 65_536
	MaxMatrixModulus            = 1_024
	MaxMaterializedClassModulus = 1_024
	MaxRotationClassModulus     = 1_024
	MaxTTOClassModulus          = 256
	MaxAnalysisModulus          = 1_024
)

// Sentinel errors returned by pcseq operations. Use errors.Is to classify an
// error; message text is not part of the public contract.
var (
	ErrInvalidModulus        = errors.New("pcseq: invalid modulus")
	ErrModulusTooLarge       = errors.New("pcseq: modulus exceeds operation limit")
	ErrInvalidValue          = errors.New("pcseq: invalid value")
	ErrModulusMismatch       = errors.New("pcseq: modulus mismatch")
	ErrInvalidNumber         = errors.New("pcseq: invalid class number")
	ErrNotRow                = errors.New("pcseq: sequence is not a row")
	ErrInvalidSegment        = errors.New("pcseq: invalid segment")
	ErrInvalidPartition      = errors.New("pcseq: invalid partition")
	ErrInvalidFormLabel      = errors.New("pcseq: invalid form label")
	ErrNotDerived            = errors.New("pcseq: row is not derived")
	ErrNotation              = errors.New("pcseq: invalid notation")
	ErrOverflow              = errors.New("pcseq: result does not fit in int")
	ErrNotModulus12          = errors.New("pcseq: operation requires modulus 12")
	ErrInvalidSegmentContent = errors.New("pcseq: invalid segment content")
)

func checkModulus(modulus int) error {
	if modulus < 1 {
		return ErrInvalidModulus
	}
	return checkModulusLimit(modulus, MaxModulus)
}

func checkModulusLimit(modulus, limit int) error {
	if modulus > limit {
		return ErrModulusTooLarge
	}
	return nil
}

func invalidValue() error {
	return ErrInvalidValue
}

func modulusMismatch(a, b int) error {
	return ErrModulusMismatch
}

func invalidNumber(number, modulus int) error {
	return ErrInvalidNumber
}

func invalidSegment() error {
	return ErrInvalidSegment
}

func invalidPartition() error {
	return ErrInvalidPartition
}

func invalidFormLabel() error {
	return ErrInvalidFormLabel
}

func invalidNotation() error {
	return ErrNotation
}

func invalidSegmentContent() error {
	return ErrInvalidSegmentContent
}
