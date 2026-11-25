package voiceleading_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: architecture 3 (labels), 12 (ownership, equality,
// concurrency) and 13 (errors, Must forms).

func TestExtensionLabels(t *testing.T) {
	t.Parallel()
	cases := []struct {
		got  string
		want string
	}{
		{voiceleading.LeapUnset.String(), "unset"},
		{voiceleading.LeapFromThird.String(), "from-third"},
		{voiceleading.LeapFromFourth.String(), "from-fourth"},
		{voiceleading.LeapDefinition(9).String(), "unknown"},
		{voiceleading.MelodyRuleUnset.String(), "unset"},
		{voiceleading.OMTLeapRecovery.String(), "omt-leap-recovery"},
		{voiceleading.OMTMelodicIntervals.String(), "omt-melodic-intervals"},
		{voiceleading.FuxAllowedMotions.String(), "wp-fux-allowed-motions"},
		{voiceleading.MT21ChoraleMelody.String(), "mt21-chorale-melody"},
		{voiceleading.MelodyRule(9).String(), "unknown"},
		{voiceleading.SpacingUnset.String(), "unset"},
		{voiceleading.SpacingME.String(), "me"},
		{voiceleading.SpacingAM.String(), "am"},
		{voiceleading.SpacingOMT.String(), "omt"},
		{voiceleading.SpacingDefinition(9).String(), "unknown"},
		{voiceleading.SpacingClassUnset.String(), "unset"},
		{voiceleading.CloseSpacing.String(), "close"},
		{voiceleading.OpenSpacing.String(), "open"},
		{voiceleading.MixedSpacing.String(), "mixed"},
		{voiceleading.OMTClosedSpacing.String(), "omt-closed"},
		{voiceleading.OMTOpenSpacing.String(), "omt-open"},
		{voiceleading.SpacingClass(9).String(), "unknown"},
		{voiceleading.DistanceUnset.String(), "unset"},
		{voiceleading.DistanceOMTSATB.String(), "omt-satb"},
		{voiceleading.DistanceMT21SATB.String(), "mt21-satb"},
		{voiceleading.DistanceOMTFirstSpecies.String(), "omt-first-species"},
		{voiceleading.DistanceCustom.String(), "custom"},
		{voiceleading.DistanceRule(9).String(), "unknown"},
		{voiceleading.BySpelling.String(), "spelling"},
		{voiceleading.BySound.String(), "sound"},
		{voiceleading.Reading(9).String(), "unknown"},
		{voiceleading.NoDoubling.String(), "none"},
		{voiceleading.UnisonDoubling.String(), "unison"},
		{voiceleading.OctaveDoubling.String(), "octave"},
		{voiceleading.UnisonAndOctaveDoubling.String(), "unison-and-octave"},
		{voiceleading.DoublingKind(9).String(), "unknown"},
		{voiceleading.ConnectionUnset.String(), "unset"},
		{voiceleading.HarmonicConnection.String(), "harmonic"},
		{voiceleading.MelodicConnection.String(), "melodic"},
		{voiceleading.ConnectionKind(9).String(), "unknown"},
		{voiceleading.MT21VariantUnset.String(), "unset"},
		{voiceleading.MT21CommonToneAndSteps.String(), "26.7.2-common-tone"},
		{voiceleading.MT21SameDirection.String(), "26.7.2-same-direction"},
		{voiceleading.MT21TwoCommonTones.String(), "26.7.1-two-common-tones"},
		{voiceleading.MT21OneCommonTone.String(), "26.13-one-common-tone"},
		{voiceleading.MT21Variant(9).String(), "unknown"},
	}
	for _, tt := range cases {
		assert.Equal(t, tt.want, tt.got)
	}
	reasons := map[string]string{
		voiceleading.ReasonNoStepBack: "no-step-back", voiceleading.ReasonEndOfLine: "end-of-line",
		voiceleading.ReasonRestFollows: "rest-follows", voiceleading.ReasonAugmented: "augmented",
		voiceleading.ReasonDiminished: "diminished", voiceleading.ReasonSeventh: "seventh",
		voiceleading.ReasonSixth: "sixth", voiceleading.ReasonWiderThanOctave: "wider-than-octave",
		voiceleading.ReasonNotAllowed: "not-allowed", voiceleading.ReasonSameDirection: "same-direction",
		voiceleading.ReasonRepeat: "repeat", voiceleading.ReasonSilentParts: "silent-parts",
		voiceleading.ReasonVoiceCount: "voice-count", voiceleading.ReasonCrossedVoices: "crossed-voices",
		voiceleading.ReasonFromNotThreeClasses: "from-not-three-classes", voiceleading.ReasonToNotThreeClasses: "to-not-three-classes",
		voiceleading.ReasonFromBassNotDoubled: "from-bass-not-doubled", voiceleading.ReasonToBassNotDoubled: "to-bass-not-doubled",
		voiceleading.ReasonBassMotion: "bass-motion-not-covered",
	}
	for got, want := range reasons {
		assert.Equal(t, want, got)
	}
	assert.Equal(t, "mt21-root-position-triads", voiceleading.MT21ConnectionRule)
	assert.Equal(t, voiceleading.BySpelling, voiceleading.Reading(0), "the zero reading is by spelling")
}

func TestNewSentinelsAndChordError(t *testing.T) {
	t.Parallel()
	for i, a := range allSentinels {
		assert.Contains(t, a.Error(), "voiceleading: ")
		for j, b := range allSentinels {
			assert.Equal(t, i == j, errors.Is(a, b), "%v and %v", a, b)
		}
	}
	var nilError *voiceleading.ChordError
	assert.Contains(t, nilError.Error(), "voiceleading: ")
	assert.NoError(t, nilError.Unwrap())
	empty := &voiceleading.ChordError{}
	assert.Contains(t, empty.Error(), "voiceleading: ")
	assert.NoError(t, empty.Unwrap())
	wrapped := &voiceleading.ChordError{Chord: 2, Note: -1, Err: voiceleading.ErrVoiceCount}
	assert.Contains(t, wrapped.Error(), "voiceleading: ")
	assert.Equal(t, voiceleading.ErrVoiceCount, wrapped.Unwrap())
	assert.ErrorIs(t, wrapped, voiceleading.ErrVoiceCount)
}

type extensionOperation struct {
	name         string
	normal, must any
	args         []any
}

func melodyOperations(m *voiceleading.Melody) []extensionOperation {
	return []extensionOperation{
		{"Elements", m.Elements, m.MustElements, nil},
		{"System", m.System, m.MustSystem, nil},
		{"Moves", m.Moves, m.MustMoves, nil},
		{"LeapRuns", m.LeapRuns, m.MustLeapRuns, []any{voiceleading.LeapFromThird, 1}},
		{"Check", m.Check, m.MustCheck, []any{voiceleading.OMTMelodicIntervals}},
	}
}

func sonorityOperations(t *testing.T, s *voiceleading.Sonority) []extensionOperation {
	return []extensionOperation{
		{"Len", s.Len, s.MustLen, nil},
		{"Notes", s.Notes, s.MustNotes, nil},
		{"Parts", s.Parts, s.MustParts, nil},
		{"Silent", s.Silent, s.MustSilent, nil},
		{"Order", s.Order, s.MustOrder, nil},
		{"System", s.System, s.MustSystem, nil},
		{"Spacing", s.Spacing, s.MustSpacing, nil},
		{"ClassifySpacing", s.ClassifySpacing, s.MustClassifySpacing, []any{voiceleading.SpacingAM}},
		{"CheckDistances", s.CheckDistances, s.MustCheckDistances, []any{voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB}}},
		{"Doublings", s.Doublings, s.MustDoublings, []any{voiceleading.BySound, namesChord(t, "C", "E", "G")}},
	}
}

