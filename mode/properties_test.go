package mode_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInclinationAndBrightness(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		value mode.Inclination
		label string
		valid bool
	}{
		{mode.InclinationUndefined, "undefined", true}, {mode.Major, "major", true}, {mode.Minor, "minor", true},
		{mode.Inclination(255), "Inclination(255)", false},
	} {
		assert.Equal(t, tt.valid, tt.value.IsValid())
		assert.Equal(t, tt.label, tt.value.String())
	}
	assert.Equal(t, mode.Inclination(0), mode.InclinationUndefined)
	assert.Equal(t, mode.Inclination(1), mode.Major)
	assert.Equal(t, mode.Inclination(2), mode.Minor)
	for _, tt := range []struct {
		code                               string
		secondwise, diatonic, onlyDiatonic bool
		inclination                        mode.Inclination
		brightness                         int
	}{
		{"lydian", true, true, true, mode.Major, 1}, {"ionian", true, true, true, mode.Major, 2},
		{"mixolydian", true, true, true, mode.Major, 3}, {"dorian", true, true, true, mode.Minor, 4},
		{"aeolian", true, true, true, mode.Minor, 5}, {"phrygian", true, true, true, mode.Minor, 6},
		{"locrian", true, true, true, mode.Minor, 7}, {"harmonic-minor", true, false, false, mode.Minor, 0},
		{"altered", true, false, false, mode.Minor, 0}, {"pentatonic-major", false, false, true, mode.Major, 0},
		{"blues", false, false, false, mode.Minor, 0}, {"whole-tone", false, false, false, mode.Major, 0},
		{"octatonic-hw", false, false, false, mode.InclinationUndefined, 0},
	} {
		t.Run(tt.code, func(t *testing.T) {
			m := mode.MustFromCode(tt.code)
			assert.Equal(t, tt.secondwise, m.IsSecondwise())
			assert.Equal(t, tt.diatonic, m.MustIsDiatonic())
			assert.Equal(t, tt.onlyDiatonic, m.MustOnlyDiatonicIntervals())
			assert.Equal(t, tt.inclination, m.MustInclination())
			rank, ok, err := m.Brightness()
			require.NoError(t, err)
			assert.Equal(t, tt.brightness > 0, ok)
			assert.Equal(t, tt.brightness, rank)
		})
	}
	open := mode.MustParse("M2 M2 m2 M2 M2 M2 |")
	assert.False(t, open.IsSecondwise())
	assert.False(t, open.MustIsDiatonic())
	assert.True(t, open.MustOnlyDiatonicIntervals())
	assert.Equal(t, mode.Major, open.MustInclination())
	rank, ok := open.MustBrightness()
	assert.False(t, ok)
	assert.Zero(t, rank)
	m := mode.MustFromCode("ionian")
	assert.Equal(t, -1, mode.MustFromCode("lydian").MustCompareBrightness(m))
	assert.Equal(t, 1, m.MustCompareBrightness(mode.MustFromCode("lydian")))
	assert.Equal(t, -1, mode.MustFromCode("dorian").MustCompareBrightness(mode.MustFromCode("aeolian")))
	assert.Zero(t, m.MustCompareBrightness(mode.MustParse(ionian)))
	rank, err := m.CompareBrightness(mode.MustFromCode("harmonic-minor"))
	checkError(t, err, mode.ErrIncomparable)
	assert.Zero(t, rank)
	assert.True(t, mode.MustParse("P8").MustOnlyDiatonicIntervals())
	assert.Equal(t, mode.InclinationUndefined, mode.MustParse("P8").MustInclination())
}

type intervalRow struct {
	degree, distance int
	name             string
}

func intervalRows(results []mode.IntervalAt) []intervalRow {
	rows := make([]intervalRow, len(results))
	for i, r := range results {
		rows[i] = intervalRow{r.Degree.Number(), r.Distance, r.Interval.String()}
	}
	return rows
}

