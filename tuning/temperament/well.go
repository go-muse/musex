package temperament

import (
	"slices"

	"github.com/go-muse/musex/cents"
)

var (
	pcMinusQuarter = cents.PythagoreanComma().MustDiv(-4)
	pcPlusQuarter  = cents.PythagoreanComma().MustDiv(4)
	pcMinusThird   = cents.PythagoreanComma().MustDiv(-3)
	pcPlusThird    = cents.PythagoreanComma().MustDiv(3)
	pcMinusFifth   = cents.PythagoreanComma().MustDiv(-5)
	pcMinusSixth   = cents.PythagoreanComma().MustDiv(-6)
	pcMinusTwelfth = cents.PythagoreanComma().MustDiv(-12)

	scMinusFull         = cents.SyntonicComma().Neg()
	scMinusHalf         = cents.SyntonicComma().MustDiv(-2)
	scMinusQuarter      = cents.SyntonicComma().MustDiv(-4)
	scMinusThreeSixteen = cents.SyntonicComma().MustScale(-3, 16)
	schismaNeg          = cents.Schisma().Neg()
	youngIEtaNeg        = pcMinusQuarter.MustSub(scMinusThreeSixteen)
	zeroCents           = cents.Zero()

	presetWerckmeisterIII = MustFromFifthDeviations("Werckmeister III", []cents.Cents{
		pcMinusQuarter, pcMinusQuarter, pcMinusQuarter, zeroCents,
		zeroCents, pcMinusQuarter, zeroCents, zeroCents,
		zeroCents, zeroCents, zeroCents, zeroCents,
	})

	presetWerckmeisterIV = MustFromFifthDeviations("Werckmeister IV", []cents.Cents{
		pcMinusThird, zeroCents, pcMinusThird, zeroCents,
		pcMinusThird, zeroCents, pcMinusThird, zeroCents,
		pcPlusThird, pcPlusThird, pcMinusThird, zeroCents,
	})

	presetWerckmeisterV = MustFromFifthDeviations("Werckmeister V", []cents.Cents{
		zeroCents, zeroCents, pcMinusQuarter, pcMinusQuarter,
		zeroCents, zeroCents, pcMinusQuarter, pcMinusQuarter,
		pcPlusQuarter, zeroCents, zeroCents, pcMinusQuarter,
	})

	presetKirnbergerI = MustFromFifthDeviations("Kirnberger I", []cents.Cents{
		zeroCents, zeroCents, scMinusFull, zeroCents,
		zeroCents, zeroCents, schismaNeg, zeroCents,
		zeroCents, zeroCents, zeroCents, zeroCents,
	})

	presetKirnbergerII = MustFromFifthDeviations("Kirnberger II", []cents.Cents{
		zeroCents, zeroCents, scMinusHalf, scMinusHalf,
		zeroCents, zeroCents, schismaNeg, zeroCents,
		zeroCents, zeroCents, zeroCents, zeroCents,
	})

	presetKirnbergerIII = MustFromFifthDeviations("Kirnberger III", []cents.Cents{
		scMinusQuarter, scMinusQuarter, scMinusQuarter, scMinusQuarter,
		zeroCents, zeroCents, schismaNeg, zeroCents,
		zeroCents, zeroCents, zeroCents, zeroCents,
	})

	presetVallotti = MustFromFifthDeviations("Vallotti", []cents.Cents{
		pcMinusSixth, pcMinusSixth, pcMinusSixth, pcMinusSixth,
		pcMinusSixth, zeroCents, zeroCents, zeroCents,
		zeroCents, zeroCents, zeroCents, pcMinusSixth,
	})

	presetYoungI = MustFromFifthDeviations("Young I", []cents.Cents{
		scMinusThreeSixteen, scMinusThreeSixteen, scMinusThreeSixteen, scMinusThreeSixteen,
		youngIEtaNeg, youngIEtaNeg, zeroCents, zeroCents,
		zeroCents, zeroCents, youngIEtaNeg, youngIEtaNeg,
	})

	presetYoungII = MustFromFifthDeviations("Young II", []cents.Cents{
		pcMinusSixth, pcMinusSixth, pcMinusSixth, pcMinusSixth,
		pcMinusSixth, pcMinusSixth, zeroCents, zeroCents,
		zeroCents, zeroCents, zeroCents, zeroCents,
	})

	presetKellner = MustFromFifthDeviations("Kellner", []cents.Cents{
		pcMinusFifth, pcMinusFifth, pcMinusFifth, pcMinusFifth,
		zeroCents, pcMinusFifth, zeroCents, zeroCents,
		zeroCents, zeroCents, zeroCents, zeroCents,
	})

	// The Neidhardt presets narrow fifths by k/12 of the Pythagorean comma,
	// a geometric share, so each deviation is exactly -k/12 Pc in cents.
	presetNeidhardtI = MustFromFifthDeviations("Neidhardt I", []cents.Cents{
		pcMinusSixth, pcMinusSixth, pcMinusSixth, pcMinusSixth,
		pcMinusTwelfth, pcMinusTwelfth, zeroCents, zeroCents,
		pcMinusTwelfth, pcMinusTwelfth, zeroCents, zeroCents,
	})

	presetNeidhardtII = MustFromFifthDeviations("Neidhardt II", []cents.Cents{
		pcMinusSixth, pcMinusSixth, pcMinusSixth, pcMinusTwelfth,
		zeroCents, pcMinusTwelfth, pcMinusTwelfth, pcMinusTwelfth,
		zeroCents, zeroCents, pcMinusTwelfth, pcMinusTwelfth,
	})

	presetNeidhardtIII = MustFromFifthDeviations("Neidhardt III", []cents.Cents{
		pcMinusSixth, pcMinusSixth, pcMinusSixth, pcMinusTwelfth,
		zeroCents, pcMinusTwelfth, pcMinusTwelfth, pcMinusTwelfth,
		zeroCents, pcMinusTwelfth, pcMinusTwelfth, zeroCents,
	})

	presetNeidhardtDorf1732 = MustFromFifthDeviations("Neidhardt Dorf (1732)", []cents.Cents{
		pcMinusTwelfth, pcMinusSixth, pcMinusQuarter, pcMinusQuarter,
		zeroCents, pcMinusTwelfth, zeroCents, pcMinusTwelfth,
		zeroCents, zeroCents, pcMinusTwelfth, zeroCents,
	})

	canonicalPresets = [31]Temperament{
		// 1. 6 N-EDO presets
		presetEqual12,
		presetEqual19,
		presetEqual22,
		presetEqual24,
		presetEqual31,
		presetEqual53,
		// 2. 2 Pythagorean presets
		presetPythagorean,
		presetPythagoreanEFlatGSharp,
		// 3. 4 Just Intonation presets
		presetJust5Limit,
		presetJust5LimitSymmetric1,
		presetJust5LimitSymmetric2,
		presetJust7Limit,
		// 4. 5 Meantone presets
		presetMeantoneQuarterComma,
		presetMeantoneThirdComma,
		presetMeantoneFifthComma,
		presetMeantoneSixthComma,
		presetMeantoneTwoSeventhComma,
		// 5. 14 Well Temperament presets
		presetWerckmeisterIII,
		presetWerckmeisterIV,
		presetWerckmeisterV,
		presetKirnbergerI,
		presetKirnbergerII,
		presetKirnbergerIII,
		presetVallotti,
		presetYoungI,
		presetYoungII,
		presetKellner,
		presetNeidhardtI,
		presetNeidhardtII,
		presetNeidhardtIII,
		presetNeidhardtDorf1732,
	}
)

