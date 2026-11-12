package tuplet

import (
	"fmt"
	"math"

	"github.com/go-muse/musex/common/fraction"
)

// Tuplet is a tuplet n:m: n equal parts (the actual count) played in the time
// of m parts of the same written value (the normal count). Both counts are at
// least 1 and at most math.MaxInt.
//
// The zero value Tuplet{} is the unset tuplet. It is not 1:1. Only
// [Tuplet.IsSet], [Tuplet.Equal], [Tuplet.Compare], and [Tuplet.String] accept
// it without an error; every other operation returns an error wrapping
// [ErrUnset].
//
// A Tuplet is an immutable comparable value of fixed size without pointers.
// The pair is stored as given: 6:4 is not reduced to 3:2. Every value has a
// single representation, so == matches [Tuplet.Equal] and a Tuplet can be
// used as a map key. Copying a Tuplet copies two integers.
type Tuplet struct {
	actual int
	normal int
}

// Kind is the kind of a tuplet by the relation of its counts. The zero value
// and every value other than the declared constants are invalid. Kind values
// are compared with ==.
type Kind uint8

// Supported tuplet kinds.
const (
	// Compressing is a tuplet with actual > normal: its notes are shorter
	// than written (a triplet 3:2).
	Compressing Kind = iota + 1

	// Expanding is a tuplet with actual < normal: its notes are longer than
	// written (a duplet 2:3).
	Expanding

	// Neutral is a tuplet with actual == normal: its coefficient is 1.
	Neutral
)

// IsValid reports whether k is one of the declared constants Compressing,
// Expanding, and Neutral.
func (k Kind) IsValid() bool {
	return k >= Compressing && k <= Neutral
}

// String returns "compressing", "expanding", or "neutral" for a valid kind
// and "Kind(invalid)" otherwise.
func (k Kind) String() string {
	switch k {
	case Compressing:
		return "compressing"
	case Expanding:
		return "expanding"
	case Neutral:
		return "neutral"
	default:
		return "Kind(invalid)"
	}
}

// New returns the tuplet actual:normal, stored as given. It returns an error
// wrapping ErrInvalidCount if actual < 1 or normal < 1. Pairs with
// actual == normal (including 1:1), actual == 1, and normal == 1 are valid.
func New(actual, normal int) (Tuplet, error) {
	if actual < 1 || normal < 1 {
		return Tuplet{}, fmt.Errorf("%w: %d:%d", ErrInvalidCount, actual, normal)
	}
	return Tuplet{actual: actual, normal: normal}, nil
}

// MustNew is like New but panics with the returned error.
func MustNew(actual, normal int) Tuplet {
	return must(New(actual, normal))
}

// coefficientCounts validates a coefficient c and returns the magnitudes of
// its stored numerator and denominator.
func coefficientCounts(c fraction.Fraction) (num, den int, err error) {
	if !c.IsValid() {
		return 0, 0, fmt.Errorf("%w: unset coefficient", ErrInvalidArgument)
	}
	if c.Sign() <= 0 {
		return 0, 0, fmt.Errorf("%w: coefficient %s is not positive", ErrInvalidCount, c)
	}
	p, q, ok := c.Pair()
	if !ok || p == math.MinInt || q == math.MinInt {
		return 0, 0, fmt.Errorf("%w: coefficient component exceeds int", ErrOutOfRange)
	}
	return abs(p), abs(q), nil
}

// FromCoefficient returns the tuplet whose coefficient is c. For a stored pair
// (p, q) of c, the result is |q|:|p|: 2/3 gives 3:2, 4/6 gives 6:4, and
// (-2)/(-3) gives 3:2. It returns an error wrapping ErrInvalidArgument if c is
// unset, ErrInvalidCount if c <= 0, and ErrOutOfRange if |p| or |q| does not
// fit in int.
func FromCoefficient(c fraction.Fraction) (Tuplet, error) {
	p, q, err := coefficientCounts(c)
	if err != nil {
		return Tuplet{}, err
	}
	return Tuplet{actual: q, normal: p}, nil
}

