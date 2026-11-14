package meter

// variantsLimit is the smallest number of units with more than
// MaxGroupingVariants grouping variants: 27 units have 816 variants and 28
// units have 1081. The number of variants does not decrease as the number of
// units grows from 2, so the limit is checked without building the list.
const variantsLimit = 28

// WithGrouping returns a meter with the same record of the signature and the
// explicit grouping groups, which replaces a previous one. Only a simple
// signature and a symbol take an explicit grouping. Any groups of at least 1
// whose sum equals the numerator are accepted, including one group, such as
// 5/8[5], and a nonstandard grouping, such as 6/8[2+2+2]. A grouping equal to
// the standard one gives the meter without an explicit grouping: 6/8 with
// 3, 3 is 6/8. The slice groups is only read during the call.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument for
// a compound numerator or a sum, then for no groups, for the first group less
// than 1, and for groups whose sum differs from the numerator.
//
// Example: MustNew(7, 8).WithGrouping(3, 2, 2) is 7/8[3+2+2].
func (m Meter) WithGrouping(groups ...int) (Meter, error) {
	const op = "Meter.WithGrouping"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if f := r.form(); f != FormSimple && f != FormSymbol {
		return Meter{}, invalidError(op, "explicit grouping of a compound numerator or a sum")
	}
	if len(groups) == 0 {
		return Meter{}, invalidError(op, "no groups")
	}
	for _, g := range groups {
		if g < 1 {
			return Meter{}, invalidError(op, "group less than 1")
		}
	}
	if !sumEquals(groups, r.numerator()) {
		return Meter{}, invalidError(op, "sum of groups differs from the numerator")
	}
	return build(op, r.withGrouping(groups))
}

// MustWithGrouping is like WithGrouping but panics with the error of
// WithGrouping.
func (m Meter) MustWithGrouping(groups ...int) Meter {
	return must(m.WithGrouping(groups...))
}

// sumEquals reports whether the sum of ns, each at least 1, equals n, at
// least 1. It never overflows.
func sumEquals(ns []int, n int) bool {
	rest := n
	for _, x := range ns {
		if x > rest {
			return false
		}
		rest -= x
	}
	return rest == 0
}

// ResetGrouping returns m without its explicit grouping: the same signature
// with the standard grouping or without a grouping (5/8[2+3] gives 5/8,
// 6/8[2+2+2] gives 6/8). A meter of any form without an explicit grouping is
// returned unchanged.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) ResetGrouping() (Meter, error) {
	const op = "Meter.ResetGrouping"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if r.explicit == nil {
		return m, nil
	}
	r.explicit = nil
	return build(op, r)
}

// MustResetGrouping is like ResetGrouping but panics with the error of
// ResetGrouping.
func (m Meter) MustResetGrouping() Meter {
	return must(m.ResetGrouping())
}

// GroupingVariants returns every way to split units into groups of two and
// three units, with the order of groups taken into account. The variants are
// in lexicographic order of their group sizes: 7 gives [[2 2 3] [2 3 2] [3 2
// 2]], 8 gives [[2 2 2 2] [2 3 3] [3 2 3] [3 3 2]]. One unit has no variants:
// the result is nil without an error. The result is new at every call.
//
// Errors, in this order: ErrInvalidArgument for units less than 1;
// ErrTooMany for more than MaxGroupingVariants variants, that is, for units
// of 28 or more, without building the list.
func GroupingVariants(units int) ([][]int, error) {
	const op = "meter.GroupingVariants"
	if units < 1 {
		return nil, invalidError(op, "number of units less than 1")
	}
	if units >= variantsLimit {
		return nil, simpleError(ErrTooMany, op, "more than MaxGroupingVariants variants")
	}
	return groupingVariants(units), nil
}

// MustGroupingVariants is like GroupingVariants but panics with the error of
// GroupingVariants.
func MustGroupingVariants(units int) [][]int {
	return must(GroupingVariants(units))
}

// groupingVariants lists the splits of n units, 1 ≤ n < variantsLimit, into
// groups of two and three in lexicographic order. It returns nil for n = 1.
func groupingVariants(n int) [][]int {
	var out [][]int
	var walk func(prefix []int, rest int)
	walk = func(prefix []int, rest int) {
		if rest == 0 {
			out = append(out, append([]int(nil), prefix...))
			return
		}
		for _, g := range [...]int{2, 3} {
			if g <= rest {
				walk(append(prefix, g), rest-g)
			}
		}
	}
	walk(make([]int, 0, n/2), n)
	return out
}

// GroupingVariants returns the grouping variants of the number of units of m,
// as the function GroupingVariants does. It does not depend on the grouping
// of m.
//
// Errors, in this order: ErrUnset for the unset meter; ErrTooMany for more
// than MaxGroupingVariants variants.
//
// Example: the variants of 5/8 are [[2 3] [3 2]].
func (m Meter) GroupingVariants() ([][]int, error) {
	const op = "Meter.GroupingVariants"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	n := r.numerator()
	if n >= variantsLimit {
		return nil, simpleError(ErrTooMany, op, "more than MaxGroupingVariants variants")
	}
	return groupingVariants(n), nil
}

// MustGroupingVariants is like GroupingVariants but panics with the error of
// GroupingVariants.
func (m Meter) MustGroupingVariants() [][]int {
	return must(m.GroupingVariants())
}
