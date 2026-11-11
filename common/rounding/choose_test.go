package rounding_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeChoose(t *testing.T) {
	t.Parallel()
	comparisons := []int{-7, -1, 0, 1, 5, math.MinInt, math.MaxInt}
	for i, mode := range testModes {
		for _, comparison := range comparisons {
			for _, side := range []rounding.Side{rounding.Negative, rounding.Positive} {
				for _, lowerEven := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/cmp_%d/side_%d/even_%t", testModeNames[i], comparison, side, lowerEven), func(t *testing.T) {
						t.Parallel()
						got, err := mode.Choose(comparison, side, lowerEven)
						require.NoError(t, err)
						assert.Equal(t, expectedNeighbor(mode, comparison, side, lowerEven), got)
					})
				}
			}
		}
	}
}

func TestModeChooseRejectsInvalidSide(t *testing.T) {
	t.Parallel()
	for i, mode := range testModes {
		for _, side := range []rounding.Side{0, 2, -2, 127, -128} {
			t.Run(fmt.Sprintf("%s/side_%d", testModeNames[i], side), func(t *testing.T) {
				t.Parallel()
				got, err := mode.Choose(0, side, false)
				checkError(t, err, rounding.ErrInvalidArgument)
				assert.Equal(t, rounding.Neighbor(0), got)
			})
		}
	}
	for _, tt := range []struct {
		mode rounding.Mode
		want error
	}{{0, rounding.ErrUnset}, {10, rounding.ErrUnknownMode}, {255, rounding.ErrUnknownMode}} {
		t.Run(fmt.Sprintf("priority_%d", tt.mode), func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.Choose(0, 0, false)
			checkError(t, err, tt.want)
			assert.Equal(t, rounding.Neighbor(0), got)
		})
	}
}

func TestModeChooseIndexControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		mode  rounding.Mode
		lower int
		cmp   int
		want  int
		err   error
	}{
		{"half_down_positive_tie", rounding.HalfDown, 0, 0, 0, nil},
		{"half_up_positive_tie", rounding.HalfUp, 0, 0, 1, nil},
		{"down_nearer_lower", rounding.Down, 0, -1, 0, nil},
		{"up_nearer_lower", rounding.Up, 0, -1, 1, nil},
		{"toward_zero_negative_nearer_lower", rounding.TowardZero, -1, -1, 0, nil},
		{"toward_zero_positive_nearer_upper", rounding.TowardZero, 0, 1, 0, nil},
		{"away_negative_nearer_upper", rounding.AwayFromZero, -1, 1, -1, nil},
		{"away_positive_nearer_lower", rounding.AwayFromZero, 0, -1, 1, nil},
		{"half_even_odd_positive", rounding.HalfEven, 3, 0, 4, nil},
		{"half_even_odd_negative", rounding.HalfEven, -3, 0, -2, nil},
		{"half_away_negative_tie", rounding.HalfAwayFromZero, -1, 0, -1, nil},
		{"half_toward_negative_tie", rounding.HalfTowardZero, -1, 0, 0, nil},
		{"upper_at_max", rounding.Up, math.MaxInt, 0, 0, rounding.ErrOutOfRange},
		{"lower_at_max", rounding.Down, math.MaxInt, 0, math.MaxInt, nil},
		{"half_up_at_min", rounding.HalfUp, math.MinInt, 0, math.MinInt + 1, nil},
		{"unset", 0, 0, 0, 0, rounding.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.ChooseIndex(tt.lower, tt.cmp)
			if tt.err != nil {
				checkError(t, err, tt.err)
				assert.Zero(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModeChooseIndexMatchesChoose(t *testing.T) {
	t.Parallel()
	lowers := []int{math.MinInt, -3, -2, -1, 0, 1, 2, 3, math.MaxInt}
	for _, lower := range lowers {
		for i, mode := range testModes {
			for _, comparison := range []int{-1, 0, 1} {
				t.Run(fmt.Sprintf("lower_%d/%s/cmp_%d", lower, testModeNames[i], comparison), func(t *testing.T) {
					t.Parallel()
					got, err := mode.ChooseIndex(lower, comparison)
					side := rounding.Positive
					if lower < 0 {
						side = rounding.Negative
					}
					neighbor := expectedNeighbor(mode, comparison, side, lower%2 == 0)
					if lower == math.MaxInt && neighbor == rounding.Upper {
						checkError(t, err, rounding.ErrOutOfRange)
						assert.Zero(t, got)
						return
					}
					require.NoError(t, err)
					want := lower
					if neighbor == rounding.Upper {
						want++
					}
					assert.Equal(t, want, got)
				})
			}
		}
	}
}