func analysisExtensionOperations(a *voiceleading.Analysis) []extensionOperation {
	return []extensionOperation{
		{"Melody", a.Melody, a.MustMelody, []any{1}},
		{"Sonorities", a.Sonorities, a.MustSonorities, nil},
		{"SonorityAt", a.SonorityAt, a.MustSonorityAt, []any{pos(1, 4)}},
		{"Connections", a.Connections, a.MustConnections, nil},
		{"ConnectionAt", a.ConnectionAt, a.MustConnectionAt, []any{pos(0, 1), pos(1, 4)}},
	}
}

func assertUnsetOperation(t *testing.T, op extensionOperation) {
	t.Helper()
	values := call(op.normal, op.args)
	err, _ := values[len(values)-1].Interface().(error)
	assertSentinel(t, err, voiceleading.ErrUnset)
	for i, v := range values[:len(values)-1] {
		assert.True(t, v.IsZero(), "%s result %d must be zero", op.name, i)
	}
	assertMustPanics(t, err, func() { call(op.must, op.args) })
}

func TestUnsetSnapshotsFailEveryOperation(t *testing.T) {
	t.Parallel()
	for _, m := range []*voiceleading.Melody{nil, new(voiceleading.Melody)} {
		assert.False(t, m.IsSet())
		for _, op := range melodyOperations(m) {
			t.Run("Melody."+op.name, func(t *testing.T) { assertUnsetOperation(t, op) })
		}
	}
	for _, s := range []*voiceleading.Sonority{nil, new(voiceleading.Sonority)} {
		assert.False(t, s.IsSet())
		for _, op := range sonorityOperations(t, s) {
			t.Run("Sonority."+op.name, func(t *testing.T) { assertUnsetOperation(t, op) })
		}
	}
	for _, a := range []*voiceleading.Analysis{nil, new(voiceleading.Analysis)} {
		for _, op := range analysisExtensionOperations(a) {
			t.Run("Analysis."+op.name, func(t *testing.T) { assertUnsetOperation(t, op) })
		}
	}
	// Empty data are set snapshots.
	assert.True(t, voiceleading.MustNewMelody(voice.MustNew(), nil).IsSet())
	assert.True(t, voiceleading.MustNewSonority(chord.MustNew(), nil).IsSet())
}

