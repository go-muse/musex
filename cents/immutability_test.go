package cents_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allOperations calls every operation with a and b (and fixed other
// arguments), ignoring the results; Must variants are called only where
// they do not panic for S (they are covered by TestMustVariants).
func allOperations(a, b cents.Cents) {
	_, _ = a.Add(b)
	_, _ = a.Sub(b)
	_ = a.Neg()
	_ = a.Abs()
	_ = a.Sign()
	_, _ = a.Mul(3)
	_, _ = a.Mul(maxInt)
	_, _ = a.Div(7)
	_, _ = a.Div(0)
	_, _ = a.Scale(5, 3)
	_, _ = a.ScaleRat(big.NewRat(-2, 9))
	_, _ = a.RoundTo(1000)
	_, _ = a.SplitOctaves()
	_, _, _ = a.NearestStep(53)
	_, _, _ = a.NearestStep(maxInt)
	_ = a.Equal(b)
	_ = a.Compare(b)
	_, _ = a.Near(b, cents.Cent())
	_, _ = a.Near(b, b.Neg())
	_ = a.Float64()
	_, _, _ = a.Fraction()
	_, _, _ = a.Ratio()
	_, _, _ = a.Composition()
	_, _ = a.Quo(b)
	_, _ = a.Multiplier()
	_, _ = a.ApplyTo(440)
	_ = a.Kind()
	_ = a.String()
	_ = a.Text(cents.Typographic | cents.WithUnit | cents.WithPlus)
	_, _ = a.TextAs(cents.KindPower, cents.WithPlus)
	_ = a.Decimal(9, cents.WithUnit)
	_ = a.Scala(5)
	_, _ = a.MarshalText()
	_, _ = a.AppendText(make([]byte, 1, 64))
	_ = cents.Cents.Compare(a, b)
}

// K-33, NF-6, G-11: for all a, b of S the operations of NF-6 return no
// error and do not panic; after any operation, also with an error, the
// inputs equal their copies made before the call.
func TestInputsUnchanged(t *testing.T) {
	t.Parallel()

	set := setS(t)
	for _, a := range set {
		for _, b := range set {
			ca, cb := a.v, b.v

			require.NotPanics(t, func() { allOperations(a.v, b.v) }, "%s, %s", a.name, b.name)
			assert.True(t, ca == a.v, "%s unchanged", a.name)
			assert.True(t, cb == b.v, "%s unchanged", b.name)
			assertComp(t, a.c, a.v, "%s unchanged", a.name)

			require.NotPanics(t, func() {
				_ = a.v.Neg()
				_ = a.v.Abs()
				_ = a.v.Sign()
				_ = a.v.Equal(b.v)
				_ = a.v.Compare(b.v)
				_, _ = a.v.SplitOctaves()
				_ = a.v.Float64()
				_, _, _ = a.v.Fraction()
				_, _, _ = a.v.Ratio()
				_, _, _ = a.v.Composition()
				_ = a.v.String()
				_ = a.v.Decimal(6, 0)
			})

			_, err := a.v.Near(b.v, cents.Cent())
			assert.NoError(t, err, "Near %s, %s", a.name, b.name)

			if b.c.sign() != 0 {
				_, err = a.v.Quo(b.v)
				assert.NoError(t, err, "Quo %s, %s", a.name, b.name)
			}

			for _, n := range []int{1, 12, 53, 1200} {
				_, _, err = a.v.NearestStep(n)
				assert.NoError(t, err, "NearestStep %s, %d", a.name, n)
			}
		}

		for _, k := range kinds() {
			_, _ = a.v.TextAs(k, 0)
		}
	}
}

// G-11: the value of a panicking Must variant leaves its inputs unchanged.
func TestInputsUnchangedAfterPanic(t *testing.T) {
	t.Parallel()

	a := noErr(cents.FromInt(9000000))(t)
	b := cents.Cent()
	ca, cb := a, b

	assertPanicsWithErrorIs(t, cents.ErrOutOfRange, func() { a.MustAdd(b) })
	assertPanicsWithErrorIs(t, cents.ErrInvalidArgument, func() { a.MustDiv(0) })
	assertPanicsWithErrorIs(t, cents.ErrOutOfRange, func() { a.MustNearestStep(maxInt) })
	assert.True(t, ca == a)
	assert.True(t, cb == b)
	assertComp(t, ratC(9000000, 1), a)
}

// K-33, G-11: *big.Int, *big.Rat and []Power inputs are not changed or kept.
func TestBigInputs(t *testing.T) {
	t.Parallel()

	t.Run("FromComposition", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name string
			num  *big.Int
			den  *big.Int
			ps   []cents.Power
			ok   bool
		}{
			{"success", big.NewInt(-2400), big.NewInt(2), []cents.Power{pw(3, 2, 4), pw(5, -3, 1), pw(3, 1, 2)}, true},
			{"error", big.NewInt(9000000), big.NewInt(-1), []cents.Power{pw(3, 6000, 1), pw(37, 1, 1)}, false},
		} {
			num, den := new(big.Int).Set(tc.num), new(big.Int).Set(tc.den)
			ps := make([]cents.Power, len(tc.ps))
			copy(ps, tc.ps)
			nums := make([]*big.Int, len(ps))
			dens := make([]*big.Int, len(ps))
			for i, p := range ps {
				nums[i], dens[i] = new(big.Int).Set(p.Num), new(big.Int).Set(p.Den)
			}

			c, err := cents.FromComposition(tc.num, tc.den, tc.ps...)
			assert.Equal(t, tc.ok, err == nil, tc.name)
			assert.Equal(t, 0, num.Cmp(tc.num), tc.name)
			assert.Equal(t, 0, den.Cmp(tc.den), tc.name)
			require.Len(t, tc.ps, len(ps))

			for i := range ps {
				assert.Equal(t, ps[i].Base, tc.ps[i].Base, tc.name)
				assert.Same(t, ps[i].Num, tc.ps[i].Num, tc.name)
				assert.Equal(t, 0, nums[i].Cmp(tc.ps[i].Num), tc.name)
				assert.Equal(t, 0, dens[i].Cmp(tc.ps[i].Den), tc.name)
			}

			if !tc.ok {
				continue
			}

			before := c
			tc.num.SetInt64(777)
			tc.den.SetInt64(5)
			tc.ps[0].Num.SetInt64(9)
			tc.ps[0].Den.SetInt64(1)
			tc.ps[0].Base = 7

			assert.True(t, before == c, "changing the inputs later does not change the value")
			// (-2400/2; 3 → 2/4 + 1/2, 5 → -3) = (-1200; 3 → 1, 5 → -3).
			assertComp(t, ratC(-1200, 1).withExp(3, rat(1, 1)).withExp(5, rat(-3, 1)), c)
		}
	})

	t.Run("ScaleRat", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name string
			c    cents.Cents
			r    *big.Rat
			ok   bool
		}{
			{"success", cents.Millioctave(), big.NewRat(1750, 3), true},
			{"error", cents.Millioctave(), new(big.Rat).SetFrac(big.NewInt(1), pow10(30)), false},
		} {
			before := new(big.Rat).Set(tc.r)
			got, err := tc.c.ScaleRat(tc.r)
			assert.Equal(t, tc.ok, err == nil, tc.name)
			assert.Equal(t, 0, before.Cmp(tc.r), "the input rational is unchanged: %s", tc.name)

			if tc.ok {
				kept := got
				tc.r.SetInt64(3)
				assert.True(t, kept == got)
				assertComp(t, ratC(700, 1), got)
			}
		}
	})
}

