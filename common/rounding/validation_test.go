package rounding_test

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
)

var (
	_ fmt.Stringer             = rounding.Mode(0)
	_ encoding.TextMarshaler   = rounding.Mode(0)
	_ encoding.TextAppender    = rounding.Mode(0)
	_ encoding.TextUnmarshaler = (*rounding.Mode)(nil)
	_ fmt.Stringer             = rounding.Side(0)
	_ fmt.Stringer             = rounding.Neighbor(0)
)

func TestContractConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 400, rounding.MaxPlaces)
	assert.Equal(t, rounding.Side(-1), rounding.Negative)
	assert.Equal(t, rounding.Side(1), rounding.Positive)
	assert.Equal(t, rounding.Neighbor(1), rounding.Lower)
	assert.Equal(t, rounding.Neighbor(2), rounding.Upper)
}

func TestModeValidationAcrossOperations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		call func(rounding.Mode) ([]any, error)
	}{
		{"Choose", func(m rounding.Mode) ([]any, error) {
			v, err := m.Choose(0, rounding.Positive, false)
			return []any{v}, err
		}},
		{"ChooseIndex", func(m rounding.Mode) ([]any, error) { v, err := m.ChooseIndex(0, 0); return []any{v}, err }},
		{"QuoRem", func(m rounding.Mode) ([]any, error) { q, r, err := m.QuoRem(7, 2); return []any{q, r}, err }},
		{"IntMultiple", func(m rounding.Mode) ([]any, error) { v, err := m.IntMultiple(1200, 480); return []any{v}, err }},
		{"QuoRemBig", func(m rounding.Mode) ([]any, error) {
			q, r, err := m.QuoRemBig(bigInt(7), bigInt(2))
			return []any{q, r}, err
		}},
		{"RatInt", func(m rounding.Mode) ([]any, error) { v, err := m.RatInt(bigRat(7, 3)); return []any{v}, err }},
		{"RatMultiple", func(m rounding.Mode) ([]any, error) {
			k, v, err := m.RatMultiple(bigRat(7, 3), bigRat(1, 4))
			return []any{k, v}, err
		}},
		{"RatDecimal", func(m rounding.Mode) ([]any, error) { v, err := m.RatDecimal(bigRat(1, 8), 2); return []any{v}, err }},
		{"Float", func(m rounding.Mode) ([]any, error) { v, err := m.Float(2.5); return []any{v}, err }},
		{"FloatInt", func(m rounding.Mode) ([]any, error) { v, err := m.FloatInt(2.5); return []any{v}, err }},
		{"FloatMultiple", func(m rounding.Mode) ([]any, error) { v, err := m.FloatMultiple(0.15, 0.1); return []any{v}, err }},
		{"FloatSteps", func(m rounding.Mode) ([]any, error) { v, err := m.FloatSteps(0.15, 0.1); return []any{v}, err }},
		{"FloatPlaces", func(m rounding.Mode) ([]any, error) { v, err := m.FloatPlaces(0.145, 2); return []any{v}, err }},
		{"FloatDecimal", func(m rounding.Mode) ([]any, error) { v, err := m.FloatDecimal(0.145, 2); return []any{v}, err }},
		{"FloatFraction", func(m rounding.Mode) ([]any, error) { v, err := m.FloatFraction(0.5, 3); return []any{v}, err }},
		{"FloatNumerator", func(m rounding.Mode) ([]any, error) { v, err := m.FloatNumerator(0.5, 3); return []any{v}, err }},
	}
	for _, tt := range tests {
		for _, mode := range []rounding.Mode{0, 10, 255} {
			t.Run(fmt.Sprintf("%s/mode_%d", tt.name, mode), func(t *testing.T) {
				t.Parallel()
				results, err := tt.call(mode)
				want := rounding.ErrUnknownMode
				if mode == 0 {
					want = rounding.ErrUnset
				}
				checkError(t, err, want)
				for _, result := range results {
					assert.Zero(t, result)
				}
			})
		}
	}
}
