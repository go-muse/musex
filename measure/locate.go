package measure

import (
	"math/big"

	"github.com/go-muse/musex/duration"
)

// checkPositionArg checks the receiver and a position argument: ErrUnset;
// ErrInvalidArgument together with duration.ErrUnset for an unset position;
// then the end of the grid as checkPosition does.
func (g Grid) checkPositionArg(op string, p duration.Position, allowEnd bool) (*big.Rat, error) {
	if !g.IsSet() {
		return nil, errUnsetGrid(op)
	}
	if !p.IsSet() {
		return nil, errUnsetTime(op, "position")
	}
	r := ratOfPosition(p)
	if err := g.checkPosition(op, r, allowEnd); err != nil {
		return nil, err
	}
	return r, nil
}

// Locate returns the address of the position p: the bar whose start is not
// after p and whose end is after p, and the offset of p from the start of
// the bar. A position on a bar line belongs to the next bar.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset p; ErrOutsideGrid if p is not
// before the end of a finite grid; ErrOutOfRange if p is not before the end
// of bar math.MaxInt of an open grid, then together with
// fraction.ErrOutOfRange if the offset is outside the range of
// common/fraction.
//
// Example: in "3/4;pickup:1/4", 0/1 is 0:0/1, 1/4 is 1:0/1, and 9/8 is
// 2:1/8.
func (g Grid) Locate(p duration.Position) (Address, error) {
	const op = "Grid.Locate"
	r, err := g.checkPositionArg(op, p, false)
	if err != nil {
		return Address{}, err
	}
	return g.locate(op, r)
}

// locate returns the address of the position r inside the grid.
func (g Grid) locate(op string, r *big.Rat) (Address, error) {
	_, k, offset := g.locateRat(r)
	o, err := toPosition(op, offset)
	if err != nil {
		return Address{}, err
	}
	return Address{bar: k, offset: o}, nil
}

// MustLocate is like Locate but panics with the error of Locate.
func (g Grid) MustLocate(p duration.Position) Address {
	return must(g.Locate(p))
}

// BarAt returns the description of the bar that contains the position p, as
// Bar returns it; its Start and End are the start and the end of the bar.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset p; ErrOutsideGrid if p is not
// before the end of a finite grid; ErrOutOfRange if p is not before the end
// of bar math.MaxInt of an open grid, then together with
// fraction.ErrOutOfRange if the start, the end, or the shift of the bar is
// outside the range of common/fraction.
func (g Grid) BarAt(p duration.Position) (Bar, error) {
	const op = "Grid.BarAt"
	r, err := g.checkPositionArg(op, p, false)
	if err != nil {
		return Bar{}, err
	}
	i, k, offset := g.locateRat(r)
	return barOf(op, &g.segs[i], k, ratSub(r, offset))
}

// MustBarAt is like BarAt but panics with the error of BarAt.
func (g Grid) MustBarAt(p duration.Position) Bar {
	return must(g.BarAt(p))
}

// Remaining returns the rest of the bar from the position p: the end of the
// bar that contains p minus p, always positive.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset p; ErrOutsideGrid if p is not
// before the end of a finite grid; ErrOutOfRange if p is not before the end
// of bar math.MaxInt of an open grid, then together with
// fraction.ErrOutOfRange if the rest is outside the range of
// common/fraction.
//
// Example: in 3/4, the rest from 1/4 is 1/2.
func (g Grid) Remaining(p duration.Position) (duration.Duration, error) {
	const op = "Grid.Remaining"
	r, err := g.checkPositionArg(op, p, false)
	if err != nil {
		return duration.Duration{}, err
	}
	i, _, offset := g.locateRat(r)
	return toDuration(op, ratSub(g.segs[i].length, offset))
}

// MustRemaining is like Remaining but panics with the error of Remaining.
func (g Grid) MustRemaining(p duration.Position) duration.Duration {
	return must(g.Remaining(p))
}

// OnBarline reports whether the position p is on a bar line: the start of a
// bar or the end of a finite grid.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset p; ErrOutsideGrid if p is
// after the end of a finite grid; ErrOutOfRange if p is not before the end
// of bar math.MaxInt of an open grid, where a bar with a number greater than
// math.MaxInt would start.
//
// Example: in 3/4, 0/1 and 3/4 are on bar lines and 1/4 is not.
func (g Grid) OnBarline(p duration.Position) (bool, error) {
	const op = "Grid.OnBarline"
	r, err := g.checkPositionArg(op, p, true)
	if err != nil {
		return false, err
	}
	if g.count > 0 && r.Cmp(g.end) == 0 {
		return true, nil
	}
	_, _, offset := g.locateRat(r)
	return offset.Sign() == 0, nil
}

