# Tuplets

A tuplet plays a number of equal notes in the time normally taken by a different number of notes of the same written value: a triplet plays three eighths in the time of two, a duplet plays two notes in the time of three. The same concept is called *дуоль*, *триоль*, *квинтоль* in Russian usage and duplet, triplet, quintuplet in English usage. The package describes tuplets exactly and computes their action on durations without rounding, so that later durations, chords, and tracks can accumulate positions such as thirds and fifths of a beat without drift. It also serves score editors, notation format exchange, rhythm generators, and quantization.

## Actual and normal counts

A tuplet is a pair *n*:*m* of two whole numbers, each at least one:

- the **actual count** *n* says how many notes sound in the group;
- the **normal count** *m* says how many notes of the same written value they replace.

A triplet is `3:2`, a quintuplet in the time of four is `5:4`, and a duplet in the time of three is `2:3`. The pair corresponds directly to the actual and normal note counts of MusicXML and to the number and base number of MEI, so a tuplet read from a score keeps both numbers as written.

The pair alone defines the tuplet; the written note value of its parts (eighths or quarters), the meter, the inner grouping of the notes (two plus two plus two or three plus three in a sextuplet), brackets, and numbers on the page are not part of it. A quarter note under a triplet of eighths belongs to the same `3:2` group and is changed by the same factor as the eighths. Note values, dots, and measures belong to the packages that build on tuplets.

A group whose notes are counted in different units — three eighths in the time of one quarter — is converted into a pair counted in the unit of the sounding notes: three eighths in the time of one quarter become `3:2`. The conversion succeeds only when the normal part becomes a whole number of those units; the original units are not kept.

## The direction of the coefficient

The **duration coefficient** of *n*:*m* is *m*/*n*, normal over actual. Multiplying a written duration by it gives the sounding duration of each note of the group:

- in a triplet `3:2` the coefficient is `2/3`, so an eighth note (`1/8` of a whole note) sounds `1/12`;
- in a duplet `2:3` the coefficient is `3/2`, so an eighth note sounds `3/16`, the length of a dotted eighth.

The **reciprocal coefficient** *n*/*m* is the "speed" factor that LilyPond writes as `3/2` for a triplet. Both directions are available under separate names, and applying the reciprocal undoes an applied tuplet, so the two directions cannot be confused silently. A tuplet can also be recovered from either coefficient.

The **total duration** of a group written with parts of a given value is *m* times that value: a quintuplet `5:4` of eighths spans `1/2`, a septuplet `7:4` of quarters spans a whole note, and fourteen sixty-fourths in the time of eight span `1/8`.

All coefficients and results are exact fractions. Durations may be zero or negative, so offsets can be shifted as well. Applied durations are returned in fully reduced form; a result can then be expressed in any chosen subdivision grid with the fraction tools of the library.

## Notation and value: `3:2` versus `6:4`

The pair is kept exactly as written. A triplet `3:2` and a sextuplet `6:4` have the same coefficient `2/3`, yet they are different notations, and a sextuplet is not the same as two triplets in a row. Two comparisons therefore exist:

- **equality of notation**, under which `3:2` and `6:4` differ — this is the main equality, and a tuplet can serve as a key that keeps both notations apart;
- **equality of coefficient**, under which `3:2` and `6:4` are equal.

This is the reverse of the fraction values of the library, whose main equality compares numeric values and whose second equality compares the written pair. A third comparison orders tuplets for sorting: by coefficient, then by actual count, then by normal count, so `3:2` comes before `6:4`.

The coefficient of the notation keeps its pair: the coefficient of `6:4` is `4/6`, equal to `2/3` by value; the reduced coefficient `2/3` is available as well. A notation can also be reduced (`6:4` to `3:2`) or scaled by a whole factor (`3:2` by two to `6:4`) without changing the coefficient.

## Compressing, expanding, and neutral tuplets

A tuplet with more actual than normal notes is **compressing**: its notes are shorter than written, as in triplets and quintuplets. A tuplet with fewer actual notes is **expanding**: its notes are longer, as in a duplet `2:3` in a dotted beat.

Pairs with equal counts, including `1:1`, are accepted and called **neutral**: their coefficient is one. This is a product decision rather than a fact of notation practice. Musical sources describe only tuplets with different counts, but nesting needs neutral pairs — a duplet `2:3` inside a triplet `3:2` flattens to `6:6` — and an empty nesting chain amounts to `1:1`.

## Default normal counts

Musicians often write only the actual count and leave the normal count implied by the division of the beat. Two kinds of normal division are distinguished:

- **binary** division splits a plain value into 2, 4, 8, … equal parts;
- **ternary** division splits a dotted value into 3, 6, 12, … equal parts.

