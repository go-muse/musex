package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: business-spec RZ-1...RZ-5 and architecture 11.

func quarterDuration() voice.Duration { return voice.MustValueDuration(duration.MustNew(1, 4)) }
func halfDuration() voice.Duration    { return voice.MustValueDuration(duration.MustNew(1, 2)) }

func simultaneity(t *testing.T, d voice.Duration, tokens ...string) voiceleading.Simultaneity {
	t.Helper()
	return voiceleading.Simultaneity{Notes: chordOf(t, tokens...), Duration: d}
}

// expectedLine is a line of quarter sounds; "-" is a rest.
type expectedElement struct {
	token string
	d     voice.Duration
}

func assertLineElements(t *testing.T, want []expectedElement, got *voice.Line, context ...any) {
	t.Helper()
	require.NotNil(t, got)
	elements := got.MustElements()
	require.Len(t, elements, len(want), context...)
	for i, w := range want {
		var e voice.Element
		if w.token == "-" {
			e = voice.MustNewRest(w.d)
		} else {
			e = voice.MustNewSound(noteOf(t, w.token, nil), w.d)
		}
		assert.True(t, e.Equal(elements[i]), "element %d: want %s, got %s %v", i, e, elements[i], context)
	}
	name, err := got.Name()
	require.NoError(t, err)
	assert.Empty(t, name)
}

func quarters(tokens ...string) []expectedElement {
	es := make([]expectedElement, len(tokens))
	for i, token := range tokens {
		es[i] = expectedElement{token, quarterDuration()}
	}
	return es
}

func TestDecompose(t *testing.T) {
	t.Parallel()
	q, h := quarterDuration(), halfDuration()
	cases := []struct {
		name        string
		seq         func(t *testing.T) []voiceleading.Simultaneity
		lines       [][]expectedElement
		ambiguities []voiceleading.Ambiguity
	}{
		{"RZ-1 two chords given in any order", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "G4", "C3", "C5", "E4"), simultaneity(t, q, "B4", "G2", "G4", "D4")}
		}, [][]expectedElement{quarters("C3", "G2"), quarters("E4", "D4"), quarters("G4", "G4"), quarters("C5", "B4")}, nil},
		{"RZ-1 a common rest", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C3", "E4", "G4", "C5"), {Notes: chord.MustNew(), Duration: h}, simultaneity(t, q, "G2", "D4", "G4", "B4")}
		}, [][]expectedElement{
			{{"C3", q}, {"-", h}, {"G2", q}}, {{"E4", q}, {"-", h}, {"D4", q}},
			{{"G4", q}, {"-", h}, {"G4", q}}, {{"C5", q}, {"-", h}, {"B4", q}},
		}, nil},
		{"a leading rest", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{{Notes: chord.MustNew(), Duration: q}, simultaneity(t, h, "E4", "C4")}
		}, [][]expectedElement{{{"-", q}, {"C4", h}}, {{"-", q}, {"E4", h}}}, nil},
		{"RZ-3 repeated notes stay separate attacks", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4"), simultaneity(t, h, "C4", "E4")}
		}, [][]expectedElement{{{"C4", q}, {"C4", h}}, {{"E4", q}, {"E4", h}}}, nil},
		{"RZ-2 B#3 is below C4", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "B#3")}
		}, [][]expectedElement{quarters("B#3"), quarters("C4")}, []voiceleading.Ambiguity{{Index: 0, Voices: []int{0, 1}}}},
		{"RZ-2 C4 is below Dbb4", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "Dbb4", "C4")}
		}, [][]expectedElement{quarters("C4"), quarters("Dbb4")}, []voiceleading.Ambiguity{{Index: 0, Voices: []int{0, 1}}}},
		{"RZ-2 several groups in one chord", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "Fb4", "B#3", "E4")}
		}, [][]expectedElement{quarters("B#3"), quarters("C4"), quarters("E4"), quarters("Fb4")},
			[]voiceleading.Ambiguity{{Index: 0, Voices: []int{0, 1}}, {Index: 0, Voices: []int{2, 3}}}},
		{"RZ-2 a unison of equal notes", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "E4", "C4", "C4")}
		}, [][]expectedElement{quarters("C4"), quarters("C4"), quarters("E4")}, []voiceleading.Ambiguity{{Index: 0, Voices: []int{0, 1}}}},
		{"RZ-2 a group of three", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "Dbb4", "C4", "B#3")}
		}, [][]expectedElement{quarters("B#3"), quarters("C4"), quarters("Dbb4")}, []voiceleading.Ambiguity{{Index: 0, Voices: []int{0, 1, 2}}}},
		{"ambiguities follow the chords", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "C4"), simultaneity(t, q, "E4", "G4"), simultaneity(t, q, "Abb4", "G4")}
		}, [][]expectedElement{quarters("C4", "E4", "G4"), quarters("C4", "G4", "Abb4")},
			[]voiceleading.Ambiguity{{Index: 0, Voices: []int{0, 1}}, {Index: 2, Voices: []int{0, 1}}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			seq := tt.seq(t)
			got, err := voiceleading.Decompose(seq, traditionalSystem())
			require.NoError(t, err)
			require.Len(t, got.Lines, len(tt.lines))
			for k := range tt.lines {
				assertLineElements(t, tt.lines[k], got.Lines[k], "line", k)
			}
			if tt.ambiguities == nil {
				assert.Empty(t, got.Ambiguities)
			} else {
				assert.Equal(t, tt.ambiguities, got.Ambiguities)
			}
			assert.Equal(t, got, voiceleading.MustDecompose(seq, traditionalSystem()))
			// The lines are ready for an analysis in the nominal order.
			if len(got.Lines) >= 2 {
				a, err := voiceleading.Analyze(got.Lines, voiceleading.Options{System: traditionalSystem(), Order: voiceleading.LowToHigh})
				require.NoError(t, err)
				assert.Equal(t, len(got.Lines), a.MustPartCount())
				assert.Empty(t, a.MustFindCrossings())
			}
		})
	}
}

