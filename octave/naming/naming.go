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
// [New] and [Catalog.WithAliases] return errors that match ErrDuplicateName,
// and [MustNew] and [Catalog.MustWithAliases] panic with such an error. Check
// for it with [errors.Is]; the text of the error is not part of the API.
var ErrDuplicateName = errors.New("naming: duplicate octave name")

// ErrUnnamedOctave reports that an alias is given for an octave that has no
// name in the catalog.
//
// [Catalog.WithAliases] returns errors that match ErrUnnamedOctave, and
// [Catalog.MustWithAliases] panics with such an error. Check for it with
// [errors.Is]; the text of the error is not part of the API.
var ErrUnnamedOctave = errors.New("naming: alias for an unnamed octave")

// Catalog is a catalog of octave names: an immutable set of pairs of an ASPN
// octave number and a non-empty name. Each octave has at most one name, and
// different octaves have different names. Octaves without a name in the
// catalog have the empty name "".
//
// Create a catalog with [New] or [MustNew], or take a built-in one with
// [Russian], [English] or [German]. No function or method changes a catalog
// after it is created.
//
// A catalog may also accept additional input names, aliases, for its named
// octaves; add them with [Catalog.WithAliases]. Aliases extend the lookup by
// name only: [Catalog.Name] and [Catalog.All] keep returning the canonical
// names, and [Catalog.Aliases] lists the aliases of an octave. The built-in
// catalogs have no aliases.
//
// The zero value, such as var c Catalog or &Catalog{}, and the nil pointer
// are valid catalogs without names. They behave like MustNew(nil): [Catalog.Name]
// returns "" for every number, [Catalog.ASPN] reports [ErrUnknownName] for
// every name, and [Catalog.Equal] reports them equal to any catalog without
// names or aliases. No method panics on them, except [Catalog.MustASPN] and
// [Catalog.MustWithAliases] as documented.
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
// catalog, keep the map of pairs it was created from and the aliases added
// with [Catalog.WithAliases].
type Catalog struct {
	// names maps an ASPN octave number to its non-empty name.
	names map[int]string
	// numbers maps each name in names back to its ASPN octave number.
	numbers map[string]int
	// ordered holds the pairs of names in ascending order of the number. It
	// is built once by New and only read afterward, so a sequence returned by
	// All may keep it without copying.
	ordered []namedOctave
	// aliases maps each additional input name to its ASPN octave number.
	// Canonical names are not stored here. It is built once by WithAliases
	// and only read afterward; copies of a catalog may share it.
	aliases map[string]int
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
// catalog, either as its canonical name or as one of its aliases.
//
// Names are compared exactly, byte for byte, without trimming,
// normalization or case folding. If the catalog has no octave with this
// name, including for the empty name, ASPN returns 0 and an error that
// matches [ErrUnknownName]. The returned 0 is not an answer: 0 is a valid
// ASPN number, so check the error to tell whether the name was found.
//
// For every named octave k of a catalog c, c.ASPN(c.Name(k)) returns k and a
// nil error. For every alias a of a named octave k, c.ASPN(a) returns k and
// c.Name(k) returns the canonical name of that octave.
func (c *Catalog) ASPN(name string) (int, error) {
	if c != nil {
		if number, ok := c.numbers[name]; ok {
			return number, nil
		}

		if number, ok := c.aliases[name]; ok {
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

// WithAliases returns a new catalog that also accepts the given additional
// input names, aliases, for its named octaves. The receiver is not modified:
// its names, aliases and lookup results stay as they were. Aliases already
// in the receiver are kept, so successive calls accumulate.
//
// aliases maps an ASPN octave number to the additional names of that octave.
// An alias is stored and compared exactly as given, byte for byte, like a
// canonical name. [Catalog.ASPN] and [Catalog.MustASPN] accept aliases as
// well as canonical names; [Catalog.Name] and [Catalog.All] keep returning
// the canonical names only. For an alias a of a named octave k,
// c.Name(c.MustASPN(a)) returns the canonical name of k. The built-in
// catalogs have no aliases.
//
// An alias must resolve to exactly one octave. WithAliases returns nil and
// an error that matches [ErrDuplicateName] when a non-empty alias would
// resolve to two different octaves: it equals another octave's canonical
// name or another octave's alias, including an alias the receiver already
// has. An alias that already resolves to the same octave is not ambiguous:
// it is accepted as a no-op and not stored twice, whether it repeats in the
// list, is already an alias of that octave in the receiver, or equals that
// octave's own canonical name. WithAliases is therefore idempotent, and
// alias lists from several sources can be merged.
//
// Empty alias strings are ignored, as in [New]: the empty string is not a
// name. A non-empty alias for an octave without a canonical name gives an
// error that matches [ErrUnnamedOctave]. The conditions are checked in this
// order:
//
//  1. an alias of an unnamed octave: the error matches [ErrUnnamedOctave];
//  2. an alias resolving to two different octaves: the error matches
//     [ErrDuplicateName].
//
// On error no catalog is created and the receiver is unchanged. The error
// message quotes a bounded excerpt of one offending alias and names the
// octaves involved; it does not depend on map iteration order.
//
// WithAliases does not modify or retain aliases: the new catalog keeps its
// own copy, and later changes to the map or its slices do not affect it. On
// success the result is a new non-nil *Catalog that the caller owns. A nil
// or zero receiver has no named octaves: an empty alias map gives a new
// catalog without names, and any non-empty alias gives [ErrUnnamedOctave].
// The caller must not modify aliases concurrently with the call.
func (c *Catalog) WithAliases(aliases map[int][]string) (*Catalog, error) {
	var names map[int]string
	var numbers, existing map[string]int

	if c != nil {
		names = c.names
		numbers = c.numbers
		existing = c.aliases
	}

	// Process the input in a fixed order, ascending octave number and then
	// alias, so the chosen error does not depend on map iteration order.
	octaves := make([]int, 0, len(aliases))
	for number := range aliases {
		octaves = append(octaves, number)
	}

	slices.Sort(octaves)

	merged := maps.Clone(existing)

	var unnamedNumbers []int
	var unnamedAliases []string

	type collision struct {
		alias   string
		numbers []int
	}

	var collisions []collision

	for _, number := range octaves {
		list := slices.Clone(aliases[number])
		slices.Sort(list)

		for _, alias := range list {
			if alias == "" {
				continue
			}

			if names[number] == "" {
				unnamedNumbers = append(unnamedNumbers, number)
				unnamedAliases = append(unnamedAliases, alias)

				continue
			}

			target, found := numbers[alias]
			if !found {
				target, found = merged[alias]
			}

			switch {
			case found && target == number:
				// Already resolves to this octave; do not store it twice.
			case found:
				collisions = append(collisions, collision{alias: alias, numbers: []int{target, number}})
			default:
				if merged == nil {
					merged = make(map[string]int)
				}

				merged[alias] = number
			}
		}
	}

	if len(unnamedNumbers) > 0 {
		// The input was processed in ascending order, so the first pair is
		// the smallest octave number with its smallest alias.
		return nil, fmt.Errorf("%w: %s for octave %d", ErrUnnamedOctave, excerpt(unnamedAliases[0]), unnamedNumbers[0])
	}

	if len(collisions) > 0 {
		chosen := collisions[0]
		for _, col := range collisions[1:] {
			if col.alias < chosen.alias {
				chosen = col
			}
		}

		var nums []int
		for _, col := range collisions {
			if col.alias == chosen.alias {
				nums = append(nums, col.numbers...)
			}
		}

		slices.Sort(nums)
		nums = slices.Compact(nums)

		return nil, fmt.Errorf("%w: %s for octaves %d and %d", ErrDuplicateName, excerpt(chosen.alias), nums[0], nums[1])
	}

	result := &Catalog{aliases: merged}
	if c != nil {
		result.names = names
		result.numbers = numbers
		result.ordered = c.ordered
	}

	return result, nil
}

// MustWithAliases is like [Catalog.WithAliases] but panics if WithAliases
// returns an error. The panic value is that error, which matches
// [ErrUnnamedOctave] or [ErrDuplicateName].
//
// MustWithAliases is intended for literals and other data known to be
// correct.
func (c *Catalog) MustWithAliases(aliases map[int][]string) *Catalog {
	result, err := c.WithAliases(aliases)
	if err != nil {
		panic(err)
	}

	return result
}

// Aliases returns the additional input names of the octave with ASPN number
// aspn, in ascending byte order. The canonical name of the octave is not
// included. The result is a new non-nil slice that the caller owns; it is
// empty when the octave has no aliases, including for an unnamed octave and
// on a nil or zero catalog. Aliases never fails or panics.
func (c *Catalog) Aliases(aspn int) []string {
	aliases := make([]string, 0)

	if c != nil {
		for alias, number := range c.aliases {
			if number == aspn {
				aliases = append(aliases, alias)
			}
		}
	}

	slices.Sort(aliases)

	return aliases
}

// Equal reports whether c and other are the same catalog: they name the same
// octaves, each octave has the same name in both, compared exactly, and both
// accept the same aliases for the same octaves. The order of aliases does
// not matter.
//
// Equal catalogs are interchangeable: Name, ASPN and All agree on them.
// Catalogs without aliases behave as before: they are equal exactly when
// they name the same octaves with the same names.
//
// How the catalogs were created and whether the pointers are identical do
// not matter: Russian().Equal(Russian()) is true although the pointers
// differ. Pairs with an empty name are not part of a catalog, so the nil
// pointer, the zero value and every catalog without names or aliases are
// equal to one another. Equal does not modify either catalog.
func (c *Catalog) Equal(other *Catalog) bool {
	return maps.Equal(c.pairs(), other.pairs()) && maps.Equal(c.aliasMap(), other.aliasMap())
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

// aliasMap returns the read-only map of aliases of c; it is nil for a nil or
// zero catalog and for a catalog without aliases.
func (c *Catalog) aliasMap() map[string]int {
	if c == nil {
		return nil
	}

	return c.aliases
}
