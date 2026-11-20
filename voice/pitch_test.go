package voice_test

import (
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type melodic struct {
	from, to    int
	interval    string
	throughRest bool
}

func assertMelodic(t *testing.T, want []melodic, got []voice.MelodicInterval) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for i, w := range want {
		assert.Equal(t, w.from, got[i].From, "interval %d", i)
		assert.Equal(t, w.to, got[i].To, "interval %d", i)
		assert.Equal(t, w.interval, got[i].Interval.String(), "interval %d", i)
		assert.Equal(t, w.throughRest, got[i].ThroughRest, "interval %d", i)
	}
}

// ОП-18, П-9, ИЛ-4: intervals between neighboring sounds with a direction;
// sounds separated by a rest give an interval marked ThroughRest.
func TestIntervals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		l    *voice.Line
		want []melodic
	}{
		{"КП-4", kp4(), []melodic{{0, 1, "M2", false}, {1, 3, "M2", true}}},
		{"КП-5", line(snd(c4(), q()), snd(d4(), q()), snd(g4(), q()), snd(e4(), q())),
			[]melodic{{0, 1, "M2", false}, {1, 2, "P4", false}, {2, 3, "-m3", false}}},
		{"КП-6 repeated pitch", line(snd(c4(), q()), snd(c4(), q())), []melodic{{0, 1, "P1", false}}},
		{"КП-47 one sound", line(snd(c4(), q())), []melodic{}},
		{"КП-9 empty", line(), []melodic{}},
		{"КП-9 one rest", line(rst(q())), []melodic{}},
		{"several rests", line(rst(q()), snd(c4(), q()), rst(e8()), rst(e8()), snd(nt(spelling.A(), 3), q()), rst(q())),
			[]melodic{{1, 4, "-m3", true}}},
		{"octave and more", line(snd(c4(), q()), snd(nt(spelling.D(), 5), q()), snd(nt(spelling.C(), 3), q())),
			[]melodic{{0, 1, "M9", false}, {1, 2, "-M16", false}}},
		{"spelling is kept", line(snd(c4(), q()), snd(nt(spelling.DFlat(), 4), q()), snd(nt(spelling.CSharp(), 4), q())),
			[]melodic{{0, 1, "m2", false}, {1, 2, "-d2", false}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.l.Intervals()
			require.NoError(t, err)
			assertMelodic(t, tt.want, got)
			assertMelodic(t, tt.want, tt.l.MustIntervals())
			in, err := tt.l.IntervalsIn(traditional())
			require.NoError(t, err)
			assertMelodic(t, tt.want, in)
			assertMelodic(t, tt.want, tt.l.MustIntervalsIn(traditional()))
		})
	}
}

// The result of interval.BetweenPitches is used as is.
func TestIntervalsMatchInterval(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), q()), snd(nt(spelling.BFlat(), 4), q()), snd(nt(spelling.FSharp(), 3), q()))
	got := l.MustIntervals()
	require.Len(t, got, 2)
	sys := traditional()
	first := interval.MustBetweenPitches(sys, c4().MustPitch(), pt(spelling.BFlat(), 4))
	second := interval.MustBetweenPitches(sys, pt(spelling.BFlat(), 4), pt(spelling.FSharp(), 3))
	assert.True(t, first.Equal(got[0].Interval), "want %v, got %v", first, got[0].Interval)
	assert.True(t, second.Equal(got[1].Interval), "want %v, got %v", second, got[1].Interval)
}

// Corrections do not take part in comparisons (interval П-15).
func TestIntervalsIgnoreCorrections(t *testing.T) {
	t.Parallel()
	l := line(snd(withCents(spelling.C(), 4, 30), q()), snd(withCents(spelling.C(), 4, -20), q()))
	assertMelodic(t, []melodic{{0, 1, "P1", false}}, l.MustIntervals())
}

type extremes struct {
	lowest, highest int
	lower, upper    pitch.Pitch
	ambitus         string
}

