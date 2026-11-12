// Package notation formats and parses written note names and written pitches
// according to explicit, unambiguous textual notations.
//
// A Notation combines immutable NameRules with an OctaveStyle and, where the
// style needs one, a separator. NameRules define a finite domain: every name
// has exactly one canonical output string and may also have explicit aliases
// accepted only during parsing. Different names cannot share a canonical form
// or an ambiguous alias. Formatting never substitutes an enharmonic spelling,
// rounds an alteration, normalizes Unicode, changes case implicitly, or trims
// whitespace. Strings are compared exactly as bytes, although invalid UTF-8 is
// allowed when a custom rule explicitly contains it.
//
// # Names and custom notations
//
// NewNameRules supports generated letter-and-accidental forms and explicit
// whole-name tables. Whole-name exceptions replace both the generated
// canonical form and its generated aliases. NewUpsAndDownsNameRules creates
// 105 names for a chromatic unit c in the range 3..65536: seven letters, five
// ordinary alterations, and three arrow states. An up arrow adds exactly 1/c
// and a down arrow subtracts exactly 1/c. The constructor rejects c=2, where
// these forms would be ambiguous, and values outside the supported range.
//
// New checks the complete grammar, including aliases, separators, letter-case
// transformations, and octave marks. Its ambiguity checks are symbolic over
// the whole octave axis rather than samples of a few octaves. Construction
// copies input slices. Values and built-in tables are immutable, contain no
// parse caches, and are safe for concurrent use.
//
// # Written pitches and context
//
// Name-only operations need no musical system. Pitch operations require an
// explicit system.System; the traditional system is never inserted silently.
// The system validates the name and is solely responsible for converting
// between the written octave w and the positional octave stored in pitch.Pitch.
// For example, in the traditional system notated B sharp 3 and C4 have
// positional octave 4, while C flat 4 has positional octave 3. This package
// does not calculate those carries itself. An ups-and-downs notation and the
// supplied system must have the same chromatic unit.
//
// OctaveStyleNumber requires an explicit octave/number.Numbering.
// OctaveStyleName uses the supplied octave/naming.Catalog. Numbering and
// catalogs are independent of the name rules and of each other. A nil, zero,
// or empty catalog is a valid empty catalog: formatting reports ErrNoOctave,
// and parsing preserves naming.ErrUnknownName. Visible numbers have a strict
// decimal syntax without a plus sign, leading zeroes, or negative zero.
//
// OctaveStyleAM supports only ASPN octaves 0 through 8. Helmholtz, numbered
// Helmholtz, LilyPond, and ABC forms are not restricted to the MIDI range;
// their limits are the technical limits of int and the result string. ABC
// parsing accepts mixed apostrophes and commas and adds them algebraically,
// while canonical output uses only one kind of mark. Arithmetic overflow is
// checked before allocating a result and reports octave.ErrOutOfRange.
//
// Convert parses a name with one notation and canonically formats it with
// another. ConvertPitch does the same for a pitch, using one explicit system
// for both directions and preserving the name and positional octave.
//
// # Built-in records and limits
//
// BuiltInIDs lists 19 records in their specified order. BuiltIn obtains the
// 18 non-parameterized records without allocating memory; ASCIIUpDownID
// requires exactly one chromatic-unit argument.
//
//   - english-unicode: 35 names, alterations -2..+2, visible octave numbers.
//   - english-ascii-mt21: 28 names, alterations -1..+2, visible numbers; it
//     has no canonical double-flat form.
//   - german-am: 33 names, alterations -2..+2, AM octaves 0..8. A double-flat
//     and E double-flat are not writable because their forms are unconfirmed.
//   - russian-am: 35 names, alterations -2..+2, AM octaves 0..8.
//   - romance-italian: 35 Italian syllabic names, alterations -2..+2, no
//     octave operation.
//   - abc: 35 names, alterations -2..+2, ABC octaves.
//   - lilypond-nederlands, lilypond-english, and lilypond-italiano: 63 names
//     each, alterations -2, -3/2, -1, -1/2, 0, +1/2, +1, +3/2, +2, with
//     LilyPond octave marks. English uses short canonical suffixes and accepts
//     the documented long aliases.
//   - unicode-stein: 63 names with Stein-Zimmermann forms, including SMuFL
//     Private Use Area code points; no octave operation.
//   - unicode-gould: 77 names, adding alterations -5/2 and +5/2 and Unicode
//     aliases; no octave operation.
//   - unicode-sims: 77 names, adding alterations -1/3, -1/6, +1/6, and +1/3;
//     no octave operation.
//   - unicode-wyschnegradsky: 175 names with values n/6 from -11/6 through
//     +11/6 plus ordinary and double accidentals; no octave operation.
//   - unicode-aeu: 77 names whose alteration values are expressed in AEU
//     sharps of four commas; no octave operation.
//   - unicode-persian: 49 names, with koron and sori nominally at -1/2 and
//     +1/2; no octave operation.
//   - unicode-arabic: 63 names with alterations -2 through +2 in half-unit
//     steps; no octave operation.
//   - unicode-stockhausen: 56 names with alterations -2, -1, -1/2, 0, +1/2,
//     +1, +3/2, and +2; no octave operation.
//   - ascii-updown: 105 names for each c in 3..65536; no octave operation.
//   - ascii-fokker: 63 names with ASCII Fokker forms; no octave operation.
//
// The SMuFL Private Use Area code points are not presented as Unicode-assigned
// symbols. There is no claim of one universal standard for fractional
// accidentals. No built-in German-letter record with an ASPN suffix such as H4
// is provided; users can construct an explicit custom combination instead.
//
// # Values, errors, and resource use
//
// Notation{} and NameRules{} are the only undefined values. They do not equal
// anything, including themselves. Both types are deliberately not comparable;
// use their Equal methods. Assignment and value copying are allowed and share
// only immutable tables. Names and BuiltInIDs return new caller-owned slices.
// A debugging name does not affect equality or behavior.
//
// Errors are classified with errors.Is. ErrUndefined, ErrUnknownID,
// ErrMissingChromaticUnit, ErrInvalidChromaticUnit, ErrInvalidDefinition,
// ErrAmbiguous, ErrParse, ErrUnwritableName, ErrNoOctave, and
// ErrIncompatibleSystem distinguish package-level causes. Errors from system,
// octave, octave/number, and octave/naming are wrapped with context and remain
// distinguishable. Every operation that returns an error has a Must wrapper;
// the wrapper returns the same result on success and panics with the same error
// value on failure.
//
// IsDefined, Equal, and non-parameterized BuiltIn lookups do not allocate.
// Formatting and parsing may allocate their result, errors, and temporary
// values whose lifetime ends with the call. Dependencies may allocate under
// their own contracts. No notation retains call results, temporary tables,
// input text, pitches, systems, numberings, or catalogs after returning.
package notation
