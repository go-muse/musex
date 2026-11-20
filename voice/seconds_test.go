package voice_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type precision struct {
	places int
	mode   rounding.Mode
}

// Any valid precision gives the same exact result (ИК-24 tempomap).
func precisions() []precision {
	return []precision{{0, rounding.Down}, {3, rounding.HalfUp}, {1000, rounding.HalfEven}}
}

type span struct {
	index      int
	start, end tempo.Seconds
}

func assertTimings(t *testing.T, want []span, got []voice.Timing) {
	t.Helper()
	require.Len(t, got, len(want))
	for i, w := range want {
		assert.Equal(t, w.index, got[i].Index)
		assertExactSeconds(t, w.start, got[i].Start)
		assertExactSeconds(t, w.end, got[i].End)
	}
}

// assertSameTimings compares timings by index and SecondsResult.Equal.
func assertSameTimings(t *testing.T, want, got []voice.Timing) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		assert.Equal(t, want[i].Index, got[i].Index)
		assert.True(t, want[i].Start.Equal(got[i].Start), "start %d: want %v, got %v", i, want[i].Start, got[i].Start)
		assert.True(t, want[i].End.Equal(got[i].End), "end %d: want %v, got %v", i, want[i].End, got[i].End)
	}
}

// heldLine is the line of КП-25: C4 quarter, D4 half with a fermata, E4
// quarter.
func heldLine() *voice.Line {
	return line(snd(c4(), q()), snd(d4(), h()).MustWithFermata(true), snd(e4(), q()))
}

// КП-25: the next sound starts after the hold at the end of the previous
// one; the fermata itself adds nothing.
func TestSoundSecondsWithHold(t *testing.T) {
	t.Parallel()
	want := []span{{0, seconds(0, 1), seconds(1, 1)}, {1, seconds(1, 1), seconds(5, 1)}, {2, seconds(5, 1), seconds(6, 1)}}
	for _, p := range precisions() {
		t.Run(p.mode.String(), func(t *testing.T) {
			t.Parallel()
			l := heldLine()
			got, err := l.SoundSeconds(map60Hold(), p.places, p.mode)
			require.NoError(t, err)
			assertTimings(t, want, got)
			assertSameTimings(t, got, l.MustSoundSeconds(map60Hold(), p.places, p.mode))
			from, err := l.SoundSecondsFrom(map60Hold(), duration.Origin(), p.places, p.mode)
			require.NoError(t, err)
			assertSameTimings(t, got, from)
		})
	}

	// П-4: without a hold in the map the fermata does not change the times.
	got := heldLine().MustSoundSeconds(map60(), 3, rounding.HalfUp)
	assertTimings(t, []span{{0, seconds(0, 1), seconds(1, 1)}, {1, seconds(1, 1), seconds(3, 1)}, {2, seconds(3, 1), seconds(4, 1)}}, got)

	fermatas := heldLine().MustFermatas()
	require.Len(t, fermatas, 1)
	assert.Equal(t, 1, fermatas[0].Index)
	assert.Equal(t, "3/4", fermatas[0].End.String())
}

// КП-26: the start of the line on the map shifts every position.
func TestSoundSecondsFrom(t *testing.T) {
	t.Parallel()
	want := []span{{0, seconds(4, 1), seconds(5, 1)}, {1, seconds(5, 1), seconds(7, 1)}, {2, seconds(7, 1), seconds(8, 1)}}
	for _, p := range precisions() {
		t.Run(p.mode.String(), func(t *testing.T) {
			t.Parallel()
			l := heldLine()
			got, err := l.SoundSecondsFrom(map60(), pos(1, 1), p.places, p.mode)
			require.NoError(t, err)
			assertTimings(t, want, got)
			assertSameTimings(t, got, l.MustSoundSecondsFrom(map60(), pos(1, 1), p.places, p.mode))
		})
	}
}

// Rests are skipped: the index of a timing is the index of the element.
func TestSoundSecondsSkipRests(t *testing.T) {
	t.Parallel()
	got := kp4().MustSoundSeconds(map60(), 0, rounding.Down)
	assertTimings(t, []span{
		{0, seconds(0, 1), seconds(3, 2)},
		{1, seconds(3, 2), seconds(2, 1)},
		{3, seconds(3, 1), seconds(4, 1)},
	}, got)
}

