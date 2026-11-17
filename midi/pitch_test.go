package midi_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/midi"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

type midiReferenceCase struct {
	name             string
	written          string
	spelling         spelling.Spelling
	writtenASPN      int
	expectedOctave   int
	expectedClass    int
	expectedPosition int
	expectedNumber   int
	valid            bool
}

var midiReferenceCases = []midiReferenceCase{
	{name: "C-1", written: "C−1", spelling: spelling.C(), writtenASPN: -1, expectedOctave: -1, expectedClass: 0, expectedPosition: -12, expectedNumber: 0, valid: true},
	{name: "C0", written: "C0", spelling: spelling.C(), writtenASPN: 0, expectedOctave: 0, expectedClass: 0, expectedPosition: 0, expectedNumber: 12, valid: true},
	{name: "A0", written: "A0", spelling: spelling.A(), writtenASPN: 0, expectedOctave: 0, expectedClass: 9, expectedPosition: 9, expectedNumber: 21, valid: true},
	{name: "C4", written: "C4", spelling: spelling.C(), writtenASPN: 4, expectedOctave: 4, expectedClass: 0, expectedPosition: 48, expectedNumber: 60, valid: true},
	{name: "A4", written: "A4", spelling: spelling.A(), writtenASPN: 4, expectedOctave: 4, expectedClass: 9, expectedPosition: 57, expectedNumber: 69, valid: true},
	{name: "C5", written: "C5", spelling: spelling.C(), writtenASPN: 5, expectedOctave: 5, expectedClass: 0, expectedPosition: 60, expectedNumber: 72, valid: true},
	{name: "C8", written: "C8", spelling: spelling.C(), writtenASPN: 8, expectedOctave: 8, expectedClass: 0, expectedPosition: 96, expectedNumber: 108, valid: true},
	{name: "G9", written: "G9", spelling: spelling.G(), writtenASPN: 9, expectedOctave: 9, expectedClass: 7, expectedPosition: 115, expectedNumber: 127, valid: true},
	{name: "B#3", written: "B♯3", spelling: spelling.BSharp(), writtenASPN: 3, expectedOctave: 4, expectedClass: 0, expectedPosition: 48, expectedNumber: 60, valid: true},
	{name: "Cb4", written: "C♭4", spelling: spelling.CFlat(), writtenASPN: 4, expectedOctave: 3, expectedClass: 11, expectedPosition: 47, expectedNumber: 59, valid: true},
	{name: "B3", written: "B3", spelling: spelling.B(), writtenASPN: 3, expectedOctave: 3, expectedClass: 11, expectedPosition: 47, expectedNumber: 59, valid: true},
	{name: "B4", written: "B4", spelling: spelling.B(), writtenASPN: 4, expectedOctave: 4, expectedClass: 11, expectedPosition: 59, expectedNumber: 71, valid: true},
	{name: "Cb-1", written: "C♭−1", spelling: spelling.CFlat(), writtenASPN: -1, expectedOctave: -2, expectedClass: 11, expectedPosition: -13, expectedNumber: -1, valid: false},
	{name: "G#9", written: "G♯9", spelling: spelling.GSharp(), writtenASPN: 9, expectedOctave: 9, expectedClass: 8, expectedPosition: 116, expectedNumber: 128, valid: false},
	{name: "B9", written: "B9", spelling: spelling.B(), writtenASPN: 9, expectedOctave: 9, expectedClass: 11, expectedPosition: 119, expectedNumber: 131, valid: false},
	{name: "C10", written: "C10", spelling: spelling.C(), writtenASPN: 10, expectedOctave: 10, expectedClass: 0, expectedPosition: 120, expectedNumber: 132, valid: false},
}

