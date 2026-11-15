package measure

import (
	"slices"
)

// Equal reports whether g and other are equal by record: the same pickup
// (its presence and length), the same marks (bars and meters by
// meter.Meter.Equal), the same explicit lengths (bars, lengths, and
// alignments), and the same number of bars (its presence and value). The
// unset grid equals only itself. A grid is not comparable with ==.
//
// Example: "4/4" and "4/4;5:4/4" differ by record.
func (g Grid) Equal(other Grid) bool {
	if !g.IsSet() || !other.IsSet() {
		return g.IsSet() == other.IsSet()
	}
	return g.pickup == other.pickup &&
		g.count == other.count &&
		slices.Equal(g.marks, other.marks) &&
		slices.Equal(g.lengths, other.lengths)
}

// run is a run of bars from first with equal compared properties.
type run struct {
	first int
	seg   *segment
}

// runs returns the segments of g with adjacent segments merged when same
// reports that their compared properties are equal.
func (g Grid) runs(same func(a, b *segment) bool) []run {
	out := make([]run, 0, len(g.segs))
	for i := range g.segs {
		s := &g.segs[i]
		if n := len(out); n > 0 && !out[n-1].seg.pickup && !s.pickup && same(out[n-1].seg, s) {
			continue
		}
		out = append(out, run{first: s.first, seg: s})
	}
	return out
}

// sameMarkup reports whether two segments have the same meter by record,
// length, and shift.
func sameMarkup(a, b *segment) bool {
	return a.meter.Equal(b.meter) && a.length.Cmp(b.length) == 0 && a.shift.Cmp(b.shift) == 0
}

// sameLength reports whether two segments have the same length.
func sameLength(a, b *segment) bool {
	return a.length.Cmp(b.length) == 0
}

// equalRuns reports whether g and other have the same pickup presence and
// number of bars and equal runs by same.
func (g Grid) equalRuns(other Grid, same func(a, b *segment) bool) bool {
	if g.pickup.IsSet() != other.pickup.IsSet() || g.count != other.count {
		return false
	}
	a, b := g.runs(same), other.runs(same)
	return slices.EqualFunc(a, b, func(x, y run) bool {
		return x.first == y.first && same(x.seg, y.seg)
	})
}

// EqualMarkup reports whether g and other are equal by markup: the same
// pickup and number of bars, and every bar has the same effective meter by
// record, length, and shift in the meter. Repeated marks and explicit
// lengths equal to the size of the meter do not matter.
//
// Errors: ErrUnset for the unset grid, then for an unset other.
//
// Example: "4/4" and "4/4;5:4/4" are equal by markup; "3/4" and "6/8" are
// not.
func (g Grid) EqualMarkup(other Grid) (bool, error) {
	const op = "Grid.EqualMarkup"
	if !g.IsSet() {
		return false, errUnsetGrid(op)
	}
	if !other.IsSet() {
		return false, errUnset(op, "grid argument")
	}
	return g.equalRuns(other, sameMarkup), nil
}

// MustEqualMarkup is like EqualMarkup but panics with the error of
// EqualMarkup.
func (g Grid) MustEqualMarkup(other Grid) bool {
	return must(g.EqualMarkup(other))
}

// EqualBarLengths reports whether g and other are equal by bar lengths: the
// same pickup (its presence and length) and number of bars, and every bar
// number has the same length; the meters can differ.
//
// Errors: ErrUnset for the unset grid, then for an unset other.
//
// Example: "3/4" and "6/8" are equal by bar lengths.
func (g Grid) EqualBarLengths(other Grid) (bool, error) {
	const op = "Grid.EqualBarLengths"
	if !g.IsSet() {
		return false, errUnsetGrid(op)
	}
	if !other.IsSet() {
		return false, errUnset(op, "grid argument")
	}
	return g.equalRuns(other, sameLength), nil
}

// MustEqualBarLengths is like EqualBarLengths but panics with the error of
// EqualBarLengths.
func (g Grid) MustEqualBarLengths(other Grid) bool {
	return must(g.EqualBarLengths(other))
}
