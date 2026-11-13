// Package frequency compares two frequencies in hertz and performs numeric
// frequency calculations. A [Criterion] is an immutable value that a consumer
// can accept from a user and apply without redefining comparison locally:
// strict, cents, hertz, relative, all-of, and any-of criteria have the same
// meaning for every consumer. The zero criterion is unset; there is no default
// comparison kind.
//
// # Comparing frequencies
//
// [Strict], [Equal], and [Compare] have no hidden tolerance: no epsilon, ULPs,
// decimal interpretation, or conversion to cents. Thus 0.1+0.2 computed in
// float64 does not equal 0.3. Inputs follow [cents.IsValidFrequency]: finite
// positive values, including subnormals. Frequency bounds are owned by cents
// and are not duplicated here.
//
// [WithinCents] compares the absolute distance from [cents.Between] with the
// tolerance exactly. That distance reads each frequency as its shortest
// decimal representation and rounds to a grid of 10^-6 cent. Consequently, a
// zero-cent tolerance is not strict: 440 Hz and the next float64 after it are
// close because their distance rounds to zero on that grid. [WithinHertz] and
// [WithinRelative], in contrast, use the exact binary values of their inputs
// and exact rational arithmetic, without float64 rounding error. A relative
// tolerance is measured from the greater frequency; a value of at least one
// makes every pair of valid frequencies close.
//
// Hertz are not a pitch distance: one hertz near 55 Hz is about 31 cents, but
// near 1760 Hz it is about one cent. Cents and relative tolerances describe
// frequency ratios and therefore pitch distances; an absolute hertz tolerance
// instead suits instruments with absolute measurement error. As domain
// guidance rather than fixed settings, musicians may distinguish differences
// of about 5–6 cents and tuners about 1–2 cents, depending on frequency,
// loudness, and timbre. In the domain example, 439 and 441 Hz are heard as the
// same A, slightly below and above a reference; this is an example, not a
// boundary of a pitch zone.
//
// Closeness with a tolerance is generally not transitive. [Criterion.Compare]
// returns zero for closeness and otherwise the strict numeric order, so it is
// not a sorting order for tolerant criteria. Use strict comparison for
// sorting.
//
// [AllOf] and [AnyOf] compose criteria. Their comparison result does not
// depend on component order, but [Criterion.Equal] and [Criterion.String]
// preserve order and distinguish that order. Nesting depth is unlimited:
// creation, comparison, equality, formatting, and parsing use iterative
// traversals rather than recursion. A zero-cent criterion can also be
// non-transitive because of the cents grid.
//
// # Numeric operations
//
// [Transpose] applies steps of an equal division of the octave; seven
// semitones of 12-EDO are not the just ratio 3:2. [Ratio] returns the reduced
// exact rational ratio of the shortest decimal representations, while
// [CentsBetween] returns a signed cents distance. [Partial] and [Partials]
// calculate f·n, and [Subharmonic] and [Subharmonics] calculate f/n. Index 1 is
// the fundamental and index 2 is the first overtone. The subharmonic formula
// and numbering are supported only by two weak domain sources and do not
// claim that all such tones are physically present in a sound.
//
// [BeatFrequency] is the absolute difference between two simple tones. It is
// not a measure of consonance or a numerical hearing threshold; zero means no
// periodic beating, not no sound. [ShiftOctaves] multiplies by a power of two.
// [NormalizeOctave] selects a representative in the exact half-open band
// [base, 2·base), using exact inputs even when 2·base is not representable as
// float64. Its result is a numeric octave equivalent, not a pitch class.
// [Period] returns 1/f seconds; milliseconds are only a derived display unit,
// and the period is neither a musical duration nor the duration of the whole
// sound. No operation returns a note, a spelled pitch, an interval name, or a
// pitch class.
//
// Ratio and the cents operations read shortest decimal representations. The
// other numeric operations use exact binary inputs. Transposition, partials,
// subharmonics, octave shifts, and octave normalization round their exact
// result once to a 53-bit mantissa with an unlimited exponent, check the
// normal-frequency bounds before converting to float64, and reject a result
// outside them. Beat frequencies and periods instead round the exact rational
// result directly to float64 without intermediate 53-bit rounding, and may be
// subnormal. Period rejects only overflow to infinity.
//
// Lists of partials and subharmonics are finite, explicit, and atomic: all
// indices are validated before calculation, and any failure returns no list.
// Every returned ratio and list is newly allocated and independently owned by
// the caller. Input slices are not modified or retained.
//
// # Errors, text, and JSON
//
// Frequency inputs are checked from left to right, then numeric parameters,
// then results. Criteria report [ErrInvalidFrequency], [ErrInvalidTolerance],
// [ErrUnset], [ErrNoComponents], and [ErrInvalidText]. Numeric operations also
// use [ErrInvalidParameter] and [ErrOutOfRange]. Errors are recognized with
// errors.Is. An invalid frequency also matches [cents.ErrInvalidArgument];
// errors caused by cents parsing or generated-frequency range checks preserve
// the corresponding cents sentinel. An invalid division count in Transpose
// also preserves cents.ErrInvalidArgument, while period overflow has no cents
// sentinel. Only Must wrappers panic, with the original error; a nil
// *Criterion receiver is a programming error rather than a criterion value.
//
// The canonical text forms are "unset", "strict", "cents(t)", "hertz(t)",
// "relative(t)", "all(c1, c2)", and "any(c1, c2)". Cents use the exact text
// rules of cents; numeric tolerances use the shortest float64 representation.
// Lists require a comma and exactly one space, keywords are lowercase ASCII,
// and the entire input must match. [Parse] checks the grammar of the whole
// text before returning the first semantic error. [Criterion.MarshalText],
// [Criterion.AppendText], and [Criterion.UnmarshalText] support text encoders,
// including JSON v1 and v2 as a string. Criterion is not comparable; use
// [Criterion.Equal] for equality and its canonical String as a map key.
//
// Strict comparisons on valid frequencies take constant time and allocate no
// memory. Other simple criteria take constant time; composites are linear in
// their tree nodes, parsing and formatting are linear in text length, and
// partial lists are linear in the number of indices. The package has no
// mutable global state or caches.
//
// Criteria and immutable results can be shared by concurrent readers. The
// caller synchronizes writes to a Criterion receiving UnmarshalText, use of an
// AppendText buffer whose spare capacity another goroutine accesses, writes
// to an input index slice during a call, and writes to a shared returned
// *big.Rat or result slice. Independent result slices and ratios need no such
// synchronization for read-only sharing.
package frequency
