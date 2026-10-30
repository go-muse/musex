# Roadmap

`musex` is planned as a set of focused Go packages across three areas. This roadmap shows the intended package order and scope. All packages are **planned**; none is implemented yet. Check a package off when its agreed API and behavior are implemented, reviewed, and accepted.

## 1. Required to get started

- [ ] **1. `letter`** — the seven note letters and their cyclic order.
- [ ] **2. `alteration`** — accidental values and chromatic changes.
- [ ] **3. `spelling`** — a note name formed from a letter and alteration.
- [ ] **4. `octave/number`** — octave numbering schemes and their visible numbers.
- [ ] **5. `octave/naming`** — catalogs of visible octave names.
- [ ] **6. `octave`** — octave identity and conversion to visible labels.
- [ ] **7. `pitch`** — a written note name in a positional octave.
- [ ] **8. `pitchclass`** — a position class within an octave.
- [ ] **9. `system`** — musical positions, steps, equivalence, and written-octave rules.
- [ ] **10. `cents`** — cent values, note offsets, and distances.
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

- [ ] **26. `common/fraction`** — fractions used to represent durations.
- [ ] **27. `tuplet`** — tuplets such as triplets.
- [ ] **28. `duration`** — musical durations.
- [ ] **29. `chord`** — chords built from notes and durations.
- [ ] **30. `track`** — tracks and their events.

Package order and scope are derived from the local research in `docs/package-roadmap.md`. That detailed source is excluded from Git; this file is the project-facing overview.
