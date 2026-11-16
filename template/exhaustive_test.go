package template_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExhaustiveTemplateProperties(t *testing.T) {
	t.Parallel()
	cases := controlTemplates()
	for _, record := range loadBuiltinRecords(t) {
		entry, ok := template.BuiltIn().ByCode(record.Code)
		require.True(t, ok)
		cases = append(cases, controlTemplate{name: "built-in " + record.Code, text: record.Adjacent})
		cases[len(cases)-1].value = entry.Template
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkAllRotationsAndPlaces(t, tc.template())
		})
	}
}

func checkAllRotationsAndPlaces(t *testing.T, original template.Template) {
	t.Helper()
	k := original.Len()
	originalAdjacent := original.AdjacentIntervals()
	rotationLimit := k
	if original.IsOpen() {
		rotationLimit = 1
	}
	for rotationPlace := 1; rotationPlace <= rotationLimit; rotationPlace++ {
		rotated, err := original.Rotate(rotationPlace)
		require.NoError(t, err)
		assert.Equal(t, k, rotated.Len())
		assert.Equal(t, original.Kind(), rotated.Kind())
		assert.Equal(t, intervalTexts(rotateIntervalSlice(originalAdjacent, rotationPlace)), intervalTexts(rotated.AdjacentIntervals()))
		if original.IsClosed() {
			before, ok := original.Cycle()
			require.True(t, ok)
			after, ok := rotated.Cycle()
			require.True(t, ok)
			assert.True(t, after.Equal(before))
		} else {
			_, ok := rotated.Cycle()
			assert.False(t, ok)
		}

		places, err := rotated.PlaceIntervals()
		require.NoError(t, err)
		if rotated.IsClosed() {
			cycle, ok := rotated.Cycle()
			require.True(t, ok)
			rebuilt, err := template.ClosedFromPlaceIntervals(places, cycle)
			require.NoError(t, err)
			assert.True(t, rebuilt.Equal(rotated))
		} else {
			rebuilt, err := template.OpenFromPlaceIntervals(places)
			require.NoError(t, err)
			assert.True(t, rebuilt.Equal(rotated))
		}

		adjacentCoordinates := coordinatesOf(rotated.AdjacentIntervals())
		limit := 2*k + 1
		if rotated.IsOpen() {
			limit = k
		}
		for place := 1; place <= limit; place++ {
			got, err := rotated.PlaceInterval(place)
			require.NoError(t, err, "place %d", place)
			requireCoordinate(t, placeCoordinate(adjacentCoordinates, rotated.IsClosed(), place), got)
			if rotated.IsClosed() || place < k {
				adjacent, err := rotated.AdjacentIntervalAfter(place)
				require.NoError(t, err)
				expected := rotated.AdjacentIntervals()[(place-1)%k]
				assert.True(t, adjacent.Equal(expected))
			}
		}

		for i := 1; i <= limit; i++ {
			for j := 1; j <= limit; j++ {
				got, err := rotated.BetweenPlaces(i, j)
				require.NoError(t, err, "places %d -> %d", i, j)
				want := subCoordinates(
					placeCoordinate(adjacentCoordinates, rotated.IsClosed(), j),
					placeCoordinate(adjacentCoordinates, rotated.IsClosed(), i),
				)
				requireCoordinate(t, want, got)
				assert.Equal(t, i == j, got.Equal(interval.PerfectUnison()))
				for m := 1; m <= limit; m++ {
					second, err := rotated.BetweenPlaces(j, m)
					require.NoError(t, err)
					total, err := rotated.BetweenPlaces(i, m)
					require.NoError(t, err)
					combined := addCoordinates(coordinateOf(got), coordinateOf(second))
					requireCoordinate(t, combined, total)
				}
			}
		}

		place, ok := rotated.RotationOf(original)
		require.True(t, ok)
		assert.Equal(t, smallestRotationPlace(original, rotated), place)
		back, ok := original.RotationOf(rotated)
		require.True(t, ok)
		backTemplate, err := rotated.Rotate(back)
		require.NoError(t, err)
		assert.True(t, backTemplate.Equal(original))
		for nextPlace := 1; nextPlace <= rotationLimit; nextPlace++ {
			composed, err := rotated.Rotate(nextPlace)
			require.NoError(t, err)
			directPlace := (rotationPlace+nextPlace-2)%k + 1
			direct, err := original.Rotate(directPlace)
			require.NoError(t, err)
			assert.True(t, composed.Equal(direct))
		}
		assert.True(t, rotated.Equal(rotated))
		assert.Zero(t, rotated.Compare(rotated))
		assert.Equal(t, rotated.Equal(original) == true, rotated.Compare(original) == 0)
		assert.Equal(t, -rotated.Compare(original), original.Compare(rotated))
		for _, measure := range []template.Measure{template.FromFirst, template.Adjacent} {
			places, err := rotated.MismatchingPlaces(rotated, measure)
			require.NoError(t, err)
			assert.NotNil(t, places)
			assert.Empty(t, places)
		}
		parsed, err := template.Parse(rotated.String())
		require.NoError(t, err)
		assert.True(t, parsed.Equal(rotated))
	}
}

