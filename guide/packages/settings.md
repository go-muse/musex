# Settings

A set of settings is everything needed to read, write, and sound a written pitch in one consistent way. It joins three parts into one value:

- the **musical system**: which note names are admissible, how a name maps to a class of positions, and in which written octave a pitch is notated;
- the **tuning**: a temperament, the class it is rooted at, and a reference pitch with its frequency;
- the **designations**: a textual notation for names and octave marks, an octave numbering, and a catalog of octave names.

There is only one system in a set: it is always the system of the tuning. The three designations are chosen independently. Changing the octave numbering does not change the notation, and a catalog of octave names matters only to a notation that writes octaves by name. A catalog is never missing; a catalog without names is simply empty.

A set never changes. Replacing a part — the tuning, the notation, the numbering, the catalog, or all designations at once — gives a new set and leaves the old one as it was. A set keeps its own copies of the tuning and the catalog, so nothing done later to the values it was built from, or to the values read from it, can change it. Sets are equal when all their parts are equal; two empty, unset sets are equal to each other.

## Parts that must fit together

An ups-and-downs notation counts its arrows in steps of one chromatic unit of its own. Such a notation fits a set only when that unit equals the chromatic unit of the system; otherwise the set is refused. A notation without octave marks, such as the Italian syllables or the Unicode accidental tables, is allowed: such a set reads and writes names, but no pitch can be expressed in it.

Changing the system of a set keeps the temperament, the root number, the reference, and the designations. It is possible only for a system with the same number of positions per octave and with the reference note admissible in it; to move to another number of positions, the whole tuning is replaced.

## Standard settings

The standard settings are the traditional twelve-position system, twelve-tone equal temperament rooted at C with A4 at 440 Hz, English names with Unicode accidentals and visible octave numbers, ASPN numbering, and the English catalog of octave names. They are one ready-made set among many, not a default: a missing part is never filled in from them silently.

## Names and pitches a set can express

A name is **expressible** by a set when the system admits it and the notation can write it. A pitch is expressible when its name is expressible and its written octave can be shown: the octave in which the letter is notated, which can differ from the positional octave. In the traditional system B♯ in the positional octave 4 is written B♯3.

Which written octaves can be shown depends on the notation:

- visible numbers, the Helmholtz styles, LilyPond, and ABC show every octave;
- the AM style shows only ASPN octaves 0 through 8;
- a notation that writes octaves by name shows only the octaves named in the catalog of the set;
- a notation without octave marks shows none.

Writing and reading through the set use its notation, system, numbering, and catalog together. Every name or pitch that the set can express can be written, and everything the set reads is expressible by it. Writing a name the system does not admit is refused before the notation is asked.

## Sources of settings

A **source** gives the current set as a whole. A set is itself a source that always gives itself; an instrument that holds a current set and can switch to another one is a source too. A reader such as a note asks its source at the moment it needs the settings and works with the set it received. A source never gives a mixture of parts from two sets.

Two sets are the same source when they are equal, because a set gives only itself and equal sets cannot be told apart. Other sources are the same only when they are the same object, such as one instrument. A missing source is not a source at all.

## Compatibility of a replacement

When one set replaces another, nothing the previous set supported may stop being supported. Two rules are checked together:

- **structure**: every name admissible in the previous system is admissible in the next one;
- **designations**: every name and every pitch the previous set can express, the next set can express too. Each set uses its own system, so the written octave of the same pitch can differ between them.

When a rule is violated, the result names a concrete counterexample: a name, or a pitch, that the previous set supports and the next one does not. When both rules are violated, both are reported. Either rule can also be checked alone, and the structure rule can be checked for two systems before a whole set is built.

Notations that show every octave are treated as showing all of them, so the check is exact and finite: name domains are finite, and the octaves that can be shown are either all of them or a short list. For example, replacing the traditional system by the system of 24 quarter tones keeps every name, but the opposite replacement loses C raised by a quarter tone. Replacing English names with visible numbers by German names with AM octave marks loses A double flat, which the German record does not write, and every pitch written outside ASPN octaves 0 through 8.

The check compares only the two sets, not the notes actually in use. It can therefore refuse a replacement that would harm no real note; that is not an error of the check, and a new set can still be built separately.

## What is not an incompatibility

These are expected consequences of a compatible change and are never reported as violations:

- frequencies, step numbers, and positions change;
- the written octave number of the same pitch can change: with 19 positions B♯ in the positional octave 4 is written B♯4, not B♯3;
- which spellings sound the same can change;
- MIDI note numbers can become unavailable, because they exist only for the traditional layout;
- the system can stop being regular.

## Kinds of change

A replacement can change several things at once. Three kinds are reported separately:

- **structure** — the system changed. Its admissible names were extended (the old names are a strict part of the new ones), narrowed (the opposite), kept the same, or became incomparable;
- **sound** — the temperament, the root number, or the reference changed. A temperament with another number of positions always counts as a change. A change of sound alone is always compatible;
- **designations** — the notation, the numbering, or the catalog changed.

Replacing a set by an equal one is no change at all. Moving to a system with another number of positions always changes the sound as well; changing the system while keeping the same number of positions changes only the structure.

All values are immutable and safe for concurrent use.

First released in v0.22.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/settings) · [Go examples](../../settings/example_test.go) · [Roadmap](../../ROADMAP.md)
