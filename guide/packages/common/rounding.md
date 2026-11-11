# Rounding

Rounding places a value on a chosen grid. A grid is an ordered set of nodes: it may be uniform, such as integers, multiples of 480 ticks, thirds, or decimal places, or non-uniform, such as the steps of a temperament. The package supplies the rounding rule; the caller chooses the grid and, when necessary, supplies its musical meaning.

There is no default rounding rule. Every operation makes the choice explicit so that different consumers can use different conventions without a hidden global setting.

## Neighbors and ties

A value that already coincides with a grid node stays unchanged. Otherwise, it lies strictly between two adjacent nodes: the lower neighbor and the upper neighbor. Directed rules always choose a particular side. Nearest rules compare the two distances and choose the closer node; equal distances are a tie, which needs its own convention.

For example, on the integer grid, 1.5 has neighbors 1 and 2 and is exactly halfway between them. On a grid of 480 ticks, 1200 has neighbors 960 and 1440 and is likewise a tie. On a non-uniform grid, the distances need not be equal just because the value appears to be “between” two nodes: the caller compares the actual distances in its own exact units.

## The nine rules

| Rule | −2.5 | −1.5 | −0.5 | 0.5 | 1.5 | 2.5 | Tie behavior |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Down, toward negative infinity | −3 | −2 | −1 | 0 | 1 | 2 | Lower neighbor |
| Up, toward positive infinity | −2 | −1 | 0 | 1 | 2 | 3 | Upper neighbor |
| Toward zero | −2 | −1 | 0 | 0 | 1 | 2 | Neighbor nearer zero |
| Away from zero | −3 | −2 | −1 | 1 | 2 | 3 | Neighbor farther from zero |
| Half up, toward positive infinity | −2 | −1 | 0 | 1 | 2 | 3 | Upper neighbor |
| Half down, toward negative infinity | −3 | −2 | −1 | 0 | 1 | 2 | Lower neighbor |
| Half away from zero | −3 | −2 | −1 | 1 | 2 | 3 | Neighbor farther from zero |
| Half toward zero | −2 | −1 | 0 | 0 | 1 | 2 | Neighbor nearer zero |
| Half to even | −2 | −2 | 0 | 0 | 2 | 2 | Neighbor with the even grid index |

The first four rules are directed. The other five choose the nearest node and differ only in how they resolve a tie. “Half up” here always means toward positive infinity, and “up” also means toward positive infinity. Some other libraries use those names for movement away from zero; the explicit distinction between up, away from zero, half up, and half away from zero avoids that ambiguity.

A zero produced by rounding is an unsigned zero: neither a floating-point negative zero nor a decimal string with a minus sign is produced.

## Zero and parity belong to the grid

“Toward zero” and “away from zero” refer to the zero of the grid, not automatically to the sign of a node value. “Half to even” refers to the parity of a node's index, not to whether its numeric value happens to be an even number.

On a uniform grid, these meanings are natural: nodes have indices …, −1, 0, 1, …, node zero is the origin, and even indices select the even nodes. For example, thirds have indices 0, 1, 2, 3 for the values 0, 1/3, 2/3, 1. A tie between 1/3 and 2/3 therefore resolves to 2/3 under half-to-even because its index 2 is even.

On a non-uniform grid, the caller defines the indexing and origin. A tuner, for example, can number the steps of a temperament relative to a reference step. It knows which index is zero, whether the lower index is even, and the exact distances from the input frequency to the adjacent steps. The rounding rule then chooses the lower or upper step without interpreting cents, frequencies, or temperament data itself.

## Uniform grids and exact choice

Uniform grids support whole integers, integer quotients and remainders, multiples of a positive step, exact rational values, fractions with a chosen denominator, and decimal places. The neighboring nodes and their distances are determined exactly before a rule is applied; intermediate floating-point subtraction is not used to decide whether a value is tied.

Integer division returns both the rounded quotient and the remainder consistent with it. Thus 7 divided by 3 gives quotient 2 and remainder 1 when rounded down, but quotient 3 and remainder −2 when rounded up. A remainder can therefore be negative.

Exact rational arithmetic supports values of arbitrary size and grids such as quarters. For example, 5/8 lies exactly halfway between 2/4 and 3/4: half-to-even selects 2/4, while half-up selects 3/4.

A denominator can also describe a grid directly. One half is exactly 1.5 thirds, so half-down selects 1/3, whereas half-up and half-to-even select 2/3. This avoids approximating one third by a decimal or binary floating-point step.

## Decimal formatting and places

Decimal output uses ordinary notation without an exponent, a leading plus sign, or unnecessary leading zeroes. A positive number of places fixes exactly that many digits after the decimal point: 1/8 becomes `0.12` under half-to-even and `0.13` under half-away-from-zero when two places are requested. Zero is written without a minus sign, including forms such as `0.00`.

Negative place counts round to tens, hundreds, and larger powers of ten. For example, 1250 to two negative places becomes `1200` under half-to-even and `1300` under half-up. The supported range is from −400 through 400 places, covering every decimal position relevant to a finite 64-bit floating-point value with additional room for exact rational calculations.

## Reading floating-point values

For magnitudes below 2⁵³, a floating-point input is interpreted through its shortest decimal representation—the decimal value a user normally sees. This makes decimal rounding meet decimal expectations: 0.145 to two places becomes `0.15` under half-up, even though formatting the underlying binary value directly with two digits can produce `0.14`. Likewise, 0.15 on a 0.1 grid is treated as the exact decimal tie between 0.1 and 0.2.

Starting at magnitude 2⁵³, every such floating-point value is an integer and is interpreted by its exact binary value. This preserves all its integer digits instead of replacing them with the shorter display form. For example, the value commonly displayed as `1e23` formats to zero places as `99999999999999991611392`, its exact value.

Calculations that need the exact binary value below 2⁵³ can use the exact rational operations instead. In every case, a numeric floating-point result is the nearest representable value to the exact rounded result, with ties resolved to an even mantissa. Reapplying a step-grid operation to that represented result need not be idempotent when the requested grid node itself is not representable.

## Configuration names

The canonical configuration values are `down`, `up`, `toward-zero`, `away-from-zero`, `half-up`, `half-down`, `half-away-from-zero`, `half-toward-zero`, and `half-even`. They are case-sensitive and have no aliases or surrounding whitespace, giving each rule one stable spelling for settings and text-based data exchange.

First released in v0.16.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/common/rounding) · [Go examples](../../../common/rounding/example_test.go) · [Roadmap](../../../ROADMAP.md)
