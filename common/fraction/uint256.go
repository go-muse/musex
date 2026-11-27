package fraction

import (
	"math"
	"math/big"
	"math/bits"
)

// u256 is an unsigned 256-bit integer stored in little-endian 64-bit limbs
// (u[0] is least significant, u[3] is most significant).
// Valid Fraction component magnitudes satisfy u <= MaxComponent (2^255 - 1),
// so bit 63 of u[3] is always 0 for stored components.
type u256 [4]uint64

// u512 is an unsigned 512-bit integer stored in little-endian 64-bit limbs.
// Since MaxComponent < 2^255, any product of two components is < 2^510 and any
// sum of two such products is < 2^511, which always fits in u512 without
// overflow.
type u512 [8]uint64

var (
	u256Zero = u256{0, 0, 0, 0}
	u256One  = u256{1, 0, 0, 0}
	u256Max  = u256{^uint64(0), ^uint64(0), ^uint64(0), (uint64(1) << 63) - 1}
)

func (u u256) isZero() bool {
	return (u[0] | u[1] | u[2] | u[3]) == 0
}

func (u u256) isOne() bool {
	return u[0] == 1 && (u[1]|u[2]|u[3]) == 0
}

func (u u256) isUint64() bool {
	return (u[1] | u[2] | u[3]) == 0
}

func (u u256) exceedsMax() bool {
	return (u[3] >> 63) != 0
}

func u256FromInt(n int) (u256, bool) {
	if n < 0 {
		mag := uint64(-(n + 1)) + 1
		return u256{mag, 0, 0, 0}, true
	}
	return u256{uint64(n), 0, 0, 0}, false
}

func u256ToInt(u u256, neg bool) (int, bool) {
	if !u.isUint64() {
		return 0, false
	}
	mag := u[0]
	if !neg {
		if mag > uint64(math.MaxInt) {
			return 0, false
		}
		return int(mag), true
	}
	const minIntMag = uint64(1) << 63
	if mag > minIntMag {
		return 0, false
	}
	if mag == minIntMag {
		return math.MinInt, true
	}
	return -int(mag), true
}

