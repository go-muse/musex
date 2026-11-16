package degree_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectedShift struct {
	number int
	cycles int
}

func shiftByContract(k, j, n int) expectedShift {
	q := n / k
	r := n % k
	if r < 0 {
		q--
		r += k
	}
	a := j - 1
	cycles := q
	remainder := a + r
	if r > 0 && a >= k-r {
		remainder = a - (k - r)
		cycles++
	}
	return expectedShift{number: remainder + 1, cycles: cycles}
}

func TestDegreeNextAndPreviousControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		k, j            int
		next            int
		nextWrapped     bool
		previous        int
		previousWrapped bool
	}{
		{"first of seven", 7, 1, 2, false, 7, true},
		{"second of seven", 7, 2, 3, false, 1, false},
		{"last of seven", 7, 7, 1, true, 6, false},
		{"single number", 1, 1, 1, true, 1, true},
		{"last of two", 2, 2, 1, true, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := degree.MustNew(tt.j)
			next, wrapped, err := d.Next(tt.k)
			require.NoError(t, err)
			assertDegreeState(t, next, tt.next, false, note.Note{})
			assert.Equal(t, tt.nextWrapped, wrapped)
			mustNext, mustWrapped := d.MustNext(tt.k)
			assertSameDegree(t, next, mustNext)
			assert.Equal(t, wrapped, mustWrapped)

			previous, wrapped, err := d.Previous(tt.k)
			require.NoError(t, err)
			assertDegreeState(t, previous, tt.previous, false, note.Note{})
			assert.Equal(t, tt.previousWrapped, wrapped)
			mustPrevious, mustWrapped := d.MustPrevious(tt.k)
			assertSameDegree(t, previous, mustPrevious)
			assert.Equal(t, wrapped, mustWrapped)
		})
	}
}

func TestDegreeShiftControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		k, j, n  int
		expected expectedShift
	}{
		{7, 1, 1, expectedShift{2, 0}},
		{7, 7, 1, expectedShift{1, 1}},
		{7, 1, -1, expectedShift{7, -1}},
		{7, 1, 0, expectedShift{1, 0}},
		{7, 3, 7, expectedShift{3, 1}},
		{7, 3, -7, expectedShift{3, -1}},
		{7, 5, 9, expectedShift{7, 1}},
		{7, 5, -9, expectedShift{3, -1}},
		{7, 2, 14, expectedShift{2, 2}},
		{7, 6, 3, expectedShift{2, 1}},
		{7, 2, -3, expectedShift{6, -1}},
		{2, 1, 1, expectedShift{2, 0}},
		{2, 2, 1, expectedShift{1, 1}},
		{2, 1, 3, expectedShift{2, 1}},
		{2, 1, -1, expectedShift{2, -1}},
		{1, 1, 5, expectedShift{1, 5}},
		{1, 1, -2, expectedShift{1, -2}},
		{5, 4, 3, expectedShift{2, 1}},
		{8, 8, 1, expectedShift{1, 1}},
		{8, 3, -4, expectedShift{7, -1}},
		{12, 11, 5, expectedShift{4, 1}},
		{7, 1, math.MaxInt, expectedShift{1, 1317624576693539401}},
		{7, 1, math.MinInt, expectedShift{7, -1317624576693539402}},
		{2, 2, math.MaxInt, expectedShift{1, 4611686018427387904}},
		{1, 1, math.MinInt, expectedShift{1, math.MinInt}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("k_%d_j_%d_n_%d", tt.k, tt.j, tt.n), func(t *testing.T) {
			t.Parallel()
			d := degree.MustNew(tt.j)
			shifted, cycles, err := d.Shift(tt.k, tt.n)
			require.NoError(t, err)
			assertDegreeState(t, shifted, tt.expected.number, false, note.Note{})
			assert.Equal(t, tt.expected.cycles, cycles)
			mustShifted, mustCycles := d.MustShift(tt.k, tt.n)
			assertSameDegree(t, shifted, mustShifted)
			assert.Equal(t, cycles, mustCycles)
			assertDegreeState(t, d, tt.j, false, note.Note{})
		})
	}
}

