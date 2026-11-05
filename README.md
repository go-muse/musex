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

A written name combines a letter with an alteration. An octave identifies a register across naming traditions. A pitch class identifies a position that repeats in each octave. In the traditional twelve-position system with C at 0, C♯ and D♭ are different spellings of class 1.

```mermaid
flowchart LR
    L["Letter · C"] -->|supplies the letter| S["Spelling · C♯"]
    A["Alteration · +1"] -->|raises by one chromatic unit| S
    S -.->|mapped by a musical system| PC["Pitch class · 1 mod 12"]
    O["Octave containing middle C"] -->|ASPN numbering| N["4"]
    O -->|Russian naming catalog| W["первая"]
```

The dotted arrow requires the planned musical system. The next diagram shows the planned model for modes and scales. In the ascending G major scale from G3 to G4, F♯ names the seventh degree and F♯4 is its written pitch.

```mermaid
flowchart LR
    S["Spelling · F♯"] -->|names| D["Degree VII"]
    M["Mode · G major"] -->|contains| D
    SC["Ascending scale · G3 to G4"] -->|traverses one octave of| M
    S -->|supplies the written name| P["Written pitch · F♯4"]
    O["Octave · ASPN 4"] -->|supplies the register| P
    SC -->|includes| P
```

## Use Cases

**Reading the same score across octave conventions.** A teacher working with a French score and a Russian textbook sees octave 3 in the score and *первая* (“first”) in the textbook. Both refer to the register containing middle C. A teaching tool can look up the register using Franco-Belgian numbering and display its Russian name alongside its ASPN number, 4. The student can then match the passages in both sources.

**Keeping triplet subdivisions exact across a measure.** An editor placing triplet eighth notes in a measure scales an eighth note (`1/8` of a whole note) by `2/3` to get `1/12`. Adding two triplet eighths on the same twelfth-note grid keeps the subdivision `2/12` while reducing to `1/6` of a whole note when only the net span is needed, and advancing to the seventh twelfth-note position `6/12` matches the half-measure mark `1/2` by value while preserving the active `1/12` grid.
