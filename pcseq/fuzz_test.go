package pcseq_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzSequenceRoundTrip(f *testing.F) {
	for _, seed := range []struct {
		modulus uint32
		a, b, c int
	}{
		{modulus: 12, a: 7, b: 11, c: 7},
		{modulus: 10, a: 0, b: 0, c: 9},
		{modulus: 1, a: math.MinInt, b: math.MaxInt, c: -1},
		{modulus: 65_536, a: math.MinInt, b: math.MaxInt, c: 65_535},
	} {
		f.Add(seed.modulus, seed.a, seed.b, seed.c)
	}
	f.Fuzz(func(t *testing.T, modulus uint32, a, b, c int) {
		if modulus == 0 || modulus > 65_536 {
			t.Skip("outside the supported modulus")
		}
		s, err := pcseq.Reduce(int(modulus), a, b, c)
		require.NoError(t, err)
		parsed, err := pcseq.Parse(s.String())
		require.NoError(t, err)
		assert.Equal(t, s, parsed)

		styles := []pitchclass.Style{pitchclass.Decimal}
		if modulus <= 12 {
			styles = append(styles, pitchclass.TE, pitchclass.LowerTE, pitchclass.AB)
		}
		for _, style := range styles {
			notation, err := s.Notation(style)
			require.NoError(t, err)
			parsed, err := pcseq.ParseNotation(notation, style, int(modulus))
			require.NoError(t, err)
			assert.Equal(t, s, parsed)
			if modulus <= 10 || style != pitchclass.Decimal {
				compact, err := s.CompactNotation(style)
				require.NoError(t, err)
				parsed, err = pcseq.ParseNotation(compact, style, int(modulus))
				require.NoError(t, err)
				assert.Equal(t, s, parsed)
			}
		}
	})
}

func FuzzTransformInvariants(f *testing.F) {
	f.Add(uint32(12), 7, 11, 9, 2, 5)
	f.Add(uint32(7), math.MinInt, math.MaxInt, -3, math.MinInt, math.MaxInt)
	f.Add(uint32(1), 0, 0, 0, math.MinInt, math.MaxInt)
	f.Fuzz(func(t *testing.T, modulus uint32, a, b, c, n, k int) {
		if modulus == 0 || modulus > 65_536 {
			t.Skip("outside the supported modulus")
		}
		s, err := pcseq.Reduce(int(modulus), a, b, c)
		require.NoError(t, err)

		inverted, err := s.Invert(n)
		require.NoError(t, err)
		inverted, err = inverted.Invert(n)
		require.NoError(t, err)
		assert.Equal(t, s, inverted)

		retro, err := s.Retrograde()
		require.NoError(t, err)
		retro, err = retro.Retrograde()
		require.NoError(t, err)
		assert.Equal(t, s, retro)

		transposed, err := s.Transpose(n)
		require.NoError(t, err)
		transposedRetro, err := transposed.Retrograde()
		require.NoError(t, err)
		retroTransposed, err := retro.Transpose(n)
		require.NoError(t, err)
		assert.Equal(t, transposedRetro, retroTransposed)

		intervals, err := s.Intervals()
		require.NoError(t, err)
		transposedIntervals, err := transposed.Intervals()
		require.NoError(t, err)
		assert.Equal(t, classNumbers(intervals), classNumbers(transposedIntervals))

		rotated, err := s.Rotate(k)
		require.NoError(t, err)
		content, err := s.Content()
		require.NoError(t, err)
		rotatedContent, err := rotated.Content()
		require.NoError(t, err)
		assert.True(t, content.Equal(rotatedContent))
	})
}
