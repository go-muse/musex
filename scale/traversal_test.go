package scale_test

import (
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O-12...16, K-2/3/6/15/16/21/22/23/34/35/42/43/44.
func TestCircularMovement(t *testing.T) {
	t.Parallel()
	ns := []note.Note{at("G", 4, cents.Zero(), nil), at("A", 4, cents.Zero(), nil), at("B", 4, cents.Zero(), nil)}
	s := scale.MustNew(ns...)
	for _, tt := range []struct{ start, distance, want int }{{2, 1, 0}, {0, -1, 2}, {1, 0, 1}, {0, 6, 0}, {0, -6, 0}, {2, math.MaxInt, 0}, {2, math.MinInt, 0}} {
		got, err := s.Move(tt.start, tt.distance)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got.Index)
		checkNotes(t, []note.Note{ns[tt.want]}, []note.Note{got.Note})
	}
	next, err := s.Next(2)
	require.NoError(t, err)
	assert.Equal(t, 0, next.Index)
	assert.True(t, ns[0].Equal(next.Note))
	prev, err := s.Previous(0)
	require.NoError(t, err)
	assert.Equal(t, 2, prev.Index)
	for i := range 3 {
		for j := range 3 {
			f, err := s.DistanceForward(i, j)
			require.NoError(t, err)
			assert.Equal(t, (j-i+3)%3, f)
			b, err := s.DistanceBackward(i, j)
			require.NoError(t, err)
			assert.Equal(t, (i-j+3)%3, b)
			reflected, err := s.MirrorIndex(i, j)
			require.NoError(t, err)
			again, err := s.MirrorIndex(reflected.Index, j)
			require.NoError(t, err)
			assert.Equal(t, i, again.Index)
		}
	}
	for _, pair := range [][2]int{{-1, 0}, {3, 0}, {0, -1}, {0, 3}} {
		_, err := s.MirrorIndex(pair[0], pair[1])
		assert.ErrorIs(t, err, scale.ErrIndex)
	}
	_, err = s.Move(3, 0)
	assert.ErrorIs(t, err, scale.ErrIndex)
	_, err = s.DistanceForward(0, 3)
	assert.ErrorIs(t, err, scale.ErrIndex)
	_, err = s.DistanceBackward(-1, 0)
	assert.ErrorIs(t, err, scale.ErrIndex)
	single := scale.MustNew(ns[0])
	got, err := single.Move(0, math.MinInt)
	require.NoError(t, err)
	assert.Zero(t, got.Index)
	for _, distance := range []int{math.MinInt, math.MaxInt} {
		got, err := single.Take(0, 5, scale.Forward)
		require.NoError(t, err)
		checkScale(t, []note.Note{ns[0], ns[0], ns[0], ns[0], ns[0]}, got)
		_, err = single.Move(0, distance)
		require.NoError(t, err)
	}
	duplicate := list("G A G")
	_, err = duplicate.MoveFrom(scale.Selection{Note: names("G")[0]}, 1)
	assert.ErrorIs(t, err, scale.ErrAmbiguous)
	index := 2
	got, err = duplicate.MoveFrom(scale.Selection{Note: names("G")[0], Index: &index}, 1)
	require.NoError(t, err)
	assert.Zero(t, got.Index)
	for _, tt := range []struct {
		selection scale.Selection
		want      error
	}{{scale.Selection{Note: names("F")[0]}, scale.ErrNotFound}, {scale.Selection{}, scale.ErrUnset}, {scale.Selection{Note: names("A")[0], Index: new(2)}, scale.ErrNotFound}, {scale.Selection{Note: names("G")[0], Index: new(3)}, scale.ErrIndex}} {
		_, err = duplicate.MoveFrom(tt.selection, 1)
		assert.ErrorIs(t, err, tt.want)
	}
}

func TestTraversalEmptyAndValidation(t *testing.T) {
	t.Parallel()
	empty := scale.MustNew()
	nonempty := list("G A B")
	for _, s := range []*scale.Scale{empty, nonempty} {
		got, err := s.Take(math.MinInt, 0, scale.Forward)
		require.NoError(t, err)
		checkScale(t, names(""), got)
		seq, err := s.Iter(math.MaxInt, 0, scale.Backward)
		require.NoError(t, err)
		require.NotNil(t, seq)
		assert.Empty(t, collect(seq, 1))
		_, err = s.Take(0, 0, scale.DirectionUnset)
		assert.ErrorIs(t, err, scale.ErrInvalidArgument)
		_, err = s.Iter(0, -1, scale.Forward)
		assert.ErrorIs(t, err, scale.ErrInvalidArgument)
	}
	round, err := empty.OneRound(-100, scale.Forward)
	require.NoError(t, err)
	checkScale(t, names(""), round)
	seq, err := empty.Round(100, scale.Backward)
	require.NoError(t, err)
	assert.Empty(t, collect(seq, 1))
	for _, f := range []func() error{
		func() error { _, e := empty.Take(100, 3, scale.Forward); return e }, func() error { _, e := empty.Iter(100, 3, scale.Forward); return e }, func() error { _, e := empty.Cycle(100, scale.Forward); return e }, func() error { _, e := empty.Move(100, 1); return e }, func() error { _, e := empty.Next(100); return e }, func() error { _, e := empty.Previous(100); return e }, func() error { _, e := empty.DistanceForward(0, 1); return e }, func() error { _, e := empty.DistanceBackward(0, 1); return e }, func() error { _, e := empty.MirrorIndex(0, 3); return e },
	} {
		assert.ErrorIs(t, f(), scale.ErrEmpty)
	}
	for _, direction := range []scale.Direction{scale.Forward, scale.Backward} {
		_, err := nonempty.Cycle(3, direction)
		assert.ErrorIs(t, err, scale.ErrIndex)
	}
	for _, d := range []scale.Direction{scale.DirectionUnset, 2, -2, 127, -128} {
		assert.False(t, d.IsValid())
		_, err := nonempty.Round(0, d)
		assert.ErrorIs(t, err, scale.ErrInvalidArgument)
	}
	assert.True(t, scale.Forward.IsValid())
	assert.True(t, scale.Backward.IsValid())
	assert.Equal(t, "forward", scale.Forward.String())
	assert.Equal(t, "backward", scale.Backward.String())
	assert.Equal(t, "unset", scale.DirectionUnset.String())
	assert.Equal(t, "Direction(-2)", scale.Direction(-2).String())
}

