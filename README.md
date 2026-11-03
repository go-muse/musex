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

> **Early development:** The library is not ready for production use, and its public API is not yet stable. The `letter`, `alteration`, `spelling`, `octave/number`, `octave/naming`, `octave` and `pitchclass` packages are implemented; `letter` was first released in v0.1.0, `alteration` in v0.2.0, `octave/naming` in v0.3.0, `octave/number` in v0.4.0, `spelling` in v0.5.0, `octave` in v0.6.0, and `pitchclass` in v0.7.0. Until v1.0.0, compatibility between releases is not guaranteed beyond what the package documentation states, and the API may change.

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

### Alterations

Package [`alteration`](alteration/) provides alteration values: how many chromatic units an accidental raises or lowers a letter.

- The five traditional accidentals: natural 0, sharp +1, flat -1, double sharp +2, and double flat -2.
- Any whole number of chromatic units, such as +3 or -4, and exact fractional values, such as +1/2 for a quarter tone up. Fractions are kept in lowest terms: 2/4 is the same value as 1/2.
- Raising or lowering a value by one unit, by any whole number of units, or by another alteration value, whole or fractional: +1/2 raised by +1/2 is a sharp.
- The difference between two values: a sharp is +2 higher than a flat.
- Multiplying a value by a whole number: +1/2 multiplied by 2 is +1.
- Comparing values, however they were obtained, and ordering them by size; checking the direction of a value and whether it is natural or whole.
- The exact value as a fraction in lowest terms, for storage and exchange.
- A guaranteed range of magnitudes up to 2^24 chromatic units with denominators up to 2^16, with exact results inside it. A result outside the range returns an error instead of a rounded or wrapped value.
- The zero value is the natural.
- Immutable values that are safe for concurrent use.

### Note names

Package [`spelling`](spelling/) provides note names: a letter together with an alteration value, without an octave.

- A note name from any letter and any alteration value, including fractional values and values beyond the traditional accidentals: C with +1/2, E with +3, or B with -7/2.
- The 35 ready-made names of the seven letters with the natural, sharp, flat, double sharp, and double flat, such as C, C sharp, and B double flat, and a list of all of them in sorting order.
- Raising or lowering a name by one chromatic unit, by any whole number of units, or by any alteration value, always keeping the letter: C sharp raised by one unit is C double sharp, never D.
- The natural name of the same letter, and the same letter with another alteration value.
- Comparing names as written, not by sound: C sharp is not equal to D flat, and B sharp is not equal to C.
- An agreed order for sorting names: by letter, then by alteration value. It is not an order by pitch.
- The letter distance between two names and the difference between their alteration values; neither is an interval.
- Checking whether a name is natural or whole, its direction (raised, lowered, or natural), and whether two names have the same letter or the same alteration value.
- A fixed machine form for storage and exchange, such as `C+1` or `B-7/2`, as text and in JSON, including map keys. Each name has exactly one form; it is not a notation.
- Errors instead of guessed results: for an invalid letter, for a result outside the guaranteed range of alteration values, and for text that is not a machine form.
- Detecting unset names: the zero value is not a name and is never treated as C.
- Immutable values that are safe for concurrent use.

### Octave numbering

Package [`octave/number`](octave/number/) provides octave numberings: the rules that give octaves their visible numbers.

- Five built-in numberings in a fixed order, each with a label: ASPN (also called SPN and IPN), Franco-Belgian, Mexican, Middle C = C3, and Middle C = C5.
- Your own numbering from a single pair, such as "the octave with ASPN number 4 has my number 5", optionally with numbers that are skipped.
- The number of an octave in any numbering from its ASPN number, and the ASPN number of an octave from its number.
- Converting an octave number between any two numberings, built-in or your own: number 3 with middle C as C3 is number 3 in the Franco-Belgian numbering.
- Skipped numbers are reported as errors and never replaced with a nearby number: no octave has number 0 in the Franco-Belgian numbering.
- Comparing numberings by the numbers they give to all octaves, regardless of label: Mexican equals Middle C = C5.
- Any `int` octave number, with no musical or MIDI limits; a result that does not fit in an `int` is an error, never wrapped around.
- Detecting undefined values: the zero value is not a numbering and is never treated as ASPN.
- Immutable values that are safe for concurrent use.

### Octave names

Package [`octave/naming`](octave/naming/) provides catalogs of verbal octave names.

- Built-in Russian, English and German catalogs: the Russian catalog names the octaves with ASPN numbers -1 to 9, the English and German catalogs the octaves 0 to 7.
- The name of an octave by its ASPN number: octave 4 is "one-line" in English and "eingestrichene" in German.
- An empty name, without an error, for an octave that has no name in a catalog. Any `int` is accepted as an ASPN number.
- The ASPN number of an octave by its name, with an error for a name that is not in the catalog.
- Custom catalogs from your own pairs of ASPN numbers and names, with an error when two octaves are given the same name.
- Comparing catalogs by their names, however they were created.
- Exact names: names are stored and looked up exactly as given, without trimming, normalization or case changes.
- Immutable catalogs that are safe for concurrent use.

### Octaves

Package [`octave`](octave/) provides octaves: elements of the axis of octaves, shown by their numbers and names.

- An octave by its number in any numbering, built-in or your own, or by its name in any catalog, built-in or your own: number 3 in the Franco-Belgian numbering, number 4 in ASPN, and "первая" in the Russian catalog are the same octave.
- The number of an octave in any numbering and its name in any catalog, with an empty name, without an error, for an octave that has no name.
- Comparing octaves, however they were obtained; two values of the same octave are also the same map key.
- Ordering and sorting octaves from lower to higher, and checking whether an octave is above or below another or lies between two octaves.
- Shifting an octave up or down by any number of octaves, and the next octave above or below.
- How many octaves one octave is higher than another, and the distance between them: the octaves numbered 3 and -1 in the Franco-Belgian numbering are 3 octaves apart, although their numbers differ by 4.
- The octaves from one to another, upward or downward, one at a time, with the option to stop at any point.
- Your own numbering from a single octave, such as "the octave of middle C is my number 1", optionally with skipped numbers, and your own catalog of names from pairs of an octave and a name.
- Errors instead of guessed results: a skipped number, an undefined numbering, an unknown name, and an invalid definition are reported with the same errors as in `octave/number` and `octave/naming`, and a nearby octave is never used instead.
- Any octave, with no musical or MIDI limits: every octave whose ASPN number fits in an `int`. A result that does not fit in an `int` is an error, never wrapped around.
- Detecting invalid values: the zero value is not an octave and is never treated as ASPN 0 or 4.
- Immutable values that are safe for concurrent use.

