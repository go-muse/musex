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

## Comparison and editing

Compare modes by structure, examine the intervals in which they differ, or locate an equal rotation. The package also describes diatonic properties, key signatures, and the numerical structure in a given system. A repository searches a catalog for exact, rotated, related, and numerical matches.

Edits include replacing, inserting, and removing degrees, rotating a closed mode, mirroring, and repeating its structure. Each edit validates its result before replacing the mode. Copies made by assignment can be edited independently, and iterators retain the structure they started from.

First released in v0.33.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/mode) · [Go examples](../../mode/examples_test.go) · [Roadmap](../../ROADMAP.md)
