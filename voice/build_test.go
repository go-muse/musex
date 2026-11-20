package voice_test

import (
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertRoles checks the name and the roles of a line.
func assertRoles(t *testing.T, l *voice.Line, name string, p voice.Placement, r voice.ThematicRole, i voice.Importance) {
	t.Helper()
	assert.Equal(t, name, l.MustName())
	assert.Equal(t, p, l.MustPlacement())
	assert.Equal(t, r, l.MustThematic())
	assert.Equal(t, i, l.MustImportance())
}

// ОП-27, П-11а: the transposition keeps rests, durations, fermatas, the
// name and the roles.
func TestTranspose(t *testing.T) {
	t.Parallel()
	src := named(line(snd(c4(), q()).MustWithFermata(true), rst(e8()), snd(d4(), qe())))
	want := named(line(snd(e4(), q()).MustWithFermata(true), rst(e8()), snd(nt(spelling.FSharp(), 4), qe())))
	got, err := src.Transpose(interval.MajorThird())
	require.NoError(t, err)
	assertLine(t, want, got)
	assertLine(t, want, src.MustTranspose(interval.MajorThird()))
	in, err := src.TransposeIn(traditional(), interval.MajorThird())
	require.NoError(t, err)
	assertLine(t, want, in)
	assertLine(t, want, src.MustTransposeIn(traditional(), interval.MajorThird()))

	down := src.MustTranspose(interval.MajorThird().MustDown())
	assertNotes(t, []note.Note{nt(spelling.AFlat(), 3), nt(spelling.BFlat(), 3)}, down.MustNotes())
}

// interval П-15: a new note has the correction and the source of the
// original note.
func TestTransposeKeepsCorrectionAndSource(t *testing.T) {
	t.Parallel()
	original := withCents(spelling.C(), 4, 15)
	got := line(snd(original, q())).MustTranspose(interval.PerfectFifth()).MustNotes()
	require.Len(t, got, 1)
	assert.True(t, withCents(spelling.G(), 4, 15).Equal(got[0]), "got %v", got[0])
	assert.True(t, original.SameSource(got[0]))
}

// КП-10, ОП-28: the doubling has the thematic role "doubling", no name and
// no other roles.
func TestDouble(t *testing.T) {
	t.Parallel()
	src := named(cde())
	want := line(snd(e4(), q()), snd(nt(spelling.FSharp(), 4), q()), snd(nt(spelling.GSharp(), 4), q())).MustWithThematic(voice.Doubling)
	got, err := src.Double(interval.MajorThird())
	require.NoError(t, err)
	assertLine(t, want, got)
	assertRoles(t, got, "", voice.PlacementUnset, voice.Doubling, voice.ImportanceUnset)
	assertLine(t, want, src.MustDouble(interval.MajorThird()))
	in, err := src.DoubleIn(traditional(), interval.MajorThird())
	require.NoError(t, err)
	assertLine(t, want, in)
	assertLine(t, want, src.MustDoubleIn(traditional(), interval.MajorThird()))
}

// КП-11, КП-12, ОП-29, ИЛ-5: the imitation enters after the shift.
func TestImitate(t *testing.T) {
	t.Parallel()
	src := named(cde())
	answer := []voice.Element{snd(g4(), q()), snd(nt(spelling.A(), 4), q()), snd(nt(spelling.B(), 4), q())}

	t.Run("КП-11 with a shift", func(t *testing.T) {
		t.Parallel()
		want := line(append([]voice.Element{rst(h())}, answer...)...).MustWithThematic(voice.Imitating)
		got, err := src.ImitateAfter(interval.PerfectFifth(), h())
		require.NoError(t, err)
		assertLine(t, want, got)
		assertRoles(t, got, "", voice.PlacementUnset, voice.Imitating, voice.ImportanceUnset)
		assert.False(t, got.MustElement(0).MustFermata(), "the rest of the entry has no fermata")
		sounds := got.MustSounds()
		require.Len(t, sounds, 3)
		assert.Equal(t, []string{"1/2", "3/4", "1/1"}, []string{sounds[0].Start.String(), sounds[1].Start.String(), sounds[2].Start.String()})
		assert.Equal(t, "5/4", got.MustLength().String())
		assertLine(t, want, src.MustImitateAfter(interval.PerfectFifth(), h()))
		in, err := src.ImitateAfterIn(traditional(), interval.PerfectFifth(), h())
		require.NoError(t, err)
		assertLine(t, want, in)
		assertLine(t, want, src.MustImitateAfterIn(traditional(), interval.PerfectFifth(), h()))
	})

	t.Run("КП-12 without a shift", func(t *testing.T) {
		t.Parallel()
		want := line(answer...).MustWithThematic(voice.Imitating)
		got, err := src.Imitate(interval.PerfectFifth())
		require.NoError(t, err)
		assertLine(t, want, got)
		starts, _ := got.MustStarts()
		assert.Equal(t, []string{"0/1", "1/4", "1/2"}, positionStrings(starts))
		assertLine(t, want, src.MustImitate(interval.PerfectFifth()))
		in, err := src.ImitateIn(traditional(), interval.PerfectFifth())
		require.NoError(t, err)
		assertLine(t, want, in)
		assertLine(t, want, src.MustImitateIn(traditional(), interval.PerfectFifth()))
	})

	t.Run("shift without notation", func(t *testing.T) {
		t.Parallel()
		got := src.MustImitateAfter(interval.PerfectUnison(), val(1, 3))
		assertLine(t, line(rst(val(1, 3)), snd(c4(), q()), snd(d4(), q()), snd(e4(), q())).MustWithThematic(voice.Imitating), got)
		assert.Equal(t, "13/12", got.MustLength().String())
	})
}

// КП-13, КП-14, ОП-30, ИЛ-6: the exact inversion around an axis.
func TestInvert(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  *voice.Line
		axis note.Note
		want []note.Note
	}{
		{"КП-13", named(cde()), c4(), []note.Note{c4(), nt(spelling.BFlat(), 3), nt(spelling.AFlat(), 3)}},
		{"КП-14 below the axis", line(snd(nt(spelling.B(), 3), q())), c4(), []note.Note{nt(spelling.DFlat(), 4)}},
		{"the axis itself", line(snd(d4(), q())), d4(), []note.Note{d4()}},
		{"axis D4", line(snd(c4(), q()), snd(nt(spelling.FSharp(), 4), q())), d4(), []note.Note{e4(), nt(spelling.BFlat(), 3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := func(t *testing.T, got *voice.Line) {
				t.Helper()
				require.NotNil(t, got)
				assertNotes(t, tt.want, got.MustNotes())
				assertRoles(t, got, "", voice.PlacementUnset, voice.Inversion, voice.ImportanceUnset)
				require.Equal(t, tt.src.MustLen(), got.MustLen())
				for i, e := range got.MustElements() {
					original := tt.src.MustElement(i)
					assert.Equal(t, original.Kind(), e.Kind(), "element %d", i)
					assert.True(t, original.MustDuration().Equal(e.MustDuration()), "element %d", i)
					assert.Equal(t, original.MustFermata(), e.MustFermata(), "element %d", i)
				}
			}
			got, err := tt.src.Invert(tt.axis)
			require.NoError(t, err)
			check(t, got)
			check(t, tt.src.MustInvert(tt.axis))
			in, err := tt.src.InvertIn(traditional(), tt.axis)
			require.NoError(t, err)
			check(t, in)
			check(t, tt.src.MustInvertIn(traditional(), tt.axis))
			byPitch, err := tt.src.InvertPitch(tt.axis.MustPitch())
			require.NoError(t, err)
			check(t, byPitch)
			check(t, tt.src.MustInvertPitch(tt.axis.MustPitch()))
			byPitchIn, err := tt.src.InvertPitchIn(traditional(), tt.axis.MustPitch())
			require.NoError(t, err)
			check(t, byPitchIn)
			check(t, tt.src.MustInvertPitchIn(traditional(), tt.axis.MustPitch()))

			// ИЛ-6: the inversion around the same axis twice gives the
			// original written pitches.
			assertNotes(t, tt.src.MustNotes(), got.MustInvert(tt.axis).MustNotes())
		})
	}
}

// ОП-30: a new note is a copy of the original sound with the new pitch:
// its correction and source; durations, rests and fermatas are kept. The
// axis is a written pitch: its source and correction are not used.
func TestInvertKeepsSoundParts(t *testing.T) {
	t.Parallel()
	original := withCents(spelling.D(), 4, 15)
	src := line(snd(original, qe()).MustWithFermata(true), rst(e8()).MustWithFermata(true))
	axes := map[string]note.Note{
		"bound axis":       c4(),
		"unbound axis":     free(spelling.C(), 4),
		"axis with cents":  withCents(spelling.C(), 4, -40),
		"axis with broken": brokenNote(spelling.C(), 4),
	}
	for name, axis := range axes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := src.MustInvert(axis)
			want := line(snd(withCents(spelling.BFlat(), 3, 15), qe()).MustWithFermata(true), rst(e8()).MustWithFermata(true)).
				MustWithThematic(voice.Inversion)
			assertLine(t, want, got)
			n, ok := got.MustElement(0).MustNote()
			require.True(t, ok)
			assert.True(t, original.SameSource(n))
		})
	}
}

