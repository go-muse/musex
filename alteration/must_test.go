package alteration_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
)

// K-21 (test part), K-17: on the inputs of K-2...K-13 every Must variant
// returns the result of its operation or panics with the same sentinel
// error; the inputs are unchanged after the call, the error and the panic.
// The boundary inputs of K-14 are covered by TestRange with the same check.
func TestMustMatchesOperation(t *testing.T) {
	t.Parallel()

	t.Run("FromInt", func(t *testing.T) {
		t.Parallel()

		var ops []fallibleOp
		for n := -10; n <= 10; n++ {
			ops = append(ops, opFromInt(n))
		}

		runFallible(t, ops)
	})

	t.Run("FromFraction", func(t *testing.T) {
		t.Parallel()

		// q == 0 gives the zero-denominator inputs of K-13.
		var ops []fallibleOp
		for p := -12; p <= 12; p++ {
			for q := -12; q <= 12; q++ {
				ops = append(ops, opFromFraction(p, q))
			}
		}

		runFallible(t, ops)
	})

	t.Run("Raise and Lower", func(t *testing.T) {
		t.Parallel()

		var ops []fallibleOp
		for _, a := range setS() {
			ops = append(ops, opRaise(a), opLower(a))
		}

		runFallible(t, ops)
	})

	t.Run("RaiseBy and LowerBy", func(t *testing.T) {
		t.Parallel()

		var ops []fallibleOp
		for _, a := range setS() {
			for n := -5; n <= 5; n++ {
				ops = append(ops, opRaiseBy(a, n), opLowerBy(a, n))
			}
		}

		runFallible(t, ops)
	})

	t.Run("Add and Sub", func(t *testing.T) {
		t.Parallel()

		var ops []fallibleOp
		for _, a := range setS() {
			for _, d := range setS() {
				ops = append(ops, opAdd(a, d), opSub(a, d))
			}
		}

		runFallible(t, ops)
	})

	t.Run("Mul", func(t *testing.T) {
		t.Parallel()

		var ops []fallibleOp
		for _, a := range setS() {
			for k := -4; k <= 4; k++ {
				ops = append(ops, opMul(a, k))
			}
		}

		runFallible(t, ops)
	})
}

// K-17: operations without errors do not change their inputs either: the
// receiver and the argument equal their copies taken before the call and keep
// their exact values.
func TestQueriesKeepInputs(t *testing.T) {
	t.Parallel()

	inputs := slices.Concat(setS(), []frac{edgeTop(), edgeBottom(), edgeInvQ(), edgeNegInvQ(), edgeNearTop(), edgeNearBottom()})

	for _, fa := range inputs {
		t.Run(fa.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, fa)
			aBefore := a

			_, _ = a.Fraction()
			_ = a.Sign()
			_ = a.IsNatural()
			_ = a.IsInteger()
			_ = a.String()

			for _, fb := range inputs {
				b := value(t, fb)
				bBefore := b

				_ = a.Equal(b)
				_ = a.Compare(b)

				assertInputsUnchanged(t, []frac{fa, fb}, []alteration.Alteration{aBefore, bBefore}, []alteration.Alteration{a, b},
					"Equal and Compare with "+fb.String())
			}

			assertInputsUnchanged(t, []frac{fa}, []alteration.Alteration{aBefore}, []alteration.Alteration{a}, "queries")
			assert.Equal(t, expectedFrac(t, fa.rat()), fractionOf(a))
		})
	}
}
