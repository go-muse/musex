// Package cents provides pitch distances in cents: exact values created from
// numbers, frequency ratios, frequencies, compositions, and texts, with
// arithmetic, comparison, division into octaves, the nearest step of an equal
// division of the octave, quotients of values, frequencies, named small
// intervals and units, text forms, and the pitch values of Scala files.
//
// The cent is the logarithmic unit of a pitch distance: the distance from a
// frequency f1 to a frequency f2 is c = 1200·log2(f2/f1) cents, so an octave
// (the ratio 2:1) is 1200 cents and an equal-tempered semitone is 100 cents.
// The definition does not depend on a tuning system or a temperament.
//
// A [Cents] value is one signed quantity. The same type serves as the
// correction of a note, as a value T(m) of a temperament, and as a distance
// between pitches. The constraints of these roles, such as 0 <= T(m) < 1200,
// increasing tables, an optional correction, or a correction that does not
// take part in the name and the step of a note, are checked by the packages
// that use them, not by this package.
//
// The package defines the following operations:
//
//   - [Zero], [Cent], [Semitone], and [Octave] return basic values;
//     [FromInt], [FromFraction], [Step], [FromRatio], and [FromComposition]
//     create exact values; [FromFloat], [FromRatioFloat], and [Between]
//     create values from float64 numbers on a grid.
//   - [Cents.Add], [Cents.Sub], [Cents.Neg], [Cents.Abs], [Cents.Mul],
//     [Cents.Div], [Cents.Scale], and [Cents.ScaleRat] compute exactly;
//     [Cents.RoundTo] rounds explicitly; [Cents.SplitOctaves] divides a value
//     into octaves; [Cents.NearestStep] finds the nearest step of an n-EDO.
//   - [Cents.Equal], [Cents.Compare], [Cents.Sign], and [Cents.Near] compare
//     values exactly.
//   - [Cents.Float64], [Cents.Fraction], [Cents.Ratio], and
//     [Cents.Composition] read a value; [Cents.Quo] returns the [Quotient] of
//     two values, such as an amount of a unit; [Cents.Multiplier] and
//     [Cents.ApplyTo] work with frequencies.
//   - [MinNormalFrequency] and [MaxFrequency] bound frequencies;
//     [IsValidFrequency], [IsNormalFrequency], [ValidFrequency], and
//     [NormalFrequency] check them.
//   - [Schisma] to [LargeLimma] and [NamedValues] give 19 named small
//     intervals; [MIDITuningUnit] to [Meride] and [NamedUnits] give 8 named
//     units.
//   - [Cents.String], [Cents.Text], [Cents.TextAs], [Cents.Decimal],
//     [Cents.Kind], [Cents.Format], [Cents.MarshalText],
//     [Cents.AppendText], [Cents.UnmarshalText], and [Parse] write and read
//     texts; [Cents.Scala], [ParseScala], and [ParseScalaApprox] write and
//     read the pitch values of Scala files.
//
// Functions and methods whose names start with Must are wrappers for
// literals and values known to be valid: they return the same results as the
// operation without Must, or panic with its error, so that calls can be
// chained:
//
//	meantone := cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4))
//
// For data from users, use the operations that return errors.
//
// # Exact values
//
// A value is exact: it is a rational number r of cents plus rational
// multiples of the sizes J(p) = 1200·log2(p) of the primes 3 to 31,
//
//	v = r + e3·J(3) + e5·J(5) + ... + e31·J(31).
//
// The prime 2 is part of r, because J(2) = 1200. Every value has exactly one
// such record, so equality and order are exact, without tolerances: a == b if
// and only if a.Equal(b), and a value can be a map key. Exact values include
// integers and decimal fractions of cents, any fraction of cents, the steps
// k·1200/n of any equal division of the octave, the cents of ratios of
// integers whose prime factors are at most 31, and fractions of such ratios,
// such as a quarter of the syntonic comma. No operation on values rounds,
// except where rounding is explicit.
//
// The limit 31 is a prime limit, not an odd limit: a ratio is exact if all
// its prime factors are at most 31, whatever its odd factors are. The ratio
// 37:32 is outside the limit and gives [ErrPrimeLimit] on an exact path; an
// approximate path, such as [FromRatioFloat] or [ParseScalaApprox], gives its
// cents on the grid.
//
// [Cents.Composition] returns the record, and [FromComposition] creates a
// value from it: the rational part r in cents and the exponents of the
// primes. It is related to the monzo (prime exponent vector) of a ratio, with
// two differences: the exponent of 2 is given in cents as part of r, and the
// exponents may be fractions. The syntonic comma 81:80, the monzo
// [-4 4 -1⟩, has the rational part -4800 cents and the exponents 3 → 4 and
// 5 → -1.
//
// # Approximate values
//
// [FromFloat], [FromRatioFloat], and [Between] take float64 numbers and
// return values on a grid of 1/GridDenominator = 10^-6 cent. Each input is
// understood as its shortest decimal representation, the one written by
// strconv.FormatFloat(x, 'g', -1, 64), as an exact number: 7.1 is 71/10,
// and 5e-7 is exactly half a step of the grid, although its binary value is
// slightly less. The exact result is then rounded to the nearest multiple of
// 10^-6 cent, with an exact half rounded away from zero, symmetrically for
// both signs. A number of cents with at most 6 digits after the point is
// therefore exact. Logarithms are rounded correctly, without math.Log2, so
// the results are the same on all platforms. After creation, an approximate
// value is an ordinary exact value: FromRatioFloat(1.5) is 701.955001 cents,
// not equal to FromRatio(3, 2).
//
// Other rounding is explicit: [Cents.RoundTo] rounds to a multiple of 1/d
// cent, and the decimal outputs round to a chosen number of digits, with an
// exact half rounded away from zero. [Cents.Float64] returns the nearest
// float64.
//
// # Range, limits and errors
//
// Every value is within the range -MaxMagnitude to MaxMagnitude, that is,
// ±9,000,000 cents (7500 octaves), bounds included; the range is checked
// exactly, also for irrational values. It is a technical limit, not a
// musical one. A result outside the range is an error, never truncated,
// wrapped, or saturated.
//
// The record of every value is within the record limits: every reduced
// denominator of r and of the exponents is at most 10^30, and every exponent
// is at most 10^24 in magnitude. The limits apply to the reduced record of a
// result, not to intermediate values or inputs. They include, with a
// margin: integers; decimals with up to 6 digits after the point; the steps
// k·1200/n for n <= 10^6; the cents of ratios of 64-bit integers with primes
// up to 31; the named values; amounts p/q of the named units and of
// "1/n of an octave" with q, n <= 10^6; the sum and the difference of two
// such values; any int multiple of such a value within the range; its
// fraction 1/n for |n| <= 10^6; and the deviation of [Cents.NearestStep] of
// such a value for n <= 10^6. A cent divided by 10^6 five times is still
// within the limits; a sixth division gives [ErrTooComplex].
//
// Texts have length limits, checked on the text before any arithmetic: a run
// of digits has at most MaxDigits = 100 digits, leading zeros included, and
// a product of powers has at most MaxFactors = 64 factors, each written
// factor counted. A text accepted by [Parse] or [Cents.UnmarshalText] is
// therefore at most 19,583 bytes long (about 19.6 KB). For [ParseScala] and
// [ParseScalaApprox] the limits apply only to the value of the line, at most
// 202 bytes; the leading spaces and the rest of the line are not limited and
// are passed over in linear time.
//
// Errors are recognized with [errors.Is], and each error matches exactly one
// of [ErrOutOfRange], [ErrTooComplex], [ErrPrimeLimit],
// [ErrInvalidArgument], and [ErrInvalidText]. When several conditions apply,
// operations on values, numbers, and compositions report the first in the
// order invalid argument, prime limit, record limits, range. Texts report the
// first in the order invalid text, length limits, prime limit, record
// limits, range: the length limits come before the primes, so that long
// numbers are never factored. On error, all other results are zero values,
// and the arguments are unchanged. Integer arguments may be any int,
// including math.MinInt and math.MaxInt: only the exact result decides an
// error. Only the Must functions panic.
//
// [MinNormalFrequency] and [MaxFrequency] bound frequencies and frequency
// multipliers for all musex packages; they are defined only here.
// [IsValidFrequency] defines a valid frequency argument as finite, positive,
// and at most MaxFrequency, so a positive subnormal number is valid.
// [IsNormalFrequency] defines a normal frequency as one between
// MinNormalFrequency and MaxFrequency, inclusive.
//
// The inputs of [Cents.ApplyTo] and [Between] and the ratio of
// [FromRatioFloat] follow the valid-frequency rule. The results of
// [Cents.Multiplier] and [Cents.ApplyTo] follow the normal-frequency rule:
// the exact result is rounded once to a 53-bit mantissa with an unlimited
// exponent, and [ErrOutOfRange] is returned if that value is below
// MinNormalFrequency or above MaxFrequency. This includes a subnormal result
// of a valid subnormal input frequency.
//
// [ValidFrequency] returns [ErrInvalidArgument] for an invalid input.
// [NormalFrequency] is a check of a computed result and returns
// [ErrOutOfRange] for every violation, NaN and infinities included. For input
// values use ValidFrequency; for computed values use NormalFrequency.
//
// # Named values
//
// The package names 19 small intervals, in increasing order: the schisma
// 32805:32768, the septimal kleisma 225:224, the kleisma 15625:15552,
// Ptolemy's comma 100:99, the diaschisma 2048:2025, the syntonic comma
// 81:80, the Holdrian comma (1/53 of an octave), the Pythagorean comma
// 531441:524288, the septimal comma 64:63, the lesser diesis 128:125, the
// septimal quarter tone 36:35, the undecimal comma 33:32, the greater diesis
// 648:625, the tridecimal comma 27:26, the minor chroma 25:24, the limma
// 256:243, the major chroma 135:128, the apotome 2187:2048, and the large
// limma 27:25. Their exact relations hold as equalities, for example the
// schisma is the Pythagorean comma minus the syntonic comma.
//
// The names have synonyms and ambiguities:
//
//   - "comma" without a qualifier usually means the syntonic comma, and
//     "diesis" the lesser diesis; "lesser" and "minor", "greater" and
//     "major" are synonyms;
//   - the schisma 32805:32768 is not the "Grad" or "schisma" of
//     Werckmeister (≈ 1.955 cents, a twelfth of the Pythagorean comma);
//   - 36:35 is the septimal quarter tone; it is also called the "septimal
//     diesis", a name used for 49:48 (≈ 35.697 cents) as well, which is not
//     a named value;
//   - 100:99 is Ptolemy's comma; Wikipedia calls it the "small undecimal
//     comma", while the Huygens-Fokker list of intervals gives that name to
//     99:98;
//   - "Mercator's comma" means either 3^53:2^84 (≈ 3.615 cents) or a value
//     near 1/53 of an octave (≈ 22.6 cents, close to the Holdrian comma), so
//     it is not a named value. The ratio 3^53:2^84 is 53 fifths 3:2 minus 31
//     octaves, MustFromRatio(3, 2).MustMul(53).MustSub(Octave().MustMul(31)),
//     or Parse("19383245667680019896796723:19342813113834066795298816").
//
// # Units
//
// A unit is a value: its size in cents. The amount of a unit in a value is
// [Cents.Quo] with the unit, and the value of an amount is [Cents.Scale] or
// [Cents.ScaleRat] of the unit: 700 cents are 1750/3 millioctaves, and
// Millioctave().Scale(1750, 3) is 700 cents. The unit "1/n of an octave" is
// Step(1, n); the cent is [Cent] and the octave [Octave]. The named units
// are [MIDITuningUnit] (1/196608 of an octave), [Jot] (1/30103), [Millioctave]
// (1/1000), [Centitone] (1/600), [Savart], [Eptameride] (1/301), [Meride]
// (1/43), and [Semitone] (1/12).
//
// The word "savart" has four meanings in the literature. The package uses
// the decimal savart, 1/1000 of a decade (the ratio 10:1), as defined by
// Wikipedia's article "Savart": about 301.03 savarts in an octave. The
// Huygens-Fokker list of measures defines the savart as 1/301 of an octave,
// which is the eptaméride of Sauveur, [Eptameride]; Wikipedia's article
// "Savart" describes the eptaméride differently, but the package follows the
// Huygens-Fokker list. A modified savart of 1/300 of an octave is Step(1,
// 300), and 100/30103 of an octave, given by the Huygens-Fokker list as
// another savart, is 100 jots.
//
// # Text
//
// A value has up to four exact text forms ([Kind]): decimal cents
// (KindDecimal, "700", "7.1", "-13.7") for a rational value with a
// terminating decimal; a fraction of cents with a unit (KindFraction,
// "1200/19 cents") for any rational value; a frequency ratio (KindRatio,
// "3:2", without a sign or a unit) for the cents of a ratio of integers; and
// a product of powers of primes (KindPower, "5^(1/4)",
// "2^(-1199/1200)*3") for any value. Every value has exactly one canonical
// form, [Cents.Kind]: decimal cents if it is rational with a terminating
// decimal, otherwise a fraction if it is rational, otherwise a ratio if its
// terms have at most MaxDigits digits, otherwise a product of powers. A
// ratio with a longer term, such as 3^5000:2^5000, has no ratio form, so that
// every written text can be read back.
//
// [Cents.String] writes the canonical form in ASCII, without "+" and without
// the optional unit; it is also the output of %v, of [Cents.MarshalText],
// and of [Cents.AppendText], so JSON and other encoders of the standard
// library write a value, also as a map key, as a string of its canonical
// form, such as "3:2". [Cents.Text] and [Cents.TextAs] apply [Options] to
// decimal cents and fractions: Typographic writes the minus "−" (U+2212) and
// the unit "¢"; WithUnit adds the unit to decimal cents; WithPlus writes "+"
// before a positive number. Without Typographic the text is ASCII, and the
// unit is "cent" if the written number without its sign is exactly "1" and
// "cents" otherwise, so 1 cent, -1 cent, but 1.000 cents.
//
// [Parse] reads all the forms strictly, with the sign "+", "-", or "−", the
// units "¢", "cent", and "cents" in any case, and at most one space or
// no-break space before the unit. It rejects spaces around a number or a
// colon, commas, exponents such as "1e3", "408." and ".5", signs and units
// of ratios and products, and a bare fraction "3/2", which is ambiguous: the
// error suggests "3:2" for a ratio and "3/2 ¢" for cents. A product of powers
// has at most MaxFactors factors; every written form has at most 11, so it
// fits. [Cents.Decimal] and %f write rounded decimals for reading; the
// decimal of an irrational value is not exact, so exchange values in their
// exact forms.
//
// # Scala
//
// [Cents.Scala] writes the pitch value of a line of a Scala tuning file: the
// ratio "n/d" if the value is the cents of a ratio with terms at most 2^31-1,
// otherwise cents with at least one digit after the point. The Scala
// specification requires programs that read the files to support numbers of
// at least 2^31-1, so such ratios can be read by any of them; it does not
// limit the numbers of the files, and [ParseScala] reads ratios of any size
// up to the length limits. Irrational values are written as rounded cents, a
// limitation of the format. [ParseScala] reads the value of a line and
// returns the rest of the line, which a file reader needs for the names of
// the notes; a line must be passed without its line break, so a "\r" of a
// Windows file is an error. [ParseScalaApprox] gives the cents of a ratio
// with a prime greater than 31 on the grid instead of an error. Reading and
// writing whole .scl files is outside the package.
//
// # Composition, quotients and steps
//
// [FromComposition] takes the rational part and powers of the bases 2 and
// the primes 3 to 31; the exponent of 2 gives 1200 cents per unit, as in a
// monzo. A base that is not a prime gives [ErrInvalidArgument], and a prime
// greater than 31 [ErrPrimeLimit], whatever the exponent; then the record
// limits and the range are checked.
//
// [Cents.Quo] returns the number c/d as a [Quotient]: exact
// ([Quotient.Exact]) when it is rational, which is the case when the records
// of c and d are proportional, and correctly rounded to any number of digits
// otherwise ([Quotient.Decimal], [Quotient.Float64]).
//
// [Cents.NearestStep] returns the step k of an n-EDO nearest to a value, with
// an exact half rounded away from zero, and the deviation, with
// |deviation| <= 600/n cents. It returns [ErrOutOfRange] if k does not fit
// in an int, and [ErrTooComplex] if the reduced denominator of the rational
// part of the deviation exceeds 10^30. That denominator divides d·n, where d
// is the denominator of the rational part of the value: the error does not
// occur for the values listed in the record limits with n <= 10^6, but with
// n near math.MaxInt it can occur already for values whose denominator is of
// the order of 10^11 to 10^12.
//
// # Performance
//
// Creating a value allocates memory for its record; ==, [Cents.Equal], and
// copying do not allocate. The frequency checks take constant time and do not
// allocate unless they return an error. Operations decide the sign, the order,
// the range, and the rounding of irrational values from intervals of growing
// precision,
// so comparing nearly equal values and writing many digits take longer than
// usual. The cost of parsing is bounded by MaxDigits and MaxFactors, so
// [Parse], [Cents.UnmarshalText], [ParseScala], and [ParseScalaApprox] are
// suitable for untrusted text, for example when decoding JSON; ParseScala
// passes over the leading spaces and the rest of a line in linear time.
// [FromComposition] and [Cents.ScaleRat] have no input limits: their cost
// grows with the length of the [big.Int] and [big.Rat] numbers and the number
// of powers (quadratically for many exponents with large coprime
// denominators), and these inputs come from the calling program.
// [Cents.AppendText] writes into the spare capacity of a large enough buffer
// without a new array for the result; the number of allocations is not
// specified.
//
// # Concurrency
//
// Values are immutable, and all functions and methods are safe for
// concurrent use without synchronization, also on the same values and with
// the named values. Data of the caller that a method writes is synchronized
// by the caller, as for any Go variable:
//
//   - [Cents.UnmarshalText] writes its receiver;
//   - [Cents.AppendText] can write into the spare capacity of its buffer b,
//     the bytes of the array of b after len(b). If one call can write bytes
//     that another goroutine reads or writes, the caller synchronizes such
//     calls or gives each goroutine its own buffer. Buffers on one array with
//     disjoint capacities, such as a[:0:8] and a[8:8:16], are independent,
//     and a buffer without spare capacity (len(b) == cap(b), including nil)
//     is not written.
package cents