// КП-48, ОП-16: seconds of one element, a sound or a rest.
func TestElementSeconds(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), rst(q()))
	for _, p := range precisions() {
		t.Run(p.mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := l.ElementSeconds(1, map60(), p.places, p.mode)
			require.NoError(t, err)
			assertTimings(t, []span{{1, seconds(1, 1), seconds(2, 1)}}, []voice.Timing{got})
			assertSameTimings(t, []voice.Timing{got}, []voice.Timing{l.MustElementSeconds(1, map60(), p.places, p.mode)})

			first, err := l.ElementSeconds(0, map60(), p.places, p.mode)
			require.NoError(t, err)
			assertTimings(t, []span{{0, seconds(0, 1), seconds(1, 1)}}, []voice.Timing{first})

			from, err := l.ElementSecondsFrom(1, map60(), pos(1, 1), p.places, p.mode)
			require.NoError(t, err)
			assertTimings(t, []span{{1, seconds(5, 1), seconds(6, 1)}}, []voice.Timing{from})
			assertSameTimings(t, []voice.Timing{from}, []voice.Timing{l.MustElementSecondsFrom(1, map60(), pos(1, 1), p.places, p.mode)})
		})
	}

	held := heldLine().MustElementSeconds(1, map60Hold(), 3, rounding.HalfUp)
	assertTimings(t, []span{{1, seconds(1, 1), seconds(5, 1)}}, []voice.Timing{held})
}

// КП-40: a gradual change gives an approximate result with the bound of
// tempomap.
func TestSoundSecondsApproximate(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), w()))
	got, err := l.SoundSeconds(mapGradual(), 3, rounding.HalfUp)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 0, got[0].Index)
	assertExactSeconds(t, seconds(0, 1), got[0].Start)

	end := got[0].End
	value, err := end.Value()
	require.NoError(t, err)
	assert.True(t, seconds(2773, 1000).Equal(value), "КП-40: about 2.773 s, got %v", end)
	approximate, err := end.Approximate()
	require.NoError(t, err)
	assert.True(t, approximate)
	bound, err := end.ErrorBound()
	require.NoError(t, err)
	assert.True(t, seconds(1, 2000).Equal(bound), "КП-40: bound 0.0005 s, got %v", bound)
	places, ok, err := end.Places()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3, places)

	direct := mapGradual().MustEnd(at(1, 1), 3, rounding.HalfUp)
	assert.True(t, direct.Equal(end), "the result is the one of tempomap: want %v, got %v", direct, end)

	element := l.MustElementSeconds(0, mapGradual(), 3, rounding.HalfUp)
	assert.True(t, direct.Equal(element.End))
}