// FromFifths creates a 12-tone temperament from 11 or 12 fifth sizes Q_i in
// cents starting at C (kMin = 0) along the circle of fifths
// (C-G-D-A-E-B-F#-C#-G#-Eb-Bb-F-[C]). If len(fifths) == 12, it verifies exact
// octave closure: sum(fifths) == 8400 cents (7 octaves).
func FromFifths(name string, fifths []cents.Cents) (Temperament, error) {
	return FromFifthsWithStart(name, fifths, 0)
}

// MustFromFifths is like [FromFifths] but panics on error.
func MustFromFifths(name string, fifths []cents.Cents) Temperament {
	return must(FromFifths(name, fifths))
}

// FromFifthsWithStart creates a 12-tone temperament from 11 or 12 fifth sizes
// Q_j in cents along the chain starting at fifth index kMin in [-11, 0]
// (k = kMin .. kMin+11, with k = 0 fixed at C = 0 cents), where fifths[j]
// (0 <= j <= 10) specifies the fifth transition from chain node k = kMin + j to
// k + 1 = kMin + j + 1. If len(fifths) == 12, fifths[11] closes the circle from
// kMin + 11 to kMin + 12 and exact closure sum(fifths) == 8400 cents is
// verified.
func FromFifthsWithStart(name string, fifths []cents.Cents, kMin int) (Temperament, error) {
	return fromFifths("FromFifthsWithStart", name, fifths, kMin)
}

