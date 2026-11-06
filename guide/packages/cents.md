# Cents

A cent is one hundredth of an equal-tempered semitone, or 1/1200 of an octave. It measures a frequency ratio on a logarithmic scale, so multiplying frequencies corresponds to adding cents. The same cent value serves for a note offset, a temperament entry, and a distance between frequencies, independent of any musical system or tuning.

## Exact and approximate values

Frequency ratios whose prime factors do not exceed 31—a 31-prime limit, not an odd limit, so 37:32 is outside it—steps of any equal division of the octave, fractions of commas, and decimal offsets are kept and combined without rounding. A pure fifth 3:2 is kept exactly; its decimal display, about 701.955 cents (1.955 cents wider than the 700-cent equal-tempered fifth), is rounded only for reading. Values from floating-point numbers and from frequencies, such as the 7.85-cent rise from 440 to 442 Hz, are placed on a grid of one millionth of a cent. Values span ±9,000,000 cents (7,500 octaves each way) as a technical bound rather than a musical one.

## Small intervals, units, and text forms

Nineteen named small intervals cover commas, dieses, the schisma, and related microtonal differences; unqualified “the comma” is the syntonic comma 81:80 (about 21.506 cents), whereas Mercator's comma has two meanings in the literature and is formed by arithmetic instead. Eight named units range from the MIDI tuning unit to the semitone, with the savart defined as the decimal savart; any 1/*N*-octave unit, such as 1/19 octave (about 63.158 cents), comes from an equal division.

Values read and write as decimal cents (`701.955 cents`, `+30 ¢`), a fraction of cents (`1200/19 cents`), a ratio (`3:2`), or a product of prime powers (`5^(1/4)`), and in the pitch-value notation of Scala `.scl` files (`100.0`, `3/2`).

First released in v0.11.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/cents) · [Go examples](../../cents/example_test.go) · [Roadmap](../../ROADMAP.md)
