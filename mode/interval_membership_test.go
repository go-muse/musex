package mode_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func altForReading(value int) alteration.Alteration { return alteration.MustFromInt(value) }
func readingPitch(sys *system.System, name spelling.Spelling, written int) pitch.Pitch {
	return sys.MustFromWritten(name, octave.FromASPN(written))
}

func TestMatchSpellings(t *testing.T) {
	t.Parallel()
	major := mode.MustFromCode("ionian")
	melodic := mode.MustFromCode("melodic-minor")
	open := mode.MustOpenFromIntervals(intervals("P8 P8"))
	for _, tc := range []struct {
		name          string
		m             mode.Mode
		tonic, a, b   spelling.Spelling
		window        mode.CycleWindow
		first, second []int
	}{
		{"melodic roles", melodic, spelling.D(), spelling.F(), spelling.CSharp(), mode.CycleWindow{}, []int{3}, []int{7}},
		{"second independently absent", major, spelling.C(), spelling.F(), spelling.CSharp(), mode.CycleWindow{}, []int{4}, nil},
		{"first independently absent", major, spelling.C(), spelling.CSharp(), spelling.F(), mode.CycleWindow{}, nil, []int{4}},
		{"identical inputs", major, spelling.C(), spelling.C(), spelling.C(), mode.CycleWindow{-1, 2}, []int{1}, []int{1}},
		{"spelling not position", major, spelling.C(), spelling.BSharp(), spelling.E(), mode.CycleWindow{}, nil, []int{3}},
		{"duplicate names", open, spelling.C(), spelling.C(), spelling.E(), mode.CycleWindow{-3, 3}, []int{1, 2, 3}, nil},
		{"open outside window", open, spelling.C(), spelling.C(), spelling.C(), mode.CycleWindow{1, 4}, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.m.MatchSpellings(tc.tonic, tc.a, tc.b, tc.window)
			require.NoError(t, err)
			assert.Equal(t, tc.first, emptyNumbers(got.First))
			assert.Equal(t, tc.second, emptyNumbers(got.Second))
			assert.Equal(t, got, tc.m.MustMatchSpellings(tc.tonic, tc.a, tc.b, tc.window))
			if len(got.First) > 0 {
				got.First[0] = 999
			}
			again, err := tc.m.MatchSpellings(tc.tonic, tc.a, tc.b, tc.window)
			require.NoError(t, err)
			assert.Equal(t, tc.first, emptyNumbers(again.First))
		})
	}
	for _, tc := range []struct {
		m           mode.Mode
		tonic, a, b spelling.Spelling
		window      mode.CycleWindow
		want        error
	}{
		{mode.Mode{}, spelling.C(), spelling.C(), spelling.C(), mode.CycleWindow{}, mode.ErrUnset},
		{major, spelling.Spelling{}, spelling.C(), spelling.C(), mode.CycleWindow{}, mode.ErrUnset},
		{major, spelling.C(), spelling.Spelling{}, spelling.C(), mode.CycleWindow{}, mode.ErrUnset},
		{major, spelling.C(), spelling.C(), spelling.Spelling{}, mode.CycleWindow{}, mode.ErrUnset},
		{major, spelling.C(), spelling.C(), spelling.C(), mode.CycleWindow{2, 1}, mode.ErrInvalidArgument},
		{open, spelling.C(), spelling.Spelling{}, spelling.C(), mode.CycleWindow{1, 2}, mode.ErrUnset},
	} {
		got, err := tc.m.MatchSpellings(tc.tonic, tc.a, tc.b, tc.window)
		assert.ErrorIs(t, err, tc.want)
		assert.Equal(t, mode.SpellingMembership{}, got)
	}
}

func emptyNumbers(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	return values
}
func emptyPlaces(values []mode.DegreePlace) []mode.DegreePlace {
	if len(values) == 0 {
		return nil
	}
	return values
}

