package temperament

import (
	"math"

	"github.com/go-muse/musex/cents"
)

// EDODeviation holds the step-by-step and summary deviation metrics of a
// temperament from N-EDO.
type EDODeviation struct {
	// StepsPerOctave is the dimension N of the analyzed temperament.
	StepsPerOctave int
	// Deviations contains the N signed deviations d_m = T(m) - Step(m, N) for m in 0..N-1.
	Deviations []cents.Cents
	// MaxAbsDeviation is the maximum absolute deviation D_max = max_m |d_m|.
	MaxAbsDeviation cents.Cents
	// MaxPosition is the smallest step index m in 0..N-1 where MaxAbsDeviation is attained.
	MaxPosition int
	// MeanAbsDeviation is the mean absolute deviation D_mean = (1/N) * sum_m |d_m|.
	MeanAbsDeviation cents.Cents
	// RMSDeviation is the root-mean-square deviation D_RMS = sqrt((1/N) * sum_m d_m^2) in cents.
	RMSDeviation float64
}

// IntervalAnalysis holds the acoustic analysis of all N cyclic intervals of a
// given step size s.
type IntervalAnalysis struct {
	// Step is the analyzed cyclic step size s in 1..N-1.
	Step int
	// Reference is the pure reference interval I_ref in cents.
	Reference cents.Cents
	// WolfThreshold is the strict threshold tau_wolf > 0 above which |e_s(m)| is classified as a wolf interval.
	WolfThreshold cents.Cents
	// Sizes contains the N cyclic interval sizes I_s(m) = CyclicInterval(m, (m+s) mod N) for m in 0..N-1.
	Sizes []cents.Cents
	// Deviations contains the N signed deviations e_s(m) = I_s(m) - Reference for m in 0..N-1.
	Deviations []cents.Cents
	// PureCount is the number of positions m in 0..N-1 with exact zero deviation (e_s(m) == 0).
	PureCount int
	// Min is the minimum interval size min_m I_s(m).
	Min cents.Cents
	// MinPosition is the smallest step index m in 0..N-1 where Min is attained.
	MinPosition int
	// Max is the maximum interval size max_m I_s(m).
	Max cents.Cents
	// MaxPosition is the smallest step index m in 0..N-1 where Max is attained.
	MaxPosition int
	// Spread is the interval spread Max - Min.
	Spread cents.Cents
	// MaxAbsDeviation is the maximum absolute deviation max_m |e_s(m)| from Reference.
	MaxAbsDeviation cents.Cents
	// WorstPosition is the smallest step index m in 0..N-1 where MaxAbsDeviation is attained.
	WorstPosition int
	// MeanAbsDeviation is the mean absolute deviation (1/N) * sum_m |e_s(m)| from Reference.
	MeanAbsDeviation cents.Cents
	// RMSDeviation is the root-mean-square deviation sqrt((1/N) * sum_m e_s(m)^2) in cents.
	RMSDeviation float64
	// WolfPositions contains all step indices m in 0..N-1 in ascending order where |e_s(m)| > WolfThreshold.
	WolfPositions []int
}

// AnalyzeEDO computes the step-by-step signed deviations
// d_m = T(m) - cents.Step(m, N) and summary metrics (MaxAbsDeviation,
// MaxPosition, MeanAbsDeviation, RMSDeviation) relative to N-EDO. It returns
// [ErrInvalidArgument] if t is invalid, or [ErrTooComplex] if exact cents
// arithmetic overflows complexity limits.
func (t Temperament) AnalyzeEDO() (EDODeviation, error) {
	if !t.IsValid() {
		return EDODeviation{}, errorf(ErrInvalidArgument, "AnalyzeEDO: invalid temperament")
	}
	n := len(t.offsets)
	devs := make([]cents.Cents, n)
	devs[0] = cents.Zero()
	maxAbs := cents.Zero()
	maxPos := 0
	sumAbs := cents.Zero()
	var sumSq float64

	for m := 1; m < n; m++ {
		edoStep, err := cents.Step(m, n)
		if err != nil {
			return EDODeviation{}, wrapCentsError("AnalyzeEDO", err)
		}
		if t.offsets[m].Equal(edoStep) {
			devs[m] = cents.Zero()
			continue
		}
		d, err := t.offsets[m].Sub(edoStep)
		if err != nil {
			return EDODeviation{}, wrapCentsError("AnalyzeEDO", err)
		}
		devs[m] = d
		absD := d.Abs()
		if absD.Compare(maxAbs) > 0 {
			maxAbs = absD
			maxPos = m
		}
		sumAbs, err = sumAbs.Add(absD)
		if err != nil {
			return EDODeviation{}, wrapCentsError("AnalyzeEDO", err)
		}
		df := d.Float64()
		sumSq += df * df
	}

	meanAbs, err := sumAbs.Div(n)
	if err != nil {
		return EDODeviation{}, wrapCentsError("AnalyzeEDO", err)
	}
	var rms float64
	if sumSq > 0 {
		rms = math.Sqrt(sumSq / float64(n))
	}
	return EDODeviation{
		StepsPerOctave:   n,
		Deviations:       devs,
		MaxAbsDeviation:  maxAbs,
		MaxPosition:      maxPos,
		MeanAbsDeviation: meanAbs,
		RMSDeviation:     rms,
	}, nil
}

