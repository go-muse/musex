package term_test

import (
	"testing"

	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuiltInContents covers business-spec 2.4 and architecture 4.5 for the
// eight lists in full (Т-16, С-23): number of places and words, names in
// order, places, groups, guides including open and equal bounds, English and
// Russian meanings verbatim (lists 1 and 2), no meanings (lists 3-8).
func TestBuiltInContents(t *testing.T) {
	t.Parallel()

	for _, want := range builtIns() {
		t.Run(want.name, func(t *testing.T) {
			t.Parallel()

			l := want.get()
			require.True(t, l.IsSet())

			n, err := l.Len()
			require.NoError(t, err)
			assert.Equal(t, len(want.places), n, "places")

			all, err := l.Terms()
			require.NoError(t, err)
			assert.Len(t, all, want.words, "words")

			var flat []string
			for _, p := range want.places {
				flat = append(flat, p.names...)
			}

			assert.Equal(t, flat, names(t, all), "names of Terms in order")

			for i, p := range want.places {
				place := i + 1

				got, err := l.Place(place)
				require.NoError(t, err, "place %d", place)

				if !assert.Equal(t, p.names, names(t, got), "place %d", place) {
					continue
				}

				for j, w := range got {
					checkWord(t, l, w, p, place, j)
				}
			}

			assert.Equal(t, stringOf(want.places), l.String())
		})
	}
}

// checkWord checks one word of a built-in list against its expected place.
func checkWord(t *testing.T, l term.List, w term.Term, p wantPlace, place, j int) {
	t.Helper()

	name := p.names[j]
	assert.True(t, w.IsSet(), name)
	assert.Equal(t, name, w.String(), name)

	gotList, err := w.List()
	require.NoError(t, err, name)
	assert.True(t, gotList.Equal(l), "%s: List", name)

	gotPlace, err := w.Place()
	require.NoError(t, err, name)
	assert.Equal(t, place, gotPlace, "%s: Place", name)

	g, err := w.Group()
	require.NoError(t, err, name)
	assert.Equal(t, p.group, g, "%s: Group", name)

	guide, ok, err := w.Guide()
	require.NoError(t, err, name)
	assert.True(t, ok, "%s: every built-in word has a guide", name)
	assert.Equal(t, guideOf(p.lo, p.hi), guide, "%s: Guide %s", name, guide)

	en, ok, err := w.English()
	require.NoError(t, err, name)
	assert.Equal(t, p.english != "", ok, "%s: English present", name)
	assert.Equal(t, p.english, en, "%s: English", name)

	ru, ok, err := w.Russian()
	require.NoError(t, err, name)
	assert.Equal(t, p.russian[j] != "", ok, "%s: Russian present", name)
	assert.Equal(t, p.russian[j], ru, "%s: Russian", name)

	parsed, err := l.Parse(name)
	require.NoError(t, err, name)
	assert.Equal(t, w, parsed, "%s: Parse by its name", name)
	assert.True(t, parsed.Equal(w), name)
}

// TestBuiltInFunctions covers architecture 4.4: Default equals
// EnglishWikipedia, BuiltIn returns the eight lists in the order 1-8 as a new
// slice, every call returns an equal value, and the eight lists differ.
func TestBuiltInFunctions(t *testing.T) {
	t.Parallel()

	assert.Equal(t, term.EnglishWikipedia(), term.Default())
	assert.True(t, term.Default().Equal(term.EnglishWikipedia()))

	want := builtIns()
	got := term.BuiltIn()
	require.Len(t, got, 8)

	for i, w := range want {
		assert.Equal(t, w.get(), got[i], "BuiltIn()[%d] is %s", i, w.name)
		assert.Equal(t, w.get(), w.get(), "%s: every call returns an equal value", w.name)
		assert.True(t, w.get().Equal(got[i]), w.name)

		for j := range want {
			if j != i {
				assert.False(t, got[i].Equal(got[j]), "%s and %s differ", w.name, want[j].name)
				assert.NotEqual(t, got[i], got[j])
			}
		}
	}

	got[0] = term.List{}
	assert.Equal(t, term.EnglishWikipedia(), term.BuiltIn()[0], "the returned slice belongs to the caller")
}

// TestBuiltInScenarios covers the values named in С-23.
func TestBuiltInScenarios(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	ru := term.RussianWikipedia()

	t.Run("С-23 Allegro of list 1", func(t *testing.T) {
		t.Parallel()

		w := en.MustParse("Allegro")
		assert.Equal(t, 15, w.MustPlace())
		assert.Equal(t, term.Fast, w.MustGroup())

		g, ok := w.MustGuide()
		assert.True(t, ok)
		assert.Equal(t, between(120, 156), g)

		e, ok := w.MustEnglish()
		assert.True(t, ok)
		assert.Equal(t, "fast and bright", e)

		r, ok := w.MustRussian()
		assert.True(t, ok)
		assert.Equal(t, "скоро", r)
	})

	t.Run("С-23 Allegretto shortened English meaning", func(t *testing.T) {
		t.Parallel()

		e, ok := en.MustParse("Allegretto").MustEnglish()
		assert.True(t, ok)
		assert.Equal(t, "by the mid-19th century, moderately fast", e)
	})

	t.Run("С-23 Larghissimo open lower bound, no Russian meaning", func(t *testing.T) {
		t.Parallel()

		w := en.MustParse("Larghissimo")
		g, ok := w.MustGuide()
		assert.True(t, ok)
		assert.Equal(t, atMost(24), g)

		lo, ok := g.Lower()
		assert.False(t, ok)
		assert.False(t, lo.IsValid(), "an open bound is fraction.Fraction{}")

		hi, ok := g.Upper()
		assert.True(t, ok)
		assert.True(t, hi.EqualExact(fi(24)))

		r, ok := w.MustRussian()
		assert.False(t, ok)
		assert.Empty(t, r)
	})

	t.Run("С-23 Grave and Adagissimo share place 2", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 2, en.MustParse("Grave").MustPlace())
		assert.Equal(t, 2, en.MustParse("Adagissimo").MustPlace())
		assert.Equal(t, 0, en.MustParse("Grave").MustComparePlace(en.MustParse("Adagissimo")))
	})

	t.Run("С-23 moderate words of list 1", func(t *testing.T) {
		t.Parallel()

		got := en.MustOfGroup(term.Moderate)
		assert.Equal(t, []string{"Andante", "Andantino", "Marcia moderato", "Andante moderato", "Moderato"},
			names(t, got))
	})

	t.Run("С-23 places of Largo and Allegro compare places", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 3, en.MustParse("Largo").MustPlace())
		assert.Equal(t, -1, en.MustParse("Largo").MustComparePlace(en.MustParse("Allegro")))
		assert.Equal(t, 1, en.MustParse("Allegro").MustComparePlace(en.MustParse("Largo")))
		// Lento (40-60) stands after Larghetto (44-66): places, not speeds.
		assert.Equal(t, 1, en.MustParse("Lento").MustComparePlace(en.MustParse("Larghetto")))
	})

	t.Run("С-23 list 2", func(t *testing.T) {
		t.Parallel()

		w := ru.MustParse("allegro")
		assert.Equal(t, 29, w.MustPlace())
		assert.Equal(t, term.Fast, w.MustGroup())

		g, _ := w.MustGuide()
		assert.Equal(t, between(120, 144), g)

		r, ok := w.MustRussian()
		assert.True(t, ok)
		assert.Equal(t, "скорый темп (буквально: «весело»)", r)

		e, ok := w.MustEnglish()
		assert.False(t, ok, "list 2 has no English meanings")
		assert.Empty(t, e)

		assert.Equal(t, 34, ru.MustParse("vivo").MustPlace())
		g, _ = ru.MustParse("vivo").MustGuide()
		assert.Equal(t, between(160, 176), g)
		assert.Equal(t, 35, ru.MustParse("vivace").MustPlace())
		g, _ = ru.MustParse("vivace").MustGuide()
		assert.Equal(t, between(176, 192), g)

		for _, n := range []string{"comodo", "comodamente"} {
			assert.Equal(t, 13, ru.MustParse(n).MustPlace(), n)
			assert.Equal(t, term.NoGroup, ru.MustParse(n).MustGroup(), n)
		}

		n, err := ru.Len()
		require.NoError(t, err)
		assert.Equal(t, 37, n)
		assert.Len(t, ru.MustTerms(), 39)
	})

	t.Run("П-29 prestissimo of list 2", func(t *testing.T) {
		t.Parallel()

		w := ru.MustParse("prestissimo")
		assert.Equal(t, "prestissimo", w.MustName())
		assert.Equal(t, 37, w.MustPlace())

		g, _ := w.MustGuide()
		assert.Equal(t, atLeast(192), g)

		_, ok := g.Upper()
		assert.False(t, ok, "the upper bound is open")

		_, err := ru.Parse("prestissimo(possible)")
		assertErr(t, err, term.ErrInvalidText)
	})

	t.Run("С-23 Japanese metronome", func(t *testing.T) {
		t.Parallel()

		l := term.JapaneseMetronome()
		assert.Equal(t, 17, l.MustLen())
		assert.Equal(t, term.NoGroup, l.MustParse("Maestoso").MustGroup())

		g, _ := l.MustParse("Prestissimo").MustGuide()
		assert.Equal(t, atLeast(208), g)

		for _, w := range l.MustTerms() {
			_, ok := w.MustEnglish()
			assert.False(t, ok)
			_, ok = w.MustRussian()
			assert.False(t, ok)
		}
	})

	t.Run("С-23 metronome of 1950 equal bounds", func(t *testing.T) {
		t.Parallel()

		g, _ := term.Metronome1950().MustParse("allegretto").MustGuide()
		assert.Equal(t, between(100, 100), g)

		lo, _ := g.Lower()
		hi, _ := g.Upper()
		assert.True(t, lo.EqualExact(fi(100)))
		assert.True(t, hi.EqualExact(fi(100)))
	})

	t.Run("С-23 the same name in two lists is two words", func(t *testing.T) {
		t.Parallel()

		a1 := en.MustParse("Allegro")
		a2 := ru.MustParse("allegro")
		assert.False(t, a1.Equal(a2))
		assert.NotEqual(t, a1, a2)
		assert.Equal(t, 15, a1.MustPlace())
		assert.Equal(t, 29, a2.MustPlace())

		g1, _ := a1.MustGuide()
		g2, _ := a2.MustGuide()
		assert.Equal(t, between(120, 156), g1)
		assert.Equal(t, between(120, 144), g2)

		// The same name in two metronome lists is two words.
		assert.False(t, term.GermanMetronome().MustParse("Largo").Equal(term.FrenchMetronome().MustParse("Largo")))
	})

	t.Run("lists 6 and 7 keep the source order", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 2, term.MaelzelMetronome().MustParse("adagio").MustPlace())
		assert.Equal(t, 3, term.MaelzelMetronome().MustParse("lento").MustPlace())
		assert.Equal(t, 9, term.Metronome1950().MustParse("vivace").MustPlace())
		assert.Equal(t, 10, term.Metronome1950().MustParse("allegro").MustPlace())

		g, _ := term.MaelzelMetronome().MustParse("prestissimo").MustGuide()
		assert.Equal(t, between(184, 240), g)

		g, _ = term.ElectronicMetronome().MustParse("andante").MustGuide()
		assert.Equal(t, between(80, 100), g)
	})
}

// TestVivoSynonym covers П-12: Vivo in list 1 parses as Vivace, ignoring
// case; in list 2 vivo and vivace are different words; the synonym is not a
// name of list 1 and is not among its terms.
func TestVivoSynonym(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	vivace := en.MustParse("Vivace")

	for _, x := range []string{"Vivo", "vivo", "VIVO", "vIvO"} {
		got, err := en.Parse(x)
		require.NoError(t, err, x)
		assert.Equal(t, vivace, got, x)
		assert.Equal(t, "Vivace", got.MustName(), x)
	}

	for _, w := range en.MustTerms() {
		assert.NotEqual(t, "Vivo", w.MustName())
	}

	ru := term.RussianWikipedia()
	vivo := ru.MustParse("Vivo")
	assert.Equal(t, "vivo", vivo.MustName())
	assert.Equal(t, 34, vivo.MustPlace())
	assert.False(t, vivo.Equal(ru.MustParse("vivace")))

	for _, l := range term.BuiltIn()[2:] {
		_, err := l.Parse("Vivo")
		assertErr(t, err, term.ErrInvalidText)
	}
}
