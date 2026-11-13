// Package settings provides a set of settings: one immutable value that
// joins a musical system, a tuning and the designations of written pitches,
// together with rules for replacing one set by another.
//
// # The set
//
// A Settings has three parts:
//   - the system (github.com/go-muse/musex/system): the classes, the
//     admissible names, the carry of a name and the written octave;
//   - the tuning (github.com/go-muse/musex/tuning): a temperament, its root and
//     a reference; the system of the set is always the system of its tuning,
//     so there is only one system;
//   - the designations: a notation (github.com/go-muse/musex/notation), an
//     octave numbering (github.com/go-muse/musex/octave/number) and a catalog
//     of octave names (github.com/go-muse/musex/octave/naming). They are
//     independent of each other.
//
// New creates a set and checks that the parts fit together: when the notation
// is an ups-and-downs notation, its chromatic unit must equal the chromatic
// unit of the system. A notation without octave marks is allowed: such a set
// formats and parses names, but expresses no pitch. A catalog is never unset:
// nil, a pointer to the zero catalog and a catalog without names are the empty
// catalog. Standard returns the standard settings: the traditional system,
// 12-EDO rooted at C with A4 at 440 Hz, the english-unicode notation, the ASPN
// numbering and the English catalog. Nothing is ever substituted silently:
// unset parts are never replaced by standard ones.
//
// WithTuning, WithNotation, WithNumbering, WithCatalog and WithDesignations
// return a new set with one part or all designations replaced. WithSystem
// translates the tuning to another system with the same number of classes by
// tuning.Tuning.WithSystem (the same temperament, root number and reference)
// and keeps the designations; to change the number of classes, replace the
// whole tuning.
//
// Equal compares all parts by their own Equal. Two unset sets are equal; an
// unset and a set one are not. String names all five parts for debugging.
//
// # Values, copies and ownership
//
// A Settings is used through a pointer and never changes. The nil pointer and
// a pointer to the zero value are the unset set: IsValid reports false,
// System, Tuning and Catalog return nil, Notation and Numbering return their
// undefined values, OctaveContext returns the zero context, String returns
// "Settings(unset)", Settings returns the receiver, Equal follows the rule
// above, and every other operation returns an error that matches ErrUnset.
// Methods on the unset set do not panic.
//
// The set protects itself with copies. New and the With methods copy the
// tuning and the catalog they receive; Tuning, Catalog and OctaveContext
// return new copies. Assigning through a pointer given to or received from
// the set does not change the set. Notation and Numbering are values whose
// copies share only immutable tables. Two things are not protected: the
// shared *system.System returned by System must not be overwritten through
// the pointer, because that would change the system for the whole program (a
// rule of package system), and assigning *s = … through a *Settings changes
// the set for every holder of that pointer. The package never makes such
// assignments, and Standard returns a new pointer on every call, so the
// standard settings cannot be changed this way.
//
// # Source
//
// A Source gives the current set whole through its only method Settings.
// A set is itself a source that returns itself; an instrument holding a
// current set is another source. A reader calls Settings at the moment of an
// operation and works with that snapshot. Implementations must be safe for
// concurrent calls, return either the previous or the new set whole and never
// a mixture of parts, not change the source, not block for long, and return
// the unset set when they have no set. Their dynamic types should be
// comparable, preferably pointers.
//
// SameSource decides whether two values are one source without calling
// them: two sets are one source when they are Equal (two results of Standard
// are one source); values of other types are one source when they are the
// same comparable value, such as the same instrument; a nil interface is no
// source; incomparable values never match and never cause a panic.
//
// # Names and pitches
//
// A name is expressible by a set when it is admissible in the system of the
// set and belongs to the name domain of its notation. A pitch is expressible
// when its name is expressible, the notation shows octaves, and its written
// octave w = Octave − k (k is the carry of the name in the system) is writable
// by the octave style of the notation with the numbering and the catalog of
// the set:
//   - number: the numbering gives w a visible number that fits in int;
//   - name: the catalog has a non-empty name for w;
//   - AM: 0 <= w <= 8 (ASPN);
//   - Helmholtz, Helmholtz with English repeats, LilyPond and ABC: the counts
//     of repeated marks and the length of the result fit in int, exactly as
//     package notation checks them before it builds the text;
//   - numbered Helmholtz: 2 − w fits in int for w <= 2, always for w >= 3.
//
// A written octave outside int (for example B♯ at the smallest Octave in the
// traditional system) and a visible number outside int give "not
// expressible", not an error.
//
// Expresses and ExpressesPitch answer the question. Format, Parse,
// FormatPitch and ParsePitch format and parse with the notation, the system,
// the numbering and the catalog of the set; OctaveContext gives the same
// context for direct calls of package notation. The operations agree:
// Expresses is true exactly when Format succeeds, ExpressesPitch is true
// exactly when FormatPitch succeeds (see "Huge octaves" below), and every
// result of a successful Parse or ParsePitch is expressible.
//
// # Huge octaves
//
// For the styles with repeated marks (Helmholtz, Helmholtz with English
// repeats, LilyPond and ABC) the text of an octave far from the middle is a
// string of very many marks. ExpressesPitch reports true for such an octave
// as long as the length fits in int, and it never builds the text.
// FormatPitch and MustFormatPitch delegate the text to package notation,
// which then needs a string of that length: for lengths of about 2^48 bytes
// and more (the limit depends on the platform; for the three-byte mark ′
// from about 2^46 marks) it panics with "runtime error: makeslice: len out of
// range", and for smaller but still enormous lengths the process can run out
// of memory. This is the behavior of package notation, not an error of this
// package; settings does not recover from it or change it. The equivalence of
// ExpressesPitch and a successful FormatPitch holds for octaves whose text
// fits in memory.
//
// # Compatibility
//
// CheckReplacement checks whether a next set can replace a previous one. Two
// rules apply together:
//   - RuleStructure: every name admissible in the previous system is
//     admissible in the next one (system.System.NamesSubsetOf). The
//     counterexample is a name admissible in the previous system and not in
//     the next one;
//   - RuleDesignations: every name expressible by the previous set is
//     expressible by the next one, and so is every pitch; each set uses its
//     own system, so the carry of a name can differ between the sets. The
//     counterexample is a name or a pitch.
//
// Expressibility in RuleDesignations treats a style that shows any octave
// (number, the Helmholtz styles, LilyPond, ABC) as showing every octave: the
// technical limits of int are not a violation. The check is exact and finite:
// the name domains are finite, and the writable octaves are either all
// octaves or a finite set (AM, a catalog). Every counterexample in a report is
// exact: for RuleStructure the name is admissible in the previous system and
// not in the next one; for RuleDesignations the name or the pitch is
// expressible by the previous set (Expresses, ExpressesPitch) and not by the
// next one. Which counterexample is chosen is not part of the contract, but
// one pair always gives an equal (==) report.
//
// CheckStructure and CheckDesignations check one rule each; CheckSystems
// checks RuleStructure for two systems before a set is built. A Report tells
// which rules were checked (Checked), whether all checked rules hold
// (Compatible), the violations in the order RuleStructure, RuleDesignations
// (Violations, Violation) and, for an owner such as an instrument, an error
// that matches ErrIncompatible (Err). A partial check speaks only about its
// own rule. Incompatibility is a result, not an error: the checks fail only
// with ErrUnset, for an unset set or system.
//
// Notes are not taken into account: the check compares only the two sets. A
// refusal can therefore be false for the notes actually in use; that is not
// an error of the check, and a new set can still be created separately,
// without a replacement. The expected consequences of a compatible change are
// not violations either: frequencies, step numbers and positions change; the
// written octave number of the same name and Octave can change (in the
// traditional system B♯ at Octave 4 is written B♯3; with 19 positions it is
// written B♯4); which spellings sound the same, the availability of MIDI and
// the regularity of the system can change.
//
// # Kinds of change
//
// Changes lists the kinds of change of a replacement; one replacement can
// have several:
//   - Structure: StructureUnchanged when the systems are Equal; otherwise
//     StructureExtended, StructureNarrowed, StructureSameNames or
//     StructureIncomparable by the admissible names of both systems;
//   - Sound: the temperament, the root number (as a number, without the
//     modulus) or the reference changes; a temperament with another number
//     of classes is always a change. A change of sound alone is always
//     compatible;
//   - Designations: the notation, the numbering or the catalog changes.
//
// Changes.None is true exactly when the two sets are Equal. A system with
// another number of classes always changes the sound too; WithSystem with the
// same number of classes changes only the structure.
//
// # Errors
//
// Errors are classified with errors.Is; their texts are not part of the API:
//   - ErrUnset: an unset set (the receiver or the next set) or an unset
//     tuning, notation, numbering or system;
//   - ErrInconsistent: the chromatic unit of an ups-and-downs notation differs
//     from the chromatic unit of the system; also matches
//     notation.ErrIncompatibleSystem;
//   - ErrUnsetName: an unset name or pitch;
//   - ErrInadmissible: a name that is not admissible in the system of the set
//     when a name or a pitch is formatted or parsed; also matches
//     system.ErrInadmissible;
//   - ErrIncompatible: only from Report.Err.
//
// The package's own errors never wrap each other. Errors of system, tuning,
// notation, octave, octave/number and octave/naming are returned with context
// and stay available through errors.Is. The first cause found is reported,
// in this order:
//   - New: ErrUnset (tuning, notation, numbering), then ErrInconsistent;
//   - WithTuning, WithNotation, WithDesignations: ErrUnset, then
//     ErrInconsistent; WithNumbering and WithCatalog: ErrUnset;
//   - WithSystem: ErrUnset, then the error of tuning.Tuning.WithSystem
//     (tuning.ErrClassCount, tuning.ErrInadmissible), then ErrInconsistent;
//   - Expresses, ExpressesPitch: ErrUnset, then ErrUnsetName; every other
//     case is an answer;
//   - Format, FormatPitch: ErrUnset, ErrUnsetName, ErrInadmissible, then the
//     errors of package notation and its dependencies in the order of the
//     notation contract;
//   - Parse: ErrUnset, the parse error of package notation, then
//     ErrInadmissible for the parsed name;
//   - ParsePitch: ErrUnset, then the errors of package notation in the order
//     of its contract. Notation recovers the written octave before the system
//     checks the name, so for text with both a wrong octave and an
//     inadmissible name the error names the octave; when the system reports
//     an inadmissible name, the error also matches ErrInadmissible;
//   - the checks of compatibility and Changes: ErrUnset only.
//
// On error a function returns the zero value of its result (nil, false, "",
// the unset name or pitch, Report{}, Changes{}) and leaves its inputs
// unchanged. Every operation with an error has a Must variant that calls it
// and panics with the same error value, which errors.Is recognizes after
// recover.
//
// # Concurrency and cost
//
// All values are immutable; the package has no mutable global state, caches
// or lazy initialization: the template of the standard settings is built when
// the package is initialized and only read afterward. Every function and
// method is safe for concurrent use on the same values, and implementations
// of Source must be as well.
//
// IsValid, System, Notation, Numbering, Settings and Equal do not allocate
// (Equal when the Equal methods of the parts do not allocate, as for all
// built-in parts), and neither does SameSource for two sets or for
// implementations that are pointers. Standard, Tuning, Catalog and
// OctaveContext allocate one new value each: the price of the protection by
// copies. New and the With methods allocate a new set and copies of the parts
// they receive. ExpressesPitch does not allocate memory that depends on the
// octave. A check of compatibility tries the names of the previous notation
// (up to a few hundred) and, for pitches, a few octaves per distinct pair of
// carries; it does not depend on the range of octaves. Finding the
// counterexample of RuleStructure follows the scheme of system.Rule.SubsetOf:
// microseconds for ordinary rules and about 10^8 simple operations in the
// worst case.
package settings
