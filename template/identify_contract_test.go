package template_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentifyContractDuplicatesOrderAndSmallestPlace(t *testing.T) {
	t.Parallel()

	first := template.MustNewEntry("first", []string{"First"}, closed("M2 m2 M2 m2"))
	other := template.MustNewEntry("other", []string{"Other"}, closed("M2 M2"))
	second := template.MustNewEntry("second", []string{"Second"}, closed("M2 m2 M2 m2"))
	rotation := template.MustNewEntry("rotation", []string{"Rotation"}, closed("m2 M2 m2 M2"))
	catalog := template.MustNewCatalog(first, other, second, rotation)

	cases := []struct {
		name       string
		tmpl       template.Template
		wantCodes  []string
		wantPlaces []int
	}{
		{
			name:       "places one and three choose one",
			tmpl:       closed("M2 m2 M2 m2"),
			wantCodes:  []string{"first", "second", "rotation"},
			wantPlaces: []int{1, 1, 2},
		},
		{
			name:       "places two and four choose two",
			tmpl:       closed("m2 M2 m2 M2"),
			wantCodes:  []string{"first", "second", "rotation"},
			wantPlaces: []int{2, 2, 1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			identified, err := catalog.Identify(tc.tmpl)
			require.NoError(t, err)
			rotations, err := catalog.IdentifyRotations(tc.tmpl)
			require.NoError(t, err)

			if tc.wantPlaces[0] == 1 {
				assert.Equal(t, []string{"first", "second"}, entryCodes(identified))
			} else {
				assert.Equal(t, []string{"rotation"}, entryCodes(identified))
			}
			assert.Equal(t, tc.wantCodes, rotationCodes(rotations))
			assert.Equal(t, tc.wantPlaces, rotationPlaces(rotations))
			requireIdentifyMatchesRelated(t, catalog, tc.tmpl)
		})
	}
}

func TestIdentifyContractOpenKindAndLength(t *testing.T) {
	t.Parallel()

	openEntry := template.MustNewEntry("open", []string{"Open"}, opened("m3 M2"))
	closedEntry := template.MustNewEntry("closed", []string{"Closed"}, closed("m3 M2"))
	shortEntry := template.MustNewEntry("short", []string{"Short"}, closed("M2"))
	catalog := template.MustNewCatalog(openEntry, closedEntry, shortEntry)

	cases := []struct {
		name       string
		tmpl       template.Template
		wantCodes  []string
		wantPlaces []int
	}{
		{name: "open matches only equal open", tmpl: opened("m3 M2"), wantCodes: []string{"open"}, wantPlaces: []int{1}},
		{name: "closed matches only equal closed", tmpl: closed("m3 M2"), wantCodes: []string{"closed"}, wantPlaces: []int{1}},
		{name: "different length is a miss", tmpl: opened("M2"), wantCodes: []string{}, wantPlaces: []int{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			identified, err := catalog.Identify(tc.tmpl)
			require.NoError(t, err)
			require.NotNil(t, identified)
			rotations, err := catalog.IdentifyRotations(tc.tmpl)
			require.NoError(t, err)
			require.NotNil(t, rotations)

			assert.Equal(t, tc.wantCodes, entryCodes(identified))
			assert.Equal(t, tc.wantCodes, rotationCodes(rotations))
			assert.Equal(t, tc.wantPlaces, rotationPlaces(rotations))
			requireIdentifyMatchesRelated(t, catalog, tc.tmpl)
		})
	}
}

func TestIdentifyContractEmptyCatalogAndMiss(t *testing.T) {
	t.Parallel()

	empty, err := template.NewCatalog()
	require.NoError(t, err)
	cases := []struct {
		name    string
		catalog template.Catalog
		tmpl    template.Template
	}{
		{name: "zero catalog", catalog: template.Catalog{}, tmpl: template.Dorian()},
		{name: "empty catalog", catalog: empty, tmpl: template.Dorian()},
		{name: "built-in miss", catalog: template.BuiltIn(), tmpl: template.MustParse("M2 M2 M2 M2 M2 M2 M2")},
		{name: "custom miss", catalog: template.MustNewCatalog(template.IonianEntry()), tmpl: closed("M2 M2 m6")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			identified, err := tc.catalog.Identify(tc.tmpl)
			require.NoError(t, err)
			require.NotNil(t, identified)
			assert.Empty(t, identified)

			rotations, err := tc.catalog.IdentifyRotations(tc.tmpl)
			require.NoError(t, err)
			require.NotNil(t, rotations)
			assert.Empty(t, rotations)
			requireIdentifyMatchesRelated(t, tc.catalog, tc.tmpl)
		})
	}
}

