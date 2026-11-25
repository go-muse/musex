package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/problemsolving/voiceleading"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Examples are business-spec SK-3, SK-4, SK-4a and SK-5 line by line; the
// identifiers, sources, scopes and the order of reasons are architecture 4.4.

type wantViolation struct {
	index   int
	reasons []string
	hasNext bool
}

func assertMelodyViolations(t *testing.T, m *voiceleading.Melody, want []wantViolation, got []voiceleading.MelodyViolation) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	moves := m.MustMoves()
	for i, w := range want {
		v := got[i]
		assert.Equal(t, w.index, v.Index, "violation %d", i)
		assert.Equal(t, w.reasons, v.Reasons, "violation %d", i)
		assert.Equal(t, w.hasNext, v.HasNext, "violation %d", i)
		require.Less(t, w.index, len(moves))
		assert.Equal(t, moves[w.index], v.Move, "violation %d move", i)
		if w.hasNext {
			require.Less(t, w.index+1, len(moves))
			assert.Equal(t, moves[w.index+1], v.Next, "violation %d next", i)
		} else {
			assert.Equal(t, voiceleading.Move{}, v.Next, "violation %d next must be zero", i)
		}
	}
}

func none() []wantViolation { return []wantViolation{} }

func TestMelodyNormIdentity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		rule                 voiceleading.MelodyRule
		label, source, scope string
	}{
		{voiceleading.OMTLeapRecovery, "omt-leap-recovery", "Open Music Theory, cantus firmus",
			"one voice; a leap of a fourth or wider is followed by a step in the opposite direction within the same fragment"},
		{voiceleading.OMTMelodicIntervals, "omt-melodic-intervals", "Open Music Theory, rules for all species",
			"one voice; no augmented, diminished, seventh, sixth other than an ascending minor sixth, or wider than an octave"},
		{voiceleading.FuxAllowedMotions, "wp-fux-allowed-motions", "Wikipedia, Counterpoint, considerations for all species",
			"one voice; only P1, P4, P5, P8, minor and major seconds and thirds, and an ascending minor sixth followed by downward motion"},
		{voiceleading.MT21ChoraleMelody, "mt21-chorale-melody", "Music Theory for the 21st-Century Classroom, 26.5",
			"one voice; no augmented intervals; the direction changes after a leap wider than a fourth or a diminished interval"},
	}
	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.label, tt.rule.String())
			for _, tokens := range [][]string{nil, {"-", "-"}, {"C4"}, {"C4", "D4"}} {
				m := melodyOf(t, tokens...)
				a, err := m.Check(tt.rule)
				require.NoError(t, err)
				require.NotNil(t, a)
				assert.Equal(t, tt.label, a.Rule)
				assert.Equal(t, tt.source, a.Source)
				assert.Equal(t, tt.scope, a.Scope)
				assert.True(t, a.Applicable)
				assert.NotNil(t, a.Violations)
				assert.Empty(t, a.Violations)
				assert.Equal(t, a, m.MustCheck(tt.rule))
			}
		})
	}
	m := melodyOf(t, "C4", "A4", "B4")
	for _, rule := range []voiceleading.MelodyRule{voiceleading.MelodyRuleUnset, voiceleading.MelodyRule(99)} {
		a, err := m.Check(rule)
		assert.Nil(t, a)
		assertSentinel(t, err, voiceleading.ErrOption)
		assertMustPanics(t, err, func() { m.MustCheck(rule) })
	}
	assert.Equal(t, "unset", voiceleading.MelodyRuleUnset.String())
	assert.Equal(t, "unknown", voiceleading.MelodyRule(99).String())
	// A new assessment is returned on every call.
	first, second := m.MustCheck(voiceleading.OMTLeapRecovery), m.MustCheck(voiceleading.OMTLeapRecovery)
	assert.NotSame(t, first, second)
}