func TestIntervalTableOrderAndFilter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, structure string
		rows            []intervalRow
	}{
		{"closed", "M2 M2 m2", []intervalRow{{1, 1, "M2"}, {1, 2, "M3"}, {2, 1, "M2"}, {2, 2, "m3"}, {3, 1, "m2"}, {3, 2, "m3"}}},
		{"open", "m3 M2 |", []intervalRow{{1, 1, "m3"}, {1, 2, "P4"}, {2, 1, "M2"}}},
		{"singleton", "P8", []intervalRow{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mode.MustParse(tt.structure)
			seq, err := m.Intervals()
			require.NoError(t, err)
			require.NotNil(t, seq)
			all := slices.Collect(seq)
			for _, r := range all {
				assert.False(t, r.Degree.HasNote())
			}
			assert.Equal(t, tt.rows, intervalRows(all))
			if len(tt.rows) > 0 {
				walked := 0
				for row := range seq {
					assert.Equal(t, tt.rows[0].name, row.Interval.String())
					walked++
					break
				}
				assert.Equal(t, 1, walked)
				all[0] = mode.IntervalAt{}
			}
			assert.Equal(t, tt.rows, intervalRows(slices.Collect(seq)))
		})
	}
	for _, tt := range []struct {
		code string
		want []intervalRow
	}{
		{"ionian", []intervalRow{}},
		{"harmonic-minor", []intervalRow{{3, 4, "A5"}, {6, 1, "A2"}, {7, 3, "d4"}, {7, 6, "d7"}}},
		{"harmonic-major", []intervalRow{{3, 3, "d4"}, {6, 1, "A2"}, {6, 4, "A5"}, {7, 6, "d7"}}},
	} {
		m := mode.MustFromCode(tt.code)
		seq, err := m.ChromaticIntervals()
		require.NoError(t, err)
		require.NotNil(t, seq)
		assert.Equal(t, tt.want, intervalRows(slices.Collect(seq)))
	}
}

func TestDegreesForInterval(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code, want string
		degrees    []int
	}{
		{"ionian", "A4", []int{4}}, {"ionian", "d5", []int{7}}, {"ionian", "M3", []int{1, 4, 5}},
		{"ionian", "P1", []int{1, 2, 3, 4, 5, 6, 7}}, {"ionian", "P8", []int{1, 2, 3, 4, 5, 6, 7}},
		{"ionian", "A2", []int{}}, {"harmonic-minor", "A2", []int{6}}, {"harmonic-minor", "A5", []int{3}},
		{"harmonic-minor", "d4", []int{7}}, {"harmonic-minor", "d7", []int{7}},
	} {
		m := mode.MustFromCode(tt.code)
		got, err := m.DegreesForInterval(interval.MustParse(tt.want))
		require.NoError(t, err)
		require.NotNil(t, got)
		checkDegrees(t, got, tt.degrees...)
	}
	open := mode.MustParse("m3 M2 |")
	checkDegrees(t, open.MustDegreesForInterval(interval.MustParse("P1")), 1, 2, 3)
	checkDegrees(t, open.MustDegreesForInterval(interval.MustParse("M2")), 2)
	checkDegrees(t, open.MustDegreesForInterval(interval.MustParse("P4")), 1)
	got, err := open.DegreesForInterval(interval.MustParse("-M2"))
	checkError(t, err, mode.ErrInvalidArgument)
	assert.Nil(t, got)
	got, err = open.DegreesForInterval(interval.Interval{})
	checkError(t, err, mode.ErrUnset)
	assert.Nil(t, got)
}

func TestIntervalTableChecksOnlyPromisedResults(t *testing.T) {
	t.Parallel()
	large := mode.MustClosedFromIntervals(extremeSeconds())
	all, err := large.Intervals()
	checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Nil(t, all)
	all, err = large.ChromaticIntervals()
	checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Nil(t, all)
	only, err := large.OnlyDiatonicIntervals()
	require.NoError(t, err)
	assert.False(t, only)
	checkDegrees(t, large.MustDegreesForInterval(interval.MustParse("P5")), 1, 2, 3, 4)

	// Between degrees 2 and 4 the fractional part is 1/65535-1/65536.
	// The denominator of that difference is too large but it is not chromatic.
	a := interval.MustFromDeviation(2, alteration.MustFromFraction(1, 65536))
	b := interval.MustFromDeviation(2, alteration.MustFromFraction(-1, 65536))
	c := interval.MustFromDeviation(2, alteration.MustFromFraction(1, 65535))
	fractional := mode.MustOpenFromIntervals([]interval.Interval{a, b, c})
	all, err = fractional.Intervals()
	checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Nil(t, all)
	chromatic, err := fractional.ChromaticIntervals()
	require.NoError(t, err)
	require.NotNil(t, chromatic)
	assert.Equal(t, []intervalRow{}, intervalRows(slices.Collect(chromatic)))
	assert.False(t, fractional.MustOnlyDiatonicIntervals())
}

