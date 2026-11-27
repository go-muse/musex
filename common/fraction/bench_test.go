package fraction_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/require"
)

// Benchmarks of task 26.6 by width class, with big.Rat references on the same
// values (architecture, "Benchmark list"). Names are stable for comparison:
// Benchmark<Operation>/<class>, Benchmark<Operation>/<class>-short and
// BenchmarkRat<Operation>/<class>. Every operation is checked once before the
// loop; the big.Rat references allocate a new receiver per operation, which is
// the semantics of the fraction operations that return a new value.

// benchEach runs body as a sub-benchmark for every input of the family, after check.
func benchEach(b *testing.B, family benchFamily, check func(in benchInput) error, body func(b *testing.B, in benchInput)) {
	for _, in := range benchInputs(family) {
		b.Run(in.class, func(b *testing.B) {
			require.NoError(b, check(in))
			b.ReportAllocs()
			body(b, in)
		})
	}
}

func noCheck(benchInput) error { return nil }

func BenchmarkAdd(b *testing.B) {
	benchEach(b, benchBinary, func(in benchInput) error {
		_, err := in.x.Add(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.Add(y)
		}
	})
}

func BenchmarkSub(b *testing.B) {
	benchEach(b, benchBinary, func(in benchInput) error {
		_, err := in.x.Sub(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.Sub(y)
		}
	})
}

func BenchmarkMul(b *testing.B) {
	benchEach(b, benchBinary, func(in benchInput) error {
		_, err := in.x.Mul(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.Mul(y)
		}
	})
}

func BenchmarkDiv(b *testing.B) {
	benchEach(b, benchBinary, func(in benchInput) error {
		_, err := in.x.Div(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.Div(y)
		}
	})
}

func BenchmarkAddWithUnreducedLCM(b *testing.B) {
	benchEach(b, benchLCM, func(in benchInput) error {
		_, err := in.x.AddWith(in.y, fraction.DenomUnreducedLCM)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.AddWith(y, fraction.DenomUnreducedLCM)
		}
	})
}

func BenchmarkAddWithLargerUnit(b *testing.B) {
	benchEach(b, benchUnit, func(in benchInput) error {
		_, err := in.x.AddWith(in.y, fraction.DenomLargerUnit)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.AddWith(y, fraction.DenomLargerUnit)
		}
	})
}

func BenchmarkAddInDenominator(b *testing.B) {
	benchEach(b, benchUnit, func(in benchInput) error {
		_, err := in.x.AddInDenominator(in.y, in.den)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y, den := in.x, in.y, in.den
		for b.Loop() {
			_, _ = x.AddInDenominator(y, den)
		}
	})
}

func BenchmarkCompare(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_ = x.Compare(y)
		}
	})
}

func BenchmarkEqual(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_ = x.Equal(y)
		}
	})
}

func BenchmarkQuoRem(b *testing.B) {
	benchEach(b, benchDivMod, func(in benchInput) error {
		_, _, err := in.x.QuoRem(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _, _ = x.QuoRem(y)
		}
	})
}

func BenchmarkDivModFloor(b *testing.B) {
	benchEach(b, benchDivMod, func(in benchInput) error {
		_, _, err := in.x.DivModFloor(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _, _ = x.DivModFloor(y)
		}
	})
}

func BenchmarkIsMultipleOf(b *testing.B) {
	benchEach(b, benchDivMod, func(in benchInput) error {
		_, err := in.x.IsMultipleOf(in.y)
		return err
	}, func(b *testing.B, in benchInput) {
		x, y := in.x, in.y
		for b.Loop() {
			_, _ = x.IsMultipleOf(y)
		}
	})
}

func BenchmarkReduce(b *testing.B) {
	benchEach(b, benchReduce, noCheck, func(b *testing.B, in benchInput) {
		x := in.x
		for b.Loop() {
			_ = x.Reduce()
		}
	})
}

func BenchmarkIsReduced(b *testing.B) {
	benchEach(b, benchReduce, noCheck, func(b *testing.B, in benchInput) {
		x := in.x
		for b.Loop() {
			_ = x.IsReduced()
		}
	})
}

func BenchmarkIsInteger(b *testing.B) {
	benchEach(b, benchSplit, noCheck, func(b *testing.B, in benchInput) {
		x := in.x
		for b.Loop() {
			_ = x.IsInteger()
		}
	})
}

func BenchmarkSplit(b *testing.B) {
	benchEach(b, benchSplit, func(in benchInput) error {
		_, _, err := in.x.Split()
		return err
	}, func(b *testing.B, in benchInput) {
		x := in.x
		for b.Loop() {
			_, _, _ = x.Split()
		}
	})
}

func BenchmarkSplitFloor(b *testing.B) {
	benchEach(b, benchSplit, func(in benchInput) error {
		_, _, err := in.x.SplitFloor()
		return err
	}, func(b *testing.B, in benchInput) {
		x := in.x
		for b.Loop() {
			_, _, _ = x.SplitFloor()
		}
	})
}

func BenchmarkRoundToInt(b *testing.B) {
	benchEach(b, benchSplit, func(in benchInput) error {
		_, err := in.x.RoundToInt(fraction.RoundHalfToEven)
		return err
	}, func(b *testing.B, in benchInput) {
		x := in.x
		for b.Loop() {
			_, _ = x.RoundToInt(fraction.RoundHalfToEven)
		}
	})
}

func BenchmarkRoundToDenominator(b *testing.B) {
	benchEach(b, benchRoundDen, func(in benchInput) error {
		_, err := in.x.RoundToDenominator(in.den, fraction.RoundHalfToEven)
		return err
	}, func(b *testing.B, in benchInput) {
		x, den := in.x, in.den
		for b.Loop() {
			_, _ = x.RoundToDenominator(den, fraction.RoundHalfToEven)
		}
	})
}

func BenchmarkWithDenominator(b *testing.B) {
	benchEach(b, benchWithDen, func(in benchInput) error {
		_, err := in.x.WithDenominator(in.den)
		return err
	}, func(b *testing.B, in benchInput) {
		x, den := in.x, in.den
		for b.Loop() {
			_, _ = x.WithDenominator(den)
		}
	})
}

// benchRats returns the big.Rat values of the operands, created before the loop.
func benchRats(in benchInput) (x, y *big.Rat) {
	return benchRat(in.x), benchRat(in.y)
}

func BenchmarkRatAdd(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := benchRats(in)
		for b.Loop() {
			new(big.Rat).Add(x, y)
		}
	})
}

func BenchmarkRatSub(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := benchRats(in)
		for b.Loop() {
			new(big.Rat).Sub(x, y)
		}
	})
}

func BenchmarkRatMul(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := benchRats(in)
		for b.Loop() {
			new(big.Rat).Mul(x, y)
		}
	})
}

func BenchmarkRatQuo(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := benchRats(in)
		for b.Loop() {
			new(big.Rat).Quo(x, y)
		}
	})
}

func BenchmarkRatCmp(b *testing.B) {
	benchEach(b, benchBinary, noCheck, func(b *testing.B, in benchInput) {
		x, y := benchRats(in)
		for b.Loop() {
			_ = x.Cmp(y)
		}
	})
}

// BenchmarkRatSetFrac reduces the same unreduced pair as BenchmarkReduce.
func BenchmarkRatSetFrac(b *testing.B) {
	benchEach(b, benchReduce, noCheck, func(b *testing.B, in benchInput) {
		num, den, ok := in.x.BigPair()
		require.True(b, ok)
		for b.Loop() {
			new(big.Rat).SetFrac(num, den)
		}
	})
}
