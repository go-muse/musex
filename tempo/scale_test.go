package tempo_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scale is the Maelzel scale of business-spec 2.4 (tempo.inv-23), indexes
// from 0.
var scale = []int{
	40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 63, 66, 69, 72, 76, 80, 84, 88, 92, 96, 100, 104, 108, 112, 116,
	120, 126, 132, 138, 144, 152, 160, 168, 176, 184, 192, 200, 208,
}

// TestScaleDivisions covers С-21 and architecture 4.12: the 39 divisions,
// the constants, the division by index, membership, and index.
func TestScaleDivisions(t *testing.T) {
	t.Parallel()

	assert.Equal(t, scale, tempo.ScaleDivisions())
	assert.Equal(t, 40, tempo.ScaleMin)
	assert.Equal(t, 208, tempo.ScaleMax)
	assert.Equal(t, 39, tempo.ScaleSize)

	got := tempo.ScaleDivisions()
	got[0] = 1
	assert.Equal(t, 40, tempo.ScaleDivisions()[0], "ScaleDivisions returns a new slice")

	for i, d := range scale {
		v, err := tempo.ScaleDivision(i)
		require.NoError(t, err)
		assert.Equal(t, d, v, "ScaleDivision(%d)", i)

		idx, err := tempo.ScaleIndex(fi(d))
		require.NoError(t, err)
		assert.Equal(t, i, idx, "ScaleIndex(%d)", d)

		ok, err := tempo.IsScaleDivision(fi(d))
		require.NoError(t, err)
		assert.True(t, ok, "%d", d)
	}

	assert.Equal(t, 21, tempo.MustScaleIndex(fi(100)), "С-21")

	for _, b := range []fraction.Fraction{fi(110), fr(123, 2), fi(30), fi(39), fi(41), fi(209), fi(250), fr(81, 2)} {
		ok, err := tempo.IsScaleDivision(b)
		require.NoError(t, err)
		assert.False(t, ok, "%s", b)

		idx, err := tempo.ScaleIndex(b)
		offScale.check(t, err)
		assert.Equal(t, 0, idx)
	}

	for _, i := range []int{39, -1, 1000} {
		v, err := tempo.ScaleDivision(i)
		argClass.check(t, err)
		assert.Equal(t, 0, v)
	}
}

// TestScaleNextPrev covers С-21, С-39 and architecture 4.12: the next
// division strictly above or below, and Е-6 where there is none.
func TestScaleNextPrev(t *testing.T) {
	t.Parallel()

	next := []struct {
		b    fraction.Fraction
		want int
	}{
		{fi(60), 63}, {fi(120), 126}, {fi(110), 112}, {fi(30), 40}, {fi(40), 42}, {fi(207), 208},
		{fr(1, 2), 40}, {fr(121, 2), 63}, {fi(200), 208},
	}

	for _, tt := range next {
		got, err := tempo.NextScaleDivision(tt.b)
		require.NoError(t, err, "%s", tt.b)
		assert.Equal(t, tt.want, got, "Next(%s)", tt.b)
	}

	prev := []struct {
		b    fraction.Fraction
		want int
	}{
		{fi(120), 116}, {fi(110), 108}, {fi(250), 208}, {fi(41), 40}, {fi(209), 208}, {fi(208), 200},
		{fr(121, 2), 60}, {fi(63), 60},
	}

	for _, tt := range prev {
		got, err := tempo.PrevScaleDivision(tt.b)
		require.NoError(t, err, "%s", tt.b)
		assert.Equal(t, tt.want, got, "Prev(%s)", tt.b)
	}

	for _, b := range []fraction.Fraction{fi(208), fi(250), fr(417, 2)} {
		got, err := tempo.NextScaleDivision(b)
		offScale.check(t, err)
		assert.Equal(t, 0, got)
	}

	for _, b := range []fraction.Fraction{fi(40), fi(30), fr(79, 2)} {
		got, err := tempo.PrevScaleDivision(b)
		offScale.check(t, err)
		assert.Equal(t, 0, got)
	}
}

// TestScaleNeighbors covers С-22, С-39, С-55, С-56 and architecture 4.12:
// the lower and upper neighbors and which is nearer by difference.
func TestScaleNeighbors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		b                 fraction.Fraction
		lower, upper, cmp int
	}{
		{fi(62), 60, 63, 1},
		{fi(130), 126, 132, 1},
		{fi(100), 100, 100, 0},
		{fi(110), 108, 112, 0},
		{fi(74), 72, 76, 0},
		{fr(123, 2), 60, 63, 0},
		{fr(121, 2), 60, 63, -1},
		{fi(40), 40, 40, 0},
		{fi(208), 208, 208, 0},
		{fi(41), 40, 42, 0},
		{fr(81, 2), 40, 42, -1},
		{fi(205), 200, 208, 1},
		{fi(201), 200, 208, -1},
	}

	for _, tt := range tests {
		lo, hi, cmp, err := tempo.ScaleNeighbors(tt.b)
		require.NoError(t, err, "%s", tt.b)
		assert.Equal(t, []int{tt.lower, tt.upper, tt.cmp}, []int{lo, hi, cmp}, "ScaleNeighbors(%s)", tt.b)
	}

	for _, b := range []fraction.Fraction{fi(30), fi(250), fi(39), fi(209), fr(79, 2), fr(417, 2)} {
		lo, hi, cmp, err := tempo.ScaleNeighbors(b)
		offScale.check(t, err)
		assert.Equal(t, []int{0, 0, 0}, []int{lo, hi, cmp})
	}
}

