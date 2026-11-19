package modal_test

import (
	"cmp"
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func position(number, diminished, minor, perfect, major, augmented, score, rank int) modal.AbsolutePosition {
	return modal.AbsolutePosition{Number: number, Counts: modal.QualityCounts{Diminished: diminished, Minor: minor, Perfect: perfect, Major: major, Augmented: augmented}, Score: score, Rank: rank}
}

func group(rank, score int, numbers ...int) modal.AbsoluteGroup {
	return modal.AbsoluteGroup{Rank: rank, Score: score, Numbers: numbers}
}

func dorianAnalysis() modal.AbsoluteAnalysis {
	return modal.AbsoluteAnalysis{
		Positions: []modal.AbsolutePosition{
			position(1, 0, 2, 3, 2, 0, 0, 4), position(2, 0, 4, 3, 0, 0, 4, 2), position(3, 0, 0, 2, 4, 1, -5, 7),
			position(4, 0, 1, 3, 3, 0, -2, 5), position(5, 0, 3, 3, 1, 0, 2, 3), position(6, 1, 4, 2, 0, 0, 5, 1), position(7, 0, 0, 3, 4, 0, -4, 6),
		},
		Order: []modal.AbsoluteGroup{group(1, 5, 6), group(2, 4, 2), group(3, 2, 5), group(4, 0, 1), group(5, -2, 4), group(6, -4, 7), group(7, -5, 3)},
	}
}

// Check every public absolute entry point against independent fixed values.
func checkAbsolute(t *testing.T, m mode.Mode, want modal.AbsoluteAnalysis) {
	t.Helper()
	got, err := modal.AnalyzeAbsolute(m)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, want, modal.MustAnalyzeAbsolute(m))
	positions, err := modal.AbsolutePositions(m)
	require.NoError(t, err)
	assert.Equal(t, want.Positions, positions)
	assert.Equal(t, want.Positions, modal.MustAbsolutePositions(m))
	order, err := modal.Order(m)
	require.NoError(t, err)
	assert.Equal(t, want.Order, order)
	assert.Equal(t, want.Order, modal.MustOrder(m))
	for _, first := range want.Positions {
		p, err := modal.Absolute(m, first.Number)
		require.NoError(t, err)
		assert.Equal(t, first, p)
		assert.Equal(t, first, modal.MustAbsolute(m, first.Number))
		counts := p.Counts
		assert.Equal(t, m.Len(), counts.Diminished+counts.Minor+counts.Perfect+counts.Major+counts.Augmented)
		for _, second := range want.Positions {
			comparison, err := modal.CompareAbsolute(m, first.Number, second.Number)
			require.NoError(t, err)
			expected := cmp.Compare(first.Rank, second.Rank)
			assert.Equal(t, expected, comparison, "degrees %d, %d", first.Number, second.Number)
			assert.Equal(t, expected, modal.MustCompareAbsolute(m, first.Number, second.Number))
		}
	}
}

