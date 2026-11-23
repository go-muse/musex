# Mode

A mode describes numbered [degrees](degree.md) in terms of written intervals. An open mode ends at its last degree. A closed mode repeats at the interval of its cycle, which may be shorter than, equal to, or wider than a perfect octave.

Use a [template](template.md), a catalog entry, or a sequence of intervals or note spellings to build a mode. A constructor can also read steps in an explicit musical system. When a catalog entry supplies the structure, the mode retains its provenance and reports whether edits have changed that structure.

## Degrees and notes

Read the interval from the first degree, the interval between two degrees, or the adjacent intervals in a segment. A closed mode supports virtual degrees beyond its first cycle. The degree values supplied by the structure have no notes attached.

Supply a starting note to derive the notes of the actual degrees or of one degree at a signed cycle offset. The notes preserve its correction and source. When the starting note has an octave, either its settings source or an explicitly supplied system determines the result's octave.

## Build a scale from a tonic

`m.FromTonic(tonic)` returns a [scale](scale.md) containing one complete note
per degree. The tonic anchors the first degree of cycle zero. To choose a
starting degree and cycle, direction, and extent, use `FromTonicCount`,
`FromTonicTo`, or `FromTonicOctaves` with `mode.BuildOptions`.
Their `In` forms take an explicit musical system. The resulting scale owns
its finite list and can be edited independently of the mode. Circular
traversal of that list preserves the octaves of its stored notes.

## Find and explain an interval

`FindInterval` searches selected `IntervalScope.Families` and explicit
`Templates`. Family I uses the Ionian profile, II melodic minor, and III
Neapolitan major, including all seven rotations of each. These are structural
profiles, not a test for every seven-degree scale. For example A5 occurs in II
and III, d3 only in III, and A2 in none of the three.

```go
witnesses := mode.MustFindInterval(interval.AugmentedFifth(), mode.IntervalScope{
    Families: []mode.IntervalFamily{mode.FamilyII, mode.FamilyIII},
})
for _, w := range witnesses {
    fmt.Println(w.Family, w.Rotation, w.From, w.To, w.Interval)
}
```

Each witness includes its exact template and degree places. `DegreePlace`
uses an actual degree number and a signed cycle count; the source is in cycle
zero. Results include full cycles, zero spans, and compound or descending
intervals. `CycleWindow` is inclusive. A closed non-octave template needs an
explicit window; an open template uses only cycle zero. Duplicate selections
are ignored. Family witnesses precede explicit-template witnesses.

`MatchSpellings` checks each requested name independently and returns sorted
unique degree numbers. `MatchPitches` additionally checks register and returns
places ordered by cycle and number. Select `WrittenPitch` for exact spelling
or `PositionalPitch` for enharmonic position equality in an explicit system.
Both calls require a finite cycle window; its zero value selects cycle zero.
Both result lists must be nonempty to confirm both notes. Invalid inputs and
unrepresentable constructions return errors, not a negative membership result.

## Read a known degree role

`ResolveRole` compares an observed `interval.Chromatic` with a known pair of
degree places from a registered tonic. `ResolveRolePitch` measures a registered
observed pitch in an explicit system before making the same comparison.

```go
major := mode.MustFromTemplate(template.Ionian())
tonic := pitch.MustFromASPN(spelling.C(), 4)
observed := interval.MustNewChromatic(system.Traditional(), 3)
result := major.MustResolveRole(tonic, mode.DegreePlace{Number: 1},
    mode.DegreePlace{Number: 3}, observed, mode.AllowTargetAlteration)
// result.Status == mode.RoleMismatch: the expected III is E4 / M3 / +4.
// result.HasCandidate: E-flat 4 / m3 / +3, Alteration == -1.
```

`StrictRole` confirms only the exact expected full shift. A matching result has
`RoleMatch` and the expected reading as its candidate. `AllowTargetAlteration`
may explain a mismatch by changing only the target's alteration, keeping its
letter and written register. It still reports `RoleMismatch`; an impossible
alteration yields no candidate. Missing mode, tonic, observation, system, or
degree number returns `NeedContext` without error, retaining the requested
places. Explicit invalid parameters return errors even if context is missing.
All error-returning operations have corresponding Must wrappers.

