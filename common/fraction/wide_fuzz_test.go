package fraction_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
)

// wFuzzValue builds a non-negative value from bytes, truncated to at most bits bits.
func wFuzzValue(b []byte, bits int) *big.Int {
	v := new(big.Int).SetBytes(b)
	if bits <= 0 {
		return v.SetInt64(0)
	}
	if v.BitLen() > bits {
		v.And(v, new(big.Int).Sub(wPow2(uint(bits)), big.NewInt(1)))
	}
	return v
}

// wFuzzOperands builds x = (±g·a1, ±g·b1) and y = (±g·a2, ±g·b2) with a common
// factor g of 1 to 128 bits and components of at most limit bits (128 or 255).
func wFuzzOperands(gb, a1b, b1b, a2b, b2b []byte, signs uint8, wide bool) (x, y fraction.Fraction) {
	limit := 128
	if wide {
		limit = fraction.MaxComponentBits
	}
	g := wFuzzValue(gb, 128)
	if g.Sign() == 0 {
		g.SetInt64(1)
	}
	rest := limit - g.BitLen()
	component := func(b []byte, nonzero bool, negative bool) *big.Int {
		v := wFuzzValue(b, rest)
		if nonzero && v.Sign() == 0 {
			v.SetInt64(1)
		}
		v.Mul(v, g)
		if negative {
			v.Neg(v)
		}
		return v
	}
	x = wFrac(component(a1b, false, signs&1 != 0), component(b1b, true, signs&2 != 0))
	y = wFrac(component(a2b, false, signs&4 != 0), component(b2b, true, signs&8 != 0))
	return x, y
}

// FuzzWideOperations compares every operation of the wide-value list with the
// contract oracle on operands with a common factor (architecture of task 26.6,
// "Accuracy tests"). Plain go test runs only the seed corpus below.
func FuzzWideOperations(f *testing.F) {
	b := func(v *big.Int) []byte { return v.Bytes() }
	p := wPow2
	m1 := func(e uint) []byte { return b(wOff(p(e), -1)) }
	seeds := []struct {
		g, a1, b1, a2, b2 []byte
		signs             uint8
		wide              bool
	}{
		{[]byte{1}, []byte{1}, []byte{3}, []byte{1}, []byte{6}, 0, false},
		{[]byte{6}, []byte{7}, []byte{5}, []byte{11}, []byte{13}, 0b0110, false},
		{b(wOff(p(64), 13)), m1(63), b(p(63)), b(wOff(p(63), 1)), m1(64), 0b1001, false},
		{b(p(32)), m1(32), b(wOff(p(32), 1)), b(p(31)), m1(96), 0b0101, false},
		{b(wOff(p(127), 1)), b(wF64a), b(wF64b), m1(64), b(big.NewInt(3)), 0b0010, true},
		{b(wOff(p(64), 1)), m1(190), m1(128), b(wOff(p(128), 1)), b(p(127)), 0b1100, true},
		{b(wOff(p(100), 7)), b(wKnuthU), b(wKnuthV), b(wKnuthV), b(wKnuthU), 0b0011, true},
		{m1(128), m1(127), m1(126), b(p(126)), b(wOff(p(126), 1)), 0b1111, true},
		{[]byte{1}, b(fraction.MaxComponentBig()), b(fraction.MaxComponentBig()), m1(255), []byte{1}, 0b0100, true},
		{[]byte{2}, nil, b(p(200)), b(p(200)), b(wOff(p(64), 1)), 0, true},
	}
	for _, s := range seeds {
		f.Add(s.g, s.a1, s.b1, s.a2, s.b2, s.signs, s.wide)
	}
	f.Fuzz(func(t *testing.T, gb, a1b, b1b, a2b, b2b []byte, signs uint8, wide bool) {
		x, y := wFuzzOperands(gb, a1b, b1b, a2b, b2b, signs, wide)
		checkWideBinary(t, x, y)
		checkWideUnary(t, x)
	})
}
