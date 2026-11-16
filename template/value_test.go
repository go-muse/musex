package template_test

import (
	"errors"
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKindAndMeasure(t *testing.T) {
	t.Parallel()
	t.Run("kind", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			value template.Kind
			text  string
			valid bool
		}{
			{template.KindUnset, "unset", false},
			{template.Closed, "closed", true},
			{template.Open, "open", true},
			{template.Kind(255), "Kind(255)", false},
		}
		for _, tc := range cases {
			assert.Equal(t, tc.text, tc.value.String())
			assert.Equal(t, tc.valid, tc.value.IsValid())
		}
		assert.Equal(t, 0, int(template.KindUnset))
		assert.Equal(t, 1, int(template.Closed))
		assert.Equal(t, 2, int(template.Open))
	})
	t.Run("measure", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			value template.Measure
			text  string
			valid bool
		}{
			{template.MeasureUnset, "unset", false},
			{template.FromFirst, "from-first", true},
			{template.Adjacent, "adjacent", true},
			{template.Measure(255), "Measure(255)", false},
		}
		for _, tc := range cases {
			assert.Equal(t, tc.text, tc.value.String())
			assert.Equal(t, tc.valid, tc.value.IsValid())
		}
		assert.Equal(t, 0, int(template.MeasureUnset))
		assert.Equal(t, 1, int(template.FromFirst))
		assert.Equal(t, 2, int(template.Adjacent))
	})
	assert.Equal(t, 1<<16, template.MaxPlaces)
}

func TestTemplateZeroValue(t *testing.T) {
	t.Parallel()
	var zero template.Template
	assert.False(t, zero.IsValid())
	assert.Equal(t, template.KindUnset, zero.Kind())
	assert.False(t, zero.IsClosed())
	assert.False(t, zero.IsOpen())
	assert.Zero(t, zero.Len())
	assert.Nil(t, zero.AdjacentIntervals())
	cycle, ok := zero.Cycle()
	assert.False(t, ok)
	assert.False(t, cycle.IsValid())
	assert.False(t, zero.IsOctave())
	assert.Equal(t, "Template(unset)", zero.String())
	assert.False(t, zero.Equal(zero))
	assert.Zero(t, zero.Compare(template.Template{}))
	place, ok := zero.RotationOf(zero)
	assert.Zero(t, place)
	assert.False(t, ok)
}

func TestFromIntervals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		text      string
		isOpen    bool
		places    int
		cycle     string
		hasCycle  bool
		octave    bool
		placeText []string
	}{
		{name: "ionian", text: "M2 M2 m2 M2 M2 M2 m2", places: 7, cycle: "P8", hasCycle: true, octave: true, placeText: []string{"P1", "M2", "M3", "P4", "P5", "M6", "M7"}},
		{name: "single octave", text: "P8", places: 1, cycle: "P8", hasCycle: true, octave: true, placeText: []string{"P1"}},
		{name: "single fifth closed", text: "P5", places: 1, cycle: "P5", hasCycle: true, placeText: []string{"P1"}},
		{name: "fifth and fourth", text: "P5 P4", places: 2, cycle: "P8", hasCycle: true, octave: true, placeText: []string{"P1", "P5"}},
		{name: "blues", text: "m3 M2 A1 m2 m3 M2", places: 6, cycle: "P8", hasCycle: true, octave: true, placeText: []string{"P1", "m3", "P4", "A4", "P5", "m7"}},
		{name: "diminished second", text: "d2 A7", places: 2, cycle: "P8", hasCycle: true, octave: true, placeText: []string{"P1", "d2"}},
		{name: "mixed interval sizes", text: "A3 dd2 A2 P4", places: 4, cycle: "P8", hasCycle: true, octave: true, placeText: []string{"P1", "A3", "P4", "P5"}},
		{name: "obikhod", text: "M2 M2 m2", places: 3, cycle: "P4", hasCycle: true, placeText: []string{"P1", "M2", "M3"}},
		{name: "wide cycle", text: "M9 m7", places: 2, cycle: "P15", hasCycle: true, placeText: []string{"P1", "M9"}},
		{name: "seven major seconds", text: "M2 M2 M2 M2 M2 M2 M2", places: 7, cycle: "AA8", hasCycle: true, placeText: []string{"P1", "M2", "M3", "A4", "A5", "A6", "A7"}},
		{name: "two fourths", text: "P4 P4", places: 2, cycle: "m7", hasCycle: true, placeText: []string{"P1", "P4"}},
		{name: "two fifths", text: "P5 P5", places: 2, cycle: "M9", hasCycle: true, placeText: []string{"P1", "P5"}},
		{name: "open trichord", text: "m3 M2", isOpen: true, places: 3, placeText: []string{"P1", "m3", "P4"}},
		{name: "open fifth", text: "P5", isOpen: true, places: 2, placeText: []string{"P1", "P5"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			input := intervals(tc.text)
			var got template.Template
			var err error
			if tc.isOpen {
				got, err = template.OpenFromIntervals(input)
			} else {
				got, err = template.ClosedFromIntervals(input)
			}
			require.NoError(t, err)
			require.True(t, got.IsValid())
			assert.Equal(t, tc.places, got.Len())
			wantAdjacent := tc.places
			if tc.isOpen {
				wantAdjacent--
			}
			assert.Equal(t, wantAdjacent, len(input))
			assert.Equal(t, !tc.isOpen, got.IsClosed())
			assert.Equal(t, tc.isOpen, got.IsOpen())
			assert.Equal(t, tc.isOpen, got.Kind() == template.Open)
			assert.Equal(t, intervalTexts(input), intervalTexts(got.AdjacentIntervals()))
			cycle, ok := got.Cycle()
			assert.Equal(t, tc.hasCycle, ok)
			if tc.hasCycle {
				requireInterval(t, tc.cycle, cycle)
			} else {
				assert.False(t, cycle.IsValid())
			}
			assert.Equal(t, tc.octave, got.IsOctave())
			places, err := got.PlaceIntervals()
			require.NoError(t, err)
			assert.Equal(t, tc.placeText, intervalTexts(places))
		})
	}
}