// MustAnalyzeEDO is like [Temperament.AnalyzeEDO] but panics on error.
func (t Temperament) MustAnalyzeEDO() EDODeviation {
	return must(t.AnalyzeEDO())
}

// AnalyzeIntervals analyzes all N cyclic intervals of step size step in
// [1, N-1] against reference interval ref, classifying intervals with
// |deviation| > wolfThreshold as wolf intervals. It returns
// [ErrInvalidArgument] if t is invalid, t.StepsPerOctave() < 2, step < 1,
// step >= N, or wolfThreshold <= 0.
func (t Temperament) AnalyzeIntervals(step int, ref, wolfThreshold cents.Cents) (IntervalAnalysis, error) {
	if !t.IsValid() {
		return IntervalAnalysis{}, errorf(ErrInvalidArgument, "AnalyzeIntervals: invalid temperament")
	}
	n := len(t.offsets)
	if n < 2 || step < 1 || step >= n {
		return IntervalAnalysis{}, errorf(ErrInvalidArgument, "AnalyzeIntervals(%d): step out of range [1, %d)", step, n)
	}
	if wolfThreshold.Sign() <= 0 {
		return IntervalAnalysis{}, errorf(ErrInvalidArgument, "AnalyzeIntervals: wolfThreshold must be positive, got %v", wolfThreshold)
	}

	sizes := make([]cents.Cents, n)
	devs := make([]cents.Cents, n)
	var wolves []int
	pureCount := 0
	var minSize, maxSize, maxAbsDev cents.Cents
	minPos, maxPos, worstPos := 0, 0, 0
	sumAbs := cents.Zero()
	var sumSq float64

	for m := range n {
		sz, err := t.CyclicInterval(m, (m+step)%n)
		if err != nil {
			return IntervalAnalysis{}, err
		}
		sizes[m] = sz
		var d cents.Cents
		if sz.Equal(ref) {
			d = cents.Zero()
			pureCount++
		} else {
			d, err = sz.Sub(ref)
			if err != nil {
				return IntervalAnalysis{}, wrapCentsError("AnalyzeIntervals", err)
			}
		}
		devs[m] = d
		absD := d.Abs()

		if m == 0 {
			minSize = sz
			maxSize = sz
			maxAbsDev = absD
		} else {
			if sz.Compare(minSize) < 0 {
				minSize = sz
				minPos = m
			}
			if sz.Compare(maxSize) > 0 {
				maxSize = sz
				maxPos = m
			}
			if absD.Compare(maxAbsDev) > 0 {
				maxAbsDev = absD
				worstPos = m
			}
		}

		if absD.Compare(wolfThreshold) > 0 {
			wolves = append(wolves, m)
		}
		sumAbs, err = sumAbs.Add(absD)
		if err != nil {
			return IntervalAnalysis{}, wrapCentsError("AnalyzeIntervals", err)
		}
		if !d.Equal(cents.Zero()) {
			df := d.Float64()
			sumSq += df * df
		}
	}

	spread, err := maxSize.Sub(minSize)
	if err != nil {
		return IntervalAnalysis{}, wrapCentsError("AnalyzeIntervals", err)
	}
	meanAbs, err := sumAbs.Div(n)
	if err != nil {
		return IntervalAnalysis{}, wrapCentsError("AnalyzeIntervals", err)
	}
	var rms float64
	if sumSq > 0 {
		rms = math.Sqrt(sumSq / float64(n))
	}

	return IntervalAnalysis{
		Step:             step,
		Reference:        ref,
		WolfThreshold:    wolfThreshold,
		Sizes:            sizes,
		Deviations:       devs,
		PureCount:        pureCount,
		Min:              minSize,
		MinPosition:      minPos,
		Max:              maxSize,
		MaxPosition:      maxPos,
		Spread:           spread,
		MaxAbsDeviation:  maxAbsDev,
		WorstPosition:    worstPos,
		MeanAbsDeviation: meanAbs,
		RMSDeviation:     rms,
		WolfPositions:    wolves,
	}, nil
}

// MustAnalyzeIntervals is like [Temperament.AnalyzeIntervals] but panics on
// error.
func (t Temperament) MustAnalyzeIntervals(step int, ref, wolfThreshold cents.Cents) IntervalAnalysis {
	return must(t.AnalyzeIntervals(step, ref, wolfThreshold))
}

// AnalyzeFifths analyzes the 12 cyclic fifths (step = 7) of a 12-tone
// temperament against pure fifth R(3:2) using [DefaultFifthWolfThreshold] (15
// cents). It returns [ErrInvalidArgument] if t is invalid or
// t.StepsPerOctave() != 12.
func (t Temperament) AnalyzeFifths() (IntervalAnalysis, error) {
	return t.AnalyzeFifthsWithThreshold(DefaultFifthWolfThreshold())
}

// MustAnalyzeFifths is like [Temperament.AnalyzeFifths] but panics on error.
func (t Temperament) MustAnalyzeFifths() IntervalAnalysis {
	return must(t.AnalyzeFifths())
}

