package pitch_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRange(t *testing.T) {
	t.Parallel()

	fSharp := spelling.FSharp()
	c := spelling.C()
	cHalf := cPlusHalf()

	tests := []struct {
		name string
		s    spelling.Spelling
		from octave.Octave
		to   octave.Octave
		want []pitch.Pitch
	}{
		{
			name: "ascending_F_sharp_from_2_to_5",
			s:    fSharp,
			from: o(2),
			to:   o(5),
			want: []pitch.Pitch{
				pa(fSharp, 2),
				pa(fSharp, 3),
				pa(fSharp, 4),
				pa(fSharp, 5),
			},
		},
		{
			name: "descending_F_sharp_from_5_to_2",
			s:    fSharp,
			from: o(5),
			to:   o(2),
			want: []pitch.Pitch{
				pa(fSharp, 5),
				pa(fSharp, 4),
				pa(fSharp, 3),
				pa(fSharp, 2),
			},
		},
		{
			name: "single_element_F_sharp_4_to_4",
			s:    fSharp,
			from: o(4),
			to:   o(4),
			want: []pitch.Pitch{
				pa(fSharp, 4),
			},
		},
		{
			name: "ascending_at_MaxInt_boundary",
			s:    c,
			from: o(math.MaxInt - 2),
			to:   o(math.MaxInt),
			want: []pitch.Pitch{
				pa(c, math.MaxInt-2),
				pa(c, math.MaxInt-1),
				pa(c, math.MaxInt),
			},
		},
		{
			name: "descending_at_MinInt_boundary",
			s:    c,
			from: o(math.MinInt + 2),
			to:   o(math.MinInt),
			want: []pitch.Pitch{
				pa(c, math.MinInt+2),
				pa(c, math.MinInt+1),
				pa(c, math.MinInt),
			},
		},
		{
			name: "single_element_at_MaxInt",
			s:    c,
			from: o(math.MaxInt),
			to:   o(math.MaxInt),
			want: []pitch.Pitch{
				pa(c, math.MaxInt),
			},
		},
		{
			name: "single_element_at_MinInt",
			s:    c,
			from: o(math.MinInt),
			to:   o(math.MinInt),
			want: []pitch.Pitch{
				pa(c, math.MinInt),
			},
		},
		{
			name: "fractional_spelling_across_negative_and_positive_octaves",
			s:    cHalf,
			from: octave.MustFromName("малая", naming.Russian()),
			to:   octave.MustFromNumber(4, number.FrancoBelgian()),
			want: []pitch.Pitch{
				pa(cHalf, 3),
				pa(cHalf, 4),
				pa(cHalf, 5),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			seq, err := pitch.Range(tc.s, tc.from, tc.to)
			require.NoError(t, err)
			require.NotNil(t, seq)

			got := collectAll(seq)
			assert.Equal(t, tc.want, got)

			// Re-iterating the same sequence starts from the beginning.
			gotSecond := collectAll(seq)
			assert.Equal(t, tc.want, gotSecond)

			for _, p := range got {
				assert.True(t, p.IsValid())
				assert.Equal(t, tc.s, p.Spelling())
			}

			mustSeq := pitch.MustRange(tc.s, tc.from, tc.to)
			require.NotNil(t, mustSeq)
			assert.Equal(t, tc.want, collectAll(mustSeq))
		})
	}
}

func TestRangeEarlyBreakAndReiteration(t *testing.T) {
	t.Parallel()

	c := spelling.C()
	seq, err := pitch.Range(c, o(math.MinInt), o(math.MaxInt))
	require.NoError(t, err)
	require.NotNil(t, seq)

	wantFirstThree := []pitch.Pitch{
		pa(c, math.MinInt),
		pa(c, math.MinInt+1),
		pa(c, math.MinInt+2),
	}

	gotFirst := collectLimit(seq, 3)
	assert.Equal(t, wantFirstThree, gotFirst)

	// Re-iterating the same sequence after an early break starts from math.MinInt again.
	gotSecond := collectLimit(seq, 3)
	assert.Equal(t, wantFirstThree, gotSecond)

	// Also verify descending full range with early break.
	descSeq := pitch.MustRange(c, o(math.MaxInt), o(math.MinInt))
	wantDescThree := []pitch.Pitch{
		pa(c, math.MaxInt),
		pa(c, math.MaxInt-1),
		pa(c, math.MaxInt-2),
	}
	assert.Equal(t, wantDescThree, collectLimit(descSeq, 3))
	assert.Equal(t, wantDescThree, collectLimit(descSeq, 3))
}

func TestRangeInvalidInputs(t *testing.T) {
	t.Parallel()

	var (
		zs spelling.Spelling
		zo octave.Octave
	)
	c := spelling.C()

	tests := []struct {
		name string
		s    spelling.Spelling
		from octave.Octave
		to   octave.Octave
	}{
		{
			name: "zero_spelling",
			s:    zs,
			from: o(2),
			to:   o(4),
		},
		{
			name: "zero_from_octave",
			s:    c,
			from: zo,
			to:   o(4),
		},
		{
			name: "zero_to_octave",
			s:    c,
			from: o(2),
			to:   zo,
		},
		{
			name: "zero_spelling_and_zero_from",
			s:    zs,
			from: zo,
			to:   o(4),
		},
		{
			name: "zero_spelling_and_zero_to",
			s:    zs,
			from: o(2),
			to:   zo,
		},
		{
			name: "zero_from_and_zero_to",
			s:    c,
			from: zo,
			to:   zo,
		},
		{
			name: "all_zero_arguments",
			s:    zs,
			from: zo,
			to:   zo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			seq, err := pitch.Range(tc.s, tc.from, tc.to)
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Nil(t, seq)

			requireMustPanicsIs(t, func() {
				_ = pitch.MustRange(tc.s, tc.from, tc.to)
			}, pitch.ErrInvalid)
		})
	}
}
