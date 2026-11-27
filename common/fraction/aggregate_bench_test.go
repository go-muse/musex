package fraction_test

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/require"
)

// Benchmarks of task 26.7 (architecture, requirement T-3 and "Tests",
// "Benchmarks"). Names are stable for comparison:
//
//   - Benchmark<Function>/<class>/<n> and BenchmarkRef<Function>/<class>/<n>
//     for Sum, Product, CommonDenominator and CommonDenominatorBig on sets of
//     n = 2, 8, 64 elements;
//   - Benchmark<Function>/<class>/<n>-overflow for Sum and Product on the
//     path of the exact continuation, with the references on the same sets,
//     for the report;
//   - BenchmarkGCD/<class>, BenchmarkLCM/<class>, BenchmarkRefLCM/<class>;
//   - BenchmarkRefBigRat/<class>, BenchmarkBigRatTo/<class>/nil and
//     BenchmarkBigRatTo/<class>/reuse.
//
// BenchmarkRef* run the computation the consumers use now, in exactly its form
// (aggregate_bench_check_test.go). Every operation is checked against its
// reference once before the loop.

// agBenchSets runs body for every set as the sub-benchmark <class>/<n> after
// check; classes keep the order of agClasses.
func agBenchSets(b *testing.B, sets []agSetInput, check func(fs []fraction.Fraction) error, body func(b *testing.B, fs []fraction.Fraction)) {
	for _, class := range agClasses {
		b.Run(class, func(b *testing.B) {
			for _, s := range sets {
				if s.class != class {
					continue
				}
				b.Run(strings.TrimPrefix(s.name(), class+"/"), func(b *testing.B) {
					require.NoError(b, check(s.fs))
					b.ReportAllocs()
					body(b, s.fs)
				})
			}
		})
	}
}

// agSameFrac returns nil when the function and the reference agree.
func agSameFrac(got fraction.Fraction, err error, ref fraction.Fraction, refErr error) error {
	switch {
	case err != nil:
		return err
	case refErr != nil:
		return refErr
	case got != ref:
		return fmt.Errorf("result %s, reference %s", got, ref)
	}
	return nil
}

func agCheckSum(fs []fraction.Fraction) error {
	got, err := fraction.Sum(fs...)
	ref, refErr := agRefSum(fs)
	return agSameFrac(got, err, ref, refErr)
}

func agCheckProduct(fs []fraction.Fraction) error {
	got, err := fraction.Product(fs...)
	ref, refErr := agRefProduct(fs)
	return agSameFrac(got, err, ref, refErr)
}

func agCheckCommonDen(fs []fraction.Fraction) error {
	got, err := fraction.CommonDenominator(fs...)
	if err != nil {
		return err
	}
	gotBig, err := fraction.CommonDenominatorBig(fs...)
	if err != nil {
		return err
	}
	ref, ok := agRefCommonDenominator(fs)
	if !ok || got != ref || gotBig.Cmp(agRefCommonDenominatorBig(fs)) != 0 {
		return fmt.Errorf("results %d and %s, reference %d", got, gotBig, ref)
	}
	return nil
}

func BenchmarkSum(b *testing.B) {
	agBenchSets(b, slices.Concat(agInputs().sum, agInputs().sumOverflow), agCheckSum, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = fraction.Sum(fs...)
		}
	})
}

func BenchmarkRefSum(b *testing.B) {
	agBenchSets(b, slices.Concat(agInputs().sum, agInputs().sumOverflow), agCheckSum, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = agRefSum(fs)
		}
	})
}

func BenchmarkProduct(b *testing.B) {
	agBenchSets(b, slices.Concat(agInputs().product, agInputs().prodOverflow), agCheckProduct, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = fraction.Product(fs...)
		}
	})
}

func BenchmarkRefProduct(b *testing.B) {
	agBenchSets(b, slices.Concat(agInputs().product, agInputs().prodOverflow), agCheckProduct, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = agRefProduct(fs)
		}
	})
}

