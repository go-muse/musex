# Time signatures

A **time signature** says how one full measure is built: how long it is, what note value the numerator counts, and how those values gather into beats and groups. The library keeps a time signature both as it is written — `3/4`, `C`, `(3+2+3)/8` — and as the structure of a measure: its size in whole notes, its units, groups, and beats, its kind in two theory traditions, and its strong beats. All lengths and offsets are exact [durations and positions](duration.md), so a measure of `4/3` or `2/4+3/8` is described as precisely as `4/4`.

A time signature describes one measure only. The sequence of measures, pickups, bar numbers, the measure at which a signature takes effect, and changing signatures from measure to measure belong to the measure grid; the speed of the beat belongs to tempo. Notes, events, and the units of file formats are outside it as well. The name of the package means the time signature, not meter as a broad topic of music theory, and the measure grid is a separate package with its own name.

## Four forms of a signature

A signature is written in one of four forms:

| Form | Example | What it holds |
| --- | --- | --- |
| simple | `3/4`, `6/8` | a numerator over a denominator |
| symbol | `C`, `C\|` | common time with the structure of `4/4`, or cut time (¢, alla breve) with the structure of `2/2` |
| compound numerator | `(3+2+3)/8` | two or more addends over one denominator |
| sum of signatures | `2/4+3/8` | two or more parts, each a simple signature or a compound numerator |

The record is kept as written: `C` and `4/4` are different records of the same structure, and so are `(2+3)/8`, `5/8[2+3]`, and `2/8+3/8`. `4/2` has no symbol of its own and is written with digits.

Any whole denominator of at least one is accepted. Denominators are usually powers of two, but `3/1` and `1/1` occur, and an irrational signature such as `4/3` — four thirds of a whole note — is a full member with a size, units, groups, beats, a kind, and strong beats. Fractional numerators such as 2½/4 are not supported: the same measure is written `5/8` or `(2+2+1)/8`. A measure has at most 1024 units, the same limit as the number of tied parts of a duration; longer stretches without a meter are a matter for the measure grid.

## Units and beats

The word "beat" is used in two senses in theory books, and the library keeps both under different names:

- a **unit** is what the numerator counts — the note value one over the denominator. `6/8` has six units, each an eighth;
- a **beat** is the beat of the English-language tradition — the pulse a conductor or a metronome follows. `6/8` has two beats, each a dotted quarter.

For the plain signatures `2/4`, `3/4`, `4/4`, `2/2`, `3/8`, `C`, and cut time the two coincide: each unit is a beat. In every other case the beats are the groups described below: `5/8` grouped `2+3` has two beats, a quarter and a dotted quarter, and five units. Units always exist; beats need a grouping, so `5/8` written without one has five units and no beats. A unit, a beat, or a group can also be read as a written note value when its length is one note without a tuplet: the beats of `6/8` are dotted quarters, but the unit of `4/3`, a third of a whole note, has no such note.

## Grouping

A **group** is several consecutive units that form one cell of the measure, and the **grouping** is the sizes of the groups in units, such as `2+3`. Groups follow one another and cover every unit; the first group starts at the beginning of the measure.

**Standard grouping.** A simple signature or a symbol is grouped by its numerator unless told otherwise:

| Units | Standard grouping |
| --- | --- |
| 1, 2, 3 | one group |
| 4 | 2+2 |
| 6, 9, 12, 15, … (multiples of three) | groups of three |
| 5, 7, 8, 10, 11, 13, … | none |

So `4/4` and `C` are grouped `2+2`, `6/8` `3+3`, `12/8` `3+3+3+3`, and `5/8`, `7/8`, and `8/8` have no grouping of their own. Groups of three for multiples of three beyond twelve, such as `15/8`, are a rule of the library: the theory examples stop at twelve.

**Explicit grouping.** A simple signature or a symbol can be given any grouping whose groups add up to the numerator: `5/8[2+3]`, `7/8[3+2+2]`, a single group `5/8[5]`, or a nonstandard grouping of an ordinary signature such as `6/8[2+2+2]`. A grouping equal to the standard one changes nothing: `6/8[3+3]` is the same value as `6/8`. Removing an explicit grouping returns the plain signature. Groups smaller than a unit are not part of a grouping: a `4/4` bar felt as `3+3+2` eighths is written `(3+3+2)/8`.

**Grouping of the other forms.** The groups of a compound numerator are its addends: `(3+2+3)/8` is grouped `3+2+3`. A sum has the groups of its parts in order, each part grouped as a signature of its own: `2/4+3/8` is grouped `2+3`, and `4/4+3/8` `2+2+3`. If one part has no grouping, as in `5/8+2/8`, neither does the sum. The groups of these two forms are part of the record, so an explicit grouping is not given to them.