// M-3...M-5, MOD-3/4: fixed tables from the business examples. Additional
// counts follow directly from their written ascending specific intervals.
func TestAbsoluteControlExamples(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		degrees string
		want    modal.AbsoluteAnalysis
	}{
		{"dorian", "P1 M2 m3 P4 P5 M6 m7", dorianAnalysis()},
		{"white keys", "P1 M2 M3 P4 P5 M6 M7", modal.AbsoluteAnalysis{
			Positions: []modal.AbsolutePosition{position(1, 0, 0, 3, 4, 0, -4, 6), position(2, 0, 2, 3, 2, 0, 0, 4), position(3, 0, 4, 3, 0, 0, 4, 2), position(4, 0, 0, 2, 4, 1, -5, 7), position(5, 0, 1, 3, 3, 0, -2, 5), position(6, 0, 3, 3, 1, 0, 2, 3), position(7, 1, 4, 2, 0, 0, 5, 1)},
			Order:     []modal.AbsoluteGroup{group(1, 5, 7), group(2, 4, 3), group(3, 2, 6), group(4, 0, 2), group(5, -2, 5), group(6, -4, 1), group(7, -5, 4)},
		}},
		{"harmonic equality", "P1 M2 m3 P4 P5 m6 M7", modal.AbsoluteAnalysis{
			Positions: []modal.AbsolutePosition{position(1, 0, 2, 3, 2, 0, 0, 4), position(2, 1, 3, 2, 1, 0, 3, 2), position(3, 0, 0, 2, 4, 1, -5, 6), position(4, 0, 2, 2, 2, 1, -1, 5), position(5, 0, 3, 3, 1, 0, 2, 3), position(6, 0, 0, 2, 3, 2, -5, 6), position(7, 3, 3, 1, 0, 0, 6, 1)},
			Order:     []modal.AbsoluteGroup{group(1, 6, 7), group(2, 3, 2), group(3, 2, 5), group(4, 0, 1), group(5, -1, 4), group(6, -5, 3, 6)},
		}},
		{"pentatonic", "P1 M2 M3 P5 M6", modal.AbsoluteAnalysis{
			Positions: []modal.AbsolutePosition{position(1, 0, 0, 2, 3, 0, -3, 5), position(2, 0, 1, 3, 1, 0, 0, 3), position(3, 0, 3, 2, 0, 0, 3, 1), position(4, 0, 0, 3, 2, 0, -2, 4), position(5, 0, 2, 3, 0, 0, 2, 2)},
			Order:     []modal.AbsoluteGroup{group(1, 3, 3), group(2, 2, 5), group(3, 0, 2), group(4, -2, 4), group(5, -3, 1)},
		}},
		{"whole tone sharp", "P1 M2 M3 A4 A5 A6", modal.AbsoluteAnalysis{
			Positions: []modal.AbsolutePosition{position(1, 0, 0, 1, 2, 3, -5, 6), position(2, 0, 1, 1, 2, 2, -3, 5), position(3, 0, 2, 1, 2, 1, -1, 4), position(4, 1, 2, 1, 2, 0, 1, 3), position(5, 2, 2, 1, 1, 0, 3, 2), position(6, 3, 2, 1, 0, 0, 5, 1)},
			Order:     []modal.AbsoluteGroup{group(1, 5, 6), group(2, 3, 5), group(3, 1, 4), group(4, -1, 3), group(5, -3, 2), group(6, -5, 1)},
		}},
		{"whole tone flat", "P1 M2 M3 A4 A5 m7", modal.AbsoluteAnalysis{
			Positions: []modal.AbsolutePosition{position(1, 0, 1, 1, 2, 2, -3, 5), position(2, 0, 2, 1, 2, 1, -1, 4), position(3, 1, 2, 1, 2, 0, 1, 3), position(4, 2, 2, 1, 1, 0, 3, 2), position(5, 3, 2, 1, 0, 0, 5, 1), position(6, 0, 0, 1, 2, 3, -5, 6)},
			Order:     []modal.AbsoluteGroup{group(1, 5, 5), group(2, 3, 4), group(3, 1, 3), group(4, -1, 2), group(5, -3, 1), group(6, -5, 6)},
		}},
		{"single P1 only", "P1", modal.AbsoluteAnalysis{Positions: []modal.AbsolutePosition{position(1, 0, 0, 1, 0, 0, 0, 1)}, Order: []modal.AbsoluteGroup{group(1, 0, 1)}}},
		// Eight written places C, C-sharp, ..., C with seven sharps. From
		// each place later targets are augmented unisons; earlier targets
		// in the next circle are diminished octaves. There is no seven-
		// degree limit and each place contributes exactly once.
		{"eight degrees", "P1 A1 AA1 AAA1 AAAA1 AAAAA1 AAAAAA1 AAAAAAA1", modal.AbsoluteAnalysis{
			Positions: []modal.AbsolutePosition{position(1, 0, 0, 1, 0, 7, -7, 8), position(2, 1, 0, 1, 0, 6, -5, 7), position(3, 2, 0, 1, 0, 5, -3, 6), position(4, 3, 0, 1, 0, 4, -1, 5), position(5, 4, 0, 1, 0, 3, 1, 4), position(6, 5, 0, 1, 0, 2, 3, 3), position(7, 6, 0, 1, 0, 1, 5, 2), position(8, 7, 0, 1, 0, 0, 7, 1)},
			Order:     []modal.AbsoluteGroup{group(1, 7, 8), group(2, 5, 7), group(3, 3, 6), group(4, 1, 5), group(5, -1, 4), group(6, -3, 3), group(7, -5, 2), group(8, -7, 1)},
		}},
		// C, D-sharp and E-flat have three distinct degree numbers. The last
		// two have equal traditional sounding heights, but d2/A7 are counted.
		{"equal sounding distinct degrees", "P1 A2 m3", modal.AbsoluteAnalysis{Positions: []modal.AbsolutePosition{position(1, 0, 1, 1, 0, 1, 0, 2), position(2, 2, 0, 1, 0, 0, 2, 1), position(3, 0, 0, 1, 1, 1, -2, 3)}, Order: []modal.AbsoluteGroup{group(1, 2, 2), group(2, 0, 1), group(3, -2, 3)}}},
		// Both P1 and the different numbered P5 are perfect. A second own P8
		// would incorrectly increase both Perfect counts from two to three.
		{"all perfect tie", "P1 P5", modal.AbsoluteAnalysis{Positions: []modal.AbsolutePosition{position(1, 0, 0, 2, 0, 0, 0, 1), position(2, 0, 0, 2, 0, 0, 0, 1)}, Order: []modal.AbsoluteGroup{group(1, 0, 1, 2)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkAbsolute(t, closedMode(t, tc.degrees, "P8"), tc.want)
		})
	}
}

