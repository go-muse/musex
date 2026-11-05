package pitch_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOctaveMovement(t *testing.T) {
	t.Parallel()

	t.Run("specific_cases_and_boundaries_K11", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			in   pitch.Pitch
			op   string
			k    int
			want pitch.Pitch
		}{
			{
				name: "FSharp_4_Up_to_FSharp_5",
				in:   pa(spelling.FSharp(), 4),
				op:   "Up",
				want: pa(spelling.FSharp(), 5),
			},
			{
				name: "FSharp_4_Down_to_FSharp_3",
				in:   pa(spelling.FSharp(), 4),
				op:   "Down",
				want: pa(spelling.FSharp(), 3),
			},
			{
				name: "FSharp_4_Shift_2_to_FSharp_6",
				in:   pa(spelling.FSharp(), 4),
				op:   "Shift",
				k:    2,
				want: pa(spelling.FSharp(), 6),
			},
			{
				name: "FSharp_4_Shift_minus_5_to_FSharp_minus_1",
				in:   pa(spelling.FSharp(), 4),
				op:   "Shift",
				k:    -5,
				want: pa(spelling.FSharp(), -1),
			},
			{
				name: "FSharp_4_ShiftDown_2_to_FSharp_2",
				in:   pa(spelling.FSharp(), 4),
				op:   "ShiftDown",
				k:    2,
				want: pa(spelling.FSharp(), 2),
			},
			{
				name: "FSharp_4_ShiftDown_minus_2_to_FSharp_6",
				in:   pa(spelling.FSharp(), 4),
				op:   "ShiftDown",
				k:    -2,
				want: pa(spelling.FSharp(), 6),
			},
			{
				name: "C_MinInt_Shift_0",
				in:   pa(spelling.C(), math.MinInt),
				op:   "Shift",
				k:    0,
				want: pa(spelling.C(), math.MinInt),
			},
			{
				name: "C_MaxInt_ShiftDown_0",
				in:   pa(spelling.C(), math.MaxInt),
				op:   "ShiftDown",
				k:    0,
				want: pa(spelling.C(), math.MaxInt),
			},
			{
				name: "C_0_Shift_MinInt_to_C_MinInt",
				in:   pa(spelling.C(), 0),
				op:   "Shift",
				k:    math.MinInt,
				want: pa(spelling.C(), math.MinInt),
			},
			{
				name: "C_MaxInt_Shift_MinInt_to_C_minus_1",
				in:   pa(spelling.C(), math.MaxInt),
				op:   "Shift",
				k:    math.MinInt,
				want: pa(spelling.C(), -1),
			},
			{
				name: "C_minus_1_ShiftDown_MinInt_to_C_MaxInt",
				in:   pa(spelling.C(), -1),
				op:   "ShiftDown",
				k:    math.MinInt,
				want: pa(spelling.C(), math.MaxInt),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				before := tc.in
				var (
					got     pitch.Pitch
					mustGot pitch.Pitch
					err     error
				)
				switch tc.op {
				case "Up":
					got, err = tc.in.Up()
					mustGot = tc.in.MustUp()
				case "Down":
					got, err = tc.in.Down()
					mustGot = tc.in.MustDown()
				case "Shift":
					got, err = tc.in.Shift(tc.k)
					mustGot = tc.in.MustShift(tc.k)
				case "ShiftDown":
					got, err = tc.in.ShiftDown(tc.k)
					mustGot = tc.in.MustShiftDown(tc.k)
				default:
					t.Fatalf("unknown op: %s", tc.op)
				}

				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.want, mustGot)
				assert.True(t, got.SameSpelling(tc.in))
				assert.Equal(t, before, tc.in, "receiver must remain unchanged")
			})
		}

		// Round-trip at math.MinInt (IV-19).
		c0 := pa(spelling.C(), 0)
		assert.Equal(t, c0, c0.MustShift(math.MinInt).MustShiftDown(math.MinInt))

		cMinus1 := pa(spelling.C(), -1)
		assert.Equal(t, cMinus1, cMinus1.MustShiftDown(math.MinInt).MustShift(math.MinInt))
	})

	t.Run("grid_identities_minus_10_to_10_K11", func(t *testing.T) {
		t.Parallel()

		s := spelling.FSharp()
		for a := -10; a <= 10; a++ {
			p := pa(s, a)
			assert.Equal(t, p.MustShift(1), p.MustUp(), "a=%d", a)
			assert.Equal(t, p.MustShift(-1), p.MustDown(), "a=%d", a)
			assert.Equal(t, p.MustShiftDown(1), p.MustDown(), "a=%d", a)

			for k1 := -10; k1 <= 10; k1++ {
				s1, err := p.Shift(k1)
				require.NoError(t, err, "a=%d k1=%d", a, k1)
				assert.Equal(t, s1, p.MustShift(k1), "a=%d k1=%d", a, k1)
				assert.True(t, s1.SameSpelling(p), "a=%d k1=%d", a, k1)

				back, err := s1.ShiftDown(k1)
				require.NoError(t, err, "a=%d k1=%d", a, k1)
				assert.Equal(t, p, back, "a=%d k1=%d", a, k1)
				assert.Equal(t, p, s1.MustShiftDown(k1), "a=%d k1=%d", a, k1)

				sd1, err := p.ShiftDown(k1)
				require.NoError(t, err, "a=%d k1=%d", a, k1)
				assert.Equal(t, p, sd1.MustShift(k1), "a=%d k1=%d", a, k1)

				for k2 := -10; k2 <= 10; k2++ {
					s12 := s1.MustShift(k2)
					want := p.MustShift(k1 + k2)
					assert.Equal(t, want, s12, "a=%d k1=%d k2=%d", a, k1, k2)
					assert.True(t, s12.SameSpelling(p), "a=%d k1=%d k2=%d", a, k1, k2)
				}
			}
		}
	})
}

