package instrument_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// builtin is a row of business-spec table 8.2: the function named by the rule
// of А 4.7, the exact name and the bounds as the standard settings write them.
type builtin struct {
	fn           func() *instrument.Preset
	name         string
	lower, upper string
}

// builtins is business-spec table 8.2 in its order.
var builtins = []builtin{
	{instrument.Violin, "Violin", "G3", "E7"},
	{instrument.Cello, "Cello", "C2", "G6"},
	{instrument.DoubleBass, "Double Bass", "E1", "D5"},
	{instrument.ClassicalGuitar, "Classical Guitar", "E2", "B5"},
	{instrument.BassGuitar, "Bass Guitar", "E1", "G4"},
	{instrument.Mandolin, "Mandolin", "G3", "F♯6"},
	{instrument.Flute, "Flute", "B3", "C7"},
	{instrument.Oboe, "Oboe", "B♭3", "G6"},
	{instrument.EnglishHorn, "English Horn", "E3", "C6"},
	{instrument.ClarinetInBFlat, "Clarinet in B-flat", "D3", "B♭6"},
	{instrument.ClarinetInA, "Clarinet in A", "C♯3", "A6"},
	{instrument.BassClarinet, "Bass Clarinet", "B♭1", "D6"},
	{instrument.Bassoon, "Bassoon", "B♭1", "E5"},
	{instrument.Contrabassoon, "Contrabassoon", "B♭0", "D4"},
	{instrument.SopranoSaxophone, "Soprano Saxophone", "A♭3", "E6"},
	{instrument.AltoSaxophone, "Alto Saxophone", "D♭3", "A♭5"},
	{instrument.TenorSaxophone, "Tenor Saxophone", "A♭2", "E5"},
	{instrument.Horn, "Horn", "B♭1", "F5"},
	{instrument.TrumpetInBFlat, "Trumpet in B-flat", "E3", "B♭5"},
	{instrument.Trombone, "Trombone", "E2", "F5"},
	{instrument.BassTrombone, "Bass Trombone", "B♭0", "B♭4"},
	{instrument.BassTuba, "Bass Tuba", "A1", "A♭4"},
	{instrument.ContrabassTuba, "Contrabass Tuba", "E1", "F4"},
	{instrument.Piano, "Piano", "A0", "C8"},
	{instrument.Piano85Keys, "Piano (85 keys)", "A0", "A7"},
	{instrument.Piano92Keys, "Piano (92 keys)", "F0", "C8"},
	{instrument.Organ, "Organ", "C0", "C9"},
	{instrument.Harpsichord, "Harpsichord", "F1", "F6"},
	{instrument.Celesta, "Celesta", "C4", "C8"},
	{instrument.Timpani, "Timpani", "D2", "A3"},
	{instrument.Xylophone, "Xylophone", "F4", "C8"},
	{instrument.Marimba, "Marimba", "C2", "C7"},
	{instrument.Vibraphone, "Vibraphone", "F3", "F6"},
	{instrument.SopranoVoice, "Soprano Voice", "C4", "G5"},
	{instrument.MezzoSopranoVoice, "Mezzo-Soprano Voice", "A3", "A5"},
	{instrument.AltoVoice, "Alto Voice", "G3", "D5"},
	{instrument.TenorVoice, "Tenor Voice", "C3", "G4"},
	{instrument.BaritoneVoice, "Baritone Voice", "G2", "A♭4"},
	{instrument.BassVoice, "Bass Voice", "F2", "D4"},
}

