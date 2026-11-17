package mode_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeKindsAndZero(t *testing.T) {
	t.Parallel()
	var m mode.Mode
	assert.False(t, m.IsValid())
	assert.False(t, m.IsClosed())
	assert.False(t, m.IsOpen())
	assert.False(t, m.IsOctave())
	assert.False(t, m.IsSecondwise())
	assert.False(t, m.IsPalindrome())
	assert.False(t, m.HasOrigin())
	assert.False(t, m.Changed())
	assert.Zero(t, m.Len())
	assert.Equal(t, mode.KindUnset, m.Kind())
	assert.Nil(t, m.AdjacentIntervals())
	assert.Nil(t, m.DegreeIntervals())
	assert.False(t, m.Template().IsValid())
	cycle, ok := m.Cycle()
	assert.Zero(t, cycle)
	assert.False(t, ok)
	origin, ok := m.Origin()
	assert.Zero(t, origin)
	assert.False(t, ok)
	name, ok := m.SelectedName()
	assert.Empty(t, name)
	assert.False(t, ok)

	closed := mode.MustParse("P8")
	open := mode.MustParse("M2 |")
	assert.Equal(t, mode.Closed, closed.Kind())
	assert.Equal(t, mode.Open, open.Kind())
	assert.True(t, closed.IsClosed())
	assert.False(t, closed.IsOpen())
	assert.True(t, open.IsOpen())
	assert.False(t, open.IsClosed())
	assert.Equal(t, template.Closed, mode.Closed)
	assert.Equal(t, template.Open, mode.Open)
	assert.Equal(t, template.KindUnset, mode.KindUnset)
	assert.Equal(t, template.FromFirst, mode.FromFirst)
	assert.Equal(t, template.Adjacent, mode.Adjacent)
	assert.Equal(t, template.MeasureUnset, mode.MeasureUnset)
	assert.False(t, reflect.TypeFor[mode.Mode]().Comparable())
	assert.False(t, reflect.TypeFor[mode.Repository]().Comparable())
}

func TestOwnDegreesAndOutputOwnership(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ code, own string }{
		{"dorian", "P1 M2 m3 P4 P5 M6 m7"},
		{"harmonic-minor", "P1 M2 m3 P4 P5 m6 M7"},
		{"pentatonic-major", "P1 M2 M3 P5 M6"},
		{"whole-tone", "P1 M2 M3 A4 A5 A6"},
	} {
		t.Run(tt.code, func(t *testing.T) {
			m := mode.MustFromCode(tt.code)
			saved := m
			ds, err := m.Degrees()
			require.NoError(t, err)
			original := m.Template()
			checkMode(t, mode.MustFromTemplate(original), original.String())
			assert.True(t, mode.MustFromTemplate(original).Equal(m))
			require.Len(t, ds, m.Len())
			for i, d := range ds {
				checkDegrees(t, []degree.Degree{d}, i+1)
			}
			fifth, err := m.Degree(5)
			require.NoError(t, err)
			checkDegrees(t, []degree.Degree{fifth}, 5)
			own := m.DegreeIntervals()
			adj := m.AdjacentIntervals()
			origin := firstOf(m.Origin())
			assert.Equal(t, tt.own, intervalText(own))
			ds[0] = degree.MustNew(42)
			ds = append(ds, degree.MustNew(43))
			adj[0] = interval.Interval{}
			own[0] = interval.Interval{}
			names := origin.Names()
			names[0] = "overwritten"
			require.NoError(t, m.Rotate(2))
			assert.True(t, original.Equal(saved.Template()))
			assert.Equal(t, tt.own, intervalText(saved.DegreeIntervals()))
			assert.Equal(t, "P1", m.DegreeIntervals()[0].String())
			assert.True(t, origin.Equal(firstOf(m.Origin())))
			assert.Equal(t, 1, m.MustDegree(1).Number())
			assert.Equal(t, 42, ds[0].Number())
			assert.False(t, adj[0].IsValid())
		})
	}

	d, err := mode.MustFromCode("dorian").Degree(8)
	checkError(t, err, mode.ErrInvalidArgument)
	assert.Zero(t, d)
}

