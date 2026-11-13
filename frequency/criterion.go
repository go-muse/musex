package frequency

import (
	"math"
	"slices"

	"github.com/go-muse/musex/cents"
)

// Criterion describes when two frequencies are close. Its zero value is an
// unset criterion, not the strict criterion. Criterion is immutable and not
// comparable; use Equal for equality and String for a map key. Copies share
// only immutable internal data.
type Criterion struct {
	kind      Kind
	number    float64
	tolerance cents.Cents
	parts     []Criterion
}

// Strict returns the strict criterion: frequencies are close exactly when
// their float64 values are equal. It has no hidden tolerance.
func Strict() Criterion {
	return Criterion{kind: KindStrict}
}

// WithinCents returns a criterion whose tolerance is t cents. Closeness uses
// the shortest-decimal, 10^-6-cent-grid distance provided by cents.Between.
// It returns ErrInvalidTolerance if t is negative.
func WithinCents(t cents.Cents) (Criterion, error) {
	if t.Sign() < 0 {
		return Criterion{}, invalidTolerance()
	}
	return Criterion{kind: KindCents, tolerance: t}, nil
}

// MustWithinCents calls WithinCents and panics with its error, if any.
func MustWithinCents(t cents.Cents) Criterion {
	c, err := WithinCents(t)
	if err != nil {
		panic(err)
	}
	return c
}

// WithinHertz returns a criterion with an absolute tolerance in hertz. The
// exact binary values of the inputs are compared without float64 rounding
// error. It returns ErrInvalidTolerance for a negative, NaN, or infinite t;
// negative zero is stored as positive zero.
func WithinHertz(t float64) (Criterion, error) {
	return withinNumber(KindHertz, t)
}

// MustWithinHertz calls WithinHertz and panics with its error, if any.
func MustWithinHertz(t float64) Criterion {
	c, err := WithinHertz(t)
	if err != nil {
		panic(err)
	}
	return c
}

// WithinRelative returns a criterion with tolerance t times the greater
// frequency. Values t >= 1 make every pair of valid frequencies close. It
// returns ErrInvalidTolerance for a negative, NaN, or infinite t; negative
// zero is stored as positive zero.
func WithinRelative(t float64) (Criterion, error) {
	return withinNumber(KindRelative, t)
}

// MustWithinRelative calls WithinRelative and panics with its error, if any.
func MustWithinRelative(t float64) Criterion {
	c, err := WithinRelative(t)
	if err != nil {
		panic(err)
	}
	return c
}

func withinNumber(kind Kind, t float64) (Criterion, error) {
	if t < 0 || math.IsNaN(t) || math.IsInf(t, 0) {
		return Criterion{}, invalidTolerance()
	}
	if t == 0 {
		t = 0 // Normalize negative zero.
	}
	return Criterion{kind: kind, number: t}, nil
}

// AllOf returns a criterion that reports closeness only when every component
// does. It returns ErrNoComponents for an empty list, then ErrUnset for the
// first unset component. The input slice is copied and is not retained.
func AllOf(cs ...Criterion) (Criterion, error) {
	return composite(KindAllOf, cs)
}

// MustAllOf calls AllOf and panics with its error, if any.
func MustAllOf(cs ...Criterion) Criterion {
	c, err := AllOf(cs...)
	if err != nil {
		panic(err)
	}
	return c
}

// AnyOf returns a criterion that reports closeness when at least one component
// does. It returns ErrNoComponents for an empty list, then ErrUnset for the
// first unset component. The input slice is copied and is not retained.
func AnyOf(cs ...Criterion) (Criterion, error) {
	return composite(KindAnyOf, cs)
}

// MustAnyOf calls AnyOf and panics with its error, if any.
func MustAnyOf(cs ...Criterion) Criterion {
	c, err := AnyOf(cs...)
	if err != nil {
		panic(err)
	}
	return c
}

func composite(kind Kind, cs []Criterion) (Criterion, error) {
	if len(cs) == 0 {
		return Criterion{}, ErrNoComponents
	}
	for _, c := range cs {
		if !c.IsValid() {
			return Criterion{}, ErrUnset
		}
	}
	return Criterion{kind: kind, parts: slices.Clone(cs)}, nil
}

// IsValid reports whether c is a criterion created by this package rather
// than the unset zero value.
func (c Criterion) IsValid() bool {
	return c.kind.IsValid()
}

// Kind returns the kind of c, or zero for the unset criterion.
func (c Criterion) Kind() Kind {
	return c.kind
}

// Tolerance returns the numeric tolerance and true for KindHertz and
// KindRelative. For all other kinds, including KindCents, it returns (0,
// false); use CentsTolerance for a cents tolerance.
func (c Criterion) Tolerance() (float64, bool) {
	if c.kind == KindHertz || c.kind == KindRelative {
		return c.number, true
	}
	return 0, false
}

// CentsTolerance returns the tolerance and true only for KindCents. For other
// kinds it returns the zero cents.Cents and false; see Tolerance for numeric
// tolerances.
func (c Criterion) CentsTolerance() (cents.Cents, bool) {
	if c.kind == KindCents {
		return c.tolerance, true
	}
	return cents.Cents{}, false
}

// Components returns a new slice of the direct components of a composite
// criterion in construction order. For other criteria it returns nil.
// Modifying the result does not modify c.
func (c Criterion) Components() []Criterion {
	if !c.kind.IsComposite() {
		return nil
	}
	return slices.Clone(c.parts)
}

// Equal reports whether c and other have the same kind, tolerance, and ordered
// components. Criteria with equivalent behavior but different descriptions,
// such as Strict and WithinHertz(0), are not equal. The traversal is iterative
// and supports arbitrarily deep nesting without recursion.
func (c Criterion) Equal(other Criterion) bool {
	if !c.kind.IsComposite() || !other.kind.IsComposite() {
		return equalSimple(c, other)
	}
	if !equalSimple(c, other) || len(c.parts) != len(other.parts) {
		return false
	}
	type pair struct {
		left  Criterion
		right Criterion
	}
	var stack []pair
	for i := range c.parts {
		stack = append(stack, pair{c.parts[i], other.parts[i]})
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		p := stack[last]
		stack = stack[:last]
		if !equalSimple(p.left, p.right) {
			return false
		}
		if !p.left.kind.IsComposite() {
			continue
		}
		if len(p.left.parts) != len(p.right.parts) {
			return false
		}
		for i := range p.left.parts {
			stack = append(stack, pair{p.left.parts[i], p.right.parts[i]})
		}
	}
	return true
}

func equalSimple(c, other Criterion) bool {
	if c.kind != other.kind {
		return false
	}
	switch c.kind {
	case 0, KindStrict:
		return true
	case KindCents:
		return c.tolerance.Equal(other.tolerance)
	case KindHertz, KindRelative:
		return c.number == other.number
	case KindAllOf, KindAnyOf:
		return true
	default:
		return false
	}
}
