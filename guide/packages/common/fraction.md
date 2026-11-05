# Fractions

An ordinary fraction represents an exact signed rational number while preserving its supplied numerator and denominator. The package belongs to the numeric foundation of the library and stays independent of musical units: higher-level packages supply musical meaning such as chromatic steps, tuplets, durations, cent scales, or measure positions.

## Why musical calculations preserve the denominator

In music theory and notation, two equal rational values often carry different structural meaning depending on their denominator:

- In a 12-step cycle or a measure subdivided into twelfths of a whole note, the position `6/12` names step 6 of the 12-unit grid, whereas the reduced value `1/2` states only that the position lies halfway through the period.
- Two consecutive triplet eighth notes have duration `(1/8) × (2/3) = 1/12` each; their sum on the twelfth-note grid is `2/12`, which reduces to `1/6` of a whole note when only the net span matters.
- Tuplet ratios such as `3/2` and `6/4` are numerically equal in the rational numbers, yet distinguish a triplet from a sextuplet subdivision.

Treating every rational number solely as an automatically reduced fraction loses the active subdivision grid. Conversely, keeping only raw pairs without rational equivalence makes it impossible to compare positions across different subdivisions. Both views are supported together.

## Stored pairs, canonical reduction, and equality

Every valid fraction stores its numerator and non-zero denominator within the symmetric 255-bit magnitude range `±(2²⁵⁵ − 1)`:

- **Preserved numerator and denominator.** Constructing a fraction without reduction (or parsing an ordinary fraction such as `"6/12"` or `"1/-2"`, or stepping the numerator by an integer delta) keeps the supplied numerator and denominator as given, including any common factor and the original sign of the denominator (for example, `6/12`, `0/3`, `1/-2`, or `-2/-4`).
- **Canonical reduced form.** Every valid fraction also has a unique irreducible form with a strictly positive denominator, relatively prime components, the sign placed in the numerator, and `0/1` for numeric zero (`6/12` reduces to `1/2`, `1/-2` and `-2/4` reduce to `-1/2`, and `0/3` and `0/-5` reduce to `0/1`).
- **Two equality and ordering relations.** Numeric comparison checks mathematical equality and order in the rational numbers (`6/12`, `1/2`, and `-2/-4` are equal; `0/1`, `0/3`, and `0/-5` are equal; `1/-2` and `-1/2` are equal). Exact pair comparison checks whether both the stored numerator and the stored denominator match (`6/12` and `1/2` are distinct pairs, as are `1/-2` and `-1/2`, whereas `0/3` and `-0/3` are the same pair).
- **Undefined zero value.** The uninitialized zero value `0/0` represents an invalid, unset fraction rather than numeric zero (`0/1` or `0/3`).

## Denominator policies in arithmetic

Unary, power, and binary arithmetic operations always produce a positive denominator with the sign carried by the numerator. By contrast, stepping the numerator by an integer delta shifts only the stored numerator and preserves the stored denominator (including its sign) without reduction (`6/12` stepped by `+1` becomes `7/12`, whereas adding `1` gives `3/2`).

When adding, subtracting, multiplying, or dividing two fractions, seven explicit denominator-selection policies are available:

| Policy | Resulting denominator | Typical musical use |
| --- | --- | --- |
| **Reduced** | Canonical irreducible denominator (`d > 0`, `0/1` for zero) | Net duration totals, interval ratios, canonical comparisons |
| **Unreduced least common multiple (LCM)** | `lcm(\|d₁\|, \|d₂\|)` for addition and subtraction without post-reduction (`1/6 + 1/6 = 2/6`, `1/4 − 1/4 = 0/4`); `\|d₁\|·\|d₂\|` for multiplication; `\|d₁·n₂\|` for division | Combining two grids into their coarsest common subdivision while keeping that grid denominator |
| **Unreduced product** | Direct product `\|d₁\|·\|d₂\|` for addition, subtraction, and multiplication without reduction (`1/6 + 1/6 = 12/36`); `\|d₁·n₂\|` for division | Preserving raw compound subdivision structure |
| **Same unit** | Shared denominator `D` when `\|d₁\| = \|d₂\| = D` without reduction | Advancing, subtracting, or scaling along a single fixed grid |
| **Larger unit** | Coarser denominator `D = min(\|d₁\|, \|d₂\|)` of the two operands when the exact result fits `1/D` | Expressing a combined value on the coarser of the two operand grids when no remainder is left |
| **Smaller unit** | Finer denominator `D = max(\|d₁\|, \|d₂\|)` of the two operands when the exact result fits `1/D` | Expressing a combined value on the finer of the two operand grids when no remainder is left |
| **Explicit target denominator** | Caller-supplied positive denominator `D > 0` when the exact result fits `1/D` | Projecting an exact calculation onto a chosen measure or beat subdivision |

