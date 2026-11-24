package mode_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the open mode built from numbers and degree sizes (issue #146,
// item 3; business-spec, KP-17 and F-14).

// openFromNumbers calls the form without a start name for start "" and the
// form with start otherwise; "unset" passes the unset name.
func openFromNumbers(sys *system.System, adjacent []system.Step, start string, sizes []int) (mode.Mode, error) {
	if start == "" {
		return mode.OpenFromNumbers(sys, adjacent, sizes)
	}
	return mode.OpenFromNumbersFrom(sys, adjacent, spelling.MustParse(start), sizes)
}

func isRegular(t *testing.T, sys *system.System) bool {
	t.Helper()
	regularity, err := sys.Regularity()
	require.NoError(t, err)
	return regularity.Regular
}

func TestOpenFromNumbers(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	for _, tt := range []struct {
		name     string
		sys      *system.System
		start    string
		adjacent []system.Step
		sizes    []int
		want     string
		names    string
	}{
		{"issue example", traditional, "C", stepList(2, 1), []int{1, 2, 3}, "M2 m2", "C D E-1"},
		{"issue example without a name", traditional, "", stepList(2, 1), []int{1, 2, 3}, "M2 m2", ""},
		{"smallest mode", traditional, "C", stepList(2), []int{1, 2}, "M2", "C D"},
		{"repeated size", traditional, "C", stepList(1), []int{1, 1}, "A1", "C C+1"},
		{"open mode of KP-12", traditional, "E", stepList(3, 2), []int{1, 3, 4}, "m3 M2", "E G A"},
		{"same numbers with another plan", traditional, "E", stepList(3, 2), []int{1, 2, 4}, "A2 d3", "E F+2 A"},
		{"repeated letter inside", traditional, "C", stepList(3, 2, 1, 1), []int{1, 3, 4, 4, 5}, "m3 M2 A1 m2", "C E-1 F F+1 G"},
		{"sizes above 8", traditional, "C", stepList(2, 2, 1, 2, 2, 2, 1, 2), []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			"M2 M2 m2 M2 M2 M2 m2 M2", "C D E F G A B C D"},
		{"ninth without a name", traditional, "", stepList(14), []int{1, 9}, "M9", ""},
		{"19-position", system.Positions19(), "C", stepList(3, 2), []int{1, 2, 3}, "M2 m2", "C D E-1"},
		{"24-position fractional quality", system.QuarterTones24(), "C", stepList(3, 4), []int{1, 2, 3}, "n2 M2", "C D-1/2 E-1/2"},
		{"31 sharps and flats without a name", system.SharpsAndFlats31(), "", stepList(4, 5), []int{1, 2, 3}, "n2 M2", ""},
		{"irregular from C", irregularSystem(), "C", stepList(2, 1), []int{1, 2, 3}, "m2 M2", "C D-1 E-1"},
		{"irregular from D", irregularSystem(), "D", stepList(2, 1), []int{1, 2, 3}, "A2 m2", "D E+1 F+1"},
		{"naturals only without a name", naturalsOnlySystem(), "", stepList(2, 1), []int{1, 2, 3}, "M2 m2", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			forms := []string{tt.start}
			if tt.start != "" && isRegular(t, tt.sys) {
				// In a regular system the form without a name gives the same mode.
				forms = append(forms, "")
			}
			want := mode.MustParse(tt.want + " |")
			for _, start := range forms {
				adjacent, sizes := slices.Clone(tt.adjacent), slices.Clone(tt.sizes)
				m, err := openFromNumbers(tt.sys, adjacent, start, sizes)
				require.NoError(t, err, "start %q", start)
				assert.Equal(t, tt.adjacent, adjacent, "the input numbers must not change")
				assert.Equal(t, tt.sizes, sizes, "the input sizes must not change")
				assert.True(t, m.Equal(want), "start %q: got %s, want %s", start, m, want)
				assert.True(t, m.IsOpen())
				assert.False(t, m.IsClosed())
				assert.False(t, m.HasOrigin())
				_, hasCycle := m.Cycle()
				assert.False(t, hasCycle)
				assert.Equal(t, len(tt.sizes), m.Len())

				// F-14: the passed numbers and sizes are read back, and the mode
				// equals the open mode of its own degree intervals.
				var numbers mode.Positions
				if start == "" {
					numbers = m.MustNumbers(tt.sys)
				} else {
					numbers = m.MustNumbersFrom(tt.sys, spelling.MustParse(start))
					assert.Equal(t, tt.names, degreeNames(t, m, tt.sys, start))
				}
				assert.Equal(t, tt.adjacent, numbers.Adjacent)
				assert.False(t, numbers.HasCycle)
				plan, closed := m.LetterPlan()
				assert.False(t, closed)
				assert.Equal(t, mode.LetterPlan{Degrees: tt.sizes, Cycle: 0}, plan)
				assert.True(t, m.Equal(mode.MustOpenFromDegreeIntervals(m.DegreeIntervals())))
			}
		})
	}

	long := mode.MustOpenFromNumbersFrom(traditional, stepList(2, 2, 1, 2, 2, 2, 1, 2), spelling.MustParse("C"), []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	assert.Equal(t, 9, long.Len())
	assert.Equal(t, "M9", long.MustDegreeInterval(9).String())
}

func TestOpenFromNumbersInputsAreNotRetained(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	adjacent := stepList(3, 2, 1, 1)
	sizes := []int{1, 3, 4, 4, 5}
	named := mode.MustOpenFromNumbersFrom(traditional, adjacent, spelling.MustParse("C"), sizes)
	unnamed := mode.MustOpenFromNumbers(traditional, adjacent, sizes)
	adjacent[0], sizes[1] = 4, 2

	want := "m3 M2 A1 m2 |"
	checkMode(t, named, want)
	checkMode(t, unnamed, want)
	plan, _ := named.LetterPlan()
	assert.Equal(t, []int{1, 3, 4, 4, 5}, plan.Degrees)
	plan.Degrees[0] = 7
	again, _ := named.LetterPlan()
	assert.Equal(t, []int{1, 3, 4, 4, 5}, again.Degrees)
	checkMode(t, named, want)
}

func TestOpenFromNumbersFailures(t *testing.T) {
	t.Parallel()
	traditional := system.Traditional()
	t0 := naturalsOnlySystem()
	t1 := singleAccidentalSystem()
	irregular := irregularSystem()
	outOfRange := []error{interval.ErrOutOfRange}
	both := []string{"", "C"}
	tooManyWithZero := ones(mode.MaxDegrees)
	tooManyWithZero[0] = 0
	manySizes := make([]int, mode.MaxDegrees+1)
	for i := range manySizes {
		manySizes[i] = 1
	}
	for _, tt := range []struct {
		name     string
		sys      *system.System
		starts   []string
		adjacent []system.Step
		sizes    []int
		// category is the mode sentinel; nil means only the listed causes
		// of other packages, without any mode sentinel.
		category error
		causes   []error
	}{
		{"fewer sizes", traditional, both, stepList(2, 1), []int{1, 2}, mode.ErrWrongDegreeCount, nil},
		{"more sizes", traditional, both, stepList(2, 1), []int{1, 2, 3, 4}, mode.ErrWrongDegreeCount, nil},
		{"one number and three sizes", traditional, both, stepList(2), []int{1, 2, 3}, mode.ErrWrongDegreeCount, nil},
		{"nil sizes", traditional, both, stepList(2, 1), nil, mode.ErrWrongDegreeCount, nil},
		{"empty sizes", traditional, both, stepList(2, 1), []int{}, mode.ErrWrongDegreeCount, nil},
		{"no numbers and one size", traditional, both, nil, []int{1}, mode.ErrEmptyInput, nil},
		{"empty numbers and one size", traditional, both, []system.Step{}, []int{1}, mode.ErrEmptyInput, nil},
		{"no numbers and no sizes", traditional, both, nil, nil, mode.ErrEmptyInput, nil},
		{"numbers checked before incorrect sizes", traditional, both, nil, []int{2, 3}, mode.ErrEmptyInput, nil},
		{"first size not 1", traditional, both, stepList(2, 1), []int{2, 3, 4}, mode.ErrInvalidArgument, nil},
		{"decreasing sizes", traditional, both, stepList(2, 1), []int{1, 3, 2}, mode.ErrInvalidArgument, nil},
		{"zero number", traditional, both, stepList(2, 0), []int{1, 2, 3}, mode.ErrInvalidStructure, nil},
		{"negative number", traditional, both, stepList(-2, 1), []int{1, 2, 3}, mode.ErrInvalidStructure, nil},
		{"sizes checked before numbers", traditional, both, stepList(2, 0), []int{1, 3, 2}, mode.ErrInvalidArgument, nil},
		{"number checked before the size count", traditional, both, stepList(2, 0), []int{1, 2}, mode.ErrInvalidStructure, nil},
		{"sizes checked before the size count", traditional, both, stepList(2, 1), []int{1, 3, 2, 4}, mode.ErrInvalidArgument, nil},
		{"too many degrees whatever the sizes", traditional, both, ones(mode.MaxDegrees), []int{1, 2, 3}, mode.ErrInvalidArgument, nil},
		{"too many degrees before incorrect sizes", traditional, both, ones(mode.MaxDegrees), []int{2, 3}, mode.ErrInvalidArgument, nil},
		{"too many degrees before reading the numbers", traditional, both, tooManyWithZero, []int{1, 2, 3}, mode.ErrInvalidArgument, nil},
		{"the degree limit itself is allowed", traditional, both, ones(mode.MaxDegrees - 1), []int{1, 2, 3}, mode.ErrWrongDegreeCount, nil},
		{"the size count is not compared with the limit", traditional, both, stepList(2, 1), manySizes, mode.ErrWrongDegreeCount, nil},
		{"nil system", nil, both, stepList(2, 1), []int{1, 2, 3}, mode.ErrUnset, nil},
		{"unset system", new(system.System), both, stepList(2, 1), []int{1, 2, 3}, mode.ErrUnset, nil},
		{"nil system before missing numbers", nil, both, nil, nil, mode.ErrUnset, nil},
		{"unset start", traditional, []string{"unset"}, stepList(2, 1), []int{1, 2, 3}, mode.ErrUnset, nil},
		{"unset start before missing numbers", traditional, []string{"unset"}, nil, nil, mode.ErrUnset, nil},
		{"irregular system without a name", irregular, []string{""}, stepList(2, 1), []int{1, 2, 3}, nil,
			[]error{interval.ErrStartRequired}},
		{"size count checked before the start requirement", irregular, []string{""}, stepList(2, 1), []int{1, 2},
			mode.ErrWrongDegreeCount, nil},
		{"inadmissible start", t1, []string{"C+2"}, stepList(2, 1), []int{1, 2, 3}, nil, []error{system.ErrInadmissible}},
		{"number checked before an inadmissible start", t1, []string{"C+2"}, stepList(2, 0), []int{1, 2, 3},
			mode.ErrInvalidStructure, nil},
		{"inadmissible degree name", system.SharpsAndFlats31(), []string{"C"}, stepList(4, 5), []int{1, 2, 3}, nil,
			[]error{system.ErrInadmissible}},
		{"inadmissible E flat", t0, []string{"C"}, stepList(2, 1), []int{1, 2, 3}, nil, []error{system.ErrInadmissible}},
		{"deviation out of range", traditional, both, stepList(1 << 40), []int{1, 2}, mode.ErrOutOfRange, outOfRange},
		{"position out of range", traditional, both, stepList(math.MaxInt, 1), []int{1, 2, 3}, mode.ErrOutOfRange, outOfRange},
		{"inadmissible degree 2 before the position out of range", t0, []string{"C"}, stepList(1, math.MaxInt),
			[]int{1, 2, 3}, nil, []error{system.ErrInadmissible}},
		{"without names the position out of range is reported", t0, []string{""}, stepList(1, math.MaxInt),
			[]int{1, 2, 3}, mode.ErrOutOfRange, outOfRange},
		{"inadmissible later degree before an unrepresentable earlier one", t1, []string{"C"}, stepList(1<<40, 1),
			[]int{1, 2, 641381782872}, nil, []error{system.ErrInadmissible}},
		{"without names the unrepresentable degree is reported", t1, []string{""}, stepList(1<<40, 1),
			[]int{1, 2, 641381782872}, mode.ErrOutOfRange, outOfRange},
		{"unrepresentable adjacent interval", traditional, both, stepList(16000002, 1), []int{1, 2, 18666669},
			mode.ErrOutOfRange, outOfRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, start := range tt.starts {
				m, err := openFromNumbers(tt.sys, tt.adjacent, start, tt.sizes)
				if tt.category == nil {
					checkForeignCause(t, err, tt.causes...)
				} else {
					checkCategory(t, err, tt.category, tt.causes...)
				}
				assert.Zero(t, m, "start %q", start)
				assert.False(t, m.IsValid())
			}
		})
	}
}
