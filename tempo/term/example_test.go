package term_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo/term"
)

// Words of the default list (English Wikipedia): place, group, guide, and
// meanings; the synonym Vivo; the search by a number of beats.
func Example() {
	l := term.EnglishWikipedia()
	allegro := l.MustParse("allegro")
	guide, _ := allegro.MustGuide()
	english, _ := allegro.MustEnglish()
	russian, _ := allegro.MustRussian()
	fmt.Println(allegro, allegro.MustPlace(), allegro.MustGroup(), guide)
	fmt.Println(english, "/", russian)
	fmt.Println(l.MustAt(fraction.FromInt(115)), l.MustParse("Vivo"))
	fmt.Println(term.Default().Equal(l), len(term.BuiltIn()))
	// Output:
	// Allegro 15 fast 120-156
	// fast and bright / скоро
	// [Moderato Allegretto] Vivace
	// true 8
}

// The same name in two lists is two different words.
func ExampleRussianWikipedia() {
	en := term.EnglishWikipedia().MustParse("Allegro")
	ru := term.RussianWikipedia().MustParse("allegro")
	g1, _ := en.MustGuide()
	g2, _ := ru.MustGuide()
	fmt.Println(en.MustPlace(), g1, ru.MustPlace(), g2, en.Equal(ru))
	fmt.Println(term.RussianWikipedia().MustParse("Vivo").MustPlace())
	// Output:
	// 15 120-156 29 120-144 false
	// 34
}

// A user list: names are any non-empty UTF-8 text; a guide is optional.
func ExampleNewList() {
	l := term.MustNewList(
		term.Row{Entries: []term.Entry{{Name: "Медленно"}}, Group: term.Slow, Guide: new(term.MustBetween(fraction.FromInt(40), fraction.FromInt(60)))},
		term.Row{Entries: []term.Entry{{Name: "Ровно"}}, Guide: new(term.MustBetween(fraction.FromInt(100), fraction.FromInt(100)))},
		term.Row{Entries: []term.Entry{{Name: "Скоро"}, {Name: "Живо"}}, Group: term.Fast, Guide: new(term.MustBetween(fraction.FromInt(120), fraction.FromInt(160)))},
	)
	fmt.Println(l)
	fmt.Println(l.MustAt(fraction.FromInt(130)), l.MustAt(fraction.FromInt(80)) == nil)
	fmt.Println(l.MustParse("скоро").MustPlace(), l.MustParse("Ровно").MustGroup())
	// Output:
	// List[Медленно; Ровно; Скоро, Живо]
	// [Скоро Живо] true
	// 3 none
}

// Guides with open and equal bounds.
func ExampleGuide() {
	fmt.Println(term.MustAtMost(fraction.FromInt(24)), term.MustAtLeast(fraction.FromInt(200)),
		term.MustBetween(fraction.FromInt(100), fraction.FromInt(100)), term.Guide{})
	ok := term.MustAtLeast(fraction.FromInt(200)).MustContains(fraction.FromInt(200))
	fmt.Println(ok)
	// Output:
	// <=24 >=200 100-100 Guide(none)
	// true
}

// Failures are distinguished with errors.Is.
func ExampleList_Parse() {
	_, err := term.EnglishWikipedia().Parse("Fast")
	fmt.Println(errors.Is(err, term.ErrInvalidText))

	_, err = term.EnglishWikipedia().At(fraction.Zero())
	fmt.Println(errors.Is(err, term.ErrNoTempo))

	_, err = term.NewList()
	fmt.Println(errors.Is(err, term.ErrInvalidArgument))

	var l term.List
	_, err = l.Len()
	fmt.Println(l, errors.Is(err, term.ErrUnset))
	// Output:
	// true
	// true
	// true
	// List(unset) true
}
