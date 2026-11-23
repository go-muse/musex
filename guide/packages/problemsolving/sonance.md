# Sonance

`problemsolving/sonance` scores the tonal consonance of a pitch-class set or
of the actual pitch pairs of a voicing: the sum, over the counted intervals,
of the number of pairs times a consonance weight. A higher score means a more
consonant inventory under the chosen profile; a negative score means a
dissonant one. The score describes interval content under stated weights, not
heard pleasantness: timbre, register, and familiarity are outside the model.

## Profiles

`Profile` carries seven weights, one per interval class 0 to 6, for
pitch-class sets. `HuronProfile()` returns the published aggregate dyadic
consonance weights of Huron (1994): 0, -1.428, -0.582, +0.594, +0.386, +1.240,
-0.453. `NewProfile(weights)` accepts any seven finite weights.

`PairProfile` carries twelve weights, one per tempered distance modulo 12
(0 to 11 semitones), for pairs of actual pitches. Because it keeps an interval
apart from its inversion, a pair profile can distinguish a fourth from a
fifth. `HuronPairProfile()` spreads the published class weights over both
members of each class; `NewPairProfile(weights)` accepts any twelve finite
weights. Both profiles are immutable copies with `Weights`, `Weight`,
`IsValid`, and `Equal`; the zero value is unset.

## Scores and contributions

`ScoreSet(set, profile)` counts the interval-class vector of a twelve-class
`pcset.Set`: one count per unordered pair of distinct pitch classes, so IC0
is always zero. `ScoreNotes(sys, notes, pairProfile)` counts every unordered
pair of note positions at its distance modulo 12, including unisons, octaves,
and doubled pitch classes; cents do not take part, and the system must have
twelve classes and chromatic unit 1.

`Score.Value()` is the sum; `Score.Contributions()` exposes one contribution
per interval (seven for a set score, twelve for a note-pair score) with
`Interval()`, `Count()`, `Weight()`, and `Value()`, including zero counts.

```go
s := sonance.MustScoreSet(pcset.MustNew(12, 0, 4, 7), sonance.HuronProfile())
// s.Value() == 2.220: one IC3, one IC4, one IC5.

weights := make([]float64, 12)
weights[5] = 1 // fourth
weights[7] = 2 // fifth
p := sonance.MustNewPairProfile(weights)
fourth := sonance.MustScoreNotes(sys, []note.Note{c4, f4}, p) // 1
fifth := sonance.MustScoreNotes(sys, []note.Note{c4, g4}, p)  // 2
```

## Errors

Sentinels matched with `errors.Is`: `ErrUnset`, `ErrInvalidArgument`,
`ErrNoOctave`, `ErrUnsupportedSystem`, `ErrInadmissible` (wrapping
`system.ErrInadmissible`), and `ErrOutOfRange` (wrapping
`system.ErrOutOfRange`), in the documented order per operation. Every checked
operation has a `Must` variant that panics with the same error.

Sources: Huron, "Interval-Class Content in Equally Tempered Pitch-Class Sets:
Common Scales Exhibit Optimum Tonal Consonance", *Music Perception* 11(3),
1994, doi:10.2307/40285624; the weights as quoted in Parncutt, "Consonance and
Dissonance of Simultaneous Trichords in Western Music" (2023),
https://hrcak.srce.hr/file/433486.