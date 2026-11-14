package interval_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-2: octaves, quality steps, inversion, sum, difference and multiple
// (OP-11...OP-17; P-6, P-8, P-9; KP-5...KP-7; architecture 4.6).

// q7 is MaxInt / 7: the largest number of octaves that can be added to a
// unison.
const q7 = maxInt / 7

// ---------------------------------------------------------------------------
// Octaves (OP-11, P-6; KP-5).

func TestArithmeticAddOctaves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    interval.Interval
		k    int
		want interval.Interval
	}{
		{"M3 by 2 (KP-5)", iv("M3"), 2, iv("M17")},
		{"P1 by 1 (P-6)", iv("P1"), 1, iv("P8")},
		{"A1 by 1 (P-6)", iv("A1"), 1, iv("A8")},
		{"descending A1 by 1", iv("↓A1"), 1, iv("↓A8")},
		{"descending M3 by 1", iv("↓M3"), 1, iv("↓M10")},
		{"P8 by 1", iv("P8"), 1, iv("P15")},
		{"d8 by 2", iv("d8"), 2, iv("d22")},
		{"n3 by 1", iv("n3"), 1, iv("n10")},
		{"by 0", iv("↓m6"), 0, iv("↓m6")},
		{"M7 up to MaxSize", iv("M7"), (maxInt - 7) / 7, basicSize(interval.MaxSize)},
		{"P1 by MaxInt/7", iv("P1"), q7 - 1, interval.MustFromDeviation(maxInt-6, ai(0))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.i.AddOctaves(tt.k)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v", got)
		})
	}
}

func TestArithmeticAddOctavesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    interval.Interval
		k    int
		want error
	}{
		{"unset", interval.Interval{}, 1, interval.ErrUnset},
		{"unset before a negative number", interval.Interval{}, -1, interval.ErrUnset},
		{"negative", iv("M3"), -1, interval.ErrInvalidArgument},
		{"MinInt", iv("P8"), minInt, interval.ErrInvalidArgument},
		{"beyond MaxSize by one octave", basicSize(interval.MaxSize), 1, interval.ErrOutOfRange},
		{"MaxInt octaves", iv("P1"), maxInt, interval.ErrOutOfRange},
		{"MaxInt/7 octaves of a second", iv("M2"), q7, interval.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.i.AddOctaves(tt.k)
			assertErr(t, err, tt.want)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Quality steps (OP-12; architecture 4.6).

func TestArithmeticExpandContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i, expand, contract string
	}{
		{"M3", "A3", "m3"},
		{"m3", "M3", "d3"},
		{"d3", "m3", "dd3"},
		{"P5", "A5", "d5"},
		{"d5", "P5", "dd5"},
		{"P4", "A4", "d4"},
		{"A1", "AA1", "P1"},
		{"↓A1", "↓AA1", "P1"},
		{"AA1", "AAA1", "A1"},
		{"P8", "A8", "d8"},
		{"n3", "{+1/2}3", "{-3/2}3"},
		{"↓M3", "↓A3", "↓m3"},
		{"↓P8", "↓A8", "↓d8"},
		{"M10", "A10", "m10"},
		{"{+1/2}1", "{+3/2}1", ""},
		{"P1", "A1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			t.Parallel()

			i := iv(tt.i)

			e, err := i.Expand()
			require.NoError(t, err)
			assert.True(t, e.Equal(iv(tt.expand)), "Expand = %v", e)

			eb, err := i.ExpandBy(1)
			require.NoError(t, err)
			assert.True(t, eb.Equal(e), "ExpandBy(1)")

			ea, err := i.ExpandByAlteration(ai(1))
			require.NoError(t, err)
			assert.True(t, ea.Equal(e), "ExpandByAlteration(+1)")

			cb, err := i.ContractBy(-1)
			require.NoError(t, err)
			assert.True(t, cb.Equal(e), "ContractBy(-1)")

			c, err := i.Contract()
			if tt.contract == "" {
				assertErr(t, err, interval.ErrInvalidArgument)
				assert.Equal(t, interval.Interval{}, c)

				return
			}

			require.NoError(t, err)
			assert.True(t, c.Equal(iv(tt.contract)), "Contract = %v", c)
			assert.True(t, i.MustContractBy(1).Equal(c), "ContractBy(1)")
			assert.True(t, i.MustContractByAlteration(ai(1)).Equal(c), "ContractByAlteration(+1)")
			assert.True(t, i.MustExpandBy(-1).Equal(c), "ExpandBy(-1)")
		})
	}
}

func TestArithmeticExpandBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  func() (interval.Interval, error)
		want interval.Interval
	}{
		{"M3 by 2", func() (interval.Interval, error) { return iv("M3").ExpandBy(2) }, iv("AA3")},
		{"M3 by 0", func() (interval.Interval, error) { return iv("M3").ExpandBy(0) }, iv("M3")},
		{"M3 by -3", func() (interval.Interval, error) { return iv("M3").ExpandBy(-3) }, iv("dd3")},
		{"contract M3 by -3", func() (interval.Interval, error) { return iv("M3").ContractBy(-3) }, iv("AAA3")},
		{"contract P5 by 2", func() (interval.Interval, error) { return iv("P5").ContractBy(2) }, iv("dd5")},
		{"m3 by +1/2", func() (interval.Interval, error) { return iv("m3").ExpandByAlteration(alt(1, 2)) }, iv("n3")},
		{"M3 by -1/2", func() (interval.Interval, error) { return iv("M3").ExpandByAlteration(alt(-1, 2)) }, iv("n3")},
		{"contract M3 by 1/2", func() (interval.Interval, error) { return iv("M3").ContractByAlteration(alt(1, 2)) }, iv("n3")},
		{"contract n3 by -1/2", func() (interval.Interval, error) { return iv("n3").ContractByAlteration(alt(-1, 2)) }, iv("M3")},
		{"by the natural value", func() (interval.Interval, error) { return iv("↓n6").ExpandByAlteration(alteration.Natural()) }, iv("↓n6")},
		{"A1 by -1/2", func() (interval.Interval, error) { return iv("A1").ExpandByAlteration(alt(-1, 2)) }, iv("{+1/2}1")},
		{"descending A1 contracted to P1", func() (interval.Interval, error) { return iv("↓A1").ContractByAlteration(ai(1)) }, iv("P1")},
		{"-M by 2M", func() (interval.Interval, error) { return dim(mag, 5).ExpandBy(2 * mag) }, aug(mag, 5)},
		{"M contracted by 2M", func() (interval.Interval, error) { return aug(mag, 4).ContractBy(2 * mag) }, dim(mag, 4)},
		{"M contracted by M", func() (interval.Interval, error) { return aug(mag, 4).ContractByAlteration(ai(mag)) }, iv("P4")},
		{"-M contracted by -M", func() (interval.Interval, error) { return dim(mag, 5).ContractByAlteration(ai(-mag)) }, iv("P5")},
		{"-M expanded by M", func() (interval.Interval, error) {
			return interval.MustFromDeviation(3, ai(-mag)).ExpandByAlteration(ai(mag))
		}, iv("M3")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.got()
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v", got)
		})
	}
}

