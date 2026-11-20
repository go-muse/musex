package voice_test

import (
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ОП-32…ОП-40, П-14: every change returns a new line with the name and the
// roles of the receiver, which stays as it was.
func TestEdit(t *testing.T) {
	t.Parallel()
	c, d, e, f := snd(c4(), q()), snd(d4(), e8()), snd(e4(), e8()), snd(f4(), h())
	r := rst(q())
	tests := []struct {
		name string
		call func(l *voice.Line) (*voice.Line, error)
		must func(l *voice.Line) *voice.Line
		want *voice.Line
	}{
		{"Append", func(l *voice.Line) (*voice.Line, error) { return l.Append(r, c) },
			func(l *voice.Line) *voice.Line { return l.MustAppend(r, c) }, line(c, d, e, f, r, c)},
		{"Append nothing", func(l *voice.Line) (*voice.Line, error) { return l.Append() },
			func(l *voice.Line) *voice.Line { return l.MustAppend() }, line(c, d, e, f)},
		{"КП-49 Insert", func(l *voice.Line) (*voice.Line, error) { return l.Insert(1, r) },
			func(l *voice.Line) *voice.Line { return l.MustInsert(1, r) }, line(c, r, d, e, f)},
		{"Insert at the start", func(l *voice.Line) (*voice.Line, error) { return l.Insert(0, r, r) },
			func(l *voice.Line) *voice.Line { return l.MustInsert(0, r, r) }, line(r, r, c, d, e, f)},
		{"Insert at the end", func(l *voice.Line) (*voice.Line, error) { return l.Insert(4, r) },
			func(l *voice.Line) *voice.Line { return l.MustInsert(4, r) }, line(c, d, e, f, r)},
		{"Replace", func(l *voice.Line) (*voice.Line, error) { return l.Replace(2, r) },
			func(l *voice.Line) *voice.Line { return l.MustReplace(2, r) }, line(c, d, r, f)},
		{"Remove", func(l *voice.Line) (*voice.Line, error) { return l.Remove(0) },
			func(l *voice.Line) *voice.Line { return l.MustRemove(0) }, line(d, e, f)},
		{"Remove the last", func(l *voice.Line) (*voice.Line, error) { return l.Remove(3) },
			func(l *voice.Line) *voice.Line { return l.MustRemove(3) }, line(c, d, e)},
		{"RemoveRange", func(l *voice.Line) (*voice.Line, error) { return l.RemoveRange(1, 3) },
			func(l *voice.Line) *voice.Line { return l.MustRemoveRange(1, 3) }, line(c, f)},
		{"RemoveRange of nothing", func(l *voice.Line) (*voice.Line, error) { return l.RemoveRange(2, 2) },
			func(l *voice.Line) *voice.Line { return l.MustRemoveRange(2, 2) }, line(c, d, e, f)},
		{"RemoveRange of all", func(l *voice.Line) (*voice.Line, error) { return l.RemoveRange(0, 4) },
			func(l *voice.Line) *voice.Line { return l.MustRemoveRange(0, 4) }, line()},
		{"Slice", func(l *voice.Line) (*voice.Line, error) { return l.Slice(1, 3) },
			func(l *voice.Line) *voice.Line { return l.MustSlice(1, 3) }, line(d, e)},
		{"Slice of nothing", func(l *voice.Line) (*voice.Line, error) { return l.Slice(4, 4) },
			func(l *voice.Line) *voice.Line { return l.MustSlice(4, 4) }, line()},
		{"WithNote", func(l *voice.Line) (*voice.Line, error) { return l.WithNote(1, g4()) },
			func(l *voice.Line) *voice.Line { return l.MustWithNote(1, g4()) }, line(c, snd(g4(), e8()), e, f)},
		{"WithDuration", func(l *voice.Line) (*voice.Line, error) { return l.WithDuration(0, qe()) },
			func(l *voice.Line) *voice.Line { return l.MustWithDuration(0, qe()) }, line(snd(c4(), qe()), d, e, f)},
		{"WithFermata", func(l *voice.Line) (*voice.Line, error) { return l.WithFermata(3, true) },
			func(l *voice.Line) *voice.Line { return l.MustWithFermata(3, true) }, line(c, d, e, f.MustWithFermata(true))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := named(kp1())
			got, err := tt.call(src)
			require.NoError(t, err)
			want := named(tt.want)
			assertLine(t, want, got)
			assertLine(t, want, tt.must(src))
			assertLine(t, named(kp1()), src)
			assert.NotSame(t, src, got, "a new line is returned")
		})
	}
}

