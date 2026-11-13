package naming

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"
	"unicode/utf8"
)

// excerptLimit is the maximum number of bytes of a name quoted in an error
// message.
const excerptLimit = 32

// ErrUnknownName reports that a catalog has no octave with the given name.
//
// [Catalog.ASPN] returns errors that match ErrUnknownName, and
// [Catalog.MustASPN] panics with such an error. Check for it with
// [errors.Is]; the text of the error is not part of the API.
var ErrUnknownName = errors.New("naming: unknown octave name")

// ErrDuplicateName reports that the same non-empty name is given to two
// different octaves while creating a catalog.
//
// [New] returns errors that match ErrDuplicateName, and [MustNew] panics
// with such an error. Check for it with [errors.Is]; the text of the error
// is not part of the API.
var ErrDuplicateName = errors.New("naming: duplicate octave name")

// Catalog is a catalog of octave names: an immutable set of pairs of an ASPN
// octave number and a non-empty name. Each octave has at most one name, and
// different octaves have different names. Octaves without a name in the
// catalog have the empty name "".
//
// Create a catalog with [New] or [MustNew], or take a built-in one with
// [Russian], [English] or [German]. No function or method changes a catalog
// after it is created.
//
// The zero value, such as var c Catalog or &Catalog{}, and the nil pointer
// are valid catalogs without names. They behave like MustNew(nil): [Catalog.Name]
// returns "" for every number, [Catalog.ASPN] reports [ErrUnknownName] for
// every name, and [Catalog.Equal] reports them equal to any catalog without
// names. No method panics on them, except [Catalog.MustASPN] as documented.
//
// Catalog values are not comparable: == on two Catalog values does not
// compile. For *Catalog pointers, == reports whether they point to the same
// object, not whether the catalogs are equal; use [Catalog.Equal] instead.
//
// A Catalog value may be copied, as in cp := *c. The copy is equal to the
// original and behaves the same way; both share immutable internal data, so
// copying is safe, including across goroutines. Assigning to *c changes only
// the object c points to: built-in catalogs returned by other calls are not
// affected.
//
// The package does not provide serialization for catalogs. To store a custom
// catalog, keep the map of pairs it was created from.
type Catalog struct {
	// names maps an ASPN octave number to its non-empty name.
	names map[int]string
	// numbers maps each name in names back to its ASPN octave number.
	numbers map[string]int
	// ordered holds the pairs of names in ascending order of the number. It
	// is built once by New and only read afterward, so a sequence returned by
	// All may keep it without copying.
	ordered []namedOctave
}

// namedOctave is a pair of an ASPN octave number and its non-empty name.
type namedOctave struct {
	number int
	name   string
}

// New returns a catalog with the octave names given by names, a map from an
// ASPN octave number to the name of that octave.
//
// Any int is a valid octave number. A name is stored exactly as given, byte
// for byte: it is not trimmed, normalized or changed in case, and it may be
// any non-empty string. An empty name means that the octave has no name: the
// pair is not included in the catalog and is not checked for duplicates. A
// nil or empty map, or a map with only empty names, gives a catalog without
// names.
//
// New sizes both internal indexes by the number of non-empty names, so
// pairs with an empty name do not increase their capacity.
//
// If the same non-empty name is given to two different octave numbers, New
// returns nil and an error that matches [ErrDuplicateName]; no catalog is
// created. Names are compared exactly, so "a" and "A", or " a" and "a", are
// different names.
//
// New does not modify or retain names: the catalog keeps its own copy of the
// pairs, and later changes to the map do not affect it. If the error is nil,
// New returns a new non-nil *Catalog that the caller owns, even for a
// catalog without names. The caller must not modify names concurrently with
// the call.
func New(names map[int]string) (*Catalog, error) {
	// Size both indexes and the ordered list by the number of non-empty
	// names, not by len(names): a map never shrinks, and pairs with an empty
	// name are left out.
	count := 0

	for _, name := range names {
		if name != "" {
			count++
		}
	}

	byNumber := make(map[int]string, count)
	byName := make(map[string]int, count)
	ordered := make([]namedOctave, 0, count)

	var duplicates []string

	for number, name := range names {
		if name == "" {
			continue
		}

		if _, ok := byName[name]; ok {
			duplicates = append(duplicates, name)

			continue
		}

		byNumber[number] = name
		byName[name] = number
		ordered = append(ordered, namedOctave{number: number, name: name})
	}

	if len(duplicates) > 0 {
		return nil, duplicateError(names, slices.Min(duplicates))
	}

	// Numbers are unique map keys, so the order is strict. cmp.Compare does
	// not overflow on math.MinInt and math.MaxInt, unlike a subtraction.
	slices.SortFunc(ordered, func(a, b namedOctave) int {
		return cmp.Compare(a.number, b.number)
	})

	return &Catalog{names: byNumber, numbers: byName, ordered: ordered}, nil
}

