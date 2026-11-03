package octave

import "math"

// Shift returns the octave k octaves higher than o, or lower for a negative
// k: the octave whose ASPN number is the ASPN number of o plus k, computed
// exactly. Any int k is accepted, including math.MinInt: FromASPN(0) shifted
// by math.MinInt is FromASPN(math.MinInt). Shift(0) returns o.
//
// If o.Shift(k1) and then Shift(k2) succeed and k1+k2 fits in an int, the
// result is o.Shift(k1 + k2); b.Shift(a.MustSub(b)) is a.
//
// On error, Shift returns the zero Octave and an error: if o is invalid, the
// error matches [ErrInvalid]; if the resulting octave is not representable,
// the error matches [ErrOutOfRange]. The result never wraps around.
//
// To move down by k, use [Octave.ShiftDown] rather than Shift(-k): for
// k = math.MinInt, -k does not fit in an int.
func (o Octave) Shift(k int) (Octave, error) {
	if !o.valid {
		return Octave{}, ErrInvalid
	}

	sum := o.aspn + k
	if (k > 0 && sum < o.aspn) || (k < 0 && sum > o.aspn) {
		return Octave{}, outOfRange("shift %v by %d", o, k)
	}

	return of(sum), nil
}

// MustShift is like [Octave.Shift] but panics if Shift returns an error. The
// panic value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustShift(k int) Octave {
	return must(o.Shift(k))
}

// ShiftDown returns the octave k octaves lower than o, or higher for a
// negative k: the octave whose ASPN number is the ASPN number of o minus k,
// computed exactly for every int k, including math.MinInt:
// FromASPN(-1).ShiftDown(math.MinInt) is FromASPN(math.MaxInt). For k other
// than math.MinInt, o.ShiftDown(k) is o.Shift(-k).
//
// ShiftDown reverses [Octave.Shift] for every int k: if o.Shift(k) succeeds,
// o.MustShift(k).ShiftDown(k) returns o without an error, and if
// o.ShiftDown(k) succeeds, o.MustShiftDown(k).Shift(k) returns o without an
// error.
//
// On error, ShiftDown returns the zero Octave and an error: if o is invalid,
// the error matches [ErrInvalid]; if the resulting octave is not
// representable, the error matches [ErrOutOfRange]. The result never wraps
// around.
func (o Octave) ShiftDown(k int) (Octave, error) {
	if !o.valid {
		return Octave{}, ErrInvalid
	}

	diff, ok := sub(o.aspn, k)
	if !ok {
		return Octave{}, outOfRange("shift %v down by %d", o, k)
	}

	return of(diff), nil
}

// MustShiftDown is like [Octave.ShiftDown] but panics if ShiftDown returns an
// error. The panic value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustShiftDown(k int) Octave {
	return must(o.ShiftDown(k))
}

// Up returns the next octave above o, the same as o.Shift(1), with the same
// errors: for the highest representable octave, FromASPN(math.MaxInt), the
// error matches [ErrOutOfRange].
func (o Octave) Up() (Octave, error) {
	return o.Shift(1)
}

// MustUp is like [Octave.Up] but panics if Up returns an error. The panic
// value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustUp() Octave {
	return must(o.Up())
}

// Down returns the next octave below o, the same as o.Shift(-1), with the
// same errors: for the lowest representable octave, FromASPN(math.MinInt),
// the error matches [ErrOutOfRange].
func (o Octave) Down() (Octave, error) {
	return o.Shift(-1)
}

// MustDown is like [Octave.Down] but panics if Down returns an error. The
// panic value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustDown() Octave {
	return must(o.Down())
}

// Sub returns how many octaves o is higher than other: the ASPN number of o
// minus the ASPN number of other, computed exactly. The result is positive if
// o is higher, negative if it is lower, and 0 only for the same octave. It is
// a difference of octaves, not of numbers: in the Franco-Belgian numbering,
// the octaves numbered 3 and -1 are 3 octaves apart, although their numbers
// differ by 4. Like time.Time.Sub, the result is o minus other, and
// o.Sub(other) is -other.Sub(o) when both succeed.
//
// On error, Sub returns 0 and an error: if o or other is invalid, the error
// matches [ErrInvalid]; if the difference does not fit in an int, the error
// matches [ErrOutOfRange]. For example, FromASPN(math.MinInt).Sub(FromASPN(0))
// is math.MinInt, while FromASPN(0).Sub(FromASPN(math.MinInt)) is an error.
func (o Octave) Sub(other Octave) (int, error) {
	if !o.valid || !other.valid {
		return 0, ErrInvalid
	}

	diff, ok := sub(o.aspn, other.aspn)
	if !ok {
		return 0, outOfRange("%v minus %v", o, other)
	}

	return diff, nil
}

// MustSub is like [Octave.Sub] but panics if Sub returns an error. The panic
// value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustSub(other Octave) int {
	return must(o.Sub(other))
}

// Distance returns how many octaves lie between o and other, without
// direction: the absolute value of o.Sub(other). It does not depend on the
// order of o and other and is 0 only for the same octave.
//
// The result is an int, like the result of [Octave.Sub] and the argument of
// [Octave.Shift]. Distance succeeds exactly when Sub succeeds and its result
// is not math.MinInt; otherwise the distance, at least 2^63, does not fit in
// an int. For example, FromASPN(math.MinInt+1).Distance(FromASPN(0)) is
// math.MaxInt, while FromASPN(math.MinInt).Distance(FromASPN(0)) is an error.
//
// On error, Distance returns 0 and an error: if o or other is invalid, the
// error matches [ErrInvalid]; if the distance does not fit in an int, the
// error matches [ErrOutOfRange].
func (o Octave) Distance(other Octave) (int, error) {
	if !o.valid || !other.valid {
		return 0, ErrInvalid
	}

	diff, ok := sub(o.aspn, other.aspn)
	if !ok || diff == math.MinInt {
		return 0, outOfRange("distance between %v and %v", o, other)
	}

	if diff < 0 {
		return -diff, nil
	}

	return diff, nil
}

// MustDistance is like [Octave.Distance] but panics if Distance returns an
// error. The panic value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustDistance(other Octave) int {
	return must(o.Distance(other))
}

// sub returns a-b and true if the exact difference fits in an int, and false
// otherwise. It never computes -b, so it is exact for b = math.MinInt too.
func sub(a, b int) (int, bool) {
	diff := a - b
	if (b > 0 && diff > a) || (b < 0 && diff < a) {
		return 0, false
	}

	return diff, true
}
