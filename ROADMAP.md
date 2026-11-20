# Roadmap

`musex` is planned as a set of focused Go packages across four areas. This roadmap shows the intended package order and scope. Checked packages are implemented; the others are **planned**. Open a checked package to read about its musical model or follow its links to the Go API and working examples. Mark a package complete when its agreed API and behavior have been implemented, reviewed, and accepted.

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
- [x] **15. [`instrument`](guide/packages/instrument.md)** — instrument settings, presets, and pitch range.
- [x] **16. [`note`](guide/packages/note.md)** — a note connected to its musical settings.
- [x] **17. [`midi`](guide/packages/midi.md)** — MIDI note numbers, positional octaves and classes, checked shifts, and conversion to and from written pitches and notes in the traditional C layout with explicit spelling rules.

## 2. Important functionality

- [x] **18. [`interval`](guide/packages/interval.md)** — named intervals between names, written pitches, and notes with their direction, size, and quality, transposition, inversion, sums, the line of fifths, positional sizes in a musical system, and the short notation and names in English and Russian.
- [x] **19. [`degree`](guide/packages/degree.md)** — ordinal degrees with optional notes and independent copies, comparison, decimal and Roman number forms, circular transitions, shifts and distances, normalization and denormalization, and lazy enumeration.
- [x] **20. [`template`](guide/packages/template.md)** — closed and open mode templates with arbitrary cycles, virtual places, rotations, comparison, text notation, and immutable built-in and custom catalogs with name lookup, identification, and related-entry search.
- [x] **21. [`mode`](guide/packages/mode.md)** — closed and open modes with numbered degrees, notes and scales from a starting note, key signatures, structural comparison and editing, system positions, and catalog search.
- [x] **22. [`scale`](guide/packages/scale.md)** — finite note lists with circular traversal, snapshot iterators, comparisons, melody transformations, editing, and multi-octave expansion through modes.
- [x] **23. [`search`](guide/packages/search.md)** — finding mode-and-tonic hypotheses from sets of notes by exact spelling or pitch class, with containment and full match, tonic families, matching degrees, and stable groupings.
- [x] **24. [`modal`](guide/packages/modal.md)** — relative categories with full specific intervals; absolute quality counts, scores, comparisons, and equal-score groups for closed perfect-octave modes.
- [x] **25. [`problemsolving/tendency`](problemsolving/tendency/README.md)** — supports, roles, tendencies, classifications, and two-voice resolutions under explicit built-in or finite user rules, with coverage, explanations, filtering, and comparisons.

## 3. Long-term plans

- [x] **26. [`common/fraction`](guide/packages/common/fraction.md)** — exact signed fractions for musical quantities. Preserve the supplied numerator and denominator, compare by numeric value or by the supplied pair, and support arithmetic with explicitly chosen result forms. Review consumers individually, starting with fractional alterations and covering position cycles, systems, cents, tuplets, durations, and track timing.
- [x] **27. [`tuplet`](guide/packages/tuplet.md)** — tuplets such as duplets, triplets, and quintuplets. Keep the written pair of actual and normal counts, compare by notation or by coefficient, apply exact coefficients to durations, nest tuplets, and derive default pairs for binary or ternary division.
- [x] **28. [`duration`](guide/packages/duration.md)** — exact durations in fractions of a whole note written with note values from maxima to 1024th, dots, tuplets, and ties, positions as offsets from a chosen origin, notation of a duration as one note or tied notes, normal subdivision, and explicit quantization to a grid.
- [x] **28a. [`meter`](guide/packages/meter.md)** — time signatures in four forms — simple, symbol, compound numerator, and sum — with their size, units and beats, standard and explicit groupings, kinds in two traditions, beat division, strong beats, places in a measure, and changes of the record and the unit.
- [x] **28b. [`measure`](guide/packages/measure.md)** — measure grids with pickups, meter changes, explicit bar lengths, and open or finite numbers of bars; bar numbers, lengths, and starts, positions given as a measure and an offset, places in the meter, unit counts and the MEI tstamp, splitting durations at bar lines and inside bars, and changes of a grid by ranges of bars.
- [x] **28c. [`tempo/term`](guide/packages/tempo/term.md)** — tempo terms such as Largo and Allegro with their places, groups, reference ranges of beats per minute, and meanings, in eight built-in lists from the English and Russian Wikipedia and six metronome scales or in lists made by the user, and finding the terms that fit a number of beats.
- [x] **28c. [`tempo`](guide/packages/tempo.md)** — tempo markings as a beat unit and BPM, compared by record or by speed, exact seconds of durations and explicit rounding to system time, unit conversion and metric modulation, the formula of bars, units, BPM, and time, approximate markings and tempo ranges, the Maelzel metronome scale, and tempo indications that join a term with a marking.
- [x] **28d. [`tempomap`](guide/packages/tempomap.md)** — tempo changes and fermata holds over time, and conversion between seconds, positions, and measures.
- [ ] **28e. [`format/ticks`](guide/packages/format/ticks.md)** — MusicXML divisions, MEI ppq and MIDI ticks for durations, positions and measure addresses; MIDI tempo and tempo maps; MIDI time signatures and their changes; SMPTE time and MIDI clocks.
- [x] **29. [`chord`](guide/packages/chord.md)** — chords as groups of notes that sound together, without durations or harmony: notes in their order of record with repeats, membership and equality by record, changes of the notes in place with independent copies, joining, intersecting, and subtracting chords, sorting, and moving all the notes by octaves.
- [x] **29a. [`event`](guide/packages/event.md)** — voice elements with a note, chord, or rest, written or exact duration, rests in measures, and ties, fermatas, arpeggios, and grace notes.
- [x] **29c. [`voice`](guide/packages/voice.md)** — melodic lines: monophonic sounds and rests with durations, names, and roles; positions without measures and seconds from a tempo map; melodic intervals, range, ambitus, and tessitura against a singing voice or an instrument; doubling, exact imitation, inversion, transposition, and retrograde.
- [x] **30. [`track`](guide/packages/track.md)** — instrument-assigned parts with parallel voices in finite measure grids, atomic edits, tied sounds, explicit layout and transformations, measured clocks, and melodic-line conversion.
- [ ] **31. `score`** — tracks with a common start and one shared tempo map.

## 4. Problem Solving

- [x] **32. [`problemsolving/scalechoice`](guide/packages/problemsolving/scalechoice.md)** — choosing pitch collections and modal readings from notes, distinguishing alternatives with added notes, and comparing a replacement throughout a set or at one phrase position, including local written and chromatic motions.
- [x] **33. [`problemsolving/voiceleading`](README.md#analyze-voice-leading)** — first version (unreleased): pairwise verticals and transitions over melodic lines, written and chromatic measurements, motion, perfect parallels, consecutive perfect families, direct arrivals, crossings and overlaps, and an explicit OMT first-species consecutive-perfect-family assessment. Chord spacing, chord distribution, leap checking, and chord connections remain planned.

Package order and scope are derived from the local research in `docs/package-roadmap.md`. That detailed source is excluded from Git; this file is the project-facing overview.