func TestOctaveSubAndDistance(t *testing.T) {
	t.Parallel()

	t.Run("specific_cases_and_boundaries_K12", func(t *testing.T) {
		t.Parallel()

		subCases := []struct {
			name string
			p    pitch.Pitch
			q    pitch.Pitch
			want int
		}{
			{
				name: "C_5_minus_GSharp_2_is_3",
				p:    pa(spelling.C(), 5),
				q:    pa(spelling.GSharp(), 2),
				want: 3,
			},
			{
				name: "GSharp_2_minus_C_5_is_minus_3",
				p:    pa(spelling.GSharp(), 2),
				q:    pa(spelling.C(), 5),
				want: -3,
			},
			{
				name: "CSharp_5_minus_BFlat_2_is_3",
				p:    pa(spelling.CSharp(), 5),
				q:    pa(spelling.BFlat(), 2),
				want: 3,
			},
			{
				name: "BFlat_2_minus_CSharp_5_is_minus_3",
				p:    pa(spelling.BFlat(), 2),
				q:    pa(spelling.CSharp(), 5),
				want: -3,
			},
			{
				name: "C_4_minus_BSharp_4_same_octave_is_0",
				p:    pa(spelling.C(), 4),
				q:    pa(spelling.BSharp(), 4),
				want: 0,
			},
			{
				name: "CDoubleFlat_4_minus_BDoubleSharp_4_same_octave_is_0",
				p:    pa(spelling.CDoubleFlat(), 4),
				q:    pa(spelling.BDoubleSharp(), 4),
				want: 0,
			},
			{
				name: "C_MinInt_minus_D_0_is_MinInt",
				p:    pa(spelling.C(), math.MinInt),
				q:    pa(spelling.D(), 0),
				want: math.MinInt,
			},
			{
				name: "C_MaxInt_minus_D_0_is_MaxInt",
				p:    pa(spelling.C(), math.MaxInt),
				q:    pa(spelling.D(), 0),
				want: math.MaxInt,
			},
		}

		for _, tc := range subCases {
			t.Run("OctaveSub_"+tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.p.OctaveSub(tc.q)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.want, tc.p.MustOctaveSub(tc.q))
			})
		}

		distCases := []struct {
			name string
			p    pitch.Pitch
			q    pitch.Pitch
			want int
		}{
			{
				name: "CDoubleFlat_4_and_BDoubleSharp_4_is_0",
				p:    pa(spelling.CDoubleFlat(), 4),
				q:    pa(spelling.BDoubleSharp(), 4),
				want: 0,
			},
			{
				name: "C_4_and_BSharp_4_is_0",
				p:    pa(spelling.C(), 4),
				q:    pa(spelling.BSharp(), 4),
				want: 0,
			},
			{
				name: "C_2_and_G_5_is_3",
				p:    pa(spelling.C(), 2),
				q:    pa(spelling.G(), 5),
				want: 3,
			},
			{
				name: "CSharp_5_and_BFlat_2_is_3",
				p:    pa(spelling.CSharp(), 5),
				q:    pa(spelling.BFlat(), 2),
				want: 3,
			},
			{
				name: "C_MinInt_plus_1_and_D_0_is_MaxInt",
				p:    pa(spelling.C(), math.MinInt+1),
				q:    pa(spelling.D(), 0),
				want: math.MaxInt,
			},
		}

		for _, tc := range distCases {
			t.Run("OctaveDistance_"+tc.name, func(t *testing.T) {
				t.Parallel()

				gotPQ, err := tc.p.OctaveDistance(tc.q)
				require.NoError(t, err)
				assert.Equal(t, tc.want, gotPQ)
				assert.Equal(t, tc.want, tc.p.MustOctaveDistance(tc.q))

				gotQP, err := tc.q.OctaveDistance(tc.p)
				require.NoError(t, err)
				assert.Equal(t, tc.want, gotQP)
				assert.Equal(t, tc.want, tc.q.MustOctaveDistance(tc.p))
			})
		}
	})

	t.Run("grid_identities_minus_10_to_10_K12", func(t *testing.T) {
		t.Parallel()

		s := spelling.CSharp()
		tName := spelling.GFlat()
		for a := -10; a <= 10; a++ {
			for b := -10; b <= 10; b++ {
				paPitch := pa(tName, a)
				pbPitch := pa(s, b)
				msg := fmt.Sprintf("a=%d b=%d", a, b)

				diff, err := paPitch.OctaveSub(pbPitch)
				require.NoError(t, err, msg)
				assert.Equal(t, a-b, diff, msg)
				assert.Equal(t, diff, paPitch.MustOctaveSub(pbPitch), msg)
				assert.Equal(t, pa(s, a), pbPitch.MustShift(diff), msg)

				dist, err := paPitch.OctaveDistance(pbPitch)
				require.NoError(t, err, msg)
				wantDist := diff
				if wantDist < 0 {
					wantDist = -wantDist
				}
				assert.Equal(t, wantDist, dist, msg)
				assert.Equal(t, wantDist, pbPitch.MustOctaveDistance(paPitch), msg)
			}
		}
	})
}

