package pcseq

import (
	"errors"
	"fmt"
)

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
		return fmt.Errorf("%w: %d", ErrInvalidModulus, modulus)
	}
	return checkModulusLimit(modulus, MaxModulus)
}

func checkModulusLimit(modulus, limit int) error {
	if modulus > limit {
		return fmt.Errorf("%w: %d exceeds %d", ErrModulusTooLarge, modulus, limit)
	}
	return nil
}

func invalidValue() error {
	return fmt.Errorf("%w: zero or invalid operand", ErrInvalidValue)
}

func modulusMismatch(a, b int) error {
	return fmt.Errorf("%w: %d and %d", ErrModulusMismatch, a, b)
}

func invalidNumber(number, modulus int) error {
	return fmt.Errorf("%w: %d is outside 0..%d", ErrInvalidNumber, number, modulus-1)
}

func invalidSegment() error {
	return fmt.Errorf("%w: position or length is out of range", ErrInvalidSegment)
}

func invalidPartition() error {
	return fmt.Errorf("%w: segment length must be positive and divide the sequence length", ErrInvalidPartition)
}

func invalidFormLabel() error {
	return fmt.Errorf("%w: unknown kind or index outside the modulus", ErrInvalidFormLabel)
}

func invalidNotation() error {
	return fmt.Errorf("%w: malformed or unsupported text", ErrNotation)
}

func invalidSegmentContent() error {
	return fmt.Errorf("%w: lengths or sets do not partition the aggregate", ErrInvalidSegmentContent)
}
