package voiceleading_test

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expectations: business-spec SA-3, SA-3a, SA-4, SA-5 and architecture 10.
// The reference below enumerates every combination of candidates as the
// contract defines it, independently of ChooseConnections.

type wantChoice struct {
	voices []string
	total  int
}

func assertChoices(t *testing.T, from *voiceleading.Sonority, want []wantChoice, got []voiceleading.Connection) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for i, w := range want {
		assert.Same(t, from, got[i].From, "choice %d", i)
		require.NotNil(t, got[i].To, "choice %d", i)
		assertNotesAre(t, w.voices, got[i].To.MustNotes(), "choice", i)
		assert.Equal(t, system.Step(w.total), got[i].Total, "choice %d total", i)
	}
}

func choicePrefix(t *testing.T, got []voiceleading.Connection, n int) []voiceleading.Connection {
	t.Helper()
	require.GreaterOrEqual(t, len(got), n)
	return got[:n]
}

func TestChooseConnectionsExamples(t *testing.T) {
	t.Parallel()
	t.Run("SA-3 C3 E4 G4 C5 to F A C over F3", func(t *testing.T) {
		t.Parallel()
		from := sonorityOf(t, "C3", "E4", "G4", "C5")
		conds := voiceleading.ConnectionConditions{
			Complete: true, CommonTones: true, NoCrossings: true,
			Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth},
		}
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "F", "A", "C"), noteOf(t, "F3", nil), conds)
		require.NoError(t, err)
		assertChoices(t, from, []wantChoice{
			{[]string{"F3", "F4", "A4", "C5"}, 8},
			{[]string{"F3", "C4", "A4", "C5"}, 11},
			{[]string{"F3", "A4", "A4", "C5"}, 12},
			{[]string{"F3", "A3", "F4", "C5"}, 14},
			{[]string{"F3", "A3", "A4", "C5"}, 14},
			{[]string{"F3", "A4", "C5", "C5"}, 15},
			{[]string{"F3", "A3", "C5", "C5"}, 17},
			{[]string{"F3", "F3", "A4", "C5"}, 18},
			{[]string{"F3", "A3", "C4", "C5"}, 19},
			{[]string{"F3", "A3", "A3", "C5"}, 22},
			{[]string{"F3", "F3", "A3", "C5"}, 26},
		}, got)
		first := got[0]
		require.Len(t, first.Moves, 4)
		assertIntervalText(t, "P4", first.Moves[0].Interval)
		assertIntervalText(t, "m2", first.Moves[1].Interval)
		assertIntervalText(t, "M2", first.Moves[2].Interval)
		assertIntervalText(t, "P1", first.Moves[3].Interval)
		assert.Equal(t, voiceleading.HarmonicConnection, first.Kind)
		require.Len(t, first.Common, 1)
		assert.True(t, spelling.C().Equal(first.Common[0]))
		// With the MT21 norm only the first variant remains.
		conds.MT21 = true
		got, err = voiceleading.ChooseConnections(from, namesChord(t, "F", "A", "C"), noteOf(t, "F3", nil), conds)
		require.NoError(t, err)
		assertChoices(t, from, []wantChoice{{[]string{"F3", "F4", "A4", "C5"}, 8}}, got)
		assert.Equal(t, got, voiceleading.MustChooseConnections(from, namesChord(t, "F", "A", "C"), noteOf(t, "F3", nil), conds))
	})
	t.Run("SA-3 C3 G3 E4 C5 to D F A over D3", func(t *testing.T) {
		t.Parallel()
		from := sonorityOf(t, "C3", "G3", "E4", "C5")
		conds := voiceleading.ConnectionConditions{Complete: true, CommonTones: true, NoCrossings: true}
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "D", "F", "A"), noteOf(t, "D3", nil), conds)
		require.NoError(t, err)
		assertChoices(t, from, []wantChoice{
			{[]string{"D3", "A3", "F4", "D5"}, 7},
			{[]string{"D3", "F3", "F4", "A4"}, 8},
			{[]string{"D3", "A3", "F4", "A4"}, 8},
		}, choicePrefix(t, got, 3))
		conds.Parallels = voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}
		got, err = voiceleading.ChooseConnections(from, namesChord(t, "D", "F", "A"), noteOf(t, "D3", nil), conds)
		require.NoError(t, err)
		assertChoices(t, from, []wantChoice{{[]string{"D3", "F3", "F4", "A4"}, 8}}, choicePrefix(t, got, 1))
		for _, c := range got {
			tenor, err := c.To.MustNotes()[1].Pitch()
			require.NoError(t, err)
			assert.False(t, writtenPitch(t, "A3").Equal(tenor), "parallel fifths C3-G3 to D3-A3 must be excluded")
		}
	})
	t.Run("equal positions are ordered by the classes of the composition", func(t *testing.T) {
		t.Parallel()
		from := sonorityOf(t, "C4", "C5")
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "F", "E#"), noteOf(t, "F3", nil), voiceleading.ConnectionConditions{})
		require.NoError(t, err)
		assertChoices(t, from, []wantChoice{
			{[]string{"F3", "F5"}, 12},
			{[]string{"F3", "E#5"}, 12},
			{[]string{"F3", "F4"}, 14},
			{[]string{"F3", "E#4"}, 14},
		}, got)
	})
	t.Run("no variant is an empty result", func(t *testing.T) {
		t.Parallel()
		from := sonorityOf(t, "C4", "E4")
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "G"), noteOf(t, "G5", nil), voiceleading.ConnectionConditions{NoCrossings: true})
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
}

