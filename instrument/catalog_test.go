package instrument_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/instrument"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lead, alto and bass are three presets with different names for catalogs.
func lead() *instrument.Preset { return preset("Lead", h0(), rng("C4", "G5")) }
func alto() *instrument.Preset { return preset("Alto part", h415(), rng("G3", "D5")) }
func bass() *instrument.Preset { return preset("Bass part", h0(), instrument.Range{}) }

// names collects the names of a catalog.
func names(c *instrument.Catalog) []string {
	return slices.Collect(c.Names())
}

// assertCatalog checks the order, the names and the presets of a catalog.
func assertCatalog(t *testing.T, want []*instrument.Preset, c *instrument.Catalog) {
	t.Helper()
	require.Equal(t, len(want), c.Len())
	got := slices.Collect(c.All())
	require.Len(t, got, len(want))
	wantNames := make([]string, len(want))
	for i, w := range want {
		wantNames[i] = w.Name()
		assert.True(t, w.Equal(got[i]), "preset %d: want %v, got %v", i, w, got[i])
	}
	// Collecting an empty sequence gives a nil slice; only the names and
	// their order are compared.
	assert.Equal(t, wantNames, append(make([]string, 0, len(want)), names(c)...))
}

// ОП-25, ОП-27, П-8, К-11, К-23, А 4.6.
func TestNewCatalog(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		presets []*instrument.Preset
		wantErr error
	}{
		{name: "order is kept", presets: []*instrument.Preset{lead(), alto(), bass()}},
		{name: "another order", presets: []*instrument.Preset{bass(), lead(), alto()}},
		{name: "Г-11 empty", presets: nil},
		{name: "one preset", presets: []*instrument.Preset{alto()}},
		{name: "names differ in case", presets: []*instrument.Preset{lead(), lead().MustWithName("lead")}},
		{name: "names differ in a space", presets: []*instrument.Preset{lead(), lead().MustWithName("Lead ")}},
		{name: "nil preset", presets: []*instrument.Preset{lead(), nil}, wantErr: instrument.ErrUnset},
		{name: "zero preset", presets: []*instrument.Preset{{}}, wantErr: instrument.ErrUnset},
		{name: "duplicate name", presets: []*instrument.Preset{lead(), alto(), lead().MustWithRange(rng("C3", "C4"))}, wantErr: instrument.ErrDuplicateName},
		{name: "К-23 unset before duplicates", presets: []*instrument.Preset{nil, lead(), lead()}, wantErr: instrument.ErrUnset},
		{name: "К-23 duplicates before a later unset", presets: []*instrument.Preset{lead(), lead(), nil}, wantErr: instrument.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := instrument.NewCatalog(tt.presets...)
			if tt.wantErr != nil {
				assertCause(t, err, tt.wantErr)
				assert.Nil(t, c)
				assertPanicsWith(t, tt.wantErr, err, func() { instrument.MustNewCatalog(tt.presets...) })
				return
			}
			require.NoError(t, err)
			require.NotNil(t, c)
			assertCatalog(t, tt.presets, c)
			assertCatalog(t, tt.presets, instrument.MustNewCatalog(tt.presets...))
		})
	}
}

// П-18а, А 4.6: nil and &Catalog{} are the empty catalog.
func TestCatalog_empty(t *testing.T) {
	t.Parallel()
	empty := map[string]*instrument.Catalog{
		"nil":          nil,
		"zero":         {},
		"NewCatalog()": instrument.MustNewCatalog(),
	}
	for name, c := range empty {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertNoPanic(t, func() {
				assert.Equal(t, 0, c.Len())
				assert.Empty(t, slices.Collect(c.All()))
				assert.Empty(t, names(c))
				ps, ok := c.Lookup("Lead")
				assert.False(t, ok)
				assert.Nil(t, ps)
				assert.Equal(t, "Catalog(empty)", c.String())
				assert.True(t, c.Equal(nil))
				assert.True(t, c.Equal(&instrument.Catalog{}))
				assert.True(t, c.Equal(instrument.MustNewCatalog()))
				assert.False(t, c.Equal(instrument.MustNewCatalog(lead())))
			})

			ps, err := c.Find("Lead")
			assertCause(t, err, instrument.ErrNotFound)
			assert.Nil(t, ps)
			assertPanicsWith(t, instrument.ErrNotFound, err, func() { c.MustFind("Lead") })

			one, err := c.With(lead())
			require.NoError(t, err)
			assertCatalog(t, []*instrument.Preset{lead()}, one)

			without, err := c.Without("Lead")
			assertCause(t, err, instrument.ErrNotFound)
			assert.Nil(t, without)
			assertPanicsWith(t, instrument.ErrNotFound, err, func() { c.MustWithout("Lead") })
		})
	}
}