func TestIdentifyContractUnsetAndMustPanics(t *testing.T) {
	t.Parallel()

	catalog := template.MustNewCatalog(template.IonianEntry())
	t.Run("identify", func(t *testing.T) {
		t.Parallel()
		got, err := catalog.Identify(template.Template{})
		require.ErrorIs(t, err, template.ErrUnset)
		assert.Nil(t, got)
		requirePanicValue(t, err, func() { catalog.MustIdentify(template.Template{}) })
	})
	t.Run("identify rotations", func(t *testing.T) {
		t.Parallel()
		got, err := catalog.IdentifyRotations(template.Template{})
		require.ErrorIs(t, err, template.ErrUnset)
		assert.Nil(t, got)
		requirePanicValue(t, err, func() { catalog.MustIdentifyRotations(template.Template{}) })
	})
}

func TestIdentifyContractCustomLengths(t *testing.T) {
	t.Parallel()

	for _, size := range []int{16, 1024} {
		t.Run(stringSize(size), func(t *testing.T) {
			t.Parallel()

			hit := repeatedMajorSeconds(size, false)
			entry := template.MustNewEntry("custom", []string{"Custom"}, hit)
			catalog := template.MustNewCatalog(entry)
			miss := repeatedMajorSeconds(size, true)

			identified, err := catalog.Identify(hit)
			require.NoError(t, err)
			assert.Equal(t, []string{"custom"}, entryCodes(identified))
			rotations, err := catalog.IdentifyRotations(hit)
			require.NoError(t, err)
			assert.Equal(t, []string{"custom"}, rotationCodes(rotations))
			assert.Equal(t, []int{1}, rotationPlaces(rotations))
			requireIdentifyMatchesRelated(t, catalog, hit)

			identified, err = catalog.Identify(miss)
			require.NoError(t, err)
			require.NotNil(t, identified)
			assert.Empty(t, identified)
			rotations, err = catalog.IdentifyRotations(miss)
			require.NoError(t, err)
			require.NotNil(t, rotations)
			assert.Empty(t, rotations)
			requireIdentifyMatchesRelated(t, catalog, miss)
		})
	}
}

func TestIdentifyContractOwnsResultSlices(t *testing.T) {
	t.Parallel()

	first := template.MustNewEntry("first", []string{"First"}, closed("M2 m2 M2 m2"))
	second := template.MustNewEntry("second", []string{"Second"}, closed("M2 m2 M2 m2"))
	catalog := template.MustNewCatalog(first, second)

	identified, err := catalog.Identify(first.Template())
	require.NoError(t, err)
	require.Len(t, identified, 2)
	identified[0] = template.Entry{}
	identifiedAgain, err := catalog.Identify(first.Template())
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second"}, entryCodes(identifiedAgain))
	assert.Equal(t, []string{"first", "second"}, catalog.Codes())

	rotations, err := catalog.IdentifyRotations(first.Template())
	require.NoError(t, err)
	require.Len(t, rotations, 2)
	rotations[0] = template.Rotation{}
	rotationsAgain, err := catalog.IdentifyRotations(first.Template())
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second"}, rotationCodes(rotationsAgain))
	assert.Equal(t, []int{1, 1}, rotationPlaces(rotationsAgain))
}

func requireIdentifyMatchesRelated(t testing.TB, catalog template.Catalog, tmpl template.Template) {
	t.Helper()

	identified, err := catalog.Identify(tmpl)
	require.NoError(t, err)
	related, err := catalog.Related(tmpl, 0, template.Adjacent, false)
	require.NoError(t, err)
	assert.Equal(t, identified, relationEntries(related))

	rotations, err := catalog.IdentifyRotations(tmpl)
	require.NoError(t, err)
	relatedRotations, err := catalog.Related(tmpl, 0, template.Adjacent, true)
	require.NoError(t, err)
	assert.Equal(t, rotations, relationRotations(relatedRotations))
}

func repeatedMajorSeconds(size int, firstMinor bool) template.Template {
	values := slices.Repeat([]interval.Interval{interval.MajorSecond()}, size)
	if firstMinor {
		values[0] = interval.MinorSecond()
	}
	return template.MustClosedFromIntervals(values)
}

func stringSize(size int) string {
	return strconv.Itoa(size)
}
