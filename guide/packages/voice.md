# Melodic line

A melodic line is a voice as a line of the texture: the soprano of a chorale, the subject of a fugue, a doubling at the octave. It is a sequence of sounds and rests that follow each other in time, with a name and three roles. A line sounds **at most one pitch at a time**.

A melodic line is not the voice of a track. In notation, a voice is also a layer of a part written in measures, as in MusicXML, with ties between events, grace notes, and fermatas at bar lines; that layer belongs to a separate package for tracks. A line knows no measures, no track, and no other lines: how lines move against each other is the subject of voice leading, which is planned as a separate problem-solving package.

## Elements

Each element of a line is either a **sound** — one [note](note.md) with an octave — or a **rest**. Each has a duration and may carry a fermata sign. A chord, even of one note, a grace note, and a rest in measures are never elements: converting [events](event.md) that contain them is refused with distinct errors.

A duration keeps its notation as given, in one of three forms:

- a **written duration**, such as a dotted quarter;
- a **value** without notation, such as 3/8 of a whole note;
- a **tied duration** of two or more written parts, such as a quarter tied to an eighth.

A rule of the product: a sound under a tie is **one element** with a tied duration, not several elements with tie signs as in the voice of a track. The element stays one sound, so its intervals, its place in time, and the sounds of the line need no rules for continuing ties. A rest cannot have a tied duration; write several rests in a row instead. Converting a line to events gives one event per written part, joined by ties, and converting the events back gives the same line.

A dotted quarter and a quarter tied to an eighth last the same time but are written differently. The ordinary equality of lines compares the notation; a separate equality compares only how long the elements last.

## Positions without measures

The start of an element is the sum of the lengths of the elements before it, an exact fraction of a whole note counted from the start of the line. The length of a line is the sum of the lengths of all its elements; the empty line has length zero.

A moment at the end of an element belongs to the next element. A moment at or after the end of the line has no element, which is an answer, not an error. A line tells which element sounds at a moment, which note sounds or whether there is silence, and which elements overlap a span of time. This lets several lines be aligned moment by moment.

## Seconds

A rule of the product: a line gives the start and the end of each sound or element in seconds from a [tempo map](tempomap.md) addressed by positions. The caller gives the position at which the line starts on the map and always chooses the number of decimal places and the rounding mode.

A fermata of an element adds no time by itself; music theory gives no fixed length for it. The caller puts a hold of the chosen length on the tempo map at the end of the element, and the line lists these positions. The next sound then starts after the hold, and the held sound lasts until its end. Seconds of gradual tempo changes are approximate and keep the approximation flag and the guaranteed error bound.

## Pitch and the musical system

A line gives the melodic intervals from each sound to the next one with their direction, its range from the lowest sound to the highest, the indexes of these sounds, and the ambitus, the interval between them. A rule of the product: sounds separated by a rest also form an interval, marked as passing through a rest, so the caller can keep or ignore it.

Intervals are measured as written, so the ambitus of C♭4 and B♯3 is a descending diminished second. When several sounds share the lowest or the highest position, the first one in the line is taken.

Comparing pitches needs one [musical system](system.md). Each operation can take it explicitly. Otherwise it comes from the reference range of a preset or an instrument, if one is given, or from the settings of the sounds, which must then all be bound and share one system. Moving sounds — transposition, doubling, imitation, inversion — works on each sound in the system of its own settings, so a line may mix systems; an unbound sound needs an explicit system.

## Range of a voice or an instrument and tessitura

A line can be compared with the range of a singing voice or an [instrument](instrument.md): a range, a preset, an instrument, or the reference range of a role. It tells for each sound whether it is below, inside, or above the range and by how many steps, and whether the whole line fits.

A rule of the product: the **tessitura** of a line is the mean position of its sounds relative to the range, where the lower bound is 0 and the upper bound is 1, weighted by the lengths of the sounds or counting each sound once. The number is exact and may fall below 0 or above 1. Its kind — low, middle, or high — depends on zones; by default they divide the range into three equal parts, at 1/3 and 2/3. These zones are a choice of the product, not a fact of music theory, and the caller can set others. A middle tessitura means only that the sounds lie in the middle on average: it does not mean that the line fits in the range.

## Roles

A line has three independent roles, each of which may be unset:

- by place in the texture: soprano, alto, tenor, bass, or upper, middle, lower;
- thematic: leading, solo, accompanying, imitating, counterpoint, doubling, inversion, pedal;
- by importance: principal, subsidiary, supplementary, filler, harmonic.

Every role has an English and a Russian name. The roles are labels set by the user: the line never derives them from its notes. A soprano, alto, tenor, or bass role has the range of the built-in preset of that singing voice as its reference range, but the role is not a preset.

## Building lines

A line builds new lines and never changes itself:

- a **transposition** moves every sound by an interval and keeps the name and the roles;
- a **doubling** does the same and becomes a new line with the thematic role of a doubling;
- an **exact imitation** repeats the line at an interval, with or without a rest before its entry;
- an **exact inversion** reflects every sound around an axis, keeping the names of the intervals;
- a **retrograde** reverses the pitches, the rhythm, or both.

A doubling, an imitation, and an inversion have no name and no role by place or by importance: their place in the texture depends on the interval and on the context. The constructions are exact: they work by intervals, not by the degrees of a mode.

## Changing and sharing lines

Adding, inserting, replacing, removing, slicing, splitting, and joining elements, renaming a line, and changing its roles all give new lines. The original line, its elements, and the values passed in stay as they were, also when an operation fails. Reading a line returns copies. Lines can therefore be shared and read concurrently. A note in a line keeps the reference to its source of settings, so retuning that source changes how the note sounds.

Included in v0.42.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/voice) · [Usage](../../README.md#build-a-melodic-line) · [Go examples](../../voice/example_test.go) · [Roadmap](../../ROADMAP.md)
