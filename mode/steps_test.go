package mode_test

import (
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The expected distances below come from business-spec 21.2, KP-13, and from
// the rules P-33 and F-8, F-9. They are not derived from the code under test.

// stepsRequest describes one call of StepsBetween (named == false) or
// StepsBetweenFrom (named == true) together with its Must form.
type stepsRequest struct {
	m        mode.Mode
	sys      *system.System
	named    bool
	start    spelling.Spelling
	from, to mode.DegreePlace
}

func place(number, cycle int) mode.DegreePlace {
	return mode.DegreePlace{Number: number, Cycle: cycle}
}

func unnamed(m mode.Mode, sys *system.System, from, to mode.DegreePlace) stepsRequest {
	return stepsRequest{m: m, sys: sys, from: from, to: to}
}

func namedFrom(m mode.Mode, sys *system.System, start string, from, to mode.DegreePlace) stepsRequest {
	return stepsRequest{m: m, sys: sys, named: true, start: spelling.MustParse(start), from: from, to: to}
}

func (r stepsRequest) run() (system.Step, error) {
	if r.named {
		return r.m.StepsBetweenFrom(r.sys, r.start, r.from, r.to)
	}
	return r.m.StepsBetween(r.sys, r.from, r.to)
}

func (r stepsRequest) must() system.Step {
	if r.named {
		return r.m.MustStepsBetweenFrom(r.sys, r.start, r.from, r.to)
	}
	return r.m.MustStepsBetween(r.sys, r.from, r.to)
}

// stepsCauses lists the causes that distinguish the failures of StepsBetween*
// besides the mode categories.
func stepsCauses() []error {
	return []error{interval.ErrStartRequired, interval.ErrFractionalSteps, system.ErrInadmissible, interval.ErrOutOfRange}
}

func checkStepsSuccess(t *testing.T, r stepsRequest, want system.Step) {
	t.Helper()
	got, err := r.run()
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, want, r.must())
}

// checkStepsFailure checks the promised causes, the absence of every other
// mode category and distance cause, the zero result, and that the Must form
// panics with the same error.
func checkStepsFailure(t *testing.T, r stepsRequest, causes ...error) {
	t.Helper()
	got, err := r.run()
	require.Error(t, err)
	assert.Zero(t, got, "a failure returns 0")
	for _, cause := range causes {
		assert.ErrorIs(t, err, cause)
	}
	for _, other := range append(modeErrors(), stepsCauses()...) {
		if !slices.Contains(causes, other) {
			assert.NotErrorIs(t, err, other, "only the first failing check is reported")
		}
	}
	recovered := capturePanic(func() { r.must() })
	require.NotNil(t, recovered, "the Must form panics on failure")
	mustErr, ok := recovered.(error)
	require.True(t, ok, "the Must form panics with an error, got %T", recovered)
	assert.Equal(t, err.Error(), mustErr.Error(), "the Must form panics with the same error")
	for _, cause := range causes {
		assert.ErrorIs(t, mustErr, cause)
	}
}