func BenchmarkCommonDenominator(b *testing.B) {
	agBenchSets(b, agInputs().commonDen, agCheckCommonDen, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = fraction.CommonDenominator(fs...)
		}
	})
}

func BenchmarkRefCommonDenominator(b *testing.B) {
	agBenchSets(b, agInputs().commonDen, agCheckCommonDen, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = agRefCommonDenominator(fs)
		}
	})
}

func BenchmarkCommonDenominatorBig(b *testing.B) {
	agBenchSets(b, agInputs().commonDen, agCheckCommonDen, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_, _ = fraction.CommonDenominatorBig(fs...)
		}
	})
}

func BenchmarkRefCommonDenominatorBig(b *testing.B) {
	agBenchSets(b, agInputs().commonDen, agCheckCommonDen, func(b *testing.B, fs []fraction.Fraction) {
		for b.Loop() {
			_ = agRefCommonDenominatorBig(fs)
		}
	})
}

// agBenchPairs runs body for every GCD/LCM pair as the sub-benchmark <class> after check.
func agBenchPairs(b *testing.B, check func(p agPairInput) error, body func(b *testing.B, x, y fraction.Fraction)) {
	for _, p := range agInputs().pairs {
		b.Run(p.class, func(b *testing.B) {
			require.NoError(b, check(p))
			b.ReportAllocs()
			body(b, p.x, p.y)
		})
	}
}

func agCheckLCM(p agPairInput) error {
	got, err := p.x.LCM(p.y)
	ref, refErr := agRefLCM(p.x, p.y)
	return agSameFrac(got, err, ref, refErr)
}

// BenchmarkGCD is for the report: T-3 does not apply to GCD.
func BenchmarkGCD(b *testing.B) {
	agBenchPairs(b, func(p agPairInput) error {
		_, err := p.x.GCD(p.y)
		return err
	}, func(b *testing.B, x, y fraction.Fraction) {
		for b.Loop() {
			_, _ = x.GCD(y)
		}
	})
}

func BenchmarkLCM(b *testing.B) {
	agBenchPairs(b, agCheckLCM, func(b *testing.B, x, y fraction.Fraction) {
		for b.Loop() {
			_, _ = x.LCM(y)
		}
	})
}

func BenchmarkRefLCM(b *testing.B) {
	agBenchPairs(b, agCheckLCM, func(b *testing.B, x, y fraction.Fraction) {
		for b.Loop() {
			_, _ = agRefLCM(x, y)
		}
	})
}

// agCheckBigRatToInput returns nil when BigRatTo into z and BigRat agree.
func agCheckBigRatToInput(x fraction.Fraction, z *big.Rat) error {
	got, ok := x.BigRatTo(z)
	want, wantOK := x.BigRat()
	if !ok || !wantOK || got == nil || got.Cmp(want) != 0 {
		return errors.New("BigRatTo and BigRat disagree")
	}
	return nil
}

// BenchmarkRefBigRat is the reference of BigRatTo: a new *big.Rat per call.
func BenchmarkRefBigRat(b *testing.B) {
	for _, in := range agBigRatToInputs() {
		b.Run(in.class, func(b *testing.B) {
			x := in.x
			require.NoError(b, agCheckBigRatToInput(x, nil))
			b.ReportAllocs()
			for b.Loop() {
				_, _ = x.BigRat()
			}
		})
	}
}

// BenchmarkBigRatTo measures a new value per call (nil) and one reused z (reuse).
func BenchmarkBigRatTo(b *testing.B) {
	for _, in := range agBigRatToInputs() {
		b.Run(in.class, func(b *testing.B) {
			x := in.x
			b.Run("nil", func(b *testing.B) {
				require.NoError(b, agCheckBigRatToInput(x, nil))
				b.ReportAllocs()
				for b.Loop() {
					_, _ = x.BigRatTo(nil)
				}
			})
			b.Run("reuse", func(b *testing.B) {
				z := new(big.Rat)
				require.NoError(b, agCheckBigRatToInput(x, z))
				b.ReportAllocs()
				for b.Loop() {
					_, _ = x.BigRatTo(z)
				}
			})
		})
	}
}
