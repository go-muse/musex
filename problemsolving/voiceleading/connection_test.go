package voiceleading_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: business-spec SA-1, SA-2, SA-5 and architecture 9.

type wantVoiceMove struct {
	iv        string
	chromatic int
}

func assertConnection(t *testing.T, from, to *voiceleading.Sonority, moves []wantVoiceMove, common []string, total int, kind voiceleading.ConnectionKind, got voiceleading.Connection) {
	t.Helper()
	assert.Same(t, from, got.From)
	assert.Same(t, to, got.To)
	require.Len(t, got.Moves, len(moves))
	fromNotes, toNotes := from.MustNotes(), to.MustNotes()
	for k, w := range moves {
		mv := got.Moves[k]
		assert.Equal(t, k, mv.Voice)
		assert.True(t, fromNotes[k].Equal(mv.From), "voice %d from", k)
		assert.True(t, toNotes[k].Equal(mv.To), "voice %d to", k)
		assertIntervalText(t, w.iv, mv.Interval, "voice", k)
		assert.Equal(t, system.Step(w.chromatic), mv.Chromatic, "voice %d chromatic", k)
	}
	require.Len(t, got.Common, len(common))
	for i, c := range common {
		assert.True(t, spellingOf(t, c).Equal(got.Common[i]), "common %d: want %s, got %s", i, c, got.Common[i])
	}
	assert.Equal(t, system.Step(total), got.Total)
	assert.Equal(t, kind, got.Kind)
}

func TestConnect(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		from, to []string
		moves    []wantVoiceMove
		common   []string
		total    int
		kind     voiceleading.ConnectionKind
	}{
		{"SA-1 C3 E4 G4 C5 to F3 F4 A4 C5", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "F4", "A4", "C5"},
			[]wantVoiceMove{{"P4", 5}, {"m2", 1}, {"M2", 2}, {"P1", 0}}, []string{"C"}, 8, voiceleading.HarmonicConnection},
		{"no common class is melodic", []string{"C3", "E4", "G4", "C5"}, []string{"D3", "F4", "A4", "D5"},
			[]wantVoiceMove{{"M2", 2}, {"m2", 1}, {"M2", 2}, {"M2", 2}}, []string{}, 7, voiceleading.MelodicConnection},
		{"a common class that no voice keeps is melodic", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "C4", "F4", "A4"},
			[]wantVoiceMove{{"P4", 5}, {"-M3", -4}, {"-M2", -2}, {"-m3", -3}}, []string{"C"}, 14, voiceleading.MelodicConnection},
		{"common classes follow the voices of the first sonority", []string{"C3", "E4", "G4", "C5"}, []string{"A2", "E4", "A4", "C5"},
			[]wantVoiceMove{{"-m3", -3}, {"P1", 0}, {"M2", 2}, {"P1", 0}}, []string{"C", "E"}, 5, voiceleading.HarmonicConnection},
		{"an enharmonic change is not a kept common tone", []string{"C4", "G#4"}, []string{"C4", "Ab4"},
			[]wantVoiceMove{{"P1", 0}, {"d2", 0}}, []string{"C"}, 0, voiceleading.HarmonicConnection},
		{"one voice", []string{"C4"}, []string{"D4"}, []wantVoiceMove{{"M2", 2}}, []string{}, 2, voiceleading.MelodicConnection},
		{"crossing voices", []string{"C4", "E4"}, []string{"G4", "D4"}, []wantVoiceMove{{"P5", 7}, {"-M2", -2}}, []string{}, 9, voiceleading.MelodicConnection},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from, to := sonorityOf(t, tt.from...), sonorityOf(t, tt.to...)
			got, err := voiceleading.Connect(from, to)
			require.NoError(t, err)
			assertConnection(t, from, to, tt.moves, tt.common, tt.total, tt.kind, got)
			assert.Equal(t, got, voiceleading.MustConnect(from, to))
			for k, mv := range got.Moves {
				assert.Equal(t, mv.Interval.IsPerfectUnison(), mv.IsRepeat(), "voice %d", k)
			}
		})
	}
}

