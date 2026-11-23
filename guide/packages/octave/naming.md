# Octave names

A naming catalog associates an [octave](../octave.md), identified by its ASPN number, with a word or phrase. For ASPN 4 the Russian name is *первая*, the English name is *one-line*, and the German name is *eingestrichene*. Changing catalogs changes the language used for the same register.

## Coverage and lookup

The Russian catalog covers ASPN −1 through 9. The names at −1 and 9 extend the Russian naming pattern in musex. The English and German catalogs cover 0 through 7. Looking up an octave outside a catalog's coverage gives an empty name. The octave still exists. A catalog can list its named octaves in ascending order of ASPN number; octaves without a name are not listed.

Within one catalog a name identifies exactly one octave. Looking up a name requires the spelling stored in that catalog, including case and whitespace. An unknown name gives an error. In a custom catalog, an empty name leaves an octave unnamed, and assigning the same name to two octaves is an error. Two catalogs are equal when they assign the same names to the same octaves and accept the same aliases.

## Aliases

A catalog can also accept additional input names for its named octaves. `WithAliases` returns a new catalog with the aliases added; the original is unchanged, successive calls accumulate, and an alias that already resolves to the same octave is a no-op, so alias lists from several sources can be merged. For example, the Russian catalog can accept *первая октава* and *FirstOctave* alongside the canonical *первая* for ASPN 4; all three lookups return 4, while the name of the octave stays *первая*. `Aliases` lists the additional names of an octave in byte order. The built-in catalogs have no aliases.

An alias must resolve to exactly one octave: an alias equal to another octave's canonical name or alias is an error, and so is an alias for an octave without a canonical name. Empty alias strings are ignored. Aliases extend the lookup only: the name of an octave and the enumeration of named octaves keep returning the canonical names.

First released in v0.3.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/octave/naming) · [Go examples](../../../octave/naming/example_test.go) · [Roadmap](../../../ROADMAP.md)