func TestStepsBetweenControlValues(t *testing.T) {
	t.Parallel()
	const maxInt = math.MaxInt
	major := mode.MustFromCode("ionian")
	fourth := mode.MustParse("M2 M2 m2")
	wide := mode.MustParse("M9 m7")
	unisonPair := mode.MustParse("d2 A7")
	below := mode.MustParse("dd2 AA7")
	triple := mode.MustParse("A1 dd2 A7")
	openThird := mode.MustParse("m3 M2 |")
	neutral := mode.MustOpenFromDegreeIntervals(intervals("P1 n3 n7"))
	augmentedUnison := mode.MustClosedFromIntervals(intervals("A1"))
	traditional, p19, q24, irregular := system.Traditional(), system.Positions19(), system.QuarterTones24(), irregularSystem()

	for _, tt := range []struct {
		name string
		r    stepsRequest
		want system.Step
	}{
		{"issue: ionian C, 1 to 3", namedFrom(major, traditional, "C", place(1, 0), place(3, 0)), 4},
		{"issue: ionian C, 7 down to 1", namedFrom(major, traditional, "C", place(7, 0), place(1, 0)), -11},
		{"issue: ionian C, 7 to next 1", namedFrom(major, traditional, "C", place(7, 0), place(1, 1)), 1},
		{"ionian without a name", unnamed(major, traditional, place(1, 0), place(3, 0)), 4},
		{"issue: ionian in 24 positions", namedFrom(major, q24, "C", place(1, 0), place(3, 0)), 8},
		{"issue: ionian in 19 positions", namedFrom(major, p19, "C", place(1, 0), place(3, 0)), 6},
		{"ionian in 24 positions without a name", unnamed(major, q24, place(1, 0), place(3, 0)), 8},
		{"ionian in 19 positions without a name", unnamed(major, p19, place(1, 0), place(3, 0)), 6},
		{"same place", namedFrom(major, traditional, "C", place(3, 0), place(3, 0)), 0},
		{"next cycle", namedFrom(major, traditional, "C", place(1, 0), place(1, 1)), 12},
		{"previous cycle", namedFrom(major, traditional, "C", place(1, 0), place(1, -1)), -12},
		{"two cycles down", namedFrom(major, traditional, "C", place(1, 1), place(1, -1)), -24},

		{"irregular from C, 1 to 2", namedFrom(major, irregular, "C", place(1, 0), place(2, 0)), 3},
		{"irregular from C, 1 to 3", namedFrom(major, irregular, "C", place(1, 0), place(3, 0)), 4},
		{"irregular from C, 2 to 3", namedFrom(major, irregular, "C", place(2, 0), place(3, 0)), 1},
		{"irregular from C, 7 to next 1", namedFrom(major, irregular, "C", place(7, 0), place(1, 1)), 1},
		{"irregular from C, 7 down to 1", namedFrom(major, irregular, "C", place(7, 0), place(1, 0)), -11},
		{"irregular from D, 1 to 2", namedFrom(major, irregular, "D", place(1, 0), place(2, 0)), 1},
		{"irregular from D, 1 to 3", namedFrom(major, irregular, "D", place(1, 0), place(3, 0)), 3},
		{"irregular from D, 2 to 3", namedFrom(major, irregular, "D", place(2, 0), place(3, 0)), 2},
		{"irregular from D, 7 to next 1", namedFrom(major, irregular, "D", place(7, 0), place(1, 1)), 2},
		{"irregular from D, 7 down to 1", namedFrom(major, irregular, "D", place(7, 0), place(1, 0)), -10},

		{"P4 cycle from G, one cycle", namedFrom(fourth, traditional, "G", place(1, 0), place(1, 1)), 5},
		{"P4 cycle from G, 3 to 1 three cycles up", namedFrom(fourth, traditional, "G", place(3, 0), place(1, 3)), 11},
		{"P4 cycle from G, down across cycles", namedFrom(fourth, traditional, "G", place(3, 2), place(1, 1)), -9},
		{"P4 cycle from G, seven cycles", namedFrom(fourth, traditional, "G", place(1, 0), place(1, 7)), 35},
		{"P4 cycle irregular from D, first cycle", namedFrom(fourth, irregular, "D", place(1, 0), place(1, 1)), 4},
		{"P4 cycle irregular from D, second cycle", namedFrom(fourth, irregular, "D", place(1, 1), place(1, 2)), 5},
		{"P4 cycle irregular from D, two cycles", namedFrom(fourth, irregular, "D", place(1, 0), place(1, 2)), 9},
		{"P4 cycle irregular from D, negative cycle", namedFrom(fourth, irregular, "D", place(1, 0), place(1, -1)), -6},

		{"P15 cycle, 1 to 2", namedFrom(wide, traditional, "C", place(1, 0), place(2, 0)), 14},
		{"P15 cycle, 2 to next 1", namedFrom(wide, traditional, "C", place(2, 0), place(1, 1)), 10},
		{"P15 cycle, previous cycle", namedFrom(wide, traditional, "C", place(1, 0), place(1, -1)), -24},

		{"d2 A7, coinciding degrees", unnamed(unisonPair, traditional, place(1, 0), place(2, 0)), 0},
		{"d2 A7, 2 to next 1", unnamed(unisonPair, traditional, place(2, 0), place(1, 1)), 12},
		{"d2 A7 in 19 positions", unnamed(unisonPair, p19, place(1, 0), place(2, 0)), 1},
		{"dd2 AA7, negative upward by number", unnamed(below, traditional, place(1, 0), place(2, 0)), -1},
		{"A1 dd2 A7, non-adjacent coinciding degrees", unnamed(triple, traditional, place(1, 0), place(3, 0)), 0},
		{"A1 dd2 A7, 2 to 3", unnamed(triple, traditional, place(2, 0), place(3, 0)), -1},
		{"open m3 M2 from E, up", namedFrom(openThird, traditional, "E", place(1, 0), place(3, 0)), 5},
		{"open m3 M2 from E, down", namedFrom(openThird, traditional, "E", place(3, 0), place(1, 0)), -5},

		{"whole distance between fractional degrees", unnamed(neutral, traditional, place(2, 0), place(3, 0)), 7},
		{"neutral degrees in 24 positions, 1 to 2", namedFrom(neutral, q24, "C", place(1, 0), place(2, 0)), 7},
		{"neutral degrees in 24 positions, 2 to 3", namedFrom(neutral, q24, "C", place(2, 0), place(3, 0)), 14},

		{"representable distance at MaxInt cycles", unnamed(major, traditional, place(1, maxInt), place(2, maxInt)), 2},
		{"representable distance at MaxInt cycles from C", namedFrom(major, traditional, "C", place(1, maxInt), place(2, maxInt)), 2},
		{"P4 cycle at MaxInt cycles without a name", unnamed(fourth, traditional, place(1, maxInt), place(2, maxInt)), 2},
		{"representable distance at MinInt cycles", unnamed(major, traditional, place(1, math.MinInt), place(2, math.MinInt)), 2},
		{"representable distance at MinInt cycles from C", namedFrom(major, traditional, "C", place(1, math.MinInt), place(2, math.MinInt)), 2},
		{"A1 cycle, 2^25 cycles without a name", unnamed(augmentedUnison, traditional, place(1, 0), place(1, 1<<25)), 1 << 25},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkStepsSuccess(t, tt.r, tt.want)
		})
	}
}

