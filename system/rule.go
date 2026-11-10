package system

import (
	"iter"
	"strings"

	"github.com/go-muse/musex/alteration"
)

// Rule is the set A of admissible alteration values of a system: the union of
// 1 to MaxProgressions progressions, kept in the order they were given, without
// exact repeats. Points outside the guaranteed range [−M, M] of package
// alteration are not part of any rule, so the integers and the integers from
// −M to M are the same set.
//
// A Rule is used through a pointer and is immutable after creation; one Rule
// can be shared by many systems and used concurrently. The nil pointer and a
// pointer to the zero Rule are the unset rule: IsValid reports false,
// Progressions returns an empty slice, String returns "Rule(unset)", Equal
// reports false, and the other methods return an error that matches
// ErrInvalid. A Rule value is not comparable: == on pointers is identity; use
// Equal for the equality of sets.
type Rule struct {
	ps []Progression
}

// NewRule returns the rule made of the progressions ps in the given order. An
// exact repeat of an earlier progression (by ==) is dropped.
//
// It returns an error that matches ErrInvalid when some progression is unset,
// and an error that matches ErrDefinition and ErrRule when there is no
// progression or more than MaxProgressions remain after dropping repeats.
func NewRule(ps ...Progression) (*Rule, error) {
	for i, p := range ps {
		if !p.IsValid() {
			return nil, errorf(ErrInvalid, "progression %d of a rule is unset", i)
		}
	}

	if len(ps) == 0 {
		return nil, definitionError(ErrRule, "a rule has no progressions")
	}

	out := make([]Progression, 0, min(len(ps), MaxProgressions))

	for _, p := range ps {
		if containsProgression(out, p) {
			continue
		}

		if len(out) == MaxProgressions {
			return nil, definitionError(ErrRule, "a rule has more than %d progressions", MaxProgressions)
		}

		out = append(out, p)
	}

	return &Rule{ps: out}, nil
}

// MustNewRule is like NewRule but panics on error.
func MustNewRule(ps ...Progression) *Rule {
	return must(NewRule(ps...))
}

// Integers returns the rule of all integers: Z. Each call returns a new Rule.
func Integers() *Rule {
	return &Rule{ps: []Progression{integers()}}
}

// Multiples returns the rule of all multiples of 1/den: (1/den)Z, for den in
// 1..MaxDenominator. Each call returns a new Rule.
//
// It returns an error that matches ErrDefinition and ErrRule when den is out of
// range.
func Multiples(den int) (*Rule, error) {
	if den < 1 || den > MaxDenominator {
		return nil, definitionError(ErrRule, "denominator %d is outside 1..%d", den, MaxDenominator)
	}

	p := MustNewProgression(alteration.Natural(), alteration.MustFromFraction(1, den))

	return &Rule{ps: []Progression{p}}, nil
}

// MustMultiples is like Multiples but panics on error.
func MustMultiples(den int) *Rule {
	return must(Multiples(den))
}

// IntegerRange returns the rule of the integers from..to, both included. Each
// call returns a new Rule.
//
// It returns an error that matches ErrDefinition and ErrRule when from > to or
// when from or to is outside [−M, M].
func IntegerRange(from, to int) (*Rule, error) {
	if from > to || int64(from) < -maxMagnitude || int64(to) > maxMagnitude {
		return nil, definitionError(ErrRule, "integer range %d..%d is empty or outside [%d, %d]",
			from, to, -maxMagnitude, maxMagnitude)
	}

	p := integers().
		MustWithLower(alteration.MustFromInt(from)).
		MustWithUpper(alteration.MustFromInt(to))

	return &Rule{ps: []Progression{p}}, nil
}

// MustIntegerRange is like IntegerRange but panics on error.
func MustIntegerRange(from, to int) *Rule {
	return must(IntegerRange(from, to))
}

// Single returns the rule with the one value a. Each call returns a new Rule.
func Single(a alteration.Alteration) *Rule {
	return &Rule{ps: []Progression{single(a)}}
}

// Union returns the rule with the progressions of all rules in order, built as
// by NewRule.
//
// It returns an error that matches ErrInvalid when some rule is nil or unset,
// and an error that matches ErrDefinition and ErrRule when there are no rules
// or the union has more than MaxProgressions progressions.
func Union(rules ...*Rule) (*Rule, error) {
	for i, r := range rules {
		if !r.IsValid() {
			return nil, errorf(ErrInvalid, "rule %d of a union is unset", i)
		}
	}

	if len(rules) == 0 {
		return nil, definitionError(ErrRule, "a union has no rules")
	}

	var ps []Progression
	for _, r := range rules {
		ps = append(ps, r.ps...)
	}

	return NewRule(ps...)
}

// MustUnion is like Union but panics on error.
func MustUnion(rules ...*Rule) *Rule {
	return must(Union(rules...))
}

