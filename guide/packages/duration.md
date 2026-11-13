# Durations and positions

Musical time in the library is measured in whole notes. A **duration** is how long something lasts — a quarter note lasts `1/4`, a dotted quarter `3/8`, an eighth of a triplet `1/12` — and a **position** is where something starts, counted from an origin: the third eighth of a voice starts at `1/4`. Both are exact fractions, so thirds, fifths, and sevenths of a beat add up without drift however long a piece is. The package serves the later time packages — meters, measures, tempos, tick conversion, events, voices, tracks, and scores — as well as score editors, notation format exchange, rhythm generators, and quantization.

The package knows lengths and offsets only. Notes, chords, rests as elements, measures, meters, tempos, seconds, and the units of file formats belong to the packages that build on it. Rests have the same durations as notes, so nothing here distinguishes them.

## Durations in whole notes

A duration is a fraction of a whole note greater than zero. It carries no notation: `3/8` is the same length whether it is written as a dotted quarter, as a quarter tied to an eighth, or as three triplet quarters of a larger group. Durations can be added, subtracted (the result must stay positive), multiplied and divided by whole numbers and by fractions, compared, and sorted. Two more operations answer common questions about lengths:

- the **ratio** of two durations is a plain number, not a duration: a dotted quarter holds three eighths, and a whole note holds sixty-four sixty-fourths;
- **whole fits with a remainder** tell how many times one duration fits in another entirely and what is left: a whole note holds two dotted quarters with a quarter left over.

Adding durations gives a duration, never a notation. Turning a length into note values is a separate, explicit step described below.

## The series of note values

A **note value** without dots is one of fourteen values, each twice as long as the next:

| Value | Length in whole notes | Text |
| --- | --- | --- |
| maxima | 8 | `maxima` |
| longa | 4 | `longa` |
| breve | 2 | `breve` |
| whole | 1 | `1` |
| half | 1/2 | `2` |
| quarter | 1/4 | `4` |
| eighth | 1/8 | `8` |
| sixteenth | 1/16 | `16` |
| thirty-second | 1/32 | `32` |
| sixty-fourth | 1/64 | `64` |
| 128th | 1/128 | `128` |
| 256th | 1/256 | `256` |
| 512th | 1/512 | `512` |
| 1024th | 1/1024 | `1024` |

From the breve to the 1024th the lengths follow modern notation. The longa and the maxima come from mensural notation, whose relations between values the library does not model; it gives them four and eight whole notes as a convention that continues the doubling of the series. Values shorter than the sixty-fourth are rare in practice but complete the series as far as LilyPond writes it.

Moving one step along the series doubles or halves a value: the next longer value after the whole is the breve, the next shorter one the half. Steps beyond the maxima or below the 1024th are reported as outside the series rather than wrapped around.

## Dots

Each dot adds half of what the previous dot or the value itself added, so a value *b* with *k* dots lasts *b*·(2 − 1/2^*k*):

- a dotted quarter lasts `3/8`;
- a quarter with two dots lasts `7/16`, with three dots `15/32`;
- a dotted 1024th lasts `3/2048`.

The number of dots is not limited by a rule. Four dots are already impractical on the page, but the only limit is the exact range of the fraction values; from 512 dots on, the length always leaves that range and is reported as an error rather than rounded.

## Tuplets in a written duration

A written duration can sit inside one or more [tuplets](tuplet.md), listed from the outermost group to the innermost. Each tuplet multiplies the length by its coefficient: an eighth in a triplet `3:2` lasts `1/12`, a quarter in a quintuplet holding a triplet lasts `1/4 · 4/5 · 2/3 = 2/15`. The chain is kept exactly as written — `6:4` is not reduced to `3:2`, and a quintuplet around a triplet is not reordered into a triplet around a quintuplet. A neutral group `1:1` is accepted: it changes nothing in the length but is still part of the notation.

