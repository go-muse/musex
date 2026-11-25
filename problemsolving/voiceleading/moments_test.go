package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: business-spec SZ-1...SZ-3 and architecture 5.

type wantElement struct {
	index      int
	start, end duration.Position
}

func assertPartSonority(t *testing.T, lines []*voice.Line, start, end duration.Position, elements []wantElement, got voiceleading.PartSonority) {
	t.Helper()
	assertPosition(t, start, got.Start)
	assertPosition(t, end, got.End)
	require.Len(t, got.Elements, len(elements))
	for p, w := range elements {
		e := got.Elements[p]
		assert.Equal(t, w.index, e.Index, "part %d", p)
		assert.True(t, lines[p].MustElement(w.index).Equal(e.Element), "part %d", p)
		assertPosition(t, w.start, e.Start)
		assertPosition(t, w.end, e.End)
	}
	require.NotNil(t, got.Voices)
	assert.True(t, got.Voices.IsSet())
}

func TestPartSonorities(t *testing.T) {
	t.Parallel()
	// SZ-1: bass C3 half, tenor G3 A3 quarters, soprano C4 half.
	bass := voice.MustNew(sound(t, "C", 3, 1, 2, nil))
	tenor := line(t, "G3", "A3")
	soprano := voice.MustNew(sound(t, "C", 4, 1, 2, nil))
	lines := []*voice.Line{bass, tenor, soprano}
	a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), lines...)
	ps, err := a.Sonorities()
	require.NoError(t, err)
	require.Len(t, ps, 2)
	assertPartSonority(t, lines, pos(0, 1), pos(1, 4), []wantElement{{0, pos(0, 1), pos(1, 2)}, {0, pos(0, 1), pos(1, 4)}, {0, pos(0, 1), pos(1, 2)}}, ps[0])
	assertPartSonority(t, lines, pos(1, 4), pos(1, 2), []wantElement{{0, pos(0, 1), pos(1, 2)}, {1, pos(1, 4), pos(1, 2)}, {0, pos(0, 1), pos(1, 2)}}, ps[1])
	assertNotesAre(t, []string{"C3", "G3", "C4"}, ps[0].Voices.MustNotes())
	assertNotesAre(t, []string{"C3", "A3", "C4"}, ps[1].Voices.MustNotes())
	for _, p := range ps {
		parts, ok := p.Voices.MustParts()
		assert.True(t, ok)
		assert.Equal(t, []int{0, 1, 2}, parts)
		assert.Empty(t, p.Voices.MustSilent())
		assert.Equal(t, voiceleading.LowToHigh, p.Voices.MustOrder())
		sys, ok := p.Voices.MustSystem()
		assert.True(t, ok)
		assert.True(t, traditionalSystem().Equal(sys))
	}
	assert.Equal(t, ps, a.MustSonorities())
	cases := []struct {
		name  string
		at    duration.Position
		found bool
		index int
	}{
		{"the first moment", pos(0, 1), true, 0},
		{"inside the first moment", pos(1, 8), true, 0},
		{"the second moment", pos(1, 4), true, 1},
		{"inside the second moment", pos(3, 8), true, 1},
		{"the common end", pos(1, 2), false, 0},
		{"beyond the common end", pos(3, 4), false, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok, err := a.SonorityAt(tt.at)
			require.NoError(t, err)
			assert.Equal(t, tt.found, ok)
			if !tt.found {
				assert.Equal(t, voiceleading.PartSonority{}, got)
				return
			}
			assertPosition(t, ps[tt.index].Start, got.Start)
			assertPosition(t, ps[tt.index].End, got.End)
			assert.True(t, ps[tt.index].Voices.Equal(got.Voices))
			mustGot, mustOK := a.MustSonorityAt(tt.at)
			assert.Equal(t, ok, mustOK)
			assert.Equal(t, got, mustGot)
		})
	}
}

