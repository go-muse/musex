package pitch_test

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

func TestStringAndGoString(t *testing.T) {
	t.Parallel()

	var zero pitch.Pitch

	tests := []struct {
		name         string
		p            pitch.Pitch
		wantString   string
		wantGoString string
	}{
		{
			name:         "C_ASPN_4",
			p:            pa(spelling.C(), 4),
			wantString:   "C ASPN 4",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("C"), 4)`,
		},
		{
			name:         "C_sharp_ASPN_4",
			p:            pa(spelling.CSharp(), 4),
			wantString:   "C+1 ASPN 4",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("C+1"), 4)`,
		},
		{
			name:         "B_sharp_ASPN_4",
			p:            pa(spelling.BSharp(), 4),
			wantString:   "B+1 ASPN 4",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("B+1"), 4)`,
		},
		{
			name:         "C_minus_half_ASPN_minus_1",
			p:            pa(cMinusHalf(), -1),
			wantString:   "C-1/2 ASPN -1",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("C-1/2"), -1)`,
		},
		{
			name:         "B_minus_seven_halves_ASPN_0",
			p:            pa(bMinusSevenHalves(), 0),
			wantString:   "B-7/2 ASPN 0",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("B-7/2"), 0)`,
		},
		{
			name:         "A_ASPN_MinInt",
			p:            pa(spelling.A(), math.MinInt),
			wantString:   "A ASPN -9223372036854775808",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("A"), -9223372036854775808)`,
		},
		{
			name:         "A_ASPN_MaxInt",
			p:            pa(spelling.A(), math.MaxInt),
			wantString:   "A ASPN 9223372036854775807",
			wantGoString: `pitch.MustFromASPN(spelling.MustParse("A"), 9223372036854775807)`,
		},
		{
			name:         "zero_value_Pitch",
			p:            zero,
			wantString:   "Pitch(invalid)",
			wantGoString: "pitch.Pitch{}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantString, tc.p.String())
			assert.Equal(t, tc.wantGoString, tc.p.GoString())

			// Verify fmt verbs.
			assert.Equal(t, tc.wantString, fmt.Sprint(tc.p))
			assert.Equal(t, tc.wantString+"\n", fmt.Sprintln(tc.p))
			assert.Equal(t, "["+tc.wantString+"]", fmt.Sprintf("[%v]", tc.p))
			assert.Equal(t, "["+tc.wantString+"]", fmt.Sprintf("[%+v]", tc.p))
			assert.Equal(t, "["+tc.wantString+"]", fmt.Sprintf("[%s]", tc.p))
			assert.Equal(t, strconv.Quote(tc.wantString), fmt.Sprintf("%q", tc.p))
			assert.Equal(t, hex.EncodeToString([]byte(tc.wantString)), fmt.Sprintf("%x", tc.p))
			assert.Equal(t, strings.ToUpper(hex.EncodeToString([]byte(tc.wantString))), fmt.Sprintf("%X", tc.p))
			assert.Equal(t, tc.wantGoString, fmt.Sprintf("%#v", tc.p))
		})
	}
}

func TestStringInvarianceAcrossOctaveConventions(t *testing.T) {
	t.Parallel()

	pASPN := pitch.MustFromASPN(spelling.CSharp(), 4)
	pNewASPN := pitch.MustNew(spelling.CSharp(), octave.FromASPN(4))
	pFB := pitch.MustNew(spelling.CSharp(), octave.MustFromNumber(3, number.FrancoBelgian()))
	pRu := pitch.MustNew(spelling.CSharp(), octave.MustFromName("первая", naming.Russian()))

	for _, p := range []pitch.Pitch{pASPN, pNewASPN, pFB, pRu} {
		assert.Equal(t, "C+1 ASPN 4", p.String())
		assert.Equal(t, `pitch.MustFromASPN(spelling.MustParse("C+1"), 4)`, p.GoString())
	}
}

func TestStringPairwiseDistinctness(t *testing.T) {
	t.Parallel()

	pitches := []pitch.Pitch{
		{},
	}
	octaves := []int{math.MinInt, -1, 0, 4, math.MaxInt}
	for _, s := range sampleSpellings() {
		for _, a := range octaves {
			pitches = append(pitches, pa(s, a))
		}
	}

	seenString := make(map[string]pitch.Pitch, len(pitches))
	seenGoString := make(map[string]pitch.Pitch, len(pitches))

	for _, p := range pitches {
		str := p.String()
		if prev, exists := seenString[str]; exists {
			assert.Equal(t, prev, p, "duplicate String() %q for distinct pitches %#v and %#v", str, prev, p)
		} else {
			seenString[str] = p
		}

		goStr := p.GoString()
		if prev, exists := seenGoString[goStr]; exists {
			assert.Equal(t, prev, p, "duplicate GoString() %q for distinct pitches %#v and %#v", goStr, prev, p)
		} else {
			seenGoString[goStr] = p
		}
	}
}