func TestDecomposeChoosesTheSystemOnlyForSeveralVoices(t *testing.T) {
	t.Parallel()
	t.Run("one voice needs no system", func(t *testing.T) {
		t.Parallel()
		got, err := voiceleading.Decompose([]voiceleading.Simultaneity{simultaneity(t, quarterDuration(), "C4"), {Notes: chord.MustNew(), Duration: quarterDuration()}, simultaneity(t, quarterDuration(), "D4")}, nil)
		require.NoError(t, err)
		require.Len(t, got.Lines, 1)
		assertLineElements(t, quarters("C4", "-", "D4"), got.Lines[0])
		assert.Empty(t, got.Ambiguities)
	})
	t.Run("one read per note in the order of record", func(t *testing.T) {
		t.Parallel()
		trace := &sourceTrace{}
		sources := make([]*testSource, 4)
		tokens := []string{"E4", "C4", "G4", "D4"}
		notes := make([]note.Note, 4)
		for i, token := range tokens {
			sources[i] = &testSource{current: settings.Standard(), id: i}
			notes[i] = noteOf(t, token, sources[i])
			sources[i].trace = trace
			sources[i].reset(settings.Standard())
		}
		seq := []voiceleading.Simultaneity{
			{Notes: chord.MustNew(notes[0], notes[1]), Duration: quarterDuration()},
			{Notes: chord.MustNew(notes[2], notes[3]), Duration: quarterDuration()},
		}
		got, err := voiceleading.Decompose(seq, nil)
		require.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2, 3}, trace.values)
		require.Len(t, got.Lines, 2)
		assertLineElements(t, quarters("C4", "D4"), got.Lines[0])
		assertLineElements(t, quarters("E4", "G4"), got.Lines[1])
		// The notes of the lines keep their sources.
		n, ok := got.Lines[0].MustElement(0).MustNote()
		require.True(t, ok)
		assert.True(t, n.SameSource(notes[1]))
	})
	t.Run("an explicit system reads no sources", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		seq := []voiceleading.Simultaneity{{Notes: chord.MustNew(noteOf(t, "E4", src), noteOf(t, "C4", nil)), Duration: quarterDuration()}}
		src.reset(nil)
		got, err := voiceleading.Decompose(seq, traditionalSystem())
		require.NoError(t, err)
		assert.Zero(t, src.count())
		assertLineElements(t, quarters("C4"), got.Lines[0])
	})
}

