package frequency_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/require"
)

func TestNumericOperations_ConcurrentUse(t *testing.T) {
	const workers = 24
	const iterations = 25
	expectedTranspose, err := cents.MustStep(7, 12).ApplyTo(440)
	require.NoError(t, err)
	expectedPeriod, ok := modelPeriod(440)
	require.True(t, ok)
	indices := []int{2, 3, 4}
	subharmonicIndices := []int{1, 2, 3, 4}

	var wait sync.WaitGroup
	errors := make(chan error, workers)
	for worker := range workers {
		wait.Go(func() {
			for range iterations {
				transposed, err := frequency.Transpose(440, 7, 12)
				if err != nil || transposed != expectedTranspose {
					errors <- fmt.Errorf("worker %d: Transpose = %v, %v", worker, transposed, err)
					return
				}
				ratio, err := frequency.Ratio(440, 660)
				if err != nil || ratio.RatString() != "3/2" {
					errors <- fmt.Errorf("worker %d: Ratio = %v, %v", worker, ratio, err)
					return
				}
				ratio.SetInt64(100)
				independentRatio, err := frequency.Ratio(440, 660)
				if err != nil || independentRatio.RatString() != "3/2" {
					errors <- fmt.Errorf("worker %d: Ratio results are shared: %v, %v", worker, independentRatio, err)
					return
				}
				distance, err := frequency.CentsBetween(440, 442)
				if err != nil || !distance.Equal(cents.MustParse("7.851415")) {
					errors <- fmt.Errorf("worker %d: CentsBetween = %v, %v", worker, distance, err)
					return
				}
				partials, err := frequency.Partials(110, indices)
				if err != nil || len(partials) != 3 || partials[0] != 220 || partials[1] != 330 || partials[2] != 440 {
					errors <- fmt.Errorf("worker %d: Partials = %v, %v", worker, partials, err)
					return
				}
				partials[0] = 0
				independentPartials, err := frequency.Partials(110, indices)
				if err != nil || independentPartials[0] != 220 {
					errors <- fmt.Errorf("worker %d: Partials results are shared: %v, %v", worker, independentPartials, err)
					return
				}
				subharmonics, err := frequency.Subharmonics(440, subharmonicIndices)
				if err != nil || len(subharmonics) != 4 || subharmonics[0] != 440 || subharmonics[1] != 220 || subharmonics[3] != 110 {
					errors <- fmt.Errorf("worker %d: Subharmonics = %v, %v", worker, subharmonics, err)
					return
				}
				subharmonics[0] = 0
				independentSubharmonics, err := frequency.Subharmonics(440, subharmonicIndices)
				if err != nil || independentSubharmonics[0] != 440 {
					errors <- fmt.Errorf("worker %d: Subharmonics results are shared: %v, %v", worker, independentSubharmonics, err)
					return
				}
				beat, err := frequency.BeatFrequency(440, 442)
				if err != nil || beat != 2 {
					errors <- fmt.Errorf("worker %d: BeatFrequency = %v, %v", worker, beat, err)
					return
				}
				shifted, err := frequency.ShiftOctaves(440, 1)
				if err != nil || shifted != 880 {
					errors <- fmt.Errorf("worker %d: ShiftOctaves = %v, %v", worker, shifted, err)
					return
				}
				representative, k, err := frequency.NormalizeOctave(880, 110)
				if err != nil || representative != 110 || k != -3 {
					errors <- fmt.Errorf("worker %d: NormalizeOctave = %v, %d, %v", worker, representative, k, err)
					return
				}
				period, err := frequency.Period(440)
				if err != nil || period != expectedPeriod {
					errors <- fmt.Errorf("worker %d: Period = %v, %v", worker, period, err)
					return
				}
			}
		})
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}