func TestOctaveOutOfRange(t *testing.T) {
	t.Parallel()

	t.Run("pitch_ returning_operations_K18", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			call   func() (pitch.Pitch, error)
			mustFn func()
		}{
			{
				name:   "C_MaxInt_Up",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), math.MaxInt).Up() },
				mustFn: func() { _ = pa(spelling.C(), math.MaxInt).MustUp() },
			},
			{
				name:   "C_MaxInt_Shift_1",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), math.MaxInt).Shift(1) },
				mustFn: func() { _ = pa(spelling.C(), math.MaxInt).MustShift(1) },
			},
			{
				name:   "C_MaxInt_ShiftDown_minus_1",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), math.MaxInt).ShiftDown(-1) },
				mustFn: func() { _ = pa(spelling.C(), math.MaxInt).MustShiftDown(-1) },
			},
			{
				name:   "C_MinInt_Down",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), math.MinInt).Down() },
				mustFn: func() { _ = pa(spelling.C(), math.MinInt).MustDown() },
			},
			{
				name:   "C_MinInt_Shift_minus_1",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), math.MinInt).Shift(-1) },
				mustFn: func() { _ = pa(spelling.C(), math.MinInt).MustShift(-1) },
			},
			{
				name:   "C_MinInt_ShiftDown_1",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), math.MinInt).ShiftDown(1) },
				mustFn: func() { _ = pa(spelling.C(), math.MinInt).MustShiftDown(1) },
			},
			{
				name:   "C_1_Shift_MaxInt",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), 1).Shift(math.MaxInt) },
				mustFn: func() { _ = pa(spelling.C(), 1).MustShift(math.MaxInt) },
			},
			{
				name:   "C_0_ShiftDown_MinInt",
				call:   func() (pitch.Pitch, error) { return pa(spelling.C(), 0).ShiftDown(math.MinInt) },
				mustFn: func() { _ = pa(spelling.C(), 0).MustShiftDown(math.MinInt) },
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.call()
				requirePitchErrorIs(t, err, octave.ErrOutOfRange)
				assert.Equal(t, pitch.Pitch{}, got)
				assert.False(t, got.IsValid())
				requireMustPanicsIs(t, tc.mustFn, octave.ErrOutOfRange)
			})
		}
	})

	t.Run("int_returning_operations_K18", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			call   func() (int, error)
			mustFn func()
		}{
			{
				name:   "C_MaxInt_OctaveSub_D_minus_1",
				call:   func() (int, error) { return pa(spelling.C(), math.MaxInt).OctaveSub(pa(spelling.D(), -1)) },
				mustFn: func() { _ = pa(spelling.C(), math.MaxInt).MustOctaveSub(pa(spelling.D(), -1)) },
			},
			{
				name:   "C_0_OctaveSub_D_MinInt",
				call:   func() (int, error) { return pa(spelling.C(), 0).OctaveSub(pa(spelling.D(), math.MinInt)) },
				mustFn: func() { _ = pa(spelling.C(), 0).MustOctaveSub(pa(spelling.D(), math.MinInt)) },
			},
			{
				name:   "C_MinInt_OctaveDistance_D_0",
				call:   func() (int, error) { return pa(spelling.C(), math.MinInt).OctaveDistance(pa(spelling.D(), 0)) },
				mustFn: func() { _ = pa(spelling.C(), math.MinInt).MustOctaveDistance(pa(spelling.D(), 0)) },
			},
			{
				name:   "C_0_OctaveDistance_D_MinInt",
				call:   func() (int, error) { return pa(spelling.C(), 0).OctaveDistance(pa(spelling.D(), math.MinInt)) },
				mustFn: func() { _ = pa(spelling.C(), 0).MustOctaveDistance(pa(spelling.D(), math.MinInt)) },
			},
			{
				name: "C_MinInt_OctaveDistance_D_MaxInt",
				call: func() (int, error) {
					return pa(spelling.C(), math.MinInt).OctaveDistance(pa(spelling.D(), math.MaxInt))
				},
				mustFn: func() { _ = pa(spelling.C(), math.MinInt).MustOctaveDistance(pa(spelling.D(), math.MaxInt)) },
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.call()
				requirePitchErrorIs(t, err, octave.ErrOutOfRange)
				assert.Equal(t, 0, got)
				requireMustPanicsIs(t, tc.mustFn, octave.ErrOutOfRange)
			})
		}
	})
}