// КП-41 and the order of checks of architecture 6.7.
func TestSoundSecondsFailures(t *testing.T) {
	t.Parallel()
	whole := line(snd(c4(), w()))
	var unset *voice.Line
	tests := []struct {
		name   string
		l      *voice.Line
		m      tempomap.Map
		places int
		mode   rounding.Mode
		want   want
	}{
		{"КП-41 unfinished gradual change", whole, mapUnfinished0(), 3, rounding.HalfUp, fails(tempomap.ErrUnfinished)},
		{"КП-41 map by measures", whole, mapMeasured(), 3, rounding.HalfUp, fails(voice.ErrMeasureMap).only()},
		{"КП-41 places 1001", whole, map60(), 1001, rounding.HalfUp, fails(tempomap.ErrInvalidArgument)},
		{"negative places", whole, map60(), -1, rounding.HalfUp, fails(tempomap.ErrInvalidArgument)},
		{"КП-41 unset mode", whole, map60(), 3, rounding.Mode(0), fails(tempomap.ErrInvalidArgument, rounding.ErrUnset)},
		{"unknown mode", whole, map60(), 3, rounding.Mode(42), fails(tempomap.ErrInvalidArgument, rounding.ErrUnknownMode)},
		{"КП-41 unset map", whole, tempomap.Map{}, 3, rounding.HalfUp, fails(voice.ErrUnset, tempomap.ErrUnset)},
		{"unset map before precision", whole, tempomap.Map{}, 1001, rounding.Mode(0), fails(voice.ErrUnset, tempomap.ErrUnset).only()},
		{"map by measures before precision", whole, mapMeasured(), 1001, rounding.Mode(0), fails(voice.ErrMeasureMap).only()},
		{"unset line first", unset, tempomap.Map{}, 1001, rounding.Mode(0), fails(voice.ErrUnset).only()},
		{"precision is checked without sounds", line(rst(q())), map60(), 1001, rounding.HalfUp, fails(tempomap.ErrInvalidArgument)},
		{"precision is checked on the empty line", line(), map60(), 3, rounding.Mode(0), fails(tempomap.ErrInvalidArgument, rounding.ErrUnset)},
		{"map is checked on the empty line", line(), mapUnfinished0(), 3, rounding.HalfUp, fails(tempomap.ErrUnfinished)},
		{"failure of a sound", line(snd(c4(), q()), snd(d4(), q()), snd(e4(), h())), mapUnfinishedHalf(), 3, rounding.HalfUp,
			failsAt(1, tempomap.ErrUnfinished)},
		{"failure of a sound after a rest", line(snd(c4(), q()), rst(q()), snd(e4(), h())), mapUnfinishedHalf(), 3, rounding.HalfUp,
			failsAt(2, tempomap.ErrUnfinished)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() ([]voice.Timing, error) { return tt.l.SoundSeconds(tt.m, tt.places, tt.mode) },
				func() []voice.Timing { return tt.l.MustSoundSeconds(tt.m, tt.places, tt.mode) })
			assertFailure(t, tt.want,
				func() ([]voice.Timing, error) {
					return tt.l.SoundSecondsFrom(tt.m, duration.Origin(), tt.places, tt.mode)
				},
				func() []voice.Timing {
					return tt.l.MustSoundSecondsFrom(tt.m, duration.Origin(), tt.places, tt.mode)
				})
		})
	}
}

func TestSoundSecondsFromFailures(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()))
	tests := []struct {
		name  string
		m     tempomap.Map
		start duration.Position
		want  want
	}{
		{"unset start", map60(), duration.Position{}, fails(voice.ErrUnset, duration.ErrUnset)},
		{"unset map before start", tempomap.Map{}, duration.Position{}, fails(voice.ErrUnset, tempomap.ErrUnset).without(duration.ErrUnset)},
		{"unset start before map by measures", mapMeasured(), duration.Position{}, fails(voice.ErrUnset, duration.ErrUnset)},
		{"unfinished change at the start of the line", mapUnfinishedHalf(), pos(1, 2), fails(tempomap.ErrUnfinished)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() ([]voice.Timing, error) { return l.SoundSecondsFrom(tt.m, tt.start, 3, rounding.HalfUp) },
				func() []voice.Timing { return l.MustSoundSecondsFrom(tt.m, tt.start, 3, rounding.HalfUp) })
		})
	}

	// The position of a sound on the map is out of range: the failure of
	// that sound.
	invP, _, invQ := overflowDurations()
	start := duration.Origin().MustAdd(invP)
	far := line(snd(c4(), voice.MustValueDuration(invQ)))
	assertFailure(t, failsAt(0, duration.ErrOutOfRange, fraction.ErrOutOfRange),
		func() ([]voice.Timing, error) { return far.SoundSecondsFrom(map60(), start, 3, rounding.HalfUp) },
		func() []voice.Timing { return far.MustSoundSecondsFrom(map60(), start, 3, rounding.HalfUp) })
}