// K-33, G-11: outputs belong to the caller; changing them changes neither
// the value nor later calls.
func TestOutputsOwned(t *testing.T) {
	t.Parallel()

	t.Run("Composition", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			num, den, ps := a.v.Composition()
			num.SetInt64(31)
			den.SetInt64(7)

			for i := range ps {
				ps[i].Num.SetInt64(11)
				ps[i].Den.SetInt64(13)
				ps[i].Base = 29
			}

			assertComp(t, a.c, a.v, a.name)
		}
	})

	t.Run("Quotient.Exact", func(t *testing.T) {
		t.Parallel()

		q := noErr(cents.FromInt(700))(t).MustQuo(cents.Millioctave())
		e, ok := q.Exact()
		require.True(t, ok)
		e.SetInt64(1)

		e2, ok := q.Exact()
		require.True(t, ok)
		assert.Equal(t, "1750/3", e2.RatString())
		assert.NotSame(t, e, e2)
	})

	t.Run("text bytes", func(t *testing.T) {
		t.Parallel()

		c := noErr(cents.FromRatio(3, 2))(t)
		text, err := c.MarshalText()
		require.NoError(t, err)
		text[0] = '9'

		again, err := c.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "3:2", string(again))

		// AppendText(nil) returns a new array; with enough spare capacity
		// the result shares the array of the buffer (KR-C16), so only the
		// value and the following calls are independent of the result.
		out, err := c.AppendText(nil)
		require.NoError(t, err)
		require.NotEmpty(t, out)
		out[0] = '7'
		assert.Equal(t, "3:2", c.String())

		out, err = c.AppendText(make([]byte, 0, 16))
		require.NoError(t, err)
		require.NotEmpty(t, out)
		out[0] = '7'
		assert.Equal(t, "3:2", c.String())

		again, err = c.AppendText(nil)
		require.NoError(t, err)
		assert.Equal(t, "3:2", string(again))
	})

	t.Run("UnmarshalText input", func(t *testing.T) {
		t.Parallel()

		var c cents.Cents
		in := []byte("3:2")
		require.NoError(t, c.UnmarshalText(in))
		in[0] = '5'
		assertCents(t, noErr(cents.FromRatio(3, 2))(t), c, "the input is not kept")
	})
}