These correspond to the division of the beat in simple and compound meters in the English classification; the package does not model meters, so the kind of division is given explicitly.

The default normal count is the nearest smaller member of the division's series:

| Actual count | Binary division | Ternary division |
| --- | --- | --- |
| 2 | none | `2:3` |
| 3 | `3:2` | none |
| 4 | none | `4:3` |
| 5 | `5:4` | `5:3` |
| 6 | `6:4` | none |
| 7 | `7:4` | `7:6` |
| 9 | `9:8` | `9:6` |
| 14 | `14:8` | `14:12` |

The duplet in ternary division is the one exception to the rule: its default is `2:3`, as in all sources. There is no default for one note, nor for a count that itself belongs to the series — four notes in binary division or six in ternary division are a normal division rather than a tuplet group.

This rule is a choice of the library. It follows the "nearest smaller normal value" rule of English-language references for binary division and extends it to ternary division, where it agrees with those references for four, five, and seven notes. It differs from other sources in several places:

- some Russian sources write a septuplet as seven in the time of eight, not four;
- abc notation uses different implied pairs for five through nine notes, such as five or six in the time of two;
- no source confirms the octuplet `8:6`, the nonuplet `9:8`, or the decuplet `10:8`, and larger groups have no established norm.

Any other pair is created explicitly.

## Named tuplets

Groups of two through ten notes have names. Each named tuplet uses the binary default, or the ternary default where binary division has none:

| Count | English | Russian | Pair |
| --- | --- | --- | --- |
| 2 | duplet | дуоль | `2:3` |
| 3 | triplet | триоль | `3:2` |
| 4 | quadruplet | квартоль | `4:3` |
| 5 | quintuplet | квинтоль | `5:4` |
| 6 | sextuplet | секстоль | `6:4` |
| 7 | septuplet | септоль | `7:4` |
| 8 | octuplet | октоль | `8:6` |
| 9 | nonuplet | нонемоль | `9:8` |
| 10 | decuplet | децимоль | `10:8` |

The name depends on the actual count alone, so `5:3` is also a quintuplet and `7:11` a septuplet. Groups of one note and of eleven or more have no name; sources describe large groups by their numbers. The alternative English forms quartole, quintolet, pentuplet, sextolet, sestole, septolet, and septimole are recognized when reading names. The Russian words *дуплет* and *триплет* are not used.

## Nested tuplets

A tuplet can sit inside one part of another — a triplet inside a quintuplet, or a triplet inside a triplet. A chain of nested tuplets is listed from the outermost to the innermost, and their coefficients multiply:

- a triplet inside a quintuplet gives `8/15`;
- a triplet inside a triplet gives `4/9`, so an eighth note sounds `1/18` of a whole note, four ninths of an eighth;
- a duplet inside a triplet gives a neutral result with coefficient one.

The chain can also be flattened into a single notation by multiplying the counts: two nested triplets flatten to `9:4`, which describes the combined action as one pair when all tuplets of the chain count parts of the same note value. The total coefficient keeps the unreduced pair that matches this flattened notation. Neither result depends on the order of the chain, and an empty chain leaves durations unchanged.

Flattened counts are limited to the machine integer range, while the chain coefficient uses the much larger exact range of the fraction values; any chain of four tuplets has a coefficient. Exceeding a limit is reported as an error rather than wrapped around.

## Text

A tuplet is written as its two counts separated by a colon, such as `3:2` or `6:4`, in plain digits without signs, spaces, leading zeros, or zero counts. The same text is used when a tuplet is stored as part of a larger structured document. The short form writes only the actual count, such as `3`, and is read back with a kind of division, which gives the default tuplet for that count — not necessarily the original pair. Names are read in lower case in English, in the alternative English forms, and in Russian.

The French notation "3 pour 2", the LilyPond fraction, the abc tuplet syntax, MusicXML, and MEI are not read: exchange with these formats happens outside the library, and their numbers map directly to the pair.

When a long text is rejected, the error quotes at most its first 64 bytes, without cutting a character in half, so error messages stay short however long the input is.

## The unset tuplet

An empty tuplet value means that no tuplet is set; it is not `1:1`. It can be compared and sorted — it precedes every set tuplet and equals only another unset value — and it prints as a recognizable placeholder, but every calculation with it reports an error. Whether a duration carries a tuplet at all is decided by the packages that use tuplets.

Tuplets are immutable values that are safe to copy and to use from many goroutines at once.

First released in v0.17.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/tuplet) · [Go examples](../../tuplet/example_test.go) · [Roadmap](../../ROADMAP.md)
