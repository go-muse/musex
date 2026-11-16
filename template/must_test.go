package template_test

import (
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mustResult struct {
	value any
	err   error
}

type mustCase struct {
	name string
	call func() mustResult
	must func() any
}

type russianResult struct {
	text string
	ok   bool
}

func mustSuccessCases() []mustCase {
	entry := template.MustNewEntry("entry", []string{"Entry"}, template.Ionian())
	other := template.MustNewEntry("other", []string{"Other"}, template.Dorian())
	catalog := template.MustNewCatalog(entry)
	return []mustCase{
		{name: "ClosedFromIntervals", call: func() mustResult {
			v, err := template.ClosedFromIntervals(intervals("M2 m2"))
			return mustResult{v, err}
		}, must: func() any { return template.MustClosedFromIntervals(intervals("M2 m2")) }},
		{name: "OpenFromIntervals", call: func() mustResult { v, err := template.OpenFromIntervals(intervals("m3 M2")); return mustResult{v, err} }, must: func() any { return template.MustOpenFromIntervals(intervals("m3 M2")) }},
		{name: "ClosedFromPlaceIntervals", call: func() mustResult {
			v, err := template.ClosedFromPlaceIntervals(intervals("P1 M2 M3"), interval.PerfectFourth())
			return mustResult{v, err}
		}, must: func() any {
			return template.MustClosedFromPlaceIntervals(intervals("P1 M2 M3"), interval.PerfectFourth())
		}},
		{name: "OpenFromPlaceIntervals", call: func() mustResult {
			v, err := template.OpenFromPlaceIntervals(intervals("P1 m3 P4"))
			return mustResult{v, err}
		}, must: func() any { return template.MustOpenFromPlaceIntervals(intervals("P1 m3 P4")) }},
		{name: "ClosedFromSpellings", call: func() mustResult {
			v, err := template.ClosedFromSpellings(names("G A B"), interval.PerfectFourth())
			return mustResult{v, err}
		}, must: func() any { return template.MustClosedFromSpellings(names("G A B"), interval.PerfectFourth()) }},
		{name: "FromSpellingsInOctave", call: func() mustResult { v, err := template.FromSpellingsInOctave(names("C D E")); return mustResult{v, err} }, must: func() any { return template.MustFromSpellingsInOctave(names("C D E")) }},
		{name: "OpenFromSpellings", call: func() mustResult { v, err := template.OpenFromSpellings(names("E G A")); return mustResult{v, err} }, must: func() any { return template.MustOpenFromSpellings(names("E G A")) }},
		{name: "AdjacentIntervalAfter", call: func() mustResult { v, err := template.Ionian().AdjacentIntervalAfter(8); return mustResult{v, err} }, must: func() any { return template.Ionian().MustAdjacentIntervalAfter(8) }},
		{name: "PlaceIntervals", call: func() mustResult { v, err := template.Ionian().PlaceIntervals(); return mustResult{v, err} }, must: func() any { return template.Ionian().MustPlaceIntervals() }},
		{name: "PlaceInterval", call: func() mustResult { v, err := template.Ionian().PlaceInterval(9); return mustResult{v, err} }, must: func() any { return template.Ionian().MustPlaceInterval(9) }},
		{name: "BetweenPlaces", call: func() mustResult { v, err := template.Ionian().BetweenPlaces(7, 4); return mustResult{v, err} }, must: func() any { return template.Ionian().MustBetweenPlaces(7, 4) }},
		{name: "Rotate", call: func() mustResult { v, err := template.Ionian().Rotate(2); return mustResult{v, err} }, must: func() any { return template.Ionian().MustRotate(2) }},
		{name: "MismatchingPlaces", call: func() mustResult {
			v, err := template.Dorian().MismatchingPlaces(template.Aeolian(), template.FromFirst)
			return mustResult{v, err}
		}, must: func() any { return template.Dorian().MustMismatchingPlaces(template.Aeolian(), template.FromFirst) }},
		{name: "Parse", call: func() mustResult { v, err := template.Parse("m3 M2 |"); return mustResult{v, err} }, must: func() any { return template.MustParse("m3 M2 |") }},
		{name: "Russian", call: func() mustResult {
			v, ok, err := template.Blues().Russian()
			return mustResult{russianResult{v, ok}, err}
		}, must: func() any { v, ok := template.Blues().MustRussian(); return russianResult{v, ok} }},
		{name: "NewEntry", call: func() mustResult {
			v, err := template.NewEntry("new", []string{"New"}, template.Ionian())
			return mustResult{v, err}
		}, must: func() any { return template.MustNewEntry("new", []string{"New"}, template.Ionian()) }},
		{name: "NewCatalog", call: func() mustResult { v, err := template.NewCatalog(entry); return mustResult{v, err} }, must: func() any { return template.MustNewCatalog(entry) }},
		{name: "With", call: func() mustResult { v, err := catalog.With(other); return mustResult{v, err} }, must: func() any { return catalog.MustWith(other) }},
		{name: "Without", call: func() mustResult { v, err := catalog.Without("entry"); return mustResult{v, err} }, must: func() any { return catalog.MustWithout("entry") }},
		{name: "Merge", call: func() mustResult { v, err := catalog.Merge(template.MustNewCatalog(other)); return mustResult{v, err} }, must: func() any { return catalog.MustMerge(template.MustNewCatalog(other)) }},
		{name: "Filter", call: func() mustResult {
			v, err := catalog.Filter(func(template.Entry) bool { return true })
			return mustResult{v, err}
		}, must: func() any { return catalog.MustFilter(func(template.Entry) bool { return true }) }},
		{name: "Identify", call: func() mustResult { v, err := catalog.Identify(template.Ionian()); return mustResult{v, err} }, must: func() any { return catalog.MustIdentify(template.Ionian()) }},
		{name: "IdentifyRotations", call: func() mustResult { v, err := catalog.IdentifyRotations(template.Dorian()); return mustResult{v, err} }, must: func() any { return catalog.MustIdentifyRotations(template.Dorian()) }},
		{name: "Related", call: func() mustResult {
			v, err := catalog.Related(template.Ionian(), 0, template.FromFirst, true)
			return mustResult{v, err}
		}, must: func() any { return catalog.MustRelated(template.Ionian(), 0, template.FromFirst, true) }},
	}
}

func mustErrorCases() []mustCase {
	entry := template.MustNewEntry("entry", []string{"Entry"}, template.Ionian())
	catalog := template.MustNewCatalog(entry)
	return []mustCase{
		{name: "ClosedFromIntervals", call: func() mustResult { v, err := template.ClosedFromIntervals(nil); return mustResult{v, err} }, must: func() any { return template.MustClosedFromIntervals(nil) }},
		{name: "OpenFromIntervals", call: func() mustResult { v, err := template.OpenFromIntervals(nil); return mustResult{v, err} }, must: func() any { return template.MustOpenFromIntervals(nil) }},
		{name: "ClosedFromPlaceIntervals", call: func() mustResult {
			v, err := template.ClosedFromPlaceIntervals(intervals("P1"), interval.Interval{})
			return mustResult{v, err}
		}, must: func() any { return template.MustClosedFromPlaceIntervals(intervals("P1"), interval.Interval{}) }},
		{name: "OpenFromPlaceIntervals", call: func() mustResult {
			v, err := template.OpenFromPlaceIntervals(intervals("P1"))
			return mustResult{v, err}
		}, must: func() any { return template.MustOpenFromPlaceIntervals(intervals("P1")) }},
		{name: "ClosedFromSpellings", call: func() mustResult {
			v, err := template.ClosedFromSpellings(names("C C"), interval.PerfectOctave())
			return mustResult{v, err}
		}, must: func() any { return template.MustClosedFromSpellings(names("C C"), interval.PerfectOctave()) }},
		{name: "FromSpellingsInOctave", call: func() mustResult { v, err := template.FromSpellingsInOctave(names("C C")); return mustResult{v, err} }, must: func() any { return template.MustFromSpellingsInOctave(names("C C")) }},
		{name: "OpenFromSpellings", call: func() mustResult {
			v, err := template.OpenFromSpellings([]spelling.Spelling{spelling.C()})
			return mustResult{v, err}
		}, must: func() any { return template.MustOpenFromSpellings([]spelling.Spelling{spelling.C()}) }},
		{name: "AdjacentIntervalAfter", call: func() mustResult { v, err := template.Ionian().AdjacentIntervalAfter(0); return mustResult{v, err} }, must: func() any { return template.Ionian().MustAdjacentIntervalAfter(0) }},
		{name: "PlaceIntervals", call: func() mustResult { v, err := template.Template{}.PlaceIntervals(); return mustResult{v, err} }, must: func() any { return template.Template{}.MustPlaceIntervals() }},
		{name: "PlaceInterval", call: func() mustResult { v, err := template.Ionian().PlaceInterval(0); return mustResult{v, err} }, must: func() any { return template.Ionian().MustPlaceInterval(0) }},
		{name: "BetweenPlaces", call: func() mustResult { v, err := template.Ionian().BetweenPlaces(0, 1); return mustResult{v, err} }, must: func() any { return template.Ionian().MustBetweenPlaces(0, 1) }},
		{name: "Rotate", call: func() mustResult {
			v, err := template.MustOpenFromIntervals(intervals("m3 M2")).Rotate(2)
			return mustResult{v, err}
		}, must: func() any { return template.MustOpenFromIntervals(intervals("m3 M2")).MustRotate(2) }},
		{name: "MismatchingPlaces", call: func() mustResult {
			v, err := template.Ionian().MismatchingPlaces(template.Aeolian(), template.MeasureUnset)
			return mustResult{v, err}
		}, must: func() any { return template.Ionian().MustMismatchingPlaces(template.Aeolian(), template.MeasureUnset) }},
		{name: "Parse", call: func() mustResult { v, err := template.Parse("M2  M2"); return mustResult{v, err} }, must: func() any { return template.MustParse("M2  M2") }},
		{name: "Russian", call: func() mustResult {
			v, ok, err := template.Template{}.Russian()
			return mustResult{russianResult{v, ok}, err}
		}, must: func() any { v, ok := template.Template{}.MustRussian(); return russianResult{v, ok} }},
		{name: "NewEntry", call: func() mustResult {
			v, err := template.NewEntry("", []string{"Name"}, template.Ionian())
			return mustResult{v, err}
		}, must: func() any { return template.MustNewEntry("", []string{"Name"}, template.Ionian()) }},
		{name: "NewCatalog", call: func() mustResult { v, err := template.NewCatalog(entry, entry); return mustResult{v, err} }, must: func() any { return template.MustNewCatalog(entry, entry) }},
		{name: "With", call: func() mustResult { v, err := catalog.With(entry); return mustResult{v, err} }, must: func() any { return catalog.MustWith(entry) }},
		{name: "Without", call: func() mustResult { v, err := catalog.Without("missing"); return mustResult{v, err} }, must: func() any { return catalog.MustWithout("missing") }},
		{name: "Merge", call: func() mustResult { v, err := catalog.Merge(catalog); return mustResult{v, err} }, must: func() any { return catalog.MustMerge(catalog) }},
		{name: "Filter", call: func() mustResult { v, err := catalog.Filter(nil); return mustResult{v, err} }, must: func() any { return catalog.MustFilter(nil) }},
		{name: "Identify", call: func() mustResult { v, err := catalog.Identify(template.Template{}); return mustResult{v, err} }, must: func() any { return catalog.MustIdentify(template.Template{}) }},
		{name: "IdentifyRotations", call: func() mustResult { v, err := catalog.IdentifyRotations(template.Template{}); return mustResult{v, err} }, must: func() any { return catalog.MustIdentifyRotations(template.Template{}) }},
		{name: "Related", call: func() mustResult {
			v, err := catalog.Related(template.Ionian(), -1, template.FromFirst, false)
			return mustResult{v, err}
		}, must: func() any { return catalog.MustRelated(template.Ionian(), -1, template.FromFirst, false) }},
	}
}

func TestMustSuccessMatchesBaseOperation(t *testing.T) {
	t.Parallel()
	for _, tc := range mustSuccessCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.call()
			require.NoError(t, got.err)
			assert.Equal(t, got.value, tc.must())
		})
	}
}

func TestMustPanicsWithBaseError(t *testing.T) {
	t.Parallel()
	for _, tc := range mustErrorCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.call()
			require.Error(t, got.err)
			requirePanicValue(t, got.err, func() { _ = tc.must() })
		})
	}
}
