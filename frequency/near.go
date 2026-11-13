package frequency

import (
	"cmp"
	"math/big"

	"github.com/go-muse/musex/cents"
)

// Equal reports whether f1 and f2 are exactly equal float64 values, without
// an epsilon, ULP tolerance, decimal interpretation, or conversion to cents.
// It validates f1 and then f2, returning ErrInvalidFrequency and
// cents.ErrInvalidArgument for the first invalid frequency. On valid inputs
// it performs no allocation.
func Equal(f1, f2 float64) (bool, error) {
	if err := validateFrequency(f1, 1); err != nil {
		return false, err
	}
	if err := validateFrequency(f2, 2); err != nil {
		return false, err
	}
	return f1 == f2, nil
}

// MustEqual calls Equal and panics with its error, if any.
func MustEqual(f1, f2 float64) bool {
	equal, err := Equal(f1, f2)
	if err != nil {
		panic(err)
	}
	return equal
}

// Compare returns -1, 0, or +1 according to the strict float64 order of f1
// and f2. It validates f1 and then f2, returning ErrInvalidFrequency and
// cents.ErrInvalidArgument for the first invalid frequency. On valid inputs
// it performs no allocation.
func Compare(f1, f2 float64) (int, error) {
	if err := validateFrequency(f1, 1); err != nil {
		return 0, err
	}
	if err := validateFrequency(f2, 2); err != nil {
		return 0, err
	}
	return cmp.Compare(f1, f2), nil
}

// MustCompare calls Compare and panics with its error, if any.
func MustCompare(f1, f2 float64) int {
	comparison, err := Compare(f1, f2)
	if err != nil {
		panic(err)
	}
	return comparison
}

// Near reports whether f1 and f2 are close according to c. It checks for an
// unset criterion, then validates f1 and f2 in that order. A valid criterion
// never fails for valid frequencies. Composite criteria are evaluated
// iteratively with short-circuiting and no recursion.
func (c Criterion) Near(f1, f2 float64) (bool, error) {
	if !c.IsValid() {
		return false, ErrUnset
	}
	if err := validateFrequency(f1, 1); err != nil {
		return false, err
	}
	if err := validateFrequency(f2, 2); err != nil {
		return false, err
	}
	if !c.kind.IsComposite() {
		return nearSimple(c, f1, f2, &nearContext{})
	}

	type frame struct {
		criterion Criterion
		next      int
		result    bool
	}
	root := frame{criterion: c, result: c.kind == KindAllOf}
	stack := []frame{root}
	context := nearContext{}
	for len(stack) > 0 {
		current := &stack[len(stack)-1]
		if current.criterion.kind.IsComposite() && current.next < len(current.criterion.parts) {
			child := current.criterion.parts[current.next]
			current.next++
			stack = append(stack, frame{criterion: child, result: child.kind == KindAllOf})
			continue
		}

		result := current.result
		if !current.criterion.kind.IsComposite() {
			var err error
			result, err = nearSimple(current.criterion, f1, f2, &context)
			if err != nil {
				return false, err
			}
		}
		stack = stack[:len(stack)-1]
		if len(stack) == 0 {
			return result, nil
		}

		parent := &stack[len(stack)-1]
		if parent.criterion.kind == KindAllOf {
			parent.result = parent.result && result
			if !parent.result {
				parent.next = len(parent.criterion.parts)
			}
		} else {
			parent.result = parent.result || result
			if parent.result {
				parent.next = len(parent.criterion.parts)
			}
		}
	}
	return false, nil
}

// MustNear calls Near and panics with its error, if any.
func (c Criterion) MustNear(f1, f2 float64) bool {
	near, err := c.Near(f1, f2)
	if err != nil {
		panic(err)
	}
	return near
}

// Compare returns 0 when f1 and f2 are close according to c; otherwise it
// returns their strict order, -1 or +1. This is not a sorting order when c has
// a tolerance because closeness is generally not transitive. Error order is
// the same as for Near.
func (c Criterion) Compare(f1, f2 float64) (int, error) {
	near, err := c.Near(f1, f2)
	if err != nil {
		return 0, err
	}
	if near {
		return 0, nil
	}
	return cmp.Compare(f1, f2), nil
}

// MustCompare calls Compare and panics with its error, if any.
func (c Criterion) MustCompare(f1, f2 float64) int {
	comparison, err := c.Compare(f1, f2)
	if err != nil {
		panic(err)
	}
	return comparison
}

func validateFrequency(f float64, position int) error {
	if !cents.IsValidFrequency(f) {
		return invalidFrequency(position)
	}
	return nil
}

type nearContext struct {
	centsDistance *cents.Cents
	differenceRat *big.Rat
}

func (ctx *nearContext) distance(f1, f2 float64) (cents.Cents, error) {
	if ctx.centsDistance == nil {
		distance, err := cents.Between(f1, f2)
		if err != nil {
			return cents.Cents{}, invalidFrequency(1)
		}
		ctx.centsDistance = &distance
	}
	return *ctx.centsDistance, nil
}

func (ctx *nearContext) difference(f1, f2 float64) *big.Rat {
	if ctx.differenceRat == nil {
		left := new(big.Rat).SetFloat64(f1)
		right := new(big.Rat).SetFloat64(f2)
		ctx.differenceRat = new(big.Rat).Sub(left, right)
		ctx.differenceRat.Abs(ctx.differenceRat)
	}
	return ctx.differenceRat
}

func nearSimple(c Criterion, f1, f2 float64, context *nearContext) (bool, error) {
	switch c.kind {
	case KindStrict:
		return f1 == f2, nil
	case KindCents:
		distance, err := context.distance(f1, f2)
		if err != nil {
			return false, err
		}
		return distance.Abs().Compare(c.tolerance) <= 0, nil
	case KindHertz:
		tolerance := new(big.Rat).SetFloat64(c.number)
		return context.difference(f1, f2).Cmp(tolerance) <= 0, nil
	case KindRelative:
		if f1 == f2 || c.number >= 1 {
			return true, nil
		}
		tolerance := new(big.Rat).SetFloat64(c.number)
		maxFrequency := f1
		if f2 > maxFrequency {
			maxFrequency = f2
		}
		tolerance.Mul(tolerance, new(big.Rat).SetFloat64(maxFrequency))
		return context.difference(f1, f2).Cmp(tolerance) <= 0, nil
	default:
		return false, ErrUnset
	}
}
