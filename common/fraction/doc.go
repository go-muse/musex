// Package fraction provides exact signed rational numbers (ordinary fractions)
// that preserve the supplied numerator and denominator pair while supporting
// canonical reduction, exact arithmetic with explicit denominator-selection
// modes, whole-period extraction, divisibility checks, explicit rounding, and
// text conversion.
//
// The package belongs to the numeric foundation of musex and is independent of
// musical units: it does not know chromatic steps, note letters, spellings,
// octaves, pitch classes, cent scales, frequencies, tuplets, durations, or
// measures. Higher-level packages supply musical meaning and domain limits.
//
// # Representation, zero value, and canonical reduction
//
// A [Fraction] is an immutable value type of fixed size that stores both the
// original numerator n and the original denominator d in a symmetric 256-bit
// signed range ([-MaxComponent, +MaxComponent], where [MaxComponent] is
// 2^255 - 1). Because the original pair is preserved, fractions such as 6/12
// (position 6 in a 12-step circle or six quarter-units), 3/2 and 6/4 (tuplet
// ratios), 1/-2, and 0/3 keep their supplied components in [Fraction.Numerator],
// [Fraction.Denominator], [Fraction.Pair], and [Fraction.String].
//
// Every valid fraction also has a unique reduced canonical form (n_red, d_red)
// with a strictly positive denominator d_red > 0, gcd(|n_red|, d_red) == 1, the
// sign carried by n_red, and (0, 1) for numeric zero (so 6/12 reduces to 1/2,
// 1/-2 reduces to -1/2, and 0/3 reduces to 0/1). [Fraction.Reduce] returns a
// new fraction whose stored pair is (n_red, d_red), and [NewReduced],
// [FromBigIntReduced], and [FromBigRat] construct reduced fractions directly.
//
// The uninitialized zero value Fraction{} has numerator 0 and denominator 0
// (0/0) and represents an invalid, unset fraction rather than numeric zero
// ([Zero], 0/1, or 0/3). For Fraction{}, [Fraction.IsValid] and all other
// boolean predicates (including [Fraction.IsZero], [Fraction.Equal], and
// [Fraction.EqualExact]) return false, [Fraction.String] returns "0/0", and
// operations that return an error fail with [ErrInvalid].
//
// # Equality, ordering, and map keys
//
// Two equality relations are provided:
//
//   - [Fraction.Equal] reports numeric equality in Q (6/12, 1/2, and -2/-4 are
//     equal; 0/1, 0/3, and 0/-5 are equal; 1/-2 and -1/2 are equal).
//   - [Fraction.EqualExact] reports equality of the stored pair (n1 == n2 and
//     d1 == d2). For valid fractions, Go's built-in == operator matches
//     [Fraction.EqualExact].
//
// Because [Fraction] is comparable, it can be used directly as a map key when
// distinct stored pairs such as 3/2 and 6/4 must remain separate keys, whereas
// f.Reduce() can be used as a map key when numerically equal fractions must map
// to the same entry.
//
// [Fraction.Compare], [Fraction.Less], [Fraction.LessOrEqual],
// [Fraction.Greater], and [Fraction.GreaterOrEqual] order fractions by numeric
// value. [Fraction.CompareExact] defines a total deterministic order on pairs
// (numeric value first, then stored denominator, then stored numerator, with
// invalid Fraction{} ordered before all valid fractions).
//
// # Denominator modes and arithmetic
//
// Binary arithmetic supports seven explicit denominator-selection modes:
//
//  1. [DenomReduced] (and default [Fraction.Add], [Fraction.Sub],
//     [Fraction.Mul], [Fraction.Div], and *Int methods): fully reduces the
//     exact result to canonical form with d > 0 (0/1 for zero).
//  2. [DenomUnreducedLCM]: uses lcm(|d1|, |d2|) for addition and subtraction
//     without post-reduction (1/6 + 1/6 = 2/6, 1/4 - 1/4 = 0/4), |d1|*|d2| for
//     multiplication, and |d1*n2| for division.
//  3. [DenomUnreducedProduct]: uses the direct product |d1|*|d2| for addition,
//     subtraction, and multiplication without reduction (1/6 + 1/6 = 12/36),
//     and |d1*n2| for division.
//  4. [DenomSameUnit]: requires |d1| == |d2| == D and keeps denominator D
//     without reduction, returning [ErrNotRepresentable] if |d1| != |d2| or if
//     a product/quotient is not an integer multiple of 1/D.
//  5. [DenomLargerUnit]: expresses the exact result in the larger unit
//     D = min(|d1|, |d2|), or returns [ErrNotRepresentable].
//  6. [DenomSmallerUnit]: expresses the exact result in the smaller unit
//     D = max(|d1|, |d2|), or returns [ErrNotRepresentable].
//  7. [Fraction.AddInDenominator], [Fraction.SubInDenominator],
//     [Fraction.MulInDenominator], [Fraction.DivInDenominator], and
//     [Fraction.WithDenominator] / [Fraction.WithBigDenominator]: express the
//     exact result in a caller-supplied positive denominator D > 0, or return
//     [ErrNotRepresentable].
//
// All unary ([Fraction.Neg], [Fraction.Abs], [Fraction.Reciprocal],
// [Fraction.Inc], [Fraction.Dec]), power ([Fraction.Pow]), and binary
// arithmetic operations always produce a strictly positive denominator (d > 0)
// with the sign in the numerator. By contrast, [Fraction.StepNumerator] shifts
// the stored numerator by an integer delta while keeping the stored denominator
// (including its sign) unchanged without reduction (for example, stepping 6/12
// by +1 yields 7/12, whereas adding 1 yields 3/2).
//
// # Division with remainder, divisibility, and rounding
//
// Whole parts and remainders can be extracted in two directions:
//
//   - Toward zero: [Fraction.Split], [Fraction.SplitBig], [Fraction.QuoRem],
//     and [Fraction.QuoRemBig] truncate the quotient toward zero so that the
//     remainder has the sign of the dividend (for -13/12, Split yields -1 and
//     -1/12).
//   - Toward negative infinity (floor): [Fraction.SplitFloor],
//     [Fraction.SplitFloorBig], [Fraction.DivModFloor], and
//     [Fraction.DivModFloorBig] compute the floor quotient so that the
//     remainder is non-negative (for -1/12, SplitFloor yields -1 and 11/12).
//
// [Fraction.ScaleIntExact], [Fraction.IsMultipleOf], and
// [Fraction.InProgression] test exact integer divisibility and arithmetic
// progression membership without rounding.
//
// Rounding is never implicit. [Fraction.RoundToInt], [Fraction.RoundToBigInt],
// [Fraction.RoundToDenominator], [Fraction.RoundToBigDenominator], and
// [Fraction.FormatDecimal] round explicitly according to one of five
// [RoundingMode] rules: [RoundHalfAwayFromZero], [RoundHalfToEven],
// [RoundFloor], [RoundCeil], or [RoundTrunc].
//
// # Range, intermediate precision, and errors
//
// Every stored numerator and denominator has magnitude at most [MaxComponent]
// (2^255 - 1). Intermediate products, sums, greatest common divisors, and
// quotients are computed in extended exact integer precision without spurious
// intermediate overflow; [ErrOutOfRange] is returned only when the final
// requested numerator or denominator exceeds [MaxComponent].
//
// Every error returned by the package matches exactly one of the six exported
// sentinel errors via [errors.Is]: [ErrInvalid], [ErrZeroDenominator],
// [ErrDivisionByZero], [ErrNotRepresentable], [ErrOutOfRange], or
// [ErrInvalidText]. Every fallible constructor, function, and method has a
// corresponding Must* wrapper that returns the value on success and panics with
// the underlying error on failure.
//
// # Concurrency
//
// [Fraction] values are immutable and the package maintains no mutable global
// state. All functions and methods are safe for concurrent use by multiple
// goroutines without external synchronization (except concurrent calls to
// [Fraction.UnmarshalText] mutating the same *Fraction pointer).
package fraction