func TestArithmeticExpandErrors(t *testing.T) {
	t.Parallel()

	var u interval.Interval

	type op = func() (interval.Interval, error)

	tests := []struct {
		name   string
		op     op
		want   error
		causes []error
	}{
		{"Expand unset", u.Expand, interval.ErrUnset, nil},
		{"Contract unset", u.Contract, interval.ErrUnset, nil},
		{"ExpandBy unset", func() (interval.Interval, error) { return u.ExpandBy(minInt) }, interval.ErrUnset, nil},
		{"ContractBy unset", func() (interval.Interval, error) { return u.ContractBy(1) }, interval.ErrUnset, nil},
		{"ExpandByAlteration unset", func() (interval.Interval, error) { return u.ExpandByAlteration(ai(1)) }, interval.ErrUnset, nil},
		{"ContractByAlteration unset", func() (interval.Interval, error) { return u.ContractByAlteration(ai(1)) }, interval.ErrUnset, nil},
		{"P1 contracted", iv("P1").Contract, interval.ErrInvalidArgument, nil},
		{"P1 expanded by -1", func() (interval.Interval, error) { return iv("P1").ExpandBy(-1) }, interval.ErrInvalidArgument, nil},
		{"A1 contracted by 2", func() (interval.Interval, error) { return iv("A1").ContractBy(2) }, interval.ErrInvalidArgument, nil},
		{"descending A1 contracted by 2", func() (interval.Interval, error) { return iv("↓A1").ContractBy(2) }, interval.ErrInvalidArgument, nil},
		{"P1 contracted by 1/2", func() (interval.Interval, error) { return iv("P1").ContractByAlteration(alt(1, 2)) }, interval.ErrInvalidArgument, nil},
		{"unison +1/2 contracted", iv("{+1/2}1").Contract, interval.ErrInvalidArgument, nil},
		{"P1 expanded by MinInt: E-2 before E-12", func() (interval.Interval, error) { return iv("P1").ExpandBy(minInt) }, interval.ErrInvalidArgument, nil},
		{"P1 contracted by MaxInt: E-2 before E-12", func() (interval.Interval, error) { return iv("P1").ContractBy(maxInt) }, interval.ErrInvalidArgument, nil},
		{"M expanded", aug(mag, 4).Expand, interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"-M contracted", dim(mag, 5).Contract, interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"M expanded by 1/2", func() (interval.Interval, error) { return aug(mag, 4).ExpandByAlteration(alt(1, 2)) },
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"M3 expanded by MaxInt", func() (interval.Interval, error) { return iv("M3").ExpandBy(maxInt) }, interval.ErrOutOfRange, nil},
		{"M3 expanded by MinInt", func() (interval.Interval, error) { return iv("M3").ExpandBy(minInt) }, interval.ErrOutOfRange, nil},
		{"M3 contracted by MinInt", func() (interval.Interval, error) { return iv("M3").ContractBy(minInt) }, interval.ErrOutOfRange, nil},
		{"M3 contracted by MaxInt", func() (interval.Interval, error) { return iv("M3").ContractBy(maxInt) }, interval.ErrOutOfRange, nil},
		{"-M contracted by M", func() (interval.Interval, error) { return dim(mag, 5).ContractByAlteration(ai(mag)) },
			interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.op()
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Inversion (OP-13, OP-14, P-8; KP-6).

func TestArithmeticInvert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i, want string
	}{
		{"M3", "m6"},
		{"A4", "d5"},
		{"P1", "P8"},
		{"AA4", "dd5"},
		{"P8", "P1"},
		{"m2", "M7"},
		{"d8", "A1"},
		{"A1", "d8"},
		{"d2", "A7"},
		{"n3", "n6"},
		{"M10", "m6"},
		{"P15", "P1"},
		{"↓M3", "↓m6"},
		{"↓P8", "P1"},
		{"↓A1", "↓d8"},
		{"AA1", "dd8"},
		{"dd8", "AA1"},
		{"{+1/2}1", "{-1/2}8"},
		{"{-1/2}8", "{+1/2}1"},
		{"{+1/4}3", "{-5/4}6"},
		{"M9", "m7"},
		{"↓M17", "↓m6"},
		{"P5", "P4"},
		{"d7", "A2"},
		{"AAA3", "ddd6"},
	}

	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			t.Parallel()

			i := iv(tt.i)
			got, err := i.Invert()
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.want)), "Invert(%s) = %v", tt.i, got)

			s, inv, err := i.InversionClass()
			require.NoError(t, err)
			assert.True(t, s.Equal(i.MustSimple()), "InversionClass simple part")
			assert.True(t, inv.Equal(got), "InversionClass inversion")
		})
	}

	t.Run("edges of the range", func(t *testing.T) {
		t.Parallel()

		assert.True(t, aug(mag, 4).MustInvert().Equal(dim(mag, 5)))
		assert.True(t, dim(mag, 5).MustInvert().Equal(aug(mag, 4)))
		assert.True(t, interval.MustFromDeviation(3, ai(mag-1)).MustInvert().Equal(interval.MustFromDeviation(6, ai(-mag))))
		assert.True(t, interval.MustFromDeviation(3, ai(-mag)).MustInvert().Equal(interval.MustFromDeviation(6, ai(mag-1))))
		assert.True(t, basicSize(interval.MaxSize).MustInvert().Equal(iv("m2")), "MaxSize has simple part M7")
	})
}

func TestArithmeticInvertErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		i      interval.Interval
		want   error
		causes []error
	}{
		{"unset", interval.Interval{}, interval.ErrUnset, nil},
		{"A8 (KP-6)", iv("A8"), interval.ErrNoInversion, nil},
		{"AA8 (KP-6)", iv("AA8"), interval.ErrNoInversion, nil},
		{"A15 (KP-6)", iv("A15"), interval.ErrNoInversion, nil},
		{"descending A8", iv("↓A8"), interval.ErrNoInversion, nil},
		{"octave +1/2", iv("{+1/2}8"), interval.ErrNoInversion, nil},
		{"octave of multiplicity M: E-4 before E-12", aug(mag, 8), interval.ErrNoInversion, nil},
		{"third with deviation M", aug(mag, 3), interval.ErrOutOfRange, nil},
		{"third with deviation M-1/2", interval.MustFromDeviation(3, alt(2*mag-1, 2)), interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.i.Invert()
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)

			s, inv, err := tt.i.InversionClass()
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, s)
			assert.Equal(t, interval.Interval{}, inv)
		})
	}
}

// TestArithmeticInvertProperties checks P-8 on the predefined intervals and
// fractional qualities.
func TestArithmeticInvertProperties(t *testing.T) {
	t.Parallel()

	set := predefinedBoth()
	set = append(set, both(ivs("n3", "n6", "{+1/2}1", "{-1/2}8", "{+1/4}3", "{-3/4}6", "{+1/2}5", "M17", "↓dd12"))...)

	for _, i := range set {
		inv, err := i.Invert()

		s := i.MustSimple()
		if s.Size() == 8 && s.Deviation().Sign() > 0 {
			assertErr(t, err, interval.ErrNoInversion)

			continue
		}

		require.NoError(t, err, "%v", i)
		assert.Equal(t, 9, s.Size()+inv.Size(), "sizes add to 9: %v", i)
		assert.True(t, inv.IsSimple(), "%v", i)

		if inv.IsPerfectUnison() {
			assert.True(t, inv.IsAscending())
		} else {
			assert.Equal(t, i.Direction(), inv.Direction(), "the direction is kept: %v", i)
		}

		assert.Equal(t, i.IsDiatonic(), inv.IsDiatonic(), "%v", i)
		assert.Equal(t, i.IsChromatic(), inv.IsChromatic(), "%v", i)
		assert.Equal(t, i.Multiplicity(), inv.Multiplicity(), "%v", i)

		// Major and minor, augmented and diminished, perfect and neutral.
		pairs := map[interval.Quality]interval.Quality{
			interval.Perfect: interval.Perfect, interval.Major: interval.Minor, interval.Minor: interval.Major,
			interval.Augmented: interval.Diminished, interval.Diminished: interval.Augmented,
			interval.Neutral: interval.Neutral, interval.Fractional: interval.Fractional,
		}
		assert.Equal(t, pairs[i.Quality()], inv.Quality(), "%v", i)

		// The inversion of the inversion of an ascending simple interval
		// with an inversion is the interval itself.
		if i.IsAscending() && i.IsSimple() {
			back, err := inv.Invert()
			if assert.NoError(t, err, "%v", i) {
				assert.True(t, back.Equal(i), "Invert twice: %v", i)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Sum, difference, multiple (OP-15...OP-17, P-9; KP-7).

func TestArithmeticAdd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b, want string
	}{
		{"M3", "m3", "P5"},
		{"M3", "M3", "A5"},
		{"P5", "P4", "P8"},
		{"A4", "d5", "P8"},
		{"M7", "m2", "P8"},
		{"P8", "P8", "P15"},
		{"M3", "↓M3", "P1"},
		{"M3", "↓P5", "↓m3"},
		{"m2", "↓A1", "d2"},
		{"P4", "↓A4", "↓A1"},
		{"d8", "A1", "P8"},
		{"n3", "n3", "P5"},
		{"P1", "↓m6", "↓m6"},
		{"A1", "↓A1", "P1"},
		{"↓A1", "↓A1", "↓AA1"},
		{"A1", "↓m2", "↓d2"},
		{"P8", "↓P8", "P1"},
		{"M9", "↓P8", "M2"},
		{"↓M10", "P8", "↓M3"},
		{"n3", "{-1/2}5", "{-1}7"},
		{"{+1/2}1", "{+1/2}1", "A1"},
	}

	for _, tt := range tests {
		t.Run(tt.a+"+"+tt.b, func(t *testing.T) {
			t.Parallel()

			a, b := iv(tt.a), iv(tt.b)
			got, err := a.Add(b)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.want)), "%s + %s = %v", tt.a, tt.b, got)
			assert.True(t, b.MustAdd(a).Equal(got), "commutative")
		})
	}

	t.Run("steps of the major scale (KP-7)", func(t *testing.T) {
		t.Parallel()

		sum := interval.PerfectUnison()
		for _, s := range ivs("M2", "M2", "m2", "M2", "M2", "M2", "m2") {
			var err error

			sum, err = sum.Add(s)
			require.NoError(t, err)
		}

		assert.True(t, sum.Equal(iv("P8")), "got %v", sum)
	})

	t.Run("edges", func(t *testing.T) {
		t.Parallel()

		big := basicSize(interval.MaxSize)
		assert.True(t, big.MustAdd(iv("P1")).Equal(big))
		assert.True(t, big.MustAdd(iv("↓M2")).Equal(basicSize(interval.MaxSize-1)), "major of MaxSize-1")
		assert.True(t, big.MustDown().MustAdd(big).Equal(iv("P1")))
		assert.True(t, aug(mag, 4).MustAdd(iv("↓A1")).Equal(aug(mag-1, 4)))
		assert.True(t, aug(mag, 4).MustAdd(dim(mag, 5)).Equal(iv("P8")))
		assert.True(t, aug(mag, 1).MustAdd(aug(mag, 1).MustDown()).Equal(iv("P1")))
	})
}

