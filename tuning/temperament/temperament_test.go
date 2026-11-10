package temperament_test

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedMeantoneOffset computes the exact reference offset T_alpha(m) for a 12-tone regular meantone
// temperament with fifth narrowed by (commaNum/commaDen)*SyntonicComma and chain start kMin.
func expectedMeantoneOffset(commaNum, commaDen, kMin, m int) cents.Cents {
	fifth := cents.MustFromRatio(3, 2)
	if commaNum != 0 {
		fifth = fifth.MustSub(cents.SyntonicComma().MustScale(commaNum, commaDen))
	}
	rem := (7*m - kMin) % 12
	if rem < 0 {
		rem += 12
	}
	k := kMin + rem
	_, rest := fifth.MustMul(k).SplitOctaves()
	return rest
}

// expectedWellOffsets computes the 12 exact reference offsets for a closed 12-tone well temperament
// from 12 fifth deviations delta_0..delta_11 starting at C (kMin = 0).
func expectedWellOffsets(deltas [12]cents.Cents) []cents.Cents {
	pureFifth := cents.MustFromRatio(3, 2)
	out := make([]cents.Cents, 12)
	acc := cents.Zero()
	for i := range 12 {
		m := (7 * i) % 12
		_, rest := acc.SplitOctaves()
		out[m] = rest
		acc = acc.MustAdd(pureFifth).MustAdd(deltas[i])
	}
	return out
}

// assertValidTemperamentInvariants verifies IV-11 and basic attributes on a valid temperament.
func assertValidTemperamentInvariants(t *testing.T, temp temperament.Temperament, wantName string, wantN int) {
	t.Helper()
	require.True(t, temp.IsValid())
	assert.Equal(t, wantName, temp.Name())
	assert.Equal(t, wantN, temp.StepsPerOctave())

	offsets := temp.Offsets()
	require.Len(t, offsets, wantN)
	assert.True(t, offsets[0].Equal(cents.Zero()), "T(0) must be 0 cents")

	for m := range wantN {
		got, err := temp.Offset(m)
		require.NoError(t, err)
		assert.True(t, got.Equal(offsets[m]), "Offset(%d) must match Offsets()[%d]", m, m)
		assert.Equal(t, got, temp.MustOffset(m))
		if m > 0 {
			assert.Less(t, offsets[m-1].Compare(offsets[m]), 0,
				"strict monotonicity IV-11 violated at step %d: %v >= %v", m, offsets[m-1], offsets[m])
		}
	}
	assert.Less(t, offsets[wantN-1].Compare(cents.Octave()), 0, "T(N-1) must be < 1200 cents")
}

// assertEDODeviationOracle independently verifies all fields of EDODeviation against the temperament offsets.
func assertEDODeviationOracle(t *testing.T, temp temperament.Temperament, edoDev temperament.EDODeviation) {
	t.Helper()
	n := temp.StepsPerOctave()
	assert.Equal(t, n, edoDev.StepsPerOctave)
	require.Len(t, edoDev.Deviations, n)
	assert.True(t, edoDev.Deviations[0].Equal(cents.Zero()))

	require.GreaterOrEqual(t, edoDev.MaxPosition, 0)
	require.Less(t, edoDev.MaxPosition, n)
	assert.True(t, edoDev.Deviations[edoDev.MaxPosition].Abs().Equal(edoDev.MaxAbsDeviation))

	sumAbs := cents.Zero()
	var sumSq float64
	for m, dev := range edoDev.Deviations {
		wantDev := temp.MustOffset(m).MustSub(cents.MustStep(m, n))
		assert.True(t, dev.Equal(wantDev), "EDO deviation at step %d", m)

		absDev := dev.Abs()
		if m < edoDev.MaxPosition {
			assert.Less(t, absDev.Compare(edoDev.MaxAbsDeviation), 0,
				"tie-breaking violated before MaxPosition at step %d", m)
		} else {
			assert.LessOrEqual(t, absDev.Compare(edoDev.MaxAbsDeviation), 0,
				"MaxAbsDeviation exceeded at step %d", m)
		}
		sumAbs = sumAbs.MustAdd(absDev)
		df := dev.Float64()
		sumSq += df * df
	}

	assert.True(t, edoDev.MeanAbsDeviation.Equal(sumAbs.MustDiv(n)))
	assert.InDelta(t, math.Sqrt(sumSq/float64(n)), edoDev.RMSDeviation, 1e-9)
}

// assertIntervalAnalysisOracle independently verifies all fields of IntervalAnalysis against Sizes, Deviations, and the temperament.
func assertIntervalAnalysisOracle(t *testing.T, temp temperament.Temperament, ia temperament.IntervalAnalysis) {
	t.Helper()
	n := temp.StepsPerOctave()
	require.Len(t, ia.Sizes, n)
	require.Len(t, ia.Deviations, n)

	require.GreaterOrEqual(t, ia.MinPosition, 0)
	require.Less(t, ia.MinPosition, n)
	assert.True(t, ia.Sizes[ia.MinPosition].Equal(ia.Min))

	require.GreaterOrEqual(t, ia.MaxPosition, 0)
	require.Less(t, ia.MaxPosition, n)
	assert.True(t, ia.Sizes[ia.MaxPosition].Equal(ia.Max))

	assert.True(t, ia.Spread.Equal(ia.Max.MustSub(ia.Min)))

	require.GreaterOrEqual(t, ia.WorstPosition, 0)
	require.Less(t, ia.WorstPosition, n)
	assert.True(t, ia.Deviations[ia.WorstPosition].Abs().Equal(ia.MaxAbsDeviation))

	sumAbs := cents.Zero()
	var sumSq float64
	pureCount := 0
	var wolves []int

	for m := range n {
		sz := ia.Sizes[m]
		dev := ia.Deviations[m]
		wantSz := temp.MustCyclicInterval(m, (m+ia.Step)%n)
		assert.True(t, sz.Equal(wantSz), "interval size at step %d", m)
		assert.True(t, dev.Equal(sz.MustSub(ia.Reference)), "interval deviation at step %d", m)

		if m < ia.MinPosition {
			assert.Greater(t, sz.Compare(ia.Min), 0, "MinPosition tie-breaking violated at step %d", m)
		} else {
			assert.GreaterOrEqual(t, sz.Compare(ia.Min), 0, "Min violated at step %d", m)
		}

		if m < ia.MaxPosition {
			assert.Less(t, sz.Compare(ia.Max), 0, "MaxPosition tie-breaking violated at step %d", m)
		} else {
			assert.LessOrEqual(t, sz.Compare(ia.Max), 0, "Max violated at step %d", m)
		}

		absDev := dev.Abs()
		if m < ia.WorstPosition {
			assert.Less(t, absDev.Compare(ia.MaxAbsDeviation), 0, "WorstPosition tie-breaking violated at step %d", m)
		} else {
			assert.LessOrEqual(t, absDev.Compare(ia.MaxAbsDeviation), 0, "MaxAbsDeviation violated at step %d", m)
		}

		sumAbs = sumAbs.MustAdd(absDev)
		df := dev.Float64()
		sumSq += df * df
		if dev.Equal(cents.Zero()) {
			pureCount++
		}
		if absDev.Compare(ia.WolfThreshold) > 0 {
			wolves = append(wolves, m)
		}
	}

	assert.Equal(t, pureCount, ia.PureCount)
	assert.True(t, ia.MeanAbsDeviation.Equal(sumAbs.MustDiv(n)))
	assert.InDelta(t, math.Sqrt(sumSq/float64(n)), ia.RMSDeviation, 1e-9)
	if len(wolves) == 0 {
		assert.Empty(t, ia.WolfPositions)
	} else {
		assert.Equal(t, wolves, ia.WolfPositions)
	}
}

// TestConstantsAndThresholds verifies exported constants and default wolf threshold functions.
func TestConstantsAndThresholds(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 65536, temperament.MaxStepsPerOctave)
	assert.Equal(t, 12, temperament.DefaultStepsPerOctave)
	assert.Equal(t, -5, temperament.DefaultPythagoreanMinFifth)
	assert.Equal(t, -3, temperament.EFlatPythagoreanMinFifth)
	assert.Equal(t, -3, temperament.DefaultMeantoneMinFifth)

	assert.True(t, temperament.DefaultFifthWolfThreshold().Equal(cents.MustFromInt(15)))
	assert.True(t, temperament.DefaultThirdWolfThreshold().Equal(cents.MustFromInt(25)))
}

// TestPresetsCatalog verifies the 31 built-in presets in Presets(), their order, names, aliases, and immutability (OP-1, S-1..S-5, K-1..K-8).
func TestPresetsCatalog(t *testing.T) {
	t.Parallel()

	presets := temperament.Presets()
	require.Len(t, presets, 31)

	wantCatalog := []struct {
		name string
		fn   func() temperament.Temperament
		n    int
	}{
		{"12-EDO", temperament.Equal12, 12},
		{"19-EDO", temperament.Equal19, 19},
		{"22-EDO", temperament.Equal22, 22},
		{"24-EDO", temperament.Equal24, 24},
		{"31-EDO", temperament.Equal31, 31},
		{"53-EDO", temperament.Equal53, 53},
		{"Pythagorean", temperament.Pythagorean, 12},
		{"Pythagorean (Eb-G#)", temperament.PythagoreanEFlatGSharp, 12},
		{"5-Limit Just Intonation", temperament.Just5Limit, 12},
		{"5-Limit Just Intonation (Symmetric I)", temperament.Just5LimitSymmetric1, 12},
		{"5-Limit Just Intonation (Symmetric II)", temperament.Just5LimitSymmetric2, 12},
		{"7-Limit Just Intonation", temperament.Just7Limit, 12},
		{"1/4-Comma Meantone", temperament.MeantoneQuarterComma, 12},
		{"1/3-Comma Meantone", temperament.MeantoneThirdComma, 12},
		{"1/5-Comma Meantone", temperament.MeantoneFifthComma, 12},
		{"1/6-Comma Meantone", temperament.MeantoneSixthComma, 12},
		{"2/7-Comma Meantone", temperament.MeantoneTwoSeventhComma, 12},
		{"Werckmeister III", temperament.WerckmeisterIII, 12},
		{"Werckmeister IV", temperament.WerckmeisterIV, 12},
		{"Werckmeister V", temperament.WerckmeisterV, 12},
		{"Kirnberger I", temperament.KirnbergerI, 12},
		{"Kirnberger II", temperament.KirnbergerII, 12},
		{"Kirnberger III", temperament.KirnbergerIII, 12},
		{"Vallotti", temperament.Vallotti, 12},
		{"Young I", temperament.YoungI, 12},
		{"Young II", temperament.YoungII, 12},
		{"Kellner", temperament.Kellner, 12},
		{"Neidhardt I", temperament.NeidhardtI, 12},
		{"Neidhardt II", temperament.NeidhardtII, 12},
		{"Neidhardt III", temperament.NeidhardtIII, 12},
		{"Neidhardt Dorf (1732)", temperament.NeidhardtDorf1732, 12},
	}

	for i, tc := range wantCatalog {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			direct := tc.fn()
			fromList := presets[i]
			assertValidTemperamentInvariants(t, direct, tc.name, tc.n)
			assert.True(t, direct.Equal(fromList))
			assert.Equal(t, direct.Name(), fromList.Name())
		})
	}

	t.Run("preset aliases", func(t *testing.T) {
		t.Parallel()

		assert.True(t, temperament.Equal().Equal(temperament.Equal12()))
		assert.Equal(t, temperament.Equal12().Name(), temperament.Equal().Name())
		assert.True(t, temperament.EDO12().Equal(temperament.Equal12()))
		assert.Equal(t, temperament.Equal12().Name(), temperament.EDO12().Name())
		assert.True(t, temperament.EDO19().Equal(temperament.Equal19()))
		assert.True(t, temperament.EDO22().Equal(temperament.Equal22()))
		assert.True(t, temperament.EDO24().Equal(temperament.Equal24()))
		assert.True(t, temperament.EDO31().Equal(temperament.Equal31()))
		assert.True(t, temperament.EDO53().Equal(temperament.Equal53()))
		assert.True(t, temperament.Just().Equal(temperament.Just5Limit()))
		assert.Equal(t, temperament.Just5Limit().Name(), temperament.Just().Name())
		assert.True(t, temperament.Meantone().Equal(temperament.MeantoneQuarterComma()))
		assert.Equal(t, temperament.MeantoneQuarterComma().Name(), temperament.Meantone().Name())
		assert.True(t, temperament.Young().Equal(temperament.YoungII()))
		assert.Equal(t, temperament.YoungII().Name(), temperament.Young().Name())
	})

	t.Run("Presets returns defensive copy", func(t *testing.T) {
		t.Parallel()

		p1 := temperament.Presets()
		p1[0] = temperament.Pythagorean()
		p2 := temperament.Presets()
		assert.Equal(t, "12-EDO", p2[0].Name())
		assert.True(t, p2[0].Equal(temperament.Equal12()))
	})
}

// TestEDOPresetsAndConstructor verifies Family 1 (N-EDO presets and EDO/MustEDO constructor, OP-1, OP-2, S-1, K-1, G-1, G-2).
func TestEDOPresetsAndConstructor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		n    int
		fn   func() temperament.Temperament
	}{
		{"12-EDO", 12, temperament.Equal12},
		{"19-EDO", 19, temperament.Equal19},
		{"22-EDO", 22, temperament.Equal22},
		{"24-EDO", 24, temperament.Equal24},
		{"31-EDO", 31, temperament.Equal31},
		{"53-EDO", 53, temperament.Equal53},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			preset := tc.fn()
			constructed, err := temperament.EDO(tc.n)
			require.NoError(t, err)
			mustConstructed := temperament.MustEDO(tc.n)

			assertValidTemperamentInvariants(t, preset, tc.name, tc.n)
			assert.True(t, preset.Equal(constructed))
			assert.True(t, preset.Equal(mustConstructed))
			assert.Equal(t, tc.name, constructed.Name())
			assert.True(t, preset.IsEqual())
			assert.False(t, preset.IsRational())

			for m := range tc.n {
				want := cents.MustStep(m, tc.n)
				got := preset.MustOffset(m)
				assert.True(t, got.Equal(want), "m=%d: got %v, want %v", m, got, want)
				num, den, ok := got.Fraction()
				require.True(t, ok)
				wantFloat := float64(1200*m) / float64(tc.n)
				assert.InDelta(t, wantFloat, float64(num)/float64(den), 1e-9)
			}
		})
	}

	t.Run("53-EDO step 1 equals HoldrianComma", func(t *testing.T) {
		t.Parallel()

		edo53 := temperament.Equal53()
		assert.True(t, edo53.MustOffset(1).Equal(cents.HoldrianComma()))
	})

	t.Run("G-1: 1-EDO single-step temperament", func(t *testing.T) {
		t.Parallel()

		edo1, err := temperament.EDO(1)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, edo1, "1-EDO", 1)
		assert.True(t, edo1.IsEqual())
		assert.True(t, edo1.IsRational(), "1-EDO has only T(0)=0 (1:1), which is rational")

		steps, err := edo1.StepSizes()
		require.NoError(t, err)
		require.Len(t, steps, 1)
		assert.True(t, steps[0].Equal(cents.Octave()))
		assert.True(t, edo1.MustMinStepSize().Equal(cents.Octave()))
		assert.True(t, edo1.MustMaxStepSize().Equal(cents.Octave()))

		for _, k := range []int{-3, -1, 0, 1, 5} {
			got, err := edo1.CumulativeOffset(k)
			require.NoError(t, err)
			assert.True(t, got.Equal(cents.Octave().MustMul(k)))
		}

		for _, r := range []int{-5, -1, 0, 1, 7} {
			rot, err := edo1.Rotate(r)
			require.NoError(t, err)
			assert.True(t, rot.Equal(edo1))
		}
	})

	t.Run("G-2: large microtonal divisions N=72, 1200, 65536", func(t *testing.T) {
		t.Parallel()

		edo72 := temperament.MustEDO(72)
		assert.Equal(t, "72-EDO", edo72.Name())
		assert.Equal(t, 72, edo72.StepsPerOctave())
		assert.True(t, edo72.IsEqual())
		assert.True(t, edo72.MustMinStepSize().Equal(cents.MustFromFraction(1200, 72)))

		edo1200 := temperament.MustEDO(1200)
		assert.Equal(t, 1200, edo1200.StepsPerOctave())
		assert.True(t, edo1200.MustOffset(1).Equal(cents.Cent()))
		assert.True(t, edo1200.MustMinStepSize().Equal(cents.Cent()))
		assert.True(t, edo1200.MustMaxStepSize().Equal(cents.Cent()))

		edoMax, err := temperament.EDO(temperament.MaxStepsPerOctave)
		require.NoError(t, err)
		assert.Equal(t, temperament.MaxStepsPerOctave, edoMax.StepsPerOctave())
		wantStep := cents.MustFromFraction(75, 4096)
		assert.True(t, edoMax.MustOffset(1).Equal(wantStep))
		assert.True(t, edoMax.MustMinStepSize().Equal(wantStep))
		assert.True(t, edoMax.MustMaxStepSize().Equal(wantStep))

		stepSizes := edoMax.MustStepSizes()
		require.Len(t, stepSizes, temperament.MaxStepsPerOctave)
		sum := cents.Zero()
		for _, s := range stepSizes {
			assert.True(t, s.Equal(wantStep))
			sum = sum.MustAdd(s)
		}
		assert.True(t, sum.Equal(cents.Octave()))
	})
}

