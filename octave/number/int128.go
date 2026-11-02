package number

import "math/bits"

// int128 is a signed 128-bit integer in two's complement form, used for exact
// intermediate arithmetic on octave numbers without overflow and without
// allocation. Its value is hi*2^64 + lo.
//
// The magnitudes the package computes stay far below 2^126, so additions and
// subtractions of int128 values never overflow.
type int128 struct {
	hi int64
	lo uint64
}

// one is the int128 value 1.
var one = int128{lo: 1}

// int128Of returns v as an int128.
func int128Of(v int) int128 {
	return int128{hi: int64(v) >> 63, lo: uint64(v)}
}

// add returns a + b.
func (a int128) add(b int128) int128 {
	lo, carry := bits.Add64(a.lo, b.lo, 0)

	return int128{hi: a.hi + b.hi + int64(carry), lo: lo}
}

// sub returns a - b.
func (a int128) sub(b int128) int128 {
	lo, borrow := bits.Sub64(a.lo, b.lo, 0)

	return int128{hi: a.hi - b.hi - int64(borrow), lo: lo}
}

// cmp returns -1, 0, or +1 depending on whether a is less than, equal to, or
// greater than b.
func (a int128) cmp(b int128) int {
	switch {
	case a.hi < b.hi:
		return -1
	case a.hi > b.hi:
		return 1
	case a.lo < b.lo:
		return -1
	case a.lo > b.lo:
		return 1
	default:
		return 0
	}
}

// toInt returns a as an int and reports whether a fits in an int.
func (a int128) toInt() (int, bool) {
	v := int64(a.lo)
	if a.hi != v>>63 {
		return 0, false
	}

	return int(v), true
}