func TestArithmeticSub(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b, want string
	}{
		{"M9", "P8", "M2"},
		{"P8", "d8", "A1"},
		{"P5", "M3", "m3"},
		{"M3", "M3", "P1"},
		{"M3", "P5", "↓m3"},
		{"P1", "M3", "↓M3"},
		{"P8", "↓P8", "P15"},
		{"A1", "A1", "P1"},
		{"P1", "A1", "↓A1"},
	}

	for _, tt := range tests {
		t.Run(tt.a+"-"+tt.b, func(t *testing.T) {
			t.Parallel()

			a, b := iv(tt.a), iv(tt.b)
			got, err := a.Sub(b)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.want)), "%s - %s = %v", tt.a, tt.b, got)
		})
	}
}

func TestArithmeticMul(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    interval.Interval
		k    int
		want interval.Interval
	}{
		{"2 P5 (KP-7)", iv("P5"), 2, iv("M9")},
		{"3 M3 (KP-7)", iv("M3"), 3, iv("A7")},
		{"4 m3 (KP-7)", iv("m3"), 4, iv("d9")},
		{"0 M3 (KP-7)", iv("M3"), 0, iv("P1")},
		{"-1 P5 (KP-7)", iv("P5"), -1, iv("↓P5")},
		{"1 M3", iv("M3"), 1, iv("M3")},
		{"-2 M3", iv("M3"), -2, iv("↓A5")},
		{"2 descending M3", iv("↓M3"), 2, iv("↓A5")},
		{"-2 descending M3", iv("↓M3"), -2, iv("A5")},
		{"12 P5", iv("P5"), 12, iv("A49")},
		{"7 P8", iv("P8"), 7, iv("P50")},
		{"2 n3", iv("n3"), 2, iv("P5")},
		{"P1 MinInt times", iv("P1"), minInt, iv("P1")},
		{"P1 MaxInt times", iv("P1"), maxInt, iv("P1")},
		{"P8 MaxInt/7-1 times", iv("P8"), q7 - 1, interval.MustFromDeviation(maxInt-6, ai(0))},
		{"P8 -(MaxInt/7-1) times", iv("P8"), -(q7 - 1), interval.MustFromDeviation(maxInt-6, ai(0)).MustDown()},
		{"second -2/7 times 2^40 (architecture 7)", interval.MustFromDeviation(2, alt(-2, 7)), 1 << 40,
			interval.MustFromDeviation(1<<40+1, alt(-4, 7))},
		{"A1 M times", iv("A1"), mag, aug(mag, 1)},
		{"A1 -M times", iv("A1"), -mag, aug(mag, 1).MustDown()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.i.Mul(tt.k)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}

	t.Run("12 P5 - 7 P8 (KP-7)", func(t *testing.T) {
		t.Parallel()

		got, err := iv("P5").MustMul(12).Sub(iv("P8").MustMul(7))
		require.NoError(t, err)
		assert.True(t, got.Equal(iv("↓d2")), "got %v", got)
	})
}

