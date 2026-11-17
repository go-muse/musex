package mode_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEqualCompareAndRotationDirection(t *testing.T) {
	t.Parallel()
	a := mode.MustFromName("ионийский")
	b := mode.MustFromName("натуральный мажор")
	assert.True(t, a.Equal(b))
	assert.Zero(t, a.Compare(b))
	assert.False(t, a.Equal(mode.MustFromCode("aeolian")))
	d := mode.MustFromCode("dorian")
	assert.True(t, d.Equal(mode.MustParse(dorian)))
	assert.Equal(t, 2, firstOf(d.RotationOf(a)))
	assert.True(t, secondOf(d.RotationOf(a)))
	assert.Equal(t, 7, firstOf(a.RotationOf(d)))
	assert.True(t, secondOf(a.RotationOf(d)))
	n, ok := mode.MustFromCode("harmonic-minor").RotationOf(a)
	assert.Zero(t, n)
	assert.False(t, ok)
	repeating := mode.MustParse("M2 m2 M2 m2")
	rotation := mode.MustFromDegree(repeating, 4)
	n, ok = rotation.RotationOf(repeating)
	assert.Equal(t, 2, n)
	assert.True(t, ok)

	var zero mode.Mode
	assert.False(t, zero.Equal(zero))
	assert.False(t, zero.Equal(a))
	assert.False(t, a.Equal(zero))
	assert.Zero(t, zero.Compare(zero))
	assert.Equal(t, -1, zero.Compare(a))
	assert.Equal(t, 1, a.Compare(zero))
	n, ok = a.RotationOf(zero)
	assert.Zero(t, n)
	assert.False(t, ok)

	open := mode.MustParse("m3 M2 |")
	closed := mode.MustParse("m3 M2")
	assert.False(t, open.Equal(closed))
	assert.False(t, secondOf(open.RotationOf(closed)))
	assert.False(t, secondOf(mode.MustParse("M2 m3 |").RotationOf(open)))
	n, ok = open.RotationOf(mode.MustParse("m3 M2 |"))
	assert.True(t, ok)
	assert.Equal(t, 1, n)
	values := []mode.Mode{a, d, zero, mode.MustFromCode("pentatonic-major"), open, closed}
	slices.SortFunc(values, mode.Mode.Compare)
	require.Len(t, values, 6)
	assert.False(t, values[0].IsValid())
	for i, want := range []string{"m3 M2", "M2 M2 m3 M2 m3", dorian, ionian, "m3 M2 |"} {
		checkMode(t, values[i+1], want)
	}
}

