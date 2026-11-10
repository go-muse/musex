package temperament

import (
	"fmt"

	"github.com/go-muse/musex/cents"
)

var (
	// minMeantoneFifth and maxMeantoneFifth are the exclusive bounds
	// (4800/7, 720) cents of a fifth generator that yields a strictly
	// increasing 12-tone regular meantone table.
	minMeantoneFifth = cents.MustFromFraction(4800, 7)
	maxMeantoneFifth = cents.MustFromInt(720)

	presetMeantoneQuarterComma    = mustBuildMeantonePreset(1, 4)
	presetMeantoneThirdComma      = mustBuildMeantonePreset(1, 3)
	presetMeantoneFifthComma      = mustBuildMeantonePreset(1, 5)
	presetMeantoneSixthComma      = mustBuildMeantonePreset(1, 6)
	presetMeantoneTwoSeventhComma = mustBuildMeantonePreset(2, 7)
)

// mustBuildMeantonePreset builds a standard kMin = -3 meantone preset.
func mustBuildMeantonePreset(commaNum, commaDen int) Temperament {
	adj := cents.SyntonicComma().MustScale(commaNum, commaDen)
	fifth := pureFifth.MustSub(adj)
	name := fmt.Sprintf("%d/%d-Comma Meantone", commaNum, commaDen)
	return must(buildMeantoneFromFifth("NewMeantone", name, fifth, DefaultMeantoneMinFifth))
}

// checkMeantoneFifth reports [ErrInvalidArgument] unless fifth lies strictly
// inside (4800/7, 720) cents. The fifth is compared as given, without any
// reduction modulo the octave.
func checkMeantoneFifth(op string, fifth cents.Cents) error {
	if fifth.Compare(minMeantoneFifth) <= 0 || fifth.Compare(maxMeantoneFifth) >= 0 {
		return errorf(ErrInvalidArgument, "%s: fifth %v outside (4800/7, 720) cents", op, fifth)
	}
	return nil
}

// buildMeantoneFromFifth constructs a 12-tone regular meantone temperament with
// the given name, fifth generator, and chain start kMin. The caller has
// already validated kMin in [-11, 0] and the fifth in (4800/7, 720) cents with
// [checkMeantoneFifth]; op names the public operation in error messages.
func buildMeantoneFromFifth(op, name string, fifth cents.Cents, kMin int) (Temperament, error) {
	offsets := make([]cents.Cents, DefaultStepsPerOctave)
	for k := kMin; k <= kMin+11; k++ {
		m := (7 * k) % 12
		if m < 0 {
			m += 12
		}
		raw, err := fifth.Mul(k)
		if err != nil {
			return Temperament{}, wrapCentsError(op, err)
		}
		_, offsets[m] = raw.SplitOctaves()
	}
	return newChecked(op, name, offsets)
}

// gcdUint64 returns the greatest common divisor of a and b.
func gcdUint64(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// absUint64 returns |x| as a uint64 without overflow.
func absUint64(x int) uint64 {
	if x < 0 {
		return -uint64(x)
	}
	return uint64(x)
}

// NewMeantone creates a 12-tone regular meantone temperament with fifth
// generator g = R(3:2) - (commaNum/commaDen)*SyntonicComma and default chain
// start kMin = -3 (Eb..G#). It returns [ErrInvalidArgument] if commaDen <= 0 or
// if the resulting fifth g is outside (4800/7, 720) cents.
func NewMeantone(commaNum, commaDen int) (Temperament, error) {
	return NewMeantoneWithRange(commaNum, commaDen, DefaultMeantoneMinFifth)
}

// MustNewMeantone is like [NewMeantone] but panics on error.
func MustNewMeantone(commaNum, commaDen int) Temperament {
	return must(NewMeantone(commaNum, commaDen))
}

// NewMeantoneWithRange creates a 12-tone regular meantone temperament with
// fifth generator g = R(3:2) - (commaNum/commaDen)*SyntonicComma and explicit
// chain start kMin in [-11, 0].
func NewMeantoneWithRange(commaNum, commaDen, kMin int) (Temperament, error) {
	if commaDen <= 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "NewMeantoneWithRange: non-positive commaDen %d", commaDen)
	}
	if kMin < -11 || kMin > 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "NewMeantoneWithRange: kMin %d out of range [-11, 0]", kMin)
	}
	if commaNum == 0 {
		return NewPythagorean(kMin)
	}
	// The fifth range (4800/7, 720) cents corresponds to a comma fraction
	// alpha in about (-0.839, 0.755), so |commaNum/commaDen| >= 1 is always
	// outside it. Rejecting it here, before any cents arithmetic, reports such
	// arguments as invalid instead of as a cents range error.
	if absUint64(commaNum) >= uint64(commaDen) {
		return Temperament{}, errorf(ErrInvalidArgument, "NewMeantoneWithRange: comma fraction %d/%d outside (-1, 1)", commaNum, commaDen)
	}
	g := int(gcdUint64(absUint64(commaNum), uint64(commaDen)))
	num := commaNum / g
	den := commaDen / g
	if kMin == DefaultMeantoneMinFifth {
		switch {
		case num == 1 && den == 4:
			return presetMeantoneQuarterComma, nil
		case num == 1 && den == 3:
			return presetMeantoneThirdComma, nil
		case num == 1 && den == 5:
			return presetMeantoneFifthComma, nil
		case num == 1 && den == 6:
			return presetMeantoneSixthComma, nil
		case num == 2 && den == 7:
			return presetMeantoneTwoSeventhComma, nil
		}
	}
	adj, err := cents.SyntonicComma().Scale(num, den)
	if err != nil {
		return Temperament{}, wrapCentsError("NewMeantoneWithRange", err)
	}
	fifth, err := pureFifth.Sub(adj)
	if err != nil {
		return Temperament{}, wrapCentsError("NewMeantoneWithRange", err)
	}
	if err := checkMeantoneFifth("NewMeantoneWithRange", fifth); err != nil {
		return Temperament{}, err
	}
	var name string
	if kMin == DefaultMeantoneMinFifth {
		name = fmt.Sprintf("%d/%d-Comma Meantone", num, den)
	} else {
		name = fmt.Sprintf("%d/%d-Comma Meantone (kMin=%d)", num, den, kMin)
	}
	return buildMeantoneFromFifth("NewMeantoneWithRange", name, fifth, kMin)
}

