package instrument

import (
	"fmt"
	"iter"
	"math"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/system"
)

// Range is a reference range: two pitches, the lower bound and the upper
// bound, both included. It is an immutable, comparable value.
//
// A range does not hold a system: every operation that needs positions takes
// the system explicitly, so one range can be valid in one system and inverted
// in another. A range is valid in a system when the names of both bounds are
// admissible in it and the lower bound is not above the upper one by position.
// The bounds are kept as written and are never swapped.
//
// The zero Range is the unset range, which means "no range": IsValid reports
// false, Lower and Upper return pitch.Pitch{}, String returns "Range(unset)",
// and every operation in a system fails with an error that matches ErrNoRange.
// For any two ranges, == agrees with Equal.
type Range struct {
	lower pitch.Pitch
	upper pitch.Pitch
}

// NewRange returns the range from lower to upper. It checks only that both
// pitches are set; the range is checked in a system by the operations that
// take one, so NewRange(G5, C4) is created and is inverted in every system
// that admits both names.
//
// It fails with an error that matches ErrUnsetPitch when lower is unset,
// then when upper is unset.
func NewRange(lower, upper pitch.Pitch) (Range, error) {
	const op = "new range"

	if !lower.IsValid() {
		return Range{}, unsetPitchError(op, "lower bound")
	}

	if !upper.IsValid() {
		return Range{}, unsetPitchError(op, "upper bound")
	}

	return Range{lower: lower, upper: upper}, nil
}

// MustNewRange is like NewRange but panics with its error on failure.
func MustNewRange(lower, upper pitch.Pitch) Range {
	r, err := NewRange(lower, upper)
	if err != nil {
		panic(err)
	}

	return r
}

// RangeOf returns the range of a set of pitches in sys: the lowest of ps by
// position is the lower bound and the highest is the upper bound. Of pitches
// at one position the first in the order of ps is taken. RangeOf(sys, x) is
// the range from x to x.
//
// Errors, in this order: ErrUnset when sys is unset; ErrNoPitches when ps is
// empty; then for each pitch in the order of ps, ErrUnsetPitch when it is
// unset and ErrInadmissible when its name is not admissible in sys.
func RangeOf(sys *system.System, ps ...pitch.Pitch) (Range, error) {
	const op = "range of"

	if !sys.IsValid() {
		return Range{}, unsetError(op, "system")
	}

	if len(ps) == 0 {
		return Range{}, noPitchesError(op)
	}

	for _, p := range ps {
		if err := checkPitch(op, "pitch", sys, p); err != nil {
			return Range{}, err
		}
	}

	r := Range{lower: ps[0], upper: ps[0]}
	for _, p := range ps[1:] {
		below, err := sys.Compare(p, r.lower)
		if err != nil {
			return Range{}, wrapError(op, err)
		}

		if below < 0 {
			r.lower = p
		}

		above, err := sys.Compare(p, r.upper)
		if err != nil {
			return Range{}, wrapError(op, err)
		}

		if above > 0 {
			r.upper = p
		}
	}

	return r, nil
}

// MustRangeOf is like RangeOf but panics with its error on failure.
func MustRangeOf(sys *system.System, ps ...pitch.Pitch) Range {
	r, err := RangeOf(sys, ps...)
	if err != nil {
		panic(err)
	}

	return r
}

// IsValid reports whether r is set.
func (r Range) IsValid() bool {
	return r.lower.IsValid()
}

// Lower returns the lower bound as written; pitch.Pitch{} for the unset range.
func (r Range) Lower() pitch.Pitch {
	return r.lower
}

// Upper returns the upper bound as written; pitch.Pitch{} for the unset range.
func (r Range) Upper() pitch.Pitch {
	return r.upper
}

// Equal reports whether r and other are equal as written: both unset, or both
// set with equal lower bounds and equal upper bounds (pitch.Pitch.Equal).
// Ranges with the same positions and other spellings are not equal; see
// Equivalent.
func (r Range) Equal(other Range) bool {
	if !r.IsValid() || !other.IsValid() {
		return r.IsValid() == other.IsValid()
	}

	return r.lower.Equal(other.lower) && r.upper.Equal(other.upper)
}

// String returns a debugging form with both bounds, such as
// "Range(C ASPN 4 – G ASPN 5)", or "Range(unset)". It is not a notation, has
// no parser and does not affect equality.
func (r Range) String() string {
	if !r.IsValid() {
		return "Range(unset)"
	}

	return "Range(" + r.lower.String() + " – " + r.upper.String() + ")"
}

