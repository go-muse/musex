package pitchclass_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// edgeSeeds are the corpus values of the plan: the limits of int, small
// numbers and the moduli of the criteria.
func edgeSeeds() []int {
	return []int{minInt, minInt + 1, -1, 0, 1, 2, 12, 19, maxInt - 1, maxInt}
}

// FuzzClassArithmetic checks creation by the circle, Tn, In, Mm and the
// interval operations against the math/big oracle for any x, y, n and N.
func FuzzClassArithmetic(f *testing.F) {
	for _, size := range edgeSeeds() {
		for _, v := range edgeSeeds() {
			f.Add(v, -v, v/2, size)
		}
	}
	f.Add(11, 2, 7, 12)
	f.Add(-13, 5, minInt, 12)

	f.Fuzz(func(t *testing.T, x, y, n, size int) {
		c, err := pitchclass.Reduce(x, size)
		if size < 1 {
			assertOnlyError(t, err, pitchclass.ErrInvalidClass, "Reduce")
			assert.Equal(t, pitchclass.Class{}, c)

			return
		}

		require.NoError(t, err)
		assertClass(t, mod(x, size), size, c, "Reduce")

		d := pitchclass.MustReduce(y, size)
		assertClass(t, addMod(x, n, size), size, c.Transpose(n), "Transpose")
		assertClass(t, subMod(n, x, size), size, c.Invert(n), "Invert")
		assertClass(t, mulMod(n, x, size), size, c.Multiply(n), "Multiply")

		i, err := c.IntervalTo(d)
		require.NoError(t, err)
		assert.Equal(t, subMod(y, x, size), i, "IntervalTo")

		ic, err := c.IntervalClass(d)
		require.NoError(t, err)
		assert.Equal(t, intervalClassOf(subMod(y, x, size), size), ic, "IntervalClass")

		s, err := c.IndexSum(d)
		require.NoError(t, err)
		assert.Equal(t, addMod(x, y, size), s, "IndexSum")

		around, err := c.InvertAround(d)
		require.NoError(t, err)
		assertClass(t, subMod(addMod(y, y, size), mod(x, size), size), size, around, "InvertAround")

		assert.Equal(t, mod(n, size), pitchclass.MustReduceInterval(n, size), "ReduceInterval")
		assert.Equal(t, negMod(n, size), pitchclass.MustComplementInterval(n, size), "ComplementInterval")
		assert.Equal(t, intervalClassOf(n, size), pitchclass.MustIntervalClass(n, size), "IntervalClass by size")

		back, err := pitchclass.Parse(c.String())
		require.NoError(t, err)
		assert.True(t, back.Equal(c), "Parse(String)")
	})
}

// FuzzOperator checks the creation, application, composition and inverse
// of operators against the oracle for any m, n, N and x.
func FuzzOperator(f *testing.F) {
	for _, size := range edgeSeeds() {
		for _, v := range edgeSeeds() {
			f.Add(v, v+1, size, -v)
		}
	}
	f.Add(7, 5, 12, 1)
	f.Add(2, 0, 12, 3)

	f.Fuzz(func(t *testing.T, m, n, size, x int) {
		op, err := pitchclass.NewOperator(m, n, size)
		switch {
		case size < 1:
			assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "NewOperator")
			assert.Equal(t, pitchclass.Operator{}, op)

			return
		case !invertibleOracle(m, size):
			assertOnlyError(t, err, pitchclass.ErrNotInvertible, "NewOperator")
			assert.Equal(t, pitchclass.Operator{}, op)

			return
		}

		require.NoError(t, err)
		spec := newOpSpec(m, n, size)
		assertOperator(t, spec, op, "NewOperator")

		c := pitchclass.MustReduce(x, size)
		got, err := op.Apply(c)
		require.NoError(t, err)
		assertClass(t, spec.apply(x), size, got, "Apply")

		g := pitchclass.MustTransposition(x, size)
		fg, err := op.Compose(g)
		require.NoError(t, err)
		assertOperator(t, spec.compose(transpositionSpec(x, size)), fg, "Compose")

		inv := op.Inverse()
		assertOperator(t, spec.inverse(), inv, "Inverse")
		id := pitchclass.MustIdentity(size)
		assert.True(t, op.MustCompose(inv).Equal(id), "f∘f^-1")
		assert.True(t, inv.MustCompose(op).Equal(id), "f^-1∘f")

		back, err := pitchclass.ParseOperator(op.String())
		require.NoError(t, err)
		assert.True(t, back.Equal(op), "ParseOperator(String)")
	})
}

// FuzzParse checks the canonical form of classes on any string: parsing
// never panics, a failure is ErrNotation with the zero Class, and an
// accepted string is the canonical form of its class.
func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"1 mod 12", "0 mod 1", "11 mod 12", "9223372036854775806 mod 9223372036854775807",
		"", "12 mod 12", "0 mod 0", "01 mod 12", "-1 mod 12", "1  mod 12", "1 MOD 12", "Class(invalid)",
		"0 mod 9223372036854775808", "\uff11 mod 12",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		var (
			c   pitchclass.Class
			err error
		)
		require.NotPanics(t, func() { c, err = pitchclass.Parse(s) })

		if err != nil {
			assertOnlyError(t, err, pitchclass.ErrNotation, "Parse")
			assert.Equal(t, pitchclass.Class{}, c)

			return
		}

		assert.True(t, c.IsValid())
		assert.Equal(t, s, c.String(), "an accepted string is the canonical form")
	})
}