func TestDifferenceAndMismatches(t *testing.T) {
	t.Parallel()
	diff := func(number, delta int) mode.DegreeDifference {
		return mode.DegreeDifference{Degree: number, Delta: alteration.MustFromInt(delta)}
	}
	for _, tt := range []struct {
		lhs, rhs string
		want     []mode.DegreeDifference
	}{
		{"dorian", "aeolian", []mode.DegreeDifference{diff(6, 1)}},
		{"lydian", "ionian", []mode.DegreeDifference{diff(4, 1)}},
		{"locrian", "aeolian", []mode.DegreeDifference{diff(2, -1), diff(5, -1)}},
		{"aeolian", "ionian", []mode.DegreeDifference{diff(3, -1), diff(6, -1), diff(7, -1)}},
		{"harmonic-major", "harmonic-minor", []mode.DegreeDifference{diff(3, 1)}},
	} {
		t.Run(tt.lhs+"_vs_"+tt.rhs, func(t *testing.T) {
			lhs, rhs := mode.MustFromCode(tt.lhs), mode.MustFromCode(tt.rhs)
			got, err := lhs.Difference(rhs)
			require.NoError(t, err)
			assert.True(t, got.HasCycle)
			assert.True(t, alteration.MustFromInt(0).Equal(got.Cycle))
			assert.Equal(t, tt.want, got.Degrees)
			mismatches, err := lhs.MismatchingDegrees(rhs, mode.FromFirst)
			require.NoError(t, err)
			require.Len(t, mismatches.Degrees, len(tt.want))
			for i, d := range tt.want {
				assert.Equal(t, d.Degree, mismatches.Degrees[i])
			}
		})
	}
	d, a := mode.MustFromCode("dorian"), mode.MustFromCode("aeolian")
	for _, tt := range []struct {
		m    mode.Measure
		want []int
	}{{mode.FromFirst, []int{6}}, {mode.Adjacent, []int{5, 6}}} {
		result, err := d.MismatchingDegrees(a, tt.m)
		require.NoError(t, err)
		assert.Equal(t, tt.want, result.Degrees)
		assert.Equal(t, len(tt.want), result.Count())
		inverse, err := a.MismatchingDegrees(d, tt.m)
		require.NoError(t, err)
		assert.Equal(t, result, inverse)
		result.Degrees[0] = 0
		assert.Equal(t, tt.want, d.MustMismatchingDegrees(a, tt.m).Degrees)
		equal, err := d.MismatchingDegrees(mode.MustParse(dorian), tt.m)
		require.NoError(t, err)
		assert.NotNil(t, equal.Degrees)
		assert.Zero(t, equal.Count())
	}

	blues, flat := mode.MustFromCode("blues"), mode.MustFromCode("blues-b5")
	got, err := blues.Difference(flat)
	checkError(t, err, mode.ErrIncomparable)
	assert.Zero(t, got)
	assert.Equal(t, []int{4}, blues.MustMismatchingDegrees(flat, mode.FromFirst).Degrees)
	got, err = mode.MustFromCode("pentatonic-major").Difference(mode.MustFromCode("pentatonic-minor"))
	checkError(t, err, mode.ErrIncomparable)
	assert.Zero(t, got)
	equal, err := d.Difference(mode.MustParse(dorian))
	require.NoError(t, err)
	assert.NotNil(t, equal.Degrees)
	assert.Empty(t, equal.Degrees)
	assert.True(t, equal.HasCycle)

	openA, openB := mode.MustParse("m3 M2 |"), mode.MustParse("m3 m2 |")
	got, err = openA.Difference(openB)
	require.NoError(t, err)
	assert.False(t, got.HasCycle)
	assert.Equal(t, []mode.DegreeDifference{diff(3, 1)}, got.Degrees)
	assert.Equal(t, []int{3}, openA.MustMismatchingDegrees(openB, mode.FromFirst).Degrees)
	assert.Equal(t, []int{2}, openA.MustMismatchingDegrees(openB, mode.Adjacent).Degrees)
	closed := mode.MustParse("m3 M2 P5")
	got, err = openA.Difference(closed)
	checkError(t, err, mode.ErrIncomparable)
	assert.Zero(t, got)
	mm, err := openA.MismatchingDegrees(closed, mode.Adjacent)
	checkError(t, err, mode.ErrIncomparable, template.ErrIncomparable)
	assert.Zero(t, mm)
	mm, err = d.MismatchingDegrees(mode.MustParse("P8"), mode.MeasureUnset)
	checkError(t, err, mode.ErrInvalidArgument, template.ErrInvalidArgument)
	assert.Zero(t, mm)

	shorter, longer := mode.MustParse("M2 m2"), mode.MustParse("M2 M2")
	got, err = shorter.Difference(longer)
	require.NoError(t, err)
	assert.True(t, got.HasCycle)
	assert.True(t, alteration.MustFromInt(-1).Equal(got.Cycle))
	assert.NotNil(t, got.Degrees)
	assert.Empty(t, got.Degrees)
	assert.Equal(t, []int{3}, shorter.MustMismatchingDegrees(longer, mode.FromFirst).Degrees)
	assert.Equal(t, []int{2}, shorter.MustMismatchingDegrees(longer, mode.Adjacent).Degrees)

	magnitude := alteration.MaxMagnitude
	high := mode.MustOpenFromIntervals([]interval.Interval{interval.MustFromDeviation(2, alteration.MustFromInt(magnitude))})
	low := mode.MustOpenFromIntervals([]interval.Interval{interval.MustFromDeviation(2, alteration.MustFromInt(-magnitude))})
	got, err = high.Difference(low)
	checkError(t, err, mode.ErrOutOfRange, alteration.ErrOutOfRange)
	assert.Zero(t, got)
}

func TestSameInSystem(t *testing.T) {
	t.Parallel()
	blues, flat := mode.MustFromCode("blues"), mode.MustFromCode("blues-b5")
	assert.True(t, blues.MustSameInSystem(system.Traditional(), flat))
	assert.False(t, blues.MustSameInSystem(system.Positions19(), flat))
	open, closed := mode.MustParse("m3 M2 |"), mode.MustParse("m3 M2")
	assert.False(t, open.MustSameInSystem(system.Traditional(), closed))
	assert.False(t, closed.MustSameInSystem(system.Traditional(), mode.MustParse("P8")))

	m := mode.MustFromCode("ionian")
	c, d := spelling.MustParse("C"), spelling.MustParse("D")
	irregular := irregularSystem()
	assert.False(t, m.MustSameInSystemFrom(irregular, c, m, d))
	assert.True(t, m.MustSameInSystemFrom(irregular, c, m, c))
	got, err := open.SameInSystem(irregular, closed)
	assert.ErrorIs(t, err, interval.ErrStartRequired)
	assert.False(t, got)
	got, err = open.SameInSystemFrom(irregular, c, closed, spelling.Spelling{})
	checkError(t, err, mode.ErrUnset)
	assert.False(t, got)
}
