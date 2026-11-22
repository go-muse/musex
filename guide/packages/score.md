# An ensemble on one timeline

A `score.Score` combines ordered [tracks](track.md) with a common musical origin and one positional [tempo map](tempomap.md). Each participation keeps its own finite measure grid and voice keys. Repeatedly adding the same track creates distinct `PartID` keys and independent sounding contributions. `Equal` compares the ordered musical records and common map, ignoring these keys.

```go
piano := instrument.MustNew(instrument.Piano())
grid := measure.MustNew(meter.MustNew(4, 4)).MustWithBarCount(2)
part := track.MustNew(piano, grid)
clock := tempomap.MustNew(tempo.MustNewQuarter(fraction.MustNew(60, 1)))
ensemble := score.MustNew(clock)
ensemble, first := ensemble.MustAdd(part)
ensemble, second := ensemble.MustAdd(part)
_ = second
independentPart := ensemble.MustPart(first)
_ = independentPart
```

`Add` and `Replace` require a track without a clock. `AddWithoutClock` and `ReplaceWithoutClock` explicitly discard a local clock and return the entire discarded schedule. To edit a part, obtain it with `Part`, derive a new track, and call `Replace`. A replacement must preserve the starts of all bar numbers present in both grids; adding or removing a grid tail is otherwise allowed. `Remove` leaves the common map intact. `Reorder` takes an exact permutation of the existing keys.

## Written and sounding slices

`At` distinguishes positive written events, explicit rests, empty record space, and the end of a shorter part's grid. `Sounds` joins ties through Track and preserves all original occurrences. `SoundingAt` selects those sustained notes in half-open intervals. Grace records remain available through `Part(...).Entries(...)` and do not create a positive interval in `At`.

Selections identify a participation and optionally its voice keys. Omitted selections mean the whole ensemble; an empty voice list means all voices of that participation. Results follow score and track order rather than selection order. Unknown keys and duplicate selections are errors, including at the global end.

`SoundSeconds` uses the common map's `End` side for both attack and release. A common hold is counted once; written fermatas add no seconds on their own. The caller chooses precision and rounding explicitly. Exact and approximate results preserve the tempo map's bounds. `Map` also exposes position-to-seconds and inverse conversion beyond the recorded ensemble.

## Extract the full common map

`Part` returns a clockless track. `ExtractWithMap` copies every common-map record to that part's grid, preserving kinds, slots, laws, and values. Tempo changes must be strictly before its end; extensions may also occur exactly at the final barline, where they become a terminal `BarHold`. A record beyond the finite grid causes an addressed error. The operation never clips the map, adds synthetic records, or extends notation. A valid unfinished gradual tail stays unfinished.

## Change meters and compare barlines

`ChangeMeter(first, last, meter)` proposes a strict change in every part. Notes stay in their original bars and offsets; the method does not call `Rebar` or move overflowing notes into the next bar. It checks every part and binds every common-map mark to each old grid before finding its new position. Different candidate positions or missing bindings reject the whole proposal. Inspect `*score.Problems` for original marks, affected participations, candidate positions, and underlying track issues. Accept a successful proposal by choosing `proposal.Score`; the source is unchanged.

`CommonBarlines` returns real common boundaries in a closed interval for two or more selected parts. It includes initial and final boundaries and identifies the neighboring bar numbers for every part. It does not continue a short grid beyond its end. `BarlinePeriod` computes the exact mathematical period for two constant grids of full bars without pickups, even when the period extends beyond their finite records. Variable or incomplete bars give `ErrPeriod`.

`BarView` displays one requested bar. A multibar rest appears as one full-bar rest in each covered bar, with its original final fermata only in the last one. The source entry retains the full original span. This display never rewrites the stored multibar sign or fills an empty voice.

## Ownership and errors

Nil and zero scores are unset; `New` creates a set empty score ending at the origin. Snapshots support concurrent reads and independent derived edits. Event and Track pointer shells returned by the API are detached from internal storage, including proposal reports. Live instruments and note settings sources keep their Track semantics. Assigning an entire returned event or track therefore cannot edit the ensemble.

Use `errors.Is` for error classes and dependency causes, and `errors.As` or `errors.AsType[*score.Problems]` for addressed diagnostics. Failed operations return zero results rather than partial ensembles or time lists. Error-returning operations have corresponding `Must` wrappers that panic with the original error.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/score) · [Roadmap](../../ROADMAP.md)
