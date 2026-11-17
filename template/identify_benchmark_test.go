package template_test

import (
	"testing"

	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	benchmarkEntriesSink   []template.Entry
	benchmarkRotationsSink []template.Rotation
	benchmarkErrorSink     error
)

type identifyBenchmarkScenario struct {
	name               string
	catalog            template.Catalog
	tmpl               template.Template
	wantIdentifyCodes  []string
	wantRotationCodes  []string
	wantRotationPlaces []int
}

func identifyBenchmarkScenarios() []identifyBenchmarkScenario {
	repeatedEntry := template.MustNewEntry("repeated", []string{"Repeated"}, closed("M2 m2 M2 m2"))
	repeatedCatalog := template.MustNewCatalog(repeatedEntry)
	customCatalog := func(size int) template.Catalog {
		return template.MustNewCatalog(template.MustNewEntry("custom", []string{"Custom"}, repeatedMajorSeconds(size, false)))
	}

	return []identifyBenchmarkScenario{
		{
			name:               "builtin_hit",
			catalog:            template.BuiltIn(),
			tmpl:               template.Dorian(),
			wantIdentifyCodes:  []string{"dorian"},
			wantRotationCodes:  []string{"ionian", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian"},
			wantRotationPlaces: []int{2, 1, 7, 6, 5, 4, 3},
		},
		{
			name:               "builtin_miss",
			catalog:            template.BuiltIn(),
			tmpl:               template.MustParse("M2 M2 M2 M2 M2 M2 M2"),
			wantIdentifyCodes:  []string{},
			wantRotationCodes:  []string{},
			wantRotationPlaces: []int{},
		},
		{
			name:               "repeated_pattern/place_1",
			catalog:            repeatedCatalog,
			tmpl:               closed("M2 m2 M2 m2"),
			wantIdentifyCodes:  []string{"repeated"},
			wantRotationCodes:  []string{"repeated"},
			wantRotationPlaces: []int{1},
		},
		{
			name:               "repeated_pattern/place_2",
			catalog:            repeatedCatalog,
			tmpl:               closed("m2 M2 m2 M2"),
			wantIdentifyCodes:  []string{},
			wantRotationCodes:  []string{"repeated"},
			wantRotationPlaces: []int{2},
		},
		{
			name:               "custom_16_hit",
			catalog:            customCatalog(16),
			tmpl:               repeatedMajorSeconds(16, false),
			wantIdentifyCodes:  []string{"custom"},
			wantRotationCodes:  []string{"custom"},
			wantRotationPlaces: []int{1},
		},
		{
			name:               "custom_16_early_miss",
			catalog:            customCatalog(16),
			tmpl:               repeatedMajorSeconds(16, true),
			wantIdentifyCodes:  []string{},
			wantRotationCodes:  []string{},
			wantRotationPlaces: []int{},
		},
		{
			name:               "custom_1024_hit",
			catalog:            customCatalog(1024),
			tmpl:               repeatedMajorSeconds(1024, false),
			wantIdentifyCodes:  []string{"custom"},
			wantRotationCodes:  []string{"custom"},
			wantRotationPlaces: []int{1},
		},
		{
			name:               "custom_1024_early_miss",
			catalog:            customCatalog(1024),
			tmpl:               repeatedMajorSeconds(1024, true),
			wantIdentifyCodes:  []string{},
			wantRotationCodes:  []string{},
			wantRotationPlaces: []int{},
		},
	}
}

func BenchmarkCatalogIdentify(b *testing.B) {
	for _, scenario := range identifyBenchmarkScenarios() {
		b.Run(scenario.name, func(b *testing.B) {
			checkIdentifyBenchmarkScenario(b, scenario)
			b.ResetTimer()
			for b.Loop() {
				benchmarkEntriesSink, benchmarkErrorSink = scenario.catalog.Identify(scenario.tmpl)
			}
		})
	}
}

func BenchmarkCatalogIdentifyRotations(b *testing.B) {
	for _, scenario := range identifyBenchmarkScenarios() {
		b.Run(scenario.name, func(b *testing.B) {
			checkIdentifyBenchmarkScenario(b, scenario)
			b.ResetTimer()
			for b.Loop() {
				benchmarkRotationsSink, benchmarkErrorSink = scenario.catalog.IdentifyRotations(scenario.tmpl)
			}
		})
	}
}

func checkIdentifyBenchmarkScenario(b *testing.B, scenario identifyBenchmarkScenario) {
	b.Helper()

	identified, err := scenario.catalog.Identify(scenario.tmpl)
	require.NoError(b, err)
	require.NotNil(b, identified)
	assert.Equal(b, scenario.wantIdentifyCodes, entryCodes(identified))

	rotations, err := scenario.catalog.IdentifyRotations(scenario.tmpl)
	require.NoError(b, err)
	require.NotNil(b, rotations)
	assert.Equal(b, scenario.wantRotationCodes, rotationCodes(rotations))
	assert.Equal(b, scenario.wantRotationPlaces, rotationPlaces(rotations))
}