func TestAbsoluteMultiplicityHasUnitContribution(t *testing.T) {
	t.Parallel()
	for _, quality := range []interval.Quality{interval.Augmented, interval.Diminished} {
		for _, multiplicity := range []int{1, 2, 3, 17, alteration.MaxMagnitude} {
			t.Run(fmt.Sprintf("%s_%d", quality, multiplicity), func(t *testing.T) {
				t.Parallel()
				value, err := interval.NewMultiple(multiplicity, quality, 4)
				require.NoError(t, err)
				m, err := mode.ClosedFromDegreeIntervals([]interval.Interval{interval.PerfectUnison(), value}, interval.PerfectOctave())
				require.NoError(t, err)
				// A(n)4 + d(n)5 = P8. Counts and Score depend on the
				// number of intervals, even at the representable boundary.
				if quality == interval.Augmented {
					checkAbsolute(t, m, modal.AbsoluteAnalysis{Positions: []modal.AbsolutePosition{position(1, 0, 0, 1, 0, 1, -1, 2), position(2, 1, 0, 1, 0, 0, 1, 1)}, Order: []modal.AbsoluteGroup{group(1, 1, 2), group(2, -1, 1)}})
				} else {
					checkAbsolute(t, m, modal.AbsoluteAnalysis{Positions: []modal.AbsolutePosition{position(1, 1, 0, 1, 0, 0, 1, 1), position(2, 0, 0, 1, 0, 1, -1, 2)}, Order: []modal.AbsoluteGroup{group(1, 1, 1), group(2, -1, 2)}})
				}
			})
		}
	}
}

func TestRotationPreservesAbsoluteAndRecomputesRelative(t *testing.T) {
	t.Parallel()
	source := dorianMode(t)
	want := dorianAnalysis()
	for first := 1; first <= 7; first++ {
		t.Run(fmt.Sprint(first), func(t *testing.T) {
			t.Parallel()
			rotated, err := mode.FromDegree(source, first)
			require.NoError(t, err)
			got, err := modal.AnalyzeAbsolute(rotated)
			require.NoError(t, err)
			require.Len(t, got.Positions, 7)
			for index, p := range got.Positions {
				expected := want.Positions[(first-1+index)%7]
				expected.Number = index + 1
				assert.Equal(t, expected, p)
			}
			root, err := modal.Relative(rotated, 1)
			require.NoError(t, err)
			assert.Equal(t, modal.RelativePosition{Number: 1, Category: modal.Neutral, Interval: interval.PerfectUnison()}, root)
		})
	}
	// Starting from the old seventh (C) gives M3 at the new third, whereas
	// the original Dorian third is m3. The categories follow the new root.
	rotated, err := mode.FromDegree(source, 7)
	require.NoError(t, err)
	p, err := modal.Relative(rotated, 3)
	require.NoError(t, err)
	assert.Equal(t, modal.High, p.Category)
	assert.True(t, p.Interval.Equal(interval.MajorThird()))
	assert.Equal(t, modal.Low, modal.MustRelative(source, 3).Category)
}
