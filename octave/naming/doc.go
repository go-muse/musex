// Package naming provides catalogs of verbal octave names: the name of an
// octave by its ASPN number, and the ASPN number of an octave by its name.
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
//   - [Catalog.Name] returns the name of an octave by its ASPN number.
//   - [Catalog.ASPN] and [Catalog.MustASPN] return the ASPN number of an
//     octave by its name.
//   - [Catalog.Equal] reports whether two catalogs are the same.
//
// # Octave numbers
//
// Octaves are identified by their numbers in the ASPN (American Standard
// Pitch Notation, also known as scientific pitch notation) octave
// numbering. An octave is the range of pitches from C up to the next B. The
// octave that contains middle C (the one-line octave) has number 4,
// neighboring octaves have consecutive numbers (the octave above it is 5,
// the octave below it is 3), and any int is a valid number, including
// negative ones. The package does not check or limit the numbers.
//
// The ASPN number identifies the octave itself. It is not a written octave
// number of a note, such as the octave in which a part for a transposing
// instrument is notated.
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
// an empty name passed to [New] is left out of the catalog.
//
// Names are stored and compared exactly, byte for byte. They are not
// trimmed, normalized or changed in case, so "One-line" or " one-line" is
// not found in the English catalog. Names are not notation symbols or
// inflected word forms; parsing and formatting of notation are outside this
// package.
//
// [Catalog.ASPN] returns an error that matches [ErrUnknownName] when a name
// is not in the catalog, and [New] returns an error that matches
// [ErrDuplicateName] when two octaves are given the same name. Check errors
// with [errors.Is]. [MustNew] and [Catalog.MustASPN] panic in the same cases
// and work the same way otherwise. No other function or method panics.
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
// copy *c.
package naming
