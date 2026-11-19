# Scale choice

`problemsolving/scalechoice` connects a composer's notes with compatible pitch
collections, modal readings, possible continuations, and one trial change.
The caller chooses the search scope, condition, comparison, and optionally a
center. The package does not infer a heard tonic or rank artistic choices.

## Choose a scope and comparison

`Classical()` selects Ionian, Dorian, Phrygian, Lydian, Mixolydian, Aeolian,
and Locrian, in that order. `Pentatonic()` selects major and minor pentatonic
entries. These are explicit search profiles. Custom `template.Catalog` values
and unions made with `catalog.With(other.Entries()...)` work too. Every selected
entry must describe a representable closed mode with a perfect-octave cycle.
An empty catalog is valid and never substitutes a built-in scope.

Use `search.Containment` for collections containing the input set, or
`search.FullMatch` for equal sets. Select `BySpelling` for exact names or
`ByPosition` for classes in one musical system. Repeated notes and octaves do
not change the searched set. A center filters readings by exact spelling or
by class according to that same comparison; it never replaces the search
family's display representative.

## Find a next note

With the imports `cents`, `note`, `search`, `spelling`, and
`problemsolving/scalechoice` from `github.com/go-muse/musex`:

```go
n := func(name string) note.Note {
    return note.MustFromSpelling(spelling.MustParse(name), cents.Zero(), nil)
}
material := []note.Note{n("C"), n("D"), n("E"), n("G"), n("A")}
opts := scalechoice.Options{
    Condition: search.Containment,
    Comparison: scalechoice.BySpelling,
}
choices := scalechoice.MustAnalyzeSet(material, scalechoice.Classical(), opts)
withF := choices.MustAfterAdding(spelling.F())
withFB := choices.MustAfterAdding(spelling.F(), spelling.B())
// Three initial groups; withF keeps two, withFB keeps one.
```

`Groups()` returns each complete collection, its `Missing` notes, and its
`Readings`. `Common()` intersects complete collections, including notes already
played. `Continuations()` lists every found but unplayed tone and partitions
current groups into `Kept` and `Dropped` indices. It includes notes that preserve
all groups. `AfterAdding` filters those groups without starting another search;
empty arguments preserve all, and duplicates do not change the result.
At full match, continuations are empty. A valid query with no matches remains a
valid analysis with empty groups.

A written `Tone` has one exact name and an unset class. A positional `Tone` has
a valid class and distinct sorted spellings seen in the corresponding result,
not every theoretical enharmonic spelling. Tones sort by spelling or class;
groups and readings retain first-search-result order.

## Compare modal readings

```go
naturals := []note.Note{n("C"), n("D"), n("E"), n("F"), n("G"), n("A"), n("B")}
opts.Condition = search.FullMatch
all := scalechoice.MustAnalyzeSet(naturals, scalechoice.Classical(), opts)
// One group, seven readings of the same collection.
opts.Center = spelling.D()
dorian := scalechoice.MustAnalyzeSet(naturals, scalechoice.Classical(), opts)
r := dorian.Groups()[0].Readings[0]
fDegrees := r.MustDegreesOf(spelling.F()) // III: []int{3}
```

A reading provides the actual catalog `Entry` with all its aliases, the tonic
or family representative, and `Degrees()` in modal order without a closing
repeat. `DegreesOf` accepts any valid new name too, returning all matching
degree numbers in increasing order. Multiple degrees of one positional class
remain visible. Aliases do not create separate readings; distinct catalog
entries retain their provenance. The same collection does not prove the same
heard tonal center.

## Try one change in a phrase

Use `AnalyzePhrase` to preserve adjacency. For example, build the seven natural
notes and a second F in register 4, with `cents`, `note`, `spelling`, and `system`:

```go
var phrase []note.Note
for _, name := range []string{"C", "D", "E", "F", "G", "A", "B", "F"} {
    phrase = append(phrase, note.MustFromASPN(spelling.MustParse(name), 4, cents.Zero(), nil))
}
opts := scalechoice.Options{
    Condition: search.FullMatch,
    Comparison: scalechoice.BySpelling,
    Center: spelling.C(),
    System: system.Traditional(),
}
line := scalechoice.MustAnalyzePhrase(phrase, scalechoice.Classical(), opts)
whole := line.MustReplaceAll(spelling.F(), spelling.FSharp())
// The C-Ionian reading is removed and C-Lydian is added.
one := line.MustReplaceAt(3,
    note.MustFromASPN(spelling.FSharp(), 4, cents.Zero(), nil))
// The second F remains: eight names fit no selected classical collection.
// one.After is still valid and its changed places and local motions are available.
```

`ReplaceAll` selects exact old spellings in every register, even for positional
comparison. It preserves cents and, when interpretable, each written octave
through the system's octave boundary. `ReplaceAt` accepts one exact replacement
record with or without a register and with its supplied cents. Its source is
ignored; the old place's captured context is inherited. Neither operation
changes the original analysis or rereads any settings source. Each successful
`After` supports further independent trials using the same captured contexts.

`Matched` identifies selected original places; `Changed` identifies records
whose spelling, register, or cents actually changed. An absent old name, a
matched no-op, and an invalid phrase index are distinct cases. `Tones` compares
the observed sets even when no modes fit. `Groups` compares complete found
collections: removed indices address the original groups, kept and added
indices address `After.Groups()`. Reading identity is catalog code and exact
tonic, or code and positional-family identity; changing a family's display
spelling alone does not remove it. Kept positional tones unite names seen on
both sides.

For a phrase, `Neighbors` lists each affected adjacent pair once, in phrase
order. Each side has independent `Named` and signed `Chromatic` results and
independent errors. Inspect `NamedError` and `ChromaticError` separately. In
E4–F4–G4, changing F to F-sharp changes shifts [+1,+2] to [+2,+1]. Enharmonic
replacement can preserve chromatic shifts while changing written intervals.
Zero shifts and octave movement remain visible. A set has no neighboring edges.

## Context, ownership, and failures

An explicit valid system applies to the entire analysis, including trial
changes and measurements, without consulting note sources. Otherwise each
distinct source is read once. Positional search requires a common equal system
and admissible names, even with an empty catalog. Written search remains
available when a source is missing, invalid, or in another system: context or
its failure is retained per place. A usable adjacent pair can still be measured
when an unrelated place lacks context.

Returned notes preserve spelling, system octave values, and cents, and are
unbound copies. Returned slices, nested name lists, and systems are independent
of internal snapshots. Catalogs and entries retain their existing immutable
value contracts. Copies, repeated getters, and independent trial branches can
be read concurrently; do not mutate caller inputs during a call.

When written replacement cannot transfer an original register because context
is unknown or the new spelling is inadmissible there, it retains the new name
without an octave. An explicit replacement octave in `ReplaceAt` is retained as
a value even without context; measurement still reports that context failure.
Computational range errors are returned rather than hidden. Missing edge
registers report `interval.ErrNoOctave` before context is consulted.

Use `errors.Is` with `ErrUnset`, `ErrEmptyInput`, `ErrInvalidArgument`,
`ErrNoSystem`, `ErrDifferentSystems`, `ErrUninterpretableName`, or `ErrContext`
and preserved dependency causes. A nonnil pointer to an unset system is an
error in both comparisons. Zero `Analysis` and `Reading` values provide empty
getters; error-returning operations reject them. All error-returning operations
have ordinary Must wrappers that panic with the same error.

[Search](../search.md) · [Intervals](../interval.md) ·
[API reference](https://pkg.go.dev/github.com/go-muse/musex/problemsolving/scalechoice) ·
[Roadmap](../../../ROADMAP.md)
