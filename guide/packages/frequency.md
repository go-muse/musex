# Frequency

Frequency provides one shared way to compare two frequencies in hertz and a set of numeric operations over frequencies. A consumer can accept a comparison criterion from a user and apply it without inventing a local rule: every kind has the same meaning everywhere. A criterion is immutable; an unset criterion is not silently replaced by a default.

## Comparing frequencies

Strict comparison means exact numeric equality, with no epsilon, units in the last place, decimal reinterpretation, or conversion to cents. Thus a sum of decimal fractions computed in binary floating-point is not necessarily equal to the nearest decimal literal. Strict comparison also gives the ordinary numeric order and performs no memory allocation for valid inputs.

A cents tolerance uses the distance between the shortest decimal representations of the two frequencies, rounded to a grid of one millionth of a cent. Consequently, zero cents is not the same as strict comparison: 440 Hz and the immediately following representable number are close because their tiny distance rounds to zero on that grid. Equality at the tolerance boundary is included.

An absolute tolerance in hertz and a relative tolerance instead use the exact binary values of the inputs, with exact difference, multiplication, and comparison. A relative tolerance is measured from the greater frequency; a value of at least one makes every pair of valid frequencies close. Positive subnormal frequencies are valid inputs.

Hertz are not a pitch distance. One hertz near 55 Hz is about 31 cents, while one hertz near 1760 Hz is about one cent. Cents and relative tolerances describe frequency ratios and therefore pitch distances; an absolute hertz tolerance suits instruments whose measurement error is expressed in hertz.

Closeness with a tolerance is generally not transitive. For example, neighboring frequencies one hertz apart can be close in pairs while the outer pair is not close. The three-way comparison returns “close” or the strict lower/higher direction, but it is not a sorting order for a tolerant criterion. Use strict comparison for sorting.

As guidance rather than fixed limits, the domain knowledge base describes discrimination thresholds of roughly 5–6 cents for a musician and 1–2 cents for a tuner, depending on frequency, loudness, and timbre. Its example treats 439 and 441 Hz as the same A, slightly below and above a reference. That example does not define the boundaries of a pitch zone.

Composite criteria require either every component or at least one component to report closeness. Their behavior does not depend on component order, but equality and canonical text preserve that order. Components may be nested without a depth limit; operations use iterative traversal rather than recursion.

## Numeric operations

Uniform transposition multiplies a frequency by a power of an equal division of the octave. Seven equal semitones are not identical to a just 3:2 fifth. A frequency ratio is returned as a reduced rational fraction of the shortest decimal representations: 660 to 440 gives 3/2, while 442 to 440 gives 221/220. A separate signed cents deviation follows the cents grid. Neither operation guesses a nearby simple ratio or assigns an interval name.

Partial tones multiply the fundamental by an integer index, and subharmonics divide it by an integer index. Index 1 is the fundamental; index 2 is the first overtone. The subharmonic definition and numbering are supported only by two weak domain sources, and the formula does not claim that all such tones are physically present in a sound. Explicit finite lists preserve index order and are atomic: every index is checked before calculation, and any failure returns no partial list. Every returned list is independently owned by the caller.

Beat frequency is the absolute difference between two simple tones. It is not a measure of consonance or a numerical hearing threshold. Zero means no periodic beating, not absence of sound. Octave shifting multiplies by a power of two. Octave normalization chooses a representative in the exact half-open interval from the base up to, but excluding, twice the base. The upper endpoint need not be representable as a floating-point number. The representative is only a numeric octave equivalent, not a pitch class or a note name.

The oscillation period is the reciprocal frequency in seconds. Milliseconds are merely a derived display unit; the period is neither a musical duration nor the duration of the entire sound. No numeric operation returns a note, a spelled pitch, or an interval name.

Ratios and cents distances use shortest decimal representations. The other calculations use exact binary inputs. Generated frequencies from transposition, partials, subharmonics, octave shifts, and normalization are rounded once to a 53-bit mantissa with an unlimited exponent and must lie within the normal-frequency range, checked before conversion to the result type. Beat frequencies and periods are instead rounded directly from their exact rational values and may be subnormal, avoiding an extra intermediate rounding. An infinite period is rejected.

Each returned rational ratio is also a new value owned by the caller. Input slices are not modified or retained. Immutable criteria and independent results can be read concurrently; callers synchronize access to mutable buffers, slices, and rational values that they share for writing.

## Text and errors

Canonical criterion text is lowercase ASCII and can be stored as a configuration string or JSON string. Examples include `strict`, `cents(5)`, `hertz(0.5)`, `relative(0.001)`, and `any(cents(1), hertz(0.5))`. Composite lists use a comma followed by exactly one space. Parsing validates the grammar of the entire text before reporting the first semantic error, and canonical text round-trips to an equal criterion, including an unset one.

Errors distinguish an invalid frequency, an invalid tolerance, an unset criterion, an empty composite, invalid text, an invalid numeric parameter, and an out-of-range result. Frequency inputs are checked from left to right, then parameters, then results. Convenience variants panic with the original error only when their names explicitly request that behavior.

First released in v0.19.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/frequency) · [Go examples](../../frequency/example_test.go)
