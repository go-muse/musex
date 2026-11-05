package pitch_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

func TestZeroValueErrorFreeOperations(t *testing.T) {
	t.Parallel()

	var (
		zVar pitch.Pitch
		zs   spelling.Spelling
		zo   octave.Octave
	)
	zLit := pitch.Pitch{}
	c4 := pa(spelling.C(), 4)
	c0 := pa(spelling.C(), 0)
	cMin := pa(spelling.C(), math.MinInt)

	for _, z := range []pitch.Pitch{zVar, zLit} {
		// K-16 item 1: selectors, predicates, Natural, String, GoString.
		assert.False(t, z.IsValid())
		assert.Equal(t, zs, z.Spelling())
		assert.False(t, z.Spelling().IsValid())
		assert.Equal(t, zo, z.Octave())
		assert.False(t, z.Octave().IsValid())

		gotS, gotO := z.Parts()
		assert.Equal(t, zs, gotS)
		assert.Equal(t, zo, gotO)

		assert.False(t, z.IsNatural())
		assert.False(t, z.IsInteger())
		assert.Equal(t, spelling.DirectionUnset, z.Direction())
		assert.Equal(t, zLit, z.Natural())
		assert.False(t, z.Natural().IsValid())
		assert.Equal(t, "Pitch(invalid)", z.String())
		assert.Equal(t, "pitch.Pitch{}", z.GoString())

		// K-16 item 2: Equal, ==, SameSpelling, SameOctave, SameLetter, SameAlteration.
		assert.False(t, z.Equal(z))
		assert.False(t, z.Equal(c4))
		assert.False(t, c4.Equal(z))
		assert.False(t, z.Equal(c0))
		assert.False(t, c0.Equal(z))

		// Built-in == is true for two zero structs, false against any valid Pitch.
		assert.Equal(t, zLit, z)
		assert.NotEqual(t, c4, z)
		assert.NotEqual(t, c0, z)
		assert.NotEqual(t, cMin, z)

		assert.False(t, z.SameSpelling(z))
		assert.False(t, z.SameSpelling(c4))
		assert.False(t, c4.SameSpelling(z))

		assert.False(t, z.SameOctave(z))
		assert.False(t, z.SameOctave(c4))
		assert.False(t, c4.SameOctave(z))

		assert.False(t, z.SameLetter(z))
		assert.False(t, z.SameLetter(c4))
		assert.False(t, c4.SameLetter(z))

		assert.False(t, z.SameAlteration(z))
		assert.False(t, z.SameAlteration(c4))
		assert.False(t, c4.SameAlteration(z))

		// K-16 item 3: Compare.
		assert.Equal(t, 0, z.Compare(z))
		assert.Equal(t, -1, z.Compare(cMin))
		assert.Equal(t, 1, cMin.Compare(z))
		assert.Equal(t, -1, z.Compare(c4))
		assert.Equal(t, 1, c4.Compare(z))
	}
}