// fromFifths implements [FromFifthsWithStart]; op names the public operation
// in error messages.
func fromFifths(op, name string, fifths []cents.Cents, kMin int) (Temperament, error) {
	if len(fifths) != 11 && len(fifths) != 12 {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: length must be 11 or 12, got %d", op, len(fifths))
	}
	if kMin < -11 || kMin > 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: kMin %d out of range [-11, 0]", op, kMin)
	}
	if len(fifths) == 12 {
		sum := cents.Zero()
		for _, q := range fifths {
			var err error
			sum, err = sum.Add(q)
			if err != nil {
				return Temperament{}, wrapCentsError(op, err)
			}
		}
		if !sum.Equal(sevenOctavesCents) {
			return Temperament{}, errorf(ErrInvalidArgument, "%s: 12 fifths do not close the circle (sum %v, want 8400 cents)", op, sum)
		}
	}

	var pos [12]cents.Cents
	j0 := -kMin
	pos[j0] = cents.Zero()
	for j := j0; j < 11; j++ {
		next, err := pos[j].Add(fifths[j])
		if err != nil {
			return Temperament{}, wrapCentsError(op, err)
		}
		pos[j+1] = next
	}
	for j := j0 - 1; j >= 0; j-- {
		prev, err := pos[j+1].Sub(fifths[j])
		if err != nil {
			return Temperament{}, wrapCentsError(op, err)
		}
		pos[j] = prev
	}

	offsets := make([]cents.Cents, DefaultStepsPerOctave)
	for j := range 12 {
		k := kMin + j
		m := (7 * k) % 12
		if m < 0 {
			m += 12
		}
		_, offsets[m] = pos[j].SplitOctaves()
	}
	return newChecked(op, name, offsets)
}

// MustFromFifthsWithStart is like [FromFifthsWithStart] but panics on error.
func MustFromFifthsWithStart(name string, fifths []cents.Cents, kMin int) Temperament {
	return must(FromFifthsWithStart(name, fifths, kMin))
}

// FromFifthDeviations creates a 12-tone temperament from 11 or 12 signed
// deviations delta_i in cents from the pure fifth R(3:2) (so
// Q_i = R(3:2) + delta_i) starting at C (kMin = 0). If len(deviations) == 12,
// it verifies exact closure: sum(deviations) == -cents.PythagoreanComma().
func FromFifthDeviations(name string, deviations []cents.Cents) (Temperament, error) {
	return FromFifthDeviationsWithStart(name, deviations, 0)
}

// MustFromFifthDeviations is like [FromFifthDeviations] but panics on error.
func MustFromFifthDeviations(name string, deviations []cents.Cents) Temperament {
	return must(FromFifthDeviations(name, deviations))
}

