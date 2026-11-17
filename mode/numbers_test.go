package mode_test

import (
	"math"
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

func TestNumbersInRegularSystems(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, code        string
		sys               *system.System
		degrees, adjacent []system.Step
		cycle             system.Step
		symmetry          mode.Symmetry
	}{
		{"traditional major", "ionian", system.Traditional(), []system.Step{0, 2, 4, 5, 7, 9, 11}, []system.Step{2, 2, 1, 2, 2, 2, 1}, 12, mode.Symmetry{Rotations: 7, Transpositions: 12}},
		{"19-position major", "ionian", system.Positions19(), []system.Step{0, 3, 6, 8, 11, 14, 17}, []system.Step{3, 3, 2, 3, 3, 3, 2}, 19, mode.Symmetry{Rotations: 7, Transpositions: 19}},
		{"31-position major", "ionian", system.Fokker31(), []system.Step{0, 5, 10, 13, 18, 23, 28}, []system.Step{5, 5, 3, 5, 5, 5, 3}, 31, mode.Symmetry{Rotations: 7, Transpositions: 31}},
		{"traditional whole tone", "whole-tone", system.Traditional(), []system.Step{0, 2, 4, 6, 8, 10}, []system.Step{2, 2, 2, 2, 2, 2}, 12, mode.Symmetry{Rotations: 1, Transpositions: 2, Limited: true}},
		{"19-position whole tone", "whole-tone", system.Positions19(), []system.Step{0, 3, 6, 9, 12, 15}, []system.Step{3, 3, 3, 3, 3, 4}, 19, mode.Symmetry{Rotations: 6, Transpositions: 19}},
		{"half whole", "octatonic-hw", system.Traditional(), []system.Step{0, 1, 3, 4, 6, 7, 9, 10}, []system.Step{1, 2, 1, 2, 1, 2, 1, 2}, 12, mode.Symmetry{Rotations: 2, Transpositions: 3, Limited: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mode.MustFromCode(tt.code)
			got, err := m.Numbers(tt.sys)
			require.NoError(t, err)
			want := mode.Positions{Degrees: tt.degrees, Adjacent: tt.adjacent, Cycle: tt.cycle, HasCycle: true}
			assert.Equal(t, want, got)
			got.Degrees[0], got.Adjacent[0] = 99, 99
			assert.Equal(t, want, m.MustNumbers(tt.sys))
			assert.Equal(t, mode.CycleInfo{Cycles: 1, Octaves: 1}, m.MustCycleIn(tt.sys))
			assert.Equal(t, tt.symmetry, m.MustSymmetryIn(tt.sys))
		})
	}
	short := mode.MustParse("M2 M2 m2")
	sys := system.Traditional()
	assert.Equal(t, mode.Positions{Degrees: []system.Step{0, 2, 4}, Adjacent: []system.Step{2, 2, 1}, Cycle: 5, HasCycle: true}, short.MustNumbers(sys))
	assert.Equal(t, mode.CycleInfo{Comparison: -1, Cycles: 12, Octaves: 5}, short.MustCycleIn(sys))
	wide := mode.MustParse("M9 m7")
	assert.Equal(t, mode.Positions{Degrees: []system.Step{0, 14}, Adjacent: []system.Step{14, 10}, Cycle: 24, HasCycle: true}, wide.MustNumbers(sys))
	assert.Equal(t, mode.CycleInfo{Comparison: 1, Cycles: 1, Octaves: 2}, wide.MustCycleIn(sys))
	gcd := mode.MustParse("P5 P4 P5")
	assert.Equal(t, mode.CycleInfo{Comparison: 1, Cycles: 12, Octaves: 19}, gcd.MustCycleIn(sys))
	assert.Equal(t, mode.CycleInfo{Comparison: -1, Cycles: 6, Octaves: 5}, mode.MustParse("P5 m3").MustCycleIn(sys))
	open := mode.MustParse("m3 M2 |")
	assert.Equal(t, mode.Positions{Degrees: []system.Step{0, 3, 5}, Adjacent: []system.Step{3, 2}}, open.MustNumbers(sys))
	assert.Equal(t, mode.Positions{Degrees: []system.Step{0}, Adjacent: []system.Step{0}, HasCycle: true}, mode.MustParse("d2").MustNumbers(sys))
}