func putUint64BE(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func getUint64BE(b []byte) uint64 {
	return uint64(b[0])<<56 |
		uint64(b[1])<<48 |
		uint64(b[2])<<40 |
		uint64(b[3])<<32 |
		uint64(b[4])<<24 |
		uint64(b[5])<<16 |
		uint64(b[6])<<8 |
		uint64(b[7])
}

func (u u256) toBigInt(neg bool) *big.Int {
	if u.isZero() {
		return new(big.Int)
	}
	if u.isUint64() {
		z := new(big.Int).SetUint64(u[0])
		if neg {
			z.Neg(z)
		}
		return z
	}
	var buf [32]byte
	putUint64BE(buf[0:8], u[3])
	putUint64BE(buf[8:16], u[2])
	putUint64BE(buf[16:24], u[1])
	putUint64BE(buf[24:32], u[0])
	z := new(big.Int).SetBytes(buf[:])
	if neg {
		z.Neg(z)
	}
	return z
}

func u256FromBigInt(b *big.Int) (u256, bool, bool) {
	sig := b.Sign()
	if sig == 0 {
		return u256Zero, false, true
	}
	if b.BitLen() > MaxComponentBits {
		return u256Zero, false, false
	}
	var buf [32]byte
	b.FillBytes(buf[:])
	u := u256{
		getUint64BE(buf[24:32]),
		getUint64BE(buf[16:24]),
		getUint64BE(buf[8:16]),
		getUint64BE(buf[0:8]),
	}
	return u, sig < 0, true
}

func u256Cmp(a, b u256) int {
	for i := 3; i >= 0; i-- {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func cmpSignedU256(a u256, aNeg bool, b u256, bNeg bool) int {
	if a.isZero() {
		aNeg = false
	}
	if b.isZero() {
		bNeg = false
	}
	if aNeg != bNeg {
		if aNeg {
			return -1
		}
		return 1
	}
	c := u256Cmp(a, b)
	if aNeg {
		return -c
	}
	return c
}

func u256Add(a, b u256) (u256, uint64) {
	var res u256
	var carry uint64
	res[0], carry = bits.Add64(a[0], b[0], 0)
	res[1], carry = bits.Add64(a[1], b[1], carry)
	res[2], carry = bits.Add64(a[2], b[2], carry)
	res[3], carry = bits.Add64(a[3], b[3], carry)
	return res, carry
}

func u256Sub(a, b u256) u256 {
	var res u256
	var borrow uint64
	res[0], borrow = bits.Sub64(a[0], b[0], 0)
	res[1], borrow = bits.Sub64(a[1], b[1], borrow)
	res[2], borrow = bits.Sub64(a[2], b[2], borrow)
	res[3], _ = bits.Sub64(a[3], b[3], borrow)
	return res
}

// gcdUint64 returns gcd(a, b), with gcd(x, 0) = x (Euclid's algorithm).
func gcdUint64(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// u256GCD returns gcd(a, b), with gcd(x, 0) = x.
func u256GCD(a, b u256) u256 {
	if (a[2] | a[3] | b[2] | b[3]) == 0 {
		g0, g1 := gcd2(a[0], a[1], b[0], b[1])
		return u256{g0, g1, 0, 0}
	}
	return u256GCDWide(a, b)
}

// u256GCDWide is u256GCD for operands of which at least one is
// longer than two words.
func u256GCDWide(a, b u256) u256 {
	var g u256
	natGCD(g[:], a[:], b[:])
	return g
}

// u256DivMod returns a / b and a mod b for b != 0.
func u256DivMod(a, b u256) (quo, rem u256) {
	if (a[2] | a[3] | b[2] | b[3]) == 0 {
		q0, q1, r0, r1 := divMod2(a[0], a[1], b[0], b[1])
		return u256{q0, q1, 0, 0}, u256{r0, r1, 0, 0}
	}
	return u256DivModWide(a, b)
}

// u256DivModWide is u256DivMod for operands of which at least one is
// longer than two words.
func u256DivModWide(a, b u256) (quo, rem u256) {
	natDivMod(quo[:], rem[:], a[:], b[:])
	return quo, rem
}

// u256Mul returns the full 512-bit product a·b.
func u256Mul(a, b u256) u512 {
	if a.isUint64() && b.isUint64() {
		hi, lo := bits.Mul64(a[0], b[0])
		return u512{lo, hi, 0, 0, 0, 0, 0, 0}
	}
	if (a[2] | a[3] | b[2] | b[3]) == 0 {
		return mul2x2(a[0], a[1], b[0], b[1])
	}
	// Row by row, skipping zero words of a.
	var res u512
	for i := range 4 {
		ai := a[i]
		if ai == 0 {
			continue
		}
		// row = ai·b in five words; res[i+4] is still zero here.
		h0, l0 := bits.Mul64(ai, b[0])
		h1, l1 := bits.Mul64(ai, b[1])
		h2, l2 := bits.Mul64(ai, b[2])
		h3, l3 := bits.Mul64(ai, b[3])
		var c uint64
		r1, c := bits.Add64(l1, h0, 0)
		r2, c := bits.Add64(l2, h1, c)
		r3, c := bits.Add64(l3, h2, c)
		r4 := h3 + c
		res[i], c = bits.Add64(res[i], l0, 0)
		res[i+1], c = bits.Add64(res[i+1], r1, c)
		res[i+2], c = bits.Add64(res[i+2], r2, c)
		res[i+3], c = bits.Add64(res[i+3], r3, c)
		res[i+4] = r4 + c
	}
	return res
}

// mulAcc adds a·b to the three-word column accumulator c2:c1:c0.
func mulAcc(a, b, c0, c1, c2 uint64) (uint64, uint64, uint64) {
	h, l := bits.Mul64(a, b)
	var c uint64
	c0, c = bits.Add64(c0, l, 0)
	c1, c = bits.Add64(c1, h, c)
	return c0, c1, c2 + c
}

// mul2x2 returns (a1:a0)·(b1:b0) by columns.
func mul2x2(a0, a1, b0, b1 uint64) u512 {
	var r u512
	h, l := bits.Mul64(a0, b0)
	r[0] = l
	c0, c1, c2 := h, uint64(0), uint64(0)
	c0, c1, c2 = mulAcc(a0, b1, c0, c1, c2)
	c0, c1, c2 = mulAcc(a1, b0, c0, c1, c2)
	r[1] = c0
	c0, c1, _ = mulAcc(a1, b1, c1, c2, 0)
	r[2], r[3] = c0, c1
	return r
}

func (u u512) isZero() bool {
	return (u[0] | u[1] | u[2] | u[3] | u[4] | u[5] | u[6] | u[7]) == 0
}

func (u u512) isOne() bool {
	return u[0] == 1 && (u[1]|u[2]|u[3]|u[4]|u[5]|u[6]|u[7]) == 0
}

func (u u512) fitsU256Max() bool {
	return (u[4]|u[5]|u[6]|u[7]) == 0 && (u[3]>>63) == 0
}

func (u u512) low256() u256 {
	return u256{u[0], u[1], u[2], u[3]}
}

func u512FromU256(u u256) u512 {
	return u512{u[0], u[1], u[2], u[3], 0, 0, 0, 0}
}

func u512Cmp(a, b u512) int {
	for i := 7; i >= 0; i-- {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func u512Add(a, b u512) u512 {
	var res u512
	var carry uint64
	for i := range 8 {
		res[i], carry = bits.Add64(a[i], b[i], carry)
	}
	return res
}

func u512Sub(a, b u512) u512 {
	var res u512
	var borrow uint64
	for i := range 8 {
		res[i], borrow = bits.Sub64(a[i], b[i], borrow)
	}
	return res
}

// u512GCD returns gcd(a, b), with gcd(x, 0) = x.
func u512GCD(a, b u512) u512 {
	if (a[2] | a[3] | a[4] | a[5] | a[6] | a[7] | b[2] | b[3] | b[4] | b[5] | b[6] | b[7]) == 0 {
		g0, g1 := gcd2(a[0], a[1], b[0], b[1])
		return u512{g0, g1, 0, 0, 0, 0, 0, 0}
	}
	return u512GCDWide(a, b)
}

// u512GCDWide is u512GCD for operands of which at least one is
// longer than two words.
func u512GCDWide(a, b u512) u512 {
	var g u512
	natGCD(g[:], a[:], b[:])
	return g
}

// u512DivMod returns a / b and a mod b for b != 0.
func u512DivMod(a, b u512) (quo, rem u512) {
	if (a[2] | a[3] | a[4] | a[5] | a[6] | a[7] | b[2] | b[3] | b[4] | b[5] | b[6] | b[7]) == 0 {
		q0, q1, r0, r1 := divMod2(a[0], a[1], b[0], b[1])
		return u512{q0, q1, 0, 0, 0, 0, 0, 0}, u512{r0, r1, 0, 0, 0, 0, 0, 0}
	}
	return u512DivModWide(a, b)
}

// u512DivModWide is u512DivMod for operands of which at least one is
// longer than two words.
func u512DivModWide(a, b u512) (quo, rem u512) {
	natDivMod(quo[:], rem[:], a[:], b[:])
	return quo, rem
}

// setBigInt sets x to the signed value (u, neg) and reuses the capacity of x:
// one-word magnitudes go through SetUint64, longer ones through SetBytes from a
// stack buffer.
func (u u256) setBigInt(x *big.Int, neg bool) {
	if u.isUint64() {
		x.SetUint64(u[0])
	} else {
		var buf [32]byte
		putUint64BE(buf[0:8], u[3])
		putUint64BE(buf[8:16], u[2])
		putUint64BE(buf[16:24], u[1])
		putUint64BE(buf[24:32], u[0])
		x.SetBytes(buf[:])
	}
	if neg {
		x.Neg(x)
	}
}

// u512MulU256 returns a·b and true if the product fits in 512 bits; otherwise
// it returns an unspecified value and false.
func u512MulU256(a u512, b u256) (u512, bool) {
	var res [12]uint64
	for i := range 4 {
		bi := b[i]
		if bi == 0 {
			continue
		}
		var carry uint64
		for j := range 8 {
			hi, lo := bits.Mul64(a[j], bi)
			var c uint64
			lo, c = bits.Add64(lo, res[i+j], 0)
			hi += c
			lo, c = bits.Add64(lo, carry, 0)
			hi += c
			res[i+j] = lo
			carry = hi
		}
		res[i+8] = carry
	}
	if (res[8] | res[9] | res[10] | res[11]) != 0 {
		return u512{}, false
	}
	return u512(res[:8]), true
}
