package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The expected moves come from business-spec SK-1 (C4 E4 G4 C5 B4) and from
// the move predicates of architecture 4.2.

type wantMove struct {
	from, to    int
	iv          string
	chromatic   int
	throughRest bool
	repeat      bool
	step        bool
	leapThird   bool
	leapFourth  bool
}

func assertMoves(t *testing.T, want []wantMove, got []voiceleading.Move) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for k, w := range want {
		mv := got[k]
		assert.Equal(t, w.from, mv.From.Index, "move %d from", k)
		assert.Equal(t, w.to, mv.To.Index, "move %d to", k)
		assertIntervalText(t, w.iv, mv.Interval, "move", k)
		assert.Equal(t, system.Step(w.chromatic), mv.Chromatic, "move %d chromatic", k)
		assert.Equal(t, w.throughRest, mv.ThroughRest, "move %d through rest", k)
		assert.Equal(t, w.repeat, mv.IsRepeat(), "move %d repeat", k)
		assert.Equal(t, w.step, mv.IsStep(), "move %d step", k)
		assert.Equal(t, w.leapThird, mv.IsLeap(voiceleading.LeapFromThird), "move %d leap from a third", k)
		assert.Equal(t, w.leapFourth, mv.IsLeap(voiceleading.LeapFromFourth), "move %d leap from a fourth", k)
		assert.False(t, mv.IsLeap(voiceleading.LeapUnset))
		assert.False(t, mv.IsLeap(voiceleading.LeapDefinition(9)))
	}
}

func TestMelodyMoves(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		tokens []string
		want   []wantMove
	}{
		{"SK-1 example C4 E4 G4 C5 B4", []string{"C4", "E4", "G4", "C5", "B4"}, []wantMove{
			{0, 1, "M3", 4, false, false, false, true, false},
			{1, 2, "m3", 3, false, false, false, true, false},
			{2, 3, "P4", 5, false, false, false, true, true},
			{3, 4, "-m2", -1, false, false, true, false, false},
		}},
		{"a rest gives a move marked through the rest", []string{"C4", "-", "-", "D4"}, []wantMove{
			{0, 3, "M2", 2, true, false, true, false, false},
		}},
		{"a perfect unison is a repeat", []string{"C4", "C4"}, []wantMove{
			{0, 1, "P1", 0, false, true, false, false, false},
		}},
		{"an augmented unison is neither a repeat nor a step nor a leap", []string{"C4", "C#4"}, []wantMove{
			{0, 1, "A1", 1, false, false, false, false, false},
		}},
		{"an enharmonic diminished second is a step with no chromatic motion", []string{"C#4", "Db4"}, []wantMove{
			{0, 1, "d2", 0, false, false, true, false, false},
		}},
		{"an augmented second is a step", []string{"F4", "G#4"}, []wantMove{
			{0, 1, "A2", 3, false, false, true, false, false},
		}},
		{"descending leaps keep the written direction", []string{"C5", "G4", "E4"}, []wantMove{
			{0, 1, "-P4", -5, false, false, false, true, true},
			{1, 2, "-m3", -3, false, false, false, true, false},
		}},
		{"an empty line", nil, []wantMove{}},
		{"a single sound", []string{"C4"}, []wantMove{}},
		{"rests only", []string{"-", "-"}, []wantMove{}},
		{"a sound between rests", []string{"-", "C4", "-"}, []wantMove{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := melodyOf(t, tt.tokens...)
			moves, err := m.Moves()
			require.NoError(t, err)
			assertMoves(t, tt.want, moves)
			assert.Equal(t, moves, m.MustMoves())
		})
	}
}

func TestMoveFromAndToKeepElementsAndPositions(t *testing.T) {
	t.Parallel()
	l := voice.MustNew(sound(t, "C", 4, 1, 2, nil), rest(1, 4), sound(t, "E", 4, 1, 4, nil))
	m := voiceleading.MustNewMelody(l, traditionalSystem())
	moves := m.MustMoves()
	require.Len(t, moves, 1)
	mv := moves[0]
	assert.True(t, l.MustElement(0).Equal(mv.From.Element))
	assert.True(t, l.MustElement(2).Equal(mv.To.Element))
	assertPosition(t, pos(0, 1), mv.From.Start)
	assertPosition(t, pos(1, 2), mv.From.End)
	assertPosition(t, pos(3, 4), mv.To.Start)
	assertPosition(t, pos(1, 1), mv.To.End)
	assert.True(t, mv.ThroughRest)
	elements := m.MustElements()
	require.Len(t, elements, 3)
	for i, e := range elements {
		assert.Equal(t, i, e.Index)
		assert.True(t, l.MustElement(i).Equal(e.Element))
	}
	assertPosition(t, pos(1, 2), elements[1].Start)
	assertPosition(t, pos(3, 4), elements[1].End)
	assert.True(t, elements[1].Element.IsRest())
}

func TestMovePredicatesUseTheStructFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                                string
		iv                                  string
		repeat, step, leapThird, leapFourth bool
	}{
		{"P1", "P1", true, false, false, false},
		{"A1", "A1", false, false, false, false},
		{"d2", "d2", false, true, false, false},
		{"A2", "A2", false, true, false, false},
		{"M3", "M3", false, false, true, false},
		{"descending m3", "-m3", false, false, true, false},
		{"P4", "P4", false, false, true, true},
		{"P8", "P8", false, false, true, true},
		{"M9", "M9", false, false, true, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mv := voiceleading.Move{Interval: ivl(t, tt.iv)}
			assert.Equal(t, tt.repeat, mv.IsRepeat())
			assert.Equal(t, tt.step, mv.IsStep())
			assert.Equal(t, tt.leapThird, mv.IsLeap(voiceleading.LeapFromThird))
			assert.Equal(t, tt.leapFourth, mv.IsLeap(voiceleading.LeapFromFourth))
			assert.False(t, mv.IsLeap(voiceleading.LeapUnset))
			vm := voiceleading.VoiceMove{Interval: ivl(t, tt.iv)}
			assert.Equal(t, tt.repeat, vm.IsRepeat())
			assert.Equal(t, tt.step, vm.IsStep())
			assert.Equal(t, tt.leapThird, vm.IsLeap(voiceleading.LeapFromThird))
			assert.Equal(t, tt.leapFourth, vm.IsLeap(voiceleading.LeapFromFourth))
			assert.False(t, vm.IsLeap(voiceleading.LeapDefinition(7)))
		})
	}
	var zero voiceleading.Move
	assert.False(t, zero.IsRepeat())
	assert.False(t, zero.IsStep())
	assert.False(t, zero.IsLeap(voiceleading.LeapFromThird))
	assert.False(t, zero.IsLeap(voiceleading.LeapFromFourth))
	var zeroVoice voiceleading.VoiceMove
	assert.False(t, zeroVoice.IsRepeat())
	assert.False(t, zeroVoice.IsStep())
	assert.False(t, zeroVoice.IsLeap(voiceleading.LeapFromThird))
}

func TestLeapRuns(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		tokens    []string
		def       voiceleading.LeapDefinition
		minLength int
		want      []voiceleading.LeapRun
	}{
		{"SK-2 WP k=3", []string{"C4", "E4", "G4", "C5", "B4"}, voiceleading.LeapFromThird, 3, []voiceleading.LeapRun{{First: 0, Last: 2}}},
		{"SK-2 WP k=2 gives the maximal run only", []string{"C4", "E4", "G4", "C5", "B4"}, voiceleading.LeapFromThird, 2, []voiceleading.LeapRun{{First: 0, Last: 2}}},
		{"SK-2 WP k=1", []string{"C4", "E4", "G4", "C5", "B4"}, voiceleading.LeapFromThird, 1, []voiceleading.LeapRun{{First: 0, Last: 2}}},
		{"SK-2 WP k=4 is longer than the run", []string{"C4", "E4", "G4", "C5", "B4"}, voiceleading.LeapFromThird, 4, []voiceleading.LeapRun{}},
		{"SK-2 AM k=2", []string{"C4", "E4", "G4", "C5", "B4"}, voiceleading.LeapFromFourth, 2, []voiceleading.LeapRun{}},
		{"SK-2 AM k=1", []string{"C4", "E4", "G4", "C5", "B4"}, voiceleading.LeapFromFourth, 1, []voiceleading.LeapRun{{First: 2, Last: 2}}},
		{"a rest breaks a run", []string{"C4", "E4", "-", "G4", "C5"}, voiceleading.LeapFromThird, 1, []voiceleading.LeapRun{{First: 0, Last: 0}, {First: 2, Last: 2}}},
		{"a rest breaks a run of two", []string{"C4", "E4", "-", "G4", "C5"}, voiceleading.LeapFromThird, 2, []voiceleading.LeapRun{}},
		{"two runs in order", []string{"C4", "E4", "G4", "A4", "C5", "E5"}, voiceleading.LeapFromThird, 2, []voiceleading.LeapRun{{First: 0, Last: 1}, {First: 3, Last: 4}}},
		{"an empty line", nil, voiceleading.LeapFromThird, 1, []voiceleading.LeapRun{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := melodyOf(t, tt.tokens...)
			runs, err := m.LeapRuns(tt.def, tt.minLength)
			require.NoError(t, err)
			require.NotNil(t, runs)
			assert.Equal(t, tt.want, runs)
			assert.Equal(t, runs, m.MustLeapRuns(tt.def, tt.minLength))
		})
	}
}

