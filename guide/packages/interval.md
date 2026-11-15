# Interval

A named interval is the distance between two notes as a score writes it: a major third, a perfect fifth, an augmented fourth. It has three parts:

- the **direction**: up or down;
- the **size**: how many letters the interval spans, counting both ends — 1 for a unison, 3 for a third, 8 for an octave, 10 for a tenth;
- the **quality**: perfect, major, minor, augmented, diminished, and their multiples, such as doubly diminished.

An interval is read from the spelling of the names — their letters and accidentals — not from how they sound. It is not a number of semitones. C to E is a major third, and C to F♭ is a diminished fourth, although in the traditional twelve-position system both span four steps. B♯ to C is a diminished second, not a unison. Because no musical system is involved, the same interval exists between any two names, in any system.

## Quality

The sizes fall into two groups by their simple part. Unisons, fourths, fifths, and octaves form the **perfect group**: their basic kind is perfect. Seconds, thirds, sixths, and sevenths form the **major-minor group**: their basic kind is major, and the minor kind is one chromatic unit narrower.

The quality of an interval is how far it is from the basic kind of its size. One unit wider is augmented, two units doubly augmented, and so on without a fixed limit. One unit narrower than perfect, or than minor, is diminished. A perfect third or a major fifth does not exist.

Systems with quarter tones and other fractional accidentals give fractional qualities. The interval exactly halfway between minor and major is **neutral**: C to E lowered by a quarter tone is a neutral third. Other fractional qualities exist as values too, but have no traditional name.

## The unison and the octave

The perfect unison has no direction: going nowhere is the same up and down. There is no diminished unison. When the same letter is lowered — C to C♭ in one register — the interval is an augmented unison down.

Between two names without octaves, the same letter with different accidentals gives two answers depending on the direction. Following the motion of the sound gives an augmented unison: C to C♯ up, or C to C♭ down. Going against it gives a diminished octave: C to C♭ up is a diminished octave.

## Simple and compound intervals

An interval up to an octave is **simple**; a wider one is **compound**. Every interval has a simple part with the same quality and direction: a major tenth is a major third plus one octave, a perfect fifteenth is a perfect octave plus one octave. The octave is its own simple part, so the simple part of an interval is never wider than an octave and never narrower than a second, except for a unison.

Intervals are also basic or derived, diatonic or chromatic, narrow or wide. **Basic** intervals are the perfect, major, and minor ones, together with the augmented fourth and the diminished fifth — the tritones; they are also the **diatonic** intervals. The other intervals with whole qualities are **chromatic**; an interval with a fractional quality is neither. Up to a fourth an interval is **narrow**, from a fifth to an octave it is **wide**; compound intervals are neither. Only the simple augmented fourth and diminished fifth count as tritones.

## Inversion

Inverting an interval takes its simple part and completes it to an octave: the sizes of an interval and its inversion add up to 9, major turns into minor, augmented into diminished of the same multiplicity, perfect stays perfect, and neutral stays neutral. A major third inverts to a minor sixth, and a major tenth inverts the same way, through its simple part. The direction is kept.

An augmented octave has no inversion: it would have to be a diminished unison, which does not exist.

## Sums and the line of fifths

Intervals add by moving names: a major third followed by a minor third reaches as far as a perfect fifth, and a perfect fifth followed by a perfect fourth makes an octave. A sum always means "move by the first, then by the second", so it works up and down: a minor second followed by an augmented unison down is a diminished second. The steps of a major scale — major, major, minor, major, major, major, minor seconds — add up to a perfect octave.

Names with whole accidentals have places on the line of fifths: F, C, G, D, A, E, B, then each sharp seven places further and each flat seven places back. The distance between two places is enough to name the interval between the names: four places is a major third, six an augmented fourth.

## Between pitches and notes

Between two names, the interval is always within an octave. Between two written pitches — names in an octave — it can be compound, and it needs a musical system. The system tells where the written octave of each pitch begins: in the traditional system the pitch at the position of C4 can be written B♯3. The direction of the interval follows the writing, not the position: from B♯3 to C4 is a diminished second up, although both are at one position. No system is ever assumed; it is given explicitly or taken from the settings of the notes.

Moving a name, a pitch, or a note by an interval never substitutes another spelling: B♯ moved up by an augmented sixth is G with three sharps. Moving by an interval and then by its opposite gives back the start.

The correction in cents of a note does not take part in intervals: C4 raised by 30 cents to E4 lowered by 20 cents is still a major third. A moved note is a new note with the correction and the source of the original; the original note does not change.

## Positional size in a system

The **positional size** of an interval in a musical system is the number of steps of the system it spans: in the traditional system a major third spans four steps, a perfect fifth seven. It depends on the system: an augmented fourth and a diminished fifth both span six steps in the traditional system, but nine and ten in the 19-position system. The size can even be zero or negative: a diminished second spans no steps in the traditional system.

In a regular system the size of an interval is the same from any starting name. In other systems it depends on where the interval starts, so a start name is required. Some intervals have no whole size in a system: the neutral third in the traditional system lies between two steps.

From a size the way back is possible too: the interval of a given letter size that spans a given number of steps, every interval that spans a number of steps, intervals enharmonically equal to a given one, and all spellings of a pair of pitches at the same positions with their intervals.

## Short notation and names

Every interval has a short notation for storage and exchange: an optional minus for a descending interval, a sign of quality, and the size in digits. P is perfect, M major, m minor, n neutral; A and d, written once, twice, or three times, are augmented and diminished of that multiplicity. So a minor sixth down is "-m6" and a doubly diminished fifth "dd5". Larger multiplicities and fractional qualities write the exact deviation in braces: "{+1/4}3" is a third a quarter of a unit wider than major. Reading also accepts other common signs, such as "maj3", "+4", and "°5". "TT" is not read: the word tritone names two different intervals.

Intervals also have names in English and Russian and a Russian abbreviation: "major third", «большая терция», «б.3». Only names established in the sources are given: sizes up to the fifteenth, multiplicities up to triple, and the neutral intervals that have traditional names. Other intervals exist and have a short notation, but have no name.

Intervals are values: they never change, and they are safe for concurrent use.

First released in v0.27.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/interval) · [Go examples](../../interval/example_test.go) · [Roadmap](../../ROADMAP.md)