func TestNumbersUseEachSuccessiveName(t *testing.T) {
	t.Parallel()
	sys := irregularSystem()
	m := mode.MustFromCode("ionian")
	d := spelling.MustParse("D")
	got, err := m.NumbersFrom(sys, d)
	require.NoError(t, err)
	assert.Equal(t, mode.Positions{Degrees: []system.Step{0, 1, 3, 4, 6, 8, 10}, Adjacent: []system.Step{1, 2, 1, 2, 2, 2, 2}, Cycle: 12, HasCycle: true}, got)
	assert.Equal(t, got, m.MustNumbersFrom(sys, spelling.MustParse("D+1")))
	assert.Equal(t, mode.CycleInfo{Cycles: 1, Octaves: 1}, m.MustCycleInFrom(sys, d))
	assert.Equal(t, mode.Symmetry{Rotations: 7, Transpositions: 12}, m.MustSymmetryInFrom(sys, d))
	got, err = m.Numbers(sys)
	assert.ErrorIs(t, err, interval.ErrStartRequired)
	assert.Zero(t, got)
	cycle, err := m.CycleIn(sys)
	assert.ErrorIs(t, err, interval.ErrStartRequired)
	assert.Zero(t, cycle)
	sym, err := m.SymmetryIn(sys)
	assert.ErrorIs(t, err, interval.ErrStartRequired)
	assert.Zero(t, sym)

	naturalOnly := system.MustNewRegular(2, 1, letter.C, system.MustIntegerRange(0, 0), "Naturals only")
	// C and D are admissible; the name of the closing degree, E flat, is not.
	closesAtFlat := mode.MustParse("M2 m2")
	assert.Equal(t, system.Step(3), closesAtFlat.MustNumbers(naturalOnly).Cycle)
	got, err = closesAtFlat.NumbersFrom(naturalOnly, spelling.MustParse("C"))
	assert.ErrorIs(t, err, system.ErrInadmissible)
	assert.Zero(t, got)
	created, err := mode.FromNumbersFromWithPlan(naturalOnly, []system.Step{2, 1}, spelling.MustParse("C"), mode.LetterPlan{Degrees: []int{1, 2}, Cycle: 3})
	assert.ErrorIs(t, err, system.ErrInadmissible)
	assert.Zero(t, created)

	got, err = m.NumbersFrom(sys, spelling.Spelling{})
	checkError(t, err, mode.ErrUnset)
	assert.Zero(t, got)
	got, err = m.Numbers(nil)
	checkError(t, err, mode.ErrUnset)
	assert.Zero(t, got)
}

func TestNumericErrorsAndPrecedence(t *testing.T) {
	t.Parallel()
	sys := system.Traditional()
	nonpositive := mode.MustParse("d2 A7")
	assert.Equal(t, []system.Step{0, 12}, nonpositive.MustNumbers(sys).Adjacent)
	negativeEdge := mode.MustClosedFromIntervals([]interval.Interval{
		interval.MustFromDeviation(2, alteration.MustFromInt(-3)), interval.MustParse("AA7"),
	})
	assert.Equal(t, []system.Step{-1, 13}, negativeEdge.MustNumbers(sys).Adjacent)
	for _, tt := range []struct {
		name string
		m    mode.Mode
		err  error
	}{
		{"zero step in octave", nonpositive, mode.ErrNonPositiveNumber},
		{"negative step in octave", negativeEdge, mode.ErrNonPositiveNumber},
		{"cycle checked before nonpositive step", mode.MustParse("d2 M2"), mode.ErrSystemOctaveRequired},
		{"short cycle", mode.MustParse("M2 M2 m2"), mode.ErrSystemOctaveRequired},
		{"fractional before octave check", mode.MustParse("n2 M2"), interval.ErrFractionalSteps},
		{"open", mode.MustParse("M2 |"), mode.ErrOpenMode},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sym, err := tt.m.SymmetryIn(sys)
			assert.ErrorIs(t, err, tt.err)
			assert.Zero(t, sym)
			sym, err = tt.m.SymmetryInFrom(sys, spelling.MustParse("C"))
			if tt.err != interval.ErrFractionalSteps {
				assert.ErrorIs(t, err, tt.err)
				assert.Zero(t, sym)
			} else {
				// Explicit names can additionally fail admissibility; either
				// way the fractional request cannot report numeric symmetry.
				assert.Error(t, err)
				assert.Zero(t, sym)
			}
		})
	}
	zero, err := mode.MustParse("d2").CycleIn(sys)
	checkError(t, err, mode.ErrNonPositiveNumber)
	assert.Zero(t, zero)
	zero, err = mode.MustParse("M2 |").CycleIn(sys)
	checkError(t, err, mode.ErrOpenMode)
	assert.Zero(t, zero)
	zero, err = mode.MustParse("M2 |").CycleInFrom(sys, spelling.Spelling{})
	checkError(t, err, mode.ErrUnset)
	assert.Zero(t, zero)
	wide := mode.MustClosedFromIntervals([]interval.Interval{interval.MustFromDeviation(math.MaxInt, alteration.MustFromInt(0))})
	got, err := wide.Numbers(sys)
	checkError(t, err, mode.ErrOutOfRange, interval.ErrOutOfRange)
	assert.Zero(t, got)
}