func TestLeapRunsRejectsOptions(t *testing.T) {
	t.Parallel()
	m := melodyOf(t, "C4", "E4", "G4")
	cases := []struct {
		name      string
		def       voiceleading.LeapDefinition
		minLength int
	}{
		{"unset definition", voiceleading.LeapUnset, 1},
		{"unknown definition", voiceleading.LeapDefinition(9), 1},
		{"zero length", voiceleading.LeapFromThird, 0},
		{"negative length", voiceleading.LeapFromFourth, -1},
		{"both wrong", voiceleading.LeapUnset, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runs, err := m.LeapRuns(tt.def, tt.minLength)
			assert.Nil(t, runs)
			assertSentinel(t, err, voiceleading.ErrOption)
			assertMustPanics(t, err, func() { m.MustLeapRuns(tt.def, tt.minLength) })
		})
	}
	// The receiver is checked before the arguments.
	var unset *voiceleading.Melody
	_, err := unset.LeapRuns(voiceleading.LeapUnset, 0)
	assertSentinel(t, err, voiceleading.ErrUnset)
}

func TestNewMelodyWithExplicitSystem(t *testing.T) {
	t.Parallel()
	t.Run("sources are never read and unbound notes are allowed", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		l := voice.MustNew(boundElement(t, "C4", src), sound(t, "D", 4, 1, 4, nil), boundElement(t, "E4", src))
		src.reset(nil)
		m, err := voiceleading.NewMelody(l, system.Positions19())
		require.NoError(t, err)
		assert.Zero(t, src.count())
		sys, ok, err := m.System()
		require.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, system.Positions19().Equal(sys))
		moves := m.MustMoves()
		require.Len(t, moves, 2)
		assert.Equal(t, system.Step(3), moves[0].Chromatic) // A major second spans 3 of 19 positions.
	})
	t.Run("an explicit system is reported even without sounds", func(t *testing.T) {
		t.Parallel()
		for _, l := range []*voice.Line{line(t), line(t, "-"), line(t, "C4")} {
			m := voiceleading.MustNewMelody(l, traditionalSystem())
			sys, ok := m.MustSystem()
			assert.True(t, ok)
			assert.True(t, traditionalSystem().Equal(sys))
		}
	})
	t.Run("an inadmissible spelling is found by its element", func(t *testing.T) {
		t.Parallel()
		l := voice.MustNew(sound(t, "C", 4, 1, 4, nil), rest(1, 4), sound(t, "C+1/2", 4, 1, 4, nil))
		m, err := voiceleading.NewMelody(l, traditionalSystem())
		assert.Nil(t, m)
		assertPartError(t, err, system.ErrInadmissible, 0, 2)
		assertSentinel(t, err, nil)
		assertMustPanics(t, err, func() { voiceleading.MustNewMelody(l, traditionalSystem()) })
		// The same spelling is admissible in the quarter-tone system.
		require.NotNil(t, voiceleading.MustNewMelody(l, system.QuarterTones24()))
	})
	t.Run("an unset explicit system", func(t *testing.T) {
		t.Parallel()
		m, err := voiceleading.NewMelody(line(t, "C4", "D4"), new(system.System))
		assert.Nil(t, m)
		assertSentinel(t, err, voiceleading.ErrUnset)
		assert.ErrorIs(t, err, system.ErrInvalid)
	})
	t.Run("an unset line precedes an unset system", func(t *testing.T) {
		t.Parallel()
		for _, l := range []*voice.Line{nil, new(voice.Line)} {
			m, err := voiceleading.NewMelody(l, new(system.System))
			assert.Nil(t, m)
			assertSentinel(t, err, voiceleading.ErrUnset)
			assertPartError(t, err, voice.ErrUnset, 0, -1)
			assertMustPanics(t, err, func() { voiceleading.MustNewMelody(l, nil) })
		}
	})
}

