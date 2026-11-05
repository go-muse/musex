package pitch_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndParts(t *testing.T) {
	t.Parallel()

	spellings := sampleSpellings()
	octaves := sampleOctaves()

	require.Len(t, spellings, 39, "expected 35 predefined + 4 boundary/fractional spellings")
	require.Len(t, octaves, 7, "expected 7 sample octaves")

	for _, s := range spellings {
		for _, oct := range octaves {
			name := fmt.Sprintf("%s_%s", s.String(), oct.String())
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				p, err := pitch.New(s, oct)
				require.NoError(t, err)
				assert.True(t, p.IsValid())
				assert.Equal(t, s, p.Spelling())
				assert.True(t, p.Spelling().Equal(s))
				assert.Equal(t, oct, p.Octave())
				assert.True(t, p.Octave().Equal(oct))

				gotS, gotO := p.Parts()
				assert.Equal(t, s, gotS)
				assert.Equal(t, oct, gotO)

				aspn, err := p.ASPN()
				require.NoError(t, err)
				assert.Equal(t, oct.MustASPN(), aspn)
				assert.Equal(t, oct.MustASPN(), p.MustASPN())

				mustP := pitch.MustNew(s, oct)
				assert.Equal(t, p, mustP)
				assert.True(t, p.Equal(mustP))
			})
		}
	}
}

func TestFromASPN(t *testing.T) {
	t.Parallel()

	spellings := []spelling.Spelling{
		spelling.C(),
		spelling.CSharp(),
		spelling.BFlat(),
		cPlusHalf(),
	}
	aspnValues := []int{
		math.MinInt,
		-100,
		-1,
		0,
		4,
		9,
		100,
		math.MaxInt,
	}

	for _, s := range spellings {
		for _, a := range aspnValues {
			name := fmt.Sprintf("%s_ASPN_%d", s.String(), a)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				p, err := pitch.FromASPN(s, a)
				require.NoError(t, err)
				assert.True(t, p.IsValid())

				want := pitch.MustNew(s, octave.FromASPN(a))
				assert.Equal(t, want, p)
				assert.True(t, p.Equal(want))
				assert.Equal(t, s, p.Spelling())
				assert.Equal(t, octave.FromASPN(a), p.Octave())

				gotS, gotO := p.Parts()
				assert.Equal(t, s, gotS)
				assert.Equal(t, octave.FromASPN(a), gotO)

				gotASPN, err := p.ASPN()
				require.NoError(t, err)
				assert.Equal(t, a, gotASPN)
				assert.Equal(t, a, p.MustASPN())

				mustP := pitch.MustFromASPN(s, a)
				assert.Equal(t, p, mustP)
				assert.True(t, p.Equal(mustP))
			})
		}
	}
}

func TestCreationInvalidInputs(t *testing.T) {
	t.Parallel()

	var (
		zeroS spelling.Spelling
		zeroO octave.Octave
	)

	newCases := []struct {
		name string
		s    spelling.Spelling
		o    octave.Octave
	}{
		{
			name: "unset_spelling_valid_octave",
			s:    zeroS,
			o:    o(4),
		},
		{
			name: "valid_spelling_invalid_octave",
			s:    spelling.C(),
			o:    zeroO,
		},
		{
			name: "unset_spelling_invalid_octave",
			s:    zeroS,
			o:    zeroO,
		},
	}

	for _, tc := range newCases {
		t.Run("New_"+tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := pitch.New(tc.s, tc.o)
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Equal(t, pitch.Pitch{}, got)
			assert.False(t, got.IsValid())

			requireMustPanicsIs(t, func() {
				_ = pitch.MustNew(tc.s, tc.o)
			}, pitch.ErrInvalid)
		})
	}

	fromASPNCases := []struct {
		name string
		s    spelling.Spelling
		aspn int
	}{
		{
			name: "unset_spelling_aspn_4",
			s:    zeroS,
			aspn: 4,
		},
		{
			name: "unset_spelling_aspn_0",
			s:    zeroS,
			aspn: 0,
		},
		{
			name: "unset_spelling_aspn_min",
			s:    zeroS,
			aspn: math.MinInt,
		},
		{
			name: "unset_spelling_aspn_max",
			s:    zeroS,
			aspn: math.MaxInt,
		},
	}

	for _, tc := range fromASPNCases {
		t.Run("FromASPN_"+tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := pitch.FromASPN(tc.s, tc.aspn)
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Equal(t, pitch.Pitch{}, got)
			assert.False(t, got.IsValid())

			requireMustPanicsIs(t, func() {
				_ = pitch.MustFromASPN(tc.s, tc.aspn)
			}, pitch.ErrInvalid)
		})
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	var zeroVar pitch.Pitch
	assert.False(t, zeroVar.IsValid())
	assert.False(t, (pitch.Pitch{}).IsValid())

	validPitches := []pitch.Pitch{
		pa(spelling.C(), 4),
		pa(spelling.CSharp(), 0),
		pa(spelling.BFlat(), -1),
		pa(cPlusHalf(), math.MinInt),
		pa(bMinusSevenHalves(), math.MaxInt),
	}
	for _, p := range validPitches {
		assert.True(t, p.IsValid(), "expected valid pitch: %v", p)
	}
}
