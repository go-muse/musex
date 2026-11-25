# Voice leading

`problemsolving/voiceleading` describes how voices move: two or more existing
melodic lines against each other, the leaps of one voice, the spacing and
doublings of a chord, and the connection of two chords. It never changes the
music, never names a chord's root, quality, inversion, or function, and does
not judge artistic quality. Every educational norm is selected explicitly and
names its rule and source; without a norm a result is only a description.

In the examples below `sys` is `system.Traditional()`, notes such as `c3` or
`e4` are `note.Note` values with an octave, and `quarter` is a
`voice.Duration`.

## Lines against each other

`Analyze` builds an immutable snapshot over every pair of `voice.Line` inputs.
Each pair has its own grid of element starts and ends at the shorter line.
Verticals keep the written interval with its direction and full size, the
signed chromatic distance in positions of the system, and the original
elements; transitions classify motion. The searches find motion kinds, strict
perfect parallels, consecutive perfect families, direct arrivals, crossings,
and overlaps; crossings and overlaps require the `LowToHigh` order.
`CheckOMTFirstSpecies` explicitly checks one Open Music Theory first-species
rule for one pair. See [Analyze voice leading](../../../README.md#analyze-voice-leading)
for a worked example.

## Leaps in one voice

`NewMelody(line, sys)` measures one line; `Analysis.Melody(part)` measures a
part of an analysis in the system of the analysis, without reading sources.
`Moves()` lists every move between two consecutive sounds: both elements,
the written interval with its direction, the chromatic distance, and
`ThroughRest` for a move across a rest. A perfect unison is a repeat, a second
of any quality is a step, and a leap depends on the definition: from a third
(`LeapFromThird`, Wikipedia) or from a fourth (`LeapFromFourth`, Alekseev and
Myasoedov).

```go
m := voiceleading.MustNewMelody(line, sys) // C4 E4 G4 C5 B4
for _, mv := range m.MustMoves() {
    fmt.Println(mv.Interval, mv.Chromatic, mv.IsLeap(voiceleading.LeapFromThird))
}
// M3 4 true, m3 3 true, P4 5 true, -m2 -1 false
fmt.Println(m.MustLeapRuns(voiceleading.LeapFromThird, 2))  // [{0 2}]
fmt.Println(m.MustLeapRuns(voiceleading.LeapFromFourth, 2)) // []
```

`LeapRuns` reports maximal runs of leaps inside a fragment; a rest breaks a
run, and nested parts of a run are not reported. `Check` applies one melodic
norm to every move that is not through a rest:

| Rule | Requirement |
|---|---|
| `OMTLeapRecovery` | after a fourth or wider, a step in the opposite direction in the same fragment; the last leap is reported with `end-of-line` or `rest-follows` |
| `OMTMelodicIntervals` | no augmented, diminished, seventh, sixth other than an ascending minor sixth, or wider than an octave |
| `FuxAllowedMotions` | only P1, P4, P5, P8, minor and major seconds and thirds, and an ascending minor sixth followed by downward motion |
| `MT21ChoraleMelody` | no augmented intervals; after a leap wider than a fourth or a diminished interval, the next move changes direction |

```go
a := voiceleading.MustNewMelody(cab, sys).MustCheck(voiceleading.OMTLeapRecovery) // C4 A4 B4
fmt.Println(a.Rule, a.Violations[0].Index, a.Violations[0].Reasons)
// omt-leap-recovery 0 [no-step-back]
```

## Sonorities and spacing

`NewSonority(chord, sys)` names the notes of a `chord.Chord` by voice from the
bass upward in the order of record; the notes are never sorted. The sonorities
of all parts of an analysis come from `Analysis.Sonorities` and `SonorityAt`:
one per element start of any part before the joint end of all parts, with the
element of every part and the sounding voices in input-slot order.
`Analysis.Connections` and `ConnectionAt` connect adjacent moments in which
every part sounds.

`Spacing()` gives the intervals between adjacent voices, the span from the
bass to the top voice, and whether the span stays within a perfect octave by
writing: B♯3 to C5 is a diminished ninth of 12 positions, not within the
octave. `ClassifySpacing` applies one definition; a fourth or narrower is
narrow and a fifth or wider is wide.

| Sonority | `SpacingME` | `SpacingAM` | `SpacingOMT` |
|---|---|---|---|
| C3 E4 G4 C5 | close | mixed | omt-open |
| C4 E4 G4 C5 | close | close | omt-closed |
| C3 C4 G4 E5 | open | open | omt-open |
| C4 E4 G4 | voice-count | close | omt-closed |

`SpacingME` considers tenor-alto and alto-soprano of exactly four voices,
`SpacingAM` every adjacent pair including the bass, and `SpacingOMT` the span.
Crossed voices, a wrong number of voices, and silent parts make a definition
inapplicable; that is a result with a reason, not an error.

`CheckDistances` applies `DistanceOMTSATB`, `DistanceMT21SATB`,
`DistanceOMTFirstSpecies`, or the composer's limits with `DistanceCustom`.
A pair exceeds a limit when its size is larger, or the size is equal and the
quality larger: P8 does not exceed P8, an augmented octave or a diminished
ninth does.

```go
s := voiceleading.MustNewSonority(chord.MustNew(c3, g3, c4, e5), sys)
d := s.MustCheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB})
fmt.Println(d.Violations[0].Pair.Interval) // M10, soprano-alto
```

## Doublings and chord members

`Doublings(reading, expected)` lists the tone classes of the voices, the voices
that take each class, and the kind of doubling: unison, octave, or both.
`BySpelling` reads written names, so C and B♯ differ; `BySound` reads the
pitch class in the system, so they coincide in the traditional system. With
an expected chord, the result adds missing and extra classes and whether the
chord is complete; octaves of the expected notes are ignored.

```go
s := voiceleading.MustNewSonority(chord.MustNew(c3, c4, e4, c5), sys)
d := s.MustDoublings(voiceleading.BySpelling, chord.MustNew(cName, eName, gName))
// C is taken by voices 0, 1, 3 (octave); G is missing; the bass is doubled.
fmt.Println(d.BassDoubled, d.Missing[0].Name, d.Complete) // true G false
```

Norms of doubling by harmonic function, such as the leading tone, need a
harmonic and tonal analysis and are not part of this package.

## Decomposing chords into lines

`Decompose` turns a sequence of `Simultaneity` values (a chord and a duration;
an empty chord is a rest in every voice) into lines from the bass upward: the
k-th line takes the k-th lowest note of every chord. It assumes that voices do
not cross and marks every chord where two voices have equal positions, since
membership does not follow from the notes there. Every chord attacks anew.
`DecomposeInOrder` keeps the order of record instead and needs no system.

```go
dec := voiceleading.MustDecompose([]voiceleading.Simultaneity{
    {Notes: chord.MustNew(c3, e4, g4, c5), Duration: quarter},
    {Notes: chord.MustNew(g2, d4, g4, b4), Duration: quarter},
}, sys)
// Bass C3 G2, tenor E4 D4, alto G4 G4, soprano C5 B4.
fifths := voiceleading.MustAnalyze(dec.Lines, voiceleading.Options{
    System: sys, Order: voiceleading.LowToHigh,
}).MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth})
```

## Connecting chords

`Connect` describes the connection of two sonorities with the same number of
sounding voices: the move of every voice, the classes common to both, the
voices that keep a common tone, and the total chromatic movement. A connection
is harmonic when a voice keeps a common tone and melodic otherwise.

`CheckMT21Connection` applies the MT21 norm for root-position triads with a
doubled bass; selecting it asserts that both chords are such triads. For a
bass moving by a fourth or fifth, one variant keeps the common tone in the
same upper voice and moves the two others by step, the other moves all three
upper voices in one direction; for a bass moving by a third or sixth, two
common tones stay and the third upper voice steps.

`ChooseConnections` lists every choice of upper voices toward a target chord
and a given bass that passes the selected conditions: completeness, common
tones, no crossings, no overlaps, a distance norm, no parallels, a doubled
bass, and the MT21 norm. Each upper voice may take every target class at the
nearest pitch not below its note or the nearest strictly below. The result is
ordered by total movement and equals a full enumeration; the number of
combinations grows exponentially with the number of voices.

```go
from := voiceleading.MustNewSonority(chord.MustNew(c3, e4, g4, c5), sys)
to := voiceleading.MustNewSonority(chord.MustNew(f3, f4, a4, c5), sys)
c := voiceleading.MustConnect(from, to)
fmt.Println(c.Kind, c.Total, c.Common) // harmonic 8 [C]
fmt.Println(voiceleading.MustCheckMT21Connection(from, to).Conforms) // true

choices := voiceleading.MustChooseConnections(from, chord.MustNew(fName, aName, cName), f3,
    voiceleading.ConnectionConditions{
        Complete: true, CommonTones: true, NoCrossings: true,
        Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth},
    })
// F3 F4 A4 C5 (8), F3 C4 A4 C5 (11), F3 A4 A4 C5 (12), F3 A3 F4 C5 (14), ...
```

The choices are possibilities under stated conditions, not a harmonization or
an artistic recommendation.

## Errors, equality, and ownership

Sentinels are matched with `errors.Is`; `PartError` and `ChordError` locate a
part element or a chord note, and every checked operation has a `Must` variant
that panics with the same error. Analyses, melodies, and sonorities are
immutable snapshots, safe to copy and to read concurrently, and compared with
`Equal`. `Analysis.Equal` compares pairwise verticals and transitions only, so
two equal analyses can differ in the tail of a part, which `Melody.Equal`
compares. Every returned slice is independent; returned notes keep their source
binding.

Sources: Open Music Theory; Wikipedia, Counterpoint; Music Theory for the
21st-Century Classroom, 26.5-26.7; Alekseev and Myasoedov, Elementary Music
Theory; Muzykalnaya entsiklopediya, Akkord.
