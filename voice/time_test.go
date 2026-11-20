package voice_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// КП-1, ОП-7…ОП-10: starts, ends and the length of the line.
func TestStartsAndLength(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		l      *voice.Line
		starts []string
		ends   []string
		end    string
	}{
		{"КП-1", kp1(), []string{"0/1", "1/4", "3/8", "1/2"}, []string{"1/4", "3/8", "1/2", "1/1"}, "1/1"},
		{"КП-4", kp4(), []string{"0/1", "3/8", "1/2", "3/4"}, []string{"3/8", "1/2", "3/4", "1/1"}, "1/1"},
		{"КП-9 empty", line(), []string{}, []string{}, "0/1"},
		{"КП-9 one rest", line(rst(q())), []string{"0/1"}, []string{"1/4"}, "1/4"},
		{"values", line(rst(val(1, 3)), snd(c4(), val(1, 6)), snd(d4(), w())), []string{"0/1", "1/3", "1/2"}, []string{"1/3", "1/2", "3/2"}, "3/2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			starts, end, err := tt.l.Starts()
			require.NoError(t, err)
			require.NotNil(t, starts)
			assert.Equal(t, tt.starts, positionStrings(starts))
			assert.Equal(t, tt.end, end.String())
			mustStarts, mustEnd := tt.l.MustStarts()
			assert.Equal(t, tt.starts, positionStrings(mustStarts))
			assert.Equal(t, tt.end, mustEnd.String())

			for i := range tt.starts {
				start, err := tt.l.StartOf(i)
				require.NoError(t, err)
				assert.Equal(t, tt.starts[i], start.String())
				assert.Equal(t, tt.starts[i], tt.l.MustStartOf(i).String())
				finish, err := tt.l.EndOf(i)
				require.NoError(t, err)
				assert.Equal(t, tt.ends[i], finish.String())
				assert.Equal(t, tt.ends[i], tt.l.MustEndOf(i).String())
			}

			length, err := tt.l.Length()
			require.NoError(t, err)
			assert.Equal(t, tt.end, length.String(), "the length is reduced and equals the end")
			assert.Equal(t, tt.end, tt.l.MustLength().String())
			last, err := tt.l.End()
			require.NoError(t, err)
			assert.Equal(t, tt.end, last.String())
			assert.Equal(t, tt.end, tt.l.MustEnd().String())
		})
	}
}

func TestStartOfFailures(t *testing.T) {
	t.Parallel()
	l := kp1()
	for _, i := range []int{-1, 4} {
		assertFailure(t, fails(voice.ErrIndex).only(),
			func() (duration.Position, error) { return l.StartOf(i) },
			func() duration.Position { return l.MustStartOf(i) })
		assertFailure(t, fails(voice.ErrIndex).only(),
			func() (duration.Position, error) { return l.EndOf(i) },
			func() duration.Position { return l.MustEndOf(i) })
	}
	var unset *voice.Line
	// The receiver is checked before the index.
	assertFailure(t, fails(voice.ErrUnset).only(),
		func() (duration.Position, error) { return unset.StartOf(-1) },
		func() duration.Position { return unset.MustStartOf(-1) })
	assertFailure(t, fails(voice.ErrUnset).only(),
		func() (duration.Position, error) { return unset.EndOf(-1) },
		func() duration.Position { return unset.MustEndOf(-1) })
	assertFailure(t, fails(voice.ErrUnset).only(),
		func() (pair[[]duration.Position, duration.Position], error) { return twoStarts(unset) },
		func() pair[[]duration.Position, duration.Position] { return mustTwo(unset.MustStarts()) })
}

func twoStarts(l *voice.Line) (pair[[]duration.Position, duration.Position], error) {
	starts, end, err := l.Starts()
	return pair[[]duration.Position, duration.Position]{A: starts, B: end}, err
}

// КП-2, ОП-11, СВ-41: a moment at the end of an element belongs to the next
// element; the end of the line and later give no element.
func TestAt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		l     *voice.Line
		t     duration.Position
		index int
		ok    bool
	}{
		{"КП-2 3/8", kp1(), pos(3, 8), 2, true},
		{"КП-2 1/3", kp1(), pos(1, 3), 1, true},
		{"КП-2 1/4", kp1(), pos(1, 4), 1, true},
		{"КП-2 0", kp1(), pos(0, 1), 0, true},
		{"КП-2 1", kp1(), pos(1, 1), 0, false},
		{"after the end", kp1(), pos(5, 1), 0, false},
		{"inside the last", kp1(), pos(99, 100), 3, true},
		{"empty line", line(), pos(0, 1), 0, false},
		{"rest", kp4(), pos(5, 8), 2, true},
		{"inside a tied sound", kp4(), pos(1, 3), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			index, ok, err := tt.l.At(tt.t)
			require.NoError(t, err)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.index, index)
			mustIndex, mustOK := tt.l.MustAt(tt.t)
			assert.Equal(t, tt.index, mustIndex)
			assert.Equal(t, tt.ok, mustOK)
		})
	}
}

