package tempo_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Architecture 26.9, decisions 1 and 2: QuartersPerMinute = 4·u·N and
// DurationIn(s) = s·(1/60)·u·N are exact products checked only at the final
// value. Each case below keeps the result in range while some pairwise
// product of a chain of multiplications is out of range, so a chain would
// refuse where the exact product succeeds. M is fraction.MaxComponent.
func TestQuartersPerMinuteExactProduct(t *testing.T) {
	t.Parallel()
	m := fraction.MaxComponentBig()

	tests := []struct {
		name string
		mark tempo.Tempo
		want fraction.Fraction
	}{
		// Decision 2 with a = 1: u·N = 1/(4M) is out of range.
		{"u·N beyond range, ♩ = 1/M", tm(q, invMaxF()), invMaxF()},
		// Decision 2 with a = 3 (M ≡ 1 mod 3): u·N = 3/(4M) is out of range.
		{"u·N beyond range, ♩ = 3/M", tm(q, fbig(big.NewInt(3), m)), fbig(big.NewInt(3), m)},
		// 4·N = 4M is out of range; u·N = M/8 and 4·u = 1/2 are not.
		{"4·N beyond range, ♪ = M", tm(e, maxF()), fbig(m, big.NewInt(2))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.mark.QuartersPerMinute()
			require.NoError(t, err)
			assertFrac(t, tt.want, got, "QuartersPerMinute")
			assertFrac(t, tt.want, tt.mark.MustQuartersPerMinute(), "MustQuartersPerMinute")
		})
	}
}

// The speed u·N of ♩ = 1/M is itself out of range: the input of the first
// QuartersPerMinute case really overflows that intermediate product.
func TestQuartersPerMinuteExactProductIntermediate(t *testing.T) {
	t.Parallel()

	_, err := tm(q, invMaxF()).Speed()
	require.Error(t, err)
	assert.ErrorIs(t, err, tempo.ErrOutOfRange)
	assert.ErrorIs(t, err, fraction.ErrOutOfRange)
}

func TestDurationInExactProduct(t *testing.T) {
	t.Parallel()
	m := fraction.MaxComponentBig()

	tests := []struct {
		name string
		mark tempo.Tempo
		s    tempo.Seconds
		want fraction.Fraction
	}{
		// Decision 1 (bpm < 1): u·N = 1/(4M) and N/60 = 1/(60M) are out of
		// range; s·N = 1, s/60 = M/60 and s·u = M/4 are not.
		{"u·N and N/60 beyond range, ♩ = 1/M, s = M", tm(q, invMaxF()), secF(maxF()), fr(1, 240)},
		// s/60 = 1/(60M) and s·u = 1/(4M) are out of range; u·N = 60 and
		// s·N = 240/M are not.
		{"s/60 and s·u beyond range, ♩ = 240, s = 1/M", tq(240), secF(invMaxF()), invMaxF()},
		// s·N = 2M is out of range; s/60, s·u and u·N are not. M is coprime
		// to 120, so the result M/120 is reduced.
		{"s·N beyond range, ♩ = 2, s = M", tq(2), secF(maxF()), fbig(m, big.NewInt(120))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.mark.DurationIn(tt.s)
			require.NoError(t, err)
			assertDur(t, durF(tt.want), got, "DurationIn")
			assertDur(t, durF(tt.want), tt.mark.MustDurationIn(tt.s), "MustDurationIn")
		})
	}
}
