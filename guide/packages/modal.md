# Modal positions

The `modal` package analyzes the numbered degrees of a [mode](mode.md) through
their specific [intervals](interval.md). Pass the mode directly to a query;
the analysis uses its written structure and needs no tonic note or musical
system. This package is unreleased.

## Relative positions

A relative position classifies the full interval from the first degree.
Major and augmented intervals are **high**, perfect intervals **neutral**,
and minor and diminished intervals **low**. Every representable integral
augmentation or diminution belongs to its quality family. Neutral and other
fractional interval qualities return `ErrUnsupportedQuality`; the neutral
modal category applies to perfect intervals.

```go
m := mode.MustFromTemplate(template.Dorian())
sixth := modal.MustRelative(m, 6)
// sixth.Category == modal.High
// sixth.Interval.Equal(interval.MajorSixth())

groups := modal.MustGroupRelative(m)
// groups.High:    [2 6]
// groups.Neutral: [1 4 5]
// groups.Low:     [3 7]
```

`Relative` returns the degree number, category, and complete `interval.Interval`.
Its generic size, direction, deviation, and multiplicity remain available:
compound intervals retain their full size. Categories express a grouping;
their enumeration values do not measure how high or low a degree is.

`RelativePositions` returns one position per degree, in numeric order.
`GroupRelative` returns all numbers in their category, also in numeric order,
with a non-nil empty slice when a category is absent. Both validate the whole
relative result. A single `Relative` query checks only the requested interval.
Open modes and closed modes with arbitrary cycles are supported.

## Absolute positions

Absolute analysis requires a closed mode whose cycle is exactly a perfect
octave, P8. It supports any valid number of degrees. From each degree it counts
one upward interval to every degree in the circle: its own P1 once and every
other degree once. A target earlier in the degree order is taken in the next
cycle. The source's next P8 is not an additional target. Different numbered
degrees remain separate even if their pitches coincide in a musical system.

`QualityCounts` reports diminished, minor, perfect, major, and augmented
counts. Each interval contributes once, regardless of integral multiplicity.
The sum equals the number of degrees. The score is:

```text
Minor + Diminished - Major - Augmented
```

Larger scores are brighter. Equal scores share a group even when their
quality counts differ. Ranks are dense: groups numbered 1, 2, 3, and so on,
without gaps. Degree numbers within a group are ascending.

```go
analysis := modal.MustAnalyzeAbsolute(m)
fourth := analysis.Positions[3] // Degree IV of the Dorian mode.
// fourth.Counts: one minor, three perfect, three major intervals.
// fourth.Score == -2; fourth.Rank == 5.

comparison := modal.MustCompareAbsolute(m, 4, 5)
// comparison == 1: IV is darker than V.
```

`AnalyzeAbsolute` computes positions in degree order and the ordered equal-score
groups together. `Absolute` selects one position, `AbsolutePositions` returns
the complete position list, and `Order` returns the groups. `CompareAbsolute`
returns -1 for a brighter first degree, 0 for equal scores, and 1 for a darker
first degree. Every form validates the complete absolute analysis, including
self-comparison. Reuse one `AbsoluteAnalysis` to read several results without
repeating the calculation. Absolute queries take O(k² + k log k) time and O(k)
additional memory for k degrees.

## Errors and ownership

All explicit numbers must identify actual degrees, from 1 through `m.Len()`.
Virtual degree numbers accepted by some `mode` operations are outside this API.
Check errors with `errors.Is` against `modal.ErrUnset`, `ErrInvalidDegree`,
`ErrUnsupportedStructure`, and `ErrUnsupportedQuality`. An unset-mode error
also retains `mode.ErrUnset`; other dependency causes remain in the error chain.

Validation checks the mode, then explicit numbers in argument order, then the
absolute structure when applicable. Absolute analysis checks representability
of the whole interval table before qualities. Quality errors follow starting
degree order and then increasing upward distance. Errors return zero structs,
nil lists, or comparison 0, together with the error, without a partial result.
Every error-returning function has a corresponding `Must` wrapper that calls
it and panics with its error on failure.

Functions leave the input mode unchanged and retain no reference to it after
the call. Returned slices belong to the caller and are independent between
calls and between groups, including within one analysis. Editing a mode later
does not change earlier results; a new query uses its current structure.
Assignment of a slice-containing result shares its elements under the normal
Go rules. Concurrent queries on an unchanged mode are safe; callers synchronize
mode mutations and writes to result slices.

[Package reference and Go examples](https://pkg.go.dev/github.com/go-muse/musex/modal) · [Roadmap](../../ROADMAP.md)