func TestPartSonoritiesWithRestsAndEmptyParts(t *testing.T) {
	t.Parallel()
	t.Run("a silent part", func(t *testing.T) {
		t.Parallel()
		lines := []*voice.Line{line(t, "C3", "-"), line(t, "E4", "F4")}
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), lines...)
		ps := a.MustSonorities()
		require.Len(t, ps, 2)
		assert.True(t, ps[1].Elements[0].Element.IsRest())
		assert.True(t, ps[1].Elements[1].Element.IsSound())
		assertNotesAre(t, []string{"F4"}, ps[1].Voices.MustNotes())
		parts, ok := ps[1].Voices.MustParts()
		assert.True(t, ok)
		assert.Equal(t, []int{1}, parts)
		assert.Equal(t, []int{0}, ps[1].Voices.MustSilent())
	})
	t.Run("an empty part leaves no moments", func(t *testing.T) {
		t.Parallel()
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C4", "D4"), line(t), line(t, "G4"))
		ps, err := a.Sonorities()
		require.NoError(t, err)
		assert.NotNil(t, ps)
		assert.Empty(t, ps)
		got, ok, err := a.SonorityAt(pos(0, 1))
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, voiceleading.PartSonority{}, got)
		cs, err := a.Connections()
		require.NoError(t, err)
		assert.NotNil(t, cs)
		assert.Empty(t, cs)
	})
	t.Run("moments come from every part up to the earliest end", func(t *testing.T) {
		t.Parallel()
		lines := []*voice.Line{
			voice.MustNew(sound(t, "C", 3, 3, 8, nil), sound(t, "D", 3, 1, 8, nil), sound(t, "E", 3, 1, 4, nil)),
			voice.MustNew(sound(t, "G", 4, 1, 4, nil), sound(t, "A", 4, 1, 4, nil)),
			line(t, "C5", "D5", "E5"),
		}
		a := analyzeParts(t, voiceleading.Unordered, traditionalSystem(), lines...)
		ps := a.MustSonorities()
		require.Len(t, ps, 3)
		assertPosition(t, pos(0, 1), ps[0].Start)
		assertPosition(t, pos(1, 4), ps[1].Start)
		assertPosition(t, pos(3, 8), ps[2].Start)
		assertPosition(t, pos(1, 2), ps[2].End)
		assertNotesAre(t, []string{"C3", "A4", "D5"}, ps[1].Voices.MustNotes())
		assertNotesAre(t, []string{"D3", "A4", "D5"}, ps[2].Voices.MustNotes())
		assert.Equal(t, voiceleading.Unordered, ps[0].Voices.MustOrder())
	})
	t.Run("an unset position", func(t *testing.T) {
		t.Parallel()
		a := analyze(t, []string{"C4"}, []string{"G4"})
		got, ok, err := a.SonorityAt(duration.Position{})
		assert.Equal(t, voiceleading.PartSonority{}, got)
		assert.False(t, ok)
		assertSentinel(t, err, voiceleading.ErrUnset)
		assert.ErrorIs(t, err, duration.ErrUnset)
		assertMustPanics(t, err, func() { a.MustSonorityAt(duration.Position{}) })
	})
}

func TestPartSonoritiesReadNoSources(t *testing.T) {
	t.Parallel()
	src := &testSource{current: settings.Standard()}
	low := voice.MustNew(boundElement(t, "C3", src), boundElement(t, "D3", src))
	middle := voice.MustNew(boundElement(t, "E4", src), boundElement(t, "F4", src))
	high := voice.MustNew(boundElement(t, "G4", src), boundElement(t, "A4", src), boundElement(t, "B4", src))
	src.reset(settings.Standard())
	a, err := voiceleading.Analyze([]*voice.Line{low, middle, high}, voiceleading.Options{Order: voiceleading.LowToHigh})
	require.NoError(t, err)
	assert.Equal(t, 6, src.count())
	src.reset(nil)
	_ = a.MustSonorities()
	_, _ = a.MustSonorityAt(pos(1, 4))
	_ = a.MustConnections()
	_, _ = a.MustConnectionAt(pos(0, 1), pos(1, 4))
	for i := range 3 {
		_ = a.MustMelody(i).MustMoves()
	}
	assert.Zero(t, src.count())
}

