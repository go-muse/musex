// Package system provides musical systems: the rules by which an admissible
// note name gets a class of positions, and a written pitch gets a position on
// the axis, a written octave, an order and distances in steps.
//
// A [System] is defined by five values (see [Definition]):
//
//   - N, the number of classes of positions in an octave, 1..[MaxClasses];
//   - b, the boundary letter: the letter at offset 0, just after the boundary
//     of the octave of positions;
//   - the offsets p(l) of the seven natural letters from b, in 0..N, which do
//     not go down along the circle of letters from b (letters may coincide);
//   - c, the chromatic unit: the number of classes by which the alteration
//     value +1 shifts a class, 1..[MaxChromatic];
//   - A, the [Rule] of admissible alteration values.
//
// The class of an admissible name (l, a) is (p(l) + a·c) mod N, a
// [github.com/go-muse/musex/pitchclass.Class] with the modulus N. The position
// of a written pitch with the Octave O is O·N + class. Positions are never
// exposed as numbers: the package has no reference pitch and no public origin
// of positions, and gives only their order and differences, measured in
// [Step] values.
//
// The package defines the following operations:
//
//   - [New], [FromSteps] and [NewRegular] create a system from offsets, from
//     steps between letters, or from a large and a small step;
//     [Traditional], [Positions19] and [QuarterTones24] return the built-in
//     systems, [Builtins] lists them and [LookupBuiltin] finds one by name.
//   - [System.Definition], [System.Classes], [System.Boundary],
//     [System.Offset], [System.Steps], [System.StepAfter],
//     [System.NaturalClass], [System.Chromatic], [System.Alterations] and
//     [System.Name] read the definition; [System.Regularity] and
//     [System.HasTraditionalLayout] describe its shape; [System.Equal] and
//     [System.NamesSubsetOf] compare systems.
//   - [NewProgression], [Progression.WithLower] and [Progression.WithUpper]
//     build progressions of alteration values; [NewRule], [Integers],
//     [Multiples], [IntegerRange], [Single] and [Union] build rules;
//     [Rule.Contains], [Rule.SubsetOf], [Rule.Equal] and [Rule.Values] query
//     them.
//   - [System.Admits], [System.Class], [System.Carry], [System.ShiftedClass],
//     [System.ShiftedClassByAlteration], [System.EquivalentNames],
//     [System.Spellings] and [System.EnharmonicNames] work with names.
//   - [System.AdmitsPitch], [System.PitchClass], [System.Compare],
//     [System.Between], [System.Sorted], [System.Distance],
//     [System.Equivalent] and [System.EnharmonicPitches] work with pitches.
//   - [System.WrittenOctave] and [System.FromWritten] convert between the
//     Octave of a pitch and its written octave; [System.RaiseWritten],
//     [System.LowerWritten], [System.RaiseWrittenBy],
//     [System.LowerWrittenBy], [System.RaiseWrittenByAlteration],
//     [System.LowerWrittenByAlteration], [System.RespellWritten] and
//     [System.NaturalWritten] change a name while keeping the written octave.
//   - [Sharps], [Flats], [ChromaticScale] and [CustomSpelling] create spelling
//     rules; [System.Spell], [System.PitchAt], [System.Move] and [System.Walk]
//     build names and pitches from classes and positions.
//
// Every operation that can return an error has a Must variant: it returns the
// same result or panics with the same error, which errors.Is recognizes after
// recover. Use the Must variants for literals and values known to be valid.
//
// # No default system
//
// Every operation on names and pitches takes the system explicitly; no system,
// not even the traditional one, is ever substituted silently. The same name can
// be admissible in one system and not in another: D raised by 1/2 is
// admissible in [QuarterTones24] but not in [Traditional]. Packages
// [github.com/go-muse/musex/pitch], [github.com/go-muse/musex/octave] and
// [github.com/go-muse/musex/spelling] know nothing about systems.
//
// # Octave and written octave
//
// The Octave of a [github.com/go-muse/musex/pitch.Pitch] is the octave of
// positions that holds the pitch, not the number written with the letter. A
// name whose shift p(l) + a·c leaves the range 0..N−1 lies in a neighboring
// octave of positions: its carry k = ⌊(p(l) + a·c)/N⌋ ([System.Carry]) is not
// 0. The written octave w, the octave whose number is written with the letter,
// is w = Octave − k. In the traditional system B♯ has the carry 1, so the
// record "B♯3" is the pitch (B♯, Octave 4) and sounds with C4; C♭ has the
// carry −1, so "C♭4" is the pitch (C♭, Octave 3). [System.WrittenOctave]
// returns w for a pitch, and [System.FromWritten] returns the pitch for a name
// and w; the written octave is an [github.com/go-muse/musex/octave.Octave],
// and its visible number is a matter of notation.
//
// # Keeping the written octave
//
// The methods of [github.com/go-muse/musex/pitch.Pitch] that change the
// alteration, such as Raise, keep the Octave: (B, Octave 3) raised gives
// (B♯, Octave 3), which is written "B♯2" and sounds with C3. The methods
// [System.RaiseWritten], [System.LowerWritten], [System.RaiseWrittenBy],
// [System.LowerWrittenBy], [System.RaiseWrittenByAlteration],
// [System.LowerWrittenByAlteration], [System.RespellWritten] and
// [System.NaturalWritten] keep the written octave instead: the record "B3"
// raised gives "B♯3", the pitch (B♯, Octave 4), one step above "B3". The
// result of a change from the name n to the name t is
// (t, Octave + k(t) − k(n)), and its position moves by s(t) − s(n), where
// s = p(letter) + a·c. Only this difference of carries is used, so a written
// octave beyond the int range is not an error when the result fits.
//
// # Letters on the next boundary
//
// An offset can equal N. Such a letter stands on the boundary of the next
// octave of positions: its natural name has the class 0 and the carry 1. In
// the five-class system with the offsets C 0, D 1, E 2, F 2, G 3, A 4, B 5, the
// record "B4" is the pitch (B, Octave 5) and sounds with C5.
//
// # Alteration rules
//
// A [Rule] is a union of at most [MaxProgressions] progressions. A
// [Progression] is a single value or all values r + n/d, n an integer, with a
// step 1/d, 1 <= d <= [MaxDenominator], optionally bounded below and above.
// Values outside the guaranteed range of
// [github.com/go-muse/musex/alteration] (|a| <= alteration.MaxMagnitude) are
// not part of any rule, so the integers and the integers from −2^24 to 2^24 are
// the same rule. Rules are compared as sets: (1/2)Z equals the union of Z and
// 1/2 + Z. A system requires 0 in A and an integer shift a·c for every
// admissible a. Membership costs O([MaxProgressions]); inclusion and equality
// are decided without enumerating values, in at most about 10^8 simple
// operations and typically in microseconds.
//
// # Choosing names
//
// Operations that build a name from a class or a position ([System.Spell],
// [System.PitchAt], [System.Move], [System.Walk]) always take an explicit
// [SpellingRule]; there is no default rule. [Sharps] and [Flats] choose the
// admissible name with the smallest |a|, preferring a > 0 or a < 0 on a tie,
// and then the first letter going up from the boundary letter.
// [ChromaticScale] uses the traditional spelling of the chromatic scale of a
// major or minor tonic and applies only to systems with the traditional layout
// ([System.HasTraditionalLayout]). [CustomSpelling] calls a user function: the
// function is called synchronously, possibly many times and possibly from many
// goroutines when the rule is used concurrently; a panic in it is not
// recovered; the package checks that the returned name is set, admissible and
// of the requested class.
//
// # Technical limits
//
// N <= [MaxClasses] (2^16), c <= [MaxChromatic] (2^16), a progression step is
// at least 1/[MaxDenominator] and a rule has at most [MaxProgressions] (8)
// progressions. These are engineering limits that keep all computations exact
// in 64-bit integers, not musical limits. An Octave, a written octave or a
// distance in steps outside int is reported with [ErrOutOfRange], never
// truncated or wrapped.
//
// # Errors
//
// Every error matches exactly one category with errors.Is: [ErrInvalid] (an
// unset input), [ErrDefinition] (an invalid definition, together with exactly
// one reason among [ErrClassCount], [ErrOffsets], [ErrDescending],
// [ErrChromatic], [ErrNoNatural], [ErrFractionalShift] and [ErrRule]),
// [ErrModulusMismatch], [ErrLimit], [ErrNotApplicable], [ErrInadmissible],
// [ErrNoSpelling], [ErrBadSpelling] or [ErrOutOfRange]. Checks run in this
// order: unset inputs, definitions, class moduli, limits, applicability of the
// spelling rule, input names, the range of a new alteration value, a new name,
// the spelling rule's choice, and the range of octaves and distances. On error
// a function returns the zero value of its result (nil for pointers and
// sequences) and leaves its inputs unchanged.
//
// # Unset values and concurrency
//
// The nil *System and *Rule, pointers to their zero values, the zero
// [Progression] and the zero [SpellingRule] are unset; methods on them do not
// panic. Systems, rules, progressions and spelling rules are immutable and
// safe for concurrent use; a custom spelling rule is as safe as its function.
// The sequences returned by the package can be stopped at any element and
// iterated many times, also concurrently, with the same result.
package system