// КП-49: insertion, then removal of [0, 2); the source lines stay.
func TestInsertThenRemove(t *testing.T) {
	t.Parallel()
	src := kp1()
	inserted := src.MustInsert(1, rst(q()))
	assertLine(t, line(snd(c4(), q()), rst(q()), snd(d4(), e8()), snd(e4(), e8()), snd(f4(), h())), inserted)
	removed := inserted.MustRemoveRange(0, 2)
	assertLine(t, line(snd(d4(), e8()), snd(e4(), e8()), snd(f4(), h())), removed)
	assertLine(t, kp1(), src)
	assertLine(t, line(snd(c4(), q()), rst(q()), snd(d4(), e8()), snd(e4(), e8()), snd(f4(), h())), inserted)
	starts, end := removed.MustStarts()
	assert.Equal(t, []string{"0/1", "1/8", "1/4"}, positionStrings(starts), "starts are computed anew")
	assert.Equal(t, "3/4", end.String())
}

// КП-39, ОП-37, ОП-38: split and join back.
func TestSplitAndJoin(t *testing.T) {
	t.Parallel()
	src := named(kp1())
	left, right, err := src.Split(2)
	require.NoError(t, err)
	assertLine(t, named(line(snd(c4(), q()), snd(d4(), e8()))), left)
	assertLine(t, named(line(snd(e4(), e8()), snd(f4(), h()))), right)
	assert.Equal(t, "0/1", right.MustStartOf(0).String(), "a part starts at zero")
	ml, mr := src.MustSplit(2)
	assertLine(t, left, ml)
	assertLine(t, right, mr)
	assertLine(t, src, voice.MustJoin(left, right))

	first, rest := src.MustSplit(0)
	assertLine(t, named(line()), first)
	assertLine(t, src, rest)
	all, none := src.MustSplit(4)
	assertLine(t, src, all)
	assertLine(t, named(line()), none)
}

// ОП-39: the name and each role are replaced alone; the unset value removes
// a role.
func TestWithNameAndRoles(t *testing.T) {
	t.Parallel()
	src := named(kp1())

	renamed, err := src.WithName("Top")
	require.NoError(t, err)
	assertRoles(t, renamed, "Top", voice.Soprano, voice.Leading, voice.Principal)
	assert.True(t, src.EqualElements(renamed))
	assert.NotSame(t, src, renamed)
	same := src.MustWithName("Melody")
	assertLine(t, src, same)
	assert.NotSame(t, src, same, "a new line is returned even without a change")

	p, err := src.WithPlacement(voice.Alto)
	require.NoError(t, err)
	assertRoles(t, p, "Melody", voice.Alto, voice.Leading, voice.Principal)
	r, err := src.WithThematic(voice.Pedal)
	require.NoError(t, err)
	assertRoles(t, r, "Melody", voice.Soprano, voice.Pedal, voice.Principal)
	i, err := src.WithImportance(voice.Harmonic)
	require.NoError(t, err)
	assertRoles(t, i, "Melody", voice.Soprano, voice.Leading, voice.Harmonic)

	cleared := src.MustWithPlacement(voice.PlacementUnset).MustWithThematic(voice.ThematicUnset).MustWithImportance(voice.ImportanceUnset)
	assertRoles(t, cleared, "Melody", voice.PlacementUnset, voice.ThematicUnset, voice.ImportanceUnset)
	assertRoles(t, src, "Melody", voice.Soprano, voice.Leading, voice.Principal)
}