### Pitch classes

Package [`pitchclass`](pitchclass/) provides pitch classes: the positions within an octave of a system with any number of positions, together with the arithmetic of pitch-class theory.

- A class from its number and the number of positions *N*, such as `1 mod 12`, strictly or by wrapping any integer around the circle: -1 becomes `11 mod 12`, and 13 becomes `1 mod 12`.
- All the classes of *N* in order, and a walk around the circle upward from any class.
- Transposition T*n* by any whole number of positions, in either direction.
- The ordered interval from one class to another, the interval class of a pair, and the index sum; the ordered, complementary, and class intervals of an interval given by its size.
- Inversion I*n* (also written T*n*I) by an index, around an axis class, or by a pair of classes it exchanges, and the classes an inversion leaves in place.
- Multiplication M*m*, the invertible multipliers of *N*, the inverse of a multiplier, and their number φ(*N*), computed quickly even for very large *N*.
- Operators x ↦ m·x + n, which include the 48 twelve-tone operators (TTOs) for 12 positions: creation, application to a class, composition, inverse, and type (transposition or inversion).
- The groups of transpositions, of transpositions and inversions, and of all operators of *N*, and their sizes, exact for any *N*.
- Canonical text forms for storage and exchange, `1 mod 12` and `T5M7 mod 12`, with strict parsing.
- The literary notation of class numbers in four styles: `10 11`, `T E`, `t e`, and `A B`, with strict parsing in the chosen style.
- Classes and operators as map keys and with an agreed order for sorting; this order is not an order by pitch.
- Any *N* from 1 to the largest `int`, with exact results for any `int` input; lists are lazy, so their first elements are available at once for any *N*.
- Errors instead of guessed results: for classes of different *N*, a multiplier that is not invertible, invalid text, and a count that does not fit in an `int`.
- Detecting unset values: the zero value is not a class and is never treated as `0 mod 12`.
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

### Alterations

- **Chromatic unit** — the unit of alteration values: the amount by which a sharp raises a letter. It is not a semitone in general or a position of a musical system: how many positions a chromatic unit spans depends on the system.
- **Alteration value** — the exact number of chromatic units by which an accidental raises a letter (a positive value) or lowers it (a negative value). It is only a number: it has no letter, octave, or pitch, and it is not an accidental symbol, a cent offset, or a modal alteration. Whether a value is allowed in a musical system is a property of that system, not of the value.
- **Natural** — the value 0. It is a value in its own right, not a missing value; the zero value of the type is the natural.
- **Traditional accidentals** — natural 0, sharp +1, flat -1, double sharp +2, and double flat -2. They are not the only values: +3 (a triple sharp) is whole too, and +1/2 and -3/2 are fractional (microtonal) values.
- **Exact value** — a value as a fraction in lowest terms with a positive denominator: 0/1 for the natural and n/1 for a whole value n. There are no decimal approximations.
- **Raising and lowering** — making a new value a + d or a - d, where d is a whole number or another alteration value. The original value does not change.
- **Difference** — a - b: how many chromatic units a is higher than b. It is not an interval, which also needs letters.
- **Order by size** — comparing values as numbers. It is not the order of note pitches, which also depend on the letter, octave, and system.
- **Direction** — raising (a value above 0), lowering (below 0), or natural (0).
- **Guaranteed range** — the values for which the package promises exact results: magnitudes up to 2^24 chromatic units and denominators up to 2^16. It is not a limit on accidentals in music and not a rule of any musical system.

### Note names

- **Note name (spelling)** — a letter together with an alteration value, such as C sharp (the letter C with +1) or B with -7/2. A note name has no octave and no pitch: it is how a note is written, not how it sounds. It is not a string of a notation such as "C♯", "Cis", or "до-диез".
- **Ready-made names** — the 35 names of the seven letters with the traditional accidentals: natural, sharp, flat, double sharp, and double flat. Any other alteration value forms a name too.
- **Equal names** — names with the same letter and the same alteration value. Names that may sound the same are not equal: C sharp and D flat are different names, and so are B sharp and C. Which names sound the same, and which names are allowed, depends on the musical system that uses them.
- **Raising and lowering a name** — changing its alteration value while keeping its letter. There is no enharmonic substitution: C sharp raised by one chromatic unit is C double sharp, not D. The original name does not change.
- **Order for sorting** — an agreed order: by letter in the canonical order C D E F G A B, then by alteration value. It is not an order by pitch: B sharp sorts after C, and C flat sorts before B.
- **Letter distance and alteration difference** — the letter distance between the letters of two names, and the difference between their alteration values. From C sharp up to E flat the letter distance is 2; C sharp minus C flat is +2. Neither is an interval: they have no interval quality or direction.
- **Machine form** — the fixed text form of a name for storage and exchange: the letter alone for a natural name, such as `C`; otherwise the letter, the sign, and the value, such as `C+1`, `D-1`, `C+1/2`, or `B-7/2`. Each name has exactly one form, and the form is part of the package API. It is not a notation: accidental symbols such as `C#` are not accepted.
- **Unset name** — the zero value: a name that is not given. It is not C and not a natural name, and it is written as `unset`.

### Octave numbering

Octaves are identified by their **ASPN numbers**, as defined in the Octave names part of this section below: the octave that contains middle C is 4. An octave numbering gives each octave a visible number of its own.

- **Octave numbering** — a rule that gives each octave a visible number. Numbers increase from lower to higher octaves and continue without limit in both directions; there is no lowest or highest octave, and MIDI limits do not apply. A numbering is not a notation: it defines only numbers, not how a note such as "C♯4" or "do3" is written.
- **Skipped number** — a number that no octave has in a numbering. In the Franco-Belgian numbering, number 0 is skipped: the octave numbered 1 is followed downward by −1.
- **Built-in numbering** — one of the five numberings provided by the library, each with a label:
  - **ASPN** — the number of an octave equals its ASPN number; middle C is in octave 4. Also called SPN and IPN; used in English-language literature and science. Labels on Roland and Korg instruments match ASPN.
  - **Franco-Belgian** — middle C is in octave 3, and number 0 is skipped. Used in France, Belgium, and parts of Spain; in Italy it is prescribed by a law of 1989, but ASPN is common there in practice. Sources confirm the numbers −2 to 7; beyond them, the library continues the numbers in sequence.
  - **Mexican** — the numbering of the Spanish translations of Riemann; middle C is in octave 5. Used in Mexico alongside ASPN. A single encyclopedia article confirms the numbers from 1, for the octave of the lowest audible C (ASPN 0); below 1, the library continues the numbers in sequence.
  - **Middle C = C3** — middle C is in octave 3. Used by Yamaha instruments and by default in Logic Pro, and reportedly in Cubase.
  - **Middle C = C5** — middle C is in octave 5. Used in FL Studio.
