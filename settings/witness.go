package settings

import (
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/system"
)

// structureWitness returns an alteration value that belongs to the rule prev
// and not to the rule next, and whether one exists. Both rules are set.
//
// A quick pass looks at the values of prev with |a| <= 2, which finds the
// difference of every pair of built-in systems. Otherwise an exact search
// follows the scheme of system.Rule.SubsetOf: the index range of each
// progression of prev is split by the bounds of the progressions of next
// into at most 2K + 1 segments; on a segment the same progressions of next
// apply, and moving by d indices (one whole alteration unit) maps every
// lattice of next onto itself, so membership repeats with the period d and at
// most min(length, d) points of a segment are checked. All arithmetic is
// exact in int.
func structureWitness(prev, next *system.Rule) (alteration.Alteration, bool) {
	values, err := prev.Values(alteration.DoubleSharp())
	if err == nil {
		for a := range values {
			if !ruleContains(next, a) {
				return a, true
			}
		}
	}
	nextProgressions := next.Progressions()
	for _, p := range prev.Progressions() {
		if a, ok := progressionWitness(p, nextProgressions, next); ok {
			return a, true
		}
	}
	return alteration.Alteration{}, false
}

// progressionWitness returns a point of the progression p that does not
// belong to the rule next made of the progressions qs.
func progressionWitness(p system.Progression, qs []system.Progression, next *system.Rule) (alteration.Alteration, bool) {
	if p.Den() == 0 {
		a := p.First()
		return a, !ruleContains(next, a)
	}
	d := p.Den()
	fNum, fDen := p.First().Fraction()
	// The points of p are x_i = F + i/d for i in 0..n.
	n := indexFloor(p.Last(), fNum, fDen, d)

	cuts := []int{0, n + 1}
	for _, q := range qs {
		if !q.IsValid() {
			continue
		}
		lo := indexCeil(q.First(), fNum, fDen, d)
		hi := indexFloor(q.Last(), fNum, fDen, d)
		cuts = append(cuts, clamp(lo, 0, n+1), clamp(hi+1, 0, n+1))
	}
	slices.Sort(cuts)
	cuts = slices.Compact(cuts)

	for k := 0; k+1 < len(cuts); k++ {
		from, to := cuts[k], cuts[k+1]
		end := to
		if to-from > d {
			end = from + d
		}
		for i := from; i < end; i++ {
			a, err := alteration.FromFraction(fNum*d+i*fDen, fDen*d)
			if err != nil {
				continue
			}
			if !ruleContains(next, a) {
				return a, true
			}
		}
	}
	return alteration.Alteration{}, false
}

// indexFloor returns ⌊(x − F)·d⌋ for F = fNum/fDen, exactly.
func indexFloor(x alteration.Alteration, fNum, fDen, d int) int {
	num, den := diffFraction(x, fNum, fDen)
	return mulDivFloor(num, den, d)
}

// indexCeil returns ⌈(x − F)·d⌉ for F = fNum/fDen, exactly.
func indexCeil(x alteration.Alteration, fNum, fDen, d int) int {
	num, den := diffFraction(x, fNum, fDen)
	return -mulDivFloor(-num, den, d)
}

// diffFraction returns x − fNum/fDen as num/den with den > 0. Numerators of
// alteration values are at most 2^40 in magnitude and denominators at most
// 2^16, so the results stay below 2^58 and 2^32.
func diffFraction(x alteration.Alteration, fNum, fDen int) (num, den int) {
	xNum, xDen := x.Fraction()
	return xNum*fDen - fNum*xDen, xDen * fDen
}

// mulDivFloor returns ⌊num·m/den⌋ for den > 0 and m > 0 without overflow when
// num/den is an alteration difference (at most 2^25 in magnitude), den is at
// most 2^32 and m is at most 2^16.
func mulDivFloor(num, den, m int) int {
	q := num / den
	r := num % den
	if r < 0 {
		q--
		r += den
	}
	return q*m + r*m/den
}

// clamp returns v limited to lo..hi.
func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// ruleContains reports whether a belongs to the set rule r.
func ruleContains(r *system.Rule, a alteration.Alteration) bool {
	ok, err := r.Contains(a)
	return err == nil && ok
}