// Check reports whether r is valid in sys: nil when both names are admissible
// in sys and the lower bound is not above the upper one.
//
// Errors, in this order: ErrUnset when sys is unset; ErrNoRange when r is
// unset; ErrInadmissible for the lower bound, then for the upper bound;
// ErrInverted when the lower bound is above the upper one in sys. Every
// operation of a range in a system checks its ranges in this order first.
func (r Range) Check(sys *system.System) error {
	return r.check("check", sys)
}

// MustCheck is like Check but panics with its error on failure; it returns r.
func (r Range) MustCheck(sys *system.System) Range {
	if err := r.Check(sys); err != nil {
		panic(err)
	}

	return r
}

// Contains reports whether x is in r in sys: whether the position of x is
// between the positions of the bounds, both included. Written octave numbers
// are not compared: in the traditional system B♯ at Octave 4 (written B♯3) is
// at the position of C4 and is in a range from C4.
//
// Errors: those of Check, then ErrUnsetPitch when x is unset and
// ErrInadmissible when the name of x is not admissible in sys.
func (r Range) Contains(sys *system.System, x pitch.Pitch) (bool, error) {
	loc, err := r.locate("contains", sys, x)
	if err != nil {
		return false, err
	}

	return loc == Inside, nil
}

// MustContains is like Contains but panics with its error on failure.
func (r Range) MustContains(sys *system.System, x pitch.Pitch) bool {
	ok, err := r.Contains(sys, x)
	if err != nil {
		panic(err)
	}

	return ok
}

// Locate returns where x is relative to r in sys: Below when its position is
// below the lower bound, Above when it is above the upper bound, and Inside
// otherwise. Errors are those of Contains; on error it returns LocationUnset.
func (r Range) Locate(sys *system.System, x pitch.Pitch) (Location, error) {
	return r.locate("locate", sys, x)
}

// MustLocate is like Locate but panics with its error on failure.
func (r Range) MustLocate(sys *system.System, x pitch.Pitch) Location {
	loc, err := r.Locate(sys, x)
	if err != nil {
		panic(err)
	}

	return loc
}

// Width returns the number of steps of sys from the lower bound to the upper
// bound: zero or more.
//
// Errors: those of Check, then an error that matches system.ErrOutOfRange
// when the width does not fit in int.
func (r Range) Width(sys *system.System) (system.Step, error) {
	return r.width("width", sys)
}

// MustWidth is like Width but panics with its error on failure.
func (r Range) MustWidth(sys *system.System) system.Step {
	w, err := r.Width(sys)
	if err != nil {
		panic(err)
	}

	return w
}

// Count returns the number of sounds of r in sys: the width plus one.
//
// Errors: those of Width; an error that matches system.ErrOutOfRange also
// when the width is math.MaxInt, so that the count does not fit in int.
func (r Range) Count(sys *system.System) (int, error) {
	const op = "count"

	w, err := r.width(op, sys)
	if err != nil {
		return 0, err
	}

	if w == math.MaxInt {
		return 0, wrapError(op, fmt.Errorf("%w: width %d plus one", system.ErrOutOfRange, w))
	}

	return int(w) + 1, nil
}

// MustCount is like Count but panics with its error on failure.
func (r Range) MustCount(sys *system.System) int {
	n, err := r.Count(sys)
	if err != nil {
		panic(err)
	}

	return n
}

// Folds returns all folds of x into r in sys: the pitches with the name of x
// and any Octave that are in r, from the lowest to the highest. There can be
// none. The sequence is lazy, can be stopped at any element, and yields the
// same pitches when iterated again or concurrently.
//
// All errors are reported before the sequence is returned: those of Contains.
// The folds are always representable, so there is no error of overflow.
func (r Range) Folds(sys *system.System, x pitch.Pitch) (iter.Seq[pitch.Pitch], error) {
	const op = "folds"

	if err := r.checkWith(op, sys, x); err != nil {
		return nil, err
	}

	first, last, ok, err := r.foldBounds(op, sys, x)
	if err != nil {
		return nil, err
	}

	return foldSeq(first, last, ok), nil
}