func TestElementSecondsFailures(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), snd(d4(), q()), snd(e4(), h()))
	var unset *voice.Line
	tests := []struct {
		name   string
		l      *voice.Line
		i      int
		m      tempomap.Map
		places int
		mode   rounding.Mode
		want   want
	}{
		{"index out of range", l, 3, map60(), 3, rounding.HalfUp, fails(voice.ErrIndex).only()},
		{"negative index", l, -1, map60(), 3, rounding.HalfUp, fails(voice.ErrIndex).only()},
		{"index before the map", l, 5, tempomap.Map{}, 3, rounding.HalfUp, fails(voice.ErrIndex).only()},
		{"unset map", l, 0, tempomap.Map{}, 3, rounding.HalfUp, fails(voice.ErrUnset, tempomap.ErrUnset)},
		{"map by measures", l, 0, mapMeasured(), 1001, rounding.HalfUp, fails(voice.ErrMeasureMap).only()},
		{"places", l, 0, map60(), 1001, rounding.HalfUp, fails(tempomap.ErrInvalidArgument)},
		{"mode", l, 0, map60(), 3, rounding.Mode(0), fails(tempomap.ErrInvalidArgument, rounding.ErrUnset)},
		{"failure of the element is not wrapped", l, 1, mapUnfinishedHalf(), 3, rounding.HalfUp, fails(tempomap.ErrUnfinished)},
		{"unset line first", unset, -1, tempomap.Map{}, 3, rounding.HalfUp, fails(voice.ErrUnset).only()},
		{"empty line has no element", line(), 0, map60(), 3, rounding.HalfUp, fails(voice.ErrIndex).only()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (voice.Timing, error) { return tt.l.ElementSeconds(tt.i, tt.m, tt.places, tt.mode) },
				func() voice.Timing { return tt.l.MustElementSeconds(tt.i, tt.m, tt.places, tt.mode) })
			assertFailure(t, tt.want,
				func() (voice.Timing, error) {
					return tt.l.ElementSecondsFrom(tt.i, tt.m, duration.Origin(), tt.places, tt.mode)
				},
				func() voice.Timing {
					return tt.l.MustElementSecondsFrom(tt.i, tt.m, duration.Origin(), tt.places, tt.mode)
				})
		})
	}

	t.Run("start", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, fails(voice.ErrUnset, duration.ErrUnset),
			func() (voice.Timing, error) {
				return l.ElementSecondsFrom(0, map60(), duration.Position{}, 3, rounding.HalfUp)
			},
			func() voice.Timing {
				return l.MustElementSecondsFrom(0, map60(), duration.Position{}, 3, rounding.HalfUp)
			})
		assertFailure(t, fails(voice.ErrIndex).only(),
			func() (voice.Timing, error) {
				return l.ElementSecondsFrom(9, map60(), duration.Position{}, 3, rounding.HalfUp)
			},
			func() voice.Timing {
				return l.MustElementSecondsFrom(9, map60(), duration.Position{}, 3, rounding.HalfUp)
			})
		assertFailure(t, fails(voice.ErrUnset, tempomap.ErrUnset).without(duration.ErrUnset),
			func() (voice.Timing, error) {
				return l.ElementSecondsFrom(0, tempomap.Map{}, duration.Position{}, 3, rounding.HalfUp)
			},
			func() voice.Timing {
				return l.MustElementSecondsFrom(0, tempomap.Map{}, duration.Position{}, 3, rounding.HalfUp)
			})
		assertFailure(t, fails(voice.ErrUnset, duration.ErrUnset),
			func() (voice.Timing, error) {
				return l.ElementSecondsFrom(0, mapMeasured(), duration.Position{}, 3, rounding.HalfUp)
			},
			func() voice.Timing {
				return l.MustElementSecondsFrom(0, mapMeasured(), duration.Position{}, 3, rounding.HalfUp)
			})
	})

	t.Run("overflow is not wrapped", func(t *testing.T) {
		t.Parallel()
		invP, _, invQ := overflowDurations()
		start := duration.Origin().MustAdd(invP)
		far := line(snd(c4(), voice.MustValueDuration(invQ)))
		assertFailure(t, overflowWant(),
			func() (voice.Timing, error) { return far.ElementSecondsFrom(0, map60(), start, 3, rounding.HalfUp) },
			func() voice.Timing { return far.MustElementSecondsFrom(0, map60(), start, 3, rounding.HalfUp) })
	})
}

// A line without sounds gives an empty, non-nil list.
func TestSoundSecondsWithoutSounds(t *testing.T) {
	t.Parallel()
	for _, l := range []*voice.Line{line(), line(rst(q()))} {
		got, err := l.SoundSeconds(map60(), 3, rounding.HalfUp)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Empty(t, got)
	}
}