func TestChooseConnectionsEachCondition(t *testing.T) {
	t.Parallel()
	from := sonorityOf(t, "C4", "E4", "G4")
	target := func(t *testing.T) chord.Chord { return namesChord(t, "C", "E", "G") }
	bass := func(t *testing.T) note.Note { return noteOf(t, "C4", nil) }
	cases := []struct {
		name  string
		conds voiceleading.ConnectionConditions
		count int
		check func(t *testing.T, got []voiceleading.Connection)
	}{
		{"no condition", voiceleading.ConnectionConditions{}, 36, nil},
		{"1 completeness", voiceleading.ConnectionConditions{Complete: true}, 8, func(t *testing.T, got []voiceleading.Connection) {
			assertNotesAre(t, []string{"C4", "E4", "G4"}, got[0].To.MustNotes())
		}},
		{"2 common tones", voiceleading.ConnectionConditions{CommonTones: true}, 1, func(t *testing.T, got []voiceleading.Connection) {
			assertChoices(t, from, []wantChoice{{[]string{"C4", "E4", "G4"}, 0}}, got)
		}},
		{"3 no crossings", voiceleading.ConnectionConditions{NoCrossings: true}, 14, nil},
		{"4 no overlaps", voiceleading.ConnectionConditions{NoOverlaps: true}, 12, nil},
		{"5 distances", voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{interval.PerfectFifth()}}}, -1, nil},
		{"6 parallels", voiceleading.ConnectionConditions{Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}}, 36, nil},
		{"7 doubled bass", voiceleading.ConnectionConditions{BassDoubled: true}, -1, func(t *testing.T, got []voiceleading.Connection) {
			for _, c := range got {
				names := c.To.MustNotes()
				assert.True(t, names[1].Spelling().Equal(spelling.C()) || names[2].Spelling().Equal(spelling.C()))
			}
		}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := voiceleading.ChooseConnections(from, target(t), bass(t), tt.conds)
			require.NoError(t, err)
			if tt.count >= 0 {
				assert.Len(t, got, tt.count)
			}
			assertMatchesReference(t, from, target(t), bass(t), tt.conds, got)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestChooseConnectionsParallelOptions(t *testing.T) {
	t.Parallel()
	octaves := voiceleading.ParallelOptions{Perfect: voiceleading.PerfectOctaves}
	cases := []struct {
		name     string
		from     []string
		target   []string
		bass     string
		parallel voiceleading.ParallelOptions
		want     []wantChoice
	}{
		{"no condition", []string{"C3", "C4"}, []string{"D", "A"}, "D3", voiceleading.ParallelOptions{},
			[]wantChoice{{[]string{"D3", "D4"}, 4}, {[]string{"D3", "A3"}, 5}, {[]string{"D3", "A4"}, 11}, {[]string{"D3", "D3"}, 12}}},
		{"parallel octaves are excluded", []string{"C3", "C4"}, []string{"D", "A"}, "D3", octaves,
			[]wantChoice{{[]string{"D3", "A3"}, 5}, {[]string{"D3", "A4"}, 11}, {[]string{"D3", "D3"}, 12}}},
		{"fifths alone keep the octaves", []string{"C3", "C4"}, []string{"D", "A"}, "D3", voiceleading.ParallelOptions{Perfect: voiceleading.PerfectFifths},
			[]wantChoice{{[]string{"D3", "D4"}, 4}, {[]string{"D3", "A3"}, 5}, {[]string{"D3", "A4"}, 11}, {[]string{"D3", "D3"}, 12}}},
		{"a double octave without compounds", []string{"C3", "C5"}, []string{"D"}, "D3", octaves,
			[]wantChoice{{[]string{"D3", "D5"}, 4}, {[]string{"D3", "D4"}, 12}}},
		{"a double octave with compounds", []string{"C3", "C5"}, []string{"D"}, "D3", voiceleading.ParallelOptions{Perfect: voiceleading.PerfectOctaves, Compound: true},
			[]wantChoice{{[]string{"D3", "D4"}, 12}}},
		{"unisons without the flag", []string{"C4", "C4"}, []string{"D"}, "D4", octaves,
			[]wantChoice{{[]string{"D4", "D4"}, 4}, {[]string{"D4", "D3"}, 12}}},
		{"unisons with the flag", []string{"C4", "C4"}, []string{"D"}, "D4", voiceleading.ParallelOptions{Perfect: voiceleading.PerfectOctaves, Unisons: true},
			[]wantChoice{{[]string{"D4", "D3"}, 12}}},
		{"unisons with fifths only add nothing", []string{"C4", "C4"}, []string{"D"}, "D4", voiceleading.ParallelOptions{Perfect: voiceleading.PerfectFifths, Unisons: true},
			[]wantChoice{{[]string{"D4", "D4"}, 4}, {[]string{"D4", "D3"}, 12}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from := sonorityOf(t, tt.from...)
			conds := voiceleading.ConnectionConditions{Parallels: tt.parallel}
			got, err := voiceleading.ChooseConnections(from, namesChord(t, tt.target...), noteOf(t, tt.bass, nil), conds)
			require.NoError(t, err)
			assertChoices(t, from, tt.want, got)
			assertMatchesReference(t, from, namesChord(t, tt.target...), noteOf(t, tt.bass, nil), conds, got)
		})
	}
}

func TestChooseConnectionsMatchesTheFullEnumeration(t *testing.T) {
	t.Parallel()
	omt := voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}
	cases := []struct {
		name   string
		from   []string
		target []string
		bass   string
		conds  voiceleading.ConnectionConditions
	}{
		{"SA-3 first example, MT21 only", []string{"C3", "E4", "G4", "C5"}, []string{"F", "A", "C"}, "F3", voiceleading.ConnectionConditions{MT21: true}},
		{"SA-3 first example, all conditions but MT21", []string{"C3", "E4", "G4", "C5"}, []string{"F", "A", "C"}, "F3", voiceleading.ConnectionConditions{
			Complete: true, CommonTones: true, NoCrossings: true, NoOverlaps: true, Distances: omt,
			Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}, BassDoubled: true,
		}},
		{"SA-3 first example, every condition", []string{"C3", "E4", "G4", "C5"}, []string{"F", "A", "C"}, "F3", voiceleading.ConnectionConditions{
			Complete: true, CommonTones: true, NoCrossings: true, NoOverlaps: true, Distances: omt,
			Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}, BassDoubled: true, MT21: true,
		}},
		{"SA-3 second example, conditions 1-3", []string{"C3", "G3", "E4", "C5"}, []string{"D", "F", "A"}, "D3", voiceleading.ConnectionConditions{Complete: true, CommonTones: true, NoCrossings: true}},
		{"SA-3 second example, conditions 1-3 and 6", []string{"C3", "G3", "E4", "C5"}, []string{"D", "F", "A"}, "D3", voiceleading.ConnectionConditions{
			Complete: true, CommonTones: true, NoCrossings: true, Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth},
		}},
		{"four classes without conditions", []string{"C3", "E4", "G4", "C5"}, []string{"D", "F", "A", "C"}, "D3", voiceleading.ConnectionConditions{}},
		{"a bass down a third under MT21 and the MT21 distances", []string{"C3", "E4", "G4", "C5"}, []string{"A", "C", "E"}, "A2", voiceleading.ConnectionConditions{
			MT21: true, Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB},
		}},
		{"five voices with overlaps, distances and compound parallels", []string{"C3", "G3", "C4", "E4", "G4"}, []string{"F", "A", "C"}, "F3", voiceleading.ConnectionConditions{
			NoOverlaps: true, Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{interval.PerfectOctave()}},
			Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth, Compound: true, Unisons: true},
		}},
		{"six voices and four classes", []string{"G2", "D3", "G3", "B3", "D4", "G4"}, []string{"C", "E", "G", "Bb"}, "C3", voiceleading.ConnectionConditions{
			Complete: true, NoCrossings: true, BassDoubled: true,
		}},
		{"six voices and four classes with every pairwise condition", []string{"G2", "D3", "G3", "B3", "D4", "G4"}, []string{"C", "E", "G", "Bb"}, "C3", voiceleading.ConnectionConditions{
			CommonTones: true, NoCrossings: true, NoOverlaps: true,
			Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{interval.PerfectTwelfth(), interval.PerfectOctave(), interval.PerfectOctave(), interval.PerfectOctave(), interval.PerfectOctave()}},
			Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth},
		}},
		{"six voices and two classes without conditions", []string{"C3", "E3", "G3", "C4", "E4", "G4"}, []string{"D", "F#"}, "D3", voiceleading.ConnectionConditions{}},
		{"crossed voices", []string{"C3", "G4", "E4", "C5"}, []string{"F", "A", "C"}, "F3", voiceleading.ConnectionConditions{NoOverlaps: true, Complete: true}},
		{"enharmonic classes", []string{"C4", "E4", "G#4"}, []string{"A", "C#", "E", "Db"}, "A3", voiceleading.ConnectionConditions{NoCrossings: true}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from := sonorityOf(t, tt.from...)
			target := namesChord(t, tt.target...)
			bass := noteOf(t, tt.bass, nil)
			got, err := voiceleading.ChooseConnections(from, target, bass, tt.conds)
			require.NoError(t, err)
			assertMatchesReference(t, from, target, bass, tt.conds, got)
		})
	}
}