// ОП-19…ОП-21, П-10: range, extremes and ambitus by position; the first of
// several sounds at one position is taken.
func TestRangeExtremesAmbitus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		l    *voice.Line
		want extremes
	}{
		{"КП-7", line(snd(c4(), q()), snd(d4(), q()), snd(g4(), q()), snd(e4(), q())),
			extremes{0, 2, pt(spelling.C(), 4), pt(spelling.G(), 4), "P5"}},
		{"КП-8 equal positions, first by order", line(snd(c4(), q()), snd(written(spelling.BSharp(), 3), q())),
			extremes{0, 0, pt(spelling.C(), 4), pt(spelling.C(), 4), "P1"}},
		{"КП-42 enharmonically shifted extremes", line(snd(written(spelling.CFlat(), 4), q()), snd(written(spelling.BSharp(), 3), q())),
			extremes{0, 1, written(spelling.CFlat(), 4).MustPitch(), written(spelling.BSharp(), 3).MustPitch(), "-dd2"}},
		{"КП-47 one sound", line(snd(c4(), q())), extremes{0, 0, pt(spelling.C(), 4), pt(spelling.C(), 4), "P1"}},
		{"rests are skipped", line(rst(q()), snd(e4(), q()), rst(q()), snd(c4(), h())),
			extremes{3, 1, pt(spelling.C(), 4), pt(spelling.E(), 4), "M3"}},
		{"repeated highest", line(snd(e4(), q()), snd(c4(), q()), snd(e4(), q()), snd(c4(), q())),
			extremes{1, 0, pt(spelling.C(), 4), pt(spelling.E(), 4), "M3"}},
		{"descending line", line(snd(g4(), q()), snd(e4(), q()), snd(c4(), q())),
			extremes{2, 0, pt(spelling.C(), 4), pt(spelling.G(), 4), "P5"}},
		{"wide line", line(snd(nt(spelling.F(), 2), q()), snd(nt(spelling.D(), 4), q())),
			extremes{0, 1, pt(spelling.F(), 2), pt(spelling.D(), 4), "M13"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lowest, highest, err := tt.l.Extremes()
			require.NoError(t, err)
			assert.Equal(t, tt.want.lowest, lowest)
			assert.Equal(t, tt.want.highest, highest)
			ml, mh := tt.l.MustExtremes()
			assert.Equal(t, tt.want.lowest, ml)
			assert.Equal(t, tt.want.highest, mh)
			il, ih, err := tt.l.ExtremesIn(traditional())
			require.NoError(t, err)
			assert.Equal(t, tt.want.lowest, il)
			assert.Equal(t, tt.want.highest, ih)

			r, err := tt.l.Range()
			require.NoError(t, err)
			assert.True(t, tt.want.lower.Equal(r.Lower()), "lower: want %v, got %v", tt.want.lower, r.Lower())
			assert.True(t, tt.want.upper.Equal(r.Upper()), "upper: want %v, got %v", tt.want.upper, r.Upper())
			assert.True(t, r.Equal(tt.l.MustRange()))
			assert.True(t, r.Equal(tt.l.MustRangeIn(traditional())))

			a, err := tt.l.Ambitus()
			require.NoError(t, err)
			assert.Equal(t, tt.want.ambitus, a.String())
			assert.Equal(t, tt.want.ambitus, tt.l.MustAmbitus().String())
			in, err := tt.l.AmbitusIn(traditional())
			require.NoError(t, err)
			assert.Equal(t, tt.want.ambitus, in.String())
			assert.Equal(t, tt.want.ambitus, tt.l.MustAmbitusIn(traditional()).String())
		})
	}
}

// КП-42: the ambitus is the result of interval.BetweenPitches as is.
func TestAmbitusIsBetweenPitches(t *testing.T) {
	t.Parallel()
	low, high := written(spelling.CFlat(), 4), written(spelling.BSharp(), 3)
	l := line(snd(low, q()), snd(high, q()))
	want := interval.MustBetweenPitches(traditional(), low.MustPitch(), high.MustPitch())
	got := l.MustAmbitus()
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
	assert.True(t, got.IsDescending())
	r := l.MustRange()
	assert.True(t, instrument.MustNewRange(low.MustPitch(), high.MustPitch()).Equal(r))
}

// КП-9, Е-8: a line without sounds has no range.
func TestNoSounds(t *testing.T) {
	t.Parallel()
	for name, l := range map[string]*voice.Line{"empty": line(), "rests": line(rst(q()), rst(h()))} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			w := fails(voice.ErrNoSounds).only()
			assertFailure(t, w, l.Range, l.MustRange)
			assertFailure(t, w, func() (instrument.Range, error) { return l.RangeIn(traditional()) },
				func() instrument.Range { return l.MustRangeIn(traditional()) })
			assertFailure(t, w, func() (pair[int, int], error) { return two(l.Extremes()) },
				func() pair[int, int] { return mustTwo(l.MustExtremes()) })
			assertFailure(t, w, func() (pair[int, int], error) { return two(l.ExtremesIn(traditional())) },
				func() pair[int, int] { return mustTwo(l.MustExtremesIn(traditional())) })
			assertFailure(t, w, l.Ambitus, l.MustAmbitus)
			assertFailure(t, w, func() (interval.Interval, error) { return l.AmbitusIn(traditional()) },
				func() interval.Interval { return l.MustAmbitusIn(traditional()) })
		})
	}
}

