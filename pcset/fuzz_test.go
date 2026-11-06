package pcset_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzParse(f *testing.F) {
	seeds := []string{
		"{} mod 1",
		"{0} mod 1",
		"{} mod 12",
		"{0, 4, 7} mod 12",
		"{0, 10, 11} mod 12",
		"{0, 65535} mod 65536",
		"{} mod 65537",
		"{0} mod 100000",
		"",
		"Set(invalid)",
		"{0, 4, 7}",
		"[0, 4, 7] mod 12",
		"{4, 0} mod 12",
		"{0, 0} mod 12",
		"{12} mod 12",
		"{-1} mod 12",
		"{01} mod 12",
		"{+1} mod 12",
		"{-0} mod 12",
		"{} mod 0",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var (
			got pcset.Set
			err error
		)
		require.NotPanics(t, func() {
			got, err = pcset.Parse(input)
		})

		if err != nil {
			assert.False(t, got.IsValid())
			assert.True(
				t,
				errors.Is(err, pcset.ErrNotation) || errors.Is(err, pcset.ErrModulusTooLarge),
				"unexpected error %v for input %q",
				err,
				input,
			)

			return
		}

		require.True(t, got.IsValid())
		assert.Equal(t, input, got.String())
	})
}

func FuzzParseNotation(f *testing.F) {
	inputs := []string{
		"{}",
		"[]",
		"{0, 4, 7}",
		"{0, 4, T, E}",
		"{04TE}",
		"[E, 2, 3, 7]",
		"[E237]",
		"{10}",
		"{047}",
		"{7, 0, 4, 0}",
		"{0, 4, 7]",
		"",
	}
	for _, inp := range inputs {
		for _, st := range []uint8{0, 1, 2, 3, 99} {
			for _, mod := range []int{0, 1, 10, 11, 12, 19, 65_536, 65_537} {
				f.Add(inp, st, mod)
			}
		}
	}

	f.Fuzz(func(t *testing.T, input string, styleByte uint8, modulus int) {
		style := pitchclass.Style(styleByte)

		var (
			got pcset.Set
			err error
		)
		require.NotPanics(t, func() {
			got, err = pcset.ParseNotation(input, style, modulus)
		})

		if err != nil {
			assert.False(t, got.IsValid())
			switch {
			case modulus < 1:
				requireSentinelError(t, err, pcset.ErrInvalidModulus)
			case modulus > pcset.MaxModulus:
				requireSentinelError(t, err, pcset.ErrModulusTooLarge)
			default:
				requireSentinelError(t, err, pcset.ErrNotation)
			}

			return
		}

		require.True(t, got.IsValid())
		assert.Equal(t, modulus, got.Modulus())

		formatted, err := got.Notation(style)
		require.NoError(t, err)
		roundTrip, err := pcset.ParseNotation(formatted, style, modulus)
		require.NoError(t, err)
		assert.Equal(t, got, roundTrip)
	})
}

func FuzzSetOperations(f *testing.F) {
	f.Add(uint16(145), uint16(2009), 12, 5, 7)
	f.Add(uint16(0), uint16(4095), 12, -3, 2)
	f.Add(uint16(27), uint16(54), 7, 11, -5)
	f.Add(uint16(1), uint16(0), 1, 0, 0)

	f.Fuzz(func(t *testing.T, maskA, maskB uint16, modRaw, step, mult int) {
		mod := int(uint(modRaw)%16) + 1
		limit := uint64(1) << mod
		ma := uint64(maskA) % limit
		mb := uint64(maskB) % limit

		a := pcset.MustFromBitmask(ma, mod)
		b := pcset.MustFromBitmask(mb, mod)
		oa := newOracleSet(mod, a.Numbers()...)
		ob := newOracleSet(mod, b.Numbers()...)

		assertSet(t, a.Transpose(step), mod, transposeOracle(oa, step).elems)
		assertSet(t, a.Invert(step), mod, invertOracle(oa, step).elems)
		assertSet(t, a.Multiply(mult), mod, multiplyOracle(oa, mult).elems)
		assertSet(t, a.Complement(), mod, complementOracle(oa).elems)
		assertSet(t, a.MustUnion(b), mod, unionOracle(oa, ob).elems)
		assertSet(t, a.MustIntersection(b), mod, intersectionOracle(oa, ob).elems)
		assertSet(t, a.MustDifference(b), mod, differenceOracle(oa, ob).elems)
		assertSet(t, a.MustSymmetricDifference(b), mod, symDiffOracle(oa, ob).elems)

		assert.Equal(t, normalFormRahnOracle(mod, oa.elems), numbersOfClasses(a.NormalForm()))
		assert.Equal(t, normalFormForteOracle(mod, oa.elems), numbersOfClasses(a.MustNormalFormWith(pcset.Forte)))
		assert.Equal(t, icvOracle(mod, oa.elems), a.IntervalVector().Slice())

		parsed, err := pcset.Parse(a.String())
		require.NoError(t, err)
		assert.Equal(t, a, parsed)
	})
}