// MustFolds is like Folds but panics with its error on failure.
func (r Range) MustFolds(sys *system.System, x pitch.Pitch) iter.Seq[pitch.Pitch] {
	seq, err := r.Folds(sys, x)
	if err != nil {
		panic(err)
	}

	return seq
}

// Fold returns the nearest fold of x into r in sys, the one shifted by the
// fewest octaves, and true: x itself when it is in r, the lowest fold when
// x is below r and the highest when x is above. When there is no fold it
// returns pitch.Pitch{} and false without an error. Errors are those of
// Contains.
func (r Range) Fold(sys *system.System, x pitch.Pitch) (pitch.Pitch, bool, error) {
	const op = "fold"

	if err := r.checkWith(op, sys, x); err != nil {
		return pitch.Pitch{}, false, err
	}

	loc, err := r.position(op, sys, x)
	if err != nil {
		return pitch.Pitch{}, false, err
	}

	if loc == Inside {
		return x, true, nil
	}

	first, last, ok, err := r.foldBounds(op, sys, x)
	if err != nil || !ok {
		return pitch.Pitch{}, false, err
	}

	if loc == Below {
		return first, true, nil
	}

	return last, true, nil
}

// MustFold is like Fold but panics with its error on failure.
func (r Range) MustFold(sys *system.System, x pitch.Pitch) (pitch.Pitch, bool) {
	p, ok, err := r.Fold(sys, x)
	if err != nil {
		panic(err)
	}

	return p, ok
}

// Intersect returns the range of the sounds common to r and other in sys and
// true. Its lower bound is the higher of the two lower bounds and its upper
// bound is the lower of the two upper bounds; of two bounds at one position,
// the bound of r is taken. When there are no common sounds it returns the
// unset Range and false without an error.
//
// Errors: those of Check for r, then for other.
func (r Range) Intersect(sys *system.System, other Range) (Range, bool, error) {
	const op = "intersect"

	if err := r.checkPair(op, sys, other); err != nil {
		return Range{}, false, err
	}

	lower, err := pick(op, sys, r.lower, other.lower, func(c int) bool { return c >= 0 })
	if err != nil {
		return Range{}, false, err
	}

	upper, err := pick(op, sys, r.upper, other.upper, func(c int) bool { return c <= 0 })
	if err != nil {
		return Range{}, false, err
	}

	c, err := sys.Compare(lower, upper)
	if err != nil {
		return Range{}, false, wrapError(op, err)
	}

	if c > 0 {
		return Range{}, false, nil
	}

	return Range{lower: lower, upper: upper}, true, nil
}

// MustIntersect is like Intersect but panics with its error on failure.
func (r Range) MustIntersect(sys *system.System, other Range) (Range, bool) {
	res, ok, err := r.Intersect(sys, other)
	if err != nil {
		panic(err)
	}

	return res, ok
}

// Includes reports whether every sound of other is in r in sys: the lower
// bound of r is not above the lower bound of other and the upper bound of
// other is not above the upper bound of r. Errors are those of Intersect.
func (r Range) Includes(sys *system.System, other Range) (bool, error) {
	const op = "includes"

	if err := r.checkPair(op, sys, other); err != nil {
		return false, err
	}

	lo, err := sys.Compare(r.lower, other.lower)
	if err != nil {
		return false, wrapError(op, err)
	}

	hi, err := sys.Compare(other.upper, r.upper)
	if err != nil {
		return false, wrapError(op, err)
	}

	return lo <= 0 && hi <= 0, nil
}

// MustIncludes is like Includes but panics with its error on failure.
func (r Range) MustIncludes(sys *system.System, other Range) bool {
	ok, err := r.Includes(sys, other)
	if err != nil {
		panic(err)
	}

	return ok
}

// Hull returns the smallest range in sys that includes both r and other:
// its lower bound is the lower of the two lower bounds and its upper bound is
// the higher of the two upper bounds; of two bounds at one position, the bound
// of r is taken. Errors are those of Intersect.
func (r Range) Hull(sys *system.System, other Range) (Range, error) {
	const op = "hull"

	if err := r.checkPair(op, sys, other); err != nil {
		return Range{}, err
	}

	lower, err := pick(op, sys, r.lower, other.lower, func(c int) bool { return c <= 0 })
	if err != nil {
		return Range{}, err
	}

	upper, err := pick(op, sys, r.upper, other.upper, func(c int) bool { return c >= 0 })
	if err != nil {
		return Range{}, err
	}

	return Range{lower: lower, upper: upper}, nil
}

