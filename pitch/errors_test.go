package pitch_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrorsPairwiseDistinct(t *testing.T) {
	t.Parallel()

	require.Error(t, pitch.ErrInvalid)

	sentinels := []struct {
		name string
		err  error
	}{
		{name: "pitch.ErrInvalid", err: pitch.ErrInvalid},
		{name: "alteration.ErrOutOfRange", err: alteration.ErrOutOfRange},
		{name: "octave.ErrOutOfRange", err: octave.ErrOutOfRange},
		{name: "letter.ErrInvalid", err: letter.ErrInvalid},
		{name: "spelling.ErrInvalidForm", err: spelling.ErrInvalidForm},
		{name: "octave.ErrInvalid", err: octave.ErrInvalid},
		{name: "alteration.ErrZeroDenominator", err: alteration.ErrZeroDenominator},
	}

	for i, a := range sentinels {
		assert.ErrorIs(t, a.err, a.err, "sentinel %s must match itself", a.name)
		for j, b := range sentinels {
			if i == j {
				continue
			}
			assert.NotErrorIs(t, a.err, b.err, "sentinel %s must not match %s", a.name, b.name)
		}
	}
}

func TestAllErrorAndMustPairs(t *testing.T) {
	t.Parallel()

	var (
		z  pitch.Pitch
		zs spelling.Spelling
		zo octave.Octave
	)
	c := spelling.C()
	d := spelling.D()
	c4 := pa(c, 4)
	d0 := pa(d, 0)
	dMinus1 := pa(d, -1)
	cMaxOct := pa(c, math.MaxInt)
	cMinOct := pa(c, math.MinInt)
	pMaxAlt := pa(cMaxAlteration(), 4)
	pMinAlt := pa(cMinAlteration(), 4)

	tests := []struct {
		name       string
		call       func() error
		mustFn     func()
		wantTarget error
	}{
		// 1. New / MustNew (E-1)
		{
			name: "New_zero_spelling",
			call: func() error {
				got, err := pitch.New(zs, o(4))
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = pitch.MustNew(zs, o(4)) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "New_zero_octave",
			call: func() error {
				got, err := pitch.New(c, zo)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = pitch.MustNew(c, zo) },
			wantTarget: pitch.ErrInvalid,
		},
		// 2. FromASPN / MustFromASPN (E-1)
		{
			name: "FromASPN_zero_spelling",
			call: func() error {
				got, err := pitch.FromASPN(zs, 4)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = pitch.MustFromASPN(zs, 4) },
			wantTarget: pitch.ErrInvalid,
		},
		// 3. Range / MustRange (E-1)
		{
			name: "Range_zero_spelling",
			call: func() error {
				got, err := pitch.Range(zs, o(2), o(4))
				assert.Nil(t, got)
				return err
			},
			mustFn:     func() { _ = pitch.MustRange(zs, o(2), o(4)) },
			wantTarget: pitch.ErrInvalid,
		},
		// 4. ASPN / MustASPN (E-2)
		{
			name: "ASPN_zero_pitch",
			call: func() error {
				got, err := z.ASPN()
				assert.Equal(t, 0, got)
				return err
			},
			mustFn:     func() { _ = z.MustASPN() },
			wantTarget: pitch.ErrInvalid,
		},
		// 5. WithSpelling / MustWithSpelling (E-1 and E-2)
		{
			name: "WithSpelling_zero_argument",
			call: func() error {
				got, err := c4.WithSpelling(zs)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = c4.MustWithSpelling(zs) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "WithSpelling_zero_receiver",
			call: func() error {
				got, err := z.WithSpelling(c)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustWithSpelling(c) },
			wantTarget: pitch.ErrInvalid,
		},
		// 6. WithOctave / MustWithOctave (E-1 and E-2)
		{
			name: "WithOctave_zero_argument",
			call: func() error {
				got, err := c4.WithOctave(zo)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = c4.MustWithOctave(zo) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "WithOctave_zero_receiver",
			call: func() error {
				got, err := z.WithOctave(o(4))
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustWithOctave(o(4)) },
			wantTarget: pitch.ErrInvalid,
		},
		// 7. WithASPN / MustWithASPN (E-2)
		{
			name: "WithASPN_zero_receiver",
			call: func() error {
				got, err := z.WithASPN(4)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustWithASPN(4) },
			wantTarget: pitch.ErrInvalid,
		},
		// 8. Raise / MustRaise (E-2 and E-3)
		{
			name: "Raise_zero_receiver",
			call: func() error {
				got, err := z.Raise()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustRaise() },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "Raise_out_of_range",
			call: func() error {
				got, err := pMaxAlt.Raise()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = pMaxAlt.MustRaise() },
			wantTarget: alteration.ErrOutOfRange,
		},
		// 9. Lower / MustLower (E-2 and E-3)
		{
			name: "Lower_zero_receiver",
			call: func() error {
				got, err := z.Lower()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustLower() },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "Lower_out_of_range",
			call: func() error {
				got, err := pMinAlt.Lower()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = pMinAlt.MustLower() },
			wantTarget: alteration.ErrOutOfRange,
		},
		// 10. RaiseBy / MustRaiseBy (E-2 and E-3)
		{
			name: "RaiseBy_zero_receiver_with_MaxInt",
			call: func() error {
				got, err := z.RaiseBy(math.MaxInt)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustRaiseBy(math.MaxInt) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "RaiseBy_out_of_range",
			call: func() error {
				got, err := c4.RaiseBy(alteration.MaxMagnitude + 1)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = c4.MustRaiseBy(alteration.MaxMagnitude + 1) },
			wantTarget: alteration.ErrOutOfRange,
		},
		// 11. LowerBy / MustLowerBy (E-2 and E-3)
		{
			name: "LowerBy_zero_receiver_with_MinInt",
			call: func() error {
				got, err := z.LowerBy(math.MinInt)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustLowerBy(math.MinInt) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "LowerBy_out_of_range",
			call: func() error {
				got, err := c4.LowerBy(alteration.MaxMagnitude + 1)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = c4.MustLowerBy(alteration.MaxMagnitude + 1) },
			wantTarget: alteration.ErrOutOfRange,
		},
		// 12. Shift / MustShift (E-2 and E-4)
		{
			name: "Shift_zero_receiver_with_MaxInt",
			call: func() error {
				got, err := z.Shift(math.MaxInt)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustShift(math.MaxInt) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "Shift_out_of_range",
			call: func() error {
				got, err := cMaxOct.Shift(1)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = cMaxOct.MustShift(1) },
			wantTarget: octave.ErrOutOfRange,
		},
		// 13. ShiftDown / MustShiftDown (E-2 and E-4)
		{
			name: "ShiftDown_zero_receiver_with_MinInt",
			call: func() error {
				got, err := z.ShiftDown(math.MinInt)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustShiftDown(math.MinInt) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "ShiftDown_out_of_range",
			call: func() error {
				got, err := cMinOct.ShiftDown(1)
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = cMinOct.MustShiftDown(1) },
			wantTarget: octave.ErrOutOfRange,
		},
		// 14. Up / MustUp (E-2 and E-4)
		{
			name: "Up_zero_receiver",
			call: func() error {
				got, err := z.Up()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustUp() },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "Up_out_of_range",
			call: func() error {
				got, err := cMaxOct.Up()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = cMaxOct.MustUp() },
			wantTarget: octave.ErrOutOfRange,
		},
		// 15. Down / MustDown (E-2 and E-4)
		{
			name: "Down_zero_receiver",
			call: func() error {
				got, err := z.Down()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = z.MustDown() },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "Down_out_of_range",
			call: func() error {
				got, err := cMinOct.Down()
				assert.Equal(t, pitch.Pitch{}, got)
				return err
			},
			mustFn:     func() { _ = cMinOct.MustDown() },
			wantTarget: octave.ErrOutOfRange,
		},
		// 16. OctaveSub / MustOctaveSub (E-2 and E-4)
		{
			name: "OctaveSub_zero_argument_priority",
			call: func() error {
				got, err := cMaxOct.OctaveSub(z)
				assert.Equal(t, 0, got)
				return err
			},
			mustFn:     func() { _ = cMaxOct.MustOctaveSub(z) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "OctaveSub_out_of_range",
			call: func() error {
				got, err := cMaxOct.OctaveSub(dMinus1)
				assert.Equal(t, 0, got)
				return err
			},
			mustFn:     func() { _ = cMaxOct.MustOctaveSub(dMinus1) },
			wantTarget: octave.ErrOutOfRange,
		},
		// 17. OctaveDistance / MustOctaveDistance (E-2 and E-4)
		{
			name: "OctaveDistance_zero_argument_priority",
			call: func() error {
				got, err := cMinOct.OctaveDistance(z)
				assert.Equal(t, 0, got)
				return err
			},
			mustFn:     func() { _ = cMinOct.MustOctaveDistance(z) },
			wantTarget: pitch.ErrInvalid,
		},
		{
			name: "OctaveDistance_out_of_range",
			call: func() error {
				got, err := cMinOct.OctaveDistance(d0)
				assert.Equal(t, 0, got)
				return err
			},
			mustFn:     func() { _ = cMinOct.MustOctaveDistance(d0) },
			wantTarget: octave.ErrOutOfRange,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			require.NotPanics(t, func() {
				err = tc.call()
			})
			requirePitchErrorIs(t, err, tc.wantTarget)
			requireMustPanicsIs(t, tc.mustFn, tc.wantTarget)
		})
	}
}
