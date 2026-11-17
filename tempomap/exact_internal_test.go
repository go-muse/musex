package tempomap

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogFormIsZero(t *testing.T) {
	t.Parallel()
	zero := logForm{
		constant: big.NewRat(0, 1),
		terms: []logTerm{
			{coefficient: big.NewRat(1, 1), base: big.NewRat(4, 1)},
			{coefficient: big.NewRat(-2, 1), base: big.NewRat(2, 1)},
		},
	}
	almostZero := logForm{
		constant: big.NewRat(0, 1),
		terms: []logTerm{
			{coefficient: big.NewRat(1, 1), base: big.NewRat(4, 1)},
			{coefficient: big.NewRat(-2, 1), base: big.NewRat(2, 1)},
			{coefficient: big.NewRat(1, 1), base: big.NewRat(1000001, 1000000)},
		},
	}
	algebraicIrrational := logForm{
		constant: big.NewRat(1, 1),
		terms: []logTerm{
			{coefficient: big.NewRat(1, 2), base: big.NewRat(2, 1)},
		},
	}

	for _, tt := range []struct {
		name string
		form logForm
		want bool
	}{
		{name: "ln4 minus two ln2 is exact zero", form: zero, want: true},
		{name: "additional small logarithm is not zero", form: almostZero, want: false},
		{name: "nonzero constant with algebraic irrational candidate", form: algebraicIrrational, want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, logFormIsZero(tt.form))
		})
	}
}
