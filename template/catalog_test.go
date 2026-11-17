package template_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntry(t *testing.T) {
	t.Parallel()
	t.Run("readings and languages", func(t *testing.T) {
		t.Parallel()
		entry := template.MustNewEntry("mixed", []string{"First", "Первый", "Second", "Второй"}, template.Ionian())
		assert.True(t, entry.IsValid())
		assert.Equal(t, "mixed", entry.Code())
		assert.True(t, entry.Template().Equal(template.Ionian()))
		assert.Equal(t, []string{"First", "Первый", "Second", "Второй"}, entry.Names())
		assert.Equal(t, []string{"Первый", "Второй"}, entry.RussianNames())
		assert.Equal(t, []string{"First", "Second"}, entry.EnglishNames())
		russian, ok := entry.RussianName()
		assert.True(t, ok)
		assert.Equal(t, "Первый", russian)
		english, ok := entry.EnglishName()
		assert.True(t, ok)
		assert.Equal(t, "First", english)
		assert.Equal(t, "mixed: "+template.Ionian().String(), entry.String())
	})

	t.Run("missing languages", func(t *testing.T) {
		t.Parallel()
		english := template.MustNewEntry("english", []string{"Only English"}, opened("m3 M2"))
		name, ok := english.RussianName()
		assert.False(t, ok)
		assert.Empty(t, name)
		assert.NotNil(t, english.RussianNames())
		assert.Empty(t, english.RussianNames())
		assert.Equal(t, []string{"Only English"}, english.EnglishNames())
		russian := template.MustNewEntry("russian", []string{"Только русское"}, template.Ionian())
		name, ok = russian.EnglishName()
		assert.False(t, ok)
		assert.Empty(t, name)
		assert.NotNil(t, russian.EnglishNames())
		assert.Empty(t, russian.EnglishNames())
	})

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		var entry template.Entry
		assert.False(t, entry.IsValid())
		assert.Empty(t, entry.Code())
		assert.False(t, entry.Template().IsValid())
		assert.Nil(t, entry.Names())
		assert.Nil(t, entry.RussianNames())
		assert.Nil(t, entry.EnglishNames())
		name, ok := entry.RussianName()
		assert.False(t, ok)
		assert.Empty(t, name)
		name, ok = entry.EnglishName()
		assert.False(t, ok)
		assert.Empty(t, name)
		assert.False(t, entry.Equal(entry))
		assert.Equal(t, "Entry(unset)", entry.String())
	})

	t.Run("equality uses original names and order", func(t *testing.T) {
		t.Parallel()
		base := template.MustNewEntry("code", []string{"Dorian", "Mode"}, template.Dorian())
		assert.True(t, base.Equal(template.MustNewEntry("code", []string{"Dorian", "Mode"}, template.Dorian())))
		assert.False(t, base.Equal(template.MustNewEntry("other", []string{"Dorian", "Mode"}, template.Dorian())))
		assert.False(t, base.Equal(template.MustNewEntry("code", []string{"dorian", "Mode"}, template.Dorian())))
		assert.False(t, base.Equal(template.MustNewEntry("code", []string{"Mode", "Dorian"}, template.Dorian())))
		assert.False(t, base.Equal(template.MustNewEntry("code", []string{"Dorian", "Mode"}, template.Aeolian())))
		assert.False(t, base.Equal(template.Entry{}))
	})
}