func TestConnectFailureOrder(t *testing.T) {
	t.Parallel()
	four := sonorityOf(t, "C3", "E4", "G4", "C5")
	three := sonorityOf(t, "C3", "E4", "G4")
	other := voiceleading.MustNewSonority(chordOf(t, "C3", "E4", "G4", "C5"), system.Positions19())
	unordered := silentSonority(t, voiceleading.Unordered)
	silent := silentSonority(t, voiceleading.LowToHigh)
	empty := voiceleading.MustNewSonority(chord.MustNew(), nil)
	cases := []struct {
		name     string
		from, to *voiceleading.Sonority
		want     error
	}{
		{"an unset first sonority", nil, four, voiceleading.ErrUnset},
		{"an unset second sonority", four, new(voiceleading.Sonority), voiceleading.ErrUnset},
		{"an unset sonority precedes a missing order", unordered, nil, voiceleading.ErrUnset},
		{"a missing order of the first", unordered, four, voiceleading.ErrNoOrder},
		{"a missing order of the second", four, unordered, voiceleading.ErrNoOrder},
		{"a missing order precedes a silent part", silent, unordered, voiceleading.ErrNoOrder},
		{"a silent part of the first", silent, three, voiceleading.ErrSilentPart},
		{"a silent part of the second", three, silent, voiceleading.ErrSilentPart},
		{"a silent part precedes the voice count", silent, four, voiceleading.ErrSilentPart},
		{"different voice counts", three, four, voiceleading.ErrVoiceCount},
		{"no voices", empty, empty, voiceleading.ErrVoiceCount},
		{"the voice count precedes the systems", three, other, voiceleading.ErrVoiceCount},
		{"different systems", four, other, voiceleading.ErrDifferentSystems},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := voiceleading.Connect(tt.from, tt.to)
			assert.Equal(t, voiceleading.Connection{}, got)
			assertSentinel(t, err, tt.want)
			assertNoChordError(t, err)
			assertNoPartError(t, err)
			assertMustPanics(t, err, func() { voiceleading.MustConnect(tt.from, tt.to) })
			a, err := voiceleading.CheckMT21Connection(tt.from, tt.to)
			assert.Nil(t, a)
			assertSentinel(t, err, tt.want)
			assertMustPanics(t, err, func() { voiceleading.MustCheckMT21Connection(tt.from, tt.to) })
		})
	}
}

func TestConnectMeasurementFailures(t *testing.T) {
	t.Parallel()
	t.Run("a move out of range", func(t *testing.T) {
		t.Parallel()
		from := voiceleading.MustNewSonority(chord.MustNew(aspnNote(t, "C", math.MinInt)), traditionalSystem())
		to := voiceleading.MustNewSonority(chord.MustNew(aspnNote(t, "C", math.MaxInt)), traditionalSystem())
		got, err := voiceleading.Connect(from, to)
		assert.Equal(t, voiceleading.Connection{}, got)
		assert.ErrorIs(t, err, interval.ErrOutOfRange)
		assertSentinel(t, err, nil)
	})
	t.Run("a total beyond int", func(t *testing.T) {
		t.Parallel()
		from := voiceleading.MustNewSonority(chord.MustNew(aspnNote(t, "C", 0), aspnNote(t, "C", 1)), traditionalSystem())
		to := voiceleading.MustNewSonority(chord.MustNew(aspnNote(t, "C", math.MaxInt/16), aspnNote(t, "C", math.MaxInt/16+1)), traditionalSystem())
		got, err := voiceleading.Connect(from, to)
		assert.Equal(t, voiceleading.Connection{}, got)
		assert.ErrorIs(t, err, system.ErrOutOfRange)
		assertSentinel(t, err, nil)
		assertNoChordError(t, err)
		_, err = voiceleading.CheckMT21Connection(from, to)
		assert.ErrorIs(t, err, system.ErrOutOfRange)
	})
}

const (
	mt21Source = "Music Theory for the 21st-Century Classroom, 26.7.1, 26.7.2"
	mt21Scope  = "four sounding voices; root-position triads with a doubled bass, as asserted by the composer"
)

func TestCheckMT21Connection(t *testing.T) {
	t.Parallel()
	v1, v2 := voiceleading.MT21CommonToneAndSteps, voiceleading.MT21SameDirection
	two := voiceleading.MT21TwoCommonTones
	fourthOrFifth := func(first, second bool) []voiceleading.VariantCheck {
		return []voiceleading.VariantCheck{{Variant: v1, Satisfied: first}, {Variant: v2, Satisfied: second}}
	}
	// A bass by a third or a sixth has the single variant of 26.7.1.
	thirdOrSixth := func(satisfied bool) []voiceleading.VariantCheck {
		return []voiceleading.VariantCheck{{Variant: two, Satisfied: satisfied}}
	}
	cases := []struct {
		name     string
		from, to []string
		variants []voiceleading.VariantCheck
		conforms bool
	}{
		{"SA-2 C3 E4 G4 C5 to F3 F4 A4 C5", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "F4", "A4", "C5"}, fourthOrFifth(true, false), true},
		{"SA-2 C3 E4 G4 C5 to F3 C4 F4 A4", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "C4", "F4", "A4"}, fourthOrFifth(false, true), true},
		{"SA-2 C3 E4 G4 C5 to A2 E4 A4 C5", []string{"C3", "E4", "G4", "C5"}, []string{"A2", "E4", "A4", "C5"}, thirdOrSixth(true), true},
		{"a bass up a fifth, one kept tone and two steps", []string{"C3", "E4", "G4", "C5"}, []string{"G3", "D4", "G4", "B4"}, fourthOrFifth(true, false), true},
		{"a bass up a fifth with leaps in two directions", []string{"C3", "E4", "G4", "C5"}, []string{"G3", "B3", "D4", "G5"}, fourthOrFifth(false, false), false},
		{"a bass down a fourth, all upper voices up", []string{"C3", "E4", "G4", "C5"}, []string{"G2", "G4", "B4", "D5"}, fourthOrFifth(false, true), true},
		{"a bass up a fourth, one kept tone and a leap", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "A3", "F4", "C5"}, fourthOrFifth(false, false), false},
		{"a bass up a sixth keeping two tones", []string{"C3", "E4", "G4", "C5"}, []string{"A3", "E4", "A4", "C5"}, thirdOrSixth(true), true},
		{"a bass up a third keeping two tones", []string{"C3", "E4", "G4", "C5"}, []string{"E3", "E4", "G4", "B4"}, thirdOrSixth(true), true},
		{"a bass down a third keeping no tone", []string{"C3", "E4", "G4", "C5"}, []string{"A2", "C4", "E4", "A4"}, thirdOrSixth(false), false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from, to := sonorityOf(t, tt.from...), sonorityOf(t, tt.to...)
			a, err := voiceleading.CheckMT21Connection(from, to)
			require.NoError(t, err)
			require.NotNil(t, a)
			assert.Equal(t, voiceleading.MT21ConnectionRule, a.Rule)
			assert.Equal(t, "mt21-root-position-triads", a.Rule)
			assert.Equal(t, mt21Source, a.Source)
			assert.Equal(t, mt21Scope, a.Scope)
			assert.True(t, a.Applicable)
			assert.Empty(t, a.Reason)
			assert.Equal(t, tt.variants, a.Variants)
			assert.Equal(t, tt.conforms, a.Conforms)
			assert.Equal(t, voiceleading.MustConnect(from, to), a.Connection)
			assert.Equal(t, a, voiceleading.MustCheckMT21Connection(from, to))
			assert.NotSame(t, a, voiceleading.MustCheckMT21Connection(from, to))
		})
	}
}