- **Own numbering** — a numbering you define with an **anchor pair**, "the octave with ASPN number *a* has number *x*", and optionally a list of skipped numbers. Without skipped numbers, it is a shift of ASPN. An own numbering has no label.
- **Label** — the name of a built-in numbering, such as "Franco-Belgian". It identifies the numbering for people and does not affect its numbers.
- **Conversion** — finding the number in one numbering of the octave that has a given number in another numbering.
- **Equal numberings** — numberings that give the same number to every octave, however they were defined and whatever their labels. Mexican and Middle C = C5 are equal; Franco-Belgian and Middle C = C3 are not, although both put middle C in octave 3.
- **Undefined value** — a value that poses as a numbering but is none. Operations recognize it and never treat it as ASPN or any other numbering.

### Octave names

- **Octave** — in the traditional system, the range of pitches from C up to the next B; see [Octaves](#octaves-1) for the general definition.
- **ASPN number** — the number of an octave in the ASPN (scientific pitch notation) octave numbering: the octave that contains middle C is 4, neighboring octaves have consecutive numbers, and any integer is a valid number, including negative ones. The number identifies the octave itself; it is not the written octave of a note, such as a note for a transposing instrument.
- **Octave name** — a word that names an octave, such as "great", "small" or "one-line" in English. A name is not a notation symbol such as c¹ or C₁, and it is not an inflected word form.
- **Catalog of octave names** — a set of pairs of an ASPN number and a name. Each octave has at most one name in a catalog, and different octaves have different names; the same name may appear in different catalogs. A catalog is not an octave numbering, and it has no name of its own.
- **Empty name** — the name of an octave that has no name in a catalog. It is not an error, and a pair with an empty name does not add a name to a catalog. A catalog may have no names at all.
- **Built-in catalogs** — the Russian, English and German catalogs. Their sets of named octaves differ: the octaves -1, 8 and 9 have names only in the Russian catalog. The Russian names for the octaves -1 and 9 are an extension of musex: they follow the pattern of the other names but were not found in the literature.
- **Exact match** — names are compared exactly, byte for byte: "Small" and " small" are not the name "small". Built-in names are in Unicode normalization form NFC, so ё and ß are single characters: "ss" in place of ß does not match.
- **Equal catalogs** — catalogs that name the same octaves with the same names, however they were created.

### Octaves

- **Octave** — an element of the ordered sequence of octaves, which has no end in either direction; neighboring octaves differ by exactly one octave. The identity of an octave is its place in this sequence, and it is separate from its number and its name. The octave of a note is the octave section of its position in a musical system, and the system decides where that section begins and ends; in the traditional system, an octave is the range of pitches from C up to the next B.
- **Octave axis** — the sequence of all octaves, ordered from lower to higher.
- **Visible number and name** — the number of an octave in a numbering and its name in a catalog. They are computed from the numbering or the catalog you choose and are not stored in the octave. How the library represents an octave internally is not visible.
- **ASPN number of an octave** — one of its visible numbers. It identifies the octave uniquely and is the common reference of octave numberings and octave names; this is why the sections above say that octaves are identified by their ASPN numbers and that the number identifies the octave itself. There is no contradiction: the identity of an octave is its place on the axis, and the ASPN number points to exactly that place.
- **Order** — lower and higher octaves. It agrees with the order of numbers in every numbering.
- **Shift** — the octave a given number of octaves higher, or lower for a negative number. The next octave above or below is a shift by one.
- **Difference** — how many octaves one octave is higher than another; negative if it is lower. It counts octaves, not numbers: in the Franco-Belgian numbering, the octaves numbered 1 and −1 are neighbors, so their difference is 1.
- **Distance** — the difference without direction.
- **Range of octaves** — the octaves from one octave to another, both included, upward or downward.
- **Representable octaves** — the octaves the library can work with: all octaves whose ASPN number fits in an `int`. This is a technical limit, not a musical one; a result beyond it is an error. A number at the very edge of `int` that converts between numberings without an error may still not give a representable octave.
- **Invalid value** — a value that poses as an octave but is none. It is not the octave with ASPN number 0 or 4, and it does not mean that a note has no octave. Operations recognize it and never treat it as any octave.

### Pitch classes

- **Pitch class** — a position within the octave of a system with *N* positions, numbered from 0 to *N* − 1 and written `k mod N`. A class holds only its number and *N*: it does not remember the system, the octave, the note name, or the pitch. Classes of different systems with the same *N* are equal; classes of different *N* are never equal. In the traditional 12-position system, C is 0, C♯ and D♭ are both 1, and B is 11 (integer notation); results depend on which class is numbered 0. Which class a note name or a step of a system has will be computed by the planned package `system`.
- **Ordered interval** — the distance from one class upward to another around the circle, from 0 to *N* − 1: from 10 to 2 modulo 12 it is 4, and from 2 to 10 it is 8. The two directions give a pair of complementary intervals.
- **Interval class** — the smaller of the two ordered intervals of a pair, from 0 to *N*/2: 0 and 7 modulo 12 are in interval class 5. The interval class of a class with itself is 0; the literature names only the interval classes 1 to 6 for 12 positions.
- **Transposition T*n*** — moving every class by *n* positions around the circle.
- **Inversion I*n*** — the reflection x ↦ n − x, written I*n* or T*n*I; *n* is its index. The inversion I without an index is I0. An inversion can also be given by an axis class, which stays in place, or by a pair of classes it exchanges, whose index sum is *n*. For 1 or 2 positions, every inversion is also a transposition.
- **Multiplication M*m*** — x ↦ m·x. If *m* has a common divisor with *N*, different classes merge; otherwise *m* is invertible, as 1, 5, 7, and 11 are for 12 positions. M7 maps the chromatic circle of 12 classes onto the circle of fifths.
- **Operator** — x ↦ m·x + n with an invertible multiplier *m*, written T*n*M*m*: multiply by *m*, then transpose by *n*. For 12 positions there are 48 operators, the twelve-tone operators of Morris; in general, *N*·φ(*N*). Composition f∘g applies g first, then f. The literature defines these operations for 12 positions; musex generalizes them to any *N* by the same formulas.
- **Canonical form** — the fixed text form of a class, `1 mod 12`, or of an operator, `T5M7 mod 12`, for storage and exchange. Each value has exactly one form.
- **Literary notation** — the number of a class without *N*, as in the literature: 10 and 11 are written `10` and `11`, `T` and `E` (MT21), `t` and `e` (Whittall, OMT), or `A` and `B` (Morris). The letter styles apply only to 12 positions or fewer.
- **Unset value** — the zero value: a class or an operator that is not given. It is not `0 mod 12` and not the identity, and it is written as `Class(invalid)` or `Operator(invalid)`.

## Use Cases

<!-- Describe practical musical or educational problems and the outcomes the library enables, in nontechnical language. Focus on why someone would use it, rather than API calls. -->

- **Reading octave numbers and names from different traditions.** A French score calls an octave the third, a Russian textbook calls the same octave "первая" (the first), an English book calls it the one-line octave, and software that follows ASPN shows it as 4. An application built on `musex` can take an octave given in any of these conventions and show it in any other, so a student, a teacher, or an editor sees that all of them mean the same octave.

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

### Alterations

All examples below are in [`alteration/example_test.go`](alteration/example_test.go); their output is checked by `go test`.

- **What can I do with alteration values?**
  `alteration.MustFromFraction(1, 2).MustRaiseBy(1)` is +3/2, and a sharp minus a flat is +2. `String` prints a fixed form such as `+3/2`; it is not a notation.
  Example: [`Example`](alteration/example_test.go).
- **How do I get the traditional accidentals?**
  `alteration.Natural()`, `Sharp()`, `Flat()`, `DoubleSharp()`, and `DoubleFlat()` return 0, +1, -1, +2, and -2. The zero value `var a alteration.Alteration` is the natural.
  Example: [`Example_signs`](alteration/example_test.go).
- **How do I create a whole value, such as a triple sharp?**
  `alteration.FromInt(3)` returns +3. A value beyond `alteration.MaxMagnitude` returns an error that matches `alteration.ErrOutOfRange` with `errors.Is`; `alteration.MustFromInt` panics instead.
  Example: [`ExampleFromInt`](alteration/example_test.go).
- **How do I create a fractional value, such as a quarter tone?**
  `alteration.FromFraction(1, 2)` returns +1/2. Fractions are reduced and the sign is kept in the numerator: 2/4 is +1/2, 1/(-2) is -1/2, and 4/2 is +2. A zero denominator returns an error that matches `alteration.ErrZeroDenominator`. `alteration.MustFromFraction` panics instead and suits literals.
  Examples: [`ExampleFromFraction`](alteration/example_test.go), [`ExampleMustFromFraction`](alteration/example_test.go).
- **How do I raise or lower a value by one chromatic unit?**
  `alteration.Flat().Raise()` returns the natural, and `Lower` goes the other way. Each operation returns a new value; the usual style is `a = a.MustRaise()`.
  Example: [`ExampleAlteration_Raise`](alteration/example_test.go).
- **How do I raise or lower a value by several units?**
  `alteration.Natural().RaiseBy(3)` is +3, and `alteration.Sharp().RaiseBy(-2)` is -1. `LowerBy` lowers instead. Any `int` is accepted, without overflow.
  Example: [`ExampleAlteration_RaiseBy`](alteration/example_test.go).
- **How do I raise a value by a fraction?**
  `a.Add(d)` raises `a` by any alteration value `d`: +1/2 plus +1/2 is +1, and +1/2 plus +1/3 is +5/6.
  Example: [`ExampleAlteration_Add`](alteration/example_test.go).
- **How do I find how much higher one value is than another?**
  `alteration.Sharp().Sub(alteration.Flat())` is +2, and +1/2 minus +1/3 is +1/6. `Sub` also lowers a value by another value.
  Example: [`ExampleAlteration_Sub`](alteration/example_test.go).
- **How do I check whether two values are the same, or sort them?**
  `MustFromFraction(2, 4).Equal(MustFromFraction(1, 2))` is true, and `==` gives the same answer. `a.Compare(b)` returns -1, 0, or +1 and works with `slices.SortFunc(xs, alteration.Alteration.Compare)`.
  Examples: [`ExampleAlteration_Equal`](alteration/example_test.go), [`ExampleAlteration_Compare`](alteration/example_test.go).
- **How do I multiply a value by a whole number?**
  `a.Mul(k)` returns `a` times `k`: +1/2 times 2 is +1, and -3/2 times 4 is -6.
  Example: [`ExampleAlteration_Mul`](alteration/example_test.go).
- **How do I tell whether a value raises or lowers, or whether it is whole?**
  `a.Sign()` is +1, -1, or 0; `a.IsNatural()` and `a.IsInteger()` report whether the value is 0 or a whole number.
  Examples: [`ExampleAlteration_Sign`](alteration/example_test.go), [`ExampleAlteration_IsInteger`](alteration/example_test.go).
- **How do I store a value and read it back?**
  `a.Fraction()` returns the exact value as a numerator and a positive denominator in lowest terms, and `alteration.FromFraction(a.Fraction())` gives back the same value. `String` is for output only.
  Examples: [`ExampleAlteration_Fraction`](alteration/example_test.go), [`ExampleAlteration_String`](alteration/example_test.go).
- **What happens when a result is too large?**
  The guaranteed range covers magnitudes up to `alteration.MaxMagnitude` (2^24) with denominators up to `alteration.MaxDenominator` (2^16). `alteration.MustFromInt(alteration.MaxMagnitude).Raise()` returns an error that matches `alteration.ErrOutOfRange`; the original value does not change.
  Example: [`Example_outOfRange`](alteration/example_test.go).

### Note names

All examples below are in [`spelling/example_test.go`](spelling/example_test.go); their output is checked by `go test`.

- **What can I do with note names?**
  `spelling.MustNew(letter.C, alteration.Sharp())` is C sharp, written `C+1`. `spelling.MustNew(letter.C, alteration.Sharp()).MustRaise()` is `C+2`, which equals `spelling.CDoubleSharp()`, while C sharp is not equal to D flat.
  Example: [`Example`](spelling/example_test.go).
- **How do I create a name from a letter and an alteration value?**
  `s, err := spelling.New(letter.B, alteration.MustFromFraction(-7, 2))` returns B with -7/2, written `B-7/2`. An invalid letter, such as the zero value `letter.Letter(0)`, returns an error that matches `letter.ErrInvalid` with `errors.Is`. `spelling.MustNew` panics instead and suits values known to be valid.
  Examples: [`ExampleNew`](spelling/example_test.go), [`ExampleNew_invalidLetter`](spelling/example_test.go), [`ExampleMustNew`](spelling/example_test.go).
- **How do I get common names such as C sharp?**
  `spelling.C()`, `spelling.CSharp()`, `spelling.DFlat()`, and `spelling.BDoubleFlat()` return ready-made names. `spelling.Predefined()` lists all 35 of them in sorting order, starting with `C-2 C-1 C C+1 C+2`; each call returns a new slice that the caller owns.
  Examples: [`Example_readyNames`](spelling/example_test.go), [`ExamplePredefined`](spelling/example_test.go).
- **How do I get the letter and the alteration value of a name?**
  `l, a := spelling.FSharp().Parts()` gives the letter F and the value +1; `Letter()` and `Alteration()` return them one at a time.
  Example: [`ExampleSpelling_Parts`](spelling/example_test.go).
- **How do I check whether two names are the same?**
  `spelling.CSharp().Equal(spelling.DFlat())` is false: names are compared as written, not by sound. C with 2/4 equals C with 1/2. For set names, `==` gives the same answer, so names can be map keys.
  Example: [`ExampleSpelling_Equal`](spelling/example_test.go).
- **How do I raise or lower a name?**
  `spelling.C().MustRaise()` is `C+1`, `spelling.CSharp().MustRaiseBy(-2)` is `C-1`, and `spelling.CDoubleSharp().MustLowerBy(4)` is `C-2`. The letter never changes, and each operation returns a new name; the usual style is `s = s.MustRaise()`.
  Examples: [`ExampleSpelling_Raise`](spelling/example_test.go), [`ExampleSpelling_Lower`](spelling/example_test.go), [`ExampleSpelling_RaiseBy`](spelling/example_test.go), [`ExampleSpelling_LowerBy`](spelling/example_test.go).
- **How do I raise a name by a quarter tone or another fractional value?**
  `spelling.CSharp().MustRaiseByAlteration(alteration.MustFromFraction(1, 2))` is `C+3/2`, and `spelling.C().MustLowerByAlteration(alteration.MustFromFraction(1, 2))` is `C-1/2`.
  Example: [`ExampleSpelling_RaiseByAlteration`](spelling/example_test.go).
- **How do I get the natural name or change the accidental?**
  `spelling.CSharp().Natural()` is C, and `spelling.CSharp().WithAlteration(alteration.Flat())` is C flat.
  Examples: [`ExampleSpelling_Natural`](spelling/example_test.go), [`ExampleSpelling_WithAlteration`](spelling/example_test.go).
- **How do I sort names?**
  `slices.SortFunc(names, spelling.Spelling.Compare)` sorts by letter, then by alteration value. The order is not by pitch: B sharp sorts after C.
  Example: [`ExampleSpelling_Compare`](spelling/example_test.go).
- **How do I compare the letters or the alteration values of two names?**
  `spelling.CSharp().LetterDistanceUp(spelling.EFlat())` is 2 and `LetterDistanceDown` is 5. `spelling.CSharp().MustAlterationDifference(spelling.CFlat())` is +2. Neither result is an interval.
  Examples: [`ExampleSpelling_LetterDistanceUp`](spelling/example_test.go), [`ExampleSpelling_LetterDistanceDown`](spelling/example_test.go), [`ExampleSpelling_AlterationDifference`](spelling/example_test.go).
- **How do I check whether a name is natural, whole, raised, or lowered?**
  `IsNatural()` and `IsInteger()` report whether the alteration value is 0 or a whole number, and `Direction()` returns `raised`, `lowered`, or `natural`. `SameLetter` and `SameAlteration` compare one part of two names: C sharp and C flat have the same letter.
  Examples: [`ExampleSpelling_IsNatural`](spelling/example_test.go), [`ExampleSpelling_IsInteger`](spelling/example_test.go), [`ExampleSpelling_Direction`](spelling/example_test.go), [`ExampleDirection_String`](spelling/example_test.go), [`ExampleSpelling_SameLetter`](spelling/example_test.go), [`ExampleSpelling_SameAlteration`](spelling/example_test.go).
- **How do I store a name and read it back?**
  `String()` and `MarshalText()` return the machine form, such as `C+1`, and `s, err := spelling.Parse("C+1/2")` reads it back. Only the exact form is accepted: `spelling.Parse("C#")` returns an error that matches `spelling.ErrInvalidForm`. `spelling.MustParse` panics instead and suits literals.
  Examples: [`ExampleSpelling_String`](spelling/example_test.go), [`ExampleSpelling_MarshalText`](spelling/example_test.go), [`ExampleParse`](spelling/example_test.go), [`ExampleMustParse`](spelling/example_test.go).
- **How do I use names in JSON?**
  Names are JSON strings with their machine form in struct fields, arrays, and map keys, for example `{"tonic":"D","notes":["F+1","unset"],"names":{"B-1":"B"}}`. This works with both `encoding/json` and `encoding/json/v2`; JSON `null` reads as the unset name.
  Example: [`Example_json`](spelling/example_test.go).
- **What happens when a value is too large?**
  `top := spelling.MustNew(letter.C, alteration.MustFromInt(alteration.MaxMagnitude))` is `C+16777216`, and `_, err := top.Raise()` returns an error that matches `alteration.ErrOutOfRange`; `top` does not change.
  Example: [`Example_outOfRange`](spelling/example_test.go).
- **How do I detect a name that was never set?**
  `s.IsValid()` is false for the zero value `var s spelling.Spelling`, which is written as `unset`. It is not C: `s.IsNatural()` is false, and operations such as `s.Raise()` return an error that matches `letter.ErrInvalid`. Its alteration part is the natural, so check `IsValid` before using the parts.
  Example: [`ExampleSpelling_IsValid`](spelling/example_test.go).

### Octave numbering

All examples below are in [`octave/number/example_test.go`](octave/number/example_test.go); their output is checked by `go test`. An overview is in [`Example`](octave/number/example_test.go).

- **How do I list the built-in numberings?**
  `number.Builtins()` returns ASPN, Franco-Belgian, Mexican, Middle C = C3, and Middle C = C5 in this order. Each call returns a new slice that the caller owns. Each numbering is also available directly, for example `number.FrancoBelgian()`.
  Example: [`ExampleBuiltins`](octave/number/example_test.go).
- **How do I find the number of an octave in a numbering?**
  `number.FrancoBelgian().FromASPN(2)` is 1, and `FromASPN(1)` is −1: number 0 is skipped. The octave is given by its ASPN number; middle C is in ASPN octave 4.
  Example: [`ExampleNumbering_FromASPN`](octave/number/example_test.go).
- **How do I find which octave has a given number?**
  `number.FrancoBelgian().ToASPN(3)` is 4, and `number.Mexican().ToASPN(5)` is 4. A skipped number returns an error that matches `number.ErrSkipped` with `errors.Is`: `number.FrancoBelgian().ToASPN(0)` is an error.
  Examples: [`ExampleNumbering_ToASPN`](octave/number/example_test.go), [`ExampleNumbering_ToASPN_skipped`](octave/number/example_test.go).
- **How do I convert an octave number from one numbering to another?**
  `number.Convert(3, number.MiddleC3(), number.FrancoBelgian())` is 3: both numbers refer to the octave of middle C. `number.MustConvert` panics instead of returning an error and suits values known to be valid.
  Examples: [`ExampleConvert`](octave/number/example_test.go), [`ExampleMustConvert`](octave/number/example_test.go).
- **How do I define my own numbering?**
  `number.New(4, 5)` gives number 5 to ASPN octave 4 and number 1 to ASPN octave 0. Skipped numbers follow the pair: `number.New(4, 3, 0)` equals the Franco-Belgian numbering. A pair whose number is among the skipped numbers, or a repeated skipped number, returns an error that matches `number.ErrInvalidDefinition`.
  Examples: [`ExampleNew`](octave/number/example_test.go), [`ExampleNew_skipped`](octave/number/example_test.go), [`ExampleNew_invalid`](octave/number/example_test.go).
- **How do I check whether two numberings are the same?**
  `number.Mexican().Equal(number.MiddleC5())` is true: they give the same numbers to all octaves, although their labels differ. `==` does not compile for numberings.
  Example: [`ExampleNumbering_Equal`](octave/number/example_test.go).
- **How do I show a numbering to a user or in a log?**
  `Label()` returns the label of a built-in numbering, such as "Franco-Belgian", and an empty string for your own numbering. `String()` always returns a fixed form: `number.MustNew(4, 5).String()` is "ASPN 4 -> 5".
  Examples: [`ExampleNumbering_Label`](octave/number/example_test.go), [`ExampleNumbering_String`](octave/number/example_test.go).
- **What happens when a number does not fit in an `int`?**
  The operation returns an error that matches `number.ErrOutOfRange`; the result never wraps around.
  Example: [`Example_outOfRange`](octave/number/example_test.go).
- **How do I detect a numbering that was never set?**
  `n.IsDefined()` is false for the zero value `var n number.Numbering`. Operations on it return an error that matches `number.ErrUndefined`; it is never treated as ASPN.
  Example: [`Example_zeroValue`](octave/number/example_test.go).

### Octave names

All examples below are in [`octave/naming/example_test.go`](octave/naming/example_test.go); their output is checked by `go test`.

- **How do I get the name of an octave?**
  `naming.Russian().Name(4)` is "первая", and `naming.Russian().MustASPN("малая")` is 3.
  Example: [`Example`](octave/naming/example_test.go).
- **Which octaves do the built-in catalogs name?**
  `naming.Russian()` names the octaves -1 to 9; `naming.English()` and `naming.German()` name the octaves 0 to 7. Each call returns a new catalog that the caller owns.
  Examples: [`ExampleRussian`](octave/naming/example_test.go), [`ExampleEnglish`](octave/naming/example_test.go), [`ExampleGerman`](octave/naming/example_test.go).
- **What do I get for an octave without a name?**
  `naming.English().Name(8)` is "", without an error. Any `int` is accepted, including negative numbers.
  Example: [`ExampleCatalog_Name`](octave/naming/example_test.go).
- **How do I find an octave by its name?**
  `naming.Russian().ASPN("первая")` returns 4. A name that is not in the catalog, such as "Первая", returns an error that matches `naming.ErrUnknownName` with `errors.Is`; the number returned with the error is not an answer. `MustASPN` panics instead and suits names known to be correct.
  Examples: [`ExampleCatalog_ASPN`](octave/naming/example_test.go), [`ExampleCatalog_MustASPN`](octave/naming/example_test.go).
- **How do I create my own catalog?**
  `naming.New(map[int]string{4: "middle", -3: "deep", 12: "high"})` creates a catalog in which `ASPN("deep")` is -3. An empty name means "no name". The same name for two octaves returns an error that matches `naming.ErrDuplicateName`; `MustNew` panics instead.
  Examples: [`ExampleNew`](octave/naming/example_test.go), [`ExampleMustNew`](octave/naming/example_test.go).
- **How do I check whether two catalogs are the same?**
  `naming.Russian().Equal(naming.Russian())` is true, although the two pointers differ. A catalog with the pairs 1 "a" and 2 "" is equal to a catalog with the pair 1 "a". Use `Equal`, not `==`, which compares pointers.
  Example: [`ExampleCatalog_Equal`](octave/naming/example_test.go).

### Octaves

All examples below are in [`octave/example_test.go`](octave/example_test.go); their output is checked by `go test`.

- **How do I show an octave from one convention in another?**
  `octave.MustFromNumber(3, number.FrancoBelgian())` is the octave of middle C: its name is "первая" in `naming.Russian()` and "one-line" in `naming.English()`, and its ASPN number is 4.
  Example: [`Example`](octave/example_test.go).
- **How do I get an octave by its ASPN number?**
  `octave.FromASPN(4)` returns the octave of middle C and never fails: every `int` is an ASPN number. It prints as `ASPN 4`.
  Example: [`ExampleFromASPN`](octave/example_test.go).
- **How do I get an octave by its number in a numbering?**
  `octave.FromNumber(3, number.FrancoBelgian())`, `octave.FromNumber(5, number.Mexican())`, and `octave.FromNumber(3, number.MiddleC3())` return the same octave, ASPN 4. A skipped number, such as 0 in the Franco-Belgian numbering, returns an error that matches `number.ErrSkipped` with `errors.Is`, and an undefined numbering returns an error that matches `number.ErrUndefined`; no other octave is used instead. `octave.MustFromNumber` panics instead and suits values known to be valid.
  Examples: [`ExampleFromNumber`](octave/example_test.go), [`ExampleFromNumber_skipped`](octave/example_test.go), [`ExampleFromNumber_undefined`](octave/example_test.go), [`ExampleMustFromNumber`](octave/example_test.go).
- **How do I get an octave by its name?**
  `octave.FromName("малая", naming.Russian())` returns ASPN 3, which is number 2 in the Franco-Belgian numbering. A name that is not in the catalog, such as "Первая", returns an error that matches `naming.ErrUnknownName`.
  Examples: [`ExampleFromName`](octave/example_test.go), [`ExampleFromName_unknown`](octave/example_test.go).
- **How do I find the number or the name of an octave?**
  `o.Number(n)` returns the number of `o` in any numbering: ASPN 4 is 4, 3, 5, 3, and 5 in the five built-in numberings. `o.ASPN()` returns the ASPN number. `o.Name(c)` returns the name in any catalog, or "" without an error: ASPN 8 has no English name.
  Examples: [`ExampleOctave_Number`](octave/example_test.go), [`ExampleOctave_ASPN`](octave/example_test.go), [`ExampleOctave_Name`](octave/example_test.go).
- **How do I check whether two octaves are the same?**
  `octave.FromASPN(4).Equal(octave.MustFromNumber(3, number.FrancoBelgian()))` is true. For octaves, `==` gives the same answer, so octaves can be map keys.
  Example: [`ExampleOctave_Equal`](octave/example_test.go).
- **How do I sort octaves or compare their heights?**
  `a.Compare(b)` returns -1, 0, or +1 and works with `slices.SortFunc(xs, octave.Octave.Compare)`. `a.Above(b)` and `a.Below(b)` compare two octaves, and `o.Between(a, b)` reports whether `o` lies between `a` and `b`, both included, in either order.
  Examples: [`ExampleOctave_Compare`](octave/example_test.go), [`ExampleOctave_Above`](octave/example_test.go), [`ExampleOctave_Between`](octave/example_test.go).
- **How do I move an octave up or down?**
  `o.Shift(k)` moves `o` up by `k` octaves, or down for a negative `k`, and `o.ShiftDown(k)` moves it down; `ShiftDown(k)` reverses `Shift(k)` for any `int`. `o.Up()` and `o.Down()` return the neighboring octaves: in the Franco-Belgian numbering, the octave above number -1 is number 1. Each operation returns a new octave.
  Examples: [`ExampleOctave_Shift`](octave/example_test.go), [`ExampleOctave_ShiftDown`](octave/example_test.go), [`ExampleOctave_Up`](octave/example_test.go).
- **How many octaves apart are two octaves?**
  `octave.FromASPN(5).MustSub(octave.FromASPN(4))` is 1: how many octaves the first is higher than the second. In the Franco-Belgian numbering, the octaves numbered 3 and -1 are 3 octaves apart. `a.Distance(b)` is the same without direction.
  Examples: [`ExampleOctave_Sub`](octave/example_test.go), [`ExampleOctave_Distance`](octave/example_test.go).
- **How do I go through the octaves from one to another?**
  `for o := range octave.MustRange(octave.FromASPN(2), octave.FromASPN(5))` gives the octaves ASPN 2 to 5; with the ends swapped, it goes downward. The octaves are produced one at a time, and the loop may stop at any point. `octave.Range` returns an error for an invalid end.
  Example: [`ExampleRange`](octave/example_test.go).
- **How do I define my own numbering or catalog by octaves?**
  `octave.NewNumbering(o, 1)` creates a numbering in which `o` has number 1, and `octave.NewNumbering(octave.FromASPN(4), 3, 0)` equals the Franco-Belgian numbering. `octave.NewCatalog(map[octave.Octave]string{octave.FromASPN(4): "middle"})` creates a catalog of names. An invalid definition returns an error that matches `number.ErrInvalidDefinition`, and the same name for two octaves an error that matches `naming.ErrDuplicateName`.
  Examples: [`ExampleNewNumbering`](octave/example_test.go), [`ExampleNewNumbering_invalid`](octave/example_test.go), [`ExampleNewCatalog`](octave/example_test.go), [`ExampleNewCatalog_duplicate`](octave/example_test.go).
- **How do I print an octave in a log?**
  `fmt.Println(o)` prints a fixed form such as `ASPN 4`, and `%#v` prints `octave.FromASPN(4)`. The forms are for logs and debugging: they are not a notation such as "C4" and have no parser.
  Examples: [`ExampleOctave_String`](octave/example_test.go), [`ExampleOctave_GoString`](octave/example_test.go).
- **What happens at the edges of `int`?**
  `octave.FromASPN(math.MaxInt).Shift(1)` returns an error that matches `octave.ErrOutOfRange`; the result never wraps around. `octave.FromNumber(math.MaxInt, number.MiddleC3())` does too, because that octave's ASPN number does not fit in an `int`, although `number.Convert` converts the same number to the Franco-Belgian numbering without an error.
  Example: [`Example_outOfRange`](octave/example_test.go).
- **How do I detect a value that is not an octave?**
  `o.IsValid()` is false for the zero value `var o octave.Octave`. It is not ASPN 0 or 4: `Number` and `ASPN` return an error that matches `octave.ErrInvalid`, `Name` returns "", and `Equal` is false even for two zero values, while `==` is true.
  Examples: [`ExampleOctave_IsValid`](octave/example_test.go), [`Example_zeroValue`](octave/example_test.go).

### Pitch classes

All examples below are in [`pitchclass/example_test.go`](pitchclass/example_test.go); their output is checked by `go test`. An overview is in [`Example`](pitchclass/example_test.go).

- **How do I create a pitch class?**
  `pitchclass.New(1, 12)` returns `1 mod 12`. `New` does not wrap numbers: `pitchclass.New(12, 12)` returns an error that matches `pitchclass.ErrInvalidClass` with `errors.Is`. `pitchclass.Reduce(13, 12)` wraps any integer around the circle and returns `1 mod 12`; `Reduce(-1, 12)` returns `11 mod 12`. `MustNew` and `MustReduce` panic instead and suit values known to be valid.
  Examples: [`ExampleNew`](pitchclass/example_test.go), [`ExampleReduce`](pitchclass/example_test.go).
- **How do I list the classes of a system?**
  `pitchclass.All(12)` returns an iterator over `0 mod 12` to `11 mod 12`; use `slices.Collect` for a slice. For any number of positions, the first classes are available at once. `c.Cycle()` walks the circle upward from `c`.
  Example: [`ExampleAll`](pitchclass/example_test.go).
- **How do I transpose a class?**
  `pitchclass.MustNew(11, 12).Transpose(2)` is `1 mod 12`, and `pitchclass.MustNew(0, 12).Transpose(-1)` is `11 mod 12`.
  Example: [`ExampleClass_Transpose`](pitchclass/example_test.go).
- **How do I find the interval between two classes?**
  `pitchclass.MustNew(10, 12).IntervalTo(pitchclass.MustNew(2, 12))` is 4, and in the other direction it is 8. `IntervalClass` gives the interval class of a pair: 4 for 10 and 2, 5 for 0 and 7. Classes of different numbers of positions, such as `1 mod 12` and `1 mod 19`, return an error that matches `pitchclass.ErrModulusMismatch`. For an interval given by its size, use `pitchclass.ReduceInterval`, `pitchclass.ComplementInterval`, and `pitchclass.IntervalClass`.
  Examples: [`ExampleClass_IntervalTo`](pitchclass/example_test.go), [`ExampleClass_IntervalTo_modulusMismatch`](pitchclass/example_test.go), [`ExampleClass_IntervalClass`](pitchclass/example_test.go).
- **How do I invert a class?**
  `pitchclass.MustNew(2, 12).Invert(8)` is `6 mod 12`: I8 maps 2, 4, and 7 to 6, 4, and 1. `c.Invert(0)` is I without an index. `InvertAround` inverts around an axis class, and `InvertSwapping` by a pair of classes that it exchanges. `pitchclass.InversionFixedClasses(4, 12)` returns the classes I4 leaves in place, `2 mod 12` and `8 mod 12`; for I1 it returns an empty slice.
  Examples: [`ExampleClass_Invert`](pitchclass/example_test.go), [`ExampleInversionFixedClasses`](pitchclass/example_test.go).
- **How do I multiply a class?**
  `pitchclass.MustNew(1, 12).Multiply(7)` is `7 mod 12`; M7 maps 0 to 11 onto 0 7 2 9 4 11 6 1 8 3 10 5, the circle of fifths. Any multiplier is accepted. `pitchclass.Invertible(m, n)`, `pitchclass.InverseMultiplier`, `pitchclass.InvertibleMultipliers`, and `pitchclass.Totient` describe the multipliers that do not merge classes: φ(12) is 4.
  Examples: [`ExampleClass_Multiply`](pitchclass/example_test.go), [`ExampleTotient`](pitchclass/example_test.go).
- **How do I work with twelve-tone operators?**
  `pitchclass.NewOperator(7, 5, 12)` returns `T5M7 mod 12`, the operator x ↦ 7x + 5; it is not a transposition. `MustApply(pitchclass.MustNew(1, 12))` maps 1 to `0 mod 12`. `t3.Compose(i8)` applies I8 first, then T3, and gives `T11M11 mod 12`; `i8.Compose(t3)` gives `T5M11 mod 12`. `Transposition`, `Inversion`, and `Multiplication` create T*n*, I*n*, and M*m*; a multiplier that is not invertible, such as 2 modulo 12, returns an error that matches `pitchclass.ErrNotInvertible`.
  Examples: [`ExampleNewOperator`](pitchclass/example_test.go), [`ExampleOperator_Apply`](pitchclass/example_test.go), [`ExampleOperator_Compose`](pitchclass/example_test.go).
- **How do I list or count the operators of a system?**
  `pitchclass.Operators(12)` returns an iterator over the 48 operators, and `pitchclass.CountOperators(12)` is 48. `Transpositions` and `TranspositionsInversions` list the smaller groups. A count that does not fit in an `int`, such as `CountOperators(math.MaxInt)`, returns an error that matches `pitchclass.ErrOverflow`; `CountOperatorsBig` returns the exact value.
  Examples: [`ExampleOperators`](pitchclass/example_test.go), [`ExampleCountOperators`](pitchclass/example_test.go).
- **How do I store a class or an operator as text?**
  `pitchclass.MustNew(1, 12).String()` is `1 mod 12`, and `pitchclass.Parse("1 mod 12")` returns an equal class. `ParseOperator` reads forms such as `T5M7 mod 12`. Parsing is strict: `01 mod 12` and `12 mod 12` are errors that match `pitchclass.ErrNotation`.
  Example: [`ExampleClass_String`](pitchclass/example_test.go).
- **How do I write 10 and 11 as T and E?**
  `pitchclass.MustNew(10, 12).Notation(pitchclass.TE)` is `T`. The four styles write 10 and 11 as `10 11`, `T E`, `t e`, and `A B`. `pitchclass.ParseNotation("E", pitchclass.TE, 12)` returns `11 mod 12`. The letter styles apply to 12 positions or fewer; for more, they return an error that matches `pitchclass.ErrNotation`.
  Examples: [`ExampleClass_Notation`](pitchclass/example_test.go), [`ExampleParseNotation`](pitchclass/example_test.go).
- **Can I use classes as map keys?**
  Yes: equal classes are equal with `==`, so `pitchclass.MustReduce(13, 12)` and `pitchclass.MustNew(1, 12)` are the same key. Use `Equal` to compare classes in code; `slices.SortFunc(classes, pitchclass.Class.Compare)` sorts them by number of positions, then by number.
  API documentation: [`Class`](https://pkg.go.dev/github.com/go-muse/musex/pitchclass#Class), [`Class.Compare`](https://pkg.go.dev/github.com/go-muse/musex/pitchclass#Class.Compare).
- **How do I detect a class that was never set?**
  `c.IsValid()` is false for the zero value `var c pitchclass.Class`, which is not `0 mod 12`. Its `String()` is `Class(invalid)`, and operations on two classes return an error that matches `pitchclass.ErrInvalidValue`.
  Example: [`Example_zeroValue`](pitchclass/example_test.go).
