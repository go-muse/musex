package template_test

import (
	"testing"

	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentify(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		tmpl  template.Template
		codes []string
	}{
		{name: "aeolian", tmpl: template.Aeolian(), codes: []string{"aeolian"}},
		{name: "duplicate template", tmpl: template.HirajoshiSachs(), codes: []string{"hirajoshi-sachs", "iwato"}},
		{name: "not present", tmpl: closed("M2 M2 m6"), codes: []string{}},
		{name: "open absent", tmpl: opened("m3 M2"), codes: []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.BuiltIn().Identify(tc.tmpl)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.codes, entryCodes(got))
		})
	}
	got, err := template.BuiltIn().Identify(template.Template{})
	require.ErrorIs(t, err, template.ErrUnset)
	assert.Nil(t, got)
}

func TestIdentifyRotations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		tmpl   template.Template
		codes  []string
		places []int
	}{
		{
			name: "dorian flat two", tmpl: template.DorianB2(),
			codes:  []string{"melodic-major", "melodic-minor", "acoustic", "dorian-b2", "lydian-augmented", "locrian-sharp2", "altered"},
			places: []int{5, 2, 6, 1, 7, 4, 3},
		},
		{
			name: "minor pentatonic", tmpl: template.PentatonicMinor(),
			codes:  []string{"pentatonic-major", "pentatonic-minor", "pentatonic-suspended", "pentatonic-blues-minor", "yo"},
			places: []int{5, 1, 4, 3, 2},
		},
		{
			name: "Messiaen five commons", tmpl: template.Messiaen5Commons(),
			codes:  []string{"messiaen-5-commons", "two-semitone-tritone"},
			places: []int{1, 5},
		},
		{name: "open absent", tmpl: opened("m3 M2"), codes: []string{}, places: []int{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.BuiltIn().IdentifyRotations(tc.tmpl)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.codes, rotationCodes(got))
			assert.Equal(t, tc.places, rotationPlaces(got))
		})
	}
	got, err := template.BuiltIn().IdentifyRotations(template.Template{})
	require.ErrorIs(t, err, template.ErrUnset)
	assert.Nil(t, got)
}

func TestIdentifyInCustomCatalog(t *testing.T) {
	t.Parallel()
	first := template.MustNewEntry("first", []string{"First"}, closed("M2 m2"))
	duplicate := template.MustNewEntry("duplicate", []string{"Duplicate"}, closed("M2 m2"))
	rotation := template.MustNewEntry("rotation", []string{"Rotation"}, closed("m2 M2"))
	open := template.MustNewEntry("open", []string{"Open"}, opened("m3 M2"))
	catalog := template.MustNewCatalog(first, duplicate, rotation, open)
	got, err := catalog.Identify(closed("M2 m2"))
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "duplicate"}, entryCodes(got))
	rotations, err := catalog.IdentifyRotations(closed("M2 m2"))
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "duplicate", "rotation"}, rotationCodes(rotations))
	assert.Equal(t, []int{1, 1, 2}, rotationPlaces(rotations))
	got, err = catalog.Identify(opened("m3 M2"))
	require.NoError(t, err)
	assert.Equal(t, []string{"open"}, entryCodes(got))
	rotations, err = catalog.IdentifyRotations(opened("m3 M2"))
	require.NoError(t, err)
	assert.Equal(t, []string{"open"}, rotationCodes(rotations))
	assert.Equal(t, []int{1}, rotationPlaces(rotations))
}

type relationExpectation struct {
	code   string
	place  int
	places []int
}

func relationSummaries(relations []template.Relation) []relationExpectation {
	result := make([]relationExpectation, len(relations))
	for i, relation := range relations {
		result[i] = relationExpectation{code: relation.Entry.Code(), place: relation.Place, places: relation.Places}
	}
	return result
}

