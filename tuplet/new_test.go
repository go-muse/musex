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

// TestNew covers K-1, G-1, and G-2: creation keeps the pair as given.
func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		actual, normal int
	}{
		{name: "triplet", actual: 3, normal: 2},
		{name: "duplet", actual: 2, normal: 3},
		{name: "sextuplet is not reduced", actual: 6, normal: 4},
		{name: "one to one", actual: 1, normal: 1},
		{name: "equal counts", actual: 4, normal: 4},
		{name: "actual one", actual: 1, normal: 5},
		{name: "normal one", actual: 7, normal: 1},
		{name: "max actual", actual: math.MaxInt, normal: 1},
		{name: "max normal", actual: 1, normal: math.MaxInt},
		{name: "max both", actual: math.MaxInt, normal: math.MaxInt},
		{name: "near max", actual: math.MaxInt - 1, normal: math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.New(tt.actual, tt.normal)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)
			assert.Equal(t, tuplet.MustNew(tt.actual, tt.normal), got)
		})
	}
}

// TestNew_InvalidCount covers E-1: a count below 1 in either argument.
func TestNew_InvalidCount(t *testing.T) {
	t.Parallel()

	for _, bad := range []int{0, -1, math.MinInt} {
		tests := []struct {
			name           string
			actual, normal int
		}{
			{name: "actual", actual: bad, normal: 2},
			{name: "normal", actual: 3, normal: bad},
			{name: "both", actual: bad, normal: bad},
		}

		for _, tt := range tests {
			t.Run(tt.name+"/"+itoa(bad), func(t *testing.T) {
				t.Parallel()

				got, err := tuplet.New(tt.actual, tt.normal)
				assertErr(t, err, tuplet.ErrInvalidCount)
				assert.Equal(t, tuplet.Tuplet{}, got)
			})
		}
	}
}

// coefficientInput is a constructor of a coefficient used by table tests.
type coefficientInput func(t *testing.T) fraction.Fraction

func fixed(f fraction.Fraction) coefficientInput {
	return func(*testing.T) fraction.Fraction { return f }
}

func bigInput(num, den *big.Int) coefficientInput {
	return func(t *testing.T) fraction.Fraction { return bigFrac(t, num, den) }
}

// TestFromCoefficient covers KP-16 and architecture 4.4: the result is
// |q|:|p| for the source pair (p, q).
func TestFromCoefficient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		c              fraction.Fraction
		actual, normal int
	}{
		{name: "two thirds", c: frac(2, 3), actual: 3, normal: 2},
		{name: "four sixths keeps the pair", c: frac(4, 6), actual: 6, normal: 4},
		{name: "three halves", c: frac(3, 2), actual: 2, normal: 3},
		{name: "negative components", c: frac(-2, -3), actual: 3, normal: 2},
		{name: "one", c: frac(1, 1), actual: 1, normal: 1},
		{name: "six sixths", c: frac(6, 6), actual: 6, normal: 6},
		{name: "max numerator", c: frac(math.MaxInt, 1), actual: 1, normal: math.MaxInt},
		{name: "max denominator", c: frac(1, math.MaxInt), actual: math.MaxInt, normal: 1},
		{
			name:   "negated near max",
			c:      frac(-math.MaxInt, -(math.MaxInt - 1)),
			actual: math.MaxInt - 1, normal: math.MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.FromCoefficient(tt.c)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)
		})
	}
}

// TestFromReciprocalCoefficient covers KP-16: the result is |p|:|q|.
func TestFromReciprocalCoefficient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		c              fraction.Fraction
		actual, normal int
	}{
		{name: "three halves", c: frac(3, 2), actual: 3, normal: 2},
		{name: "six fourths keeps the pair", c: frac(6, 4), actual: 6, normal: 4},
		{name: "two thirds", c: frac(2, 3), actual: 2, normal: 3},
		{name: "negative components", c: frac(-3, -2), actual: 3, normal: 2},
		{name: "one", c: frac(1, 1), actual: 1, normal: 1},
		{name: "max numerator", c: frac(math.MaxInt, 1), actual: math.MaxInt, normal: 1},
		{name: "max denominator", c: frac(1, math.MaxInt), actual: 1, normal: math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.FromReciprocalCoefficient(tt.c)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)
		})
	}
}

// TestFromCoefficient_Errors covers KP-16, E-1, E-5, E-6 and the check order
// E-6 -> E-1 -> E-5 of architecture 4.2 for both coefficient constructors.
func TestFromCoefficient_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    coefficientInput
		want error
	}{
		{name: "unset", c: fixed(fraction.Fraction{}), want: tuplet.ErrInvalidArgument},
		{name: "zero", c: fixed(frac(0, 1)), want: tuplet.ErrInvalidCount},
		{name: "zero with denominator five", c: fixed(frac(0, 5)), want: tuplet.ErrInvalidCount},
		{name: "zero with negative denominator", c: fixed(frac(0, -5)), want: tuplet.ErrInvalidCount},
		{name: "negative numerator", c: fixed(frac(-2, 3)), want: tuplet.ErrInvalidCount},
		{name: "negative denominator", c: fixed(frac(2, -3)), want: tuplet.ErrInvalidCount},
		{name: "negative min int numerator before range", c: fixed(frac(math.MinInt, 1)), want: tuplet.ErrInvalidCount},
		{name: "negative big numerator before range", c: bigInput(new(big.Int).Neg(pow2(64)), big.NewInt(3)), want: tuplet.ErrInvalidCount},
		{name: "min int numerator", c: fixed(frac(math.MinInt, -1)), want: tuplet.ErrOutOfRange},
		{name: "min int denominator", c: fixed(frac(-1, math.MinInt)), want: tuplet.ErrOutOfRange},
		{name: "numerator beyond int", c: bigInput(pow2(64), big.NewInt(3)), want: tuplet.ErrOutOfRange},
		{name: "denominator beyond int", c: bigInput(big.NewInt(3), pow2(64)), want: tuplet.ErrOutOfRange},
		{name: "reducible pair beyond int", c: bigInput(pow2(64), pow2(65)), want: tuplet.ErrOutOfRange},
		{name: "max component", c: maxComponentFrac, want: tuplet.ErrOutOfRange},
	}

	constructors := []struct {
		name string
		f    func(fraction.Fraction) (tuplet.Tuplet, error)
	}{
		{name: "FromCoefficient", f: tuplet.FromCoefficient},
		{name: "FromReciprocalCoefficient", f: tuplet.FromReciprocalCoefficient},
	}

	for _, ctor := range constructors {
		for _, tt := range tests {
			t.Run(ctor.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := ctor.f(tt.c(t))
				assertErr(t, err, tt.want)
				assert.Equal(t, tuplet.Tuplet{}, got)
			})
		}
	}
}