// FuzzParseNotation checks the literary notation on any string, style and
// N: no panic; a failure is ErrInvalidModulus or ErrNotation; an accepted
// string is the notation of its class in that style.
func FuzzParseNotation(f *testing.F) {
	for _, s := range []string{"0", "1", "10", "11", "12", "T", "E", "t", "e", "A", "B", "01", "+1", "-0", "\u22120", " 1", ""} {
		for _, st := range []uint8{0, 1, 2, 3, 9} {
			for _, size := range []int{1, 10, 11, 12, 13, 19, 0, -1} {
				f.Add(s, st, size)
			}
		}
	}

	f.Fuzz(func(t *testing.T, s string, style uint8, size int) {
		st := pitchclass.Style(style)

		var (
			c   pitchclass.Class
			err error
		)
		require.NotPanics(t, func() { c, err = pitchclass.ParseNotation(s, st, size) })

		if err != nil {
			if size < 1 {
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "ParseNotation")
			} else {
				assertOnlyError(t, err, pitchclass.ErrNotation, "ParseNotation")
			}
			assert.Equal(t, pitchclass.Class{}, c)

			return
		}

		require.True(t, c.IsValid())
		assert.Equal(t, size, c.Modulus())

		back, err := c.Notation(st)
		require.NoError(t, err)
		assert.Equal(t, s, back, "an accepted string is the notation of its class")
	})
}

// FuzzParseOperator checks the canonical form of operators on any string:
// no panic; a failure is ErrNotation or ErrNotInvertible with the zero
// Operator; an accepted string is the canonical form of its operator.
func FuzzParseOperator(f *testing.F) {
	for _, s := range []string{
		"T5M7 mod 12", "T0M1 mod 12", "T0M0 mod 1", "T11M11 mod 12", "T0M2 mod 12", "T12M2 mod 12",
		"", "T5M7", "t5m7 mod 12", "T05M7 mod 12", "Operator(invalid)", "T0M1 mod 9223372036854775808",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		var (
			op  pitchclass.Operator
			err error
		)
		require.NotPanics(t, func() { op, err = pitchclass.ParseOperator(s) })

		if err != nil {
			if !assert.True(t, isOneOf(err, pitchclass.ErrNotation, pitchclass.ErrNotInvertible), "error %v", err) {
				return
			}
			assert.Equal(t, pitchclass.Operator{}, op)

			return
		}

		assert.True(t, op.IsValid())
		assert.Equal(t, s, op.String(), "an accepted string is the canonical form")
	})
}

// isOneOf reports whether err matches exactly one of the allowed sentinels
// and no other sentinel of the package.
func isOneOf(err error, allowed ...error) bool {
	for _, want := range allowed {
		if errorIsOnly(err, want) {
			return true
		}
	}

	return false
}

// errorIsOnly reports whether err matches want and no other sentinel.
func errorIsOnly(err, want error) bool {
	if !errors.Is(err, want) {
		return false
	}

	for _, e := range allErrors() {
		if e != want && errors.Is(err, e) {
			return false
		}
	}

	return true
}

// totientTrial returns phi(n) by trial division; only for n <= 2^40.
func totientTrial(n int) int {
	result, m := n, n
	for p := 2; p*p <= m; p++ {
		if m%p != 0 {
			continue
		}

		for m%p == 0 {
			m /= p
		}
		result -= result / p
	}

	if m > 1 {
		result -= result / m
	}

	return result
}

// FuzzTotient checks phi(N) for any N: exactly by trial division up to
// 2^40, N - 1 for a prime N, and general bounds otherwise.
func FuzzTotient(f *testing.F) {
	for _, n := range edgeSeeds() {
		f.Add(n)
	}
	for _, tc := range hardTotients() {
		f.Add(tc.n)
	}
	f.Add(1 << 40)
	f.Add(1<<40 + 1)

	f.Fuzz(func(t *testing.T, n int) {
		phi, err := pitchclass.Totient(n)
		if n < 1 {
			assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "Totient")
			assert.Zero(t, phi)

			return
		}

		require.NoError(t, err)

		switch {
		case n <= 1<<40:
			assert.Equal(t, totientTrial(n), phi, "Totient(%d)", n)
		case big.NewInt(int64(n)).ProbablyPrime(20):
			assert.Equal(t, n-1, phi, "Totient of the prime %d", n)
		default:
			assert.True(t, phi >= 1 && phi < n-1, "Totient(%d) = %d of a composite", n, phi)
			assert.Zero(t, phi%2, "Totient(%d) = %d is even", n, phi)
		}
	})
}