func TestRelatedWithoutRotations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		tmpl    template.Template
		limit   int
		measure template.Measure
		want    []relationExpectation
	}{
		{
			name: "dorian one from first", tmpl: template.Dorian(), limit: 1, measure: template.FromFirst,
			want: []relationExpectation{
				{code: "dorian", place: 1, places: []int{}},
				{code: "mixolydian", place: 1, places: []int{3}},
				{code: "aeolian", place: 1, places: []int{6}},
				{code: "melodic-minor", place: 1, places: []int{7}},
				{code: "dorian-b2", place: 1, places: []int{2}},
				{code: "ukrainian-dorian", place: 1, places: []int{4}},
				{code: "dorian-b5", place: 1, places: []int{5}},
			},
		},
		{
			name: "dorian two from first", tmpl: template.Dorian(), limit: 2, measure: template.FromFirst,
			want: []relationExpectation{
				{code: "dorian", place: 1, places: []int{}},
				{code: "mixolydian", place: 1, places: []int{3}},
				{code: "aeolian", place: 1, places: []int{6}},
				{code: "melodic-minor", place: 1, places: []int{7}},
				{code: "dorian-b2", place: 1, places: []int{2}},
				{code: "ukrainian-dorian", place: 1, places: []int{4}},
				{code: "dorian-b5", place: 1, places: []int{5}},
				{code: "ionian", place: 1, places: []int{3, 7}},
				{code: "phrygian", place: 1, places: []int{2, 6}},
				{code: "melodic-major", place: 1, places: []int{3, 6}},
				{code: "harmonic-minor", place: 1, places: []int{6, 7}},
				{code: "acoustic", place: 1, places: []int{3, 4}},
				{code: "locrian-sharp2", place: 1, places: []int{5, 6}},
				{code: "locrian-natural6", place: 1, places: []int{2, 5}},
				{code: "lydian-b3", place: 1, places: []int{4, 7}},
				{code: "mixolydian-b2", place: 1, places: []int{2, 3}},
				{code: "aeolian-sharp4", place: 1, places: []int{4, 6}},
				{code: "neapolitan-major", place: 1, places: []int{2, 7}},
			},
		},
		{
			name: "dorian one adjacent", tmpl: template.Dorian(), limit: 1, measure: template.Adjacent,
			want: []relationExpectation{{code: "dorian", place: 1, places: []int{}}},
		},
		{
			name: "dorian two adjacent", tmpl: template.Dorian(), limit: 2, measure: template.Adjacent,
			want: []relationExpectation{
				{code: "dorian", place: 1, places: []int{}},
				{code: "mixolydian", place: 1, places: []int{2, 3}},
				{code: "aeolian", place: 1, places: []int{5, 6}},
				{code: "melodic-minor", place: 1, places: []int{6, 7}},
				{code: "acoustic", place: 1, places: []int{2, 4}},
				{code: "dorian-b2", place: 1, places: []int{1, 2}},
				{code: "locrian-sharp2", place: 1, places: []int{4, 6}},
				{code: "ukrainian-dorian", place: 1, places: []int{3, 4}},
				{code: "dorian-b5", place: 1, places: []int{4, 5}},
				{code: "hungarian-major", place: 1, places: []int{1, 4}},
			},
		},
		{
			name: "blues one", tmpl: template.Blues(), limit: 1, measure: template.FromFirst,
			want: []relationExpectation{{code: "blues", place: 1, places: []int{}}, {code: "blues-b5", place: 1, places: []int{4}}},
		},
		{
			name: "pentatonic one", tmpl: template.PentatonicMajor(), limit: 1, measure: template.FromFirst,
			want: []relationExpectation{{code: "pentatonic-major", place: 1, places: []int{}}, {code: "yo", place: 1, places: []int{3}}, {code: "kumoi", place: 1, places: []int{3}}},
		},
		{name: "no equal size", tmpl: closed("M2 M2 m6"), limit: 1, measure: template.FromFirst, want: []relationExpectation{}},
		{name: "no open built in", tmpl: opened("m3 M2"), limit: 1, measure: template.Adjacent, want: []relationExpectation{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.BuiltIn().Related(tc.tmpl, tc.limit, tc.measure, false)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.want, relationSummaries(got))
			for _, relation := range got {
				assert.Equal(t, len(relation.Places), relation.Count())
				assert.NotNil(t, relation.Places)
			}
		})
	}
}

