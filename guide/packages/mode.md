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

## Build modes from steps

A mode can be described by the numbers of its adjacent intervals in a system:
how many positions each step spans. The numbers alone do not say which letter
each degree uses; a *letter plan* does. `LetterPlan.Degrees` holds the size of
the interval of every degree from the first one (1 for the first degree, equal
sizes for two degrees written with one letter), and `LetterPlan.Cycle` holds
the size of the cycle.

`SuggestLetterPlan` asks a spelling rule of package `system` (`Sharps`,
`Flats`, `ChromaticScale`, or `CustomSpelling`) for the plan of a one-octave
mode with a named first degree. The numbers include the closing one and must
add up to the positions of the system octave. The first degree keeps its name;
every other degree gets the name the rule chooses for its class, and its size
is the size of the written interval at its real place, so F♯ followed by F
eleven positions higher has size 8. `FromNumbersFromWithRule` builds that mode
directly. The plan is only a suggestion: edit it and pass it to
`FromNumbersFromWithPlan` to write the degrees differently.

```go
traditional := system.Traditional()
c := spelling.MustParse("C")
steps := []system.Step{3, 2, 2, 3, 2}

plan := mode.MustSuggestLetterPlan(traditional, steps, c, system.Flats())
// plan.Degrees == [1 3 4 5 7], plan.Cycle == 8: C E♭ F G B♭
flats := mode.MustFromNumbersFromWithRule(traditional, steps, c, system.Flats())
// flats equals mode.MustFromNumbersFromWithPlan(traditional, steps, c, plan)

plan.Degrees[1] = 2 // write the second degree as D♯ instead
edited := mode.MustFromNumbersFromWithPlan(traditional, steps, c, plan)
```

When the names of a rule are written out of order (a degree below the first
or the previous one, or above the first degree of the next cycle),
`SuggestLetterPlan` and `FromNumbersFromWithRule` return `ErrInvalidStructure`
instead of an incorrect plan; pass an explicit plan in that case. The checks
come in this order: `ErrUnset` (system, first name, rule), `ErrEmptyInput`,
`ErrInvalidArgument` (more than `MaxDegrees` numbers), `ErrInvalidStructure`
(a number below 1), `ErrSystemOctaveRequired` (the exact sum), then the
`system` errors for the first name and for the rule at degrees 2..k, and
finally `ErrInvalidStructure` for a plan out of order. With a single number the
plan is `[1]; 8` and the rule is not applied. `FromNumbers` and
`FromNumbersFrom` without a plan are unchanged: they still accept exactly seven
numbers and give each degree the next letter.

`OpenFromNumbers` and `OpenFromNumbersFrom` build an open mode from k−1 numbers
and the k sizes of its degrees; there is no cycle size. Sizes start at 1, do
not decrease, and may exceed 8. The form without a name works only in a
regular system; the named form checks the names of all degrees.

```go
open := mode.MustOpenFromNumbersFrom(traditional, []system.Step{2, 1}, c, []int{1, 2, 3})
// the open mode M2 m2: C D E♭
```

`Mode.LetterPlan` reads the plan of any mode without errors. A closed mode
reports its cycle size and `true`; an open mode reports `Cycle == 0` and
`false`; the unset mode gives an empty plan. Together with the numbers of the
mode the plan rebuilds it, as long as every number is at least 1:

```go
pentatonic := mode.MustFromCode("pentatonic-major")
read, closed := pentatonic.LetterPlan() // [1 2 3 5 6], 8, true
numbers := pentatonic.MustNumbers(traditional).Adjacent // 2 2 3 2 3
again := mode.MustFromNumbersWithPlan(traditional, numbers, read)
// again.Equal(pentatonic)
```

Every plan the package returns is a new slice owned by the caller; editing it
changes neither the mode nor later plans. Passed numbers, sizes, and rules are
neither changed nor retained.

To find catalog entries by numbers without building a mode, use
`Repository.IdentifyNumbers` or, with a name for the first degree,
`Repository.IdentifyNumbersFrom`. The kind is explicit (`mode.Closed` or
`mode.Open`), matches are exact from the first degree without rotations, and
the entries come in catalog order. A query without matches returns an empty
slice, not an error.

```go
entries := mode.BuiltInRepository().MustIdentifyNumbers(mode.Closed, traditional,
    []system.Step{3, 2, 1, 1, 3, 2})
// blues, blues-b5: different templates with the same numbers
```

## Comparison and editing

Compare modes by structure, examine the intervals in which they differ, or locate an equal rotation. The package also describes diatonic properties, key signatures, and the numerical structure in a given system. A repository searches a catalog for exact, rotated, related, and numerical matches, either by a mode or directly by the numbers of adjacent intervals (`IdentifyNumbers`, `IdentifyNumbersFrom`).

Edits include replacing, inserting, and removing degrees, rotating a closed mode, mirroring, and repeating its structure. Each edit validates its result before replacing the mode. Copies made by assignment can be edited independently, and iterators retain the structure they started from.

First released in v0.33.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/mode) · [Go examples](../../mode/examples_test.go) · [Roadmap](../../ROADMAP.md)
