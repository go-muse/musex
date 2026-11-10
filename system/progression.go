package system

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/alteration"
)

// maxMagnitude is M, the largest magnitude of an alteration value.
const maxMagnitude int64 = alteration.MaxMagnitude

// Progression is one arithmetic progression of alteration values: the set
// {First + n/d : n is an integer, First <= x <= Last} for a step 1/d, or the
// single value {First} for the step 0. Every progression lies within the
// guaranteed range [−M, M] of package alteration, M = alteration.MaxMagnitude,
// and all its points are alteration values.
//
// A Progression is created by NewProgression and narrowed by WithLower and
// WithUpper. It is kept in a canonical form: First is the smallest point, Last
// the largest, and a progression with one point has the step 0. Two set
// progressions are therefore equal by == and by Equal exactly when they are the
// same set; a Progression can be used as a map key.
//
// The zero value is the unset progression: IsValid reports false, Den returns
// −1, Step, First and Last return the natural value, BoundedBelow, BoundedAbove,
// Contains and Equal report false, and String returns "Progression(unset)".
//
// A Progression is an immutable value, safe for concurrent use.
type Progression struct {
	lo, hi int64 // First·unit and Last·unit
	unit   int64 // common denominator D of the points; 0 when unset
	den    int64 // d of the step 1/d; 0 for a single value
}

// NewProgression returns the progression that starts at start and has the
// step step: the single value start for the step 0, or all points start + n/d
// within [−M, M] for the step 1/d, 1 <= d <= MaxDenominator.
//
// It returns an error that matches ErrDefinition and ErrRule when the step is
// neither 0 nor 1/d (for example negative, 2 or 2/3), or when the least common
// multiple of the denominator of start and d exceeds MaxDenominator, so that
// some points would not be alteration values.
func NewProgression(start, step alteration.Alteration) (Progression, error) {
	sp, sq := step.Fraction()
	if sp == 0 {
		return single(start), nil
	}

	if sp != 1 {
		return Progression{}, definitionError(ErrRule, "step %v is neither 0 nor 1/d", step)
	}

	d := int64(sq)
	p, q := start.Fraction()

	unit := lcm(int64(q), d)
	if unit > MaxDenominator {
		return Progression{}, definitionError(ErrRule,
			"points of %v + Z/%d have the denominator %d, more than %d", start, d, unit, MaxDenominator)
	}

	t0 := int64(p) * (unit / int64(q))
	s := unit / d
	bound := maxMagnitude * unit

	return Progression{
		lo:   -bound + floorMod(t0+bound, s),
		hi:   bound - floorMod(bound-t0, s),
		unit: unit,
		den:  d,
	}, nil
}

// MustNewProgression is like NewProgression but panics on error.
func MustNewProgression(start, step alteration.Alteration) Progression {
	return must(NewProgression(start, step))
}

// WithLower returns the progression narrowed to the points x >= lower. Calls
// can be repeated; each narrows further.
//
// It returns an error that matches ErrInvalid for the unset progression, and an
// error that matches ErrDefinition and ErrRule when no point remains.
func (p Progression) WithLower(lower alteration.Alteration) (Progression, error) {
	if !p.IsValid() {
		return Progression{}, errorf(ErrInvalid, "lower bound of an unset progression")
	}

	lp, lq := lower.Fraction()
	lo := ceilDiv(int64(lp)*p.unit, int64(lq))

	if p.den != 0 {
		lo += floorMod(p.lo-lo, p.stride())
	}

	lo = max(lo, p.lo)
	if lo > p.hi {
		return Progression{}, definitionError(ErrRule, "no point of %v is at least %v", p, lower)
	}

	return canonical(lo, p.hi, p.unit, p.den), nil
}

// MustWithLower is like WithLower but panics on error.
func (p Progression) MustWithLower(lower alteration.Alteration) Progression {
	return must(p.WithLower(lower))
}

// WithUpper returns the progression narrowed to the points x <= upper. Calls
// can be repeated; each narrows further.
//
// It returns an error that matches ErrInvalid for the unset progression, and an
// error that matches ErrDefinition and ErrRule when no point remains.
func (p Progression) WithUpper(upper alteration.Alteration) (Progression, error) {
	if !p.IsValid() {
		return Progression{}, errorf(ErrInvalid, "upper bound of an unset progression")
	}

	up, uq := upper.Fraction()
	hi := floorDiv(int64(up)*p.unit, int64(uq))

	if p.den != 0 {
		hi -= floorMod(hi-p.lo, p.stride())
	}

	hi = min(hi, p.hi)
	if hi < p.lo {
		return Progression{}, definitionError(ErrRule, "no point of %v is at most %v", p, upper)
	}

	return canonical(p.lo, hi, p.unit, p.den), nil
}

