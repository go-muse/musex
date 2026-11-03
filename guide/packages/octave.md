# Octaves

An octave is a place on an ordered axis of registers. The octave containing middle C can appear as ASPN 4, as 3 in Franco-Belgian numbering, and as *one-line* in an English catalog. All three identify the same octave. An octave value retains that identity whichever convention was used to find it.

## Position and labels

Moving one octave up always reaches the next register. In Franco-Belgian numbering, −1 and 1 are neighbors because the numbering skips 0. The distance between those registers is one octave. This is why the library measures shifts, differences, and ranges along the octave axis itself.

The axis has no musical endpoints. The Go representation supports every octave whose ASPN number fits in an `int`. Operations report overflow at its bounds. An uninitialized octave is invalid, distinct from the octave numbered 0 in ASPN.

A [numbering](octave/number.md) supplies a number for each octave. A [naming catalog](octave/naming.md) supplies words for the octaves it names. Both use ASPN numbers as a common reference, which allows an application to find an octave under one convention and display it under another.

First released in v0.6.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/octave) · [Go examples](../../octave/example_test.go) · [Roadmap](../../ROADMAP.md)
