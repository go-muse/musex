# Chord

A chord is a combination of [notes](note.md) that sound together. In musex it is just that: the notes, nothing more.

- A chord has **no duration** and no place in time. How long a chord lasts, where it starts, whether its notes are tied, arpeggiated, or held under a fermata — all this belongs to an event of a voice, which holds a single note or a chord.
- A chord has **no harmony**. Its intervals, quality, root, inversion, function, and name are left to a separate package for the harmony of chords. Whether three notes form a major triad is not a question for the chord itself.
- A chord is **not a set of pitch classes**. It keeps its notes as they are written; the classes of the notes and the sets they form come from the notes and from [pitch-class sets](pcset.md).

Any number of notes makes a chord, from none: a chord of one note is allowed, and so is the empty chord, a working value to which notes are added one by one. Sources disagree on whether a chord needs two or three sounds; that threshold belongs to harmony and is not checked here.

## The notes of a chord

A chord can hold any notes, in any combination:

- notes with an octave, such as C4, and notes without one, such as C — a chord of names only is useful when only the record matters;
- notes with corrections, such as C4 raised by 30 cents;
- notes bound to different sources of settings, and unbound notes.

The chord never reads the settings of its notes. It does not check whether their names are admissible, whether they can be written, or whether their systems are the same. Everything that needs settings — positions, steps, frequencies, ordering by pitch or by sound, writing in a notation — is done by the notes themselves: take the notes of a chord and ask each of them. In the standard settings the notes of the chord C4–E4 sound at 261.6256 Hz and 329.6276 Hz.

## Order and repeats

A chord keeps the **order of record**: the notes stay in the order in which they were given, and they come back in that order. The chord never reorders them by itself. Removing or replacing notes keeps the order of the rest; a replacing note takes the place of the one it replaces. Only inserting notes at a chosen place and sorting change the order. Sorting can follow an agreed order of records, a rule of the user — for example, by position in the settings of the notes — or simply reverse the order.

A note can enter a chord **more than once**. Nothing written is dropped: two notes that are equal by record may still be bound to different sources and sound differently, and losing one would lose a sounding note. A repeat is the second and every further occurrence of a note. Finding a note gives its first occurrence; the chord also tells how many times a note enters it and at which places. When a chord should work as a set, its repeats can be removed, notes can be added only when missing, and the distinct notes can be read in the order of record.

## Equality by record

Two notes are the same note when their records and their corrections are equal; the source does not count. So C4 and C4 raised by 5 cents are different notes, C4 and C without an octave are different notes, and B♯ in octave 4 and C4 are different notes although they may share a position. Two C4 notes bound to different settings are one note entering a chord twice.

Two chords are **equal by record** when they have the same notes with the same number of occurrences, in any order: C4, E4, G4 equals G4, C4, E4. The order of record does not count, because the notes of a chord sound at once. Two other equalities are available:

- **equality in order**, which also needs the notes at the same places to be equal;
- **equality without the corrections**, under which C4 and C4 raised by 5 cents count as one record.

A chord can also include another — contain all its notes, counting repeats — or share notes with it. Chords can be joined, kept to their common notes, or one subtracted from another; repeats are counted in each case. An agreed order sorts chords for steady output; it agrees with equality by record and says nothing about which chord is higher.

The empty chord is a real chord: it equals another empty chord and is included in every set chord. A chord that was never made is unset: it accepts no notes and equals nothing, not even another unset chord.

## Changing a chord and copies

A chord is a value. Adding, inserting, removing, or replacing notes, sorting them, moving all of them by octaves, or changing every note by a rule of the user — each change changes that chord itself; no second chord appears. Several changes can be written one after another as a chain, and each step changes the same chord.

Each change is all or nothing: when a change is refused — a note to remove is not in the chord, a place is outside it, a note without an octave cannot be moved by octaves, a rule of the user fails — the chord stays exactly as it was. A chain is not all or nothing: when a step fails, the steps before it stay done and the rest are not performed.

A copy of a chord is independent from the start: changing the copy does not change the original, and changing the original does not change the copy. Only the sources of the notes are shared, because they belong to their owners. A copy is cheap: the notes are copied only when one of the chords changes.

The notes a chord gives out are copies. Changing such a note does not change the chord; a note inside a chord is changed only through the chord. The notes given to a chord are copied too, so changing them afterwards does not change it.

Reading one chord is safe for concurrent use; changing a chord needs exclusive access to it, and different copies can be changed at the same time.

First released in v0.28.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/chord) · [Go examples](../../chord/example_test.go) · [Roadmap](../../ROADMAP.md)
