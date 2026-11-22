package voiceleading_test

import (
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sources are private to each test. Calls are reset after construction because
// note constructors are entitled to read a bound source themselves.
type sourceTrace struct {
	mu     sync.Mutex
	values []int
}

type testSource struct {
	mu      sync.Mutex
	current *settings.Settings
	calls   int
	trace   *sourceTrace
	id      int
}

func (s *testSource) Settings() *settings.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.trace != nil {
		s.trace.mu.Lock()
		s.trace.values = append(s.trace.values, s.id)
		s.trace.mu.Unlock()
	}
	return s.current
}
func (s *testSource) reset(next *settings.Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = next
	s.calls = 0
}
func (s *testSource) count() int { s.mu.Lock(); defer s.mu.Unlock(); return s.calls }

func TestSystemSelectionAndSnapshot(t *testing.T) {
	t.Parallel()
	src := &testSource{current: settings.Standard()}
	low := voice.MustNew(sound(t, "C", 4, 1, 2, src))
	high := voice.MustNew(sound(t, "G", 4, 1, 4, src), sound(t, "A", 4, 1, 4, src))
	src.reset(settings.Standard())
	a, err := voiceleading.Analyze([]*voice.Line{low, high}, voiceleading.Options{})
	require.NoError(t, err)
	assert.Equal(t, 3, src.count()) // One per participating element, not one per source or vertical.
	sys, ok, err := a.System()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, system.Traditional().Equal(sys))
	assert.Equal(t, system.Step(7), a.MustVerticals(pair)[0].Chromatic)
	changed := settings.Standard().MustWithSystem(system.Positions19())
	src.reset(changed)
	assert.Equal(t, system.Step(7), a.MustVerticals(pair)[0].Chromatic)
	assert.Len(t, a.MustFindMotion(voiceleading.MotionOblique), 1)
	assert.True(t, a.Equal(a))
	assert.Zero(t, src.count())
	b, err := voiceleading.Analyze([]*voice.Line{low, high}, voiceleading.Options{})
	require.NoError(t, err)
	assert.Equal(t, 3, src.count())
	assert.Equal(t, system.Step(11), b.MustVerticals(pair)[0].Chromatic)
	assert.False(t, a.Equal(b))
	// Returned notes retain the live source, although measured values stay frozen.
	n, ok := a.MustVerticals(pair)[0].First.Element.MustNote()
	require.True(t, ok)
	got, err := n.Settings()
	require.NoError(t, err)
	assert.True(t, changed.Equal(got))
}

func TestExplicitSystemNeverReadsSources(t *testing.T) {
	t.Parallel()
	src := &testSource{current: settings.Standard()}
	low := voice.MustNew(sound(t, "C", 4, 1, 4, src))
	high := voice.MustNew(sound(t, "G", 4, 1, 4, src))
	src.reset(nil)
	a, err := voiceleading.Analyze([]*voice.Line{low, high}, voiceleading.Options{System: system.Positions19()})
	require.NoError(t, err)
	assert.Zero(t, src.count())
	assert.Equal(t, system.Step(11), a.MustVerticals(pair)[0].Chromatic)
	assertInterval(t, a.MustVerticals(pair)[0].Specific, 5, interval.Perfect, interval.Up)
	// Even unbound notes are valid when the common context is explicit.
	a = voiceleading.MustAnalyze([]*voice.Line{line(t, "C4"), line(t, "G4")}, voiceleading.Options{System: system.QuarterTones24()})
	assert.Equal(t, system.Step(14), a.MustVerticals(pair)[0].Chromatic)
}

func TestOnlyMeasuredElementsRequireSystems(t *testing.T) {
	t.Parallel()
	src := &testSource{current: settings.Standard()}
	low := voice.MustNew(sound(t, "C", 4, 1, 4, src), sound(t, "D", 4, 1, 4, nil), sound(t, "E", 4, 1, 4, nil))
	high := voice.MustNew(sound(t, "G", 4, 1, 4, src), rest(1, 4))
	src.reset(settings.Standard())
	a, err := voiceleading.Analyze([]*voice.Line{low, high}, voiceleading.Options{})
	require.NoError(t, err)
	assert.Equal(t, 2, src.count())
	require.Len(t, a.MustVerticals(pair), 2)
	// That previously unmeasured second note becomes measurable against a third part.
	_, err = voiceleading.Analyze([]*voice.Line{low, high, voice.MustNew(rest(1, 4), sound(t, "A", 4, 1, 4, settings.Standard()))}, voiceleading.Options{})
	assertPartError(t, err, voiceleading.ErrNoSystem, 0, 1)
	for _, lines := range [][]*voice.Line{{line(t, "C4"), line(t, "-")}, {line(t), line(t, "C4")}, {line(t, "-"), line(t, "-")}} {
		a, err := voiceleading.Analyze(lines, voiceleading.Options{})
		require.NoError(t, err)
		sys, ok, err := a.System()
		require.NoError(t, err)
		assert.Nil(t, sys)
		assert.False(t, ok)
		a = voiceleading.MustAnalyze(lines, voiceleading.Options{System: system.Traditional()})
		sys, ok = a.MustSystem()
		assert.True(t, ok)
		assert.True(t, sys.Equal(system.Traditional()))
	}
	// A bound but failing source on a note opposite a rest is also never consulted.
	silentSrc := &testSource{current: settings.Standard()}
	silent := voice.MustNew(sound(t, "C", 4, 1, 4, silentSrc))
	silentSrc.reset(nil)
	_, err = voiceleading.Analyze([]*voice.Line{silent, line(t, "-")}, voiceleading.Options{})
	require.NoError(t, err)
	assert.Zero(t, silentSrc.count())
}