func TestMustFormsReturnTheOrdinaryResults(t *testing.T) {
	t.Parallel()
	m := melodyOf(t, "C4", "A4", "B4", "-", "E4", "F#4")
	s := sonorityOf(t, "C3", "E4", "G4", "C5")
	a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3", "D3"), line(t, "E4", "F4", "G4"))
	var ops []extensionOperation
	ops = append(ops, melodyOperations(m)...)
	ops = append(ops, sonorityOperations(t, s)...)
	ops = append(ops, analysisExtensionOperations(a)...)
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			normal := call(op.normal, op.args)
			require.True(t, normal[len(normal)-1].IsNil(), "%s: %v", op.name, normal[len(normal)-1])
			must := call(op.must, op.args)
			require.Len(t, must, len(normal)-1)
			for i := range must {
				assert.Equal(t, normal[i].Interface(), must[i].Interface(), "%s result %d", op.name, i)
			}
		})
	}
	from, to := sonorityOf(t, "C3", "E4", "G4", "C5"), sonorityOf(t, "F3", "F4", "A4", "C5")
	c, err := voiceleading.Connect(from, to)
	require.NoError(t, err)
	assert.Equal(t, c, voiceleading.MustConnect(from, to))
	mt, err := voiceleading.CheckMT21Connection(from, to)
	require.NoError(t, err)
	assert.Equal(t, mt, voiceleading.MustCheckMT21Connection(from, to))
	n, err := voiceleading.NewSonority(chordOf(t, "C4", "E4"), nil)
	require.Error(t, err) // Unbound notes need a system: Must panics with the same error.
	assert.Nil(t, n)
	assertMustPanics(t, err, func() { voiceleading.MustNewSonority(chordOf(t, "C4", "E4"), nil) })
}