**Grouping variants.** For a given number of units the library lists every way to split them into groups of two and three, in order of the group sizes: five units give `2+3` and `3+2`, seven units `2+2+3`, `2+3+2`, and `3+2+2`. One unit has no variants. Twenty-eight units would already have more than 1024 variants, so the list is refused from there on without being built.

## Size and fitting note values

The **size** of a signature is the length of a full measure: the sum, over its parts, of numerator over denominator. `3/4` and `6/8` both last `3/4` of a whole note, `12/8` lasts `3/2`, `C` and `(3+2+3)/8` last a whole note, and `2/4+3/8` lasts `7/8`. The library can also tell how many times a given length fits wholly into a measure and what is left: a measure of `6/8` holds six eighths, or two dotted quarters, or one half with a quarter left over.

## Kinds in two traditions

The English-language and the Russian traditions classify signatures differently, and the same words mean different things in them: a "compound" signature in an English textbook is not the "сложный" signature of a Russian one. The library computes both kinds independently, each by its own rules, from the structure of the measure.

**English-language tradition.** A signature is simple when its beat divides into two, compound when it divides into three, and asymmetrical when its beats are unequal; the number of beats completes the name. Equal groups of two units count as simple beats: `6/8[2+2+2]` is simple triple. A signature without a grouping whose units are equal and whose number is odd, at least five, and not a multiple of three — `5/8`, `7/8` — is asymmetrical even though its beats are unknown.

**Russian tradition.** A signature of one group of two or three units is simple, several equal groups of two or three units make it compound uniform, and unequal groups make it compound mixed. `5/8` and `7/8` without a grouping are compound mixed.

Groups are equal only when both their number of units and their unit are equal: in `2/4+2/8` the two groups of two have different units, so the signature is asymmetrical and compound mixed.

| Signature | English-language kind | Beats | Russian kind |
| --- | --- | --- | --- |
| `2/4`, `2/2`, `C\|` | simple duple | 2 | simple |
| `3/4`, `3/8` | simple triple | 3 | simple |
| `4/4`, `C`, `4/3` | simple quadruple | 4 | compound uniform |
| `6/8`, `6/4` | compound duple | 2 | compound uniform |
| `9/8` | compound triple | 3 | compound uniform |
| `12/8` | compound quadruple | 4 | compound uniform |
| `15/8` | compound | 5 | compound uniform |
| `6/8[2+2+2]` | simple triple | 3 | compound uniform |
| `5/8[2+3]`, `2/4+3/8` | asymmetrical | 2 | compound mixed |
| `7/8[3+2+2]`, `(3+2+3)/8` | asymmetrical | 3 | compound mixed |
| `5/8`, `7/8` | asymmetrical | unknown | compound mixed |
| `1/4` | undetermined | 1 | undetermined |
| `(4+4)/8` | undetermined | 2 | undetermined |
| `4/4[3+1]` | asymmetrical | 2 | undetermined |

"Undetermined" is an answer, not an error: it marks a structure that the rules of the tradition do not cover, such as a single unit or groups of four. A signature such as `8/8` without a grouping has no kind in either tradition until it is grouped. Readings that depend on tempo or style — a fast waltz counted in one, `6/4` felt as three half notes — are not guessed; a different grouping is given explicitly.

**Beat division.** In a simple kind the beat divides into two and subdivides into four; in a compound kind it divides into three and subdivides into six. Asymmetrical and undetermined signatures have no beat division.

## Strong beats

The first unit of the first group is the **strong** beat, the first units of the other groups are **relatively strong**, and every other unit is **weak**. In `4/4` the strong beat is at the start and a relatively strong one at the half; in `6/8` at the start and at `3/8`; in `7/8[3+2+2]` at the start and at `3/8` and `5/8`; `3/4` has no relatively strong beat. Here, as in Russian theory, a "beat" is a unit. A point inside a unit is reported as off the units, and a signature without a grouping has no strong beats until it is grouped.

The theory sources describe strong beats for simple, compound, and mixed signatures but give no general rule for any grouping and no deeper hierarchy. Applying the same three levels to every grouping is a rule of the library; which of the relatively strong beats of `12/8` is heavier is not modeled.

## Places in a measure

Units, groups, and beats are numbered from one, as musicians count them; offsets are measured from the start of the measure, which is zero. For any offset inside the measure the library finds the unit, the group, and the beat that contain it and the offset inside each, and the weight at that point. In `6/8` the offset `1/2` is in unit five at its start, in beat two and group two an eighth after their start, and on a weak unit. Conversely, the start of any unit, group, or beat is found by its number: beat three of `4/4` starts at `1/2`. A point exactly at the start of an element belongs to it, and the end of the measure — an offset equal to its size — belongs to the next measure, not to this one.

## Three kinds of equality

Two signatures can be compared in three ways:

- **by record**, the main equality: the form, the symbol, the parts with their addends and denominators, and the explicit grouping all match. `C` and `4/4` differ, and so do `5/8` and `5/8[2+3]`;
- **by structure**: the same units in order and the same grouping. `C` and `4/4` are equal, and so are `(2+3)/8`, `5/8[2+3]`, and `2/8+3/8`;
- **by size**: `3/4` and `6/8` are equal although their structures differ.

Equal records have equal structures, and equal structures have equal sizes. For sorting, signatures are ordered by size and, at an equal size, by their text, so the order never treats two different records as equal: `2/4` comes before `3/4`, and `3/4` before `6/8`.

## Changing a signature

**Changing the record** gives the same structure in another form:

- with digits: `C` becomes `4/4`, cut time `2/2`, and `C[3+1]` becomes `4/4[3+1]`;
- with a symbol: `4/4` becomes `C` and `2/2` cut time; other signatures have no symbol;
- as a simple signature: `(2+3)/8` becomes `5/8[2+3]` and `(3+3)/8` becomes `6/8`; `5/8+2/8` becomes `7/8` without a grouping, while `5/8+1/8` cannot be written so, because `6/8` always has the grouping `3+3`;
- as a compound numerator: the groups become the addends, so `6/8` becomes `(3+3)/8` and `C` `(2+2)/4`;
- as a sum: each group becomes a part, so `6/8` becomes `3/8+3/8`, `C` `2/4+2/4`, and `C[3+1]` `3/4+1/4`. A group of more than three units cannot become a part, because the part would be regrouped.

A compound numerator and a sum need at least two groups, so `3/4` and cut time have neither; a sum with different denominators has no simple form and no compound numerator. A signature already in the requested form is returned as it is.

**Changing the unit** halves or doubles every denominator, keeping the addends and the grouping: `3/4` becomes `3/8` or `3/2`, `5/8[2+3]` becomes `5/4[2+3]`, and `2/4+3/8` becomes `2/8+3/16`. Units cannot be doubled when a denominator is odd, as in `3/1` or `4/3`. A signature with one denominator can also be given another denominator directly: `6/8` with the denominator 4 is `6/4`. A symbol is written with digits after any such change.

**From a size.** A size and a denominator give the simple signature of that size: `3/4` of a whole note in eighths is `6/8`, and a whole note in thirds is `3/3`. A size that is not a whole number of units is refused.

## Names

The library returns names only from its theory references:

| Value | English | Russian |
| --- | --- | --- |
| simple kind | simple | простой |
| compound kind | compound | сложный |
| asymmetrical kind | asymmetrical | асимметричный |
| two, three, four beats of a simple or compound kind | duple, triple, quadruple | двухдольный, трёхдольный, четырёхдольный |
| simple kind of the Russian tradition | — | простой |
| compound uniform kind | — | сложный однородный |
| compound mixed kind | — | сложный смешанный |
| strong, relatively strong, weak beat | — | сильная, относительно сильная, слабая |
| common time | common time | — |
| cut time | cut time | alla breve |

The full name of an English-language kind adds the number of beats: `3/4` is simple triple (простой трёхдольный), and `6/8` compound duple (сложный двухдольный). Only simple and compound kinds with two, three, or four beats name their number of beats; `15/8` is simply compound, with five beats as a number, and an asymmetrical signature is asymmetrical with its beats as a number. Undetermined kinds and points off the units have no name. A missing name is reported as "no name", not as an error, and no name is invented: the labels in the first column of the table are descriptions, not returned names. Signatures are not named in words such as "three-four".

## Text

Every signature has a short canonical text without spaces: `3/4`, `5/8[2+3]`, `C`, `C|`, `C[3+1]`, `(3+2+3)/8`, `2/4+3/8`, `(2+3)/8+2/4`. Cut time is written `C|`, as in ABC notation, so the text stays plain ASCII and easy to type; reading also accepts the ¢ sign. Reading accepts exactly this grammar — no spaces, signs, leading zeros, lowercase `c`, or names — and also an explicit grouping equal to the standard one, such as `6/8[3+3]`. A compound numerator or a sum with an explicit grouping, a part of a sum written as a symbol, and a compound numerator of one addend are refused. Texts of different records differ, and reading the text of any signature gives the same signature back. MusicXML time elements, MIDI time-signature events, and LilyPond time commands are not read; exchange with those formats happens outside the library.

When a long text is rejected, the error quotes at most its first 64 bytes, without cutting a character in half, so error messages stay short however long the input is.

## The unset value

An empty signature means that nothing is set; it is not a measure. It can be compared and sorted — it comes before every set signature and equals only another unset value — and it prints as a recognizable placeholder that reading refuses, but every calculation with it reports an error. An optional signature in a structured document is therefore left out rather than written as a placeholder.

Signatures are immutable values that are safe to copy, to use as keys, and to use from many goroutines at once.

First released in v0.24.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/meter) · [Go examples](../../meter/example_test.go) · [Roadmap](../../ROADMAP.md)