func TestChooseConnectionsOwnership(t *testing.T) {
	t.Parallel()
	src := &testSource{current: settings.Standard()}
	tunedF, err := note.FromSpelling(spelling.F(), cents.MustParse("7"), src)
	require.NoError(t, err)
	tunedA, err := note.FromPitch(writtenPitch(t, "A5"), cents.MustParse("-3"), src)
	require.NoError(t, err)
	bass, err := note.FromPitch(writtenPitch(t, "F3"), cents.MustParse("11"), src)
	require.NoError(t, err)
	src.reset(nil)
	from := sonorityOf(t, "C3", "E4", "G4", "C5")
	target := chord.MustNew(tunedF, tunedA, nameOf(t, "C"), nameOf(t, "F"))
	got, err := voiceleading.ChooseConnections(from, target, bass, voiceleading.ConnectionConditions{Complete: true, CommonTones: true, NoCrossings: true})
	require.NoError(t, err)
	assert.Zero(t, src.count())
	require.NotEmpty(t, got)
	for i, c := range got {
		to := c.To
		assert.Equal(t, voiceleading.LowToHigh, to.MustOrder())
		parts, ok := to.MustParts()
		assert.Nil(t, parts)
		assert.False(t, ok)
		sys, ok := to.MustSystem()
		assert.True(t, ok)
		assert.True(t, traditionalSystem().Equal(sys))
		notes := to.MustNotes()
		assert.True(t, notes[0].Equal(bass), "choice %d: the bass is the given note", i)
		assert.True(t, notes[0].SameSource(bass), "choice %d: the bass keeps its source", i)
		assert.True(t, notes[0].SameCents(bass))
		require.Len(t, c.Moves, 4)
		assert.True(t, c.Moves[0].To.Equal(bass))
		for k := 1; k < 4; k++ {
			n := notes[k]
			assert.False(t, n.IsBound(), "choice %d voice %d is unbound", i, k)
			switch {
			case n.Spelling().Equal(spelling.F()):
				assert.True(t, n.SameCents(tunedF), "choice %d voice %d keeps the correction of F", i, k)
			case n.Spelling().Equal(spelling.A()):
				assert.True(t, n.SameCents(tunedA), "choice %d voice %d keeps the correction of A", i, k)
			default:
				assert.True(t, n.Spelling().Equal(spelling.C()))
			}
			assert.True(t, c.Moves[k].To.Equal(n))
		}
	}
	// Results are independent: changing one does not change the next call.
	again := voiceleading.MustChooseConnections(from, target, bass, voiceleading.ConnectionConditions{Complete: true, CommonTones: true, NoCrossings: true})
	got[0].Moves[0].Chromatic = 99
	got[0].Common = nil
	got[0].To = nil
	fresh := voiceleading.MustChooseConnections(from, target, bass, voiceleading.ConnectionConditions{Complete: true, CommonTones: true, NoCrossings: true})
	assert.Equal(t, again, fresh)
	assert.Zero(t, src.count())
}