// ОП-26, П-8, К-11: the search is exact.
func TestCatalog_LookupAndFind(t *testing.T) {
	t.Parallel()
	c := instrument.MustNewCatalog(lead(), alto(), bass())
	tests := []struct {
		name  string
		query string
		want  *instrument.Preset
	}{
		{name: "first", query: "Lead", want: lead()},
		{name: "middle", query: "Alto part", want: alto()},
		{name: "last", query: "Bass part", want: bass()},
		{name: "other case", query: "lead"},
		{name: "leading space", query: " Lead"},
		{name: "trailing space", query: "Lead "},
		{name: "prefix", query: "Alto"},
		{name: "empty", query: ""},
		{name: "unknown", query: "Violin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ps, ok := c.Lookup(tt.query)
			found, err := c.Find(tt.query)
			if tt.want == nil {
				assert.False(t, ok)
				assert.Nil(t, ps)
				assertCause(t, err, instrument.ErrNotFound)
				assert.Nil(t, found)
				assertPanicsWith(t, instrument.ErrNotFound, err, func() { c.MustFind(tt.query) })
				return
			}
			require.True(t, ok)
			assert.True(t, tt.want.Equal(ps))
			require.NoError(t, err)
			assert.True(t, tt.want.Equal(found))
			assert.True(t, tt.want.Equal(c.MustFind(tt.query)))
		})
	}
}

// ОП-28, П-8, К-11: changes give a new catalog; the source does not change.
func TestCatalog_WithAndWithout(t *testing.T) {
	t.Parallel()
	t.Run("with appends at the end", func(t *testing.T) {
		t.Parallel()
		src := instrument.MustNewCatalog(lead(), alto())
		got, err := src.With(bass())
		require.NoError(t, err)
		assertCatalog(t, []*instrument.Preset{lead(), alto(), bass()}, got)
		assertCatalog(t, []*instrument.Preset{lead(), alto()}, src)
		assertCatalog(t, []*instrument.Preset{lead(), alto(), bass()}, src.MustWith(bass()))
		assert.NotSame(t, src, got)
	})

	t.Run("with a duplicate name", func(t *testing.T) {
		t.Parallel()
		src := instrument.MustNewCatalog(lead(), alto())
		dup := alto().MustWithSettingsSet(h0())
		got, err := src.With(dup)
		assertCause(t, err, instrument.ErrDuplicateName)
		assert.Nil(t, got)
		assertPanicsWith(t, instrument.ErrDuplicateName, err, func() { src.MustWith(dup) })
		assertCatalog(t, []*instrument.Preset{lead(), alto()}, src)
	})

	t.Run("with an unset preset", func(t *testing.T) {
		t.Parallel()
		src := instrument.MustNewCatalog(lead())
		for _, ps := range []*instrument.Preset{nil, {}} {
			got, err := src.With(ps)
			assertCause(t, err, instrument.ErrUnset)
			assert.Nil(t, got)
			assertPanicsWith(t, instrument.ErrUnset, err, func() { src.MustWith(ps) })
		}
		assertCatalog(t, []*instrument.Preset{lead()}, src)
	})

	t.Run("without keeps the order of the others", func(t *testing.T) {
		t.Parallel()
		src := instrument.MustNewCatalog(lead(), alto(), bass())
		got, err := src.Without("Alto part")
		require.NoError(t, err)
		assertCatalog(t, []*instrument.Preset{lead(), bass()}, got)
		assertCatalog(t, []*instrument.Preset{lead(), alto(), bass()}, src)
		assertCatalog(t, []*instrument.Preset{alto(), bass()}, src.MustWithout("Lead"))
		assertCatalog(t, []*instrument.Preset{}, instrument.MustNewCatalog(bass()).MustWithout("Bass part"))
	})

	t.Run("without a missing name", func(t *testing.T) {
		t.Parallel()
		src := instrument.MustNewCatalog(lead(), alto())
		for _, name := range []string{"Bass part", "lead", "", "Lead "} {
			got, err := src.Without(name)
			assertCause(t, err, instrument.ErrNotFound)
			assert.Nil(t, got)
			assertPanicsWith(t, instrument.ErrNotFound, err, func() { src.MustWithout(name) })
		}
		assertCatalog(t, []*instrument.Preset{lead(), alto()}, src)
	})

	t.Run("the built-in catalog does not change", func(t *testing.T) {
		t.Parallel()
		catalog := instrument.Builtins()
		n := catalog.Len()
		_, err := catalog.With(lead())
		require.NoError(t, err)
		_, err = catalog.Without(catalog.MustFind("Violin").Name())
		require.NoError(t, err)
		assert.Equal(t, n, catalog.Len())
		assert.True(t, catalog.Equal(instrument.Builtins()))
	})
}

