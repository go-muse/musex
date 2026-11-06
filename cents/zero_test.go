package cents_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1, G-10: 0, the octave and the semitone; the zero value of the type is 0.
func TestBasicValues(t *testing.T) {
	t.Parallel()

	var zero cents.Cents

	t.Run("zero", func(t *testing.T) {
		t.Parallel()

		zeros := []struct {
			name string
			v    cents.Cents
		}{
			{"Zero()", cents.Zero()},
			{"Cents{}", cents.Cents{}},
			{"var", zero},
			{"FromInt(0)", noErr(cents.FromInt(0))(t)},
			{"MustFromInt(0)", cents.MustFromInt(0)},
			{"Step(0, 19)", noErr(cents.Step(0, 19))(t)},
			{"FromRatio(1, 1)", noErr(cents.FromRatio(1, 1))(t)},
			{"FromFraction(0, 5)", noErr(cents.FromFraction(0, 5))(t)},
		}

		for _, a := range zeros {
			for _, b := range zeros {
				assert.True(t, a.v.Equal(b.v), "%s.Equal(%s)", a.name, b.name)
				assert.True(t, a.v == b.v, "%s == %s", a.name, b.name)
			}

			assertComp(t, zeroComp(), a.v, a.name)
		}
	})

	cases := []struct {
		name string
		got  cents.Cents
		want []cents.Cents
		comp comp
	}{
		{
			name: "octave",
			got:  cents.Octave(),
			want: []cents.Cents{noErr(cents.FromInt(1200))(t), noErr(cents.FromRatio(2, 1))(t), noErr(cents.Step(12, 12))(t)},
			comp: ratC(1200, 1),
		},
		{
			name: "semitone",
			got:  cents.Semitone(),
			want: []cents.Cents{noErr(cents.FromInt(100))(t), noErr(cents.Step(1, 12))(t)},
			comp: ratC(100, 1),
		},
		{
			name: "cent",
			got:  cents.Cent(),
			want: []cents.Cents{noErr(cents.FromInt(1))(t), noErr(cents.Step(1, 1200))(t)},
			comp: ratC(1, 1),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, w := range tc.want {
				assertCents(t, w, tc.got)
			}

			assertComp(t, tc.comp, tc.got)
		})
	}
}