func TestFromPitch_ReferenceRows(t *testing.T) {
	t.Parallel()

	require.Len(t, midiReferenceCases, 16)
	for _, tt := range midiReferenceCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			written := pitch.MustFromASPN(tt.spelling, tt.expectedOctave)
			require.Equal(t, tt.expectedOctave, written.MustASPN())
			number, err := midi.FromPitch(written, system.Traditional())
			if !tt.valid {
				assertMidiError(t, err, midi.ErrOutOfRange)
				require.Equal(t, midi.Number{}, number, "%s has raw number %d before range validation", tt.written, tt.expectedNumber)
				requirePanicError(t, err, func() { midi.MustFromPitch(written, system.Traditional()) })
				return
			}

			require.NoError(t, err)
			requireNumberValue(t, tt.expectedNumber, number)
			requireNumberValue(t, tt.expectedNumber, midi.MustFromPitch(written, system.Traditional()))
			o, class, position, err := number.Parts()
			require.NoError(t, err)
			require.Equal(t, tt.expectedOctave, o.MustASPN())
			require.Equal(t, tt.expectedClass, class.Number())
			require.Equal(t, 12, class.Modulus())
			require.Equal(t, tt.expectedPosition, position)

			for _, rule := range []system.SpellingRule{system.Sharps(), system.Flats()} {
				reversed, err := number.Pitch(system.Traditional(), rule)
				require.NoError(t, err)
				expectedNatural := tt.spelling.Natural()
				require.True(t, reversed.Equal(pitch.MustFromASPN(expectedNatural, tt.expectedOctave)),
					"neutral rule must not reconstruct the original enharmonic spelling")
				roundTrip, err := midi.FromPitch(reversed, system.Traditional())
				require.NoError(t, err)
				requireNumberValue(t, tt.expectedNumber, roundTrip)
			}

			custom := system.MustCustomSpelling(func(_ *system.System, class pitchclass.Class) spelling.Spelling {
				require.Equal(t, tt.expectedClass, class.Number())
				return tt.spelling
			})
			reversed := number.MustPitch(system.Traditional(), custom)
			require.True(t, reversed.Equal(written), "explicit custom rule selects the requested enharmonic spelling")
		})
	}
}

func TestFromPitch_EnharmonicEquivalenceAndWrittenRegisters(t *testing.T) {
	t.Parallel()

	bSharp3 := pitch.MustFromASPN(spelling.BSharp(), 4)
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	cFlat4 := pitch.MustFromASPN(spelling.CFlat(), 3)
	b3 := pitch.MustFromASPN(spelling.B(), 3)
	require.False(t, bSharp3.Equal(c4))
	require.False(t, cFlat4.Equal(b3))
	requireNumberValue(t, 60, midi.MustFromPitch(bSharp3, system.Traditional()))
	requireNumberValue(t, 60, midi.MustFromPitch(c4, system.Traditional()))
	requireNumberValue(t, 59, midi.MustFromPitch(cFlat4, system.Traditional()))
	requireNumberValue(t, 59, midi.MustFromPitch(b3, system.Traditional()))

	for _, tt := range []struct {
		name      string
		rule      system.SpellingRule
		number    int
		expected  pitch.Pitch
		ruleLabel string
	}{
		{name: "sharps C#", rule: system.Sharps(), number: 61, expected: pitch.MustFromASPN(spelling.CSharp(), 4), ruleLabel: "sharps"},
		{name: "flats Db", rule: system.Flats(), number: 61, expected: pitch.MustFromASPN(spelling.DFlat(), 4), ruleLabel: "flats"},
		{name: "sharps D#", rule: system.Sharps(), number: 63, expected: pitch.MustFromASPN(spelling.DSharp(), 4), ruleLabel: "sharps"},
		{name: "flats Eb", rule: system.Flats(), number: 63, expected: pitch.MustFromASPN(spelling.EFlat(), 4), ruleLabel: "flats"},
		{name: "sharps G#", rule: system.Sharps(), number: 68, expected: pitch.MustFromASPN(spelling.GSharp(), 4), ruleLabel: "sharps"},
		{name: "flats Ab", rule: system.Flats(), number: 68, expected: pitch.MustFromASPN(spelling.AFlat(), 4), ruleLabel: "flats"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			number := midi.MustNew(tt.number)
			actual, err := number.Pitch(system.Traditional(), tt.rule)
			require.NoError(t, err)
			require.True(t, actual.Equal(tt.expected), "%s must select the documented spelling", tt.ruleLabel)
			require.True(t, number.MustPitch(system.Traditional(), tt.rule).Equal(tt.expected))
		})
	}
}

func TestFromPitch_UnsetInputs(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		pitch      pitch.Pitch
		system     *system.System
		dependency error
	}{
		{name: "invalid pitch", pitch: pitch.Pitch{}, system: system.Traditional(), dependency: pitch.ErrInvalid},
		{name: "nil system", pitch: pitch.MustFromASPN(spelling.C(), 4), system: nil, dependency: system.ErrInvalid},
		{name: "zero system", pitch: pitch.MustFromASPN(spelling.C(), 4), system: &system.System{}, dependency: system.ErrInvalid},
	} {
		t.Run(tt.name, func(t *testing.T) {
			number, err := midi.FromPitch(tt.pitch, tt.system)
			assertMidiError(t, err, midi.ErrUnset)
			require.ErrorIs(t, err, tt.dependency)
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromPitch(tt.pitch, tt.system) })
		})
	}
}

func naturalOnlySystem() *system.System {
	definition := system.Traditional().Definition()
	definition.Name = "natural-only C layout"
	definition.Alterations = system.Single(alteration.Natural())
	return system.MustNew(definition)
}

