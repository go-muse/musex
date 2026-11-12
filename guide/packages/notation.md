# Textual notation

Textual notation turns a note spelling into a string and reads that string back without losing the written letter or its exact alteration. A written pitch can also carry an octave designation. The notation is always chosen explicitly: the same characters can mean different names or registers in different traditions, and no default naming system is inserted silently.

This package concerns textual names, not graphical scores, staves, clefs, key signatures, octave transposition marks, frequencies, tuning, or machine serialization of spellings.

## Canonical output and aliases

Every notation has a finite domain of names. Within that domain, each name has exactly one **canonical output**. Parsing that output returns the original name, and two different names cannot have the same canonical string.

A notation may also accept explicit **aliases**. Aliases provide additional ways to read the same name but never create another canonical output. For example, a long accidental suffix may be accepted while formatting still chooses the short form. Matching is exact: whitespace, letter case, Unicode composition, and visually similar accidental signs are not normalized or exchanged implicitly.

Names outside the domain are rejected rather than rounded, replaced enharmonically, or approximated with a nearby sign. The same rule applies to unknown aliases and strings with extra characters.

## Built-in notations and their domains

There are 19 built-in records. Eighteen are complete notations; the ups-and-downs record is a parameterized family whose chromatic unit must be supplied.

| Record | Names and alteration domain | Octave support |
|---|---|---|
| English Unicode | 35 names with whole alterations from −2 through +2 | Visible numbers |
| English ASCII MT21 | 28 names from −1 through +2; no canonical double flat | Visible numbers |
| German AM | 33 names from −2 through +2 | AM, ASPN 0 through 8 |
| Russian AM | 35 names from −2 through +2 | AM, ASPN 0 through 8 |
| Italian syllabic | 35 names from −2 through +2 | None |
| ABC | 35 names from −2 through +2 | ABC marks |
| LilyPond Nederlands | 63 names from −2 through +2 in half-unit steps | LilyPond marks |
| LilyPond English | 63 names in the same half-unit domain, with short canonical suffixes and longer aliases | LilyPond marks |
| Stein–Zimmermann Unicode | 63 names in the same half-unit domain | None |
| Gould Unicode | 77 names, additionally including −5/2 and +5/2 | None |
| LilyPond Italiano | 63 names in the half-unit domain | LilyPond marks |
| Sims Unicode | 77 names, additionally including ±1/6 and ±1/3 | None |
| Wyschnegradsky Unicode | 175 names covering ±n/6 for n from 1 through 11, plus ordinary and double signs | None |
| AEU Unicode | 77 names whose values are expressed in AEU sharps of four commas | None |
| Persian Unicode | 49 names, including koron and sori nominally at ∓1/2 | None |
| Arabic Unicode | 63 names from −2 through +2 in half-unit steps | None |
| Stockhausen Unicode | 56 names at −2, −1, −1/2, 0, +1/2, +1, +3/2, and +2 | None |
| Ups and downs | 105 names for every chromatic unit from 3 through 65,536 | None |
| Fokker ASCII | 63 names in the half-unit domain | None |

The German AM domain deliberately excludes A double flat and E double flat because their forms are not confirmed. No separate built-in record writes German letters with an ASPN suffix such as H4; a user can define that combination explicitly without attributing it to a built-in tradition. AM octave designations are limited to ASPN 0 through 8.

Several Unicode families use SMuFL **Private Use Area** code points. These are not presented as symbols assigned by the Unicode standard. The package also does not claim that one universal standard exists for fractional accidentals.

## Written octave and positional octave

The octave written next to a name is called **w**. It is not necessarily the positional octave stored in a written pitch. In the traditional system, B sharp written with register 3 and C written with register 4 both belong to positional octave 4; C flat written with register 4 belongs to positional octave 3.

A musical system is required whenever a pitch is formatted or parsed. That system alone checks the name and translates between w and the positional octave. The notation package does not calculate octave carry itself and does not silently select the traditional system. For ups and downs, the notation's chromatic unit must also match the system's chromatic unit.

Names without octaves need no musical system, numbering, or naming catalog.

## Numbers, names, and special octave forms

A visible number uses an independently selected numbering. The octave containing middle C can therefore appear as 4 in ASPN, 3 in Franco-Belgian numbering, or 5 in a numbering that calls middle C C5. Changing the numbering changes only the visible label, not the octave or the name. Numeric text is strict: no plus sign, leading zeroes, or negative zero.

A visible name uses an independently selected catalog. An empty catalog is still a valid explicit choice, but it cannot provide a nonempty output name and cannot resolve an input name. The separator between the note name and octave name is exact and belongs to the notation.

AM, Helmholtz, numbered Helmholtz, LilyPond, and ABC have their own case and mark rules and do not need an independent numbering. ABC accepts mixed apostrophes and commas during parsing and adds their effects algebraically; canonical output uses only one kind of mark. The other supported special forms are not artificially limited to MIDI octaves, subject to their notation's domain and technical representation limits.

## Custom notations and conversion

Custom rules can be generated from letters and alteration forms or listed as a complete whole-name table. Whole-name exceptions replace both the generated canonical output and its generated aliases. Built-in name rules can be reused with a different octave style.

Construction verifies the entire grammar, including aliases, separators, case changes, and interactions with octave marks. Checks cover the complete octave axis symbolically rather than sampling a few registers. Ambiguous definitions are rejected before they can produce inconsistent output.

Conversion first parses with the source notation and then emits the canonical form in the target notation. Name conversion needs no system. Pitch conversion uses one explicit system at both ends and preserves the written name and positional octave, not merely the appearance of the string.

All notation values are immutable and safe for concurrent calls. Returned collections are independent copies, and parsing does not populate hidden caches.

First released in v0.18.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/notation) · [Go examples](../../notation/example_test.go) · [Roadmap](../../ROADMAP.md)
