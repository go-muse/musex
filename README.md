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

**Interval** is the distance between two spellings, written pitches, or notes as they are written: a direction, a size counted in letters, and a quality such as major or augmented. It is neither a number of semitones nor a frequency ratio: B♯ to C is a diminished second. The **positional size** of an interval is the number of steps it spans in a musical system.

**Chord** is a group of notes that sound together. It keeps the notes in their order of record, repeats included, and has no duration and no harmonic name.

**Degree** is an ordinal number with an optional note. It is not tied to a tonic or mode: its number can be formatted and parsed, and circular operations count positions with an explicit cycle count.

## Use Cases

- **Track positions in a caller-defined cycle.** Create numbered degrees with or without notes, move to neighboring positions or by an arbitrary shift, measure directed and shortest distances, and convert between continuous and normalized numbers while keeping cycle counts explicit.