func rotateIntervalSlice(input []interval.Interval, place int) []interval.Interval {
	k := len(input)
	if k == 0 {
		return nil
	}
	result := make([]interval.Interval, k)
	for i := range k {
		result[i] = input[(i+place-1)%k]
	}
	return result
}

func smallestRotationPlace(original, rotated template.Template) int {
	for place := 1; place <= original.Len(); place++ {
		candidate, err := original.Rotate(place)
		if err == nil && candidate.Equal(rotated) {
			return place
		}
	}
	return 0
}

func TestNamedRotationFamilies(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code      string
		rotations [][]string
	}{
		{code: "ionian", rotations: [][]string{{"ionian"}, {"dorian"}, {"phrygian"}, {"lydian"}, {"mixolydian"}, {"aeolian"}, {"locrian"}}},
		{code: "melodic-minor", rotations: [][]string{{"melodic-minor"}, {"dorian-b2"}, {"lydian-augmented"}, {"acoustic"}, {"melodic-major"}, {"locrian-sharp2"}, {"altered"}}},
		{code: "harmonic-minor", rotations: [][]string{{"harmonic-minor"}, {"locrian-natural6"}, {"ionian-sharp5"}, {"ukrainian-dorian"}, {"phrygian-dominant"}, {"lydian-sharp2"}, {"ultralocrian"}}},
		{code: "harmonic-major", rotations: [][]string{{"harmonic-major"}, {"dorian-b5"}, {"phrygian-b4"}, {"lydian-b3"}, {"mixolydian-b2"}, {"lydian-sharp2-sharp5"}, {"locrian-bb7"}}},
		{code: "double-harmonic-major", rotations: [][]string{{"double-harmonic-major"}, {"lydian-sharp2-sharp6"}, {"ultraphrygian"}, {"double-harmonic-minor"}, {"oriental"}, {"ionian-augmented-sharp2"}, {"locrian-bb3-bb7"}}},
		{code: "pentatonic-major", rotations: [][]string{{"pentatonic-major"}, {"pentatonic-suspended"}, {"pentatonic-blues-minor"}, {"yo"}, {"pentatonic-minor"}}},
		{code: "hirajoshi-kostka", rotations: [][]string{{"hirajoshi-kostka"}, {"hirajoshi-sachs", "iwato"}, {}, {"in"}, {"hirajoshi-burrows"}}},
		{code: "neapolitan-minor", rotations: [][]string{{"neapolitan-minor"}, {}, {}, {"aeolian-sharp4"}, {}, {}, {}}},
		{code: "bebop-major", rotations: [][]string{{"bebop-major"}, {}, {}, {}, {}, {}, {"bebop-harmonic-minor"}, {}}},
		{code: "bebop-dominant", rotations: [][]string{{"bebop-dominant"}, {}, {}, {}, {"bebop-dorian"}, {}, {}, {}}},
		{code: "insen", rotations: [][]string{{"insen"}, {}, {}, {}, {"kumoi"}}},
		{code: "blues", rotations: [][]string{{"blues"}, {"blues-major"}, {}, {}, {}, {}}},
		{code: "octatonic-wh", rotations: [][]string{{"octatonic-wh"}, {}, {}, {"octatonic-hw"}, {}, {}, {}, {}}},
		{code: "two-semitone-tritone", rotations: [][]string{{"two-semitone-tritone"}, {}, {}, {}, {"messiaen-5-commons"}, {}}},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()
			entry, ok := template.BuiltIn().ByCode(tc.code)
			require.True(t, ok)
			require.Len(t, tc.rotations, entry.Template().Len())
			for i, want := range tc.rotations {
				t.Run(fmt.Sprintf("place_%d", i+1), func(t *testing.T) {
					t.Parallel()
					rotated, err := entry.Template().Rotate(i + 1)
					require.NoError(t, err)
					identified, err := template.BuiltIn().Identify(rotated)
					require.NoError(t, err)
					assert.Equal(t, want, entryCodes(identified))
				})
			}
		})
	}
}