func TestFromIntervalsErrorsAndOrder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		isOpen   bool
		input    []interval.Interval
		want     error
		interval error
	}{
		{name: "unset closed", input: []interval.Interval{{}}, want: template.ErrUnset, interval: interval.ErrUnset},
		{name: "unset open", isOpen: true, input: []interval.Interval{interval.MajorSecond(), {}}, want: template.ErrUnset, interval: interval.ErrUnset},
		{name: "empty closed", want: template.ErrEmptyInput},
		{name: "empty open", isOpen: true, want: template.ErrEmptyInput},
		{name: "unison closed", input: intervals("M2 P1 M7"), want: template.ErrInvalidInterval},
		{name: "descending closed", input: intervals("M2 -M2 M7"), want: template.ErrInvalidInterval},
		{name: "unison open", isOpen: true, input: intervals("M2 P1"), want: template.ErrInvalidInterval},
		{name: "descending open", isOpen: true, input: intervals("-A1"), want: template.ErrInvalidInterval},
		{name: "unset before invalid", input: []interval.Interval{{}, interval.PerfectUnison()}, want: template.ErrUnset, interval: interval.ErrUnset},
		{name: "overflow closed", want: template.ErrOutOfRange},
	}
	for _, tc := range cases {
		if tc.name == "overflow closed" {
			tc.input = []interval.Interval{
				interval.MustFromDeviation(2, alteration.MustFromInt(alteration.MaxMagnitude)),
				interval.MustFromDeviation(2, alteration.MustFromInt(alteration.MaxMagnitude)),
			}
			tc.interval = interval.ErrOutOfRange
		}
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got template.Template
			var err error
			if tc.isOpen {
				got, err = template.OpenFromIntervals(tc.input)
			} else {
				got, err = template.ClosedFromIntervals(tc.input)
			}
			require.ErrorIs(t, err, tc.want)
			assert.False(t, got.IsValid())
			if tc.interval != nil {
				assert.ErrorIs(t, err, tc.interval)
			}
			if errors.Is(tc.want, template.ErrOutOfRange) {
				assert.ErrorIs(t, err, alteration.ErrOutOfRange)
			}
		})
	}
}

func TestPlaceIntervalConstructors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		places []string
		cycle  string
		isOpen bool
		want   string
	}{
		{name: "blues", places: []string{"P1", "m3", "P4", "A4", "P5", "m7"}, cycle: "P8", want: "m3 M2 A1 m2 m3 M2"},
		{name: "blues flat five", places: []string{"P1", "m3", "P4", "d5", "P5", "m7"}, cycle: "P8", want: "m3 M2 m2 A1 m3 M2"},
		{name: "one place", places: []string{"P1"}, cycle: "P8", want: "P8"},
		{name: "obikhod", places: []string{"P1", "M2", "M3"}, cycle: "P4", want: "M2 M2 m2"},
		{name: "wide cycle", places: []string{"P1", "M9"}, cycle: "P15", want: "M9 m7"},
		{name: "open trichord", places: []string{"P1", "m3", "P4"}, isOpen: true, want: "m3 M2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			places := intervals(stringsJoin(tc.places))
			var got template.Template
			var err error
			if tc.isOpen {
				got, err = template.OpenFromPlaceIntervals(places)
			} else {
				got, err = template.ClosedFromPlaceIntervals(places, interval.MustParse(tc.cycle))
			}
			require.NoError(t, err)
			want := closed(tc.want)
			if tc.isOpen {
				want = opened(tc.want)
			}
			requireTemplate(t, want, got)
			readPlaces, err := got.PlaceIntervals()
			require.NoError(t, err)
			assert.Equal(t, tc.places, intervalTexts(readPlaces))
			if tc.isOpen {
				roundTrip, err := template.OpenFromPlaceIntervals(readPlaces)
				require.NoError(t, err)
				assert.True(t, roundTrip.Equal(got))
			} else {
				cycle, ok := got.Cycle()
				require.True(t, ok)
				roundTrip, err := template.ClosedFromPlaceIntervals(readPlaces, cycle)
				require.NoError(t, err)
				assert.True(t, roundTrip.Equal(got))
			}
		})
	}
}

