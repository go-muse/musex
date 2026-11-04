package octave

import (
	"fmt"

	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
)

// NewNumbering returns an own numbering in which octave o has number num and
// the numbers in skipped are skipped. The result is equal by
// [number.Numbering.Equal] to number.New with the ASPN number of o, num, and
// skipped: it has the same form as [number.New], with an octave in place of an
// ASPN number. Like any own numbering, it has no label.
//
// NewNumbering does not retain skipped: changing the slice after the call
// does not affect the numbering. The caller must not modify skipped
// concurrently with the call.
//
// On error, NewNumbering returns the zero (undefined) Numbering and an error.
// The conditions are checked in this order:
//
//  1. o is invalid: the error matches [ErrInvalid];
//  2. num is among the skipped numbers, or a skipped number is repeated: the
//     error matches [number.ErrInvalidDefinition].
func NewNumbering(o Octave, num int, skipped ...int) (number.Numbering, error) {
	if !o.valid {
		return number.Numbering{}, ErrInvalid
	}

	n, err := number.New(o.aspn, num, skipped...)
	if err != nil {
		return number.Numbering{}, fmt.Errorf("octave: numbering with number %d for %v: %w", num, o, err)
	}

	return n, nil
}

// MustNewNumbering is like [NewNumbering] but panics if NewNumbering returns
// an error. The panic value is that error.
//
// It is intended for literals and values known to be valid.
func MustNewNumbering(o Octave, num int, skipped ...int) number.Numbering {
	return must(NewNumbering(o, num, skipped...))
}

// NewCatalog returns a catalog of octave names from names, a map from an
// octave to its name. The result is equal by [naming.Catalog.Equal] to the
// catalog that [naming.New] creates from the same pairs with the ASPN
// numbers of the octaves. An empty name means that the octave has no name;
// a nil or empty map gives a new non-nil catalog without names.
//
// Because Octave is comparable and each octave has one representation, equal
// octaves are one map key, so one octave cannot be given twice in one call.
// In a map literal with equal non-constant keys, Go does not specify which
// name is kept. For explicit last-write-wins behavior, assign the entries
// sequentially to a map variable before calling NewCatalog.
//
// NewCatalog does not modify or retain names, including on error: changing
// the map after the call does not affect the catalog. The caller must not
// modify names concurrently with the call.
//
// On error, NewCatalog returns nil and an error. The conditions are checked in
// this order:
//
//  1. a key of names is the invalid value, whatever its name: the error
//     matches [ErrInvalid];
//  2. the same non-empty name is given to two different octaves: the error
//     matches [naming.ErrDuplicateName].
func NewCatalog(names map[Octave]string) (*naming.Catalog, error) {
	pairs := make(map[int]string, len(names))
	for o, name := range names {
		if !o.valid {
			return nil, ErrInvalid
		}

		pairs[o.aspn] = name
	}

	c, err := naming.New(pairs)
	if err != nil {
		return nil, fmt.Errorf("octave: catalog: %w", err)
	}

	return c, nil
}

// MustNewCatalog is like [NewCatalog] but panics if NewCatalog returns an
// error. The panic value is that error.
//
// It is intended for map literals and other data known to be valid.
func MustNewCatalog(names map[Octave]string) *naming.Catalog {
	return must(NewCatalog(names))
}
