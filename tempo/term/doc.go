// Package term provides tempo terms, the words that name a tempo in a score,
// such as Largo, Andante, or Allegro, and the lists that give them an order,
// a group, a reference range of beats per minute, and meanings. Package tempo
// uses it for tempo indications and for finding terms by a number of beats.
//
// # Lists and terms
//
// A [List] is an ordered list of terms. Each row of a list is a place
// numbered from 1, and a place holds one or more terms that share a group and
// a guide: Adagissimo and Grave share place 2 of the list of the English
// Wikipedia. A [Term] belongs to its list: it knows its name, place, group,
// guide, and meanings, and terms with the same name in different lists are
// different terms with their own places and guides. List.Place, List.Terms,
// and List.OfGroup read a list, List.Parse finds a term by name, and
// Term.ComparePlace compares the places of two terms of one list. Places are
// rows of the source, not tempos: the order of a list need not follow its
// numbers.
//
// # The eight built-in lists
//
//   - [EnglishWikipedia]: the article "Tempo" of the English Wikipedia,
//     section "Approximately from the slowest to the fastest", 20 places and
//     23 terms; English meanings from the article, Russian meanings from the
//     textbook "Элементарная теория музыки" by B. Alekseev and A. Myasoedov
//     where it gives them. It is the default list, returned also by
//     [Default].
//   - [RussianWikipedia]: the article "Темп (музыка)" of the Russian
//     Wikipedia, table "Темповые обозначения", 37 places and 39 terms in lower
//     case; Russian meanings from the article, no English meanings.
//   - [JapaneseMetronome], [GermanMetronome], and [FrenchMetronome]: the
//     scales of a Japanese metronome of the 21st century, a German one of the
//     20th, and a French one of the 19th, from the article "Tempo (Musik)"
//     of the German Wikipedia; 17, 8, and 6 places.
//   - [MaelzelMetronome], [Metronome1950], and [ElectronicMetronome]: a
//     19th-century Mälzel metronome, a metronome of 1950, and a modern
//     electronic metronome, from the table of tempo markings by Dolmetsch; 10,
//     12, and 8 places in lower case. A single number of the source is a guide
//     with equal bounds.
//
// The scales of metronomes have one term per place and no meanings. [BuiltIn]
// returns all eight lists in this order. The texts of names and meanings are
// those of the sources, verbatim, except for the conventions listed below.
//
// # Guides are references, not tempos
//
// A [Guide] is the range of beats per minute that a list gives for a term, with
// both bounds included; one bound may be open, as in "up to 24" or "from 200".
// The numbers come from weak sources: they are rough approximations, and other
// sources give other numbers for the same word — Presto is 168–200 in one
// list, 184–200 in another, and 100–152 elsewhere. A term alone does not
// determine a tempo, and the package never derives a tempo from a term. The
// numbers carry no duration: List.At compares a number of beats with them as
// it is, for every list. Where a user needs quarters, the tempo is converted
// to a quarter unit first. The sources of the two Wikipedia lists give their
// numbers for a quarter beat, and the scales of metronomes give none.
//
// # Conventions of the product
//
// The following rules are conventions of the library, not facts of the
// sources:
//
//   - Groups. A term has the group of its main word: Adagio, Largo, Lento, and
//     Grave are slow; Andante, Moderato, and Sostenuto moderate; Allegro, Vivo
//     (vivace), and Presto fast, as the textbook of Alekseev and Myasoedov
//     groups them. A derived word, such as Larghetto, Andantino, Allegretto,
//     or Prestissimo, takes the group of the word it derives from, and a
//     compound one, such as Marcia moderato or Molto Allegro, the group of its
//     first main word, so Allegro moderato is fast. Terms without a main word,
//     such as comodo, con moto, animato, tempo di marcia, Maestoso, and Assai,
//     have no group ([NoGroup]), which is a result, not an error. The sources
//     disagree on groups; these are the groups of the built-in lists.
//   - Bounds. Both bounds of a guide belong to it, and an open bound includes
//     the named number, following "24 bpm and under" and "200 bpm and over".
//     Guides overlap, so List.At returns every matching term.
//   - Vivo. In the list of the English Wikipedia, List.Parse reads "Vivo" as
//     Vivace, as several sources identify the two words; in the list of the
//     Russian Wikipedia vivo and vivace are different terms, as in its source.
//   - Allegretto. Its English meaning in the list of the English Wikipedia is
//     "by the mid-19th century, moderately fast", without the remainder of
//     the source text, which refers to another paragraph of the article.
//   - prestissimo. In the list of the Russian Wikipedia its name omits the
//     mark "(possible)" of the source, and its guide "192–208 and above" is
//     stored as "from 192", with an open upper bound.
//
// # The default list and user lists
//
// [Default] returns the list of the English Wikipedia, the only source that
// gives every term numbers and states an order from slow to fast. No operation
// uses a list implicitly: a list is always passed explicitly, and Default
// exists for that.
//
// [NewList] makes a user list from rows of [Entry] values, each with an
// optional group, guide, and meanings. A name is any non-empty valid UTF-8
// text, including "=", "≈", commas, and quotation marks; names of a list must
// differ up to case, as strings.EqualFold compares them. A guide is optional,
// but a given one must have at least one bound; [Between], [AtLeast], and
// [AtMost] create guides with positive bounds, the lower not greater than the
// upper, and equal bounds name a single number. Rows need not be ordered by
// their guides. A user list has no synonyms of parsing, so it is never equal to
// the list of the English Wikipedia, which parses "Vivo".
//
// # Search and parsing
//
// List.At returns all terms whose guide contains a number of beats, in the
// order of the list; terms without a guide are skipped, and an empty result is
// nil, not an error. List.Parse finds a term by name ignoring case, without
// trimming spaces; a name that is not in the list fails. Term.String,
// MarshalText, and AppendText write the name as stored, and List.Parse of that
// text returns the term. Term and List do not implement
// encoding.TextUnmarshaler, because a name identifies a term only within a
// named list.
//
// Every error of List.Parse and of its Must variant quotes the rejected name,
// except the error for an unset list, which wraps [ErrUnset] without a
// quotation. The quotation gives the name in full if it is at most 64 bytes
// long, otherwise its longest prefix of whole UTF-8 characters of at most 64
// bytes followed by "...". The message consists only of the text of the
// sentinel and this excerpt, is at most 512 bytes long, and depends on the
// input only through its first 67 bytes.
//
// # Unset values
//
// The zero values List{} and Term{} are unset. Only IsSet, Equal, String, and
// Term.ComparePlace accept them without an error: they are equal only to an
// unset value of their type, an unset term precedes every set term, List{}
// prints as "List(unset)", and Term{} prints as the empty string, which no
// name equals. Every other operation whose receiver or argument is unset
// returns an error wrapping [ErrUnset]. [Group] and [Guide] have no unset
// state: the zero Group is NoGroup, a valid value, and the zero Guide has no
// bounds and is not valid.
//
// # Errors
//
// Every error matches exactly one of four sentinels via [errors.Is]:
// [ErrUnset], [ErrInvalidArgument], [ErrNoTempo] for a number of beats that is
// zero or negative, and [ErrInvalidText] for a name that is not in the list.
// When several conditions fail, the first one in the documented order of the
// operation is reported. A failing call returns zero results, except that
// AppendText returns its argument unchanged. "No group", "no meaning", "no
// guide", and an empty result of a search are results, not errors.
//
// # Must variants
//
// Every fallible function and method X, except MarshalText and AppendText, has
// a variant MustX that returns the same results without the error and panics
// with the same error value on failure, such as
// term.EnglishWikipedia().MustParse("allegro").MustPlace().
//
// # Concurrency and comparison
//
// All values are immutable, and the package has no mutable state, caches, or
// goroutines; each built-in list is packed once on first use. No operation
// modifies its receiver or arguments; rows passed to NewList are only read
// and not retained, and slices returned are new and belong to the caller. All
// functions and methods are safe for concurrent use without synchronization.
// Group, Guide, Entry, List, and Term are comparable with ==, which matches
// Equal where a type has it, and can be used as map keys; Row is not
// comparable, as it holds a slice.
package term