func TestCheckMT21ConnectionScope(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		from, to []string
		reason   string
	}{
		{"two voices", []string{"C3", "C4"}, []string{"F3", "F4"}, voiceleading.ReasonVoiceCount},
		{"five voices", []string{"C3", "G3", "E4", "G4", "C5"}, []string{"F3", "A3", "F4", "A4", "C5"}, voiceleading.ReasonVoiceCount},
		{"four classes in the first chord", []string{"C3", "E4", "G4", "B4"}, []string{"F3", "F4", "A4", "C5"}, voiceleading.ReasonFromNotThreeClasses},
		{"the first chord precedes the second", []string{"C3", "E4", "G4", "B4"}, []string{"F3", "F4", "F5", "C5"}, voiceleading.ReasonFromNotThreeClasses},
		{"four classes in the second chord", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "A3", "C4", "E4"}, voiceleading.ReasonToNotThreeClasses},
		{"the classes of both chords precede the doubling of the bass", []string{"C3", "E4", "G4", "G5"}, []string{"F3", "F4", "F5", "C5"}, voiceleading.ReasonToNotThreeClasses},
		{"an undoubled bass in the first chord", []string{"C3", "E4", "G4", "G5"}, []string{"F3", "F4", "A4", "C5"}, voiceleading.ReasonFromBassNotDoubled},
		{"the first bass precedes the second", []string{"C3", "E4", "G4", "G5"}, []string{"F3", "A3", "C4", "A4"}, voiceleading.ReasonFromBassNotDoubled},
		{"an undoubled bass in the second chord", []string{"C3", "E4", "G4", "C5"}, []string{"F3", "A3", "C4", "A4"}, voiceleading.ReasonToBassNotDoubled},
		{"the doubling precedes the motion of the bass", []string{"C3", "E4", "G4", "C5"}, []string{"D3", "F4", "A4", "A5"}, voiceleading.ReasonToBassNotDoubled},
		{"SA-2 a bass by a second", []string{"C3", "E4", "G4", "C5"}, []string{"D3", "F4", "A4", "D5"}, voiceleading.ReasonBassMotion},
		{"a bass by an octave", []string{"C3", "E4", "G4", "C5"}, []string{"C4", "E4", "G4", "C5"}, voiceleading.ReasonBassMotion},
		{"a bass that stays", []string{"C3", "E4", "G4", "C5"}, []string{"C3", "G4", "C5", "E5"}, voiceleading.ReasonBassMotion},
		{"a bass by a seventh", []string{"C3", "E4", "G4", "C5"}, []string{"B3", "D4", "F#4", "B4"}, voiceleading.ReasonBassMotion},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from, to := sonorityOf(t, tt.from...), sonorityOf(t, tt.to...)
			a, err := voiceleading.CheckMT21Connection(from, to)
			require.NoError(t, err)
			require.NotNil(t, a)
			assert.False(t, a.Applicable)
			assert.Equal(t, tt.reason, a.Reason)
			assert.Empty(t, a.Variants)
			assert.False(t, a.Conforms)
			assert.Equal(t, voiceleading.MT21ConnectionRule, a.Rule)
			assert.Equal(t, mt21Source, a.Source)
			assert.Equal(t, mt21Scope, a.Scope)
			assert.Equal(t, voiceleading.MustConnect(from, to), a.Connection)
		})
	}
}
