package pitch_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRaiseAndLower(t *testing.T) {
	t.Parallel()

	t.Run("specific_cases_K10", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			in   pitch.Pitch
			op   string
			n    int
			want pitch.Pitch
		}{
			{
				name: "C_4_Raise_to_CSharp_4",
				in:   pa(spelling.C(), 4),
				op:   "Raise",
				want: pa(spelling.CSharp(), 4),
			},
			{
				name: "CSharp_4_Raise_to_CDoubleSharp_4",
				in:   pa(spelling.CSharp(), 4),
				op:   "Raise",
				want: pa(spelling.CDoubleSharp(), 4),
			},
			{
				name: "B_4_Raise_to_BSharp_4_keeps_ASPN_4",
				in:   pa(spelling.B(), 4),
				op:   "Raise",
				want: pa(spelling.BSharp(), 4),
			},
			{
				name: "C_4_Lower_to_CFlat_4",
				in:   pa(spelling.C(), 4),
				op:   "Lower",
				want: pa(spelling.CFlat(), 4),
			},
			{
				name: "B_4_Lower_to_BFlat_4",
				in:   pa(spelling.B(), 4),
				op:   "Lower",
				want: pa(spelling.BFlat(), 4),
			},
			{
				name: "DFlat_4_Lower_to_DDoubleFlat_4",
				in:   pa(spelling.DFlat(), 4),
				op:   "Lower",
				want: pa(spelling.DDoubleFlat(), 4),
			},
			{
				name: "C_plus_half_4_Raise_to_C_plus_3_halves_4",
				in:   pa(cPlusHalf(), 4),
				op:   "Raise",
				want: pa(cPlusThreeHalves(), 4),
			},
			{
				name: "C_plus_half_4_Lower_to_C_minus_half_4",
				in:   pa(cPlusHalf(), 4),
				op:   "Lower",
				want: pa(cMinusHalf(), 4),
			},
			{
				name: "C_plus_half_4_RaiseBy_1_to_C_plus_3_halves_4",
				in:   pa(cPlusHalf(), 4),
				op:   "RaiseBy",
				n:    1,
				want: pa(cPlusThreeHalves(), 4),
			},
			{
				name: "C_plus_half_4_LowerBy_1_to_C_minus_half_4",
				in:   pa(cPlusHalf(), 4),
				op:   "LowerBy",
				n:    1,
				want: pa(cMinusHalf(), 4),
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
				case "Raise":
					got, err = tc.in.Raise()
					mustGot = tc.in.MustRaise()
				case "Lower":
					got, err = tc.in.Lower()
					mustGot = tc.in.MustLower()
				case "RaiseBy":
					got, err = tc.in.RaiseBy(tc.n)
					mustGot = tc.in.MustRaiseBy(tc.n)
				case "LowerBy":
					got, err = tc.in.LowerBy(tc.n)
					mustGot = tc.in.MustLowerBy(tc.n)
				default:
					t.Fatalf("unknown op: %s", tc.op)
				}

				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.want, mustGot)
				assert.True(t, got.SameLetter(tc.in))
				assert.True(t, got.SameOctave(tc.in))
				assert.Equal(t, before, tc.in, "receiver must remain unchanged")
			})
		}
	})

	t.Run("algebraic_identities_K10", func(t *testing.T) {
		t.Parallel()

		spellings := []spelling.Spelling{
			spelling.C(),
			spelling.CSharp(),
			spelling.DFlat(),
			cPlusHalf(),
		}
		ns := []int{-5, -1, 0, 1, 5}

		for _, s := range spellings {
			p := pa(s, 4)

			r0, err := p.RaiseBy(0)
			require.NoError(t, err)
			assert.Equal(t, p, r0)
			assert.Equal(t, p, p.MustRaiseBy(0))

			l0, err := p.LowerBy(0)
			require.NoError(t, err)
			assert.Equal(t, p, l0)
			assert.Equal(t, p, p.MustLowerBy(0))

			r1, err := p.Raise()
			require.NoError(t, err)
			assert.Equal(t, p.MustRaiseBy(1), r1)
			assert.Equal(t, p.MustRaiseBy(1), p.MustRaise())

			l1, err := p.Lower()
			require.NoError(t, err)
			assert.Equal(t, p.MustLowerBy(1), l1)
			assert.Equal(t, p.MustLowerBy(1), p.MustLower())

			for _, n := range ns {
				msg := fmt.Sprintf("s=%v n=%d", s, n)

				raised, err := p.RaiseBy(n)
				require.NoError(t, err, msg)
				assert.Equal(t, raised, p.MustRaiseBy(n), msg)
				assert.True(t, raised.SameLetter(p), msg)
				assert.True(t, raised.SameOctave(p), msg)

				lowered, err := p.LowerBy(n)
				require.NoError(t, err, msg)
				assert.Equal(t, lowered, p.MustLowerBy(n), msg)
				assert.True(t, lowered.SameLetter(p), msg)
				assert.True(t, lowered.SameOctave(p), msg)

				assert.Equal(t, p.MustRaiseBy(-n), lowered, msg)
				assert.Equal(t, p, raised.MustLowerBy(n), msg)
				assert.Equal(t, p, lowered.MustRaiseBy(n), msg)
			}
		}
	})
}