func TestStepsBetweenFailures(t *testing.T) {
	t.Parallel()
	const maxInt = math.MaxInt
	major := mode.MustFromCode("ionian")
	fourth := mode.MustParse("M2 M2 m2")
	openThird := mode.MustParse("m3 M2 |")
	neutral := mode.MustOpenFromDegreeIntervals(intervals("P1 n3 n7"))
	augmentedUnison := mode.MustClosedFromIntervals(intervals("A1"))
	traditional, irregular := system.Traditional(), irregularSystem()
	// From G, every seven P4 cycles lower the alteration by one: place (1, 21)
	// is G with -3, outside this system, while (1, MaxInt) has no representable
	// alteration at all.
	limited := system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(-2, 2), "Limited accidentals")
	unsetStart := stepsRequest{m: major, sys: traditional, named: true, from: place(1, 0), to: place(3, 0)}
	unsetStartBadNumber := stepsRequest{m: major, sys: traditional, named: true, from: place(0, 0), to: place(3, 0)}
	outOfName := []error{mode.ErrOutOfRange, interval.ErrOutOfRange, alteration.ErrOutOfRange}
	outOfResult := []error{mode.ErrOutOfRange, interval.ErrOutOfRange}

	for _, tt := range []struct {
		name   string
		r      stepsRequest
		causes []error
	}{
		// E-1: unset inputs.
		{"unset mode", unnamed(mode.Mode{}, traditional, place(1, 0), place(1, 0)), []error{mode.ErrUnset}},
		{"unset mode, named", namedFrom(mode.Mode{}, traditional, "C", place(1, 0), place(1, 0)), []error{mode.ErrUnset}},
		{"nil system", unnamed(major, nil, place(1, 0), place(3, 0)), []error{mode.ErrUnset}},
		{"nil system, named", namedFrom(major, nil, "C", place(1, 0), place(3, 0)), []error{mode.ErrUnset}},
		{"zero system", unnamed(major, new(system.System), place(1, 0), place(3, 0)), []error{mode.ErrUnset}},
		{"zero system, named", namedFrom(major, new(system.System), "C", place(1, 0), place(3, 0)), []error{mode.ErrUnset}},
		{"unset start", unsetStart, []error{mode.ErrUnset}},
		{"unset mode before invalid number", unnamed(mode.Mode{}, traditional, place(0, 0), place(9, 0)), []error{mode.ErrUnset}},
		{"nil system before invalid number", unnamed(major, nil, place(0, 0), place(9, 0)), []error{mode.ErrUnset}},
		{"unset start before invalid number", unsetStartBadNumber, []error{mode.ErrUnset}},

		// E-2: places outside the mode.
		{"from number 0", namedFrom(major, traditional, "C", place(0, 0), place(1, 0)), []error{mode.ErrInvalidArgument}},
		{"to number 8", namedFrom(major, traditional, "C", place(1, 0), place(8, 0)), []error{mode.ErrInvalidArgument}},
		{"from number -1", namedFrom(major, traditional, "C", place(-1, 0), place(1, 0)), []error{mode.ErrInvalidArgument}},
		{"to number 0 without a name", unnamed(major, traditional, place(1, 0), place(0, 0)), []error{mode.ErrInvalidArgument}},
		{"from number MinInt", unnamed(major, traditional, place(math.MinInt, 0), place(1, 0)), []error{mode.ErrInvalidArgument}},
		{"to number MaxInt", unnamed(major, traditional, place(1, 0), place(maxInt, 0)), []error{mode.ErrInvalidArgument}},
		{"open mode, from cycle 1", namedFrom(openThird, traditional, "E", place(1, 0), place(1, 1)), []error{mode.ErrInvalidArgument}},
		{"open mode, to number 4", namedFrom(openThird, traditional, "E", place(1, 0), place(4, 0)), []error{mode.ErrInvalidArgument}},
		{"open mode, from cycle -1 without a name", unnamed(openThird, traditional, place(1, -1), place(1, 0)), []error{mode.ErrInvalidArgument}},
		{"invalid number before irregular system", unnamed(major, irregular, place(1, 0), place(8, 0)), []error{mode.ErrInvalidArgument}},
		{"invalid number before inadmissible start", namedFrom(major, traditional, "C+1/2", place(0, 0), place(1, 0)), []error{mode.ErrInvalidArgument}},
		{"open cycle before inadmissible name", namedFrom(neutral, traditional, "C", place(2, 1), place(3, 0)), []error{mode.ErrInvalidArgument}},

		// Causes of interval and system.
		{"irregular system without a name", unnamed(major, irregular, place(1, 0), place(3, 0)), []error{interval.ErrStartRequired}},
		{"irregular system without a name, same place", unnamed(major, irregular, place(1, 0), place(1, 0)), []error{interval.ErrStartRequired}},
		{"start required before out of range result", unnamed(major, irregular, place(1, 0), place(1, maxInt)), []error{interval.ErrStartRequired}},
		{"start required before fractional distance", unnamed(neutral, irregular, place(1, 0), place(2, 0)), []error{interval.ErrStartRequired}},
		{"inadmissible start", namedFrom(major, traditional, "C+1/2", place(1, 0), place(3, 0)), []error{system.ErrInadmissible}},
		{"inadmissible start before unrepresentable place name", namedFrom(fourth, traditional, "G+1/2", place(1, maxInt), place(2, maxInt)), []error{system.ErrInadmissible}},
		{"inadmissible name of the to place", namedFrom(neutral, traditional, "C", place(1, 0), place(2, 0)), []error{system.ErrInadmissible}},
		{"inadmissible name of the from place", namedFrom(neutral, traditional, "C", place(2, 0), place(3, 0)), []error{system.ErrInadmissible}},
		{"inadmissible to before unrepresentable from", namedFrom(fourth, limited, "G", place(1, maxInt), place(1, 21)), []error{system.ErrInadmissible}},
		{"fractional distance without a name", unnamed(neutral, traditional, place(1, 0), place(2, 0)), []error{interval.ErrFractionalSteps}},

		// E-11: unrepresentable names and results.
		{"unrepresentable names of both places", namedFrom(fourth, traditional, "G", place(1, maxInt), place(2, maxInt)), outOfName},
		{"unrepresentable name of the to place", namedFrom(fourth, traditional, "G", place(1, 0), place(1, maxInt)), outOfName},
		{"unrepresentable name after 2^25 A1 cycles", namedFrom(augmentedUnison, traditional, "C", place(1, 0), place(1, 1<<25)), outOfName},
		{"result out of int without a name", unnamed(major, traditional, place(1, 0), place(1, maxInt)), outOfResult},
		{"result out of int downward without a name", unnamed(major, traditional, place(1, 0), place(1, -maxInt)), outOfResult},
		{"result out of int from C", namedFrom(major, traditional, "C", place(1, 0), place(1, maxInt)), outOfResult},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkStepsFailure(t, tt.r, tt.causes...)
		})
	}
}

