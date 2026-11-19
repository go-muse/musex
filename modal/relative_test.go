package modal_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategory(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		category modal.Category
		number   uint8
		text     string
		valid    bool
	}{
		{modal.CategoryUnset, 0, "unset", false},
		{modal.High, 1, "high", true},
		{modal.Neutral, 2, "neutral", true},
		{modal.Low, 3, "low", true},
		{modal.Category(4), 4, "Category(4)", false},
		{modal.Category(255), 255, "Category(255)", false},
	} {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.number, uint8(tc.category))
			assert.Equal(t, tc.text, tc.category.String())
			assert.Equal(t, tc.valid, tc.category.IsValid())
			assert.True(t, map[modal.Category]bool{tc.category: true}[tc.category])
		})
	}
}

// M-1, M-2, M-8: expected intervals and categories are from the Dorian table.
func TestRelativeDorian(t *testing.T) {
	t.Parallel()
	m := dorianMode(t)
	intervals := degreeIntervals(t, "P1 M2 m3 P4 P5 M6 m7")
	categories := []modal.Category{modal.Neutral, modal.High, modal.Low, modal.Neutral, modal.Neutral, modal.High, modal.Low}
	want := make([]modal.RelativePosition, len(intervals))
	for index, value := range intervals {
		want[index] = modal.RelativePosition{Number: index + 1, Category: categories[index], Interval: value}
		got, err := modal.Relative(m, index+1)
		require.NoError(t, err)
		assert.Equal(t, want[index], got)
		assert.True(t, got.Interval.Equal(value))
		assert.Equal(t, want[index], modal.MustRelative(m, index+1))
	}
	all, err := modal.RelativePositions(m)
	require.NoError(t, err)
	assert.Equal(t, want, all)
	assert.Equal(t, want, modal.MustRelativePositions(m))
	groups, err := modal.GroupRelative(m)
	require.NoError(t, err)
	expected := modal.RelativeGroups{High: []int{2, 6}, Neutral: []int{1, 4, 5}, Low: []int{3, 7}}
	assert.Equal(t, expected, groups)
	assert.Equal(t, expected, modal.MustGroupRelative(m))
}

func TestRelativeStructureAndEmptyGroups(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		degrees    string
		cycle      string
		categories []modal.Category
		groups     modal.RelativeGroups
	}{
		{"open", "P1 M2 M3", "", []modal.Category{modal.Neutral, modal.High, modal.High}, modal.RelativeGroups{High: []int{2, 3}, Neutral: []int{1}, Low: []int{}}},
		{"fourth cycle", "P1 M2 M3", "P4", []modal.Category{modal.Neutral, modal.High, modal.High}, modal.RelativeGroups{High: []int{2, 3}, Neutral: []int{1}, Low: []int{}}},
		{"double octave", "P1 M9", "P15", []modal.Category{modal.Neutral, modal.High}, modal.RelativeGroups{High: []int{2}, Neutral: []int{1}, Low: []int{}}},
		{"augmented cycle", "P1 m2", "A2", []modal.Category{modal.Neutral, modal.Low}, modal.RelativeGroups{High: []int{}, Neutral: []int{1}, Low: []int{2}}},
		{"pentatonic", "P1 M2 M3 P5 M6", "P8", []modal.Category{modal.Neutral, modal.High, modal.High, modal.Neutral, modal.High}, modal.RelativeGroups{High: []int{2, 3, 5}, Neutral: []int{1, 4}, Low: []int{}}},
		{"single degree", "P1", "P8", []modal.Category{modal.Neutral}, modal.RelativeGroups{High: []int{}, Neutral: []int{1}, Low: []int{}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var m mode.Mode
			if tc.cycle == "" {
				m = openMode(t, tc.degrees)
			} else {
				m = closedMode(t, tc.degrees, tc.cycle)
			}
			values := degreeIntervals(t, tc.degrees)
			want := make([]modal.RelativePosition, len(values))
			for i, v := range values {
				want[i] = modal.RelativePosition{Number: i + 1, Category: tc.categories[i], Interval: v}
				got, err := modal.Relative(m, i+1)
				require.NoError(t, err)
				assert.Equal(t, want[i], got)
			}
			got, err := modal.RelativePositions(m)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			groups, err := modal.GroupRelative(m)
			require.NoError(t, err)
			assert.Equal(t, tc.groups, groups)
			assert.NotNil(t, groups.High)
			assert.NotNil(t, groups.Neutral)
			assert.NotNil(t, groups.Low)
			assert.Equal(t, want, modal.MustRelativePositions(m))
			assert.Equal(t, tc.groups, modal.MustGroupRelative(m))
		})
	}
}

