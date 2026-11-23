package naming_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/interval/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntries covers the vocabulary enumeration: the generic sizes 1..15
// first, then every named specific interval in ascending size, with the full
// name always present and the short form exactly when the language gives one.
func TestEntries(t *testing.T) {
	t.Parallel()

	for _, lang := range []struct {
		name       string
		cat        naming.Catalog
		specifics  int // sizes 1..15: 4 at the unison, 7 at each of the 6 other perfect sizes (4, 5, 8, 11, 12, 15), 8 at each of the 8 major-minor sizes (2, 3, 6, 7, 9, 10, 13, 14), plus the named neutrals
		neutralCnt int
	}{
		{"russian", naming.Russian(), 4 + 6*7 + 8*8 + 1, 1},
		{"english", naming.English(), 4 + 6*7 + 8*8 + 5, 5},
	} {
		t.Run(lang.name, func(t *testing.T) {
			t.Parallel()

			entries, err := lang.cat.Entries()
			require.NoError(t, err)
			require.Len(t, entries, 15+lang.specifics)
			assert.Greater(t, len(entries), 57, "the vocabulary is wider than the predefined constructors")

			// The first 15 entries are the generic sizes in ascending order.
			for n := 1; n <= 15; n++ {
				e := entries[n-1]
				assert.True(t, e.IsGeneric(), "entry %d", n-1)
				assert.Equal(t, n, e.Generic().Size())
				assert.False(t, e.Interval().IsValid())
				full, ok := e.MustNameFull()
				require.True(t, ok)
				g, err := lang.cat.ParseGeneric(full)
				require.NoError(t, err)
				assert.Equal(t, n, g.Size())
			}

			// The rest are the named specific intervals in ascending size.
			var specifics, neutrals int
			prevSize := 0
			for _, e := range entries[15:] {
				assert.False(t, e.IsGeneric())
				assert.False(t, e.Generic().IsValid())
				i := e.Interval()
				require.True(t, i.IsValid())
				assert.Equal(t, interval.Up, i.Direction(), "the subjects are ascending")
				assert.GreaterOrEqual(t, i.Generic().Size(), prevSize)
				prevSize = i.Generic().Size()

				full, ok := e.MustNameFull()
				require.True(t, ok, "every entry has a full name")
				back, err := lang.cat.ParseInterval(full)
				require.NoError(t, err, "full name %q", full)
				assert.True(t, back.Equal(i), "full name %q round-trips", full)

				if short, ok := e.MustName(); ok {
					back, err := lang.cat.ParseInterval(short)
					require.NoError(t, err, "short form %q", short)
					assert.True(t, back.Equal(i), "short form %q round-trips", short)
				}

				specifics++
				if i.Deviation().Equal(alteration.MustFromFraction(-1, 2)) {
					neutrals++
				}
			}
			assert.Equal(t, lang.specifics, specifics)
			assert.Equal(t, lang.neutralCnt, neutrals)

			// Repeated calls give the same order and subjects.
			again := lang.cat.MustEntries()
			require.Len(t, again, len(entries))
			for k := range entries {
				assert.Equal(t, entries[k].IsGeneric(), again[k].IsGeneric())
				assert.True(t, entries[k].Generic() == again[k].Generic())
				assert.True(t, entries[k].Interval() == again[k].Interval())
			}
		})
	}

	t.Run("russian major third", func(t *testing.T) {
		t.Parallel()
		e := findSpecific(t, naming.Russian(), interval.MustNew(interval.Major, 3))
		full, ok := e.MustNameFull()
		require.True(t, ok)
		assert.Equal(t, "большая терция", full)
		short, ok := e.MustName()
		require.True(t, ok)
		assert.Equal(t, "б.3", short)
	})

	t.Run("english major third", func(t *testing.T) {
		t.Parallel()
		e := findSpecific(t, naming.English(), interval.MustNew(interval.Major, 3))
		full, ok := e.MustNameFull()
		require.True(t, ok)
		assert.Equal(t, "major third", full)
		short, ok := e.MustName()
		require.True(t, ok)
		assert.Equal(t, "M3", short)
	})

	t.Run("russian neutral third has no short form", func(t *testing.T) {
		t.Parallel()
		e := findSpecific(t, naming.Russian(), interval.MustNew(interval.Neutral, 3))
		full, ok := e.MustNameFull()
		require.True(t, ok)
		assert.Equal(t, "нейтральная терция", full)
		short, ok := e.MustName()
		assert.False(t, ok)
		assert.Empty(t, short)
	})

	t.Run("english neutral second has both forms", func(t *testing.T) {
		t.Parallel()
		e := findSpecific(t, naming.English(), interval.MustNew(interval.Neutral, 2))
		full, ok := e.MustNameFull()
		require.True(t, ok)
		assert.Equal(t, "neutral second", full)
		short, ok := e.MustName()
		require.True(t, ok)
		assert.Equal(t, "n2", short)
	})

	t.Run("multiplicities up to three are discoverable", func(t *testing.T) {
		t.Parallel()
		e := findSpecific(t, naming.English(), interval.MustNewMultiple(3, interval.Augmented, 5))
		full, ok := e.MustNameFull()
		require.True(t, ok)
		assert.Equal(t, "triply augmented fifth", full)
	})

	t.Run("zero catalog", func(t *testing.T) {
		t.Parallel()
		var c naming.Catalog
		entries, err := c.Entries()
		assert.Nil(t, entries)
		assert.ErrorIs(t, err, interval.ErrUnset)
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			c.MustEntries()
		}()
		panicErr, ok := recovered.(error)
		require.True(t, ok, "the panic value is an error")
		assert.ErrorIs(t, panicErr, interval.ErrUnset)
	})
}

// findSpecific returns the entry of the given ascending interval.
func findSpecific(t *testing.T, c naming.Catalog, want interval.Interval) naming.Entry {
	t.Helper()
	for _, e := range c.MustEntries() {
		if !e.IsGeneric() && e.Interval().Equal(want) {
			return e
		}
	}
	t.Fatalf("no entry for %v", want)
	return naming.Entry{}
}

// The Russian major third is discoverable with its full name and its short
// form.
func ExampleCatalog_Entries() {
	for _, e := range naming.Russian().MustEntries() {
		if !e.IsGeneric() && e.Interval().Equal(interval.MustNew(interval.Major, 3)) {
			full, _ := e.MustNameFull()
			short, ok := e.MustName()
			fmt.Println(full, short, ok)
		}
	}
	// Output:
	// большая терция б.3 true
}