func TestDecomposeInOrder(t *testing.T) {
	t.Parallel()
	q := quarterDuration()
	// RZ-5: crossings are kept and no system is needed.
	seq := []voiceleading.Simultaneity{simultaneity(t, q, "C3", "G4", "E4", "C5"), simultaneity(t, q, "C3", "F4", "F4", "C5")}
	lines, err := voiceleading.DecomposeInOrder(seq)
	require.NoError(t, err)
	require.Len(t, lines, 4)
	assertLineElements(t, quarters("C3", "C3"), lines[0])
	assertLineElements(t, quarters("G4", "F4"), lines[1])
	assertLineElements(t, quarters("E4", "F4"), lines[2])
	assertLineElements(t, quarters("C5", "C5"), lines[3])
	assert.Equal(t, lines, voiceleading.MustDecomposeInOrder(seq))
	a := voiceleading.MustAnalyze(lines, voiceleading.Options{System: traditionalSystem(), Order: voiceleading.LowToHigh})
	crossings := a.MustFindCrossings()
	require.Len(t, crossings, 1)
	assert.Equal(t, voiceleading.Pair{First: 1, Second: 2}, crossings[0].Pair)
	// Sources are never read, even when they would fail.
	src := &testSource{current: settings.Standard()}
	bound := []voiceleading.Simultaneity{
		{Notes: chord.MustNew(noteOf(t, "E4", src), noteOf(t, "C4", nil)), Duration: q},
		{Notes: chord.MustNew(), Duration: halfDuration()},
		{Notes: chord.MustNew(noteOf(t, "D4", nil), noteOf(t, "F4", settings.Standard())), Duration: q},
	}
	src.reset(nil)
	lines, err = voiceleading.DecomposeInOrder(bound)
	require.NoError(t, err)
	assert.Zero(t, src.count())
	assertLineElements(t, []expectedElement{{"E4", q}, {"-", halfDuration()}, {"D4", q}}, lines[0])
	assertLineElements(t, []expectedElement{{"C4", q}, {"-", halfDuration()}, {"F4", q}}, lines[1])
}

func TestDecompositionFailureOrder(t *testing.T) {
	t.Parallel()
	q := quarterDuration()
	tied := voice.MustTiedDuration(duration.MustParseTied("4~4"))
	half, err := note.FromPitch(pitch.MustFromASPN(spelling.MustParse("C+1/2"), 4), cents.Zero(), nil)
	require.NoError(t, err)
	cases := []struct {
		name    string
		seq     func(t *testing.T) []voiceleading.Simultaneity
		sys     *system.System
		both    failure
		inOrder *failure
	}{
		{"a nil sequence", func(*testing.T) []voiceleading.Simultaneity { return nil }, new(system.System),
			failure{voiceleading.ErrEmpty, voiceleading.ErrEmpty, 0, 0, false}, nil},
		{"an empty sequence", func(*testing.T) []voiceleading.Simultaneity { return []voiceleading.Simultaneity{} }, nil,
			failure{voiceleading.ErrEmpty, voiceleading.ErrEmpty, 0, 0, false}, nil},
		{"an unset system precedes the chords", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{{Duration: q}}
		}, new(system.System), failure{voiceleading.ErrUnset, system.ErrInvalid, 0, 0, false},
			&failure{voiceleading.ErrUnset, chord.ErrUnset, 0, -1, true}},
		{"an unset chord", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4"), {Duration: q}}
		}, nil, failure{voiceleading.ErrUnset, chord.ErrUnset, 1, -1, true}, nil},
		{"a note without an octave", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4"), {Notes: chord.MustNew(noteOf(t, "C4", nil), nameOf(t, "E")), Duration: q}}
		}, nil, failure{voiceleading.ErrNoOctave, voiceleading.ErrNoOctave, 1, 1, true}, nil},
		{"an unset duration of a chord", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{{Notes: chordOf(t, "C4", "E4")}}
		}, nil, failure{nil, voice.ErrUnset, 0, -1, true}, nil},
		{"a tied rest", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4"), {Notes: chord.MustNew(), Duration: tied}}
		}, nil, failure{nil, voice.ErrTiedRest, 1, -1, true}, nil},
		{"chords are checked before the count of voices", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4", "G4"), simultaneity(t, q, "C4", "E4"), {Notes: chord.MustNew(noteOf(t, "C4", nil), nameOf(t, "E")), Duration: q}}
		}, nil, failure{voiceleading.ErrNoOctave, voiceleading.ErrNoOctave, 2, 1, true}, nil},
		{"rests only", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{{Notes: chord.MustNew(), Duration: q}, {Notes: chord.MustNew(), Duration: q}}
		}, nil, failure{voiceleading.ErrNoVoices, voiceleading.ErrNoVoices, 0, 0, false}, nil},
		{"another count of voices than the first sounding chord", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{{Notes: chord.MustNew(), Duration: q}, simultaneity(t, q, "C4", "E4"), simultaneity(t, q, "C4", "E4"), simultaneity(t, q, "C4")}
		}, nil, failure{voiceleading.ErrVoiceCount, voiceleading.ErrVoiceCount, 3, -1, true}, nil},
		{"the count of voices precedes the system", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4"), simultaneity(t, q, "C4", "E4", "G4")}
		}, nil, failure{voiceleading.ErrVoiceCount, voiceleading.ErrVoiceCount, 1, -1, true}, nil},
		{"an unbound note without a system", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4")}
		}, nil, failure{voiceleading.ErrNoSystem, voiceleading.ErrNoSystem, 0, 0, true}, &failure{}},
		{"an inadmissible note under the explicit system", func(t *testing.T) []voiceleading.Simultaneity {
			return []voiceleading.Simultaneity{simultaneity(t, q, "C4", "E4"), {Notes: chord.MustNew(noteOf(t, "C4", nil), half), Duration: q}}
		}, traditionalSystem(), failure{nil, system.ErrInadmissible, 1, 1, true}, &failure{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			seq := tt.seq(t)
			got, err := voiceleading.Decompose(seq, tt.sys)
			assert.Equal(t, voiceleading.Decomposition{}, got)
			assertDecompositionFailure(t, err, tt.both)
			assertMustPanics(t, err, func() { voiceleading.MustDecompose(seq, tt.sys) })
			lines, err := voiceleading.DecomposeInOrder(seq)
			want := tt.both
			if tt.inOrder != nil {
				want = *tt.inOrder
			}
			if want.target == nil {
				require.NoError(t, err)
				return
			}
			assert.Nil(t, lines)
			assertDecompositionFailure(t, err, want)
			assertMustPanics(t, err, func() { voiceleading.MustDecomposeInOrder(seq) })
		})
	}
	// A tied duration is valid for a chord.
	lines, err := voiceleading.DecomposeInOrder([]voiceleading.Simultaneity{{Notes: chordOf(t, "C4"), Duration: tied}})
	require.NoError(t, err)
	assertLineElements(t, []expectedElement{{"C4", tied}}, lines[0])
}

