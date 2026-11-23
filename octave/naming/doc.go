// Package naming provides catalogs of verbal octave names: the name of an
// octave by its ASPN number, the ASPN number of an octave by its name, and
// the named octaves of a catalog in ascending order of their numbers.
//
// A [Catalog] is an immutable set of pairs of an ASPN octave number and a
// non-empty name. Each octave has at most one name in a catalog, and
// different octaves have different names. A catalog is a set of names only:
// it is not an octave numbering, a registry of catalogs or a notation, and
// it has no name of its own.
//
// The package defines the following operations:
//
//   - [Russian], [English] and [German] return the built-in catalogs.
//   - [New] and [MustNew] create a catalog from a map of pairs.
//   - [Catalog.WithAliases] and [Catalog.MustWithAliases] return a new
//     catalog that also accepts additional input names, aliases, for its
//     named octaves.
//   - [Catalog.Name] returns the name of an octave by its ASPN number.
//   - [Catalog.ASPN] and [Catalog.MustASPN] return the ASPN number of an
//     octave by its name or alias.
//   - [Catalog.Aliases] returns the aliases of an octave in ascending byte
//     order.
//   - [Catalog.Equal] reports whether two catalogs are the same.
//   - [Catalog.All] returns the named octaves of a catalog in ascending
//     order of their ASPN numbers.
//
// # Octave numbers
//
// Octaves are identified by their numbers in the ASPN (American Standard
// Pitch Notation, also known as scientific pitch notation) octave
// numbering. In the traditional system, an octave is the range of pitches
// from C up to the next B. The octave that contains middle C (the one-line
// octave) has number 4, neighboring octaves have consecutive numbers (the
// octave above it is 5, the octave below it is 3), and any int is a valid
// number, including negative ones. The package does not check or limit the
// numbers.
//
// The ASPN number identifies the octave itself. It is not a written octave
// number of a note, such as the octave in which a part for a transposing
// instrument is notated.
//
// Octave numberings, including ASPN, are provided by package
// [github.com/go-muse/musex/octave/number].
//
// # Built-in catalogs
//
// The package provides three built-in catalogs:
//
//   - [Russian] names the octaves with ASPN numbers -1 through 9.
//   - [English] names the octaves with ASPN numbers 0 through 7.
//   - [German] names the octaves with ASPN numbers 0 through 7.
//
// The sets of named octaves differ: the octaves with ASPN numbers -1, 8
// and 9 have names only in the Russian catalog. In every catalog, the other
// octaves have the empty name. The Russian names "субсубконтроктава"
// (ASPN -1) and "шестая" (ASPN 9) are an extension of musex: they were not
// found in the literature.
//
// Built-in catalogs follow the same rules as custom ones. For example,
// German() is equal by [Catalog.Equal] to a catalog created with [MustNew]
// from the same eight pairs. Each call of a built-in function returns a new
// *Catalog, so overwriting the object it points to does not affect other
// callers.
//
// # Empty names and lookup
//
// The empty name "" means that an octave has no name. It is not an error:
// [Catalog.Name] returns "" for every octave without a name, and a pair with
// an empty name passed to [New] is left out of the catalog. [Catalog.All]
// leaves out octaves without a name.
//
// Names are stored and compared exactly, byte for byte. They are not
// trimmed, normalized or changed in case, so "One-line" or " one-line" is
// not found in the English catalog. Names are not notation symbols or
// inflected word forms; parsing and formatting of notation are outside this
// package.
//
// # Aliases
//
// [Catalog.WithAliases] returns a new catalog that also accepts additional
// exact input names for its named octaves, for example "первая октава" and
// "FirstOctave" alongside the canonical "первая". The receiver is not
// modified, successive calls accumulate, and an alias that already resolves
// to the same octave is accepted as a no-op, so alias lists from several
// sources can be merged. Aliases extend the lookup only: [Catalog.Name] and
// [Catalog.All] keep returning the canonical names, and [Catalog.Aliases]
// lists the aliases of an octave. The built-in catalogs have no aliases.
//
// An alias must resolve to exactly one octave: an alias equal to another
// octave's canonical name or alias is an error. An alias for an octave
// without a canonical name is an error too; empty alias strings are ignored,
// as in [New]. [Catalog.Equal] compares the canonical names and the
// alias-to-octave mapping, so equal catalogs accept the same input.
//
// [Catalog.ASPN] returns an error that matches [ErrUnknownName] when a name
// is not in the catalog, [New] returns an error that matches
// [ErrDuplicateName] when two octaves are given the same name, and
// [Catalog.WithAliases] returns errors that match [ErrDuplicateName] for an
// ambiguous alias and [ErrUnnamedOctave] for an alias of an unnamed octave.
// Check errors with [errors.Is]. [MustNew], [Catalog.MustASPN] and
// [Catalog.MustWithAliases] panic in the same cases and work the same way
// otherwise. No other function or method panics.
// An error quotes a name of up to 32 bytes in full. A longer name is
// quoted by at most its first 32 bytes, without splitting a valid UTF-8
// character, followed by its length in bytes, so the length of an error
// message is bounded whatever the length of the name. The exact text of
// errors is not part of the API.
//
// The zero value of [Catalog] and the nil *Catalog are catalogs without
// names: they behave like a catalog created from an empty map.
//
// # Concurrency
//
// Catalogs are immutable, and the package has no mutable state. All
// functions and methods are safe for concurrent use by multiple goroutines,
// on the same catalog or on different ones, under the usual Go rules for
// writes: do not modify the map passed to [New] or [MustNew] during the
// call, and do not assign to *c while other goroutines call methods on c or
// copy *c. A sequence returned by [Catalog.All] may be traversed by
// multiple goroutines at once.
package naming