// TestPythagoreanPresetsAndConstructors verifies Family 2 (Pythagorean presets and NewPythagorean/NewPythagoreanWithWolf, OP-1, OP-3, S-2, K-2).
func TestPythagoreanPresetsAndConstructors(t *testing.T) {
	t.Parallel()

	wantStandardRatios := [12][2]int{
		{1, 1}, {256, 243}, {9, 8}, {32, 27}, {81, 64}, {4, 3},
		{729, 512}, {3, 2}, {128, 81}, {27, 16}, {16, 9}, {243, 128},
	}
	wantStandardCents := [12]float64{
		0.0, 90.224996, 203.910002, 294.134997, 407.820003, 498.044999,
		611.730005, 701.955001, 792.179997, 905.865003, 996.089998, 1109.775004,
	}

	t.Run("Pythagorean C-symmetric preset (kMin=-5, wolf at 6->1)", func(t *testing.T) {
		t.Parallel()

		p := temperament.Pythagorean()
		assertValidTemperamentInvariants(t, p, "Pythagorean", 12)
		assert.False(t, p.IsEqual())
		assert.True(t, p.IsRational())

		for m := range 12 {
			want := cents.MustFromRatio(wantStandardRatios[m][0], wantStandardRatios[m][1])
			got := p.MustOffset(m)
			assert.True(t, got.Equal(want), "step %d: got %v, want %v", m, got, want)
			assert.InDelta(t, wantStandardCents[m], got.Float64(), 1e-6)
		}

		fromKMin, err := temperament.NewPythagorean(-5)
		require.NoError(t, err)
		assert.True(t, p.Equal(fromKMin))
		assert.Equal(t, "Pythagorean", fromKMin.Name())
		assert.True(t, p.Equal(temperament.MustNewPythagorean(-5)))

		fromWolf, err := temperament.NewPythagoreanWithWolf(6)
		require.NoError(t, err)
		assert.True(t, p.Equal(fromWolf))
		assert.Equal(t, "Pythagorean", fromWolf.Name())
		assert.True(t, p.Equal(temperament.MustNewPythagoreanWithWolf(6)))
	})

	t.Run("PythagoreanEFlatGSharp preset (kMin=-3, wolf at 8->3)", func(t *testing.T) {
		t.Parallel()

		ebGs := temperament.PythagoreanEFlatGSharp()
		assertValidTemperamentInvariants(t, ebGs, "Pythagorean (Eb-G#)", 12)
		assert.False(t, ebGs.IsEqual())
		assert.True(t, ebGs.IsRational())

		for m := range 12 {
			var want cents.Cents
			switch m {
			case 1:
				want = cents.MustFromRatio(2187, 2048) // Apotome C#
				assert.True(t, want.Equal(cents.Apotome()))
				assert.InDelta(t, 113.685006, ebGs.MustOffset(m).Float64(), 1e-6)
			case 8:
				want = cents.MustFromRatio(6561, 4096) // G#
				assert.InDelta(t, 815.640007, ebGs.MustOffset(m).Float64(), 1e-6)
			default:
				want = cents.MustFromRatio(wantStandardRatios[m][0], wantStandardRatios[m][1])
			}
			assert.True(t, ebGs.MustOffset(m).Equal(want), "step %d", m)
		}

		fromKMin, err := temperament.NewPythagorean(-3)
		require.NoError(t, err)
		assert.True(t, ebGs.Equal(fromKMin))
		assert.Equal(t, "Pythagorean (Eb-G#)", fromKMin.Name())

		// Decision 8 of the architecture: the chain start constant of this preset is EFlatPythagoreanMinFifth = -3,
		// and the preset is the eighth element of Presets().
		fromConst, err := temperament.NewPythagorean(temperament.EFlatPythagoreanMinFifth)
		require.NoError(t, err)
		assert.True(t, ebGs.Equal(fromConst))
		assert.Equal(t, "Pythagorean (Eb-G#)", fromConst.Name())
		assert.True(t, reflect.DeepEqual(ebGs, temperament.Presets()[7]))

		fromWolf, err := temperament.NewPythagoreanWithWolf(8)
		require.NoError(t, err)
		assert.True(t, ebGs.Equal(fromWolf))
	})

	t.Run("all 12 kMin in [-11, 0] and wolfPosition in [0, 11]", func(t *testing.T) {
		t.Parallel()

		pureFifth := cents.MustFromRatio(3, 2)
		wolfFifth := pureFifth.MustSub(cents.PythagoreanComma()) // 262144:177147

		for kMin := -11; kMin <= 0; kMin++ {
			temp, err := temperament.NewPythagorean(kMin)
			require.NoError(t, err)
			assert.Equal(t, 12, temp.StepsPerOctave())
			assert.True(t, temp.IsRational())

			switch kMin {
			case -5:
				assert.Equal(t, "Pythagorean", temp.Name())
			case -3:
				assert.Equal(t, "Pythagorean (Eb-G#)", temp.Name())
			default:
				assert.Equal(t, fmt.Sprintf("Pythagorean (kMin=%d)", kMin), temp.Name())
			}

			wolfPos := (7 * (kMin + 11)) % 12
			if wolfPos < 0 {
				wolfPos += 12
			}
			fromWolf, err := temperament.NewPythagoreanWithWolf(wolfPos)
			require.NoError(t, err)
			assert.True(t, temp.Equal(fromWolf), "kMin=%d wolfPos=%d", kMin, wolfPos)

			for m := range 12 {
				fifth := temp.MustCyclicInterval(m, (m+7)%12)
				if m == wolfPos {
					assert.True(t, fifth.Equal(wolfFifth), "kMin=%d wolf at %d: got %v", kMin, m, fifth)
				} else {
					assert.True(t, fifth.Equal(pureFifth), "kMin=%d pure fifth at %d: got %v", kMin, m, fifth)
				}
			}
		}
	})
}

// TestJustIntonationPresets verifies Family 3 (5-limit and 7-limit Just Intonation presets, OP-1, S-3, K-3, G-6).
func TestJustIntonationPresets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		fn        func() temperament.Temperament
		ratios    [12][2]int
		wantCents [12]float64
		symmetric bool
	}{
		{
			name: "5-Limit Just Intonation",
			fn:   temperament.Just5Limit,
			ratios: [12][2]int{
				{1, 1}, {16, 15}, {9, 8}, {6, 5}, {5, 4}, {4, 3},
				{45, 32}, {3, 2}, {8, 5}, {5, 3}, {9, 5}, {15, 8},
			},
			wantCents: [12]float64{
				0.0, 111.731285, 203.910002, 315.641287, 386.313714, 498.044999,
				590.223716, 701.955001, 813.686286, 884.358713, 1017.596288, 1088.268715,
			},
			symmetric: false,
		},
		{
			name: "5-Limit Just Intonation (Symmetric I)",
			fn:   temperament.Just5LimitSymmetric1,
			ratios: [12][2]int{
				{1, 1}, {16, 15}, {9, 8}, {6, 5}, {5, 4}, {4, 3},
				{45, 32}, {3, 2}, {8, 5}, {5, 3}, {16, 9}, {15, 8},
			},
			wantCents: [12]float64{
				0.0, 111.731285, 203.910002, 315.641287, 386.313714, 498.044999,
				590.223716, 701.955001, 813.686286, 884.358713, 996.089998, 1088.268715,
			},
			symmetric: true,
		},
		{
			name: "5-Limit Just Intonation (Symmetric II)",
			fn:   temperament.Just5LimitSymmetric2,
			ratios: [12][2]int{
				{1, 1}, {16, 15}, {10, 9}, {6, 5}, {5, 4}, {4, 3},
				{45, 32}, {3, 2}, {8, 5}, {5, 3}, {9, 5}, {15, 8},
			},
			wantCents: [12]float64{
				0.0, 111.731285, 182.403712, 315.641287, 386.313714, 498.044999,
				590.223716, 701.955001, 813.686286, 884.358713, 1017.596288, 1088.268715,
			},
			symmetric: true,
		},
		{
			name: "7-Limit Just Intonation",
			fn:   temperament.Just7Limit,
			ratios: [12][2]int{
				{1, 1}, {15, 14}, {8, 7}, {6, 5}, {5, 4}, {4, 3},
				{7, 5}, {3, 2}, {8, 5}, {5, 3}, {7, 4}, {15, 8},
			},
			wantCents: [12]float64{
				0.0, 119.442808, 231.174094, 315.641287, 386.313714, 498.044999,
				582.512193, 701.955001, 813.686286, 884.358713, 968.825906, 1088.268715,
			},
			symmetric: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			temp := tc.fn()
			assertValidTemperamentInvariants(t, temp, tc.name, 12)
			assert.False(t, temp.IsEqual())
			assert.True(t, temp.IsRational())

			for m := range 12 {
				want := cents.MustFromRatio(tc.ratios[m][0], tc.ratios[m][1])
				got := temp.MustOffset(m)
				assert.True(t, got.Equal(want), "step %d: got %v, want %v", m, got, want)
				assert.InDelta(t, tc.wantCents[m], got.Float64(), 1e-6)

				num, den, ok := got.Ratio()
				require.True(t, ok)
				assert.Equal(t, tc.ratios[m][0], num)
				assert.Equal(t, tc.ratios[m][1], den)
			}

			if tc.symmetric {
				for m := 1; m <= 5; m++ {
					sum := temp.MustOffset(m).MustAdd(temp.MustOffset(12 - m))
					assert.True(t, sum.Equal(cents.Octave()), "inversion symmetry at m=%d: sum=%v", m, sum)
				}
			}
		})
	}
}

// TestMeantonePresetsAndConstructors verifies Family 4 (regular meantone presets and constructors, OP-1, OP-4, S-4, K-4, G-5, G-6, OTK-4).
func TestMeantonePresetsAndConstructors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		fn        func() temperament.Temperament
		commaNum  int
		commaDen  int
		wantCents [12]float64
	}{
		{
			name:     "1/4-Comma Meantone",
			fn:       temperament.MeantoneQuarterComma,
			commaNum: 1, commaDen: 4,
			wantCents: [12]float64{
				0.0, 76.048999, 193.156857, 310.264715, 386.313714, 503.421572,
				579.470571, 696.578428, 772.627428, 889.735285, 1006.843143, 1082.892142,
			},
		},
		{
			name:     "1/3-Comma Meantone",
			fn:       temperament.MeantoneThirdComma,
			commaNum: 1, commaDen: 3,
			wantCents: [12]float64{
				0.0, 63.503664, 189.572475, 315.641287, 379.144951, 505.213762,
				568.717426, 694.786238, 758.289901, 884.358713, 1010.427525, 1073.931188,
			},
		},
		{
			name:     "1/5-Comma Meantone",
			fn:       temperament.MeantoneFifthComma,
			commaNum: 1, commaDen: 5,
			wantCents: [12]float64{
				0.0, 83.576201, 195.307486, 307.038771, 390.614972, 502.346257,
				585.922458, 697.653743, 781.229944, 892.961229, 1004.692514, 1088.268715,
			},
		},
		{
			name:     "1/6-Comma Meantone",
			fn:       temperament.MeantoneSixthComma,
			commaNum: 1, commaDen: 6,
			wantCents: [12]float64{
				0.0, 88.594335, 196.741239, 304.888142, 393.482477, 501.629381,
				590.223716, 698.370619, 786.964954, 895.111858, 1003.258761, 1091.853096,
			},
		},
		{
			name:     "2/7-Comma Meantone",
			fn:       temperament.MeantoneTwoSeventhComma,
			commaNum: 2, commaDen: 7,
			wantCents: [12]float64{
				0.0, 70.672427, 191.620693, 312.568960, 383.241387, 504.189653,
				574.862080, 695.810347, 766.482774, 887.431040, 1008.379307, 1079.051733,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			preset := tc.fn()
			assertValidTemperamentInvariants(t, preset, tc.name, 12)
			assert.False(t, preset.IsEqual())
			assert.False(t, preset.IsRational())

			for m := range 12 {
				want := expectedMeantoneOffset(tc.commaNum, tc.commaDen, -3, m)
				got := preset.MustOffset(m)
				assert.True(t, got.Equal(want), "step %d: got %v, want %v", m, got, want)
				assert.InDelta(t, tc.wantCents[m], got.Float64(), 1e-6)
			}

			constructed, err := temperament.NewMeantone(tc.commaNum, tc.commaDen)
			require.NoError(t, err)
			assert.True(t, preset.Equal(constructed))
			assert.Equal(t, tc.name, constructed.Name())
			assert.True(t, preset.Equal(temperament.MustNewMeantone(tc.commaNum, tc.commaDen)))

			withRange, err := temperament.NewMeantoneWithRange(tc.commaNum, tc.commaDen, -3)
			require.NoError(t, err)
			assert.True(t, preset.Equal(withRange))
			assert.Equal(t, tc.name, withRange.Name())
			assert.True(t, preset.Equal(temperament.MustNewMeantoneWithRange(tc.commaNum, tc.commaDen, -3)))

			fifth := preset.MustOffset(7)
			fromFifth, err := temperament.NewMeantoneFromFifth(fifth)
			require.NoError(t, err)
			assert.True(t, preset.Equal(fromFifth))
			assert.Equal(t, fmt.Sprintf("Meantone (fifth=%s)", fifth.String()), fromFifth.Name())
			assert.True(t, preset.Equal(temperament.MustNewMeantoneFromFifth(fifth)))

			fromFifthRange, err := temperament.NewMeantoneFromFifthWithRange(fifth, -3)
			require.NoError(t, err)
			assert.True(t, preset.Equal(fromFifthRange))
			assert.Equal(t, fmt.Sprintf("Meantone (fifth=%s)", fifth.String()), fromFifthRange.Name())
			assert.True(t, preset.Equal(temperament.MustNewMeantoneFromFifthWithRange(fifth, -3)))
		})
	}

	t.Run("OTK-4 and G-6: fixed old meantone bugs and exact pure intervals", func(t *testing.T) {
		t.Parallel()

		q := temperament.MeantoneQuarterComma()
		// Step 6 (F#) is 5*sqrt(5)/8 (~579.471 cents), NOT duplicate of step 9 (A ~889.735 cents).
		assert.False(t, q.MustOffset(6).Equal(q.MustOffset(9)))
		assert.InDelta(t, 579.470571, q.MustOffset(6).Float64(), 1e-6)
		// Step 10 (Bb) is 4/sqrt(5) (~1006.843 cents), NOT duplicate of step 3 (Eb ~310.265 cents).
		assert.False(t, q.MustOffset(10).Equal(q.MustOffset(3)))
		assert.InDelta(t, 1006.843143, q.MustOffset(10).Float64(), 1e-6)
		// Step 8 (G#) is exact 25:16.
		assert.True(t, q.MustOffset(8).Equal(cents.MustFromRatio(25, 16)))
		// Step 4 (E) is exact 5:4.
		assert.True(t, q.MustOffset(4).Equal(cents.MustFromRatio(5, 4)))
		assert.True(t, q.MustCyclicInterval(0, 4).Equal(cents.MustFromRatio(5, 4)))
		// Wolf fifth G#-Eb (8->3) is ~737.637287 cents.
		assert.InDelta(t, 737.637287, q.MustCyclicInterval(8, 3).Float64(), 1e-6)

		// 1/3-comma meantone has exact 6:5 at m=3 (Eb), 25:18 at m=6 (F#), 5:3 at m=9 (A).
		m3 := temperament.MeantoneThirdComma()
		assert.True(t, m3.MustOffset(3).Equal(cents.MustFromRatio(6, 5)))
		assert.True(t, m3.MustCyclicInterval(0, 3).Equal(cents.MustFromRatio(6, 5)))
		assert.True(t, m3.MustOffset(6).Equal(cents.MustFromRatio(25, 18)))
		assert.True(t, m3.MustOffset(9).Equal(cents.MustFromRatio(5, 3)))

		// 1/5-comma meantone has exact 15:8 at m=11 (B).
		m5 := temperament.MeantoneFifthComma()
		assert.True(t, m5.MustOffset(11).Equal(cents.MustFromRatio(15, 8)))

		// 1/6-comma meantone has exact 45:32 at m=6 (F#).
		m6 := temperament.MeantoneSixthComma()
		assert.True(t, m6.MustOffset(6).Equal(cents.MustFromRatio(45, 32)))

		// 2/7-comma meantone has exact 25:24 (MinorChroma) at m=1 (C#).
		m27 := temperament.MeantoneTwoSeventhComma()
		assert.True(t, m27.MustOffset(1).Equal(cents.MustFromRatio(25, 24)))
	})

	t.Run("NewMeantone naming and fraction reduction", func(t *testing.T) {
		t.Parallel()

		// Reducible fraction 2/8 -> "1/4-Comma Meantone"
		m28, err := temperament.NewMeantone(2, 8)
		require.NoError(t, err)
		assert.Equal(t, "1/4-Comma Meantone", m28.Name())
		assert.True(t, m28.Equal(temperament.MeantoneQuarterComma()))

		// kMin != -3 -> "<num>/<den>-Comma Meantone (kMin=<kMin>)"
		mAlt, err := temperament.NewMeantoneWithRange(1, 4, -4)
		require.NoError(t, err)
		assert.Equal(t, "1/4-Comma Meantone (kMin=-4)", mAlt.Name())
		// In kMin=-4 chain, m=8 is Ab = 8:5.
		assert.True(t, mAlt.MustOffset(8).Equal(cents.MustFromRatio(8, 5)))

		// commaNum == 0 -> Pythagorean naming
		p0, err := temperament.NewMeantone(0, 4)
		require.NoError(t, err)
		assert.Equal(t, "Pythagorean (Eb-G#)", p0.Name())
		assert.True(t, p0.Equal(temperament.PythagoreanEFlatGSharp()))

		pSym, err := temperament.NewMeantoneWithRange(0, 1, -5)
		require.NoError(t, err)
		assert.Equal(t, "Pythagorean", pSym.Name())
		assert.True(t, pSym.Equal(temperament.Pythagorean()))

		pOther, err := temperament.NewMeantoneWithRange(0, 1, -6)
		require.NoError(t, err)
		assert.Equal(t, "Pythagorean (kMin=-6)", pOther.Name())

		fifth := cents.MustFromInt(700)
		mfRange, err := temperament.NewMeantoneFromFifthWithRange(fifth, -5)
		require.NoError(t, err)
		assert.Equal(t, "Meantone (fifth=700, kMin=-5)", mfRange.Name())
		assert.True(t, mfRange.Equal(temperament.Equal12()))
	})

	t.Run("G-5: existence boundaries for regular meantone fifths (4800/7, 720)", func(t *testing.T) {
		t.Parallel()

		// Exact lower boundary 4800/7 cents -> chromatic semitone = 0 -> error
		lowBound := cents.MustFromFraction(4800, 7)
		_, err := temperament.NewMeantoneFromFifth(lowBound)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Exact upper boundary 720 cents -> diatonic semitone = 0 -> error
		highBound := cents.MustFromInt(720)
		_, err = temperament.NewMeantoneFromFifth(highBound)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Strictly inside lower boundary: 4800/7 + 1/1000 cents -> valid
		justAboveLow := lowBound.MustAdd(cents.MustFromFraction(1, 1000))
		tLow, err := temperament.NewMeantoneFromFifth(justAboveLow)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, tLow, tLow.Name(), 12)

		// Strictly inside upper boundary: 720 - 1/1000 cents -> valid
		justBelowHigh := highBound.MustSub(cents.MustFromFraction(1, 1000))
		tHigh, err := temperament.NewMeantoneFromFifth(justBelowHigh)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, tHigh, tHigh.Name(), 12)
	})
}