// MustHull is like Hull but panics with its error on failure.
func (r Range) MustHull(sys *system.System, other Range) Range {
	res, err := r.Hull(sys, other)
	if err != nil {
		panic(err)
	}

	return res
}

// Pitches returns the pitches of r in sys, one step apart, from the position
// of the lower bound to the position of the upper bound, with names chosen by
// rule (system.System.Walk). Their number is Count. The sequence is lazy and
// can be stopped at any element.
//
// All errors are reported before the sequence is returned: those of Check,
// then the errors of system.System.Walk with their causes, such as
// system.ErrInvalid for an unset rule, system.ErrNotApplicable,
// system.ErrNoSpelling and system.ErrBadSpelling. An inadmissible tonic of the
// rule matches system.ErrInadmissible but not ErrInadmissible of this
// package: the bounds are already checked.
func (r Range) Pitches(sys *system.System, rule system.SpellingRule) (iter.Seq[pitch.Pitch], error) {
	const op = "pitches"

	if err := r.check(op, sys); err != nil {
		return nil, err
	}

	seq, err := sys.Walk(r.lower, r.upper, rule)
	if err != nil {
		return nil, wrapError(op, err)
	}

	return seq, nil
}

// MustPitches is like Pitches but panics with its error on failure.
func (r Range) MustPitches(sys *system.System, rule system.SpellingRule) iter.Seq[pitch.Pitch] {
	seq, err := r.Pitches(sys, rule)
	if err != nil {
		panic(err)
	}

	return seq
}

// Equivalent reports whether r and other have the same sounds in sys: whether
// their lower bounds are at one position and their upper bounds are at one
// position. Errors are those of Intersect.
func (r Range) Equivalent(sys *system.System, other Range) (bool, error) {
	const op = "equivalent"

	if err := r.checkPair(op, sys, other); err != nil {
		return false, err
	}

	lo, err := sys.Equivalent(r.lower, other.lower)
	if err != nil {
		return false, wrapError(op, err)
	}

	hi, err := sys.Equivalent(r.upper, other.upper)
	if err != nil {
		return false, wrapError(op, err)
	}

	return lo && hi, nil
}

// MustEquivalent is like Equivalent but panics with its error on failure.
func (r Range) MustEquivalent(sys *system.System, other Range) bool {
	ok, err := r.Equivalent(sys, other)
	if err != nil {
		panic(err)
	}

	return ok
}

// Location tells where a pitch is relative to a range. The zero value
// LocationUnset is no answer; it is returned on error.
type Location uint8

// Locations. The numbers of the constants are stable.
const (
	// LocationUnset is the zero value: no answer.
	LocationUnset Location = iota
	// Below: the position of the pitch is below the lower bound.
	Below
	// Inside: the pitch is in the range, the bounds included.
	Inside
	// Above: the position of the pitch is above the upper bound.
	Above
)

// String returns "below", "inside" or "above", "Location(unset)" for
// LocationUnset and "Location(<n>)" for other values. It is a debugging form.
func (l Location) String() string {
	switch l {
	case LocationUnset:
		return "Location(unset)"
	case Below:
		return "below"
	case Inside:
		return "inside"
	case Above:
		return "above"
	default:
		return fmt.Sprintf("Location(%d)", uint8(l))
	}
}

// check checks sys and then r in sys: ErrUnset, ErrNoRange, then validity.
func (r Range) check(op string, sys *system.System) error {
	if !sys.IsValid() {
		return unsetError(op, "system")
	}

	if !r.IsValid() {
		return noRangeError(op)
	}

	return r.validIn(op, sys)
}

// checkWith checks sys, r and then the pitch x.
func (r Range) checkWith(op string, sys *system.System, x pitch.Pitch) error {
	if err := r.check(op, sys); err != nil {
		return err
	}

	return checkPitch(op, "pitch", sys, x)
}

// checkPair checks sys, then r and then other.
func (r Range) checkPair(op string, sys *system.System, other Range) error {
	if err := r.check(op, sys); err != nil {
		return err
	}

	if !other.IsValid() {
		return noRangeError(op)
	}

	return other.validIn(op, sys)
}