func TestNewMelodySelectsTheSystemFromItsSounds(t *testing.T) {
	t.Parallel()
	t.Run("one read per sound in order", func(t *testing.T) {
		t.Parallel()
		trace := &sourceTrace{}
		elements := make([]voice.Element, 0, 4)
		sources := make([]*testSource, 3)
		for i, token := range []string{"C4", "E4", "G4"} {
			sources[i] = &testSource{current: settings.Standard(), id: i}
			elements = append(elements, boundElement(t, token, sources[i]))
			if i == 1 {
				elements = append(elements, rest(1, 4))
			}
		}
		for _, src := range sources {
			src.trace = trace
			src.reset(settings.Standard())
		}
		m, err := voiceleading.NewMelody(voice.MustNew(elements...), nil)
		require.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2}, trace.values)
		sys, ok := m.MustSystem()
		assert.True(t, ok)
		assert.True(t, traditionalSystem().Equal(sys))
		moves := m.MustMoves()
		require.Len(t, moves, 2)
		assert.True(t, moves[1].ThroughRest)
		// The snapshot does not read the sources again.
		for _, src := range sources {
			src.reset(nil)
		}
		_ = m.MustMoves()
		_, _ = m.MustSystem()
		_ = m.MustCheck(voiceleading.OMTLeapRecovery)
		for _, src := range sources {
			assert.Zero(t, src.count())
		}
	})
	t.Run("an unbound sound fails before any source is read", func(t *testing.T) {
		t.Parallel()
		good := &testSource{current: settings.Standard()}
		failing := &testSource{current: settings.Standard()}
		l := voice.MustNew(boundElement(t, "C4", good), rest(1, 4), sound(t, "D", 4, 1, 4, nil), boundElement(t, "E4", failing))
		good.reset(settings.Standard())
		failing.reset(nil)
		m, err := voiceleading.NewMelody(l, nil)
		assert.Nil(t, m)
		assertPartError(t, err, voiceleading.ErrNoSystem, 0, 2)
		assertSentinel(t, err, voiceleading.ErrNoSystem)
		assert.Zero(t, good.count())
		assert.Zero(t, failing.count())
	})
	t.Run("all settings are read before systems are compared", func(t *testing.T) {
		t.Parallel()
		one := &testSource{current: settings.Standard()}
		two := &testSource{current: settings.Standard()}
		three := &testSource{current: settings.Standard()}
		l := voice.MustNew(boundElement(t, "C4", one), boundElement(t, "E4", two), boundElement(t, "G4", three))
		one.reset(settings.Standard())
		two.reset(settings19(t))
		three.reset(nil)
		m, err := voiceleading.NewMelody(l, nil)
		assert.Nil(t, m)
		assertPartError(t, err, note.ErrUnsetSettings, 0, 2)
		assertSentinel(t, err, nil)
		assert.Equal(t, 1, one.count())
		assert.Equal(t, 1, two.count())
		assert.Equal(t, 1, three.count())
	})
	t.Run("different systems", func(t *testing.T) {
		t.Parallel()
		one := &testSource{current: settings.Standard()}
		two := &testSource{current: settings.Standard()}
		l := voice.MustNew(boundElement(t, "C4", one), rest(1, 4), boundElement(t, "E4", two))
		one.reset(settings.Standard())
		two.reset(settings19(t))
		m, err := voiceleading.NewMelody(l, nil)
		assert.Nil(t, m)
		assertPartError(t, err, voiceleading.ErrDifferentSystems, 0, 2)
		assertSentinel(t, err, voiceleading.ErrDifferentSystems)
	})
	t.Run("fewer than two sounds read nothing", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		for _, l := range []*voice.Line{voice.MustNew(), voice.MustNew(rest(1, 4)), voice.MustNew(rest(1, 4), boundElement(t, "C4", src), rest(1, 4)), voice.MustNew(sound(t, "C", 4, 1, 4, nil))} {
			src.reset(nil)
			m, err := voiceleading.NewMelody(l, nil)
			require.NoError(t, err)
			assert.Zero(t, src.count())
			sys, ok, err := m.System()
			require.NoError(t, err)
			assert.Nil(t, sys)
			assert.False(t, ok)
			assert.Empty(t, m.MustMoves())
		}
	})
}