// КП-45, П-9а: without an explicit system every sound is moved in its own
// system; sounds of different systems are allowed.
func TestBuildInSystemsOfSounds(t *testing.T) {
	t.Parallel()
	src := line(snd(c4(), q()), snd(n19(spelling.C(), 4), q()))
	got, err := src.Transpose(interval.MajorThird())
	require.NoError(t, err)
	notes := got.MustNotes()
	require.Len(t, notes, 2)
	assertNote(t, e4(), notes[0])
	assertNote(t, n19(spelling.E(), 4), notes[1])
	assert.True(t, notes[0].MustSettings().System().Equal(system.Traditional()))
	assert.True(t, notes[1].MustSettings().System().Equal(system.Positions19()))

	inverted := src.MustInvert(d4())
	assertNotes(t, []note.Note{e4(), e4()}, inverted.MustNotes())
	assert.True(t, inverted.MustNotes()[1].MustSettings().System().Equal(system.Positions19()))
}

// With an explicit system unbound notes are moved.
func TestBuildExplicitSystemForUnboundNotes(t *testing.T) {
	t.Parallel()
	src := line(snd(free(spelling.C(), 4), q()), rst(q()), snd(free(spelling.E(), 4), q()))
	got := src.MustTransposeIn(traditional(), interval.MajorSecond())
	assertNotes(t, []note.Note{free(spelling.D(), 4), free(spelling.FSharp(), 4)}, got.MustNotes())
	assert.False(t, got.MustNotes()[0].IsBound())
	inverted := src.MustInvertIn(traditional(), free(spelling.D(), 4))
	assertNotes(t, []note.Note{free(spelling.E(), 4), free(spelling.C(), 4)}, inverted.MustNotes())
	byPitch := src.MustInvertPitchIn(traditional(), pt(spelling.D(), 4))
	assertNotes(t, []note.Note{free(spelling.E(), 4), free(spelling.C(), 4)}, byPitch.MustNotes())
}

