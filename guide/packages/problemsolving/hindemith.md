# Hindemith analysis

`problemsolving/hindemith` reads dyads, voiced chords, and chord progressions
through Paul Hindemith's harmonic theory: Series II, interval roots, the table
of chord groups, and harmonic fluctuation. Every operation takes an explicit
system with twelve classes and chromatic unit 1; the model assumes twelve-tone
equal temperament and reduces compound intervals to their simple part. The
result is an analysis under the stated rules, not a claim about the composer's
intention or the heard center.

## Dyads: Series II position and root

`AnalyzeDyad(sys, from, to)` returns a `Dyad` with the Series II position
(1 to 13: unison, octave, fifth, fourth, major third, minor sixth, minor
third, major sixth, major second, minor seventh, minor second, major seventh,
tritone), the tempered interval name, and the root: the lower tone of
unisons, octaves, fifths, thirds, and sevenths, the upper tone of fourths,
sixths, and seconds. The tritone has no root. `From` and `To` return the
arguments in call order; `Root` is one of the two notes.

```go
sys := system.Traditional()
d := hindemith.MustAnalyzeDyad(sys, c4, g4)
// d.Series() == 3, d.Name() == "perfect fifth", d.Root() equals c4.
t := hindemith.MustAnalyzeDyad(sys, c4, fSharp4)
// t.Series() == 13, t.IsTritone(), !t.HasRoot().
```

## Chords: group, subgroup, bass, root, evidence

`AnalyzeChord(sys, notes)` classifies a voiced chord of at least two notes
with octaves. `Group()` returns I to VI, `Subgroup()` the subdivision
(`1`, `2`, `a`, `b1`, `b2`, `b3`, or none for V and VI), `Bass()` the lowest
note, `Root()` the chord root when determinate, and `Dyads()` every
intra-chord pair with `Best()` the pair whose root is the chord root.

- Without a tritone: three different tones a major third apart (augmented
  triad) or a perfect fourth apart (quartal triad), in any voicing and with
  doublings, are group V without a root; otherwise absence of seconds and
  sevenths gives I and their presence III.
- With a tritone: content of only tritones, minor thirds, and major sixths
  (diminished triad, diminished seventh) is VI without a root; otherwise
  absence of minor seconds and major sevenths gives II and their presence IV.
- In the other groups the root is the root of the strongest interval between
  different tones. Unison and octave pairs are doublings and carry no root
  information; ties between equal best intervals are broken by the lower
  pair, then the upper, so the listing order never changes the result.
  Tritones count as different pairs of pitch classes, so a doubled tritone
  member is one tritone.
- Subgroups 1 and 2 mean root in the bass and root above the bass. Group II
  adds `a` (one tritone, no major second, root in the bass), `b1` and `b2`
  (one tritone, root in the bass or above it), and `b3` (more than one
  tritone).

```go
c := hindemith.MustAnalyzeChord(sys, []note.Note{c4, e4, g4})
// c.Group() == hindemith.GroupI, c.Subgroup() == hindemith.Sub1,
// c.Root() equals c4, c.Bass() equals c4.
inv := hindemith.MustAnalyzeChord(sys, []note.Note{e4, g4, c5})
// inv.Group() == hindemith.GroupI, inv.Subgroup() == hindemith.Sub2,
// inv.Root() equals c5.
v := hindemith.MustAnalyzeChord(sys, []note.Note{c4, f4, bFlat4})
// v.Group() == hindemith.GroupV, !v.HasRoot(): the quartal triad.
```

Doublings never decide the root; they can move the bass or add a second or
seventh between the doubled note and its neighbours, and the subgroup then
follows the new content.

## Progressions: harmonic fluctuation

`AnalyzeProgression(sys, chords)` classifies every chord and reports a `Step`
per adjacent pair: `From()`, `To()`, `Delta()`, and `Direction()`
(`Rise`, `Fall`, `Level`). The fluctuation order is I.1, I.2, II.a, II.b1,
II.b2, II.b3, III.1, III.2, IV.1, IV.2, V, VI; `Chord.Value()` returns the
position, 1 to 12. A step to a higher value raises harmonic tension.

## Errors and equality

Errors are sentinels matched with `errors.Is`: `ErrUnset`, `ErrNoOctave`,
`ErrUnsupportedSystem`, `ErrTooFewNotes`, `ErrInadmissible` (wrapping
`system.ErrInadmissible`), and `ErrOutOfRange` (wrapping
`system.ErrOutOfRange`), in the documented order per operation. Every checked
operation has a `Must` variant that panics with the same error. `Dyad.Equal`
and `Chord.Equal` compare the analyzed values; two unset values are not
equal, as for `note.Note.Equal`.

Sources: Hindemith, *The Craft of Musical Composition*, Book I, pp. 68-106;
the chord-group table reproduction at the UNT digital library; the group
summary at https://ux1.eiu.edu/~pdhesterman/old/analysis/hindemith.html; and
the ontocomc formalization (CEUR Vol-4091, paper 16).