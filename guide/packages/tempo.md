# Tempo

A **tempo marking** says how fast the music goes: a note value, the **unit**, and the number of units per minute, the **bpm**. ♩ = 60 is sixty quarters a minute, and ♩. = 60 is sixty dotted quarters. The library keeps a marking as it is written and turns the exact [durations](duration.md) of a score into exact seconds while the tempo stays the same. It also offers the formula that connects bars, units, bpm, and time, approximate markings and tempo ranges, the scale of a metronome, and tempo indications that join a marking with a word such as Allegro from the [tempo terms](tempo/term.md).

The package knows one constant tempo at a time. Tempo changes, gradual accelerations and slowdowns, a tempo, Tempo I, and fermata holds belong to tempo maps; MIDI ticks and MusicXML divisions belong to tick conversion; time signatures and measure grids belong to their own packages. The length of a bar comes into the formula as a plain duration, so nothing here needs to know a time signature.

## Unit and bpm

The unit is any written note value of one sign: a quarter, a dotted quarter, a half, an eighth in a triplet. A unit tied from two notes is not a unit, because a marking names one sign. The bpm is any exact positive number: whole, such as `60`, or fractional, such as `185/2`, which a format with decimals writes as 92.5. The library does not limit bpm to the 40 to 208 of a metronome: sources name slower and faster tempos, and conversions give fractions. A bpm of zero or below is not a tempo and is refused.

A number of bpm alone, with no unit, means a quarter unit: "bpm 60" and ♩ = 60 are the same marking.

## Speed and the two equalities

The **speed** of a marking is the unit times the bpm, in whole notes per minute. ♩. = 60, ♪ = 180, and ♩ = 90 all move at 22½ whole notes a minute, and ♩. = 60 is ninety quarters a minute. So two equalities are needed:

- **equality of record**, the main one, compares the unit as written and the bpm by value: ♩ = 60 and ♩ = 120/2 are the same marking, while ♩. = 60 and ♪ = 180 are not;
- **equality of speed** compares only how fast the music goes: ♩. = 60, ♪ = 180, and ♩ = 90 are equal.

Markings can also be compared by speed — faster, slower, or equal — and divided to see how many times one is faster: ♩ = 120 is 4/3 times as fast as ♩. = 60. For sorting, markings are ordered by speed, and markings of equal speed by their unit, shorter units first: ♪ = 180, then ♩ = 90, then ♩. = 60.

## Changing a marking

- **Another unit at the same speed.** ♩. = 60 becomes ♪ = 180, ♩ = 90 becomes 𝅗𝅥 = 45, and ♩ = 100 becomes ♩. = 200/3. A bpm may become fractional; it is not rounded.
- **Faster or slower.** Multiplying the bpm by a positive number keeps the unit: ♩ = 60 twice as fast — the doppio movimento — is ♩ = 120, half as fast ♩ = 30, and one and a half times as fast ♩ = 90. A change by a percentage is a multiplication too.
- **Metric modulation.** "New dotted quarter equals old quarter" at ♩ = 120 gives ♩. = 120: in the new tempo the new unit lasts as long as the old note did. The old note can be any length, including a tied one or a triplet value. Going the other way, from two markings the library finds which old length equals the new unit, so the change can be written as an equation of two notes.

## Seconds

A length *d* in whole notes lasts 60·*d* divided by the speed, in seconds. At ♩ = 60 a half lasts 2 seconds; at ♩. = 60 an eighth lasts 1/3 of a second, and a bar of 6/8 lasts 2 seconds; at ♩ = 90 a triplet eighth lasts 2/9 of a second. Times stay exact, so thirds and other fractions of a second never drift, however long a piece is.

The same relation answers the reverse questions:

- how long one unit lasts and how many units sound in a second: at ♩ = 120, half a second and two units;
- which length sounds for a given time: at ♩ = 60, two seconds is a half;
- which bpm makes a given length last a given time: a quarter lasting half a second means ♩ = 120, which is the MIDI tempo of 500,000 microseconds a quarter;
- how many units fit in a time, and how long a number of units lasts: at ♩ = 90, forty seconds hold sixty quarters. Zero time holds zero units, and zero units take zero time.