func TestMelodyNorms(t *testing.T) {
	t.Parallel()
	recovery := voiceleading.OMTLeapRecovery
	intervals := voiceleading.OMTMelodicIntervals
	fux := voiceleading.FuxAllowedMotions
	chorale := voiceleading.MT21ChoraleMelody
	cases := []struct {
		name   string
		rule   voiceleading.MelodyRule
		tokens []string
		want   []wantViolation
	}{
		// SK-3: leap recovery.
		{"SK-3 C4 A4 B4", recovery, []string{"C4", "A4", "B4"}, []wantViolation{{0, []string{voiceleading.ReasonNoStepBack}, true}}},
		{"SK-3 C4 A4 G4", recovery, []string{"C4", "A4", "G4"}, none()},
		{"SK-3 C4 D4 G4 ends the line", recovery, []string{"C4", "D4", "G4"}, []wantViolation{{1, []string{voiceleading.ReasonEndOfLine}, false}}},
		{"SK-3 C4 A4 A4 a repeat is not a step", recovery, []string{"C4", "A4", "A4"}, []wantViolation{{0, []string{voiceleading.ReasonNoStepBack}, true}}},
		{"SK-3 a rest follows the leap", recovery, []string{"C4", "G4", "-"}, []wantViolation{{0, []string{voiceleading.ReasonRestFollows}, false}}},
		{"SK-3 a rest ends the fragment before the next sound", recovery, []string{"C4", "G4", "-", "F4"}, []wantViolation{{0, []string{voiceleading.ReasonRestFollows}, false}}},
		{"SK-3 a leap through a rest is not checked", recovery, []string{"C4", "-", "A4", "G4"}, none()},
		{"SK-3 a third is not checked whatever the leap definition", recovery, []string{"C4", "E4", "F4"}, none()},
		{"SK-3 a step back after a descending leap", recovery, []string{"C5", "G4", "A4"}, none()},
		{"SK-3 a leap back is not a step back", recovery, []string{"C4", "G4", "C4"}, []wantViolation{{0, []string{voiceleading.ReasonNoStepBack}, true}, {1, []string{voiceleading.ReasonEndOfLine}, false}}},
		// SK-4: melodic intervals of counterpoint.
		{"SK-4 C4 B4 seventh", intervals, []string{"C4", "B4"}, []wantViolation{{0, []string{voiceleading.ReasonSeventh}, false}}},
		{"SK-4 C5 E4 descending minor sixth", intervals, []string{"C5", "E4"}, []wantViolation{{0, []string{voiceleading.ReasonSixth}, false}}},
		{"SK-4 C4 A4 major sixth", intervals, []string{"C4", "A4"}, []wantViolation{{0, []string{voiceleading.ReasonSixth}, false}}},
		{"SK-4 F4 B4 augmented fourth", intervals, []string{"F4", "B4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented}, false}}},
		{"SK-4 B3 F4 diminished fifth", intervals, []string{"B3", "F4"}, []wantViolation{{0, []string{voiceleading.ReasonDiminished}, false}}},
		{"SK-4 C4 D5 wider than an octave", intervals, []string{"C4", "D5"}, []wantViolation{{0, []string{voiceleading.ReasonWiderThanOctave}, false}}},
		{"SK-4 E4 C5 ascending minor sixth", intervals, []string{"E4", "C5"}, none()},
		{"SK-4 C4 C5 octave", intervals, []string{"C4", "C5"}, none()},
		{"SK-4 C4 G4 fifth", intervals, []string{"C4", "G4"}, none()},
		{"SK-4 C4 D4 second", intervals, []string{"C4", "D4"}, none()},
		{"SK-4 augmented unison", intervals, []string{"C4", "C#4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented}, false}}},
		{"SK-4 augmented second", intervals, []string{"C4", "D#4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented}, false}}},
		{"SK-4 diminished second", intervals, []string{"C#4", "Db4"}, []wantViolation{{0, []string{voiceleading.ReasonDiminished}, false}}},
		{"SK-4 augmented seventh lists both reasons", intervals, []string{"C4", "B#4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented, voiceleading.ReasonSeventh}, false}}},
		{"SK-4 augmented sixth lists both reasons", intervals, []string{"C4", "A#4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented, voiceleading.ReasonSixth}, false}}},
		{"SK-4 diminished seventh lists both reasons", intervals, []string{"C#4", "Bb4"}, []wantViolation{{0, []string{voiceleading.ReasonDiminished, voiceleading.ReasonSeventh}, false}}},
		{"SK-4 augmented fifteenth lists both reasons", intervals, []string{"C4", "C#6"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented, voiceleading.ReasonWiderThanOctave}, false}}},
		{"SK-4 a move through a rest is not checked", intervals, []string{"C4", "-", "B4"}, none()},
		{"SK-4 every move is checked", intervals, []string{"C4", "B4", "C4", "D4", "B4"}, []wantViolation{
			{0, []string{voiceleading.ReasonSeventh}, false},
			{1, []string{voiceleading.ReasonSeventh}, false},
			{3, []string{voiceleading.ReasonSixth}, false},
		}},
		// SK-4a: motions allowed by Fux.
		{"SK-4a C4 Ab4 F4", fux, []string{"C4", "Ab4", "F4"}, none()},
		{"SK-4a C4 G4 A4", fux, []string{"C4", "G4", "A4"}, none()},
		{"SK-4a E4 C5 D5 same direction", fux, []string{"E4", "C5", "D5"}, []wantViolation{{0, []string{voiceleading.ReasonSameDirection}, true}}},
		{"SK-4a C4 F#4 not allowed", fux, []string{"C4", "F#4"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a C4 B4 not allowed", fux, []string{"C4", "B4"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a a repeat after the minor sixth", fux, []string{"E4", "C5", "C5"}, []wantViolation{{0, []string{voiceleading.ReasonRepeat}, true}}},
		{"SK-4a nothing follows the minor sixth", fux, []string{"E4", "C5"}, none()},
		{"SK-4a a rest follows the minor sixth", fux, []string{"E4", "C5", "-", "D5"}, none()},
		{"SK-4a every allowed motion", fux, []string{"C4", "C4", "F4", "C5", "C4", "D4", "Eb4", "G4", "Bb4", "F4", "G4"}, none()},
		{"SK-4a a descending minor sixth", fux, []string{"C5", "E4"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a a major sixth", fux, []string{"C4", "A4"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a an augmented unison", fux, []string{"C4", "C#4"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a an augmented second", fux, []string{"C4", "D#4"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a a twelfth", fux, []string{"C4", "G5"}, []wantViolation{{0, []string{voiceleading.ReasonNotAllowed}, false}}},
		{"SK-4a a move through a rest is not checked", fux, []string{"C4", "-", "B4"}, none()},
		// SK-5: chorale melody of MT21.
		{"SK-5 F4 G#4 augmented", chorale, []string{"F4", "G#4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented}, false}}},
		{"SK-5 C4 A4 B4", chorale, []string{"C4", "A4", "B4"}, []wantViolation{{0, []string{voiceleading.ReasonSameDirection}, true}}},
		{"SK-5 C4 A4 G4", chorale, []string{"C4", "A4", "G4"}, none()},
		{"SK-5 B3 F4 G4 after a diminished fifth", chorale, []string{"B3", "F4", "G4"}, []wantViolation{{0, []string{voiceleading.ReasonSameDirection}, true}}},
		{"SK-5 B3 F4 E4", chorale, []string{"B3", "F4", "E4"}, none()},
		{"SK-5 C4 F4 G4 a fourth is not wider than a fourth", chorale, []string{"C4", "F4", "G4"}, none()},
		{"SK-5 C4 G#4 A4 both kinds", chorale, []string{"C4", "G#4", "A4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented, voiceleading.ReasonSameDirection}, true}}},
		{"SK-5 a repeat after a leap", chorale, []string{"C4", "A4", "A4"}, []wantViolation{{0, []string{voiceleading.ReasonRepeat}, true}}},
		{"SK-5 the cadential bass G2 C2", chorale, []string{"G2", "C2"}, none()},
		{"SK-5 a rest follows the leap", chorale, []string{"C4", "A4", "-", "B4"}, none()},
		{"SK-5 a diminished second", chorale, []string{"C#4", "Db4", "Eb4"}, []wantViolation{{0, []string{voiceleading.ReasonSameDirection}, true}}},
		{"SK-5 an augmented unison", chorale, []string{"C4", "C#4"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented}, false}}},
		{"SK-5 a descending sixth then down", chorale, []string{"A4", "C4", "B3"}, []wantViolation{{0, []string{voiceleading.ReasonSameDirection}, true}}},
		{"SK-5 an augmented move with an unchecked size", chorale, []string{"F4", "B4", "C5"}, []wantViolation{{0, []string{voiceleading.ReasonAugmented}, false}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := melodyOf(t, tt.tokens...)
			a, err := m.Check(tt.rule)
			require.NoError(t, err)
			require.NotNil(t, a)
			assert.Equal(t, tt.rule.String(), a.Rule)
			assert.True(t, a.Applicable)
			assertMelodyViolations(t, m, tt.want, a.Violations)
		})
	}
}

func TestMelodyNormsOfAnalysisParts(t *testing.T) {
	t.Parallel()
	// Each part is checked on its own, including a tail beyond the other part.
	a := analyzeParts(t, voiceleading.LowToHigh, traditionalSystem(), line(t, "C3", "E3"), line(t, "C4", "A4", "B4"))
	got := a.MustMelody(1).MustCheck(voiceleading.OMTLeapRecovery)
	assertMelodyViolations(t, a.MustMelody(1), []wantViolation{{0, []string{voiceleading.ReasonNoStepBack}, true}}, got.Violations)
	assert.Empty(t, a.MustMelody(0).MustCheck(voiceleading.OMTLeapRecovery).Violations)
}
