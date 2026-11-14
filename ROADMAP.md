# Roadmap

`musex` is planned as a set of focused Go packages across three areas. This roadmap shows the intended package order and scope. Checked packages are implemented; the others are **planned**. Open a checked package to read about its musical model or follow its links to the Go API and working examples. Mark a package complete when its agreed API and behavior have been implemented, reviewed, and accepted.

<!-- When a package is implemented, link its checked item to guide/packages/<package path>.md. Keep its musical model and release status in that chapter, and its technical examples in the package's example tests. -->

## 1. Required to get started

- [x] **1. [`letter`](guide/packages/letter.md)** — the seven note letters and their cyclic order.
- [x] **2. [`alteration`](guide/packages/alteration.md)** — accidental values and chromatic changes.
- [x] **3. [`spelling`](guide/packages/spelling.md)** — a note name formed from a letter and alteration.
- [x] **4. [`octave/number`](guide/packages/octave/number.md)** — octave numbering schemes and their visible numbers.
- [x] **5. [`octave/naming`](guide/packages/octave/naming.md)** — catalogs of visible octave names.
- [x] **6. [`octave`](guide/packages/octave.md)** — octave identity and conversion to visible labels.
- [x] **7. [`pitch`](guide/packages/pitch.md)** — a written note name in a positional octave.
- [x] **8. [`pitchclass`](guide/packages/pitchclass.md)** — a position class within an octave.
- [x] **8a. [`pcset`](guide/packages/pcset.md)** — sets of pitch classes and set operations.
- [x] **8b. [`setclass/catalog`](guide/packages/setclass-catalog.md)** — Forte and Carter set-class catalogs for 12 positions.
- [x] **8c. [`setclass`](guide/packages/setclass.md)** — set classes, prime forms, and relations between them.
- [x] **8d. [`pcseq`](guide/packages/pcseq.md)** — pitch-class sequences and twelve-tone rows.
- [x] **9. [`system`](guide/packages/system.md)** — musical positions, steps, equivalence, and written-octave rules.
- [x] **10. [`cents`](guide/packages/cents.md)** — cent values, note offsets, and distances.
- [x] **11. [`tuning/temperament`](guide/packages/tuning/temperament.md)** — temperament tables expressed in cents.
- [x] **12. [`tuning`](guide/packages/tuning.md)** — a temperament, reference pitch, and frequency calculations.
- [x] **13. [`notation`](guide/packages/notation.md)** — parsing and formatting note names and pitches.
- [x] **14. [`settings`](guide/packages/settings.md)** — a compatible set of system, tuning, and notation settings.
- [ ] **15. `instrument`** — instrument settings, presets, and pitch range.
- [x] **16. [`note`](guide/packages/note.md)** — a note connected to its musical settings.
- [ ] **17. `midi`** — conversion between pitches and MIDI note numbers.

## 2. Important functionality

- [ ] **18. `interval`** — named intervals, distance, and transposition.
- [ ] **19. `degree`** — scale degrees and their names.
- [ ] **20. `template`** — named and positional mode templates.
- [ ] **21. `mode`** — mode membership, ordering, and transposition.
- [ ] **22. `scale`** — scales and their multi-octave expansion.
- [ ] **23. `search`** — finding modes from sets of notes.
- [ ] **24. `modal`** — relative and absolute modal positions.
- [ ] **25. `tendency`** — tendencies and resolutions under rule sets.

## 3. Long-term plans

- [x] **26. [`common/fraction`](guide/packages/common/fraction.md)** — exact signed fractions for musical quantities. Preserve the supplied numerator and denominator, compare by numeric value or by the supplied pair, and support arithmetic with explicitly chosen result forms. Review consumers individually, starting with fractional alterations and covering position cycles, systems, cents, tuplets, durations, and track timing.
- [x] **27. [`tuplet`](guide/packages/tuplet.md)** — tuplets such as duplets, triplets, and quintuplets. Keep the written pair of actual and normal counts, compare by notation or by coefficient, apply exact coefficients to durations, nest tuplets, and derive default pairs for binary or ternary division.
- [x] **28. [`duration`](guide/packages/duration.md)** — exact durations in fractions of a whole note written with note values from maxima to 1024th, dots, tuplets, and ties, positions as offsets from a chosen origin, notation of a duration as one note or tied notes, normal subdivision, and explicit quantization to a grid.
- [ ] **28a. `meter`** — time signatures with their beats, beat units, groupings, and strong beats.
- [ ] **28b. `measure`** — measure grids with pickups and meter changes, and positions given as a measure and an offset.
- [ ] **28c. `tempo`** — tempo markings as a beat unit and BPM, and durations in seconds.
- [ ] **28d. `tempomap`** — tempo changes and fermata holds over time, and conversion between seconds, positions, and measures.
- [ ] **28e. `ticks`** — MusicXML divisions, MIDI ticks, and MIDI tempo for durations, positions, and tempo maps.
- [ ] **29. `chord`** — chords as groups of notes that sound together, without durations.
- [ ] **29a. `event`** — voice elements: a note, chord, or rest with its duration, ties, fermatas, arpeggios, and grace notes.
- [ ] **29b. `voice`** — sequences of events with positions, ties, grace notes, and sounding notes in time.
- [ ] **30. `track`** — parts as sequences of measures holding the events of each voice.
- [ ] **31. `score`** — tracks with a common start and one shared tempo map.

Package order and scope are derived from the local research in `docs/package-roadmap.md`. That detailed source is excluded from Git; this file is the project-facing overview.