// A line without sounds is built without a system.
func TestBuildWithoutSounds(t *testing.T) {
	t.Parallel()
	src := named(line(rst(q()), rst(e8()).MustWithFermata(true)))
	assertLine(t, src, src.MustTranspose(interval.MajorThird()))
	assertLine(t, src.MustWithName("").MustWithPlacement(voice.PlacementUnset).MustWithImportance(voice.ImportanceUnset).
		MustWithThematic(voice.Doubling), src.MustDouble(interval.MajorThird()))
	assertLine(t, line(rst(q()), rst(e8()).MustWithFermata(true)).MustWithThematic(voice.Inversion),
		src.MustInvert(c4()))
	// The axis is checked in no system without an explicit one (6.11).
	assertLine(t, line().MustWithThematic(voice.Inversion), line().MustInvert(halfSharpC()))
	assertLine(t, line().MustWithThematic(voice.Imitating), line().MustImitate(interval.PerfectFifth()))
}

// Architecture 6.11: the order of the checks of the building operations.
func TestTransposeFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	std := line(snd(c4(), q()))
	freeLine := line(rst(q()), snd(free(spelling.C(), 4), q()))
	neutralThird := interval.MustParse("n3")
	tests := []struct {
		name string
		l    *voice.Line
		in   bool
		sys  *system.System
		i    interval.Interval
		want want
	}{
		{"unset line", unset, false, nil, interval.Interval{}, fails(voice.ErrUnset).only()},
		{"unset line with a system", unset, true, nil, interval.Interval{}, fails(voice.ErrUnset).only()},
		{"unset system before the interval", std, true, nil, interval.Interval{}, fails(voice.ErrUnset, system.ErrInvalid).without(interval.ErrUnset)},
		{"unset interval", std, false, nil, interval.Interval{}, fails(voice.ErrUnset, interval.ErrUnset)},
		{"unset interval before Е-16", freeLine, false, nil, interval.Interval{}, fails(voice.ErrUnset, interval.ErrUnset)},
		{"КП-45 unbound note without a system", freeLine, false, nil, interval.MajorThird(), failsAt(1, voice.ErrNoSystem)},
		{"Е-16 before the sounds", line(snd(c4(), q()), snd(free(spelling.C(), 4), q())), false, nil, neutralThird,
			failsAt(1, voice.ErrNoSystem)},
		{"КП-46 name not admissible in the set of the note", std, false, nil, neutralThird,
			failsAt(0, note.ErrInadmissible, system.ErrInadmissible)},
		{"name not admissible in the explicit system", line(rst(q()), snd(c4(), q())), true, traditional(), neutralThird,
			failsAt(1, system.ErrInadmissible).without(note.ErrInadmissible)},
		{"name admissible only in the explicit system", std, true, settings24().System(), neutralThird,
			failsAt(0, note.ErrInadmissible)},
		{"settings of a note cannot be read", line(snd(brokenNote(spelling.C(), 4), q())), false, nil, interval.MajorThird(),
			failsAt(0, note.ErrUnsetSettings)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, sys, i := tt.l, tt.sys, tt.i
			type op struct {
				name string
				call func() (*voice.Line, error)
				must func() *voice.Line
			}
			var ops []op
			if tt.in {
				ops = []op{
					{"TransposeIn", func() (*voice.Line, error) { return l.TransposeIn(sys, i) }, func() *voice.Line { return l.MustTransposeIn(sys, i) }},
					{"DoubleIn", func() (*voice.Line, error) { return l.DoubleIn(sys, i) }, func() *voice.Line { return l.MustDoubleIn(sys, i) }},
					{"ImitateIn", func() (*voice.Line, error) { return l.ImitateIn(sys, i) }, func() *voice.Line { return l.MustImitateIn(sys, i) }},
					{"ImitateAfterIn", func() (*voice.Line, error) { return l.ImitateAfterIn(sys, i, q()) },
						func() *voice.Line { return l.MustImitateAfterIn(sys, i, q()) }},
				}
			} else {
				ops = []op{
					{"Transpose", func() (*voice.Line, error) { return l.Transpose(i) }, func() *voice.Line { return l.MustTranspose(i) }},
					{"Double", func() (*voice.Line, error) { return l.Double(i) }, func() *voice.Line { return l.MustDouble(i) }},
					{"Imitate", func() (*voice.Line, error) { return l.Imitate(i) }, func() *voice.Line { return l.MustImitate(i) }},
					{"ImitateAfter", func() (*voice.Line, error) { return l.ImitateAfter(i, q()) },
						func() *voice.Line { return l.MustImitateAfter(i, q()) }},
				}
			}
			for _, o := range ops {
				// ImitateAfter keeps the indexes of the receiver in
				// *voice.ItemError (architecture 6.2).
				t.Run(o.name, func(t *testing.T) {
					t.Parallel()
					assertFailure(t, tt.want, o.call, o.must)
				})
			}
		})
	}
}

