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

## Measure and reuse a chromatic shift

`Chromatic` records a signed number of system positions and the immutable
musical system used to measure it. It preserves register and ignores cents.
An unset value differs from a constructed zero shift.

```go
sys := system.Traditional()
c4 := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), nil)
eb4 := note.MustFromPitch(pitch.MustFromASPN(spelling.EFlat(), 4), cents.Zero(), nil)
shift := interval.MustChromaticBetweenNotesIn(sys, c4, eb4)
fmt.Println(shift.Steps()) // 3
```

`ChromaticBetweenNotes` instead reads the bound notes' settings. Each distinct
source is read once. Both forms require octaves. Compare measurements with
`SameShift`, which checks musical system equality; an equal step count in a
different system is not an equal shift. `TransposeNote` takes an explicit
`system.SpellingRule`, preserves cents and the settings source, and chooses a
spelling even when moving by zero steps. A bound note must use an equal system.
To measure a named interval, use `i.Chromatic(sys, start)`.

`MeasureNotes(from, to)` and `MeasureNotesIn(sys, from, to)`, with Must
variants, return both measurements at once: the written interval and the
signed chromatic displacement of two notes with octaves. The bound form reads
each distinct source once and uses the same snapshot for both results; the
explicit-system form reads no sources. Cents do not take part; read the generic
size from the written interval with `Generic()`. C4 to D#4 gives A2 and +3,
C4 to E-flat4 gives m3 and +3, and D#4 to E-flat4 gives an ascending diminished
second and 0.

## Choose quality or recover spelling

`Generic` retains the complete letter size without choosing quality. Its
`Variants` iterator lists unique qualities from narrow to wide. A zero
`Palette` is `DefaultPalette`: basic qualities, single augmentation/diminution
for the major-minor group, and double augmentation/diminution for the perfect
group. Unisons exclude negative deviations. `Through(k)` chooses a common
multiplicity limit; `Through(0)` keeps only basic qualities. `Deviations` selects
an exact finite set, including fractions; no arguments selects none.
`WithDeviations` adds values without changing the original palette.

```go
third := interval.MustNewGeneric(3)
for candidate := range third.MustVariants(interval.Up, interval.DefaultPalette()) {
    fmt.Println(candidate) // d3, m3, M3, A3
}
written := third.MustResolve(shift, spelling.C(), interval.Up) // m3
endpoint := interval.MustTransposePitch(sys, c4.MustPitch(), written) // E-flat 4
```

`FromPitch` returns interval/endpoint pairs and skips individual impossible
endpoints. Invalid starting inputs still return errors for an empty palette.
`Resolve` recovers the quality from a signed shift, complete letter size, and
written direction. For +3 from C, a second gives A2 and a third gives m3.
`ForShift` finds all matching sizes in the chosen palette; it bounds its search
by reachable octaves. These finite iterators support repeated and concurrent
reads and early stopping. In irregular systems, recovery requires the actual
starting spelling. Written direction and the sign of the positional shift are
independent: an ascending doubly diminished second may move down one position.

## Short notation and names

Every interval has a short notation for storage and exchange: an optional minus for a descending interval, a sign of quality, and the size in digits. P is perfect, M major, m minor, n neutral; A and d, written once, twice, or three times, are augmented and diminished of that multiplicity. So a minor sixth down is "-m6" and a doubly diminished fifth "dd5". Larger multiplicities and fractional qualities write the exact deviation in braces: "{+1/4}3" is a third a quarter of a unit wider than major. Reading also accepts other common signs, such as "maj3", "+4", and "°5". "TT" is not read: the word tritone names two different intervals.

Use explicit collections from `interval/naming` for user-facing names:

```go
entry := naming.Russian().Specific(interval.MinorThird().MustDown())
short, _ := entry.MustName()     // "м.3"
full, _ := entry.MustNameFull()  // "малая терция"
english, _ := naming.English().Specific(interval.MinorThird()).MustName() // "m3"
size, _ := naming.English().Generic(interval.MustNewGeneric(3)).MustName() // "third"
```

`Name` is short and `NameFull` is full; both return `(string, bool, error)`.
A valid but unavailable form returns `"", false, nil`. `ParseGeneric` and
`ParseInterval` accept only their selected language, with no direction sign.
English short forms preserve the M/m distinction. Russian neutral thirds have
only a full name; English neutral 2, 3, 6, 7, and 9 have both forms. Other
fractional qualities have neither form. Sizes above 15 have no dictionary name.

`Catalog.Entries()` (and `MustEntries`) enumerates the whole dictionary in a
fixed order: the generic sizes 1 to 15, then the named specific intervals in
ascending size, within a size perfect, major, minor, neutral, then augmented
and diminished in multiplicity order 1 to 3. Every entry has a full name; the
short form is present exactly when the language gives one, so the Russian
neutral third is discoverable without a short form. `Entry.IsGeneric`,
`Entry.Generic`, and `Entry.Interval` read the subject of an entry; the
enumeration round-trips through `ParseGeneric` and `ParseInterval`.
Use `interval.Parse` for directed canonical notation such as `-m3`.

The historical `Interval.Name` still returns the full English name;
`RussianName`, `RussianAbbreviation`, and `ParseName` also retain their released
behavior. Legacy forms include: "major third", «большая терция», «б.3». Only names established in the sources are given: sizes up to the fifteenth, multiplicities up to triple, and the neutral intervals that have traditional names. Other intervals exist and have a short notation, but have no name.

Intervals are values: they never change, and they are safe for concurrent use.

First released in v0.27.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/interval) · [Go examples](../../interval/example_test.go) · [Roadmap](../../ROADMAP.md)