// Architecture 6.9: the order of checks of the pitch operations.
func TestPitchOperationFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	bad := halfSharpC()
	tests := []struct {
		name string
		l    *voice.Line
		sys  *system.System
		in   bool
		want want
	}{
		{"unset line", unset, nil, true, fails(voice.ErrUnset).only()},
		{"unset line without system", unset, nil, false, fails(voice.ErrUnset).only()},
		{"unset system before no sounds", line(), nil, true, fails(voice.ErrUnset, system.ErrInvalid)},
		{"КП-32 unbound note without system", line(snd(free(spelling.C(), 4), q()), snd(free(spelling.E(), 4), q())), nil, false,
			failsAt(0, voice.ErrNoSystem)},
		{"one unbound sound is checked", line(rst(q()), snd(free(spelling.C(), 4), q())), nil, false, failsAt(1, voice.ErrNoSystem)},
		{"name not admissible in the explicit system", line(snd(free(spelling.C(), 4), q()), snd(bad, q())), traditional(), true,
			failsAt(1, system.ErrInadmissible)},
		{"КП-45 different systems", line(snd(c4(), q()), snd(n19(spelling.C(), 4), q())), nil, false,
			failsAt(1, voice.ErrDifferentSystems)},
		{"the first sound gives the system", line(snd(n19(spelling.C(), 4), q()), snd(c4(), q()), snd(n19(spelling.E(), 4), q())), nil, false,
			failsAt(1, voice.ErrDifferentSystems)},
		{"settings of a note cannot be read", line(snd(c4(), q()), snd(brokenNote(spelling.E(), 4), q())), nil, false,
			failsAt(1, note.ErrUnsetSettings)},
		{"unbound before settings", line(snd(brokenNote(spelling.E(), 4), q()), snd(free(spelling.C(), 4), q())), nil, false,
			failsAt(1, voice.ErrNoSystem)},
		{"settings before different systems", line(snd(n19(spelling.C(), 4), q()), snd(c4(), q()), snd(brokenNote(spelling.E(), 4), q())), nil, false,
			failsAt(2, note.ErrUnsetSettings)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, sys := tt.l, tt.sys
			if tt.in {
				assertFailure(t, tt.want, func() (instrument.Range, error) { return l.RangeIn(sys) },
					func() instrument.Range { return l.MustRangeIn(sys) })
				assertFailure(t, tt.want, func() (pair[int, int], error) { return two(l.ExtremesIn(sys)) },
					func() pair[int, int] { return mustTwo(l.MustExtremesIn(sys)) })
				assertFailure(t, tt.want, func() (interval.Interval, error) { return l.AmbitusIn(sys) },
					func() interval.Interval { return l.MustAmbitusIn(sys) })
				assertFailure(t, tt.want, func() ([]voice.MelodicInterval, error) { return l.IntervalsIn(sys) },
					func() []voice.MelodicInterval { return l.MustIntervalsIn(sys) })
				return
			}
			assertFailure(t, tt.want, l.Range, l.MustRange)
			assertFailure(t, tt.want, func() (pair[int, int], error) { return two(l.Extremes()) },
				func() pair[int, int] { return mustTwo(l.MustExtremes()) })
			assertFailure(t, tt.want, l.Ambitus, l.MustAmbitus)
			assertFailure(t, tt.want, l.Intervals, l.MustIntervals)
		})
	}
}

// The explicit system is checked before the state of the line: an unset
// system fails also for a line without sounds (architecture 6.9).
func TestIntervalsInUnsetSystem(t *testing.T) {
	t.Parallel()
	assertFailure(t, fails(voice.ErrUnset, system.ErrInvalid),
		func() ([]voice.MelodicInterval, error) { return line().IntervalsIn(nil) },
		func() []voice.MelodicInterval { return line().MustIntervalsIn(nil) })
}

// КП-32: with the explicit traditional system unbound notes are compared.
func TestExplicitSystemForUnboundNotes(t *testing.T) {
	t.Parallel()
	l := line(snd(free(spelling.C(), 4), q()), snd(free(spelling.E(), 4), q()))
	got, err := l.IntervalsIn(traditional())
	require.NoError(t, err)
	assertMelodic(t, []melodic{{0, 1, "M3", false}}, got)
	assert.Equal(t, "M3", l.MustAmbitusIn(traditional()).String())
	r := l.MustRangeIn(traditional())
	assert.True(t, instrument.MustNewRange(pt(spelling.C(), 4), pt(spelling.E(), 4)).Equal(r))
}

// П-9а: with an explicit system the sources of notes are not read.
func TestExplicitSystemDoesNotReadSources(t *testing.T) {
	t.Parallel()
	l := line(snd(brokenNote(spelling.C(), 4), q()), snd(brokenNote(spelling.E(), 4), q()))
	assertMelodic(t, []melodic{{0, 1, "M3", false}}, l.MustIntervalsIn(traditional()))
	assert.Equal(t, "M3", l.MustAmbitusIn(traditional()).String())
	lowest, highest := l.MustExtremesIn(traditional())
	assert.Equal(t, 0, lowest)
	assert.Equal(t, 1, highest)

	mixed := line(snd(c4(), q()), snd(n19(spelling.E(), 4), q()))
	assertMelodic(t, []melodic{{0, 1, "M3", false}}, mixed.MustIntervalsIn(traditional()))
}

// Notes of the 19-position system are compared in their system.
func TestIntervalsInOtherSystem(t *testing.T) {
	t.Parallel()
	l := line(snd(n19(spelling.C(), 4), q()), snd(n19(spelling.E(), 4), q()), snd(n19(spelling.DSharp(), 4), q()))
	assertMelodic(t, []melodic{{0, 1, "M3", false}, {1, 2, "-m2", false}}, l.MustIntervals())
	lowest, highest := l.MustExtremes()
	assert.Equal(t, 0, lowest)
	assert.Equal(t, 1, highest, "E4 is above D sharp 4 in the 19-position system")
}
