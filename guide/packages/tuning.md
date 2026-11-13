# Tuning

A tuning assigns a frequency to every written pitch by combining four musical choices: a musical system, a temperament, the class from which that temperament is laid out, and a reference. The reference is a written pitch together with its frequency in hertz. A name without an octave has no single frequency, because its pitch classes differ by a factor of two from one octave to the next.

The reference note is step zero. The next higher position is step one, and the previous position is step minus one. Enharmonic spellings of the same position have the same step, exact cent height, and frequency. Heights and distances are measured in exact cents; converting a height to a frequency applies that exact value once to the reference frequency. A signed cent shift can be added before this conversion without changing the written pitch.

## Level, root, and comparison

Moving the reference to another note at the frequency already assigned to that note preserves the pitch level. Replacing only the reference frequency scales every assigned frequency and changes the level. Two levels can be compared by a signed cent difference or by a frequency ratio, even when their reference notes differ.

Changing the temperament root can change the frequencies of some classes while preserving the class of the reference note. In an equal temperament it changes no frequencies. The classes affected by a root change can be listed explicitly.

Strict sound comparison is a different question from level comparison. It asks whether two tunings compute exactly the same numeric frequencies, with no tolerance. Reexpressing a reference can therefore produce mathematically equivalent frequencies that differ in their last floating-point bits and fail this strict test. Compare levels when domain-level equivalence is intended; do not confuse either check with equality of all four constituent parts.

## Finding a pitch from a frequency

Given a measured frequency, a tuner selects one of the two adjacent steps and reports the signed deviation in cents: positive means the measurement is above the selected pitch. The default chooses the nearest step and resolves an exact midpoint downward. All nine shared rounding rules are available, including directed rounding and midpoint rules relative to zero or step parity. Producing a written pitch from the selected step always requires an explicit spelling rule.

## Standard tuning and ready-made references

The standard tuning combines the traditional twelve-position system, twelve-tone equal temperament, root C, and A4 at 440 Hz. The reference standard does not by itself prescribe this temperament.

Eight ready-made references are available:

- **A4 = 440 Hz** — the current international reference standard.
- **A4 = 435 Hz** — the former international reference standard in Vienna, preceding the 440 Hz standard.
- **A4 = 442 Hz** — the lower end of the 442–444 Hz range reflecting a modern tendency toward a higher orchestra pitch; not a standard.
- **A4 = 444 Hz** — the upper end of that range, with the same attribution; not a standard.
- **A4 = 453 Hz** — a historical tuning-fork frequency reached at times in the Paris and Vienna opera houses; not a standard.
- **A4 = 466 Hz** — another historical tuning-fork frequency reached at times in those opera houses; not a standard.
- **C5 = 523 Hz** — an ordinary tuning fork in the second octave, used less often than A4 at 440 Hz; not a standard. Its source does not establish a temperament. In twelve-tone equal temperament with A4 at 440 Hz, C5 is approximately 523.2511 Hz, so this reference is not at the same level.
- **C0 = 16 Hz** — the scientific pitch standard.

First released in v0.21.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/tuning) · [Go examples](../../../tuning/example_test.go) · [Roadmap](../../../ROADMAP.md)
