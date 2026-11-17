package degree_test

import (
	"math"
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var degreeCanonicalRomanPattern = regexp.MustCompile(`^M{0,3}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})$`)

func degreeParseExpectation(text string) (int, error) {
	if text == "" {
		return 0, degree.ErrParse
	}

	decimal := true
	for i := range len(text) {
		if text[i] < '0' || text[i] > '9' {
			decimal = false
			break
		}
	}
	if decimal {
		if text == "0" {
			return 0, degree.ErrInvalidNumber
		}
		if text[0] == '0' {
			return 0, degree.ErrParse
		}
		number, ok := new(big.Int).SetString(text, 10)
		if !ok {
			return 0, degree.ErrParse
		}
		if number.Cmp(big.NewInt(int64(degree.MaxNumber))) > 0 {
			return 0, degree.ErrOutOfRange
		}
		return int(number.Int64()), nil
	}

	if !degreeCanonicalRomanPattern.MatchString(text) {
		return 0, degree.ErrParse
	}
	number := 0
	previous := 0
	for i := len(text) - 1; i >= 0; i-- {
		value := 0
		switch text[i] {
		case 'I':
			value = 1
		case 'V':
			value = 5
		case 'X':
			value = 10
		case 'L':
			value = 50
		case 'C':
			value = 100
		case 'D':
			value = 500
		case 'M':
			value = 1000
		}
		if value < previous {
			number -= value
		} else {
			number += value
		}
		previous = value
	}
	return number, nil
}

func FuzzDegreeParseNeverPanics(f *testing.F) {
	oversized := strings.Repeat("9", 40)
	for _, text := range []string{
		"", "0", "4", "IV", "3999", "MMMCMXCIX", "4000", "iv", "IIII", "99999999999999999999", "Ⅳ", "\x00\xff",
		oversized, oversized + "I", oversized + " ", oversized + "Ⅳ",
	} {
		f.Add(text)
	}
	f.Fuzz(func(t *testing.T, text string) {
		expectedNumber, expectedError := degreeParseExpectation(text)
		d, err := degree.Parse(text)
		if expectedError != nil {
			assertDegreeError(t, err, expectedError)
			assertUnsetDegree(t, d)
			return
		}
		require.NoError(t, err)
		assertDegreeState(t, d, expectedNumber, false, note.Note{})

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