// TestWellTemperamentPresets verifies Family 5 (all 14 historical well temperament presets, OP-1, S-5, K-5..K-8, K-25, G-4, G-6, OTK-4).
func TestWellTemperamentPresets(t *testing.T) {
	t.Parallel()

	pc := cents.PythagoreanComma()
	sc := cents.SyntonicComma()
	sch := cents.Schisma()
	z := cents.Zero()

	qPc := pc.MustScale(-1, 4)   // -1/4 Pc
	pPc4 := pc.MustScale(1, 4)   // +1/4 Pc
	tPc := pc.MustScale(-1, 3)   // -1/3 Pc
	pPc3 := pc.MustScale(1, 3)   // +1/3 Pc
	sPc := pc.MustScale(-1, 6)   // -1/6 Pc
	twPc := pc.MustScale(-1, 12) // -1/12 Pc
	fPc := pc.MustScale(-1, 5)   // -1/5 Pc

	negSc := sc.Neg()             // -Sc
	halfSc := sc.MustScale(-1, 2) // -1/2 Sc
	qSc := sc.MustScale(-1, 4)    // -1/4 Sc
	negSch := sch.Neg()           // -Schisma

	// Young I: -3/16 Sc on C-G-D-A-E, -(1/4 Pc - 3/16 Sc) on E-B, B-F#, Bb-F, F-C
	y1Main := sc.MustScale(-3, 16)
	y1Side := pc.MustScale(1, 4).MustSub(sc.MustScale(3, 16)).Neg()

	cases := []struct {
		name      string
		fn        func() temperament.Temperament
		deltas    [12]cents.Cents
		wantCents [12]float64
	}{
		{
			name:   "Werckmeister III",
			fn:     temperament.WerckmeisterIII,
			deltas: [12]cents.Cents{qPc, qPc, qPc, z, z, qPc, z, z, z, z, z, z},
			wantCents: [12]float64{
				0.0, 90.224996, 192.179997, 294.134997, 390.224996, 498.044999,
				588.269995, 696.089998, 792.179997, 888.269995, 996.089998, 1092.179997,
			},
		},
		{
			name:   "Werckmeister IV",
			fn:     temperament.WerckmeisterIV,
			deltas: [12]cents.Cents{tPc, z, tPc, z, tPc, z, tPc, z, pPc3, pPc3, tPc, z},
			wantCents: [12]float64{
				0.0, 82.404992, 196.089998, 294.134997, 392.179997, 498.044999,
				588.269995, 694.134997, 784.359993, 890.224996, 1003.910002, 1086.314994,
			},
		},
		{
			name:   "Werckmeister V",
			fn:     temperament.WerckmeisterV,
			deltas: [12]cents.Cents{z, z, qPc, qPc, z, z, qPc, qPc, pPc4, z, z, qPc},
			wantCents: [12]float64{
				0.0, 96.089998, 203.910002, 300.000000, 396.089998, 503.910002,
				600.000000, 701.955001, 792.179997, 900.000000, 1001.955001, 1098.044999,
			},
		},
		{
			name:   "Kirnberger I",
			fn:     temperament.KirnbergerI,
			deltas: [12]cents.Cents{z, z, negSc, z, z, z, negSch, z, z, z, z, z},
			wantCents: [12]float64{
				0.0, 90.224996, 203.910002, 294.134997, 386.313714, 498.044999,
				590.223716, 701.955001, 792.179997, 884.358713, 996.089998, 1088.268715,
			},
		},
		{
			name:   "Kirnberger II",
			fn:     temperament.KirnbergerII,
			deltas: [12]cents.Cents{z, z, halfSc, halfSc, z, z, negSch, z, z, z, z, z},
			wantCents: [12]float64{
				0.0, 90.224996, 203.910002, 294.134997, 386.313714, 498.044999,
				590.223716, 701.955001, 792.179997, 895.111858, 996.089998, 1088.268715,
			},
		},
		{
			name:   "Kirnberger III",
			fn:     temperament.KirnbergerIII,
			deltas: [12]cents.Cents{qSc, qSc, qSc, qSc, z, z, negSch, z, z, z, z, z},
			wantCents: [12]float64{
				0.0, 90.224996, 193.156857, 294.134997, 386.313714, 498.044999,
				590.223716, 696.578428, 792.179997, 889.735285, 996.089998, 1088.268715,
			},
		},
		{
			name:   "Vallotti",
			fn:     temperament.Vallotti,
			deltas: [12]cents.Cents{sPc, sPc, sPc, sPc, sPc, z, z, z, z, z, z, sPc},
			wantCents: [12]float64{
				0.0, 94.134997, 196.089998, 298.044999, 392.179997, 501.955001,
				592.179997, 698.044999, 796.089998, 894.134997, 1000.000000, 1090.224996,
			},
		},
		{
			name:   "Young I",
			fn:     temperament.YoungI,
			deltas: [12]cents.Cents{y1Main, y1Main, y1Main, y1Main, y1Side, y1Side, z, z, z, z, y1Side, y1Side},
			wantCents: [12]float64{
				0.0, 93.890142, 195.845143, 297.800144, 391.690286, 499.877572,
				591.935141, 697.922572, 795.845143, 893.767715, 999.755145, 1091.812714,
			},
		},
		{
			name:   "Young II",
			fn:     temperament.YoungII,
			deltas: [12]cents.Cents{sPc, sPc, sPc, sPc, sPc, sPc, z, z, z, z, z, z},
			wantCents: [12]float64{
				0.0, 90.224996, 196.089998, 294.134997, 392.179997, 498.044999,
				588.269995, 698.044999, 792.179997, 894.134997, 996.089998, 1090.224996,
			},
		},
		{
			name:   "Kellner",
			fn:     temperament.Kellner,
			deltas: [12]cents.Cents{fPc, fPc, fPc, fPc, z, fPc, z, z, z, z, z, z},
			wantCents: [12]float64{
				0.0, 90.224996, 194.525998, 294.134997, 389.051995, 498.044999,
				588.269995, 697.262999, 792.179997, 891.788996, 996.089998, 1091.006996,
			},
		},
		// Neidhardt presets: schemes and tables of the knowledge base (temperament topic, data/temperament-fifth-schemes.csv
		// rows neidhardt-*, data/temperament-presets.csv columns "центы"); fifths narrowed by k/12 Pc, sum -Pc (K-8).
		{
			name:   "Neidhardt I", // neidhardt-dorf-1724: 2,2,2,2,1,1,0,0,1,1,0,0 twelfths
			fn:     temperament.NeidhardtI,
			deltas: [12]cents.Cents{sPc, sPc, sPc, sPc, twPc, twPc, z, z, twPc, twPc, z, z},
			wantCents: [12]float64{
				0.0, 94.134997, 196.089998, 296.089998, 392.179997, 498.044999,
				592.179997, 698.044999, 796.089998, 894.134997, 996.089998, 1092.179997,
			},
		},
		{
			name:   "Neidhardt II", // neidhardt-kleine-stadt-1724: 2,2,2,1,0,1,1,1,0,0,1,1 twelfths
			fn:     temperament.NeidhardtII,
			deltas: [12]cents.Cents{sPc, sPc, sPc, twPc, z, twPc, twPc, twPc, z, z, twPc, twPc},
			wantCents: [12]float64{
				0.0, 96.089998, 196.089998, 298.044999, 394.134997, 500.000000,
				596.089998, 698.044999, 796.089998, 894.134997, 1000.000000, 1096.089998,
			},
		},
		{
			name:   "Neidhardt III", // neidhardt-grosse-stadt-1724: 2,2,2,1,0,1,1,1,0,1,1,0 twelfths
			fn:     temperament.NeidhardtIII,
			deltas: [12]cents.Cents{sPc, sPc, sPc, twPc, z, twPc, twPc, twPc, z, twPc, twPc, z},
			wantCents: [12]float64{
				0.0, 96.089998, 196.089998, 298.044999, 394.134997, 498.044999,
				596.089998, 698.044999, 796.089998, 894.134997, 998.044999, 1096.089998,
			},
		},
		{
			name:   "Neidhardt Dorf (1732)", // neidhardt-dorf-1732: 1,2,3,3,0,1,0,1,0,0,1,0 twelfths
			fn:     temperament.NeidhardtDorf1732,
			deltas: [12]cents.Cents{twPc, sPc, qPc, qPc, z, twPc, z, twPc, z, z, twPc, z},
			wantCents: [12]float64{
				0.0, 94.134997, 198.044999, 296.089998, 390.224996, 498.044999,
				592.179997, 700.000000, 794.134997, 894.134997, 998.044999, 1092.179997,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			preset := tc.fn()
			assertValidTemperamentInvariants(t, preset, tc.name, 12)
			assert.False(t, preset.IsEqual())
			if tc.name == "Kirnberger I" {
				assert.True(t, preset.IsRational())
			} else {
				assert.False(t, preset.IsRational())
			}

			wantOffsets := expectedWellOffsets(tc.deltas)
			for m := range 12 {
				got := preset.MustOffset(m)
				assert.True(t, got.Equal(wantOffsets[m]), "step %d: got %v, want %v", m, got, wantOffsets[m])
				assert.InDelta(t, tc.wantCents[m], got.Float64(), 1e-6)
			}

			// Reconstruct via FromFifthDeviations (12 elements and 11 elements)
			from12, err := temperament.FromFifthDeviations(tc.name, tc.deltas[:])
			require.NoError(t, err)
			assert.True(t, preset.Equal(from12))

			from11, err := temperament.FromFifthDeviations(tc.name, tc.deltas[:11])
			require.NoError(t, err)
			assert.True(t, preset.Equal(from11))
		})
	}

	t.Run("OTK-4: fixed old Kirnberger, Vallotti, and Young bugs", func(t *testing.T) {
		t.Parallel()

		// Kirnberger I & II have D = 9:8 (~203.910 cents); Kirnberger III has D = sqrt(5)/2 (~193.157 cents), NOT ~452.933 cents.
		assert.True(t, temperament.KirnbergerI().MustOffset(2).Equal(cents.MustFromRatio(9, 8)))
		assert.True(t, temperament.KirnbergerII().MustOffset(2).Equal(cents.MustFromRatio(9, 8)))
		assert.True(t, temperament.KirnbergerIII().MustOffset(2).Equal(temperament.MeantoneQuarterComma().MustOffset(2)))
		assert.InDelta(t, 193.156857, temperament.KirnbergerIII().MustOffset(2).Float64(), 1e-6)

		// All three Kirnberger presets have exact pure major third C-E = 5:4.
		for _, k := range []temperament.Temperament{
			temperament.KirnbergerI(),
			temperament.KirnbergerII(),
			temperament.KirnbergerIII(),
		} {
			assert.True(t, k.MustOffset(4).Equal(cents.MustFromRatio(5, 4)))
			assert.True(t, k.MustCyclicInterval(0, 4).Equal(cents.MustFromRatio(5, 4)))
		}

		// Vallotti, WerckmeisterIII, YoungI, YoungII are pairwise distinct.
		assert.False(t, temperament.Vallotti().Equal(temperament.WerckmeisterIII()))
		assert.False(t, temperament.Vallotti().Equal(temperament.YoungII()))
		assert.False(t, temperament.YoungI().Equal(temperament.YoungII()))
		assert.False(t, temperament.Vallotti().Equal(temperament.YoungI()))
	})

	t.Run("K-8: Neidhardt fifth corrections close the circle and NeidhardtI keeps pure F and Bb", func(t *testing.T) {
		t.Parallel()

		for _, tc := range cases {
			if !strings.HasPrefix(tc.name, "Neidhardt") {
				continue
			}
			sum := cents.Zero()
			for _, d := range tc.deltas {
				sum = sum.MustAdd(d)
			}
			assert.True(t, sum.Equal(pc.Neg()), "%s: fifth corrections sum to %v, want -Pc", tc.name, sum)
		}

		// neidhardt-dorf-1724: the fifths Bb-F and F-C are pure, so F = 4:3 and Bb = 16:9 exactly.
		assert.True(t, temperament.NeidhardtI().MustOffset(5).Equal(cents.MustFromRatio(4, 3)))
		assert.True(t, temperament.NeidhardtI().MustOffset(10).Equal(cents.MustFromRatio(16, 9)))
	})

	t.Run("G-4 and K-25: exact algebraic cancellation of commas", func(t *testing.T) {
		t.Parallel()

		flatSteps := []struct {
			m   int
			num int
			den int
		}{
			{1, 256, 243},
			{3, 32, 27},
			{5, 4, 3},
			{8, 128, 81},
			{10, 16, 9},
		}
		presetsWithPureFlats := []temperament.Temperament{
			temperament.WerckmeisterIII(),
			temperament.KirnbergerI(),
			temperament.KirnbergerII(),
			temperament.KirnbergerIII(),
			temperament.YoungII(),
			temperament.Kellner(),
		}
		for _, p := range presetsWithPureFlats {
			for _, fs := range flatSteps {
				got := p.MustOffset(fs.m)
				want := cents.MustFromRatio(fs.num, fs.den)
				assert.True(t, got.Equal(want), "%s step %d: got %v, want %v", p.Name(), fs.m, got, want)
				rNum, rDen, ok := got.Ratio()
				require.True(t, ok, "%s step %d must be an exact ratio", p.Name(), fs.m)
				assert.Equal(t, fs.num, rNum)
				assert.Equal(t, fs.den, rDen)
			}
		}

		// Step 6 (F#) in WerckmeisterIII, WerckmeisterIV, YoungII, Kellner is exact Pythagorean 1024:729.
		wantPythTritone := cents.MustFromRatio(1024, 729)
		for _, p := range []temperament.Temperament{
			temperament.WerckmeisterIII(),
			temperament.WerckmeisterIV(),
			temperament.YoungII(),
			temperament.Kellner(),
		} {
			got := p.MustOffset(6)
			assert.True(t, got.Equal(wantPythTritone), "%s step 6: got %v", p.Name(), got)
			rNum, rDen, ok := got.Ratio()
			require.True(t, ok)
			assert.Equal(t, 1024, rNum)
			assert.Equal(t, 729, rDen)
		}

		// Step 6 (F#) in Kirnberger I, II, III is exact 5-limit tritone 45:32.
		wantJustTritone := cents.MustFromRatio(45, 32)
		for _, p := range []temperament.Temperament{
			temperament.KirnbergerI(),
			temperament.KirnbergerII(),
			temperament.KirnbergerIII(),
		} {
			got := p.MustOffset(6)
			assert.True(t, got.Equal(wantJustTritone), "%s step 6: got %v", p.Name(), got)
			rNum, rDen, ok := got.Ratio()
			require.True(t, ok)
			assert.Equal(t, 45, rNum)
			assert.Equal(t, 32, rDen)
		}
	})
}

// TestCustomConstructors verifies New, FromOffsets, FromFloats, FromRatios, FromFifths*, FromFifthDeviations*, FromGenerator (OP-5..OP-8, S-6..S-9, K-9..K-12, G-10).
func TestCustomConstructors(t *testing.T) {
	t.Parallel()

	t.Run("OP-5: New, FromOffsets, FromFloats and slice defensive copying (G-10)", func(t *testing.T) {
		t.Parallel()

		input := []cents.Cents{
			cents.Zero(),
			cents.MustFromInt(204),
			cents.MustFromInt(386),
			cents.MustFromInt(498),
			cents.MustFromInt(702),
			cents.MustFromInt(884),
			cents.MustFromInt(1088),
		}
		origSecond := input[1]

		temp, err := temperament.New("7-Tone Diatonic", input)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, temp, "7-Tone Diatonic", 7)

		fromOff, err := temperament.FromOffsets("7-Tone Diatonic", input)
		require.NoError(t, err)
		assert.True(t, temp.Equal(fromOff))
		assert.True(t, temp.Equal(temperament.MustNew("7-Tone Diatonic", input)))
		assert.True(t, temp.Equal(temperament.MustFromOffsets("7-Tone Diatonic", input)))

		// Mutate caller's input slice -> must not affect temp.
		input[1] = cents.MustFromInt(999)
		assert.True(t, temp.MustOffset(1).Equal(origSecond))

		// Mutate returned Offsets() slice -> must not affect temp.
		out := temp.Offsets()
		out[1] = cents.MustFromInt(888)
		assert.True(t, temp.MustOffset(1).Equal(origSecond))

		// FromFloats
		floats := []float64{0, 100.5, 350.25, 700.125, 1050.75}
		tf, err := temperament.FromFloats("Custom Float", floats)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, tf, "Custom Float", 5)
		assert.True(t, tf.Equal(temperament.MustFromFloats("Custom Float", floats)))
		assert.True(t, tf.MustOffset(1).Equal(cents.MustFromFloat(100.5)))
	})

	t.Run("OP-6: FromRatios in 31-limit including 11-, 13-, 31-limit ratios", func(t *testing.T) {
		t.Parallel()

		justRatios := []temperament.Ratio{
			{1, 1}, {16, 15}, {9, 8}, {6, 5}, {5, 4}, {4, 3},
			{45, 32}, {3, 2}, {8, 5}, {5, 3}, {9, 5}, {15, 8},
		}
		justCustom, err := temperament.FromRatios("Custom Just", justRatios)
		require.NoError(t, err)
		assert.True(t, justCustom.Equal(temperament.Just()))
		assert.True(t, justCustom.IsRational())
		assert.True(t, justCustom.Equal(temperament.MustFromRatios("Custom Just", justRatios)))

		// Non-reduced ratios with prime factor > 31 that cancels out (e.g. 37*3 / 37*2 = 3:2) and 31-limit primes (11, 13, 31)
		higherLimit := []temperament.Ratio{
			{2, 2},    // 1:1 (unison with Num == Den > 1)
			{31, 30},  // 31-limit semitone
			{11, 9},   // 11-limit neutral third
			{13, 9},   // 13-limit tritone
			{111, 74}, // 37*3 : 37*2 -> reduces to 3:2 (3-limit)
			{62, 33},  // 2*31 : 3*11
		}
		hlTemp, err := temperament.FromRatios("31-Limit Microtonal", higherLimit)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, hlTemp, "31-Limit Microtonal", 6)
		assert.True(t, hlTemp.IsRational())
		assert.True(t, hlTemp.MustOffset(4).Equal(cents.MustFromRatio(3, 2)))
	})

	t.Run("OP-7: FromFifths* and FromFifthDeviations* including Vallotti's original scheme", func(t *testing.T) {
		t.Parallel()

		// 1. Reconstruct Vallotti (1/6 Pc) via FromFifths and FromFifthsWithStart
		pureFifth := cents.MustFromRatio(3, 2)
		narrow := pureFifth.MustSub(cents.PythagoreanComma().MustDiv(6))
		fifths12 := []cents.Cents{
			narrow, narrow, narrow, narrow, narrow,
			pureFifth, pureFifth, pureFifth, pureFifth, pureFifth, pureFifth,
			narrow, // F-C
		}
		v12, err := temperament.FromFifths("Vallotti 12", fifths12)
		require.NoError(t, err)
		assert.True(t, v12.Equal(temperament.Vallotti()))
		assert.True(t, v12.Equal(temperament.MustFromFifths("Vallotti 12", fifths12)))

		v11, err := temperament.FromFifths("Vallotti 11", fifths12[:11])
		require.NoError(t, err)
		assert.True(t, v11.Equal(temperament.Vallotti()))

		// 2. Reconstruct 1/4-comma meantone via FromFifthsWithStart (11 fifths on kMin = -3)
		mtFifth := pureFifth.MustSub(cents.SyntonicComma().MustDiv(4))
		mtFifths11 := make([]cents.Cents, 11)
		for i := range 11 {
			mtFifths11[i] = mtFifth
		}
		mtFromChain, err := temperament.FromFifthsWithStart("Meantone Chain", mtFifths11, -3)
		require.NoError(t, err)
		assert.True(t, mtFromChain.Equal(temperament.MeantoneQuarterComma()))
		assert.True(t, mtFromChain.Equal(temperament.MustFromFifthsWithStart("Meantone Chain", mtFifths11, -3)))

		// Same via FromFifthDeviationsWithStart
		mtDev := cents.SyntonicComma().MustScale(-1, 4)
		mtDevs11 := make([]cents.Cents, 11)
		for i := range 11 {
			mtDevs11[i] = mtDev
		}
		mtFromDevs, err := temperament.FromFifthDeviationsWithStart("Meantone Devs", mtDevs11, -3)
		require.NoError(t, err)
		assert.True(t, mtFromDevs.Equal(temperament.MeantoneQuarterComma()))
		assert.True(t, mtFromDevs.Equal(temperament.MustFromFifthDeviationsWithStart("Meantone Devs", mtDevs11, -3)))

		// 3. Vallotti's original (manuscript) scheme, taken from the domain knowledge base
		// (data/temperament-fifth-schemes.csv, row vallotti-original, source WP-en-Vallotti "Vallotti's original"):
		// C-G, G-D, D-A, A-E, E-B and F-C narrowed by 1/6 syntonic comma, B-F#, F#-C#, C#-G#, G#-Eb, Eb-Bb pure,
		// and Bb-F narrowed by one schisma. The six -1/6 Sc fifths and the one -Schisma fifth sum to -Pc.
		dSc6 := cents.SyntonicComma().MustScale(-1, 6)
		z := cents.Zero()
		vallottiMsDevs := []cents.Cents{
			dSc6, dSc6, dSc6, dSc6, dSc6, // C-G, G-D, D-A, A-E, E-B
			z, z, z, z, z, // B-F#, F#-C#, C#-G#, G#-Eb, Eb-Bb
			cents.Schisma().Neg(), // Bb-F
			dSc6,                  // F-C
		}
		vallottiMs, err := temperament.FromFifthDeviations("Vallotti Original", vallottiMsDevs)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, vallottiMs, "Vallotti Original", 12)
		assert.True(t, vallottiMs.Equal(temperament.MustFromFifthDeviations("Vallotti Original", vallottiMsDevs)))

		// Reference offsets of the vallotti-original table in data/temperament-presets.csv (rounded to 1e-6 cent).
		wantVallottiMs := [12]float64{
			0, 95.763098, 196.741239, 299.673100, 393.482477, 501.629381,
			593.808097, 698.370619, 797.718099, 895.111858, 1001.628101, 1091.853096,
		}
		for m, want := range wantVallottiMs {
			assert.InDelta(t, want, vallottiMs.MustOffset(m).Float64(), 1e-6, "Vallotti original offset at m=%d", m)
		}

		// S-15, K-18: the two Vallotti forms are distinct. Along the circle from C both reach Bb after 10 fifths,
		// five of which are narrowed: by 5/6 Pc in Vallotti() and by 5/6 Sc in the original, so the largest
		// offset difference is exactly 5/6 (Pc - Sc) = 5/6 Schisma (~1.628101 cents, Bb: 1001.628101 vs 1000.000000).
		vallotti := temperament.Vallotti()
		assert.False(t, vallottiMs.Equal(vallotti))
		maxDiff := cents.Schisma().MustScale(5, 6)
		assert.InDelta(t, 1.628101, maxDiff.Float64(), 1e-6)
		assert.True(t, vallottiMs.MustOffset(10).MustSub(vallotti.MustOffset(10)).Equal(maxDiff))
		assert.True(t, vallottiMs.Near(vallotti, maxDiff))
		assert.True(t, vallotti.Near(vallottiMs, maxDiff))
		assert.False(t, vallottiMs.Near(vallotti, maxDiff.MustSub(cents.MustFromFraction(1, 1_000_000))))
		assert.False(t, vallottiMs.Near(vallotti, cents.MustFromInt(1)), "S-15: Near must be false at 1.0 cent")
		assert.True(t, vallottiMs.Near(vallotti, cents.MustFromInt(2)), "S-15: Near must be true at 2.0 cents")
		assert.True(t, vallotti.Near(vallottiMs, cents.MustFromInt(2)), "S-15: Near symmetry at 2.0 cents")
	})

	t.Run("OP-8: FromGenerator for N=12 meantone and N=7 diatonic chain", func(t *testing.T) {
		t.Parallel()

		// N=12, kMin=-3, g = R(3:2) - 1/4 Sc -> equals MeantoneQuarterComma
		gQuarter := cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4))
		gen12, err := temperament.FromGenerator("Generated Meantone", 12, gQuarter, -3)
		require.NoError(t, err)
		assert.True(t, gen12.Equal(temperament.MeantoneQuarterComma()))
		assert.True(t, gen12.Equal(temperament.MustFromGenerator("Generated Meantone", 12, gQuarter, -3)))

		// N=7 Pythagorean diatonic scale (F-C-G-D-A-E-B, kMin=-1..+5)
		gen7, err := temperament.FromGenerator("Pythagorean Diatonic", 7, cents.MustFromRatio(3, 2), -1)
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, gen7, "Pythagorean Diatonic", 7)
		assert.True(t, gen7.IsRational())
	})
}