// MustWithUpper is like WithUpper but panics on error.
func (p Progression) MustWithUpper(upper alteration.Alteration) Progression {
	return must(p.WithUpper(upper))
}

// IsValid reports whether p is set.
func (p Progression) IsValid() bool {
	return p.unit != 0
}

// Step returns the step: 0 for a single value, 1/d otherwise. It returns the
// natural value 0 for the unset progression.
func (p Progression) Step() alteration.Alteration {
	if p.den == 0 {
		return alteration.Natural()
	}

	return alteration.MustFromFraction(1, int(p.den))
}

// Den returns d of the step 1/d, 0 for a single value and −1 for the unset
// progression.
func (p Progression) Den() int {
	if !p.IsValid() {
		return -1
	}

	return int(p.den)
}

// First returns the smallest point; the natural value for the unset
// progression.
func (p Progression) First() alteration.Alteration {
	if !p.IsValid() {
		return alteration.Natural()
	}

	return alteration.MustFromFraction(int(p.lo), int(p.unit))
}

// Last returns the largest point; the natural value for the unset progression.
func (p Progression) Last() alteration.Alteration {
	if !p.IsValid() {
		return alteration.Natural()
	}

	return alteration.MustFromFraction(int(p.hi), int(p.unit))
}

// BoundedBelow reports whether the progression excludes a point of its lattice
// below First within [−M, M]. A single value is bounded. A progression bounded
// exactly at −M is the same set as an unbounded one and reports false.
func (p Progression) BoundedBelow() bool {
	if !p.IsValid() {
		return false
	}

	return p.den == 0 || p.lo-p.stride() >= -maxMagnitude*p.unit
}

// BoundedAbove reports whether the progression excludes a point of its lattice
// above Last within [−M, M]. A single value is bounded. A progression bounded
// exactly at M is the same set as an unbounded one and reports false.
func (p Progression) BoundedAbove() bool {
	if !p.IsValid() {
		return false
	}

	return p.den == 0 || p.hi+p.stride() <= maxMagnitude*p.unit
}

// Contains reports whether a is a point of p. It does not enumerate points.
func (p Progression) Contains(a alteration.Alteration) bool {
	if !p.IsValid() {
		return false
	}

	ap, aq := a.Fraction()
	if p.unit%int64(aq) != 0 {
		return false
	}

	t := int64(ap) * (p.unit / int64(aq))
	if t < p.lo || t > p.hi {
		return false
	}

	return p.den == 0 || (t-p.lo)%p.stride() == 0
}

// Equal reports whether p and other are set and are the same set of values.
func (p Progression) Equal(other Progression) bool {
	return p.IsValid() && p == other
}

// String returns a debugging form. A single value is written as "{0}" or
// "{+1/2}". A progression with a step is written as "<r> + Z/<d> in
// [<lo>, <hi>]", where r is the point of the lattice First + (1/d)Z in
// [0, 1/d), lo is First or "-inf" when the progression is not bounded below,
// and hi is Last or "+inf" when it is not bounded above. For example, the
// integers are "0 + Z/1 in [-inf, +inf]" and the integers from −2 to 2 are
// "0 + Z/1 in [-2, +2]". The unset progression is "Progression(unset)".
func (p Progression) String() string {
	if !p.IsValid() {
		return "Progression(unset)"
	}

	if p.den == 0 {
		return "{" + p.First().String() + "}"
	}

	var b strings.Builder

	r := alteration.MustFromFraction(int(floorMod(p.lo, p.stride())), int(p.unit))

	b.WriteString(r.String())
	b.WriteString(" + Z/")
	b.WriteString(strconv.FormatInt(p.den, 10))
	b.WriteString(" in [")

	if p.BoundedBelow() {
		b.WriteString(p.First().String())
	} else {
		b.WriteString("-inf")
	}

	b.WriteString(", ")

	if p.BoundedAbove() {
		b.WriteString(p.Last().String())
	} else {
		b.WriteString("+inf")
	}

	b.WriteString("]")

	return b.String()
}

// stride returns the distance between neighboring points in units 1/unit; the
// caller ensures den >= 1.
func (p Progression) stride() int64 {
	return p.unit / p.den
}

// single returns the progression with the one value a.
func single(a alteration.Alteration) Progression {
	p, q := a.Fraction()

	return Progression{lo: int64(p), hi: int64(p), unit: int64(q)}
}

// canonical returns the canonical progression with the points lo..hi (in units
// 1/unit) of the lattice of step 1/den. A progression with one point becomes a
// single value in lowest terms.
func canonical(lo, hi, unit, den int64) Progression {
	if lo == hi {
		g := gcd(lo, unit)

		return Progression{lo: lo / g, hi: lo / g, unit: unit / g}
	}

	return Progression{lo: lo, hi: hi, unit: unit, den: den}
}