// stepsFixture is a mode in a system, with or without the name of the first
// degree, whose accepted Numbers succeed.
type stepsFixture struct {
	name  string
	m     mode.Mode
	sys   *system.System
	start string
}

func (f stepsFixture) request(from, to mode.DegreePlace) stepsRequest {
	if f.start == "" {
		return unnamed(f.m, f.sys, from, to)
	}
	return namedFrom(f.m, f.sys, f.start, from, to)
}

func (f stepsFixture) numbers(t *testing.T) mode.Positions {
	t.Helper()
	var p mode.Positions
	var err error
	if f.start == "" {
		p, err = f.m.Numbers(f.sys)
	} else {
		p, err = f.m.NumbersFrom(f.sys, spelling.MustParse(f.start))
	}
	require.NoError(t, err, "the comparison needs accepted numbers")
	return p
}

// places lists every actual degree in the cycles -1, 0, and 2, or only in
// cycle 0 for an open mode.
func (f stepsFixture) places() []mode.DegreePlace {
	cycles := []int{-1, 0, 2}
	if f.m.IsOpen() {
		cycles = []int{0}
	}
	var result []mode.DegreePlace
	for number := 1; number <= f.m.Len(); number++ {
		for _, cycle := range cycles {
			result = append(result, place(number, cycle))
		}
	}
	return result
}