A time is a value of its own, at least zero: it can be created from seconds, from minutes, or from a system time, read in seconds or minutes, compared, added, and subtracted as long as the result does not become negative. Zero is a valid time, the start of a count.

**System time.** Turning an exact time into the system time of the computer rounds it to whole nanoseconds, and only then. The rounding rule is one of the nine [rounding rules](common/rounding.md) and must always be given; there is no default. One third of a second is 333,333,333 nanoseconds rounded down or half even and 333,333,334 rounded up, and a whole number of nanoseconds stays as it is under every rule. Times are never negative, so four pairs of rules coincide here: toward zero and down, away from zero and up, half toward zero and half down, half away from zero and half up. The longest system time is about 292 years; a longer time is refused rather than cut off, and stays exact in seconds. A system time turns back into an exact time without rounding.

## The formula of bars and time

*n* bars, each holding *r* units, last *t* minutes at *N* bpm when *n*·*r* = *N*·*t*. The number *r* is the length of the bar divided by the length of the unit: four for a 4/4 bar counted in quarters, two for 6/8 counted in dotted quarters. Given any three of the four numbers, the library finds the fourth:

- 4/4 at ♩ = 120 for two minutes is 60 bars, and 60 bars take two minutes;
- 3/4 at ♩ = 100 for one minute is 100/3 bars: the time ends inside a bar, and the number of bars stays fractional;
- 30 bars of two units in one minute need 60 bpm, and 30 bars in one minute at ♩ = 60 mean two quarters, a bar of 1/2.

Bars per minute are the bpm divided by *r*: 3/4 at ♩ = 180 is 60 bars a minute, and 60 bars of 3/4 a minute is ♩ = 180 or, counted in dotted halves, 𝅗𝅥. = 60.

The formula holds only for a **constant tempo and equal full bars**. The library does not check this: it knows no bars, only the length it is given. Numbers of zero or below, and a time of zero, have no solution and are refused.

## Rounding the bpm

A bpm can be moved to a grid — whole numbers, halves, or any positive step — with an explicitly chosen rounding rule, for formats that store whole or decimal bpm: ♩. = 200/3 on a whole-number grid becomes ♩. = 67 rounded half up and ♩. = 66 rounded down. A bpm that would round to zero is refused.

## Approximate markings and ranges

An **approximate marking** such as ♩ ≈ 108 names a number without promising it. Scores write it as c. 108, ca. 108, or ≈ 108; the sources do not distinguish these signs, so the library keeps only the unit and the number. An approximate marking is never equal to the exact ♩ = 108, has no tolerance, and does not answer whether a number lies "near" it; it can change its unit at the same speed (♩. ≈ 60 becomes ♪ ≈ 180) and becomes exact only when the caller explicitly takes its named number.

A **tempo range** such as ♪ = 132–144 has both a lower and an upper number, the lower first and smaller. Both numbers belong to the range — a convention of the library, since the sources do not say — so 132, 138, and 144 lie in it and 131 does not. A marking lies in a range when its speed does: ♩ = 69 lies in ♪ = 132–144, ♩ = 65 does not. A range becomes an exact marking only by an explicit choice: the lower bound, the upper bound, or the exact middle, ♪ = 138 here and ♩ = 251/2 for ♩ = 120–131. A rounded middle, as some notation programs print, comes from rounding the bpm afterwards: ♩ = 126 rounded half even. Open ranges, such as "200 and over", exist only as guides of tempo terms.

Seconds are computed only from an exact marking; an approximate marking or a range first has to be turned into one by the caller.

## The metronome scale

The scale of Maelzel's metronome has 39 divisions from 40 to 208: by twos up to 60, by threes up to 72, by fours up to 120, by sixes up to 144, and by eights up to 208. A division is a number of beats per minute without a note value; a marking placed on the scale keeps its unit.