func TestDegreeShiftIgnoresAndDropsInputNote(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	d := degree.MustNewWithNote(7, n)
	shifted, cycles, err := d.Shift(7, 1)
	require.NoError(t, err)
	assertDegreeState(t, shifted, 1, false, note.Note{})
	assert.Equal(t, 1, cycles)
	assertDegreeState(t, d, 7, true, n)

	normalized, normalizedCycles, err := d.Normalize(3)
	require.NoError(t, err)
	assertDegreeState(t, normalized, 1, false, note.Note{})
	assert.Equal(t, 2, normalizedCycles)
	within := degree.MustNewWithNote(2, n)
	denormalized, err := within.Denormalize(3, 1)
	require.NoError(t, err)
	assertDegreeState(t, denormalized, 5, false, note.Note{})
	assertDegreeState(t, within, 2, true, n)
}

func TestDegreeShiftLargeIntegerCarryBranches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		k, j, n int
	}{
		{"positive without carry", 7, 1, 1},
		{"positive with carry", 7, 7, 1},
		{"negative without additional cycle", 7, 2, -1},
		{"negative with borrow", 7, 1, -1},
		{"large k positive carry", math.MaxInt, math.MaxInt, 2},
		{"large k positive near remainder boundary", math.MaxInt, 2, math.MaxInt - 1},
		{"large k negative borrow", math.MaxInt, 1, -1},
		{"large k maximum n", math.MaxInt, math.MaxInt, math.MaxInt},
		{"large k minimum n", math.MaxInt, math.MaxInt, math.MinInt},
		{"large k minimum n from first", math.MaxInt, 1, math.MinInt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expected := shiftByContract(tt.k, tt.j, tt.n)
			shifted, cycles, err := degree.MustNew(tt.j).Shift(tt.k, tt.n)
			require.NoError(t, err)
			assertDegreeState(t, shifted, expected.number, false, note.Note{})
			assert.Equal(t, expected.cycles, cycles)
		})
	}
}

func TestDegreeInCircle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		k, j int
		want bool
	}{
		{7, 1, true},
		{7, 7, true},
		{7, 8, false},
		{1, 1, true},
		{1, 2, false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("k_%d_j_%d", tt.k, tt.j), func(t *testing.T) {
			t.Parallel()
			d := degree.MustNew(tt.j)
			got, err := d.InCircle(tt.k)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, got, d.MustInCircle(tt.k))
			withNote := degree.MustNewWithNote(tt.j, degreeTestNote(t, spelling.B(), cents.Zero(), nil))
			withNoteResult, err := withNote.InCircle(tt.k)
			require.NoError(t, err)
			assert.Equal(t, got, withNoteResult)
		})
	}
}

func TestDegreeDistancesControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		k, a, b  int
		up, down int
		shortest int
	}{
		{7, 1, 5, 4, 3, -3},
		{7, 5, 1, 3, 4, 3},
		{7, 7, 1, 1, 6, 1},
		{7, 1, 7, 6, 1, -1},
		{7, 1, 4, 3, 4, 3},
		{7, 3, 3, 0, 0, 0},
		{2, 1, 2, 1, 1, 1},
		{2, 2, 1, 1, 1, 1},
		{8, 2, 8, 6, 2, -2},
		{8, 1, 5, 4, 4, 4},
		{8, 5, 1, 4, 4, 4},
		{1, 1, 1, 0, 0, 0},
		{12, 1, 7, 6, 6, 6},
		{12, 1, 8, 7, 5, -5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("k_%d_a_%d_b_%d", tt.k, tt.a, tt.b), func(t *testing.T) {
			t.Parallel()
			a := degree.MustNew(tt.a)
			b := degree.MustNew(tt.b)
			up, err := a.DistanceUp(b, tt.k)
			require.NoError(t, err)
			assert.Equal(t, tt.up, up)
			assert.Equal(t, up, a.MustDistanceUp(b, tt.k))
			down, err := a.DistanceDown(b, tt.k)
			require.NoError(t, err)
			assert.Equal(t, tt.down, down)
			assert.Equal(t, down, a.MustDistanceDown(b, tt.k))
			shortest, err := a.ShortestDistance(b, tt.k)
			require.NoError(t, err)
			assert.Equal(t, tt.shortest, shortest)
			assert.Equal(t, shortest, a.MustShortestDistance(b, tt.k))
			shifted, _, err := a.Shift(tt.k, shortest)
			require.NoError(t, err)
			assertSameDegree(t, b, shifted)
		})
	}
}

func TestDegreeNormalizeAndDenormalizeControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		m, k   int
		number int
		cycles int
	}{
		{8, 7, 1, 1},
		{7, 7, 7, 0},
		{1, 7, 1, 0},
		{9, 7, 2, 1},
		{14, 7, 7, 1},
		{15, 7, 1, 2},
		{6, 5, 1, 1},
		{5, 1, 1, 4},
		{3, 2, 1, 1},
		{8, 8, 8, 0},
		{math.MaxInt, 7, 7, 1317624576693539400},
		{math.MaxInt, 1, 1, math.MaxInt - 1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("m_%d_k_%d", tt.m, tt.k), func(t *testing.T) {
			t.Parallel()
			d := degree.MustNew(tt.m)
			normalized, cycles, err := d.Normalize(tt.k)
			require.NoError(t, err)
			assertDegreeState(t, normalized, tt.number, false, note.Note{})
			assert.Equal(t, tt.cycles, cycles)
			mustNormalized, mustCycles := d.MustNormalize(tt.k)
			assertSameDegree(t, normalized, mustNormalized)
			assert.Equal(t, cycles, mustCycles)

			denormalized, err := normalized.Denormalize(tt.k, cycles)
			require.NoError(t, err)
			assertDegreeState(t, denormalized, tt.m, false, note.Note{})
			mustDenormalized := normalized.MustDenormalize(tt.k, cycles)
			assertSameDegree(t, denormalized, mustDenormalized)
			assertDegreeState(t, d, tt.m, false, note.Note{})
		})
	}
}

func TestDegreeDenormalizeLastRepresentableResult(t *testing.T) {
	t.Parallel()
	d := degree.MustNew(7)
	const cycles = 1317624576693539400
	result, err := d.Denormalize(7, cycles)
	require.NoError(t, err)
	assertDegreeState(t, result, math.MaxInt, false, note.Note{})
	normalized, normalizedCycles, err := result.Normalize(7)
	require.NoError(t, err)
	assertSameDegree(t, d, normalized)
	assert.Equal(t, cycles, normalizedCycles)
}

func TestDegreeRenumberingByDistance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		first int
		want  []int
	}{
		{2, []int{7, 1, 2, 3, 4, 5, 6}},
		{5, []int{4, 5, 6, 7, 1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("first_%d", tt.first), func(t *testing.T) {
			t.Parallel()
			first := degree.MustNew(tt.first)
			for j, expected := range tt.want {
				d := degree.MustNew(j + 1)
				distance, err := first.DistanceUp(d, 7)
				require.NoError(t, err)
				assert.Equal(t, expected, distance+1)
			}
		})
	}
}