// TestFromCoefficient_RoundTrip covers K-1: creation from the coefficient and
// from the reciprocal coefficient of any set tuplet returns the same tuplet.
func TestFromCoefficient_RoundTrip(t *testing.T) {
	t.Parallel()

	for _, tp := range sampleTuplets() {
		t.Run(tp.String(), func(t *testing.T) {
			t.Parallel()

			fromCoefficient, err := tuplet.FromCoefficient(tp.MustCoefficient())
			require.NoError(t, err)
			assert.Equal(t, tp, fromCoefficient)

			fromReciprocal, err := tuplet.FromReciprocalCoefficient(tp.MustReciprocalCoefficient())
			require.NoError(t, err)
			assert.Equal(t, tp, fromReciprocal)
		})
	}
}

// TestFromUnitRatio covers KP-14 and architecture 4.4: actual:(normal*ratio).
func TestFromUnitRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		actual, normal int
		ratio          fraction.Fraction
		wantNormal     int
	}{
		{name: "three eighths in the time of one quarter", actual: 3, normal: 1, ratio: frac(2, 1), wantNormal: 2},
		{name: "same unit", actual: 3, normal: 2, ratio: frac(1, 1), wantNormal: 2},
		{name: "unreduced ratio", actual: 5, normal: 2, ratio: frac(2, 4), wantNormal: 1},
		{name: "fractional ratio with integer product", actual: 3, normal: 2, ratio: frac(3, 2), wantNormal: 3},
		{name: "half ratio with even normal", actual: 3, normal: 4, ratio: frac(1, 2), wantNormal: 2},
		{name: "negative components with positive value", actual: 3, normal: 1, ratio: frac(-2, -1), wantNormal: 2},
		{name: "max normal", actual: 1, normal: math.MaxInt, ratio: frac(1, 1), wantNormal: math.MaxInt},
		{name: "product reaches max", actual: 7, normal: 1, ratio: frac(math.MaxInt, 1), wantNormal: math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.FromUnitRatio(tt.actual, tt.normal, tt.ratio)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.wantNormal)
		})
	}
}

// TestFromUnitRatio_Errors covers KP-14, K-10 and the check order
// E-1 -> E-6 -> E-7 -> E-5 of architecture 4.2.
func TestFromUnitRatio_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		actual, normal int
		ratio          fraction.Fraction
		want           error
		causes         []error
	}{
		{name: "three in three halves", actual: 3, normal: 3, ratio: frac(1, 2), want: tuplet.ErrNotRepresentable},
		{name: "three in one half", actual: 3, normal: 1, ratio: frac(1, 2), want: tuplet.ErrNotRepresentable},
		{name: "zero actual", actual: 0, normal: 1, ratio: frac(2, 1), want: tuplet.ErrInvalidCount},
		{name: "negative actual", actual: -1, normal: 1, ratio: frac(2, 1), want: tuplet.ErrInvalidCount},
		{name: "zero normal", actual: 3, normal: 0, ratio: frac(2, 1), want: tuplet.ErrInvalidCount},
		{name: "min int normal", actual: 3, normal: math.MinInt, ratio: frac(2, 1), want: tuplet.ErrInvalidCount},
		{name: "unset ratio", actual: 3, normal: 1, ratio: fraction.Fraction{}, want: tuplet.ErrInvalidArgument},
		{name: "zero ratio", actual: 3, normal: 1, ratio: frac(0, 1), want: tuplet.ErrInvalidArgument},
		{name: "negative ratio", actual: 3, normal: 1, ratio: frac(-1, 2), want: tuplet.ErrInvalidArgument},
		{name: "negative ratio denominator", actual: 3, normal: 1, ratio: frac(1, -2), want: tuplet.ErrInvalidArgument},
		{
			name: "integer product beyond int", actual: 1, normal: math.MaxInt, ratio: frac(2, 1),
			want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrNotRepresentable},
		},
		{name: "count before unset ratio", actual: 0, normal: 1, ratio: fraction.Fraction{}, want: tuplet.ErrInvalidCount},
		{name: "count before negative ratio", actual: 3, normal: 0, ratio: frac(-1, 1), want: tuplet.ErrInvalidCount},
		{name: "ratio before representability", actual: 3, normal: 3, ratio: frac(-1, 2), want: tuplet.ErrInvalidArgument},
		{name: "representability before range", actual: 1, normal: math.MaxInt, ratio: frac(3, 2), want: tuplet.ErrNotRepresentable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.FromUnitRatio(tt.actual, tt.normal, tt.ratio)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}
}