// ОП-23, П-21, К-9: every built-in preset is the row of table 8.2 with the
// standard settings; every call gives an equal preset and a new pointer.
func TestBuiltinPresets(t *testing.T) {
	t.Parallel()
	require.Len(t, builtins, 39)
	for _, b := range builtins {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()
			ps := b.fn()
			require.NotNil(t, ps)
			assert.True(t, ps.IsValid())
			assert.Equal(t, b.name, ps.Name())
			assert.NotEmpty(t, ps.Name())
			assert.True(t, ps.SettingsSet().Equal(settings.Standard()), "the standard settings")
			require.True(t, ps.HasRange())
			lower, upper := p(b.lower), p(b.upper)
			r := ps.Range()
			assert.True(t, lower.SameSpelling(r.Lower()) && lower.SameOctave(r.Lower()), "lower %v, want %v", r.Lower(), lower)
			assert.True(t, upper.SameSpelling(r.Upper()) && upper.SameOctave(r.Upper()), "upper %v, want %v", r.Upper(), upper)
			assert.True(t, instrument.MustNewRange(lower, upper).Equal(r))
			assert.NoError(t, r.Check(trad()), "valid in traditional")
			assert.Equal(t, -1, trad().MustCompare(r.Lower(), r.Upper()), "the lower bound is below the upper one")

			again := b.fn()
			assert.True(t, ps.Equal(again), "two calls give equal presets")
			assert.NotSame(t, ps, again, "every call gives a new pointer (КР-12)")
			assert.NotSame(t, ps.SettingsSet(), again.SettingsSet(), "every call gives a new standard set (КР-12)")
			for _, ch := range ps.Name() {
				assert.Less(t, ch, rune(128), "the name is ASCII")
			}
		})
	}

	t.Run("assigning through a result changes nothing else", func(t *testing.T) {
		t.Parallel()
		v := instrument.Violin()
		*v = *instrument.Cello()
		assert.Equal(t, "Violin", instrument.Violin().Name())
		assert.True(t, instrument.Violin().Range().Equal(rng("G3", "E7")))
		found, ok := instrument.Builtins().Lookup("Violin")
		require.True(t, ok)
		assert.Equal(t, "Violin", found.Name())
	})

	t.Run("names are unique", func(t *testing.T) {
		t.Parallel()
		seen := map[string]bool{}
		for _, b := range builtins {
			assert.False(t, seen[b.fn().Name()], "duplicate %q", b.name)
			seen[b.fn().Name()] = true
		}
	})
}

// ОП-24, К-10: the built-in catalog holds the 39 presets in the order of
// table 8.2 and nothing else.
func TestBuiltins(t *testing.T) {
	t.Parallel()
	c := instrument.Builtins()
	require.NotNil(t, c)
	require.Equal(t, 39, c.Len())

	wantNames := make([]string, len(builtins))
	for i, b := range builtins {
		wantNames[i] = b.name
	}
	assert.Equal(t, wantNames, slices.Collect(c.Names()), "the order of table 8.2")

	all := slices.Collect(c.All())
	require.Len(t, all, len(builtins))
	for i, b := range builtins {
		assert.True(t, b.fn().Equal(all[i]), "preset %d %q", i, b.name)
	}

	again := instrument.Builtins()
	assert.NotSame(t, c, again, "every call builds a new catalog")
	assert.True(t, c.Equal(again))

	for _, b := range builtins {
		t.Run("lookup "+b.name, func(t *testing.T) {
			t.Parallel()
			found, ok := instrument.Builtins().Lookup(b.name)
			require.True(t, ok)
			assert.True(t, b.fn().Equal(found))
			assert.True(t, b.fn().Equal(instrument.Builtins().MustFind(b.name)))
		})
	}

	misses := []string{
		"violin", "VIOLIN", " Violin", "Violin ", "", "Viola", "Harp", "Ukulele", "Piccolo",
		"Baritone Saxophone", "Accordion", "Banjo", "Glockenspiel", "Tuba", "Clarinet in Bb",
		"Bass", "Alto", "Tenor", "Soprano", "Baritone", "Mezzo-Soprano", "Piano (88 keys)",
	}
	for _, name := range misses {
		t.Run(fmt.Sprintf("miss %q", name), func(t *testing.T) {
			t.Parallel()
			found, ok := instrument.Builtins().Lookup(name)
			assert.False(t, ok)
			assert.Nil(t, found)
			_, err := instrument.Builtins().Find(name)
			assertCause(t, err, instrument.ErrNotFound)
			assertPanicsWith(t, instrument.ErrNotFound, err, func() { instrument.Builtins().MustFind(name) })
		})
	}

	t.Run("К-10 six different basses", func(t *testing.T) {
		t.Parallel()
		basses := []string{"Bass Voice", "Bass Guitar", "Double Bass", "Bass Clarinet", "Bass Trombone", "Bass Tuba"}
		found := make([]*instrument.Preset, len(basses))
		for i, name := range basses {
			found[i] = instrument.Builtins().MustFind(name)
		}
		for i := range found {
			for j := range found {
				if i != j {
					assert.False(t, found[i].Equal(found[j]), "%s and %s", basses[i], basses[j])
				}
			}
		}
	})
}

