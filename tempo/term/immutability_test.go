package term_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewListDoesNotRetainInput covers architecture 3.2: rows, their Entries
// slices and Guide pointers are only read during NewList; changing them later
// does not change the list, and NewList does not change them.
func TestNewListDoesNotRetainInput(t *testing.T) {
	t.Parallel()

	g := between(40, 60)
	entries := []term.Entry{{Name: "a", English: "e"}, {Name: "b"}}
	rows := []term.Row{{Entries: entries, Group: term.Slow, Guide: &g}, {Entries: []term.Entry{{Name: "c"}}}}

	l := term.MustNewList(rows...)
	want := term.MustNewList(customRowsCopy(rows)...)

	assert.Equal(t, between(40, 60), g, "the guide is not changed")
	assert.Equal(t, []term.Entry{{Name: "a", English: "e"}, {Name: "b"}}, entries, "the entries are not changed")

	g = between(1, 2)
	entries[0] = term.Entry{Name: "z", English: "changed"}
	rows[0].Group = term.Fast
	rows[1].Entries = nil

	assert.Equal(t, want, l)
	assert.Equal(t, "List[a, b; c]", l.String())

	a := l.MustParse("a")
	gg, ok := a.MustGuide()
	assert.True(t, ok)
	assert.Equal(t, between(40, 60), gg)
	assert.Equal(t, term.Slow, a.MustGroup())

	en, _ := a.MustEnglish()
	assert.Equal(t, "e", en)
}

// customRowsCopy returns a deep copy of rows.
func customRowsCopy(rows []term.Row) []term.Row {
	out := make([]term.Row, len(rows))
	for i, r := range rows {
		out[i] = term.Row{Entries: append([]term.Entry(nil), r.Entries...), Group: r.Group}
		if r.Guide != nil {
			out[i].Guide = guidePtr(*r.Guide)
		}
	}

	return out
}

// TestResultSlicesBelongToCaller covers architecture 3.2: Place, Terms,
// OfGroup, At, BuiltIn and Groups return new slices; changing them does not
// change later results.
func TestResultSlicesBelongToCaller(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	other := term.RussianWikipedia().MustParse("grave")

	calls := map[string]func() []term.Term{
		"Place":   func() []term.Term { return en.MustPlace(2) },
		"Terms":   func() []term.Term { return en.MustTerms() },
		"OfGroup": func() []term.Term { return en.MustOfGroup(term.Fast) },
		"At":      func() []term.Term { return en.MustAt(fi(120)) },
	}

	for name, call := range calls {
		first := call()
		require.NotEmpty(t, first, name)

		want := append([]term.Term(nil), first...)

		for i := range first {
			first[i] = other
		}

		assert.Equal(t, want, call(), name)
	}
}

// TestValuesAreNotChanged covers С-58 for tempo/term: operations, including
// failing ones and Must panics, leave their receivers and arguments equal to
// their earlier values.
func TestValuesAreNotChanged(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	w := en.MustParse("Largo")
	g := between(40, 60)
	before := []any{en, w, g}

	_, _ = en.Parse("nothing")
	_, _ = en.At(fi(-1))
	_, _ = en.Place(0)
	_, _ = w.ComparePlace(term.RussianWikipedia().MustParse("largo"))
	_, _ = g.Contains(fi(0))
	_, _ = capturePanic(func() { en.MustParse("nothing") })
	_, _ = capturePanic(func() { g.MustContains(fi(-1)) })
	_, _ = w.AppendText(make([]byte, 0, 64))

	assert.Equal(t, before, []any{en, w, g})
}

// TestAppendTextKeepsPrefix covers architecture 3.2: AppendText(b) does not
// change b[:len(b)], even when b has spare capacity.
func TestAppendTextKeepsPrefix(t *testing.T) {
	t.Parallel()

	buf := make([]byte, 3, 64)
	copy(buf, "abc")

	got, err := term.EnglishWikipedia().MustParse("Allegro").AppendText(buf)
	require.NoError(t, err)
	assert.Equal(t, "abcAllegro", string(got))
	assert.Equal(t, "abc", string(buf))

	text, err := term.EnglishWikipedia().MustParse("Allegro").MarshalText()
	require.NoError(t, err)

	text[0] = 'X'
	assert.Equal(t, "Allegro", term.EnglishWikipedia().MustParse("Allegro").String(), "MarshalText returns a new slice")
}

// TestConcurrentReads covers architecture 3.2, "Конкурентность": built-in
// lists, a user list, and their terms are read from many goroutines at once
// without synchronization; run with -race.
func TestConcurrentReads(t *testing.T) {
	t.Parallel()

	shared := customList()
	allegro := term.EnglishWikipedia().MustParse("Allegro")

	const workers = 16

	type outcome struct {
		at, parse, place, str string
		err                   error
	}

	results := make(chan outcome, workers)

	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			var o outcome

			for _, l := range term.BuiltIn() {
				if _, err := l.Terms(); err != nil {
					o.err = err
				}
			}

			at, err := term.EnglishWikipedia().At(fi(130))
			if err != nil {
				o.err = err
			}

			for _, w := range at {
				o.at += w.String() + ";"
			}

			o.parse = shared.MustParse("живо").String()
			o.place = allegro.String()
			o.str = term.Default().String()
			results <- o
		})
	}

	wg.Wait()
	close(results)

	for o := range results {
		require.NoError(t, o.err)
		assert.Equal(t, "Allegro;Molto Allegro;Allegro vivace;", o.at)
		assert.Equal(t, "Живо", o.parse)
		assert.Equal(t, "Allegro", o.place)
		assert.Equal(t, stringOf(list1), o.str)
	}
}
