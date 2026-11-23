# musex

<p align="center">
  <img src="musex-logo-readme.png" alt="musex logo" width="720">
</p>

[![Latest release](https://custom-icon-badges.demolab.com/github/v/release/go-muse/musex?style=flat&logo=tag&label=release)](https://github.com/go-muse/musex/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/go-muse/musex?style=flat&logo=go)](https://github.com/go-muse/musex/blob/main/go.mod)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-muse/musex.svg)](https://pkg.go.dev/github.com/go-muse/musex)
[![Roadmap](https://go-muse.github.io/musex/roadmap.svg)](ROADMAP.md)
<br>
[![CI](https://img.shields.io/github/actions/workflow/status/go-muse/musex/ci.yml?branch=main&style=flat&logo=github&label=CI)](https://github.com/go-muse/musex/actions/workflows/ci.yml?query=branch%3Amain)
[![coverage](https://go-muse.github.io/musex/coverage.svg)](https://go-muse.github.io/musex/)
[![lint](https://img.shields.io/github/check-runs/go-muse/musex/main?nameFilter=lint&style=flat&logo=go&label=lint)](https://github.com/go-muse/musex/actions/workflows/ci.yml?query=branch%3Amain)
[![govulncheck](https://img.shields.io/github/check-runs/go-muse/musex/main?nameFilter=vulnerabilities&style=flat&logo=go&label=govulncheck)](https://github.com/go-muse/musex/actions/workflows/ci.yml?query=branch%3Amain)
[![MIT License](https://custom-icon-badges.demolab.com/github/license/go-muse/musex?style=flat&logo=law&label=license)](LICENSE)

`musex` is a Go library for representing and working with musical concepts in software. It is intended to provide building blocks for music editors, analyzers, educational tools, and music generators.

> **Early development.** The library is not yet ready for production use. The public API may change before v1.0.

The [Roadmap and package guide](ROADMAP.md) lists the planned packages and links each implemented package to its chapter. For Go types and functions, see the [API reference](https://pkg.go.dev/github.com/go-muse/musex).

## Concept

`musex` keeps what is written apart from how it sounds. A note is a written record: a spelling such as C♯, usually placed in an octave, with a correction in cents that is zero unless given. It holds no musical system and no tuning. It reads them from a set of settings, for example the one an instrument owns, so retuning the instrument changes the sound of the notes bound to it without rewriting them. A chord holds notes that sound together. An interval is measured between two notes as they are written and is a value of its own: a direction, a size, and a quality. A degree pairs an ordinal number with an optional note and counts numbers in a caller-defined circle without prescribing a mode.

```mermaid
flowchart LR
    subgraph NOTE["Note"]
        subgraph PITCH["Written pitch"]
            SP("Spelling") ~~~ OC("Octave")
        end
        CO("Correction#nbsp;in#nbsp;cents")
    end

    subgraph INSTRUMENT["Instrument"]
        subgraph SETTINGS["Settings"]
            TU("Tuning")
            DE("Designations")
            SY("Musical system")
        end
    end

    NOTE -. "reads" .-> SETTINGS

    subgraph CHORD["Chord"]
        C1(("Note")) ~~~ C2(("Note")) ~~~ C3(("Note")) ~~~ C4["• • •"]
    end

    I1(("Note")) <-->|"Interval"| I2(("Note"))

    classDef note fill:#4f46e5,stroke:#4338ca,color:#ffffff
    classDef setting fill:#0f766e,stroke:#115e59,color:#ffffff
    classDef more fill:none,stroke:none,color:#8b949e
    class SP,OC,CO,C1,C2,C3,I1,I2 note
    class TU,DE,SY setting
    class C4 more
    style NOTE fill:#4f46e514,stroke:#4f46e5,stroke-width:1.5px,rx:12,ry:12
    style PITCH fill:#4f46e514,stroke:#4f46e5,stroke-width:1px,rx:10,ry:10
    style INSTRUMENT fill:#0f766e14,stroke:#0f766e,stroke-width:1.5px,rx:12,ry:12
    style SETTINGS fill:#0f766e14,stroke:#0f766e,stroke-width:1px,rx:10,ry:10
    style CHORD fill:#8b949e14,stroke:#8b949e,stroke-width:1.5px,rx:12,ry:12
```

## Definitions

Definitions may differ from conventional ones.

**Spelling** is the name of a note: a **letter**, one of C D E F G A B, with an **alteration**, the change an accidental makes, counted in sharps. C♯ and D♭ are different spellings.

**Octave** is a register on an ordered axis of registers that has no musical ends. The octave that contains middle C is the same octave whether it is numbered 4 in ASPN, numbered 3 in Franco-Belgian numbering, or called one-line in English.

**Musical system** is the set of rules that connects spellings to positions. It fixes how many positions an octave has and gives each spelling it admits a **pitch class**, a position that repeats in every octave; a spelling can be admissible in one system and not in another. In the traditional system of twelve positions C♯ and D♭ share a pitch class; in the system of nineteen positions they do not. No system is assumed by default.

**Written pitch** is a spelling in a positional octave, such as C♯ in octave 4. The octave is the octave of positions that holds the pitch, not the number printed next to the letter: in the traditional system the record B♯3 is B♯ in octave 4. C♯ and D♭ in the same octave are different written pitches.

**Tuning** assigns a frequency to every written pitch. It combines four parts: a musical system; a **temperament**, which fixes the height of every position of the octave in cents; the pitch class the temperament starts from; and a reference such as A4 at 440 Hz.

**Settings** is one unchangeable value that joins a musical system, its tuning, and its **designations**: a textual **notation** for names and octave marks, an octave numbering, and a catalog of octave names. Replacing a part gives a new set of settings.

**Instrument** is the owner of a set of settings. It gives its current set to the notes that read it and replaces the set as a whole, after checking that the new set can stand in for the old one. It also has a name and an optional reference range of pitches.

**Note** is a written record, a spelling or a written pitch, with a **correction** in cents and an optional source of settings. The correction moves the sound and leaves the record as it is. A note has no duration and stores no settings: a bound note reads them from its source at each use, and an unbound note is given them for each operation.

**[Interval](guide/packages/interval.md)** is the distance between two spellings, written pitches, or notes as they are written: a direction, a size counted in letters, and a quality such as major or augmented. It is neither a number of semitones nor a frequency ratio: B♯ to C is a diminished second. The **positional size** of an interval is the number of steps it spans in a musical system. `interval.Chromatic` measures and reuses a signed shift in that system; `Generic` and quality palettes recover or enumerate written intervals. Explicit Russian and English dictionaries live in `interval/naming`.

**Mode template** is an ordered record of the intervals between neighboring places of a mode. A closed template repeats after an arbitrary cycle; an open template ends after its last place. Templates can be rotated, compared, serialized, and stored in immutable catalogs with codes and English and Russian names. A template is not a mode and has no tonic, degrees, or musical system.

**[Mode](guide/packages/mode.md)** is a sequence of numbered degrees described by written intervals. It can be open or repeat after an arbitrary cycle. A caller supplies a starting note to build notes or a musical system to compare positional sizes. Its `FromTonic` methods build an independent scale, with optional count, pitch, or octave bounds. Interval witnesses explain family membership, while spelling and pitch membership checks identify actual degree places. Role readings distinguish confirmation, mismatch, and missing context. In an explicit system it measures signed distances in positions between degree places of any cycles and lists or walks the degrees at each offset. It builds modes from the numbers of their adjacent intervals with a letter plan suggested by a spelling rule or given explicitly, builds open modes from numbers and degree sizes, reads the letter plan of any mode, and finds catalog entries directly by numbers.

**[Scale](guide/packages/scale.md)** keeps a finite ordered list of complete notes. It supports circular traversal and snapshot iterators, exact comparisons and set operations, melody shifts and reflections, editing, transposition, enharmonic spellings, and key signatures. Scales can be built from spellings or text tokens, written back as tokens, projected to an ordered pitch-class sequence, and built as playable chromatic runs between two pitches with an explicit spelling rule.

**[Search](guide/packages/search.md)** finds mode-and-tonic hypotheses whose actual degree notes contain or exactly match a set of input notes. It supports exact-spelling and positional matching in a chosen or context-derived system, tonic-family spellings, and stable grouping without ranking musical likelihood.

**[Modal positions](guide/packages/modal.md)** describe degrees through their specific intervals. Relative analysis returns high, neutral, or low categories together with the full interval from the first degree. Absolute analysis of a closed perfect-octave mode counts five quality families from each degree and orders equal-score groups from brightest to darkest. Integral augmentations and diminutions count once regardless of multiplicity; fractional qualities are unsupported. Relations extend the relative view to any source degree, and two modes can be compared degree by degree with flags for changed intervals, categories, scores, and ranks.

**[Tendency and resolution](problemsolving/tendency/README.md)** answers support, degree-label, direction, target, consonance, and two-voice resolution questions under explicit built-in or finite user rules. Coverage and explanations preserve missing context and uncertainty; filters and comparisons retain the original claims.

**[Scale choice](guide/packages/problemsolving/scalechoice.md)** supports a composer's choice of pitch collection: inspect compatible modal readings, common and missing notes, possible continuations, and the consequences of one trial replacement in a set or a phrase. The caller chooses the scope, comparison, condition, and optional center.

**[Voice leading](#analyze-voice-leading)** compares existing melodic lines: written and chromatic intervals, kinds of two-part motion, perfect parallels, consecutive perfect families, direct arrivals, crossings, and overlaps. A separate, explicit assessment checks one Open Music Theory first-species rule.

**[Hindemith analysis](guide/packages/problemsolving/hindemith.md)** reads dyads, voiced chords, and progressions through Hindemith's harmonic theory: Series II positions and interval roots, the chord groups I-VI with subgroups, bass and root with the intra-chord intervals behind them, and harmonic fluctuation. The model and its twelve-tone equal-temperament assumption are documented; the result is an analysis under stated rules, not a heard center.

**[Sonance](guide/packages/problemsolving/sonance.md)** scores the interval inventory of a pitch-class set or of actual pitch pairs under an explicit weight profile: Huron's published aggregate dyadic consonance weights, or caller-supplied weights that can distinguish a fourth from a fifth. Every contribution is exposed.

**Chord** is a group of notes that sound together. It keeps the notes in their order of record, repeats included, and has no duration and no harmonic name.

**Tempo map** is an immutable timeline of tempo changes and explicit fermata holds at positions or measure addresses. It distinguishes main and current tempo, supports jumps, gradual changes, a tempo, Tempo I, and metric modulation, and converts seconds to and from positions and measures, with one explicitly selected rounding for the complete result, to decimal places or to any step.

**[Ticks](guide/packages/format/ticks.md)** are the whole-number units of exchange formats: MIDI ticks, MusicXML divisions, and MEI pulses per quarter. The `format/ticks` package converts durations, positions, and measure addresses to them and back exactly, chooses a resolution, and translates tempo and tempo maps, time signatures and their changes, SMPTE time, and MIDI clocks; any rounding is explicit and happens once.

**[Track](guide/packages/track.md)** is a part assigned to one live instrument: parallel voices written in a finite measure grid, atomic edits, tied sounds, explicit layout and rebar proposals, optional seconds, and conversion to and from melodic lines.

**[Score](guide/packages/score.md)** is an immutable ensemble of tracks with one common start and positional tempo map. It supports record and sounding slices, shared seconds, full-map extraction, strict meter changes, and common barlines across independent grids.

**[Event](guide/packages/event.md)** is one written element of a voice. It combines a note, chord, or rest with a duration and any ties, fermata, arpeggio, or grace sign.

**Degree** is an ordinal number with an optional note. It is not tied to a tonic or mode: its number can be formatted and parsed, and circular operations count positions with an explicit cycle count.

**MIDI note number** is an integer position from 0 through 127 in the ordinary MIDI map. C4 is 60 and A4 is 69. The number depends on the positional octave, not the register digit: B♯3 and C4 both give 60, while C♭4 and B3 both give 59. Reverse conversion requires an explicit spelling rule; a correction in cents is passed separately and does not change the number.

## Use Cases

- **Convert between the written model and MIDI note numbers.** Create any of the 128 numbers, inspect its positional octave and class, enumerate the numbers of a positional octave, shift it without overflow, and convert written pitches and notes in a traditional C-layout system. Choose the return spelling explicitly with sharps, flats, or a custom rule, attach zero or a specific correction, and handle missing inputs, unsupported layouts, and out-of-range results as distinct errors.
- **Track positions in a caller-defined cycle.** Create numbered degrees with or without notes, move to neighboring positions or by an arbitrary shift, measure directed and shortest distances, and convert between continuous and normalized numbers while keeping cycle counts explicit.
- **Find modes from a set of notes.** Search an immutable built-in or custom catalog by exact spelling or pitch class, choose containment or full match, inspect every matching degree and compatible tonic spelling, and group hypotheses without treating them as musical recognition or ranking.
- **Analyze modal positions.** Pass a mode directly to `modal.Relative` or `modal.GroupRelative` to inspect its degrees. Use `modal.AnalyzeAbsolute` for quality counts, scores, and dense ranks in one complete analysis; equal scores share a rank. Relative analysis also supports open modes and arbitrary cycles. Use `modal.RelationsFrom`, `modal.RelationTo`, or `modal.RelationTable` for the relations from any degree, and `modal.CompareRelativePositions` or `modal.ComparePositions` to see what changes between two modes.

- **Choose a collection and try a change.** Use `problemsolving/scalechoice` to compare collection groups and readings, filter them by added notes, or branch into a replacement trial with observed-set differences and local written/chromatic motions. Analyses snapshot source contexts and preserve the original material.
- **Write and analyze a melodic line.** Use `voice` to build a monophonic line of sounds and rests with a name and roles, read positions and the element at a moment without measures, get seconds from a tempo map, melodic intervals, the range, the ambitus, and the tessitura against a singing voice or an instrument, and build transpositions, doublings, exact imitations, inversions, and retrogrades.
- **Inspect how parts move together.** Use `problemsolving/voiceleading` to analyze two or more existing lines on independent pairwise grids. Keep the original notes and element boundaries, choose interval-family filters explicitly, and supply nominal order when looking for crossings or overlaps.
- **Classify chords and progressions by Hindemith.** Use `problemsolving/hindemith` to read a dyad's Series II position and root, a voiced chord's group, subgroup, bass, and root with its intra-chord evidence, and a progression's harmonic fluctuation. A chord is classified by the lowest note of each pitch class, so higher doublings and the listing order never change the result; groups V and VI report no root in isolation.
- **Score consonance of sets and voicings.** Use `problemsolving/sonance` to score a pitch-class set's interval-class vector or the actual pitch pairs of a voicing under Huron's published weights or a caller-supplied profile, and inspect which intervals drive the value.
- **Write rhythms and import notated material.** Divide written durations into tuplet parts, build tuplet groups with contents and sounding spans, and find the exact rhythmic readings of a sequence of sounding durations under allowed ratios. Read and write ABC and LilyPond tuplet markers, group spellings into a keyboard palette by pitch class, and build scales from names or tokens with a chromatic spelling rule.

### Build a voice event

Given two `note.Note` values `c4` and `e4`, mark the second note of a chord for a tie into the next event. Copy the marks to add a fermata to a new event.

```go
content := event.MustChordContent(chord.MustNew(c4, e4))
written := event.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 1))
entry := event.MustNew(content, written, event.Marks{Ties: []int{1}})
marks := entry.MustMarks()
marks.Fermata = true
held := entry.MustWithMarks(marks) // entry retains its original marks.
fmt.Println(held.MustMarks().Fermata) // true
```

To represent a three-measure rest, store the measure count. The caller uses its measure grid to find the corresponding length.

```go
rest := event.MustNew(event.RestContent(), event.MustMeasureDuration(3), event.Marks{})
_, err := rest.Length()
if errors.Is(err, event.ErrNeedsMeasureContext) {
    fmt.Println(rest.MustDuration().MustMeasureCount()) // 3
}
```

### Build a melodic line

Given `note.Note` values `c4`, `d4`, and `e4` bound to standard settings, build a subject, answer it at the fifth after a half rest, and measure its tessitura against the soprano range.

```go
q := voice.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 0))
h := voice.MustWrittenDuration(duration.MustNewWritten(duration.Half, 0))
subject := voice.MustNew(
    voice.MustNewSound(c4, q),
    voice.MustNewSound(d4, q),
    voice.MustNewSound(e4, h),
).MustWithName("Subject").MustWithPlacement(voice.Soprano)
for _, m := range subject.MustIntervals() {
    fmt.Println(m.Interval) // M2, then M2
}
answer := subject.MustImitateAfter(interval.PerfectFifth(), h) // a half rest, then G4, A4, B4
fmt.Println(answer.MustThematic(), answer.MustLength())        // imitating 3/2
fmt.Println(answer.MustAt(duration.MustNewPosition(3, 4)))       // 2 true: A4 sounds at 3/4
t := subject.MustTessitura(voice.PlacementReference(voice.Soprano))
fmt.Println(t.Number, t.Kind) // 5/38 low
```

### Analyze voice leading

Given two set `voice.Line` values `bass` and `upper`, declare their nominal
order and choose a system explicitly. Input slots identify the parts; their
names may repeat. A later entry is represented by an initial rest.

```go
a, err := voiceleading.Analyze([]*voice.Line{bass, upper}, voiceleading.Options{
    System: system.Traditional(),
    Order:  voiceleading.LowToHigh,
})
if err != nil {
    return err
}
pair := voiceleading.Pair{First: 0, Second: 1}
v, ok := a.MustVerticalAt(pair, duration.Origin())
if ok && v.Sounding {
    fmt.Println(v.Specific, v.Specific.Size(), v.Chromatic)
}
parallels := a.MustFindParallels(voiceleading.ParallelOptions{
    Perfect: voiceleading.PerfectBoth,
})
for _, t := range parallels {
    fmt.Println(t.From.Start, t.To.Start, t.First, t.Second)
}
assessment := a.MustCheckOMTFirstSpecies(pair)
fmt.Println(assessment.Rule, assessment.Applicable, assessment.Issues, assessment.Violations)
```

Each pair ends at its shorter line and uses only the starts of its own
elements. Rests remain visible and break classified motion. `VerticalAt`
accepts a position inside a segment; `TransitionAt` requires two adjacent
increasing segment starts and otherwise returns `ErrTransition`.

Specific intervals preserve written direction, quality, and full generic size.
Chromatic distances count signed system positions; cents, tuning frequencies,
and performance timing do not change these measurements. `Options{}` is
valid: it leaves nominal order unspecified and reads a common system once per
participating sound element. An explicit system permits unbound notes and
avoids reading their sources. Subsequent source changes do not alter the
measured snapshot.

`FindMotion(MotionDirect)` includes parallel motion. Strict `FindParallels`
requires parallel written movement and equal full perfect sizes: normally P5
or P8, with explicit `Compound` and `Unisons` extensions. `FindConsecutive`
instead matches perfect families regardless of movement or octave extension.
`FindDirect` requires an explicit `AnyOrigin` or `DifferentFamily` choice.
Crossings inspect all pairs; overlaps inspect adjacent nominal parts. Both
require `LowToHigh` and compare positional heights.

The explicit OMT assessment checks consecutive perfect families only. Its
note-against-note scope requires equal original sounding-element starts and
ends. If that scope fails, issues explain the affected verticals and violations
are empty. It does not apply a general stylistic judgment or check every rule
of first-species counterpoint.

Analyses are immutable, safe to copy and read concurrently, and compared with
`Equal`. Every returned slice, including nested crossing and assessment
slices, is independent. For example, changing `a.MustVerticals(pair)[0]`
changes only that returned copy. Returned notes keep their original source
binding; returned systems are borrowed immutable values. Successful empty
collections are non-nil. Invalid inputs return zero results and errors
inspectable with `errors.Is` or contextual `PartError` and `MeasurementError`.
Every fallible operation has a `Must` wrapper that panics with its error.

Chord spacing, chord distribution, leap checking, and chord connections remain
planned work beyond this first version.
