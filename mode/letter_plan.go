package mode

import (
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// octaveSize is the size of the perfect octave, the cycle of every mode built
// by a spelling rule.
const octaveSize = 8

// SuggestLetterPlan returns the letter plan that rule gives for a closed mode
// spanning one octave of sys, with the adjacent numbers adjacent and the name
// start of degree 1. The plan is a suggestion: the caller can read or edit it
// and pass it to FromNumbersFromWithPlan, or build the same mode directly with
// FromNumbersFromWithRule.
//
// adjacent holds k >= 1 numbers of positions, the last one closing the cycle;
// every number must be at least 1 and their exact sum must equal
// sys.Classes(). The place of degree j is the sum of the first j-1 numbers.
// Degree 1 keeps the name start; the rule is not applied to it, so D♭ stays D♭
// under Sharps. Degree j >= 2 gets the name that sys.Spell chooses with rule
// for the class of start transposed by its place. The size of degree j is the
// size of the written ascending interval from start to that name at its real
// place: F♯ followed by F eleven positions higher gives size 8, not 1. The
// cycle size is 8.
//
// With one number (k = 1, the number equals sys.Classes()) the plan is
// {Degrees: [1], Cycle: 8}; the rule is neither applied nor checked for
// applicability, but start must still be admissible.
//
// The result is never an incorrect plan: on success 1 = Degrees[0] <=
// Degrees[1] <= ... <= Degrees[k-1] <= 8. Degrees is a new slice owned by the
// caller; changing it changes no mode and no other plan.
//
// Errors, in this order, each category checked over all relevant arguments
// before the next: ErrUnset for a nil or unset sys, an unset start, or an
// unset rule; ErrEmptyInput for no numbers; ErrInvalidArgument for more than
// MaxDegrees numbers, checked before any number is read; ErrInvalidStructure
// for a number less than 1; ErrSystemOctaveRequired when the exact sum is not
// sys.Classes() (the sum never overflows: MaxInt MaxInt 14 is rejected in the
// traditional system); system.ErrInadmissible when start is not admissible in
// sys; then, for degrees 2..k in order, the first error of applying rule:
// system.ErrNotApplicable, system.ErrInadmissible for an inadmissible tonic of
// the rule, system.ErrNoSpelling, or system.ErrBadSpelling; finally
// ErrInvalidStructure when the chosen names give no correct letter plan, that
// is, a name is written below degree 1 or below the previous degree, or above
// degree 1 of the next cycle. The system errors are returned without any mode
// sentinel in the chain; the last ErrInvalidStructure has no system cause.
// When the rule gives no correct plan, choose an explicit plan and use
// FromNumbersFromWithPlan. On error the result is the zero LetterPlan.
//
// A custom rule is called synchronously in the caller's goroutine, only with
// the classes of degrees 2..j in degree order, where j is k or the first degree
// whose name fails; each of these classes is passed at least once, and the
// rule is never called for k = 1 or when an earlier check fails. How many
// times one class is passed is up to package system. A panic of the rule is
// not recovered. The rule must return one name for one class.
//
// adjacent and rule are neither changed nor retained. The function has no
// state and is safe for concurrent use whenever the rule is. The time is O(k)
// plus the cost of k-1 applications of the rule.
func SuggestLetterPlan(sys *system.System, adjacent []system.Step, start spelling.Spelling, rule system.SpellingRule) (LetterPlan, error) {
	return suggestLetterPlan("SuggestLetterPlan", sys, adjacent, start, rule)
}

// FromNumbersFromWithRule creates the closed one-octave mode of sys with the
// adjacent numbers adjacent, the name start of degree 1, and the names that
// rule chooses for the other degrees. It equals FromNumbersFromWithPlan with
// the same numbers and start and the plan of SuggestLetterPlan with the same
// arguments; it fails exactly when SuggestLetterPlan fails, with the same
// errors in the same order. When the rule gives no correct plan
// (ErrInvalidStructure), choose an explicit plan and use
// FromNumbersFromWithPlan.
//
// The mode has the cycle P8 and no catalog origin; neither the rule nor start
// is stored in it. Its LetterPlan is the suggested plan, its NumbersFrom(sys,
// start).Adjacent are the passed numbers, and the names of its notes from a
// note named start are start followed by the names of the rule.
//
// Unlike FromNumbers and FromNumbersFrom, which accept only seven numbers and
// the default plan "one letter per degree", this function accepts any number
// of degrees, and with seven degrees the rule can give another plan.
//
// A custom rule is called with the same sequence of classes as by
// SuggestLetterPlan; building the mode from the plan does not apply the rule
// again. adjacent and rule are neither changed nor retained.
func FromNumbersFromWithRule(sys *system.System, adjacent []system.Step, start spelling.Spelling, rule system.SpellingRule) (Mode, error) {
	plan, err := suggestLetterPlan("FromNumbersFromWithRule", sys, adjacent, start, rule)
	if err != nil {
		return Mode{}, err
	}
	return FromNumbersFromWithPlan(sys, adjacent, start, plan)
}

// suggestLetterPlan performs the checks of SuggestLetterPlan in their
// documented order and returns the suggested plan.
func suggestLetterPlan(op string, sys *system.System, adjacent []system.Step, start spelling.Spelling, rule system.SpellingRule) (LetterPlan, error) {
	if !sys.IsValid() || !start.IsValid() || !rule.IsValid() {
		return LetterPlan{}, ErrUnset
	}
	k := len(adjacent)
	if k == 0 {
		return LetterPlan{}, ErrEmptyInput
	}
	if k > MaxDegrees {
		return LetterPlan{}, ErrInvalidArgument
	}
	for _, n := range adjacent {
		if n < 1 {
			return LetterPlan{}, ErrInvalidStructure
		}
	}
	classes := sys.Classes()
	// Every number is positive, so the running sum stays in 0..classes and
	// the comparison below never overflows.
	total := 0
	for _, n := range adjacent {
		if int(n) > classes-total {
			return LetterPlan{}, ErrSystemOctaveRequired
		}
		total += int(n)
	}
	if total != classes {
		return LetterPlan{}, ErrSystemOctaveRequired
	}
	first, err := sys.Class(start)
	if err != nil {
		return LetterPlan{}, fmt.Errorf("%s: start %v: %w", op, start, err)
	}

	// Choose all names first: a rule error of any degree is reported before
	// an incorrect letter plan of an earlier degree.
	names := make([]spelling.Spelling, k)
	places := make([]int, k)
	names[0] = start
	for j := 1; j < k; j++ {
		places[j] = places[j-1] + int(adjacent[j-1])
		name, err := sys.Spell(first.Transpose(places[j]), rule)
		if err != nil {
			return LetterPlan{}, fmt.Errorf("%s: degree %d: %w", op, j+1, err)
		}
		names[j] = name
	}

	degrees := make([]int, k)
	degrees[0] = 1
	for j := 1; j < k; j++ {
		size, ok := writtenSize(sys, start, names[j], places[j])
		if !ok || size < degrees[j-1] || size > octaveSize {
			return LetterPlan{}, fmt.Errorf("%s: degree %d (%v) written out of order: %w", op, j+1, names[j], ErrInvalidStructure)
		}
		degrees[j] = size
	}
	return LetterPlan{Degrees: degrees, Cycle: octaveSize}, nil
}

// writtenSize returns the size of the ascending interval from start to name
// when name lies place positions above start: n = d + 1 + 7t, where d is the
// letter distance up from the letter of start to the letter of name, and t is
// the number of octaves that the written interval adds,
// t = (place - S - (α_name - α_start)·c) / N, with S the positions of the
// natural letters over d letter steps. The division is exact because name has
// the class of start transposed by place. The result is reported only for
// t in {0, 1}, the only values that can give a size from 1 to 8.
func writtenSize(sys *system.System, start, name spelling.Spelling, place int) (int, bool) {
	from := start.Letter()
	d := from.DistanceUp(name.Letter())
	numerator := new(big.Rat).SetInt64(int64(place))
	for i := range d {
		step, _ := sys.StepAfter(from.Shift(i))
		numerator.Sub(numerator, new(big.Rat).SetInt64(int64(step)))
	}
	shift := new(big.Rat).Sub(rational(name.Alteration()), rational(start.Alteration()))
	shift.Mul(shift, new(big.Rat).SetInt64(int64(sys.Chromatic())))
	numerator.Sub(numerator, shift)
	octaves := numerator.Quo(numerator, new(big.Rat).SetInt64(int64(sys.Classes())))
	if !octaves.IsInt() {
		return 0, false
	}
	t := octaves.Num()
	if t.Sign() < 0 || t.Cmp(big.NewInt(1)) > 0 {
		return 0, false
	}
	return d + 1 + lettersPerOctave*int(t.Int64()), true
}

// OpenFromNumbers creates an open mode of k = len(adjacent)+1 degrees in a
// regular sys from the adjacent numbers adjacent and the sizes of the degrees
// sizes. The interval of degree j is the ascending interval of size sizes[j-1]
// whose positional size is the exact sum of the first j-1 numbers
// (interval.FromSteps); its quality can be fractional. No name is built or
// checked.
//
// sizes lists the k sizes of the degrees, without a cycle size; it is correct
// when it is not empty, sizes[0] is 1, and the sizes do not decrease. Sizes
// above 8 are allowed. Equal sizes of neighboring degrees give a unison with a
// positive deviation as their adjacent interval (A1 in the traditional system
// for the number 1).
//
// The mode has no cycle and no catalog origin. Its Numbers(sys).Adjacent are
// the passed numbers, its LetterPlan is (LetterPlan{Degrees: sizes}, false),
// and it equals OpenFromDegreeIntervals of its degree intervals.
//
// Errors, in this order: ErrUnset for a nil or unset sys; ErrEmptyInput for no
// numbers, whatever sizes is; ErrInvalidArgument for len(adjacent) >=
// MaxDegrees, checked before any number is read and whatever sizes is, and for
// a non-empty sizes whose first size is not 1 or whose sizes decrease;
// ErrInvalidStructure for a number less than 1; ErrWrongDegreeCount when
// len(sizes) != len(adjacent)+1, including an empty or nil sizes;
// interval.ErrStartRequired when sys is not regular; ErrOutOfRange together
// with interval.ErrOutOfRange when the position of a degree does not fit in an
// int, the interval of a degree cannot be represented, or an adjacent interval
// between two representable degree intervals cannot be represented. On error
// the result is the zero Mode.
//
// adjacent and sizes are neither changed nor retained. The function has no
// state and is safe for concurrent use. The time is O(k + len(sizes)).
func OpenFromNumbers(sys *system.System, adjacent []system.Step, sizes []int) (Mode, error) {
	return openFromNumbers("OpenFromNumbers", sys, adjacent, spelling.Spelling{}, sizes, false)
}

// OpenFromNumbersFrom is OpenFromNumbers with start as the name of degree 1.
// It works in any system, including an irregular one; start and the names of
// all k degrees must be admissible in sys.
//
// Errors, in this order: ErrUnset for a nil or unset sys or an unset start;
// then ErrEmptyInput, ErrInvalidArgument, ErrInvalidStructure, and
// ErrWrongDegreeCount as for OpenFromNumbers; system.ErrInadmissible when
// start is not admissible, then when the built name of a degree 2..k, in
// degree order, is not admissible; ErrOutOfRange together with
// interval.ErrOutOfRange as for OpenFromNumbers and also when the name of a
// degree cannot be represented. Every degree is checked for admissibility
// before any range error is reported, so an inadmissible name of a late degree
// is reported before an unrepresentable interval of an early one; a degree
// that fails the range check takes no part in the admissibility check. The
// system errors are returned without any mode sentinel in the chain. On error
// the result is the zero Mode.
func OpenFromNumbersFrom(sys *system.System, adjacent []system.Step, start spelling.Spelling, sizes []int) (Mode, error) {
	return openFromNumbers("OpenFromNumbersFrom", sys, adjacent, start, sizes, true)
}

func validSizes(sizes []int) bool {
	return len(sizes) == 0 || (sizes[0] == 1 && slices.IsSorted(sizes))
}

func openFromNumbers(op string, sys *system.System, adjacent []system.Step, start spelling.Spelling, sizes []int, named bool) (Mode, error) {
	if !sys.IsValid() || (named && !start.IsValid()) {
		return Mode{}, ErrUnset
	}
	if len(adjacent) == 0 {
		return Mode{}, ErrEmptyInput
	}
	if len(adjacent) >= MaxDegrees || !validSizes(sizes) {
		return Mode{}, ErrInvalidArgument
	}
	for _, n := range adjacent {
		if n < 1 {
			return Mode{}, ErrInvalidStructure
		}
	}
	if len(sizes) != len(adjacent)+1 {
		return Mode{}, ErrWrongDegreeCount
	}
	if named {
		if !admits(sys, start) {
			return Mode{}, fmt.Errorf("%s: start name %v: %w", op, start, system.ErrInadmissible)
		}
	} else if reg, err := sys.Regularity(); err != nil || !reg.Regular {
		return Mode{}, fmt.Errorf("%s: %w", op, interval.ErrStartRequired)
	}

	// The first admissibility failure of any degree wins over the first
	// range failure of any degree, so both are collected in one pass.
	degrees := make([]interval.Interval, len(sizes))
	var nameErr, rangeErr error
	position := new(big.Int)
	for i, size := range sizes {
		if i > 0 {
			position.Add(position, big.NewInt(int64(adjacent[i-1])))
		}
		if !position.IsInt64() {
			// Positions only grow, so no later degree can be built either.
			if rangeErr == nil {
				rangeErr = fmt.Errorf("%s: degree %d: %w: %w", op, i+1, ErrOutOfRange, interval.ErrOutOfRange)
			}
			break
		}
		degreeInterval, err := interval.FromSteps(sys, size, system.Step(position.Int64()), start)
		switch {
		case err == nil:
			degrees[i] = degreeInterval
		case errors.Is(err, interval.ErrOutOfRange):
			if rangeErr == nil {
				rangeErr = fmt.Errorf("%s: degree %d: %w: %w", op, i+1, ErrOutOfRange, err)
			}
		default:
			if nameErr == nil {
				nameErr = fmt.Errorf("%s: degree %d: %w", op, i+1, err)
			}
		}
	}
	if nameErr != nil {
		return Mode{}, nameErr
	}
	if rangeErr != nil {
		return Mode{}, rangeErr
	}
	m, err := OpenFromDegreeIntervals(degrees)
	if err != nil {
		if errors.Is(err, ErrOutOfRange) && !errors.Is(err, interval.ErrOutOfRange) {
			return Mode{}, fmt.Errorf("%s: %w: %w", op, interval.ErrOutOfRange, err)
		}
		return Mode{}, fmt.Errorf("%s: %w", op, err)
	}
	return m, nil
}

// LetterPlan returns the letter plan of m: the sizes of the k degree intervals
// and, for a closed mode, the size of the cycle. The second result reports
// whether m has a cycle size and agrees with the second result of Cycle.
//
// For a closed mode the result is (LetterPlan{Degrees: sizes, Cycle: cycle
// size}, true), with Cycle >= 1. For an open mode it is (LetterPlan{Degrees:
// sizes, Cycle: 0}, false). For the unset mode it is (LetterPlan{}, false)
// with nil Degrees. The plan of a set mode is correct: Degrees[0] is 1, the
// sizes do not decrease, and the cycle size is not less than the last size.
// Origin and the Changed mark do not affect the plan.
//
// The plan together with the numbers of m rebuilds m: when NumbersFrom(sys,
// start) (or Numbers(sys) in a regular system) succeeds and every adjacent
// number is at least 1, a closed mode equals FromNumbersFromWithPlan (or
// FromNumbersWithPlan) of those numbers and this plan, and an open mode equals
// OpenFromNumbersFrom (or OpenFromNumbers) of those numbers and plan.Degrees.
//
// LetterPlan never fails. Every call returns a new Degrees slice owned by the
// caller; changing it changes neither m nor any other plan. The time is O(k).
func (m Mode) LetterPlan() (LetterPlan, bool) {
	if !m.IsValid() {
		return LetterPlan{}, false
	}
	degrees := make([]int, len(m.prefixes))
	for i, prefix := range m.prefixes {
		degrees[i] = prefix.Size()
	}
	if !m.IsClosed() {
		return LetterPlan{Degrees: degrees}, false
	}
	return LetterPlan{Degrees: degrees, Cycle: m.cycle.Size()}, true
}