func stepsFixtures() []stepsFixture {
	traditional, irregular := system.Traditional(), irregularSystem()
	return []stepsFixture{
		{"ionian traditional", mode.MustFromCode("ionian"), traditional, ""},
		{"ionian traditional from F+1", mode.MustFromCode("ionian"), traditional, "F+1"},
		{"ionian 19 positions", mode.MustFromCode("ionian"), system.Positions19(), ""},
		{"ionian 24 positions from C", mode.MustFromCode("ionian"), system.QuarterTones24(), "C"},
		{"ionian irregular from C", mode.MustFromCode("ionian"), irregular, "C"},
		{"ionian irregular from D", mode.MustFromCode("ionian"), irregular, "D"},
		{"P4 cycle traditional", mode.MustParse("M2 M2 m2"), traditional, ""},
		{"P4 cycle irregular from D", mode.MustParse("M2 M2 m2"), irregular, "D"},
		{"P15 cycle traditional", mode.MustParse("M9 m7"), traditional, ""},
		{"d2 A7 traditional", mode.MustParse("d2 A7"), traditional, ""},
		{"dd2 AA7 traditional", mode.MustParse("dd2 AA7"), traditional, ""},
		{"A1 dd2 A7 traditional", mode.MustParse("A1 dd2 A7"), traditional, ""},
		{"open m3 M2 traditional from E", mode.MustParse("m3 M2 |"), traditional, "E"},
	}
}

func TestStepsBetweenMatchesNumbersInCycleZero(t *testing.T) {
	t.Parallel()
	for _, f := range stepsFixtures() {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			p := f.numbers(t)
			require.Len(t, p.Degrees, f.m.Len())
			for i := 1; i <= f.m.Len(); i++ {
				for j := 1; j <= f.m.Len(); j++ {
					got, err := f.request(place(i, 0), place(j, 0)).run()
					require.NoError(t, err, "%d to %d", i, j)
					assert.Equal(t, p.Degrees[j-1]-p.Degrees[i-1], got, "%d to %d", i, j)
				}
			}
		})
	}
}

