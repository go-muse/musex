package rounding_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeConstants(t *testing.T) {
	t.Parallel()
	for i, mode := range testModes {
		t.Run(testModeNames[i], func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, rounding.Mode(i+1), mode)
		})
	}
}

func TestModes(t *testing.T) {
	t.Parallel()
	got := rounding.Modes()
	require.Len(t, got, 9)
	assert.Equal(t, testModes, got)
	assert.Equal(t, 9, cap(got))
	got[0] = rounding.Mode(255)
	assert.Equal(t, testModes, rounding.Modes())
}

func TestModePredicates(t *testing.T) {
	t.Parallel()
	values := []rounding.Mode{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 255}
	mirrors := []rounding.Mode{
		0,
		rounding.Up,
		rounding.Down,
		rounding.TowardZero,
		rounding.AwayFromZero,
		rounding.HalfDown,
		rounding.HalfUp,
		rounding.HalfAwayFromZero,
		rounding.HalfTowardZero,
		rounding.HalfEven,
		10,
		255,
	}
	for i, mode := range values {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			t.Parallel()
			valid := i >= 1 && i <= 9
			assert.Equal(t, valid, mode.IsValid())
			assert.Equal(t, i >= 1 && i <= 4, mode.IsDirected())
			assert.Equal(t, i >= 5 && i <= 9, mode.IsNearest())
			assert.Equal(t, mirrors[i], mode.Mirror())
			assert.Equal(t, mode, mode.Mirror().Mirror())
		})
	}
}

func TestModeEqual(t *testing.T) {
	t.Parallel()
	values := []rounding.Mode{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 255}
	for _, left := range values {
		for _, right := range values {
			t.Run(fmt.Sprintf("%d_%d", left, right), func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, left == right, left.Equal(right))
			})
		}
	}
}