func stringsJoin(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	return result
}

func TestPlaceIntervalConstructorErrors(t *testing.T) {
	t.Parallel()
	p1 := interval.PerfectUnison()
	cases := []struct {
		name   string
		places []interval.Interval
		cycle  interval.Interval
		isOpen bool
		want   error
	}{
		{name: "closed unset cycle", places: []interval.Interval{p1}, want: template.ErrUnset},
		{name: "closed unset place", places: []interval.Interval{p1, {}}, cycle: interval.PerfectOctave(), want: template.ErrUnset},
		{name: "closed empty and unset cycle", cycle: interval.PerfectOctave(), want: template.ErrEmptyInput},
		{name: "open unset before empty", places: []interval.Interval{{}}, isOpen: true, want: template.ErrUnset},
		{name: "open empty", isOpen: true, want: template.ErrEmptyInput},
		{name: "open one place", places: []interval.Interval{p1}, isOpen: true, want: template.ErrEmptyInput},
		{name: "first not unison closed", places: intervals("M2 M3"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "first not unison open", places: intervals("M2 M3"), isOpen: true, want: template.ErrInvalidInterval},
		{name: "decreasing places", places: intervals("P1 M3 M2"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "equal places", places: intervals("P1 P4 P4"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "descending place", places: intervals("P1 -M2"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "cycle equal to last", places: intervals("P1 P5"), cycle: interval.PerfectFifth(), want: template.ErrInvalidInterval},
		{name: "cycle below last", places: intervals("P1 P5"), cycle: interval.PerfectFourth(), want: template.ErrInvalidInterval},
		{name: "cycle below compound place", places: intervals("P1 P5 P8"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "open decreasing", places: intervals("P1 M3 M2"), isOpen: true, want: template.ErrInvalidInterval},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got template.Template
			var err error
			if tc.isOpen {
				got, err = template.OpenFromPlaceIntervals(tc.places)
			} else {
				got, err = template.ClosedFromPlaceIntervals(tc.places, tc.cycle)
			}
			require.ErrorIs(t, err, tc.want)
			assert.False(t, got.IsValid())
		})
	}

	t.Run("unrepresentable difference", func(t *testing.T) {
		t.Parallel()
		large := interval.MustFromDeviation(2, alteration.MustFromInt(-alteration.MaxMagnitude))
		larger := interval.MustFromDeviation(3, alteration.MustFromInt(alteration.MaxMagnitude))
		got, err := template.OpenFromPlaceIntervals([]interval.Interval{p1, large, larger})
		require.ErrorIs(t, err, template.ErrOutOfRange)
		assert.ErrorIs(t, err, interval.ErrOutOfRange)
		assert.False(t, got.IsValid())
	})
}

func TestSpellingsConstructors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		names  string
		cycle  string
		isOpen bool
		octave bool
		want   string
	}{
		{name: "harmonic minor", names: "C D E♭ F G A♭ B", cycle: "P8", want: "M2 m2 M2 M2 m2 A2 m2"},
		{name: "other tonic", names: "D E F G A B♭ C♯", cycle: "P8", want: "M2 m2 M2 M2 m2 A2 m2"},
		{name: "double sharps", names: "G♯ A♯ B C♯ D♯ E F𝄪", cycle: "P8", want: "M2 m2 M2 M2 m2 A2 m2"},
		{name: "blues augmented fourth", names: "C E♭ F F♯ G B♭", cycle: "P8", want: "m3 M2 A1 m2 m3 M2"},
		{name: "blues flat fifth", names: "C E♭ F G♭ G B♭", cycle: "P8", want: "m3 M2 m2 A1 m3 M2"},
		{name: "wrap from B", names: "B C D", cycle: "P8", want: "m2 M2 M6"},
		{name: "diminished seventh", names: "C D C♭", cycle: "P8", want: "M2 d7 A1"},
		{name: "one name", names: "C", cycle: "P8", want: "P8"},
		{name: "obikhod", names: "G A B", cycle: "P4", want: "M2 M2 m2"},
		{name: "wider than octave", names: "C E G C E", cycle: "P15", want: "M3 m3 P4 M3 m6"},
		{name: "open trichord", names: "E G A", isOpen: true, want: "m3 M2"},
		{name: "open written upwards", names: "C E D", isOpen: true, want: "M3 m7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got template.Template
			var err error
			switch {
			case tc.octave:
				got, err = template.FromSpellingsInOctave(names(tc.names))
			case tc.isOpen:
				got, err = template.OpenFromSpellings(names(tc.names))
			default:
				got, err = template.ClosedFromSpellings(names(tc.names), interval.MustParse(tc.cycle))
			}
			require.NoError(t, err)
			want := closed(tc.want)
			if tc.isOpen {
				want = opened(tc.want)
			}
			requireTemplate(t, want, got)
			if tc.cycle == "P8" && !tc.isOpen {
				fromOctave, err := template.FromSpellingsInOctave(names(tc.names))
				require.NoError(t, err)
				assert.True(t, fromOctave.Equal(got))
			}
		})
	}
}

func TestSpellingsConstructorErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		names  []spelling.Spelling
		cycle  interval.Interval
		isOpen bool
		octave bool
		want   error
	}{
		{name: "closed unset cycle", names: names("C"), want: template.ErrUnset},
		{name: "closed unset name", names: []spelling.Spelling{spelling.C(), {}}, cycle: interval.PerfectOctave(), want: template.ErrUnset},
		{name: "open unset name", names: []spelling.Spelling{spelling.C(), {}}, isOpen: true, want: template.ErrUnset},
		{name: "closed empty", cycle: interval.PerfectOctave(), want: template.ErrEmptyInput},
		{name: "open empty", isOpen: true, want: template.ErrEmptyInput},
		{name: "open one name", names: names("C"), isOpen: true, want: template.ErrEmptyInput},
		{name: "descending spelling closes above last", names: names("C E D"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "equal names closed", names: names("C C"), cycle: interval.PerfectOctave(), want: template.ErrInvalidInterval},
		{name: "cycle not above names", names: names("G A B C"), cycle: interval.PerfectFourth(), want: template.ErrInvalidInterval},
		{name: "equal names open", names: names("C C D"), isOpen: true, want: template.ErrInvalidInterval},
		{name: "octave descending spelling", names: names("C E D"), octave: true, want: template.ErrInvalidInterval},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got template.Template
			var err error
			switch {
			case tc.octave:
				got, err = template.FromSpellingsInOctave(tc.names)
			case tc.isOpen:
				got, err = template.OpenFromSpellings(tc.names)
			default:
				got, err = template.ClosedFromSpellings(tc.names, tc.cycle)
			}
			require.ErrorIs(t, err, tc.want)
			assert.False(t, got.IsValid())
			if errors.Is(tc.want, template.ErrUnset) && len(tc.names) > 1 && !tc.names[1].IsValid() {
				assert.ErrorIs(t, err, interval.ErrUnset)
			}
		})
	}
}

func majorSecondPlaces(count int) []interval.Interval {
	places := make([]interval.Interval, count)
	places[0] = interval.PerfectUnison()
	for i := 1; i < count; i++ {
		places[i] = places[i-1].MustAdd(interval.MajorSecond())
	}
	return places
}

func TestMaxPlaces(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 65536, template.MaxPlaces)
	cases := []struct {
		name   string
		count  int
		isOpen bool
		want   error
	}{
		{name: "closed at limit", count: template.MaxPlaces},
		{name: "closed over limit", count: template.MaxPlaces + 1, want: template.ErrInvalidArgument},
		{name: "open at limit", count: template.MaxPlaces, isOpen: true},
		{name: "open over limit", count: template.MaxPlaces + 1, isOpen: true, want: template.ErrInvalidArgument},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			places := majorSecondPlaces(tc.count)
			adjacentCount := tc.count
			if tc.isOpen {
				adjacentCount--
			}
			adjacent := slices.Repeat([]interval.Interval{interval.MajorSecond()}, adjacentCount)
			var fromAdjacent template.Template
			var err error
			if tc.isOpen {
				fromAdjacent, err = template.OpenFromIntervals(adjacent)
			} else {
				fromAdjacent, err = template.ClosedFromIntervals(adjacent)
			}
			if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
				assert.False(t, fromAdjacent.IsValid())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.count, fromAdjacent.Len())
			}

			var fromPlaces template.Template
			if tc.isOpen {
				fromPlaces, err = template.OpenFromPlaceIntervals(places)
			} else {
				cycle := places[len(places)-1].MustAdd(interval.MajorSecond())
				fromPlaces, err = template.ClosedFromPlaceIntervals(places, cycle)
			}
			if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
				assert.False(t, fromPlaces.IsValid())
			} else {
				require.NoError(t, err)
				assert.True(t, fromPlaces.Equal(fromAdjacent))
			}
		})
	}

	t.Run("limit precedes invalid interval", func(t *testing.T) {
		t.Parallel()
		input := slices.Repeat([]interval.Interval{interval.PerfectUnison()}, template.MaxPlaces+1)
		got, err := template.ClosedFromIntervals(input)
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.False(t, errors.Is(err, template.ErrInvalidInterval))
		assert.False(t, got.IsValid())
		places := majorSecondPlaces(template.MaxPlaces + 1)
		places[0] = interval.MajorSecond()
		got, err = template.OpenFromPlaceIntervals(places)
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.False(t, errors.Is(err, template.ErrInvalidInterval))
		assert.False(t, got.IsValid())
	})

	t.Run("spellings at and over limit", func(t *testing.T) {
		// These cases are sequential because they intentionally allocate large inputs.
		alternating := func(count int) []spelling.Spelling {
			result := make([]spelling.Spelling, count)
			for i := range count {
				result[i] = spelling.C()
				if i%2 == 1 {
					result[i] = spelling.D()
				}
			}
			return result
		}
		got, err := template.OpenFromSpellings(alternating(template.MaxPlaces))
		require.NoError(t, err)
		assert.Equal(t, template.MaxPlaces, got.Len())
		got, err = template.OpenFromSpellings(alternating(template.MaxPlaces + 1))
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.False(t, got.IsValid())

		// C-D pairs have M2/m7 intervals; the complete natural cycle has this size.
		cycle := interval.MustFromDeviation(7*(template.MaxPlaces/2)+1, alteration.Natural())
		got, err = template.ClosedFromSpellings(alternating(template.MaxPlaces), cycle)
		require.NoError(t, err)
		assert.Equal(t, template.MaxPlaces, got.Len())
		got, err = template.ClosedFromSpellings(alternating(template.MaxPlaces+1), cycle)
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.False(t, got.IsValid())
		got, err = template.FromSpellingsInOctave(alternating(template.MaxPlaces + 1))
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.False(t, got.IsValid())
	})
}