// КП-35 and the order of the checks of architecture 6.12.
func TestEditFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	l := line(snd(c4(), q()), rst(q()), snd(d4(), q()))
	r := rst(q())
	tests := []struct {
		name string
		call func() (*voice.Line, error)
		must func() *voice.Line
		want want
	}{
		{"Append to an unset line", func() (*voice.Line, error) { return unset.Append(voice.Element{}) },
			func() *voice.Line { return unset.MustAppend(voice.Element{}) }, fails(voice.ErrUnset).only()},
		{"Append an unset element", func() (*voice.Line, error) { return l.Append(r, voice.Element{}) },
			func() *voice.Line { return l.MustAppend(r, voice.Element{}) }, failsAt(1, voice.ErrUnset).only()},
		{"КП-35 Insert at 4", func() (*voice.Line, error) { return l.Insert(4, r) },
			func() *voice.Line { return l.MustInsert(4, r) }, fails(voice.ErrIndex).only()},
		{"Insert at −1", func() (*voice.Line, error) { return l.Insert(-1, r) },
			func() *voice.Line { return l.MustInsert(-1, r) }, fails(voice.ErrIndex).only()},
		{"Insert index before elements", func() (*voice.Line, error) { return l.Insert(9, voice.Element{}) },
			func() *voice.Line { return l.MustInsert(9, voice.Element{}) }, fails(voice.ErrIndex).only()},
		{"Insert an unset element", func() (*voice.Line, error) { return l.Insert(0, r, r, voice.Element{}) },
			func() *voice.Line { return l.MustInsert(0, r, r, voice.Element{}) }, failsAt(2, voice.ErrUnset).only()},
		{"Insert into an unset line", func() (*voice.Line, error) { return unset.Insert(9, voice.Element{}) },
			func() *voice.Line { return unset.MustInsert(9, voice.Element{}) }, fails(voice.ErrUnset).only()},
		{"Replace at 3", func() (*voice.Line, error) { return l.Replace(3, r) },
			func() *voice.Line { return l.MustReplace(3, r) }, fails(voice.ErrIndex).only()},
		{"Replace index before element", func() (*voice.Line, error) { return l.Replace(3, voice.Element{}) },
			func() *voice.Line { return l.MustReplace(3, voice.Element{}) }, fails(voice.ErrIndex).only()},
		{"Replace with an unset element", func() (*voice.Line, error) { return l.Replace(0, voice.Element{}) },
			func() *voice.Line { return l.MustReplace(0, voice.Element{}) }, fails(voice.ErrUnset).only()},
		{"Remove at 3", func() (*voice.Line, error) { return l.Remove(3) },
			func() *voice.Line { return l.MustRemove(3) }, fails(voice.ErrIndex).only()},
		{"Remove at −1", func() (*voice.Line, error) { return l.Remove(-1) },
			func() *voice.Line { return l.MustRemove(-1) }, fails(voice.ErrIndex).only()},
		{"Remove from the empty line", func() (*voice.Line, error) { return line().Remove(0) },
			func() *voice.Line { return line().MustRemove(0) }, fails(voice.ErrIndex).only()},
		{"КП-35 RemoveRange [2, 1)", func() (*voice.Line, error) { return l.RemoveRange(2, 1) },
			func() *voice.Line { return l.MustRemoveRange(2, 1) }, fails(voice.ErrIndex).only()},
		{"RemoveRange beyond the end", func() (*voice.Line, error) { return l.RemoveRange(1, 4) },
			func() *voice.Line { return l.MustRemoveRange(1, 4) }, fails(voice.ErrIndex).only()},
		{"RemoveRange from −1", func() (*voice.Line, error) { return l.RemoveRange(-1, 1) },
			func() *voice.Line { return l.MustRemoveRange(-1, 1) }, fails(voice.ErrIndex).only()},
		{"КП-35 Slice [2, 1)", func() (*voice.Line, error) { return l.Slice(2, 1) },
			func() *voice.Line { return l.MustSlice(2, 1) }, fails(voice.ErrIndex).only()},
		{"Slice beyond the end", func() (*voice.Line, error) { return l.Slice(0, 4) },
			func() *voice.Line { return l.MustSlice(0, 4) }, fails(voice.ErrIndex).only()},
		{"WithName of an unset line", func() (*voice.Line, error) { return unset.WithName("x") },
			func() *voice.Line { return unset.MustWithName("x") }, fails(voice.ErrUnset).only()},
		{"WithPlacement out of the list", func() (*voice.Line, error) { return l.WithPlacement(voice.Placement(42)) },
			func() *voice.Line { return l.MustWithPlacement(voice.Placement(42)) }, fails(voice.ErrUnset).only()},
		{"WithPlacement of an unset line", func() (*voice.Line, error) { return unset.WithPlacement(voice.Soprano) },
			func() *voice.Line { return unset.MustWithPlacement(voice.Soprano) }, fails(voice.ErrUnset).only()},
		{"WithThematic out of the list", func() (*voice.Line, error) { return l.WithThematic(voice.ThematicRole(42)) },
			func() *voice.Line { return l.MustWithThematic(voice.ThematicRole(42)) }, fails(voice.ErrUnset).only()},
		{"WithImportance out of the list", func() (*voice.Line, error) { return l.WithImportance(voice.Importance(42)) },
			func() *voice.Line { return l.MustWithImportance(voice.Importance(42)) }, fails(voice.ErrUnset).only()},
		{"WithNote index", func() (*voice.Line, error) { return l.WithNote(3, note.Note{}) },
			func() *voice.Line { return l.MustWithNote(3, note.Note{}) }, fails(voice.ErrIndex).only()},
		{"WithNote unset note", func() (*voice.Line, error) { return l.WithNote(0, note.Note{}) },
			func() *voice.Line { return l.MustWithNote(0, note.Note{}) }, fails(voice.ErrUnset, note.ErrUnset)},
		{"WithNote without octave", func() (*voice.Line, error) { return l.WithNote(1, noOctave()) },
			func() *voice.Line { return l.MustWithNote(1, noOctave()) }, fails(voice.ErrNoOctave).only()},
		{"WithDuration index", func() (*voice.Line, error) { return l.WithDuration(-1, voice.Duration{}) },
			func() *voice.Line { return l.MustWithDuration(-1, voice.Duration{}) }, fails(voice.ErrIndex).only()},
		{"WithDuration unset", func() (*voice.Line, error) { return l.WithDuration(0, voice.Duration{}) },
			func() *voice.Line { return l.MustWithDuration(0, voice.Duration{}) }, fails(voice.ErrUnset).only()},
		{"WithDuration tied on a rest", func() (*voice.Line, error) { return l.WithDuration(1, qe()) },
			func() *voice.Line { return l.MustWithDuration(1, qe()) }, fails(voice.ErrTiedRest).only()},
		{"WithFermata index", func() (*voice.Line, error) { return l.WithFermata(3, true) },
			func() *voice.Line { return l.MustWithFermata(3, true) }, fails(voice.ErrIndex).only()},
		{"WithFermata of an unset line", func() (*voice.Line, error) { return unset.WithFermata(9, true) },
			func() *voice.Line { return unset.MustWithFermata(9, true) }, fails(voice.ErrUnset).only()},
	}
	// The receiver is checked after all parallel subtests have finished.
	t.Cleanup(func() { assertLine(t, line(snd(c4(), q()), rst(q()), snd(d4(), q())), l) })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want, tt.call, tt.must)
		})
	}
}

func TestSplitFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	for _, tt := range []struct {
		name string
		l    *voice.Line
		i    int
		want want
	}{
		{"index above", kp1(), 5, fails(voice.ErrIndex).only()},
		{"negative index", kp1(), -1, fails(voice.ErrIndex).only()},
		{"unset line", unset, 0, fails(voice.ErrUnset).only()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (pair[*voice.Line, *voice.Line], error) { return two(tt.l.Split(tt.i)) },
				func() pair[*voice.Line, *voice.Line] { return mustTwo(tt.l.MustSplit(tt.i)) })
		})
	}
}

// Architecture 6.1: building a line with a new composition or order of
// lengths fails when a partial sum is out of the range of common/fraction,
// even if every sum of the receiver is representable; the failure has no
// sentinel of voice, no *voice.ItemError and comes after the other checks.
func TestOverflowOfPartialSums(t *testing.T) {
	t.Parallel()
	invP, _, invQ := overflowDurations()
	a := rst(voice.MustValueDuration(invP))
	c := rst(voice.MustValueDuration(invQ))
	// overflowLine() is the rest 1/p, the sound (p−1)/p and the rest 1/q:
	// starts 0, 1/p, 1; a and c are its first and last elements.
	tests := []struct {
		name string
		call func(l *voice.Line) (*voice.Line, error)
		must func(l *voice.Line) *voice.Line
	}{
		{"Remove", func(l *voice.Line) (*voice.Line, error) { return l.Remove(1) },
			func(l *voice.Line) *voice.Line { return l.MustRemove(1) }},
		{"RemoveRange", func(l *voice.Line) (*voice.Line, error) { return l.RemoveRange(1, 2) },
			func(l *voice.Line) *voice.Line { return l.MustRemoveRange(1, 2) }},
		{"Slice", func(l *voice.Line) (*voice.Line, error) { return l.Slice(1, 3) },
			func(l *voice.Line) *voice.Line { return l.MustSlice(1, 3) }},
		{"Append", func(l *voice.Line) (*voice.Line, error) { return l.MustSlice(0, 1).Append(c) },
			func(l *voice.Line) *voice.Line { return l.MustSlice(0, 1).MustAppend(c) }},
		{"Insert", func(l *voice.Line) (*voice.Line, error) { return l.MustSlice(0, 1).Insert(1, c) },
			func(l *voice.Line) *voice.Line { return l.MustSlice(0, 1).MustInsert(1, c) }},
		{"Replace", func(l *voice.Line) (*voice.Line, error) { return l.Replace(1, c) },
			func(l *voice.Line) *voice.Line { return l.MustReplace(1, c) }},
		{"WithDuration", func(l *voice.Line) (*voice.Line, error) { return l.WithDuration(1, voice.MustValueDuration(invQ)) },
			func(l *voice.Line) *voice.Line { return l.MustWithDuration(1, voice.MustValueDuration(invQ)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := overflowLine()
			assertFailure(t, overflowWant(),
				func() (*voice.Line, error) { return tt.call(src) },
				func() *voice.Line { return tt.must(src) })
			assertLine(t, overflowLine(), src)
		})
	}

	t.Run("Split", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, overflowWant(),
			func() (pair[*voice.Line, *voice.Line], error) { return two(overflowLine().Split(1)) },
			func() pair[*voice.Line, *voice.Line] { return mustTwo(overflowLine().MustSplit(1)) })
	})

	t.Run("New", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, overflowWant(),
			func() (*voice.Line, error) { return voice.New(a, c) },
			func() *voice.Line { return voice.MustNew(a, c) })
	})

	t.Run("FromPairs", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, overflowWant(),
			func() (*voice.Line, error) {
				return voice.FromPairs(voice.Pair{Duration: voice.MustValueDuration(invP)}, voice.Pair{Duration: voice.MustValueDuration(invQ)})
			},
			func() *voice.Line {
				return voice.MustFromPairs(voice.Pair{Duration: voice.MustValueDuration(invP)}, voice.Pair{Duration: voice.MustValueDuration(invQ)})
			})
	})

	t.Run("Join", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, overflowWant(),
			func() (*voice.Line, error) { return voice.Join(line(a), line(c)) },
			func() *voice.Line { return voice.MustJoin(line(a), line(c)) })
	})

	// Other failures come before the overflow.
	t.Run("order", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, failsAt(2, voice.ErrUnset).only(),
			func() (*voice.Line, error) { return voice.New(a, c, voice.Element{}) },
			func() *voice.Line { return voice.MustNew(a, c, voice.Element{}) })
		assertFailure(t, failsAt(1, voice.ErrUnset).only(),
			func() (*voice.Line, error) { return voice.Join(line(a), nil, line(c)) },
			func() *voice.Line { return voice.MustJoin(line(a), nil, line(c)) })
		assertFailure(t, failsAt(1, voice.ErrUnset).only(),
			func() (*voice.Line, error) { return line(a).Append(c, voice.Element{}) },
			func() *voice.Line { return line(a).MustAppend(c, voice.Element{}) })
	})

	// Operations that keep lengths and their order do not fail.
	t.Run("kept lengths", func(t *testing.T) {
		t.Parallel()
		src := overflowLine()
		_, err := src.WithName("x")
		require.NoError(t, err)
		_, err = src.WithNote(0, d4())
		require.NoError(t, err)
		_, err = src.WithFermata(2, true)
		require.NoError(t, err)
		_, err = src.WithPlacement(voice.Bass)
		require.NoError(t, err)
	})
}
