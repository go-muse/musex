# Parts written in measures

A `track.Track` is one part assigned to a particular live instrument. It has a finite [measure grid](measure.md), at least one voice of notation, optional fermata signs at barlines, and an optional measured [tempo map](tempomap.md). Voices share the part's grid and run in parallel. Their opaque `VoiceID` keys remain valid in derived snapshots where those voices exist.

The assigned instrument and the settings source of each note are independent. An unbound note stays unbound; notes using another instrument keep that source. Retuning the assigned instrument changes only notes that actually use it. The instrument's reference range does not restrict notation.

## Read and edit the record

`Entries` returns the original notes, chords, rests, grace signs, and duration forms in one bar. `At` returns a positive event in its half-open musical interval, including a rest; an empty remainder or bar has no event. Grace notes take zero musical time and remain visible through `Entries`. `Sounds` instead joins tied note occurrences, including partial ties of chords and ties across barlines. Each sound lists all its original occurrences.

Events in one bar follow one another from its beginning. Write an explicit rest before a later entrance in that bar. Unrecorded full bars and trailing space stay empty. A rest in measures uses the actual lengths of its covered bars and requires one continuous meter run.

`Splice` inserts, replaces, or removes records. `Edit` replaces several bar cells, the grid, the clock, and barline signs atomically. In particular, write both sides of a new tie in one edit. Invalid edits return `*track.Problems` with locations and causes; use `errors.Is` for classifications. Every edit preserves the original snapshot.

## Layout and transformations

`WriteEvents` lays out a sequence from a measure address. `Vacant` rejects occupied time; `ReplaceVoice` explicitly replaces the selected voice. A `Proposal` contains the complete new `Track` plus before/after entries and input indexes. Accept it by choosing `proposal.Track`.

Crossing notes and chords are divided at barlines with ties between all newly created parts. The attack signs stay on the first part, and the original outgoing ties and final fermata on the last. Exact notation is required; `KeepGroups` retains tuplet context and `WithoutGroups` removes it. An unrepresentable part is an error, never an approximation.

`Split` and `Merge` change compatible records within a bar. `PackRests` and `UnpackRest` convert explicit bar rests to and from measured signs without losing intermediate fermatas. `Rebar` preserves absolute musical positions on a new finite grid; an ordinary grid edit instead preserves bar numbers and offsets. `Quantize` rounds one voice's shared boundaries with an explicit step and rounding mode, preserving actual barlines and refusing collapsed notes, crossings, or detached scheduled fermatas.

## Extract a melodic line

```go
piano := instrument.MustNew(instrument.Piano())
grid := measure.MustNew(meter.MustNew(4, 4)).MustWithBarCount(3)
part := track.MustNew(piano, grid)
upper := part.MustVoices()[0]
quarter := event.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 0))
e4 := note.MustParsePitch(piano, "E4")
e := event.MustNew(event.MustNoteContent(e4), quarter, event.Marks{})
part = part.MustSplice(upper, 2, 0, 0, e)

out, err := part.ExtractLine(upper, track.Exact)
if err != nil {
    return err
}
line := out.Line // Rest of 1, then E4 on [1, 5/4).
_ = line
```

Leading and internal unrecorded gaps become exact rests; the empty grid tail is omitted. Measured rests become rests of their actual length. The report distinguishes gaps, measured records, skipped grace notes, barline signs, and scheduled extensions. `PrincipalOnly` explicitly skips grace records. Chords, even one-note chords, and notes without octaves cannot become a melodic line.

A complete tie becomes one sound with the first note's source. Writing that sound back uses the first source for every expanded part. Later links' distinct sources remain accessible in the original part but cannot be recovered from the line alone. Line names and roles are independent of the part's voice keys.

## Write a line into a part

```go
piano := instrument.MustNew(instrument.Piano())
part := track.MustNew(piano, measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(2))
upper := part.MustVoices()[0]
c4 := note.MustParsePitch(piano, "C4")
half := voice.MustWrittenDuration(duration.MustNewWritten(duration.Half, 0))
line := voice.MustNew(voice.MustNewSound(c4, half))

start := measure.MustNewAddress(1, duration.MustNewPosition(1, 2))
proposal, err := part.WriteLine(upper, start, line, track.WriteOptions{})
if err != nil {
    return err
}
part = proposal.Track
```

The first bar contains a half rest followed by a tied quarter C4; the second starts with the continuing quarter. Extracting again gives the rest and one C4 sound on [1/2, 1).

## Seconds and fermatas

A measured `Clock.Map` schedules tempo changes and extensions inside the grid. `Clock.BarHolds` schedules additional delays after specific bars, including the final bar. A hold stays attached to that bar if the grid grows. Fermata signs alone add no seconds.

`PerformanceMap` resolves the schedule into a positional map. `SoundSeconds` uses that map's `End` at both attack and release, so a shared hold delays the next attack and sustains ending notes through the hold. The caller selects decimal places and rounding; approximate results retain their error bounds. With quarter = 60, a whole note and a one-second final hold end at five seconds while the grid still ends at musical position 1.

Snapshots are safe for concurrent reads and derived edits. Returned containers are independent; immutable events and the live instrument are intentionally shared. `nil` and zero tracks are unset. Error-returning public operations have `Must` wrappers that panic with the original error.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/track) · [Roadmap](../../ROADMAP.md)