func TestPlacesAndBetweenPlaces(t *testing.T) {
	t.Parallel()
	t.Run("obikhod virtual places", func(t *testing.T) {
		t.Parallel()
		tmpl := closed("M2 M2 m2")
		got, err := tmpl.PlaceIntervals()
		require.NoError(t, err)
		assert.Equal(t, []string{"P1", "M2", "M3"}, intervalTexts(got))
		want := []string{"P1", "M2", "M3", "P4", "P5", "M6", "m7", "P8", "M9", "m10"}
		for place, expected := range want {
			got, err := tmpl.PlaceInterval(place + 1)
			require.NoError(t, err)
			requireInterval(t, expected, got)
		}
		for place, expected := range []string{"M2", "M2", "m2", "M2", "M2", "m2"} {
			got, err := tmpl.AdjacentIntervalAfter(place + 1)
			require.NoError(t, err)
			requireInterval(t, expected, got)
		}
	})

	t.Run("control intervals", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			template string
			i, j     int
			want     string
		}{
			{template: "harmonic-minor", i: 6, j: 7, want: "A2"},
			{template: "ionian", i: 4, j: 7, want: "A4"},
			{template: "ionian", i: 7, j: 4, want: "-A4"},
			{template: "ionian", i: 3, j: 3, want: "P1"},
			{template: "ionian", i: 7, j: 8, want: "m2"},
			{template: "ionian", i: 1, j: 8, want: "P8"},
			{template: "ionian", i: 1, j: 9, want: "M9"},
			{template: "ionian", i: 1, j: 15, want: "P15"},
			{template: "ionian", i: 2, j: 1, want: "-M2"},
			{template: "harmonic-minor", i: 3, j: 7, want: "A5"},
			{template: "harmonic-minor", i: 7, j: 10, want: "d4"},
			{template: "harmonic-minor", i: 7, j: 13, want: "d7"},
			{template: "harmonic-major", i: 3, j: 6, want: "d4"},
			{template: "harmonic-major", i: 6, j: 10, want: "A5"},
			{template: "pentatonic-major", i: 3, j: 4, want: "m3"},
			{template: "pentatonic-major", i: 5, j: 6, want: "m3"},
			{template: "blues", i: 3, j: 4, want: "A1"},
			{template: "blues", i: 4, j: 3, want: "-A1"},
			{template: "M2 M2 m2", i: 1, j: 4, want: "P4"},
			{template: "M2 M2 m2", i: 3, j: 10, want: "d8"},
		}
		for _, tc := range cases {
			t.Run(tc.template+"/"+tc.want, func(t *testing.T) {
				t.Parallel()
				var tmpl template.Template
				switch tc.template {
				case "harmonic-minor":
					tmpl = template.HarmonicMinor()
				case "ionian":
					tmpl = template.Ionian()
				case "harmonic-major":
					tmpl = template.HarmonicMajor()
				case "pentatonic-major":
					tmpl = template.PentatonicMajor()
				case "blues":
					tmpl = template.Blues()
				default:
					tmpl = closed(tc.template)
				}
				got, err := tmpl.BetweenPlaces(tc.i, tc.j)
				require.NoError(t, err)
				requireInterval(t, tc.want, got)
			})
		}
	})

	t.Run("open boundaries", func(t *testing.T) {
		t.Parallel()
		tmpl := opened("m3 M2")
		got, err := tmpl.PlaceInterval(3)
		require.NoError(t, err)
		requireInterval(t, "P4", got)
		got, err = tmpl.BetweenPlaces(3, 1)
		require.NoError(t, err)
		requireInterval(t, "-P4", got)
		got, err = tmpl.AdjacentIntervalAfter(2)
		require.NoError(t, err)
		requireInterval(t, "M2", got)
		_, ok := tmpl.Cycle()
		assert.False(t, ok)
		assert.False(t, tmpl.IsOctave())
		for _, call := range []func() error{
			func() error { _, err := tmpl.PlaceInterval(4); return err },
			func() error { _, err := tmpl.AdjacentIntervalAfter(3); return err },
			func() error { _, err := tmpl.BetweenPlaces(1, 4); return err },
		} {
			require.ErrorIs(t, call(), template.ErrInvalidArgument)
		}
	})

	t.Run("invalid and unset places", func(t *testing.T) {
		t.Parallel()
		valid := template.Ionian()
		cases := []struct {
			name string
			tmpl template.Template
			call func(template.Template) error
			want error
		}{
			{name: "zero place", tmpl: valid, call: func(v template.Template) error { _, err := v.PlaceInterval(0); return err }, want: template.ErrInvalidArgument},
			{name: "negative between", tmpl: valid, call: func(v template.Template) error { _, err := v.BetweenPlaces(-1, 2); return err }, want: template.ErrInvalidArgument},
			{name: "unset place", call: func(v template.Template) error { _, err := v.PlaceInterval(1); return err }, want: template.ErrUnset},
			{name: "unset adjacent", call: func(v template.Template) error { _, err := v.AdjacentIntervalAfter(1); return err }, want: template.ErrUnset},
			{name: "unset between", call: func(v template.Template) error { _, err := v.BetweenPlaces(1, 2); return err }, want: template.ErrUnset},
			{name: "unset place intervals", call: func(v template.Template) error { _, err := v.PlaceIntervals(); return err }, want: template.ErrUnset},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				require.ErrorIs(t, tc.call(tc.tmpl), tc.want)
			})
		}
	})
}

