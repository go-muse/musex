package pitch_test

import (
	"iter"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pa(s spelling.Spelling, aspn int) pitch.Pitch {
	return pitch.MustFromASPN(s, aspn)
}

func o(aspn int) octave.Octave {
	return octave.FromASPN(aspn)
}

func cPlusHalf() spelling.Spelling {
	return spelling.MustNew(letter.C, alteration.MustFromFraction(1, 2))
}

func cMinusHalf() spelling.Spelling {
	return spelling.MustNew(letter.C, alteration.MustFromFraction(-1, 2))
}

func cPlusThreeHalves() spelling.Spelling {
	return spelling.MustNew(letter.C, alteration.MustFromFraction(3, 2))
}

func bMinusSevenHalves() spelling.Spelling {
	return spelling.MustNew(letter.B, alteration.MustFromFraction(-7, 2))
}

func cMaxAlteration() spelling.Spelling {
	return spelling.MustNew(letter.C, alteration.MustFromInt(alteration.MaxMagnitude))
}

func cMinAlteration() spelling.Spelling {
	return spelling.MustNew(letter.C, alteration.MustFromInt(-alteration.MaxMagnitude))
}

// sampleSpellings returns all 35 predefined spellings plus fractional and boundary spellings for K-1.
func sampleSpellings() []spelling.Spelling {
	predefined := spelling.Predefined()
	out := make([]spelling.Spelling, 0, len(predefined)+4)
	out = append(out, predefined...)
	out = append(out,
		cPlusHalf(),
		bMinusSevenHalves(),
		cMaxAlteration(),
		cMinAlteration(),
	)
	return out
}

// sampleOctaves returns the representative and boundary octaves for K-1.
func sampleOctaves() []octave.Octave {
	return []octave.Octave{
		octave.FromASPN(math.MinInt),
		octave.FromASPN(-1),
		octave.FromASPN(0),
		octave.FromASPN(4),
		octave.FromASPN(math.MaxInt),
		octave.MustFromNumber(3, number.FrancoBelgian()),
		octave.MustFromName("первая", naming.Russian()),
	}
}

// requirePitchErrorIs verifies that err matches wantTarget via errors.Is and does not match
// any other sentinel error from pitch or its component domains.
func requirePitchErrorIs(t *testing.T, err error, wantTarget error) {
	t.Helper()
	require.Error(t, err)
	require.ErrorIs(t, err, wantTarget)

	allTargets := []error{
		pitch.ErrInvalid,
		alteration.ErrOutOfRange,
		octave.ErrOutOfRange,
		letter.ErrInvalid,
		spelling.ErrInvalidForm,
		octave.ErrInvalid,
		alteration.ErrZeroDenominator,
	}
	for _, target := range allTargets {
		if target == wantTarget {
			continue
		}
		assert.NotErrorIs(t, err, target)
	}
}

// requireMustPanicsIs verifies that calling fn panics with an error matching wantTarget via errors.Is
// and not matching any other sentinel error.
func requireMustPanicsIs(t *testing.T, fn func(), wantTarget error) {
	t.Helper()
	var (
		panicked  bool
		recovered any
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				recovered = r
			}
		}()
		fn()
	}()
	require.True(t, panicked, "expected Must function to panic")
	panickedErr, ok := recovered.(error)
	require.True(t, ok, "expected panic value to implement error, got %T", recovered)
	requirePitchErrorIs(t, panickedErr, wantTarget)
}

func collectAll(seq iter.Seq[pitch.Pitch]) []pitch.Pitch {
	var out []pitch.Pitch
	for p := range seq {
		out = append(out, p)
	}
	return out
}

func collectLimit(seq iter.Seq[pitch.Pitch], limit int) []pitch.Pitch {
	if limit <= 0 {
		return nil
	}
	out := make([]pitch.Pitch, 0, limit)
	for p := range seq {
		out = append(out, p)
		if len(out) >= limit {
			break
		}
	}
	return out
}