func assertPartError(t *testing.T, err, target error, part, element int) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, target)
	pe, ok := errors.AsType[*voiceleading.PartError](err)
	require.True(t, ok)
	require.NotNil(t, pe)
	assert.Equal(t, part, pe.Part)
	assert.Equal(t, element, pe.Element)
	assert.ErrorIs(t, pe.Unwrap(), target)
	assert.Contains(t, pe.Error(), "voiceleading: ")
}

func TestSystemFailureOrder(t *testing.T) {
	t.Parallel()
	t.Run("all bindings precede any settings read", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		low := voice.MustNew(sound(t, "C", 4, 1, 4, src))
		src.reset(nil)
		a, err := voiceleading.Analyze([]*voice.Line{low, line(t, "G4")}, voiceleading.Options{})
		assert.Nil(t, a)
		assertPartError(t, err, voiceleading.ErrNoSystem, 1, 0)
		assert.Zero(t, src.count())
	})
	t.Run("read all settings before comparing systems", func(t *testing.T) {
		t.Parallel()
		one := &testSource{current: settings.Standard()}
		two := &testSource{current: settings.Standard()}
		three := &testSource{current: settings.Standard()}
		low := voice.MustNew(sound(t, "C", 4, 1, 4, one))
		middle := voice.MustNew(sound(t, "E", 4, 1, 4, two))
		high := voice.MustNew(sound(t, "G", 4, 1, 4, three))
		one.reset(settings.Standard())
		two.reset(settings.Standard().MustWithSystem(system.Positions19()))
		three.reset(nil)
		a, err := voiceleading.Analyze([]*voice.Line{low, middle, high}, voiceleading.Options{})
		assert.Nil(t, a)
		assertPartError(t, err, note.ErrUnsetSettings, 2, 0)
		assert.NotErrorIs(t, err, voiceleading.ErrDifferentSystems)
	})
	t.Run("first system mismatch in part order", func(t *testing.T) {
		t.Parallel()
		a, err := voiceleading.Analyze([]*voice.Line{voice.MustNew(sound(t, "C", 4, 1, 4, settings.Standard())), voice.MustNew(sound(t, "E", 4, 1, 4, settings.Standard().MustWithSystem(system.Positions19())))}, voiceleading.Options{})
		assert.Nil(t, a)
		assertPartError(t, err, voiceleading.ErrDifferentSystems, 1, 0)
	})
	t.Run("settings reads follow part and element order", func(t *testing.T) {
		t.Parallel()
		trace := &sourceTrace{}
		sources := make([]*testSource, 4)
		elements := make([]voice.Element, 4)
		for i := range sources {
			sources[i] = &testSource{current: settings.Standard(), id: i}
			elements[i] = sound(t, "C", 4, 1, 4, sources[i])
			sources[i].trace = trace
			sources[i].reset(settings.Standard())
		}
		_, err := voiceleading.Analyze([]*voice.Line{voice.MustNew(elements[:2]...), voice.MustNew(elements[2:]...)}, voiceleading.Options{})
		require.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2, 3}, trace.values)
	})
	t.Run("inadmissible fractional spelling context", func(t *testing.T) {
		t.Parallel()
		// The quarter-tone system admits half sharps; the traditional system does not.
		half := sound(t, "C+1/2", 4, 1, 4, nil)
		a, err := voiceleading.Analyze([]*voice.Line{voice.MustNew(half), line(t, "G4")}, voiceleading.Options{System: system.Traditional()})
		assert.Nil(t, a)
		assertPartError(t, err, system.ErrInadmissible, 0, 0)
	})
}

func TestMeasurementRangeErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		lines       func(t *testing.T) []*voice.Line
		atNumerator int
		cause       error
	}{
		{"vertical interval too wide", func(t *testing.T) []*voice.Line {
			return []*voice.Line{voice.MustNew(sound(t, "C", math.MinInt, 1, 4, nil)), voice.MustNew(sound(t, "C", math.MaxInt, 1, 4, nil))}
		}, 0, interval.ErrOutOfRange},
		{"melodic interval too wide", func(t *testing.T) []*voice.Line {
			return []*voice.Line{voice.MustNew(sound(t, "C", math.MinInt, 1, 4, nil), sound(t, "C", math.MaxInt, 1, 4, nil)), voice.MustNew(sound(t, "G", math.MinInt, 1, 4, nil), sound(t, "G", math.MaxInt, 1, 4, nil))}
		}, 1, interval.ErrOutOfRange},
		{"melodic chromatic distance too wide", func(t *testing.T) []*voice.Line {
			return []*voice.Line{voice.MustNew(sound(t, "C", 0, 1, 4, nil), sound(t, "C", math.MaxInt/8, 1, 4, nil)), voice.MustNew(sound(t, "G", 0, 1, 4, nil), sound(t, "G", math.MaxInt/8, 1, 4, nil))}
		}, 1, system.ErrOutOfRange},
		{"vertical chromatic distance too wide", func(t *testing.T) []*voice.Line {
			return []*voice.Line{voice.MustNew(sound(t, "C", 0, 1, 4, nil)), voice.MustNew(sound(t, "C", math.MaxInt/8, 1, 4, nil))}
		}, 0, system.ErrOutOfRange},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, err := voiceleading.Analyze(tt.lines(t), voiceleading.Options{System: system.Traditional()})
			assert.Nil(t, a)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.cause)
			me, ok := errors.AsType[*voiceleading.MeasurementError](err)
			require.True(t, ok)
			require.NotNil(t, me)
			assert.Equal(t, pair, me.Pair)
			assertPosition(t, pos(tt.atNumerator, 4), me.At)
			assert.Contains(t, me.Error(), "voiceleading: ")
		})
	}
	// Absolute positions overflow int, but nearby interval/distance must remain exact.
	a := voiceleading.MustAnalyze([]*voice.Line{voice.MustNew(sound(t, "C", math.MaxInt, 1, 4, nil)), voice.MustNew(sound(t, "G", math.MaxInt, 1, 4, nil))}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	assert.Equal(t, system.Step(7), a.MustVerticals(pair)[0].Chromatic)
	assert.Empty(t, a.MustFindCrossings())
}

func TestCentsDoNotAlterMeasurements(t *testing.T) {
	t.Parallel()
	low := sound(t, "C", 4, 1, 4, nil)
	n, ok := low.MustNote()
	require.True(t, ok)
	p, err := n.Pitch()
	require.NoError(t, err)
	shifted, err := note.FromPitch(p, cents.MustParse("1000"), nil)
	require.NoError(t, err)
	a := voiceleading.MustAnalyze([]*voice.Line{voice.MustNew(low.MustWithNote(shifted)), line(t, "G4")}, voiceleading.Options{System: system.Traditional()})
	v := a.MustVerticals(pair)[0]
	assert.Equal(t, system.Step(7), v.Chromatic)
	assertInterval(t, v.Specific, 5, interval.Perfect, interval.Up)
	// Spelling, not pitch-class distance alone, distinguishes A2 from m3.
	a = voiceleading.MustAnalyze([]*voice.Line{line(t, "C4", "C4"), line(t, "D#4", "Eb4")}, voiceleading.Options{System: system.Traditional()})
	vs := a.MustVerticals(pair)
	assert.Equal(t, system.Step(3), vs[0].Chromatic)
	assert.Equal(t, system.Step(3), vs[1].Chromatic)
	assertInterval(t, vs[0].Specific, 2, interval.Augmented, interval.Up)
	assertInterval(t, vs[1].Specific, 3, interval.Minor, interval.Up)
}

func TestWrittenIntervalPreservesQualityAndFullSize(t *testing.T) {
	t.Parallel()
	a := analyze(t, []string{"Ab3", "G3"}, []string{"D#4", "D4"})
	vs := a.MustVerticals(pair)
	assert.Equal(t, system.Step(7), vs[0].Chromatic)
	assert.Equal(t, system.Step(7), vs[1].Chromatic)
	assertInterval(t, vs[0].Specific, 4, interval.Augmented, interval.Up)
	assert.Equal(t, 2, vs[0].Specific.Multiplicity())
	assertInterval(t, vs[1].Specific, 5, interval.Perfect, interval.Up)
	compound := analyze(t, []string{"C3", "D3"}, []string{"G5", "A5"})
	found := compound.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectFifths, Compound: true})
	require.Len(t, found, 1)
	assertInterval(t, found[0].From.Specific, 19, interval.Perfect, interval.Up)
	assertInterval(t, found[0].To.Specific, 19, interval.Perfect, interval.Up)
}
