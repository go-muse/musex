package tuplet_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApply_KP2 covers KP-2: an eighth in a triplet is 1/12, an eighth in a
// duplet is 3/16 (a dotted eighth).
func TestApply_KP2(t *testing.T) {
	t.Parallel()

	eighth := frac(1, 8)

	got, err := tuplet.Triplet().Apply(eighth)
	require.NoError(t, err)
	assertPair(t, got, 1, 12)

	got, err = tuplet.Duplet().Apply(eighth)
	require.NoError(t, err)
	assertPair(t, got, 3, 16)
	assert.True(t, got.Equal(frac(1, 8).MustAdd(frac(1, 16))))
}

// TestSpan_KP3 covers KP-3 and IP-5: the total duration of a tuplet written
// with a unit is normal*unit, fully reduced.
func TestSpan_KP3(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tp       tuplet.Tuplet
		unit     fraction.Fraction
		num, den int
	}{
		{name: "5:4 eighths", tp: tuplet.MustNew(5, 4), unit: frac(1, 8), num: 1, den: 2},
		{name: "5:3 eighths", tp: tuplet.MustNew(5, 3), unit: frac(1, 8), num: 3, den: 8},
		{name: "7:4 quarters", tp: tuplet.MustNew(7, 4), unit: frac(1, 4), num: 1, den: 1},
		{name: "7:8 eighths", tp: tuplet.MustNew(7, 8), unit: frac(1, 8), num: 1, den: 1},
		{name: "3:2 quarters", tp: tuplet.MustNew(3, 2), unit: frac(1, 4), num: 1, den: 2},
		{name: "14:8 sixty-fourths", tp: tuplet.MustNew(14, 8), unit: frac(1, 64), num: 1, den: 8},
		{name: "3:2 halves", tp: tuplet.MustNew(3, 2), unit: frac(1, 2), num: 1, den: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.tp.Span(tt.unit)
			require.NoError(t, err)
			assertPair(t, got, tt.num, tt.den)
			assert.Equal(t, got, tt.tp.MustSpan(tt.unit))
		})
	}
}

// applyDurations returns durations for property checks: zero, negative, a
// negative denominator, unreduced pairs, and large components.
func applyDurations(t *testing.T) []fraction.Fraction {
	t.Helper()

	return []fraction.Fraction{
		frac(1, 8),
		frac(2, 16),
		frac(3, 16),
		frac(1, 1),
		frac(0, 1),
		frac(0, 7),
		frac(-1, 8),
		frac(1, -8),
		frac(-3, -4),
		frac(7, 3),
		frac(math.MaxInt, 1),
		frac(1, math.MaxInt),
		frac(math.MinInt, 3),
		bigFrac(t, pow2(200), pow2(100)),
	}
}

// TestApply_Oracle covers IP-4, O-11, G-5 and architecture 4.9: Apply,
// Unapply and Span give the exact value in fully reduced form.
func TestApply_Oracle(t *testing.T) {
	t.Parallel()

	tuplets := []tuplet.Tuplet{
		tuplet.Triplet(), tuplet.Duplet(), tuplet.Sextuplet(), tuplet.MustNew(1, 1),
		tuplet.MustNew(6, 6), tuplet.MustNew(7, 11), tuplet.MustNew(math.MaxInt, 1),
		tuplet.MustNew(1, math.MaxInt), tuplet.MustNew(math.MaxInt, math.MaxInt-1),
	}

	durations := applyDurations(t)

	for _, tp := range tuplets {
		t.Run(tp.String(), func(t *testing.T) {
			t.Parallel()

			actual, normal := tp.MustPair()
			coefficient := ratInts(normal, actual)
			reciprocal := ratInts(actual, normal)

			for _, d := range durations {
				dv := ratOf(t, d)

				applied, err := tp.Apply(d)
				if assert.NoError(t, err, "Apply(%s)", d) {
					assertReducedValue(t, applied, new(big.Rat).Mul(dv, coefficient))
				}

				unapplied, err := tp.Unapply(d)
				if assert.NoError(t, err, "Unapply(%s)", d) {
					assertReducedValue(t, unapplied, new(big.Rat).Mul(dv, reciprocal))
				}

				span, err := tp.Span(d)
				if assert.NoError(t, err, "Span(%s)", d) {
					assertReducedValue(t, span, new(big.Rat).Mul(dv, ratInts(normal, 1)))
				}
			}
		})
	}
}