func TestSnapshotsOwnTheirData(t *testing.T) {
	t.Parallel()
	t.Run("melody", func(t *testing.T) {
		t.Parallel()
		m := melodyOf(t, "C4", "A4", "B4", "-", "E4", "F#4")
		elements, moves := m.MustElements(), m.MustMoves()
		check := m.MustCheck(voiceleading.OMTLeapRecovery)
		runs := m.MustLeapRuns(voiceleading.LeapFromThird, 1)
		wantElements, wantMoves := m.MustElements(), m.MustMoves()
		wantCheck := m.MustCheck(voiceleading.OMTLeapRecovery)
		wantRuns := m.MustLeapRuns(voiceleading.LeapFromThird, 1)
		require.NotEmpty(t, check.Violations)
		elements[0].Index = 99
		moves[0].ThroughRest = true
		moves[0].Chromatic = 99
		check.Violations[0].Reasons[0] = "changed"
		check.Violations[0].Move.Chromatic = 99
		check.Rule = "changed"
		runs[0].First = 99
		assert.Equal(t, wantElements, m.MustElements())
		assert.Equal(t, wantMoves, m.MustMoves())
		assert.Equal(t, wantCheck, m.MustCheck(voiceleading.OMTLeapRecovery))
		assert.Equal(t, wantRuns, m.MustLeapRuns(voiceleading.LeapFromThird, 1))
	})
	t.Run("sonority and its inputs", func(t *testing.T) {
		t.Parallel()
		notes := chordOf(t, "C3", "E4", "G4", "C5")
		s := voiceleading.MustNewSonority(notes, traditionalSystem())
		want := sonorityOf(t, "C3", "E4", "G4", "C5")
		require.NoError(t, notes.Add(noteOf(t, "E5", nil)))
		require.NoError(t, notes.ReplaceAt(0, noteOf(t, "D3", nil)))
		assert.True(t, want.Equal(s))
		got := s.MustNotes()
		require.NoError(t, got[0].SetCents(cents.MustParse("30")))
		got[1] = noteOf(t, "B7", nil)
		spacing := s.MustSpacing()
		spacing.Adjacent[0].Lower = 99
		classes := s.MustClassifySpacing(voiceleading.SpacingME)
		classes.Considered[0].Chromatic = 99
		distances := voiceleading.MustNewSonority(chordOf(t, "C3", "G3", "C4", "E5"), traditionalSystem())
		check := distances.MustCheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB})
		require.NotEmpty(t, check.Violations)
		check.Violations[0].Pair.Lower = 99
		doublings := s.MustDoublings(voiceleading.BySpelling, namesChord(t, "C", "E", "G", "B"))
		doublings.Classes[0].Voices[0] = 99
		doublings.Missing[0].Voices = append(doublings.Missing[0].Voices, 99)
		doubled := doublings.Doubled()
		require.NotEmpty(t, doubled)
		doubled[0].Voices[1] = 99
		assert.True(t, want.Equal(s))
		assert.Equal(t, want.MustNotes(), s.MustNotes())
		assert.Equal(t, want.MustSpacing(), s.MustSpacing())
		assert.Equal(t, want.MustClassifySpacing(voiceleading.SpacingME), s.MustClassifySpacing(voiceleading.SpacingME))
		assert.Equal(t, 2, distances.MustCheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}).Violations[0].Pair.Lower)
		fresh := s.MustDoublings(voiceleading.BySpelling, namesChord(t, "C", "E", "G", "B"))
		assert.Equal(t, []int{0, 3}, fresh.Classes[0].Voices)
		assert.Empty(t, fresh.Missing[0].Voices)
		assert.Equal(t, []int{0, 3}, fresh.Doubled()[0].Voices)
		// Doubled returns a new slice: changing it does not change Classes.
		again := s.MustDoublings(voiceleading.BySpelling, chord.Chord{})
		d := again.Doubled()
		d[0].Voices = nil
		assert.Equal(t, []int{0, 3}, again.Classes[0].Voices)
	})
	t.Run("part sonorities and connections", func(t *testing.T) {
		t.Parallel()
		a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3", "-", "F3"), line(t, "E4", "F4", "A4"), line(t, "C5", "C5", "C5"))
		ps := a.MustSonorities()
		want := a.MustSonorities()
		ps[0].Elements[0].Index = 99
		ps[0].Voices = nil
		silent := want[1].Voices.MustSilent()
		silent[0] = 99
		parts, _ := want[1].Voices.MustParts()
		parts[0] = 99
		assert.Equal(t, want, a.MustSonorities())
		assert.Equal(t, []int{0}, want[1].Voices.MustSilent())
		gotParts, _ := want[1].Voices.MustParts()
		assert.Equal(t, []int{1, 2}, gotParts)
		cs := a.MustConnections()
		require.Empty(t, cs) // Every pair of adjacent moments has a silent bass.
		c := voiceleading.MustConnect(sonorityOf(t, "C3", "E4", "G4", "C5"), sonorityOf(t, "F3", "F4", "A4", "C5"))
		from := c.From.MustNotes()
		require.NoError(t, c.Moves[0].From.SetCents(cents.MustParse("40")))
		c.Common = nil
		assert.Equal(t, from, c.From.MustNotes())
	})
	t.Run("decomposition", func(t *testing.T) {
		t.Parallel()
		seq := []voiceleading.Simultaneity{simultaneity(t, quarterDuration(), "C4", "C4"), simultaneity(t, quarterDuration(), "E4", "G4")}
		got := voiceleading.MustDecompose(seq, traditionalSystem())
		want := voiceleading.MustDecompose(seq, traditionalSystem())
		require.NoError(t, seq[0].Notes.ReplaceAt(0, noteOf(t, "D4", nil)))
		seq[1].Duration = halfDuration()
		got.Ambiguities[0].Voices[0] = 99
		got.Lines[0] = nil
		assertLineElements(t, quarters("C4", "E4"), want.Lines[0])
		assert.Equal(t, []int{0, 1}, want.Ambiguities[0].Voices)
		inOrder := voiceleading.MustDecomposeInOrder([]voiceleading.Simultaneity{simultaneity(t, quarterDuration(), "C4", "E4")})
		inOrder[0] = nil
		assert.Len(t, voiceleading.MustDecomposeInOrder([]voiceleading.Simultaneity{simultaneity(t, quarterDuration(), "C4", "E4")}), 2)
	})
}

