package pitchclass_test

import (
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/require"
)

// BenchmarkTotient measures phi(N) on the hard moduli of the plan: the
// largest prime below 2^63, a semiprime and a prime square near 2^63, a
// strong pseudoprime to small bases, a power of 3, MAX and 2^62. There are
// no thresholds; QA records the results.
func BenchmarkTotient(b *testing.B) {
	benchmarks := []struct {
		name string
		n    int
		want int
	}{
		{name: "prime=9223372036854775783", n: 9223372036854775783, want: 9223372036854775782},
		{name: "semiprime=9223371873002223329", n: 9223371873002223329, want: 9223371866928222384},
		{name: "square=9223371994482243049", n: 9223371994482243049, want: 9223371991445242556},
		{name: "pseudoprime=3825123056546413051", n: 3825123056546413051, want: 3825092239639605000},
		{name: "3^39", n: 4052555153018976267, want: 2701703435345984178},
		{name: "MAX", n: maxInt, want: 7713001620195508224},
		{name: "2^62", n: 1 << 62, want: 1 << 61},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			phi, err := pitchclass.Totient(bm.n)
			require.NoError(b, err)
			require.Equal(b, bm.want, phi)

			for b.Loop() {
				_, _ = pitchclass.Totient(bm.n)
			}
		})
	}
}
