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

func (u u256) bitLen() int {
	for i := 3; i >= 0; i-- {
		if u[i] != 0 {
			return i*64 + bits.Len64(u[i])
		}
	}
	return 0
}

func (u u256) trailingZeros() int {
	for i := range 4 {
		if u[i] != 0 {
			return i*64 + bits.TrailingZeros64(u[i])
		}
	}
	return 256
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

func u256Shr(u u256, k int) u256 {
	if k <= 0 {
		return u
	}
	if k >= 256 {
		return u256Zero
	}
	wordShift := k / 64
	bitShift := uint(k % 64)
	var res u256
	for i := 0; i < 4-wordShift; i++ {
		res[i] = u[i+wordShift] >> bitShift
		if bitShift != 0 && i+wordShift+1 < 4 {
			res[i] |= u[i+wordShift+1] << (64 - bitShift)
		}
	}
	return res
}

func u256Shr1(u u256) u256 {
	return u256{
		(u[0] >> 1) | (u[1] << 63),
		(u[1] >> 1) | (u[2] << 63),
		(u[2] >> 1) | (u[3] << 63),
		u[3] >> 1,
	}
}

func u256Shl(u u256, k int) u256 {
	if k <= 0 {
		return u
	}
	if k >= 256 {
		return u256Zero
	}
	wordShift := k / 64
	bitShift := uint(k % 64)
	var res u256
	for i := wordShift; i < 4; i++ {
		res[i] = u[i-wordShift] << bitShift
		if bitShift != 0 && i-wordShift-1 >= 0 {
			res[i] |= u[i-wordShift-1] >> (64 - bitShift)
		}
	}
	return res
}

func gcdUint64(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func u256GCD(a, b u256) u256 {
	if a.isUint64() && b.isUint64() {
		return u256{gcdUint64(a[0], b[0]), 0, 0, 0}
	}
	if a.isZero() {
		return b
	}
	if b.isZero() {
		return a
	}
	za := a.trailingZeros()
	zb := b.trailingZeros()
	shift := min(za, zb)
	a = u256Shr(a, za)
	for !b.isZero() {
		b = u256Shr(b, b.trailingZeros())
		if u256Cmp(a, b) > 0 {
			a, b = b, a
		}
		b = u256Sub(b, a)
	}
	return u256Shl(a, shift)
}

func u256DivMod(a, b u256) (quo, rem u256) {
	if a.isUint64() && b.isUint64() {
		return u256{a[0] / b[0], 0, 0, 0}, u256{a[0] % b[0], 0, 0, 0}
	}
	if u256Cmp(a, b) < 0 {
		return u256Zero, a
	}
	shift := a.bitLen() - b.bitLen()
	divisor := u256Shl(b, shift)
	rem = a
	for i := shift; i >= 0; i-- {
		if u256Cmp(rem, divisor) >= 0 {
			rem = u256Sub(rem, divisor)
			quo[i/64] |= uint64(1) << uint(i%64)
		}
		divisor = u256Shr1(divisor)
	}
	return quo, rem
}

func u256Mul(a, b u256) u512 {
	if a.isUint64() && b.isUint64() {
		hi, lo := bits.Mul64(a[0], b[0])
		return u512{lo, hi, 0, 0, 0, 0, 0, 0}
	}
	var res u512
	for i := range 4 {
		if a[i] == 0 {
			continue
		}
		var carry uint64
		for j := range 4 {
			hi, lo := bits.Mul64(a[i], b[j])
			var c1, c2 uint64
			res[i+j], c1 = bits.Add64(res[i+j], lo, 0)
			res[i+j], c2 = bits.Add64(res[i+j], carry, 0)
			carry = hi + c1 + c2
		}
		res[i+4] = carry
	}
	return res
}

func (u u512) isZero() bool {
	return (u[0] | u[1] | u[2] | u[3] | u[4] | u[5] | u[6] | u[7]) == 0
}

func (u u512) isOne() bool {
	return u[0] == 1 && (u[1]|u[2]|u[3]|u[4]|u[5]|u[6]|u[7]) == 0
}

func (u u512) isUint64() bool {
	return (u[1] | u[2] | u[3] | u[4] | u[5] | u[6] | u[7]) == 0
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

func (u u512) bitLen() int {
	for i := 7; i >= 0; i-- {
		if u[i] != 0 {
			return i*64 + bits.Len64(u[i])
		}
	}
	return 0
}

func (u u512) trailingZeros() int {
	for i := range 8 {
		if u[i] != 0 {
			return i*64 + bits.TrailingZeros64(u[i])
		}
	}
	return 512
}

func u512Shr(u u512, k int) u512 {
	if k <= 0 {
		return u
	}
	if k >= 512 {
		return u512{}
	}
	wordShift := k / 64
	bitShift := uint(k % 64)
	var res u512
	for i := 0; i < 8-wordShift; i++ {
		res[i] = u[i+wordShift] >> bitShift
		if bitShift != 0 && i+wordShift+1 < 8 {
			res[i] |= u[i+wordShift+1] << (64 - bitShift)
		}
	}
	return res
}

func u512Shr1(u u512) u512 {
	return u512{
		(u[0] >> 1) | (u[1] << 63),
		(u[1] >> 1) | (u[2] << 63),
		(u[2] >> 1) | (u[3] << 63),
		(u[3] >> 1) | (u[4] << 63),
		(u[4] >> 1) | (u[5] << 63),
		(u[5] >> 1) | (u[6] << 63),
		(u[6] >> 1) | (u[7] << 63),
		u[7] >> 1,
	}
}

func u512Shl(u u512, k int) u512 {
	if k <= 0 {
		return u
	}
	if k >= 512 {
		return u512{}
	}
	wordShift := k / 64
	bitShift := uint(k % 64)
	var res u512
	for i := wordShift; i < 8; i++ {
		res[i] = u[i-wordShift] << bitShift
		if bitShift != 0 && i-wordShift-1 >= 0 {
			res[i] |= u[i-wordShift-1] >> (64 - bitShift)
		}
	}
	return res
}

func u512GCD(a, b u512) u512 {
	if a.isUint64() && b.isUint64() {
		return u512{gcdUint64(a[0], b[0]), 0, 0, 0, 0, 0, 0, 0}
	}
	if a.isZero() {
		return b
	}
	if b.isZero() {
		return a
	}
	za := a.trailingZeros()
	zb := b.trailingZeros()
	shift := min(za, zb)
	a = u512Shr(a, za)
	for !b.isZero() {
		b = u512Shr(b, b.trailingZeros())
		if u512Cmp(a, b) > 0 {
			a, b = b, a
		}
		b = u512Sub(b, a)
	}
	return u512Shl(a, shift)
}

func u512DivMod(a, b u512) (quo, rem u512) {
	if a.isUint64() && b.isUint64() {
		return u512{a[0] / b[0]}, u512{a[0] % b[0]}
	}
	if u512Cmp(a, b) < 0 {
		return u512{}, a
	}
	shift := a.bitLen() - b.bitLen()
	divisor := u512Shl(b, shift)
	rem = a
	for i := shift; i >= 0; i-- {
		if u512Cmp(rem, divisor) >= 0 {
			rem = u512Sub(rem, divisor)
			quo[i/64] |= uint64(1) << uint(i%64)
		}
		divisor = u512Shr1(divisor)
	}
	return quo, rem
}