func TestAlterationBoundariesAndOutOfRange(t *testing.T) {
	t.Parallel()

	const m = alteration.MaxMagnitude
	pMax := pa(cMaxAlteration(), 4)
	pMin := pa(cMinAlteration(), 4)
	c4 := pa(spelling.C(), 4)

	t.Run("exact_boundary_succeeds_K17", func(t *testing.T) {
		t.Parallel()

		gotMax, err := c4.RaiseBy(m)
		require.NoError(t, err)
		assert.Equal(t, pMax, gotMax)
		assert.Equal(t, pMax, c4.MustRaiseBy(m))

		gotMin, err := c4.LowerBy(m)
		require.NoError(t, err)
		assert.Equal(t, pMin, gotMin)
		assert.Equal(t, pMin, c4.MustLowerBy(m))

		assert.Equal(t, c4, pMax.MustLowerBy(m))
		assert.Equal(t, c4, pMin.MustRaiseBy(m))
		assert.Equal(t, pMin, pMax.MustLowerBy(2*m))
		assert.Equal(t, pMax, pMin.MustRaiseBy(2*m))
	})

	t.Run("out_of_range_returns_alteration_ErrOutOfRange_K17", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			call   func() (pitch.Pitch, error)
			mustFn func()
		}{
			{
				name:   "pMax_Raise",
				call:   func() (pitch.Pitch, error) { return pMax.Raise() },
				mustFn: func() { _ = pMax.MustRaise() },
			},
			{
				name:   "pMax_RaiseBy_1",
				call:   func() (pitch.Pitch, error) { return pMax.RaiseBy(1) },
				mustFn: func() { _ = pMax.MustRaiseBy(1) },
			},
			{
				name:   "pMax_LowerBy_minus_1",
				call:   func() (pitch.Pitch, error) { return pMax.LowerBy(-1) },
				mustFn: func() { _ = pMax.MustLowerBy(-1) },
			},
			{
				name:   "pMin_Lower",
				call:   func() (pitch.Pitch, error) { return pMin.Lower() },
				mustFn: func() { _ = pMin.MustLower() },
			},
			{
				name:   "pMin_LowerBy_1",
				call:   func() (pitch.Pitch, error) { return pMin.LowerBy(1) },
				mustFn: func() { _ = pMin.MustLowerBy(1) },
			},
			{
				name:   "pMin_RaiseBy_minus_1",
				call:   func() (pitch.Pitch, error) { return pMin.RaiseBy(-1) },
				mustFn: func() { _ = pMin.MustRaiseBy(-1) },
			},
			{
				name:   "C4_RaiseBy_M_plus_1",
				call:   func() (pitch.Pitch, error) { return c4.RaiseBy(m + 1) },
				mustFn: func() { _ = c4.MustRaiseBy(m + 1) },
			},
			{
				name:   "C4_LowerBy_M_plus_1",
				call:   func() (pitch.Pitch, error) { return c4.LowerBy(m + 1) },
				mustFn: func() { _ = c4.MustLowerBy(m + 1) },
			},
			{
				name:   "C4_RaiseBy_MaxInt",
				call:   func() (pitch.Pitch, error) { return c4.RaiseBy(math.MaxInt) },
				mustFn: func() { _ = c4.MustRaiseBy(math.MaxInt) },
			},
			{
				name:   "C4_LowerBy_MinInt",
				call:   func() (pitch.Pitch, error) { return c4.LowerBy(math.MinInt) },
				mustFn: func() { _ = c4.MustLowerBy(math.MinInt) },
			},
			{
				name:   "C4_RaiseBy_MinInt",
				call:   func() (pitch.Pitch, error) { return c4.RaiseBy(math.MinInt) },
				mustFn: func() { _ = c4.MustRaiseBy(math.MinInt) },
			},
			{
				name:   "C4_LowerBy_MaxInt",
				call:   func() (pitch.Pitch, error) { return c4.LowerBy(math.MaxInt) },
				mustFn: func() { _ = c4.MustLowerBy(math.MaxInt) },
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.call()
				requirePitchErrorIs(t, err, alteration.ErrOutOfRange)
				assert.Equal(t, pitch.Pitch{}, got)
				assert.False(t, got.IsValid())
				requireMustPanicsIs(t, tc.mustFn, alteration.ErrOutOfRange)
			})
		}
	})

	t.Run("zero_pitch_returns_pitch_ErrInvalid_even_for_extreme_n_K16", func(t *testing.T) {
		t.Parallel()

		var z pitch.Pitch
		cases := []struct {
			name   string
			call   func() (pitch.Pitch, error)
			mustFn func()
		}{
			{
				name:   "zero_Raise",
				call:   func() (pitch.Pitch, error) { return z.Raise() },
				mustFn: func() { _ = z.MustRaise() },
			},
			{
				name:   "zero_Lower",
				call:   func() (pitch.Pitch, error) { return z.Lower() },
				mustFn: func() { _ = z.MustLower() },
			},
			{
				name:   "zero_RaiseBy_0",
				call:   func() (pitch.Pitch, error) { return z.RaiseBy(0) },
				mustFn: func() { _ = z.MustRaiseBy(0) },
			},
			{
				name:   "zero_LowerBy_0",
				call:   func() (pitch.Pitch, error) { return z.LowerBy(0) },
				mustFn: func() { _ = z.MustLowerBy(0) },
			},
			{
				name:   "zero_RaiseBy_MaxInt",
				call:   func() (pitch.Pitch, error) { return z.RaiseBy(math.MaxInt) },
				mustFn: func() { _ = z.MustRaiseBy(math.MaxInt) },
			},
			{
				name:   "zero_LowerBy_MinInt",
				call:   func() (pitch.Pitch, error) { return z.LowerBy(math.MinInt) },
				mustFn: func() { _ = z.MustLowerBy(math.MinInt) },
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.call()
				requirePitchErrorIs(t, err, pitch.ErrInvalid)
				assert.Equal(t, pitch.Pitch{}, got)
				requireMustPanicsIs(t, tc.mustFn, pitch.ErrInvalid)
			})
		}
	})
}
