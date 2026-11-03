package pitchclass

// ReduceInterval returns the ordered interval of size i modulo modulus:
// i mod N, in the range 0 to N-1. Any int i is accepted: ReduceInterval(14,
// 12) is 2, and ReduceInterval(-1, 12) is 11.
//
// If modulus is less than 1, ReduceInterval returns 0 and an error that
// wraps [ErrInvalidModulus].
func ReduceInterval(i, modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	return mod(i, modulus), nil
}

// MustReduceInterval is like [ReduceInterval] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func MustReduceInterval(i, modulus int) int {
	r, err := ReduceInterval(i, modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// ComplementInterval returns the complementary interval of an interval of
// size i modulo modulus: (-i) mod N, in the range 0 to N-1. Together with
// i mod N it completes the circle: ComplementInterval(4, 12) is 8, and
// ComplementInterval(0, 12) is 0. Any int i is accepted, including
// math.MinInt.
//
// If modulus is less than 1, ComplementInterval returns 0 and an error that
// wraps [ErrInvalidModulus].
func ComplementInterval(i, modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	return negMod(mod(i, modulus), modulus), nil
}

// MustComplementInterval is like [ComplementInterval] but panics with its
// error instead of returning it. It is intended for values known to be
// valid.
func MustComplementInterval(i, modulus int) int {
	r, err := ComplementInterval(i, modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// IntervalClass returns the interval class of an interval of size i modulo
// modulus: the smaller of i mod N and (-i) mod N, in the range 0 to N/2
// (rounded down). An interval and its complement have the same interval
// class: IntervalClass(7, 12) and IntervalClass(5, 12) are both 5. Any int
// i is accepted.
//
// IntervalClass(i, N) equals the interval class of 0 mod N and i mod N; for
// the interval class of two classes, use the method [Class.IntervalClass].
//
// If modulus is less than 1, IntervalClass returns 0 and an error that
// wraps [ErrInvalidModulus].
func IntervalClass(i, modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	r := mod(i, modulus)

	return min(r, negMod(r, modulus)), nil
}

// MustIntervalClass is like [IntervalClass] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func MustIntervalClass(i, modulus int) int {
	r, err := IntervalClass(i, modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// InversionFixedClasses returns the classes z of modulus positions that the
// inversion In with n = index leaves in place, In(z) = z, in ascending order
// of their numbers. Any int index is accepted. There are 0, 1, or 2 such
// classes:
//
//   - for even N and even n mod N, the classes (n mod N)/2 and
//     (n mod N)/2 + N/2;
//   - for even N and odd n mod N, none;
//   - for odd N, exactly one class.
//
// For example, I4 mod 12 leaves 2 mod 12 and 8 mod 12 in place, and I1 mod
// 12 leaves no class in place.
//
// An empty result is an answer, not an error: it is an empty, non-nil
// slice. Each call returns a new slice whose length and capacity equal the
// number of classes; the caller owns it and may modify or retain it.
//
// If modulus is less than 1, InversionFixedClasses returns nil and an error
// that wraps [ErrInvalidModulus].
func InversionFixedClasses(index, modulus int) ([]Class, error) {
	if modulus < 1 {
		return nil, errModulus(modulus)
	}

	n := mod(index, modulus)

	if modulus%2 == 0 {
		if n%2 != 0 {
			return []Class{}, nil
		}

		return []Class{{k: n / 2, n: modulus}, {k: n/2 + modulus/2, n: modulus}}, nil
	}

	// For odd N, the class is n * 2⁻¹ mod N. For odd n it equals
	// (n + N)/2, computed as (n-1)/2 + N/2 + 1 to avoid overflow.
	k := n / 2
	if n%2 != 0 {
		k = (n-1)/2 + modulus/2 + 1
	}

	return []Class{{k: k, n: modulus}}, nil
}

// MustInversionFixedClasses is like [InversionFixedClasses] but panics with
// its error instead of returning it. It is intended for values known to be
// valid.
func MustInversionFixedClasses(index, modulus int) []Class {
	r, err := InversionFixedClasses(index, modulus)
	if err != nil {
		panic(err)
	}

	return r
}