The library tells whether a number is a division and which one, gives the next division up and down — 63 after 60, 40 after 30, 208 before 250 — and finds the two **neighbors** of a number between divisions together with which one is nearer: 62 lies between 60 and 63, nearer to 63; 110 lies exactly between 108 and 112. Choosing one division, or moving a marking onto the scale, takes a rounding rule chosen by the caller: 110 becomes 112 rounded half up and 108 rounded half down. A number on a division stays on it under every rule.

Three rules of the scale are conventions of the library, not facts of the sources:

- **nearness is measured by difference**, so 110 is exactly between 108 and 112;
- **half even** breaks a tie by the position of the lower neighbor, counting 40 as position zero: 108 is at an odd position, so 110 goes up to 112, while 72 is at an even position, so 74 goes down to 72;
- **a number outside 40 to 208 is refused** when neighbors or the nearest division are asked for. Sources suggest moving it to the nearest end of the scale; the library does not do that silently, because a metronome cannot play such a tempo.

## Words and indications

Tempo words such as Largo or Allegro, their groups, and the numbers that lists give for them are described in the chapter on [tempo terms](tempo/term.md). Here they meet markings in two ways.

**Words for a marking.** The library finds every word of a chosen list whose numbers contain the bpm of a marking. The unit is not taken into account: the numbers of the lists name beats without a note value, so ♩. = 60 finds the words around 60 just as ♩ = 60 does. Where quarters are meant, the marking is first converted to a quarter unit. The numbers of a word can also be turned into a tempo range with a unit chosen by the caller: Allegro with a quarter is ♩ = 120–156.

**Tempo indications.** Above a score a word and a marking often stand together: "Allegro, ♩ = 120". The marking sets the speed and the word the character, so the library does not check that they agree — "Allegro, ♩ = 60" is a valid indication. An indication holds a word, a marking of any of the three kinds, or both; one of them must be present. Its exact marking, if it has one, gives the seconds; "Moderato, ♩ ≈ 108" has none.

## Text

Every marking has a short canonical text in which the unit is written as in the chapter on [durations](duration.md):

- an exact marking is the unit, an equals sign, and the bpm: `4=60` for ♩ = 60, `4.=60` for ♩. = 60, `4=185/2`, `8[3:2]=360`;
- an approximate marking uses the sign ≈: `4≈108`;
- a range joins its numbers with a hyphen: `8=132-144`;
- a time is a reduced fraction of seconds: `0/1`, `1/3`, `2/1`;
- an indication is a word, a marking, or both separated by a comma and a space: `Allegro`, `4=60`, `Allegro, 4=120`, `Moderato, 4≈108`.

Numbers print as whole numbers or reduced fractions; reading also accepts a finite decimal such as `92.5`. The approximate sign is ≈ only: `4=~108` and `4~108` are not read. Note symbols such as ♩ and ♪ are neither written nor read, and neither are MusicXML, MEI, or LilyPond markup. Reading the text of any value gives the same value back.

A word of a user list may contain any characters. When its name contains `=` or `≈` or starts with a quotation mark, an indication writes it in Go quotation marks, so the word named `4=60` prints as `"4=60"` and is never confused with the marking `4=60`; other names, such as `Скоро, но не слишком`, stay as they are. Quotation marks are accepted around any name when reading. An indication is read only together with the list in which its word is looked up, ignoring case.

When a long text is rejected, the error quotes at most its first 64 bytes, without cutting a character in half.

## The unset value

An empty marking, time, range, or indication means that nothing is set; an empty time is not zero. It can be compared and sorted — it comes before every set value and equals only another unset value — and it prints as a recognizable placeholder that reading refuses, but every calculation with it reports an error.

Markings, times, ranges, and indications are immutable values that are safe to copy, to use as keys, and to use from many goroutines at once.

First released in v0.26.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/tempo) · [Go examples](../../tempo/example_test.go) · [Roadmap](../../ROADMAP.md)