// TestNearestScaleDivision covers С-22, С-55, С-59 and architecture 4.8,
// 4.12: the division by the mode, the parity of the lower index for HalfEven,
// a division in any mode.
func TestNearestScaleDivision(t *testing.T) {
	t.Parallel()

	modes := []roundingMode{
		down, up, towardZero, awayFromZero, halfUp, halfDown, halfAwayFromZero, halfTowardZero, halfEven,
	}

	tests := []struct {
		name string
		b    fraction.Fraction
		want []int // in the order of modes
	}{
		{"С-22 110, lower 108 has index 23", fi(110), []int{108, 112, 108, 112, 112, 108, 112, 108, 112}},
		{"С-22 74, lower 72 has index 14", fi(74), []int{72, 76, 72, 76, 76, 72, 76, 72, 72}},
		{"С-22 123/2, lower 60 has index 10", fr(123, 2), []int{60, 63, 60, 63, 63, 60, 63, 60, 60}},
		{"С-22 62", fi(62), []int{60, 63, 60, 63, 63, 63, 63, 63, 63}},
		{"С-22 130", fi(130), []int{126, 132, 126, 132, 132, 132, 132, 132, 132}},
		{"С-56 121/2", fr(121, 2), []int{60, 63, 60, 63, 60, 60, 60, 60, 60}},
		{"С-55 120", fi(120), []int{120, 120, 120, 120, 120, 120, 120, 120, 120}},
		{"С-55 40", fi(40), []int{40, 40, 40, 40, 40, 40, 40, 40, 40}},
		{"С-55 208", fi(208), []int{208, 208, 208, 208, 208, 208, 208, 208, 208}},
		{"41, lower 40 has index 0", fi(41), []int{40, 42, 40, 42, 42, 40, 42, 40, 40}},
		{"204, lower 200 has index 37", fi(204), []int{200, 208, 200, 208, 208, 200, 208, 200, 208}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for i, m := range modes {
				got, err := tempo.NearestScaleDivision(tt.b, m)
				require.NoError(t, err, "%s", m)
				assert.Equal(t, tt.want[i], got, "%s", m)
			}
		})
	}
}

// TestScaleErrors covers С-39, С-44 and architecture 4.3, 4.12: Е-3 for a
// number <= 0 in every function with a number, Е-2 for an unset number, and
// the order of NearestScaleDivision: Е-2 (number) -> Е-3 -> Е-2 (mode) -> Е-6.
func TestScaleErrors(t *testing.T) {
	t.Parallel()

	type call = func(b fraction.Fraction) (any, error)

	funcs := map[string]call{
		"IsScaleDivision": func(b fraction.Fraction) (any, error) { return tempo.IsScaleDivision(b) },
		"ScaleIndex":      func(b fraction.Fraction) (any, error) { return tempo.ScaleIndex(b) },
		"Next":            func(b fraction.Fraction) (any, error) { return tempo.NextScaleDivision(b) },
		"Prev":            func(b fraction.Fraction) (any, error) { return tempo.PrevScaleDivision(b) },
		"ScaleNeighbors": func(b fraction.Fraction) (any, error) {
			lo, hi, cmp, err := tempo.ScaleNeighbors(b)
			return []int{lo, hi, cmp}, err
		},
		"NearestScaleDivision": func(b fraction.Fraction) (any, error) {
			return tempo.NearestScaleDivision(b, halfUp)
		},
	}

	zeros := map[string]any{
		"IsScaleDivision": false, "ScaleIndex": 0, "Next": 0, "Prev": 0, "ScaleNeighbors": []int{0, 0, 0},
		"NearestScaleDivision": 0,
	}

	for name, f := range funcs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, b := range []fraction.Fraction{fi(0), fi(-5), fr(-1, 2)} {
				got, err := f(b)
				noTempoClass.check(t, err)
				assert.Equal(t, zeros[name], got, "С-44 %s(%s)", name, b)
			}

			got, err := f(fraction.Fraction{})
			argClass.check(t, err)
			assert.Equal(t, zeros[name], got)
		})
	}

	t.Run("NearestScaleDivision off scale", func(t *testing.T) {
		t.Parallel()

		for _, b := range []fraction.Fraction{fi(30), fi(250)} {
			got, err := tempo.NearestScaleDivision(b, halfUp)
			offScale.check(t, err)
			assert.Equal(t, 0, got, "С-39 %s", b)
		}
	})

	t.Run("NearestScaleDivision order", func(t *testing.T) {
		t.Parallel()

		_, err := tempo.NearestScaleDivision(fraction.Fraction{}, roundingMode(0))
		argClass.check(t, err)

		_, err = tempo.NearestScaleDivision(fi(0), roundingMode(0))
		noTempoClass.check(t, err)

		_, err = tempo.NearestScaleDivision(fi(30), roundingMode(0))
		modeUnset.check(t, err)

		_, err = tempo.NearestScaleDivision(fi(120), roundingMode(0))
		modeUnset.check(t, err)

		_, err = tempo.NearestScaleDivision(fi(110), roundingMode(99))
		modeUnknown.check(t, err)
	})
}