// TestAttributesIntervalsAndRotation verifies OP-9..OP-14 (WithName, Offset, SplitStep, CumulativeOffset, DirectedInterval, CyclicInterval, StepSizes, Rotate, S-10..S-14, S-20..S-22, G-3, G-8, G-9).
func TestAttributesIntervalsAndRotation(t *testing.T) {
	t.Parallel()

	t.Run("OP-9 and OP-20: WithName and String()", func(t *testing.T) {
		t.Parallel()

		w3 := temperament.WerckmeisterIII()
		assert.Equal(t, "Werckmeister III (12 steps/octave)", w3.String())

		renamed, err := w3.WithName("Custom W3")
		require.NoError(t, err)
		assert.Equal(t, "Custom W3", renamed.Name())
		assert.Equal(t, "Werckmeister III", w3.Name(), "original temperament must remain unchanged")
		assert.True(t, w3.Equal(renamed))
		assert.Equal(t, "Custom W3 (12 steps/octave)", renamed.String())
		assert.Equal(t, renamed.Name(), w3.MustWithName("Custom W3").Name())

		unnamed, err := w3.WithName("")
		require.NoError(t, err)
		assert.Equal(t, "", unnamed.Name())
		assert.Equal(t, "Custom (12 steps/octave)", unnamed.String())
	})

	t.Run("OP-11 and G-3: SplitStep and CumulativeOffset across negative, zero, and positive octaves", func(t *testing.T) {
		t.Parallel()

		just := temperament.Just()
		t11 := just.MustOffset(11) // 15:8

		cases := []struct {
			step    int
			wantOct int
			wantOff cents.Cents
			wantCum cents.Cents
		}{
			{0, 0, cents.Zero(), cents.Zero()},
			{1, 0, just.MustOffset(1), just.MustOffset(1)},
			{11, 0, t11, t11},
			{12, 1, cents.Zero(), cents.Octave()},
			{13, 1, just.MustOffset(1), cents.Octave().MustAdd(just.MustOffset(1))},
			{24, 2, cents.Zero(), cents.MustFromInt(2400)},
			{-1, -1, t11, cents.MustFromInt(-1200).MustAdd(t11)},
			{-11, -1, just.MustOffset(1), cents.MustFromInt(-1200).MustAdd(just.MustOffset(1))},
			{-12, -1, cents.Zero(), cents.MustFromInt(-1200)},
			{-13, -2, t11, cents.MustFromInt(-2400).MustAdd(t11)},
			{-24, -2, cents.Zero(), cents.MustFromInt(-2400)},
		}

		for _, tc := range cases {
			oct, off, err := just.SplitStep(tc.step)
			require.NoError(t, err)
			assert.Equal(t, tc.wantOct, oct, "step=%d", tc.step)
			assert.True(t, off.Equal(tc.wantOff), "step=%d", tc.step)

			mOct, mOff := just.MustSplitStep(tc.step)
			assert.Equal(t, oct, mOct)
			assert.True(t, mOff.Equal(off))

			cum, err := just.CumulativeOffset(tc.step)
			require.NoError(t, err)
			assert.True(t, cum.Equal(tc.wantCum), "step=%d: got %v, want %v", tc.step, cum, tc.wantCum)
			assert.True(t, just.MustCumulativeOffset(tc.step).Equal(tc.wantCum))
		}

		// G-3: strict monotonicity and boundary step size across [-2N..2N]
		lastStep := cents.Octave().MustSub(t11)
		for k := -24; k <= 24; k++ {
			ck := just.MustCumulativeOffset(k)
			ckNext := just.MustCumulativeOffset(k + 1)
			assert.Less(t, ck.Compare(ckNext), 0, "c(%d) < c(%d)", k, k+1)
			if k%12 == 0 {
				ckPrev := just.MustCumulativeOffset(k - 1)
				diff := ck.MustSub(ckPrev)
				assert.True(t, diff.Equal(lastStep), "boundary step at k=%d", k)
			}
		}

		// S-21: non-12 temperament (19-EDO)
		edo19 := temperament.Equal19()
		assert.True(t, edo19.MustCumulativeOffset(19).Equal(cents.Octave()))
		assert.True(t, edo19.MustCumulativeOffset(-1).Equal(cents.MustStep(-1, 19)))
	})

	t.Run("OP-12 and G-9: DirectedInterval and CyclicInterval", func(t *testing.T) {
		t.Parallel()

		w3 := temperament.WerckmeisterIII()
		for _, k1 := range []int{-15, -12, -1, 0, 4, 7, 12, 19} {
			for _, k2 := range []int{-15, -12, -1, 0, 4, 7, 12, 19} {
				dir, err := w3.DirectedInterval(k1, k2)
				require.NoError(t, err)
				assert.True(t, dir.Equal(w3.MustDirectedInterval(k1, k2)))

				rev := w3.MustDirectedInterval(k2, k1)
				assert.True(t, dir.Equal(rev.Neg()), "antisymmetry for (%d, %d)", k1, k2)

				// Additivity with k3 = 5
				dir23 := w3.MustDirectedInterval(k2, 5)
				dir13 := w3.MustDirectedInterval(k1, 5)
				assert.True(t, dir.MustAdd(dir23).Equal(dir13), "additivity for (%d, %d, 5)", k1, k2)
			}
		}

		for m1 := range 12 {
			for m2 := range 12 {
				cyc, err := w3.CyclicInterval(m1, m2)
				require.NoError(t, err)
				assert.True(t, cyc.Equal(w3.MustCyclicInterval(m1, m2)))
				if m1 == m2 {
					assert.True(t, cyc.Equal(cents.Zero()), "G-9: CyclicInterval(%d, %d) == 0", m1, m2)
				} else {
					rev := w3.MustCyclicInterval(m2, m1)
					assert.True(t, cyc.MustAdd(rev).Equal(cents.Octave()), "CyclicInterval complement for (%d, %d)", m1, m2)
				}
			}
		}
	})

	t.Run("S-20: acoustic frequency calculation with cents.ApplyTo", func(t *testing.T) {
		t.Parallel()

		// For Equal12 with A4 (k=9) = 440 Hz, A5 (k=21) = 880 Hz, C4 (k=0) = 440 * 2^(-9/12)
		eq := temperament.Equal12()
		a5Delta := eq.MustDirectedInterval(9, 21)
		assert.InDelta(t, 880.0, a5Delta.MustApplyTo(440.0), 1e-12)

		// For Just with C4 (k=0) = 264 Hz, E4 (k=4, 5:4) = 330 Hz, G4 (k=7, 3:2) = 396 Hz, A4 (k=9, 5:3) = 440 Hz
		just := temperament.Just()
		assert.InDelta(t, 264.0, just.MustDirectedInterval(9, 0).MustApplyTo(440.0), 1e-9)
		assert.InDelta(t, 330.0, just.MustDirectedInterval(9, 4).MustApplyTo(440.0), 1e-9)
		assert.InDelta(t, 396.0, just.MustDirectedInterval(9, 7).MustApplyTo(440.0), 1e-9)
	})

	t.Run("OP-13: StepSizes, MinStepSize, MaxStepSize (S-14)", func(t *testing.T) {
		t.Parallel()

		just := temperament.Just()
		steps, err := just.StepSizes()
		require.NoError(t, err)
		require.Len(t, steps, 12)
		assert.Equal(t, steps, just.MustStepSizes())

		sum := cents.Zero()
		for _, s := range steps {
			assert.Greater(t, s.Sign(), 0)
			sum = sum.MustAdd(s)
		}
		assert.True(t, sum.Equal(cents.Octave()))

		// In asymmetric 5-limit Just Intonation (Just / Just5Limit), elementary steps take four values:
		// 16:15 at {0, 2, 4, 6, 7, 11}, 135:128 at {1, 5}, 25:24 at {3, 8, 10} (min), and 27:25 at {9} (max).
		s16_15 := cents.MustFromRatio(16, 15)
		s135_128 := cents.MustFromRatio(135, 128)
		s25_24 := cents.MustFromRatio(25, 24)
		s27_25 := cents.MustFromRatio(27, 25)
		wantJustSteps := [12]cents.Cents{
			s16_15, s135_128, s16_15, s25_24, s16_15, s135_128,
			s16_15, s16_15, s25_24, s27_25, s25_24, s16_15,
		}
		for m, want := range wantJustSteps {
			assert.True(t, steps[m].Equal(want), "Just step size at m=%d", m)
		}

		minStep, err := just.MinStepSize()
		require.NoError(t, err)
		assert.True(t, minStep.Equal(s25_24))
		assert.True(t, just.MustMinStepSize().Equal(minStep))

		maxStep, err := just.MaxStepSize()
		require.NoError(t, err)
		assert.True(t, maxStep.Equal(s27_25))
		assert.True(t, just.MustMaxStepSize().Equal(maxStep))

		// Defensive copy check on StepSizes()
		steps[0] = cents.MustFromInt(999)
		assert.True(t, just.MustStepSizes()[0].Equal(s16_15))

		// In inversionally symmetric Just5LimitSymmetric1 (m=10 is 16:9), elementary steps take three values:
		// 16:15 at {0, 2, 4, 6, 7, 9, 11} (max), 135:128 at {1, 5, 10}, and 25:24 at {3, 8} (min).
		sym1 := temperament.Just5LimitSymmetric1()
		sym1Steps := sym1.MustStepSizes()
		require.Len(t, sym1Steps, 12)
		wantSym1Steps := [12]cents.Cents{
			s16_15, s135_128, s16_15, s25_24, s16_15, s135_128,
			s16_15, s16_15, s25_24, s16_15, s135_128, s16_15,
		}
		for m, want := range wantSym1Steps {
			assert.True(t, sym1Steps[m].Equal(want), "Just5LimitSymmetric1 step size at m=%d", m)
		}
		assert.True(t, sym1.MustMinStepSize().Equal(s25_24))
		assert.True(t, sym1.MustMaxStepSize().Equal(s16_15))
	})

	t.Run("OP-14 and G-8: Rotate modal rotation and group laws (S-10)", func(t *testing.T) {
		t.Parallel()

		// YoungII narrows the six fifths C-G-D-A-E-B-F# by 1/6 Pc (S-5, K-7). Rotating its root to G (step 7)
		// relabels every pitch one fifth lower, so the narrowed fifths become F-C-G-D-A-E-B, which is exactly
		// the Vallotti scheme; with T(0) = 0 in both, the tables coincide (P-12 item 6 describes the converse
		// direction: Vallotti rotated by 7 differs from YoungII).
		young2 := temperament.YoungII()
		rotToG, err := young2.Rotate(7)
		require.NoError(t, err)
		assert.Equal(t, young2.Name(), rotToG.Name(), "Rotate preserves original Name()")
		assert.True(t, rotToG.Equal(temperament.Vallotti()), "YoungII rotated to G (step 7) must equal Vallotti")

		// Group laws on WerckmeisterIII
		w3 := temperament.WerckmeisterIII()
		assert.True(t, w3.MustRotate(0).Equal(w3))
		assert.True(t, w3.MustRotate(12).Equal(w3))
		assert.True(t, w3.MustRotate(-12).Equal(w3))
		assert.True(t, w3.MustRotate(-1).Equal(w3.MustRotate(11)))
		assert.True(t, w3.MustRotate(1203).Equal(w3.MustRotate(3)))

		for a := -5; a <= 7; a++ {
			for b := -5; b <= 7; b++ {
				lhs := w3.MustRotate(a).MustRotate(b)
				rhs := w3.MustRotate(a + b)
				assert.True(t, lhs.Equal(rhs), "group law failed for a=%d, b=%d", a, b)
			}
		}

		// Any N-EDO is invariant under rotation
		edo31 := temperament.Equal31()
		for r := -10; r <= 35; r++ {
			assert.True(t, edo31.MustRotate(r).Equal(edo31))
		}
	})
}

// TestComparisonAndPredicates verifies Equal, Near, IsEqual, and IsRational (OP-15, OP-16, S-15, S-16, K-18, K-19, G-11).
func TestComparisonAndPredicates(t *testing.T) {
	t.Parallel()

	t.Run("OP-15 and G-11: Equal and Near", func(t *testing.T) {
		t.Parallel()

		w3 := temperament.WerckmeisterIII()
		v := temperament.Vallotti()
		edo19 := temperament.Equal19()

		// Reflexivity and name independence
		assert.True(t, w3.Equal(w3))
		assert.True(t, w3.Equal(w3.MustWithName("Another Name")))
		assert.False(t, w3.Equal(v))
		assert.False(t, w3.Equal(edo19))

		// Near with negative tolerance -> false
		assert.False(t, w3.Near(w3, cents.MustFromInt(-1)))

		// Near with zero tolerance -> true iff Equal
		assert.True(t, w3.Near(w3, cents.Zero()))
		assert.False(t, w3.Near(v, cents.Zero()))

		// Near with different dimensions -> false
		assert.False(t, w3.Near(edo19, cents.MustFromInt(1000)))

		// Exact boundary tolerance max_m |T1(m) - T2(m)| between WerckmeisterIII and Vallotti
		maxDiff := cents.Zero()
		for m := range 12 {
			diff := w3.MustOffset(m).MustSub(v.MustOffset(m)).Abs()
			if diff.Compare(maxDiff) > 0 {
				maxDiff = diff
			}
		}
		assert.True(t, w3.Near(v, maxDiff), "Near must be true at exact maxDiff")
		assert.True(t, v.Near(w3, maxDiff), "Near symmetry")
		justBelow := maxDiff.MustSub(cents.MustFromFraction(1, 1000000))
		assert.False(t, w3.Near(v, justBelow), "Near must be false just below maxDiff")
	})

	t.Run("OP-16: IsEqual and IsRational across all 31 presets (K-19)", func(t *testing.T) {
		t.Parallel()

		for i, p := range temperament.Presets() {
			wantEqual := i < 6 // first 6 presets are N-EDO
			assert.Equal(t, wantEqual, p.IsEqual(), "IsEqual for %s", p.Name())

			// Rational 12-tone presets in catalog: Pythagorean (6), PythagoreanEFlatGSharp (7),
			// Just5Limit (8), Just5LimitSymmetric1 (9), Just5LimitSymmetric2 (10), Just7Limit (11), KirnbergerI (20).
			wantRational := (i >= 6 && i <= 11) || i == 20
			assert.Equal(t, wantRational, p.IsRational(), "IsRational for %s", p.Name())
		}
	})
}