func TestImitateAfterFailures(t *testing.T) {
	t.Parallel()
	freeLine := line(snd(free(spelling.C(), 4), q()))
	tests := []struct {
		name  string
		l     *voice.Line
		i     interval.Interval
		shift voice.Duration
		want  want
	}{
		{"unset interval before the shift", cde(), interval.Interval{}, voice.Duration{}, fails(voice.ErrUnset, interval.ErrUnset)},
		{"unset shift", cde(), interval.PerfectFifth(), voice.Duration{}, fails(voice.ErrUnset).only()},
		{"unset shift before Е-16", freeLine, interval.PerfectFifth(), voice.Duration{}, fails(voice.ErrUnset).only()},
		{"tied shift", cde(), interval.PerfectFifth(), qe(), fails(voice.ErrTiedRest).only()},
		{"tied shift before Е-16", freeLine, interval.PerfectFifth(), qe(), fails(voice.ErrTiedRest).only()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (*voice.Line, error) { return tt.l.ImitateAfter(tt.i, tt.shift) },
				func() *voice.Line { return tt.l.MustImitateAfter(tt.i, tt.shift) })
			assertFailure(t, tt.want,
				func() (*voice.Line, error) { return tt.l.ImitateAfterIn(traditional(), tt.i, tt.shift) },
				func() *voice.Line { return tt.l.MustImitateAfterIn(traditional(), tt.i, tt.shift) })
		})
	}

	t.Run("overflow of the length", func(t *testing.T) {
		t.Parallel()
		invP, _, invQ := overflowDurations()
		l := line(snd(c4(), voice.MustValueDuration(invQ)))
		shift := voice.MustValueDuration(invP)
		assertFailure(t, overflowWant(),
			func() (*voice.Line, error) { return l.ImitateAfter(interval.PerfectFifth(), shift) },
			func() *voice.Line { return l.MustImitateAfter(interval.PerfectFifth(), shift) })
	})

	t.Run("the failure of a sound keeps its index in the receiver", func(t *testing.T) {
		t.Parallel()
		l := line(snd(c4(), q()), snd(free(spelling.C(), 4), q()))
		assertFailure(t, failsAt(1, voice.ErrNoSystem),
			func() (*voice.Line, error) { return l.ImitateAfter(interval.PerfectFifth(), h()) },
			func() *voice.Line { return l.MustImitateAfter(interval.PerfectFifth(), h()) })
	})
}

func TestInvertFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	freeLine := line(rst(q()), snd(free(spelling.C(), 4), q()))
	std := line(rst(q()), snd(c4(), q()))
	badAxis := halfSharpC()
	tests := []struct {
		name string
		l    *voice.Line
		in   bool
		sys  *system.System
		axis note.Note
		want want
	}{
		{"unset line", unset, false, nil, note.Note{}, fails(voice.ErrUnset).only()},
		{"unset system before the axis", std, true, nil, note.Note{}, fails(voice.ErrUnset, system.ErrInvalid).without(note.ErrUnset)},
		{"unset axis", freeLine, false, nil, note.Note{}, fails(voice.ErrUnset, note.ErrUnset)},
		{"axis without an octave", freeLine, false, nil, noOctave(), fails(voice.ErrNoOctave).only()},
		{"axis without an octave with a system", freeLine, true, traditional(), noOctave(), fails(voice.ErrNoOctave).only()},
		{"axis not admissible in the explicit system", freeLine, true, traditional(), badAxis, fails(system.ErrInadmissible).only()},
		{"axis not admissible in the explicit system without sounds", line(), true, traditional(), badAxis, fails(system.ErrInadmissible).only()},
		{"axis not admissible in the system of a sound", std, false, nil, badAxis, failsAt(1, system.ErrInadmissible).only()},
		{"unbound sound without a system", freeLine, false, nil, c4(), failsAt(1, voice.ErrNoSystem)},
		{"Е-16 before the axis in the system of a sound", line(snd(c4(), q()), snd(free(spelling.C(), 4), q())), false, nil, badAxis,
			failsAt(1, voice.ErrNoSystem)},
		{"sound not admissible in the explicit system", line(snd(c4(), q()), snd(halfSharpC(), q())), true, traditional(), c4(),
			failsAt(1, system.ErrInadmissible)},
		{"settings of a sound cannot be read", line(snd(c4(), q()), snd(brokenNote(spelling.E(), 4), q())), false, nil, c4(),
			failsAt(1, note.ErrUnsetSettings)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, sys, axis := tt.l, tt.sys, tt.axis
			if tt.in {
				assertFailure(t, tt.want, func() (*voice.Line, error) { return l.InvertIn(sys, axis) },
					func() *voice.Line { return l.MustInvertIn(sys, axis) })
			} else {
				assertFailure(t, tt.want, func() (*voice.Line, error) { return l.Invert(axis) },
					func() *voice.Line { return l.MustInvert(axis) })
			}
			if !axis.IsValid() || !axis.HasOctave() {
				return
			}
			p := axis.MustPitch()
			if tt.in {
				assertFailure(t, tt.want, func() (*voice.Line, error) { return l.InvertPitchIn(sys, p) },
					func() *voice.Line { return l.MustInvertPitchIn(sys, p) })
			} else {
				assertFailure(t, tt.want, func() (*voice.Line, error) { return l.InvertPitch(p) },
					func() *voice.Line { return l.MustInvertPitch(p) })
			}
		})
	}

	t.Run("unset pitch axis", func(t *testing.T) {
		t.Parallel()
		assertFailure(t, fails(voice.ErrUnset, pitch.ErrInvalid),
			func() (*voice.Line, error) { return cde().InvertPitch(pitch.Pitch{}) },
			func() *voice.Line { return cde().MustInvertPitch(pitch.Pitch{}) })
		assertFailure(t, fails(voice.ErrUnset, pitch.ErrInvalid),
			func() (*voice.Line, error) { return cde().InvertPitchIn(traditional(), pitch.Pitch{}) },
			func() *voice.Line { return cde().MustInvertPitchIn(traditional(), pitch.Pitch{}) })
		assertFailure(t, fails(voice.ErrUnset, system.ErrInvalid).without(pitch.ErrInvalid),
			func() (*voice.Line, error) { return cde().InvertPitchIn(nil, pitch.Pitch{}) },
			func() *voice.Line { return cde().MustInvertPitchIn(nil, pitch.Pitch{}) })
	})
}