func TestExactArithmeticRegressions(t *testing.T) {
	t.Parallel()
	deviation := func(value int) interval.Interval {
		return interval.MustFromDeviation(2, alteration.MustFromInt(value))
	}
	t.Run("representable place with unrepresentable double cycle", func(t *testing.T) {
		t.Parallel()
		tmpl, err := template.ClosedFromIntervals([]interval.Interval{deviation(-6_000_000), deviation(alteration.MaxMagnitude)})
		require.NoError(t, err)
		cycle, ok := tmpl.Cycle()
		require.True(t, ok)
		assert.Equal(t, 3, cycle.Size())
		num, den := cycle.Deviation().Fraction()
		assert.Equal(t, 10_777_216, num)
		assert.Equal(t, 1, den)
		_, err = cycle.Mul(2)
		require.ErrorIs(t, err, interval.ErrOutOfRange)
		got, err := tmpl.PlaceInterval(6)
		require.NoError(t, err)
		assert.Equal(t, 6, got.Size())
		num, den = got.Deviation().Fraction()
		assert.Equal(t, 15_554_433, num)
		assert.Equal(t, 1, den)
		got, err = tmpl.BetweenPlaces(1, 6)
		require.NoError(t, err)
		assert.True(t, got.Equal(interval.MustFromDeviation(6, alteration.MustFromInt(15_554_433))))
	})

	t.Run("cycle after rotation with unrepresentable partial sum", func(t *testing.T) {
		t.Parallel()
		tmpl, err := template.ClosedFromIntervals([]interval.Interval{
			deviation(-6_000_000), deviation(alteration.MaxMagnitude), deviation(1_000_000),
		})
		require.NoError(t, err)
		rotated, err := tmpl.Rotate(2)
		require.NoError(t, err)
		cycle, ok := rotated.Cycle()
		require.True(t, ok)
		assert.Equal(t, 4, cycle.Size())
		num, den := cycle.Deviation().Fraction()
		assert.Equal(t, 11_777_217, num)
		assert.Equal(t, 1, den)
		places, err := rotated.MismatchingPlaces(rotated, template.FromFirst)
		require.NoError(t, err)
		assert.NotNil(t, places)
		assert.Empty(t, places)

		changedAdjacent := tmpl.AdjacentIntervals()
		changedAdjacent[0] = deviation(-6_000_001)
		changedOriginal, err := template.ClosedFromIntervals(changedAdjacent)
		require.NoError(t, err)
		changed, err := changedOriginal.Rotate(2)
		require.NoError(t, err)
		places, err = rotated.MismatchingPlaces(changed, template.FromFirst)
		require.NoError(t, err)
		assert.Equal(t, []int{4}, places)
		places, err = rotated.MismatchingPlaces(changed, template.Adjacent)
		require.NoError(t, err)
		assert.Equal(t, []int{3}, places)
	})

	t.Run("unrepresentable final place", func(t *testing.T) {
		t.Parallel()
		tmpl := template.MustClosedFromIntervals([]interval.Interval{deviation(alteration.MaxMagnitude)})
		_, err := tmpl.PlaceInterval(math.MaxInt)
		require.ErrorIs(t, err, template.ErrOutOfRange)
		assert.ErrorIs(t, err, interval.ErrOutOfRange)
		assert.ErrorIs(t, err, alteration.ErrOutOfRange)
		_, err = tmpl.BetweenPlaces(1, math.MaxInt)
		require.ErrorIs(t, err, template.ErrOutOfRange)
	})
}

