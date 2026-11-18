package naming_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/interval/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkPanic(t *testing.T, call func(), want error) {
	t.Helper()
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered)
		err, ok := recovered.(error)
		require.True(t, ok, "panic must contain error, got %T", recovered)
		assert.ErrorIs(t, err, want)
	}()
	call()
}

func TestCatalogGeneric(t *testing.T) {
	t.Parallel()
	ru := []string{"прима", "секунда", "терция", "кварта", "квинта", "секста", "септима", "октава", "нона", "децима", "ундецима", "дуодецима", "терцдецима", "квартдецима", "квинтдецима"}
	en := []string{"unison", "second", "third", "fourth", "fifth", "sixth", "seventh", "octave", "ninth", "tenth", "eleventh", "twelfth", "thirteenth", "fourteenth", "fifteenth"}
	for _, tc := range []struct {
		name  string
		c     naming.Catalog
		words []string
	}{{"Russian", naming.Russian(), ru}, {"English", naming.English(), en}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for index, word := range tc.words {
				g := interval.MustNewGeneric(index + 1)
				entry := tc.c.Generic(g)
				for _, read := range []func() (string, bool, error){entry.Name, entry.NameFull} {
					name, ok, err := read()
					require.NoError(t, err)
					assert.True(t, ok)
					assert.Equal(t, word, name)
				}
				got, err := tc.c.ParseGeneric(strings.ToUpper(word))
				require.NoError(t, err)
				assert.Equal(t, g, got)
				assert.Equal(t, g, tc.c.MustParseGeneric(word))
				name, ok := entry.MustName()
				assert.True(t, ok)
				assert.Equal(t, word, name)
				name, ok = entry.MustNameFull()
				assert.True(t, ok)
				assert.Equal(t, word, name)
			}
			for _, size := range []int{16, interval.MaxSize} {
				entry := tc.c.Generic(interval.MustNewGeneric(size))
				name, ok, err := entry.Name()
				require.NoError(t, err)
				assert.False(t, ok)
				assert.Empty(t, name)
				name, ok, err = entry.NameFull()
				require.NoError(t, err)
				assert.False(t, ok)
				assert.Empty(t, name)
			}
		})
	}
}

func TestCatalogSpecific(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, token, ruShort, ruFull, enShort, enFull string
		ruOK, enOK                                    bool
	}{
		{"minor", "-m3", "м.3", "малая терция", "m3", "minor third", true, true},
		{"major", "M3", "б.3", "большая терция", "M3", "major third", true, true},
		{"perfect compound", "P15", "ч.15", "чистая квинтдецима", "P15", "perfect fifteenth", true, true},
		{"doubly diminished", "dd5", "дв.ум.5", "дважды уменьшенная квинта", "dd5", "doubly diminished fifth", true, true},
		{"triply augmented", "AAA4", "", "трижды увеличенная кварта", "AAA4", "triply augmented fourth", true, true},
		{"outside dictionary", "M16", "", "", "", "", false, false},
		{"four augmentations", "{+4}4", "", "", "", "", false, false},
		{"fractional", "{+1/4}3", "", "", "", "", false, false},
		{"neutral third", "n3", "", "нейтральная терция", "n3", "neutral third", true, true},
		{"neutral second", "n2", "", "", "n2", "neutral second", false, true},
		{"neutral sixth", "n6", "", "", "n6", "neutral sixth", false, true},
		{"neutral seventh", "n7", "", "", "n7", "neutral seventh", false, true},
		{"neutral ninth", "n9", "", "", "n9", "neutral ninth", false, true},
		{"neutral tenth absent", "n10", "", "", "", "", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := interval.MustParse(tc.token)
			assert.Equal(t, tc.token, i.String())
			for _, lang := range []struct {
				name        string
				c           naming.Catalog
				short, full string
				ok          bool
			}{{"ru", naming.Russian(), tc.ruShort, tc.ruFull, tc.ruOK}, {"en", naming.English(), tc.enShort, tc.enFull, tc.enOK}} {
				t.Run(lang.name, func(t *testing.T) {
					t.Parallel()
					e := lang.c.Specific(i)
					short, ok, err := e.Name()
					require.NoError(t, err)
					assert.Equal(t, lang.short, short)
					assert.Equal(t, lang.short != "", ok)
					full, ok, err := e.NameFull()
					require.NoError(t, err)
					assert.Equal(t, lang.full, full)
					assert.Equal(t, lang.ok, ok)
					for _, text := range []string{short, full} {
						if text == "" {
							continue
						}
						parsed, err := lang.c.ParseInterval(text)
						require.NoError(t, err)
						assert.True(t, i.MustUp().Equal(parsed))
						assert.True(t, parsed.Equal(lang.c.MustParseInterval(text)))
					}
				})
			}
		})
	}
	// Legacy names and mixed-language parsing remain available unchanged.
	third := interval.MustParse("-m3")
	name, ok, err := third.Name()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "minor third", name)
	name, ok, err = third.RussianName()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "малая терция", name)
	name, ok, err = third.RussianAbbreviation()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "м.3", name)
	for _, text := range []string{"minor third", "малая терция", "м.3"} {
		assert.True(t, interval.MustParseName(text).Equal(third.MustUp()))
	}
	neutral := interval.MustFromDeviation(3, alteration.MustFromFraction(-1, 2))
	name, ok, err = neutral.Name()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "neutral third", name)
	assert.True(t, interval.MustParseName("нейтральная терция").Equal(neutral))
}

