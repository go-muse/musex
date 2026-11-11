package rounding_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
)

func TestBigNilArgumentsReturnZeroResults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		call func() ([]any, error)
	}{
		{"quo_rem_nil_n", func() ([]any, error) { q, r, err := rounding.Down.QuoRemBig(nil, bigInt(2)); return []any{q, r}, err }},
		{"quo_rem_nil_d", func() ([]any, error) { q, r, err := rounding.Down.QuoRemBig(bigInt(2), nil); return []any{q, r}, err }},
		{"rat_int_nil", func() ([]any, error) { v, err := rounding.Down.RatInt(nil); return []any{v}, err }},
		{"rat_multiple_nil_x", func() ([]any, error) {
			k, v, err := rounding.Down.RatMultiple(nil, bigRat(1, 2))
			return []any{k, v}, err
		}},
		{"rat_multiple_nil_step", func() ([]any, error) {
			k, v, err := rounding.Down.RatMultiple(bigRat(1, 2), nil)
			return []any{k, v}, err
		}},
		{"rat_decimal_nil", func() ([]any, error) { v, err := rounding.Down.RatDecimal(nil, 2); return []any{v}, err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			results, err := tt.call()
			checkError(t, err, rounding.ErrUnset)
			for _, result := range results {
				assert.Zero(t, result)
			}
		})
	}
}