// FromFifthDeviationsWithStart creates a 12-tone temperament from 11 or 12
// signed deviations delta_j in cents from R(3:2) along the chain starting at
// fifth index kMin in [-11, 0] (with k = 0 fixed at C = 0 cents), where
// deviations[j] (0 <= j <= 10) specifies the deviation of the fifth transition
// from chain node k = kMin + j to k + 1 = kMin + j + 1. If
// len(deviations) == 12, deviations[11] closes the circle from kMin + 11 to
// kMin + 12 and exact closure sum(deviations) == -cents.PythagoreanComma() is
// verified.
func FromFifthDeviationsWithStart(name string, deviations []cents.Cents, kMin int) (Temperament, error) {
	if len(deviations) != 11 && len(deviations) != 12 {
		return Temperament{}, errorf(ErrInvalidArgument, "FromFifthDeviationsWithStart: deviations length must be 11 or 12, got %d", len(deviations))
	}
	if kMin < -11 || kMin > 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "FromFifthDeviationsWithStart: kMin %d out of range [-11, 0]", kMin)
	}
	fifths := make([]cents.Cents, len(deviations))
	for j, d := range deviations {
		q, err := pureFifth.Add(d)
		if err != nil {
			return Temperament{}, wrapCentsError("FromFifthDeviationsWithStart", err)
		}
		fifths[j] = q
	}
	return fromFifths("FromFifthDeviationsWithStart", name, fifths, kMin)
}

// MustFromFifthDeviationsWithStart is like [FromFifthDeviationsWithStart] but
// panics on error.
func MustFromFifthDeviationsWithStart(name string, deviations []cents.Cents, kMin int) Temperament {
	return must(FromFifthDeviationsWithStart(name, deviations, kMin))
}

// FromGenerator creates a regular temperament of dimension N
// (1 <= stepsPerOctave <= MaxStepsPerOctave) from the N reduced chain points
// P_k = (k * generator) mod 1200 cents in [0, 1200) for k in
// [kMin, kMin+stepsPerOctave-1], where kMin must be in
// [-(stepsPerOctave-1), 0], so the chain always contains k = 0 with P_0 = 0.
// Any generator in the range of package cents is accepted, including zero,
// negative values, and values of 1200 cents or more; only generator mod 1200
// affects the result. The points are computed exactly without forming
// k * generator, so FromGenerator never returns [ErrOutOfRange]. The N points
// are sorted in ascending order and must be pairwise distinct. It returns
// [ErrInvalidArgument] if stepsPerOctave or kMin is out of range or if two
// chain points coincide, and [ErrTooComplex] if the exact value of some chain
// point exceeds the representation limits of package cents.
func FromGenerator(name string, stepsPerOctave int, generator cents.Cents, kMin int) (Temperament, error) {
	const op = "FromGenerator"
	if stepsPerOctave < 1 || stepsPerOctave > MaxStepsPerOctave {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: stepsPerOctave %d out of range [1, %d]", op, stepsPerOctave, MaxStepsPerOctave)
	}
	if kMin < -(stepsPerOctave-1) || kMin > 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "%s: kMin %d out of range [%d, 0]", op, kMin, -(stepsPerOctave - 1))
	}
	// Reduce the generator to [0, 1200) and walk the chain from k = 0 in both
	// directions, reducing every point modulo the octave. Each intermediate
	// sum lies in (-1200, 2400) cents, so no range error can occur, and each
	// point is exactly (k * generator) mod 1200 because cents values are exact
	// and canonical.
	_, gr := generator.SplitOctaves()
	j0 := -kMin // index of k = 0 in points
	points := make([]cents.Cents, stepsPerOctave)
	points[j0] = cents.Zero()
	for j := j0 + 1; j < stepsPerOctave; j++ {
		sum, err := points[j-1].Add(gr)
		if err != nil {
			return Temperament{}, wrapCentsError(op, err)
		}
		_, points[j] = sum.SplitOctaves()
	}
	for j := j0 - 1; j >= 0; j-- {
		diff, err := points[j+1].Sub(gr)
		if err != nil {
			return Temperament{}, wrapCentsError(op, err)
		}
		_, points[j] = diff.SplitOctaves()
	}
	slices.SortFunc(points, cents.Cents.Compare)
	return newChecked(op, name, points)
}

// MustFromGenerator is like [FromGenerator] but panics on error.
func MustFromGenerator(name string, stepsPerOctave int, generator cents.Cents, kMin int) Temperament {
	return must(FromGenerator(name, stepsPerOctave, generator, kMin))
}

// WerckmeisterIII returns Werckmeister's III (1691) well temperament
// ("Werckmeister III", N = 12).
func WerckmeisterIII() Temperament {
	return presetWerckmeisterIII
}

