package pitch_test

import (
	"math"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type concurrencySnapshot struct {
	createdNew      pitch.Pitch
	createdNewErr   error
	createdASPN     pitch.Pitch
	createdASPNErr  error
	isValid         bool
	zeroValid       bool
	sp              spelling.Spelling
	oct             octave.Octave
	partsS          spelling.Spelling
	partsO          octave.Octave
	aspn            int
	aspnErr         error
	mustASPN        int
	isNatural       bool
	isInteger       bool
	direction       spelling.Direction
	sameSpelling    bool
	sameOctave      bool
	sameLetter      bool
	sameAlteration  bool
	equalSelf       bool
	equalOther      bool
	compareOther    int
	withSpelling    pitch.Pitch
	withOctave      pitch.Pitch
	withASPN        pitch.Pitch
	natural         pitch.Pitch
	raised          pitch.Pitch
	lowered         pitch.Pitch
	raisedBy2       pitch.Pitch
	loweredBy2      pitch.Pitch
	shiftedUp3      pitch.Pitch
	shiftedDown2    pitch.Pitch
	upPitch         pitch.Pitch
	downPitch       pitch.Pitch
	octSub          int
	octDist         int
	str             string
	goStr           string
	zeroStr         string
	zeroGoStr       string
	zeroASPN        int
	zeroASPNErr     error
	rangeItems      []pitch.Pitch
	earlyBreakItems []pitch.Pitch
}

func computeSnapshot(
	p1, p2, pHalf, zero pitch.Pitch,
	sharedRangeSeq, sharedFullSeq func(func(pitch.Pitch) bool),
) concurrencySnapshot {
	createdNew, createdNewErr := pitch.New(spelling.FSharp(), o(4))
	createdASPN, createdASPNErr := pitch.FromASPN(spelling.FSharp(), 4)
	partsS, partsO := p1.Parts()
	aspn, aspnErr := p1.ASPN()
	zeroASPN, zeroASPNErr := zero.ASPN()

	return concurrencySnapshot{
		createdNew:      createdNew,
		createdNewErr:   createdNewErr,
		createdASPN:     createdASPN,
		createdASPNErr:  createdASPNErr,
		isValid:         p1.IsValid(),
		zeroValid:       zero.IsValid(),
		sp:              p1.Spelling(),
		oct:             p1.Octave(),
		partsS:          partsS,
		partsO:          partsO,
		aspn:            aspn,
		aspnErr:         aspnErr,
		mustASPN:        p1.MustASPN(),
		isNatural:       pHalf.IsNatural(),
		isInteger:       pHalf.IsInteger(),
		direction:       pHalf.Direction(),
		sameSpelling:    p1.SameSpelling(pa(spelling.FSharp(), 7)),
		sameOctave:      p1.SameOctave(pa(spelling.C(), 4)),
		sameLetter:      p1.SameLetter(pa(spelling.FFlat(), 2)),
		sameAlteration:  p1.SameAlteration(pa(spelling.CSharp(), 2)),
		equalSelf:       p1.Equal(pitch.MustNew(spelling.FSharp(), o(4))),
		equalOther:      p1.Equal(p2),
		compareOther:    p1.Compare(p2),
		withSpelling:    p1.MustWithSpelling(spelling.C()),
		withOctave:      p1.MustWithOctave(o(7)),
		withASPN:        p1.MustWithASPN(8),
		natural:         p1.Natural(),
		raised:          p1.MustRaise(),
		lowered:         p1.MustLower(),
		raisedBy2:       p1.MustRaiseBy(2),
		loweredBy2:      p1.MustLowerBy(2),
		shiftedUp3:      p1.MustShift(3),
		shiftedDown2:    p1.MustShiftDown(2),
		upPitch:         p1.MustUp(),
		downPitch:       p1.MustDown(),
		octSub:          p2.MustOctaveSub(p1),
		octDist:         p1.MustOctaveDistance(p2),
		str:             p1.String(),
		goStr:           p1.GoString(),
		zeroStr:         zero.String(),
		zeroGoStr:       zero.GoString(),
		zeroASPN:        zeroASPN,
		zeroASPNErr:     zeroASPNErr,
		rangeItems:      collectAll(sharedRangeSeq),
		earlyBreakItems: collectLimit(sharedFullSeq, 3),
	}
}

func TestConcurrentOperationsAndSharedRange(t *testing.T) {
	t.Parallel()

	p1 := pa(spelling.FSharp(), 4)
	p2 := pa(spelling.DFlat(), 6)
	pHalf := pa(cPlusHalf(), -1)
	var zero pitch.Pitch

	sharedRangeSeq := pitch.MustRange(spelling.FSharp(), o(2), o(5))
	sharedFullSeq := pitch.MustRange(spelling.C(), o(math.MinInt), o(math.MaxInt))

	want := computeSnapshot(p1, p2, pHalf, zero, sharedRangeSeq, sharedFullSeq)
	require.NoError(t, want.createdNewErr)
	require.NoError(t, want.createdASPNErr)
	require.NoError(t, want.aspnErr)
	requirePitchErrorIs(t, want.zeroASPNErr, pitch.ErrInvalid)

	workers := max(4, runtime.GOMAXPROCS(0)*4)
	const iterations = 25

	results := make([][]concurrencySnapshot, workers)
	for w := range workers {
		results[w] = make([]concurrencySnapshot, iterations)
	}

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for iterIdx := range iterations {
				results[w][iterIdx] = computeSnapshot(p1, p2, pHalf, zero, sharedRangeSeq, sharedFullSeq)
			}
		})
	}
	wg.Wait()

	// Verify shared values were never mutated.
	assert.Equal(t, pa(spelling.FSharp(), 4), p1)
	assert.Equal(t, pa(spelling.DFlat(), 6), p2)
	assert.Equal(t, pa(cPlusHalf(), -1), pHalf)
	assert.Equal(t, pitch.Pitch{}, zero)

	for w := range workers {
		for iterIdx, got := range results[w] {
			require.NoError(t, got.createdNewErr, "worker=%d iter=%d", w, iterIdx)
			require.NoError(t, got.createdASPNErr, "worker=%d iter=%d", w, iterIdx)
			require.NoError(t, got.aspnErr, "worker=%d iter=%d", w, iterIdx)
			requirePitchErrorIs(t, got.zeroASPNErr, pitch.ErrInvalid)

			got.zeroASPNErr = nil
			expected := want
			expected.zeroASPNErr = nil
			assert.Equal(t, expected, got, "worker=%d iter=%d", w, iterIdx)
		}
	}
}
