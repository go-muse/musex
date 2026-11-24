package mode

import (
	"cmp"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// LetterPlan supplies the interval sizes of the degrees and the cycle.
//
// Degrees[i] is the size of the interval of degree i+1 from degree 1: the
// number of letters it spans, counting both ends. Equal neighboring sizes
// write two degrees with one letter; a skipped size skips a letter. Cycle is
// the size of the cycle. A plan is correct when Degrees[0] is 1, the sizes do
// not decrease, and Cycle is not less than the last size; the function that
// accepts a plan checks it. A plan without a cycle, as returned for an open
// mode by Mode.LetterPlan, has Cycle 0; the second result of Mode.LetterPlan
// tells whether there is a cycle.
//
// FromNumbersWithPlan and FromNumbersFromWithPlan build a closed mode from a
// plan; SuggestLetterPlan suggests a plan from a spelling rule, and
// Mode.LetterPlan reads the plan of a mode. A plan is ordinary data without
// an unset state: a plan returned by this package is owned by the caller and
// shares no data with any mode.
type LetterPlan struct {
	Degrees []int
	Cycle   int
}

// Positions describes the sizes of a mode in a given musical system. It is
// ordinary data that callers can read and change.
//
// The offset methods DegreesAt, OccupiedSteps, OccupiedStepsBackward, StepsIn,
// and StepsInBackward read only Degrees: the element at index i is the number
// of degree i+1, its offset in positions of the system from degree 1. Adjacent,
// Cycle, and HasCycle are neither used nor checked, and the value need not
// match any mode; unsorted, repeated, and negative numbers are allowed. The
// cycle boundary is never crossed: degrees of other cycles are not searched.
// None of these methods fails. Every degree list they return, including each
// list yielded by an iterator, is a new slice owned by the caller; an empty
// list is a non-nil slice of length zero. The methods have value receivers, so
// they can be called on a temporary result such as m.MustNumbers(sys).
type Positions struct {
	Degrees  []system.Step
	Adjacent []system.Step
	Cycle    system.Step
	HasCycle bool
}

// CycleInfo compares the cycle with the system octave and gives the smallest
// alignment of whole cycles and octaves.
type CycleInfo struct {
	Comparison int
	Cycles     int
	Octaves    int
}

// Symmetry describes rotations and transpositions of the numerical structure.
type Symmetry struct {
	Rotations      int
	Transpositions int
	Limited        bool
}

// FromNumbers creates a seven-degree mode spanning one octave of a regular
// system, using the default letter plan.
//
// The default plan gives every degree the next letter, so FromNumbers accepts
// exactly seven numbers whose sum is the number of positions of the system;
// other counts give ErrWrongDegreeCount. For any number of degrees, use
// FromNumbersFromWithRule to let a spelling rule choose the letters,
// FromNumbersWithPlan to pass an explicit letter plan, or OpenFromNumbers for
// an open mode.
func FromNumbers(sys *system.System, adjacent []system.Step) (Mode, error) {
	return fromNumbers(sys, adjacent, spelling.Spelling{}, nil, false)
}

// FromNumbersFrom creates a seven-degree, one-octave mode from start.
//
// Like FromNumbers, it uses the default letter plan and accepts exactly seven
// numbers. For any number of degrees, use FromNumbersFromWithRule (the letters
// chosen by a spelling rule), FromNumbersFromWithPlan (an explicit letter
// plan), or OpenFromNumbersFrom (an open mode).
func FromNumbersFrom(sys *system.System, adjacent []system.Step, start spelling.Spelling) (Mode, error) {
	return fromNumbers(sys, adjacent, start, nil, true)
}

// FromNumbersWithPlan creates a closed mode of the specified sizes in a regular system.
func FromNumbersWithPlan(sys *system.System, adjacent []system.Step, plan LetterPlan) (Mode, error) {
	return fromNumbers(sys, adjacent, spelling.Spelling{}, &plan, false)
}

// FromNumbersFromWithPlan creates a closed mode of the specified sizes from start.
func FromNumbersFromWithPlan(sys *system.System, adjacent []system.Step, start spelling.Spelling, plan LetterPlan) (Mode, error) {
	return fromNumbers(sys, adjacent, start, &plan, true)
}

func validPlan(plan LetterPlan) bool {
	if len(plan.Degrees) == 0 {
		return plan.Cycle > 0
	}
	return plan.Degrees[0] == 1 && slices.IsSorted(plan.Degrees) &&
		plan.Cycle >= plan.Degrees[len(plan.Degrees)-1]
}

func fromNumbers(sys *system.System, adjacent []system.Step, start spelling.Spelling, plan *LetterPlan, requireStart bool) (Mode, error) {
	if !sys.IsValid() || (requireStart && !start.IsValid()) {
		return Mode{}, ErrUnset
	}
	k := len(adjacent)
	if k == 0 {
		return Mode{}, ErrEmptyInput
	}
	if k > MaxDegrees || (plan != nil && !validPlan(*plan)) {
		return Mode{}, ErrInvalidArgument
	}
	total := new(big.Int)
	for _, n := range adjacent {
		if n < 1 {
			return Mode{}, ErrInvalidStructure
		}
		total.Add(total, big.NewInt(int64(n)))
	}
	if plan == nil {
		const octaveDegrees = 7
		if k != octaveDegrees {
			return Mode{}, ErrWrongDegreeCount
		}
		if total.Cmp(big.NewInt(int64(sys.Classes()))) != 0 {
			return Mode{}, ErrSystemOctaveRequired
		}
		plan = &LetterPlan{Degrees: []int{1, 2, 3, 4, 5, 6, 7}, Cycle: 8}
	} else if len(plan.Degrees) != k {
		return Mode{}, ErrWrongDegreeCount
	}

	degrees := make([]interval.Interval, k)
	position := new(big.Int)
	for i, size := range plan.Degrees {
		if !position.IsInt64() {
			return Mode{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange)
		}
		degreeInterval, err := interval.FromSteps(sys, size, system.Step(position.Int64()), start)
		if err != nil {
			return Mode{}, positionError(fmt.Sprintf("degree %d", i+1), err)
		}
		degrees[i] = degreeInterval
		position.Add(position, big.NewInt(int64(adjacent[i])))
	}
	if !position.IsInt64() {
		return Mode{}, errors.Join(ErrOutOfRange, interval.ErrOutOfRange)
	}
	cycle, err := interval.FromSteps(sys, plan.Cycle, system.Step(position.Int64()), start)
	if err != nil {
		return Mode{}, positionError("cycle", err)
	}
	return ClosedFromDegreeIntervals(degrees, cycle)
}

// Numbers returns the positions and adjacent sizes without an initial spelling.
func (m Mode) Numbers(sys *system.System) (Positions, error) {
	return m.numbers(sys, spelling.Spelling{}, false)
}

// NumbersFrom returns the positions and adjacent sizes starting at start.
func (m Mode) NumbersFrom(sys *system.System, start spelling.Spelling) (Positions, error) {
	return m.numbers(sys, start, true)
}

func (m Mode) numbers(sys *system.System, start spelling.Spelling, requireStart bool) (Positions, error) {
	if !m.IsValid() || !sys.IsValid() || (requireStart && !start.IsValid()) {
		return Positions{}, ErrUnset
	}
	return numbersFromIntervals(sys, start, m.prefixes, m.cycle)
}

func numbersFromIntervals(sys *system.System, start spelling.Spelling, prefixes []interval.Interval, cycle interval.Interval) (Positions, error) {
	return collectPositions(len(prefixes), cycle.IsValid(), func(i int) (system.Step, error) {
		if i == len(prefixes) {
			result, err := cycle.Steps(sys, start)
			if err != nil {
				return 0, positionError("cycle", err)
			}
			return result, nil
		}
		result, err := prefixes[i].Steps(sys, start)
		if err != nil {
			return 0, positionError(fmt.Sprintf("degree %d", i+1), err)
		}
		return result, nil
	})
}

// collectPositions calls at in order, including the cycle at count when present.
func collectPositions(count int, hasCycle bool, at func(int) (system.Step, error)) (Positions, error) {
	result := Positions{
		Degrees:  make([]system.Step, count),
		Adjacent: make([]system.Step, count-1),
		HasCycle: hasCycle,
	}
	for i := range result.Degrees {
		n, err := at(i)
		if err != nil {
			return Positions{}, err
		}
		result.Degrees[i] = n
	}
	if hasCycle {
		n, err := at(count)
		if err != nil {
			return Positions{}, err
		}
		result.Cycle = n
	}
	for i := range result.Adjacent {
		d, err := stepDifference(result.Degrees[i+1], result.Degrees[i])
		if err != nil {
			return Positions{}, err
		}
		result.Adjacent[i] = d
	}
	if result.HasCycle {
		d, err := stepDifference(result.Cycle, result.Degrees[count-1])
		if err != nil {
			return Positions{}, err
		}
		result.Adjacent = append(result.Adjacent, d)
	}
	return result, nil
}

func stepDifference(a, b system.Step) (system.Step, error) {
	difference := new(big.Int).Sub(big.NewInt(int64(a)), big.NewInt(int64(b)))
	if !difference.IsInt64() {
		return 0, errors.Join(ErrOutOfRange, interval.ErrOutOfRange)
	}
	return system.Step(difference.Int64()), nil
}

// SameInSystem reports whether the modes have the same kind and adjacent sizes.
func (m Mode) SameInSystem(sys *system.System, other Mode) (bool, error) {
	return m.sameInSystem(sys, spelling.Spelling{}, other, spelling.Spelling{}, false)
}

// SameInSystemFrom compares adjacent sizes under two given initial spellings.
func (m Mode) SameInSystemFrom(sys *system.System, start spelling.Spelling, other Mode, otherStart spelling.Spelling) (bool, error) {
	return m.sameInSystem(sys, start, other, otherStart, true)
}

func (m Mode) sameInSystem(sys *system.System, start spelling.Spelling, other Mode, otherStart spelling.Spelling, requireStart bool) (bool, error) {
	if !m.IsValid() || !other.IsValid() || !sys.IsValid() || (requireStart && (!start.IsValid() || !otherStart.IsValid())) {
		return false, ErrUnset
	}
	first, err := m.numbers(sys, start, requireStart)
	if err != nil {
		return false, err
	}
	second, err := other.numbers(sys, otherStart, requireStart)
	if err != nil {
		return false, err
	}
	return m.Kind() == other.Kind() && slices.Equal(first.Adjacent, second.Adjacent), nil
}

// CycleIn compares the cycle with the octave of a regular system.
// Only the size of the cycle needs to fit in a Step.
func (m Mode) CycleIn(sys *system.System) (CycleInfo, error) {
	return m.cycleIn(sys, spelling.Spelling{}, false)
}

// CycleInFrom compares the cycle with the system octave from start.
// The initial spelling, all degree spellings, and the spelling at the end of
// the cycle must be admissible in sys.
func (m Mode) CycleInFrom(sys *system.System, start spelling.Spelling) (CycleInfo, error) {
	return m.cycleIn(sys, start, true)
}

func (m Mode) checkClosedIn(sys *system.System, start spelling.Spelling, requireStart bool) error {
	if !sys.IsValid() || (requireStart && !start.IsValid()) {
		return ErrUnset
	}
	return m.checkClosed()
}

func (m Mode) cycleIn(sys *system.System, start spelling.Spelling, requireStart bool) (CycleInfo, error) {
	if err := m.checkClosedIn(sys, start, requireStart); err != nil {
		return CycleInfo{}, err
	}
	if requireStart {
		if err := m.checkDegreeNames(sys, start); err != nil {
			return CycleInfo{}, err
		}
	}
	cycleSize, err := m.cycle.Steps(sys, start)
	if err != nil {
		return CycleInfo{}, positionError("cycle", err)
	}
	cycle, octave := int(cycleSize), sys.Classes()
	if cycle < 1 {
		return CycleInfo{}, ErrNonPositiveNumber
	}
	gcd := cycle
	for remainder := octave; remainder != 0; {
		gcd, remainder = remainder, gcd%remainder
	}
	return CycleInfo{Comparison: cmp.Compare(cycle, octave), Cycles: octave / gcd, Octaves: cycle / gcd}, nil
}

func (m Mode) checkDegreeNames(sys *system.System, start spelling.Spelling) error {
	for i, prefix := range m.prefixes {
		// The first prefix is P1, so this also validates start before any
		// subsequent names. No intermediate position is needed here.
		name, err := interval.Transpose(start, prefix)
		if err != nil {
			return positionError(fmt.Sprintf("degree %d", i+1), err)
		}
		admitted, err := sys.Admits(name)
		if err != nil {
			return fmt.Errorf("degree %d: %w", i+1, err)
		}
		if !admitted {
			return fmt.Errorf("degree %d (%v): %w", i+1, name, system.ErrInadmissible)
		}
	}
	return nil
}

// SymmetryIn describes the numerical symmetry in a regular system.
func (m Mode) SymmetryIn(sys *system.System) (Symmetry, error) {
	return m.symmetryIn(sys, spelling.Spelling{}, false)
}

// SymmetryInFrom describes the numerical symmetry from start.
func (m Mode) SymmetryInFrom(sys *system.System, start spelling.Spelling) (Symmetry, error) {
	return m.symmetryIn(sys, start, true)
}

func (m Mode) symmetryIn(sys *system.System, start spelling.Spelling, requireStart bool) (Symmetry, error) {
	if err := m.checkClosedIn(sys, start, requireStart); err != nil {
		return Symmetry{}, err
	}
	positions, err := m.numbers(sys, start, requireStart)
	if err != nil {
		return Symmetry{}, err
	}
	octave := sys.Classes()
	if int(positions.Cycle) != octave {
		return Symmetry{}, ErrSystemOctaveRequired
	}
	for _, n := range positions.Adjacent {
		if n < 1 {
			return Symmetry{}, ErrNonPositiveNumber
		}
	}
	period := leastPeriod(positions.Adjacent)
	return Symmetry{
		Rotations:      period,
		Transpositions: octave / (m.Len() / period),
		Limited:        period < m.Len(),
	}, nil
}
