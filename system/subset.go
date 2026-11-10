package system

import "slices"

// SubsetOf reports whether every value of r belongs to other, within the
// guaranteed range [−M, M] of package alteration.
//
// The check does not enumerate values: each progression of r is split into at
// most 2·MaxProgressions + 1 index ranges on which the same progressions of
// other apply, and each range is checked over one period of at most
// MaxDenominator indices. The worst case costs about 10^8 simple operations;
// rules with a few progressions and small denominators take microseconds.
//
// It returns an error that matches ErrInvalid when r or other is unset.
func (r *Rule) SubsetOf(other *Rule) (bool, error) {
	if !r.IsValid() || !other.IsValid() {
		return false, errorf(ErrInvalid, "inclusion of unset rules")
	}

	if r == other {
		return true, nil
	}

	for _, p := range r.ps {
		if !covered(p, other) {
			return false, nil
		}
	}

	return true, nil
}

// MustSubsetOf is like SubsetOf but panics on error.
func (r *Rule) MustSubsetOf(other *Rule) bool {
	return must(r.SubsetOf(other))
}

// Equal reports whether r and other are set and are the same set of values
// within [−M, M]; how the sets are split into progressions does not matter.
// It reports false when either rule is unset.
func (r *Rule) Equal(other *Rule) bool {
	if !r.IsValid() || !other.IsValid() {
		return false
	}

	return r.MustSubsetOf(other) && other.MustSubsetOf(r)
}

// indexCover is the set of indices n of the points of a progression P that one
// progression Q of another rule contains: lo <= n <= hi and n ≡ alpha (mod e).
type indexCover struct {
	lo, hi, alpha, e int64
}

// covered reports whether every point of p belongs to the set rule other.
func covered(p Progression, other *Rule) bool {
	if p.den == 0 {
		return other.contains(p.First())
	}

	nmax := (p.hi - p.lo) / p.stride()

	var (
		covers [MaxProgressions]indexCover
		count  int
	)

	for _, q := range other.ps {
		if c, ok := coverOf(p, q, nmax); ok {
			covers[count] = c
			count++
		}
	}

	cs := covers[:count]

	cuts := make([]int64, 0, 2*count+2)
	cuts = append(cuts, 0, nmax+1)

	for _, c := range cs {
		cuts = append(cuts, c.lo, c.hi+1)
	}

	slices.Sort(cuts)
	cuts = slices.Compact(cuts)

	var bitset []uint64

	for i := 0; i+1 < len(cuts); i++ {
		from, to := cuts[i], cuts[i+1] // indices from..to-1
		period := int64(1)
		active := make([]indexCover, 0, count)

		for _, c := range cs {
			if c.lo <= from && to-1 <= c.hi {
				active = append(active, c)
				period = lcm(period, c.e)
			}
		}

		if len(active) == 0 {
			return false
		}

		if to-from < period {
			for n := from; n < to; n++ {
				if !coveredIndex(active, n) {
					return false
				}
			}

			continue
		}

		words := int((period + 63) / 64)
		if cap(bitset) < words {
			bitset = make([]uint64, words)
		}

		bitset = bitset[:words]
		clear(bitset)

		for _, c := range active {
			for n := c.alpha; n < period; n += c.e {
				bitset[n/64] |= 1 << (n % 64)
			}
		}

		for n := range period {
			if bitset[n/64]&(1<<(n%64)) == 0 {
				return false
			}
		}
	}

	return true
}

// coveredIndex reports whether some cover contains the index n.
func coveredIndex(cs []indexCover, n int64) bool {
	for _, c := range cs {
		if (n-c.alpha)%c.e == 0 {
			return true
		}
	}

	return false
}

// coverOf returns the indices n in 0..nmax of the points of p (a progression
// with a step) that q contains; ok is false when there are none.
//
// Points are compared in units 1/w with w = lcm(p.unit, q.unit) <= 2^32, so
// every coordinate is at most M·w <= 2^56 in magnitude.
func coverOf(p, q Progression, nmax int64) (indexCover, bool) {
	w := lcm(p.unit, q.unit)
	a := w / p.den // distance between neighboring points of p
	t0 := p.lo * (w / p.unit)
	qlo := q.lo * (w / q.unit)
	qhi := q.hi * (w / q.unit)

	var c indexCover

	if q.den == 0 {
		diff := qlo - t0
		if diff%a != 0 {
			return indexCover{}, false
		}

		n := diff / a
		c = indexCover{lo: n, hi: n, alpha: 0, e: 1}
	} else {
		alpha, e, ok := solveCongruence(a, qlo-t0, w/q.den)
		if !ok {
			return indexCover{}, false
		}

		c = indexCover{lo: ceilDiv(qlo-t0, a), hi: floorDiv(qhi-t0, a), alpha: alpha, e: e}
	}

	c.lo, c.hi = max(c.lo, 0), min(c.hi, nmax)
	if c.lo > c.hi {
		return indexCover{}, false
	}

	return c, true
}
