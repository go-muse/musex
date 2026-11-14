package term_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTermEqual covers architecture 3.1, 4.4: the same list and the same
// name; it matches ==; terms are usable as map keys.
func TestTermEqual(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()

	allegro := en.MustParse("Allegro")
	assert.Equal(t, allegro, term.EnglishWikipedia().MustParse("allegro"), "terms of equal lists are equal")
	assert.Equal(t, allegro, en.MustPlace(15)[0])
	assert.Equal(t, allegro, term.Default().MustParse("ALLEGRO"))

	values := []term.Term{
		{}, allegro, en.MustParse("Grave"), en.MustParse("Adagissimo"), en.MustParse("Vivace"),
		term.RussianWikipedia().MustParse("allegro"), customList().MustParse("Скоро"), customList().MustParse("Живо"),
		term.MustNewList(rowsOf(list1)...).MustParse("Allegro"),
	}

	for _, a := range values {
		for _, b := range values {
			assert.Equal(t, a == b, a.Equal(b), "%q vs %q", a, b)
		}
	}

	assert.Equal(t, en.MustParse("Vivo"), en.MustParse("Vivace"), "the synonym gives the same term")

	m := map[term.Term]int{allegro: 1}
	m[en.MustParse("allegro")]++
	assert.Equal(t, 2, m[allegro])
}

// TestTermComparePlace covers architecture 4.4: sign of the place difference,
// 0 for one place, unset before set without an error (С-29), Е-2 for words of
// different lists (С-45).
func TestTermComparePlace(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	c := customList()

	tests := []struct {
		name string
		a, b term.Term
		want int
	}{
		{"before", en.MustParse("Largo"), en.MustParse("Allegro"), -1},
		{"after", en.MustParse("Prestissimo"), en.MustParse("Larghissimo"), 1},
		{"itself", en.MustParse("Largo"), en.MustParse("Largo"), 0},
		{"one place", en.MustParse("Molto Allegro"), en.MustParse("Allegro vivace"), 0},
		{"custom one place", c.MustParse("Скоро"), c.MustParse("Живо"), 0},
		{"custom", c.MustParse("Ровно"), c.MustParse("Медленно"), 1},
		{"unset before set", term.Term{}, en.MustParse("Larghissimo"), -1},
		{"set after unset", en.MustParse("Larghissimo"), term.Term{}, 1},
		{"two unset", term.Term{}, term.Term{}, 0},
		{"unset and a custom term", term.Term{}, c.MustParse("Скоро"), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.a.ComparePlace(tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.a.MustComparePlace(tt.b))
		})
	}

	errTests := []struct {
		name string
		a, b term.Term
	}{
		{"С-45 list 1 and a custom list", en.MustParse("Allegro"), c.MustParse("Скоро")},
		{"С-45 Allegro of lists 1 and 2", en.MustParse("Allegro"), term.RussianWikipedia().MustParse("allegro")},
		{"list 1 and its copy without the synonym", en.MustParse("Largo"), term.MustNewList(rowsOf(list1)...).MustParse("Largo")},
	}

	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.a.ComparePlace(tt.b)
			assertErr(t, err, term.ErrInvalidArgument)
			assert.Equal(t, 0, got)

			_, err = tt.b.ComparePlace(tt.a)
			assertErr(t, err, term.ErrInvalidArgument)
		})
	}
}

// TestTermText covers architecture 4.4: String, MarshalText, and AppendText
// give the name; l.Parse(w.String()) returns w (4.14, "Обратимость").
func TestTermText(t *testing.T) {
	t.Parallel()

	for _, want := range builtIns() {
		l := want.get()
		for _, w := range l.MustTerms() {
			assert.Equal(t, w, l.MustParse(w.String()), "%s: %s", want.name, w)

			text, err := w.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, w.String(), string(text))

			dst := []byte("prefix:")
			got, err := w.AppendText(dst)
			require.NoError(t, err)
			assert.Equal(t, "prefix:"+w.String(), string(got))
		}
	}

	c := customList()
	for _, w := range c.MustTerms() {
		assert.Equal(t, w, c.MustParse(w.String()))
	}
}