func TestMelodyEquality(t *testing.T) {
	t.Parallel()
	var unset *voiceleading.Melody
	set := melodyOf(t, "C4", "D4")
	assert.True(t, unset.Equal(nil))
	assert.True(t, unset.Equal(new(voiceleading.Melody)))
	assert.True(t, new(voiceleading.Melody).Equal(unset))
	assert.False(t, unset.Equal(set))
	assert.False(t, set.Equal(unset))
	assert.True(t, set.Equal(set))
	assert.True(t, set.Equal(melodyOf(t, "C4", "D4")))
	assert.False(t, set.Equal(melodyOf(t, "C4", "E4")))
	assert.False(t, set.Equal(melodyOf(t, "C4", "-", "D4")))
	assert.False(t, set.Equal(voiceleading.MustNewMelody(line(t, "C4", "D4"), system.Positions19())))
	longer := voice.MustNew(sound(t, "C", 4, 1, 2, nil), sound(t, "D", 4, 1, 4, nil))
	assert.False(t, set.Equal(voiceleading.MustNewMelody(longer, traditionalSystem())))
	// The presence of a system takes part in equality.
	one := line(t, "C4")
	assert.False(t, voiceleading.MustNewMelody(one, nil).Equal(voiceleading.MustNewMelody(one, traditionalSystem())))
	assert.True(t, voiceleading.MustNewMelody(one, nil).Equal(voiceleading.MustNewMelody(one, nil)))
	assert.False(t, reflect.TypeFor[voiceleading.Melody]().Comparable())
	assert.False(t, reflect.TypeFor[voiceleading.Sonority]().Comparable())
	// A copy by assignment shares only immutable data.
	copied := *set
	assert.True(t, copied.Equal(set))
}