func TestDegreeIntervalsAndSegments(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code     string
		from, to int
		want     string
	}{
		{"dorian", 3, 6, "A4"}, {"dorian", 6, 3, "-A4"},
		{"harmonic-minor", 6, 7, "A2"}, {"harmonic-minor", 3, 7, "A5"}, {"harmonic-minor", 7, 10, "d4"},
		{"ionian", 4, 7, "A4"}, {"ionian", 7, 4, "-A4"}, {"ionian", 3, 3, "P1"}, {"ionian", 7, 8, "m2"},
		{"blues", 3, 4, "A1"}, {"blues", 4, 3, "-A1"},
	} {
		got, err := mode.MustFromCode(tt.code).BetweenDegrees(tt.from, tt.to)
		require.NoError(t, err)
		assert.True(t, interval.MustParse(tt.want).Equal(got), "%s %d to %d: %s", tt.code, tt.from, tt.to, got)
	}
	m := mode.MustFromCode("ionian")
	for _, tt := range []struct {
		number int
		want   string
	}{{8, "P8"}, {9, "M9"}, {15, "P15"}} {
		got, err := m.DegreeInterval(tt.number)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got.String())
	}
	edge, err := m.AdjacentIntervalAfter(9)
	require.NoError(t, err)
	assert.Equal(t, "M2", edge.String())
	for _, tt := range []struct {
		from, to int
		want     string
	}{
		{1, 4, "M2 M2 m2"}, {7, 9, "m2 M2"}, {1, 8, ionian}, {math.MaxInt - 2, math.MaxInt, "M2 M2"},
	} {
		got, err := m.Segment(tt.from, tt.to)
		require.NoError(t, err)
		assert.Equal(t, tt.want, intervalText(got))
		got[0] = interval.Interval{}
		assert.Equal(t, ionian, intervalText(m.AdjacentIntervals()))
	}
	for _, tt := range [][2]int{{1, 9}, {3, 3}, {4, 2}, {0, 1}, {math.MinInt, math.MaxInt}} {
		got, err := m.Segment(tt[0], tt[1])
		checkError(t, err, mode.ErrInvalidArgument)
		assert.Nil(t, got)
	}
	open := mode.MustParse("m3 M2 |")
	assert.Equal(t, "-P4", open.MustBetweenDegrees(3, 1).String())
	assert.Equal(t, "m3 M2", intervalText(open.MustSegment(1, 3)))
	for _, call := range []func() (interval.Interval, error){
		func() (interval.Interval, error) { return open.DegreeInterval(4) },
		func() (interval.Interval, error) { return open.AdjacentIntervalAfter(3) },
		func() (interval.Interval, error) { return open.BetweenDegrees(1, 4) },
		func() (interval.Interval, error) { return m.DegreeInterval(0) },
		func() (interval.Interval, error) { return m.AdjacentIntervalAfter(-1) },
		func() (interval.Interval, error) { return m.BetweenDegrees(2, 0) },
	} {
		got, err := call()
		checkError(t, err, mode.ErrInvalidArgument)
		assert.Zero(t, got)
	}
	segment, err := open.Segment(1, 4)
	checkError(t, err, mode.ErrInvalidArgument)
	assert.Nil(t, segment)

	singleton := mode.MustParse("P8")
	got, err := singleton.DegreeInterval(math.MaxInt)
	checkError(t, err, mode.ErrOutOfRange, template.ErrOutOfRange)
	assert.Zero(t, got)
	assert.Equal(t, "P8", singleton.MustBetweenDegrees(math.MaxInt-1, math.MaxInt).String())
	assert.Equal(t, "P1", singleton.MustBetweenDegrees(math.MaxInt, math.MaxInt).String())
}

func TestCircularDegreeNavigation(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("dorian")
	for _, tt := range []struct {
		name       string
		call       func(int) (degree.Degree, bool, error)
		from, want int
		wraps      bool
	}{
		{"next", m.Next, 5, 6, false}, {"next wraps", m.Next, 7, 1, true},
		{"previous", m.Previous, 5, 4, false}, {"previous wraps", m.Previous, 1, 7, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, wraps, err := tt.call(tt.from)
			require.NoError(t, err)
			assert.Equal(t, tt.wraps, wraps)
			checkDegrees(t, []degree.Degree{d}, tt.want)
		})
	}
	for _, tt := range []struct{ from, shift, number, cycles int }{
		{5, 9, 7, 1}, {1, -1, 7, -1}, {5, 0, 5, 0},
		{5, math.MinInt, 4, -1317624576693539401}, {5, math.MaxInt, 5, 1317624576693539401},
	} {
		d, cycles, err := m.Shift(tt.from, tt.shift)
		require.NoError(t, err)
		assert.Equal(t, tt.cycles, cycles)
		checkDegrees(t, []degree.Degree{d}, tt.number)
	}
	assert.Equal(t, 3, m.MustDistanceUp(3, 6))
	assert.Equal(t, 4, m.MustDistanceDown(3, 6))
	assert.Equal(t, -3, m.MustShortestDistance(6, 3))
	assert.Zero(t, m.MustShortestDistance(5, 5))
	assert.Equal(t, 3, mode.MustFromCode("blues").MustShortestDistance(1, 4))

	open := mode.MustParse("m3 M2 |")
	for _, tt := range []struct {
		m   mode.Mode
		n   int
		err error
	}{
		{open, 2, mode.ErrOpenMode}, {open, 4, mode.ErrInvalidArgument}, {m, 8, mode.ErrInvalidArgument},
	} {
		next, wrap, err := tt.m.Next(tt.n)
		checkError(t, err, tt.err)
		assert.Zero(t, next)
		assert.False(t, wrap)
		prev, wrap, err := tt.m.Previous(tt.n)
		checkError(t, err, tt.err)
		assert.Zero(t, prev)
		assert.False(t, wrap)
		shifted, cycles, err := tt.m.Shift(tt.n, 0)
		checkError(t, err, tt.err)
		assert.Zero(t, shifted)
		assert.Zero(t, cycles)
		for _, call := range []func(int, int) (int, error){tt.m.DistanceUp, tt.m.DistanceDown, tt.m.ShortestDistance} {
			got, err := call(1, tt.n)
			checkError(t, err, tt.err)
			assert.Zero(t, got)
		}
	}
}