// failure describes an expected error of a decomposition: the only sentinel
// (nil for none), the cause and, when inContext is set, the ChordError indices.
type failure struct {
	sentinel  error
	target    error
	chord     int
	noteIndex int
	inContext bool
}

func assertDecompositionFailure(t *testing.T, err error, want failure) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, want.target)
	assertSentinel(t, err, want.sentinel)
	if want.inContext {
		assertChordError(t, err, want.target, want.chord, want.noteIndex)
	} else {
		assertNoChordError(t, err)
	}
}

func TestDecomposeSystemFailures(t *testing.T) {
	t.Parallel()
	t.Run("an unbound note fails before any read", func(t *testing.T) {
		t.Parallel()
		src := &testSource{current: settings.Standard()}
		seq := []voiceleading.Simultaneity{
			{Notes: chord.MustNew(noteOf(t, "C4", src), noteOf(t, "E4", src)), Duration: quarterDuration()},
			{Notes: chord.MustNew(noteOf(t, "C4", src), noteOf(t, "E4", nil)), Duration: quarterDuration()},
		}
		src.reset(nil)
		_, err := voiceleading.Decompose(seq, nil)
		assertChordError(t, err, voiceleading.ErrNoSystem, 1, 1)
		assert.Zero(t, src.count())
	})
	t.Run("all reads precede the comparison of systems", func(t *testing.T) {
		t.Parallel()
		sources := []*testSource{{current: settings.Standard()}, {current: settings.Standard()}, {current: settings.Standard()}, {current: settings.Standard()}}
		notes := make([]note.Note, 4)
		for i, token := range []string{"C4", "E4", "C4", "E4"} {
			notes[i] = noteOf(t, token, sources[i])
		}
		sources[0].reset(settings.Standard())
		sources[1].reset(settings19(t))
		sources[2].reset(settings.Standard())
		sources[3].reset(nil)
		seq := []voiceleading.Simultaneity{
			{Notes: chord.MustNew(notes[0], notes[1]), Duration: quarterDuration()},
			{Notes: chord.MustNew(notes[2], notes[3]), Duration: quarterDuration()},
		}
		_, err := voiceleading.Decompose(seq, nil)
		assertChordError(t, err, note.ErrUnsetSettings, 1, 1)
		assertSentinel(t, err, nil)
		for _, src := range sources {
			assert.Equal(t, 1, src.count())
		}
	})
	t.Run("different systems", func(t *testing.T) {
		t.Parallel()
		one := &testSource{current: settings.Standard()}
		two := &testSource{current: settings.Standard()}
		seq := []voiceleading.Simultaneity{{Notes: chord.MustNew(noteOf(t, "C4", one), noteOf(t, "E4", two)), Duration: quarterDuration()}}
		one.reset(settings.Standard())
		two.reset(settings19(t))
		_, err := voiceleading.Decompose(seq, nil)
		assertChordError(t, err, voiceleading.ErrDifferentSystems, 0, 1)
		assertSentinel(t, err, voiceleading.ErrDifferentSystems)
	})
}