// А 4.6: the catalog keeps and gives its own copies, so assigning through a
// passed or a received pointer does not change it.
func TestCatalog_copies(t *testing.T) {
	t.Parallel()
	t.Run("passed pointer", func(t *testing.T) {
		t.Parallel()
		passed := lead()
		c := instrument.MustNewCatalog(passed, alto())
		*passed = *bass()
		assertCatalog(t, []*instrument.Preset{lead(), alto()}, c)

		added := bass()
		c2 := c.MustWith(added)
		*added = *lead().MustWithName("Changed")
		assertCatalog(t, []*instrument.Preset{lead(), alto(), bass()}, c2)
	})

	t.Run("received pointer", func(t *testing.T) {
		t.Parallel()
		c := instrument.MustNewCatalog(lead(), alto())
		got, ok := c.Lookup("Lead")
		require.True(t, ok)
		*got = *bass()
		found := c.MustFind("Alto part")
		*found = *bass()
		for ps := range c.All() {
			*ps = *bass().MustWithName(ps.Name() + " changed")
		}
		assertCatalog(t, []*instrument.Preset{lead(), alto()}, c)
	})

	t.Run("every lookup is a new copy", func(t *testing.T) {
		t.Parallel()
		c := instrument.MustNewCatalog(lead())
		a, _ := c.Lookup("Lead")
		b, _ := c.Lookup("Lead")
		assert.NotSame(t, a, b)
		assert.True(t, a.Equal(b))
	})
}

// ОП-27: the sequences can be stopped and iterated again.
func TestCatalog_sequences(t *testing.T) {
	t.Parallel()
	c := instrument.MustNewCatalog(lead(), alto(), bass())
	var first []string
	for ps := range c.All() {
		first = append(first, ps.Name())
		if len(first) == 2 {
			break
		}
	}
	assert.Equal(t, []string{"Lead", "Alto part"}, first)

	var firstName string
	for name := range c.Names() {
		firstName = name
		break
	}
	assert.Equal(t, "Lead", firstName)
	assert.Equal(t, names(c), names(c))
}

// ОП-29, П-20, К-11.
func TestCatalog_Equal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b *instrument.Catalog
		want bool
	}{
		{name: "same presets in the same order", a: instrument.MustNewCatalog(lead(), alto()), b: instrument.MustNewCatalog(lead(), alto()), want: true},
		{name: "same presets in another order", a: instrument.MustNewCatalog(lead(), alto()), b: instrument.MustNewCatalog(alto(), lead()), want: false},
		{name: "one preset more", a: instrument.MustNewCatalog(lead(), alto()), b: instrument.MustNewCatalog(lead(), alto(), bass()), want: false},
		{name: "a preset differs", a: instrument.MustNewCatalog(lead(), alto()), b: instrument.MustNewCatalog(lead(), alto().MustWithSettingsSet(h0())), want: false},
		{name: "nil and empty", a: nil, b: instrument.MustNewCatalog(), want: true},
		{name: "nil and nil", a: nil, b: nil, want: true},
		{name: "nil and non-empty", a: nil, b: instrument.MustNewCatalog(lead()), want: false},
		{name: "built-ins", a: instrument.Builtins(), b: instrument.Builtins(), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.a.Equal(tt.b))
			assert.Equal(t, tt.want, tt.b.Equal(tt.a), "symmetric")
			assert.True(t, tt.a.Equal(tt.a), "reflexive")
		})
	}
}

// ОП-29, К-26, А 4.10: the string names all presets in order.
func TestCatalog_String(t *testing.T) {
	t.Parallel()
	c := instrument.MustNewCatalog(lead(), alto(), bass())
	s := c.String()
	last := -1
	for _, name := range []string{"Lead", "Alto part", "Bass part"} {
		at := strings.Index(s, name)
		require.GreaterOrEqual(t, at, 0, "%q contains %q", s, name)
		assert.Greater(t, at, last, "names in order in %q", s)
		last = at
	}
	assert.NotEqual(t, "Catalog(empty)", s)
}