func TestStepsBetweenIsAnOrientedDistance(t *testing.T) {
	t.Parallel()
	for _, f := range stepsFixtures() {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			places := f.places()
			distance := make(map[[2]mode.DegreePlace]system.Step, len(places)*len(places))
			for _, a := range places {
				for _, b := range places {
					got, err := f.request(a, b).run()
					require.NoError(t, err, "%v to %v", a, b)
					distance[[2]mode.DegreePlace{a, b}] = got
				}
			}
			for _, a := range places {
				assert.Zero(t, distance[[2]mode.DegreePlace{a, a}], "%v to itself", a)
				for _, b := range places {
					ab := distance[[2]mode.DegreePlace{a, b}]
					assert.Equal(t, -ab, distance[[2]mode.DegreePlace{b, a}], "%v and %v differ only in sign", a, b)
					for _, c := range places {
						assert.Equal(t, distance[[2]mode.DegreePlace{a, c}], ab+distance[[2]mode.DegreePlace{b, c}], "%v to %v to %v", a, b, c)
					}
				}
			}
		})
	}
}

func TestStepsBetweenCyclesInRegularAndIrregularSystems(t *testing.T) {
	t.Parallel()
	t.Run("one cycle has the same size everywhere in a regular system", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name  string
			m     mode.Mode
			sys   *system.System
			cycle system.Step
		}{
			{"ionian traditional", mode.MustFromCode("ionian"), system.Traditional(), 12},
			{"ionian 19 positions", mode.MustFromCode("ionian"), system.Positions19(), 19},
			{"ionian 24 positions", mode.MustFromCode("ionian"), system.QuarterTones24(), 24},
			{"P4 cycle traditional", mode.MustParse("M2 M2 m2"), system.Traditional(), 5},
			{"P15 cycle traditional", mode.MustParse("M9 m7"), system.Traditional(), 24},
		} {
			for number := 1; number <= tt.m.Len(); number++ {
				for _, cycle := range []int{-3, -1, 0, 4} {
					checkStepsSuccess(t, unnamed(tt.m, tt.sys, place(number, cycle), place(number, cycle+1)), tt.cycle)
				}
			}
		}
	})
	t.Run("the forms with and without a name agree in a regular system", func(t *testing.T) {
		t.Parallel()
		for _, f := range stepsFixtures() {
			// Fixtures without a name are all in regular systems.
			if f.start != "" {
				continue
			}
			for _, start := range []string{"C", "D", "F+1", "B-1"} {
				for _, a := range f.places() {
					for _, b := range f.places() {
						want, err := f.request(a, b).run()
						require.NoError(t, err)
						got, err := namedFrom(f.m, f.sys, start, a, b).run()
						require.NoError(t, err, "%s from %s, %v to %v", f.name, start, a, b)
						assert.Equal(t, want, got, "%s from %s, %v to %v", f.name, start, a, b)
					}
				}
			}
		}
	})
	t.Run("a cycle of the P4 mode depends on the place in an irregular system", func(t *testing.T) {
		t.Parallel()
		fourth := mode.MustParse("M2 M2 m2")
		first := namedFrom(fourth, irregularSystem(), "D", place(1, 0), place(1, 1))
		second := namedFrom(fourth, irregularSystem(), "D", place(1, 1), place(1, 2))
		checkStepsSuccess(t, first, 4)
		checkStepsSuccess(t, second, 5)
	})
}

func TestStepsBetweenDoesNotChangeInputs(t *testing.T) {
	t.Parallel()
	m := mode.MustParse("M2 M2 m2")
	saved := m
	sys := system.Traditional()
	from, to := place(3, 2), place(1, 1)
	for range 3 {
		checkStepsSuccess(t, namedFrom(m, sys, "G", from, to), -9)
		checkStepsSuccess(t, unnamed(m, sys, from, to), -9)
	}
	assert.True(t, saved.Equal(m))
	assert.Equal(t, place(3, 2), from)
	assert.Equal(t, place(1, 1), to)
	assert.Equal(t, mode.Positions{Degrees: []system.Step{0, 2, 4}, Adjacent: []system.Step{2, 2, 1}, Cycle: 5, HasCycle: true}, m.MustNumbers(sys))
}

func TestStepsBetweenConcurrentCalls(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("ionian")
	sys := irregularSystem()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 50 {
				got, err := m.StepsBetweenFrom(sys, spelling.MustParse("D"), place(7, 0), place(1, 1))
				assert.NoError(t, err)
				assert.Equal(t, system.Step(2), got)
				got, err = m.StepsBetween(system.Traditional(), place(1, -1), place(3, 2))
				assert.NoError(t, err)
				assert.Equal(t, system.Step(40), got)
			}
		})
	}
	wg.Wait()
}