func TestEntryValidation(t *testing.T) {
	t.Parallel()
	valid := template.Ionian()
	cases := []struct {
		name  string
		code  string
		names []string
		tmpl  template.Template
		want  error
	}{
		{name: "unset template", code: "code", names: []string{"Name"}, tmpl: template.Template{}, want: template.ErrUnset},
		{name: "unset template before invalid code and names", names: nil, tmpl: template.Template{}, want: template.ErrUnset},
		{name: "empty code", names: []string{"Name"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "code with ASCII space", code: "my code", names: []string{"Name"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "code with Unicode space", code: "my\u2003code", names: []string{"Name"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "no names", code: "code", tmpl: valid, want: template.ErrInvalidArgument},
		{name: "empty name", code: "code", names: []string{""}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "spaces only name", code: "code", names: []string{" \t"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "leading name space", code: "code", names: []string{" Name"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "trailing name space", code: "code", names: []string{"Name\u2003"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "normalized duplicate case", code: "code", names: []string{"Dorian", "dorian"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "normalized duplicate sign", code: "code", names: []string{"Dorian ♭2", "dorian b2"}, tmpl: valid, want: template.ErrInvalidArgument},
		{name: "normalized duplicate whitespace", code: "code", names: []string{"My mode", "my  mode"}, tmpl: valid, want: template.ErrInvalidArgument},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.NewEntry(tc.code, tc.names, tc.tmpl)
			require.ErrorIs(t, err, tc.want)
			assert.False(t, got.IsValid())
		})
	}
}

func TestEntryOwnsNameSlices(t *testing.T) {
	t.Parallel()
	input := []string{"Первый", "First"}
	entry := template.MustNewEntry("entry", input, template.Ionian())
	input[0] = "Changed"
	assert.Equal(t, []string{"Первый", "First"}, entry.Names())
	for _, read := range []func() []string{entry.Names, entry.RussianNames, entry.EnglishNames} {
		got := read()
		got[0] = "Changed"
		assert.NotContains(t, entry.Names(), "Changed")
	}
}

func TestCatalogZeroAndEmpty(t *testing.T) {
	t.Parallel()
	empty, err := template.NewCatalog()
	require.NoError(t, err)
	var zero template.Catalog
	for _, catalog := range []template.Catalog{zero, empty} {
		assert.Zero(t, catalog.Len())
		assert.NotNil(t, catalog.All())
		assert.Empty(t, slices.Collect(catalog.All()))
		assert.NotNil(t, catalog.Entries())
		assert.Empty(t, catalog.Entries())
		assert.NotNil(t, catalog.Codes())
		assert.Empty(t, catalog.Codes())
		assert.NotNil(t, catalog.Templates())
		assert.Empty(t, catalog.Templates())
		entry, ok := catalog.ByCode("ionian")
		assert.False(t, ok)
		assert.False(t, entry.IsValid())
		assert.NotNil(t, catalog.ByName("anything"))
		assert.Empty(t, catalog.ByName("anything"))
		assert.NotNil(t, catalog.MustIdentify(template.Ionian()))
		assert.Empty(t, catalog.MustIdentify(template.Ionian()))
		assert.NotNil(t, catalog.MustIdentifyRotations(template.Ionian()))
		assert.Empty(t, catalog.MustIdentifyRotations(template.Ionian()))
		assert.Equal(t, "Catalog[]", catalog.String())
		assert.True(t, catalog.Equal(template.Catalog{}))
	}
	assert.True(t, zero.Equal(empty))
}

func catalogFixture(t testing.TB) (template.Catalog, template.Entry, template.Entry, template.Entry) {
	t.Helper()
	a := template.MustNewEntry("a", []string{"Alpha", "Shared"}, closed("M2 m2"))
	b := template.MustNewEntry("b", []string{"Beta", "Shared"}, closed("M2 M2"))
	open := template.MustNewEntry("open", []string{"Open"}, opened("m3 M2"))
	catalog, err := template.NewCatalog(a, b)
	require.NoError(t, err)
	return catalog, a, b, open
}

func TestCatalogReadAndSequence(t *testing.T) {
	t.Parallel()
	catalog, a, b, _ := catalogFixture(t)
	assert.Equal(t, 2, catalog.Len())
	assert.Equal(t, []string{"a", "b"}, catalog.Codes())
	assert.Equal(t, []template.Entry{a, b}, catalog.Entries())
	templates := catalog.Templates()
	require.Len(t, templates, 2)
	assert.True(t, templates[0].Equal(a.Template()))
	assert.True(t, templates[1].Equal(b.Template()))
	assert.Equal(t, "Catalog[a, b]", catalog.String())
	got, ok := catalog.ByCode("a")
	assert.True(t, ok)
	assert.True(t, got.Equal(a))
	for _, code := range []string{"", "A", "missing"} {
		got, ok = catalog.ByCode(code)
		assert.False(t, ok)
		assert.False(t, got.IsValid())
	}
	assert.Equal(t, []string{"a", "b"}, entryCodes(catalog.ByName("shared")))
	assert.Empty(t, catalog.ByName("a"), "codes are not names")

	first := slices.Collect(catalog.All())
	second := slices.Collect(catalog.All())
	assert.Equal(t, []string{"a", "b"}, entryCodes(first))
	assert.Equal(t, first, second)
	var stopped []string
	catalog.All()(func(entry template.Entry) bool {
		stopped = append(stopped, entry.Code())
		return false
	})
	assert.Equal(t, []string{"a"}, stopped)

	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			assert.Equal(t, []string{"a", "b"}, entryCodes(slices.Collect(catalog.All())))
			assert.Equal(t, []string{"a", "b"}, entryCodes(catalog.ByName("SHARED")))
		})
	}
	wg.Wait()
}

func TestCatalogChanges(t *testing.T) {
	t.Parallel()
	catalog, a, b, open := catalogFixture(t)
	c := template.MustNewEntry("c", []string{"Gamma"}, closed("P8"))
	d := template.MustNewEntry("d", []string{"Delta"}, closed("P5"))

	t.Run("with", func(t *testing.T) {
		t.Parallel()
		got, err := catalog.With(c, d)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c", "d"}, got.Codes())
		assert.Equal(t, []string{"a", "b"}, catalog.Codes())
		unchanged, err := catalog.With()
		require.NoError(t, err)
		assert.True(t, unchanged.Equal(catalog))
	})

	t.Run("without", func(t *testing.T) {
		t.Parallel()
		got, err := catalog.With(open)
		require.NoError(t, err)
		got, err = got.Without("a", "open")
		require.NoError(t, err)
		assert.Equal(t, []string{"b"}, got.Codes())
		assert.Equal(t, []string{"a", "b"}, catalog.Codes())
		unchanged, err := catalog.Without()
		require.NoError(t, err)
		assert.True(t, unchanged.Equal(catalog))
	})

	t.Run("merge", func(t *testing.T) {
		t.Parallel()
		other := template.MustNewCatalog(c, d)
		got, err := catalog.Merge(other)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c", "d"}, got.Codes())
		assert.Equal(t, []string{"c", "d"}, other.Codes())
		empty, err := catalog.Merge(template.Catalog{})
		require.NoError(t, err)
		assert.True(t, empty.Equal(catalog))
	})

	t.Run("filter", func(t *testing.T) {
		t.Parallel()
		var calls []string
		got, err := catalog.Filter(func(entry template.Entry) bool {
			calls = append(calls, entry.Code())
			return entry.Code() == "b"
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, calls)
		assert.Equal(t, []string{"b"}, got.Codes())
		got, err = catalog.Filter(func(template.Entry) bool { return false })
		require.NoError(t, err)
		assert.Zero(t, got.Len())
		assert.Equal(t, []string{"a", "b"}, catalog.Codes())
	})

	t.Run("equality and order", func(t *testing.T) {
		t.Parallel()
		assert.True(t, catalog.Equal(template.MustNewCatalog(a, b)))
		assert.False(t, catalog.Equal(template.MustNewCatalog(b, a)))
		assert.False(t, catalog.Equal(template.MustNewCatalog(a)))
	})
}

func TestCatalogErrorsLeaveInputsUnchanged(t *testing.T) {
	t.Parallel()
	catalog, a, b, _ := catalogFixture(t)
	other := template.MustNewCatalog(template.MustNewEntry("c", []string{"Gamma"}, template.Ionian()))
	cases := []struct {
		name string
		want error
		call func() (template.Catalog, error)
	}{
		{name: "new unset entry", want: template.ErrUnset, call: func() (template.Catalog, error) { return template.NewCatalog(a, template.Entry{}) }},
		{name: "new duplicate", want: template.ErrDuplicateCode, call: func() (template.Catalog, error) { return template.NewCatalog(a, a) }},
		{name: "with unset before duplicate", want: template.ErrUnset, call: func() (template.Catalog, error) { return catalog.With(template.Entry{}, a) }},
		{name: "with existing duplicate", want: template.ErrDuplicateCode, call: func() (template.Catalog, error) { return catalog.With(a) }},
		{name: "with duplicate among additions", want: template.ErrDuplicateCode, call: func() (template.Catalog, error) { return catalog.With(other.Entries()[0], other.Entries()[0]) }},
		{name: "without unknown", want: template.ErrUnknownCode, call: func() (template.Catalog, error) { return catalog.Without("a", "missing") }},
		{name: "without repeated code", want: template.ErrUnknownCode, call: func() (template.Catalog, error) { return catalog.Without("a", "a") }},
		{name: "merge duplicate", want: template.ErrDuplicateCode, call: func() (template.Catalog, error) { return catalog.Merge(template.MustNewCatalog(b)) }},
		{name: "nil filter", want: template.ErrUnset, call: func() (template.Catalog, error) { return catalog.Filter(nil) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.call()
			require.ErrorIs(t, err, tc.want)
			assert.Zero(t, got.Len())
			assert.Equal(t, []string{"a", "b"}, catalog.Codes())
			assert.Equal(t, []string{"c"}, other.Codes())
		})
	}

	t.Run("predicate panic is not intercepted", func(t *testing.T) {
		t.Parallel()
		marker := errorsNewForTest("predicate")
		require.PanicsWithValue(t, marker, func() {
			_, _ = catalog.Filter(func(template.Entry) bool { panic(marker) })
		})
	})
}

func errorsNewForTest(text string) error { return &testError{text: text} }

type testError struct{ text string }

func (e *testError) Error() string { return e.text }

func TestCatalogOwnsSlices(t *testing.T) {
	t.Parallel()
	catalog, a, b, _ := catalogFixture(t)
	input := []template.Entry{a, b}
	fromInput := template.MustNewCatalog(input...)
	input[0] = template.Entry{}
	assert.Equal(t, []string{"a", "b"}, fromInput.Codes())
	entries := catalog.Entries()
	entries[0] = template.Entry{}
	assert.Equal(t, []string{"a", "b"}, catalog.Codes())
	codes := catalog.Codes()
	codes[0] = "changed"
	assert.Equal(t, []string{"a", "b"}, catalog.Codes())
	templates := catalog.Templates()
	templates[0] = template.Template{}
	assert.True(t, catalog.Templates()[0].Equal(closed("M2 m2")))
	names := catalog.Entries()[0].Names()
	names[0] = "Changed"
	assert.Equal(t, "Alpha", catalog.Entries()[0].Names()[0])
}