// TestUnset covers С-29, С-32 and architecture 4.3 for List{} and Term{}:
// only IsSet, Equal, ComparePlace and String work; everything else is Е-1
// with zero results.
func TestUnset(t *testing.T) {
	t.Parallel()

	var (
		l term.List
		w term.Term
	)

	assert.False(t, l.IsSet())
	assert.False(t, w.IsSet())
	assert.True(t, l.Equal(term.List{}))
	assert.False(t, l.Equal(term.EnglishWikipedia()))
	assert.False(t, term.EnglishWikipedia().Equal(l))
	assert.True(t, w.Equal(term.Term{}))
	assert.False(t, w.Equal(term.EnglishWikipedia().MustParse("Largo")))
	assert.False(t, term.EnglishWikipedia().MustParse("Largo").Equal(w))
	assert.Equal(t, "List(unset)", l.String())
	assert.Empty(t, w.String(), "КР-TP-16: the special string of Term{} is empty")

	unsetCalls := []struct {
		name string
		call func() (any, error)
		zero any
	}{
		{"List.Len", func() (any, error) { return l.Len() }, 0},
		{"List.Place", func() (any, error) { return l.Place(1) }, []term.Term(nil)},
		{"List.Place out of range", func() (any, error) { return l.Place(0) }, []term.Term(nil)},
		{"List.Terms", func() (any, error) { return l.Terms() }, []term.Term(nil)},
		{"List.Parse", func() (any, error) { return l.Parse("Allegro") }, term.Term{}},
		{"List.Parse empty", func() (any, error) { return l.Parse("") }, term.Term{}},
		{"List.OfGroup", func() (any, error) { return l.OfGroup(term.Fast) }, []term.Term(nil)},
		{"List.OfGroup invalid", func() (any, error) { return l.OfGroup(term.Group(9)) }, []term.Term(nil)},
		{"List.At", func() (any, error) { return l.At(fi(100)) }, []term.Term(nil)},
		{"List.At unset number", func() (any, error) { return l.At(fraction.Fraction{}) }, []term.Term(nil)},
		{"List.At zero", func() (any, error) { return l.At(fi(0)) }, []term.Term(nil)},
		{"Term.Name", func() (any, error) { return w.Name() }, ""},
		{"Term.List", func() (any, error) { return w.List() }, term.List{}},
		{"Term.Place", func() (any, error) { return w.Place() }, 0},
		{"Term.Group", func() (any, error) { return w.Group() }, term.NoGroup},
		{"Term.MarshalText", func() (any, error) { return w.MarshalText() }, []byte(nil)},
	}

	for _, tt := range unsetCalls {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.call()
			assertErr(t, err, term.ErrUnset)
			assert.Equal(t, tt.zero, got)
		})
	}

	t.Run("Term.Guide", func(t *testing.T) {
		t.Parallel()

		g, ok, err := w.Guide()
		assertErr(t, err, term.ErrUnset)
		assert.Equal(t, term.Guide{}, g)
		assert.False(t, ok)
	})

	t.Run("Term.English and Russian", func(t *testing.T) {
		t.Parallel()

		s, ok, err := w.English()
		assertErr(t, err, term.ErrUnset)
		assert.Empty(t, s)
		assert.False(t, ok)

		s, ok, err = w.Russian()
		assertErr(t, err, term.ErrUnset)
		assert.Empty(t, s)
		assert.False(t, ok)
	})

	t.Run("Term.AppendText", func(t *testing.T) {
		t.Parallel()

		dst := []byte("keep")
		got, err := w.AppendText(dst)
		assertErr(t, err, term.ErrUnset)
		assert.Equal(t, []byte("keep"), got, "AppendText returns dst unchanged")
	})
}