// MustOnBarline is like OnBarline but panics with the error of OnBarline.
func (g Grid) MustOnBarline(p duration.Position) bool {
	return must(g.OnBarline(p))
}

// checkAddress checks the receiver and an address argument: ErrUnset for the
// unset grid, then for the unset address.
func (g Grid) checkAddress(op string, a Address) error {
	if !g.IsSet() {
		return errUnsetGrid(op)
	}
	if !a.IsSet() {
		return errUnset(op, "address")
	}
	return nil
}

// addressInBar checks that the bar of a exists and that its offset is less
// than the length of the bar: ErrOutsideGrid, then ErrOutsideBar. It returns
// the segment of the bar and the offset.
func (g Grid) addressInBar(op string, a Address) (*segment, *big.Rat, error) {
	if err := g.checkExists(op, a.bar); err != nil {
		return nil, nil, err
	}
	s := &g.segs[g.segIndex(a.bar)]
	offset := ratOfPosition(a.offset)
	if offset.Cmp(s.length) >= 0 {
		return nil, nil, errSimple(ErrOutsideBar, op, "the offset is not less than the length of the bar")
	}
	return s, offset, nil
}

// PositionOf returns the position of the address a: the start of its bar
// plus its offset. The address must be valid: Normalize converts any offset.
//
// Errors, in this order: ErrUnset for the unset grid, then for the unset
// address; ErrOutsideGrid if g has no bar of a; ErrOutsideBar if the offset
// is not less than the length of the bar; ErrOutOfRange together with
// fraction.ErrOutOfRange if the position is outside the range of
// common/fraction.
//
// Example: in "3/4;pickup:1/4", 1:1/2 is at 3/4.
func (g Grid) PositionOf(a Address) (duration.Position, error) {
	const op = "Grid.PositionOf"
	if err := g.checkAddress(op, a); err != nil {
		return duration.Position{}, err
	}
	s, offset, err := g.addressInBar(op, a)
	if err != nil {
		return duration.Position{}, err
	}
	start := ratAdd(s.start, ratMulInt(s.length, a.bar-s.first))
	return toPosition(op, ratAdd(start, offset))
}

// MustPositionOf is like PositionOf but panics with the error of PositionOf.
func (g Grid) MustPositionOf(a Address) duration.Position {
	return must(g.PositionOf(a))
}

// Contains reports whether the address a is valid in g: its bar exists and
// its offset is less than the length of the bar. An invalid address is a
// result, not an error.
//
// Errors, in this order: ErrUnset for the unset grid, then for the unset
// address.
//
// Example: in "4/4;bars:4", 1:1/2 is valid, and 1:1/1, 5:0/1, and 0:0/1 are
// not.
func (g Grid) Contains(a Address) (bool, error) {
	if err := g.checkAddress("Grid.Contains", a); err != nil {
		return false, err
	}
	if !g.hasBar(a.bar) {
		return false, nil
	}
	s := &g.segs[g.segIndex(a.bar)]
	return ratOfPosition(a.offset).Cmp(s.length) < 0, nil
}

// MustContains is like Contains but panics with the error of Contains.
func (g Grid) MustContains(a Address) bool {
	return must(g.Contains(a))
}

// Normalize returns the valid address of the moment of a: the bar of a must
// exist, and an offset not less than the length of the bar carries over into
// the next bars.
//
// Errors, in this order: ErrUnset for the unset grid, then for the unset
// address; ErrOutsideGrid if g has no bar of a, then if the moment is not
// before the end of a finite grid; ErrOutOfRange if the moment is not before
// the end of bar math.MaxInt of an open grid, then together with
// fraction.ErrOutOfRange if the offset of the result is outside the range of
// common/fraction.
//
// Example: in 4/4, 2:5/4 is 3:1/4; in "3/4;pickup:1/4", 0:1/1 is 2:0/1.
func (g Grid) Normalize(a Address) (Address, error) {
	const op = "Grid.Normalize"
	if err := g.checkAddress(op, a); err != nil {
		return Address{}, err
	}
	if err := g.checkExists(op, a.bar); err != nil {
		return Address{}, err
	}
	r := ratAdd(g.barStartRat(a.bar), ratOfPosition(a.offset))
	if err := g.checkPosition(op, r, false); err != nil {
		return Address{}, err
	}
	return g.locate(op, r)
}

// MustNormalize is like Normalize but panics with the error of Normalize.
func (g Grid) MustNormalize(a Address) Address {
	return must(g.Normalize(a))
}