// K-34 (test part), NF-2: the named values and units cannot be changed for
// other parts of the program through copies or the lists.
func TestNamedValuesImmutable(t *testing.T) {
	t.Parallel()

	specs := append(namedSpecs(), unitSpecs()...)

	for _, s := range specs {
		x := s.get()
		x = x.MustAdd(cents.Cent())
		x = x.MustMul(3)
		x = x.Neg()
		require.NoError(t, x.UnmarshalText([]byte("3:2")))

		assertComp(t, s.def, s.get(), "%s after changing copies", s.name)
	}

	values, units := cents.NamedValues(), cents.NamedUnits()
	for i := range values {
		values[i].Value = cents.Octave()
		values[i].Name = "changed"
	}

	for i := range units {
		units[i] = cents.NamedValue{Name: "changed", Value: cents.Zero()}
	}

	values = append(values[:0], cents.NamedValue{Name: "appended"})
	assert.Len(t, values, 1)

	for i, s := range namedSpecs() {
		assertComp(t, s.def, s.get(), s.name)
		nv := cents.NamedValues()[i]
		assert.Equal(t, s.name, nv.Name)
		assertComp(t, s.def, nv.Value, s.name)
	}

	for i, s := range unitSpecs() {
		assertComp(t, s.def, s.get(), s.name)
		nv := cents.NamedUnits()[i]
		assert.Equal(t, s.name, nv.Name)
		assertComp(t, s.def, nv.Value, s.name)
	}

	assert.Len(t, cents.NamedValues(), 19)
	assert.Len(t, cents.NamedUnits(), 8)
	assertComp(t, ratC(1200, 1), cents.Octave())
	assertComp(t, zeroComp(), cents.Zero())
	assertComp(t, ratC(1, 1), cents.Cent())
}
