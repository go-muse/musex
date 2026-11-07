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
- [ ] **8d. `pcseq`** — pitch-class sequences and twelve-tone rows.
- [ ] **9. `system`** — musical positions, steps, equivalence, and written-octave rules.
- [x] **10. [`cents`](guide/packages/cents.md)** — cent values, note offsets, and distances.
- [ ] **11. `tuning/temperament`** — temperament tables expressed in cents.
- [ ] **12. `tuning`** — a temperament, reference pitch, and frequency calculations.
- [ ] **13. `notation`** — parsing and formatting note names and pitches.
- [ ] **14. `settings`** — a compatible set of system, tuning, and notation settings.
- [ ] **15. `instrument`** — instrument settings, presets, and pitch range.
- [ ] **16. `note`** — a note connected to its musical settings.
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
- [ ] **27. `tuplet`** — tuplets such as triplets.
- [ ] **28. `duration`** — musical durations.
- [ ] **29. `chord`** — chords built from notes and durations.
- [ ] **30. `track`** — tracks and their events.

Package order and scope are derived from the local research in `docs/package-roadmap.md`. That detailed source is excluded from Git; this file is the project-facing overview.
