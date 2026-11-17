package degree

import "iter"

// InCircle reports whether the number of d belongs to a circle of k numbers.
// A number above k returns false without an error. Errors are ErrUnset for an
// unset receiver and ErrInvalidNumber for k below one.
func (d Degree) InCircle(k int) (bool, error) {
	if !d.IsValid() {
		return false, unsetError("InCircle", "unset degree")
	}

	if k < 1 {
		return false, invalidNumberError("InCircle", "circle size is below one")
	}

	return d.number <= k, nil
}

// Next returns the next number-only degree in a circle of k numbers and
// whether the transition wrapped from k to 1. The note of d is not carried
// over. Errors are ErrUnset for an unset receiver and ErrInvalidNumber for an
// invalid k or a number above k.
func (d Degree) Next(k int) (next Degree, wrapped bool, err error) {
	if err := d.checkInside("Next", k); err != nil {
		return Degree{}, false, err
	}

	if d.number == k {
		return Degree{number: 1}, true, nil
	}

	return Degree{number: d.number + 1}, false, nil
}

// Previous returns the previous number-only degree in a circle of k numbers
// and whether the transition wrapped from 1 to k. The note of d is not carried
// over. Its errors are those of Next.
func (d Degree) Previous(k int) (previous Degree, wrapped bool, err error) {
	if err := d.checkInside("Previous", k); err != nil {
		return Degree{}, false, err
	}

	if d.number == 1 {
		return Degree{number: k}, true, nil
	}

	return Degree{number: d.number - 1}, false, nil
}

// Shift moves d by n positions in a circle of k numbers and returns the
// resulting number-only degree and the signed number of completed cycles. It
// is exact for every int n, including the minimum and maximum values, and does
// not change d or carry its note over.
//
// Errors are ErrUnset for an unset receiver and ErrInvalidNumber for an
// invalid k or a number above k.
func (d Degree) Shift(k, n int) (shifted Degree, cycles int, err error) {
	if err := d.checkInside("Shift", k); err != nil {
		return Degree{}, 0, err
	}

	// Decompose n into floor division and a nonnegative remainder. This avoids
	// evaluating (d.number-1)+n, which can overflow even when the result is
	// representable.
	quotient, remainder := n/k, n%k
	if remainder < 0 {
		quotient--
		remainder += k
	}

	position := d.number - 1
	cycles = quotient

	if remainder > 0 && position >= k-remainder {
		position -= k - remainder
		cycles++
	} else {
		position += remainder
	}

	return Degree{number: position + 1}, cycles, nil
}

// DistanceUp returns the distance from d upward to another degree in a circle
// of k numbers, in the range 0 through k-1. Neither degree changes, and notes
// do not take part.
//
// Errors, in this order: ErrUnset for the receiver, then the argument;
// ErrInvalidNumber for k, then a receiver above k, then an argument above k.
func (d Degree) DistanceUp(to Degree, k int) (int, error) {
	up, _, err := d.distances(to, k, "DistanceUp")

	return up, err
}

// DistanceDown returns the distance from d downward to another degree in a
// circle of k numbers, in the range 0 through k-1. Its errors and input order
// are those of DistanceUp.
func (d Degree) DistanceDown(to Degree, k int) (int, error) {
	_, down, err := d.distances(to, k, "DistanceDown")

	return down, err
}

// ShortestDistance returns the signed shortest distance from d to another
// degree: positive upward, negative downward. Equal distances choose upward;
// equal numbers return 0. Its errors and input order are those of DistanceUp.
func (d Degree) ShortestDistance(to Degree, k int) (int, error) {
	up, down, err := d.distances(to, k, "ShortestDistance")
	if err != nil {
		return 0, err
	}

	if up <= down {
		return up, nil
	}

	return -down, nil
}

// Normalize converts any positive continuous number of d to a number-only
// degree in a circle of k numbers and a nonnegative cycle count. The note of d
// is not carried over. Errors are ErrUnset for an unset receiver and
// ErrInvalidNumber for k below one.
func (d Degree) Normalize(k int) (normalized Degree, cycles int, err error) {
	if !d.IsValid() {
		return Degree{}, 0, unsetError("Normalize", "unset degree")
	}

	if k < 1 {
		return Degree{}, 0, invalidNumberError("Normalize", "circle size is below one")
	}

	position := d.number - 1

	return Degree{number: position%k + 1}, position / k, nil
}

// Denormalize converts a number-only degree inside a circle of k numbers and
// a nonnegative cycle count back to a continuous degree number. It reports
// ErrOutOfRange when cycles*k+Number() exceeds MaxNumber.
//
// Errors, in this order: ErrUnset for the receiver; ErrInvalidNumber for k,
// a number above k, or negative cycles; ErrOutOfRange for the result.
func (d Degree) Denormalize(k, cycles int) (Degree, error) {
	if err := d.checkInside("Denormalize", k); err != nil {
		return Degree{}, err
	}

	if cycles < 0 {
		return Degree{}, invalidNumberError("Denormalize", "cycle count is negative")
	}

	if cycles > (MaxNumber-d.number)/k {
		return Degree{}, outOfRangeError("Denormalize", "continuous number exceeds MaxNumber")
	}

	return Degree{number: cycles*k + d.number}, nil
}

// All returns a lazy sequence of number-only degrees from 1 through k. It
// validates k before returning and reports ErrInvalidNumber with a nil
// sequence when k is below one.
//
// The sequence uses constant memory, supports early termination, and can be
// iterated repeatedly and concurrently. Its index runs from 0 through k-1,
// so the transition after MaxNumber cannot overflow a degree counter.
func All(k int) (iter.Seq[Degree], error) {
	if k < 1 {
		return nil, invalidNumberError("All", "circle size is below one")
	}

	return func(yield func(Degree) bool) {
		for i := 0; i < k; i++ {
			if !yield(Degree{number: i + 1}) {
				return
			}
		}
	}, nil
}

// checkInside validates a receiver that must belong to the circle.
func (d Degree) checkInside(op string, k int) error {
	if !d.IsValid() {
		return unsetError(op, "unset degree")
	}

	if k < 1 {
		return invalidNumberError(op, "circle size is below one")
	}

	if d.number > k {
		return invalidNumberError(op, "degree number is above the circle size")
	}

	return nil
}

// distances validates both degrees and computes both directed distances
// without overflowing at large numbers.
func (d Degree) distances(to Degree, k int, op string) (up, down int, err error) {
	if !d.IsValid() {
		return 0, 0, unsetError(op, "unset degree")
	}

	if !to.IsValid() {
		return 0, 0, unsetError(op, "unset degree argument")
	}

	if k < 1 {
		return 0, 0, invalidNumberError(op, "circle size is below one")
	}

	if d.number > k {
		return 0, 0, invalidNumberError(op, "receiver number is above the circle size")
	}

	if to.number > k {
		return 0, 0, invalidNumberError(op, "argument number is above the circle size")
	}

	if to.number >= d.number {
		up = to.number - d.number
	} else {
		up = k - d.number + to.number
	}

	if up != 0 {
		down = k - up
	}

	return up, down, nil
}