// validIn checks a set range in a valid system: ErrInadmissible for the lower
// bound, then for the upper bound, then ErrInverted.
func (r Range) validIn(op string, sys *system.System) error {
	if err := admit(op, "lower bound", sys, r.lower); err != nil {
		return err
	}

	if err := admit(op, "upper bound", sys, r.upper); err != nil {
		return err
	}

	c, err := sys.Compare(r.lower, r.upper)
	if err != nil {
		return wrapError(op, err)
	}

	if c > 0 {
		return invertedError(op, r, sys)
	}

	return nil
}

// locate checks its inputs and returns the location of x.
func (r Range) locate(op string, sys *system.System, x pitch.Pitch) (Location, error) {
	if err := r.checkWith(op, sys, x); err != nil {
		return LocationUnset, err
	}

	return r.position(op, sys, x)
}

// position returns the location of x relative to a range valid in sys.
func (r Range) position(op string, sys *system.System, x pitch.Pitch) (Location, error) {
	c, err := sys.Compare(x, r.lower)
	if err != nil {
		return LocationUnset, wrapError(op, err)
	}

	if c < 0 {
		return Below, nil
	}

	c, err = sys.Compare(x, r.upper)
	if err != nil {
		return LocationUnset, wrapError(op, err)
	}

	if c > 0 {
		return Above, nil
	}

	return Inside, nil
}

// width checks its inputs and returns the width.
func (r Range) width(op string, sys *system.System) (system.Step, error) {
	if err := r.check(op, sys); err != nil {
		return 0, err
	}

	w, err := sys.Distance(r.lower, r.upper)
	if err != nil {
		return 0, wrapError(op, err)
	}

	return w, nil
}

// foldBounds returns the lowest and the highest fold of x into a range valid
// in sys and whether there is any. It works with Octaves of the bounds, so no
// position is computed and no fold can overflow: every fold has an Octave
// between the Octaves of the bounds.
func (r Range) foldBounds(op string, sys *system.System, x pitch.Pitch) (first, last pitch.Pitch, ok bool, err error) {
	first, err = x.WithOctave(r.lower.Octave())
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, false, wrapError(op, err)
	}

	c, err := sys.Compare(first, r.lower)
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, false, wrapError(op, err)
	}

	if c < 0 {
		if first, err = first.Up(); err != nil {
			return pitch.Pitch{}, pitch.Pitch{}, false, nil
		}
	}

	last, err = x.WithOctave(r.upper.Octave())
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, false, wrapError(op, err)
	}

	c, err = sys.Compare(last, r.upper)
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, false, wrapError(op, err)
	}

	if c > 0 {
		if last, err = last.Down(); err != nil {
			return pitch.Pitch{}, pitch.Pitch{}, false, nil
		}
	}

	c, err = sys.Compare(first, last)
	if err != nil {
		return pitch.Pitch{}, pitch.Pitch{}, false, wrapError(op, err)
	}

	if c > 0 {
		return pitch.Pitch{}, pitch.Pitch{}, false, nil
	}

	return first, last, true, nil
}

// foldSeq returns the folds from first to last, one octave apart; an empty
// sequence when ok is false.
func foldSeq(first, last pitch.Pitch, ok bool) iter.Seq[pitch.Pitch] {
	return func(yield func(pitch.Pitch) bool) {
		if !ok {
			return
		}

		for p := first; ; {
			if !yield(p) || p == last {
				return
			}

			next, err := p.Up()
			if err != nil {
				return
			}

			p = next
		}
	}
}

// pick returns a if keep(Compare(a, b)) holds and b otherwise.
func pick(op string, sys *system.System, a, b pitch.Pitch, keep func(int) bool) (pitch.Pitch, error) {
	c, err := sys.Compare(a, b)
	if err != nil {
		return pitch.Pitch{}, wrapError(op, err)
	}

	if keep(c) {
		return a, nil
	}

	return b, nil
}

// checkPitch checks a pitch argument in a valid system: ErrUnsetPitch, then
// ErrInadmissible.
func checkPitch(op, what string, sys *system.System, x pitch.Pitch) error {
	if !x.IsValid() {
		return unsetPitchError(op, what)
	}

	return admit(op, what, sys, x)
}

// admit reports ErrInadmissible when the name of a set pitch is not
// admissible in a valid system.
func admit(op, what string, sys *system.System, p pitch.Pitch) error {
	ok, err := sys.AdmitsPitch(p)
	if err != nil {
		return wrapError(op, err)
	}

	if !ok {
		return inadmissibleError(op, what, p, sys)
	}

	return nil
}