func TestPartConnections(t *testing.T) {
	t.Parallel()
	lines := []*voice.Line{line(t, "C3", "D3", "-", "E3"), line(t, "E4", "F4", "G4", "G4")}
	a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), lines...)
	ps := a.MustSonorities()
	require.Len(t, ps, 4)
	cs, err := a.Connections()
	require.NoError(t, err)
	require.Len(t, cs, 1) // SZ-2: a rest gives no connection and no gluing across it.
	c := cs[0]
	assertPosition(t, pos(0, 1), c.From.Start)
	assertPosition(t, pos(1, 4), c.To.Start)
	assert.Same(t, c.From.Voices, c.Connection.From)
	assert.Same(t, c.To.Voices, c.Connection.To)
	require.Len(t, c.Connection.Moves, 2)
	assert.Equal(t, 0, c.Connection.Moves[0].Voice)
	assertIntervalText(t, "M2", c.Connection.Moves[0].Interval)
	assert.Equal(t, system.Step(2), c.Connection.Moves[0].Chromatic)
	assert.Equal(t, 1, c.Connection.Moves[1].Voice)
	assertIntervalText(t, "m2", c.Connection.Moves[1].Interval)
	assert.Equal(t, system.Step(1), c.Connection.Moves[1].Chromatic)
	assert.Empty(t, c.Connection.Common)
	assert.Equal(t, system.Step(3), c.Connection.Total)
	assert.Equal(t, voiceleading.MelodicConnection, c.Connection.Kind)
	got, ok, err := a.ConnectionAt(pos(0, 1), pos(1, 4))
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, got.From.Voices.Equal(c.From.Voices))
	assert.True(t, got.To.Voices.Equal(c.To.Voices))
	assert.Same(t, got.From.Voices, got.Connection.From)
	assert.Same(t, got.To.Voices, got.Connection.To)
	assert.Equal(t, c.Connection.Total, got.Connection.Total)
	for _, silent := range [][2]duration.Position{{pos(1, 4), pos(1, 2)}, {pos(1, 2), pos(3, 4)}} {
		got, ok, err := a.ConnectionAt(silent[0], silent[1])
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, voiceleading.PartConnection{}, got)
	}
	transitions := []struct {
		name     string
		from, to duration.Position
	}{
		{"a repeated start", pos(0, 1), pos(0, 1)},
		{"the reverse order", pos(1, 4), pos(0, 1)},
		{"a point inside a moment", pos(1, 8), pos(1, 4)},
		{"a skipped boundary", pos(1, 4), pos(3, 4)},
		{"the common end", pos(3, 4), pos(1, 1)},
		{"beyond the common end", pos(1, 1), pos(5, 4)},
	}
	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok, err := a.ConnectionAt(tt.from, tt.to)
			assert.Equal(t, voiceleading.PartConnection{}, got)
			assert.False(t, ok)
			assertSentinel(t, err, voiceleading.ErrTransition)
			assertMustPanics(t, err, func() { a.MustConnectionAt(tt.from, tt.to) })
		})
	}
	for _, positions := range [][2]duration.Position{{{}, pos(1, 4)}, {pos(0, 1), {}}, {{}, {}}} {
		_, _, err := a.ConnectionAt(positions[0], positions[1])
		assertSentinel(t, err, voiceleading.ErrUnset)
		assert.ErrorIs(t, err, duration.ErrUnset)
	}
}

func TestPartConnectionsNeedTheNominalOrder(t *testing.T) {
	t.Parallel()
	unordered := analyzeParts(t, voiceleading.Unordered, traditionalSystem(), line(t, "C3", "D3"), line(t, "E4", "F4"))
	empty := analyzeParts(t, voiceleading.Unordered, nil, line(t), line(t))
	for _, a := range []*voiceleading.Analysis{unordered, empty} {
		cs, err := a.Connections()
		assert.Nil(t, cs)
		assertSentinel(t, err, voiceleading.ErrNoOrder)
		assertMustPanics(t, err, func() { a.MustConnections() })
		// The order precedes the positions.
		_, ok, err := a.ConnectionAt(duration.Position{}, pos(1, 4))
		assert.False(t, ok)
		assertSentinel(t, err, voiceleading.ErrNoOrder)
		// Sonorities need no order.
		_, err = a.Sonorities()
		require.NoError(t, err)
	}
	// A connection of an unordered sonority is refused by Connect itself.
	s := unordered.MustSonorities()
	_, err := voiceleading.Connect(s[0].Voices, s[1].Voices)
	assertSentinel(t, err, voiceleading.ErrNoOrder)
}

func TestEqualAnalysesGiveEqualPartSonorities(t *testing.T) {
	t.Parallel()
	build := func() *voiceleading.Analysis {
		return analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3", "D3"), line(t, "E4", "F4", "G4"))
	}
	a, b := build(), build()
	require.True(t, a.Equal(b))
	as, bs := a.MustSonorities(), b.MustSonorities()
	require.Len(t, bs, len(as))
	for i := range as {
		assert.True(t, as[i].Voices.Equal(bs[i].Voices))
	}
	ac, bc := a.MustConnections(), b.MustConnections()
	require.Len(t, bc, len(ac))
	assert.Equal(t, ac[0].Connection.Total, bc[0].Connection.Total)
	// Melodies see the tail beyond every pairwise overlap; Equal of analyses does not.
	c := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3", "D3"), line(t, "E4", "F4", "A4"))
	assert.True(t, a.Equal(c))
	assert.False(t, a.MustMelody(1).Equal(c.MustMelody(1)))
}