func TestConcurrentUseOfSnapshotsAndFunctions(t *testing.T) {
	t.Parallel()
	m := melodyOf(t, "C4", "A4", "B4", "-", "E4", "F#4")
	s := sonorityOf(t, "C3", "E4", "G4", "C5")
	to := sonorityOf(t, "F3", "F4", "A4", "C5")
	a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3", "D3"), line(t, "E4", "F4", "G4"))
	target := namesChord(t, "F", "A", "C")
	bass := noteOf(t, "F3", nil)
	conds := voiceleading.ConnectionConditions{Complete: true, NoCrossings: true}
	seq := []voiceleading.Simultaneity{simultaneity(t, quarterDuration(), "G4", "C3", "C5", "E4"), simultaneity(t, quarterDuration(), "B4", "G2", "G4", "D4")}
	partMelody := voiceleading.MustNewMelody(line(t, "E4", "F4", "G4"), traditionalSystem())
	type observation struct {
		moves       []voiceleading.Move
		check       *voiceleading.MelodyAssessment
		spacing     voiceleading.SpacingClassification
		doublings   voiceleading.Doublings
		distances   *voiceleading.DistanceAssessment
		sonorities  []voiceleading.PartSonority
		connections []voiceleading.PartConnection
		melody      bool
		connection  voiceleading.Connection
		mt21        *voiceleading.ConnectionAssessment
		choices     []voiceleading.Connection
		lines       voiceleading.Decomposition
		errs        []error
	}
	observe := func() observation {
		var o observation
		var err error
		record := func(e error) { o.errs = append(o.errs, e) }
		o.moves, err = m.Moves()
		record(err)
		o.check, err = m.Check(voiceleading.MT21ChoraleMelody)
		record(err)
		o.spacing, err = s.ClassifySpacing(voiceleading.SpacingAM)
		record(err)
		o.doublings, err = s.Doublings(voiceleading.BySound, target)
		record(err)
		o.distances, err = s.CheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB})
		record(err)
		o.sonorities, err = a.Sonorities()
		record(err)
		o.connections, err = a.Connections()
		record(err)
		melody, err := a.Melody(1)
		record(err)
		o.melody = melody.Equal(partMelody)
		o.connection, err = voiceleading.Connect(s, to)
		record(err)
		o.mt21, err = voiceleading.CheckMT21Connection(s, to)
		record(err)
		o.choices, err = voiceleading.ChooseConnections(s, target, bass, conds)
		record(err)
		o.lines, err = voiceleading.Decompose(seq, traditionalSystem())
		record(err)
		return o
	}
	expected := observe()
	for _, err := range expected.errs {
		require.NoError(t, err)
	}
	require.True(t, expected.melody)
	results := make(chan observation, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { results <- observe() })
	}
	wg.Wait()
	close(results)
	for got := range results {
		assert.Equal(t, expected, got)
	}
}

func TestExistingOperationsKeepTheirBehavior(t *testing.T) {
	t.Parallel()
	// The v1 surface is unchanged: Analyze still measures only the pairwise
	// verticals, so a tail with an invalid spelling under the explicit system
	// and parts without a system remain valid input.
	high := voice.MustNew(sound(t, "G", 4, 1, 4, nil), sound(t, "C+1/2", 4, 1, 4, nil))
	a, err := voiceleading.Analyze([]*voice.Line{line(t, "C4"), high}, voiceleading.Options{System: traditionalSystem()})
	require.NoError(t, err)
	assert.Len(t, a.MustVerticals(pair), 1)
	b := voiceleading.MustAnalyze([]*voice.Line{line(t, "C4", "-"), line(t, "-", "G4")}, voiceleading.Options{})
	sys, ok := b.MustSystem()
	assert.Nil(t, sys)
	assert.False(t, ok)
}