// ОП-12, П-7, КП-4: a rest and a moment outside the line are silence.
func TestSoundAt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		t    duration.Position
		note note.Note
		ok   bool
	}{
		{"start of the tied sound", pos(0, 1), c4(), true},
		{"inside the tied sound", pos(1, 4), c4(), true},
		{"end of the tied sound", pos(3, 8), d4(), true},
		{"КП-4 rest at 5/8", pos(5, 8), note.Note{}, false},
		{"last sound", pos(3, 4), e4(), true},
		{"end of the line", pos(1, 1), note.Note{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := kp4()
			got, ok, err := l.SoundAt(tt.t)
			require.NoError(t, err)
			assert.Equal(t, tt.ok, ok)
			mustGot, mustOK := l.MustSoundAt(tt.t)
			assert.Equal(t, tt.ok, mustOK)
			if tt.ok {
				assertNote(t, tt.note, got)
				assertNote(t, tt.note, mustGot)
			} else {
				assert.Zero(t, got)
				assert.Zero(t, mustGot)
			}
		})
	}
}

func TestAtFailures(t *testing.T) {
	t.Parallel()
	l := kp1()
	assertFailure(t, fails(voice.ErrUnset, duration.ErrUnset),
		func() (pair[int, bool], error) { return two(l.At(duration.Position{})) },
		func() pair[int, bool] { return mustTwo(l.MustAt(duration.Position{})) })
	assertFailure(t, fails(voice.ErrUnset, duration.ErrUnset),
		func() (pair[note.Note, bool], error) { return two(l.SoundAt(duration.Position{})) },
		func() pair[note.Note, bool] { return mustTwo(l.MustSoundAt(duration.Position{})) })

	var unset *voice.Line
	// The receiver is checked before the moment.
	assertFailure(t, fails(voice.ErrUnset).only(),
		func() (pair[int, bool], error) { return two(unset.At(duration.Position{})) },
		func() pair[int, bool] { return mustTwo(unset.MustAt(duration.Position{})) })
	assertFailure(t, fails(voice.ErrUnset).only(),
		func() (pair[note.Note, bool], error) { return two(unset.SoundAt(duration.Position{})) },
		func() pair[note.Note, bool] { return mustTwo(unset.MustSoundAt(duration.Position{})) })
}

// КП-3, ОП-13, П-7: elements that intersect [a, b); at a = b the element at
// the moment a.
func TestSpan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		l    *voice.Line
		a, b duration.Position
		want []int
	}{
		{"КП-3 [1/4, 1/2)", kp1(), pos(1, 4), pos(1, 2), []int{1, 2}},
		{"КП-3 [3/8, 3/4)", kp1(), pos(3, 8), pos(3, 4), []int{2, 3}},
		{"КП-3 [1/2, 1/2]", kp1(), pos(1, 2), pos(1, 2), []int{3}},
		{"whole line", kp1(), pos(0, 1), pos(1, 1), []int{0, 1, 2, 3}},
		{"beyond the end", kp1(), pos(1, 2), pos(7, 1), []int{3}},
		{"after the end", kp1(), pos(1, 1), pos(2, 1), []int{}},
		{"at the end with a = b", kp1(), pos(1, 1), pos(1, 1), []int{}},
		{"at zero with a = b", kp1(), pos(0, 1), pos(0, 1), []int{0}},
		{"touching the start of an element", kp1(), pos(0, 1), pos(1, 4), []int{0}},
		{"inside one element", kp1(), pos(5, 8), pos(3, 4), []int{3}},
		{"empty line", line(), pos(0, 1), pos(1, 1), []int{}},
		{"rests are elements", kp4(), pos(1, 2), pos(3, 4), []int{2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.l.Span(tt.a, tt.b)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.l.MustSpan(tt.a, tt.b))
		})
	}
}

func TestSpanFailures(t *testing.T) {
	t.Parallel()
	l := kp1()
	var unset *voice.Line
	tests := []struct {
		name string
		l    *voice.Line
		a, b duration.Position
		want want
	}{
		{"КП-35 a > b", l, pos(1, 2), pos(1, 4), fails(voice.ErrIndex).only()},
		{"unset a", l, duration.Position{}, pos(1, 4), fails(voice.ErrUnset, duration.ErrUnset)},
		{"unset b", l, pos(1, 4), duration.Position{}, fails(voice.ErrUnset, duration.ErrUnset)},
		{"unset line first", unset, duration.Position{}, duration.Position{}, fails(voice.ErrUnset).only()},
		{"unset a before a > b", l, duration.Position{}, pos(0, 1), fails(voice.ErrUnset, duration.ErrUnset)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() ([]int, error) { return tt.l.Span(tt.a, tt.b) },
				func() []int { return tt.l.MustSpan(tt.a, tt.b) })
		})
	}
}

