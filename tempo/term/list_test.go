package term_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomList covers С-26 and architecture 4.4, "Свой перечень": places,
// groups, search by number, parse in the list.
func TestCustomList(t *testing.T) {
	t.Parallel()

	l, err := term.NewList(customRows()...)
	require.NoError(t, err)
	require.True(t, l.IsSet())

	assert.Equal(t, 3, l.MustLen())
	assert.Equal(t, []string{"Медленно", "Ровно", "Скоро", "Живо"}, names(t, l.MustTerms()))
	assert.Equal(t, "List[Медленно; Ровно; Скоро, Живо]", l.String())

	t.Run("С-26 places", func(t *testing.T) {
		t.Parallel()

		for name, place := range map[string]int{"Медленно": 1, "Ровно": 2, "Скоро": 3, "Живо": 3} {
			assert.Equal(t, place, l.MustParse(name).MustPlace(), name)
		}

		assert.Equal(t, []string{"Скоро", "Живо"}, names(t, l.MustPlace(3)))
	})

	t.Run("С-26 groups", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, term.Slow, l.MustParse("Медленно").MustGroup())
		assert.Equal(t, term.NoGroup, l.MustParse("Ровно").MustGroup())
		assert.Equal(t, term.Fast, l.MustParse("Живо").MustGroup())
		assert.Equal(t, []string{"Ровно"}, names(t, l.MustOfGroup(term.NoGroup)))
		assert.Equal(t, []string{"Скоро", "Живо"}, names(t, l.MustOfGroup(term.Fast)))
		assert.Nil(t, l.MustOfGroup(term.Moderate), "no words of the group is nil")
	})

	t.Run("С-26 search by number", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"Медленно"}, names(t, l.MustAt(fi(50))))
		assert.Equal(t, []string{"Ровно"}, names(t, l.MustAt(fi(100))))
		assert.Equal(t, []string{"Скоро", "Живо"}, names(t, l.MustAt(fi(130))))
		assert.Nil(t, l.MustAt(fi(80)), "an empty answer is nil, not an error")
	})

	t.Run("С-26 parse", func(t *testing.T) {
		t.Parallel()

		got, err := l.Parse("скоро")
		require.NoError(t, err)
		assert.Equal(t, "Скоро", got.MustName())

		_, err = l.Parse("Allegro")
		assertErr(t, err, term.ErrInvalidText)

		_, err = l.Parse("Vivo")
		assertErr(t, err, term.ErrInvalidText)
	})

	t.Run("values and missing guide", func(t *testing.T) {
		t.Parallel()

		m := term.MustNewList(
			term.Row{Entries: []term.Entry{{Name: "a", English: "e", Russian: "р"}, {Name: "b"}}},
			term.Row{Entries: []term.Entry{{Name: "c", Russian: "только"}}, Guide: guidePtr(atMost(10))},
		)

		a := m.MustParse("A")
		en, ok := a.MustEnglish()
		assert.True(t, ok)
		assert.Equal(t, "e", en)

		ru, ok := a.MustRussian()
		assert.True(t, ok)
		assert.Equal(t, "р", ru)

		_, ok = m.MustParse("b").MustEnglish()
		assert.False(t, ok)

		g, ok := a.MustGuide()
		assert.False(t, ok, "no guide")
		assert.Equal(t, term.Guide{}, g)
		assert.Equal(t, term.NoGroup, a.MustGroup(), "Row{} without a group means no group")

		_, ok = m.MustParse("c").MustEnglish()
		assert.False(t, ok)
		assert.Equal(t, []string{"c"}, names(t, m.MustAt(fi(5))), "words without a guide are not found")
	})
}

// TestCustomListNames covers architecture 4.4 and ВА-1: any non-empty UTF-8
// name is allowed and stored as given.
func TestCustomListNames(t *testing.T) {
	t.Parallel()

	odd := []string{
		"4=60", "a≈b", "Скоро, но не слишком", `"Presto"`, "q\"uote", " spaced ", "\t", "\x00", "a\\b",
		"line\nbreak", ", ", "=", "≈", "-", "😀", "ё",
	}

	rows := make([]term.Row, 0, len(odd))
	for _, n := range odd {
		rows = append(rows, term.Row{Entries: []term.Entry{{Name: n}}})
	}

	l, err := term.NewList(rows...)
	require.NoError(t, err)

	for i, n := range odd {
		w, err := l.Parse(n)
		if !assert.NoError(t, err, "%q", n) {
			continue
		}

		assert.Equal(t, n, w.MustName(), "%q is stored as given", n)
		assert.Equal(t, n, w.String())
		assert.Equal(t, i+1, w.MustPlace())

		text, err := w.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, n, string(text))
	}

	_, err = l.Parse("spaced")
	assertErr(t, err, term.ErrInvalidText)

	_, err = l.Parse(" SPACED ")
	assert.NoError(t, err, "case is ignored, spaces are not trimmed")
}

