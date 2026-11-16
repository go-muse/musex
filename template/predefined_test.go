package template_test

import (
	"encoding/json/v2"
	"os"
	"strings"
	"testing"

	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type builtinRecord struct {
	Code      string   `json:"code"`
	Function  string   `json:"function"`
	Adjacent  string   `json:"adjacent"`
	Names     []string `json:"names"`
	Semitones []int    `json:"semitones"`
	Spellings []string `json:"spellings"`
}

func loadBuiltinRecords(t testing.TB) []builtinRecord {
	t.Helper()
	data, err := os.ReadFile("testdata/builtin-catalog.json")
	require.NoError(t, err)
	var records []builtinRecord
	require.NoError(t, json.Unmarshal(data, &records))
	require.Len(t, records, 81)
	return records
}

type directEntry struct {
	template func() template.Template
	entry    func() template.Entry
}

func directEntries() map[string]directEntry {
	return map[string]directEntry{
		"ionian":                   {template.Ionian, template.IonianEntry},
		"dorian":                   {template.Dorian, template.DorianEntry},
		"phrygian":                 {template.Phrygian, template.PhrygianEntry},
		"lydian":                   {template.Lydian, template.LydianEntry},
		"mixolydian":               {template.Mixolydian, template.MixolydianEntry},
		"aeolian":                  {template.Aeolian, template.AeolianEntry},
		"locrian":                  {template.Locrian, template.LocrianEntry},
		"harmonic-major":           {template.HarmonicMajor, template.HarmonicMajorEntry},
		"melodic-major":            {template.MelodicMajor, template.MelodicMajorEntry},
		"harmonic-minor":           {template.HarmonicMinor, template.HarmonicMinorEntry},
		"melodic-minor":            {template.MelodicMinor, template.MelodicMinorEntry},
		"double-harmonic-major":    {template.DoubleHarmonicMajor, template.DoubleHarmonicMajorEntry},
		"double-harmonic-minor":    {template.DoubleHarmonicMinor, template.DoubleHarmonicMinorEntry},
		"acoustic":                 {template.Acoustic, template.AcousticEntry},
		"dorian-b2":                {template.DorianB2, template.DorianB2Entry},
		"lydian-augmented":         {template.LydianAugmented, template.LydianAugmentedEntry},
		"locrian-sharp2":           {template.LocrianSharp2, template.LocrianSharp2Entry},
		"altered":                  {template.Altered, template.AlteredEntry},
		"pentatonic-major":         {template.PentatonicMajor, template.PentatonicMajorEntry},
		"pentatonic-minor":         {template.PentatonicMinor, template.PentatonicMinorEntry},
		"blues":                    {template.Blues, template.BluesEntry},
		"blues-b5":                 {template.BluesB5, template.BluesB5Entry},
		"blues-major":              {template.BluesMajor, template.BluesMajorEntry},
		"whole-tone":               {template.WholeTone, template.WholeToneEntry},
		"octatonic-wh":             {template.OctatonicWH, template.OctatonicWHEntry},
		"octatonic-hw":             {template.OctatonicHW, template.OctatonicHWEntry},
		"hexatonic":                {template.Hexatonic, template.HexatonicEntry},
		"messiaen-3":               {template.Messiaen3, template.Messiaen3Entry},
		"messiaen-3-commons":       {template.Messiaen3Commons, template.Messiaen3CommonsEntry},
		"messiaen-4-papadopoulos":  {template.Messiaen4Papadopoulos, template.Messiaen4PapadopoulosEntry},
		"messiaen-4-degraaf":       {template.Messiaen4Degraaf, template.Messiaen4DegraafEntry},
		"messiaen-4-commons":       {template.Messiaen4Commons, template.Messiaen4CommonsEntry},
		"messiaen-5-papadopoulos":  {template.Messiaen5Papadopoulos, template.Messiaen5PapadopoulosEntry},
		"messiaen-5-degraaf":       {template.Messiaen5Degraaf, template.Messiaen5DegraafEntry},
		"messiaen-5-commons":       {template.Messiaen5Commons, template.Messiaen5CommonsEntry},
		"messiaen-6":               {template.Messiaen6, template.Messiaen6Entry},
		"messiaen-7-degraaf":       {template.Messiaen7Degraaf, template.Messiaen7DegraafEntry},
		"messiaen-7-commons":       {template.Messiaen7Commons, template.Messiaen7CommonsEntry},
		"bebop-dominant":           {template.BebopDominant, template.BebopDominantEntry},
		"bebop-major":              {template.BebopMajor, template.BebopMajorEntry},
		"locrian-natural6":         {template.LocrianNatural6, template.LocrianNatural6Entry},
		"ionian-sharp5":            {template.IonianSharp5, template.IonianSharp5Entry},
		"ukrainian-dorian":         {template.UkrainianDorian, template.UkrainianDorianEntry},
		"phrygian-dominant":        {template.PhrygianDominant, template.PhrygianDominantEntry},
		"lydian-sharp2":            {template.LydianSharp2, template.LydianSharp2Entry},
		"ultralocrian":             {template.Ultralocrian, template.UltralocrianEntry},
		"dorian-b5":                {template.DorianB5, template.DorianB5Entry},
		"phrygian-b4":              {template.PhrygianB4, template.PhrygianB4Entry},
		"lydian-b3":                {template.LydianB3, template.LydianB3Entry},
		"mixolydian-b2":            {template.MixolydianB2, template.MixolydianB2Entry},
		"lydian-sharp2-sharp5":     {template.LydianSharp2Sharp5, template.LydianSharp2Sharp5Entry},
		"locrian-bb7":              {template.LocrianBB7, template.LocrianBB7Entry},
		"lydian-sharp2-sharp6":     {template.LydianSharp2Sharp6, template.LydianSharp2Sharp6Entry},
		"ultraphrygian":            {template.Ultraphrygian, template.UltraphrygianEntry},
		"oriental":                 {template.Oriental, template.OrientalEntry},
		"ionian-augmented-sharp2":  {template.IonianAugmentedSharp2, template.IonianAugmentedSharp2Entry},
		"locrian-bb3-bb7":          {template.LocrianBB3BB7, template.LocrianBB3BB7Entry},
		"hungarian-major":          {template.HungarianMajor, template.HungarianMajorEntry},
		"aeolian-sharp4":           {template.AeolianSharp4, template.AeolianSharp4Entry},
		"neapolitan-minor":         {template.NeapolitanMinor, template.NeapolitanMinorEntry},
		"neapolitan-major":         {template.NeapolitanMajor, template.NeapolitanMajorEntry},
		"persian":                  {template.Persian, template.PersianEntry},
		"enigmatic":                {template.Enigmatic, template.EnigmaticEntry},
		"prometheus":               {template.Prometheus, template.PrometheusEntry},
		"tritone":                  {template.Tritone, template.TritoneEntry},
		"two-semitone-tritone":     {template.TwoSemitoneTritone, template.TwoSemitoneTritoneEntry},
		"augmented":                {template.Augmented, template.AugmentedEntry},
		"pentatonic-suspended":     {template.PentatonicSuspended, template.PentatonicSuspendedEntry},
		"pentatonic-blues-minor":   {template.PentatonicBluesMinor, template.PentatonicBluesMinorEntry},
		"yo":                       {template.Yo, template.YoEntry},
		"hirajoshi-burrows":        {template.HirajoshiBurrows, template.HirajoshiBurrowsEntry},
		"hirajoshi-sachs":          {template.HirajoshiSachs, template.HirajoshiSachsEntry},
		"hirajoshi-kostka":         {template.HirajoshiKostka, template.HirajoshiKostkaEntry},
		"iwato":                    {template.Iwato, template.IwatoEntry},
		"in":                       {template.In, template.InEntry},
		"insen":                    {template.Insen, template.InsenEntry},
		"bebop-melodic-minor":      {template.BebopMelodicMinor, template.BebopMelodicMinorEntry},
		"bebop-harmonic-minor":     {template.BebopHarmonicMinor, template.BebopHarmonicMinorEntry},
		"seventh-flat5-diminished": {template.SeventhFlat5Diminished, template.SeventhFlat5DiminishedEntry},
		"bebop-dorian":             {template.BebopDorian, template.BebopDorianEntry},
		"kumoi":                    {template.Kumoi, template.KumoiEntry},
	}
}

func TestBuiltInCatalogMatchesOracle(t *testing.T) {
	t.Parallel()
	records := loadBuiltinRecords(t)
	direct := directEntries()
	require.Len(t, direct, 81)
	catalog := template.BuiltIn()
	require.Equal(t, 81, catalog.Len())
	entries := catalog.Entries()
	require.Len(t, entries, len(records))
	for i, record := range records {
		t.Run(record.Code, func(t *testing.T) {
			t.Parallel()
			entry := entries[i]
			assert.Equal(t, record.Code, entry.Code())
			assert.Equal(t, record.Adjacent, entry.Template().String())
			assert.Equal(t, record.Names, entry.Names())
			cycle, ok := entry.Template().Cycle()
			require.True(t, ok)
			requireInterval(t, "P8", cycle)
			assert.True(t, entry.Template().IsOctave())

			directFunctions, ok := direct[record.Code]
			require.True(t, ok)
			assert.True(t, directFunctions.template().Equal(entry.Template()))
			assert.True(t, directFunctions.entry().Equal(entry))
			byCode, ok := catalog.ByCode(record.Code)
			require.True(t, ok)
			assert.True(t, byCode.Equal(entry))
			for _, name := range record.Names {
				matches := catalog.ByName(name)
				assert.NotNil(t, matches)
				assert.True(t, slicesContainsEntry(matches, entry), "name %q does not find %q", name, record.Code)
			}
		})
	}
}

func slicesContainsEntry(entries []template.Entry, want template.Entry) bool {
	return slicesContainsFunc(entries, want.Equal)
}

func slicesContainsFunc(values []template.Entry, predicate func(template.Entry) bool) bool {
	for _, value := range values {
		if predicate(value) {
			return true
		}
	}
	return false
}

func TestBuiltInCatalogDomainCrossChecks(t *testing.T) {
	t.Parallel()
	for _, record := range loadBuiltinRecords(t) {
		t.Run(record.Code, func(t *testing.T) {
			t.Parallel()
			entry, ok := template.BuiltIn().ByCode(record.Code)
			require.True(t, ok)
			adjacent := entry.Template().AdjacentIntervals()
			require.Len(t, adjacent, len(record.Semitones))
			for i, value := range adjacent {
				steps, err := value.Steps(system.Traditional(), spelling.Spelling{})
				require.NoError(t, err)
				assert.EqualValues(t, record.Semitones[i], steps, "interval %d", i+1)
			}
			fromNames, err := template.FromSpellingsInOctave(names(strings.Join(record.Spellings, " ")))
			require.NoError(t, err)
			assert.True(t, fromNames.Equal(entry.Template()))
		})
	}
}

func TestBuiltInIsNotChangedByUserOperations(t *testing.T) {
	t.Parallel()
	before := template.BuiltIn()
	_, err := before.Without("ionian")
	require.NoError(t, err)
	_, err = before.With(template.MustNewEntry("custom", []string{"Custom"}, template.Ionian()))
	require.NoError(t, err)
	filtered, err := before.Filter(func(template.Entry) bool { return false })
	require.NoError(t, err)
	assert.Zero(t, filtered.Len())
	assert.Equal(t, 81, template.BuiltIn().Len())
	assert.True(t, template.BuiltIn().Equal(before))
	entries := template.BuiltIn().Entries()
	entries[0] = template.Entry{}
	assert.Equal(t, "ionian", template.BuiltIn().Entries()[0].Code())
}

func TestBuiltInZeroDifferenceRelations(t *testing.T) {
	t.Parallel()
	catalog := template.BuiltIn()
	for _, entry := range catalog.Entries() {
		t.Run(entry.Code(), func(t *testing.T) {
			t.Parallel()
			tmpl := entry.Template()
			identified, err := catalog.Identify(tmpl)
			require.NoError(t, err)
			rotations, err := catalog.IdentifyRotations(tmpl)
			require.NoError(t, err)
			for _, measure := range []template.Measure{template.FromFirst, template.Adjacent} {
				related, err := catalog.Related(tmpl, 0, measure, false)
				require.NoError(t, err)
				assert.Equal(t, entryCodes(identified), entryCodes(relationEntries(related)))
				relatedRotations, err := catalog.Related(tmpl, 0, measure, true)
				require.NoError(t, err)
				assert.Equal(t, rotationCodes(rotations), rotationCodes(relationRotations(relatedRotations)))
				assert.Equal(t, rotationPlaces(rotations), rotationPlaces(relationRotations(relatedRotations)))
				for _, relation := range relatedRotations {
					rotated, err := relation.Entry.Template().Rotate(relation.Place)
					require.NoError(t, err)
					places, err := rotated.MismatchingPlaces(tmpl, measure)
					require.NoError(t, err)
					assert.Equal(t, relation.Places, places)
					assert.Equal(t, len(places), relation.Count())
				}
			}
		})
	}
}

func relationEntries(relations []template.Relation) []template.Entry {
	result := make([]template.Entry, len(relations))
	for i, relation := range relations {
		result[i] = relation.Entry
	}
	return result
}

func relationRotations(relations []template.Relation) []template.Rotation {
	result := make([]template.Rotation, len(relations))
	for i, relation := range relations {
		result[i] = template.Rotation{Entry: relation.Entry, Place: relation.Place}
	}
	return result
}