// MustFromCoefficient is like FromCoefficient but panics with the returned
// error.
func MustFromCoefficient(c fraction.Fraction) Tuplet {
	return must(FromCoefficient(c))
}

// FromReciprocalCoefficient returns the tuplet whose reciprocal coefficient
// is c (the LilyPond fraction n/m). For a stored pair (p, q) of c, the result
// is |p|:|q|: 3/2 gives 3:2 and 6/4 gives 6:4. Errors are as for
// FromCoefficient.
func FromReciprocalCoefficient(c fraction.Fraction) (Tuplet, error) {
	p, q, err := coefficientCounts(c)
	if err != nil {
		return Tuplet{}, err
	}
	return Tuplet{actual: p, normal: q}, nil
}

// MustFromReciprocalCoefficient is like FromReciprocalCoefficient but panics
// with the returned error.
func MustFromReciprocalCoefficient(c fraction.Fraction) Tuplet {
	return must(FromReciprocalCoefficient(c))
}

// FromUnitRatio converts the explicit notation "actual parts of unit a in the
// time of normal parts of unit b" with ratio = b/a into a tuplet counted in
// the unit of the actual parts: actual:(normal·ratio). For example, three
// eighths in the time of one quarter is FromUnitRatio(3, 1, 2/1) = 3:2. The
// difference between the original units is not kept.
//
// In order of checking, it returns an error wrapping ErrInvalidCount if
// actual < 1 or normal < 1; ErrInvalidArgument if ratio is unset or not
// positive; ErrNotRepresentable if normal·ratio is not an integer; and
// ErrOutOfRange together with fraction.ErrNotRepresentable if normal·ratio is
// an integer greater than math.MaxInt.
func FromUnitRatio(actual, normal int, ratio fraction.Fraction) (Tuplet, error) {
	if actual < 1 || normal < 1 {
		return Tuplet{}, fmt.Errorf("%w: %d:%d", ErrInvalidCount, actual, normal)
	}
	if !ratio.IsValid() {
		return Tuplet{}, fmt.Errorf("%w: unset ratio", ErrInvalidArgument)
	}
	if ratio.Sign() <= 0 {
		return Tuplet{}, fmt.Errorf("%w: ratio %s is not positive", ErrInvalidArgument, ratio)
	}
	product, exact, err := ratio.ScaleIntExact(normal)
	if err != nil {
		return Tuplet{}, fmt.Errorf("%w: normal count %d times ratio %s: %w",
			ErrOutOfRange, normal, ratio, err)
	}
	if !exact {
		return Tuplet{}, fmt.Errorf("%w: normal count %d times ratio %s",
			ErrNotRepresentable, normal, ratio)
	}
	return Tuplet{actual: actual, normal: product}, nil
}

// MustFromUnitRatio is like FromUnitRatio but panics with the returned error.
func MustFromUnitRatio(actual, normal int, ratio fraction.Fraction) Tuplet {
	return must(FromUnitRatio(actual, normal, ratio))
}

// Duplet returns the duplet 2:3 (Russian "дуоль").
func Duplet() Tuplet { return Tuplet{actual: 2, normal: 3} }

// Triplet returns the triplet 3:2 (Russian "триоль").
func Triplet() Tuplet { return Tuplet{actual: 3, normal: 2} }

// Quadruplet returns the quadruplet 4:3 (Russian "квартоль").
func Quadruplet() Tuplet { return Tuplet{actual: 4, normal: 3} }

// Quintuplet returns the quintuplet 5:4 (Russian "квинтоль").
func Quintuplet() Tuplet { return Tuplet{actual: 5, normal: 4} }

// Sextuplet returns the sextuplet 6:4 (Russian "секстоль").
func Sextuplet() Tuplet { return Tuplet{actual: 6, normal: 4} }