func TestFromPitch_LimitedAlterationRule(t *testing.T) {
	t.Parallel()

	sys := naturalOnlySystem()
	requireNumberValue(t, 60, midi.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), sys))
	requireNumberValue(t, 69, midi.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), sys))

	number, err := midi.FromPitch(pitch.MustFromASPN(spelling.CSharp(), 4), sys)
	assertMidiError(t, err, midi.ErrInadmissible)
	require.ErrorIs(t, err, system.ErrInadmissible)
	require.Equal(t, midi.Number{}, number)
	requirePanicError(t, err, func() { midi.MustFromPitch(pitch.MustFromASPN(spelling.CSharp(), 4), sys) })
}

func TestPitch_LayoutValidation(t *testing.T) {
	t.Parallel()

	differentOffsets := system.Traditional().Definition()
	differentOffsets.Offsets[letter.D.Ordinal()] = 3
	differentChromatic := system.Traditional().Definition()
	differentChromatic.Chromatic = 2
	differentBoundary := system.Traditional().Definition()
	differentBoundary.Boundary = letter.A

	for _, tt := range []struct {
		name                 string
		system               *system.System
		hasTraditionalLayout bool
	}{
		{name: "19 positions", system: system.Positions19(), hasTraditionalLayout: false},
		{name: "D offset 3", system: system.MustNew(differentOffsets), hasTraditionalLayout: false},
		{name: "chromatic 2", system: system.MustNew(differentChromatic), hasTraditionalLayout: false},
		{name: "boundary A", system: system.MustNew(differentBoundary), hasTraditionalLayout: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hasTraditional, err := tt.system.HasTraditionalLayout()
			require.NoError(t, err)
			require.Equal(t, tt.hasTraditionalLayout, hasTraditional)

			number, err := midi.FromPitch(pitch.MustFromASPN(spelling.C(), 4), tt.system)
			assertMidiError(t, err, midi.ErrNotTraditionalC)
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), tt.system) })

			written, err := midi.MustNew(60).Pitch(tt.system, system.Sharps())
			assertMidiError(t, err, midi.ErrNotTraditionalC)
			require.False(t, written.IsValid())
			requirePanicError(t, err, func() { midi.MustNew(60).MustPitch(tt.system, system.Sharps()) })
		})
	}
}

func TestFromPitch_LayoutErrorPrecedesInadmissibleName(t *testing.T) {
	t.Parallel()

	definition := system.Traditional().Definition()
	definition.Offsets[letter.D.Ordinal()] = 3
	definition.Alterations = system.Single(alteration.Natural())
	sys := system.MustNew(definition)
	number, err := midi.FromPitch(pitch.MustFromASPN(spelling.CSharp(), 4), sys)
	assertMidiError(t, err, midi.ErrNotTraditionalC)
	require.NotErrorIs(t, err, system.ErrInadmissible)
	require.Equal(t, midi.Number{}, number)
}

func TestFromPitch_ExtremeOctaves(t *testing.T) {
	t.Parallel()

	for _, aspn := range []int{math.MinInt, math.MaxInt} {
		t.Run(fmt.Sprintf("ASPN_%d", aspn), func(t *testing.T) {
			number, err := midi.FromPitch(pitch.MustFromASPN(spelling.C(), aspn), system.Traditional())
			assertMidiError(t, err, midi.ErrOutOfRange)
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromPitch(pitch.MustFromASPN(spelling.C(), aspn), system.Traditional()) })
		})
	}
}

func TestNumber_PitchRuleErrors(t *testing.T) {
	t.Parallel()

	number := midi.MustNew(61)
	written, err := number.Pitch(system.Traditional(), system.SpellingRule{})
	assertMidiError(t, err, midi.ErrNoSpellingRule)
	require.ErrorIs(t, err, system.ErrInvalid)
	require.False(t, written.IsValid())
	requirePanicError(t, err, func() { number.MustPitch(system.Traditional(), system.SpellingRule{}) })

	created, err := number.Note(settingsForSystem(system.Traditional()), system.SpellingRule{}, cents.Zero())
	assertMidiError(t, err, midi.ErrNoSpellingRule)
	require.ErrorIs(t, err, system.ErrInvalid)
	require.False(t, created.IsValid())
	requirePanicError(t, err, func() { number.MustNote(settingsForSystem(system.Traditional()), system.SpellingRule{}, cents.Zero()) })

	for _, tt := range []struct {
		name   string
		choose func(*system.System, pitchclass.Class) spelling.Spelling
	}{
		{name: "unset result", choose: func(*system.System, pitchclass.Class) spelling.Spelling { return spelling.Spelling{} }},
		{name: "inadmissible result", choose: func(*system.System, pitchclass.Class) spelling.Spelling { return spelling.CSharp() }},
		{name: "wrong class result", choose: func(*system.System, pitchclass.Class) spelling.Spelling { return spelling.D() }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rule := system.MustCustomSpelling(tt.choose)
			written, err := number.Pitch(naturalOnlySystem(), rule)
			assertMidiError(t, err, nil)
			require.True(t, errors.Is(err, system.ErrBadSpelling))
			require.False(t, written.IsValid())
			requirePanicError(t, err, func() { number.MustPitch(naturalOnlySystem(), rule) })

			created, err := number.Note(settingsForSystem(naturalOnlySystem()), rule, cents.Zero())
			assertMidiError(t, err, nil)
			require.True(t, errors.Is(err, system.ErrBadSpelling))
			require.False(t, created.IsValid())
			requirePanicError(t, err, func() { number.MustNote(settingsForSystem(naturalOnlySystem()), rule, cents.Zero()) })
		})
	}
}