// duplicateError returns an error that wraps ErrDuplicateName for name, which
// is given to at least two numbers in names. The message names the two
// smallest of these numbers, so it does not depend on map iteration order.
// The name is chosen and matched in full; only the message quotes an excerpt.
func duplicateError(names map[int]string, name string) error {
	var numbers []int

	for number, n := range names {
		if n == name {
			numbers = append(numbers, number)
		}
	}

	slices.Sort(numbers)

	return fmt.Errorf("%w: %s for octaves %d and %d", ErrDuplicateName, excerpt(name), numbers[0], numbers[1])
}

// excerpt returns name quoted for an error message, with a length bounded
// whatever the length of name. A name of at most excerptLimit bytes is
// quoted in full. A longer name is cut at the end of the last rune that ends
// within the first excerptLimit bytes, where an invalid byte counts as a rune
// of length 1, and the quoted prefix is followed by an ellipsis and the length
// of the full name in bytes.
func excerpt(name string) string {
	if len(name) <= excerptLimit {
		return strconv.Quote(name)
	}

	end := 0

	for end < excerptLimit {
		_, size := utf8.DecodeRuneInString(name[end:])
		if end+size > excerptLimit {
			break
		}

		end += size
	}

	return strconv.Quote(name[:end]) + "... (" + strconv.Itoa(len(name)) + " bytes)"
}

// MustNew is like [New] but panics if New returns an error. The panic value
// is that error, which matches [ErrDuplicateName].
//
// MustNew is intended for map literals and other data known to be correct.
func MustNew(names map[int]string) *Catalog {
	c, err := New(names)
	if err != nil {
		panic(err)
	}

	return c
}

// Name returns the name of the octave with ASPN number aspn in the catalog,
// or "" if the catalog has no name for that octave.
//
// Any int is accepted; Name never fails or panics, including on a nil or
// zero catalog. The returned name is exactly the string given when the
// catalog was created.
func (c *Catalog) Name(aspn int) string {
	if c == nil {
		return ""
	}

	return c.names[aspn]
}

// ASPN returns the ASPN number of the octave that has the given name in the
// catalog.
//
// Names are compared exactly, byte for byte, without trimming,
// normalization or case folding. If the catalog has no octave with this
// name, including for the empty name, ASPN returns 0 and an error that
// matches [ErrUnknownName]. The returned 0 is not an answer: 0 is a valid
// ASPN number, so check the error to tell whether the name was found.
//
// For every named octave k of a catalog c, c.ASPN(c.Name(k)) returns k and a
// nil error.
func (c *Catalog) ASPN(name string) (int, error) {
	if c != nil {
		if number, ok := c.numbers[name]; ok {
			return number, nil
		}
	}

	return 0, fmt.Errorf("%w: %s", ErrUnknownName, excerpt(name))
}

// MustASPN is like [Catalog.ASPN] but panics if ASPN returns an error. The
// panic value is that error, which matches [ErrUnknownName].
func (c *Catalog) MustASPN(name string) int {
	number, err := c.ASPN(name)
	if err != nil {
		panic(err)
	}

	return number
}

// Equal reports whether c and other are the same catalog: they name the same
// octaves, and each octave has the same name in both, compared exactly.
//
// How the catalogs were created and whether the pointers are identical do
// not matter: Russian().Equal(Russian()) is true although the pointers
// differ. Pairs with an empty name are not part of a catalog, so the nil
// pointer, the zero value and every catalog without names are equal to one
// another. Equal does not modify either catalog.
func (c *Catalog) Equal(other *Catalog) bool {
	return maps.Equal(c.pairs(), other.pairs())
}

// All returns the named octaves of the catalog as a sequence of pairs of an
// ASPN octave number and its name, in ascending order of the number. Each
// named octave appears once; octaves with the empty name are not included,
// so a nil or zero catalog, or a catalog without names, gives an empty
// sequence. The result is never nil.
//
// The sequence may be stopped early, as with break in a range loop, and
// may be traversed any number of times, each time giving the same pairs.
// It is bound to the pairs that c holds when All is called: assigning to
// *c afterward does not change it. All and the traversal do not change the
// catalog, and the loop body may call any method of the catalog, including
// All. A sequence may be traversed by multiple goroutines at once.
//
// All does not copy the pairs: a full traversal takes time proportional to
// the number of named octaves, whatever their numbers are. All cannot fail,
// so it has no Must variant.
func (c *Catalog) All() iter.Seq2[int, string] {
	// Read *c once, here: the sequence captures the immutable ordered list,
	// not c, so a later assignment to *c neither changes nor races with it.
	var ordered []namedOctave
	if c != nil {
		ordered = c.ordered
	}

	return func(yield func(int, string) bool) {
		for _, o := range ordered {
			if !yield(o.number, o.name) {
				return
			}
		}
	}
}

// pairs returns the read-only map of named octaves of c; it is nil for a nil
// or zero catalog.
func (c *Catalog) pairs() map[int]string {
	if c == nil {
		return nil
	}

	return c.names
}