A **written duration** therefore consists of a value, a number of dots, and a chain of tuplets. It can be read back part by part, asked whether it has dots, tuplets, or neither (a plain value), and changed into a new written duration: another value, a dot more or less, a tuplet added outside or inside, the outermost or innermost tuplet removed, all tuplets removed, or the value moved a number of steps along the series while the dots and tuplets stay. Each change gives a new written duration with a new length; nothing is changed in place.

## Two kinds of equality

A written duration is a notation, so two equalities are needed:

- **equality of notation**, the main one, compares the value, the dots, and the tuplets in order: a quarter in a triplet `3:2` and a quarter in a sextuplet `6:4` differ, and so do the chains `[3:2, 5:4]` and `[5:4, 3:2]`;
- **equality of length** compares only how long they last: all four of these are equal, as are a dotted half and a half tied to a quarter.

For sorting, written durations are ordered by length first; equal lengths are ordered by the number of tuplets (none first), then by value (longer first), then by the number of dots, then by the tuplets themselves. Durations and positions, which carry no notation, have only equality of value: `6/16` and `3/8` are the same duration value.

## Ties

A **tied duration** is two or more written durations under a tie, in order; its length is the sum of their lengths, so a half tied to a quarter lasts `3/4`. Each part keeps its own tuplets, so a tie can join a triplet note to a plain one. A single note is never a tie of one part. Parts can be appended at the end, and two ties can be joined into one.

## Writing a length as notes

The library writes a length as notes only when asked, and the caller always chooses the tuplets: a length of `1/5` could belong to a quintuplet of many shapes, and guessing would be wrong as often as right. A length is first divided by the coefficient of the chosen chain and then written in plain values; every resulting note carries the same chain.

**As one note.** A length can be written as a single value with dots exactly when it has the form (2^(*k*+1) − 1)·2^*j* and the value lies in the series; such a note is unique. `3/8` is a dotted quarter, `7/16` a quarter with two dots, `3` a dotted breve, and `15` a maxima with three dots, while `5/16`, `1/3`, `16`, and `1/2048` cannot be one note. Under a triplet, `1/3` is a half in `3:2`. A tie can be reduced the same way: a half tied to a quarter becomes a dotted half.

**As tied notes.** The rule takes the longest allowed note that does not exceed what is left, again and again, so the parts go from long to short. By default the allowed notes are plain values from the breve down: `5/16` becomes a quarter and a sixteenth, `11/16` a half, an eighth, and a sixteenth, and `5` two breves and a whole. The caller can choose a different longest value and allow dotted notes up to a number of dots: with one dot `11/16` becomes a half and a dotted eighth, and up to the maxima `12` becomes a maxima and a longa, or a single dotted maxima when one dot is allowed. If the length is one allowed note, the result is that one note.

The rule has two limits:

- **No notation by this rule.** A length the rule cannot finish — a third of a whole note without a tuplet, or anything finer than a 1024th — is refused. The refusal means that this rule gives no notation, not that none exists: `1/512` with values down to the 1024th and one dot allowed is refused, although two tied 1024ths would do. Other notations of the same length are not searched for.
- **At most 1024 parts.** A length of `2048` whole notes is 1024 breves; `2049` would need 1025 parts and is refused without building them. Long lengths are split across measures rather than tied into one endless note.

Splitting notes at bar lines and beats is the job of the measure packages, not of this rule.

## Normal subdivision

A plain value divides normally into two notes one step shorter, and a dotted value into three: a quarter into two eighths, a dotted quarter into three eighths, a dotted half into three quarters, and a quarter in a triplet into two eighths of the same triplet. A value with two or more dots has no normal subdivision, and the 1024th has none because its parts would be shorter than the series allows.

## Positions and their origin

A **position** is an offset of at least zero from an origin chosen by whoever uses it — the start of a voice, of a measure, or of a whole piece. Zero is the origin itself and is a valid position, unlike a zero length, which is not a duration at all. Starting from the origin and adding a quarter, an eighth, an eighth, and a half gives the positions `1/4`, `3/8`, `1/2`, and `1`.

A duration can be added to a position or subtracted from it, as long as the result does not fall before the origin. Between two positions there are two different answers:

- the **offset** from one to the other is itself a position and may be zero: from `1/2` to `1/2` it is the origin;
- the **duration** from one to the other must be positive: from `1/4` to `1/2` it is `1/4`, and between equal positions there is none.

## Quantization

Rounding happens only when explicitly requested; every other operation is exact. Quantization moves a duration or a position to a node of a uniform grid — the multiples of a positive step, such as every sixteenth — starting at zero. A grid starting elsewhere is obtained by subtracting its start first.

The rule for choosing between the two neighboring nodes is one of the nine [rounding rules](common/rounding.md) and must always be given; there is no default. A value already on a node stays where it is under every rule. Since durations and positions are never negative, four pairs of rules coincide here: toward zero and down, away from zero and up, half toward zero and half down, half away from zero and half up. For example, the position `5/16` on an eighth grid becomes `3/8` with half up and `1/4` with half down or half even, and the duration `1/3` on a sixteenth grid becomes `5/16` with half even.

A span — a start and a length — is quantized by rounding its start and its end on the same grid with the same rule, so the end does not drift by the rounding of the length: a span from `1/10` lasting `1/5` on an eighth grid becomes a span from `1/8` to `1/4`. A duration or span that rounds to zero is refused, because a duration must stay positive. Quantization returns durations and positions, never a notation.

## Names

Note values have English names in American and British usage and Russian names, but only where the library's references give them:

| Value | American | British | Russian |
| --- | --- | --- | --- |
| breve | double whole note | breve | бревис |
| whole | whole note | semibreve | целая |
| half | half note | minim | половинная |
| quarter | quarter note | crotchet | четвертная |
| eighth | eighth note | quaver | восьмая |
| sixteenth | sixteenth note | semiquaver | шестнадцатая |
| thirty-second | thirty-second note | — | тридцатьвторая |
| sixty-fourth | sixty-fourth note | — | шестьдесятчетвёртая |
| 128th | — | — | стодвадцатьвосьмая |

The maxima, the longa, and the values from the 256th down have no names here, and neither do the cells marked with a dash. A missing name is reported as "no name", not as an error, and no name is invented: the labels in this chapter's first column are descriptions, not names returned by the library. Names can be added later if the references gain them. Reading a value accepts its text, the LilyPond words `\maxima`, `\longa`, and `\breve`, and all names above, ignoring the case of letters.

## Text

Every value has a short canonical text modeled on LilyPond:

- a duration or a position is a reduced fraction such as `3/8`, `1/1`, or `0/1`;
- a written duration is its value, one period per dot, and its tuplets in brackets: `4.` is a dotted quarter, `8[3:2]` an eighth in a triplet, `4.[5:4,3:2]` a dotted quarter in a quintuplet holding a triplet, and `breve..` a breve with two dots;
- a tied duration joins its parts with a tilde: `2~4` is a half tied to a quarter, `4[3:2]~8[3:2]` a triplet quarter tied to a triplet eighth.

Reading accepts exactly this grammar: no spaces, signs, leading zeros, other case, or names. An unreduced fraction such as `6/16` is read as `3/8`. The digits of a value are read as a whole, so `16.` is a dotted sixteenth, and `3` is not a value. Reading the text of any value gives the same value back. The same texts are used inside structured documents such as JSON. MusicXML note types, MEI durations, and LilyPond markup for tuplets and ties are not read; exchange with those formats happens outside the library.

When a long text is rejected, the error quotes at most its first 64 bytes, without cutting a character in half, so error messages stay short however long the input is.

## The unset value

An empty duration, written duration, tie, or position means that nothing is set; it is neither zero nor the origin. It can be compared and sorted — it comes before every set value and equals only another unset value — and it prints as a recognizable placeholder that reading refuses, but every calculation with it reports an error. An optional field in a structured document is therefore left out rather than written as a placeholder.

Durations and positions are immutable values that are safe to copy, to use as keys, and to use from many goroutines at once.

First released in v0.20.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/duration) · [Go examples](../../duration/example_test.go) · [Roadmap](../../ROADMAP.md)