When an exact unit policy (same unit, larger unit, smaller unit, or explicit target denominator) cannot represent the true rational result as an integer multiple of `1/D`, the operation reports that the value is not representable rather than silently rounding or changing the grid. Intermediate calculations use extended exact integer precision so that products and cross-terms do not overflow prematurely when the final result fits within `±(2²⁵⁵ − 1)`.

## Whole-part extraction and division with remainder

Separating a fraction into an integer whole part and a fractional remainder — or dividing a dividend by a non-zero step — supports two rounding conventions:

- **Truncation toward zero.** The integer quotient rounds toward zero and a non-zero remainder keeps the sign of the dividend. For example, `-13/12` splits into whole part `-1` and fractional remainder `-1/12`, matching signed mixed-number decomposition.
- **Floor (Euclidean) division.** For a positive step, the integer quotient rounds toward negative infinity (`floor`) so that the remainder is always non-negative and strictly smaller than the step (`0 <= remainder < step`). For example, at position `-1/12` relative to a cycle of length `1`, floor division gives cycle index `-1` and in-cycle offset `11/12`.

The remainder's representation depends on the operation, not on the rounding convention:

- **Whole-part extraction.** The fractional remainder preserves the magnitude of the stored denominator without reduction, including when it is zero: `6/3 → 2 + 0/3` and `14/12 → 1 + 2/12` under either convention. For a negative value, `-13/12 → -1 + (-1/12)` toward zero and `-13/12 → -2 + 11/12` toward negative infinity.
- **Division by a supplied step.** The remainder is always in canonical reduced form, with a positive denominator and `0/1` for zero. For example, dividing `6/3` by step `1` gives quotient `2` and remainder `0/1` under either convention; dividing `-1/12` by step `1/4` toward negative infinity gives quotient `-1` and remainder `1/6`.

## Divisibility and arithmetic progressions

Musical timelines and pitch grids frequently test whether a value lands on an exact subdivision without remainder:

- **Exact integer scaling and multiples.** Scaling a fraction by an integer factor that must divide its reduced denominator, or testing whether a fraction is an exact integer multiple of a non-zero step (for example, whether `6/12` is a multiple of `1/6`), succeeds only when the division leaves zero remainder.
- **Membership in an arithmetic progression.** Given a starting value `start` and a positive step `step > 0`, progression testing checks whether a fraction equals `start + k × step` for **any integer** step count `k` (negative, zero, or positive) and returns `k` when it does. For example, from start `1/4` with step `1/2`, the value `5/4` lies at `k = 2` and the value `-1/4` lies at `k = -1`.

## Explicit rounding and floating-point conversion

Exact rational arithmetic never rounds implicitly. When a calculation must be quantized to an integer or to a positive denominator `D > 0` (preserving `D` without reduction, such as rounding `1/3` to denominator `12` as `4/12` or `3/8` to denominator `4` as `2/4`), the caller selects one of five explicit rounding rules:

- **Half away from zero** (ties round away from zero: `+1/2 → +1`, `−1/2 → −1`)
- **Half to even** (banker's rounding: ties round to the nearest even integer or even numerator)
- **Floor** (toward `−∞`)
- **Ceiling** (toward `+∞`)
- **Truncation** (toward zero)

Explicit conversion to a 64-bit IEEE 754 floating-point number is also available when interfacing with external synthesis or rendering pipelines.

## Text formatting and parsing

Fractions format to and parse from plain ASCII text:

- **Stored pair notation** (`"6/12"`, `"1/-2"`, `"0/3"`, and `"0/0"` for the undefined zero value) and **canonical reduced notation** (`"1/2"`, `"-1/2"`, `"0/1"`).
- **Compact notation**, which formats the reduced fraction and omits `/1` for whole numbers (`"2"`, `"-3"`, `"1/2"`).
- **Mixed-number notation**, which separates a non-zero integer part from a proper reduced fractional part (`"1 3/4"`, `"-1 3/4"`, `"3/4"`, `"2"`, `"0"`).
- **Fixed-precision decimal formatting**, which formats a fraction with a chosen number of digits after the decimal point (`0` through `100`), rounding either half away from zero by default or according to any of the five explicit rounding rules (and never producing `"-0"` or `"-0.00"` when the rounded value is zero).
- **Parsing**, which accepts four exact forms without surrounding whitespace: signed integers (`"3"`, `"-2"`), ordinary fractions (`"6/12"`, `"1/-2"`), mixed fractions (`"1 3/4"`, `"-1 3/4"`), and finite decimal numbers (`"1.75"`, `"-0.125"`, `"2."`), converting decimal input to its reduced rational form.

First released in v0.8.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/common/fraction) · [Go examples](../../../common/fraction/example_test.go) · [Roadmap](../../../ROADMAP.md)