func TestRelatedWithRotations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		tmpl  template.Template
		limit int
		want  []relationExpectation
	}{
		{
			name: "ionian zero", tmpl: template.Ionian(), limit: 0,
			want: []relationExpectation{
				{code: "ionian", place: 1, places: []int{}}, {code: "dorian", place: 7, places: []int{}},
				{code: "phrygian", place: 6, places: []int{}}, {code: "lydian", place: 5, places: []int{}},
				{code: "mixolydian", place: 4, places: []int{}}, {code: "aeolian", place: 3, places: []int{}},
				{code: "locrian", place: 2, places: []int{}},
			},
		},
		{
			name: "ionian one", tmpl: template.Ionian(), limit: 1,
			want: []relationExpectation{
				{code: "ionian", place: 1, places: []int{}}, {code: "dorian", place: 7, places: []int{}},
				{code: "phrygian", place: 6, places: []int{}}, {code: "lydian", place: 5, places: []int{}},
				{code: "mixolydian", place: 4, places: []int{}}, {code: "aeolian", place: 3, places: []int{}},
				{code: "locrian", place: 2, places: []int{}},
				{code: "harmonic-major", place: 1, places: []int{6}}, {code: "melodic-major", place: 4, places: []int{3}},
				{code: "harmonic-minor", place: 3, places: []int{5}}, {code: "melodic-minor", place: 1, places: []int{3}},
				{code: "acoustic", place: 5, places: []int{3}}, {code: "dorian-b2", place: 7, places: []int{3}},
				{code: "lydian-augmented", place: 6, places: []int{3}}, {code: "locrian-sharp2", place: 3, places: []int{3}},
				{code: "altered", place: 2, places: []int{3}}, {code: "locrian-natural6", place: 2, places: []int{5}},
				{code: "ionian-sharp5", place: 1, places: []int{5}}, {code: "ukrainian-dorian", place: 7, places: []int{5}},
				{code: "phrygian-dominant", place: 6, places: []int{5}}, {code: "lydian-sharp2", place: 5, places: []int{5}},
				{code: "ultralocrian", place: 4, places: []int{5}}, {code: "dorian-b5", place: 7, places: []int{6}},
				{code: "phrygian-b4", place: 6, places: []int{6}}, {code: "lydian-b3", place: 5, places: []int{6}},
				{code: "mixolydian-b2", place: 4, places: []int{6}}, {code: "lydian-sharp2-sharp5", place: 3, places: []int{6}},
				{code: "locrian-bb7", place: 2, places: []int{6}}, {code: "aeolian-sharp4", place: 3, places: []int{2}},
				{code: "neapolitan-minor", place: 6, places: []int{2}},
			},
		},
		{
			name: "dorian one with tie breaking", tmpl: template.Dorian(), limit: 1,
			want: []relationExpectation{
				{code: "ionian", place: 2, places: []int{}}, {code: "dorian", place: 1, places: []int{}},
				{code: "phrygian", place: 7, places: []int{}}, {code: "lydian", place: 6, places: []int{}},
				{code: "mixolydian", place: 5, places: []int{}}, {code: "aeolian", place: 4, places: []int{}},
				{code: "locrian", place: 3, places: []int{}},
				{code: "harmonic-major", place: 2, places: []int{5}}, {code: "melodic-major", place: 4, places: []int{7}},
				{code: "harmonic-minor", place: 4, places: []int{4}}, {code: "melodic-minor", place: 1, places: []int{7}},
				{code: "acoustic", place: 5, places: []int{7}}, {code: "dorian-b2", place: 1, places: []int{2}},
				{code: "lydian-augmented", place: 6, places: []int{7}}, {code: "locrian-sharp2", place: 3, places: []int{7}},
				{code: "altered", place: 2, places: []int{7}}, {code: "locrian-natural6", place: 3, places: []int{4}},
				{code: "ionian-sharp5", place: 2, places: []int{4}}, {code: "ukrainian-dorian", place: 1, places: []int{4}},
				{code: "phrygian-dominant", place: 7, places: []int{4}}, {code: "lydian-sharp2", place: 6, places: []int{4}},
				{code: "ultralocrian", place: 5, places: []int{4}}, {code: "dorian-b5", place: 1, places: []int{5}},
				{code: "phrygian-b4", place: 7, places: []int{5}}, {code: "lydian-b3", place: 6, places: []int{5}},
				{code: "mixolydian-b2", place: 5, places: []int{5}}, {code: "lydian-sharp2-sharp5", place: 4, places: []int{5}},
				{code: "locrian-bb7", place: 3, places: []int{5}},
			},
		},
		{name: "different size", tmpl: closed("M2 M2 m6"), limit: 1, want: []relationExpectation{}},
		{name: "open absent", tmpl: opened("m3 M2"), limit: 1, want: []relationExpectation{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := template.BuiltIn().Related(tc.tmpl, tc.limit, template.FromFirst, true)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.want, relationSummaries(got))
		})
	}
}

func TestRelatedCustomCatalogAndErrors(t *testing.T) {
	t.Parallel()
	open := template.MustNewEntry("open", []string{"Open"}, opened("m3 M2"))
	otherOpen := template.MustNewEntry("other-open", []string{"Other open"}, opened("m3 m2"))
	differentSize := template.MustNewEntry("short", []string{"Short"}, closed("M2"))
	differentKind := template.MustNewEntry("closed", []string{"Closed"}, closed("m3 M2"))
	catalog := template.MustNewCatalog(open, otherOpen, differentSize, differentKind)
	got, err := catalog.Related(opened("m3 M2"), 1, template.FromFirst, true)
	require.NoError(t, err)
	assert.Equal(t, []relationExpectation{
		{code: "open", place: 1, places: []int{}},
		{code: "other-open", place: 1, places: []int{3}},
	}, relationSummaries(got))

	got, err = catalog.Related(opened("m3 M2"), 0, template.Adjacent, false)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 0, got[0].Count())
	got[0].Places = append(got[0].Places, 99)
	again, err := catalog.Related(opened("m3 M2"), 0, template.Adjacent, false)
	require.NoError(t, err)
	assert.Empty(t, again[0].Places)

	got, err = catalog.Related(template.Template{}, -1, template.MeasureUnset, true)
	require.ErrorIs(t, err, template.ErrUnset)
	assert.Nil(t, got)
	for _, tc := range []struct {
		limit   int
		measure template.Measure
	}{
		{limit: -1, measure: template.FromFirst},
		{limit: 0, measure: template.MeasureUnset},
		{limit: 0, measure: template.Measure(255)},
	} {
		got, err = catalog.Related(opened("m3 M2"), tc.limit, tc.measure, false)
		require.ErrorIs(t, err, template.ErrInvalidArgument)
		assert.Nil(t, got)
	}
}