func TestMatchPitches(t *testing.T) {
	t.Parallel()
	sys := system.Traditional()
	c4, c5 := readingPitch(sys, spelling.C(), 4), readingPitch(sys, spelling.C(), 5)
	bSharp3 := readingPitch(sys, spelling.BSharp(), 3)
	major := mode.MustFromCode("ionian")
	for _, tc := range []struct {
		name          string
		a, b          pitch.Pitch
		by            mode.PitchComparison
		window        mode.CycleWindow
		first, second []mode.DegreePlace
	}{
		{"registers", c4, c5, mode.WrittenPitch, mode.CycleWindow{0, 1}, []mode.DegreePlace{{1, 0}}, []mode.DegreePlace{{1, 1}}},
		{"only next cycle", c4, c5, mode.WrittenPitch, mode.CycleWindow{1, 1}, nil, []mode.DegreePlace{{1, 1}}},
		{"exact spelling", bSharp3, c4, mode.WrittenPitch, mode.CycleWindow{-1, 1}, nil, []mode.DegreePlace{{1, 0}}},
		{"enharmonic position", bSharp3, c4, mode.PositionalPitch, mode.CycleWindow{-1, 1}, []mode.DegreePlace{{1, 0}}, []mode.DegreePlace{{1, 0}}},
		{"first absent", readingPitch(sys, spelling.CSharp(), 4), c4, mode.PositionalPitch, mode.CycleWindow{}, nil, []mode.DegreePlace{{1, 0}}},
		{"both same", c5, c5, mode.WrittenPitch, mode.CycleWindow{-1, 2}, []mode.DegreePlace{{1, 1}}, []mode.DegreePlace{{1, 1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := major.MatchPitches(sys, c4, tc.a, tc.b, tc.by, tc.window)
			require.NoError(t, err)
			assert.Equal(t, tc.first, emptyPlaces(got.First))
			assert.Equal(t, tc.second, emptyPlaces(got.Second))
			assert.Equal(t, got, major.MustMatchPitches(sys, c4, tc.a, tc.b, tc.by, tc.window))
			if len(got.First) > 0 {
				got.First[0].Number = 999
			}
			again, err := major.MatchPitches(sys, c4, tc.a, tc.b, tc.by, tc.window)
			require.NoError(t, err)
			assert.Equal(t, tc.first, emptyPlaces(again.First))
		})
	}
	fourth := mode.MustClosedFromIntervals(intervals("M2 M2 m2"))
	got, err := fourth.MatchPitches(sys, c4, readingPitch(sys, spelling.F(), 4), readingPitch(sys, spelling.G(), 4), mode.WrittenPitch, mode.CycleWindow{0, 2})
	require.NoError(t, err)
	assert.Equal(t, []mode.DegreePlace{{1, 1}}, got.First)
	assert.Equal(t, []mode.DegreePlace{{2, 1}}, got.Second)
	open := mode.MustOpenFromIntervals(intervals("M2 m2"))
	got, err = open.MatchPitches(sys, c4, c4, c5, mode.WrittenPitch, mode.CycleWindow{1, 2})
	require.NoError(t, err)
	assert.Empty(t, got.First)
	assert.Empty(t, got.Second)
	// A duplicate spelling is still a distinct written pitch.
	octaves := mode.MustOpenFromIntervals(intervals("P8 P8"))
	got, err = octaves.MatchPitches(sys, c4, c5, readingPitch(sys, spelling.C(), 6), mode.WrittenPitch, mode.CycleWindow{})
	require.NoError(t, err)
	assert.Equal(t, []mode.DegreePlace{{2, 0}}, got.First)
	assert.Equal(t, []mode.DegreePlace{{3, 0}}, got.Second)
	half := pitch.MustFromASPN(spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2)), 4)
	for _, tc := range []struct {
		m           mode.Mode
		sys         *system.System
		tonic, a, b pitch.Pitch
		by          mode.PitchComparison
		window      mode.CycleWindow
		want        error
	}{
		{mode.Mode{}, sys, c4, c4, c4, mode.WrittenPitch, mode.CycleWindow{}, mode.ErrUnset},
		{major, nil, c4, c4, c4, mode.WrittenPitch, mode.CycleWindow{}, mode.ErrUnset},
		{major, sys, pitch.Pitch{}, c4, c4, mode.WrittenPitch, mode.CycleWindow{}, mode.ErrUnset},
		{major, sys, c4, pitch.Pitch{}, c4, mode.WrittenPitch, mode.CycleWindow{}, mode.ErrUnset},
		{major, sys, c4, c4, pitch.Pitch{}, mode.WrittenPitch, mode.CycleWindow{}, mode.ErrUnset},
		{major, sys, c4, c4, c4, 0, mode.CycleWindow{}, mode.ErrInvalidArgument},
		{major, sys, c4, c4, c4, 99, mode.CycleWindow{}, mode.ErrInvalidArgument},
		{major, sys, c4, c4, c4, mode.WrittenPitch, mode.CycleWindow{2, 1}, mode.ErrInvalidArgument},
		{major, sys, c4, half, c4, mode.WrittenPitch, mode.CycleWindow{}, system.ErrInadmissible},
		{major, sys, c4, c4, half, mode.PositionalPitch, mode.CycleWindow{}, system.ErrInadmissible},
		{open, sys, c4, half, c4, mode.WrittenPitch, mode.CycleWindow{1, 2}, system.ErrInadmissible},
	} {
		got, err := tc.m.MatchPitches(tc.sys, tc.tonic, tc.a, tc.b, tc.by, tc.window)
		assert.ErrorIs(t, err, tc.want)
		assert.Equal(t, mode.PitchMembership{}, got)
	}
}

// The cycle window limits which degrees are built, including at register
// boundaries. Unrequested cycles must not create an overflow error.
func TestMatchPitchesOnlyBuildsRequestedExtremeCycle(t *testing.T) {
	t.Parallel()
	major := mode.MustFromCode("ionian")
	sys := system.Traditional()
	tonic := pitch.MustFromASPN(spelling.B(), math.MaxInt)
	a := pitch.MustFromASPN(spelling.CSharp(), math.MaxInt)
	b := pitch.MustFromASPN(spelling.DSharp(), math.MaxInt)
	// B major in cycle -1 is B/(MaxInt-1), C-sharp, D-sharp, E,
	// F-sharp, G-sharp, A-sharp in MaxInt. Every requested degree fits.
	window := mode.CycleWindow{First: -1, Last: -1}
	for _, tc := range []struct {
		name string
		by   mode.PitchComparison
	}{
		{"written", mode.WrittenPitch},
		{"positional", mode.PositionalPitch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := major.MatchPitches(sys, tonic, a, b, tc.by, window)
			require.NoError(t, err)
			assert.Equal(t, []mode.DegreePlace{{Number: 2, Cycle: -1}}, got.First)
			assert.Equal(t, []mode.DegreePlace{{Number: 3, Cycle: -1}}, got.Second)
		})
	}
}
