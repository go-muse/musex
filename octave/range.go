package octave

import "iter"

// Range returns the octaves from one to another, inclusive, one octave at a
// time: upward if to is higher than from, downward if it is lower, and a
// single octave if they are equal. The number of octaves is the distance
// between from and to plus one; it may be up to 2^64 and fit neither in an
// int nor in a slice.
//
// The sequence is lazy: it yields the octaves one by one, builds no slice,
// and starts no goroutine; it runs in the goroutine of its caller. The
// caller may stop the iteration early with break or return, and the sequence
// then stops calling yield. At either end of the axis, the iteration stops
// after yielding to, without an error or a wraparound. Each call of the
// sequence starts again from from, so it can be used many times and
// concurrently by multiple goroutines. It does not panic.
//
// If from or to is invalid, Range returns a nil sequence and an error that
// matches [ErrInvalid]; ranging over a nil sequence panics, so check the
// error or use [MustRange]. There are no other errors.
func Range(from, to Octave) (iter.Seq[Octave], error) {
	if !from.valid || !to.valid {
		return nil, ErrInvalid
	}

	step := 1
	if to.aspn < from.aspn {
		step = -1
	}

	return func(yield func(Octave) bool) {
		for cur := from.aspn; ; cur += step {
			if !yield(of(cur)) || cur == to.aspn {
				return
			}
		}
	}, nil
}

// MustRange is like [Range] but panics if Range returns an error. The panic
// value is that error, which matches [ErrInvalid].
//
// It is intended for octaves known to be valid, as in
// for o := range octave.MustRange(lo, hi) { ... }.
func MustRange(from, to Octave) iter.Seq[Octave] {
	return must(Range(from, to))
}