func TestRelativeExactQualityAndMultiplicity(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name         string
		quality      interval.Quality
		size         int
		multiplicity int
		deviation    int
		category     modal.Category
	}{
		{"M2", interval.Major, 2, 0, 0, modal.High},
		{"m2", interval.Minor, 2, 0, -1, modal.Low},
		{"A2", interval.Augmented, 2, 1, 1, modal.High},
		{"AA2", interval.Augmented, 2, 2, 2, modal.High},
		{"AAA4", interval.Augmented, 4, 3, 3, modal.High},
		{"17-fold augmented", interval.Augmented, 4, 17, 17, modal.High},
		{"maximum augmented", interval.Augmented, 4, alteration.MaxMagnitude, alteration.MaxMagnitude, modal.High},
		{"d5", interval.Diminished, 5, 1, -1, modal.Low},
		{"dd2", interval.Diminished, 2, 2, -3, modal.Low},
		{"17-fold diminished", interval.Diminished, 5, 17, -17, modal.Low},
		{"maximum diminished", interval.Diminished, 5, alteration.MaxMagnitude, -alteration.MaxMagnitude, modal.Low},
		{"M9", interval.Major, 9, 0, 0, modal.High},
		{"compound AA11", interval.Augmented, 11, 2, 2, modal.High},
		{"P15", interval.Perfect, 15, 0, 0, modal.Neutral},
		{"m16", interval.Minor, 16, 0, -1, modal.Low},
		{"maximum full generic", interval.Major, math.MaxInt, 0, 0, modal.High},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var value interval.Interval
			var err error
			if tc.multiplicity == 0 {
				value, err = interval.New(tc.quality, tc.size)
			} else {
				value, err = interval.NewMultiple(tc.multiplicity, tc.quality, tc.size)
			}
			require.NoError(t, err)
			m, err := mode.OpenFromDegreeIntervals([]interval.Interval{interval.PerfectUnison(), value})
			require.NoError(t, err)
			want := modal.RelativePosition{Number: 2, Category: tc.category, Interval: value}
			got, err := modal.Relative(m, 2)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.True(t, got.Interval.Equal(value))
			assert.Equal(t, tc.size, got.Interval.Generic().Size())
			assert.Equal(t, interval.Up, got.Interval.Direction())
			assert.Equal(t, tc.quality, got.Interval.Quality())
			assert.Equal(t, tc.multiplicity, got.Interval.Multiplicity())
			assert.True(t, got.Interval.Deviation().Equal(alteration.MustFromInt(tc.deviation)))
			all, err := modal.RelativePositions(m)
			require.NoError(t, err)
			require.Len(t, all, 2)
			assert.Equal(t, want, all[1])
			assert.Equal(t, want, modal.MustRelative(m, 2))
			groups, err := modal.GroupRelative(m)
			require.NoError(t, err)
			assert.Contains(t, map[modal.Category][]int{modal.High: groups.High, modal.Neutral: groups.Neutral, modal.Low: groups.Low}[tc.category], 2)
		})
	}
}

func TestRelativePreservesSpecificSpelling(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		specific string
		category modal.Category
	}{{"A4", modal.High}, {"d5", modal.Low}} {
		t.Run(fmt.Sprint(tc.specific), func(t *testing.T) {
			t.Parallel()
			m := openMode(t, "P1 "+tc.specific)
			got, err := modal.Relative(m, 2)
			require.NoError(t, err)
			assert.Equal(t, tc.category, got.Category)
			assert.Equal(t, tc.specific, got.Interval.String())
		})
	}
}
