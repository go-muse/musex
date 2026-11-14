# Tempo terms

A **tempo term** is a word that names a tempo in a score: Largo, Andante, Allegro, Presto. Lists of such words give each one a place, a **group** — slow, moderate, or fast tempos — a **guide**, which is a range of beats per minute, and sometimes its meaning. The library offers eight such lists and lets you make your own. The [tempo](../tempo.md) package uses them to find the words that fit a tempo marking and to write indications such as "Allegro, ♩ = 120".

## Places and words

A list is a sequence of places numbered from 1. A place usually holds one word, but can hold several that share a group and a guide: in the list of the English Wikipedia, Adagissimo and Grave share place 2. A word belongs to its list. The same name in two lists is two different words with their own places and guides: Allegro is at place 15 with 120–156 in one list and at place 29 with 120–144 in another.

Places follow the rows of a source, not the tempos. Comparing the places of two words — Largo at 3 comes before Allegro at 15 — compares rows. The English list calls its order approximate, and its numbers are not monotonic: Lento (40–60) follows Larghetto (44–66), and Presto (168–200) follows Vivacissimo (172–176). The Russian list does not claim an order by speed at all. Words of different lists cannot be compared by place.

## The eight lists

| List | Source | Places | Words | Meanings |
| --- | --- | --- | --- | --- |
| English Wikipedia (default) | article "Tempo", section "Approximately from the slowest to the fastest" | 20 | 23 | English from the article; Russian from the textbook of Alekseev and Myasoedov where it gives them |
| Russian Wikipedia | article "Темп (музыка)", table "Темповые обозначения" | 37 | 39 | Russian from the article |
| Japanese metronome, 21st century | German Wikipedia, article "Tempo (Musik)" | 17 | 17 | none |
| German metronome, 20th century | German Wikipedia, article "Tempo (Musik)" | 8 | 8 | none |
| French metronome, 19th century | German Wikipedia, article "Tempo (Musik)" | 6 | 6 | none |
| Maelzel metronome, 19th century | Dolmetsch, table of tempo markings | 10 | 10 | none |
| Metronome of 1950 | Dolmetsch, table of tempo markings | 12 | 12 | none |
| Modern electronic metronome | Dolmetsch, table of tempo markings | 8 | 8 | none |

Names and meanings are copied from the sources, verbatim; the lists of the Russian Wikipedia and of the three Dolmetsch metronomes are in lower case, as their sources write them. Each metronome is a list of its own because each places the words differently: Andante is 69–76 on one, 76–108 on another, and 124–152 on a third. Words in other languages and Russian names of the words are not included; a user list can add them.

## Guides are references, not tempos

A guide is the range of numbers a list gives for a word. Both ends belong to it, and one end may be open: Larghissimo is "up to 24" and Prestissimo "from 200". The metronomes of Dolmetsch often name a single number, which becomes a guide whose two ends are equal.

The numbers come from **weak sources**. They are rough approximations, and other sources give other numbers for the same word — Presto is 168–200 in one list, 184–200 in another, and 100–152 elsewhere. A word alone therefore does not give a tempo: the library never derives a marking from a word, and a guide is reference data, not a tempo of the word. Where a different source matters to you, make a list of your own.

The numbers name beats without a note value. The two Wikipedia sources give theirs for a quarter, and the metronomes give none, so the library treats every list the same way: a number is compared as it is.

## Finding words by number

Asking a list which words fit a number returns every word whose guide contains it, in the order of the list, because guides overlap and the library does not pick one. In the English list, 115 gives Moderato and Allegretto; 120 gives Moderato, Allegretto, Allegro moderato, and Allegro, since both ends belong to a guide; 10 gives Larghissimo and 250 Prestissimo. Thanks to its open ends, the English list has at least one word for every positive number. Elsewhere the answer can be empty — the Russian list has nothing for 30 — and an empty answer is a result, not an error. Words without a guide are never found by number.

The same number gives different words in different lists: 100 is con moto, allegretto moderato, allegretto, allegretto mosso, and animato in the Russian list, Moderato on the Japanese metronome, Andante on the German one, Adagio on the French one, and allegretto or andante on the others.

## Groups

Every word of the built-in lists has the group of its **main word**, after the textbook of Alekseev and Myasoedov: Adagio, Largo, Lento, and Grave are slow; Andante, Moderato, and Sostenuto moderate; Allegro, Vivo (vivace), and Presto fast. Extending this to the other words is a convention of the library:

- a **derived** word takes the group of the word it comes from: Larghetto and Larghissimo are slow, Andantino moderate, Allegretto, Vivacissimo, and Prestissimo fast;
- a **compound** term takes the group of its first main word: Marcia moderato and Andante moderato are moderate, Molto Allegro is fast, and so is **Allegro moderato**, although other sources treat moderato as a moderate tempo;
- a word with no main word — comodo, con moto, animato, tempo di marcia, Maestoso, Assai — has **no group**, which is an answer, not an error.

The sources disagree on groups of single words too: one puts Allegretto among the fast tempos, another among the moderate ones. The groups of the built-in lists follow the rule above; in a list of your own you assign groups yourself.

## Other conventions

- **Vivo.** In the English list, reading the name Vivo gives Vivace, as several sources treat the two words as one. In the Russian list vivo (160–176) and vivace (176–192) are different words, as in its source.
- **Allegretto.** Its English meaning in the English list is "by the mid-19th century, moderately fast"; the rest of the source text only refers to another paragraph of the article.
- **prestissimo in the Russian list.** Its name leaves out the source's mark "(possible)", and its guide "192–208 and above" is kept as "from 192" with no upper end.

## The default list and your own lists

The English list is the default, because it is the only source that gives numbers for every word and states an order from slow to fast. Nothing uses it behind your back, though: every operation that needs a list is given one explicitly.

A list of your own is made of rows, each with one or more names, a group or none, a guide or none, and optional English and Russian meanings. A name can be any non-empty text, including commas, quotation marks, and equals signs; two names that differ only in case are not allowed in one list. A guide needs at least one end, its numbers must be positive, and the lower end must not exceed the upper one; rows need not be in order of their numbers. Your list works with every operation of a built-in list. It never equals the English list, even with the same rows, because only the English list reads Vivo.

## Reading names

A name is read in a given list, ignoring the case of letters but not spaces: allegro and ALLEGRO both give Allegro, while Fast in the English list or Larghissimo in the Russian list are not words of those lists and are refused. A word prints as its name, and reading that name in its list gives the word back. Because a name means a word only within a list, a word is always read together with its list, never on its own.

When a long name is rejected, the error quotes at most its first 64 bytes, without cutting a character in half.

## The unset value

An empty list or word means that nothing is set. It can be compared — an empty word comes before every word in place order and equals only another empty word — and an empty list prints as a recognizable placeholder, while an empty word prints as an empty name, which no word has; every other operation with them reports an error.

Lists and words are immutable values that are safe to copy, to use as keys, and to use from many goroutines at once.

First released in v0.26.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/tempo/term) · [Go examples](../../../tempo/term/example_test.go) · [Roadmap](../../../ROADMAP.md)