// TestNewListErrors covers С-42 and architecture 4.3, 4.4: every Е-2 condition
// of NewList. The result is List{}.
func TestNewListErrors(t *testing.T) {
	t.Parallel()

	one := func(names ...string) term.Row {
		r := term.Row{}
		for _, n := range names {
			r.Entries = append(r.Entries, term.Entry{Name: n})
		}

		return r
	}

	tests := []struct {
		name string
		rows []term.Row
	}{
		{"С-42 no rows", nil},
		{"empty rows slice", []term.Row{}},
		{"row without terms", []term.Row{one("a"), {}}},
		{"row with an empty Entries slice", []term.Row{{Entries: []term.Entry{}}}},
		{"С-42 empty name", []term.Row{one("")}},
		{"empty name among others", []term.Row{one("a", "")}},
		{"С-42 name not in UTF-8", []term.Row{one("\xff")}},
		{"name with a truncated rune", []term.Row{one("a\xd0")}},
		{"С-42 Скоро and скоро", []term.Row{one("Скоро"), one("скоро")}},
		{"repeat in one row", []term.Row{one("Allegro", "ALLEGRO")}},
		{"exact repeat", []term.Row{one("a"), one("b"), one("a")}},
		{"group out of range", []term.Row{{Entries: []term.Entry{{Name: "a"}}, Group: term.Group(4)}}},
		{"С-42 guide without bounds", []term.Row{{Entries: []term.Entry{{Name: "a"}}, Guide: &term.Guide{}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l, err := term.NewList(tt.rows...)
			assertErr(t, err, term.ErrInvalidArgument)
			assert.Equal(t, term.List{}, l)
			assert.False(t, l.IsSet())
		})
	}

	t.Run("two different invalid byte names", func(t *testing.T) {
		t.Parallel()

		_, err := term.NewList(one("\xff"), one("\xfe"))
		assertErr(t, err, term.ErrInvalidArgument)
	})
}

// TestListRead covers architecture 4.4, "Чтение перечня": Len, Place, Terms,
// OfGroup and their Е-2 conditions.
func TestListRead(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()

	for _, p := range []int{0, -1, 21, 1000} {
		got, err := en.Place(p)
		assertErr(t, err, term.ErrInvalidArgument)
		assert.Nil(t, got, "place %d", p)
	}

	assert.Equal(t, []string{"Larghissimo"}, names(t, en.MustPlace(1)))
	assert.Equal(t, []string{"Prestissimo"}, names(t, en.MustPlace(20)))
	assert.Equal(t, []string{"Vivacissimo", "Allegrissimo"}, names(t, en.MustPlace(18)))

	for _, g := range []term.Group{term.Group(4), term.Group(200)} {
		got, err := en.OfGroup(g)
		assertErr(t, err, term.ErrInvalidArgument)
		assert.Nil(t, got)
	}

	assert.Nil(t, en.MustOfGroup(term.NoGroup), "list 1 has no word without a group")
	assert.Equal(t, []string{"Maestoso", "Animato", "Assai"}, names(t, term.JapaneseMetronome().MustOfGroup(term.NoGroup)))
	assert.Equal(t,
		[]string{"comodo", "comodamente", "con moto", "animato", "animato assai", "tempo di marcia"},
		names(t, term.RussianWikipedia().MustOfGroup(term.NoGroup)))
	assert.Equal(t,
		[]string{"Larghissimo", "Adagissimo", "Grave", "Largo", "Larghetto", "Lento", "Adagio", "Adagietto"},
		names(t, en.MustOfGroup(term.Slow)))

	for _, want := range builtIns() {
		l := want.get()

		var total int

		for _, g := range append([]term.Group{term.NoGroup}, term.Groups()...) {
			total += len(l.MustOfGroup(g))
		}

		assert.Equal(t, want.words, total, "%s: every word is in exactly one group", want.name)
	}
}

// TestListParse covers architecture 4.4: case-insensitive match
// (strings.EqualFold), no trimming, Е-10 for unknown names, "" and text not
// in UTF-8, and С-41 for Fast and Larghissimo.
func TestListParse(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	ru := term.RussianWikipedia()

	for _, x := range []string{"allegro", "ALLEGRO", "Allegro", "aLlEgRo"} {
		got, err := en.Parse(x)
		require.NoError(t, err, x)
		assert.Equal(t, "Allegro", got.MustName(), "С-27: %s", x)
	}

	assert.Equal(t, "Marcia moderato", en.MustParse("MARCIA MODERATO").MustName())

	bad := []struct {
		l    term.List
		name string
	}{
		{en, "Fast"},
		{ru, "Larghissimo"},
		{en, ""},
		{en, " Allegro"},
		{en, "Allegro "},
		{en, "Allegro\n"},
		{en, "Allegro  moderato"},
		{en, "\xff"},
		{en, "Allegro\xff"},
		{en, "List(unset)"},
		{en, "Allegretto; see paragraph above"},
		{ru, "Allegro moderato\x00"},
	}

	for _, tt := range bad {
		got, err := tt.l.Parse(tt.name)
		assertErr(t, err, term.ErrInvalidText)
		assert.Equal(t, term.Term{}, got, "%q", tt.name)
	}
}

// TestListAt covers С-24, С-44, С-57 and architecture 4.4: all words whose
// guide contains the number, in list order, bounds included, open bounds,
// empty answer nil.
func TestListAt(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()

	tests := []struct {
		name  string
		l     term.List
		beats fraction.Fraction
		want  []string
	}{
		{"С-24 115", en, fi(115), []string{"Moderato", "Allegretto"}},
		{"С-24 130", en, fi(130), []string{"Allegro", "Molto Allegro", "Allegro vivace"}},
		{"С-24 120 bounds included", en, fi(120), []string{"Moderato", "Allegretto", "Allegro moderato", "Allegro"}},
		{"С-24 24", en, fi(24), []string{"Larghissimo", "Adagissimo", "Grave"}},
		{"С-24 10", en, fi(10), []string{"Larghissimo"}},
		{"С-24 250", en, fi(250), []string{"Prestissimo"}},
		{"С-24 60", en, fi(60), []string{"Largo", "Larghetto", "Lento", "Adagio", "Adagietto", "Andante"}},
		{"200 both bounds", en, fi(200), []string{"Presto", "Prestissimo"}},
		{"tiny", en, invMaxF(), []string{"Larghissimo"}},
		{"huge", en, maxF(), []string{"Prestissimo"}},
		{"fraction 49/2", en, fr(49, 2), []string{"Adagissimo", "Grave"}},
		{"С-24 100 Russian", term.RussianWikipedia(), fi(100),
			[]string{"con moto", "allegretto moderato", "allegretto", "allegretto mosso", "animato"}},
		{"С-24 100 Japanese", term.JapaneseMetronome(), fi(100), []string{"Moderato"}},
		{"С-24 100 German", term.GermanMetronome(), fi(100), []string{"Andante"}},
		{"С-24 100 French", term.FrenchMetronome(), fi(100), []string{"Adagio"}},
		{"С-24 100 Maelzel", term.MaelzelMetronome(), fi(100), []string{"allegretto"}},
		{"С-24 100 1950", term.Metronome1950(), fi(100), []string{"allegretto"}},
		{"С-24 100 electronic", term.ElectronicMetronome(), fi(100), []string{"andante"}},
		{"С-24 30 Russian empty", term.RussianWikipedia(), fi(30), nil},
		{"С-24 210 Russian open bound", term.RussianWikipedia(), fi(210), []string{"prestissimo"}},
		{"Japanese gap 200-208 empty", term.JapaneseMetronome(), fi(204), nil},
		{"Japanese open bound 208", term.JapaneseMetronome(), fi(208), []string{"Prestissimo"}},
		{"1950 between values empty", term.Metronome1950(), fi(101), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.l.At(tt.beats)
			require.NoError(t, err)

			if tt.want == nil {
				assert.Nil(t, got)

				return
			}

			assert.Equal(t, tt.want, names(t, got))
		})
	}

	t.Run("С-57 list 1 answers every number", func(t *testing.T) {
		t.Parallel()

		for n := 1; n <= 400; n++ {
			for _, b := range []fraction.Fraction{fi(n), fr(2*n-1, 2), fr(3*n+1, 3)} {
				got, err := en.At(b)
				require.NoError(t, err)
				assert.NotEmpty(t, got, "At(%s)", b)
			}
		}
	})

	t.Run("С-44 errors", func(t *testing.T) {
		t.Parallel()

		for _, b := range []fraction.Fraction{fi(0), fi(-5), fr(-1, 2)} {
			got, err := en.At(b)
			assertErr(t, err, term.ErrNoTempo)
			assert.Nil(t, got)
		}

		got, err := en.At(fraction.Fraction{})
		assertErr(t, err, term.ErrInvalidArgument)
		assert.Nil(t, got)
	})
}

