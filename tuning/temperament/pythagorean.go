package temperament

import (
	"fmt"

	"github.com/go-muse/musex/cents"
)

var (
	pureFifth         = cents.MustFromRatio(3, 2)
	pureMajorThird    = cents.MustFromRatio(5, 4)
	pureMinorThird    = cents.MustFromRatio(6, 5)
	sevenOctavesCents = cents.Octave().MustMul(7)

	presetPythagorean            = buildPythagorean(DefaultPythagoreanMinFifth)
	presetPythagoreanEFlatGSharp = buildPythagorean(EFlatPythagoreanMinFifth)
)

// pythagoreanName returns the canonical name for a 12-tone Pythagorean chain starting at kMin.
func pythagoreanName(kMin int) string {
	switch kMin {
	case DefaultPythagoreanMinFifth:
		return "Pythagorean"
	case EFlatPythagoreanMinFifth:
		return "Pythagorean (Eb-G#)"
	default:
		return fmt.Sprintf("Pythagorean (kMin=%d)", kMin)
	}
}

// buildPythagorean constructs a 12-tone Pythagorean temperament for a valid kMin in [-11, 0].
func buildPythagorean(kMin int) Temperament {
	offsets := make([]cents.Cents, DefaultStepsPerOctave)
	for k := kMin; k <= kMin+11; k++ {
		m := ((7 * k) % 12)
		if m < 0 {
			m += 12
		}
		_, offsets[m] = pureFifth.MustMul(k).SplitOctaves()
	}
	return makeTemperament(pythagoreanName(kMin), offsets)
}

// NewPythagorean creates a 12-tone Pythagorean tuning from a chain of 11 pure
// fifths (3:2) spanning fifth indices k in [kMin, kMin+11], where kMin must be
// in [-11, 0]. For kMin == -5 the name is "Pythagorean"; for kMin == -3 it is
// "Pythagorean (Eb-G#)"; for other values it is
// fmt.Sprintf("Pythagorean (kMin=%d)", kMin).
func NewPythagorean(kMin int) (Temperament, error) {
	if kMin < -11 || kMin > 0 {
		return Temperament{}, errorf(ErrInvalidArgument, "NewPythagorean(%d): kMin out of range [-11, 0]", kMin)
	}
	switch kMin {
	case DefaultPythagoreanMinFifth:
		return presetPythagorean, nil
	case EFlatPythagoreanMinFifth:
		return presetPythagoreanEFlatGSharp, nil
	default:
		return buildPythagorean(kMin), nil
	}
}

// MustNewPythagorean is like [NewPythagorean] but panics on error.
func MustNewPythagorean(kMin int) Temperament {
	return must(NewPythagorean(kMin))
}

// NewPythagoreanWithWolf creates a 12-tone Pythagorean tuning placing the wolf
// fifth at chromatic step wolfPosition -> (wolfPosition+7)%12, where
// wolfPosition must be in [0, 11]. It is equivalent to
// NewPythagorean(((7*wolfPosition)%12) - 11).
func NewPythagoreanWithWolf(wolfPosition int) (Temperament, error) {
	if wolfPosition < 0 || wolfPosition > 11 {
		return Temperament{}, errorf(ErrInvalidArgument, "NewPythagoreanWithWolf(%d): wolfPosition out of range [0, 11]", wolfPosition)
	}
	kMin := ((7 * wolfPosition) % 12) - 11
	return NewPythagorean(kMin)
}

// MustNewPythagoreanWithWolf is like [NewPythagoreanWithWolf] but panics on
// error.
func MustNewPythagoreanWithWolf(wolfPosition int) Temperament {
	return must(NewPythagoreanWithWolf(wolfPosition))
}

// Pythagorean returns the C-symmetric 12-tone Pythagorean tuning
// ("Pythagorean", k in [-5, +6], wolf fifth F#-Db at 6->1).
func Pythagorean() Temperament {
	return presetPythagorean
}

// PythagoreanEFlatGSharp returns the 12-tone Pythagorean tuning on the Eb..G#
// chain of fifths ("Pythagorean (Eb-G#)", k in [-3, +8], wolf fifth G#-Eb at
// 8->3).
func PythagoreanEFlatGSharp() Temperament {
	return presetPythagoreanEFlatGSharp
}