func TestZeroValueErrorReturningOperations(t *testing.T) {
	t.Parallel()

	var (
		z  pitch.Pitch
		zs spelling.Spelling
		zo octave.Octave
	)
	c := spelling.C()
	c4 := pa(c, 4)
	cMax := pa(c, math.MaxInt)
	cMin := pa(c, math.MinInt)

	pitchOps := []struct {
		name   string
		call   func() (pitch.Pitch, error)
		mustFn func()
	}{
		{
			name:   "WithSpelling_valid",
			call:   func() (pitch.Pitch, error) { return z.WithSpelling(c) },
			mustFn: func() { _ = z.MustWithSpelling(c) },
		},
		{
			name:   "WithSpelling_zero",
			call:   func() (pitch.Pitch, error) { return z.WithSpelling(zs) },
			mustFn: func() { _ = z.MustWithSpelling(zs) },
		},
		{
			name:   "WithOctave_valid",
			call:   func() (pitch.Pitch, error) { return z.WithOctave(o(4)) },
			mustFn: func() { _ = z.MustWithOctave(o(4)) },
		},
		{
			name:   "WithOctave_zero",
			call:   func() (pitch.Pitch, error) { return z.WithOctave(zo) },
			mustFn: func() { _ = z.MustWithOctave(zo) },
		},
		{
			name:   "WithASPN_4",
			call:   func() (pitch.Pitch, error) { return z.WithASPN(4) },
			mustFn: func() { _ = z.MustWithASPN(4) },
		},
		{
			name:   "Raise",
			call:   func() (pitch.Pitch, error) { return z.Raise() },
			mustFn: func() { _ = z.MustRaise() },
		},
		{
			name:   "Lower",
			call:   func() (pitch.Pitch, error) { return z.Lower() },
			mustFn: func() { _ = z.MustLower() },
		},
		{
			name:   "RaiseBy_0",
			call:   func() (pitch.Pitch, error) { return z.RaiseBy(0) },
			mustFn: func() { _ = z.MustRaiseBy(0) },
		},
		{
			name:   "RaiseBy_1",
			call:   func() (pitch.Pitch, error) { return z.RaiseBy(1) },
			mustFn: func() { _ = z.MustRaiseBy(1) },
		},
		{
			name:   "RaiseBy_MaxInt_priority_over_out_of_range",
			call:   func() (pitch.Pitch, error) { return z.RaiseBy(math.MaxInt) },
			mustFn: func() { _ = z.MustRaiseBy(math.MaxInt) },
		},
		{
			name:   "LowerBy_0",
			call:   func() (pitch.Pitch, error) { return z.LowerBy(0) },
			mustFn: func() { _ = z.MustLowerBy(0) },
		},
		{
			name:   "LowerBy_1",
			call:   func() (pitch.Pitch, error) { return z.LowerBy(1) },
			mustFn: func() { _ = z.MustLowerBy(1) },
		},
		{
			name:   "LowerBy_MinInt_priority_over_out_of_range",
			call:   func() (pitch.Pitch, error) { return z.LowerBy(math.MinInt) },
			mustFn: func() { _ = z.MustLowerBy(math.MinInt) },
		},
		{
			name:   "Up",
			call:   func() (pitch.Pitch, error) { return z.Up() },
			mustFn: func() { _ = z.MustUp() },
		},
		{
			name:   "Down",
			call:   func() (pitch.Pitch, error) { return z.Down() },
			mustFn: func() { _ = z.MustDown() },
		},
		{
			name:   "Shift_0",
			call:   func() (pitch.Pitch, error) { return z.Shift(0) },
			mustFn: func() { _ = z.MustShift(0) },
		},
		{
			name:   "Shift_1",
			call:   func() (pitch.Pitch, error) { return z.Shift(1) },
			mustFn: func() { _ = z.MustShift(1) },
		},
		{
			name:   "Shift_MaxInt_priority_over_out_of_range",
			call:   func() (pitch.Pitch, error) { return z.Shift(math.MaxInt) },
			mustFn: func() { _ = z.MustShift(math.MaxInt) },
		},
		{
			name:   "ShiftDown_0",
			call:   func() (pitch.Pitch, error) { return z.ShiftDown(0) },
			mustFn: func() { _ = z.MustShiftDown(0) },
		},
		{
			name:   "ShiftDown_1",
			call:   func() (pitch.Pitch, error) { return z.ShiftDown(1) },
			mustFn: func() { _ = z.MustShiftDown(1) },
		},
		{
			name:   "ShiftDown_MinInt_priority_over_out_of_range",
			call:   func() (pitch.Pitch, error) { return z.ShiftDown(math.MinInt) },
			mustFn: func() { _ = z.MustShiftDown(math.MinInt) },
		},
	}

	for _, tc := range pitchOps {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.call()
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Equal(t, pitch.Pitch{}, got)
			assert.False(t, got.IsValid())

			requireMustPanicsIs(t, tc.mustFn, pitch.ErrInvalid)
		})
	}

	intOps := []struct {
		name   string
		call   func() (int, error)
		mustFn func()
	}{
		{
			name:   "ASPN",
			call:   func() (int, error) { return z.ASPN() },
			mustFn: func() { _ = z.MustASPN() },
		},
		{
			name:   "OctaveSub_zero_receiver",
			call:   func() (int, error) { return z.OctaveSub(c4) },
			mustFn: func() { _ = z.MustOctaveSub(c4) },
		},
		{
			name:   "OctaveSub_zero_argument",
			call:   func() (int, error) { return c4.OctaveSub(z) },
			mustFn: func() { _ = c4.MustOctaveSub(z) },
		},
		{
			name:   "OctaveSub_MaxInt_receiver_and_zero_argument",
			call:   func() (int, error) { return cMax.OctaveSub(z) },
			mustFn: func() { _ = cMax.MustOctaveSub(z) },
		},
		{
			name:   "OctaveSub_both_zero",
			call:   func() (int, error) { return z.OctaveSub(z) },
			mustFn: func() { _ = z.MustOctaveSub(z) },
		},
		{
			name:   "OctaveDistance_zero_receiver",
			call:   func() (int, error) { return z.OctaveDistance(c4) },
			mustFn: func() { _ = z.MustOctaveDistance(c4) },
		},
		{
			name:   "OctaveDistance_zero_argument",
			call:   func() (int, error) { return c4.OctaveDistance(z) },
			mustFn: func() { _ = c4.MustOctaveDistance(z) },
		},
		{
			name:   "OctaveDistance_zero_receiver_and_MinInt_argument",
			call:   func() (int, error) { return z.OctaveDistance(cMin) },
			mustFn: func() { _ = z.MustOctaveDistance(cMin) },
		},
		{
			name:   "OctaveDistance_MinInt_receiver_and_zero_argument",
			call:   func() (int, error) { return cMin.OctaveDistance(z) },
			mustFn: func() { _ = cMin.MustOctaveDistance(z) },
		},
		{
			name:   "OctaveDistance_both_zero",
			call:   func() (int, error) { return z.OctaveDistance(z) },
			mustFn: func() { _ = z.MustOctaveDistance(z) },
		},
	}

	for _, tc := range intOps {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.call()
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Equal(t, 0, got)

			requireMustPanicsIs(t, tc.mustFn, pitch.ErrInvalid)
		})
	}
}