func TestIteratorSnapshotReentrancyAndEarlyBreak(t *testing.T) {
	t.Parallel()
	s := list("G A B")
	cycle, err := s.Cycle(1, scale.Forward)
	require.NoError(t, err)
	seq, err := s.Iter(1, 5, scale.Forward)
	require.NoError(t, err)
	round, err := s.Round(1, scale.Backward)
	require.NoError(t, err)
	huge, err := s.Iter(0, math.MaxInt, scale.Forward)
	require.NoError(t, err)
	require.NoError(t, s.Replace(names("C D")...)) // Snapshot is taken before the first range.
	checkNotes(t, names("A B G A B"), collect(seq, 100))
	checkNotes(t, names("A G B"), collect(round, 100))
	checkNotes(t, names("G A"), collect(huge, 2))
	checkNotes(t, names("A B G A"), collect(cycle, 4))
	checkNotes(t, names("A B G A B"), collect(seq, 100))
	for n := range cycle {
		n.MustSetCents(cents.MustFromInt(77))
		break
	}
	checkNotes(t, names("A B G"), collect(cycle, 3))
	fresh, err := s.Cycle(0, scale.Forward)
	require.NoError(t, err)
	checkNotes(t, names("C D C D"), collect(fresh, 4))
	var wg sync.WaitGroup
	results := make(chan []note.Note, 8)
	for range 8 {
		wg.Go(func() { results <- collect(seq, 100) })
	}
	wg.Wait()
	close(results)
	for result := range results {
		checkNotes(t, names("A B G A B"), result)
	}
	take, err := list("G A B").Take(1, 5, scale.Forward)
	require.NoError(t, err)
	checkScale(t, names("A B G A B"), take)
	take, err = list("G A B").OneRound(1, scale.Backward)
	require.NoError(t, err)
	checkScale(t, names("A G B"), take)
}

func TestMelodySelection(t *testing.T) {
	t.Parallel()
	s := list("C D E F G A B")
	for _, tt := range []struct {
		name, in, want string
		distance       int
	}{{"K17", "D F", "E G", 1}, {"K18", "B D", "C E", 1}, {"negative", "C E", "B D", -1}} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.ShiftMelody(names(tt.in), tt.distance)
			require.NoError(t, err)
			checkScale(t, names(tt.want), got)
		})
	}
	mirror, err := s.MirrorMelody(names("D F"), 0)
	require.NoError(t, err)
	checkScale(t, names("B G"), mirror)
	again, err := s.MirrorMelody(mirror.Notes(), 0)
	require.NoError(t, err)
	checkScale(t, names("D F"), again)
	input := []note.Note{at("D", 4, cents.MustFromInt(30), settings.Standard())}
	got, err := s.ShiftMelody(input, 1, scale.ByName())
	require.NoError(t, err)
	checkScale(t, names("E"), got)
	assert.True(t, input[0].HasOctave())
	duplicates := list("G A G")
	selected := []scale.Selection{{Note: names("G")[0], Index: new(2)}, {Note: names("A")[0]}}
	got, err = duplicates.ShiftSelected(selected, 1)
	require.NoError(t, err)
	checkScale(t, names("G G"), got)
	got, err = duplicates.MirrorSelected(selected, 0)
	require.NoError(t, err)
	checkScale(t, names("A G"), got)
	for _, f := range []func() (*scale.Scale, error){func() (*scale.Scale, error) { return duplicates.ShiftMelody(names("A G"), 1) }, func() (*scale.Scale, error) { return duplicates.MirrorMelody(names("A G"), 0) }} {
		got, err := f()
		assert.Nil(t, got)
		assert.ErrorIs(t, err, scale.ErrAmbiguous)
	}
	got, err = s.ShiftMelody(names("D G+1"), 1)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, scale.ErrNotFound)
	got, err = s.MirrorMelody(nil, 7)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, scale.ErrIndex)
	empty := scale.MustNew()
	got, err = empty.ShiftSelected(nil, 1)
	require.NoError(t, err)
	checkScale(t, names(""), got)
	_, err = empty.MirrorSelected(nil, 0)
	assert.ErrorIs(t, err, scale.ErrEmpty)
	_, err = s.ShiftMelody(nil, 0, scale.ByName(), scale.ByRecord())
	assert.ErrorIs(t, err, scale.ErrInvalidArgument)
	checkScale(t, names("C D E F G A B"), s)
}