// IsValid reports whether r is set.
func (r *Rule) IsValid() bool {
	return r != nil && len(r.ps) > 0
}

// Progressions returns a new slice with the progressions of r in the order they
// were given; the slice belongs to the caller. It is empty for the unset rule.
func (r *Rule) Progressions() []Progression {
	if !r.IsValid() {
		return []Progression{}
	}

	return append([]Progression(nil), r.ps...)
}

// Contains reports whether a belongs to r. It costs O(MaxProgressions) and does
// not enumerate values.
//
// It returns an error that matches ErrInvalid for the unset rule.
func (r *Rule) Contains(a alteration.Alteration) (bool, error) {
	if !r.IsValid() {
		return false, errorf(ErrInvalid, "membership in an unset rule")
	}

	return r.contains(a), nil
}

// MustContains is like Contains but panics on error.
func (r *Rule) MustContains(a alteration.Alteration) bool {
	return must(r.Contains(a))
}

// Values returns the values a of r with |a| <= limit in strictly increasing
// order, without repeats. The sequence is lazy, can be stopped at any element
// and can be iterated again with the same result; it keeps O(MaxProgressions)
// state and never builds the whole set.
//
// It returns an error that matches ErrInvalid for the unset rule and, after
// that check, an error that matches ErrLimit when limit is negative.
func (r *Rule) Values(limit alteration.Alteration) (iter.Seq[alteration.Alteration], error) {
	if !r.IsValid() {
		return nil, errorf(ErrInvalid, "values of an unset rule")
	}

	if limit.Sign() < 0 {
		return nil, errorf(ErrLimit, "limit %v", limit)
	}

	ps := r.ps

	return func(yield func(alteration.Alteration) bool) {
		var streams [MaxProgressions]pointStream

		n := 0

		for _, p := range ps {
			if st, ok := newPointStream(p, limit); ok {
				streams[n] = st
				n++
			}
		}

		var (
			last    alteration.Alteration
			started bool
		)

		for {
			best := -1

			for i := range n {
				if streams[i].done() {
					continue
				}

				if best < 0 || streams[i].less(streams[best]) {
					best = i
				}
			}

			if best < 0 {
				return
			}

			v := streams[best].value()
			streams[best].next()

			if started && v == last {
				continue
			}

			if !yield(v) {
				return
			}

			last, started = v, true
		}
	}, nil
}

// MustValues is like Values but panics on error.
func (r *Rule) MustValues(limit alteration.Alteration) iter.Seq[alteration.Alteration] {
	return must(r.Values(limit))
}

// String returns a debugging form: the progressions in order, separated by
// " | ", for example "0 + Z/1 in [-inf, +inf] | {+1/2} | {-1/2}". The unset
// rule is "Rule(unset)".
func (r *Rule) String() string {
	if !r.IsValid() {
		return "Rule(unset)"
	}

	parts := make([]string, len(r.ps))
	for i, p := range r.ps {
		parts[i] = p.String()
	}

	return strings.Join(parts, " | ")
}

// contains reports whether a belongs to the set rule r.
func (r *Rule) contains(a alteration.Alteration) bool {
	for _, p := range r.ps {
		if p.Contains(a) {
			return true
		}
	}

	return false
}

// integers returns the progression of all integers.
func integers() Progression {
	return MustNewProgression(alteration.Natural(), alteration.Sharp())
}

// containsProgression reports whether ps has p.
func containsProgression(ps []Progression, p Progression) bool {
	for _, q := range ps {
		if q == p {
			return true
		}
	}

	return false
}

// pointStream yields the points of one progression within a limit in
// increasing order. Points are t/unit for t = cur, cur + step, ..., end.
type pointStream struct {
	cur, end, step, unit int64
}

// newPointStream returns the stream of the points of p within [−limit, limit];
// ok is false when there are none.
func newPointStream(p Progression, limit alteration.Alteration) (pointStream, bool) {
	lp, lq := limit.Fraction()
	bound := floorDiv(int64(lp)*p.unit, int64(lq))

	lo, hi := max(p.lo, -bound), min(p.hi, bound)
	step := int64(1)

	if p.den != 0 {
		step = p.stride()
		lo += floorMod(p.lo-lo, step)
		hi -= floorMod(hi-p.lo, step)
	}

	if lo > hi {
		return pointStream{}, false
	}

	return pointStream{cur: lo, end: hi, step: step, unit: p.unit}, true
}

func (s *pointStream) done() bool {
	return s.cur > s.end
}

func (s *pointStream) next() {
	s.cur += s.step
}

func (s *pointStream) value() alteration.Alteration {
	return alteration.MustFromFraction(int(s.cur), int(s.unit))
}

// less reports whether the current point of s is below the current point of o.
func (s *pointStream) less(o pointStream) bool {
	return s.cur*o.unit < o.cur*s.unit
}