// TestApply_RoundTrip covers K-8 and architecture 4.9: Unapply(Apply(d)) and
// Apply(Unapply(d)) equal d.Reduce() as a pair.
func TestApply_RoundTrip(t *testing.T) {
	t.Parallel()

	durations := applyDurations(t)

	for _, tp := range sampleTuplets() {
		t.Run(tp.String(), func(t *testing.T) {
			t.Parallel()

			for _, d := range durations {
				if applied, err := tp.Apply(d); err == nil {
					back, err := tp.Unapply(applied)
					if assert.NoError(t, err, "Unapply(Apply(%s))", d) {
						assert.Equal(t, d.Reduce(), back, "Unapply(Apply(%s))", d)
					}
				}

				if unapplied, err := tp.Unapply(d); err == nil {
					back, err := tp.Apply(unapplied)
					if assert.NoError(t, err, "Apply(Unapply(%s))", d) {
						assert.Equal(t, d.Reduce(), back, "Apply(Unapply(%s))", d)
					}
				}
			}
		})
	}
}

// TestApply_ZeroAndNegative covers G-5 explicitly.
func TestApply_ZeroAndNegative(t *testing.T) {
	t.Parallel()

	tp := tuplet.Triplet()

	got, err := tp.Apply(frac(0, 5))
	require.NoError(t, err)
	assertPair(t, got, 0, 1)

	got, err = tp.Apply(frac(-1, 8))
	require.NoError(t, err)
	assertPair(t, got, -1, 12)

	got, err = tp.Apply(frac(1, -8))
	require.NoError(t, err)
	assertPair(t, got, -1, 12)

	got, err = tp.Unapply(frac(-1, 12))
	require.NoError(t, err)
	assertPair(t, got, -1, 8)

	got, err = tp.Span(frac(-1, 8))
	require.NoError(t, err)
	assertPair(t, got, -1, 4)

	got, err = tp.Span(frac(0, 3))
	require.NoError(t, err)
	assertPair(t, got, 0, 1)
}

// TestApply_Errors covers E-2, E-5, E-6 and the order E-2 -> E-6 -> E-5 for
// Apply, Unapply and Span; E-5 also matches fraction.ErrOutOfRange (K-13).
func TestApply_Errors(t *testing.T) {
	t.Parallel()

	maxFrac := maxComponentFrac(t)

	ops := []struct {
		name string
		f    func(tuplet.Tuplet, fraction.Fraction) (fraction.Fraction, error)
		// overflow is a tuplet whose operation on MaxComponent/1 is out of range.
		overflow tuplet.Tuplet
	}{
		{name: "Apply", f: tuplet.Tuplet.Apply, overflow: tuplet.Duplet()},
		{name: "Unapply", f: tuplet.Tuplet.Unapply, overflow: tuplet.Triplet()},
		{name: "Span", f: tuplet.Tuplet.Span, overflow: tuplet.Triplet()},
	}

	for _, op := range ops {
		tests := []struct {
			name   string
			tp     tuplet.Tuplet
			d      fraction.Fraction
			want   error
			causes []error
		}{
			{name: "unset duration", tp: tuplet.Triplet(), d: fraction.Fraction{}, want: tuplet.ErrInvalidArgument},
			{name: "unset tuplet", tp: tuplet.Tuplet{}, d: frac(1, 8), want: tuplet.ErrUnset},
			{name: "unset tuplet before unset duration", tp: tuplet.Tuplet{}, d: fraction.Fraction{}, want: tuplet.ErrUnset},
			{name: "unset tuplet before range", tp: tuplet.Tuplet{}, d: maxFrac, want: tuplet.ErrUnset},
			{
				name: "out of range", tp: op.overflow, d: maxFrac,
				want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrOutOfRange},
			},
		}

		for _, tt := range tests {
			t.Run(op.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := op.f(tt.tp, tt.d)
				assertErr(t, err, tt.want, tt.causes...)
				assert.Equal(t, fraction.Fraction{}, got)
			})
		}
	}
}

// TestApply_InRange covers architecture 4.9: a neutral tuplet keeps a value
// at the edge of the common/fraction range.
func TestApply_InRange(t *testing.T) {
	t.Parallel()

	maxFrac := maxComponentFrac(t)

	got, err := tuplet.MustNew(1, 1).Apply(maxFrac)
	require.NoError(t, err)
	assertBigPair(t, got, fraction.MaxComponentBig(), big.NewInt(1))

	got, err = tuplet.MustNew(3, 3).Unapply(maxFrac)
	require.NoError(t, err)
	assertBigPair(t, got, fraction.MaxComponentBig(), big.NewInt(1))
}