func TestNumber_PitchNoAdmissibleSpelling(t *testing.T) {
	t.Parallel()

	for _, rule := range []system.SpellingRule{system.Sharps(), system.Flats()} {
		written, err := midi.MustNew(61).Pitch(naturalOnlySystem(), rule)
		assertMidiError(t, err, nil)
		require.True(t, errors.Is(err, system.ErrNoSpelling))
		require.False(t, written.IsValid())
		requirePanicError(t, err, func() { midi.MustNew(61).MustPitch(naturalOnlySystem(), rule) })

		created, err := midi.MustNew(61).Note(settingsForSystem(naturalOnlySystem()), rule, cents.Zero())
		assertMidiError(t, err, nil)
		require.True(t, errors.Is(err, system.ErrNoSpelling))
		require.False(t, created.IsValid())
		requirePanicError(t, err, func() { midi.MustNew(61).MustNote(settingsForSystem(naturalOnlySystem()), rule, cents.Zero()) })
	}
}

func TestNumber_PitchRejectsChromaticScale(t *testing.T) {
	t.Parallel()

	tonics := []spelling.Spelling{spelling.C(), spelling.G(), spelling.CSharp()}
	for _, tonic := range tonics {
		for _, mode := range []system.ScaleMode{system.Major, system.Minor} {
			for _, motion := range []system.Motion{system.Ascending, system.Descending} {
				name := fmt.Sprintf("tonic_%s_mode_%s_motion_%s", tonic, mode, motion)
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					rule := system.MustChromaticScale(tonic, mode, motion)
					written, err := midi.MustNew(60).Pitch(system.Traditional(), rule)
					assertMidiError(t, err, midi.ErrNotApplicable)
					require.NotErrorIs(t, err, system.ErrNotApplicable)
					require.False(t, written.IsValid())
					requirePanicError(t, err, func() { midi.MustNew(60).MustPitch(system.Traditional(), rule) })

					created, err := midi.MustNew(60).Note(settingsForSystem(system.Traditional()), rule, cents.Zero())
					assertMidiError(t, err, midi.ErrNotApplicable)
					require.NotErrorIs(t, err, system.ErrNotApplicable)
					require.False(t, created.IsValid())
					requirePanicError(t, err, func() { midi.MustNew(60).MustNote(settingsForSystem(system.Traditional()), rule, cents.Zero()) })
				})
			}
		}
	}
}

func TestNumber_PitchRejectsChromaticScaleBeforeSystem(t *testing.T) {
	t.Parallel()

	rule := system.MustChromaticScale(spelling.CSharp(), system.Major, system.Ascending)
	limited := naturalOnlySystem()
	written, err := midi.MustNew(61).Pitch(limited, rule)
	assertMidiError(t, err, midi.ErrNotApplicable)
	require.NotErrorIs(t, err, system.ErrNotApplicable)
	require.NotErrorIs(t, err, system.ErrInadmissible)
	require.NotErrorIs(t, err, system.ErrNoSpelling)
	require.NotErrorIs(t, err, system.ErrBadSpelling)
	require.False(t, written.IsValid())

	created, err := midi.MustNew(61).Note(settingsForSystem(limited), rule, cents.Zero())
	assertMidiError(t, err, midi.ErrNotApplicable)
	require.NotErrorIs(t, err, system.ErrNotApplicable)
	require.NotErrorIs(t, err, system.ErrInadmissible)
	require.NotErrorIs(t, err, system.ErrNoSpelling)
	require.NotErrorIs(t, err, system.ErrBadSpelling)
	require.False(t, created.IsValid())
}
