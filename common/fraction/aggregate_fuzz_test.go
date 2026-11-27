package fraction_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
)

// Fuzz shape bits of FuzzAggregate.
const (
	agFuzzCount      = 0x0f // number of base elements: shape & 0x0f, modulo 10
	agFuzzNegations  = 0x10 // append −xₙ₋₁, …, −x₁: the sum equals x₀
	agFuzzInverses   = 0x20 // append 1/xₙ₋₁, …, 1/x₁ of nonzero elements: the product equals x₀
	agFuzzInvalid    = 0x40 // insert the zero value Fraction{}
	agFuzzInvalidPos = 0x80 // insert it at the end instead of the start
)

// agFuzzSet builds a set from fuzz data: base elements (±gN·aᵢ·e, ±gD·bᵢ·e)
// with common factors gN of the numerators and gD of the denominators, the
// factor e (applied only where the bit 3i of signs is set: an unreduced pair)
// bounding the components as well, and components of at most 128 or 255 bits. Shape bits add mirrored elements, so
// prefixes leave the range while the result stays in it, and an invalid element.
func agFuzzSet(gn, gd, e, raw []byte, signs uint64, shape uint8, wide bool) []fraction.Fraction {
	limit := 128
	if wide {
		limit = fraction.MaxComponentBits
	}
	nonzero := func(v *big.Int) *big.Int {
		if v.Sign() == 0 {
			return v.SetInt64(1)
		}
		return v
	}
	gN := nonzero(wFuzzValue(gn, limit/4))
	gD := nonzero(wFuzzValue(gd, limit/4))
	ev := nonzero(wFuzzValue(e, limit/4))
	// bounded returns a value from bytes not above (2^limit − 1) / (g·ev), so
	// the component g·v·ev stays within limit bits and MaxComponent is reachable.
	top := wOff(wPow2(uint(limit)), -1)
	bounded := func(b []byte, g *big.Int) *big.Int {
		m := new(big.Int).Quo(top, wMul(g, ev))
		return new(big.Int).Mod(new(big.Int).SetBytes(b), m.Add(m, big.NewInt(1)))
	}
	n := int(shape&agFuzzCount) % 10
	var fs []fraction.Fraction
	if n > 0 {
		chunk := len(raw) / (2 * n)
		part := func(k int) []byte { return raw[k*chunk : (k+1)*chunk] }
		for i := range n {
			a := bounded(part(2*i), gN)
			b := nonzero(bounded(part(2*i+1), gD))
			num, den := wMul(gN, a), wMul(gD, b)
			bits := signs >> (3 * uint(i%21))
			if bits&1 != 0 {
				num.Mul(num, ev)
				den.Mul(den, ev)
			}
			if bits&2 != 0 {
				num.Neg(num)
			}
			if bits&4 != 0 {
				den.Neg(den)
			}
			fs = append(fs, wFrac(num, den))
		}
	}
	base := slices.Clone(fs)
	if shape&agFuzzNegations != 0 {
		for i := len(base) - 1; i >= 1; i-- {
			fs = append(fs, base[i].MustNeg())
		}
	}
	if shape&agFuzzInverses != 0 {
		for i := len(base) - 1; i >= 1; i-- {
			if !base[i].IsZero() {
				fs = append(fs, base[i].MustReciprocal())
			}
		}
	}
	if shape&agFuzzInvalid != 0 {
		if shape&agFuzzInvalidPos != 0 {
			fs = append(fs, fraction.Fraction{})
		} else {
			fs = slices.Insert(fs, 0, fraction.Fraction{})
		}
	}
	return fs
}

// FuzzAggregate compares every function of task 26.7 with the contract oracle
// on sets with common factors and components up to 255 bits (architecture,
// "Tests", "Fuzzing"): Sum, Product, CommonDenominator, CommonDenominatorBig
// and their Must forms on the set, GCD and LCM of its first two elements, and
// BigRatTo of every element into nil and into one reused z. Plain go test runs
// only the seed corpus below.
func FuzzAggregate(f *testing.F) {
	enc := func(vs ...*big.Int) []byte {
		var out []byte
		for _, v := range vs {
			out = append(out, v.FillBytes(make([]byte, 32))...)
		}
		return out
	}
	one := big.NewInt(1)
	pa, pb := bp(200, 0), bp(200, 1)
	maxC := fraction.MaxComponentBig()
	p := wPow2
	bs := func(v *big.Int) []byte { return v.Bytes() }
	seeds := []struct {
		gn, gd, e, raw []byte
		signs          uint64
		shape          uint8
		wide           bool
	}{
		{[]byte{1}, []byte{1}, []byte{6}, enc(big.NewInt(1), big.NewInt(4), big.NewInt(1), big.NewInt(6), big.NewInt(5), big.NewInt(12)), 0b001_010_001, 3, false},
		{[]byte{2}, []byte{3}, []byte{5}, enc(big.NewInt(7), big.NewInt(11), big.NewInt(0), big.NewInt(13)), 0b101_011, 2, false},
		{[]byte{1}, []byte{1}, []byte{1}, enc(one, pa, one, pb), 0, 2 | agFuzzNegations, true},
		{[]byte{1}, []byte{1}, []byte{1}, enc(p(200), one, p(200), one, one, p(200)), 0, 3, true},
		{[]byte{1}, []byte{1}, []byte{1}, enc(maxC, one, one, one), 0, 2 | agFuzzNegations, true},
		{bs(p(60)), bs(wOff(p(60), 1)), bs(wOff(p(30), 3)), enc(p(120), wOff(p(150), -1), wOff(p(100), 7), p(140), p(130), wOff(p(129), 1)), 0x1ff, 3 | agFuzzNegations | agFuzzInverses, true},
		{bs(wOff(p(64), 13)), bs(p(63)), bs(big.NewInt(3)), enc(wOff(p(63), -1), wOff(p(64), 1), p(40), wOff(p(40), -1)), 0b110_011, 2 | agFuzzInverses, false},
		{[]byte{1}, []byte{1}, []byte{1}, enc(one, big.NewInt(3), one, big.NewInt(9)), 0, 2 | agFuzzInvalid, false},
		{[]byte{1}, []byte{1}, []byte{1}, enc(one, pa, one, pb), 0, 2 | agFuzzInvalid | agFuzzInvalidPos, true},
		{nil, nil, nil, nil, 0, 0, false},
	}
	for _, s := range seeds {
		f.Add(s.gn, s.gd, s.e, s.raw, s.signs, s.shape, s.wide)
	}
	f.Fuzz(func(t *testing.T, gn, gd, e, raw []byte, signs uint64, shape uint8, wide bool) {
		fs := agFuzzSet(gn, gd, e, raw, signs, shape, wide)
		agCheckSumProduct(t, "set", fs)
		agCheckCommonDenominator(t, "set", fs)
		switch len(fs) {
		case 0:
		case 1:
			agCheckGCDLCM(t, "pair", fs[0], fs[0])
		default:
			agCheckGCDLCM(t, "pair", fs[0], fs[1])
		}
		z := big.NewRat(5, 7)
		for _, x := range fs {
			agCheckBigRatTo(t, "BigRatTo nil", x, nil)
			agCheckBigRatTo(t, "BigRatTo reused", x, z)
		}
	})
}