// Septuplet returns the septuplet 7:4 (Russian "септоль").
func Septuplet() Tuplet { return Tuplet{actual: 7, normal: 4} }

// Octuplet returns the octuplet 8:6 (Russian "октоль").
func Octuplet() Tuplet { return Tuplet{actual: 8, normal: 6} }

// Nonuplet returns the nonuplet 9:8 (Russian "нонемоль").
func Nonuplet() Tuplet { return Tuplet{actual: 9, normal: 8} }

// Decuplet returns the decuplet 10:8 (Russian "децимоль").
func Decuplet() Tuplet { return Tuplet{actual: 10, normal: 8} }

// IsSet reports whether t is a set tuplet rather than the zero value.
func (t Tuplet) IsSet() bool {
	return t.actual != 0
}

// Actual returns the actual count n of t. It returns an error wrapping
// ErrUnset if t is unset.
func (t Tuplet) Actual() (int, error) {
	if !t.IsSet() {
		return 0, unsetError("receiver")
	}
	return t.actual, nil
}

// MustActual is like Actual but panics with the returned error.
func (t Tuplet) MustActual() int {
	return must(t.Actual())
}

// Normal returns the normal count m of t. It returns an error wrapping
// ErrUnset if t is unset.
func (t Tuplet) Normal() (int, error) {
	if !t.IsSet() {
		return 0, unsetError("receiver")
	}
	return t.normal, nil
}

// MustNormal is like Normal but panics with the returned error.
func (t Tuplet) MustNormal() int {
	return must(t.Normal())
}

// Pair returns the actual and normal counts of t. It returns an error
// wrapping ErrUnset if t is unset.
func (t Tuplet) Pair() (actual, normal int, err error) {
	if !t.IsSet() {
		return 0, 0, unsetError("receiver")
	}
	return t.actual, t.normal, nil
}

// MustPair is like Pair but panics with the returned error.
func (t Tuplet) MustPair() (actual, normal int) {
	actual, normal, err := t.Pair()
	if err != nil {
		panic(err)
	}
	return actual, normal
}

// Kind returns Compressing if actual > normal, Expanding if actual < normal,
// and Neutral if they are equal. It returns an error wrapping ErrUnset if t is
// unset.
func (t Tuplet) Kind() (Kind, error) {
	if !t.IsSet() {
		return 0, unsetError("receiver")
	}
	switch {
	case t.actual > t.normal:
		return Compressing, nil
	case t.actual < t.normal:
		return Expanding, nil
	default:
		return Neutral, nil
	}
}

// MustKind is like Kind but panics with the returned error.
func (t Tuplet) MustKind() Kind {
	return must(t.Kind())
}

// IsCompressing reports whether actual > normal. It returns an error wrapping
// ErrUnset if t is unset.
func (t Tuplet) IsCompressing() (bool, error) {
	k, err := t.Kind()
	return k == Compressing, err
}

// MustIsCompressing is like IsCompressing but panics with the returned error.
func (t Tuplet) MustIsCompressing() bool {
	return must(t.IsCompressing())
}

// IsExpanding reports whether actual < normal. It returns an error wrapping
// ErrUnset if t is unset.
func (t Tuplet) IsExpanding() (bool, error) {
	k, err := t.Kind()
	return k == Expanding, err
}

// MustIsExpanding is like IsExpanding but panics with the returned error.
func (t Tuplet) MustIsExpanding() bool {
	return must(t.IsExpanding())
}

// IsNeutral reports whether actual == normal. It returns an error wrapping
// ErrUnset if t is unset.
func (t Tuplet) IsNeutral() (bool, error) {
	k, err := t.Kind()
	return k == Neutral, err
}

// MustIsNeutral is like IsNeutral but panics with the returned error.
func (t Tuplet) MustIsNeutral() bool {
	return must(t.IsNeutral())
}

// abs returns the magnitude of n; n must not be math.MinInt.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
