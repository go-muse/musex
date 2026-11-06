package catalog

import "errors"

var (
	// ErrInvalidNotation is returned when a Forte or Tn-type notation string is syntactically invalid
	// or violates Z-flag or suffix rules.
	ErrInvalidNotation = errors.New("catalog: invalid Forte notation")

	// ErrInvalidCardinality is returned when cardinality is outside the valid range [0, 12].
	ErrInvalidCardinality = errors.New("catalog: invalid cardinality")

	// ErrInvalidOrdinal is returned when a Forte ordinal is outside [1, k_max(cardinality)].
	ErrInvalidOrdinal = errors.New("catalog: invalid Forte ordinal")

	// ErrInvalidSuffix is returned when a Suffix is invalid or incompatible with the chirality of the ForteClass.
	ErrInvalidSuffix = errors.New("catalog: invalid or incompatible orientation suffix")

	// ErrNoCarterCatalog is returned when querying Elliott Carter's catalog for a valid cardinality
	// in {0, 1, 5, 8, 9, 10, 11, 12} where Carter's numbering is not defined.
	ErrNoCarterCatalog = errors.New("catalog: Carter catalog is not defined for cardinality")

	// ErrInvalidCarterNumber is returned when a Carter number is outside [1, k_max(cardinality)]
	// for a supported Carter cardinality in {2, 3, 4, 6, 7}.
	ErrInvalidCarterNumber = errors.New("catalog: invalid Carter number")

	// ErrNotPrimeForm is returned when a bitmask or pitch-class number slice is not a canonical prime form
	// under the requested algorithm.
	ErrNotPrimeForm = errors.New("catalog: not a canonical prime form")

	// ErrInvalidValue is returned when a method returning an error is called on a zero (invalid) catalog value.
	ErrInvalidValue = errors.New("catalog: invalid value")

	// ErrInvalidAlgorithm is returned when an unknown Algorithm value (not Rahn or Forte) is supplied.
	ErrInvalidAlgorithm = errors.New("catalog: invalid prime form algorithm")
)
