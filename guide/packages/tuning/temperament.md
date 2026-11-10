# Temperaments

A temperament divides the octave into *N* positions and fixes the pitch of each position above a reference position 0 as an offset in [cents](../cents.md): 0 = *T*(0) < *T*(1) < … < *T*(*N* − 1) < 1200 cents. The table is periodic in the octave, so step *k* = *N*·*q* + *m* (0 ≤ *m* < *N*) lies 1200·*q* + *T*(*m*) cents above position 0 for every integer *k*, including negative steps. A temperament is independent of note names, of the note chosen as position 0, and of a reference frequency in Hz; those belong to a tuning.

## Families and built-in tables

All offsets are exact cent values, so pure ratios, fractions of commas, and equal steps keep their exact identities: the major third C–E of quarter-comma meantone is exactly 5:4, and the step sizes of any table add up to exactly one octave. Thirty-one built-in tables cover five families, and each family can also be built from parameters:

- **Equal divisions of the octave** (*N*-EDO) for any *N* from 1 to 65,536, with built-in 12-, 19-, 22-, 24-, 31-, and 53-EDO; the step of 53-EDO is the Holdrian comma.
- **Pythagorean tuning**: eleven pure 3:2 fifths with the remaining wolf fifth, about 678.495 cents, placed at a chosen point of the chain; built-in chains D♭…F♯ and E♭…G♯.
- **Just intonation**: twelve pure ratios, with three 5-limit tables and one 7-limit table.
- **Regular meantone**: every fifth narrowed by the same fraction of the syntonic comma (built-in 1/4, 1/3, 1/5, 1/6, and 2/7) or given directly as a size between 4800/7 and 720 cents, the range in which the twelve positions stay in strictly increasing order.
- **Well temperaments**: the Pythagorean comma distributed over some fifths of the circle, as in Werckmeister III, IV, and V (1691), Kirnberger I, II, and III, the temperament now known as Vallotti's, Young's first and second, Kellner's (1975), and four schemes by Neidhardt. Any circle of fifths, given as fifth sizes or as deviations from 3:2, and any regular generator of an *N*-step system build further tables.

Tables can also come directly from cent values, decimal cents, or frequency ratios with prime factors up to 31; an entry that breaks the order or leaves the octave is rejected rather than sorted or wrapped.

## Intervals, rotation, and analysis

Intervals are measured between any two steps across octaves, or upward within the octave between two positions. Rotation re-reads a table from another position, as when an unequal temperament is examined in another key; rotating an equal division gives the same table. The nearest position or step to any cent value is found together with its signed deviation.

A table can be compared exactly or within a tolerance with another, and checked for being an equal division or for consisting only of exact ratios. Its deviations from the equal division with the same *N* are summarized by their maximum, mean, and root mean square. For any interval class, such as the twelve fifths, major thirds, or minor thirds of a twelve-position table, the analysis lists every size and its deviation from a pure reference, counts the pure intervals, and marks the wolves: the fifths more than 15 cents and the thirds more than 25 cents from pure by default, such as the G♯–E♭ fifth of quarter-comma meantone.

First released in v0.15.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/tuning/temperament) · [Go examples](../../../tuning/temperament/example_test.go) · [Roadmap](../../../ROADMAP.md)