// TestNearestStep verifies NearestStep and NearestCumulativeStep including boundary midpoints and tie-breaking (OP-17, S-17, K-20, G-1, G-7).
func TestNearestStep(t *testing.T) {
	t.Parallel()

	t.Run("Equal12 standard, boundary, and tie-breaking cases (S-17, G-7)", func(t *testing.T) {
		t.Parallel()

		eq := temperament.Equal12()

		// Pure major third 5:4 (~386.313714 cents) -> step 4 (400 cents), deviation = R(5:4) - 400 (~-13.686286 cents)
		r54 := cents.MustFromRatio(5, 4)
		step, dev, err := eq.NearestStep(r54)
		require.NoError(t, err)
		assert.Equal(t, 4, step)
		assert.True(t, dev.Equal(r54.MustSub(cents.MustFromInt(400))))

		// 1190 cents -> wraps to step 0 via upper octave 1200 cents, deviation = -10 cents, cumulative step = 12
		step, dev, err = eq.NearestStep(cents.MustFromInt(1190))
		require.NoError(t, err)
		assert.Equal(t, 0, step)
		assert.True(t, dev.Equal(cents.MustFromInt(-10)))

		cumStep, cumDev, err := eq.NearestCumulativeStep(cents.MustFromInt(1190))
		require.NoError(t, err)
		assert.Equal(t, 12, cumStep)
		assert.True(t, cumDev.Equal(cents.MustFromInt(-10)))

		// -10 cents -> step 0, deviation = -10 cents, cumulative step = 0
		step, dev = eq.MustNearestStep(cents.MustFromInt(-10))
		assert.Equal(t, 0, step)
		assert.True(t, dev.Equal(cents.MustFromInt(-10)))

		cumStep, cumDev = eq.MustNearestCumulativeStep(cents.MustFromInt(-10))
		assert.Equal(t, 0, cumStep)
		assert.True(t, cumDev.Equal(cents.MustFromInt(-10)))

		// -1210 cents -> step 0, deviation = -10 cents, cumulative step = -12
		cumStep, cumDev = eq.MustNearestCumulativeStep(cents.MustFromInt(-1210))
		assert.Equal(t, -12, cumStep)
		assert.True(t, cumDev.Equal(cents.MustFromInt(-10)))

		// G-7: octave boundary neighborhood
		milli := cents.MustFromFraction(1, 1000)
		step, dev = eq.MustNearestStep(milli.Neg())
		assert.Equal(t, 0, step)
		assert.True(t, dev.Equal(milli.Neg()))

		step, dev = eq.MustNearestStep(cents.Octave().MustSub(milli))
		assert.Equal(t, 0, step)
		assert.True(t, dev.Equal(milli.Neg()))

		for _, mult := range []int{-2, -1, 0, 1, 2} {
			x := cents.Octave().MustMul(mult)
			step, dev = eq.MustNearestStep(x)
			assert.Equal(t, 0, step)
			assert.True(t, dev.Equal(cents.Zero()))

			cumStep, cumDev = eq.MustNearestCumulativeStep(x)
			assert.Equal(t, 12*mult, cumStep)
			assert.True(t, cumDev.Equal(cents.Zero()))
		}

		// G-7: exact midpoints for all 12 steps m = 0..11 (including last step midpoint 1150 cents)
		for m := range 12 {
			mid := cents.MustFromInt(100*m + 50)
			step, dev = eq.MustNearestStep(mid)
			assert.Equal(t, m, step, "midpoint at %v must choose lower index %d", mid, m)
			assert.True(t, dev.Equal(cents.MustFromInt(50)), "midpoint at %v must have +50 deviation", mid)

			cumStep, cumDev = eq.MustNearestCumulativeStep(mid.MustAdd(cents.MustFromInt(1200)))
			assert.Equal(t, 12+m, cumStep)
			assert.True(t, cumDev.Equal(cents.MustFromInt(50)))
		}
	})

	t.Run("P-14: x exactly on a step T(m), in any octave, returns that step with zero deviation", func(t *testing.T) {
		t.Parallel()

		// P-14: x_mod = T(m) lies in the segment [T(m), T(m+1)) with d_low = 0 < d_high, so m* = m and delta = 0.
		// The same holds for x = T(m) + 1200*q in any octave q; NearestCumulativeStep returns k* = N*q + m.
		temps := append(temperament.Presets(), temperament.MustEDO(1), temperament.MustEDO(72),
			temperament.MustFromGenerator("Pythagorean 100", 100, cents.MustFromRatio(3, 2), -50))
		for _, tm := range temps {
			n := tm.StepsPerOctave()
			for m := range n {
				tmOffset := tm.MustOffset(m)
				for _, q := range []int{-3, -1, 0, 1, 2} {
					x := tmOffset.MustAdd(cents.Octave().MustMul(q))

					step, dev, err := tm.NearestStep(x)
					require.NoError(t, err)
					assert.Equal(t, m, step, "%s: NearestStep(T(%d)%+d*1200)", tm.Name(), m, q)
					assert.True(t, dev.Equal(cents.Zero()), "%s: deviation at T(%d)%+d*1200 is %v", tm.Name(), m, q, dev)

					cumStep, cumDev, err := tm.NearestCumulativeStep(x)
					require.NoError(t, err)
					assert.Equal(t, n*q+m, cumStep, "%s: NearestCumulativeStep(T(%d)%+d*1200)", tm.Name(), m, q)
					assert.True(t, cumDev.Equal(cents.Zero()), "%s: cumulative deviation at T(%d)%+d*1200 is %v", tm.Name(), m, q, cumDev)
				}
			}
		}

		// Explicit 12-EDO examples: 700 cents -> (7, 0); 1900 cents -> cumulative step 19.
		step, dev := temperament.Equal12().MustNearestStep(cents.MustFromInt(700))
		assert.Equal(t, 7, step)
		assert.True(t, dev.Equal(cents.Zero()))
		cumStep, cumDev := temperament.Equal12().MustNearestCumulativeStep(cents.MustFromInt(1900))
		assert.Equal(t, 19, cumStep)
		assert.True(t, cumDev.Equal(cents.Zero()))
	})

	t.Run("G-1 and G-7: 1-EDO midpoint at 600 cents", func(t *testing.T) {
		t.Parallel()

		edo1 := temperament.MustEDO(1)

		// At x = 600 cents (exact midpoint between T(0)=0 and T(1)=1200), lower candidate m=0 is chosen with +600 cents.
		step, dev := edo1.MustNearestStep(cents.MustFromInt(600))
		assert.Equal(t, 0, step)
		assert.True(t, dev.Equal(cents.MustFromInt(600)))

		cumStep, cumDev := edo1.MustNearestCumulativeStep(cents.MustFromInt(600))
		assert.Equal(t, 0, cumStep)
		assert.True(t, cumDev.Equal(cents.MustFromInt(600)))

		// At x = 601 cents, upper octave candidate T(1)=1200 is strictly closer -> step=0, dev=-599, cumStep=1.
		step, dev = edo1.MustNearestStep(cents.MustFromInt(601))
		assert.Equal(t, 0, step)
		assert.True(t, dev.Equal(cents.MustFromInt(-599)))

		cumStep, cumDev = edo1.MustNearestCumulativeStep(cents.MustFromInt(601))
		assert.Equal(t, 1, cumStep)
		assert.True(t, cumDev.Equal(cents.MustFromInt(-599)))
	})
}

// TestAnalyzeEDOAndIntervals verifies AnalyzeEDO and AnalyzeIntervals/AnalyzeFifths*/AnalyzeMajorThirds*/AnalyzeMinorThirds*
// against the reference table of all built-in presets in business-spec.md (OP-18, OP-19, S-18, S-19, K-21, K-22, G-9, G-10).
func TestAnalyzeEDOAndIntervals(t *testing.T) {
	t.Parallel()

	// Reference table from business-spec.md section "Сводная таблица акустических метрик встроенных темпераций"
	refTable := []struct {
		name           string
		fn             func() temperament.Temperament
		dMax           float64
		dMean          float64
		dRMS           float64
		pureFifths     int
		minFifth       float64
		maxFifth       float64
		fifthSpread    float64
		pureMajThirds  int
		minMajThird    float64
		maxMajThird    float64
		majThirdSpread float64
		wolfFifths15   []int
	}{
		{"12-EDO", temperament.Equal12, 0.0, 0.0, 0.0, 0, 700.000, 700.000, 0.000, 0, 400.000, 400.000, 0.000, nil},
		{"Pythagorean", temperament.Pythagorean, 11.730005, 5.865003, 6.819189, 11, 678.495, 701.955, 23.460, 0, 384.360, 407.820, 23.460, []int{6}},
		{"Pythagorean (Eb-G#)", temperament.PythagoreanEFlatGSharp, 15.640007, 6.842503, 8.332678, 11, 678.495, 701.955, 23.460, 0, 384.360, 407.820, 23.460, []int{8}},
		{"5-Limit Just Intonation", temperament.Just5Limit, 17.596288, 9.775858, 11.428444, 9, 680.449, 721.508, 41.059, 8, 386.314, 427.373, 41.059, []int{2, 6, 10}},
		{"5-Limit Just Intonation (Symmetric I)", temperament.Just5LimitSymmetric1, 15.641287, 8.635334, 10.299558, 9, 680.449, 721.508, 41.059, 7, 386.314, 427.373, 41.059, []int{2, 3, 6}},
		{"5-Limit Just Intonation (Symmetric II)", temperament.Just5LimitSymmetric2, 17.596288, 10.916382, 12.455431, 9, 680.449, 721.508, 41.059, 6, 364.807, 427.373, 62.565, []int{6, 7, 10}},
		{"7-Limit Just Intonation", temperament.Just7Limit, 31.174094, 14.464603, 17.408547, 5, 653.185, 736.931, 83.746, 6, 351.338, 462.348, 111.010, []int{2, 3, 6, 7, 10}},
		{"1/4-Comma Meantone", temperament.MeantoneQuarterComma, 27.372572, 11.975500, 14.583550, 0, 696.578, 737.637, 41.059, 8, 386.314, 427.373, 41.059, []int{8}},
		{"1/3-Comma Meantone", temperament.MeantoneThirdComma, 41.710099, 18.248168, 22.222292, 0, 694.786, 757.351, 62.565, 0, 379.145, 441.710, 62.565, []int{8}},
		{"1/5-Comma Meantone", temperament.MeantoneFifthComma, 18.770056, 8.211900, 10.000304, 0, 697.654, 725.809, 28.155, 0, 390.615, 418.770, 28.155, []int{8}},
		{"1/6-Comma Meantone", temperament.MeantoneSixthComma, 13.035046, 5.702833, 6.944807, 0, 698.371, 717.923, 19.553, 0, 393.482, 413.035, 19.553, []int{8}},
		{"2/7-Comma Meantone", temperament.MeantoneTwoSeventhComma, 33.517226, 14.663787, 17.857297, 0, 695.810, 746.086, 50.276, 0, 383.241, 433.517, 50.276, []int{8}},
		{"Werckmeister III", temperament.WerckmeisterIII, 11.730005, 6.842503, 7.738116, 8, 696.090, 701.955, 5.865, 0, 390.225, 407.820, 17.595, nil},
		{"Werckmeister IV", temperament.WerckmeisterIV, 17.595008, 8.145837, 9.742366, 5, 694.135, 709.775, 15.640, 0, 392.180, 415.640, 23.460, nil},
		{"Werckmeister V", temperament.WerckmeisterV, 7.820003, 2.443751, 3.338800, 6, 696.090, 707.820, 11.730, 0, 396.090, 407.820, 11.730, nil},
		{"Kirnberger I", temperament.KirnbergerI, 15.641287, 7.168763, 8.633678, 10, 680.449, 701.955, 21.506, 4, 386.314, 407.820, 21.506, []int{2}},
		{"Kirnberger II", temperament.KirnbergerII, 13.686286, 6.272668, 7.492935, 9, 691.202, 701.955, 10.753, 3, 386.314, 407.820, 21.506, nil},
		{"Kirnberger III", temperament.KirnbergerIII, 13.686286, 7.087358, 8.137490, 7, 696.578, 701.955, 5.377, 1, 386.314, 407.820, 21.506, nil},
		{"Vallotti", temperament.Vallotti, 9.775004, 4.235835, 5.233661, 6, 698.045, 701.955, 3.910, 0, 392.180, 407.820, 15.640, nil},
		{"Young I", temperament.YoungI, 8.309714, 4.154857, 5.172462, 4, 697.923, 701.955, 4.032, 0, 391.690, 407.820, 16.130, nil},
		{"Young II", temperament.YoungII, 11.730005, 5.865003, 6.819189, 6, 698.045, 701.955, 3.910, 0, 392.180, 407.820, 15.640, nil},
		{"Kellner", temperament.Kellner, 11.730005, 6.451503, 7.379103, 7, 697.263, 701.955, 4.692, 0, 389.052, 407.820, 18.768, nil},
		{"Neidhardt I", temperament.NeidhardtI, 7.820003, 4.561669, 5.172446, 4, 698.045, 701.955, 3.910, 0, 392.180, 403.910, 11.730, nil},
		{"Neidhardt II", temperament.NeidhardtII, 5.865003, 2.932501, 3.569327, 3, 698.045, 701.955, 3.910, 0, 394.135, 403.910, 9.775, nil},
		{"Neidhardt III", temperament.NeidhardtIII, 5.865003, 3.258335, 3.657472, 3, 698.045, 701.955, 3.910, 0, 394.135, 403.910, 9.775, nil},
		{"Neidhardt Dorf (1732)", temperament.NeidhardtDorf1732, 9.775004, 4.398752, 5.383653, 5, 696.090, 701.955, 5.865, 0, 390.225, 405.865, 15.640, nil},
	}

	for _, tc := range refTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			temp := tc.fn()

			// 1. AnalyzeEDO
			edoDev, err := temp.AnalyzeEDO()
			require.NoError(t, err)
			assertEDODeviationOracle(t, temp, edoDev)
			assert.InDelta(t, tc.dMax, edoDev.MaxAbsDeviation.Float64(), 1e-5)
			assert.InDelta(t, tc.dMean, edoDev.MeanAbsDeviation.Float64(), 1e-5)
			assert.InDelta(t, tc.dRMS, edoDev.RMSDeviation, 1e-5)
			assert.Equal(t, edoDev, temp.MustAnalyzeEDO())

			// 2. AnalyzeFifths
			fifths, err := temp.AnalyzeFifths()
			require.NoError(t, err)
			assert.Equal(t, 7, fifths.Step)
			assert.True(t, fifths.Reference.Equal(cents.MustFromRatio(3, 2)))
			assert.True(t, fifths.WolfThreshold.Equal(temperament.DefaultFifthWolfThreshold()))
			assertIntervalAnalysisOracle(t, temp, fifths)
			assert.Equal(t, tc.pureFifths, fifths.PureCount)
			assert.InDelta(t, tc.minFifth, fifths.Min.Float64(), 1e-2)
			assert.InDelta(t, tc.maxFifth, fifths.Max.Float64(), 1e-2)
			assert.InDelta(t, tc.fifthSpread, fifths.Spread.Float64(), 1e-2)
			assert.Equal(t, fifths, temp.MustAnalyzeFifths())

			if len(tc.wolfFifths15) == 0 {
				assert.Empty(t, fifths.WolfPositions)
			} else {
				assert.Equal(t, tc.wolfFifths15, fifths.WolfPositions)
			}

			// Sum of all 12 cyclic fifths must equal 7 * 1200 = 8400 cents
			sumFifths := cents.Zero()
			for _, sz := range fifths.Sizes {
				sumFifths = sumFifths.MustAdd(sz)
			}
			assert.True(t, sumFifths.Equal(cents.MustFromInt(8400)))

			// 3. AnalyzeMajorThirds
			maj3, err := temp.AnalyzeMajorThirds()
			require.NoError(t, err)
			assert.Equal(t, 4, maj3.Step)
			assert.True(t, maj3.Reference.Equal(cents.MustFromRatio(5, 4)))
			assert.True(t, maj3.WolfThreshold.Equal(temperament.DefaultThirdWolfThreshold()))
			assertIntervalAnalysisOracle(t, temp, maj3)
			assert.Equal(t, tc.pureMajThirds, maj3.PureCount)
			assert.InDelta(t, tc.minMajThird, maj3.Min.Float64(), 1e-2)
			assert.InDelta(t, tc.maxMajThird, maj3.Max.Float64(), 1e-2)
			assert.InDelta(t, tc.majThirdSpread, maj3.Spread.Float64(), 1e-2)
			assert.Equal(t, maj3, temp.MustAnalyzeMajorThirds())

			// 4. AnalyzeMinorThirds and triad identity I_4(m) + I_3((m+4)%12) == I_7(m)
			min3, err := temp.AnalyzeMinorThirds()
			require.NoError(t, err)
			assert.Equal(t, 3, min3.Step)
			assert.True(t, min3.Reference.Equal(cents.MustFromRatio(6, 5)))
			assert.True(t, min3.WolfThreshold.Equal(temperament.DefaultThirdWolfThreshold()))
			assertIntervalAnalysisOracle(t, temp, min3)
			assert.Equal(t, min3, temp.MustAnalyzeMinorThirds())
			for m := range 12 {
				triadFifth := maj3.Sizes[m].MustAdd(min3.Sizes[(m+4)%12])
				assert.True(t, triadFifth.Equal(fifths.Sizes[m]), "triad identity at step %d", m)
			}
		})
	}

	t.Run("Non-12 EDO presets have zero EDO deviation", func(t *testing.T) {
		t.Parallel()

		for _, fn := range []func() temperament.Temperament{
			temperament.Equal19,
			temperament.Equal22,
			temperament.Equal24,
			temperament.Equal31,
			temperament.Equal53,
		} {
			p := fn()
			dev := p.MustAnalyzeEDO()
			assertEDODeviationOracle(t, p, dev)
			assert.True(t, dev.MaxAbsDeviation.Equal(cents.Zero()))
			assert.Equal(t, 0, dev.MaxPosition)
			assert.True(t, dev.MeanAbsDeviation.Equal(cents.Zero()))
			assert.Equal(t, 0.0, dev.RMSDeviation)
		}
	})

	t.Run("WithThreshold variants and G-9 boundary steps s=1 and s=N-1", func(t *testing.T) {
		t.Parallel()

		// Kirnberger I has 1 wolf fifth at threshold 15 cents (D-A = 40:27, dev ~ -21.506 cents),
		// but 0 wolf fifths at threshold 22 cents.
		k1 := temperament.KirnbergerI()
		f22, err := k1.AnalyzeFifthsWithThreshold(cents.MustFromInt(22))
		require.NoError(t, err)
		assertIntervalAnalysisOracle(t, k1, f22)
		assert.Empty(t, f22.WolfPositions)
		assert.Empty(t, k1.MustAnalyzeFifthsWithThreshold(cents.MustFromInt(22)).WolfPositions)

		majTh, err := k1.AnalyzeMajorThirdsWithThreshold(cents.MustFromInt(10))
		require.NoError(t, err)
		assertIntervalAnalysisOracle(t, k1, majTh)
		assert.NotEmpty(t, majTh.WolfPositions)
		assert.Equal(t, majTh.WolfPositions, k1.MustAnalyzeMajorThirdsWithThreshold(cents.MustFromInt(10)).WolfPositions)

		minTh, err := k1.AnalyzeMinorThirdsWithThreshold(cents.MustFromInt(10))
		require.NoError(t, err)
		assertIntervalAnalysisOracle(t, k1, minTh)
		assert.NotEmpty(t, minTh.WolfPositions)
		assert.Equal(t, minTh.WolfPositions, k1.MustAnalyzeMinorThirdsWithThreshold(cents.MustFromInt(10)).WolfPositions)

		// G-9: AnalyzeIntervals at s=1 matches StepSizes, and at s=N-1 complements StepSizes to 1200 cents
		stepSizes := k1.MustStepSizes()
		a1, err := k1.AnalyzeIntervals(1, cents.MustFromInt(100), cents.MustFromInt(20))
		require.NoError(t, err)
		assertIntervalAnalysisOracle(t, k1, a1)
		assert.Equal(t, a1.Sizes, k1.MustAnalyzeIntervals(1, cents.MustFromInt(100), cents.MustFromInt(20)).Sizes)
		a11, err := k1.AnalyzeIntervals(11, cents.MustFromInt(1100), cents.MustFromInt(20))
		require.NoError(t, err)
		assertIntervalAnalysisOracle(t, k1, a11)

		for m := range 12 {
			assert.True(t, a1.Sizes[m].Equal(stepSizes[m]))
			prevStep := stepSizes[(m+11)%12]
			assert.True(t, a11.Sizes[m].Equal(cents.Octave().MustSub(prevStep)))
		}
	})

	t.Run("G-10: Defensive copying of slices in EDODeviation and IntervalAnalysis", func(t *testing.T) {
		t.Parallel()

		tm := temperament.MeantoneQuarterComma()

		edo1 := tm.MustAnalyzeEDO()
		origDev0 := edo1.Deviations[0]
		edo1.Deviations[0] = cents.MustFromInt(999)
		edo2 := tm.MustAnalyzeEDO()
		assert.True(t, edo2.Deviations[0].Equal(origDev0), "mutating EDODeviation.Deviations must not affect subsequent AnalyzeEDO calls")
		assert.True(t, tm.MustOffset(0).Equal(cents.Zero()))

		ia1 := tm.MustAnalyzeFifths()
		require.NotEmpty(t, ia1.WolfPositions)
		origSize0 := ia1.Sizes[0]
		origIaDev0 := ia1.Deviations[0]
		origWolf0 := ia1.WolfPositions[0]

		ia1.Sizes[0] = cents.MustFromInt(999)
		ia1.Deviations[0] = cents.MustFromInt(999)
		ia1.WolfPositions[0] = 99

		ia2 := tm.MustAnalyzeFifths()
		assert.True(t, ia2.Sizes[0].Equal(origSize0), "mutating IntervalAnalysis.Sizes must not affect subsequent calls")
		assert.True(t, ia2.Deviations[0].Equal(origIaDev0), "mutating IntervalAnalysis.Deviations must not affect subsequent calls")
		assert.Equal(t, origWolf0, ia2.WolfPositions[0], "mutating IntervalAnalysis.WolfPositions must not affect subsequent calls")
	})
}