// TestListEqual covers architecture 3.1, 4.4 and КР-TP-15, ВА-2: equality by
// contents in order, including meanings, groups, guides, and parse synonyms;
// it matches ==.
func TestListEqual(t *testing.T) {
	t.Parallel()

	base := func() []term.Row {
		return []term.Row{
			{Entries: []term.Entry{{Name: "a", English: "e", Russian: "r"}}, Group: term.Slow, Guide: guidePtr(between(1, 2))},
			{Entries: []term.Entry{{Name: "b"}, {Name: "c"}}},
		}
	}

	variants := map[string]func(rows []term.Row){
		"name":           func(r []term.Row) { r[0].Entries[0].Name = "x" },
		"name case":      func(r []term.Row) { r[0].Entries[0].Name = "A" },
		"English":        func(r []term.Row) { r[0].Entries[0].English = "" },
		"Russian":        func(r []term.Row) { r[0].Entries[0].Russian = "R" },
		"group":          func(r []term.Row) { r[0].Group = term.Fast },
		"guide":          func(r []term.Row) { r[0].Guide = guidePtr(between(1, 3)) },
		"no guide":       func(r []term.Row) { r[0].Guide = nil },
		"open guide":     func(r []term.Row) { r[0].Guide = guidePtr(atLeast(1)) },
		"word order":     func(r []term.Row) { r[1].Entries[0], r[1].Entries[1] = r[1].Entries[1], r[1].Entries[0] },
		"row order":      func(r []term.Row) { r[0], r[1] = r[1], r[0] },
		"split row":      func(r []term.Row) { r[1].Entries = r[1].Entries[:1] },
		"no meanings":    func(r []term.Row) { r[0].Entries[0] = term.Entry{Name: "a"} },
		"guide fraction": func(r []term.Row) { r[0].Guide = guidePtr(term.MustBetween(fr(1, 2), fi(2))) },
	}

	a := term.MustNewList(base()...)
	b := term.MustNewList(base()...)
	assert.True(t, a.Equal(b))
	assert.Equal(t, a, b, "equal lists are ==")

	same := base()
	same[0].Guide = guidePtr(term.MustBetween(fraction.MustNew(2, 2), fraction.MustNew(4, 2)))
	assert.Equal(t, a, term.MustNewList(same...), "guide bounds are compared by value")

	values := []term.List{{}, a, term.EnglishWikipedia(), term.RussianWikipedia()}

	for name, change := range variants {
		rows := base()
		change(rows)

		c := term.MustNewList(rows...)
		assert.False(t, a.Equal(c), name)
		assert.NotEqual(t, a, c, name)

		values = append(values, c)
	}

	for _, x := range values {
		for _, y := range values {
			assert.Equal(t, x == y, x.Equal(y), "%s vs %s", x, y)
		}
	}

	m := map[term.List]int{a: 1}
	m[b]++
	assert.Equal(t, 2, m[a], "a List is usable as a map key")
}