func TestDegreeCircleExhaustiveProperties(t *testing.T) {
	t.Parallel()
	n := degreeTestNote(t, spelling.B(), cents.Zero(), nil)
	for k := 1; k <= 13; k++ {
		t.Run(fmt.Sprintf("k_%d", k), func(t *testing.T) {
			t.Parallel()
			for j := 1; j <= k; j++ {
				d := degree.MustNew(j)
				withNote := degree.MustNewWithNote(j, n)

				next, wrapped, err := d.Next(k)
				require.NoError(t, err)
				assertDegreeState(t, next, j%k+1, false, note.Note{})
				assert.Equal(t, j == k, wrapped)
				nextWithNote, wrappedWithNote, err := withNote.Next(k)
				require.NoError(t, err)
				assertSameDegree(t, next, nextWithNote)
				assert.Equal(t, wrapped, wrappedWithNote)

				previous, wrapped, err := d.Previous(k)
				require.NoError(t, err)
				expectedPrevious := j - 1
				if expectedPrevious == 0 {
					expectedPrevious = k
				}
				assertDegreeState(t, previous, expectedPrevious, false, note.Note{})
				assert.Equal(t, j == 1, wrapped)
				previousWithNote, wrappedWithNote, err := withNote.Previous(k)
				require.NoError(t, err)
				assertSameDegree(t, previous, previousWithNote)
				assert.Equal(t, wrapped, wrappedWithNote)

				inside, err := d.InCircle(k)
				require.NoError(t, err)
				assert.True(t, inside)

				for delta := -3 * k; delta <= 3*k; delta++ {
					expected := shiftByContract(k, j, delta)
					shifted, cycles, err := d.Shift(k, delta)
					require.NoError(t, err)
					assertDegreeState(t, shifted, expected.number, false, note.Note{})
					assert.Equal(t, expected.cycles, cycles)
					shiftedWithNote, cyclesWithNote, err := withNote.Shift(k, delta)
					require.NoError(t, err)
					assertSameDegree(t, shifted, shiftedWithNote)
					assert.Equal(t, cycles, cyclesWithNote)

					if delta != math.MinInt {
						returned, returnCycles, err := shifted.Shift(k, -delta)
						require.NoError(t, err)
						assertSameDegree(t, d, returned)
						assert.Zero(t, cycles+returnCycles)
					}
				}

				for target := 1; target <= k; target++ {
					to := degree.MustNew(target)
					up, err := d.DistanceUp(to, k)
					require.NoError(t, err)
					expectedUp := (target - j + k) % k
					assert.Equal(t, expectedUp, up)
					assert.GreaterOrEqual(t, up, 0)
					assert.Less(t, up, k)
					down, err := d.DistanceDown(to, k)
					require.NoError(t, err)
					expectedDown := (j - target + k) % k
					assert.Equal(t, expectedDown, down)
					assert.GreaterOrEqual(t, down, 0)
					assert.Less(t, down, k)
					if up == 0 {
						assert.Zero(t, down)
					} else {
						assert.Equal(t, k, up+down)
					}

					shortest, err := d.ShortestDistance(to, k)
					require.NoError(t, err)
					expectedShortest := up
					if up > down {
						expectedShortest = -down
					}
					assert.Equal(t, expectedShortest, shortest)
					shifted, _, err := d.Shift(k, shortest)
					require.NoError(t, err)
					assertSameDegree(t, to, shifted)

					upWithNote, err := withNote.DistanceUp(degree.MustNewWithNote(target, n), k)
					require.NoError(t, err)
					assert.Equal(t, up, upWithNote)
				}
			}

			for m := 1; m <= 5*k; m++ {
				d := degree.MustNew(m)
				normalized, cycles, err := d.Normalize(k)
				require.NoError(t, err)
				assertDegreeState(t, normalized, (m-1)%k+1, false, note.Note{})
				assert.Equal(t, (m-1)/k, cycles)
				denormalized, err := normalized.Denormalize(k, cycles)
				require.NoError(t, err)
				assertSameDegree(t, d, denormalized)
			}
		})
	}
}

func TestDegreeAllSmallCirclesAndEarlyTermination(t *testing.T) {
	t.Parallel()
	for _, k := range []int{1, 2, 5} {
		t.Run(fmt.Sprintf("k_%d", k), func(t *testing.T) {
			t.Parallel()
			sequence, err := degree.All(k)
			require.NoError(t, err)
			require.NotNil(t, sequence)
			collect := func(limit int) []degree.Degree {
				var values []degree.Degree
				for d := range sequence {
					values = append(values, d)
					if len(values) == limit {
						break
					}
				}
				return values
			}
			values := collect(k)
			require.Len(t, values, k)
			for i, d := range values {
				assertDegreeState(t, d, i+1, false, note.Note{})
			}
			repeated := collect(k)
			require.Len(t, repeated, k)
			for i := range values {
				assertSameDegree(t, values[i], repeated[i])
			}

			early := collect(1)
			require.Len(t, early, 1)
			assertDegreeState(t, early[0], 1, false, note.Note{})
			mustSequence := degree.MustAll(k)
			for d := range mustSequence {
				assertSameDegree(t, degree.First(), d)
				break
			}
		})
	}
}

func TestDegreeAllMaxNumberIsLazyAtBeginning(t *testing.T) {
	t.Parallel()
	sequence, err := degree.All(degree.MaxNumber)
	require.NoError(t, err)
	require.NotNil(t, sequence)
	const prefix = 3
	count := 0
	for d := range sequence {
		count++
		assertDegreeState(t, d, count, false, note.Note{})
		if count == prefix {
			break
		}
	}
	assert.Equal(t, prefix, count)
	// This deliberately checks only the lazy prefix. The terminal MaxInt transition
	// is a separate static implementation-review obligation, not a dynamic claim here.
}
