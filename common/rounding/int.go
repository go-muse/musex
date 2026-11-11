package rounding

import "math"

// QuoRem rounds the exact quotient n/d to q and returns r = n-q*d. A zero
// divisor returns ErrInvalidArgument. math.MinInt/-1 returns ErrOutOfRange
// because its exact quotient is not representable as int.
func (m Mode) QuoRem(n, d int) (q, r int, err error) {
	if err := validateMode(m); err != nil {
		return 0, 0, err
	}
	if d == 0 {
		return 0, 0, ErrInvalidArgument
	}
	if n == math.MinInt && d == -1 {
		return 0, 0, ErrOutOfRange
	}

	q0, r0 := n/d, n%d
	if r0 == 0 {
		return q0, 0, nil
	}

	sameSign := (n < 0) == (d < 0)
	side := Negative
	if sameSign {
		side = Positive
	}
	lower, lowerRemainder := q0, r0
	if !sameSign {
		lower--
		lowerRemainder += d
	}
	cmp := compareDistances(r0, d, sameSign)
	if m.choose(cmp, side, lower%2 == 0) == Upper {
		return lower + 1, lowerRemainder - d, nil
	}
	return lower, lowerRemainder, nil
}

// MustQuoRem calls QuoRem and panics with its error on failure.
func (m Mode) MustQuoRem(n, d int) (q, r int) {
	q, r, err := m.QuoRem(n, d)
	if err != nil {
		panic(err)
	}
	return q, r
}

// IntMultiple rounds n to a multiple of the positive step. It returns
// ErrInvalidArgument for step <= 0 and ErrOutOfRange if the selected multiple
// is not representable as int.
func (m Mode) IntMultiple(n, step int) (int, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if step <= 0 {
		return 0, ErrInvalidArgument
	}
	k, _, err := m.QuoRem(n, step)
	if err != nil {
		return 0, err
	}
	if k > math.MaxInt/step || k < math.MinInt/step {
		return 0, ErrOutOfRange
	}
	return k * step, nil
}

// MustIntMultiple calls IntMultiple and panics with its error on failure.
func (m Mode) MustIntMultiple(n, step int) int {
	v, err := m.IntMultiple(n, step)
	if err != nil {
		panic(err)
	}
	return v
}

func compareDistances(remainder, divisor int, positive bool) int {
	lower := absInt(remainder)
	divisorMagnitude := absInt(divisor)
	if !positive {
		lower = divisorMagnitude - lower
	}
	upper := divisorMagnitude - lower
	switch {
	case lower < upper:
		return -1
	case lower > upper:
		return 1
	default:
		return 0
	}
}

func absInt(v int) uint64 {
	if v < 0 {
		return uint64(-(v + 1)) + 1
	}
	return uint64(v)
}