// КП-16: the built-in pianos.
func TestBuiltin_pianos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		fn           func() *instrument.Preset
		lower, upper string
		count        int
	}{
		{name: "88 keys", fn: instrument.Piano, lower: "A0", upper: "C8", count: 88},
		{name: "85 keys", fn: instrument.Piano85Keys, lower: "A0", upper: "A7", count: 85},
		{name: "92 keys", fn: instrument.Piano92Keys, lower: "F0", upper: "C8", count: 92},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ps := tt.fn()
			assertRange(t, rng(tt.lower, tt.upper), ps.Range())
			assert.Equal(t, tt.count, ps.Range().MustCount(trad()))
			g := instrument.MustNew(ps)
			assert.Equal(t, tt.count, g.MustCount())
		})
	}

	t.Run("membership of the 88 keys", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustNew(instrument.Piano())
		for x, want := range map[string]bool{"A0": true, "C8": true, "B♯7": true, "G♯0": false, "C♯8": false} {
			got, err := g.Contains(p(x))
			require.NoError(t, err)
			assert.Equal(t, want, got, x)
			got, err = instrument.Piano().Range().Contains(trad(), p(x))
			require.NoError(t, err)
			assert.Equal(t, want, got, x)
		}
	})
}

// КП-17: the built-in voices.
func TestBuiltin_voices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		fn           func() *instrument.Preset
		lower, upper string
	}{
		{name: "Soprano Voice", fn: instrument.SopranoVoice, lower: "C4", upper: "G5"},
		{name: "Alto Voice", fn: instrument.AltoVoice, lower: "G3", upper: "D5"},
		{name: "Tenor Voice", fn: instrument.TenorVoice, lower: "C3", upper: "G4"},
		{name: "Bass Voice", fn: instrument.BassVoice, lower: "F2", upper: "D4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ps := tt.fn()
			assert.Equal(t, tt.name, ps.Name())
			assertRange(t, rng(tt.lower, tt.upper), ps.Range())
		})
	}

	t.Run("К-20 the soprano voice checks КП-1", func(t *testing.T) {
		t.Parallel()
		g := instrument.MustNew(instrument.SopranoVoice())
		assert.True(t, g.MustContains(p("B♯3")))
		assert.False(t, g.MustContains(p("C♭4")))
		assert.Equal(t, instrument.Above, g.MustLocate(p("A5")))
	})

	t.Run("КП-9 alto and tenor voices", func(t *testing.T) {
		t.Parallel()
		got, ok := instrument.AltoVoice().Range().MustIntersect(trad(), instrument.TenorVoice().Range())
		require.True(t, ok)
		assertRange(t, rng("G3", "G4"), got)
	})
}

// An instrument from a built-in preset is independent of it and of every
// other instrument made from it.
func TestBuiltin_instruments(t *testing.T) {
	t.Parallel()
	a := instrument.MustNew(instrument.Violin())
	b := instrument.MustNew(instrument.Violin())
	require.NoError(t, a.Retune(h415()))
	require.NoError(t, a.SetName("First violin"))
	assert.True(t, b.Settings().Equal(settings.Standard()))
	assert.Equal(t, "Violin", b.Name())
	assert.True(t, instrument.Violin().SettingsSet().Equal(settings.Standard()))
	assert.False(t, settings.SameSource(a, b))
}

// An instrument from a built-in preset; the built-in catalog is searched by
// the exact name.
func ExampleViolin() {
	std := settings.Standard()
	p := std.MustParsePitch

	g := instrument.MustNew(instrument.Violin())
	fmt.Println(g.Name(), g.Range().Lower(), g.Range().Upper())
	fmt.Println(g.MustContains(p("G3")), g.MustContains(p("F3")))
	fmt.Println(g.Settings().Equal(std))
	// Output:
	// Violin G ASPN 3 E ASPN 7
	// true false
	// true
}

// The built-in catalog in the order of the table; the search is exact.
func ExampleBuiltins() {
	c := instrument.Builtins()
	fmt.Println(c.Len())

	piano, ok := c.Lookup("Piano (85 keys)")
	fmt.Println(piano.Range().Lower(), piano.Range().Upper(), ok)

	_, ok = c.Lookup("violin")
	fmt.Println(ok)
	_, err := c.Find("Bass")
	fmt.Println(errors.Is(err, instrument.ErrNotFound))
	// Output:
	// 39
	// A ASPN 0 A ASPN 7 true
	// false
	// true
}