func TestArithmeticSumErrors(t *testing.T) {
	t.Parallel()

	var u interval.Interval

	big := basicSize(interval.MaxSize)

	tests := []struct {
		name   string
		op     func() (interval.Interval, error)
		want   error
		causes []error
	}{
		{"Add unset receiver", func() (interval.Interval, error) { return u.Add(iv("M3")) }, interval.ErrUnset, nil},
		{"Add unset argument", func() (interval.Interval, error) { return iv("M3").Add(u) }, interval.ErrUnset, nil},
		{"Sub unset receiver", func() (interval.Interval, error) { return u.Sub(iv("M3")) }, interval.ErrUnset, nil},
		{"Sub unset argument", func() (interval.Interval, error) { return iv("M3").Sub(u) }, interval.ErrUnset, nil},
		{"Mul unset", func() (interval.Interval, error) { return u.Mul(0) }, interval.ErrUnset, nil},
		{"Add beyond MaxSize", func() (interval.Interval, error) { return big.Add(iv("M2")) }, interval.ErrOutOfRange, nil},
		{"Add descending beyond MaxSize", func() (interval.Interval, error) { return big.MustDown().Add(iv("↓M2")) }, interval.ErrOutOfRange, nil},
		{"Sub beyond MaxSize", func() (interval.Interval, error) { return big.Sub(iv("↓M2")) }, interval.ErrOutOfRange, nil},
		{"Add beyond M", func() (interval.Interval, error) { return aug(mag, 4).Add(iv("A1")) }, interval.ErrOutOfRange, nil},
		{"Add beyond -M", func() (interval.Interval, error) { return dim(mag, 5).Add(iv("↓A1")) }, interval.ErrOutOfRange, nil},
		{"Mul P8 MaxInt/7 times", func() (interval.Interval, error) { return iv("P8").Mul(q7) }, interval.ErrOutOfRange, nil},
		{"Mul P8 MaxInt times", func() (interval.Interval, error) { return iv("P8").Mul(maxInt) }, interval.ErrOutOfRange, nil},
		{"Mul P8 MinInt times", func() (interval.Interval, error) { return iv("P8").Mul(minInt) }, interval.ErrOutOfRange, nil},
		{"Mul A1 MinInt times", func() (interval.Interval, error) { return iv("A1").Mul(minInt) }, interval.ErrOutOfRange, nil},
		{"Mul A1 M+1 times", func() (interval.Interval, error) { return iv("A1").Mul(mag + 1) }, interval.ErrOutOfRange, nil},
		{"Mul M2 MaxInt times", func() (interval.Interval, error) { return iv("M2").Mul(maxInt) }, interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.op()
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}

// coordinates returns the linear coordinates of 4.6: t = σ(n − 1) and
// v = σ(b(n) + d), for small sizes.
func coordinates(i interval.Interval) (int, alteration.Alteration) {
	n := i.Size()
	v := ai(baseSteps(n)).MustAdd(i.Deviation())

	if i.IsDescending() {
		return -(n - 1), alteration.Natural().MustSub(v)
	}

	return n - 1, v
}

// baseSteps is b(n) = 12·⌊(n − 1)/7⌋ + B[(n − 1) mod 7] of 4.6.
func baseSteps(n int) int {
	b := [7]int{0, 2, 4, 5, 7, 9, 11}

	return 12*((n-1)/7) + b[(n-1)%7]
}

// fromCoordinates is the inverse map of 4.6.
func fromCoordinates(t int, v alteration.Alteration) interval.Interval {
	if t == 0 {
		if v.Sign() < 0 {
			return interval.MustFromDeviation(1, alteration.Natural().MustSub(v)).MustDown()
		}

		return interval.MustFromDeviation(1, v)
	}

	if t < 0 {
		n := -t + 1

		return interval.MustFromDeviation(n, alteration.Natural().MustSub(v).MustSub(ai(baseSteps(n)))).MustDown()
	}

	n := t + 1

	return interval.MustFromDeviation(n, v.MustSub(ai(baseSteps(n))))
}

// TestArithmeticSumProperties checks P-9 and the definition of 4.6 on all
// pairs of a sweep: the sum by the linear coordinates, commutativity,
// associativity, the neutral element, the opposite, Sub and Mul.
func TestArithmeticSumProperties(t *testing.T) {
	t.Parallel()

	set := predefinedBoth()
	set = append(set, both(ivs("n3", "{+1/2}1", "{-1/4}6", "{+1/3}4", "M17", "dd22"))...)

	small := both(ivs("P1", "A1", "m2", "M3", "P4", "d5", "P8", "n3", "M10", "AA4"))

	for _, a := range set {
		assert.True(t, a.MustAdd(interval.PerfectUnison()).Equal(a), "neutral: %v", a)
		assert.True(t, a.MustAdd(a.MustOpposite()).Equal(interval.PerfectUnison()), "opposite: %v", a)

		ta, va := coordinates(a)

		for _, b := range set {
			got, err := a.Add(b)
			require.NoError(t, err, "%v + %v", a, b)

			tb, vb := coordinates(b)
			want := fromCoordinates(ta+tb, va.MustAdd(vb))
			assert.True(t, got.Equal(want), "%v + %v = %v, want %v", a, b, got, want)
			assert.True(t, got.Equal(b.MustAdd(a)), "commutative: %v %v", a, b)
			assert.True(t, a.MustSub(b).Equal(a.MustAdd(b.MustOpposite())), "Sub: %v %v", a, b)
		}

		for _, b := range small {
			for _, c := range small {
				assert.True(t, a.MustAdd(b).MustAdd(c).Equal(a.MustAdd(b.MustAdd(c))), "associative: %v %v %v", a, b, c)
			}
		}

		sum := interval.PerfectUnison()
		for k := 1; k <= 4; k++ {
			sum = sum.MustAdd(a)
			assert.True(t, a.MustMul(k).Equal(sum), "%d times %v", k, a)
			assert.True(t, a.MustMul(-k).Equal(sum.MustOpposite()), "%d times %v", -k, a)
		}
	}
}