// TestErrorScenariosAndZeroValue verifies all error scenarios E-1..E-6, priority of ErrInvalidArgument over ErrPrimeLimit,
// zero-value behavior G-12 without panics, and Must... panics (K-24, NF-5).
func TestErrorScenariosAndZeroValue(t *testing.T) {
	t.Parallel()

	t.Run("E-1: New, FromOffsets, FromFloats validation errors", func(t *testing.T) {
		t.Parallel()

		// Empty slice
		_, err := temperament.New("Empty", nil)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromOffsets("Empty", []cents.Cents{})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Exceeds MaxStepsPerOctave
		tooMany := make([]cents.Cents, temperament.MaxStepsPerOctave+1)
		_, err = temperament.New("TooMany", tooMany)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Non-zero T(0)
		_, err = temperament.New("NonZeroRoot", []cents.Cents{cents.MustFromInt(10), cents.MustFromInt(200)})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Non-monotonic (duplicate or decreasing)
		_, err = temperament.New("Duplicate", []cents.Cents{cents.Zero(), cents.MustFromInt(300), cents.MustFromInt(300)})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.New("Decreasing", []cents.Cents{cents.Zero(), cents.MustFromInt(500), cents.MustFromInt(400)})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// T(N-1) >= 1200 cents
		_, err = temperament.New("AtOctave", []cents.Cents{cents.Zero(), cents.Octave()})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.New("AboveOctave", []cents.Cents{cents.Zero(), cents.MustFromInt(1250)})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// FromFloats with NaN, +Inf, -Inf, and non-monotonic input
		_, err = temperament.FromFloats("NaN", []float64{0, math.NaN()})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromFloats("PosInf", []float64{0, math.Inf(1)})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromFloats("NegInf", []float64{0, math.Inf(-1)})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromFloats("NonMonotonicFloat", []float64{0, 200, 150})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
	})

	t.Run("E-1: FromFloats octave bounds, values beyond the cents range, and grid collisions", func(t *testing.T) {
		t.Parallel()

		// OP-5 and E-1 classify every table violation (T(0) != 0, T(m) <= T(m-1), T(m) >= 1200) as OA.
		// The architecture error priority OA -> OP -> OZ -> OD (P-7) applies when an input value is also outside
		// the range of package cents (|c| > 9e6 cents): the table violation is reported as ErrInvalidArgument,
		// not as ErrOutOfRange.
		invalid := []struct {
			name    string
			offsets []float64
		}{
			{"AtOctave", []float64{0, 1200}},
			{"JustAboveOctave", []float64{0, 1200.000001}},
			{"WellAboveOctave", []float64{0, 1e6}},
			{"BeyondCentsRangeAboveOctave", []float64{0, 1e7}},
			{"BeyondCentsRangeNegative", []float64{0, -1e7}},
			{"BeyondCentsRangeAtRoot", []float64{1e7, 100}},
			{"NonZeroRootAndBeyondCentsRange", []float64{5, 1e7}},
			{"NegativeStep", []float64{0, -100}},
			{"NonZeroRoot", []float64{0.000001, 100}},
			// P-6: two distinct float64 inputs rounding to the same 1e-6 cent grid point violate strict monotonicity.
			{"GridCollision", []float64{0, 100, 100.0000004}},
		}
		for _, tc := range invalid {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				_, err := temperament.FromFloats(tc.name, tc.offsets)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument)
				assert.NotErrorIs(t, err, temperament.ErrOutOfRange)
				assert.Panics(t, func() { temperament.MustFromFloats(tc.name, tc.offsets) })
			})
		}

		// Largest grid value below the octave is accepted (IV-11 bound is strict at 1200 cents only).
		edge, err := temperament.FromFloats("Edge", []float64{0, 1199.999999})
		require.NoError(t, err)
		assertValidTemperamentInvariants(t, edge, "Edge", 2)
		assert.True(t, edge.MustOffset(1).Equal(cents.MustFromFraction(1199999999, 1_000_000)))

		// E-1: an empty table and a table longer than MaxStepsPerOctave are OA.
		_, err = temperament.FromFloats("Empty", nil)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		maxSteps := make([]float64, temperament.MaxStepsPerOctave)
		for m := range maxSteps {
			maxSteps[m] = float64(m) * 1200 / temperament.MaxStepsPerOctave
		}
		atMax, err := temperament.FromFloats("AtMax", maxSteps)
		require.NoError(t, err)
		assert.Equal(t, temperament.MaxStepsPerOctave, atMax.StepsPerOctave())
		tooMany := append(slices.Clone(maxSteps), 1199.99999)
		_, err = temperament.FromFloats("TooMany", tooMany)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromFloats("TooManyZeros", make([]float64, temperament.MaxStepsPerOctave+1))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Negative zero converts to 0 cents (cents P-6) and is a valid root.
		negZero, err := temperament.FromFloats("NegZero", []float64{math.Copysign(0, -1), 700})
		require.NoError(t, err)
		assert.True(t, negZero.MustOffset(0).Equal(cents.Zero()))
	})

	t.Run("P-6: FromFloats checks T(0) = 0 after the conversion to cents", func(t *testing.T) {
		t.Parallel()

		// P-6: IV-11 is checked on the values after cents.FromFloat. Inputs smaller than half of the 1e-6 cent grid
		// step convert to exactly 0 cents (cents P-6, Г-6), so they are a valid root T(0) = 0.
		for _, root := range []float64{1e-9, -1e-9, 4e-7, -4e-7, 4.999e-7, math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64} {
			tm, err := temperament.FromFloats("TinyRoot", []float64{root, 100})
			if !assert.NoError(t, err, "root=%g", root) {
				continue
			}
			assertValidTemperamentInvariants(t, tm, "TinyRoot", 2)
			assert.True(t, tm.MustOffset(1).Equal(cents.MustFromInt(100)), "root=%g", root)
			assert.True(t, tm.Equal(temperament.MustFromFloats("Zero", []float64{0, 100})), "root=%g", root)
		}
		// Half of the grid step rounds away from zero to +/-1e-6 cent, which is not a valid root.
		for _, root := range []float64{5e-7, -5e-7, 1e-6} {
			_, err := temperament.FromFloats("GridRoot", []float64{root, 100})
			require.ErrorIs(t, err, temperament.ErrInvalidArgument, "root=%g", root)
		}
		// A tiny non-root value also converts to 0 cents and therefore duplicates T(0).
		_, err := temperament.FromFloats("TinyStep", []float64{0, 1e-9})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
	})

	t.Run("E-2: EDO and FromGenerator dimension bounds", func(t *testing.T) {
		t.Parallel()

		for _, badN := range []int{-5, 0, temperament.MaxStepsPerOctave + 1, 100000} {
			_, err := temperament.EDO(badN)
			require.ErrorIs(t, err, temperament.ErrInvalidArgument)
			_, err = temperament.FromGenerator("BadN", badN, cents.MustFromInt(700), 0)
			require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		}
	})

	t.Run("E-3 and P-7: FromRatios errors and priority of ErrInvalidArgument before ErrPrimeLimit", func(t *testing.T) {
		t.Parallel()

		// Empty slice
		_, err := temperament.FromRatios("Empty", nil)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Non-positive numerator or denominator
		_, err = temperament.FromRatios("ZeroNum", []temperament.Ratio{{0, 1}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromRatios("NegDen", []temperament.Ratio{{1, 1}, {3, -2}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// First ratio != 1:1
		_, err = temperament.FromRatios("BadRoot", []temperament.Ratio{{9, 8}, {3, 2}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Ratio < 1 or >= 2
		_, err = temperament.FromRatios("BelowOne", []temperament.Ratio{{1, 1}, {3, 4}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromRatios("AtTwo", []temperament.Ratio{{1, 1}, {2, 1}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromRatios("AboveTwo", []temperament.Ratio{{1, 1}, {5, 2}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Non-increasing ratios
		_, err = temperament.FromRatios("EqualRatios", []temperament.Ratio{{1, 1}, {3, 2}, {6, 4}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromRatios("DecreasingRatios", []temperament.Ratio{{1, 1}, {3, 2}, {4, 3}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Prime factor > 31 in the reduced ratio (E-3: e.g. the primes 37, 131, and 139) -> ErrPrimeLimit
		_, err = temperament.FromRatios("Prime37", []temperament.Ratio{{1, 1}, {37, 32}})
		require.ErrorIs(t, err, temperament.ErrPrimeLimit)
		_, err = temperament.FromRatios("Prime139", []temperament.Ratio{{1, 1}, {139, 128}})
		require.ErrorIs(t, err, temperament.ErrPrimeLimit)
		_, err = temperament.FromRatios("Prime131", []temperament.Ratio{{1, 1}, {131, 128}, {3, 2}})
		require.ErrorIs(t, err, temperament.ErrPrimeLimit)
		_, err = temperament.FromRatios("Prime131Denominator", []temperament.Ratio{{1, 1}, {196, 131}})
		require.ErrorIs(t, err, temperament.ErrPrimeLimit)

		// Priority check (P-7): prime > 31 at index 1 AND invalid argument (non-monotonic or Den<=0) at index 2
		// must return ErrInvalidArgument, NOT ErrPrimeLimit!
		_, err = temperament.FromRatios("Priority1", []temperament.Ratio{{1, 1}, {37, 32}, {0, 1}})
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		assert.NotErrorIs(t, err, temperament.ErrPrimeLimit)

		_, err = temperament.FromRatios("Priority2", []temperament.Ratio{{1, 1}, {37, 32}, {9, 8}}) // 37/32 = 1.15625 > 9/8 = 1.125
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		assert.NotErrorIs(t, err, temperament.ErrPrimeLimit)
	})

	t.Run("E-4: Pythagorean, Meantone, Circle-of-Fifths, and Generator validation errors", func(t *testing.T) {
		t.Parallel()

		// NewPythagorean / NewPythagoreanWithWolf out of bounds
		_, err := temperament.NewPythagorean(-12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewPythagorean(1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewPythagoreanWithWolf(-1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewPythagoreanWithWolf(12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// NewMeantone* invalid commaDen <= 0, kMin out of [-11, 0], or fifth outside (4800/7, 720)
		_, err = temperament.NewMeantone(1, 0)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewMeantone(1, -4)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewMeantone(1, 1) // 1-comma meantone: g ~ 680.449 < 4800/7 ~ 685.714
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewMeantone(-1, 1) // -1-comma meantone: g ~ 723.461 > 720
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewMeantoneWithRange(1, 4, -12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewMeantoneWithRange(1, 4, 1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.NewMeantoneFromFifthWithRange(cents.MustFromInt(700), -12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// FromFifths* / FromFifthDeviations* wrong length, bad kMin, unclosed 12-fifth circle, or non-monotonic result
		_, err = temperament.FromFifths("Short", make([]cents.Cents, 10))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromFifthDeviations("Long", make([]cents.Cents, 13))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// 12 pure fifths do not close the circle (sum = 12*R(3:2) != 8400 cents)
		pure12 := make([]cents.Cents, 12)
		for i := range 12 {
			pure12[i] = cents.MustFromRatio(3, 2)
		}
		_, err = temperament.FromFifths("Unclosed", pure12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		zeroDevs12 := make([]cents.Cents, 12)
		_, err = temperament.FromFifthDeviations("UnclosedDevs", zeroDevs12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Bad kMin in FromFifthsWithStart / FromFifthDeviationsWithStart
		_, err = temperament.FromFifthsWithStart("BadKMin", pure12[:11], -12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromFifthDeviationsWithStart("BadKMin", zeroDevs12[:11], 1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// Degenerate 11 fifths (all 600 cents -> duplicate offsets in [0, 1200))
		deg11 := make([]cents.Cents, 11)
		for i := range 11 {
			deg11[i] = cents.MustFromInt(600)
		}
		_, err = temperament.FromFifths("Degenerate", deg11)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// FromGenerator with kMin out of [-(N-1), 0] or generator producing duplicates (e.g. g = 600 cents for N = 12)
		_, err = temperament.FromGenerator("BadKMinLow", 12, cents.MustFromInt(700), -12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromGenerator("BadKMinHigh", 12, cents.MustFromInt(700), 1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = temperament.FromGenerator("DuplicatePoints", 12, cents.MustFromInt(600), 0)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
	})

	t.Run("E-4 and G-5: meantone fifth outside (4800/7, 720) cents is rejected even when congruent modulo the octave", func(t *testing.T) {
		t.Parallel()

		// OP-4 and the architecture contract of NewMeantone*/NewMeantoneFromFifth*: the fifth generator g itself
		// must lie in (4800/7, 720) cents; there is no reduction of g modulo 1200 cents in these constructors.
		// A fifth such as -500 cents (congruent to 700) or 1900 cents (congruent to 700) is not a meantone fifth.
		lowBound := cents.MustFromFraction(4800, 7)
		highBound := cents.MustFromInt(720)
		fifths := []struct {
			name  string
			fifth cents.Cents
		}{
			{"MinusFiveHundred", cents.MustFromInt(-500)},
			{"NineteenHundred", cents.MustFromInt(1900)},
			{"SevenHundredPlusOctave", cents.MustFromInt(700 + 1200)},
			{"SevenHundredMinusOctave", cents.MustFromInt(700 - 1200)},
			{"LowBoundMinusOctave", lowBound.MustSub(cents.Octave())},
			{"HighBoundPlusOctave", highBound.MustAdd(cents.Octave())},
			{"Zero", cents.Zero()},
			{"Octave", cents.Octave()},
		}
		for _, tc := range fifths {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				_, err := temperament.NewMeantoneFromFifth(tc.fifth)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument)
				for kMin := -11; kMin <= 0; kMin++ {
					_, err = temperament.NewMeantoneFromFifthWithRange(tc.fifth, kMin)
					require.ErrorIs(t, err, temperament.ErrInvalidArgument, "kMin=%d", kMin)
				}
				assert.Panics(t, func() { temperament.MustNewMeantoneFromFifth(tc.fifth) })
			})
		}

		// Comma fractions whose fifth g = R(3:2) - (num/den)*Sc leaves (4800/7, 720) cents, including fractions
		// with g congruent to a valid meantone fifth modulo 1200 cents (56 commas: g ~ -502.398 ~ 697.602 mod 1200;
		// -56 commas: g ~ 1906.307 ~ 706.307 mod 1200) and huge fractions whose intermediate product exceeds the
		// range of package cents. All are OA by OP-4; by the priority OA -> OP -> OZ -> OD (P-7) the out-of-range
		// intermediate value does not turn the error into ErrOutOfRange.
		fractions := []struct {
			name     string
			num, den int
		}{
			{"FiftySixCommas", 56, 1},
			{"MinusFiftySixCommas", -56, 1},
			{"OneHundredTwelveHalfCommas", 112, 2},
			{"OneMillionCommas", 1_000_000, 1},
			{"MinusOneMillionCommas", -1_000_000, 1},
			{"MaxIntCommas", math.MaxInt, 1},
			{"MinIntThirdCommas", math.MinInt, 3},
			{"MinIntCommas", math.MinInt, 1},
			{"MaxIntOverMaxIntMinusOne", math.MaxInt, math.MaxInt - 1},
			// |alpha| < 1 outside the existence interval (-0.83905683..., +0.75516119...) of E-4 and G-5.
			{"FourFifthsComma", 4, 5},
			{"MinusNineTenthsComma", -9, 10},
			{"SixteenTwentiethsComma", 16, 20},
			{"MinusSixSeventhsComma", -6, 7},
		}
		for _, tc := range fractions {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				_, err := temperament.NewMeantone(tc.num, tc.den)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument)
				assert.NotErrorIs(t, err, temperament.ErrOutOfRange)
				_, err = temperament.NewMeantoneWithRange(tc.num, tc.den, temperament.DefaultMeantoneMinFifth)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument)
				assert.NotErrorIs(t, err, temperament.ErrOutOfRange)
				_, err = temperament.NewMeantoneWithRange(tc.num, tc.den, -11)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument)
				assert.NotErrorIs(t, err, temperament.ErrOutOfRange)
				assert.Panics(t, func() { temperament.MustNewMeantone(tc.num, tc.den) })
			})
		}

		// Fractions with |alpha| < 1 just inside the existence interval of G-5 are valid regular meantones.
		for _, f := range [][2]int{{3, 4}, {755, 1000}, {-5, 6}, {-839, 1000}} {
			tm, err := temperament.NewMeantone(f[0], f[1])
			require.NoError(t, err, "alpha=%d/%d", f[0], f[1])
			assertValidTemperamentInvariants(t, tm, tm.Name(), 12)
		}

		// MaxInt/MaxInt commas is exactly one syntonic comma (g = 40/27 ~ 680.449 < 4800/7): OA.
		_, err := temperament.NewMeantone(math.MaxInt, math.MaxInt)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		// MaxInt/MaxInt * 1/4: a reducible fraction equal to 1/4 is a valid 1/4-comma meantone (OP-4 reduces by gcd).
		quarter, err := temperament.NewMeantone(math.MaxInt/7*1, math.MaxInt/7*4)
		require.NoError(t, err)
		assert.True(t, quarter.Equal(temperament.MeantoneQuarterComma()))
		assert.Equal(t, "1/4-Comma Meantone", quarter.Name())
	})

	t.Run("OP-8: FromGenerator reduces any generator modulo the octave", func(t *testing.T) {
		t.Parallel()

		// P-9: points P_k = (k*g) mod 1200 for any g in Cents. With g = -500 cents (congruent to 700 cents)
		// and N = 12 the twelve points are the multiples of 100 cents, i.e. 12-EDO.
		for _, g := range []int{-500, 700, 1900, -1700} {
			gen, err := temperament.FromGenerator("Generated", 12, cents.MustFromInt(g), -3)
			require.NoError(t, err, "g=%d", g)
			assert.True(t, gen.Equal(temperament.Equal12()), "g=%d", g)
		}
	})

	t.Run("E-5 and E-6: out-of-bounds indices, invalid analysis parameters, and N != 12 checks", func(t *testing.T) {
		t.Parallel()

		eq12 := temperament.Equal12()
		_, err := eq12.Offset(-1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = eq12.Offset(12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		_, err = eq12.CyclicInterval(-1, 5)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = eq12.CyclicInterval(0, 12)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// AnalyzeIntervals invalid step or non-positive wolfThreshold
		ref := cents.MustFromRatio(3, 2)
		th := cents.MustFromInt(15)
		_, err = eq12.AnalyzeIntervals(0, ref, th)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = eq12.AnalyzeIntervals(12, ref, th)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = eq12.AnalyzeIntervals(7, ref, cents.Zero())
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = eq12.AnalyzeIntervals(7, ref, cents.MustFromInt(-5))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// AnalyzeIntervals on N=1 -> ErrInvalidArgument
		edo1 := temperament.MustEDO(1)
		_, err = edo1.AnalyzeIntervals(1, ref, th)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// E-6: 12-tone helpers on N=19 -> ErrInvalidArgument
		edo19 := temperament.Equal19()
		_, err = edo19.AnalyzeFifths()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = edo19.AnalyzeFifthsWithThreshold(th)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = edo19.AnalyzeMajorThirds()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = edo19.AnalyzeMajorThirdsWithThreshold(th)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = edo19.AnalyzeMinorThirds()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = edo19.AnalyzeMinorThirdsWithThreshold(th)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)

		// ErrOutOfRange on huge step index in CumulativeOffset and DirectedInterval. The architecture ("Errors":
		// relation to errors of package cents) requires the chain to hold both the temperament error and the cents
		// error of the same class, whatever way the overflow is detected.
		_, err = eq12.CumulativeOffset(12 * 10000) // 10000 octaves = 12,000,000 cents > 9,000,000
		require.ErrorIs(t, err, temperament.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
		_, err = eq12.CumulativeOffset(-12*7500 - 1) // just below -9,000,000 cents
		require.ErrorIs(t, err, temperament.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
		edge, err := eq12.CumulativeOffset(-12 * 7500) // exactly -9,000,000 cents is inside the cents range
		require.NoError(t, err)
		assert.True(t, edge.Equal(cents.MustFromInt(-9_000_000)))
		_, err = eq12.DirectedInterval(-12*5000, 12*5000)
		require.ErrorIs(t, err, temperament.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)

		// OP-11, OP-12 and the error table of the architecture: c(k) = 1200*q + T(m) and every intermediate c(k)
		// of DirectedInterval must lie in the cents range. Just5Limit has non-zero remainders T(1) = 16:15 and
		// T(11) = 15:8, so the octave shift alone is in range while the full offset is not:
		// c(90000) = 9 000 000 (in range), c(90001) = 9 000 000 + T(1) (out of range),
		// c(-90000) = -9 000 000 (in range), c(-90001) = -9 001 200 + T(11) ~ -9 000 111.73 (out of range).
		just := temperament.Just5Limit()
		top, err := just.CumulativeOffset(90000)
		require.NoError(t, err)
		assert.True(t, top.Equal(cents.MustFromInt(9_000_000)))
		bottom, err := just.CumulativeOffset(-90000)
		require.NoError(t, err)
		assert.True(t, bottom.Equal(cents.MustFromInt(-9_000_000)))
		for _, k := range []int{90001, 90011, 90012, -90001, -90011, -90012, math.MaxInt, math.MinInt} {
			_, err = just.CumulativeOffset(k)
			require.ErrorIs(t, err, temperament.ErrOutOfRange, "k=%d", k)
			assert.ErrorIs(t, err, cents.ErrOutOfRange, "k=%d", k)
		}

		// A directed interval with one or both bounds outside the cents range is OD even when the interval itself
		// would be small (e.g. 90012 -> 90013 is one 12-EDO semitone between two out-of-range offsets).
		outOfRangePairs := []struct {
			tm       temperament.Temperament
			from, to int
		}{
			{eq12, 90012, 90013},
			{eq12, 90013, 90012},
			{eq12, 90012, 0},
			{eq12, 0, 90012},
			{eq12, -90012, 0},
			{eq12, 0, -90012},
			{eq12, 90001, 90001},
			{just, 90001, 90000},
			{just, 90000, 90001},
			{just, -90000, -90001},
			{just, -90001, -90000},
			{just, 1, 90001},
			{just, -90001, 0},
			{just, math.MinInt, math.MaxInt},
			// Both bounds in range, interval out of range: c(-89999) = -9 000 000 + T(1) and
			// c(89999) = 8 998 800 + T(11), so the interval is ~17 999 776.54 cents.
			{just, -89999, 89999},
			{just, 89999, -89999},
			// Octave shift of exactly 9 000 000 cents plus a non-zero remainder: c(2) - c(-89999) =
			// 9 000 000 + T(2) - T(1) ~ 9 000 092.18 cents.
			{just, -89999, 2},
			{just, 2, -89999},
		}
		for _, tc := range outOfRangePairs {
			_, err = tc.tm.DirectedInterval(tc.from, tc.to)
			require.ErrorIs(t, err, temperament.ErrOutOfRange, "%s: DirectedInterval(%d, %d)", tc.tm.Name(), tc.from, tc.to)
			assert.ErrorIs(t, err, cents.ErrOutOfRange, "%s: DirectedInterval(%d, %d)", tc.tm.Name(), tc.from, tc.to)
		}

		// Both bounds at the edges of the range are fine as long as the interval itself is in range.
		for _, tc := range []struct {
			from, to int
			want     cents.Cents
		}{
			{90000, 89999, just.MustOffset(11).MustSub(cents.Octave())},
			{-90000, -89999, just.MustOffset(1)},
			{0, 90000, cents.MustFromInt(9_000_000)},
			{-90000, 0, cents.MustFromInt(9_000_000)},
			{90000, 90000, cents.Zero()},
			// c(1) - c(-89999) = T(1) - (-9 000 000 + T(1)) is exactly the edge 9 000 000 cents.
			{-89999, 1, cents.MustFromInt(9_000_000)},
			{1, -89999, cents.MustFromInt(-9_000_000)},
		} {
			got, err := just.DirectedInterval(tc.from, tc.to)
			require.NoError(t, err, "DirectedInterval(%d, %d)", tc.from, tc.to)
			assert.True(t, got.Equal(tc.want), "DirectedInterval(%d, %d) = %v, want %v", tc.from, tc.to, got, tc.want)
		}

		// ErrTooComplex when exact cents denominator overflows 10^30
		bigDen := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
		maxComplexityOffset := cents.MustFromComposition(big.NewInt(1), bigDen)
		complexTemp := temperament.MustNew("MaxComplexity", []cents.Cents{cents.Zero(), maxComplexityOffset})
		_, err = complexTemp.AnalyzeEDO()
		require.ErrorIs(t, err, temperament.ErrTooComplex)
	})

	t.Run("G-12: zero-value Temperament{} never panics on non-Must methods", func(t *testing.T) {
		t.Parallel()

		var zero temperament.Temperament
		assert.False(t, zero.IsValid())
		assert.Equal(t, 0, zero.StepsPerOctave())
		assert.Equal(t, "", zero.Name())
		assert.Nil(t, zero.Offsets())
		assert.Equal(t, "Temperament(<invalid>)", zero.String())
		assert.False(t, zero.IsEqual())
		assert.False(t, zero.IsRational())
		assert.True(t, zero.Equal(temperament.Temperament{}))
		assert.False(t, zero.Equal(temperament.Equal12()))
		assert.False(t, temperament.Equal12().Equal(zero))
		assert.False(t, zero.Near(zero, cents.Zero()))
		assert.False(t, zero.Near(temperament.Equal12(), cents.MustFromInt(100)))

		_, err := zero.WithName("X")
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.Offset(0)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, _, err = zero.SplitStep(0)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.CumulativeOffset(0)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.DirectedInterval(0, 1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.CyclicInterval(0, 1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.StepSizes()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.MinStepSize()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.MaxStepSize()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.Rotate(1)
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, _, err = zero.NearestStep(cents.Zero())
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, _, err = zero.NearestCumulativeStep(cents.Zero())
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeEDO()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeIntervals(1, cents.MustFromInt(100), cents.MustFromInt(10))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeFifths()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeFifthsWithThreshold(cents.MustFromInt(15))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeMajorThirds()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeMajorThirdsWithThreshold(cents.MustFromInt(25))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeMinorThirds()
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
		_, err = zero.AnalyzeMinorThirdsWithThreshold(cents.MustFromInt(25))
		require.ErrorIs(t, err, temperament.ErrInvalidArgument)
	})

	t.Run("Must... wrappers panic on error", func(t *testing.T) {
		t.Parallel()

		var zero temperament.Temperament
		require.Panics(t, func() { temperament.MustEDO(0) })
		require.Panics(t, func() { temperament.MustNewPythagorean(-12) })
		require.Panics(t, func() { temperament.MustNewPythagoreanWithWolf(12) })
		require.Panics(t, func() { temperament.MustNewMeantone(1, 0) })
		require.Panics(t, func() { temperament.MustNewMeantoneWithRange(1, 4, 1) })
		require.Panics(t, func() { temperament.MustNewMeantoneFromFifth(cents.MustFromInt(720)) })
		require.Panics(t, func() { temperament.MustNewMeantoneFromFifthWithRange(cents.MustFromInt(700), 1) })
		require.Panics(t, func() { temperament.MustNew("Bad", nil) })
		require.Panics(t, func() { temperament.MustFromOffsets("Bad", nil) })
		require.Panics(t, func() { temperament.MustFromFloats("Bad", []float64{math.NaN()}) })
		require.Panics(t, func() { temperament.MustFromRatios("Bad", nil) })
		require.Panics(t, func() { temperament.MustFromFifths("Bad", nil) })
		require.Panics(t, func() { temperament.MustFromFifthsWithStart("Bad", nil, 0) })
		require.Panics(t, func() { temperament.MustFromFifthDeviations("Bad", nil) })
		require.Panics(t, func() { temperament.MustFromFifthDeviationsWithStart("Bad", nil, 0) })
		require.Panics(t, func() { temperament.MustFromGenerator("Bad", 0, cents.Zero(), 0) })

		require.Panics(t, func() { zero.MustWithName("X") })
		require.Panics(t, func() { zero.MustOffset(0) })
		require.Panics(t, func() { zero.MustSplitStep(0) })
		require.Panics(t, func() { zero.MustCumulativeOffset(0) })
		require.Panics(t, func() { zero.MustDirectedInterval(0, 1) })
		require.Panics(t, func() { zero.MustCyclicInterval(0, 1) })
		require.Panics(t, func() { zero.MustStepSizes() })
		require.Panics(t, func() { zero.MustMinStepSize() })
		require.Panics(t, func() { zero.MustMaxStepSize() })
		require.Panics(t, func() { zero.MustRotate(1) })
		require.Panics(t, func() { zero.MustNearestStep(cents.Zero()) })
		require.Panics(t, func() { zero.MustNearestCumulativeStep(cents.Zero()) })
		require.Panics(t, func() { zero.MustAnalyzeEDO() })
		require.Panics(t, func() { zero.MustAnalyzeIntervals(1, cents.MustFromInt(100), cents.MustFromInt(10)) })
		require.Panics(t, func() { zero.MustAnalyzeFifths() })
		require.Panics(t, func() { zero.MustAnalyzeFifthsWithThreshold(cents.MustFromInt(15)) })
		require.Panics(t, func() { zero.MustAnalyzeMajorThirds() })
		require.Panics(t, func() { zero.MustAnalyzeMajorThirdsWithThreshold(cents.MustFromInt(25)) })
		require.Panics(t, func() { zero.MustAnalyzeMinorThirds() })
		require.Panics(t, func() { zero.MustAnalyzeMinorThirdsWithThreshold(cents.MustFromInt(25)) })
	})
}

// TestConcurrencySafety verifies concurrent read and analysis operations across goroutines (NF-2).
func TestConcurrencySafety(t *testing.T) {
	t.Parallel()

	presets := temperament.Presets()
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Go(func() {
			for i, p := range presets {
				_ = p.Name()
				_ = p.StepsPerOctave()
				_ = p.Offsets()
				_ = p.MustOffset((g + i) % p.StepsPerOctave())
				_ = p.MustCumulativeOffset(g*12 - i)
				_ = p.MustRotate(g)
				_, _ = p.MustNearestStep(cents.MustFromInt(386))
				_ = p.MustAnalyzeEDO()
				if p.StepsPerOctave() == 12 {
					_ = p.MustAnalyzeFifths()
				}
			}
		})
	}
	wg.Wait()
}

// TestConcurrentReadsAndDeepComparison verifies NF-2 and API property 6 under the race detector: read operations on
// a temperament may run concurrently with any other reader of the same value, including a structural comparison
// with reflect.DeepEqual (as performed by assert.Equal); an immutable value has no writes that could race.
// The test is deliberately sequential (an exception to docs/rules/dev-test.md "Parallelism") so that a race report,
// if any, is attributed to this test only and not to unrelated tests running in parallel.
func TestConcurrentReadsAndDeepComparison(t *testing.T) {
	for _, source := range []temperament.Temperament{temperament.Vallotti(), temperament.Kellner(), temperament.MustEDO(72)} {
		read := temperament.MustNew(source.Name(), source.Offsets())
		other := temperament.MustNew(source.Name(), source.Offsets())
		var wg sync.WaitGroup
		for range 4 {
			wg.Go(func() { exerciseReadOperations(read) })
			wg.Go(func() {
				for range 50 {
					_ = reflect.DeepEqual(read, other)
				}
			})
		}
		wg.Wait()
		assert.True(t, read.Equal(other), source.Name())
	}
}

// cyclicFromOffsets is an independent oracle of the cyclic interval I_cyc(from, to) built only from Offsets() (S-13).
func cyclicFromOffsets(offsets []cents.Cents, from, to int) cents.Cents {
	diff := offsets[to].MustSub(offsets[from])
	if to < from {
		diff = diff.MustAdd(cents.Octave())
	}
	return diff
}

// largeTemperaments returns valid temperaments with N > 64 steps per octave, uniform and non-uniform.
func largeTemperaments() []temperament.Temperament {
	const irregularN = 128
	irregular := make([]cents.Cents, irregularN)
	for m := range irregularN {
		irregular[m] = cents.MustStep(m, irregularN).MustAdd(cents.MustFromFraction(m%5, 7))
	}
	pureFifth := cents.MustFromRatio(3, 2)
	return []temperament.Temperament{
		temperament.MustEDO(65),
		temperament.MustEDO(72),
		temperament.MustFromGenerator("Pythagorean 65", 65, pureFifth, -32),
		temperament.MustFromGenerator("Pythagorean 100", 100, pureFifth, -50),
		temperament.MustNew("Irregular 128", irregular),
		temperament.MustFromGenerator("Meantone 200", 200, pureFifth.MustSub(cents.SyntonicComma().MustDiv(4)), -199),
	}
}

// TestLargeDimensionOperations verifies intervals, rotation, step sizes, and interval analysis for N > 64,
// with non-zero starting steps (OP-12, OP-13, OP-14, OP-19, S-13, S-14, S-10, S-19, S-21, G-2).
func TestLargeDimensionOperations(t *testing.T) {
	t.Parallel()

	for _, tm := range largeTemperaments() {
		t.Run(tm.Name(), func(t *testing.T) {
			t.Parallel()

			n := tm.StepsPerOctave()
			require.Greater(t, n, 64)
			offsets := tm.Offsets()
			require.Len(t, offsets, n)
			fromSteps := []int{0, 1, 2, 7, n/2 - 1, n / 2, n - 2, n - 1}

			// OP-12: cyclic intervals from non-zero steps match the oracle and are complementary.
			for _, from := range fromSteps {
				for to := range n {
					got, err := tm.CyclicInterval(from, to)
					require.NoError(t, err)
					want := cyclicFromOffsets(offsets, from, to)
					require.True(t, got.Equal(want), "CyclicInterval(%d, %d) = %v, want %v", from, to, got, want)
					if from != to {
						assert.True(t, got.MustAdd(tm.MustCyclicInterval(to, from)).Equal(cents.Octave()))
					}
				}
			}

			// OP-12: directed intervals across octaves from non-zero steps equal c(to) - c(from).
			for _, from := range []int{1, n/2 + 1, n - 1, -1, -n - 3, 2*n + 5} {
				for _, to := range []int{0, 3, n - 1, n, n + 7, -n + 1, 3 * n} {
					got, err := tm.DirectedInterval(from, to)
					require.NoError(t, err)
					want := tm.MustCumulativeOffset(to).MustSub(tm.MustCumulativeOffset(from))
					assert.True(t, got.Equal(want), "DirectedInterval(%d, %d)", from, to)
				}
			}

			// OP-13: step sizes are the cyclic intervals between adjacent steps and sum to one octave.
			stepSizes, err := tm.StepSizes()
			require.NoError(t, err)
			require.Len(t, stepSizes, n)
			sum := cents.Zero()
			for m, s := range stepSizes {
				assert.True(t, s.Equal(cyclicFromOffsets(offsets, m, (m+1)%n)), "StepSizes()[%d]", m)
				sum = sum.MustAdd(s)
			}
			assert.True(t, sum.Equal(cents.Octave()))

			// OP-14: rotation to a non-zero root follows T^(r)(m) = I_cyc(r0, (r0+m) mod N) and the group laws.
			for _, r := range []int{1, 7, n / 2, n - 1, -1, -n - 3, 5*n + 2} {
				rotated, err := tm.Rotate(r)
				require.NoError(t, err)
				assertValidTemperamentInvariants(t, rotated, tm.Name(), n)
				r0 := ((r % n) + n) % n
				for m := range n {
					want := cyclicFromOffsets(offsets, r0, (r0+m)%n)
					require.True(t, rotated.MustOffset(m).Equal(want), "Rotate(%d) offset %d", r, m)
				}
				assert.True(t, rotated.MustRotate(-r).Equal(tm), "Rotate(%d).Rotate(%d) must restore the table", r, -r)
				assert.True(t, rotated.MustRotate(3).Equal(tm.MustRotate(r+3)), "Rotate(%d).Rotate(3) == Rotate(%d)", r, r+3)
			}

			// OP-19: interval analysis for several step sizes; Sizes[m] = I_cyc(m, (m+s) mod N) from the offsets oracle.
			for _, s := range []int{1, 2, 7, n / 2, n - 1} {
				ref := cents.MustStep(s, n)
				analysis, err := tm.AnalyzeIntervals(s, ref, cents.MustFromFraction(1, 10))
				require.NoError(t, err)
				assert.Equal(t, s, analysis.Step)
				assert.True(t, analysis.Reference.Equal(ref))
				require.Len(t, analysis.Sizes, n)
				sumSizes := cents.Zero()
				for m := range n {
					want := cyclicFromOffsets(offsets, m, (m+s)%n)
					require.True(t, analysis.Sizes[m].Equal(want), "AnalyzeIntervals(%d).Sizes[%d]", s, m)
					sumSizes = sumSizes.MustAdd(analysis.Sizes[m])
				}
				// P-15 item 1: the N cyclic intervals of step s sum to s octaves.
				assert.True(t, sumSizes.Equal(cents.Octave().MustMul(s)))
				assertIntervalAnalysisOracle(t, tm, analysis)
			}
		})
	}
}

// presetsWithAliases returns every built-in preset function together with its documented aliases (OP-1).
func presetsWithAliases() [][]func() temperament.Temperament {
	return [][]func() temperament.Temperament{
		{temperament.Equal12, temperament.Equal, temperament.EDO12},
		{temperament.Equal19, temperament.EDO19},
		{temperament.Equal22, temperament.EDO22},
		{temperament.Equal24, temperament.EDO24},
		{temperament.Equal31, temperament.EDO31},
		{temperament.Equal53, temperament.EDO53},
		{temperament.Pythagorean},
		{temperament.PythagoreanEFlatGSharp},
		{temperament.Just5Limit, temperament.Just},
		{temperament.Just5LimitSymmetric1},
		{temperament.Just5LimitSymmetric2},
		{temperament.Just7Limit},
		{temperament.MeantoneQuarterComma, temperament.Meantone},
		{temperament.MeantoneThirdComma},
		{temperament.MeantoneFifthComma},
		{temperament.MeantoneSixthComma},
		{temperament.MeantoneTwoSeventhComma},
		{temperament.WerckmeisterIII},
		{temperament.WerckmeisterIV},
		{temperament.WerckmeisterV},
		{temperament.KirnbergerI},
		{temperament.KirnbergerII},
		{temperament.KirnbergerIII},
		{temperament.Vallotti},
		{temperament.YoungI},
		{temperament.YoungII, temperament.Young},
		{temperament.Kellner},
		{temperament.NeidhardtI},
		{temperament.NeidhardtII},
		{temperament.NeidhardtIII},
		{temperament.NeidhardtDorf1732},
	}
}

// exerciseReadOperations calls every non-mutating operation of the public API on tm and returns nothing;
// the contract (architecture, API properties 5 and 6) requires that none of them modifies tm.
func exerciseReadOperations(tm temperament.Temperament) {
	n := tm.StepsPerOctave()
	_ = tm.IsValid()
	_ = tm.Name()
	_ = tm.Offsets()
	_ = tm.String()
	_, _ = tm.WithName("renamed")
	for _, step := range []int{0, 1, n - 1, n, n + 1, 13, -1, -13, -n, 7 * n, -7 * n} {
		_, _ = tm.Offset(((step % n) + n) % n)
		_, _, _ = tm.SplitStep(step)
		_, _ = tm.CumulativeOffset(step)
		_, _ = tm.DirectedInterval(step, -step+5)
		_, _ = tm.Rotate(step)
	}
	_, _ = tm.CyclicInterval(n-1, 0)
	_, _ = tm.StepSizes()
	_, _ = tm.MinStepSize()
	_, _ = tm.MaxStepSize()
	_ = tm.Equal(temperament.Equal12())
	_ = tm.Near(temperament.Equal12(), cents.MustFromInt(30))
	_ = tm.IsEqual()
	_ = tm.IsRational()
	for _, x := range []int{-1210, -10, 0, 50, 386, 1150, 1190, 2450} {
		_, _, _ = tm.NearestStep(cents.MustFromInt(x))
		_, _, _ = tm.NearestCumulativeStep(cents.MustFromInt(x))
	}
	_, _ = tm.AnalyzeEDO()
	if n >= 2 {
		_, _ = tm.AnalyzeIntervals(1, cents.MustStep(1, n), cents.MustFromInt(15))
	}
	_, _ = tm.AnalyzeFifths()
	_, _ = tm.AnalyzeFifthsWithThreshold(cents.MustFromInt(22))
	_, _ = tm.AnalyzeMajorThirds()
	_, _ = tm.AnalyzeMajorThirdsWithThreshold(cents.MustFromInt(10))
	_, _ = tm.AnalyzeMinorThirds()
	_, _ = tm.AnalyzeMinorThirdsWithThreshold(cents.MustFromInt(10))
}

// TestImmutabilityUnderReads verifies that read operations never change a Temperament value
// (architecture API properties 4-6, P-5, P-16, NF-2). Two temperaments built from the same input are compared
// structurally with reflect.DeepEqual (the comparison used by assert.Equal and require.Equal) before and after
// all read operations are applied to one of them: an immutable value cannot change that comparison.
// A built-in preset and its aliases are the same immutable instance and stay deeply equal.
// The test is deliberately sequential (an exception to docs/rules/dev-test.md "Parallelism"): it compares values
// structurally, and running it in parallel with other tests would let a contract violation surface as unrelated
// race reports in those tests.
func TestImmutabilityUnderReads(t *testing.T) {
	for _, group := range presetsWithAliases() {
		name := group[0]().Name()
		t.Run(name, func(t *testing.T) {
			preset := group[0]()
			for _, alias := range group[1:] {
				assert.True(t, reflect.DeepEqual(preset, alias()), "%s: an alias must return the same instance", name)
			}

			// Architecture, "reflect.DeepEqual": DeepEqual holds iff Name() and Equal agree, whatever the constructor.
			assert.True(t, reflect.DeepEqual(preset, temperament.MustNew(name, preset.Offsets())), "%s: same name and table", name)
			assert.False(t, reflect.DeepEqual(preset, preset.MustWithName(name+" (renamed)")), "%s: different name", name)
			rotated := preset.MustRotate(1).MustWithName(name) // same name; same table only for N-EDO and N = 1
			assert.Equal(t, rotated.Equal(preset), reflect.DeepEqual(preset, rotated), "%s: rotated by one step", name)

			read := temperament.MustNew(name, preset.Offsets())
			untouched := temperament.MustNew(name, preset.Offsets())
			before := reflect.DeepEqual(read, untouched)
			assert.True(t, before, "%s: two temperaments built from the same input must be deeply equal", name)

			exerciseReadOperations(read)
			assert.Equal(t, before, reflect.DeepEqual(read, untouched), "%s: read operations changed the value", name)
			assert.True(t, read.Equal(untouched))

			exerciseReadOperations(untouched)
			assert.Equal(t, before, reflect.DeepEqual(read, untouched), "%s: read operations changed the value", name)

			for _, alias := range group[1:] {
				assert.True(t, reflect.DeepEqual(group[0](), alias()), "%s: an alias must stay the same instance", name)
			}
		})
	}

	t.Run("large and single-step temperaments", func(t *testing.T) {
		for _, tm := range append(largeTemperaments(), temperament.MustEDO(1)) {
			read := temperament.MustNew(tm.Name(), tm.Offsets())
			untouched := temperament.MustNew(tm.Name(), tm.Offsets())
			before := reflect.DeepEqual(read, untouched)
			exerciseReadOperations(read)
			assert.Equal(t, before, reflect.DeepEqual(read, untouched), "%s: read operations changed the value", tm.Name())
		}
	})
}

// TestZeroAllocationReads verifies NF-3: StepsPerOctave, Name, Offset (OP-10) for valid steps, and Equal perform no heap
// allocations, on varied temperaments (presets, single-step, custom, N > 64, zero value) and varied steps and pairs.
// Each input is measured separately, so a single allocating input is not hidden by averaging.
// The other read operations may allocate inside package cents (NF-3) and are only benchmarked (bench_test.go).
// The test is deliberately sequential (an exception to docs/rules/dev-test.md "Parallelism"):
// testing.AllocsPerRun counts allocations of the whole process, so parallel tests would distort the count.
func TestZeroAllocationReads(t *testing.T) {
	temps := append(temperament.Presets(), temperament.MustEDO(1), temperament.MustEDO(7), temperament.Temperament{})
	temps = append(temps, largeTemperaments()...)

	const runs = 20
	assertNoAllocs := func(f func(), format string, args ...any) {
		t.Helper()
		assert.Zero(t, testing.AllocsPerRun(runs, f), append([]any{format + " must not allocate (NF-3)"}, args...)...)
	}

	for k, tm := range temps {
		name := tm.Name()
		assertNoAllocs(func() { _ = tm.StepsPerOctave() }, "StepsPerOctave of %q", name)
		assertNoAllocs(func() { _ = tm.Name() }, "Name of %q", name)
		n := tm.StepsPerOctave()
		for _, m := range []int{0, 1, n / 2, n - 1} {
			if m >= 0 && m < n {
				assertNoAllocs(func() { _, _ = tm.Offset(m) }, "Offset(%d) of %q", m, name)
			}
		}

		same := tm
		if tm.IsValid() {
			same = temperament.MustNew("Copy", tm.Offsets())
		}
		other := temps[(k+1)%len(temps)]
		assertNoAllocs(func() { _ = tm.Equal(same) }, "Equal of %q with an equal table", name)
		assertNoAllocs(func() { _ = tm.Equal(other) }, "Equal of %q with %q", name, other.Name())
	}
}

// generatorOracle builds the expected FromGenerator table directly from the formula of OP-8 and P-9
// (architecture, "Checkable consequences": points (k*g).SplitOctaves() for k in the chain, sorted ascending).
// It reports ok == false when two points coincide. Valid only when every k*g lies in the range of package cents.
func generatorOracle(n int, g cents.Cents, kMin int) (offsets []cents.Cents, ok bool) {
	offsets = make([]cents.Cents, 0, n)
	for k := kMin; k < kMin+n; k++ {
		_, rest := g.MustMul(k).SplitOctaves()
		offsets = append(offsets, rest)
	}
	slices.SortFunc(offsets, cents.Cents.Compare)
	for m := 1; m < n; m++ {
		if offsets[m-1].Equal(offsets[m]) {
			return offsets, false
		}
	}
	return offsets, true
}

// TestFromGeneratorContract verifies the FromGenerator contract of the architecture (OP-8, P-9, S-9, K-12, decision 7):
// agreement with the direct formula, periodicity and symmetry in the generator, kMin bounds, N = 1, coinciding points,
// the presets, large N without ErrOutOfRange, and ErrTooComplex for an artificially complex generator.
func TestFromGeneratorContract(t *testing.T) {
	t.Parallel()

	pureFifth := cents.MustFromRatio(3, 2)
	quarterComma := pureFifth.MustSub(cents.SyntonicComma().MustDiv(4))
	generators := []cents.Cents{
		pureFifth,
		quarterComma,
		cents.MustFromRatio(5, 4),
		cents.MustFromRatio(7, 4),
		cents.MustFromFloat(701.955),
		cents.MustFromInt(-317),
		cents.MustFromInt(700),
		cents.MustFromFraction(1200, 7),
		cents.MustStep(5, 19),
		pureFifth.Neg(),
		pureFifth.MustAdd(cents.MustFromInt(3600)),
	}

	t.Run("agreement with the direct formula for small chains", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{1, 2, 5, 7, 12, 13, 19, 31, 53} {
			for _, g := range generators {
				for _, kMin := range []int{0, -1, -(n - 1) / 2, -(n - 1)} {
					if kMin < -(n-1) || kMin > 0 {
						continue
					}
					want, ok := generatorOracle(n, g, kMin)
					got, err := temperament.FromGenerator("Generated", n, g, kMin)
					if !ok {
						require.ErrorIs(t, err, temperament.ErrInvalidArgument, "N=%d g=%v kMin=%d: coinciding points", n, g, kMin)
						continue
					}
					require.NoError(t, err, "N=%d g=%v kMin=%d", n, g, kMin)
					assertValidTemperamentInvariants(t, got, "Generated", n)
					assert.True(t, got.Equal(temperament.MustNew("Generated", want)), "N=%d g=%v kMin=%d", n, g, kMin)
				}
			}
		}
	})

	t.Run("periodicity in the generator modulo the octave", func(t *testing.T) {
		t.Parallel()

		for _, g := range []cents.Cents{pureFifth, quarterComma, cents.MustFromInt(700), cents.MustFromRatio(7, 4)} {
			base := temperament.MustFromGenerator("Base", 12, g, -3)
			for _, j := range []int{-7000, -5, -1, 1, 2, 7000} {
				shifted := g.MustAdd(cents.Octave().MustMul(j))
				got, err := temperament.FromGenerator("Shifted", 12, shifted, -3)
				require.NoError(t, err, "g=%v j=%d", g, j)
				assert.True(t, got.Equal(base), "g=%v j=%d", g, j)
			}
		}

		// -500 cents and the large generator 8 400 700 cents give the same table as 700 cents (12-EDO for N = 12).
		edo := temperament.MustFromGenerator("700", 12, cents.MustFromInt(700), -9)
		assert.True(t, edo.Equal(temperament.Equal12()))
		for _, g := range []int{-500, 8_400_700, -8_399_300} {
			got, err := temperament.FromGenerator("Shifted", 12, cents.MustFromInt(g), -9)
			if !assert.NoError(t, err, "g=%d", g) {
				continue
			}
			assert.True(t, got.Equal(edo), "g=%d", g)
		}
	})

	t.Run("symmetry under negating the generator and mirroring the chain", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{1, 5, 12, 19, 53, 101} {
			for _, g := range []cents.Cents{pureFifth, quarterComma, cents.MustFromRatio(5, 4), cents.MustFromInt(-317)} {
				for _, kMin := range []int{0, -1, -(n - 1) / 3, -(n - 1)} {
					if kMin < -(n-1) || kMin > 0 {
						continue
					}
					direct, err := temperament.FromGenerator("Direct", n, g, kMin)
					require.NoError(t, err, "N=%d g=%v kMin=%d", n, g, kMin)
					mirrored, err := temperament.FromGenerator("Mirrored", n, g.Neg(), -kMin-(n-1))
					require.NoError(t, err, "N=%d g=%v kMin=%d", n, g, -kMin-(n-1))
					assert.True(t, mirrored.Equal(direct), "N=%d g=%v kMin=%d", n, g, kMin)
				}
			}
		}
	})

	t.Run("kMin bounds", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{1, 2, 12, 53} {
			for _, kMin := range []int{0, -(n - 1)} {
				_, err := temperament.FromGenerator("Bound", n, pureFifth, kMin)
				require.NoError(t, err, "N=%d kMin=%d", n, kMin)
			}
			for _, kMin := range []int{1, -n, math.MinInt, math.MaxInt} {
				_, err := temperament.FromGenerator("Bound", n, pureFifth, kMin)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument, "N=%d kMin=%d", n, kMin)
			}
		}
	})

	t.Run("N = 1 gives the table [0] for any generator", func(t *testing.T) {
		t.Parallel()

		for _, g := range []cents.Cents{cents.Zero(), cents.Octave(), cents.MustFromInt(-2400), pureFifth, cents.MustFromInt(9_000_000)} {
			got, err := temperament.FromGenerator("One", 1, g, 0)
			require.NoError(t, err, "g=%v", g)
			assertValidTemperamentInvariants(t, got, "One", 1)
			assert.True(t, got.Equal(temperament.MustEDO(1)), "g=%v", g)
		}
	})

	t.Run("coinciding points are ErrInvalidArgument", func(t *testing.T) {
		t.Parallel()

		for _, g := range []cents.Cents{cents.Zero(), cents.Octave(), cents.MustFromInt(-2400)} {
			for _, n := range []int{2, 12, 53} {
				_, err := temperament.FromGenerator("Degenerate", n, g, 0)
				require.ErrorIs(t, err, temperament.ErrInvalidArgument, "g=%v N=%d", g, n)
			}
		}
		for _, n := range []int{13, 24} {
			_, err := temperament.FromGenerator("Degenerate", n, cents.MustFromInt(700), 0)
			require.ErrorIs(t, err, temperament.ErrInvalidArgument, "g=700 N=%d", n)
		}
		for kMin := -11; kMin <= 0; kMin++ {
			got, err := temperament.FromGenerator("700", 12, cents.MustFromInt(700), kMin)
			require.NoError(t, err, "kMin=%d", kMin)
			assert.True(t, got.Equal(temperament.Equal12()), "kMin=%d", kMin)
		}
	})

	t.Run("presets (S-9)", func(t *testing.T) {
		t.Parallel()

		assert.True(t, temperament.MustFromGenerator("P", 12, pureFifth, temperament.DefaultPythagoreanMinFifth).Equal(temperament.Pythagorean()))
		assert.True(t, temperament.MustFromGenerator("P", 12, pureFifth, temperament.EFlatPythagoreanMinFifth).Equal(temperament.PythagoreanEFlatGSharp()))
		assert.True(t, temperament.MustFromGenerator("M", 12, quarterComma, -3).Equal(temperament.MeantoneQuarterComma()))
		named := temperament.MustFromGenerator("Name Kept", 12, pureFifth, -5)
		assert.Equal(t, "Name Kept", named.Name())
	})

	t.Run("large N without ErrOutOfRange (S-9, K-12)", func(t *testing.T) {
		t.Parallel()

		// Every point of the chain k = 0..19999 is reduced exactly; the former product k*g would exceed 9e6 cents.
		const n = 20_000
		got, err := temperament.FromGenerator("Pythagorean 20000", n, pureFifth, 0)
		require.NoError(t, err)
		assert.NotErrorIs(t, err, temperament.ErrOutOfRange)
		require.Equal(t, n, got.StepsPerOctave())
		offsets := got.Offsets()
		require.Len(t, offsets, n)
		assert.True(t, offsets[0].Equal(cents.Zero()))
		for m := 1; m < n; m++ {
			require.Negative(t, offsets[m-1].Compare(offsets[m]), "strict monotonicity at m=%d", m)
		}
		assert.Negative(t, offsets[n-1].Compare(cents.Octave()))

		// Every point equals (k*g) mod 1200 for the chain; spot-check points whose product is in the cents range.
		for _, k := range []int{1, 2, 7, 12, 5000} {
			_, want := pureFifth.MustMul(k).SplitOctaves()
			assert.True(t, slices.ContainsFunc(offsets, want.Equal), "point k=%d", k)
		}

		mirrored, err := temperament.FromGenerator("Pythagorean 20000", n, pureFifth.Neg(), -(n - 1))
		require.NoError(t, err)
		assert.True(t, mirrored.Equal(got))
	})

	t.Run("ErrTooComplex for an artificially complex generator", func(t *testing.T) {
		t.Parallel()

		// g = r + e3*c(3:1) with e3 = 1e22 and an integer rational part r chosen so that g lies in [0, 1) cent.
		// A chain point k has the exponent k*e3 for prime 3; package cents limits |e_p| <= 1e24 (architecture,
		// FromGenerator algorithm item 5), so k = 100 is still representable and k = 101 is not.
		e3 := new(big.Int).Exp(big.NewInt(10), big.NewInt(22), nil)
		r, ok := new(big.Int).SetString("-19019550008653874177444867", 10)
		require.True(t, ok)
		g, err := cents.FromComposition(r, big.NewInt(1), cents.Power{Base: 3, Num: e3})
		require.NoError(t, err)
		require.GreaterOrEqual(t, g.Sign(), 0)
		require.Negative(t, g.Compare(cents.Cent()))

		// Chains whose largest |k| is 100 stay within the limit.
		for _, tc := range []struct {
			n, kMin int
		}{{101, 0}, {101, -100}, {201, -100}} {
			withinLimit, err := temperament.FromGenerator("Complex", tc.n, g, tc.kMin)
			require.NoError(t, err, "N=%d kMin=%d", tc.n, tc.kMin)
			assert.Equal(t, tc.n, withinLimit.StepsPerOctave())
		}

		// Chains containing |k| = 101 exceed it.

		for _, tc := range []struct {
			n, kMin int
		}{{102, 0}, {102, -101}, {200, -101}} {
			_, err := temperament.FromGenerator("Complex", tc.n, g, tc.kMin)
			require.ErrorIs(t, err, temperament.ErrTooComplex, "N=%d kMin=%d", tc.n, tc.kMin)
			assert.ErrorIs(t, err, cents.ErrTooComplex, "N=%d kMin=%d", tc.n, tc.kMin)
			assert.NotErrorIs(t, err, temperament.ErrOutOfRange, "N=%d kMin=%d", tc.n, tc.kMin)
		}
	})
}