// КП-15, ОП-31, П-15: the three kinds of the retrograde.
func TestRetrograde(t *testing.T) {
	t.Parallel()
	src := named(line(snd(c4(), q()), rst(e8()), snd(e4(), h()).MustWithFermata(true)))
	tests := []struct {
		kind voice.RetrogradeKind
		want *voice.Line
	}{
		{voice.RetrogradeBoth, line(snd(e4(), h()).MustWithFermata(true), rst(e8()), snd(c4(), q()))},
		{voice.RetrogradePitches, line(snd(e4(), q()), rst(e8()), snd(c4(), h()).MustWithFermata(true))},
		{voice.RetrogradeRhythm, line(snd(c4(), h()).MustWithFermata(true), rst(e8()), snd(e4(), q()))},
	}
	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			t.Parallel()
			want := named(tt.want)
			got, err := src.Retrograde(tt.kind)
			require.NoError(t, err)
			assertLine(t, want, got)
			assertLine(t, want, src.MustRetrograde(tt.kind))
			// ИЛ-6: the same retrograde twice gives the original line.
			assertLine(t, src, got.MustRetrograde(tt.kind))
		})
	}
}

// П-15: a tied duration is reversed as the duration of one sound.
func TestRetrogradeTiedSound(t *testing.T) {
	t.Parallel()
	src := line(snd(c4(), qe()), snd(d4(), e8()), rst(dq()))
	assertLine(t, line(rst(dq()), snd(d4(), e8()), snd(c4(), qe())), src.MustRetrograde(voice.RetrogradeBoth))
	assertLine(t, line(snd(d4(), qe()), snd(c4(), e8()), rst(dq())), src.MustRetrograde(voice.RetrogradePitches))
	assertLine(t, line(rst(dq()), snd(c4(), e8()), snd(d4(), qe())), src.MustRetrograde(voice.RetrogradeRhythm))
}

