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

The library focuses on musical concepts and their relationships. It is not a standalone application, digital audio workstation, or sound player.

> **Early development:** The library is not ready for production use, and its public API is not yet stable. Only the `letter` package is implemented; it was first released in v0.1.0. Until v1.0.0, compatibility between releases is not guaranteed beyond what the package documentation states, and the API may change.

<!--
Update the relevant sections in the same PR that adds or changes public functionality.
Describe implemented behavior here; keep future work in ROADMAP.md.
Mark functionality not included in the latest release as unreleased.
-->

## Features

<!-- List implemented capabilities from the user's perspective, grouped by musical topic as needed. -->

### Letters

Package [`letter`](letter/) provides the seven letters used to name notes.

- The seven letters C D E F G A B in canonical order, starting with C.
- Walking the circle of letters upward from any letter, for example A B C D E F G.
- Letter distance between two letters, upward or downward, from 0 to 6.
- Shifting a letter up or down by any number of letter steps, wrapping around the circle: B shifted up by 1 is C.
- Comparing letters, however they were obtained: C shifted up by 9 is E.
- Converting between letters and their ordinals 0 to 6, with an error for numbers out of range.
- Detecting invalid values: the zero value is not a letter and is never treated as C.
- Immutable values that are safe for concurrent use.

## Concept

<!-- Define the library's core musical entities, their relationships, and any differences from conventional terminology. Keep definitions aligned with the implemented model. -->

### Letters

- **Letter (basic degree)** — one of the seven letters C, D, E, F, G, A, B. A letter is only a letter name: it has no accidental, octave, or pitch. It is not a notation symbol such as "H" or "do", and it is not a degree of a scale or mode.
- **Circle of letters** — the letters follow one another in a circle: after B comes C again. The circle has no beginning and does not describe pitch height.
- **Canonical order** — the agreed order C D E F G A B, starting with the first letter C. The **ordinal** of a letter is its place in this order, from 0 for C to 6 for B.
- **Letter step** — a move to the neighboring letter around the circle.
- **Letter distance** — the number of letter steps from one letter to another, upward or downward, from 0 to 6. It differs from the conventional interval number, which counts both letters: from C up to E the letter distance is 2, while the interval is a third.
- **Shift** — the letter reached after a given number of letter steps: up for a positive number, down for a negative one.
- **Seven-letter set** — the set of all seven letters. The range of distances 0 to 6 and the shift period of 7 are properties of this set. Sets with a different number of letters are not provided.
- **Invalid value** — a value that poses as a letter but is none of the seven. Operations recognize it and never treat it as C or any other letter.

## Use Cases

<!-- Describe practical musical or educational problems and the outcomes the library enables, in nontechnical language. Focus on why someone would use it, rather than API calls. -->

## Examples

<!-- Use a technical question-and-answer format: how to perform a task, a short explanation, and a link to a working Go example or the relevant API documentation. Add only examples supported by the implemented API. -->

### Letters

All examples below are in [`letter/example_test.go`](letter/example_test.go); their output is checked by `go test`.

- **How do I list all letters?**
  `letter.All()` returns C D E F G A B in canonical order; `letter.Count` is 7. Each call returns a new slice that the caller owns.
  Example: [`ExampleAll`](letter/example_test.go).
- **How do I walk the circle of letters from a given letter?**
  `letter.A.Cycle()` returns A B C D E F G: seven letters going upward, starting with A.
  Example: [`ExampleLetter_Cycle`](letter/example_test.go).
- **How do I find how many letter steps lie between two letters?**
  `letter.B.DistanceUp(letter.D)` is 2: counting upward wraps from B through C to D. `letter.C.DistanceDown(letter.E)` is 5. Both results are from 0 to 6.
  Examples: [`ExampleLetter_DistanceUp`](letter/example_test.go), [`ExampleLetter_DistanceDown`](letter/example_test.go).
- **How do I move a letter up or down by several steps?**
  `letter.C.Shift(9)` is E and `letter.C.Shift(-1)` is B. Any `int` is accepted, without overflow.
  Example: [`ExampleLetter_Shift`](letter/example_test.go).
- **How do I check whether two letters are the same?**
  `letter.C.Shift(9).Equal(letter.E)` is true. An invalid value is not equal to any letter.
  Example: [`ExampleLetter_Equal`](letter/example_test.go).
- **How do I create a letter from a number stored in my data?**
  `letter.FromOrdinal(4)` returns G. A number outside 0 to 6 returns an error that matches `letter.ErrInvalid` with `errors.Is`. `letter.MustFromOrdinal` panics instead and suits values known to be correct.
  Examples: [`ExampleFromOrdinal`](letter/example_test.go), [`ExampleMustFromOrdinal`](letter/example_test.go).
- **How do I detect a value that is not a letter?**
  `l.IsValid()` is false for the zero value `var l letter.Letter` and for any value other than the seven constants.
  Example: [`ExampleLetter_IsValid`](letter/example_test.go).
