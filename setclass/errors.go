package setclass

import (
	"errors"

	"github.com/go-muse/musex/pcset"
)

const (
	// MaxModulus is the maximum supported pitch-class modulus N for SetClass (65_536).
	MaxModulus = pcset.MaxModulus

	// MaxOrbitModulus is the maximum supported modulus N for orbit-based operations on Tn and TnIn (1_024).
	MaxOrbitModulus = 1_024

	// MaxTTOModulus is the maximum supported modulus N for TTO set classes and TTO operations (256).
	MaxTTOModulus = 256

	// MaxBitmaskModulus is the maximum supported modulus N for uint64 bitmask operations (64).
	MaxBitmaskModulus = 64

	// MaxAllEnumerationModulus is the maximum supported modulus N for full enumeration All (20).
	MaxAllEnumerationModulus = 20

	// MaxEnumerationSets is the maximum number of sets C(N, k) allowed in OfCardinality, SubclassesOfCardinality, SuperclassesOfCardinality, and ByIntervalVector (200_000).
	MaxEnumerationSets = 200_000
)

var (
	// ErrInvalidModulus is returned when a pitch-class modulus is less than 1.
	ErrInvalidModulus = errors.New("setclass: invalid modulus")

	// ErrModulusTooLarge is returned when a modulus exceeds the maximum supported bound for an operation.
	ErrModulusTooLarge = errors.New("setclass: modulus exceeds maximum supported limit")

	// ErrEnumerationLimitExceeded is returned when C(N, k) exceeds MaxEnumerationSets (200_000).
	ErrEnumerationLimitExceeded = errors.New("setclass: enumeration size C(N, k) exceeds maximum supported limit")

	// ErrInvalidCardinality is returned when a cardinality k is outside [0, N] or unsupported by a catalog.
	ErrInvalidCardinality = errors.New("setclass: invalid cardinality")

	// ErrInvalidPitchClass is returned when a pitch-class number is outside [0, N-1] or an invalid pitchclass.Class is passed.
	ErrInvalidPitchClass = errors.New("setclass: invalid pitch class")

	// ErrInvalidBitmask is returned when a bitmask contains bits at or above N.
	ErrInvalidBitmask = errors.New("setclass: bitmask has bits set at or above modulus")

	// ErrNonInvertibleMultiplier is returned when a multiplier m is not coprime to the modulus (gcd(m, N) != 1).
	ErrNonInvertibleMultiplier = errors.New("setclass: multiplier is not invertible modulo N")

	// ErrModulusMismatch is returned when two operands have different moduli.
	ErrModulusMismatch = errors.New("setclass: modulus mismatch")

	// ErrKindMismatch is returned when two SetClasses have different Kinds or an operation does not support the receiver's Kind.
	ErrKindMismatch = errors.New("setclass: equivalence group kind mismatch")

	// ErrNotModulus12 is returned when a mod-12-only operation (catalog bridge, M5, Morris notation) is called with N != 12.
	ErrNotModulus12 = errors.New("setclass: operation requires modulus 12")

	// ErrInvalidValue is returned when a method or function is called with a zero/invalid SetClass, pcset.Set, or catalog value.
	ErrInvalidValue = errors.New("setclass: invalid value")

	// ErrInvalidKind is returned when an unknown Kind value (not Tn, TnIn, or TTO) is supplied.
	ErrInvalidKind = errors.New("setclass: invalid equivalence group kind")

	// ErrInvalidAlgorithm is returned when an unknown Algorithm value (not Rahn or Forte) is supplied.
	ErrInvalidAlgorithm = errors.New("setclass: invalid prime form algorithm")

	// ErrNotation is returned when a notation string cannot be parsed or fails round-trip verification.
	ErrNotation = errors.New("setclass: invalid set class notation")

	// ErrOverflow is returned when a 64-bit integer or bitmask overflows (e.g. N > 64 for uint64 masks or Count overflow).
	ErrOverflow = errors.New("setclass: integer or bitmask overflow")
)