func TestTetrachordsAndPeriod(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code, lower, connector, upper, lowerRU, lowerEN, upperRU, upperEN string
		tone                                                              bool
	}{
		{"ionian", "M2 M2 m2", "M2", "M2 M2 m2", "ионийский", "Major", "ионийский", "Major", true},
		{"lydian", "M2 M2 M2", "m2", "M2 M2 m2", "лидийский", "", "ионийский", "Major", false},
		{"harmonic-minor", "M2 m2 M2", "M2", "m2 A2 m2", "дорийский", "Minor", "", "Harmonic", true},
		{"phrygian", "m2 M2 M2", "M2", "m2 M2 M2", "фригийский", "Upper Minor", "фригийский", "Upper Minor", true},
	} {
		t.Run(tt.code, func(t *testing.T) {
			pair, err := mode.MustFromCode(tt.code).Tetrachords()
			require.NoError(t, err)
			assert.Equal(t, tt.lower, intervalText(pair.Lower.Intervals[:]))
			assert.Equal(t, tt.upper, intervalText(pair.Upper.Intervals[:]))
			assert.Equal(t, tt.connector, pair.Connector.String())
			assert.Equal(t, tt.lowerRU, pair.Lower.RussianName)
			assert.Equal(t, tt.lowerEN, pair.Lower.EnglishName)
			assert.Equal(t, tt.upperRU, pair.Upper.RussianName)
			assert.Equal(t, tt.upperEN, pair.Upper.EnglishName)
			assert.Equal(t, tt.tone, pair.SeparatedByTone())
		})
	}
	assert.False(t, (mode.TetrachordPair{}).SeparatedByTone())
	unnamed := mode.MustParse("M3 m2 M2 M2 M2 m2 m3").MustTetrachords()
	assert.Empty(t, unnamed.Lower.RussianName)
	assert.Empty(t, unnamed.Lower.EnglishName)
	pair, err := mode.MustFromCode("pentatonic-major").Tetrachords()
	checkError(t, err, mode.ErrWrongDegreeCount)
	assert.Zero(t, pair)
	pair, err = mode.MustParse("M2 |").Tetrachords()
	checkError(t, err, mode.ErrOpenMode)
	assert.Zero(t, pair)
	for _, tt := range []struct {
		code   string
		period int
	}{
		{"ionian", 7}, {"pentatonic-major", 5}, {"blues", 6}, {"whole-tone", 6},
	} {
		m := mode.MustFromCode(tt.code)
		assert.Equal(t, tt.period, m.MustPeriod())
		assert.Equal(t, tt.period, m.MustDistinctRotations())
		part, repeat := m.MustMinimalPart()
		assert.Equal(t, intervalText(m.AdjacentIntervals()), intervalText(part))
		assert.Equal(t, 1, repeat)
	}
	m := mode.MustParse("M2 m2 M2 m2")
	assert.Equal(t, 2, m.MustPeriod())
	assert.Equal(t, 2, m.MustDistinctRotations())
	part, repeat := m.MustMinimalPart()
	assert.Equal(t, "M2 m2", intervalText(part))
	assert.Equal(t, 2, repeat)
	rebuilt := mode.MustClosedFromIntervals(part)
	require.NoError(t, rebuilt.Repeat(repeat))
	assert.True(t, m.Equal(rebuilt))
	part[0] = interval.Interval{}
	assert.Equal(t, "M2 m2 M2 m2", intervalText(m.AdjacentIntervals()))
	assert.True(t, mode.MustFromCode("dorian").IsPalindrome())
	assert.False(t, mode.MustFromCode("harmonic-minor").IsPalindrome())
	assert.True(t, mode.MustParse("M2 m2 M2 |").IsPalindrome())
	assert.False(t, mode.MustParse("M2 m3 |").IsPalindrome())

	open := mode.MustParse("M2 |")
	p, err := open.Period()
	checkError(t, err, mode.ErrOpenMode)
	assert.Zero(t, p)
	p, err = open.DistinctRotations()
	checkError(t, err, mode.ErrOpenMode)
	assert.Zero(t, p)
	part, repeat, err = open.MinimalPart()
	checkError(t, err, mode.ErrOpenMode)
	assert.Nil(t, part)
	assert.Zero(t, repeat)
}