// MustNewMeantoneWithRange is like [NewMeantoneWithRange] but panics on error.
func MustNewMeantoneWithRange(commaNum, commaDen, kMin int) Temperament {
	return must(NewMeantoneWithRange(commaNum, commaDen, kMin))
}

// NewMeantoneFromFifth creates a 12-tone regular meantone temperament from an
// explicit fifth generator in (4800/7, 720) cents and default chain start
// kMin = -3 (Eb..G#).
func NewMeantoneFromFifth(fifth cents.Cents) (Temperament, error) {
	return NewMeantoneFromFifthWithRange(fifth, DefaultMeantoneMinFifth)
}

// MustNewMeantoneFromFifth is like [NewMeantoneFromFifth] but panics on error.
func MustNewMeantoneFromFifth(fifth cents.Cents) Temperament {
	return must(NewMeantoneFromFifth(fifth))
}

// NewMeantoneFromFifthWithRange creates a 12-tone regular meantone temperament
// from an explicit fifth generator in (4800/7, 720) cents and explicit chain
// start kMin in [-11, 0].
func NewMeantoneFromFifthWithRange(fifth cents.Cents, kMin int) (Temperament, error) {
	if kMin < -11 || kMin > 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "NewMeantoneFromFifthWithRange: kMin %d out of range [-11, 0]", kMin)
	}
	const op = "NewMeantoneFromFifthWithRange"
	if err := checkMeantoneFifth(op, fifth); err != nil {
		return Temperament{}, err
	}
	var name string
	if kMin == DefaultMeantoneMinFifth {
		name = fmt.Sprintf("Meantone (fifth=%s)", fifth.String())
	} else {
		name = fmt.Sprintf("Meantone (fifth=%s, kMin=%d)", fifth.String(), kMin)
	}
	return buildMeantoneFromFifth(op, name, fifth, kMin)
}

// MustNewMeantoneFromFifthWithRange is like [NewMeantoneFromFifthWithRange] but
// panics on error.
func MustNewMeantoneFromFifthWithRange(fifth cents.Cents, kMin int) Temperament {
	return must(NewMeantoneFromFifthWithRange(fifth, kMin))
}

// MeantoneQuarterComma returns the canonical 1/4-comma meantone temperament
// ("1/4-Comma Meantone", N = 12, k in [-3, +8]).
func MeantoneQuarterComma() Temperament {
	return presetMeantoneQuarterComma
}

// Meantone is an alias for [MeantoneQuarterComma] ("1/4-Comma Meantone",
// N = 12).
func Meantone() Temperament {
	return presetMeantoneQuarterComma
}

// MeantoneThirdComma returns the 1/3-comma meantone temperament
// ("1/3-Comma Meantone", N = 12, k in [-3, +8]).
func MeantoneThirdComma() Temperament {
	return presetMeantoneThirdComma
}

// MeantoneFifthComma returns the 1/5-comma meantone temperament
// ("1/5-Comma Meantone", N = 12, k in [-3, +8]).
func MeantoneFifthComma() Temperament {
	return presetMeantoneFifthComma
}

// MeantoneSixthComma returns the 1/6-comma meantone temperament
// ("1/6-Comma Meantone", N = 12, k in [-3, +8]).
func MeantoneSixthComma() Temperament {
	return presetMeantoneSixthComma
}

// MeantoneTwoSeventhComma returns the 2/7-comma meantone temperament
// ("2/7-Comma Meantone", N = 12, k in [-3, +8]).
func MeantoneTwoSeventhComma() Temperament {
	return presetMeantoneTwoSeventhComma
}