func TestAnalysisMelody(t *testing.T) {
	t.Parallel()
	t.Run("part numbers", func(t *testing.T) {
		t.Parallel()
		a := analyze(t, []string{"C4"}, []string{"G4"})
		for _, part := range []int{-1, 2, 99} {
			m, err := a.Melody(part)
			assert.Nil(t, m)
			assertSentinel(t, err, voiceleading.ErrPart)
			assertMustPanics(t, err, func() { a.MustMelody(part) })
		}
		var unset *voiceleading.Analysis
		_, err := unset.Melody(-1)
		assertSentinel(t, err, voiceleading.ErrUnset)
		_, err = new(voiceleading.Analysis).Melody(0)
		assertSentinel(t, err, voiceleading.ErrUnset)
	})
	t.Run("a tail beyond every pairwise overlap is measured without reading sources", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		low := voice.MustNew(boundElement(t, "C4", src))
		high := voice.MustNew(boundElement(t, "G4", src), boundElement(t, "A4", src), boundElement(t, "B4", src))
		src.reset(settings.Standard())
		a, err := voiceleading.Analyze([]*voice.Line{low, high}, voiceleading.Options{})
		require.NoError(t, err)
		assert.Equal(t, 2, src.count()) // v1 reads only the measured C4 and G4.
		src.reset(nil)
		m, err := a.Melody(1)
		require.NoError(t, err)
		assertMoves(t, []wantMove{
			{0, 1, "M2", 2, false, false, true, false, false},
			{1, 2, "M2", 2, false, false, true, false, false},
		}, m.MustMoves())
		sys, ok := m.MustSystem()
		assert.True(t, ok)
		assert.True(t, traditionalSystem().Equal(sys))
		_ = a.MustMelody(0)
		assert.Zero(t, src.count())
	})
	t.Run("an inadmissible tail sound under the explicit system", func(t *testing.T) {
		t.Parallel()
		low := line(t, "C4")
		high := voice.MustNew(sound(t, "G", 4, 1, 4, nil), sound(t, "C+1/2", 4, 1, 4, nil))
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), low, high)
		m, err := a.Melody(1)
		assert.Nil(t, m)
		assertPartError(t, err, system.ErrInadmissible, 1, 1)
		assertSentinel(t, err, nil)
		assert.NotNil(t, a.MustMelody(0))
	})
	t.Run("an analysis without a system", func(t *testing.T) {
		t.Parallel()
		a := analyzeParts(t, voiceleading.Unordered, nil, line(t, "C4", "D4", "-", "-"), line(t, "-", "-", "G4", "A4"))
		sys, ok := a.MustSystem()
		require.Nil(t, sys)
		require.False(t, ok)
		m, err := a.Melody(0)
		assert.Nil(t, m)
		assertPartError(t, err, voiceleading.ErrNoSystem, 0, 0)
		assertSentinel(t, err, voiceleading.ErrNoSystem)
		_, err = a.Melody(1)
		assertPartError(t, err, voiceleading.ErrNoSystem, 1, 2)
		// One sound per part needs no system.
		single := analyzeParts(t, voiceleading.Unordered, nil, line(t, "C4", "-"), line(t, "-", "G4"))
		m, err = single.Melody(1)
		require.NoError(t, err)
		assert.Empty(t, m.MustMoves())
		sys, ok, err = m.System()
		require.NoError(t, err)
		assert.Nil(t, sys)
		assert.False(t, ok)
		assert.Len(t, m.MustElements(), 2)
	})
	t.Run("an explicit system gives the same melody as NewMelody", func(t *testing.T) {
		t.Parallel()
		lines := []*voice.Line{line(t, "C4", "D4", "E4", "F4"), line(t, "G4", "-", "A4"), line(t)}
		a := analyzeParts(t, voiceleading.Unordered, traditionalSystem(), lines...)
		for i, l := range lines {
			got := a.MustMelody(i)
			want := voiceleading.MustNewMelody(l, traditionalSystem())
			assert.True(t, got.Equal(want), "part %d", i)
			assert.True(t, want.Equal(got), "part %d", i)
		}
		assert.False(t, a.MustMelody(0).Equal(a.MustMelody(1)))
	})
}