func TestRotationEqualityAndCompare(t *testing.T) {
	t.Parallel()
	t.Run("diatonic rotations", func(t *testing.T) {
		t.Parallel()
		want := []template.Template{
			template.Ionian(), template.Dorian(), template.Phrygian(), template.Lydian(),
			template.Mixolydian(), template.Aeolian(), template.Locrian(),
		}
		for i, expected := range want {
			got, err := template.Ionian().Rotate(i + 1)
			require.NoError(t, err)
			requireTemplate(t, expected, got)
		}
		got, err := template.Dorian().Rotate(7)
		require.NoError(t, err)
		assert.True(t, got.Equal(template.Ionian()))
		_, err = template.Ionian().Rotate(8)
		require.ErrorIs(t, err, template.ErrInvalidArgument)
	})

	t.Run("rotation of", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name  string
			a, b  template.Template
			place int
			ok    bool
		}{
			{name: "dorian of ionian", a: template.Ionian(), b: template.Dorian(), place: 2, ok: true},
			{name: "ionian of dorian", a: template.Dorian(), b: template.Ionian(), place: 7, ok: true},
			{name: "self", a: template.Ionian(), b: template.Ionian(), place: 1, ok: true},
			{name: "unrelated", a: template.Ionian(), b: template.HarmonicMinor()},
			{name: "open self", a: opened("m3 M2"), b: opened("m3 M2"), place: 1, ok: true},
			{name: "open different", a: opened("m3 M2"), b: opened("M2 m3")},
			{name: "closed open", a: opened("m3 M2"), b: closed("m3 M2")},
			{name: "open closed", a: closed("m3 M2"), b: opened("m3 M2")},
			{name: "unset left", a: template.Template{}, b: template.Ionian()},
			{name: "unset right", a: template.Ionian(), b: template.Template{}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				place, ok := tc.b.RotationOf(tc.a)
				assert.Equal(t, tc.ok, ok)
				assert.Equal(t, tc.place, place)
			})
		}
	})

	t.Run("composition and repeating pattern", func(t *testing.T) {
		t.Parallel()
		tmpl := closed("M2 m2 M2 m2")
		rotated, err := tmpl.Rotate(3)
		require.NoError(t, err)
		assert.True(t, rotated.Equal(tmpl))
		place, ok := tmpl.RotationOf(tmpl)
		assert.True(t, ok)
		assert.Equal(t, 1, place)
		original := template.Ionian()
		first, err := original.Rotate(4)
		require.NoError(t, err)
		second, err := first.Rotate(3)
		require.NoError(t, err)
		direct, err := original.Rotate(6)
		require.NoError(t, err)
		assert.True(t, second.Equal(direct))
		place, ok = direct.RotationOf(original)
		assert.True(t, ok)
		assert.Equal(t, 6, place)
		place, ok = original.RotationOf(direct)
		assert.True(t, ok)
		assert.Equal(t, 3, place)
	})

	t.Run("open rotation", func(t *testing.T) {
		t.Parallel()
		tmpl := opened("m3 M2")
		got, err := tmpl.Rotate(1)
		require.NoError(t, err)
		assert.True(t, got.Equal(tmpl))
		got, err = tmpl.Rotate(2)
		require.ErrorIs(t, err, template.ErrOpenRotation)
		assert.False(t, got.IsValid())
		got, err = tmpl.Rotate(4)
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.False(t, got.IsValid())
	})

	t.Run("equality and order", func(t *testing.T) {
		t.Parallel()
		assert.True(t, template.Ionian().Equal(template.Ionian()))
		assert.False(t, closed("M2 m2").Equal(opened("M2 m2")))
		assert.False(t, template.Template{}.Equal(template.Template{}))
		ordered := []template.Template{{}, closed("P8"), closed("M2 m2"), opened("P5")}
		for i, left := range ordered {
			for j, right := range ordered {
				got := left.Compare(right)
				assert.Contains(t, []int{-1, 0, 1}, got)
				if i < j {
					assert.Equal(t, -1, got, "indexes %d and %d", i, j)
				} else if i > j {
					assert.Equal(t, 1, got, "indexes %d and %d", i, j)
				} else {
					assert.Zero(t, got)
				}
				assert.Equal(t, -got, right.Compare(left))
			}
		}
		assert.Negative(t, closed("M2 m2").Compare(closed("M2 M2")))
		assert.Negative(t, closed("M2 m2").Compare(closed("M2 M2 m2")))
	})
}

