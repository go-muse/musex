# Note

A note is what a score writes on a staff, plus how far it is meant to sound from that record. It has three parts:

- the **record**: a note name such as C♯, or a written pitch — a name in a positional octave, such as C♯ in octave 4;
- the **correction**: a deviation in cents, zero unless given. It does not change the name, the octave, or the position of the note; it only moves the sound. There is no limit on it: C4 raised by 150 cents still sits below C♯4 by position, while in twelve-tone equal temperament it sounds above it;
- the **source of settings**: where the note reads its settings from. It is optional.

A note without an octave is a name with a correction, such as "C♯ +30 cents". It has no frequency, no step number, no position, and no written octave: no octave is ever filled in for it. Asking for any of these is an error, not a guess.

## Bound and unbound notes

A note **bound** to a source of settings reads the source at the moment it needs the settings: to check a name, to write the note in a notation, to find its position, or to compute its frequency. The note does not keep the settings it read. When the owner of the source retunes it, the same note gives the new answers at its next use, without being recreated.

The source is only a role. It can be a set of settings, the standard settings, or any other owner that gives a current set — an instrument is one such owner. The note does not know what its source is, and it knows nothing of instruments.

An **unbound** note has no source. Everything that does not depend on settings works for it as for a bound note. Whatever needs settings has to be given them explicitly: a set of settings for that one operation, or just a tuning when only the sound and positions matter. Standard settings are never substituted silently. Giving settings for one operation does not bind the note.

A bound note keeps a name admissible in the system of its settings: creating it, renaming it, or deriving notes from it is refused when the new name is not admissible there. An unbound note is not checked.

## The frequency of a note

The note does not store a frequency and does not compute one itself. When asked for its frequency, it reads its settings, takes their tuning, and hands it the written pitch and the correction. The tuning counts how many steps the pitch is from its reference note, turns that into cents by its temperament, adds the correction, and applies the result to the reference frequency with a single rounding. So the frequency depends only on the record, the correction, and the tuning — not on the notation or the octave numbering.

For example, in the standard settings C4 sounds at 261.6256 Hz, and A4 raised by 100 cents sounds at 466.1638 Hz, exactly as A♯4 does. When the source is retuned to A4 at 415 Hz, the same C4 sounds at 246.7605 Hz.

The opposite direction works too: from a frequency, such as one measured by a tuner, a note is built with the nearest pitch of the tuning and with the deviation of the frequency as its correction.

## Changing a note and copies

A note is a value. Changing it — renaming it, raising its alteration, moving it by octaves or steps, replacing its correction, binding it to another source — changes that note itself; no second note appears. The correction and the source are kept through every change of the record. Several changes can be written one after another as a chain, and each step changes the same note.

Each change is all or nothing: when a change is refused, the note stays exactly as it was — record, correction, and source. A chain is not all or nothing: when a step fails, the steps before it stay done and the rest are not performed.

A copy of a note is independent: changing the copy does not change the original, and changing the original does not change the copy. Only the source is shared, because it belongs to its owner, not to the notes. Collections that hold notes, such as chords and tracks, keep their own notes and give out copies; a note inside such a collection is changed only through the collection.

Operations that give other notes — the same position under other names, the same name in other octaves, every step between two notes — return new notes and never change the note they start from. They take its correction and its source.

There are two ways to change the alteration of a note with an octave:

- keeping the **positional octave**: raising B in octave 3 gives B♯ in octave 3, which is written B♯2. This works without settings;
- keeping the **written octave**: raising B written in octave 3 gives B♯ written B♯3, which is in the positional octave 4. This needs the system of the settings.

Removing the sign works the same two ways: C♭ in the positional octave 3, written C♭4, becomes C3 by position or C4 by record.

## Comparing notes

Notes can be compared at three levels.

**By record**, without settings. Two notes are equal when their records and corrections are equal, whatever their sources are. C4 and C4 raised by 5 cents are different notes with one written pitch. A note with an octave never equals a note without one, and an unset note equals nothing. Parts can be compared separately: the name, the letter, the alteration, the octave, the correction, or whether two notes share one source. An agreed order sorts records: names without octaves first, then written pitches, then corrections. It is only a convention, neither equality nor sound.

**By position**, in the system of the settings. B♯3 and C4 are at one position, zero steps apart, though their records differ. A note with a correction keeps its position. Positions can be compared, measured in steps, and sorted only when the systems of all the notes are the same.

**By sound**, each note in its own settings, which may differ. C4 of the standard settings sounds lower than C4 of a 19-position equal temperament (261.6256 Hz against 264.0226 Hz), though by position the two cannot even be compared. When two notes share an equal tuning, the comparison is exact: the heights in cents, with their corrections, are compared instead of rounded frequencies. Otherwise frequencies are compared, strictly by default or within a tolerance in cents, in hertz, or relative to the frequency. A note can also be compared with a bare frequency: whether the frequency is below, within, or above the sound of the note, and by how many cents it deviates.

## Changing the source

A note can move to another source. Its name, correction, and record stay; two conditions are checked:

- the name must be admissible in the new system;
- whatever was expressible by the previous settings must stay expressible by the new ones. A note that its previous settings could not write loses nothing.

A refused move leaves the note with its previous source. A move can be checked in advance; the check names every condition that would be violated.

With the same system, nothing else changes. With another system, the move is made in one of two ways:

- keeping the **positional octave**: the written octave can change. B♯ in the positional octave 4 is written B♯3 in the traditional system and B♯4 with 19 positions;
- keeping the **written octave**: the record stays B♯3, and the positional octave is recomputed in the new system.

Either way the frequency, the step number, and the position follow the new settings, and which names sound the same can change. Copies of the note made before the move keep the previous source.

## Notes and time

A note has no duration and no place in time. Durations, onsets, and ties belong to events, voices, and tracks, which hold notes.

Reading one note is safe for concurrent use, provided reading its source is; changing a note needs exclusive access to it, and different copies can be changed at the same time.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/note) · [Go examples](../../note/example_test.go)
