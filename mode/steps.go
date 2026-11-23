package mode

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// StepsBetween returns by how many positions of sys the note of the place to
// is above the note of the place from. The note of a place (j, c) is the note
// of degree 1 moved by DegreeInterval(j) plus c cycles; the sign follows the
// positions, not the degree numbers, so it can be zero or negative even for an
// upward pair of numbers. The result does not depend on the octave of the
// first note, which is therefore not passed.
//
// This form has no name for degree 1 and is allowed only in a regular system,
// where the result does not depend on that name. Names are neither built nor
// checked. For cycle-zero places the result equals the difference of the
// degree numbers of Numbers in sys whenever Numbers succeeds and the
// difference fits in an int. StepsBetween can succeed where Numbers fails:
// only the final distance must be a whole number that fits in a Step; the
// numbers of the individual places, the interval of a place from degree 1,
// and the product of the cycle count and the cycle need not.
//
// Places are checked from, then to. A place has Number in 1..Len() and, for an
// open mode, Cycle 0; any cycle is allowed for a closed mode. Number 0 is an
// invalid place, not "no context".
//
// Errors, in this order: ErrUnset for an unset mode or a nil or unset sys;
// ErrInvalidArgument for an invalid place; interval.ErrStartRequired when sys
// is not regular, for any places, including from == to;
// interval.ErrFractionalSteps when the exact distance is not whole, as for
// the open mode P1 n3 n7 from degree 1 to degree 2 in the traditional system;
// ErrOutOfRange together with interval.ErrOutOfRange when the distance does
// not fit in a Step. On error the result is 0. The mode is not changed and
// nothing is retained. The time is O(1) operations of exact arithmetic,
// independent of Len() and of the cycle counts.
func (m Mode) StepsBetween(sys *system.System, from, to DegreePlace) (system.Step, error) {
	return m.stepsBetween("StepsBetween", sys, spelling.Spelling{}, from, to, false)
}

// StepsBetweenFrom is StepsBetween with start as the name of degree 1. The
// names of the notes of both places are computed from start exactly, as
// NoteFrom computes its final name; no intermediate interval is built. The
// form works in any system, including an irregular one, where the number of
// positions of a cycle that is not an octave can depend on the place.
//
// start and the names of the notes of both places must be admissible in sys;
// the names of the other degrees and of the next cycle are not checked. In a
// regular system both forms return the same value whenever this one succeeds.
// For cycle-zero places the result equals the difference of the degree
// numbers of NumbersFrom(sys, start) whenever that succeeds and the difference
// fits in an int.
//
// Errors, in this order: ErrUnset for an unset mode, a nil or unset sys, or
// an unset start; ErrInvalidArgument for an invalid place, from before to;
// system.ErrInadmissible when start is not admissible, then when the built
// name of the note of from, then of to, is not admissible; ErrOutOfRange
// together with interval.ErrOutOfRange and alteration.ErrOutOfRange when the
// name of the note of from, then of to, is not representable, even if the
// distance itself is; interval.ErrFractionalSteps when the distance is not
// whole, which does not happen in a correct system; ErrOutOfRange together
// with interval.ErrOutOfRange when the distance does not fit in a Step. An
// admissibility failure of either built name is reported before an
// unrepresentable name of the other place. On error the result is 0.
func (m Mode) StepsBetweenFrom(sys *system.System, start spelling.Spelling, from, to DegreePlace) (system.Step, error) {
	return m.stepsBetween("StepsBetweenFrom", sys, start, from, to, true)
}

func (m Mode) stepsBetween(op string, sys *system.System, start spelling.Spelling, from, to DegreePlace, named bool) (system.Step, error) {
	if !m.IsValid() || !sys.IsValid() || (named && !start.IsValid()) {
		return 0, ErrUnset
	}
	if err := m.checkPlace(op, "from", from); err != nil {
		return 0, err
	}
	if err := m.checkPlace(op, "to", to); err != nil {
		return 0, err
	}
	a, b := m.placeCoordinates(from), m.placeCoordinates(to)
	var distance *big.Rat
	if named {
		d, err := namedDistance(op, sys, start, a, b)
		if err != nil {
			return 0, err
		}
		distance = d
	} else {
		reg, err := sys.Regularity()
		if err != nil || !reg.Regular {
			return 0, fmt.Errorf("%s: %w", op, interval.ErrStartRequired)
		}
		distance = regularDistance(reg, b.sub(a))
	}
	if !distance.IsInt() {
		return 0, fmt.Errorf("%s: %w", op, interval.ErrFractionalSteps)
	}
	if !distance.Num().IsInt64() {
		return 0, fmt.Errorf("%s: %w: %w", op, ErrOutOfRange, interval.ErrOutOfRange)
	}
	return system.Step(distance.Num().Int64()), nil
}

