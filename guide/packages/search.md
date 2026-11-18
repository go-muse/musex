# Search

The `search` package answers a deliberately narrow question: which catalog entries and tonics produce notes containing a supplied set, or producing exactly that set? Every answer is a **hypothesis**, not a claim about the actual musical context. The package does not rank answers or prefer major and minor over other structures.

## Scope and conditions

`search.BuiltIn()` uses the built-in template catalog. `search.New` snapshots a custom catalog, and `search.NewFromRepository` uses the catalog of a mode repository. A pointer to an empty catalog is a valid empty scope.

- `search.Containment`: every distinct input name or class occurs among the mode's notes.
- `search.FullMatch`: the input and the mode's notes are equal as sets.

Input order, duplicates, octaves, corrections in cents, and source metadata do not participate. Open modes and modes with a non-octave cycle use exactly their actual k degree notes; search does not continue an open mode or read the next cycle.

The following complete program finds exact-spelling containment hypotheses and prints the degrees on which each input name occurs. Names use the machine form of `spelling.Spelling` (`F+1` is F sharp).

```go
package main

import (
    "fmt"

    "github.com/go-muse/musex/cents"
    "github.com/go-muse/musex/note"
    "github.com/go-muse/musex/search"
    "github.com/go-muse/musex/spelling"
)

func main() {
    input := []note.Note{
        note.MustFromSpelling(spelling.D(), cents.Zero(), nil),
        note.MustFromSpelling(spelling.E(), cents.Zero(), nil),
        note.MustFromSpelling(spelling.FSharp(), cents.Zero(), nil),
        note.MustFromSpelling(spelling.A(), cents.Zero(), nil),
        note.MustFromSpelling(spelling.B(), cents.Zero(), nil),
    }
    matches := search.BuiltIn().MustBySpelling(input, search.Containment)
    for _, match := range matches {
        entry, _ := match.Mode().Origin()
        fmt.Println(entry.Code(), match.Tonic(), match.DegreesFor(spelling.FSharp()))
    }
}
```

For a full match, replace the condition and supply the complete set. With the seven natural names and a scope containing the seven diatonic entries, the answers include ionian/C, dorian/D, phrygian/E, and the other rotations; none is declared the correct musical interpretation:

```go
matches := searcher.MustBySpelling(input, search.FullMatch)
```

## Exact spelling and position

`BySpelling` never translates names enharmonically. C sharp and D flat therefore remain different. `ByPosition` instead interprets every input name in one explicit system:

```go
matches := search.BuiltIn().MustByPosition(system.Traditional(), input, search.Containment)
```

The explicit form never reads the sources of the input notes, even when they belong to different systems. `ByPositionFromContext` is the alternative when the common system should come from the notes themselves:

```go
matches := search.BuiltIn().MustByPositionFromContext(boundInput, search.Containment)
```

Each distinct source is read once through `Note.SameSource` and `Note.Settings`. Unbound notes, unreadable or unset settings, and unequal systems are distinct errors. Once the search finishes, its families retain the captured system rather than rereading a source.

## Tonic families and alternative spellings

A positional match stores the complete tonic family and a deterministic representative. In a system regular by classes, one family unites every letter with the tonic class. Otherwise, the family identity also contains a fixed letter. `Family().Letter()` distinguishes these cases.

The representative minimizes absolute alteration, then maximizes exact matches with input names, then follows `spelling.Compare`. It is a display convention, not a musical norm. To inspect all compatible names within an explicit limit:

```go
family := matches[0].Family()
for name := range family.MustSpellings(alteration.Sharp()) {
    fmt.Println(name)
}
```

The sequence is lazy and interruptible, and each traversal is independent. The limit restricts only this enumeration: search itself does not discard families for needing more signs. A small limit can produce an empty sequence even though the family exists.

## Degrees and grouping

`SpellingMatch.InputNames` and `PositionMatch.InputClasses` return normalized inputs. `DegreesFor` and `DegreesForClass` return every matching degree in increasing order, including several degrees with the same class. All returned slices are new.

Three representations are available for each result type:

```go
byCollection := search.MustGroupPositionByCollection(matches)
byTonic := search.MustGroupPositionByTonic(matches)
byStructure := search.MustGroupPositionByStructure(matches)
```

Equivalent functions exist for spelling matches. Grouping by collection compares sets of actual degree names or classes; grouping by tonic compares exact names or family identities; grouping by structure uses `mode.Equal`. Groups follow first occurrence, and their elements retain result order. Different catalog entries remain separate even when their structures are equal.

## Errors, ownership, and concurrency

Check errors with `errors.Is` against `search.ErrUnset`, `ErrEmptyInput`, `ErrInvalidArgument`, `ErrNoSystem`, `ErrUninterpretableName`, `ErrDifferentSystems`, and `ErrContext`. Dependency causes remain available in the same chain. For example, a name inadmissible in the selected system matches both `search.ErrUninterpretableName` and `system.ErrInadmissible`:

```go
matches, err := searcher.ByPosition(sys, input, search.Containment)
if errors.Is(err, search.ErrUninterpretableName) {
    // The input itself cannot be interpreted in sys; no partial result exists.
}
```

An invalid input rejects the whole operation. A particular candidate that is inadmissible or unrepresentable is skipped without cancelling other candidates. No compatible modes and a valid empty catalog both produce a successful, non-nil empty result.

Searchers, matches, and families are immutable and safe for concurrent reading. Returned slices are independent, and each returned mode can be edited without changing a match or another result. The package starts no background work and keeps no global cache.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/search) · [Roadmap](../../ROADMAP.md)
