package degree_test

import (
	"errors"
	"math"
	"testing"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzDegreeParseNeverPanics(f *testing.F) {
	for _, text := range []string{"", "0", "4", "IV", "3999", "MMMCMXCIX", "4000", "iv", "IIII", "99999999999999999999", "Ⅳ", "\x00\xff"} {
		f.Add(text)
	}
	f.Fuzz(func(t *testing.T, text string) {
		d, err := degree.Parse(text)
		if err != nil {
			assertUnsetDegree(t, d)
			assert.True(t,
				errors.Is(err, degree.ErrParse) || errors.Is(err, degree.ErrInvalidNumber) || errors.Is(err, degree.ErrOutOfRange),
				"unexpected parse error category: %v", err,
			)
			return
		}
		require.True(t, d.IsValid())
		assert.Positive(t, d.Number())
		assert.False(t, d.HasNote())
		decimal, err := d.Decimal()
		require.NoError(t, err)
		fromDecimal, err := degree.Parse(decimal)
		require.NoError(t, err)
		assert.Equal(t, d.Number(), fromDecimal.Number())
		if d.Number() <= degree.MaxRomanNumber {
			roman, err := d.Roman()
			require.NoError(t, err)
			fromRoman, err := degree.Parse(roman)
			require.NoError(t, err)
			assert.Equal(t, d.Number(), fromRoman.Number())
		}
	})
}

func FuzzDegreeCanonicalFormattingRoundTrip(f *testing.F) {
	for _, number := range []int{1, 4, 8, 9, 40, 53, 90, 400, 900, 1994, 3999} {
		f.Add(number)
	}
	f.Fuzz(func(t *testing.T, number int) {
		if number < 1 || number > degree.MaxRomanNumber {
			t.Skip("outside the canonical Roman range")
		}
		d := degree.MustNew(number)
		decimal, err := d.Decimal()
		require.NoError(t, err)
		fromDecimal, err := degree.Parse(decimal)
		require.NoError(t, err)
		assertDegreeState(t, fromDecimal, number, false, note.Note{})
		roman, err := d.Roman()
		require.NoError(t, err)
		fromRoman, err := degree.Parse(roman)
		require.NoError(t, err)
		assertDegreeState(t, fromRoman, number, false, note.Note{})
	})
}

func FuzzDegreeCircleProperties(f *testing.F) {
	seeds := []struct {
		k, j, n, target, m int
	}{
		{1, 1, math.MinInt, 1, math.MaxInt},
		{2, 2, math.MaxInt, 1, 3},
		{7, 1, math.MinInt, 7, 8},
		{7, 7, math.MaxInt, 1, math.MaxInt},
		{13, 11, -39, 4, 65},
		{math.MaxInt, math.MaxInt, math.MinInt, 1, math.MaxInt},
	}
	for _, seed := range seeds {
		f.Add(seed.k, seed.j, seed.n, seed.target, seed.m)
	}
	f.Fuzz(func(t *testing.T, k, j, n, target, m int) {
		if k < 1 || j < 1 || j > k || target < 1 || target > k || m < 1 {
			t.Skip("outside valid circular-counting inputs")
		}
		d := degree.MustNew(j)
		expected := shiftByContract(k, j, n)
		shifted, cycles, err := d.Shift(k, n)
		require.NoError(t, err)
		assertDegreeState(t, shifted, expected.number, false, note.Note{})
		assert.Equal(t, expected.cycles, cycles)
		assert.GreaterOrEqual(t, shifted.Number(), 1)
		assert.LessOrEqual(t, shifted.Number(), k)

		to := degree.MustNew(target)
		up, err := d.DistanceUp(to, k)
		require.NoError(t, err)
		down, err := d.DistanceDown(to, k)
		require.NoError(t, err)
		expectedUp := target - j
		if expectedUp < 0 {
			expectedUp = k - j + target
		}
		expectedDown := j - target
		if expectedDown < 0 {
			expectedDown = k - target + j
		}
		assert.Equal(t, expectedUp, up)
		assert.Equal(t, expectedDown, down)
		assert.GreaterOrEqual(t, up, 0)
		assert.Less(t, up, k)
		assert.GreaterOrEqual(t, down, 0)
		assert.Less(t, down, k)
		if up == 0 {
			assert.Zero(t, down)
		} else {
			assert.Equal(t, k, up+down)
		}

		shortest, err := d.ShortestDistance(to, k)
		require.NoError(t, err)
		expectedShortest := up
		if up > down {
			expectedShortest = -down
		}
		assert.Equal(t, expectedShortest, shortest)
		reached, _, err := d.Shift(k, shortest)
		require.NoError(t, err)
		assertSameDegree(t, to, reached)

		normalized, normalizedCycles, err := degree.MustNew(m).Normalize(k)
		require.NoError(t, err)
		assertDegreeState(t, normalized, (m-1)%k+1, false, note.Note{})
		assert.Equal(t, (m-1)/k, normalizedCycles)
		denormalized, err := normalized.Denormalize(k, normalizedCycles)
		require.NoError(t, err)
		assertDegreeState(t, denormalized, m, false, note.Note{})
	})
}