func TestChooseConnectionsFailureOrder(t *testing.T) {
	t.Parallel()
	four := sonorityOf(t, "C3", "E4", "G4", "C5")
	three := sonorityOf(t, "C3", "E4", "G4")
	one := sonorityOf(t, "C4")
	empty := voiceleading.MustNewSonority(chord.MustNew(), nil)
	unordered := silentSonority(t, voiceleading.Unordered)
	silent := silentSonority(t, voiceleading.LowToHigh)
	fac := namesChord(t, "F", "A", "C")
	f3 := noteOf(t, "F3", nil)
	halfName, err := note.FromSpelling(spelling.MustParse("C+1/2"), cents.Zero(), nil)
	require.NoError(t, err)
	halfBass, err := note.FromPitch(pitch.MustFromASPN(spelling.MustParse("C+1/2"), 3), cents.Zero(), nil)
	require.NoError(t, err)
	badConds := voiceleading.ConnectionConditions{Parallels: voiceleading.ParallelOptions{Perfect: 4}}
	p8 := interval.PerfectOctave()
	type chooseFailure struct {
		sentinel error
		target   error
		chord    bool
		note     int
	}
	cases := []struct {
		name   string
		from   *voiceleading.Sonority
		target chord.Chord
		bass   note.Note
		conds  voiceleading.ConnectionConditions
		want   chooseFailure
	}{
		// Stage 1.
		{"an unset sonority precedes everything", nil, chord.Chord{}, note.Note{}, badConds, chooseFailure{voiceleading.ErrUnset, voiceleading.ErrUnset, false, 0}},
		{"an unset zero sonority", new(voiceleading.Sonority), fac, f3, voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrUnset, voiceleading.ErrUnset, false, 0}},
		{"an unset composition precedes an unset bass", four, chord.Chord{}, note.Note{}, voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrUnset, chord.ErrUnset, true, -1}},
		{"an unset bass", four, fac, note.Note{}, voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrUnset, note.ErrUnset, false, 0}},
		{"an unset bass precedes a missing order", unordered, fac, note.Note{}, voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrUnset, note.ErrUnset, false, 0}},
		// Stage 2.
		{"a missing order", unordered, chord.MustNew(), f3, badConds, chooseFailure{voiceleading.ErrNoOrder, voiceleading.ErrNoOrder, false, 0}},
		{"a silent part", silent, chord.MustNew(), f3, badConds, chooseFailure{voiceleading.ErrSilentPart, voiceleading.ErrSilentPart, false, 0}},
		{"one voice", one, chord.MustNew(), f3, badConds, chooseFailure{voiceleading.ErrVoiceCount, voiceleading.ErrVoiceCount, false, 0}},
		{"no voices", empty, fac, f3, voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrVoiceCount, voiceleading.ErrVoiceCount, false, 0}},
		// Stage 3.
		{"an empty composition precedes a bass without an octave", four, chord.MustNew(), nameOf(t, "F"), badConds, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, true, -1}},
		{"a bass without an octave precedes the conditions", four, fac, nameOf(t, "F"), badConds, chooseFailure{voiceleading.ErrNoOctave, voiceleading.ErrNoOctave, false, 0}},
		// Stage 4.
		{"an unknown distance rule", four, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceRule(99)}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"a custom distance norm without limits", four, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"limits of a built-in distance norm", four, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB, Limits: []interval.Interval{p8}}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"limits without a distance rule", four, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Limits: []interval.Interval{p8}}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"an unknown set of perfect intervals", four, fac, f3, badConds, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"compounds without perfect intervals", four, fac, f3, voiceleading.ConnectionConditions{Parallels: voiceleading.ParallelOptions{Compound: true}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"unisons without perfect intervals", four, fac, f3, voiceleading.ConnectionConditions{Parallels: voiceleading.ParallelOptions{Unisons: true}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"the conditions precede an inadmissible composition", four, chord.MustNew(nameOf(t, "F"), halfName), f3, badConds, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		// Stage 5.
		{"an inadmissible name of the composition", four, chord.MustNew(nameOf(t, "F"), nameOf(t, "A"), halfName), f3, voiceleading.ConnectionConditions{}, chooseFailure{nil, system.ErrInadmissible, true, 2}},
		{"an inadmissible composition precedes a bass outside it", four, chord.MustNew(nameOf(t, "A"), halfName), f3, voiceleading.ConnectionConditions{}, chooseFailure{nil, system.ErrInadmissible, true, 1}},
		{"an inadmissible bass", four, chord.MustNew(nameOf(t, "F"), halfName), halfBass, voiceleading.ConnectionConditions{}, chooseFailure{nil, system.ErrInadmissible, true, 1}},
		{"an inadmissible bass alone", four, fac, halfBass, voiceleading.ConnectionConditions{}, chooseFailure{nil, system.ErrInadmissible, false, 0}},
		// Stage 6.
		{"a bass outside the composition", four, fac, noteOf(t, "G3", nil), voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"a bass outside the composition by spelling", four, namesChord(t, "F", "Ab", "C"), noteOf(t, "G#3", nil), voiceleading.ConnectionConditions{}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		// Stage 8.
		{"a built-in distance norm for another count of voices", three, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTSATB}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"first species for four voices", four, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceOMTFirstSpecies}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"two custom limits for three pairs", four, fac, f3, voiceleading.ConnectionConditions{Distances: voiceleading.DistanceNorm{Rule: voiceleading.DistanceCustom, Limits: []interval.Interval{p8, p8}}}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"MT21 for three voices", three, fac, f3, voiceleading.ConnectionConditions{MT21: true}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"MT21 from four classes", sonorityOf(t, "C3", "E4", "G4", "B4"), fac, f3, voiceleading.ConnectionConditions{MT21: true}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"MT21 from an undoubled bass", sonorityOf(t, "C3", "E4", "G4", "G5"), fac, f3, voiceleading.ConnectionConditions{MT21: true}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"MT21 to four classes", four, namesChord(t, "F", "A", "C", "Eb"), f3, voiceleading.ConnectionConditions{MT21: true}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"MT21 with a bass by a second", four, namesChord(t, "D", "F", "A"), noteOf(t, "D3", nil), voiceleading.ConnectionConditions{MT21: true}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
		{"MT21 with a bass by an octave", four, namesChord(t, "C", "E", "G"), noteOf(t, "C4", nil), voiceleading.ConnectionConditions{MT21: true}, chooseFailure{voiceleading.ErrOption, voiceleading.ErrOption, false, 0}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := voiceleading.ChooseConnections(tt.from, tt.target, tt.bass, tt.conds)
			assert.Nil(t, got)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.want.target)
			assertSentinel(t, err, tt.want.sentinel)
			if tt.want.chord {
				assertChordError(t, err, tt.want.target, 0, tt.want.note)
			} else {
				assertNoChordError(t, err)
			}
			assertNoPartError(t, err)
			assertMustPanics(t, err, func() { voiceleading.MustChooseConnections(tt.from, tt.target, tt.bass, tt.conds) })
		})
	}
}