## Distances between degree places in a system

`StepsBetween` and `StepsBetweenFrom` return by how many positions of a
musical system the note of one `DegreePlace` is above the note of another. A
place may be in any cycle of a closed mode, so the distance can cross cycle
boundaries; the sign follows the positions, not the degree numbers, so it can
be zero or negative. The octave of the first note does not matter.

```go
ionian := mode.MustFromTemplate(template.Ionian())
traditional := system.Traditional()
d := ionian.MustStepsBetweenFrom(traditional, spelling.MustParse("C"),
    mode.DegreePlace{Number: 7}, mode.DegreePlace{Number: 1, Cycle: 1})
up := ionian.MustStepsBetween(traditional, mode.DegreePlace{Number: 1}, mode.DegreePlace{Number: 3})
// d == 1, up == 4: no name is needed in a regular system.
```

`StepsBetweenFrom` names the first degree and works in any system, including
an irregular one, where a cycle that is not an octave can span a different
number of positions at different places. Only that name and the names of the
two places must be admissible. `StepsBetween` has no name and is allowed only in
a regular system; in an irregular one it returns `interval.ErrStartRequired`.
Only the final distance must be a whole number that fits in a `system.Step`:
fractional or huge intermediate numbers of the individual places do not cause
an error. For two places in cycle zero the result equals the difference of the
corresponding numbers of `Numbers` or `NumbersFrom` whenever those succeed
and the difference fits in an `int`; otherwise the distance is `ErrOutOfRange`.

Errors follow this order: `ErrUnset`; `ErrInvalidArgument` for a place number
outside 1..`Len()` or a nonzero cycle of an open mode, from before to; then
`interval.ErrStartRequired` or `system.ErrInadmissible`; `ErrOutOfRange` for an
unrepresentable place name in the named form; `interval.ErrFractionalSteps`;
and `ErrOutOfRange` for a distance outside `int`.

## Walk offsets in a system

The `Positions` value returned by `Numbers` and `NumbersFrom` describes where
each degree stands, in positions from the first degree. Its offset methods read
only `Degrees` and never fail. `DegreesAt` lists the degrees at one offset;
several degrees can share an offset and a free offset gives an empty list.
`OccupiedSteps` yields each occupied offset once with its degrees, and
`StepsIn(from, to)` yields every offset of an inclusive range, free ones with
an empty list. The `Backward` forms walk from the largest offset down; the
bounds of `StepsInBackward` are still given as lower, upper. The cycle boundary
is not crossed: degrees of other cycles are not searched.

```go
numbers := ionian.MustNumbers(traditional) // Degrees: 0 2 4 5 7 9 11
fmt.Println(numbers.DegreesAt(4), numbers.DegreesAt(1)) // [3] []
for offset, degrees := range numbers.StepsIn(0, 4) {
    fmt.Println(offset, degrees) // 0 [1], 1 [], 2 [2], 3 [], 4 [3]
}
```

Each iterator copies the degree numbers when the method is called, so later
changes of the value do not affect it. Iterators can be walked again, stopped
with `break`, and walked from several goroutines. Ranges are walked lazily,
without overflow at `math.MinInt` and `math.MaxInt`, and the gaps between
distant occupied offsets are not visited. Every returned list is a new slice
owned by the caller.

## Comparison and editing

Compare modes by structure, examine the intervals in which they differ, or locate an equal rotation. The package also describes diatonic properties, key signatures, and the numerical structure in a given system. A repository searches a catalog for exact, rotated, related, and numerical matches.

Edits include replacing, inserting, and removing degrees, rotating a closed mode, mirroring, and repeating its structure. Each edit validates its result before replacing the mode. Copies made by assignment can be edited independently, and iterators retain the structure they started from.

First released in v0.33.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/mode) · [Go examples](../../mode/examples_test.go) · [Roadmap](../../ROADMAP.md)