func TestMismatchingPlaces(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		a, b      template.Template
		fromFirst []int
		adjacent  []int
	}{
		{name: "dorian aeolian", a: template.Dorian(), b: template.Aeolian(), fromFirst: []int{6}, adjacent: []int{5, 6}},
		{name: "dorian mixolydian", a: template.Dorian(), b: template.Mixolydian(), fromFirst: []int{3}, adjacent: []int{2, 3}},
		{name: "blues enharmonic difference", a: template.Blues(), b: template.BluesB5(), fromFirst: []int{4}, adjacent: []int{3, 4}},
		{name: "equal", a: template.Ionian(), b: template.Ionian(), fromFirst: []int{}, adjacent: []int{}},
		{name: "cycle only", a: closed("M2 M2 m2"), b: closed("M2 M2 M2"), fromFirst: []int{4}, adjacent: []int{3}},
		{name: "open last place", a: opened("m3 M2"), b: opened("m3 m2"), fromFirst: []int{3}, adjacent: []int{2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, measure := range []template.Measure{template.FromFirst, template.Adjacent} {
				want := tc.fromFirst
				if measure == template.Adjacent {
					want = tc.adjacent
				}
				got, err := tc.a.MismatchingPlaces(tc.b, measure)
				require.NoError(t, err)
				assert.Equal(t, want, got)
				assert.NotNil(t, got)
				reverse, err := tc.b.MismatchingPlaces(tc.a, measure)
				require.NoError(t, err)
				assert.Equal(t, want, reverse)
			}
		})
	}

	t.Run("different size or kind", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct{ a, b template.Template }{
			{a: template.Dorian(), b: template.PentatonicMajor()},
			{a: closed("m3 M2"), b: opened("m3 M2")},
		} {
			got, err := tc.a.MismatchingPlaces(tc.b, template.FromFirst)
			require.ErrorIs(t, err, template.ErrIncomparable)
			assert.Nil(t, got)
		}
	})

	t.Run("validation order", func(t *testing.T) {
		t.Parallel()
		got, err := template.Template{}.MismatchingPlaces(template.Template{}, template.MeasureUnset)
		require.ErrorIs(t, err, template.ErrUnset)
		assert.Nil(t, got)
		for _, measure := range []template.Measure{template.MeasureUnset, template.Measure(255)} {
			got, err = template.Ionian().MismatchingPlaces(template.Aeolian(), measure)
			require.ErrorIs(t, err, template.ErrInvalidArgument)
			assert.Nil(t, got)
			got, err = template.Ionian().MismatchingPlaces(template.PentatonicMajor(), measure)
			require.ErrorIs(t, err, template.ErrInvalidArgument)
			assert.Nil(t, got)
		}
	})

	t.Run("exact accumulated coordinates", func(t *testing.T) {
		t.Parallel()
		left := closed("M2 M2 P4")
		right := closed("M2 M3 M3")
		leftCycle, ok := left.Cycle()
		require.True(t, ok)
		rightCycle, ok := right.Cycle()
		require.True(t, ok)
		requireInterval(t, "M6", leftCycle)
		requireInterval(t, "A6", rightCycle)
		got, err := left.MismatchingPlaces(right, template.FromFirst)
		require.NoError(t, err)
		assert.Equal(t, []int{3, 4}, got)
		got, err = left.MismatchingPlaces(right, template.Adjacent)
		require.NoError(t, err)
		assert.Equal(t, []int{2, 3}, got)
	})
}

func TestTemplateSliceOwnershipAndConcurrency(t *testing.T) {
	t.Parallel()
	t.Run("input and output slices", func(t *testing.T) {
		t.Parallel()
		input := intervals("M2 M2 m2")
		tmpl := template.MustClosedFromIntervals(input)
		input[0] = interval.PerfectOctave()
		assert.Equal(t, []string{"M2", "M2", "m2"}, intervalTexts(tmpl.AdjacentIntervals()))
		adjacent := tmpl.AdjacentIntervals()
		adjacent[1] = interval.PerfectUnison()
		assert.Equal(t, []string{"M2", "M2", "m2"}, intervalTexts(tmpl.AdjacentIntervals()))
		places := tmpl.MustPlaceIntervals()
		places[0] = interval.PerfectOctave()
		assert.Equal(t, "P1", tmpl.MustPlaceIntervals()[0].String())
	})

	t.Run("concurrent reads and independent unmarshalling", func(t *testing.T) {
		t.Parallel()
		tmpl := template.HarmonicMinor()
		const workers = 16
		var wg sync.WaitGroup
		for range workers {
			wg.Go(func() {
				for range 20 {
					assert.Equal(t, 7, tmpl.Len())
					assert.True(t, tmpl.IsOctave())
					assert.Equal(t, "M2", tmpl.MustAdjacentIntervalAfter(8).String())
					assert.Equal(t, "P8", tmpl.MustPlaceInterval(8).String())
					assert.Equal(t, "A2", tmpl.MustBetweenPlaces(6, 7).String())
					assert.True(t, tmpl.MustRotate(2).Equal(template.LocrianNatural6()))
					var decoded template.Template
					assert.NoError(t, decoded.UnmarshalText([]byte(tmpl.String())))
					assert.True(t, decoded.Equal(tmpl))
				}
			})
		}
		wg.Wait()
	})
}
