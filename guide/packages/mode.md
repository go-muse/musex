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

## Comparison and editing

Compare modes by structure, examine the intervals in which they differ, or locate an equal rotation. The package also describes diatonic properties, key signatures, and the numerical structure in a given system. A repository searches a catalog for exact, rotated, related, and numerical matches.

Edits include replacing, inserting, and removing degrees, rotating a closed mode, mirroring, and repeating its structure. Each edit validates its result before replacing the mode. Copies made by assignment can be edited independently, and iterators retain the structure they started from.

First released in v0.33.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/mode) · [Go examples](../../mode/examples_test.go) · [Roadmap](../../ROADMAP.md)
