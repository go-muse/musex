# Octave names

A naming catalog associates an [octave](../octave.md), identified by its ASPN number, with a word or phrase. For ASPN 4 the Russian name is *первая*, the English name is *one-line*, and the German name is *eingestrichene*. Changing catalogs changes the language used for the same register.

## Coverage and lookup

The Russian catalog covers ASPN −1 through 9. The names at −1 and 9 extend the Russian naming pattern in musex. The English and German catalogs cover 0 through 7. Looking up an octave outside a catalog's coverage gives an empty name. The octave still exists. A catalog can list its named octaves in ascending order of ASPN number; octaves without a name are not listed.

Within one catalog a name identifies exactly one octave. Looking up a name requires the spelling stored in that catalog, including case and whitespace. An unknown name gives an error. In a custom catalog, an empty name leaves an octave unnamed, and assigning the same name to two octaves is an error. Two catalogs are equal when they assign the same names to the same octaves.

First released in v0.3.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/octave/naming) · [Go examples](../../../octave/naming/example_test.go) · [Roadmap](../../../ROADMAP.md)