// The table "Zero value" of the architecture, row by row; the zero value
// passes K-10...K-20 as 0 without errors (K-1).
func TestZeroValue(t *testing.T) {
	t.Parallel()

	var z cents.Cents

	t.Run("equal", func(t *testing.T) {
		t.Parallel()

		assert.True(t, z.Equal(cents.Zero()))
		assert.True(t, z == cents.Zero())
		assert.True(t, z == cents.Cents{})
		assert.Equal(t, 0, z.Sign())
		assert.Equal(t, 0, z.Compare(cents.Zero()))
		assert.Equal(t, 0, cents.Zero().Compare(z))
	})

	t.Run("results are the zero value", func(t *testing.T) {
		t.Parallel()

		assert.True(t, z.Neg() == cents.Cents{}, "Neg")
		assert.True(t, z.Abs() == cents.Cents{}, "Abs")

		for _, k := range []int{minInt, -4, -1, 0, 1, 2, 7501, maxInt} {
			assert.True(t, noErr(z.Mul(k))(t) == cents.Cents{}, "Mul(%d)", k)
			assert.True(t, noErr(z.Scale(k, 3))(t) == cents.Cents{}, "Scale(%d, 3)", k)
		}

		for _, n := range []int{minInt, -12, -1, 1, 2, 12, maxInt} {
			assert.True(t, noErr(z.Div(n))(t) == cents.Cents{}, "Div(%d)", n)
			assert.True(t, noErr(z.Scale(1, n))(t) == cents.Cents{}, "Scale(1, %d)", n)
		}

		for _, d := range []int{1, 2, 1000, maxInt} {
			assert.True(t, noErr(z.RoundTo(d))(t) == cents.Cents{}, "RoundTo(%d)", d)
		}

		assert.True(t, noErr(z.ScaleRat(big.NewRat(3, 7)))(t) == cents.Cents{}, "ScaleRat")
		assert.True(t, noErr(z.ScaleRat(nil))(t) == cents.Cents{}, "ScaleRat(nil)")
	})

	t.Run("add and sub", func(t *testing.T) {
		t.Parallel()

		for _, d := range setS(t) {
			assertCents(t, d.v, noErr(z.Add(d.v))(t), "0 + %s", d.name)
			assertCents(t, d.v, noErr(d.v.Add(z))(t), "%s + 0", d.name)
			assertCents(t, d.v.Neg(), noErr(z.Sub(d.v))(t), "0 - %s", d.name)
			assertCents(t, d.v, noErr(d.v.Sub(z))(t), "%s - 0", d.name)
			assert.True(t, noErr(d.v.Sub(d.v))(t) == cents.Cents{}, "%s - %s is the zero value", d.name, d.name)
		}
	})

	t.Run("split, nearest step, round", func(t *testing.T) {
		t.Parallel()

		octaves, rest := z.SplitOctaves()
		assert.Equal(t, 0, octaves)
		assert.True(t, rest == cents.Cents{})

		for _, n := range []int{1, 12, 19, 53, maxInt} {
			k, dev, err := z.NearestStep(n)
			require.NoError(t, err)
			assert.Equal(t, 0, k, "NearestStep(%d)", n)
			assert.True(t, dev == cents.Cents{}, "NearestStep(%d)", n)
		}
	})

	t.Run("quotient", func(t *testing.T) {
		t.Parallel()

		for _, d := range setS(t) {
			if d.c.sign() == 0 {
				continue
			}

			q, err := z.Quo(d.v)
			require.NoError(t, err, d.name)

			exact, ok := q.Exact()
			require.True(t, ok, d.name)
			assert.Equal(t, 0, exact.Sign(), d.name)

			_, err = d.v.Quo(z)
			assertOnlyError(t, err, cents.ErrInvalidArgument, "%s / zero value", d.name)

			_, err = d.v.Quo(cents.Zero())
			assertOnlyError(t, err, cents.ErrInvalidArgument, "%s / Zero()", d.name)
		}

		_, err := z.Quo(z)
		assertOnlyError(t, err, cents.ErrInvalidArgument, "0 / 0")
	})

	t.Run("values", func(t *testing.T) {
		t.Parallel()

		f := z.Float64()
		assert.Zero(t, f)
		assert.False(t, math.Signbit(f), "Float64 is +0")

		num, den, ok := z.Fraction()
		assert.Equal(t, []any{0, 1, true}, []any{num, den, ok})

		num, den, ok = z.Ratio()
		assert.Equal(t, []any{1, 1, true}, []any{num, den, ok})

		bn, bd, ps := z.Composition()
		require.NotNil(t, bn)
		require.NotNil(t, bd)
		assert.Equal(t, 0, bn.Sign())
		assert.Equal(t, 0, bd.Cmp(big.NewInt(1)))
		assert.Empty(t, ps)
	})

	t.Run("frequencies", func(t *testing.T) {
		t.Parallel()

		m, err := z.Multiplier()
		require.NoError(t, err)
		assert.Equal(t, 1.0, m)

		for _, f := range []float64{440, 0.1, 0x1p-1022, math.MaxFloat64} {
			got, err := z.ApplyTo(f)
			require.NoError(t, err, "%v", f)
			assert.Equal(t, f, got)
		}

		got, err := z.ApplyTo(0x1p-1074)
		assertOnlyError(t, err, cents.ErrOutOfRange, "a subnormal result")
		assert.Zero(t, got)
	})

	t.Run("text", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, cents.KindDecimal, z.Kind())
		assert.Equal(t, "0", z.String())
		assert.Equal(t, "0", z.Text(0))
		assert.Equal(t, "1/1", z.Scala(1))
		assert.Equal(t, "1/1", z.Scala(5))
		assert.Equal(t, "0.000", z.Decimal(3, 0))

		text, err := z.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "0", string(text))
	})

	t.Run("comparison", func(t *testing.T) {
		t.Parallel()

		for _, d := range setS(t) {
			assert.Equal(t, -d.c.sign(), z.Compare(d.v), d.name)

			near, err := z.Near(d.v, cents.Zero())
			require.NoError(t, err)
			assert.Equal(t, d.c.sign() == 0, near, d.name)
		}

		near, err := z.Near(z, z)
		require.NoError(t, err)
		assert.True(t, near)
	})
}