func TestRetrogradeFailures(t *testing.T) {
	t.Parallel()
	var unset *voice.Line
	tests := []struct {
		name string
		l    *voice.Line
		kind voice.RetrogradeKind
		want want
	}{
		{"unset line", unset, voice.RetrogradeUnset, fails(voice.ErrUnset).only()},
		{"unset kind", cde(), voice.RetrogradeUnset, fails(voice.ErrUnset).only()},
		{"kind out of the list", cde(), voice.RetrogradeKind(42), fails(voice.ErrUnset).only()},
		{"overflow of both", overflowLine(), voice.RetrogradeBoth, overflowWant()},
		{"overflow of the rhythm", overflowLine(), voice.RetrogradeRhythm, overflowWant()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (*voice.Line, error) { return tt.l.Retrograde(tt.kind) },
				func() *voice.Line { return tt.l.MustRetrograde(tt.kind) })
		})
	}

	// Lengths and their order are kept: no overflow (architecture 6.1).
	got, err := overflowLine().Retrograde(voice.RetrogradePitches)
	require.NoError(t, err)
	assertLine(t, overflowLine(), got)
}

// Architecture 6.14: the strings of RetrogradeKind.
func TestRetrogradeKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind  voice.RetrogradeKind
		valid bool
		text  string
	}{
		{voice.RetrogradeUnset, false, "RetrogradeKind(unset)"},
		{voice.RetrogradePitches, true, "pitches"},
		{voice.RetrogradeRhythm, true, "rhythm"},
		{voice.RetrogradeBoth, true, "both"},
		{voice.RetrogradeKind(42), false, "RetrogradeKind(invalid)"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.valid, tt.kind.IsValid())
			assert.Equal(t, tt.text, tt.kind.String())
		})
	}
}

// ИЛ-5: building keeps the number, kinds, durations and fermatas of
// elements; the imitation adds one rest at the start.
func TestBuildingInvariants(t *testing.T) {
	t.Parallel()
	for name, src := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			built := map[string]*voice.Line{
				"transpose": src.MustTranspose(interval.MinorSixth()),
				"double":    src.MustDouble(interval.PerfectOctave()),
				"imitate":   src.MustImitate(interval.PerfectFourth().MustDown()),
				"invert":    src.MustInvert(e4()),
			}
			for op, got := range built {
				require.Equal(t, src.MustLen(), got.MustLen(), op)
				for i := range src.MustLen() {
					a, b := src.MustElement(i), got.MustElement(i)
					assert.Equal(t, a.Kind(), b.Kind(), "%s %d", op, i)
					assert.True(t, a.MustDuration().Equal(b.MustDuration()), "%s %d", op, i)
					assert.Equal(t, a.MustFermata(), b.MustFermata(), "%s %d", op, i)
				}
				assert.True(t, src.MustLength().Equal(got.MustLength()), op)
			}

			after := src.MustImitateAfter(interval.PerfectFifth(), dq())
			require.Equal(t, src.MustLen()+1, after.MustLen())
			assertElement(t, rst(dq()), after.MustElement(0))
			for i := range src.MustLen() {
				a, b := src.MustElement(i), after.MustElement(i+1)
				assert.Equal(t, a.Kind(), b.Kind())
				assert.True(t, a.MustDuration().Equal(b.MustDuration()))
				assert.Equal(t, a.MustFermata(), b.MustFermata())
			}
			want := src.MustLength().MustAdd(frac(3, 8))
			assert.True(t, want.Equal(after.MustLength()), "length plus the shift")
		})
	}
}