func TestCatalogParsingBoundaries(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		c    naming.Catalog
		text string
		want interval.Interval
		err  error
	}{
		{"major case", naming.English(), "M3", interval.MajorThird(), nil},
		{"minor case", naming.English(), "m3", interval.MinorThird(), nil},
		{"English words case", naming.English(), "MINOR THIRD", interval.MinorThird(), nil},
		{"Russian words case", naming.Russian(), "МАЛАЯ ТЕРЦИЯ", interval.MinorThird(), nil},
		{"Russian synonym", naming.Russian(), "увеличенная кварта", interval.AugmentedFourth(), nil},
		{"foreign Russian", naming.English(), "малая терция", interval.Interval{}, interval.ErrParse},
		{"foreign English", naming.Russian(), "minor third", interval.Interval{}, interval.ErrParse},
		{"foreign neutral short", naming.Russian(), "n3", interval.Interval{}, interval.ErrParse},
		{"absent Russian neutral second", naming.Russian(), "нейтральная секунда", interval.Interval{}, interval.ErrParse},
		{"absent English neutral tenth", naming.English(), "neutral tenth", interval.Interval{}, interval.ErrParse},
		{"English synonym", naming.English(), "double octave", interval.PerfectFifteenth(), nil},
		{"English ordinal", naming.English(), "minor 3rd", interval.MinorThird(), nil},
		{"Russian orthography", naming.Russian(), "уменьшённая квинта", interval.DiminishedFifth(), nil},
		{"minus English", naming.English(), "-m3", interval.Interval{}, interval.ErrParse},
		{"minus Russian", naming.Russian(), "-м.3", interval.Interval{}, interval.ErrParse},
		{"leading space", naming.English(), " m3", interval.Interval{}, interval.ErrParse},
		{"trailing space", naming.Russian(), "м.3 ", interval.Interval{}, interval.ErrParse},
		{"double space", naming.English(), "minor  third", interval.Interval{}, interval.ErrParse},
		{"unknown", naming.English(), "tritone", interval.Interval{}, interval.ErrParse},
		{"missing quality", naming.Russian(), "терция", interval.Interval{}, interval.ErrParse},
		{"wrong group", naming.English(), "perfect third", interval.Interval{}, interval.ErrInvalidArgument},
		{"diminished unison", naming.Russian(), "уменьшенная прима", interval.Interval{}, interval.ErrInvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.c.ParseInterval(tc.text)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				assert.Equal(t, interval.Interval{}, got)
			} else {
				require.NoError(t, err)
				assert.True(t, tc.want.Equal(got))
			}
		})
	}
	for _, tc := range []struct {
		c    naming.Catalog
		text string
	}{{naming.English(), "терция"}, {naming.Russian(), "third"}, {naming.English(), " third"}, {naming.Russian(), "терция "}, {naming.English(), "sixteenth"}, {naming.Russian(), ""}} {
		g, err := tc.c.ParseGeneric(tc.text)
		assert.ErrorIs(t, err, interval.ErrParse)
		assert.Equal(t, interval.Generic{}, g)
	}
	checkPanic(t, func() { naming.English().MustParseGeneric("терция") }, interval.ErrParse)
	checkPanic(t, func() { naming.Russian().MustParseInterval("minor third") }, interval.ErrParse)
}

