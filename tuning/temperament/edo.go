package temperament

import (
	"strconv"

	"github.com/go-muse/musex/cents"
)

var (
	presetEqual12 = buildEDO(12)
	presetEqual19 = buildEDO(19)
	presetEqual22 = buildEDO(22)
	presetEqual24 = buildEDO(24)
	presetEqual31 = buildEDO(31)
	presetEqual53 = buildEDO(53)
)

// buildEDO constructs an N-EDO temperament for a valid stepsPerOctave in [1, MaxStepsPerOctave].
func buildEDO(stepsPerOctave int) Temperament {
	offsets := make([]cents.Cents, stepsPerOctave)
	offsets[0] = cents.Zero()
	for m := 1; m < stepsPerOctave; m++ {
		offsets[m] = cents.MustStep(m, stepsPerOctave)
	}
	return makeTemperament(strconv.Itoa(stepsPerOctave)+"-EDO", offsets)
}

// EDO creates an N-tone equal division of the octave
// (1 <= stepsPerOctave <= MaxStepsPerOctave) with canonical name "<N>-EDO" and
// exact steps T(m) = cents.Step(m, N).
func EDO(stepsPerOctave int) (Temperament, error) {
	if stepsPerOctave < 1 || stepsPerOctave > MaxStepsPerOctave {
		return Temperament{}, errorf(ErrInvalidArgument, "EDO(%d): stepsPerOctave out of range [1, %d]", stepsPerOctave, MaxStepsPerOctave)
	}
	switch stepsPerOctave {
	case 12:
		return presetEqual12, nil
	case 19:
		return presetEqual19, nil
	case 22:
		return presetEqual22, nil
	case 24:
		return presetEqual24, nil
	case 31:
		return presetEqual31, nil
	case 53:
		return presetEqual53, nil
	default:
		return buildEDO(stepsPerOctave), nil
	}
}

// MustEDO is like [EDO] but panics if stepsPerOctave is out of range
// [1, MaxStepsPerOctave].
func MustEDO(stepsPerOctave int) Temperament {
	return must(EDO(stepsPerOctave))
}

// Equal12 returns the standard 12-tone equal temperament ("12-EDO", N = 12).
func Equal12() Temperament {
	return presetEqual12
}

// Equal is an alias for [Equal12] ("12-EDO", N = 12).
func Equal() Temperament {
	return presetEqual12
}

// EDO12 is an alias for [Equal12] ("12-EDO", N = 12).
func EDO12() Temperament {
	return presetEqual12
}

// Equal19 returns the 19-tone equal temperament ("19-EDO", N = 19).
func Equal19() Temperament {
	return presetEqual19
}

// EDO19 is an alias for [Equal19] ("19-EDO", N = 19).
func EDO19() Temperament {
	return presetEqual19
}

// Equal22 returns the 22-tone equal temperament ("22-EDO", N = 22).
func Equal22() Temperament {
	return presetEqual22
}

// EDO22 is an alias for [Equal22] ("22-EDO", N = 22).
func EDO22() Temperament {
	return presetEqual22
}

// Equal24 returns the 24-tone quarter-tone equal temperament ("24-EDO", N = 24).
func Equal24() Temperament {
	return presetEqual24
}

// EDO24 is an alias for [Equal24] ("24-EDO", N = 24).
func EDO24() Temperament {
	return presetEqual24
}

// Equal31 returns the 31-tone equal temperament ("31-EDO", N = 31).
func Equal31() Temperament {
	return presetEqual31
}

// EDO31 is an alias for [Equal31] ("31-EDO", N = 31).
func EDO31() Temperament {
	return presetEqual31
}

// Equal53 returns the 53-tone equal temperament ("53-EDO", N = 53).
func Equal53() Temperament {
	return presetEqual53
}

// EDO53 is an alias for [Equal53] ("53-EDO", N = 53).
func EDO53() Temperament {
	return presetEqual53
}