// AnalyzeFifthsWithThreshold analyzes the 12 cyclic fifths (step = 7) of a
// 12-tone temperament against pure fifth R(3:2) using the given
// wolfThreshold > 0. It returns [ErrInvalidArgument] if t is invalid, N != 12,
// or wolfThreshold <= 0.
func (t Temperament) AnalyzeFifthsWithThreshold(wolfThreshold cents.Cents) (IntervalAnalysis, error) {
	if !t.IsValid() || len(t.offsets) != DefaultStepsPerOctave {
		return IntervalAnalysis{}, errorf(ErrInvalidArgument, "AnalyzeFifthsWithThreshold: requires a 12-tone temperament, got N=%d", len(t.offsets))
	}
	return t.AnalyzeIntervals(7, pureFifth, wolfThreshold)
}

// MustAnalyzeFifthsWithThreshold is like
// [Temperament.AnalyzeFifthsWithThreshold] but panics on error.
func (t Temperament) MustAnalyzeFifthsWithThreshold(wolfThreshold cents.Cents) IntervalAnalysis {
	return must(t.AnalyzeFifthsWithThreshold(wolfThreshold))
}

// AnalyzeMajorThirds analyzes the 12 cyclic major thirds (step = 4) of a
// 12-tone temperament against pure major third R(5:4) using
// [DefaultThirdWolfThreshold] (25 cents). It returns [ErrInvalidArgument] if t
// is invalid or t.StepsPerOctave() != 12.
func (t Temperament) AnalyzeMajorThirds() (IntervalAnalysis, error) {
	return t.AnalyzeMajorThirdsWithThreshold(DefaultThirdWolfThreshold())
}

// MustAnalyzeMajorThirds is like [Temperament.AnalyzeMajorThirds] but panics on
// error.
func (t Temperament) MustAnalyzeMajorThirds() IntervalAnalysis {
	return must(t.AnalyzeMajorThirds())
}

// AnalyzeMajorThirdsWithThreshold analyzes the 12 cyclic major thirds
// (step = 4) of a 12-tone temperament against R(5:4) using the given
// wolfThreshold > 0. It returns [ErrInvalidArgument] if t is invalid, N != 12,
// or wolfThreshold <= 0.
func (t Temperament) AnalyzeMajorThirdsWithThreshold(wolfThreshold cents.Cents) (IntervalAnalysis, error) {
	if !t.IsValid() || len(t.offsets) != DefaultStepsPerOctave {
		return IntervalAnalysis{}, errorf(ErrInvalidArgument, "AnalyzeMajorThirdsWithThreshold: requires a 12-tone temperament, got N=%d", len(t.offsets))
	}
	return t.AnalyzeIntervals(4, pureMajorThird, wolfThreshold)
}

// MustAnalyzeMajorThirdsWithThreshold is like
// [Temperament.AnalyzeMajorThirdsWithThreshold] but panics on error.
func (t Temperament) MustAnalyzeMajorThirdsWithThreshold(wolfThreshold cents.Cents) IntervalAnalysis {
	return must(t.AnalyzeMajorThirdsWithThreshold(wolfThreshold))
}

// AnalyzeMinorThirds analyzes the 12 cyclic minor thirds (step = 3) of a
// 12-tone temperament against pure minor third R(6:5) using
// [DefaultThirdWolfThreshold] (25 cents). It returns [ErrInvalidArgument] if t
// is invalid or t.StepsPerOctave() != 12.
func (t Temperament) AnalyzeMinorThirds() (IntervalAnalysis, error) {
	return t.AnalyzeMinorThirdsWithThreshold(DefaultThirdWolfThreshold())
}

// MustAnalyzeMinorThirds is like [Temperament.AnalyzeMinorThirds] but panics on
// error.
func (t Temperament) MustAnalyzeMinorThirds() IntervalAnalysis {
	return must(t.AnalyzeMinorThirds())
}

// AnalyzeMinorThirdsWithThreshold analyzes the 12 cyclic minor thirds
// (step = 3) of a 12-tone temperament against R(6:5) using the given
// wolfThreshold > 0. It returns [ErrInvalidArgument] if t is invalid, N != 12,
// or wolfThreshold <= 0.
func (t Temperament) AnalyzeMinorThirdsWithThreshold(wolfThreshold cents.Cents) (IntervalAnalysis, error) {
	if !t.IsValid() || len(t.offsets) != DefaultStepsPerOctave {
		return IntervalAnalysis{}, errorf(ErrInvalidArgument, "AnalyzeMinorThirdsWithThreshold: requires a 12-tone temperament, got N=%d", len(t.offsets))
	}
	return t.AnalyzeIntervals(3, pureMinorThird, wolfThreshold)
}

// MustAnalyzeMinorThirdsWithThreshold is like
// [Temperament.AnalyzeMinorThirdsWithThreshold] but panics on error.
func (t Temperament) MustAnalyzeMinorThirdsWithThreshold(wolfThreshold cents.Cents) IntervalAnalysis {
	return must(t.AnalyzeMinorThirdsWithThreshold(wolfThreshold))
}