func TestCatalogUnset(t *testing.T) {
	t.Parallel()
	var c naming.Catalog
	for _, e := range []naming.Entry{{}, c.Generic(interval.MustNewGeneric(3)), c.Specific(interval.MajorThird()), naming.Russian().Generic(interval.Generic{}), naming.English().Specific(interval.Interval{})} {
		for _, read := range []func() (string, bool, error){e.Name, e.NameFull} {
			name, ok, err := read()
			assert.ErrorIs(t, err, interval.ErrUnset)
			assert.Empty(t, name)
			assert.False(t, ok)
		}
		checkPanic(t, func() { e.MustName() }, interval.ErrUnset)
		checkPanic(t, func() { e.MustNameFull() }, interval.ErrUnset)
	}
	g, err := c.ParseGeneric("third")
	assert.ErrorIs(t, err, interval.ErrUnset)
	assert.Equal(t, interval.Generic{}, g)
	i, err := c.ParseInterval("m3")
	assert.ErrorIs(t, err, interval.ErrUnset)
	assert.Equal(t, interval.Interval{}, i)
	checkPanic(t, func() { c.MustParseGeneric("third") }, interval.ErrUnset)
	checkPanic(t, func() { c.MustParseInterval("m3") }, interval.ErrUnset)
}

func TestCatalogConcurrentRead(t *testing.T) {
	t.Parallel()
	ru, en := naming.Russian(), naming.English()
	results := make([]error, 16)
	var wg sync.WaitGroup
	for worker := range results {
		wg.Go(func() {
			for range 25 {
				r, rok, re := ru.Specific(interval.MinorThird()).NameFull()
				e, eok, ee := en.Specific(interval.MinorThird()).NameFull()
				if re != nil || ee != nil || !rok || !eok || r != "малая терция" || e != "minor third" {
					results[worker] = errors.New("catalogs are not independent immutable readers")
					return
				}
			}
		})
	}
	wg.Wait()
	for _, err := range results {
		assert.NoError(t, err)
	}
}

func TestSpecificDictionaryAtEverySize(t *testing.T) {
	t.Parallel()
	ru := []string{"прима", "секунда", "терция", "кварта", "квинта", "секста", "септима", "октава", "нона", "децима", "ундецима", "дуодецима", "терцдецима", "квартдецима", "квинтдецима"}
	en := []string{"unison", "second", "third", "fourth", "fifth", "sixth", "seventh", "octave", "ninth", "tenth", "eleventh", "twelfth", "thirteenth", "fourteenth", "fifteenth"}
	for index := range 15 {
		size := index + 1
		perfect := size == 1 || size%7 == 1 || size%7 == 4 || size%7 == 5
		choices := []struct {
			deviation                   int
			short, ruQuality, enQuality string
		}{{0, "M", "большая", "major"}, {-1, "m", "малая", "minor"}, {1, "A", "увеличенная", "augmented"}, {-2, "d", "уменьшенная", "diminished"}}
		if perfect {
			choices = []struct {
				deviation                   int
				short, ruQuality, enQuality string
			}{{0, "P", "чистая", "perfect"}, {1, "A", "увеличенная", "augmented"}, {-1, "d", "уменьшенная", "diminished"}}
		}
		for _, quality := range choices {
			if size == 1 && quality.deviation < 0 {
				continue
			}
			i := interval.MustFromDeviation(size, alteration.MustFromInt(quality.deviation))
			for _, tc := range []struct {
				c             naming.Catalog
				word, quality string
			}{{naming.Russian(), ru[index], quality.ruQuality}, {naming.English(), en[index], quality.enQuality}} {
				text, ok, err := tc.c.Specific(i).NameFull()
				require.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, tc.quality+" "+tc.word, text)
				parsed, err := tc.c.ParseInterval(text)
				require.NoError(t, err)
				assert.True(t, i.Equal(parsed))
			}
		}
	}
}