// TestListEqualSynonym covers ВА-2 (business-spec 2.1, architecture 4.4): a
// user list with the rows of list 1 has no synonym Vivo, so it is not equal to
// list 1 and does not parse Vivo; its words differ from the words of list 1.
func TestListEqualSynonym(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	mine := term.MustNewList(rowsOf(list1)...)

	assert.Equal(t, en.String(), mine.String(), "the same rows")
	assert.Equal(t, names(t, en.MustTerms()), names(t, mine.MustTerms()))
	assert.False(t, en.Equal(mine))
	assert.False(t, mine.Equal(en))
	assert.NotEqual(t, en, mine)

	_, err := mine.Parse("Vivo")
	assertErr(t, err, term.ErrInvalidText)

	assert.Equal(t, "Vivace", en.MustParse("Vivo").MustName())

	a := mine.MustParse("Allegro")
	assert.False(t, a.Equal(en.MustParse("Allegro")), "words of different lists differ")
	assert.Equal(t, en.MustParse("Allegro").MustPlace(), a.MustPlace())

	_, err = a.ComparePlace(en.MustParse("Largo"))
	assertErr(t, err, term.ErrInvalidArgument)

	mine2 := term.MustNewList(rowsOf(list2)...)
	assert.Equal(t, term.RussianWikipedia(), mine2, "list 2 has no synonyms: the same rows give an equal list")
	assert.True(t, term.RussianWikipedia().Equal(mine2))

	for i, want := range builtIns()[2:] {
		got := term.MustNewList(rowsOf(want.places)...)
		assert.True(t, want.get().Equal(got), "list %d has no synonyms", i+3)
	}
}
