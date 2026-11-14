# Instrument

An instrument is the owner and operator of a set of [settings](settings.md). It holds the current set, gives it to whoever reads it, and replaces it with another set only after a check. Besides the set, it has a name and a reference range of pitches.

## Settings, instrument, and note

Three roles are kept apart:

- a **set of settings** is a value: a musical system, a tuning, and the designations of written pitches. It never changes;
- an **instrument** owns its current set and can switch to another one. It is a source of settings: it gives its current set as a whole, never a mixture of two sets;
- a **note** reads its source at the moment it needs the settings. It can hold an instrument as its source, but it knows nothing of instruments.

The dependency goes one way. An instrument does not hold, register, or check notes, and it does not compute frequencies, positions, or written forms itself: the tuning, the system, and the settings do that through the current set. When an instrument is retuned, the [notes](note.md) bound to it give the new answers at their next use, without being recreated.

## Presets and instances

A **preset** describes an instrument: a name, a set of settings, and a reference range, which can be absent. A preset never changes and has no identity: two presets with equal parts are simply equal. Replacing its name, its set, or its range gives a new preset.

An **instance** of an instrument is created from a preset or directly from a name, a set, and a range. It does not remember the preset it came from: changing the instance never changes the preset, other instances, or a catalog, and there is no way to return an instance to its preset. Two instances made from one preset are two different instruments with equal states. A copy of an instance is a new, independent instrument.

An instance has identity, and identity is what makes it a source: an instrument is the same source only as itself. Two instances with equal states are different sources, and an instrument and the set it gives are different sources too. Whether two instances are the same instrument and whether their states are equal are two separate questions.

The name of an instance can be changed, and its range can be replaced or removed. Its state is always correct: the name is not empty, the set is present, and the range is absent or valid in the system of the current set. A change that would break this is refused and changes nothing.

## Catalogs

A **catalog** is an ordered collection of presets with different names. Lookup by name is exact: another case of letters, an extra space, or a synonym finds nothing. A catalog never changes; adding or removing a preset gives a new catalog. Adding a name that is already there, or removing a name that is missing, is refused, so a typo never passes silently. There is no registry of all presets: a preset made by a program works exactly like a ready one, whether it is in a catalog or not.

## Ready-made presets

The library offers 39 ready-made presets with English names and the standard settings:

- bowed strings: violin, cello, double bass;
- plucked strings: classical guitar, bass guitar, mandolin;
- woodwinds: flute, oboe, English horn, clarinets in B-flat and in A, bass clarinet, bassoon, contrabassoon, soprano, alto, and tenor saxophones;
- brass: horn, trumpet in B-flat, trombone, bass trombone, bass tuba, contrabass tuba;
- keyboards: piano with 88, 85, and 92 keys, organ, harpsichord, celesta;
- pitched percussion: timpani, xylophone, marimba, vibraphone;
- voices: soprano, mezzo-soprano, alto, tenor, baritone, and bass.

The names of the voices end with the word "Voice", so that "Bass Voice" is not confused with the bass guitar or the double bass. The groups above are only the order of the catalog; the presets themselves carry no group. The ranges come from the domain knowledge of the library: only sounding ranges are used, for the usual variant of each instrument, and where sources disagree the wider bound is taken. Different kinds of one instrument, such as the three pianos or the two tubas, are separate presets. Instruments for which the sources give no sounding bounds — the viola, the harp, the ukulele, the piccolo, the baritone saxophone, the accordion, the banjo, and the glockenspiel — are not in the catalog yet.

Every request for a ready-made preset or for the ready-made catalog gives new values, so nothing a program does with them can change them for anyone else.

## The reference range

A range is two written pitches: a lower bound and an upper bound, both included. It does not belong to a musical system: the same two pitches can form a valid range in one system and an inverted one in another. A range is **valid** in a system when the system admits both names and the lower bound is not above the upper one. An inverted range is refused; its bounds are never swapped.

Pitches are compared with a range by **position**, never by the written octave number. In the traditional system B♯3 stands at the position of C4, so it is in a range that starts at C4; C♭4 stands at the position of B3 and is below it. The soprano range C4–G5 is 19 steps wide and holds 20 sounds. A correction in cents plays no part: a note is checked by its written pitch.

A range is only a reference. A pitch outside it is still a valid pitch: nothing is forbidden because of a range, neither creating notes nor retuning.

A range can tell whether a pitch is below, inside, or above it, how wide it is, and how many sounds it holds. A pitch can be **folded** into it: moved by whole octaves, keeping its name, until it lands inside. E1 folds into C3–G5 as E3, E4, and E5, and the nearest fold is E3. A range narrower than an octave can have no fold at all; that is an answer, not an error. Two ranges have an intersection, which can be empty, and a hull — the smallest range that includes both; one range can include another. A set of pitches has its own range, from its lowest to its highest pitch. A range can list its pitches one step apart with names chosen by a spelling rule, and two ranges can be compared by record or by sound.

An instrument applies these operations to its own range in the system of its current set. Retuning to another system can therefore change the answer for the same pitch: A double flat in octave 5 is inside C4–G5 in the traditional system, but above it with 19 positions per octave.

## Ranges of transposing instruments

The ranges of the ready-made presets are **sounding** ranges. For a transposing instrument — the clarinet in B-flat, the trumpet in B-flat, the English horn, the saxophones, the horn — and for instruments written an octave away from their sound — the guitar, the double bass, the bass guitar, the contrabassoon, the xylophone, the celesta — the range lists the pitches that sound, not the pitches that are written. Translating written pitches of such instruments into sounding ones is not provided yet.

## Retuning

Retuning replaces the whole set of an instrument by another set. It can be given a ready set, a change of the current set, or just one part to replace: the tuning, the system, the notation, the octave numbering, the catalog of octave names, or all designations at once. In every case a new whole set is built and checked; a part is never changed in place. There is no separate operation to change the reference note: the tuning is replaced, for example by the same pitch level expressed through another reference note.

A retuning succeeds only when both conditions hold:

- the new set is **compatible** with the current one by the rules of the settings: every name and pitch the current set supports stays supported;
- the range of the instrument, if it has one, is **valid** in the new system.

When either condition fails, the instrument stays exactly as it was: name, set, and range. An incompatible set is refused with the report of the settings, which names the broken rule and a concrete counterexample; a range that would become inverted is refused with its own reason, distinguishable from incompatibility. The range is never removed silently. To move to such a system, the instrument is retuned together with a new range, or without a range, in one action. A retuning can also be checked in advance without changing anything.

The check compares only the two sets and the range, not the notes in use, so it can refuse a change that would harm no actual note. That is not an error; a new instrument with the needed set can still be created.

These are expected consequences of a successful retuning and never count as incompatibility: frequencies, step numbers, positions, and written octave numbers of the notes change, and with another system the positions of the range bounds change too. Retuning to an equal set succeeds and changes nothing visible. Retuning never changes identity: the instrument stays the same source for every note that holds it.

## Working from several threads

An instrument can be read and changed from several threads at once without extra locking. A reader always gets the previous state or the new one as a whole, never a mixture, and never waits for a check in progress. Changes are made one at a time, and a check always belongs to the state it replaces. A change computed from the current set is indivisible: concurrent changes of this kind lose nothing. While computing such a change, a program may read instruments but must not change any instrument; doing so would make the threads wait for each other forever.

Ranges, presets, and catalogs are immutable and safe for concurrent use.

First released in v0.25.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/instrument) · [Go examples](../../instrument/example_test.go) · [Roadmap](../../ROADMAP.md)