func TestChooseConnectionsMeasurementFailures(t *testing.T) {
	t.Parallel()
	sys := traditionalSystem()
	sonority := func(t *testing.T, notes ...note.Note) *voiceleading.Sonority {
		t.Helper()
		s, err := voiceleading.NewSonority(chord.MustNew(notes...), sys)
		require.NoError(t, err)
		return s
	}
	const m16 = math.MaxInt / 16
	t.Run("stage 7: the move of the bass precedes the scope of MT21", func(t *testing.T) {
		t.Parallel()
		from := sonority(t, aspnNote(t, "C", math.MinInt), aspnNote(t, "E", math.MinInt))
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "C", "E"), aspnNote(t, "C", math.MaxInt), voiceleading.ConnectionConditions{MT21: true})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, interval.ErrOutOfRange)
		assertSentinel(t, err, nil)
		assertNoChordError(t, err)
	})
	t.Run("stage 9.1: a candidate beyond the range of octaves", func(t *testing.T) {
		t.Parallel()
		from := sonority(t, aspnNote(t, "B", math.MaxInt-1), aspnNote(t, "B", math.MaxInt))
		target := namesChord(t, "B", "B", "C")
		got, err := voiceleading.ChooseConnections(from, target, aspnNote(t, "B", math.MaxInt-1), voiceleading.ConnectionConditions{})
		assert.Nil(t, got)
		require.Error(t, err)
		assertSentinel(t, err, nil)
		ce, ok := errors.AsType[*voiceleading.ChordError](err)
		require.True(t, ok, "want a ChordError, got %v", err)
		assert.Equal(t, 0, ce.Chord)
		assert.Equal(t, 2, ce.Note) // The record of the first C in the composition.
		// The scope of MT21 is checked before any candidate.
		_, err = voiceleading.ChooseConnections(from, target, aspnNote(t, "B", math.MaxInt-1), voiceleading.ConnectionConditions{MT21: true})
		assertSentinel(t, err, voiceleading.ErrOption)
		assertNoChordError(t, err)
	})
	t.Run("stage 9.3: a vertical of the first sonority", func(t *testing.T) {
		t.Parallel()
		from := sonority(t, aspnNote(t, "C", 0), aspnNote(t, "C", m16), aspnNote(t, "C", 2*m16), aspnNote(t, "C", m16))
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "C"), aspnNote(t, "C", 0), voiceleading.ConnectionConditions{})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, system.ErrOutOfRange)
		assertSentinel(t, err, nil)
		assertNoChordError(t, err)
	})
	t.Run("stage 9.4: a vertical of a variant", func(t *testing.T) {
		t.Parallel()
		from := sonority(t, aspnNote(t, "C", 0), aspnNote(t, "C", m16))
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "C"), aspnNote(t, "C", -m16), voiceleading.ConnectionConditions{})
		assert.Nil(t, got)
		assertChordError(t, err, system.ErrOutOfRange, 0, 0)
		assertSentinel(t, err, nil)
	})
	t.Run("stage 11: a total beyond int", func(t *testing.T) {
		t.Parallel()
		from := sonority(t, aspnNote(t, "C", 0), aspnNote(t, "C", 1))
		bass := aspnNote(t, "G", (math.MaxInt-7)/12)
		got, err := voiceleading.ChooseConnections(from, namesChord(t, "G"), bass, voiceleading.ConnectionConditions{})
		assert.Nil(t, got)
		assert.ErrorIs(t, err, system.ErrOutOfRange)
		assertSentinel(t, err, nil)
		assertNoChordError(t, err)
		// Without any surviving variant there is no total to overflow.
		got, err = voiceleading.ChooseConnections(from, namesChord(t, "G"), bass, voiceleading.ConnectionConditions{NoCrossings: true})
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
}

// ---- Reference enumeration by the contract of architecture 10 ----

type referenceCandidate struct {
	p     pitch.Pitch
	class int
}

type referenceVariant struct {
	voices  []pitch.Pitch
	classes []int // class index of each voice; the bass is -1
	total   int
}

func assertMatchesReference(t *testing.T, from *voiceleading.Sonority, target chord.Chord, bass note.Note, conds voiceleading.ConnectionConditions, got []voiceleading.Connection) {
	t.Helper()
	want := referenceChoices(t, from, target, bass, conds)
	require.Len(t, got, len(want), "count of variants")
	for i, w := range want {
		notes := got[i].To.MustNotes()
		require.Len(t, notes, len(w.voices))
		for k, p := range w.voices {
			gp, err := notes[k].Pitch()
			require.NoError(t, err)
			if !assert.True(t, p.Equal(gp), "variant %d voice %d: want %s, got %s (want %s)", i, k, p, gp, describeVariant(w)) {
				return
			}
		}
		assert.Equal(t, system.Step(w.total), got[i].Total, "variant %d total", i)
	}
}

func describeVariant(v referenceVariant) string {
	parts := make([]string, len(v.voices))
	for i, p := range v.voices {
		parts[i] = p.String()
	}
	return fmt.Sprintf("[%s] %d", strings.Join(parts, " "), v.total)
}

func referenceChoices(t *testing.T, from *voiceleading.Sonority, target chord.Chord, bass note.Note, conds voiceleading.ConnectionConditions) []referenceVariant {
	t.Helper()
	sys, ok := from.MustSystem()
	require.True(t, ok)
	fromNotes := from.MustNotes()
	n := len(fromNotes)
	current := make([]pitch.Pitch, n)
	for k, fn := range fromNotes {
		p, err := fn.Pitch()
		require.NoError(t, err)
		current[k] = p
	}
	var classes []spelling.Spelling
	for _, tn := range target.Notes() {
		if !slices.ContainsFunc(classes, tn.Spelling().Equal) {
			classes = append(classes, tn.Spelling())
		}
	}
	bassPitch, err := bass.Pitch()
	require.NoError(t, err)
	return referenceEnumerate(t, sys, current, classes, bassPitch, conds)
}

// referenceEnumerate builds every combination of candidates of the upper
// voices, keeps those that pass the enabled conditions and sorts them.
func referenceEnumerate(t *testing.T, sys *system.System, current []pitch.Pitch, classes []spelling.Spelling, bassPitch pitch.Pitch, conds voiceleading.ConnectionConditions) []referenceVariant {
	t.Helper()
	n := len(current)
	candidates := make([][]referenceCandidate, n)
	for k := 1; k < n; k++ {
		cur := current[k]
		for c, name := range classes {
			var upper, lower pitch.Pitch
			hasUpper, hasLower := false, false
			for o := cur.MustASPN() - 2; o <= cur.MustASPN()+2; o++ {
				q := pitch.MustFromASPN(name, o)
				switch {
				case sys.MustCompare(q, cur) >= 0:
					if !hasUpper || sys.MustCompare(q, upper) < 0 {
						upper, hasUpper = q, true
					}
				default:
					if !hasLower || sys.MustCompare(q, lower) > 0 {
						lower, hasLower = q, true
					}
				}
			}
			require.True(t, hasUpper && hasLower)
			candidates[k] = append(candidates[k], referenceCandidate{upper, c}, referenceCandidate{lower, c})
		}
	}
	var result []referenceVariant
	chosen := make([]referenceCandidate, n)
	chosen[0] = referenceCandidate{bassPitch, -1}
	var walk func(k int)
	walk = func(k int) {
		if k == n {
			voices := make([]pitch.Pitch, n)
			cls := make([]int, n)
			for i, c := range chosen {
				voices[i], cls[i] = c.p, c.class
			}
			if referencePasses(t, sys, current, voices, classes, conds) {
				total := 0
				for i := range voices {
					d := sys.MustDistance(current[i], voices[i])
					total += int(max(d, -d))
				}
				result = append(result, referenceVariant{voices, cls, total})
			}
			return
		}
		for _, c := range candidates[k] {
			chosen[k] = c
			walk(k + 1)
		}
	}
	walk(1)
	slices.SortStableFunc(result, func(a, b referenceVariant) int {
		if c := cmp.Compare(a.total, b.total); c != 0 {
			return c
		}
		for k := n - 1; k >= 0; k-- {
			if c := sys.MustCompare(a.voices[k], b.voices[k]); c != 0 {
				return c
			}
		}
		for k := n - 1; k >= 1; k-- {
			if c := cmp.Compare(a.classes[k], b.classes[k]); c != 0 {
				return c
			}
		}
		return 0
	})
	return result
}

func referencePasses(t *testing.T, sys *system.System, from, to []pitch.Pitch, classes []spelling.Spelling, conds voiceleading.ConnectionConditions) bool {
	t.Helper()
	n := len(from)
	between := func(x, y pitch.Pitch) interval.Interval {
		i, err := interval.BetweenPitches(sys, x, y)
		require.NoError(t, err)
		return i
	}
	inTarget := func(s spelling.Spelling) bool { return slices.ContainsFunc(classes, s.Equal) }
	if conds.Complete {
		for _, c := range classes {
			if !slices.ContainsFunc(to, func(p pitch.Pitch) bool { return p.Spelling().Equal(c) }) {
				return false
			}
		}
	}
	if conds.CommonTones {
		for k := 1; k < n; k++ {
			if inTarget(from[k].Spelling()) && !from[k].Equal(to[k]) {
				return false
			}
		}
	}
	if conds.NoCrossings {
		for k := 0; k+1 < n; k++ {
			if sys.MustCompare(to[k], to[k+1]) > 0 {
				return false
			}
		}
	}
	if conds.NoOverlaps {
		for k := 0; k+1 < n; k++ {
			x0, y0, x1, y1 := from[k], from[k+1], to[k], to[k+1]
			firstAbove := sys.MustCompare(x0, y0) <= 0 && sys.MustCompare(x1, y0) > 0
			secondBelow := sys.MustCompare(y0, x0) >= 0 && sys.MustCompare(y1, x0) < 0
			if firstAbove || secondBelow {
				return false
			}
		}
	}
	if conds.Distances.Rule != voiceleading.DistanceUnset {
		limits := referenceLimits(conds.Distances, n-1)
		for k := 0; k+1 < n; k++ {
			limit := limits[k]
			if limit.IsValid() && between(to[k], to[k+1]).MustUp().Compare(limit.MustUp()) > 0 {
				return false
			}
		}
	}
	if conds.Parallels.Perfect != 0 {
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				mi, mj := between(from[i], to[i]), between(from[j], to[j])
				if mi.IsPerfectUnison() || mj.IsPerfectUnison() || mi.Direction() != mj.Direction() || mi.Size() != mj.Size() {
					continue
				}
				before, after := between(from[i], from[j]), between(to[i], to[j])
				if before.Quality() != interval.Perfect || after.Quality() != interval.Perfect || before.Size() != after.Size() {
					continue
				}
				if referenceSelected(conds.Parallels, before.Size()) {
					return false
				}
			}
		}
	}
	upperHasBass := func(voices []pitch.Pitch) bool {
		return slices.ContainsFunc(voices[1:], func(p pitch.Pitch) bool { return p.Spelling().Equal(voices[0].Spelling()) })
	}
	if conds.BassDoubled && !upperHasBass(to) {
		return false
	}
	if conds.MT21 {
		var distinct []spelling.Spelling
		for _, p := range to {
			if !slices.ContainsFunc(distinct, p.Spelling().Equal) {
				distinct = append(distinct, p.Spelling())
			}
		}
		if len(distinct) != 3 || !upperHasBass(to) {
			return false
		}
		kept, steps, ups, downs := 0, 0, 0, 0
		for k := 1; k < n; k++ {
			mv := between(from[k], to[k])
			switch {
			case mv.IsPerfectUnison():
				kept++
			case mv.Size() == 2:
				steps++
			}
			if !mv.IsPerfectUnison() {
				if mv.Direction() == interval.Up {
					ups++
				} else {
					downs++
				}
			}
		}
		switch between(from[0], to[0]).SimpleSize() {
		case 4, 5:
			commonTone := kept == 1 && steps == 2
			sameDirection := kept == 0 && (ups == 3 || downs == 3)
			if !commonTone && !sameDirection {
				return false
			}
		case 3, 6:
			if kept != 2 || steps != 1 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func referenceLimits(norm voiceleading.DistanceNorm, pairs int) []interval.Interval {
	limits := make([]interval.Interval, pairs)
	switch norm.Rule {
	case voiceleading.DistanceOMTSATB:
		copy(limits, []interval.Interval{interval.PerfectTwelfth(), interval.PerfectOctave(), interval.PerfectOctave()})
	case voiceleading.DistanceMT21SATB:
		copy(limits, []interval.Interval{{}, interval.PerfectOctave(), interval.PerfectOctave()})
	case voiceleading.DistanceOMTFirstSpecies:
		copy(limits, []interval.Interval{interval.PerfectTwelfth()})
	case voiceleading.DistanceCustom:
		for k := range limits {
			if len(norm.Limits) == 1 {
				limits[k] = norm.Limits[0]
			} else {
				limits[k] = norm.Limits[k]
			}
		}
	}
	return limits
}

func referenceSelected(opts voiceleading.ParallelOptions, size int) bool {
	if opts.Perfect&voiceleading.PerfectFifths != 0 {
		if size == 5 || (opts.Compound && size > 5 && (size-5)%7 == 0) {
			return true
		}
	}
	if opts.Perfect&voiceleading.PerfectOctaves != 0 {
		if size == 8 || (opts.Compound && size > 8 && (size-1)%7 == 0) || (opts.Unisons && size == 1) {
			return true
		}
	}
	return false
}