// WerckmeisterIV returns Werckmeister's IV (1691) well temperament
// ("Werckmeister IV", N = 12).
func WerckmeisterIV() Temperament {
	return presetWerckmeisterIV
}

// WerckmeisterV returns Werckmeister's V (1691) well temperament
// ("Werckmeister V", N = 12).
func WerckmeisterV() Temperament {
	return presetWerckmeisterV
}

// KirnbergerI returns Kirnberger's first tuning ("Kirnberger I", N = 12).
func KirnbergerI() Temperament {
	return presetKirnbergerI
}

// KirnbergerII returns Kirnberger's second well temperament
// ("Kirnberger II", N = 12).
func KirnbergerII() Temperament {
	return presetKirnbergerII
}

// KirnbergerIII returns Kirnberger's third well temperament
// ("Kirnberger III", N = 12).
func KirnbergerIII() Temperament {
	return presetKirnbergerIII
}

// Vallotti returns the well temperament now known as Vallotti's ("Vallotti",
// N = 12, 6 fifths F..B narrowed by 1/6 Pc, the other 6 fifths pure).
func Vallotti() Temperament {
	return presetVallotti
}

// YoungI returns Young's first well temperament ("Young I", N = 12).
func YoungI() Temperament {
	return presetYoungI
}

// YoungII returns Young's second well temperament ("Young II", N = 12,
// 6 fifths C..F# narrowed by 1/6 Pc).
func YoungII() Temperament {
	return presetYoungII
}

// Young is an alias for [YoungII] ("Young II", N = 12).
func Young() Temperament {
	return presetYoungII
}

// Kellner returns Kellner's (1975) well temperament ("Kellner", N = 12,
// 5 fifths narrowed by 1/5 Pc).
func Kellner() Temperament {
	return presetKellner
}

// NeidhardtI returns Neidhardt's first temperament of 1724, for a village; the
// same scheme is his 1732 temperament for a small town ("Neidhardt I", N = 12:
// fifths C-G, G-D, D-A, A-E narrowed by 1/6 Pc; E-B, B-F#, G#-Eb, Eb-Bb by 1/12
// Pc; F#-C#, C#-G#, Bb-F, F-C pure).
func NeidhardtI() Temperament {
	return presetNeidhardtI
}

// NeidhardtII returns Neidhardt's second temperament of 1724, for a small town;
// the same scheme is his 1732 temperament for a large town ("Neidhardt II",
// N = 12: fifths C-G, G-D, D-A narrowed by 1/6 Pc; A-E, B-F#, F#-C#, C#-G#,
// Bb-F, F-C by 1/12 Pc; E-B, G#-Eb, Eb-Bb pure).
func NeidhardtII() Temperament {
	return presetNeidhardtII
}

// NeidhardtIII returns Neidhardt's third temperament of 1724, for a large town
// ("Neidhardt III", N = 12: fifths C-G, G-D, D-A narrowed by 1/6 Pc; A-E, B-F#,
// F#-C#, C#-G#, Eb-Bb, Bb-F by 1/12 Pc; E-B, G#-Eb, F-C pure).
func NeidhardtIII() Temperament {
	return presetNeidhardtIII
}

// NeidhardtDorf1732 returns Neidhardt's 1732 temperament for a village, the
// first temperament of his circle of thirds ("Neidhardt Dorf (1732)", N = 12:
// fifth C-G narrowed by 1/12 Pc, G-D by 1/6 Pc, D-A and A-E by 1/4 Pc, B-F#,
// C#-G#, Bb-F by 1/12 Pc; E-B, F#-C#, G#-Eb, Eb-Bb, F-C pure).
func NeidhardtDorf1732() Temperament {
	return presetNeidhardtDorf1732
}

// Presets returns a new slice containing all 31 canonical built-in temperaments
// in canonical family order: 6 EDO presets, 2 Pythagorean presets, 4 Just
// presets, 5 Meantone presets, and 14 Well temperament presets.
func Presets() []Temperament {
	out := make([]Temperament, len(canonicalPresets))
	copy(out, canonicalPresets[:])
	return out
}