// checkPlace validates a degree place without arithmetic on its fields.
func (m Mode) checkPlace(op, role string, p DegreePlace) error {
	if p.Number < 1 || p.Number > m.Len() {
		return fmt.Errorf("%s: %s place number %d outside 1..%d: %w", op, role, p.Number, m.Len(), ErrInvalidArgument)
	}
	if m.IsOpen() && p.Cycle != 0 {
		return fmt.Errorf("%s: %s place cycle %d of an open mode: %w", op, role, p.Cycle, ErrInvalidArgument)
	}
	return nil
}

// regularDistance returns the positional size of exact coordinates in a
// regular system. With (q, r) the floor division of the letter steps D by 7,
// the size is q·N + S(C, r) + (V − 12q − n(r))·c, the formula of
// interval.Interval.Steps without a start name. For the step pattern
// L L S L L L S with c = L − S it reduces to D·(2S − L) + V·c, which is
// linear in (D, V) and therefore valid for a difference of any sign.
func regularDistance(reg system.Regularity, d coordinates) *big.Rat {
	perLetter := big.NewInt(int64(2*reg.Small - reg.Large))
	chromatic := new(big.Rat).SetInt64(int64(reg.Large - reg.Small))
	result := new(big.Rat).SetInt(perLetter.Mul(perLetter, d.letters))
	return result.Add(result, chromatic.Mul(chromatic, d.chromatic))
}

// namedDistance builds the names of both places from start, checks them in
// the documented order, and subtracts their exact positions.
func namedDistance(op string, sys *system.System, start spelling.Spelling, a, b coordinates) (*big.Rat, error) {
	if !admits(sys, start) {
		return nil, fmt.Errorf("%s: start name %v: %w", op, start, system.ErrInadmissible)
	}
	nameFrom, errFrom := a.transposeName(start)
	nameTo, errTo := b.transposeName(start)
	if errFrom == nil && !admits(sys, nameFrom) {
		return nil, fmt.Errorf("%s: name %v of the from place: %w", op, nameFrom, system.ErrInadmissible)
	}
	if errTo == nil && !admits(sys, nameTo) {
		return nil, fmt.Errorf("%s: name %v of the to place: %w", op, nameTo, system.ErrInadmissible)
	}
	if errFrom != nil {
		return nil, nameRangeError(op, "from", errFrom)
	}
	if errTo != nil {
		return nil, nameRangeError(op, "to", errTo)
	}
	upper := namedPosition(sys, start.Letter(), b.letters, nameTo)
	lower := namedPosition(sys, start.Letter(), a.letters, nameFrom)
	return upper.Sub(upper, lower), nil
}

func admits(sys *system.System, name spelling.Spelling) bool {
	ok, err := sys.Admits(name)
	return err == nil && ok
}

// nameRangeError reports an unrepresentable place name. The cause already
// carries interval.ErrOutOfRange and the alteration cause.
func nameRangeError(op, role string, err error) error {
	if errors.Is(err, interval.ErrOutOfRange) {
		return fmt.Errorf("%s: name of the %s place: %w: %w", op, role, ErrOutOfRange, err)
	}
	return fmt.Errorf("%s: name of the %s place: %w: %w: %w", op, role, ErrOutOfRange, interval.ErrOutOfRange, err)
}

// namedPosition returns the exact position of name, reached by the given
// letter steps from the letter l, relative to the natural l:
// q·N + S(l, r) + α·c, where (q, r) is the floor division of the letter steps
// by 7 and α is the alteration value of name.
func namedPosition(sys *system.System, l letter.Letter, letters *big.Int, name spelling.Spelling) *big.Rat {
	octaves, rem := splitLetters(letters)
	position := octaves.Mul(octaves, big.NewInt(int64(sys.Classes())))
	for j := range rem {
		step, _ := sys.StepAfter(l.Shift(j))
		position.Add(position, big.NewInt(int64(step)))
	}
	shift := rational(name.Alteration())
	shift.Mul(shift, new(big.Rat).SetInt64(int64(sys.Chromatic())))
	return shift.Add(shift, new(big.Rat).SetInt(position))
}
