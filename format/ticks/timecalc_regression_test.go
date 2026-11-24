package ticks_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireUnboundedExport(t *testing.T, m tempomap.Map, options ticks.ExportOptions, want []string) {
	t.Helper()
	before := m.String()
	var got ticks.MIDITempoMap
	var skipped []tempomap.Extension
	var err error
	require.NotPanics(t, func() { got, skipped, err = ticks.MIDITempoMapOf(m, res(480), options) })
	require.NoError(t, err)
	assert.Nil(t, skipped)
	assert.Equal(t, want, tempoChanges(t, got))
	var must ticks.MIDITempoMap
	require.NotPanics(t, func() { must, skipped = ticks.MustMIDITempoMapOf(m, res(480), options) })
	assert.Nil(t, skipped)
	assert.True(t, must.Equal(got))
	assert.Equal(t, before, m.String())
}

// Architecture 15.4: a=1/M and h=1/2 are valid inputs, but a+h has
// denominator 2*M. That internal position must stay in big.Rat until
// quantization. Decimal(110) independently gives the two EqualTime tempos
// 810930.216216328763956026... and 575364.144903561854878438... .
func TestMIDITempoMapOfEqualTimeWideComputedStep(t *testing.T) {
	if testing.Short() {
		t.Skip("equal-time export with a wide computed step: heavy single-threaded check; runs in the coverage job without -short")
	}
	t.Parallel()
	f, err := fraction.FromBigInt(big.NewInt(1), fraction.MaxComponentBig())
	require.NoError(t, err)
	p, err := duration.PositionFromFraction(f)
	require.NoError(t, err)
	a, err := tempomap.At(p)
	require.NoError(t, err)
	m, err := tempomap.New(quarterTempo(60, 1))
	require.NoError(t, err)
	gradual, err := tempomap.NewGradual(a, tempomap.LinearSpeed)
	require.NoError(t, err)
	m, err = m.WithChange(gradual)
	require.NoError(t, err)
	main, err := tempomap.NewMain(at(1, 1), quarterTempo(120, 1))
	require.NoError(t, err)
	m, err = m.WithChange(main)
	require.NoError(t, err)
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, mode)
			o, err := o.WithSteps(dur(1, 2), ticks.StepEqualTime)
			require.NoError(t, err)
			o, err = o.WithPositionMode(rounding.Down)
			require.NoError(t, err)
			want := changeTexts(0, 1000000, 0, int(irrationalIndex(mode, 810930, false)), 960, int(irrationalIndex(mode, 575364, false)), 1920, 500000)
			requireUnboundedExport(t, m, o, want)
		})
	}
}

// N=(M-1)/3 is an integer, L=N/M is valid, and p=1/2 gives
// start=(M+2)/(6*M). Its reduced denominator is 2*M, outside fraction.
// Down quantizes start to tick 320. The original quarter steps remain
// unchanged: the split inherits m0, and S=(m0/3+m1)/1e6 seconds from the
// final 160- and 480-tick fragments. E=.001 gives k=1+.001/S.
// Start/End use exact rational forms; EqualTime uses the two-form cycle.
// Independent Decimal(110) gives the stretched pairs:
//
//	Start: 500726.315789473684..., 522497.025171624713...
//	End:   522464.405160057333..., 546212.787212787212...
//	Equal: 511441.171930075404..., 534179.217881831295...
//
// No unrounded or pre-quantization fragment may replace S.
func TestMIDITempoMapOfStretchWideCutAcrossGradualSteps(t *testing.T) {
	t.Parallel()
	maximum := fraction.MaxComponentBig()
	numerator := new(big.Int).Quo(new(big.Int).Sub(maximum, big.NewInt(1)), big.NewInt(3))
	f, err := fraction.FromBigInt(numerator, maximum)
	require.NoError(t, err)
	length, err := duration.FromFraction(f)
	require.NoError(t, err)
	// Prove the fixture's computed split really exceeds the public bound.
	split := new(big.Rat).Sub(big.NewRat(1, 2), new(big.Rat).SetFrac(numerator, maximum))
	assert.Positive(t, split.Denom().Cmp(maximum))
	extension, err := tempomap.NewExtension(at(1, 2), sec(1, 1000))
	require.NoError(t, err)
	m, err := ritardandoMap().WithExtension(extension)
	require.NoError(t, err)
	for _, tt := range []struct {
		method ticks.StepMethod
		lower  [5]int64
		up     [5]bool
	}{
		{method: ticks.StepStart, lower: [5]int64{500000, 500726, 522497, 545454, 571428}, up: [5]bool{false, false, false, true, true}},
		{method: ticks.StepEnd, lower: [5]int64{521739, 522464, 546212, 571428, 600000}, up: [5]bool{false, false, true, true, false}},
		{method: ticks.StepEqualTime, lower: [5]int64{510715, 511441, 534179, 558240, 585481}, up: [5]bool{false, false, false, false, true}},
	} {
		t.Run(tt.method.String(), func(t *testing.T) {
			t.Parallel()
			for _, mode := range allModes {
				t.Run(mode.String(), func(t *testing.T) {
					t.Parallel()
					values := make([]int, len(tt.lower))
					for i, lower := range tt.lower {
						values[i] = int(irrationalIndex(mode, lower, tt.up[i]))
					}
					// These two values are integers, so directed modes cannot add one.
					if tt.method == ticks.StepStart {
						values[0] = 500000
					}
					if tt.method == ticks.StepEnd {
						values[4] = 600000
					}
					o := exportOptions(t, mode).MustWithPositionMode(rounding.Down).MustWithSteps(dur(1, 4), tt.method).MustWithHoldsStretched(length)
					want := changeTexts(0, values[0], 320, values[1], 480, values[2], 960, values[3], 1440, values[4], 1920, 600000)
					requireUnboundedExport(t, m, o, want)
				})
			}
		})
	}
}