// ОП-14, КП-4: sounds with starts and ends in positions; a tied sound is
// one sound.
func TestSounds(t *testing.T) {
	t.Parallel()
	l := kp4()
	got, err := l.Sounds()
	require.NoError(t, err)
	want := []struct {
		index      int
		note       note.Note
		start, end string
	}{
		{0, c4(), "0/1", "3/8"},
		{1, d4(), "3/8", "1/2"},
		{3, e4(), "3/4", "1/1"},
	}
	require.Len(t, got, len(want))
	for i, w := range want {
		assert.Equal(t, w.index, got[i].Index)
		assertNote(t, w.note, got[i].Note)
		assert.Equal(t, w.start, got[i].Start.String())
		assert.Equal(t, w.end, got[i].End.String())
	}
	assert.Equal(t, got, l.MustSounds())

	for _, l := range []*voice.Line{line(), line(rst(q()), rst(h()))} {
		sounds, err := l.Sounds()
		require.NoError(t, err)
		require.NotNil(t, sounds)
		assert.Empty(t, sounds)
	}
}

// ОП-17, КП-25: elements with a fermata and the positions of their ends.
func TestFermatas(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), snd(d4(), h()).MustWithFermata(true), snd(e4(), q()))
	got, err := l.Fermatas()
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].Index)
	assert.Equal(t, "3/4", got[0].End.String())
	assert.Equal(t, got, l.MustFermatas())

	from, err := l.FermatasFrom(pos(1, 1))
	require.NoError(t, err)
	require.Len(t, from, 1)
	assert.Equal(t, 1, from[0].Index)
	assert.Equal(t, "7/4", from[0].End.String())
	assert.Equal(t, from, l.MustFermatasFrom(pos(1, 1)))

	zero, err := l.FermatasFrom(duration.Origin())
	require.NoError(t, err)
	assert.Equal(t, got, zero)

	several := line(rst(q()).MustWithFermata(true), snd(c4(), qe()), snd(d4(), e8()).MustWithFermata(true))
	all := several.MustFermatas()
	require.Len(t, all, 2)
	assert.Equal(t, 0, all[0].Index)
	assert.Equal(t, "1/4", all[0].End.String())
	assert.Equal(t, 2, all[1].Index)
	assert.Equal(t, "3/4", all[1].End.String())

	none, err := kp1().Fermatas()
	require.NoError(t, err)
	require.NotNil(t, none)
	assert.Empty(t, none)
	noneFrom, err := line().FermatasFrom(pos(1, 1))
	require.NoError(t, err)
	require.NotNil(t, noneFrom)
	assert.Empty(t, noneFrom)
}

func TestFermatasFromFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	assertFailure(t, fails(voice.ErrUnset).only(),
		func() ([]voice.Fermata, error) { return unset.FermatasFrom(duration.Position{}) },
		func() []voice.Fermata { return unset.MustFermatasFrom(duration.Position{}) })
	assertFailure(t, fails(voice.ErrUnset, duration.ErrUnset),
		func() ([]voice.Fermata, error) { return kp1().FermatasFrom(duration.Position{}) },
		func() []voice.Fermata { return kp1().MustFermatasFrom(duration.Position{}) })

	// The sum of the start and the end is out of range (architecture 6.6).
	invP, _, invQ := overflowDurations()
	l := line(snd(c4(), voice.MustValueDuration(invQ)).MustWithFermata(true))
	start := duration.Origin().MustAdd(invP)
	assertFailure(t, overflowWant(),
		func() ([]voice.Fermata, error) { return l.FermatasFrom(start) },
		func() []voice.Fermata { return l.MustFermatasFrom(start) })
}

// ИЛ-2: starts do not decrease, each start equals the end of the previous
// element, and the length is the end of the last element.
func TestPositionInvariants(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			starts, end := l.MustStarts()
			require.Len(t, starts, l.MustLen())
			previous := duration.Origin()
			for i, start := range starts {
				assert.True(t, previous.Equal(start), "start %d: want %v, got %v", i, previous, start)
				finish := l.MustEndOf(i)
				assert.Equal(t, 1, finish.Compare(start), "ИЛ-3: the length of element %d is positive", i)
				previous = finish
			}
			assert.True(t, previous.Equal(end))
			assert.True(t, end.MustFraction().Equal(l.MustLength()))
		})
	}
}
